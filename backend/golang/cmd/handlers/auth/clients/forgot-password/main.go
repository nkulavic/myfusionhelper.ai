package forgotpassword

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
	authMiddleware "github.com/myfusionhelper/api/internal/middleware/auth"
	"github.com/myfusionhelper/api/internal/notifications"
)

type ForgotPasswordRequest struct {
	Email string `json:"email"`
}

var (
	cognitoClientID = os.Getenv("COGNITO_CLIENT_ID")
)

// Handle triggers Cognito ForgotPassword flow (public, no auth required)
func Handle(ctx context.Context, event events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	log.Printf("ForgotPassword handler called")

	if cognitoClientID == "" {
		log.Printf("ERROR: Missing Cognito configuration")
		return authMiddleware.CreateErrorResponse(500, "Authentication service not configured"), nil
	}

	if event.Body == "" {
		return authMiddleware.CreateErrorResponse(400, "Request body is required"), nil
	}

	var req ForgotPasswordRequest
	if err := json.Unmarshal([]byte(event.Body), &req); err != nil {
		return authMiddleware.CreateErrorResponse(400, "Invalid JSON format"), nil
	}

	if req.Email == "" {
		return authMiddleware.CreateErrorResponse(400, "Email is required"), nil
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

	_, err = cognitoClient.ForgotPassword(ctx, &cognitoidentityprovider.ForgotPasswordInput{
		ClientId: aws.String(cognitoClientID),
		Username: aws.String(req.Email),
	})
	if err != nil {
		log.Printf("ForgotPassword failed: %v", err)
		return handleForgotPasswordError(err), nil
	}

	// Send branded notification email asynchronously (Cognito will also send the actual reset code email)
	go func() {
		notifSvc, err := notifications.New(ctx)
		if err != nil {
			log.Printf("Failed to create notification service for password reset notification: %v", err)
			return
		}
		// Note: We're not sending the actual code here - Cognito handles that securely
		// This is just a branded notification confirming the reset request
		if err := notifSvc.SendPasswordResetEmail(ctx, req.Email, ""); err != nil {
			log.Printf("Failed to send password reset notification email: %v", err)
		}
	}()

	// Always return success to prevent email enumeration
	return authMiddleware.CreateSuccessResponse(200, "If an account exists with this email, a reset code has been sent.", nil), nil
}

func handleForgotPasswordError(err error) events.APIGatewayV2HTTPResponse {
	if err == nil {
		return authMiddleware.CreateErrorResponse(500, "Unknown error")
	}

	var tooMany *cognitotypes.TooManyRequestsException
	var limitExceeded *cognitotypes.LimitExceededException

	switch {
	case errors.As(err, &tooMany):
		return authMiddleware.CreateErrorResponse(429, "Too many requests. Please try again later.")
	case errors.As(err, &limitExceeded):
		return authMiddleware.CreateErrorResponse(429, "Too many requests. Please try again later.")
	default:
		// Return success even for user-not-found to prevent email enumeration
		return authMiddleware.CreateSuccessResponse(200, "If an account exists with this email, a reset code has been sent.", nil)
	}
}
