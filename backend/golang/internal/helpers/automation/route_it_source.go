package automation

import (
	"context"
	"fmt"
	"strings"

	"github.com/myfusionhelper/api/internal/helpers"
)

func init() {
	helpers.Register("route_it_source", func() helpers.Helper { return &RouteItSource{} })
}

// RouteItSource routes contacts based on traffic source (UTM parameters, referrer, lead source field)
type RouteItSource struct{}

func (h *RouteItSource) GetName() string     { return "Route It - Source" }
func (h *RouteItSource) GetType() string     { return "route_it_source" }
func (h *RouteItSource) GetCategory() string { return "automation" }
func (h *RouteItSource) GetDescription() string {
	return "Route contacts to different URLs based on traffic source (UTM parameters, referrer, or lead source field)"
}
func (h *RouteItSource) RequiresCRM() bool       { return true }
func (h *RouteItSource) SupportedCRMs() []string { return nil }

func (h *RouteItSource) GetConfigSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"source_field": map[string]interface{}{
				"type":        "string",
				"description": "Contact field containing the traffic source (e.g., 'LeadSource', 'utm_source', 'utm_campaign', 'Referrer')",
				"default":     "LeadSource",
			},
			"source_routes": map[string]interface{}{
				"type":        "object",
				"description": "Map of source values to redirect URLs (e.g., {\"google\": \"https://example.com/google\", \"facebook\": \"https://example.com/fb\"})",
			},
			"match_mode": map[string]interface{}{
				"type":        "string",
				"enum":        []string{"exact", "contains", "starts_with"},
				"description": "How to match source values: 'exact', 'contains', or 'starts_with'",
				"default":     "exact",
			},
			"case_sensitive": map[string]interface{}{
				"type":        "boolean",
				"description": "Whether source matching is case-sensitive",
				"default":     false,
			},
			"fallback_url": map[string]interface{}{
				"type":        "string",
				"description": "Default URL if no source matches",
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
		"required": []string{"source_routes"},
	}
}

func (h *RouteItSource) ValidateConfig(config map[string]interface{}) error {
	sourceRoutes, ok := config["source_routes"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("source_routes is required and must be an object")
	}
	if len(sourceRoutes) == 0 {
		return fmt.Errorf("source_routes must contain at least one route")
	}

	matchMode, _ := config["match_mode"].(string)
	if matchMode != "" && matchMode != "exact" && matchMode != "contains" && matchMode != "starts_with" {
		return fmt.Errorf("match_mode must be 'exact', 'contains', or 'starts_with'")
	}

	return nil
}

