package email

import (
	"fmt"
	"os"
	"time"
)

// TemplateData holds common template variables
type TemplateData struct {
	UserName   string
	UserEmail  string
	AppName    string
	BaseURL    string
	ResetCode  string
	HelperName string
	PlanName   string
	ErrorMsg   string
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
		mainContent = `
			<p>We were unable to process your most recent payment. Please update your payment method to avoid any interruption to your service.</p>
			<p>We'll retry the payment automatically, but you can also update your billing details from your Settings page.</p>
		`
		textContent = "We were unable to process your most recent payment. Please update your payment method."

	default:
		subject = fmt.Sprintf("Billing Update - %s", data.AppName)
		mainContent = "<p>There has been an update to your billing status. Check your Settings page for details.</p>"
		textContent = "There has been an update to your billing status."
	}

	htmlBody := generateHTML(data, emailContent{
		headerTitle:    "Billing Update",
		headerSubtitle: "Your subscription details",
		greetingIcon:   "creditcard",
		mainContent:    mainContent,
		ctaText:        "Manage Billing",
		ctaURL:         fmt.Sprintf("%s/settings", data.BaseURL),
	})

	textBody := fmt.Sprintf(`%s

Hello %s,

%s

Manage Billing: %s/settings

-- %s`, subject, data.UserName, textContent, data.BaseURL, data.AppName)

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
