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
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

var queueURL = os.Getenv("HELPER_EXECUTION_QUEUE_URL")

func main() {
	lambda.Start(handler)
}

func handler(ctx context.Context, event events.DynamoDBEvent) error {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(os.Getenv("COGNITO_REGION")))
	if err != nil {
		log.Printf("Failed to load AWS config: %v", err)
		return err
	}
	sqsClient := sqs.NewFromConfig(cfg)

	for _, record := range event.Records {
		if record.EventName != "INSERT" {
			continue
		}

		img := record.Change.NewImage
		status, ok := img["status"]
		if !ok || status.DataType() != events.DataTypeString || status.String() != "queued" {
			continue
		}

		executionID := img["execution_id"].String()
		helperID := img["helper_id"].String()
		contactID := ""
		if v, ok := img["contact_id"]; ok && v.DataType() == events.DataTypeString {
			contactID = v.String()
		}

		// Build the SQS message body from the DynamoDB stream image
		msgBody := map[string]interface{}{
			"execution_id": executionID,
			"helper_id":    helperID,
			"account_id":   img["account_id"].String(),
			"contact_id":   contactID,
		}
		if v, ok := img["connection_id"]; ok && v.DataType() == events.DataTypeString {
			msgBody["connection_id"] = v.String()
		}
		if v, ok := img["user_id"]; ok && v.DataType() == events.DataTypeString {
			msgBody["user_id"] = v.String()
		}

		body, err := json.Marshal(msgBody)
		if err != nil {
			log.Printf("Failed to marshal SQS message for execution %s: %v", executionID, err)
			continue
		}

		// MessageGroupId = helper_id:contact_id for per-helper-per-contact ordering
		groupID := helperID
		if contactID != "" {
			groupID = helperID + ":" + contactID
		}

		_, err = sqsClient.SendMessage(ctx, &sqs.SendMessageInput{
			QueueUrl:               aws.String(queueURL),
			MessageBody:            aws.String(string(body)),
			MessageGroupId:         aws.String(groupID),
			MessageDeduplicationId: aws.String(executionID),
		})
		if err != nil {
			log.Printf("Failed to send SQS message for execution %s: %v", executionID, err)
			return err
		}

		log.Printf("Dispatched execution %s to SQS (group: %s)", executionID, groupID)
	}

	return nil
}
