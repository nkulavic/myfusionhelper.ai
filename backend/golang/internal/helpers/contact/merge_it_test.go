package contact

import (
	"context"
	"fmt"
	"testing"

	"github.com/myfusionhelper/api/internal/connectors"
	"github.com/myfusionhelper/api/internal/helpers"
)

type mockConnectorForMergeIt struct {
	fieldValues   map[string]interface{}
	getFieldError map[string]error
	setFieldError error
}

func (m *mockConnectorForMergeIt) GetContactFieldValue(ctx context.Context, contactID, fieldKey string) (interface{}, error) {
	if m.getFieldError != nil {
		if err, ok := m.getFieldError[fieldKey]; ok {
			return nil, err
		}
	}
	if m.fieldValues != nil {
		return m.fieldValues[fieldKey], nil
	}
	return nil, nil
}

func (m *mockConnectorForMergeIt) SetContactFieldValue(ctx context.Context, contactID, fieldKey string, value interface{}) error {
	if m.setFieldError != nil {
		return m.setFieldError
	}
	if m.fieldValues == nil {
		m.fieldValues = make(map[string]interface{})
	}
	m.fieldValues[fieldKey] = value
	return nil
}

func (m *mockConnectorForMergeIt) GetContacts(ctx context.Context, opts connectors.QueryOptions) (*connectors.ContactList, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForMergeIt) GetContact(ctx context.Context, contactID string) (*connectors.NormalizedContact, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForMergeIt) CreateContact(ctx context.Context, contact connectors.CreateContactInput) (*connectors.NormalizedContact, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForMergeIt) UpdateContact(ctx context.Context, contactID string, updates connectors.UpdateContactInput) (*connectors.NormalizedContact, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForMergeIt) DeleteContact(ctx context.Context, contactID string) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorForMergeIt) GetTags(ctx context.Context) ([]connectors.Tag, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForMergeIt) ApplyTag(ctx context.Context, contactID, tagID string) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorForMergeIt) RemoveTag(ctx context.Context, contactID, tagID string) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorForMergeIt) GetCustomFields(ctx context.Context) ([]connectors.CustomField, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForMergeIt) TriggerAutomation(ctx context.Context, contactID, automationID string) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorForMergeIt) AchieveGoal(ctx context.Context, contactID, goalName, integration string) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorForMergeIt) TestConnection(ctx context.Context) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorForMergeIt) GetMetadata() connectors.ConnectorMetadata {
	return connectors.ConnectorMetadata{}
}
func (m *mockConnectorForMergeIt) GetCapabilities() []connectors.Capability {
	return nil
}

func TestMergeIt_GetMetadata(t *testing.T) {
	h := &MergeIt{}
	if h.GetName() != "Merge It" {
		t.Error("wrong name")
	}
	if h.GetType() != "merge_it" {
		t.Error("wrong type")
	}
	if h.GetCategory() != "contact" {
		t.Error("wrong category")
	}
	if !h.RequiresCRM() {
		t.Error("should require CRM")
	}
}

func TestMergeIt_ValidateConfig_MissingSourceFields(t *testing.T) {
	err := (&MergeIt{}).ValidateConfig(map[string]interface{}{
		"target_field": "FullName",
	})
	if err == nil {
		t.Error("should error on missing source_fields")
	}
}

func TestMergeIt_ValidateConfig_MissingTargetField(t *testing.T) {
	err := (&MergeIt{}).ValidateConfig(map[string]interface{}{
		"source_fields": []interface{}{"FirstName", "LastName"},
	})
	if err == nil {
		t.Error("should error on missing target_field")
	}
}

func TestMergeIt_ValidateConfig_TooFewSourceFields(t *testing.T) {
	err := (&MergeIt{}).ValidateConfig(map[string]interface{}{
		"source_fields": []interface{}{"FirstName"},
		"target_field":  "FullName",
	})
	if err == nil {
		t.Error("should error when source_fields has less than 2 fields")
	}
}

func TestMergeIt_ValidateConfig_EmptyTargetField(t *testing.T) {
	err := (&MergeIt{}).ValidateConfig(map[string]interface{}{
		"source_fields": []interface{}{"FirstName", "LastName"},
		"target_field":  "",
	})
	if err == nil {
		t.Error("should error on empty target_field")
	}
}

