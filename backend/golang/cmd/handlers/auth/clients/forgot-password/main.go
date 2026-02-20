package forgotpassword

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"log"
	"math/big"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/google/uuid"
	"github.com/myfusionhelper/api/internal/apiutil"
	"github.com/myfusionhelper/api/internal/database"
	authMiddleware "github.com/myfusionhelper/api/internal/middleware/auth"
	"github.com/myfusionhelper/api/internal/types"
)

type ForgotPasswordRequest struct {
	Email string `json:"email"`
}

var (
	usersTable              = os.Getenv("USERS_TABLE")
	emailVerificationsTable = os.Getenv("EMAIL_VERIFICATIONS_TABLE")
)

// Handle triggers a password reset by generating a code and sending a branded email (public, no auth required)
func Handle(ctx context.Context, event events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	log.Printf("ForgotPassword handler called")

	body := apiutil.GetBody(event)
	if body == "" {
		return authMiddleware.CreateErrorResponse(400, "Request body is required"), nil
	}

	var req ForgotPasswordRequest
	if err := json.Unmarshal([]byte(body), &req); err != nil {
		return authMiddleware.CreateErrorResponse(400, "Invalid JSON format"), nil
	}

	if req.Email == "" {
		return authMiddleware.CreateErrorResponse(400, "Email is required"), nil
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
	usersRepo := database.NewUsersRepository(db, usersTable)
	verificationsRepo := database.NewEmailVerificationsRepository(db, emailVerificationsTable)

	// Look up user by email (silent failure to prevent email enumeration)
	user, err := usersRepo.GetByEmail(ctx, email)
	if err != nil {
		log.Printf("User lookup error for %s: %v", email, err)
	}
	if user == nil {
		// Return success even if user doesn't exist to prevent email enumeration
		log.Printf("No user found for %s, returning success anyway", email)
		return authMiddleware.CreateSuccessResponse(200, "If an account exists with this email, a reset code has been sent.", nil), nil
	}

	// Expire any existing pending codes for this email
	pendingCodes, err := verificationsRepo.GetPendingByEmail(ctx, email)
	if err != nil {
		log.Printf("Failed to fetch pending codes for %s: %v", email, err)
	}
	for _, pending := range pendingCodes {
		if err := verificationsRepo.MarkAsExpired(ctx, pending.VerificationID); err != nil {
			log.Printf("Failed to expire old code %s: %v", pending.VerificationID, err)
		}
	}

	// Generate secure 6-digit code
	code, err := generateSecureCode()
	if err != nil {
		log.Printf("Failed to generate reset code: %v", err)
		return authMiddleware.CreateErrorResponse(500, "Service unavailable"), nil
	}

	// Store in EMAIL_VERIFICATIONS_TABLE with 15-minute TTL
	verification := &types.EmailVerification{
		VerificationID: "verify:" + uuid.Must(uuid.NewV7()).String(),
		Email:          email,
		Token:          code,
		ExpiresAt:      time.Now().Add(15 * time.Minute).Unix(),
	}
	if err := verificationsRepo.Create(ctx, verification); err != nil {
		log.Printf("Failed to store verification code: %v", err)
		return authMiddleware.CreateErrorResponse(500, "Service unavailable"), nil
	}

	log.Printf("Created reset code for %s (verification_id: %s)", email, verification.VerificationID)

	// Enqueue branded password reset notification via SQS
	notifQueueURL := os.Getenv("NOTIFICATION_QUEUE_URL")
	if notifQueueURL != "" {
		sqsClient := sqs.NewFromConfig(cfg)
		userName := user.Name
		if userName == "" {
			userName = "there"
		}
		jobJSON, _ := json.Marshal(map[string]interface{}{
			"type":    "password_reset",
			"user_id": user.UserID,
			"data": map[string]interface{}{
				"user_email": email,
				"user_name":  userName,
				"reset_code": code,
			},
		})
		_, sqsErr := sqsClient.SendMessage(ctx, &sqs.SendMessageInput{
			QueueUrl:       aws.String(notifQueueURL),
			MessageGroupId: aws.String("pwd-reset"),
			MessageBody:    aws.String(string(jobJSON)),
		})
		if sqsErr != nil {
			log.Printf("Failed to enqueue password reset notification: %v", sqsErr)
		}
	}

	// Always return success to prevent email enumeration
	return authMiddleware.CreateSuccessResponse(200, "If an account exists with this email, a reset code has been sent.", nil), nil
}

// generateSecureCode generates a cryptographically secure 6-digit numeric code
func generateSecureCode() (string, error) {
	max := big.NewInt(1000000) // 0-999999
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", err
	}
	// Pad to 6 digits
	return padCode(n.Int64()), nil
}

func padCode(n int64) string {
	s := make([]byte, 6)
	for i := 5; i >= 0; i-- {
		s[i] = byte('0' + n%10)
		n /= 10
	}
	return string(s)
}
