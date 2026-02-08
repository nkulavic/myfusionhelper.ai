package integration

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/myfusionhelper/api/internal/connectors"
	"github.com/myfusionhelper/api/internal/helpers"
)

// mockConnectorForExcelIt implements the CRMConnector interface for testing excel_it
type mockConnectorForExcelIt struct {
	contact       *connectors.NormalizedContact
	getContactErr error
}

func (m *mockConnectorForExcelIt) GetContact(ctx context.Context, contactID string) (*connectors.NormalizedContact, error) {
	if m.getContactErr != nil {
		return nil, m.getContactErr
	}
	return m.contact, nil
}

// Stub implementations
func (m *mockConnectorForExcelIt) GetContacts(ctx context.Context, opts connectors.QueryOptions) (*connectors.ContactList, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *mockConnectorForExcelIt) CreateContact(ctx context.Context, input connectors.CreateContactInput) (*connectors.NormalizedContact, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *mockConnectorForExcelIt) UpdateContact(ctx context.Context, contactID string, updates connectors.UpdateContactInput) (*connectors.NormalizedContact, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *mockConnectorForExcelIt) DeleteContact(ctx context.Context, contactID string) error {
	return fmt.Errorf("not implemented")
}

func (m *mockConnectorForExcelIt) GetTags(ctx context.Context) ([]connectors.Tag, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *mockConnectorForExcelIt) ApplyTag(ctx context.Context, contactID string, tagID string) error {
	return fmt.Errorf("not implemented")
}

func (m *mockConnectorForExcelIt) RemoveTag(ctx context.Context, contactID string, tagID string) error {
	return fmt.Errorf("not implemented")
}

func (m *mockConnectorForExcelIt) GetCustomFields(ctx context.Context) ([]connectors.CustomField, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *mockConnectorForExcelIt) GetContactFieldValue(ctx context.Context, contactID, fieldKey string) (interface{}, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *mockConnectorForExcelIt) SetContactFieldValue(ctx context.Context, contactID, fieldKey string, value interface{}) error {
	return fmt.Errorf("not implemented")
}

func (m *mockConnectorForExcelIt) TriggerAutomation(ctx context.Context, contactID string, automationID string) error {
	return fmt.Errorf("not implemented")
}

func (m *mockConnectorForExcelIt) AchieveGoal(ctx context.Context, contactID string, goalName string, integration string) error {
	return fmt.Errorf("not implemented")
}

func (m *mockConnectorForExcelIt) TestConnection(ctx context.Context) error {
	return fmt.Errorf("not implemented")
}

func (m *mockConnectorForExcelIt) GetMetadata() connectors.ConnectorMetadata {
	return connectors.ConnectorMetadata{
		PlatformSlug: "test",
		PlatformName: "Test Platform",
	}
}

func (m *mockConnectorForExcelIt) GetCapabilities() []connectors.Capability {
	return []connectors.Capability{connectors.CapContacts}
}

// Tests

func TestExcelIt_GetMetadata(t *testing.T) {
	helper := &ExcelIt{}

	if helper.GetName() != "Excel It" {
		t.Errorf("Expected name 'Excel It', got '%s'", helper.GetName())
	}
	if helper.GetType() != "excel_it" {
		t.Errorf("Expected type 'excel_it', got '%s'", helper.GetType())
	}
	if helper.GetCategory() != "integration" {
		t.Errorf("Expected category 'integration', got '%s'", helper.GetCategory())
	}
	if !helper.RequiresCRM() {
		t.Error("Expected RequiresCRM to be true")
	}
}

func TestExcelIt_GetConfigSchema(t *testing.T) {
	helper := &ExcelIt{}
	schema := helper.GetConfigSchema()

	if schema == nil {
		t.Fatal("Schema should not be nil")
	}

	props, ok := schema["properties"].(map[string]interface{})
	if !ok {
		t.Fatal("Schema should have properties")
	}

	if _, ok := props["fields"]; !ok {
		t.Error("Schema should have fields property")
	}
	if _, ok := props["format"]; !ok {
		t.Error("Schema should have format property")
	}
	if _, ok := props["delimiter"]; !ok {
		t.Error("Schema should have delimiter property")
	}
	if _, ok := props["include_headers"]; !ok {
		t.Error("Schema should have include_headers property")
	}
}

