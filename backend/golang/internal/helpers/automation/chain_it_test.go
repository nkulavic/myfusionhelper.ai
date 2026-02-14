package automation

import (
	"context"
	"fmt"
	"testing"

	"github.com/myfusionhelper/api/internal/connectors"
	"github.com/myfusionhelper/api/internal/helpers"
)

// mockConnectorForChainIt - minimal mock since chain_it doesn't require CRM
type mockConnectorForChainIt struct{}

func (m *mockConnectorForChainIt) GetContacts(ctx context.Context, opts connectors.QueryOptions) (*connectors.ContactList, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForChainIt) GetContact(ctx context.Context, contactID string) (*connectors.NormalizedContact, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForChainIt) CreateContact(ctx context.Context, input connectors.CreateContactInput) (*connectors.NormalizedContact, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForChainIt) UpdateContact(ctx context.Context, contactID string, updates connectors.UpdateContactInput) (*connectors.NormalizedContact, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForChainIt) DeleteContact(ctx context.Context, contactID string) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorForChainIt) GetTags(ctx context.Context) ([]connectors.Tag, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForChainIt) ApplyTag(ctx context.Context, contactID string, tagID string) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorForChainIt) RemoveTag(ctx context.Context, contactID string, tagID string) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorForChainIt) GetCustomFields(ctx context.Context) ([]connectors.CustomField, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForChainIt) GetContactFieldValue(ctx context.Context, contactID, fieldKey string) (interface{}, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForChainIt) SetContactFieldValue(ctx context.Context, contactID, fieldKey string, value interface{}) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorForChainIt) TriggerAutomation(ctx context.Context, contactID string, automationID string) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorForChainIt) AchieveGoal(ctx context.Context, contactID string, goalName string, integration string) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorForChainIt) TestConnection(ctx context.Context) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorForChainIt) GetMetadata() connectors.ConnectorMetadata {
	return connectors.ConnectorMetadata{PlatformSlug: "test", PlatformName: "Test"}
}
func (m *mockConnectorForChainIt) GetCapabilities() []connectors.Capability {
	return []connectors.Capability{connectors.CapContacts}
}
func (m *mockConnectorForChainIt) SetOptInStatus(ctx context.Context, contactID string, optIn bool, reason string) error {
	return fmt.Errorf("not implemented")
}

// Tests

func TestChainIt_GetMetadata(t *testing.T) {
	helper := &ChainIt{}

	if helper.GetName() != "Chain It" {
		t.Errorf("Expected name 'Chain It', got '%s'", helper.GetName())
	}
	if helper.GetType() != "chain_it" {
		t.Errorf("Expected type 'chain_it', got '%s'", helper.GetType())
	}
	if helper.GetCategory() != "automation" {
		t.Errorf("Expected category 'automation', got '%s'", helper.GetCategory())
	}
	if helper.RequiresCRM() {
		t.Error("Expected RequiresCRM to be false")
	}
}

func TestChainIt_GetConfigSchema(t *testing.T) {
	helper := &ChainIt{}
	schema := helper.GetConfigSchema()

	props, ok := schema["properties"].(map[string]interface{})
	if !ok {
		t.Fatal("Schema should have properties")
	}

	if _, ok := props["helpers"]; !ok {
		t.Error("Schema should have helpers property")
	}
}

func TestChainIt_ValidateConfig_MissingHelpers(t *testing.T) {
	helper := &ChainIt{}
	err := helper.ValidateConfig(map[string]interface{}{})
	if err == nil {
		t.Error("Expected error for missing helpers")
	}
}

func TestChainIt_ValidateConfig_EmptyHelpers(t *testing.T) {
	helper := &ChainIt{}
	err := helper.ValidateConfig(map[string]interface{}{
		"helpers": []interface{}{},
	})
	if err == nil {
		t.Error("Expected error for empty helpers")
	}
}

func TestChainIt_ValidateConfig_InvalidHelpersType(t *testing.T) {
	helper := &ChainIt{}
	err := helper.ValidateConfig(map[string]interface{}{
		"helpers": "not-an-array",
	})
	if err == nil {
		t.Error("Expected error for non-array helpers")
	}
}

func TestChainIt_ValidateConfig_Valid(t *testing.T) {
	helper := &ChainIt{}
	err := helper.ValidateConfig(map[string]interface{}{
		"helpers": []interface{}{"tag_it", "copy_it"},
	})
	if err != nil {
		t.Errorf("Expected no error for valid config, got: %v", err)
	}
}

