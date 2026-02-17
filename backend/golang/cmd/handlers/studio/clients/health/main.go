package health

import (
	"context"

	"github.com/aws/aws-lambda-go/events"

	authMiddleware "github.com/myfusionhelper/api/internal/middleware/auth"
)

// Handle returns a simple health check response.
func Handle(ctx context.Context, event events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	return authMiddleware.CreateSuccessResponse(200, "Studio service healthy", map[string]interface{}{
		"service": "studio",
		"status":  "ok",
	}), nil
}
