package logout

import (
	"context"
	"log"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	authMiddleware "github.com/myfusionhelper/api/internal/middleware/auth"
	apitypes "github.com/myfusionhelper/api/internal/types"
)

// HandleWithAuth is the logout handler (requires auth)
func HandleWithAuth(ctx context.Context, event events.APIGatewayV2HTTPRequest, authCtx *apitypes.AuthContext) (events.APIGatewayV2HTTPResponse, error) {
	log.Printf("Logout handler called for user: %s", authCtx.UserID)

	// Extract access token from Authorization header
	authHeader := event.Headers["Authorization"]
	if authHeader == "" {
		authHeader = event.Headers["authorization"]
	}
	accessToken := strings.TrimPrefix(authHeader, "Bearer ")

	if accessToken == "" || accessToken == authHeader {
		return authMiddleware.CreateErrorResponse(400, "Access token is required"), nil
	}

	region := os.Getenv("COGNITO_REGION")
	if region == "" {
		region = "us-west-2"
	}

	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return authMiddleware.CreateErrorResponse(500, "Logout failed"), nil
	}

	cognitoClient := cognitoidentityprovider.NewFromConfig(cfg)

	_, err = cognitoClient.GlobalSignOut(ctx, &cognitoidentityprovider.GlobalSignOutInput{
		AccessToken: aws.String(accessToken),
	})
	if err != nil {
		log.Printf("Failed to sign out: %v", err)
		return authMiddleware.CreateErrorResponse(500, "Logout failed"), nil
	}

	return authMiddleware.CreateSuccessResponse(200, "Logout successful", map[string]interface{}{
		"user_id": authCtx.UserID,
	}), nil
}
