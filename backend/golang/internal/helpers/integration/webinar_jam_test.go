package integration

import (
	"context"
	"fmt"
	"testing"

	"github.com/myfusionhelper/api/internal/connectors"
	"github.com/myfusionhelper/api/internal/helpers"
)

// Mock connector for WebinarJam tests
type mockConnectorForWebinarJam struct {
	contact          *connectors.NormalizedContact
	getContactError  error
	applyTagError    error
	tagApplied       bool
	appliedTagID     string
}

func (m *mockConnectorForWebinarJam) GetContact(ctx context.Context, contactID string) (*connectors.NormalizedContact, error) {
	if m.getContactError != nil {
		return nil, m.getContactError
	}
	return m.contact, nil
}

func (m *mockConnectorForWebinarJam) ApplyTag(ctx context.Context, contactID, tagID string) error {
	if m.applyTagError != nil {
		return m.applyTagError
	}
	m.tagApplied = true
	m.appliedTagID = tagID
	return nil
}

func (m *mockConnectorForWebinarJam) GetContacts(ctx context.Context, opts connectors.QueryOptions) (*connectors.ContactList, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *mockConnectorForWebinarJam) CreateContact(ctx context.Context, input connectors.CreateContactInput) (*connectors.NormalizedContact, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *mockConnectorForWebinarJam) UpdateContact(ctx context.Context, contactID string, updates connectors.UpdateContactInput) (*connectors.NormalizedContact, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *mockConnectorForWebinarJam) DeleteContact(ctx context.Context, contactID string) error {
	return fmt.Errorf("not implemented")
}

func (m *mockConnectorForWebinarJam) GetTags(ctx context.Context) ([]connectors.Tag, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *mockConnectorForWebinarJam) RemoveTag(ctx context.Context, contactID, tagID string) error {
	return fmt.Errorf("not implemented")
}

func (m *mockConnectorForWebinarJam) GetCustomFields(ctx context.Context) ([]connectors.CustomField, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *mockConnectorForWebinarJam) GetContactFieldValue(ctx context.Context, contactID, fieldKey string) (interface{}, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *mockConnectorForWebinarJam) SetContactFieldValue(ctx context.Context, contactID, fieldKey string, value interface{}) error {
	return fmt.Errorf("not implemented")
}

func (m *mockConnectorForWebinarJam) TriggerAutomation(ctx context.Context, contactID, automationID string) error {
	return fmt.Errorf("not implemented")
}

func (m *mockConnectorForWebinarJam) AchieveGoal(ctx context.Context, contactID, goalName, integration string) error {
	return fmt.Errorf("not implemented")
}

func (m *mockConnectorForWebinarJam) TestConnection(ctx context.Context) error {
	return fmt.Errorf("not implemented")
}

func (m *mockConnectorForWebinarJam) GetMetadata() connectors.ConnectorMetadata {
	return connectors.ConnectorMetadata{PlatformSlug: "mock", PlatformName: "Mock"}
}

func (m *mockConnectorForWebinarJam) GetCapabilities() []connectors.Capability {
	return []connectors.Capability{}
}

func TestWebinarJam_GetMetadata(t *testing.T) {
	helper := &WebinarJam{}
	if helper.GetName() != "WebinarJam" {
		t.Errorf("Expected name 'WebinarJam', got '%s'", helper.GetName())
	}
	if helper.GetType() != "webinar_jam" {
		t.Errorf("Expected type 'webinar_jam', got '%s'", helper.GetType())
	}
	if helper.GetCategory() != "integration" {
		t.Errorf("Expected category 'integration', got '%s'", helper.GetCategory())
	}
}

func TestWebinarJam_ValidateConfig_Success(t *testing.T) {
	helper := &WebinarJam{}
	err := helper.ValidateConfig(map[string]interface{}{"webinar_id": "webinar123"})
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
}

func TestWebinarJam_ValidateConfig_MissingWebinarID(t *testing.T) {
	helper := &WebinarJam{}
	err := helper.ValidateConfig(map[string]interface{}{})
	if err == nil {
		t.Error("Expected error for missing webinar_id")
	}
}

