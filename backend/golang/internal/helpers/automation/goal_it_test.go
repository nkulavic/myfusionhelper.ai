package automation

import (
	"context"
	"fmt"
	"testing"

	"github.com/myfusionhelper/api/internal/connectors"
	"github.com/myfusionhelper/api/internal/helpers"
)

type mockConnectorForGoalIt struct {
	achieveGoalCalls []struct{ contactID, goalName, integration string }
	achieveGoalError error
}

func (m *mockConnectorForGoalIt) AchieveGoal(ctx context.Context, contactID, goalName, integration string) error {
	if m.achieveGoalError != nil { return m.achieveGoalError }
	if m.achieveGoalCalls == nil {
		m.achieveGoalCalls = make([]struct{ contactID, goalName, integration string }, 0)
	}
	m.achieveGoalCalls = append(m.achieveGoalCalls, struct{ contactID, goalName, integration string }{contactID, goalName, integration})
	return nil
}

// Stub implementations
func (m *mockConnectorForGoalIt) GetContacts(ctx context.Context, opts connectors.QueryOptions) (*connectors.ContactList, error) { return nil, fmt.Errorf("not implemented") }
func (m *mockConnectorForGoalIt) GetContact(ctx context.Context, contactID string) (*connectors.NormalizedContact, error) { return nil, fmt.Errorf("not implemented") }
func (m *mockConnectorForGoalIt) CreateContact(ctx context.Context, input connectors.CreateContactInput) (*connectors.NormalizedContact, error) { return nil, fmt.Errorf("not implemented") }
func (m *mockConnectorForGoalIt) UpdateContact(ctx context.Context, contactID string, updates connectors.UpdateContactInput) (*connectors.NormalizedContact, error) { return nil, fmt.Errorf("not implemented") }
func (m *mockConnectorForGoalIt) DeleteContact(ctx context.Context, contactID string) error { return fmt.Errorf("not implemented") }
func (m *mockConnectorForGoalIt) GetTags(ctx context.Context) ([]connectors.Tag, error) { return nil, fmt.Errorf("not implemented") }
func (m *mockConnectorForGoalIt) ApplyTag(ctx context.Context, contactID string, tagID string) error { return fmt.Errorf("not implemented") }
func (m *mockConnectorForGoalIt) RemoveTag(ctx context.Context, contactID string, tagID string) error { return fmt.Errorf("not implemented") }
func (m *mockConnectorForGoalIt) GetCustomFields(ctx context.Context) ([]connectors.CustomField, error) { return nil, fmt.Errorf("not implemented") }
func (m *mockConnectorForGoalIt) GetContactFieldValue(ctx context.Context, contactID, fieldKey string) (interface{}, error) { return nil, fmt.Errorf("not implemented") }
func (m *mockConnectorForGoalIt) SetContactFieldValue(ctx context.Context, contactID, fieldKey string, value interface{}) error { return fmt.Errorf("not implemented") }
func (m *mockConnectorForGoalIt) TriggerAutomation(ctx context.Context, contactID string, automationID string) error { return fmt.Errorf("not implemented") }
func (m *mockConnectorForGoalIt) TestConnection(ctx context.Context) error { return fmt.Errorf("not implemented") }
func (m *mockConnectorForGoalIt) GetMetadata() connectors.ConnectorMetadata { return connectors.ConnectorMetadata{PlatformSlug: "keap"} }
func (m *mockConnectorForGoalIt) GetCapabilities() []connectors.Capability { return []connectors.Capability{connectors.CapGoals} }

func TestGoalIt_GetMetadata(t *testing.T) {
	h := &GoalIt{}
	if h.GetName() != "Goal It" { t.Error("Wrong name") }
	if h.GetType() != "goal_it" { t.Error("Wrong type") }
	if h.GetCategory() != "automation" { t.Error("Wrong category") }
	if !h.RequiresCRM() { t.Error("Should require CRM") }
	if len(h.SupportedCRMs()) == 0 || h.SupportedCRMs()[0] != "keap" {
		t.Error("Should support keap")
	}
}

func TestGoalIt_ValidateConfig_MissingGoalName(t *testing.T) {
	h := &GoalIt{}
	if err := h.ValidateConfig(map[string]interface{}{}); err == nil {
		t.Error("Expected error for missing goal_name")
	}
}

func TestGoalIt_ValidateConfig_EmptyGoalName(t *testing.T) {
	h := &GoalIt{}
	if err := h.ValidateConfig(map[string]interface{}{"goal_name": ""}); err == nil {
		t.Error("Expected error for empty goal_name")
	}
}

func TestGoalIt_ValidateConfig_Valid(t *testing.T) {
	h := &GoalIt{}
	if err := h.ValidateConfig(map[string]interface{}{"goal_name": "Purchase Complete"}); err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestGoalIt_Execute_Success(t *testing.T) {
	h := &GoalIt{}
	mockConn := &mockConnectorForGoalIt{}
	output, err := h.Execute(context.Background(), helpers.HelperInput{
		ContactID: "c123",
		Config: map[string]interface{}{
			"goal_name": "Purchase Complete",
		},
		Connector: mockConn,
	})
	if err != nil { t.Fatalf("Error: %v", err) }
	if !output.Success { t.Error("Should succeed") }
	if len(mockConn.achieveGoalCalls) != 1 { t.Error("Should call AchieveGoal") }
	if mockConn.achieveGoalCalls[0].goalName != "Purchase Complete" {
		t.Error("Wrong goal name")
	}
	if mockConn.achieveGoalCalls[0].integration != "mfh" {
		t.Error("Should use default integration mfh")
	}
}

func TestGoalIt_Execute_CustomIntegration(t *testing.T) {
	h := &GoalIt{}
	mockConn := &mockConnectorForGoalIt{}
	_, _ = h.Execute(context.Background(), helpers.HelperInput{
		ContactID: "c123",
		Config: map[string]interface{}{
			"goal_name": "Goal X",
			"integration": "custom",
		},
		Connector: mockConn,
	})
	if mockConn.achieveGoalCalls[0].integration != "custom" {
		t.Error("Should use custom integration")
	}
}

func TestGoalIt_Execute_Error(t *testing.T) {
	h := &GoalIt{}
	mockConn := &mockConnectorForGoalIt{achieveGoalError: fmt.Errorf("goal failed")}
	output, err := h.Execute(context.Background(), helpers.HelperInput{
		ContactID: "c123",
		Config: map[string]interface{}{"goal_name": "Goal"},
		Connector: mockConn,
	})
	if err == nil { t.Error("Expected error") }
	if output.Success { t.Error("Should not succeed") }
}

func TestGoalIt_Execute_ActionsRecorded(t *testing.T) {
	h := &GoalIt{}
	mockConn := &mockConnectorForGoalIt{}
	output, _ := h.Execute(context.Background(), helpers.HelperInput{
		ContactID: "c123",
		Config: map[string]interface{}{"goal_name": "Goal"},
		Connector: mockConn,
	})
	if len(output.Actions) != 1 { t.Error("Expected 1 action") }
	if output.Actions[0].Type != "goal_achieved" { t.Error("Wrong action type") }
	if output.Actions[0].Value != "Goal" { t.Error("Wrong action value") }
}

func TestGoalIt_Execute_LogsRecorded(t *testing.T) {
	h := &GoalIt{}
	mockConn := &mockConnectorForGoalIt{}
	output, _ := h.Execute(context.Background(), helpers.HelperInput{
		ContactID: "c123",
		Config: map[string]interface{}{"goal_name": "Goal"},
		Connector: mockConn,
	})
	if len(output.Logs) == 0 { t.Error("Expected logs") }
}
