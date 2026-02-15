package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	ddbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"
	"github.com/myfusionhelper/api/internal/apiutil"
	"github.com/myfusionhelper/api/internal/services"
	"github.com/myfusionhelper/api/internal/types"
)

var (
	conversationsTable = os.Getenv("CHAT_CONVERSATIONS_TABLE")
	messagesTable      = os.Getenv("CHAT_MESSAGES_TABLE")
	phoneMappingsTable = os.Getenv("PHONE_MAPPINGS_TABLE")
)

// TwilioWebhookRequest represents incoming Twilio SMS webhook
type TwilioWebhookRequest struct {
	From    string `json:"From"`
	To      string `json:"To"`
	Body    string `json:"Body"`
	NumSms  string `json:"NumSms"`
	SmsId   string `json:"SmsMessageSid"`
	Account string `json:"AccountSid"`
}

// PhoneMapping represents phone number to user mapping
type PhoneMapping struct {
	PhoneNumber string `dynamodbav:"phone_number"`
	UserID      string `dynamodbav:"user_id"`
	AccountID   string `dynamodbav:"account_id"`
	Status      string `dynamodbav:"status"` // "active", "blocked"
	CreatedAt   string `dynamodbav:"created_at"`
	UpdatedAt   string `dynamodbav:"updated_at"`
}

// RateLimit tracks SMS rate limiting per phone number
type RateLimit struct {
	PhoneNumber string `dynamodbav:"phone_number"`
	Count       int    `dynamodbav:"count"`
	WindowStart int64  `dynamodbav:"window_start"`
	TTL         *int64 `dynamodbav:"ttl,omitempty"`
}

const (
	MaxSMSPerHour     = 10
	RateLimitWindow   = 3600 // 1 hour in seconds
	MaxSMSLength      = 1600 // Twilio SMS limit
	TwilioContentType = "application/x-www-form-urlencoded"
)

func main() {
	lambda.Start(Handle)
}

// Handle processes incoming Twilio SMS webhooks
func Handle(ctx context.Context, event events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	// Parse Twilio webhook (form-encoded)
	twilioReq, err := parseTwilioWebhook(apiutil.GetBody(event))
	if err != nil {
		return createTwiMLResponse(fmt.Sprintf("Error: %v", err)), nil
	}

	// Validate required fields
	if twilioReq.From == "" || twilioReq.Body == "" {
		return createTwiMLResponse("Invalid request"), nil
	}

	// Create DynamoDB client
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return createTwiMLResponse("Service unavailable"), nil
	}
	dbClient := dynamodb.NewFromConfig(cfg)

	// Check rate limit
	if limited, err := checkRateLimit(ctx, dbClient, twilioReq.From); err != nil || limited {
		return createTwiMLResponse("Too many messages. Please wait before sending more."), nil
	}

	// Get phone mapping
	mapping, err := getPhoneMapping(ctx, dbClient, twilioReq.From)
	if err != nil || mapping == nil {
		return createTwiMLResponse("Phone number not registered. Please register at app.myfusionhelper.ai"), nil
	}

	if mapping.Status != "active" {
		return createTwiMLResponse("Your account is blocked. Contact support."), nil
	}

	// Get or create conversation
	conversationID, err := getOrCreateConversation(ctx, dbClient, mapping)
	if err != nil {
		return createTwiMLResponse("Failed to create conversation"), nil
	}

	// Get access token (we need a service token for SMS users)
	// For now, we'll use a placeholder - in production, this should be a service account token
	accessToken := os.Getenv("SERVICE_ACCESS_TOKEN")
	if accessToken == "" {
		// Fallback: try to get from user's stored token (if we stored it)
		// For MVP, we'll return an error
		return createTwiMLResponse("Service not configured. Please contact support."), nil
	}

	// Save user message
	userMessage := types.ChatMessage{
		MessageID:      "msg:" + uuid.New().String(),
		ConversationID: conversationID,
		Sequence:       1, // We'll need to get actual sequence
		Role:           "user",
		Content:        twilioReq.Body,
		CreatedAt:      time.Now().UTC().Format(time.RFC3339),
	}
	if err := saveMessage(ctx, dbClient, &userMessage); err != nil {
		return createTwiMLResponse("Failed to save message"), nil
	}

	// Get message history
	messages, err := getMessages(ctx, dbClient, conversationID)
	if err != nil {
		messages = []types.ChatMessage{} // Continue with empty history
	}

	// Process with MCP service
	mcpService := services.NewMCPService(ctx)
	response := processMessage(ctx, mcpService, messages, twilioReq.Body, accessToken)

	// Save assistant message
	assistantMessage := types.ChatMessage{
		MessageID:      "msg:" + uuid.New().String(),
		ConversationID: conversationID,
		Sequence:       userMessage.Sequence + 1,
		Role:           "assistant",
		Content:        response,
		CreatedAt:      time.Now().UTC().Format(time.RFC3339),
	}
	saveMessage(ctx, dbClient, &assistantMessage)

	// Update rate limit
	updateRateLimit(ctx, dbClient, twilioReq.From)

	// Send SMS response via TwiML
	return createTwiMLResponse(truncateSMS(response)), nil
}

