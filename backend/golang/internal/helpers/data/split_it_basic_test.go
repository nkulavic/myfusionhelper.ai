package data

import (
	"context"
	"fmt"
	"testing"

	"github.com/myfusionhelper/api/internal/connectors"
	"github.com/myfusionhelper/api/internal/helpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Mock connector for split_it_basic testing
type mockConnectorForSplitItBasic struct {
	appliedTags      []string
	removedTags      []string
	achievedGoals    []string
}

func (m *mockConnectorForSplitItBasic) ApplyTag(ctx context.Context, contactID, tagID string) error {
	if m.appliedTags == nil {
		m.appliedTags = make([]string, 0)
	}
	m.appliedTags = append(m.appliedTags, tagID)
	return nil
}

func (m *mockConnectorForSplitItBasic) RemoveTag(ctx context.Context, contactID, tagID string) error {
	if m.removedTags == nil {
		m.removedTags = make([]string, 0)
	}
	m.removedTags = append(m.removedTags, tagID)
	return nil
}

func (m *mockConnectorForSplitItBasic) AchieveGoal(ctx context.Context, contactID, goalName, integration string) error {
	if m.achievedGoals == nil {
		m.achievedGoals = make([]string, 0)
	}
	m.achievedGoals = append(m.achievedGoals, goalName)
	return nil
}

// Stub implementations
func (m *mockConnectorForSplitItBasic) GetContact(ctx context.Context, contactID string) (*connectors.NormalizedContact, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForSplitItBasic) GetContacts(ctx context.Context, opts connectors.QueryOptions) (*connectors.ContactList, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForSplitItBasic) CreateContact(ctx context.Context, input connectors.CreateContactInput) (*connectors.NormalizedContact, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForSplitItBasic) UpdateContact(ctx context.Context, contactID string, updates connectors.UpdateContactInput) (*connectors.NormalizedContact, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForSplitItBasic) DeleteContact(ctx context.Context, contactID string) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorForSplitItBasic) GetTags(ctx context.Context) ([]connectors.Tag, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForSplitItBasic) GetCustomFields(ctx context.Context) ([]connectors.CustomField, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForSplitItBasic) GetContactFieldValue(ctx context.Context, contactID, fieldKey string) (interface{}, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForSplitItBasic) SetContactFieldValue(ctx context.Context, contactID, fieldKey string, value interface{}) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorForSplitItBasic) TriggerAutomation(ctx context.Context, contactID, automationID string) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorForSplitItBasic) TestConnection(ctx context.Context) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorForSplitItBasic) GetMetadata() connectors.ConnectorMetadata {
	return connectors.ConnectorMetadata{}
}
func (m *mockConnectorForSplitItBasic) GetCapabilities() []connectors.Capability {
	return nil
}

func TestSplitItBasic_Metadata(t *testing.T) {
	h := &SplitItBasic{}

	assert.Equal(t, "Split It Basic", h.GetName())
	assert.Equal(t, "split_it_basic", h.GetType())
	assert.Equal(t, "data", h.GetCategory())
	assert.NotEmpty(t, h.GetDescription())
	assert.True(t, h.RequiresCRM())
	assert.Nil(t, h.SupportedCRMs())
}

func TestSplitItBasic_GetConfigSchema(t *testing.T) {
	h := &SplitItBasic{}
	schema := h.GetConfigSchema()

	assert.Equal(t, "object", schema["type"])

	props, ok := schema["properties"].(map[string]interface{})
	require.True(t, ok)
	assert.Contains(t, props, "split_type")
	assert.Contains(t, props, "last_group")

	required, ok := schema["required"].([]string)
	require.True(t, ok)
	assert.Contains(t, required, "split_type")
	assert.Contains(t, required, "last_group")
}

func TestSplitItBasic_ValidateConfig_Success(t *testing.T) {
	h := &SplitItBasic{}

	tests := []struct {
		name   string
		config map[string]interface{}
	}{
		{
			name: "tag_split valid",
			config: map[string]interface{}{
				"split_type":  "tag_split",
				"last_group":  "a",
				"split_tag_a": "tag_123",
				"split_tag_b": "tag_456",
			},
		},
		{
			name: "goal_split valid",
			config: map[string]interface{}{
				"split_type":   "goal_split",
				"last_group":   "b",
				"split_goal_a": "goal_a",
				"split_goal_b": "goal_b",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := h.ValidateConfig(tt.config)
			assert.NoError(t, err)
		})
	}
}

func TestSplitItBasic_ValidateConfig_Errors(t *testing.T) {
	h := &SplitItBasic{}

	tests := []struct {
		name    string
		config  map[string]interface{}
		wantErr string
	}{
		{
			name:    "missing split_type",
			config:  map[string]interface{}{"last_group": "a"},
			wantErr: "split_type is required",
		},
		{
			name:    "invalid split_type",
			config:  map[string]interface{}{"split_type": "invalid", "last_group": "a"},
			wantErr: "must be 'tag_split' or 'goal_split'",
		},
		{
			name:    "missing last_group",
			config:  map[string]interface{}{"split_type": "tag_split"},
			wantErr: "last_group is required",
		},
		{
			name:    "invalid last_group",
			config:  map[string]interface{}{"split_type": "tag_split", "last_group": "c"},
			wantErr: "must be 'a' or 'b'",
		},
		{
			name: "tag_split missing split_tag_a",
			config: map[string]interface{}{
				"split_type":  "tag_split",
				"last_group":  "a",
				"split_tag_b": "tag_456",
			},
			wantErr: "split_tag_a is required",
		},
		{
			name: "tag_split missing split_tag_b",
			config: map[string]interface{}{
				"split_type":  "tag_split",
				"last_group":  "a",
				"split_tag_a": "tag_123",
			},
			wantErr: "split_tag_b is required",
		},
		{
			name: "goal_split missing split_goal_a",
			config: map[string]interface{}{
				"split_type":   "goal_split",
				"last_group":   "a",
				"split_goal_b": "goal_b",
			},
			wantErr: "split_goal_a is required",
		},
		{
			name: "goal_split missing split_goal_b",
			config: map[string]interface{}{
				"split_type":   "goal_split",
				"last_group":   "a",
				"split_goal_a": "goal_a",
			},
			wantErr: "split_goal_b is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := h.ValidateConfig(tt.config)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErr)
		})
	}
}

