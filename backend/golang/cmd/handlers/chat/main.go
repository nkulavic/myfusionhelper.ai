package main

import (
	"context"
	"log"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"

	authMiddleware "github.com/myfusionhelper/api/internal/middleware/auth"
	"github.com/myfusionhelper/api/cmd/handlers/chat/clients/conversations"
	"github.com/myfusionhelper/api/cmd/handlers/chat/clients/health"
	"github.com/myfusionhelper/api/cmd/handlers/chat/clients/messages"
)

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
