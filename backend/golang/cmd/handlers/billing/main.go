package main

import (
	"context"
	"log"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"

	authMiddleware "github.com/myfusionhelper/api/internal/middleware/auth"

	// Protected endpoints (require auth)
	getBillingClient "github.com/myfusionhelper/api/cmd/handlers/billing/clients/get-billing"
	invoicesClient "github.com/myfusionhelper/api/cmd/handlers/billing/clients/invoices"
	portalClient "github.com/myfusionhelper/api/cmd/handlers/billing/clients/portal-session"

	// Public endpoints (webhook)
	webhookClient "github.com/myfusionhelper/api/cmd/handlers/billing/clients/webhook"
)

// Handle is the main entry point for the billing service
func Handle(ctx context.Context, event events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	log.Printf("Billing Handler: path=%s method=%s", event.RequestContext.HTTP.Path, event.RequestContext.HTTP.Method)

	// Handle OPTIONS for CORS
	if event.RequestContext.HTTP.Method == "OPTIONS" {
		return events.APIGatewayV2HTTPResponse{
			StatusCode: 200,
			Headers: map[string]string{
				"Access-Control-Allow-Origin":  "*",
				"Access-Control-Allow-Methods": "GET, POST, OPTIONS",
				"Access-Control-Allow-Headers": "Content-Type, Authorization, Stripe-Signature",
			},
			Body: "",
		}, nil
	}

	switch event.RequestContext.HTTP.Path {
	// Protected endpoints
	case "/billing":
		return routeToProtectedHandler(ctx, event, getBillingClient.HandleWithAuth)
	case "/billing/portal-session":
		return routeToProtectedHandler(ctx, event, portalClient.HandleWithAuth)
	case "/billing/invoices":
		return routeToProtectedHandler(ctx, event, invoicesClient.HandleWithAuth)

	// Public endpoint (Stripe webhook -- verified by signature, not JWT)
	case "/billing/webhook":
		return webhookClient.Handle(ctx, event)

	default:
		log.Printf("No handler found for path: %s", event.RequestContext.HTTP.Path)
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
