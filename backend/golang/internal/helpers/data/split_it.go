package data

import (
	"context"
	"fmt"

	"github.com/myfusionhelper/api/internal/helpers"
)

func init() {
	helpers.Register("split_it", func() helpers.Helper { return &SplitIt{} })
}

// SplitIt performs A/B split testing by alternating between two options
type SplitIt struct{}

func (h *SplitIt) GetName() string        { return "Split It" }
func (h *SplitIt) GetType() string        { return "split_it" }
func (h *SplitIt) GetCategory() string    { return "data" }
func (h *SplitIt) GetDescription() string { return "A/B split testing - alternate between two options on each execution" }
func (h *SplitIt) RequiresCRM() bool      { return true }
func (h *SplitIt) SupportedCRMs() []string { return nil }

func (h *SplitIt) GetConfigSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"mode": map[string]interface{}{
				"type":        "string",
				"enum":        []string{"tag", "goal"},
				"description": "Whether to apply tags or achieve goals for the split",
			},
			"option_a": map[string]interface{}{
				"type":        "string",
				"description": "Tag ID or goal name for option A",
			},
			"option_b": map[string]interface{}{
				"type":        "string",
				"description": "Tag ID or goal name for option B",
			},
			"state_field": map[string]interface{}{
				"type":        "string",
				"description": "Field to store the last choice (A or B) for alternation",
			},
		},
		"required": []string{"mode", "option_a", "option_b", "state_field"},
	}
}

func (h *SplitIt) ValidateConfig(config map[string]interface{}) error {
	mode, ok := config["mode"].(string)
	if !ok || mode == "" {
		return fmt.Errorf("mode is required")
	}
	if mode != "tag" && mode != "goal" {
		return fmt.Errorf("mode must be 'tag' or 'goal'")
	}
	if _, ok := config["option_a"].(string); !ok || config["option_a"] == "" {
		return fmt.Errorf("option_a is required")
	}
	if _, ok := config["option_b"].(string); !ok || config["option_b"] == "" {
		return fmt.Errorf("option_b is required")
	}
	if _, ok := config["state_field"].(string); !ok || config["state_field"] == "" {
		return fmt.Errorf("state_field is required")
	}
	return nil
}

func (h *SplitIt) Execute(ctx context.Context, input helpers.HelperInput) (*helpers.HelperOutput, error) {
	mode := input.Config["mode"].(string)
	optionA := input.Config["option_a"].(string)
	optionB := input.Config["option_b"].(string)
	stateField := input.Config["state_field"].(string)

	output := &helpers.HelperOutput{
		Actions:      make([]helpers.HelperAction, 0),
		ModifiedData: make(map[string]interface{}),
		Logs:         make([]string, 0),
	}

	// Read current state to determine which option to use
	lastChoice := ""
	stateValue, err := input.Connector.GetContactFieldValue(ctx, input.ContactID, stateField)
	if err == nil && stateValue != nil {
		strVal := fmt.Sprintf("%v", stateValue)
		if strVal != "<nil>" {
			lastChoice = strVal
		}
	}

	// Alternate: if last was A, use B; otherwise use A
	currentChoice := "A"
	currentOption := optionA
	if lastChoice == "A" {
		currentChoice = "B"
		currentOption = optionB
	}

	// Apply the selected option
	switch mode {
	case "tag":
		err = input.Connector.ApplyTag(ctx, input.ContactID, currentOption)
		if err != nil {
			output.Message = fmt.Sprintf("Failed to apply split tag %s: %v", currentOption, err)
			return output, err
		}
		output.Actions = append(output.Actions, helpers.HelperAction{
			Type:   "tag_applied",
			Target: input.ContactID,
			Value:  currentOption,
		})

	case "goal":
		err = input.Connector.AchieveGoal(ctx, input.ContactID, currentOption, "mfh")
		if err != nil {
			output.Message = fmt.Sprintf("Failed to achieve split goal '%s': %v", currentOption, err)
			return output, err
		}
		output.Actions = append(output.Actions, helpers.HelperAction{
			Type:   "goal_achieved",
			Target: input.ContactID,
			Value:  currentOption,
		})
	}

	// Update state field
	err = input.Connector.SetContactFieldValue(ctx, input.ContactID, stateField, currentChoice)
	if err != nil {
		output.Logs = append(output.Logs, fmt.Sprintf("Warning: failed to update state field '%s': %v", stateField, err))
	} else {
		output.Actions = append(output.Actions, helpers.HelperAction{
			Type:   "field_updated",
			Target: stateField,
			Value:  currentChoice,
		})
		output.ModifiedData[stateField] = currentChoice
	}

	output.Success = true
	output.Message = fmt.Sprintf("Split test: selected option %s (%s mode)", currentChoice, mode)
	output.Logs = append(output.Logs, fmt.Sprintf("A/B split on contact %s: chose %s (option: %s, mode: %s)", input.ContactID, currentChoice, currentOption, mode))

	return output, nil
}
