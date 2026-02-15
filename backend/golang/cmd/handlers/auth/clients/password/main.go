package password

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"os"
	"strings"
	"unicode"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	cognitotypes "github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider/types"
	"github.com/myfusionhelper/api/internal/apiutil"
	authMiddleware "github.com/myfusionhelper/api/internal/middleware/auth"
	apitypes "github.com/myfusionhelper/api/internal/types"
)

type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

// HandleWithAuth handles password change requests (requires authentication)
func HandleWithAuth(ctx context.Context, event events.APIGatewayV2HTTPRequest, authCtx *apitypes.AuthContext) (events.APIGatewayV2HTTPResponse, error) {
	if event.RequestContext.HTTP.Method != "PUT" {
		return authMiddleware.CreateErrorResponse(405, "Method not allowed"), nil
	}

	log.Printf("Password change requested for user: %s", authCtx.UserID)

	var req ChangePasswordRequest
	if err := json.Unmarshal([]byte(apiutil.GetBody(event)), &req); err != nil {
		return authMiddleware.CreateErrorResponse(400, "Invalid request format"), nil
	}

	if req.CurrentPassword == "" {
		return authMiddleware.CreateErrorResponse(400, "Current password is required"), nil
	}
	if req.NewPassword == "" {
		return authMiddleware.CreateErrorResponse(400, "New password is required"), nil
	}
	if req.CurrentPassword == req.NewPassword {
		return authMiddleware.CreateErrorResponse(400, "New password must be different from current password"), nil
	}

	// Validate new password meets Cognito policy
	if err := validatePassword(req.NewPassword); err != nil {
		return authMiddleware.CreateErrorResponse(400, err.Error()), nil
	}

	// Extract the access token from the Authorization header
	accessToken := extractAccessToken(event)
	if accessToken == "" {
		return authMiddleware.CreateErrorResponse(401, "Missing access token"), nil
	}

	region := os.Getenv("COGNITO_REGION")
	if region == "" {
		region = "us-west-2"
	}

	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		log.Printf("Failed to load AWS config: %v", err)
		return authMiddleware.CreateErrorResponse(500, "Internal error"), nil
	}

	cognitoClient := cognitoidentityprovider.NewFromConfig(cfg)

	_, err = cognitoClient.ChangePassword(ctx, &cognitoidentityprovider.ChangePasswordInput{
		AccessToken:      aws.String(accessToken),
		PreviousPassword: aws.String(req.CurrentPassword),
		ProposedPassword: aws.String(req.NewPassword),
	})
	if err != nil {
		log.Printf("Password change failed for user %s: %v", authCtx.UserID, err)
		return handlePasswordError(err), nil
	}

	log.Printf("Password changed successfully for user: %s", authCtx.UserID)
	return authMiddleware.CreateSuccessResponse(200, "Password updated successfully", nil), nil
}

func extractAccessToken(event events.APIGatewayV2HTTPRequest) string {
	authHeader := event.Headers["Authorization"]
	if authHeader == "" {
		authHeader = event.Headers["authorization"]
	}
	if authHeader == "" {
		return ""
	}
	return strings.TrimPrefix(authHeader, "Bearer ")
}

func validatePassword(password string) error {
	if len(password) < 8 {
		return errors.New("Password must be at least 8 characters")
	}

	var hasUpper, hasLower, hasDigit bool
	for _, ch := range password {
		switch {
		case unicode.IsUpper(ch):
			hasUpper = true
		case unicode.IsLower(ch):
			hasLower = true
		case unicode.IsDigit(ch):
			hasDigit = true
		}
	}

	if !hasUpper {
		return errors.New("Password must contain at least one uppercase letter")
	}
	if !hasLower {
		return errors.New("Password must contain at least one lowercase letter")
	}
	if !hasDigit {
		return errors.New("Password must contain at least one number")
	}

	return nil
}

func handlePasswordError(err error) events.APIGatewayV2HTTPResponse {
	var notAuth *cognitotypes.NotAuthorizedException
	var invalidPassword *cognitotypes.InvalidPasswordException
	var limitExceeded *cognitotypes.LimitExceededException
	var tooMany *cognitotypes.TooManyRequestsException

	switch {
	case errors.As(err, &notAuth):
		return authMiddleware.CreateErrorResponse(401, "Current password is incorrect")
	case errors.As(err, &invalidPassword):
		return authMiddleware.CreateErrorResponse(400, "New password does not meet requirements")
	case errors.As(err, &limitExceeded):
		return authMiddleware.CreateErrorResponse(429, "Too many attempts, please try again later")
	case errors.As(err, &tooMany):
		return authMiddleware.CreateErrorResponse(429, "Too many requests, please try again later")
	default:
		return authMiddleware.CreateErrorResponse(500, "Failed to change password")
	}
}
