package integration

import (
	"context"
	"fmt"
	"testing"

	"github.com/myfusionhelper/api/internal/connectors"
	"github.com/myfusionhelper/api/internal/helpers"
)

type mockConnectorForHookIt struct {
	goalsAchieved    []goalCall
	achieveGoalError error
}

type goalCall struct {
	contactID   string
	goalName    string
	integration string
}

func (m *mockConnectorForHookIt) AchieveGoal(ctx context.Context, contactID, goalName, integration string) error {
	if m.achieveGoalError != nil {
		return m.achieveGoalError
	}
	if m.goalsAchieved == nil {
		m.goalsAchieved = make([]goalCall, 0)
	}
	m.goalsAchieved = append(m.goalsAchieved, goalCall{
		contactID:   contactID,
		goalName:    goalName,
		integration: integration,
	})
	return nil
}

func (m *mockConnectorForHookIt) GetContacts(ctx context.Context, opts connectors.QueryOptions) (*connectors.ContactList, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForHookIt) GetContact(ctx context.Context, contactID string) (*connectors.NormalizedContact, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForHookIt) CreateContact(ctx context.Context, contact connectors.CreateContactInput) (*connectors.NormalizedContact, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForHookIt) UpdateContact(ctx context.Context, contactID string, updates connectors.UpdateContactInput) (*connectors.NormalizedContact, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForHookIt) DeleteContact(ctx context.Context, contactID string) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorForHookIt) GetTags(ctx context.Context) ([]connectors.Tag, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForHookIt) ApplyTag(ctx context.Context, contactID, tagID string) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorForHookIt) RemoveTag(ctx context.Context, contactID, tagID string) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorForHookIt) GetCustomFields(ctx context.Context) ([]connectors.CustomField, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForHookIt) GetContactFieldValue(ctx context.Context, contactID, fieldKey string) (interface{}, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForHookIt) SetContactFieldValue(ctx context.Context, contactID, fieldKey string, value interface{}) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorForHookIt) TriggerAutomation(ctx context.Context, contactID, automationID string) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorForHookIt) TestConnection(ctx context.Context) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorForHookIt) GetMetadata() connectors.ConnectorMetadata {
	return connectors.ConnectorMetadata{}
}
func (m *mockConnectorForHookIt) GetCapabilities() []connectors.Capability {
	return nil
}

func TestHookIt_GetMetadata(t *testing.T) {
	h := &HookIt{}
	if h.GetName() != "Hook It" {
		t.Error("wrong name")
	}
	if h.GetType() != "hook_it" {
		t.Error("wrong type")
	}
	if h.GetCategory() != "integration" {
		t.Error("wrong category")
	}
	if !h.RequiresCRM() {
		t.Error("should require CRM")
	}
}

func TestHookIt_ValidateConfig_AlwaysValid(t *testing.T) {
	// Config is flexible, should not error
	err := (&HookIt{}).ValidateConfig(map[string]interface{}{})
	if err != nil {
		t.Errorf("should be valid: %v", err)
	}
}

func TestHookIt_Execute_ContactAddAction(t *testing.T) {
	mock := &mockConnectorForHookIt{}

	output, err := (&HookIt{}).Execute(context.Background(), helpers.HelperInput{
		ContactID: "contact_123",
		HelperID:  "helper_456",
		Config: map[string]interface{}{
			"hook_action": "contact.add",
		},
		Connector: mock,
	})

	if err != nil {
		t.Fatal(err)
	}
	if !output.Success {
		t.Error("should succeed")
	}
	if len(mock.goalsAchieved) != 1 {
		t.Fatalf("expected 1 goal, got %d", len(mock.goalsAchieved))
	}
	expectedGoal := "newcontacthelper_456"
	if mock.goalsAchieved[0].goalName != expectedGoal {
		t.Errorf("expected goal %s, got %s", expectedGoal, mock.goalsAchieved[0].goalName)
	}
	if mock.goalsAchieved[0].integration != "myfusionhelper" {
		t.Error("integration should be myfusionhelper")
	}
}

