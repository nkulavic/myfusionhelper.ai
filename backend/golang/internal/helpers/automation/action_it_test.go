package automation

import (
	"context"
	"fmt"
	"testing"

	"github.com/myfusionhelper/api/internal/connectors"
	"github.com/myfusionhelper/api/internal/helpers"
)

// mockConnectorForActionIt implements the CRMConnector interface for testing action_it
type mockConnectorForActionIt struct {
	triggerCalls        []struct{ contactID, automationID string }
	triggerAutomationError error
	failOnAutomationID  string // Fail only for specific automation ID
}

func (m *mockConnectorForActionIt) TriggerAutomation(ctx context.Context, contactID, automationID string) error {
	if m.triggerAutomationError != nil {
		return m.triggerAutomationError
	}
	if m.failOnAutomationID != "" && automationID == m.failOnAutomationID {
		return fmt.Errorf("automation %s failed", automationID)
	}
	if m.triggerCalls == nil {
		m.triggerCalls = make([]struct{ contactID, automationID string }, 0)
	}
	m.triggerCalls = append(m.triggerCalls, struct{contactID, automationID string}{contactID, automationID})
	return nil
}

// Stub implementations
func (m *mockConnectorForActionIt) GetContacts(ctx context.Context, opts connectors.QueryOptions) (*connectors.ContactList, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *mockConnectorForActionIt) GetContact(ctx context.Context, contactID string) (*connectors.NormalizedContact, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *mockConnectorForActionIt) CreateContact(ctx context.Context, input connectors.CreateContactInput) (*connectors.NormalizedContact, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *mockConnectorForActionIt) UpdateContact(ctx context.Context, contactID string, updates connectors.UpdateContactInput) (*connectors.NormalizedContact, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *mockConnectorForActionIt) DeleteContact(ctx context.Context, contactID string) error {
	return fmt.Errorf("not implemented")
}

func (m *mockConnectorForActionIt) GetTags(ctx context.Context) ([]connectors.Tag, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *mockConnectorForActionIt) ApplyTag(ctx context.Context, contactID string, tagID string) error {
	return fmt.Errorf("not implemented")
}

func (m *mockConnectorForActionIt) RemoveTag(ctx context.Context, contactID string, tagID string) error {
	return fmt.Errorf("not implemented")
}

func (m *mockConnectorForActionIt) GetCustomFields(ctx context.Context) ([]connectors.CustomField, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *mockConnectorForActionIt) GetContactFieldValue(ctx context.Context, contactID, fieldKey string) (interface{}, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *mockConnectorForActionIt) SetContactFieldValue(ctx context.Context, contactID, fieldKey string, value interface{}) error {
	return fmt.Errorf("not implemented")
}

func (m *mockConnectorForActionIt) AchieveGoal(ctx context.Context, contactID string, goalName string, integration string) error {
	return fmt.Errorf("not implemented")
}

func (m *mockConnectorForActionIt) TestConnection(ctx context.Context) error {
	return fmt.Errorf("not implemented")
}

func (m *mockConnectorForActionIt) GetMetadata() connectors.ConnectorMetadata {
	return connectors.ConnectorMetadata{
		PlatformSlug: "test",
		PlatformName: "Test Platform",
	}
}

func (m *mockConnectorForActionIt) GetCapabilities() []connectors.Capability {
	return []connectors.Capability{connectors.CapContacts, connectors.CapAutomations}
}

func (m *mockConnectorForActionIt) SetOptInStatus(ctx context.Context, contactID string, optIn bool, reason string) error {
	return nil
}

// Tests

func TestActionIt_GetMetadata(t *testing.T) {
	helper := &ActionIt{}

	if helper.GetName() != "Action It" {
		t.Errorf("Expected name 'Action It', got '%s'", helper.GetName())
	}
	if helper.GetType() != "action_it" {
		t.Errorf("Expected type 'action_it', got '%s'", helper.GetType())
	}
	if helper.GetCategory() != "automation" {
		t.Errorf("Expected category 'automation', got '%s'", helper.GetCategory())
	}
	if !helper.RequiresCRM() {
		t.Error("Expected RequiresCRM to be true")
	}
}

func TestActionIt_GetConfigSchema(t *testing.T) {
	helper := &ActionIt{}
	schema := helper.GetConfigSchema()

	if schema == nil {
		t.Fatal("Schema should not be nil")
	}

	props, ok := schema["properties"].(map[string]interface{})
	if !ok {
		t.Fatal("Schema should have properties")
	}

	if _, ok := props["automation_ids"]; !ok {
		t.Error("Schema should have automation_ids property")
	}
}

func TestActionIt_ValidateConfig_MissingAutomationIDs(t *testing.T) {
	helper := &ActionIt{}

	config := map[string]interface{}{}
	err := helper.ValidateConfig(config)
	if err == nil {
		t.Error("Expected error for missing automation_ids")
	}
}

func TestActionIt_ValidateConfig_EmptyAutomationIDs(t *testing.T) {
	helper := &ActionIt{}

	config := map[string]interface{}{
		"automation_ids": []interface{}{},
	}
	err := helper.ValidateConfig(config)
	if err == nil {
		t.Error("Expected error for empty automation_ids")
	}
}

func TestActionIt_ValidateConfig_InvalidAutomationIDsType(t *testing.T) {
	helper := &ActionIt{}

	config := map[string]interface{}{
		"automation_ids": "not-an-array",
	}
	err := helper.ValidateConfig(config)
	if err == nil {
		t.Error("Expected error for non-array automation_ids")
	}
}

