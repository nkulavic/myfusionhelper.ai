package integration

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/myfusionhelper/api/internal/connectors"
	"github.com/myfusionhelper/api/internal/helpers"
)

type mockConnectorForSlackIt struct {
	contact  *connectors.NormalizedContact
	getError error
}

func (m *mockConnectorForSlackIt) GetContact(ctx context.Context, contactID string) (*connectors.NormalizedContact, error) {
	if m.getError != nil {
		return nil, m.getError
	}
	return m.contact, nil
}

func (m *mockConnectorForSlackIt) GetContacts(ctx context.Context, opts connectors.QueryOptions) (*connectors.ContactList, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForSlackIt) CreateContact(ctx context.Context, contact connectors.CreateContactInput) (*connectors.NormalizedContact, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForSlackIt) UpdateContact(ctx context.Context, contactID string, updates connectors.UpdateContactInput) (*connectors.NormalizedContact, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForSlackIt) DeleteContact(ctx context.Context, contactID string) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorForSlackIt) GetTags(ctx context.Context) ([]connectors.Tag, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForSlackIt) ApplyTag(ctx context.Context, contactID, tagID string) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorForSlackIt) RemoveTag(ctx context.Context, contactID, tagID string) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorForSlackIt) GetCustomFields(ctx context.Context) ([]connectors.CustomField, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForSlackIt) GetContactFieldValue(ctx context.Context, contactID, fieldKey string) (interface{}, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForSlackIt) SetContactFieldValue(ctx context.Context, contactID, fieldKey string, value interface{}) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorForSlackIt) TriggerAutomation(ctx context.Context, contactID, automationID string) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorForSlackIt) AchieveGoal(ctx context.Context, contactID, goalName, integration string) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorForSlackIt) TestConnection(ctx context.Context) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorForSlackIt) GetMetadata() connectors.ConnectorMetadata {
	return connectors.ConnectorMetadata{}
}
func (m *mockConnectorForSlackIt) GetCapabilities() []connectors.Capability {
	return nil
}

func TestSlackIt_GetMetadata(t *testing.T) {
	h := &SlackIt{}
	if h.GetName() != "Slack It" {
		t.Error("wrong name")
	}
	if h.GetType() != "slack_it" {
		t.Error("wrong type")
	}
	if h.GetCategory() != "integration" {
		t.Error("wrong category")
	}
	if !h.RequiresCRM() {
		t.Error("should require CRM")
	}
}

func TestSlackIt_ValidateConfig_MissingWebhook(t *testing.T) {
	err := (&SlackIt{}).ValidateConfig(map[string]interface{}{
		"message":  "Test message",
		"username": "Bot",
	})
	if err == nil {
		t.Error("should error on missing webhook")
	}
}

func TestSlackIt_ValidateConfig_MissingMessage(t *testing.T) {
	err := (&SlackIt{}).ValidateConfig(map[string]interface{}{
		"webhook":  "https://hooks.slack.com/test",
		"username": "Bot",
	})
	if err == nil {
		t.Error("should error on missing message")
	}
}

func TestSlackIt_ValidateConfig_MissingUsername(t *testing.T) {
	err := (&SlackIt{}).ValidateConfig(map[string]interface{}{
		"webhook": "https://hooks.slack.com/test",
		"message": "Test message",
	})
	if err == nil {
		t.Error("should error on missing username")
	}
}

func TestSlackIt_ValidateConfig_EmptyWebhook(t *testing.T) {
	err := (&SlackIt{}).ValidateConfig(map[string]interface{}{
		"webhook":  "",
		"message":  "Test",
		"username": "Bot",
	})
	if err == nil {
		t.Error("should error on empty webhook")
	}
}

func TestSlackIt_ValidateConfig_Valid(t *testing.T) {
	err := (&SlackIt{}).ValidateConfig(map[string]interface{}{
		"webhook":  "https://hooks.slack.com/test",
		"message":  "Test message",
		"username": "Bot",
	})
	if err != nil {
		t.Errorf("should be valid: %v", err)
	}
}

