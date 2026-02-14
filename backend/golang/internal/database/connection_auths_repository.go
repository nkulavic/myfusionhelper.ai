package database

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	ddbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/myfusionhelper/api/internal/types"
)

// ConnectionAuthsRepository provides access to the platform-connection-auths DynamoDB table.
type ConnectionAuthsRepository struct {
	client    *dynamodb.Client
	tableName string
}

// NewConnectionAuthsRepository creates a new ConnectionAuthsRepository.
func NewConnectionAuthsRepository(client *dynamodb.Client, tableName string) *ConnectionAuthsRepository {
	return &ConnectionAuthsRepository{client: client, tableName: tableName}
}

// GetByID fetches a connection auth by its auth_id (primary key).
func (r *ConnectionAuthsRepository) GetByID(ctx context.Context, authID string) (*types.PlatformConnectionAuth, error) {
	return getItem[types.PlatformConnectionAuth](ctx, r.client, r.tableName, stringKey("auth_id", authID))
}

// GetByConnectionID fetches a connection auth by connection_id using the ConnectionIdIndex GSI.
func (r *ConnectionAuthsRepository) GetByConnectionID(ctx context.Context, connectionID string) (*types.PlatformConnectionAuth, error) {
	indexName := "ConnectionIdIndex"
	return querySingleItem[types.PlatformConnectionAuth](ctx, r.client, &dynamodb.QueryInput{
		TableName:              &r.tableName,
		IndexName:              &indexName,
		KeyConditionExpression: aws.String("connection_id = :connection_id"),
		ExpressionAttributeValues: map[string]ddbtypes.AttributeValue{
			":connection_id": stringVal(connectionID),
		},
	})
}

// Create inserts a new connection auth record.
func (r *ConnectionAuthsRepository) Create(ctx context.Context, auth *types.PlatformConnectionAuth) error {
	return putItem(ctx, r.client, r.tableName, auth)
}

// Update performs a full replace of the connection auth record.
func (r *ConnectionAuthsRepository) Update(ctx context.Context, auth *types.PlatformConnectionAuth) error {
	return putItem(ctx, r.client, r.tableName, auth)
}

// Revoke sets the auth record status to "revoked" and records the revocation timestamp.
func (r *ConnectionAuthsRepository) Revoke(ctx context.Context, authID string) error {
	now := time.Now().Unix()
	nowStr := fmt.Sprintf("%d", now)

	_, err := r.client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: &r.tableName,
		Key:       stringKey("auth_id", authID),
		UpdateExpression: aws.String("SET #s = :status, revoked_at = :revoked_at, updated_at = :updated_at"),
		ExpressionAttributeNames: map[string]string{
			"#s": "status",
		},
		ExpressionAttributeValues: map[string]ddbtypes.AttributeValue{
			":status":     stringVal("revoked"),
			":revoked_at": numVal(nowStr),
			":updated_at": numVal(nowStr),
		},
	})
	if err != nil {
		return fmt.Errorf("revoke connection auth: %w", err)
	}
	return nil
}
