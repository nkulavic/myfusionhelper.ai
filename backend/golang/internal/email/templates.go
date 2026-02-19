package email

import (
	"fmt"
	"os"
	"time"
)

// TemplateData holds common template variables
type TemplateData struct {
	UserName       string
	UserEmail      string
	AppName        string
	BaseURL        string
	ResetCode      string
	HelperName     string
	PlanName       string
	ErrorMsg       string
	ResourceName   string
	UsagePercent   int
	UsageCurrent   int
	UsageLimit     int
	TotalHelpers   int
	TotalExecuted  int
	TotalSucceeded int
	TotalFailed    int
	SuccessRate    string
	TopHelper      string
	WeekStart      string
	WeekEnd        string
	InviterName    string
	InviterEmail   string
	RoleName       string
	AccountName    string
	InviteToken    string
	VerifyCode     string
	InvoiceURL     string
	CardLast4      string
	CardBrand      string
	CardExpMonth   string
	CardExpYear    string
	Amount         string
	InvoiceNumber  string
	RefundReason   string
}

// EmailTemplate represents a rendered email template
type EmailTemplate struct {
	Subject  string
	HTMLBody string
	TextBody string
}

// GetDefaultTemplateData returns template data with common defaults
func GetDefaultTemplateData() TemplateData {
	return TemplateData{
		AppName: "MyFusion Helper",
		BaseURL: getAppBaseURL(),
	}
}

func getAppBaseURL() string {
	if url := os.Getenv("APP_URL"); url != "" {
		return url
	}
	stage := os.Getenv("STAGE")
	if stage == "dev" {
		return "http://localhost:3001"
	}
	return "https://app.myfusionhelper.ai"
}

func getCurrentYear() string {
	return fmt.Sprintf("%d", time.Now().Year())
}

// GetWelcomeEmailTemplate returns the welcome email for new signups
func GetWelcomeEmailTemplate(data TemplateData) EmailTemplate {
	subject := fmt.Sprintf("Welcome to %s!", data.AppName)

	htmlBody := generateHTML(data, emailContent{
		headerTitle:    "Welcome to MyFusion Helper!",
		headerSubtitle: "Your CRM automation journey starts here",
		greetingIcon:   "rocket",
		mainContent: `
			<p>Welcome to MyFusion Helper! We're excited to have you on board.</p>
			<p>Your account has been created and you're ready to start automating your CRM workflows. Connect your CRM platform, set up Helpers, and let automation do the heavy lifting.</p>

			<div style="background: #f0f4ff; border-radius: 8px; padding: 20px; margin: 20px 0;">
				<h3 style="color: #1b3a6b; margin: 0 0 12px 0; font-size: 16px;">Get Started in 3 Steps:</h3>
				<div style="margin-bottom: 8px;"><strong>1.</strong> Connect your CRM platform (Keap, GoHighLevel, ActiveCampaign, etc.)</div>
				<div style="margin-bottom: 8px;"><strong>2.</strong> Browse and activate Helpers for your workflows</div>
				<div><strong>3.</strong> Watch your automations run on autopilot</div>
			</div>
		`,
		ctaText: "Go to Dashboard",
		ctaURL:  data.BaseURL,
	})

	textBody := fmt.Sprintf(`Welcome to %s!

Hello %s,

Your account has been created and you're ready to start automating your CRM workflows.

Get Started in 3 Steps:
1. Connect your CRM platform (Keap, GoHighLevel, ActiveCampaign, etc.)
2. Browse and activate Helpers for your workflows
3. Watch your automations run on autopilot

Go to Dashboard: %s

-- %s`, data.AppName, data.UserName, data.BaseURL, data.AppName)

	return EmailTemplate{Subject: subject, HTMLBody: htmlBody, TextBody: textBody}
}

