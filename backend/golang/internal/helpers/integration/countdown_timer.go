package integration

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	ddbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/golang-jwt/jwt/v5"
	"github.com/myfusionhelper/api/internal/database"
	"github.com/myfusionhelper/api/internal/helpers"
)

func init() {
	helpers.Register("countdown_timer", func() helpers.Helper { return &CountdownTimer{} })
}

// CountdownTimer generates secure countdown timer URLs with JWT signing.
// Supports multiple modes: dynamic, standard, evergreen, and contact_field.
// Evergreen mode uses DynamoDB for per-contact persistent timers.
type CountdownTimer struct{}

func (h *CountdownTimer) GetName() string     { return "Countdown Timer" }
func (h *CountdownTimer) GetType() string     { return "countdown_timer" }
func (h *CountdownTimer) GetCategory() string { return "integration" }
func (h *CountdownTimer) GetDescription() string {
	return "Generate secure countdown timer URLs with JWT signing (dynamic, standard, evergreen, contact_field modes)"
}
func (h *CountdownTimer) RequiresCRM() bool       { return true }
func (h *CountdownTimer) SupportedCRMs() []string { return nil }

func (h *CountdownTimer) GetConfigSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"timer_mode": map[string]interface{}{
				"type":        "string",
				"enum":        []string{"dynamic", "standard", "evergreen", "contact_field"},
				"description": "Timer mode: dynamic (from now), standard (fixed date), evergreen (persistent per-contact), contact_field (read from CRM field)",
			},
			"expire_date": map[string]interface{}{
				"type":        "string",
				"description": "Expiry date for standard mode (ISO 8601: 2024-12-31T23:59:59Z)",
			},
			"duration_hours": map[string]interface{}{
				"type":        "number",
				"description": "Duration in hours from now for dynamic mode",
			},
			"timer_field": map[string]interface{}{
				"type":        "string",
				"description": "CRM field containing expiry date for contact_field mode",
			},
			"save_url_field": map[string]interface{}{
				"type":        "string",
				"description": "Optional CRM field to save timer URL to",
			},
			"jwt_secret": map[string]interface{}{
				"type":        "string",
				"description": "Secret key for JWT signing (required)",
			},
			"timer_url_base": map[string]interface{}{
				"type":        "string",
				"description": "Base URL for timer display (defaults to app.myfusionhelper.ai/timer)",
				"default":     "https://app.myfusionhelper.ai/timer",
			},
		},
		"required": []string{"timer_mode", "jwt_secret"},
	}
}

func (h *CountdownTimer) ValidateConfig(config map[string]interface{}) error {
	mode := getStringConfigValue(config, "timer_mode", "")
	if mode == "" {
		return fmt.Errorf("timer_mode is required")
	}

	validModes := map[string]bool{
		"dynamic":       true,
		"standard":      true,
		"evergreen":     true,
		"contact_field": true,
	}

	if !validModes[mode] {
		return fmt.Errorf("invalid timer_mode: %s", mode)
	}

	if _, ok := config["jwt_secret"].(string); !ok || config["jwt_secret"].(string) == "" {
		return fmt.Errorf("jwt_secret is required")
	}

	// Mode-specific validation
	switch mode {
	case "standard":
		if _, ok := config["expire_date"].(string); !ok || config["expire_date"].(string) == "" {
			return fmt.Errorf("expire_date is required for standard mode")
		}
	case "dynamic":
		if _, ok := config["duration_hours"]; !ok {
			return fmt.Errorf("duration_hours is required for dynamic mode")
		}
	case "contact_field":
		if _, ok := config["timer_field"].(string); !ok || config["timer_field"].(string) == "" {
			return fmt.Errorf("timer_field is required for contact_field mode")
		}
	}

	return nil
}

