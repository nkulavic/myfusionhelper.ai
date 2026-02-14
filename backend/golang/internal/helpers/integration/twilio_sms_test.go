package integration

import (
	"context"
	"fmt"
	"testing"

	"github.com/myfusionhelper/api/internal/connectors"
	"github.com/myfusionhelper/api/internal/helpers"
)

// mockConnectorForTwilioSMS mocks the CRMConnector interface for testing twilio_sms
type mockConnectorForTwilioSMS struct {
	contact         *connectors.NormalizedContact
	getContactError error
}

func (m *mockConnectorForTwilioSMS) GetContact(ctx context.Context, contactID string) (*connectors.NormalizedContact, error) {
	if m.getContactError != nil {
		return nil, m.getContactError
	}
	return m.contact, nil
}

// Stub implementations
func (m *mockConnectorForTwilioSMS) GetContacts(ctx context.Context, opts connectors.QueryOptions) (*connectors.ContactList, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *mockConnectorForTwilioSMS) CreateContact(ctx context.Context, input connectors.CreateContactInput) (*connectors.NormalizedContact, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *mockConnectorForTwilioSMS) UpdateContact(ctx context.Context, contactID string, updates connectors.UpdateContactInput) (*connectors.NormalizedContact, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *mockConnectorForTwilioSMS) DeleteContact(ctx context.Context, contactID string) error {
	return fmt.Errorf("not implemented")
}

func (m *mockConnectorForTwilioSMS) GetTags(ctx context.Context) ([]connectors.Tag, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *mockConnectorForTwilioSMS) ApplyTag(ctx context.Context, contactID, tagID string) error {
	return fmt.Errorf("not implemented")
}

func (m *mockConnectorForTwilioSMS) RemoveTag(ctx context.Context, contactID, tagID string) error {
	return fmt.Errorf("not implemented")
}

func (m *mockConnectorForTwilioSMS) GetCustomFields(ctx context.Context) ([]connectors.CustomField, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *mockConnectorForTwilioSMS) GetContactFieldValue(ctx context.Context, contactID, fieldKey string) (interface{}, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *mockConnectorForTwilioSMS) SetContactFieldValue(ctx context.Context, contactID, fieldKey string, value interface{}) error {
	return fmt.Errorf("not implemented")
}

func (m *mockConnectorForTwilioSMS) TriggerAutomation(ctx context.Context, contactID, automationID string) error {
	return fmt.Errorf("not implemented")
}

func (m *mockConnectorForTwilioSMS) AchieveGoal(ctx context.Context, contactID, goalName, integration string) error {
	return fmt.Errorf("not implemented")
}

func (m *mockConnectorForTwilioSMS) TestConnection(ctx context.Context) error {
	return fmt.Errorf("not implemented")
}

func (m *mockConnectorForTwilioSMS) GetMetadata() connectors.ConnectorMetadata {
	return connectors.ConnectorMetadata{PlatformSlug: "mock", PlatformName: "Mock"}
}

func (m *mockConnectorForTwilioSMS) GetCapabilities() []connectors.Capability {
	return []connectors.Capability{}
}

func TestTwilioSMS_GetMetadata(t *testing.T) {
	helper := &TwilioSMS{}

	if helper.GetName() != "Twilio SMS" {
		t.Errorf("Expected name 'Twilio SMS', got '%s'", helper.GetName())
	}
	if helper.GetType() != "twilio_sms" {
		t.Errorf("Expected type 'twilio_sms', got '%s'", helper.GetType())
	}
	if helper.GetCategory() != "integration" {
		t.Errorf("Expected category 'integration', got '%s'", helper.GetCategory())
	}
	if !helper.RequiresCRM() {
		t.Error("Expected RequiresCRM to be true")
	}
}

