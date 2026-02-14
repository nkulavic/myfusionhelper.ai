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

// ChatConversationsRepository provides access to the chat conversations DynamoDB table.
type ChatConversationsRepository struct {
	client    *dynamodb.Client
	tableName string
}

// NewChatConversationsRepository creates a new ChatConversationsRepository.
func NewChatConversationsRepository(client *dynamodb.Client, tableName string) *ChatConversationsRepository {
	return &ChatConversationsRepository{client: client, tableName: tableName}
}

// GetConversation fetches a conversation by its conversation_id (primary key).
func (r *ChatConversationsRepository) GetConversation(ctx context.Context, conversationID string) (*types.ChatConversation, error) {
	return getItem[types.ChatConversation](ctx, r.client, r.tableName, stringKey("conversation_id", conversationID))
}

// ListConversationsByAccount fetches conversations for an account using the AccountIdIndex GSI
// with cursor-based pagination. Excludes soft-deleted conversations.
func (r *ChatConversationsRepository) ListConversationsByAccount(ctx context.Context, accountID string, limit int, cursor string) ([]types.ChatConversation, string, error) {
	indexName := "AccountIdIndex"
	input := &dynamodb.QueryInput{
		TableName:              &r.tableName,
		IndexName:              &indexName,
		KeyConditionExpression: aws.String("account_id = :account_id"),
		FilterExpression:       aws.String("attribute_not_exists(deleted_at)"),
		ExpressionAttributeValues: map[string]ddbtypes.AttributeValue{
			":account_id": stringVal(accountID),
		},
		ScanIndexForward: aws.Bool(false), // Most recent first
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

	items := make([]types.ChatConversation, 0, len(result.Items))
	for _, item := range result.Items {
		var conv types.ChatConversation
		if err := attributevalue.UnmarshalMap(item, &conv); err != nil {
			return nil, "", err
		}
		items = append(items, conv)
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

// CreateConversation inserts a new conversation with a condition that the conversation_id does not exist.
func (r *ChatConversationsRepository) CreateConversation(ctx context.Context, conversation *types.ChatConversation) error {
	return putItemWithCondition(ctx, r.client, r.tableName, conversation, "attribute_not_exists(conversation_id)")
}

// DeleteConversation performs a soft delete by setting deleted_at timestamp.
func (r *ChatConversationsRepository) DeleteConversation(ctx context.Context, conversationID string) error {
	now := time.Now().UTC()
	_, err := r.client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: &r.tableName,
		Key:       stringKey("conversation_id", conversationID),
		UpdateExpression: aws.String("SET deleted_at = :deleted_at, updated_at = :updated_at"),
		ExpressionAttributeValues: map[string]ddbtypes.AttributeValue{
			":deleted_at": stringVal(now.Format(time.RFC3339)),
			":updated_at": stringVal(now.Format(time.RFC3339)),
		},
		ConditionExpression: aws.String("attribute_exists(conversation_id) AND attribute_not_exists(deleted_at)"),
	})
	if err != nil {
		return fmt.Errorf("soft delete conversation: %w", err)
	}
	return nil
}

// UpdateConversationMetadata updates conversation metadata fields (title, updated_at).
func (r *ChatConversationsRepository) UpdateConversationMetadata(ctx context.Context, conversationID string, updates map[string]interface{}) error {
	now := time.Now().UTC()

	// Build dynamic update expression
	updateExpr := "SET updated_at = :updated_at"
	exprValues := map[string]ddbtypes.AttributeValue{
		":updated_at": stringVal(now.Format(time.RFC3339)),
	}

	if title, ok := updates["title"].(string); ok {
		updateExpr += ", title = :title"
		exprValues[":title"] = stringVal(title)
	}

	_, err := r.client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName:                 &r.tableName,
		Key:                       stringKey("conversation_id", conversationID),
		UpdateExpression:          aws.String(updateExpr),
		ExpressionAttributeValues: exprValues,
		ConditionExpression:       aws.String("attribute_exists(conversation_id) AND attribute_not_exists(deleted_at)"),
	})
	if err != nil {
		return fmt.Errorf("update conversation metadata: %w", err)
	}
	return nil
}

// IncrementMessageCount atomically increments the message_count field.
func (r *ChatConversationsRepository) IncrementMessageCount(ctx context.Context, conversationID string) error {
	now := time.Now().UTC()
	_, err := r.client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: &r.tableName,
		Key:       stringKey("conversation_id", conversationID),
		UpdateExpression: aws.String("ADD message_count :inc SET updated_at = :updated_at"),
		ExpressionAttributeValues: map[string]ddbtypes.AttributeValue{
			":inc":        numVal("1"),
			":updated_at": stringVal(now.Format(time.RFC3339)),
		},
		ConditionExpression: aws.String("attribute_exists(conversation_id)"),
	})
	if err != nil {
		return fmt.Errorf("increment message count: %w", err)
	}
	return nil
}
