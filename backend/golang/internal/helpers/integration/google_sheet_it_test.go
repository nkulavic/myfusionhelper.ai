package integration

import (
	"context"
	"fmt"
	"testing"

	"github.com/myfusionhelper/api/internal/connectors"
	"github.com/myfusionhelper/api/internal/helpers"
)

// mockConnectorForGoogleSheetIt implements the CRMConnector interface for testing google_sheet_it
type mockConnectorForGoogleSheetIt struct {
	contact       *connectors.NormalizedContact
	getContactErr error
}

func (m *mockConnectorForGoogleSheetIt) GetContact(ctx context.Context, contactID string) (*connectors.NormalizedContact, error) {
	if m.getContactErr != nil {
		return nil, m.getContactErr
	}
	return m.contact, nil
}

// Stub implementations
func (m *mockConnectorForGoogleSheetIt) GetContacts(ctx context.Context, opts connectors.QueryOptions) (*connectors.ContactList, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *mockConnectorForGoogleSheetIt) CreateContact(ctx context.Context, input connectors.CreateContactInput) (*connectors.NormalizedContact, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *mockConnectorForGoogleSheetIt) UpdateContact(ctx context.Context, contactID string, updates connectors.UpdateContactInput) (*connectors.NormalizedContact, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *mockConnectorForGoogleSheetIt) DeleteContact(ctx context.Context, contactID string) error {
	return fmt.Errorf("not implemented")
}

func (m *mockConnectorForGoogleSheetIt) GetTags(ctx context.Context) ([]connectors.Tag, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *mockConnectorForGoogleSheetIt) ApplyTag(ctx context.Context, contactID string, tagID string) error {
	return fmt.Errorf("not implemented")
}

func (m *mockConnectorForGoogleSheetIt) RemoveTag(ctx context.Context, contactID string, tagID string) error {
	return fmt.Errorf("not implemented")
}

func (m *mockConnectorForGoogleSheetIt) GetCustomFields(ctx context.Context) ([]connectors.CustomField, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *mockConnectorForGoogleSheetIt) GetContactFieldValue(ctx context.Context, contactID, fieldKey string) (interface{}, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *mockConnectorForGoogleSheetIt) SetContactFieldValue(ctx context.Context, contactID, fieldKey string, value interface{}) error {
	return fmt.Errorf("not implemented")
}

func (m *mockConnectorForGoogleSheetIt) TriggerAutomation(ctx context.Context, contactID string, automationID string) error {
	return fmt.Errorf("not implemented")
}

func (m *mockConnectorForGoogleSheetIt) AchieveGoal(ctx context.Context, contactID string, goalName string, integration string) error {
	return fmt.Errorf("not implemented")
}

func (m *mockConnectorForGoogleSheetIt) TestConnection(ctx context.Context) error {
	return fmt.Errorf("not implemented")
}

func (m *mockConnectorForGoogleSheetIt) GetMetadata() connectors.ConnectorMetadata {
	return connectors.ConnectorMetadata{
		PlatformSlug: "test",
		PlatformName: "Test Platform",
	}
}

func (m *mockConnectorForGoogleSheetIt) GetCapabilities() []connectors.Capability {
	return []connectors.Capability{connectors.CapContacts}
}

// Tests

func TestGoogleSheetIt_GetMetadata(t *testing.T) {
	helper := &GoogleSheetIt{}

	if helper.GetName() != "Google Sheet It" {
		t.Errorf("Expected name 'Google Sheet It', got '%s'", helper.GetName())
	}
	if helper.GetType() != "google_sheet_it" {
		t.Errorf("Expected type 'google_sheet_it', got '%s'", helper.GetType())
	}
	if helper.GetCategory() != "integration" {
		t.Errorf("Expected category 'integration', got '%s'", helper.GetCategory())
	}
	if !helper.RequiresCRM() {
		t.Error("Expected RequiresCRM to be true")
	}
}

