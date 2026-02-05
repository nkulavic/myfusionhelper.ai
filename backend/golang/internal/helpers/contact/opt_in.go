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

	output := &helpers.HelperOutput{
		Logs: make([]string, 0),
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
		"email":      email,
		"contact_id": input.ContactID,
		"action":     "opt_in",
		"reason":     reason,
	}

	output.Success = true
	output.Message = fmt.Sprintf("Opt-in prepared for %s", email)
	output.Actions = []helpers.HelperAction{
		{
			Type:   "opt_in_requested",
			Target: input.ContactID,
			Value:  optInData,
		},
	}
	output.ModifiedData = optInData
	output.Logs = append(output.Logs, fmt.Sprintf("Opt-in requested for contact %s email %s", input.ContactID, email))

	return output, nil
}