func TestChainIt_Execute_Success(t *testing.T) {
	helper := &ChainIt{}
	mockConn := &mockConnectorForChainIt{}
	ctx := context.Background()

	input := helpers.HelperInput{
		ContactID: "contact-123",
		Config: map[string]interface{}{
			"helpers": []interface{}{"tag_it", "copy_it", "email_it"},
		},
		Connector: mockConn,
	}

	output, err := helper.Execute(ctx, input)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if !output.Success {
		t.Error("Expected success to be true")
	}

	if output.Message != "Chained 3 helper(s) for sequential execution" {
		t.Errorf("Unexpected message: %s", output.Message)
	}
}

func TestChainIt_Execute_ActionsRecorded(t *testing.T) {
	helper := &ChainIt{}
	mockConn := &mockConnectorForChainIt{}
	ctx := context.Background()

	input := helpers.HelperInput{
		ContactID: "contact-123",
		Config: map[string]interface{}{
			"helpers": []interface{}{"helper1", "helper2", "helper3"},
		},
		Connector: mockConn,
	}

	output, err := helper.Execute(ctx, input)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(output.Actions) != 3 {
		t.Fatalf("Expected 3 actions, got %d", len(output.Actions))
	}

	for i, action := range output.Actions {
		if action.Type != "helper_chain" {
			t.Errorf("Expected action type 'helper_chain', got '%s'", action.Type)
		}
		expectedHelper := fmt.Sprintf("helper%d", i+1)
		if action.Target != expectedHelper {
			t.Errorf("Expected action target '%s', got '%s'", expectedHelper, action.Target)
		}
		if action.Value != i {
			t.Errorf("Expected action value %d, got '%v'", i, action.Value)
		}
	}
}

func TestChainIt_Execute_LogsRecorded(t *testing.T) {
	helper := &ChainIt{}
	mockConn := &mockConnectorForChainIt{}
	ctx := context.Background()

	input := helpers.HelperInput{
		ContactID: "contact-123",
		Config: map[string]interface{}{
			"helpers": []interface{}{"tag_it", "copy_it"},
		},
		Connector: mockConn,
	}

	output, err := helper.Execute(ctx, input)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(output.Logs) != 2 {
		t.Fatalf("Expected 2 logs, got %d", len(output.Logs))
	}
}

func TestChainIt_Execute_SingleHelper(t *testing.T) {
	helper := &ChainIt{}
	mockConn := &mockConnectorForChainIt{}
	ctx := context.Background()

	input := helpers.HelperInput{
		ContactID: "contact-123",
		Config: map[string]interface{}{
			"helpers": []interface{}{"single_helper"},
		},
		Connector: mockConn,
	}

	output, err := helper.Execute(ctx, input)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if !output.Success {
		t.Error("Expected success to be true")
	}

	if len(output.Actions) != 1 {
		t.Errorf("Expected 1 action, got %d", len(output.Actions))
	}
}

func TestChainIt_Execute_ManyHelpers(t *testing.T) {
	helper := &ChainIt{}
	mockConn := &mockConnectorForChainIt{}
	ctx := context.Background()

	input := helpers.HelperInput{
		ContactID: "contact-123",
		Config: map[string]interface{}{
			"helpers": []interface{}{"h1", "h2", "h3", "h4", "h5", "h6", "h7"},
		},
		Connector: mockConn,
	}

	output, err := helper.Execute(ctx, input)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(output.Actions) != 7 {
		t.Errorf("Expected 7 actions, got %d", len(output.Actions))
	}

	if output.Message != "Chained 7 helper(s) for sequential execution" {
		t.Errorf("Unexpected message: %s", output.Message)
	}
}

func TestChainIt_Execute_HelperOrder(t *testing.T) {
	helper := &ChainIt{}
	mockConn := &mockConnectorForChainIt{}
	ctx := context.Background()

	input := helpers.HelperInput{
		ContactID: "contact-123",
		Config: map[string]interface{}{
			"helpers": []interface{}{"first", "second", "third"},
		},
		Connector: mockConn,
	}

	output, err := helper.Execute(ctx, input)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Verify order is preserved
	if output.Actions[0].Target != "first" {
		t.Errorf("Expected first action target 'first', got '%s'", output.Actions[0].Target)
	}
	if output.Actions[1].Target != "second" {
		t.Errorf("Expected second action target 'second', got '%s'", output.Actions[1].Target)
	}
	if output.Actions[2].Target != "third" {
		t.Errorf("Expected third action target 'third', got '%s'", output.Actions[2].Target)
	}
}