func TestGoogleSheetIt_GetConfigSchema(t *testing.T) {
	helper := &GoogleSheetIt{}
	schema := helper.GetConfigSchema()

	if schema == nil {
		t.Fatal("Schema should not be nil")
	}

	props, ok := schema["properties"].(map[string]interface{})
	if !ok {
		t.Fatal("Schema should have properties")
	}

	if _, ok := props["spreadsheet_id"]; !ok {
		t.Error("Schema should have spreadsheet_id property")
	}
	if _, ok := props["sheet_id"]; !ok {
		t.Error("Schema should have sheet_id property")
	}
	if _, ok := props["google_account_id"]; !ok {
		t.Error("Schema should have google_account_id property")
	}
	if _, ok := props["mode"]; !ok {
		t.Error("Schema should have mode property")
	}

	required, ok := schema["required"].([]string)
	if !ok {
		t.Fatal("Schema should have required array")
	}

	if len(required) < 3 {
		t.Error("Schema should require at least 3 fields")
	}
}

func TestGoogleSheetIt_ValidateConfig_MissingSpreadsheetID(t *testing.T) {
	helper := &GoogleSheetIt{}

	config := map[string]interface{}{
		"sheet_id":          "sheet-123",
		"google_account_id": "account-456",
	}
	err := helper.ValidateConfig(config)
	if err == nil {
		t.Error("Expected error for missing spreadsheet_id")
	}
}

func TestGoogleSheetIt_ValidateConfig_EmptySpreadsheetID(t *testing.T) {
	helper := &GoogleSheetIt{}

	config := map[string]interface{}{
		"spreadsheet_id":    "",
		"sheet_id":          "sheet-123",
		"google_account_id": "account-456",
	}
	err := helper.ValidateConfig(config)
	if err == nil {
		t.Error("Expected error for empty spreadsheet_id")
	}
}

func TestGoogleSheetIt_ValidateConfig_MissingSheetID(t *testing.T) {
	helper := &GoogleSheetIt{}

	config := map[string]interface{}{
		"spreadsheet_id":    "spreadsheet-123",
		"google_account_id": "account-456",
	}
	err := helper.ValidateConfig(config)
	if err == nil {
		t.Error("Expected error for missing sheet_id")
	}
}

func TestGoogleSheetIt_ValidateConfig_MissingGoogleAccountID(t *testing.T) {
	helper := &GoogleSheetIt{}

	config := map[string]interface{}{
		"spreadsheet_id": "spreadsheet-123",
		"sheet_id":       "sheet-123",
	}
	err := helper.ValidateConfig(config)
	if err == nil {
		t.Error("Expected error for missing google_account_id")
	}
}

func TestGoogleSheetIt_ValidateConfig_Valid(t *testing.T) {
	helper := &GoogleSheetIt{}

	config := map[string]interface{}{
		"spreadsheet_id":    "spreadsheet-123",
		"sheet_id":          "sheet-456",
		"google_account_id": "account-789",
	}
	err := helper.ValidateConfig(config)
	if err != nil {
		t.Errorf("Expected no error for valid config, got: %v", err)
	}
}

func TestGoogleSheetIt_Execute_Success(t *testing.T) {
	helper := &GoogleSheetIt{}
	mockConn := &mockConnectorForGoogleSheetIt{
		contact: &connectors.NormalizedContact{
			ID:        "contact-123",
			FirstName: "John",
			LastName:  "Doe",
			Email:     "john.doe@example.com",
		},
	}
	ctx := context.Background()

	input := helpers.HelperInput{
		ContactID: "contact-123",
		Config: map[string]interface{}{
			"spreadsheet_id":    "spreadsheet-123",
			"sheet_id":          "sheet-456",
			"google_account_id": "account-789",
		},
		Connector: mockConn,
	}

	output, err := helper.Execute(ctx, input)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if !output.Success {
		t.Error("Expected success to be true")
	}

	if output.Message == "" {
		t.Error("Expected non-empty message")
	}
}

