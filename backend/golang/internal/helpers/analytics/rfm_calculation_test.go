package analytics

import (
	"context"
	"fmt"
	"testing"

	"github.com/myfusionhelper/api/internal/connectors"
	"github.com/myfusionhelper/api/internal/helpers"
)

type mockConnectorForRFM struct {
	fieldValues    map[string]interface{}
	getFieldError  map[string]error
	setFieldError  error
	goalsAchieved  []string
	achieveGoalErr error
}

func (m *mockConnectorForRFM) GetContactFieldValue(ctx context.Context, contactID, fieldKey string) (interface{}, error) {
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

func (m *mockConnectorForRFM) SetContactFieldValue(ctx context.Context, contactID, fieldKey string, value interface{}) error {
	if m.setFieldError != nil {
		return m.setFieldError
	}
	if m.fieldValues == nil {
		m.fieldValues = make(map[string]interface{})
	}
	m.fieldValues[fieldKey] = value
	return nil
}

func (m *mockConnectorForRFM) AchieveGoal(ctx context.Context, contactID, goalName, integration string) error {
	if m.achieveGoalErr != nil {
		return m.achieveGoalErr
	}
	if m.goalsAchieved == nil {
		m.goalsAchieved = make([]string, 0)
	}
	m.goalsAchieved = append(m.goalsAchieved, goalName)
	return nil
}

func (m *mockConnectorForRFM) GetContacts(ctx context.Context, opts connectors.QueryOptions) (*connectors.ContactList, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForRFM) GetContact(ctx context.Context, contactID string) (*connectors.NormalizedContact, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForRFM) CreateContact(ctx context.Context, contact connectors.CreateContactInput) (*connectors.NormalizedContact, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForRFM) UpdateContact(ctx context.Context, contactID string, updates connectors.UpdateContactInput) (*connectors.NormalizedContact, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForRFM) DeleteContact(ctx context.Context, contactID string) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorForRFM) GetTags(ctx context.Context) ([]connectors.Tag, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForRFM) ApplyTag(ctx context.Context, contactID, tagID string) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorForRFM) RemoveTag(ctx context.Context, contactID, tagID string) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorForRFM) GetCustomFields(ctx context.Context) ([]connectors.CustomField, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForRFM) TriggerAutomation(ctx context.Context, contactID, automationID string) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorForRFM) TestConnection(ctx context.Context) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorForRFM) GetMetadata() connectors.ConnectorMetadata {
	return connectors.ConnectorMetadata{}
}
func (m *mockConnectorForRFM) GetCapabilities() []connectors.Capability {
	return nil
}
func (m *mockConnectorForRFM) SetOptInStatus(ctx context.Context, contactID string, optIn bool, reason string) error {
	return nil
}

func TestRFMCalculation_GetMetadata(t *testing.T) {
	h := &RFMCalculation{}
	if h.GetName() != "RFM Calculation" {
		t.Error("wrong name")
	}
	if h.GetType() != "rfm_calculation" {
		t.Error("wrong type")
	}
	if h.GetCategory() != "analytics" {
		t.Error("wrong category")
	}
	if !h.RequiresCRM() {
		t.Error("should require CRM")
	}
}

func TestRFMCalculation_ValidateConfig_MissingOptions(t *testing.T) {
	err := (&RFMCalculation{}).ValidateConfig(map[string]interface{}{})
	if err == nil {
		t.Error("should error on missing options")
	}
}

func TestRFMCalculation_ValidateConfig_Valid(t *testing.T) {
	err := (&RFMCalculation{}).ValidateConfig(map[string]interface{}{
		"options": map[string]interface{}{},
	})
	if err != nil {
		t.Errorf("should be valid: %v", err)
	}
}

func TestRFMCalculation_Execute_NoOrders(t *testing.T) {
	mock := &mockConnectorForRFM{
		fieldValues: map[string]interface{}{
			"_related.invoice.count": 0.0,
		},
	}

	output, err := (&RFMCalculation{}).Execute(context.Background(), helpers.HelperInput{
		ContactID: "contact_123",
		Config: map[string]interface{}{
			"options": map[string]interface{}{},
		},
		Connector: mock,
	})

	if err != nil {
		t.Fatal(err)
	}
	if !output.Success {
		t.Error("should succeed")
	}
	if output.Message != "No orders found for contact" {
		t.Errorf("wrong message: %s", output.Message)
	}
}

