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
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	ddbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/google/uuid"

	"github.com/myfusionhelper/api/internal/connectors"
	"github.com/myfusionhelper/api/internal/connectors/loader"
	"github.com/myfusionhelper/api/internal/services/parquet"
	apitypes "github.com/myfusionhelper/api/internal/types"

	// Register all connectors via init()
	_ "github.com/myfusionhelper/api/internal/connectors"
)

const (
	// chunkSize is the number of records per chunk parquet file.
	chunkSize = 10000
	// maxDuration is how long we run before triggering a continuation.
	// Lambda timeout is 900s (15 min); we stop at 13 min to flush + send continuation.
	maxDuration = 13 * time.Minute
	// pageSize is the number of records per CRM API call.
	pageSize = 200
	// executionTTLDays is how long sync execution records are kept.
	executionTTLDays = 7
)

var (
	connectionsTable     = os.Getenv("CONNECTIONS_TABLE")
	syncExecutionsTable  = os.Getenv("SYNC_EXECUTIONS_TABLE")
	analyticsBucket      = os.Getenv("ANALYTICS_BUCKET")
	continuationQueueURL = os.Getenv("DATA_SYNC_CONTINUATION_QUEUE_URL")
)

// SyncMessage represents a data sync job received from SQS.
type SyncMessage struct {
	AccountID    string   `json:"account_id"`
	ConnectionID string   `json:"connection_id"`
	ObjectTypes  []string `json:"object_types"`

	// Continuation fields (populated when IsContinuation=true)
	IsContinuation  bool   `json:"is_continuation"`
	ExecutionID     string `json:"execution_id,omitempty"`
	ObjectTypeIndex int    `json:"object_type_index,omitempty"`
	Cursor          string `json:"cursor,omitempty"`
	ChunkNumber     int    `json:"chunk_number,omitempty"`
	RecordsSoFar    int    `json:"records_so_far,omitempty"`
}

func main() {
	lambda.Start(handleSQSEvent)
}

func handleSQSEvent(ctx context.Context, event events.SQSEvent) error {
	log.Printf("Processing %d SQS messages for data sync", len(event.Records))

	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(os.Getenv("COGNITO_REGION")))
	if err != nil {
		log.Printf("Failed to load AWS config: %v", err)
		return err
	}

	db := dynamodb.NewFromConfig(cfg)
	s3Client := s3.NewFromConfig(cfg)
	sqsClient := sqs.NewFromConfig(cfg)

	for _, record := range event.Records {
		var msg SyncMessage
		if err := json.Unmarshal([]byte(record.Body), &msg); err != nil {
			log.Printf("Failed to unmarshal SQS message: %v", err)
			continue
		}

		log.Printf("Starting data sync for connection %s (account: %s, objects: %v, continuation: %v)",
			msg.ConnectionID, msg.AccountID, msg.ObjectTypes, msg.IsContinuation)

		if err := processSync(ctx, db, s3Client, sqsClient, msg); err != nil {
			log.Printf("Sync failed for connection %s: %v", msg.ConnectionID, err)
			updateSyncStatus(ctx, db, msg.ConnectionID, "failed", nil, nil)
			if msg.ExecutionID != "" {
				failSyncExecution(ctx, db, msg.ExecutionID, err.Error())
			}
		}
	}

	return nil
}

