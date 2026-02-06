package main

import (
	"context"
	"encoding/json"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	ddbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/myfusionhelper/api/internal/notifications"
	apitypes "github.com/myfusionhelper/api/internal/types"
)

var (
	usersTable = os.Getenv("USERS_TABLE")
)

// NotificationJob represents a notification message from the SQS queue
type NotificationJob struct {
	Type      string                 `json:"type"`
	UserID    string                 `json:"user_id"`
	AccountID string                 `json:"account_id"`
	Data      map[string]interface{} `json:"data"`
}

func main() {
	lambda.Start(handleSQSEvent)
}

func handleSQSEvent(ctx context.Context, event events.SQSEvent) error {
	log.Printf("Processing %d notification messages", len(event.Records))

	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(os.Getenv("COGNITO_REGION")))
	if err != nil {
		log.Printf("Failed to load AWS config: %v", err)
		return err
	}
	db := dynamodb.NewFromConfig(cfg)

	notifSvc, err := notifications.New(ctx)
	if err != nil {
		log.Printf("Failed to create notification service: %v", err)
		return err
	}

	for _, record := range event.Records {
		var job NotificationJob
		if err := json.Unmarshal([]byte(record.Body), &job); err != nil {
			log.Printf("Failed to unmarshal notification message: %v", err)
			continue
		}

		log.Printf("Processing notification type=%s for user=%s", job.Type, job.UserID)

		// Look up user info
		userName, userEmail, prefs, err := lookupUser(ctx, db, job.UserID)
		if err != nil {
			log.Printf("Failed to look up user %s: %v", job.UserID, err)
			continue
		}

		if userEmail == "" {
			log.Printf("No email found for user %s, skipping", job.UserID)
			continue
		}

		// Dispatch based on notification type
		switch job.Type {
		case "execution_failure":
			helperName := getStringData(job.Data, "helper_name")
			errorMsg := getStringData(job.Data, "error_message")
			if err := notifSvc.SendExecutionAlert(ctx, userName, userEmail, helperName, errorMsg, prefs); err != nil {
				log.Printf("Failed to send execution alert: %v", err)
			}

		case "connection_issue":
			connectionName := getStringData(job.Data, "connection_name")
			errorMsg := getStringData(job.Data, "error_message")
			if err := notifSvc.SendConnectionAlert(ctx, userName, userEmail, connectionName, errorMsg, prefs); err != nil {
				log.Printf("Failed to send connection alert: %v", err)
			}

		case "usage_alert":
			resourceName := getStringData(job.Data, "resource_name")
			current := getIntData(job.Data, "current")
			limit := getIntData(job.Data, "limit")
			percent := getIntData(job.Data, "percent")
			if err := notifSvc.SendUsageAlert(ctx, userName, userEmail, resourceName, current, limit, percent, prefs); err != nil {
				log.Printf("Failed to send usage alert: %v", err)
			}

		case "billing_event":
			eventType := getStringData(job.Data, "event_type")
			planName := getStringData(job.Data, "plan_name")
			if err := notifSvc.SendBillingEvent(ctx, userName, userEmail, eventType, planName); err != nil {
				log.Printf("Failed to send billing event email: %v", err)
			}

		case "weekly_summary":
			totalHelpers := getIntData(job.Data, "total_helpers")
			totalExecuted := getIntData(job.Data, "total_executed")
			totalSucceeded := getIntData(job.Data, "total_succeeded")
			totalFailed := getIntData(job.Data, "total_failed")
			successRate := getStringData(job.Data, "success_rate")
			weekStart := getStringData(job.Data, "week_start")
			weekEnd := getStringData(job.Data, "week_end")
			if err := notifSvc.SendWeeklySummary(ctx, userName, userEmail, totalHelpers, totalExecuted, totalSucceeded, totalFailed, successRate, weekStart, weekEnd, prefs); err != nil {
				log.Printf("Failed to send weekly summary: %v", err)
			}

		case "team_invite":
			recipientEmail := getStringData(job.Data, "recipient_email")
			inviterName := getStringData(job.Data, "inviter_name")
			inviterEmail := getStringData(job.Data, "inviter_email")
			roleName := getStringData(job.Data, "role_name")
			accountName := getStringData(job.Data, "account_name")
			inviteToken := getStringData(job.Data, "invite_token")
			if err := notifSvc.SendTeamInviteEmail(ctx, recipientEmail, inviterName, inviterEmail, roleName, accountName, inviteToken); err != nil {
				log.Printf("Failed to send team invite email: %v", err)
			}

		default:
			log.Printf("Unknown notification type: %s", job.Type)
		}
	}

	return nil
}

func lookupUser(ctx context.Context, db *dynamodb.Client, userID string) (string, string, *apitypes.NotificationPreferences, error) {
	result, err := db.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(usersTable),
		Key: map[string]ddbtypes.AttributeValue{
			"user_id": &ddbtypes.AttributeValueMemberS{Value: userID},
		},
		ProjectionExpression: aws.String("#n, email, notification_preferences"),
		ExpressionAttributeNames: map[string]string{
			"#n": "name",
		},
	})
	if err != nil {
		return "", "", nil, err
	}
	if result.Item == nil {
		return "", "", nil, nil
	}

	var user struct {
		Name                    string                            `dynamodbav:"name"`
		Email                   string                            `dynamodbav:"email"`
		NotificationPreferences *apitypes.NotificationPreferences `dynamodbav:"notification_preferences"`
	}
	if err := attributevalue.UnmarshalMap(result.Item, &user); err != nil {
		return "", "", nil, err
	}

	return user.Name, user.Email, user.NotificationPreferences, nil
}

func getStringData(data map[string]interface{}, key string) string {
	if v, ok := data[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func getIntData(data map[string]interface{}, key string) int {
	if v, ok := data[key]; ok {
		switch n := v.(type) {
		case float64:
			return int(n)
		case int:
			return n
		}
	}
	return 0
}
