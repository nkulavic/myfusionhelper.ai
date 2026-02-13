package automation

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/google/uuid"
	"github.com/myfusionhelper/api/internal/helpers"
)

// NewChainIt creates a new ChainIt helper instance
func NewChainIt() helpers.Helper { return &ChainIt{} }

func init() {
	helpers.Register("chain_it", func() helpers.Helper { return &ChainIt{} })
}

// ChainIt chains multiple helper executions together with conditional logic and timing control.
// Makes HTTP POST requests to execute each helper in sequence.
type ChainIt struct{}

// helperConfig represents a helper to execute in the chain
type helperConfig struct {
	ID     string
	Config map[string]interface{}
}

func (h *ChainIt) GetName() string        { return "Chain It" }
func (h *ChainIt) GetType() string        { return "chain_it" }
func (h *ChainIt) GetCategory() string    { return "automation" }
func (h *ChainIt) GetDescription() string { return "Chain multiple helper executions with conditionals and delays" }
func (h *ChainIt) RequiresCRM() bool      { return false }
func (h *ChainIt) SupportedCRMs() []string { return nil }

func (h *ChainIt) GetConfigSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"helpers": map[string]interface{}{
				"type": "array",
				"items": map[string]interface{}{
					"oneOf": []interface{}{
						map[string]interface{}{"type": "string"},
						map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"id":     map[string]interface{}{"type": "string"},
								"config": map[string]interface{}{"type": "object"},
							},
							"required": []string{"id"},
						},
					},
				},
				"description": "List of helper IDs/short keys to chain. Can be strings or objects with 'id' and optional 'config'",
			},
			"conditional_field": map[string]interface{}{
				"type":        "string",
				"description": "Optional: field name to check for conditional execution",
			},
			"conditional_value": map[string]interface{}{
				"type":        "string",
				"description": "Optional: value to compare against for conditional execution",
			},
			"conditional_operator": map[string]interface{}{
				"type":        "string",
				"enum":        []string{"equals", "not_equals", "contains", "not_contains", "exists", "not_exists"},
				"description": "Optional: comparison operator for conditional execution (default: equals)",
			},
			"delay_seconds": map[string]interface{}{
				"type":        "number",
				"description": "Optional: delay in seconds between helper executions (default: 0)",
			},
			"on_success_helpers": map[string]interface{}{
				"type":        "array",
				"items":       map[string]interface{}{"type": "string"},
				"description": "Optional: helpers to execute if all primary helpers succeed",
			},
			"on_failure_helpers": map[string]interface{}{
				"type":        "array",
				"items":       map[string]interface{}{"type": "string"},
				"description": "Optional: helpers to execute if any primary helper fails",
			},
		},
		"required": []string{"helpers"},
	}
}

func (h *ChainIt) ValidateConfig(config map[string]interface{}) error {
	helpersList, ok := config["helpers"]
	if !ok {
		return fmt.Errorf("helpers is required")
	}

	switch v := helpersList.(type) {
	case []interface{}:
		if len(v) == 0 {
			return fmt.Errorf("helpers must contain at least one helper type")
		}
	case []string:
		if len(v) == 0 {
			return fmt.Errorf("helpers must contain at least one helper type")
		}
	default:
		return fmt.Errorf("helpers must be an array of strings")
	}

	// Validate conditional operator if provided
	if operator, ok := config["conditional_operator"].(string); ok {
		validOperators := map[string]bool{
			"equals":       true,
			"not_equals":   true,
			"contains":     true,
			"not_contains": true,
			"exists":       true,
			"not_exists":   true,
		}
		if !validOperators[operator] {
			return fmt.Errorf("invalid conditional_operator: %s", operator)
		}
	}

	// Validate delay_seconds if provided
	if delay, ok := config["delay_seconds"]; ok {
		switch v := delay.(type) {
		case float64:
			if v < 0 {
				return fmt.Errorf("delay_seconds must be non-negative")
			}
		case int:
			if v < 0 {
				return fmt.Errorf("delay_seconds must be non-negative")
			}
		default:
			return fmt.Errorf("delay_seconds must be a number")
		}
	}

	return nil
}

