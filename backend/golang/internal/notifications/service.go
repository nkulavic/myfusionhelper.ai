package notifications

import (
	"context"
	"log"

	"github.com/myfusionhelper/api/internal/email"
	apitypes "github.com/myfusionhelper/api/internal/types"
)

// Service sends notifications via configured channels based on user preferences
type Service struct {
	sesClient *email.SESClient
}

// New creates a new notification service
func New(ctx context.Context) (*Service, error) {
	sesClient, err := email.NewSESClient(ctx)
	if err != nil {
		return nil, err
	}
	return &Service{sesClient: sesClient}, nil
}

// SendWelcomeEmail sends a welcome email to a new user
func (s *Service) SendWelcomeEmail(ctx context.Context, userName, userEmail string) error {
	data := email.GetDefaultTemplateData()
	data.UserName = userName
	data.UserEmail = userEmail

	tpl := email.GetWelcomeEmailTemplate(data)
	_, err := s.sesClient.SendEmail(ctx, email.EmailMessage{
		To:       []string{userEmail},
		Subject:  tpl.Subject,
		HTMLBody: tpl.HTMLBody,
		TextBody: tpl.TextBody,
		Tags:     map[string]string{"type": "welcome"},
	})
	if err != nil {
		log.Printf("Failed to send welcome email to %s: %v", userEmail, err)
	}
	return err
}

// SendPasswordResetEmail sends a password reset code email
func (s *Service) SendPasswordResetEmail(ctx context.Context, userName, userEmail, resetCode string) error {
	data := email.GetDefaultTemplateData()
	data.UserName = userName
	data.UserEmail = userEmail
	data.ResetCode = resetCode

	tpl := email.GetPasswordResetEmailTemplate(data)
	_, err := s.sesClient.SendEmail(ctx, email.EmailMessage{
		To:       []string{userEmail},
		Subject:  tpl.Subject,
		HTMLBody: tpl.HTMLBody,
		TextBody: tpl.TextBody,
		Tags:     map[string]string{"type": "password_reset"},
	})
	if err != nil {
		log.Printf("Failed to send password reset email to %s: %v", userEmail, err)
	}
	return err
}

// SendExecutionAlert sends an execution failure notification if the user has it enabled
func (s *Service) SendExecutionAlert(ctx context.Context, userName, userEmail, helperName, errorMsg string, prefs *apitypes.NotificationPreferences) error {
	if prefs != nil && !prefs.ExecutionFailures {
		log.Printf("Execution failure notifications disabled for %s, skipping", userEmail)
		return nil
	}

	data := email.GetDefaultTemplateData()
	data.UserName = userName
	data.UserEmail = userEmail
	data.HelperName = helperName
	data.ErrorMsg = errorMsg

	tpl := email.GetExecutionAlertEmailTemplate(data)
	_, err := s.sesClient.SendEmail(ctx, email.EmailMessage{
		To:       []string{userEmail},
		Subject:  tpl.Subject,
		HTMLBody: tpl.HTMLBody,
		TextBody: tpl.TextBody,
		Tags:     map[string]string{"type": "execution_alert"},
	})
	if err != nil {
		log.Printf("Failed to send execution alert email to %s: %v", userEmail, err)
	}
	return err
}

// SendBillingEvent sends a billing notification email
func (s *Service) SendBillingEvent(ctx context.Context, userName, userEmail, eventType, planName string) error {
	data := email.GetDefaultTemplateData()
	data.UserName = userName
	data.UserEmail = userEmail
	data.PlanName = planName

	tpl := email.GetBillingEventEmailTemplate(data, eventType)
	_, err := s.sesClient.SendEmail(ctx, email.EmailMessage{
		To:       []string{userEmail},
		Subject:  tpl.Subject,
		HTMLBody: tpl.HTMLBody,
		TextBody: tpl.TextBody,
		Tags:     map[string]string{"type": "billing", "event": eventType},
	})
	if err != nil {
		log.Printf("Failed to send billing event email to %s: %v", userEmail, err)
	}
	return err
}

