package analytics

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/myfusionhelper/api/internal/helpers"
)

func init() {
	helpers.Register("rfm_calculation", func() helpers.Helper { return &RFMCalculation{} })
}

// RFMCalculation computes Recency, Frequency, and Monetary scores for a contact
// based on their invoice/order history and configurable thresholds.
// Ported from legacy PHP rfm_calculation helper.
type RFMCalculation struct{}

func (h *RFMCalculation) GetName() string     { return "RFM Calculation" }
func (h *RFMCalculation) GetType() string     { return "rfm_calculation" }
func (h *RFMCalculation) GetCategory() string { return "analytics" }
func (h *RFMCalculation) GetDescription() string {
	return "Calculate Recency, Frequency, and Monetary (RFM) customer metrics from order history"
}
func (h *RFMCalculation) RequiresCRM() bool       { return true }
func (h *RFMCalculation) SupportedCRMs() []string { return nil }

func (h *RFMCalculation) GetConfigSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"options": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"recency_calculation": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"1_2_threshold": map[string]interface{}{"type": "number", "description": "Days threshold for score 2 (e.g. 120)"},
							"2_3_threshold": map[string]interface{}{"type": "number", "description": "Days threshold for score 3 (e.g. 90)"},
							"3_4_threshold": map[string]interface{}{"type": "number", "description": "Days threshold for score 4 (e.g. 60)"},
							"4_5_threshold": map[string]interface{}{"type": "number", "description": "Days threshold for score 5 (e.g. 30)"},
						},
					},
					"frequency_calculation": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"1_2_threshold": map[string]interface{}{"type": "number", "description": "Order count threshold for score 2"},
							"2_3_threshold": map[string]interface{}{"type": "number", "description": "Order count threshold for score 3"},
							"3_4_threshold": map[string]interface{}{"type": "number", "description": "Order count threshold for score 4"},
							"4_5_threshold": map[string]interface{}{"type": "number", "description": "Order count threshold for score 5"},
						},
					},
					"monetary_calculation": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"1_2_threshold": map[string]interface{}{"type": "number", "description": "Total spend threshold for score 2"},
							"2_3_threshold": map[string]interface{}{"type": "number", "description": "Total spend threshold for score 3"},
							"3_4_threshold": map[string]interface{}{"type": "number", "description": "Total spend threshold for score 4"},
							"4_5_threshold": map[string]interface{}{"type": "number", "description": "Total spend threshold for score 5"},
						},
					},
					"save_data": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"recency_score":         map[string]interface{}{"type": "string"},
							"frequency_score":       map[string]interface{}{"type": "string"},
							"monetary_score":        map[string]interface{}{"type": "string"},
							"rfm_composite_score":   map[string]interface{}{"type": "string"},
							"total_order_value":     map[string]interface{}{"type": "string"},
							"average_order_value":   map[string]interface{}{"type": "string"},
							"total_order_count":     map[string]interface{}{"type": "string"},
							"first_order_date":      map[string]interface{}{"type": "string"},
							"first_order_value":     map[string]interface{}{"type": "string"},
							"last_order_date":       map[string]interface{}{"type": "string"},
							"last_order_value":      map[string]interface{}{"type": "string"},
							"days_since_last_order": map[string]interface{}{"type": "string"},
						},
						"description": "Map of metric names to contact field keys for saving results",
					},
				},
				"description": "RFM calculation options including thresholds and save field mappings",
			},
		},
		"required": []string{"options"},
	}
}

func (h *RFMCalculation) ValidateConfig(config map[string]interface{}) error {
	if _, ok := config["options"]; !ok {
		return fmt.Errorf("options is required")
	}
	return nil
}

