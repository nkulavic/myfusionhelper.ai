package data

import (
	"context"
	"fmt"
	"time"

	"github.com/myfusionhelper/api/internal/helpers"
)

func init() {
	helpers.Register("date_calc", func() helpers.Helper { return &DateCalc{} })
}

// DateCalc performs date calculations on contact field values
type DateCalc struct{}

func (h *DateCalc) GetName() string        { return "Date Calc" }
func (h *DateCalc) GetType() string        { return "date_calc" }
func (h *DateCalc) GetCategory() string    { return "data" }
func (h *DateCalc) GetDescription() string { return "Perform date calculations on contact field values" }
func (h *DateCalc) RequiresCRM() bool      { return true }
func (h *DateCalc) SupportedCRMs() []string { return nil }

func (h *DateCalc) GetConfigSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"field": map[string]interface{}{
				"type":        "string",
				"description": "The date field to operate on",
			},
			"operation": map[string]interface{}{
				"type":        "string",
				"enum":        []string{"add_days", "subtract_days", "add_months", "subtract_months", "add_years", "subtract_years", "set_now", "diff_days", "format"},
				"description": "The date operation to perform",
			},
			"amount": map[string]interface{}{
				"type":        "integer",
				"description": "Number of days/months/years (for add/subtract operations)",
			},
			"target_field": map[string]interface{}{
				"type":        "string",
				"description": "Field to store the result (defaults to same field)",
			},
			"compare_field": map[string]interface{}{
				"type":        "string",
				"description": "Second date field for diff_days operation",
			},
			"output_format": map[string]interface{}{
				"type":        "string",
				"description": "Output date format (Go time format string, default: 2006-01-02)",
				"default":     "2006-01-02",
			},
		},
		"required": []string{"operation"},
	}
}

func (h *DateCalc) ValidateConfig(config map[string]interface{}) error {
	op, ok := config["operation"].(string)
	if !ok || op == "" {
		return fmt.Errorf("operation is required")
	}

	validOps := map[string]bool{
		"add_days": true, "subtract_days": true,
		"add_months": true, "subtract_months": true,
		"add_years": true, "subtract_years": true,
		"set_now": true, "diff_days": true, "format": true,
	}
	if !validOps[op] {
		return fmt.Errorf("invalid operation: %s", op)
	}

	needsField := map[string]bool{
		"add_days": true, "subtract_days": true,
		"add_months": true, "subtract_months": true,
		"add_years": true, "subtract_years": true,
		"diff_days": true, "format": true,
	}
	if needsField[op] {
		if _, ok := config["field"].(string); !ok || config["field"] == "" {
			return fmt.Errorf("field is required for operation '%s'", op)
		}
	}

	needsAmount := map[string]bool{
		"add_days": true, "subtract_days": true,
		"add_months": true, "subtract_months": true,
		"add_years": true, "subtract_years": true,
	}
	if needsAmount[op] {
		if _, ok := config["amount"]; !ok {
			return fmt.Errorf("amount is required for operation '%s'", op)
		}
	}

	if op == "diff_days" {
		if _, ok := config["compare_field"].(string); !ok || config["compare_field"] == "" {
			return fmt.Errorf("compare_field is required for diff_days operation")
		}
	}

	return nil
}

