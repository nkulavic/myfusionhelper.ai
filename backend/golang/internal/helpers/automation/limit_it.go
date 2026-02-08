package automation

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	ddbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/myfusionhelper/api/internal/helpers"
)

func init() {
	helpers.Register("limit_it", func() helpers.Helper { return &LimitIt{} })
}

// LimitIt enforces rate limits on helper executions per contact
type LimitIt struct{}

func (h *LimitIt) GetName() string     { return "Limit It" }
func (h *LimitIt) GetType() string     { return "limit_it" }
func (h *LimitIt) GetCategory() string { return "automation" }
func (h *LimitIt) GetDescription() string {
	return "Enforce rate limits on helper executions per contact (e.g., max 1 execution per day)"
}
func (h *LimitIt) RequiresCRM() bool       { return false }
func (h *LimitIt) SupportedCRMs() []string { return nil }

func (h *LimitIt) GetConfigSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"limit_type": map[string]interface{}{
				"type":        "string",
				"enum":        []string{"per_hour", "per_day", "per_week", "per_month"},
				"description": "Time window for the limit",
				"default":     "per_day",
			},
			"max_executions": map[string]interface{}{
				"type":        "number",
				"description": "Maximum number of executions allowed in the time window",
				"default":     1,
			},
			"scope": map[string]interface{}{
				"type":        "string",
				"enum":        []string{"contact", "helper", "account"},
				"description": "Scope of the limit (per contact, per helper, or per account)",
				"default":     "contact",
			},
			"action_on_limit": map[string]interface{}{
				"type":        "string",
				"enum":        []string{"skip", "error"},
				"description": "What to do when limit is reached: skip (success=true but no action) or error (success=false)",
				"default":     "skip",
			},
		},
		"required": []string{},
	}
}

func (h *LimitIt) ValidateConfig(config map[string]interface{}) error {
	// All fields are optional with defaults
	return nil
}

func (h *LimitIt) Execute(ctx context.Context, input helpers.HelperInput) (*helpers.HelperOutput, error) {
	limitType, _ := input.Config["limit_type"].(string)
	if limitType == "" {
		limitType = "per_day"
	}

	maxExecutions := 1
	if me, ok := input.Config["max_executions"].(float64); ok && me > 0 {
		maxExecutions = int(me)
	}

	scope, _ := input.Config["scope"].(string)
	if scope == "" {
		scope = "contact"
	}

	actionOnLimit, _ := input.Config["action_on_limit"].(string)
	if actionOnLimit == "" {
		actionOnLimit = "skip"
	}

	output := &helpers.HelperOutput{
		Logs: make([]string, 0),
	}

	// Build rate limit key based on scope
	var rateLimitKey string
	switch scope {
	case "contact":
		rateLimitKey = fmt.Sprintf("limit:%s:%s:%s", input.HelperID, input.ContactID, limitType)
	case "helper":
		rateLimitKey = fmt.Sprintf("limit:%s:%s", input.HelperID, limitType)
	case "account":
		rateLimitKey = fmt.Sprintf("limit:%s:%s:%s", input.AccountID, input.HelperID, limitType)
	default:
		rateLimitKey = fmt.Sprintf("limit:%s:%s:%s", input.HelperID, input.ContactID, limitType)
	}

	output.Logs = append(output.Logs, fmt.Sprintf("Checking rate limit: %s (max: %d)", rateLimitKey, maxExecutions))

	// Check current count in DynamoDB rate-limits table
	count, err := h.getRateLimitCount(ctx, rateLimitKey)
	if err != nil {
		output.Logs = append(output.Logs, fmt.Sprintf("Warning: Failed to check rate limit: %v", err))
		// On error, allow execution (fail open)
		count = 0
	}

	output.Logs = append(output.Logs, fmt.Sprintf("Current execution count: %d / %d", count, maxExecutions))

	if count >= maxExecutions {
		// Limit reached
		output.Logs = append(output.Logs, fmt.Sprintf("Rate limit reached for %s (scope: %s)", rateLimitKey, scope))

		if actionOnLimit == "error" {
			output.Success = false
			output.Message = fmt.Sprintf("Rate limit reached: max %d executions %s", maxExecutions, limitType)
			return output, fmt.Errorf("rate limit exceeded")
		} else {
			// skip
			output.Success = true
			output.Message = fmt.Sprintf("Rate limit reached, skipping execution (max %d executions %s)", maxExecutions, limitType)
			output.Actions = []helpers.HelperAction{
				{
					Type:   "rate_limit_reached",
					Target: rateLimitKey,
					Value:  count,
				},
			}
			output.ModifiedData = map[string]interface{}{
				"limit_reached": true,
				"current_count": count,
				"max_allowed":   maxExecutions,
				"limit_type":    limitType,
				"scope":         scope,
			}
			return output, nil
		}
	}

	// Increment counter
	newCount, err := h.incrementRateLimitCount(ctx, rateLimitKey, limitType)
	if err != nil {
		output.Logs = append(output.Logs, fmt.Sprintf("Warning: Failed to increment rate limit counter: %v", err))
	} else {
		output.Logs = append(output.Logs, fmt.Sprintf("Incremented rate limit counter: %d", newCount))
	}

	output.Success = true
	output.Message = fmt.Sprintf("Rate limit check passed (%d / %d executions %s)", newCount, maxExecutions, limitType)
	output.Actions = []helpers.HelperAction{
		{
			Type:   "rate_limit_checked",
			Target: rateLimitKey,
			Value:  newCount,
		},
	}
	output.ModifiedData = map[string]interface{}{
		"limit_reached": false,
		"current_count": newCount,
		"max_allowed":   maxExecutions,
		"limit_type":    limitType,
		"scope":         scope,
	}

	return output, nil
}

