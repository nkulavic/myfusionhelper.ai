package automation

import (
	"context"
	"fmt"
	"strconv"

	"github.com/myfusionhelper/api/internal/helpers"
)

// NewDripIt creates a new DripIt helper instance
func NewDripIt() helpers.Helper { return &DripIt{} }

func init() {
	helpers.Register("drip_it", func() helpers.Helper { return &DripIt{} })
}

// DripIt manages drip campaign step tracking
type DripIt struct{}

func (h *DripIt) GetName() string        { return "Drip It" }
func (h *DripIt) GetType() string        { return "drip_it" }
func (h *DripIt) GetCategory() string    { return "automation" }
func (h *DripIt) GetDescription() string { return "Manage drip campaign step tracking - trigger next step in sequence" }
func (h *DripIt) RequiresCRM() bool      { return true }
func (h *DripIt) SupportedCRMs() []string { return nil }

func (h *DripIt) GetConfigSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"steps": map[string]interface{}{
				"type":        "array",
				"items":       map[string]interface{}{"type": "string"},
				"description": "Ordered list of automation IDs representing drip steps",
			},
			"state_field": map[string]interface{}{
				"type":        "string",
				"description": "Field to track the current step index",
			},
		},
		"required": []string{"steps", "state_field"},
	}
}

func (h *DripIt) ValidateConfig(config map[string]interface{}) error {
	steps, ok := config["steps"]
	if !ok {
		return fmt.Errorf("steps is required")
	}

	switch v := steps.(type) {
	case []interface{}:
		if len(v) == 0 {
			return fmt.Errorf("steps must contain at least one automation")
		}
	case []string:
		if len(v) == 0 {
			return fmt.Errorf("steps must contain at least one automation")
		}
	default:
		return fmt.Errorf("steps must be an array of strings")
	}

	if _, ok := config["state_field"].(string); !ok || config["state_field"] == "" {
		return fmt.Errorf("state_field is required")
	}

	return nil
}

func (h *DripIt) Execute(ctx context.Context, input helpers.HelperInput) (*helpers.HelperOutput, error) {
	steps := extractStringSlice(input.Config["steps"])
	stateField := input.Config["state_field"].(string)

	output := &helpers.HelperOutput{
		Actions:      make([]helpers.HelperAction, 0),
		ModifiedData: make(map[string]interface{}),
		Logs:         make([]string, 0),
	}

	// Read current step from state field
	currentStep := 0
	stateValue, err := input.Connector.GetContactFieldValue(ctx, input.ContactID, stateField)
	if err == nil && stateValue != nil {
		strVal := fmt.Sprintf("%v", stateValue)
		if strVal != "<nil>" && strVal != "" {
			if parsed, parseErr := strconv.Atoi(strVal); parseErr == nil {
				currentStep = parsed
			}
		}
	}

	// Check if all steps are completed
	if currentStep >= len(steps) {
		output.Success = true
		output.Message = fmt.Sprintf("Drip campaign complete: all %d steps executed", len(steps))
		output.Logs = append(output.Logs, output.Message)
		return output, nil
	}

	// Trigger the current step
	automationID := steps[currentStep]
	err = input.Connector.TriggerAutomation(ctx, input.ContactID, automationID)
	if err != nil {
		output.Message = fmt.Sprintf("Failed to trigger drip step %d (automation %s): %v", currentStep, automationID, err)
		return output, err
	}

	output.Actions = append(output.Actions, helpers.HelperAction{
		Type:   "automation_triggered",
		Target: input.ContactID,
		Value:  automationID,
	})

	// Advance to next step
	nextStep := currentStep + 1
	nextStepStr := fmt.Sprintf("%d", nextStep)

	err = input.Connector.SetContactFieldValue(ctx, input.ContactID, stateField, nextStepStr)
	if err != nil {
		output.Logs = append(output.Logs, fmt.Sprintf("Warning: failed to update state field '%s': %v", stateField, err))
	} else {
		output.Actions = append(output.Actions, helpers.HelperAction{
			Type:   "field_updated",
			Target: stateField,
			Value:  nextStepStr,
		})
		output.ModifiedData[stateField] = nextStepStr
	}

	output.Success = true
	output.Message = fmt.Sprintf("Drip step %d of %d: triggered automation %s", currentStep+1, len(steps), automationID)
	output.Logs = append(output.Logs, fmt.Sprintf("Drip campaign for contact %s: executed step %d/%d (automation %s)", input.ContactID, currentStep+1, len(steps), automationID))

	return output, nil
}
