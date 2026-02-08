package contact

import (
	"context"
	"fmt"
	"testing"

	"github.com/myfusionhelper/api/internal/connectors"
	"github.com/myfusionhelper/api/internal/helpers"
)

// mockConnectorForOwnIt implements connectors.CRMConnector for testing
type mockConnectorForOwnIt struct {
	fieldsSet      map[string]interface{}
	setFieldError  error
}

func (m *mockConnectorForOwnIt) SetContactFieldValue(ctx context.Context, contactID, fieldKey string, value interface{}) error {
	if m.setFieldError != nil {
		return m.setFieldError
	}
	if m.fieldsSet == nil {
		m.fieldsSet = make(map[string]interface{})
	}
	m.fieldsSet[fieldKey] = value
	return nil
}

// Stub implementations
func (m *mockConnectorForOwnIt) GetContacts(ctx context.Context, opts connectors.QueryOptions) (*connectors.ContactList, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForOwnIt) GetContact(ctx context.Context, contactID string) (*connectors.NormalizedContact, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForOwnIt) CreateContact(ctx context.Context, contact connectors.CreateContactInput) (*connectors.NormalizedContact, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForOwnIt) UpdateContact(ctx context.Context, contactID string, updates connectors.UpdateContactInput) (*connectors.NormalizedContact, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForOwnIt) DeleteContact(ctx context.Context, contactID string) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorForOwnIt) GetTags(ctx context.Context) ([]connectors.Tag, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForOwnIt) ApplyTag(ctx context.Context, contactID, tagID string) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorForOwnIt) RemoveTag(ctx context.Context, contactID, tagID string) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorForOwnIt) GetCustomFields(ctx context.Context) ([]connectors.CustomField, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForOwnIt) GetContactFieldValue(ctx context.Context, contactID, fieldKey string) (interface{}, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForOwnIt) TriggerAutomation(ctx context.Context, contactID, automationID string) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorForOwnIt) AchieveGoal(ctx context.Context, contactID, goalName, integration string) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorForOwnIt) TestConnection(ctx context.Context) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorForOwnIt) GetMetadata() connectors.ConnectorMetadata {
	return connectors.ConnectorMetadata{}
}
func (m *mockConnectorForOwnIt) GetCapabilities() []connectors.Capability {
	return nil
}

func TestOwnIt_GetMetadata(t *testing.T) {
	h := &OwnIt{}
	if h.GetName() != "Own It" {
		t.Errorf("expected name 'Own It', got '%s'", h.GetName())
	}
	if h.GetType() != "own_it" {
		t.Errorf("expected type 'own_it', got '%s'", h.GetType())
	}
	if h.GetCategory() != "contact" {
		t.Errorf("expected category 'contact', got '%s'", h.GetCategory())
	}
	if !h.RequiresCRM() {
		t.Error("expected RequiresCRM to be true")
	}
	if h.SupportedCRMs() != nil {
		t.Errorf("expected nil SupportedCRMs, got %v", h.SupportedCRMs())
	}
}

func TestOwnIt_GetConfigSchema(t *testing.T) {
	h := &OwnIt{}
	schema := h.GetConfigSchema()
	if schema["type"] != "object" {
		t.Errorf("expected type 'object', got '%v'", schema["type"])
	}
	props := schema["properties"].(map[string]interface{})
	if _, ok := props["owner_id"]; !ok {
		t.Error("schema missing owner_id property")
	}

	required := schema["required"].([]string)
	if len(required) != 1 || required[0] != "owner_id" {
		t.Errorf("expected required field owner_id, got %v", required)
	}
}

func TestOwnIt_ValidateConfig_MissingOwnerID(t *testing.T) {
	h := &OwnIt{}
	config := map[string]interface{}{}
	err := h.ValidateConfig(config)
	if err == nil {
		t.Error("expected error for missing owner_id")
	}
}

func TestOwnIt_ValidateConfig_EmptyOwnerID(t *testing.T) {
	h := &OwnIt{}
	config := map[string]interface{}{
		"owner_id": "",
	}
	err := h.ValidateConfig(config)
	if err == nil {
		t.Error("expected error for empty owner_id")
	}
}

