package tagging

import (
	"context"
	"fmt"

	"github.com/myfusionhelper/api/internal/helpers"
)

func init() {
	helpers.Register("clear_tags", func() helpers.Helper { return &ClearTags{} })
}

// ClearTags removes tags from a contact matching specified criteria
type ClearTags struct{}

func (h *ClearTags) GetName() string        { return "Clear Tags" }
func (h *ClearTags) GetType() string        { return "clear_tags" }
func (h *ClearTags) GetCategory() string    { return "tagging" }
func (h *ClearTags) GetDescription() string { return "Remove tags from a contact by IDs, category, or all" }
func (h *ClearTags) RequiresCRM() bool      { return true }
func (h *ClearTags) SupportedCRMs() []string { return nil }

func (h *ClearTags) GetConfigSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"mode": map[string]interface{}{
				"type":        "string",
				"enum":        []string{"specific", "all", "prefix", "category"},
				"description": "How to select tags to remove",
			},
			"tag_ids": map[string]interface{}{
				"type":        "array",
				"items":       map[string]interface{}{"type": "string"},
				"description": "Specific tag IDs to remove (for 'specific' mode)",
			},
			"prefix": map[string]interface{}{
				"type":        "string",
				"description": "Remove tags whose name starts with this prefix (for 'prefix' mode)",
			},
			"category": map[string]interface{}{
				"type":        "string",
				"description": "Remove tags in this category (for 'category' mode)",
			},
		},
		"required": []string{"mode"},
	}
}

func (h *ClearTags) ValidateConfig(config map[string]interface{}) error {
	mode, ok := config["mode"].(string)
	if !ok || mode == "" {
		return fmt.Errorf("mode is required")
	}

	validModes := map[string]bool{"specific": true, "all": true, "prefix": true, "category": true}
	if !validModes[mode] {
		return fmt.Errorf("invalid mode: %s", mode)
	}

	if mode == "specific" {
		tagIDs, ok := config["tag_ids"]
		if !ok {
			return fmt.Errorf("tag_ids is required for 'specific' mode")
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
		}
	}

	if mode == "prefix" {
		if _, ok := config["prefix"].(string); !ok || config["prefix"] == "" {
			return fmt.Errorf("prefix is required for 'prefix' mode")
		}
	}

	if mode == "category" {
		if _, ok := config["category"].(string); !ok || config["category"] == "" {
			return fmt.Errorf("category is required for 'category' mode")
		}
	}

	return nil
}

func (h *ClearTags) Execute(ctx context.Context, input helpers.HelperInput) (*helpers.HelperOutput, error) {
	mode := input.Config["mode"].(string)

	output := &helpers.HelperOutput{
		Actions: make([]helpers.HelperAction, 0),
		Logs:    make([]string, 0),
	}

	// Get the contact to find current tags
	contact, err := input.Connector.GetContact(ctx, input.ContactID)
	if err != nil {
		output.Message = fmt.Sprintf("Failed to get contact: %v", err)
		return output, err
	}

	var tagsToRemove []string

	switch mode {
	case "specific":
		tagsToRemove = extractStringSlice(input.Config["tag_ids"])

	case "all":
		for _, tag := range contact.Tags {
			tagsToRemove = append(tagsToRemove, tag.ID)
		}

	case "prefix":
		prefix := input.Config["prefix"].(string)
		for _, tag := range contact.Tags {
			if len(tag.Name) >= len(prefix) && tag.Name[:len(prefix)] == prefix {
				tagsToRemove = append(tagsToRemove, tag.ID)
			}
		}

	case "category":
		category := input.Config["category"].(string)
		// Need to get all tags to find category info
		allTags, err := input.Connector.GetTags(ctx)
		if err != nil {
			output.Message = fmt.Sprintf("Failed to get tags: %v", err)
			return output, err
		}

		tagCategories := make(map[string]string)
		for _, t := range allTags {
			tagCategories[t.ID] = t.Category
		}

		for _, tag := range contact.Tags {
			if tagCategories[tag.ID] == category {
				tagsToRemove = append(tagsToRemove, tag.ID)
			}
		}
	}

	if len(tagsToRemove) == 0 {
		output.Success = true
		output.Message = "No matching tags to remove"
		return output, nil
	}

	// Remove tags
	removed := 0
	for _, tagID := range tagsToRemove {
		err := input.Connector.RemoveTag(ctx, input.ContactID, tagID)
		if err != nil {
			output.Logs = append(output.Logs, fmt.Sprintf("Failed to remove tag %s: %v", tagID, err))
			continue
		}

		output.Actions = append(output.Actions, helpers.HelperAction{
			Type:   "tag_removed",
			Target: input.ContactID,
			Value:  tagID,
		})
		removed++
	}

	output.Success = removed > 0
	output.Message = fmt.Sprintf("Removed %d of %d tag(s) (%s mode)", removed, len(tagsToRemove), mode)
	output.Logs = append(output.Logs, output.Message)

	return output, nil
}
