package health

import (
	"context"
	"os"

	"github.com/aws/aws-lambda-go/events"
	authMiddleware "github.com/myfusionhelper/api/internal/middleware/auth"
)

func Handle(ctx context.Context, event events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	return authMiddleware.CreateSuccessResponse(200, "Emails service healthy", map[string]interface{}{
		"service": "mfh-emails",
		"version": os.Getenv("SERVICE_VERSION"),
		"stage":   os.Getenv("STAGE"),
	}), nil
}