func TestGoogleSheetIt_Execute_ModeReplace(t *testing.T) {
	helper := &GoogleSheetIt{}
	mockConn := &mockConnectorForGoogleSheetIt{
		contact: &connectors.NormalizedContact{
			ID:    "contact-123",
			Email: "test@example.com",
		},
	}
	ctx := context.Background()

	input := helpers.HelperInput{
		ContactID: "contact-123",
		Config: map[string]interface{}{
			"spreadsheet_id":    "spreadsheet-123",
			"sheet_id":          "sheet-456",
			"google_account_id": "account-789",
			"mode":              "replace",
		},
		Connector: mockConn,
	}

	output, err := helper.Execute(ctx, input)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if output.ModifiedData["mode"] != "replace" {
		t.Errorf("Expected mode to be 'replace', got: %v", output.ModifiedData["mode"])
	}
}

func TestGoogleSheetIt_Execute_ModeAppend(t *testing.T) {
	helper := &GoogleSheetIt{}
	mockConn := &mockConnectorForGoogleSheetIt{
		contact: &connectors.NormalizedContact{
			ID:    "contact-123",
			Email: "test@example.com",
		},
	}
	ctx := context.Background()

	input := helpers.HelperInput{
		ContactID: "contact-123",
		Config: map[string]interface{}{
			"spreadsheet_id":    "spreadsheet-123",
			"sheet_id":          "sheet-456",
			"google_account_id": "account-789",
			"mode":              "append",
		},
		Connector: mockConn,
	}

	output, err := helper.Execute(ctx, input)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if output.ModifiedData["mode"] != "append" {
		t.Errorf("Expected mode to be 'append', got: %v", output.ModifiedData["mode"])
	}
}

func TestGoogleSheetIt_Execute_TranslateTrue(t *testing.T) {
	helper := &GoogleSheetIt{}
	mockConn := &mockConnectorForGoogleSheetIt{
		contact: &connectors.NormalizedContact{
			ID:    "contact-123",
			Email: "test@example.com",
		},
	}
	ctx := context.Background()

	input := helpers.HelperInput{
		ContactID: "contact-123",
		Config: map[string]interface{}{
			"spreadsheet_id":    "spreadsheet-123",
			"sheet_id":          "sheet-456",
			"google_account_id": "account-789",
			"translate":         "true",
		},
		Connector: mockConn,
	}

	output, err := helper.Execute(ctx, input)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if output.ModifiedData["translate"] != true {
		t.Errorf("Expected translate to be true, got: %v", output.ModifiedData["translate"])
	}
}

func TestGoogleSheetIt_Execute_TranslateFalse(t *testing.T) {
	helper := &GoogleSheetIt{}
	mockConn := &mockConnectorForGoogleSheetIt{
		contact: &connectors.NormalizedContact{
			ID:    "contact-123",
			Email: "test@example.com",
		},
	}
	ctx := context.Background()

	input := helpers.HelperInput{
		ContactID: "contact-123",
		Config: map[string]interface{}{
			"spreadsheet_id":    "spreadsheet-123",
			"sheet_id":          "sheet-456",
			"google_account_id": "account-789",
			"translate":         "false",
		},
		Connector: mockConn,
	}

	output, err := helper.Execute(ctx, input)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if output.ModifiedData["translate"] != false {
		t.Errorf("Expected translate to be false, got: %v", output.ModifiedData["translate"])
	}
}

func TestGoogleSheetIt_Execute_WithSearchData(t *testing.T) {
	helper := &GoogleSheetIt{}
	mockConn := &mockConnectorForGoogleSheetIt{
		contact: &connectors.NormalizedContact{
			ID:    "contact-123",
			Email: "test@example.com",
		},
	}
	ctx := context.Background()

	input := helpers.HelperInput{
		ContactID: "contact-123",
		Config: map[string]interface{}{
			"spreadsheet_id":    "spreadsheet-123",
			"sheet_id":          "sheet-456",
			"google_account_id": "account-789",
			"search_data":       "search-999,user-888",
		},
		Connector: mockConn,
	}

	output, err := helper.Execute(ctx, input)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if output.ModifiedData["search_id"] != "search-999" {
		t.Errorf("Expected search_id to be 'search-999', got: %v", output.ModifiedData["search_id"])
	}

	if output.ModifiedData["search_user_id"] != "user-888" {
		t.Errorf("Expected search_user_id to be 'user-888', got: %v", output.ModifiedData["search_user_id"])
	}
}

