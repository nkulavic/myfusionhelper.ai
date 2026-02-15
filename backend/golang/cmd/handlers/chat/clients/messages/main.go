package messages

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	ddbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"
	"github.com/myfusionhelper/api/internal/apiutil"
	authMiddleware "github.com/myfusionhelper/api/internal/middleware/auth"
	"github.com/myfusionhelper/api/internal/services"
	"github.com/myfusionhelper/api/internal/types"
	apitypes "github.com/myfusionhelper/api/internal/types"
)

const (
	GroqAPIURL   = "https://api.groq.com/openai/v1/chat/completions"
	GroqModel    = "llama-3.3-70b-versatile"
	SystemPrompt = "You are a helpful AI assistant for MyFusionHelper, a CRM automation platform. You can help users query their CRM data, manage contacts, execute automation helpers, and answer questions about their data. Use the available tools to interact with the user's CRM data and helpers."
)

var (
	conversationsTable = os.Getenv("CHAT_CONVERSATIONS_TABLE")
	messagesTable      = os.Getenv("CHAT_MESSAGES_TABLE")
)

// HandleSendWithAuth sends a new message and streams the response
func HandleSendWithAuth(ctx context.Context, event events.APIGatewayV2HTTPRequest, authCtx *apitypes.AuthContext) (events.APIGatewayV2HTTPResponse, error) {
	// Extract conversation ID from path
	conversationID := event.PathParameters["conversation_id"]
	if conversationID == "" {
		return authMiddleware.CreateErrorResponse(400, "conversation_id is required"), nil
	}

	// Parse request body
	var req types.SendMessageRequest
	if err := json.Unmarshal([]byte(apiutil.GetBody(event)), &req); err != nil {
		return authMiddleware.CreateErrorResponse(400, "Invalid request body"), nil
	}

	if req.Content == "" {
		return authMiddleware.CreateErrorResponse(400, "content is required"), nil
	}

	// Extract access token from Authorization header
	accessToken := strings.TrimPrefix(event.Headers["authorization"], "Bearer ")
	if accessToken == "" {
		accessToken = strings.TrimPrefix(event.Headers["Authorization"], "Bearer ")
	}
	if accessToken == "" {
		return authMiddleware.CreateErrorResponse(401, "Access token is required"), nil
	}

	// Get Groq API key
	groqAPIKey := getGroqAPIKey()
	if groqAPIKey == "" {
		return authMiddleware.CreateErrorResponse(500, "Groq API key not configured"), nil
	}

	// Create DynamoDB client
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return authMiddleware.CreateErrorResponse(500, "Failed to initialize service"), nil
	}
	dbClient := dynamodb.NewFromConfig(cfg)

	// Verify conversation exists and user has access
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

	if conversation.UserID != authCtx.UserID {
		return authMiddleware.CreateErrorResponse(403, "Access denied"), nil
	}

	// Get message history
	messages, err := getMessages(ctx, dbClient, conversationID)
	if err != nil {
		return authMiddleware.CreateErrorResponse(500, fmt.Sprintf("Failed to load history: %v", err)), nil
	}

	// Save user message
	userMessage := types.ChatMessage{
		MessageID:      "msg:" + uuid.New().String(),
		ConversationID: conversationID,
		Sequence:       len(messages) + 1,
		Role:           "user",
		Content:        req.Content,
		CreatedAt:      time.Now().UTC().Format(time.RFC3339),
	}
	if err := saveMessage(ctx, dbClient, &userMessage); err != nil {
		return authMiddleware.CreateErrorResponse(500, fmt.Sprintf("Failed to save message: %v", err)), nil
	}

	// Process streaming response
	sseData, err := processStreaming(ctx, dbClient, groqAPIKey, accessToken, conversationID, messages, req.Content)
	if err != nil {
		return authMiddleware.CreateErrorResponse(500, fmt.Sprintf("Failed to process message: %v", err)), nil
	}

	return events.APIGatewayV2HTTPResponse{
		StatusCode: 200,
		Headers: map[string]string{
			"Content-Type":                "text/event-stream",
			"Cache-Control":               "no-cache",
			"Connection":                  "keep-alive",
			"Access-Control-Allow-Origin": "*",
		},
		Body: sseData,
	}, nil
}

