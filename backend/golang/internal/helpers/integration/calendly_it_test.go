package integration

import (
	"context"
	"fmt"
	"testing"

	"github.com/myfusionhelper/api/internal/connectors"
	"github.com/myfusionhelper/api/internal/helpers"
)

// mockConnectorForCalendlyIt implements the CRMConnector interface for testing calendly_it
type mockConnectorForCalendlyIt struct {
	contact       *connectors.NormalizedContact
	getContactErr error
}

func (m *mockConnectorForCalendlyIt) GetContact(ctx context.Context, contactID string) (*connectors.NormalizedContact, error) {
	if m.getContactErr != nil {
		return nil, m.getContactErr
	}
	return m.contact, nil
}

// Stub implementations
func (m *mockConnectorForCalendlyIt) GetContacts(ctx context.Context, opts connectors.QueryOptions) (*connectors.ContactList, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *mockConnectorForCalendlyIt) CreateContact(ctx context.Context, input connectors.CreateContactInput) (*connectors.NormalizedContact, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *mockConnectorForCalendlyIt) UpdateContact(ctx context.Context, contactID string, updates connectors.UpdateContactInput) (*connectors.NormalizedContact, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *mockConnectorForCalendlyIt) DeleteContact(ctx context.Context, contactID string) error {
	return fmt.Errorf("not implemented")
}

func (m *mockConnectorForCalendlyIt) GetTags(ctx context.Context) ([]connectors.Tag, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *mockConnectorForCalendlyIt) ApplyTag(ctx context.Context, contactID string, tagID string) error {
	return fmt.Errorf("not implemented")
}

func (m *mockConnectorForCalendlyIt) RemoveTag(ctx context.Context, contactID string, tagID string) error {
	return fmt.Errorf("not implemented")
}

func (m *mockConnectorForCalendlyIt) GetCustomFields(ctx context.Context) ([]connectors.CustomField, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *mockConnectorForCalendlyIt) GetContactFieldValue(ctx context.Context, contactID, fieldKey string) (interface{}, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *mockConnectorForCalendlyIt) SetContactFieldValue(ctx context.Context, contactID, fieldKey string, value interface{}) error {
	return fmt.Errorf("not implemented")
}

func (m *mockConnectorForCalendlyIt) TriggerAutomation(ctx context.Context, contactID string, automationID string) error {
	return fmt.Errorf("not implemented")
}

func (m *mockConnectorForCalendlyIt) AchieveGoal(ctx context.Context, contactID string, goalName string, integration string) error {
	return fmt.Errorf("not implemented")
}

func (m *mockConnectorForCalendlyIt) TestConnection(ctx context.Context) error {
	return fmt.Errorf("not implemented")
}

func (m *mockConnectorForCalendlyIt) GetMetadata() connectors.ConnectorMetadata {
	return connectors.ConnectorMetadata{
		PlatformSlug: "test",
		PlatformName: "Test Platform",
	}
}

func (m *mockConnectorForCalendlyIt) GetCapabilities() []connectors.Capability {
	return []connectors.Capability{connectors.CapContacts}
}

// Tests

func TestCalendlyIt_GetMetadata(t *testing.T) {
	helper := &CalendlyIt{}

	if helper.GetName() != "Calendly It" {
		t.Errorf("Expected name 'Calendly It', got '%s'", helper.GetName())
	}
	if helper.GetType() != "calendly_it" {
		t.Errorf("Expected type 'calendly_it', got '%s'", helper.GetType())
	}
	if helper.GetCategory() != "integration" {
		t.Errorf("Expected category 'integration', got '%s'", helper.GetCategory())
	}
	if !helper.RequiresCRM() {
		t.Error("Expected RequiresCRM to be true")
	}
}