func processSync(ctx context.Context, db *dynamodb.Client, s3Client *s3.Client, sqsClient *sqs.Client, msg SyncMessage) error {
	startTime := time.Now()

	// Load the CRM connector with field translation
	connector, err := loader.LoadConnectorWithTranslation(ctx, db, msg.ConnectionID, msg.AccountID)
	if err != nil {
		return fmt.Errorf("failed to load connector: %w", err)
	}

	// Determine which capabilities the connector supports
	capabilities := connector.GetCapabilities()
	capSet := make(map[connectors.Capability]bool, len(capabilities))
	for _, cap := range capabilities {
		capSet[cap] = true
	}

	// Create or resume sync execution
	var exec apitypes.SyncExecution
	if !msg.IsContinuation {
		// New sync: mark connection as syncing + create execution record
		updateSyncStatus(ctx, db, msg.ConnectionID, "syncing", nil, nil)

		exec = apitypes.SyncExecution{
			ExecutionID:          uuid.New().String(),
			ConnectionID:         msg.ConnectionID,
			AccountID:            msg.AccountID,
			Status:               "syncing",
			StartedAt:            startTime.UTC().Format(time.RFC3339),
			ObjectTypes:          msg.ObjectTypes,
			ObjectTypeIndex:      0,
			ChunkNumber:          0,
			RecordsSoFar:         0,
			CompletedObjectTypes: make(map[string]int),
			UpdatedAt:            startTime.UTC().Format(time.RFC3339),
			TTL:                  startTime.Add(time.Duration(executionTTLDays) * 24 * time.Hour).Unix(),
		}
		if err := putSyncExecution(ctx, db, &exec); err != nil {
			return fmt.Errorf("failed to create sync execution: %w", err)
		}

		// Clear old chunk files for all object types that will be synced
		for _, ot := range msg.ObjectTypes {
			clearOldChunks(ctx, s3Client, msg.AccountID, msg.ConnectionID, ot)
		}
	} else {
		// Continuation: resume from saved state
		var err error
		exec, err = getSyncExecution(ctx, db, msg.ExecutionID)
		if err != nil {
			return fmt.Errorf("failed to get sync execution %s: %w", msg.ExecutionID, err)
		}
		exec.ObjectTypes = msg.ObjectTypes
	}

	recordCounts := make(map[string]int)
	// Copy over previously completed counts
	for k, v := range exec.CompletedObjectTypes {
		recordCounts[k] = v
	}

	startIdx := 0
	if msg.IsContinuation {
		startIdx = msg.ObjectTypeIndex
	}

	for i := startIdx; i < len(msg.ObjectTypes); i++ {
		objectType := msg.ObjectTypes[i]

		// Check capability
		switch objectType {
		case "contacts":
			if !capSet[connectors.CapContacts] {
				log.Printf("Connector does not support contacts, skipping")
				continue
			}
		case "tags":
			if !capSet[connectors.CapTags] {
				log.Printf("Connector does not support tags, skipping")
				continue
			}
		case "custom_fields":
			if !capSet[connectors.CapCustomFields] {
				log.Printf("Connector does not support custom_fields, skipping")
				continue
			}
		default:
			log.Printf("Unknown object type %q, skipping", objectType)
			continue
		}

		cursor := ""
		if i == startIdx && msg.IsContinuation {
			cursor = msg.Cursor
		}

		chunkNum := exec.ChunkNumber
		if i != startIdx || !msg.IsContinuation {
			chunkNum = 0 // Reset chunk numbering for new object types
		}

		count, newChunkNum, timedOut, err := syncObjectType(
			ctx, connector, s3Client, msg.AccountID, msg.ConnectionID,
			objectType, cursor, chunkNum, startTime, capSet,
		)
		if err != nil {
			return fmt.Errorf("%s sync failed: %w", objectType, err)
		}

		exec.RecordsSoFar += count
		recordCounts[objectType] = (recordCounts[objectType]) + count

		if timedOut {
			// Save state and send continuation
			exec.ChunkNumber = newChunkNum
			exec.ObjectTypeIndex = i
			exec.CurrentObjectType = objectType
			exec.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
			if err := updateSyncExecution(ctx, db, &exec); err != nil {
				log.Printf("Failed to save sync execution state: %v", err)
			}

			contMsg := SyncMessage{
				AccountID:       msg.AccountID,
				ConnectionID:    msg.ConnectionID,
				ObjectTypes:     msg.ObjectTypes,
				IsContinuation:  true,
				ExecutionID:     exec.ExecutionID,
				ObjectTypeIndex: i,
				Cursor:          exec.Cursor,
				ChunkNumber:     newChunkNum,
				RecordsSoFar:    exec.RecordsSoFar,
			}
			if err := sendContinuation(ctx, sqsClient, contMsg); err != nil {
				return fmt.Errorf("failed to send continuation: %w", err)
			}

			log.Printf("Continuation queued for connection %s (execution: %s, objectType: %s, chunk: %d, records: %d)",
				msg.ConnectionID, exec.ExecutionID, objectType, newChunkNum, exec.RecordsSoFar)
			return nil // Exit without marking completed
		}

		// Object type fully synced — update completed tracking
		exec.CompletedObjectTypes[objectType] = recordCounts[objectType]
		exec.ChunkNumber = 0 // Reset for next object type

		// Write schema.json for this object type
		// (schema is written after each object type completes, even mid-continuation)
	}

	// All object types completed
	now := time.Now().UTC()
	updateSyncStatus(ctx, db, msg.ConnectionID, "completed", &now, recordCounts)

	exec.Status = "completed"
	exec.CompletedAt = now.Format(time.RFC3339)
	exec.UpdatedAt = now.Format(time.RFC3339)
	if err := updateSyncExecution(ctx, db, &exec); err != nil {
		log.Printf("Failed to update sync execution as completed: %v", err)
	}

	log.Printf("Data sync completed for connection %s (execution: %s): %v",
		msg.ConnectionID, exec.ExecutionID, recordCounts)
	return nil
}

