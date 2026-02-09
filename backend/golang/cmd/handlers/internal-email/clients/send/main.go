package send

import (
	"context"
	"encoding/json"
	"log"

	"github.com/aws/aws-lambda-go/events"
	authMiddleware "github.com/myfusionhelper/api/internal/middleware/auth"
	"github.com/myfusionhelper/api/internal/email"
	"github.com/myfusionhelper/api/internal/types"
)

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
	if err := json.Unmarshal([]byte(event.Body), &req); err != nil {
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

	// Get the appropriate template
	var emailTemplate email.EmailTemplate
	switch req.TemplateType {
	case "welcome":
		emailTemplate = email.GetWelcomeEmailTemplate(templateData)
	case "password_reset":
		emailTemplate = email.GetPasswordResetEmailTemplate(templateData)
	case "execution_alert":
		emailTemplate = email.GetExecutionAlertEmailTemplate(templateData)
	case "billing_event":
		eventType, _ := req.Data["event_type"].(string)
		emailTemplate = email.GetBillingEventEmailTemplate(templateData, eventType)
	case "connection_alert":
		connectionName, _ := req.Data["connection_name"].(string)
		emailTemplate = email.GetConnectionAlertEmailTemplate(templateData, connectionName)
	case "usage_alert":
		emailTemplate = email.GetUsageAlertEmailTemplate(templateData)
	case "weekly_summary":
		emailTemplate = email.GetWeeklySummaryEmailTemplate(templateData)
	case "team_invite":
		emailTemplate = email.GetTeamInviteEmailTemplate(templateData)
	default:
		return authMiddleware.CreateErrorResponse(400, "Unknown template type: "+req.TemplateType), nil
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

	return td
}
