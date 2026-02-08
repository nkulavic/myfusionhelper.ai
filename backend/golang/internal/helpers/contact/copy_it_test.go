package contact

import (
	"context"
	"fmt"
	"testing"

	"github.com/myfusionhelper/api/internal/connectors"
	"github.com/myfusionhelper/api/internal/helpers"
)

// mockConnectorForCopyIt implements the CRMConnector interface for testing copy_it
type mockConnectorForCopyIt struct {
	fieldsSet     map[string]interface{}
	fieldValues   map[string]interface{}
	setFieldError error
	getFieldError error
}

func (m *mockConnectorForCopyIt) SetContactFieldValue(ctx context.Context, contactID, fieldKey string, value interface{}) error {
	if m.setFieldError != nil {
		return m.setFieldError
	}
	if m.fieldsSet == nil {
		m.fieldsSet = make(map[string]interface{})
	}
	m.fieldsSet[fieldKey] = value
	return nil
}

func (m *mockConnectorForCopyIt) GetContactFieldValue(ctx context.Context, contactID, fieldKey string) (interface{}, error) {
	if m.getFieldError != nil {
		return nil, m.getFieldError
	}
	if m.fieldValues == nil {
		return nil, nil
	}
	return m.fieldValues[fieldKey], nil
}

// Stub implementations with correct signatures
func (m *mockConnectorForCopyIt) GetContacts(ctx context.Context, opts connectors.QueryOptions) (*connectors.ContactList, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *mockConnectorForCopyIt) GetContact(ctx context.Context, contactID string) (*connectors.NormalizedContact, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *mockConnectorForCopyIt) CreateContact(ctx context.Context, input connectors.CreateContactInput) (*connectors.NormalizedContact, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *mockConnectorForCopyIt) UpdateContact(ctx context.Context, contactID string, updates connectors.UpdateContactInput) (*connectors.NormalizedContact, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *mockConnectorForCopyIt) DeleteContact(ctx context.Context, contactID string) error {
	return fmt.Errorf("not implemented")
}

func (m *mockConnectorForCopyIt) GetTags(ctx context.Context) ([]connectors.Tag, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *mockConnectorForCopyIt) ApplyTag(ctx context.Context, contactID string, tagID string) error {
	return fmt.Errorf("not implemented")
}

func (m *mockConnectorForCopyIt) RemoveTag(ctx context.Context, contactID string, tagID string) error {
	return fmt.Errorf("not implemented")
}

func (m *mockConnectorForCopyIt) GetCustomFields(ctx context.Context) ([]connectors.CustomField, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *mockConnectorForCopyIt) TriggerAutomation(ctx context.Context, contactID string, automationID string) error {
	return fmt.Errorf("not implemented")
}

func (m *mockConnectorForCopyIt) AchieveGoal(ctx context.Context, contactID string, goalName string, integration string) error {
	return fmt.Errorf("not implemented")
}

func (m *mockConnectorForCopyIt) TestConnection(ctx context.Context) error {
	return fmt.Errorf("not implemented")
}

func (m *mockConnectorForCopyIt) GetMetadata() connectors.ConnectorMetadata {
	return connectors.ConnectorMetadata{
		PlatformSlug: "test",
		PlatformName: "Test Platform",
	}
}

func (m *mockConnectorForCopyIt) GetCapabilities() []connectors.Capability {
	return []connectors.Capability{connectors.CapContacts, connectors.CapCustomFields}
}

// Tests

func TestCopyIt_GetMetadata(t *testing.T) {
	helper := &CopyIt{}

	if helper.GetName() != "Copy It" {
		t.Errorf("Expected name 'Copy It', got '%s'", helper.GetName())
	}
	if helper.GetType() != "copy_it" {
		t.Errorf("Expected type 'copy_it', got '%s'", helper.GetType())
	}
	if helper.GetCategory() != "contact" {
		t.Errorf("Expected category 'contact', got '%s'", helper.GetCategory())
	}
	if !helper.RequiresCRM() {
		t.Error("Expected RequiresCRM to be true")
	}
}

