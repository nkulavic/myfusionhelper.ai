package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/google/uuid"
	"github.com/myfusionhelper/api/internal/services"
	"github.com/myfusionhelper/api/internal/types"
)

var (
	conversationsTable = os.Getenv("CHAT_CONVERSATIONS_TABLE")
	messagesTable      = os.Getenv("CHAT_MESSAGES_TABLE")
)

// AlexaRequest represents an Alexa skill request
type AlexaRequest struct {
	Version string             `json:"version"`
	Session AlexaSession       `json:"session"`
	Context AlexaContext       `json:"context"`
	Request AlexaRequestDetail `json:"request"`
}

// AlexaSession represents Alexa session data
type AlexaSession struct {
	New         bool                   `json:"new"`
	SessionID   string                 `json:"sessionId"`
	Application map[string]interface{} `json:"application"`
	User        AlexaUser              `json:"user"`
	Attributes  map[string]interface{} `json:"attributes,omitempty"`
}

// AlexaUser represents Alexa user data
type AlexaUser struct {
	UserID      string `json:"userId"`
	AccessToken string `json:"accessToken,omitempty"`
}

// AlexaContext represents Alexa context data
type AlexaContext struct {
	System AlexaSystem `json:"System"`
}

// AlexaSystem represents Alexa system data
type AlexaSystem struct {
	Device      map[string]interface{} `json:"device"`
	Application map[string]interface{} `json:"application"`
}

// AlexaRequestDetail represents the actual request
type AlexaRequestDetail struct {
	Type      string                 `json:"type"` // LaunchRequest, IntentRequest, SessionEndedRequest
	RequestID string                 `json:"requestId"`
	Timestamp string                 `json:"timestamp"`
	Locale    string                 `json:"locale"`
	Intent    *AlexaIntent           `json:"intent,omitempty"`
	Reason    string                 `json:"reason,omitempty"`
	Error     map[string]interface{} `json:"error,omitempty"`
}

// AlexaIntent represents an intent with slots
type AlexaIntent struct {
	Name               string                 `json:"name"`
	ConfirmationStatus string                 `json:"confirmationStatus"`
	Slots              map[string]AlexaSlot   `json:"slots,omitempty"`
}

// AlexaSlot represents a slot value
type AlexaSlot struct {
	Name        string      `json:"name"`
	Value       string      `json:"value,omitempty"`
	Resolutions interface{} `json:"resolutions,omitempty"`
}

// AlexaResponse represents an Alexa skill response
type AlexaResponse struct {
	Version           string                 `json:"version"`
	SessionAttributes map[string]interface{} `json:"sessionAttributes,omitempty"`
	Response          AlexaResponseBody      `json:"response"`
}

// AlexaResponseBody represents the response body
type AlexaResponseBody struct {
	OutputSpeech     *AlexaSpeech `json:"outputSpeech,omitempty"`
	Card             *AlexaCard   `json:"card,omitempty"`
	Reprompt         *AlexaSpeech `json:"reprompt,omitempty"`
	ShouldEndSession bool         `json:"shouldEndSession"`
}

// AlexaSpeech represents speech output
type AlexaSpeech struct {
	Type string `json:"type"` // PlainText or SSML
	Text string `json:"text,omitempty"`
	SSML string `json:"ssml,omitempty"`
}

// AlexaCard represents a card
type AlexaCard struct {
	Type    string `json:"type"` // Simple, Standard, LinkAccount
	Title   string `json:"title,omitempty"`
	Content string `json:"content,omitempty"`
	Text    string `json:"text,omitempty"`
}

func main() {
	lambda.Start(Handle)
}

// Handle processes incoming Alexa skill requests
func Handle(ctx context.Context, event events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	// Verify Alexa request signature (production requirement)
	// In Lambda behind API Gateway, signature verification is typically done at API Gateway level
	// For this implementation, we'll skip signature verification but note it should be added

	// Parse Alexa request
	var alexaReq AlexaRequest
	if err := json.Unmarshal([]byte(event.Body), &alexaReq); err != nil {
		return createErrorResponse("Invalid request"), nil
	}

	// Check for access token (OAuth account linking)
	if alexaReq.Session.User.AccessToken == "" {
		return createLinkAccountResponse(), nil
	}

	accessToken := alexaReq.Session.User.AccessToken

	// Route based on request type
	switch alexaReq.Request.Type {
	case "LaunchRequest":
		return handleLaunchRequest(ctx, alexaReq), nil
	case "IntentRequest":
		return handleIntentRequest(ctx, alexaReq, accessToken), nil
	case "SessionEndedRequest":
		return handleSessionEndedRequest(ctx, alexaReq), nil
	default:
		return createErrorResponse("Unsupported request type"), nil
	}
}