func TestSplitItBasic_Execute_TagSplit_LastGroupA_RunB(t *testing.T) {
	h := &SplitItBasic{}
	mockConnector := &mockConnectorForSplitItBasic{}

	input := helpers.HelperInput{
		ContactID: "contact_123",
		Connector: mockConnector,
		Config: map[string]interface{}{
			"split_type":  "tag_split",
			"last_group":  "a",
			"split_tag_a": "tag_a",
			"split_tag_b": "tag_b",
		},
	}

	output, err := h.Execute(context.Background(), input)

	assert.NoError(t, err)
	assert.True(t, output.Success)
	assert.Contains(t, output.Message, "group b")
	assert.Equal(t, "b", output.ModifiedData["run_group"])
	assert.Contains(t, mockConnector.appliedTags, "tag_b")
	assert.Contains(t, mockConnector.removedTags, "tag_b")
}

func TestSplitItBasic_Execute_TagSplit_LastGroupB_RunA(t *testing.T) {
	h := &SplitItBasic{}
	mockConnector := &mockConnectorForSplitItBasic{}

	input := helpers.HelperInput{
		ContactID: "contact_456",
		Connector: mockConnector,
		Config: map[string]interface{}{
			"split_type":  "tag_split",
			"last_group":  "b",
			"split_tag_a": "tag_a",
			"split_tag_b": "tag_b",
		},
	}

	output, err := h.Execute(context.Background(), input)

	assert.NoError(t, err)
	assert.True(t, output.Success)
	assert.Contains(t, output.Message, "group a")
	assert.Equal(t, "a", output.ModifiedData["run_group"])
	assert.Contains(t, mockConnector.appliedTags, "tag_a")
}

func TestSplitItBasic_Execute_GoalSplit_LastGroupA_RunB(t *testing.T) {
	h := &SplitItBasic{}
	mockConnector := &mockConnectorForSplitItBasic{}

	input := helpers.HelperInput{
		ContactID: "contact_789",
		Connector: mockConnector,
		Config: map[string]interface{}{
			"split_type":   "goal_split",
			"last_group":   "a",
			"split_goal_a": "goal_a",
			"split_goal_b": "goal_b",
		},
	}

	output, err := h.Execute(context.Background(), input)

	assert.NoError(t, err)
	assert.True(t, output.Success)
	assert.Contains(t, output.Message, "group b")
	assert.Equal(t, "b", output.ModifiedData["run_group"])
	assert.Contains(t, mockConnector.achievedGoals, "goal_b")
}

func TestSplitItBasic_Execute_GoalSplit_LastGroupB_RunA(t *testing.T) {
	h := &SplitItBasic{}
	mockConnector := &mockConnectorForSplitItBasic{}

	input := helpers.HelperInput{
		ContactID: "contact_101",
		Connector: mockConnector,
		Config: map[string]interface{}{
			"split_type":   "goal_split",
			"last_group":   "b",
			"split_goal_a": "goal_a",
			"split_goal_b": "goal_b",
		},
	}

	output, err := h.Execute(context.Background(), input)

	assert.NoError(t, err)
	assert.True(t, output.Success)
	assert.Contains(t, output.Message, "group a")
	assert.Equal(t, "a", output.ModifiedData["run_group"])
	assert.Contains(t, mockConnector.achievedGoals, "goal_a")
}

func TestSplitItBasic_Execute_Actions(t *testing.T) {
	h := &SplitItBasic{}
	mockConnector := &mockConnectorForSplitItBasic{}

	input := helpers.HelperInput{
		ContactID: "contact_actions",
		Connector: mockConnector,
		Config: map[string]interface{}{
			"split_type":  "tag_split",
			"last_group":  "a",
			"split_tag_a": "tag_a",
			"split_tag_b": "tag_b",
		},
	}

	output, err := h.Execute(context.Background(), input)

	assert.NoError(t, err)
	assert.Len(t, output.Actions, 1)
	assert.Equal(t, "tag_applied", output.Actions[0].Type)
	assert.Equal(t, "contact_actions", output.Actions[0].Target)
	assert.Equal(t, "tag_b", output.Actions[0].Value)
}

func TestSplitItBasic_Execute_Logs(t *testing.T) {
	h := &SplitItBasic{}
	mockConnector := &mockConnectorForSplitItBasic{}

	input := helpers.HelperInput{
		ContactID: "contact_logs",
		Connector: mockConnector,
		Config: map[string]interface{}{
			"split_type":  "tag_split",
			"last_group":  "b",
			"split_tag_a": "tag_a",
			"split_tag_b": "tag_b",
		},
	}

	output, err := h.Execute(context.Background(), input)

	assert.NoError(t, err)
	assert.NotEmpty(t, output.Logs)
	assert.Contains(t, output.Logs[0], "Last group was 'b'")
	assert.Contains(t, output.Logs[0], "running group 'a'")
}
