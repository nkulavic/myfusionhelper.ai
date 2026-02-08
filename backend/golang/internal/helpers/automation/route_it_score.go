package automation

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/myfusionhelper/api/internal/helpers"
)

func init() {
	helpers.Register("route_it_score", func() helpers.Helper { return &RouteItScore{} })
}

// RouteItScore routes contacts based on lead score ranges
type RouteItScore struct{}

func (h *RouteItScore) GetName() string     { return "Route It - Score" }
func (h *RouteItScore) GetType() string     { return "route_it_score" }
func (h *RouteItScore) GetCategory() string { return "automation" }
func (h *RouteItScore) GetDescription() string {
	return "Route contacts to different URLs based on lead score ranges"
}
func (h *RouteItScore) RequiresCRM() bool       { return true }
func (h *RouteItScore) SupportedCRMs() []string { return nil }

func (h *RouteItScore) GetConfigSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"score_field": map[string]interface{}{
				"type":        "string",
				"description": "Contact field containing the lead score (numeric)",
				"default":     "Score",
			},
			"score_ranges": map[string]interface{}{
				"type":        "array",
				"description": "Array of score range configurations (evaluated in order)",
				"items": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"label": map[string]interface{}{
							"type":        "string",
							"description": "Label for this score range (e.g., 'Hot Lead', 'Warm Lead')",
						},
						"min_score": map[string]interface{}{
							"type":        "number",
							"description": "Minimum score (inclusive)",
						},
						"max_score": map[string]interface{}{
							"type":        "number",
							"description": "Maximum score (inclusive). Omit for open-ended range.",
						},
						"redirect_url": map[string]interface{}{
							"type":        "string",
							"description": "URL to redirect to if score is in this range",
						},
					},
					"required": []string{"redirect_url"},
				},
			},
			"fallback_url": map[string]interface{}{
				"type":        "string",
				"description": "Default URL if no score ranges match",
			},
			"save_to_field": map[string]interface{}{
				"type":        "string",
				"description": "Optional: CRM field to save the selected URL to",
			},
			"apply_tag": map[string]interface{}{
				"type":        "string",
				"description": "Optional: Tag ID to apply when routing occurs",
			},
		},
		"required": []string{"score_ranges"},
	}
}

func (h *RouteItScore) ValidateConfig(config map[string]interface{}) error {
	scoreRanges, ok := config["score_ranges"].([]interface{})
	if !ok {
		return fmt.Errorf("score_ranges is required and must be an array")
	}
	if len(scoreRanges) == 0 {
		return fmt.Errorf("score_ranges must contain at least one range")
	}

	// Validate each range
	for i, rangeInterface := range scoreRanges {
		rangeMap, ok := rangeInterface.(map[string]interface{})
		if !ok {
			return fmt.Errorf("score_ranges[%d] must be an object", i)
		}

		redirectURL, ok := rangeMap["redirect_url"].(string)
		if !ok || redirectURL == "" {
			return fmt.Errorf("score_ranges[%d].redirect_url is required", i)
		}
	}

	return nil
}

