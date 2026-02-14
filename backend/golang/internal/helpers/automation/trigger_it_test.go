package automation

import (
	"context"
	"fmt"
	"testing"

	"github.com/myfusionhelper/api/internal/connectors"
	"github.com/myfusionhelper/api/internal/helpers"
)

type mockConnectorForTriggerIt struct {
	triggerCalls           []struct{ contactID, automationID string }
	triggerAutomationError error
}

func (m *mockConnectorForTriggerIt) TriggerAutomation(ctx context.Context, contactID, automationID string) error {
	if m.triggerAutomationError != nil { return m.triggerAutomationError }
	if m.triggerCalls == nil {
		m.triggerCalls = make([]struct{ contactID, automationID string }, 0)
	}
	m.triggerCalls = append(m.triggerCalls, struct{ contactID, automationID string }{contactID, automationID})
	return nil
}

// Stub implementations
func (m *mockConnectorForTriggerIt) GetContacts(ctx context.Context, opts connectors.QueryOptions) (*connectors.ContactList, error) { return nil, fmt.Errorf("not implemented") }
func (m *mockConnectorForTriggerIt) GetContact(ctx context.Context, contactID string) (*connectors.NormalizedContact, error) { return nil, fmt.Errorf("not implemented") }
func (m *mockConnectorForTriggerIt) CreateContact(ctx context.Context, input connectors.CreateContactInput) (*connectors.NormalizedContact, error) { return nil, fmt.Errorf("not implemented") }
func (m *mockConnectorForTriggerIt) UpdateContact(ctx context.Context, contactID string, updates connectors.UpdateContactInput) (*connectors.NormalizedContact, error) { return nil, fmt.Errorf("not implemented") }
func (m *mockConnectorForTriggerIt) DeleteContact(ctx context.Context, contactID string) error { return fmt.Errorf("not implemented") }
func (m *mockConnectorForTriggerIt) GetTags(ctx context.Context) ([]connectors.Tag, error) { return nil, fmt.Errorf("not implemented") }
func (m *mockConnectorForTriggerIt) ApplyTag(ctx context.Context, contactID string, tagID string) error { return fmt.Errorf("not implemented") }
func (m *mockConnectorForTriggerIt) RemoveTag(ctx context.Context, contactID string, tagID string) error { return fmt.Errorf("not implemented") }
func (m *mockConnectorForTriggerIt) GetCustomFields(ctx context.Context) ([]connectors.CustomField, error) { return nil, fmt.Errorf("not implemented") }
func (m *mockConnectorForTriggerIt) GetContactFieldValue(ctx context.Context, contactID, fieldKey string) (interface{}, error) { return nil, fmt.Errorf("not implemented") }
func (m *mockConnectorForTriggerIt) SetContactFieldValue(ctx context.Context, contactID, fieldKey string, value interface{}) error { return fmt.Errorf("not implemented") }
func (m *mockConnectorForTriggerIt) AchieveGoal(ctx context.Context, contactID string, goalName string, integration string) error { return fmt.Errorf("not implemented") }
func (m *mockConnectorForTriggerIt) TestConnection(ctx context.Context) error { return fmt.Errorf("not implemented") }
func (m *mockConnectorForTriggerIt) GetMetadata() connectors.ConnectorMetadata { return connectors.ConnectorMetadata{PlatformSlug: "test"} }
func (m *mockConnectorForTriggerIt) GetCapabilities() []connectors.Capability { return []connectors.Capability{connectors.CapAutomations} }
func (m *mockConnectorForTriggerIt) SetOptInStatus(ctx context.Context, contactID string, optIn bool, reason string) error { return fmt.Errorf("not implemented") }

func TestTriggerIt_GetMetadata(t *testing.T) {
	h := &TriggerIt{}
	if h.GetName() != "Trigger It" { t.Error("Wrong name") }
	if h.GetType() != "trigger_it" { t.Error("Wrong type") }
	if h.GetCategory() != "automation" { t.Error("Wrong category") }
	if !h.RequiresCRM() { t.Error("Should require CRM") }
}

func TestTriggerIt_GetConfigSchema(t *testing.T) {
	h := &TriggerIt{}
	schema := h.GetConfigSchema()
	props, ok := schema["properties"].(map[string]interface{})
	if !ok { t.Fatal("Schema should have properties") }
	if _, ok := props["automation_id"]; !ok {
		t.Error("Schema should have automation_id property")
	}
}

