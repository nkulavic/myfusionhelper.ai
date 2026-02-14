package analytics

import (
	"context"
	"fmt"
	"testing"

	"github.com/myfusionhelper/api/internal/connectors"
	"github.com/myfusionhelper/api/internal/helpers"
)

type mockConnectorForCLV struct {
	fieldValues   map[string]interface{}
	getFieldError map[string]error
	setFieldError error
}

func (m *mockConnectorForCLV) GetContactFieldValue(ctx context.Context, contactID, fieldKey string) (interface{}, error) {
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

func (m *mockConnectorForCLV) SetContactFieldValue(ctx context.Context, contactID, fieldKey string, value interface{}) error {
	if m.setFieldError != nil {
		return m.setFieldError
	}
	if m.fieldValues == nil {
		m.fieldValues = make(map[string]interface{})
	}
	m.fieldValues[fieldKey] = value
	return nil
}

func (m *mockConnectorForCLV) GetContacts(ctx context.Context, opts connectors.QueryOptions) (*connectors.ContactList, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForCLV) GetContact(ctx context.Context, contactID string) (*connectors.NormalizedContact, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForCLV) CreateContact(ctx context.Context, contact connectors.CreateContactInput) (*connectors.NormalizedContact, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForCLV) UpdateContact(ctx context.Context, contactID string, updates connectors.UpdateContactInput) (*connectors.NormalizedContact, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForCLV) DeleteContact(ctx context.Context, contactID string) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorForCLV) GetTags(ctx context.Context) ([]connectors.Tag, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForCLV) ApplyTag(ctx context.Context, contactID, tagID string) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorForCLV) RemoveTag(ctx context.Context, contactID, tagID string) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorForCLV) GetCustomFields(ctx context.Context) ([]connectors.CustomField, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForCLV) TriggerAutomation(ctx context.Context, contactID, automationID string) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorForCLV) AchieveGoal(ctx context.Context, contactID, goalName, integration string) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorForCLV) TestConnection(ctx context.Context) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorForCLV) GetMetadata() connectors.ConnectorMetadata {
	return connectors.ConnectorMetadata{}
}
func (m *mockConnectorForCLV) GetCapabilities() []connectors.Capability {
	return nil
}
func (m *mockConnectorForCLV) SetOptInStatus(ctx context.Context, contactID string, optIn bool, reason string) error {
	return nil
}

func TestCustomerLifetimeValue_GetMetadata(t *testing.T) {
	h := &CustomerLifetimeValue{}
	if h.GetName() != "Customer Lifetime Value" {
		t.Error("wrong name")
	}
	if h.GetType() != "customer_lifetime_value" {
		t.Error("wrong type")
	}
	if h.GetCategory() != "analytics" {
		t.Error("wrong category")
	}
	if !h.RequiresCRM() {
		t.Error("should require CRM")
	}
}

func TestCustomerLifetimeValue_ValidateConfig_NoSaveFields(t *testing.T) {
	err := (&CustomerLifetimeValue{}).ValidateConfig(map[string]interface{}{})
	if err == nil {
		t.Error("should error when no save fields provided")
	}
}

func TestCustomerLifetimeValue_ValidateConfig_AllNoSave(t *testing.T) {
	err := (&CustomerLifetimeValue{}).ValidateConfig(map[string]interface{}{
		"lcv_total_orders":  "no_save",
		"lcv_total_spend":   "no_save",
		"lcv_average_order": "no_save",
		"lcv_total_due":     "no_save",
	})
	if err == nil {
		t.Error("should error when all fields are no_save")
	}
}

func TestCustomerLifetimeValue_ValidateConfig_Valid(t *testing.T) {
	err := (&CustomerLifetimeValue{}).ValidateConfig(map[string]interface{}{
		"lcv_total_orders": "TotalOrders",
	})
	if err != nil {
		t.Errorf("should be valid: %v", err)
	}
}

