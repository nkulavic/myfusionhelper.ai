package integration

import (
	"context"
	"fmt"
	"strings"

	"github.com/myfusionhelper/api/internal/helpers"
)

func init() {
	helpers.Register("gotowebinar", func() helpers.Helper { return &GoToWebinar{} })
}

// GoToWebinar registers a CRM contact for a GoToWebinar session.
// It prepares the registration payload and produces an API request for downstream
// processing by the execution layer which handles authentication and the HTTP call.
type GoToWebinar struct{}

func (h *GoToWebinar) GetName() string     { return "GoToWebinar" }
func (h *GoToWebinar) GetType() string     { return "gotowebinar" }
func (h *GoToWebinar) GetCategory() string { return "integration" }
func (h *GoToWebinar) GetDescription() string {
	return "Registers a CRM contact for a GoToWebinar session"
}
func (h *GoToWebinar) RequiresCRM() bool       { return true }
func (h *GoToWebinar) SupportedCRMs() []string { return nil }

func (h *GoToWebinar) GetConfigSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"organizer_key": map[string]interface{}{
				"type":        "string",
				"description": "The GoToWebinar organizer key",
			},
			"webinar_key": map[string]interface{}{
				"type":        "string",
				"description": "The GoToWebinar webinar key",
			},
			"apply_tag": map[string]interface{}{
				"type":        "string",
				"description": "Optional tag ID to apply to the contact after successful registration",
			},
			"service_connection_ids": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"gotowebinar": map[string]interface{}{
						"type":        "string",
						"description": "The service connection ID for GoToWebinar authentication",
					},
				},
				"description": "Service connection IDs for third-party authentication",
			},
		},
		"required": []string{"organizer_key", "webinar_key"},
	}
}

func (h *GoToWebinar) ValidateConfig(config map[string]interface{}) error {
	if _, ok := config["organizer_key"].(string); !ok || config["organizer_key"] == "" {
		return fmt.Errorf("organizer_key is required")
	}
	if _, ok := config["webinar_key"].(string); !ok || config["webinar_key"] == "" {
		return fmt.Errorf("webinar_key is required")
	}
	return nil
}

func (h *GoToWebinar) Execute(ctx context.Context, input helpers.HelperInput) (*helpers.HelperOutput, error) {
	organizerKey := input.Config["organizer_key"].(string)
	webinarKey := input.Config["webinar_key"].(string)

	output := &helpers.HelperOutput{
		Logs: make([]string, 0),
	}

	// Get service auth for GoToWebinar
	auth := input.ServiceAuths["gotowebinar"]
	if auth == nil {
		output.Message = "GoToWebinar service connection is not configured"
		return output, fmt.Errorf("gotowebinar service auth not found")
	}

	// Resolve the auth token
	authToken := auth.AccessToken
	if authToken == "" {
		authToken = auth.APIKey
	}
	if authToken == "" {
		output.Message = "GoToWebinar authentication token is missing"
		return output, fmt.Errorf("gotowebinar auth token not found")
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
		output.Message = "Contact email is required for GoToWebinar registration"
		return output, fmt.Errorf("contact email is empty")
	}

	// Build the GoToWebinar registration payload
	registrationPayload := map[string]interface{}{
		"firstName": firstName,
		"lastName":  lastName,
		"email":     email,
	}

	// Build the API URL
	apiURL := fmt.Sprintf("https://api.getgo.com/G2W/rest/v2/organizers/%s/webinars/%s/registrants",
		organizerKey, webinarKey)

	output.Success = true
	output.Message = fmt.Sprintf("GoToWebinar registration prepared for %s (webinar: %s)", email, webinarKey)
	output.Actions = []helpers.HelperAction{
		{
			Type:   "webhook_queued",
			Target: apiURL,
			Value: map[string]interface{}{
				"method":  "POST",
				"url":     apiURL,
				"payload": registrationPayload,
				"headers": map[string]string{
					"Authorization": "Bearer " + authToken,
					"Content-Type":  "application/json",
				},
				"auth_type": "bearer",
			},
		},
	}
	output.ModifiedData = map[string]interface{}{
		"webinar_key":   webinarKey,
		"organizer_key": organizerKey,
		"email":         email,
		"first_name":    firstName,
		"last_name":     lastName,
		"full_name":     strings.TrimSpace(firstName + " " + lastName),
		"api_url":       apiURL,
		"payload":       registrationPayload,
	}
	output.Logs = append(output.Logs, fmt.Sprintf("GoToWebinar registration for contact %s (%s) to webinar %s", input.ContactID, email, webinarKey))

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
