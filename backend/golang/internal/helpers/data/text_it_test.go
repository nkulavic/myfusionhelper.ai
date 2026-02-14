package data

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/myfusionhelper/api/internal/connectors"
	"github.com/myfusionhelper/api/internal/helpers"
)

// Mock connector for text_it testing
type mockConnectorForText struct {
	fieldValues   map[string]interface{}
	updatedFields map[string]interface{}
}

func (m *mockConnectorForText) GetContactFieldValue(ctx context.Context, contactID, fieldKey string) (interface{}, error) {
	if m.fieldValues == nil {
		return nil, nil
	}
	val, ok := m.fieldValues[fieldKey]
	if !ok {
		return nil, nil
	}
	return val, nil
}

func (m *mockConnectorForText) SetContactFieldValue(ctx context.Context, contactID, fieldKey string, value interface{}) error {
	if m.updatedFields == nil {
		m.updatedFields = make(map[string]interface{})
	}
	m.updatedFields[fieldKey] = value
	return nil
}

// Stub implementations for CRMConnector interface
func (m *mockConnectorForText) GetContact(ctx context.Context, contactID string) (*connectors.NormalizedContact, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForText) GetContacts(ctx context.Context, opts connectors.QueryOptions) (*connectors.ContactList, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForText) CreateContact(ctx context.Context, input connectors.CreateContactInput) (*connectors.NormalizedContact, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForText) UpdateContact(ctx context.Context, contactID string, updates connectors.UpdateContactInput) (*connectors.NormalizedContact, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForText) DeleteContact(ctx context.Context, contactID string) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorForText) GetTags(ctx context.Context) ([]connectors.Tag, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForText) ApplyTag(ctx context.Context, contactID, tagID string) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorForText) RemoveTag(ctx context.Context, contactID, tagID string) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorForText) GetCustomFields(ctx context.Context) ([]connectors.CustomField, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForText) TriggerAutomation(ctx context.Context, contactID, automationID string) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorForText) AchieveGoal(ctx context.Context, contactID, goalName, integration string) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorForText) TestConnection(ctx context.Context) error {
	return nil
}
func (m *mockConnectorForText) GetMetadata() connectors.ConnectorMetadata {
	return connectors.ConnectorMetadata{}
}
func (m *mockConnectorForText) GetCapabilities() []connectors.Capability {
	return nil
}

