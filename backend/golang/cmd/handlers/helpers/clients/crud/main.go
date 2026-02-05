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
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/google/uuid"
	helperEngine "github.com/myfusionhelper/api/internal/helpers"
	authMiddleware "github.com/myfusionhelper/api/internal/middleware/auth"
	apitypes "github.com/myfusionhelper/api/internal/types"
)

var (
	helpersTable             = os.Getenv("HELPERS_TABLE")
	executionsTable          = os.Getenv("EXECUTIONS_TABLE")
	helperExecutionQueueURL  = os.Getenv("HELPER_EXECUTION_QUEUE_URL")
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
	Name         string                 `json:"name"`
	Description  string                 `json:"description"`
	Config       map[string]interface{} `json:"config"`
	Enabled      *bool                  `json:"enabled"`
	ConnectionID string                 `json:"connection_id"`
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
	helperID := "helper:" + uuid.New().String()

	helper := apitypes.Helper{
		HelperID:     helperID,
		AccountID:    authCtx.AccountID,
		CreatedBy:    authCtx.UserID,
		ConnectionID: req.ConnectionID,
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
		"name":        req.Name,
		"helper_type": req.HelperType,
		"category":    req.Category,
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
		configJSON, _ := json.Marshal(req.Config)
		updateParts = append(updateParts, "config = :config")
		exprValues[":config"] = &ddbtypes.AttributeValueMemberS{Value: string(configJSON)}
	}
	if req.Enabled != nil {
		updateParts = append(updateParts, "enabled = :enabled")
		exprValues[":enabled"] = &ddbtypes.AttributeValueMemberBOOL{Value: *req.Enabled}
	}
	if req.ConnectionID != "" {
		updateParts = append(updateParts, "connection_id = :connection_id")
		exprValues[":connection_id"] = &ddbtypes.AttributeValueMemberS{Value: req.ConnectionID}
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

	// Soft delete by setting status to deleted
	_, err = db.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(helpersTable),
		Key: map[string]ddbtypes.AttributeValue{
			"helper_id": &ddbtypes.AttributeValueMemberS{Value: helperID},
		},
		UpdateExpression: aws.String("SET #s = :status, updated_at = :updated_at, enabled = :enabled"),
		ExpressionAttributeNames: map[string]string{
			"#s": "status",
		},
		ExpressionAttributeValues: map[string]ddbtypes.AttributeValue{
			":status":     &ddbtypes.AttributeValueMemberS{Value: "deleted"},
			":updated_at": &ddbtypes.AttributeValueMemberS{Value: time.Now().UTC().Format(time.RFC3339)},
			":enabled":    &ddbtypes.AttributeValueMemberBOOL{Value: false},
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

	// Create execution record
	now := time.Now().UTC()
	executionID := "exec:" + uuid.New().String()

	execution := apitypes.Execution{
		ExecutionID:  executionID,
		HelperID:     helperID,
		AccountID:    authCtx.AccountID,
		UserID:       authCtx.UserID,
		ConnectionID: helper.ConnectionID,
		ContactID:    req.ContactID,
		Status:       "pending",
		TriggerType:  "manual",
		Input:        req.Input,
		CreatedAt:    now.Format(time.RFC3339),
		StartedAt:    now,
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

	// Send to SQS for async execution
	sqsMessage := map[string]interface{}{
		"execution_id":  executionID,
		"helper_id":     helperID,
		"helper_type":   helper.HelperType,
		"account_id":    authCtx.AccountID,
		"user_id":       authCtx.UserID,
		"connection_id": helper.ConnectionID,
		"contact_id":    req.ContactID,
		"config":        helper.Config,
		"input":         req.Input,
		"retry_count":   0,
	}

	sqsBody, err := json.Marshal(sqsMessage)
	if err != nil {
		log.Printf("Failed to marshal SQS message: %v", err)
		return authMiddleware.CreateErrorResponse(500, "Failed to queue execution"), nil
	}

	sqsClient := sqs.NewFromConfig(cfg)
	_, err = sqsClient.SendMessage(ctx, &sqs.SendMessageInput{
		QueueUrl:               aws.String(helperExecutionQueueURL),
		MessageBody:            aws.String(string(sqsBody)),
		MessageGroupId:         aws.String(authCtx.AccountID), // FIFO: group by account
		MessageDeduplicationId: aws.String(executionID),
	})
	if err != nil {
		log.Printf("Failed to send SQS message: %v", err)
		// Update execution status to failed
		_, _ = db.UpdateItem(ctx, &dynamodb.UpdateItemInput{
			TableName: aws.String(executionsTable),
			Key: map[string]ddbtypes.AttributeValue{
				"execution_id": &ddbtypes.AttributeValueMemberS{Value: executionID},
			},
			UpdateExpression: aws.String("SET #s = :status, error_message = :error"),
			ExpressionAttributeNames: map[string]string{"#s": "status"},
			ExpressionAttributeValues: map[string]ddbtypes.AttributeValue{
				":status": &ddbtypes.AttributeValueMemberS{Value: "failed"},
				":error":  &ddbtypes.AttributeValueMemberS{Value: "Failed to queue for execution"},
			},
		})
		return authMiddleware.CreateErrorResponse(500, "Failed to queue execution"), nil
	}

	// Update execution status to queued
	_, _ = db.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(executionsTable),
		Key: map[string]ddbtypes.AttributeValue{
			"execution_id": &ddbtypes.AttributeValueMemberS{Value: executionID},
		},
		UpdateExpression: aws.String("SET #s = :status"),
		ExpressionAttributeNames: map[string]string{"#s": "status"},
		ExpressionAttributeValues: map[string]ddbtypes.AttributeValue{
			":status": &ddbtypes.AttributeValueMemberS{Value: "queued"},
		},
	})

	return authMiddleware.CreateSuccessResponse(202, "Helper execution queued", map[string]interface{}{
		"execution_id": executionID,
		"helper_id":    helperID,
		"status":       "queued",
		"started_at":   now,
	}), nil
}
