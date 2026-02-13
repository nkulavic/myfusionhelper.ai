package integration

import (
	"context"
	"fmt"

	"github.com/myfusionhelper/api/internal/helpers"
)

// NewHookItByTag creates a new HookItByTag helper instance
func NewHookItByTag() helpers.Helper { return &HookItByTag{} }

func init() {
	helpers.Register("hook_it_by_tag", func() helpers.Helper { return &HookItByTag{} })
}

// HookItByTag handles webhook events with conditional tag-based processing.
// Only processes webhooks if the contact has specific required tags
// and doesn't have any forbidden tags.
type HookItByTag struct{}

func (h *HookItByTag) GetName() string     { return "Hook It By Tag (Conditional)" }
func (h *HookItByTag) GetType() string     { return "hook_it_by_tag" }
func (h *HookItByTag) GetCategory() string { return "integration" }
func (h *HookItByTag) GetDescription() string {
	return "Process webhook only if contact has required tags and lacks forbidden tags"
}
func (h *HookItByTag) RequiresCRM() bool       { return true }
func (h *HookItByTag) SupportedCRMs() []string { return nil }

func (h *HookItByTag) GetConfigSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"required_tags": map[string]interface{}{
				"type": "array",
				"items": map[string]interface{}{
					"type": "string",
				},
				"description": "Tag IDs that contact must have for webhook to process (all required)",
			},
			"forbidden_tags": map[string]interface{}{
				"type": "array",
				"items": map[string]interface{}{
					"type": "string",
				},
				"description": "Tag IDs that contact must NOT have for webhook to process",
			},
			"match_mode": map[string]interface{}{
				"type":        "string",
				"enum":        []string{"all", "any"},
				"description": "For required_tags: 'all' = must have all tags, 'any' = must have at least one tag (defaults to 'all')",
			},
			"goal_name": map[string]interface{}{
				"type":        "string",
				"description": "Goal to achieve if conditions are met",
			},
			"integration": map[string]interface{}{
				"type":        "string",
				"description": "Integration name for goal calls (defaults to 'myfusionhelper')",
			},
			"apply_tag_on_success": map[string]interface{}{
				"type":        "string",
				"description": "Tag ID to apply if webhook is processed successfully",
			},
			"apply_tag_on_skip": map[string]interface{}{
				"type":        "string",
				"description": "Tag ID to apply if webhook is skipped due to tag conditions",
			},
		},
		"required": []string{},
	}
}

func (h *HookItByTag) ValidateConfig(config map[string]interface{}) error {
	// At least one of required_tags or forbidden_tags should be specified
	hasRequiredTags := config["required_tags"] != nil
	hasForbiddenTags := config["forbidden_tags"] != nil

	if !hasRequiredTags && !hasForbiddenTags {
		return fmt.Errorf("at least one of required_tags or forbidden_tags must be specified")
	}

	// Validate match_mode if present
	if matchMode, ok := config["match_mode"].(string); ok {
		if matchMode != "all" && matchMode != "any" {
			return fmt.Errorf("match_mode must be 'all' or 'any'")
		}
	}

	return nil
}

// hasTag checks if a tag ID is in the contact's tag list
func (h *HookItByTag) hasTag(contactTags []string, tagID string) bool {
	for _, t := range contactTags {
		if t == tagID {
			return true
		}
	}
	return false
}