// GetPasswordResetEmailTemplate returns the password reset email
func GetPasswordResetEmailTemplate(data TemplateData) EmailTemplate {
	subject := fmt.Sprintf("Reset your %s password", data.AppName)

	htmlBody := generateHTML(data, emailContent{
		headerTitle:    "Password Reset Request",
		headerSubtitle: "Secure access to your account",
		greetingIcon:   "lock",
		mainContent: fmt.Sprintf(`
			<p>We received a request to reset the password for your %s account.</p>
			<p>Use the code below to reset your password:</p>
			<div style="background: #f0f4ff; border: 2px solid #1b3a6b; border-radius: 8px; padding: 16px; font-family: monospace; font-size: 24px; font-weight: 700; text-align: center; letter-spacing: 4px; color: #1b3a6b; margin: 20px 0;">%s</div>
			<p>This code will expire in 15 minutes.</p>
			<p>If you didn't request this reset, you can safely ignore this email.</p>
		`, data.AppName, data.ResetCode),
		ctaText: "Reset Password",
		ctaURL:  fmt.Sprintf("%s/reset-password?email=%s", data.BaseURL, data.UserEmail),
	})

	textBody := fmt.Sprintf(`Reset your %s password

Hello %s,

We received a request to reset the password for your account.

Your reset code is: %s

This code will expire in 15 minutes.

Reset Password: %s/reset-password?email=%s

If you didn't request this reset, you can safely ignore this email.

-- %s`, data.AppName, data.UserName, data.ResetCode, data.BaseURL, data.UserEmail, data.AppName)

	return EmailTemplate{Subject: subject, HTMLBody: htmlBody, TextBody: textBody}
}

// GetExecutionAlertEmailTemplate returns the execution failure alert email
func GetExecutionAlertEmailTemplate(data TemplateData) EmailTemplate {
	subject := fmt.Sprintf("[%s] Helper Execution Failed: %s", data.AppName, data.HelperName)

	htmlBody := generateHTML(data, emailContent{
		headerTitle:    "Execution Alert",
		headerSubtitle: "A helper needs your attention",
		greetingIcon:   "warning",
		mainContent: fmt.Sprintf(`
			<p>A helper execution has failed and may need your attention.</p>

			<div style="background: #fff5f5; border-left: 4px solid #e53e3e; border-radius: 4px; padding: 16px; margin: 16px 0;">
				<strong>Helper:</strong> %s<br>
				<strong>Error:</strong> %s
			</div>

			<p>Check your execution logs for more details and retry the helper if needed.</p>
		`, data.HelperName, data.ErrorMsg),
		ctaText: "View Executions",
		ctaURL:  fmt.Sprintf("%s/executions", data.BaseURL),
	})

	textBody := fmt.Sprintf(`[%s] Helper Execution Failed

Hello %s,

A helper execution has failed and may need your attention.

Helper: %s
Error: %s

View Executions: %s/executions

-- %s`, data.AppName, data.UserName, data.HelperName, data.ErrorMsg, data.BaseURL, data.AppName)

	return EmailTemplate{Subject: subject, HTMLBody: htmlBody, TextBody: textBody}
}

