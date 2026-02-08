package data

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/myfusionhelper/api/internal/connectors"
	"github.com/myfusionhelper/api/internal/helpers"
)

// Mock connector for get_the_first testing
type mockConnectorForGetFirst struct {
	fieldValues   map[string]interface{}
	updatedFields map[string]interface{}
	contactData   *connectors.NormalizedContact
}

func (m *mockConnectorForGetFirst) GetContactFieldValue(ctx context.Context, contactID, fieldKey string) (interface{}, error) {
	if m.fieldValues == nil {
		return nil, nil
	}
	val, ok := m.fieldValues[fieldKey]
	if !ok {
		return nil, fmt.Errorf("field not found: %s", fieldKey)
	}
	return val, nil
}

func (m *mockConnectorForGetFirst) SetContactFieldValue(ctx context.Context, contactID, fieldKey string, value interface{}) error {
	if m.updatedFields == nil {
		m.updatedFields = make(map[string]interface{})
	}
	m.updatedFields[fieldKey] = value
	return nil
}

func (m *mockConnectorForGetFirst) GetContact(ctx context.Context, contactID string) (*connectors.NormalizedContact, error) {
	if m.contactData != nil {
		return m.contactData, nil
	}
	return nil, fmt.Errorf("contact not found")
}

// Stub implementations for CRMConnector interface
func (m *mockConnectorForGetFirst) GetContacts(ctx context.Context, opts connectors.QueryOptions) (*connectors.ContactList, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForGetFirst) CreateContact(ctx context.Context, input connectors.CreateContactInput) (*connectors.NormalizedContact, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForGetFirst) UpdateContact(ctx context.Context, contactID string, updates connectors.UpdateContactInput) (*connectors.NormalizedContact, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForGetFirst) DeleteContact(ctx context.Context, contactID string) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorForGetFirst) GetTags(ctx context.Context) ([]connectors.Tag, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForGetFirst) ApplyTag(ctx context.Context, contactID, tagID string) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorForGetFirst) RemoveTag(ctx context.Context, contactID, tagID string) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorForGetFirst) GetCustomFields(ctx context.Context) ([]connectors.CustomField, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForGetFirst) TriggerAutomation(ctx context.Context, contactID, automationID string) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorForGetFirst) AchieveGoal(ctx context.Context, contactID, goalName, integration string) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorForGetFirst) TestConnection(ctx context.Context) error {
	return nil
}
func (m *mockConnectorForGetFirst) GetMetadata() connectors.ConnectorMetadata {
	return connectors.ConnectorMetadata{}
}
func (m *mockConnectorForGetFirst) GetCapabilities() []connectors.Capability {
	return nil
}

