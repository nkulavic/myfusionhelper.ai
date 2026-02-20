package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	ddbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

var (
	sqsClient        *sqs.Client
	dbClient         *dynamodb.Client
	notificationQURL string
	usersTable       string
)

// NotificationJob matches the message format expected by the notification-worker
type NotificationJob struct {
	Type      string                 `json:"type"`
	UserID    string                 `json:"user_id"`
	AccountID string                 `json:"account_id"`
	Data      map[string]interface{} `json:"data"`
}

func init() {
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		log.Fatalf("Failed to load AWS config: %v", err)
	}
	sqsClient = sqs.NewFromConfig(cfg)
	dbClient = dynamodb.NewFromConfig(cfg)

	notificationQURL = os.Getenv("NOTIFICATION_QUEUE_URL")
	usersTable = os.Getenv("USERS_TABLE")
}

func Handle(ctx context.Context, event events.DynamoDBEvent) (events.DynamoDBEventResponse, error) {
	var failures []events.DynamoDBBatchItemFailure

	for _, record := range event.Records {
		table := extractTableName(record.EventSourceArn)
		var err error

		switch {
		case strings.Contains(table, "-users"):
			err = handleUsersStream(ctx, record)
		case strings.Contains(table, "-accounts"):
			err = handleAccountsStream(ctx, record)
		case strings.Contains(table, "-webhook-events"):
			err = handleWebhookEventsStream(ctx, record)
		default:
			log.Printf("Unknown source table: %s", table)
		}

		if err != nil {
			log.Printf("ERROR processing record from %s: %v", table, err)
			failures = append(failures, events.DynamoDBBatchItemFailure{
				ItemIdentifier: record.Change.SequenceNumber,
			})
		}
	}

	return events.DynamoDBEventResponse{BatchItemFailures: failures}, nil
}

// handleUsersStream processes INSERT events from the Users table to send welcome emails
func handleUsersStream(ctx context.Context, record events.DynamoDBEventRecord) error {
	if record.EventName != "INSERT" {
		return nil
	}

	newImage := record.Change.NewImage
	email := getStreamString(newImage, "email")
	name := getStreamString(newImage, "name")
	userID := getStreamString(newImage, "user_id")

	if email == "" {
		log.Printf("No email in Users INSERT, skipping")
		return nil
	}

	log.Printf("Users INSERT detected: user=%s email=%s — enqueuing welcome email", userID, email)

	return enqueueNotification(ctx, NotificationJob{
		Type:   "welcome",
		UserID: userID,
		Data: map[string]interface{}{
			"user_name":  name,
			"user_email": email,
		},
	}, userID)
}

// handleAccountsStream processes MODIFY events from the Accounts table to detect trial expiration
func handleAccountsStream(ctx context.Context, record events.DynamoDBEventRecord) error {
	if record.EventName != "MODIFY" {
		return nil
	}

	oldImage := record.Change.OldImage
	newImage := record.Change.NewImage

	// Detect trial_expired changing from false to true
	oldExpired := getStreamBool(oldImage, "trial_expired")
	newExpired := getStreamBool(newImage, "trial_expired")

	if oldExpired || !newExpired {
		return nil // Not a trial expiration event
	}

	accountID := getStreamString(newImage, "account_id")
	ownerUserID := getStreamString(newImage, "owner_user_id")

	log.Printf("Accounts MODIFY detected: account=%s trial_expired flipped true — enqueuing trial_expired email", accountID)

	// Look up user email/name
	userName, userEmail, err := lookupUser(ctx, ownerUserID)
	if err != nil {
		log.Printf("Failed to look up user %s for trial_expired email: %v", ownerUserID, err)
		return err
	}
	if userEmail == "" {
		log.Printf("No email for user %s, skipping trial_expired email", ownerUserID)
		return nil
	}

	return enqueueNotification(ctx, NotificationJob{
		Type:      "billing_event",
		UserID:    ownerUserID,
		AccountID: accountID,
		Data: map[string]interface{}{
			"event_type": "trial_expired",
			"plan_name":  "Trial",
			"email":      userEmail,
			"user_name":  userName,
		},
	}, accountID)
}

