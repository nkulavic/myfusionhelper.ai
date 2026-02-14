package data

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/myfusionhelper/api/internal/connectors"
	"github.com/myfusionhelper/api/internal/helpers"
)

// Mock connector for date_calc testing
type mockConnectorForDateCalc struct {
	fieldValues   map[string]interface{}
	updatedFields map[string]interface{}
}

func (m *mockConnectorForDateCalc) GetContactFieldValue(ctx context.Context, contactID, fieldKey string) (interface{}, error) {
	if m.fieldValues == nil {
		return nil, nil
	}
	val, ok := m.fieldValues[fieldKey]
	if !ok {
		return nil, nil
	}
	return val, nil
}

func (m *mockConnectorForDateCalc) SetContactFieldValue(ctx context.Context, contactID, fieldKey string, value interface{}) error {
	if m.updatedFields == nil {
		m.updatedFields = make(map[string]interface{})
	}
	m.updatedFields[fieldKey] = value
	return nil
}

// Stub implementations for CRMConnector interface
func (m *mockConnectorForDateCalc) GetContact(ctx context.Context, contactID string) (*connectors.NormalizedContact, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForDateCalc) GetContacts(ctx context.Context, opts connectors.QueryOptions) (*connectors.ContactList, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForDateCalc) CreateContact(ctx context.Context, input connectors.CreateContactInput) (*connectors.NormalizedContact, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForDateCalc) UpdateContact(ctx context.Context, contactID string, updates connectors.UpdateContactInput) (*connectors.NormalizedContact, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForDateCalc) DeleteContact(ctx context.Context, contactID string) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorForDateCalc) GetTags(ctx context.Context) ([]connectors.Tag, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForDateCalc) ApplyTag(ctx context.Context, contactID, tagID string) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorForDateCalc) RemoveTag(ctx context.Context, contactID, tagID string) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorForDateCalc) GetCustomFields(ctx context.Context) ([]connectors.CustomField, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForDateCalc) TriggerAutomation(ctx context.Context, contactID, automationID string) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorForDateCalc) AchieveGoal(ctx context.Context, contactID, goalName, integration string) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorForDateCalc) TestConnection(ctx context.Context) error {
	return nil
}
func (m *mockConnectorForDateCalc) GetMetadata() connectors.ConnectorMetadata {
	return connectors.ConnectorMetadata{}
}
func (m *mockConnectorForDateCalc) GetCapabilities() []connectors.Capability {
	return nil
}

// Test helper metadata
func TestDateCalc_Metadata(t *testing.T) {
	helper := &DateCalc{}

	if helper.GetName() != "Date Calc" {
		t.Errorf("Expected name 'Date Calc', got '%s'", helper.GetName())
	}
	if helper.GetType() != "date_calc" {
		t.Errorf("Expected type 'date_calc', got '%s'", helper.GetType())
	}
	if helper.GetCategory() != "data" {
		t.Errorf("Expected category 'data', got '%s'", helper.GetCategory())
	}
	if !helper.RequiresCRM() {
		t.Error("Expected RequiresCRM to be true")
	}

	schema := helper.GetConfigSchema()
	if schema == nil {
		t.Fatal("Expected config schema, got nil")
	}
}

// Test validation - missing operation
func TestDateCalc_ValidateConfig_MissingOperation(t *testing.T) {
	helper := &DateCalc{}
	config := map[string]interface{}{
		"field":  "created_date",
		"amount": 7,
	}

	err := helper.ValidateConfig(config)
	if err == nil {
		t.Error("Expected validation error for missing operation")
	}
	if !strings.Contains(err.Error(), "operation") {
		t.Errorf("Expected error about operation, got: %v", err)
	}
}

// Test validation - invalid operation
func TestDateCalc_ValidateConfig_InvalidOperation(t *testing.T) {
	helper := &DateCalc{}
	config := map[string]interface{}{
		"operation": "invalid_op",
		"field":     "created_date",
	}

	err := helper.ValidateConfig(config)
	if err == nil {
		t.Error("Expected validation error for invalid operation")
	}
	if !strings.Contains(err.Error(), "invalid operation") {
		t.Errorf("Expected error about invalid operation, got: %v", err)
	}
}

