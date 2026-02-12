package config

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
)

type SecretsConfig struct {
	Stripe StripeSecrets `json:"stripe"`
	Groq   GroqSecrets   `json:"groq"`
	Twilio TwilioSecrets `json:"twilio"`
}

type StripeSecrets struct {
	SecretKey      string `json:"secret_key"`
	PublishableKey string `json:"publishable_key"`
	WebhookSecret  string `json:"webhook_secret"`
	PriceStart     string `json:"price_start"`
	PriceGrow      string `json:"price_grow"`
	PriceDeliver   string `json:"price_deliver"`
}

type GroqSecrets struct {
	APIKey string `json:"api_key"`
}

type TwilioSecrets struct {
	AccountSID   string `json:"account_sid"`
	AuthToken    string `json:"auth_token"`
	FromNumber   string `json:"from_number"`
	MessagingSID string `json:"messaging_sid"`
}

var (
	secrets     SecretsConfig
	secretsOnce sync.Once
	secretsErr  error
)

func LoadSecrets(ctx context.Context) (*SecretsConfig, error) {
	secretsOnce.Do(func() {
		paramName := os.Getenv("INTERNAL_SECRETS_PARAM")
		if paramName == "" {
			secretsErr = fmt.Errorf("INTERNAL_SECRETS_PARAM not set")
			return
		}

		cfg, err := config.LoadDefaultConfig(ctx)
		if err != nil {
			secretsErr = err
			return
		}

		ssmClient := ssm.NewFromConfig(cfg)
		result, err := ssmClient.GetParameter(ctx, &ssm.GetParameterInput{
			Name:           aws.String(paramName),
			WithDecryption: aws.Bool(true),
		})
		if err != nil {
			secretsErr = err
			return
		}

		secretsErr = json.Unmarshal([]byte(*result.Parameter.Value), &secrets)
	})

	if secretsErr != nil {
		return nil, secretsErr
	}
	return &secrets, nil
}
