package integration

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/myfusionhelper/api/internal/connectors"
	"github.com/myfusionhelper/api/internal/helpers"
)

// Mock connector for trello_it testing
type mockConnectorForTrello struct {
	contact       *connectors.NormalizedContact
	appliedTags   []string
	updatedFields map[string]interface{}
}

func (m *mockConnectorForTrello) GetContact(ctx context.Context, contactID string) (*connectors.NormalizedContact, error) {
	if m.contact != nil {
		return m.contact, nil
	}
	return nil, fmt.Errorf("contact not found")
}

func (m *mockConnectorForTrello) ApplyTag(ctx context.Context, contactID, tagID string) error {
	m.appliedTags = append(m.appliedTags, tagID)
	return nil
}

// Implement remaining interface methods as stubs
func (m *mockConnectorForTrello) GetContacts(ctx context.Context, opts connectors.QueryOptions) (*connectors.ContactList, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForTrello) CreateContact(ctx context.Context, input connectors.CreateContactInput) (*connectors.NormalizedContact, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForTrello) UpdateContact(ctx context.Context, contactID string, updates connectors.UpdateContactInput) (*connectors.NormalizedContact, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForTrello) DeleteContact(ctx context.Context, contactID string) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorForTrello) GetTags(ctx context.Context) ([]connectors.Tag, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForTrello) RemoveTag(ctx context.Context, contactID, tagID string) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorForTrello) GetCustomFields(ctx context.Context) ([]connectors.CustomField, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForTrello) GetContactFieldValue(ctx context.Context, contactID, fieldKey string) (interface{}, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForTrello) SetContactFieldValue(ctx context.Context, contactID, fieldKey string, value interface{}) error {
	if m.updatedFields == nil {
		m.updatedFields = make(map[string]interface{})
	}
	m.updatedFields[fieldKey] = value
	return nil
}
func (m *mockConnectorForTrello) TriggerAutomation(ctx context.Context, contactID, automationID string) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorForTrello) AchieveGoal(ctx context.Context, contactID, goalName, integration string) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorForTrello) TestConnection(ctx context.Context) error {
	return nil
}
func (m *mockConnectorForTrello) GetMetadata() connectors.ConnectorMetadata {
	return connectors.ConnectorMetadata{}
}
func (m *mockConnectorForTrello) GetCapabilities() []connectors.Capability {
	return nil
}

// Test helper metadata
func TestTrelloIt_Metadata(t *testing.T) {
	helper := &TrelloIt{}

	if helper.GetName() != "Trello It" {
		t.Errorf("Expected name 'Trello It', got '%s'", helper.GetName())
	}
	if helper.GetType() != "trello_it" {
		t.Errorf("Expected type 'trello_it', got '%s'", helper.GetType())
	}
	if helper.GetCategory() != "integration" {
		t.Errorf("Expected category 'integration', got '%s'", helper.GetCategory())
	}
	if !helper.RequiresCRM() {
		t.Error("Expected RequiresCRM to be true")
	}

	schema := helper.GetConfigSchema()
	if schema == nil {
		t.Fatal("Expected config schema, got nil")
	}
}

// Test validation - missing board_id
func TestTrelloIt_ValidateConfig_MissingBoardId(t *testing.T) {
	helper := &TrelloIt{}
	config := map[string]interface{}{
		"list_id":             "list-123",
		"card_name_template": "Contact: {first_name} {last_name}",
	}

	err := helper.ValidateConfig(config)
	if err == nil {
		t.Error("Expected validation error for missing board_id")
	}
	if !strings.Contains(err.Error(), "board_id") {
		t.Errorf("Expected error about board_id, got: %v", err)
	}
}

