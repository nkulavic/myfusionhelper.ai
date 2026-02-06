package preferences

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
	authMiddleware "github.com/myfusionhelper/api/internal/middleware/auth"
	apitypes "github.com/myfusionhelper/api/internal/types"
)

// HandleWithAuth routes to get or update notification preferences
func HandleWithAuth(ctx context.Context, event events.APIGatewayV2HTTPRequest, authCtx *apitypes.AuthContext) (events.APIGatewayV2HTTPResponse, error) {
	method := event.RequestContext.HTTP.Method

	switch method {
	case "GET":
		return getPreferences(ctx, authCtx)
	case "PUT":
		return updatePreferences(ctx, event, authCtx)
	default:
		return authMiddleware.CreateErrorResponse(405, "Method not allowed"), nil
	}
}

func getPreferences(ctx context.Context, authCtx *apitypes.AuthContext) (events.APIGatewayV2HTTPResponse, error) {
	log.Printf("Get notification preferences for user: %s", authCtx.UserID)

	region := os.Getenv("COGNITO_REGION")
	if region == "" {
		region = "us-west-2"
	}

	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return authMiddleware.CreateErrorResponse(500, "Internal server error"), nil
	}

	db := dynamodb.NewFromConfig(cfg)
	usersTable := os.Getenv("USERS_TABLE")

	result, err := db.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(usersTable),
		Key: map[string]ddbtypes.AttributeValue{
			"user_id": &ddbtypes.AttributeValueMemberS{Value: authCtx.UserID},
		},
		ProjectionExpression: aws.String("notification_preferences"),
	})
	if err != nil {
		log.Printf("Failed to get user preferences: %v", err)
		return authMiddleware.CreateErrorResponse(500, "Failed to get preferences"), nil
	}

	// Return defaults if no preferences are set yet
	prefs := defaultPreferences()
	if result.Item != nil {
		var user struct {
			NotificationPreferences *apitypes.NotificationPreferences `dynamodbav:"notification_preferences"`
		}
		if err := attributevalue.UnmarshalMap(result.Item, &user); err == nil && user.NotificationPreferences != nil {
			prefs = user.NotificationPreferences
		}
	}

	return authMiddleware.CreateSuccessResponse(200, "Preferences retrieved successfully", prefs), nil
}

func updatePreferences(ctx context.Context, event events.APIGatewayV2HTTPRequest, authCtx *apitypes.AuthContext) (events.APIGatewayV2HTTPResponse, error) {
	log.Printf("Update notification preferences for user: %s", authCtx.UserID)

	var prefs apitypes.NotificationPreferences
	if err := json.Unmarshal([]byte(event.Body), &prefs); err != nil {
		return authMiddleware.CreateErrorResponse(400, "Invalid request format"), nil
	}

	region := os.Getenv("COGNITO_REGION")
	if region == "" {
		region = "us-west-2"
	}

	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return authMiddleware.CreateErrorResponse(500, "Internal server error"), nil
	}

	db := dynamodb.NewFromConfig(cfg)
	usersTable := os.Getenv("USERS_TABLE")

	prefsAV, err := attributevalue.MarshalMap(prefs)
	if err != nil {
		return authMiddleware.CreateErrorResponse(500, "Internal server error"), nil
	}

	_, err = db.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(usersTable),
		Key: map[string]ddbtypes.AttributeValue{
			"user_id": &ddbtypes.AttributeValueMemberS{Value: authCtx.UserID},
		},
		UpdateExpression: aws.String("SET notification_preferences = :prefs, updated_at = :updated_at"),
		ExpressionAttributeValues: map[string]ddbtypes.AttributeValue{
			":prefs":      &ddbtypes.AttributeValueMemberM{Value: prefsAV},
			":updated_at": &ddbtypes.AttributeValueMemberS{Value: time.Now().UTC().Format(time.RFC3339)},
		},
	})
	if err != nil {
		log.Printf("Failed to update preferences: %v", err)
		return authMiddleware.CreateErrorResponse(500, "Failed to update preferences"), nil
	}

	return authMiddleware.CreateSuccessResponse(200, "Preferences updated successfully", prefs), nil
}

func defaultPreferences() *apitypes.NotificationPreferences {
	return &apitypes.NotificationPreferences{
		ExecutionFailures: true,
		ConnectionIssues:  true,
		UsageAlerts:       true,
		WeeklySummary:     false,
		NewFeatures:       true,
		TeamActivity:      false,
		RealtimeStatus:    false,
		AiInsights:        true,
		SystemMaintenance: true,
	}
}
