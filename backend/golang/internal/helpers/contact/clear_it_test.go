package contact

import (
	"context"
	"fmt"
	"testing"

	"github.com/myfusionhelper/api/internal/connectors"
	"github.com/myfusionhelper/api/internal/helpers"
)

type mockConnectorForClearIt struct {
	fieldsSet     map[string]interface{}
	setFieldError map[string]error
}

func (m *mockConnectorForClearIt) SetContactFieldValue(ctx context.Context, contactID, fieldKey string, value interface{}) error {
	if m.setFieldError != nil {
		if err, ok := m.setFieldError[fieldKey]; ok {
			return err
		}
	}
	if m.fieldsSet == nil {
		m.fieldsSet = make(map[string]interface{})
	}
	m.fieldsSet[fieldKey] = value
	return nil
}

func (m *mockConnectorForClearIt) GetContacts(ctx context.Context, opts connectors.QueryOptions) (*connectors.ContactList, error) { return nil, fmt.Errorf("not implemented") }
func (m *mockConnectorForClearIt) GetContact(ctx context.Context, contactID string) (*connectors.NormalizedContact, error) { return nil, fmt.Errorf("not implemented") }
func (m *mockConnectorForClearIt) CreateContact(ctx context.Context, contact connectors.CreateContactInput) (*connectors.NormalizedContact, error) { return nil, fmt.Errorf("not implemented") }
func (m *mockConnectorForClearIt) UpdateContact(ctx context.Context, contactID string, updates connectors.UpdateContactInput) (*connectors.NormalizedContact, error) { return nil, fmt.Errorf("not implemented") }
func (m *mockConnectorForClearIt) DeleteContact(ctx context.Context, contactID string) error { return fmt.Errorf("not implemented") }
func (m *mockConnectorForClearIt) GetTags(ctx context.Context) ([]connectors.Tag, error) { return nil, fmt.Errorf("not implemented") }
func (m *mockConnectorForClearIt) ApplyTag(ctx context.Context, contactID, tagID string) error { return fmt.Errorf("not implemented") }
func (m *mockConnectorForClearIt) RemoveTag(ctx context.Context, contactID, tagID string) error { return fmt.Errorf("not implemented") }
func (m *mockConnectorForClearIt) GetCustomFields(ctx context.Context) ([]connectors.CustomField, error) { return nil, fmt.Errorf("not implemented") }
func (m *mockConnectorForClearIt) GetContactFieldValue(ctx context.Context, contactID, fieldKey string) (interface{}, error) { return nil, fmt.Errorf("not implemented") }
func (m *mockConnectorForClearIt) TriggerAutomation(ctx context.Context, contactID, automationID string) error { return fmt.Errorf("not implemented") }
func (m *mockConnectorForClearIt) AchieveGoal(ctx context.Context, contactID, goalName, integration string) error { return fmt.Errorf("not implemented") }
func (m *mockConnectorForClearIt) TestConnection(ctx context.Context) error { return fmt.Errorf("not implemented") }
func (m *mockConnectorForClearIt) GetMetadata() connectors.ConnectorMetadata { return connectors.ConnectorMetadata{} }
func (m *mockConnectorForClearIt) GetCapabilities() []connectors.Capability { return nil }

func TestClearIt_GetMetadata(t *testing.T) {
	h := &ClearIt{}
	if h.GetName() != "Clear It" { t.Error("wrong name") }
	if h.GetType() != "clear_it" { t.Error("wrong type") }
	if h.GetCategory() != "contact" { t.Error("wrong category") }
	if !h.RequiresCRM() { t.Error("should require CRM") }
}

func TestClearIt_ValidateConfig_Missing(t *testing.T) {
	if err := (&ClearIt{}).ValidateConfig(map[string]interface{}{}); err == nil { t.Error("should error on missing fields") }
}

func TestClearIt_ValidateConfig_Empty(t *testing.T) {
	if err := (&ClearIt{}).ValidateConfig(map[string]interface{}{"fields": []interface{}{}}); err == nil { t.Error("should error on empty fields") }
}

func TestClearIt_ValidateConfig_Valid(t *testing.T) {
	if err := (&ClearIt{}).ValidateConfig(map[string]interface{}{"fields": []interface{}{"field1"}}); err != nil { t.Error(err) }
}

func TestClearIt_Execute_SingleField(t *testing.T) {
	mock := &mockConnectorForClearIt{}
	output, err := (&ClearIt{}).Execute(context.Background(), helpers.HelperInput{
		ContactID: "123",
		Config: map[string]interface{}{"fields": []interface{}{"Company"}},
		Connector: mock,
	})
	if err != nil { t.Fatal(err) }
	if !output.Success { t.Error("should succeed") }
	if mock.fieldsSet["Company"] != "" { t.Error("Company should be cleared") }
}

func TestClearIt_Execute_MultipleFields(t *testing.T) {
	mock := &mockConnectorForClearIt{}
	output, err := (&ClearIt{}).Execute(context.Background(), helpers.HelperInput{
		ContactID: "123",
		Config: map[string]interface{}{"fields": []interface{}{"Phone", "Email", "Company"}},
		Connector: mock,
	})
	if err != nil { t.Fatal(err) }
	if !output.Success { t.Error("should succeed") }
	if len(output.Actions) != 3 { t.Errorf("expected 3 actions, got %d", len(output.Actions)) }
	if mock.fieldsSet["Phone"] != "" || mock.fieldsSet["Email"] != "" || mock.fieldsSet["Company"] != "" {
		t.Error("all fields should be cleared")
	}
}

func TestClearIt_Execute_PartialFailure(t *testing.T) {
	mock := &mockConnectorForClearIt{
		setFieldError: map[string]error{"Email": fmt.Errorf("API error")},
	}
	output, err := (&ClearIt{}).Execute(context.Background(), helpers.HelperInput{
		ContactID: "123",
		Config: map[string]interface{}{"fields": []interface{}{"Phone", "Email"}},
		Connector: mock,
	})
	if err != nil { t.Fatal(err) }
	if !output.Success { t.Error("should succeed if at least one field cleared") }
	if mock.fieldsSet["Phone"] != "" { t.Error("Phone should be cleared") }
	if _, ok := mock.fieldsSet["Email"]; ok { t.Error("Email should not be set due to error") }
}

func TestClearIt_Execute_AllFail(t *testing.T) {
	mock := &mockConnectorForClearIt{
		setFieldError: map[string]error{"Phone": fmt.Errorf("error")},
	}
	output, _ := (&ClearIt{}).Execute(context.Background(), helpers.HelperInput{
		ContactID: "123",
		Config: map[string]interface{}{"fields": []interface{}{"Phone"}},
		Connector: mock,
	})
	if output.Success { t.Error("should fail if no fields cleared") }
}
