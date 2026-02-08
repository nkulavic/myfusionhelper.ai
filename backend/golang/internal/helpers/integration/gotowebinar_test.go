package integration

import (
	"context"
	"fmt"
	"testing"

	"github.com/myfusionhelper/api/internal/connectors"
	"github.com/myfusionhelper/api/internal/helpers"
)

type mockConnectorForGoToWebinar struct {
	contact       *connectors.NormalizedContact
	getError      error
	tagsApplied   []string
	applyTagError error
}

func (m *mockConnectorForGoToWebinar) GetContact(ctx context.Context, contactID string) (*connectors.NormalizedContact, error) {
	if m.getError != nil {
		return nil, m.getError
	}
	return m.contact, nil
}

func (m *mockConnectorForGoToWebinar) ApplyTag(ctx context.Context, contactID, tagID string) error {
	if m.applyTagError != nil {
		return m.applyTagError
	}
	if m.tagsApplied == nil {
		m.tagsApplied = make([]string, 0)
	}
	m.tagsApplied = append(m.tagsApplied, tagID)
	return nil
}

func (m *mockConnectorForGoToWebinar) GetContacts(ctx context.Context, opts connectors.QueryOptions) (*connectors.ContactList, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForGoToWebinar) CreateContact(ctx context.Context, contact connectors.CreateContactInput) (*connectors.NormalizedContact, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForGoToWebinar) UpdateContact(ctx context.Context, contactID string, updates connectors.UpdateContactInput) (*connectors.NormalizedContact, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForGoToWebinar) DeleteContact(ctx context.Context, contactID string) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorForGoToWebinar) GetTags(ctx context.Context) ([]connectors.Tag, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForGoToWebinar) RemoveTag(ctx context.Context, contactID, tagID string) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorForGoToWebinar) GetCustomFields(ctx context.Context) ([]connectors.CustomField, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForGoToWebinar) GetContactFieldValue(ctx context.Context, contactID, fieldKey string) (interface{}, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForGoToWebinar) SetContactFieldValue(ctx context.Context, contactID, fieldKey string, value interface{}) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorForGoToWebinar) TriggerAutomation(ctx context.Context, contactID, automationID string) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorForGoToWebinar) AchieveGoal(ctx context.Context, contactID, goalName, integration string) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorForGoToWebinar) TestConnection(ctx context.Context) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorForGoToWebinar) GetMetadata() connectors.ConnectorMetadata {
	return connectors.ConnectorMetadata{}
}
func (m *mockConnectorForGoToWebinar) GetCapabilities() []connectors.Capability {
	return nil
}

func TestGoToWebinar_GetMetadata(t *testing.T) {
	h := &GoToWebinar{}
	if h.GetName() != "GoToWebinar" {
		t.Error("wrong name")
	}
	if h.GetType() != "gotowebinar" {
		t.Error("wrong type")
	}
	if h.GetCategory() != "integration" {
		t.Error("wrong category")
	}
	if !h.RequiresCRM() {
		t.Error("should require CRM")
	}
}

func TestGoToWebinar_ValidateConfig_MissingOrganizerKey(t *testing.T) {
	err := (&GoToWebinar{}).ValidateConfig(map[string]interface{}{
		"webinar_key": "123456",
	})
	if err == nil {
		t.Error("should error on missing organizer_key")
	}
}

func TestGoToWebinar_ValidateConfig_MissingWebinarKey(t *testing.T) {
	err := (&GoToWebinar{}).ValidateConfig(map[string]interface{}{
		"organizer_key": "org123",
	})
	if err == nil {
		t.Error("should error on missing webinar_key")
	}
}

func TestGoToWebinar_ValidateConfig_EmptyKeys(t *testing.T) {
	err := (&GoToWebinar{}).ValidateConfig(map[string]interface{}{
		"organizer_key": "",
		"webinar_key":   "",
	})
	if err == nil {
		t.Error("should error on empty keys")
	}
}

func TestGoToWebinar_ValidateConfig_Valid(t *testing.T) {
	err := (&GoToWebinar{}).ValidateConfig(map[string]interface{}{
		"organizer_key": "org123",
		"webinar_key":   "web456",
	})
	if err != nil {
		t.Errorf("should be valid: %v", err)
	}
}

func TestGoToWebinar_Execute_Success(t *testing.T) {
	mock := &mockConnectorForGoToWebinar{
		contact: &connectors.NormalizedContact{
			ID:        "contact_123",
			FirstName: "John",
			LastName:  "Doe",
			Email:     "john@example.com",
		},
	}

	output, err := (&GoToWebinar{}).Execute(context.Background(), helpers.HelperInput{
		ContactID: "contact_123",
		Config: map[string]interface{}{
			"organizer_key": "org123",
			"webinar_key":   "web456",
		},
		Connector: mock,
		ServiceAuths: map[string]*connectors.ConnectorConfig{
			"gotowebinar": {
				AccessToken: "test_access_token",
			},
		},
	})

	if err != nil {
		t.Fatal(err)
	}
	if !output.Success {
		t.Error("should succeed")
	}
	if len(output.Actions) != 1 {
		t.Fatalf("expected 1 action, got %d", len(output.Actions))
	}
	if output.Actions[0].Type != "webhook_queued" {
		t.Errorf("expected webhook_queued action, got %s", output.Actions[0].Type)
	}

	// Verify API URL
	actionValue, ok := output.Actions[0].Value.(map[string]interface{})
	if !ok {
		t.Fatal("action value should be a map")
	}
	expectedURL := "https://api.getgo.com/G2W/rest/v2/organizers/org123/webinars/web456/registrants"
	if actionValue["url"] != expectedURL {
		t.Errorf("expected URL %s, got %v", expectedURL, actionValue["url"])
	}
}

