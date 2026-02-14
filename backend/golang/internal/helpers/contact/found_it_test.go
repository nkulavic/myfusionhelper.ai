package contact

import (
	"context"
	"fmt"
	"testing"

	"github.com/myfusionhelper/api/internal/connectors"
	"github.com/myfusionhelper/api/internal/helpers"
)

type mockConnectorForFoundIt struct {
	fieldValues    map[string]interface{}
	getFieldError  error
	tagsApplied    []string
	applyTagError  error
	goalsAchieved  []string
	achieveGoalError error
}

func (m *mockConnectorForFoundIt) GetContactFieldValue(ctx context.Context, contactID, fieldKey string) (interface{}, error) {
	if m.getFieldError != nil {
		return nil, m.getFieldError
	}
	if m.fieldValues != nil {
		return m.fieldValues[fieldKey], nil
	}
	return nil, nil
}

func (m *mockConnectorForFoundIt) ApplyTag(ctx context.Context, contactID, tagID string) error {
	if m.applyTagError != nil {
		return m.applyTagError
	}
	if m.tagsApplied == nil {
		m.tagsApplied = make([]string, 0)
	}
	m.tagsApplied = append(m.tagsApplied, tagID)
	return nil
}

func (m *mockConnectorForFoundIt) AchieveGoal(ctx context.Context, contactID, goalName, integration string) error {
	if m.achieveGoalError != nil {
		return m.achieveGoalError
	}
	if m.goalsAchieved == nil {
		m.goalsAchieved = make([]string, 0)
	}
	m.goalsAchieved = append(m.goalsAchieved, goalName)
	return nil
}

func (m *mockConnectorForFoundIt) GetContacts(ctx context.Context, opts connectors.QueryOptions) (*connectors.ContactList, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForFoundIt) GetContact(ctx context.Context, contactID string) (*connectors.NormalizedContact, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForFoundIt) CreateContact(ctx context.Context, contact connectors.CreateContactInput) (*connectors.NormalizedContact, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForFoundIt) UpdateContact(ctx context.Context, contactID string, updates connectors.UpdateContactInput) (*connectors.NormalizedContact, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForFoundIt) DeleteContact(ctx context.Context, contactID string) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorForFoundIt) GetTags(ctx context.Context) ([]connectors.Tag, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForFoundIt) RemoveTag(ctx context.Context, contactID, tagID string) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorForFoundIt) GetCustomFields(ctx context.Context) ([]connectors.CustomField, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForFoundIt) SetContactFieldValue(ctx context.Context, contactID, fieldKey string, value interface{}) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorForFoundIt) TriggerAutomation(ctx context.Context, contactID, automationID string) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorForFoundIt) TestConnection(ctx context.Context) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorForFoundIt) GetMetadata() connectors.ConnectorMetadata {
	return connectors.ConnectorMetadata{}
}
func (m *mockConnectorForFoundIt) GetCapabilities() []connectors.Capability {
	return nil
}

func TestFoundIt_GetMetadata(t *testing.T) {
	h := &FoundIt{}
	if h.GetName() != "Found It" {
		t.Error("wrong name")
	}
	if h.GetType() != "found_it" {
		t.Error("wrong type")
	}
	if h.GetCategory() != "contact" {
		t.Error("wrong category")
	}
	if !h.RequiresCRM() {
		t.Error("should require CRM")
	}
}

func TestFoundIt_ValidateConfig_MissingCheckField(t *testing.T) {
	err := (&FoundIt{}).ValidateConfig(map[string]interface{}{})
	if err == nil {
		t.Error("should error on missing check_field")
	}
}

func TestFoundIt_ValidateConfig_EmptyCheckField(t *testing.T) {
	err := (&FoundIt{}).ValidateConfig(map[string]interface{}{
		"check_field": "",
	})
	if err == nil {
		t.Error("should error on empty check_field")
	}
}

func TestFoundIt_ValidateConfig_Valid(t *testing.T) {
	err := (&FoundIt{}).ValidateConfig(map[string]interface{}{
		"check_field": "Email",
	})
	if err != nil {
		t.Errorf("should be valid: %v", err)
	}
}

