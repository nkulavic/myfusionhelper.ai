package contact

import (
	"context"
	"fmt"
	"testing"

	"github.com/myfusionhelper/api/internal/connectors"
	"github.com/myfusionhelper/api/internal/helpers"
)

// mockConnectorForOptOut implements connectors.CRMConnector for testing
type mockConnectorForOptOut struct {
	fieldValues   map[string]interface{}
	getFieldError error
}

func (m *mockConnectorForOptOut) GetContactFieldValue(ctx context.Context, contactID, fieldKey string) (interface{}, error) {
	if m.getFieldError != nil {
		return nil, m.getFieldError
	}
	if m.fieldValues == nil {
		return nil, fmt.Errorf("field not found")
	}
	return m.fieldValues[fieldKey], nil
}

// Stub implementations
func (m *mockConnectorForOptOut) GetContacts(ctx context.Context, opts connectors.QueryOptions) (*connectors.ContactList, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForOptOut) GetContact(ctx context.Context, contactID string) (*connectors.NormalizedContact, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForOptOut) CreateContact(ctx context.Context, contact connectors.CreateContactInput) (*connectors.NormalizedContact, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForOptOut) UpdateContact(ctx context.Context, contactID string, updates connectors.UpdateContactInput) (*connectors.NormalizedContact, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForOptOut) DeleteContact(ctx context.Context, contactID string) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorForOptOut) GetTags(ctx context.Context) ([]connectors.Tag, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForOptOut) ApplyTag(ctx context.Context, contactID, tagID string) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorForOptOut) RemoveTag(ctx context.Context, contactID, tagID string) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorForOptOut) GetCustomFields(ctx context.Context) ([]connectors.CustomField, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForOptOut) SetContactFieldValue(ctx context.Context, contactID, fieldKey string, value interface{}) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorForOptOut) TriggerAutomation(ctx context.Context, contactID, automationID string) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorForOptOut) AchieveGoal(ctx context.Context, contactID, goalName, integration string) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorForOptOut) TestConnection(ctx context.Context) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorForOptOut) GetMetadata() connectors.ConnectorMetadata {
	return connectors.ConnectorMetadata{}
}
func (m *mockConnectorForOptOut) GetCapabilities() []connectors.Capability {
	return nil
}

func TestOptOut_GetMetadata(t *testing.T) {
	h := &OptOut{}
	if h.GetName() != "Opt Out" {
		t.Errorf("expected name 'Opt Out', got '%s'", h.GetName())
	}
	if h.GetType() != "opt_out" {
		t.Errorf("expected type 'opt_out', got '%s'", h.GetType())
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

func TestOptOut_GetConfigSchema(t *testing.T) {
	h := &OptOut{}
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

func TestOptOut_ValidateConfig(t *testing.T) {
	h := &OptOut{}
	config := map[string]interface{}{}
	err := h.ValidateConfig(config)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestOptOut_Execute_BasicOptOut(t *testing.T) {
	h := &OptOut{}
	mock := &mockConnectorForOptOut{
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
	if output.Message != "Opt-out prepared for test@example.com" {
		t.Errorf("unexpected message: %s", output.Message)
	}

	optOutData := output.ModifiedData
	if optOutData["email"] != "test@example.com" {
		t.Errorf("unexpected email: %v", optOutData["email"])
	}
	if optOutData["action"] != "opt_out" {
		t.Errorf("unexpected action: %v", optOutData["action"])
	}
	if optOutData["contact_id"] != "12345" {
		t.Errorf("unexpected contact_id: %v", optOutData["contact_id"])
	}
}

func TestOptOut_Execute_WithCustomEmailField(t *testing.T) {
	h := &OptOut{}
	mock := &mockConnectorForOptOut{
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

	optOutData := output.ModifiedData
	if optOutData["email"] != "custom@example.com" {
		t.Errorf("unexpected email: %v", optOutData["email"])
	}
}

func TestOptOut_Execute_WithReason(t *testing.T) {
	h := &OptOut{}
	mock := &mockConnectorForOptOut{
		fieldValues: map[string]interface{}{
			"email": "test@example.com",
		},
	}

	input := helpers.HelperInput{
		ContactID: "12345",
		Config: map[string]interface{}{
			"reason": "User requested unsubscribe",
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

	optOutData := output.ModifiedData
	if optOutData["reason"] != "User requested unsubscribe" {
		t.Errorf("unexpected reason: %v", optOutData["reason"])
	}
}

func TestOptOut_Execute_EmptyEmail(t *testing.T) {
	h := &OptOut{}
	mock := &mockConnectorForOptOut{
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
	if output.Message != "Email field 'email' is empty, cannot opt out" {
		t.Errorf("unexpected message: %s", output.Message)
	}
}

func TestOptOut_Execute_NilEmail(t *testing.T) {
	h := &OptOut{}
	mock := &mockConnectorForOptOut{
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

func TestOptOut_Execute_GetFieldError(t *testing.T) {
	h := &OptOut{}
	mock := &mockConnectorForOptOut{
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

func TestOptOut_Execute_ActionVerification(t *testing.T) {
	h := &OptOut{}
	mock := &mockConnectorForOptOut{
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
	if output.Actions[0].Type != "opt_out_requested" {
		t.Errorf("expected action type 'opt_out_requested', got '%s'", output.Actions[0].Type)
	}
	if output.Actions[0].Target != "12345" {
		t.Errorf("expected action target '12345', got '%s'", output.Actions[0].Target)
	}

	if len(output.Logs) == 0 {
		t.Error("expected logs to be populated")
	}
}

func TestOptOut_Execute_DefaultEmailField(t *testing.T) {
	h := &OptOut{}
	mock := &mockConnectorForOptOut{
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

	optOutData := output.ModifiedData
	if optOutData["email"] != "default@example.com" {
		t.Errorf("expected default email field to be used")
	}
}

func TestOptOut_Execute_EmptyReason(t *testing.T) {
	h := &OptOut{}
	mock := &mockConnectorForOptOut{
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

	optOutData := output.ModifiedData
	if optOutData["reason"] != "" {
		t.Errorf("expected empty reason, got %v", optOutData["reason"])
	}
}
