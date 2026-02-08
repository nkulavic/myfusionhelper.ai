package crud

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	ddbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge"
	ebtypes "github.com/aws/aws-sdk-go-v2/service/eventbridge/types"
	lambdasvc "github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/google/uuid"
	helperEngine "github.com/myfusionhelper/api/internal/helpers"
	authMiddleware "github.com/myfusionhelper/api/internal/middleware/auth"
	"github.com/myfusionhelper/api/internal/nanoid"
	apitypes "github.com/myfusionhelper/api/internal/types"
)

var (
	helpersTable        = os.Getenv("HELPERS_TABLE")
	executionsTable     = os.Getenv("EXECUTIONS_TABLE")
	schedulerFunctionARN = os.Getenv("SCHEDULER_FUNCTION_ARN")
)

type CreateHelperRequest struct {
	Name         string                 `json:"name"`
	Description  string                 `json:"description"`
	HelperType   string                 `json:"helper_type"`
	Category     string                 `json:"category"`
	ConnectionID string                 `json:"connection_id"`
	Config       map[string]interface{} `json:"config"`
}

type UpdateHelperRequest struct {
	Name            string                 `json:"name"`
	Description     string                 `json:"description"`
	Config          map[string]interface{} `json:"config"`
	Enabled         *bool                  `json:"enabled"`
	ConnectionID    string                 `json:"connection_id"`
	ScheduleEnabled *bool                  `json:"schedule_enabled"`
	CronExpression  string                 `json:"cron_expression"`
}

type ExecuteHelperRequest struct {
	ContactID string                 `json:"contact_id"`
	Input     map[string]interface{} `json:"input"`
}

// HandleWithAuth routes to the appropriate operation based on path and method
func HandleWithAuth(ctx context.Context, event events.APIGatewayV2HTTPRequest, authCtx *apitypes.AuthContext) (events.APIGatewayV2HTTPResponse, error) {
	path := event.RequestContext.HTTP.Path
	method := event.RequestContext.HTTP.Method

	switch {
	case path == "/helpers" && method == "GET":
		return listHelpers(ctx, event, authCtx)
	case path == "/helpers" && method == "POST":
		return createHelper(ctx, event, authCtx)
	case strings.HasSuffix(path, "/execute") && method == "POST":
		return executeHelper(ctx, event, authCtx)
	case strings.HasPrefix(path, "/helpers/") && method == "GET":
		return getHelper(ctx, event, authCtx)
	case strings.HasPrefix(path, "/helpers/") && method == "PUT":
		return updateHelper(ctx, event, authCtx)
	case strings.HasPrefix(path, "/helpers/") && method == "DELETE":
		return deleteHelper(ctx, event, authCtx)
	default:
		return authMiddleware.CreateErrorResponse(404, "Not Found"), nil
	}
}

func listHelpers(ctx context.Context, event events.APIGatewayV2HTTPRequest, authCtx *apitypes.AuthContext) (events.APIGatewayV2HTTPResponse, error) {
	log.Printf("List helpers for account: %s", authCtx.AccountID)

	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(os.Getenv("COGNITO_REGION")))
	if err != nil {
		return authMiddleware.CreateErrorResponse(500, "Internal server error"), nil
	}
	db := dynamodb.NewFromConfig(cfg)

	result, err := db.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(helpersTable),
		IndexName:              aws.String("AccountIdIndex"),
		KeyConditionExpression: aws.String("account_id = :account_id"),
		FilterExpression:       aws.String("#s <> :deleted_status"),
		ExpressionAttributeNames: map[string]string{
			"#s": "status",
		},
		ExpressionAttributeValues: map[string]ddbtypes.AttributeValue{
			":account_id":      &ddbtypes.AttributeValueMemberS{Value: authCtx.AccountID},
			":deleted_status":  &ddbtypes.AttributeValueMemberS{Value: "deleted"},
		},
	})
	if err != nil {
		log.Printf("Failed to query helpers: %v", err)
		return authMiddleware.CreateErrorResponse(500, "Failed to list helpers"), nil
	}

	var helpers []map[string]interface{}
	for _, item := range result.Items {
		var helper apitypes.Helper
		if err := attributevalue.UnmarshalMap(item, &helper); err != nil {
			continue
		}
		helpers = append(helpers, map[string]interface{}{
			"helper_id":        helper.HelperID,
			"short_key":        helper.ShortKey,
			"name":             helper.Name,
			"description":      helper.Description,
			"helper_type":      helper.HelperType,
			"category":         helper.Category,
			"status":           helper.Status,
			"enabled":          helper.Enabled,
			"execution_count":  helper.ExecutionCount,
			"last_executed_at": helper.LastExecutedAt,
			"created_at":       helper.CreatedAt,
			"updated_at":       helper.UpdatedAt,
		})
	}

	return authMiddleware.CreateSuccessResponse(200, "Helpers retrieved successfully", map[string]interface{}{
		"helpers":     helpers,
		"total_count": len(helpers),
	}), nil
}

