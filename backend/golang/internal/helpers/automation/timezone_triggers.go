package automation

import (
	"context"
	"fmt"
	"strings"

	"github.com/myfusionhelper/api/internal/helpers"
)

func init() {
	helpers.Register("timezone_triggers", func() helpers.Helper { return &TimezoneTriggers{} })
}

// TimezoneTriggers resolves a contact's timezone from their address data and schedules
// a time-zone-aware automation trigger. It geocodes the contact's address, determines
// their timezone, and optionally saves timezone data to contact fields.
// Ported from legacy PHP timezone_triggers helper.
type TimezoneTriggers struct{}

func (h *TimezoneTriggers) GetName() string     { return "Timezone Triggers" }
func (h *TimezoneTriggers) GetType() string     { return "timezone_triggers" }
func (h *TimezoneTriggers) GetCategory() string { return "automation" }
func (h *TimezoneTriggers) GetDescription() string {
	return "Resolve contact timezone from address and schedule time-zone-aware automation triggers"
}
func (h *TimezoneTriggers) RequiresCRM() bool       { return true }
func (h *TimezoneTriggers) SupportedCRMs() []string { return nil }

func (h *TimezoneTriggers) GetConfigSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"day": map[string]interface{}{
				"type":        "string",
				"description": "Day of the week for the trigger (e.g., Monday, Tuesday)",
			},
			"time": map[string]interface{}{
				"type":        "string",
				"description": "Time of day for the trigger (e.g., 9:00 AM, 14:00)",
			},
			"save_time_zone": map[string]interface{}{
				"type":        "string",
				"description": "Optional field to save the resolved timezone ID",
			},
			"save_lat_lng": map[string]interface{}{
				"type":        "string",
				"description": "Optional field to save the resolved lat,lng coordinates",
			},
			"save_time_zone_offset": map[string]interface{}{
				"type":        "string",
				"description": "Optional field to save the timezone UTC offset in hours",
			},
			"trigger_goal": map[string]interface{}{
				"type":        "string",
				"description": "Goal call name to achieve when the timezone trigger fires",
			},
			"failed_goal": map[string]interface{}{
				"type":        "string",
				"description": "Goal call name to achieve when no address is found",
			},
		},
		"required": []string{"day", "time"},
	}
}

func (h *TimezoneTriggers) ValidateConfig(config map[string]interface{}) error {
	if _, ok := config["day"].(string); !ok || config["day"] == "" {
		return fmt.Errorf("day is required")
	}
	if _, ok := config["time"].(string); !ok || config["time"] == "" {
		return fmt.Errorf("time is required")
	}
	return nil
}

func (h *TimezoneTriggers) Execute(ctx context.Context, input helpers.HelperInput) (*helpers.HelperOutput, error) {
	day := input.Config["day"].(string)
	triggerTime := input.Config["time"].(string)
	integration := "myfusionhelper"

	output := &helpers.HelperOutput{
		Actions: make([]helpers.HelperAction, 0),
		Logs:    make([]string, 0),
	}

	// Get contact data for address resolution
	contact, err := input.Connector.GetContact(ctx, input.ContactID)
	if err != nil {
		output.Message = fmt.Sprintf("Failed to get contact: %v", err)
		return output, err
	}

	// Build address from contact fields (try billing, then shipping, then other)
	address := buildAddress(contact.CustomFields, [][]string{
		{"StreetAddress1", "StreetAddress2", "City", "State", "PostalCode", "Country"},
		{"Address2Street1", "Address2Street2", "City2", "State2", "PostalCode2", "Country2"},
		{"Address3Street1", "Address3Street2", "City3", "State3", "PostalCode3", "Country3"},
	})

	if address == "" {
		output.Logs = append(output.Logs, "No address found on contact")

		// Fire failed goal if configured
		if failedGoal, ok := input.Config["failed_goal"].(string); ok && failedGoal != "" {
			goalErr := input.Connector.AchieveGoal(ctx, input.ContactID, failedGoal, integration)
			if goalErr != nil {
				output.Logs = append(output.Logs, fmt.Sprintf("Failed to achieve failed goal: %v", goalErr))
			} else {
				output.Actions = append(output.Actions, helpers.HelperAction{
					Type:   "goal_achieved",
					Target: input.ContactID,
					Value:  failedGoal,
				})
			}
		}

		output.Success = true
		output.Message = "No address found for timezone resolution"
		return output, nil
	}

	output.Logs = append(output.Logs, fmt.Sprintf("Resolved address: %s", address))

	// Store timezone resolution request as output data for downstream processing
	// The actual geocoding and timezone resolution is handled by the execution layer
	timezoneRequest := map[string]interface{}{
		"address":    address,
		"day":        day,
		"time":       triggerTime,
		"contact_id": input.ContactID,
		"helper_id":  input.HelperID,
	}

	// Save timezone data fields if configured
	updateFields := map[string]interface{}{}

	if saveTimezone, ok := input.Config["save_time_zone"].(string); ok && saveTimezone != "" && saveTimezone != "no_select" {
		updateFields[saveTimezone] = "" // Will be populated by execution layer after geocoding
		timezoneRequest["save_time_zone_field"] = saveTimezone
	}

	if saveLatLng, ok := input.Config["save_lat_lng"].(string); ok && saveLatLng != "" && saveLatLng != "no_select" {
		timezoneRequest["save_lat_lng_field"] = saveLatLng
	}

	if saveOffset, ok := input.Config["save_time_zone_offset"].(string); ok && saveOffset != "" && saveOffset != "no_select" {
		timezoneRequest["save_time_zone_offset_field"] = saveOffset
	}

	// Fire the trigger goal
	if triggerGoal, ok := input.Config["trigger_goal"].(string); ok && triggerGoal != "" {
		timezoneRequest["trigger_goal"] = triggerGoal
	}

	output.Success = true
	output.Message = fmt.Sprintf("Timezone trigger scheduled for %s %s based on contact address", day, triggerTime)
	output.Actions = []helpers.HelperAction{
		{
			Type:   "timezone_trigger_scheduled",
			Target: input.ContactID,
			Value:  timezoneRequest,
		},
	}
	output.ModifiedData = timezoneRequest
	output.Logs = append(output.Logs, fmt.Sprintf("Timezone trigger for contact %s: %s %s", input.ContactID, day, triggerTime))

	return output, nil
}

// buildAddress tries multiple address field sets and returns the first non-empty one.
func buildAddress(fields map[string]interface{}, fieldSets [][]string) string {
	if fields == nil {
		return ""
	}

	for _, fieldSet := range fieldSets {
		var parts []string
		for _, fieldName := range fieldSet {
			if v, ok := fields[fieldName]; ok && v != nil {
				s := fmt.Sprintf("%v", v)
				if s != "" {
					parts = append(parts, s)
				}
			}
		}
		combined := strings.Join(parts, " ")
		if strings.TrimSpace(combined) != "" {
			return strings.TrimSpace(combined)
		}
	}

	return ""
}
