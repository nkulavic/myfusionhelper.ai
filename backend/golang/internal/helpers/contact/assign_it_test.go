package contact

import (
	"context"
	"fmt"
	"testing"

	"github.com/myfusionhelper/api/internal/connectors"
	"github.com/myfusionhelper/api/internal/helpers"
)

// mockConnectorForAssignIt implements the CRMConnector interface for testing assign_it
type mockConnectorForAssignIt struct {
	fieldsSet     map[string]interface{}
	setFieldError error
}

func (m *mockConnectorForAssignIt) SetContactFieldValue(ctx context.Context, contactID, fieldKey string, value interface{}) error {
	if m.setFieldError != nil {
		return m.setFieldError
	}
	if m.fieldsSet == nil {
		m.fieldsSet = make(map[string]interface{})
	}
	m.fieldsSet[fieldKey] = value
	return nil
}

// Stub implementations with correct signatures
func (m *mockConnectorForAssignIt) GetContacts(ctx context.Context, opts connectors.QueryOptions) (*connectors.ContactList, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *mockConnectorForAssignIt) GetContact(ctx context.Context, contactID string) (*connectors.NormalizedContact, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *mockConnectorForAssignIt) CreateContact(ctx context.Context, input connectors.CreateContactInput) (*connectors.NormalizedContact, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *mockConnectorForAssignIt) UpdateContact(ctx context.Context, contactID string, updates connectors.UpdateContactInput) (*connectors.NormalizedContact, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *mockConnectorForAssignIt) DeleteContact(ctx context.Context, contactID string) error {
	return fmt.Errorf("not implemented")
}

func (m *mockConnectorForAssignIt) GetTags(ctx context.Context) ([]connectors.Tag, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *mockConnectorForAssignIt) ApplyTag(ctx context.Context, contactID string, tagID string) error {
	return fmt.Errorf("not implemented")
}

func (m *mockConnectorForAssignIt) RemoveTag(ctx context.Context, contactID string, tagID string) error {
	return fmt.Errorf("not implemented")
}

func (m *mockConnectorForAssignIt) GetCustomFields(ctx context.Context) ([]connectors.CustomField, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *mockConnectorForAssignIt) GetContactFieldValue(ctx context.Context, contactID, fieldKey string) (interface{}, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *mockConnectorForAssignIt) TriggerAutomation(ctx context.Context, contactID string, automationID string) error {
	return fmt.Errorf("not implemented")
}

func (m *mockConnectorForAssignIt) AchieveGoal(ctx context.Context, contactID string, goalName string, integration string) error {
	return fmt.Errorf("not implemented")
}

func (m *mockConnectorForAssignIt) TestConnection(ctx context.Context) error {
	return fmt.Errorf("not implemented")
}

func (m *mockConnectorForAssignIt) GetMetadata() connectors.ConnectorMetadata {
	return connectors.ConnectorMetadata{
		PlatformSlug: "test",
		PlatformName: "Test Platform",
	}
}

func (m *mockConnectorForAssignIt) GetCapabilities() []connectors.Capability {
	return []connectors.Capability{connectors.CapContacts, connectors.CapCustomFields}
}

// Tests

func TestAssignIt_GetMetadata(t *testing.T) {
	helper := &AssignIt{}

	if helper.GetName() != "Assign It" {
		t.Errorf("Expected name 'Assign It', got '%s'", helper.GetName())
	}
	if helper.GetType() != "assign_it" {
		t.Errorf("Expected type 'assign_it', got '%s'", helper.GetType())
	}
	if helper.GetCategory() != "contact" {
		t.Errorf("Expected category 'contact', got '%s'", helper.GetCategory())
	}
	if !helper.RequiresCRM() {
		t.Error("Expected RequiresCRM to be true")
	}
}

func TestAssignIt_GetConfigSchema(t *testing.T) {
	helper := &AssignIt{}
	schema := helper.GetConfigSchema()

	if schema == nil {
		t.Fatal("Schema should not be nil")
	}

	props, ok := schema["properties"].(map[string]interface{})
	if !ok {
		t.Fatal("Schema should have properties")
	}

	if _, ok := props["owner_id"]; !ok {
		t.Error("Schema should have owner_id property")
	}

	if _, ok := props["owner_field"]; !ok {
		t.Error("Schema should have owner_field property")
	}

	required, ok := schema["required"].([]string)
	if !ok {
		t.Fatal("Schema should have required array")
	}

	if len(required) == 0 || required[0] != "owner_id" {
		t.Error("Schema should require owner_id")
	}
}

func TestAssignIt_ValidateConfig_MissingOwnerID(t *testing.T) {
	helper := &AssignIt{}

	config := map[string]interface{}{}
	err := helper.ValidateConfig(config)
	if err == nil {
		t.Error("Expected error for missing owner_id")
	}
}

func TestAssignIt_ValidateConfig_EmptyOwnerID(t *testing.T) {
	helper := &AssignIt{}

	config := map[string]interface{}{
		"owner_id": "",
	}
	err := helper.ValidateConfig(config)
	if err == nil {
		t.Error("Expected error for empty owner_id")
	}
}

func TestAssignIt_ValidateConfig_InvalidOwnerIDType(t *testing.T) {
	helper := &AssignIt{}

	config := map[string]interface{}{
		"owner_id": 123, // number instead of string
	}
	err := helper.ValidateConfig(config)
	if err == nil {
		t.Error("Expected error for non-string owner_id")
	}
}

