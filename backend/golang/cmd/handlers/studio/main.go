package main

import (
	"context"
	"log"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"

	authMiddleware "github.com/myfusionhelper/api/internal/middleware/auth"

	dashboardsClient "github.com/myfusionhelper/api/cmd/handlers/studio/clients/dashboards"
	healthClient "github.com/myfusionhelper/api/cmd/handlers/studio/clients/health"
	templatesClient "github.com/myfusionhelper/api/cmd/handlers/studio/clients/templates"
)

// Handle is the main entry point for the studio service.
func Handle(ctx context.Context, event events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	path := event.RequestContext.HTTP.Path
	method := event.RequestContext.HTTP.Method

	log.Printf("Studio Handler: path=%s method=%s", path, method)

	if method == "OPTIONS" {
		return events.APIGatewayV2HTTPResponse{
			StatusCode: 200,
			Headers: map[string]string{
				"Access-Control-Allow-Origin":  "*",
				"Access-Control-Allow-Methods": "GET, POST, PUT, DELETE, OPTIONS",
				"Access-Control-Allow-Headers": "Content-Type, Authorization",
			},
			Body: "",
		}, nil
	}

	switch {
	// Public endpoints
	case path == "/studio/health" && method == "GET":
		return healthClient.Handle(ctx, event)

	// Dashboard CRUD
	case path == "/studio/dashboards" && method == "GET":
		return routeToProtectedHandler(ctx, event, dashboardsClient.HandleWithAuth)
	case path == "/studio/dashboards" && method == "POST":
		return routeToProtectedHandler(ctx, event, dashboardsClient.HandleWithAuth)
	case strings.HasPrefix(path, "/studio/dashboards/") && method == "GET":
		return routeToProtectedHandler(ctx, event, dashboardsClient.HandleWithAuth)
	case strings.HasPrefix(path, "/studio/dashboards/") && method == "PUT":
		return routeToProtectedHandler(ctx, event, dashboardsClient.HandleWithAuth)
	case strings.HasPrefix(path, "/studio/dashboards/") && method == "DELETE":
		return routeToProtectedHandler(ctx, event, dashboardsClient.HandleWithAuth)

	// Templates
	case path == "/studio/templates" && method == "GET":
		return routeToProtectedHandler(ctx, event, templatesClient.HandleWithAuth)
	case strings.HasPrefix(path, "/studio/templates/") && strings.HasSuffix(path, "/apply") && method == "POST":
		return routeToProtectedHandler(ctx, event, templatesClient.HandleWithAuth)

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
