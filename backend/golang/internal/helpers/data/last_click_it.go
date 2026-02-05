package data

import (
	"context"
	"fmt"

	"github.com/myfusionhelper/api/internal/helpers"
)

func init() {
	helpers.Register("last_click_it", func() helpers.Helper { return &LastClickIt{} })
}

// LastClickIt retrieves the last email click date for a contact and stores it
// in a specified field. Queries email engagement stats via the CRM connector.
// Ported from legacy PHP last_click_it helper.
type LastClickIt struct{}

func (h *LastClickIt) GetName() string     { return "Last Click It" }
func (h *LastClickIt) GetType() string     { return "last_click_it" }
func (h *LastClickIt) GetCategory() string { return "data" }
func (h *LastClickIt) GetDescription() string {
	return "Retrieve the last email click date for a contact and save it to a field"
}
func (h *LastClickIt) RequiresCRM() bool       { return true }
func (h *LastClickIt) SupportedCRMs() []string { return nil }

func (h *LastClickIt) GetConfigSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"email_field": map[string]interface{}{
				"type":        "string",
				"description": "The email field to look up (e.g., Email, Email2, Email3)",
				"default":     "Email",
			},
			"save_to": map[string]interface{}{
				"type":        "string",
				"description": "The contact field to store the last click date",
			},
		},
		"required": []string{"save_to"},
	}
}

func (h *LastClickIt) ValidateConfig(config map[string]interface{}) error {
	if _, ok := config["save_to"].(string); !ok || config["save_to"] == "" {
		return fmt.Errorf("save_to field is required")
	}
	return nil
}

func (h *LastClickIt) Execute(ctx context.Context, input helpers.HelperInput) (*helpers.HelperOutput, error) {
	saveTo := input.Config["save_to"].(string)

	emailField := "Email"
	if ef, ok := input.Config["email_field"].(string); ok && ef != "" {
		emailField = ef
	}

	// Normalize email field names
	if emailField == "Email2" {
		emailField = "EmailAddress2"
	}
	if emailField == "Email3" {
		emailField = "EmailAddress3"
	}

	output := &helpers.HelperOutput{
		Logs: make([]string, 0),
	}

	// Get the email address from the contact
	emailValue, err := input.Connector.GetContactFieldValue(ctx, input.ContactID, emailField)
	if err != nil || emailValue == nil || fmt.Sprintf("%v", emailValue) == "" {
		output.Success = true
		output.Message = fmt.Sprintf("Email field '%s' is empty, nothing to look up", emailField)
		output.Logs = append(output.Logs, output.Message)
		return output, nil
	}

	// Query email engagement stats via the connector
	// Use a composite key to indicate we want email stats
	lastClickKey := fmt.Sprintf("_email_stats.%s.LastClickDate", fmt.Sprintf("%v", emailValue))
	lastClickDate, err := input.Connector.GetContactFieldValue(ctx, input.ContactID, lastClickKey)
	if err != nil {
		output.Logs = append(output.Logs, fmt.Sprintf("Email stats query not directly supported: %v", err))

		// Fallback: try a generic email stats field
		lastClickDate, err = input.Connector.GetContactFieldValue(ctx, input.ContactID, "LastClickDate")
		if err != nil {
			output.Success = true
			output.Message = "Could not retrieve email click stats"
			output.Logs = append(output.Logs, fmt.Sprintf("Failed to get LastClickDate: %v", err))
			return output, nil
		}
	}

	if lastClickDate == nil || fmt.Sprintf("%v", lastClickDate) == "" {
		output.Success = true
		output.Message = "No click date found for contact"
		output.Logs = append(output.Logs, output.Message)
		return output, nil
	}

	dateStr := fmt.Sprintf("%v", lastClickDate)

	// Save the date to the target field
	err = input.Connector.SetContactFieldValue(ctx, input.ContactID, saveTo, dateStr)
	if err != nil {
		output.Message = fmt.Sprintf("Failed to save last click date to '%s': %v", saveTo, err)
		return output, err
	}

	output.Success = true
	output.Message = fmt.Sprintf("Last click date saved to '%s'", saveTo)
	output.Actions = []helpers.HelperAction{
		{
			Type:   "field_updated",
			Target: saveTo,
			Value:  dateStr,
		},
	}
	output.ModifiedData = map[string]interface{}{
		saveTo: dateStr,
	}
	output.Logs = append(output.Logs, fmt.Sprintf("Last click date '%s' saved to '%s' for contact %s", dateStr, saveTo, input.ContactID))

	return output, nil
}