func TestTwilioSMS_ValidateConfig_Success(t *testing.T) {
	helper := &TwilioSMS{}

	config := map[string]interface{}{
		"account_sid":      "ACtest123",
		"auth_token":       "token456",
		"from_number":      "+15551234567",
		"message_template": "Hello {{FirstName}}!",
	}

	err := helper.ValidateConfig(config)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
}

func TestTwilioSMS_ValidateConfig_MissingAccountSID(t *testing.T) {
	helper := &TwilioSMS{}

	config := map[string]interface{}{
		"auth_token":       "token456",
		"from_number":      "+15551234567",
		"message_template": "Hello!",
	}

	err := helper.ValidateConfig(config)
	if err == nil {
		t.Error("Expected error for missing account_sid")
	}
	if err != nil && err.Error() != "account_sid is required" {
		t.Errorf("Expected 'account_sid is required', got: %v", err)
	}
}

func TestTwilioSMS_ValidateConfig_MissingAuthToken(t *testing.T) {
	helper := &TwilioSMS{}

	config := map[string]interface{}{
		"account_sid":      "ACtest123",
		"from_number":      "+15551234567",
		"message_template": "Hello!",
	}

	err := helper.ValidateConfig(config)
	if err == nil {
		t.Error("Expected error for missing auth_token")
	}
}

func TestTwilioSMS_ValidateConfig_MissingFromNumber(t *testing.T) {
	helper := &TwilioSMS{}

	config := map[string]interface{}{
		"account_sid":      "ACtest123",
		"auth_token":       "token456",
		"message_template": "Hello!",
	}

	err := helper.ValidateConfig(config)
	if err == nil {
		t.Error("Expected error for missing from_number")
	}
}

func TestTwilioSMS_ValidateConfig_MissingMessageTemplate(t *testing.T) {
	helper := &TwilioSMS{}

	config := map[string]interface{}{
		"account_sid": "ACtest123",
		"auth_token":  "token456",
		"from_number": "+15551234567",
	}

	err := helper.ValidateConfig(config)
	if err == nil {
		t.Error("Expected error for missing message_template")
	}
}

func TestTwilioSMS_Execute_Success(t *testing.T) {
	helper := &TwilioSMS{}
	mock := &mockConnectorForTwilioSMS{
		contact: &connectors.NormalizedContact{
			ID:        "123",
			FirstName: "John",
			LastName:  "Doe",
			Email:     "john@example.com",
			Phone:     "+15559876543",
		},
	}

	input := helpers.HelperInput{
		ContactID: "123",
		Config: map[string]interface{}{
			"account_sid":      "ACtest123",
			"auth_token":       "token456",
			"from_number":      "+15551234567",
			"message_template": "Hello {{FirstName}}!",
		},
		Connector: mock,
	}

	output, err := helper.Execute(context.Background(), input)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if !output.Success {
		t.Error("Expected success to be true")
	}

	// Verify webhook action
	if len(output.Actions) != 1 {
		t.Errorf("Expected 1 action, got %d", len(output.Actions))
	}
	if output.Actions[0].Type != "webhook_queued" {
		t.Errorf("Expected action type 'webhook_queued', got: %s", output.Actions[0].Type)
	}

	// Verify ModifiedData
	if output.ModifiedData["to_number"] != "+15559876543" {
		t.Errorf("Expected to_number to be '+15559876543', got: %v", output.ModifiedData["to_number"])
	}
	if output.ModifiedData["message"] != "Hello John!" {
		t.Errorf("Expected message to be 'Hello John!', got: %v", output.ModifiedData["message"])
	}
}

