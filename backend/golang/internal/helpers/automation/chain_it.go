package automation

import (
	"context"
	"fmt"
	"strconv"

	"github.com/myfusionhelper/api/internal/helpers"
)

func init() {
	helpers.Register("chain_it", func() helpers.Helper { return &ChainIt{} })
}

// ChainIt chains multiple helper executions together with conditional logic and timing control.
// Returns output with actions listing helpers to chain (actual chaining done by execution layer).
type ChainIt struct{}

func (h *ChainIt) GetName() string        { return "Chain It" }
func (h *ChainIt) GetType() string        { return "chain_it" }
func (h *ChainIt) GetCategory() string    { return "automation" }
func (h *ChainIt) GetDescription() string { return "Chain multiple helper executions with conditionals and delays" }
func (h *ChainIt) RequiresCRM() bool      { return false }
func (h *ChainIt) SupportedCRMs() []string { return nil }

func (h *ChainIt) GetConfigSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"helpers": map[string]interface{}{
				"type":        "array",
				"items":       map[string]interface{}{"type": "string"},
				"description": "List of helper type strings to chain in sequence",
			},
			"conditional_field": map[string]interface{}{
				"type":        "string",
				"description": "Optional: field name to check for conditional execution",
			},
			"conditional_value": map[string]interface{}{
				"type":        "string",
				"description": "Optional: value to compare against for conditional execution",
			},
			"conditional_operator": map[string]interface{}{
				"type":        "string",
				"enum":        []string{"equals", "not_equals", "contains", "not_contains", "exists", "not_exists"},
				"description": "Optional: comparison operator for conditional execution (default: equals)",
			},
			"delay_seconds": map[string]interface{}{
				"type":        "number",
				"description": "Optional: delay in seconds between helper executions (default: 0)",
			},
			"on_success_helpers": map[string]interface{}{
				"type":        "array",
				"items":       map[string]interface{}{"type": "string"},
				"description": "Optional: helpers to execute if all primary helpers succeed",
			},
			"on_failure_helpers": map[string]interface{}{
				"type":        "array",
				"items":       map[string]interface{}{"type": "string"},
				"description": "Optional: helpers to execute if any primary helper fails",
			},
		},
		"required": []string{"helpers"},
	}
}

func (h *ChainIt) ValidateConfig(config map[string]interface{}) error {
	helpersList, ok := config["helpers"]
	if !ok {
		return fmt.Errorf("helpers is required")
	}

	switch v := helpersList.(type) {
	case []interface{}:
		if len(v) == 0 {
			return fmt.Errorf("helpers must contain at least one helper type")
		}
	case []string:
		if len(v) == 0 {
			return fmt.Errorf("helpers must contain at least one helper type")
		}
	default:
		return fmt.Errorf("helpers must be an array of strings")
	}

	// Validate conditional operator if provided
	if operator, ok := config["conditional_operator"].(string); ok {
		validOperators := map[string]bool{
			"equals":       true,
			"not_equals":   true,
			"contains":     true,
			"not_contains": true,
			"exists":       true,
			"not_exists":   true,
		}
		if !validOperators[operator] {
			return fmt.Errorf("invalid conditional_operator: %s", operator)
		}
	}

	// Validate delay_seconds if provided
	if delay, ok := config["delay_seconds"]; ok {
		switch v := delay.(type) {
		case float64:
			if v < 0 {
				return fmt.Errorf("delay_seconds must be non-negative")
			}
		case int:
			if v < 0 {
				return fmt.Errorf("delay_seconds must be non-negative")
			}
		default:
			return fmt.Errorf("delay_seconds must be a number")
		}
	}

	return nil
}

