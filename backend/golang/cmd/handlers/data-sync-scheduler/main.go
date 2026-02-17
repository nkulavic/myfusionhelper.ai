package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	ddbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go-v2/service/sqs"

	apitypes "github.com/myfusionhelper/api/internal/types"
)

var (
	connectionsTable     = os.Getenv("CONNECTIONS_TABLE")
	syncExecutionsTable  = os.Getenv("SYNC_EXECUTIONS_TABLE")
	dataSyncQueueURL     = os.Getenv("DATA_SYNC_QUEUE_URL")
	continuationQueueURL = os.Getenv("DATA_SYNC_CONTINUATION_QUEUE_URL")
)

// stuckThreshold is how long a sync execution can be idle before the scheduler
// considers it stuck and resends a continuation message.
const stuckThreshold = 5 * time.Minute

// defaultSyncFrequency is used when a connection has no sync_frequency set.
const defaultSyncFrequency = "6h"

// SyncMessage matches the message format expected by the data-sync worker.
type SyncMessage struct {
	AccountID       string   `json:"account_id"`
	ConnectionID    string   `json:"connection_id"`
	ObjectTypes     []string `json:"object_types"`
	IsContinuation  bool     `json:"is_continuation"`
	ExecutionID     string   `json:"execution_id,omitempty"`
	ObjectTypeIndex int      `json:"object_type_index,omitempty"`
	Cursor          string   `json:"cursor,omitempty"`
	ChunkNumber     int      `json:"chunk_number,omitempty"`
	RecordsSoFar    int      `json:"records_so_far,omitempty"`
}

func main() {
	lambda.Start(handleScheduleEvent)
}

func handleScheduleEvent(ctx context.Context) error {
	log.Println("Data sync scheduler triggered")

	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(os.Getenv("COGNITO_REGION")))
	if err != nil {
		log.Printf("Failed to load AWS config: %v", err)
		return err
	}

	db := dynamodb.NewFromConfig(cfg)
	sqsClient := sqs.NewFromConfig(cfg)

	// 1. Recover stuck sync executions
	recoveredCount := recoverStuckExecutions(ctx, db, sqsClient)
	if recoveredCount > 0 {
		log.Printf("Recovered %d stuck sync executions", recoveredCount)
	}

	// 2. Schedule new syncs for due connections
	activeConnections, err := scanActiveConnections(ctx, db)
	if err != nil {
		log.Printf("Failed to scan active connections: %v", err)
		return err
	}

	log.Printf("Found %d active connections", len(activeConnections))
	now := time.Now().UTC()

	sentCount := 0
	skippedSyncing := 0
	skippedManual := 0
	skippedNotDue := 0

	for _, conn := range activeConnections {
		// Skip connections set to manual sync
		freq := conn.SyncFrequency
		if freq == "" {
			freq = defaultSyncFrequency
		}
		if freq == "manual" {
			skippedManual++
			continue
		}

		// Skip connections currently syncing
		if conn.SyncStatus == "syncing" {
			skippedSyncing++
			continue
		}

		// Check if sync is due (next_scheduled_sync <= now, or empty for first sync)
		if conn.NextScheduledSync != nil && conn.NextScheduledSync.After(now) {
			skippedNotDue++
			continue
		}

		msg := SyncMessage{
			AccountID:    conn.AccountID,
			ConnectionID: conn.ConnectionID,
			ObjectTypes:  []string{"contacts", "tags", "custom_fields"},
		}

		msgBody, err := json.Marshal(msg)
		if err != nil {
			log.Printf("Failed to marshal sync message for connection %s: %v", conn.ConnectionID, err)
			continue
		}

		_, err = sqsClient.SendMessage(ctx, &sqs.SendMessageInput{
			QueueUrl:    aws.String(dataSyncQueueURL),
			MessageBody: aws.String(string(msgBody)),
		})
		if err != nil {
			log.Printf("Failed to send sync message for connection %s: %v", conn.ConnectionID, err)
			continue
		}

		// Set next_scheduled_sync = now + sync_frequency
		nextSync := now.Add(parseSyncFrequency(freq))
		updateNextScheduledSync(ctx, db, conn.ConnectionID, nextSync)

		sentCount++
	}

	log.Printf("Scheduler results: sent=%d, skippedSyncing=%d, skippedManual=%d, skippedNotDue=%d, total=%d",
		sentCount, skippedSyncing, skippedManual, skippedNotDue, len(activeConnections))
	return nil
}

// parseSyncFrequency converts a frequency string to a duration.
func parseSyncFrequency(freq string) time.Duration {
	switch freq {
	case "1h":
		return 1 * time.Hour
	case "6h":
		return 6 * time.Hour
	case "12h":
		return 12 * time.Hour
	case "24h":
		return 24 * time.Hour
	default:
		return 6 * time.Hour
	}
}

// updateNextScheduledSync sets the next_scheduled_sync for a connection.
func updateNextScheduledSync(ctx context.Context, db *dynamodb.Client, connectionID string, nextSync time.Time) {
	_, err := db.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(connectionsTable),
		Key: map[string]ddbtypes.AttributeValue{
			"connection_id": &ddbtypes.AttributeValueMemberS{Value: connectionID},
		},
		UpdateExpression: aws.String("SET next_scheduled_sync = :next"),
		ExpressionAttributeValues: map[string]ddbtypes.AttributeValue{
			":next": &ddbtypes.AttributeValueMemberS{Value: nextSync.Format(time.RFC3339)},
		},
	})
	if err != nil {
		log.Printf("Failed to update next_scheduled_sync for connection %s: %v", connectionID, err)
	}
}