func TestTwilioSMS_Execute_MergeFields(t *testing.T) {
	helper := &TwilioSMS{}
	mock := &mockConnectorForTwilioSMS{
		contact: &connectors.NormalizedContact{
			ID:        "123",
			FirstName: "Jane",
			LastName:  "Smith",
			Email:     "jane@example.com",
			Phone:     "+15559876543",
			Company:   "Acme Corp",
		},
	}

	input := helpers.HelperInput{
		ContactID: "123",
		Config: map[string]interface{}{
			"account_sid":      "ACtest123",
			"auth_token":       "token456",
			"from_number":      "+15551234567",
			"message_template": "Hi {{FirstName}} {{LastName}} from {{Company}}!",
		},
		Connector: mock,
	}

	output, err := helper.Execute(context.Background(), input)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if output.ModifiedData["message"] != "Hi Jane Smith from Acme Corp!" {
		t.Errorf("Expected message with merged fields, got: %v", output.ModifiedData["message"])
	}
}

func TestTwilioSMS_Execute_CustomField(t *testing.T) {
	helper := &TwilioSMS{}
	mock := &mockConnectorForTwilioSMS{
		contact: &connectors.NormalizedContact{
			ID:        "123",
			FirstName: "Bob",
			Phone:     "+15559876543",
			CustomFields: map[string]interface{}{
				"custom_field1": "CustomValue",
			},
		},
	}

	input := helpers.HelperInput{
		ContactID: "123",
		Config: map[string]interface{}{
			"account_sid":      "ACtest123",
			"auth_token":       "token456",
			"from_number":      "+15551234567",
			"message_template": "Custom: {{custom_field1}}",
		},
		Connector: mock,
	}

	output, err := helper.Execute(context.Background(), input)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if output.ModifiedData["message"] != "Custom: CustomValue" {
		t.Errorf("Expected custom field merged, got: %v", output.ModifiedData["message"])
	}
}

func TestTwilioSMS_Execute_CustomToField(t *testing.T) {
	helper := &TwilioSMS{}
	mock := &mockConnectorForTwilioSMS{
		contact: &connectors.NormalizedContact{
			ID:        "123",
			FirstName: "Alice",
			CustomFields: map[string]interface{}{
				"mobile_phone": "+15551112222",
			},
		},
	}

	input := helpers.HelperInput{
		ContactID: "123",
		Config: map[string]interface{}{
			"account_sid":      "ACtest123",
			"auth_token":       "token456",
			"from_number":      "+15551234567",
			"message_template": "Test message",
			"to_field":         "mobile_phone",
		},
		Connector: mock,
	}

	output, err := helper.Execute(context.Background(), input)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if output.ModifiedData["to_number"] != "+15551112222" {
		t.Errorf("Expected to_number from custom field, got: %v", output.ModifiedData["to_number"])
	}
}

func TestTwilioSMS_Execute_EmptyPhoneField(t *testing.T) {
	helper := &TwilioSMS{}
	mock := &mockConnectorForTwilioSMS{
		contact: &connectors.NormalizedContact{
			ID:        "123",
			FirstName: "NoPhone",
			Phone:     "",
		},
	}

	input := helpers.HelperInput{
		ContactID: "123",
		Config: map[string]interface{}{
			"account_sid":      "ACtest123",
			"auth_token":       "token456",
			"from_number":      "+15551234567",
			"message_template": "Test",
		},
		Connector: mock,
	}

	output, err := helper.Execute(context.Background(), input)
	if err == nil {
		t.Error("Expected error for empty phone field")
	}
	if output.Success {
		t.Error("Expected success to be false")
	}
}

func TestTwilioSMS_Execute_GetContactError(t *testing.T) {
	helper := &TwilioSMS{}
	mock := &mockConnectorForTwilioSMS{
		getContactError: fmt.Errorf("contact not found"),
	}

	input := helpers.HelperInput{
		ContactID: "123",
		Config: map[string]interface{}{
			"account_sid":      "ACtest123",
			"auth_token":       "token456",
			"from_number":      "+15551234567",
			"message_template": "Test",
		},
		Connector: mock,
	}

	output, err := helper.Execute(context.Background(), input)
	if err == nil {
		t.Error("Expected error for GetContact failure")
	}
	if output.Success {
		t.Error("Expected success to be false")
	}
}