func (h *RouteItSource) Execute(ctx context.Context, input helpers.HelperInput) (*helpers.HelperOutput, error) {
	sourceField := "LeadSource"
	if sf, ok := input.Config["source_field"].(string); ok && sf != "" {
		sourceField = sf
	}

	matchMode := "exact"
	if mm, ok := input.Config["match_mode"].(string); ok && mm != "" {
		matchMode = mm
	}

	caseSensitive := false
	if cs, ok := input.Config["case_sensitive"].(bool); ok {
		caseSensitive = cs
	}

	fallbackURL, _ := input.Config["fallback_url"].(string)
	saveToField, _ := input.Config["save_to_field"].(string)
	applyTag, _ := input.Config["apply_tag"].(string)

	// Parse source routes
	sourceRoutes := make(map[string]string)
	if sr, ok := input.Config["source_routes"].(map[string]interface{}); ok {
		for source, url := range sr {
			if urlStr, ok := url.(string); ok {
				sourceRoutes[source] = urlStr
			}
		}
	}

	output := &helpers.HelperOutput{
		Actions: make([]helpers.HelperAction, 0),
		Logs:    make([]string, 0),
	}

	// Get the source value from the contact
	sourceValue, err := input.Connector.GetContactFieldValue(ctx, input.ContactID, sourceField)
	if err != nil {
		output.Message = fmt.Sprintf("Failed to get source field '%s': %v", sourceField, err)
		return output, err
	}

	sourceStr := strings.TrimSpace(fmt.Sprintf("%v", sourceValue))
	if sourceValue == nil || sourceStr == "" || sourceStr == "<nil>" {
		output.Logs = append(output.Logs, fmt.Sprintf("Source field '%s' is empty", sourceField))
		sourceStr = ""
	} else {
		output.Logs = append(output.Logs, fmt.Sprintf("Traffic source: %s = %s", sourceField, sourceStr))
	}

	// Normalize for case-insensitive matching
	compareSource := sourceStr
	if !caseSensitive {
		compareSource = strings.ToLower(compareSource)
	}

	// Find matching route
	var selectedURL string
	var routingReason string
	var matchedSourceKey string

	for sourceKey, redirectURL := range sourceRoutes {
		compareKey := sourceKey
		if !caseSensitive {
			compareKey = strings.ToLower(compareKey)
		}

		matched := false
		switch matchMode {
		case "exact":
			matched = (compareSource == compareKey)
		case "contains":
			matched = strings.Contains(compareSource, compareKey)
		case "starts_with":
			matched = strings.HasPrefix(compareSource, compareKey)
		}

		if matched {
			selectedURL = redirectURL
			matchedSourceKey = sourceKey
			routingReason = fmt.Sprintf("source_match:%s (mode:%s)", sourceKey, matchMode)
			output.Logs = append(output.Logs, fmt.Sprintf("Source '%s' matched '%s' -> %s", sourceStr, sourceKey, redirectURL))
			break
		}
	}

	// Fall back if no source matched
	if selectedURL == "" {
		if fallbackURL != "" {
			selectedURL = fallbackURL
			routingReason = "fallback"
			output.Logs = append(output.Logs, "No source matched, using fallback URL")
		} else {
			output.Message = fmt.Sprintf("No source matched for '%s' and no fallback URL configured", sourceStr)
			return output, fmt.Errorf("no route found")
		}
	}

	// Optional: Save URL to CRM field
	if saveToField != "" {
		err := input.Connector.SetContactFieldValue(ctx, input.ContactID, saveToField, selectedURL)
		if err != nil {
			output.Logs = append(output.Logs, fmt.Sprintf("Warning: Failed to save URL to field '%s': %v", saveToField, err))
		} else {
			output.Actions = append(output.Actions, helpers.HelperAction{
				Type:   "field_updated",
				Target: saveToField,
				Value:  selectedURL,
			})
			output.Logs = append(output.Logs, fmt.Sprintf("Saved URL to field '%s'", saveToField))
		}
	}

	// Optional: Apply tag
	if applyTag != "" {
		err := input.Connector.ApplyTag(ctx, input.ContactID, applyTag)
		if err != nil {
			output.Logs = append(output.Logs, fmt.Sprintf("Warning: Failed to apply tag '%s': %v", applyTag, err))
		} else {
			output.Actions = append(output.Actions, helpers.HelperAction{
				Type:   "tag_applied",
				Target: input.ContactID,
				Value:  applyTag,
			})
			output.Logs = append(output.Logs, fmt.Sprintf("Applied tag '%s'", applyTag))
		}
	}

	output.Success = true
	output.Message = fmt.Sprintf("Routed to %s based on traffic source: %s", selectedURL, sourceStr)
	output.ModifiedData = map[string]interface{}{
		"redirect_url":    selectedURL,
		"routing_reason":  routingReason,
		"source":          sourceStr,
		"matched_source":  matchedSourceKey,
		"match_mode":      matchMode,
		"case_sensitive":  caseSensitive,
	}
	output.Actions = append(output.Actions, helpers.HelperAction{
		Type:   "contact_routed",
		Target: input.ContactID,
		Value:  selectedURL,
	})

	return output, nil
}
