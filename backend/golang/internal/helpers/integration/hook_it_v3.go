package integration

import (
	"context"
	"fmt"

	"github.com/myfusionhelper/api/internal/helpers"
)

// NewHookItV3 creates a new HookItV3 helper instance
func NewHookItV3() helpers.Helper { return &HookItV3{} }

func init() {
	helpers.Register("hook_it_v3", func() helpers.Helper { return &HookItV3{} })
}

// HookItV3 handles webhook events with field-based data extraction.
// Saves webhook payload fields to CRM contact custom fields.
// Example: {"total_amount": "OrderTotal", "order_id": "LastOrderID"}
// extracts total_amount from webhook and saves to OrderTotal custom field
type HookItV3 struct{}

func (h *HookItV3) GetName() string     { return "Hook It V3 (Field Mapper)" }
func (h *HookItV3) GetType() string     { return "hook_it_v3" }
func (h *HookItV3) GetCategory() string { return "integration" }
func (h *HookItV3) GetDescription() string {
	return "Extract webhook payload fields and save to CRM contact custom fields"
}
func (h *HookItV3) RequiresCRM() bool       { return true }
func (h *HookItV3) SupportedCRMs() []string { return nil }

func (h *HookItV3) GetConfigSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"field_mappings": map[string]interface{}{
				"type": "object",
				"description": "Map webhook payload field names to CRM custom field names (e.g., {\"total_amount\": \"OrderTotal\", \"order_id\": \"LastOrderID\"})",
				"additionalProperties": map[string]interface{}{
					"type": "string",
				},
			},
			"nested_separator": map[string]interface{}{
				"type":        "string",
				"description": "Separator for nested field access (e.g., '.' for 'order.id', defaults to '.')",
			},
			"skip_null_values": map[string]interface{}{
				"type":        "boolean",
				"description": "Skip setting fields when webhook value is null or empty (defaults to true)",
			},
		},
		"required": []string{"field_mappings"},
	}
}

func (h *HookItV3) ValidateConfig(config map[string]interface{}) error {
	if config["field_mappings"] == nil {
		return fmt.Errorf("field_mappings is required")
	}

	fieldMappings, ok := config["field_mappings"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("field_mappings must be an object")
	}

	if len(fieldMappings) == 0 {
		return fmt.Errorf("field_mappings must contain at least one field mapping")
	}

	// Validate that all values are strings
	for webhookField, crmField := range fieldMappings {
		if _, ok := crmField.(string); !ok {
			return fmt.Errorf("field_mappings[%s] must be a string CRM field name", webhookField)
		}
	}

	return nil
}

// getNestedValue retrieves a value from a nested map using dot notation
func (h *HookItV3) getNestedValue(data map[string]interface{}, path string, separator string) (interface{}, bool) {
	// Simple implementation for nested access
	// For "order.id" with separator ".", split into ["order", "id"]
	parts := []string{path}
	if separator != "" {
		// Split path by separator
		var currentParts []string
		currentPart := ""
		for _, char := range path {
			if string(char) == separator {
				if currentPart != "" {
					currentParts = append(currentParts, currentPart)
					currentPart = ""
				}
			} else {
				currentPart += string(char)
			}
		}
		if currentPart != "" {
			currentParts = append(currentParts, currentPart)
		}
		parts = currentParts
	}

	// Navigate nested structure
	var current interface{} = data
	for i, part := range parts {
		if currentMap, ok := current.(map[string]interface{}); ok {
			val, exists := currentMap[part]
			if !exists {
				return nil, false
			}
			current = val
			// If this is the last part, return the value
			if i == len(parts)-1 {
				return current, true
			}
		} else {
			return nil, false
		}
	}

	return current, true
}

func (h *HookItV3) Execute(ctx context.Context, input helpers.HelperInput) (*helpers.HelperOutput, error) {
	output := &helpers.HelperOutput{
		Actions: make([]helpers.HelperAction, 0),
		Logs:    make([]string, 0),
	}

	// Extract field_mappings from config
	fieldMappings, ok := input.Config["field_mappings"].(map[string]interface{})
	if !ok {
		output.Message = "Invalid field_mappings configuration"
		output.Logs = append(output.Logs, output.Message)
		return output, fmt.Errorf("field_mappings must be an object")
	}

	// Get nested separator (defaults to ".")
	separator := "."
	if sep, ok := input.Config["nested_separator"].(string); ok && sep != "" {
		separator = sep
	}

	// Get skip_null_values flag (defaults to true)
	skipNullValues := true
	if skip, ok := input.Config["skip_null_values"].(bool); ok {
		skipNullValues = skip
	}

	// Process each field mapping
	fieldsUpdated := 0
	for webhookField, crmFieldInterface := range fieldMappings {
		crmField, ok := crmFieldInterface.(string)
		if !ok {
			output.Logs = append(output.Logs, fmt.Sprintf("Skipping invalid mapping for %s", webhookField))
			continue
		}

		// Extract value from webhook data (stored in config)
		var value interface{}
		var found bool
		if webhookData, ok := input.Config["webhook_data"].(map[string]interface{}); ok {
			value, found = h.getNestedValue(webhookData, webhookField, separator)
		} else {
			// Fallback: try to get directly from config
			value, found = h.getNestedValue(input.Config, webhookField, separator)
		}

		if !found {
			output.Logs = append(output.Logs, fmt.Sprintf("Webhook field '%s' not found in payload", webhookField))
			continue
		}

		// Skip null/empty values if configured
		if skipNullValues && (value == nil || value == "") {
			output.Logs = append(output.Logs, fmt.Sprintf("Skipping null/empty value for webhook field '%s'", webhookField))
			continue
		}

		// Set value in CRM custom field
		err := input.Connector.SetContactFieldValue(ctx, input.ContactID, crmField, value)
		if err != nil {
			output.Logs = append(output.Logs, fmt.Sprintf("Failed to set field '%s' from webhook field '%s': %v", crmField, webhookField, err))
			continue
		}

		output.Actions = append(output.Actions, helpers.HelperAction{
			Type:   "field_updated",
			Target: input.ContactID,
			Value:  fmt.Sprintf("%s=%v", crmField, value),
		})
		output.Logs = append(output.Logs, fmt.Sprintf("Set CRM field '%s' to '%v' from webhook field '%s'", crmField, value, webhookField))
		fieldsUpdated++
	}

	output.Success = fieldsUpdated > 0 || len(fieldMappings) == 0
	output.Message = fmt.Sprintf("Field mapping complete: updated %d of %d fields", fieldsUpdated, len(fieldMappings))
	output.ModifiedData = map[string]interface{}{
		"fields_updated":   fieldsUpdated,
		"fields_attempted": len(fieldMappings),
	}

	return output, nil
}
