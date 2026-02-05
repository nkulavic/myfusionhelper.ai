package integration

import (
	"context"
	"fmt"
	"strings"

	"github.com/myfusionhelper/api/internal/helpers"
)

func init() {
	helpers.Register("zoom_webinar", func() helpers.Helper { return &ZoomWebinar{} })
}

// ZoomWebinar registers a contact for a Zoom webinar via the Zoom API.
// It prepares the registration payload and produces an API request for downstream
// processing by the execution layer which handles authentication and the HTTP call.
type ZoomWebinar struct{}

func (h *ZoomWebinar) GetName() string     { return "Zoom Webinar" }
func (h *ZoomWebinar) GetType() string     { return "zoom_webinar" }
func (h *ZoomWebinar) GetCategory() string { return "integration" }
func (h *ZoomWebinar) GetDescription() string {
	return "Register a contact for a Zoom webinar via the Zoom API"
}
func (h *ZoomWebinar) RequiresCRM() bool       { return true }
func (h *ZoomWebinar) SupportedCRMs() []string { return nil }

func (h *ZoomWebinar) GetConfigSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"webinar_id": map[string]interface{}{
				"type":        "string",
				"description": "The Zoom webinar ID to register the contact for",
			},
			"api_key": map[string]interface{}{
				"type":        "string",
				"description": "Zoom API key for authentication",
			},
			"api_secret": map[string]interface{}{
				"type":        "string",
				"description": "Zoom API secret for authentication",
			},
			"name_field": map[string]interface{}{
				"type":        "string",
				"description": "Contact field containing the registrant's name (defaults to FirstName + LastName)",
				"default":     "FirstName",
			},
			"email_field": map[string]interface{}{
				"type":        "string",
				"description": "Contact field containing the registrant's email",
				"default":     "Email",
			},
			"custom_questions": map[string]interface{}{
				"type": "array",
				"items": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"title": map[string]interface{}{"type": "string"},
						"value": map[string]interface{}{"type": "string"},
					},
				},
				"description": "Optional custom questions for Zoom webinar registration",
			},
		},
		"required": []string{"webinar_id", "api_key", "api_secret"},
	}
}

func (h *ZoomWebinar) ValidateConfig(config map[string]interface{}) error {
	if _, ok := config["webinar_id"].(string); !ok || config["webinar_id"] == "" {
		return fmt.Errorf("webinar_id is required")
	}
	if _, ok := config["api_key"].(string); !ok || config["api_key"] == "" {
		return fmt.Errorf("api_key is required")
	}
	if _, ok := config["api_secret"].(string); !ok || config["api_secret"] == "" {
		return fmt.Errorf("api_secret is required")
	}
	return nil
}

func (h *ZoomWebinar) Execute(ctx context.Context, input helpers.HelperInput) (*helpers.HelperOutput, error) {
	webinarID := input.Config["webinar_id"].(string)
	apiKey := input.Config["api_key"].(string)
	apiSecret := input.Config["api_secret"].(string)

	nameField := "FirstName"
	if nf, ok := input.Config["name_field"].(string); ok && nf != "" {
		nameField = nf
	}

	emailField := "Email"
	if ef, ok := input.Config["email_field"].(string); ok && ef != "" {
		emailField = ef
	}

	output := &helpers.HelperOutput{
		Logs: make([]string, 0),
	}

	// Get contact data
	contact, err := input.Connector.GetContact(ctx, input.ContactID)
	if err != nil {
		output.Message = fmt.Sprintf("Failed to get contact data: %v", err)
		return output, err
	}

	// Build field data map
	fieldData := map[string]string{
		"Id":        contact.ID,
		"FirstName": contact.FirstName,
		"LastName":  contact.LastName,
		"Email":     contact.Email,
		"Phone1":    contact.Phone,
		"Company":   contact.Company,
		"JobTitle":  contact.JobTitle,
		"full_name": strings.TrimSpace(contact.FirstName + " " + contact.LastName),
	}

	// Add custom fields
	if contact.CustomFields != nil {
		for key, value := range contact.CustomFields {
			fieldData[key] = fmt.Sprintf("%v", value)
		}
	}

	// Resolve email
	email := ""
	if val, exists := fieldData[emailField]; exists && val != "" {
		email = val
	}
	if email == "" {
		output.Message = fmt.Sprintf("Email field '%s' is empty for contact %s", emailField, input.ContactID)
		return output, fmt.Errorf("email field '%s' is empty", emailField)
	}

	// Resolve name
	firstName := contact.FirstName
	lastName := contact.LastName
	if nameField != "FirstName" {
		if val, exists := fieldData[nameField]; exists && val != "" {
			firstName = val
			lastName = ""
		}
	}

	// Build the Zoom webinar registration payload
	registrationPayload := map[string]interface{}{
		"email":      email,
		"first_name": firstName,
		"last_name":  lastName,
	}

	// Add custom questions if configured
	if customQuestions, ok := input.Config["custom_questions"].([]interface{}); ok {
		var questions []map[string]interface{}
		for _, q := range customQuestions {
			if qMap, ok := q.(map[string]interface{}); ok {
				questions = append(questions, qMap)
			}
		}
		if len(questions) > 0 {
			registrationPayload["custom_questions"] = questions
		}
	}

	// Build the API request for downstream processing
	apiURL := fmt.Sprintf("https://api.zoom.us/v2/webinars/%s/registrants", webinarID)

	output.Success = true
	output.Message = fmt.Sprintf("Zoom webinar registration prepared for %s (webinar: %s)", email, webinarID)
	output.Actions = []helpers.HelperAction{
		{
			Type:   "webhook_queued",
			Target: apiURL,
			Value: map[string]interface{}{
				"method":     "POST",
				"url":        apiURL,
				"payload":    registrationPayload,
				"api_key":    apiKey,
				"api_secret": apiSecret,
				"auth_type":  "zoom_jwt",
			},
		},
	}
	output.ModifiedData = map[string]interface{}{
		"webinar_id": webinarID,
		"email":      email,
		"first_name": firstName,
		"last_name":  lastName,
		"api_url":    apiURL,
		"payload":    registrationPayload,
	}
	output.Logs = append(output.Logs, fmt.Sprintf("Zoom webinar registration for contact %s (%s) to webinar %s", input.ContactID, email, webinarID))

	return output, nil
}
