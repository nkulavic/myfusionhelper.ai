package verifyemail

import (
	"context"
	"encoding/json"
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
	ddbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/myfusionhelper/api/internal/apiutil"
	"github.com/myfusionhelper/api/internal/database"
	authMiddleware "github.com/myfusionhelper/api/internal/middleware/auth"
)

type VerifyEmailRequest struct {
	Email string `json:"email"`
	Code  string `json:"code"`
}

var (
	cognitoUserPoolID       = os.Getenv("COGNITO_USER_POOL_ID")
	usersTable              = os.Getenv("USERS_TABLE")
	emailVerificationsTable = os.Getenv("EMAIL_VERIFICATIONS_TABLE")
)

// Handle verifies a user's email with a 6-digit code (public, no auth required)
func Handle(ctx context.Context, event events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	log.Printf("VerifyEmail handler called")

	body := apiutil.GetBody(event)
	if body == "" {
		return authMiddleware.CreateErrorResponse(400, "Request body is required"), nil
	}

	var req VerifyEmailRequest
	if err := json.Unmarshal([]byte(body), &req); err != nil {
		return authMiddleware.CreateErrorResponse(400, "Invalid JSON format"), nil
	}

	if req.Email == "" {
		return authMiddleware.CreateErrorResponse(400, "Email is required"), nil
	}
	if req.Code == "" {
		return authMiddleware.CreateErrorResponse(400, "Verification code is required"), nil
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

	// Update Cognito email_verified attribute to true
	cognitoClient := cognitoidentityprovider.NewFromConfig(cfg)
	_, err = cognitoClient.AdminUpdateUserAttributes(ctx, &cognitoidentityprovider.AdminUpdateUserAttributesInput{
		UserPoolId: aws.String(cognitoUserPoolID),
		Username:   aws.String(email),
		UserAttributes: []cognitotypes.AttributeType{
			{Name: aws.String("email_verified"), Value: aws.String("true")},
		},
	})
	if err != nil {
		log.Printf("Failed to update Cognito email_verified for %s: %v", email, err)
		return authMiddleware.CreateErrorResponse(500, "Verification failed"), nil
	}

	// Update user record in DynamoDB: set email_verified=true
	if usersTable != "" {
		usersRepo := database.NewUsersRepository(db, usersTable)
		user, err := usersRepo.GetByEmail(ctx, email)
		if err != nil {
			log.Printf("Failed to look up user for %s: %v", email, err)
		}
		if user != nil {
			_, updateErr := db.UpdateItem(ctx, &dynamodb.UpdateItemInput{
				TableName: aws.String(usersTable),
				Key: map[string]ddbtypes.AttributeValue{
					"user_id": &ddbtypes.AttributeValueMemberS{Value: user.UserID},
				},
				UpdateExpression: aws.String("SET email_verified = :verified, updated_at = :updated_at"),
				ExpressionAttributeValues: map[string]ddbtypes.AttributeValue{
					":verified":   &ddbtypes.AttributeValueMemberBOOL{Value: true},
					":updated_at": &ddbtypes.AttributeValueMemberS{Value: time.Now().UTC().Format(time.RFC3339)},
				},
			})
			if updateErr != nil {
				log.Printf("Failed to update user email_verified for %s: %v", email, updateErr)
			}
		}
	}

	// Mark the verification code as verified
	if err := verificationsRepo.MarkAsVerified(ctx, matchedVerificationID); err != nil {
		log.Printf("Failed to mark verification %s as verified: %v", matchedVerificationID, err)
	}

	log.Printf("Email verification successful for %s", email)
	return authMiddleware.CreateSuccessResponse(200, "Email verified successfully", nil), nil
}