// handleLaunchRequest handles skill launch
func handleLaunchRequest(ctx context.Context, req AlexaRequest) events.APIGatewayV2HTTPResponse {
	response := AlexaResponse{
		Version: "1.0",
		Response: AlexaResponseBody{
			OutputSpeech: &AlexaSpeech{
				Type: "SSML",
				SSML: "<speak>Welcome to Fusion Helper! You can ask me about your CRM data, like \"show my contacts\" or \"what helpers are available\".</speak>",
			},
			Card: &AlexaCard{
				Type:    "Simple",
				Title:   "Welcome to Fusion Helper",
				Content: "Ask me about your CRM data and automations!",
			},
			ShouldEndSession: false,
		},
	}

	body, _ := json.Marshal(response)
	return events.APIGatewayV2HTTPResponse{
		StatusCode: 200,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: string(body),
	}
}

// handleIntentRequest handles intent requests
func handleIntentRequest(ctx context.Context, req AlexaRequest, accessToken string) events.APIGatewayV2HTTPResponse {
	if req.Request.Intent == nil {
		return createErrorResponse("No intent provided")
	}

	switch req.Request.Intent.Name {
	case "QueryDataIntent":
		return handleQueryDataIntent(ctx, req, accessToken)
	case "InvokeHelperIntent":
		return handleInvokeHelperIntent(ctx, req, accessToken)
	case "GetSummaryIntent":
		return handleGetSummaryIntent(ctx, req, accessToken)
	case "AMAZON.HelpIntent":
		return handleHelpIntent(ctx, req)
	case "AMAZON.StopIntent", "AMAZON.CancelIntent":
		return handleStopIntent(ctx, req)
	default:
		return createErrorResponse("I don't understand that request")
	}
}

// handleQueryDataIntent handles data queries
func handleQueryDataIntent(ctx context.Context, req AlexaRequest, accessToken string) events.APIGatewayV2HTTPResponse {
	query := ""
	if req.Request.Intent.Slots != nil {
		if querySlot, ok := req.Request.Intent.Slots["query"]; ok {
			query = querySlot.Value
		}
	}

	if query == "" {
		return createSpeechResponse("What would you like to know about your CRM data?", false)
	}

	// Get or create conversation
	conversationID, err := getOrCreateAlexaConversation(ctx, req.Session.User.UserID)
	if err != nil {
		return createSpeechResponse("Failed to create conversation", true)
	}

	// Process query with MCP service
	mcpService := services.NewMCPService(ctx)
	responseText := processAlexaQuery(ctx, mcpService, conversationID, query, accessToken)

	return createSpeechResponse(responseText, true)
}

// handleInvokeHelperIntent handles helper invocation
func handleInvokeHelperIntent(ctx context.Context, req AlexaRequest, accessToken string) events.APIGatewayV2HTTPResponse {
	helperType := ""
	action := ""

	if req.Request.Intent.Slots != nil {
		if slot, ok := req.Request.Intent.Slots["helper_type"]; ok {
			helperType = slot.Value
		}
		if slot, ok := req.Request.Intent.Slots["action"]; ok {
			action = slot.Value
		}
	}

	if helperType == "" || action == "" {
		return createSpeechResponse("What helper would you like to run?", false)
	}

	query := fmt.Sprintf("Run %s helper with action: %s", helperType, action)

	// Get or create conversation
	conversationID, err := getOrCreateAlexaConversation(ctx, req.Session.User.UserID)
	if err != nil {
		return createSpeechResponse("Failed to create conversation", true)
	}

	// Process with MCP service
	mcpService := services.NewMCPService(ctx)
	responseText := processAlexaQuery(ctx, mcpService, conversationID, query, accessToken)

	return createSpeechResponse(responseText, true)
}

