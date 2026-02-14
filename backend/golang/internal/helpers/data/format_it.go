package data

import (
	"context"
	"fmt"
	"strings"
	"unicode"

	"github.com/myfusionhelper/api/internal/helpers"
)

// NewFormatIt creates a new FormatIt helper instance
func NewFormatIt() helpers.Helper { return &FormatIt{} }

func init() {
	helpers.Register("format_it", func() helpers.Helper { return &FormatIt{} })
}

// FormatIt formats a contact field value (uppercase, lowercase, title case, trim, etc.)
type FormatIt struct{}

func (h *FormatIt) GetName() string        { return "Format It" }
func (h *FormatIt) GetType() string        { return "format_it" }
func (h *FormatIt) GetCategory() string    { return "data" }
func (h *FormatIt) GetDescription() string { return "Format a contact field value (uppercase, lowercase, title case, trim, etc.)" }
func (h *FormatIt) RequiresCRM() bool      { return true }
func (h *FormatIt) SupportedCRMs() []string { return nil }

func (h *FormatIt) GetConfigSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"field": map[string]interface{}{
				"type":        "string",
				"description": "The field key to format",
			},
			"format": map[string]interface{}{
				"type":        "string",
				"enum":        []string{"uppercase", "lowercase", "title_case", "trim", "trim_uppercase", "trim_lowercase", "trim_title_case"},
				"description": "The format to apply",
			},
			"target_field": map[string]interface{}{
				"type":        "string",
				"description": "Optional: write result to a different field (defaults to same field)",
			},
		},
		"required": []string{"field", "format"},
	}
}

func (h *FormatIt) ValidateConfig(config map[string]interface{}) error {
	if _, ok := config["field"].(string); !ok || config["field"] == "" {
		return fmt.Errorf("field is required")
	}

	format, ok := config["format"].(string)
	if !ok || format == "" {
		return fmt.Errorf("format is required")
	}

	validFormats := map[string]bool{
		"uppercase": true, "lowercase": true, "title_case": true,
		"trim": true, "trim_uppercase": true, "trim_lowercase": true,
		"trim_title_case": true,
	}
	if !validFormats[format] {
		return fmt.Errorf("invalid format: %s", format)
	}

	return nil
}

func (h *FormatIt) Execute(ctx context.Context, input helpers.HelperInput) (*helpers.HelperOutput, error) {
	field := input.Config["field"].(string)
	format := input.Config["format"].(string)

	targetField := field
	if tf, ok := input.Config["target_field"].(string); ok && tf != "" {
		targetField = tf
	}

	output := &helpers.HelperOutput{
		Logs: make([]string, 0),
	}

	// Get current value
	value, err := input.Connector.GetContactFieldValue(ctx, input.ContactID, field)
	if err != nil {
		output.Message = fmt.Sprintf("Failed to read field '%s': %v", field, err)
		return output, err
	}

	strValue := fmt.Sprintf("%v", value)
	if value == nil || strValue == "" || strValue == "<nil>" {
		output.Success = true
		output.Message = fmt.Sprintf("Field '%s' is empty, nothing to format", field)
		return output, nil
	}

	// Apply formatting
	formatted := applyFormat(strValue, format)

	// Skip if value didn't change
	if formatted == strValue {
		output.Success = true
		output.Message = fmt.Sprintf("Field '%s' already formatted correctly", field)
		output.Logs = append(output.Logs, output.Message)
		return output, nil
	}

	// Write formatted value
	err = input.Connector.SetContactFieldValue(ctx, input.ContactID, targetField, formatted)
	if err != nil {
		output.Message = fmt.Sprintf("Failed to update field '%s': %v", targetField, err)
		return output, err
	}

	output.Success = true
	output.Message = fmt.Sprintf("Formatted '%s' from '%s' to '%s'", field, strValue, formatted)
	output.Actions = []helpers.HelperAction{
		{
			Type:   "field_updated",
			Target: targetField,
			Value:  formatted,
		},
	}
	output.ModifiedData = map[string]interface{}{
		targetField: formatted,
	}
	output.Logs = append(output.Logs, fmt.Sprintf("Formatted field '%s': '%s' → '%s'", field, strValue, formatted))

	return output, nil
}

func applyFormat(value, format string) string {
	switch format {
	case "uppercase":
		return strings.ToUpper(value)
	case "lowercase":
		return strings.ToLower(value)
	case "title_case":
		return toTitleCase(value)
	case "trim":
		return strings.TrimSpace(value)
	case "trim_uppercase":
		return strings.ToUpper(strings.TrimSpace(value))
	case "trim_lowercase":
		return strings.ToLower(strings.TrimSpace(value))
	case "trim_title_case":
		return toTitleCase(strings.TrimSpace(value))
	default:
		return value
	}
}

func toTitleCase(s string) string {
	prev := ' '
	return strings.Map(func(r rune) rune {
		if unicode.IsSpace(rune(prev)) || prev == '-' || prev == '\'' {
			prev = r
			return unicode.ToUpper(r)
		}
		prev = r
		return unicode.ToLower(r)
	}, s)
}