// GetBillingEventEmailTemplate returns billing-related notification emails
func GetBillingEventEmailTemplate(data TemplateData, eventType string) EmailTemplate {
	var subject, mainContent, textContent string

	switch eventType {
	case "subscription_created":
		subject = fmt.Sprintf("Subscription Confirmed - %s %s", data.AppName, data.PlanName)
		mainContent = fmt.Sprintf(`
			<p>Your subscription to the <strong>%s</strong> plan has been confirmed.</p>
			<p>You now have access to all the features included in your plan. Head to your dashboard to start building automations.</p>
		`, data.PlanName)
		textContent = fmt.Sprintf("Your subscription to the %s plan has been confirmed.", data.PlanName)

	case "subscription_cancelled":
		subject = fmt.Sprintf("Subscription Cancelled - %s", data.AppName)
		mainContent = `
			<p>Your subscription has been cancelled. You'll continue to have access to your current plan features until the end of your billing period.</p>
			<p>After that, your account will revert to the Free plan. Your data and configurations will be preserved.</p>
			<p>You can resubscribe at any time from your Settings page.</p>
		`
		textContent = "Your subscription has been cancelled. You'll continue to have access until the end of your billing period."

	case "payment_failed":
		subject = fmt.Sprintf("[Action Required] Payment Failed - %s", data.AppName)
		if data.InvoiceURL != "" {
			mainContent = fmt.Sprintf(`
				<p>We were unable to process your most recent payment. Please update your payment method to avoid any interruption to your service.</p>
				<p>We'll retry the payment automatically, but you can also resolve this now by clicking the button below.</p>
			`)
		} else {
			mainContent = `
				<p>We were unable to process your most recent payment. Please update your payment method to avoid any interruption to your service.</p>
				<p>We'll retry the payment automatically, but you can also update your billing details from your Settings page.</p>
			`
		}
		textContent = "We were unable to process your most recent payment. Please update your payment method."

	case "trial_ending":
		subject = fmt.Sprintf("Your %s trial ends soon", data.AppName)
		mainContent = fmt.Sprintf(`
			<p>Your 14-day free trial is ending soon. To keep your Helpers running and avoid any interruption, choose a plan that fits your needs.</p>

			<div style="background: #f0f4ff; border-radius: 8px; padding: 16px; margin: 16px 0;">
				<strong>What happens next:</strong>
				<ul style="margin: 8px 0 0 0; padding-left: 20px; color: #374151;">
					<li>Your Helpers will pause when the trial expires</li>
					<li>Your data and configurations are preserved</li>
					<li>Subscribe to any plan to resume instantly</li>
				</ul>
			</div>

			<p>Plans start at just $39/month for 10 active Helpers and 5,000 executions.</p>
		`)
		textContent = fmt.Sprintf("Your %s trial is ending soon. Subscribe to a plan to keep your Helpers running. Plans start at $39/month.", data.AppName)

	case "plan_upgraded":
		subject = fmt.Sprintf("Plan Upgraded to %s - %s", data.PlanName, data.AppName)
		mainContent = fmt.Sprintf(`
			<p>Your plan has been upgraded to <strong>%s</strong>. Your new features and limits are available immediately.</p>
			<p>Thank you for growing with MyFusion Helper! Check your dashboard to take advantage of your expanded capabilities.</p>
		`, data.PlanName)
		textContent = fmt.Sprintf("Your plan has been upgraded to %s. New features are available immediately.", data.PlanName)

	case "plan_downgraded":
		subject = fmt.Sprintf("Plan Changed to %s - %s", data.PlanName, data.AppName)
		mainContent = fmt.Sprintf(`
			<p>Your plan has been changed to <strong>%s</strong>. The change will take effect at the start of your next billing cycle.</p>
			<p>Until then, you'll continue to have access to your current plan features. If any of your active Helpers exceed the new plan limits, they'll be paused automatically.</p>
		`, data.PlanName)
		textContent = fmt.Sprintf("Your plan has been changed to %s. The change takes effect at your next billing cycle.", data.PlanName)

	case "payment_recovered":
		subject = fmt.Sprintf("Payment Successful - %s", data.AppName)
		mainContent = `
			<p>Great news! Your recent payment has been processed successfully and your account is back to full access.</p>
			<p>No further action is needed. Thank you for staying with us!</p>
		`
		textContent = "Your recent payment has been processed successfully. Your account is fully active."

	case "trial_expired":
		subject = fmt.Sprintf("Your %s trial has ended", data.AppName)
		mainContent = `
			<p>Your 14-day free trial has ended. Your Helpers have been paused, but your data and configurations are safe.</p>

			<div style="background: #f0f4ff; border-radius: 8px; padding: 16px; margin: 16px 0;">
				<strong>What you can still do:</strong>
				<ul style="margin: 8px 0 0 0; padding-left: 20px; color: #374151;">
					<li>Log in and view your dashboard</li>
					<li>Browse your existing Helpers and data</li>
					<li>Subscribe to resume everything instantly</li>
				</ul>
			</div>

			<p>Plans start at just $39/month. Pick up right where you left off.</p>
		`
		textContent = fmt.Sprintf("Your %s trial has ended. Subscribe to any plan to resume your Helpers. Plans start at $39/month.", data.AppName)

	case "payment_receipt":
		subject = fmt.Sprintf("Payment Receipt - %s", data.AppName)
		invoiceRef := ""
		if data.InvoiceNumber != "" {
			invoiceRef = fmt.Sprintf(`<strong>Invoice:</strong> %s<br>`, data.InvoiceNumber)
		}
		mainContent = fmt.Sprintf(`
			<p>Your payment has been processed successfully. Here are the details:</p>

			<div style="background: #f0fdf4; border-left: 4px solid #22c55e; border-radius: 4px; padding: 16px; margin: 16px 0;">
				%s
				<strong>Amount:</strong> %s<br>
				<strong>Plan:</strong> %s
			</div>

			<p>Thank you for your continued support!</p>
		`, invoiceRef, data.Amount, data.PlanName)
		textContent = fmt.Sprintf("Your payment of %s for the %s plan has been processed successfully.", data.Amount, data.PlanName)

	case "card_expiring":
		subject = fmt.Sprintf("[Action Required] Card Expiring Soon - %s", data.AppName)
		mainContent = fmt.Sprintf(`
			<p>The payment method on file for your account is expiring soon.</p>

			<div style="background: #fffbeb; border-left: 4px solid #f59e0b; border-radius: 4px; padding: 16px; margin: 16px 0;">
				<strong>Card:</strong> %s ending in %s<br>
				<strong>Expires:</strong> %s/%s
			</div>

			<p>Please update your payment method before your next billing date to avoid any interruption to your service.</p>
		`, data.CardBrand, data.CardLast4, data.CardExpMonth, data.CardExpYear)
		textContent = fmt.Sprintf("Your %s card ending in %s expires %s/%s. Please update your payment method.", data.CardBrand, data.CardLast4, data.CardExpMonth, data.CardExpYear)

	case "refund_processed":
		subject = fmt.Sprintf("Refund Processed - %s", data.AppName)
		reasonBlock := ""
		if data.RefundReason != "" {
			reasonBlock = fmt.Sprintf(`<strong>Reason:</strong> %s<br>`, data.RefundReason)
		}
		mainContent = fmt.Sprintf(`
			<p>A refund has been processed for your account. Here are the details:</p>

			<div style="background: #f0f4ff; border-left: 4px solid #3b82f6; border-radius: 4px; padding: 16px; margin: 16px 0;">
				<strong>Refund Amount:</strong> %s<br>
				%s
			</div>

			<p>The refund should appear on your statement within 5-10 business days, depending on your bank.</p>
		`, data.Amount, reasonBlock)
		textContent = fmt.Sprintf("A refund of %s has been processed for your account. It should appear on your statement within 5-10 business days.", data.Amount)

	default:
		subject = fmt.Sprintf("Billing Update - %s", data.AppName)
		mainContent = "<p>There has been an update to your billing status. Check your Settings page for details.</p>"
		textContent = "There has been an update to your billing status."
	}

	// Default CTA for most billing emails
	ctaText := "Manage Billing"
	ctaURL := fmt.Sprintf("%s/settings?tab=billing", data.BaseURL)

	// Customize CTA per event type
	switch eventType {
	case "payment_failed":
		if data.InvoiceURL != "" {
			ctaText = "Pay Invoice Now"
			ctaURL = data.InvoiceURL
		} else {
			ctaText = "Update Payment Method"
		}
	case "card_expiring":
		ctaText = "Update Payment Method"
	case "payment_receipt":
		if data.InvoiceURL != "" {
			ctaText = "View Invoice"
			ctaURL = data.InvoiceURL
		}
	}

	htmlBody := generateHTML(data, emailContent{
		headerTitle:    "Billing Update",
		headerSubtitle: "Your subscription details",
		greetingIcon:   "creditcard",
		mainContent:    mainContent,
		ctaText:        ctaText,
		ctaURL:         ctaURL,
	})

	textBody := fmt.Sprintf(`%s

Hello %s,

%s

%s: %s

-- %s`, subject, data.UserName, textContent, ctaText, ctaURL, data.AppName)

	return EmailTemplate{Subject: subject, HTMLBody: htmlBody, TextBody: textBody}
}

