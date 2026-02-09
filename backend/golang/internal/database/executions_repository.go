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

// ExecutionsRepository provides access to the executions DynamoDB table.
type ExecutionsRepository struct {
	client    *dynamodb.Client
	tableName string
}

// NewExecutionsRepository creates a new ExecutionsRepository.
func NewExecutionsRepository(client *dynamodb.Client, tableName string) *ExecutionsRepository {
	return &ExecutionsRepository{client: client, tableName: tableName}
}

// GetByID fetches an execution by its execution_id (primary key).
func (r *ExecutionsRepository) GetByID(ctx context.Context, executionID string) (*types.Execution, error) {
	return getItem[types.Execution](ctx, r.client, r.tableName, stringKey("execution_id", executionID))
}

// ListByAccount fetches executions for an account using the AccountIdCreatedAtIndex GSI
// with cursor-based pagination. The cursor is the created_at value of the last item.
func (r *ExecutionsRepository) ListByAccount(ctx context.Context, accountID string, limit int, cursor string) ([]types.Execution, string, error) {
	indexName := "AccountIdCreatedAtIndex"
	input := &dynamodb.QueryInput{
		TableName:              &r.tableName,
		IndexName:              &indexName,
		KeyConditionExpression: aws.String("account_id = :account_id"),
		ExpressionAttributeValues: map[string]ddbtypes.AttributeValue{
			":account_id": stringVal(accountID),
		},
		ScanIndexForward: aws.Bool(false),
		Limit:            aws.Int32(int32(limit)),
	}

	if cursor != "" {
		startKey, err := decodeCursor(cursor)
		if err != nil {
			return nil, "", fmt.Errorf("invalid cursor: %w", err)
		}
		input.ExclusiveStartKey = startKey
	}

	result, err := r.client.Query(ctx, input)
	if err != nil {
		return nil, "", err
	}

	items := make([]types.Execution, 0, len(result.Items))
	for _, item := range result.Items {
		var exec types.Execution
		if err := attributevalue.UnmarshalMap(item, &exec); err != nil {
			return nil, "", err
		}
		items = append(items, exec)
	}

	var nextCursor string
	if result.LastEvaluatedKey != nil {
		nextCursor, err = encodeCursor(result.LastEvaluatedKey)
		if err != nil {
			return nil, "", fmt.Errorf("encode cursor: %w", err)
		}
	}

	return items, nextCursor, nil
}

// ListByHelper fetches executions for a helper using the HelperIdCreatedAtIndex GSI
// with cursor-based pagination.
func (r *ExecutionsRepository) ListByHelper(ctx context.Context, helperID string, limit int, cursor string) ([]types.Execution, string, error) {
	indexName := "HelperIdCreatedAtIndex"
	input := &dynamodb.QueryInput{
		TableName:              &r.tableName,
		IndexName:              &indexName,
		KeyConditionExpression: aws.String("helper_id = :helper_id"),
		ExpressionAttributeValues: map[string]ddbtypes.AttributeValue{
			":helper_id": stringVal(helperID),
		},
		ScanIndexForward: aws.Bool(false),
		Limit:            aws.Int32(int32(limit)),
	}

	if cursor != "" {
		startKey, err := decodeCursor(cursor)
		if err != nil {
			return nil, "", fmt.Errorf("invalid cursor: %w", err)
		}
		input.ExclusiveStartKey = startKey
	}

	result, err := r.client.Query(ctx, input)
	if err != nil {
		return nil, "", err
	}

	items := make([]types.Execution, 0, len(result.Items))
	for _, item := range result.Items {
		var exec types.Execution
		if err := attributevalue.UnmarshalMap(item, &exec); err != nil {
			return nil, "", err
		}
		items = append(items, exec)
	}

	var nextCursor string
	if result.LastEvaluatedKey != nil {
		nextCursor, err = encodeCursor(result.LastEvaluatedKey)
		if err != nil {
			return nil, "", fmt.Errorf("encode cursor: %w", err)
		}
	}

	return items, nextCursor, nil
}

// Create inserts a new execution record.
func (r *ExecutionsRepository) Create(ctx context.Context, exec *types.Execution) error {
	return putItem(ctx, r.client, r.tableName, exec)
}

// UpdateResult updates the status, output, and duration of a completed execution.
func (r *ExecutionsRepository) UpdateResult(ctx context.Context, executionID, status string, output map[string]interface{}, durationMs int64) error {
	now := time.Now().UTC()
	updateExpr := "SET #s = :status, duration_ms = :duration_ms, completed_at = :completed_at"
	exprValues := map[string]ddbtypes.AttributeValue{
		":status":       stringVal(status),
		":duration_ms":  numVal(fmt.Sprintf("%d", durationMs)),
		":completed_at": stringVal(now.Format(time.RFC3339)),
	}
	exprNames := map[string]string{
		"#s": "status",
	}

	if output != nil {
		outputAV, err := attributevalue.MarshalMap(output)
		if err != nil {
			return fmt.Errorf("marshal output: %w", err)
		}
		updateExpr += ", output = :output"
		exprValues[":output"] = &ddbtypes.AttributeValueMemberM{Value: outputAV}
	}

	_, err := r.client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName:                 &r.tableName,
		Key:                       stringKey("execution_id", executionID),
		UpdateExpression:          &updateExpr,
		ExpressionAttributeValues: exprValues,
		ExpressionAttributeNames:  exprNames,
	})
	if err != nil {
		return fmt.Errorf("update execution result: %w", err)
	}
	return nil
}
