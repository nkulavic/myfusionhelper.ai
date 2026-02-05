package integration

import (
	"context"
	"fmt"
	"strings"

	"github.com/myfusionhelper/api/internal/helpers"
)

func init() {
	helpers.Register("twilio_sms", func() helpers.Helper { return &TwilioSMS{} })
}

// TwilioSMS sends an SMS message via the Twilio API with contact data template interpolation.
// Supports {{field}} merge syntax for dynamic message content.
// The actual SMS delivery is handled by the downstream execution layer.
type TwilioSMS struct{}

func (h *TwilioSMS) GetName() string     { return "Twilio SMS" }
func (h *TwilioSMS) GetType() string     { return "twilio_sms" }
func (h *TwilioSMS) GetCategory() string { return "integration" }
func (h *TwilioSMS) GetDescription() string {
	return "Send an SMS message via Twilio with contact data merge fields"
}
func (h *TwilioSMS) RequiresCRM() bool       { return true }
func (h *TwilioSMS) SupportedCRMs() []string { return nil }

func (h *TwilioSMS) GetConfigSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"account_sid": map[string]interface{}{
				"type":        "string",
				"description": "Twilio account SID",
			},
			"auth_token": map[string]interface{}{
				"type":        "string",
				"description": "Twilio auth token",
			},
			"from_number": map[string]interface{}{
				"type":        "string",
				"description": "Twilio phone number to send from (E.164 format, e.g., +15551234567)",
			},
			"to_field": map[string]interface{}{
				"type":        "string",
				"description": "Contact field containing the recipient phone number",
				"default":     "Phone1",
			},
			"message_template": map[string]interface{}{
				"type":        "string",
				"description": "SMS message body. Supports {{field_name}} merge fields",
			},
		},
		"required": []string{"account_sid", "auth_token", "from_number", "message_template"},
	}
}

func (h *TwilioSMS) ValidateConfig(config map[string]interface{}) error {
	if _, ok := config["account_sid"].(string); !ok || config["account_sid"] == "" {
		return fmt.Errorf("account_sid is required")
	}
	if _, ok := config["auth_token"].(string); !ok || config["auth_token"] == "" {
		return fmt.Errorf("auth_token is required")
	}
	if _, ok := config["from_number"].(string); !ok || config["from_number"] == "" {
		return fmt.Errorf("from_number is required")
	}
	if _, ok := config["message_template"].(string); !ok || config["message_template"] == "" {
		return fmt.Errorf("message_template is required")
	}
	return nil
}

func (h *TwilioSMS) Execute(ctx context.Context, input helpers.HelperInput) (*helpers.HelperOutput, error) {
	accountSID := input.Config["account_sid"].(string)
	authToken := input.Config["auth_token"].(string)
	fromNumber := input.Config["from_number"].(string)
	messageTemplate := input.Config["message_template"].(string)

	toField := "Phone1"
	if tf, ok := input.Config["to_field"].(string); ok && tf != "" {
		toField = tf
	}

	output := &helpers.HelperOutput{
		Logs: make([]string, 0),
	}

	// Get contact data for merge field interpolation
	contact, err := input.Connector.GetContact(ctx, input.ContactID)
	if err != nil {
		output.Message = fmt.Sprintf("Failed to get contact data: %v", err)
		return output, err
	}

	// Build field data map for merge field replacement
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

	// Resolve the recipient phone number from the contact field
	toNumber := ""
	if val, exists := fieldData[toField]; exists && val != "" {
		toNumber = val
	}
	if toNumber == "" {
		output.Message = fmt.Sprintf("Phone field '%s' is empty for contact %s", toField, input.ContactID)
		return output, fmt.Errorf("phone field '%s' is empty", toField)
	}

	// Replace {{field}} syntax in message template
	message := messageTemplate
	for key, value := range fieldData {
		message = strings.ReplaceAll(message, "{{"+key+"}}", value)
	}

	// Build the Twilio API request
	apiURL := fmt.Sprintf("https://api.twilio.com/2010-04-01/Accounts/%s/Messages.json", accountSID)

	twilioPayload := map[string]interface{}{
		"From": fromNumber,
		"To":   toNumber,
		"Body": message,
	}

	output.Success = true
	output.Message = fmt.Sprintf("SMS prepared for %s via Twilio", toNumber)
	output.Actions = []helpers.HelperAction{
		{
			Type:   "webhook_queued",
			Target: apiURL,
			Value: map[string]interface{}{
				"method":      "POST",
				"url":         apiURL,
				"payload":     twilioPayload,
				"auth_type":   "basic",
				"auth_user":   accountSID,
				"auth_pass":   authToken,
				"content_type": "application/x-www-form-urlencoded",
			},
		},
	}
	output.ModifiedData = map[string]interface{}{
		"from_number":  fromNumber,
		"to_number":    toNumber,
		"message":      message,
		"api_url":      apiURL,
		"account_sid":  accountSID,
	}
	output.Logs = append(output.Logs, fmt.Sprintf("Twilio SMS for contact %s (%s): %s", input.ContactID, toNumber, message))

	return output, nil
}
