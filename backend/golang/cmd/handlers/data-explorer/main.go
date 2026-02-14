package main

import (
	"context"
	"log"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"

	authMiddleware "github.com/myfusionhelper/api/internal/middleware/auth"

	catalogClient "github.com/myfusionhelper/api/cmd/handlers/data-explorer/clients/catalog"
	exportClient "github.com/myfusionhelper/api/cmd/handlers/data-explorer/clients/export"
	healthClient "github.com/myfusionhelper/api/cmd/handlers/data-explorer/clients/health"
	queryClient "github.com/myfusionhelper/api/cmd/handlers/data-explorer/clients/query"
	recordClient "github.com/myfusionhelper/api/cmd/handlers/data-explorer/clients/record"
	syncClient "github.com/myfusionhelper/api/cmd/handlers/data-explorer/clients/sync"

	// Register all connectors via init()
	_ "github.com/myfusionhelper/api/internal/connectors"
)

// Handle is the main entry point for the data explorer service
func Handle(ctx context.Context, event events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	path := event.RequestContext.HTTP.Path
	method := event.RequestContext.HTTP.Method

	log.Printf("Data Explorer Handler: path=%s method=%s", path, method)

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
	// Public endpoints
	case path == "/data/health" && method == "GET":
		return healthClient.Handle(ctx, event)

	// Catalog
	case path == "/data/catalog" && method == "GET":
		return routeToProtectedHandler(ctx, event, catalogClient.HandleWithAuth)

	// Query
	case path == "/data/query" && method == "POST":
		return routeToProtectedHandler(ctx, event, queryClient.HandleWithAuth)

	// Single record: /data/record/{connectionId}/{objectType}/{recordId}
	case strings.HasPrefix(path, "/data/record/") && method == "GET":
		return routeToProtectedHandler(ctx, event, recordClient.HandleWithAuth)

	// Export
	case path == "/data/export" && method == "POST":
		return routeToProtectedHandler(ctx, event, exportClient.HandleWithAuth)

	// Sync
	case path == "/data/sync" && method == "POST":
		return routeToProtectedHandler(ctx, event, syncClient.HandleWithAuth)

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
