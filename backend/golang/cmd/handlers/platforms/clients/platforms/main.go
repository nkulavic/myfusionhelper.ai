package platforms

import (
	"context"
	"log"
	"os"
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

var platformsTable = os.Getenv("PLATFORMS_TABLE")
var connectionsTable = os.Getenv("CONNECTIONS_TABLE")

// HandleWithAuth handles platform list and get operations
func HandleWithAuth(ctx context.Context, event events.APIGatewayV2HTTPRequest, authCtx *apitypes.AuthContext) (events.APIGatewayV2HTTPResponse, error) {
	switch event.RequestContext.HTTP.Method {
	case "GET":
		if event.RequestContext.HTTP.Path == "/platforms" {
			return listPlatforms(ctx, event, authCtx)
		}
		return getPlatform(ctx, event, authCtx)
	default:
		return authMiddleware.CreateErrorResponse(405, "Method not allowed"), nil
	}
}

func listPlatforms(ctx context.Context, event events.APIGatewayV2HTTPRequest, authCtx *apitypes.AuthContext) (events.APIGatewayV2HTTPResponse, error) {
	log.Printf("List platforms for account: %s", authCtx.AccountID)

	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(os.Getenv("COGNITO_REGION")))
	if err != nil {
		return authMiddleware.CreateErrorResponse(500, "Internal server error"), nil
	}
	db := dynamodb.NewFromConfig(cfg)

	// Build scan with optional filters
	scanInput := &dynamodb.ScanInput{
		TableName: aws.String(platformsTable),
	}

	queryParams := event.QueryStringParameters
	var filterExpressions []string
	expressionValues := map[string]ddbtypes.AttributeValue{}
	expressionNames := map[string]string{}

	if category, ok := queryParams["category"]; ok && category != "" {
		filterExpressions = append(filterExpressions, "category = :category")
		expressionValues[":category"] = &ddbtypes.AttributeValueMemberS{Value: category}
	}
	if status, ok := queryParams["status"]; ok && status != "" {
		filterExpressions = append(filterExpressions, "#s = :status")
		expressionValues[":status"] = &ddbtypes.AttributeValueMemberS{Value: status}
		expressionNames["#s"] = "status"
	}

	if len(filterExpressions) > 0 {
		expr := strings.Join(filterExpressions, " AND ")
		scanInput.FilterExpression = aws.String(expr)
		scanInput.ExpressionAttributeValues = expressionValues
		if len(expressionNames) > 0 {
			scanInput.ExpressionAttributeNames = expressionNames
		}
	}

	result, err := db.Scan(ctx, scanInput)
	if err != nil {
		log.Printf("Failed to scan platforms: %v", err)
		return authMiddleware.CreateErrorResponse(500, "Failed to list platforms"), nil
	}

	var platforms []apitypes.Platform
	if err := attributevalue.UnmarshalListOfMaps(result.Items, &platforms); err != nil {
		log.Printf("Failed to unmarshal platforms: %v", err)
		return authMiddleware.CreateErrorResponse(500, "Internal server error"), nil
	}

	return authMiddleware.CreateSuccessResponse(200, "Platforms listed successfully", map[string]interface{}{
		"platforms": platforms,
		"total":     len(platforms),
	}), nil
}

func getPlatform(ctx context.Context, event events.APIGatewayV2HTTPRequest, authCtx *apitypes.AuthContext) (events.APIGatewayV2HTTPResponse, error) {
	platformIDOrSlug := event.PathParameters["platform_id"]
	if platformIDOrSlug == "" {
		return authMiddleware.CreateErrorResponse(400, "Platform ID is required"), nil
	}

	log.Printf("Get platform %s for account: %s", platformIDOrSlug, authCtx.AccountID)

	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(os.Getenv("COGNITO_REGION")))
	if err != nil {
		return authMiddleware.CreateErrorResponse(500, "Internal server error"), nil
	}
	db := dynamodb.NewFromConfig(cfg)

	platform, err := resolvePlatform(ctx, db, platformIDOrSlug)
	if err != nil {
		log.Printf("Failed to resolve platform %s: %v", platformIDOrSlug, err)
		return authMiddleware.CreateErrorResponse(404, "Platform not found"), nil
	}

	// Check if user has any connections to this platform
	hasConnection := false
	if connectionsTable != "" {
		connResult, connErr := db.Query(ctx, &dynamodb.QueryInput{
			TableName:              aws.String(connectionsTable),
			IndexName:              aws.String("AccountIdIndex"),
			KeyConditionExpression: aws.String("account_id = :account_id"),
			FilterExpression:       aws.String("platform_id = :platform_id"),
			ExpressionAttributeValues: map[string]ddbtypes.AttributeValue{
				":account_id":  &ddbtypes.AttributeValueMemberS{Value: authCtx.AccountID},
				":platform_id": &ddbtypes.AttributeValueMemberS{Value: platform.PlatformID},
			},
		})
		if connErr == nil && connResult.Count > 0 {
			hasConnection = true
		}
	}

	return authMiddleware.CreateSuccessResponse(200, "Platform retrieved successfully", map[string]interface{}{
		"platform":       platform,
		"has_connection": hasConnection,
	}), nil
}

// resolvePlatform looks up a platform by ID or slug
func resolvePlatform(ctx context.Context, db *dynamodb.Client, platformIDOrSlug string) (*apitypes.Platform, error) {
	// Try by platform_id first
	result, err := db.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(platformsTable),
		Key: map[string]ddbtypes.AttributeValue{
			"platform_id": &ddbtypes.AttributeValueMemberS{Value: platformIDOrSlug},
		},
	})
	if err == nil && result.Item != nil {
		var platform apitypes.Platform
		if err := attributevalue.UnmarshalMap(result.Item, &platform); err != nil {
			return nil, err
		}
		return &platform, nil
	}

	// Try by slug using GSI
	slugResult, err := db.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(platformsTable),
		IndexName:              aws.String("SlugIndex"),
		KeyConditionExpression: aws.String("slug = :slug"),
		ExpressionAttributeValues: map[string]ddbtypes.AttributeValue{
			":slug": &ddbtypes.AttributeValueMemberS{Value: platformIDOrSlug},
		},
	})
	if err != nil {
		return nil, err
	}
	if len(slugResult.Items) == 0 {
		return nil, err
	}

	var platform apitypes.Platform
	if err := attributevalue.UnmarshalMap(slugResult.Items[0], &platform); err != nil {
		return nil, err
	}
	return &platform, nil
}