func TestRFMCalculation_Execute_RecencyScoring(t *testing.T) {
	mock := &mockConnectorForRFM{
		fieldValues: map[string]interface{}{
			"_related.invoice.sum.TotalPaid":            1000.0,
			"_related.invoice.count":                    5.0,
			"_related.invoice.last.DateCreated":         "2026-02-01",
			"_related.invoice.last.InvoiceTotal":        200.0,
			"_related.invoice.first.DateCreated":        "2025-12-01",
			"_related.invoice.first.InvoiceTotal":       100.0,
			"_related.invoice.sum.InvoiceTotal":         1200.0,
		},
	}

	output, err := (&RFMCalculation{}).Execute(context.Background(), helpers.HelperInput{
		ContactID: "contact_123",
		HelperID:  "helper_123",
		Config: map[string]interface{}{
			"options": map[string]interface{}{
				"recency_calculation": map[string]interface{}{
					"1_2_threshold": 120.0,
					"2_3_threshold": 90.0,
					"3_4_threshold": 60.0,
					"4_5_threshold": 30.0,
				},
				"frequency_calculation": map[string]interface{}{
					"1_2_threshold": 2.0,
					"2_3_threshold": 5.0,
					"3_4_threshold": 10.0,
					"4_5_threshold": 20.0,
				},
				"monetary_calculation": map[string]interface{}{
					"1_2_threshold": 100.0,
					"2_3_threshold": 500.0,
					"3_4_threshold": 1000.0,
					"4_5_threshold": 5000.0,
				},
				"save_data": map[string]interface{}{
					"recency_score":  "RecencyScore",
					"frequency_score": "FrequencyScore",
					"monetary_score":  "MonetaryScore",
				},
			},
		},
		Connector: mock,
	})

	if err != nil {
		t.Fatal(err)
	}
	if !output.Success {
		t.Error("should succeed")
	}

	// Recent order (7 days) = score 5
	// 5 orders = score 3
	// $1000 spent = score 4 (>= 1000 threshold)
	modData := output.ModifiedData
	if modData["recency_score"] != 5 {
		t.Errorf("expected recency_score=5, got %v", modData["recency_score"])
	}
	if modData["frequency_score"] != 3 {
		t.Errorf("expected frequency_score=3, got %v", modData["frequency_score"])
	}
	if modData["monetary_score"] != 4 {
		t.Errorf("expected monetary_score=4, got %v", modData["monetary_score"])
	}
	if modData["composite_score"] != 534 {
		t.Errorf("expected composite=534, got %v", modData["composite_score"])
	}
}

func TestRFMCalculation_Execute_GoalsAchieved(t *testing.T) {
	mock := &mockConnectorForRFM{
		fieldValues: map[string]interface{}{
			"_related.invoice.sum.TotalPaid":            500.0,
			"_related.invoice.count":                    3.0,
			"_related.invoice.last.DateCreated":         "2026-02-01",
			"_related.invoice.sum.InvoiceTotal":         600.0,
		},
	}

	output, err := (&RFMCalculation{}).Execute(context.Background(), helpers.HelperInput{
		ContactID: "contact_123",
		HelperID:  "helper_xyz",
		Config: map[string]interface{}{
			"options": map[string]interface{}{
				"recency_calculation": map[string]interface{}{
					"4_5_threshold": 30.0,
				},
				"frequency_calculation": map[string]interface{}{
					"1_2_threshold": 2.0,
					"2_3_threshold": 5.0,
				},
				"monetary_calculation": map[string]interface{}{
					"1_2_threshold": 100.0,
					"2_3_threshold": 500.0,
				},
				"save_data": map[string]interface{}{},
			},
		},
		Connector: mock,
	})

	if err != nil {
		t.Fatal(err)
	}
	if !output.Success {
		t.Error("should succeed")
	}

	// Should fire 3 goals: recency5, frequency5 (3 >= threshold 2), monetary5 (500 >= threshold 500)
	if len(mock.goalsAchieved) != 3 {
		t.Fatalf("expected 3 goals, got %d: %v", len(mock.goalsAchieved), mock.goalsAchieved)
	}
	expectedGoals := map[string]bool{
		"helper_xyzrecency5":   true,
		"helper_xyzfrequency5": true,
		"helper_xyzmonetary5":  true,
	}
	for _, goal := range mock.goalsAchieved {
		if !expectedGoals[goal] {
			t.Errorf("unexpected goal: %s", goal)
		}
	}
}

