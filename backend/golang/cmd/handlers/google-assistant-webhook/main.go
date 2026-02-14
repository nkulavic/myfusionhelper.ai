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

// GoogleAssistantRequest represents Actions on Google request
type GoogleAssistantRequest struct {
	Handler GoogleHandler `json:"handler"`
	Intent  GoogleIntent  `json:"intent"`
	Scene   GoogleScene   `json:"scene"`
	Session GoogleSession `json:"session"`
	User    GoogleUser    `json:"user"`
	Device  GoogleDevice  `json:"device"`
	Context GoogleContext `json:"context"`
}

// GoogleHandler represents the handler info
type GoogleHandler struct {
	Name string `json:"name"`
}

// GoogleIntent represents the intent
type GoogleIntent struct {
	Name   string                 `json:"name"`
	Params map[string]interface{} `json:"params"`
	Query  string                 `json:"query"`
}

// GoogleScene represents the scene
type GoogleScene struct {
	Name           string                 `json:"name"`
	SlotFillingStatus string              `json:"slotFillingStatus"`
	Slots          map[string]interface{} `json:"slots"`
}

// GoogleSession represents session data
type GoogleSession struct {
	ID         string                 `json:"id"`
	Params     map[string]interface{} `json:"params"`
	TypeOverrides []interface{}        `json:"typeOverrides"`
	LanguageCode string                `json:"languageCode"`
}

// GoogleUser represents user data
type GoogleUser struct {
	Locale            string                 `json:"locale"`
	Params            map[string]interface{} `json:"params"`
	AccountLinkingStatus string              `json:"accountLinkingStatus"`
	VerificationStatus string                `json:"verificationStatus"`
	PackageEntitlements []interface{}        `json:"packageEntitlements"`
	LastSeenTime       string                `json:"lastSeenTime"`
}

// GoogleDevice represents device info
type GoogleDevice struct {
	Capabilities []string `json:"capabilities"`
}

// GoogleContext represents context info
type GoogleContext struct {
	Media []interface{} `json:"media"`
}

// GoogleAssistantResponse represents Actions on Google response
type GoogleAssistantResponse struct {
	Session GoogleSessionResponse `json:"session"`
	Prompt  GooglePrompt          `json:"prompt"`
	Scene   *GoogleSceneResponse  `json:"scene,omitempty"`
}

// GoogleSessionResponse represents session in response
type GoogleSessionResponse struct {
	ID     string                 `json:"id"`
	Params map[string]interface{} `json:"params"`
}

// GooglePrompt represents the prompt
type GooglePrompt struct {
	Override    bool                `json:"override"`
	FirstSimple GoogleSimpleResponse `json:"firstSimple"`
	Content     *GoogleContent      `json:"content,omitempty"`
}

// GoogleSimpleResponse represents simple speech
type GoogleSimpleResponse struct {
	Speech string `json:"speech"`
	Text   string `json:"text"`
}

// GoogleContent represents rich content
type GoogleContent struct {
	Card *GoogleCard `json:"card,omitempty"`
}

// GoogleCard represents a card
type GoogleCard struct {
	Title    string `json:"title"`
	Subtitle string `json:"subtitle"`
	Text     string `json:"text"`
}

// GoogleSceneResponse represents scene transition
type GoogleSceneResponse struct {
	Name  string `json:"name"`
	SlotFillingStatus string `json:"slotFillingStatus"`
	Next  *GoogleNext `json:"next,omitempty"`
}

// GoogleNext represents next action
type GoogleNext struct {
	Name string `json:"name"`
}

func main() {
	lambda.Start(Handle)
}

// Handle processes incoming Google Assistant webhook requests
func Handle(ctx context.Context, event events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	// Parse Google Assistant request
	var googleReq GoogleAssistantRequest
	if err := json.Unmarshal([]byte(event.Body), &googleReq); err != nil {
		return createErrorResponse("Invalid request"), nil
	}

	// Check account linking status
	if googleReq.User.AccountLinkingStatus != "LINKED" {
		return createAccountLinkingResponse(), nil
	}

	// Get access token from session params (set during account linking)
	accessToken := ""
	if googleReq.Session.Params != nil {
		if token, ok := googleReq.Session.Params["access_token"].(string); ok {
			accessToken = token
		}
	}

	if accessToken == "" {
		return createSimpleResponse("Please link your Fusion Helper account in the Google Home app.", true), nil
	}

	// Route based on handler name
	switch googleReq.Handler.Name {
	case "welcome":
		return handleWelcome(ctx, googleReq), nil
	case "query_data":
		return handleQueryData(ctx, googleReq, accessToken), nil
	case "invoke_helper":
		return handleInvokeHelper(ctx, googleReq, accessToken), nil
	case "get_summary":
		return handleGetSummary(ctx, googleReq, accessToken), nil
	default:
		return createSimpleResponse("I didn't understand that. Try asking me about your CRM data.", false), nil
	}
}

