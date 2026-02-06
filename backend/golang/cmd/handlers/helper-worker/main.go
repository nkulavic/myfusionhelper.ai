package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	ddbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	"github.com/myfusionhelper/api/internal/connectors"
	"github.com/myfusionhelper/api/internal/connectors/loader"
	helperEngine "github.com/myfusionhelper/api/internal/helpers"

	// Register all connectors via init()
	_ "github.com/myfusionhelper/api/internal/connectors"

	// Register all helpers via init()
	_ "github.com/myfusionhelper/api/internal/helpers/analytics"
	_ "github.com/myfusionhelper/api/internal/helpers/automation"
	_ "github.com/myfusionhelper/api/internal/helpers/contact"
	_ "github.com/myfusionhelper/api/internal/helpers/data"
	_ "github.com/myfusionhelper/api/internal/helpers/integration"
	_ "github.com/myfusionhelper/api/internal/helpers/notification"
	_ "github.com/myfusionhelper/api/internal/helpers/tagging"
)

var (
	executionsTable      = os.Getenv("EXECUTIONS_TABLE")
	helpersTable         = os.Getenv("HELPERS_TABLE")
	notificationQueueURL = os.Getenv("NOTIFICATION_QUEUE_URL")
)

// HelperExecutionJob represents a job from the SQS queue
type HelperExecutionJob struct {
	ExecutionID  string                 `json:"execution_id"`
	HelperID     string                 `json:"helper_id"`
	HelperType   string                 `json:"helper_type"`
	AccountID    string                 `json:"account_id"`
	UserID       string                 `json:"user_id"`
	ConnectionID string                 `json:"connection_id"`
	ContactID    string                 `json:"contact_id"`
	Config       map[string]interface{} `json:"config"`
	Input        map[string]interface{} `json:"input"`
	RetryCount   int                    `json:"retry_count"`
}

func main() {
	lambda.Start(handleSQSEvent)
}

func handleSQSEvent(ctx context.Context, event events.SQSEvent) error {
	log.Printf("Processing %d SQS messages", len(event.Records))

	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(os.Getenv("COGNITO_REGION")))
	if err != nil {
		log.Printf("Failed to load AWS config: %v", err)
		return err
	}
	db := dynamodb.NewFromConfig(cfg)
	sqsClient := sqs.NewFromConfig(cfg)

	for _, record := range event.Records {
		var job HelperExecutionJob
		if err := json.Unmarshal([]byte(record.Body), &job); err != nil {
			log.Printf("Failed to unmarshal SQS message: %v", err)
			continue
		}

		log.Printf("Processing execution %s (helper: %s, type: %s)", job.ExecutionID, job.HelperID, job.HelperType)

		// Update execution status to running
		updateExecutionStatus(ctx, db, job.ExecutionID, "running", "", 0)

		// Execute the helper
		result, execErr := processJob(ctx, db, job)

		// Update execution record with results
		now := time.Now().UTC()
		if execErr != nil {
			log.Printf("Execution %s failed: %v", job.ExecutionID, execErr)
			updateExecutionResult(ctx, db, job.ExecutionID, "failed", execErr.Error(), result, &now)
			sendFailureNotification(ctx, sqsClient, job, execErr.Error())
		} else if result != nil && result.Success {
			log.Printf("Execution %s completed successfully", job.ExecutionID)
			updateExecutionResult(ctx, db, job.ExecutionID, "completed", "", result, &now)
		} else {
			errMsg := "execution returned unsuccessful result"
			if result != nil && result.Error != "" {
				errMsg = result.Error
			}
			log.Printf("Execution %s completed with errors: %s", job.ExecutionID, errMsg)
			updateExecutionResult(ctx, db, job.ExecutionID, "failed", errMsg, result, &now)
			sendFailureNotification(ctx, sqsClient, job, errMsg)
		}

		// Update helper execution count
		updateHelperStats(ctx, db, job.HelperID, &now)
	}

	return nil
}

func processJob(ctx context.Context, db *dynamodb.Client, job HelperExecutionJob) (*helperEngine.ExecutionResult, error) {
	// Load CRM connector if connection ID is specified
	var connector connectors.CRMConnector
	if job.ConnectionID != "" {
		var err error
		connector, err = loader.LoadConnectorWithTranslation(ctx, db, job.ConnectionID, job.AccountID)
		if err != nil {
			return nil, err
		}
	}

	// Execute via the helper engine
	executor := helperEngine.NewExecutor()
	execReq := helperEngine.ExecutionRequest{
		HelperType:   job.HelperType,
		ContactID:    job.ContactID,
		Config:       job.Config,
		UserID:       job.UserID,
		AccountID:    job.AccountID,
		HelperID:     job.HelperID,
		ConnectionID: job.ConnectionID,
	}

	result, err := executor.Execute(ctx, execReq, connector)
	return result, err
}