// SendUsageAlert sends a usage limit warning email if the user has it enabled
func (s *Service) SendUsageAlert(ctx context.Context, userName, userEmail, resourceName string, current, limit, percent int, prefs *apitypes.NotificationPreferences) error {
	if prefs != nil && !prefs.UsageAlerts {
		log.Printf("Usage alert notifications disabled for %s, skipping", userEmail)
		return nil
	}

	data := email.GetDefaultTemplateData()
	data.UserName = userName
	data.UserEmail = userEmail
	data.ResourceName = resourceName
	data.UsageCurrent = current
	data.UsageLimit = limit
	data.UsagePercent = percent

	tpl := email.GetUsageAlertEmailTemplate(data)
	_, err := s.sesClient.SendEmail(ctx, email.EmailMessage{
		To:       []string{userEmail},
		Subject:  tpl.Subject,
		HTMLBody: tpl.HTMLBody,
		TextBody: tpl.TextBody,
		Tags:     map[string]string{"type": "usage_alert", "resource": resourceName},
	})
	if err != nil {
		log.Printf("Failed to send usage alert email to %s: %v", userEmail, err)
	}
	return err
}

// SendWeeklySummary sends a weekly execution summary email if the user has it enabled
func (s *Service) SendWeeklySummary(ctx context.Context, userName, userEmail string, totalHelpers, totalExecuted, totalSucceeded, totalFailed int, successRate, weekStart, weekEnd string, prefs *apitypes.NotificationPreferences) error {
	if prefs != nil && !prefs.WeeklySummary {
		log.Printf("Weekly summary notifications disabled for %s, skipping", userEmail)
		return nil
	}

	data := email.GetDefaultTemplateData()
	data.UserName = userName
	data.UserEmail = userEmail
	data.TotalHelpers = totalHelpers
	data.TotalExecuted = totalExecuted
	data.TotalSucceeded = totalSucceeded
	data.TotalFailed = totalFailed
	data.SuccessRate = successRate
	data.WeekStart = weekStart
	data.WeekEnd = weekEnd

	tpl := email.GetWeeklySummaryEmailTemplate(data)
	_, err := s.sesClient.SendEmail(ctx, email.EmailMessage{
		To:       []string{userEmail},
		Subject:  tpl.Subject,
		HTMLBody: tpl.HTMLBody,
		TextBody: tpl.TextBody,
		Tags:     map[string]string{"type": "weekly_summary"},
	})
	if err != nil {
		log.Printf("Failed to send weekly summary email to %s: %v", userEmail, err)
	}
	return err
}

// SendTeamInviteEmail sends a team invitation email
func (s *Service) SendTeamInviteEmail(ctx context.Context, recipientEmail, inviterName, inviterEmail, roleName, accountName, inviteToken string) error {
	data := email.GetDefaultTemplateData()
	data.UserName = ""
	data.UserEmail = recipientEmail
	data.InviterName = inviterName
	data.InviterEmail = inviterEmail
	data.RoleName = roleName
	data.AccountName = accountName
	data.InviteToken = inviteToken

	tpl := email.GetTeamInviteEmailTemplate(data)
	_, err := s.sesClient.SendEmail(ctx, email.EmailMessage{
		To:       []string{recipientEmail},
		Subject:  tpl.Subject,
		HTMLBody: tpl.HTMLBody,
		TextBody: tpl.TextBody,
		Tags:     map[string]string{"type": "team_invite"},
	})
	if err != nil {
		log.Printf("Failed to send team invite email to %s: %v", recipientEmail, err)
	}
	return err
}

// SendConnectionAlert sends a connection issue notification if the user has it enabled
func (s *Service) SendConnectionAlert(ctx context.Context, userName, userEmail, connectionName, errorMsg string, prefs *apitypes.NotificationPreferences) error {
	if prefs != nil && !prefs.ConnectionIssues {
		log.Printf("Connection issue notifications disabled for %s, skipping", userEmail)
		return nil
	}

	data := email.GetDefaultTemplateData()
	data.UserName = userName
	data.UserEmail = userEmail
	data.ErrorMsg = errorMsg

	tpl := email.GetConnectionAlertEmailTemplate(data, connectionName)
	_, err := s.sesClient.SendEmail(ctx, email.EmailMessage{
		To:       []string{userEmail},
		Subject:  tpl.Subject,
		HTMLBody: tpl.HTMLBody,
		TextBody: tpl.TextBody,
		Tags:     map[string]string{"type": "connection_alert"},
	})
	if err != nil {
		log.Printf("Failed to send connection alert email to %s: %v", userEmail, err)
	}
	return err
}