func (h *DateCalc) Execute(ctx context.Context, input helpers.HelperInput) (*helpers.HelperOutput, error) {
	operation := input.Config["operation"].(string)
	targetField := ""
	if f, ok := input.Config["field"].(string); ok {
		targetField = f
	}
	if tf, ok := input.Config["target_field"].(string); ok && tf != "" {
		targetField = tf
	}
	outputFormat := "2006-01-02"
	if f, ok := input.Config["output_format"].(string); ok && f != "" {
		outputFormat = f
	}

	output := &helpers.HelperOutput{
		Logs: make([]string, 0),
	}

	// Handle set_now (doesn't need to read a field)
	if operation == "set_now" {
		if targetField == "" {
			if f, ok := input.Config["target_field"].(string); ok {
				targetField = f
			}
			if targetField == "" {
				output.Message = "target_field is required for set_now operation"
				return output, fmt.Errorf("target_field required")
			}
		}

		now := time.Now().UTC().Format(outputFormat)
		err := input.Connector.SetContactFieldValue(ctx, input.ContactID, targetField, now)
		if err != nil {
			output.Message = fmt.Sprintf("Failed to set date: %v", err)
			return output, err
		}

		output.Success = true
		output.Message = fmt.Sprintf("Set '%s' to current date: %s", targetField, now)
		output.Actions = []helpers.HelperAction{{Type: "field_updated", Target: targetField, Value: now}}
		output.ModifiedData = map[string]interface{}{targetField: now}
		return output, nil
	}

	// Read date field
	field := input.Config["field"].(string)
	rawValue, err := input.Connector.GetContactFieldValue(ctx, input.ContactID, field)
	if err != nil {
		output.Message = fmt.Sprintf("Failed to read field '%s': %v", field, err)
		return output, err
	}

	dateStr := fmt.Sprintf("%v", rawValue)
	parsedDate, err := parseFlexibleDate(dateStr)
	if err != nil {
		output.Message = fmt.Sprintf("Could not parse date '%s': %v", dateStr, err)
		return output, err
	}

	var resultStr string

	switch operation {
	case "add_days":
		amount := int(toFloat64(input.Config["amount"]))
		result := parsedDate.AddDate(0, 0, amount)
		resultStr = result.Format(outputFormat)

	case "subtract_days":
		amount := int(toFloat64(input.Config["amount"]))
		result := parsedDate.AddDate(0, 0, -amount)
		resultStr = result.Format(outputFormat)

	case "add_months":
		amount := int(toFloat64(input.Config["amount"]))
		result := parsedDate.AddDate(0, amount, 0)
		resultStr = result.Format(outputFormat)

	case "subtract_months":
		amount := int(toFloat64(input.Config["amount"]))
		result := parsedDate.AddDate(0, -amount, 0)
		resultStr = result.Format(outputFormat)

	case "add_years":
		amount := int(toFloat64(input.Config["amount"]))
		result := parsedDate.AddDate(amount, 0, 0)
		resultStr = result.Format(outputFormat)

	case "subtract_years":
		amount := int(toFloat64(input.Config["amount"]))
		result := parsedDate.AddDate(-amount, 0, 0)
		resultStr = result.Format(outputFormat)

	case "diff_days":
		compareField := input.Config["compare_field"].(string)
		compareRaw, err := input.Connector.GetContactFieldValue(ctx, input.ContactID, compareField)
		if err != nil {
			output.Message = fmt.Sprintf("Failed to read compare field '%s': %v", compareField, err)
			return output, err
		}

		compareDate, err := parseFlexibleDate(fmt.Sprintf("%v", compareRaw))
		if err != nil {
			output.Message = fmt.Sprintf("Could not parse compare date: %v", err)
			return output, err
		}

		diff := int(compareDate.Sub(parsedDate).Hours() / 24)
		resultStr = fmt.Sprintf("%d", diff)

	case "format":
		resultStr = parsedDate.Format(outputFormat)
	}

	// Set target field if not diff_days, or if target specified
	if targetField == "" {
		targetField = field
	}

	err = input.Connector.SetContactFieldValue(ctx, input.ContactID, targetField, resultStr)
	if err != nil {
		output.Message = fmt.Sprintf("Failed to set result: %v", err)
		return output, err
	}

	output.Success = true
	output.Message = fmt.Sprintf("Date %s: result = %s", operation, resultStr)
	output.Actions = []helpers.HelperAction{{Type: "field_updated", Target: targetField, Value: resultStr}}
	output.ModifiedData = map[string]interface{}{targetField: resultStr}
	output.Logs = append(output.Logs, fmt.Sprintf("Date calc '%s' on contact %s: %s -> %s", operation, input.ContactID, dateStr, resultStr))

	return output, nil
}

func parseFlexibleDate(s string) (time.Time, error) {
	formats := []string{
		time.RFC3339,
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
		"2006-01-02",
		"01/02/2006",
		"01-02-2006",
		"Jan 2, 2006",
		"January 2, 2006",
		"02 Jan 2006",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, s); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("unrecognized date format: %s", s)
}