// Test validation - missing field for add_days
func TestDateCalc_ValidateConfig_MissingFieldForAddDays(t *testing.T) {
	helper := &DateCalc{}
	config := map[string]interface{}{
		"operation": "add_days",
		"amount":    7,
	}

	err := helper.ValidateConfig(config)
	if err == nil {
		t.Error("Expected validation error for missing field")
	}
	if !strings.Contains(err.Error(), "field") {
		t.Errorf("Expected error about field, got: %v", err)
	}
}

// Test validation - missing amount for add_days
func TestDateCalc_ValidateConfig_MissingAmountForAddDays(t *testing.T) {
	helper := &DateCalc{}
	config := map[string]interface{}{
		"operation": "add_days",
		"field":     "created_date",
	}

	err := helper.ValidateConfig(config)
	if err == nil {
		t.Error("Expected validation error for missing amount")
	}
	if !strings.Contains(err.Error(), "amount") {
		t.Errorf("Expected error about amount, got: %v", err)
	}
}

// Test validation - missing compare_field for diff_days
func TestDateCalc_ValidateConfig_MissingCompareField(t *testing.T) {
	helper := &DateCalc{}
	config := map[string]interface{}{
		"operation": "diff_days",
		"field":     "start_date",
	}

	err := helper.ValidateConfig(config)
	if err == nil {
		t.Error("Expected validation error for missing compare_field")
	}
	if !strings.Contains(err.Error(), "compare_field") {
		t.Errorf("Expected error about compare_field, got: %v", err)
	}
}

// Test validation - valid configs
func TestDateCalc_ValidateConfig_Valid(t *testing.T) {
	helper := &DateCalc{}

	validConfigs := []map[string]interface{}{
		{
			"operation": "add_days",
			"field":     "created_date",
			"amount":    7,
		},
		{
			"operation": "subtract_months",
			"field":     "expiry_date",
			"amount":    3,
		},
		{
			"operation": "set_now",
		},
		{
			"operation":     "diff_days",
			"field":         "start_date",
			"compare_field": "end_date",
		},
		{
			"operation": "format",
			"field":     "created_date",
		},
	}

	for _, config := range validConfigs {
		err := helper.ValidateConfig(config)
		if err != nil {
			t.Errorf("Expected no validation error for %v, got: %v", config["operation"], err)
		}
	}
}

// Test execution - add_days
func TestDateCalc_Execute_AddDays(t *testing.T) {
	helper := &DateCalc{}

	mockConn := &mockConnectorForDateCalc{
		fieldValues: map[string]interface{}{
			"start_date": "2024-01-15",
		},
	}

	input := helpers.HelperInput{
		ContactID: "123",
		Connector: mockConn,
		Config: map[string]interface{}{
			"operation": "add_days",
			"field":     "start_date",
			"amount":    7,
		},
	}

	output, err := helper.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if !output.Success {
		t.Error("Expected success=true")
	}

	result := mockConn.updatedFields["start_date"]
	if result != "2024-01-22" {
		t.Errorf("Expected '2024-01-22', got: %v", result)
	}
}

// Test execution - subtract_days
func TestDateCalc_Execute_SubtractDays(t *testing.T) {
	helper := &DateCalc{}

	mockConn := &mockConnectorForDateCalc{
		fieldValues: map[string]interface{}{
			"end_date": "2024-01-15",
		},
	}

	input := helpers.HelperInput{
		ContactID: "123",
		Connector: mockConn,
		Config: map[string]interface{}{
			"operation": "subtract_days",
			"field":     "end_date",
			"amount":    10,
		},
	}

	output, err := helper.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if !output.Success {
		t.Error("Expected success=true")
	}

	result := mockConn.updatedFields["end_date"]
	if result != "2024-01-05" {
		t.Errorf("Expected '2024-01-05', got: %v", result)
	}
}

// Test execution - add_months
func TestDateCalc_Execute_AddMonths(t *testing.T) {
	helper := &DateCalc{}

	mockConn := &mockConnectorForDateCalc{
		fieldValues: map[string]interface{}{
			"subscription_date": "2024-01-15",
		},
	}

	input := helpers.HelperInput{
		ContactID: "123",
		Connector: mockConn,
		Config: map[string]interface{}{
			"operation": "add_months",
			"field":     "subscription_date",
			"amount":    3,
		},
	}

	output, err := helper.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if !output.Success {
		t.Error("Expected success=true")
	}

	result := mockConn.updatedFields["subscription_date"]
	if result != "2024-04-15" {
		t.Errorf("Expected '2024-04-15', got: %v", result)
	}
}

