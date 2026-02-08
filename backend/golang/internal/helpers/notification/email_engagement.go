package notification

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/myfusionhelper/api/internal/helpers"
)

func init() {
	helpers.Register("email_engagement", func() helpers.Helper { return &EmailEngagement{} })
}

// EmailEngagement tracks email open, click, and send events and
// applies tags or triggers automations based on engagement level.
// Primarily designed for Keap's email tracking data.
type EmailEngagement struct{}

func (h *EmailEngagement) GetName() string     { return "Email Engagement" }
func (h *EmailEngagement) GetType() string     { return "email_engagement" }
func (h *EmailEngagement) GetCategory() string { return "notification" }
func (h *EmailEngagement) GetDescription() string {
	return "Track email opens, clicks, and trigger automations based on engagement"
}
func (h *EmailEngagement) RequiresCRM() bool       { return true }
func (h *EmailEngagement) SupportedCRMs() []string { return []string{"keap"} }

func (h *EmailEngagement) GetConfigSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"engagement_type": map[string]interface{}{
				"type":        "string",
				"enum":        []string{"opens", "clicks", "sends", "all"},
				"description": "Type of email engagement to evaluate",
				"default":     "all",
			},
			"lookback_days": map[string]interface{}{
				"type":        "number",
				"description": "Number of days to look back for engagement data",
				"default":     30,
			},
			"time_decay_days": map[string]interface{}{
				"type":        "number",
				"description": "Number of days for time decay calculation (older engagements are weighted less)",
				"default":     90,
			},
			"engagement_weights": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"opens": map[string]interface{}{
						"type":        "number",
						"description": "Weight multiplier for email opens",
						"default":     1,
					},
					"clicks": map[string]interface{}{
						"type":        "number",
						"description": "Weight multiplier for email clicks",
						"default":     3,
					},
				},
			},
			"thresholds": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"highly_engaged": map[string]interface{}{
						"type":        "number",
						"description": "Minimum engagements to be considered highly engaged",
						"default":     5,
					},
					"engaged": map[string]interface{}{
						"type":        "number",
						"description": "Minimum engagements to be considered engaged",
						"default":     2,
					},
				},
			},
			"segments": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"hot_threshold": map[string]interface{}{
						"type":        "number",
						"description": "Minimum score for hot segment",
						"default":     10,
					},
					"warm_threshold": map[string]interface{}{
						"type":        "number",
						"description": "Minimum score for warm segment (hot_threshold > score >= warm_threshold)",
						"default":     5,
					},
				},
			},
			"trend_analysis": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"enabled": map[string]interface{}{
						"type":        "boolean",
						"description": "Enable trend detection (increasing/decreasing engagement)",
						"default":     false,
					},
					"comparison_days": map[string]interface{}{
						"type":        "number",
						"description": "Days to compare for trend analysis",
						"default":     30,
					},
				},
			},
			"tags": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"highly_engaged_tag": map[string]interface{}{
						"type":        "string",
						"description": "Tag ID to apply for highly engaged contacts",
					},
					"engaged_tag": map[string]interface{}{
						"type":        "string",
						"description": "Tag ID to apply for engaged contacts",
					},
					"disengaged_tag": map[string]interface{}{
						"type":        "string",
						"description": "Tag ID to apply for disengaged contacts",
					},
					"hot_segment_tag": map[string]interface{}{
						"type":        "string",
						"description": "Tag ID to apply for hot segment contacts",
					},
					"warm_segment_tag": map[string]interface{}{
						"type":        "string",
						"description": "Tag ID to apply for warm segment contacts",
					},
					"cold_segment_tag": map[string]interface{}{
						"type":        "string",
						"description": "Tag ID to apply for cold segment contacts",
					},
				},
			},
			"score_field": map[string]interface{}{
				"type":        "string",
				"description": "Custom field to store the calculated engagement score",
			},
			"last_engagement_field": map[string]interface{}{
				"type":        "string",
				"description": "Custom field to store the last engagement date",
			},
			"segment_field": map[string]interface{}{
				"type":        "string",
				"description": "Custom field to store the engagement segment (hot/warm/cold)",
			},
			"trend_field": map[string]interface{}{
				"type":        "string",
				"description": "Custom field to store the engagement trend (increasing/decreasing/stable)",
			},
		},
	}
}

