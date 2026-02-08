package data

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/myfusionhelper/api/internal/connectors"
	"github.com/myfusionhelper/api/internal/helpers"
)

// Mock connector for split_it testing
type mockConnectorForSplit struct {
	fieldValues   map[string]interface{}
	updatedFields map[string]interface{}
	appliedTags   []string
	achievedGoals []string
}

func (m *mockConnectorForSplit) GetContactFieldValue(ctx context.Context, contactID, fieldKey string) (interface{}, error) {
	if m.fieldValues == nil {
		return nil, nil // Return nil for non-existent fields (no error)
	}
	val, ok := m.fieldValues[fieldKey]
	if !ok {
		return nil, nil
	}
	return val, nil
}

func (m *mockConnectorForSplit) SetContactFieldValue(ctx context.Context, contactID, fieldKey string, value interface{}) error {
	if m.updatedFields == nil {
		m.updatedFields = make(map[string]interface{})
	}
	m.updatedFields[fieldKey] = value
	return nil
}

func (m *mockConnectorForSplit) ApplyTag(ctx context.Context, contactID, tagID string) error {
	m.appliedTags = append(m.appliedTags, tagID)
	return nil
}

func (m *mockConnectorForSplit) AchieveGoal(ctx context.Context, contactID, goalName, integration string) error {
	m.achievedGoals = append(m.achievedGoals, goalName)
	return nil
}

// Stub implementations for CRMConnector interface
func (m *mockConnectorForSplit) GetContact(ctx context.Context, contactID string) (*connectors.NormalizedContact, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForSplit) GetContacts(ctx context.Context, opts connectors.QueryOptions) (*connectors.ContactList, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForSplit) CreateContact(ctx context.Context, input connectors.CreateContactInput) (*connectors.NormalizedContact, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForSplit) UpdateContact(ctx context.Context, contactID string, updates connectors.UpdateContactInput) (*connectors.NormalizedContact, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForSplit) DeleteContact(ctx context.Context, contactID string) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorForSplit) GetTags(ctx context.Context) ([]connectors.Tag, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForSplit) RemoveTag(ctx context.Context, contactID, tagID string) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorForSplit) GetCustomFields(ctx context.Context) ([]connectors.CustomField, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForSplit) TriggerAutomation(ctx context.Context, contactID, automationID string) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorForSplit) TestConnection(ctx context.Context) error {
	return nil
}
func (m *mockConnectorForSplit) GetMetadata() connectors.ConnectorMetadata {
	return connectors.ConnectorMetadata{}
}
func (m *mockConnectorForSplit) GetCapabilities() []connectors.Capability {
	return nil
}

// Test helper metadata
func TestSplitIt_Metadata(t *testing.T) {
	helper := &SplitIt{}

	if helper.GetName() != "Split It" {
		t.Errorf("Expected name 'Split It', got '%s'", helper.GetName())
	}
	if helper.GetType() != "split_it" {
		t.Errorf("Expected type 'split_it', got '%s'", helper.GetType())
	}
	if helper.GetCategory() != "data" {
		t.Errorf("Expected category 'data', got '%s'", helper.GetCategory())
	}
	if !helper.RequiresCRM() {
		t.Error("Expected RequiresCRM to be true")
	}

	schema := helper.GetConfigSchema()
	if schema == nil {
		t.Fatal("Expected config schema, got nil")
	}
}

// Test validation - missing mode
func TestSplitIt_ValidateConfig_MissingMode(t *testing.T) {
	helper := &SplitIt{}
	config := map[string]interface{}{
		"option_a":    "tag-a",
		"option_b":    "tag-b",
		"state_field": "split_state",
	}

	err := helper.ValidateConfig(config)
	if err == nil {
		t.Error("Expected validation error for missing mode")
	}
	if !strings.Contains(err.Error(), "mode") {
		t.Errorf("Expected error about mode, got: %v", err)
	}
}

// Test validation - invalid mode
func TestSplitIt_ValidateConfig_InvalidMode(t *testing.T) {
	helper := &SplitIt{}
	config := map[string]interface{}{
		"mode":        "invalid",
		"option_a":    "tag-a",
		"option_b":    "tag-b",
		"state_field": "split_state",
	}

	err := helper.ValidateConfig(config)
	if err == nil {
		t.Error("Expected validation error for invalid mode")
	}
	if !strings.Contains(err.Error(), "tag") || !strings.Contains(err.Error(), "goal") {
		t.Errorf("Expected error about tag/goal, got: %v", err)
	}
}

