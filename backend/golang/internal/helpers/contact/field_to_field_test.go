package contact

import (
	"context"
	"fmt"
	"testing"

	"github.com/myfusionhelper/api/internal/connectors"
	"github.com/myfusionhelper/api/internal/helpers"
)

type mockConnectorForFieldToField struct {
	fieldValues   map[string]interface{}
	getFieldError map[string]error
	setFieldError map[string]error
}

func (m *mockConnectorForFieldToField) GetContactFieldValue(ctx context.Context, contactID, fieldKey string) (interface{}, error) {
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

func (m *mockConnectorForFieldToField) SetContactFieldValue(ctx context.Context, contactID, fieldKey string, value interface{}) error {
	if m.setFieldError != nil {
		if err, ok := m.setFieldError[fieldKey]; ok {
			return err
		}
	}
	if m.fieldValues == nil {
		m.fieldValues = make(map[string]interface{})
	}
	m.fieldValues[fieldKey] = value
	return nil
}

func (m *mockConnectorForFieldToField) GetContacts(ctx context.Context, opts connectors.QueryOptions) (*connectors.ContactList, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForFieldToField) GetContact(ctx context.Context, contactID string) (*connectors.NormalizedContact, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForFieldToField) CreateContact(ctx context.Context, contact connectors.CreateContactInput) (*connectors.NormalizedContact, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForFieldToField) UpdateContact(ctx context.Context, contactID string, updates connectors.UpdateContactInput) (*connectors.NormalizedContact, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForFieldToField) DeleteContact(ctx context.Context, contactID string) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorForFieldToField) GetTags(ctx context.Context) ([]connectors.Tag, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForFieldToField) ApplyTag(ctx context.Context, contactID, tagID string) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorForFieldToField) RemoveTag(ctx context.Context, contactID, tagID string) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorForFieldToField) GetCustomFields(ctx context.Context) ([]connectors.CustomField, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForFieldToField) TriggerAutomation(ctx context.Context, contactID, automationID string) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorForFieldToField) AchieveGoal(ctx context.Context, contactID, goalName, integration string) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorForFieldToField) TestConnection(ctx context.Context) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorForFieldToField) GetMetadata() connectors.ConnectorMetadata {
	return connectors.ConnectorMetadata{}
}
func (m *mockConnectorForFieldToField) GetCapabilities() []connectors.Capability {
	return nil
}

func TestFieldToField_GetMetadata(t *testing.T) {
	h := &FieldToField{}
	if h.GetName() != "Field to Field" {
		t.Error("wrong name")
	}
	if h.GetType() != "field_to_field" {
		t.Error("wrong type")
	}
	if h.GetCategory() != "contact" {
		t.Error("wrong category")
	}
	if !h.RequiresCRM() {
		t.Error("should require CRM")
	}
}

func TestFieldToField_ValidateConfig_MissingMappings(t *testing.T) {
	err := (&FieldToField{}).ValidateConfig(map[string]interface{}{})
	if err == nil {
		t.Error("should error on missing mappings")
	}
}

func TestFieldToField_ValidateConfig_MappingsNotArray(t *testing.T) {
	err := (&FieldToField{}).ValidateConfig(map[string]interface{}{
		"mappings": "not an array",
	})
	if err == nil {
		t.Error("should error when mappings is not an array")
	}
}

func TestFieldToField_ValidateConfig_EmptyMappings(t *testing.T) {
	err := (&FieldToField{}).ValidateConfig(map[string]interface{}{
		"mappings": []interface{}{},
	})
	if err == nil {
		t.Error("should error on empty mappings")
	}
}

func TestFieldToField_ValidateConfig_MissingSource(t *testing.T) {
	err := (&FieldToField{}).ValidateConfig(map[string]interface{}{
		"mappings": []interface{}{
			map[string]interface{}{
				"target": "Email2",
			},
		},
	})
	if err == nil {
		t.Error("should error on missing source field")
	}
}

func TestFieldToField_ValidateConfig_MissingTarget(t *testing.T) {
	err := (&FieldToField{}).ValidateConfig(map[string]interface{}{
		"mappings": []interface{}{
			map[string]interface{}{
				"source": "Email",
			},
		},
	})
	if err == nil {
		t.Error("should error on missing target field")
	}
}

