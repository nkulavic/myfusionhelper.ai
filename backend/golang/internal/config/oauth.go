package config

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
)

// PlatformOAuthConfig holds OAuth credentials and URLs for a single platform.
type PlatformOAuthConfig struct {
	ClientID    string   `json:"client_id"`
	ClientSecret string  `json:"client_secret"`
	Slug        string   `json:"slug"`
	AuthURL     string   `json:"auth_url"`
	TokenURL    string   `json:"token_url"`
	UserInfoURL string   `json:"user_info_url,omitempty"`
	Scopes      []string `json:"scopes,omitempty"`
}

// OAuthCredentials maps platform slug to its OAuth config.
type OAuthCredentials map[string]PlatformOAuthConfig

var (
	oauthCreds     OAuthCredentials
	oauthCredsOnce sync.Once
	oauthCredsErr  error
)

// LoadOAuthCredentials loads the unified OAuth credentials from SSM.
// Uses sync.Once to cache after first call.
// Reads from the OAUTH_CREDENTIALS_PARAM env var (SSM parameter name).
func LoadOAuthCredentials(ctx context.Context) (OAuthCredentials, error) {
	oauthCredsOnce.Do(func() {
		paramName := os.Getenv("OAUTH_CREDENTIALS_PARAM")
		if paramName == "" {
			// Fallback: construct from STAGE
			stage := os.Getenv("STAGE")
			if stage == "" {
				stage = "dev"
			}
			paramName = fmt.Sprintf("/myfusionhelper/%s/platforms/oauth/credentials", stage)
		}

		cfg, err := awsconfig.LoadDefaultConfig(ctx)
		if err != nil {
			oauthCredsErr = fmt.Errorf("failed to load AWS config: %w", err)
			return
		}

		ssmClient := ssm.NewFromConfig(cfg)
		result, err := ssmClient.GetParameter(ctx, &ssm.GetParameterInput{
			Name:           aws.String(paramName),
			WithDecryption: aws.Bool(true),
		})
		if err != nil {
			oauthCredsErr = fmt.Errorf("failed to get OAuth credentials from SSM (%s): %w", paramName, err)
			return
		}

		oauthCreds = make(OAuthCredentials)
		oauthCredsErr = json.Unmarshal([]byte(*result.Parameter.Value), &oauthCreds)
	})

	if oauthCredsErr != nil {
		return nil, oauthCredsErr
	}
	return oauthCreds, nil
}

// GetPlatformOAuth returns OAuth config for a specific platform slug.
// Returns an error if credentials haven't been loaded or the platform isn't found.
func GetPlatformOAuth(ctx context.Context, slug string) (*PlatformOAuthConfig, error) {
	creds, err := LoadOAuthCredentials(ctx)
	if err != nil {
		return nil, err
	}

	config, ok := creds[slug]
	if !ok {
		return nil, fmt.Errorf("no OAuth credentials configured for platform: %s", slug)
	}

	if config.ClientID == "" || config.ClientSecret == "" {
		return nil, fmt.Errorf("incomplete OAuth credentials for platform: %s", slug)
	}

	return &config, nil
}
