package updateTemplate

import (
	"context"
	"encoding/json"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	authMiddleware "github.com/myfusionhelper/api/internal/middleware/auth"
	"github.com/myfusionhelper/api/internal/services"
)

type UpdateTemplateRequest struct {
	Name      *string `json:"name,omitempty"`
	Category  *string `json:"category,omitempty"`
	Subject   *string `json:"subject,omitempty"`
	Body      *string `json:"body,omitempty"`
	IsStarred *bool   `json:"is_starred,omitempty"`
}

func HandleWithAuth(ctx context.Context, event events.APIGatewayV2HTTPRequest, authCtx *authMiddleware.AuthContext) (events.APIGatewayV2HTTPResponse, error) {
	templateID := event.PathParameters["id"]
	if templateID == "" {
		return authMiddleware.CreateErrorResponse(400, "Template ID is required"), nil
	}

	var req UpdateTemplateRequest
	if err := json.Unmarshal([]byte(event.Body), &req); err != nil {
		log.Printf("Failed to parse request: %v", err)
		return authMiddleware.CreateErrorResponse(400, "Invalid request body"), nil
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

	// Get existing template
	tmpl, err := emailService.GetTemplate(ctx, templateID)
	if err != nil {
		log.Printf("Failed to get template: %v", err)
		return authMiddleware.CreateErrorResponse(404, "Template not found"), nil
	}

	// Security: ensure template belongs to account
	if tmpl.AccountID != authCtx.AccountID {
		return authMiddleware.CreateErrorResponse(403, "Access denied"), nil
	}

	// Update fields if provided
	if req.Name != nil {
		tmpl.Name = *req.Name
	}
	if req.Subject != nil {
		tmpl.Subject = *req.Subject
	}
	if req.Body != nil {
		tmpl.HTMLTemplate = *req.Body
		tmpl.TextTemplate = *req.Body // TODO: Strip HTML for text version
	}

	// Update template
	if err := emailService.UpdateTemplate(ctx, tmpl); err != nil {
		log.Printf("Failed to update template: %v", err)
		return authMiddleware.CreateErrorResponse(500, "Failed to update template"), nil
	}

	// Return updated template
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

	return authMiddleware.CreateSuccessResponse(200, "Template updated", templateData), nil
}