func TestHookIt_Execute_InvoiceAddAction(t *testing.T) {
	mock := &mockConnectorForHookIt{}

	output, err := (&HookIt{}).Execute(context.Background(), helpers.HelperInput{
		ContactID: "contact_123",
		HelperID:  "helper_789",
		Config: map[string]interface{}{
			"hook_action": "invoice.add",
		},
		Connector: mock,
	})

	if err != nil {
		t.Fatal(err)
	}
	if !output.Success {
		t.Error("should succeed")
	}
	expectedGoal := "newinvoicehelper_789"
	if len(mock.goalsAchieved) != 1 || mock.goalsAchieved[0].goalName != expectedGoal {
		t.Errorf("expected goal %s, got %v", expectedGoal, mock.goalsAchieved)
	}
}

func TestHookIt_Execute_OrderAddAction(t *testing.T) {
	mock := &mockConnectorForHookIt{}

	output, err := (&HookIt{}).Execute(context.Background(), helpers.HelperInput{
		ContactID: "contact_123",
		HelperID:  "helper_999",
		Config: map[string]interface{}{
			"hook_action": "order.add",
		},
		Connector: mock,
	})

	if err != nil {
		t.Fatal(err)
	}
	if !output.Success {
		t.Error("should succeed")
	}
	expectedGoal := "neworderhelper_999"
	if len(mock.goalsAchieved) != 1 || mock.goalsAchieved[0].goalName != expectedGoal {
		t.Errorf("expected goal %s", expectedGoal)
	}
}

func TestHookIt_Execute_CustomAction(t *testing.T) {
	mock := &mockConnectorForHookIt{}

	output, err := (&HookIt{}).Execute(context.Background(), helpers.HelperInput{
		ContactID: "contact_123",
		HelperID:  "helper_abc",
		Config: map[string]interface{}{
			"hook_action": "custom.event",
		},
		Connector: mock,
	})

	if err != nil {
		t.Fatal(err)
	}
	if !output.Success {
		t.Error("should succeed")
	}
	expectedGoal := "custom.eventhelper_abc"
	if len(mock.goalsAchieved) != 1 || mock.goalsAchieved[0].goalName != expectedGoal {
		t.Errorf("expected goal %s", expectedGoal)
	}
}

func TestHookIt_Execute_CustomGoalPrefix(t *testing.T) {
	mock := &mockConnectorForHookIt{}

	output, err := (&HookIt{}).Execute(context.Background(), helpers.HelperInput{
		ContactID: "contact_123",
		HelperID:  "helper_xyz",
		Config: map[string]interface{}{
			"hook_action": "contact.add",
			"goal_prefix": "customprefix",
		},
		Connector: mock,
	})

	if err != nil {
		t.Fatal(err)
	}
	if !output.Success {
		t.Error("should succeed")
	}
	expectedGoal := "customprefixhelper_xyz"
	if len(mock.goalsAchieved) != 1 || mock.goalsAchieved[0].goalName != expectedGoal {
		t.Errorf("expected goal %s, got %s", expectedGoal, mock.goalsAchieved[0].goalName)
	}
}

func TestHookIt_Execute_ActionsArray(t *testing.T) {
	mock := &mockConnectorForHookIt{}

	output, err := (&HookIt{}).Execute(context.Background(), helpers.HelperInput{
		ContactID: "contact_123",
		HelperID:  "helper_def",
		Config: map[string]interface{}{
			"actions": []interface{}{
				map[string]interface{}{
					"event":     "custom.event1",
					"goal_name": "goal1",
				},
				map[string]interface{}{
					"event":     "custom.event2",
					"goal_name": "goal2",
				},
			},
		},
		Connector: mock,
	})

	if err != nil {
		t.Fatal(err)
	}
	if !output.Success {
		t.Error("should succeed")
	}
	if len(mock.goalsAchieved) != 2 {
		t.Errorf("expected 2 goals, got %d", len(mock.goalsAchieved))
	}
	if mock.goalsAchieved[0].goalName != "goal1" {
		t.Error("first goal should be goal1")
	}
	if mock.goalsAchieved[1].goalName != "goal2" {
		t.Error("second goal should be goal2")
	}
}