// handleGetSummaryIntent handles summary requests
func handleGetSummaryIntent(ctx context.Context, req AlexaRequest, accessToken string) events.APIGatewayV2HTTPResponse {
	query := "Give me a summary of my CRM data and recent activity"

	conversationID, err := getOrCreateAlexaConversation(ctx, req.Session.User.UserID)
	if err != nil {
		return createSpeechResponse("Failed to get summary", true)
	}

	mcpService := services.NewMCPService(ctx)
	responseText := processAlexaQuery(ctx, mcpService, conversationID, query, accessToken)

	return createSpeechResponse(responseText, true)
}

// handleHelpIntent handles help requests
func handleHelpIntent(ctx context.Context, req AlexaRequest) events.APIGatewayV2HTTPResponse {
	return createSpeechResponse("You can ask me to show your contacts, list helpers, or get a summary of your data. What would you like to do?", false)
}

// handleStopIntent handles stop/cancel
func handleStopIntent(ctx context.Context, req AlexaRequest) events.APIGatewayV2HTTPResponse {
	return createSpeechResponse("Goodbye!", true)
}

// handleSessionEndedRequest handles session end
func handleSessionEndedRequest(ctx context.Context, req AlexaRequest) events.APIGatewayV2HTTPResponse {
	// Nothing to do, just return empty response
	response := AlexaResponse{
		Version: "1.0",
		Response: AlexaResponseBody{
			ShouldEndSession: true,
		},
	}

	body, _ := json.Marshal(response)
	return events.APIGatewayV2HTTPResponse{
		StatusCode: 200,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: string(body),
	}
}

// createSpeechResponse creates a speech response
func createSpeechResponse(text string, shouldEnd bool) events.APIGatewayV2HTTPResponse {
	response := AlexaResponse{
		Version: "1.0",
		Response: AlexaResponseBody{
			OutputSpeech: &AlexaSpeech{
				Type: "SSML",
				SSML: fmt.Sprintf("<speak>%s</speak>", text),
			},
			ShouldEndSession: shouldEnd,
		},
	}

	body, _ := json.Marshal(response)
	return events.APIGatewayV2HTTPResponse{
		StatusCode: 200,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: string(body),
	}
}

// createLinkAccountResponse creates a link account card response
func createLinkAccountResponse() events.APIGatewayV2HTTPResponse {
	response := AlexaResponse{
		Version: "1.0",
		Response: AlexaResponseBody{
			OutputSpeech: &AlexaSpeech{
				Type: "PlainText",
				Text: "Please link your Fusion Helper account using the Alexa app.",
			},
			Card: &AlexaCard{
				Type: "LinkAccount",
			},
			ShouldEndSession: true,
		},
	}

	body, _ := json.Marshal(response)
	return events.APIGatewayV2HTTPResponse{
		StatusCode: 200,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: string(body),
	}
}

// createErrorResponse creates an error response
func createErrorResponse(message string) events.APIGatewayV2HTTPResponse {
	return createSpeechResponse(message, true)
}

// getOrCreateAlexaConversation gets or creates conversation for Alexa user
func getOrCreateAlexaConversation(ctx context.Context, alexaUserID string) (string, error) {
	// For simplicity, create new conversation each session
	// In production, you'd want to maintain session state and save to DynamoDB
	conversationID := "conv:" + uuid.New().String()

	// In a full implementation, we would:
	// 1. Load AWS config
	// 2. Create DynamoDB client
	// 3. Query for existing conversation
	// 4. If not found, create new conversation
	// 5. Save conversation to DynamoDB

	return conversationID, nil
}

// processAlexaQuery processes a query with the MCP service
func processAlexaQuery(ctx context.Context, mcpService *services.MCPService, conversationID, query, accessToken string) string {
	// Build Groq messages
	groqMessages := []types.GroqMessage{
		{Role: "system", Content: "You are a voice assistant for MyFusionHelper. Keep responses brief and conversational for voice output."},
		{Role: "user", Content: query},
	}

	// Get tool definitions
	tools := mcpService.GetToolDefinitions()

	// Call Groq API (simplified - no streaming for Alexa)
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
		return "Failed to process request. Please try again."
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

	// Handle tool calls (simplified)
	if len(choice.Message.ToolCalls) > 0 {
		for _, tc := range choice.Message.ToolCalls {
			_, err := mcpService.ExecuteTool(ctx, tc, accessToken)
			if err != nil {
				content += fmt.Sprintf(" I encountered an error executing %s.", tc.Function.Name)
			} else {
				// Summarize tool result for voice
				content += fmt.Sprintf(" I checked your %s.", tc.Function.Name)
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
