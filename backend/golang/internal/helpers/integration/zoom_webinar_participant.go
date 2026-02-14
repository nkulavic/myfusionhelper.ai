package integration

import (
	"context"
	"fmt"
	"strings"

	"github.com/myfusionhelper/api/internal/helpers"
)

// NewZoomWebinarParticipant creates a new ZoomWebinarParticipant helper instance
func NewZoomWebinarParticipant() helpers.Helper { return &ZoomWebinarParticipant{} }

func init() {
	helpers.Register("zoom_webinar_participant", func() helpers.Helper { return &ZoomWebinarParticipant{} })
}

// ZoomWebinarParticipant processes webinar attendees after a webinar ends.
// Applies segmentation tags based on attendance metrics (attended_percent, duration_minutes)
// and engagement level (high/medium/low).
type ZoomWebinarParticipant struct{}

func (h *ZoomWebinarParticipant) GetName() string     { return "Zoom Webinar Participant" }
func (h *ZoomWebinarParticipant) GetType() string     { return "zoom_webinar_participant" }
func (h *ZoomWebinarParticipant) GetCategory() string { return "integration" }
func (h *ZoomWebinarParticipant) GetDescription() string {
	return "Process webinar attendees and apply segmentation tags based on engagement"
}
func (h *ZoomWebinarParticipant) RequiresCRM() bool       { return true }
func (h *ZoomWebinarParticipant) SupportedCRMs() []string { return nil }

func (h *ZoomWebinarParticipant) GetConfigSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"webinar_id": map[string]interface{}{
				"type":        "string",
				"description": "Zoom webinar ID to fetch participants for",
			},
			"attended_percent": map[string]interface{}{
				"type":        "number",
				"description": "Percent of webinar attended (0-100)",
			},
			"duration_minutes": map[string]interface{}{
				"type":        "number",
				"description": "Duration in minutes the participant attended",
			},
			"tag_prefix": map[string]interface{}{
				"type":        "string",
				"description": "Prefix for applied tags (default: 'Webinar')",
				"default":     "Webinar",
			},
			"high_engagement_threshold": map[string]interface{}{
				"type":        "number",
				"description": "Percent threshold for high engagement (default: 75)",
				"default":     75,
			},
			"medium_engagement_threshold": map[string]interface{}{
				"type":        "number",
				"description": "Percent threshold for medium engagement (default: 50)",
				"default":     50,
			},
			"apply_attendance_tag": map[string]interface{}{
				"type":        "boolean",
				"description": "Apply 'Attended' tag (default: true)",
				"default":     true,
			},
			"apply_engagement_tags": map[string]interface{}{
				"type":        "boolean",
				"description": "Apply engagement level tags (high/medium/low) (default: true)",
				"default":     true,
			},
			"apply_duration_tag": map[string]interface{}{
				"type":        "boolean",
				"description": "Apply duration-based tag (default: false)",
				"default":     false,
			},
			"custom_tags": map[string]interface{}{
				"type": "array",
				"items": map[string]interface{}{
					"type": "string",
				},
				"description": "Additional custom tags to apply to all participants",
			},
		},
		"required": []string{},
	}
}

func (h *ZoomWebinarParticipant) ValidateConfig(config map[string]interface{}) error {
	// Either webinar_id or attended_percent should be present for meaningful processing
	// but we'll allow flexible usage
	return nil
}

