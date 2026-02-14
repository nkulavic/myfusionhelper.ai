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

// APIKeysRepository provides access to the api-keys DynamoDB table.
type APIKeysRepository struct {
	client    *dynamodb.Client
	tableName string
}

// NewAPIKeysRepository creates a new APIKeysRepository.
func NewAPIKeysRepository(client *dynamodb.Client, tableName string) *APIKeysRepository {
	return &APIKeysRepository{client: client, tableName: tableName}
}

// GetByID fetches an API key by its key_id (primary key).
func (r *APIKeysRepository) GetByID(ctx context.Context, keyID string) (*types.APIKey, error) {
	return getItem[types.APIKey](ctx, r.client, r.tableName, stringKey("key_id", keyID))
}

// GetByHash fetches an API key by its hash using the KeyHashIndex GSI.
func (r *APIKeysRepository) GetByHash(ctx context.Context, keyHash string) (*types.APIKey, error) {
	indexName := "KeyHashIndex"
	return querySingleItem[types.APIKey](ctx, r.client, &dynamodb.QueryInput{
		TableName:              &r.tableName,
		IndexName:              &indexName,
		KeyConditionExpression: aws.String("key_hash = :key_hash"),
		ExpressionAttributeValues: map[string]ddbtypes.AttributeValue{
			":key_hash": stringVal(keyHash),
		},
	})
}

// ListByAccount fetches all API keys for a given account using the AccountIdIndex GSI.
func (r *APIKeysRepository) ListByAccount(ctx context.Context, accountID string) ([]types.APIKey, error) {
	indexName := "AccountIdIndex"
	return queryIndex[types.APIKey](ctx, r.client, &dynamodb.QueryInput{
		TableName:              &r.tableName,
		IndexName:              &indexName,
		KeyConditionExpression: aws.String("account_id = :account_id"),
		ExpressionAttributeValues: map[string]ddbtypes.AttributeValue{
			":account_id": stringVal(accountID),
		},
	})
}

// Create inserts a new API key with a condition that the key_id does not already exist.
func (r *APIKeysRepository) Create(ctx context.Context, key *types.APIKey) error {
	return putItemWithCondition(ctx, r.client, r.tableName, key, "attribute_not_exists(key_id)")
}

// Revoke sets the API key status to "revoked".
func (r *APIKeysRepository) Revoke(ctx context.Context, keyID string) error {
	now := time.Now().UTC()
	_, err := r.client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: &r.tableName,
		Key:       stringKey("key_id", keyID),
		UpdateExpression: aws.String("SET #s = :status, updated_at = :updated_at"),
		ExpressionAttributeNames: map[string]string{
			"#s": "status",
		},
		ExpressionAttributeValues: map[string]ddbtypes.AttributeValue{
			":status":     stringVal("revoked"),
			":updated_at": stringVal(now.Format(time.RFC3339)),
		},
		ConditionExpression: aws.String("attribute_exists(key_id)"),
	})
	if err != nil {
		return fmt.Errorf("revoke api key: %w", err)
	}
	return nil
}
