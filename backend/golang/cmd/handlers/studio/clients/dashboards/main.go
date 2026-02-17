package dashboards

import (
	"context"
	"encoding/json"
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
	"github.com/google/uuid"

	"github.com/myfusionhelper/api/internal/apiutil"
	authMiddleware "github.com/myfusionhelper/api/internal/middleware/auth"
	apitypes "github.com/myfusionhelper/api/internal/types"
)

var dashboardsTable = os.Getenv("DASHBOARDS_TABLE")

// HandleWithAuth routes dashboard requests by path + method.
func HandleWithAuth(ctx context.Context, event events.APIGatewayV2HTTPRequest, authCtx *apitypes.AuthContext) (events.APIGatewayV2HTTPResponse, error) {
	path := event.RequestContext.HTTP.Path
	method := event.RequestContext.HTTP.Method

	switch {
	case path == "/studio/dashboards" && method == "GET":
		return listDashboards(ctx, event, authCtx)
	case path == "/studio/dashboards" && method == "POST":
		return createDashboard(ctx, event, authCtx)
	case strings.HasPrefix(path, "/studio/dashboards/") && method == "GET":
		return getDashboard(ctx, event, authCtx)
	case strings.HasPrefix(path, "/studio/dashboards/") && method == "PUT":
		return updateDashboard(ctx, event, authCtx)
	case strings.HasPrefix(path, "/studio/dashboards/") && method == "DELETE":
		return deleteDashboard(ctx, event, authCtx)
	default:
		return authMiddleware.CreateErrorResponse(404, "Not Found"), nil
	}
}

func getDashboardID(path string) string {
	// /studio/dashboards/{id}
	parts := strings.Split(strings.TrimPrefix(path, "/"), "/")
	if len(parts) >= 3 {
		return parts[2]
	}
	return ""
}

func listDashboards(ctx context.Context, event events.APIGatewayV2HTTPRequest, authCtx *apitypes.AuthContext) (events.APIGatewayV2HTTPResponse, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(os.Getenv("COGNITO_REGION")))
	if err != nil {
		log.Printf("Failed to load AWS config: %v", err)
		return authMiddleware.CreateErrorResponse(500, "Internal server error"), nil
	}

	ddb := dynamodb.NewFromConfig(cfg)
	result, err := ddb.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(dashboardsTable),
		IndexName:              aws.String("AccountIdIndex"),
		KeyConditionExpression: aws.String("account_id = :aid"),
		FilterExpression:       aws.String("#s <> :deleted"),
		ExpressionAttributeNames: map[string]string{
			"#s": "status",
		},
		ExpressionAttributeValues: map[string]ddbtypes.AttributeValue{
			":aid":     &ddbtypes.AttributeValueMemberS{Value: authCtx.AccountID},
			":deleted": &ddbtypes.AttributeValueMemberS{Value: "deleted"},
		},
	})
	if err != nil {
		log.Printf("Failed to query dashboards: %v", err)
		return authMiddleware.CreateErrorResponse(500, "Failed to list dashboards"), nil
	}

	var dashboards []apitypes.Dashboard
	if err := attributevalue.UnmarshalListOfMaps(result.Items, &dashboards); err != nil {
		log.Printf("Failed to unmarshal dashboards: %v", err)
		return authMiddleware.CreateErrorResponse(500, "Failed to parse dashboards"), nil
	}

	if dashboards == nil {
		dashboards = []apitypes.Dashboard{}
	}

	return authMiddleware.CreateSuccessResponse(200, "Dashboards retrieved", map[string]interface{}{
		"dashboards": dashboards,
	}), nil
}

type createDashboardRequest struct {
	Name         string                    `json:"name"`
	Description  string                    `json:"description,omitempty"`
	Widgets      []apitypes.DashboardWidget `json:"widgets,omitempty"`
	ConnectionID string                    `json:"connection_id,omitempty"`
	TemplateID   string                    `json:"template_id,omitempty"`
}

