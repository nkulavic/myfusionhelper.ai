package contact

import (
	"context"
	"fmt"

	"github.com/myfusionhelper/api/internal/helpers"
)

func init() {
	helpers.Register("found_it", func() helpers.Helper { return &FoundIt{} })
}

// FoundIt checks if a field has a value and branches accordingly (applies tags/goals)
type FoundIt struct{}

func (h *FoundIt) GetName() string        { return "Found It" }
func (h *FoundIt) GetType() string        { return "found_it" }
func (h *FoundIt) GetCategory() string    { return "contact" }
func (h *FoundIt) GetDescription() string { return "Check if a field has a value and branch accordingly with tags or goals" }
func (h *FoundIt) RequiresCRM() bool      { return true }
func (h *FoundIt) SupportedCRMs() []string { return nil }

func (h *FoundIt) GetConfigSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"check_field": map[string]interface{}{
				"type":        "string",
				"description": "The field key to check for a value",
			},
			"found_tag_id": map[string]interface{}{
				"type":        "string",
				"description": "Tag ID to apply if field has a value",
			},
			"not_found_tag_id": map[string]interface{}{
				"type":        "string",
				"description": "Tag ID to apply if field is empty",
			},
			"found_goal": map[string]interface{}{
				"type":        "string",
				"description": "Goal name to achieve if field has a value",
			},
			"not_found_goal": map[string]interface{}{
				"type":        "string",
				"description": "Goal name to achieve if field is empty",
			},
		},
		"required": []string{"check_field"},
	}
}

func (h *FoundIt) ValidateConfig(config map[string]interface{}) error {
	if _, ok := config["check_field"].(string); !ok || config["check_field"] == "" {
		return fmt.Errorf("check_field is required")
	}
	return nil
}

func (h *FoundIt) Execute(ctx context.Context, input helpers.HelperInput) (*helpers.HelperOutput, error) {
	checkField := input.Config["check_field"].(string)

	output := &helpers.HelperOutput{
		Actions: make([]helpers.HelperAction, 0),
		Logs:    make([]string, 0),
	}

	// Get the field value
	value, err := input.Connector.GetContactFieldValue(ctx, input.ContactID, checkField)
	if err != nil {
		output.Message = fmt.Sprintf("Failed to read field '%s': %v", checkField, err)
		return output, err
	}

	strVal := fmt.Sprintf("%v", value)
	found := value != nil && strVal != "" && strVal != "<nil>"

	if found {
		output.Logs = append(output.Logs, fmt.Sprintf("Field '%s' has value: %s", checkField, strVal))

		// Apply found tag
		if tagID, ok := input.Config["found_tag_id"].(string); ok && tagID != "" {
			err := input.Connector.ApplyTag(ctx, input.ContactID, tagID)
			if err != nil {
				output.Logs = append(output.Logs, fmt.Sprintf("Failed to apply found tag %s: %v", tagID, err))
			} else {
				output.Actions = append(output.Actions, helpers.HelperAction{
					Type:   "tag_applied",
					Target: input.ContactID,
					Value:  tagID,
				})
			}
		}

		// Achieve found goal
		if goal, ok := input.Config["found_goal"].(string); ok && goal != "" {
			err := input.Connector.AchieveGoal(ctx, input.ContactID, goal, "mfh")
			if err != nil {
				output.Logs = append(output.Logs, fmt.Sprintf("Failed to achieve found goal '%s': %v", goal, err))
			} else {
				output.Actions = append(output.Actions, helpers.HelperAction{
					Type:   "goal_achieved",
					Target: input.ContactID,
					Value:  goal,
				})
			}
		}
	} else {
		output.Logs = append(output.Logs, fmt.Sprintf("Field '%s' is empty", checkField))

		// Apply not_found tag
		if tagID, ok := input.Config["not_found_tag_id"].(string); ok && tagID != "" {
			err := input.Connector.ApplyTag(ctx, input.ContactID, tagID)
			if err != nil {
				output.Logs = append(output.Logs, fmt.Sprintf("Failed to apply not_found tag %s: %v", tagID, err))
			} else {
				output.Actions = append(output.Actions, helpers.HelperAction{
					Type:   "tag_applied",
					Target: input.ContactID,
					Value:  tagID,
				})
			}
		}

		// Achieve not_found goal
		if goal, ok := input.Config["not_found_goal"].(string); ok && goal != "" {
			err := input.Connector.AchieveGoal(ctx, input.ContactID, goal, "mfh")
			if err != nil {
				output.Logs = append(output.Logs, fmt.Sprintf("Failed to achieve not_found goal '%s': %v", goal, err))
			} else {
				output.Actions = append(output.Actions, helpers.HelperAction{
					Type:   "goal_achieved",
					Target: input.ContactID,
					Value:  goal,
				})
			}
		}
	}

	output.Success = true
	if found {
		output.Message = fmt.Sprintf("Field '%s' has a value", checkField)
	} else {
		output.Message = fmt.Sprintf("Field '%s' is empty", checkField)
	}

	return output, nil
}
