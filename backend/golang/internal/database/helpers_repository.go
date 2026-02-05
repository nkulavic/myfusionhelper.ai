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

// HelpersRepository provides access to the helpers DynamoDB table.
type HelpersRepository struct {
	client    *dynamodb.Client
	tableName string
}

// NewHelpersRepository creates a new HelpersRepository.
func NewHelpersRepository(client *dynamodb.Client, tableName string) *HelpersRepository {
	return &HelpersRepository{client: client, tableName: tableName}
}

// GetByID fetches a helper by its helper_id (primary key).
func (r *HelpersRepository) GetByID(ctx context.Context, helperID string) (*types.Helper, error) {
	return getItem[types.Helper](ctx, r.client, r.tableName, stringKey("helper_id", helperID))
}

// ListByAccount fetches all helpers for a given account using the AccountIdIndex GSI.
func (r *HelpersRepository) ListByAccount(ctx context.Context, accountID string) ([]types.Helper, error) {
	indexName := "AccountIdIndex"
	return queryIndex[types.Helper](ctx, r.client, &dynamodb.QueryInput{
		TableName:              &r.tableName,
		IndexName:              &indexName,
		KeyConditionExpression: aws.String("account_id = :account_id"),
		ExpressionAttributeValues: map[string]ddbtypes.AttributeValue{
			":account_id": stringVal(accountID),
		},
	})
}

// Create inserts a new helper with a condition that the helper_id does not already exist.
func (r *HelpersRepository) Create(ctx context.Context, helper *types.Helper) error {
	return putItemWithCondition(ctx, r.client, r.tableName, helper, "attribute_not_exists(helper_id)")
}

// Update performs a full replace of the helper record.
func (r *HelpersRepository) Update(ctx context.Context, helper *types.Helper) error {
	return putItem(ctx, r.client, r.tableName, helper)
}

// SoftDelete sets the helper status to "deleted".
func (r *HelpersRepository) SoftDelete(ctx context.Context, helperID string) error {
	now := time.Now().UTC()
	_, err := r.client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: &r.tableName,
		Key:       stringKey("helper_id", helperID),
		UpdateExpression: aws.String("SET #s = :status, updated_at = :updated_at"),
		ExpressionAttributeNames: map[string]string{
			"#s": "status",
		},
		ExpressionAttributeValues: map[string]ddbtypes.AttributeValue{
			":status":     stringVal("deleted"),
			":updated_at": stringVal(now.Format(time.RFC3339)),
		},
		ConditionExpression: aws.String("attribute_exists(helper_id)"),
	})
	if err != nil {
		return fmt.Errorf("soft delete helper: %w", err)
	}
	return nil
}

// UpdateExecutionStats updates the execution count and last executed timestamp.
func (r *HelpersRepository) UpdateExecutionStats(ctx context.Context, helperID string, count int64, lastAt time.Time) error {
	_, err := r.client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: &r.tableName,
		Key:       stringKey("helper_id", helperID),
		UpdateExpression: aws.String("SET execution_count = :count, last_executed_at = :last_at, updated_at = :updated_at"),
		ExpressionAttributeValues: map[string]ddbtypes.AttributeValue{
			":count":      numVal(fmt.Sprintf("%d", count)),
			":last_at":    stringVal(lastAt.UTC().Format(time.RFC3339)),
			":updated_at": stringVal(time.Now().UTC().Format(time.RFC3339)),
		},
	})
	if err != nil {
		return fmt.Errorf("update execution stats: %w", err)
	}
	return nil
}