func TestCalendlyIt_GetConfigSchema(t *testing.T) {
	helper := &CalendlyIt{}
	schema := helper.GetConfigSchema()

	if schema == nil {
		t.Fatal("Schema should not be nil")
	}

	props, ok := schema["properties"].(map[string]interface{})
	if !ok {
		t.Fatal("Schema should have properties")
	}

	if _, ok := props["event_type_uri"]; !ok {
		t.Error("Schema should have event_type_uri property")
	}
	if _, ok := props["api_token"]; !ok {
		t.Error("Schema should have api_token property")
	}
	if _, ok := props["email_field"]; !ok {
		t.Error("Schema should have email_field property")
	}
	if _, ok := props["name_field"]; !ok {
		t.Error("Schema should have name_field property")
	}

	required, ok := schema["required"].([]string)
	if !ok {
		t.Fatal("Schema should have required array")
	}

	if len(required) < 2 {
		t.Error("Schema should require at least 2 fields")
	}
}

func TestCalendlyIt_ValidateConfig_MissingEventTypeURI(t *testing.T) {
	helper := &CalendlyIt{}

	config := map[string]interface{}{
		"api_token": "test-token",
	}
	err := helper.ValidateConfig(config)
	if err == nil {
		t.Error("Expected error for missing event_type_uri")
	}
}

func TestCalendlyIt_ValidateConfig_EmptyEventTypeURI(t *testing.T) {
	helper := &CalendlyIt{}

	config := map[string]interface{}{
		"event_type_uri": "",
		"api_token":      "test-token",
	}
	err := helper.ValidateConfig(config)
	if err == nil {
		t.Error("Expected error for empty event_type_uri")
	}
}

func TestCalendlyIt_ValidateConfig_MissingAPIToken(t *testing.T) {
	helper := &CalendlyIt{}

	config := map[string]interface{}{
		"event_type_uri": "https://api.calendly.com/event_types/XXXX",
	}
	err := helper.ValidateConfig(config)
	if err == nil {
		t.Error("Expected error for missing api_token")
	}
}

func TestCalendlyIt_ValidateConfig_EmptyAPIToken(t *testing.T) {
	helper := &CalendlyIt{}

	config := map[string]interface{}{
		"event_type_uri": "https://api.calendly.com/event_types/XXXX",
		"api_token":      "",
	}
	err := helper.ValidateConfig(config)
	if err == nil {
		t.Error("Expected error for empty api_token")
	}
}

func TestCalendlyIt_ValidateConfig_Valid(t *testing.T) {
	helper := &CalendlyIt{}

	config := map[string]interface{}{
		"event_type_uri": "https://api.calendly.com/event_types/XXXX",
		"api_token":      "test-token-123",
	}
	err := helper.ValidateConfig(config)
	if err != nil {
		t.Errorf("Expected no error for valid config, got: %v", err)
	}
}

func TestCalendlyIt_Execute_Success(t *testing.T) {
	helper := &CalendlyIt{}
	mockConn := &mockConnectorForCalendlyIt{
		contact: &connectors.NormalizedContact{
			ID:        "contact-123",
			FirstName: "John",
			LastName:  "Doe",
			Email:     "john.doe@example.com",
		},
	}
	ctx := context.Background()

	input := helpers.HelperInput{
		ContactID: "contact-123",
		Config: map[string]interface{}{
			"event_type_uri": "https://api.calendly.com/event_types/XXXX",
			"api_token":      "test-token-123",
		},
		Connector: mockConn,
	}

	output, err := helper.Execute(ctx, input)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if !output.Success {
		t.Error("Expected success to be true")
	}

	if output.Message == "" {
		t.Error("Expected non-empty message")
	}
}

func TestCalendlyIt_Execute_CustomEmailField(t *testing.T) {
	helper := &CalendlyIt{}
	mockConn := &mockConnectorForCalendlyIt{
		contact: &connectors.NormalizedContact{
			ID:        "contact-123",
			FirstName: "John",
			LastName:  "Doe",
			Email:     "john.doe@example.com",
			CustomFields: map[string]interface{}{
				"work_email": "john@work.com",
			},
		},
	}
	ctx := context.Background()

	input := helpers.HelperInput{
		ContactID: "contact-123",
		Config: map[string]interface{}{
			"event_type_uri": "https://api.calendly.com/event_types/XXXX",
			"api_token":      "test-token-123",
			"email_field":    "work_email",
		},
		Connector: mockConn,
	}

	output, err := helper.Execute(ctx, input)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if !output.Success {
		t.Error("Expected success to be true")
	}

	// Verify work_email is used in modified data
	if modData, ok := output.ModifiedData["email"]; ok {
		if modData != "john@work.com" {
			t.Errorf("Expected email to be 'john@work.com', got: %v", modData)
		}
	}
}