func TestActionIt_ValidateConfig_Valid(t *testing.T) {
	helper := &ActionIt{}

	config := map[string]interface{}{
		"automation_ids": []interface{}{"auto-1", "auto-2"},
	}
	err := helper.ValidateConfig(config)
	if err != nil {
		t.Errorf("Expected no error for valid config, got: %v", err)
	}
}

func TestActionIt_Execute_Success(t *testing.T) {
	helper := &ActionIt{}
	mockConn := &mockConnectorForActionIt{}
	ctx := context.Background()

	input := helpers.HelperInput{
		ContactID: "contact-123",
		Config: map[string]interface{}{
			"automation_ids": []interface{}{"automation-1", "automation-2"},
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

	if len(mockConn.triggerCalls) != 2 {
		t.Fatalf("Expected 2 automation triggers, got %d", len(mockConn.triggerCalls))
	}

	if mockConn.triggerCalls[0].automationID != "automation-1" {
		t.Errorf("Expected first automation to be 'automation-1', got '%s'", mockConn.triggerCalls[0].automationID)
	}
	if mockConn.triggerCalls[1].automationID != "automation-2" {
		t.Errorf("Expected second automation to be 'automation-2', got '%s'", mockConn.triggerCalls[1].automationID)
	}
}

func TestActionIt_Execute_PartialFailure(t *testing.T) {
	helper := &ActionIt{}
	mockConn := &mockConnectorForActionIt{
		failOnAutomationID: "automation-2",
	}
	ctx := context.Background()

	input := helpers.HelperInput{
		ContactID: "contact-123",
		Config: map[string]interface{}{
			"automation_ids": []interface{}{"automation-1", "automation-2", "automation-3"},
		},
		Connector: mockConn,
	}

	output, err := helper.Execute(ctx, input)
	if err != nil {
		t.Fatalf("Expected no error (partial success), got: %v", err)
	}

	if !output.Success {
		t.Error("Expected success to be true with partial success")
	}

	// Should trigger automation-1 and automation-3, skip automation-2
	if len(mockConn.triggerCalls) != 2 {
		t.Fatalf("Expected 2 successful triggers, got %d", len(mockConn.triggerCalls))
	}

	if output.Message != "Triggered 2 of 3 automation(s)" {
		t.Errorf("Unexpected message: %s", output.Message)
	}
}

func TestActionIt_Execute_AllFail(t *testing.T) {
	helper := &ActionIt{}
	mockConn := &mockConnectorForActionIt{
		triggerAutomationError: fmt.Errorf("CRM API error"),
	}
	ctx := context.Background()

	input := helpers.HelperInput{
		ContactID: "contact-123",
		Config: map[string]interface{}{
			"automation_ids": []interface{}{"automation-1"},
		},
		Connector: mockConn,
	}

	output, err := helper.Execute(ctx, input)
	if err != nil {
		t.Fatalf("Expected no error (graceful failure), got: %v", err)
	}

	if output.Success {
		t.Error("Expected success to be false when all automations fail")
	}

	if output.Message != "Failed to trigger any automations" {
		t.Errorf("Unexpected message: %s", output.Message)
	}
}

func TestActionIt_Execute_ActionsRecorded(t *testing.T) {
	helper := &ActionIt{}
	mockConn := &mockConnectorForActionIt{}
	ctx := context.Background()

	input := helpers.HelperInput{
		ContactID: "contact-123",
		Config: map[string]interface{}{
			"automation_ids": []interface{}{"automation-1", "automation-2"},
		},
		Connector: mockConn,
	}

	output, err := helper.Execute(ctx, input)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(output.Actions) != 2 {
		t.Fatalf("Expected 2 actions, got %d", len(output.Actions))
	}

	for i, action := range output.Actions {
		if action.Type != "automation_triggered" {
			t.Errorf("Expected action type 'automation_triggered', got '%s'", action.Type)
		}
		if action.Target != "contact-123" {
			t.Errorf("Expected action target 'contact-123', got '%s'", action.Target)
		}
		expectedAutomation := fmt.Sprintf("automation-%d", i+1)
		if action.Value != expectedAutomation {
			t.Errorf("Expected action value '%s', got '%v'", expectedAutomation, action.Value)
		}
	}
}

func TestActionIt_Execute_LogsRecorded(t *testing.T) {
	helper := &ActionIt{}
	mockConn := &mockConnectorForActionIt{}
	ctx := context.Background()

	input := helpers.HelperInput{
		ContactID: "contact-123",
		Config: map[string]interface{}{
			"automation_ids": []interface{}{"automation-1"},
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

func TestActionIt_Execute_SingleAutomation(t *testing.T) {
	helper := &ActionIt{}
	mockConn := &mockConnectorForActionIt{}
	ctx := context.Background()

	input := helpers.HelperInput{
		ContactID: "contact-123",
		Config: map[string]interface{}{
			"automation_ids": []interface{}{"single-automation"},
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

	if output.Message != "Triggered 1 of 1 automation(s)" {
		t.Errorf("Unexpected message: %s", output.Message)
	}
}

func TestActionIt_Execute_MultipleAutomations(t *testing.T) {
	helper := &ActionIt{}
	mockConn := &mockConnectorForActionIt{}
	ctx := context.Background()

	input := helpers.HelperInput{
		ContactID: "contact-123",
		Config: map[string]interface{}{
			"automation_ids": []interface{}{"auto-1", "auto-2", "auto-3", "auto-4", "auto-5"},
		},
		Connector: mockConn,
	}

	output, err := helper.Execute(ctx, input)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(mockConn.triggerCalls) != 5 {
		t.Fatalf("Expected 5 automation triggers, got %d", len(mockConn.triggerCalls))
	}

	if output.Message != "Triggered 5 of 5 automation(s)" {
		t.Errorf("Unexpected message: %s", output.Message)
	}
}
