package services

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"html/template"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	"github.com/myfusionhelper/api/internal/email"
)

// EmailService provides high-level email operations with template management
type EmailService struct {
	sesClient *email.SESClient
	dbClient  *dynamodb.Client
	validator *email.EmailValidator

	// Table names
	logsTable          string
	templatesTable     string
	verificationsTable string
}

// EmailTemplate represents an email template stored in DynamoDB
type EmailTemplate struct {
	TemplateID   string                 `json:"template_id" dynamodbav:"template_id"`
	AccountID    string                 `json:"account_id,omitempty" dynamodbav:"account_id,omitempty"` // Empty for system templates
	Name         string                 `json:"name" dynamodbav:"name"`
	Subject      string                 `json:"subject" dynamodbav:"subject"`
	HTMLTemplate string                 `json:"html_template" dynamodbav:"html_template"`
	TextTemplate string                 `json:"text_template" dynamodbav:"text_template"`
	Variables    []EmailVariable        `json:"variables,omitempty" dynamodbav:"variables,omitempty"`
	IsSystem     bool                   `json:"is_system" dynamodbav:"is_system"`
	IsActive     bool                   `json:"is_active" dynamodbav:"is_active"`
	CreatedAt    string                 `json:"created_at" dynamodbav:"created_at"`
	UpdatedAt    string                 `json:"updated_at" dynamodbav:"updated_at"`
}

// EmailVariable represents a template variable
type EmailVariable struct {
	Name        string `json:"name" dynamodbav:"name"`
	Description string `json:"description" dynamodbav:"description"`
	Required    bool   `json:"required" dynamodbav:"required"`
}

// EmailLog represents a sent email record
type EmailLog struct {
	EmailID        string            `json:"email_id" dynamodbav:"email_id"`
	AccountID      string            `json:"account_id" dynamodbav:"account_id"`
	RecipientEmail string            `json:"recipient_email" dynamodbav:"recipient_email"`
	Subject        string            `json:"subject" dynamodbav:"subject"`
	TemplateID     string            `json:"template_id,omitempty" dynamodbav:"template_id,omitempty"`
	Status         string            `json:"status" dynamodbav:"status"` // sent, failed, bounced
	MessageID      string            `json:"message_id,omitempty" dynamodbav:"message_id,omitempty"`
	ErrorMessage   string            `json:"error_message,omitempty" dynamodbav:"error_message,omitempty"`
	CreatedAt      string            `json:"created_at" dynamodbav:"created_at"`
	SentAt         string            `json:"sent_at,omitempty" dynamodbav:"sent_at,omitempty"`
	TTL            int64             `json:"ttl" dynamodbav:"ttl"` // Auto-delete after 90 days
}

// EmailVerification represents an email verification record
type EmailVerification struct {
	VerificationID string `json:"verification_id" dynamodbav:"verification_id"`
	Email          string `json:"email" dynamodbav:"email"`
	Token          string `json:"token" dynamodbav:"token"`
	ExpiresAt      int64  `json:"expires_at" dynamodbav:"expires_at"` // Unix timestamp, also used for TTL
	CreatedAt      string `json:"created_at" dynamodbav:"created_at"`
	VerifiedAt     string `json:"verified_at,omitempty" dynamodbav:"verified_at,omitempty"`
	Status         string `json:"status" dynamodbav:"status"` // pending, verified, expired
}

// NewEmailService creates a new email service
func NewEmailService(ctx context.Context, dbClient *dynamodb.Client, logsTable, templatesTable, verificationsTable string) (*EmailService, error) {
	sesClient, err := email.NewSESClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create SES client: %w", err)
	}

	return &EmailService{
		sesClient:          sesClient,
		dbClient:           dbClient,
		validator:          email.NewEmailValidator(),
		logsTable:          logsTable,
		templatesTable:     templatesTable,
		verificationsTable: verificationsTable,
	}, nil
}

