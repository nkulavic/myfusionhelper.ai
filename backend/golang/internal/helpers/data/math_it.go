package data

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strconv"

	"github.com/myfusionhelper/api/internal/helpers"
)

// NewMathIt creates a new MathIt helper instance
func NewMathIt() helpers.Helper { return &MathIt{} }

func init() {
	helpers.Register("math_it", func() helpers.Helper { return &MathIt{} })
}

// MathIt performs math operations on contact field values
type MathIt struct{}

func (h *MathIt) GetName() string        { return "Math It" }
func (h *MathIt) GetType() string        { return "math_it" }
func (h *MathIt) GetCategory() string    { return "data" }
func (h *MathIt) GetDescription() string { return "Perform math operations on contact field values" }
func (h *MathIt) RequiresCRM() bool      { return true }
func (h *MathIt) SupportedCRMs() []string { return nil }

func (h *MathIt) GetConfigSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"field": map[string]interface{}{
				"type":        "string",
				"description": "The field to perform the operation on",
			},
			"operation": map[string]interface{}{
				"type":        "string",
				"enum":        []string{"add", "subtract", "multiply", "divide", "round", "ceil", "floor", "abs", "percent"},
				"description": "The math operation to perform",
			},
			"operand": map[string]interface{}{
				"type":        "number",
				"description": "The value to use in the operation (not needed for round/ceil/floor/abs)",
			},
			"target_field": map[string]interface{}{
				"type":        "string",
				"description": "Optional field to store result (defaults to same field)",
			},
			"decimal_places": map[string]interface{}{
				"type":        "integer",
				"description": "Number of decimal places for rounding (default: 2)",
				"default":     2,
			},
		},
		"required": []string{"field", "operation"},
	}
}

func (h *MathIt) ValidateConfig(config map[string]interface{}) error {
	if _, ok := config["field"].(string); !ok || config["field"] == "" {
		return fmt.Errorf("field is required")
	}

	op, ok := config["operation"].(string)
	if !ok || op == "" {
		return fmt.Errorf("operation is required")
	}

	validOps := map[string]bool{
		"add": true, "subtract": true, "multiply": true, "divide": true,
		"round": true, "ceil": true, "floor": true, "abs": true, "percent": true,
	}
	if !validOps[op] {
		return fmt.Errorf("invalid operation: %s", op)
	}

	// Operations that require an operand
	needsOperand := map[string]bool{"add": true, "subtract": true, "multiply": true, "divide": true, "percent": true}
	if needsOperand[op] {
		if _, ok := config["operand"]; !ok {
			return fmt.Errorf("operand is required for operation '%s'", op)
		}
	}

	return nil
}

func (h *MathIt) Execute(ctx context.Context, input helpers.HelperInput) (*helpers.HelperOutput, error) {
	field := input.Config["field"].(string)
	operation := input.Config["operation"].(string)
	targetField := field
	if tf, ok := input.Config["target_field"].(string); ok && tf != "" {
		targetField = tf
	}
	decimalPlaces := 2
	if dp, ok := input.Config["decimal_places"].(float64); ok {
		decimalPlaces = int(dp)
	}

	output := &helpers.HelperOutput{
		Logs: make([]string, 0),
	}

	// Get current field value
	rawValue, err := input.Connector.GetContactFieldValue(ctx, input.ContactID, field)
	if err != nil {
		output.Message = fmt.Sprintf("Failed to read field '%s': %v", field, err)
		return output, err
	}

	currentValue := toFloat64(rawValue)

	// Get operand
	operand := 0.0
	if op, ok := input.Config["operand"]; ok {
		operand = toFloat64(op)
	}

	// Perform operation
	var result float64
	switch operation {
	case "add":
		result = currentValue + operand
	case "subtract":
		result = currentValue - operand
	case "multiply":
		result = currentValue * operand
	case "divide":
		if operand == 0 {
			output.Message = "Cannot divide by zero"
			return output, fmt.Errorf("division by zero")
		}
		result = currentValue / operand
	case "round":
		factor := math.Pow(10, float64(decimalPlaces))
		result = math.Round(currentValue*factor) / factor
	case "ceil":
		result = math.Ceil(currentValue)
	case "floor":
		result = math.Floor(currentValue)
	case "abs":
		result = math.Abs(currentValue)
	case "percent":
		result = currentValue * (operand / 100)
	}

	// Round result to decimal places
	if operation != "round" {
		factor := math.Pow(10, float64(decimalPlaces))
		result = math.Round(result*factor) / factor
	}

	// Format result - use integer format if no decimals
	var resultStr string
	if result == math.Trunc(result) {
		resultStr = fmt.Sprintf("%.0f", result)
	} else {
		resultStr = strconv.FormatFloat(result, 'f', decimalPlaces, 64)
	}

	// Set result
	err = input.Connector.SetContactFieldValue(ctx, input.ContactID, targetField, resultStr)
	if err != nil {
		output.Message = fmt.Sprintf("Failed to set result: %v", err)
		return output, err
	}

	output.Success = true
	output.Message = fmt.Sprintf("%s %s %v = %s", field, operation, operand, resultStr)
	output.Actions = []helpers.HelperAction{
		{
			Type:   "field_updated",
			Target: targetField,
			Value:  resultStr,
		},
	}
	output.ModifiedData = map[string]interface{}{
		targetField: resultStr,
	}
	output.Logs = append(output.Logs, fmt.Sprintf("Math: %v %s %v = %s (stored in '%s')", currentValue, operation, operand, resultStr, targetField))

	return output, nil
}

func toFloat64(v interface{}) float64 {
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
		f, _ := strconv.ParseFloat(val, 64)
		return f
	case json.Number:
		f, _ := val.Float64()
		return f
	default:
		s := fmt.Sprintf("%v", val)
		f, _ := strconv.ParseFloat(s, 64)
		return f
	}
}
