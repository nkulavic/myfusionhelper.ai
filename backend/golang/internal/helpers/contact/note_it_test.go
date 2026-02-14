package contact

import (
	"context"
	"fmt"
	"testing"

	"github.com/myfusionhelper/api/internal/connectors"
	"github.com/myfusionhelper/api/internal/helpers"
)

// mockConnectorForNoteIt implements connectors.CRMConnector for testing
type mockConnectorForNoteIt struct {
	contact         *connectors.NormalizedContact
	getContactError error
}

func (m *mockConnectorForNoteIt) GetContact(ctx context.Context, contactID string) (*connectors.NormalizedContact, error) {
	if m.getContactError != nil {
		return nil, m.getContactError
	}
	return m.contact, nil
}

// Stub implementations for unused interface methods
func (m *mockConnectorForNoteIt) GetContacts(ctx context.Context, opts connectors.QueryOptions) (*connectors.ContactList, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForNoteIt) CreateContact(ctx context.Context, contact connectors.CreateContactInput) (*connectors.NormalizedContact, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForNoteIt) UpdateContact(ctx context.Context, contactID string, updates connectors.UpdateContactInput) (*connectors.NormalizedContact, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForNoteIt) DeleteContact(ctx context.Context, contactID string) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorForNoteIt) GetTags(ctx context.Context) ([]connectors.Tag, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForNoteIt) ApplyTag(ctx context.Context, contactID, tagID string) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorForNoteIt) RemoveTag(ctx context.Context, contactID, tagID string) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorForNoteIt) GetCustomFields(ctx context.Context) ([]connectors.CustomField, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForNoteIt) GetContactFieldValue(ctx context.Context, contactID, fieldKey string) (interface{}, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForNoteIt) SetContactFieldValue(ctx context.Context, contactID, fieldKey string, value interface{}) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorForNoteIt) TriggerAutomation(ctx context.Context, contactID, automationID string) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorForNoteIt) AchieveGoal(ctx context.Context, contactID, goalName, integration string) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorForNoteIt) TestConnection(ctx context.Context) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorForNoteIt) GetMetadata() connectors.ConnectorMetadata {
	return connectors.ConnectorMetadata{}
}
func (m *mockConnectorForNoteIt) GetCapabilities() []connectors.Capability {
	return nil
}

func TestNoteIt_GetMetadata(t *testing.T) {
	h := &NoteIt{}
	if h.GetName() != "Note It" {
		t.Errorf("expected name 'Note It', got '%s'", h.GetName())
	}
	if h.GetType() != "note_it" {
		t.Errorf("expected type 'note_it', got '%s'", h.GetType())
	}
	if h.GetCategory() != "contact" {
		t.Errorf("expected category 'contact', got '%s'", h.GetCategory())
	}
	if !h.RequiresCRM() {
		t.Error("expected RequiresCRM to be true")
	}
	if h.SupportedCRMs() != nil {
		t.Errorf("expected nil SupportedCRMs, got %v", h.SupportedCRMs())
	}
}

func TestNoteIt_GetConfigSchema(t *testing.T) {
	h := &NoteIt{}
	schema := h.GetConfigSchema()
	if schema["type"] != "object" {
		t.Errorf("expected type 'object', got '%v'", schema["type"])
	}
	props := schema["properties"].(map[string]interface{})
	if _, ok := props["subject"]; !ok {
		t.Error("schema missing subject property")
	}
	if _, ok := props["body"]; !ok {
		t.Error("schema missing body property")
	}
	if _, ok := props["note_type"]; !ok {
		t.Error("schema missing note_type property")
	}

	required := schema["required"].([]string)
	if len(required) != 2 {
		t.Errorf("expected 2 required fields, got %d", len(required))
	}
}

func TestNoteIt_ValidateConfig_MissingSubject(t *testing.T) {
	h := &NoteIt{}
	config := map[string]interface{}{
		"body": "Test body",
	}
	err := h.ValidateConfig(config)
	if err == nil {
		t.Error("expected error for missing subject")
	}
}

func TestNoteIt_ValidateConfig_EmptySubject(t *testing.T) {
	h := &NoteIt{}
	config := map[string]interface{}{
		"subject": "",
		"body":    "Test body",
	}
	err := h.ValidateConfig(config)
	if err == nil {
		t.Error("expected error for empty subject")
	}
}

