package automation

import (
	"context"
	"fmt"

	"github.com/myfusionhelper/api/internal/helpers"
)

func init() {
	helpers.Register("action_it", func() helpers.Helper { return &ActionIt{} })
}

// ActionIt triggers multiple automations in sequence
type ActionIt struct{}

func (h *ActionIt) GetName() string        { return "Action It" }
func (h *ActionIt) GetType() string        { return "action_it" }
func (h *ActionIt) GetCategory() string    { return "automation" }
func (h *ActionIt) GetDescription() string { return "Trigger multiple automations in sequence for a contact" }
func (h *ActionIt) RequiresCRM() bool      { return true }
func (h *ActionIt) SupportedCRMs() []string { return nil }

func (h *ActionIt) GetConfigSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"automation_ids": map[string]interface{}{
				"type":        "array",
				"items":       map[string]interface{}{"type": "string"},
				"description": "List of automation IDs to trigger in sequence",
			},
		},
		"required": []string{"automation_ids"},
	}
}

func (h *ActionIt) ValidateConfig(config map[string]interface{}) error {
	ids, ok := config["automation_ids"]
	if !ok {
		return fmt.Errorf("automation_ids is required")
	}

	switch v := ids.(type) {
	case []interface{}:
		if len(v) == 0 {
			return fmt.Errorf("automation_ids must contain at least one automation")
		}
	case []string:
		if len(v) == 0 {
			return fmt.Errorf("automation_ids must contain at least one automation")
		}
	default:
		return fmt.Errorf("automation_ids must be an array of strings")
	}

	return nil
}

func (h *ActionIt) Execute(ctx context.Context, input helpers.HelperInput) (*helpers.HelperOutput, error) {
	automationIDs := extractStringSlice(input.Config["automation_ids"])

	output := &helpers.HelperOutput{
		Actions: make([]helpers.HelperAction, 0),
		Logs:    make([]string, 0),
	}

	triggered := 0
	for _, automationID := range automationIDs {
		err := input.Connector.TriggerAutomation(ctx, input.ContactID, automationID)
		if err != nil {
			output.Logs = append(output.Logs, fmt.Sprintf("Failed to trigger automation %s: %v", automationID, err))
			continue
		}

		output.Actions = append(output.Actions, helpers.HelperAction{
			Type:   "automation_triggered",
			Target: input.ContactID,
			Value:  automationID,
		})
		output.Logs = append(output.Logs, fmt.Sprintf("Triggered automation %s", automationID))
		triggered++
	}

	output.Success = triggered > 0
	if output.Success {
		output.Message = fmt.Sprintf("Triggered %d of %d automation(s)", triggered, len(automationIDs))
	} else {
		output.Message = "Failed to trigger any automations"
	}

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
