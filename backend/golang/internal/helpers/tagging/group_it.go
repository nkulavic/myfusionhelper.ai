package tagging

import (
	"context"
	"fmt"

	"github.com/myfusionhelper/api/internal/helpers"
)

func init() {
	helpers.Register("group_it", func() helpers.Helper { return &GroupIt{} })
}

// GroupIt creates and assigns tags based on a field value with a prefix
type GroupIt struct{}

func (h *GroupIt) GetName() string        { return "Group It" }
func (h *GroupIt) GetType() string        { return "group_it" }
func (h *GroupIt) GetCategory() string    { return "tagging" }
func (h *GroupIt) GetDescription() string { return "Create and assign tags based on a field value with a prefix" }
func (h *GroupIt) RequiresCRM() bool      { return true }
func (h *GroupIt) SupportedCRMs() []string { return nil }

func (h *GroupIt) GetConfigSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"field": map[string]interface{}{
				"type":        "string",
				"description": "The field whose value will be used to create the tag name",
			},
			"tag_prefix": map[string]interface{}{
				"type":        "string",
				"description": "Prefix to prepend to the field value for the tag name",
			},
		},
		"required": []string{"field", "tag_prefix"},
	}
}

func (h *GroupIt) ValidateConfig(config map[string]interface{}) error {
	if _, ok := config["field"].(string); !ok || config["field"] == "" {
		return fmt.Errorf("field is required")
	}
	if _, ok := config["tag_prefix"].(string); !ok || config["tag_prefix"] == "" {
		return fmt.Errorf("tag_prefix is required")
	}
	return nil
}

func (h *GroupIt) Execute(ctx context.Context, input helpers.HelperInput) (*helpers.HelperOutput, error) {
	field := input.Config["field"].(string)
	tagPrefix := input.Config["tag_prefix"].(string)

	output := &helpers.HelperOutput{
		Actions: make([]helpers.HelperAction, 0),
		Logs:    make([]string, 0),
	}

	// Get field value
	rawValue, err := input.Connector.GetContactFieldValue(ctx, input.ContactID, field)
	if err != nil {
		output.Message = fmt.Sprintf("Failed to read field '%s': %v", field, err)
		return output, err
	}

	strValue := fmt.Sprintf("%v", rawValue)
	if rawValue == nil || strValue == "" || strValue == "<nil>" {
		output.Success = true
		output.Message = fmt.Sprintf("Field '%s' is empty, no tag to create", field)
		return output, nil
	}

	// Build tag name: prefix + value
	tagName := tagPrefix + strValue

	// Look up tag by name to find its ID
	allTags, err := input.Connector.GetTags(ctx)
	if err != nil {
		output.Message = fmt.Sprintf("Failed to get tags: %v", err)
		return output, err
	}

	var tagID string
	for _, t := range allTags {
		if t.Name == tagName {
			tagID = t.ID
			break
		}
	}

	if tagID == "" {
		output.Success = false
		output.Message = fmt.Sprintf("Tag '%s' not found in CRM", tagName)
		output.Logs = append(output.Logs, output.Message)
		return output, nil
	}

	// Apply the tag
	err = input.Connector.ApplyTag(ctx, input.ContactID, tagID)
	if err != nil {
		output.Message = fmt.Sprintf("Failed to apply tag '%s': %v", tagName, err)
		return output, err
	}

	output.Success = true
	output.Message = fmt.Sprintf("Applied tag '%s' to contact", tagName)
	output.Actions = []helpers.HelperAction{
		{
			Type:   "tag_applied",
			Target: input.ContactID,
			Value:  tagID,
		},
	}
	output.Logs = append(output.Logs, fmt.Sprintf("Applied group tag '%s' (ID: %s) to contact %s based on field '%s'", tagName, tagID, input.ContactID, field))

	return output, nil
}