func TestMergeIt_ValidateConfig_Valid(t *testing.T) {
	err := (&MergeIt{}).ValidateConfig(map[string]interface{}{
		"source_fields": []interface{}{"FirstName", "LastName"},
		"target_field":  "FullName",
	})
	if err != nil {
		t.Errorf("should be valid: %v", err)
	}
}

func TestMergeIt_Execute_TwoFields(t *testing.T) {
	mock := &mockConnectorForMergeIt{
		fieldValues: map[string]interface{}{
			"FirstName": "John",
			"LastName":  "Doe",
		},
	}
	output, err := (&MergeIt{}).Execute(context.Background(), helpers.HelperInput{
		ContactID: "123",
		Config: map[string]interface{}{
			"source_fields": []interface{}{"FirstName", "LastName"},
			"target_field":  "FullName",
		},
		Connector: mock,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !output.Success {
		t.Error("should succeed")
	}
	if mock.fieldValues["FullName"] != "John Doe" {
		t.Errorf("expected 'John Doe', got %v", mock.fieldValues["FullName"])
	}
	if len(output.Actions) != 1 {
		t.Errorf("expected 1 action, got %d", len(output.Actions))
	}
}

func TestMergeIt_Execute_ThreeFields(t *testing.T) {
	mock := &mockConnectorForMergeIt{
		fieldValues: map[string]interface{}{
			"FirstName":  "Jane",
			"MiddleName": "Marie",
			"LastName":   "Smith",
		},
	}
	output, err := (&MergeIt{}).Execute(context.Background(), helpers.HelperInput{
		ContactID: "123",
		Config: map[string]interface{}{
			"source_fields": []interface{}{"FirstName", "MiddleName", "LastName"},
			"target_field":  "FullName",
		},
		Connector: mock,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !output.Success {
		t.Error("should succeed")
	}
	if mock.fieldValues["FullName"] != "Jane Marie Smith" {
		t.Errorf("expected 'Jane Marie Smith', got %v", mock.fieldValues["FullName"])
	}
}

func TestMergeIt_Execute_CustomSeparator(t *testing.T) {
	mock := &mockConnectorForMergeIt{
		fieldValues: map[string]interface{}{
			"City":  "San Francisco",
			"State": "CA",
		},
	}
	output, err := (&MergeIt{}).Execute(context.Background(), helpers.HelperInput{
		ContactID: "123",
		Config: map[string]interface{}{
			"source_fields": []interface{}{"City", "State"},
			"target_field":  "Location",
			"separator":     ", ",
		},
		Connector: mock,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !output.Success {
		t.Error("should succeed")
	}
	if mock.fieldValues["Location"] != "San Francisco, CA" {
		t.Errorf("expected 'San Francisco, CA', got %v", mock.fieldValues["Location"])
	}
}

func TestMergeIt_Execute_SkipEmpty(t *testing.T) {
	mock := &mockConnectorForMergeIt{
		fieldValues: map[string]interface{}{
			"FirstName":  "Bob",
			"MiddleName": "",
			"LastName":   "Jones",
		},
	}
	output, err := (&MergeIt{}).Execute(context.Background(), helpers.HelperInput{
		ContactID: "123",
		Config: map[string]interface{}{
			"source_fields": []interface{}{"FirstName", "MiddleName", "LastName"},
			"target_field":  "FullName",
			"skip_empty":    true,
		},
		Connector: mock,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !output.Success {
		t.Error("should succeed")
	}
	if mock.fieldValues["FullName"] != "Bob Jones" {
		t.Errorf("expected 'Bob Jones', got %v", mock.fieldValues["FullName"])
	}
}

func TestMergeIt_Execute_DontSkipEmpty(t *testing.T) {
	mock := &mockConnectorForMergeIt{
		fieldValues: map[string]interface{}{
			"FirstName":  "Bob",
			"MiddleName": "",
			"LastName":   "Jones",
		},
	}
	output, err := (&MergeIt{}).Execute(context.Background(), helpers.HelperInput{
		ContactID: "123",
		Config: map[string]interface{}{
			"source_fields": []interface{}{"FirstName", "MiddleName", "LastName"},
			"target_field":  "FullName",
			"skip_empty":    false,
		},
		Connector: mock,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !output.Success {
		t.Error("should succeed")
	}
	if mock.fieldValues["FullName"] != "Bob  Jones" {
		t.Errorf("expected 'Bob  Jones' (with double space), got %v", mock.fieldValues["FullName"])
	}
}

func TestMergeIt_Execute_AllFieldsEmpty(t *testing.T) {
	mock := &mockConnectorForMergeIt{
		fieldValues: map[string]interface{}{
			"Field1": "",
			"Field2": nil,
		},
	}
	output, err := (&MergeIt{}).Execute(context.Background(), helpers.HelperInput{
		ContactID: "123",
		Config: map[string]interface{}{
			"source_fields": []interface{}{"Field1", "Field2"},
			"target_field":  "Result",
		},
		Connector: mock,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !output.Success {
		t.Error("should succeed")
	}
	if output.Message != "All source fields are empty, nothing to merge" {
		t.Errorf("wrong message: %s", output.Message)
	}
	if _, exists := mock.fieldValues["Result"]; exists {
		t.Error("Result should not be set when all sources empty")
	}
}

func TestMergeIt_Execute_GetFieldError(t *testing.T) {
	mock := &mockConnectorForMergeIt{
		fieldValues: map[string]interface{}{
			"FirstName": "John",
		},
		getFieldError: map[string]error{
			"LastName": fmt.Errorf("field read error"),
		},
	}
	output, err := (&MergeIt{}).Execute(context.Background(), helpers.HelperInput{
		ContactID: "123",
		Config: map[string]interface{}{
			"source_fields": []interface{}{"FirstName", "LastName"},
			"target_field":  "FullName",
		},
		Connector: mock,
	})
	if err != nil {
		t.Fatal(err)
	}
	// Should succeed with partial data
	if !output.Success {
		t.Error("should succeed with at least one field")
	}
	if mock.fieldValues["FullName"] != "John" {
		t.Errorf("expected 'John', got %v", mock.fieldValues["FullName"])
	}
}

func TestMergeIt_Execute_SetFieldError(t *testing.T) {
	mock := &mockConnectorForMergeIt{
		fieldValues: map[string]interface{}{
			"FirstName": "John",
			"LastName":  "Doe",
		},
		setFieldError: fmt.Errorf("field write error"),
	}
	output, err := (&MergeIt{}).Execute(context.Background(), helpers.HelperInput{
		ContactID: "123",
		Config: map[string]interface{}{
			"source_fields": []interface{}{"FirstName", "LastName"},
			"target_field":  "FullName",
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

func TestMergeIt_Execute_NilValuesSkipped(t *testing.T) {
	mock := &mockConnectorForMergeIt{
		fieldValues: map[string]interface{}{
			"Field1": "Value1",
			"Field2": nil,
			"Field3": "Value3",
		},
	}
	output, err := (&MergeIt{}).Execute(context.Background(), helpers.HelperInput{
		ContactID: "123",
		Config: map[string]interface{}{
			"source_fields": []interface{}{"Field1", "Field2", "Field3"},
			"target_field":  "Result",
			"skip_empty":    true,
		},
		Connector: mock,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !output.Success {
		t.Error("should succeed")
	}
	if mock.fieldValues["Result"] != "Value1 Value3" {
		t.Errorf("expected 'Value1 Value3', got %v", mock.fieldValues["Result"])
	}
}

func TestMergeIt_Execute_DifferentSeparators(t *testing.T) {
	tests := []struct {
		name      string
		separator string
		expected  string
	}{
		{"Comma", ",", "A,B,C"},
		{"Pipe", "|", "A|B|C"},
		{"Newline", "\n", "A\nB\nC"},
		{"Tab", "\t", "A\tB\tC"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockConnectorForMergeIt{
				fieldValues: map[string]interface{}{
					"F1": "A",
					"F2": "B",
					"F3": "C",
				},
			}
			output, err := (&MergeIt{}).Execute(context.Background(), helpers.HelperInput{
				ContactID: "123",
				Config: map[string]interface{}{
					"source_fields": []interface{}{"F1", "F2", "F3"},
					"target_field":  "Result",
					"separator":     tt.separator,
				},
				Connector: mock,
			})
			if err != nil {
				t.Fatal(err)
			}
			if !output.Success {
				t.Error("should succeed")
			}
			if mock.fieldValues["Result"] != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, mock.fieldValues["Result"])
			}
		})
	}
}