// GetConnectionAlertEmailTemplate returns connection issue notification emails
func GetConnectionAlertEmailTemplate(data TemplateData, connectionName string) EmailTemplate {
	subject := fmt.Sprintf("[%s] Connection Issue: %s", data.AppName, connectionName)

	htmlBody := generateHTML(data, emailContent{
		headerTitle:    "Connection Alert",
		headerSubtitle: "A CRM connection needs attention",
		greetingIcon:   "warning",
		mainContent: fmt.Sprintf(`
			<p>We've detected an issue with your <strong>%s</strong> connection.</p>

			<div style="background: #fffbeb; border-left: 4px solid #f59e0b; border-radius: 4px; padding: 16px; margin: 16px 0;">
				<strong>Connection:</strong> %s<br>
				<strong>Issue:</strong> %s
			</div>

			<p>This may affect any Helpers that use this connection. Please check your connection settings and re-authorize if needed.</p>
		`, connectionName, connectionName, data.ErrorMsg),
		ctaText: "View Connections",
		ctaURL:  fmt.Sprintf("%s/connections", data.BaseURL),
	})

	textBody := fmt.Sprintf(`[%s] Connection Issue: %s

Hello %s,

We've detected an issue with your %s connection.

Issue: %s

This may affect any Helpers that use this connection. Please check your connection settings.

View Connections: %s/connections

-- %s`, data.AppName, connectionName, data.UserName, connectionName, data.ErrorMsg, data.BaseURL, data.AppName)

	return EmailTemplate{Subject: subject, HTMLBody: htmlBody, TextBody: textBody}
}