// Test helper metadata
func TestTextIt_Metadata(t *testing.T) {
	helper := &TextIt{}

	if helper.GetName() != "Text It" {
		t.Errorf("Expected name 'Text It', got '%s'", helper.GetName())
	}
	if helper.GetType() != "text_it" {
		t.Errorf("Expected type 'text_it', got '%s'", helper.GetType())
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
func TestTextIt_ValidateConfig_MissingField(t *testing.T) {
	helper := &TextIt{}
	config := map[string]interface{}{
		"operation": "prepend",
		"value":     "Hello ",
	}

	err := helper.ValidateConfig(config)
	if err == nil {
		t.Error("Expected validation error for missing field")
	}
	if !strings.Contains(err.Error(), "field") {
		t.Errorf("Expected error about field, got: %v", err)
	}
}

// Test validation - missing operation
func TestTextIt_ValidateConfig_MissingOperation(t *testing.T) {
	helper := &TextIt{}
	config := map[string]interface{}{
		"field": "name",
		"value": "Hello ",
	}

	err := helper.ValidateConfig(config)
	if err == nil {
		t.Error("Expected validation error for missing operation")
	}
	if !strings.Contains(err.Error(), "operation") {
		t.Errorf("Expected error about operation, got: %v", err)
	}
}

// Test validation - invalid operation
func TestTextIt_ValidateConfig_InvalidOperation(t *testing.T) {
	helper := &TextIt{}
	config := map[string]interface{}{
		"field":     "name",
		"operation": "invalid_op",
	}

	err := helper.ValidateConfig(config)
	if err == nil {
		t.Error("Expected validation error for invalid operation")
	}
	if !strings.Contains(err.Error(), "invalid operation") {
		t.Errorf("Expected error about invalid operation, got: %v", err)
	}
}

// Test validation - missing value for prepend
func TestTextIt_ValidateConfig_MissingValueForPrepend(t *testing.T) {
	helper := &TextIt{}
	config := map[string]interface{}{
		"field":     "name",
		"operation": "prepend",
	}

	err := helper.ValidateConfig(config)
	if err == nil {
		t.Error("Expected validation error for missing value")
	}
	if !strings.Contains(err.Error(), "value") {
		t.Errorf("Expected error about value, got: %v", err)
	}
}

// Test validation - missing max_length for truncate
func TestTextIt_ValidateConfig_MissingMaxLengthForTruncate(t *testing.T) {
	helper := &TextIt{}
	config := map[string]interface{}{
		"field":     "name",
		"operation": "truncate",
	}

	err := helper.ValidateConfig(config)
	if err == nil {
		t.Error("Expected validation error for missing max_length")
	}
	if !strings.Contains(err.Error(), "max_length") {
		t.Errorf("Expected error about max_length, got: %v", err)
	}
}

// Test validation - valid configs
func TestTextIt_ValidateConfig_Valid(t *testing.T) {
	helper := &TextIt{}

	validConfigs := []map[string]interface{}{
		{
			"field":     "name",
			"operation": "prepend",
			"value":     "Hello ",
		},
		{
			"field":     "name",
			"operation": "append",
			"value":     " Jr.",
		},
		{
			"field":     "name",
			"operation": "truncate",
			"max_length": 10,
		},
		{
			"field":     "email",
			"operation": "extract_email_domain",
		},
		{
			"field":     "phone",
			"operation": "extract_numbers",
		},
		{
			"field":     "title",
			"operation": "slug",
		},
		{
			"field":     "name",
			"operation": "reverse",
		},
	}

	for _, config := range validConfigs {
		err := helper.ValidateConfig(config)
		if err != nil {
			t.Errorf("Expected no validation error for %v, got: %v", config, err)
		}
	}
}

// Test execution - prepend operation
func TestTextIt_Execute_Prepend(t *testing.T) {
	helper := &TextIt{}

	mockConn := &mockConnectorForText{
		fieldValues: map[string]interface{}{
			"name": "John",
		},
	}

	input := helpers.HelperInput{
		ContactID: "123",
		Connector: mockConn,
		Config: map[string]interface{}{
			"field":     "name",
			"operation": "prepend",
			"value":     "Dr. ",
		},
	}

	output, err := helper.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if !output.Success {
		t.Error("Expected success=true")
	}

	result := mockConn.updatedFields["name"]
	if result != "Dr. John" {
		t.Errorf("Expected 'Dr. John', got: %v", result)
	}
}

// Test execution - append operation
func TestTextIt_Execute_Append(t *testing.T) {
	helper := &TextIt{}

	mockConn := &mockConnectorForText{
		fieldValues: map[string]interface{}{
			"name": "John Smith",
		},
	}

	input := helpers.HelperInput{
		ContactID: "123",
		Connector: mockConn,
		Config: map[string]interface{}{
			"field":     "name",
			"operation": "append",
			"value":     ", Jr.",
		},
	}

	output, err := helper.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if !output.Success {
		t.Error("Expected success=true")
	}

	result := mockConn.updatedFields["name"]
	if result != "John Smith, Jr." {
		t.Errorf("Expected 'John Smith, Jr.', got: %v", result)
	}
}

// Test execution - replace operation
func TestTextIt_Execute_Replace(t *testing.T) {
	helper := &TextIt{}

	mockConn := &mockConnectorForText{
		fieldValues: map[string]interface{}{
			"message": "Hello World! Hello Universe!",
		},
	}

	input := helpers.HelperInput{
		ContactID: "123",
		Connector: mockConn,
		Config: map[string]interface{}{
			"field":        "message",
			"operation":    "replace",
			"value":        "Hello",
			"replace_with": "Hi",
		},
	}

	output, err := helper.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if !output.Success {
		t.Error("Expected success=true")
	}

	result := mockConn.updatedFields["message"]
	if result != "Hi World! Hi Universe!" {
		t.Errorf("Expected 'Hi World! Hi Universe!', got: %v", result)
	}
}

// Test execution - remove operation
func TestTextIt_Execute_Remove(t *testing.T) {
	helper := &TextIt{}

	mockConn := &mockConnectorForText{
		fieldValues: map[string]interface{}{
			"text": "Hello-World-Test",
		},
	}

	input := helpers.HelperInput{
		ContactID: "123",
		Connector: mockConn,
		Config: map[string]interface{}{
			"field":     "text",
			"operation": "remove",
			"value":     "-",
		},
	}

	output, err := helper.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if !output.Success {
		t.Error("Expected success=true")
	}

	result := mockConn.updatedFields["text"]
	if result != "HelloWorldTest" {
		t.Errorf("Expected 'HelloWorldTest', got: %v", result)
	}
}

// Test execution - truncate operation
func TestTextIt_Execute_Truncate(t *testing.T) {
	helper := &TextIt{}

	mockConn := &mockConnectorForText{
		fieldValues: map[string]interface{}{
			"description": "This is a very long description that needs to be truncated",
		},
	}

	input := helpers.HelperInput{
		ContactID: "123",
		Connector: mockConn,
		Config: map[string]interface{}{
			"field":      "description",
			"operation":  "truncate",
			"max_length": 20,
		},
	}

	output, err := helper.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if !output.Success {
		t.Error("Expected success=true")
	}

	result := mockConn.updatedFields["description"]
	if result != "This is a very long " {
		t.Errorf("Expected 'This is a very long ', got: %v", result)
	}
}

// Test execution - truncate no change (already short enough)
func TestTextIt_Execute_TruncateNoChange(t *testing.T) {
	helper := &TextIt{}

	mockConn := &mockConnectorForText{
		fieldValues: map[string]interface{}{
			"description": "Short",
		},
	}

	input := helpers.HelperInput{
		ContactID: "123",
		Connector: mockConn,
		Config: map[string]interface{}{
			"field":      "description",
			"operation":  "truncate",
			"max_length": 20,
		},
	}

	output, err := helper.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if !output.Success {
		t.Error("Expected success=true")
	}

	if output.Message != "Value unchanged, no update needed" {
		t.Errorf("Expected 'Value unchanged' message, got: %s", output.Message)
	}

	// Should not have updated the field
	if _, ok := mockConn.updatedFields["description"]; ok {
		t.Error("Should not update field when value is unchanged")
	}
}

// Test execution - extract_email_domain operation
func TestTextIt_Execute_ExtractEmailDomain(t *testing.T) {
	helper := &TextIt{}

	mockConn := &mockConnectorForText{
		fieldValues: map[string]interface{}{
			"email": "john.doe@example.com",
		},
	}

	input := helpers.HelperInput{
		ContactID: "123",
		Connector: mockConn,
		Config: map[string]interface{}{
			"field":        "email",
			"operation":    "extract_email_domain",
			"target_field": "email_domain",
		},
	}

	output, err := helper.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if !output.Success {
		t.Error("Expected success=true")
	}

	result := mockConn.updatedFields["email_domain"]
	if result != "example.com" {
		t.Errorf("Expected 'example.com', got: %v", result)
	}
}

// Test execution - extract_numbers operation
func TestTextIt_Execute_ExtractNumbers(t *testing.T) {
	helper := &TextIt{}

	mockConn := &mockConnectorForText{
		fieldValues: map[string]interface{}{
			"phone": "Call me at (555) 123-4567 ext. 890",
		},
	}

	input := helpers.HelperInput{
		ContactID: "123",
		Connector: mockConn,
		Config: map[string]interface{}{
			"field":        "phone",
			"operation":    "extract_numbers",
			"target_field": "phone_digits",
		},
	}

	output, err := helper.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if !output.Success {
		t.Error("Expected success=true")
	}

	result := mockConn.updatedFields["phone_digits"]
	// extract_numbers regex includes periods, so "ext. 890" becomes "890" as separate match
	// Result is joined: "555" + "123" + "4567" + "890" with periods preserved in matches
	if result != "5551234567.890" {
		t.Errorf("Expected '5551234567.890', got: %v", result)
	}
}

// Test execution - slug operation
func TestTextIt_Execute_Slug(t *testing.T) {
	helper := &TextIt{}

	mockConn := &mockConnectorForText{
		fieldValues: map[string]interface{}{
			"title": "Hello World! This is a TEST!!!",
		},
	}

	input := helpers.HelperInput{
		ContactID: "123",
		Connector: mockConn,
		Config: map[string]interface{}{
			"field":     "title",
			"operation": "slug",
		},
	}

	output, err := helper.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if !output.Success {
		t.Error("Expected success=true")
	}

	result := mockConn.updatedFields["title"]
	if result != "hello-world-this-is-a-test" {
		t.Errorf("Expected 'hello-world-this-is-a-test', got: %v", result)
	}
}

// Test execution - reverse operation
func TestTextIt_Execute_Reverse(t *testing.T) {
	helper := &TextIt{}

	mockConn := &mockConnectorForText{
		fieldValues: map[string]interface{}{
			"text": "Hello",
		},
	}

	input := helpers.HelperInput{
		ContactID: "123",
		Connector: mockConn,
		Config: map[string]interface{}{
			"field":     "text",
			"operation": "reverse",
		},
	}

	output, err := helper.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if !output.Success {
		t.Error("Expected success=true")
	}

	result := mockConn.updatedFields["text"]
	if result != "olleH" {
		t.Errorf("Expected 'olleH', got: %v", result)
	}
}

// Test execution - target_field usage
func TestTextIt_Execute_TargetField(t *testing.T) {
	helper := &TextIt{}

	mockConn := &mockConnectorForText{
		fieldValues: map[string]interface{}{
			"first_name": "John",
		},
	}

	input := helpers.HelperInput{
		ContactID: "123",
		Connector: mockConn,
		Config: map[string]interface{}{
			"field":        "first_name",
			"operation":    "prepend",
			"value":        "Mr. ",
			"target_field": "formal_name",
		},
	}

	output, err := helper.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if !output.Success {
		t.Error("Expected success=true")
	}

	// Original field should not be updated
	if _, ok := mockConn.updatedFields["first_name"]; ok {
		t.Error("Should not update source field when target_field is specified")
	}

	// Target field should be updated
	result := mockConn.updatedFields["formal_name"]
	if result != "Mr. John" {
		t.Errorf("Expected 'Mr. John' in formal_name, got: %v", result)
	}
}

// Test execution - nil value handling
func TestTextIt_Execute_NilValue(t *testing.T) {
	helper := &TextIt{}

	mockConn := &mockConnectorForText{
		fieldValues: map[string]interface{}{
			"name": nil,
		},
	}

	input := helpers.HelperInput{
		ContactID: "123",
		Connector: mockConn,
		Config: map[string]interface{}{
			"field":     "name",
			"operation": "append",
			"value":     " (Unknown)",
		},
	}

	output, err := helper.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if !output.Success {
		t.Error("Expected success=true")
	}

	result := mockConn.updatedFields["name"]
	if result != " (Unknown)" {
		t.Errorf("Expected ' (Unknown)', got: %v", result)
	}
}

// Test action logging
func TestTextIt_Execute_ActionLogging(t *testing.T) {
	helper := &TextIt{}

	mockConn := &mockConnectorForText{
		fieldValues: map[string]interface{}{
			"name": "John",
		},
	}

	input := helpers.HelperInput{
		ContactID: "123",
		Connector: mockConn,
		Config: map[string]interface{}{
			"field":     "name",
			"operation": "prepend",
			"value":     "Dr. ",
		},
	}

	output, err := helper.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Verify action was logged
	if len(output.Actions) != 1 {
		t.Fatalf("Expected 1 action, got %d", len(output.Actions))
	}

	action := output.Actions[0]
	if action.Type != "field_updated" {
		t.Errorf("Expected action type 'field_updated', got '%s'", action.Type)
	}
	if action.Target != "name" {
		t.Errorf("Expected action target 'name', got '%s'", action.Target)
	}

	// Verify logs
	if len(output.Logs) == 0 {
		t.Error("Expected logs to be generated")
	}

	// Verify modified data
	if output.ModifiedData == nil {
		t.Fatal("Expected ModifiedData to be set")
	}
	if output.ModifiedData["name"] != "Dr. John" {
		t.Errorf("Expected ModifiedData['name'] = 'Dr. John', got: %v", output.ModifiedData["name"])
	}
}
