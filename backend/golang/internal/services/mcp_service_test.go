package services

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/myfusionhelper/api/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGetToolDefinitions verifies all 7 tools are defined with correct schemas
func TestGetToolDefinitions(t *testing.T) {
	service := &MCPService{
		httpClient: &http.Client{},
		baseURL:    "http://test",
		groqAPIKey: "test-key",
	}

	tools := service.GetToolDefinitions()

	// Verify we have exactly 7 tools
	assert.Len(t, tools, 7, "Should have 7 tools defined")

	// Verify each tool exists and has correct structure
	toolMap := make(map[string]types.GroqTool)
	for _, tool := range tools {
		assert.Equal(t, "function", tool.Type, "Tool type should be 'function'")
		assert.NotEmpty(t, tool.Function.Name, "Tool name should not be empty")
		assert.NotEmpty(t, tool.Function.Description, "Tool description should not be empty")
		assert.NotNil(t, tool.Function.Parameters, "Tool parameters should not be nil")
		toolMap[tool.Function.Name] = tool
	}

	// Verify all expected tools are present
	expectedTools := []string{
		"query_crm_data",
		"get_contacts",
		"get_contact_detail",
		"invoke_helper",
		"list_helpers",
		"get_helper_config",
		"get_connections",
	}

	for _, expectedTool := range expectedTools {
		tool, exists := toolMap[expectedTool]
		assert.True(t, exists, "Tool %s should exist", expectedTool)
		if exists {
			assert.NotEmpty(t, tool.Function.Description, "Tool %s should have description", expectedTool)
		}
	}

	// Verify query_crm_data tool has correct parameters
	queryCRMTool := toolMap["query_crm_data"]
	params := queryCRMTool.Function.Parameters
	assert.Equal(t, "object", params["type"])

	properties, ok := params["properties"].(map[string]interface{})
	require.True(t, ok, "Properties should be a map")
	assert.Contains(t, properties, "nl_query")
	assert.Contains(t, properties, "connection_id")
	assert.Contains(t, properties, "limit")
	assert.Contains(t, properties, "offset")

	required, ok := params["required"].([]interface{})
	require.True(t, ok, "Required should be an array")
	assert.Contains(t, required, "connection_id")
}

// TestExecuteTool_UnknownTool verifies error handling for unknown tool names
func TestExecuteTool_UnknownTool(t *testing.T) {
	service := &MCPService{
		httpClient: &http.Client{},
		baseURL:    "http://test",
		groqAPIKey: "test-key",
	}

	toolCall := types.GroqToolCall{
		Function: types.GroqFunctionCall{
			Name:      "unknown_tool_name",
			Arguments: `{}`,
		},
	}

	result, err := service.ExecuteTool(context.Background(), toolCall, "fake-token")

	assert.Error(t, err)
	assert.Empty(t, result)
	assert.Contains(t, err.Error(), "unknown tool")
}

// TestExecuteTool_InvalidJSON verifies error handling for invalid JSON arguments
func TestExecuteTool_InvalidJSON(t *testing.T) {
	service := &MCPService{
		httpClient: &http.Client{},
		baseURL:    "http://test",
		groqAPIKey: "test-key",
	}

	toolCall := types.GroqToolCall{
		Function: types.GroqFunctionCall{
			Name:      "query_crm_data",
			Arguments: `{invalid json}`,
		},
	}

	result, err := service.ExecuteTool(context.Background(), toolCall, "fake-token")

	assert.Error(t, err)
	assert.Empty(t, result)
	assert.Contains(t, err.Error(), "failed to parse tool arguments")
}

