package contact

import (
	"context"
	"fmt"

	"github.com/myfusionhelper/api/internal/helpers"
)

// NewMergeIt creates a new MergeIt helper instance
func NewMergeIt() helpers.Helper { return &MergeIt{} }

func init() {
	helpers.Register("merge_it", func() helpers.Helper { return &MergeIt{} })
}

// MergeIt merges field values from multiple source fields into a target field
type MergeIt struct{}

func (h *MergeIt) GetName() string        { return "Merge It" }
func (h *MergeIt) GetType() string        { return "merge_it" }
func (h *MergeIt) GetCategory() string    { return "contact" }
func (h *MergeIt) GetDescription() string { return "Merge values from multiple fields into a single target field" }
func (h *MergeIt) RequiresCRM() bool      { return true }
func (h *MergeIt) SupportedCRMs() []string { return nil }

func (h *MergeIt) GetConfigSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"source_fields": map[string]interface{}{
				"type":        "array",
				"items":       map[string]interface{}{"type": "string"},
				"description": "List of field keys to merge from",
			},
			"target_field": map[string]interface{}{
				"type":        "string",
				"description": "The field key to merge into",
			},
			"separator": map[string]interface{}{
				"type":        "string",
				"description": "Separator between merged values",
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

func (h *MergeIt) ValidateConfig(config map[string]interface{}) error {
	sources, ok := config["source_fields"]
	if !ok {
		return fmt.Errorf("source_fields is required")
	}
	switch v := sources.(type) {
	case []interface{}:
		if len(v) < 2 {
			return fmt.Errorf("source_fields must contain at least two fields")
		}
	case []string:
		if len(v) < 2 {
			return fmt.Errorf("source_fields must contain at least two fields")
		}
	default:
		return fmt.Errorf("source_fields must be an array of strings")
	}

	if _, ok := config["target_field"].(string); !ok || config["target_field"] == "" {
		return fmt.Errorf("target_field is required")
	}
	return nil
}

func (h *MergeIt) Execute(ctx context.Context, input helpers.HelperInput) (*helpers.HelperOutput, error) {
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
		output.Message = "All source fields are empty, nothing to merge"
		return output, nil
	}

	// Merge and set
	merged := ""
	for i, v := range values {
		if i > 0 {
			merged += separator
		}
		merged += v
	}

	err := input.Connector.SetContactFieldValue(ctx, input.ContactID, targetField, merged)
	if err != nil {
		output.Message = fmt.Sprintf("Failed to set merged value: %v", err)
		return output, err
	}

	output.Success = true
	output.Message = fmt.Sprintf("Merged %d fields into '%s'", len(values), targetField)
	output.Actions = []helpers.HelperAction{
		{
			Type:   "field_updated",
			Target: targetField,
			Value:  merged,
		},
	}
	output.ModifiedData = map[string]interface{}{
		targetField: merged,
	}
	output.Logs = append(output.Logs, fmt.Sprintf("Merged %d source values into '%s' on contact %s", len(values), targetField, input.ContactID))

	return output, nil
}

func extractStringSlice(v interface{}) []string {
	switch val := v.(type) {
	case []string:
		return val
	case []interface{}:
		result := make([]string, 0, len(val))
		for _, item := range val {
			if s, ok := item.(string); ok {
				result = append(result, s)
			} else {
				result = append(result, fmt.Sprintf("%v", item))
			}
		}
		return result
	default:
		return nil
	}
}