func (h *RFMCalculation) Execute(ctx context.Context, input helpers.HelperInput) (*helpers.HelperOutput, error) {
	options := extractMap(input.Config["options"])

	output := &helpers.HelperOutput{
		Actions: make([]helpers.HelperAction, 0),
		Logs:    make([]string, 0),
	}

	// Get order/invoice data from the CRM
	// Query for total spend, total orders, etc. via composite keys
	totalSpendVal, _ := input.Connector.GetContactFieldValue(ctx, input.ContactID, "_related.invoice.sum.TotalPaid")
	totalOrdersVal, _ := input.Connector.GetContactFieldValue(ctx, input.ContactID, "_related.invoice.count")
	lastOrderDateVal, _ := input.Connector.GetContactFieldValue(ctx, input.ContactID, "_related.invoice.last.DateCreated")
	lastOrderValueVal, _ := input.Connector.GetContactFieldValue(ctx, input.ContactID, "_related.invoice.last.InvoiceTotal")
	firstOrderDateVal, _ := input.Connector.GetContactFieldValue(ctx, input.ContactID, "_related.invoice.first.DateCreated")
	firstOrderValueVal, _ := input.Connector.GetContactFieldValue(ctx, input.ContactID, "_related.invoice.first.InvoiceTotal")
	totalInvoiceVal, _ := input.Connector.GetContactFieldValue(ctx, input.ContactID, "_related.invoice.sum.InvoiceTotal")

	totalSpend := toFloat64(totalSpendVal)
	totalOrders := toFloat64(totalOrdersVal)
	lastOrderDate := fmt.Sprintf("%v", lastOrderDateVal)
	lastOrderValue := toFloat64(lastOrderValueVal)
	firstOrderDate := fmt.Sprintf("%v", firstOrderDateVal)
	firstOrderValue := toFloat64(firstOrderValueVal)
	totalInvoiceAmount := toFloat64(totalInvoiceVal)

	if totalOrders == 0 {
		output.Success = true
		output.Message = "No orders found for contact"
		output.Logs = append(output.Logs, output.Message)
		return output, nil
	}

	averageOrderSize := math.Round(totalInvoiceAmount / totalOrders)
	actualPaymentPerOrder := math.Round(totalSpend / totalOrders)
	totalSpendRounded := math.Round(totalSpend)
	totalInvoiceRounded := math.Round(totalInvoiceAmount)

	// Calculate days since last order
	daysSinceLastOrder := 0
	if lastOrderDate != "" && lastOrderDate != "<nil>" {
		parsedDate, err := time.Parse("2006-01-02", lastOrderDate[:minInt(10, len(lastOrderDate))])
		if err == nil {
			daysSinceLastOrder = int(time.Since(parsedDate).Hours() / 24)
		}
	}

	// Calculate Recency score (1-5)
	recencyCalc := extractMap(options["recency_calculation"])
	recencyScore := 1
	if daysSinceLastOrder <= int(toFloat64(recencyCalc["1_2_threshold"])) {
		recencyScore = 2
	}
	if daysSinceLastOrder <= int(toFloat64(recencyCalc["2_3_threshold"])) {
		recencyScore = 3
	}
	if daysSinceLastOrder <= int(toFloat64(recencyCalc["3_4_threshold"])) {
		recencyScore = 4
	}
	if daysSinceLastOrder <= int(toFloat64(recencyCalc["4_5_threshold"])) {
		recencyScore = 5
	}

	// Calculate Frequency score (1-5)
	frequencyCalc := extractMap(options["frequency_calculation"])
	frequencyScore := 1
	if totalOrders >= toFloat64(frequencyCalc["1_2_threshold"]) {
		frequencyScore = 2
	}
	if totalOrders >= toFloat64(frequencyCalc["2_3_threshold"]) {
		frequencyScore = 3
	}
	if totalOrders >= toFloat64(frequencyCalc["3_4_threshold"]) {
		frequencyScore = 4
	}
	if totalOrders >= toFloat64(frequencyCalc["4_5_threshold"]) {
		frequencyScore = 5
	}

	// Calculate Monetary score (1-5)
	monetaryCalc := extractMap(options["monetary_calculation"])
	monetaryScore := 1
	if totalSpendRounded >= toFloat64(monetaryCalc["1_2_threshold"]) {
		monetaryScore = 2
	}
	if totalSpendRounded >= toFloat64(monetaryCalc["2_3_threshold"]) {
		monetaryScore = 3
	}
	if totalSpendRounded >= toFloat64(monetaryCalc["3_4_threshold"]) {
		monetaryScore = 4
	}
	if totalSpendRounded >= toFloat64(monetaryCalc["4_5_threshold"]) {
		monetaryScore = 5
	}

	// Composite score (concatenated digits, e.g. 543)
	compositeScore := recencyScore*100 + frequencyScore*10 + monetaryScore

	// Build update data based on save_data field mappings
	saveData := extractMap(options["save_data"])
	updateData := map[string]interface{}{}

	fieldMappings := map[string]interface{}{
		"recency_score":         recencyScore,
		"frequency_score":       frequencyScore,
		"monetary_score":        monetaryScore,
		"rfm_composite_score":   compositeScore,
		"total_order_value":     totalSpendRounded,
		"average_order_value":   averageOrderSize,
		"total_order_count":     int(totalOrders),
		"first_order_date":      firstOrderDate,
		"first_order_value":     firstOrderValue,
		"last_order_date":       lastOrderDate,
		"last_order_value":      lastOrderValue,
		"days_since_last_order": daysSinceLastOrder,
	}

	for metricKey, metricValue := range fieldMappings {
		if fieldName, ok := saveData[metricKey].(string); ok && fieldName != "" && fieldName != "no_save" {
			valueStr := fmt.Sprintf("%v", metricValue)
			updateData[fieldName] = valueStr
		}
	}

	// Update contact fields
	for field, value := range updateData {
		err := input.Connector.SetContactFieldValue(ctx, input.ContactID, field, value)
		if err != nil {
			output.Logs = append(output.Logs, fmt.Sprintf("Failed to set field '%s': %v", field, err))
		} else {
			output.Actions = append(output.Actions, helpers.HelperAction{
				Type:   "field_updated",
				Target: field,
				Value:  value,
			})
		}
	}

	// Fire RFM score goals
	integration := "myfusionhelper"
	helperID := input.HelperID

	goalNames := []string{
		fmt.Sprintf("%srecency%d", helperID, recencyScore),
		fmt.Sprintf("%sfrequency%d", helperID, frequencyScore),
		fmt.Sprintf("%smonetary%d", helperID, monetaryScore),
	}

	for _, goalName := range goalNames {
		goalErr := input.Connector.AchieveGoal(ctx, input.ContactID, goalName, integration)
		if goalErr != nil {
			output.Logs = append(output.Logs, fmt.Sprintf("Failed to achieve goal '%s': %v", goalName, goalErr))
		} else {
			output.Actions = append(output.Actions, helpers.HelperAction{
				Type:   "goal_achieved",
				Target: input.ContactID,
				Value:  goalName,
			})
		}
	}

	output.Success = true
	output.Message = fmt.Sprintf("RFM scores calculated: R=%d F=%d M=%d (composite: %d)", recencyScore, frequencyScore, monetaryScore, compositeScore)
	output.ModifiedData = map[string]interface{}{
		"recency_score":         recencyScore,
		"frequency_score":       frequencyScore,
		"monetary_score":        monetaryScore,
		"composite_score":       compositeScore,
		"total_spend":           totalSpendRounded,
		"total_invoice_amount":  totalInvoiceRounded,
		"total_orders":          int(totalOrders),
		"average_order_size":    averageOrderSize,
		"payment_per_order":     actualPaymentPerOrder,
		"days_since_last_order": daysSinceLastOrder,
		"first_order_date":      firstOrderDate,
		"first_order_value":     firstOrderValue,
		"last_order_date":       lastOrderDate,
		"last_order_value":      lastOrderValue,
	}
	output.Logs = append(output.Logs, fmt.Sprintf("RFM for contact %s: R=%d F=%d M=%d composite=%d", input.ContactID, recencyScore, frequencyScore, monetaryScore, compositeScore))

	return output, nil
}

func extractMap(v interface{}) map[string]interface{} {
	if m, ok := v.(map[string]interface{}); ok {
		return m
	}
	return map[string]interface{}{}
}

func toFloat64(v interface{}) float64 {
	if v == nil {
		return 0
	}
	switch val := v.(type) {
	case float64:
		return val
	case float32:
		return float64(val)
	case int:
		return float64(val)
	case int64:
		return float64(val)
	case string:
		var f float64
		_, _ = fmt.Sscanf(val, "%f", &f)
		return f
	default:
		var f float64
		_, _ = fmt.Sscanf(fmt.Sprintf("%v", v), "%f", &f)
		return f
	}
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
