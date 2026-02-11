package listEmails

import (
	"context"
	"log"
	"os"
	"strconv"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	authMiddleware "github.com/myfusionhelper/api/internal/middleware/auth"
	"github.com/myfusionhelper/api/internal/services"
)

func HandleWithAuth(ctx context.Context, event events.APIGatewayV2HTTPRequest, authCtx *authMiddleware.AuthContext) (events.APIGatewayV2HTTPResponse, error) {
	// Get query parameters
	limitStr := event.QueryStringParameters["limit"]
	limit := 50
	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
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

	// Get email history for account
	emails, err := emailService.GetEmailHistory(ctx, authCtx.AccountID, limit)
	if err != nil {
		log.Printf("Failed to get email history: %v", err)
		return authMiddleware.CreateErrorResponse(500, "Failed to retrieve emails"), nil
	}

	// Convert to frontend format
	emailList := make([]map[string]interface{}, 0, len(emails))
	for _, email := range emails {
		emailList = append(emailList, map[string]interface{}{
			"id":          email.EmailID,
			"subject":     email.Subject,
			"to":          email.RecipientEmail,
			"status":      email.Status,
			"sent_at":     email.SentAt,
			"created_at":  email.CreatedAt,
			"template_id": email.TemplateID,
			"message_id":  email.MessageID,
			"opens":       0, // TODO: Track opens when engagement tracking is implemented
			"clicks":      0, // TODO: Track clicks when engagement tracking is implemented
		})
	}

	return authMiddleware.CreateSuccessResponse(200, "Emails retrieved", emailList), nil
}
