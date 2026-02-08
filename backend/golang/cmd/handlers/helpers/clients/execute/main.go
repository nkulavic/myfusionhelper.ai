package execute

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	ddbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"
	authMiddleware "github.com/myfusionhelper/api/internal/middleware/auth"
	"github.com/myfusionhelper/api/internal/ratelimit"
	apitypes "github.com/myfusionhelper/api/internal/types"
)

var (
	executionsTable = os.Getenv("EXECUTIONS_TABLE")
	accountsTable   = os.Getenv("ACCOUNTS_TABLE")
	rateLimitsTable = os.Getenv("RATE_LIMITS_TABLE")
)

// Handle processes API-key-authenticated execute requests.
// Auth context comes from the Lambda authorizer (not Cognito JWT).
func Handle(ctx context.Context, event events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	// Extract auth context from Lambda authorizer
	lambdaCtx := event.RequestContext.Authorizer.Lambda
	if lambdaCtx == nil {
		return authMiddleware.CreateErrorResponse(401, "Unauthorized"), nil
	}

	accountID, _ := lambdaCtx["accountId"].(string)
	apiKeyID, _ := lambdaCtx["apiKeyId"].(string)
	helperID, _ := lambdaCtx["helperId"].(string)

	if accountID == "" || helperID == "" {
		return authMiddleware.CreateErrorResponse(401, "Unauthorized"), nil
	}

	log.Printf("API key execute: helper=%s account=%s apiKey=%s", helperID, accountID, apiKeyID)

	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(os.Getenv("COGNITO_REGION")))
	if err != nil {
		return authMiddleware.CreateErrorResponse(500, "Internal server error"), nil
	}
	db := dynamodb.NewFromConfig(cfg)

	// Rate limit checks before creating execution
	limiter := ratelimit.New(db, accountsTable, rateLimitsTable)

	// 1. Get account to check plan limits
	accountResult, err := db.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(accountsTable),
		Key: map[string]ddbtypes.AttributeValue{
			"account_id": &ddbtypes.AttributeValueMemberS{Value: accountID},
		},
	})
	if err != nil || accountResult.Item == nil {
		return authMiddleware.CreateErrorResponse(500, "Failed to load account"), nil
	}

	var account apitypes.Account
	if err := attributevalue.UnmarshalMap(accountResult.Item, &account); err != nil {
		return authMiddleware.CreateErrorResponse(500, "Internal server error"), nil
	}

	// 2. Check monthly execution limit
	monthlyResult, err := limiter.CheckMonthlyLimit(ctx, accountID, account.Settings.MaxExecutions)
	if err != nil {
		log.Printf("Failed to check monthly limit: %v", err)
		// Don't block on rate limit errors — allow execution
	} else if !monthlyResult.Allowed {
		return createRateLimitResponse(monthlyResult, account.Plan), nil
	}

	// 3. Check per-helper burst limit
	burstResult, err := limiter.CheckBurstLimit(ctx, helperID, 100)
	if err != nil {
		log.Printf("Failed to check burst limit: %v", err)
	} else if !burstResult.Allowed {
		return createRateLimitResponse(burstResult, account.Plan), nil
	}

	// Parse input — POST body first, then query string fallback
	var body map[string]interface{}
	if event.Body != "" {
		_ = json.Unmarshal([]byte(event.Body), &body)
	}

	contactID := ""
	connectionID := ""
	var input map[string]interface{}

	if body != nil {
		if v, ok := body["contact_id"].(string); ok {
			contactID = v
		}
		if v, ok := body["connection_id"].(string); ok {
			connectionID = v
		}
		if v, ok := body["input"].(map[string]interface{}); ok {
			input = v
		}
	}

	// Query string fallback
	if contactID == "" {
		contactID = event.QueryStringParameters["contact_id"]
	}
	if connectionID == "" {
		connectionID = event.QueryStringParameters["connection_id"]
	}

	// Create execution record with status="queued".
	// DynamoDB Streams auto-dispatches to SQS FIFO.
	now := time.Now().UTC()
	executionID := "exec:" + uuid.Must(uuid.NewV7()).String()
	ttl := now.Add(7 * 24 * time.Hour).Unix()

	execution := map[string]interface{}{
		"execution_id":  executionID,
		"helper_id":     helperID,
		"account_id":    accountID,
		"api_key_id":    apiKeyID,
		"connection_id": connectionID,
		"contact_id":    contactID,
		"status":        "queued",
		"trigger_type":  "api",
		"input":         input,
		"created_at":    now.Format(time.RFC3339),
		"started_at":    now.Format(time.RFC3339),
		"ttl":           ttl,
	}

	item, err := attributevalue.MarshalMap(execution)
	if err != nil {
		return authMiddleware.CreateErrorResponse(500, "Failed to create execution"), nil
	}

	_, err = db.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(executionsTable),
		Item:      item,
	})
	if err != nil {
		log.Printf("Failed to store execution: %v", err)
		return authMiddleware.CreateErrorResponse(500, "Failed to create execution"), nil
	}

	return authMiddleware.CreateSuccessResponse(202, "Helper execution queued", map[string]interface{}{
		"execution_id": executionID,
		"helper_id":    helperID,
		"status":       "queued",
	}), nil
}

func createRateLimitResponse(result *ratelimit.Result, plan string) events.APIGatewayV2HTTPResponse {
	body := map[string]interface{}{
		"success": false,
		"error":   "Rate limit exceeded",
		"data": map[string]interface{}{
			"limit":       result.Limit,
			"used":        result.Used,
			"reset_at":    result.ResetAt,
			"plan":        plan,
			"upgrade_url": "https://app.myfusionhelper.ai/billing",
		},
	}
	bodyJSON, _ := json.Marshal(body)

	return events.APIGatewayV2HTTPResponse{
		StatusCode: 429,
		Headers: map[string]string{
			"Content-Type":                 "application/json",
			"Access-Control-Allow-Origin":  "*",
			"Access-Control-Allow-Methods": "GET, POST, PUT, DELETE, OPTIONS",
			"Access-Control-Allow-Headers": "Content-Type, Authorization, X-API-Key",
		},
		Body: string(bodyJSON),
	}
}