func (h *ChainIt) Execute(ctx context.Context, input helpers.HelperInput) (*helpers.HelperOutput, error) {
	helperTypes := extractStringSlice(input.Config["helpers"])

	output := &helpers.HelperOutput{
		Actions: make([]helpers.HelperAction, 0),
		Logs:    make([]string, 0),
	}

	// Check conditional execution if specified
	shouldExecute, conditionLog := h.evaluateCondition(input)
	if conditionLog != "" {
		output.Logs = append(output.Logs, conditionLog)
	}

	if !shouldExecute {
		output.Success = true
		output.Message = "Chain skipped due to conditional check"
		return output, nil
	}

	// Extract delay configuration
	delaySeconds := h.getDelaySeconds(input.Config)

	// Build chain actions for the execution layer to process
	for i, helperType := range helperTypes {
		actionValue := map[string]interface{}{
			"index": i,
		}

		// Add delay metadata if configured
		if delaySeconds > 0 {
			actionValue["delay_seconds"] = delaySeconds
			output.Logs = append(output.Logs, fmt.Sprintf("Chained helper %d: %s (with %d second delay)", i+1, helperType, delaySeconds))
		} else {
			output.Logs = append(output.Logs, fmt.Sprintf("Chained helper %d: %s", i+1, helperType))
		}

		output.Actions = append(output.Actions, helpers.HelperAction{
			Type:   "helper_chain",
			Target: helperType,
			Value:  actionValue,
		})
	}

	// Add success/failure handler metadata
	if successHelpers := extractStringSlice(input.Config["on_success_helpers"]); len(successHelpers) > 0 {
		output.Actions = append(output.Actions, helpers.HelperAction{
			Type:   "helper_chain_success_handler",
			Target: "chain_completion",
			Value:  successHelpers,
		})
		output.Logs = append(output.Logs, fmt.Sprintf("Configured %d success handler(s)", len(successHelpers)))
	}

	if failureHelpers := extractStringSlice(input.Config["on_failure_helpers"]); len(failureHelpers) > 0 {
		output.Actions = append(output.Actions, helpers.HelperAction{
			Type:   "helper_chain_failure_handler",
			Target: "chain_completion",
			Value:  failureHelpers,
		})
		output.Logs = append(output.Logs, fmt.Sprintf("Configured %d failure handler(s)", len(failureHelpers)))
	}

	output.Success = true
	output.Message = fmt.Sprintf("Chained %d helper(s) for sequential execution", len(helperTypes))

	return output, nil
}

// evaluateCondition checks if the chain should execute based on conditional configuration
func (h *ChainIt) evaluateCondition(input helpers.HelperInput) (bool, string) {
	conditionalField, hasField := input.Config["conditional_field"].(string)
	if !hasField || conditionalField == "" {
		return true, "" // No condition specified, execute unconditionally
	}

	operator := "equals" // default
	if op, ok := input.Config["conditional_operator"].(string); ok {
		operator = op
	}

	conditionalValue, _ := input.Config["conditional_value"].(string)

	// Get actual field value from contact data
	var actualValue string
	var fieldExists bool

	if input.ContactData != nil {
		// Check standard fields
		switch conditionalField {
		case "email":
			actualValue = input.ContactData.Email
			fieldExists = actualValue != ""
		case "first_name":
			actualValue = input.ContactData.FirstName
			fieldExists = actualValue != ""
		case "last_name":
			actualValue = input.ContactData.LastName
			fieldExists = actualValue != ""
		case "phone":
			actualValue = input.ContactData.Phone
			fieldExists = actualValue != ""
		default:
			// Check custom fields
			if val, ok := input.ContactData.CustomFields[conditionalField]; ok {
				actualValue = fmt.Sprintf("%v", val)
				fieldExists = true
			}
		}
	}

	// Evaluate condition based on operator
	var result bool
	var logMessage string

	switch operator {
	case "equals":
		result = actualValue == conditionalValue
		logMessage = fmt.Sprintf("Condition: %s == %s (actual: %s) = %v", conditionalField, conditionalValue, actualValue, result)
	case "not_equals":
		result = actualValue != conditionalValue
		logMessage = fmt.Sprintf("Condition: %s != %s (actual: %s) = %v", conditionalField, conditionalValue, actualValue, result)
	case "contains":
		result = fieldExists && len(actualValue) > 0 && len(conditionalValue) > 0 && stringContains(actualValue, conditionalValue)
		logMessage = fmt.Sprintf("Condition: %s contains %s (actual: %s) = %v", conditionalField, conditionalValue, actualValue, result)
	case "not_contains":
		result = !fieldExists || !stringContains(actualValue, conditionalValue)
		logMessage = fmt.Sprintf("Condition: %s not contains %s (actual: %s) = %v", conditionalField, conditionalValue, actualValue, result)
	case "exists":
		result = fieldExists
		logMessage = fmt.Sprintf("Condition: %s exists = %v", conditionalField, result)
	case "not_exists":
		result = !fieldExists
		logMessage = fmt.Sprintf("Condition: %s not exists = %v", conditionalField, result)
	default:
		result = true
		logMessage = fmt.Sprintf("Unknown operator: %s, executing unconditionally", operator)
	}

	return result, logMessage
}

// getDelaySeconds extracts and converts delay_seconds from config
func (h *ChainIt) getDelaySeconds(config map[string]interface{}) int {
	delay, ok := config["delay_seconds"]
	if !ok {
		return 0
	}

	switch v := delay.(type) {
	case float64:
		return int(v)
	case int:
		return v
	case string:
		if intVal, err := strconv.Atoi(v); err == nil {
			return intVal
		}
	}

	return 0
}

// stringContains checks if a string contains a substring (case-sensitive)
func stringContains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || containsSubstring(s, substr))
}

// containsSubstring is a simple substring check
func containsSubstring(s, substr string) bool {
	if len(substr) == 0 {
		return true
	}
	if len(s) < len(substr) {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
