package stripeusage

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	ddbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	stripe "github.com/stripe/stripe-go/v82"
	"github.com/stripe/stripe-go/v82/billing/meterevent"
	apitypes "github.com/myfusionhelper/api/internal/types"
)

// ReportExecution reports a single execution to Stripe via Billing Meter Events.
// Best-effort: failures are logged but do not block the execution flow.
// Uses execution_id as the event identifier for idempotency (24-hour dedup window).
func ReportExecution(ctx context.Context, db *dynamodb.Client, executionID string, accountID string, completedAtUnix int64) {
	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")
	if stripe.Key == "" {
		log.Printf("STRIPE_SECRET_KEY not set, skipping usage report for execution %s", executionID)
		return
	}

	accountsTable := os.Getenv("ACCOUNTS_TABLE")
	executionsTable := os.Getenv("EXECUTIONS_TABLE")

	// Get account to check plan and Stripe customer ID
	accountResult, err := db.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(accountsTable),
		Key: map[string]ddbtypes.AttributeValue{
			"account_id": &ddbtypes.AttributeValueMemberS{Value: accountID},
		},
	})
	if err != nil || accountResult.Item == nil {
		log.Printf("Failed to get account %s for usage report: %v", accountID, err)
		return
	}

	var account apitypes.Account
	if err := attributevalue.UnmarshalMap(accountResult.Item, &account); err != nil {
		log.Printf("Failed to unmarshal account: %v", err)
		return
	}

	// Only report for paid plans
	if account.Plan == "" || account.Plan == "free" {
		return
	}

	if account.StripeCustomerID == "" {
		log.Printf("No stripe_customer_id for account %s, skipping usage report", accountID)
		return
	}

	// Create Stripe billing meter event with idempotency via Identifier
	params := &stripe.BillingMeterEventParams{
		EventName:  stripe.String("helper_execution"),
		Identifier: stripe.String(fmt.Sprintf("exec-%s", executionID)),
		Payload: map[string]string{
			"stripe_customer_id": account.StripeCustomerID,
			"value":              "1",
		},
		Timestamp: stripe.Int64(completedAtUnix),
	}

	meterEvent, err := meterevent.New(params)
	if err != nil {
		log.Printf("Failed to create Stripe meter event for execution %s: %v", executionID, err)
		return
	}

	log.Printf("Reported execution %s to Stripe (meter event: %s)", executionID, meterEvent.Identifier)

	// Update execution record with Stripe reporting status
	_, err = db.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(executionsTable),
		Key: map[string]ddbtypes.AttributeValue{
			"execution_id": &ddbtypes.AttributeValueMemberS{Value: executionID},
		},
		UpdateExpression: aws.String("SET stripe_reported = :reported, stripe_usage_record_id = :record_id"),
		ExpressionAttributeValues: map[string]ddbtypes.AttributeValue{
			":reported":  &ddbtypes.AttributeValueMemberBOOL{Value: true},
			":record_id": &ddbtypes.AttributeValueMemberS{Value: meterEvent.Identifier},
		},
	})
	if err != nil {
		log.Printf("Failed to update execution %s with Stripe status: %v", executionID, err)
	}
}
