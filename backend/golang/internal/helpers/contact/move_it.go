package contact

import (
	"context"
	"fmt"

	"github.com/myfusionhelper/api/internal/helpers"
)

func init() {
	helpers.Register("move_it", func() helpers.Helper { return &MoveIt{} })
}

// MoveIt moves a field value from source to target and clears the source
type MoveIt struct{}

func (h *MoveIt) GetName() string        { return "Move It" }
func (h *MoveIt) GetType() string        { return "move_it" }
func (h *MoveIt) GetCategory() string    { return "contact" }
func (h *MoveIt) GetDescription() string { return "Move a field value from source to target and clear the source" }
func (h *MoveIt) RequiresCRM() bool      { return true }
func (h *MoveIt) SupportedCRMs() []string { return nil }

func (h *MoveIt) GetConfigSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"source_field": map[string]interface{}{
				"type":        "string",
				"description": "The field key to move from",
			},
			"target_field": map[string]interface{}{
				"type":        "string",
				"description": "The field key to move to",
			},
			"preserve": map[string]interface{}{
				"type":        "boolean",
				"description": "If true, don't overwrite a non-empty target field",
				"default":     false,
			},
		},
		"required": []string{"source_field", "target_field"},
	}
}

func (h *MoveIt) ValidateConfig(config map[string]interface{}) error {
	if _, ok := config["source_field"].(string); !ok || config["source_field"] == "" {
		return fmt.Errorf("source_field is required")
	}
	if _, ok := config["target_field"].(string); !ok || config["target_field"] == "" {
		return fmt.Errorf("target_field is required")
	}
	return nil
}

func (h *MoveIt) Execute(ctx context.Context, input helpers.HelperInput) (*helpers.HelperOutput, error) {
	sourceField := input.Config["source_field"].(string)
	targetField := input.Config["target_field"].(string)

	preserve := false
	if p, ok := input.Config["preserve"].(bool); ok {
		preserve = p
	}

	output := &helpers.HelperOutput{
		Actions:      make([]helpers.HelperAction, 0),
		ModifiedData: make(map[string]interface{}),
		Logs:         make([]string, 0),
	}

	// Get source field value
	sourceValue, err := input.Connector.GetContactFieldValue(ctx, input.ContactID, sourceField)
	if err != nil {
		output.Message = fmt.Sprintf("Failed to read source field '%s': %v", sourceField, err)
		return output, err
	}

	if sourceValue == nil || sourceValue == "" {
		output.Success = true
		output.Message = fmt.Sprintf("Source field '%s' is empty, nothing to move", sourceField)
		output.Logs = append(output.Logs, output.Message)
		return output, nil
	}

	// Check if target already has a value when preserve is true
	if preserve {
		targetValue, err := input.Connector.GetContactFieldValue(ctx, input.ContactID, targetField)
		if err == nil && targetValue != nil && targetValue != "" {
			output.Success = true
			output.Message = fmt.Sprintf("Target field '%s' already has a value, skipping (preserve=true)", targetField)
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

	output.Actions = append(output.Actions, helpers.HelperAction{
		Type:   "field_updated",
		Target: targetField,
		Value:  sourceValue,
	})
	output.ModifiedData[targetField] = sourceValue

	// Clear source field
	err = input.Connector.SetContactFieldValue(ctx, input.ContactID, sourceField, "")
	if err != nil {
		output.Message = fmt.Sprintf("Moved value to '%s' but failed to clear source '%s': %v", targetField, sourceField, err)
		output.Success = true
		output.Logs = append(output.Logs, output.Message)
		return output, nil
	}

	output.Actions = append(output.Actions, helpers.HelperAction{
		Type:   "field_updated",
		Target: sourceField,
		Value:  "",
	})
	output.ModifiedData[sourceField] = ""

	output.Success = true
	output.Message = fmt.Sprintf("Moved value from '%s' to '%s' and cleared source", sourceField, targetField)
	output.Logs = append(output.Logs, fmt.Sprintf("Moved value from '%s' to '%s' on contact %s", sourceField, targetField, input.ContactID))

	return output, nil
}