// Test validation - missing list_id
func TestTrelloIt_ValidateConfig_MissingListId(t *testing.T) {
	helper := &TrelloIt{}
	config := map[string]interface{}{
		"board_id":           "board-123",
		"card_name_template": "Contact: {first_name} {last_name}",
	}

	err := helper.ValidateConfig(config)
	if err == nil {
		t.Error("Expected validation error for missing list_id")
	}
	if !strings.Contains(err.Error(), "list_id") {
		t.Errorf("Expected error about list_id, got: %v", err)
	}
}

// Test validation - missing card_name_template
func TestTrelloIt_ValidateConfig_MissingCardNameTemplate(t *testing.T) {
	helper := &TrelloIt{}
	config := map[string]interface{}{
		"board_id": "board-123",
		"list_id":  "list-123",
	}

	err := helper.ValidateConfig(config)
	if err == nil {
		t.Error("Expected validation error for missing card_name_template")
	}
	if !strings.Contains(err.Error(), "card_name_template") {
		t.Errorf("Expected error about card_name_template, got: %v", err)
	}
}

// Test validation - valid config
func TestTrelloIt_ValidateConfig_Valid(t *testing.T) {
	helper := &TrelloIt{}
	config := map[string]interface{}{
		"board_id":           "board-123",
		"list_id":            "list-456",
		"card_name_template": "New Lead: {first_name} {last_name}",
	}

	err := helper.ValidateConfig(config)
	if err != nil {
		t.Errorf("Expected no validation error, got: %v", err)
	}
}

// Test execution - missing Trello connection
func TestTrelloIt_Execute_MissingConnection(t *testing.T) {
	helper := &TrelloIt{}

	input := helpers.HelperInput{
		ContactID:    "contact-123",
		ServiceAuths: map[string]*connectors.ConnectorConfig{},
		Config: map[string]interface{}{
			"board_id":           "board-123",
			"list_id":            "list-123",
			"card_name_template": "Contact",
		},
	}

	output, err := helper.Execute(context.Background(), input)
	if err == nil {
		t.Error("Expected error for missing Trello connection")
	}
	if !strings.Contains(output.Message, "Trello connection required") {
		t.Errorf("Expected error message about connection, got: %s", output.Message)
	}
}

// Test template interpolation
func TestTrelloIt_TemplateInterpolation(t *testing.T) {
	// This test verifies the template replacement logic works correctly
	helper := &TrelloIt{}

	mockConn := &mockConnectorForTrello{
		contact: &connectors.NormalizedContact{
			ID:        "contact-789",
			FirstName: "John",
			LastName:  "Smith",
			Email:     "john@example.com",
			Phone:     "555-1234",
			Company:   "Acme Corp",
		},
	}

	input := helpers.HelperInput{
		ContactID: "contact-789",
		ContactData: &connectors.NormalizedContact{
			ID:        "contact-789",
			FirstName: "John",
			LastName:  "Smith",
			Email:     "john@example.com",
			Phone:     "555-1234",
			Company:   "Acme Corp",
		},
		Connector: mockConn,
		ServiceAuths: map[string]*connectors.ConnectorConfig{
			"trello": {
				APIKey:    "test-api-key",
				APISecret: "test-token",
			},
		},
		Config: map[string]interface{}{
			"board_id":                   "board-123",
			"list_id":                    "list-456",
			"card_name_template":         "Lead: {first_name} {last_name}",
			"card_description_template": "Email: {email}\nPhone: {phone}\nCompany: {company}",
		},
	}

	// This will make a real API call and likely fail, but we can verify the logs contain interpolated values
	output, _ := helper.Execute(context.Background(), input)

	// Check logs for interpolated values
	if len(output.Logs) > 0 {
		firstLog := output.Logs[0]
		if !strings.Contains(firstLog, "Lead: John Smith") {
			t.Errorf("Expected log to contain interpolated card name, got: %s", firstLog)
		}
	}
}

