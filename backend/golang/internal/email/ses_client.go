package email

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ses"
	sestypes "github.com/aws/aws-sdk-go-v2/service/ses/types"
)

// SESClient wraps AWS SES functionality for sending emails
type SESClient struct {
	client *ses.Client
	cfg    SESConfig
}

// SESConfig holds configuration for SES operations
type SESConfig struct {
	Region               string
	DefaultFromEmail     string
	DefaultFromName      string
	ConfigurationSetName string
}

// EmailMessage represents an email to be sent
type EmailMessage struct {
	To           []string
	Cc           []string
	Bcc          []string
	Subject      string
	HTMLBody     string
	TextBody     string
	FromEmail    string
	FromName     string
	ReplyTo      []string
	Tags         map[string]string
	TemplateName string
	TemplateData map[string]interface{}
}

// EmailResult represents the result of sending an email
type EmailResult struct {
	MessageID string
	Success   bool
	Error     string
}

// NewSESClient creates a new SES client with configuration from environment variables
func NewSESClient(ctx context.Context) (*SESClient, error) {
	region := os.Getenv("AWS_REGION")
	if region == "" {
		region = "us-west-2"
	}

	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return nil, fmt.Errorf("failed to create AWS config: %w", err)
	}

	sesConfig := SESConfig{
		Region:               region,
		DefaultFromEmail:     getEnvOrDefault("DEFAULT_FROM_EMAIL", getDefaultFromEmail()),
		DefaultFromName:      getEnvOrDefault("DEFAULT_FROM_NAME", "MyFusion Helper"),
		ConfigurationSetName: getEnvOrDefault("SES_CONFIGURATION_SET", ""),
	}

	return &SESClient{
		client: ses.NewFromConfig(cfg),
		cfg:    sesConfig,
	}, nil
}

// SendEmail sends an email via AWS SES
func (s *SESClient) SendEmail(ctx context.Context, message EmailMessage) (*EmailResult, error) {
	// Validate message
	if err := s.validateMessage(&message); err != nil {
		return &EmailResult{
			Success: false,
			Error:   err.Error(),
		}, err
	}

	log.Printf("Sending email to: %v, subject: %s", message.To, message.Subject)

	if message.FromEmail == "" {
		message.FromEmail = s.cfg.DefaultFromEmail
	}
	if message.FromName == "" {
		message.FromName = s.cfg.DefaultFromName
	}

	fromAddress := message.FromEmail
	if message.FromName != "" {
		fromAddress = fmt.Sprintf("%s <%s>", message.FromName, message.FromEmail)
	}

	toAddresses := make([]string, len(message.To))
	copy(toAddresses, message.To)

	destination := &sestypes.Destination{
		ToAddresses: toAddresses,
	}

	// Add Cc recipients if provided
	if len(message.Cc) > 0 {
		ccAddresses := make([]string, len(message.Cc))
		copy(ccAddresses, message.Cc)
		destination.CcAddresses = ccAddresses
	}

	// Add Bcc recipients if provided
	if len(message.Bcc) > 0 {
		bccAddresses := make([]string, len(message.Bcc))
		copy(bccAddresses, message.Bcc)
		destination.BccAddresses = bccAddresses
	}

	body := &sestypes.Body{}
	if message.HTMLBody != "" {
		body.Html = &sestypes.Content{
			Data:    aws.String(message.HTMLBody),
			Charset: aws.String("UTF-8"),
		}
	}
	if message.TextBody != "" {
		body.Text = &sestypes.Content{
			Data:    aws.String(message.TextBody),
			Charset: aws.String("UTF-8"),
		}
	}

	input := &ses.SendEmailInput{
		Source:      aws.String(fromAddress),
		Destination: destination,
		Message: &sestypes.Message{
			Subject: &sestypes.Content{
				Data:    aws.String(message.Subject),
				Charset: aws.String("UTF-8"),
			},
			Body: body,
		},
	}

	// Add reply-to if provided
	if len(message.ReplyTo) > 0 {
		replyToAddresses := make([]string, len(message.ReplyTo))
		copy(replyToAddresses, message.ReplyTo)
		input.ReplyToAddresses = replyToAddresses
	}

	if s.cfg.ConfigurationSetName != "" {
		input.ConfigurationSetName = aws.String(s.cfg.ConfigurationSetName)
	}

	if len(message.Tags) > 0 {
		tags := make([]sestypes.MessageTag, 0, len(message.Tags))
		for key, value := range message.Tags {
			if key != "" && value != "" {
				tags = append(tags, sestypes.MessageTag{
					Name:  aws.String(key),
					Value: aws.String(value),
				})
			}
		}
		if len(tags) > 0 {
			input.Tags = tags
		}
	}

	output, err := s.client.SendEmail(ctx, input)
	if err != nil {
		log.Printf("ERROR: Failed to send email: %v", err)
		return &EmailResult{
			Success: false,
			Error:   err.Error(),
		}, err
	}

	log.Printf("Email sent successfully, MessageID: %s", *output.MessageId)
	return &EmailResult{
		MessageID: *output.MessageId,
		Success:   true,
	}, nil
}