// TestQueryCRMData verifies the query_crm_data tool handler
func TestQueryCRMData(t *testing.T) {
	// Create mock server
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/data/explorer/query", r.URL.Path)
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		// Verify request body
		var reqBody map[string]interface{}
		err := json.NewDecoder(r.Body).Decode(&reqBody)
		require.NoError(t, err)
		assert.Equal(t, "conn_123", reqBody["connection_id"])
		assert.Equal(t, "show contacts", reqBody["nl_query"])
		assert.Equal(t, float64(10), reqBody["limit"])

		// Send mock response
		response := map[string]interface{}{
			"success": true,
			"data": map[string]interface{}{
				"rows": []map[string]interface{}{
					{"id": "1", "name": "John Doe", "email": "john@example.com"},
					{"id": "2", "name": "Jane Smith", "email": "jane@example.com"},
				},
				"total_count": 2,
				"query_time":  "0.5s",
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer mockServer.Close()

	service := &MCPService{
		httpClient: &http.Client{},
		baseURL:    mockServer.URL,
		groqAPIKey: "test-key",
	}

	toolCall := types.GroqToolCall{
		Function: types.GroqFunctionCall{
			Name:      "query_crm_data",
			Arguments: `{"nl_query": "show contacts", "connection_id": "conn_123", "limit": 10}`,
		},
	}

	result, err := service.ExecuteTool(context.Background(), toolCall, "test-token")

	require.NoError(t, err)
	assert.Contains(t, result, "John Doe")
	assert.Contains(t, result, "Jane Smith")
	assert.Contains(t, result, "john@example.com")
}

// TestQueryCRMData_MissingConnectionID verifies error handling for missing connection_id
func TestQueryCRMData_MissingConnectionID(t *testing.T) {
	service := &MCPService{
		httpClient: &http.Client{},
		baseURL:    "http://test",
		groqAPIKey: "test-key",
	}

	toolCall := types.GroqToolCall{
		Function: types.GroqFunctionCall{
			Name:      "query_crm_data",
			Arguments: `{"nl_query": "show contacts"}`,
		},
	}

	result, err := service.ExecuteTool(context.Background(), toolCall, "test-token")

	assert.Error(t, err)
	assert.Empty(t, result)
	assert.Contains(t, err.Error(), "connection_id is required")
}

// TestGetContacts verifies the get_contacts tool handler
func TestGetContacts(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/contacts", r.URL.Path)
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

		// Verify query parameters
		query := r.URL.Query()
		assert.Equal(t, "conn_123", query.Get("connection_id"))
		assert.Equal(t, "25", query.Get("limit"))
		assert.Equal(t, "10", query.Get("offset"))

		response := map[string]interface{}{
			"success": true,
			"data": map[string]interface{}{
				"contacts": []map[string]interface{}{
					{"id": "1", "first_name": "John", "last_name": "Doe"},
					{"id": "2", "first_name": "Jane", "last_name": "Smith"},
				},
				"total": 2,
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer mockServer.Close()

	service := &MCPService{
		httpClient: &http.Client{},
		baseURL:    mockServer.URL,
		groqAPIKey: "test-key",
	}

	toolCall := types.GroqToolCall{
		Function: types.GroqFunctionCall{
			Name:      "get_contacts",
			Arguments: `{"connection_id": "conn_123", "limit": 25, "offset": 10}`,
		},
	}

	result, err := service.ExecuteTool(context.Background(), toolCall, "test-token")

	require.NoError(t, err)
	assert.Contains(t, result, "John")
	assert.Contains(t, result, "Jane")
}

// TestGetContactDetail verifies the get_contact_detail tool handler
func TestGetContactDetail(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/contacts/contact_456", r.URL.Path)
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

		query := r.URL.Query()
		assert.Equal(t, "conn_123", query.Get("connection_id"))

		response := map[string]interface{}{
			"success": true,
			"data": map[string]interface{}{
				"id":         "contact_456",
				"first_name": "John",
				"last_name":  "Doe",
				"email":      "john@example.com",
				"phone":      "+1234567890",
				"tags":       []string{"customer", "vip"},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer mockServer.Close()

	service := &MCPService{
		httpClient: &http.Client{},
		baseURL:    mockServer.URL,
		groqAPIKey: "test-key",
	}

	toolCall := types.GroqToolCall{
		Function: types.GroqFunctionCall{
			Name:      "get_contact_detail",
			Arguments: `{"connection_id": "conn_123", "contact_id": "contact_456"}`,
		},
	}

	result, err := service.ExecuteTool(context.Background(), toolCall, "test-token")

	require.NoError(t, err)
	assert.Contains(t, result, "contact_456")
	assert.Contains(t, result, "john@example.com")
	assert.Contains(t, result, "customer")
}

// TestGetContactDetail_MissingContactID verifies error handling
func TestGetContactDetail_MissingContactID(t *testing.T) {
	service := &MCPService{
		httpClient: &http.Client{},
		baseURL:    "http://test",
		groqAPIKey: "test-key",
	}

	toolCall := types.GroqToolCall{
		Function: types.GroqFunctionCall{
			Name:      "get_contact_detail",
			Arguments: `{"connection_id": "conn_123"}`,
		},
	}

	result, err := service.ExecuteTool(context.Background(), toolCall, "test-token")

	assert.Error(t, err)
	assert.Empty(t, result)
	assert.Contains(t, err.Error(), "contact_id is required")
}

// TestInvokeHelper verifies the invoke_helper tool handler
func TestInvokeHelper(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/helpers/invoke", r.URL.Path)
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

		var reqBody map[string]interface{}
		err := json.NewDecoder(r.Body).Decode(&reqBody)
		require.NoError(t, err)
		assert.Equal(t, "helper_789", reqBody["helper_id"])
		assert.Equal(t, "conn_123", reqBody["connection_id"])

		inputData := reqBody["input_data"].(map[string]interface{})
		assert.Equal(t, "contact_456", inputData["contact_id"])

		response := map[string]interface{}{
			"success": true,
			"data": map[string]interface{}{
				"execution_id": "exec_999",
				"status":       "success",
				"output": map[string]interface{}{
					"message": "Helper executed successfully",
					"result":  "Tag applied to contact",
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer mockServer.Close()

	service := &MCPService{
		httpClient: &http.Client{},
		baseURL:    mockServer.URL,
		groqAPIKey: "test-key",
	}

	toolCall := types.GroqToolCall{
		Function: types.GroqFunctionCall{
			Name:      "invoke_helper",
			Arguments: `{"helper_id": "helper_789", "connection_id": "conn_123", "input_data": {"contact_id": "contact_456"}}`,
		},
	}

	result, err := service.ExecuteTool(context.Background(), toolCall, "test-token")

	require.NoError(t, err)
	assert.Contains(t, result, "exec_999")
	assert.Contains(t, result, "success")
	assert.Contains(t, result, "Tag applied to contact")
}

// TestInvokeHelper_MissingHelperID verifies error handling
func TestInvokeHelper_MissingHelperID(t *testing.T) {
	service := &MCPService{
		httpClient: &http.Client{},
		baseURL:    "http://test",
		groqAPIKey: "test-key",
	}

	toolCall := types.GroqToolCall{
		Function: types.GroqFunctionCall{
			Name:      "invoke_helper",
			Arguments: `{"input_data": {"contact_id": "contact_456"}}`,
		},
	}

	result, err := service.ExecuteTool(context.Background(), toolCall, "test-token")

	assert.Error(t, err)
	assert.Empty(t, result)
	assert.Contains(t, err.Error(), "helper_id is required")
}

// TestListHelpers verifies the list_helpers tool handler
func TestListHelpers(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/helpers", r.URL.Path)
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

		response := map[string]interface{}{
			"success": true,
			"data": map[string]interface{}{
				"helpers": []map[string]interface{}{
					{
						"helper_id":   "helper_1",
						"name":        "Tag It",
						"description": "Add tags to contacts",
						"category":    "tagging",
						"status":      "active",
					},
					{
						"helper_id":   "helper_2",
						"name":        "Email Validator",
						"description": "Validate email addresses",
						"category":    "data",
						"status":      "active",
					},
				},
				"total": 2,
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer mockServer.Close()

	service := &MCPService{
		httpClient: &http.Client{},
		baseURL:    mockServer.URL,
		groqAPIKey: "test-key",
	}

	toolCall := types.GroqToolCall{
		Function: types.GroqFunctionCall{
			Name:      "list_helpers",
			Arguments: `{}`,
		},
	}

	result, err := service.ExecuteTool(context.Background(), toolCall, "test-token")

	require.NoError(t, err)
	assert.Contains(t, result, "Tag It")
	assert.Contains(t, result, "Email Validator")
	assert.Contains(t, result, "tagging")
}

// TestGetHelperConfig verifies the get_helper_config tool handler
func TestGetHelperConfig(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/helpers/helper_789", r.URL.Path)
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

		response := map[string]interface{}{
			"success": true,
			"data": map[string]interface{}{
				"helper_id":   "helper_789",
				"name":        "Tag It",
				"description": "Add tags to contacts",
				"config_schema": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"tag_name": map[string]interface{}{
							"type":        "string",
							"description": "Name of tag to apply",
						},
						"contact_id": map[string]interface{}{
							"type":        "string",
							"description": "ID of contact",
						},
					},
					"required": []string{"tag_name", "contact_id"},
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer mockServer.Close()

	service := &MCPService{
		httpClient: &http.Client{},
		baseURL:    mockServer.URL,
		groqAPIKey: "test-key",
	}

	toolCall := types.GroqToolCall{
		Function: types.GroqFunctionCall{
			Name:      "get_helper_config",
			Arguments: `{"helper_id": "helper_789"}`,
		},
	}

	result, err := service.ExecuteTool(context.Background(), toolCall, "test-token")

	require.NoError(t, err)
	assert.Contains(t, result, "helper_789")
	assert.Contains(t, result, "Tag It")
	assert.Contains(t, result, "config_schema")
	assert.Contains(t, result, "tag_name")
}

// TestGetHelperConfig_MissingHelperID verifies error handling
func TestGetHelperConfig_MissingHelperID(t *testing.T) {
	service := &MCPService{
		httpClient: &http.Client{},
		baseURL:    "http://test",
		groqAPIKey: "test-key",
	}

	toolCall := types.GroqToolCall{
		Function: types.GroqFunctionCall{
			Name:      "get_helper_config",
			Arguments: `{}`,
		},
	}

	result, err := service.ExecuteTool(context.Background(), toolCall, "test-token")

	assert.Error(t, err)
	assert.Empty(t, result)
	assert.Contains(t, err.Error(), "helper_id is required")
}

// TestGetConnections verifies the get_connections tool handler
func TestGetConnections(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/platform-connections", r.URL.Path)
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

		response := map[string]interface{}{
			"success": true,
			"data": map[string]interface{}{
				"connections": []map[string]interface{}{
					{
						"connection_id": "conn_1",
						"platform":      "keap",
						"display_name":  "My Keap Account",
						"status":        "active",
						"extra_field":   "should be filtered",
					},
					{
						"connection_id": "conn_2",
						"platform":      "gohighlevel",
						"display_name":  "GHL Agency",
						"status":        "active",
						"extra_field":   "should be filtered",
					},
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer mockServer.Close()

	service := &MCPService{
		httpClient: &http.Client{},
		baseURL:    mockServer.URL,
		groqAPIKey: "test-key",
	}

	toolCall := types.GroqToolCall{
		Function: types.GroqFunctionCall{
			Name:      "get_connections",
			Arguments: `{}`,
		},
	}

	result, err := service.ExecuteTool(context.Background(), toolCall, "test-token")

	require.NoError(t, err)
	assert.Contains(t, result, "conn_1")
	assert.Contains(t, result, "keap")
	assert.Contains(t, result, "My Keap Account")
	assert.Contains(t, result, "GHL Agency")
	// Verify condensed format doesn't include extra fields
	assert.NotContains(t, result, "should be filtered")
}

// TestFormatResponse verifies JSON formatting
func TestFormatResponse(t *testing.T) {
	service := &MCPService{
		httpClient: &http.Client{},
		baseURL:    "http://test",
		groqAPIKey: "test-key",
	}

	t.Run("formats data field when present", func(t *testing.T) {
		resp := map[string]interface{}{
			"success": true,
			"data": map[string]interface{}{
				"id":   "123",
				"name": "Test",
			},
			"metadata": "should be excluded",
		}

		result := service.formatResponse(resp)

		// Should only include data field
		assert.Contains(t, result, `"id"`)
		assert.Contains(t, result, `"name"`)
		assert.Contains(t, result, "Test")
		assert.NotContains(t, result, "metadata")
		assert.NotContains(t, result, "success")
	})

	t.Run("formats full response when data field missing", func(t *testing.T) {
		resp := map[string]interface{}{
			"status":  "ok",
			"message": "Success",
		}

		result := service.formatResponse(resp)

		assert.Contains(t, result, "status")
		assert.Contains(t, result, "message")
		assert.Contains(t, result, "Success")
	})
}

// TestFormatConnectionsResponse verifies connection response condensing
func TestFormatConnectionsResponse(t *testing.T) {
	service := &MCPService{
		httpClient: &http.Client{},
		baseURL:    "http://test",
		groqAPIKey: "test-key",
	}

	t.Run("condenses connection data", func(t *testing.T) {
		resp := map[string]interface{}{
			"success": true,
			"data": map[string]interface{}{
				"connections": []interface{}{
					map[string]interface{}{
						"connection_id":  "conn_1",
						"platform":       "keap",
						"display_name":   "My Account",
						"status":         "active",
						"access_token":   "secret123",
						"refresh_token":  "refresh456",
						"extra_metadata": "not needed",
					},
				},
			},
		}

		result := service.formatConnectionsResponse(resp)

		// Should include essential fields
		assert.Contains(t, result, "conn_1")
		assert.Contains(t, result, "keap")
		assert.Contains(t, result, "My Account")
		assert.Contains(t, result, "active")
		assert.Contains(t, result, `"total": 1`)

		// Should exclude sensitive/extra fields
		assert.NotContains(t, result, "secret123")
		assert.NotContains(t, result, "refresh456")
		assert.NotContains(t, result, "not needed")
	})

	t.Run("falls back to standard format for invalid data", func(t *testing.T) {
		resp := map[string]interface{}{
			"success": true,
			"data":    "invalid",
		}

		result := service.formatConnectionsResponse(resp)

		// formatResponse extracts just the data field when connections parsing fails
		// Result will be just the JSON-encoded data value
		assert.Contains(t, result, "invalid")
	})
}

// TestAPIRequestFailure verifies error handling for failed API requests
func TestAPIRequestFailure(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "Internal server error"}`))
	}))
	defer mockServer.Close()

	service := &MCPService{
		httpClient: &http.Client{},
		baseURL:    mockServer.URL,
		groqAPIKey: "test-key",
	}

	toolCall := types.GroqToolCall{
		Function: types.GroqFunctionCall{
			Name:      "list_helpers",
			Arguments: `{}`,
		},
	}

	result, err := service.ExecuteTool(context.Background(), toolCall, "test-token")

	assert.Error(t, err)
	assert.Empty(t, result)
	assert.Contains(t, err.Error(), "API returned status 500")
}

// TestAPIRequestTimeout verifies timeout handling
func TestAPIRequestTimeout(t *testing.T) {
	// Create a server that never responds
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate timeout by never responding
		<-r.Context().Done()
	}))
	defer mockServer.Close()

	service := &MCPService{
		httpClient: &http.Client{
			Timeout: 1, // 1 nanosecond to force immediate timeout
		},
		baseURL:    mockServer.URL,
		groqAPIKey: "test-key",
	}

	toolCall := types.GroqToolCall{
		Function: types.GroqFunctionCall{
			Name:      "list_helpers",
			Arguments: `{}`,
		},
	}

	result, err := service.ExecuteTool(context.Background(), toolCall, "test-token")

	assert.Error(t, err)
	assert.Empty(t, result)
	// Error should indicate a connection/timeout issue
	assert.True(t, err.Error() != "")
}

// TestNewMCPService verifies service initialization
func TestNewMCPService(t *testing.T) {
	ctx := context.Background()

	t.Run("uses environment variables", func(t *testing.T) {
		t.Setenv("API_BASE_URL", "https://custom.api.example.com")
		t.Setenv("GROQ_API_KEY", "custom-groq-key")

		service := NewMCPService(ctx)

		assert.NotNil(t, service)
		assert.Equal(t, "https://custom.api.example.com", service.baseURL)
		assert.Equal(t, "custom-groq-key", service.groqAPIKey)
		assert.NotNil(t, service.httpClient)
	})

	t.Run("uses default values when env vars not set", func(t *testing.T) {
		// Clear env vars
		t.Setenv("API_BASE_URL", "")
		t.Setenv("GROQ_API_KEY", "")

		service := NewMCPService(ctx)

		assert.NotNil(t, service)
		assert.Equal(t, "https://dev.api.myfusionhelper.ai", service.baseURL)
		assert.Equal(t, "", service.groqAPIKey)
		assert.NotNil(t, service.httpClient)
	})
}
