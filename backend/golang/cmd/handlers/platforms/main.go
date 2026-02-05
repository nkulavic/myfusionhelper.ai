package main

import (
	"context"
	"log"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"

	authMiddleware "github.com/myfusionhelper/api/internal/middleware/auth"

	connectionsClient "github.com/myfusionhelper/api/cmd/handlers/platforms/clients/connections"
	healthClient "github.com/myfusionhelper/api/cmd/handlers/platforms/clients/health"
	platformsClient "github.com/myfusionhelper/api/cmd/handlers/platforms/clients/platforms"
)

// Handle is the main entry point for the consolidated platforms service
func Handle(ctx context.Context, event events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	path := event.RequestContext.HTTP.Path
	method := event.RequestContext.HTTP.Method

	log.Printf("Platforms Handler: path=%s method=%s", path, method)

	if method == "OPTIONS" {
		return events.APIGatewayV2HTTPResponse{
			StatusCode: 200,
			Headers: map[string]string{
				"Access-Control-Allow-Origin":  "*",
				"Access-Control-Allow-Methods": "GET, POST, PUT, DELETE, OPTIONS",
				"Access-Control-Allow-Headers": "Content-Type, Authorization, X-API-Key, X-Account-Context",
			},
			Body: "",
		}, nil
	}

	switch {
	// Public endpoints
	case path == "/platforms/health" && method == "GET":
		return healthClient.Handle(ctx, event)
	case path == "/platforms/oauth/callback" && method == "GET":
		return connectionsClient.HandlePublic(ctx, event)

	// Platform listing
	case path == "/platforms" && method == "GET":
		return routeToProtectedHandler(ctx, event, platformsClient.HandleWithAuth)
	case strings.HasPrefix(path, "/platforms/") && event.PathParameters["platform_id"] != "" &&
		!strings.Contains(path, "/connections") && !strings.Contains(path, "/oauth") &&
		!strings.Contains(path, "/health") && method == "GET":
		return routeToProtectedHandler(ctx, event, platformsClient.HandleWithAuth)

	// Platform connections (scoped to platform)
	case strings.HasPrefix(path, "/platforms/") && strings.HasSuffix(path, "/connections") &&
		event.PathParameters["platform_id"] != "" && event.PathParameters["connection_id"] == "" &&
		(method == "GET" || method == "POST"):
		return routeToProtectedHandler(ctx, event, connectionsClient.HandleWithAuth)
	case strings.HasPrefix(path, "/platforms/") && strings.Contains(path, "/connections/") &&
		event.PathParameters["platform_id"] != "" && event.PathParameters["connection_id"] != "" &&
		!strings.HasSuffix(path, "/test") && (method == "GET" || method == "PUT" || method == "DELETE"):
		return routeToProtectedHandler(ctx, event, connectionsClient.HandleWithAuth)
	case strings.HasPrefix(path, "/platforms/") && strings.HasSuffix(path, "/test") &&
		event.PathParameters["platform_id"] != "" && event.PathParameters["connection_id"] != "" &&
		method == "POST":
		return routeToProtectedHandler(ctx, event, connectionsClient.HandleWithAuth)

	// List ALL connections (no platform filter)
	case path == "/platform-connections" && method == "GET":
		return routeToProtectedHandler(ctx, event, connectionsClient.HandleWithAuth)

	// OAuth flow
	case strings.HasPrefix(path, "/platforms/") && strings.HasSuffix(path, "/oauth/start") &&
		event.PathParameters["platform_id"] != "" && method == "POST":
		return routeToProtectedHandler(ctx, event, connectionsClient.HandleWithAuth)

	default:
		return authMiddleware.CreateErrorResponse(404, "Not Found"), nil
	}
}

func routeToProtectedHandler(ctx context.Context, event events.APIGatewayV2HTTPRequest, handler authMiddleware.AuthHandlerFunc) (events.APIGatewayV2HTTPResponse, error) {
	authMiddlewareInstance, err := authMiddleware.NewAuthMiddleware(ctx)
	if err != nil {
		log.Printf("Failed to create auth middleware: %v", err)
		return authMiddleware.CreateErrorResponse(500, "Internal server error"), nil
	}
	return authMiddlewareInstance.WithAuth(handler)(ctx, event)
}

func main() {
	lambda.Start(Handle)
}
