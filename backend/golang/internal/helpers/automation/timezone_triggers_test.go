package automation

import (
	"context"
	"fmt"
	"testing"

	"github.com/myfusionhelper/api/internal/connectors"
	"github.com/myfusionhelper/api/internal/helpers"
)

type mockConnectorForTimezoneTriggers struct {
	contact          *connectors.NormalizedContact
	getContactError  error
	achieveGoalCalls []struct{ contactID, goalName, integration string }
	achieveGoalError error
}

func (m *mockConnectorForTimezoneTriggers) GetContact(ctx context.Context, contactID string) (*connectors.NormalizedContact, error) {
	if m.getContactError != nil {
		return nil, m.getContactError
	}
	if m.contact != nil {
		return m.contact, nil
	}
	return &connectors.NormalizedContact{
		ID:           contactID,
		CustomFields: map[string]interface{}{},
	}, nil
}

func (m *mockConnectorForTimezoneTriggers) AchieveGoal(ctx context.Context, contactID, goalName, integration string) error {
	if m.achieveGoalError != nil {
		return m.achieveGoalError
	}
	if m.achieveGoalCalls == nil {
		m.achieveGoalCalls = make([]struct{ contactID, goalName, integration string }, 0)
	}
	m.achieveGoalCalls = append(m.achieveGoalCalls, struct{ contactID, goalName, integration string }{contactID, goalName, integration})
	return nil
}

// Stub implementations
func (m *mockConnectorForTimezoneTriggers) GetContacts(ctx context.Context, opts connectors.QueryOptions) (*connectors.ContactList, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForTimezoneTriggers) CreateContact(ctx context.Context, input connectors.CreateContactInput) (*connectors.NormalizedContact, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForTimezoneTriggers) UpdateContact(ctx context.Context, contactID string, updates connectors.UpdateContactInput) (*connectors.NormalizedContact, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForTimezoneTriggers) DeleteContact(ctx context.Context, contactID string) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorForTimezoneTriggers) GetTags(ctx context.Context) ([]connectors.Tag, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForTimezoneTriggers) ApplyTag(ctx context.Context, contactID string, tagID string) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorForTimezoneTriggers) RemoveTag(ctx context.Context, contactID string, tagID string) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorForTimezoneTriggers) GetCustomFields(ctx context.Context) ([]connectors.CustomField, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForTimezoneTriggers) GetContactFieldValue(ctx context.Context, contactID, fieldKey string) (interface{}, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForTimezoneTriggers) SetContactFieldValue(ctx context.Context, contactID, fieldKey string, value interface{}) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorForTimezoneTriggers) TriggerAutomation(ctx context.Context, contactID string, automationID string) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorForTimezoneTriggers) TestConnection(ctx context.Context) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorForTimezoneTriggers) GetMetadata() connectors.ConnectorMetadata {
	return connectors.ConnectorMetadata{PlatformSlug: "test"}
}
func (m *mockConnectorForTimezoneTriggers) GetCapabilities() []connectors.Capability {
	return []connectors.Capability{connectors.CapContacts, connectors.CapGoals}
}

func TestTimezoneTriggers_GetMetadata(t *testing.T) {
	h := &TimezoneTriggers{}
	if h.GetName() != "Timezone Triggers" {
		t.Error("Wrong name")
	}
	if h.GetType() != "timezone_triggers" {
		t.Error("Wrong type")
	}
	if h.GetCategory() != "automation" {
		t.Error("Wrong category")
	}
	if !h.RequiresCRM() {
		t.Error("Should require CRM")
	}
}

func TestTimezoneTriggers_GetConfigSchema(t *testing.T) {
	h := &TimezoneTriggers{}
	schema := h.GetConfigSchema()
	props, ok := schema["properties"].(map[string]interface{})
	if !ok {
		t.Fatal("Schema should have properties")
	}
	if _, ok := props["day"]; !ok {
		t.Error("Schema should have day property")
	}
	if _, ok := props["time"]; !ok {
		t.Error("Schema should have time property")
	}
}

func TestTimezoneTriggers_ValidateConfig_MissingDay(t *testing.T) {
	h := &TimezoneTriggers{}
	if err := h.ValidateConfig(map[string]interface{}{"time": "9:00 AM"}); err == nil {
		t.Error("Expected error for missing day")
	}
}

func TestTimezoneTriggers_ValidateConfig_MissingTime(t *testing.T) {
	h := &TimezoneTriggers{}
	if err := h.ValidateConfig(map[string]interface{}{"day": "Monday"}); err == nil {
		t.Error("Expected error for missing time")
	}
}

func TestTimezoneTriggers_ValidateConfig_EmptyDay(t *testing.T) {
	h := &TimezoneTriggers{}
	if err := h.ValidateConfig(map[string]interface{}{"day": "", "time": "9:00 AM"}); err == nil {
		t.Error("Expected error for empty day")
	}
}

