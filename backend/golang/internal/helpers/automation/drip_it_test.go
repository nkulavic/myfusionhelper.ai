package automation

import (
	"context"
	"fmt"
	"testing"

	"github.com/myfusionhelper/api/internal/connectors"
	"github.com/myfusionhelper/api/internal/helpers"
)

type mockConnectorForDripIt struct {
	fieldValues            map[string]interface{}
	fieldsSet              map[string]interface{}
	triggerCalls           []string
	getFieldError          error
	setFieldError          error
	triggerAutomationError error
}

func (m *mockConnectorForDripIt) GetContactFieldValue(ctx context.Context, contactID, fieldKey string) (interface{}, error) {
	if m.getFieldError != nil {
		return nil, m.getFieldError
	}
	if m.fieldValues == nil {
		return nil, nil
	}
	return m.fieldValues[fieldKey], nil
}

func (m *mockConnectorForDripIt) SetContactFieldValue(ctx context.Context, contactID, fieldKey string, value interface{}) error {
	if m.setFieldError != nil {
		return m.setFieldError
	}
	if m.fieldsSet == nil {
		m.fieldsSet = make(map[string]interface{})
	}
	m.fieldsSet[fieldKey] = value
	return nil
}

func (m *mockConnectorForDripIt) TriggerAutomation(ctx context.Context, contactID, automationID string) error {
	if m.triggerAutomationError != nil {
		return m.triggerAutomationError
	}
	if m.triggerCalls == nil {
		m.triggerCalls = make([]string, 0)
	}
	m.triggerCalls = append(m.triggerCalls, automationID)
	return nil
}

// Stub implementations
func (m *mockConnectorForDripIt) GetContacts(ctx context.Context, opts connectors.QueryOptions) (*connectors.ContactList, error) { return nil, fmt.Errorf("not implemented") }
func (m *mockConnectorForDripIt) GetContact(ctx context.Context, contactID string) (*connectors.NormalizedContact, error) { return nil, fmt.Errorf("not implemented") }
func (m *mockConnectorForDripIt) CreateContact(ctx context.Context, input connectors.CreateContactInput) (*connectors.NormalizedContact, error) { return nil, fmt.Errorf("not implemented") }
func (m *mockConnectorForDripIt) UpdateContact(ctx context.Context, contactID string, updates connectors.UpdateContactInput) (*connectors.NormalizedContact, error) { return nil, fmt.Errorf("not implemented") }
func (m *mockConnectorForDripIt) DeleteContact(ctx context.Context, contactID string) error { return fmt.Errorf("not implemented") }
func (m *mockConnectorForDripIt) GetTags(ctx context.Context) ([]connectors.Tag, error) { return nil, fmt.Errorf("not implemented") }
func (m *mockConnectorForDripIt) ApplyTag(ctx context.Context, contactID string, tagID string) error { return fmt.Errorf("not implemented") }
func (m *mockConnectorForDripIt) RemoveTag(ctx context.Context, contactID string, tagID string) error { return fmt.Errorf("not implemented") }
func (m *mockConnectorForDripIt) GetCustomFields(ctx context.Context) ([]connectors.CustomField, error) { return nil, fmt.Errorf("not implemented") }
func (m *mockConnectorForDripIt) AchieveGoal(ctx context.Context, contactID string, goalName string, integration string) error { return fmt.Errorf("not implemented") }
func (m *mockConnectorForDripIt) TestConnection(ctx context.Context) error { return fmt.Errorf("not implemented") }
func (m *mockConnectorForDripIt) GetMetadata() connectors.ConnectorMetadata { return connectors.ConnectorMetadata{PlatformSlug: "test"} }
func (m *mockConnectorForDripIt) GetCapabilities() []connectors.Capability { return []connectors.Capability{connectors.CapContacts} }
func (m *mockConnectorForDripIt) SetOptInStatus(ctx context.Context, contactID string, optIn bool, reason string) error { return fmt.Errorf("not implemented") }

func TestDripIt_GetMetadata(t *testing.T) {
	h := &DripIt{}
	if h.GetName() != "Drip It" { t.Error("Wrong name") }
	if h.GetType() != "drip_it" { t.Error("Wrong type") }
	if h.GetCategory() != "automation" { t.Error("Wrong category") }
	if !h.RequiresCRM() { t.Error("Should require CRM") }
}

func TestDripIt_ValidateConfig_MissingSteps(t *testing.T) {
	h := &DripIt{}
	if err := h.ValidateConfig(map[string]interface{}{"state_field": "step"}); err == nil {
		t.Error("Expected error for missing steps")
	}
}

