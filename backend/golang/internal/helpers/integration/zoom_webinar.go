package integration

import (
	"context"
	"fmt"

	"github.com/myfusionhelper/api/internal/helpers"
)

// NewZoomWebinar creates a new ZoomWebinar helper instance
func NewZoomWebinar() helpers.Helper { return &ZoomWebinar{} }

func init() {
	helpers.Register("zoom_webinar", func() helpers.Helper { return &ZoomWebinar{} })
}

// ZoomWebinar creates Zoom webinars and registers contacts via Zoom API v2.
// COMPLETE REWRITE with OAuth implementation via ServiceAuths["zoom"].
type ZoomWebinar struct{}

func (h *ZoomWebinar) GetName() string     { return "Zoom Webinar" }
func (h *ZoomWebinar) GetType() string     { return "zoom_webinar" }
func (h *ZoomWebinar) GetCategory() string { return "integration" }
func (h *ZoomWebinar) GetDescription() string {
	return "Create Zoom webinars and register contacts via the Zoom API"
}
func (h *ZoomWebinar) RequiresCRM() bool       { return true }
func (h *ZoomWebinar) SupportedCRMs() []string { return nil }

func (h *ZoomWebinar) GetConfigSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"action": map[string]interface{}{
				"type":        "string",
				"enum":        []string{"create", "register"},
				"description": "Action: 'create' a webinar or 'register' contact for existing webinar",
				"default":     "create",
			},
			"user_id": map[string]interface{}{
				"type":        "string",
				"description": "Zoom user ID (email or user ID) - required for create action",
			},
			"webinar_id": map[string]interface{}{
				"type":        "string",
				"description": "Existing webinar ID - required for register action",
			},
			"topic": map[string]interface{}{
				"type":        "string",
				"description": "Webinar topic/title - for create action",
			},
			"start_time": map[string]interface{}{
				"type":        "string",
				"description": "Webinar start time (ISO 8601 format: 2024-01-15T10:00:00Z) - for create action",
			},
			"duration": map[string]interface{}{
				"type":        "number",
				"description": "Webinar duration in minutes (default: 60)",
				"default":     60,
			},
			"timezone": map[string]interface{}{
				"type":        "string",
				"description": "Timezone (e.g., America/Los_Angeles, UTC)",
				"default":     "UTC",
			},
			"password": map[string]interface{}{
				"type":        "string",
				"description": "Webinar password (optional)",
			},
			"name_field": map[string]interface{}{
				"type":        "string",
				"description": "Contact field for registrant name (defaults to FirstName + LastName)",
				"default":     "FirstName",
			},
			"email_field": map[string]interface{}{
				"type":        "string",
				"description": "Contact field for registrant email",
				"default":     "Email",
			},
			"save_join_url_to": map[string]interface{}{
				"type":        "string",
				"description": "CRM field to save join URL to (optional)",
			},
			"approval_type": map[string]interface{}{
				"type":        "number",
				"description": "Registration approval: 0=auto approve, 1=manual, 2=no registration (default: 0)",
				"default":     0,
			},
			"registration_type": map[string]interface{}{
				"type":        "number",
				"description": "Registration type: 1=once, 2=each occurrence, 3=series (default: 1)",
				"default":     1,
			},
			"auto_recording": map[string]interface{}{
				"type":        "string",
				"enum":        []string{"none", "local", "cloud"},
				"description": "Auto recording setting (default: cloud)",
				"default":     "cloud",
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
				"description": "Custom questions for webinar registration",
			},
		},
		"required": []string{"action"},
	}
}

func (h *ZoomWebinar) ValidateConfig(config map[string]interface{}) error {
	action, ok := config["action"].(string)
	if !ok || action == "" {
		return fmt.Errorf("action is required")
	}

	if action == "create" {
		if _, ok := config["user_id"].(string); !ok || config["user_id"] == "" {
			return fmt.Errorf("user_id is required for create action")
		}
	} else if action == "register" {
		if _, ok := config["webinar_id"].(string); !ok || config["webinar_id"] == "" {
			return fmt.Errorf("webinar_id is required for register action")
		}
	} else {
		return fmt.Errorf("action must be 'create' or 'register'")
	}

	return nil
}

func (h *ZoomWebinar) Execute(ctx context.Context, input helpers.HelperInput) (*helpers.HelperOutput, error) {
	output := &helpers.HelperOutput{
		Logs: make([]string, 0),
	}

	// Get Zoom auth from ServiceAuths
	zoomAuth, ok := input.ServiceAuths["zoom"]
	if !ok || zoomAuth == nil {
		output.Message = "Zoom service connection not configured"
		return output, fmt.Errorf("zoom service connection required")
	}

	action := input.Config["action"].(string)

	// Get contact data
	contact, err := input.Connector.GetContact(ctx, input.ContactID)
	if err != nil {
		output.Message = fmt.Sprintf("Failed to get contact: %v", err)
		return output, err
	}

	// Build field data map
	fieldData := buildFieldDataMap(contact)

	// Get email
	emailField := getStringConfigValue(input.Config, "email_field", "Email")
	email, exists := fieldData[emailField]
	if !exists || email == "" {
		output.Message = fmt.Sprintf("Email field '%s' is empty", emailField)
		return output, fmt.Errorf("email required")
	}

	// Get name
	nameField := getStringConfigValue(input.Config, "name_field", "FirstName")
	firstName := contact.FirstName
	lastName := contact.LastName
	if nameField != "FirstName" {
		if val, exists := fieldData[nameField]; exists && val != "" {
			firstName = val
			lastName = ""
		}
	}

	if action == "create" {
		return h.createWebinar(ctx, input, output, zoomAuth, contact, email, firstName, lastName)
	}

	return h.registerForWebinar(ctx, input, output, zoomAuth, contact, email, firstName, lastName)
}

