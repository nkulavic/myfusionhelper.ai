package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/myfusionhelper/api/internal/types"
)

// MCPService handles execution of MCP tools by calling existing backend APIs
type MCPService struct {
	httpClient  *http.Client
	baseURL     string
	groqAPIKey  string
}

// NewMCPService creates a new MCP service instance
func NewMCPService(ctx context.Context) *MCPService {
	baseURL := os.Getenv("API_BASE_URL")
	if baseURL == "" {
		baseURL = "https://dev.api.myfusionhelper.ai"
	}

	groqAPIKey := os.Getenv("GROQ_API_KEY")

	return &MCPService{
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
		baseURL:    baseURL,
		groqAPIKey: groqAPIKey,
	}
}

// GetToolDefinitions returns all available MCP tool definitions in Groq/OpenAI format
func (s *MCPService) GetToolDefinitions() []types.GroqTool {
	return []types.GroqTool{
		// CRM Data Query Tool
		{
			Type: "function",
			Function: types.GroqFunctionDef{
				Name:        "query_crm_data",
				Description: "Query CRM data using natural language or filters via Data Explorer API. Returns paginated results from Infusionsoft CRM.",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"nl_query": map[string]interface{}{
							"type":        "string",
							"description": "Natural language query to filter CRM data (e.g., 'contacts created last 30 days')",
						},
						"connection_id": map[string]interface{}{
							"type":        "string",
							"description": "Platform connection ID (required - use get_connections to find)",
						},
						"limit": map[string]interface{}{
							"type":        "integer",
							"description": "Maximum number of results to return",
							"default":     50,
						},
						"offset": map[string]interface{}{
							"type":        "integer",
							"description": "Number of results to skip for pagination",
							"default":     0,
						},
					},
					"required": []string{"connection_id"},
				},
			},
		},
		// Get Contacts Tool
		{
			Type: "function",
			Function: types.GroqFunctionDef{
				Name:        "get_contacts",
				Description: "Get list of contacts from CRM with optional filters. Returns paginated contact list.",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"connection_id": map[string]interface{}{
							"type":        "string",
							"description": "Platform connection ID (required)",
						},
						"limit": map[string]interface{}{
							"type":        "integer",
							"description": "Maximum number of contacts to return",
							"default":     50,
						},
						"offset": map[string]interface{}{
							"type":        "integer",
							"description": "Number of contacts to skip for pagination",
							"default":     0,
						},
						"filter": map[string]interface{}{
							"type":        "string",
							"description": "Optional filter query (e.g., 'email contains @example.com')",
						},
					},
					"required": []string{"connection_id"},
				},
			},
		},
		// Get Contact Detail Tool
		{
			Type: "function",
			Function: types.GroqFunctionDef{
				Name:        "get_contact_detail",
				Description: "Get detailed information for a single contact by ID. Returns full contact record with all fields.",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"connection_id": map[string]interface{}{
							"type":        "string",
							"description": "Platform connection ID (required)",
						},
						"contact_id": map[string]interface{}{
							"type":        "string",
							"description": "Unique ID of the contact to retrieve (required)",
						},
					},
					"required": []string{"connection_id", "contact_id"},
				},
			},
		},
		// Invoke Helper Tool
		{
			Type: "function",
			Function: types.GroqFunctionDef{
				Name:        "invoke_helper",
				Description: "Execute a Fusion helper with specified input data. Returns the helper execution result.",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"helper_id": map[string]interface{}{
							"type":        "string",
							"description": "Unique ID of the helper to execute (required)",
						},
						"input_data": map[string]interface{}{
							"type":        "object",
							"description": "Input data for the helper execution (JSON object)",
						},
						"connection_id": map[string]interface{}{
							"type":        "string",
							"description": "Platform connection ID (optional, if helper needs CRM access)",
						},
					},
					"required": []string{"helper_id"},
				},
			},
		},
		// List Helpers Tool
		{
			Type: "function",
			Function: types.GroqFunctionDef{
				Name:        "list_helpers",
				Description: "List all available Fusion helpers for the account. Returns helper list with names, IDs, and descriptions.",
				Parameters: map[string]interface{}{
					"type":       "object",
					"properties": map[string]interface{}{},
				},
			},
		},
		// Get Helper Config Tool
		{
			Type: "function",
			Function: types.GroqFunctionDef{
				Name:        "get_helper_config",
				Description: "Get configuration details for a specific helper including input schema and settings.",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"helper_id": map[string]interface{}{
							"type":        "string",
							"description": "Unique ID of the helper (required)",
						},
					},
					"required": []string{"helper_id"},
				},
			},
		},
		// Get Platform Connections Tool
		{
			Type: "function",
			Function: types.GroqFunctionDef{
				Name:        "get_connections",
				Description: "Get list of platform connections for the user. Returns available CRM connections with IDs.",
				Parameters: map[string]interface{}{
					"type":       "object",
					"properties": map[string]interface{}{},
				},
			},
		},
	}
}

