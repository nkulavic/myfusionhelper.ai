package contact

import (
	"context"
	"fmt"
	"testing"

	"github.com/myfusionhelper/api/internal/connectors"
	"github.com/myfusionhelper/api/internal/helpers"
)

// mockConnectorForCombineIt implements the CRMConnector interface for testing combine_it
type mockConnectorForCombineIt struct {
	fieldsSet     map[string]interface{}
	fieldValues   map[string]interface{}
	setFieldError error
	getFieldError error
}

func (m *mockConnectorForCombineIt) SetContactFieldValue(ctx context.Context, contactID, fieldKey string, value interface{}) error {
	if m.setFieldError != nil {
		return m.setFieldError
	}
	if m.fieldsSet == nil {
		m.fieldsSet = make(map[string]interface{})
	}
	m.fieldsSet[fieldKey] = value
	return nil
}

func (m *mockConnectorForCombineIt) GetContactFieldValue(ctx context.Context, contactID, fieldKey string) (interface{}, error) {
	if m.getFieldError != nil {
		return nil, m.getFieldError
	}
	if m.fieldValues == nil {
		return nil, nil
	}
	return m.fieldValues[fieldKey], nil
}

// Stub implementations with correct signatures
func (m *mockConnectorForCombineIt) GetContacts(ctx context.Context, opts connectors.QueryOptions) (*connectors.ContactList, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *mockConnectorForCombineIt) GetContact(ctx context.Context, contactID string) (*connectors.NormalizedContact, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *mockConnectorForCombineIt) CreateContact(ctx context.Context, input connectors.CreateContactInput) (*connectors.NormalizedContact, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *mockConnectorForCombineIt) UpdateContact(ctx context.Context, contactID string, updates connectors.UpdateContactInput) (*connectors.NormalizedContact, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *mockConnectorForCombineIt) DeleteContact(ctx context.Context, contactID string) error {
	return fmt.Errorf("not implemented")
}

func (m *mockConnectorForCombineIt) GetTags(ctx context.Context) ([]connectors.Tag, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *mockConnectorForCombineIt) ApplyTag(ctx context.Context, contactID string, tagID string) error {
	return fmt.Errorf("not implemented")
}

func (m *mockConnectorForCombineIt) RemoveTag(ctx context.Context, contactID string, tagID string) error {
	return fmt.Errorf("not implemented")
}

func (m *mockConnectorForCombineIt) GetCustomFields(ctx context.Context) ([]connectors.CustomField, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *mockConnectorForCombineIt) TriggerAutomation(ctx context.Context, contactID string, automationID string) error {
	return fmt.Errorf("not implemented")
}

func (m *mockConnectorForCombineIt) AchieveGoal(ctx context.Context, contactID string, goalName string, integration string) error {
	return fmt.Errorf("not implemented")
}

func (m *mockConnectorForCombineIt) TestConnection(ctx context.Context) error {
	return fmt.Errorf("not implemented")
}

func (m *mockConnectorForCombineIt) GetMetadata() connectors.ConnectorMetadata {
	return connectors.ConnectorMetadata{
		PlatformSlug: "test",
		PlatformName: "Test Platform",
	}
}

func (m *mockConnectorForCombineIt) GetCapabilities() []connectors.Capability {
	return []connectors.Capability{connectors.CapContacts, connectors.CapCustomFields}
}

// Tests

func TestCombineIt_GetMetadata(t *testing.T) {
	helper := &CombineIt{}

	if helper.GetName() != "Combine It" {
		t.Errorf("Expected name 'Combine It', got '%s'", helper.GetName())
	}
	if helper.GetType() != "combine_it" {
		t.Errorf("Expected type 'combine_it', got '%s'", helper.GetType())
	}
	if helper.GetCategory() != "contact" {
		t.Errorf("Expected category 'contact', got '%s'", helper.GetCategory())
	}
	if !helper.RequiresCRM() {
		t.Error("Expected RequiresCRM to be true")
	}
}

func TestCombineIt_GetConfigSchema(t *testing.T) {
	helper := &CombineIt{}
	schema := helper.GetConfigSchema()

	if schema == nil {
		t.Fatal("Schema should not be nil")
	}

	props, ok := schema["properties"].(map[string]interface{})
	if !ok {
		t.Fatal("Schema should have properties")
	}

	if _, ok := props["source_fields"]; !ok {
		t.Error("Schema should have source_fields property")
	}
	if _, ok := props["target_field"]; !ok {
		t.Error("Schema should have target_field property")
	}
	if _, ok := props["separator"]; !ok {
		t.Error("Schema should have separator property")
	}
	if _, ok := props["skip_empty"]; !ok {
		t.Error("Schema should have skip_empty property")
	}
}

