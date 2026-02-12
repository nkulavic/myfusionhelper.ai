package sendEmail

import (
	"context"
	"encoding/json"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	authMiddleware "github.com/myfusionhelper/api/internal/middleware/auth"
	"github.com/myfusionhelper/api/internal/types"
	"github.com/myfusionhelper/api/internal/email"
	"github.com/myfusionhelper/api/internal/services"
)

type SendEmailRequest struct {
	To          string `json:"to"`
	Subject     string `json:"subject"`
	Body        string `json:"body"`
	TemplateID  string `json:"template_id,omitempty"`
	ScheduledAt string `json:"scheduled_at,omitempty"`
}

func HandleWithAuth(ctx context.Context, event events.APIGatewayV2HTTPRequest, authCtx *types.AuthContext) (events.APIGatewayV2HTTPResponse, error) {
	var req SendEmailRequest
	if err := json.Unmarshal([]byte(event.Body), &req); err != nil {
		log.Printf("Failed to parse request: %v", err)
		return authMiddleware.CreateErrorResponse(400, "Invalid request body"), nil
	}

	// Validate request
	if req.To == "" {
		return authMiddleware.CreateErrorResponse(400, "Recipient (to) is required"), nil
	}
	if req.Subject == "" {
		return authMiddleware.CreateErrorResponse(400, "Subject is required"), nil
	}

	// TODO: Handle scheduled emails (store in DB, trigger via EventBridge)
	if req.ScheduledAt != "" {
		return authMiddleware.CreateErrorResponse(501, "Scheduled emails not yet implemented"), nil
	}

	// Initialize AWS config
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(os.Getenv("COGNITO_REGION")))
	if err != nil {
		log.Printf("Failed to load AWS config: %v", err)
		return authMiddleware.CreateErrorResponse(500, "Failed to initialize AWS"), nil
	}

	dbClient := dynamodb.NewFromConfig(cfg)

	// Initialize email service
	emailService, err := services.NewEmailService(
		ctx,
		dbClient,
		os.Getenv("EMAIL_LOGS_TABLE"),
		os.Getenv("EMAIL_TEMPLATES_TABLE"),
		os.Getenv("EMAIL_VERIFICATIONS_TABLE"),
	)
	if err != nil {
		log.Printf("Failed to initialize email service: %v", err)
		return authMiddleware.CreateErrorResponse(500, "Failed to initialize email service"), nil
	}

	// Build email message
	msg := &email.EmailMessage{
		To:       []string{req.To},
		Subject:  req.Subject,
		HTMLBody: req.Body,
		TextBody: req.Body, // TODO: Strip HTML for text version
		Tags: map[string]string{
			"account_id": authCtx.AccountID,
			"sent_via":   "user_interface",
		},
	}

	// Send email
	result, err := emailService.SendEmail(ctx, authCtx.AccountID, msg)
	if err != nil {
		log.Printf("Failed to send email: %v", err)
		return authMiddleware.CreateErrorResponse(500, "Failed to send email: "+err.Error()), nil
	}

	// Return success response
	return authMiddleware.CreateSuccessResponse(200, "Email sent successfully", map[string]interface{}{
		"id":         result.MessageID,
		"subject":    req.Subject,
		"to":         req.To,
		"status":     "sent",
		"message_id": result.MessageID,
	}), nil
}
