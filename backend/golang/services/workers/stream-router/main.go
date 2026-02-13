package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
)

var (
	sqsClient    *sqs.Client
	ssmClient    *ssm.Client
	queueCache   = make(map[string]string)
	cacheMutex   sync.RWMutex
	stage        string
)

func init() {
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		log.Fatalf("Failed to load AWS config: %v", err)
	}
	sqsClient = sqs.NewFromConfig(cfg)
	ssmClient = ssm.NewFromConfig(cfg)
	stage = os.Getenv("STAGE")
	if stage == "" {
		stage = "dev"
	}
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

		// Get queue URL for this helper type
		queueURL := getQueueForHelper(helperType)

		if queueURL == "" {
			// Passthrough mode: use fallback queue during migration
			queueURL = os.Getenv("FALLBACK_QUEUE_URL")
			log.Printf("No queue found for helper_type=%s, using fallback queue", helperType)
		}

		if queueURL == "" {
			log.Printf("ERROR: No queue URL available for helper_type=%s and no fallback configured", helperType)
			return fmt.Errorf("no queue available for helper_type: %s", helperType)
		}

		// Send message to SQS queue
		_, err := sqsClient.SendMessage(ctx, &sqs.SendMessageInput{
			QueueUrl:    aws.String(queueURL),
			MessageBody: aws.String(executionID),
			// Use first 8 chars of execution_id as MessageGroupId for FIFO queue ordering
			MessageGroupId: aws.String(executionID[:8]),
		})

		if err != nil {
			log.Printf("ERROR: Failed to send message to queue %s for execution %s: %v", queueURL, executionID, err)
			return err
		}

		log.Printf("Routed execution %s (helper_type=%s) to queue %s", executionID, helperType, queueURL)
	}

	return nil
}

func getQueueForHelper(helperType string) string {
	// Check cache first
	cacheMutex.RLock()
	if url, ok := queueCache[helperType]; ok {
		cacheMutex.RUnlock()
		return url
	}
	cacheMutex.RUnlock()

	// Try environment variable first (for initial deployment compatibility)
	envVar := strings.ToUpper(helperType) + "_QUEUE_URL"
	if url := os.Getenv(envVar); url != "" {
		cacheMutex.Lock()
		queueCache[helperType] = url
		cacheMutex.Unlock()
		return url
	}

	// Fetch from SSM Parameter Store
	// Parameter naming convention: /mfh/{stage}/sqs/{helper_type}/queue-url
	paramName := fmt.Sprintf("/mfh/%s/sqs/%s/queue-url", stage, helperType)

	ctx := context.Background()
	result, err := ssmClient.GetParameter(ctx, &ssm.GetParameterInput{
		Name: aws.String(paramName),
	})

	if err != nil {
		log.Printf("Failed to fetch queue URL from SSM for %s: %v", helperType, err)
		return ""
	}

	url := *result.Parameter.Value

	// Cache the result
	cacheMutex.Lock()
	queueCache[helperType] = url
	cacheMutex.Unlock()

	return url
}

func main() {
	lambda.Start(Handle)
}
