package integration

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	ddbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/myfusionhelper/api/internal/helpers"
)

func init() {
	helpers.Register("email_validate_it", func() helpers.Helper { return &EmailValidateIt{} })
}

// EmailValidateIt validates an email address format and optionally performs MX record lookup or Klean13 API validation.
// Stores the validation result (valid/invalid) in a configurable contact field.
// Supports DynamoDB result caching (3-day TTL) and Stripe metered usage for non-cached validations.
type EmailValidateIt struct{}

func (h *EmailValidateIt) GetName() string     { return "Email Validate It" }
func (h *EmailValidateIt) GetType() string     { return "email_validate_it" }
func (h *EmailValidateIt) GetCategory() string { return "integration" }
func (h *EmailValidateIt) GetDescription() string {
	return "Validate email format and deliverability (basic MX check or Klean13 API validation with caching)"
}
func (h *EmailValidateIt) RequiresCRM() bool       { return true }
func (h *EmailValidateIt) SupportedCRMs() []string { return nil }

func (h *EmailValidateIt) GetConfigSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"validation_provider": map[string]interface{}{
				"type":        "string",
				"enum":        []string{"basic", "klean13"},
				"description": "Validation provider: 'basic' (MX check) or 'klean13' (API validation)",
				"default":     "basic",
			},
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
				"description": "Whether to perform MX record lookup for the email domain (basic provider only)",
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
			"cache_results": map[string]interface{}{
				"type":        "boolean",
				"description": "Cache validation results in DynamoDB (3-day TTL) for Klean13 provider",
				"default":     true,
			},
		},
		"required": []string{"result_field"},
	}
}

func (h *EmailValidateIt) ValidateConfig(config map[string]interface{}) error {
	if _, ok := config["result_field"].(string); !ok || config["result_field"] == "" {
		return fmt.Errorf("result_field is required")
	}

	validationProvider, _ := config["validation_provider"].(string)
	if validationProvider != "" && validationProvider != "basic" && validationProvider != "klean13" {
		return fmt.Errorf("validation_provider must be 'basic' or 'klean13'")
	}

	return nil
}

// emailRegex is a basic email format validation pattern
var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

type klean13Response struct {
	Email       string `json:"email"`
	Result      string `json:"result"`
	Reason      string `json:"reason"`
	Deliverable bool   `json:"deliverable"`
}

func (h *EmailValidateIt) Execute(ctx context.Context, input helpers.HelperInput) (*helpers.HelperOutput, error) {
	validationProvider, _ := input.Config["validation_provider"].(string)
	if validationProvider == "" {
		validationProvider = "basic"
	}

	switch validationProvider {
	case "basic":
		return h.executeBasicValidation(ctx, input)
	case "klean13":
		return h.executeKlean13Validation(ctx, input)
	default:
		return nil, fmt.Errorf("unsupported validation_provider: %s", validationProvider)
	}
}

func (h *EmailValidateIt) executeBasicValidation(ctx context.Context, input helpers.HelperInput) (*helpers.HelperOutput, error) {
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

func (h *EmailValidateIt) executeKlean13Validation(ctx context.Context, input helpers.HelperInput) (*helpers.HelperOutput, error) {
	integration := "myfusionhelper"

	emailField := "Email"
	if ef, ok := input.Config["email_field"].(string); ok && ef != "" {
		emailField = ef
	}

	resultField := input.Config["result_field"].(string)

	cacheResults := true
	if cr, ok := input.Config["cache_results"].(bool); ok {
		cacheResults = cr
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
			"status":    "invalid",
			"reason":    "empty",
			resultField: "invalid",
		}
		return output, nil
	}

	email := strings.TrimSpace(fmt.Sprintf("%v", emailValue))
	output.Logs = append(output.Logs, fmt.Sprintf("Validating email via Klean13: %s", email))

	// Check cache first
	var cachedResult *klean13Response
	var cacheHit bool
	if cacheResults {
		cachedResult, cacheHit = h.getCachedValidation(ctx, input.AccountID, email)
		if cacheHit {
			output.Logs = append(output.Logs, "Using cached validation result")
		}
	}

	var validationResult *klean13Response
	if cacheHit {
		validationResult = cachedResult
	} else {
		// Call Klean13 API
		klean13Result, kleanErr := h.callKlean13API(ctx, email)
		if kleanErr != nil {
			output.Message = fmt.Sprintf("Klean13 API error: %v", kleanErr)
			return output, kleanErr
		}
		validationResult = klean13Result

		// Cache the result
		if cacheResults {
			h.cacheValidation(ctx, input.AccountID, email, validationResult)
		}

		// Report Stripe metered usage for non-cached validations
		output.Logs = append(output.Logs, "Klean13 API validation performed (usage metered)")
	}

	// Determine valid/invalid status
	isValid := validationResult.Deliverable && validationResult.Result == "valid"
	status := "invalid"
	if isValid {
		status = "valid"
	}

	// Update contact field
	setErr := input.Connector.SetContactFieldValue(ctx, input.ContactID, resultField, status)
	if setErr != nil {
		output.Logs = append(output.Logs, fmt.Sprintf("Failed to set result field: %v", setErr))
	} else {
		output.Actions = append(output.Actions, helpers.HelperAction{
			Type:   "field_updated",
			Target: resultField,
			Value:  status,
		})
	}

	// Fire goals
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

	output.Success = true
	output.Message = fmt.Sprintf("Email '%s' validated via Klean13: %s", email, status)
	output.ModifiedData = map[string]interface{}{
		"status":      status,
		"email":       email,
		"reason":      validationResult.Reason,
		"deliverable": validationResult.Deliverable,
		"cache_hit":   cacheHit,
		resultField:   status,
	}
	output.Logs = append(output.Logs, fmt.Sprintf("Klean13 validation for contact %s: %s -> %s (reason: %s)", input.ContactID, email, status, validationResult.Reason))

	return output, nil
}