func TestSlackIt_Execute_BasicMessage(t *testing.T) {
	mock := &mockConnectorForSlackIt{
		contact: &connectors.NormalizedContact{
			ID:        "contact_123",
			FirstName: "John",
			Email:     "john@example.com",
		},
	}

	output, err := (&SlackIt{}).Execute(context.Background(), helpers.HelperInput{
		ContactID: "contact_123",
		Config: map[string]interface{}{
			"webhook":  "https://hooks.slack.com/test",
			"message":  "New contact registered!",
			"username": "CRM Bot",
		},
		Connector: mock,
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
}

func TestSlackIt_Execute_CurlyBraceMergeFields(t *testing.T) {
	mock := &mockConnectorForSlackIt{
		contact: &connectors.NormalizedContact{
			ID:        "contact_123",
			FirstName: "Jane",
			LastName:  "Smith",
			Email:     "jane@example.com",
			Company:   "ACME Corp",
		},
	}

	output, err := (&SlackIt{}).Execute(context.Background(), helpers.HelperInput{
		ContactID: "contact_123",
		Config: map[string]interface{}{
			"webhook":  "https://hooks.slack.com/test",
			"message":  "New contact: {{FirstName}} {{LastName}} from {{Company}}",
			"username": "Bot",
		},
		Connector: mock,
	})

	if err != nil {
		t.Fatal(err)
	}
	if !output.Success {
		t.Error("should succeed")
	}

	actionValue := output.Actions[0].Value.(map[string]interface{})
	payload := actionValue["payload"].(map[string]interface{})
	expectedMessage := "New contact: Jane Smith from ACME Corp"
	if payload["text"] != expectedMessage {
		t.Errorf("expected message '%s', got %v", expectedMessage, payload["text"])
	}
}

func TestSlackIt_Execute_AtSignMergeFields(t *testing.T) {
	mock := &mockConnectorForSlackIt{
		contact: &connectors.NormalizedContact{
			ID:        "contact_123",
			FirstName: "Bob",
			Email:     "bob@example.com",
		},
	}

	output, err := (&SlackIt{}).Execute(context.Background(), helpers.HelperInput{
		ContactID: "contact_123",
		Config: map[string]interface{}{
			"webhook":  "https://hooks.slack.com/test",
			"message":  "Contact @FirstName has email @Email",
			"username": "Bot",
		},
		Connector: mock,
	})

	if err != nil {
		t.Fatal(err)
	}

	actionValue := output.Actions[0].Value.(map[string]interface{})
	payload := actionValue["payload"].(map[string]interface{})
	expectedMessage := "Contact Bob has email bob@example.com"
	if payload["text"] != expectedMessage {
		t.Errorf("expected message '%s', got %v", expectedMessage, payload["text"])
	}
}

func TestSlackIt_Execute_BothMergeSyntaxes(t *testing.T) {
	mock := &mockConnectorForSlackIt{
		contact: &connectors.NormalizedContact{
			ID:        "contact_123",
			FirstName: "Alice",
			LastName:  "Johnson",
			Email:     "alice@example.com",
		},
	}

	output, err := (&SlackIt{}).Execute(context.Background(), helpers.HelperInput{
		ContactID: "contact_123",
		Config: map[string]interface{}{
			"webhook":  "https://hooks.slack.com/test",
			"message":  "{{FirstName}} @LastName registered with @Email",
			"username": "Bot",
		},
		Connector: mock,
	})

	if err != nil {
		t.Fatal(err)
	}

	actionValue := output.Actions[0].Value.(map[string]interface{})
	payload := actionValue["payload"].(map[string]interface{})
	expectedMessage := "Alice Johnson registered with alice@example.com"
	if payload["text"] != expectedMessage {
		t.Errorf("expected message '%s', got %v", expectedMessage, payload["text"])
	}
}

func TestSlackIt_Execute_CustomFields(t *testing.T) {
	mock := &mockConnectorForSlackIt{
		contact: &connectors.NormalizedContact{
			ID:    "contact_123",
			Email: "test@example.com",
			CustomFields: map[string]interface{}{
				"Status": "Premium",
			},
		},
	}

	output, err := (&SlackIt{}).Execute(context.Background(), helpers.HelperInput{
		ContactID: "contact_123",
		Config: map[string]interface{}{
			"webhook":  "https://hooks.slack.com/test",
			"message":  "Contact status: {{Status}}",
			"username": "Bot",
		},
		Connector: mock,
	})

	if err != nil {
		t.Fatal(err)
	}

	actionValue := output.Actions[0].Value.(map[string]interface{})
	payload := actionValue["payload"].(map[string]interface{})
	if payload["text"] != "Contact status: Premium" {
		t.Errorf("should merge custom field, got %v", payload["text"])
	}
}

func TestSlackIt_Execute_WithChannel(t *testing.T) {
	mock := &mockConnectorForSlackIt{
		contact: &connectors.NormalizedContact{
			ID:    "contact_123",
			Email: "test@example.com",
		},
	}

	output, err := (&SlackIt{}).Execute(context.Background(), helpers.HelperInput{
		ContactID: "contact_123",
		Config: map[string]interface{}{
			"webhook":  "https://hooks.slack.com/test",
			"message":  "Test message",
			"username": "Bot",
			"channel":  "#notifications",
		},
		Connector: mock,
	})

	if err != nil {
		t.Fatal(err)
	}

	actionValue := output.Actions[0].Value.(map[string]interface{})
	payload := actionValue["payload"].(map[string]interface{})
	if payload["channel"] != "#notifications" {
		t.Errorf("expected channel #notifications, got %v", payload["channel"])
	}
}

func TestSlackIt_Execute_WithIconEmoji(t *testing.T) {
	mock := &mockConnectorForSlackIt{
		contact: &connectors.NormalizedContact{
			ID:    "contact_123",
			Email: "test@example.com",
		},
	}

	output, err := (&SlackIt{}).Execute(context.Background(), helpers.HelperInput{
		ContactID: "contact_123",
		Config: map[string]interface{}{
			"webhook":    "https://hooks.slack.com/test",
			"message":    "Test message",
			"username":   "Bot",
			"icon_emoji": ":robot_face:",
		},
		Connector: mock,
	})

	if err != nil {
		t.Fatal(err)
	}

	actionValue := output.Actions[0].Value.(map[string]interface{})
	payload := actionValue["payload"].(map[string]interface{})
	if payload["icon_emoji"] != ":robot_face:" {
		t.Errorf("expected icon_emoji :robot_face:, got %v", payload["icon_emoji"])
	}
}

func TestSlackIt_Execute_GetContactError(t *testing.T) {
	mock := &mockConnectorForSlackIt{
		getError: fmt.Errorf("API error"),
	}

	output, err := (&SlackIt{}).Execute(context.Background(), helpers.HelperInput{
		ContactID: "contact_123",
		Config: map[string]interface{}{
			"webhook":  "https://hooks.slack.com/test",
			"message":  "Test",
			"username": "Bot",
		},
		Connector: mock,
	})

	if err == nil {
		t.Error("should return error on GetContact failure")
	}
	if output.Success {
		t.Error("should not succeed")
	}
}

func TestSlackIt_Execute_FullNameMergeField(t *testing.T) {
	mock := &mockConnectorForSlackIt{
		contact: &connectors.NormalizedContact{
			ID:        "contact_123",
			FirstName: "John",
			LastName:  "Doe",
			Email:     "john@example.com",
		},
	}

	output, err := (&SlackIt{}).Execute(context.Background(), helpers.HelperInput{
		ContactID: "contact_123",
		Config: map[string]interface{}{
			"webhook":  "https://hooks.slack.com/test",
			"message":  "Welcome {{full_name}}!",
			"username": "Bot",
		},
		Connector: mock,
	})

	if err != nil {
		t.Fatal(err)
	}

	actionValue := output.Actions[0].Value.(map[string]interface{})
	payload := actionValue["payload"].(map[string]interface{})
	if !strings.Contains(payload["text"].(string), "John Doe") {
		t.Error("should contain full name")
	}
}

func TestSlackIt_Execute_WebhookURL(t *testing.T) {
	mock := &mockConnectorForSlackIt{
		contact: &connectors.NormalizedContact{
			ID:    "contact_123",
			Email: "test@example.com",
		},
	}

	webhookURL := "https://hooks.slack.com/services/T00/B00/XXX"
	output, err := (&SlackIt{}).Execute(context.Background(), helpers.HelperInput{
		ContactID: "contact_123",
		Config: map[string]interface{}{
			"webhook":  webhookURL,
			"message":  "Test",
			"username": "Bot",
		},
		Connector: mock,
	})

	if err != nil {
		t.Fatal(err)
	}

	if output.Actions[0].Target != webhookURL {
		t.Error("action target should be webhook URL")
	}

	actionValue := output.Actions[0].Value.(map[string]interface{})
	if actionValue["url"] != webhookURL {
		t.Error("action value should include webhook URL")
	}
	if actionValue["method"] != "POST" {
		t.Error("action should use POST method")
	}
}

func TestSlackIt_Execute_ModifiedDataStructure(t *testing.T) {
	mock := &mockConnectorForSlackIt{
		contact: &connectors.NormalizedContact{
			ID:        "contact_123",
			FirstName: "Test",
			Email:     "test@example.com",
		},
	}

	output, err := (&SlackIt{}).Execute(context.Background(), helpers.HelperInput{
		ContactID: "contact_123",
		Config: map[string]interface{}{
			"webhook":    "https://hooks.slack.com/test",
			"message":    "Hello {{FirstName}}",
			"username":   "Bot",
			"channel":    "#general",
			"icon_emoji": ":tada:",
		},
		Connector: mock,
	})

	if err != nil {
		t.Fatal(err)
	}

	modData := output.ModifiedData
	if modData["webhook"] != "https://hooks.slack.com/test" {
		t.Error("modified data should include webhook")
	}
	if modData["message"] != "Hello Test" {
		t.Error("modified data should include processed message")
	}
	if modData["username"] != "Bot" {
		t.Error("modified data should include username")
	}
	if modData["channel"] != "#general" {
		t.Error("modified data should include channel")
	}
	if modData["payload"] == nil {
		t.Error("modified data should include payload")
	}
}