func getHelper(ctx context.Context, event events.APIGatewayV2HTTPRequest, authCtx *apitypes.AuthContext) (events.APIGatewayV2HTTPResponse, error) {
	helperID := event.PathParameters["helper_id"]
	if helperID == "" {
		return authMiddleware.CreateErrorResponse(400, "Helper ID is required"), nil
	}

	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(os.Getenv("COGNITO_REGION")))
	if err != nil {
		return authMiddleware.CreateErrorResponse(500, "Internal server error"), nil
	}
	db := dynamodb.NewFromConfig(cfg)

	result, err := db.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(helpersTable),
		Key: map[string]ddbtypes.AttributeValue{
			"helper_id": &ddbtypes.AttributeValueMemberS{Value: helperID},
		},
	})
	if err != nil || result.Item == nil {
		return authMiddleware.CreateErrorResponse(404, "Helper not found"), nil
	}

	var helper apitypes.Helper
	if err := attributevalue.UnmarshalMap(result.Item, &helper); err != nil {
		return authMiddleware.CreateErrorResponse(500, "Internal server error"), nil
	}

	if helper.AccountID != authCtx.AccountID {
		return authMiddleware.CreateErrorResponse(404, "Helper not found"), nil
	}

	return authMiddleware.CreateSuccessResponse(200, "Helper retrieved successfully", map[string]interface{}{
		"helper_id":        helper.HelperID,
		"short_key":        helper.ShortKey,
		"name":             helper.Name,
		"description":      helper.Description,
		"helper_type":      helper.HelperType,
		"category":         helper.Category,
		"status":           helper.Status,
		"enabled":          helper.Enabled,
		"config":           helper.Config,
		"config_schema":    helper.ConfigSchema,
		"connection_id":    helper.ConnectionID,
		"execution_count":  helper.ExecutionCount,
		"last_executed_at": helper.LastExecutedAt,
		"created_at":       helper.CreatedAt,
		"updated_at":       helper.UpdatedAt,
	}), nil
}