// Test validation - missing option_a
func TestSplitIt_ValidateConfig_MissingOptionA(t *testing.T) {
	helper := &SplitIt{}
	config := map[string]interface{}{
		"mode":        "tag",
		"option_b":    "tag-b",
		"state_field": "split_state",
	}

	err := helper.ValidateConfig(config)
	if err == nil {
		t.Error("Expected validation error for missing option_a")
	}
	if !strings.Contains(err.Error(), "option_a") {
		t.Errorf("Expected error about option_a, got: %v", err)
	}
}

// Test validation - missing option_b
func TestSplitIt_ValidateConfig_MissingOptionB(t *testing.T) {
	helper := &SplitIt{}
	config := map[string]interface{}{
		"mode":        "tag",
		"option_a":    "tag-a",
		"state_field": "split_state",
	}

	err := helper.ValidateConfig(config)
	if err == nil {
		t.Error("Expected validation error for missing option_b")
	}
	if !strings.Contains(err.Error(), "option_b") {
		t.Errorf("Expected error about option_b, got: %v", err)
	}
}

// Test validation - missing state_field
func TestSplitIt_ValidateConfig_MissingStateField(t *testing.T) {
	helper := &SplitIt{}
	config := map[string]interface{}{
		"mode":     "tag",
		"option_a": "tag-a",
		"option_b": "tag-b",
	}

	err := helper.ValidateConfig(config)
	if err == nil {
		t.Error("Expected validation error for missing state_field")
	}
	if !strings.Contains(err.Error(), "state_field") {
		t.Errorf("Expected error about state_field, got: %v", err)
	}
}

// Test validation - valid config
func TestSplitIt_ValidateConfig_Valid(t *testing.T) {
	helper := &SplitIt{}

	validConfigs := []map[string]interface{}{
		{
			"mode":        "tag",
			"option_a":    "tag-a",
			"option_b":    "tag-b",
			"state_field": "split_state",
		},
		{
			"mode":        "goal",
			"option_a":    "goal-a",
			"option_b":    "goal-b",
			"state_field": "split_state",
		},
	}

	for _, config := range validConfigs {
		err := helper.ValidateConfig(config)
		if err != nil {
			t.Errorf("Expected no validation error for %v, got: %v", config, err)
		}
	}
}

// Test execution - first run (no state) should choose A
func TestSplitIt_Execute_FirstRun_Tag(t *testing.T) {
	helper := &SplitIt{}

	mockConn := &mockConnectorForSplit{
		fieldValues: map[string]interface{}{},
	}

	input := helpers.HelperInput{
		ContactID: "123",
		Connector: mockConn,
		Config: map[string]interface{}{
			"mode":        "tag",
			"option_a":    "tag-a",
			"option_b":    "tag-b",
			"state_field": "split_state",
		},
	}

	output, err := helper.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if !output.Success {
		t.Error("Expected success=true")
	}

	// First run should choose A
	if len(mockConn.appliedTags) != 1 || mockConn.appliedTags[0] != "tag-a" {
		t.Errorf("Expected tag 'tag-a' to be applied, got: %v", mockConn.appliedTags)
	}

	// State should be set to A
	if state := mockConn.updatedFields["split_state"]; state != "A" {
		t.Errorf("Expected state 'A', got: %v", state)
	}
}

// Test execution - alternation (A -> B)
func TestSplitIt_Execute_Alternation_AtoB_Tag(t *testing.T) {
	helper := &SplitIt{}

	mockConn := &mockConnectorForSplit{
		fieldValues: map[string]interface{}{
			"split_state": "A", // Last choice was A
		},
	}

	input := helpers.HelperInput{
		ContactID: "123",
		Connector: mockConn,
		Config: map[string]interface{}{
			"mode":        "tag",
			"option_a":    "tag-a",
			"option_b":    "tag-b",
			"state_field": "split_state",
		},
	}

	_, err := helper.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Should choose B (alternating from A)
	if len(mockConn.appliedTags) != 1 || mockConn.appliedTags[0] != "tag-b" {
		t.Errorf("Expected tag 'tag-b' to be applied, got: %v", mockConn.appliedTags)
	}

	// State should be updated to B
	if state := mockConn.updatedFields["split_state"]; state != "B" {
		t.Errorf("Expected state 'B', got: %v", state)
	}
}

// Test execution - alternation (B -> A)
func TestSplitIt_Execute_Alternation_BtoA_Tag(t *testing.T) {
	helper := &SplitIt{}

	mockConn := &mockConnectorForSplit{
		fieldValues: map[string]interface{}{
			"split_state": "B", // Last choice was B
		},
	}

	input := helpers.HelperInput{
		ContactID: "123",
		Connector: mockConn,
		Config: map[string]interface{}{
			"mode":        "tag",
			"option_a":    "tag-a",
			"option_b":    "tag-b",
			"state_field": "split_state",
		},
	}

	_, err := helper.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Should choose A (alternating from B)
	if len(mockConn.appliedTags) != 1 || mockConn.appliedTags[0] != "tag-a" {
		t.Errorf("Expected tag 'tag-a' to be applied, got: %v", mockConn.appliedTags)
	}

	// State should be updated to A
	if state := mockConn.updatedFields["split_state"]; state != "A" {
		t.Errorf("Expected state 'A', got: %v", state)
	}
}