func TestGoToWebinar_Execute_MissingServiceAuth(t *testing.T) {
	mock := &mockConnectorForGoToWebinar{
		contact: &connectors.NormalizedContact{
			ID:    "contact_123",
			Email: "test@example.com",
		},
	}

	output, err := (&GoToWebinar{}).Execute(context.Background(), helpers.HelperInput{
		ContactID: "contact_123",
		Config: map[string]interface{}{
			"organizer_key": "org123",
			"webinar_key":   "web456",
		},
		Connector:    mock,
		ServiceAuths: map[string]*connectors.ConnectorConfig{},
	})

	if err == nil {
		t.Error("should return error on missing service auth")
	}
	if output.Success {
		t.Error("should not succeed")
	}
}

func TestGoToWebinar_Execute_MissingAuthToken(t *testing.T) {
	mock := &mockConnectorForGoToWebinar{
		contact: &connectors.NormalizedContact{
			ID:    "contact_123",
			Email: "test@example.com",
		},
	}

	output, err := (&GoToWebinar{}).Execute(context.Background(), helpers.HelperInput{
		ContactID: "contact_123",
		Config: map[string]interface{}{
			"organizer_key": "org123",
			"webinar_key":   "web456",
		},
		Connector: mock,
		ServiceAuths: map[string]*connectors.ConnectorConfig{
			"gotowebinar": {
				AccessToken: "",
				APIKey:      "",
			},
		},
	})

	if err == nil {
		t.Error("should return error on missing auth token")
	}
	if output.Success {
		t.Error("should not succeed")
	}
}

func TestGoToWebinar_Execute_UsesAPIKeyWhenNoAccessToken(t *testing.T) {
	mock := &mockConnectorForGoToWebinar{
		contact: &connectors.NormalizedContact{
			ID:    "contact_123",
			Email: "test@example.com",
		},
	}

	output, err := (&GoToWebinar{}).Execute(context.Background(), helpers.HelperInput{
		ContactID: "contact_123",
		Config: map[string]interface{}{
			"organizer_key": "org123",
			"webinar_key":   "web456",
		},
		Connector: mock,
		ServiceAuths: map[string]*connectors.ConnectorConfig{
			"gotowebinar": {
				AccessToken: "",
				APIKey:      "test_api_key",
			},
		},
	})

	if err != nil {
		t.Fatal(err)
	}
	if !output.Success {
		t.Error("should succeed with API key")
	}

	// Verify Authorization header uses API key
	actionValue := output.Actions[0].Value.(map[string]interface{})
	headers := actionValue["headers"].(map[string]string)
	if headers["Authorization"] != "Bearer test_api_key" {
		t.Errorf("expected Bearer test_api_key, got %s", headers["Authorization"])
	}
}

func TestGoToWebinar_Execute_MissingEmail(t *testing.T) {
	mock := &mockConnectorForGoToWebinar{
		contact: &connectors.NormalizedContact{
			ID:        "contact_123",
			FirstName: "John",
			LastName:  "Doe",
			Email:     "",
		},
	}

	output, err := (&GoToWebinar{}).Execute(context.Background(), helpers.HelperInput{
		ContactID: "contact_123",
		Config: map[string]interface{}{
			"organizer_key": "org123",
			"webinar_key":   "web456",
		},
		Connector: mock,
		ServiceAuths: map[string]*connectors.ConnectorConfig{
			"gotowebinar": {
				AccessToken: "test_token",
			},
		},
	})

	if err == nil {
		t.Error("should return error when contact email is empty")
	}
	if output.Success {
		t.Error("should not succeed")
	}
}

func TestGoToWebinar_Execute_GetContactError(t *testing.T) {
	mock := &mockConnectorForGoToWebinar{
		getError: fmt.Errorf("API error"),
	}

	output, err := (&GoToWebinar{}).Execute(context.Background(), helpers.HelperInput{
		ContactID: "contact_123",
		Config: map[string]interface{}{
			"organizer_key": "org123",
			"webinar_key":   "web456",
		},
		Connector: mock,
		ServiceAuths: map[string]*connectors.ConnectorConfig{
			"gotowebinar": {
				AccessToken: "test_token",
			},
		},
	})

	if err == nil {
		t.Error("should return error on GetContact failure")
	}
	if output.Success {
		t.Error("should not succeed")
	}
}