func (h *RouteItScore) Execute(ctx context.Context, input helpers.HelperInput) (*helpers.HelperOutput, error) {
	scoreField := "Score"
	if sf, ok := input.Config["score_field"].(string); ok && sf != "" {
		scoreField = sf
	}

	scoreRanges := input.Config["score_ranges"].([]interface{})
	fallbackURL, _ := input.Config["fallback_url"].(string)
	saveToField, _ := input.Config["save_to_field"].(string)
	applyTag, _ := input.Config["apply_tag"].(string)

	output := &helpers.HelperOutput{
		Actions: make([]helpers.HelperAction, 0),
		Logs:    make([]string, 0),
	}

	// Get the score value from the contact
	scoreValue, err := input.Connector.GetContactFieldValue(ctx, input.ContactID, scoreField)
	if err != nil {
		output.Message = fmt.Sprintf("Failed to get score field '%s': %v", scoreField, err)
		return output, err
	}

	// Convert to float64
	score, err := toFloat64Score(scoreValue)
	if err != nil {
		output.Message = fmt.Sprintf("Score field '%s' is not a valid number: %v", scoreField, err)
		return output, err
	}

	output.Logs = append(output.Logs, fmt.Sprintf("Contact score: %s = %.2f", scoreField, score))

	// Evaluate score ranges in order
	var selectedURL string
	var routingReason string
	var matchedLabel string

	for i, rangeInterface := range scoreRanges {
		rangeMap := rangeInterface.(map[string]interface{})
		redirectURL := rangeMap["redirect_url"].(string)
		label, _ := rangeMap["label"].(string)

		// Get min/max scores
		var minScore, maxScore float64
		var hasMin, hasMax bool

		if min, ok := rangeMap["min_score"].(float64); ok {
			minScore = min
			hasMin = true
		}
		if max, ok := rangeMap["max_score"].(float64); ok {
			maxScore = max
			hasMax = true
		}

		// Check if score is in range
		inRange := true
		if hasMin && score < minScore {
			inRange = false
		}
		if hasMax && score > maxScore {
			inRange = false
		}

		if inRange {
			selectedURL = redirectURL
			matchedLabel = label
			if label != "" {
				routingReason = fmt.Sprintf("score_range=%s (%.2f)", label, score)
			} else {
				routingReason = fmt.Sprintf("score_range_index=%d (%.2f)", i, score)
			}
			output.Logs = append(output.Logs, fmt.Sprintf("Score %.2f matched range: %s", score, redirectURL))
			break
		}
	}

	// Fall back if no range matched
	if selectedURL == "" {
		if fallbackURL != "" {
			selectedURL = fallbackURL
			routingReason = fmt.Sprintf("fallback (score=%.2f, no match)", score)
			output.Logs = append(output.Logs, "No score range matched, using fallback URL")
		} else {
			output.Message = fmt.Sprintf("No score range matched for score %.2f and no fallback URL configured", score)
			return output, fmt.Errorf("no route found")
		}
	}

	// Optional: Save URL to CRM field
	if saveToField != "" {
		err := input.Connector.SetContactFieldValue(ctx, input.ContactID, saveToField, selectedURL)
		if err != nil {
			output.Logs = append(output.Logs, fmt.Sprintf("Warning: Failed to save URL to field '%s': %v", saveToField, err))
		} else {
			output.Actions = append(output.Actions, helpers.HelperAction{
				Type:   "field_updated",
				Target: saveToField,
				Value:  selectedURL,
			})
			output.Logs = append(output.Logs, fmt.Sprintf("Saved URL to field '%s'", saveToField))
		}
	}

	// Optional: Apply tag
	if applyTag != "" {
		err := input.Connector.ApplyTag(ctx, input.ContactID, applyTag)
		if err != nil {
			output.Logs = append(output.Logs, fmt.Sprintf("Warning: Failed to apply tag '%s': %v", applyTag, err))
		} else {
			output.Actions = append(output.Actions, helpers.HelperAction{
				Type:   "tag_applied",
				Target: input.ContactID,
				Value:  applyTag,
			})
			output.Logs = append(output.Logs, fmt.Sprintf("Applied tag '%s'", applyTag))
		}
	}

	output.Success = true
	output.Message = fmt.Sprintf("Routed to %s based on score %.2f", selectedURL, score)
	output.ModifiedData = map[string]interface{}{
		"redirect_url":   selectedURL,
		"routing_reason": routingReason,
		"score":          score,
		"matched_label":  matchedLabel,
	}
	output.Actions = append(output.Actions, helpers.HelperAction{
		Type:   "contact_routed",
		Target: input.ContactID,
		Value:  selectedURL,
	})

	return output, nil
}

// toFloat64Score converts various types to float64 for score comparison
func toFloat64Score(val interface{}) (float64, error) {
	switch v := val.(type) {
	case float64:
		return v, nil
	case float32:
		return float64(v), nil
	case int:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case string:
		// Try parsing string as number
		if strings.TrimSpace(v) == "" {
			return 0, nil
		}
		f, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return 0, fmt.Errorf("cannot parse '%s' as number", v)
		}
		return f, nil
	default:
		return 0, fmt.Errorf("cannot convert %T to float64", val)
	}
}
