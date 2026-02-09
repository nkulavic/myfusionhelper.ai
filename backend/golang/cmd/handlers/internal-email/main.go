package main

import (
	"context"
	"log"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"

	authMiddleware "github.com/myfusionhelper/api/internal/middleware/auth"

	// Internal email service endpoints (no auth required - internal only)
	healthClient "github.com/myfusionhelper/api/cmd/handlers/internal-email/clients/health"
	historyClient "github.com/myfusionhelper/api/cmd/handlers/internal-email/clients/history"
	sendClient "github.com/myfusionhelper/api/cmd/handlers/internal-email/clients/send"
)

// Handle is the main entry point for the internal email service
func Handle(ctx context.Context, event events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	log.Printf("Internal Email Handler: path=%s method=%s", event.RequestContext.HTTP.Path, event.RequestContext.HTTP.Method)

	// Handle OPTIONS request for CORS
	if event.RequestContext.HTTP.Method == "OPTIONS" {
		return events.APIGatewayV2HTTPResponse{
			StatusCode: 200,
			Headers: map[string]string{
				"Access-Control-Allow-Origin":  "*",
				"Access-Control-Allow-Methods": "GET, POST, OPTIONS",
				"Access-Control-Allow-Headers": "Content-Type",
			},
			Body: "",
		}, nil
	}

	// Route to handler based on path
	switch event.RequestContext.HTTP.Path {
	case "/internal/emails/health":
		return healthClient.Handle(ctx, event)
	case "/internal/emails/send":
		return sendClient.Handle(ctx, event)
	case "/internal/emails/history":
		return historyClient.Handle(ctx, event)

	default:
		log.Printf("No handler found for path: %s", event.RequestContext.HTTP.Path)
		return authMiddleware.CreateErrorResponse(404, "Not Found"), nil
	}
}

func main() {
	lambda.Start(Handle)
}
