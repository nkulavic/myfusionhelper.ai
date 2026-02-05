package analytics

import (
	"context"
	"fmt"
	"math"

	"github.com/myfusionhelper/api/internal/helpers"
)

func init() {
	helpers.Register("customer_lifetime_value", func() helpers.Helper { return &CustomerLifetimeValue{} })
}

// CustomerLifetimeValue calculates customer lifetime value metrics from invoice data,
// including total orders, total spend, average order value, and total due.
// Ported from legacy PHP get_clv helper.
type CustomerLifetimeValue struct{}

func (h *CustomerLifetimeValue) GetName() string     { return "Customer Lifetime Value" }
func (h *CustomerLifetimeValue) GetType() string     { return "customer_lifetime_value" }
func (h *CustomerLifetimeValue) GetCategory() string { return "analytics" }
func (h *CustomerLifetimeValue) GetDescription() string {
	return "Calculate customer lifetime value metrics: total orders, total spend, average order, and outstanding balance"
}
func (h *CustomerLifetimeValue) RequiresCRM() bool       { return true }
func (h *CustomerLifetimeValue) SupportedCRMs() []string { return nil }

func (h *CustomerLifetimeValue) GetConfigSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"LCVTotalOrders": map[string]interface{}{
				"type":        "string",
				"description": "Contact field to store total order count",
			},
			"LCVTotalSpend": map[string]interface{}{
				"type":        "string",
				"description": "Contact field to store total spend amount",
			},
			"LCVAverageOrder": map[string]interface{}{
				"type":        "string",
				"description": "Contact field to store average order value",
			},
			"LCVTotalDue": map[string]interface{}{
				"type":        "string",
				"description": "Contact field to store total outstanding balance",
			},
			"IncludeZero": map[string]interface{}{
				"type":        "string",
				"enum":        []string{"Yes", "No"},
				"description": "Whether to include zero-value invoices in calculations",
				"default":     "No",
			},
		},
		"required": []string{},
	}
}

func (h *CustomerLifetimeValue) ValidateConfig(config map[string]interface{}) error {
	// At least one save field should be provided
	fields := []string{"LCVTotalOrders", "LCVTotalSpend", "LCVAverageOrder", "LCVTotalDue"}
	hasField := false
	for _, f := range fields {
		if v, ok := config[f].(string); ok && v != "" && v != "no_save" {
			hasField = true
			break
		}
	}
	if !hasField {
		return fmt.Errorf("at least one save field (LCVTotalOrders, LCVTotalSpend, LCVAverageOrder, LCVTotalDue) is required")
	}
	return nil
}

func (h *CustomerLifetimeValue) Execute(ctx context.Context, input helpers.HelperInput) (*helpers.HelperOutput, error) {
	output := &helpers.HelperOutput{
		Actions: make([]helpers.HelperAction, 0),
		Logs:    make([]string, 0),
	}

	includeZero := "No"
	if iz, ok := input.Config["IncludeZero"].(string); ok {
		includeZero = iz
	}

	// Query invoice data from the CRM via composite keys
	totalPaidVal, _ := input.Connector.GetContactFieldValue(ctx, input.ContactID, "_related.invoice.sum.TotalPaid")
	totalDueVal, _ := input.Connector.GetContactFieldValue(ctx, input.ContactID, "_related.invoice.sum.TotalDue")

	var countKey string
	if includeZero == "Yes" {
		countKey = "_related.invoice.count"
	} else {
		countKey = "_related.invoice.count.nonzero"
	}
	countVal, _ := input.Connector.GetContactFieldValue(ctx, input.ContactID, countKey)

	totalPaid := toFloat64(totalPaidVal)
	totalDue := toFloat64(totalDueVal)
	count := toFloat64(countVal)

	if count == 0 {
		output.Success = true
		output.Message = "No invoices found for contact"
		output.Logs = append(output.Logs, output.Message)
		return output, nil
	}

	totalOwe := totalDue - totalPaid
	average := 0.0
	if totalPaid > 0 {
		average = totalPaid / count
	}

	// Round values
	totalPaid = math.Round(totalPaid*100) / 100
	average = math.Round(average*100) / 100
	totalOwe = math.Round(totalOwe*100) / 100

	// Build update data
	updateData := map[string]interface{}{}
	fieldMappings := map[string]interface{}{
		"LCVTotalOrders":  int(count),
		"LCVTotalSpend":   totalPaid,
		"LCVAverageOrder": average,
		"LCVTotalDue":     totalOwe,
	}

	for configKey, value := range fieldMappings {
		if fieldName, ok := input.Config[configKey].(string); ok && fieldName != "" && fieldName != "no_save" {
			updateData[fieldName] = value
		}
	}

	// Update contact fields
	for field, value := range updateData {
		valueStr := fmt.Sprintf("%v", value)
		err := input.Connector.SetContactFieldValue(ctx, input.ContactID, field, valueStr)
		if err != nil {
			output.Logs = append(output.Logs, fmt.Sprintf("Failed to set field '%s': %v", field, err))
		} else {
			output.Actions = append(output.Actions, helpers.HelperAction{
				Type:   "field_updated",
				Target: field,
				Value:  valueStr,
			})
		}
	}

	output.Success = len(output.Actions) > 0
	output.Message = fmt.Sprintf("CLV calculated: %d orders, $%.2f total spend, $%.2f avg order, $%.2f outstanding",
		int(count), totalPaid, average, totalOwe)
	output.ModifiedData = map[string]interface{}{
		"total_orders":   int(count),
		"total_spend":    totalPaid,
		"average_order":  average,
		"total_owe":      totalOwe,
		"include_zero":   includeZero,
	}
	output.Logs = append(output.Logs, fmt.Sprintf("CLV for contact %s: orders=%d spend=%.2f avg=%.2f owe=%.2f",
		input.ContactID, int(count), totalPaid, average, totalOwe))

	return output, nil
}
