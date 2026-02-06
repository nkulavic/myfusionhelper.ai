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
	To        []string
	Subject   string
	HTMLBody  string
	TextBody  string
	FromEmail string
	FromName  string
	Tags      map[string]string
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
		DefaultFromEmail:     getEnvOrDefault("DEFAULT_FROM_EMAIL", "noreply@myfusionhelper.ai"),
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
		Source: aws.String(fromAddress),
		Destination: &sestypes.Destination{
			ToAddresses: toAddresses,
		},
		Message: &sestypes.Message{
			Subject: &sestypes.Content{
				Data:    aws.String(message.Subject),
				Charset: aws.String("UTF-8"),
			},
			Body: body,
		},
	}

	if s.cfg.ConfigurationSetName != "" {
		input.ConfigurationSetName = aws.String(s.cfg.ConfigurationSetName)
	}

	if len(message.Tags) > 0 {
		tags := make([]sestypes.MessageTag, 0, len(message.Tags))
		for key, value := range message.Tags {
			tags = append(tags, sestypes.MessageTag{
				Name:  aws.String(key),
				Value: aws.String(value),
			})
		}
		input.Tags = tags
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

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