func TestCombineIt_ValidateConfig_MissingSourceFields(t *testing.T) {
	helper := &CombineIt{}

	config := map[string]interface{}{
		"target_field": "full_name",
	}
	err := helper.ValidateConfig(config)
	if err == nil {
		t.Error("Expected error for missing source_fields")
	}
}

func TestCombineIt_ValidateConfig_EmptySourceFields(t *testing.T) {
	helper := &CombineIt{}

	config := map[string]interface{}{
		"source_fields": []interface{}{},
		"target_field":  "full_name",
	}
	err := helper.ValidateConfig(config)
	if err == nil {
		t.Error("Expected error for empty source_fields")
	}
}

func TestCombineIt_ValidateConfig_InvalidSourceFieldsType(t *testing.T) {
	helper := &CombineIt{}

	config := map[string]interface{}{
		"source_fields": "not-an-array",
		"target_field":  "full_name",
	}
	err := helper.ValidateConfig(config)
	if err == nil {
		t.Error("Expected error for non-array source_fields")
	}
}

func TestCombineIt_ValidateConfig_MissingTargetField(t *testing.T) {
	helper := &CombineIt{}

	config := map[string]interface{}{
		"source_fields": []interface{}{"first_name", "last_name"},
	}
	err := helper.ValidateConfig(config)
	if err == nil {
		t.Error("Expected error for missing target_field")
	}
}

func TestCombineIt_ValidateConfig_EmptyTargetField(t *testing.T) {
	helper := &CombineIt{}

	config := map[string]interface{}{
		"source_fields": []interface{}{"first_name", "last_name"},
		"target_field":  "",
	}
	err := helper.ValidateConfig(config)
	if err == nil {
		t.Error("Expected error for empty target_field")
	}
}

func TestCombineIt_ValidateConfig_Valid(t *testing.T) {
	helper := &CombineIt{}

	config := map[string]interface{}{
		"source_fields": []interface{}{"first_name", "last_name"},
		"target_field":  "full_name",
	}
	err := helper.ValidateConfig(config)
	if err != nil {
		t.Errorf("Expected no error for valid config, got: %v", err)
	}
}