// handleWebhookEventsStream processes MODIFY events (status→processed) to send billing emails
func handleWebhookEventsStream(ctx context.Context, record events.DynamoDBEventRecord) error {
	if record.EventName != "MODIFY" {
		return nil
	}

	newImage := record.Change.NewImage

	status := getStreamString(newImage, "status")
	if status != "processed" {
		return nil
	}

	// Check for notification_data field — if absent, no email needed for this event
	notifDataAttr, ok := newImage["notification_data"]
	if !ok || notifDataAttr.DataType() != events.DataTypeMap {
		return nil
	}

	notifMap := convertStreamImage(notifDataAttr.Map())
	eventID := getStreamString(newImage, "event_id")
	eventType := getStreamString(newImage, "event_type")

	log.Printf("WebhookEvents MODIFY detected: event=%s type=%s status=processed — enqueuing billing email", eventID, eventType)

	// Build notification job from embedded notification_data
	notifType, _ := notifMap["notification_type"].(string)
	if notifType == "" {
		notifType = "billing_event"
	}

	userID, _ := notifMap["user_id"].(string)
	accountID, _ := notifMap["account_id"].(string)

	data := map[string]interface{}{}

	// Copy known fields into data
	for _, key := range []string{"event_type", "plan_name", "email", "user_name"} {
		if v, ok := notifMap[key]; ok {
			data[key] = v
		}
	}

	// Merge extra data if present
	if extra, ok := notifMap["extra"]; ok {
		if extraMap, ok := extra.(map[string]interface{}); ok {
			data["extra"] = extraMap
		}
	}

	return enqueueNotification(ctx, NotificationJob{
		Type:      notifType,
		UserID:    userID,
		AccountID: accountID,
		Data:      data,
	}, eventID)
}

// enqueueNotification sends a notification job to the SQS FIFO queue
func enqueueNotification(ctx context.Context, job NotificationJob, groupKey string) error {
	body, err := json.Marshal(job)
	if err != nil {
		return err
	}

	msgGroupID := groupKey
	if len(msgGroupID) > 8 {
		msgGroupID = msgGroupID[:8]
	}
	if msgGroupID == "" {
		msgGroupID = "notif"
	}

	_, err = sqsClient.SendMessage(ctx, &sqs.SendMessageInput{
		QueueUrl:       aws.String(notificationQURL),
		MessageBody:    aws.String(string(body)),
		MessageGroupId: aws.String(msgGroupID),
	})
	if err != nil {
		log.Printf("ERROR: Failed to send notification to SQS: %v", err)
		return err
	}

	log.Printf("Enqueued notification type=%s for user=%s", job.Type, job.UserID)
	return nil
}

// lookupUser fetches name and email from the Users table
func lookupUser(ctx context.Context, userID string) (string, string, error) {
	result, err := dbClient.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(usersTable),
		Key: map[string]ddbtypes.AttributeValue{
			"user_id": &ddbtypes.AttributeValueMemberS{Value: userID},
		},
		ProjectionExpression:     aws.String("#n, email"),
		ExpressionAttributeNames: map[string]string{"#n": "name"},
	})
	if err != nil {
		return "", "", err
	}
	if result.Item == nil {
		return "", "", nil
	}

	var user struct {
		Name  string `dynamodbav:"name"`
		Email string `dynamodbav:"email"`
	}
	if err := attributevalue.UnmarshalMap(result.Item, &user); err != nil {
		return "", "", err
	}
	return user.Name, user.Email, nil
}

// extractTableName parses the DynamoDB table name from the stream ARN
func extractTableName(arn string) string {
	// ARN format: arn:aws:dynamodb:region:account:table/TABLE_NAME/stream/TIMESTAMP
	parts := strings.Split(arn, "/")
	if len(parts) >= 2 {
		return parts[1]
	}
	return arn
}

// getStreamString extracts a string value from a DynamoDB stream image
func getStreamString(image map[string]events.DynamoDBAttributeValue, key string) string {
	if v, ok := image[key]; ok && v.DataType() == events.DataTypeString {
		return v.String()
	}
	return ""
}

// getStreamBool extracts a boolean value from a DynamoDB stream image
func getStreamBool(image map[string]events.DynamoDBAttributeValue, key string) bool {
	if v, ok := image[key]; ok && v.DataType() == events.DataTypeBoolean {
		return v.Boolean()
	}
	return false
}

// convertStreamImage converts a DynamoDB stream image to a Go map
func convertStreamImage(image map[string]events.DynamoDBAttributeValue) map[string]interface{} {
	result := make(map[string]interface{}, len(image))
	for k, v := range image {
		result[k] = convertStreamValue(v)
	}
	return result
}

func convertStreamValue(v events.DynamoDBAttributeValue) interface{} {
	switch v.DataType() {
	case events.DataTypeString:
		return v.String()
	case events.DataTypeNumber:
		return v.Number()
	case events.DataTypeBoolean:
		return v.Boolean()
	case events.DataTypeMap:
		return convertStreamImage(v.Map())
	case events.DataTypeList:
		list := v.List()
		result := make([]interface{}, len(list))
		for i, item := range list {
			result[i] = convertStreamValue(item)
		}
		return result
	case events.DataTypeNull:
		return nil
	default:
		return nil
	}
}

func main() {
	lambda.Start(Handle)
}