// parseTwilioWebhook parses form-encoded Twilio webhook
func parseTwilioWebhook(body string) (*TwilioWebhookRequest, error) {
	values, err := url.ParseQuery(body)
	if err != nil {
		return nil, err
	}

	return &TwilioWebhookRequest{
		From:    values.Get("From"),
		To:      values.Get("To"),
		Body:    values.Get("Body"),
		NumSms:  values.Get("NumSms"),
		SmsId:   values.Get("SmsMessageSid"),
		Account: values.Get("AccountSid"),
	}, nil
}

// createTwiMLResponse creates Twilio TwiML response
func createTwiMLResponse(message string) events.APIGatewayV2HTTPResponse {
	twiml := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<Response>
  <Message>%s</Message>
</Response>`, message)

	return events.APIGatewayV2HTTPResponse{
		StatusCode: 200,
		Headers: map[string]string{
			"Content-Type": "text/xml",
		},
		Body: twiml,
	}
}

// checkRateLimit checks if phone number exceeds rate limit
func checkRateLimit(ctx context.Context, dbClient *dynamodb.Client, phoneNumber string) (bool, error) {
	result, err := dbClient.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String("mfh-" + os.Getenv("STAGE") + "-rate-limits"),
		Key: map[string]ddbtypes.AttributeValue{
			"phone_number": &ddbtypes.AttributeValueMemberS{Value: phoneNumber},
		},
	})
	if err != nil {
		return false, err
	}

	if result.Item == nil {
		return false, nil // No rate limit yet
	}

	var limit RateLimit
	if err := attributevalue.UnmarshalMap(result.Item, &limit); err != nil {
		return false, err
	}

	now := time.Now().Unix()
	if now-limit.WindowStart > RateLimitWindow {
		return false, nil // Window expired
	}

	return limit.Count >= MaxSMSPerHour, nil
}

// updateRateLimit updates rate limit counter
func updateRateLimit(ctx context.Context, dbClient *dynamodb.Client, phoneNumber string) error {
	now := time.Now().Unix()
	ttl := now + RateLimitWindow

	_, err := dbClient.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String("mfh-" + os.Getenv("STAGE") + "-rate-limits"),
		Key: map[string]ddbtypes.AttributeValue{
			"phone_number": &ddbtypes.AttributeValueMemberS{Value: phoneNumber},
		},
		UpdateExpression: aws.String("ADD #count :inc SET #start = if_not_exists(#start, :now), #ttl = :ttl"),
		ExpressionAttributeNames: map[string]string{
			"#count": "count",
			"#start": "window_start",
			"#ttl":   "ttl",
		},
		ExpressionAttributeValues: map[string]ddbtypes.AttributeValue{
			":inc": &ddbtypes.AttributeValueMemberN{Value: "1"},
			":now": &ddbtypes.AttributeValueMemberN{Value: fmt.Sprintf("%d", now)},
			":ttl": &ddbtypes.AttributeValueMemberN{Value: fmt.Sprintf("%d", ttl)},
		},
	})
	return err
}

// getPhoneMapping gets phone number to user mapping
func getPhoneMapping(ctx context.Context, dbClient *dynamodb.Client, phoneNumber string) (*PhoneMapping, error) {
	result, err := dbClient.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(phoneMappingsTable),
		Key: map[string]ddbtypes.AttributeValue{
			"phone_number": &ddbtypes.AttributeValueMemberS{Value: phoneNumber},
		},
	})
	if err != nil {
		return nil, err
	}

	if result.Item == nil {
		return nil, nil
	}

	var mapping PhoneMapping
	if err := attributevalue.UnmarshalMap(result.Item, &mapping); err != nil {
		return nil, err
	}

	return &mapping, nil
}

// getOrCreateConversation gets existing or creates new conversation
func getOrCreateConversation(ctx context.Context, dbClient *dynamodb.Client, mapping *PhoneMapping) (string, error) {
	// Query for existing active conversation
	result, err := dbClient.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(conversationsTable),
		IndexName:              aws.String("AccountIdIndex"),
		KeyConditionExpression: aws.String("account_id = :account_id"),
		ExpressionAttributeValues: map[string]ddbtypes.AttributeValue{
			":account_id": &ddbtypes.AttributeValueMemberS{Value: mapping.AccountID},
		},
		Limit:            aws.Int32(1),
		ScanIndexForward: aws.Bool(false), // Most recent first
	})
	if err != nil {
		return "", err
	}

	if len(result.Items) > 0 {
		var conv types.ChatConversation
		if err := attributevalue.UnmarshalMap(result.Items[0], &conv); err == nil {
			return conv.ConversationID, nil
		}
	}

	// Create new conversation
	conversationID := "conv:" + uuid.New().String()
	now := time.Now().UTC().Format(time.RFC3339)
	ttl := time.Now().Add(90 * 24 * time.Hour).Unix()

	conversation := types.ChatConversation{
		ConversationID: conversationID,
		UserID:         mapping.UserID,
		AccountID:      mapping.AccountID,
		Title:          "SMS Conversation",
		CreatedAt:      now,
		UpdatedAt:      now,
		MessageCount:   0,
		TTL:            &ttl,
	}

	av, err := attributevalue.MarshalMap(conversation)
	if err != nil {
		return "", err
	}

	_, err = dbClient.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(conversationsTable),
		Item:      av,
	})
	if err != nil {
		return "", err
	}

	return conversationID, nil
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
		Limit: aws.Int32(10), // Last 10 messages for context
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

// processMessage processes message with MCP service
func processMessage(ctx context.Context, mcpService *services.MCPService, history []types.ChatMessage, userMessage, accessToken string) string {
	// Build Groq messages from history
	groqMessages := []types.GroqMessage{
		{Role: "system", Content: "You are a helpful AI assistant for MyFusionHelper. Keep responses brief and SMS-friendly (under 1600 characters)."},
	}

	for _, msg := range history {
		groqMessages = append(groqMessages, types.GroqMessage{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	groqMessages = append(groqMessages, types.GroqMessage{
		Role:    "user",
		Content: userMessage,
	})

	// Get tool definitions
	tools := mcpService.GetToolDefinitions()

	// Call Groq API (simplified - no streaming for SMS)
	groqAPIKey := getGroqAPIKey()
	if groqAPIKey == "" {
		return "Service not configured. Please contact support."
	}

	reqBody := types.GroqChatRequest{
		Model:       "llama-3.3-70b-versatile",
		Messages:    groqMessages,
		Temperature: 0.7,
		Stream:      false,
		Tools:       tools,
	}

	bodyJSON, _ := json.Marshal(reqBody)
	req, _ := http.NewRequestWithContext(ctx, "POST", "https://api.groq.com/openai/v1/chat/completions", strings.NewReader(string(bodyJSON)))
	req.Header.Set("Authorization", "Bearer "+groqAPIKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "Failed to process message. Please try again."
	}
	defer resp.Body.Close()

	var groqResp types.GroqChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&groqResp); err != nil {
		return "Failed to process response. Please try again."
	}

	if len(groqResp.Choices) == 0 {
		return "No response generated. Please try again."
	}

	choice := groqResp.Choices[0]
	content := choice.Message.Content

	// Handle tool calls
	if len(choice.Message.ToolCalls) > 0 {
		for _, tc := range choice.Message.ToolCalls {
			result, err := mcpService.ExecuteTool(ctx, tc, accessToken)
			if err != nil {
				content += fmt.Sprintf("\n\nTool error: %v", err)
			} else {
				// Append tool result to context and make another call
				// For SMS simplicity, we'll just include the result
				content += fmt.Sprintf("\n\n[%s result: %s]", tc.Function.Name, truncateToolResult(result))
			}
		}
	}

	return content
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

// truncateSMS truncates message to SMS length limit
func truncateSMS(message string) string {
	if len(message) <= MaxSMSLength {
		return message
	}
	return message[:MaxSMSLength-3] + "..."
}

// truncateToolResult truncates tool result for SMS context
func truncateToolResult(result string) string {
	if len(result) <= 200 {
		return result
	}
	return result[:197] + "..."
}
