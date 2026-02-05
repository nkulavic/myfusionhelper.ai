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
	apitypes "github.com/myfusionhelper/api/internal/types"
)

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

var clientID = os.Getenv("COGNITO_CLIENT_ID")

// HandleWithAuth is the token refresh handler (requires auth)
func HandleWithAuth(ctx context.Context, event events.APIGatewayV2HTTPRequest, authCtx *apitypes.AuthContext) (events.APIGatewayV2HTTPResponse, error) {
	log.Printf("Refresh handler called for user: %s", authCtx.UserID)

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
		return authMiddleware.CreateErrorResponse(400, "Failed to refresh token"), nil
	}

	var accessToken, idToken string
	if result.AuthenticationResult != nil {
		if result.AuthenticationResult.AccessToken != nil {
			accessToken = *result.AuthenticationResult.AccessToken
		}
		if result.AuthenticationResult.IdToken != nil {
			idToken = *result.AuthenticationResult.IdToken
		}
	}

	return authMiddleware.CreateSuccessResponse(200, "Token refreshed successfully", map[string]interface{}{
		"user_id":      authCtx.UserID,
		"access_token": accessToken,
		"id_token":     idToken,
		"token_type":   "Bearer",
	}), nil
}
