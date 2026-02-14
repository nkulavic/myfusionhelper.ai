package contact

import (
	"context"
	"fmt"
	"testing"

	"github.com/myfusionhelper/api/internal/connectors"
	"github.com/myfusionhelper/api/internal/helpers"
)

// mockConnectorForNameParseIt mocks the CRMConnector interface for testing name_parse_it
type mockConnectorForNameParseIt struct {
	fieldValues     map[string]interface{}
	getFieldError   error
	setFieldError   error
	contact         *connectors.NormalizedContact
	getContactError error
}

func (m *mockConnectorForNameParseIt) GetContact(ctx context.Context, contactID string) (*connectors.NormalizedContact, error) {
	if m.getContactError != nil {
		return nil, m.getContactError
	}
	return m.contact, nil
}

func (m *mockConnectorForNameParseIt) SetContactFieldValue(ctx context.Context, contactID, fieldKey string, value interface{}) error {
	if m.setFieldError != nil {
		return m.setFieldError
	}
	if m.fieldValues == nil {
		m.fieldValues = make(map[string]interface{})
	}
	m.fieldValues[fieldKey] = value
	return nil
}

func (m *mockConnectorForNameParseIt) GetContactFieldValue(ctx context.Context, contactID, fieldKey string) (interface{}, error) {
	if m.getFieldError != nil {
		return nil, m.getFieldError
	}
	if m.fieldValues == nil {
		return nil, nil
	}
	return m.fieldValues[fieldKey], nil
}

// Stub implementations
func (m *mockConnectorForNameParseIt) GetContacts(ctx context.Context, opts connectors.QueryOptions) (*connectors.ContactList, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *mockConnectorForNameParseIt) CreateContact(ctx context.Context, input connectors.CreateContactInput) (*connectors.NormalizedContact, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *mockConnectorForNameParseIt) UpdateContact(ctx context.Context, contactID string, updates connectors.UpdateContactInput) (*connectors.NormalizedContact, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *mockConnectorForNameParseIt) DeleteContact(ctx context.Context, contactID string) error {
	return fmt.Errorf("not implemented")
}

func (m *mockConnectorForNameParseIt) GetTags(ctx context.Context) ([]connectors.Tag, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *mockConnectorForNameParseIt) ApplyTag(ctx context.Context, contactID, tagID string) error {
	return fmt.Errorf("not implemented")
}

func (m *mockConnectorForNameParseIt) RemoveTag(ctx context.Context, contactID, tagID string) error {
	return fmt.Errorf("not implemented")
}

func (m *mockConnectorForNameParseIt) GetCustomFields(ctx context.Context) ([]connectors.CustomField, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *mockConnectorForNameParseIt) TriggerAutomation(ctx context.Context, contactID, automationID string) error {
	return fmt.Errorf("not implemented")
}

func (m *mockConnectorForNameParseIt) AchieveGoal(ctx context.Context, contactID, goalName, integration string) error {
	return fmt.Errorf("not implemented")
}

func (m *mockConnectorForNameParseIt) TestConnection(ctx context.Context) error {
	return fmt.Errorf("not implemented")
}

func (m *mockConnectorForNameParseIt) GetMetadata() connectors.ConnectorMetadata {
	return connectors.ConnectorMetadata{PlatformSlug: "mock", PlatformName: "Mock"}
}

func (m *mockConnectorForNameParseIt) GetCapabilities() []connectors.Capability {
	return []connectors.Capability{}
}

func TestNameParseIt_GetMetadata(t *testing.T) {
	helper := &NameParseIt{}

	if helper.GetName() != "Name Parse It" {
		t.Errorf("Expected name 'Name Parse It', got '%s'", helper.GetName())
	}
	if helper.GetType() != "name_parse_it" {
		t.Errorf("Expected type 'name_parse_it', got '%s'", helper.GetType())
	}
	if helper.GetCategory() != "contact" {
		t.Errorf("Expected category 'contact', got '%s'", helper.GetCategory())
	}
	if !helper.RequiresCRM() {
		t.Error("Expected RequiresCRM to be true")
	}
}

func TestNameParseIt_ValidateConfig_Success(t *testing.T) {
	helper := &NameParseIt{}

	config := map[string]interface{}{
		"source_field": "full_name",
	}

	err := helper.ValidateConfig(config)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
}

func TestNameParseIt_ValidateConfig_MissingSourceField(t *testing.T) {
	helper := &NameParseIt{}

	config := map[string]interface{}{}

	err := helper.ValidateConfig(config)
	if err == nil {
		t.Error("Expected error for missing source_field")
	}
	if err != nil && err.Error() != "source_field is required" {
		t.Errorf("Expected 'source_field is required', got: %v", err)
	}
}

func TestNameParseIt_ValidateConfig_EmptySourceField(t *testing.T) {
	helper := &NameParseIt{}

	config := map[string]interface{}{
		"source_field": "",
	}

	err := helper.ValidateConfig(config)
	if err == nil {
		t.Error("Expected error for empty source_field")
	}
}

