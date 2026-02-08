package tagging

import (
	"context"
	"fmt"
	"testing"

	"github.com/myfusionhelper/api/internal/connectors"
	"github.com/myfusionhelper/api/internal/helpers"
)

type mockConnectorForClearTags struct {
	contact       *connectors.NormalizedContact
	getContactErr error
	allTags       []connectors.Tag
	getTagsErr    error
	tagsRemoved   []string
	removeTagErr  map[string]error
}

func (m *mockConnectorForClearTags) GetContact(ctx context.Context, contactID string) (*connectors.NormalizedContact, error) {
	if m.getContactErr != nil {
		return nil, m.getContactErr
	}
	if m.contact != nil {
		return m.contact, nil
	}
	return &connectors.NormalizedContact{
		ID: contactID,
		Tags: []connectors.TagRef{
			{ID: "tag1", Name: "Marketing"},
			{ID: "tag2", Name: "VIP"},
		},
	}, nil
}

func (m *mockConnectorForClearTags) GetTags(ctx context.Context) ([]connectors.Tag, error) {
	if m.getTagsErr != nil {
		return nil, m.getTagsErr
	}
	if m.allTags != nil {
		return m.allTags, nil
	}
	return []connectors.Tag{
		{ID: "tag1", Name: "Marketing", Category: "sales"},
		{ID: "tag2", Name: "VIP", Category: "customer"},
	}, nil
}

func (m *mockConnectorForClearTags) RemoveTag(ctx context.Context, contactID, tagID string) error {
	if m.removeTagErr != nil {
		if err, ok := m.removeTagErr[tagID]; ok {
			return err
		}
	}
	if m.tagsRemoved == nil {
		m.tagsRemoved = make([]string, 0)
	}
	m.tagsRemoved = append(m.tagsRemoved, tagID)
	return nil
}

func (m *mockConnectorForClearTags) GetContacts(ctx context.Context, opts connectors.QueryOptions) (*connectors.ContactList, error) { return nil, fmt.Errorf("not implemented") }
func (m *mockConnectorForClearTags) CreateContact(ctx context.Context, contact connectors.CreateContactInput) (*connectors.NormalizedContact, error) { return nil, fmt.Errorf("not implemented") }
func (m *mockConnectorForClearTags) UpdateContact(ctx context.Context, contactID string, updates connectors.UpdateContactInput) (*connectors.NormalizedContact, error) { return nil, fmt.Errorf("not implemented") }
func (m *mockConnectorForClearTags) DeleteContact(ctx context.Context, contactID string) error { return fmt.Errorf("not implemented") }
func (m *mockConnectorForClearTags) ApplyTag(ctx context.Context, contactID, tagID string) error { return fmt.Errorf("not implemented") }
func (m *mockConnectorForClearTags) GetCustomFields(ctx context.Context) ([]connectors.CustomField, error) { return nil, fmt.Errorf("not implemented") }
func (m *mockConnectorForClearTags) GetContactFieldValue(ctx context.Context, contactID, fieldKey string) (interface{}, error) { return nil, fmt.Errorf("not implemented") }
func (m *mockConnectorForClearTags) SetContactFieldValue(ctx context.Context, contactID, fieldKey string, value interface{}) error { return fmt.Errorf("not implemented") }
func (m *mockConnectorForClearTags) TriggerAutomation(ctx context.Context, contactID, automationID string) error { return fmt.Errorf("not implemented") }
func (m *mockConnectorForClearTags) AchieveGoal(ctx context.Context, contactID, goalName, integration string) error { return fmt.Errorf("not implemented") }
func (m *mockConnectorForClearTags) TestConnection(ctx context.Context) error { return fmt.Errorf("not implemented") }
func (m *mockConnectorForClearTags) GetMetadata() connectors.ConnectorMetadata { return connectors.ConnectorMetadata{} }
func (m *mockConnectorForClearTags) GetCapabilities() []connectors.Capability { return nil }

func TestClearTags_GetMetadata(t *testing.T) {
	h := &ClearTags{}
	if h.GetName() != "Clear Tags" { t.Error("wrong name") }
	if h.GetType() != "clear_tags" { t.Error("wrong type") }
	if h.GetCategory() != "tagging" { t.Error("wrong category") }
	if !h.RequiresCRM() { t.Error("should require CRM") }
}