func TestExcelIt_ValidateConfig_MissingFields(t *testing.T) {
	helper := &ExcelIt{}

	config := map[string]interface{}{}
	err := helper.ValidateConfig(config)
	if err == nil {
		t.Error("Expected error for missing fields")
	}
}

func TestExcelIt_ValidateConfig_EmptyFields(t *testing.T) {
	helper := &ExcelIt{}

	config := map[string]interface{}{
		"fields": []interface{}{},
	}
	err := helper.ValidateConfig(config)
	if err == nil {
		t.Error("Expected error for empty fields")
	}
}

func TestExcelIt_ValidateConfig_InvalidFormat(t *testing.T) {
	helper := &ExcelIt{}

	config := map[string]interface{}{
		"fields": []interface{}{"FirstName", "Email"},
		"format": "pdf",
	}
	err := helper.ValidateConfig(config)
	if err == nil {
		t.Error("Expected error for invalid format")
	}
}

func TestExcelIt_ValidateConfig_Valid(t *testing.T) {
	helper := &ExcelIt{}

	config := map[string]interface{}{
		"fields": []interface{}{"FirstName", "LastName", "Email"},
	}
	err := helper.ValidateConfig(config)
	if err != nil {
		t.Errorf("Expected no error for valid config, got: %v", err)
	}
}

func TestExcelIt_ValidateConfig_ValidWithFormat(t *testing.T) {
	helper := &ExcelIt{}

	config := map[string]interface{}{
		"fields": []interface{}{"FirstName", "Email"},
		"format": "xlsx",
	}
	err := helper.ValidateConfig(config)
	if err != nil {
		t.Errorf("Expected no error for valid config with xlsx format, got: %v", err)
	}
}

