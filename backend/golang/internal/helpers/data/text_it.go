package data

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/myfusionhelper/api/internal/helpers"
)

// NewTextIt creates a new TextIt helper instance
func NewTextIt() helpers.Helper { return &TextIt{} }

func init() {
	helpers.Register("text_it", func() helpers.Helper { return &TextIt{} })
}

// TextIt performs text manipulation on contact field values
type TextIt struct{}

func (h *TextIt) GetName() string        { return "Text It" }
func (h *TextIt) GetType() string        { return "text_it" }
func (h *TextIt) GetCategory() string    { return "data" }
func (h *TextIt) GetDescription() string { return "Perform text manipulation on contact field values" }
func (h *TextIt) RequiresCRM() bool      { return true }
func (h *TextIt) SupportedCRMs() []string { return nil }

func (h *TextIt) GetConfigSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"field": map[string]interface{}{
				"type":        "string",
				"description": "The field to manipulate",
			},
			"operation": map[string]interface{}{
				"type":        "string",
				"enum":        []string{"prepend", "append", "replace", "remove", "truncate", "extract_email_domain", "extract_numbers", "slug", "reverse"},
				"description": "The text operation to perform",
			},
			"value": map[string]interface{}{
				"type":        "string",
				"description": "Value to use in the operation (text to prepend/append/replace/remove)",
			},
			"replace_with": map[string]interface{}{
				"type":        "string",
				"description": "Replacement text (for replace operation)",
			},
			"max_length": map[string]interface{}{
				"type":        "integer",
				"description": "Maximum length for truncate operation",
			},
			"target_field": map[string]interface{}{
				"type":        "string",
				"description": "Optional field to store result (defaults to same field)",
			},
		},
		"required": []string{"field", "operation"},
	}
}

func (h *TextIt) ValidateConfig(config map[string]interface{}) error {
	if _, ok := config["field"].(string); !ok || config["field"] == "" {
		return fmt.Errorf("field is required")
	}

	op, ok := config["operation"].(string)
	if !ok || op == "" {
		return fmt.Errorf("operation is required")
	}

	validOps := map[string]bool{
		"prepend": true, "append": true, "replace": true, "remove": true,
		"truncate": true, "extract_email_domain": true, "extract_numbers": true,
		"slug": true, "reverse": true,
	}
	if !validOps[op] {
		return fmt.Errorf("invalid operation: %s", op)
	}

	needsValue := map[string]bool{"prepend": true, "append": true, "replace": true, "remove": true}
	if needsValue[op] {
		if _, ok := config["value"].(string); !ok {
			return fmt.Errorf("value is required for operation '%s'", op)
		}
	}

	if op == "truncate" {
		if _, ok := config["max_length"]; !ok {
			return fmt.Errorf("max_length is required for truncate operation")
		}
	}

	return nil
}

func (h *TextIt) Execute(ctx context.Context, input helpers.HelperInput) (*helpers.HelperOutput, error) {
	field := input.Config["field"].(string)
	operation := input.Config["operation"].(string)
	targetField := field
	if tf, ok := input.Config["target_field"].(string); ok && tf != "" {
		targetField = tf
	}

	output := &helpers.HelperOutput{
		Logs: make([]string, 0),
	}

	// Get current value
	rawValue, err := input.Connector.GetContactFieldValue(ctx, input.ContactID, field)
	if err != nil {
		output.Message = fmt.Sprintf("Failed to read field '%s': %v", field, err)
		return output, err
	}

	currentValue := fmt.Sprintf("%v", rawValue)
	if currentValue == "<nil>" {
		currentValue = ""
	}

	var result string

	switch operation {
	case "prepend":
		prefix := input.Config["value"].(string)
		result = prefix + currentValue

	case "append":
		suffix := input.Config["value"].(string)
		result = currentValue + suffix

	case "replace":
		search := input.Config["value"].(string)
		replaceWith := ""
		if rw, ok := input.Config["replace_with"].(string); ok {
			replaceWith = rw
		}
		result = strings.ReplaceAll(currentValue, search, replaceWith)

	case "remove":
		removeStr := input.Config["value"].(string)
		result = strings.ReplaceAll(currentValue, removeStr, "")

	case "truncate":
		maxLen := int(toFloat64(input.Config["max_length"]))
		if len(currentValue) > maxLen {
			result = currentValue[:maxLen]
		} else {
			result = currentValue
		}

	case "extract_email_domain":
		parts := strings.Split(currentValue, "@")
		if len(parts) == 2 {
			result = parts[1]
		} else {
			result = currentValue
		}

	case "extract_numbers":
		re := regexp.MustCompile(`[0-9.]+`)
		matches := re.FindAllString(currentValue, -1)
		result = strings.Join(matches, "")

	case "slug":
		result = strings.ToLower(currentValue)
		re := regexp.MustCompile(`[^a-z0-9]+`)
		result = re.ReplaceAllString(result, "-")
		result = strings.Trim(result, "-")

	case "reverse":
		runes := []rune(currentValue)
		for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
			runes[i], runes[j] = runes[j], runes[i]
		}
		result = string(runes)
	}

	// Skip if unchanged
	if result == currentValue && targetField == field {
		output.Success = true
		output.Message = "Value unchanged, no update needed"
		return output, nil
	}

	// Set result
	err = input.Connector.SetContactFieldValue(ctx, input.ContactID, targetField, result)
	if err != nil {
		output.Message = fmt.Sprintf("Failed to set result: %v", err)
		return output, err
	}

	output.Success = true
	output.Message = fmt.Sprintf("Text %s: '%s' -> '%s'", operation, currentValue, result)
	output.Actions = []helpers.HelperAction{{Type: "field_updated", Target: targetField, Value: result}}
	output.ModifiedData = map[string]interface{}{targetField: result}
	output.Logs = append(output.Logs, fmt.Sprintf("Text '%s' on field '%s': result stored in '%s'", operation, field, targetField))

	return output, nil
}
