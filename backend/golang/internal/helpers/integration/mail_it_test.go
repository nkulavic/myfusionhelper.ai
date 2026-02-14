package integration

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/myfusionhelper/api/internal/connectors"
	"github.com/myfusionhelper/api/internal/helpers"
)

type mockConnectorForMailIt struct {
	contact  *connectors.NormalizedContact
	getError error
}

func (m *mockConnectorForMailIt) GetContact(ctx context.Context, contactID string) (*connectors.NormalizedContact, error) {
	if m.getError != nil {
		return nil, m.getError
	}
	return m.contact, nil
}

func (m *mockConnectorForMailIt) GetContacts(ctx context.Context, opts connectors.QueryOptions) (*connectors.ContactList, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForMailIt) CreateContact(ctx context.Context, contact connectors.CreateContactInput) (*connectors.NormalizedContact, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForMailIt) UpdateContact(ctx context.Context, contactID string, updates connectors.UpdateContactInput) (*connectors.NormalizedContact, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForMailIt) DeleteContact(ctx context.Context, contactID string) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorForMailIt) GetTags(ctx context.Context) ([]connectors.Tag, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForMailIt) ApplyTag(ctx context.Context, contactID, tagID string) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorForMailIt) RemoveTag(ctx context.Context, contactID, tagID string) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorForMailIt) GetCustomFields(ctx context.Context) ([]connectors.CustomField, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForMailIt) GetContactFieldValue(ctx context.Context, contactID, fieldKey string) (interface{}, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForMailIt) SetContactFieldValue(ctx context.Context, contactID, fieldKey string, value interface{}) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorForMailIt) TriggerAutomation(ctx context.Context, contactID, automationID string) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorForMailIt) AchieveGoal(ctx context.Context, contactID, goalName, integration string) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorForMailIt) TestConnection(ctx context.Context) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorForMailIt) GetMetadata() connectors.ConnectorMetadata {
	return connectors.ConnectorMetadata{}
}
func (m *mockConnectorForMailIt) GetCapabilities() []connectors.Capability {
	return nil
}

func TestMailIt_GetMetadata(t *testing.T) {
	h := &MailIt{}
	if h.GetName() != "Mail It" {
		t.Error("wrong name")
	}
	if h.GetType() != "mail_it" {
		t.Error("wrong type")
	}
	if h.GetCategory() != "integration" {
		t.Error("wrong category")
	}
	if !h.RequiresCRM() {
		t.Error("should require CRM")
	}
}

func TestMailIt_ValidateConfig_MissingSubject(t *testing.T) {
	err := (&MailIt{}).ValidateConfig(map[string]interface{}{
		"body_template": "Body",
		"from_name":     "Sender",
		"from_email":    "sender@example.com",
	})
	if err == nil {
		t.Error("should error on missing subject_template")
	}
}

func TestMailIt_ValidateConfig_MissingBody(t *testing.T) {
	err := (&MailIt{}).ValidateConfig(map[string]interface{}{
		"subject_template": "Subject",
		"from_name":        "Sender",
		"from_email":       "sender@example.com",
	})
	if err == nil {
		t.Error("should error on missing body_template")
	}
}

func TestMailIt_ValidateConfig_MissingFromName(t *testing.T) {
	err := (&MailIt{}).ValidateConfig(map[string]interface{}{
		"subject_template": "Subject",
		"body_template":    "Body",
		"from_email":       "sender@example.com",
	})
	if err == nil {
		t.Error("should error on missing from_name")
	}
}

func TestMailIt_ValidateConfig_MissingFromEmail(t *testing.T) {
	err := (&MailIt{}).ValidateConfig(map[string]interface{}{
		"subject_template": "Subject",
		"body_template":    "Body",
		"from_name":        "Sender",
	})
	if err == nil {
		t.Error("should error on missing from_email")
	}
}

func TestMailIt_ValidateConfig_Valid(t *testing.T) {
	err := (&MailIt{}).ValidateConfig(map[string]interface{}{
		"subject_template": "Subject",
		"body_template":    "Body",
		"from_name":        "Sender",
		"from_email":       "sender@example.com",
	})
	if err != nil {
		t.Errorf("should be valid: %v", err)
	}
}