func TestNoteIt_ValidateConfig_MissingBody(t *testing.T) {
	h := &NoteIt{}
	config := map[string]interface{}{
		"subject": "Test subject",
	}
	err := h.ValidateConfig(config)
	if err == nil {
		t.Error("expected error for missing body")
	}
}

func TestNoteIt_ValidateConfig_EmptyBody(t *testing.T) {
	h := &NoteIt{}
	config := map[string]interface{}{
		"subject": "Test subject",
		"body":    "",
	}
	err := h.ValidateConfig(config)
	if err == nil {
		t.Error("expected error for empty body")
	}
}

func TestNoteIt_ValidateConfig_Valid(t *testing.T) {
	h := &NoteIt{}
	config := map[string]interface{}{
		"subject": "Test subject",
		"body":    "Test body",
	}
	err := h.ValidateConfig(config)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestNoteIt_Execute_BasicNote(t *testing.T) {
	h := &NoteIt{}
	mock := &mockConnectorForNoteIt{
		contact: &connectors.NormalizedContact{
			ID:        "12345",
			FirstName: "John",
			LastName:  "Doe",
			Email:     "john@example.com",
		},
	}

	input := helpers.HelperInput{
		ContactID: "12345",
		Config: map[string]interface{}{
			"subject": "Follow up call",
			"body":    "Need to discuss project details",
		},
		Connector: mock,
	}

	output, err := h.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !output.Success {
		t.Error("expected success=true")
	}
	if output.Message != "Note prepared: Follow up call" {
		t.Errorf("unexpected message: %s", output.Message)
	}

	noteData := output.ModifiedData
	if noteData["subject"] != "Follow up call" {
		t.Errorf("unexpected subject: %v", noteData["subject"])
	}
	if noteData["body"] != "Need to discuss project details" {
		t.Errorf("unexpected body: %v", noteData["body"])
	}
	if noteData["note_type"] != "general" {
		t.Errorf("expected default note_type=general, got %v", noteData["note_type"])
	}
}

func TestNoteIt_Execute_WithTemplateInterpolation(t *testing.T) {
	h := &NoteIt{}
	mock := &mockConnectorForNoteIt{
		contact: &connectors.NormalizedContact{
			ID:        "12345",
			FirstName: "Jane",
			LastName:  "Smith",
			Email:     "jane@example.com",
			Phone:     "555-1234",
			Company:   "Acme Corp",
		},
	}

	input := helpers.HelperInput{
		ContactID: "12345",
		Config: map[string]interface{}{
			"subject": "Call {{first_name}} at {{company}}",
			"body":    "Hi {{full_name}}, please call {{phone}}. Email: {{email}}",
		},
		Connector: mock,
	}

	output, err := h.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !output.Success {
		t.Error("expected success=true")
	}

	noteData := output.ModifiedData
	expectedSubject := "Call Jane at Acme Corp"
	if noteData["subject"] != expectedSubject {
		t.Errorf("expected subject '%s', got '%v'", expectedSubject, noteData["subject"])
	}

	expectedBody := "Hi Jane Smith, please call 555-1234. Email: jane@example.com"
	if noteData["body"] != expectedBody {
		t.Errorf("expected body '%s', got '%v'", expectedBody, noteData["body"])
	}
}

func TestNoteIt_Execute_WithCustomFields(t *testing.T) {
	h := &NoteIt{}
	mock := &mockConnectorForNoteIt{
		contact: &connectors.NormalizedContact{
			ID:        "12345",
			FirstName: "Bob",
			LastName:  "Jones",
			CustomFields: map[string]interface{}{
				"account_status": "Premium",
				"last_purchase":  "2024-01-15",
			},
		},
	}

	input := helpers.HelperInput{
		ContactID: "12345",
		Config: map[string]interface{}{
			"subject": "Account update for {{full_name}}",
			"body":    "Status: {{account_status}}, Last purchase: {{last_purchase}}",
		},
		Connector: mock,
	}

	output, err := h.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !output.Success {
		t.Error("expected success=true")
	}

	noteData := output.ModifiedData
	expectedBody := "Status: Premium, Last purchase: 2024-01-15"
	if noteData["body"] != expectedBody {
		t.Errorf("expected body '%s', got '%v'", expectedBody, noteData["body"])
	}
}

func TestNoteIt_Execute_WithNoteType(t *testing.T) {
	h := &NoteIt{}
	mock := &mockConnectorForNoteIt{
		contact: &connectors.NormalizedContact{
			ID:        "12345",
			FirstName: "Alice",
			LastName:  "Brown",
		},
	}

	input := helpers.HelperInput{
		ContactID: "12345",
		Config: map[string]interface{}{
			"subject":   "Important reminder",
			"body":      "Followup needed",
			"note_type": "urgent",
		},
		Connector: mock,
	}

	output, err := h.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !output.Success {
		t.Error("expected success=true")
	}

	noteData := output.ModifiedData
	if noteData["note_type"] != "urgent" {
		t.Errorf("expected note_type=urgent, got %v", noteData["note_type"])
	}
}

func TestNoteIt_Execute_GetContactError(t *testing.T) {
	h := &NoteIt{}
	mock := &mockConnectorForNoteIt{
		getContactError: fmt.Errorf("CRM API error"),
	}

	input := helpers.HelperInput{
		ContactID: "12345",
		Config: map[string]interface{}{
			"subject": "Test",
			"body":    "Test body",
		},
		Connector: mock,
	}

	output, err := h.Execute(context.Background(), input)
	if err == nil {
		t.Error("expected error from GetContact")
	}
	if output.Success {
		t.Error("expected success=false on error")
	}
	if output.Message == "" {
		t.Error("expected error message to be set")
	}
}

func TestNoteIt_Execute_MultipleTemplates(t *testing.T) {
	h := &NoteIt{}
	mock := &mockConnectorForNoteIt{
		contact: &connectors.NormalizedContact{
			ID:        "12345",
			FirstName: "Test",
			LastName:  "User",
		},
	}

	input := helpers.HelperInput{
		ContactID: "12345",
		Config: map[string]interface{}{
			"subject": "Hello {{first_name}} {{last_name}}",
			"body":    "Dear {{first_name}}, your ID is {{contact_id}}. Regards, {{first_name}}",
		},
		Connector: mock,
	}

	output, err := h.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	noteData := output.ModifiedData
	expectedSubject := "Hello Test User"
	if noteData["subject"] != expectedSubject {
		t.Errorf("expected subject '%s', got '%v'", expectedSubject, noteData["subject"])
	}

	expectedBody := "Dear Test, your ID is 12345. Regards, Test"
	if noteData["body"] != expectedBody {
		t.Errorf("expected body '%s', got '%v'", expectedBody, noteData["body"])
	}
}

func TestNoteIt_Execute_ActionLogging(t *testing.T) {
	h := &NoteIt{}
	mock := &mockConnectorForNoteIt{
		contact: &connectors.NormalizedContact{
			ID:        "12345",
			FirstName: "Test",
			LastName:  "User",
		},
	}

	input := helpers.HelperInput{
		ContactID: "12345",
		Config: map[string]interface{}{
			"subject": "Test note",
			"body":    "Test body",
		},
		Connector: mock,
	}

	output, err := h.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(output.Actions) != 1 {
		t.Errorf("expected 1 action, got %d", len(output.Actions))
	}
	if output.Actions[0].Type != "notification_queued" {
		t.Errorf("expected action type 'notification_queued', got '%s'", output.Actions[0].Type)
	}
	if output.Actions[0].Target != "12345" {
		t.Errorf("expected action target '12345', got '%s'", output.Actions[0].Target)
	}

	if len(output.Logs) == 0 {
		t.Error("expected logs to be populated")
	}
}

func Test_interpolateNoteTemplate(t *testing.T) {
	tests := []struct {
		name     string
		template string
		data     map[string]string
		want     string
	}{
		{
			name:     "single replacement",
			template: "Hello {{name}}",
			data:     map[string]string{"name": "World"},
			want:     "Hello World",
		},
		{
			name:     "multiple replacements",
			template: "{{first}} {{last}}",
			data:     map[string]string{"first": "John", "last": "Doe"},
			want:     "John Doe",
		},
		{
			name:     "no placeholders",
			template: "Plain text",
			data:     map[string]string{"name": "Test"},
			want:     "Plain text",
		},
		{
			name:     "unused placeholders",
			template: "Hello {{name}}, {{age}}",
			data:     map[string]string{"name": "Bob"},
			want:     "Hello Bob, {{age}}",
		},
		{
			name:     "repeated placeholder",
			template: "{{x}} and {{x}} again",
			data:     map[string]string{"x": "value"},
			want:     "value and value again",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := interpolateNoteTemplate(tt.template, tt.data)
			if got != tt.want {
				t.Errorf("interpolateNoteTemplate() = %v, want %v", got, tt.want)
			}
		})
	}
}