func TestClearTags_ValidateConfig_MissingMode(t *testing.T) {
	err := (&ClearTags{}).ValidateConfig(map[string]interface{}{})
	if err == nil { t.Error("should error on missing mode") }
}

func TestClearTags_ValidateConfig_InvalidMode(t *testing.T) {
	err := (&ClearTags{}).ValidateConfig(map[string]interface{}{
		"mode": "invalid",
	})
	if err == nil { t.Error("should error on invalid mode") }
}

func TestClearTags_ValidateConfig_SpecificModeMissingTagIDs(t *testing.T) {
	err := (&ClearTags{}).ValidateConfig(map[string]interface{}{
		"mode": "specific",
	})
	if err == nil { t.Error("should error on missing tag_ids in specific mode") }
}

func TestClearTags_ValidateConfig_SpecificModeEmptyTagIDs(t *testing.T) {
	err := (&ClearTags{}).ValidateConfig(map[string]interface{}{
		"mode":    "specific",
		"tag_ids": []interface{}{},
	})
	if err == nil { t.Error("should error on empty tag_ids in specific mode") }
}

func TestClearTags_ValidateConfig_PrefixModeMissingPrefix(t *testing.T) {
	err := (&ClearTags{}).ValidateConfig(map[string]interface{}{
		"mode": "prefix",
	})
	if err == nil { t.Error("should error on missing prefix in prefix mode") }
}

func TestClearTags_ValidateConfig_CategoryModeMissingCategory(t *testing.T) {
	err := (&ClearTags{}).ValidateConfig(map[string]interface{}{
		"mode": "category",
	})
	if err == nil { t.Error("should error on missing category in category mode") }
}

func TestClearTags_ValidateConfig_ValidConfigs(t *testing.T) {
	tests := []struct {
		name   string
		config map[string]interface{}
	}{
		{
			name:   "specific mode",
			config: map[string]interface{}{"mode": "specific", "tag_ids": []string{"tag1"}},
		},
		{
			name:   "all mode",
			config: map[string]interface{}{"mode": "all"},
		},
		{
			name:   "prefix mode",
			config: map[string]interface{}{"mode": "prefix", "prefix": "test"},
		},
		{
			name:   "category mode",
			config: map[string]interface{}{"mode": "category", "category": "sales"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := (&ClearTags{}).ValidateConfig(tt.config)
			if err != nil { t.Errorf("should accept valid config: %v", err) }
		})
	}
}

func TestClearTags_Execute_SpecificMode(t *testing.T) {
	mock := &mockConnectorForClearTags{
		contact: &connectors.NormalizedContact{
			ID: "123",
			Tags: []connectors.TagRef{
				{ID: "tag1", Name: "Marketing"},
				{ID: "tag2", Name: "VIP"},
				{ID: "tag3", Name: "Customer"},
			},
		},
	}

	output, err := (&ClearTags{}).Execute(context.Background(), helpers.HelperInput{
		ContactID: "123",
		Config: map[string]interface{}{
			"mode":    "specific",
			"tag_ids": []interface{}{"tag1", "tag3"},
		},
		Connector: mock,
	})
	if err != nil { t.Fatal(err) }
	if !output.Success { t.Error("should succeed") }

	if len(mock.tagsRemoved) != 2 {
		t.Errorf("expected 2 tags removed, got %d", len(mock.tagsRemoved))
	}
	if mock.tagsRemoved[0] != "tag1" || mock.tagsRemoved[1] != "tag3" {
		t.Errorf("expected tag1 and tag3 removed, got %v", mock.tagsRemoved)
	}

	if len(output.Actions) != 2 {
		t.Errorf("expected 2 actions, got %d", len(output.Actions))
	}
}

