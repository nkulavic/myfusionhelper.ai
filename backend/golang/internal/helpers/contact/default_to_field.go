package contact

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/myfusionhelper/api/internal/helpers"
)

// NewDefaultToField creates a new DefaultToField helper instance
func NewDefaultToField() helpers.Helper { return &DefaultToField{} }

func init() {
	helpers.Register("default_to_field", func() helpers.Helper { return &DefaultToField{} })
}

// DefaultToField sets a default value (static text, date macro, or merge field) into a contact field.
// Ported from legacy PHP default_to_field helper.
type DefaultToField struct{}

func (h *DefaultToField) GetName() string     { return "Default To Field" }
func (h *DefaultToField) GetType() string     { return "default_to_field" }
func (h *DefaultToField) GetCategory() string { return "contact" }
func (h *DefaultToField) GetDescription() string {
	return "Set a default or computed value into a contact field with support for date macros and merge fields"
}
func (h *DefaultToField) RequiresCRM() bool       { return true }
func (h *DefaultToField) SupportedCRMs() []string { return nil }

func (h *DefaultToField) GetConfigSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"default": map[string]interface{}{
				"type":        "string",
				"description": "The default value to set. Supports @field_name merge fields and date macros: 'now', 'today', @date_now, @date_today",
			},
			"to_field": map[string]interface{}{
				"type":        "string",
				"description": "The target field key to write the value into",
			},
		},
		"required": []string{"default", "to_field"},
	}
}

func (h *DefaultToField) ValidateConfig(config map[string]interface{}) error {
	if _, ok := config["default"].(string); !ok || config["default"] == "" {
		return fmt.Errorf("default value is required")
	}
	if _, ok := config["to_field"].(string); !ok || config["to_field"] == "" {
		return fmt.Errorf("to_field is required")
	}
	return nil
}

func (h *DefaultToField) Execute(ctx context.Context, input helpers.HelperInput) (*helpers.HelperOutput, error) {
	fromField := input.Config["default"].(string)
	toField := input.Config["to_field"].(string)

	output := &helpers.HelperOutput{
		Logs: make([]string, 0),
	}

	now := time.Now()
	nowStr := now.Format("2006-01-02 15:04:05")
	todayStr := now.Format("2006-01-02")

	// Handle date macros
	fromField = strings.ReplaceAll(fromField, "@date_now", nowStr)
	fromField = strings.ReplaceAll(fromField, "@date_today", todayStr)

	// Handle special keywords "now" and "today"
	cleaned := strings.ToLower(strings.ReplaceAll(fromField, "\"", ""))
	if cleaned == "now" {
		fromField = nowStr
	} else if cleaned == "today" {
		fromField = todayStr
	}

	// Handle merge fields (@FieldName) by fetching contact data
	if strings.Contains(fromField, "@") {
		contact, err := input.Connector.GetContact(ctx, input.ContactID)
		if err != nil {
			output.Logs = append(output.Logs, fmt.Sprintf("Warning: could not fetch contact for merge fields: %v", err))
		} else {
			// Replace standard fields
			fieldMap := map[string]string{
				"FirstName": contact.FirstName,
				"LastName":  contact.LastName,
				"Email":     contact.Email,
				"Phone1":    contact.Phone,
				"Company":   contact.Company,
				"JobTitle":  contact.JobTitle,
				"Id":        contact.ID,
			}
			for key, value := range fieldMap {
				fromField = strings.ReplaceAll(fromField, "@"+key, value)
			}

			// Replace custom fields
			if contact.CustomFields != nil {
				for key, value := range contact.CustomFields {
					fromField = strings.ReplaceAll(fromField, "@"+key, fmt.Sprintf("%v", value))
				}
			}
		}
	}

	if fromField == "" {
		output.Success = true
		output.Message = "Default value resolved to empty, nothing to set"
		output.Logs = append(output.Logs, output.Message)
		return output, nil
	}

	// Set the target field
	err := input.Connector.SetContactFieldValue(ctx, input.ContactID, toField, fromField)
	if err != nil {
		output.Message = fmt.Sprintf("Failed to set field '%s': %v", toField, err)
		return output, err
	}

	output.Success = true
	output.Message = fmt.Sprintf("Set field '%s' to default value", toField)
	output.Actions = []helpers.HelperAction{
		{
			Type:   "field_updated",
			Target: toField,
			Value:  fromField,
		},
	}
	output.ModifiedData = map[string]interface{}{
		toField: fromField,
	}
	output.Logs = append(output.Logs, fmt.Sprintf("Set field '%s' to '%s' on contact %s", toField, fromField, input.ContactID))

	return output, nil
}
