package conversations

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	ddbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"
	authMiddleware "github.com/myfusionhelper/api/internal/middleware/auth"
	"github.com/myfusionhelper/api/internal/types"
	apitypes "github.com/myfusionhelper/api/internal/types"
)

var (
	conversationsTable = os.Getenv("CHAT_CONVERSATIONS_TABLE")
	messagesTable      = os.Getenv("CHAT_MESSAGES_TABLE")
)

// HandleWithAuth creates a new conversation
func HandleWithAuth(ctx context.Context, event events.APIGatewayV2HTTPRequest, authCtx *apitypes.AuthContext) (events.APIGatewayV2HTTPResponse, error) {
	// Parse request body
	var req types.CreateConversationRequest
	if event.Body != "" {
		if err := json.Unmarshal([]byte(event.Body), &req); err != nil {
			return authMiddleware.CreateErrorResponse(400, "Invalid request body"), nil
		}
	}

	// Create DynamoDB client
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return authMiddleware.CreateErrorResponse(500, "Failed to initialize service"), nil
	}
	dbClient := dynamodb.NewFromConfig(cfg)

	// Generate conversation ID
	conversationID := "conv:" + uuid.New().String()
	title := req.Title
	if title == "" {
		title = "New Conversation"
	}

	now := time.Now().UTC().Format(time.RFC3339)
	ttl := time.Now().Add(90 * 24 * time.Hour).Unix()

	conversation := types.ChatConversation{
		ConversationID: conversationID,
		UserID:         authCtx.UserID,
		AccountID:      authCtx.AccountID,
		Title:          title,
		CreatedAt:      now,
		UpdatedAt:      now,
		MessageCount:   0,
		TTL:            &ttl,
	}

	// Save to DynamoDB
	av, err := attributevalue.MarshalMap(conversation)
	if err != nil {
		return authMiddleware.CreateErrorResponse(500, "Failed to create conversation"), nil
	}

	_, err = dbClient.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(conversationsTable),
		Item:      av,
	})
	if err != nil {
		return authMiddleware.CreateErrorResponse(500, fmt.Sprintf("Failed to save conversation: %v", err)), nil
	}

	// Build response
	response := types.CreateConversationResponse{
		ConversationID: conversation.ConversationID,
		Title:          conversation.Title,
		Status:         "active",
		CreatedAt:      conversation.CreatedAt,
	}

	return authMiddleware.CreateSuccessResponse(201, "Conversation created successfully", response), nil
}

// HandleListWithAuth lists all conversations for the current user
func HandleListWithAuth(ctx context.Context, event events.APIGatewayV2HTTPRequest, authCtx *apitypes.AuthContext) (events.APIGatewayV2HTTPResponse, error) {
	// Create DynamoDB client
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return authMiddleware.CreateErrorResponse(500, "Failed to initialize service"), nil
	}
	dbClient := dynamodb.NewFromConfig(cfg)

	// Query conversations by user_id using GSI
	result, err := dbClient.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(conversationsTable),
		IndexName:              aws.String("UserIdIndex"),
		KeyConditionExpression: aws.String("user_id = :user_id"),
		ExpressionAttributeValues: map[string]ddbtypes.AttributeValue{
			":user_id": &ddbtypes.AttributeValueMemberS{Value: authCtx.UserID},
		},
	})
	if err != nil {
		return authMiddleware.CreateErrorResponse(500, fmt.Sprintf("Failed to list conversations: %v", err)), nil
	}

	// Unmarshal conversations
	var conversations []types.ChatConversation
	for _, item := range result.Items {
		var conv types.ChatConversation
		if err := attributevalue.UnmarshalMap(item, &conv); err != nil {
			continue
		}
		// Filter out deleted conversations
		if conv.DeletedAt == nil {
			conversations = append(conversations, conv)
		}
	}

	// Build response
	response := types.ListConversationsResponse{
		Conversations: conversations,
	}

	return authMiddleware.CreateSuccessResponse(200, "Conversations retrieved successfully", response), nil
}