// SendEmail sends an email and logs it
func (s *EmailService) SendEmail(ctx context.Context, accountID string, msg *email.EmailMessage) (*email.EmailResult, error) {
	// Validate email addresses
	for _, recipient := range msg.To {
		validation := s.validator.ValidateEmail(recipient)
		if !validation.IsValid {
			return &email.EmailResult{
				Success: false,
				Error:   validation.Error,
			}, fmt.Errorf("invalid recipient email: %s", validation.Error)
		}
	}

	// Send email via SES
	result, err := s.sesClient.SendEmail(ctx, *msg)

	// Log the email (even if it failed)
	logEntry := &EmailLog{
		EmailID:        generateID("email"),
		AccountID:      accountID,
		RecipientEmail: msg.To[0], // Log first recipient
		Subject:        msg.Subject,
		CreatedAt:      time.Now().UTC().Format(time.RFC3339),
		TTL:            time.Now().Add(90 * 24 * time.Hour).Unix(), // 90 days
	}

	if err != nil {
		logEntry.Status = "failed"
		logEntry.ErrorMessage = err.Error()
	} else {
		logEntry.Status = "sent"
		logEntry.MessageID = result.MessageID
		logEntry.SentAt = time.Now().UTC().Format(time.RFC3339)
	}

	// Save log to DynamoDB (don't fail the send if logging fails)
	if logErr := s.logEmail(ctx, logEntry); logErr != nil {
		log.Printf("WARNING: Failed to log email: %v", logErr)
	}

	return result, err
}

// SendTemplatedEmail sends an email using a stored template
func (s *EmailService) SendTemplatedEmail(ctx context.Context, accountID, templateID string, data map[string]interface{}, recipients []string) (*email.EmailResult, error) {
	// Get template from DynamoDB
	tmpl, err := s.GetTemplate(ctx, templateID)
	if err != nil {
		return &email.EmailResult{
			Success: false,
			Error:   fmt.Sprintf("failed to get template: %v", err),
		}, err
	}

	if !tmpl.IsActive {
		return &email.EmailResult{
			Success: false,
			Error:   "template is not active",
		}, fmt.Errorf("template is not active")
	}

	// Render template
	htmlBody, textBody, err := s.renderTemplate(tmpl, data)
	if err != nil {
		return &email.EmailResult{
			Success: false,
			Error:   fmt.Sprintf("failed to render template: %v", err),
		}, err
	}

	// Render subject
	subject, err := s.renderString(tmpl.Subject, data)
	if err != nil {
		return &email.EmailResult{
			Success: false,
			Error:   fmt.Sprintf("failed to render subject: %v", err),
		}, err
	}

	// Send email
	msg := &email.EmailMessage{
		To:       recipients,
		Subject:  subject,
		HTMLBody: htmlBody,
		TextBody: textBody,
		Tags: map[string]string{
			"template_id": templateID,
			"account_id":  accountID,
		},
	}

	return s.SendEmail(ctx, accountID, msg)
}

// Convenience methods for system emails

// SendWelcomeEmail sends a welcome email to a new user
func (s *EmailService) SendWelcomeEmail(ctx context.Context, accountID, email, firstName string) error {
	data := map[string]interface{}{
		"FirstName": firstName,
		"Email":     email,
		"AppURL":    "https://app.myfusionhelper.ai",
	}

	_, err := s.SendTemplatedEmail(ctx, accountID, "welcome", data, []string{email})
	return err
}

// SendPasswordResetEmail sends a password reset email
func (s *EmailService) SendPasswordResetEmail(ctx context.Context, accountID, email, resetToken string) error {
	data := map[string]interface{}{
		"Email":     email,
		"ResetURL":  fmt.Sprintf("https://app.myfusionhelper.ai/reset-password?token=%s", resetToken),
		"ExpiresIn": "24 hours",
	}

	_, err := s.SendTemplatedEmail(ctx, accountID, "password_reset", data, []string{email})
	return err
}

// SendEmailVerificationEmail sends an email verification email
func (s *EmailService) SendEmailVerificationEmail(ctx context.Context, accountID, email, verificationToken string) error {
	data := map[string]interface{}{
		"Email":         email,
		"VerificationURL": fmt.Sprintf("https://app.myfusionhelper.ai/verify-email?token=%s", verificationToken),
		"ExpiresIn":     "24 hours",
	}

	_, err := s.SendTemplatedEmail(ctx, accountID, "email_verification", data, []string{email})
	return err
}

// SendHelperReportEmail sends a helper execution report
func (s *EmailService) SendHelperReportEmail(ctx context.Context, accountID, email, helperName string, executionDetails map[string]interface{}) error {
	data := map[string]interface{}{
		"HelperName": helperName,
		"Details":    executionDetails,
		"AppURL":     "https://app.myfusionhelper.ai",
	}

	_, err := s.SendTemplatedEmail(ctx, accountID, "helper_report", data, []string{email})
	return err
}

// SendAccountNotificationEmail sends an account notification
func (s *EmailService) SendAccountNotificationEmail(ctx context.Context, accountID, email, subject, message string) error {
	data := map[string]interface{}{
		"Subject": subject,
		"Message": message,
		"AppURL":  "https://app.myfusionhelper.ai",
	}

	_, err := s.SendTemplatedEmail(ctx, accountID, "account_notification", data, []string{email})
	return err
}