func TestExcelIt_Execute_Success_CSV(t *testing.T) {
	helper := &ExcelIt{}
	mockConn := &mockConnectorForExcelIt{
		contact: &connectors.NormalizedContact{
			ID:        "contact-123",
			FirstName: "John",
			LastName:  "Doe",
			Email:     "john.doe@example.com",
		},
	}
	ctx := context.Background()

	input := helpers.HelperInput{
		ContactID: "contact-123",
		Config: map[string]interface{}{
			"fields": []interface{}{"FirstName", "LastName", "Email"},
			"format": "csv",
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

	if output.ModifiedData["format"] != "csv" {
		t.Errorf("Expected format to be 'csv', got: %v", output.ModifiedData["format"])
	}
}

func TestExcelIt_Execute_Success_XLSX(t *testing.T) {
	helper := &ExcelIt{}
	mockConn := &mockConnectorForExcelIt{
		contact: &connectors.NormalizedContact{
			ID:        "contact-123",
			FirstName: "Jane",
			LastName:  "Smith",
			Email:     "jane.smith@example.com",
		},
	}
	ctx := context.Background()

	input := helpers.HelperInput{
		ContactID: "contact-123",
		Config: map[string]interface{}{
			"fields": []interface{}{"FirstName", "LastName", "Email"},
			"format": "xlsx",
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

	if output.ModifiedData["format"] != "xlsx" {
		t.Errorf("Expected format to be 'xlsx', got: %v", output.ModifiedData["format"])
	}
}

func TestExcelIt_Execute_CustomDelimiter(t *testing.T) {
	helper := &ExcelIt{}
	mockConn := &mockConnectorForExcelIt{
		contact: &connectors.NormalizedContact{
			ID:        "contact-123",
			FirstName: "John",
			LastName:  "Doe",
			Email:     "john.doe@example.com",
		},
	}
	ctx := context.Background()

	input := helpers.HelperInput{
		ContactID: "contact-123",
		Config: map[string]interface{}{
			"fields":    []interface{}{"FirstName", "LastName"},
			"delimiter": "|",
		},
		Connector: mockConn,
	}

	output, err := helper.Execute(ctx, input)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if output.ModifiedData["delimiter"] != "|" {
		t.Errorf("Expected delimiter to be '|', got: %v", output.ModifiedData["delimiter"])
	}

	dataRow := output.ModifiedData["data_row"].(string)
	if !strings.Contains(dataRow, "|") {
		t.Error("Expected data_row to contain custom delimiter")
	}
}

func TestExcelIt_Execute_HeadersIncluded(t *testing.T) {
	helper := &ExcelIt{}
	mockConn := &mockConnectorForExcelIt{
		contact: &connectors.NormalizedContact{
			ID:        "contact-123",
			FirstName: "John",
			LastName:  "Doe",
		},
	}
	ctx := context.Background()

	input := helpers.HelperInput{
		ContactID: "contact-123",
		Config: map[string]interface{}{
			"fields":          []interface{}{"FirstName", "LastName"},
			"include_headers": true,
		},
		Connector: mockConn,
	}

	output, err := helper.Execute(ctx, input)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if _, ok := output.ModifiedData["header_row"]; !ok {
		t.Error("Expected header_row to be present when include_headers is true")
	}

	headerRow := output.ModifiedData["header_row"].(string)
	if !strings.Contains(headerRow, "FirstName") || !strings.Contains(headerRow, "LastName") {
		t.Errorf("Expected header_row to contain field names, got: %s", headerRow)
	}
}

func TestExcelIt_Execute_HeadersExcluded(t *testing.T) {
	helper := &ExcelIt{}
	mockConn := &mockConnectorForExcelIt{
		contact: &connectors.NormalizedContact{
			ID:        "contact-123",
			FirstName: "John",
			LastName:  "Doe",
		},
	}
	ctx := context.Background()

	input := helpers.HelperInput{
		ContactID: "contact-123",
		Config: map[string]interface{}{
			"fields":          []interface{}{"FirstName", "LastName"},
			"include_headers": false,
		},
		Connector: mockConn,
	}

	output, err := helper.Execute(ctx, input)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if _, ok := output.ModifiedData["header_row"]; ok {
		t.Error("Expected header_row to be absent when include_headers is false")
	}
}

func TestExcelIt_Execute_CustomFields(t *testing.T) {
	helper := &ExcelIt{}
	mockConn := &mockConnectorForExcelIt{
		contact: &connectors.NormalizedContact{
			ID:        "contact-123",
			FirstName: "John",
			LastName:  "Doe",
			CustomFields: map[string]interface{}{
				"custom_score": 85,
				"custom_notes": "VIP customer",
			},
		},
	}
	ctx := context.Background()

	input := helpers.HelperInput{
		ContactID: "contact-123",
		Config: map[string]interface{}{
			"fields": []interface{}{"FirstName", "custom_score", "custom_notes"},
		},
		Connector: mockConn,
	}

	output, err := helper.Execute(ctx, input)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	dataRow := output.ModifiedData["data_row"].(string)
	if !strings.Contains(dataRow, "85") {
		t.Error("Expected data_row to contain custom_score value")
	}
	if !strings.Contains(dataRow, "VIP customer") {
		t.Error("Expected data_row to contain custom_notes value")
	}
}

func TestExcelIt_Execute_EscapesSpecialCharacters(t *testing.T) {
	helper := &ExcelIt{}
	mockConn := &mockConnectorForExcelIt{
		contact: &connectors.NormalizedContact{
			ID:        "contact-123",
			FirstName: "John, Jr.",
			LastName:  "Doe",
		},
	}
	ctx := context.Background()

	input := helpers.HelperInput{
		ContactID: "contact-123",
		Config: map[string]interface{}{
			"fields":    []interface{}{"FirstName", "LastName"},
			"delimiter": ",",
		},
		Connector: mockConn,
	}

	output, err := helper.Execute(ctx, input)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	dataRow := output.ModifiedData["data_row"].(string)
	// Value with comma should be quoted
	if !strings.Contains(dataRow, "\"John, Jr.\"") {
		t.Errorf("Expected FirstName to be quoted because it contains delimiter, got: %s", dataRow)
	}
}

func TestExcelIt_Execute_GetContactError(t *testing.T) {
	helper := &ExcelIt{}
	mockConn := &mockConnectorForExcelIt{
		getContactErr: fmt.Errorf("CRM API error"),
	}
	ctx := context.Background()

	input := helpers.HelperInput{
		ContactID: "contact-123",
		Config: map[string]interface{}{
			"fields": []interface{}{"FirstName", "Email"},
		},
		Connector: mockConn,
	}

	output, err := helper.Execute(ctx, input)
	if err == nil {
		t.Error("Expected error for GetContact failure")
	}

	if output.Success {
		t.Error("Expected success to be false")
	}

	if output.Message != "Failed to get contact data: CRM API error" {
		t.Errorf("Unexpected error message: %s", output.Message)
	}
}

func TestExcelIt_Execute_ActionsRecorded(t *testing.T) {
	helper := &ExcelIt{}
	mockConn := &mockConnectorForExcelIt{
		contact: &connectors.NormalizedContact{
			ID:        "contact-123",
			FirstName: "John",
			LastName:  "Doe",
			Email:     "john.doe@example.com",
		},
	}
	ctx := context.Background()

	input := helpers.HelperInput{
		ContactID: "contact-123",
		Config: map[string]interface{}{
			"fields": []interface{}{"FirstName", "Email"},
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
	if action.Type != "export_queued" {
		t.Errorf("Expected action type 'export_queued', got '%s'", action.Type)
	}
	if action.Target != "contact-123" {
		t.Errorf("Expected action target to be 'contact-123', got '%s'", action.Target)
	}
}

func TestExcelIt_Execute_LogsRecorded(t *testing.T) {
	helper := &ExcelIt{}
	mockConn := &mockConnectorForExcelIt{
		contact: &connectors.NormalizedContact{
			ID:        "contact-123",
			FirstName: "John",
			Email:     "john@example.com",
		},
	}
	ctx := context.Background()

	input := helpers.HelperInput{
		ContactID: "contact-123",
		Config: map[string]interface{}{
			"fields": []interface{}{"FirstName", "Email"},
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
}

func TestExcelIt_Execute_DefaultFormatCSV(t *testing.T) {
	helper := &ExcelIt{}
	mockConn := &mockConnectorForExcelIt{
		contact: &connectors.NormalizedContact{
			ID:        "contact-123",
			FirstName: "John",
		},
	}
	ctx := context.Background()

	input := helpers.HelperInput{
		ContactID: "contact-123",
		Config: map[string]interface{}{
			"fields": []interface{}{"FirstName"},
			// format not specified, should default to csv
		},
		Connector: mockConn,
	}

	output, err := helper.Execute(ctx, input)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if output.ModifiedData["format"] != "csv" {
		t.Errorf("Expected format to default to 'csv', got: %v", output.ModifiedData["format"])
	}
}

func TestExcelIt_Execute_EmptyFieldValue(t *testing.T) {
	helper := &ExcelIt{}
	mockConn := &mockConnectorForExcelIt{
		contact: &connectors.NormalizedContact{
			ID:        "contact-123",
			FirstName: "John",
			LastName:  "",
			Email:     "john@example.com",
		},
	}
	ctx := context.Background()

	input := helpers.HelperInput{
		ContactID: "contact-123",
		Config: map[string]interface{}{
			"fields": []interface{}{"FirstName", "LastName", "Email"},
		},
		Connector: mockConn,
	}

	output, err := helper.Execute(ctx, input)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if !output.Success {
		t.Error("Expected success to be true even with empty field values")
	}

	// Data row should have empty value for LastName
	dataRow := output.ModifiedData["data_row"].(string)
	parts := strings.Split(dataRow, ",")
	if len(parts) != 3 {
		t.Errorf("Expected 3 fields in data_row, got %d", len(parts))
	}
}
