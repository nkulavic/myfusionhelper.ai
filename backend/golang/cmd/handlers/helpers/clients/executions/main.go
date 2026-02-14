package executions

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	ddbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	authMiddleware "github.com/myfusionhelper/api/internal/middleware/auth"
	apitypes "github.com/myfusionhelper/api/internal/types"
)

var (
	executionsTable = os.Getenv("EXECUTIONS_TABLE")
)

// HandleWithAuth routes execution requests
func HandleWithAuth(ctx context.Context, event events.APIGatewayV2HTTPRequest, authCtx *apitypes.AuthContext) (events.APIGatewayV2HTTPResponse, error) {
	path := event.RequestContext.HTTP.Path
	method := event.RequestContext.HTTP.Method

	switch {
	case path == "/executions" && method == "GET":
		return listExecutions(ctx, event, authCtx)
	case strings.HasPrefix(path, "/executions/") && method == "GET":
		return getExecution(ctx, event, authCtx)
	default:
		return authMiddleware.CreateErrorResponse(404, "Not Found"), nil
	}
}

func listExecutions(ctx context.Context, event events.APIGatewayV2HTTPRequest, authCtx *apitypes.AuthContext) (events.APIGatewayV2HTTPResponse, error) {
	log.Printf("List executions for account: %s", authCtx.AccountID)

	helperID := event.QueryStringParameters["helper_id"]
	statusFilter := event.QueryStringParameters["status"]
	limitStr := event.QueryStringParameters["limit"]
	nextToken := event.QueryStringParameters["next_token"]

	// Parse limit (default 20, max 100)
	limit := int32(20)
	if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
		limit = int32(l)
	}

	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(os.Getenv("COGNITO_REGION")))
	if err != nil {
		return authMiddleware.CreateErrorResponse(500, "Internal server error"), nil
	}
	db := dynamodb.NewFromConfig(cfg)

	// Build query based on parameters
	var indexName string
	var keyExpr string
	exprValues := map[string]ddbtypes.AttributeValue{}
	exprNames := map[string]string{}
	var filterParts []string

	if helperID != "" {
		// Query by helper_id using HelperIdCreatedAtIndex
		indexName = "HelperIdCreatedAtIndex"
		keyExpr = "helper_id = :helper_id"
		exprValues[":helper_id"] = &ddbtypes.AttributeValueMemberS{Value: helperID}
		// Must also filter by account_id for security
		filterParts = append(filterParts, "account_id = :account_id")
		exprValues[":account_id"] = &ddbtypes.AttributeValueMemberS{Value: authCtx.AccountID}
	} else {
		// Query by account_id using AccountIdCreatedAtIndex
		indexName = "AccountIdCreatedAtIndex"
		keyExpr = "account_id = :account_id"
		exprValues[":account_id"] = &ddbtypes.AttributeValueMemberS{Value: authCtx.AccountID}
	}

	// Optional status filter
	if statusFilter != "" {
		filterParts = append(filterParts, "#s = :status_filter")
		exprNames["#s"] = "status"
		exprValues[":status_filter"] = &ddbtypes.AttributeValueMemberS{Value: statusFilter}
	}

	queryInput := &dynamodb.QueryInput{
		TableName:                 aws.String(executionsTable),
		IndexName:                 aws.String(indexName),
		KeyConditionExpression:    aws.String(keyExpr),
		ExpressionAttributeValues: exprValues,
		ScanIndexForward:          aws.Bool(false), // newest first
		Limit:                     aws.Int32(limit),
	}

	if len(filterParts) > 0 {
		queryInput.FilterExpression = aws.String(strings.Join(filterParts, " AND "))
	}
	if len(exprNames) > 0 {
		queryInput.ExpressionAttributeNames = exprNames
	}

	// Cursor-based pagination
	if nextToken != "" {
		startKey, err := decodePageToken(nextToken)
		if err == nil && startKey != nil {
			queryInput.ExclusiveStartKey = startKey
		}
	}

	result, err := db.Query(ctx, queryInput)
	if err != nil {
		log.Printf("Failed to query executions: %v", err)
		return authMiddleware.CreateErrorResponse(500, "Failed to list executions"), nil
	}

	// Unmarshal results
	executionItems := make([]map[string]interface{}, 0, len(result.Items))
	for _, item := range result.Items {
		var exec apitypes.Execution
		if err := attributevalue.UnmarshalMap(item, &exec); err != nil {
			continue
		}
		// Double-check account ownership when querying by helper_id
		if exec.AccountID != authCtx.AccountID {
			continue
		}
		executionItems = append(executionItems, map[string]interface{}{
			"execution_id":  exec.ExecutionID,
			"helper_id":     exec.HelperID,
			"account_id":    exec.AccountID,
			"user_id":       exec.UserID,
			"connection_id": exec.ConnectionID,
			"contact_id":    exec.ContactID,
			"status":        exec.Status,
			"trigger_type":  exec.TriggerType,
			"error_message": exec.ErrorMessage,
			"duration_ms":   exec.DurationMs,
			"created_at":    exec.CreatedAt,
			"started_at":    exec.StartedAt,
			"completed_at":  exec.CompletedAt,
		})
	}

	// Build next_token if there are more results
	var responseNextToken string
	hasMore := false
	if result.LastEvaluatedKey != nil {
		hasMore = true
		responseNextToken = encodePageToken(result.LastEvaluatedKey)
	}

	return authMiddleware.CreateSuccessResponse(200, "Executions retrieved successfully", map[string]interface{}{
		"executions": executionItems,
		"total_count": len(executionItems),
		"next_token":  responseNextToken,
		"has_more":    hasMore,
	}), nil
}

