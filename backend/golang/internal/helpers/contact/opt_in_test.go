package contact

import (
	"context"
	"fmt"
	"testing"

	"github.com/myfusionhelper/api/internal/connectors"
	"github.com/myfusionhelper/api/internal/helpers"
)

// mockConnectorForOptIn implements connectors.CRMConnector for testing
type mockConnectorForOptIn struct {
	fieldValues     map[string]interface{}
	getFieldError   error
}

func (m *mockConnectorForOptIn) GetContactFieldValue(ctx context.Context, contactID, fieldKey string) (interface{}, error) {
	if m.getFieldError != nil {
		return nil, m.getFieldError
	}
	if m.fieldValues == nil {
		return nil, fmt.Errorf("field not found")
	}
	return m.fieldValues[fieldKey], nil
}

// Stub implementations
func (m *mockConnectorForOptIn) GetContacts(ctx context.Context, opts connectors.QueryOptions) (*connectors.ContactList, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForOptIn) GetContact(ctx context.Context, contactID string) (*connectors.NormalizedContact, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForOptIn) CreateContact(ctx context.Context, contact connectors.CreateContactInput) (*connectors.NormalizedContact, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForOptIn) UpdateContact(ctx context.Context, contactID string, updates connectors.UpdateContactInput) (*connectors.NormalizedContact, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForOptIn) DeleteContact(ctx context.Context, contactID string) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorForOptIn) GetTags(ctx context.Context) ([]connectors.Tag, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForOptIn) ApplyTag(ctx context.Context, contactID, tagID string) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorForOptIn) RemoveTag(ctx context.Context, contactID, tagID string) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorForOptIn) GetCustomFields(ctx context.Context) ([]connectors.CustomField, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForOptIn) SetContactFieldValue(ctx context.Context, contactID, fieldKey string, value interface{}) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorForOptIn) TriggerAutomation(ctx context.Context, contactID, automationID string) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorForOptIn) AchieveGoal(ctx context.Context, contactID, goalName, integration string) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorForOptIn) TestConnection(ctx context.Context) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorForOptIn) GetMetadata() connectors.ConnectorMetadata {
	return connectors.ConnectorMetadata{}
}
func (m *mockConnectorForOptIn) GetCapabilities() []connectors.Capability {
	return nil
}

func TestOptIn_GetMetadata(t *testing.T) {
	h := &OptIn{}
	if h.GetName() != "Opt In" {
		t.Errorf("expected name 'Opt In', got '%s'", h.GetName())
	}
	if h.GetType() != "opt_in" {
		t.Errorf("expected type 'opt_in', got '%s'", h.GetType())
	}
	if h.GetCategory() != "contact" {
		t.Errorf("expected category 'contact', got '%s'", h.GetCategory())
	}
	if !h.RequiresCRM() {
		t.Error("expected RequiresCRM to be true")
	}
	supported := h.SupportedCRMs()
	if len(supported) != 1 || supported[0] != "keap" {
		t.Errorf("expected SupportedCRMs ['keap'], got %v", supported)
	}
}

func TestOptIn_GetConfigSchema(t *testing.T) {
	h := &OptIn{}
	schema := h.GetConfigSchema()
	if schema["type"] != "object" {
		t.Errorf("expected type 'object', got '%v'", schema["type"])
	}
	props := schema["properties"].(map[string]interface{})
	if _, ok := props["email_field"]; !ok {
		t.Error("schema missing email_field property")
	}
	if _, ok := props["reason"]; !ok {
		t.Error("schema missing reason property")
	}

	required := schema["required"].([]string)
	if len(required) != 0 {
		t.Errorf("expected 0 required fields, got %d", len(required))
	}
}