// SendTeamInvitationEmail sends a team invitation email
func (s *EmailService) SendTeamInvitationEmail(ctx context.Context, accountID, inviteeEmail, inviterName, teamName, inviteToken string) error {
	data := map[string]interface{}{
		"InviterName": inviterName,
		"TeamName":    teamName,
		"InviteURL":   fmt.Sprintf("https://app.myfusionhelper.ai/join-team?token=%s", inviteToken),
	}

	_, err := s.SendTemplatedEmail(ctx, accountID, "team_invitation", data, []string{inviteeEmail})
	return err
}

// Template management methods

// CreateTemplate creates a new email template
func (s *EmailService) CreateTemplate(ctx context.Context, tmpl *EmailTemplate) error {
	if tmpl.TemplateID == "" {
		tmpl.TemplateID = generateID("tmpl")
	}

	now := time.Now().UTC().Format(time.RFC3339)
	tmpl.CreatedAt = now
	tmpl.UpdatedAt = now
	tmpl.IsActive = true

	av, err := attributevalue.MarshalMap(tmpl)
	if err != nil {
		return fmt.Errorf("failed to marshal template: %w", err)
	}

	_, err = s.dbClient.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(s.templatesTable),
		Item:      av,
	})

	return err
}

// GetTemplate retrieves a template by ID
func (s *EmailService) GetTemplate(ctx context.Context, templateID string) (*EmailTemplate, error) {
	result, err := s.dbClient.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(s.templatesTable),
		Key: map[string]types.AttributeValue{
			"template_id": &types.AttributeValueMemberS{Value: templateID},
		},
	})
	if err != nil {
		return nil, err
	}

	if result.Item == nil {
		return nil, fmt.Errorf("template not found")
	}

	var tmpl EmailTemplate
	if err := attributevalue.UnmarshalMap(result.Item, &tmpl); err != nil {
		return nil, fmt.Errorf("failed to unmarshal template: %w", err)
	}

	return &tmpl, nil
}

// UpdateTemplate updates an existing template
func (s *EmailService) UpdateTemplate(ctx context.Context, tmpl *EmailTemplate) error {
	tmpl.UpdatedAt = time.Now().UTC().Format(time.RFC3339)

	av, err := attributevalue.MarshalMap(tmpl)
	if err != nil {
		return fmt.Errorf("failed to marshal template: %w", err)
	}

	_, err = s.dbClient.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(s.templatesTable),
		Item:      av,
	})

	return err
}

// DeleteTemplate deletes a template (soft delete by marking inactive)
func (s *EmailService) DeleteTemplate(ctx context.Context, templateID string) error {
	_, err := s.dbClient.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(s.templatesTable),
		Key: map[string]types.AttributeValue{
			"template_id": &types.AttributeValueMemberS{Value: templateID},
		},
		UpdateExpression: aws.String("SET is_active = :inactive, updated_at = :now"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":inactive": &types.AttributeValueMemberBOOL{Value: false},
			":now":      &types.AttributeValueMemberS{Value: time.Now().UTC().Format(time.RFC3339)},
		},
	})

	return err
}

// ListTemplates lists all templates for an account (or system templates)
func (s *EmailService) ListTemplates(ctx context.Context, accountID string, includeSystem bool) ([]*EmailTemplate, error) {
	var templates []*EmailTemplate

	// Query user templates
	if accountID != "" {
		result, err := s.dbClient.Query(ctx, &dynamodb.QueryInput{
			TableName:              aws.String(s.templatesTable),
			IndexName:              aws.String("AccountIdIndex"),
			KeyConditionExpression: aws.String("account_id = :account_id"),
			ExpressionAttributeValues: map[string]types.AttributeValue{
				":account_id": &types.AttributeValueMemberS{Value: accountID},
			},
		})
		if err != nil {
			return nil, err
		}

		for _, item := range result.Items {
			var tmpl EmailTemplate
			if err := attributevalue.UnmarshalMap(item, &tmpl); err != nil {
				continue
			}
			templates = append(templates, &tmpl)
		}
	}

	// Query system templates if requested
	if includeSystem {
		// System templates have empty account_id
		// Would need to scan with filter for is_system = true
		// For now, skip implementation
	}

	return templates, nil
}

// Email verification methods