// Test execution - goal mode
func TestSplitIt_Execute_GoalMode(t *testing.T) {
	helper := &SplitIt{}

	mockConn := &mockConnectorForSplit{
		fieldValues: map[string]interface{}{},
	}

	input := helpers.HelperInput{
		ContactID: "123",
		Connector: mockConn,
		Config: map[string]interface{}{
			"mode":        "goal",
			"option_a":    "goal-a",
			"option_b":    "goal-b",
			"state_field": "split_state",
		},
	}

	output, err := helper.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if !output.Success {
		t.Error("Expected success=true")
	}

	// First run should choose A
	if len(mockConn.achievedGoals) != 1 || mockConn.achievedGoals[0] != "goal-a" {
		t.Errorf("Expected goal 'goal-a' to be achieved, got: %v", mockConn.achievedGoals)
	}

	// State should be set to A
	if state := mockConn.updatedFields["split_state"]; state != "A" {
		t.Errorf("Expected state 'A', got: %v", state)
	}
}

// Test execution - goal mode alternation
func TestSplitIt_Execute_GoalMode_Alternation(t *testing.T) {
	helper := &SplitIt{}

	mockConn := &mockConnectorForSplit{
		fieldValues: map[string]interface{}{
			"split_state": "A",
		},
	}

	input := helpers.HelperInput{
		ContactID: "123",
		Connector: mockConn,
		Config: map[string]interface{}{
			"mode":        "goal",
			"option_a":    "goal-a",
			"option_b":    "goal-b",
			"state_field": "split_state",
		},
	}

	_, err := helper.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Should choose B (alternating from A)
	if len(mockConn.achievedGoals) != 1 || mockConn.achievedGoals[0] != "goal-b" {
		t.Errorf("Expected goal 'goal-b' to be achieved, got: %v", mockConn.achievedGoals)
	}
}

// Test action logging
func TestSplitIt_Execute_ActionLogging(t *testing.T) {
	helper := &SplitIt{}

	mockConn := &mockConnectorForSplit{
		fieldValues: map[string]interface{}{},
	}

	input := helpers.HelperInput{
		ContactID: "123",
		Connector: mockConn,
		Config: map[string]interface{}{
			"mode":        "tag",
			"option_a":    "tag-a",
			"option_b":    "tag-b",
			"state_field": "split_state",
		},
	}

	output, err := helper.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Verify actions were logged
	if len(output.Actions) < 2 {
		t.Fatal("Expected at least 2 actions (tag_applied, field_updated)")
	}

	// Verify tag action
	foundTagAction := false
	foundFieldAction := false
	for _, action := range output.Actions {
		if action.Type == "tag_applied" {
			foundTagAction = true
		}
		if action.Type == "field_updated" {
			foundFieldAction = true
		}
	}

	if !foundTagAction {
		t.Error("Expected tag_applied action")
	}
	if !foundFieldAction {
		t.Error("Expected field_updated action")
	}

	// Verify logs
	if len(output.Logs) == 0 {
		t.Fatal("Expected logs to be generated")
	}
}

// Test multiple executions verify alternation
func TestSplitIt_Execute_MultipleExecutions(t *testing.T) {
	helper := &SplitIt{}

	mockConn := &mockConnectorForSplit{
		fieldValues: map[string]interface{}{},
	}

	config := map[string]interface{}{
		"mode":        "tag",
		"option_a":    "tag-a",
		"option_b":    "tag-b",
		"state_field": "split_state",
	}

	// Execute 4 times, should alternate A, B, A, B
	expectedSequence := []string{"tag-a", "tag-b", "tag-a", "tag-b"}

	for i := 0; i < 4; i++ {
		mockConn.appliedTags = make([]string, 0) // Reset for each iteration

		input := helpers.HelperInput{
			ContactID: "123",
			Connector: mockConn,
			Config:    config,
		}

		_, err := helper.Execute(context.Background(), input)
		if err != nil {
			t.Fatalf("Execution %d failed: %v", i+1, err)
		}

		if len(mockConn.appliedTags) != 1 || mockConn.appliedTags[0] != expectedSequence[i] {
			t.Errorf("Execution %d: expected tag '%s', got: %v", i+1, expectedSequence[i], mockConn.appliedTags)
		}

		// Simulate state persistence: copy updated state to fieldValues for next iteration
		if updatedState, ok := mockConn.updatedFields["split_state"]; ok {
			mockConn.fieldValues["split_state"] = updatedState
		}
	}
}
