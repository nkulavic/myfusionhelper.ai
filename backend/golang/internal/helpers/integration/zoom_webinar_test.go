package integration

import (
	"context"
	"fmt"
	"testing"

	"github.com/myfusionhelper/api/internal/connectors"
	"github.com/myfusionhelper/api/internal/helpers"
)

type mockConnectorForZoomWebinar struct {
	contact         *connectors.NormalizedContact
	getContactError error
}

func (m *mockConnectorForZoomWebinar) GetContact(ctx context.Context, contactID string) (*connectors.NormalizedContact, error) {
	if m.getContactError != nil {
		return nil, m.getContactError
	}
	return m.contact, nil
}

func (m *mockConnectorForZoomWebinar) GetContacts(ctx context.Context, opts connectors.QueryOptions) (*connectors.ContactList, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *mockConnectorForZoomWebinar) CreateContact(ctx context.Context, input connectors.CreateContactInput) (*connectors.NormalizedContact, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *mockConnectorForZoomWebinar) UpdateContact(ctx context.Context, contactID string, updates connectors.UpdateContactInput) (*connectors.NormalizedContact, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *mockConnectorForZoomWebinar) DeleteContact(ctx context.Context, contactID string) error {
	return fmt.Errorf("not implemented")
}

func (m *mockConnectorForZoomWebinar) GetTags(ctx context.Context) ([]connectors.Tag, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *mockConnectorForZoomWebinar) ApplyTag(ctx context.Context, contactID, tagID string) error {
	return fmt.Errorf("not implemented")
}

func (m *mockConnectorForZoomWebinar) RemoveTag(ctx context.Context, contactID, tagID string) error {
	return fmt.Errorf("not implemented")
}

func (m *mockConnectorForZoomWebinar) GetCustomFields(ctx context.Context) ([]connectors.CustomField, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *mockConnectorForZoomWebinar) GetContactFieldValue(ctx context.Context, contactID, fieldKey string) (interface{}, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *mockConnectorForZoomWebinar) SetContactFieldValue(ctx context.Context, contactID, fieldKey string, value interface{}) error {
	return fmt.Errorf("not implemented")
}

func (m *mockConnectorForZoomWebinar) TriggerAutomation(ctx context.Context, contactID, automationID string) error {
	return fmt.Errorf("not implemented")
}

func (m *mockConnectorForZoomWebinar) AchieveGoal(ctx context.Context, contactID, goalName, integration string) error {
	return fmt.Errorf("not implemented")
}

func (m *mockConnectorForZoomWebinar) TestConnection(ctx context.Context) error {
	return fmt.Errorf("not implemented")
}

func (m *mockConnectorForZoomWebinar) GetMetadata() connectors.ConnectorMetadata {
	return connectors.ConnectorMetadata{PlatformSlug: "mock", PlatformName: "Mock"}
}

func (m *mockConnectorForZoomWebinar) GetCapabilities() []connectors.Capability {
	return []connectors.Capability{}
}

func TestZoomWebinar_GetMetadata(t *testing.T) {
	helper := &ZoomWebinar{}
	if helper.GetName() != "Zoom Webinar" {
		t.Errorf("Expected name 'Zoom Webinar', got '%s'", helper.GetName())
	}
	if helper.GetType() != "zoom_webinar" {
		t.Errorf("Expected type 'zoom_webinar', got '%s'", helper.GetType())
	}
	if helper.GetCategory() != "integration" {
		t.Errorf("Expected category 'integration', got '%s'", helper.GetCategory())
	}
}

func TestZoomWebinar_ValidateConfig_Success(t *testing.T) {
	helper := &ZoomWebinar{}
	err := helper.ValidateConfig(map[string]interface{}{
		"webinar_id": "123456789",
		"api_key": "key123",
		"api_secret": "secret456",
	})
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
}

func TestZoomWebinar_ValidateConfig_MissingWebinarID(t *testing.T) {
	helper := &ZoomWebinar{}
	err := helper.ValidateConfig(map[string]interface{}{"api_key": "key", "api_secret": "secret"})
	if err == nil {
		t.Error("Expected error for missing webinar_id")
	}
}

func TestZoomWebinar_ValidateConfig_MissingAPIKey(t *testing.T) {
	helper := &ZoomWebinar{}
	err := helper.ValidateConfig(map[string]interface{}{"webinar_id": "123", "api_secret": "secret"})
	if err == nil {
		t.Error("Expected error for missing api_key")
	}
}

func TestZoomWebinar_ValidateConfig_MissingAPISecret(t *testing.T) {
	helper := &ZoomWebinar{}
	err := helper.ValidateConfig(map[string]interface{}{"webinar_id": "123", "api_key": "key"})
	if err == nil {
		t.Error("Expected error for missing api_secret")
	}
}

