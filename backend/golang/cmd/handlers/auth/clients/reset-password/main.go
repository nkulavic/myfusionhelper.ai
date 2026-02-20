package resetpassword

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	cognitotypes "github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider/types"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/myfusionhelper/api/internal/apiutil"
	"github.com/myfusionhelper/api/internal/database"
	authMiddleware "github.com/myfusionhelper/api/internal/middleware/auth"
)

type ResetPasswordRequest struct {
	Email       string `json:"email"`
	Code        string `json:"code"`
	NewPassword string `json:"new_password"`
}

var (
	cognitoUserPoolID       = os.Getenv("COGNITO_USER_POOL_ID")
	emailVerificationsTable = os.Getenv("EMAIL_VERIFICATIONS_TABLE")
)

// Handle confirms a password reset with a code from the branded email (public, no auth required)
func Handle(ctx context.Context, event events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	log.Printf("ResetPassword handler called")

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

	email := strings.ToLower(req.Email)

	region := os.Getenv("COGNITO_REGION")
	if region == "" {
		region = "us-west-2"
	}

	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		log.Printf("Failed to load AWS config: %v", err)
		return authMiddleware.CreateErrorResponse(500, "Service unavailable"), nil
	}

	db := dynamodb.NewFromConfig(cfg)
	verificationsRepo := database.NewEmailVerificationsRepository(db, emailVerificationsTable)

	// Look up pending verification codes for this email
	pendingCodes, err := verificationsRepo.GetPendingByEmail(ctx, email)
	if err != nil {
		log.Printf("Failed to look up verification codes for %s: %v", email, err)
		return authMiddleware.CreateErrorResponse(500, "Service unavailable"), nil
	}

	// Find matching code
	var matchedVerificationID string
	now := time.Now().Unix()
	for _, v := range pendingCodes {
		if v.Token == req.Code && v.ExpiresAt > now {
			matchedVerificationID = v.VerificationID
			break
		}
	}

	if matchedVerificationID == "" {
		log.Printf("No matching verification code for %s", email)
		return authMiddleware.CreateErrorResponse(400, "Invalid or expired verification code. Please request a new one."), nil
	}

	// Set the new password via Cognito AdminSetUserPassword
	cognitoClient := cognitoidentityprovider.NewFromConfig(cfg)
	_, err = cognitoClient.AdminSetUserPassword(ctx, &cognitoidentityprovider.AdminSetUserPasswordInput{
		UserPoolId: aws.String(cognitoUserPoolID),
		Username:   aws.String(email),
		Password:   aws.String(req.NewPassword),
		Permanent:  true,
	})
	if err != nil {
		log.Printf("AdminSetUserPassword failed for %s: %v", email, err)
		return handleSetPasswordError(err), nil
	}

	// Mark the verification as used
	if err := verificationsRepo.MarkAsVerified(ctx, matchedVerificationID); err != nil {
		log.Printf("Failed to mark verification %s as verified: %v", matchedVerificationID, err)
		// Non-fatal: password was already changed successfully
	}

	log.Printf("Password reset successful for %s", email)
	return authMiddleware.CreateSuccessResponse(200, "Password reset successfully", nil), nil
}

func handleSetPasswordError(err error) events.APIGatewayV2HTTPResponse {
	if err == nil {
		return authMiddleware.CreateErrorResponse(500, "Unknown error")
	}

	var invalidPassword *cognitotypes.InvalidPasswordException
	var tooMany *cognitotypes.TooManyRequestsException
	var limitExceeded *cognitotypes.LimitExceededException
	var userNotFound *cognitotypes.UserNotFoundException

	switch {
	case errors.As(err, &invalidPassword):
		return authMiddleware.CreateErrorResponse(400, "Password does not meet requirements. Use at least 8 characters with uppercase, lowercase, numbers, and symbols.")
	case errors.As(err, &tooMany):
		return authMiddleware.CreateErrorResponse(429, "Too many requests. Please try again later.")
	case errors.As(err, &limitExceeded):
		return authMiddleware.CreateErrorResponse(429, "Too many requests. Please try again later.")
	case errors.As(err, &userNotFound):
		return authMiddleware.CreateErrorResponse(400, "Invalid or expired verification code. Please request a new one.")
	default:
		return authMiddleware.CreateErrorResponse(500, "Password reset failed")
	}
}