func TestTimezoneTriggers_ValidateConfig_EmptyTime(t *testing.T) {
	h := &TimezoneTriggers{}
	if err := h.ValidateConfig(map[string]interface{}{"day": "Monday", "time": ""}); err == nil {
		t.Error("Expected error for empty time")
	}
}

func TestTimezoneTriggers_ValidateConfig_Valid(t *testing.T) {
	h := &TimezoneTriggers{}
	if err := h.ValidateConfig(map[string]interface{}{
		"day":  "Monday",
		"time": "9:00 AM",
	}); err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestTimezoneTriggers_Execute_WithAddress(t *testing.T) {
	h := &TimezoneTriggers{}
	mockConn := &mockConnectorForTimezoneTriggers{
		contact: &connectors.NormalizedContact{
			ID: "c123",
			CustomFields: map[string]interface{}{
				"StreetAddress1": "123 Main St",
				"City":           "San Francisco",
				"State":          "CA",
				"PostalCode":     "94105",
				"Country":        "USA",
			},
		},
	}

	output, err := h.Execute(context.Background(), helpers.HelperInput{
		ContactID: "c123",
		Config: map[string]interface{}{
			"day":  "Monday",
			"time": "9:00 AM",
		},
		Connector: mockConn,
	})

	if err != nil {
		t.Fatalf("Error: %v", err)
	}
	if !output.Success {
		t.Error("Should succeed")
	}
	if len(output.Actions) != 1 {
		t.Fatal("Expected 1 action")
	}
	if output.Actions[0].Type != "timezone_trigger_scheduled" {
		t.Error("Wrong action type")
	}
}

func TestTimezoneTriggers_Execute_NoAddress_FailedGoal(t *testing.T) {
	h := &TimezoneTriggers{}
	mockConn := &mockConnectorForTimezoneTriggers{
		contact: &connectors.NormalizedContact{
			ID:           "c123",
			CustomFields: map[string]interface{}{},
		},
	}

	output, err := h.Execute(context.Background(), helpers.HelperInput{
		ContactID: "c123",
		Config: map[string]interface{}{
			"day":         "Monday",
			"time":        "9:00 AM",
			"failed_goal": "No Address Found",
		},
		Connector: mockConn,
	})

	if err != nil {
		t.Fatalf("Error: %v", err)
	}
	if !output.Success {
		t.Error("Should succeed")
	}
	if len(mockConn.achieveGoalCalls) != 1 {
		t.Error("Should achieve failed goal")
	}
	if mockConn.achieveGoalCalls[0].goalName != "No Address Found" {
		t.Error("Wrong goal name")
	}
}

func TestTimezoneTriggers_Execute_NoAddress_NoFailedGoal(t *testing.T) {
	h := &TimezoneTriggers{}
	mockConn := &mockConnectorForTimezoneTriggers{
		contact: &connectors.NormalizedContact{
			ID:           "c123",
			CustomFields: map[string]interface{}{},
		},
	}

	output, err := h.Execute(context.Background(), helpers.HelperInput{
		ContactID: "c123",
		Config: map[string]interface{}{
			"day":  "Monday",
			"time": "9:00 AM",
		},
		Connector: mockConn,
	})

	if err != nil {
		t.Fatalf("Error: %v", err)
	}
	if !output.Success {
		t.Error("Should succeed")
	}
	if len(mockConn.achieveGoalCalls) != 0 {
		t.Error("Should not achieve any goal")
	}
	if output.Message != "No address found for timezone resolution" {
		t.Errorf("Wrong message: %s", output.Message)
	}
}

func TestTimezoneTriggers_Execute_GetContactError(t *testing.T) {
	h := &TimezoneTriggers{}
	mockConn := &mockConnectorForTimezoneTriggers{
		getContactError: fmt.Errorf("contact not found"),
	}

	output, err := h.Execute(context.Background(), helpers.HelperInput{
		ContactID: "c123",
		Config: map[string]interface{}{
			"day":  "Monday",
			"time": "9:00 AM",
		},
		Connector: mockConn,
	})

	if err == nil {
		t.Error("Expected error")
	}
	if output.Success {
		t.Error("Should not succeed")
	}
}

func TestTimezoneTriggers_Execute_AlternateAddress(t *testing.T) {
	h := &TimezoneTriggers{}
	mockConn := &mockConnectorForTimezoneTriggers{
		contact: &connectors.NormalizedContact{
			ID: "c123",
			CustomFields: map[string]interface{}{
				"Address2Street1": "456 Oak Ave",
				"City2":           "Seattle",
				"State2":          "WA",
				"PostalCode2":     "98101",
			},
		},
	}

	output, err := h.Execute(context.Background(), helpers.HelperInput{
		ContactID: "c123",
		Config: map[string]interface{}{
			"day":  "Tuesday",
			"time": "14:00",
		},
		Connector: mockConn,
	})

	if err != nil {
		t.Fatalf("Error: %v", err)
	}
	if !output.Success {
		t.Error("Should succeed")
	}
	if len(output.Actions) != 1 {
		t.Fatal("Expected 1 action")
	}
}

func TestTimezoneTriggers_Execute_WithOptionalFields(t *testing.T) {
	h := &TimezoneTriggers{}
	mockConn := &mockConnectorForTimezoneTriggers{
		contact: &connectors.NormalizedContact{
			ID: "c123",
			CustomFields: map[string]interface{}{
				"StreetAddress1": "789 Pine Rd",
				"City":           "Austin",
				"State":          "TX",
			},
		},
	}

	output, err := h.Execute(context.Background(), helpers.HelperInput{
		ContactID: "c123",
		Config: map[string]interface{}{
			"day":                    "Wednesday",
			"time":                   "10:30 AM",
			"save_time_zone":         "timezone_field",
			"save_lat_lng":           "latlng_field",
			"save_time_zone_offset":  "offset_field",
			"trigger_goal":           "Timezone Resolved",
		},
		Connector: mockConn,
	})

	if err != nil {
		t.Fatalf("Error: %v", err)
	}
	if !output.Success {
		t.Error("Should succeed")
	}

	// Verify modified data contains timezone request
	if output.ModifiedData == nil {
		t.Fatal("ModifiedData should not be nil")
	}
	if output.ModifiedData["save_time_zone_field"] != "timezone_field" {
		t.Error("Should include save_time_zone_field")
	}
	if output.ModifiedData["save_lat_lng_field"] != "latlng_field" {
		t.Error("Should include save_lat_lng_field")
	}
	if output.ModifiedData["save_time_zone_offset_field"] != "offset_field" {
		t.Error("Should include save_time_zone_offset_field")
	}
	if output.ModifiedData["trigger_goal"] != "Timezone Resolved" {
		t.Error("Should include trigger_goal")
	}
}

func TestTimezoneTriggers_Execute_ActionsRecorded(t *testing.T) {
	h := &TimezoneTriggers{}
	mockConn := &mockConnectorForTimezoneTriggers{
		contact: &connectors.NormalizedContact{
			ID: "c123",
			CustomFields: map[string]interface{}{
				"StreetAddress1": "123 Test St",
				"City":           "Portland",
			},
		},
	}

	output, err := h.Execute(context.Background(), helpers.HelperInput{
		ContactID: "c123",
		Config: map[string]interface{}{
			"day":  "Friday",
			"time": "15:00",
		},
		Connector: mockConn,
	})

	if err != nil {
		t.Fatalf("Error: %v", err)
	}
	if len(output.Actions) != 1 {
		t.Fatalf("Expected 1 action, got %d", len(output.Actions))
	}
	if output.Actions[0].Type != "timezone_trigger_scheduled" {
		t.Error("Wrong action type")
	}
	if output.Actions[0].Target != "c123" {
		t.Error("Wrong action target")
	}
}

func TestTimezoneTriggers_Execute_LogsRecorded(t *testing.T) {
	h := &TimezoneTriggers{}
	mockConn := &mockConnectorForTimezoneTriggers{
		contact: &connectors.NormalizedContact{
			ID: "c123",
			CustomFields: map[string]interface{}{
				"StreetAddress1": "123 Test St",
				"City":           "Denver",
			},
		},
	}

	output, err := h.Execute(context.Background(), helpers.HelperInput{
		ContactID: "c123",
		Config: map[string]interface{}{
			"day":  "Saturday",
			"time": "8:00 AM",
		},
		Connector: mockConn,
	})

	if err != nil {
		t.Fatalf("Error: %v", err)
	}
	if len(output.Logs) == 0 {
		t.Error("Expected logs")
	}
}

func TestTimezoneTriggers_Execute_MessageFormat(t *testing.T) {
	h := &TimezoneTriggers{}
	mockConn := &mockConnectorForTimezoneTriggers{
		contact: &connectors.NormalizedContact{
			ID: "c123",
			CustomFields: map[string]interface{}{
				"StreetAddress1": "100 Broadway",
				"City":           "New York",
			},
		},
	}

	output, err := h.Execute(context.Background(), helpers.HelperInput{
		ContactID: "c123",
		Config: map[string]interface{}{
			"day":  "Thursday",
			"time": "11:45 AM",
		},
		Connector: mockConn,
	})

	if err != nil {
		t.Fatalf("Error: %v", err)
	}
	expected := "Timezone trigger scheduled for Thursday 11:45 AM based on contact address"
	if output.Message != expected {
		t.Errorf("Expected message '%s', got '%s'", expected, output.Message)
	}
}
