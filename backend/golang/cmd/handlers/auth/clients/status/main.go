package status

import (
	"context"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	ddbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	authMiddleware "github.com/myfusionhelper/api/internal/middleware/auth"
	apitypes "github.com/myfusionhelper/api/internal/types"
)

// HandleWithAuth is the status handler (requires auth)
func HandleWithAuth(ctx context.Context, event events.APIGatewayV2HTTPRequest, authCtx *apitypes.AuthContext) (events.APIGatewayV2HTTPResponse, error) {
	log.Printf("Status handler called for user: %s", authCtx.UserID)

	region := os.Getenv("COGNITO_REGION")
	if region == "" {
		region = "us-west-2"
	}

	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return authMiddleware.CreateErrorResponse(500, "Internal server error"), nil
	}

	db := dynamodb.NewFromConfig(cfg)

	// Get full user details
	userResult, err := db.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(os.Getenv("USERS_TABLE")),
		Key: map[string]ddbtypes.AttributeValue{
			"user_id": &ddbtypes.AttributeValueMemberS{Value: authCtx.UserID},
		},
	})
	if err != nil || userResult.Item == nil {
		return authMiddleware.CreateErrorResponse(404, "User not found"), nil
	}

	var user struct {
		UserID             string `json:"user_id" dynamodbav:"user_id"`
		Email              string `json:"email" dynamodbav:"email"`
		Name               string `json:"name" dynamodbav:"name"`
		Status             string `json:"status" dynamodbav:"status"`
		CurrentAccountID   string `json:"current_account_id" dynamodbav:"current_account_id"`
		OnboardingComplete bool   `json:"onboarding_complete" dynamodbav:"onboarding_complete"`
		CreatedAt          string `json:"created_at" dynamodbav:"created_at"`
		UpdatedAt          string `json:"updated_at" dynamodbav:"updated_at"`
	}
	if err := attributevalue.UnmarshalMap(userResult.Item, &user); err != nil {
		return authMiddleware.CreateErrorResponse(500, "Internal server error"), nil
	}

	// Build account context
	accountContext := map[string]interface{}{
		"account_id":  authCtx.AccountID,
		"role":        authCtx.Role,
		"permissions": authCtx.Permissions,
	}

	// Build available accounts list
	availableAccounts := make([]map[string]interface{}, len(authCtx.AvailableAccounts))
	for i, aa := range authCtx.AvailableAccounts {
		availableAccounts[i] = map[string]interface{}{
			"account_id":   aa.AccountID,
			"account_name": aa.AccountName,
			"role":         aa.Role,
			"permissions":  aa.Permissions,
			"is_current":   aa.IsCurrent,
		}
	}

	return authMiddleware.CreateSuccessResponse(200, "Status retrieved successfully", map[string]interface{}{
		"user_id":              user.UserID,
		"email":                user.Email,
		"name":                 user.Name,
		"status":               user.Status,
		"current_account_id":   user.CurrentAccountID,
		"onboarding_complete":  user.OnboardingComplete,
		"created_at":           user.CreatedAt,
		"updated_at":           user.UpdatedAt,
		"account_context":      accountContext,
		"available_accounts":   availableAccounts,
	}), nil
}
