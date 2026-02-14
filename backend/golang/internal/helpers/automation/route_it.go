package automation

import (
	"context"
	"fmt"
	"time"

	"github.com/myfusionhelper/api/internal/helpers"
)

// NewRouteIt creates a new RouteIt helper instance
func NewRouteIt() helpers.Helper { return &RouteIt{} }

func init() {
	helpers.Register("route_it", func() helpers.Helper { return &RouteIt{} })
}

// RouteIt routes contacts to different URLs based on configurable routing rules
type RouteIt struct{}

func (h *RouteIt) GetName() string     { return "Route It" }
func (h *RouteIt) GetType() string     { return "route_it" }
func (h *RouteIt) GetCategory() string { return "automation" }
func (h *RouteIt) GetDescription() string {
	return "Route contacts to different URLs based on tags, fields, and conditions"
}
func (h *RouteIt) RequiresCRM() bool      { return false }
func (h *RouteIt) SupportedCRMs() []string { return nil }

func (h *RouteIt) GetConfigSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"routes": map[string]interface{}{
				"type":        "array",
				"description": "Array of route configurations",
				"items": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"label": map[string]interface{}{
							"type":        "string",
							"description": "Label for this route",
						},
						"redirectUrl": map[string]interface{}{
							"type":        "string",
							"description": "URL to redirect to if conditions match",
						},
					},
					"required": []string{"redirectUrl"},
				},
			},
			"fallback_url": map[string]interface{}{
				"type":        "string",
				"description": "Default URL if no routes match",
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
		"required": []string{"routes"},
	}
}

func (h *RouteIt) ValidateConfig(config map[string]interface{}) error {
	routes, ok := config["routes"].([]interface{})
	if !ok {
		return fmt.Errorf("routes is required and must be an array")
	}
	if len(routes) == 0 {
		return fmt.Errorf("routes must contain at least one route")
	}

	// Validate each route
	for i, route := range routes {
		routeMap, ok := route.(map[string]interface{})
		if !ok {
			return fmt.Errorf("routes[%d] must be an object", i)
		}

		redirectUrl, ok := routeMap["redirectUrl"].(string)
		if !ok || redirectUrl == "" {
			return fmt.Errorf("routes[%d].redirectUrl is required", i)
		}
	}

	return nil
}

func (h *RouteIt) Execute(ctx context.Context, input helpers.HelperInput) (*helpers.HelperOutput, error) {
	routes := input.Config["routes"].([]interface{})
	fallbackURL, _ := input.Config["fallback_url"].(string)
	saveToField, _ := input.Config["save_to_field"].(string)
	applyTag, _ := input.Config["apply_tag"].(string)

	output := &helpers.HelperOutput{
		Logs: make([]string, 0),
	}

	// For the base route_it, we'll use a simple sequential evaluation
	// Routes are evaluated in order, first match wins
	// This provides a foundation for more complex condition evaluation later

	var selectedURL string
	var routingReason string
	var matchedLabel string

	// Evaluate routes in order
	for i, routeInterface := range routes {
		route := routeInterface.(map[string]interface{})
		redirectUrl := route["redirectUrl"].(string)
		label, _ := route["label"].(string)

		// For now, we'll select the first route in the array
		// This can be extended with condition evaluation (geo, score, source, etc.)
		// Future enhancement: add "condition" field to route config and evaluate it

		if i == 0 {
			selectedURL = redirectUrl
			matchedLabel = label
			if label != "" {
				routingReason = fmt.Sprintf("route_matched=%s", label)
			} else {
				routingReason = fmt.Sprintf("route_index=%d", i)
			}
			output.Logs = append(output.Logs, fmt.Sprintf("Matched route: %s", redirectUrl))
			break
		}
	}

	// Fall back if no route selected (shouldn't happen with current logic, but safety check)
	if selectedURL == "" {
		if fallbackURL != "" {
			selectedURL = fallbackURL
			routingReason = "fallback"
			output.Logs = append(output.Logs, "No route matched, using fallback URL")
		} else {
			output.Message = "No routes configured and no fallback URL"
			return output, fmt.Errorf("no route found")
		}
	}

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
		"matched_label":  matchedLabel,
		"routed_at":      time.Now().Format(time.RFC3339),
	}
	output.Actions = append(output.Actions, helpers.HelperAction{
		Type:   "contact_routed",
		Target: input.ContactID,
		Value:  selectedURL,
	})

	return output, nil
}