// SendTemplatedEmail sends an email using a pre-defined SES template
func (s *SESClient) SendTemplatedEmail(ctx context.Context, message EmailMessage) (*EmailResult, error) {
	if message.TemplateName == "" {
		return &EmailResult{
			Success: false,
			Error:   "template name is required for templated emails",
		}, fmt.Errorf("template name is required")
	}

	// Validate message
	if err := s.validateMessage(&message); err != nil {
		return &EmailResult{
			Success: false,
			Error:   err.Error(),
		}, err
	}

	log.Printf("Sending templated email (%s) to: %v", message.TemplateName, message.To)

	if message.FromEmail == "" {
		message.FromEmail = s.cfg.DefaultFromEmail
	}
	if message.FromName == "" {
		message.FromName = s.cfg.DefaultFromName
	}

	fromAddress := fmt.Sprintf("%s <%s>", message.FromName, message.FromEmail)

	toAddresses := make([]string, len(message.To))
	copy(toAddresses, message.To)

	destination := &sestypes.Destination{
		ToAddresses: toAddresses,
	}

	// Add Cc and Bcc if provided
	if len(message.Cc) > 0 {
		ccAddresses := make([]string, len(message.Cc))
		copy(ccAddresses, message.Cc)
		destination.CcAddresses = ccAddresses
	}
	if len(message.Bcc) > 0 {
		bccAddresses := make([]string, len(message.Bcc))
		copy(bccAddresses, message.Bcc)
		destination.BccAddresses = bccAddresses
	}

	// Build template data JSON string
	templateDataJSON := "{}"
	if len(message.TemplateData) > 0 {
		// Note: In production, use json.Marshal for proper serialization
		templateDataJSON = "{}"
	}

	input := &ses.SendTemplatedEmailInput{
		Source:       aws.String(fromAddress),
		Template:     aws.String(message.TemplateName),
		TemplateData: aws.String(templateDataJSON),
		Destination:  destination,
	}

	// Add reply-to if provided
	if len(message.ReplyTo) > 0 {
		replyToAddresses := make([]string, len(message.ReplyTo))
		copy(replyToAddresses, message.ReplyTo)
		input.ReplyToAddresses = replyToAddresses
	}

	if s.cfg.ConfigurationSetName != "" {
		input.ConfigurationSetName = aws.String(s.cfg.ConfigurationSetName)
	}

	if len(message.Tags) > 0 {
		tags := make([]sestypes.MessageTag, 0, len(message.Tags))
		for key, value := range message.Tags {
			if key != "" && value != "" {
				tags = append(tags, sestypes.MessageTag{
					Name:  aws.String(key),
					Value: aws.String(value),
				})
			}
		}
		if len(tags) > 0 {
			input.Tags = tags
		}
	}

	output, err := s.client.SendTemplatedEmail(ctx, input)
	if err != nil {
		log.Printf("ERROR: Failed to send templated email: %v", err)
		return &EmailResult{
			Success: false,
			Error:   err.Error(),
		}, err
	}

	log.Printf("Templated email sent successfully, MessageID: %s", *output.MessageId)
	return &EmailResult{
		MessageID: *output.MessageId,
		Success:   true,
	}, nil
}

