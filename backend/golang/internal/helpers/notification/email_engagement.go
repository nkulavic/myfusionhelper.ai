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

	// Calculate engagement score from available data
	cutoff := time.Now().AddDate(0, 0, -lookbackDays)

	if openCount, ok := parseFloat(fieldValues["open_count"]); ok {
		engagementScore += int(openCount)
	}
	if clickCount, ok := parseFloat(fieldValues["click_count"]); ok {
		// Clicks are weighted more than opens
		engagementScore += int(clickCount) * 2
	}

	// Track last engagement date
	for _, dateField := range []string{"last_open_date", "last_click_date", "last_send_date"} {
		if dateStr, ok := fieldValues[dateField].(string); ok && dateStr != "" {
			if t, err := time.Parse(time.RFC3339, dateStr); err == nil {
				if t.After(lastEngagement) {
					lastEngagement = t
				}
			}
		}
	}

	// Determine engagement level
	engagementLevel := "disengaged"
	if float64(engagementScore) >= highlyEngagedThreshold && lastEngagement.After(cutoff) {
		engagementLevel = "highly_engaged"
	} else if float64(engagementScore) >= engagedThreshold && lastEngagement.After(cutoff) {
		engagementLevel = "engaged"
	}

	output.Logs = append(output.Logs, fmt.Sprintf("Contact %s: engagement_score=%d, level=%s, last_engagement=%s",
		input.ContactID, engagementScore, engagementLevel, lastEngagement.Format(time.RFC3339)))

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

	output.Success = true
	output.Message = fmt.Sprintf("Email engagement evaluated: level=%s, score=%d", engagementLevel, engagementScore)
	output.Actions = actions
	output.ModifiedData = map[string]interface{}{
		"engagement_level": engagementLevel,
		"engagement_score": engagementScore,
		"last_engagement":  lastEngagement.Format(time.RFC3339),
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