// HandleListWithAuth lists all messages for a conversation
func HandleListWithAuth(ctx context.Context, event events.APIGatewayV2HTTPRequest, authCtx *apitypes.AuthContext) (events.APIGatewayV2HTTPResponse, error) {
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

	// Verify conversation exists and user has access
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

	if conversation.UserID != authCtx.UserID {
		return authMiddleware.CreateErrorResponse(403, "Access denied"), nil
	}

	// Get messages
	messages, err := getMessages(ctx, dbClient, conversationID)
	if err != nil {
		return authMiddleware.CreateErrorResponse(500, fmt.Sprintf("Failed to get messages: %v", err)), nil
	}

	// Build response
	data := map[string]interface{}{
		"conversation_id": conversationID,
		"messages":        messages,
	}

	return authMiddleware.CreateSuccessResponse(200, "Messages retrieved successfully", data), nil
}

// getMessages retrieves all messages for a conversation
func getMessages(ctx context.Context, dbClient *dynamodb.Client, conversationID string) ([]types.ChatMessage, error) {
	result, err := dbClient.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(messagesTable),
		IndexName:              aws.String("ConversationIdIndex"),
		KeyConditionExpression: aws.String("conversation_id = :conversation_id"),
		ExpressionAttributeValues: map[string]ddbtypes.AttributeValue{
			":conversation_id": &ddbtypes.AttributeValueMemberS{Value: conversationID},
		},
	})
	if err != nil {
		return nil, err
	}

	var messages []types.ChatMessage
	for _, item := range result.Items {
		var msg types.ChatMessage
		if err := attributevalue.UnmarshalMap(item, &msg); err != nil {
			continue
		}
		messages = append(messages, msg)
	}
	return messages, nil
}

// saveMessage saves a message to DynamoDB
func saveMessage(ctx context.Context, dbClient *dynamodb.Client, message *types.ChatMessage) error {
	av, err := attributevalue.MarshalMap(message)
	if err != nil {
		return err
	}
	_, err = dbClient.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(messagesTable),
		Item:      av,
	})
	return err
}

// processStreaming handles the streaming LLM response
func processStreaming(ctx context.Context, dbClient *dynamodb.Client, groqAPIKey, accessToken, conversationID string, history []types.ChatMessage, userContent string) (string, error) {
	var sseData strings.Builder
	mcpService := services.NewMCPService(ctx)

	// Build Groq messages
	groqMessages := buildGroqMessages(history, userContent)
	tools := mcpService.GetToolDefinitions()

	httpClient := &http.Client{Timeout: 60 * time.Second}
	maxIterations := 10

	for iteration := 0; iteration < maxIterations; iteration++ {
		// Call Groq API
		content, toolCalls, err := callGroqAPI(ctx, httpClient, groqAPIKey, groqMessages, tools, &sseData)
		if err != nil {
			return "", err
		}

		// If no tool calls, we're done
		if len(toolCalls) == 0 {
			// Save assistant message
			assistantMsg := types.ChatMessage{
				MessageID:      "msg:" + uuid.New().String(),
				ConversationID: conversationID,
				Sequence:       len(groqMessages),
				Role:           "assistant",
				Content:        content,
				CreatedAt:      time.Now().UTC().Format(time.RFC3339),
			}
			saveMessage(ctx, dbClient, &assistantMsg)

			// Send done
			sseData.WriteString("data: " + toJSON(types.StreamChatResponse{Type: "done", Done: true}) + "\n\n")
			sseData.WriteString("data: [DONE]\n\n")
			return sseData.String(), nil
		}

		// Execute tool calls
		var toolResults []types.ToolResult
		var savedToolCalls []types.ToolCall
		for _, tc := range toolCalls {
			// Convert to ToolCall for response
			toolCall := types.ToolCall{
				ID:       tc.ID,
				Type:     tc.Type,
				Function: types.FunctionCall{Name: tc.Function.Name, Arguments: tc.Function.Arguments},
			}
			savedToolCalls = append(savedToolCalls, toolCall)

			sseData.WriteString("data: " + toJSON(types.StreamChatResponse{Type: "tool_call", ToolCall: &toolCall}) + "\n\n")

			result, err := mcpService.ExecuteTool(ctx, tc, accessToken)
			if err != nil {
				result = fmt.Sprintf("Error: %v", err)
			}

			tr := types.ToolResult{
				ToolCallID: tc.ID,
				Result:     result,
			}
			toolResults = append(toolResults, tr)

			sseData.WriteString("data: " + toJSON(types.StreamChatResponse{Type: "tool_result", ToolResult: &tr}) + "\n\n")
		}

		// Save assistant message with tool calls
		assistantMsg := types.ChatMessage{
			MessageID:      "msg:" + uuid.New().String(),
			ConversationID: conversationID,
			Sequence:       len(groqMessages),
			Role:           "assistant",
			Content:        content,
			ToolCalls:      savedToolCalls,
			ToolResults:    toolResults,
			CreatedAt:      time.Now().UTC().Format(time.RFC3339),
		}
		saveMessage(ctx, dbClient, &assistantMsg)

		// Add to history
		groqMessages = append(groqMessages, types.GroqMessage{Role: "assistant", Content: content})
		for _, tr := range toolResults {
			groqMessages = append(groqMessages, types.GroqMessage{Role: "tool", Content: tr.Result, ToolCallID: tr.ToolCallID, Name: "tool_result"})
		}
	}

	return sseData.String(), fmt.Errorf("max iterations reached")
}