// syncObjectType syncs a single object type with chunked parquet output and time-limit checks.
// Returns (recordCount, chunkNumber, timedOut, error).
func syncObjectType(
	ctx context.Context,
	connector connectors.CRMConnector,
	s3Client *s3.Client,
	accountID, connectionID, objectType, startCursor string,
	startChunkNum int,
	startTime time.Time,
	capSet map[connectors.Capability]bool,
) (int, int, bool, error) {
	switch objectType {
	case "contacts":
		return syncContactsChunked(ctx, connector, s3Client, accountID, connectionID, startCursor, startChunkNum, startTime)
	case "tags":
		count, err := syncTags(ctx, connector, s3Client, accountID, connectionID, startChunkNum)
		return count, startChunkNum + 1, false, err
	case "custom_fields":
		count, err := syncCustomFields(ctx, connector, s3Client, accountID, connectionID, startChunkNum)
		return count, startChunkNum + 1, false, err
	default:
		return 0, startChunkNum, false, nil
	}
}

// syncContactsChunked fetches contacts page-by-page, writing chunk parquet files every chunkSize records.
// If the Lambda is running out of time, it returns timedOut=true with the current cursor saved in the execution.
func syncContactsChunked(
	ctx context.Context,
	connector connectors.CRMConnector,
	s3Client *s3.Client,
	accountID, connectionID, startCursor string,
	startChunkNum int,
	startTime time.Time,
) (int, int, bool, error) {
	var batch []connectors.NormalizedContact
	cursor := startCursor
	chunkNum := startChunkNum
	totalCount := 0

	for {
		opts := connectors.QueryOptions{
			Limit:  pageSize,
			Cursor: cursor,
		}

		result, err := connector.GetContacts(ctx, opts)
		if err != nil {
			// If we have accumulated records, flush them before returning error
			if len(batch) > 0 {
				s3Key := fmt.Sprintf("%s/%s/contacts/chunk_%03d.parquet", accountID, connectionID, chunkNum)
				if _, flushErr := parquet.WriteContactsParquet(ctx, s3Client, analyticsBucket, s3Key, batch); flushErr != nil {
					log.Printf("Failed to flush contacts chunk on error: %v", flushErr)
				}
			}
			return totalCount, chunkNum, false, fmt.Errorf("failed to get contacts page: %w", err)
		}

		batch = append(batch, result.Contacts...)
		totalCount += len(result.Contacts)
		log.Printf("Fetched %d contacts (total so far: %d, has_more: %v)",
			len(result.Contacts), totalCount, result.HasMore)

		// Flush chunk when batch is big enough
		if len(batch) >= chunkSize {
			s3Key := fmt.Sprintf("%s/%s/contacts/chunk_%03d.parquet", accountID, connectionID, chunkNum)
			schemaInfo, err := parquet.WriteContactsParquet(ctx, s3Client, analyticsBucket, s3Key, batch)
			if err != nil {
				return totalCount, chunkNum, false, fmt.Errorf("failed to write contacts chunk %d: %w", chunkNum, err)
			}
			// Write schema alongside chunk
			schemaInfo.ConnectionID = connectionID
			schemaKey := fmt.Sprintf("%s/%s/contacts/schema.json", accountID, connectionID)
			if err := parquet.WriteSchema(ctx, s3Client, analyticsBucket, schemaKey, schemaInfo); err != nil {
				log.Printf("Failed to write contacts schema: %v", err)
			}
			chunkNum++
			batch = batch[:0] // Reset
		}

		// Check time limit before next page
		if time.Since(startTime) > maxDuration {
			// Flush remaining records
			if len(batch) > 0 {
				s3Key := fmt.Sprintf("%s/%s/contacts/chunk_%03d.parquet", accountID, connectionID, chunkNum)
				schemaInfo, err := parquet.WriteContactsParquet(ctx, s3Client, analyticsBucket, s3Key, batch)
				if err != nil {
					return totalCount, chunkNum, false, fmt.Errorf("failed to flush contacts chunk %d on timeout: %w", chunkNum, err)
				}
				schemaInfo.ConnectionID = connectionID
				schemaKey := fmt.Sprintf("%s/%s/contacts/schema.json", accountID, connectionID)
				if err := parquet.WriteSchema(ctx, s3Client, analyticsBucket, schemaKey, schemaInfo); err != nil {
					log.Printf("Failed to write contacts schema: %v", err)
				}
				chunkNum++
			}
			log.Printf("Time limit reached after %v, saving cursor for continuation", time.Since(startTime))
			return totalCount, chunkNum, true, nil
		}

		if !result.HasMore || result.NextCursor == "" {
			break
		}
		cursor = result.NextCursor
	}

	// Flush remaining records
	if len(batch) > 0 {
		s3Key := fmt.Sprintf("%s/%s/contacts/chunk_%03d.parquet", accountID, connectionID, chunkNum)
		schemaInfo, err := parquet.WriteContactsParquet(ctx, s3Client, analyticsBucket, s3Key, batch)
		if err != nil {
			return totalCount, chunkNum, false, fmt.Errorf("failed to write final contacts chunk %d: %w", chunkNum, err)
		}
		schemaInfo.ConnectionID = connectionID
		schemaKey := fmt.Sprintf("%s/%s/contacts/schema.json", accountID, connectionID)
		if err := parquet.WriteSchema(ctx, s3Client, analyticsBucket, schemaKey, schemaInfo); err != nil {
			log.Printf("Failed to write contacts schema: %v", err)
		}
		chunkNum++
	}

	return totalCount, chunkNum, false, nil
}

