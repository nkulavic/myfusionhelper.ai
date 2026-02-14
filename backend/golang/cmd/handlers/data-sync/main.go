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

	"github.com/myfusionhelper/api/internal/connectors"
	"github.com/myfusionhelper/api/internal/connectors/loader"
	"github.com/myfusionhelper/api/internal/services/parquet"

	// Register all connectors via init()
	_ "github.com/myfusionhelper/api/internal/connectors"
)

var (
	connectionsTable = os.Getenv("CONNECTIONS_TABLE")
	analyticsBucket  = os.Getenv("ANALYTICS_BUCKET")
)

// SyncMessage represents a data sync job received from SQS.
type SyncMessage struct {
	AccountID    string   `json:"account_id"`
	ConnectionID string   `json:"connection_id"`
	ObjectTypes  []string `json:"object_types"`
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

	for _, record := range event.Records {
		var msg SyncMessage
		if err := json.Unmarshal([]byte(record.Body), &msg); err != nil {
			log.Printf("Failed to unmarshal SQS message: %v", err)
			continue
		}

		log.Printf("Starting data sync for connection %s (account: %s, objects: %v)",
			msg.ConnectionID, msg.AccountID, msg.ObjectTypes)

		if err := processSync(ctx, db, s3Client, msg); err != nil {
			log.Printf("Sync failed for connection %s: %v", msg.ConnectionID, err)
			updateSyncStatus(ctx, db, msg.ConnectionID, "failed", nil, nil)
		}
	}

	return nil
}

func processSync(ctx context.Context, db *dynamodb.Client, s3Client *s3.Client, msg SyncMessage) error {
	// Mark connection as syncing
	updateSyncStatus(ctx, db, msg.ConnectionID, "syncing", nil, nil)

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

	recordCounts := make(map[string]int)
	now := time.Now().UTC()

	for _, objectType := range msg.ObjectTypes {
		switch objectType {
		case "contacts":
			if !capSet[connectors.CapContacts] {
				log.Printf("Connector does not support contacts, skipping")
				continue
			}
			count, err := syncContacts(ctx, connector, s3Client, msg.AccountID, msg.ConnectionID)
			if err != nil {
				return fmt.Errorf("contacts sync failed: %w", err)
			}
			recordCounts["contacts"] = count

		case "tags":
			if !capSet[connectors.CapTags] {
				log.Printf("Connector does not support tags, skipping")
				continue
			}
			count, err := syncTags(ctx, connector, s3Client, msg.AccountID, msg.ConnectionID)
			if err != nil {
				return fmt.Errorf("tags sync failed: %w", err)
			}
			recordCounts["tags"] = count

		case "custom_fields":
			if !capSet[connectors.CapCustomFields] {
				log.Printf("Connector does not support custom_fields, skipping")
				continue
			}
			count, err := syncCustomFields(ctx, connector, s3Client, msg.AccountID, msg.ConnectionID)
			if err != nil {
				return fmt.Errorf("custom_fields sync failed: %w", err)
			}
			recordCounts["custom_fields"] = count

		default:
			log.Printf("Unknown object type %q, skipping", objectType)
		}
	}

	// Mark sync as completed
	updateSyncStatus(ctx, db, msg.ConnectionID, "completed", &now, recordCounts)
	log.Printf("Data sync completed for connection %s: %v", msg.ConnectionID, recordCounts)

	return nil
}

func syncContacts(ctx context.Context, connector connectors.CRMConnector, s3Client *s3.Client, accountID, connectionID string) (int, error) {
	var allContacts []connectors.NormalizedContact
	cursor := ""

	for {
		opts := connectors.QueryOptions{
			Limit:  200,
			Cursor: cursor,
		}

		result, err := connector.GetContacts(ctx, opts)
		if err != nil {
			return 0, fmt.Errorf("failed to get contacts page: %w", err)
		}

		allContacts = append(allContacts, result.Contacts...)
		log.Printf("Fetched %d contacts (total so far: %d, has_more: %v)",
			len(result.Contacts), len(allContacts), result.HasMore)

		if !result.HasMore || result.NextCursor == "" {
			break
		}
		cursor = result.NextCursor
	}

	// Write contacts parquet file to S3 (writer also builds SchemaInfo)
	s3Key := fmt.Sprintf("%s/%s/contacts/data.parquet", accountID, connectionID)
	schemaInfo, err := parquet.WriteContactsParquet(ctx, s3Client, analyticsBucket, s3Key, allContacts)
	if err != nil {
		return 0, fmt.Errorf("failed to write contacts parquet: %w", err)
	}

	// Populate connection ID on the schema and write schema.json alongside the parquet data
	schemaInfo.ConnectionID = connectionID
	schemaKey := fmt.Sprintf("%s/%s/contacts/schema.json", accountID, connectionID)
	if err := parquet.WriteSchema(ctx, s3Client, analyticsBucket, schemaKey, schemaInfo); err != nil {
		return 0, fmt.Errorf("failed to write contacts schema: %w", err)
	}

	return len(allContacts), nil
}

func syncTags(ctx context.Context, connector connectors.CRMConnector, s3Client *s3.Client, accountID, connectionID string) (int, error) {
	tags, err := connector.GetTags(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to get tags: %w", err)
	}

	s3Key := fmt.Sprintf("%s/%s/tags/data.parquet", accountID, connectionID)
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

func syncCustomFields(ctx context.Context, connector connectors.CRMConnector, s3Client *s3.Client, accountID, connectionID string) (int, error) {
	fields, err := connector.GetCustomFields(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to get custom fields: %w", err)
	}

	s3Key := fmt.Sprintf("%s/%s/custom_fields/data.parquet", accountID, connectionID)
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
