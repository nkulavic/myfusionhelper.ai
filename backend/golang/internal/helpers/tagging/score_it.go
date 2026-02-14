package tagging

import (
	"context"
	"fmt"

	"github.com/myfusionhelper/api/internal/helpers"
)

func init() {
	helpers.Register("score_it", func() helpers.Helper { return &ScoreIt{} })
}

// NewScoreIt creates a new ScoreIt helper instance
func NewScoreIt() helpers.Helper { return &ScoreIt{} }

// ScoreIt scores contacts based on tag criteria rules
type ScoreIt struct{}

func (h *ScoreIt) GetName() string        { return "Score It" }
func (h *ScoreIt) GetType() string        { return "score_it" }
func (h *ScoreIt) GetCategory() string    { return "tagging" }
func (h *ScoreIt) GetDescription() string { return "Score contacts based on tag criteria and store the result" }
func (h *ScoreIt) RequiresCRM() bool      { return true }
func (h *ScoreIt) SupportedCRMs() []string { return nil }

func (h *ScoreIt) GetConfigSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"rules": map[string]interface{}{
				"type": "array",
				"items": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"tag_id":  map[string]interface{}{"type": "string"},
						"has_tag": map[string]interface{}{"type": "boolean"},
						"points":  map[string]interface{}{"type": "integer"},
					},
				},
				"description": "Array of scoring rules: {tag_id, has_tag, points}",
			},
			"target_field": map[string]interface{}{
				"type":        "string",
				"description": "The field to store the total score",
			},
		},
		"required": []string{"rules", "target_field"},
	}
}

func (h *ScoreIt) ValidateConfig(config map[string]interface{}) error {
	rules, ok := config["rules"]
	if !ok {
		return fmt.Errorf("rules is required")
	}

	rulesList, ok := rules.([]interface{})
	if !ok {
		return fmt.Errorf("rules must be an array")
	}
	if len(rulesList) == 0 {
		return fmt.Errorf("rules must contain at least one rule")
	}

	for i, r := range rulesList {
		rule, ok := r.(map[string]interface{})
		if !ok {
			return fmt.Errorf("rule %d must be an object", i)
		}
		if _, ok := rule["tag_id"].(string); !ok {
			return fmt.Errorf("rule %d requires a tag_id", i)
		}
	}

	if _, ok := config["target_field"].(string); !ok || config["target_field"] == "" {
		return fmt.Errorf("target_field is required")
	}

	return nil
}

func (h *ScoreIt) Execute(ctx context.Context, input helpers.HelperInput) (*helpers.HelperOutput, error) {
	rulesList := input.Config["rules"].([]interface{})
	targetField := input.Config["target_field"].(string)

	output := &helpers.HelperOutput{
		Logs: make([]string, 0),
	}

	// Get contact to find current tags
	contact, err := input.Connector.GetContact(ctx, input.ContactID)
	if err != nil {
		output.Message = fmt.Sprintf("Failed to get contact: %v", err)
		return output, err
	}

	// Build set of contact tag IDs for fast lookup
	contactTags := make(map[string]bool)
	for _, tag := range contact.Tags {
		contactTags[tag.ID] = true
	}

	// Evaluate rules and sum points
	totalScore := 0
	rulesMatched := 0

	for _, r := range rulesList {
		rule := r.(map[string]interface{})
		tagID := rule["tag_id"].(string)

		// Default: has_tag = true
		hasTag := true
		if ht, ok := rule["has_tag"].(bool); ok {
			hasTag = ht
		}

		// Default: points = 0
		points := 0
		if p, ok := rule["points"].(float64); ok {
			points = int(p)
		}

		// Check if rule matches
		contactHasTag := contactTags[tagID]
		if contactHasTag == hasTag {
			totalScore += points
			rulesMatched++
			output.Logs = append(output.Logs, fmt.Sprintf("Rule matched: tag %s (has_tag=%v), +%d points", tagID, hasTag, points))
		}
	}

	scoreStr := fmt.Sprintf("%d", totalScore)

	// Set the score field
	err = input.Connector.SetContactFieldValue(ctx, input.ContactID, targetField, scoreStr)
	if err != nil {
		output.Message = fmt.Sprintf("Failed to set score field '%s': %v", targetField, err)
		return output, err
	}

	output.Success = true
	output.Message = fmt.Sprintf("Scored contact: %d points (%d of %d rules matched)", totalScore, rulesMatched, len(rulesList))
	output.Actions = []helpers.HelperAction{
		{
			Type:   "field_updated",
			Target: targetField,
			Value:  scoreStr,
		},
	}
	output.ModifiedData = map[string]interface{}{
		targetField: scoreStr,
	}
	output.Logs = append(output.Logs, fmt.Sprintf("Score for contact %s: %d points, stored in '%s'", input.ContactID, totalScore, targetField))

	return output, nil
}
