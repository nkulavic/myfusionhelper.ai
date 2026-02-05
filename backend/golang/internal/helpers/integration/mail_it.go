package integration

import (
	"context"
	"fmt"
	"strings"

	"github.com/myfusionhelper/api/internal/helpers"
)

func init() {
	helpers.Register("mail_it", func() helpers.Helper { return &MailIt{} })
}

// MailIt sends an email via SMTP or API with contact data template interpolation.
// Supports {{field}} merge syntax for dynamic subject and body content.
// The actual email delivery is handled by the downstream execution layer.
type MailIt struct{}

func (h *MailIt) GetName() string     { return "Mail It" }
func (h *MailIt) GetType() string     { return "mail_it" }
func (h *MailIt) GetCategory() string { return "integration" }
func (h *MailIt) GetDescription() string {
	return "Send an email via SMTP or API with contact data merge fields"
}
func (h *MailIt) RequiresCRM() bool       { return true }
func (h *MailIt) SupportedCRMs() []string { return nil }

func (h *MailIt) GetConfigSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"to_field": map[string]interface{}{
				"type":        "string",
				"description": "The contact field containing the recipient email address",
				"default":     "Email",
			},
			"subject_template": map[string]interface{}{
				"type":        "string",
				"description": "Email subject line. Supports {{field_name}} merge fields",
			},
			"body_template": map[string]interface{}{
				"type":        "string",
				"description": "Email body content. Supports {{field_name}} merge fields",
			},
			"from_name": map[string]interface{}{
				"type":        "string",
				"description": "Sender display name",
			},
			"from_email": map[string]interface{}{
				"type":        "string",
				"description": "Sender email address",
			},
			"reply_to": map[string]interface{}{
				"type":        "string",
				"description": "Optional reply-to email address",
			},
			"content_type": map[string]interface{}{
				"type":        "string",
				"enum":        []string{"text/plain", "text/html"},
				"description": "Email content type",
				"default":     "text/html",
			},
		},
		"required": []string{"subject_template", "body_template", "from_name", "from_email"},
	}
}

func (h *MailIt) ValidateConfig(config map[string]interface{}) error {
	if _, ok := config["subject_template"].(string); !ok || config["subject_template"] == "" {
		return fmt.Errorf("subject_template is required")
	}
	if _, ok := config["body_template"].(string); !ok || config["body_template"] == "" {
		return fmt.Errorf("body_template is required")
	}
	if _, ok := config["from_name"].(string); !ok || config["from_name"] == "" {
		return fmt.Errorf("from_name is required")
	}
	if _, ok := config["from_email"].(string); !ok || config["from_email"] == "" {
		return fmt.Errorf("from_email is required")
	}
	return nil
}

func (h *MailIt) Execute(ctx context.Context, input helpers.HelperInput) (*helpers.HelperOutput, error) {
	subjectTemplate := input.Config["subject_template"].(string)
	bodyTemplate := input.Config["body_template"].(string)
	fromName := input.Config["from_name"].(string)
	fromEmail := input.Config["from_email"].(string)

	toField := "Email"
	if tf, ok := input.Config["to_field"].(string); ok && tf != "" {
		toField = tf
	}

	replyTo := ""
	if rt, ok := input.Config["reply_to"].(string); ok {
		replyTo = rt
	}

	contentType := "text/html"
	if ct, ok := input.Config["content_type"].(string); ok && ct != "" {
		contentType = ct
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

	// Resolve the recipient email from the contact field
	toEmail := ""
	if val, exists := fieldData[toField]; exists && val != "" {
		toEmail = val
	}
	if toEmail == "" {
		output.Message = fmt.Sprintf("Recipient email field '%s' is empty for contact %s", toField, input.ContactID)
		return output, fmt.Errorf("recipient email field '%s' is empty", toField)
	}

	// Replace {{field}} syntax in subject and body
	subject := subjectTemplate
	body := bodyTemplate
	for key, value := range fieldData {
		subject = strings.ReplaceAll(subject, "{{"+key+"}}", value)
		body = strings.ReplaceAll(body, "{{"+key+"}}", value)
	}

	// Build the email payload
	emailPayload := map[string]interface{}{
		"to":           toEmail,
		"subject":      subject,
		"body":         body,
		"from_name":    fromName,
		"from_email":   fromEmail,
		"content_type": contentType,
		"contact_id":   input.ContactID,
		"account_id":   input.AccountID,
		"helper_id":    input.HelperID,
	}
	if replyTo != "" {
		emailPayload["reply_to"] = replyTo
	}

	output.Success = true
	output.Message = fmt.Sprintf("Email prepared for %s: %s", toEmail, subject)
	output.Actions = []helpers.HelperAction{
		{
			Type:   "email_queued",
			Target: toEmail,
			Value:  emailPayload,
		},
	}
	output.ModifiedData = emailPayload
	output.Logs = append(output.Logs, fmt.Sprintf("Email queued for contact %s (%s): subject '%s'", input.ContactID, toEmail, subject))

	return output, nil
}
