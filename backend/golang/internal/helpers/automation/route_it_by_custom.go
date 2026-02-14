package automation

import (
	"context"
	"fmt"
	"time"

	"github.com/myfusionhelper/api/internal/helpers"
)

// NewRouteItByCustom creates a new RouteItByCustom helper instance
func NewRouteItByCustom() helpers.Helper { return &RouteItByCustom{} }

func init() {
	helpers.Register("route_it_by_custom", func() helpers.Helper { return &RouteItByCustom{} })
}

// RouteItByCustom routes contacts to different URLs based on custom field values
type RouteItByCustom struct{}

func (h *RouteItByCustom) GetName() string     { return "Route It By Custom Field" }
func (h *RouteItByCustom) GetType() string     { return "route_it_by_custom" }
func (h *RouteItByCustom) GetCategory() string { return "automation" }
func (h *RouteItByCustom) GetDescription() string {
	return "Route contacts to different URLs based on custom field values"
}
func (h *RouteItByCustom) RequiresCRM() bool      { return true }
func (h *RouteItByCustom) SupportedCRMs() []string { return nil }

func (h *RouteItByCustom) GetConfigSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"field_name": map[string]interface{}{
				"type":        "string",
				"description": "Custom field name to check",
			},
			"value_routes": map[string]interface{}{
				"type":        "object",
				"description": "Map of field values to redirect URLs",
				"additionalProperties": map[string]interface{}{
					"type": "string",
				},
			},
			"fallback_url": map[string]interface{}{
				"type":        "string",
				"description": "Default URL if field value doesn't match any route",
			},
			"save_to_field": map[string]interface{}{
				"type":        "string",
				"description": "Optional: CRM field to save the selected URL to",
			},
			"apply_tag": map[string]interface{}{
				"type":        "string",
				"description": "Optional: Tag ID to apply when routing occurs",
			},
		},
		"required": []string{"field_name", "value_routes"},
	}
}

func (h *RouteItByCustom) ValidateConfig(config map[string]interface{}) error {
	if _, ok := config["field_name"].(string); !ok || config["field_name"] == "" {
		return fmt.Errorf("field_name is required")
	}
	valueRoutes, ok := config["value_routes"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("value_routes is required and must be an object")
	}
	if len(valueRoutes) == 0 {
		return fmt.Errorf("value_routes must contain at least one value mapping")
	}
	return nil
}

func (h *RouteItByCustom) Execute(ctx context.Context, input helpers.HelperInput) (*helpers.HelperOutput, error) {
	fieldName := input.Config["field_name"].(string)
	valueRoutes := input.Config["value_routes"].(map[string]interface{})
	fallbackURL, _ := input.Config["fallback_url"].(string)
	saveToField, _ := input.Config["save_to_field"].(string)
	applyTag, _ := input.Config["apply_tag"].(string)

	output := &helpers.HelperOutput{
		Logs: make([]string, 0),
	}

	// Get field value from contact
	fieldValue, err := input.Connector.GetContactFieldValue(ctx, input.ContactID, fieldName)
	if err != nil {
		output.Message = fmt.Sprintf("Failed to get field '%s': %v", fieldName, err)
		return output, err
	}

	// Convert field value to string for comparison
	fieldValueStr := fmt.Sprintf("%v", fieldValue)
	output.Logs = append(output.Logs, fmt.Sprintf("Field '%s' value: %s", fieldName, fieldValueStr))

	// Select URL based on field value
	var selectedURL string
	var routingReason string

	if urlInterface, ok := valueRoutes[fieldValueStr]; ok {
		if url, ok := urlInterface.(string); ok && url != "" {
			selectedURL = url
			routingReason = fmt.Sprintf("field_%s=%s", fieldName, fieldValueStr)
		}
	}

	// Fall back if no matching value found
	if selectedURL == "" {
		if fallbackURL != "" {
			selectedURL = fallbackURL
			routingReason = "fallback"
		} else {
			output.Message = fmt.Sprintf("No route found for field '%s' value '%s' and no fallback URL configured", fieldName, fieldValueStr)
			return output, fmt.Errorf("no route found")
		}
	}

	output.Logs = append(output.Logs, fmt.Sprintf("Selected URL: %s (reason: %s)", selectedURL, routingReason))

	// Optional: Save to CRM field
	if saveToField != "" {
		err := input.Connector.SetContactFieldValue(ctx, input.ContactID, saveToField, selectedURL)
		if err != nil {
			output.Logs = append(output.Logs, fmt.Sprintf("Warning: Failed to save URL to field '%s': %v", saveToField, err))
		} else {
			output.Logs = append(output.Logs, fmt.Sprintf("Saved URL to field '%s'", saveToField))
		}
	}

	// Optional: Apply tag
	if applyTag != "" {
		err := input.Connector.ApplyTag(ctx, input.ContactID, applyTag)
		if err != nil {
			output.Logs = append(output.Logs, fmt.Sprintf("Warning: Failed to apply tag '%s': %v", applyTag, err))
		} else {
			output.Logs = append(output.Logs, fmt.Sprintf("Applied tag '%s'", applyTag))
			output.Actions = append(output.Actions, helpers.HelperAction{
				Type:   "tag_applied",
				Target: input.ContactID,
				Value:  applyTag,
			})
		}
	}

	output.Success = true
	output.Message = fmt.Sprintf("Routed contact to %s based on %s", selectedURL, routingReason)
	output.ModifiedData = map[string]interface{}{
		"redirect_url":   selectedURL,
		"routing_reason": routingReason,
		"field_name":     fieldName,
		"field_value":    fieldValueStr,
		"routed_at":      time.Now().Format(time.RFC3339),
	}
	output.Actions = append(output.Actions, helpers.HelperAction{
		Type:   "contact_routed",
		Target: input.ContactID,
		Value:  selectedURL,
	})

	return output, nil
}
