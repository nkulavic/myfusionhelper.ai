package automation

import (
	"context"
	"errors"
	"testing"

	"github.com/myfusionhelper/api/internal/connectors"
	"github.com/myfusionhelper/api/internal/helpers"
)

// mockConnectorSimpleOptIn is a minimal mock for testing simple_opt_in
type mockConnectorSimpleOptIn struct {
	setOptInStatusFunc func(ctx context.Context, contactID string, optIn bool, reason string) error
}

func (m *mockConnectorSimpleOptIn) SetOptInStatus(ctx context.Context, contactID string, optIn bool, reason string) error {
	if m.setOptInStatusFunc != nil {
		return m.setOptInStatusFunc(ctx, contactID, optIn, reason)
	}
	return nil
}

// Stub methods to satisfy CRMConnector interface
func (m *mockConnectorSimpleOptIn) GetContacts(ctx context.Context, opts connectors.QueryOptions) (*connectors.ContactList, error) {
	return nil, nil
}
func (m *mockConnectorSimpleOptIn) GetContact(ctx context.Context, contactID string) (*connectors.NormalizedContact, error) {
	return nil, nil
}
func (m *mockConnectorSimpleOptIn) CreateContact(ctx context.Context, contact connectors.CreateContactInput) (*connectors.NormalizedContact, error) {
	return nil, nil
}
func (m *mockConnectorSimpleOptIn) UpdateContact(ctx context.Context, contactID string, updates connectors.UpdateContactInput) (*connectors.NormalizedContact, error) {
	return nil, nil
}
func (m *mockConnectorSimpleOptIn) DeleteContact(ctx context.Context, contactID string) error {
	return nil
}
func (m *mockConnectorSimpleOptIn) GetTags(ctx context.Context) ([]connectors.Tag, error) {
	return nil, nil
}
func (m *mockConnectorSimpleOptIn) ApplyTag(ctx context.Context, contactID string, tagID string) error {
	return nil
}
func (m *mockConnectorSimpleOptIn) RemoveTag(ctx context.Context, contactID string, tagID string) error {
	return nil
}
func (m *mockConnectorSimpleOptIn) GetCustomFields(ctx context.Context) ([]connectors.CustomField, error) {
	return nil, nil
}
func (m *mockConnectorSimpleOptIn) GetContactFieldValue(ctx context.Context, contactID string, fieldKey string) (interface{}, error) {
	return nil, nil
}
func (m *mockConnectorSimpleOptIn) SetContactFieldValue(ctx context.Context, contactID string, fieldKey string, value interface{}) error {
	return nil
}
func (m *mockConnectorSimpleOptIn) TriggerAutomation(ctx context.Context, contactID string, automationID string) error {
	return nil
}
func (m *mockConnectorSimpleOptIn) AchieveGoal(ctx context.Context, contactID string, goalName string, integration string) error {
	return nil
}
func (m *mockConnectorSimpleOptIn) TestConnection(ctx context.Context) error {
	return nil
}
func (m *mockConnectorSimpleOptIn) GetMetadata() connectors.ConnectorMetadata {
	return connectors.ConnectorMetadata{}
}
func (m *mockConnectorSimpleOptIn) GetCapabilities() []connectors.Capability {
	return nil
}

func TestSimpleOptIn_GetMetadata(t *testing.T) {
	helper := &SimpleOptIn{}

	if helper.GetName() != "Simple Opt-In" {
		t.Errorf("Expected name 'Simple Opt-In', got '%s'", helper.GetName())
	}

	if helper.GetType() != "simple_opt_in" {
		t.Errorf("Expected type 'simple_opt_in', got '%s'", helper.GetType())
	}

	if helper.GetCategory() != "automation" {
		t.Errorf("Expected category 'automation', got '%s'", helper.GetCategory())
	}

	if helper.GetDescription() != "Set marketable status for email opt-in" {
		t.Errorf("Expected description 'Set marketable status for email opt-in', got '%s'", helper.GetDescription())
	}

	if !helper.RequiresCRM() {
		t.Error("Expected RequiresCRM to be true")
	}

	// simple_opt_in supports all CRMs
	if helper.SupportedCRMs() != nil {
		t.Errorf("Expected SupportedCRMs to be nil (all CRMs), got %v", helper.SupportedCRMs())
	}
}

func TestSimpleOptIn_GetConfigSchema(t *testing.T) {
	helper := &SimpleOptIn{}
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

	// Check opt_in property
	optInProp, ok := props["opt_in"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected opt_in property in schema")
	}

	if optInProp["type"] != "boolean" {
		t.Error("Expected opt_in type to be 'boolean'")
	}

	// Check reason property (optional)
	reasonProp, ok := props["reason"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected reason property in schema")
	}

	if reasonProp["type"] != "string" {
		t.Error("Expected reason type to be 'string'")
	}

	// Check required fields
	required, ok := schema["required"].([]string)
	if !ok {
		t.Fatal("Expected required fields in schema")
	}

	if len(required) != 1 || required[0] != "opt_in" {
		t.Errorf("Expected required=['opt_in'], got %v", required)
	}
}

