package createTemplate

import (
	"context"
	"encoding/json"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/myfusionhelper/api/internal/apiutil"
	authMiddleware "github.com/myfusionhelper/api/internal/middleware/auth"
	"github.com/myfusionhelper/api/internal/services"
	"github.com/myfusionhelper/api/internal/types"
)

type CreateTemplateRequest struct {
	Name     string `json:"name"`
	Category string `json:"category"`
	Subject  string `json:"subject"`
	Body     string `json:"body"`
}

func HandleWithAuth(ctx context.Context, event events.APIGatewayV2HTTPRequest, authCtx *types.AuthContext) (events.APIGatewayV2HTTPResponse, error) {
	var req CreateTemplateRequest
	if err := json.Unmarshal([]byte(apiutil.GetBody(event)), &req); err != nil {
		log.Printf("Failed to parse request: %v", err)
		return authMiddleware.CreateErrorResponse(400, "Invalid request body"), nil
	}

	// Validate request
	if req.Name == "" {
		return authMiddleware.CreateErrorResponse(400, "Template name is required"), nil
	}
	if req.Subject == "" {
		return authMiddleware.CreateErrorResponse(400, "Template subject is required"), nil
	}
	if req.Body == "" {
		return authMiddleware.CreateErrorResponse(400, "Template body is required"), nil
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

	// Create template
	tmpl := &services.EmailTemplate{
		AccountID:    authCtx.AccountID,
		Name:         req.Name,
		Subject:      req.Subject,
		HTMLTemplate: req.Body,
		TextTemplate: req.Body, // TODO: Strip HTML for text version
		IsSystem:     false,
		IsActive:     true,
	}

	if err := emailService.CreateTemplate(ctx, tmpl); err != nil {
		log.Printf("Failed to create template: %v", err)
		return authMiddleware.CreateErrorResponse(500, "Failed to create template"), nil
	}

	// Return created template
	templateData := map[string]interface{}{
		"id":           tmpl.TemplateID,
		"name":         tmpl.Name,
		"category":     req.Category,
		"subject":      tmpl.Subject,
		"body":         tmpl.HTMLTemplate,
		"is_starred":   false,
		"usage_count":  0,
		"created_at":   tmpl.CreatedAt,
		"updated_at":   tmpl.UpdatedAt,
	}

	return authMiddleware.CreateSuccessResponse(201, "Template created", templateData), nil
}