func TestHookIt_Execute_ActionsArrayWithMatchingEvent(t *testing.T) {
	mock := &mockConnectorForHookIt{}

	output, err := (&HookIt{}).Execute(context.Background(), helpers.HelperInput{
		ContactID: "contact_123",
		HelperID:  "helper_ghi",
		Config: map[string]interface{}{
			"hook_action": "specific.event",
			"actions": []interface{}{
				map[string]interface{}{
					"event":     "specific.event",
					"goal_name": "matched_goal",
				},
				map[string]interface{}{
					"event":     "other.event",
					"goal_name": "unmatched_goal",
				},
			},
		},
		Connector: mock,
	})

	if err != nil {
		t.Fatal(err)
	}
	if !output.Success {
		t.Error("should succeed")
	}
	// Should fire default hook_action goal + matching event goal
	if len(mock.goalsAchieved) != 2 {
		t.Errorf("expected 2 goals, got %d", len(mock.goalsAchieved))
	}
}

func TestHookIt_Execute_NoActionsNoEvents(t *testing.T) {
	mock := &mockConnectorForHookIt{}

	output, err := (&HookIt{}).Execute(context.Background(), helpers.HelperInput{
		ContactID: "contact_123",
		HelperID:  "helper_jkl",
		Config:    map[string]interface{}{},
		Connector: mock,
	})

	if err != nil {
		t.Fatal(err)
	}
	if !output.Success {
		t.Error("should succeed even with no actions")
	}
	if len(mock.goalsAchieved) != 0 {
		t.Error("should not achieve any goals")
	}
	if output.Message != "Webhook received, no matching events to process" {
		t.Errorf("wrong message: %s", output.Message)
	}
}

func TestHookIt_Execute_AchieveGoalError(t *testing.T) {
	mock := &mockConnectorForHookIt{
		achieveGoalError: fmt.Errorf("goal error"),
	}

	output, err := (&HookIt{}).Execute(context.Background(), helpers.HelperInput{
		ContactID: "contact_123",
		HelperID:  "helper_mno",
		Config: map[string]interface{}{
			"hook_action": "contact.add",
		},
		Connector: mock,
	})

	if err == nil {
		t.Error("should return error on AchieveGoal failure")
	}
	if output.Success {
		t.Error("should not succeed")
	}
}

func TestHookIt_Execute_ActionsArraySkipsEmpty(t *testing.T) {
	mock := &mockConnectorForHookIt{}

	output, err := (&HookIt{}).Execute(context.Background(), helpers.HelperInput{
		ContactID: "contact_123",
		HelperID:  "helper_pqr",
		Config: map[string]interface{}{
			"actions": []interface{}{
				map[string]interface{}{
					"event":     "event1",
					"goal_name": "goal1",
				},
				map[string]interface{}{
					"event":     "",
					"goal_name": "goal2",
				},
				map[string]interface{}{
					"event":     "event3",
					"goal_name": "",
				},
			},
		},
		Connector: mock,
	})

	if err != nil {
		t.Fatal(err)
	}
	// Should succeed with only valid actions
	if !output.Success {
		t.Error("should succeed")
	}
	if len(mock.goalsAchieved) != 1 {
		t.Errorf("expected 1 goal (skipped empty event/goal_name), got %d", len(mock.goalsAchieved))
	}
	if mock.goalsAchieved[0].goalName != "goal1" {
		t.Error("should only achieve goal1")
	}
}

func TestHookIt_Execute_InvalidActionsFormat(t *testing.T) {
	mock := &mockConnectorForHookIt{}

	output, err := (&HookIt{}).Execute(context.Background(), helpers.HelperInput{
		ContactID: "contact_123",
		HelperID:  "helper_stu",
		Config: map[string]interface{}{
			"actions": []interface{}{
				"invalid_string",
				map[string]interface{}{
					"event": "valid.event",
					// missing goal_name
				},
			},
		},
		Connector: mock,
	})

	if err != nil {
		t.Fatal(err)
	}
	// Should succeed but skip invalid actions
	if !output.Success {
		t.Error("should succeed")
	}
	if len(mock.goalsAchieved) != 0 {
		t.Error("should not achieve any goals from invalid actions")
	}
}
