package automation

import (
	"context"
	"errors"
	"testing"

	"github.com/myfusionhelper/api/internal/connectors"
	"github.com/myfusionhelper/api/internal/helpers"
)

// mockConnectorSimpleOptOut is a minimal mock for testing simple_opt_out
type mockConnectorSimpleOptOut struct {
	setOptInStatusFunc func(ctx context.Context, contactID string, optIn bool, reason string) error
}

func (m *mockConnectorSimpleOptOut) SetOptInStatus(ctx context.Context, contactID string, optIn bool, reason string) error {
	if m.setOptInStatusFunc != nil {
		return m.setOptInStatusFunc(ctx, contactID, optIn, reason)
	}
	return nil
}

// Stub methods to satisfy CRMConnector interface
func (m *mockConnectorSimpleOptOut) GetContacts(ctx context.Context, opts connectors.QueryOptions) (*connectors.ContactList, error) {
	return nil, nil
}
func (m *mockConnectorSimpleOptOut) GetContact(ctx context.Context, contactID string) (*connectors.NormalizedContact, error) {
	return nil, nil
}
func (m *mockConnectorSimpleOptOut) CreateContact(ctx context.Context, contact connectors.CreateContactInput) (*connectors.NormalizedContact, error) {
	return nil, nil
}
func (m *mockConnectorSimpleOptOut) UpdateContact(ctx context.Context, contactID string, updates connectors.UpdateContactInput) (*connectors.NormalizedContact, error) {
	return nil, nil
}
func (m *mockConnectorSimpleOptOut) DeleteContact(ctx context.Context, contactID string) error {
	return nil
}
func (m *mockConnectorSimpleOptOut) GetTags(ctx context.Context) ([]connectors.Tag, error) {
	return nil, nil
}
func (m *mockConnectorSimpleOptOut) ApplyTag(ctx context.Context, contactID string, tagID string) error {
	return nil
}
func (m *mockConnectorSimpleOptOut) RemoveTag(ctx context.Context, contactID string, tagID string) error {
	return nil
}
func (m *mockConnectorSimpleOptOut) GetCustomFields(ctx context.Context) ([]connectors.CustomField, error) {
	return nil, nil
}
func (m *mockConnectorSimpleOptOut) GetContactFieldValue(ctx context.Context, contactID string, fieldKey string) (interface{}, error) {
	return nil, nil
}
func (m *mockConnectorSimpleOptOut) SetContactFieldValue(ctx context.Context, contactID string, fieldKey string, value interface{}) error {
	return nil
}
func (m *mockConnectorSimpleOptOut) TriggerAutomation(ctx context.Context, contactID string, automationID string) error {
	return nil
}
func (m *mockConnectorSimpleOptOut) AchieveGoal(ctx context.Context, contactID string, goalName string, integration string) error {
	return nil
}
func (m *mockConnectorSimpleOptOut) TestConnection(ctx context.Context) error {
	return nil
}
func (m *mockConnectorSimpleOptOut) GetMetadata() connectors.ConnectorMetadata {
	return connectors.ConnectorMetadata{}
}
func (m *mockConnectorSimpleOptOut) GetCapabilities() []connectors.Capability {
	return nil
}

func TestSimpleOptOut_GetMetadata(t *testing.T) {
	helper := &SimpleOptOut{}

	if helper.GetName() != "Simple Opt-Out" {
		t.Errorf("Expected name 'Simple Opt-Out', got '%s'", helper.GetName())
	}

	if helper.GetType() != "simple_opt_out" {
		t.Errorf("Expected type 'simple_opt_out', got '%s'", helper.GetType())
	}

	if helper.GetCategory() != "automation" {
		t.Errorf("Expected category 'automation', got '%s'", helper.GetCategory())
	}

	if helper.GetDescription() != "Remove marketable status for email opt-out" {
		t.Errorf("Expected description 'Remove marketable status for email opt-out', got '%s'", helper.GetDescription())
	}

	if !helper.RequiresCRM() {
		t.Error("Expected RequiresCRM to be true")
	}

	// simple_opt_out supports all CRMs
	if helper.SupportedCRMs() != nil {
		t.Errorf("Expected SupportedCRMs to be nil (all CRMs), got %v", helper.SupportedCRMs())
	}
}

