package data

import (
	"context"
	"fmt"

	"github.com/myfusionhelper/api/internal/helpers"
)

// NewGetTheLast creates a new GetTheLast helper instance
func NewGetTheLast() helpers.Helper { return &GetTheLast{} }

func init() {
	helpers.Register("get_the_last", func() helpers.Helper { return &GetTheLast{} })
}

// GetTheLast retrieves the last (newest) record of a given type (invoice, order, subscription, etc.)
// for a contact and stores a selected field value into a contact field.
// Ported from legacy PHP get_the_last helper.
type GetTheLast struct{}

func (h *GetTheLast) GetName() string     { return "Get The Last" }
func (h *GetTheLast) GetType() string     { return "get_the_last" }
func (h *GetTheLast) GetCategory() string { return "data" }
func (h *GetTheLast) GetDescription() string {
	return "Retrieve the last (newest) invoice, order, subscription, or opportunity record and store a field value"
}
func (h *GetTheLast) RequiresCRM() bool       { return true }
func (h *GetTheLast) SupportedCRMs() []string { return nil }

func (h *GetTheLast) GetConfigSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"type": map[string]interface{}{
				"type":        "string",
				"enum":        []string{"invoice", "job", "subscription", "lead", "creditcard", "payment"},
				"description": "The record type to query for the last entry",
			},
			"from_field": map[string]interface{}{
				"type":        "string",
				"description": "The field on the record to read the value from",
			},
			"to_field": map[string]interface{}{
				"type":        "string",
				"description": "The contact field to store the retrieved value",
			},
		},
		"required": []string{"type", "from_field", "to_field"},
	}
}

func (h *GetTheLast) ValidateConfig(config map[string]interface{}) error {
	recordType, ok := config["type"].(string)
	if !ok || recordType == "" {
		return fmt.Errorf("type is required")
	}
	validTypes := map[string]bool{
		"invoice": true, "job": true, "subscription": true,
		"lead": true, "creditcard": true, "payment": true,
	}
	if !validTypes[recordType] {
		return fmt.Errorf("invalid type: %s", recordType)
	}
	if _, ok := config["from_field"].(string); !ok || config["from_field"] == "" {
		return fmt.Errorf("from_field is required")
	}
	if _, ok := config["to_field"].(string); !ok || config["to_field"] == "" {
		return fmt.Errorf("to_field is required")
	}
	return nil
}

func (h *GetTheLast) Execute(ctx context.Context, input helpers.HelperInput) (*helpers.HelperOutput, error) {
	recordType := input.Config["type"].(string)
	fromField := input.Config["from_field"].(string)
	toField := input.Config["to_field"].(string)

	output := &helpers.HelperOutput{
		Logs: make([]string, 0),
	}

	// Use the connector to query for the last record (descending order = newest first)
	queryKey := fmt.Sprintf("_related.%s.last.%s", recordType, fromField)

	value, err := input.Connector.GetContactFieldValue(ctx, input.ContactID, queryKey)
	if err != nil {
		output.Logs = append(output.Logs, fmt.Sprintf("Related record query not supported, attempting field lookup: %v", err))

		contact, cErr := input.Connector.GetContact(ctx, input.ContactID)
		if cErr != nil {
			output.Message = fmt.Sprintf("Failed to get contact: %v", cErr)
			return output, cErr
		}

		if contact.CustomFields != nil {
			if v, ok := contact.CustomFields[fromField]; ok {
				value = v
			}
		}
	}

	if value == nil || fmt.Sprintf("%v", value) == "" {
		output.Success = true
		output.Message = fmt.Sprintf("No %s records found for contact or field '%s' is empty", recordType, fromField)
		output.Logs = append(output.Logs, output.Message)
		return output, nil
	}

	valueStr := fmt.Sprintf("%v", value)

	// Set the target field on the contact
	err = input.Connector.SetContactFieldValue(ctx, input.ContactID, toField, valueStr)
	if err != nil {
		output.Message = fmt.Sprintf("Failed to set field '%s': %v", toField, err)
		return output, err
	}

	output.Success = true
	output.Message = fmt.Sprintf("Retrieved last %s field '%s' and stored in '%s'", recordType, fromField, toField)
	output.Actions = []helpers.HelperAction{
		{
			Type:   "field_updated",
			Target: toField,
			Value:  valueStr,
		},
	}
	output.ModifiedData = map[string]interface{}{
		toField: valueStr,
	}
	output.Logs = append(output.Logs, fmt.Sprintf("Last %s.%s = '%s' stored in '%s' for contact %s", recordType, fromField, valueStr, toField, input.ContactID))

	return output, nil
}
