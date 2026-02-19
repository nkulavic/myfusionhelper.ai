package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	ddbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

var (
	accountsTable = os.Getenv("ACCOUNTS_TABLE")
)

// trialAccount is a minimal projection of an account for the scan.
type trialAccount struct {
	AccountID    string     `dynamodbav:"account_id"`
	Plan         string     `dynamodbav:"plan"`
	TrialEndsAt  *time.Time `dynamodbav:"trial_ends_at"`
	TrialExpired bool       `dynamodbav:"trial_expired"`
}

func main() {
	lambda.Start(handleScheduleEvent)
}

func handleScheduleEvent(ctx context.Context) error {
	log.Println("Trial expiration worker triggered")

	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(os.Getenv("COGNITO_REGION")))
	if err != nil {
		log.Printf("Failed to load AWS config: %v", err)
		return err
	}

	db := dynamodb.NewFromConfig(cfg)
	now := time.Now().UTC()

	// Scan for accounts where plan is "trial", trial_expired is false,
	// and trial_ends_at is in the past.
	expiredCount := 0
	var lastKey map[string]ddbtypes.AttributeValue

	for {
		input := &dynamodb.ScanInput{
			TableName:        aws.String(accountsTable),
			FilterExpression: aws.String("#plan = :trial AND trial_expired = :false AND trial_ends_at < :now"),
			ExpressionAttributeNames: map[string]string{
				"#plan": "plan",
			},
			ExpressionAttributeValues: map[string]ddbtypes.AttributeValue{
				":trial": &ddbtypes.AttributeValueMemberS{Value: "trial"},
				":false":  &ddbtypes.AttributeValueMemberBOOL{Value: false},
				":now":   &ddbtypes.AttributeValueMemberS{Value: now.Format(time.RFC3339)},
			},
			ProjectionExpression: aws.String("account_id, #plan, trial_ends_at, trial_expired"),
		}

		if lastKey != nil {
			input.ExclusiveStartKey = lastKey
		}

		result, err := db.Scan(ctx, input)
		if err != nil {
			log.Printf("Failed to scan accounts: %v", err)
			return err
		}

		for _, item := range result.Items {
			var account trialAccount
			if err := attributevalue.UnmarshalMap(item, &account); err != nil {
				log.Printf("Failed to unmarshal account: %v", err)
				continue
			}

			if err := markTrialExpired(ctx, db, account.AccountID); err != nil {
				log.Printf("Failed to expire account %s: %v", account.AccountID, err)
				continue
			}

			expiredCount++
			log.Printf("Expired trial for account %s (trial ended: %v)", account.AccountID, account.TrialEndsAt)
		}

		lastKey = result.LastEvaluatedKey
		if lastKey == nil {
			break
		}
	}

	log.Printf("Trial expiration complete: %d accounts expired", expiredCount)
	return nil
}

func markTrialExpired(ctx context.Context, db *dynamodb.Client, accountID string) error {
	_, err := db.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(accountsTable),
		Key: map[string]ddbtypes.AttributeValue{
			"account_id": &ddbtypes.AttributeValueMemberS{Value: accountID},
		},
		UpdateExpression: aws.String("SET trial_expired = :true, updated_at = :now"),
		ConditionExpression: aws.String("trial_expired = :false"),
		ExpressionAttributeValues: map[string]ddbtypes.AttributeValue{
			":true": &ddbtypes.AttributeValueMemberBOOL{Value: true},
			":false": &ddbtypes.AttributeValueMemberBOOL{Value: false},
			":now":  &ddbtypes.AttributeValueMemberS{Value: time.Now().UTC().Format(time.RFC3339)},
		},
	})
	return err
}
