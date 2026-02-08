package tagging

import (
	"context"
	"fmt"

	"github.com/myfusionhelper/api/internal/connectors"
	"github.com/myfusionhelper/api/internal/helpers"
)

func init() {
	helpers.Register("count_it_tags", func() helpers.Helper { return &CountItTags{} })
}

// CountItTags counts the total number of contacts that have a specific tag applied,
// and optionally applies threshold-based tags. Also determines the current contact's
// position in the tag assignment list.
// Ported from legacy PHP count_it_tags helper.
type CountItTags struct{}

func (h *CountItTags) GetName() string     { return "Count It Tags" }
func (h *CountItTags) GetType() string     { return "count_it_tags" }
func (h *CountItTags) GetCategory() string { return "tagging" }
func (h *CountItTags) GetDescription() string {
	return "Count total contacts with a specific tag and apply threshold-based tags"
}
func (h *CountItTags) RequiresCRM() bool       { return true }
func (h *CountItTags) SupportedCRMs() []string { return nil }

func (h *CountItTags) GetConfigSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"tag_id": map[string]interface{}{
				"type":        "string",
				"description": "The tag ID to count contacts for",
			},
			"threshold": map[string]interface{}{
				"type":        "number",
				"description": "The count threshold to check against",
			},
			"threshold_met_tag": map[string]interface{}{
				"type":        "string",
				"description": "Tag ID to apply when threshold is met (use 'no_tag' to skip)",
			},
			"threshold_not_met_tag": map[string]interface{}{
				"type":        "string",
				"description": "Tag ID to apply when threshold is not met (use 'no_tag' to skip)",
			},
			"save_count_to": map[string]interface{}{
				"type":        "string",
				"description": "Optional field to store the total count",
			},
			"save_position_to": map[string]interface{}{
				"type":        "string",
				"description": "Optional field to store the contact's position in the tag list",
			},
		},
		"required": []string{"tag_id"},
	}
}

func (h *CountItTags) ValidateConfig(config map[string]interface{}) error {
	if _, ok := config["tag_id"].(string); !ok || config["tag_id"] == "" {
		return fmt.Errorf("tag_id is required")
	}
	return nil
}

func (h *CountItTags) Execute(ctx context.Context, input helpers.HelperInput) (*helpers.HelperOutput, error) {
	tagID := input.Config["tag_id"].(string)

	output := &helpers.HelperOutput{
		Actions: make([]helpers.HelperAction, 0),
		Logs:    make([]string, 0),
	}

	// Get all contacts with this tag using the connector
	contactList, err := input.Connector.GetContacts(ctx, connectors.QueryOptions{TagID: tagID, Limit: 1000})
	if err != nil {
		output.Message = fmt.Sprintf("Failed to query contacts with tag %s: %v", tagID, err)
		return output, err
	}

	totalCount := contactList.Total
	output.Logs = append(output.Logs, fmt.Sprintf("Total contacts with tag %s: %d", tagID, totalCount))

	// Check threshold and apply tags
	threshold := 0.0
	if t, ok := input.Config["threshold"].(float64); ok {
		threshold = t
	}

	if threshold > 0 {
		if totalCount >= int(threshold) {
			// Threshold met
			if metTag, ok := input.Config["threshold_met_tag"].(string); ok && metTag != "" && metTag != "no_tag" {
				tagErr := input.Connector.ApplyTag(ctx, input.ContactID, metTag)
				if tagErr != nil {
					output.Logs = append(output.Logs, fmt.Sprintf("Failed to apply threshold_met_tag %s: %v", metTag, tagErr))
				} else {
					output.Actions = append(output.Actions, helpers.HelperAction{
						Type:   "tag_applied",
						Target: input.ContactID,
						Value:  metTag,
					})
					output.Logs = append(output.Logs, fmt.Sprintf("Threshold met (%d >= %d), applied tag %s", totalCount, int(threshold), metTag))
				}
			}
		} else {
			// Threshold not met
			if notMetTag, ok := input.Config["threshold_not_met_tag"].(string); ok && notMetTag != "" && notMetTag != "no_tag" {
				tagErr := input.Connector.ApplyTag(ctx, input.ContactID, notMetTag)
				if tagErr != nil {
					output.Logs = append(output.Logs, fmt.Sprintf("Failed to apply threshold_not_met_tag %s: %v", notMetTag, tagErr))
				} else {
					output.Actions = append(output.Actions, helpers.HelperAction{
						Type:   "tag_applied",
						Target: input.ContactID,
						Value:  notMetTag,
					})
					output.Logs = append(output.Logs, fmt.Sprintf("Threshold not met (%d < %d), applied tag %s", totalCount, int(threshold), notMetTag))
				}
			}
		}
	}

	// Find the current contact's position in the tag list
	position := 0
	for i, c := range contactList.Contacts {
		if c.ID == input.ContactID {
			position = i + 1
			break
		}
	}

	// Save count and position to fields if configured
	modifiedData := map[string]interface{}{
		"total_count": totalCount,
		"position":    position,
	}

	if saveCountTo, ok := input.Config["save_count_to"].(string); ok && saveCountTo != "" {
		countStr := fmt.Sprintf("%d", totalCount)
		setErr := input.Connector.SetContactFieldValue(ctx, input.ContactID, saveCountTo, countStr)
		if setErr != nil {
			output.Logs = append(output.Logs, fmt.Sprintf("Failed to save count to '%s': %v", saveCountTo, setErr))
		} else {
			output.Actions = append(output.Actions, helpers.HelperAction{
				Type:   "field_updated",
				Target: saveCountTo,
				Value:  countStr,
			})
		}
	}

	if savePosTo, ok := input.Config["save_position_to"].(string); ok && savePosTo != "" && position > 0 {
		posStr := fmt.Sprintf("%d", position)
		setErr := input.Connector.SetContactFieldValue(ctx, input.ContactID, savePosTo, posStr)
		if setErr != nil {
			output.Logs = append(output.Logs, fmt.Sprintf("Failed to save position to '%s': %v", savePosTo, setErr))
		} else {
			output.Actions = append(output.Actions, helpers.HelperAction{
				Type:   "field_updated",
				Target: savePosTo,
				Value:  posStr,
			})
		}
	}

	output.Success = true
	output.Message = fmt.Sprintf("Tag %s has %d contacts", tagID, totalCount)
	if position > 0 {
		output.Message += fmt.Sprintf(", contact is #%d", position)
	}
	output.ModifiedData = modifiedData

	return output, nil
}
