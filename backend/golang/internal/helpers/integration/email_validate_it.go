package integration

import (
	"context"
	"fmt"
	"net"
	"regexp"
	"strings"

	"github.com/myfusionhelper/api/internal/helpers"
)

func init() {
	helpers.Register("email_validate_it", func() helpers.Helper { return &EmailValidateIt{} })
}

// EmailValidateIt validates an email address format and optionally performs MX record lookup.
// Stores the validation result (valid/invalid) in a configurable contact field.
type EmailValidateIt struct{}

func (h *EmailValidateIt) GetName() string     { return "Email Validate It" }
func (h *EmailValidateIt) GetType() string     { return "email_validate_it" }
func (h *EmailValidateIt) GetCategory() string { return "integration" }
func (h *EmailValidateIt) GetDescription() string {
	return "Validate email format and optionally check MX records for deliverability"
}
func (h *EmailValidateIt) RequiresCRM() bool       { return true }
func (h *EmailValidateIt) SupportedCRMs() []string { return nil }

func (h *EmailValidateIt) GetConfigSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"email_field": map[string]interface{}{
				"type":        "string",
				"description": "The contact field containing the email address to validate",
				"default":     "Email",
			},
			"result_field": map[string]interface{}{
				"type":        "string",
				"description": "The contact field to store the validation result (valid/invalid)",
			},
			"check_mx": map[string]interface{}{
				"type":        "boolean",
				"description": "Whether to perform MX record lookup for the email domain",
				"default":     true,
			},
			"valid_goal": map[string]interface{}{
				"type":        "string",
				"description": "Goal call name to achieve when email is valid",
			},
			"invalid_goal": map[string]interface{}{
				"type":        "string",
				"description": "Goal call name to achieve when email is invalid",
			},
		},
		"required": []string{"result_field"},
	}
}

func (h *EmailValidateIt) ValidateConfig(config map[string]interface{}) error {
	if _, ok := config["result_field"].(string); !ok || config["result_field"] == "" {
		return fmt.Errorf("result_field is required")
	}
	return nil
}

// emailRegex is a basic email format validation pattern
var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

func (h *EmailValidateIt) Execute(ctx context.Context, input helpers.HelperInput) (*helpers.HelperOutput, error) {
	integration := "myfusionhelper"

	emailField := "Email"
	if ef, ok := input.Config["email_field"].(string); ok && ef != "" {
		emailField = ef
	}

	resultField := input.Config["result_field"].(string)

	checkMX := true
	if cm, ok := input.Config["check_mx"].(bool); ok {
		checkMX = cm
	}

	output := &helpers.HelperOutput{
		Actions: make([]helpers.HelperAction, 0),
		Logs:    make([]string, 0),
	}

	// Get the email value from the contact
	emailValue, err := input.Connector.GetContactFieldValue(ctx, input.ContactID, emailField)
	if err != nil || emailValue == nil || fmt.Sprintf("%v", emailValue) == "" {
		output.Logs = append(output.Logs, fmt.Sprintf("Email field '%s' is empty", emailField))

		// Store invalid result
		setErr := input.Connector.SetContactFieldValue(ctx, input.ContactID, resultField, "invalid")
		if setErr != nil {
			output.Logs = append(output.Logs, fmt.Sprintf("Failed to set result field: %v", setErr))
		}

		output.Success = true
		output.Message = "Email field is empty"
		output.ModifiedData = map[string]interface{}{
			"status":      "invalid",
			"reason":      "empty",
			resultField:   "invalid",
		}
		return output, nil
	}

	email := strings.TrimSpace(fmt.Sprintf("%v", emailValue))
	output.Logs = append(output.Logs, fmt.Sprintf("Validating email: %s", email))

	// Step 1: Format validation
	if !emailRegex.MatchString(email) {
		output.Logs = append(output.Logs, "Email format is invalid")

		setErr := input.Connector.SetContactFieldValue(ctx, input.ContactID, resultField, "invalid")
		if setErr != nil {
			output.Logs = append(output.Logs, fmt.Sprintf("Failed to set result field: %v", setErr))
		} else {
			output.Actions = append(output.Actions, helpers.HelperAction{
				Type:   "field_updated",
				Target: resultField,
				Value:  "invalid",
			})
		}

		// Fire invalid goal if configured
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
		output.Message = fmt.Sprintf("Email '%s' has invalid format", email)
		output.ModifiedData = map[string]interface{}{
			"status":    "invalid",
			"reason":    "format",
			"email":     email,
			resultField: "invalid",
		}
		return output, nil
	}

	// Step 2: MX record lookup (if enabled)
	mxValid := true
	mxReason := ""
	if checkMX {
		parts := strings.SplitN(email, "@", 2)
		if len(parts) != 2 {
			mxValid = false
			mxReason = "invalid_domain"
		} else {
			domain := parts[1]
			mxRecords, mxErr := net.LookupMX(domain)
			if mxErr != nil || len(mxRecords) == 0 {
				mxValid = false
				mxReason = "no_mx_records"
				output.Logs = append(output.Logs, fmt.Sprintf("No MX records found for domain '%s'", domain))
			} else {
				output.Logs = append(output.Logs, fmt.Sprintf("MX records found for domain '%s': %d record(s)", domain, len(mxRecords)))
			}
		}
	}

	if !mxValid {
		setErr := input.Connector.SetContactFieldValue(ctx, input.ContactID, resultField, "invalid")
		if setErr != nil {
			output.Logs = append(output.Logs, fmt.Sprintf("Failed to set result field: %v", setErr))
		} else {
			output.Actions = append(output.Actions, helpers.HelperAction{
				Type:   "field_updated",
				Target: resultField,
				Value:  "invalid",
			})
		}

		// Fire invalid goal if configured
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
		output.Message = fmt.Sprintf("Email '%s' failed MX validation", email)
		output.ModifiedData = map[string]interface{}{
			"status":    "invalid",
			"reason":    mxReason,
			"email":     email,
			resultField: "invalid",
		}
		return output, nil
	}

	// Email is valid
	setErr := input.Connector.SetContactFieldValue(ctx, input.ContactID, resultField, "valid")
	if setErr != nil {
		output.Logs = append(output.Logs, fmt.Sprintf("Failed to set result field: %v", setErr))
	} else {
		output.Actions = append(output.Actions, helpers.HelperAction{
			Type:   "field_updated",
			Target: resultField,
			Value:  "valid",
		})
	}

	// Fire valid goal if configured
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

	output.Success = true
	output.Message = fmt.Sprintf("Email '%s' is valid", email)
	output.ModifiedData = map[string]interface{}{
		"status":    "valid",
		"email":     email,
		"check_mx":  checkMX,
		resultField: "valid",
	}
	output.Logs = append(output.Logs, fmt.Sprintf("Email validation for contact %s: %s -> valid", input.ContactID, email))

	return output, nil
}
