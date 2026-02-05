package database

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	ddbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/myfusionhelper/api/internal/types"
)

// ConnectionsRepository provides access to the connections DynamoDB table.
type ConnectionsRepository struct {
	client    *dynamodb.Client
	tableName string
}

// NewConnectionsRepository creates a new ConnectionsRepository.
func NewConnectionsRepository(client *dynamodb.Client, tableName string) *ConnectionsRepository {
	return &ConnectionsRepository{client: client, tableName: tableName}
}

// GetByID fetches a connection by its connection_id (primary key).
func (r *ConnectionsRepository) GetByID(ctx context.Context, connectionID string) (*types.PlatformConnection, error) {
	return getItem[types.PlatformConnection](ctx, r.client, r.tableName, stringKey("connection_id", connectionID))
}

// ListByAccount fetches all connections for a given account using the AccountIdIndex GSI.
func (r *ConnectionsRepository) ListByAccount(ctx context.Context, accountID string) ([]types.PlatformConnection, error) {
	indexName := "AccountIdIndex"
	return queryIndex[types.PlatformConnection](ctx, r.client, &dynamodb.QueryInput{
		TableName:              &r.tableName,
		IndexName:              &indexName,
		KeyConditionExpression: aws.String("account_id = :account_id"),
		ExpressionAttributeValues: map[string]ddbtypes.AttributeValue{
			":account_id": stringVal(accountID),
		},
	})
}

// Create inserts a new connection.
func (r *ConnectionsRepository) Create(ctx context.Context, conn *types.PlatformConnection) error {
	return putItem(ctx, r.client, r.tableName, conn)
}

// Update performs a full replace of the connection record.
func (r *ConnectionsRepository) Update(ctx context.Context, conn *types.PlatformConnection) error {
	return putItem(ctx, r.client, r.tableName, conn)
}

// Delete removes a connection by its connection_id.
func (r *ConnectionsRepository) Delete(ctx context.Context, connectionID string) error {
	_, err := r.client.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: &r.tableName,
		Key:       stringKey("connection_id", connectionID),
	})
	return err
}

// UpdateSyncStatus updates the sync-related fields on a connection.
func (r *ConnectionsRepository) UpdateSyncStatus(ctx context.Context, connectionID, status string, counts map[string]int) error {
	now := time.Now().UTC()

	updateExpr := "SET sync_status = :sync_status, last_synced_at = :last_synced_at, updated_at = :updated_at"
	exprValues := map[string]ddbtypes.AttributeValue{
		":sync_status":   stringVal(status),
		":last_synced_at": stringVal(now.Format(time.RFC3339)),
		":updated_at":    stringVal(now.Format(time.RFC3339)),
	}

	if counts != nil {
		countsAV, err := attributevalue.MarshalMap(counts)
		if err != nil {
			return fmt.Errorf("marshal sync counts: %w", err)
		}
		updateExpr += ", sync_record_counts = :counts"
		exprValues[":counts"] = &ddbtypes.AttributeValueMemberM{Value: countsAV}
	}

	_, err := r.client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName:                 &r.tableName,
		Key:                       stringKey("connection_id", connectionID),
		UpdateExpression:          &updateExpr,
		ExpressionAttributeValues: exprValues,
	})
	if err != nil {
		return fmt.Errorf("update sync status: %w", err)
	}
	return nil
}
