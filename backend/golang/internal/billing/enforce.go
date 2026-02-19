package billing

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	ddbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/myfusionhelper/api/internal/types"
)

// LimitExceededError is returned when an account has hit a plan limit.
type LimitExceededError struct {
	Resource string // "helpers", "connections", "api_keys", "executions"
	Current  int
	Limit    int
	Plan     string
	Message  string
}

func (e *LimitExceededError) Error() string {
	return e.Message
}

// fetchAccount loads the account from DynamoDB.
func fetchAccount(ctx context.Context, db *dynamodb.Client, accountsTable, accountID string) (*types.Account, error) {
	result, err := db.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(accountsTable),
		Key: map[string]ddbtypes.AttributeValue{
			"account_id": &ddbtypes.AttributeValueMemberS{Value: accountID},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch account: %w", err)
	}
	if result.Item == nil {
		return nil, fmt.Errorf("account not found: %s", accountID)
	}

	var account types.Account
	if err := attributevalue.UnmarshalMap(result.Item, &account); err != nil {
		return nil, fmt.Errorf("failed to unmarshal account: %w", err)
	}
	return &account, nil
}

// checkTrialExpired returns a LimitExceededError if the account's trial has expired.
func checkTrialExpired(account *types.Account) *LimitExceededError {
	if !IsTrialPlan(account.Plan) {
		return nil
	}
	// Check the TrialExpired flag first (set by the trial-expiration worker)
	if account.TrialExpired {
		return &LimitExceededError{
			Resource: "trial",
			Message:  "Your free trial has expired. Please choose a plan to continue.",
		}
	}
	// Also check TrialEndsAt in case the worker hasn't run yet
	if account.TrialEndsAt != nil && time.Now().UTC().After(*account.TrialEndsAt) {
		return &LimitExceededError{
			Resource: "trial",
			Message:  "Your free trial has expired. Please choose a plan to continue.",
		}
	}
	return nil
}

// CheckHelperLimit verifies the account hasn't exceeded its helper limit.
func CheckHelperLimit(ctx context.Context, db *dynamodb.Client, accountsTable, accountID string) error {
	account, err := fetchAccount(ctx, db, accountsTable, accountID)
	if err != nil {
		log.Printf("Billing check failed (allowing): %v", err)
		return nil // fail open
	}

	if expired := checkTrialExpired(account); expired != nil {
		return expired
	}

	if account.Settings.MaxHelpers > 0 && account.Usage.Helpers >= account.Settings.MaxHelpers {
		planLabel := GetPlanLabel(account.Plan)
		return &LimitExceededError{
			Resource: "helpers",
			Current:  account.Usage.Helpers,
			Limit:    account.Settings.MaxHelpers,
			Plan:     account.Plan,
			Message:  fmt.Sprintf("Your %s plan allows %d helpers. Upgrade to add more.", planLabel, account.Settings.MaxHelpers),
		}
	}
	return nil
}

// CheckConnectionLimit verifies the account hasn't exceeded its connection limit.
func CheckConnectionLimit(ctx context.Context, db *dynamodb.Client, accountsTable, accountID string) error {
	account, err := fetchAccount(ctx, db, accountsTable, accountID)
	if err != nil {
		log.Printf("Billing check failed (allowing): %v", err)
		return nil
	}

	if expired := checkTrialExpired(account); expired != nil {
		return expired
	}

	if account.Settings.MaxConnections > 0 && account.Usage.Connections >= account.Settings.MaxConnections {
		planLabel := GetPlanLabel(account.Plan)
		return &LimitExceededError{
			Resource: "connections",
			Current:  account.Usage.Connections,
			Limit:    account.Settings.MaxConnections,
			Plan:     account.Plan,
			Message:  fmt.Sprintf("Your %s plan allows %d connections. Upgrade to add more.", planLabel, account.Settings.MaxConnections),
		}
	}
	return nil
}

// CheckAPIKeyLimit verifies the account hasn't exceeded its API key limit.
func CheckAPIKeyLimit(ctx context.Context, db *dynamodb.Client, accountsTable, accountID string) error {
	account, err := fetchAccount(ctx, db, accountsTable, accountID)
	if err != nil {
		log.Printf("Billing check failed (allowing): %v", err)
		return nil
	}

	if expired := checkTrialExpired(account); expired != nil {
		return expired
	}

	if account.Settings.MaxAPIKeys > 0 && account.Usage.APIKeys >= account.Settings.MaxAPIKeys {
		planLabel := GetPlanLabel(account.Plan)
		return &LimitExceededError{
			Resource: "api_keys",
			Current:  account.Usage.APIKeys,
			Limit:    account.Settings.MaxAPIKeys,
			Plan:     account.Plan,
			Message:  fmt.Sprintf("Your %s plan allows %d API keys. Upgrade to add more.", planLabel, account.Settings.MaxAPIKeys),
		}
	}
	return nil
}

// CheckExecutionLimit verifies the account hasn't exceeded its monthly execution limit.
// For paid plans, this returns nil (overage is metered by Stripe).
// For trial and free plans, this blocks execution at the limit.
func CheckExecutionLimit(ctx context.Context, db *dynamodb.Client, accountsTable, accountID string) error {
	account, err := fetchAccount(ctx, db, accountsTable, accountID)
	if err != nil {
		log.Printf("Billing check failed (allowing): %v", err)
		return nil
	}

	if expired := checkTrialExpired(account); expired != nil {
		return expired
	}

	// Paid plans allow overage — Stripe meters it automatically
	if IsPaidPlan(account.Plan) {
		return nil
	}

	// Trial and free plans: hard-block at limit
	if account.Settings.MaxExecutions > 0 && account.Usage.MonthlyExecutions >= account.Settings.MaxExecutions {
		msg := fmt.Sprintf("You've used all %d executions this month. Choose a plan to continue.", account.Settings.MaxExecutions)
		return &LimitExceededError{
			Resource: "executions",
			Current:  account.Usage.MonthlyExecutions,
			Limit:    account.Settings.MaxExecutions,
			Plan:     account.Plan,
			Message:  msg,
		}
	}
	return nil
}

// IncrementUsage atomically increments a usage counter on the account.
// field should be one of: "helpers", "connections", "api_keys", "team_members", "monthly_executions"
func IncrementUsage(ctx context.Context, db *dynamodb.Client, accountsTable, accountID, field string, delta int) error {
	_, err := db.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(accountsTable),
		Key: map[string]ddbtypes.AttributeValue{
			"account_id": &ddbtypes.AttributeValueMemberS{Value: accountID},
		},
		UpdateExpression: aws.String(fmt.Sprintf("ADD usage.%s :delta", field)),
		ExpressionAttributeValues: map[string]ddbtypes.AttributeValue{
			":delta": &ddbtypes.AttributeValueMemberN{Value: fmt.Sprintf("%d", delta)},
		},
	})
	if err != nil {
		log.Printf("Failed to increment usage.%s for account %s: %v", field, accountID, err)
	}
	return err
}
