package data

import (
	"context"
	"fmt"
	"time"

	"github.com/myfusionhelper/api/internal/helpers"
)

// NewLastClickIt creates a new LastClickIt helper instance
func NewLastClickIt() helpers.Helper { return &LastClickIt{} }

func init() {
	helpers.Register("last_click_it", func() helpers.Helper { return &LastClickIt{} })
}

// LastClickIt retrieves the last email click date for a contact and stores it
// in a specified field. Queries email engagement stats via the CRM connector.
// Ported from legacy PHP last_click_it helper.
type LastClickIt struct{}

func (h *LastClickIt) GetName() string     { return "Last Click It" }
func (h *LastClickIt) GetType() string     { return "last_click_it" }
func (h *LastClickIt) GetCategory() string { return "data" }
func (h *LastClickIt) GetDescription() string {
	return "Retrieve the last email click date for a contact and save it to a field"
}
func (h *LastClickIt) RequiresCRM() bool       { return true }
func (h *LastClickIt) SupportedCRMs() []string { return nil }

func (h *LastClickIt) GetConfigSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"email_field": map[string]interface{}{
				"type":        "string",
				"description": "The email field to look up (e.g., Email, Email2, Email3)",
				"default":     "Email",
			},
			"save_to": map[string]interface{}{
				"type":        "string",
				"description": "The contact field to store the last click date",
			},
			"attribution_window_days": map[string]interface{}{
				"type":        "number",
				"description": "Only track clicks within this many days (0 = all clicks)",
				"default":     0,
			},
			"click_source_field": map[string]interface{}{
				"type":        "string",
				"description": "Field to store click source (email, sms, web)",
			},
			"conversion_tracking": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"enabled": map[string]interface{}{
						"type":        "boolean",
						"description": "Enable conversion tracking",
						"default":     false,
					},
					"goal_name": map[string]interface{}{
						"type":        "string",
						"description": "Goal to check for conversion",
					},
					"conversion_field": map[string]interface{}{
						"type":        "string",
						"description": "Field to store conversion status (true/false)",
					},
				},
			},
			"click_frequency_analysis": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"enabled": map[string]interface{}{
						"type":        "boolean",
						"description": "Enable click frequency analysis",
						"default":     false,
					},
					"analysis_period_days": map[string]interface{}{
						"type":        "number",
						"description": "Period for frequency analysis (e.g., 7 for per week)",
						"default":     7,
					},
					"frequency_field": map[string]interface{}{
						"type":        "string",
						"description": "Field to store clicks per period",
					},
				},
			},
		},
		"required": []string{"save_to"},
	}
}

func (h *LastClickIt) ValidateConfig(config map[string]interface{}) error {
	if _, ok := config["save_to"].(string); !ok || config["save_to"] == "" {
		return fmt.Errorf("save_to field is required")
	}
	return nil
}