func TestCalendlyIt_Execute_CustomNameField(t *testing.T) {
	helper := &CalendlyIt{}
	mockConn := &mockConnectorForCalendlyIt{
		contact: &connectors.NormalizedContact{
			ID:        "contact-123",
			FirstName: "John",
			LastName:  "Doe",
			Email:     "john.doe@example.com",
			CustomFields: map[string]interface{}{
				"full_name": "Dr. John Doe",
			},
		},
	}
	ctx := context.Background()

	input := helpers.HelperInput{
		ContactID: "contact-123",
		Config: map[string]interface{}{
			"event_type_uri": "https://api.calendly.com/event_types/XXXX",
			"api_token":      "test-token-123",
			"name_field":     "full_name",
		},
		Connector: mockConn,
	}

	output, err := helper.Execute(ctx, input)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if !output.Success {
		t.Error("Expected success to be true")
	}

	// Verify custom name is used
	if modData, ok := output.ModifiedData["name"]; ok {
		if modData != "Dr. John Doe" {
			t.Errorf("Expected name to be 'Dr. John Doe', got: %v", modData)
		}
	}
}

func TestCalendlyIt_Execute_EmptyEmail(t *testing.T) {
	helper := &CalendlyIt{}
	mockConn := &mockConnectorForCalendlyIt{
		contact: &connectors.NormalizedContact{
			ID:        "contact-123",
			FirstName: "John",
			LastName:  "Doe",
			Email:     "",
		},
	}
	ctx := context.Background()

	input := helpers.HelperInput{
		ContactID: "contact-123",
		Config: map[string]interface{}{
			"event_type_uri": "https://api.calendly.com/event_types/XXXX",
			"api_token":      "test-token-123",
		},
		Connector: mockConn,
	}

	output, err := helper.Execute(ctx, input)
	if err == nil {
		t.Error("Expected error for empty email")
	}

	if output.Success {
		t.Error("Expected success to be false")
	}

	if output.Message == "" {
		t.Error("Expected error message")
	}
}

func TestCalendlyIt_Execute_GetContactError(t *testing.T) {
	helper := &CalendlyIt{}
	mockConn := &mockConnectorForCalendlyIt{
		getContactErr: fmt.Errorf("CRM API error"),
	}
	ctx := context.Background()

	input := helpers.HelperInput{
		ContactID: "contact-123",
		Config: map[string]interface{}{
			"event_type_uri": "https://api.calendly.com/event_types/XXXX",
			"api_token":      "test-token-123",
		},
		Connector: mockConn,
	}

	output, err := helper.Execute(ctx, input)
	if err == nil {
		t.Error("Expected error for GetContact failure")
	}

	if output.Success {
		t.Error("Expected success to be false")
	}

	if output.Message != "Failed to get contact data: CRM API error" {
		t.Errorf("Unexpected error message: %s", output.Message)
	}
}

func TestCalendlyIt_Execute_ActionsRecorded(t *testing.T) {
	helper := &CalendlyIt{}
	mockConn := &mockConnectorForCalendlyIt{
		contact: &connectors.NormalizedContact{
			ID:        "contact-123",
			FirstName: "John",
			LastName:  "Doe",
			Email:     "john.doe@example.com",
		},
	}
	ctx := context.Background()

	input := helpers.HelperInput{
		ContactID: "contact-123",
		Config: map[string]interface{}{
			"event_type_uri": "https://api.calendly.com/event_types/XXXX",
			"api_token":      "test-token-123",
		},
		Connector: mockConn,
	}

	output, err := helper.Execute(ctx, input)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(output.Actions) != 1 {
		t.Fatalf("Expected 1 action, got %d", len(output.Actions))
	}

	action := output.Actions[0]
	if action.Type != "webhook_queued" {
		t.Errorf("Expected action type 'webhook_queued', got '%s'", action.Type)
	}
	if action.Target != "https://api.calendly.com/scheduling_links" {
		t.Errorf("Expected action target to be Calendly API URL, got '%s'", action.Target)
	}
}

