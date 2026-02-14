package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	ddbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"
	apitypes "github.com/myfusionhelper/api/internal/types"
)

var (
	helpersTable    = os.Getenv("HELPERS_TABLE")
	executionsTable = os.Getenv("EXECUTIONS_TABLE")
)

// ScheduleEvent is the payload sent by EventBridge rules
type ScheduleEvent struct {
	HelperID  string `json:"helper_id"`
	AccountID string `json:"account_id"`
}

func main() {
	lambda.Start(handler)
}

func handler(ctx context.Context, rawEvent json.RawMessage) error {
	var event ScheduleEvent
	if err := json.Unmarshal(rawEvent, &event); err != nil {
		log.Printf("Failed to unmarshal schedule event: %v", err)
		return err
	}

	log.Printf("Scheduled execution for helper %s (account: %s)", event.HelperID, event.AccountID)

	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(os.Getenv("COGNITO_REGION")))
	if err != nil {
		log.Printf("Failed to load AWS config: %v", err)
		return err
	}
	db := dynamodb.NewFromConfig(cfg)

	// Fetch the helper to verify it's still active and scheduled
	helperResult, err := db.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(helpersTable),
		Key: map[string]ddbtypes.AttributeValue{
			"helper_id": &ddbtypes.AttributeValueMemberS{Value: event.HelperID},
		},
	})
	if err != nil || helperResult.Item == nil {
		log.Printf("Helper %s not found, skipping scheduled execution", event.HelperID)
		return nil
	}

	var helper apitypes.Helper
	if err := attributevalue.UnmarshalMap(helperResult.Item, &helper); err != nil {
		log.Printf("Failed to unmarshal helper: %v", err)
		return nil
	}

	// Verify helper is active, enabled, and scheduled
	if helper.Status != "active" || !helper.Enabled {
		log.Printf("Helper %s is not active/enabled, skipping", event.HelperID)
		return nil
	}
	if !helper.ScheduleEnabled {
		log.Printf("Helper %s schedule is disabled, skipping", event.HelperID)
		return nil
	}

	// Create execution record with trigger_type "scheduled"
	now := time.Now().UTC()
	executionID := fmt.Sprintf("exec:%s", uuid.New().String())
	ttl := now.Add(7 * 24 * time.Hour).Unix()

	execution := map[string]ddbtypes.AttributeValue{
		"execution_id":  &ddbtypes.AttributeValueMemberS{Value: executionID},
		"helper_id":     &ddbtypes.AttributeValueMemberS{Value: helper.HelperID},
		"account_id":    &ddbtypes.AttributeValueMemberS{Value: helper.AccountID},
		"helper_type":   &ddbtypes.AttributeValueMemberS{Value: helper.HelperType},
		"connection_id": &ddbtypes.AttributeValueMemberS{Value: helper.ConnectionID},
		"status":        &ddbtypes.AttributeValueMemberS{Value: "queued"},
		"trigger_type":  &ddbtypes.AttributeValueMemberS{Value: "scheduled"},
		"created_at":    &ddbtypes.AttributeValueMemberS{Value: now.Format(time.RFC3339)},
		"ttl":           &ddbtypes.AttributeValueMemberN{Value: fmt.Sprintf("%d", ttl)},
	}

	// Include helper config as input
	if helper.Config != nil {
		configAV, err := attributevalue.MarshalMap(helper.Config)
		if err == nil {
			execution["config"] = &ddbtypes.AttributeValueMemberM{Value: configAV}
		}
	}

	_, err = db.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(executionsTable),
		Item:      execution,
	})
	if err != nil {
		log.Printf("Failed to create scheduled execution for helper %s: %v", event.HelperID, err)
		return err
	}

	// Update helper's last_scheduled_at
	_, err = db.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(helpersTable),
		Key: map[string]ddbtypes.AttributeValue{
			"helper_id": &ddbtypes.AttributeValueMemberS{Value: helper.HelperID},
		},
		UpdateExpression: aws.String("SET last_scheduled_at = :ts"),
		ExpressionAttributeValues: map[string]ddbtypes.AttributeValue{
			":ts": &ddbtypes.AttributeValueMemberS{Value: now.Format(time.RFC3339)},
		},
	})
	if err != nil {
		log.Printf("Failed to update last_scheduled_at for helper %s: %v", event.HelperID, err)
	}

	log.Printf("Created scheduled execution %s for helper %s", executionID, event.HelperID)
	return nil
}
