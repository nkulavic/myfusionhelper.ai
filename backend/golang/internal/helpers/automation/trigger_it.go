package automation

import (
	"context"
	"fmt"

	"github.com/myfusionhelper/api/internal/helpers"
)

func init() {
	helpers.Register("trigger_it", func() helpers.Helper { return &TriggerIt{} })
}

// TriggerIt triggers an automation or campaign sequence for a contact
type TriggerIt struct{}

func (h *TriggerIt) GetName() string        { return "Trigger It" }
func (h *TriggerIt) GetType() string        { return "trigger_it" }
func (h *TriggerIt) GetCategory() string    { return "automation" }
func (h *TriggerIt) GetDescription() string { return "Trigger an automation or campaign sequence for a contact" }
func (h *TriggerIt) RequiresCRM() bool      { return true }
func (h *TriggerIt) SupportedCRMs() []string { return nil }

func (h *TriggerIt) GetConfigSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"automation_id": map[string]interface{}{
				"type":        "string",
				"description": "The automation/campaign/workflow ID to trigger",
			},
		},
		"required": []string{"automation_id"},
	}
}

func (h *TriggerIt) ValidateConfig(config map[string]interface{}) error {
	if _, ok := config["automation_id"].(string); !ok || config["automation_id"] == "" {
		return fmt.Errorf("automation_id is required")
	}
	return nil
}

func (h *TriggerIt) Execute(ctx context.Context, input helpers.HelperInput) (*helpers.HelperOutput, error) {
	automationID := input.Config["automation_id"].(string)

	output := &helpers.HelperOutput{
		Logs: make([]string, 0),
	}

	err := input.Connector.TriggerAutomation(ctx, input.ContactID, automationID)
	if err != nil {
		output.Message = fmt.Sprintf("Failed to trigger automation %s: %v", automationID, err)
		return output, err
	}

	output.Success = true
	output.Message = fmt.Sprintf("Triggered automation %s for contact %s", automationID, input.ContactID)
	output.Actions = []helpers.HelperAction{
		{
			Type:   "automation_triggered",
			Target: input.ContactID,
			Value:  automationID,
		},
	}
	output.Logs = append(output.Logs, output.Message)

	return output, nil
}
