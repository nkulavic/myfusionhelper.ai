package data

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/myfusionhelper/api/internal/connectors"
	"github.com/myfusionhelper/api/internal/helpers"
)

// Mock connector for format_it testing
type mockConnectorForFormat struct {
	fieldValues   map[string]interface{}
	updatedFields map[string]interface{}
}

func (m *mockConnectorForFormat) GetContactFieldValue(ctx context.Context, contactID, fieldKey string) (interface{}, error) {
	if m.fieldValues == nil {
		return nil, fmt.Errorf("field not found")
	}
	val, ok := m.fieldValues[fieldKey]
	if !ok {
		return nil, fmt.Errorf("field '%s' not found", fieldKey)
	}
	return val, nil
}

func (m *mockConnectorForFormat) SetContactFieldValue(ctx context.Context, contactID, fieldKey string, value interface{}) error {
	if m.updatedFields == nil {
		m.updatedFields = make(map[string]interface{})
	}
	m.updatedFields[fieldKey] = value
	return nil
}

// Stub implementations for CRMConnector interface
func (m *mockConnectorForFormat) GetContact(ctx context.Context, contactID string) (*connectors.NormalizedContact, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForFormat) GetContacts(ctx context.Context, opts connectors.QueryOptions) (*connectors.ContactList, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForFormat) CreateContact(ctx context.Context, input connectors.CreateContactInput) (*connectors.NormalizedContact, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForFormat) UpdateContact(ctx context.Context, contactID string, updates connectors.UpdateContactInput) (*connectors.NormalizedContact, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForFormat) DeleteContact(ctx context.Context, contactID string) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorForFormat) GetTags(ctx context.Context) ([]connectors.Tag, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForFormat) ApplyTag(ctx context.Context, contactID, tagID string) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorForFormat) RemoveTag(ctx context.Context, contactID, tagID string) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorForFormat) GetCustomFields(ctx context.Context) ([]connectors.CustomField, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForFormat) TriggerAutomation(ctx context.Context, contactID, automationID string) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorForFormat) AchieveGoal(ctx context.Context, contactID, goalName, integration string) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorForFormat) TestConnection(ctx context.Context) error {
	return nil
}
func (m *mockConnectorForFormat) GetMetadata() connectors.ConnectorMetadata {
	return connectors.ConnectorMetadata{}
}
func (m *mockConnectorForFormat) GetCapabilities() []connectors.Capability {
	return nil
}

// Test helper metadata
func TestFormatIt_Metadata(t *testing.T) {
	helper := &FormatIt{}

	if helper.GetName() != "Format It" {
		t.Errorf("Expected name 'Format It', got '%s'", helper.GetName())
	}
	if helper.GetType() != "format_it" {
		t.Errorf("Expected type 'format_it', got '%s'", helper.GetType())
	}
	if helper.GetCategory() != "data" {
		t.Errorf("Expected category 'data', got '%s'", helper.GetCategory())
	}
	if !helper.RequiresCRM() {
		t.Error("Expected RequiresCRM to be true")
	}

	schema := helper.GetConfigSchema()
	if schema == nil {
		t.Fatal("Expected config schema, got nil")
	}
}

// Test validation - missing field
func TestFormatIt_ValidateConfig_MissingField(t *testing.T) {
	helper := &FormatIt{}
	config := map[string]interface{}{
		"format": "uppercase",
	}

	err := helper.ValidateConfig(config)
	if err == nil {
		t.Error("Expected validation error for missing field")
	}
	if !strings.Contains(err.Error(), "field") {
		t.Errorf("Expected error about field, got: %v", err)
	}
}

// Test validation - missing format
func TestFormatIt_ValidateConfig_MissingFormat(t *testing.T) {
	helper := &FormatIt{}
	config := map[string]interface{}{
		"field": "first_name",
	}

	err := helper.ValidateConfig(config)
	if err == nil {
		t.Error("Expected validation error for missing format")
	}
	if !strings.Contains(err.Error(), "format") {
		t.Errorf("Expected error about format, got: %v", err)
	}
}

// Test validation - invalid format
func TestFormatIt_ValidateConfig_InvalidFormat(t *testing.T) {
	helper := &FormatIt{}
	config := map[string]interface{}{
		"field":  "first_name",
		"format": "invalid_format",
	}

	err := helper.ValidateConfig(config)
	if err == nil {
		t.Error("Expected validation error for invalid format")
	}
	if !strings.Contains(err.Error(), "invalid format") {
		t.Errorf("Expected error about invalid format, got: %v", err)
	}
}

// Test validation - valid config
func TestFormatIt_ValidateConfig_Valid(t *testing.T) {
	helper := &FormatIt{}
	validFormats := []string{"uppercase", "lowercase", "title_case", "trim", "trim_uppercase", "trim_lowercase", "trim_title_case"}

	for _, format := range validFormats {
		config := map[string]interface{}{
			"field":  "first_name",
			"format": format,
		}

		err := helper.ValidateConfig(config)
		if err != nil {
			t.Errorf("Expected no validation error for format '%s', got: %v", format, err)
		}
	}
}