// getRateLimitCount retrieves the current count from DynamoDB
func (h *LimitIt) getRateLimitCount(ctx context.Context, key string) (int, error) {
	tableName := os.Getenv("RATE_LIMITS_TABLE")
	if tableName == "" {
		return 0, fmt.Errorf("RATE_LIMITS_TABLE environment variable not set")
	}

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to load AWS config: %w", err)
	}

	client := dynamodb.NewFromConfig(cfg)

	result, err := client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(tableName),
		Key: map[string]ddbtypes.AttributeValue{
			"key": &ddbtypes.AttributeValueMemberS{Value: key},
		},
	})
	if err != nil {
		return 0, fmt.Errorf("failed to get rate limit from DynamoDB: %w", err)
	}

	if result.Item == nil {
		return 0, nil // No record yet, count is 0
	}

	var record struct {
		Key   string `dynamodbav:"key"`
		Count int    `dynamodbav:"count"`
	}
	if err := attributevalue.UnmarshalMap(result.Item, &record); err != nil {
		return 0, fmt.Errorf("failed to unmarshal rate limit record: %w", err)
	}

	return record.Count, nil
}

// incrementRateLimitCount atomically increments the counter in DynamoDB
func (h *LimitIt) incrementRateLimitCount(ctx context.Context, key string, limitType string) (int, error) {
	tableName := os.Getenv("RATE_LIMITS_TABLE")
	if tableName == "" {
		return 0, fmt.Errorf("RATE_LIMITS_TABLE environment variable not set")
	}

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to load AWS config: %w", err)
	}

	client := dynamodb.NewFromConfig(cfg)

	// Calculate TTL based on limit type
	var ttl int64
	now := time.Now()
	switch limitType {
	case "per_hour":
		ttl = now.Add(1 * time.Hour).Unix()
	case "per_day":
		ttl = now.Add(24 * time.Hour).Unix()
	case "per_week":
		ttl = now.Add(7 * 24 * time.Hour).Unix()
	case "per_month":
		ttl = now.Add(30 * 24 * time.Hour).Unix()
	default:
		ttl = now.Add(24 * time.Hour).Unix()
	}

	// Atomic increment with UpdateItem
	result, err := client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(tableName),
		Key: map[string]ddbtypes.AttributeValue{
			"key": &ddbtypes.AttributeValueMemberS{Value: key},
		},
		UpdateExpression: aws.String("ADD #count :inc SET #ttl = :ttl"),
		ExpressionAttributeNames: map[string]string{
			"#count": "count",
			"#ttl":   "ttl",
		},
		ExpressionAttributeValues: map[string]ddbtypes.AttributeValue{
			":inc": &ddbtypes.AttributeValueMemberN{Value: "1"},
			":ttl": &ddbtypes.AttributeValueMemberN{Value: fmt.Sprintf("%d", ttl)},
		},
		ReturnValues: ddbtypes.ReturnValueAllNew,
	})
	if err != nil {
		return 0, fmt.Errorf("failed to increment rate limit: %w", err)
	}

	var record struct {
		Count int `dynamodbav:"count"`
	}
	if err := attributevalue.UnmarshalMap(result.Attributes, &record); err != nil {
		return 0, fmt.Errorf("failed to unmarshal updated record: %w", err)
	}

	return record.Count, nil
}
