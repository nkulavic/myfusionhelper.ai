package contact

import (
	"context"
	"fmt"
	"strings"

	"github.com/myfusionhelper/api/internal/helpers"
)

func init() {
	helpers.Register("note_it", func() helpers.Helper { return &NoteIt{} })
}

// NoteIt adds a note to a contact with template interpolation
type NoteIt struct{}

func (h *NoteIt) GetName() string        { return "Note It" }
func (h *NoteIt) GetType() string        { return "note_it" }
func (h *NoteIt) GetCategory() string    { return "contact" }
func (h *NoteIt) GetDescription() string { return "Add a note to a contact with template interpolation" }
func (h *NoteIt) RequiresCRM() bool      { return true }
func (h *NoteIt) SupportedCRMs() []string { return nil }

func (h *NoteIt) GetConfigSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"subject": map[string]interface{}{
				"type":        "string",
				"description": "Note subject (supports {{field_name}} placeholders)",
			},
			"body": map[string]interface{}{
				"type":        "string",
				"description": "Note body (supports {{field_name}} placeholders)",
			},
			"note_type": map[string]interface{}{
				"type":        "string",
				"description": "Type of note",
				"default":     "general",
			},
		},
		"required": []string{"subject", "body"},
	}
}

func (h *NoteIt) ValidateConfig(config map[string]interface{}) error {
	if _, ok := config["subject"].(string); !ok || config["subject"] == "" {
		return fmt.Errorf("subject is required")
	}
	if _, ok := config["body"].(string); !ok || config["body"] == "" {
		return fmt.Errorf("body is required")
	}
	return nil
}

func (h *NoteIt) Execute(ctx context.Context, input helpers.HelperInput) (*helpers.HelperOutput, error) {
	subject := input.Config["subject"].(string)
	body := input.Config["body"].(string)
	noteType := "general"
	if nt, ok := input.Config["note_type"].(string); ok && nt != "" {
		noteType = nt
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

	// Add custom fields
	if contact.CustomFields != nil {
		for key, val := range contact.CustomFields {
			fieldData[key] = fmt.Sprintf("%v", val)
		}
	}

	// Interpolate templates
	subject = interpolateNoteTemplate(subject, fieldData)
	body = interpolateNoteTemplate(body, fieldData)

	// Build note data
	noteData := map[string]interface{}{
		"subject":    subject,
		"body":       body,
		"note_type":  noteType,
		"contact_id": input.ContactID,
	}

	output.Success = true
	output.Message = fmt.Sprintf("Note prepared: %s", subject)
	output.Actions = []helpers.HelperAction{
		{
			Type:   "notification_queued",
			Target: input.ContactID,
			Value:  noteData,
		},
	}
	output.ModifiedData = noteData
	output.Logs = append(output.Logs, fmt.Sprintf("Note created for contact %s: %s", input.ContactID, subject))

	return output, nil
}

func interpolateNoteTemplate(template string, data map[string]string) string {
	result := template
	for key, value := range data {
		result = strings.ReplaceAll(result, "{{"+key+"}}", value)
	}
	return result
}