func TestRFMCalculation_Execute_SaveAllFields(t *testing.T) {
	mock := &mockConnectorForRFM{
		fieldValues: map[string]interface{}{
			"_related.invoice.sum.TotalPaid":            750.0,
			"_related.invoice.count":                    4.0,
			"_related.invoice.last.DateCreated":         "2026-02-05",
			"_related.invoice.last.InvoiceTotal":        200.0,
			"_related.invoice.first.DateCreated":        "2026-01-01",
			"_related.invoice.first.InvoiceTotal":       150.0,
			"_related.invoice.sum.InvoiceTotal":         800.0,
		},
	}

	_, err := (&RFMCalculation{}).Execute(context.Background(), helpers.HelperInput{
		ContactID: "contact_123",
		HelperID:  "helper_123",
		Config: map[string]interface{}{
			"options": map[string]interface{}{
				"recency_calculation":   map[string]interface{}{},
				"frequency_calculation": map[string]interface{}{},
				"monetary_calculation":  map[string]interface{}{},
				"save_data": map[string]interface{}{
					"total_order_value":     "TotalOrderValue",
					"average_order_value":   "AvgOrderValue",
					"total_order_count":     "TotalOrders",
					"first_order_date":      "FirstOrderDate",
					"first_order_value":     "FirstOrderValue",
					"last_order_date":       "LastOrderDate",
					"last_order_value":      "LastOrderValue",
					"days_since_last_order": "DaysSinceLastOrder",
				},
			},
		},
		Connector: mock,
	})

	if err != nil {
		t.Fatal(err)
	}

	// Verify all fields were saved
	if mock.fieldValues["TotalOrderValue"] != "750" {
		t.Error("TotalOrderValue should be saved")
	}
	if mock.fieldValues["AvgOrderValue"] != "200" {
		t.Error("AvgOrderValue should be saved")
	}
	if mock.fieldValues["TotalOrders"] != "4" {
		t.Error("TotalOrders should be saved")
	}
	if mock.fieldValues["FirstOrderDate"] != "2026-01-01" {
		t.Error("FirstOrderDate should be saved")
	}
	if mock.fieldValues["LastOrderDate"] != "2026-02-05" {
		t.Error("LastOrderDate should be saved")
	}
}

func TestRFMCalculation_Execute_CompositeScore(t *testing.T) {
	mock := &mockConnectorForRFM{
		fieldValues: map[string]interface{}{
			"_related.invoice.sum.TotalPaid":    10000.0,
			"_related.invoice.count":            25.0,
			"_related.invoice.last.DateCreated": "2026-02-07",
			"_related.invoice.sum.InvoiceTotal": 10000.0,
		},
	}

	output, err := (&RFMCalculation{}).Execute(context.Background(), helpers.HelperInput{
		ContactID: "contact_123",
		HelperID:  "helper_123",
		Config: map[string]interface{}{
			"options": map[string]interface{}{
				"recency_calculation": map[string]interface{}{
					"4_5_threshold": 30.0,
				},
				"frequency_calculation": map[string]interface{}{
					"4_5_threshold": 20.0,
				},
				"monetary_calculation": map[string]interface{}{
					"4_5_threshold": 5000.0,
				},
				"save_data": map[string]interface{}{
					"rfm_composite_score": "CompositeScore",
				},
			},
		},
		Connector: mock,
	})

	if err != nil {
		t.Fatal(err)
	}

	// R=5, F=5, M=5 -> composite = 555
	if mock.fieldValues["CompositeScore"] != "555" {
		t.Errorf("expected composite score 555, got %v", mock.fieldValues["CompositeScore"])
	}
	if output.ModifiedData["composite_score"] != 555 {
		t.Error("modified data should include composite score")
	}
}

func TestRFMCalculation_Execute_SetFieldError(t *testing.T) {
	mock := &mockConnectorForRFM{
		fieldValues: map[string]interface{}{
			"_related.invoice.sum.TotalPaid":    500.0,
			"_related.invoice.count":            2.0,
			"_related.invoice.last.DateCreated": "2026-02-01",
			"_related.invoice.sum.InvoiceTotal": 500.0,
		},
		setFieldError: fmt.Errorf("field update error"),
	}

	output, err := (&RFMCalculation{}).Execute(context.Background(), helpers.HelperInput{
		ContactID: "contact_123",
		HelperID:  "helper_123",
		Config: map[string]interface{}{
			"options": map[string]interface{}{
				"recency_calculation":   map[string]interface{}{},
				"frequency_calculation": map[string]interface{}{},
				"monetary_calculation":  map[string]interface{}{},
				"save_data": map[string]interface{}{
					"recency_score": "RecencyScore",
				},
			},
		},
		Connector: mock,
	})

	if err != nil {
		t.Fatal(err)
	}
	// Should still succeed (logs errors)
	if !output.Success {
		t.Error("should succeed even with field update errors")
	}
}

func TestRFMCalculation_Execute_AchieveGoalError(t *testing.T) {
	mock := &mockConnectorForRFM{
		fieldValues: map[string]interface{}{
			"_related.invoice.sum.TotalPaid":    500.0,
			"_related.invoice.count":            2.0,
			"_related.invoice.last.DateCreated": "2026-02-01",
			"_related.invoice.sum.InvoiceTotal": 500.0,
		},
		achieveGoalErr: fmt.Errorf("goal error"),
	}

	output, err := (&RFMCalculation{}).Execute(context.Background(), helpers.HelperInput{
		ContactID: "contact_123",
		HelperID:  "helper_123",
		Config: map[string]interface{}{
			"options": map[string]interface{}{
				"recency_calculation":   map[string]interface{}{},
				"frequency_calculation": map[string]interface{}{},
				"monetary_calculation":  map[string]interface{}{},
				"save_data":             map[string]interface{}{},
			},
		},
		Connector: mock,
	})

	if err != nil {
		t.Fatal(err)
	}
	// Should still succeed (logs goal errors)
	if !output.Success {
		t.Error("should succeed even with goal errors")
	}
}

