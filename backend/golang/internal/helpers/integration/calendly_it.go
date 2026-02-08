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
			"webhook_event": map[string]interface{}{
				"type":        "string",
				"description": "Type of Calendly webhook event to process",
				"enum":        []string{"invitee.created", "invitee.canceled", "invitee.rescheduled"},
			},
			"created_tag": map[string]interface{}{
				"type":        "string",
				"description": "Tag to apply when meeting is created (invitee.created event)",
			},
			"canceled_tag": map[string]interface{}{
				"type":        "string",
				"description": "Tag to apply when meeting is canceled (invitee.canceled event)",
			},
			"rescheduled_tag": map[string]interface{}{
				"type":        "string",
				"description": "Tag to apply when meeting is rescheduled (invitee.rescheduled event)",
			},
			"save_meeting_time_to": map[string]interface{}{
				"type":        "string",
				"description": "Optional contact field to save the meeting start time (ISO 8601 format)",
			},
			"save_duration_to": map[string]interface{}{
				"type":        "string",
				"description": "Optional contact field to save the meeting duration (in minutes)",
			},
			"save_event_type_to": map[string]interface{}{
				"type":        "string",
				"description": "Optional contact field to save the Calendly event type name",
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

	// Extract webhook event type if provided
	webhookEvent := ""
	if we, ok := input.Config["webhook_event"].(string); ok {
		webhookEvent = we
	}

	output := &helpers.HelperOutput{
		Logs: make([]string, 0),
	}

	// If this is a webhook event, process it differently
	if webhookEvent != "" {
		return h.processWebhookEvent(ctx, input, webhookEvent, output)
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

// processWebhookEvent handles Calendly webhook events (invitee.created, invitee.canceled, invitee.rescheduled)
func (h *CalendlyIt) processWebhookEvent(ctx context.Context, input helpers.HelperInput, eventType string, output *helpers.HelperOutput) (*helpers.HelperOutput, error) {
	// Extract event-specific tags
	createdTag := ""
	if ct, ok := input.Config["created_tag"].(string); ok {
		createdTag = ct
	}

	canceledTag := ""
	if ct, ok := input.Config["canceled_tag"].(string); ok {
		canceledTag = ct
	}

	rescheduledTag := ""
	if rt, ok := input.Config["rescheduled_tag"].(string); ok {
		rescheduledTag = rt
	}

	// Extract field mapping configs
	saveMeetingTimeTo := ""
	if mt, ok := input.Config["save_meeting_time_to"].(string); ok {
		saveMeetingTimeTo = mt
	}

	saveDurationTo := ""
	if dt, ok := input.Config["save_duration_to"].(string); ok {
		saveDurationTo = dt
	}

	saveEventTypeTo := ""
	if et, ok := input.Config["save_event_type_to"].(string); ok {
		saveEventTypeTo = et
	}

	// Extract webhook payload data from config (webhook data is merged into config during execution)
	meetingTime := ""
	duration := ""
	eventTypeName := ""

	if mt, ok := input.Config["meeting_time"].(string); ok {
		meetingTime = mt
	}

	if dur, ok := input.Config["duration"].(float64); ok {
		duration = fmt.Sprintf("%.0f", dur)
	} else if dur, ok := input.Config["duration"].(int); ok {
		duration = fmt.Sprintf("%d", dur)
	}

	if et, ok := input.Config["event_type_name"].(string); ok {
		eventTypeName = et
	}

	// Determine which tag to apply based on event type
	tagToApply := ""
	switch eventType {
	case "invitee.created":
		tagToApply = createdTag
		output.Logs = append(output.Logs, "Processing invitee.created event")
	case "invitee.canceled":
		tagToApply = canceledTag
		output.Logs = append(output.Logs, "Processing invitee.canceled event")
	case "invitee.rescheduled":
		tagToApply = rescheduledTag
		output.Logs = append(output.Logs, "Processing invitee.rescheduled event")
	default:
		output.Message = fmt.Sprintf("Unknown webhook event type: %s", eventType)
		return output, fmt.Errorf("unknown webhook event type: %s", eventType)
	}

	// Apply tag if configured
	if tagToApply != "" && input.Connector != nil {
		tags, err := input.Connector.GetTags(ctx)
		if err == nil {
			for _, tag := range tags {
				if tag.Name == tagToApply || tag.ID == tagToApply {
					err = input.Connector.ApplyTag(ctx, input.ContactID, tag.ID)
					if err != nil {
						output.Logs = append(output.Logs, fmt.Sprintf("Failed to apply tag '%s': %v", tagToApply, err))
					} else {
						output.Logs = append(output.Logs, fmt.Sprintf("Applied tag '%s' for %s event", tagToApply, eventType))
						output.Actions = append(output.Actions, helpers.HelperAction{
							Type:   "tag_applied",
							Target: input.ContactID,
							Value:  tagToApply,
						})
					}
					break
				}
			}
		}
	}

	// Save field data if configured
	fieldUpdates := make(map[string]interface{})
	if saveMeetingTimeTo != "" && meetingTime != "" {
		fieldUpdates[saveMeetingTimeTo] = meetingTime
		output.Logs = append(output.Logs, fmt.Sprintf("Saving meeting time to field '%s': %s", saveMeetingTimeTo, meetingTime))
	}
	if saveDurationTo != "" && duration != "" {
		fieldUpdates[saveDurationTo] = duration
		output.Logs = append(output.Logs, fmt.Sprintf("Saving duration to field '%s': %s", saveDurationTo, duration))
	}
	if saveEventTypeTo != "" && eventTypeName != "" {
		fieldUpdates[saveEventTypeTo] = eventTypeName
		output.Logs = append(output.Logs, fmt.Sprintf("Saving event type to field '%s': %s", saveEventTypeTo, eventTypeName))
	}

	// Apply field updates
	if len(fieldUpdates) > 0 && input.Connector != nil {
		for field, value := range fieldUpdates {
			err := input.Connector.SetContactFieldValue(ctx, input.ContactID, field, value)
			if err != nil {
				output.Logs = append(output.Logs, fmt.Sprintf("Failed to set field '%s': %v", field, err))
			} else {
				output.Actions = append(output.Actions, helpers.HelperAction{
					Type:   "field_updated",
					Target: field,
					Value:  value,
				})
			}
		}
	}

	output.Success = true
	output.Message = fmt.Sprintf("Processed Calendly %s event for contact %s", eventType, input.ContactID)
	output.ModifiedData = map[string]interface{}{
		"event_type":       eventType,
		"tag_applied":      tagToApply,
		"field_updates":    fieldUpdates,
		"meeting_time":     meetingTime,
		"duration":         duration,
		"event_type_name":  eventTypeName,
	}

	return output, nil
}
