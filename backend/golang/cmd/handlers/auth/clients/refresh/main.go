package refresh

import (
	"context"
	"encoding/json"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	cognitotypes "github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider/types"
	authMiddleware "github.com/myfusionhelper/api/internal/middleware/auth"
)

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

var clientID = os.Getenv("COGNITO_CLIENT_ID")

// Handle is the token refresh handler (public, no auth required)
func Handle(ctx context.Context, event events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	log.Printf("Refresh handler called")

	var req RefreshRequest
	if err := json.Unmarshal([]byte(event.Body), &req); err != nil {
		return authMiddleware.CreateErrorResponse(400, "Invalid request format"), nil
	}

	if req.RefreshToken == "" {
		return authMiddleware.CreateErrorResponse(400, "Refresh token is required"), nil
	}

	region := os.Getenv("COGNITO_REGION")
	if region == "" {
		region = "us-west-2"
	}

	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return authMiddleware.CreateErrorResponse(500, "Failed to refresh token"), nil
	}

	cognitoClient := cognitoidentityprovider.NewFromConfig(cfg)

	result, err := cognitoClient.InitiateAuth(ctx, &cognitoidentityprovider.InitiateAuthInput{
		AuthFlow: cognitotypes.AuthFlowTypeRefreshTokenAuth,
		ClientId: aws.String(clientID),
		AuthParameters: map[string]string{
			"REFRESH_TOKEN": req.RefreshToken,
		},
	})
	if err != nil {
		log.Printf("Failed to refresh token: %v", err)
		return authMiddleware.CreateErrorResponse(401, "Invalid or expired refresh token"), nil
	}

	if result.AuthenticationResult == nil {
		return authMiddleware.CreateErrorResponse(500, "Failed to refresh token"), nil
	}

	response := map[string]interface{}{
		"token":      *result.AuthenticationResult.AccessToken,
		"token_type": "Bearer",
	}

	// Cognito refresh does not return a new refresh token, keep the existing one
	if result.AuthenticationResult.RefreshToken != nil {
		response["refresh_token"] = *result.AuthenticationResult.RefreshToken
	}

	return authMiddleware.CreateSuccessResponse(200, "Token refreshed successfully", response), nil
}
