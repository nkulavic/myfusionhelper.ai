package contact

import (
	"context"
	"fmt"
	"testing"

	"github.com/myfusionhelper/api/internal/connectors"
	"github.com/myfusionhelper/api/internal/helpers"
)

// mockConnectorForCompanyLink implements the CRMConnector interface for testing company_link
type mockConnectorForCompanyLink struct {
	fieldsSet     map[string]interface{}
	fieldValues   map[string]interface{}
	setFieldError error
	getFieldError error
}

func (m *mockConnectorForCompanyLink) SetContactFieldValue(ctx context.Context, contactID, fieldKey string, value interface{}) error {
	if m.setFieldError != nil {
		return m.setFieldError
	}
	if m.fieldsSet == nil {
		m.fieldsSet = make(map[string]interface{})
	}
	m.fieldsSet[fieldKey] = value
	return nil
}

func (m *mockConnectorForCompanyLink) GetContactFieldValue(ctx context.Context, contactID, fieldKey string) (interface{}, error) {
	if m.getFieldError != nil {
		return nil, m.getFieldError
	}
	if m.fieldValues == nil {
		return nil, nil
	}
	return m.fieldValues[fieldKey], nil
}

// Stub implementations with correct signatures
func (m *mockConnectorForCompanyLink) GetContacts(ctx context.Context, opts connectors.QueryOptions) (*connectors.ContactList, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *mockConnectorForCompanyLink) GetContact(ctx context.Context, contactID string) (*connectors.NormalizedContact, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *mockConnectorForCompanyLink) CreateContact(ctx context.Context, input connectors.CreateContactInput) (*connectors.NormalizedContact, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *mockConnectorForCompanyLink) UpdateContact(ctx context.Context, contactID string, updates connectors.UpdateContactInput) (*connectors.NormalizedContact, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *mockConnectorForCompanyLink) DeleteContact(ctx context.Context, contactID string) error {
	return fmt.Errorf("not implemented")
}

func (m *mockConnectorForCompanyLink) GetTags(ctx context.Context) ([]connectors.Tag, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *mockConnectorForCompanyLink) ApplyTag(ctx context.Context, contactID string, tagID string) error {
	return fmt.Errorf("not implemented")
}

func (m *mockConnectorForCompanyLink) RemoveTag(ctx context.Context, contactID string, tagID string) error {
	return fmt.Errorf("not implemented")
}

func (m *mockConnectorForCompanyLink) GetCustomFields(ctx context.Context) ([]connectors.CustomField, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *mockConnectorForCompanyLink) TriggerAutomation(ctx context.Context, contactID string, automationID string) error {
	return fmt.Errorf("not implemented")
}

func (m *mockConnectorForCompanyLink) AchieveGoal(ctx context.Context, contactID string, goalName string, integration string) error {
	return fmt.Errorf("not implemented")
}

func (m *mockConnectorForCompanyLink) TestConnection(ctx context.Context) error {
	return fmt.Errorf("not implemented")
}

func (m *mockConnectorForCompanyLink) GetMetadata() connectors.ConnectorMetadata {
	return connectors.ConnectorMetadata{
		PlatformSlug: "test",
		PlatformName: "Test Platform",
	}
}

func (m *mockConnectorForCompanyLink) GetCapabilities() []connectors.Capability {
	return []connectors.Capability{connectors.CapContacts, connectors.CapCustomFields}
}

// Tests

func TestCompanyLink_GetMetadata(t *testing.T) {
	helper := &CompanyLink{}

	if helper.GetName() != "Company Link" {
		t.Errorf("Expected name 'Company Link', got '%s'", helper.GetName())
	}
	if helper.GetType() != "company_link" {
		t.Errorf("Expected type 'company_link', got '%s'", helper.GetType())
	}
	if helper.GetCategory() != "contact" {
		t.Errorf("Expected category 'contact', got '%s'", helper.GetCategory())
	}
	if !helper.RequiresCRM() {
		t.Error("Expected RequiresCRM to be true")
	}
}