func syncTags(ctx context.Context, connector connectors.CRMConnector, s3Client *s3.Client, accountID, connectionID string, chunkNum int) (int, error) {
	tags, err := connector.GetTags(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to get tags: %w", err)
	}

	s3Key := fmt.Sprintf("%s/%s/tags/chunk_%03d.parquet", accountID, connectionID, chunkNum)
	schemaInfo, err := parquet.WriteTagsParquet(ctx, s3Client, analyticsBucket, s3Key, tags)
	if err != nil {
		return 0, fmt.Errorf("failed to write tags parquet: %w", err)
	}

	schemaInfo.ConnectionID = connectionID
	schemaKey := fmt.Sprintf("%s/%s/tags/schema.json", accountID, connectionID)
	if err := parquet.WriteSchema(ctx, s3Client, analyticsBucket, schemaKey, schemaInfo); err != nil {
		return 0, fmt.Errorf("failed to write tags schema: %w", err)
	}

	return len(tags), nil
}

func syncCustomFields(ctx context.Context, connector connectors.CRMConnector, s3Client *s3.Client, accountID, connectionID string, chunkNum int) (int, error) {
	fields, err := connector.GetCustomFields(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to get custom fields: %w", err)
	}

	s3Key := fmt.Sprintf("%s/%s/custom_fields/chunk_%03d.parquet", accountID, connectionID, chunkNum)
	schemaInfo, err := parquet.WriteCustomFieldsParquet(ctx, s3Client, analyticsBucket, s3Key, fields)
	if err != nil {
		return 0, fmt.Errorf("failed to write custom_fields parquet: %w", err)
	}

	schemaInfo.ConnectionID = connectionID
	schemaKey := fmt.Sprintf("%s/%s/custom_fields/schema.json", accountID, connectionID)
	if err := parquet.WriteSchema(ctx, s3Client, analyticsBucket, schemaKey, schemaInfo); err != nil {
		return 0, fmt.Errorf("failed to write custom_fields schema: %w", err)
	}

	return len(fields), nil
}

// clearOldChunks deletes all existing chunk_*.parquet files for an object type
// before writing fresh data. Also deletes the old data.parquet if present.
func clearOldChunks(ctx context.Context, s3Client *s3.Client, accountID, connectionID, objectType string) {
	prefix := fmt.Sprintf("%s/%s/%s/", accountID, connectionID, objectType)

	paginator := s3.NewListObjectsV2Paginator(s3Client, &s3.ListObjectsV2Input{
		Bucket: aws.String(analyticsBucket),
		Prefix: aws.String(prefix),
	})

	var objectsToDelete []s3types.ObjectIdentifier
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			log.Printf("Failed to list objects for cleanup (prefix: %s): %v", prefix, err)
			return
		}
		for _, obj := range page.Contents {
			objectsToDelete = append(objectsToDelete, s3types.ObjectIdentifier{
				Key: obj.Key,
			})
		}
	}

	if len(objectsToDelete) == 0 {
		return
	}

	// Delete in batches of 1000 (S3 limit)
	for i := 0; i < len(objectsToDelete); i += 1000 {
		end := i + 1000
		if end > len(objectsToDelete) {
			end = len(objectsToDelete)
		}
		_, err := s3Client.DeleteObjects(ctx, &s3.DeleteObjectsInput{
			Bucket: aws.String(analyticsBucket),
			Delete: &s3types.Delete{
				Objects: objectsToDelete[i:end],
				Quiet:   aws.Bool(true),
			},
		})
		if err != nil {
			log.Printf("Failed to delete old chunks (prefix: %s): %v", prefix, err)
		}
	}

	log.Printf("Cleared %d old files for %s/%s/%s", len(objectsToDelete), accountID, connectionID, objectType)
}