// handleWelcome handles welcome/launch
func handleWelcome(ctx context.Context, req GoogleAssistantRequest) events.APIGatewayV2HTTPResponse {
	response := GoogleAssistantResponse{
		Session: GoogleSessionResponse{
			ID:     req.Session.ID,
			Params: req.Session.Params,
		},
		Prompt: GooglePrompt{
			Override: false,
			FirstSimple: GoogleSimpleResponse{
				Speech: "Welcome to Fusion Helper! You can ask me about your CRM data, like show my contacts or list my helpers.",
				Text:   "Welcome to Fusion Helper",
			},
			Content: &GoogleContent{
				Card: &GoogleCard{
					Title:    "Fusion Helper",
					Subtitle: "Your CRM Assistant",
					Text:     "Ask me about your contacts, helpers, and data.",
				},
			},
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

// handleQueryData handles data queries
func handleQueryData(ctx context.Context, req GoogleAssistantRequest, accessToken string) events.APIGatewayV2HTTPResponse {
	query := req.Intent.Query
	if query == "" {
		return createSimpleResponse("What would you like to know about your CRM data?", false)
	}

	// Get or create conversation
	conversationID := "conv:" + uuid.New().String()

	// Process query with MCP service
	mcpService := services.NewMCPService(ctx)
	responseText := processGoogleQuery(ctx, mcpService, conversationID, query, accessToken)

	return createSimpleResponse(responseText, false)
}

// handleInvokeHelper handles helper invocation
func handleInvokeHelper(ctx context.Context, req GoogleAssistantRequest, accessToken string) events.APIGatewayV2HTTPResponse {
	query := req.Intent.Query
	if query == "" {
		return createSimpleResponse("What helper would you like to run?", false)
	}

	conversationID := "conv:" + uuid.New().String()
	mcpService := services.NewMCPService(ctx)
	responseText := processGoogleQuery(ctx, mcpService, conversationID, query, accessToken)

	return createSimpleResponse(responseText, false)
}

// handleGetSummary handles summary requests
func handleGetSummary(ctx context.Context, req GoogleAssistantRequest, accessToken string) events.APIGatewayV2HTTPResponse {
	query := "Give me a summary of my CRM data and recent activity"
	conversationID := "conv:" + uuid.New().String()

	mcpService := services.NewMCPService(ctx)
	responseText := processGoogleQuery(ctx, mcpService, conversationID, query, accessToken)

	return createSimpleResponse(responseText, false)
}

// createSimpleResponse creates a simple speech response
func createSimpleResponse(text string, shouldEnd bool) events.APIGatewayV2HTTPResponse {
	response := GoogleAssistantResponse{
		Prompt: GooglePrompt{
			Override: false,
			FirstSimple: GoogleSimpleResponse{
				Speech: text,
				Text:   text,
			},
		},
	}

	if shouldEnd {
		response.Scene = &GoogleSceneResponse{
			Name: "actions.scene.END_CONVERSATION",
		}
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

// createAccountLinkingResponse creates account linking response
func createAccountLinkingResponse() events.APIGatewayV2HTTPResponse {
	response := GoogleAssistantResponse{
		Prompt: GooglePrompt{
			Override: false,
			FirstSimple: GoogleSimpleResponse{
				Speech: "To use Fusion Helper, you need to link your account. Please use the Google Home app to complete account linking.",
				Text:   "Account linking required",
			},
		},
		Scene: &GoogleSceneResponse{
			Name: "actions.scene.END_CONVERSATION",
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
	return createSimpleResponse(message, true)
}

// processGoogleQuery processes a query with the MCP service
func processGoogleQuery(ctx context.Context, mcpService *services.MCPService, conversationID, query, accessToken string) string {
	// Build Groq messages
	groqMessages := []types.GroqMessage{
		{Role: "system", Content: "You are a voice assistant for MyFusionHelper. Keep responses brief and conversational for voice output."},
		{Role: "user", Content: query},
	}

	// Get tool definitions
	tools := mcpService.GetToolDefinitions()

	// Call Groq API (simplified - no streaming for Google Assistant)
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
				content += fmt.Sprintf(" I encountered an error with %s.", tc.Function.Name)
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