func createDashboard(ctx context.Context, event events.APIGatewayV2HTTPRequest, authCtx *apitypes.AuthContext) (events.APIGatewayV2HTTPResponse, error) {
	var req createDashboardRequest
	if err := json.Unmarshal([]byte(apiutil.GetBody(event)), &req); err != nil {
		return authMiddleware.CreateErrorResponse(400, "Invalid request body"), nil
	}

	if req.Name == "" {
		return authMiddleware.CreateErrorResponse(400, "Name is required"), nil
	}

	now := time.Now().UTC().Format(time.RFC3339)
	id := "dash:" + uuid.Must(uuid.NewV7()).String()

	widgets := req.Widgets
	if widgets == nil {
		widgets = []apitypes.DashboardWidget{}
	}

	dashboard := apitypes.Dashboard{
		DashboardID:  id,
		AccountID:    authCtx.AccountID,
		Name:         req.Name,
		Description:  req.Description,
		Widgets:      widgets,
		TemplateID:   req.TemplateID,
		ConnectionID: req.ConnectionID,
		Status:       "active",
		CreatedBy:    authCtx.UserID,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	item, err := attributevalue.MarshalMap(dashboard)
	if err != nil {
		log.Printf("Failed to marshal dashboard: %v", err)
		return authMiddleware.CreateErrorResponse(500, "Failed to create dashboard"), nil
	}

	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(os.Getenv("COGNITO_REGION")))
	if err != nil {
		log.Printf("Failed to load AWS config: %v", err)
		return authMiddleware.CreateErrorResponse(500, "Internal server error"), nil
	}

	ddb := dynamodb.NewFromConfig(cfg)
	_, err = ddb.PutItem(ctx, &dynamodb.PutItemInput{
		TableName:           aws.String(dashboardsTable),
		Item:                item,
		ConditionExpression: aws.String("attribute_not_exists(dashboard_id)"),
	})
	if err != nil {
		log.Printf("Failed to create dashboard: %v", err)
		return authMiddleware.CreateErrorResponse(500, "Failed to create dashboard"), nil
	}

	return authMiddleware.CreateSuccessResponse(201, "Dashboard created", dashboard), nil
}

func getDashboard(ctx context.Context, event events.APIGatewayV2HTTPRequest, authCtx *apitypes.AuthContext) (events.APIGatewayV2HTTPResponse, error) {
	dashboardID := getDashboardID(event.RequestContext.HTTP.Path)
	if dashboardID == "" {
		return authMiddleware.CreateErrorResponse(400, "Dashboard ID is required"), nil
	}

	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(os.Getenv("COGNITO_REGION")))
	if err != nil {
		log.Printf("Failed to load AWS config: %v", err)
		return authMiddleware.CreateErrorResponse(500, "Internal server error"), nil
	}

	ddb := dynamodb.NewFromConfig(cfg)
	result, err := ddb.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(dashboardsTable),
		Key: map[string]ddbtypes.AttributeValue{
			"dashboard_id": &ddbtypes.AttributeValueMemberS{Value: dashboardID},
		},
	})
	if err != nil || result.Item == nil {
		return authMiddleware.CreateErrorResponse(404, "Dashboard not found"), nil
	}

	var dashboard apitypes.Dashboard
	if err := attributevalue.UnmarshalMap(result.Item, &dashboard); err != nil {
		log.Printf("Failed to unmarshal dashboard: %v", err)
		return authMiddleware.CreateErrorResponse(500, "Failed to parse dashboard"), nil
	}

	if dashboard.AccountID != authCtx.AccountID {
		return authMiddleware.CreateErrorResponse(403, "Access denied"), nil
	}

	if dashboard.Status == "deleted" {
		return authMiddleware.CreateErrorResponse(404, "Dashboard not found"), nil
	}

	return authMiddleware.CreateSuccessResponse(200, "Dashboard retrieved", dashboard), nil
}

type updateDashboardRequest struct {
	Name        *string                    `json:"name,omitempty"`
	Description *string                    `json:"description,omitempty"`
	Widgets     *[]apitypes.DashboardWidget `json:"widgets,omitempty"`
}

