package contact

import (
	"context"
	"fmt"

	"github.com/myfusionhelper/api/internal/helpers"
)

// NewClearIt creates a new ClearIt helper instance
func NewClearIt() helpers.Helper { return &ClearIt{} }

func init() {
	helpers.Register("clear_it", func() helpers.Helper { return &ClearIt{} })
}

// ClearIt clears specified fields to empty on a contact
type ClearIt struct{}

func (h *ClearIt) GetName() string        { return "Clear It" }
func (h *ClearIt) GetType() string        { return "clear_it" }
func (h *ClearIt) GetCategory() string    { return "contact" }
func (h *ClearIt) GetDescription() string { return "Clear specified fields to empty on a contact" }
func (h *ClearIt) RequiresCRM() bool      { return true }
func (h *ClearIt) SupportedCRMs() []string { return nil }

func (h *ClearIt) GetConfigSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"fields": map[string]interface{}{
				"type":        "array",
				"items":       map[string]interface{}{"type": "string"},
				"description": "List of field keys to clear",
			},
		},
		"required": []string{"fields"},
	}
}

func (h *ClearIt) ValidateConfig(config map[string]interface{}) error {
	fields, ok := config["fields"]
	if !ok {
		return fmt.Errorf("fields is required")
	}

	switch v := fields.(type) {
	case []interface{}:
		if len(v) == 0 {
			return fmt.Errorf("fields must contain at least one field")
		}
	case []string:
		if len(v) == 0 {
			return fmt.Errorf("fields must contain at least one field")
		}
	default:
		return fmt.Errorf("fields must be an array of strings")
	}

	return nil
}

func (h *ClearIt) Execute(ctx context.Context, input helpers.HelperInput) (*helpers.HelperOutput, error) {
	fields := extractStringSlice(input.Config["fields"])

	output := &helpers.HelperOutput{
		Actions:      make([]helpers.HelperAction, 0),
		ModifiedData: make(map[string]interface{}),
		Logs:         make([]string, 0),
	}

	cleared := 0
	for _, field := range fields {
		err := input.Connector.SetContactFieldValue(ctx, input.ContactID, field, "")
		if err != nil {
			output.Logs = append(output.Logs, fmt.Sprintf("Failed to clear field '%s': %v", field, err))
			continue
		}

		output.Actions = append(output.Actions, helpers.HelperAction{
			Type:   "field_updated",
			Target: field,
			Value:  "",
		})
		output.ModifiedData[field] = ""
		cleared++
	}

	output.Success = cleared > 0
	if output.Success {
		output.Message = fmt.Sprintf("Cleared %d of %d field(s)", cleared, len(fields))
	} else {
		output.Message = "Failed to clear any fields"
	}
	output.Logs = append(output.Logs, output.Message)

	return output, nil
}
