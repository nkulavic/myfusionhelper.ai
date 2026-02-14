package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"

	authMiddleware "github.com/myfusionhelper/api/internal/middleware/auth"
	"github.com/myfusionhelper/api/cmd/handlers/chat/clients/conversations"
	"github.com/myfusionhelper/api/cmd/handlers/chat/clients/health"
	"github.com/myfusionhelper/api/cmd/handlers/chat/clients/messages"
)

// InternalSecrets holds the structure of the unified secrets JSON
type InternalSecrets struct {
	JWT struct {
		Secret        string `json:"secret"`
		RefreshSecret string `json:"refresh_secret"`
	} `json:"jwt"`
	Stripe struct {
		SecretKey      string `json:"secret_key"`
		PublishableKey string `json:"publishable_key"`
		WebhookSecret  string `json:"webhook_secret"`
	} `json:"stripe"`
	Cognito struct {
		ClientID     string `json:"client_id"`
		ClientSecret string `json:"client_secret"`
		Issuer       string `json:"issuer"`
		JwksURI      string `json:"jwks_uri"`
		Region       string `json:"region"`
		UserPoolID   string `json:"user_pool_id"`
	} `json:"cognito"`
	Groq struct {
		APIKey string `json:"api_key"`
	} `json:"groq"`
}

var (
	secrets    InternalSecrets
	groqAPIKey string
)

func init() {
	// Parse the unified secrets JSON from environment variable
	secretsJSON := os.Getenv("INTERNAL_SECRETS")
	if secretsJSON == "" {
		fmt.Println("Warning: INTERNAL_SECRETS environment variable not set")
	} else {
		if err := json.Unmarshal([]byte(secretsJSON), &secrets); err != nil {
			fmt.Printf("Warning: failed to parse INTERNAL_SECRETS: %v\n", err)
		} else {
			groqAPIKey = secrets.Groq.APIKey
			if groqAPIKey == "" {
				fmt.Println("Warning: Groq API key not found in secrets")
			}
		}
	}
}

func handler(ctx context.Context, event events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	method := event.RequestContext.HTTP.Method
	path := event.RequestContext.HTTP.Path

	// Remove /chat prefix if present
	path = strings.TrimPrefix(path, "/chat")

	// Route to appropriate handler
	switch {
	// Health endpoint (no auth required)
	case method == "GET" && path == "/health":
		return health.HandlePublic(ctx, event)

	// Conversations endpoints (auth required)
	case method == "POST" && path == "/conversations":
		return routeToProtectedHandler(ctx, event, conversations.HandleWithAuth)
	case method == "GET" && path == "/conversations":
		return routeToProtectedHandler(ctx, event, conversations.HandleListWithAuth)
	case method == "GET" && strings.HasPrefix(path, "/conversations/") && !strings.Contains(path, "/messages"):
		return routeToProtectedHandler(ctx, event, conversations.HandleGetWithAuth)
	case method == "DELETE" && strings.HasPrefix(path, "/conversations/"):
		return routeToProtectedHandler(ctx, event, conversations.HandleDeleteWithAuth)

	// Messages endpoints (auth required)
	case method == "POST" && strings.Contains(path, "/messages"):
		return routeToProtectedHandler(ctx, event, messages.HandleSendWithAuth)
	case method == "GET" && strings.Contains(path, "/messages"):
		return routeToProtectedHandler(ctx, event, messages.HandleListWithAuth)

	default:
		return authMiddleware.CreateErrorResponse(404, "Not Found"), nil
	}
}

func routeToProtectedHandler(ctx context.Context, event events.APIGatewayV2HTTPRequest, handler authMiddleware.AuthHandlerFunc) (events.APIGatewayV2HTTPResponse, error) {
	mw, err := authMiddleware.NewAuthMiddleware(ctx)
	if err != nil {
		log.Printf("Failed to create auth middleware: %v", err)
		return authMiddleware.CreateErrorResponse(500, "Internal server error"), nil
	}
	return mw.WithAuth(handler)(ctx, event)
}

func main() {
	lambda.Start(handler)
}