func TestRFMCalculation_Execute_DaysSinceLastOrder(t *testing.T) {
	mock := &mockConnectorForRFM{
		fieldValues: map[string]interface{}{
			"_related.invoice.sum.TotalPaid":    500.0,
			"_related.invoice.count":            2.0,
			"_related.invoice.last.DateCreated": "2025-12-01",
			"_related.invoice.sum.InvoiceTotal": 500.0,
		},
	}

	output, err := (&RFMCalculation{}).Execute(context.Background(), helpers.HelperInput{
		ContactID: "contact_123",
		HelperID:  "helper_123",
		Config: map[string]interface{}{
			"options": map[string]interface{}{
				"recency_calculation":   map[string]interface{}{},
				"frequency_calculation": map[string]interface{}{},
				"monetary_calculation":  map[string]interface{}{},
				"save_data":             map[string]interface{}{},
			},
		},
		Connector: mock,
	})

	if err != nil {
		t.Fatal(err)
	}

	daysSince := output.ModifiedData["days_since_last_order"].(int)
	if daysSince < 60 || daysSince > 75 {
		t.Errorf("expected days since last order ~69 (2025-12-01 to 2026-02-08), got %d", daysSince)
	}
}

func TestRFMCalculation_Execute_NoSaveFieldsConfigured(t *testing.T) {
	mock := &mockConnectorForRFM{
		fieldValues: map[string]interface{}{
			"_related.invoice.sum.TotalPaid":    500.0,
			"_related.invoice.count":            2.0,
			"_related.invoice.last.DateCreated": "2026-02-01",
			"_related.invoice.sum.InvoiceTotal": 500.0,
		},
	}

	output, err := (&RFMCalculation{}).Execute(context.Background(), helpers.HelperInput{
		ContactID: "contact_123",
		HelperID:  "helper_123",
		Config: map[string]interface{}{
			"options": map[string]interface{}{
				"recency_calculation":   map[string]interface{}{},
				"frequency_calculation": map[string]interface{}{},
				"monetary_calculation":  map[string]interface{}{},
				"save_data":             map[string]interface{}{},
			},
		},
		Connector: mock,
	})

	if err != nil {
		t.Fatal(err)
	}
	if !output.Success {
		t.Error("should succeed")
	}
	// Should only have goal actions, no field updates
	fieldUpdateCount := 0
	for _, action := range output.Actions {
		if action.Type == "field_updated" {
			fieldUpdateCount++
		}
	}
	if fieldUpdateCount > 0 {
		t.Error("should not have field updates when no save fields configured")
	}
}

func TestRFMCalculation_Execute_FrequencyScoreLevels(t *testing.T) {
	testCases := []struct {
		orderCount     float64
		expectedScore  int
	}{
		{1.0, 1},
		{2.0, 2},
		{5.0, 3},
		{10.0, 4},
		{20.0, 5},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Orders_%v", tc.orderCount), func(t *testing.T) {
			mock := &mockConnectorForRFM{
				fieldValues: map[string]interface{}{
					"_related.invoice.sum.TotalPaid":    500.0,
					"_related.invoice.count":            tc.orderCount,
					"_related.invoice.last.DateCreated": "2026-02-01",
					"_related.invoice.sum.InvoiceTotal": 500.0,
				},
			}

			output, err := (&RFMCalculation{}).Execute(context.Background(), helpers.HelperInput{
				ContactID: "contact_123",
				HelperID:  "helper_123",
				Config: map[string]interface{}{
					"options": map[string]interface{}{
						"recency_calculation": map[string]interface{}{},
						"frequency_calculation": map[string]interface{}{
							"1_2_threshold": 2.0,
							"2_3_threshold": 5.0,
							"3_4_threshold": 10.0,
							"4_5_threshold": 20.0,
						},
						"monetary_calculation": map[string]interface{}{},
						"save_data":            map[string]interface{}{},
					},
				},
				Connector: mock,
			})

			if err != nil {
				t.Fatal(err)
			}

			if output.ModifiedData["frequency_score"] != tc.expectedScore {
				t.Errorf("expected frequency_score=%d for %v orders, got %v",
					tc.expectedScore, tc.orderCount, output.ModifiedData["frequency_score"])
			}
		})
	}
}