func createHelper(ctx context.Context, event events.APIGatewayV2HTTPRequest, authCtx *apitypes.AuthContext) (events.APIGatewayV2HTTPResponse, error) {
	log.Printf("Create helper for account: %s", authCtx.AccountID)

	if !authCtx.Permissions.CanManageHelpers {
		return authMiddleware.CreateErrorResponse(403, "Permission denied"), nil
	}

	var req CreateHelperRequest
	if err := json.Unmarshal([]byte(event.Body), &req); err != nil {
		return authMiddleware.CreateErrorResponse(400, "Invalid request format"), nil
	}

	if req.Name == "" {
		return authMiddleware.CreateErrorResponse(400, "Name is required"), nil
	}
	if req.HelperType == "" {
		return authMiddleware.CreateErrorResponse(400, "Helper type is required"), nil
	}

	// Validate helper type is registered
	if !helperEngine.IsRegistered(req.HelperType) {
		return authMiddleware.CreateErrorResponse(400, fmt.Sprintf("Unknown helper type: %s", req.HelperType)), nil
	}

	// Validate config against helper schema
	helperInstance, err := helperEngine.NewHelper(req.HelperType)
	if err != nil {
		return authMiddleware.CreateErrorResponse(400, fmt.Sprintf("Invalid helper type: %s", req.HelperType)), nil
	}
	if req.Config != nil {
		if err := helperInstance.ValidateConfig(req.Config); err != nil {
			return authMiddleware.CreateErrorResponse(400, fmt.Sprintf("Invalid config: %v", err)), nil
		}
	}

	// Auto-populate category and config schema from registry
	category := req.Category
	if category == "" {
		category = helperInstance.GetCategory()
	}

	now := time.Now().UTC()
	helperID := "helper:" + uuid.Must(uuid.NewV7()).String()

	shortKey, err := nanoid.New()
	if err != nil {
		log.Printf("Failed to generate short key: %v", err)
		return authMiddleware.CreateErrorResponse(500, "Failed to create helper"), nil
	}

	helper := apitypes.Helper{
		HelperID:     helperID,
		AccountID:    authCtx.AccountID,
		CreatedBy:    authCtx.UserID,
		ConnectionID: req.ConnectionID,
		ShortKey:     shortKey,
		Name:         req.Name,
		Description:  req.Description,
		HelperType:   req.HelperType,
		Category:     category,
		Status:       "active",
		Config:       req.Config,
		ConfigSchema: helperInstance.GetConfigSchema(),
		Enabled:      true,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(os.Getenv("COGNITO_REGION")))
	if err != nil {
		return authMiddleware.CreateErrorResponse(500, "Internal server error"), nil
	}
	db := dynamodb.NewFromConfig(cfg)

	item, err := attributevalue.MarshalMap(helper)
	if err != nil {
		return authMiddleware.CreateErrorResponse(500, "Failed to create helper"), nil
	}

	_, err = db.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(helpersTable),
		Item:      item,
	})
	if err != nil {
		log.Printf("Failed to store helper: %v", err)
		return authMiddleware.CreateErrorResponse(500, "Failed to create helper"), nil
	}

	return authMiddleware.CreateSuccessResponse(201, "Helper created successfully", map[string]interface{}{
		"helper_id":   helperID,
		"short_key":   shortKey,
		"name":        req.Name,
		"helper_type": req.HelperType,
		"category":    category,
		"created_at":  now,
	}), nil
}

