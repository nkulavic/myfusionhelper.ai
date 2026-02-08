package notification

import (
	"context"
	"fmt"
	"testing"

	"github.com/myfusionhelper/api/internal/connectors"
	"github.com/myfusionhelper/api/internal/helpers"
)

type mockConnectorForNotifyMe struct {
	contact       *connectors.NormalizedContact
	getContactErr error
	fieldValues   map[string]interface{}
	getFieldError map[string]error
}

func (m *mockConnectorForNotifyMe) GetContact(ctx context.Context, contactID string) (*connectors.NormalizedContact, error) {
	if m.getContactErr != nil {
		return nil, m.getContactErr
	}
	if m.contact != nil {
		return m.contact, nil
	}
	return &connectors.NormalizedContact{
		ID:        contactID,
		FirstName: "John",
		LastName:  "Doe",
		Email:     "john@example.com",
		Phone:     "555-1234",
		Company:   "ACME Inc",
	}, nil
}

func (m *mockConnectorForNotifyMe) GetContactFieldValue(ctx context.Context, contactID, fieldKey string) (interface{}, error) {
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

func (m *mockConnectorForNotifyMe) GetContacts(ctx context.Context, opts connectors.QueryOptions) (*connectors.ContactList, error) { return nil, fmt.Errorf("not implemented") }
func (m *mockConnectorForNotifyMe) CreateContact(ctx context.Context, contact connectors.CreateContactInput) (*connectors.NormalizedContact, error) { return nil, fmt.Errorf("not implemented") }
func (m *mockConnectorForNotifyMe) UpdateContact(ctx context.Context, contactID string, updates connectors.UpdateContactInput) (*connectors.NormalizedContact, error) { return nil, fmt.Errorf("not implemented") }
func (m *mockConnectorForNotifyMe) DeleteContact(ctx context.Context, contactID string) error { return fmt.Errorf("not implemented") }
func (m *mockConnectorForNotifyMe) GetTags(ctx context.Context) ([]connectors.Tag, error) { return nil, fmt.Errorf("not implemented") }
func (m *mockConnectorForNotifyMe) ApplyTag(ctx context.Context, contactID, tagID string) error { return fmt.Errorf("not implemented") }
func (m *mockConnectorForNotifyMe) RemoveTag(ctx context.Context, contactID, tagID string) error { return fmt.Errorf("not implemented") }
func (m *mockConnectorForNotifyMe) GetCustomFields(ctx context.Context) ([]connectors.CustomField, error) { return nil, fmt.Errorf("not implemented") }
func (m *mockConnectorForNotifyMe) SetContactFieldValue(ctx context.Context, contactID, fieldKey string, value interface{}) error { return fmt.Errorf("not implemented") }
func (m *mockConnectorForNotifyMe) TriggerAutomation(ctx context.Context, contactID, automationID string) error { return fmt.Errorf("not implemented") }
func (m *mockConnectorForNotifyMe) AchieveGoal(ctx context.Context, contactID, goalName, integration string) error { return fmt.Errorf("not implemented") }
func (m *mockConnectorForNotifyMe) TestConnection(ctx context.Context) error { return fmt.Errorf("not implemented") }
func (m *mockConnectorForNotifyMe) GetMetadata() connectors.ConnectorMetadata { return connectors.ConnectorMetadata{} }
func (m *mockConnectorForNotifyMe) GetCapabilities() []connectors.Capability { return nil }

func TestNotifyMe_GetMetadata(t *testing.T) {
	h := &NotifyMe{}
	if h.GetName() != "Notify Me" { t.Error("wrong name") }
	if h.GetType() != "notify_me" { t.Error("wrong type") }
	if h.GetCategory() != "notification" { t.Error("wrong category") }
	if !h.RequiresCRM() { t.Error("should require CRM") }
}

func TestNotifyMe_ValidateConfig_MissingChannel(t *testing.T) {
	err := (&NotifyMe{}).ValidateConfig(map[string]interface{}{
		"message": "test",
	})
	if err == nil { t.Error("should error on missing channel") }
}

func TestNotifyMe_ValidateConfig_MissingMessage(t *testing.T) {
	err := (&NotifyMe{}).ValidateConfig(map[string]interface{}{
		"channel": "email",
	})
	if err == nil { t.Error("should error on missing message") }
}

func TestNotifyMe_ValidateConfig_InvalidChannel(t *testing.T) {
	err := (&NotifyMe{}).ValidateConfig(map[string]interface{}{
		"channel": "sms",
		"message": "test",
	})
	if err == nil { t.Error("should error on invalid channel") }
}

func TestNotifyMe_ValidateConfig_Valid(t *testing.T) {
	channels := []string{"email", "slack", "webhook"}
	for _, ch := range channels {
		err := (&NotifyMe{}).ValidateConfig(map[string]interface{}{
			"channel": ch,
			"message": "test",
		})
		if err != nil { t.Errorf("should accept %s: %v", ch, err) }
	}
}

func TestNotifyMe_Execute_EmailNotification(t *testing.T) {
	mock := &mockConnectorForNotifyMe{}

	output, err := (&NotifyMe{}).Execute(context.Background(), helpers.HelperInput{
		ContactID: "123",
		Config: map[string]interface{}{
			"channel":   "email",
			"subject":   "Alert for {{first_name}} {{last_name}}",
			"message":   "Contact {{email}} from {{company}}",
			"recipient": "admin@example.com",
		},
		Connector: mock,
	})
	if err != nil { t.Fatal(err) }
	if !output.Success { t.Error("should succeed") }

	// Check interpolated subject
	if output.ModifiedData["subject"] != "Alert for John Doe" {
		t.Errorf("expected interpolated subject, got %v", output.ModifiedData["subject"])
	}

	// Check interpolated message
	if output.ModifiedData["message"] != "Contact john@example.com from ACME Inc" {
		t.Errorf("expected interpolated message, got %v", output.ModifiedData["message"])
	}

	// Check action
	if len(output.Actions) != 1 { t.Fatal("expected 1 action") }
	if output.Actions[0].Type != "notification_queued" {
		t.Errorf("expected notification_queued, got %s", output.Actions[0].Type)
	}
	if output.Actions[0].Target != "email" {
		t.Errorf("expected target email, got %s", output.Actions[0].Target)
	}
}

func TestNotifyMe_Execute_SlackNotification(t *testing.T) {
	mock := &mockConnectorForNotifyMe{}

	output, err := (&NotifyMe{}).Execute(context.Background(), helpers.HelperInput{
		ContactID: "123",
		Config: map[string]interface{}{
			"channel":   "slack",
			"message":   "New contact: {{full_name}}",
			"recipient": "#general",
		},
		Connector: mock,
	})
	if err != nil { t.Fatal(err) }
	if !output.Success { t.Error("should succeed") }

	notifData := output.ModifiedData
	if notifData["channel"] != "slack" {
		t.Errorf("expected slack channel, got %v", notifData["channel"])
	}
	if notifData["recipient"] != "#general" {
		t.Errorf("expected #general recipient, got %v", notifData["recipient"])
	}
	if notifData["message"] != "New contact: John Doe" {
		t.Errorf("expected full_name interpolated, got %v", notifData["message"])
	}
}

func TestNotifyMe_Execute_WebhookNotification(t *testing.T) {
	mock := &mockConnectorForNotifyMe{}

	output, err := (&NotifyMe{}).Execute(context.Background(), helpers.HelperInput{
		ContactID: "123",
		Config: map[string]interface{}{
			"channel":   "webhook",
			"message":   "Contact {{contact_id}} updated",
			"recipient": "https://api.example.com/webhook",
		},
		Connector: mock,
	})
	if err != nil { t.Fatal(err) }
	if !output.Success { t.Error("should succeed") }

	notifData := output.ModifiedData
	if notifData["message"] != "Contact 123 updated" {
		t.Errorf("expected contact_id interpolated, got %v", notifData["message"])
	}
}

func TestNotifyMe_Execute_IncludeAdditionalFields(t *testing.T) {
	mock := &mockConnectorForNotifyMe{
		fieldValues: map[string]interface{}{
			"custom_field_1": "Value1",
			"custom_field_2": 42,
		},
	}

	output, err := (&NotifyMe{}).Execute(context.Background(), helpers.HelperInput{
		ContactID: "123",
		Config: map[string]interface{}{
			"channel": "email",
			"message": "Field1: {{custom_field_1}}, Field2: {{custom_field_2}}",
			"include_fields": []interface{}{"custom_field_1", "custom_field_2"},
		},
		Connector: mock,
	})
	if err != nil { t.Fatal(err) }
	if !output.Success { t.Error("should succeed") }

	notifData := output.ModifiedData
	if notifData["message"] != "Field1: Value1, Field2: 42" {
		t.Errorf("expected custom fields interpolated, got %v", notifData["message"])
	}

	// Check contact data includes custom fields
	contactData := notifData["contact"].(map[string]string)
	if contactData["custom_field_1"] != "Value1" {
		t.Errorf("expected custom_field_1 in contact data, got %v", contactData)
	}
}

func TestNotifyMe_Execute_NoSubjectOrRecipient(t *testing.T) {
	mock := &mockConnectorForNotifyMe{}

	output, err := (&NotifyMe{}).Execute(context.Background(), helpers.HelperInput{
		ContactID: "123",
		Config: map[string]interface{}{
			"channel": "email",
			"message": "Simple notification",
		},
		Connector: mock,
	})
	if err != nil { t.Fatal(err) }
	if !output.Success { t.Error("should succeed") }

	notifData := output.ModifiedData
	if notifData["subject"] != "" {
		t.Errorf("expected empty subject, got %v", notifData["subject"])
	}
	if notifData["recipient"] != "" {
		t.Errorf("expected empty recipient, got %v", notifData["recipient"])
	}
}

func TestNotifyMe_Execute_AllStandardFields(t *testing.T) {
	mock := &mockConnectorForNotifyMe{}

	output, err := (&NotifyMe{}).Execute(context.Background(), helpers.HelperInput{
		ContactID: "123",
		Config: map[string]interface{}{
			"channel": "email",
			"message": "ID:{{contact_id}} Name:{{first_name}} {{last_name}} Email:{{email}} Phone:{{phone}} Company:{{company}} Full:{{full_name}}",
		},
		Connector: mock,
	})
	if err != nil { t.Fatal(err) }

	notifData := output.ModifiedData
	expected := "ID:123 Name:John Doe Email:john@example.com Phone:555-1234 Company:ACME Inc Full:John Doe"
	if notifData["message"] != expected {
		t.Errorf("expected all fields interpolated\nwant: %s\ngot:  %v", expected, notifData["message"])
	}
}

func TestNotifyMe_Execute_GetContactError(t *testing.T) {
	mock := &mockConnectorForNotifyMe{
		getContactErr: fmt.Errorf("API error"),
	}

	output, err := (&NotifyMe{}).Execute(context.Background(), helpers.HelperInput{
		ContactID: "123",
		Config: map[string]interface{}{
			"channel": "email",
			"message": "test",
		},
		Connector: mock,
	})
	if err == nil { t.Error("should error when GetContact fails") }
	if output.Success { t.Error("should not succeed on error") }
}

func TestNotifyMe_Execute_GetFieldError(t *testing.T) {
	mock := &mockConnectorForNotifyMe{
		getFieldError: map[string]error{
			"custom_field": fmt.Errorf("field error"),
		},
	}

	output, err := (&NotifyMe{}).Execute(context.Background(), helpers.HelperInput{
		ContactID: "123",
		Config: map[string]interface{}{
			"channel":        "email",
			"message":        "Field: {{custom_field}}",
			"include_fields": []interface{}{"custom_field"},
		},
		Connector: mock,
	})
	if err != nil { t.Fatal(err) }

	// Should still succeed, just skip failed field
	if !output.Success { t.Error("should succeed even if field fetch fails") }

	// Field should not be interpolated (placeholder remains)
	notifData := output.ModifiedData
	if notifData["message"] != "Field: {{custom_field}}" {
		t.Errorf("expected placeholder to remain, got %v", notifData["message"])
	}
}

func TestNotifyMe_Execute_ContactDataInAction(t *testing.T) {
	mock := &mockConnectorForNotifyMe{}

	output, err := (&NotifyMe{}).Execute(context.Background(), helpers.HelperInput{
		ContactID: "123",
		Config: map[string]interface{}{
			"channel": "email",
			"message": "test",
		},
		Connector: mock,
	})
	if err != nil { t.Fatal(err) }

	// Check action value contains full notification data
	actionValue := output.Actions[0].Value.(map[string]interface{})
	if actionValue["channel"] != "email" {
		t.Errorf("expected channel in action value, got %v", actionValue)
	}
	if actionValue["message"] != "test" {
		t.Errorf("expected message in action value, got %v", actionValue)
	}

	contactData := actionValue["contact"].(map[string]string)
	if contactData["first_name"] != "John" {
		t.Errorf("expected contact data in action, got %v", contactData)
	}
}

func TestNotifyMe_Execute_EmptyStringFields(t *testing.T) {
	mock := &mockConnectorForNotifyMe{
		contact: &connectors.NormalizedContact{
			ID:        "456",
			FirstName: "",
			LastName:  "",
			Email:     "",
		},
	}

	output, err := (&NotifyMe{}).Execute(context.Background(), helpers.HelperInput{
		ContactID: "456",
		Config: map[string]interface{}{
			"channel": "email",
			"message": "Name: {{first_name}} {{last_name}}",
		},
		Connector: mock,
	})
	if err != nil { t.Fatal(err) }

	// Empty fields should interpolate to empty strings
	notifData := output.ModifiedData
	if notifData["message"] != "Name:  " {
		t.Errorf("expected empty name interpolation, got %v", notifData["message"])
	}
}

func TestNotifyMe_Execute_ExtractStringSliceTypes(t *testing.T) {
	tests := []struct {
		name   string
		fields interface{}
		want   []string
	}{
		{
			name:   "string slice",
			fields: []string{"field1", "field2"},
			want:   []string{"field1", "field2"},
		},
		{
			name:   "interface slice",
			fields: []interface{}{"field1", "field2"},
			want:   []string{"field1", "field2"},
		},
		{
			name:   "mixed types",
			fields: []interface{}{"field1", 123},
			want:   []string{"field1", "123"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockConnectorForNotifyMe{
				fieldValues: map[string]interface{}{
					"field1": "value1",
					"field2": "value2",
				},
			}

			output, err := (&NotifyMe{}).Execute(context.Background(), helpers.HelperInput{
				ContactID: "123",
				Config: map[string]interface{}{
					"channel":        "email",
					"message":        "test",
					"include_fields": tt.fields,
				},
				Connector: mock,
			})
			if err != nil { t.Fatal(err) }
			if !output.Success { t.Error("should succeed") }
		})
	}
}
