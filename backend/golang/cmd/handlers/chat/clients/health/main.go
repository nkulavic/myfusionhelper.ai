package health

import (
	"context"
	"os"

	"github.com/aws/aws-lambda-go/events"
	authMiddleware "github.com/myfusionhelper/api/internal/middleware/auth"
)

// HandlePublic handles the health check endpoint (no auth required)
func HandlePublic(ctx context.Context, event events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	stage := os.Getenv("STAGE")
	if stage == "" {
		stage = "unknown"
	}

	data := map[string]interface{}{
		"status":  "healthy",
		"service": "chat",
		"stage":   stage,
	}

	return authMiddleware.CreateSuccessResponse(200, "Chat service is healthy", data), nil
}