func TestNameParseIt_Execute_SuccessSimpleName(t *testing.T) {
	helper := &NameParseIt{}
	mock := &mockConnectorForNameParseIt{
		fieldValues: map[string]interface{}{
			"full_name": "John Doe",
		},
	}

	input := helpers.HelperInput{
		ContactID: "123",
		Config: map[string]interface{}{
			"source_field": "full_name",
		},
		Connector: mock,
	}

	output, err := helper.Execute(context.Background(), input)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if !output.Success {
		t.Error("Expected success to be true")
	}

	// Verify parsed names
	if mock.fieldValues["first_name"] != "John" {
		t.Errorf("Expected first_name to be 'John', got: %v", mock.fieldValues["first_name"])
	}
	if mock.fieldValues["last_name"] != "Doe" {
		t.Errorf("Expected last_name to be 'Doe', got: %v", mock.fieldValues["last_name"])
	}

	// Verify 2 actions (first + last)
	if len(output.Actions) != 2 {
		t.Errorf("Expected 2 actions, got %d", len(output.Actions))
	}
}

func TestNameParseIt_Execute_WithSuffix(t *testing.T) {
	helper := &NameParseIt{}
	mock := &mockConnectorForNameParseIt{
		fieldValues: map[string]interface{}{
			"full_name": "John Doe Jr.",
		},
	}

	input := helpers.HelperInput{
		ContactID: "123",
		Config: map[string]interface{}{
			"source_field": "full_name",
			"suffix_field": "suffix",
		},
		Connector: mock,
	}

	output, err := helper.Execute(context.Background(), input)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if !output.Success {
		t.Error("Expected success to be true")
	}

	// Verify parsed names
	if mock.fieldValues["first_name"] != "John" {
		t.Errorf("Expected first_name to be 'John', got: %v", mock.fieldValues["first_name"])
	}
	if mock.fieldValues["last_name"] != "Doe" {
		t.Errorf("Expected last_name to be 'Doe', got: %v", mock.fieldValues["last_name"])
	}
	if mock.fieldValues["suffix"] != "Jr." {
		t.Errorf("Expected suffix to be 'Jr.', got: %v", mock.fieldValues["suffix"])
	}

	// Verify 3 actions (first + last + suffix)
	if len(output.Actions) != 3 {
		t.Errorf("Expected 3 actions, got %d", len(output.Actions))
	}
}

func TestNameParseIt_Execute_WithMiddleName(t *testing.T) {
	helper := &NameParseIt{}
	mock := &mockConnectorForNameParseIt{
		fieldValues: map[string]interface{}{
			"full_name": "John Michael Doe",
		},
	}

	input := helpers.HelperInput{
		ContactID: "123",
		Config: map[string]interface{}{
			"source_field": "full_name",
		},
		Connector: mock,
	}

	_, err := helper.Execute(context.Background(), input)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Middle name should be included in last name
	if mock.fieldValues["first_name"] != "John" {
		t.Errorf("Expected first_name to be 'John', got: %v", mock.fieldValues["first_name"])
	}
	if mock.fieldValues["last_name"] != "Michael Doe" {
		t.Errorf("Expected last_name to be 'Michael Doe', got: %v", mock.fieldValues["last_name"])
	}
}

func TestNameParseIt_Execute_SingleName(t *testing.T) {
	helper := &NameParseIt{}
	mock := &mockConnectorForNameParseIt{
		fieldValues: map[string]interface{}{
			"full_name": "Madonna",
		},
	}

	input := helpers.HelperInput{
		ContactID: "123",
		Config: map[string]interface{}{
			"source_field": "full_name",
		},
		Connector: mock,
	}

	_, err := helper.Execute(context.Background(), input)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Single name goes to first name, last name is empty
	if mock.fieldValues["first_name"] != "Madonna" {
		t.Errorf("Expected first_name to be 'Madonna', got: %v", mock.fieldValues["first_name"])
	}
	if mock.fieldValues["last_name"] != "" {
		t.Errorf("Expected last_name to be empty, got: %v", mock.fieldValues["last_name"])
	}
}

func TestNameParseIt_Execute_EmptySourceField(t *testing.T) {
	helper := &NameParseIt{}
	mock := &mockConnectorForNameParseIt{
		fieldValues: map[string]interface{}{
			"full_name": "",
		},
	}

	input := helpers.HelperInput{
		ContactID: "123",
		Config: map[string]interface{}{
			"source_field": "full_name",
		},
		Connector: mock,
	}

	output, err := helper.Execute(context.Background(), input)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if !output.Success {
		t.Error("Expected success to be true")
	}
	if output.Message != "Field 'full_name' is empty, nothing to parse" {
		t.Errorf("Unexpected message: %s", output.Message)
	}
}