// HandleGetWithAuth gets a specific conversation with its messages
func HandleGetWithAuth(ctx context.Context, event events.APIGatewayV2HTTPRequest, authCtx *apitypes.AuthContext) (events.APIGatewayV2HTTPResponse, error) {
	// Extract conversation ID from path
	conversationID := event.PathParameters["conversation_id"]
	if conversationID == "" {
		return authMiddleware.CreateErrorResponse(400, "conversation_id is required"), nil
	}

	// Create DynamoDB client
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return authMiddleware.CreateErrorResponse(500, "Failed to initialize service"), nil
	}
	dbClient := dynamodb.NewFromConfig(cfg)

	// Get conversation
	convResult, err := dbClient.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(conversationsTable),
		Key: map[string]ddbtypes.AttributeValue{
			"conversation_id": &ddbtypes.AttributeValueMemberS{Value: conversationID},
		},
	})
	if err != nil || convResult.Item == nil {
		return authMiddleware.CreateErrorResponse(404, "Conversation not found"), nil
	}

	var conversation types.ChatConversation
	if err := attributevalue.UnmarshalMap(convResult.Item, &conversation); err != nil {
		return authMiddleware.CreateErrorResponse(500, "Failed to parse conversation"), nil
	}

	// Verify user has access
	if conversation.UserID != authCtx.UserID {
		return authMiddleware.CreateErrorResponse(403, "Access denied"), nil
	}

	// Get messages
	msgsResult, err := dbClient.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(messagesTable),
		IndexName:              aws.String("ConversationIdIndex"),
		KeyConditionExpression: aws.String("conversation_id = :conversation_id"),
		ExpressionAttributeValues: map[string]ddbtypes.AttributeValue{
			":conversation_id": &ddbtypes.AttributeValueMemberS{Value: conversationID},
		},
	})
	if err != nil {
		return authMiddleware.CreateErrorResponse(500, fmt.Sprintf("Failed to get messages: %v", err)), nil
	}

	var messages []types.ChatMessage
	for _, item := range msgsResult.Items {
		var msg types.ChatMessage
		if err := attributevalue.UnmarshalMap(item, &msg); err != nil {
			continue
		}
		messages = append(messages, msg)
	}

	// Build response
	response := types.GetConversationResponse{
		Conversation: conversation,
		Messages:     messages,
	}

	return authMiddleware.CreateSuccessResponse(200, "Conversation retrieved successfully", response), nil
}

// HandleDeleteWithAuth deletes a conversation (soft delete)
func HandleDeleteWithAuth(ctx context.Context, event events.APIGatewayV2HTTPRequest, authCtx *apitypes.AuthContext) (events.APIGatewayV2HTTPResponse, error) {
	// Extract conversation ID from path
	conversationID := event.PathParameters["conversation_id"]
	if conversationID == "" {
		return authMiddleware.CreateErrorResponse(400, "conversation_id is required"), nil
	}

	// Create DynamoDB client
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return authMiddleware.CreateErrorResponse(500, "Failed to initialize service"), nil
	}
	dbClient := dynamodb.NewFromConfig(cfg)

	// Get conversation to verify ownership
	convResult, err := dbClient.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(conversationsTable),
		Key: map[string]ddbtypes.AttributeValue{
			"conversation_id": &ddbtypes.AttributeValueMemberS{Value: conversationID},
		},
	})
	if err != nil || convResult.Item == nil {
		return authMiddleware.CreateErrorResponse(404, "Conversation not found"), nil
	}

	var conversation types.ChatConversation
	if err := attributevalue.UnmarshalMap(convResult.Item, &conversation); err != nil {
		return authMiddleware.CreateErrorResponse(500, "Failed to parse conversation"), nil
	}

	// Verify user has access
	if conversation.UserID != authCtx.UserID {
		return authMiddleware.CreateErrorResponse(403, "Access denied"), nil
	}

	// Soft delete by setting deleted_at
	now := time.Now().UTC().Format(time.RFC3339)
	_, err = dbClient.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(conversationsTable),
		Key: map[string]ddbtypes.AttributeValue{
			"conversation_id": &ddbtypes.AttributeValueMemberS{Value: conversationID},
		},
		UpdateExpression: aws.String("SET deleted_at = :deleted_at, updated_at = :updated_at"),
		ExpressionAttributeValues: map[string]ddbtypes.AttributeValue{
			":deleted_at": &ddbtypes.AttributeValueMemberS{Value: now},
			":updated_at": &ddbtypes.AttributeValueMemberS{Value: now},
		},
	})
	if err != nil {
		return authMiddleware.CreateErrorResponse(500, fmt.Sprintf("Failed to delete conversation: %v", err)), nil
	}

	data := map[string]interface{}{
		"conversation_id": conversationID,
		"deleted":         true,
	}

	return authMiddleware.CreateSuccessResponse(200, "Conversation deleted successfully", data), nil
}
