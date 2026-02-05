package integration

import (
	"context"
	"fmt"

	"github.com/myfusionhelper/api/internal/helpers"
)

func init() {
	helpers.Register("hook_it", func() helpers.Helper { return &HookIt{} })
}

// HookIt handles generic webhook events (contact.add, invoice.add, etc.)
// and fires corresponding goals in the CRM automation system.
// Ported from legacy PHP hook_it_contact and hook_it_invoice helpers.
type HookIt struct{}

func (h *HookIt) GetName() string     { return "Hook It" }
func (h *HookIt) GetType() string     { return "hook_it" }
func (h *HookIt) GetCategory() string { return "integration" }
func (h *HookIt) GetDescription() string {
	return "Handle webhook events and trigger corresponding automation goals"
}
func (h *HookIt) RequiresCRM() bool       { return true }
func (h *HookIt) SupportedCRMs() []string { return nil }

func (h *HookIt) GetConfigSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"hook_action": map[string]interface{}{
				"type":        "string",
				"description": "The webhook action to listen for (e.g., contact.add, invoice.add, order.add)",
			},
			"goal_prefix": map[string]interface{}{
				"type":        "string",
				"description": "Custom prefix for the goal call name (defaults to action-based name)",
			},
			"actions": map[string]interface{}{
				"type": "array",
				"items": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"event":     map[string]interface{}{"type": "string", "description": "The event name to match"},
						"goal_name": map[string]interface{}{"type": "string", "description": "Goal call name to achieve"},
					},
				},
				"description": "List of event-to-goal mappings",
			},
		},
		"required": []string{},
	}
}

func (h *HookIt) ValidateConfig(config map[string]interface{}) error {
	// Config is flexible - hook_action or actions array
	return nil
}

func (h *HookIt) Execute(ctx context.Context, input helpers.HelperInput) (*helpers.HelperOutput, error) {
	integration := "myfusionhelper"
	helperID := input.HelperID

	output := &helpers.HelperOutput{
		Actions: make([]helpers.HelperAction, 0),
		Logs:    make([]string, 0),
	}

	hookAction := ""
	if ha, ok := input.Config["hook_action"].(string); ok {
		hookAction = ha
	}

	// Handle specific hook actions with default goal names
	if hookAction != "" {
		goalName := ""
		switch hookAction {
		case "contact.add":
			goalName = fmt.Sprintf("newcontact%s", helperID)
		case "invoice.add":
			goalName = fmt.Sprintf("newinvoice%s", helperID)
		case "order.add":
			goalName = fmt.Sprintf("neworder%s", helperID)
		default:
			// Generic: strip dots and use as goal name
			goalName = fmt.Sprintf("%s%s", hookAction, helperID)
		}

		// Allow custom goal prefix override
		if prefix, ok := input.Config["goal_prefix"].(string); ok && prefix != "" {
			goalName = fmt.Sprintf("%s%s", prefix, helperID)
		}

		err := input.Connector.AchieveGoal(ctx, input.ContactID, goalName, integration)
		if err != nil {
			output.Message = fmt.Sprintf("Failed to achieve goal '%s' for hook '%s': %v", goalName, hookAction, err)
			output.Logs = append(output.Logs, output.Message)
			return output, err
		}

		output.Actions = append(output.Actions, helpers.HelperAction{
			Type:   "goal_achieved",
			Target: input.ContactID,
			Value:  goalName,
		})
		output.Logs = append(output.Logs, fmt.Sprintf("Hook '%s' fired goal '%s' for contact %s", hookAction, goalName, input.ContactID))
	}

	// Handle event-to-goal mappings array
	if actions, ok := input.Config["actions"].([]interface{}); ok {
		for _, a := range actions {
			actionMap, ok := a.(map[string]interface{})
			if !ok {
				continue
			}

			event, _ := actionMap["event"].(string)
			goalName, _ := actionMap["goal_name"].(string)

			if event == "" || goalName == "" {
				continue
			}

			// Check if the current hook action matches this event
			if hookAction == event || hookAction == "" {
				err := input.Connector.AchieveGoal(ctx, input.ContactID, goalName, integration)
				if err != nil {
					output.Logs = append(output.Logs, fmt.Sprintf("Failed to achieve goal '%s' for event '%s': %v", goalName, event, err))
					continue
				}

				output.Actions = append(output.Actions, helpers.HelperAction{
					Type:   "goal_achieved",
					Target: input.ContactID,
					Value:  goalName,
				})
				output.Logs = append(output.Logs, fmt.Sprintf("Event '%s' fired goal '%s'", event, goalName))
			}
		}
	}

	output.Success = len(output.Actions) > 0
	if output.Success {
		output.Message = fmt.Sprintf("Webhook handler fired %d goal(s)", len(output.Actions))
	} else {
		output.Success = true // Still successful even with no actions (no matching events)
		output.Message = "Webhook received, no matching events to process"
	}

	output.ModifiedData = map[string]interface{}{
		"hook_action": hookAction,
		"goals_fired": len(output.Actions),
	}

	return output, nil
}