func (h *EmailEngagement) ValidateConfig(config map[string]interface{}) error {
	if et, ok := config["engagement_type"].(string); ok {
		valid := map[string]bool{"opens": true, "clicks": true, "sends": true, "all": true}
		if !valid[et] {
			return fmt.Errorf("invalid engagement_type: %s", et)
		}
	}
	return nil
}

func (h *EmailEngagement) Execute(ctx context.Context, input helpers.HelperInput) (*helpers.HelperOutput, error) {
	output := &helpers.HelperOutput{
		Logs: make([]string, 0),
	}

	// Parse config
	engagementType := "all"
	if et, ok := input.Config["engagement_type"].(string); ok && et != "" {
		engagementType = et
	}

	lookbackDays := 30
	if lb, ok := input.Config["lookback_days"].(float64); ok && lb > 0 {
		lookbackDays = int(lb)
	}

	timeDecayDays := 90
	if td, ok := input.Config["time_decay_days"].(float64); ok && td > 0 {
		timeDecayDays = int(td)
	}

	opensWeight := 1.0
	clicksWeight := 3.0
	if weights, ok := input.Config["engagement_weights"].(map[string]interface{}); ok {
		if ow, ok := weights["opens"].(float64); ok {
			opensWeight = ow
		}
		if cw, ok := weights["clicks"].(float64); ok {
			clicksWeight = cw
		}
	}

	highlyEngagedThreshold := 5.0
	engagedThreshold := 2.0
	if thresholds, ok := input.Config["thresholds"].(map[string]interface{}); ok {
		if he, ok := thresholds["highly_engaged"].(float64); ok {
			highlyEngagedThreshold = he
		}
		if e, ok := thresholds["engaged"].(float64); ok {
			engagedThreshold = e
		}
	}

	hotThreshold := 10.0
	warmThreshold := 5.0
	if segments, ok := input.Config["segments"].(map[string]interface{}); ok {
		if ht, ok := segments["hot_threshold"].(float64); ok {
			hotThreshold = ht
		}
		if wt, ok := segments["warm_threshold"].(float64); ok {
			warmThreshold = wt
		}
	}

	trendEnabled := false
	trendComparisonDays := 30
	if trend, ok := input.Config["trend_analysis"].(map[string]interface{}); ok {
		if te, ok := trend["enabled"].(bool); ok {
			trendEnabled = te
		}
		if tcd, ok := trend["comparison_days"].(float64); ok && tcd > 0 {
			trendComparisonDays = int(tcd)
		}
	}

	// Read engagement tracking fields from the contact
	engagementScore := 0
	var lastEngagement time.Time

	fields := []string{}
	switch engagementType {
	case "opens":
		fields = []string{"last_open_date", "open_count"}
	case "clicks":
		fields = []string{"last_click_date", "click_count"}
	case "sends":
		fields = []string{"last_send_date", "send_count"}
	default: // "all"
		fields = []string{"last_open_date", "open_count", "last_click_date", "click_count", "last_send_date", "send_count"}
	}

	fieldValues := map[string]interface{}{}
	for _, fieldKey := range fields {
		val, err := input.Connector.GetContactFieldValue(ctx, input.ContactID, fieldKey)
		if err == nil && val != nil {
			fieldValues[fieldKey] = val
		}
	}

	// Calculate engagement score from available data with time decay
	cutoff := time.Now().AddDate(0, 0, -lookbackDays)
	now := time.Now()

	baseScore := 0.0
	if openCount, ok := parseFloat(fieldValues["open_count"]); ok {
		baseScore += openCount * opensWeight
	}
	if clickCount, ok := parseFloat(fieldValues["click_count"]); ok {
		baseScore += clickCount * clicksWeight
	}

	// Apply time decay if last engagement is available
	for _, dateField := range []string{"last_open_date", "last_click_date", "last_send_date"} {
		if dateStr, ok := fieldValues[dateField].(string); ok && dateStr != "" {
			if t, err := time.Parse(time.RFC3339, dateStr); err == nil {
				if t.After(lastEngagement) {
					lastEngagement = t
				}
			}
		}
	}

	// Apply time decay: newer engagements are weighted higher
	decayFactor := 1.0
	if !lastEngagement.IsZero() && timeDecayDays > 0 {
		daysSinceLastEngagement := int(now.Sub(lastEngagement).Hours() / 24)
		if daysSinceLastEngagement > 0 && daysSinceLastEngagement <= timeDecayDays {
			decayFactor = 1.0 - (float64(daysSinceLastEngagement) / float64(timeDecayDays) * 0.5)
		} else if daysSinceLastEngagement > timeDecayDays {
			decayFactor = 0.5
		}
	}

	engagementScore = int(baseScore * decayFactor)

	// Determine engagement level
	engagementLevel := "disengaged"
	if float64(engagementScore) >= highlyEngagedThreshold && lastEngagement.After(cutoff) {
		engagementLevel = "highly_engaged"
	} else if float64(engagementScore) >= engagedThreshold && lastEngagement.After(cutoff) {
		engagementLevel = "engaged"
	}

	// Determine engagement segment based on score
	segment := "cold"
	if float64(engagementScore) >= hotThreshold {
		segment = "hot"
	} else if float64(engagementScore) >= warmThreshold {
		segment = "warm"
	}

	// Calculate trend if enabled
	trend := "stable"
	if trendEnabled && !lastEngagement.IsZero() {
		trendCutoff := now.AddDate(0, 0, -trendComparisonDays)
		recentEngagements := 0
		olderEngagements := 0

		// Count recent vs older engagements (simplified heuristic based on last engagement date)
		if lastEngagement.After(trendCutoff) {
			recentEngagements = engagementScore
		} else {
			olderEngagements = engagementScore
		}

		if recentEngagements > olderEngagements {
			trend = "increasing"
		} else if recentEngagements < olderEngagements {
			trend = "decreasing"
		}
	}

	output.Logs = append(output.Logs, fmt.Sprintf("Contact %s: engagement_score=%d, level=%s, segment=%s, trend=%s, last_engagement=%s",
		input.ContactID, engagementScore, engagementLevel, segment, trend, lastEngagement.Format(time.RFC3339)))

	// Apply tags based on engagement level
	tags, _ := input.Config["tags"].(map[string]interface{})
	actions := make([]helpers.HelperAction, 0)

	tagFields := map[string]string{
		"highly_engaged": "highly_engaged_tag",
		"engaged":        "engaged_tag",
		"disengaged":     "disengaged_tag",
	}

	for level, tagField := range tagFields {
		if tagID, ok := tags[tagField].(string); ok && tagID != "" {
			if level == engagementLevel {
				if err := input.Connector.ApplyTag(ctx, input.ContactID, tagID); err != nil {
					output.Logs = append(output.Logs, fmt.Sprintf("Failed to apply %s tag: %v", level, err))
				} else {
					actions = append(actions, helpers.HelperAction{
						Type:   "tag_applied",
						Target: tagID,
						Value:  level,
					})
					output.Logs = append(output.Logs, fmt.Sprintf("Applied %s tag (%s)", level, tagID))
				}
			} else {
				// Remove tags for other levels
				if err := input.Connector.RemoveTag(ctx, input.ContactID, tagID); err == nil {
					output.Logs = append(output.Logs, fmt.Sprintf("Removed %s tag (%s)", level, tagID))
				}
			}
		}
	}

	// Apply segment tags
	segmentTagFields := map[string]string{
		"hot":  "hot_segment_tag",
		"warm": "warm_segment_tag",
		"cold": "cold_segment_tag",
	}

	for seg, tagField := range segmentTagFields {
		if tagID, ok := tags[tagField].(string); ok && tagID != "" {
			if seg == segment {
				if err := input.Connector.ApplyTag(ctx, input.ContactID, tagID); err != nil {
					output.Logs = append(output.Logs, fmt.Sprintf("Failed to apply %s segment tag: %v", seg, err))
				} else {
					actions = append(actions, helpers.HelperAction{
						Type:   "tag_applied",
						Target: tagID,
						Value:  seg,
					})
					output.Logs = append(output.Logs, fmt.Sprintf("Applied %s segment tag (%s)", seg, tagID))
				}
			} else {
				if err := input.Connector.RemoveTag(ctx, input.ContactID, tagID); err == nil {
					output.Logs = append(output.Logs, fmt.Sprintf("Removed %s segment tag (%s)", seg, tagID))
				}
			}
		}
	}

	// Store engagement score in custom field
	if scoreField, ok := input.Config["score_field"].(string); ok && scoreField != "" {
		if err := input.Connector.SetContactFieldValue(ctx, input.ContactID, scoreField, engagementScore); err != nil {
			output.Logs = append(output.Logs, fmt.Sprintf("Failed to set score field: %v", err))
		} else {
			actions = append(actions, helpers.HelperAction{
				Type:   "field_updated",
				Target: scoreField,
				Value:  engagementScore,
			})
		}
	}

	// Store last engagement date
	if leField, ok := input.Config["last_engagement_field"].(string); ok && leField != "" && !lastEngagement.IsZero() {
		leStr := lastEngagement.Format(time.RFC3339)
		if err := input.Connector.SetContactFieldValue(ctx, input.ContactID, leField, leStr); err != nil {
			output.Logs = append(output.Logs, fmt.Sprintf("Failed to set last engagement field: %v", err))
		} else {
			actions = append(actions, helpers.HelperAction{
				Type:   "field_updated",
				Target: leField,
				Value:  leStr,
			})
		}
	}

	// Store engagement segment
	if segField, ok := input.Config["segment_field"].(string); ok && segField != "" {
		if err := input.Connector.SetContactFieldValue(ctx, input.ContactID, segField, segment); err != nil {
			output.Logs = append(output.Logs, fmt.Sprintf("Failed to set segment field: %v", err))
		} else {
			actions = append(actions, helpers.HelperAction{
				Type:   "field_updated",
				Target: segField,
				Value:  segment,
			})
		}
	}

	// Store engagement trend
	if trendField, ok := input.Config["trend_field"].(string); ok && trendField != "" && trendEnabled {
		if err := input.Connector.SetContactFieldValue(ctx, input.ContactID, trendField, trend); err != nil {
			output.Logs = append(output.Logs, fmt.Sprintf("Failed to set trend field: %v", err))
		} else {
			actions = append(actions, helpers.HelperAction{
				Type:   "field_updated",
				Target: trendField,
				Value:  trend,
			})
		}
	}

	output.Success = true
	output.Message = fmt.Sprintf("Email engagement evaluated: level=%s, segment=%s, score=%d, trend=%s", engagementLevel, segment, engagementScore, trend)
	output.Actions = actions
	output.ModifiedData = map[string]interface{}{
		"engagement_level":   engagementLevel,
		"engagement_score":   engagementScore,
		"engagement_segment": segment,
		"engagement_trend":   trend,
		"last_engagement":    lastEngagement.Format(time.RFC3339),
	}

	return output, nil
}

func parseFloat(v interface{}) (float64, bool) {
	switch val := v.(type) {
	case float64:
		return val, true
	case int:
		return float64(val), true
	case string:
		if strings.TrimSpace(val) == "" {
			return 0, false
		}
		var f float64
		_, err := fmt.Sscanf(val, "%f", &f)
		return f, err == nil
	default:
		return 0, false
	}
}
