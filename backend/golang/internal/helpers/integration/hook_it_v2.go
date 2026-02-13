package integration

import (
	"context"
	"fmt"

	"github.com/myfusionhelper/api/internal/helpers"
)

// NewHookItV2 creates a new HookItV2 helper instance
func NewHookItV2() helpers.Helper { return &HookItV2{} }

func init() {
	helpers.Register("hook_it_v2", func() helpers.Helper { return &HookItV2{} })
}

// HookItV2 handles webhook events with tag-based routing.
// Applies different tags to contacts based on the webhook event type.
// Example: "order.created" -> apply tag "New Order", "order.canceled" -> apply tag "Canceled Order"
type HookItV2 struct{}

func (h *HookItV2) GetName() string     { return "Hook It V2 (Tag Router)" }
func (h *HookItV2) GetType() string     { return "hook_it_v2" }
func (h *HookItV2) GetCategory() string { return "integration" }
func (h *HookItV2) GetDescription() string {
	return "Route webhook events to apply different tags based on event type"
}
func (h *HookItV2) RequiresCRM() bool       { return true }
func (h *HookItV2) SupportedCRMs() []string { return nil }

func (h *HookItV2) GetConfigSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"event_tag_map": map[string]interface{}{
				"type": "object",
				"description": "Map of event names to tag IDs (e.g., {\"order.created\": \"tag123\", \"order.canceled\": \"tag456\"})",
				"additionalProperties": map[string]interface{}{
					"type": "string",
				},
			},
			"event_field": map[string]interface{}{
				"type":        "string",
				"description": "Field name in webhook payload containing the event type (defaults to 'event')",
			},
			"default_tag_id": map[string]interface{}{
				"type":        "string",
				"description": "Tag ID to apply if event type doesn't match any mapping",
			},
		},
		"required": []string{"event_tag_map"},
	}
}

func (h *HookItV2) ValidateConfig(config map[string]interface{}) error {
	if config["event_tag_map"] == nil {
		return fmt.Errorf("event_tag_map is required")
	}

	eventTagMap, ok := config["event_tag_map"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("event_tag_map must be an object")
	}

	if len(eventTagMap) == 0 {
		return fmt.Errorf("event_tag_map must contain at least one event mapping")
	}

	// Validate that all values are strings
	for eventName, tagID := range eventTagMap {
		if _, ok := tagID.(string); !ok {
			return fmt.Errorf("event_tag_map[%s] must be a string tag ID", eventName)
		}
	}

	return nil
}

func (h *HookItV2) Execute(ctx context.Context, input helpers.HelperInput) (*helpers.HelperOutput, error) {
	output := &helpers.HelperOutput{
		Actions: make([]helpers.HelperAction, 0),
		Logs:    make([]string, 0),
	}

	// Extract event_tag_map from config
	eventTagMap, ok := input.Config["event_tag_map"].(map[string]interface{})
	if !ok {
		output.Message = "Invalid event_tag_map configuration"
		output.Logs = append(output.Logs, output.Message)
		return output, fmt.Errorf("event_tag_map must be an object")
	}

	// Get event field name (defaults to "event")
	eventField := "event"
	if ef, ok := input.Config["event_field"].(string); ok && ef != "" {
		eventField = ef
	}

	// Extract event type from webhook data (stored in config)
	var eventType string
	if webhookData, ok := input.Config["webhook_data"].(map[string]interface{}); ok {
		if et, ok := webhookData[eventField].(string); ok {
			eventType = et
		}
	}

	// If no event type found in webhook_data, try to use hook_action from config
	if eventType == "" {
		if ha, ok := input.Config["hook_action"].(string); ok {
			eventType = ha
		}
	}

	// Fallback: try to get event directly from config
	if eventType == "" {
		if et, ok := input.Config[eventField].(string); ok {
			eventType = et
		}
	}

	// Find matching tag ID
	var tagID string
	if eventType != "" {
		if tid, ok := eventTagMap[eventType].(string); ok {
			tagID = tid
		}
	}

	// Use default tag if no match found
	if tagID == "" {
		if defaultTag, ok := input.Config["default_tag_id"].(string); ok && defaultTag != "" {
			tagID = defaultTag
			output.Logs = append(output.Logs, fmt.Sprintf("No mapping found for event '%s', using default tag %s", eventType, tagID))
		} else {
			output.Success = true
			output.Message = fmt.Sprintf("No tag mapping found for event '%s' and no default tag configured", eventType)
			output.Logs = append(output.Logs, output.Message)
			output.ModifiedData = map[string]interface{}{
				"event_type": eventType,
				"tag_applied": false,
			}
			return output, nil
		}
	}

	// Apply tag to contact
	err := input.Connector.ApplyTag(ctx, input.ContactID, tagID)
	if err != nil {
		output.Message = fmt.Sprintf("Failed to apply tag %s for event '%s': %v", tagID, eventType, err)
		output.Logs = append(output.Logs, output.Message)
		return output, err
	}

	output.Actions = append(output.Actions, helpers.HelperAction{
		Type:   "tag_applied",
		Target: input.ContactID,
		Value:  tagID,
	})
	output.Logs = append(output.Logs, fmt.Sprintf("Applied tag %s for event '%s'", tagID, eventType))

	output.Success = true
	output.Message = fmt.Sprintf("Tag routing complete: applied tag %s for event '%s'", tagID, eventType)
	output.ModifiedData = map[string]interface{}{
		"event_type": eventType,
		"tag_id":     tagID,
		"tag_applied": true,
	}

	return output, nil
}
