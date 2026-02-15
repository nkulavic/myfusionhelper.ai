package resetpassword

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	cognitotypes "github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider/types"
	"github.com/myfusionhelper/api/internal/apiutil"
	authMiddleware "github.com/myfusionhelper/api/internal/middleware/auth"
)

type ResetPasswordRequest struct {
	Email       string `json:"email"`
	Code        string `json:"code"`
	NewPassword string `json:"new_password"`
}

var (
	cognitoClientID = os.Getenv("COGNITO_CLIENT_ID")
)

// Handle confirms a password reset with the code from email (public, no auth required)
func Handle(ctx context.Context, event events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	log.Printf("ResetPassword handler called")

	if cognitoClientID == "" {
		log.Printf("ERROR: Missing Cognito configuration")
		return authMiddleware.CreateErrorResponse(500, "Authentication service not configured"), nil
	}

	body := apiutil.GetBody(event)
	if body == "" {
		return authMiddleware.CreateErrorResponse(400, "Request body is required"), nil
	}

	var req ResetPasswordRequest
	if err := json.Unmarshal([]byte(body), &req); err != nil {
		return authMiddleware.CreateErrorResponse(400, "Invalid JSON format"), nil
	}

	if req.Email == "" {
		return authMiddleware.CreateErrorResponse(400, "Email is required"), nil
	}
	if req.Code == "" {
		return authMiddleware.CreateErrorResponse(400, "Verification code is required"), nil
	}
	if req.NewPassword == "" {
		return authMiddleware.CreateErrorResponse(400, "New password is required"), nil
	}

	region := os.Getenv("COGNITO_REGION")
	if region == "" {
		region = "us-west-2"
	}

	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		log.Printf("Failed to load AWS config: %v", err)
		return authMiddleware.CreateErrorResponse(500, "Service unavailable"), nil
	}

	cognitoClient := cognitoidentityprovider.NewFromConfig(cfg)

	_, err = cognitoClient.ConfirmForgotPassword(ctx, &cognitoidentityprovider.ConfirmForgotPasswordInput{
		ClientId:         aws.String(cognitoClientID),
		Username:         aws.String(req.Email),
		ConfirmationCode: aws.String(req.Code),
		Password:         aws.String(req.NewPassword),
	})
	if err != nil {
		log.Printf("ConfirmForgotPassword failed: %v", err)
		return handleResetPasswordError(err), nil
	}

	return authMiddleware.CreateSuccessResponse(200, "Password reset successfully", nil), nil
}

func handleResetPasswordError(err error) events.APIGatewayV2HTTPResponse {
	if err == nil {
		return authMiddleware.CreateErrorResponse(500, "Unknown error")
	}

	var codeMismatch *cognitotypes.CodeMismatchException
	var expiredCode *cognitotypes.ExpiredCodeException
	var invalidPassword *cognitotypes.InvalidPasswordException
	var tooMany *cognitotypes.TooManyRequestsException
	var limitExceeded *cognitotypes.LimitExceededException
	var userNotFound *cognitotypes.UserNotFoundException

	switch {
	case errors.As(err, &codeMismatch):
		return authMiddleware.CreateErrorResponse(400, "Invalid verification code")
	case errors.As(err, &expiredCode):
		return authMiddleware.CreateErrorResponse(400, "Verification code has expired. Please request a new one.")
	case errors.As(err, &invalidPassword):
		return authMiddleware.CreateErrorResponse(400, "Password does not meet requirements. Use at least 8 characters with uppercase, lowercase, numbers, and symbols.")
	case errors.As(err, &tooMany):
		return authMiddleware.CreateErrorResponse(429, "Too many requests. Please try again later.")
	case errors.As(err, &limitExceeded):
		return authMiddleware.CreateErrorResponse(429, "Too many requests. Please try again later.")
	case errors.As(err, &userNotFound):
		return authMiddleware.CreateErrorResponse(400, "Invalid verification code")
	default:
		return authMiddleware.CreateErrorResponse(500, "Password reset failed")
	}
}