func updateHelper(ctx context.Context, event events.APIGatewayV2HTTPRequest, authCtx *apitypes.AuthContext) (events.APIGatewayV2HTTPResponse, error) {
	helperID := event.PathParameters["helper_id"]
	if helperID == "" {
		return authMiddleware.CreateErrorResponse(400, "Helper ID is required"), nil
	}

	if !authCtx.Permissions.CanManageHelpers {
		return authMiddleware.CreateErrorResponse(403, "Permission denied"), nil
	}

	var req UpdateHelperRequest
	if err := json.Unmarshal([]byte(event.Body), &req); err != nil {
		return authMiddleware.CreateErrorResponse(400, "Invalid request format"), nil
	}

	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(os.Getenv("COGNITO_REGION")))
	if err != nil {
		return authMiddleware.CreateErrorResponse(500, "Internal server error"), nil
	}
	db := dynamodb.NewFromConfig(cfg)

	// Verify ownership
	existing, err := db.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(helpersTable),
		Key: map[string]ddbtypes.AttributeValue{
			"helper_id": &ddbtypes.AttributeValueMemberS{Value: helperID},
		},
	})
	if err != nil || existing.Item == nil {
		return authMiddleware.CreateErrorResponse(404, "Helper not found"), nil
	}

	var existingHelper apitypes.Helper
	if err := attributevalue.UnmarshalMap(existing.Item, &existingHelper); err != nil || existingHelper.AccountID != authCtx.AccountID {
		return authMiddleware.CreateErrorResponse(404, "Helper not found"), nil
	}

	// Build update expression
	updateParts := []string{"updated_at = :updated_at"}
	exprValues := map[string]ddbtypes.AttributeValue{
		":updated_at": &ddbtypes.AttributeValueMemberS{Value: time.Now().UTC().Format(time.RFC3339)},
	}
	exprNames := map[string]string{}

	if req.Name != "" {
		updateParts = append(updateParts, "#n = :name")
		exprValues[":name"] = &ddbtypes.AttributeValueMemberS{Value: req.Name}
		exprNames["#n"] = "name"
	}
	if req.Description != "" {
		updateParts = append(updateParts, "description = :description")
		exprValues[":description"] = &ddbtypes.AttributeValueMemberS{Value: req.Description}
	}
	if req.Config != nil {
		// Validate config against helper schema
		if helperEngine.IsRegistered(existingHelper.HelperType) {
			helperInstance, err := helperEngine.NewHelper(existingHelper.HelperType)
			if err == nil {
				if err := helperInstance.ValidateConfig(req.Config); err != nil {
					return authMiddleware.CreateErrorResponse(400, fmt.Sprintf("Invalid config: %v", err)), nil
				}
			}
		}
		configAV, err := attributevalue.MarshalMap(req.Config)
		if err != nil {
			return authMiddleware.CreateErrorResponse(500, "Failed to process config"), nil
		}
		updateParts = append(updateParts, "config = :config")
		exprValues[":config"] = &ddbtypes.AttributeValueMemberM{Value: configAV}
	}
	if req.Enabled != nil {
		updateParts = append(updateParts, "enabled = :enabled")
		exprValues[":enabled"] = &ddbtypes.AttributeValueMemberBOOL{Value: *req.Enabled}
	}
	if req.ConnectionID != "" {
		updateParts = append(updateParts, "connection_id = :connection_id")
		exprValues[":connection_id"] = &ddbtypes.AttributeValueMemberS{Value: req.ConnectionID}
	}

	// Handle schedule changes
	if req.ScheduleEnabled != nil || req.CronExpression != "" {
		scheduleEnabled := existingHelper.ScheduleEnabled
		cronExpr := existingHelper.CronExpression
		if req.ScheduleEnabled != nil {
			scheduleEnabled = *req.ScheduleEnabled
		}
		if req.CronExpression != "" {
			cronExpr = req.CronExpression
		}

		if scheduleEnabled && cronExpr != "" {
			// Create or update EventBridge rule
			ruleARN, err := upsertScheduleRule(ctx, cfg, helperID, existingHelper.AccountID, cronExpr)
			if err != nil {
				log.Printf("Failed to manage schedule rule: %v", err)
				return authMiddleware.CreateErrorResponse(500, "Failed to update schedule"), nil
			}
			updateParts = append(updateParts, "schedule_enabled = :sched_enabled, cron_expression = :cron, schedule_rule_arn = :rule_arn")
			exprValues[":sched_enabled"] = &ddbtypes.AttributeValueMemberBOOL{Value: true}
			exprValues[":cron"] = &ddbtypes.AttributeValueMemberS{Value: cronExpr}
			exprValues[":rule_arn"] = &ddbtypes.AttributeValueMemberS{Value: ruleARN}
		} else if !scheduleEnabled {
			// Disable schedule rule
			if existingHelper.ScheduleRuleARN != "" {
				disableScheduleRule(ctx, cfg, helperID)
			}
			updateParts = append(updateParts, "schedule_enabled = :sched_enabled")
			exprValues[":sched_enabled"] = &ddbtypes.AttributeValueMemberBOOL{Value: false}
		}
	}

	updateInput := &dynamodb.UpdateItemInput{
		TableName: aws.String(helpersTable),
		Key: map[string]ddbtypes.AttributeValue{
			"helper_id": &ddbtypes.AttributeValueMemberS{Value: helperID},
		},
		UpdateExpression:          aws.String("SET " + strings.Join(updateParts, ", ")),
		ExpressionAttributeValues: exprValues,
	}
	if len(exprNames) > 0 {
		updateInput.ExpressionAttributeNames = exprNames
	}

	_, err = db.UpdateItem(ctx, updateInput)
	if err != nil {
		log.Printf("Failed to update helper: %v", err)
		return authMiddleware.CreateErrorResponse(500, "Failed to update helper"), nil
	}

	return authMiddleware.CreateSuccessResponse(200, "Helper updated successfully", map[string]interface{}{
		"helper_id": helperID,
	}), nil
}