// Test execution - subtract_months
func TestDateCalc_Execute_SubtractMonths(t *testing.T) {
	helper := &DateCalc{}

	mockConn := &mockConnectorForDateCalc{
		fieldValues: map[string]interface{}{
			"review_date": "2024-06-15",
		},
	}

	input := helpers.HelperInput{
		ContactID: "123",
		Connector: mockConn,
		Config: map[string]interface{}{
			"operation": "subtract_months",
			"field":     "review_date",
			"amount":    2,
		},
	}

	output, err := helper.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if !output.Success {
		t.Error("Expected success=true")
	}

	result := mockConn.updatedFields["review_date"]
	if result != "2024-04-15" {
		t.Errorf("Expected '2024-04-15', got: %v", result)
	}
}

// Test execution - add_years
func TestDateCalc_Execute_AddYears(t *testing.T) {
	helper := &DateCalc{}

	mockConn := &mockConnectorForDateCalc{
		fieldValues: map[string]interface{}{
			"birth_date": "2000-01-15",
		},
	}

	input := helpers.HelperInput{
		ContactID: "123",
		Connector: mockConn,
		Config: map[string]interface{}{
			"operation": "add_years",
			"field":     "birth_date",
			"amount":    25,
		},
	}

	output, err := helper.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if !output.Success {
		t.Error("Expected success=true")
	}

	result := mockConn.updatedFields["birth_date"]
	if result != "2025-01-15" {
		t.Errorf("Expected '2025-01-15', got: %v", result)
	}
}

// Test execution - subtract_years
func TestDateCalc_Execute_SubtractYears(t *testing.T) {
	helper := &DateCalc{}

	mockConn := &mockConnectorForDateCalc{
		fieldValues: map[string]interface{}{
			"warranty_end": "2026-12-31",
		},
	}

	input := helpers.HelperInput{
		ContactID: "123",
		Connector: mockConn,
		Config: map[string]interface{}{
			"operation": "subtract_years",
			"field":     "warranty_end",
			"amount":    1,
		},
	}

	output, err := helper.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if !output.Success {
		t.Error("Expected success=true")
	}

	result := mockConn.updatedFields["warranty_end"]
	if result != "2025-12-31" {
		t.Errorf("Expected '2025-12-31', got: %v", result)
	}
}

// Test execution - set_now
func TestDateCalc_Execute_SetNow(t *testing.T) {
	helper := &DateCalc{}

	mockConn := &mockConnectorForDateCalc{
		fieldValues: map[string]interface{}{},
	}

	input := helpers.HelperInput{
		ContactID: "123",
		Connector: mockConn,
		Config: map[string]interface{}{
			"operation":    "set_now",
			"target_field": "last_updated",
		},
	}

	output, err := helper.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if !output.Success {
		t.Error("Expected success=true")
	}

	result := mockConn.updatedFields["last_updated"]
	if result == nil {
		t.Fatal("Expected date to be set")
	}

	// Verify format is YYYY-MM-DD
	resultStr := result.(string)
	if len(resultStr) != 10 {
		t.Errorf("Expected format YYYY-MM-DD, got: %s", resultStr)
	}

	// Verify it's today's date
	today := time.Now().UTC().Format("2006-01-02")
	if resultStr != today {
		t.Errorf("Expected today's date '%s', got: %s", today, resultStr)
	}
}

// Test execution - diff_days
func TestDateCalc_Execute_DiffDays(t *testing.T) {
	helper := &DateCalc{}

	mockConn := &mockConnectorForDateCalc{
		fieldValues: map[string]interface{}{
			"start_date": "2024-01-01",
			"end_date":   "2024-01-15",
		},
	}

	input := helpers.HelperInput{
		ContactID: "123",
		Connector: mockConn,
		Config: map[string]interface{}{
			"operation":     "diff_days",
			"field":         "start_date",
			"compare_field": "end_date",
			"target_field":  "duration_days",
		},
	}

	output, err := helper.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if !output.Success {
		t.Error("Expected success=true")
	}

	result := mockConn.updatedFields["duration_days"]
	if result != "14" {
		t.Errorf("Expected '14', got: %v", result)
	}
}

