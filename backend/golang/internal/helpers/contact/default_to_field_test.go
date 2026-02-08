package contact

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/myfusionhelper/api/internal/connectors"
	"github.com/myfusionhelper/api/internal/helpers"
)

type mockConnectorForDefaultToField struct {
	contact         *connectors.NormalizedContact
	getContactError error
	fieldsSet       map[string]interface{}
	setFieldError   error
}

func (m *mockConnectorForDefaultToField) GetContact(ctx context.Context, contactID string) (*connectors.NormalizedContact, error) {
	if m.getContactError != nil {
		return nil, m.getContactError
	}
	return m.contact, nil
}

func (m *mockConnectorForDefaultToField) SetContactFieldValue(ctx context.Context, contactID, fieldKey string, value interface{}) error {
	if m.setFieldError != nil {
		return m.setFieldError
	}
	if m.fieldsSet == nil {
		m.fieldsSet = make(map[string]interface{})
	}
	m.fieldsSet[fieldKey] = value
	return nil
}

func (m *mockConnectorForDefaultToField) GetContacts(ctx context.Context, opts connectors.QueryOptions) (*connectors.ContactList, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForDefaultToField) CreateContact(ctx context.Context, contact connectors.CreateContactInput) (*connectors.NormalizedContact, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForDefaultToField) UpdateContact(ctx context.Context, contactID string, updates connectors.UpdateContactInput) (*connectors.NormalizedContact, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForDefaultToField) DeleteContact(ctx context.Context, contactID string) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorForDefaultToField) GetTags(ctx context.Context) ([]connectors.Tag, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForDefaultToField) ApplyTag(ctx context.Context, contactID, tagID string) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorForDefaultToField) RemoveTag(ctx context.Context, contactID, tagID string) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorForDefaultToField) GetCustomFields(ctx context.Context) ([]connectors.CustomField, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForDefaultToField) GetContactFieldValue(ctx context.Context, contactID, fieldKey string) (interface{}, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForDefaultToField) TriggerAutomation(ctx context.Context, contactID, automationID string) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorForDefaultToField) AchieveGoal(ctx context.Context, contactID, goalName, integration string) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorForDefaultToField) TestConnection(ctx context.Context) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorForDefaultToField) GetMetadata() connectors.ConnectorMetadata {
	return connectors.ConnectorMetadata{}
}
func (m *mockConnectorForDefaultToField) GetCapabilities() []connectors.Capability {
	return nil
}

func TestDefaultToField_GetMetadata(t *testing.T) {
	h := &DefaultToField{}
	if h.GetName() != "Default To Field" {
		t.Error("wrong name")
	}
	if h.GetType() != "default_to_field" {
		t.Error("wrong type")
	}
	if h.GetCategory() != "contact" {
		t.Error("wrong category")
	}
	if !h.RequiresCRM() {
		t.Error("should require CRM")
	}
}

func TestDefaultToField_ValidateConfig_MissingDefault(t *testing.T) {
	err := (&DefaultToField{}).ValidateConfig(map[string]interface{}{
		"to_field": "Status",
	})
	if err == nil {
		t.Error("should error on missing default")
	}
}

func TestDefaultToField_ValidateConfig_MissingToField(t *testing.T) {
	err := (&DefaultToField{}).ValidateConfig(map[string]interface{}{
		"default": "active",
	})
	if err == nil {
		t.Error("should error on missing to_field")
	}
}

func TestDefaultToField_ValidateConfig_EmptyDefault(t *testing.T) {
	err := (&DefaultToField{}).ValidateConfig(map[string]interface{}{
		"default":  "",
		"to_field": "Status",
	})
	if err == nil {
		t.Error("should error on empty default")
	}
}

func TestDefaultToField_ValidateConfig_Valid(t *testing.T) {
	err := (&DefaultToField{}).ValidateConfig(map[string]interface{}{
		"default":  "active",
		"to_field": "Status",
	})
	if err != nil {
		t.Errorf("should be valid: %v", err)
	}
}

