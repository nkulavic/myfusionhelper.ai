package onboarding

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	ddbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	authMiddleware "github.com/myfusionhelper/api/internal/middleware/auth"
	apitypes "github.com/myfusionhelper/api/internal/types"
)

// HandleWithAuth handles PATCH /auth/onboarding-complete
func HandleWithAuth(ctx context.Context, event events.APIGatewayV2HTTPRequest, authCtx *apitypes.AuthContext) (events.APIGatewayV2HTTPResponse, error) {
	log.Printf("Onboarding complete for user: %s", authCtx.UserID)

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

	_, err = db.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(usersTable),
		Key: map[string]ddbtypes.AttributeValue{
			"user_id": &ddbtypes.AttributeValueMemberS{Value: authCtx.UserID},
		},
		UpdateExpression: aws.String("SET onboarding_complete = :val, updated_at = :updated_at"),
		ExpressionAttributeValues: map[string]ddbtypes.AttributeValue{
			":val":        &ddbtypes.AttributeValueMemberBOOL{Value: true},
			":updated_at": &ddbtypes.AttributeValueMemberS{Value: time.Now().UTC().Format(time.RFC3339)},
		},
	})
	if err != nil {
		log.Printf("Failed to update onboarding status: %v", err)
		return authMiddleware.CreateErrorResponse(500, "Failed to update onboarding status"), nil
	}

	return authMiddleware.CreateSuccessResponse(200, "Onboarding completed", map[string]interface{}{
		"onboarding_complete": true,
	}), nil
}
