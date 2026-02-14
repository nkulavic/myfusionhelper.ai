package tagging

import (
	"context"
	"fmt"

	"github.com/myfusionhelper/api/internal/helpers"
)

func init() {
	helpers.Register("tag_it", func() helpers.Helper { return &TagIt{} })
}

// NewTagIt creates a new TagIt helper instance
func NewTagIt() helpers.Helper { return &TagIt{} }

// TagIt applies or removes tags from a contact based on configuration
type TagIt struct{}

func (h *TagIt) GetName() string        { return "Tag It" }
func (h *TagIt) GetType() string        { return "tag_it" }
func (h *TagIt) GetCategory() string    { return "tagging" }
func (h *TagIt) GetDescription() string { return "Apply or remove tags from a contact" }
func (h *TagIt) RequiresCRM() bool      { return true }
func (h *TagIt) SupportedCRMs() []string { return nil } // All CRMs

func (h *TagIt) GetConfigSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"action": map[string]interface{}{
				"type":        "string",
				"enum":        []string{"apply", "remove"},
				"description": "Whether to apply or remove the tags",
			},
			"tag_ids": map[string]interface{}{
				"type":        "array",
				"items":       map[string]interface{}{"type": "string"},
				"description": "List of tag IDs to apply or remove",
			},
		},
		"required": []string{"action", "tag_ids"},
	}
}

func (h *TagIt) ValidateConfig(config map[string]interface{}) error {
	action, ok := config["action"].(string)
	if !ok || (action != "apply" && action != "remove") {
		return fmt.Errorf("action must be 'apply' or 'remove'")
	}

	tagIDs, ok := config["tag_ids"]
	if !ok {
		return fmt.Errorf("tag_ids is required")
	}

	switch v := tagIDs.(type) {
	case []interface{}:
		if len(v) == 0 {
			return fmt.Errorf("tag_ids must contain at least one tag")
		}
	case []string:
		if len(v) == 0 {
			return fmt.Errorf("tag_ids must contain at least one tag")
		}
	default:
		return fmt.Errorf("tag_ids must be an array of strings")
	}

	return nil
}

func (h *TagIt) Execute(ctx context.Context, input helpers.HelperInput) (*helpers.HelperOutput, error) {
	action := input.Config["action"].(string)
	tagIDs := extractStringSlice(input.Config["tag_ids"])

	output := &helpers.HelperOutput{
		Actions: make([]helpers.HelperAction, 0),
		Logs:    make([]string, 0),
	}

	for _, tagID := range tagIDs {
		var err error
		if action == "apply" {
			err = input.Connector.ApplyTag(ctx, input.ContactID, tagID)
		} else {
			err = input.Connector.RemoveTag(ctx, input.ContactID, tagID)
		}

		if err != nil {
			output.Logs = append(output.Logs, fmt.Sprintf("Failed to %s tag %s: %v", action, tagID, err))
			continue
		}

		actionType := "tag_applied"
		if action == "remove" {
			actionType = "tag_removed"
		}

		output.Actions = append(output.Actions, helpers.HelperAction{
			Type:   actionType,
			Target: input.ContactID,
			Value:  tagID,
		})
		output.Logs = append(output.Logs, fmt.Sprintf("Tag %s %sd on contact %s", tagID, action, input.ContactID))
	}

	output.Success = len(output.Actions) > 0
	if output.Success {
		output.Message = fmt.Sprintf("Successfully %sd %d tag(s)", action, len(output.Actions))
	} else {
		output.Message = fmt.Sprintf("Failed to %s any tags", action)
	}

	return output, nil
}

func extractStringSlice(v interface{}) []string {
	switch val := v.(type) {
	case []string:
		return val
	case []interface{}:
		result := make([]string, 0, len(val))
		for _, item := range val {
			if s, ok := item.(string); ok {
				result = append(result, s)
			} else {
				result = append(result, fmt.Sprintf("%v", item))
			}
		}
		return result
	default:
		return nil
	}
}