func TestAssignIt_ValidateConfig_Valid(t *testing.T) {
	helper := &AssignIt{}

	config := map[string]interface{}{
		"owner_id": "user-123",
	}
	err := helper.ValidateConfig(config)
	if err != nil {
		t.Errorf("Expected no error for valid config, got: %v", err)
	}
}

func TestAssignIt_ValidateConfig_ValidWithOwnerField(t *testing.T) {
	helper := &AssignIt{}

	config := map[string]interface{}{
		"owner_id":     "user-123",
		"owner_field": "custom_owner",
	}
	err := helper.ValidateConfig(config)
	if err != nil {
		t.Errorf("Expected no error for valid config with owner_field, got: %v", err)
	}
}

func TestAssignIt_Execute_Success(t *testing.T) {
	helper := &AssignIt{}
	mockConn := &mockConnectorForAssignIt{}
	ctx := context.Background()

	input := helpers.HelperInput{
		ContactID: "contact-123",
		Config: map[string]interface{}{
			"owner_id": "user-456",
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

	if mockConn.fieldsSet["owner_id"] != "user-456" {
		t.Errorf("Expected owner_id to be set to 'user-456', got: %v", mockConn.fieldsSet["owner_id"])
	}

	if output.Message != "Assigned contact to owner user-456" {
		t.Errorf("Unexpected message: %s", output.Message)
	}
}

func TestAssignIt_Execute_SuccessWithCustomOwnerField(t *testing.T) {
	helper := &AssignIt{}
	mockConn := &mockConnectorForAssignIt{}
	ctx := context.Background()

	input := helpers.HelperInput{
		ContactID: "contact-123",
		Config: map[string]interface{}{
			"owner_id":     "user-456",
			"owner_field": "assigned_to",
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

	if mockConn.fieldsSet["assigned_to"] != "user-456" {
		t.Errorf("Expected assigned_to to be set to 'user-456', got: %v", mockConn.fieldsSet["assigned_to"])
	}
}

func TestAssignIt_Execute_SetFieldError(t *testing.T) {
	helper := &AssignIt{}
	mockConn := &mockConnectorForAssignIt{
		setFieldError: fmt.Errorf("CRM API error"),
	}
	ctx := context.Background()

	input := helpers.HelperInput{
		ContactID: "contact-123",
		Config: map[string]interface{}{
			"owner_id": "user-456",
		},
		Connector: mockConn,
	}

	output, err := helper.Execute(ctx, input)
	if err == nil {
		t.Error("Expected error, got nil")
	}

	if output.Success {
		t.Error("Expected success to be false")
	}

	if output.Message != "Failed to assign owner: CRM API error" {
		t.Errorf("Unexpected error message: %s", output.Message)
	}
}

func TestAssignIt_Execute_ActionsRecorded(t *testing.T) {
	helper := &AssignIt{}
	mockConn := &mockConnectorForAssignIt{}
	ctx := context.Background()

	input := helpers.HelperInput{
		ContactID: "contact-123",
		Config: map[string]interface{}{
			"owner_id": "user-456",
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
	if action.Type != "field_updated" {
		t.Errorf("Expected action type 'field_updated', got '%s'", action.Type)
	}
	if action.Target != "owner_id" {
		t.Errorf("Expected action target 'owner_id', got '%s'", action.Target)
	}
	if action.Value != "user-456" {
		t.Errorf("Expected action value 'user-456', got '%v'", action.Value)
	}
}

func TestAssignIt_Execute_ModifiedDataRecorded(t *testing.T) {
	helper := &AssignIt{}
	mockConn := &mockConnectorForAssignIt{}
	ctx := context.Background()

	input := helpers.HelperInput{
		ContactID: "contact-123",
		Config: map[string]interface{}{
			"owner_id": "user-456",
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

	if output.ModifiedData["owner_id"] != "user-456" {
		t.Errorf("Expected ModifiedData owner_id to be 'user-456', got '%v'", output.ModifiedData["owner_id"])
	}
}

func TestAssignIt_Execute_LogsRecorded(t *testing.T) {
	helper := &AssignIt{}
	mockConn := &mockConnectorForAssignIt{}
	ctx := context.Background()

	input := helpers.HelperInput{
		ContactID: "contact-123",
		Config: map[string]interface{}{
			"owner_id": "user-456",
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

	expectedLog := "Assigned contact contact-123 to owner user-456 via field 'owner_id'"
	if output.Logs[0] != expectedLog {
		t.Errorf("Expected log '%s', got '%s'", expectedLog, output.Logs[0])
	}
}

func TestAssignIt_Execute_EmptyOwnerFieldUsesDefault(t *testing.T) {
	helper := &AssignIt{}
	mockConn := &mockConnectorForAssignIt{}
	ctx := context.Background()

	input := helpers.HelperInput{
		ContactID: "contact-123",
		Config: map[string]interface{}{
			"owner_id":     "user-456",
			"owner_field": "", // empty string should use default
		},
		Connector: mockConn,
	}

	_, err := helper.Execute(ctx, input)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if mockConn.fieldsSet["owner_id"] != "user-456" {
		t.Error("Expected default owner_id field to be used")
	}
}