// ExecuteTool executes a tool call by routing to the appropriate handler
func (s *MCPService) ExecuteTool(ctx context.Context, toolCall types.GroqToolCall, accessToken string) (string, error) {
	// Parse the arguments
	var args map[string]interface{}
	if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &args); err != nil {
		return "", fmt.Errorf("failed to parse tool arguments: %w", err)
	}

	// Route to appropriate tool handler
	switch toolCall.Function.Name {
	case "query_crm_data":
		return s.queryCRMData(ctx, args, accessToken)
	case "get_contacts":
		return s.getContacts(ctx, args, accessToken)
	case "get_contact_detail":
		return s.getContactDetail(ctx, args, accessToken)
	case "invoke_helper":
		return s.invokeHelper(ctx, args, accessToken)
	case "list_helpers":
		return s.listHelpers(ctx, accessToken)
	case "get_helper_config":
		return s.getHelperConfig(ctx, args, accessToken)
	case "get_connections":
		return s.getConnections(ctx, accessToken)
	default:
		return "", fmt.Errorf("unknown tool: %s", toolCall.Function.Name)
	}
}

// queryCRMData calls POST /data/query for natural language CRM data queries
func (s *MCPService) queryCRMData(ctx context.Context, args map[string]interface{}, accessToken string) (string, error) {
	nlQuery, ok := args["nl_query"].(string)
	if !ok || nlQuery == "" {
		return "", fmt.Errorf("nl_query is required")
	}

	connectionID, ok := args["connection_id"].(string)
	if !ok || connectionID == "" {
		return "", fmt.Errorf("connection_id is required")
	}

	// Optional limit parameter (default 50)
	limit := 50
	if l, ok := args["limit"].(float64); ok {
		limit = int(l)
	}

	body := map[string]interface{}{
		"nl_query":      nlQuery,
		"connection_id": connectionID,
		"limit":         limit,
	}

	bodyJSON, err := json.Marshal(body)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request body: %w", err)
	}

	resp, err := s.makeAPIRequest(ctx, "POST", "/data/query", bodyJSON, accessToken)
	if err != nil {
		return "", fmt.Errorf("data query failed: %w", err)
	}
	return s.formatResponse(resp), nil
}

// getContacts calls POST /data/query to list contacts from a CRM connection
func (s *MCPService) getContacts(ctx context.Context, args map[string]interface{}, accessToken string) (string, error) {
	connectionID, ok := args["connection_id"].(string)
	if !ok || connectionID == "" {
		return "", fmt.Errorf("connection_id is required")
	}

	// Optional pagination parameters
	page := 1
	if p, ok := args["page"].(float64); ok {
		page = int(p)
	}

	pageSize := 50
	if ps, ok := args["page_size"].(float64); ok {
		pageSize = int(ps)
	}

	// Use data/query endpoint with contacts object type
	body := map[string]interface{}{
		"connection_id": connectionID,
		"object_type":   "contacts",
		"page":          page,
		"page_size":     pageSize,
	}

	// Add optional filter if provided
	if filter, ok := args["filter"].(string); ok && filter != "" {
		body["nl_query"] = filter
	}

	bodyJSON, err := json.Marshal(body)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request body: %w", err)
	}

	resp, err := s.makeAPIRequest(ctx, "POST", "/data/query", bodyJSON, accessToken)
	if err != nil {
		return "", fmt.Errorf("get contacts failed: %w", err)
	}
	return s.formatContactsResponse(resp), nil
}