func updateExecutionStatus(ctx context.Context, db *dynamodb.Client, executionID, status, errorMsg string, durationMs int64) {
	updateExpr := "SET #s = :status"
	exprNames := map[string]string{"#s": "status"}
	exprValues := map[string]ddbtypes.AttributeValue{
		":status": &ddbtypes.AttributeValueMemberS{Value: status},
	}

	if errorMsg != "" {
		updateExpr += ", error_message = :error"
		exprValues[":error"] = &ddbtypes.AttributeValueMemberS{Value: errorMsg}
	}
	if durationMs > 0 {
		updateExpr += ", duration_ms = :duration"
		exprValues[":duration"] = &ddbtypes.AttributeValueMemberN{Value: fmt.Sprintf("%d", durationMs)}
	}

	_, err := db.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(executionsTable),
		Key: map[string]ddbtypes.AttributeValue{
			"execution_id": &ddbtypes.AttributeValueMemberS{Value: executionID},
		},
		UpdateExpression:          aws.String(updateExpr),
		ExpressionAttributeNames:  exprNames,
		ExpressionAttributeValues: exprValues,
	})
	if err != nil {
		log.Printf("Failed to update execution status: %v", err)
	}
}

func updateExecutionResult(ctx context.Context, db *dynamodb.Client, executionID, status, errorMsg string, result *helperEngine.ExecutionResult, completedAt *time.Time) {
	updateExpr := "SET #s = :status, completed_at = :completed_at"
	exprNames := map[string]string{"#s": "status"}
	exprValues := map[string]ddbtypes.AttributeValue{
		":status":       &ddbtypes.AttributeValueMemberS{Value: status},
		":completed_at": &ddbtypes.AttributeValueMemberS{Value: completedAt.Format(time.RFC3339)},
	}

	if errorMsg != "" {
		updateExpr += ", error_message = :error"
		exprValues[":error"] = &ddbtypes.AttributeValueMemberS{Value: errorMsg}
	}

	if result != nil {
		updateExpr += ", duration_ms = :duration"
		exprValues[":duration"] = &ddbtypes.AttributeValueMemberN{Value: fmt.Sprintf("%d", result.DurationMs)}

		if result.Output != nil {
			outputJSON, err := json.Marshal(result.Output)
			if err == nil {
				updateExpr += ", output = :output"
				exprValues[":output"] = &ddbtypes.AttributeValueMemberS{Value: string(outputJSON)}
			}
		}
	}

	_, err := db.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(executionsTable),
		Key: map[string]ddbtypes.AttributeValue{
			"execution_id": &ddbtypes.AttributeValueMemberS{Value: executionID},
		},
		UpdateExpression:          aws.String(updateExpr),
		ExpressionAttributeNames:  exprNames,
		ExpressionAttributeValues: exprValues,
	})
	if err != nil {
		log.Printf("Failed to update execution result: %v", err)
	}
}

func sendFailureNotification(ctx context.Context, sqsClient *sqs.Client, job HelperExecutionJob, errorMsg string) {
	if notificationQueueURL == "" {
		log.Printf("NOTIFICATION_QUEUE_URL not set, skipping failure notification")
		return
	}

	notification := map[string]interface{}{
		"type":       "execution_failure",
		"user_id":    job.UserID,
		"account_id": job.AccountID,
		"data": map[string]interface{}{
			"helper_name":   job.HelperType,
			"error_message": errorMsg,
			"execution_id":  job.ExecutionID,
			"helper_id":     job.HelperID,
		},
	}

	body, err := json.Marshal(notification)
	if err != nil {
		log.Printf("Failed to marshal notification: %v", err)
		return
	}

	_, err = sqsClient.SendMessage(ctx, &sqs.SendMessageInput{
		QueueUrl:       aws.String(notificationQueueURL),
		MessageBody:    aws.String(string(body)),
		MessageGroupId: aws.String(job.AccountID),
	})
	if err != nil {
		log.Printf("Failed to send notification to SQS: %v", err)
	} else {
		log.Printf("Sent failure notification for execution %s", job.ExecutionID)
	}
}

func updateHelperStats(ctx context.Context, db *dynamodb.Client, helperID string, executedAt *time.Time) {
	_, err := db.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(helpersTable),
		Key: map[string]ddbtypes.AttributeValue{
			"helper_id": &ddbtypes.AttributeValueMemberS{Value: helperID},
		},
		UpdateExpression: aws.String("SET execution_count = if_not_exists(execution_count, :zero) + :one, last_executed_at = :executed_at"),
		ExpressionAttributeValues: map[string]ddbtypes.AttributeValue{
			":zero":        &ddbtypes.AttributeValueMemberN{Value: "0"},
			":one":         &ddbtypes.AttributeValueMemberN{Value: "1"},
			":executed_at": &ddbtypes.AttributeValueMemberS{Value: executedAt.Format(time.RFC3339)},
		},
	})
	if err != nil {
		log.Printf("Failed to update helper stats: %v", err)
	}
}
