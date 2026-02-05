package automation

import (
	"context"
	"fmt"

	"github.com/myfusionhelper/api/internal/helpers"
)

func init() {
	helpers.Register("goal_it", func() helpers.Helper { return &GoalIt{} })
}

// GoalIt achieves a campaign goal for a contact in the CRM
type GoalIt struct{}

func (h *GoalIt) GetName() string        { return "Goal It" }
func (h *GoalIt) GetType() string        { return "goal_it" }
func (h *GoalIt) GetCategory() string    { return "automation" }
func (h *GoalIt) GetDescription() string { return "Achieve a campaign/automation goal for a contact" }
func (h *GoalIt) RequiresCRM() bool      { return true }
func (h *GoalIt) SupportedCRMs() []string { return []string{"keap"} }

func (h *GoalIt) GetConfigSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"goal_name": map[string]interface{}{
				"type":        "string",
				"description": "The name/call_name of the goal to achieve",
			},
			"integration": map[string]interface{}{
				"type":        "string",
				"description": "The integration name (defaults to 'mfh')",
				"default":     "mfh",
			},
		},
		"required": []string{"goal_name"},
	}
}

func (h *GoalIt) ValidateConfig(config map[string]interface{}) error {
	if _, ok := config["goal_name"].(string); !ok || config["goal_name"] == "" {
		return fmt.Errorf("goal_name is required")
	}
	return nil
}

func (h *GoalIt) Execute(ctx context.Context, input helpers.HelperInput) (*helpers.HelperOutput, error) {
	goalName := input.Config["goal_name"].(string)

	integration := "mfh"
	if i, ok := input.Config["integration"].(string); ok && i != "" {
		integration = i
	}

	output := &helpers.HelperOutput{
		Logs: make([]string, 0),
	}

	err := input.Connector.AchieveGoal(ctx, input.ContactID, goalName, integration)
	if err != nil {
		output.Message = fmt.Sprintf("Failed to achieve goal '%s': %v", goalName, err)
		output.Logs = append(output.Logs, output.Message)
		return output, err
	}

	output.Success = true
	output.Message = fmt.Sprintf("Goal '%s' achieved for contact %s", goalName, input.ContactID)
	output.Actions = []helpers.HelperAction{
		{
			Type:   "goal_achieved",
			Target: input.ContactID,
			Value:  goalName,
		},
	}
	output.Logs = append(output.Logs, fmt.Sprintf("Achieved goal '%s' (integration: %s) for contact %s", goalName, integration, input.ContactID))

	return output, nil
}