func TestGoogleSheetIt_Execute_WithSpecificFields(t *testing.T) {
	helper := &GoogleSheetIt{}
	mockConn := &mockConnectorForGoogleSheetIt{
		contact: &connectors.NormalizedContact{
			ID:        "contact-123",
			FirstName: "John",
			Email:     "john@example.com",
		},
	}
	ctx := context.Background()

	fields := []interface{}{"FirstName", "Email", "Company"}
	input := helpers.HelperInput{
		ContactID: "contact-123",
		Config: map[string]interface{}{
			"spreadsheet_id":    "spreadsheet-123",
			"sheet_id":          "sheet-456",
			"google_account_id": "account-789",
			"fields":            fields,
		},
		Connector: mockConn,
	}

	output, err := helper.Execute(ctx, input)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if output.ModifiedData["fields"] == nil {
		t.Error("Expected fields to be included in ModifiedData")
	}
}

func TestGoogleSheetIt_Execute_ContactDataIncluded(t *testing.T) {
	helper := &GoogleSheetIt{}
	mockConn := &mockConnectorForGoogleSheetIt{
		contact: &connectors.NormalizedContact{
			ID:        "contact-123",
			FirstName: "John",
			LastName:  "Doe",
			Email:     "john.doe@example.com",
			Phone:     "+1-555-0123",
			Company:   "Acme Corp",
		},
	}
	ctx := context.Background()

	input := helpers.HelperInput{
		ContactID: "contact-123",
		Config: map[string]interface{}{
			"spreadsheet_id":    "spreadsheet-123",
			"sheet_id":          "sheet-456",
			"google_account_id": "account-789",
		},
		Connector: mockConn,
	}

	output, err := helper.Execute(ctx, input)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	contactData, ok := output.ModifiedData["contact_data"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected contact_data to be present in ModifiedData")
	}

	if contactData["Id"] != "contact-123" {
		t.Errorf("Expected contact Id to be 'contact-123', got: %v", contactData["Id"])
	}
	if contactData["Email"] != "john.doe@example.com" {
		t.Errorf("Expected contact Email to be 'john.doe@example.com', got: %v", contactData["Email"])
	}
}

func TestGoogleSheetIt_Execute_BulkReport(t *testing.T) {
	helper := &GoogleSheetIt{}
	mockConn := &mockConnectorForGoogleSheetIt{
		contact: &connectors.NormalizedContact{
			ID: "google_report",
		},
	}
	ctx := context.Background()

	input := helpers.HelperInput{
		ContactID: "google_report", // Special contact ID for bulk reports
		Config: map[string]interface{}{
			"spreadsheet_id":    "spreadsheet-123",
			"sheet_id":          "sheet-456",
			"google_account_id": "account-789",
		},
		Connector: mockConn,
	}

	output, err := helper.Execute(ctx, input)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Should not include contact_data for bulk reports
	if _, ok := output.ModifiedData["contact_data"]; ok {
		t.Error("Expected contact_data to be absent for bulk reports")
	}
}

func TestGoogleSheetIt_Execute_GetContactError(t *testing.T) {
	helper := &GoogleSheetIt{}
	mockConn := &mockConnectorForGoogleSheetIt{
		getContactErr: fmt.Errorf("CRM API error"),
	}
	ctx := context.Background()

	input := helpers.HelperInput{
		ContactID: "contact-123",
		Config: map[string]interface{}{
			"spreadsheet_id":    "spreadsheet-123",
			"sheet_id":          "sheet-456",
			"google_account_id": "account-789",
		},
		Connector: mockConn,
	}

	output, err := helper.Execute(ctx, input)
	// Should succeed even if GetContact fails (logged as warning)
	if err != nil {
		t.Fatalf("Expected no error (graceful handling), got: %v", err)
	}

	if !output.Success {
		t.Error("Expected success to be true even with GetContact error")
	}

	// Check logs for warning
	if len(output.Logs) == 0 {
		t.Error("Expected warning log for GetContact error")
	}
}

