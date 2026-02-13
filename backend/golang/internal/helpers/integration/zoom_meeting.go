package integration

import (
	"context"
	"fmt"

	"github.com/myfusionhelper/api/internal/helpers"
)

// NewZoomMeeting creates a new ZoomMeeting helper instance
func NewZoomMeeting() helpers.Helper { return &ZoomMeeting{} }

func init() {
	helpers.Register("zoom_meeting", func() helpers.Helper { return &ZoomMeeting{} })
}

// ZoomMeeting creates Zoom meetings and registers contacts via Zoom API v2.
// Uses Server-to-Server OAuth via ServiceAuths["zoom"].
type ZoomMeeting struct{}

func (h *ZoomMeeting) GetName() string     { return "Zoom Meeting" }
func (h *ZoomMeeting) GetType() string     { return "zoom_meeting" }
func (h *ZoomMeeting) GetCategory() string { return "integration" }
func (h *ZoomMeeting) GetDescription() string {
	return "Create Zoom meetings and register contacts via the Zoom API"
}
func (h *ZoomMeeting) RequiresCRM() bool       { return true }
func (h *ZoomMeeting) SupportedCRMs() []string { return nil }

func (h *ZoomMeeting) GetConfigSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"action": map[string]interface{}{
				"type":        "string",
				"enum":        []string{"create", "register"},
				"description": "Action: 'create' a meeting or 'register' contact for existing meeting",
				"default":     "create",
			},
			"user_id": map[string]interface{}{
				"type":        "string",
				"description": "Zoom user ID (email or user ID) - required for create action",
			},
			"meeting_id": map[string]interface{}{
				"type":        "string",
				"description": "Existing meeting ID - required for register action",
			},
			"topic": map[string]interface{}{
				"type":        "string",
				"description": "Meeting topic/title - for create action",
			},
			"start_time": map[string]interface{}{
				"type":        "string",
				"description": "Meeting start time (ISO 8601 format: 2024-01-15T10:00:00Z) - for create action",
			},
			"duration": map[string]interface{}{
				"type":        "number",
				"description": "Meeting duration in minutes (default: 60)",
				"default":     60,
			},
			"timezone": map[string]interface{}{
				"type":        "string",
				"description": "Timezone (e.g., America/Los_Angeles, UTC)",
				"default":     "UTC",
			},
			"password": map[string]interface{}{
				"type":        "string",
				"description": "Meeting password (optional)",
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
			"registration_type": map[string]interface{}{
				"type":        "number",
				"description": "Registration type: 1=once, 2=each occurrence, 3=series (default: 1)",
				"default":     1,
			},
			"auto_recording": map[string]interface{}{
				"type":        "string",
				"enum":        []string{"none", "local", "cloud"},
				"description": "Auto recording setting (default: none)",
				"default":     "none",
			},
		},
		"required": []string{"action"},
	}
}

func (h *ZoomMeeting) ValidateConfig(config map[string]interface{}) error {
	action, ok := config["action"].(string)
	if !ok || action == "" {
		return fmt.Errorf("action is required")
	}

	if action == "create" {
		if _, ok := config["user_id"].(string); !ok || config["user_id"] == "" {
			return fmt.Errorf("user_id is required for create action")
		}
	} else if action == "register" {
		if _, ok := config["meeting_id"].(string); !ok || config["meeting_id"] == "" {
			return fmt.Errorf("meeting_id is required for register action")
		}
	} else {
		return fmt.Errorf("action must be 'create' or 'register'")
	}

	return nil
}

func (h *ZoomMeeting) Execute(ctx context.Context, input helpers.HelperInput) (*helpers.HelperOutput, error) {
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
		return h.createMeeting(ctx, input, output, zoomAuth, contact, email, firstName, lastName)
	}

	return h.registerForMeeting(ctx, input, output, zoomAuth, contact, email, firstName, lastName)
}