// sendContinuation sends a continuation message to the continuation SQS queue.
func sendContinuation(ctx context.Context, sqsClient *sqs.Client, msg SyncMessage) error {
	body, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal continuation message: %w", err)
	}

	_, err = sqsClient.SendMessage(ctx, &sqs.SendMessageInput{
		QueueUrl:    aws.String(continuationQueueURL),
		MessageBody: aws.String(string(body)),
	})
	return err
}

// ========== DynamoDB helpers for SyncExecution ==========

func putSyncExecution(ctx context.Context, db *dynamodb.Client, exec *apitypes.SyncExecution) error {
	item, err := attributevalue.MarshalMap(exec)
	if err != nil {
		return err
	}
	_, err = db.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(syncExecutionsTable),
		Item:      item,
	})
	return err
}

func getSyncExecution(ctx context.Context, db *dynamodb.Client, executionID string) (apitypes.SyncExecution, error) {
	result, err := db.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(syncExecutionsTable),
		Key: map[string]ddbtypes.AttributeValue{
			"execution_id": &ddbtypes.AttributeValueMemberS{Value: executionID},
		},
	})
	if err != nil {
		return apitypes.SyncExecution{}, err
	}
	if result.Item == nil {
		return apitypes.SyncExecution{}, fmt.Errorf("sync execution %s not found", executionID)
	}
	var exec apitypes.SyncExecution
	if err := attributevalue.UnmarshalMap(result.Item, &exec); err != nil {
		return apitypes.SyncExecution{}, err
	}
	return exec, nil
}

func updateSyncExecution(ctx context.Context, db *dynamodb.Client, exec *apitypes.SyncExecution) error {
	exec.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
	return putSyncExecution(ctx, db, exec)
}

func failSyncExecution(ctx context.Context, db *dynamodb.Client, executionID, errMsg string) {
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := db.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(syncExecutionsTable),
		Key: map[string]ddbtypes.AttributeValue{
			"execution_id": &ddbtypes.AttributeValueMemberS{Value: executionID},
		},
		UpdateExpression: aws.String("SET #status = :status, error_message = :err, updated_at = :updated"),
		ExpressionAttributeNames: map[string]string{
			"#status": "status",
		},
		ExpressionAttributeValues: map[string]ddbtypes.AttributeValue{
			":status":  &ddbtypes.AttributeValueMemberS{Value: "failed"},
			":err":     &ddbtypes.AttributeValueMemberS{Value: errMsg},
			":updated": &ddbtypes.AttributeValueMemberS{Value: now},
		},
	})
	if err != nil {
		log.Printf("Failed to mark sync execution %s as failed: %v", executionID, err)
	}
}

// updateSyncStatus updates the connection's sync status, last_synced_at, and record counts in DynamoDB.
func updateSyncStatus(ctx context.Context, db *dynamodb.Client, connectionID, status string, syncedAt *time.Time, recordCounts map[string]int) {
	updateExpr := "SET sync_status = :status, updated_at = :updated_at"
	exprValues := map[string]ddbtypes.AttributeValue{
		":status":     &ddbtypes.AttributeValueMemberS{Value: status},
		":updated_at": &ddbtypes.AttributeValueMemberS{Value: time.Now().UTC().Format(time.RFC3339)},
	}

	if syncedAt != nil {
		updateExpr += ", last_synced_at = :synced_at"
		exprValues[":synced_at"] = &ddbtypes.AttributeValueMemberS{Value: syncedAt.Format(time.RFC3339)}
	}

	if recordCounts != nil {
		countsAV, err := attributevalue.MarshalMap(recordCounts)
		if err == nil {
			updateExpr += ", sync_record_counts = :counts"
			exprValues[":counts"] = &ddbtypes.AttributeValueMemberM{Value: countsAV}
		}
	}

	_, err := db.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(connectionsTable),
		Key: map[string]ddbtypes.AttributeValue{
			"connection_id": &ddbtypes.AttributeValueMemberS{Value: connectionID},
		},
		UpdateExpression:          aws.String(updateExpr),
		ExpressionAttributeValues: exprValues,
	})
	if err != nil {
		log.Printf("Failed to update sync status for connection %s: %v", connectionID, err)
	}
}
