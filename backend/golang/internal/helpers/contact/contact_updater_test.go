package contact

import (
	"context"
	"errors"
	"testing"

	"github.com/myfusionhelper/api/internal/connectors"
	"github.com/myfusionhelper/api/internal/helpers"
)

// mockConnector implements the CRMConnector interface for testing
type mockConnector struct {
	updateContactCalls     []updateContactCall
	achieveGoalCalls       []achieveGoalCall
	updateContactError     error
	achieveGoalError       error
	updatedContact         *connectors.NormalizedContact
}

type updateContactCall struct {
	contactID string
	input     connectors.UpdateContactInput
}

type achieveGoalCall struct {
	contactID   string
	goalName    string
	integration string
}

func (m *mockConnector) UpdateContact(ctx context.Context, contactID string, updates connectors.UpdateContactInput) (*connectors.NormalizedContact, error) {
	m.updateContactCalls = append(m.updateContactCalls, updateContactCall{
		contactID: contactID,
		input:     updates,
	})
	if m.updateContactError != nil {
		return nil, m.updateContactError
	}
	if m.updatedContact != nil {
		return m.updatedContact, nil
	}
	return &connectors.NormalizedContact{
		ID:        contactID,
		FirstName: "Updated",
		LastName:  "Contact",
	}, nil
}

func (m *mockConnector) AchieveGoal(ctx context.Context, contactID string, goalName string, integration string) error {
	m.achieveGoalCalls = append(m.achieveGoalCalls, achieveGoalCall{
		contactID:   contactID,
		goalName:    goalName,
		integration: integration,
	})
	return m.achieveGoalError
}

// Minimal implementation of other required CRMConnector methods
func (m *mockConnector) GetContacts(ctx context.Context, opts connectors.QueryOptions) (*connectors.ContactList, error) {
	return nil, nil
}
func (m *mockConnector) GetContact(ctx context.Context, contactID string) (*connectors.NormalizedContact, error) {
	return nil, nil
}
func (m *mockConnector) CreateContact(ctx context.Context, contact connectors.CreateContactInput) (*connectors.NormalizedContact, error) {
	return nil, nil
}
func (m *mockConnector) DeleteContact(ctx context.Context, contactID string) error {
	return nil
}
func (m *mockConnector) GetTags(ctx context.Context) ([]connectors.Tag, error) {
	return nil, nil
}
func (m *mockConnector) ApplyTag(ctx context.Context, contactID string, tagID string) error {
	return nil
}
func (m *mockConnector) RemoveTag(ctx context.Context, contactID string, tagID string) error {
	return nil
}
func (m *mockConnector) GetCustomFields(ctx context.Context) ([]connectors.CustomField, error) {
	return nil, nil
}
func (m *mockConnector) GetContactFieldValue(ctx context.Context, contactID string, fieldKey string) (interface{}, error) {
	return nil, nil
}
func (m *mockConnector) SetContactFieldValue(ctx context.Context, contactID string, fieldKey string, value interface{}) error {
	return nil
}
func (m *mockConnector) TriggerAutomation(ctx context.Context, contactID string, automationID string) error {
	return nil
}
func (m *mockConnector) TestConnection(ctx context.Context) error {
	return nil
}
func (m *mockConnector) GetMetadata() connectors.ConnectorMetadata {
	return connectors.ConnectorMetadata{}
}
func (m *mockConnector) GetCapabilities() []connectors.Capability {
	return nil
}

func TestContactUpdater_GetMetadata(t *testing.T) {
	h := &ContactUpdater{}

	if got := h.GetName(); got != "Contact Updater" {
		t.Errorf("GetName() = %v, want %v", got, "Contact Updater")
	}

	if got := h.GetType(); got != "contact_updater" {
		t.Errorf("GetType() = %v, want %v", got, "contact_updater")
	}

	if got := h.GetCategory(); got != "contact" {
		t.Errorf("GetCategory() = %v, want %v", got, "contact")
	}

	if !h.RequiresCRM() {
		t.Error("RequiresCRM() = false, want true")
	}

	if crms := h.SupportedCRMs(); crms != nil {
		t.Errorf("SupportedCRMs() = %v, want nil (all CRMs)", crms)
	}
}

