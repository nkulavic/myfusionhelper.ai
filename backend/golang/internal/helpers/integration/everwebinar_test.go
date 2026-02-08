package integration

import (
	"context"
	"fmt"
	"testing"

	"github.com/myfusionhelper/api/internal/connectors"
	"github.com/myfusionhelper/api/internal/helpers"
)

type mockConnectorForEverWebinar struct {
	contact          *connectors.NormalizedContact
	getContactError  error
	applyTagError    error
	tagApplied       bool
	appliedTagID     string
}

func (m *mockConnectorForEverWebinar) GetContact(ctx context.Context, contactID string) (*connectors.NormalizedContact, error) {
	if m.getContactError != nil {
		return nil, m.getContactError
	}
	return m.contact, nil
}

func (m *mockConnectorForEverWebinar) ApplyTag(ctx context.Context, contactID, tagID string) error {
	if m.applyTagError != nil {
		return m.applyTagError
	}
	m.tagApplied = true
	m.appliedTagID = tagID
	return nil
}

func (m *mockConnectorForEverWebinar) GetContacts(ctx context.Context, opts connectors.QueryOptions) (*connectors.ContactList, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *mockConnectorForEverWebinar) CreateContact(ctx context.Context, input connectors.CreateContactInput) (*connectors.NormalizedContact, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *mockConnectorForEverWebinar) UpdateContact(ctx context.Context, contactID string, updates connectors.UpdateContactInput) (*connectors.NormalizedContact, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *mockConnectorForEverWebinar) DeleteContact(ctx context.Context, contactID string) error {
	return fmt.Errorf("not implemented")
}

func (m *mockConnectorForEverWebinar) GetTags(ctx context.Context) ([]connectors.Tag, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *mockConnectorForEverWebinar) RemoveTag(ctx context.Context, contactID, tagID string) error {
	return fmt.Errorf("not implemented")
}

func (m *mockConnectorForEverWebinar) GetCustomFields(ctx context.Context) ([]connectors.CustomField, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *mockConnectorForEverWebinar) GetContactFieldValue(ctx context.Context, contactID, fieldKey string) (interface{}, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *mockConnectorForEverWebinar) SetContactFieldValue(ctx context.Context, contactID, fieldKey string, value interface{}) error {
	return fmt.Errorf("not implemented")
}

func (m *mockConnectorForEverWebinar) TriggerAutomation(ctx context.Context, contactID, automationID string) error {
	return fmt.Errorf("not implemented")
}

func (m *mockConnectorForEverWebinar) AchieveGoal(ctx context.Context, contactID, goalName, integration string) error {
	return fmt.Errorf("not implemented")
}

func (m *mockConnectorForEverWebinar) TestConnection(ctx context.Context) error {
	return fmt.Errorf("not implemented")
}

func (m *mockConnectorForEverWebinar) GetMetadata() connectors.ConnectorMetadata {
	return connectors.ConnectorMetadata{PlatformSlug: "mock", PlatformName: "Mock"}
}

func (m *mockConnectorForEverWebinar) GetCapabilities() []connectors.Capability {
	return []connectors.Capability{}
}

func TestEverWebinar_GetMetadata(t *testing.T) {
	helper := &EverWebinar{}
	if helper.GetName() != "EverWebinar" {
		t.Errorf("Expected name 'EverWebinar', got '%s'", helper.GetName())
	}
	if helper.GetType() != "everwebinar" {
		t.Errorf("Expected type 'everwebinar', got '%s'", helper.GetType())
	}
	if helper.GetCategory() != "integration" {
		t.Errorf("Expected category 'integration', got '%s'", helper.GetCategory())
	}
}

func TestEverWebinar_ValidateConfig_Success(t *testing.T) {
	helper := &EverWebinar{}
	err := helper.ValidateConfig(map[string]interface{}{"webinar_id": "webinar123"})
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
}

func TestEverWebinar_ValidateConfig_MissingWebinarID(t *testing.T) {
	helper := &EverWebinar{}
	err := helper.ValidateConfig(map[string]interface{}{})
	if err == nil {
		t.Error("Expected error for missing webinar_id")
	}
}

func TestEverWebinar_Execute_Success(t *testing.T) {
	helper := &EverWebinar{}
	mock := &mockConnectorForEverWebinar{
		contact: &connectors.NormalizedContact{ID: "123", FirstName: "John", LastName: "Doe", Email: "john@example.com"},
	}

	output, err := helper.Execute(context.Background(), helpers.HelperInput{
		ContactID: "123",
		Config: map[string]interface{}{"webinar_id": "webinar123"},
		ServiceAuths: map[string]*connectors.ConnectorConfig{"everwebinar": {APIKey: "test-key"}},
		Connector: mock,
	})

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if !output.Success {
		t.Error("Expected success")
	}
	if len(output.Actions) != 1 || output.Actions[0].Type != "webhook_queued" {
		t.Error("Expected webhook action")
	}
}

