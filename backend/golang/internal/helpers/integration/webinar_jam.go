package integration

import (
	"context"
	"fmt"

	"github.com/myfusionhelper/api/internal/helpers"
)

// NewWebinarJam creates a new WebinarJam helper instance
func NewWebinarJam() helpers.Helper { return &WebinarJam{} }

func init() {
	helpers.Register("webinar_jam", func() helpers.Helper { return &WebinarJam{} })
}

// WebinarJam registers a CRM contact for a WebinarJam session.
// It prepares the registration payload and produces an API request for downstream
// processing by the execution layer which handles authentication and the HTTP call.
type WebinarJam struct{}

func (h *WebinarJam) GetName() string     { return "WebinarJam" }
func (h *WebinarJam) GetType() string     { return "webinar_jam" }
func (h *WebinarJam) GetCategory() string { return "integration" }
func (h *WebinarJam) GetDescription() string {
	return "Registers a CRM contact for a WebinarJam session"
}
func (h *WebinarJam) RequiresCRM() bool       { return true }
func (h *WebinarJam) SupportedCRMs() []string { return nil }

func (h *WebinarJam) GetConfigSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"webinar_id": map[string]interface{}{
				"type":        "string",
				"description": "The WebinarJam webinar ID",
			},
			"schedule": map[string]interface{}{
				"type":        "string",
				"description": "Optional schedule identifier for the webinar session",
			},
			"apply_tag": map[string]interface{}{
				"type":        "string",
				"description": "Optional tag ID to apply to the contact after successful registration",
			},
			"service_connection_ids": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"webinarjam": map[string]interface{}{
						"type":        "string",
						"description": "The service connection ID for WebinarJam authentication",
					},
				},
				"description": "Service connection IDs for third-party authentication",
			},
		},
		"required": []string{"webinar_id"},
	}
}

func (h *WebinarJam) ValidateConfig(config map[string]interface{}) error {
	if _, ok := config["webinar_id"].(string); !ok || config["webinar_id"] == "" {
		return fmt.Errorf("webinar_id is required")
	}
	return nil
}

func (h *WebinarJam) Execute(ctx context.Context, input helpers.HelperInput) (*helpers.HelperOutput, error) {
	webinarID := input.Config["webinar_id"].(string)

	output := &helpers.HelperOutput{
		Logs: make([]string, 0),
	}

	// Get service auth for WebinarJam
	auth := input.ServiceAuths["webinarjam"]
	if auth == nil {
		output.Message = "WebinarJam service connection is not configured"
		return output, fmt.Errorf("webinarjam service auth not found")
	}

	apiKey := auth.APIKey
	if apiKey == "" {
		output.Message = "WebinarJam API key is missing"
		return output, fmt.Errorf("webinarjam api key not found")
	}

	// Get contact data for registration
	contact := input.ContactData
	if contact == nil {
		var err error
		contact, err = input.Connector.GetContact(ctx, input.ContactID)
		if err != nil {
			output.Message = fmt.Sprintf("Failed to get contact data: %v", err)
			return output, err
		}
	}

	firstName := contact.FirstName
	lastName := contact.LastName
	email := contact.Email

	if email == "" {
		output.Message = "Contact email is required for WebinarJam registration"
		return output, fmt.Errorf("contact email is empty")
	}

	// Build the WebinarJam registration payload
	registrationPayload := map[string]interface{}{
		"api_key":    apiKey,
		"webinar_id": webinarID,
		"first_name": firstName,
		"last_name":  lastName,
		"email":      email,
	}

	// Include schedule if configured
	if schedule, ok := input.Config["schedule"].(string); ok && schedule != "" {
		registrationPayload["schedule"] = schedule
	}

	// Build the API URL
	apiURL := "https://api.webinarjam.com/webinarjam/register"

	output.Success = true
	output.Message = fmt.Sprintf("WebinarJam registration prepared for %s (webinar: %s)", email, webinarID)
	output.Actions = []helpers.HelperAction{
		{
			Type:   "webhook_queued",
			Target: apiURL,
			Value: map[string]interface{}{
				"method":  "POST",
				"url":     apiURL,
				"payload": registrationPayload,
				"headers": map[string]string{
					"Content-Type": "application/json",
				},
				"auth_type": "api_key_in_body",
			},
		},
	}
	output.ModifiedData = map[string]interface{}{
		"webinar_id": webinarID,
		"email":      email,
		"first_name": firstName,
		"last_name":  lastName,
		"api_url":    apiURL,
	}
	output.Logs = append(output.Logs, fmt.Sprintf("WebinarJam registration for contact %s (%s) to webinar %s", input.ContactID, email, webinarID))

	// Apply tag if configured
	if applyTag, ok := input.Config["apply_tag"].(string); ok && applyTag != "" {
		err := input.Connector.ApplyTag(ctx, input.ContactID, applyTag)
		if err != nil {
			output.Logs = append(output.Logs, fmt.Sprintf("Failed to apply tag '%s': %v", applyTag, err))
		} else {
			output.Actions = append(output.Actions, helpers.HelperAction{
				Type:   "tag_applied",
				Target: input.ContactID,
				Value:  applyTag,
			})
			output.Logs = append(output.Logs, fmt.Sprintf("Applied tag '%s' to contact %s", applyTag, input.ContactID))
		}
	}

	return output, nil
}