func TestClearTags_Execute_AllMode(t *testing.T) {
	mock := &mockConnectorForClearTags{
		contact: &connectors.NormalizedContact{
			ID: "123",
			Tags: []connectors.TagRef{
				{ID: "tag1", Name: "Marketing"},
				{ID: "tag2", Name: "VIP"},
				{ID: "tag3", Name: "Customer"},
			},
		},
	}

	output, err := (&ClearTags{}).Execute(context.Background(), helpers.HelperInput{
		ContactID: "123",
		Config: map[string]interface{}{
			"mode": "all",
		},
		Connector: mock,
	})
	if err != nil { t.Fatal(err) }
	if !output.Success { t.Error("should succeed") }

	if len(mock.tagsRemoved) != 3 {
		t.Errorf("expected 3 tags removed (all), got %d", len(mock.tagsRemoved))
	}
}

func TestClearTags_Execute_PrefixMode(t *testing.T) {
	mock := &mockConnectorForClearTags{
		contact: &connectors.NormalizedContact{
			ID: "123",
			Tags: []connectors.TagRef{
				{ID: "tag1", Name: "Marketing_New"},
				{ID: "tag2", Name: "Marketing_Old"},
				{ID: "tag3", Name: "VIP"},
			},
		},
	}

	output, err := (&ClearTags{}).Execute(context.Background(), helpers.HelperInput{
		ContactID: "123",
		Config: map[string]interface{}{
			"mode":   "prefix",
			"prefix": "Marketing",
		},
		Connector: mock,
	})
	if err != nil { t.Fatal(err) }
	if !output.Success { t.Error("should succeed") }

	if len(mock.tagsRemoved) != 2 {
		t.Errorf("expected 2 tags removed (prefix Marketing), got %d", len(mock.tagsRemoved))
	}
	if mock.tagsRemoved[0] != "tag1" || mock.tagsRemoved[1] != "tag2" {
		t.Errorf("expected tag1 and tag2 removed, got %v", mock.tagsRemoved)
	}
}

func TestClearTags_Execute_CategoryMode(t *testing.T) {
	mock := &mockConnectorForClearTags{
		contact: &connectors.NormalizedContact{
			ID: "123",
			Tags: []connectors.TagRef{
				{ID: "tag1", Name: "Marketing"},
				{ID: "tag2", Name: "Sales Lead"},
				{ID: "tag3", Name: "VIP"},
			},
		},
		allTags: []connectors.Tag{
			{ID: "tag1", Name: "Marketing", Category: "sales"},
			{ID: "tag2", Name: "Sales Lead", Category: "sales"},
			{ID: "tag3", Name: "VIP", Category: "customer"},
		},
	}

	output, err := (&ClearTags{}).Execute(context.Background(), helpers.HelperInput{
		ContactID: "123",
		Config: map[string]interface{}{
			"mode":     "category",
			"category": "sales",
		},
		Connector: mock,
	})
	if err != nil { t.Fatal(err) }
	if !output.Success { t.Error("should succeed") }

	if len(mock.tagsRemoved) != 2 {
		t.Errorf("expected 2 tags removed (category sales), got %d", len(mock.tagsRemoved))
	}
}

func TestClearTags_Execute_NoMatchingTags(t *testing.T) {
	mock := &mockConnectorForClearTags{
		contact: &connectors.NormalizedContact{
			ID: "123",
			Tags: []connectors.TagRef{
				{ID: "tag1", Name: "Marketing"},
			},
		},
	}

	output, err := (&ClearTags{}).Execute(context.Background(), helpers.HelperInput{
		ContactID: "123",
		Config: map[string]interface{}{
			"mode":   "prefix",
			"prefix": "Sales",
		},
		Connector: mock,
	})
	if err != nil { t.Fatal(err) }
	if !output.Success { t.Error("should succeed even with no matches") }

	if output.Message != "No matching tags to remove" {
		t.Errorf("expected 'No matching tags to remove', got %s", output.Message)
	}

	if len(mock.tagsRemoved) != 0 {
		t.Errorf("expected 0 tags removed, got %d", len(mock.tagsRemoved))
	}
}