func TestNameParseIt_Execute_NilSourceField(t *testing.T) {
	helper := &NameParseIt{}
	mock := &mockConnectorForNameParseIt{
		fieldValues: map[string]interface{}{
			"full_name": nil,
		},
	}

	input := helpers.HelperInput{
		ContactID: "123",
		Config: map[string]interface{}{
			"source_field": "full_name",
		},
		Connector: mock,
	}

	output, err := helper.Execute(context.Background(), input)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if !output.Success {
		t.Error("Expected success to be true")
	}
}

func TestNameParseIt_Execute_CustomFieldNames(t *testing.T) {
	helper := &NameParseIt{}
	mock := &mockConnectorForNameParseIt{
		fieldValues: map[string]interface{}{
			"full_name": "Jane Smith",
		},
	}

	input := helpers.HelperInput{
		ContactID: "123",
		Config: map[string]interface{}{
			"source_field":      "full_name",
			"first_name_field": "given_name",
			"last_name_field":  "family_name",
		},
		Connector: mock,
	}

	_, err := helper.Execute(context.Background(), input)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Verify custom field names were used
	if mock.fieldValues["given_name"] != "Jane" {
		t.Errorf("Expected given_name to be 'Jane', got: %v", mock.fieldValues["given_name"])
	}
	if mock.fieldValues["family_name"] != "Smith" {
		t.Errorf("Expected family_name to be 'Smith', got: %v", mock.fieldValues["family_name"])
	}
}

func TestNameParseIt_Execute_GetFieldError(t *testing.T) {
	helper := &NameParseIt{}
	mock := &mockConnectorForNameParseIt{
		getFieldError: fmt.Errorf("field read error"),
	}

	input := helpers.HelperInput{
		ContactID: "123",
		Config: map[string]interface{}{
			"source_field": "full_name",
		},
		Connector: mock,
	}

	output, err := helper.Execute(context.Background(), input)
	if err == nil {
		t.Error("Expected error for GetContactFieldValue failure")
	}
	if output.Success {
		t.Error("Expected success to be false")
	}
}

func TestNameParseIt_Execute_WithMultipleSuffixes(t *testing.T) {
	helper := &NameParseIt{}
	mock := &mockConnectorForNameParseIt{
		fieldValues: map[string]interface{}{
			"full_name": "Dr. Robert Smith III",
		},
	}

	input := helpers.HelperInput{
		ContactID: "123",
		Config: map[string]interface{}{
			"source_field": "full_name",
			"suffix_field": "suffix",
		},
		Connector: mock,
	}

	_, err := helper.Execute(context.Background(), input)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// "Dr." is not recognized as a suffix, "III" should be parsed
	if mock.fieldValues["first_name"] != "Dr." {
		t.Errorf("Expected first_name to be 'Dr.', got: %v", mock.fieldValues["first_name"])
	}
	if mock.fieldValues["last_name"] != "Robert Smith" {
		t.Errorf("Expected last_name to be 'Robert Smith', got: %v", mock.fieldValues["last_name"])
	}
	if mock.fieldValues["suffix"] != "III" {
		t.Errorf("Expected suffix to be 'III', got: %v", mock.fieldValues["suffix"])
	}
}

func TestNameParseIt_Execute_SuffixWithoutFieldConfig(t *testing.T) {
	helper := &NameParseIt{}
	mock := &mockConnectorForNameParseIt{
		fieldValues: map[string]interface{}{
			"full_name": "John Doe Sr.",
		},
	}

	input := helpers.HelperInput{
		ContactID: "123",
		Config: map[string]interface{}{
			"source_field": "full_name",
			// No suffix_field configured
		},
		Connector: mock,
	}

	output, err := helper.Execute(context.Background(), input)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Suffix is still parsed but not stored
	if mock.fieldValues["first_name"] != "John" {
		t.Errorf("Expected first_name to be 'John', got: %v", mock.fieldValues["first_name"])
	}
	if mock.fieldValues["last_name"] != "Doe" {
		t.Errorf("Expected last_name to be 'Doe', got: %v", mock.fieldValues["last_name"])
	}
	// Should only have 2 actions (no suffix action)
	if len(output.Actions) != 2 {
		t.Errorf("Expected 2 actions (no suffix), got %d", len(output.Actions))
	}
}

func TestNameParseIt_Execute_WhitespaceHandling(t *testing.T) {
	helper := &NameParseIt{}
	mock := &mockConnectorForNameParseIt{
		fieldValues: map[string]interface{}{
			"full_name": "  John   Doe  ",
		},
	}

	input := helpers.HelperInput{
		ContactID: "123",
		Config: map[string]interface{}{
			"source_field": "full_name",
		},
		Connector: mock,
	}

	_, err := helper.Execute(context.Background(), input)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Extra whitespace should be normalized
	if mock.fieldValues["first_name"] != "John" {
		t.Errorf("Expected first_name to be 'John', got: %v", mock.fieldValues["first_name"])
	}
	if mock.fieldValues["last_name"] != "Doe" {
		t.Errorf("Expected last_name to be 'Doe', got: %v", mock.fieldValues["last_name"])
	}
}
