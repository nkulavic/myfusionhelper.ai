package main

import (
	"context"
	"log"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"

	authMiddleware "github.com/myfusionhelper/api/internal/middleware/auth"

	// Protected endpoints (require auth)
	statusClient "github.com/myfusionhelper/api/cmd/handlers/auth/clients/status"
	logoutClient "github.com/myfusionhelper/api/cmd/handlers/auth/clients/logout"

	// Public endpoints (no auth required)
	healthClient "github.com/myfusionhelper/api/cmd/handlers/auth/clients/health"
	loginClient "github.com/myfusionhelper/api/cmd/handlers/auth/clients/login"
	registerClient "github.com/myfusionhelper/api/cmd/handlers/auth/clients/register"
	refreshClient "github.com/myfusionhelper/api/cmd/handlers/auth/clients/refresh"
)

// Handle is the main entry point for the consolidated auth service
func Handle(ctx context.Context, event events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	log.Printf("Auth Handler: path=%s method=%s", event.RequestContext.HTTP.Path, event.RequestContext.HTTP.Method)

	// Handle OPTIONS request for CORS
	if event.RequestContext.HTTP.Method == "OPTIONS" {
		return events.APIGatewayV2HTTPResponse{
			StatusCode: 200,
			Headers: map[string]string{
				"Access-Control-Allow-Origin":  "*",
				"Access-Control-Allow-Methods": "GET, POST, OPTIONS",
				"Access-Control-Allow-Headers": "Content-Type, Authorization, X-Account-Context",
			},
			Body: "",
		}, nil
	}

	// Route to handler based on path
	switch event.RequestContext.HTTP.Path {
	// Protected endpoints
	case "/auth/status":
		return routeToProtectedHandler(ctx, event, statusClient.HandleWithAuth)
	case "/auth/logout":
		return routeToProtectedHandler(ctx, event, logoutClient.HandleWithAuth)

	// Public endpoints
	case "/auth/health":
		return healthClient.Handle(ctx, event)
	case "/auth/login":
		return loginClient.Handle(ctx, event)
	case "/auth/register":
		return registerClient.Handle(ctx, event)
	case "/auth/refresh":
		return refreshClient.Handle(ctx, event)

	default:
		log.Printf("No handler found for path: %s", event.RequestContext.HTTP.Path)
		return authMiddleware.CreateErrorResponse(404, "Not Found"), nil
	}
}

// routeToProtectedHandler routes requests to handlers that require authentication
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
