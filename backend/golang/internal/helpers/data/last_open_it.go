package data

import (
	"context"
	"fmt"

	"github.com/myfusionhelper/api/internal/helpers"
)

func init() {
	helpers.Register("last_open_it", func() helpers.Helper { return &LastOpenIt{} })
}

// LastOpenIt retrieves the last email open date for a contact and stores it
// in a specified field. Queries email engagement stats via the CRM connector.
// Ported from legacy PHP last_open_it helper.
type LastOpenIt struct{}

func (h *LastOpenIt) GetName() string     { return "Last Open It" }
func (h *LastOpenIt) GetType() string     { return "last_open_it" }
func (h *LastOpenIt) GetCategory() string { return "data" }
func (h *LastOpenIt) GetDescription() string {
	return "Retrieve the last email open date for a contact and save it to a field"
}
func (h *LastOpenIt) RequiresCRM() bool       { return true }
func (h *LastOpenIt) SupportedCRMs() []string { return nil }

func (h *LastOpenIt) GetConfigSchema() map[string]interface{} {
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
				"description": "The contact field to store the last open date",
			},
		},
		"required": []string{"save_to"},
	}
}

func (h *LastOpenIt) ValidateConfig(config map[string]interface{}) error {
	if _, ok := config["save_to"].(string); !ok || config["save_to"] == "" {
		return fmt.Errorf("save_to field is required")
	}
	return nil
}

func (h *LastOpenIt) Execute(ctx context.Context, input helpers.HelperInput) (*helpers.HelperOutput, error) {
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
	lastOpenKey := fmt.Sprintf("_email_stats.%s.LastOpenDate", fmt.Sprintf("%v", emailValue))
	lastOpenDate, err := input.Connector.GetContactFieldValue(ctx, input.ContactID, lastOpenKey)
	if err != nil {
		output.Logs = append(output.Logs, fmt.Sprintf("Email stats query not directly supported: %v", err))

		// Fallback: try a generic email stats field
		lastOpenDate, err = input.Connector.GetContactFieldValue(ctx, input.ContactID, "LastOpenDate")
		if err != nil {
			output.Success = true
			output.Message = "Could not retrieve email open stats"
			output.Logs = append(output.Logs, fmt.Sprintf("Failed to get LastOpenDate: %v", err))
			return output, nil
		}
	}

	if lastOpenDate == nil || fmt.Sprintf("%v", lastOpenDate) == "" {
		output.Success = true
		output.Message = "No open date found for contact"
		output.Logs = append(output.Logs, output.Message)
		return output, nil
	}

	dateStr := fmt.Sprintf("%v", lastOpenDate)

	// Save the date to the target field
	err = input.Connector.SetContactFieldValue(ctx, input.ContactID, saveTo, dateStr)
	if err != nil {
		output.Message = fmt.Sprintf("Failed to save last open date to '%s': %v", saveTo, err)
		return output, err
	}

	output.Success = true
	output.Message = fmt.Sprintf("Last open date saved to '%s'", saveTo)
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
	output.Logs = append(output.Logs, fmt.Sprintf("Last open date '%s' saved to '%s' for contact %s", dateStr, saveTo, input.ContactID))

	return output, nil
}