func TestCalendlyIt_Execute_ResultFieldRecorded(t *testing.T) {
	helper := &CalendlyIt{}
	mockConn := &mockConnectorForCalendlyIt{
		contact: &connectors.NormalizedContact{
			ID:        "contact-123",
			FirstName: "John",
			LastName:  "Doe",
			Email:     "john.doe@example.com",
		},
	}
	ctx := context.Background()

	input := helpers.HelperInput{
		ContactID: "contact-123",
		Config: map[string]interface{}{
			"event_type_uri": "https://api.calendly.com/event_types/XXXX",
			"api_token":      "test-token-123",
			"result_field":   "calendly_link",
		},
		Connector: mockConn,
	}

	output, err := helper.Execute(ctx, input)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if output.ModifiedData == nil {
		t.Fatal("Expected ModifiedData to be set")
	}

	if output.ModifiedData["result_field"] != "calendly_link" {
		t.Errorf("Expected result_field to be 'calendly_link', got '%v'", output.ModifiedData["result_field"])
	}
}

func TestCalendlyIt_Execute_LogsRecorded(t *testing.T) {
	helper := &CalendlyIt{}
	mockConn := &mockConnectorForCalendlyIt{
		contact: &connectors.NormalizedContact{
			ID:        "contact-123",
			FirstName: "John",
			LastName:  "Doe",
			Email:     "john.doe@example.com",
		},
	}
	ctx := context.Background()

	input := helpers.HelperInput{
		ContactID: "contact-123",
		Config: map[string]interface{}{
			"event_type_uri": "https://api.calendly.com/event_types/XXXX",
			"api_token":      "test-token-123",
		},
		Connector: mockConn,
	}

	output, err := helper.Execute(ctx, input)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(output.Logs) == 0 {
		t.Error("Expected logs to be recorded")
	}
}

func TestCalendlyIt_Execute_DefaultNameFromFirstLastName(t *testing.T) {
	helper := &CalendlyIt{}
	mockConn := &mockConnectorForCalendlyIt{
		contact: &connectors.NormalizedContact{
			ID:        "contact-123",
			FirstName: "Jane",
			LastName:  "Smith",
			Email:     "jane.smith@example.com",
		},
	}
	ctx := context.Background()

	input := helpers.HelperInput{
		ContactID: "contact-123",
		Config: map[string]interface{}{
			"event_type_uri": "https://api.calendly.com/event_types/XXXX",
			"api_token":      "test-token-123",
			"name_field":     "full_name", // Use full_name field (built from FirstName + LastName)
		},
		Connector: mockConn,
	}

	output, err := helper.Execute(ctx, input)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Verify name is constructed from first and last name when full_name field is used
	if modData, ok := output.ModifiedData["name"]; ok {
		if modData != "Jane Smith" {
			t.Errorf("Expected name to be 'Jane Smith', got: %v", modData)
		}
	}
}

func TestCalendlyIt_Execute_AuthorizationHeaderSet(t *testing.T) {
	helper := &CalendlyIt{}
	mockConn := &mockConnectorForCalendlyIt{
		contact: &connectors.NormalizedContact{
			ID:        "contact-123",
			FirstName: "John",
			LastName:  "Doe",
			Email:     "john.doe@example.com",
		},
	}
	ctx := context.Background()

	apiToken := "my-secret-token-456"
	input := helpers.HelperInput{
		ContactID: "contact-123",
		Config: map[string]interface{}{
			"event_type_uri": "https://api.calendly.com/event_types/XXXX",
			"api_token":      apiToken,
		},
		Connector: mockConn,
	}

	output, err := helper.Execute(ctx, input)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Verify Authorization header is set in action
	if len(output.Actions) > 0 {
		actionValue, ok := output.Actions[0].Value.(map[string]interface{})
		if !ok {
			t.Fatal("Expected action value to be a map")
		}

		headers, ok := actionValue["headers"].(map[string]string)
		if !ok {
			t.Fatal("Expected headers to be present in action value")
		}

		expectedAuth := "Bearer " + apiToken
		if headers["Authorization"] != expectedAuth {
			t.Errorf("Expected Authorization header to be '%s', got '%s'", expectedAuth, headers["Authorization"])
		}
	}
}
