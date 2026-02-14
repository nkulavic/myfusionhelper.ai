package automation

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	ddbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/golang-jwt/jwt/v5"
	"github.com/myfusionhelper/api/internal/helpers"
)

// NewCountdownTimer creates a new CountdownTimer helper instance
func NewCountdownTimer() helpers.Helper { return &CountdownTimer{} }

func init() {
	helpers.Register("countdown_timer", func() helpers.Helper { return &CountdownTimer{} })
}

// CountdownTimer generates personalized countdown timer URLs
type CountdownTimer struct{}

func (h *CountdownTimer) GetName() string     { return "Countdown Timer" }
func (h *CountdownTimer) GetType() string     { return "countdown_timer" }
func (h *CountdownTimer) GetCategory() string { return "automation" }
func (h *CountdownTimer) GetDescription() string {
	return "Generate personalized countdown timer URLs with styling options"
}
func (h *CountdownTimer) RequiresCRM() bool      { return false }
func (h *CountdownTimer) SupportedCRMs() []string { return nil }

func (h *CountdownTimer) GetConfigSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"timerType": map[string]interface{}{
				"type":        "string",
				"enum":        []string{"standard", "contact_field", "evergreen"},
				"description": "Timer mode: standard (fixed date), contact_field (date from contact field), evergreen (per-contact persistent)",
			},
			"endTime": map[string]interface{}{
				"type":        "string",
				"description": "Fixed deadline (ISO 8601 datetime) for standard mode",
			},
			"contactField": map[string]interface{}{
				"type":        "string",
				"description": "Contact field name containing deadline for contact_field mode",
			},
			"addDays": map[string]interface{}{
				"type":        "number",
				"description": "Days to add from view time for evergreen mode",
				"default":     0,
			},
			"addHours": map[string]interface{}{
				"type":        "number",
				"description": "Hours to add from view time for evergreen mode",
				"default":     0,
			},
			"addMinutes": map[string]interface{}{
				"type":        "number",
				"description": "Minutes to add from view time for evergreen mode",
				"default":     0,
			},
			"backgroundColor": map[string]interface{}{
				"type":        "string",
				"description": "Background color (hex)",
				"default":     "#000000",
			},
			"digitColor": map[string]interface{}{
				"type":        "string",
				"description": "Digit color (hex)",
				"default":     "#FFFFFF",
			},
			"labelColor": map[string]interface{}{
				"type":        "string",
				"description": "Label color (hex)",
				"default":     "#CCCCCC",
			},
			"transparentBg": map[string]interface{}{
				"type":        "boolean",
				"description": "Use transparent background",
				"default":     false,
			},
		},
		"required": []string{"timerType"},
	}
}

func (h *CountdownTimer) ValidateConfig(config map[string]interface{}) error {
	timerType, ok := config["timerType"].(string)
	if !ok || timerType == "" {
		return fmt.Errorf("timerType is required")
	}

	switch timerType {
	case "standard":
		endTime, ok := config["endTime"].(string)
		if !ok || endTime == "" {
			return fmt.Errorf("endTime is required for standard timer mode")
		}
		// Validate ISO 8601 datetime format
		if _, err := time.Parse(time.RFC3339, endTime); err != nil {
			return fmt.Errorf("endTime must be a valid ISO 8601 datetime: %w", err)
		}

	case "contact_field":
		contactField, ok := config["contactField"].(string)
		if !ok || contactField == "" {
			return fmt.Errorf("contactField is required for contact_field timer mode")
		}

	case "evergreen":
		// Validate that at least one duration is provided
		addDays, _ := config["addDays"].(float64)
		addHours, _ := config["addHours"].(float64)
		addMinutes, _ := config["addMinutes"].(float64)

		if addDays == 0 && addHours == 0 && addMinutes == 0 {
			return fmt.Errorf("evergreen timer mode requires at least one of addDays, addHours, or addMinutes to be greater than 0")
		}

	default:
		return fmt.Errorf("invalid timerType: %s (must be standard, contact_field, or evergreen)", timerType)
	}

	return nil
}