func TestSimpleOptOut_GetConfigSchema(t *testing.T) {
	helper := &SimpleOptOut{}
	schema := helper.GetConfigSchema()

	if schema == nil {
		t.Fatal("Expected config schema, got nil")
	}

	schemaType, ok := schema["type"].(string)
	if !ok || schemaType != "object" {
		t.Error("Expected schema type 'object'")
	}

	props, ok := schema["properties"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected properties in schema")
	}

	// Check reason property (optional)
	reasonProp, ok := props["reason"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected reason property in schema")
	}

	if reasonProp["type"] != "string" {
		t.Error("Expected reason type to be 'string'")
	}

	// No required fields for opt-out
	if required, exists := schema["required"]; exists {
		if reqArr, ok := required.([]string); ok && len(reqArr) > 0 {
			t.Errorf("Expected no required fields, got %v", reqArr)
		}
	}
}

func TestSimpleOptOut_ValidateConfig(t *testing.T) {
	helper := &SimpleOptOut{}

	tests := []struct {
		name      string
		config    map[string]interface{}
		expectErr bool
	}{
		{
			name:      "empty config",
			config:    map[string]interface{}{},
			expectErr: false,
		},
		{
			name:      "with reason",
			config:    map[string]interface{}{"reason": "User requested"},
			expectErr: false,
		},
		{
			name:      "nil config",
			config:    nil,
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := helper.ValidateConfig(tt.config)
			if tt.expectErr && err == nil {
				t.Error("Expected error, got nil")
			}
			if !tt.expectErr && err != nil {
				t.Errorf("Expected no error, got %v", err)
			}
		})
	}
}

func TestSimpleOptOut_Execute_Success(t *testing.T) {
	helper := &SimpleOptOut{}

	var capturedContactID string
	var capturedOptIn bool
	var capturedReason string

	connector := &mockConnectorSimpleOptOut{
		setOptInStatusFunc: func(ctx context.Context, contactID string, optIn bool, reason string) error {
			capturedContactID = contactID
			capturedOptIn = optIn
			capturedReason = reason
			return nil
		},
	}

	input := helpers.HelperInput{
		Connector: connector,
		ContactID: "contact_123",
		Config: map[string]interface{}{
			"reason": "User unsubscribed",
		},
	}

	output, err := helper.Execute(context.Background(), input)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if !output.Success {
		t.Error("Expected success to be true")
	}

	if output.Message != "Contact opted out from email marketing" {
		t.Errorf("Expected message 'Contact opted out from email marketing', got '%s'", output.Message)
	}

	if capturedContactID != "contact_123" {
		t.Errorf("Expected contact ID 'contact_123', got '%s'", capturedContactID)
	}

	// Simple opt-out always sets optIn to false
	if capturedOptIn {
		t.Error("Expected opt-in to be false for opt-out")
	}

	if capturedReason != "User unsubscribed" {
		t.Errorf("Expected reason 'User unsubscribed', got '%s'", capturedReason)
	}

	if len(output.Actions) != 1 {
		t.Fatalf("Expected 1 action, got %d", len(output.Actions))
	}

	action := output.Actions[0]
	if action.Type != "opt_out" {
		t.Errorf("Expected action type 'opt_out', got '%s'", action.Type)
	}

	if action.Target != "contact_123" {
		t.Errorf("Expected action target 'contact_123', got '%s'", action.Target)
	}

	if action.Value != "opted_out" {
		t.Errorf("Expected action value 'opted_out', got '%v'", action.Value)
	}

	if len(output.Logs) == 0 {
		t.Error("Expected logs to be populated")
	}
}

func TestSimpleOptOut_Execute_NoReason(t *testing.T) {
	helper := &SimpleOptOut{}

	var capturedReason string

	connector := &mockConnectorSimpleOptOut{
		setOptInStatusFunc: func(ctx context.Context, contactID string, optIn bool, reason string) error {
			capturedReason = reason
			return nil
		},
	}

	input := helpers.HelperInput{
		Connector: connector,
		ContactID: "contact_456",
		Config:    map[string]interface{}{}, // No reason provided
	}

	output, err := helper.Execute(context.Background(), input)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if !output.Success {
		t.Error("Expected success to be true")
	}

	if capturedReason != "" {
		t.Errorf("Expected reason to be empty, got '%s'", capturedReason)
	}
}

func TestSimpleOptOut_Execute_ConnectorError(t *testing.T) {
	helper := &SimpleOptOut{}

	connector := &mockConnectorSimpleOptOut{
		setOptInStatusFunc: func(ctx context.Context, contactID string, optIn bool, reason string) error {
			return errors.New("CRM API error")
		},
	}

	input := helpers.HelperInput{
		Connector: connector,
		ContactID: "contact_error",
		Config:    map[string]interface{}{},
	}

	output, err := helper.Execute(context.Background(), input)

	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	if output.Success {
		t.Error("Expected success to be false")
	}

	if output.Message == "" {
		t.Error("Expected error message in output")
	}
}