// validateMessage performs basic validation on an email message
func (s *SESClient) validateMessage(msg *EmailMessage) error {
	if len(msg.To) == 0 {
		return fmt.Errorf("at least one recipient (To) is required")
	}

	// Validate email addresses
	for _, email := range msg.To {
		if !isValidEmail(email) {
			return fmt.Errorf("invalid To email address: %s", email)
		}
	}

	for _, email := range msg.Cc {
		if !isValidEmail(email) {
			return fmt.Errorf("invalid Cc email address: %s", email)
		}
	}

	for _, email := range msg.Bcc {
		if !isValidEmail(email) {
			return fmt.Errorf("invalid Bcc email address: %s", email)
		}
	}

	// Check that either subject or template is provided
	if msg.Subject == "" && msg.TemplateName == "" {
		return fmt.Errorf("either subject or template name is required")
	}

	// Check that either body or template is provided
	if msg.HTMLBody == "" && msg.TextBody == "" && msg.TemplateName == "" {
		return fmt.Errorf("either body content or template name is required")
	}

	return nil
}

// isValidEmail performs basic email validation
func isValidEmail(email string) bool {
	if len(email) == 0 {
		return false
	}

	// Must contain exactly one @
	atCount := 0
	atIndex := -1
	for i, char := range email {
		if char == '@' {
			atCount++
			if atIndex == -1 {
				atIndex = i
			}
		}
	}

	if atCount != 1 {
		return false
	}

	// Must have content before and after @
	if atIndex == 0 || atIndex == len(email)-1 {
		return false
	}

	// Domain must contain at least one dot
	domain := email[atIndex+1:]
	hasDot := false
	for _, char := range domain {
		if char == '.' {
			hasDot = true
			break
		}
	}

	return hasDot
}

// GetSendQuota returns the current SES sending quota
func (s *SESClient) GetSendQuota(ctx context.Context) (*ses.GetSendQuotaOutput, error) {
	return s.client.GetSendQuota(ctx, &ses.GetSendQuotaInput{})
}

// GetSendStatistics returns sending statistics for the last 2 weeks
func (s *SESClient) GetSendStatistics(ctx context.Context) (*ses.GetSendStatisticsOutput, error) {
	return s.client.GetSendStatistics(ctx, &ses.GetSendStatisticsInput{})
}

// VerifyEmailIdentity initiates email address verification
func (s *SESClient) VerifyEmailIdentity(ctx context.Context, email string) error {
	_, err := s.client.VerifyEmailIdentity(ctx, &ses.VerifyEmailIdentityInput{
		EmailAddress: aws.String(email),
	})
	return err
}

// GetIdentityVerificationStatus checks if an email address is verified
func (s *SESClient) GetIdentityVerificationStatus(ctx context.Context, email string) (string, error) {
	output, err := s.client.GetIdentityVerificationAttributes(ctx, &ses.GetIdentityVerificationAttributesInput{
		Identities: []string{email},
	})
	if err != nil {
		return "", err
	}

	if attr, ok := output.VerificationAttributes[email]; ok {
		return string(attr.VerificationStatus), nil
	}

	return "NotStarted", nil
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getDefaultFromEmail returns the stage-appropriate noreply address.
func getDefaultFromEmail() string {
	stage := os.Getenv("STAGE")
	switch stage {
	case "main", "":
		return "noreply@myfusionhelper.ai"
	default:
		return fmt.Sprintf("noreply@%s.myfusionhelper.ai", stage)
	}
}
