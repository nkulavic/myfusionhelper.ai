package automation

import (
	"context"
	"fmt"
	"time"

	"github.com/myfusionhelper/api/internal/helpers"
)

func init() {
	helpers.Register("route_it_by_day", func() helpers.Helper { return &RouteItByDay{} })
}

// RouteItByDay routes contacts to different URLs based on the day of the week
type RouteItByDay struct{}

func (h *RouteItByDay) GetName() string     { return "Route It By Day" }
func (h *RouteItByDay) GetType() string     { return "route_it_by_day" }
func (h *RouteItByDay) GetCategory() string { return "automation" }
func (h *RouteItByDay) GetDescription() string {
	return "Route contacts to different URLs based on the day of the week"
}
func (h *RouteItByDay) RequiresCRM() bool      { return false }
func (h *RouteItByDay) SupportedCRMs() []string { return nil }

func (h *RouteItByDay) GetConfigSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"day_routes": map[string]interface{}{
				"type":        "object",
				"description": "Map of day names (monday, tuesday, etc.) to redirect URLs",
				"additionalProperties": map[string]interface{}{
					"type": "string",
				},
			},
			"fallback_url": map[string]interface{}{
				"type":        "string",
				"description": "Default URL if no day mapping exists",
			},
			"timezone": map[string]interface{}{
				"type":        "string",
				"description": "Timezone for day calculation (e.g., 'America/New_York'). Defaults to UTC.",
				"default":     "UTC",
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
		"required": []string{"day_routes"},
	}
}

func (h *RouteItByDay) ValidateConfig(config map[string]interface{}) error {
	dayRoutes, ok := config["day_routes"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("day_routes is required and must be an object")
	}
	if len(dayRoutes) == 0 {
		return fmt.Errorf("day_routes must contain at least one day mapping")
	}
	return nil
}

func (h *RouteItByDay) Execute(ctx context.Context, input helpers.HelperInput) (*helpers.HelperOutput, error) {
	dayRoutes := input.Config["day_routes"].(map[string]interface{})
	fallbackURL, _ := input.Config["fallback_url"].(string)
	timezone, _ := input.Config["timezone"].(string)
	if timezone == "" {
		timezone = "UTC"
	}
	saveToField, _ := input.Config["save_to_field"].(string)
	applyTag, _ := input.Config["apply_tag"].(string)

	output := &helpers.HelperOutput{
		Logs: make([]string, 0),
	}

	// Load timezone
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		output.Message = fmt.Sprintf("Invalid timezone '%s': %v", timezone, err)
		return output, err
	}

	// Get current day of week in the specified timezone
	now := time.Now().In(loc)
	dayOfWeek := now.Weekday().String() // "Monday", "Tuesday", etc.
	dayKey := dayOfWeek // Try exact match first

	// Try lowercase version
	if _, ok := dayRoutes[dayKey]; !ok {
		dayKey = fmt.Sprintf("%s", dayOfWeek)
		// Convert to lowercase for case-insensitive lookup
		for key := range dayRoutes {
			if fmt.Sprintf("%s", key) == fmt.Sprintf("%s", dayOfWeek) {
				dayKey = key
				break
			}
		}
	}

	// Select URL based on day
	var selectedURL string
	var routingReason string

	if urlInterface, ok := dayRoutes[dayKey]; ok {
		if url, ok := urlInterface.(string); ok && url != "" {
			selectedURL = url
			routingReason = fmt.Sprintf("day_of_week=%s", dayOfWeek)
		}
	}

	// Fall back if no matching day found
	if selectedURL == "" {
		if fallbackURL != "" {
			selectedURL = fallbackURL
			routingReason = "fallback"
		} else {
			output.Message = fmt.Sprintf("No route found for %s and no fallback URL configured", dayOfWeek)
			return output, fmt.Errorf("no route found")
		}
	}

	output.Logs = append(output.Logs, fmt.Sprintf("Current day: %s (timezone: %s)", dayOfWeek, timezone))
	output.Logs = append(output.Logs, fmt.Sprintf("Selected URL: %s (reason: %s)", selectedURL, routingReason))

	// Optional: Save to CRM field
	if saveToField != "" && input.Connector != nil {
		err := input.Connector.SetContactFieldValue(ctx, input.ContactID, saveToField, selectedURL)
		if err != nil {
			output.Logs = append(output.Logs, fmt.Sprintf("Warning: Failed to save URL to field '%s': %v", saveToField, err))
		} else {
			output.Logs = append(output.Logs, fmt.Sprintf("Saved URL to field '%s'", saveToField))
		}
	}

	// Optional: Apply tag
	if applyTag != "" && input.Connector != nil {
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
		"day_of_week":    dayOfWeek,
		"routed_at":      now.Format(time.RFC3339),
	}
	output.Actions = append(output.Actions, helpers.HelperAction{
		Type:   "contact_routed",
		Target: input.ContactID,
		Value:  selectedURL,
	})

	return output, nil
}