func TestCompanyLink_SupportedCRMs(t *testing.T) {
	helper := &CompanyLink{}
	supportedCRMs := helper.SupportedCRMs()

	if len(supportedCRMs) == 0 {
		t.Error("Expected at least one supported CRM")
	}

	if supportedCRMs[0] != "keap" {
		t.Errorf("Expected 'keap' to be supported, got '%s'", supportedCRMs[0])
	}
}

func TestCompanyLink_GetConfigSchema(t *testing.T) {
	helper := &CompanyLink{}
	schema := helper.GetConfigSchema()

	if schema == nil {
		t.Fatal("Schema should not be nil")
	}

	props, ok := schema["properties"].(map[string]interface{})
	if !ok {
		t.Fatal("Schema should have properties")
	}

	if _, ok := props["company_field"]; !ok {
		t.Error("Schema should have company_field property")
	}

	required, ok := schema["required"].([]string)
	if !ok {
		t.Fatal("Schema should have required array")
	}

	if len(required) == 0 || required[0] != "company_field" {
		t.Error("Schema should require company_field")
	}
}

func TestCompanyLink_ValidateConfig_MissingCompanyField(t *testing.T) {
	helper := &CompanyLink{}

	config := map[string]interface{}{}
	err := helper.ValidateConfig(config)
	if err == nil {
		t.Error("Expected error for missing company_field")
	}
}

func TestCompanyLink_ValidateConfig_EmptyCompanyField(t *testing.T) {
	helper := &CompanyLink{}

	config := map[string]interface{}{
		"company_field": "",
	}
	err := helper.ValidateConfig(config)
	if err == nil {
		t.Error("Expected error for empty company_field")
	}
}

func TestCompanyLink_ValidateConfig_InvalidCompanyFieldType(t *testing.T) {
	helper := &CompanyLink{}

	config := map[string]interface{}{
		"company_field": 123,
	}
	err := helper.ValidateConfig(config)
	if err == nil {
		t.Error("Expected error for non-string company_field")
	}
}

func TestCompanyLink_ValidateConfig_Valid(t *testing.T) {
	helper := &CompanyLink{}

	config := map[string]interface{}{
		"company_field": "company_name",
	}
	err := helper.ValidateConfig(config)
	if err != nil {
		t.Errorf("Expected no error for valid config, got: %v", err)
	}
}