// Test execution - uppercase
func TestFormatIt_Execute_Uppercase(t *testing.T) {
	helper := &FormatIt{}

	mockConn := &mockConnectorForFormat{
		fieldValues: map[string]interface{}{
			"first_name": "john",
		},
	}

	input := helpers.HelperInput{
		ContactID: "123",
		Connector: mockConn,
		Config: map[string]interface{}{
			"field":  "first_name",
			"format": "uppercase",
		},
	}

	output, err := helper.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if !output.Success {
		t.Error("Expected success=true")
	}

	// Verify field was updated
	if updated, ok := mockConn.updatedFields["first_name"].(string); !ok || updated != "JOHN" {
		t.Errorf("Expected field value 'JOHN', got: %v", mockConn.updatedFields["first_name"])
	}
}

// Test execution - lowercase
func TestFormatIt_Execute_Lowercase(t *testing.T) {
	helper := &FormatIt{}

	mockConn := &mockConnectorForFormat{
		fieldValues: map[string]interface{}{
			"last_name": "SMITH",
		},
	}

	input := helpers.HelperInput{
		ContactID: "123",
		Connector: mockConn,
		Config: map[string]interface{}{
			"field":  "last_name",
			"format": "lowercase",
		},
	}

	output, err := helper.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if !output.Success {
		t.Error("Expected success=true")
	}

	// Verify field was updated
	if updated, ok := mockConn.updatedFields["last_name"].(string); !ok || updated != "smith" {
		t.Errorf("Expected field value 'smith', got: %v", mockConn.updatedFields["last_name"])
	}
}

// Test execution - title case
func TestFormatIt_Execute_TitleCase(t *testing.T) {
	helper := &FormatIt{}

	mockConn := &mockConnectorForFormat{
		fieldValues: map[string]interface{}{
			"company": "acme corporation",
		},
	}

	input := helpers.HelperInput{
		ContactID: "123",
		Connector: mockConn,
		Config: map[string]interface{}{
			"field":  "company",
			"format": "title_case",
		},
	}

	output, err := helper.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if !output.Success {
		t.Error("Expected success=true")
	}

	// Verify field was updated
	if updated, ok := mockConn.updatedFields["company"].(string); !ok || updated != "Acme Corporation" {
		t.Errorf("Expected field value 'Acme Corporation', got: %v", mockConn.updatedFields["company"])
	}
}

// Test execution - trim
func TestFormatIt_Execute_Trim(t *testing.T) {
	helper := &FormatIt{}

	mockConn := &mockConnectorForFormat{
		fieldValues: map[string]interface{}{
			"email": "  test@example.com  ",
		},
	}

	input := helpers.HelperInput{
		ContactID: "123",
		Connector: mockConn,
		Config: map[string]interface{}{
			"field":  "email",
			"format": "trim",
		},
	}

	output, err := helper.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if !output.Success {
		t.Error("Expected success=true")
	}

	// Verify field was updated
	if updated, ok := mockConn.updatedFields["email"].(string); !ok || updated != "test@example.com" {
		t.Errorf("Expected field value 'test@example.com', got: %v", mockConn.updatedFields["email"])
	}
}

// Test execution - trim_uppercase
func TestFormatIt_Execute_TrimUppercase(t *testing.T) {
	helper := &FormatIt{}

	mockConn := &mockConnectorForFormat{
		fieldValues: map[string]interface{}{
			"state": "  california  ",
		},
	}

	input := helpers.HelperInput{
		ContactID: "123",
		Connector: mockConn,
		Config: map[string]interface{}{
			"field":  "state",
			"format": "trim_uppercase",
		},
	}

	output, err := helper.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if !output.Success {
		t.Error("Expected success=true")
	}

	// Verify field was updated
	if updated, ok := mockConn.updatedFields["state"].(string); !ok || updated != "CALIFORNIA" {
		t.Errorf("Expected field value 'CALIFORNIA', got: %v", mockConn.updatedFields["state"])
	}
}

// Test execution - target_field (write to different field)
func TestFormatIt_Execute_TargetField(t *testing.T) {
	helper := &FormatIt{}

	mockConn := &mockConnectorForFormat{
		fieldValues: map[string]interface{}{
			"first_name": "john",
		},
	}

	input := helpers.HelperInput{
		ContactID: "123",
		Connector: mockConn,
		Config: map[string]interface{}{
			"field":        "first_name",
			"format":       "uppercase",
			"target_field": "display_name",
		},
	}

	output, err := helper.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if !output.Success {
		t.Error("Expected success=true")
	}

	// Verify target field was updated (not source field)
	if updated, ok := mockConn.updatedFields["display_name"].(string); !ok || updated != "JOHN" {
		t.Errorf("Expected display_name 'JOHN', got: %v", mockConn.updatedFields["display_name"])
	}
	if _, ok := mockConn.updatedFields["first_name"]; ok {
		t.Error("Source field should not have been updated when target_field is specified")
	}
}

