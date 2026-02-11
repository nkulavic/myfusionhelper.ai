package main

import (
	"context"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	authMiddleware "github.com/myfusionhelper/api/internal/middleware/auth"

	"github.com/myfusionhelper/api/cmd/handlers/emails/clients/create-template"
	"github.com/myfusionhelper/api/cmd/handlers/emails/clients/delete-email"
	"github.com/myfusionhelper/api/cmd/handlers/emails/clients/delete-template"
	"github.com/myfusionhelper/api/cmd/handlers/emails/clients/get-template"
	"github.com/myfusionhelper/api/cmd/handlers/emails/clients/health"
	"github.com/myfusionhelper/api/cmd/handlers/emails/clients/list-emails"
	"github.com/myfusionhelper/api/cmd/handlers/emails/clients/list-templates"
	"github.com/myfusionhelper/api/cmd/handlers/emails/clients/send-email"
	"github.com/myfusionhelper/api/cmd/handlers/emails/clients/update-template"
)

func main() {
	lambda.Start(Handle)
}

func Handle(ctx context.Context, event events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	path := event.RequestContext.HTTP.Path
	method := event.RequestContext.HTTP.Method

	// Health check (public)
	if path == "/emails/health" && method == "GET" {
		return health.Handle(ctx, event)
	}

	// Email list/send/delete (protected)
	if path == "/emails" && method == "GET" {
		return routeToProtectedHandler(ctx, event, listEmails.HandleWithAuth)
	}
	if path == "/emails" && method == "POST" {
		return routeToProtectedHandler(ctx, event, sendEmail.HandleWithAuth)
	}

	// Delete email - extract ID from path
	if method == "DELETE" && len(event.PathParameters) > 0 {
		if _, ok := event.PathParameters["id"]; ok && path != "/emails/templates/"+event.PathParameters["id"] {
			return routeToProtectedHandler(ctx, event, deleteEmail.HandleWithAuth)
		}
	}

	// Template endpoints (protected)
	if path == "/emails/templates" && method == "GET" {
		return routeToProtectedHandler(ctx, event, listTemplates.HandleWithAuth)
	}
	if path == "/emails/templates" && method == "POST" {
		return routeToProtectedHandler(ctx, event, createTemplate.HandleWithAuth)
	}

	// Template by ID - GET/PUT/DELETE
	if len(event.PathParameters) > 0 {
		if _, ok := event.PathParameters["id"]; ok {
			if method == "GET" {
				return routeToProtectedHandler(ctx, event, getTemplate.HandleWithAuth)
			}
			if method == "PUT" {
				return routeToProtectedHandler(ctx, event, updateTemplate.HandleWithAuth)
			}
			if method == "DELETE" {
				return routeToProtectedHandler(ctx, event, deleteTemplate.HandleWithAuth)
			}
		}
	}

	return authMiddleware.CreateErrorResponse(404, "Not found"), nil
}

// routeToProtectedHandler wraps a handler with auth middleware
func routeToProtectedHandler(
	ctx context.Context,
	event events.APIGatewayV2HTTPRequest,
	handler func(context.Context, events.APIGatewayV2HTTPRequest, *authMiddleware.AuthContext) (events.APIGatewayV2HTTPResponse, error),
) (events.APIGatewayV2HTTPResponse, error) {
	return authMiddleware.WithAuth(ctx, event, handler)
}
