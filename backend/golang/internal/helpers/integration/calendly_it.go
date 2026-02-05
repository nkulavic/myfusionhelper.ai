package integration

import (
	"context"
	"fmt"
	"strings"

	"github.com/myfusionhelper/api/internal/helpers"
)

func init() {
	helpers.Register("calendly_it", func() helpers.Helper { return &CalendlyIt{} })
}

// CalendlyIt creates a Calendly scheduling link for a contact via the Calendly API.
// It prepares the invite payload and produces an API request for downstream
// processing by the execution layer which handles the HTTP call.
type CalendlyIt struct{}

func (h *CalendlyIt) GetName() string     { return "Calendly It" }
func (h *CalendlyIt) GetType() string     { return "calendly_it" }
func (h *CalendlyIt) GetCategory() string { return "integration" }
func (h *CalendlyIt) GetDescription() string {
	return "Create a Calendly scheduling link for a contact via the Calendly API"
}
func (h *CalendlyIt) RequiresCRM() bool       { return true }
func (h *CalendlyIt) SupportedCRMs() []string { return nil }

func (h *CalendlyIt) GetConfigSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"event_type_uri": map[string]interface{}{
				"type":        "string",
				"description": "The Calendly event type URI (e.g., https://api.calendly.com/event_types/XXXX)",
			},
			"api_token": map[string]interface{}{
				"type":        "string",
				"description": "Calendly personal access token for authentication",
			},
			"email_field": map[string]interface{}{
				"type":        "string",
				"description": "Contact field containing the invitee email address",
				"default":     "Email",
			},
			"name_field": map[string]interface{}{
				"type":        "string",
				"description": "Contact field containing the invitee name",
				"default":     "FirstName",
			},
			"result_field": map[string]interface{}{
				"type":        "string",
				"description": "Optional contact field to store the generated scheduling link",
			},
		},
		"required": []string{"event_type_uri", "api_token"},
	}
}

func (h *CalendlyIt) ValidateConfig(config map[string]interface{}) error {
	if _, ok := config["event_type_uri"].(string); !ok || config["event_type_uri"] == "" {
		return fmt.Errorf("event_type_uri is required")
	}
	if _, ok := config["api_token"].(string); !ok || config["api_token"] == "" {
		return fmt.Errorf("api_token is required")
	}
	return nil
}

func (h *CalendlyIt) Execute(ctx context.Context, input helpers.HelperInput) (*helpers.HelperOutput, error) {
	eventTypeURI := input.Config["event_type_uri"].(string)
	apiToken := input.Config["api_token"].(string)

	emailField := "Email"
	if ef, ok := input.Config["email_field"].(string); ok && ef != "" {
		emailField = ef
	}

	nameField := "FirstName"
	if nf, ok := input.Config["name_field"].(string); ok && nf != "" {
		nameField = nf
	}

	resultField := ""
	if rf, ok := input.Config["result_field"].(string); ok {
		resultField = rf
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
	name := ""
	if val, exists := fieldData[nameField]; exists && val != "" {
		name = val
	}
	if name == "" {
		name = strings.TrimSpace(contact.FirstName + " " + contact.LastName)
	}

	// Build the Calendly scheduling link request payload
	invitePayload := map[string]interface{}{
		"max_event_count": 1,
		"email":           email,
		"name":            name,
	}

	// Build the API request for downstream processing
	apiURL := "https://api.calendly.com/scheduling_links"

	calendlyRequest := map[string]interface{}{
		"owner":           eventTypeURI,
		"owner_type":      "EventType",
		"max_event_count": 1,
	}

	output.Success = true
	output.Message = fmt.Sprintf("Calendly scheduling link request prepared for %s", email)
	output.Actions = []helpers.HelperAction{
		{
			Type:   "webhook_queued",
			Target: apiURL,
			Value: map[string]interface{}{
				"method":  "POST",
				"url":     apiURL,
				"payload": calendlyRequest,
				"headers": map[string]string{
					"Authorization": "Bearer " + apiToken,
					"Content-Type":  "application/json",
				},
				"auth_type": "bearer",
			},
		},
	}
	output.ModifiedData = map[string]interface{}{
		"event_type_uri": eventTypeURI,
		"email":          email,
		"name":           name,
		"api_url":        apiURL,
		"invite":         invitePayload,
	}

	if resultField != "" {
		output.ModifiedData["result_field"] = resultField
	}

	output.Logs = append(output.Logs, fmt.Sprintf("Calendly scheduling link request for contact %s (%s) with event type %s", input.ContactID, email, eventTypeURI))

	return output, nil
}