func TestSimpleOptIn_ValidateConfig(t *testing.T) {
	helper := &SimpleOptIn{}

	tests := []struct {
		name      string
		config    map[string]interface{}
		expectErr bool
	}{
		{
			name:      "valid opt-in",
			config:    map[string]interface{}{"opt_in": true},
			expectErr: false,
		},
		{
			name:      "valid opt-out",
			config:    map[string]interface{}{"opt_in": false},
			expectErr: false,
		},
		{
			name:      "valid with reason",
			config:    map[string]interface{}{"opt_in": true, "reason": "User requested"},
			expectErr: false,
		},
		{
			name:      "missing opt_in",
			config:    map[string]interface{}{},
			expectErr: true,
		},
		{
			name:      "invalid opt_in type (string)",
			config:    map[string]interface{}{"opt_in": "true"},
			expectErr: true,
		},
		{
			name:      "invalid opt_in type (number)",
			config:    map[string]interface{}{"opt_in": 1},
			expectErr: true,
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

func TestSimpleOptIn_Execute_OptIn(t *testing.T) {
	helper := &SimpleOptIn{}

	var capturedContactID string
	var capturedOptIn bool
	var capturedReason string

	connector := &mockConnectorSimpleOptIn{
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
			"opt_in": true,
			"reason": "User signed up",
		},
	}

	output, err := helper.Execute(context.Background(), input)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if !output.Success {
		t.Error("Expected success to be true")
	}

	if output.Message != "Contact opted in for email marketing" {
		t.Errorf("Expected message 'Contact opted in for email marketing', got '%s'", output.Message)
	}

	if capturedContactID != "contact_123" {
		t.Errorf("Expected contact ID 'contact_123', got '%s'", capturedContactID)
	}

	if !capturedOptIn {
		t.Error("Expected opt-in to be true")
	}

	if capturedReason != "User signed up" {
		t.Errorf("Expected reason 'User signed up', got '%s'", capturedReason)
	}

	if len(output.Actions) != 1 {
		t.Fatalf("Expected 1 action, got %d", len(output.Actions))
	}

	action := output.Actions[0]
	if action.Type != "opt_in_status_changed" {
		t.Errorf("Expected action type 'opt_in_status_changed', got '%s'", action.Type)
	}

	if action.Target != "contact_123" {
		t.Errorf("Expected action target 'contact_123', got '%s'", action.Target)
	}

	if action.Value != "opted in" {
		t.Errorf("Expected action value 'opted in', got '%v'", action.Value)
	}

	if len(output.Logs) == 0 {
		t.Error("Expected logs to be populated")
	}
}

func TestSimpleOptIn_Execute_OptOut(t *testing.T) {
	helper := &SimpleOptIn{}

	var capturedOptIn bool

	connector := &mockConnectorSimpleOptIn{
		setOptInStatusFunc: func(ctx context.Context, contactID string, optIn bool, reason string) error {
			capturedOptIn = optIn
			return nil
		},
	}

	input := helpers.HelperInput{
		Connector: connector,
		ContactID: "contact_456",
		Config: map[string]interface{}{
			"opt_in": false,
		},
	}

	output, err := helper.Execute(context.Background(), input)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if !output.Success {
		t.Error("Expected success to be true")
	}

	if output.Message != "Contact opted out for email marketing" {
		t.Errorf("Expected message 'Contact opted out for email marketing', got '%s'", output.Message)
	}

	if capturedOptIn {
		t.Error("Expected opt-in to be false")
	}

	if len(output.Actions) != 1 {
		t.Fatalf("Expected 1 action, got %d", len(output.Actions))
	}

	action := output.Actions[0]
	if action.Value != "opted out" {
		t.Errorf("Expected action value 'opted out', got '%v'", action.Value)
	}
}

func TestSimpleOptIn_Execute_NoReason(t *testing.T) {
	helper := &SimpleOptIn{}

	var capturedReason string

	connector := &mockConnectorSimpleOptIn{
		setOptInStatusFunc: func(ctx context.Context, contactID string, optIn bool, reason string) error {
			capturedReason = reason
			return nil
		},
	}

	input := helpers.HelperInput{
		Connector: connector,
		ContactID: "contact_789",
		Config: map[string]interface{}{
			"opt_in": true,
		},
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

func TestSimpleOptIn_Execute_ConnectorError(t *testing.T) {
	helper := &SimpleOptIn{}

	connector := &mockConnectorSimpleOptIn{
		setOptInStatusFunc: func(ctx context.Context, contactID string, optIn bool, reason string) error {
			return errors.New("CRM API error")
		},
	}

	input := helpers.HelperInput{
		Connector: connector,
		ContactID: "contact_error",
		Config: map[string]interface{}{
			"opt_in": true,
		},
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

func TestSimpleOptIn_Execute_DurationTracking(t *testing.T) {
	helper := &SimpleOptIn{}

	connector := &mockConnectorSimpleOptIn{
		setOptInStatusFunc: func(ctx context.Context, contactID string, optIn bool, reason string) error {
			return nil
		},
	}

	input := helpers.HelperInput{
		Connector: connector,
		ContactID: "contact_duration",
		Config: map[string]interface{}{
			"opt_in": true,
		},
	}

	output, err := helper.Execute(context.Background(), input)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Duration tracking happens in the executor, not the helper itself
	// This test just verifies the helper executes without panicking
	if output == nil {
		t.Fatal("Expected output, got nil")
	}
}