func TestGoToWebinar_Execute_WithContactData(t *testing.T) {
	mock := &mockConnectorForGoToWebinar{}

	contactData := &connectors.NormalizedContact{
		ID:        "contact_123",
		FirstName: "Jane",
		LastName:  "Smith",
		Email:     "jane@example.com",
	}

	output, err := (&GoToWebinar{}).Execute(context.Background(), helpers.HelperInput{
		ContactID:   "contact_123",
		ContactData: contactData,
		Config: map[string]interface{}{
			"organizer_key": "org123",
			"webinar_key":   "web456",
		},
		Connector: mock,
		ServiceAuths: map[string]*connectors.ConnectorConfig{
			"gotowebinar": {
				AccessToken: "test_token",
			},
		},
	})

	if err != nil {
		t.Fatal(err)
	}
	if !output.Success {
		t.Error("should succeed with pre-provided contact data")
	}

	// Verify payload contains contact data
	actionValue := output.Actions[0].Value.(map[string]interface{})
	payload := actionValue["payload"].(map[string]interface{})
	if payload["email"] != "jane@example.com" {
		t.Error("payload should contain contact email")
	}
	if payload["firstName"] != "Jane" {
		t.Error("payload should contain first name")
	}
}

func TestGoToWebinar_Execute_WithApplyTag(t *testing.T) {
	mock := &mockConnectorForGoToWebinar{
		contact: &connectors.NormalizedContact{
			ID:    "contact_123",
			Email: "test@example.com",
		},
	}

	output, err := (&GoToWebinar{}).Execute(context.Background(), helpers.HelperInput{
		ContactID: "contact_123",
		Config: map[string]interface{}{
			"organizer_key": "org123",
			"webinar_key":   "web456",
			"apply_tag":     "tag_registered",
		},
		Connector: mock,
		ServiceAuths: map[string]*connectors.ConnectorConfig{
			"gotowebinar": {
				AccessToken: "test_token",
			},
		},
	})

	if err != nil {
		t.Fatal(err)
	}
	if !output.Success {
		t.Error("should succeed")
	}
	if len(mock.tagsApplied) != 1 || mock.tagsApplied[0] != "tag_registered" {
		t.Errorf("expected tag_registered to be applied, got %v", mock.tagsApplied)
	}
	if len(output.Actions) != 2 {
		t.Errorf("expected 2 actions (webhook + tag), got %d", len(output.Actions))
	}
}

func TestGoToWebinar_Execute_ApplyTagError(t *testing.T) {
	mock := &mockConnectorForGoToWebinar{
		contact: &connectors.NormalizedContact{
			ID:    "contact_123",
			Email: "test@example.com",
		},
		applyTagError: fmt.Errorf("tag error"),
	}

	output, err := (&GoToWebinar{}).Execute(context.Background(), helpers.HelperInput{
		ContactID: "contact_123",
		Config: map[string]interface{}{
			"organizer_key": "org123",
			"webinar_key":   "web456",
			"apply_tag":     "tag_registered",
		},
		Connector: mock,
		ServiceAuths: map[string]*connectors.ConnectorConfig{
			"gotowebinar": {
				AccessToken: "test_token",
			},
		},
	})

	if err != nil {
		t.Fatal(err)
	}
	// Should still succeed even if tag application fails
	if !output.Success {
		t.Error("should succeed even if tag fails")
	}
	if len(output.Actions) != 1 {
		t.Error("should only have webhook action (tag failed)")
	}
}

func TestGoToWebinar_Execute_PayloadStructure(t *testing.T) {
	mock := &mockConnectorForGoToWebinar{
		contact: &connectors.NormalizedContact{
			ID:        "contact_123",
			FirstName: "Alice",
			LastName:  "Johnson",
			Email:     "alice@example.com",
		},
	}

	output, err := (&GoToWebinar{}).Execute(context.Background(), helpers.HelperInput{
		ContactID: "contact_123",
		Config: map[string]interface{}{
			"organizer_key": "org999",
			"webinar_key":   "web888",
		},
		Connector: mock,
		ServiceAuths: map[string]*connectors.ConnectorConfig{
			"gotowebinar": {
				AccessToken: "bearer_token_xyz",
			},
		},
	})

	if err != nil {
		t.Fatal(err)
	}

	actionValue := output.Actions[0].Value.(map[string]interface{})
	payload := actionValue["payload"].(map[string]interface{})

	if payload["firstName"] != "Alice" {
		t.Error("payload should have firstName")
	}
	if payload["lastName"] != "Johnson" {
		t.Error("payload should have lastName")
	}
	if payload["email"] != "alice@example.com" {
		t.Error("payload should have email")
	}

	headers := actionValue["headers"].(map[string]string)
	if headers["Authorization"] != "Bearer bearer_token_xyz" {
		t.Error("headers should include Bearer authorization")
	}
	if headers["Content-Type"] != "application/json" {
		t.Error("headers should include JSON content type")
	}
}