func TestClearTags_Execute_PartialFailure(t *testing.T) {
	mock := &mockConnectorForClearTags{
		contact: &connectors.NormalizedContact{
			ID: "123",
			Tags: []connectors.TagRef{
				{ID: "tag1", Name: "Marketing"},
				{ID: "tag2", Name: "VIP"},
				{ID: "tag3", Name: "Customer"},
			},
		},
		removeTagErr: map[string]error{
			"tag2": fmt.Errorf("API error"),
		},
	}

	output, err := (&ClearTags{}).Execute(context.Background(), helpers.HelperInput{
		ContactID: "123",
		Config: map[string]interface{}{
			"mode": "all",
		},
		Connector: mock,
	})
	if err != nil { t.Fatal(err) }
	if !output.Success { t.Error("should succeed if at least one tag removed") }

	// Should remove tag1 and tag3, skip tag2 due to error
	if len(mock.tagsRemoved) != 2 {
		t.Errorf("expected 2 tags removed (tag2 failed), got %d", len(mock.tagsRemoved))
	}

	// Check logs mention failure
	foundError := false
	for _, log := range output.Logs {
		if len(log) > 6 && log[:6] == "Failed" {
			foundError = true
			break
		}
	}
	if !foundError { t.Error("expected error log for failed tag removal") }
}

func TestClearTags_Execute_AllFailure(t *testing.T) {
	mock := &mockConnectorForClearTags{
		contact: &connectors.NormalizedContact{
			ID: "123",
			Tags: []connectors.TagRef{
				{ID: "tag1", Name: "Marketing"},
			},
		},
		removeTagErr: map[string]error{
			"tag1": fmt.Errorf("API error"),
		},
	}

	output, err := (&ClearTags{}).Execute(context.Background(), helpers.HelperInput{
		ContactID: "123",
		Config: map[string]interface{}{
			"mode": "all",
		},
		Connector: mock,
	})
	if err != nil { t.Fatal(err) }
	if output.Success { t.Error("should fail if no tags removed") }

	if len(mock.tagsRemoved) != 0 {
		t.Errorf("expected 0 tags removed, got %d", len(mock.tagsRemoved))
	}
}

func TestClearTags_Execute_GetContactError(t *testing.T) {
	mock := &mockConnectorForClearTags{
		getContactErr: fmt.Errorf("API error"),
	}

	output, err := (&ClearTags{}).Execute(context.Background(), helpers.HelperInput{
		ContactID: "123",
		Config: map[string]interface{}{
			"mode": "all",
		},
		Connector: mock,
	})
	if err == nil { t.Error("should error when GetContact fails") }
	if output.Success { t.Error("should not succeed on error") }
}

func TestClearTags_Execute_GetTagsError(t *testing.T) {
	mock := &mockConnectorForClearTags{
		contact: &connectors.NormalizedContact{
			ID: "123",
			Tags: []connectors.TagRef{
				{ID: "tag1", Name: "Marketing"},
			},
		},
		getTagsErr: fmt.Errorf("API error"),
	}

	output, err := (&ClearTags{}).Execute(context.Background(), helpers.HelperInput{
		ContactID: "123",
		Config: map[string]interface{}{
			"mode":     "category",
			"category": "sales",
		},
		Connector: mock,
	})
	if err == nil { t.Error("should error when GetTags fails in category mode") }
	if output.Success { t.Error("should not succeed on error") }
}

func TestClearTags_Execute_ActionDetails(t *testing.T) {
	mock := &mockConnectorForClearTags{
		contact: &connectors.NormalizedContact{
			ID: "123",
			Tags: []connectors.TagRef{
				{ID: "tag1", Name: "Marketing"},
			},
		},
	}

	output, err := (&ClearTags{}).Execute(context.Background(), helpers.HelperInput{
		ContactID: "123",
		Config: map[string]interface{}{
			"mode":    "specific",
			"tag_ids": []string{"tag1"},
		},
		Connector: mock,
	})
	if err != nil { t.Fatal(err) }

	if len(output.Actions) != 1 { t.Fatal("expected 1 action") }

	action := output.Actions[0]
	if action.Type != "tag_removed" {
		t.Errorf("expected tag_removed action, got %s", action.Type)
	}
	if action.Target != "123" {
		t.Errorf("expected target 123, got %s", action.Target)
	}
	if action.Value != "tag1" {
		t.Errorf("expected value tag1, got %v", action.Value)
	}
}