func deleteHelper(ctx context.Context, event events.APIGatewayV2HTTPRequest, authCtx *apitypes.AuthContext) (events.APIGatewayV2HTTPResponse, error) {
	helperID := event.PathParameters["helper_id"]
	if helperID == "" {
		return authMiddleware.CreateErrorResponse(400, "Helper ID is required"), nil
	}

	if !authCtx.Permissions.CanManageHelpers {
		return authMiddleware.CreateErrorResponse(403, "Permission denied"), nil
	}

	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(os.Getenv("COGNITO_REGION")))
	if err != nil {
		return authMiddleware.CreateErrorResponse(500, "Internal server error"), nil
	}
	db := dynamodb.NewFromConfig(cfg)

	// Verify ownership
	existing, err := db.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(helpersTable),
		Key: map[string]ddbtypes.AttributeValue{
			"helper_id": &ddbtypes.AttributeValueMemberS{Value: helperID},
		},
	})
	if err != nil || existing.Item == nil {
		return authMiddleware.CreateErrorResponse(404, "Helper not found"), nil
	}

	var existingHelper apitypes.Helper
	if err := attributevalue.UnmarshalMap(existing.Item, &existingHelper); err != nil || existingHelper.AccountID != authCtx.AccountID {
		return authMiddleware.CreateErrorResponse(404, "Helper not found"), nil
	}

	// Clean up EventBridge schedule rule if one exists
	if existingHelper.ScheduleRuleARN != "" {
		deleteScheduleRule(ctx, cfg, helperID)
	}

	// Soft delete by setting status to deleted
	_, err = db.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(helpersTable),
		Key: map[string]ddbtypes.AttributeValue{
			"helper_id": &ddbtypes.AttributeValueMemberS{Value: helperID},
		},
		UpdateExpression: aws.String("SET #s = :status, updated_at = :updated_at, enabled = :enabled, schedule_enabled = :sched"),
		ExpressionAttributeNames: map[string]string{
			"#s": "status",
		},
		ExpressionAttributeValues: map[string]ddbtypes.AttributeValue{
			":status":     &ddbtypes.AttributeValueMemberS{Value: "deleted"},
			":updated_at": &ddbtypes.AttributeValueMemberS{Value: time.Now().UTC().Format(time.RFC3339)},
			":enabled":    &ddbtypes.AttributeValueMemberBOOL{Value: false},
			":sched":      &ddbtypes.AttributeValueMemberBOOL{Value: false},
		},
	})
	if err != nil {
		log.Printf("Failed to delete helper: %v", err)
		return authMiddleware.CreateErrorResponse(500, "Failed to delete helper"), nil
	}

	return authMiddleware.CreateSuccessResponse(200, "Helper deleted successfully", map[string]interface{}{
		"helper_id": helperID,
	}), nil
}

