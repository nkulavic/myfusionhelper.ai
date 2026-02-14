package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/myfusionhelper/api/internal/database"
	"github.com/myfusionhelper/api/internal/types"
)

// System email templates for MyFusion Helper
var systemTemplates = []types.EmailTemplate{
	{
		TemplateID:   "welcome",
		Name:         "Welcome Email",
		Subject:      "Welcome to MyFusion Helper! 🎉",
		IsSystem:     true,
		IsActive:     true,
		HTMLTemplate: welcomeHTMLTemplate,
		TextTemplate: welcomeTextTemplate,
		Variables: []types.EmailVariable{
			{Name: "FirstName", Description: "User's first name", Required: true},
			{Name: "Email", Description: "User's email address", Required: true},
			{Name: "AppURL", Description: "Application URL", Required: true},
		},
	},
	{
		TemplateID:   "password_reset",
		Name:         "Password Reset",
		Subject:      "Reset Your MyFusion Helper Password",
		IsSystem:     true,
		IsActive:     true,
		HTMLTemplate: passwordResetHTMLTemplate,
		TextTemplate: passwordResetTextTemplate,
		Variables: []types.EmailVariable{
			{Name: "Email", Description: "User's email address", Required: true},
			{Name: "ResetURL", Description: "Password reset URL with token", Required: true},
			{Name: "ExpiresIn", Description: "Token expiration time", Required: true},
		},
	},
	{
		TemplateID:   "email_verification",
		Name:         "Email Verification",
		Subject:      "Verify Your Email Address",
		IsSystem:     true,
		IsActive:     true,
		HTMLTemplate: emailVerificationHTMLTemplate,
		TextTemplate: emailVerificationTextTemplate,
		Variables: []types.EmailVariable{
			{Name: "Email", Description: "User's email address", Required: true},
			{Name: "VerificationURL", Description: "Verification URL with token", Required: true},
			{Name: "ExpiresIn", Description: "Token expiration time", Required: true},
		},
	},
	{
		TemplateID:   "helper_report",
		Name:         "Helper Execution Report",
		Subject:      "Helper Report: {{.HelperName}}",
		IsSystem:     true,
		IsActive:     true,
		HTMLTemplate: helperReportHTMLTemplate,
		TextTemplate: helperReportTextTemplate,
		Variables: []types.EmailVariable{
			{Name: "HelperName", Description: "Name of the helper", Required: true},
			{Name: "Details", Description: "Execution details", Required: true},
			{Name: "AppURL", Description: "Application URL", Required: true},
		},
	},
	{
		TemplateID:   "account_notification",
		Name:         "Account Notification",
		Subject:      "{{.Subject}}",
		IsSystem:     true,
		IsActive:     true,
		HTMLTemplate: accountNotificationHTMLTemplate,
		TextTemplate: accountNotificationTextTemplate,
		Variables: []types.EmailVariable{
			{Name: "Subject", Description: "Notification subject", Required: true},
			{Name: "Message", Description: "Notification message", Required: true},
			{Name: "AppURL", Description: "Application URL", Required: true},
		},
	},
	{
		TemplateID:   "team_invitation",
		Name:         "Team Invitation",
		Subject:      "You've been invited to join {{.TeamName}}",
		IsSystem:     true,
		IsActive:     true,
		HTMLTemplate: teamInvitationHTMLTemplate,
		TextTemplate: teamInvitationTextTemplate,
		Variables: []types.EmailVariable{
			{Name: "InviterName", Description: "Name of person who sent invitation", Required: true},
			{Name: "TeamName", Description: "Name of the team", Required: true},
			{Name: "InviteURL", Description: "Invitation acceptance URL", Required: true},
		},
	},
}

// HTML Templates

const welcomeHTMLTemplate = `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Welcome to MyFusion Helper</title>
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333; max-width: 600px; margin: 0 auto; padding: 20px;">
    <div style="background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); padding: 30px; text-align: center; border-radius: 10px 10px 0 0;">
        <h1 style="color: white; margin: 0;">Welcome to MyFusion Helper!</h1>
    </div>
    <div style="background: #f9f9f9; padding: 30px; border-radius: 0 0 10px 10px;">
        <p>Hi {{.FirstName}},</p>
        <p>Welcome to MyFusion Helper! We're excited to have you on board.</p>
        <p>Your account ({{.Email}}) has been successfully created. You can now start connecting your CRM platforms and creating powerful automation helpers.</p>
        <div style="text-align: center; margin: 30px 0;">
            <a href="{{.AppURL}}" style="background: #667eea; color: white; padding: 12px 30px; text-decoration: none; border-radius: 5px; display: inline-block;">Get Started</a>
        </div>
        <p>If you have any questions, feel free to reach out to our support team.</p>
        <p>Best regards,<br>The MyFusion Helper Team</p>
    </div>
    <div style="text-align: center; margin-top: 20px; color: #999; font-size: 12px;">
        <p>MyFusion Helper | Automate Your CRM Workflows</p>
    </div>
</body>
</html>`

