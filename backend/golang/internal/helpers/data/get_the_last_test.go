package data

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/myfusionhelper/api/internal/connectors"
	"github.com/myfusionhelper/api/internal/helpers"
)

// Mock connector for get_the_last testing (reusing pattern from get_the_first)
type mockConnectorForGetLast struct {
	fieldValues   map[string]interface{}
	updatedFields map[string]interface{}
	contactData   *connectors.NormalizedContact
}

func (m *mockConnectorForGetLast) GetContactFieldValue(ctx context.Context, contactID, fieldKey string) (interface{}, error) {
	if m.fieldValues == nil {
		return nil, nil
	}
	val, ok := m.fieldValues[fieldKey]
	if !ok {
		return nil, fmt.Errorf("field not found: %s", fieldKey)
	}
	return val, nil
}

func (m *mockConnectorForGetLast) SetContactFieldValue(ctx context.Context, contactID, fieldKey string, value interface{}) error {
	if m.updatedFields == nil {
		m.updatedFields = make(map[string]interface{})
	}
	m.updatedFields[fieldKey] = value
	return nil
}

func (m *mockConnectorForGetLast) GetContact(ctx context.Context, contactID string) (*connectors.NormalizedContact, error) {
	if m.contactData != nil {
		return m.contactData, nil
	}
	return nil, fmt.Errorf("contact not found")
}

// Stub implementations for CRMConnector interface
func (m *mockConnectorForGetLast) GetContacts(ctx context.Context, opts connectors.QueryOptions) (*connectors.ContactList, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForGetLast) CreateContact(ctx context.Context, input connectors.CreateContactInput) (*connectors.NormalizedContact, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForGetLast) UpdateContact(ctx context.Context, contactID string, updates connectors.UpdateContactInput) (*connectors.NormalizedContact, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForGetLast) DeleteContact(ctx context.Context, contactID string) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorForGetLast) GetTags(ctx context.Context) ([]connectors.Tag, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForGetLast) ApplyTag(ctx context.Context, contactID, tagID string) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorForGetLast) RemoveTag(ctx context.Context, contactID, tagID string) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorForGetLast) GetCustomFields(ctx context.Context) ([]connectors.CustomField, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForGetLast) TriggerAutomation(ctx context.Context, contactID, automationID string) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorForGetLast) AchieveGoal(ctx context.Context, contactID, goalName, integration string) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorForGetLast) TestConnection(ctx context.Context) error {
	return nil
}
func (m *mockConnectorForGetLast) GetMetadata() connectors.ConnectorMetadata {
	return connectors.ConnectorMetadata{}
}
func (m *mockConnectorForGetLast) GetCapabilities() []connectors.Capability {
	return nil
}