func (h *ZoomMeeting) createMeeting(ctx context.Context, input helpers.HelperInput, output *helpers.HelperOutput,
	zoomAuth interface{}, contact interface{}, email, firstName, lastName string) (*helpers.HelperOutput, error) {

	userID := input.Config["user_id"].(string)
	topic := getStringConfigValue(input.Config, "topic", "Meeting")
	duration := getIntConfigValue(input.Config, "duration", 60)
	timezone := getStringConfigValue(input.Config, "timezone", "UTC")
	password := getStringConfigValue(input.Config, "password", "")
	autoRecording := getStringConfigValue(input.Config, "auto_recording", "none")
	startTime := getStringConfigValue(input.Config, "start_time", "")

	// Build meeting creation payload
	payload := map[string]interface{}{
		"topic":    topic,
		"type":     2, // Scheduled meeting
		"duration": duration,
		"timezone": timezone,
		"settings": map[string]interface{}{
			"auto_recording": autoRecording,
		},
	}

	if startTime != "" {
		payload["start_time"] = startTime
	}

	if password != "" {
		payload["password"] = password
		if settings, ok := payload["settings"].(map[string]interface{}); ok {
			settings["meeting_authentication"] = true
		}
	}

	// Make API request
	apiURL := fmt.Sprintf("https://api.zoom.us/v2/users/%s/meetings", userID)
	meetingResp, err := makeZoomAPIRequest(ctx, "POST", apiURL, payload, zoomAuth)
	if err != nil {
		output.Message = fmt.Sprintf("Failed to create meeting: %v", err)
		return output, err
	}

	meetingID := getStringFromMap(meetingResp, "id", "")
	joinURL := getStringFromMap(meetingResp, "join_url", "")

	output.Logs = append(output.Logs, fmt.Sprintf("Created Zoom meeting: %s (ID: %s)", topic, meetingID))

	// Optionally save join URL to CRM field
	if saveField, ok := input.Config["save_join_url_to"].(string); ok && saveField != "" && joinURL != "" {
		if err := input.Connector.SetContactFieldValue(ctx, input.ContactID, saveField, joinURL); err != nil {
			output.Logs = append(output.Logs, fmt.Sprintf("Warning: failed to save join URL to field '%s': %v", saveField, err))
		} else {
			output.Logs = append(output.Logs, fmt.Sprintf("Saved join URL to field '%s'", saveField))
		}
	}

	output.Success = true
	output.Message = fmt.Sprintf("Created meeting '%s'", topic)
	output.ModifiedData = map[string]interface{}{
		"meeting_id": meetingID,
		"join_url":   joinURL,
		"topic":      topic,
	}
	output.Actions = []helpers.HelperAction{
		{
			Type:   "meeting_created",
			Target: meetingID,
			Value:  joinURL,
		},
	}

	return output, nil
}

func (h *ZoomMeeting) registerForMeeting(ctx context.Context, input helpers.HelperInput, output *helpers.HelperOutput,
	zoomAuth interface{}, contact interface{}, email, firstName, lastName string) (*helpers.HelperOutput, error) {

	meetingID := input.Config["meeting_id"].(string)

	// Build registration payload
	payload := map[string]interface{}{
		"email":      email,
		"first_name": firstName,
	}

	if lastName != "" {
		payload["last_name"] = lastName
	}

	// Make API request
	apiURL := fmt.Sprintf("https://api.zoom.us/v2/meetings/%s/registrants", meetingID)
	regResp, err := makeZoomAPIRequest(ctx, "POST", apiURL, payload, zoomAuth)
	if err != nil {
		output.Message = fmt.Sprintf("Failed to register for meeting: %v", err)
		return output, err
	}

	registrantID := getStringFromMap(regResp, "registrant_id", "")
	joinURL := getStringFromMap(regResp, "join_url", "")

	output.Logs = append(output.Logs, fmt.Sprintf("Registered %s for meeting %s", email, meetingID))

	// Optionally save join URL to CRM field
	if saveField, ok := input.Config["save_join_url_to"].(string); ok && saveField != "" && joinURL != "" {
		if err := input.Connector.SetContactFieldValue(ctx, input.ContactID, saveField, joinURL); err != nil {
			output.Logs = append(output.Logs, fmt.Sprintf("Warning: failed to save join URL to field '%s': %v", saveField, err))
		} else {
			output.Logs = append(output.Logs, fmt.Sprintf("Saved join URL to field '%s'", saveField))
		}
	}

	output.Success = true
	output.Message = fmt.Sprintf("Registered for meeting %s", meetingID)
	output.ModifiedData = map[string]interface{}{
		"meeting_id":    meetingID,
		"registrant_id": registrantID,
		"join_url":      joinURL,
	}
	output.Actions = []helpers.HelperAction{
		{
			Type:   "meeting_registered",
			Target: meetingID,
			Value:  registrantID,
		},
	}

	return output, nil
}
