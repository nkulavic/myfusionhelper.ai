package data

import (
	"context"
	"fmt"

	"github.com/myfusionhelper/api/internal/helpers"
)

func init() {
	helpers.Register("split_it_basic", func() helpers.Helper { return &SplitItBasic{} })
}

// SplitItBasic performs basic A/B split testing with tag or goal routing.
// It alternates between group A and group B based on the last_group config value,
// then applies a tag or achieves a goal for the selected group.
// Ported from legacy PHP split_it_basic helper.
type SplitItBasic struct{}

func (h *SplitItBasic) GetName() string     { return "Split It Basic" }
func (h *SplitItBasic) GetType() string     { return "split_it_basic" }
func (h *SplitItBasic) GetCategory() string { return "data" }
func (h *SplitItBasic) GetDescription() string {
	return "Basic A/B split testing - alternate between two groups using tags or goals"
}
func (h *SplitItBasic) RequiresCRM() bool       { return true }
func (h *SplitItBasic) SupportedCRMs() []string { return nil }

func (h *SplitItBasic) GetConfigSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"split_type": map[string]interface{}{
				"type":        "string",
				"enum":        []string{"tag_split", "goal_split"},
				"description": "Whether to split using tags or goals",
			},
			"last_group": map[string]interface{}{
				"type":        "string",
				"enum":        []string{"a", "b"},
				"description": "The last group that was run (will alternate to the other)",
			},
			"split_tag_a": map[string]interface{}{
				"type":        "string",
				"description": "Tag ID for group A (used when split_type is tag_split)",
			},
			"split_tag_b": map[string]interface{}{
				"type":        "string",
				"description": "Tag ID for group B (used when split_type is tag_split)",
			},
			"split_goal_a": map[string]interface{}{
				"type":        "string",
				"description": "Goal call name for group A (used when split_type is goal_split)",
			},
			"split_goal_b": map[string]interface{}{
				"type":        "string",
				"description": "Goal call name for group B (used when split_type is goal_split)",
			},
		},
		"required": []string{"split_type", "last_group"},
	}
}

func (h *SplitItBasic) ValidateConfig(config map[string]interface{}) error {
	splitType, ok := config["split_type"].(string)
	if !ok || splitType == "" {
		return fmt.Errorf("split_type is required")
	}
	if splitType != "tag_split" && splitType != "goal_split" {
		return fmt.Errorf("split_type must be 'tag_split' or 'goal_split'")
	}

	lastGroup, ok := config["last_group"].(string)
	if !ok || lastGroup == "" {
		return fmt.Errorf("last_group is required")
	}
	if lastGroup != "a" && lastGroup != "b" {
		return fmt.Errorf("last_group must be 'a' or 'b'")
	}

	if splitType == "tag_split" {
		if _, ok := config["split_tag_a"].(string); !ok || config["split_tag_a"] == "" {
			return fmt.Errorf("split_tag_a is required for tag_split")
		}
		if _, ok := config["split_tag_b"].(string); !ok || config["split_tag_b"] == "" {
			return fmt.Errorf("split_tag_b is required for tag_split")
		}
	}
	if splitType == "goal_split" {
		if _, ok := config["split_goal_a"].(string); !ok || config["split_goal_a"] == "" {
			return fmt.Errorf("split_goal_a is required for goal_split")
		}
		if _, ok := config["split_goal_b"].(string); !ok || config["split_goal_b"] == "" {
			return fmt.Errorf("split_goal_b is required for goal_split")
		}
	}

	return nil
}

func (h *SplitItBasic) Execute(ctx context.Context, input helpers.HelperInput) (*helpers.HelperOutput, error) {
	splitType := input.Config["split_type"].(string)
	lastGroup := input.Config["last_group"].(string)

	output := &helpers.HelperOutput{
		Actions:      make([]helpers.HelperAction, 0),
		ModifiedData: make(map[string]interface{}),
		Logs:         make([]string, 0),
	}

	// Alternate: if last was B, run A; if last was A, run B
	runGroup := "a"
	if lastGroup == "a" {
		runGroup = "b"
	}

	output.Logs = append(output.Logs, fmt.Sprintf("Last group was '%s', running group '%s'", lastGroup, runGroup))

	switch splitType {
	case "tag_split":
		var tagID string
		if runGroup == "a" {
			tagID = input.Config["split_tag_a"].(string)
		} else {
			tagID = input.Config["split_tag_b"].(string)
		}

		// Remove then re-apply tag (ensures clean state)
		_ = input.Connector.RemoveTag(ctx, input.ContactID, tagID)
		err := input.Connector.ApplyTag(ctx, input.ContactID, tagID)
		if err != nil {
			output.Message = fmt.Sprintf("Failed to apply split tag %s: %v", tagID, err)
			return output, err
		}

		output.Actions = append(output.Actions, helpers.HelperAction{
			Type:   "tag_applied",
			Target: input.ContactID,
			Value:  tagID,
		})

	case "goal_split":
		var goalName string
		if runGroup == "a" {
			goalName = input.Config["split_goal_a"].(string)
		} else {
			goalName = input.Config["split_goal_b"].(string)
		}

		integration := "myfusionhelper"
		err := input.Connector.AchieveGoal(ctx, input.ContactID, goalName, integration)
		if err != nil {
			output.Message = fmt.Sprintf("Failed to achieve split goal '%s': %v", goalName, err)
			return output, err
		}

		output.Actions = append(output.Actions, helpers.HelperAction{
			Type:   "goal_achieved",
			Target: input.ContactID,
			Value:  goalName,
		})
	}

	output.Success = true
	output.Message = fmt.Sprintf("Split test: ran group %s (%s mode)", runGroup, splitType)
	output.ModifiedData["run_group"] = runGroup
	output.Logs = append(output.Logs, fmt.Sprintf("A/B split basic on contact %s: group %s (type: %s)", input.ContactID, runGroup, splitType))

	return output, nil
}