func (h *CountdownTimer) Execute(ctx context.Context, input helpers.HelperInput) (*helpers.HelperOutput, error) {
	output := &helpers.HelperOutput{
		Logs: make([]string, 0),
	}

	mode := getStringConfigValue(input.Config, "timer_mode", "")
	jwtSecret := getStringConfigValue(input.Config, "jwt_secret", "")
	baseURL := getStringConfigValue(input.Config, "timer_url_base", "https://app.myfusionhelper.ai/timer")

	output.Logs = append(output.Logs, fmt.Sprintf("Countdown timer mode: %s", mode))

	var expiresAt time.Time
	var err error

	switch mode {
	case "dynamic":
		expiresAt, err = h.executeDynamic(ctx, input, output)
	case "standard":
		expiresAt, err = h.executeStandard(ctx, input, output)
	case "evergreen":
		expiresAt, err = h.executeEvergreen(ctx, input, output)
	case "contact_field":
		expiresAt, err = h.executeContactField(ctx, input, output)
	default:
		output.Success = false
		output.Message = fmt.Sprintf("Invalid timer mode: %s", mode)
		return output, fmt.Errorf("invalid timer mode: %s", mode)
	}

	if err != nil {
		output.Success = false
		output.Message = fmt.Sprintf("Failed to calculate expiry: %v", err)
		return output, err
	}

	// Generate JWT token
	timerURL, err := h.generateTimerURL(baseURL, jwtSecret, input.ContactID, input.HelperID, expiresAt)
	if err != nil {
		output.Success = false
		output.Message = fmt.Sprintf("Failed to generate timer URL: %v", err)
		return output, err
	}

	output.Logs = append(output.Logs, fmt.Sprintf("Generated timer URL expires at: %s", expiresAt.Format(time.RFC3339)))

	// Optionally save URL to CRM field
	if saveField := getStringConfigValue(input.Config, "save_url_field", ""); saveField != "" {
		if err := input.Connector.SetContactFieldValue(ctx, input.ContactID, saveField, timerURL); err != nil {
			output.Logs = append(output.Logs, fmt.Sprintf("Warning: failed to save timer URL to field '%s': %v", saveField, err))
		} else {
			output.Logs = append(output.Logs, fmt.Sprintf("Saved timer URL to field '%s'", saveField))
		}
	}

	output.Success = true
	output.Message = "Countdown timer URL generated"
	output.ModifiedData = map[string]interface{}{
		"timer_url":  timerURL,
		"expires_at": expiresAt.Format(time.RFC3339),
		"mode":       mode,
	}
	output.Actions = []helpers.HelperAction{
		{
			Type:   "timer_generated",
			Target: input.ContactID,
			Value:  timerURL,
		},
	}

	return output, nil
}

// executeDynamic calculates expiry as duration_hours from now
func (h *CountdownTimer) executeDynamic(ctx context.Context, input helpers.HelperInput, output *helpers.HelperOutput) (time.Time, error) {
	durationHours := getFloatConfigValue(input.Config, "duration_hours", 24.0)
	expiresAt := time.Now().UTC().Add(time.Duration(durationHours * float64(time.Hour)))
	output.Logs = append(output.Logs, fmt.Sprintf("Dynamic mode: expires in %.1f hours", durationHours))
	return expiresAt, nil
}

// executeStandard uses fixed expire_date from config
func (h *CountdownTimer) executeStandard(ctx context.Context, input helpers.HelperInput, output *helpers.HelperOutput) (time.Time, error) {
	expireDateStr := getStringConfigValue(input.Config, "expire_date", "")
	expiresAt, err := time.Parse(time.RFC3339, expireDateStr)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid expire_date format (must be ISO 8601): %w", err)
	}
	output.Logs = append(output.Logs, fmt.Sprintf("Standard mode: fixed expiry date %s", expireDateStr))
	return expiresAt, nil
}

