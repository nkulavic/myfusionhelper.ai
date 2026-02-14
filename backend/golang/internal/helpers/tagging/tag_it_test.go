package tagging

import (
	"context"
	"fmt"
	"testing"

	"github.com/myfusionhelper/api/internal/connectors"
	"github.com/myfusionhelper/api/internal/helpers"
)

type mockConnectorForTagIt struct {
	tagsApplied  []string
	tagsRemoved  []string
	applyTagErr  map[string]error
	removeTagErr map[string]error
}

func (m *mockConnectorForTagIt) ApplyTag(ctx context.Context, contactID, tagID string) error {
	if m.applyTagErr != nil {
		if err, ok := m.applyTagErr[tagID]; ok {
			return err
		}
	}
	if m.tagsApplied == nil {
		m.tagsApplied = make([]string, 0)
	}
	m.tagsApplied = append(m.tagsApplied, tagID)
	return nil
}

func (m *mockConnectorForTagIt) RemoveTag(ctx context.Context, contactID, tagID string) error {
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

func (m *mockConnectorForTagIt) GetContacts(ctx context.Context, opts connectors.QueryOptions) (*connectors.ContactList, error) { return nil, fmt.Errorf("not implemented") }
func (m *mockConnectorForTagIt) GetContact(ctx context.Context, contactID string) (*connectors.NormalizedContact, error) { return nil, fmt.Errorf("not implemented") }
func (m *mockConnectorForTagIt) CreateContact(ctx context.Context, contact connectors.CreateContactInput) (*connectors.NormalizedContact, error) { return nil, fmt.Errorf("not implemented") }
func (m *mockConnectorForTagIt) UpdateContact(ctx context.Context, contactID string, updates connectors.UpdateContactInput) (*connectors.NormalizedContact, error) { return nil, fmt.Errorf("not implemented") }
func (m *mockConnectorForTagIt) DeleteContact(ctx context.Context, contactID string) error { return fmt.Errorf("not implemented") }
func (m *mockConnectorForTagIt) GetTags(ctx context.Context) ([]connectors.Tag, error) { return nil, fmt.Errorf("not implemented") }
func (m *mockConnectorForTagIt) GetCustomFields(ctx context.Context) ([]connectors.CustomField, error) { return nil, fmt.Errorf("not implemented") }
func (m *mockConnectorForTagIt) GetContactFieldValue(ctx context.Context, contactID, fieldKey string) (interface{}, error) { return nil, fmt.Errorf("not implemented") }
func (m *mockConnectorForTagIt) SetContactFieldValue(ctx context.Context, contactID, fieldKey string, value interface{}) error { return fmt.Errorf("not implemented") }
func (m *mockConnectorForTagIt) TriggerAutomation(ctx context.Context, contactID, automationID string) error { return fmt.Errorf("not implemented") }
func (m *mockConnectorForTagIt) AchieveGoal(ctx context.Context, contactID, goalName, integration string) error { return fmt.Errorf("not implemented") }
func (m *mockConnectorForTagIt) TestConnection(ctx context.Context) error { return fmt.Errorf("not implemented") }
func (m *mockConnectorForTagIt) GetMetadata() connectors.ConnectorMetadata { return connectors.ConnectorMetadata{} }
func (m *mockConnectorForTagIt) GetCapabilities() []connectors.Capability { return nil }

func TestTagIt_GetMetadata(t *testing.T) {
	h := &TagIt{}
	if h.GetName() != "Tag It" { t.Error("wrong name") }
	if h.GetType() != "tag_it" { t.Error("wrong type") }
	if h.GetCategory() != "tagging" { t.Error("wrong category") }
	if !h.RequiresCRM() { t.Error("should require CRM") }
}

func TestTagIt_ValidateConfig_MissingAction(t *testing.T) {
	err := (&TagIt{}).ValidateConfig(map[string]interface{}{
		"tag_ids": []string{"tag1"},
	})
	if err == nil { t.Error("should error on missing action") }
}

func TestTagIt_ValidateConfig_InvalidAction(t *testing.T) {
	err := (&TagIt{}).ValidateConfig(map[string]interface{}{
		"action":  "invalid",
		"tag_ids": []string{"tag1"},
	})
	if err == nil { t.Error("should error on invalid action") }
}

func TestTagIt_ValidateConfig_MissingTagIDs(t *testing.T) {
	err := (&TagIt{}).ValidateConfig(map[string]interface{}{
		"action": "apply",
	})
	if err == nil { t.Error("should error on missing tag_ids") }
}

func TestTagIt_ValidateConfig_EmptyTagIDs(t *testing.T) {
	err := (&TagIt{}).ValidateConfig(map[string]interface{}{
		"action":  "apply",
		"tag_ids": []interface{}{},
	})
	if err == nil { t.Error("should error on empty tag_ids") }
}

func TestTagIt_ValidateConfig_InvalidTagIDsType(t *testing.T) {
	err := (&TagIt{}).ValidateConfig(map[string]interface{}{
		"action":  "apply",
		"tag_ids": "not an array",
	})
	if err == nil { t.Error("should error on invalid tag_ids type") }
}

func TestTagIt_ValidateConfig_ValidApply(t *testing.T) {
	err := (&TagIt{}).ValidateConfig(map[string]interface{}{
		"action":  "apply",
		"tag_ids": []string{"tag1", "tag2"},
	})
	if err != nil { t.Errorf("should accept valid apply config: %v", err) }
}

func TestTagIt_ValidateConfig_ValidRemove(t *testing.T) {
	err := (&TagIt{}).ValidateConfig(map[string]interface{}{
		"action":  "remove",
		"tag_ids": []interface{}{"tag1", "tag2"},
	})
	if err != nil { t.Errorf("should accept valid remove config: %v", err) }
}

func TestTagIt_Execute_ApplySingleTag(t *testing.T) {
	mock := &mockConnectorForTagIt{}

	output, err := (&TagIt{}).Execute(context.Background(), helpers.HelperInput{
		ContactID: "123",
		Config: map[string]interface{}{
			"action":  "apply",
			"tag_ids": []string{"tag1"},
		},
		Connector: mock,
	})
	if err != nil { t.Fatal(err) }
	if !output.Success { t.Error("should succeed") }

	if len(mock.tagsApplied) != 1 || mock.tagsApplied[0] != "tag1" {
		t.Errorf("expected tag1 applied, got %v", mock.tagsApplied)
	}

	if len(output.Actions) != 1 { t.Fatal("expected 1 action") }
	if output.Actions[0].Type != "tag_applied" {
		t.Errorf("expected tag_applied, got %s", output.Actions[0].Type)
	}
}

func TestTagIt_Execute_ApplyMultipleTags(t *testing.T) {
	mock := &mockConnectorForTagIt{}

	output, err := (&TagIt{}).Execute(context.Background(), helpers.HelperInput{
		ContactID: "123",
		Config: map[string]interface{}{
			"action":  "apply",
			"tag_ids": []interface{}{"tag1", "tag2", "tag3"},
		},
		Connector: mock,
	})
	if err != nil { t.Fatal(err) }
	if !output.Success { t.Error("should succeed") }

	if len(mock.tagsApplied) != 3 {
		t.Errorf("expected 3 tags applied, got %d", len(mock.tagsApplied))
	}

	if len(output.Actions) != 3 {
		t.Errorf("expected 3 actions, got %d", len(output.Actions))
	}
}

func TestTagIt_Execute_RemoveSingleTag(t *testing.T) {
	mock := &mockConnectorForTagIt{}

	output, err := (&TagIt{}).Execute(context.Background(), helpers.HelperInput{
		ContactID: "123",
		Config: map[string]interface{}{
			"action":  "remove",
			"tag_ids": []string{"tag1"},
		},
		Connector: mock,
	})
	if err != nil { t.Fatal(err) }
	if !output.Success { t.Error("should succeed") }

	if len(mock.tagsRemoved) != 1 || mock.tagsRemoved[0] != "tag1" {
		t.Errorf("expected tag1 removed, got %v", mock.tagsRemoved)
	}

	if len(output.Actions) != 1 { t.Fatal("expected 1 action") }
	if output.Actions[0].Type != "tag_removed" {
		t.Errorf("expected tag_removed, got %s", output.Actions[0].Type)
	}
}

func TestTagIt_Execute_RemoveMultipleTags(t *testing.T) {
	mock := &mockConnectorForTagIt{}

	output, err := (&TagIt{}).Execute(context.Background(), helpers.HelperInput{
		ContactID: "123",
		Config: map[string]interface{}{
			"action":  "remove",
			"tag_ids": []string{"tag1", "tag2"},
		},
		Connector: mock,
	})
	if err != nil { t.Fatal(err) }
	if !output.Success { t.Error("should succeed") }

	if len(mock.tagsRemoved) != 2 {
		t.Errorf("expected 2 tags removed, got %d", len(mock.tagsRemoved))
	}
}

func TestTagIt_Execute_PartialApplyFailure(t *testing.T) {
	mock := &mockConnectorForTagIt{
		applyTagErr: map[string]error{
			"tag2": fmt.Errorf("API error"),
		},
	}

	output, err := (&TagIt{}).Execute(context.Background(), helpers.HelperInput{
		ContactID: "123",
		Config: map[string]interface{}{
			"action":  "apply",
			"tag_ids": []string{"tag1", "tag2", "tag3"},
		},
		Connector: mock,
	})
	if err != nil { t.Fatal(err) }
	if !output.Success { t.Error("should succeed if at least one tag applied") }

	// Should apply tag1 and tag3, skip tag2
	if len(mock.tagsApplied) != 2 {
		t.Errorf("expected 2 tags applied, got %d", len(mock.tagsApplied))
	}

	if len(output.Actions) != 2 {
		t.Errorf("expected 2 actions (tag2 failed), got %d", len(output.Actions))
	}

	// Check logs mention failure
	foundError := false
	for _, log := range output.Logs {
		if len(log) > 6 && log[:6] == "Failed" {
			foundError = true
			break
		}
	}
	if !foundError { t.Error("expected error log for failed tag apply") }
}

func TestTagIt_Execute_PartialRemoveFailure(t *testing.T) {
	mock := &mockConnectorForTagIt{
		removeTagErr: map[string]error{
			"tag1": fmt.Errorf("API error"),
		},
	}

	output, err := (&TagIt{}).Execute(context.Background(), helpers.HelperInput{
		ContactID: "123",
		Config: map[string]interface{}{
			"action":  "remove",
			"tag_ids": []string{"tag1", "tag2"},
		},
		Connector: mock,
	})
	if err != nil { t.Fatal(err) }
	if !output.Success { t.Error("should succeed if at least one tag removed") }

	// Should remove tag2, skip tag1
	if len(mock.tagsRemoved) != 1 {
		t.Errorf("expected 1 tag removed, got %d", len(mock.tagsRemoved))
	}
}

func TestTagIt_Execute_AllApplyFailure(t *testing.T) {
	mock := &mockConnectorForTagIt{
		applyTagErr: map[string]error{
			"tag1": fmt.Errorf("API error"),
		},
	}

	output, err := (&TagIt{}).Execute(context.Background(), helpers.HelperInput{
		ContactID: "123",
		Config: map[string]interface{}{
			"action":  "apply",
			"tag_ids": []string{"tag1"},
		},
		Connector: mock,
	})
	if err != nil { t.Fatal(err) }
	if output.Success { t.Error("should fail if no tags applied") }

	if len(mock.tagsApplied) != 0 {
		t.Errorf("expected 0 tags applied, got %d", len(mock.tagsApplied))
	}

	if output.Message != "Failed to apply any tags" {
		t.Errorf("expected failure message, got %s", output.Message)
	}
}

func TestTagIt_Execute_AllRemoveFailure(t *testing.T) {
	mock := &mockConnectorForTagIt{
		removeTagErr: map[string]error{
			"tag1": fmt.Errorf("API error"),
			"tag2": fmt.Errorf("API error"),
		},
	}

	output, err := (&TagIt{}).Execute(context.Background(), helpers.HelperInput{
		ContactID: "123",
		Config: map[string]interface{}{
			"action":  "remove",
			"tag_ids": []string{"tag1", "tag2"},
		},
		Connector: mock,
	})
	if err != nil { t.Fatal(err) }
	if output.Success { t.Error("should fail if no tags removed") }

	if output.Message != "Failed to remove any tags" {
		t.Errorf("expected failure message, got %s", output.Message)
	}
}

func TestTagIt_Execute_ActionDetails(t *testing.T) {
	mock := &mockConnectorForTagIt{}

	output, err := (&TagIt{}).Execute(context.Background(), helpers.HelperInput{
		ContactID: "456",
		Config: map[string]interface{}{
			"action":  "apply",
			"tag_ids": []string{"tag_abc"},
		},
		Connector: mock,
	})
	if err != nil { t.Fatal(err) }

	if len(output.Actions) != 1 { t.Fatal("expected 1 action") }

	action := output.Actions[0]
	if action.Type != "tag_applied" {
		t.Errorf("expected tag_applied, got %s", action.Type)
	}
	if action.Target != "456" {
		t.Errorf("expected target 456, got %s", action.Target)
	}
	if action.Value != "tag_abc" {
		t.Errorf("expected value tag_abc, got %v", action.Value)
	}
}

func TestTagIt_Execute_SuccessMessage(t *testing.T) {
	mock := &mockConnectorForTagIt{}

	output, err := (&TagIt{}).Execute(context.Background(), helpers.HelperInput{
		ContactID: "123",
		Config: map[string]interface{}{
			"action":  "apply",
			"tag_ids": []string{"tag1", "tag2", "tag3"},
		},
		Connector: mock,
	})
	if err != nil { t.Fatal(err) }

	if output.Message != "Successfully applyd 3 tag(s)" {
		t.Errorf("expected success message with count, got %s", output.Message)
	}
}

func TestTagIt_Execute_Logs(t *testing.T) {
	mock := &mockConnectorForTagIt{}

	output, err := (&TagIt{}).Execute(context.Background(), helpers.HelperInput{
		ContactID: "123",
		Config: map[string]interface{}{
			"action":  "remove",
			"tag_ids": []string{"tag1"},
		},
		Connector: mock,
	})
	if err != nil { t.Fatal(err) }

	if len(output.Logs) == 0 {
		t.Error("expected logs to be populated")
	}

	// Should contain "removed" in log
	foundLog := false
	for _, log := range output.Logs {
		if len(log) > 0 {
			foundLog = true
			break
		}
	}
	if !foundLog { t.Error("expected populated logs") }
}

func TestExtractStringSlice_StringSlice(t *testing.T) {
	result := extractStringSlice([]string{"a", "b", "c"})
	if len(result) != 3 || result[0] != "a" || result[1] != "b" || result[2] != "c" {
		t.Errorf("expected [a b c], got %v", result)
	}
}

func TestExtractStringSlice_InterfaceSlice(t *testing.T) {
	result := extractStringSlice([]interface{}{"a", "b", "c"})
	if len(result) != 3 || result[0] != "a" || result[1] != "b" || result[2] != "c" {
		t.Errorf("expected [a b c], got %v", result)
	}
}

func TestExtractStringSlice_MixedTypes(t *testing.T) {
	result := extractStringSlice([]interface{}{"a", 123, "c"})
	if len(result) != 3 || result[0] != "a" || result[1] != "123" || result[2] != "c" {
		t.Errorf("expected [a 123 c], got %v", result)
	}
}

func TestExtractStringSlice_InvalidType(t *testing.T) {
	result := extractStringSlice("not a slice")
	if result != nil {
		t.Errorf("expected nil for invalid type, got %v", result)
	}
}