func (h *HookItByTag) Execute(ctx context.Context, input helpers.HelperInput) (*helpers.HelperOutput, error) {
	output := &helpers.HelperOutput{
		Actions: make([]helpers.HelperAction, 0),
		Logs:    make([]string, 0),
	}

	// Get contact to check tags
	contact, err := input.Connector.GetContact(ctx, input.ContactID)
	if err != nil {
		output.Message = fmt.Sprintf("Failed to fetch contact: %v", err)
		output.Logs = append(output.Logs, output.Message)
		return output, err
	}

	// Extract tag IDs from contact
	contactTagIDs := make([]string, 0, len(contact.Tags))
	for _, tag := range contact.Tags {
		contactTagIDs = append(contactTagIDs, tag.ID)
	}

	// Get match mode (defaults to "all")
	matchMode := "all"
	if mode, ok := input.Config["match_mode"].(string); ok && mode != "" {
		matchMode = mode
	}

	// Check required_tags
	if requiredTagsInterface, ok := input.Config["required_tags"].([]interface{}); ok && len(requiredTagsInterface) > 0 {
		requiredTags := make([]string, 0, len(requiredTagsInterface))
		for _, tagInterface := range requiredTagsInterface {
			if tagID, ok := tagInterface.(string); ok {
				requiredTags = append(requiredTags, tagID)
			}
		}

		if matchMode == "all" {
			// Contact must have ALL required tags
			for _, requiredTag := range requiredTags {
				if !h.hasTag(contactTagIDs, requiredTag) {
					// Apply skip tag if configured
					if skipTag, ok := input.Config["apply_tag_on_skip"].(string); ok && skipTag != "" {
						_ = input.Connector.ApplyTag(ctx, input.ContactID, skipTag)
						output.Actions = append(output.Actions, helpers.HelperAction{
							Type:   "tag_applied",
							Target: input.ContactID,
							Value:  skipTag,
						})
					}

					output.Success = true
					output.Message = fmt.Sprintf("Webhook skipped: contact missing required tag %s", requiredTag)
					output.Logs = append(output.Logs, output.Message)
					output.ModifiedData = map[string]interface{}{
						"processed":     false,
						"skip_reason":   "missing_required_tag",
						"missing_tag":   requiredTag,
						"contact_tags":  contactTagIDs,
						"required_tags": requiredTags,
					}
					return output, nil
				}
			}
		} else {
			// Contact must have at least ONE required tag
			hasAtLeastOne := false
			for _, requiredTag := range requiredTags {
				if h.hasTag(contactTagIDs, requiredTag) {
					hasAtLeastOne = true
					break
				}
			}

			if !hasAtLeastOne {
				// Apply skip tag if configured
				if skipTag, ok := input.Config["apply_tag_on_skip"].(string); ok && skipTag != "" {
					_ = input.Connector.ApplyTag(ctx, input.ContactID, skipTag)
					output.Actions = append(output.Actions, helpers.HelperAction{
						Type:   "tag_applied",
						Target: input.ContactID,
						Value:  skipTag,
					})
				}

				output.Success = true
				output.Message = "Webhook skipped: contact has none of the required tags"
				output.Logs = append(output.Logs, output.Message)
				output.ModifiedData = map[string]interface{}{
					"processed":     false,
					"skip_reason":   "missing_any_required_tag",
					"contact_tags":  contactTagIDs,
					"required_tags": requiredTags,
				}
				return output, nil
			}
		}
	}

	// Check forbidden_tags
	if forbiddenTagsInterface, ok := input.Config["forbidden_tags"].([]interface{}); ok && len(forbiddenTagsInterface) > 0 {
		for _, tagInterface := range forbiddenTagsInterface {
			if tagID, ok := tagInterface.(string); ok {
				if h.hasTag(contactTagIDs, tagID) {
					// Apply skip tag if configured
					if skipTag, ok := input.Config["apply_tag_on_skip"].(string); ok && skipTag != "" {
						_ = input.Connector.ApplyTag(ctx, input.ContactID, skipTag)
						output.Actions = append(output.Actions, helpers.HelperAction{
							Type:   "tag_applied",
							Target: input.ContactID,
							Value:  skipTag,
						})
					}

					output.Success = true
					output.Message = fmt.Sprintf("Webhook skipped: contact has forbidden tag %s", tagID)
					output.Logs = append(output.Logs, output.Message)
					output.ModifiedData = map[string]interface{}{
						"processed":      false,
						"skip_reason":    "has_forbidden_tag",
						"forbidden_tag":  tagID,
						"contact_tags":   contactTagIDs,
					}
					return output, nil
				}
			}
		}
	}

	// Conditions met - process webhook
	output.Logs = append(output.Logs, "Tag conditions met, processing webhook")

	// Fire goal if configured
	if goalName, ok := input.Config["goal_name"].(string); ok && goalName != "" {
		integration := "myfusionhelper"
		if intg, ok := input.Config["integration"].(string); ok && intg != "" {
			integration = intg
		}

		err := input.Connector.AchieveGoal(ctx, input.ContactID, goalName, integration)
		if err != nil {
			output.Message = fmt.Sprintf("Failed to achieve goal '%s': %v", goalName, err)
			output.Logs = append(output.Logs, output.Message)
			return output, err
		}

		output.Actions = append(output.Actions, helpers.HelperAction{
			Type:   "goal_achieved",
			Target: input.ContactID,
			Value:  goalName,
		})
		output.Logs = append(output.Logs, fmt.Sprintf("Achieved goal '%s'", goalName))
	}

	// Apply success tag if configured
	if successTag, ok := input.Config["apply_tag_on_success"].(string); ok && successTag != "" {
		err := input.Connector.ApplyTag(ctx, input.ContactID, successTag)
		if err != nil {
			output.Logs = append(output.Logs, fmt.Sprintf("Failed to apply success tag: %v", err))
		} else {
			output.Actions = append(output.Actions, helpers.HelperAction{
				Type:   "tag_applied",
				Target: input.ContactID,
				Value:  successTag,
			})
			output.Logs = append(output.Logs, "Applied success tag")
		}
	}

	output.Success = true
	output.Message = "Webhook processed successfully - tag conditions met"
	output.ModifiedData = map[string]interface{}{
		"processed":    true,
		"contact_tags": contactTagIDs,
	}

	return output, nil
}