// executeEvergreen creates persistent per-contact timer in DynamoDB
func (h *CountdownTimer) executeEvergreen(ctx context.Context, input helpers.HelperInput, output *helpers.HelperOutput) (time.Time, error) {
	// Generate unique timer key based on helper_id + contact_id
	timerKey := h.generateTimerKey(input.HelperID, input.ContactID)

	// Get or create timer in DynamoDB
	stage := os.Getenv("STAGE")
	if stage == "" {
		stage = "dev"
	}
	tableName := fmt.Sprintf("mfh-%s-countdown-timers", stage)

	// Create DynamoDB client
	ddbClient, err := database.NewDynamoDBClient(ctx)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to create DynamoDB client: %w", err)
	}

	// Try to get existing timer
	key := map[string]ddbtypes.AttributeValue{
		"timer_key": &ddbtypes.AttributeValueMemberS{Value: timerKey},
	}

	result, err := ddbClient.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(tableName),
		Key:       key,
	})
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to query timer: %w", err)
	}

	// If timer exists and not expired, return existing expiry
	if result.Item != nil {
		if expiresAtAttr, ok := result.Item["expires_at"]; ok {
			if expiresAtNum, ok := expiresAtAttr.(*ddbtypes.AttributeValueMemberN); ok {
				expiresAtUnix := parseUnixTimestamp(expiresAtNum.Value)
				expiresAt := time.Unix(expiresAtUnix, 0).UTC()

				if time.Now().UTC().Before(expiresAt) {
					output.Logs = append(output.Logs, fmt.Sprintf("Evergreen mode: reusing existing timer (expires %s)", expiresAt.Format(time.RFC3339)))
					return expiresAt, nil
				}
			}
		}
	}

	// Create new timer (24 hours from now by default)
	durationHours := getFloatConfigValue(input.Config, "duration_hours", 24.0)
	expiresAt := time.Now().UTC().Add(time.Duration(durationHours * float64(time.Hour)))
	expiresAtUnix := expiresAt.Unix()

	// Save to DynamoDB with TTL
	_, err = ddbClient.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(tableName),
		Item: map[string]ddbtypes.AttributeValue{
			"timer_key":  &ddbtypes.AttributeValueMemberS{Value: timerKey},
			"helper_id":  &ddbtypes.AttributeValueMemberS{Value: input.HelperID},
			"contact_id": &ddbtypes.AttributeValueMemberS{Value: input.ContactID},
			"expires_at": &ddbtypes.AttributeValueMemberN{Value: fmt.Sprintf("%d", expiresAtUnix)},
			"ttl":        &ddbtypes.AttributeValueMemberN{Value: fmt.Sprintf("%d", expiresAtUnix+86400)}, // TTL = expires_at + 24h
		},
	})
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to save timer: %w", err)
	}

	output.Logs = append(output.Logs, fmt.Sprintf("Evergreen mode: created new timer (expires in %.1f hours)", durationHours))
	return expiresAt, nil
}

// executeContactField reads expiry date from CRM custom field
func (h *CountdownTimer) executeContactField(ctx context.Context, input helpers.HelperInput, output *helpers.HelperOutput) (time.Time, error) {
	timerField := getStringConfigValue(input.Config, "timer_field", "")

	// Get field value from contact
	fieldValue, err := input.Connector.GetContactFieldValue(ctx, input.ContactID, timerField)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to get timer field '%s': %w", timerField, err)
	}

	if fieldValue == nil {
		return time.Time{}, fmt.Errorf("timer field '%s' is empty", timerField)
	}

	// Parse as ISO 8601 date
	fieldValueStr := fmt.Sprintf("%v", fieldValue)
	expiresAt, err := time.Parse(time.RFC3339, fieldValueStr)
	if err != nil {
		// Try alternative formats
		expiresAt, err = time.Parse("2006-01-02T15:04:05", fieldValueStr)
		if err != nil {
			expiresAt, err = time.Parse("2006-01-02", fieldValueStr)
			if err != nil {
				return time.Time{}, fmt.Errorf("invalid date format in field '%s' (expected ISO 8601): %s", timerField, fieldValueStr)
			}
		}
	}

	output.Logs = append(output.Logs, fmt.Sprintf("Contact field mode: read expiry from field '%s' = %s", timerField, expiresAt.Format(time.RFC3339)))
	return expiresAt, nil
}

// generateTimerURL creates a JWT-signed timer URL
func (h *CountdownTimer) generateTimerURL(baseURL, secret, contactID, helperID string, expiresAt time.Time) (string, error) {
	// Create JWT claims
	claims := jwt.MapClaims{
		"contact_id": contactID,
		"helper_id":  helperID,
		"expires_at": expiresAt.Unix(),
		"iat":        time.Now().Unix(),
	}

	// Sign with HS256
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", fmt.Errorf("failed to sign JWT: %w", err)
	}

	// Build URL
	timerURL := fmt.Sprintf("%s?token=%s", baseURL, tokenString)
	return timerURL, nil
}

// generateTimerKey creates a unique key for evergreen timers
func (h *CountdownTimer) generateTimerKey(helperID, contactID string) string {
	data := fmt.Sprintf("%s:%s", helperID, contactID)
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// parseUnixTimestamp safely parses a string to int64
func parseUnixTimestamp(s string) int64 {
	var ts int64
	fmt.Sscanf(s, "%d", &ts)
	return ts
}
