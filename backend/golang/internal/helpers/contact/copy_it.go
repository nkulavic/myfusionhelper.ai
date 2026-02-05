package contact

import (
	"context"
	"fmt"

	"github.com/myfusionhelper/api/internal/helpers"
)

func init() {
	helpers.Register("copy_it", func() helpers.Helper { return &CopyIt{} })
}

// CopyIt copies field values from one field to another on a contact
type CopyIt struct{}

func (h *CopyIt) GetName() string        { return "Copy It" }
func (h *CopyIt) GetType() string        { return "copy_it" }
func (h *CopyIt) GetCategory() string    { return "contact" }
func (h *CopyIt) GetDescription() string { return "Copy a field value from one field to another on a contact" }
func (h *CopyIt) RequiresCRM() bool      { return true }
func (h *CopyIt) SupportedCRMs() []string { return nil }

func (h *CopyIt) GetConfigSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"source_field": map[string]interface{}{
				"type":        "string",
				"description": "The field key to copy from",
			},
			"target_field": map[string]interface{}{
				"type":        "string",
				"description": "The field key to copy to",
			},
			"overwrite": map[string]interface{}{
				"type":        "boolean",
				"description": "Whether to overwrite existing target field value",
				"default":     true,
			},
		},
		"required": []string{"source_field", "target_field"},
	}
}

func (h *CopyIt) ValidateConfig(config map[string]interface{}) error {
	if _, ok := config["source_field"].(string); !ok || config["source_field"] == "" {
		return fmt.Errorf("source_field is required")
	}
	if _, ok := config["target_field"].(string); !ok || config["target_field"] == "" {
		return fmt.Errorf("target_field is required")
	}
	return nil
}

func (h *CopyIt) Execute(ctx context.Context, input helpers.HelperInput) (*helpers.HelperOutput, error) {
	sourceField := input.Config["source_field"].(string)
	targetField := input.Config["target_field"].(string)

	overwrite := true
	if ow, ok := input.Config["overwrite"].(bool); ok {
		overwrite = ow
	}

	output := &helpers.HelperOutput{
		Logs: make([]string, 0),
	}

	// Get source field value
	sourceValue, err := input.Connector.GetContactFieldValue(ctx, input.ContactID, sourceField)
	if err != nil {
		output.Message = fmt.Sprintf("Failed to read source field '%s': %v", sourceField, err)
		return output, err
	}

	if sourceValue == nil || sourceValue == "" {
		output.Success = true
		output.Message = fmt.Sprintf("Source field '%s' is empty, nothing to copy", sourceField)
		output.Logs = append(output.Logs, output.Message)
		return output, nil
	}

	// Check if target already has a value
	if !overwrite {
		targetValue, err := input.Connector.GetContactFieldValue(ctx, input.ContactID, targetField)
		if err == nil && targetValue != nil && targetValue != "" {
			output.Success = true
			output.Message = fmt.Sprintf("Target field '%s' already has a value, skipping (overwrite=false)", targetField)
			output.Logs = append(output.Logs, output.Message)
			return output, nil
		}
	}

	// Set target field value
	err = input.Connector.SetContactFieldValue(ctx, input.ContactID, targetField, sourceValue)
	if err != nil {
		output.Message = fmt.Sprintf("Failed to write to target field '%s': %v", targetField, err)
		return output, err
	}

	output.Success = true
	output.Message = fmt.Sprintf("Copied '%s' to '%s'", sourceField, targetField)
	output.Actions = []helpers.HelperAction{
		{
			Type:   "field_updated",
			Target: targetField,
			Value:  sourceValue,
		},
	}
	output.ModifiedData = map[string]interface{}{
		targetField: sourceValue,
	}
	output.Logs = append(output.Logs, fmt.Sprintf("Copied value from '%s' to '%s' on contact %s", sourceField, targetField, input.ContactID))

	return output, nil
}
