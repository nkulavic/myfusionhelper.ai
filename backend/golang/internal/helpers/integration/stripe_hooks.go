package integration

import (
	"context"
	"fmt"

	"github.com/myfusionhelper/api/internal/helpers"
)

// StripeHooks processes Stripe webhook events and triggers CRM actions
type StripeHooks struct{}

func init() {
	helpers.Register("stripe_hooks", func() helpers.Helper {
		return &StripeHooks{}
	})
}

func (h *StripeHooks) GetName() string     { return "Stripe Hooks" }
func (h *StripeHooks) GetType() string     { return "stripe_hooks" }
func (h *StripeHooks) GetCategory() string { return "integration" }
func (h *StripeHooks) GetDescription() string {
	return "Process Stripe webhook events and trigger CRM actions (goals, tags) based on subscription and payment events"
}

func (h *StripeHooks) GetConfigSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"required": []string{"selected_events"},
		"properties": map[string]interface{}{
			"selected_events": map[string]interface{}{
				"type":        "array",
				"items":       map[string]interface{}{"type": "string"},
				"minItems":    1,
				"description": "Stripe event types to process",
			},
			"goal_name": map[string]interface{}{
				"type":        "string",
				"description": "Goal to achieve when event occurs (optional)",
			},
			"event_tags": map[string]interface{}{
				"type":        "array",
				"items":       map[string]interface{}{"type": "string"},
				"description": "Tags to apply when event occurs (optional)",
			},
		},
	}
}

func (h *StripeHooks) RequiresCRM() bool { return true }

func (h *StripeHooks) SupportedCRMs() []string {
	return nil // All CRMs supported
}

func (h *StripeHooks) ValidateConfig(config map[string]interface{}) error {
	selectedEvents, ok := config["selected_events"].([]interface{})
	if !ok {
		return fmt.Errorf("selected_events must be an array")
	}

	if len(selectedEvents) == 0 {
		return fmt.Errorf("at least one event must be selected")
	}

	// Validate each event is a string
	for i, evt := range selectedEvents {
		if _, ok := evt.(string); !ok {
			return fmt.Errorf("selected_events[%d] must be a string", i)
		}
	}

	// Validate event_tags if provided
	if eventTags, ok := config["event_tags"].([]interface{}); ok {
		for i, tag := range eventTags {
			if _, ok := tag.(string); !ok {
				return fmt.Errorf("event_tags[%d] must be a string", i)
			}
		}
	}

	return nil
}

func (h *StripeHooks) Execute(ctx context.Context, input helpers.HelperInput) (*helpers.HelperOutput, error) {
	selectedEvents := input.Config["selected_events"].([]interface{})
	goalName, _ := input.Config["goal_name"].(string)
	eventTagsRaw, _ := input.Config["event_tags"].([]interface{})

	// Convert event tags to string slice
	var eventTags []string
	for _, tag := range eventTagsRaw {
		if tagStr, ok := tag.(string); ok && tagStr != "" {
			eventTags = append(eventTags, tagStr)
		}
	}

	output := &helpers.HelperOutput{
		Success: true,
		Message: fmt.Sprintf("Stripe hooks configured for %d event types", len(selectedEvents)),
		ModifiedData: map[string]interface{}{
			"configured_events": selectedEvents,
			"goal_name":         goalName,
			"event_tags":        eventTags,
		},
		Actions: []helpers.HelperAction{},
		Logs:    []string{},
	}

	// Note: This helper configures webhook processing, actual event handling
	// would be done by the Stripe webhook handler when events arrive
	output.Logs = append(output.Logs, fmt.Sprintf("Configured to process %d Stripe event types", len(selectedEvents)))

	if goalName != "" {
		output.Logs = append(output.Logs, fmt.Sprintf("Will trigger goal: %s", goalName))
	}

	if len(eventTags) > 0 {
		output.Logs = append(output.Logs, fmt.Sprintf("Will apply %d tags on events", len(eventTags)))
	}

	return output, nil
}
