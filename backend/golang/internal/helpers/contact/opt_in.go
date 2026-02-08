package contact

import (
	"context"
	"fmt"

	"github.com/myfusionhelper/api/internal/helpers"
)

func init() {
	helpers.Register("opt_in", func() helpers.Helper { return &OptIn{} })
}

// OptIn manages email opt-in for a contact
type OptIn struct{}

func (h *OptIn) GetName() string        { return "Opt In" }
func (h *OptIn) GetType() string        { return "opt_in" }
func (h *OptIn) GetCategory() string    { return "contact" }
func (h *OptIn) GetDescription() string { return "Manage email opt-in for a contact" }
func (h *OptIn) RequiresCRM() bool      { return true }
func (h *OptIn) SupportedCRMs() []string { return []string{"keap"} }

func (h *OptIn) GetConfigSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"email_field": map[string]interface{}{
				"type":        "string",
				"description": "The field containing the email address",
				"default":     "email",
			},
			"reason": map[string]interface{}{
				"type":        "string",
				"description": "Reason for the opt-in",
			},
			"double_opt_in": map[string]interface{}{
				"type":        "boolean",
				"description": "Require email confirmation before opt-in (double opt-in)",
				"default":     false,
			},
			"confirmation_tag": map[string]interface{}{
				"type":        "string",
				"description": "Tag ID to apply when waiting for confirmation (double opt-in only)",
			},
			"confirmed_tag": map[string]interface{}{
				"type":        "string",
				"description": "Tag ID to apply when opt-in is confirmed",
			},
		},
		"required": []string{},
	}
}

func (h *OptIn) ValidateConfig(config map[string]interface{}) error {
	return nil
}

func (h *OptIn) Execute(ctx context.Context, input helpers.HelperInput) (*helpers.HelperOutput, error) {
	emailField := "email"
	if ef, ok := input.Config["email_field"].(string); ok && ef != "" {
		emailField = ef
	}
	reason := ""
	if r, ok := input.Config["reason"].(string); ok {
		reason = r
	}
	doubleOptIn := false
	if doi, ok := input.Config["double_opt_in"].(bool); ok {
		doubleOptIn = doi
	}
	confirmationTag, _ := input.Config["confirmation_tag"].(string)
	confirmedTag, _ := input.Config["confirmed_tag"].(string)

	output := &helpers.HelperOutput{
		Actions: make([]helpers.HelperAction, 0),
		Logs:    make([]string, 0),
	}

	// Get email value
	emailValue, err := input.Connector.GetContactFieldValue(ctx, input.ContactID, emailField)
	if err != nil {
		output.Message = fmt.Sprintf("Failed to read email field '%s': %v", emailField, err)
		return output, err
	}

	email := fmt.Sprintf("%v", emailValue)
	if emailValue == nil || email == "" || email == "<nil>" {
		output.Success = false
		output.Message = fmt.Sprintf("Email field '%s' is empty, cannot opt in", emailField)
		output.Logs = append(output.Logs, output.Message)
		return output, nil
	}

	// Create opt-in action for the execution layer
	optInData := map[string]interface{}{
		"email":          email,
		"contact_id":     input.ContactID,
		"action":         "opt_in",
		"reason":         reason,
		"double_opt_in":  doubleOptIn,
	}

	// Handle double opt-in mode
	if doubleOptIn {
		// Apply confirmation pending tag
		if confirmationTag != "" {
			err := input.Connector.ApplyTag(ctx, input.ContactID, confirmationTag)
			if err != nil {
				output.Logs = append(output.Logs, fmt.Sprintf("Warning: Failed to apply confirmation tag: %v", err))
			} else {
				output.Actions = append(output.Actions, helpers.HelperAction{
					Type:   "tag_applied",
					Target: input.ContactID,
					Value:  confirmationTag,
				})
				output.Logs = append(output.Logs, fmt.Sprintf("Applied confirmation pending tag '%s'", confirmationTag))
			}
		}

		optInData["confirmation_tag"] = confirmationTag
		optInData["confirmed_tag"] = confirmedTag
		output.Logs = append(output.Logs, "Double opt-in mode: confirmation email will be sent")
	} else {
		// Single opt-in: apply confirmed tag immediately
		if confirmedTag != "" {
			err := input.Connector.ApplyTag(ctx, input.ContactID, confirmedTag)
			if err != nil {
				output.Logs = append(output.Logs, fmt.Sprintf("Warning: Failed to apply confirmed tag: %v", err))
			} else {
				output.Actions = append(output.Actions, helpers.HelperAction{
					Type:   "tag_applied",
					Target: input.ContactID,
					Value:  confirmedTag,
				})
				output.Logs = append(output.Logs, fmt.Sprintf("Applied confirmed tag '%s'", confirmedTag))
			}
		}
	}

	output.Success = true
	if doubleOptIn {
		output.Message = fmt.Sprintf("Double opt-in initiated for %s (confirmation required)", email)
	} else {
		output.Message = fmt.Sprintf("Opt-in prepared for %s", email)
	}
	output.Actions = append(output.Actions, helpers.HelperAction{
		Type:   "opt_in_requested",
		Target: input.ContactID,
		Value:  optInData,
	})
	output.ModifiedData = optInData
	output.Logs = append(output.Logs, fmt.Sprintf("Opt-in requested for contact %s email %s (double opt-in: %v)", input.ContactID, email, doubleOptIn))

	return output, nil
}