const welcomeTextTemplate = `Welcome to MyFusion Helper!

Hi {{.FirstName}},

Welcome to MyFusion Helper! We're excited to have you on board.

Your account ({{.Email}}) has been successfully created. You can now start connecting your CRM platforms and creating powerful automation helpers.

Get started: {{.AppURL}}

If you have any questions, feel free to reach out to our support team.

Best regards,
The MyFusion Helper Team

---
MyFusion Helper | Automate Your CRM Workflows`

const passwordResetHTMLTemplate = `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Reset Your Password</title>
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333; max-width: 600px; margin: 0 auto; padding: 20px;">
    <div style="background: #667eea; padding: 30px; text-align: center; border-radius: 10px 10px 0 0;">
        <h1 style="color: white; margin: 0;">Reset Your Password</h1>
    </div>
    <div style="background: #f9f9f9; padding: 30px; border-radius: 0 0 10px 10px;">
        <p>Hi there,</p>
        <p>We received a request to reset the password for your MyFusion Helper account ({{.Email}}).</p>
        <p>Click the button below to reset your password:</p>
        <div style="text-align: center; margin: 30px 0;">
            <a href="{{.ResetURL}}" style="background: #667eea; color: white; padding: 12px 30px; text-decoration: none; border-radius: 5px; display: inline-block;">Reset Password</a>
        </div>
        <p><strong>This link will expire in {{.ExpiresIn}}.</strong></p>
        <p>If you didn't request this password reset, you can safely ignore this email. Your password will remain unchanged.</p>
        <p>Best regards,<br>The MyFusion Helper Team</p>
    </div>
    <div style="text-align: center; margin-top: 20px; color: #999; font-size: 12px;">
        <p>If the button doesn't work, copy and paste this link into your browser:</p>
        <p style="word-break: break-all;">{{.ResetURL}}</p>
    </div>
</body>
</html>`

const passwordResetTextTemplate = `Reset Your Password

Hi there,

We received a request to reset the password for your MyFusion Helper account ({{.Email}}).

Click the link below to reset your password:
{{.ResetURL}}

This link will expire in {{.ExpiresIn}}.

If you didn't request this password reset, you can safely ignore this email. Your password will remain unchanged.

Best regards,
The MyFusion Helper Team`

const emailVerificationHTMLTemplate = `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Verify Your Email</title>
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333; max-width: 600px; margin: 0 auto; padding: 20px;">
    <div style="background: #667eea; padding: 30px; text-align: center; border-radius: 10px 10px 0 0;">
        <h1 style="color: white; margin: 0;">Verify Your Email Address</h1>
    </div>
    <div style="background: #f9f9f9; padding: 30px; border-radius: 0 0 10px 10px;">
        <p>Hi there,</p>
        <p>Thanks for signing up for MyFusion Helper! Please verify your email address ({{.Email}}) to complete your account setup.</p>
        <div style="text-align: center; margin: 30px 0;">
            <a href="{{.VerificationURL}}" style="background: #667eea; color: white; padding: 12px 30px; text-decoration: none; border-radius: 5px; display: inline-block;">Verify Email</a>
        </div>
        <p><strong>This link will expire in {{.ExpiresIn}}.</strong></p>
        <p>If you didn't create this account, you can safely ignore this email.</p>
        <p>Best regards,<br>The MyFusion Helper Team</p>
    </div>
    <div style="text-align: center; margin-top: 20px; color: #999; font-size: 12px;">
        <p>If the button doesn't work, copy and paste this link into your browser:</p>
        <p style="word-break: break-all;">{{.VerificationURL}}</p>
    </div>
</body>
</html>`

const emailVerificationTextTemplate = `Verify Your Email Address

Hi there,

Thanks for signing up for MyFusion Helper! Please verify your email address ({{.Email}}) to complete your account setup.

Click the link below to verify:
{{.VerificationURL}}

This link will expire in {{.ExpiresIn}}.

If you didn't create this account, you can safely ignore this email.

Best regards,
The MyFusion Helper Team`

const helperReportHTMLTemplate = `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Helper Report</title>
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333; max-width: 600px; margin: 0 auto; padding: 20px;">
    <div style="background: #667eea; padding: 30px; text-align: center; border-radius: 10px 10px 0 0;">
        <h1 style="color: white; margin: 0;">Helper Execution Report</h1>
    </div>
    <div style="background: #f9f9f9; padding: 30px; border-radius: 0 0 10px 10px;">
        <p>Hi there,</p>
        <p>Your helper <strong>{{.HelperName}}</strong> has completed execution.</p>
        <div style="background: white; padding: 20px; border-radius: 5px; margin: 20px 0;">
            <h3>Execution Details:</h3>
            <pre style="background: #f5f5f5; padding: 15px; border-radius: 5px; overflow-x: auto;">{{.Details}}</pre>
        </div>
        <div style="text-align: center; margin: 30px 0;">
            <a href="{{.AppURL}}/helpers" style="background: #667eea; color: white; padding: 12px 30px; text-decoration: none; border-radius: 5px; display: inline-block;">View Helper</a>
        </div>
        <p>Best regards,<br>The MyFusion Helper Team</p>
    </div>
</body>
</html>`