func TestTriggerIt_ValidateConfig_MissingAutomationID(t *testing.T) {
	h := &TriggerIt{}
	if err := h.ValidateConfig(map[string]interface{}{}); err == nil {
		t.Error("Expected error for missing automation_id")
	}
}

func TestTriggerIt_ValidateConfig_EmptyAutomationID(t *testing.T) {
	h := &TriggerIt{}
	if err := h.ValidateConfig(map[string]interface{}{"automation_id": ""}); err == nil {
		t.Error("Expected error for empty automation_id")
	}
}

func TestTriggerIt_ValidateConfig_InvalidAutomationIDType(t *testing.T) {
	h := &TriggerIt{}
	if err := h.ValidateConfig(map[string]interface{}{"automation_id": 123}); err == nil {
		t.Error("Expected error for non-string automation_id")
	}
}

func TestTriggerIt_ValidateConfig_Valid(t *testing.T) {
	h := &TriggerIt{}
	if err := h.ValidateConfig(map[string]interface{}{"automation_id": "auto-123"}); err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestTriggerIt_Execute_Success(t *testing.T) {
	h := &TriggerIt{}
	mockConn := &mockConnectorForTriggerIt{}
	output, err := h.Execute(context.Background(), helpers.HelperInput{
		ContactID: "contact-456",
		Config: map[string]interface{}{
			"automation_id": "automation-789",
		},
		Connector: mockConn,
	})
	if err != nil { t.Fatalf("Error: %v", err) }
	if !output.Success { t.Error("Should succeed") }
	if len(mockConn.triggerCalls) != 1 { t.Fatal("Should call TriggerAutomation") }
	if mockConn.triggerCalls[0].contactID != "contact-456" {
		t.Error("Wrong contact ID")
	}
	if mockConn.triggerCalls[0].automationID != "automation-789" {
		t.Error("Wrong automation ID")
	}
}

func TestTriggerIt_Execute_Error(t *testing.T) {
	h := &TriggerIt{}
	mockConn := &mockConnectorForTriggerIt{
		triggerAutomationError: fmt.Errorf("trigger failed"),
	}
	output, err := h.Execute(context.Background(), helpers.HelperInput{
		ContactID: "c1",
		Config: map[string]interface{}{"automation_id": "auto1"},
		Connector: mockConn,
	})
	if err == nil { t.Error("Expected error") }
	if output.Success { t.Error("Should not succeed") }
	if output.Message != "Failed to trigger automation auto1: trigger failed" {
		t.Errorf("Unexpected message: %s", output.Message)
	}
}

func TestTriggerIt_Execute_ActionsRecorded(t *testing.T) {
	h := &TriggerIt{}
	mockConn := &mockConnectorForTriggerIt{}
	output, _ := h.Execute(context.Background(), helpers.HelperInput{
		ContactID: "c1",
		Config: map[string]interface{}{"automation_id": "auto1"},
		Connector: mockConn,
	})
	if len(output.Actions) != 1 { t.Fatal("Expected 1 action") }
	if output.Actions[0].Type != "automation_triggered" { t.Error("Wrong action type") }
	if output.Actions[0].Target != "c1" { t.Error("Wrong action target") }
	if output.Actions[0].Value != "auto1" { t.Error("Wrong action value") }
}

func TestTriggerIt_Execute_LogsRecorded(t *testing.T) {
	h := &TriggerIt{}
	mockConn := &mockConnectorForTriggerIt{}
	output, _ := h.Execute(context.Background(), helpers.HelperInput{
		ContactID: "c1",
		Config: map[string]interface{}{"automation_id": "auto1"},
		Connector: mockConn,
	})
	if len(output.Logs) == 0 { t.Error("Expected logs") }
}

func TestTriggerIt_Execute_Message(t *testing.T) {
	h := &TriggerIt{}
	mockConn := &mockConnectorForTriggerIt{}
	output, _ := h.Execute(context.Background(), helpers.HelperInput{
		ContactID: "contact-123",
		Config: map[string]interface{}{"automation_id": "automation-456"},
		Connector: mockConn,
	})
	expected := "Triggered automation automation-456 for contact contact-123"
	if output.Message != expected {
		t.Errorf("Expected message '%s', got '%s'", expected, output.Message)
	}
}