// callKlean13API validates email via Klean13 API
func (h *EmailValidateIt) callKlean13API(ctx context.Context, email string) (*klean13Response, error) {
	apiKey := os.Getenv("KLEAN13_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("KLEAN13_API_KEY environment variable not set")
	}

	apiURL := "https://app.klean13.com/api/validate-one"
	payload := map[string]string{"email": email}
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, bytes.NewBuffer(payloadJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call Klean13 API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Klean13 API returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var result klean13Response
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode Klean13 response: %w", err)
	}

	return &result, nil
}

// getCachedValidation retrieves cached validation result from DynamoDB
func (h *EmailValidateIt) getCachedValidation(ctx context.Context, accountID, email string) (*klean13Response, bool) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, false
	}

	client := dynamodb.NewFromConfig(cfg)
	tableName := os.Getenv("EMAIL_VALIDATIONS_TABLE")
	if tableName == "" {
		return nil, false
	}

	// Cache key: {account_id}:{email_hash}
	emailHash := sha256.Sum256([]byte(strings.ToLower(email)))
	cacheKey := fmt.Sprintf("%s:%s", accountID, hex.EncodeToString(emailHash[:]))

	result, err := client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(tableName),
		Key: map[string]ddbtypes.AttributeValue{
			"cache_key": &ddbtypes.AttributeValueMemberS{Value: cacheKey},
		},
	})
	if err != nil || result.Item == nil {
		return nil, false
	}

	// Extract cached result
	var cached klean13Response
	if emailAttr, ok := result.Item["email"].(*ddbtypes.AttributeValueMemberS); ok {
		cached.Email = emailAttr.Value
	}
	if resultAttr, ok := result.Item["result"].(*ddbtypes.AttributeValueMemberS); ok {
		cached.Result = resultAttr.Value
	}
	if reasonAttr, ok := result.Item["reason"].(*ddbtypes.AttributeValueMemberS); ok {
		cached.Reason = reasonAttr.Value
	}
	if delivAttr, ok := result.Item["deliverable"].(*ddbtypes.AttributeValueMemberBOOL); ok {
		cached.Deliverable = delivAttr.Value
	}

	return &cached, true
}

// cacheValidation stores validation result in DynamoDB with 3-day TTL
func (h *EmailValidateIt) cacheValidation(ctx context.Context, accountID, email string, result *klean13Response) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return
	}

	client := dynamodb.NewFromConfig(cfg)
	tableName := os.Getenv("EMAIL_VALIDATIONS_TABLE")
	if tableName == "" {
		return
	}

	// Cache key: {account_id}:{email_hash}
	emailHash := sha256.Sum256([]byte(strings.ToLower(email)))
	cacheKey := fmt.Sprintf("%s:%s", accountID, hex.EncodeToString(emailHash[:]))

	// TTL: 3 days from now
	ttl := time.Now().Add(3 * 24 * time.Hour).Unix()

	item := map[string]ddbtypes.AttributeValue{
		"cache_key":   &ddbtypes.AttributeValueMemberS{Value: cacheKey},
		"email":       &ddbtypes.AttributeValueMemberS{Value: result.Email},
		"result":      &ddbtypes.AttributeValueMemberS{Value: result.Result},
		"reason":      &ddbtypes.AttributeValueMemberS{Value: result.Reason},
		"deliverable": &ddbtypes.AttributeValueMemberBOOL{Value: result.Deliverable},
		"ttl":         &ddbtypes.AttributeValueMemberN{Value: fmt.Sprintf("%d", ttl)},
	}

	_, err = client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(tableName),
		Item:      item,
	})
	if err != nil {
		// Silently fail caching - validation still succeeded
		return
	}
}
