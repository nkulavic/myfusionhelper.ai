package contact

import (
	"context"
	"fmt"

	"github.com/myfusionhelper/api/internal/helpers"
)

// NewFieldToField creates a new FieldToField helper instance
func NewFieldToField() helpers.Helper { return &FieldToField{} }

func init() {
	helpers.Register("field_to_field", func() helpers.Helper { return &FieldToField{} })
}

// FieldToField maps multiple fields from source to target in a single operation
type FieldToField struct{}

func (h *FieldToField) GetName() string        { return "Field to Field" }
func (h *FieldToField) GetType() string        { return "field_to_field" }
func (h *FieldToField) GetCategory() string    { return "contact" }
func (h *FieldToField) GetDescription() string { return "Map multiple field values from source fields to target fields" }
func (h *FieldToField) RequiresCRM() bool      { return true }
func (h *FieldToField) SupportedCRMs() []string { return nil }

func (h *FieldToField) GetConfigSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"mappings": map[string]interface{}{
				"type": "array",
				"items": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"source": map[string]interface{}{"type": "string"},
						"target": map[string]interface{}{"type": "string"},
					},
				},
				"description": "Array of {source, target} field mappings",
			},
			"overwrite": map[string]interface{}{
				"type":        "boolean",
				"description": "Whether to overwrite existing target field values",
				"default":     true,
			},
		},
		"required": []string{"mappings"},
	}
}

func (h *FieldToField) ValidateConfig(config map[string]interface{}) error {
	mappings, ok := config["mappings"]
	if !ok {
		return fmt.Errorf("mappings is required")
	}

	mappingList, ok := mappings.([]interface{})
	if !ok {
		return fmt.Errorf("mappings must be an array")
	}
	if len(mappingList) == 0 {
		return fmt.Errorf("mappings must contain at least one mapping")
	}

	for i, m := range mappingList {
		mapping, ok := m.(map[string]interface{})
		if !ok {
			return fmt.Errorf("mapping %d must be an object", i)
		}
		if _, ok := mapping["source"].(string); !ok {
			return fmt.Errorf("mapping %d requires a source field", i)
		}
		if _, ok := mapping["target"].(string); !ok {
			return fmt.Errorf("mapping %d requires a target field", i)
		}
	}
	return nil
}

func (h *FieldToField) Execute(ctx context.Context, input helpers.HelperInput) (*helpers.HelperOutput, error) {
	mappings := input.Config["mappings"].([]interface{})
	overwrite := true
	if ow, ok := input.Config["overwrite"].(bool); ok {
		overwrite = ow
	}

	output := &helpers.HelperOutput{
		Actions:      make([]helpers.HelperAction, 0),
		ModifiedData: make(map[string]interface{}),
		Logs:         make([]string, 0),
	}

	copied := 0
	skipped := 0

	for _, m := range mappings {
		mapping := m.(map[string]interface{})
		source := mapping["source"].(string)
		target := mapping["target"].(string)

		// Read source
		val, err := input.Connector.GetContactFieldValue(ctx, input.ContactID, source)
		if err != nil {
			output.Logs = append(output.Logs, fmt.Sprintf("Failed to read '%s': %v", source, err))
			continue
		}

		if val == nil || val == "" {
			output.Logs = append(output.Logs, fmt.Sprintf("Source '%s' is empty, skipping", source))
			skipped++
			continue
		}

		// Check target if not overwriting
		if !overwrite {
			targetVal, err := input.Connector.GetContactFieldValue(ctx, input.ContactID, target)
			if err == nil && targetVal != nil && targetVal != "" {
				output.Logs = append(output.Logs, fmt.Sprintf("Target '%s' already has value, skipping", target))
				skipped++
				continue
			}
		}

		// Write target
		err = input.Connector.SetContactFieldValue(ctx, input.ContactID, target, val)
		if err != nil {
			output.Logs = append(output.Logs, fmt.Sprintf("Failed to write '%s': %v", target, err))
			continue
		}

		output.Actions = append(output.Actions, helpers.HelperAction{
			Type:   "field_updated",
			Target: target,
			Value:  val,
		})
		output.ModifiedData[target] = val
		copied++
	}

	output.Success = copied > 0
	output.Message = fmt.Sprintf("Mapped %d field(s), skipped %d", copied, skipped)
	output.Logs = append(output.Logs, output.Message)

	return output, nil
}