// Test helper metadata
func TestGetTheFirst_Metadata(t *testing.T) {
	helper := &GetTheFirst{}

	if helper.GetName() != "Get The First" {
		t.Errorf("Expected name 'Get The First', got '%s'", helper.GetName())
	}
	if helper.GetType() != "get_the_first" {
		t.Errorf("Expected type 'get_the_first', got '%s'", helper.GetType())
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

// Test validation - missing type
func TestGetTheFirst_ValidateConfig_MissingType(t *testing.T) {
	helper := &GetTheFirst{}
	config := map[string]interface{}{
		"from_field": "total",
		"to_field":   "first_invoice_total",
	}

	err := helper.ValidateConfig(config)
	if err == nil {
		t.Error("Expected validation error for missing type")
	}
	if !strings.Contains(err.Error(), "type") {
		t.Errorf("Expected error about type, got: %v", err)
	}
}

// Test validation - invalid type
func TestGetTheFirst_ValidateConfig_InvalidType(t *testing.T) {
	helper := &GetTheFirst{}
	config := map[string]interface{}{
		"type":       "invalid_type",
		"from_field": "total",
		"to_field":   "first_total",
	}

	err := helper.ValidateConfig(config)
	if err == nil {
		t.Error("Expected validation error for invalid type")
	}
	if !strings.Contains(err.Error(), "invalid type") {
		t.Errorf("Expected error about invalid type, got: %v", err)
	}
}

// Test validation - missing from_field
func TestGetTheFirst_ValidateConfig_MissingFromField(t *testing.T) {
	helper := &GetTheFirst{}
	config := map[string]interface{}{
		"type":     "invoice",
		"to_field": "first_invoice_total",
	}

	err := helper.ValidateConfig(config)
	if err == nil {
		t.Error("Expected validation error for missing from_field")
	}
	if !strings.Contains(err.Error(), "from_field") {
		t.Errorf("Expected error about from_field, got: %v", err)
	}
}

// Test validation - missing to_field
func TestGetTheFirst_ValidateConfig_MissingToField(t *testing.T) {
	helper := &GetTheFirst{}
	config := map[string]interface{}{
		"type":       "invoice",
		"from_field": "total",
	}

	err := helper.ValidateConfig(config)
	if err == nil {
		t.Error("Expected validation error for missing to_field")
	}
	if !strings.Contains(err.Error(), "to_field") {
		t.Errorf("Expected error about to_field, got: %v", err)
	}
}

// Test validation - valid configs
func TestGetTheFirst_ValidateConfig_Valid(t *testing.T) {
	helper := &GetTheFirst{}

	validConfigs := []map[string]interface{}{
		{
			"type":       "invoice",
			"from_field": "total",
			"to_field":   "first_invoice_total",
		},
		{
			"type":       "job",
			"from_field": "status",
			"to_field":   "first_job_status",
		},
		{
			"type":       "subscription",
			"from_field": "plan_name",
			"to_field":   "first_subscription_plan",
		},
		{
			"type":       "lead",
			"from_field": "source",
			"to_field":   "first_lead_source",
		},
		{
			"type":       "creditcard",
			"from_field": "last_four",
			"to_field":   "first_card_last_four",
		},
		{
			"type":       "payment",
			"from_field": "amount",
			"to_field":   "first_payment_amount",
		},
	}

	for _, config := range validConfigs {
		err := helper.ValidateConfig(config)
		if err != nil {
			t.Errorf("Expected no validation error for %v, got: %v", config["type"], err)
		}
	}
}

// Test execution - invoice record found
func TestGetTheFirst_Execute_InvoiceFound(t *testing.T) {
	helper := &GetTheFirst{}

	mockConn := &mockConnectorForGetFirst{
		fieldValues: map[string]interface{}{
			"_related.invoice.first.total": "299.99",
		},
	}

	input := helpers.HelperInput{
		ContactID: "123",
		Connector: mockConn,
		Config: map[string]interface{}{
			"type":       "invoice",
			"from_field": "total",
			"to_field":   "first_invoice_total",
		},
	}

	output, err := helper.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if !output.Success {
		t.Error("Expected success=true")
	}

	result := mockConn.updatedFields["first_invoice_total"]
	if result != "299.99" {
		t.Errorf("Expected '299.99', got: %v", result)
	}
}

// Test execution - subscription record found
func TestGetTheFirst_Execute_SubscriptionFound(t *testing.T) {
	helper := &GetTheFirst{}

	mockConn := &mockConnectorForGetFirst{
		fieldValues: map[string]interface{}{
			"_related.subscription.first.plan_name": "Premium Plan",
		},
	}

	input := helpers.HelperInput{
		ContactID: "123",
		Connector: mockConn,
		Config: map[string]interface{}{
			"type":       "subscription",
			"from_field": "plan_name",
			"to_field":   "first_subscription_plan",
		},
	}

	output, err := helper.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if !output.Success {
		t.Error("Expected success=true")
	}

	result := mockConn.updatedFields["first_subscription_plan"]
	if result != "Premium Plan" {
		t.Errorf("Expected 'Premium Plan', got: %v", result)
	}
}

// Test execution - job record found
func TestGetTheFirst_Execute_JobFound(t *testing.T) {
	helper := &GetTheFirst{}

	mockConn := &mockConnectorForGetFirst{
		fieldValues: map[string]interface{}{
			"_related.job.first.status": "Completed",
		},
	}

	input := helpers.HelperInput{
		ContactID: "123",
		Connector: mockConn,
		Config: map[string]interface{}{
			"type":       "job",
			"from_field": "status",
			"to_field":   "first_job_status",
		},
	}

	output, err := helper.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if !output.Success {
		t.Error("Expected success=true")
	}

	result := mockConn.updatedFields["first_job_status"]
	if result != "Completed" {
		t.Errorf("Expected 'Completed', got: %v", result)
	}
}

// Test execution - no records found (nil value)
func TestGetTheFirst_Execute_NoRecordsFound(t *testing.T) {
	helper := &GetTheFirst{}

	mockConn := &mockConnectorForGetFirst{
		fieldValues: map[string]interface{}{
			"_related.payment.first.amount": nil,
		},
	}

	input := helpers.HelperInput{
		ContactID: "123",
		Connector: mockConn,
		Config: map[string]interface{}{
			"type":       "payment",
			"from_field": "amount",
			"to_field":   "first_payment_amount",
		},
	}

	output, err := helper.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if !output.Success {
		t.Error("Expected success=true")
	}

	if !strings.Contains(output.Message, "No payment records found") {
		t.Errorf("Expected message about no records, got: %s", output.Message)
	}

	// Should not have updated the field
	if _, ok := mockConn.updatedFields["first_payment_amount"]; ok {
		t.Error("Should not update field when no records found")
	}
}

// Test execution - fallback to contact custom fields
func TestGetTheFirst_Execute_FallbackToContactFields(t *testing.T) {
	helper := &GetTheFirst{}

	mockConn := &mockConnectorForGetFirst{
		fieldValues: map[string]interface{}{}, // Empty, will trigger fallback
		contactData: &connectors.NormalizedContact{
			ID: "123",
			CustomFields: map[string]interface{}{
				"total": "450.00",
			},
		},
	}

	input := helpers.HelperInput{
		ContactID: "123",
		Connector: mockConn,
		Config: map[string]interface{}{
			"type":       "invoice",
			"from_field": "total",
			"to_field":   "first_invoice_total",
		},
	}

	output, err := helper.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if !output.Success {
		t.Error("Expected success=true")
	}

	result := mockConn.updatedFields["first_invoice_total"]
	if result != "450.00" {
		t.Errorf("Expected '450.00', got: %v", result)
	}
}

// Test execution - fallback when field not in custom fields
func TestGetTheFirst_Execute_FallbackNoField(t *testing.T) {
	helper := &GetTheFirst{}

	mockConn := &mockConnectorForGetFirst{
		fieldValues: map[string]interface{}{}, // Empty, will trigger fallback
		contactData: &connectors.NormalizedContact{
			ID:           "123",
			CustomFields: map[string]interface{}{}, // No matching field
		},
	}

	input := helpers.HelperInput{
		ContactID: "123",
		Connector: mockConn,
		Config: map[string]interface{}{
			"type":       "lead",
			"from_field": "source",
			"to_field":   "first_lead_source",
		},
	}

	output, err := helper.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if !output.Success {
		t.Error("Expected success=true")
	}

	// Should not have updated the field
	if _, ok := mockConn.updatedFields["first_lead_source"]; ok {
		t.Error("Should not update field when no value found")
	}
}

// Test action logging
func TestGetTheFirst_Execute_ActionLogging(t *testing.T) {
	helper := &GetTheFirst{}

	mockConn := &mockConnectorForGetFirst{
		fieldValues: map[string]interface{}{
			"_related.creditcard.first.last_four": "1234",
		},
	}

	input := helpers.HelperInput{
		ContactID: "123",
		Connector: mockConn,
		Config: map[string]interface{}{
			"type":       "creditcard",
			"from_field": "last_four",
			"to_field":   "first_card_last_four",
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
	if action.Target != "first_card_last_four" {
		t.Errorf("Expected action target 'first_card_last_four', got '%s'", action.Target)
	}

	// Verify logs
	if len(output.Logs) == 0 {
		t.Error("Expected logs to be generated")
	}

	// Verify modified data
	if output.ModifiedData == nil {
		t.Fatal("Expected ModifiedData to be set")
	}
	if output.ModifiedData["first_card_last_four"] != "1234" {
		t.Errorf("Expected ModifiedData['first_card_last_four'] = '1234', got: %v", output.ModifiedData["first_card_last_four"])
	}
}