func TestContactUpdater_ValidateConfig(t *testing.T) {
	h := &ContactUpdater{}

	tests := []struct {
		name    string
		config  map[string]interface{}
		wantErr bool
	}{
		{
			name: "valid config with fields",
			config: map[string]interface{}{
				"fields": map[string]interface{}{
					"status": "active",
					"score":  100,
				},
			},
			wantErr: false,
		},
		{
			name: "valid config with fields and secondary contacts",
			config: map[string]interface{}{
				"fields": map[string]interface{}{
					"status": "active",
				},
				"secondary_contact_ids": []interface{}{"contact_1", "contact_2"},
			},
			wantErr: false,
		},
		{
			name:    "missing fields",
			config:  map[string]interface{}{},
			wantErr: true,
		},
		{
			name: "invalid fields type (not a map)",
			config: map[string]interface{}{
				"fields": "invalid",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := h.ValidateConfig(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestContactUpdater_Execute_Success(t *testing.T) {
	h := &ContactUpdater{}
	mock := &mockConnector{
		updatedContact: &connectors.NormalizedContact{
			ID:        "contact_123",
			FirstName: "John",
			LastName:  "Doe",
		},
	}

	input := helpers.HelperInput{
		ContactID: "contact_123",
		Config: map[string]interface{}{
			"fields": map[string]interface{}{
				"status":      "active",
				"lead_score":  85,
				"last_update": "2024-01-15",
			},
		},
		Connector: mock,
	}

	output, err := h.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("Execute() error = %v, want nil", err)
	}

	// Check output success
	if !output.Success {
		t.Error("Execute() success = false, want true")
	}

	// Check message
	expectedMsg := "Updated 3 fields on contact contact_123"
	if output.Message != expectedMsg {
		t.Errorf("Execute() message = %v, want %v", output.Message, expectedMsg)
	}

	// Check UpdateContact was called correctly
	if len(mock.updateContactCalls) != 1 {
		t.Fatalf("UpdateContact called %d times, want 1", len(mock.updateContactCalls))
	}

	call := mock.updateContactCalls[0]
	if call.contactID != "contact_123" {
		t.Errorf("UpdateContact contactID = %v, want %v", call.contactID, "contact_123")
	}

	// Verify custom fields were passed
	if len(call.input.CustomFields) != 3 {
		t.Errorf("UpdateContact CustomFields length = %d, want 3", len(call.input.CustomFields))
	}

	if call.input.CustomFields["status"] != "active" {
		t.Errorf("CustomFields[status] = %v, want active", call.input.CustomFields["status"])
	}

	// Check actions were logged
	if len(output.Actions) != 3 {
		t.Errorf("Actions count = %d, want 3", len(output.Actions))
	}

	// Verify action types
	for _, action := range output.Actions {
		if action.Type != "field_updated" {
			t.Errorf("Action type = %v, want field_updated", action.Type)
		}
	}

	// Check logs
	if len(output.Logs) != 3 {
		t.Errorf("Logs count = %d, want 3", len(output.Logs))
	}

	// Check modified data
	if len(output.ModifiedData) != 3 {
		t.Errorf("ModifiedData count = %d, want 3", len(output.ModifiedData))
	}
}

func TestContactUpdater_Execute_WithSecondaryContacts(t *testing.T) {
	h := &ContactUpdater{}
	mock := &mockConnector{}

	input := helpers.HelperInput{
		ContactID: "contact_123",
		Config: map[string]interface{}{
			"fields": map[string]interface{}{
				"status": "active",
			},
			"secondary_contact_ids": []interface{}{"contact_456", "contact_789"},
		},
		Connector: mock,
	}

	output, err := h.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("Execute() error = %v, want nil", err)
	}

	if !output.Success {
		t.Error("Execute() success = false, want true")
	}

	// Check UpdateContact was called for primary contact
	if len(mock.updateContactCalls) != 1 {
		t.Fatalf("UpdateContact called %d times, want 1", len(mock.updateContactCalls))
	}

	// Check AchieveGoal was called for secondary contacts
	if len(mock.achieveGoalCalls) != 2 {
		t.Fatalf("AchieveGoal called %d times, want 2", len(mock.achieveGoalCalls))
	}

	// Verify first secondary contact goal
	if mock.achieveGoalCalls[0].contactID != "contact_456" {
		t.Errorf("AchieveGoal[0] contactID = %v, want contact_456", mock.achieveGoalCalls[0].contactID)
	}
	if mock.achieveGoalCalls[0].goalName != "contact_updated" {
		t.Errorf("AchieveGoal[0] goalName = %v, want contact_updated", mock.achieveGoalCalls[0].goalName)
	}
	if mock.achieveGoalCalls[0].integration != "contact_updater" {
		t.Errorf("AchieveGoal[0] integration = %v, want contact_updater", mock.achieveGoalCalls[0].integration)
	}

	// Verify second secondary contact goal
	if mock.achieveGoalCalls[1].contactID != "contact_789" {
		t.Errorf("AchieveGoal[1] contactID = %v, want contact_789", mock.achieveGoalCalls[1].contactID)
	}

	// Check that goal actions were logged
	goalActions := 0
	for _, action := range output.Actions {
		if action.Type == "goal_achieved" {
			goalActions++
		}
	}
	if goalActions != 2 {
		t.Errorf("goal_achieved actions = %d, want 2", goalActions)
	}

	// Check logs include goal achievements
	goalLogs := 0
	for _, log := range output.Logs {
		if len(log) > 15 && log[:15] == "Triggered goal " {
			goalLogs++
		}
	}
	if goalLogs != 2 {
		t.Errorf("goal trigger logs = %d, want 2", goalLogs)
	}
}

func TestContactUpdater_Execute_UpdateError(t *testing.T) {
	h := &ContactUpdater{}
	mock := &mockConnector{
		updateContactError: errors.New("update contact failed"),
	}

	input := helpers.HelperInput{
		ContactID: "contact_123",
		Config: map[string]interface{}{
			"fields": map[string]interface{}{
				"status": "active",
			},
		},
		Connector: mock,
	}

	output, err := h.Execute(context.Background(), input)
	if err == nil {
		t.Fatal("Execute() error = nil, want error")
	}

	// Output should still be returned but with error message
	if output.Message == "" {
		t.Error("Execute() output.Message is empty, want error message")
	}

	if output.Success {
		t.Error("Execute() success = true, want false on error")
	}
}

func TestContactUpdater_Execute_SecondaryContactGoalError(t *testing.T) {
	h := &ContactUpdater{}
	mock := &mockConnector{
		achieveGoalError: errors.New("achieve goal failed"),
	}

	input := helpers.HelperInput{
		ContactID: "contact_123",
		Config: map[string]interface{}{
			"fields": map[string]interface{}{
				"status": "active",
			},
			"secondary_contact_ids": []interface{}{"contact_456"},
		},
		Connector: mock,
	}

	output, err := h.Execute(context.Background(), input)
	// Should succeed even if goal achievement fails
	if err != nil {
		t.Fatalf("Execute() error = %v, want nil (goal errors are non-fatal)", err)
	}

	if !output.Success {
		t.Error("Execute() success = false, want true (goal errors are non-fatal)")
	}

	// Should have error log for failed goal
	hasErrorLog := false
	for _, log := range output.Logs {
		if len(log) > 18 && log[:18] == "Failed to trigger " {
			hasErrorLog = true
			break
		}
	}
	if !hasErrorLog {
		t.Error("Execute() logs missing error message for failed goal achievement")
	}
}

func TestContactUpdater_GetConfigSchema(t *testing.T) {
	h := &ContactUpdater{}
	schema := h.GetConfigSchema()

	// Verify schema has correct structure
	if schema["type"] != "object" {
		t.Errorf("Schema type = %v, want object", schema["type"])
	}

	props, ok := schema["properties"].(map[string]interface{})
	if !ok {
		t.Fatal("Schema properties is not a map")
	}

	// Check fields property
	if _, exists := props["fields"]; !exists {
		t.Error("Schema missing 'fields' property")
	}

	// Check secondary_contact_ids property
	if _, exists := props["secondary_contact_ids"]; !exists {
		t.Error("Schema missing 'secondary_contact_ids' property")
	}

	// Check required fields
	required, ok := schema["required"].([]string)
	if !ok {
		t.Fatal("Schema required is not a string slice")
	}

	if len(required) != 1 || required[0] != "fields" {
		t.Errorf("Schema required = %v, want [fields]", required)
	}
}