func TestTwilioSMS_Execute_FullNameMergeField(t *testing.T) {
	helper := &TwilioSMS{}
	mock := &mockConnectorForTwilioSMS{
		contact: &connectors.NormalizedContact{
			ID:        "123",
			FirstName: "John",
			LastName:  "Doe",
			Phone:     "+15559876543",
		},
	}

	input := helpers.HelperInput{
		ContactID: "123",
		Config: map[string]interface{}{
			"account_sid":      "ACtest123",
			"auth_token":       "token456",
			"from_number":      "+15551234567",
			"message_template": "Hello {{full_name}}!",
		},
		Connector: mock,
	}

	output, err := helper.Execute(context.Background(), input)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if output.ModifiedData["message"] != "Hello John Doe!" {
		t.Errorf("Expected full_name merge field, got: %v", output.ModifiedData["message"])
	}
}

func TestTwilioSMS_Execute_WebhookActionStructure(t *testing.T) {
	helper := &TwilioSMS{}
	mock := &mockConnectorForTwilioSMS{
		contact: &connectors.NormalizedContact{
			ID:        "123",
			FirstName: "Test",
			Phone:     "+15559876543",
		},
	}

	input := helpers.HelperInput{
		ContactID: "123",
		Config: map[string]interface{}{
			"account_sid":      "ACtest123",
			"auth_token":       "token456",
			"from_number":      "+15551234567",
			"message_template": "Test",
		},
		Connector: mock,
	}

	output, err := helper.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Verify webhook action value structure
	actionValue, ok := output.Actions[0].Value.(map[string]interface{})
	if !ok {
		t.Fatal("Expected action value to be map[string]interface{}")
	}

	if actionValue["method"] != "POST" {
		t.Errorf("Expected method 'POST', got: %v", actionValue["method"])
	}
	if actionValue["auth_type"] != "basic" {
		t.Errorf("Expected auth_type 'basic', got: %v", actionValue["auth_type"])
	}
	if actionValue["auth_user"] != "ACtest123" {
		t.Errorf("Expected auth_user 'ACtest123', got: %v", actionValue["auth_user"])
	}
	if actionValue["content_type"] != "application/x-www-form-urlencoded" {
		t.Errorf("Expected content_type 'application/x-www-form-urlencoded', got: %v", actionValue["content_type"])
	}

	// Verify payload
	payload, ok := actionValue["payload"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected payload to be map[string]interface{}")
	}
	if payload["From"] != "+15551234567" {
		t.Errorf("Expected From '+15551234567', got: %v", payload["From"])
	}
	if payload["To"] != "+15559876543" {
		t.Errorf("Expected To '+15559876543', got: %v", payload["To"])
	}
	if payload["Body"] != "Test" {
		t.Errorf("Expected Body 'Test', got: %v", payload["Body"])
	}
}

func TestTwilioSMS_Execute_APIURLFormat(t *testing.T) {
	helper := &TwilioSMS{}
	mock := &mockConnectorForTwilioSMS{
		contact: &connectors.NormalizedContact{
			ID:    "123",
			Phone: "+15559876543",
		},
	}

	input := helpers.HelperInput{
		ContactID: "123",
		Config: map[string]interface{}{
			"account_sid":      "ACtest123",
			"auth_token":       "token456",
			"from_number":      "+15551234567",
			"message_template": "Test",
		},
		Connector: mock,
	}

	output, err := helper.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	expectedURL := "https://api.twilio.com/2010-04-01/Accounts/ACtest123/Messages.json"
	if output.ModifiedData["api_url"] != expectedURL {
		t.Errorf("Expected api_url '%s', got: %v", expectedURL, output.ModifiedData["api_url"])
	}
	if output.Actions[0].Target != expectedURL {
		t.Errorf("Expected action target '%s', got: %s", expectedURL, output.Actions[0].Target)
	}
}
