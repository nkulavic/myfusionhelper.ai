package contact

import (
	"context"
	"fmt"

	"github.com/myfusionhelper/api/internal/connectors"
	"github.com/myfusionhelper/api/internal/helpers"
)

func init() {
	helpers.Register("contact_updater", func() helpers.Helper { return &ContactUpdater{} })
}

// ContactUpdater updates contact fields with specified values and optionally
// triggers goals for secondary contacts.
type ContactUpdater struct{}

func (h *ContactUpdater) GetName() string     { return "Contact Updater" }
func (h *ContactUpdater) GetType() string     { return "contact_updater" }
func (h *ContactUpdater) GetCategory() string { return "contact" }
func (h *ContactUpdater) GetDescription() string {
	return "Updates contact fields with specified values and optionally triggers goals for secondary contacts"
}
func (h *ContactUpdater) RequiresCRM() bool       { return true }
func (h *ContactUpdater) SupportedCRMs() []string { return nil }

func (h *ContactUpdater) GetConfigSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"fields": map[string]interface{}{
				"type":        "object",
				"description": "Map of field names to values to set on the contact",
			},
			"secondary_contact_ids": map[string]interface{}{
				"type":        "array",
				"items":       map[string]interface{}{"type": "string"},
				"description": "Additional contact IDs to trigger goals for",
			},
		},
		"required": []string{"fields"},
	}
}

func (h *ContactUpdater) ValidateConfig(config map[string]interface{}) error {
	fields, ok := config["fields"]
	if !ok {
		return fmt.Errorf("fields is required")
	}
	if _, ok := fields.(map[string]interface{}); !ok {
		return fmt.Errorf("fields must be a map of field names to values")
	}
	return nil
}

func (h *ContactUpdater) Execute(ctx context.Context, input helpers.HelperInput) (*helpers.HelperOutput, error) {
	fields, _ := input.Config["fields"].(map[string]interface{})

	output := &helpers.HelperOutput{
		Logs: make([]string, 0),
	}

	// Build UpdateContactInput with custom fields
	updateInput := connectors.UpdateContactInput{
		CustomFields: make(map[string]interface{}),
	}
	for fieldName, value := range fields {
		updateInput.CustomFields[fieldName] = value
	}

	// Update the primary contact
	_, err := input.Connector.UpdateContact(ctx, input.ContactID, updateInput)
	if err != nil {
		output.Message = fmt.Sprintf("Failed to update contact: %v", err)
		return output, err
	}

	// Log each field update as an action
	actions := make([]helpers.HelperAction, 0, len(fields))
	for fieldName, value := range fields {
		actions = append(actions, helpers.HelperAction{
			Type:   "field_updated",
			Target: fieldName,
			Value:  value,
		})
		output.Logs = append(output.Logs, fmt.Sprintf("Set field '%s' to '%v' on contact %s", fieldName, value, input.ContactID))
	}

	// Handle secondary contacts - trigger goals for each
	if secondaryRaw, ok := input.Config["secondary_contact_ids"]; ok {
		if secondaryIDs, ok := secondaryRaw.([]interface{}); ok {
			for _, idRaw := range secondaryIDs {
				if contactID, ok := idRaw.(string); ok && contactID != "" {
					err := input.Connector.AchieveGoal(ctx, contactID, "contact_updated", "contact_updater")
					if err != nil {
						output.Logs = append(output.Logs, fmt.Sprintf("Failed to trigger goal for secondary contact %s: %v", contactID, err))
					} else {
						actions = append(actions, helpers.HelperAction{
							Type:   "goal_achieved",
							Target: contactID,
							Value:  "contact_updated",
						})
						output.Logs = append(output.Logs, fmt.Sprintf("Triggered goal 'contact_updated' for secondary contact %s", contactID))
					}
				}
			}
		}
	}

	output.Success = true
	output.Message = fmt.Sprintf("Updated %d fields on contact %s", len(fields), input.ContactID)
	output.Actions = actions
	output.ModifiedData = fields

	return output, nil
}
