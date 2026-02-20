package send

import (
	"context"
	"encoding/json"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/myfusionhelper/api/internal/apiutil"
	"github.com/myfusionhelper/api/internal/email"
	authMiddleware "github.com/myfusionhelper/api/internal/middleware/auth"
	"github.com/myfusionhelper/api/internal/types"
)

var templateLoader *email.TemplateLoader

func init() {
	bucket := os.Getenv("TEMPLATE_BUCKET")
	if bucket == "" {
		log.Fatal("TEMPLATE_BUCKET environment variable is required")
	}
	loader, err := email.NewTemplateLoader(context.Background(), bucket, "email-templates/")
	if err != nil {
		log.Fatalf("Failed to initialize S3 template loader: %v", err)
	}
	templateLoader = loader
	log.Printf("S3 template loader initialized (bucket: %s)", bucket)
}

// SendEmailRequest matches the types.SendEmailRequest
type SendEmailRequest struct {
	TemplateType string                 `json:"template_type"` // welcome, password_reset, etc.
	To           []string               `json:"to"`
	Data         map[string]interface{} `json:"data"` // Template variables
	AccountID    string                 `json:"account_id,omitempty"`
}

// Handle sends an email using pre-defined templates
func Handle(ctx context.Context, event events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	var req SendEmailRequest
	if err := json.Unmarshal([]byte(apiutil.GetBody(event)), &req); err != nil {
		log.Printf("Failed to parse request: %v", err)
		return authMiddleware.CreateErrorResponse(400, "Invalid request body"), nil
	}

	// Validate request
	if len(req.To) == 0 {
		return authMiddleware.CreateErrorResponse(400, "At least one recipient required"), nil
	}
	if req.TemplateType == "" {
		return authMiddleware.CreateErrorResponse(400, "template_type is required"), nil
	}

	// Create SES client
	sesClient, err := email.NewSESClient(ctx)
	if err != nil {
		log.Printf("Failed to create SES client: %v", err)
		return authMiddleware.CreateErrorResponse(500, "Failed to initialize email service"), nil
	}

	// Convert data map to TemplateData struct
	templateData := mapToTemplateData(req.Data)

	// Render email from S3 Liquid template
	templatePath := email.ResolveTemplatePath(req.TemplateType, req.Data)
	emailTemplate, err := email.RenderEmailFromS3(ctx, templateLoader, templatePath, templateData, req.Data)
	if err != nil {
		log.Printf("Failed to render template %s: %v", templatePath, err)
		return authMiddleware.CreateErrorResponse(500, "Failed to render email template: "+err.Error()), nil
	}

	// Send email
	emailMsg := email.EmailMessage{
		To:       req.To,
		Subject:  emailTemplate.Subject,
		HTMLBody: emailTemplate.HTMLBody,
		TextBody: emailTemplate.TextBody,
		Tags: map[string]string{
			"template_type": req.TemplateType,
			"account_id":    req.AccountID,
		},
	}

	result, err := sesClient.SendEmail(ctx, emailMsg)
	if err != nil {
		log.Printf("Failed to send email: %v", err)
		return authMiddleware.CreateErrorResponse(500, "Failed to send email: "+err.Error()), nil
	}

	// Return success response
	return authMiddleware.CreateSuccessResponse(200, "Email sent successfully", types.SendEmailResponse{
		EmailID:   result.MessageID,
		MessageID: result.MessageID,
		Status:    "sent",
	}), nil
}

// mapToTemplateData converts a generic map to TemplateData struct
func mapToTemplateData(data map[string]interface{}) email.TemplateData {
	td := email.GetDefaultTemplateData()

	if v, ok := data["UserName"].(string); ok {
		td.UserName = v
	}
	if v, ok := data["UserEmail"].(string); ok {
		td.UserEmail = v
	}
	if v, ok := data["ResetCode"].(string); ok {
		td.ResetCode = v
	}
	if v, ok := data["HelperName"].(string); ok {
		td.HelperName = v
	}
	if v, ok := data["PlanName"].(string); ok {
		td.PlanName = v
	}
	if v, ok := data["ErrorMsg"].(string); ok {
		td.ErrorMsg = v
	}
	if v, ok := data["ResourceName"].(string); ok {
		td.ResourceName = v
	}
	if v, ok := data["UsagePercent"].(int); ok {
		td.UsagePercent = v
	}
	if v, ok := data["UsageCurrent"].(int); ok {
		td.UsageCurrent = v
	}
	if v, ok := data["UsageLimit"].(int); ok {
		td.UsageLimit = v
	}
	if v, ok := data["TotalHelpers"].(int); ok {
		td.TotalHelpers = v
	}
	if v, ok := data["TotalExecuted"].(int); ok {
		td.TotalExecuted = v
	}
	if v, ok := data["TotalSucceeded"].(int); ok {
		td.TotalSucceeded = v
	}
	if v, ok := data["TotalFailed"].(int); ok {
		td.TotalFailed = v
	}
	if v, ok := data["SuccessRate"].(string); ok {
		td.SuccessRate = v
	}
	if v, ok := data["TopHelper"].(string); ok {
		td.TopHelper = v
	}
	if v, ok := data["WeekStart"].(string); ok {
		td.WeekStart = v
	}
	if v, ok := data["WeekEnd"].(string); ok {
		td.WeekEnd = v
	}
	if v, ok := data["InviterName"].(string); ok {
		td.InviterName = v
	}
	if v, ok := data["InviterEmail"].(string); ok {
		td.InviterEmail = v
	}
	if v, ok := data["RoleName"].(string); ok {
		td.RoleName = v
	}
	if v, ok := data["AccountName"].(string); ok {
		td.AccountName = v
	}
	if v, ok := data["InviteToken"].(string); ok {
		td.InviteToken = v
	}
	if v, ok := data["VerifyCode"].(string); ok {
		td.VerifyCode = v
	}
	if v, ok := data["InvoiceURL"].(string); ok {
		td.InvoiceURL = v
	}
	if v, ok := data["CardLast4"].(string); ok {
		td.CardLast4 = v
	}
	if v, ok := data["CardBrand"].(string); ok {
		td.CardBrand = v
	}
	if v, ok := data["CardExpMonth"].(string); ok {
		td.CardExpMonth = v
	}
	if v, ok := data["CardExpYear"].(string); ok {
		td.CardExpYear = v
	}
	if v, ok := data["Amount"].(string); ok {
		td.Amount = v
	}
	if v, ok := data["InvoiceNumber"].(string); ok {
		td.InvoiceNumber = v
	}
	if v, ok := data["RefundReason"].(string); ok {
		td.RefundReason = v
	}

	return td
}