func TestOptIn_ValidateConfig(t *testing.T) {
	h := &OptIn{}
	config := map[string]interface{}{}
	err := h.ValidateConfig(config)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestOptIn_Execute_BasicOptIn(t *testing.T) {
	h := &OptIn{}
	mock := &mockConnectorForOptIn{
		fieldValues: map[string]interface{}{
			"email": "test@example.com",
		},
	}

	input := helpers.HelperInput{
		ContactID: "12345",
		Config:    map[string]interface{}{},
		Connector: mock,
	}

	output, err := h.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !output.Success {
		t.Error("expected success=true")
	}
	if output.Message != "Opt-in prepared for test@example.com" {
		t.Errorf("unexpected message: %s", output.Message)
	}

	optInData := output.ModifiedData
	if optInData["email"] != "test@example.com" {
		t.Errorf("unexpected email: %v", optInData["email"])
	}
	if optInData["action"] != "opt_in" {
		t.Errorf("unexpected action: %v", optInData["action"])
	}
	if optInData["contact_id"] != "12345" {
		t.Errorf("unexpected contact_id: %v", optInData["contact_id"])
	}
}

func TestOptIn_Execute_WithCustomEmailField(t *testing.T) {
	h := &OptIn{}
	mock := &mockConnectorForOptIn{
		fieldValues: map[string]interface{}{
			"Email1": "custom@example.com",
		},
	}

	input := helpers.HelperInput{
		ContactID: "12345",
		Config: map[string]interface{}{
			"email_field": "Email1",
		},
		Connector: mock,
	}

	output, err := h.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !output.Success {
		t.Error("expected success=true")
	}

	optInData := output.ModifiedData
	if optInData["email"] != "custom@example.com" {
		t.Errorf("unexpected email: %v", optInData["email"])
	}
}

func TestOptIn_Execute_WithReason(t *testing.T) {
	h := &OptIn{}
	mock := &mockConnectorForOptIn{
		fieldValues: map[string]interface{}{
			"email": "test@example.com",
		},
	}

	input := helpers.HelperInput{
		ContactID: "12345",
		Config: map[string]interface{}{
			"reason": "Newsletter subscription",
		},
		Connector: mock,
	}

	output, err := h.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !output.Success {
		t.Error("expected success=true")
	}

	optInData := output.ModifiedData
	if optInData["reason"] != "Newsletter subscription" {
		t.Errorf("unexpected reason: %v", optInData["reason"])
	}
}

func TestOptIn_Execute_EmptyEmail(t *testing.T) {
	h := &OptIn{}
	mock := &mockConnectorForOptIn{
		fieldValues: map[string]interface{}{
			"email": "",
		},
	}

	input := helpers.HelperInput{
		ContactID: "12345",
		Config:    map[string]interface{}{},
		Connector: mock,
	}

	output, err := h.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if output.Success {
		t.Error("expected success=false for empty email")
	}
	if output.Message != "Email field 'email' is empty, cannot opt in" {
		t.Errorf("unexpected message: %s", output.Message)
	}
}

func TestOptIn_Execute_NilEmail(t *testing.T) {
	h := &OptIn{}
	mock := &mockConnectorForOptIn{
		fieldValues: map[string]interface{}{
			"email": nil,
		},
	}

	input := helpers.HelperInput{
		ContactID: "12345",
		Config:    map[string]interface{}{},
		Connector: mock,
	}

	output, err := h.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if output.Success {
		t.Error("expected success=false for nil email")
	}
}

func TestOptIn_Execute_GetFieldError(t *testing.T) {
	h := &OptIn{}
	mock := &mockConnectorForOptIn{
		getFieldError: fmt.Errorf("CRM API error"),
	}

	input := helpers.HelperInput{
		ContactID: "12345",
		Config:    map[string]interface{}{},
		Connector: mock,
	}

	output, err := h.Execute(context.Background(), input)
	if err == nil {
		t.Error("expected error from GetContactFieldValue")
	}
	if output.Success {
		t.Error("expected success=false on error")
	}
	if output.Message == "" {
		t.Error("expected error message to be set")
	}
}

func TestOptIn_Execute_ActionVerification(t *testing.T) {
	h := &OptIn{}
	mock := &mockConnectorForOptIn{
		fieldValues: map[string]interface{}{
			"email": "action@example.com",
		},
	}

	input := helpers.HelperInput{
		ContactID: "12345",
		Config:    map[string]interface{}{},
		Connector: mock,
	}

	output, err := h.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(output.Actions) != 1 {
		t.Errorf("expected 1 action, got %d", len(output.Actions))
	}
	if output.Actions[0].Type != "opt_in_requested" {
		t.Errorf("expected action type 'opt_in_requested', got '%s'", output.Actions[0].Type)
	}
	if output.Actions[0].Target != "12345" {
		t.Errorf("expected action target '12345', got '%s'", output.Actions[0].Target)
	}

	if len(output.Logs) == 0 {
		t.Error("expected logs to be populated")
	}
}

func TestOptIn_Execute_DefaultEmailField(t *testing.T) {
	h := &OptIn{}
	mock := &mockConnectorForOptIn{
		fieldValues: map[string]interface{}{
			"email": "default@example.com",
		},
	}

	input := helpers.HelperInput{
		ContactID: "12345",
		Config: map[string]interface{}{
			"email_field": "", // Empty should use default
		},
		Connector: mock,
	}

	output, err := h.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !output.Success {
		t.Error("expected success=true with default email field")
	}

	optInData := output.ModifiedData
	if optInData["email"] != "default@example.com" {
		t.Errorf("expected default email field to be used")
	}
}

func TestOptIn_Execute_EmptyReason(t *testing.T) {
	h := &OptIn{}
	mock := &mockConnectorForOptIn{
		fieldValues: map[string]interface{}{
			"email": "test@example.com",
		},
	}

	input := helpers.HelperInput{
		ContactID: "12345",
		Config: map[string]interface{}{
			"reason": "",
		},
		Connector: mock,
	}

	output, err := h.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !output.Success {
		t.Error("expected success=true")
	}

	optInData := output.ModifiedData
	if optInData["reason"] != "" {
		t.Errorf("expected empty reason, got %v", optInData["reason"])
	}
}