const helperReportTextTemplate = `Helper Execution Report

Hi there,

Your helper "{{.HelperName}}" has completed execution.

Execution Details:
{{.Details}}

View helper: {{.AppURL}}/helpers

Best regards,
The MyFusion Helper Team`

const accountNotificationHTMLTemplate = `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Account Notification</title>
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333; max-width: 600px; margin: 0 auto; padding: 20px;">
    <div style="background: #667eea; padding: 30px; text-align: center; border-radius: 10px 10px 0 0;">
        <h1 style="color: white; margin: 0;">{{.Subject}}</h1>
    </div>
    <div style="background: #f9f9f9; padding: 30px; border-radius: 0 0 10px 10px;">
        <p>Hi there,</p>
        <div style="background: white; padding: 20px; border-radius: 5px; margin: 20px 0;">
            {{.Message}}
        </div>
        <div style="text-align: center; margin: 30px 0;">
            <a href="{{.AppURL}}" style="background: #667eea; color: white; padding: 12px 30px; text-decoration: none; border-radius: 5px; display: inline-block;">Go to Dashboard</a>
        </div>
        <p>Best regards,<br>The MyFusion Helper Team</p>
    </div>
</body>
</html>`

const accountNotificationTextTemplate = `{{.Subject}}

Hi there,

{{.Message}}

Go to dashboard: {{.AppURL}}

Best regards,
The MyFusion Helper Team`

const teamInvitationHTMLTemplate = `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Team Invitation</title>
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333; max-width: 600px; margin: 0 auto; padding: 20px;">
    <div style="background: #667eea; padding: 30px; text-align: center; border-radius: 10px 10px 0 0;">
        <h1 style="color: white; margin: 0;">You're Invited!</h1>
    </div>
    <div style="background: #f9f9f9; padding: 30px; border-radius: 0 0 10px 10px;">
        <p>Hi there,</p>
        <p><strong>{{.InviterName}}</strong> has invited you to join their team <strong>{{.TeamName}}</strong> on MyFusion Helper.</p>
        <p>Join the team to collaborate on CRM automation helpers and share access to platform connections.</p>
        <div style="text-align: center; margin: 30px 0;">
            <a href="{{.InviteURL}}" style="background: #667eea; color: white; padding: 12px 30px; text-decoration: none; border-radius: 5px; display: inline-block;">Accept Invitation</a>
        </div>
        <p>If you don't want to join this team, you can safely ignore this email.</p>
        <p>Best regards,<br>The MyFusion Helper Team</p>
    </div>
    <div style="text-align: center; margin-top: 20px; color: #999; font-size: 12px;">
        <p>If the button doesn't work, copy and paste this link into your browser:</p>
        <p style="word-break: break-all;">{{.InviteURL}}</p>
    </div>
</body>
</html>`

const teamInvitationTextTemplate = `You're Invited!

Hi there,

{{.InviterName}} has invited you to join their team "{{.TeamName}}" on MyFusion Helper.

Join the team to collaborate on CRM automation helpers and share access to platform connections.

Accept invitation: {{.InviteURL}}

If you don't want to join this team, you can safely ignore this email.

Best regards,
The MyFusion Helper Team`

func main() {
	ctx := context.Background()

	// Get stage from environment or default to dev
	stage := os.Getenv("STAGE")
	if stage == "" {
		stage = "dev"
	}

	// Load AWS config
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion("us-west-2"))
	if err != nil {
		log.Fatalf("Failed to load AWS config: %v", err)
	}

	dbClient := dynamodb.NewFromConfig(cfg)
	tableName := fmt.Sprintf("mfh-%s-email-templates", stage)

	repo := database.NewEmailTemplatesRepository(dbClient, tableName)

	log.Printf("Seeding %d system email templates to table: %s", len(systemTemplates), tableName)

	for _, tmpl := range systemTemplates {
		log.Printf("Creating template: %s (%s)", tmpl.Name, tmpl.TemplateID)
		err := repo.Create(ctx, &tmpl)
		if err != nil {
			// If template already exists, try updating it
			log.Printf("  Template exists, attempting update...")
			err = repo.Update(ctx, &tmpl)
			if err != nil {
				log.Printf("  ERROR: Failed to update template: %v", err)
				continue
			}
			log.Printf("  ✓ Updated successfully")
		} else {
			log.Printf("  ✓ Created successfully")
		}
	}

	log.Println("\n✅ All templates seeded successfully!")
}