// Test helper metadata
func TestGetTheLast_Metadata(t *testing.T) {
	helper := &GetTheLast{}

	if helper.GetName() != "Get The Last" {
		t.Errorf("Expected name 'Get The Last', got '%s'", helper.GetName())
	}
	if helper.GetType() != "get_the_last" {
		t.Errorf("Expected type 'get_the_last', got '%s'", helper.GetType())
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
func TestGetTheLast_ValidateConfig_MissingType(t *testing.T) {
	helper := &GetTheLast{}
	config := map[string]interface{}{
		"from_field": "total",
		"to_field":   "last_invoice_total",
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
func TestGetTheLast_ValidateConfig_InvalidType(t *testing.T) {
	helper := &GetTheLast{}
	config := map[string]interface{}{
		"type":       "invalid_type",
		"from_field": "total",
		"to_field":   "last_total",
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
func TestGetTheLast_ValidateConfig_MissingFromField(t *testing.T) {
	helper := &GetTheLast{}
	config := map[string]interface{}{
		"type":     "invoice",
		"to_field": "last_invoice_total",
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
func TestGetTheLast_ValidateConfig_MissingToField(t *testing.T) {
	helper := &GetTheLast{}
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
func TestGetTheLast_ValidateConfig_Valid(t *testing.T) {
	helper := &GetTheLast{}

	validConfigs := []map[string]interface{}{
		{
			"type":       "invoice",
			"from_field": "total",
			"to_field":   "last_invoice_total",
		},
		{
			"type":       "job",
			"from_field": "status",
			"to_field":   "last_job_status",
		},
		{
			"type":       "subscription",
			"from_field": "plan_name",
			"to_field":   "last_subscription_plan",
		},
		{
			"type":       "lead",
			"from_field": "source",
			"to_field":   "last_lead_source",
		},
		{
			"type":       "creditcard",
			"from_field": "last_four",
			"to_field":   "last_card_last_four",
		},
		{
			"type":       "payment",
			"from_field": "amount",
			"to_field":   "last_payment_amount",
		},
	}

	for _, config := range validConfigs {
		err := helper.ValidateConfig(config)
		if err != nil {
			t.Errorf("Expected no validation error for %v, got: %v", config["type"], err)
		}
	}
}

// Test execution - invoice record found (newest)
func TestGetTheLast_Execute_InvoiceFound(t *testing.T) {
	helper := &GetTheLast{}

	mockConn := &mockConnectorForGetLast{
		fieldValues: map[string]interface{}{
			"_related.invoice.last.total": "599.99",
		},
	}

	input := helpers.HelperInput{
		ContactID: "123",
		Connector: mockConn,
		Config: map[string]interface{}{
			"type":       "invoice",
			"from_field": "total",
			"to_field":   "last_invoice_total",
		},
	}

	output, err := helper.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if !output.Success {
		t.Error("Expected success=true")
	}

	result := mockConn.updatedFields["last_invoice_total"]
	if result != "599.99" {
		t.Errorf("Expected '599.99', got: %v", result)
	}
}

// Test execution - subscription record found (newest)
func TestGetTheLast_Execute_SubscriptionFound(t *testing.T) {
	helper := &GetTheLast{}

	mockConn := &mockConnectorForGetLast{
		fieldValues: map[string]interface{}{
			"_related.subscription.last.plan_name": "Enterprise Plan",
		},
	}

	input := helpers.HelperInput{
		ContactID: "123",
		Connector: mockConn,
		Config: map[string]interface{}{
			"type":       "subscription",
			"from_field": "plan_name",
			"to_field":   "last_subscription_plan",
		},
	}

	output, err := helper.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if !output.Success {
		t.Error("Expected success=true")
	}

	result := mockConn.updatedFields["last_subscription_plan"]
	if result != "Enterprise Plan" {
		t.Errorf("Expected 'Enterprise Plan', got: %v", result)
	}
}

// Test execution - payment record found (newest)
func TestGetTheLast_Execute_PaymentFound(t *testing.T) {
	helper := &GetTheLast{}

	mockConn := &mockConnectorForGetLast{
		fieldValues: map[string]interface{}{
			"_related.payment.last.amount": "125.50",
		},
	}

	input := helpers.HelperInput{
		ContactID: "123",
		Connector: mockConn,
		Config: map[string]interface{}{
			"type":       "payment",
			"from_field": "amount",
			"to_field":   "last_payment_amount",
		},
	}

	output, err := helper.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if !output.Success {
		t.Error("Expected success=true")
	}

	result := mockConn.updatedFields["last_payment_amount"]
	if result != "125.50" {
		t.Errorf("Expected '125.50', got: %v", result)
	}
}

// Test execution - no records found (nil value)
func TestGetTheLast_Execute_NoRecordsFound(t *testing.T) {
	helper := &GetTheLast{}

	mockConn := &mockConnectorForGetLast{
		fieldValues: map[string]interface{}{
			"_related.lead.last.source": nil,
		},
	}

	input := helpers.HelperInput{
		ContactID: "123",
		Connector: mockConn,
		Config: map[string]interface{}{
			"type":       "lead",
			"from_field": "source",
			"to_field":   "last_lead_source",
		},
	}

	output, err := helper.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if !output.Success {
		t.Error("Expected success=true")
	}

	if !strings.Contains(output.Message, "No lead records found") {
		t.Errorf("Expected message about no records, got: %s", output.Message)
	}

	// Should not have updated the field
	if _, ok := mockConn.updatedFields["last_lead_source"]; ok {
		t.Error("Should not update field when no records found")
	}
}

// Test execution - fallback to contact custom fields
func TestGetTheLast_Execute_FallbackToContactFields(t *testing.T) {
	helper := &GetTheLast{}

	mockConn := &mockConnectorForGetLast{
		fieldValues: map[string]interface{}{}, // Empty, will trigger fallback
		contactData: &connectors.NormalizedContact{
			ID: "123",
			CustomFields: map[string]interface{}{
				"status": "Active",
			},
		},
	}

	input := helpers.HelperInput{
		ContactID: "123",
		Connector: mockConn,
		Config: map[string]interface{}{
			"type":       "job",
			"from_field": "status",
			"to_field":   "last_job_status",
		},
	}

	output, err := helper.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if !output.Success {
		t.Error("Expected success=true")
	}

	result := mockConn.updatedFields["last_job_status"]
	if result != "Active" {
		t.Errorf("Expected 'Active', got: %v", result)
	}
}

// Test execution - fallback when field not in custom fields
func TestGetTheLast_Execute_FallbackNoField(t *testing.T) {
	helper := &GetTheLast{}

	mockConn := &mockConnectorForGetLast{
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
			"type":       "creditcard",
			"from_field": "last_four",
			"to_field":   "last_card_last_four",
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
	if _, ok := mockConn.updatedFields["last_card_last_four"]; ok {
		t.Error("Should not update field when no value found")
	}
}

// Test action logging
func TestGetTheLast_Execute_ActionLogging(t *testing.T) {
	helper := &GetTheLast{}

	mockConn := &mockConnectorForGetLast{
		fieldValues: map[string]interface{}{
			"_related.invoice.last.total": "899.00",
		},
	}

	input := helpers.HelperInput{
		ContactID: "123",
		Connector: mockConn,
		Config: map[string]interface{}{
			"type":       "invoice",
			"from_field": "total",
			"to_field":   "last_invoice_total",
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
	if action.Target != "last_invoice_total" {
		t.Errorf("Expected action target 'last_invoice_total', got '%s'", action.Target)
	}

	// Verify logs
	if len(output.Logs) == 0 {
		t.Error("Expected logs to be generated")
	}

	// Verify modified data
	if output.ModifiedData == nil {
		t.Fatal("Expected ModifiedData to be set")
	}
	if output.ModifiedData["last_invoice_total"] != "899.00" {
		t.Errorf("Expected ModifiedData['last_invoice_total'] = '899.00', got: %v", output.ModifiedData["last_invoice_total"])
	}
}