// Test execution - format with custom format
func TestDateCalc_Execute_FormatCustom(t *testing.T) {
	helper := &DateCalc{}

	mockConn := &mockConnectorForDateCalc{
		fieldValues: map[string]interface{}{
			"created_date": "2024-01-15",
		},
	}

	input := helpers.HelperInput{
		ContactID: "123",
		Connector: mockConn,
		Config: map[string]interface{}{
			"operation":     "format",
			"field":         "created_date",
			"target_field":  "formatted_date",
			"output_format": "January 2, 2006",
		},
	}

	output, err := helper.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if !output.Success {
		t.Error("Expected success=true")
	}

	result := mockConn.updatedFields["formatted_date"]
	if result != "January 15, 2024" {
		t.Errorf("Expected 'January 15, 2024', got: %v", result)
	}
}

// Test execution - flexible date parsing (RFC3339)
func TestDateCalc_Execute_FlexibleDateParsing_RFC3339(t *testing.T) {
	helper := &DateCalc{}

	mockConn := &mockConnectorForDateCalc{
		fieldValues: map[string]interface{}{
			"created_date": "2024-01-15T10:30:00Z",
		},
	}

	input := helpers.HelperInput{
		ContactID: "123",
		Connector: mockConn,
		Config: map[string]interface{}{
			"operation": "add_days",
			"field":     "created_date",
			"amount":    5,
		},
	}

	output, err := helper.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if !output.Success {
		t.Error("Expected success=true")
	}

	result := mockConn.updatedFields["created_date"]
	if result != "2024-01-20" {
		t.Errorf("Expected '2024-01-20', got: %v", result)
	}
}

// Test execution - flexible date parsing (MM/DD/YYYY)
func TestDateCalc_Execute_FlexibleDateParsing_MMDDYYYY(t *testing.T) {
	helper := &DateCalc{}

	mockConn := &mockConnectorForDateCalc{
		fieldValues: map[string]interface{}{
			"event_date": "01/15/2024",
		},
	}

	input := helpers.HelperInput{
		ContactID: "123",
		Connector: mockConn,
		Config: map[string]interface{}{
			"operation": "add_days",
			"field":     "event_date",
			"amount":    3,
		},
	}

	output, err := helper.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if !output.Success {
		t.Error("Expected success=true")
	}

	result := mockConn.updatedFields["event_date"]
	if result != "2024-01-18" {
		t.Errorf("Expected '2024-01-18', got: %v", result)
	}
}

// Test execution - target_field different from source
func TestDateCalc_Execute_TargetField(t *testing.T) {
	helper := &DateCalc{}

	mockConn := &mockConnectorForDateCalc{
		fieldValues: map[string]interface{}{
			"order_date": "2024-01-15",
		},
	}

	input := helpers.HelperInput{
		ContactID: "123",
		Connector: mockConn,
		Config: map[string]interface{}{
			"operation":    "add_days",
			"field":        "order_date",
			"amount":       30,
			"target_field": "estimated_delivery",
		},
	}

	output, err := helper.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if !output.Success {
		t.Error("Expected success=true")
	}

	// Original field should not be updated
	if _, ok := mockConn.updatedFields["order_date"]; ok {
		t.Error("Should not update source field when target_field is specified")
	}

	// Target field should be updated
	result := mockConn.updatedFields["estimated_delivery"]
	if result != "2024-02-14" {
		t.Errorf("Expected '2024-02-14', got: %v", result)
	}
}

// Test action logging
func TestDateCalc_Execute_ActionLogging(t *testing.T) {
	helper := &DateCalc{}

	mockConn := &mockConnectorForDateCalc{
		fieldValues: map[string]interface{}{
			"renewal_date": "2024-06-15",
		},
	}

	input := helpers.HelperInput{
		ContactID: "123",
		Connector: mockConn,
		Config: map[string]interface{}{
			"operation": "add_years",
			"field":     "renewal_date",
			"amount":    1,
		},
	}

	output, err := helper.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Verify action was logged
	if len(output.Actions) != 1 {
		t.Fatalf("Expected 1 action, got %d", len(output.Actions))
	}

	action := output.Actions[0]
	if action.Type != "field_updated" {
		t.Errorf("Expected action type 'field_updated', got '%s'", action.Type)
	}
	if action.Target != "renewal_date" {
		t.Errorf("Expected action target 'renewal_date', got '%s'", action.Target)
	}

	// Verify logs
	if len(output.Logs) == 0 {
		t.Error("Expected logs to be generated")
	}

	// Verify modified data
	if output.ModifiedData == nil {
		t.Fatal("Expected ModifiedData to be set")
	}
	if output.ModifiedData["renewal_date"] != "2025-06-15" {
		t.Errorf("Expected ModifiedData['renewal_date'] = '2025-06-15', got: %v", output.ModifiedData["renewal_date"])
	}
}