func getExecution(ctx context.Context, event events.APIGatewayV2HTTPRequest, authCtx *apitypes.AuthContext) (events.APIGatewayV2HTTPResponse, error) {
	// Extract execution_id from path: /executions/{execution_id}
	path := event.RequestContext.HTTP.Path
	parts := strings.Split(strings.TrimPrefix(path, "/executions/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		return authMiddleware.CreateErrorResponse(400, "Execution ID is required"), nil
	}
	executionID := parts[0]

	log.Printf("Get execution %s for account: %s", executionID, authCtx.AccountID)

	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(os.Getenv("COGNITO_REGION")))
	if err != nil {
		return authMiddleware.CreateErrorResponse(500, "Internal server error"), nil
	}
	db := dynamodb.NewFromConfig(cfg)

	result, err := db.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(executionsTable),
		Key: map[string]ddbtypes.AttributeValue{
			"execution_id": &ddbtypes.AttributeValueMemberS{Value: executionID},
		},
	})
	if err != nil || result.Item == nil {
		return authMiddleware.CreateErrorResponse(404, "Execution not found"), nil
	}

	var exec apitypes.Execution
	if err := attributevalue.UnmarshalMap(result.Item, &exec); err != nil {
		return authMiddleware.CreateErrorResponse(500, "Internal server error"), nil
	}

	// Verify account ownership
	if exec.AccountID != authCtx.AccountID {
		return authMiddleware.CreateErrorResponse(404, "Execution not found"), nil
	}

	return authMiddleware.CreateSuccessResponse(200, "Execution retrieved successfully", map[string]interface{}{
		"execution_id":  exec.ExecutionID,
		"helper_id":     exec.HelperID,
		"account_id":    exec.AccountID,
		"user_id":       exec.UserID,
		"connection_id": exec.ConnectionID,
		"contact_id":    exec.ContactID,
		"status":        exec.Status,
		"trigger_type":  exec.TriggerType,
		"input":         exec.Input,
		"output":        exec.Output,
		"error_message": exec.ErrorMessage,
		"duration_ms":   exec.DurationMs,
		"created_at":    exec.CreatedAt,
		"started_at":    exec.StartedAt,
		"completed_at":  exec.CompletedAt,
	}), nil
}

// encodePageToken encodes a DynamoDB LastEvaluatedKey as a base64 JSON string
func encodePageToken(key map[string]ddbtypes.AttributeValue) string {
	// Convert to a simpler format for encoding
	simpleKey := make(map[string]string)
	for k, v := range key {
		if sv, ok := v.(*ddbtypes.AttributeValueMemberS); ok {
			simpleKey[k] = sv.Value
		}
	}
	data, err := json.Marshal(simpleKey)
	if err != nil {
		return ""
	}
	return base64.URLEncoding.EncodeToString(data)
}

// decodePageToken decodes a base64 page token back to a DynamoDB ExclusiveStartKey
func decodePageToken(token string) (map[string]ddbtypes.AttributeValue, error) {
	data, err := base64.URLEncoding.DecodeString(token)
	if err != nil {
		return nil, err
	}
	var simpleKey map[string]string
	if err := json.Unmarshal(data, &simpleKey); err != nil {
		return nil, err
	}
	key := make(map[string]ddbtypes.AttributeValue)
	for k, v := range simpleKey {
		key[k] = &ddbtypes.AttributeValueMemberS{Value: v}
	}
	return key, nil
}