func TestDefaultToField_Execute_StaticText(t *testing.T) {
	mock := &mockConnectorForDefaultToField{}
	output, err := (&DefaultToField{}).Execute(context.Background(), helpers.HelperInput{
		ContactID: "123",
		Config: map[string]interface{}{
			"default":  "active",
			"to_field": "Status",
		},
		Connector: mock,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !output.Success {
		t.Error("should succeed")
	}
	if mock.fieldsSet["Status"] != "active" {
		t.Errorf("expected Status=active, got %v", mock.fieldsSet["Status"])
	}
	if len(output.Actions) != 1 {
		t.Errorf("expected 1 action, got %d", len(output.Actions))
	}
	if output.Actions[0].Type != "field_updated" || output.Actions[0].Target != "Status" {
		t.Error("wrong action")
	}
}

func TestDefaultToField_Execute_DateMacroNow(t *testing.T) {
	mock := &mockConnectorForDefaultToField{}
	output, err := (&DefaultToField{}).Execute(context.Background(), helpers.HelperInput{
		ContactID: "123",
		Config: map[string]interface{}{
			"default":  "@date_now",
			"to_field": "LastUpdated",
		},
		Connector: mock,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !output.Success {
		t.Error("should succeed")
	}
	result, ok := mock.fieldsSet["LastUpdated"].(string)
	if !ok {
		t.Fatal("LastUpdated should be a string")
	}
	// Should be in format "2006-01-02 15:04:05"
	_, parseErr := time.Parse("2006-01-02 15:04:05", result)
	if parseErr != nil {
		t.Errorf("expected timestamp format, got %s: %v", result, parseErr)
	}
}

func TestDefaultToField_Execute_DateMacroToday(t *testing.T) {
	mock := &mockConnectorForDefaultToField{}
	output, err := (&DefaultToField{}).Execute(context.Background(), helpers.HelperInput{
		ContactID: "123",
		Config: map[string]interface{}{
			"default":  "@date_today",
			"to_field": "DateJoined",
		},
		Connector: mock,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !output.Success {
		t.Error("should succeed")
	}
	result, ok := mock.fieldsSet["DateJoined"].(string)
	if !ok {
		t.Fatal("DateJoined should be a string")
	}
	// Should be in format "2006-01-02"
	_, parseErr := time.Parse("2006-01-02", result)
	if parseErr != nil {
		t.Errorf("expected date format, got %s: %v", result, parseErr)
	}
}

func TestDefaultToField_Execute_KeywordNow(t *testing.T) {
	mock := &mockConnectorForDefaultToField{}
	output, err := (&DefaultToField{}).Execute(context.Background(), helpers.HelperInput{
		ContactID: "123",
		Config: map[string]interface{}{
			"default":  "now",
			"to_field": "Timestamp",
		},
		Connector: mock,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !output.Success {
		t.Error("should succeed")
	}
	result, ok := mock.fieldsSet["Timestamp"].(string)
	if !ok {
		t.Fatal("Timestamp should be a string")
	}
	_, parseErr := time.Parse("2006-01-02 15:04:05", result)
	if parseErr != nil {
		t.Errorf("expected timestamp format, got %s: %v", result, parseErr)
	}
}

func TestDefaultToField_Execute_KeywordToday(t *testing.T) {
	mock := &mockConnectorForDefaultToField{}
	output, err := (&DefaultToField{}).Execute(context.Background(), helpers.HelperInput{
		ContactID: "123",
		Config: map[string]interface{}{
			"default":  "today",
			"to_field": "Date",
		},
		Connector: mock,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !output.Success {
		t.Error("should succeed")
	}
	result, ok := mock.fieldsSet["Date"].(string)
	if !ok {
		t.Fatal("Date should be a string")
	}
	_, parseErr := time.Parse("2006-01-02", result)
	if parseErr != nil {
		t.Errorf("expected date format, got %s: %v", result, parseErr)
	}
}

func TestDefaultToField_Execute_MergeFieldFirstName(t *testing.T) {
	mock := &mockConnectorForDefaultToField{
		contact: &connectors.NormalizedContact{
			ID:        "123",
			FirstName: "John",
			LastName:  "Doe",
			Email:     "john@example.com",
		},
	}
	output, err := (&DefaultToField{}).Execute(context.Background(), helpers.HelperInput{
		ContactID: "123",
		Config: map[string]interface{}{
			"default":  "Hello @FirstName!",
			"to_field": "Greeting",
		},
		Connector: mock,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !output.Success {
		t.Error("should succeed")
	}
	if mock.fieldsSet["Greeting"] != "Hello John!" {
		t.Errorf("expected 'Hello John!', got %v", mock.fieldsSet["Greeting"])
	}
}

func TestDefaultToField_Execute_MergeFieldMultiple(t *testing.T) {
	mock := &mockConnectorForDefaultToField{
		contact: &connectors.NormalizedContact{
			ID:        "123",
			FirstName: "Jane",
			LastName:  "Smith",
			Email:     "jane@example.com",
			Company:   "ACME Corp",
		},
	}
	output, err := (&DefaultToField{}).Execute(context.Background(), helpers.HelperInput{
		ContactID: "123",
		Config: map[string]interface{}{
			"default":  "@FirstName @LastName from @Company",
			"to_field": "FullProfile",
		},
		Connector: mock,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !output.Success {
		t.Error("should succeed")
	}
	if mock.fieldsSet["FullProfile"] != "Jane Smith from ACME Corp" {
		t.Errorf("expected 'Jane Smith from ACME Corp', got %v", mock.fieldsSet["FullProfile"])
	}
}

func TestDefaultToField_Execute_MergeFieldCustomField(t *testing.T) {
	mock := &mockConnectorForDefaultToField{
		contact: &connectors.NormalizedContact{
			ID:        "123",
			FirstName: "Bob",
			CustomFields: map[string]interface{}{
				"VIPLevel": "Gold",
			},
		},
	}
	output, err := (&DefaultToField{}).Execute(context.Background(), helpers.HelperInput{
		ContactID: "123",
		Config: map[string]interface{}{
			"default":  "Welcome @FirstName, your level is @VIPLevel",
			"to_field": "WelcomeMessage",
		},
		Connector: mock,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !output.Success {
		t.Error("should succeed")
	}
	if mock.fieldsSet["WelcomeMessage"] != "Welcome Bob, your level is Gold" {
		t.Errorf("expected 'Welcome Bob, your level is Gold', got %v", mock.fieldsSet["WelcomeMessage"])
	}
}

func TestDefaultToField_Execute_GetContactError(t *testing.T) {
	mock := &mockConnectorForDefaultToField{
		getContactError: fmt.Errorf("API error"),
	}
	output, err := (&DefaultToField{}).Execute(context.Background(), helpers.HelperInput{
		ContactID: "123",
		Config: map[string]interface{}{
			"default":  "@FirstName test",
			"to_field": "Message",
		},
		Connector: mock,
	})
	if err != nil {
		t.Fatal(err)
	}
	// Should succeed but log warning and use original text with unreplaced merge field
	if !output.Success {
		t.Error("should succeed even if GetContact fails")
	}
	// Value should still be set, just with unreplaced merge field
	result, ok := mock.fieldsSet["Message"].(string)
	if !ok {
		t.Fatal("Message should be set")
	}
	if !strings.Contains(result, "@FirstName") {
		t.Error("should contain unreplaced merge field")
	}
}

func TestDefaultToField_Execute_EmptyAfterResolution(t *testing.T) {
	mock := &mockConnectorForDefaultToField{
		contact: &connectors.NormalizedContact{
			ID:        "123",
			FirstName: "",
		},
	}
	output, err := (&DefaultToField{}).Execute(context.Background(), helpers.HelperInput{
		ContactID: "123",
		Config: map[string]interface{}{
			"default":  "@FirstName",
			"to_field": "Name",
		},
		Connector: mock,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !output.Success {
		t.Error("should succeed")
	}
	if output.Message != "Default value resolved to empty, nothing to set" {
		t.Errorf("expected empty resolution message, got %s", output.Message)
	}
	if len(mock.fieldsSet) > 0 {
		t.Error("should not set field when value resolves to empty")
	}
}

func TestDefaultToField_Execute_SetFieldError(t *testing.T) {
	mock := &mockConnectorForDefaultToField{
		setFieldError: fmt.Errorf("field update failed"),
	}
	output, err := (&DefaultToField{}).Execute(context.Background(), helpers.HelperInput{
		ContactID: "123",
		Config: map[string]interface{}{
			"default":  "test value",
			"to_field": "Status",
		},
		Connector: mock,
	})
	if err == nil {
		t.Error("should return error on SetContactFieldValue failure")
	}
	if output.Success {
		t.Error("should not succeed")
	}
}

func TestDefaultToField_Execute_CombinedMacrosAndMerge(t *testing.T) {
	mock := &mockConnectorForDefaultToField{
		contact: &connectors.NormalizedContact{
			ID:        "123",
			FirstName: "Alice",
		},
	}
	output, err := (&DefaultToField{}).Execute(context.Background(), helpers.HelperInput{
		ContactID: "123",
		Config: map[string]interface{}{
			"default":  "Hello @FirstName, today is @date_today",
			"to_field": "Message",
		},
		Connector: mock,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !output.Success {
		t.Error("should succeed")
	}
	result, ok := mock.fieldsSet["Message"].(string)
	if !ok {
		t.Fatal("Message should be set")
	}
	if !strings.HasPrefix(result, "Hello Alice, today is ") {
		t.Errorf("expected prefix 'Hello Alice, today is ', got %s", result)
	}
	// Extract date portion and verify it's a valid date
	datePart := strings.TrimPrefix(result, "Hello Alice, today is ")
	_, parseErr := time.Parse("2006-01-02", datePart)
	if parseErr != nil {
		t.Errorf("date portion should be valid: %v", parseErr)
	}
}