func TestCopyIt_GetConfigSchema(t *testing.T) {
	helper := &CopyIt{}
	schema := helper.GetConfigSchema()

	if schema == nil {
		t.Fatal("Schema should not be nil")
	}

	props, ok := schema["properties"].(map[string]interface{})
	if !ok {
		t.Fatal("Schema should have properties")
	}

	if _, ok := props["source_field"]; !ok {
		t.Error("Schema should have source_field property")
	}
	if _, ok := props["target_field"]; !ok {
		t.Error("Schema should have target_field property")
	}
	if _, ok := props["overwrite"]; !ok {
		t.Error("Schema should have overwrite property")
	}
}

func TestCopyIt_ValidateConfig_MissingSourceField(t *testing.T) {
	helper := &CopyIt{}

	config := map[string]interface{}{
		"target_field": "backup_email",
	}
	err := helper.ValidateConfig(config)
	if err == nil {
		t.Error("Expected error for missing source_field")
	}
}

func TestCopyIt_ValidateConfig_EmptySourceField(t *testing.T) {
	helper := &CopyIt{}

	config := map[string]interface{}{
		"source_field": "",
		"target_field": "backup_email",
	}
	err := helper.ValidateConfig(config)
	if err == nil {
		t.Error("Expected error for empty source_field")
	}
}

func TestCopyIt_ValidateConfig_MissingTargetField(t *testing.T) {
	helper := &CopyIt{}

	config := map[string]interface{}{
		"source_field": "email",
	}
	err := helper.ValidateConfig(config)
	if err == nil {
		t.Error("Expected error for missing target_field")
	}
}

func TestCopyIt_ValidateConfig_EmptyTargetField(t *testing.T) {
	helper := &CopyIt{}

	config := map[string]interface{}{
		"source_field": "email",
		"target_field": "",
	}
	err := helper.ValidateConfig(config)
	if err == nil {
		t.Error("Expected error for empty target_field")
	}
}

func TestCopyIt_ValidateConfig_Valid(t *testing.T) {
	helper := &CopyIt{}

	config := map[string]interface{}{
		"source_field": "email",
		"target_field": "backup_email",
	}
	err := helper.ValidateConfig(config)
	if err != nil {
		t.Errorf("Expected no error for valid config, got: %v", err)
	}
}

