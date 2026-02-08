package integration

import (
	"context"
	"fmt"
	"strings"

	"github.com/myfusionhelper/api/internal/helpers"
)

func init() {
	helpers.Register("zoom_webinar_absentee", func() helpers.Helper { return &ZoomWebinarAbsentee{} })
}

// ZoomWebinarAbsentee processes contacts who registered for a webinar but did not attend.
// Applies no-show tags for re-engagement campaigns.
type ZoomWebinarAbsentee struct{}

func (h *ZoomWebinarAbsentee) GetName() string     { return "Zoom Webinar Absentee" }
func (h *ZoomWebinarAbsentee) GetType() string     { return "zoom_webinar_absentee" }
func (h *ZoomWebinarAbsentee) GetCategory() string { return "integration" }
func (h *ZoomWebinarAbsentee) GetDescription() string {
	return "Process webinar registrants who did not attend (no-shows)"
}
func (h *ZoomWebinarAbsentee) RequiresCRM() bool       { return true }
func (h *ZoomWebinarAbsentee) SupportedCRMs() []string { return nil }

func (h *ZoomWebinarAbsentee) GetConfigSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"webinar_id": map[string]interface{}{
				"type":        "string",
				"description": "Zoom webinar ID",
			},
			"tag_prefix": map[string]interface{}{
				"type":        "string",
				"description": "Prefix for applied tags (default: 'Webinar')",
				"default":     "Webinar",
			},
			"apply_no_show_tag": map[string]interface{}{
				"type":        "boolean",
				"description": "Apply 'No Show' tag (default: true)",
				"default":     true,
			},
			"apply_registered_tag": map[string]interface{}{
				"type":        "boolean",
				"description": "Apply 'Registered' tag (default: true)",
				"default":     true,
			},
			"custom_tags": map[string]interface{}{
				"type": "array",
				"items": map[string]interface{}{
					"type": "string",
				},
				"description": "Additional custom tags to apply to absentees",
			},
		},
		"required": []string{},
	}
}

func (h *ZoomWebinarAbsentee) ValidateConfig(config map[string]interface{}) error {
	return nil
}

func (h *ZoomWebinarAbsentee) Execute(ctx context.Context, input helpers.HelperInput) (*helpers.HelperOutput, error) {
	output := &helpers.HelperOutput{
		Actions: make([]helpers.HelperAction, 0),
		Logs:    make([]string, 0),
	}

	// Get configuration
	tagPrefix := getStringConfigValue(input.Config, "tag_prefix", "Webinar")
	applyNoShow := getBoolConfigValue(input.Config, "apply_no_show_tag", true)
	applyRegistered := getBoolConfigValue(input.Config, "apply_registered_tag", true)
	webinarID := getStringConfigValue(input.Config, "webinar_id", "")

	tagsToApply := make([]string, 0)

	// Apply no-show tag
	if applyNoShow {
		tagsToApply = append(tagsToApply, fmt.Sprintf("%s No Show", tagPrefix))
	}

	// Apply registered tag
	if applyRegistered {
		tagsToApply = append(tagsToApply, fmt.Sprintf("%s Registered", tagPrefix))
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
		output.Message = fmt.Sprintf("Applied %d webinar absentee tag(s)", appliedCount)
	} else {
		output.Success = true // No errors, just no tags to apply
		output.Message = "No webinar absentee tags applied"
	}

	output.ModifiedData = map[string]interface{}{
		"tags_applied": appliedCount,
		"status":       "no_show",
	}

	return output, nil
}
