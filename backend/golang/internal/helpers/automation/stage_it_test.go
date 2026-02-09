package automation

import (
	"context"
	"fmt"
	"testing"

	"github.com/myfusionhelper/api/internal/connectors"
	"github.com/myfusionhelper/api/internal/helpers"
)

type mockConnectorForStageIt struct {
	fieldValues       map[string]interface{}
	fieldsSet         map[string]interface{}
	goalCalls         []string
	getFieldError     error
	setFieldError     error
	achieveGoalError  error
}

func (m *mockConnectorForStageIt) GetContactFieldValue(ctx context.Context, contactID, fieldKey string) (interface{}, error) {
	if m.getFieldError != nil { return nil, m.getFieldError }
	if m.fieldValues == nil { return nil, nil }
	return m.fieldValues[fieldKey], nil
}

func (m *mockConnectorForStageIt) SetContactFieldValue(ctx context.Context, contactID, fieldKey string, value interface{}) error {
	if m.setFieldError != nil { return m.setFieldError }
	if m.fieldsSet == nil { m.fieldsSet = make(map[string]interface{}) }
	m.fieldsSet[fieldKey] = value
	return nil
}

func (m *mockConnectorForStageIt) AchieveGoal(ctx context.Context, contactID, goalName, integration string) error {
	if m.achieveGoalError != nil { return m.achieveGoalError }
	if m.goalCalls == nil { m.goalCalls = make([]string, 0) }
	m.goalCalls = append(m.goalCalls, goalName)
	return nil
}

// Stub implementations
func (m *mockConnectorForStageIt) GetContacts(ctx context.Context, opts connectors.QueryOptions) (*connectors.ContactList, error) { return nil, fmt.Errorf("not implemented") }
func (m *mockConnectorForStageIt) GetContact(ctx context.Context, contactID string) (*connectors.NormalizedContact, error) { return nil, fmt.Errorf("not implemented") }
func (m *mockConnectorForStageIt) CreateContact(ctx context.Context, input connectors.CreateContactInput) (*connectors.NormalizedContact, error) { return nil, fmt.Errorf("not implemented") }
func (m *mockConnectorForStageIt) UpdateContact(ctx context.Context, contactID string, updates connectors.UpdateContactInput) (*connectors.NormalizedContact, error) { return nil, fmt.Errorf("not implemented") }
func (m *mockConnectorForStageIt) DeleteContact(ctx context.Context, contactID string) error { return fmt.Errorf("not implemented") }
func (m *mockConnectorForStageIt) GetTags(ctx context.Context) ([]connectors.Tag, error) { return nil, fmt.Errorf("not implemented") }
func (m *mockConnectorForStageIt) ApplyTag(ctx context.Context, contactID string, tagID string) error { return fmt.Errorf("not implemented") }
func (m *mockConnectorForStageIt) RemoveTag(ctx context.Context, contactID string, tagID string) error { return fmt.Errorf("not implemented") }
func (m *mockConnectorForStageIt) GetCustomFields(ctx context.Context) ([]connectors.CustomField, error) { return nil, fmt.Errorf("not implemented") }
func (m *mockConnectorForStageIt) TriggerAutomation(ctx context.Context, contactID string, automationID string) error { return fmt.Errorf("not implemented") }
func (m *mockConnectorForStageIt) TestConnection(ctx context.Context) error { return fmt.Errorf("not implemented") }
func (m *mockConnectorForStageIt) GetMetadata() connectors.ConnectorMetadata { return connectors.ConnectorMetadata{PlatformSlug: "keap"} }
func (m *mockConnectorForStageIt) GetCapabilities() []connectors.Capability { return []connectors.Capability{connectors.CapDeals} }
func (m *mockConnectorForStageIt) SetOptInStatus(ctx context.Context, contactID string, optIn bool, reason string) error { return fmt.Errorf("not implemented") }

func TestStageIt_GetMetadata(t *testing.T) {
	h := &StageIt{}
	if h.GetName() != "Stage It" { t.Error("Wrong name") }
	if h.GetType() != "stage_it" { t.Error("Wrong type") }
	if h.GetCategory() != "automation" { t.Error("Wrong category") }
	if !h.RequiresCRM() { t.Error("Should require CRM") }
	if len(h.SupportedCRMs()) == 0 || h.SupportedCRMs()[0] != "keap" {
		t.Error("Should support keap")
	}
}

func TestStageIt_ValidateConfig_MissingBasicMatch(t *testing.T) {
	h := &StageIt{}
	if err := h.ValidateConfig(map[string]interface{}{"to_stage": "stage2"}); err == nil {
		t.Error("Expected error for missing basic_match")
	}
}

func TestStageIt_ValidateConfig_MissingToStage(t *testing.T) {
	h := &StageIt{}
	if err := h.ValidateConfig(map[string]interface{}{"basic_match": "stage1"}); err == nil {
		t.Error("Expected error for missing to_stage")
	}
}

