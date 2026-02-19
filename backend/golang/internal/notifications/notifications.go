package notifications

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

// Service handles sending notifications via the internal-email service
type Service struct {
	httpClient *http.Client
	apiBaseURL string
}

// New creates a new notification service instance
func New(ctx context.Context) (*Service, error) {
	// Get API base URL from environment or use default
	apiBaseURL := os.Getenv("INTERNAL_EMAIL_API_URL")
	if apiBaseURL == "" {
		// Derive from stage -- use custom API domain
		stage := os.Getenv("STAGE")
		switch stage {
		case "main", "":
			apiBaseURL = "https://api.myfusionhelper.ai"
		case "staging":
			apiBaseURL = "https://api-staging.myfusionhelper.ai"
		default:
			apiBaseURL = fmt.Sprintf("https://api-%s.myfusionhelper.ai", stage)
		}
	}

	return &Service{
		httpClient: &http.Client{},
		apiBaseURL: apiBaseURL,
	}, nil
}

// SendEmailRequest matches the internal-email API request format
type SendEmailRequest struct {
	TemplateType string                 `json:"template_type"`
	To           []string               `json:"to"`
	Data         map[string]interface{} `json:"data"`
	AccountID    string                 `json:"account_id,omitempty"`
}

// SendEmailResponse matches the internal-email API response format
type SendEmailResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    struct {
		EmailID   string `json:"email_id"`
		MessageID string `json:"message_id"`
		Status    string `json:"status"`
	} `json:"data,omitempty"`
	Error string `json:"error,omitempty"`
}

// SendWelcomeEmail sends a welcome email to a new user
func (s *Service) SendWelcomeEmail(ctx context.Context, name, email string) error {
	req := SendEmailRequest{
		TemplateType: "welcome",
		To:           []string{email},
		Data: map[string]interface{}{
			"UserName":  name,
			"UserEmail": email,
		},
	}

	return s.sendEmail(ctx, req)
}

// SendPasswordResetEmail sends a password reset email
func (s *Service) SendPasswordResetEmail(ctx context.Context, email, resetCode string) error {
	req := SendEmailRequest{
		TemplateType: "password_reset",
		To:           []string{email},
		Data: map[string]interface{}{
			"UserEmail": email,
			"ResetCode": resetCode,
		},
	}

	return s.sendEmail(ctx, req)
}

// SendEmailVerificationEmail sends an email verification message
func (s *Service) SendEmailVerificationEmail(ctx context.Context, email, verifyCode string) error {
	req := SendEmailRequest{
		TemplateType: "email_verification",
		To:           []string{email},
		Data: map[string]interface{}{
			"UserEmail":  email,
			"VerifyCode": verifyCode,
		},
	}

	return s.sendEmail(ctx, req)
}

// SendHelperExecutionAlert sends an alert about a helper execution issue
func (s *Service) SendHelperExecutionAlert(ctx context.Context, accountID, email, helperName, errorMsg string) error {
	req := SendEmailRequest{
		TemplateType: "execution_alert",
		To:           []string{email},
		Data: map[string]interface{}{
			"HelperName": helperName,
			"ErrorMsg":   errorMsg,
		},
		AccountID: accountID,
	}

	return s.sendEmail(ctx, req)
}

// SendBillingEvent sends a billing-related notification.
// Optional extraData keys are merged into the template data (e.g., "InvoiceURL").
func (s *Service) SendBillingEvent(ctx context.Context, accountID, email, eventType, planName string, extraData ...map[string]interface{}) error {
	data := map[string]interface{}{
		"event_type": eventType,
		"PlanName":   planName,
	}
	if len(extraData) > 0 && extraData[0] != nil {
		for k, v := range extraData[0] {
			data[k] = v
		}
	}

	req := SendEmailRequest{
		TemplateType: "billing_event",
		To:           []string{email},
		Data:         data,
		AccountID:    accountID,
	}

	return s.sendEmail(ctx, req)
}

// SendConnectionAlert sends an alert about a platform connection issue
func (s *Service) SendConnectionAlert(ctx context.Context, accountID, email, connectionName string) error {
	req := SendEmailRequest{
		TemplateType: "connection_alert",
		To:           []string{email},
		Data: map[string]interface{}{
			"connection_name": connectionName,
		},
		AccountID: accountID,
	}

	return s.sendEmail(ctx, req)
}

// SendUsageAlert sends a usage limit alert
func (s *Service) SendUsageAlert(ctx context.Context, accountID, email, resourceName string, usagePercent, usageCurrent, usageLimit int) error {
	req := SendEmailRequest{
		TemplateType: "usage_alert",
		To:           []string{email},
		Data: map[string]interface{}{
			"ResourceName": resourceName,
			"UsagePercent": usagePercent,
			"UsageCurrent": usageCurrent,
			"UsageLimit":   usageLimit,
		},
		AccountID: accountID,
	}

	return s.sendEmail(ctx, req)
}

// SendWeeklySummary sends a weekly activity summary
func (s *Service) SendWeeklySummary(ctx context.Context, accountID, email string, summaryData map[string]interface{}) error {
	req := SendEmailRequest{
		TemplateType: "weekly_summary",
		To:           []string{email},
		Data:         summaryData,
		AccountID:    accountID,
	}

	return s.sendEmail(ctx, req)
}

// SendTeamInvite sends a team invitation email
func (s *Service) SendTeamInvite(ctx context.Context, inviterName, inviterEmail, inviteeEmail, roleName, accountName, inviteToken string) error {
	req := SendEmailRequest{
		TemplateType: "team_invite",
		To:           []string{inviteeEmail},
		Data: map[string]interface{}{
			"InviterName":  inviterName,
			"InviterEmail": inviterEmail,
			"RoleName":     roleName,
			"AccountName":  accountName,
			"InviteToken":  inviteToken,
		},
	}

	return s.sendEmail(ctx, req)
}

// sendEmail makes the HTTP request to the internal-email service
func (s *Service) sendEmail(ctx context.Context, req SendEmailRequest) error {
	// Marshal request body
	body, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal email request: %w", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", s.apiBaseURL+"/internal/emails/send", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	// Send request
	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		log.Printf("Failed to send email via internal-email API: %v", err)
		return fmt.Errorf("failed to send HTTP request: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	// Parse response
	var emailResp SendEmailResponse
	if err := json.Unmarshal(respBody, &emailResp); err != nil {
		log.Printf("Failed to parse email response: %v, body: %s", err, string(respBody))
		return fmt.Errorf("failed to parse response: %w", err)
	}

	// Check for errors
	if !emailResp.Success {
		return fmt.Errorf("email service error: %s", emailResp.Error)
	}

	if resp.StatusCode >= 400 {
		return fmt.Errorf("email service returned status %d: %s", resp.StatusCode, emailResp.Error)
	}

	log.Printf("Email sent successfully: %s (message_id: %s)", req.TemplateType, emailResp.Data.MessageID)
	return nil
}