func updateDashboard(ctx context.Context, event events.APIGatewayV2HTTPRequest, authCtx *apitypes.AuthContext) (events.APIGatewayV2HTTPResponse, error) {
	dashboardID := getDashboardID(event.RequestContext.HTTP.Path)
	if dashboardID == "" {
		return authMiddleware.CreateErrorResponse(400, "Dashboard ID is required"), nil
	}

	var req updateDashboardRequest
	if err := json.Unmarshal([]byte(apiutil.GetBody(event)), &req); err != nil {
		return authMiddleware.CreateErrorResponse(400, "Invalid request body"), nil
	}

	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(os.Getenv("COGNITO_REGION")))
	if err != nil {
		log.Printf("Failed to load AWS config: %v", err)
		return authMiddleware.CreateErrorResponse(500, "Internal server error"), nil
	}

	ddb := dynamodb.NewFromConfig(cfg)

	// Verify ownership
	getResult, err := ddb.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(dashboardsTable),
		Key: map[string]ddbtypes.AttributeValue{
			"dashboard_id": &ddbtypes.AttributeValueMemberS{Value: dashboardID},
		},
	})
	if err != nil || getResult.Item == nil {
		return authMiddleware.CreateErrorResponse(404, "Dashboard not found"), nil
	}

	var existing apitypes.Dashboard
	if err := attributevalue.UnmarshalMap(getResult.Item, &existing); err != nil {
		return authMiddleware.CreateErrorResponse(500, "Failed to parse dashboard"), nil
	}
	if existing.AccountID != authCtx.AccountID {
		return authMiddleware.CreateErrorResponse(403, "Access denied"), nil
	}

	// Build update expression
	now := time.Now().UTC().Format(time.RFC3339)
	updateExpr := "SET updated_at = :now"
	exprValues := map[string]ddbtypes.AttributeValue{
		":now": &ddbtypes.AttributeValueMemberS{Value: now},
	}
	exprNames := map[string]string{}

	if req.Name != nil {
		updateExpr += ", #n = :name"
		exprNames["#n"] = "name"
		exprValues[":name"] = &ddbtypes.AttributeValueMemberS{Value: *req.Name}
	}
	if req.Description != nil {
		updateExpr += ", description = :desc"
		exprValues[":desc"] = &ddbtypes.AttributeValueMemberS{Value: *req.Description}
	}
	if req.Widgets != nil {
		widgetsAV, err := attributevalue.MarshalList(*req.Widgets)
		if err != nil {
			return authMiddleware.CreateErrorResponse(500, "Failed to process widgets"), nil
		}
		updateExpr += ", widgets = :widgets"
		exprValues[":widgets"] = &ddbtypes.AttributeValueMemberL{Value: widgetsAV}
	}

	input := &dynamodb.UpdateItemInput{
		TableName: aws.String(dashboardsTable),
		Key: map[string]ddbtypes.AttributeValue{
			"dashboard_id": &ddbtypes.AttributeValueMemberS{Value: dashboardID},
		},
		UpdateExpression:          aws.String(updateExpr),
		ExpressionAttributeValues: exprValues,
		ReturnValues:              ddbtypes.ReturnValueAllNew,
	}
	if len(exprNames) > 0 {
		input.ExpressionAttributeNames = exprNames
	}

	updateResult, err := ddb.UpdateItem(ctx, input)
	if err != nil {
		log.Printf("Failed to update dashboard: %v", err)
		return authMiddleware.CreateErrorResponse(500, "Failed to update dashboard"), nil
	}

	var updated apitypes.Dashboard
	if err := attributevalue.UnmarshalMap(updateResult.Attributes, &updated); err != nil {
		return authMiddleware.CreateErrorResponse(500, "Failed to parse updated dashboard"), nil
	}

	return authMiddleware.CreateSuccessResponse(200, "Dashboard updated", updated), nil
}

func deleteDashboard(ctx context.Context, event events.APIGatewayV2HTTPRequest, authCtx *apitypes.AuthContext) (events.APIGatewayV2HTTPResponse, error) {
	dashboardID := getDashboardID(event.RequestContext.HTTP.Path)
	if dashboardID == "" {
		return authMiddleware.CreateErrorResponse(400, "Dashboard ID is required"), nil
	}

	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(os.Getenv("COGNITO_REGION")))
	if err != nil {
		log.Printf("Failed to load AWS config: %v", err)
		return authMiddleware.CreateErrorResponse(500, "Internal server error"), nil
	}

	ddb := dynamodb.NewFromConfig(cfg)

	// Verify ownership
	getResult, err := ddb.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(dashboardsTable),
		Key: map[string]ddbtypes.AttributeValue{
			"dashboard_id": &ddbtypes.AttributeValueMemberS{Value: dashboardID},
		},
	})
	if err != nil || getResult.Item == nil {
		return authMiddleware.CreateErrorResponse(404, "Dashboard not found"), nil
	}

	var existing apitypes.Dashboard
	if err := attributevalue.UnmarshalMap(getResult.Item, &existing); err != nil {
		return authMiddleware.CreateErrorResponse(500, "Failed to parse dashboard"), nil
	}
	if existing.AccountID != authCtx.AccountID {
		return authMiddleware.CreateErrorResponse(403, "Access denied"), nil
	}

	// Soft delete
	now := time.Now().UTC().Format(time.RFC3339)
	_, err = ddb.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(dashboardsTable),
		Key: map[string]ddbtypes.AttributeValue{
			"dashboard_id": &ddbtypes.AttributeValueMemberS{Value: dashboardID},
		},
		UpdateExpression: aws.String("SET #s = :deleted, updated_at = :now"),
		ExpressionAttributeNames: map[string]string{
			"#s": "status",
		},
		ExpressionAttributeValues: map[string]ddbtypes.AttributeValue{
			":deleted": &ddbtypes.AttributeValueMemberS{Value: "deleted"},
			":now":     &ddbtypes.AttributeValueMemberS{Value: now},
		},
	})
	if err != nil {
		log.Printf("Failed to delete dashboard: %v", err)
		return authMiddleware.CreateErrorResponse(500, "Failed to delete dashboard"), nil
	}

	return authMiddleware.CreateSuccessResponse(200, "Dashboard deleted", map[string]interface{}{
		"dashboard_id": dashboardID,
	}), nil
}
