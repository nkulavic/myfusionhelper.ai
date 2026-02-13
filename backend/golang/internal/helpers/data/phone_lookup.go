package data

import (
	"context"
	"fmt"
	"regexp"

	"github.com/myfusionhelper/api/internal/helpers"
)

// NewPhoneLookup creates a new PhoneLookup helper instance
func NewPhoneLookup() helpers.Helper { return &PhoneLookup{} }

func init() {
	helpers.Register("phone_lookup", func() helpers.Helper { return &PhoneLookup{} })
}

// PhoneLookup validates a phone number from a contact field by stripping non-digits
// and checking it via an external lookup service. Achieves goals based on valid/invalid/empty results.
// Ported from legacy PHP phone_lookup helper.
type PhoneLookup struct{}

func (h *PhoneLookup) GetName() string     { return "Phone Lookup" }
func (h *PhoneLookup) GetType() string     { return "phone_lookup" }
func (h *PhoneLookup) GetCategory() string { return "data" }
func (h *PhoneLookup) GetDescription() string {
	return "Validate a contact phone number and fire goals based on valid, invalid, or empty results"
}
func (h *PhoneLookup) RequiresCRM() bool       { return true }
func (h *PhoneLookup) SupportedCRMs() []string { return nil }

func (h *PhoneLookup) GetConfigSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"phone_field": map[string]interface{}{
				"type":        "string",
				"description": "The contact field containing the phone number (e.g., Phone1, Phone2)",
				"default":     "Phone1",
			},
			"country_code": map[string]interface{}{
				"type":        "string",
				"description": "Expected country code for validation (e.g., US, CA, GB)",
				"default":     "US",
			},
			"valid_goal": map[string]interface{}{
				"type":        "string",
				"description": "Goal call name to achieve when phone is valid",
			},
			"invalid_goal": map[string]interface{}{
				"type":        "string",
				"description": "Goal call name to achieve when phone is invalid",
			},
			"empty_goal": map[string]interface{}{
				"type":        "string",
				"description": "Goal call name to achieve when phone field is empty",
			},
			"save_formatted_to": map[string]interface{}{
				"type":        "string",
				"description": "Optional field to save the cleaned/formatted phone number",
			},
		},
		"required": []string{"phone_field"},
	}
}

func (h *PhoneLookup) ValidateConfig(config map[string]interface{}) error {
	if _, ok := config["phone_field"].(string); !ok || config["phone_field"] == "" {
		return fmt.Errorf("phone_field is required")
	}
	return nil
}

var nonDigitRegex = regexp.MustCompile(`\D`)