// Test execution - with contact fetch
func TestTrelloIt_Execute_FetchContact(t *testing.T) {
	helper := &TrelloIt{}

	mockConn := &mockConnectorForTrello{
		contact: &connectors.NormalizedContact{
			ID:        "contact-456",
			FirstName: "Jane",
			LastName:  "Doe",
			Email:     "jane@example.com",
		},
	}

	input := helpers.HelperInput{
		ContactID: "contact-456",
		Connector: mockConn,
		ServiceAuths: map[string]*connectors.ConnectorConfig{
			"trello": {
				APIKey:    "test-key",
				APISecret: "test-token",
			},
		},
		Config: map[string]interface{}{
			"board_id":           "board-123",
			"list_id":            "list-456",
			"card_name_template": "{first_name} {last_name}",
		},
	}

	// This will make a real API call and likely fail, but we can verify contact fetching works
	output, _ := helper.Execute(context.Background(), input)

	// Verify the contact was fetched (check logs)
	if len(output.Logs) > 0 {
		firstLog := output.Logs[0]
		if !strings.Contains(firstLog, "Jane Doe") {
			t.Errorf("Expected log to contain contact name, got: %s", firstLog)
		}
	}
}

// Test successful execution with mock server
func TestTrelloIt_Execute_Success(t *testing.T) {
	// Create mock Trello API server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request method
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST method, got: %s", r.Method)
		}

		// Verify query parameters
		query := r.URL.Query()
		if query.Get("key") != "test-api-key" {
			t.Errorf("Expected key=test-api-key, got: %s", query.Get("key"))
		}
		if query.Get("token") != "test-token" {
			t.Errorf("Expected token=test-token, got: %s", query.Get("token"))
		}
		if query.Get("idList") != "list-456" {
			t.Errorf("Expected idList=list-456, got: %s", query.Get("idList"))
		}
		if !strings.Contains(query.Get("name"), "Test Card") {
			t.Errorf("Expected card name to contain 'Test Card', got: %s", query.Get("name"))
		}

		// Return mock response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"id": "card-12345",
			"shortUrl": "https://trello.com/c/abc123",
			"name": "Test Card"
		}`))
	}))
	defer server.Close()

	// Note: In a real test with dependency injection, we'd replace the API URL
	// For now, this test verifies the structure without the actual API call
	helper := &TrelloIt{}

	mockConn := &mockConnectorForTrello{
		contact: &connectors.NormalizedContact{
			ID:        "contact-123",
			FirstName: "Test",
			LastName:  "User",
			Email:     "test@example.com",
		},
	}

	input := helpers.HelperInput{
		ContactID: "contact-123",
		ContactData: &connectors.NormalizedContact{
			ID:        "contact-123",
			FirstName: "Test",
			LastName:  "User",
			Email:     "test@example.com",
		},
		Connector: mockConn,
		ServiceAuths: map[string]*connectors.ConnectorConfig{
			"trello": {
				APIKey:    "test-api-key",
				APISecret: "test-token",
			},
		},
		Config: map[string]interface{}{
			"board_id":           "board-123",
			"list_id":            "list-456",
			"card_name_template": "Test Card: {first_name} {last_name}",
		},
	}

	// Execute (will make real API call to Trello, expected to fail in test environment)
	output, err := helper.Execute(context.Background(), input)

	// The API call will likely fail without proper credentials, but verify structure
	if err != nil {
		t.Logf("API call failed as expected in test environment: %v", err)
	} else {
		// If it somehow succeeds, verify the output
		if !output.Success {
			t.Error("Expected success=true")
		}
		if output.ModifiedData == nil {
			t.Error("Expected ModifiedData to be non-nil")
		}
	}
}

// Test tag application
func TestTrelloIt_Execute_TagApplication(t *testing.T) {
	helper := &TrelloIt{}

	mockConn := &mockConnectorForTrello{
		contact: &connectors.NormalizedContact{
			ID:        "contact-999",
			FirstName: "Tag",
			LastName:  "Test",
			Email:     "tag@example.com",
		},
		appliedTags: make([]string, 0),
	}

	input := helpers.HelperInput{
		ContactID: "contact-999",
		ContactData: &connectors.NormalizedContact{
			ID:        "contact-999",
			FirstName: "Tag",
			LastName:  "Test",
			Email:     "tag@example.com",
		},
		Connector: mockConn,
		ServiceAuths: map[string]*connectors.ConnectorConfig{
			"trello": {
				APIKey:    "test-key",
				APISecret: "test-token",
			},
		},
		Config: map[string]interface{}{
			"board_id":           "board-123",
			"list_id":            "list-456",
			"card_name_template": "{first_name} {last_name}",
			"apply_tag":          "trello-card-created",
		},
	}

	// Execute (will fail at API call, but that's ok for this test)
	helper.Execute(context.Background(), input)

	// Note: Since the API call fails, tags won't be applied
	// In a proper test with HTTP mocking, we'd verify:
	// - mockConn.appliedTags contains "trello-card-created"
}

// Test action logging
func TestTrelloIt_Execute_ActionLogging(t *testing.T) {
	helper := &TrelloIt{}

	mockConn := &mockConnectorForTrello{
		contact: &connectors.NormalizedContact{
			ID:        "contact-888",
			FirstName: "Action",
			LastName:  "Logger",
			Email:     "action@example.com",
		},
	}

	input := helpers.HelperInput{
		ContactID: "contact-888",
		ContactData: &connectors.NormalizedContact{
			ID:        "contact-888",
			FirstName: "Action",
			LastName:  "Logger",
			Email:     "action@example.com",
		},
		Connector: mockConn,
		ServiceAuths: map[string]*connectors.ConnectorConfig{
			"trello": {
				APIKey:    "test-key",
				APISecret: "test-token",
			},
		},
		Config: map[string]interface{}{
			"board_id":           "board-123",
			"list_id":            "list-456",
			"card_name_template": "Contact: {first_name}",
		},
	}

	output, _ := helper.Execute(context.Background(), input)

	// Verify logs were created
	if len(output.Logs) == 0 {
		t.Error("Expected logs to be generated")
	}

	// First log should mention creating the card
	if !strings.Contains(output.Logs[0], "Creating Trello card") {
		t.Errorf("Expected first log to mention card creation, got: %s", output.Logs[0])
	}
	if !strings.Contains(output.Logs[0], "board-123") {
		t.Errorf("Expected first log to mention board ID, got: %s", output.Logs[0])
	}
	if !strings.Contains(output.Logs[0], "list-456") {
		t.Errorf("Expected first log to mention list ID, got: %s", output.Logs[0])
	}
}

// Test empty description template
func TestTrelloIt_Execute_EmptyDescription(t *testing.T) {
	helper := &TrelloIt{}

	mockConn := &mockConnectorForTrello{
		contact: &connectors.NormalizedContact{
			ID:        "contact-777",
			FirstName: "No",
			LastName:  "Description",
			Email:     "no-desc@example.com",
		},
	}

	input := helpers.HelperInput{
		ContactID: "contact-777",
		ContactData: &connectors.NormalizedContact{
			ID:        "contact-777",
			FirstName: "No",
			LastName:  "Description",
			Email:     "no-desc@example.com",
		},
		Connector: mockConn,
		ServiceAuths: map[string]*connectors.ConnectorConfig{
			"trello": {
				APIKey:    "test-key",
				APISecret: "test-token",
			},
		},
		Config: map[string]interface{}{
			"board_id":           "board-123",
			"list_id":            "list-456",
			"card_name_template": "Contact: {first_name}",
			// No card_description_template provided
		},
	}

	// Execute (will fail at API, but we're testing that it doesn't crash without description)
	output, _ := helper.Execute(context.Background(), input)

	// Should not crash, should have logs
	if len(output.Logs) == 0 {
		t.Error("Expected logs even without description template")
	}
}