func (h *CountdownTimer) Execute(ctx context.Context, input helpers.HelperInput) (*helpers.HelperOutput, error) {
	timerType := input.Config["timerType"].(string)

	output := &helpers.HelperOutput{
		Logs: make([]string, 0),
	}

	var endTime time.Time
	var err error

	switch timerType {
	case "standard":
		endTimeStr := input.Config["endTime"].(string)
		endTime, err = time.Parse(time.RFC3339, endTimeStr)
		if err != nil {
			output.Message = fmt.Sprintf("Failed to parse endTime: %v", err)
			return output, err
		}
		output.Logs = append(output.Logs, fmt.Sprintf("Standard timer mode: deadline=%s", endTime.Format(time.RFC3339)))

	case "contact_field":
		contactField := input.Config["contactField"].(string)
		if input.Connector == nil {
			output.Message = "CRM connector required for contact_field timer mode"
			return output, fmt.Errorf("connector required")
		}

		fieldValue, err := input.Connector.GetContactFieldValue(ctx, input.ContactID, contactField)
		if err != nil {
			output.Message = fmt.Sprintf("Failed to read contact field '%s': %v", contactField, err)
			return output, err
		}

		// Parse field value as datetime
		fieldValueStr, ok := fieldValue.(string)
		if !ok {
			output.Message = fmt.Sprintf("Contact field '%s' is not a string", contactField)
			return output, fmt.Errorf("field value not a string")
		}

		endTime, err = time.Parse(time.RFC3339, fieldValueStr)
		if err != nil {
			output.Message = fmt.Sprintf("Failed to parse field '%s' value as datetime: %v", contactField, err)
			return output, err
		}
		output.Logs = append(output.Logs, fmt.Sprintf("Contact field timer mode: field=%s, deadline=%s", contactField, endTime.Format(time.RFC3339)))

	case "evergreen":
		addDays, _ := input.Config["addDays"].(float64)
		addHours, _ := input.Config["addHours"].(float64)
		addMinutes, _ := input.Config["addMinutes"].(float64)

		// Check for existing timer in DynamoDB
		timerKey := fmt.Sprintf("%s:%s:%s", input.AccountID, input.HelperID, input.ContactID)
		existingTimer, err := h.getEvergreenTimer(ctx, timerKey)
		if err == nil && existingTimer != nil {
			endTime = existingTimer.ExpiresAt
			output.Logs = append(output.Logs, fmt.Sprintf("Evergreen timer mode: existing timer found, expires_at=%s", endTime.Format(time.RFC3339)))
		} else {
			// Create new timer
			duration := time.Duration(addDays*24)*time.Hour +
				time.Duration(addHours)*time.Hour +
				time.Duration(addMinutes)*time.Minute
			endTime = time.Now().Add(duration)

			// Save to DynamoDB
			if err := h.saveEvergreenTimer(ctx, timerKey, endTime); err != nil {
				output.Logs = append(output.Logs, fmt.Sprintf("Warning: Failed to save evergreen timer to DynamoDB: %v", err))
			} else {
				output.Logs = append(output.Logs, fmt.Sprintf("Evergreen timer mode: created new timer, duration=%.0fd%.0fh%.0fm, expires_at=%s", addDays, addHours, addMinutes, endTime.Format(time.RFC3339)))
			}
		}
	}

	// Generate signed JWT token
	timerURL, err := h.generateTimerURL(input, endTime)
	if err != nil {
		output.Message = fmt.Sprintf("Failed to generate timer URL: %v", err)
		return output, err
	}

	output.Success = true
	output.Message = fmt.Sprintf("Generated countdown timer URL (expires: %s)", endTime.Format(time.RFC3339))
	output.ModifiedData = map[string]interface{}{
		"timer_url":    timerURL,
		"expires_at":   endTime.Format(time.RFC3339),
		"timer_type":   timerType,
		"generated_at": time.Now().Format(time.RFC3339),
	}
	output.Actions = append(output.Actions, helpers.HelperAction{
		Type:   "timer_generated",
		Target: input.ContactID,
		Value:  timerURL,
	})

	return output, nil
}