func (h *LastClickIt) Execute(ctx context.Context, input helpers.HelperInput) (*helpers.HelperOutput, error) {
	saveTo := input.Config["save_to"].(string)

	emailField := "Email"
	if ef, ok := input.Config["email_field"].(string); ok && ef != "" {
		emailField = ef
	}

	// Normalize email field names
	if emailField == "Email2" {
		emailField = "EmailAddress2"
	}
	if emailField == "Email3" {
		emailField = "EmailAddress3"
	}

	output := &helpers.HelperOutput{
		Logs:    make([]string, 0),
		Actions: make([]helpers.HelperAction, 0),
	}

	// Parse configuration
	attributionWindowDays := 0
	if awd, ok := input.Config["attribution_window_days"].(float64); ok && awd > 0 {
		attributionWindowDays = int(awd)
	}

	clickSourceField := ""
	if csf, ok := input.Config["click_source_field"].(string); ok && csf != "" {
		clickSourceField = csf
	}

	conversionEnabled := false
	conversionGoal := ""
	conversionField := ""
	if ct, ok := input.Config["conversion_tracking"].(map[string]interface{}); ok {
		if enabled, ok := ct["enabled"].(bool); ok {
			conversionEnabled = enabled
		}
		if goal, ok := ct["goal_name"].(string); ok {
			conversionGoal = goal
		}
		if field, ok := ct["conversion_field"].(string); ok {
			conversionField = field
		}
	}

	freqEnabled := false
	freqPeriodDays := 7
	freqField := ""
	if cfa, ok := input.Config["click_frequency_analysis"].(map[string]interface{}); ok {
		if enabled, ok := cfa["enabled"].(bool); ok {
			freqEnabled = enabled
		}
		if period, ok := cfa["analysis_period_days"].(float64); ok && period > 0 {
			freqPeriodDays = int(period)
		}
		if field, ok := cfa["frequency_field"].(string); ok {
			freqField = field
		}
	}

	// Get the email address from the contact
	emailValue, err := input.Connector.GetContactFieldValue(ctx, input.ContactID, emailField)
	if err != nil || emailValue == nil || fmt.Sprintf("%v", emailValue) == "" {
		output.Success = true
		output.Message = fmt.Sprintf("Email field '%s' is empty, nothing to look up", emailField)
		output.Logs = append(output.Logs, output.Message)
		return output, nil
	}

	// Query email engagement stats via the connector
	// Use a composite key to indicate we want email stats
	lastClickKey := fmt.Sprintf("_email_stats.%s.LastClickDate", fmt.Sprintf("%v", emailValue))
	lastClickDate, err := input.Connector.GetContactFieldValue(ctx, input.ContactID, lastClickKey)
	if err != nil {
		output.Logs = append(output.Logs, fmt.Sprintf("Email stats query not directly supported: %v", err))

		// Fallback: try a generic email stats field
		lastClickDate, err = input.Connector.GetContactFieldValue(ctx, input.ContactID, "LastClickDate")
		if err != nil {
			output.Success = true
			output.Message = "Could not retrieve email click stats"
			output.Logs = append(output.Logs, fmt.Sprintf("Failed to get LastClickDate: %v", err))
			return output, nil
		}
	}

	if lastClickDate == nil || fmt.Sprintf("%v", lastClickDate) == "" {
		output.Success = true
		output.Message = "No click date found for contact"
		output.Logs = append(output.Logs, output.Message)
		return output, nil
	}

	dateStr := fmt.Sprintf("%v", lastClickDate)

	// Parse click date
	var clickTime time.Time
	clickTime, parseErr := time.Parse(time.RFC3339, dateStr)
	if parseErr != nil {
		// Try alternative formats
		clickTime, parseErr = time.Parse("2006-01-02", dateStr[:minLen(10, len(dateStr))])
		if parseErr != nil {
			output.Logs = append(output.Logs, fmt.Sprintf("Warning: could not parse click date '%s'", dateStr))
		}
	}

	// Apply attribution window filter
	withinWindow := true
	if attributionWindowDays > 0 && !clickTime.IsZero() {
		cutoff := time.Now().AddDate(0, 0, -attributionWindowDays)
		if clickTime.Before(cutoff) {
			withinWindow = false
			output.Success = true
			output.Message = fmt.Sprintf("Last click is outside attribution window (%d days)", attributionWindowDays)
			output.Logs = append(output.Logs, output.Message)
			return output, nil
		}
	}

	// Save the date to the target field
	if withinWindow {
		err = input.Connector.SetContactFieldValue(ctx, input.ContactID, saveTo, dateStr)
		if err != nil {
			output.Message = fmt.Sprintf("Failed to save last click date to '%s': %v", saveTo, err)
			return output, err
		}

		output.Actions = append(output.Actions, helpers.HelperAction{
			Type:   "field_updated",
			Target: saveTo,
			Value:  dateStr,
		})
	}

	// Track click source (default to email for this helper)
	clickSource := "email"
	if clickSourceField != "" {
		err = input.Connector.SetContactFieldValue(ctx, input.ContactID, clickSourceField, clickSource)
		if err != nil {
			output.Logs = append(output.Logs, fmt.Sprintf("Failed to set click source field: %v", err))
		} else {
			output.Actions = append(output.Actions, helpers.HelperAction{
				Type:   "field_updated",
				Target: clickSourceField,
				Value:  clickSource,
			})
		}
	}

	// Check conversion if enabled
	converted := false
	if conversionEnabled && conversionGoal != "" {
		// Query if goal has been achieved
		// Note: This is a simplified check - in real implementation, would query goal history
		goalCheckKey := fmt.Sprintf("_goal.%s.achieved", conversionGoal)
		goalValue, goalErr := input.Connector.GetContactFieldValue(ctx, input.ContactID, goalCheckKey)
		if goalErr == nil && goalValue != nil {
			goalStr := fmt.Sprintf("%v", goalValue)
			if goalStr == "true" || goalStr == "1" {
				converted = true
			}
		}

		if conversionField != "" {
			conversionStr := fmt.Sprintf("%t", converted)
			err = input.Connector.SetContactFieldValue(ctx, input.ContactID, conversionField, conversionStr)
			if err != nil {
				output.Logs = append(output.Logs, fmt.Sprintf("Failed to set conversion field: %v", err))
			} else {
				output.Actions = append(output.Actions, helpers.HelperAction{
					Type:   "field_updated",
					Target: conversionField,
					Value:  conversionStr,
				})
			}
		}
	}

	// Calculate click frequency if enabled
	clicksPerPeriod := 0
	if freqEnabled && !clickTime.IsZero() {
		// Query click count field
		clickCountKey := "click_count"
		clickCountVal, countErr := input.Connector.GetContactFieldValue(ctx, input.ContactID, clickCountKey)
		if countErr == nil && clickCountVal != nil {
			if count, ok := parseIntValue(clickCountVal); ok {
				clicksPerPeriod = count
			}
		}

		// Calculate frequency (simplified: total clicks / periods)
		daysSinceFirstClick := 30 // Simplified assumption
		if freqPeriodDays > 0 {
			periods := daysSinceFirstClick / freqPeriodDays
			if periods > 0 {
				clicksPerPeriod = clicksPerPeriod / periods
			}
		}

		if freqField != "" {
			err = input.Connector.SetContactFieldValue(ctx, input.ContactID, freqField, clicksPerPeriod)
			if err != nil {
				output.Logs = append(output.Logs, fmt.Sprintf("Failed to set frequency field: %v", err))
			} else {
				output.Actions = append(output.Actions, helpers.HelperAction{
					Type:   "field_updated",
					Target: freqField,
					Value:  clicksPerPeriod,
				})
			}
		}
	}

	output.Success = true
	output.Message = fmt.Sprintf("Last click date saved to '%s' (source: %s, converted: %t, freq: %d/%d days)", saveTo, clickSource, converted, clicksPerPeriod, freqPeriodDays)
	output.ModifiedData = map[string]interface{}{
		"last_click_date":    dateStr,
		"click_source":       clickSource,
		"converted":          converted,
		"clicks_per_period":  clicksPerPeriod,
		"within_window":      withinWindow,
	}
	output.Logs = append(output.Logs, fmt.Sprintf("Last click date '%s' saved to '%s' for contact %s", dateStr, saveTo, input.ContactID))

	return output, nil
}

func minLen(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func parseIntValue(v interface{}) (int, bool) {
	switch val := v.(type) {
	case int:
		return val, true
	case float64:
		return int(val), true
	case string:
		var i int
		_, err := fmt.Sscanf(val, "%d", &i)
		return i, err == nil
	default:
		return 0, false
	}
}