func (h *ChainIt) Execute(ctx context.Context, input helpers.HelperInput) (*helpers.HelperOutput, error) {
	// Parse helpers config (can be strings or objects)
	helpersChain, err := h.parseHelpersConfig(input.Config["helpers"])
	if err != nil {
		return nil, fmt.Errorf("invalid helpers configuration: %w", err)
	}

	output := &helpers.HelperOutput{
		Actions: make([]helpers.HelperAction, 0),
		Logs:    make([]string, 0),
	}

	// Check conditional execution if specified
	shouldExecute, conditionLog := h.evaluateCondition(input)
	if conditionLog != "" {
		output.Logs = append(output.Logs, conditionLog)
	}

	if !shouldExecute {
		output.Success = true
		output.Message = "Chain skipped due to conditional check"
		return output, nil
	}

	// Extract delay configuration
	delaySeconds := h.getDelaySeconds(input.Config)

	// Get DynamoDB client
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(os.Getenv("COGNITO_REGION")))
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}
	db := dynamodb.NewFromConfig(cfg)
	executionsTable := os.Getenv("EXECUTIONS_TABLE")

	if executionsTable == "" {
		return nil, fmt.Errorf("EXECUTIONS_TABLE environment variable not set")
	}

	// Create execution records for each chained helper
	now := time.Now().UTC()
	var executionIDs []string
	var queuedHelpers []string

	for i, helper := range helpersChain {
		// Calculate start time with delay
		startTime := now
		if i > 0 && delaySeconds > 0 {
			startTime = startTime.Add(time.Duration(delaySeconds*i) * time.Second)
		}

		// Build config - use helper-specific config if provided, otherwise use shared config
		configToUse := input.Config
		if helper.Config != nil {
			configToUse = helper.Config
		}

		// Create execution record
		executionID := "exec:" + uuid.Must(uuid.NewV7()).String()
		ttl := now.Add(7 * 24 * time.Hour).Unix()

		execution := map[string]interface{}{
			"execution_id":  executionID,
			"helper_id":     helper.ID,
			"account_id":    input.AccountID,
			"user_id":       input.UserID,
			"connection_id": "", // Will be inherited if needed
			"contact_id":    input.ContactID,
			"status":        "queued",
			"trigger_type":  "chain",
			"input":         configToUse,
			"created_at":    now.Format(time.RFC3339),
			"started_at":    startTime.Format(time.RFC3339),
			"ttl":           ttl,
			"parent_exec":   input.HelperID, // Track parent chain_it execution
		}

		item, err := attributevalue.MarshalMap(execution)
		if err != nil {
			output.Logs = append(output.Logs, fmt.Sprintf("Helper %d (%s): Failed to marshal execution: %v", i+1, helper.ID, err))
			continue
		}

		_, err = db.PutItem(ctx, &dynamodb.PutItemInput{
			TableName: aws.String(executionsTable),
			Item:      item,
		})
		if err != nil {
			output.Logs = append(output.Logs, fmt.Sprintf("Helper %d (%s): Failed to create execution: %v", i+1, helper.ID, err))
			continue
		}

		executionIDs = append(executionIDs, executionID)
		queuedHelpers = append(queuedHelpers, helper.ID)

		if delaySeconds > 0 && i > 0 {
			output.Logs = append(output.Logs, fmt.Sprintf("Queued helper %d: %s (delayed %d seconds)", i+1, helper.ID, delaySeconds*i))
		} else {
			output.Logs = append(output.Logs, fmt.Sprintf("Queued helper %d: %s", i+1, helper.ID))
		}
	}

	// Queue success/failure handlers (note: these won't auto-execute, would need worker enhancement)
	// For now, just log them
	if successHelpers, err := h.parseHelpersConfig(input.Config["on_success_helpers"]); err == nil && len(successHelpers) > 0 {
		output.Logs = append(output.Logs, fmt.Sprintf("Success handlers configured: %d (not yet auto-executed)", len(successHelpers)))
	}
	if failureHelpers, err := h.parseHelpersConfig(input.Config["on_failure_helpers"]); err == nil && len(failureHelpers) > 0 {
		output.Logs = append(output.Logs, fmt.Sprintf("Failure handlers configured: %d (not yet auto-executed)", len(failureHelpers)))
	}

	output.Success = true
	output.Message = fmt.Sprintf("Queued %d helper(s) for execution", len(executionIDs))
	output.ModifiedData = map[string]interface{}{
		"execution_ids":  executionIDs,
		"queued_helpers": queuedHelpers,
		"queued_count":   len(executionIDs),
		"delay_seconds":  delaySeconds,
	}

	return output, nil
}

