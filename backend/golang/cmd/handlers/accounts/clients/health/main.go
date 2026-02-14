package health

import (
	"context"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/events"
	authMiddleware "github.com/myfusionhelper/api/internal/middleware/auth"
)

// Handle is the health check handler (public, no auth required)
func Handle(ctx context.Context, event events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	return authMiddleware.CreateSuccessResponse(200, "Service is healthy", map[string]interface{}{
		"service":   "mfh-accounts",
		"version":   os.Getenv("SERVICE_VERSION"),
		"stage":     os.Getenv("STAGE"),
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}), nil
}