// Test execution - empty field
func TestFormatIt_Execute_EmptyField(t *testing.T) {
	helper := &FormatIt{}

	mockConn := &mockConnectorForFormat{
		fieldValues: map[string]interface{}{
			"notes": "",
		},
	}

	input := helpers.HelperInput{
		ContactID: "123",
		Connector: mockConn,
		Config: map[string]interface{}{
			"field":  "notes",
			"format": "uppercase",
		},
	}

	output, err := helper.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if !output.Success {
		t.Error("Expected success=true for empty field")
	}

	// Verify no update was performed
	if len(mockConn.updatedFields) > 0 {
		t.Error("Expected no field updates for empty field")
	}
}

// Test execution - nil field value
func TestFormatIt_Execute_NilField(t *testing.T) {
	helper := &FormatIt{}

	mockConn := &mockConnectorForFormat{
		fieldValues: map[string]interface{}{
			"custom_field": nil,
		},
	}

	input := helpers.HelperInput{
		ContactID: "123",
		Connector: mockConn,
		Config: map[string]interface{}{
			"field":  "custom_field",
			"format": "uppercase",
		},
	}

	output, err := helper.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if !output.Success {
		t.Error("Expected success=true for nil field")
	}

	// Verify no update was performed
	if len(mockConn.updatedFields) > 0 {
		t.Error("Expected no field updates for nil field")
	}
}

// Test execution - already formatted (no change)
func TestFormatIt_Execute_AlreadyFormatted(t *testing.T) {
	helper := &FormatIt{}

	mockConn := &mockConnectorForFormat{
		fieldValues: map[string]interface{}{
			"first_name": "JOHN",
		},
	}

	input := helpers.HelperInput{
		ContactID: "123",
		Connector: mockConn,
		Config: map[string]interface{}{
			"field":  "first_name",
			"format": "uppercase",
		},
	}

	output, err := helper.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if !output.Success {
		t.Error("Expected success=true")
	}

	// Verify no update was performed (value already correct)
	if len(mockConn.updatedFields) > 0 {
		t.Error("Expected no field updates when value is already formatted correctly")
	}
}

// Test execution - field not found error
func TestFormatIt_Execute_FieldNotFound(t *testing.T) {
	helper := &FormatIt{}

	mockConn := &mockConnectorForFormat{
		fieldValues: map[string]interface{}{},
	}

	input := helpers.HelperInput{
		ContactID: "123",
		Connector: mockConn,
		Config: map[string]interface{}{
			"field":  "nonexistent_field",
			"format": "uppercase",
		},
	}

	output, err := helper.Execute(context.Background(), input)
	if err == nil {
		t.Error("Expected error for nonexistent field")
	}

	if output.Success {
		t.Error("Expected success=false for error case")
	}

	if !strings.Contains(output.Message, "Failed to read field") {
		t.Errorf("Expected error message about failed read, got: %s", output.Message)
	}
}

// Test title case with special characters
func TestFormatIt_Execute_TitleCaseSpecialChars(t *testing.T) {
	helper := &FormatIt{}

	testCases := []struct {
		input    string
		expected string
	}{
		{"mary-jane", "Mary-Jane"},
		{"o'brien", "O'Brien"},
		{"new york city", "New York City"},
		{"the-quick-brown-fox", "The-Quick-Brown-Fox"},
	}

	for _, tc := range testCases {
		mockConn := &mockConnectorForFormat{
			fieldValues: map[string]interface{}{
				"test_field": tc.input,
			},
		}

		input := helpers.HelperInput{
			ContactID: "123",
			Connector: mockConn,
			Config: map[string]interface{}{
				"field":  "test_field",
				"format": "title_case",
			},
		}

		_, err := helper.Execute(context.Background(), input)
		if err != nil {
			t.Fatalf("Expected no error for input '%s', got: %v", tc.input, err)
		}

		if updated, ok := mockConn.updatedFields["test_field"].(string); !ok || updated != tc.expected {
			t.Errorf("Title case for '%s': expected '%s', got '%s'", tc.input, tc.expected, updated)
		}
	}
}

// Test action logging
func TestFormatIt_Execute_ActionLogging(t *testing.T) {
	helper := &FormatIt{}

	mockConn := &mockConnectorForFormat{
		fieldValues: map[string]interface{}{
			"city": "san francisco",
		},
	}

	input := helpers.HelperInput{
		ContactID: "123",
		Connector: mockConn,
		Config: map[string]interface{}{
			"field":  "city",
			"format": "title_case",
		},
	}

	output, err := helper.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Verify action was logged
	if len(output.Actions) == 0 {
		t.Fatal("Expected actions to be logged")
	}

	action := output.Actions[0]
	if action.Type != "field_updated" {
		t.Errorf("Expected action type 'field_updated', got '%s'", action.Type)
	}
	if action.Target != "city" {
		t.Errorf("Expected action target 'city', got '%s'", action.Target)
	}
	if action.Value != "San Francisco" {
		t.Errorf("Expected action value 'San Francisco', got '%s'", action.Value)
	}

	// Verify logs
	if len(output.Logs) == 0 {
		t.Fatal("Expected logs to be generated")
	}
}
