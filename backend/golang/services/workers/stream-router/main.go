package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

var (
	sqsClient *sqs.Client
	stage     string
	region    string
	accountID string
)

func init() {
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		log.Fatalf("Failed to load AWS config: %v", err)
	}
	sqsClient = sqs.NewFromConfig(cfg)

	stage = os.Getenv("STAGE")
	if stage == "" {
		stage = "dev"
	}
	region = os.Getenv("AWS_REGION")
	if region == "" {
		region = "us-west-2"
	}
	accountID = os.Getenv("AWS_ACCOUNT_ID")
}

func Handle(ctx context.Context, event events.DynamoDBEvent) error {
	for _, record := range event.Records {
		// Only process INSERT events (new executions)
		if record.EventName != "INSERT" {
			continue
		}

		// Extract execution_id and helper_type from DynamoDB Stream record
		executionID := record.Change.NewImage["execution_id"].String()
		helperType := record.Change.NewImage["helper_type"].String()

		if executionID == "" || helperType == "" {
			log.Printf("Missing execution_id or helper_type in record")
			continue
		}

		// Construct queue URL using naming convention
		queueURL := buildQueueURL(helperType)

		if queueURL == "" {
			// Fallback to old monolith queue during migration
			queueURL = os.Getenv("FALLBACK_QUEUE_URL")
			log.Printf("Cannot construct queue URL for helper_type=%s (missing account ID), using fallback", helperType)
		}

		if queueURL == "" {
			log.Printf("ERROR: No queue URL available for helper_type=%s and no fallback configured", helperType)
			return fmt.Errorf("no queue available for helper_type: %s", helperType)
		}

		// Convert the full DynamoDB Stream record to JSON for the worker.
		// The execution record contains all data frozen at execution time
		// (helper_type, connection_id, config, input, query_params, etc.)
		messageBody, err := streamImageToJSON(record.Change.NewImage)
		if err != nil {
			log.Printf("ERROR: Failed to convert stream image to JSON for execution %s: %v", executionID, err)
			return err
		}

		// Use first 8 chars of execution_id as MessageGroupId for FIFO queue ordering
		groupID := executionID
		if len(groupID) > 8 {
			groupID = groupID[:8]
		}

		_, err = sqsClient.SendMessage(ctx, &sqs.SendMessageInput{
			QueueUrl:       aws.String(queueURL),
			MessageBody:    aws.String(messageBody),
			MessageGroupId: aws.String(groupID),
		})

		if err != nil {
			log.Printf("ERROR: Failed to send message to queue %s for execution %s: %v", queueURL, executionID, err)
			return err
		}

		log.Printf("Routed execution %s (helper_type=%s) to queue %s", executionID, helperType, queueURL)
	}

	return nil
}

// buildQueueURL constructs an SQS queue URL from the helper type using a naming convention.
// helper_type "tag_it" → queue name "mfh-{stage}-tag-it-executions.fifo"
// → URL "https://sqs.{region}.amazonaws.com/{account}/mfh-{stage}-tag-it-executions.fifo"
func buildQueueURL(helperType string) string {
	if accountID == "" {
		return ""
	}

	// Convert helper_type from snake_case to kebab-case: "tag_it" → "tag-it"
	queueName := strings.ReplaceAll(helperType, "_", "-")

	return fmt.Sprintf("https://sqs.%s.amazonaws.com/%s/mfh-%s-%s-executions.fifo",
		region, accountID, stage, queueName)
}

// streamImageToJSON converts a DynamoDB Stream NewImage to a JSON string.
// This allows the full execution record to be forwarded to the worker via SQS.
func streamImageToJSON(image map[string]events.DynamoDBAttributeValue) (string, error) {
	m := convertStreamImage(image)
	bytes, err := json.Marshal(m)
	if err != nil {
		return "", fmt.Errorf("failed to marshal stream image: %w", err)
	}
	return string(bytes), nil
}

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
		n := v.Number()
		if i, err := strconv.ParseInt(n, 10, 64); err == nil {
			return i
		}
		if f, err := strconv.ParseFloat(n, 64); err == nil {
			return f
		}
		return n
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
	case events.DataTypeStringSet:
		return v.StringSet()
	case events.DataTypeNumberSet:
		return v.NumberSet()
	case events.DataTypeNull:
		return nil
	default:
		return nil
	}
}

func main() {
	lambda.Start(Handle)
}