func TestEverWebinar_Execute_NoServiceAuth(t *testing.T) {
	helper := &EverWebinar{}
	mock := &mockConnectorForEverWebinar{contact: &connectors.NormalizedContact{ID: "123", Email: "test@test.com"}}

	_, err := helper.Execute(context.Background(), helpers.HelperInput{
		ContactID: "123",
		Config: map[string]interface{}{"webinar_id": "webinar123"},
		ServiceAuths: map[string]*connectors.ConnectorConfig{},
		Connector: mock,
	})

	if err == nil {
		t.Error("Expected error for missing service auth")
	}
}

func TestEverWebinar_Execute_NoAPIKey(t *testing.T) {
	helper := &EverWebinar{}
	mock := &mockConnectorForEverWebinar{contact: &connectors.NormalizedContact{ID: "123", Email: "test@test.com"}}

	_, err := helper.Execute(context.Background(), helpers.HelperInput{
		ContactID: "123",
		Config: map[string]interface{}{"webinar_id": "webinar123"},
		ServiceAuths: map[string]*connectors.ConnectorConfig{"everwebinar": {APIKey: ""}},
		Connector: mock,
	})

	if err == nil {
		t.Error("Expected error for missing API key")
	}
}

func TestEverWebinar_Execute_EmptyEmail(t *testing.T) {
	helper := &EverWebinar{}
	mock := &mockConnectorForEverWebinar{contact: &connectors.NormalizedContact{ID: "123", Email: ""}}

	_, err := helper.Execute(context.Background(), helpers.HelperInput{
		ContactID: "123",
		Config: map[string]interface{}{"webinar_id": "webinar123"},
		ServiceAuths: map[string]*connectors.ConnectorConfig{"everwebinar": {APIKey: "test-key"}},
		Connector: mock,
	})

	if err == nil {
		t.Error("Expected error for empty email")
	}
}

func TestEverWebinar_Execute_WithSchedule(t *testing.T) {
	helper := &EverWebinar{}
	mock := &mockConnectorForEverWebinar{contact: &connectors.NormalizedContact{ID: "123", Email: "test@test.com"}}

	output, err := helper.Execute(context.Background(), helpers.HelperInput{
		ContactID: "123",
		Config: map[string]interface{}{"webinar_id": "webinar123", "schedule": "evening"},
		ServiceAuths: map[string]*connectors.ConnectorConfig{"everwebinar": {APIKey: "test-key"}},
		Connector: mock,
	})

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	actionValue := output.Actions[0].Value.(map[string]interface{})
	payload := actionValue["payload"].(map[string]interface{})
	if payload["schedule"] != "evening" {
		t.Error("Expected schedule in payload")
	}
}

func TestEverWebinar_Execute_WithApplyTag(t *testing.T) {
	helper := &EverWebinar{}
	mock := &mockConnectorForEverWebinar{contact: &connectors.NormalizedContact{ID: "123", Email: "test@test.com"}}

	_, err := helper.Execute(context.Background(), helpers.HelperInput{
		ContactID: "123",
		Config: map[string]interface{}{"webinar_id": "webinar123", "apply_tag": "tag789"},
		ServiceAuths: map[string]*connectors.ConnectorConfig{"everwebinar": {APIKey: "test-key"}},
		Connector: mock,
	})

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if !mock.tagApplied || mock.appliedTagID != "tag789" {
		t.Error("Expected tag to be applied")
	}
}

func TestEverWebinar_Execute_UseContactData(t *testing.T) {
	helper := &EverWebinar{}
	mock := &mockConnectorForEverWebinar{}

	contactData := &connectors.NormalizedContact{ID: "123", FirstName: "Jane", Email: "jane@test.com"}
	output, err := helper.Execute(context.Background(), helpers.HelperInput{
		ContactID: "123",
		Config: map[string]interface{}{"webinar_id": "webinar123"},
		ServiceAuths: map[string]*connectors.ConnectorConfig{"everwebinar": {APIKey: "test-key"}},
		ContactData: contactData,
		Connector: mock,
	})

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if output.ModifiedData["email"] != "jane@test.com" {
		t.Error("Expected to use ContactData")
	}
}