func TestOwnIt_ValidateConfig_ValidString(t *testing.T) {
	h := &OwnIt{}
	config := map[string]interface{}{
		"owner_id": "user123",
	}
	err := h.ValidateConfig(config)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestOwnIt_ValidateConfig_ValidNumber(t *testing.T) {
	h := &OwnIt{}
	config := map[string]interface{}{
		"owner_id": 12345.0,
	}
	err := h.ValidateConfig(config)
	if err != nil {
		t.Errorf("expected no error for numeric owner_id, got %v", err)
	}
}

func TestOwnIt_Execute_BasicOwnerUpdate(t *testing.T) {
	h := &OwnIt{}
	mock := &mockConnectorForOwnIt{}

	input := helpers.HelperInput{
		ContactID: "12345",
		Config: map[string]interface{}{
			"owner_id": "user999",
		},
		Connector: mock,
	}

	output, err := h.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !output.Success {
		t.Error("expected success=true")
	}
	if output.Message != "Contact owner updated to user999" {
		t.Errorf("unexpected message: %s", output.Message)
	}

	if mock.fieldsSet["OwnerID"] != "user999" {
		t.Errorf("expected OwnerID to be set to user999, got %v", mock.fieldsSet["OwnerID"])
	}
}

func TestOwnIt_Execute_NumericOwnerID(t *testing.T) {
	h := &OwnIt{}
	mock := &mockConnectorForOwnIt{}

	input := helpers.HelperInput{
		ContactID: "12345",
		Config: map[string]interface{}{
			"owner_id": 789.0,
		},
		Connector: mock,
	}

	output, err := h.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !output.Success {
		t.Error("expected success=true")
	}

	if mock.fieldsSet["OwnerID"] != "789" {
		t.Errorf("expected OwnerID to be set to 789, got %v", mock.fieldsSet["OwnerID"])
	}
}

func TestOwnIt_Execute_SetFieldError(t *testing.T) {
	h := &OwnIt{}
	mock := &mockConnectorForOwnIt{
		setFieldError: fmt.Errorf("CRM API error"),
	}

	input := helpers.HelperInput{
		ContactID: "12345",
		Config: map[string]interface{}{
			"owner_id": "user123",
		},
		Connector: mock,
	}

	output, err := h.Execute(context.Background(), input)
	if err == nil {
		t.Error("expected error from SetContactFieldValue")
	}
	if output.Success {
		t.Error("expected success=false on error")
	}
	if output.Message == "" {
		t.Error("expected error message to be set")
	}
}

func TestOwnIt_Execute_ActionVerification(t *testing.T) {
	h := &OwnIt{}
	mock := &mockConnectorForOwnIt{}

	input := helpers.HelperInput{
		ContactID: "12345",
		Config: map[string]interface{}{
			"owner_id": "user456",
		},
		Connector: mock,
	}

	output, err := h.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(output.Actions) != 1 {
		t.Errorf("expected 1 action, got %d", len(output.Actions))
	}
	if output.Actions[0].Type != "field_updated" {
		t.Errorf("expected action type 'field_updated', got '%s'", output.Actions[0].Type)
	}
	if output.Actions[0].Target != "OwnerID" {
		t.Errorf("expected action target 'OwnerID', got '%s'", output.Actions[0].Target)
	}
	if output.Actions[0].Value != "user456" {
		t.Errorf("expected action value 'user456', got '%v'", output.Actions[0].Value)
	}

	if len(output.Logs) == 0 {
		t.Error("expected logs to be populated")
	}
}

func TestOwnIt_Execute_ModifiedDataVerification(t *testing.T) {
	h := &OwnIt{}
	mock := &mockConnectorForOwnIt{}

	input := helpers.HelperInput{
		ContactID: "12345",
		Config: map[string]interface{}{
			"owner_id": "user789",
		},
		Connector: mock,
	}

	output, err := h.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	modifiedData := output.ModifiedData
	if modifiedData["OwnerID"] != "user789" {
		t.Errorf("expected ModifiedData OwnerID=user789, got %v", modifiedData["OwnerID"])
	}
}