func TestCompanyLink_Execute_Success(t *testing.T) {
	helper := &CompanyLink{}
	mockConn := &mockConnectorForCompanyLink{
		fieldValues: map[string]interface{}{
			"company_name": "Acme Corp",
		},
	}
	ctx := context.Background()

	input := helpers.HelperInput{
		ContactID: "contact-123",
		Config: map[string]interface{}{
			"company_field": "company_name",
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

	if mockConn.fieldsSet["company"] != "Acme Corp" {
		t.Errorf("Expected company to be set to 'Acme Corp', got: %v", mockConn.fieldsSet["company"])
	}

	if output.Message != "Linked contact to company 'Acme Corp'" {
		t.Errorf("Unexpected message: %s", output.Message)
	}
}

func TestCompanyLink_Execute_EmptyCompanyField(t *testing.T) {
	helper := &CompanyLink{}
	mockConn := &mockConnectorForCompanyLink{
		fieldValues: map[string]interface{}{
			"company_name": "",
		},
	}
	ctx := context.Background()

	input := helpers.HelperInput{
		ContactID: "contact-123",
		Config: map[string]interface{}{
			"company_field": "company_name",
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

	if output.Message != "Company field 'company_name' is empty, nothing to link" {
		t.Errorf("Unexpected message: %s", output.Message)
	}

	// Should not set company field when empty
	if _, set := mockConn.fieldsSet["company"]; set {
		t.Error("Expected company not to be set when company_name is empty")
	}
}

func TestCompanyLink_Execute_NilCompanyField(t *testing.T) {
	helper := &CompanyLink{}
	mockConn := &mockConnectorForCompanyLink{
		fieldValues: map[string]interface{}{
			"company_name": nil,
		},
	}
	ctx := context.Background()

	input := helpers.HelperInput{
		ContactID: "contact-123",
		Config: map[string]interface{}{
			"company_field": "company_name",
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

	// Should not set company field when nil
	if _, set := mockConn.fieldsSet["company"]; set {
		t.Error("Expected company not to be set when company_name is nil")
	}
}

func TestCompanyLink_Execute_GetFieldError(t *testing.T) {
	helper := &CompanyLink{}
	mockConn := &mockConnectorForCompanyLink{
		getFieldError: fmt.Errorf("field read error"),
	}
	ctx := context.Background()

	input := helpers.HelperInput{
		ContactID: "contact-123",
		Config: map[string]interface{}{
			"company_field": "company_name",
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

	if output.Message != "Failed to read company field 'company_name': field read error" {
		t.Errorf("Unexpected error message: %s", output.Message)
	}
}

func TestCompanyLink_Execute_SetFieldError(t *testing.T) {
	helper := &CompanyLink{}
	mockConn := &mockConnectorForCompanyLink{
		fieldValues: map[string]interface{}{
			"company_name": "Acme Corp",
		},
		setFieldError: fmt.Errorf("field write error"),
	}
	ctx := context.Background()

	input := helpers.HelperInput{
		ContactID: "contact-123",
		Config: map[string]interface{}{
			"company_field": "company_name",
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

	if output.Message != "Failed to link company 'Acme Corp': field write error" {
		t.Errorf("Unexpected error message: %s", output.Message)
	}
}

func TestCompanyLink_Execute_ActionsRecorded(t *testing.T) {
	helper := &CompanyLink{}
	mockConn := &mockConnectorForCompanyLink{
		fieldValues: map[string]interface{}{
			"company_name": "Acme Corp",
		},
	}
	ctx := context.Background()

	input := helpers.HelperInput{
		ContactID: "contact-123",
		Config: map[string]interface{}{
			"company_field": "company_name",
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
	if action.Type != "company_linked" {
		t.Errorf("Expected action type 'company_linked', got '%s'", action.Type)
	}
	if action.Target != "contact-123" {
		t.Errorf("Expected action target 'contact-123', got '%s'", action.Target)
	}
	if action.Value != "Acme Corp" {
		t.Errorf("Expected action value 'Acme Corp', got '%v'", action.Value)
	}
}

func TestCompanyLink_Execute_ModifiedDataRecorded(t *testing.T) {
	helper := &CompanyLink{}
	mockConn := &mockConnectorForCompanyLink{
		fieldValues: map[string]interface{}{
			"company_name": "Acme Corp",
		},
	}
	ctx := context.Background()

	input := helpers.HelperInput{
		ContactID: "contact-123",
		Config: map[string]interface{}{
			"company_field": "company_name",
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

	if output.ModifiedData["company"] != "Acme Corp" {
		t.Errorf("Expected ModifiedData company to be 'Acme Corp', got '%v'", output.ModifiedData["company"])
	}
}

func TestCompanyLink_Execute_LogsRecorded(t *testing.T) {
	helper := &CompanyLink{}
	mockConn := &mockConnectorForCompanyLink{
		fieldValues: map[string]interface{}{
			"company_name": "Acme Corp",
		},
	}
	ctx := context.Background()

	input := helpers.HelperInput{
		ContactID: "contact-123",
		Config: map[string]interface{}{
			"company_field": "company_name",
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

	expectedLog := "Linked contact contact-123 to company 'Acme Corp' via field 'company_name'"
	if output.Logs[0] != expectedLog {
		t.Errorf("Expected log '%s', got '%s'", expectedLog, output.Logs[0])
	}
}