func executeHelper(ctx context.Context, event events.APIGatewayV2HTTPRequest, authCtx *apitypes.AuthContext) (events.APIGatewayV2HTTPResponse, error) {
	// Extract helper_id from path (e.g., /helpers/{helper_id}/execute)
	path := event.RequestContext.HTTP.Path
	parts := strings.Split(strings.TrimPrefix(path, "/helpers/"), "/")
	if len(parts) < 2 {
		return authMiddleware.CreateErrorResponse(400, "Invalid path"), nil
	}
	helperID := parts[0]

	log.Printf("Execute helper %s for account: %s", helperID, authCtx.AccountID)

	if !authCtx.Permissions.CanExecuteHelpers {
		return authMiddleware.CreateErrorResponse(403, "Permission denied"), nil
	}

	var req ExecuteHelperRequest
	if event.Body != "" {
		if err := json.Unmarshal([]byte(event.Body), &req); err != nil {
			return authMiddleware.CreateErrorResponse(400, "Invalid request format"), nil
		}
	}

	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(os.Getenv("COGNITO_REGION")))
	if err != nil {
		return authMiddleware.CreateErrorResponse(500, "Internal server error"), nil
	}
	db := dynamodb.NewFromConfig(cfg)

	// Verify helper exists and belongs to account
	result, err := db.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(helpersTable),
		Key: map[string]ddbtypes.AttributeValue{
			"helper_id": &ddbtypes.AttributeValueMemberS{Value: helperID},
		},
	})
	if err != nil || result.Item == nil {
		return authMiddleware.CreateErrorResponse(404, "Helper not found"), nil
	}

	var helper apitypes.Helper
	if err := attributevalue.UnmarshalMap(result.Item, &helper); err != nil || helper.AccountID != authCtx.AccountID {
		return authMiddleware.CreateErrorResponse(404, "Helper not found"), nil
	}

	if !helper.Enabled {
		return authMiddleware.CreateErrorResponse(400, "Helper is disabled"), nil
	}

	// Create execution record with status="queued" directly.
	// DynamoDB Streams will auto-dispatch to SQS FIFO for async processing.
	now := time.Now().UTC()
	executionID := "exec:" + uuid.Must(uuid.NewV7()).String()
	ttl := now.Add(7 * 24 * time.Hour).Unix()

	execution := apitypes.Execution{
		ExecutionID:  executionID,
		HelperID:     helperID,
		AccountID:    authCtx.AccountID,
		UserID:       authCtx.UserID,
		ConnectionID: helper.ConnectionID,
		ContactID:    req.ContactID,
		Status:       "queued",
		TriggerType:  "manual",
		Input:        req.Input,
		CreatedAt:    now.Format(time.RFC3339),
		StartedAt:    now,
		TTL:          &ttl,
	}

	execItem, err := attributevalue.MarshalMap(execution)
	if err != nil {
		return authMiddleware.CreateErrorResponse(500, "Failed to create execution"), nil
	}

	_, err = db.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(executionsTable),
		Item:      execItem,
	})
	if err != nil {
		log.Printf("Failed to store execution: %v", err)
		return authMiddleware.CreateErrorResponse(500, "Failed to create execution"), nil
	}

	return authMiddleware.CreateSuccessResponse(202, "Helper execution queued", map[string]interface{}{
		"execution_id": executionID,
		"helper_id":    helperID,
		"status":       "queued",
		"started_at":   now,
	}), nil
}

// scheduleRuleName returns the EventBridge rule name for a helper
func scheduleRuleName(helperID string) string {
	// Replace colons with dashes for valid rule names
	safe := strings.ReplaceAll(helperID, ":", "-")
	stage := os.Getenv("STAGE")
	return fmt.Sprintf("mfh-%s-helper-schedule-%s", stage, safe)
}