func TestCombineIt_Execute_Success(t *testing.T) {
	helper := &CombineIt{}
	mockConn := &mockConnectorForCombineIt{
		fieldValues: map[string]interface{}{
			"first_name": "John",
			"last_name":  "Doe",
		},
	}
	ctx := context.Background()

	input := helpers.HelperInput{
		ContactID: "contact-123",
		Config: map[string]interface{}{
			"source_fields": []interface{}{"first_name", "last_name"},
			"target_field":  "full_name",
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

	if mockConn.fieldsSet["full_name"] != "John Doe" {
		t.Errorf("Expected full_name to be 'John Doe', got: %v", mockConn.fieldsSet["full_name"])
	}
}

func TestCombineIt_Execute_CustomSeparator(t *testing.T) {
	helper := &CombineIt{}
	mockConn := &mockConnectorForCombineIt{
		fieldValues: map[string]interface{}{
			"first_name": "John",
			"last_name":  "Doe",
		},
	}
	ctx := context.Background()

	input := helpers.HelperInput{
		ContactID: "contact-123",
		Config: map[string]interface{}{
			"source_fields": []interface{}{"first_name", "last_name"},
			"target_field":  "full_name",
			"separator":     ", ",
		},
		Connector: mockConn,
	}

	_, err := helper.Execute(ctx, input)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if mockConn.fieldsSet["full_name"] != "John, Doe" {
		t.Errorf("Expected full_name to be 'John, Doe', got: %v", mockConn.fieldsSet["full_name"])
	}
}

func TestCombineIt_Execute_SkipEmpty(t *testing.T) {
	helper := &CombineIt{}
	mockConn := &mockConnectorForCombineIt{
		fieldValues: map[string]interface{}{
			"first_name":  "John",
			"middle_name": "",
			"last_name":   "Doe",
		},
	}
	ctx := context.Background()

	input := helpers.HelperInput{
		ContactID: "contact-123",
		Config: map[string]interface{}{
			"source_fields": []interface{}{"first_name", "middle_name", "last_name"},
			"target_field":  "full_name",
			"skip_empty":    true,
		},
		Connector: mockConn,
	}

	_, err := helper.Execute(ctx, input)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if mockConn.fieldsSet["full_name"] != "John Doe" {
		t.Errorf("Expected full_name to be 'John Doe', got: %v", mockConn.fieldsSet["full_name"])
	}
}

func TestCombineIt_Execute_DontSkipEmpty(t *testing.T) {
	helper := &CombineIt{}
	mockConn := &mockConnectorForCombineIt{
		fieldValues: map[string]interface{}{
			"first_name":  "John",
			"middle_name": "",
			"last_name":   "Doe",
		},
	}
	ctx := context.Background()

	input := helpers.HelperInput{
		ContactID: "contact-123",
		Config: map[string]interface{}{
			"source_fields": []interface{}{"first_name", "middle_name", "last_name"},
			"target_field":  "full_name",
			"skip_empty":    false,
		},
		Connector: mockConn,
	}

	_, err := helper.Execute(ctx, input)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if mockConn.fieldsSet["full_name"] != "John  Doe" {
		t.Errorf("Expected full_name to be 'John  Doe', got: %v", mockConn.fieldsSet["full_name"])
	}
}

func TestCombineIt_Execute_AllFieldsEmpty(t *testing.T) {
	helper := &CombineIt{}
	mockConn := &mockConnectorForCombineIt{
		fieldValues: map[string]interface{}{
			"first_name": "",
			"last_name":  "",
		},
	}
	ctx := context.Background()

	input := helpers.HelperInput{
		ContactID: "contact-123",
		Config: map[string]interface{}{
			"source_fields": []interface{}{"first_name", "last_name"},
			"target_field":  "full_name",
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

	if output.Message != "All source fields are empty, nothing to combine" {
		t.Errorf("Unexpected message: %s", output.Message)
	}

	// Should not set field when all empty
	if _, set := mockConn.fieldsSet["full_name"]; set {
		t.Error("Expected full_name not to be set when all fields empty")
	}
}

func TestCombineIt_Execute_GetFieldError(t *testing.T) {
	helper := &CombineIt{}
	mockConn := &mockConnectorForCombineIt{
		getFieldError: fmt.Errorf("field read error"),
	}
	ctx := context.Background()

	input := helpers.HelperInput{
		ContactID: "contact-123",
		Config: map[string]interface{}{
			"source_fields": []interface{}{"first_name"},
			"target_field":  "full_name",
		},
		Connector: mockConn,
	}

	output, err := helper.Execute(ctx, input)
	// Should still succeed even if field read fails (logged as warning)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Check warning log
	if len(output.Logs) == 0 {
		t.Error("Expected warning log for field read error")
	}
}

func TestCombineIt_Execute_SetFieldError(t *testing.T) {
	helper := &CombineIt{}
	mockConn := &mockConnectorForCombineIt{
		fieldValues: map[string]interface{}{
			"first_name": "John",
		},
		setFieldError: fmt.Errorf("field write error"),
	}
	ctx := context.Background()

	input := helpers.HelperInput{
		ContactID: "contact-123",
		Config: map[string]interface{}{
			"source_fields": []interface{}{"first_name"},
			"target_field":  "full_name",
		},
		Connector: mockConn,
	}

	output, err := helper.Execute(ctx, input)
	if err == nil {
		t.Error("Expected error for set field failure")
	}

	if output.Success {
		t.Error("Expected success to be false")
	}
}

func TestCombineIt_Execute_ActionsRecorded(t *testing.T) {
	helper := &CombineIt{}
	mockConn := &mockConnectorForCombineIt{
		fieldValues: map[string]interface{}{
			"first_name": "John",
			"last_name":  "Doe",
		},
	}
	ctx := context.Background()

	input := helpers.HelperInput{
		ContactID: "contact-123",
		Config: map[string]interface{}{
			"source_fields": []interface{}{"first_name", "last_name"},
			"target_field":  "full_name",
		},
		Connector: mockConn,
	}

	output, err := helper.Execute(ctx, input)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(output.Actions) != 1 {
		t.Fatalf("Expected 1 action, got %d", len(output.Actions))
	}

	action := output.Actions[0]
	if action.Type != "field_updated" {
		t.Errorf("Expected action type 'field_updated', got '%s'", action.Type)
	}
	if action.Target != "full_name" {
		t.Errorf("Expected action target 'full_name', got '%s'", action.Target)
	}
	if action.Value != "John Doe" {
		t.Errorf("Expected action value 'John Doe', got '%v'", action.Value)
	}
}

func TestCombineIt_Execute_NilFieldValue(t *testing.T) {
	helper := &CombineIt{}
	mockConn := &mockConnectorForCombineIt{
		fieldValues: map[string]interface{}{
			"first_name": "John",
			"last_name":  nil,
		},
	}
	ctx := context.Background()

	input := helpers.HelperInput{
		ContactID: "contact-123",
		Config: map[string]interface{}{
			"source_fields": []interface{}{"first_name", "last_name"},
			"target_field":  "full_name",
			"skip_empty":    true,
		},
		Connector: mockConn,
	}

	_, err := helper.Execute(ctx, input)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if mockConn.fieldsSet["full_name"] != "John" {
		t.Errorf("Expected full_name to be 'John' (nil skipped), got: %v", mockConn.fieldsSet["full_name"])
	}
}
