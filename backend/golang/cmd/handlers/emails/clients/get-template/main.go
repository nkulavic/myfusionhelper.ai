package getTemplate

import (
	"context"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	authMiddleware "github.com/myfusionhelper/api/internal/middleware/auth"
	"github.com/myfusionhelper/api/internal/services"
)

func HandleWithAuth(ctx context.Context, event events.APIGatewayV2HTTPRequest, authCtx *authMiddleware.AuthContext) (events.APIGatewayV2HTTPResponse, error) {
	templateID := event.PathParameters["id"]
	if templateID == "" {
		return authMiddleware.CreateErrorResponse(400, "Template ID is required"), nil
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

	// Get template
	tmpl, err := emailService.GetTemplate(ctx, templateID)
	if err != nil {
		log.Printf("Failed to get template: %v", err)
		return authMiddleware.CreateErrorResponse(404, "Template not found"), nil
	}

	// Security: ensure template belongs to account (or is a system template)
	if tmpl.AccountID != "" && tmpl.AccountID != authCtx.AccountID {
		return authMiddleware.CreateErrorResponse(403, "Access denied"), nil
	}

	// Convert to frontend format
	templateData := map[string]interface{}{
		"id":           tmpl.TemplateID,
		"name":         tmpl.Name,
		"category":     "custom",
		"subject":      tmpl.Subject,
		"body":         tmpl.HTMLTemplate,
		"is_starred":   false,
		"usage_count":  0,
		"created_at":   tmpl.CreatedAt,
		"updated_at":   tmpl.UpdatedAt,
	}

	return authMiddleware.CreateSuccessResponse(200, "Template retrieved", templateData), nil
}