// recoverStuckExecutions queries the sync executions table for executions
// that have been in "syncing" status but idle for more than stuckThreshold.
// It resends continuation messages for these executions so they can resume.
func recoverStuckExecutions(ctx context.Context, db *dynamodb.Client, sqsClient *sqs.Client) int {
	if syncExecutionsTable == "" || continuationQueueURL == "" {
		return 0
	}

	cutoffTime := time.Now().UTC().Add(-stuckThreshold).Format(time.RFC3339)

	// Query the StatusIndex for syncing executions that haven't been updated recently
	result, err := db.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(syncExecutionsTable),
		IndexName:              aws.String("StatusIndex"),
		KeyConditionExpression: aws.String("#status = :syncing"),
		FilterExpression:       aws.String("updated_at < :cutoff"),
		ExpressionAttributeNames: map[string]string{
			"#status": "status",
		},
		ExpressionAttributeValues: map[string]ddbtypes.AttributeValue{
			":syncing": &ddbtypes.AttributeValueMemberS{Value: "syncing"},
			":cutoff":  &ddbtypes.AttributeValueMemberS{Value: cutoffTime},
		},
	})
	if err != nil {
		log.Printf("Failed to query stuck sync executions: %v", err)
		return 0
	}

	var stuckExecs []apitypes.SyncExecution
	if err := attributevalue.UnmarshalListOfMaps(result.Items, &stuckExecs); err != nil {
		log.Printf("Failed to unmarshal stuck executions: %v", err)
		return 0
	}

	recoveredCount := 0
	for _, exec := range stuckExecs {
		log.Printf("Recovering stuck sync execution %s (connection: %s, last updated: %s)",
			exec.ExecutionID, exec.ConnectionID, exec.UpdatedAt)

		msg := SyncMessage{
			AccountID:       exec.AccountID,
			ConnectionID:    exec.ConnectionID,
			ObjectTypes:     exec.ObjectTypes,
			IsContinuation:  true,
			ExecutionID:     exec.ExecutionID,
			ObjectTypeIndex: exec.ObjectTypeIndex,
			Cursor:          exec.Cursor,
			ChunkNumber:     exec.ChunkNumber,
			RecordsSoFar:    exec.RecordsSoFar,
		}

		msgBody, err := json.Marshal(msg)
		if err != nil {
			log.Printf("Failed to marshal recovery message for execution %s: %v", exec.ExecutionID, err)
			continue
		}

		_, err = sqsClient.SendMessage(ctx, &sqs.SendMessageInput{
			QueueUrl:    aws.String(continuationQueueURL),
			MessageBody: aws.String(string(msgBody)),
		})
		if err != nil {
			log.Printf("Failed to send recovery message for execution %s: %v", exec.ExecutionID, err)
			continue
		}

		// Update the execution's updated_at so we don't immediately re-recover it
		_, updateErr := db.UpdateItem(ctx, &dynamodb.UpdateItemInput{
			TableName: aws.String(syncExecutionsTable),
			Key: map[string]ddbtypes.AttributeValue{
				"execution_id": &ddbtypes.AttributeValueMemberS{Value: exec.ExecutionID},
			},
			UpdateExpression: aws.String("SET updated_at = :now"),
			ExpressionAttributeValues: map[string]ddbtypes.AttributeValue{
				":now": &ddbtypes.AttributeValueMemberS{Value: time.Now().UTC().Format(time.RFC3339)},
			},
		})
		if updateErr != nil {
			log.Printf("Failed to update stuck execution %s timestamp: %v", exec.ExecutionID, updateErr)
		}

		recoveredCount++
	}

	return recoveredCount
}

// scanActiveConnections scans the connections table for all connections with status="active".
func scanActiveConnections(ctx context.Context, db *dynamodb.Client) ([]apitypes.PlatformConnection, error) {
	var connections []apitypes.PlatformConnection
	var lastEvaluatedKey map[string]ddbtypes.AttributeValue

	for {
		input := &dynamodb.ScanInput{
			TableName:        aws.String(connectionsTable),
			FilterExpression: aws.String("#s = :active"),
			ExpressionAttributeNames: map[string]string{
				"#s": "status",
			},
			ExpressionAttributeValues: map[string]ddbtypes.AttributeValue{
				":active": &ddbtypes.AttributeValueMemberS{Value: "active"},
			},
		}
		if lastEvaluatedKey != nil {
			input.ExclusiveStartKey = lastEvaluatedKey
		}

		result, err := db.Scan(ctx, input)
		if err != nil {
			return nil, err
		}

		var page []apitypes.PlatformConnection
		if err := attributevalue.UnmarshalListOfMaps(result.Items, &page); err != nil {
			return nil, err
		}
		connections = append(connections, page...)

		if result.LastEvaluatedKey == nil {
			break
		}
		lastEvaluatedKey = result.LastEvaluatedKey
	}

	return connections, nil
}
