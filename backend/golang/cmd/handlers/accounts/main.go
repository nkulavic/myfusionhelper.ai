package main

import (
	"context"
	"log"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"

	authMiddleware "github.com/myfusionhelper/api/internal/middleware/auth"

	// Protected endpoints (require auth)
	crudClient "github.com/myfusionhelper/api/cmd/handlers/accounts/clients/crud"

	// Public endpoints (no auth required)
	healthClient "github.com/myfusionhelper/api/cmd/handlers/accounts/clients/health"
)

// Handle is the main entry point for the consolidated accounts service
func Handle(ctx context.Context, event events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	path := event.RequestContext.HTTP.Path
	method := event.RequestContext.HTTP.Method

	log.Printf("Accounts Handler: path=%s method=%s", path, method)

	// Handle OPTIONS request for CORS
	if method == "OPTIONS" {
		return events.APIGatewayV2HTTPResponse{
			StatusCode: 200,
			Headers: map[string]string{
				"Access-Control-Allow-Origin":  "*",
				"Access-Control-Allow-Methods": "GET, POST, PUT, DELETE, OPTIONS",
				"Access-Control-Allow-Headers": "Content-Type, Authorization, X-Account-Context",
			},
			Body: "",
		}, nil
	}

	// Route to handler based on path and method
	switch {
	// Public endpoints
	case path == "/accounts/health" && method == "GET":
		return healthClient.Handle(ctx, event)

	// Protected endpoints
	case path == "/accounts/switch" && method == "POST":
		return routeToProtectedHandler(ctx, event, crudClient.HandleWithAuth)
	case path == "/accounts" && method == "GET":
		return routeToProtectedHandler(ctx, event, crudClient.HandleWithAuth)
	case strings.HasPrefix(path, "/accounts/") && method == "GET":
		return routeToProtectedHandler(ctx, event, crudClient.HandleWithAuth)
	case strings.HasPrefix(path, "/accounts/") && method == "PUT":
		return routeToProtectedHandler(ctx, event, crudClient.HandleWithAuth)

	default:
		log.Printf("No handler found for path: %s method: %s", path, method)
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
