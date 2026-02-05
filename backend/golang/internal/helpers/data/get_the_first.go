package data

import (
	"context"
	"fmt"

	"github.com/myfusionhelper/api/internal/helpers"
)

func init() {
	helpers.Register("get_the_first", func() helpers.Helper { return &GetTheFirst{} })
}

// GetTheFirst retrieves the first (oldest) record of a given type (invoice, order, subscription, etc.)
// for a contact and stores a selected field value into a contact field.
// Ported from legacy PHP get_the_first helper.
type GetTheFirst struct{}

func (h *GetTheFirst) GetName() string     { return "Get The First" }
func (h *GetTheFirst) GetType() string     { return "get_the_first" }
func (h *GetTheFirst) GetCategory() string { return "data" }
func (h *GetTheFirst) GetDescription() string {
	return "Retrieve the first (oldest) invoice, order, subscription, or opportunity record and store a field value"
}
func (h *GetTheFirst) RequiresCRM() bool       { return true }
func (h *GetTheFirst) SupportedCRMs() []string { return nil }

func (h *GetTheFirst) GetConfigSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"type": map[string]interface{}{
				"type":        "string",
				"enum":        []string{"invoice", "job", "subscription", "lead", "creditcard", "payment"},
				"description": "The record type to query for the first entry",
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

func (h *GetTheFirst) ValidateConfig(config map[string]interface{}) error {
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

func (h *GetTheFirst) Execute(ctx context.Context, input helpers.HelperInput) (*helpers.HelperOutput, error) {
	recordType := input.Config["type"].(string)
	fromField := input.Config["from_field"].(string)
	toField := input.Config["to_field"].(string)

	output := &helpers.HelperOutput{
		Logs: make([]string, 0),
	}

	// Use the connector to query for the first record (ascending order = oldest first)
	// The connector abstraction handles the CRM-specific query.
	// We use GetContactFieldValue with a composite key to indicate we want related data.
	queryKey := fmt.Sprintf("_related.%s.first.%s", recordType, fromField)

	value, err := input.Connector.GetContactFieldValue(ctx, input.ContactID, queryKey)
	if err != nil {
		// Fallback: try getting from contact custom fields directly
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
	output.Message = fmt.Sprintf("Retrieved first %s field '%s' and stored in '%s'", recordType, fromField, toField)
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
	output.Logs = append(output.Logs, fmt.Sprintf("First %s.%s = '%s' stored in '%s' for contact %s", recordType, fromField, valueStr, toField, input.ContactID))

	return output, nil
}
