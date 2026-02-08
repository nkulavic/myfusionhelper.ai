package integration

import (
	"context"
	"fmt"

	"github.com/myfusionhelper/api/internal/helpers"
)

func init() {
	helpers.Register("hook_it_v4", func() helpers.Helper { return &HookItV4{} })
}

// HookItV4 handles webhook events with multi-goal achievement.
// Fires multiple goals sequentially for each webhook event.
// Example: goals: ["goal1", "goal2", "goal3"] - achieves all goals in order
type HookItV4 struct{}

func (h *HookItV4) GetName() string     { return "Hook It V4 (Multi-Goal)" }
func (h *HookItV4) GetType() string     { return "hook_it_v4" }
func (h *HookItV4) GetCategory() string { return "integration" }
func (h *HookItV4) GetDescription() string {
	return "Fire multiple goals sequentially for each webhook event"
}
func (h *HookItV4) RequiresCRM() bool       { return true }
func (h *HookItV4) SupportedCRMs() []string { return nil }

func (h *HookItV4) GetConfigSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"goals": map[string]interface{}{
				"type": "array",
				"items": map[string]interface{}{
					"type": "string",
				},
				"description": "Array of goal names to achieve sequentially (e.g., [\"goal1\", \"goal2\", \"goal3\"])",
			},
			"integration": map[string]interface{}{
				"type":        "string",
				"description": "Integration name for goal calls (defaults to 'myfusionhelper')",
			},
			"stop_on_error": map[string]interface{}{
				"type":        "boolean",
				"description": "Stop processing remaining goals if one fails (defaults to false)",
			},
			"goal_prefix": map[string]interface{}{
				"type":        "string",
				"description": "Prefix to prepend to all goal names (e.g., 'webhook_')",
			},
			"goal_suffix": map[string]interface{}{
				"type":        "string",
				"description": "Suffix to append to all goal names (e.g., '_event')",
			},
		},
		"required": []string{"goals"},
	}
}

func (h *HookItV4) ValidateConfig(config map[string]interface{}) error {
	if config["goals"] == nil {
		return fmt.Errorf("goals array is required")
	}

	goalsInterface, ok := config["goals"].([]interface{})
	if !ok {
		return fmt.Errorf("goals must be an array")
	}

	if len(goalsInterface) == 0 {
		return fmt.Errorf("goals array must contain at least one goal name")
	}

	// Validate that all goals are strings
	for i, goalInterface := range goalsInterface {
		if _, ok := goalInterface.(string); !ok {
			return fmt.Errorf("goals[%d] must be a string goal name", i)
		}
	}

	return nil
}

func (h *HookItV4) Execute(ctx context.Context, input helpers.HelperInput) (*helpers.HelperOutput, error) {
	output := &helpers.HelperOutput{
		Actions: make([]helpers.HelperAction, 0),
		Logs:    make([]string, 0),
	}

	// Extract goals array from config
	goalsInterface, ok := input.Config["goals"].([]interface{})
	if !ok {
		output.Message = "Invalid goals configuration"
		output.Logs = append(output.Logs, output.Message)
		return output, fmt.Errorf("goals must be an array")
	}

	// Get integration name (defaults to "myfusionhelper")
	integration := "myfusionhelper"
	if intg, ok := input.Config["integration"].(string); ok && intg != "" {
		integration = intg
	}

	// Get stop_on_error flag (defaults to false)
	stopOnError := false
	if stop, ok := input.Config["stop_on_error"].(bool); ok {
		stopOnError = stop
	}

	// Get goal prefix/suffix
	goalPrefix := ""
	if prefix, ok := input.Config["goal_prefix"].(string); ok {
		goalPrefix = prefix
	}

	goalSuffix := ""
	if suffix, ok := input.Config["goal_suffix"].(string); ok {
		goalSuffix = suffix
	}

	// Process each goal sequentially
	goalsAchieved := 0
	goalsFailed := 0
	for i, goalInterface := range goalsInterface {
		goalName, ok := goalInterface.(string)
		if !ok {
			output.Logs = append(output.Logs, fmt.Sprintf("Skipping invalid goal at index %d", i))
			continue
		}

		if goalName == "" {
			output.Logs = append(output.Logs, fmt.Sprintf("Skipping empty goal name at index %d", i))
			continue
		}

		// Apply prefix and suffix
		fullGoalName := goalPrefix + goalName + goalSuffix

		// Achieve the goal
		err := input.Connector.AchieveGoal(ctx, input.ContactID, fullGoalName, integration)
		if err != nil {
			goalsFailed++
			output.Logs = append(output.Logs, fmt.Sprintf("Failed to achieve goal '%s' (#%d): %v", fullGoalName, i+1, err))

			// Stop processing if configured to do so
			if stopOnError {
				output.Success = false
				output.Message = fmt.Sprintf("Multi-goal processing stopped at goal '%s' due to error: %v", fullGoalName, err)
				output.ModifiedData = map[string]interface{}{
					"goals_total":    len(goalsInterface),
					"goals_achieved": goalsAchieved,
					"goals_failed":   goalsFailed,
					"stopped_at":     i + 1,
				}
				return output, err
			}
			continue
		}

		goalsAchieved++
		output.Actions = append(output.Actions, helpers.HelperAction{
			Type:   "goal_achieved",
			Target: input.ContactID,
			Value:  fullGoalName,
		})
		output.Logs = append(output.Logs, fmt.Sprintf("Achieved goal '%s' (#%d of %d)", fullGoalName, i+1, len(goalsInterface)))
	}

	output.Success = goalsAchieved > 0 || goalsFailed == 0
	output.Message = fmt.Sprintf("Multi-goal processing complete: %d achieved, %d failed out of %d total", goalsAchieved, goalsFailed, len(goalsInterface))
	output.ModifiedData = map[string]interface{}{
		"goals_total":    len(goalsInterface),
		"goals_achieved": goalsAchieved,
		"goals_failed":   goalsFailed,
	}

	return output, nil
}
