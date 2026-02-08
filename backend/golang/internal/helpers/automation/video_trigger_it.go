package automation

import (
	"context"
	"fmt"
	"time"

	"github.com/myfusionhelper/api/internal/helpers"
)

func init() {
	helpers.Register("video_trigger_it", func() helpers.Helper { return &VideoTriggerIt{} })
}

// VideoTriggerIt triggers actions based on video engagement events (view, completion)
type VideoTriggerIt struct{}

func (h *VideoTriggerIt) GetName() string     { return "Video Trigger It" }
func (h *VideoTriggerIt) GetType() string     { return "video_trigger_it" }
func (h *VideoTriggerIt) GetCategory() string { return "automation" }
func (h *VideoTriggerIt) GetDescription() string {
	return "Trigger CRM actions based on video engagement events (view start, percentage watched, completion)"
}
func (h *VideoTriggerIt) RequiresCRM() bool       { return true }
func (h *VideoTriggerIt) SupportedCRMs() []string { return nil }

func (h *VideoTriggerIt) GetConfigSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"video_id": map[string]interface{}{
				"type":        "string",
				"description": "Video identifier (URL, platform ID, or custom ID)",
			},
			"event_type": map[string]interface{}{
				"type":        "string",
				"enum":        []string{"view_start", "25_percent", "50_percent", "75_percent", "100_percent", "custom_percent"},
				"description": "Video engagement event to trigger on",
			},
			"custom_percent": map[string]interface{}{
				"type":        "number",
				"description": "Custom percentage threshold (required if event_type is custom_percent)",
			},
			"watch_duration": map[string]interface{}{
				"type":        "number",
				"description": "Optional: actual watch duration in seconds (from video platform)",
			},
			"video_title": map[string]interface{}{
				"type":        "string",
				"description": "Optional: video title for logging",
			},
			"apply_tag": map[string]interface{}{
				"type":        "string",
				"description": "Tag ID to apply when event occurs",
			},
			"achieve_goal": map[string]interface{}{
				"type":        "string",
				"description": "Goal to achieve when event occurs",
			},
			"save_to_field": map[string]interface{}{
				"type":        "string",
				"description": "Optional: field to save watch timestamp",
			},
		},
		"required": []string{"video_id", "event_type"},
	}
}

func (h *VideoTriggerIt) ValidateConfig(config map[string]interface{}) error {
	if _, ok := config["video_id"].(string); !ok || config["video_id"] == "" {
		return fmt.Errorf("video_id is required")
	}

	eventType, ok := config["event_type"].(string)
	if !ok || eventType == "" {
		return fmt.Errorf("event_type is required")
	}

	if eventType == "custom_percent" {
		customPercent, ok := config["custom_percent"].(float64)
		if !ok || customPercent <= 0 || customPercent > 100 {
			return fmt.Errorf("custom_percent must be between 0 and 100 when event_type is custom_percent")
		}
	}

	return nil
}

func (h *VideoTriggerIt) Execute(ctx context.Context, input helpers.HelperInput) (*helpers.HelperOutput, error) {
	videoID := input.Config["video_id"].(string)
	eventType := input.Config["event_type"].(string)
	videoTitle, _ := input.Config["video_title"].(string)
	if videoTitle == "" {
		videoTitle = videoID
	}
	watchDuration, _ := input.Config["watch_duration"].(float64)
	applyTag, _ := input.Config["apply_tag"].(string)
	achieveGoal, _ := input.Config["achieve_goal"].(string)
	saveToField, _ := input.Config["save_to_field"].(string)

	output := &helpers.HelperOutput{
		Actions: make([]helpers.HelperAction, 0),
		Logs:    make([]string, 0),
	}

	// Determine the engagement level
	var engagementLevel string
	var percentageWatched float64
	switch eventType {
	case "view_start":
		engagementLevel = "Started watching"
		percentageWatched = 0
	case "25_percent":
		engagementLevel = "Watched 25%"
		percentageWatched = 25
	case "50_percent":
		engagementLevel = "Watched 50%"
		percentageWatched = 50
	case "75_percent":
		engagementLevel = "Watched 75%"
		percentageWatched = 75
	case "100_percent":
		engagementLevel = "Completed video"
		percentageWatched = 100
	case "custom_percent":
		customPercent := input.Config["custom_percent"].(float64)
		engagementLevel = fmt.Sprintf("Watched %.0f%%", customPercent)
		percentageWatched = customPercent
	}

	output.Logs = append(output.Logs, fmt.Sprintf("Video engagement event: %s - %s (%s)", eventType, engagementLevel, videoTitle))
	if watchDuration > 0 {
		output.Logs = append(output.Logs, fmt.Sprintf("Watch duration: %.0f seconds", watchDuration))
	}

	// Save timestamp to field if configured
	timestamp := time.Now().Format(time.RFC3339)
	if saveToField != "" {
		err := input.Connector.SetContactFieldValue(ctx, input.ContactID, saveToField, timestamp)
		if err != nil {
			output.Logs = append(output.Logs, fmt.Sprintf("Warning: Failed to save timestamp to field '%s': %v", saveToField, err))
		} else {
			output.Actions = append(output.Actions, helpers.HelperAction{
				Type:   "field_updated",
				Target: saveToField,
				Value:  timestamp,
			})
			output.Logs = append(output.Logs, fmt.Sprintf("Saved watch timestamp to field '%s'", saveToField))
		}
	}

	// Apply tag if configured
	if applyTag != "" {
		err := input.Connector.ApplyTag(ctx, input.ContactID, applyTag)
		if err != nil {
			output.Logs = append(output.Logs, fmt.Sprintf("Failed to apply tag '%s': %v", applyTag, err))
		} else {
			output.Actions = append(output.Actions, helpers.HelperAction{
				Type:   "tag_applied",
				Target: input.ContactID,
				Value:  applyTag,
			})
			output.Logs = append(output.Logs, fmt.Sprintf("Applied tag '%s'", applyTag))
		}
	}

	// Achieve goal if configured
	if achieveGoal != "" {
		err := input.Connector.AchieveGoal(ctx, input.ContactID, achieveGoal, "video_trigger_it")
		if err != nil {
			output.Logs = append(output.Logs, fmt.Sprintf("Failed to achieve goal '%s': %v", achieveGoal, err))
		} else {
			output.Actions = append(output.Actions, helpers.HelperAction{
				Type:   "goal_achieved",
				Target: input.ContactID,
				Value:  achieveGoal,
			})
			output.Logs = append(output.Logs, fmt.Sprintf("Achieved goal '%s'", achieveGoal))
		}
	}

	output.Success = true
	output.Message = fmt.Sprintf("Video engagement tracked: %s - %s", videoTitle, engagementLevel)
	output.ModifiedData = map[string]interface{}{
		"video_id":           videoID,
		"video_title":        videoTitle,
		"event_type":         eventType,
		"engagement_level":   engagementLevel,
		"percentage_watched": percentageWatched,
		"watch_duration":     watchDuration,
		"timestamp":          timestamp,
	}

	return output, nil
}