// GetUsageAlertEmailTemplate returns usage limit warning emails
func GetUsageAlertEmailTemplate(data TemplateData) EmailTemplate {
	subject := fmt.Sprintf("[%s] Usage Alert: %s at %d%%", data.AppName, data.ResourceName, data.UsagePercent)

	htmlBody := generateHTML(data, emailContent{
		headerTitle:    "Usage Alert",
		headerSubtitle: "Approaching your plan limit",
		greetingIcon:   "warning",
		mainContent: fmt.Sprintf(`
			<p>You're approaching the limit for <strong>%s</strong> on your current plan.</p>

			<div style="background: #fffbeb; border-left: 4px solid #f59e0b; border-radius: 4px; padding: 16px; margin: 16px 0;">
				<strong>Resource:</strong> %s<br>
				<strong>Usage:</strong> %d of %d (%d%% used)
			</div>

			<p>Consider upgrading your plan to avoid any interruptions to your automation workflows.</p>
		`, data.ResourceName, data.ResourceName, data.UsageCurrent, data.UsageLimit, data.UsagePercent),
		ctaText: "Upgrade Plan",
		ctaURL:  fmt.Sprintf("%s/settings", data.BaseURL),
	})

	textBody := fmt.Sprintf(`[%s] Usage Alert

Hello %s,

You're approaching the limit for %s on your current plan.

Resource: %s
Usage: %d of %d (%d%% used)

Consider upgrading your plan to avoid interruptions.

Upgrade Plan: %s/settings

-- %s`, data.AppName, data.UserName, data.ResourceName, data.ResourceName, data.UsageCurrent, data.UsageLimit, data.UsagePercent, data.BaseURL, data.AppName)

	return EmailTemplate{Subject: subject, HTMLBody: htmlBody, TextBody: textBody}
}