func TestWebinarJam_Execute_Success(t *testing.T) {
	helper := &WebinarJam{}
	mock := &mockConnectorForWebinarJam{
		contact: &connectors.NormalizedContact{
			ID: "123", FirstName: "John", LastName: "Doe", Email: "john@example.com",
		},
	}

	output, err := helper.Execute(context.Background(), helpers.HelperInput{
		ContactID: "123",
		Config: map[string]interface{}{"webinar_id": "webinar123"},
		ServiceAuths: map[string]*connectors.ConnectorConfig{
			"webinarjam": {APIKey: "test-key"},
		},
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

func TestWebinarJam_Execute_NoServiceAuth(t *testing.T) {
	helper := &WebinarJam{}
	mock := &mockConnectorForWebinarJam{contact: &connectors.NormalizedContact{ID: "123", Email: "test@test.com"}}

	output, err := helper.Execute(context.Background(), helpers.HelperInput{
		ContactID: "123",
		Config: map[string]interface{}{"webinar_id": "webinar123"},
		ServiceAuths: map[string]*connectors.ConnectorConfig{},
		Connector: mock,
	})

	if err == nil {
		t.Error("Expected error for missing service auth")
	}
	if output.Success {
		t.Error("Expected failure")
	}
}

func TestWebinarJam_Execute_NoAPIKey(t *testing.T) {
	helper := &WebinarJam{}
	mock := &mockConnectorForWebinarJam{contact: &connectors.NormalizedContact{ID: "123", Email: "test@test.com"}}

	_, err := helper.Execute(context.Background(), helpers.HelperInput{
		ContactID: "123",
		Config: map[string]interface{}{"webinar_id": "webinar123"},
		ServiceAuths: map[string]*connectors.ConnectorConfig{"webinarjam": {APIKey: ""}},
		Connector: mock,
	})

	if err == nil {
		t.Error("Expected error for missing API key")
	}
}

func TestWebinarJam_Execute_EmptyEmail(t *testing.T) {
	helper := &WebinarJam{}
	mock := &mockConnectorForWebinarJam{contact: &connectors.NormalizedContact{ID: "123", Email: ""}}

	_, err := helper.Execute(context.Background(), helpers.HelperInput{
		ContactID: "123",
		Config: map[string]interface{}{"webinar_id": "webinar123"},
		ServiceAuths: map[string]*connectors.ConnectorConfig{"webinarjam": {APIKey: "test-key"}},
		Connector: mock,
	})

	if err == nil {
		t.Error("Expected error for empty email")
	}
}

func TestWebinarJam_Execute_WithSchedule(t *testing.T) {
	helper := &WebinarJam{}
	mock := &mockConnectorForWebinarJam{contact: &connectors.NormalizedContact{ID: "123", Email: "test@test.com"}}

	output, err := helper.Execute(context.Background(), helpers.HelperInput{
		ContactID: "123",
		Config: map[string]interface{}{"webinar_id": "webinar123", "schedule": "morning"},
		ServiceAuths: map[string]*connectors.ConnectorConfig{"webinarjam": {APIKey: "test-key"}},
		Connector: mock,
	})

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	
	actionValue := output.Actions[0].Value.(map[string]interface{})
	payload := actionValue["payload"].(map[string]interface{})
	if payload["schedule"] != "morning" {
		t.Error("Expected schedule in payload")
	}
}

func TestWebinarJam_Execute_WithApplyTag(t *testing.T) {
	helper := &WebinarJam{}
	mock := &mockConnectorForWebinarJam{contact: &connectors.NormalizedContact{ID: "123", Email: "test@test.com"}}

	output, err := helper.Execute(context.Background(), helpers.HelperInput{
		ContactID: "123",
		Config: map[string]interface{}{"webinar_id": "webinar123", "apply_tag": "tag456"},
		ServiceAuths: map[string]*connectors.ConnectorConfig{"webinarjam": {APIKey: "test-key"}},
		Connector: mock,
	})

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if !mock.tagApplied || mock.appliedTagID != "tag456" {
		t.Error("Expected tag to be applied")
	}
	if len(output.Actions) != 2 || output.Actions[1].Type != "tag_applied" {
		t.Error("Expected tag_applied action")
	}
}

func TestWebinarJam_Execute_ApplyTagError(t *testing.T) {
	helper := &WebinarJam{}
	mock := &mockConnectorForWebinarJam{
		contact: &connectors.NormalizedContact{ID: "123", Email: "test@test.com"},
		applyTagError: fmt.Errorf("tag error"),
	}

	output, err := helper.Execute(context.Background(), helpers.HelperInput{
		ContactID: "123",
		Config: map[string]interface{}{"webinar_id": "webinar123", "apply_tag": "tag456"},
		ServiceAuths: map[string]*connectors.ConnectorConfig{"webinarjam": {APIKey: "test-key"}},
		Connector: mock,
	})

	if err != nil {
		t.Fatalf("Expected no error (tag failure should be logged), got: %v", err)
	}
	if len(output.Actions) != 1 {
		t.Error("Expected only webhook action (tag should fail)")
	}
}

func TestWebinarJam_Execute_UseContactData(t *testing.T) {
	helper := &WebinarJam{}
	mock := &mockConnectorForWebinarJam{}

	contactData := &connectors.NormalizedContact{ID: "123", FirstName: "Jane", Email: "jane@test.com"}
	output, err := helper.Execute(context.Background(), helpers.HelperInput{
		ContactID: "123",
		Config: map[string]interface{}{"webinar_id": "webinar123"},
		ServiceAuths: map[string]*connectors.ConnectorConfig{"webinarjam": {APIKey: "test-key"}},
		ContactData: contactData,
		Connector: mock,
	})

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if output.ModifiedData["email"] != "jane@test.com" {
		t.Error("Expected to use ContactData instead of fetching")
	}
}

func TestWebinarJam_Execute_GetContactError(t *testing.T) {
	helper := &WebinarJam{}
	mock := &mockConnectorForWebinarJam{getContactError: fmt.Errorf("contact error")}

	_, err := helper.Execute(context.Background(), helpers.HelperInput{
		ContactID: "123",
		Config: map[string]interface{}{"webinar_id": "webinar123"},
		ServiceAuths: map[string]*connectors.ConnectorConfig{"webinarjam": {APIKey: "test-key"}},
		Connector: mock,
	})

	if err == nil {
		t.Error("Expected error when GetContact fails")
	}
}
