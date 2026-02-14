package database

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	ddbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/myfusionhelper/api/internal/types"
)

// ChatMessagesRepository provides access to the chat messages DynamoDB table.
type ChatMessagesRepository struct {
	client    *dynamodb.Client
	tableName string
}

// NewChatMessagesRepository creates a new ChatMessagesRepository.
func NewChatMessagesRepository(client *dynamodb.Client, tableName string) *ChatMessagesRepository {
	return &ChatMessagesRepository{client: client, tableName: tableName}
}

// GetMessage fetches a message by its message_id (primary key).
func (r *ChatMessagesRepository) GetMessage(ctx context.Context, messageID string) (*types.ChatMessage, error) {
	return getItem[types.ChatMessage](ctx, r.client, r.tableName, stringKey("message_id", messageID))
}

// ListMessagesByConversation fetches messages for a conversation using the ConversationIdSequenceIndex GSI
// ordered by sequence number. Supports pagination using lastSequence.
func (r *ChatMessagesRepository) ListMessagesByConversation(ctx context.Context, conversationID string, limit int, lastSequence int) ([]types.ChatMessage, int, error) {
	indexName := "ConversationIdSequenceIndex"
	input := &dynamodb.QueryInput{
		TableName:              &r.tableName,
		IndexName:              &indexName,
		KeyConditionExpression: aws.String("conversation_id = :conversation_id"),
		ExpressionAttributeValues: map[string]ddbtypes.AttributeValue{
			":conversation_id": stringVal(conversationID),
		},
		ScanIndexForward: aws.Bool(true), // Ascending order by sequence
		Limit:            aws.Int32(int32(limit)),
	}

	// If lastSequence is provided, start from the next sequence
	if lastSequence > 0 {
		input.KeyConditionExpression = aws.String("conversation_id = :conversation_id AND #seq > :last_seq")
		input.ExpressionAttributeNames = map[string]string{
			"#seq": "sequence",
		}
		input.ExpressionAttributeValues[":last_seq"] = numVal(fmt.Sprintf("%d", lastSequence))
	}

	result, err := r.client.Query(ctx, input)
	if err != nil {
		return nil, 0, err
	}

	items := make([]types.ChatMessage, 0, len(result.Items))
	var nextSequence int
	for _, item := range result.Items {
		var msg types.ChatMessage
		if err := attributevalue.UnmarshalMap(item, &msg); err != nil {
			return nil, 0, err
		}
		items = append(items, msg)
		nextSequence = msg.Sequence
	}

	// Return 0 if no more results
	if result.LastEvaluatedKey == nil {
		nextSequence = 0
	}

	return items, nextSequence, nil
}

// CountMessagesByConversation returns the total message count for a conversation.
func (r *ChatMessagesRepository) CountMessagesByConversation(ctx context.Context, conversationID string) (int, error) {
	indexName := "ConversationIdSequenceIndex"
	input := &dynamodb.QueryInput{
		TableName:              &r.tableName,
		IndexName:              &indexName,
		KeyConditionExpression: aws.String("conversation_id = :conversation_id"),
		ExpressionAttributeValues: map[string]ddbtypes.AttributeValue{
			":conversation_id": stringVal(conversationID),
		},
		Select: ddbtypes.SelectCount, // Only return count, not items
	}

	result, err := r.client.Query(ctx, input)
	if err != nil {
		return 0, err
	}

	return int(result.Count), nil
}

// CreateMessage inserts a new message with auto-incrementing sequence.
// The sequence is determined by fetching the max sequence for the conversation and adding 1.
func (r *ChatMessagesRepository) CreateMessage(ctx context.Context, message *types.ChatMessage) error {
	// Get the next sequence number for this conversation
	nextSeq, err := r.getNextSequence(ctx, message.ConversationID)
	if err != nil {
		return fmt.Errorf("get next sequence: %w", err)
	}

	message.Sequence = nextSeq

	// Insert with condition that message_id doesn't exist
	return putItemWithCondition(ctx, r.client, r.tableName, message, "attribute_not_exists(message_id)")
}

// getNextSequence determines the next sequence number for a conversation.
// It queries the ConversationIdSequenceIndex in descending order to get the highest sequence.
func (r *ChatMessagesRepository) getNextSequence(ctx context.Context, conversationID string) (int, error) {
	indexName := "ConversationIdSequenceIndex"
	input := &dynamodb.QueryInput{
		TableName:              &r.tableName,
		IndexName:              &indexName,
		KeyConditionExpression: aws.String("conversation_id = :conversation_id"),
		ExpressionAttributeValues: map[string]ddbtypes.AttributeValue{
			":conversation_id": stringVal(conversationID),
		},
		ScanIndexForward: aws.Bool(false), // Descending to get max sequence
		Limit:            aws.Int32(1),
		ProjectionExpression: aws.String("#seq"),
		ExpressionAttributeNames: map[string]string{
			"#seq": "sequence",
		},
	}

	result, err := r.client.Query(ctx, input)
	if err != nil {
		return 0, fmt.Errorf("query max sequence: %w", err)
	}

	// If no messages exist, start at 1
	if len(result.Items) == 0 {
		return 1, nil
	}

	// Parse the sequence from the first (and only) result
	var msg types.ChatMessage
	if err := attributevalue.UnmarshalMap(result.Items[0], &msg); err != nil {
		return 0, fmt.Errorf("unmarshal sequence: %w", err)
	}

	return msg.Sequence + 1, nil
}