func (h *ZoomWebinar) createWebinar(ctx context.Context, input helpers.HelperInput, output *helpers.HelperOutput,
	zoomAuth interface{}, contact interface{}, email, firstName, lastName string) (*helpers.HelperOutput, error) {

	userID := input.Config["user_id"].(string)
	topic := getStringConfigValue(input.Config, "topic", "Webinar")
	duration := getIntConfigValue(input.Config, "duration", 60)
	timezone := getStringConfigValue(input.Config, "timezone", "UTC")
	password := getStringConfigValue(input.Config, "password", "")
	approvalType := getIntConfigValue(input.Config, "approval_type", 0)
	registrationType := getIntConfigValue(input.Config, "registration_type", 1)
	autoRecording := getStringConfigValue(input.Config, "auto_recording", "cloud")
	startTime := getStringConfigValue(input.Config, "start_time", "")

	// Build webinar creation payload
	payload := map[string]interface{}{
		"topic":    topic,
		"type":     5, // Webinar
		"duration": duration,
		"timezone": timezone,
		"settings": map[string]interface{}{
			"approval_type":       approvalType,
			"registration_type":   registrationType,
			"auto_recording":      autoRecording,
			"registrants_email_notification": true,
		},
	}

	if startTime != "" {
		payload["start_time"] = startTime
	}

	if password != "" {
		payload["password"] = password
	}

	// Make API request
	apiURL := fmt.Sprintf("https://api.zoom.us/v2/users/%s/webinars", userID)
	webinarResp, err := makeZoomAPIRequest(ctx, "POST", apiURL, payload, zoomAuth)
	if err != nil {
		output.Message = fmt.Sprintf("Failed to create webinar: %v", err)
		return output, err
	}

	webinarID := getStringFromMap(webinarResp, "id", "")
	joinURL := getStringFromMap(webinarResp, "join_url", "")

	output.Logs = append(output.Logs, fmt.Sprintf("Created Zoom webinar: %s (ID: %s)", topic, webinarID))

	// Optionally save join URL to CRM field
	if saveField, ok := input.Config["save_join_url_to"].(string); ok && saveField != "" && joinURL != "" {
		if err := input.Connector.SetContactFieldValue(ctx, input.ContactID, saveField, joinURL); err != nil {
			output.Logs = append(output.Logs, fmt.Sprintf("Warning: failed to save join URL to field '%s': %v", saveField, err))
		} else {
			output.Logs = append(output.Logs, fmt.Sprintf("Saved join URL to field '%s'", saveField))
		}
	}

	output.Success = true
	output.Message = fmt.Sprintf("Created webinar '%s'", topic)
	output.ModifiedData = map[string]interface{}{
		"webinar_id": webinarID,
		"join_url":   joinURL,
		"topic":      topic,
	}
	output.Actions = []helpers.HelperAction{
		{
			Type:   "webinar_created",
			Target: webinarID,
			Value:  joinURL,
		},
	}

	return output, nil
}

func (h *ZoomWebinar) registerForWebinar(ctx context.Context, input helpers.HelperInput, output *helpers.HelperOutput,
	zoomAuth interface{}, contact interface{}, email, firstName, lastName string) (*helpers.HelperOutput, error) {

	webinarID := input.Config["webinar_id"].(string)

	// Build registration payload
	payload := map[string]interface{}{
		"email":      email,
		"first_name": firstName,
	}

	if lastName != "" {
		payload["last_name"] = lastName
	}

	// Add custom questions if configured
	if customQuestions, ok := input.Config["custom_questions"].([]interface{}); ok && len(customQuestions) > 0 {
		var questions []map[string]interface{}
		for _, q := range customQuestions {
			if qMap, ok := q.(map[string]interface{}); ok {
				questions = append(questions, qMap)
			}
		}
		if len(questions) > 0 {
			payload["custom_questions"] = questions
		}
	}

	// Make API request
	apiURL := fmt.Sprintf("https://api.zoom.us/v2/webinars/%s/registrants", webinarID)
	regResp, err := makeZoomAPIRequest(ctx, "POST", apiURL, payload, zoomAuth)
	if err != nil {
		output.Message = fmt.Sprintf("Failed to register for webinar: %v", err)
		return output, err
	}

	registrantID := getStringFromMap(regResp, "registrant_id", "")
	joinURL := getStringFromMap(regResp, "join_url", "")

	output.Logs = append(output.Logs, fmt.Sprintf("Registered %s for webinar %s", email, webinarID))

	// Optionally save join URL to CRM field
	if saveField, ok := input.Config["save_join_url_to"].(string); ok && saveField != "" && joinURL != "" {
		if err := input.Connector.SetContactFieldValue(ctx, input.ContactID, saveField, joinURL); err != nil {
			output.Logs = append(output.Logs, fmt.Sprintf("Warning: failed to save join URL to field '%s': %v", saveField, err))
		} else {
			output.Logs = append(output.Logs, fmt.Sprintf("Saved join URL to field '%s'", saveField))
		}
	}

	output.Success = true
	output.Message = fmt.Sprintf("Registered for webinar %s", webinarID)
	output.ModifiedData = map[string]interface{}{
		"webinar_id":    webinarID,
		"registrant_id": registrantID,
		"join_url":      joinURL,
	}
	output.Actions = []helpers.HelperAction{
		{
			Type:   "webinar_registered",
			Target: webinarID,
			Value:  registrantID,
		},
	}

	return output, nil
}