// upsertScheduleRule creates or updates an EventBridge rule for a helper's schedule
func upsertScheduleRule(ctx context.Context, cfg aws.Config, helperID, accountID, cronExpr string) (string, error) {
	if schedulerFunctionARN == "" {
		return "", fmt.Errorf("SCHEDULER_FUNCTION_ARN not configured")
	}

	ebClient := eventbridge.NewFromConfig(cfg)
	ruleName := scheduleRuleName(helperID)

	// Create/update the EventBridge rule
	ruleResult, err := ebClient.PutRule(ctx, &eventbridge.PutRuleInput{
		Name:               aws.String(ruleName),
		ScheduleExpression: aws.String(cronExpr),
		State:              ebtypes.RuleStateEnabled,
		Description:        aws.String(fmt.Sprintf("Schedule for helper %s", helperID)),
	})
	if err != nil {
		return "", fmt.Errorf("failed to create EventBridge rule: %w", err)
	}

	// Build the target input payload
	targetInput, _ := json.Marshal(map[string]string{
		"helper_id":  helperID,
		"account_id": accountID,
	})

	// Set the Lambda function as the target
	_, err = ebClient.PutTargets(ctx, &eventbridge.PutTargetsInput{
		Rule: aws.String(ruleName),
		Targets: []ebtypes.Target{
			{
				Id:    aws.String("scheduler"),
				Arn:   aws.String(schedulerFunctionARN),
				Input: aws.String(string(targetInput)),
			},
		},
	})
	if err != nil {
		return "", fmt.Errorf("failed to set EventBridge target: %w", err)
	}

	// Add permission for EventBridge to invoke the scheduler Lambda
	lambdaClient := lambdasvc.NewFromConfig(cfg)
	statementID := fmt.Sprintf("eb-%s", ruleName)
	_, _ = lambdaClient.RemovePermission(ctx, &lambdasvc.RemovePermissionInput{
		FunctionName: aws.String(schedulerFunctionARN),
		StatementId:  aws.String(statementID),
	})
	_, err = lambdaClient.AddPermission(ctx, &lambdasvc.AddPermissionInput{
		FunctionName: aws.String(schedulerFunctionARN),
		StatementId:  aws.String(statementID),
		Action:       aws.String("lambda:InvokeFunction"),
		Principal:    aws.String("events.amazonaws.com"),
		SourceArn:    ruleResult.RuleArn,
	})
	if err != nil {
		log.Printf("Warning: failed to add Lambda permission for EventBridge rule %s: %v", ruleName, err)
	}

	log.Printf("Created/updated EventBridge rule %s for helper %s", ruleName, helperID)
	return aws.ToString(ruleResult.RuleArn), nil
}

// disableScheduleRule disables an EventBridge rule for a helper
func disableScheduleRule(ctx context.Context, cfg aws.Config, helperID string) {
	ebClient := eventbridge.NewFromConfig(cfg)
	ruleName := scheduleRuleName(helperID)

	_, err := ebClient.DisableRule(ctx, &eventbridge.DisableRuleInput{
		Name: aws.String(ruleName),
	})
	if err != nil {
		log.Printf("Failed to disable EventBridge rule %s: %v", ruleName, err)
	} else {
		log.Printf("Disabled EventBridge rule %s", ruleName)
	}
}

// deleteScheduleRule removes an EventBridge rule and its targets for a helper
func deleteScheduleRule(ctx context.Context, cfg aws.Config, helperID string) {
	ebClient := eventbridge.NewFromConfig(cfg)
	ruleName := scheduleRuleName(helperID)

	// Remove targets first (required before deleting rule)
	_, _ = ebClient.RemoveTargets(ctx, &eventbridge.RemoveTargetsInput{
		Rule: aws.String(ruleName),
		Ids:  []string{"scheduler"},
	})

	// Remove Lambda invoke permission
	lambdaClient := lambdasvc.NewFromConfig(cfg)
	statementID := fmt.Sprintf("eb-%s", ruleName)
	_, _ = lambdaClient.RemovePermission(ctx, &lambdasvc.RemovePermissionInput{
		FunctionName: aws.String(schedulerFunctionARN),
		StatementId:  aws.String(statementID),
	})

	// Delete the rule
	_, err := ebClient.DeleteRule(ctx, &eventbridge.DeleteRuleInput{
		Name: aws.String(ruleName),
	})
	if err != nil {
		log.Printf("Failed to delete EventBridge rule %s: %v", ruleName, err)
	} else {
		log.Printf("Deleted EventBridge rule %s", ruleName)
	}
}