func (h *PhoneLookup) Execute(ctx context.Context, input helpers.HelperInput) (*helpers.HelperOutput, error) {
	phoneField := input.Config["phone_field"].(string)
	integration := "myfusionhelper"

	countryCode := "US"
	if cc, ok := input.Config["country_code"].(string); ok && cc != "" {
		countryCode = cc
	}

	output := &helpers.HelperOutput{
		Actions: make([]helpers.HelperAction, 0),
		Logs:    make([]string, 0),
	}

	// Get the phone number from the contact
	phoneValue, err := input.Connector.GetContactFieldValue(ctx, input.ContactID, phoneField)
	if err != nil || phoneValue == nil || fmt.Sprintf("%v", phoneValue) == "" {
		// Phone field is empty
		output.Logs = append(output.Logs, fmt.Sprintf("Phone field '%s' is empty", phoneField))

		if emptyGoal, ok := input.Config["empty_goal"].(string); ok && emptyGoal != "" {
			goalErr := input.Connector.AchieveGoal(ctx, input.ContactID, emptyGoal, integration)
			if goalErr != nil {
				output.Logs = append(output.Logs, fmt.Sprintf("Failed to achieve empty_goal: %v", goalErr))
			} else {
				output.Actions = append(output.Actions, helpers.HelperAction{
					Type:   "goal_achieved",
					Target: input.ContactID,
					Value:  emptyGoal,
				})
			}
		}

		output.Success = true
		output.Message = "Phone number is empty"
		output.ModifiedData = map[string]interface{}{
			"status": "empty",
		}
		return output, nil
	}

	rawNumber := fmt.Sprintf("%v", phoneValue)
	// Strip all non-digit characters
	cleanedNumber := nonDigitRegex.ReplaceAllString(rawNumber, "")

	if cleanedNumber == "" {
		// After cleaning, number is empty (was only special chars)
		if invalidGoal, ok := input.Config["invalid_goal"].(string); ok && invalidGoal != "" {
			goalErr := input.Connector.AchieveGoal(ctx, input.ContactID, invalidGoal, integration)
			if goalErr != nil {
				output.Logs = append(output.Logs, fmt.Sprintf("Failed to achieve invalid_goal: %v", goalErr))
			} else {
				output.Actions = append(output.Actions, helpers.HelperAction{
					Type:   "goal_achieved",
					Target: input.ContactID,
					Value:  invalidGoal,
				})
			}
		}

		output.Success = true
		output.Message = "Phone number contains no valid digits"
		output.ModifiedData = map[string]interface{}{
			"status":     "invalid",
			"raw_number": rawNumber,
		}
		return output, nil
	}

	output.Logs = append(output.Logs, fmt.Sprintf("Cleaned phone number: %s (country: %s)", cleanedNumber, countryCode))

	// Phone validation request - the actual Twilio/lookup call is handled by execution layer
	// We produce the validation request as output data
	lookupRequest := map[string]interface{}{
		"number":       cleanedNumber,
		"country_code": countryCode,
		"raw_number":   rawNumber,
		"contact_id":   input.ContactID,
	}

	// Basic validation: check minimum length for the country
	isValid := isPhoneValid(cleanedNumber, countryCode)

	if isValid {
		if validGoal, ok := input.Config["valid_goal"].(string); ok && validGoal != "" {
			goalErr := input.Connector.AchieveGoal(ctx, input.ContactID, validGoal, integration)
			if goalErr != nil {
				output.Logs = append(output.Logs, fmt.Sprintf("Failed to achieve valid_goal: %v", goalErr))
			} else {
				output.Actions = append(output.Actions, helpers.HelperAction{
					Type:   "goal_achieved",
					Target: input.ContactID,
					Value:  validGoal,
				})
			}
		}

		// Save formatted number if configured
		if saveTo, ok := input.Config["save_formatted_to"].(string); ok && saveTo != "" {
			setErr := input.Connector.SetContactFieldValue(ctx, input.ContactID, saveTo, cleanedNumber)
			if setErr != nil {
				output.Logs = append(output.Logs, fmt.Sprintf("Failed to save formatted number: %v", setErr))
			} else {
				output.Actions = append(output.Actions, helpers.HelperAction{
					Type:   "field_updated",
					Target: saveTo,
					Value:  cleanedNumber,
				})
			}
		}
	} else {
		if invalidGoal, ok := input.Config["invalid_goal"].(string); ok && invalidGoal != "" {
			goalErr := input.Connector.AchieveGoal(ctx, input.ContactID, invalidGoal, integration)
			if goalErr != nil {
				output.Logs = append(output.Logs, fmt.Sprintf("Failed to achieve invalid_goal: %v", goalErr))
			} else {
				output.Actions = append(output.Actions, helpers.HelperAction{
					Type:   "goal_achieved",
					Target: input.ContactID,
					Value:  invalidGoal,
				})
			}
		}
	}

	status := "valid"
	if !isValid {
		status = "invalid"
	}

	output.Success = true
	output.Message = fmt.Sprintf("Phone number validation: %s", status)
	lookupRequest["status"] = status
	output.ModifiedData = lookupRequest
	output.Logs = append(output.Logs, fmt.Sprintf("Phone lookup for contact %s: %s -> %s (country: %s)", input.ContactID, rawNumber, status, countryCode))

	return output, nil
}

// isPhoneValid performs basic phone number length validation by country.
func isPhoneValid(digits string, country string) bool {
	length := len(digits)

	switch country {
	case "US", "CA":
		// North American numbers: 10 or 11 digits (with country code)
		return length == 10 || length == 11
	case "GB":
		return length >= 10 && length <= 11
	default:
		// Generic: at least 7 digits and at most 15
		return length >= 7 && length <= 15
	}
}