// buildGroqMessages converts history to Groq format
func buildGroqMessages(history []types.ChatMessage, newUserMessage string) []types.GroqMessage {
	messages := []types.GroqMessage{{Role: "system", Content: SystemPrompt}}

	for _, msg := range history {
		groqMsg := types.GroqMessage{Role: msg.Role, Content: msg.Content}
		if len(msg.ToolCalls) > 0 {
			var toolCalls []types.GroqToolCall
			for _, tc := range msg.ToolCalls {
				toolCalls = append(toolCalls, types.GroqToolCall{
					ID:       tc.ID,
					Type:     tc.Type,
					Function: types.GroqFunctionCall{Name: tc.Function.Name, Arguments: tc.Function.Arguments},
				})
			}
			groqMsg.ToolCalls = toolCalls
		}
		messages = append(messages, groqMsg)

		for _, tr := range msg.ToolResults {
			messages = append(messages, types.GroqMessage{Role: "tool", Content: tr.Result, ToolCallID: tr.ToolCallID, Name: "tool_result"})
		}
	}

	messages = append(messages, types.GroqMessage{Role: "user", Content: newUserMessage})
	return messages
}

// callGroqAPI calls Groq API with streaming
func callGroqAPI(ctx context.Context, httpClient *http.Client, groqAPIKey string, messages []types.GroqMessage, tools []types.GroqTool, sseData *strings.Builder) (string, []types.GroqToolCall, error) {
	reqBody := types.GroqChatRequest{
		Model:       GroqModel,
		Messages:    messages,
		Temperature: 0.7,
		Stream:      true,
		Tools:       tools,
	}

	bodyJSON, _ := json.Marshal(reqBody)
	req, _ := http.NewRequestWithContext(ctx, "POST", GroqAPIURL, bytes.NewReader(bodyJSON))
	req.Header.Set("Authorization", "Bearer "+groqAPIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", nil, fmt.Errorf("Groq API error: %d - %s", resp.StatusCode, string(body))
	}

	// Handle streaming response
	scanner := bufio.NewScanner(resp.Body)
	var fullContent strings.Builder
	var toolCalls []types.GroqToolCall

	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data: ") {
			continue
		}

		data := strings.TrimPrefix(line, "data: ")
		if data == "[DONE]" {
			break
		}

		var chunk types.GroqStreamChunk
		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			continue
		}

		if len(chunk.Choices) == 0 {
			continue
		}

		delta := chunk.Choices[0].Delta

		if delta.Content != "" {
			fullContent.WriteString(delta.Content)
			sseData.WriteString("data: " + toJSON(types.StreamChatResponse{Type: "content", Content: delta.Content}) + "\n\n")
		}

		if len(delta.ToolCalls) > 0 {
			for _, tc := range delta.ToolCalls {
				toolCalls = append(toolCalls, tc)
			}
		}
	}

	return fullContent.String(), toolCalls, nil
}

// getGroqAPIKey retrieves the Groq API key from environment
func getGroqAPIKey() string {
	secretsJSON := os.Getenv("INTERNAL_SECRETS")
	if secretsJSON == "" {
		return ""
	}

	type InternalSecrets struct {
		Groq struct {
			APIKey string `json:"api_key"`
		} `json:"groq"`
	}

	var secrets InternalSecrets
	if err := json.Unmarshal([]byte(secretsJSON), &secrets); err != nil {
		return ""
	}
	return secrets.Groq.APIKey
}

// toJSON converts a value to JSON string
func toJSON(v interface{}) string {
	data, _ := json.Marshal(v)
	return string(data)
}
