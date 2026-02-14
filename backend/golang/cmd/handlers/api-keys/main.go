package main

import (
	"context"
	"log"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"

	authMiddleware "github.com/myfusionhelper/api/internal/middleware/auth"

	crudClient "github.com/myfusionhelper/api/cmd/handlers/api-keys/clients/crud"
	healthClient "github.com/myfusionhelper/api/cmd/handlers/api-keys/clients/health"
)

// Handle is the main entry point for the consolidated API keys service
func Handle(ctx context.Context, event events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	path := event.RequestContext.HTTP.Path
	method := event.RequestContext.HTTP.Method

	log.Printf("API Keys Handler: path=%s method=%s", path, method)

	if method == "OPTIONS" {
		return events.APIGatewayV2HTTPResponse{
			StatusCode: 200,
			Headers: map[string]string{
				"Access-Control-Allow-Origin":  "*",
				"Access-Control-Allow-Methods": "GET, POST, DELETE, OPTIONS",
				"Access-Control-Allow-Headers": "Content-Type, Authorization",
			},
			Body: "",
		}, nil
	}

	switch {
	case path == "/api-keys/health" && method == "GET":
		return healthClient.Handle(ctx, event)
	case path == "/api-keys" && method == "GET":
		return routeToProtectedHandler(ctx, event, crudClient.HandleWithAuth)
	case path == "/api-keys" && method == "POST":
		return routeToProtectedHandler(ctx, event, crudClient.HandleWithAuth)
	case strings.HasPrefix(path, "/api-keys/") && method == "DELETE":
		return routeToProtectedHandler(ctx, event, crudClient.HandleWithAuth)
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