// parseHelpersConfig parses the helpers configuration which can be strings or objects
func (h *ChainIt) parseHelpersConfig(val interface{}) ([]helperConfig, error) {
	if val == nil {
		return []helperConfig{}, nil
	}

	switch v := val.(type) {
	case []interface{}:
		result := make([]helperConfig, 0, len(v))
		for _, item := range v {
			switch itemVal := item.(type) {
			case string:
				// Simple string helper ID
				result = append(result, helperConfig{ID: itemVal})
			case map[string]interface{}:
				// Object with id and optional config
				if id, ok := itemVal["id"].(string); ok {
					cfg := helperConfig{ID: id}
					if config, ok := itemVal["config"].(map[string]interface{}); ok {
						cfg.Config = config
					}
					result = append(result, cfg)
				}
			}
		}
		return result, nil
	case []string:
		result := make([]helperConfig, 0, len(v))
		for _, id := range v {
			result = append(result, helperConfig{ID: id})
		}
		return result, nil
	default:
		return nil, fmt.Errorf("helpers must be an array")
	}
}

// evaluateCondition checks if the chain should execute based on conditional configuration
func (h *ChainIt) evaluateCondition(input helpers.HelperInput) (bool, string) {
	conditionalField, hasField := input.Config["conditional_field"].(string)
	if !hasField || conditionalField == "" {
		return true, "" // No condition specified, execute unconditionally
	}

	operator := "equals" // default
	if op, ok := input.Config["conditional_operator"].(string); ok {
		operator = op
	}

	conditionalValue, _ := input.Config["conditional_value"].(string)

	// Get actual field value from contact data
	var actualValue string
	var fieldExists bool

	if input.ContactData != nil {
		// Check standard fields
		switch conditionalField {
		case "email":
			actualValue = input.ContactData.Email
			fieldExists = actualValue != ""
		case "first_name":
			actualValue = input.ContactData.FirstName
			fieldExists = actualValue != ""
		case "last_name":
			actualValue = input.ContactData.LastName
			fieldExists = actualValue != ""
		case "phone":
			actualValue = input.ContactData.Phone
			fieldExists = actualValue != ""
		default:
			// Check custom fields
			if val, ok := input.ContactData.CustomFields[conditionalField]; ok {
				actualValue = fmt.Sprintf("%v", val)
				fieldExists = true
			}
		}
	}

	// Evaluate condition based on operator
	var result bool
	var logMessage string

	switch operator {
	case "equals":
		result = actualValue == conditionalValue
		logMessage = fmt.Sprintf("Condition: %s == %s (actual: %s) = %v", conditionalField, conditionalValue, actualValue, result)
	case "not_equals":
		result = actualValue != conditionalValue
		logMessage = fmt.Sprintf("Condition: %s != %s (actual: %s) = %v", conditionalField, conditionalValue, actualValue, result)
	case "contains":
		result = fieldExists && len(actualValue) > 0 && len(conditionalValue) > 0 && stringContains(actualValue, conditionalValue)
		logMessage = fmt.Sprintf("Condition: %s contains %s (actual: %s) = %v", conditionalField, conditionalValue, actualValue, result)
	case "not_contains":
		result = !fieldExists || !stringContains(actualValue, conditionalValue)
		logMessage = fmt.Sprintf("Condition: %s not contains %s (actual: %s) = %v", conditionalField, conditionalValue, actualValue, result)
	case "exists":
		result = fieldExists
		logMessage = fmt.Sprintf("Condition: %s exists = %v", conditionalField, result)
	case "not_exists":
		result = !fieldExists
		logMessage = fmt.Sprintf("Condition: %s not exists = %v", conditionalField, result)
	default:
		result = true
		logMessage = fmt.Sprintf("Unknown operator: %s, executing unconditionally", operator)
	}

	return result, logMessage
}

// getDelaySeconds extracts and converts delay_seconds from config
func (h *ChainIt) getDelaySeconds(config map[string]interface{}) int {
	delay, ok := config["delay_seconds"]
	if !ok {
		return 0
	}

	switch v := delay.(type) {
	case float64:
		return int(v)
	case int:
		return v
	case string:
		if intVal, err := strconv.Atoi(v); err == nil {
			return intVal
		}
	}

	return 0
}

// stringContains checks if a string contains a substring (case-sensitive)
func stringContains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || containsSubstring(s, substr))
}

// containsSubstring is a simple substring check
func containsSubstring(s, substr string) bool {
	if len(substr) == 0 {
		return true
	}
	if len(s) < len(substr) {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