func TestCopyIt_Execute_Success(t *testing.T) {
	helper := &CopyIt{}
	mockConn := &mockConnectorForCopyIt{
		fieldValues: map[string]interface{}{
			"email": "test@example.com",
		},
	}
	ctx := context.Background()

	input := helpers.HelperInput{
		ContactID: "contact-123",
		Config: map[string]interface{}{
			"source_field": "email",
			"target_field": "backup_email",
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

	if mockConn.fieldsSet["backup_email"] != "test@example.com" {
		t.Errorf("Expected backup_email to be 'test@example.com', got: %v", mockConn.fieldsSet["backup_email"])
	}

	if output.Message != "Copied 'email' to 'backup_email'" {
		t.Errorf("Unexpected message: %s", output.Message)
	}
}

func TestCopyIt_Execute_EmptySourceField(t *testing.T) {
	helper := &CopyIt{}
	mockConn := &mockConnectorForCopyIt{
		fieldValues: map[string]interface{}{
			"email": "",
		},
	}
	ctx := context.Background()

	input := helpers.HelperInput{
		ContactID: "contact-123",
		Config: map[string]interface{}{
			"source_field": "email",
			"target_field": "backup_email",
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

	if output.Message != "Source field 'email' is empty, nothing to copy" {
		t.Errorf("Unexpected message: %s", output.Message)
	}

	// Should not set target field when source is empty
	if _, set := mockConn.fieldsSet["backup_email"]; set {
		t.Error("Expected backup_email not to be set when source is empty")
	}
}

func TestCopyIt_Execute_NilSourceField(t *testing.T) {
	helper := &CopyIt{}
	mockConn := &mockConnectorForCopyIt{
		fieldValues: map[string]interface{}{
			"email": nil,
		},
	}
	ctx := context.Background()

	input := helpers.HelperInput{
		ContactID: "contact-123",
		Config: map[string]interface{}{
			"source_field": "email",
			"target_field": "backup_email",
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

	// Should not set target field when source is nil
	if _, set := mockConn.fieldsSet["backup_email"]; set {
		t.Error("Expected backup_email not to be set when source is nil")
	}
}

func TestCopyIt_Execute_OverwriteTrue(t *testing.T) {
	helper := &CopyIt{}
	mockConn := &mockConnectorForCopyIt{
		fieldValues: map[string]interface{}{
			"email":        "test@example.com",
			"backup_email": "old@example.com",
		},
	}
	ctx := context.Background()

	input := helpers.HelperInput{
		ContactID: "contact-123",
		Config: map[string]interface{}{
			"source_field": "email",
			"target_field": "backup_email",
			"overwrite":    true,
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

	if mockConn.fieldsSet["backup_email"] != "test@example.com" {
		t.Errorf("Expected backup_email to be overwritten to 'test@example.com', got: %v", mockConn.fieldsSet["backup_email"])
	}
}

func TestCopyIt_Execute_OverwriteFalse_TargetHasValue(t *testing.T) {
	helper := &CopyIt{}
	mockConn := &mockConnectorForCopyIt{
		fieldValues: map[string]interface{}{
			"email":        "test@example.com",
			"backup_email": "old@example.com",
		},
	}
	ctx := context.Background()

	input := helpers.HelperInput{
		ContactID: "contact-123",
		Config: map[string]interface{}{
			"source_field": "email",
			"target_field": "backup_email",
			"overwrite":    false,
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

	if output.Message != "Target field 'backup_email' already has a value, skipping (overwrite=false)" {
		t.Errorf("Unexpected message: %s", output.Message)
	}

	// Should not set target field when overwrite=false and target has value
	if _, set := mockConn.fieldsSet["backup_email"]; set {
		t.Error("Expected backup_email not to be set when overwrite=false and target has value")
	}
}

func TestCopyIt_Execute_OverwriteFalse_TargetEmpty(t *testing.T) {
	helper := &CopyIt{}
	mockConn := &mockConnectorForCopyIt{
		fieldValues: map[string]interface{}{
			"email":        "test@example.com",
			"backup_email": "",
		},
	}
	ctx := context.Background()

	input := helpers.HelperInput{
		ContactID: "contact-123",
		Config: map[string]interface{}{
			"source_field": "email",
			"target_field": "backup_email",
			"overwrite":    false,
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

	// Should copy when target is empty even if overwrite=false
	if mockConn.fieldsSet["backup_email"] != "test@example.com" {
		t.Errorf("Expected backup_email to be set when target is empty, got: %v", mockConn.fieldsSet["backup_email"])
	}
}

func TestCopyIt_Execute_GetSourceFieldError(t *testing.T) {
	helper := &CopyIt{}
	mockConn := &mockConnectorForCopyIt{
		getFieldError: fmt.Errorf("field read error"),
	}
	ctx := context.Background()

	input := helpers.HelperInput{
		ContactID: "contact-123",
		Config: map[string]interface{}{
			"source_field": "email",
			"target_field": "backup_email",
		},
		Connector: mockConn,
	}

	output, err := helper.Execute(ctx, input)
	if err == nil {
		t.Error("Expected error for get field failure")
	}

	if output.Success {
		t.Error("Expected success to be false")
	}

	if output.Message != "Failed to read source field 'email': field read error" {
		t.Errorf("Unexpected error message: %s", output.Message)
	}
}

func TestCopyIt_Execute_SetTargetFieldError(t *testing.T) {
	helper := &CopyIt{}
	mockConn := &mockConnectorForCopyIt{
		fieldValues: map[string]interface{}{
			"email": "test@example.com",
		},
		setFieldError: fmt.Errorf("field write error"),
	}
	ctx := context.Background()

	input := helpers.HelperInput{
		ContactID: "contact-123",
		Config: map[string]interface{}{
			"source_field": "email",
			"target_field": "backup_email",
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

	if output.Message != "Failed to write to target field 'backup_email': field write error" {
		t.Errorf("Unexpected error message: %s", output.Message)
	}
}

func TestCopyIt_Execute_ActionsRecorded(t *testing.T) {
	helper := &CopyIt{}
	mockConn := &mockConnectorForCopyIt{
		fieldValues: map[string]interface{}{
			"email": "test@example.com",
		},
	}
	ctx := context.Background()

	input := helpers.HelperInput{
		ContactID: "contact-123",
		Config: map[string]interface{}{
			"source_field": "email",
			"target_field": "backup_email",
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
	if action.Target != "backup_email" {
		t.Errorf("Expected action target 'backup_email', got '%s'", action.Target)
	}
	if action.Value != "test@example.com" {
		t.Errorf("Expected action value 'test@example.com', got '%v'", action.Value)
	}
}

func TestCopyIt_Execute_ModifiedDataRecorded(t *testing.T) {
	helper := &CopyIt{}
	mockConn := &mockConnectorForCopyIt{
		fieldValues: map[string]interface{}{
			"email": "test@example.com",
		},
	}
	ctx := context.Background()

	input := helpers.HelperInput{
		ContactID: "contact-123",
		Config: map[string]interface{}{
			"source_field": "email",
			"target_field": "backup_email",
		},
		Connector: mockConn,
	}

	output, err := helper.Execute(ctx, input)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if output.ModifiedData == nil {
		t.Fatal("Expected ModifiedData to be set")
	}

	if output.ModifiedData["backup_email"] != "test@example.com" {
		t.Errorf("Expected ModifiedData backup_email to be 'test@example.com', got '%v'", output.ModifiedData["backup_email"])
	}
}

func TestCopyIt_Execute_LogsRecorded(t *testing.T) {
	helper := &CopyIt{}
	mockConn := &mockConnectorForCopyIt{
		fieldValues: map[string]interface{}{
			"email": "test@example.com",
		},
	}
	ctx := context.Background()

	input := helpers.HelperInput{
		ContactID: "contact-123",
		Config: map[string]interface{}{
			"source_field": "email",
			"target_field": "backup_email",
		},
		Connector: mockConn,
	}

	output, err := helper.Execute(ctx, input)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(output.Logs) == 0 {
		t.Error("Expected logs to be recorded")
	}

	expectedLog := "Copied value from 'email' to 'backup_email' on contact contact-123"
	if output.Logs[0] != expectedLog {
		t.Errorf("Expected log '%s', got '%s'", expectedLog, output.Logs[0])
	}
}

func TestCopyIt_Execute_DefaultOverwriteTrue(t *testing.T) {
	helper := &CopyIt{}
	mockConn := &mockConnectorForCopyIt{
		fieldValues: map[string]interface{}{
			"email":        "test@example.com",
			"backup_email": "old@example.com",
		},
	}
	ctx := context.Background()

	input := helpers.HelperInput{
		ContactID: "contact-123",
		Config: map[string]interface{}{
			"source_field": "email",
			"target_field": "backup_email",
			// overwrite not specified, should default to true
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

	// Should overwrite by default
	if mockConn.fieldsSet["backup_email"] != "test@example.com" {
		t.Errorf("Expected backup_email to be overwritten by default, got: %v", mockConn.fieldsSet["backup_email"])
	}
}