// CreateEmailVerification creates a new email verification record
func (s *EmailService) CreateEmailVerification(ctx context.Context, email string) (*EmailVerification, error) {
	token := generateToken(32)
	verification := &EmailVerification{
		VerificationID: generateID("verify"),
		Email:          email,
		Token:          token,
		ExpiresAt:      time.Now().Add(24 * time.Hour).Unix(),
		CreatedAt:      time.Now().UTC().Format(time.RFC3339),
		Status:         "pending",
	}

	av, err := attributevalue.MarshalMap(verification)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal verification: %w", err)
	}

	_, err = s.dbClient.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(s.verificationsTable),
		Item:      av,
	})
	if err != nil {
		return nil, err
	}

	return verification, nil
}

// VerifyEmail verifies an email using a token
func (s *EmailService) VerifyEmail(ctx context.Context, token string) error {
	// Query by token (would need a GSI on token field)
	// For now, simplified implementation
	_, err := s.dbClient.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(s.verificationsTable),
		Key: map[string]types.AttributeValue{
			"verification_id": &types.AttributeValueMemberS{Value: token}, // Simplified
		},
		UpdateExpression: aws.String("SET #status = :verified, verified_at = :now"),
		ExpressionAttributeNames: map[string]string{
			"#status": "status",
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":verified": &types.AttributeValueMemberS{Value: "verified"},
			":now":      &types.AttributeValueMemberS{Value: time.Now().UTC().Format(time.RFC3339)},
		},
	})

	return err
}

// Email history methods

// GetEmailHistory retrieves email history for an account
func (s *EmailService) GetEmailHistory(ctx context.Context, accountID string, limit int) ([]*EmailLog, error) {
	if limit <= 0 {
		limit = 50
	}

	result, err := s.dbClient.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(s.logsTable),
		IndexName:              aws.String("AccountIdIndex"),
		KeyConditionExpression: aws.String("account_id = :account_id"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":account_id": &types.AttributeValueMemberS{Value: accountID},
		},
		Limit:            aws.Int32(int32(limit)),
		ScanIndexForward: aws.Bool(false), // Most recent first
	})
	if err != nil {
		return nil, err
	}

	var logs []*EmailLog
	for _, item := range result.Items {
		var log EmailLog
		if err := attributevalue.UnmarshalMap(item, &log); err != nil {
			continue
		}
		logs = append(logs, &log)
	}

	return logs, nil
}

// Private helper methods

func (s *EmailService) logEmail(ctx context.Context, log *EmailLog) error {
	av, err := attributevalue.MarshalMap(log)
	if err != nil {
		return err
	}

	_, err = s.dbClient.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(s.logsTable),
		Item:      av,
	})

	return err
}

func (s *EmailService) renderTemplate(tmpl *EmailTemplate, data map[string]interface{}) (htmlBody, textBody string, err error) {
	// Render HTML template
	if tmpl.HTMLTemplate != "" {
		htmlTmpl, err := template.New("html").Parse(tmpl.HTMLTemplate)
		if err != nil {
			return "", "", fmt.Errorf("failed to parse HTML template: %w", err)
		}

		var htmlBuffer bytes.Buffer
		if err := htmlTmpl.Execute(&htmlBuffer, data); err != nil {
			return "", "", fmt.Errorf("failed to execute HTML template: %w", err)
		}
		htmlBody = htmlBuffer.String()
	}

	// Render text template
	if tmpl.TextTemplate != "" {
		textTmpl, err := template.New("text").Parse(tmpl.TextTemplate)
		if err != nil {
			return "", "", fmt.Errorf("failed to parse text template: %w", err)
		}

		var textBuffer bytes.Buffer
		if err := textTmpl.Execute(&textBuffer, data); err != nil {
			return "", "", fmt.Errorf("failed to execute text template: %w", err)
		}
		textBody = textBuffer.String()
	}

	return htmlBody, textBody, nil
}

func (s *EmailService) renderString(tmplStr string, data map[string]interface{}) (string, error) {
	tmpl, err := template.New("string").Parse(tmplStr)
	if err != nil {
		return "", err
	}

	var buffer bytes.Buffer
	if err := tmpl.Execute(&buffer, data); err != nil {
		return "", err
	}

	return buffer.String(), nil
}

// Utility functions

func generateID(prefix string) string {
	b := make([]byte, 8)
	rand.Read(b)
	return fmt.Sprintf("%s_%s", prefix, hex.EncodeToString(b))
}

func generateToken(length int) string {
	b := make([]byte, length)
	rand.Read(b)
	return hex.EncodeToString(b)
}