func (h *ZoomWebinarParticipant) Execute(ctx context.Context, input helpers.HelperInput) (*helpers.HelperOutput, error) {
	output := &helpers.HelperOutput{
		Actions: make([]helpers.HelperAction, 0),
		Logs:    make([]string, 0),
	}

	// Get configuration
	tagPrefix := getStringConfigValue(input.Config, "tag_prefix", "Webinar")
	highThreshold := getFloatConfigValue(input.Config, "high_engagement_threshold", 75.0)
	mediumThreshold := getFloatConfigValue(input.Config, "medium_engagement_threshold", 50.0)
	applyAttendance := getBoolConfigValue(input.Config, "apply_attendance_tag", true)
	applyEngagement := getBoolConfigValue(input.Config, "apply_engagement_tags", true)
	applyDuration := getBoolConfigValue(input.Config, "apply_duration_tag", false)

	attendedPercent := getFloatConfigValue(input.Config, "attended_percent", 0.0)
	durationMinutes := getFloatConfigValue(input.Config, "duration_minutes", 0.0)
	webinarID := getStringConfigValue(input.Config, "webinar_id", "")

	tagsToApply := make([]string, 0)

	// Apply attendance tag
	if applyAttendance {
		tagsToApply = append(tagsToApply, fmt.Sprintf("%s Attended", tagPrefix))
	}

	// Apply engagement level tags
	if applyEngagement && attendedPercent > 0 {
		var engagementLevel string
		if attendedPercent >= highThreshold {
			engagementLevel = "High Engagement"
		} else if attendedPercent >= mediumThreshold {
			engagementLevel = "Medium Engagement"
		} else {
			engagementLevel = "Low Engagement"
		}
		tagsToApply = append(tagsToApply, fmt.Sprintf("%s %s", tagPrefix, engagementLevel))
		output.Logs = append(output.Logs, fmt.Sprintf("Engagement level: %s (%.1f%% attended)", engagementLevel, attendedPercent))
	}

	// Apply duration-based tag
	if applyDuration && durationMinutes > 0 {
		var durationBucket string
		if durationMinutes >= 60 {
			durationBucket = "60+ min"
		} else if durationMinutes >= 30 {
			durationBucket = "30-60 min"
		} else if durationMinutes >= 15 {
			durationBucket = "15-30 min"
		} else {
			durationBucket = "<15 min"
		}
		tagsToApply = append(tagsToApply, fmt.Sprintf("%s Duration: %s", tagPrefix, durationBucket))
		output.Logs = append(output.Logs, fmt.Sprintf("Duration: %.0f minutes (%s)", durationMinutes, durationBucket))
	}

	// Apply webinar-specific tag if webinar ID is provided
	if webinarID != "" {
		tagsToApply = append(tagsToApply, fmt.Sprintf("%s ID:%s", tagPrefix, webinarID))
	}

	// Add custom tags
	if customTags, ok := input.Config["custom_tags"].([]interface{}); ok {
		for _, ct := range customTags {
			if tag, ok := ct.(string); ok && tag != "" {
				tagsToApply = append(tagsToApply, tag)
			}
		}
	}

	// Get available tags from CRM
	availableTags, err := input.Connector.GetTags(ctx)
	if err != nil {
		output.Message = fmt.Sprintf("Failed to get tags: %v", err)
		return output, err
	}

	tagMap := make(map[string]string) // name -> ID
	for _, tag := range availableTags {
		tagMap[strings.ToLower(tag.Name)] = tag.ID
	}

	// Apply each tag
	appliedCount := 0
	for _, tagName := range tagsToApply {
		tagID, exists := tagMap[strings.ToLower(tagName)]
		if !exists {
			output.Logs = append(output.Logs, fmt.Sprintf("Warning: tag '%s' not found in CRM", tagName))
			continue
		}

		err := input.Connector.ApplyTag(ctx, input.ContactID, tagID)
		if err != nil {
			output.Logs = append(output.Logs, fmt.Sprintf("Failed to apply tag '%s': %v", tagName, err))
			continue
		}

		output.Actions = append(output.Actions, helpers.HelperAction{
			Type:   "tag_applied",
			Target: input.ContactID,
			Value:  tagName,
		})
		output.Logs = append(output.Logs, fmt.Sprintf("Applied tag: %s", tagName))
		appliedCount++
	}

	output.Success = appliedCount > 0
	if output.Success {
		output.Message = fmt.Sprintf("Applied %d webinar participant tag(s)", appliedCount)
	} else {
		output.Success = true // No errors, just no tags to apply
		output.Message = "No webinar participant tags applied"
	}

	output.ModifiedData = map[string]interface{}{
		"attended_percent": attendedPercent,
		"duration_minutes": durationMinutes,
		"tags_applied":     appliedCount,
		"engagement_level": determineEngagementLevel(attendedPercent, highThreshold, mediumThreshold),
	}

	return output, nil
}

func determineEngagementLevel(percent, high, medium float64) string {
	if percent >= high {
		return "high"
	} else if percent >= medium {
		return "medium"
	} else if percent > 0 {
		return "low"
	}
	return "none"
}