// generateTimerURL creates a signed JWT-based timer URL
func (h *CountdownTimer) generateTimerURL(input helpers.HelperInput, expiresAt time.Time) (string, error) {
	// Generate JWT signing secret (use environment variable or generate random)
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		// Generate random secret for this execution (not ideal for production, but safe fallback)
		secretBytes := make([]byte, 32)
		if _, err := rand.Read(secretBytes); err != nil {
			return "", fmt.Errorf("failed to generate JWT secret: %w", err)
		}
		jwtSecret = base64.StdEncoding.EncodeToString(secretBytes)
	}

	// Extract styling config
	backgroundColor, _ := input.Config["backgroundColor"].(string)
	if backgroundColor == "" {
		backgroundColor = "#000000"
	}
	digitColor, _ := input.Config["digitColor"].(string)
	if digitColor == "" {
		digitColor = "#FFFFFF"
	}
	labelColor, _ := input.Config["labelColor"].(string)
	if labelColor == "" {
		labelColor = "#CCCCCC"
	}
	transparentBg, _ := input.Config["transparentBg"].(bool)

	// Create JWT claims
	claims := jwt.MapClaims{
		"contact_id":       input.ContactID,
		"account_id":       input.AccountID,
		"helper_id":        input.HelperID,
		"expires_at":       expiresAt.Unix(),
		"bg_color":         backgroundColor,
		"digit_color":      digitColor,
		"label_color":      labelColor,
		"transparent_bg":   transparentBg,
		"iat":              time.Now().Unix(),
		"exp":              expiresAt.Add(7 * 24 * time.Hour).Unix(), // JWT valid for 7 days past timer expiry
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		return "", fmt.Errorf("failed to sign JWT: %w", err)
	}

	// Construct timer URL (placeholder for actual timer rendering endpoint)
	baseURL := os.Getenv("APP_URL")
	if baseURL == "" {
		baseURL = "https://app.myfusionhelper.ai"
	}

	timerURL := fmt.Sprintf("%s/timer?t=%s", baseURL, tokenString)
	return timerURL, nil
}

// EvergreenTimer represents a persisted timer in DynamoDB
type EvergreenTimer struct {
	TimerKey  string    `dynamodbav:"timer_key"`
	ExpiresAt time.Time `dynamodbav:"expires_at"`
}

// getEvergreenTimer retrieves an existing evergreen timer from DynamoDB
func (h *CountdownTimer) getEvergreenTimer(ctx context.Context, timerKey string) (*EvergreenTimer, error) {
	tableName := os.Getenv("COUNTDOWN_TIMERS_TABLE")
	if tableName == "" {
		return nil, fmt.Errorf("COUNTDOWN_TIMERS_TABLE environment variable not set")
	}

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	client := dynamodb.NewFromConfig(cfg)

	result, err := client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(tableName),
		Key: map[string]ddbtypes.AttributeValue{
			"timer_key": &ddbtypes.AttributeValueMemberS{Value: timerKey},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get timer from DynamoDB: %w", err)
	}

	if result.Item == nil {
		return nil, nil // Timer not found
	}

	var timer EvergreenTimer
	if err := attributevalue.UnmarshalMap(result.Item, &timer); err != nil {
		return nil, fmt.Errorf("failed to unmarshal timer: %w", err)
	}

	return &timer, nil
}

// saveEvergreenTimer persists a new evergreen timer to DynamoDB
func (h *CountdownTimer) saveEvergreenTimer(ctx context.Context, timerKey string, expiresAt time.Time) error {
	tableName := os.Getenv("COUNTDOWN_TIMERS_TABLE")
	if tableName == "" {
		return fmt.Errorf("COUNTDOWN_TIMERS_TABLE environment variable not set")
	}

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return fmt.Errorf("failed to load AWS config: %w", err)
	}

	client := dynamodb.NewFromConfig(cfg)

	timer := EvergreenTimer{
		TimerKey:  timerKey,
		ExpiresAt: expiresAt,
	}

	item, err := attributevalue.MarshalMap(timer)
	if err != nil {
		return fmt.Errorf("failed to marshal timer: %w", err)
	}

	_, err = client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(tableName),
		Item:      item,
	})
	if err != nil {
		return fmt.Errorf("failed to put timer to DynamoDB: %w", err)
	}

	return nil
}