// getContactDetail calls GET /data/record/{connectionId}/{objectType}/{recordId} to get detailed contact information
func (s *MCPService) getContactDetail(ctx context.Context, args map[string]interface{}, accessToken string) (string, error) {
	contactID, ok := args["contact_id"].(string)
	if !ok || contactID == "" {
		return "", fmt.Errorf("contact_id is required")
	}

	connectionID, ok := args["connection_id"].(string)
	if !ok || connectionID == "" {
		return "", fmt.Errorf("connection_id is required")
	}

	// Use the data/record endpoint with the correct path format
	path := fmt.Sprintf("/data/record/%s/contacts/%s", connectionID, contactID)

	resp, err := s.makeAPIRequest(ctx, "GET", path, nil, accessToken)
	if err != nil {
		return "", fmt.Errorf("get contact detail failed: %w", err)
	}
	return s.formatResponse(resp), nil
}

// invokeHelper calls POST /helpers/{id}/execute to run a helper
func (s *MCPService) invokeHelper(ctx context.Context, args map[string]interface{}, accessToken string) (string, error) {
	helperID, ok := args["helper_id"].(string)
	if !ok || helperID == "" {
		return "", fmt.Errorf("helper_id is required")
	}

	// Extract input parameters (optional)
	input := make(map[string]interface{})
	if i, ok := args["input"].(map[string]interface{}); ok {
		input = i
	}

	// Optional connection_id
	if connID, ok := args["connection_id"].(string); ok && connID != "" {
		input["connection_id"] = connID
	}

	// Optional contact_id
	if contactID, ok := args["contact_id"].(string); ok && contactID != "" {
		input["contact_id"] = contactID
	}

	body := map[string]interface{}{
		"input": input,
	}

	bodyJSON, err := json.Marshal(body)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request body: %w", err)
	}

	path := fmt.Sprintf("/helpers/%s/execute", helperID)

	resp, err := s.makeAPIRequest(ctx, "POST", path, bodyJSON, accessToken)
	if err != nil {
		return "", fmt.Errorf("helper execution failed: %w", err)
	}
	return s.formatResponse(resp), nil
}

// listHelpers calls GET /helpers to list available helpers
func (s *MCPService) listHelpers(ctx context.Context, accessToken string) (string, error) {
	resp, err := s.makeAPIRequest(ctx, "GET", "/helpers", nil, accessToken)
	if err != nil {
		return "", fmt.Errorf("list helpers failed: %w", err)
	}
	return s.formatHelpersListResponse(resp), nil
}

// getHelperConfig calls GET /helpers/{id} to get helper configuration
func (s *MCPService) getHelperConfig(ctx context.Context, args map[string]interface{}, accessToken string) (string, error) {
	helperID, ok := args["helper_id"].(string)
	if !ok || helperID == "" {
		return "", fmt.Errorf("helper_id is required")
	}

	path := fmt.Sprintf("/helpers/%s", helperID)

	resp, err := s.makeAPIRequest(ctx, "GET", path, nil, accessToken)
	if err != nil {
		return "", fmt.Errorf("get helper config failed: %w", err)
	}
	return s.formatResponse(resp), nil
}

// getConnections calls GET /platform-connections
func (s *MCPService) getConnections(ctx context.Context, accessToken string) (string, error) {
	resp, err := s.makeAPIRequest(ctx, "GET", "/platform-connections", nil, accessToken)
	if err != nil {
		return "", fmt.Errorf("get connections failed: %w", err)
	}
	return s.formatConnectionsResponse(resp), nil
}

// makeAPIRequest makes an HTTP request to the backend API
func (s *MCPService) makeAPIRequest(ctx context.Context, method, path string, body []byte, accessToken string) (map[string]interface{}, error) {
	url := s.baseURL + path

	var reqBody io.Reader
	if body != nil {
		reqBody = strings.NewReader(string(body))
	}

	req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return result, nil
}

