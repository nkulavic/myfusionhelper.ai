package listTemplates

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

	// Get templates for account (include system templates)
	templates, err := emailService.ListTemplates(ctx, authCtx.AccountID, true)
	if err != nil {
		log.Printf("Failed to list templates: %v", err)
		return authMiddleware.CreateErrorResponse(500, "Failed to retrieve templates"), nil
	}

	// Convert to frontend format
	templateList := make([]map[string]interface{}, 0, len(templates))
	for _, tmpl := range templates {
		templateList = append(templateList, map[string]interface{}{
			"id":           tmpl.TemplateID,
			"name":         tmpl.Name,
			"category":     "custom", // TODO: Add category field to EmailTemplate
			"subject":      tmpl.Subject,
			"body":         tmpl.HTMLTemplate,
			"is_starred":   false, // TODO: Add starred field
			"usage_count":  0,      // TODO: Track usage
			"created_at":   tmpl.CreatedAt,
			"updated_at":   tmpl.UpdatedAt,
		})
	}

	return authMiddleware.CreateSuccessResponse(200, "Templates retrieved", templateList), nil
}
