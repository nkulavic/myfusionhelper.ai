package templates

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

var (
	connectionsTable = os.Getenv("CONNECTIONS_TABLE")
	platformsTable   = os.Getenv("PLATFORMS_TABLE")
	dashboardsTable  = os.Getenv("DASHBOARDS_TABLE")
)

// HandleWithAuth routes template requests by path + method.
func HandleWithAuth(ctx context.Context, event events.APIGatewayV2HTTPRequest, authCtx *apitypes.AuthContext) (events.APIGatewayV2HTTPResponse, error) {
	path := event.RequestContext.HTTP.Path
	method := event.RequestContext.HTTP.Method

	switch {
	case path == "/studio/templates" && method == "GET":
		return listTemplates(ctx, event, authCtx)
	case strings.HasPrefix(path, "/studio/templates/") && strings.HasSuffix(path, "/apply") && method == "POST":
		return applyTemplate(ctx, event, authCtx)
	default:
		return authMiddleware.CreateErrorResponse(404, "Not Found"), nil
	}
}

// listTemplates returns templates filtered by the user's connected platforms.
func listTemplates(ctx context.Context, event events.APIGatewayV2HTTPRequest, authCtx *apitypes.AuthContext) (events.APIGatewayV2HTTPResponse, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(os.Getenv("COGNITO_REGION")))
	if err != nil {
		log.Printf("Failed to load AWS config: %v", err)
		return authMiddleware.CreateErrorResponse(500, "Internal server error"), nil
	}

	ddb := dynamodb.NewFromConfig(cfg)

	// Get user's connections to determine which platforms are connected
	connResult, err := ddb.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(connectionsTable),
		IndexName:              aws.String("AccountIdIndex"),
		KeyConditionExpression: aws.String("account_id = :aid"),
		FilterExpression:       aws.String("#s = :active"),
		ExpressionAttributeNames: map[string]string{
			"#s": "status",
		},
		ExpressionAttributeValues: map[string]ddbtypes.AttributeValue{
			":aid":    &ddbtypes.AttributeValueMemberS{Value: authCtx.AccountID},
			":active": &ddbtypes.AttributeValueMemberS{Value: "active"},
		},
	})
	if err != nil {
		log.Printf("Failed to query connections: %v", err)
		return authMiddleware.CreateErrorResponse(500, "Failed to load connections"), nil
	}

	// Extract platform IDs from connections
	platformIDs := make(map[string]bool)
	for _, item := range connResult.Items {
		if pidAttr, ok := item["platform_id"]; ok {
			if pidStr, ok := pidAttr.(*ddbtypes.AttributeValueMemberS); ok {
				platformIDs[pidStr.Value] = true
			}
		}
	}

	// Look up platform slugs from platform IDs
	var platformSlugs []string
	for pid := range platformIDs {
		platResult, err := ddb.GetItem(ctx, &dynamodb.GetItemInput{
			TableName: aws.String(platformsTable),
			Key: map[string]ddbtypes.AttributeValue{
				"platform_id": &ddbtypes.AttributeValueMemberS{Value: pid},
			},
		})
		if err != nil || platResult.Item == nil {
			continue
		}
		if slugAttr, ok := platResult.Item["slug"]; ok {
			if slugStr, ok := slugAttr.(*ddbtypes.AttributeValueMemberS); ok {
				platformSlugs = append(platformSlugs, slugStr.Value)
			}
		}
	}

	templates := GetTemplatesForPlatforms(platformSlugs)

	return authMiddleware.CreateSuccessResponse(200, "Templates retrieved", map[string]interface{}{
		"templates": templates,
	}), nil
}

func getTemplateID(path string) string {
	// /studio/templates/{templateId}/apply
	parts := strings.Split(strings.TrimPrefix(path, "/"), "/")
	if len(parts) >= 3 {
		return parts[2]
	}
	return ""
}

type applyTemplateRequest struct {
	ConnectionID string `json:"connection_id"`
	Name         string `json:"name,omitempty"`
}

// applyTemplate creates a new dashboard from a template.
func applyTemplate(ctx context.Context, event events.APIGatewayV2HTTPRequest, authCtx *apitypes.AuthContext) (events.APIGatewayV2HTTPResponse, error) {
	templateID := getTemplateID(event.RequestContext.HTTP.Path)
	if templateID == "" {
		return authMiddleware.CreateErrorResponse(400, "Template ID is required"), nil
	}

	tpl := GetTemplateByID(templateID)
	if tpl == nil {
		return authMiddleware.CreateErrorResponse(404, "Template not found"), nil
	}

	var req applyTemplateRequest
	if err := json.Unmarshal([]byte(apiutil.GetBody(event)), &req); err != nil {
		return authMiddleware.CreateErrorResponse(400, "Invalid request body"), nil
	}

	if req.ConnectionID == "" {
		return authMiddleware.CreateErrorResponse(400, "connection_id is required"), nil
	}

	// Copy template widgets and inject connection_id
	widgets := make([]apitypes.DashboardWidget, len(tpl.Widgets))
	for i, w := range tpl.Widgets {
		widgets[i] = w
		widgets[i].WidgetID = "wdg:" + uuid.Must(uuid.NewV7()).String()
		widgets[i].ConnectionID = req.ConnectionID
	}

	dashboardName := req.Name
	if dashboardName == "" {
		dashboardName = tpl.Name
	}

	now := time.Now().UTC().Format(time.RFC3339)
	id := "dash:" + uuid.Must(uuid.NewV7()).String()

	dashboard := apitypes.Dashboard{
		DashboardID:  id,
		AccountID:    authCtx.AccountID,
		Name:         dashboardName,
		Description:  tpl.Description,
		Widgets:      widgets,
		TemplateID:   tpl.ID,
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
		log.Printf("Failed to create dashboard from template: %v", err)
		return authMiddleware.CreateErrorResponse(500, "Failed to create dashboard"), nil
	}

	return authMiddleware.CreateSuccessResponse(201, "Dashboard created from template", dashboard), nil
}