func TestGoogleSheetIt_Execute_ActionsRecorded(t *testing.T) {
	helper := &GoogleSheetIt{}
	mockConn := &mockConnectorForGoogleSheetIt{
		contact: &connectors.NormalizedContact{
			ID:    "contact-123",
			Email: "test@example.com",
		},
	}
	ctx := context.Background()

	input := helpers.HelperInput{
		ContactID: "contact-123",
		Config: map[string]interface{}{
			"spreadsheet_id":    "spreadsheet-123",
			"sheet_id":          "sheet-456",
			"google_account_id": "account-789",
		},
		Connector: mockConn,
	}

	output, err := helper.Execute(ctx, input)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(output.Actions) != 1 {
		t.Fatalf("Expected 1 action, got %d", len(output.Actions))
	}

	action := output.Actions[0]
	if action.Type != "google_sheet_sync_queued" {
		t.Errorf("Expected action type 'google_sheet_sync_queued', got '%s'", action.Type)
	}
	if action.Target != "spreadsheet-123" {
		t.Errorf("Expected action target to be 'spreadsheet-123', got '%s'", action.Target)
	}
}

func TestGoogleSheetIt_Execute_LogsRecorded(t *testing.T) {
	helper := &GoogleSheetIt{}
	mockConn := &mockConnectorForGoogleSheetIt{
		contact: &connectors.NormalizedContact{
			ID:    "contact-123",
			Email: "test@example.com",
		},
	}
	ctx := context.Background()

	input := helpers.HelperInput{
		ContactID: "contact-123",
		Config: map[string]interface{}{
			"spreadsheet_id":    "spreadsheet-123",
			"sheet_id":          "sheet-456",
			"google_account_id": "account-789",
		},
		Connector: mockConn,
	}

	output, err := helper.Execute(ctx, input)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(output.Logs) == 0 {
		t.Error("Expected logs to be recorded")
	}
}

func TestGoogleSheetIt_Execute_DefaultModeReplace(t *testing.T) {
	helper := &GoogleSheetIt{}
	mockConn := &mockConnectorForGoogleSheetIt{
		contact: &connectors.NormalizedContact{
			ID:    "contact-123",
			Email: "test@example.com",
		},
	}
	ctx := context.Background()

	input := helpers.HelperInput{
		ContactID: "contact-123",
		Config: map[string]interface{}{
			"spreadsheet_id":    "spreadsheet-123",
			"sheet_id":          "sheet-456",
			"google_account_id": "account-789",
			// mode not specified, should default to replace
		},
		Connector: mockConn,
	}

	output, err := helper.Execute(ctx, input)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if output.ModifiedData["mode"] != "replace" {
		t.Errorf("Expected mode to default to 'replace', got: %v", output.ModifiedData["mode"])
	}
}

func TestGoogleSheetIt_Execute_CustomFieldsIncluded(t *testing.T) {
	helper := &GoogleSheetIt{}
	mockConn := &mockConnectorForGoogleSheetIt{
		contact: &connectors.NormalizedContact{
			ID:        "contact-123",
			FirstName: "John",
			CustomFields: map[string]interface{}{
				"loyalty_points": 1500,
				"vip_status":     true,
			},
		},
	}
	ctx := context.Background()

	input := helpers.HelperInput{
		ContactID: "contact-123",
		Config: map[string]interface{}{
			"spreadsheet_id":    "spreadsheet-123",
			"sheet_id":          "sheet-456",
			"google_account_id": "account-789",
		},
		Connector: mockConn,
	}

	output, err := helper.Execute(ctx, input)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	contactData, ok := output.ModifiedData["contact_data"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected contact_data to be present")
	}

	if contactData["loyalty_points"] != 1500 {
		t.Error("Expected custom field loyalty_points to be included")
	}
	if contactData["vip_status"] != true {
		t.Error("Expected custom field vip_status to be included")
	}
}
