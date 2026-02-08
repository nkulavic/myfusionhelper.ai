package contact

import (
	"context"
	"testing"

	"github.com/myfusionhelper/api/internal/connectors"
	"github.com/myfusionhelper/api/internal/helpers"
)

// TestContactUpdater_RegistryIntegration verifies the helper can be created from the registry
func TestContactUpdater_RegistryIntegration(t *testing.T) {
	// Verify helper is registered
	if !helpers.IsRegistered("contact_updater") {
		t.Fatal("contact_updater is not registered in the helper registry")
	}

	// Create helper from registry
	h, err := helpers.NewHelper("contact_updater")
	if err != nil {
		t.Fatalf("NewHelper(contact_updater) failed: %v", err)
	}

	// Verify it's the correct type
	updater, ok := h.(*ContactUpdater)
	if !ok {
		t.Fatalf("NewHelper returned type %T, want *ContactUpdater", h)
	}

	// Verify metadata
	if updater.GetType() != "contact_updater" {
		t.Errorf("GetType() = %v, want contact_updater", updater.GetType())
	}

	if updater.GetName() != "Contact Updater" {
		t.Errorf("GetName() = %v, want Contact Updater", updater.GetName())
	}

	if updater.GetCategory() != "contact" {
		t.Errorf("GetCategory() = %v, want contact", updater.GetCategory())
	}
}

// TestContactUpdater_ListHelperInfo verifies the helper appears in ListHelperInfo
func TestContactUpdater_ListHelperInfo(t *testing.T) {
	allHelpers := helpers.ListHelperInfo()

	// Find contact_updater in the list
	var found *helpers.HelperInfo
	for i := range allHelpers {
		if allHelpers[i].Type == "contact_updater" {
			found = &allHelpers[i]
			break
		}
	}

	if found == nil {
		t.Fatal("contact_updater not found in ListHelperInfo")
	}

	// Verify metadata
	if found.Name != "Contact Updater" {
		t.Errorf("HelperInfo.Name = %v, want Contact Updater", found.Name)
	}

	if found.Category != "contact" {
		t.Errorf("HelperInfo.Category = %v, want contact", found.Category)
	}

	if !found.RequiresCRM {
		t.Error("HelperInfo.RequiresCRM = false, want true")
	}

	if found.SupportedCRMs != nil {
		t.Errorf("HelperInfo.SupportedCRMs = %v, want nil (all CRMs)", found.SupportedCRMs)
	}

	// Verify config schema exists
	if found.ConfigSchema == nil {
		t.Error("HelperInfo.ConfigSchema is nil, want schema object")
	}

	// Verify schema has required properties
	props, ok := found.ConfigSchema["properties"].(map[string]interface{})
	if !ok {
		t.Fatal("ConfigSchema properties is not a map")
	}

	if _, exists := props["fields"]; !exists {
		t.Error("ConfigSchema missing 'fields' property")
	}

	if _, exists := props["secondary_contact_ids"]; !exists {
		t.Error("ConfigSchema missing 'secondary_contact_ids' property")
	}
}

// TestContactUpdater_EndToEndExecution tests a complete execution flow
func TestContactUpdater_EndToEndExecution(t *testing.T) {
	// Create helper from registry
	h, err := helpers.NewHelper("contact_updater")
	if err != nil {
		t.Fatalf("NewHelper failed: %v", err)
	}

	// Prepare test config
	config := map[string]interface{}{
		"fields": map[string]interface{}{
			"lead_status":    "qualified",
			"lead_score":     95,
			"last_contacted": "2024-02-08",
		},
		"secondary_contact_ids": []interface{}{"contact_sec_1", "contact_sec_2"},
	}

	// Validate config
	if err := h.ValidateConfig(config); err != nil {
		t.Fatalf("ValidateConfig failed: %v", err)
	}

	// Create mock connector
	mock := &mockConnector{
		updatedContact: &connectors.NormalizedContact{
			ID:        "test_contact_123",
			FirstName: "Integration",
			LastName:  "Test",
		},
	}

	// Prepare input
	input := helpers.HelperInput{
		ContactID: "test_contact_123",
		Config:    config,
		Connector: mock,
		UserID:    "user:test_user",
		AccountID: "account:test_account",
		HelperID:  "helper:contact_updater_test",
	}

	// Execute
	output, err := h.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	// Verify success
	if !output.Success {
		t.Errorf("Execute success = false, want true. Message: %s", output.Message)
	}

	// Verify UpdateContact was called
	if len(mock.updateContactCalls) != 1 {
		t.Errorf("UpdateContact called %d times, want 1", len(mock.updateContactCalls))
	}

	// Verify custom fields were passed correctly
	if len(mock.updateContactCalls[0].input.CustomFields) != 3 {
		t.Errorf("CustomFields count = %d, want 3", len(mock.updateContactCalls[0].input.CustomFields))
	}

	expectedFields := map[string]interface{}{
		"lead_status":    "qualified",
		"lead_score":     95,
		"last_contacted": "2024-02-08",
	}

	for fieldName, expectedValue := range expectedFields {
		actualValue := mock.updateContactCalls[0].input.CustomFields[fieldName]
		if actualValue != expectedValue {
			t.Errorf("CustomFields[%s] = %v, want %v", fieldName, actualValue, expectedValue)
		}
	}

	// Verify AchieveGoal was called for secondary contacts
	if len(mock.achieveGoalCalls) != 2 {
		t.Errorf("AchieveGoal called %d times, want 2", len(mock.achieveGoalCalls))
	}

	// Verify goal parameters
	for i, call := range mock.achieveGoalCalls {
		expectedContactID := config["secondary_contact_ids"].([]interface{})[i].(string)
		if call.contactID != expectedContactID {
			t.Errorf("AchieveGoal[%d] contactID = %v, want %v", i, call.contactID, expectedContactID)
		}
		if call.goalName != "contact_updated" {
			t.Errorf("AchieveGoal[%d] goalName = %v, want contact_updated", i, call.goalName)
		}
		if call.integration != "contact_updater" {
			t.Errorf("AchieveGoal[%d] integration = %v, want contact_updater", i, call.integration)
		}
	}

	// Verify output structure
	if len(output.Actions) != 5 { // 3 field updates + 2 goal achievements
		t.Errorf("Actions count = %d, want 5", len(output.Actions))
	}

	// Count action types
	fieldUpdateActions := 0
	goalAchievedActions := 0
	for _, action := range output.Actions {
		switch action.Type {
		case "field_updated":
			fieldUpdateActions++
		case "goal_achieved":
			goalAchievedActions++
		}
	}

	if fieldUpdateActions != 3 {
		t.Errorf("field_updated actions = %d, want 3", fieldUpdateActions)
	}

	if goalAchievedActions != 2 {
		t.Errorf("goal_achieved actions = %d, want 2", goalAchievedActions)
	}

	// Verify logs
	if len(output.Logs) != 5 { // 3 field updates + 2 goal achievements
		t.Errorf("Logs count = %d, want 5", len(output.Logs))
	}

	// Verify modified data
	if len(output.ModifiedData) != 3 {
		t.Errorf("ModifiedData count = %d, want 3", len(output.ModifiedData))
	}

	// Verify message format
	expectedMessage := "Updated 3 fields on contact test_contact_123"
	if output.Message != expectedMessage {
		t.Errorf("Message = %v, want %v", output.Message, expectedMessage)
	}
}
