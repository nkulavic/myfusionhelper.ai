package notification

import (
	"context"
	"fmt"
	"strings"

	"github.com/myfusionhelper/api/internal/helpers"
)

// NewNotifyMe creates a new NotifyMe helper instance
func NewNotifyMe() helpers.Helper { return &NotifyMe{} }

func init() {
	helpers.Register("notify_me", func() helpers.Helper { return &NotifyMe{} })
}

// NotifyMe generates a notification with contact data for the user.
// The actual notification delivery (email, Slack, webhook) is handled by the execution layer
// based on the notification actions in the output.
type NotifyMe struct{}

func (h *NotifyMe) GetName() string        { return "Notify Me" }
func (h *NotifyMe) GetType() string        { return "notify_me" }
func (h *NotifyMe) GetCategory() string    { return "notification" }
func (h *NotifyMe) GetDescription() string { return "Send a notification with contact data via configured channels" }
func (h *NotifyMe) RequiresCRM() bool      { return true }
func (h *NotifyMe) SupportedCRMs() []string { return nil }

func (h *NotifyMe) GetConfigSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"channel": map[string]interface{}{
				"type":        "string",
				"enum":        []string{"email", "slack", "webhook"},
				"description": "Notification delivery channel",
			},
			"subject": map[string]interface{}{
				"type":        "string",
				"description": "Notification subject/title (supports {{field_name}} placeholders)",
			},
			"message": map[string]interface{}{
				"type":        "string",
				"description": "Notification message body (supports {{field_name}} placeholders)",
			},
			"recipient": map[string]interface{}{
				"type":        "string",
				"description": "Recipient email, Slack channel, or webhook URL",
			},
			"include_fields": map[string]interface{}{
				"type":        "array",
				"items":       map[string]interface{}{"type": "string"},
				"description": "Contact fields to include in notification data",
			},
		},
		"required": []string{"channel", "message"},
	}
}

func (h *NotifyMe) ValidateConfig(config map[string]interface{}) error {
	channel, ok := config["channel"].(string)
	if !ok || channel == "" {
		return fmt.Errorf("channel is required")
	}

	validChannels := map[string]bool{"email": true, "slack": true, "webhook": true}
	if !validChannels[channel] {
		return fmt.Errorf("invalid channel: %s", channel)
	}

	if _, ok := config["message"].(string); !ok || config["message"] == "" {
		return fmt.Errorf("message is required")
	}

	return nil
}

func (h *NotifyMe) Execute(ctx context.Context, input helpers.HelperInput) (*helpers.HelperOutput, error) {
	channel := input.Config["channel"].(string)
	message := input.Config["message"].(string)
	subject := ""
	if s, ok := input.Config["subject"].(string); ok {
		subject = s
	}
	recipient := ""
	if r, ok := input.Config["recipient"].(string); ok {
		recipient = r
	}

	output := &helpers.HelperOutput{
		Logs: make([]string, 0),
	}

	// Get contact data for template interpolation
	contact, err := input.Connector.GetContact(ctx, input.ContactID)
	if err != nil {
		output.Message = fmt.Sprintf("Failed to get contact data: %v", err)
		return output, err
	}

	// Build field data for templates
	fieldData := map[string]string{
		"contact_id": contact.ID,
		"first_name": contact.FirstName,
		"last_name":  contact.LastName,
		"email":      contact.Email,
		"phone":      contact.Phone,
		"company":    contact.Company,
		"full_name":  strings.TrimSpace(contact.FirstName + " " + contact.LastName),
	}

	// Include additional fields if specified
	if includeFields, ok := input.Config["include_fields"]; ok {
		fields := extractStringSlice(includeFields)
		for _, fieldKey := range fields {
			val, err := input.Connector.GetContactFieldValue(ctx, input.ContactID, fieldKey)
			if err == nil && val != nil {
				fieldData[fieldKey] = fmt.Sprintf("%v", val)
			}
		}
	}

	// Interpolate templates
	message = interpolateTemplate(message, fieldData)
	subject = interpolateTemplate(subject, fieldData)

	// Build notification data
	notificationData := map[string]interface{}{
		"channel":   channel,
		"subject":   subject,
		"message":   message,
		"recipient": recipient,
		"contact":   fieldData,
	}

	output.Success = true
	output.Message = fmt.Sprintf("Notification prepared for %s channel", channel)
	output.Actions = []helpers.HelperAction{
		{
			Type:   "notification_queued",
			Target: channel,
			Value:  notificationData,
		},
	}
	output.ModifiedData = notificationData
	output.Logs = append(output.Logs, fmt.Sprintf("Notification for contact %s via %s: %s", input.ContactID, channel, subject))

	return output, nil
}

func interpolateTemplate(template string, data map[string]string) string {
	result := template
	for key, value := range data {
		result = strings.ReplaceAll(result, "{{"+key+"}}", value)
	}
	return result
}

func extractStringSlice(v interface{}) []string {
	switch val := v.(type) {
	case []string:
		return val
	case []interface{}:
		result := make([]string, 0, len(val))
		for _, item := range val {
			if s, ok := item.(string); ok {
				result = append(result, s)
			} else {
				result = append(result, fmt.Sprintf("%v", item))
			}
		}
		return result
	default:
		return nil
	}
}