func TestDripIt_ValidateConfig_EmptySteps(t *testing.T) {
	h := &DripIt{}
	if err := h.ValidateConfig(map[string]interface{}{"steps": []interface{}{}, "state_field": "step"}); err == nil {
		t.Error("Expected error for empty steps")
	}
}

func TestDripIt_ValidateConfig_MissingStateField(t *testing.T) {
	h := &DripIt{}
	if err := h.ValidateConfig(map[string]interface{}{"steps": []interface{}{"auto1"}}); err == nil {
		t.Error("Expected error for missing state_field")
	}
}

func TestDripIt_ValidateConfig_Valid(t *testing.T) {
	h := &DripIt{}
	if err := h.ValidateConfig(map[string]interface{}{"steps": []interface{}{"auto1"}, "state_field": "step"}); err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestDripIt_Execute_FirstStep(t *testing.T) {
	h := &DripIt{}
	mockConn := &mockConnectorForDripIt{fieldValues: map[string]interface{}{}}
	output, err := h.Execute(context.Background(), helpers.HelperInput{
		ContactID: "c1",
		Config: map[string]interface{}{
			"steps": []interface{}{"auto1", "auto2"},
			"state_field": "step",
		},
		Connector: mockConn,
	})
	if err != nil { t.Fatalf("Error: %v", err) }
	if !output.Success { t.Error("Should succeed") }
	if len(mockConn.triggerCalls) != 1 || mockConn.triggerCalls[0] != "auto1" {
		t.Error("Should trigger auto1")
	}
	if mockConn.fieldsSet["step"] != "1" { t.Error("Should update state to 1") }
}

func TestDripIt_Execute_SecondStep(t *testing.T) {
	h := &DripIt{}
	mockConn := &mockConnectorForDripIt{fieldValues: map[string]interface{}{"step": "1"}}
	output, err := h.Execute(context.Background(), helpers.HelperInput{
		ContactID: "c1",
		Config: map[string]interface{}{
			"steps": []interface{}{"auto1", "auto2"},
			"state_field": "step",
		},
		Connector: mockConn,
	})
	if err != nil { t.Fatalf("Error: %v", err) }
	if !output.Success { t.Error("Should succeed") }
	if len(mockConn.triggerCalls) != 1 || mockConn.triggerCalls[0] != "auto2" {
		t.Error("Should trigger auto2")
	}
	if mockConn.fieldsSet["step"] != "2" { t.Error("Should update state to 2") }
}

func TestDripIt_Execute_Complete(t *testing.T) {
	h := &DripIt{}
	mockConn := &mockConnectorForDripIt{fieldValues: map[string]interface{}{"step": "2"}}
	output, err := h.Execute(context.Background(), helpers.HelperInput{
		ContactID: "c1",
		Config: map[string]interface{}{
			"steps": []interface{}{"auto1", "auto2"},
			"state_field": "step",
		},
		Connector: mockConn,
	})
	if err != nil { t.Fatalf("Error: %v", err) }
	if !output.Success { t.Error("Should succeed") }
	if len(mockConn.triggerCalls) != 0 { t.Error("Should not trigger when complete") }
}

func TestDripIt_Execute_TriggerError(t *testing.T) {
	h := &DripIt{}
	mockConn := &mockConnectorForDripIt{triggerAutomationError: fmt.Errorf("trigger failed")}
	_, err := h.Execute(context.Background(), helpers.HelperInput{
		ContactID: "c1",
		Config: map[string]interface{}{
			"steps": []interface{}{"auto1"},
			"state_field": "step",
		},
		Connector: mockConn,
	})
	if err == nil { t.Error("Expected error") }
}

func TestDripIt_Execute_ActionsRecorded(t *testing.T) {
	h := &DripIt{}
	mockConn := &mockConnectorForDripIt{}
	output, _ := h.Execute(context.Background(), helpers.HelperInput{
		ContactID: "c1",
		Config: map[string]interface{}{
			"steps": []interface{}{"auto1"},
			"state_field": "step",
		},
		Connector: mockConn,
	})
	if len(output.Actions) != 2 { t.Errorf("Expected 2 actions, got %d", len(output.Actions)) }
	if output.Actions[0].Type != "automation_triggered" { t.Error("Wrong action type") }
	if output.Actions[1].Type != "field_updated" { t.Error("Wrong action type") }
}

func TestDripIt_Execute_LogsRecorded(t *testing.T) {
	h := &DripIt{}
	mockConn := &mockConnectorForDripIt{}
	output, _ := h.Execute(context.Background(), helpers.HelperInput{
		ContactID: "c1",
		Config: map[string]interface{}{
			"steps": []interface{}{"auto1"},
			"state_field": "step",
		},
		Connector: mockConn,
	})
	if len(output.Logs) == 0 { t.Error("Expected logs") }
}