func TestFoundIt_Execute_FieldHasValue_NoActions(t *testing.T) {
	mock := &mockConnectorForFoundIt{
		fieldValues: map[string]interface{}{
			"Email": "john@example.com",
		},
	}
	output, err := (&FoundIt{}).Execute(context.Background(), helpers.HelperInput{
		ContactID: "123",
		Config: map[string]interface{}{
			"check_field": "Email",
		},
		Connector: mock,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !output.Success {
		t.Error("should succeed")
	}
	if output.Message != "Field 'Email' has a value" {
		t.Errorf("wrong message: %s", output.Message)
	}
	if len(output.Actions) != 0 {
		t.Errorf("expected 0 actions, got %d", len(output.Actions))
	}
}

func TestFoundIt_Execute_FieldEmpty_NoActions(t *testing.T) {
	mock := &mockConnectorForFoundIt{
		fieldValues: map[string]interface{}{
			"Email": "",
		},
	}
	output, err := (&FoundIt{}).Execute(context.Background(), helpers.HelperInput{
		ContactID: "123",
		Config: map[string]interface{}{
			"check_field": "Email",
		},
		Connector: mock,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !output.Success {
		t.Error("should succeed")
	}
	if output.Message != "Field 'Email' is empty" {
		t.Errorf("wrong message: %s", output.Message)
	}
}

func TestFoundIt_Execute_FieldHasValue_FoundTag(t *testing.T) {
	mock := &mockConnectorForFoundIt{
		fieldValues: map[string]interface{}{
			"Email": "john@example.com",
		},
	}
	output, err := (&FoundIt{}).Execute(context.Background(), helpers.HelperInput{
		ContactID: "123",
		Config: map[string]interface{}{
			"check_field":  "Email",
			"found_tag_id": "tag_123",
		},
		Connector: mock,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !output.Success {
		t.Error("should succeed")
	}
	if len(mock.tagsApplied) != 1 || mock.tagsApplied[0] != "tag_123" {
		t.Errorf("expected tag_123 to be applied, got %v", mock.tagsApplied)
	}
	if len(output.Actions) != 1 {
		t.Errorf("expected 1 action, got %d", len(output.Actions))
	}
	if output.Actions[0].Type != "tag_applied" || output.Actions[0].Value != "tag_123" {
		t.Error("wrong action")
	}
}

func TestFoundIt_Execute_FieldEmpty_NotFoundTag(t *testing.T) {
	mock := &mockConnectorForFoundIt{
		fieldValues: map[string]interface{}{
			"Email": "",
		},
	}
	output, err := (&FoundIt{}).Execute(context.Background(), helpers.HelperInput{
		ContactID: "123",
		Config: map[string]interface{}{
			"check_field":      "Email",
			"not_found_tag_id": "tag_456",
		},
		Connector: mock,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !output.Success {
		t.Error("should succeed")
	}
	if len(mock.tagsApplied) != 1 || mock.tagsApplied[0] != "tag_456" {
		t.Errorf("expected tag_456 to be applied, got %v", mock.tagsApplied)
	}
}

func TestFoundIt_Execute_FieldHasValue_FoundGoal(t *testing.T) {
	mock := &mockConnectorForFoundIt{
		fieldValues: map[string]interface{}{
			"Phone": "555-1234",
		},
	}
	output, err := (&FoundIt{}).Execute(context.Background(), helpers.HelperInput{
		ContactID: "123",
		Config: map[string]interface{}{
			"check_field": "Phone",
			"found_goal":  "PhoneProvided",
		},
		Connector: mock,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !output.Success {
		t.Error("should succeed")
	}
	if len(mock.goalsAchieved) != 1 || mock.goalsAchieved[0] != "PhoneProvided" {
		t.Errorf("expected PhoneProvided goal, got %v", mock.goalsAchieved)
	}
	if len(output.Actions) != 1 {
		t.Errorf("expected 1 action, got %d", len(output.Actions))
	}
	if output.Actions[0].Type != "goal_achieved" {
		t.Error("wrong action type")
	}
}

func TestFoundIt_Execute_FieldEmpty_NotFoundGoal(t *testing.T) {
	mock := &mockConnectorForFoundIt{
		fieldValues: map[string]interface{}{
			"Phone": nil,
		},
	}
	output, err := (&FoundIt{}).Execute(context.Background(), helpers.HelperInput{
		ContactID: "123",
		Config: map[string]interface{}{
			"check_field":    "Phone",
			"not_found_goal": "PhoneMissing",
		},
		Connector: mock,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !output.Success {
		t.Error("should succeed")
	}
	if len(mock.goalsAchieved) != 1 || mock.goalsAchieved[0] != "PhoneMissing" {
		t.Errorf("expected PhoneMissing goal, got %v", mock.goalsAchieved)
	}
}

func TestFoundIt_Execute_AllActions(t *testing.T) {
	mock := &mockConnectorForFoundIt{
		fieldValues: map[string]interface{}{
			"Email": "test@example.com",
		},
	}
	output, err := (&FoundIt{}).Execute(context.Background(), helpers.HelperInput{
		ContactID: "123",
		Config: map[string]interface{}{
			"check_field":      "Email",
			"found_tag_id":     "tag_found",
			"not_found_tag_id": "tag_not_found",
			"found_goal":       "GoalFound",
			"not_found_goal":   "GoalNotFound",
		},
		Connector: mock,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !output.Success {
		t.Error("should succeed")
	}
	// Only found actions should execute
	if len(mock.tagsApplied) != 1 || mock.tagsApplied[0] != "tag_found" {
		t.Errorf("expected only tag_found, got %v", mock.tagsApplied)
	}
	if len(mock.goalsAchieved) != 1 || mock.goalsAchieved[0] != "GoalFound" {
		t.Errorf("expected only GoalFound, got %v", mock.goalsAchieved)
	}
	if len(output.Actions) != 2 {
		t.Errorf("expected 2 actions, got %d", len(output.Actions))
	}
}

func TestFoundIt_Execute_GetFieldError(t *testing.T) {
	mock := &mockConnectorForFoundIt{
		getFieldError: fmt.Errorf("field read error"),
	}
	output, err := (&FoundIt{}).Execute(context.Background(), helpers.HelperInput{
		ContactID: "123",
		Config: map[string]interface{}{
			"check_field": "Email",
		},
		Connector: mock,
	})
	if err == nil {
		t.Error("should return error on GetContactFieldValue failure")
	}
	if output.Success {
		t.Error("should not succeed")
	}
}

func TestFoundIt_Execute_ApplyTagError(t *testing.T) {
	mock := &mockConnectorForFoundIt{
		fieldValues: map[string]interface{}{
			"Email": "test@example.com",
		},
		applyTagError: fmt.Errorf("tag error"),
	}
	output, err := (&FoundIt{}).Execute(context.Background(), helpers.HelperInput{
		ContactID: "123",
		Config: map[string]interface{}{
			"check_field":  "Email",
			"found_tag_id": "tag_123",
		},
		Connector: mock,
	})
	if err != nil {
		t.Fatal(err)
	}
	// Should still succeed, just log the error
	if !output.Success {
		t.Error("should succeed even if tag application fails")
	}
	if len(output.Actions) != 0 {
		t.Error("should not add action for failed tag")
	}
}

func TestFoundIt_Execute_AchieveGoalError(t *testing.T) {
	mock := &mockConnectorForFoundIt{
		fieldValues: map[string]interface{}{
			"Email": "test@example.com",
		},
		achieveGoalError: fmt.Errorf("goal error"),
	}
	output, err := (&FoundIt{}).Execute(context.Background(), helpers.HelperInput{
		ContactID: "123",
		Config: map[string]interface{}{
			"check_field": "Email",
			"found_goal":  "TestGoal",
		},
		Connector: mock,
	})
	if err != nil {
		t.Fatal(err)
	}
	// Should still succeed, just log the error
	if !output.Success {
		t.Error("should succeed even if goal achievement fails")
	}
	if len(output.Actions) != 0 {
		t.Error("should not add action for failed goal")
	}
}

func TestFoundIt_Execute_NilValueTreatedAsEmpty(t *testing.T) {
	mock := &mockConnectorForFoundIt{
		fieldValues: map[string]interface{}{
			"Email": nil,
		},
	}
	output, err := (&FoundIt{}).Execute(context.Background(), helpers.HelperInput{
		ContactID: "123",
		Config: map[string]interface{}{
			"check_field": "Email",
		},
		Connector: mock,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !output.Success {
		t.Error("should succeed")
	}
	if output.Message != "Field 'Email' is empty" {
		t.Errorf("nil should be treated as empty, got: %s", output.Message)
	}
}
