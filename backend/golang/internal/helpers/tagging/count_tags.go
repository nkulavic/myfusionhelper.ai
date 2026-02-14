package tagging

import (
	"context"
	"fmt"

	"github.com/myfusionhelper/api/internal/helpers"
)

func init() {
	helpers.Register("count_tags", func() helpers.Helper { return &CountTags{} })
}

// NewCountTags creates a new CountTags helper instance
func NewCountTags() helpers.Helper { return &CountTags{} }

// CountTags counts the tags on a contact and stores the count
type CountTags struct{}

func (h *CountTags) GetName() string        { return "Count Tags" }
func (h *CountTags) GetType() string        { return "count_tags" }
func (h *CountTags) GetCategory() string    { return "tagging" }
func (h *CountTags) GetDescription() string { return "Count tags on a contact and store the count in a field" }
func (h *CountTags) RequiresCRM() bool      { return true }
func (h *CountTags) SupportedCRMs() []string { return nil }

func (h *CountTags) GetConfigSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"target_field": map[string]interface{}{
				"type":        "string",
				"description": "The field to store the tag count",
			},
			"category": map[string]interface{}{
				"type":        "string",
				"description": "Optional: only count tags in this category",
			},
		},
		"required": []string{"target_field"},
	}
}

func (h *CountTags) ValidateConfig(config map[string]interface{}) error {
	if _, ok := config["target_field"].(string); !ok || config["target_field"] == "" {
		return fmt.Errorf("target_field is required")
	}
	return nil
}

func (h *CountTags) Execute(ctx context.Context, input helpers.HelperInput) (*helpers.HelperOutput, error) {
	targetField := input.Config["target_field"].(string)
	category := ""
	if c, ok := input.Config["category"].(string); ok {
		category = c
	}

	output := &helpers.HelperOutput{
		Logs: make([]string, 0),
	}

	// Get contact to find current tags
	contact, err := input.Connector.GetContact(ctx, input.ContactID)
	if err != nil {
		output.Message = fmt.Sprintf("Failed to get contact: %v", err)
		return output, err
	}

	count := 0

	if category == "" {
		// Count all tags
		count = len(contact.Tags)
	} else {
		// Need to get full tag definitions to check category
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
				count++
			}
		}
	}

	countStr := fmt.Sprintf("%d", count)

	// Set target field
	err = input.Connector.SetContactFieldValue(ctx, input.ContactID, targetField, countStr)
	if err != nil {
		output.Message = fmt.Sprintf("Failed to set count field '%s': %v", targetField, err)
		return output, err
	}

	output.Success = true
	if category != "" {
		output.Message = fmt.Sprintf("Counted %d tags in category '%s'", count, category)
	} else {
		output.Message = fmt.Sprintf("Counted %d total tags", count)
	}
	output.Actions = []helpers.HelperAction{
		{
			Type:   "field_updated",
			Target: targetField,
			Value:  countStr,
		},
	}
	output.ModifiedData = map[string]interface{}{
		targetField: countStr,
	}
	output.Logs = append(output.Logs, fmt.Sprintf("Tag count for contact %s: %d (category: %s), stored in '%s'", input.ContactID, count, category, targetField))

	return output, nil
}
