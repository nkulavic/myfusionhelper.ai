package main

import (
	"context"
	"log"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"

	authMiddleware "github.com/myfusionhelper/api/internal/middleware/auth"

	crudClient "github.com/myfusionhelper/api/cmd/handlers/helpers/clients/crud"
	executeClient "github.com/myfusionhelper/api/cmd/handlers/helpers/clients/execute"
	executionsClient "github.com/myfusionhelper/api/cmd/handlers/helpers/clients/executions"
	healthClient "github.com/myfusionhelper/api/cmd/handlers/helpers/clients/health"
	typesClient "github.com/myfusionhelper/api/cmd/handlers/helpers/clients/types"

	// Register all helpers via init() so the registry is populated
	_ "github.com/myfusionhelper/api/internal/connectors"
	_ "github.com/myfusionhelper/api/internal/helpers/analytics"
	_ "github.com/myfusionhelper/api/internal/helpers/automation"
	_ "github.com/myfusionhelper/api/internal/helpers/contact"
	_ "github.com/myfusionhelper/api/internal/helpers/data"
	_ "github.com/myfusionhelper/api/internal/helpers/integration"
	_ "github.com/myfusionhelper/api/internal/helpers/notification"
	_ "github.com/myfusionhelper/api/internal/helpers/tagging"
)

// Handle is the main entry point for the consolidated helpers service
func Handle(ctx context.Context, event events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	path := event.RequestContext.HTTP.Path
	method := event.RequestContext.HTTP.Method

	log.Printf("Helpers Handler: path=%s method=%s", path, method)

	if method == "OPTIONS" {
		return events.APIGatewayV2HTTPResponse{
			StatusCode: 200,
			Headers: map[string]string{
				"Access-Control-Allow-Origin":  "*",
				"Access-Control-Allow-Methods": "GET, POST, PUT, DELETE, OPTIONS",
				"Access-Control-Allow-Headers": "Content-Type, Authorization, X-API-Key",
			},
			Body: "",
		}, nil
	}

	switch {
	// API-key-authenticated execute endpoints (Lambda authorizer handles auth)
	case strings.HasPrefix(path, "/helper/"):
		return executeClient.Handle(ctx, event)

	// Public endpoints
	case path == "/helpers/health" && method == "GET":
		return healthClient.Handle(ctx, event)

	// Helper types catalog (must be before generic /helpers/{id} routes)
	case path == "/helpers/types" && method == "GET":
		return routeToProtectedHandler(ctx, event, typesClient.HandleWithAuth)
	case strings.HasPrefix(path, "/helpers/types/") && method == "GET":
		return routeToProtectedHandler(ctx, event, typesClient.HandleWithAuth)

	// Executions endpoints
	case path == "/executions" && method == "GET":
		return routeToProtectedHandler(ctx, event, executionsClient.HandleWithAuth)
	case strings.HasPrefix(path, "/executions/") && method == "GET":
		return routeToProtectedHandler(ctx, event, executionsClient.HandleWithAuth)

	// Protected endpoints
	case path == "/helpers" && method == "GET":
		return routeToProtectedHandler(ctx, event, crudClient.HandleWithAuth)
	case path == "/helpers" && method == "POST":
		return routeToProtectedHandler(ctx, event, crudClient.HandleWithAuth)
	case strings.HasSuffix(path, "/execute") && method == "POST":
		return routeToProtectedHandler(ctx, event, crudClient.HandleWithAuth)
	case strings.HasPrefix(path, "/helpers/") && method == "GET":
		return routeToProtectedHandler(ctx, event, crudClient.HandleWithAuth)
	case strings.HasPrefix(path, "/helpers/") && method == "PUT":
		return routeToProtectedHandler(ctx, event, crudClient.HandleWithAuth)
	case strings.HasPrefix(path, "/helpers/") && method == "DELETE":
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