func TestStageIt_ValidateConfig_Valid(t *testing.T) {
	h := &StageIt{}
	if err := h.ValidateConfig(map[string]interface{}{"basic_match": "stage1", "to_stage": "stage2"}); err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestStageIt_Execute_OpportunitiesFound(t *testing.T) {
	h := &StageIt{}
	mockConn := &mockConnectorForStageIt{
		fieldValues: map[string]interface{}{
			"_related.lead.stage.stage1": "1",
		},
	}
	output, err := h.Execute(context.Background(), helpers.HelperInput{
		ContactID: "c1",
		Config: map[string]interface{}{
			"basic_match": "stage1",
			"to_stage": "stage2",
		},
		Connector: mockConn,
	})
	if err != nil { t.Fatalf("Error: %v", err) }
	if !output.Success { t.Error("Should succeed") }
}

func TestStageIt_Execute_NoOpportunitiesFound(t *testing.T) {
	h := &StageIt{}
	mockConn := &mockConnectorForStageIt{}
	output, err := h.Execute(context.Background(), helpers.HelperInput{
		ContactID: "c1",
		Config: map[string]interface{}{
			"basic_match": "stage1",
			"to_stage": "stage2",
		},
		Connector: mockConn,
	})
	if err != nil { t.Fatalf("Error: %v", err) }
	if !output.Success { t.Error("Should succeed") }
	if output.Message != "No opportunities found matching stage stage1" {
		t.Errorf("Unexpected message: %s", output.Message)
	}
}

func TestStageIt_Execute_FoundGoal(t *testing.T) {
	h := &StageIt{}
	mockConn := &mockConnectorForStageIt{
		fieldValues: map[string]interface{}{
			"_related.lead.stage.stage1": "1",
		},
	}
	_, _ = h.Execute(context.Background(), helpers.HelperInput{
		ContactID: "c1",
		Config: map[string]interface{}{
			"basic_match": "stage1",
			"to_stage": "stage2",
			"found_goal": "Opportunity Found",
		},
		Connector: mockConn,
	})
	if len(mockConn.goalCalls) != 1 || mockConn.goalCalls[0] != "Opportunity Found" {
		t.Error("Should achieve found_goal")
	}
}

func TestStageIt_Execute_NotFoundGoal(t *testing.T) {
	h := &StageIt{}
	mockConn := &mockConnectorForStageIt{}
	_, _ = h.Execute(context.Background(), helpers.HelperInput{
		ContactID: "c1",
		Config: map[string]interface{}{
			"basic_match": "stage1",
			"to_stage": "stage2",
			"not_found_goal": "No Opportunity",
		},
		Connector: mockConn,
	})
	if len(mockConn.goalCalls) != 1 || mockConn.goalCalls[0] != "No Opportunity" {
		t.Error("Should achieve not_found_goal")
	}
}

func TestStageIt_Execute_UpdateFirst(t *testing.T) {
	h := &StageIt{}
	mockConn := &mockConnectorForStageIt{
		fieldValues: map[string]interface{}{
			"_related.lead.stage.stage1": "1",
		},
	}
	_, _ = h.Execute(context.Background(), helpers.HelperInput{
		ContactID: "c1",
		Config: map[string]interface{}{
			"basic_match": "stage1",
			"to_stage": "stage2",
			"opportunity_count": "first",
		},
		Connector: mockConn,
	})
	updateKey := "_related.lead.stage.stage1.update_first"
	if mockConn.fieldsSet[updateKey] != "stage2" {
		t.Error("Should update first opportunity")
	}
}

func TestStageIt_Execute_UpdateAll(t *testing.T) {
	h := &StageIt{}
	mockConn := &mockConnectorForStageIt{
		fieldValues: map[string]interface{}{
			"_related.lead.stage.stage1": "2",
		},
	}
	_, _ = h.Execute(context.Background(), helpers.HelperInput{
		ContactID: "c1",
		Config: map[string]interface{}{
			"basic_match": "stage1",
			"to_stage": "stage2",
			"opportunity_count": "all",
		},
		Connector: mockConn,
	})
	updateKey := "_related.lead.stage.stage1.update_all"
	if mockConn.fieldsSet[updateKey] != "stage2" {
		t.Error("Should update all opportunities")
	}
}

func TestStageIt_Execute_ActionsRecorded(t *testing.T) {
	h := &StageIt{}
	mockConn := &mockConnectorForStageIt{
		fieldValues: map[string]interface{}{
			"_related.lead.stage.stage1": "1",
		},
	}
	output, _ := h.Execute(context.Background(), helpers.HelperInput{
		ContactID: "c1",
		Config: map[string]interface{}{
			"basic_match": "stage1",
			"to_stage": "stage2",
		},
		Connector: mockConn,
	})
	if len(output.Actions) == 0 { t.Error("Expected actions") }
}

func TestStageIt_Execute_LogsRecorded(t *testing.T) {
	h := &StageIt{}
	mockConn := &mockConnectorForStageIt{}
	output, _ := h.Execute(context.Background(), helpers.HelperInput{
		ContactID: "c1",
		Config: map[string]interface{}{
			"basic_match": "stage1",
			"to_stage": "stage2",
		},
		Connector: mockConn,
	})
	if len(output.Logs) == 0 { t.Error("Expected logs") }
}