func TestCustomerLifetimeValue_Execute_AllFields(t *testing.T) {
	mock := &mockConnectorForCLV{
		fieldValues: map[string]interface{}{
			"_related.invoice.sum.TotalPaid":    1500.50,
			"_related.invoice.sum.TotalDue":     2000.00,
			"_related.invoice.count.nonzero":    10.0,
		},
	}

	output, err := (&CustomerLifetimeValue{}).Execute(context.Background(), helpers.HelperInput{
		ContactID: "contact_123",
		Config: map[string]interface{}{
			"lcv_total_orders":  "TotalOrders",
			"lcv_total_spend":   "TotalSpend",
			"lcv_average_order": "AvgOrder",
			"lcv_total_due":     "TotalDue",
		},
		Connector: mock,
	})

	if err != nil {
		t.Fatal(err)
	}
	if !output.Success {
		t.Error("should succeed")
	}
	if len(output.Actions) != 4 {
		t.Fatalf("expected 4 actions, got %d", len(output.Actions))
	}

	// Verify calculations
	if mock.fieldValues["TotalOrders"] != "10" {
		t.Errorf("expected TotalOrders=10, got %v", mock.fieldValues["TotalOrders"])
	}
	if mock.fieldValues["TotalSpend"] != "1500.5" {
		t.Errorf("expected TotalSpend=1500.5, got %v", mock.fieldValues["TotalSpend"])
	}
	if mock.fieldValues["AvgOrder"] != "150.05" {
		t.Errorf("expected AvgOrder=150.05, got %v", mock.fieldValues["AvgOrder"])
	}
	if mock.fieldValues["TotalDue"] != "499.5" {
		t.Errorf("expected TotalDue=499.5 (2000-1500.5), got %v", mock.fieldValues["TotalDue"])
	}
}

func TestCustomerLifetimeValue_Execute_NoInvoices(t *testing.T) {
	mock := &mockConnectorForCLV{
		fieldValues: map[string]interface{}{
			"_related.invoice.count.nonzero": 0,
		},
	}

	output, err := (&CustomerLifetimeValue{}).Execute(context.Background(), helpers.HelperInput{
		ContactID: "contact_123",
		Config: map[string]interface{}{
			"lcv_total_orders": "TotalOrders",
		},
		Connector: mock,
	})

	if err != nil {
		t.Fatal(err)
	}
	if !output.Success {
		t.Error("should succeed")
	}
	if output.Message != "No invoices found for contact" {
		t.Errorf("wrong message: %s", output.Message)
	}
	if len(output.Actions) > 0 {
		t.Error("should not have actions when no invoices")
	}
}

func TestCustomerLifetimeValue_Execute_IncludeZeroYes(t *testing.T) {
	mock := &mockConnectorForCLV{
		fieldValues: map[string]interface{}{
			"_related.invoice.sum.TotalPaid": 500.0,
			"_related.invoice.sum.TotalDue":  500.0,
			"_related.invoice.count":         15.0,
		},
	}

	_, err := (&CustomerLifetimeValue{}).Execute(context.Background(), helpers.HelperInput{
		ContactID: "contact_123",
		Config: map[string]interface{}{
			"lcv_total_orders": "TotalOrders",
			"include_zero":     "Yes",
		},
		Connector: mock,
	})

	if err != nil {
		t.Fatal(err)
	}
	// Should query "_related.invoice.count" instead of "_related.invoice.count.nonzero"
	if mock.fieldValues["TotalOrders"] != "15" {
		t.Errorf("expected TotalOrders=15 (with zero invoices), got %v", mock.fieldValues["TotalOrders"])
	}
}

func TestCustomerLifetimeValue_Execute_IncludeZeroNo(t *testing.T) {
	mock := &mockConnectorForCLV{
		fieldValues: map[string]interface{}{
			"_related.invoice.sum.TotalPaid":    500.0,
			"_related.invoice.sum.TotalDue":     500.0,
			"_related.invoice.count.nonzero":    10.0,
		},
	}

	_, err := (&CustomerLifetimeValue{}).Execute(context.Background(), helpers.HelperInput{
		ContactID: "contact_123",
		Config: map[string]interface{}{
			"lcv_total_orders": "TotalOrders",
			"include_zero":     "No",
		},
		Connector: mock,
	})

	if err != nil {
		t.Fatal(err)
	}
	if mock.fieldValues["TotalOrders"] != "10" {
		t.Errorf("expected TotalOrders=10 (non-zero only), got %v", mock.fieldValues["TotalOrders"])
	}
}

func TestCustomerLifetimeValue_Execute_RoundingBehavior(t *testing.T) {
	mock := &mockConnectorForCLV{
		fieldValues: map[string]interface{}{
			"_related.invoice.sum.TotalPaid":    1234.567,
			"_related.invoice.sum.TotalDue":     2000.123,
			"_related.invoice.count.nonzero":    7.0,
		},
	}

	_, err := (&CustomerLifetimeValue{}).Execute(context.Background(), helpers.HelperInput{
		ContactID: "contact_123",
		Config: map[string]interface{}{
			"lcv_total_spend":   "TotalSpend",
			"lcv_average_order": "AvgOrder",
			"lcv_total_due":     "TotalDue",
		},
		Connector: mock,
	})

	if err != nil {
		t.Fatal(err)
	}

	// Values should be rounded to 2 decimal places
	if mock.fieldValues["TotalSpend"] != "1234.57" {
		t.Errorf("expected rounded TotalSpend=1234.57, got %v", mock.fieldValues["TotalSpend"])
	}
	if mock.fieldValues["AvgOrder"] != "176.37" {
		t.Errorf("expected rounded AvgOrder=176.37, got %v", mock.fieldValues["AvgOrder"])
	}
}