// formatResponse converts the API response to a JSON string for the LLM
func (s *MCPService) formatResponse(resp map[string]interface{}) string {
	// Try to get the data field if it exists
	if data, ok := resp["data"]; ok {
		jsonBytes, err := json.MarshalIndent(data, "", "  ")
		if err == nil {
			return string(jsonBytes)
		}
	}

	// Fall back to full response
	jsonBytes, err := json.MarshalIndent(resp, "", "  ")
	if err != nil {
		return fmt.Sprintf("%v", resp)
	}
	return string(jsonBytes)
}

// formatConnectionsResponse condenses connection data to reduce token usage
func (s *MCPService) formatConnectionsResponse(resp map[string]interface{}) string {
	data, ok := resp["data"]
	if !ok {
		return s.formatResponse(resp)
	}

	dataMap, ok := data.(map[string]interface{})
	if !ok {
		return s.formatResponse(resp)
	}

	connections, ok := dataMap["connections"].([]interface{})
	if !ok {
		return s.formatResponse(resp)
	}

	// Extract only essential fields from each connection
	condensedConnections := make([]map[string]interface{}, 0, len(connections))
	for _, c := range connections {
		connection, ok := c.(map[string]interface{})
		if !ok {
			continue
		}

		condensed := map[string]interface{}{
			"connection_id": connection["connection_id"],
			"platform":      connection["platform"],
			"display_name":  connection["display_name"],
			"status":        connection["status"],
		}

		condensedConnections = append(condensedConnections, condensed)
	}

	result := map[string]interface{}{
		"connections": condensedConnections,
		"total":       len(condensedConnections),
	}

	jsonBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return s.formatResponse(resp)
	}
	return string(jsonBytes)
}

// formatContactsResponse condenses contact list data to reduce token usage
func (s *MCPService) formatContactsResponse(resp map[string]interface{}) string {
	data, ok := resp["data"].(map[string]interface{})
	if !ok {
		return s.formatResponse(resp)
	}

	contacts, ok := data["contacts"].([]interface{})
	if !ok {
		return s.formatResponse(resp)
	}

	// Extract only essential fields from each contact
	condensedContacts := make([]map[string]interface{}, 0, len(contacts))
	for _, c := range contacts {
		contact, ok := c.(map[string]interface{})
		if !ok {
			continue
		}

		condensed := map[string]interface{}{
			"id":    contact["id"],
			"name":  contact["name"],
			"email": contact["email"],
		}

		// Include phone if available
		if phone, ok := contact["phone"].(string); ok && phone != "" {
			condensed["phone"] = phone
		}

		// Include tags if available
		if tags, ok := contact["tags"].([]interface{}); ok && len(tags) > 0 {
			condensed["tags"] = tags
		}

		condensedContacts = append(condensedContacts, condensed)
	}

	result := map[string]interface{}{
		"contacts":  condensedContacts,
		"total":     data["total"],
		"page":      data["page"],
		"page_size": data["page_size"],
		"has_more":  data["has_more"],
	}

	jsonBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return s.formatResponse(resp)
	}
	return string(jsonBytes)
}

// formatHelpersListResponse condenses helpers list to reduce token usage
func (s *MCPService) formatHelpersListResponse(resp map[string]interface{}) string {
	data, ok := resp["data"].(map[string]interface{})
	if !ok {
		return s.formatResponse(resp)
	}

	helpers, ok := data["helpers"].([]interface{})
	if !ok {
		return s.formatResponse(resp)
	}

	// Extract only essential fields from each helper
	condensedHelpers := make([]map[string]interface{}, 0, len(helpers))
	for _, h := range helpers {
		helper, ok := h.(map[string]interface{})
		if !ok {
			continue
		}

		condensed := map[string]interface{}{
			"helper_id":   helper["helper_id"],
			"name":        helper["name"],
			"description": helper["description"],
			"helper_type": helper["helper_type"],
			"category":    helper["category"],
			"enabled":     helper["enabled"],
		}

		condensedHelpers = append(condensedHelpers, condensed)
	}

	result := map[string]interface{}{
		"helpers": condensedHelpers,
		"total":   len(condensedHelpers),
	}

	jsonBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return s.formatResponse(resp)
	}
	return string(jsonBytes)
}
