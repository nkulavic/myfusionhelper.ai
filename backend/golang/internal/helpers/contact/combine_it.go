package contact

import (
	"context"
	"fmt"

	"github.com/myfusionhelper/api/internal/helpers"
)

// NewCombineIt creates a new CombineIt helper instance
func NewCombineIt() helpers.Helper { return &CombineIt{} }

func init() {
	helpers.Register("combine_it", func() helpers.Helper { return &CombineIt{} })
}

// CombineIt combines multiple field values into one field with a separator
type CombineIt struct{}

func (h *CombineIt) GetName() string        { return "Combine It" }
func (h *CombineIt) GetType() string        { return "combine_it" }
func (h *CombineIt) GetCategory() string    { return "contact" }
func (h *CombineIt) GetDescription() string { return "Combine multiple field values into one field with a separator" }
func (h *CombineIt) RequiresCRM() bool      { return true }
func (h *CombineIt) SupportedCRMs() []string { return nil }

func (h *CombineIt) GetConfigSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"source_fields": map[string]interface{}{
				"type":        "array",
				"items":       map[string]interface{}{"type": "string"},
				"description": "List of field keys to combine",
			},
			"target_field": map[string]interface{}{
				"type":        "string",
				"description": "The field key to store the combined value",
			},
			"separator": map[string]interface{}{
				"type":        "string",
				"description": "Separator between combined values",
				"default":     " ",
			},
			"skip_empty": map[string]interface{}{
				"type":        "boolean",
				"description": "Skip empty source fields",
				"default":     true,
			},
		},
		"required": []string{"source_fields", "target_field"},
	}
}

func (h *CombineIt) ValidateConfig(config map[string]interface{}) error {
	sources, ok := config["source_fields"]
	if !ok {
		return fmt.Errorf("source_fields is required")
	}
	switch v := sources.(type) {
	case []interface{}:
		if len(v) == 0 {
			return fmt.Errorf("source_fields must contain at least one field")
		}
	case []string:
		if len(v) == 0 {
			return fmt.Errorf("source_fields must contain at least one field")
		}
	default:
		return fmt.Errorf("source_fields must be an array of strings")
	}

	if _, ok := config["target_field"].(string); !ok || config["target_field"] == "" {
		return fmt.Errorf("target_field is required")
	}
	return nil
}

func (h *CombineIt) Execute(ctx context.Context, input helpers.HelperInput) (*helpers.HelperOutput, error) {
	sourceFields := extractStringSlice(input.Config["source_fields"])
	targetField := input.Config["target_field"].(string)
	separator := " "
	if s, ok := input.Config["separator"].(string); ok {
		separator = s
	}
	skipEmpty := true
	if se, ok := input.Config["skip_empty"].(bool); ok {
		skipEmpty = se
	}

	output := &helpers.HelperOutput{
		Logs: make([]string, 0),
	}

	// Collect source values
	values := make([]string, 0, len(sourceFields))
	for _, field := range sourceFields {
		val, err := input.Connector.GetContactFieldValue(ctx, input.ContactID, field)
		if err != nil {
			output.Logs = append(output.Logs, fmt.Sprintf("Warning: could not read field '%s': %v", field, err))
			if !skipEmpty {
				values = append(values, "")
			}
			continue
		}

		strVal := fmt.Sprintf("%v", val)
		if strVal == "<nil>" {
			strVal = ""
		}

		if skipEmpty && strVal == "" {
			continue
		}
		values = append(values, strVal)
	}

	if len(values) == 0 {
		output.Success = true
		output.Message = "All source fields are empty, nothing to combine"
		return output, nil
	}

	// Concatenate with separator
	combined := ""
	for i, v := range values {
		if i > 0 {
			combined += separator
		}
		combined += v
	}

	err := input.Connector.SetContactFieldValue(ctx, input.ContactID, targetField, combined)
	if err != nil {
		output.Message = fmt.Sprintf("Failed to set combined value: %v", err)
		return output, err
	}

	output.Success = true
	output.Message = fmt.Sprintf("Combined %d fields into '%s'", len(values), targetField)
	output.Actions = []helpers.HelperAction{
		{
			Type:   "field_updated",
			Target: targetField,
			Value:  combined,
		},
	}
	output.ModifiedData = map[string]interface{}{
		targetField: combined,
	}
	output.Logs = append(output.Logs, fmt.Sprintf("Combined %d source values into '%s' on contact %s", len(values), targetField, input.ContactID))

	return output, nil
}