func TestFieldToField_ValidateConfig_Valid(t *testing.T) {
	err := (&FieldToField{}).ValidateConfig(map[string]interface{}{
		"mappings": []interface{}{
			map[string]interface{}{
				"source": "Email",
				"target": "Email2",
			},
		},
	})
	if err != nil {
		t.Errorf("should be valid: %v", err)
	}
}

func TestFieldToField_Execute_SingleMapping(t *testing.T) {
	mock := &mockConnectorForFieldToField{
		fieldValues: map[string]interface{}{
			"Email": "john@example.com",
		},
	}
	output, err := (&FieldToField{}).Execute(context.Background(), helpers.HelperInput{
		ContactID: "123",
		Config: map[string]interface{}{
			"mappings": []interface{}{
				map[string]interface{}{
					"source": "Email",
					"target": "Email2",
				},
			},
		},
		Connector: mock,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !output.Success {
		t.Error("should succeed")
	}
	if mock.fieldValues["Email2"] != "john@example.com" {
		t.Errorf("expected Email2=john@example.com, got %v", mock.fieldValues["Email2"])
	}
	if len(output.Actions) != 1 {
		t.Errorf("expected 1 action, got %d", len(output.Actions))
	}
	if output.Actions[0].Type != "field_updated" || output.Actions[0].Target != "Email2" {
		t.Error("wrong action")
	}
}

func TestFieldToField_Execute_MultipleMappings(t *testing.T) {
	mock := &mockConnectorForFieldToField{
		fieldValues: map[string]interface{}{
			"FirstName": "John",
			"LastName":  "Doe",
			"Email":     "john@example.com",
		},
	}
	output, err := (&FieldToField{}).Execute(context.Background(), helpers.HelperInput{
		ContactID: "123",
		Config: map[string]interface{}{
			"mappings": []interface{}{
				map[string]interface{}{"source": "FirstName", "target": "FirstName2"},
				map[string]interface{}{"source": "LastName", "target": "LastName2"},
				map[string]interface{}{"source": "Email", "target": "Email2"},
			},
		},
		Connector: mock,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !output.Success {
		t.Error("should succeed")
	}
	if mock.fieldValues["FirstName2"] != "John" {
		t.Error("FirstName2 should be John")
	}
	if mock.fieldValues["LastName2"] != "Doe" {
		t.Error("LastName2 should be Doe")
	}
	if mock.fieldValues["Email2"] != "john@example.com" {
		t.Error("Email2 should be john@example.com")
	}
	if len(output.Actions) != 3 {
		t.Errorf("expected 3 actions, got %d", len(output.Actions))
	}
}

func TestFieldToField_Execute_EmptySourceSkipped(t *testing.T) {
	mock := &mockConnectorForFieldToField{
		fieldValues: map[string]interface{}{
			"Email": "",
		},
	}
	output, err := (&FieldToField{}).Execute(context.Background(), helpers.HelperInput{
		ContactID: "123",
		Config: map[string]interface{}{
			"mappings": []interface{}{
				map[string]interface{}{
					"source": "Email",
					"target": "Email2",
				},
			},
		},
		Connector: mock,
	})
	if err != nil {
		t.Fatal(err)
	}
	if output.Success {
		t.Error("should fail when no fields copied")
	}
	if _, exists := mock.fieldValues["Email2"]; exists {
		t.Error("Email2 should not be set")
	}
}

func TestFieldToField_Execute_NilSourceSkipped(t *testing.T) {
	mock := &mockConnectorForFieldToField{
		fieldValues: map[string]interface{}{
			"Email": nil,
		},
	}
	output, err := (&FieldToField{}).Execute(context.Background(), helpers.HelperInput{
		ContactID: "123",
		Config: map[string]interface{}{
			"mappings": []interface{}{
				map[string]interface{}{
					"source": "Email",
					"target": "Email2",
				},
			},
		},
		Connector: mock,
	})
	if err != nil {
		t.Fatal(err)
	}
	if output.Success {
		t.Error("should fail when no fields copied")
	}
}

func TestFieldToField_Execute_OverwriteTrue(t *testing.T) {
	mock := &mockConnectorForFieldToField{
		fieldValues: map[string]interface{}{
			"Email":  "new@example.com",
			"Email2": "old@example.com",
		},
	}
	output, err := (&FieldToField{}).Execute(context.Background(), helpers.HelperInput{
		ContactID: "123",
		Config: map[string]interface{}{
			"mappings": []interface{}{
				map[string]interface{}{
					"source": "Email",
					"target": "Email2",
				},
			},
			"overwrite": true,
		},
		Connector: mock,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !output.Success {
		t.Error("should succeed")
	}
	if mock.fieldValues["Email2"] != "new@example.com" {
		t.Errorf("expected Email2=new@example.com, got %v", mock.fieldValues["Email2"])
	}
}

func TestFieldToField_Execute_OverwriteFalse(t *testing.T) {
	mock := &mockConnectorForFieldToField{
		fieldValues: map[string]interface{}{
			"Email":  "new@example.com",
			"Email2": "old@example.com",
		},
	}
	output, err := (&FieldToField{}).Execute(context.Background(), helpers.HelperInput{
		ContactID: "123",
		Config: map[string]interface{}{
			"mappings": []interface{}{
				map[string]interface{}{
					"source": "Email",
					"target": "Email2",
				},
			},
			"overwrite": false,
		},
		Connector: mock,
	})
	if err != nil {
		t.Fatal(err)
	}
	if output.Success {
		t.Error("should fail when no fields copied")
	}
	if mock.fieldValues["Email2"] != "old@example.com" {
		t.Errorf("Email2 should remain old@example.com, got %v", mock.fieldValues["Email2"])
	}
}

func TestFieldToField_Execute_GetFieldError(t *testing.T) {
	mock := &mockConnectorForFieldToField{
		getFieldError: map[string]error{
			"Email": fmt.Errorf("field read error"),
		},
	}
	output, err := (&FieldToField{}).Execute(context.Background(), helpers.HelperInput{
		ContactID: "123",
		Config: map[string]interface{}{
			"mappings": []interface{}{
				map[string]interface{}{
					"source": "Email",
					"target": "Email2",
				},
			},
		},
		Connector: mock,
	})
	if err != nil {
		t.Fatal(err)
	}
	if output.Success {
		t.Error("should fail when no fields copied")
	}
}

func TestFieldToField_Execute_SetFieldError(t *testing.T) {
	mock := &mockConnectorForFieldToField{
		fieldValues: map[string]interface{}{
			"Email": "test@example.com",
		},
		setFieldError: map[string]error{
			"Email2": fmt.Errorf("field write error"),
		},
	}
	output, err := (&FieldToField{}).Execute(context.Background(), helpers.HelperInput{
		ContactID: "123",
		Config: map[string]interface{}{
			"mappings": []interface{}{
				map[string]interface{}{
					"source": "Email",
					"target": "Email2",
				},
			},
		},
		Connector: mock,
	})
	if err != nil {
		t.Fatal(err)
	}
	if output.Success {
		t.Error("should fail when no fields copied")
	}
}

func TestFieldToField_Execute_PartialSuccess(t *testing.T) {
	mock := &mockConnectorForFieldToField{
		fieldValues: map[string]interface{}{
			"Email": "test@example.com",
			"Phone": "",
		},
	}
	output, err := (&FieldToField{}).Execute(context.Background(), helpers.HelperInput{
		ContactID: "123",
		Config: map[string]interface{}{
			"mappings": []interface{}{
				map[string]interface{}{"source": "Email", "target": "Email2"},
				map[string]interface{}{"source": "Phone", "target": "Phone2"},
			},
		},
		Connector: mock,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !output.Success {
		t.Error("should succeed if at least one field copied")
	}
	if mock.fieldValues["Email2"] != "test@example.com" {
		t.Error("Email2 should be set")
	}
	if _, exists := mock.fieldValues["Phone2"]; exists {
		t.Error("Phone2 should not be set (empty source)")
	}
}