func TestCustomerLifetimeValue_Execute_PartialFieldSelection(t *testing.T) {
	mock := &mockConnectorForCLV{
		fieldValues: map[string]interface{}{
			"_related.invoice.sum.TotalPaid":    1000.0,
			"_related.invoice.sum.TotalDue":     1000.0,
			"_related.invoice.count.nonzero":    5.0,
		},
	}

	output, err := (&CustomerLifetimeValue{}).Execute(context.Background(), helpers.HelperInput{
		ContactID: "contact_123",
		Config: map[string]interface{}{
			"lcv_total_orders": "TotalOrders",
			"lcv_total_spend":  "TotalSpend",
			// Not saving average_order or total_due
		},
		Connector: mock,
	})

	if err != nil {
		t.Fatal(err)
	}
	if len(output.Actions) != 2 {
		t.Errorf("expected 2 actions (only 2 fields configured), got %d", len(output.Actions))
	}
	if _, exists := mock.fieldValues["AvgOrder"]; exists {
		t.Error("should not set AvgOrder (not configured)")
	}
}

func TestCustomerLifetimeValue_Execute_SetFieldError(t *testing.T) {
	mock := &mockConnectorForCLV{
		fieldValues: map[string]interface{}{
			"_related.invoice.sum.TotalPaid":    500.0,
			"_related.invoice.sum.TotalDue":     500.0,
			"_related.invoice.count.nonzero":    5.0,
		},
		setFieldError: fmt.Errorf("field update error"),
	}

	output, err := (&CustomerLifetimeValue{}).Execute(context.Background(), helpers.HelperInput{
		ContactID: "contact_123",
		Config: map[string]interface{}{
			"lcv_total_orders": "TotalOrders",
		},
		Connector: mock,
	})

	if err != nil {
		t.Fatal(err)
	}
	// Should still succeed (logs error)
	if output.Success {
		t.Error("should fail when no fields updated")
	}
	if len(output.Actions) > 0 {
		t.Error("should have no actions on field update error")
	}
}

func TestCustomerLifetimeValue_Execute_ZeroTotalPaidAverageIsZero(t *testing.T) {
	mock := &mockConnectorForCLV{
		fieldValues: map[string]interface{}{
			"_related.invoice.sum.TotalPaid":    0.0,
			"_related.invoice.sum.TotalDue":     1000.0,
			"_related.invoice.count.nonzero":    5.0,
		},
	}

	_, err := (&CustomerLifetimeValue{}).Execute(context.Background(), helpers.HelperInput{
		ContactID: "contact_123",
		Config: map[string]interface{}{
			"lcv_average_order": "AvgOrder",
		},
		Connector: mock,
	})

	if err != nil {
		t.Fatal(err)
	}
	if mock.fieldValues["AvgOrder"] != "0" {
		t.Errorf("expected AvgOrder=0 when total paid is 0, got %v", mock.fieldValues["AvgOrder"])
	}
}

func TestCustomerLifetimeValue_Execute_ModifiedDataStructure(t *testing.T) {
	mock := &mockConnectorForCLV{
		fieldValues: map[string]interface{}{
			"_related.invoice.sum.TotalPaid":    750.0,
			"_related.invoice.sum.TotalDue":     1000.0,
			"_related.invoice.count.nonzero":    3.0,
		},
	}

	output, err := (&CustomerLifetimeValue{}).Execute(context.Background(), helpers.HelperInput{
		ContactID: "contact_123",
		Config: map[string]interface{}{
			"lcv_total_orders": "TotalOrders",
			"include_zero":     "No",
		},
		Connector: mock,
	})

	if err != nil {
		t.Fatal(err)
	}

	modData := output.ModifiedData
	if modData["total_orders"] != 3 {
		t.Error("modified data should include total_orders")
	}
	if modData["total_spend"] != 750.0 {
		t.Error("modified data should include total_spend")
	}
	if modData["average_order"] != 250.0 {
		t.Error("modified data should include average_order")
	}
	if modData["total_owe"] != 250.0 {
		t.Error("modified data should include total_owe")
	}
	if modData["include_zero"] != "No" {
		t.Error("modified data should include include_zero setting")
	}
}
