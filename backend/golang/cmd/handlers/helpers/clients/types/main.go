package types

import (
	"context"
	"encoding/json"
	"log"
	"sort"
	"strings"

	"github.com/aws/aws-lambda-go/events"

	helperEngine "github.com/myfusionhelper/api/internal/helpers"
	authMiddleware "github.com/myfusionhelper/api/internal/middleware/auth"
	apitypes "github.com/myfusionhelper/api/internal/types"
)

// Ensure apitypes is used (for future use)
var _ = apitypes.Helper{}

// HandleWithAuth routes helper types requests
func HandleWithAuth(ctx context.Context, event events.APIGatewayV2HTTPRequest, authCtx *apitypes.AuthContext) (events.APIGatewayV2HTTPResponse, error) {
	path := event.RequestContext.HTTP.Path
	method := event.RequestContext.HTTP.Method

	switch {
	case path == "/helpers/types" && method == "GET":
		return listHelperTypes(ctx, event, authCtx)
	case strings.HasPrefix(path, "/helpers/types/") && method == "GET":
		return getHelperType(ctx, event, authCtx)
	default:
		return authMiddleware.CreateErrorResponse(404, "Not Found"), nil
	}
}

func listHelperTypes(ctx context.Context, event events.APIGatewayV2HTTPRequest, authCtx *apitypes.AuthContext) (events.APIGatewayV2HTTPResponse, error) {
	log.Printf("List helper types for account: %s", authCtx.AccountID)

	infos := helperEngine.ListHelperInfo()

	// Sort by category then name for consistent ordering
	sort.Slice(infos, func(i, j int) bool {
		if infos[i].Category != infos[j].Category {
			return infos[i].Category < infos[j].Category
		}
		return infos[i].Name < infos[j].Name
	})

	// Collect unique categories
	categorySet := make(map[string]bool)
	for _, info := range infos {
		categorySet[info.Category] = true
	}
	categories := make([]string, 0, len(categorySet))
	for cat := range categorySet {
		categories = append(categories, cat)
	}
	sort.Strings(categories)

	// Build response items
	typeItems := make([]map[string]interface{}, 0, len(infos))
	for _, info := range infos {
		item := map[string]interface{}{
			"type":           info.Type,
			"name":           info.Name,
			"category":       info.Category,
			"description":    info.Description,
			"requires_crm":   info.RequiresCRM,
			"supported_crms": info.SupportedCRMs,
			"config_schema":  info.ConfigSchema,
		}
		typeItems = append(typeItems, item)
	}

	return authMiddleware.CreateSuccessResponse(200, "Helper types retrieved successfully", map[string]interface{}{
		"types":       typeItems,
		"total_count": len(typeItems),
		"categories":  categories,
	}), nil
}

func getHelperType(ctx context.Context, event events.APIGatewayV2HTTPRequest, authCtx *apitypes.AuthContext) (events.APIGatewayV2HTTPResponse, error) {
	// Extract type from path: /helpers/types/{type}
	path := event.RequestContext.HTTP.Path
	parts := strings.Split(strings.TrimPrefix(path, "/helpers/types/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		return authMiddleware.CreateErrorResponse(400, "Helper type is required"), nil
	}
	helperType := parts[0]

	log.Printf("Get helper type %s for account: %s", helperType, authCtx.AccountID)

	if !helperEngine.IsRegistered(helperType) {
		return authMiddleware.CreateErrorResponse(404, "Helper type not found"), nil
	}

	helper, err := helperEngine.NewHelper(helperType)
	if err != nil {
		return authMiddleware.CreateErrorResponse(500, "Failed to load helper type"), nil
	}

	item := map[string]interface{}{
		"type":           helperType,
		"name":           helper.GetName(),
		"category":       helper.GetCategory(),
		"description":    helper.GetDescription(),
		"requires_crm":   helper.RequiresCRM(),
		"supported_crms": helper.SupportedCRMs(),
		"config_schema":  helper.GetConfigSchema(),
	}

	responseBody, _ := json.Marshal(map[string]interface{}{
		"success": true,
		"message": "Helper type retrieved successfully",
		"data":    item,
	})

	return events.APIGatewayV2HTTPResponse{
		StatusCode: 200,
		Headers: map[string]string{
			"Content-Type":                 "application/json",
			"Access-Control-Allow-Origin":  "*",
			"Access-Control-Allow-Methods": "GET, OPTIONS",
			"Access-Control-Allow-Headers": "Content-Type, Authorization",
		},
		Body: string(responseBody),
	}, nil
}
