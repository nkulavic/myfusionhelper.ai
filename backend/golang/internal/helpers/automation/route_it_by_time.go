package automation

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/myfusionhelper/api/internal/helpers"
)

// NewRouteItByTime creates a new RouteItByTime helper instance
func NewRouteItByTime() helpers.Helper { return &RouteItByTime{} }

func init() {
	helpers.Register("route_it_by_time", func() helpers.Helper { return &RouteItByTime{} })
}

// RouteItByTime routes contacts to different URLs based on time of day (timezone-aware)
type RouteItByTime struct{}

func (h *RouteItByTime) GetName() string     { return "Route It By Time" }
func (h *RouteItByTime) GetType() string     { return "route_it_by_time" }
func (h *RouteItByTime) GetCategory() string { return "automation" }
func (h *RouteItByTime) GetDescription() string {
	return "Route contacts to different URLs based on time of day with timezone awareness"
}
func (h *RouteItByTime) RequiresCRM() bool      { return false }
func (h *RouteItByTime) SupportedCRMs() []string { return nil }

func (h *RouteItByTime) GetConfigSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"time_routes": map[string]interface{}{
				"type":        "array",
				"description": "Array of time range routes",
				"items": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"start_time": map[string]interface{}{
							"type":        "string",
							"description": "Start time in HH:MM format (24-hour)",
						},
						"end_time": map[string]interface{}{
							"type":        "string",
							"description": "End time in HH:MM format (24-hour)",
						},
						"url": map[string]interface{}{
							"type":        "string",
							"description": "URL to redirect to if current time falls in this range",
						},
						"label": map[string]interface{}{
							"type":        "string",
							"description": "Optional label for this time range",
						},
					},
					"required": []string{"start_time", "end_time", "url"},
				},
			},
			"fallback_url": map[string]interface{}{
				"type":        "string",
				"description": "Default URL if no time range matches",
			},
			"timezone": map[string]interface{}{
				"type":        "string",
				"description": "Timezone for time calculation (e.g., 'America/New_York'). Defaults to UTC.",
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
		"required": []string{"time_routes"},
	}
}

func (h *RouteItByTime) ValidateConfig(config map[string]interface{}) error {
	timeRoutes, ok := config["time_routes"].([]interface{})
	if !ok {
		return fmt.Errorf("time_routes is required and must be an array")
	}
	if len(timeRoutes) == 0 {
		return fmt.Errorf("time_routes must contain at least one time range")
	}

	// Validate each time route
	for i, route := range timeRoutes {
		routeMap, ok := route.(map[string]interface{})
		if !ok {
			return fmt.Errorf("time_routes[%d] must be an object", i)
		}

		startTime, ok := routeMap["start_time"].(string)
		if !ok || startTime == "" {
			return fmt.Errorf("time_routes[%d].start_time is required", i)
		}

		endTime, ok := routeMap["end_time"].(string)
		if !ok || endTime == "" {
			return fmt.Errorf("time_routes[%d].end_time is required", i)
		}

		url, ok := routeMap["url"].(string)
		if !ok || url == "" {
			return fmt.Errorf("time_routes[%d].url is required", i)
		}

		// Validate time format
		if _, _, err := parseTimeString(startTime); err != nil {
			return fmt.Errorf("time_routes[%d].start_time invalid format (use HH:MM): %v", i, err)
		}
		if _, _, err := parseTimeString(endTime); err != nil {
			return fmt.Errorf("time_routes[%d].end_time invalid format (use HH:MM): %v", i, err)
		}
	}

	return nil
}

// parseTimeString parses HH:MM format and returns hour and minute
func parseTimeString(timeStr string) (hour, minute int, err error) {
	parts := strings.Split(timeStr, ":")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("invalid time format, expected HH:MM")
	}

	hour, err = strconv.Atoi(parts[0])
	if err != nil || hour < 0 || hour > 23 {
		return 0, 0, fmt.Errorf("invalid hour: %s", parts[0])
	}

	minute, err = strconv.Atoi(parts[1])
	if err != nil || minute < 0 || minute > 59 {
		return 0, 0, fmt.Errorf("invalid minute: %s", parts[1])
	}

	return hour, minute, nil
}

// timeInRange checks if current time is within the specified range
func timeInRange(currentTime time.Time, startHour, startMin, endHour, endMin int) bool {
	currentMinutes := currentTime.Hour()*60 + currentTime.Minute()
	startMinutes := startHour*60 + startMin
	endMinutes := endHour*60 + endMin

	// Handle overnight ranges (e.g., 22:00 - 06:00)
	if endMinutes < startMinutes {
		return currentMinutes >= startMinutes || currentMinutes <= endMinutes
	}

	return currentMinutes >= startMinutes && currentMinutes <= endMinutes
}

func (h *RouteItByTime) Execute(ctx context.Context, input helpers.HelperInput) (*helpers.HelperOutput, error) {
	timeRoutes := input.Config["time_routes"].([]interface{})
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

	// Get current time in the specified timezone
	now := time.Now().In(loc)
	currentTimeStr := now.Format("15:04")
	output.Logs = append(output.Logs, fmt.Sprintf("Current time: %s (timezone: %s)", currentTimeStr, timezone))

	// Check each time route
	var selectedURL string
	var routingReason string

	for _, routeInterface := range timeRoutes {
		route := routeInterface.(map[string]interface{})
		startTime := route["start_time"].(string)
		endTime := route["end_time"].(string)
		url := route["url"].(string)
		label, _ := route["label"].(string)

		startHour, startMin, _ := parseTimeString(startTime)
		endHour, endMin, _ := parseTimeString(endTime)

		if timeInRange(now, startHour, startMin, endHour, endMin) {
			selectedURL = url
			if label != "" {
				routingReason = fmt.Sprintf("time_range=%s (%s-%s)", label, startTime, endTime)
			} else {
				routingReason = fmt.Sprintf("time_range=%s-%s", startTime, endTime)
			}
			output.Logs = append(output.Logs, fmt.Sprintf("Matched time range: %s - %s", startTime, endTime))
			break
		}
	}

	// Fall back if no time range matched
	if selectedURL == "" {
		if fallbackURL != "" {
			selectedURL = fallbackURL
			routingReason = "fallback"
			output.Logs = append(output.Logs, "No time range matched, using fallback URL")
		} else {
			output.Message = fmt.Sprintf("No time range matched current time %s and no fallback URL configured", currentTimeStr)
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
		"current_time":   currentTimeStr,
		"timezone":       timezone,
		"routed_at":      now.Format(time.RFC3339),
	}
	output.Actions = append(output.Actions, helpers.HelperAction{
		Type:   "contact_routed",
		Target: input.ContactID,
		Value:  selectedURL,
	})

	return output, nil
}
