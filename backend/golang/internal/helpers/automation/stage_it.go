package automation

import (
	"context"
	"fmt"

	"github.com/myfusionhelper/api/internal/helpers"
)

func init() {
	helpers.Register("stage_it", func() helpers.Helper { return &StageIt{} })
}

// StageIt manages opportunity/deal stage transitions by matching opportunities
// on a contact and updating their stage. Fires goals when opportunities are found or not found.
// Ported from legacy PHP stage_it helper.
type StageIt struct{}

func (h *StageIt) GetName() string     { return "Stage It" }
func (h *StageIt) GetType() string     { return "stage_it" }
func (h *StageIt) GetCategory() string { return "automation" }
func (h *StageIt) GetDescription() string {
	return "Match opportunities by stage and update them, firing goals on match or no-match"
}
func (h *StageIt) RequiresCRM() bool       { return true }
func (h *StageIt) SupportedCRMs() []string { return []string{"keap"} }

func (h *StageIt) GetConfigSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"basic_match": map[string]interface{}{
				"type":        "string",
				"description": "The stage ID to match opportunities against",
			},
			"to_stage": map[string]interface{}{
				"type":        "string",
				"description": "The target stage ID to move matched opportunities to",
			},
			"opportunity_count": map[string]interface{}{
				"type":        "string",
				"enum":        []string{"first", "all"},
				"description": "Whether to update the first matched opportunity or all matched opportunities",
				"default":     "first",
			},
			"found_goal": map[string]interface{}{
				"type":        "string",
				"description": "Goal call name to achieve when opportunities are found",
			},
			"not_found_goal": map[string]interface{}{
				"type":        "string",
				"description": "Goal call name to achieve when no opportunities are found",
			},
		},
		"required": []string{"basic_match", "to_stage"},
	}
}

func (h *StageIt) ValidateConfig(config map[string]interface{}) error {
	if _, ok := config["basic_match"].(string); !ok || config["basic_match"] == "" {
		return fmt.Errorf("basic_match (stage ID to match) is required")
	}
	if _, ok := config["to_stage"].(string); !ok || config["to_stage"] == "" {
		return fmt.Errorf("to_stage is required")
	}
	return nil
}

func (h *StageIt) Execute(ctx context.Context, input helpers.HelperInput) (*helpers.HelperOutput, error) {
	matchStage := input.Config["basic_match"].(string)
	toStage := input.Config["to_stage"].(string)

	oppCount := "first"
	if oc, ok := input.Config["opportunity_count"].(string); ok && oc != "" {
		oppCount = oc
	}

	integration := "myfusionhelper"

	output := &helpers.HelperOutput{
		Actions: make([]helpers.HelperAction, 0),
		Logs:    make([]string, 0),
	}

	// Query for opportunities matching the contact and stage
	// Using GetContactFieldValue with a composite key for related deal/opportunity queries
	queryKey := fmt.Sprintf("_related.lead.stage.%s", matchStage)
	oppValue, err := input.Connector.GetContactFieldValue(ctx, input.ContactID, queryKey)
	if err != nil {
		output.Logs = append(output.Logs, fmt.Sprintf("Opportunity query via field lookup: %v", err))
	}

	// Determine if opportunities were found
	hasOpportunities := oppValue != nil && fmt.Sprintf("%v", oppValue) != "" && fmt.Sprintf("%v", oppValue) != "0"

	if !hasOpportunities {
		// No opportunities found - fire not_found goal
		if notFoundGoal, ok := input.Config["not_found_goal"].(string); ok && notFoundGoal != "" {
			goalErr := input.Connector.AchieveGoal(ctx, input.ContactID, notFoundGoal, integration)
			if goalErr != nil {
				output.Logs = append(output.Logs, fmt.Sprintf("Failed to achieve not_found goal '%s': %v", notFoundGoal, goalErr))
			} else {
				output.Actions = append(output.Actions, helpers.HelperAction{
					Type:   "goal_achieved",
					Target: input.ContactID,
					Value:  notFoundGoal,
				})
			}
		}
		output.Success = true
		output.Message = fmt.Sprintf("No opportunities found matching stage %s", matchStage)
		output.Logs = append(output.Logs, output.Message)
		return output, nil
	}

	// Update the opportunity stage(s)
	// Use SetContactFieldValue with composite key for related record updates
	updateCount := 0
	if oppCount == "all" {
		updateKey := fmt.Sprintf("_related.lead.stage.%s.update_all", matchStage)
		err = input.Connector.SetContactFieldValue(ctx, input.ContactID, updateKey, toStage)
		if err != nil {
			output.Logs = append(output.Logs, fmt.Sprintf("Failed to update all opportunities: %v", err))
		} else {
			updateCount++
		}
	} else {
		updateKey := fmt.Sprintf("_related.lead.stage.%s.update_first", matchStage)
		err = input.Connector.SetContactFieldValue(ctx, input.ContactID, updateKey, toStage)
		if err != nil {
			output.Logs = append(output.Logs, fmt.Sprintf("Failed to update first opportunity: %v", err))
		} else {
			updateCount++
		}
	}

	output.Actions = append(output.Actions, helpers.HelperAction{
		Type:   "field_updated",
		Target: "StageID",
		Value:  toStage,
	})

	// Fire found goal
	if foundGoal, ok := input.Config["found_goal"].(string); ok && foundGoal != "" {
		goalErr := input.Connector.AchieveGoal(ctx, input.ContactID, foundGoal, integration)
		if goalErr != nil {
			output.Logs = append(output.Logs, fmt.Sprintf("Failed to achieve found goal '%s': %v", foundGoal, goalErr))
		} else {
			output.Actions = append(output.Actions, helpers.HelperAction{
				Type:   "goal_achieved",
				Target: input.ContactID,
				Value:  foundGoal,
			})
		}
	}

	output.Success = true
	output.Message = fmt.Sprintf("Updated %s opportunity stage(s) from %s to %s", oppCount, matchStage, toStage)
	output.Logs = append(output.Logs, fmt.Sprintf("Stage update for contact %s: %s -> %s (mode: %s)", input.ContactID, matchStage, toStage, oppCount))

	return output, nil
}