func TestMailIt_Execute_BasicEmail(t *testing.T) {
	mock := &mockConnectorForMailIt{
		contact: &connectors.NormalizedContact{
			ID:        "contact_123",
			FirstName: "John",
			LastName:  "Doe",
			Email:     "john@example.com",
		},
	}

	output, err := (&MailIt{}).Execute(context.Background(), helpers.HelperInput{
		ContactID: "contact_123",
		Config: map[string]interface{}{
			"subject_template": "Welcome",
			"body_template":    "Hello, welcome to our service!",
			"from_name":        "Support Team",
			"from_email":       "support@company.com",
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
	if output.Actions[0].Type != "email_queued" {
		t.Errorf("expected email_queued action, got %s", output.Actions[0].Type)
	}
	if output.Actions[0].Target != "john@example.com" {
		t.Errorf("expected target john@example.com, got %s", output.Actions[0].Target)
	}
}

func TestMailIt_Execute_MergeFields(t *testing.T) {
	mock := &mockConnectorForMailIt{
		contact: &connectors.NormalizedContact{
			ID:        "contact_123",
			FirstName: "Jane",
			LastName:  "Smith",
			Email:     "jane@example.com",
			Company:   "ACME Corp",
		},
	}

	output, err := (&MailIt{}).Execute(context.Background(), helpers.HelperInput{
		ContactID: "contact_123",
		Config: map[string]interface{}{
			"subject_template": "Hello {{FirstName}}!",
			"body_template":    "Dear {{FirstName}} {{LastName}}, welcome to {{Company}}",
			"from_name":        "Team",
			"from_email":       "team@example.com",
		},
		Connector: mock,
	})

	if err != nil {
		t.Fatal(err)
	}
	if !output.Success {
		t.Error("should succeed")
	}

	payload := output.Actions[0].Value.(map[string]interface{})
	if payload["subject"] != "Hello Jane!" {
		t.Errorf("expected subject 'Hello Jane!', got %v", payload["subject"])
	}
	if payload["body"] != "Dear Jane Smith, welcome to ACME Corp" {
		t.Errorf("wrong body: %v", payload["body"])
	}
}

func TestMailIt_Execute_CustomFields(t *testing.T) {
	mock := &mockConnectorForMailIt{
		contact: &connectors.NormalizedContact{
			ID:    "contact_123",
			Email: "test@example.com",
			CustomFields: map[string]interface{}{
				"VIPLevel": "Gold",
			},
		},
	}

	output, err := (&MailIt{}).Execute(context.Background(), helpers.HelperInput{
		ContactID: "contact_123",
		Config: map[string]interface{}{
			"subject_template": "Your {{VIPLevel}} Status",
			"body_template":    "You are a {{VIPLevel}} member",
			"from_name":        "VIP Team",
			"from_email":       "vip@example.com",
		},
		Connector: mock,
	})

	if err != nil {
		t.Fatal(err)
	}
	if !output.Success {
		t.Error("should succeed")
	}

	payload := output.Actions[0].Value.(map[string]interface{})
	if payload["subject"] != "Your Gold Status" {
		t.Errorf("expected subject 'Your Gold Status', got %v", payload["subject"])
	}
}

func TestMailIt_Execute_CustomToField(t *testing.T) {
	mock := &mockConnectorForMailIt{
		contact: &connectors.NormalizedContact{
			ID:    "contact_123",
			Email: "john@example.com",
			CustomFields: map[string]interface{}{
				"AlternateEmail": "jane@example.com",
			},
		},
	}

	output, err := (&MailIt{}).Execute(context.Background(), helpers.HelperInput{
		ContactID: "contact_123",
		Config: map[string]interface{}{
			"to_field":         "AlternateEmail",
			"subject_template": "Test",
			"body_template":    "Body",
			"from_name":        "Sender",
			"from_email":       "sender@example.com",
		},
		Connector: mock,
	})

	if err != nil {
		t.Fatal(err)
	}
	if !output.Success {
		t.Error("should succeed")
	}

	payload := output.Actions[0].Value.(map[string]interface{})
	if payload["to"] != "jane@example.com" {
		t.Errorf("expected to jane@example.com, got %v", payload["to"])
	}
}

func TestMailIt_Execute_EmptyToField(t *testing.T) {
	mock := &mockConnectorForMailIt{
		contact: &connectors.NormalizedContact{
			ID:    "contact_123",
			Email: "",
		},
	}

	output, err := (&MailIt{}).Execute(context.Background(), helpers.HelperInput{
		ContactID: "contact_123",
		Config: map[string]interface{}{
			"subject_template": "Test",
			"body_template":    "Body",
			"from_name":        "Sender",
			"from_email":       "sender@example.com",
		},
		Connector: mock,
	})

	if err == nil {
		t.Error("should return error when recipient email is empty")
	}
	if output.Success {
		t.Error("should not succeed")
	}
}

func TestMailIt_Execute_WithReplyTo(t *testing.T) {
	mock := &mockConnectorForMailIt{
		contact: &connectors.NormalizedContact{
			ID:    "contact_123",
			Email: "test@example.com",
		},
	}

	output, err := (&MailIt{}).Execute(context.Background(), helpers.HelperInput{
		ContactID: "contact_123",
		Config: map[string]interface{}{
			"subject_template": "Test",
			"body_template":    "Body",
			"from_name":        "Sender",
			"from_email":       "sender@example.com",
			"reply_to":         "replyto@example.com",
		},
		Connector: mock,
	})

	if err != nil {
		t.Fatal(err)
	}
	if !output.Success {
		t.Error("should succeed")
	}

	payload := output.Actions[0].Value.(map[string]interface{})
	if payload["reply_to"] != "replyto@example.com" {
		t.Errorf("expected reply_to, got %v", payload["reply_to"])
	}
}

func TestMailIt_Execute_PlainTextContentType(t *testing.T) {
	mock := &mockConnectorForMailIt{
		contact: &connectors.NormalizedContact{
			ID:    "contact_123",
			Email: "test@example.com",
		},
	}

	output, err := (&MailIt{}).Execute(context.Background(), helpers.HelperInput{
		ContactID: "contact_123",
		Config: map[string]interface{}{
			"subject_template": "Test",
			"body_template":    "Plain text body",
			"from_name":        "Sender",
			"from_email":       "sender@example.com",
			"content_type":     "text/plain",
		},
		Connector: mock,
	})

	if err != nil {
		t.Fatal(err)
	}
	if !output.Success {
		t.Error("should succeed")
	}

	payload := output.Actions[0].Value.(map[string]interface{})
	if payload["content_type"] != "text/plain" {
		t.Errorf("expected text/plain content type, got %v", payload["content_type"])
	}
}

func TestMailIt_Execute_DefaultHTMLContentType(t *testing.T) {
	mock := &mockConnectorForMailIt{
		contact: &connectors.NormalizedContact{
			ID:    "contact_123",
			Email: "test@example.com",
		},
	}

	output, err := (&MailIt{}).Execute(context.Background(), helpers.HelperInput{
		ContactID: "contact_123",
		Config: map[string]interface{}{
			"subject_template": "Test",
			"body_template":    "Body",
			"from_name":        "Sender",
			"from_email":       "sender@example.com",
		},
		Connector: mock,
	})

	if err != nil {
		t.Fatal(err)
	}

	payload := output.Actions[0].Value.(map[string]interface{})
	if payload["content_type"] != "text/html" {
		t.Errorf("expected default text/html content type, got %v", payload["content_type"])
	}
}

func TestMailIt_Execute_GetContactError(t *testing.T) {
	mock := &mockConnectorForMailIt{
		getError: fmt.Errorf("API error"),
	}

	output, err := (&MailIt{}).Execute(context.Background(), helpers.HelperInput{
		ContactID: "contact_123",
		Config: map[string]interface{}{
			"subject_template": "Test",
			"body_template":    "Body",
			"from_name":        "Sender",
			"from_email":       "sender@example.com",
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

func TestMailIt_Execute_FullNameMergeField(t *testing.T) {
	mock := &mockConnectorForMailIt{
		contact: &connectors.NormalizedContact{
			ID:        "contact_123",
			FirstName: "John",
			LastName:  "Doe",
			Email:     "john@example.com",
		},
	}

	output, err := (&MailIt{}).Execute(context.Background(), helpers.HelperInput{
		ContactID: "contact_123",
		Config: map[string]interface{}{
			"subject_template": "Hello {{full_name}}",
			"body_template":    "Dear {{full_name}}, welcome!",
			"from_name":        "Team",
			"from_email":       "team@example.com",
		},
		Connector: mock,
	})

	if err != nil {
		t.Fatal(err)
	}

	payload := output.Actions[0].Value.(map[string]interface{})
	if !strings.Contains(payload["subject"].(string), "John Doe") {
		t.Error("should contain full name")
	}
}

func TestMailIt_Execute_PayloadMetadata(t *testing.T) {
	mock := &mockConnectorForMailIt{
		contact: &connectors.NormalizedContact{
			ID:    "contact_123",
			Email: "test@example.com",
		},
	}

	output, err := (&MailIt{}).Execute(context.Background(), helpers.HelperInput{
		ContactID: "contact_123",
		AccountID: "account_456",
		HelperID:  "helper_789",
		Config: map[string]interface{}{
			"subject_template": "Test",
			"body_template":    "Body",
			"from_name":        "Sender",
			"from_email":       "sender@example.com",
		},
		Connector: mock,
	})

	if err != nil {
		t.Fatal(err)
	}

	payload := output.Actions[0].Value.(map[string]interface{})
	if payload["contact_id"] != "contact_123" {
		t.Error("payload should include contact_id")
	}
	if payload["account_id"] != "account_456" {
		t.Error("payload should include account_id")
	}
	if payload["helper_id"] != "helper_789" {
		t.Error("payload should include helper_id")
	}
}
