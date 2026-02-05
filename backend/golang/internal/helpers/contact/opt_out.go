package contact

import (
	"context"
	"fmt"

	"github.com/myfusionhelper/api/internal/helpers"
)

func init() {
	helpers.Register("opt_out", func() helpers.Helper { return &OptOut{} })
}

// OptOut manages email opt-out for a contact
type OptOut struct{}

func (h *OptOut) GetName() string        { return "Opt Out" }
func (h *OptOut) GetType() string        { return "opt_out" }
func (h *OptOut) GetCategory() string    { return "contact" }
func (h *OptOut) GetDescription() string { return "Manage email opt-out for a contact" }
func (h *OptOut) RequiresCRM() bool      { return true }
func (h *OptOut) SupportedCRMs() []string { return []string{"keap"} }

func (h *OptOut) GetConfigSchema() map[string]interface{} {
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
				"description": "Reason for the opt-out",
			},
		},
		"required": []string{},
	}
}

func (h *OptOut) ValidateConfig(config map[string]interface{}) error {
	return nil
}

func (h *OptOut) Execute(ctx context.Context, input helpers.HelperInput) (*helpers.HelperOutput, error) {
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
		output.Message = fmt.Sprintf("Email field '%s' is empty, cannot opt out", emailField)
		output.Logs = append(output.Logs, output.Message)
		return output, nil
	}

	// Create opt-out action for the execution layer
	optOutData := map[string]interface{}{
		"email":      email,
		"contact_id": input.ContactID,
		"action":     "opt_out",
		"reason":     reason,
	}

	output.Success = true
	output.Message = fmt.Sprintf("Opt-out prepared for %s", email)
	output.Actions = []helpers.HelperAction{
		{
			Type:   "opt_out_requested",
			Target: input.ContactID,
			Value:  optOutData,
		},
	}
	output.ModifiedData = optOutData
	output.Logs = append(output.Logs, fmt.Sprintf("Opt-out requested for contact %s email %s", input.ContactID, email))

	return output, nil
}