// GetWeeklySummaryEmailTemplate returns weekly execution summary emails
func GetWeeklySummaryEmailTemplate(data TemplateData) EmailTemplate {
	subject := fmt.Sprintf("[%s] Weekly Summary: %s - %s", data.AppName, data.WeekStart, data.WeekEnd)

	htmlBody := generateHTML(data, emailContent{
		headerTitle:    "Weekly Summary",
		headerSubtitle: fmt.Sprintf("%s - %s", data.WeekStart, data.WeekEnd),
		greetingIcon:   "rocket",
		mainContent: fmt.Sprintf(`
			<p>Here's a summary of your CRM automation activity this week.</p>

			<div style="display: grid; grid-template-columns: 1fr 1fr; gap: 12px; margin: 20px 0;">
				<div style="background: #f0f4ff; border-radius: 8px; padding: 16px; text-align: center;">
					<div style="font-size: 24px; font-weight: 700; color: #1b3a6b;">%d</div>
					<div style="font-size: 13px; color: #6b7280;">Executions</div>
				</div>
				<div style="background: #f0fdf4; border-radius: 8px; padding: 16px; text-align: center;">
					<div style="font-size: 24px; font-weight: 700; color: #16a34a;">%s</div>
					<div style="font-size: 13px; color: #6b7280;">Success Rate</div>
				</div>
				<div style="background: #f0f4ff; border-radius: 8px; padding: 16px; text-align: center;">
					<div style="font-size: 24px; font-weight: 700; color: #1b3a6b;">%d</div>
					<div style="font-size: 13px; color: #6b7280;">Active Helpers</div>
				</div>
				<div style="background: #fef2f2; border-radius: 8px; padding: 16px; text-align: center;">
					<div style="font-size: 24px; font-weight: 700; color: #dc2626;">%d</div>
					<div style="font-size: 13px; color: #6b7280;">Failed</div>
				</div>
			</div>
		`, data.TotalExecuted, data.SuccessRate, data.TotalHelpers, data.TotalFailed),
		ctaText: "View Dashboard",
		ctaURL:  data.BaseURL,
	})

	textBody := fmt.Sprintf(`[%s] Weekly Summary: %s - %s

Hello %s,

Here's a summary of your CRM automation activity this week.

Executions: %d
Success Rate: %s
Active Helpers: %d
Failed: %d

View Dashboard: %s

-- %s`, data.AppName, data.WeekStart, data.WeekEnd, data.UserName, data.TotalExecuted, data.SuccessRate, data.TotalHelpers, data.TotalFailed, data.BaseURL, data.AppName)

	return EmailTemplate{Subject: subject, HTMLBody: htmlBody, TextBody: textBody}
}

// GetTeamInviteEmailTemplate returns the team invitation email
func GetTeamInviteEmailTemplate(data TemplateData) EmailTemplate {
	subject := fmt.Sprintf("%s invited you to %s on MyFusion Helper", data.InviterName, data.AccountName)

	roleDescription := getRoleDescription(data.RoleName)

	htmlBody := generateHTML(data, emailContent{
		headerTitle:    "You're Invited!",
		headerSubtitle: fmt.Sprintf("Join %s on MyFusion Helper", data.AccountName),
		greetingIcon:   "rocket",
		mainContent: fmt.Sprintf(`
			<p><strong>%s</strong> (%s) has invited you to join the <strong>%s</strong> team on MyFusion Helper.</p>

			<div style="background: #f0f4ff; border-radius: 8px; padding: 16px; margin: 16px 0;">
				<strong>Your role:</strong> %s<br>
				<span style="font-size: 13px; color: #6b7280;">%s</span>
			</div>

			<p>MyFusion Helper is a CRM automation platform with 62 pre-built Helpers that extend Keap, GoHighLevel, ActiveCampaign, Ontraport, and HubSpot. Accept the invitation below to get started.</p>
		`, data.InviterName, data.InviterEmail, data.AccountName, data.RoleName, roleDescription),
		ctaText: "Accept Invitation",
		ctaURL:  fmt.Sprintf("%s/invite/%s", data.BaseURL, data.InviteToken),
	})

	textBody := fmt.Sprintf(`You're Invited to %s on MyFusion Helper!

%s (%s) has invited you to join the %s team.

Your role: %s
%s

Accept Invitation: %s/invite/%s

-- MyFusion Helper`, data.AccountName, data.InviterName, data.InviterEmail, data.AccountName, data.RoleName, roleDescription, data.BaseURL, data.InviteToken)

	return EmailTemplate{Subject: subject, HTMLBody: htmlBody, TextBody: textBody}
}