func TestZoomWebinar_Execute_Success(t *testing.T) {
	helper := &ZoomWebinar{}
	mock := &mockConnectorForZoomWebinar{
		contact: &connectors.NormalizedContact{
			ID: "123", FirstName: "John", LastName: "Doe", Email: "john@example.com",
		},
	}

	output, err := helper.Execute(context.Background(), helpers.HelperInput{
		ContactID: "123",
		Config: map[string]interface{}{
			"webinar_id": "123456789",
			"api_key": "key123",
			"api_secret": "secret456",
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

func TestZoomWebinar_Execute_EmptyEmail(t *testing.T) {
	helper := &ZoomWebinar{}
	mock := &mockConnectorForZoomWebinar{
		contact: &connectors.NormalizedContact{ID: "123", Email: ""},
	}

	_, err := helper.Execute(context.Background(), helpers.HelperInput{
		ContactID: "123",
		Config: map[string]interface{}{
			"webinar_id": "123456789",
			"api_key": "key123",
			"api_secret": "secret456",
		},
		Connector: mock,
	})

	if err == nil {
		t.Error("Expected error for empty email")
	}
}

func TestZoomWebinar_Execute_CustomEmailField(t *testing.T) {
	helper := &ZoomWebinar{}
	mock := &mockConnectorForZoomWebinar{
		contact: &connectors.NormalizedContact{
			ID: "123",
			CustomFields: map[string]interface{}{"work_email": "work@test.com"},
		},
	}

	output, err := helper.Execute(context.Background(), helpers.HelperInput{
		ContactID: "123",
		Config: map[string]interface{}{
			"webinar_id": "123456789",
			"api_key": "key123",
			"api_secret": "secret456",
			"email_field": "work_email",
		},
		Connector: mock,
	})

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if output.ModifiedData["email"] != "work@test.com" {
		t.Error("Expected custom email field to be used")
	}
}

func TestZoomWebinar_Execute_CustomNameField(t *testing.T) {
	helper := &ZoomWebinar{}
	mock := &mockConnectorForZoomWebinar{
		contact: &connectors.NormalizedContact{
			ID: "123",
			Email: "test@test.com",
			CustomFields: map[string]interface{}{"nickname": "Johnny"},
		},
	}

	output, err := helper.Execute(context.Background(), helpers.HelperInput{
		ContactID: "123",
		Config: map[string]interface{}{
			"webinar_id": "123456789",
			"api_key": "key123",
			"api_secret": "secret456",
			"name_field": "nickname",
		},
		Connector: mock,
	})

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if output.ModifiedData["first_name"] != "Johnny" {
		t.Error("Expected custom name field to be used")
	}
}

func TestZoomWebinar_Execute_WithCustomQuestions(t *testing.T) {
	helper := &ZoomWebinar{}
	mock := &mockConnectorForZoomWebinar{
		contact: &connectors.NormalizedContact{ID: "123", Email: "test@test.com"},
	}

	output, err := helper.Execute(context.Background(), helpers.HelperInput{
		ContactID: "123",
		Config: map[string]interface{}{
			"webinar_id": "123456789",
			"api_key": "key123",
			"api_secret": "secret456",
			"custom_questions": []interface{}{
				map[string]interface{}{"title": "Company", "value": "Acme"},
			},
		},
		Connector: mock,
	})

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	payload := output.ModifiedData["payload"].(map[string]interface{})
	questions := payload["custom_questions"]
	if questions == nil {
		t.Error("Expected custom_questions in payload")
	}
}

func TestZoomWebinar_Execute_GetContactError(t *testing.T) {
	helper := &ZoomWebinar{}
	mock := &mockConnectorForZoomWebinar{getContactError: fmt.Errorf("contact error")}

	_, err := helper.Execute(context.Background(), helpers.HelperInput{
		ContactID: "123",
		Config: map[string]interface{}{
			"webinar_id": "123456789",
			"api_key": "key123",
			"api_secret": "secret456",
		},
		Connector: mock,
	})

	if err == nil {
		t.Error("Expected error when GetContact fails")
	}
}

func TestZoomWebinar_Execute_APIURLFormat(t *testing.T) {
	helper := &ZoomWebinar{}
	mock := &mockConnectorForZoomWebinar{
		contact: &connectors.NormalizedContact{ID: "123", Email: "test@test.com"},
	}

	output, err := helper.Execute(context.Background(), helpers.HelperInput{
		ContactID: "123",
		Config: map[string]interface{}{
			"webinar_id": "987654321",
			"api_key": "key123",
			"api_secret": "secret456",
		},
		Connector: mock,
	})

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	expectedURL := "https://api.zoom.us/v2/webinars/987654321/registrants"
	if output.ModifiedData["api_url"] != expectedURL {
		t.Errorf("Expected api_url '%s', got: %v", expectedURL, output.ModifiedData["api_url"])
	}
}

func TestZoomWebinar_Execute_AuthType(t *testing.T) {
	helper := &ZoomWebinar{}
	mock := &mockConnectorForZoomWebinar{
		contact: &connectors.NormalizedContact{ID: "123", Email: "test@test.com"},
	}

	output, err := helper.Execute(context.Background(), helpers.HelperInput{
		ContactID: "123",
		Config: map[string]interface{}{
			"webinar_id": "123456789",
			"api_key": "key123",
			"api_secret": "secret456",
		},
		Connector: mock,
	})

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	actionValue := output.Actions[0].Value.(map[string]interface{})
	if actionValue["auth_type"] != "zoom_jwt" {
		t.Errorf("Expected auth_type 'zoom_jwt', got: %v", actionValue["auth_type"])
	}
}
