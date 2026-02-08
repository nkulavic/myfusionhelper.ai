package data

import (
	"context"
	"fmt"
	"math"

	"github.com/myfusionhelper/api/internal/helpers"
)

func init() {
	helpers.Register("advance_math", func() helpers.Helper { return &AdvanceMath{} })
}

// AdvanceMath performs advanced mathematical operations on contact field values
type AdvanceMath struct{}

func (h *AdvanceMath) GetName() string     { return "Advance Math" }
func (h *AdvanceMath) GetType() string     { return "advance_math" }
func (h *AdvanceMath) GetCategory() string { return "data" }
func (h *AdvanceMath) GetDescription() string {
	return "Perform advanced mathematical operations (power, square root, absolute value, rounding, min, max) on contact field values"
}
func (h *AdvanceMath) RequiresCRM() bool       { return true }
func (h *AdvanceMath) SupportedCRMs() []string { return nil }

func (h *AdvanceMath) GetConfigSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"operation": map[string]interface{}{
				"type": "string",
				"enum": []string{"power", "sqrt", "abs", "round", "ceil", "floor", "min", "max"},
				"description": "Mathematical operation to perform",
			},
			"source_field": map[string]interface{}{
				"type":        "string",
				"description": "Source field containing the numeric value",
			},
			"operand": map[string]interface{}{
				"type":        "number",
				"description": "Operand for operations that require a second value (power, min, max)",
			},
			"second_field": map[string]interface{}{
				"type":        "string",
				"description": "Second source field for min/max operations (alternative to operand)",
			},
			"target_field": map[string]interface{}{
				"type":        "string",
				"description": "Target field to store the result",
			},
		},
		"required": []string{"operation", "source_field", "target_field"},
	}
}

func (h *AdvanceMath) ValidateConfig(config map[string]interface{}) error {
	operation, ok := config["operation"].(string)
	if !ok || operation == "" {
		return fmt.Errorf("operation is required")
	}

	validOps := map[string]bool{
		"power": true, "sqrt": true, "abs": true, "round": true,
		"ceil": true, "floor": true, "min": true, "max": true,
	}
	if !validOps[operation] {
		return fmt.Errorf("invalid operation: %s", operation)
	}

	if _, ok := config["source_field"].(string); !ok || config["source_field"] == "" {
		return fmt.Errorf("source_field is required")
	}

	if _, ok := config["target_field"].(string); !ok || config["target_field"] == "" {
		return fmt.Errorf("target_field is required")
	}

	// power, min, max require either operand or second_field
	if operation == "power" || operation == "min" || operation == "max" {
		_, hasOperand := config["operand"].(float64)
		_, hasSecondField := config["second_field"].(string)
		if !hasOperand && !hasSecondField {
			return fmt.Errorf("%s operation requires either 'operand' or 'second_field'", operation)
		}
	}

	return nil
}

func (h *AdvanceMath) Execute(ctx context.Context, input helpers.HelperInput) (*helpers.HelperOutput, error) {
	operation := input.Config["operation"].(string)
	sourceField := input.Config["source_field"].(string)
	targetField := input.Config["target_field"].(string)

	output := &helpers.HelperOutput{
		Logs: make([]string, 0),
	}

	// Get source field value
	sourceValue, err := input.Connector.GetContactFieldValue(ctx, input.ContactID, sourceField)
	if err != nil {
		output.Message = fmt.Sprintf("Failed to get source field '%s': %v", sourceField, err)
		return output, err
	}

	// Convert to float64
	sourceNum, err := toFloat64Advanced(sourceValue)
	if err != nil {
		output.Message = fmt.Sprintf("Source field '%s' is not a valid number: %v", sourceField, err)
		return output, err
	}

	output.Logs = append(output.Logs, fmt.Sprintf("Source value: %s = %f", sourceField, sourceNum))

	var result float64

	switch operation {
	case "sqrt":
		if sourceNum < 0 {
			output.Message = "Cannot compute square root of negative number"
			return output, fmt.Errorf("negative number for sqrt")
		}
		result = math.Sqrt(sourceNum)

	case "abs":
		result = math.Abs(sourceNum)

	case "round":
		result = math.Round(sourceNum)

	case "ceil":
		result = math.Ceil(sourceNum)

	case "floor":
		result = math.Floor(sourceNum)

	case "power", "min", "max":
		// Get second operand
		var secondNum float64
		if secondField, ok := input.Config["second_field"].(string); ok && secondField != "" {
			secondValue, err := input.Connector.GetContactFieldValue(ctx, input.ContactID, secondField)
			if err != nil {
				output.Message = fmt.Sprintf("Failed to get second field '%s': %v", secondField, err)
				return output, err
			}
			secondNum, err = toFloat64Advanced(secondValue)
			if err != nil {
				output.Message = fmt.Sprintf("Second field '%s' is not a valid number: %v", secondField, err)
				return output, err
			}
			output.Logs = append(output.Logs, fmt.Sprintf("Second value: %s = %f", secondField, secondNum))
		} else if operand, ok := input.Config["operand"].(float64); ok {
			secondNum = operand
			output.Logs = append(output.Logs, fmt.Sprintf("Operand value: %f", secondNum))
		} else {
			output.Message = fmt.Sprintf("%s operation requires 'operand' or 'second_field'", operation)
			return output, fmt.Errorf("missing operand")
		}

		switch operation {
		case "power":
			result = math.Pow(sourceNum, secondNum)
		case "min":
			result = math.Min(sourceNum, secondNum)
		case "max":
			result = math.Max(sourceNum, secondNum)
		}
	}

	// Save result to target field
	err = input.Connector.SetContactFieldValue(ctx, input.ContactID, targetField, result)
	if err != nil {
		output.Message = fmt.Sprintf("Failed to set target field '%s': %v", targetField, err)
		return output, err
	}

	output.Success = true
	output.Message = fmt.Sprintf("Computed %s(%f) = %f, saved to '%s'", operation, sourceNum, result, targetField)
	output.Actions = []helpers.HelperAction{
		{
			Type:   "field_updated",
			Target: targetField,
			Value:  result,
		},
	}
	output.ModifiedData = map[string]interface{}{
		targetField:   result,
		"operation":   operation,
		"source":      sourceNum,
		"result":      result,
	}
	output.Logs = append(output.Logs, fmt.Sprintf("Result: %s = %f", targetField, result))

	return output, nil
}

// toFloat64Advanced converts various types to float64
func toFloat64Advanced(val interface{}) (float64, error) {
	switch v := val.(type) {
	case float64:
		return v, nil
	case float32:
		return float64(v), nil
	case int:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case string:
		var f float64
		_, err := fmt.Sscanf(v, "%f", &f)
		return f, err
	default:
		return 0, fmt.Errorf("cannot convert %T to float64", val)
	}
}