func getRoleDescription(role string) string {
	switch role {
	case "admin":
		return "Full access to manage Helpers, connections, team members, and billing."
	case "member":
		return "Can create and manage Helpers and connections. Cannot manage billing or team settings."
	case "viewer":
		return "Read-only access to view Helpers, executions, and reports."
	default:
		return "Access to the team workspace."
	}
}

// GetSMSVerificationTemplate returns the 2FA SMS message body
func GetSMSVerificationTemplate(code string) string {
	return fmt.Sprintf("MyFusion Helper: Your verification code is %s. It expires in 10 minutes. Do not share this code.", code)
}

type emailContent struct {
	headerTitle    string
	headerSubtitle string
	greetingIcon   string
	mainContent    string
	ctaText        string
	ctaURL         string
}

func generateHTML(data TemplateData, content emailContent) string {
	ctaHTML := ""
	if content.ctaText != "" && content.ctaURL != "" {
		ctaHTML = fmt.Sprintf(`
			<div style="text-align: center; margin: 28px 0;">
				<a href="%s" style="display: inline-block; background: #1b3a6b; color: #ffffff; text-decoration: none; padding: 14px 28px; border-radius: 6px; font-size: 15px; font-weight: 600;">%s</a>
			</div>
		`, content.ctaURL, content.ctaText)
	}

	iconHTML := getIconSVG(content.greetingIcon)

	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>%s</title>
</head>
<body style="margin: 0; padding: 0; font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; background-color: #f4f5f7; -webkit-text-size-adjust: 100%%;">
    <div style="width: 100%%; background-color: #f4f5f7; padding: 40px 0;">
        <div style="max-width: 580px; margin: 0 auto; background: #ffffff; border-radius: 8px; overflow: hidden; box-shadow: 0 4px 16px rgba(0,0,0,0.06);">
            <!-- Header -->
            <div style="background: #1b3a6b; padding: 36px 32px; text-align: center;">
                <div style="font-size: 20px; font-weight: 700; color: #ffffff; letter-spacing: 0.5px; margin-bottom: 6px;">MyFusion Helper</div>
                <div style="font-size: 22px; font-weight: 600; color: #ffffff; margin-bottom: 6px;">%s</div>
                <div style="font-size: 14px; color: rgba(255,255,255,0.8);">%s</div>
            </div>

            <!-- Body -->
            <div style="padding: 32px;">
                <div style="font-size: 18px; font-weight: 600; color: #1b3a6b; margin-bottom: 20px; display: flex; align-items: center; gap: 8px;">
                    %s
                    <span>Hello %s!</span>
                </div>
                <div style="font-size: 15px; line-height: 1.6; color: #374151;">
                    %s
                </div>
                %s
            </div>

            <!-- Footer -->
            <div style="background: #f9fafb; padding: 24px 32px; text-align: center; border-top: 1px solid #e5e7eb;">
                <div style="font-size: 12px; color: #9ca3af;">
                    &copy; %s MyFusion Helper. All rights reserved.<br>
                    This email was sent to %s
                </div>
            </div>
        </div>
    </div>
</body>
</html>`,
		content.headerTitle,
		content.headerTitle, content.headerSubtitle,
		iconHTML, data.UserName,
		content.mainContent,
		ctaHTML,
		getCurrentYear(), data.UserEmail,
	)
}

func getIconSVG(icon string) string {
	switch icon {
	case "rocket":
		return `<span style="font-size: 22px;">&#128640;</span>`
	case "lock":
		return `<span style="font-size: 22px;">&#128274;</span>`
	case "warning":
		return `<span style="font-size: 22px;">&#9888;&#65039;</span>`
	case "creditcard":
		return `<span style="font-size: 22px;">&#128179;</span>`
	default:
		return ""
	}
}
