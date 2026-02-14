package connectiontags

import (
	"context"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"

	"github.com/myfusionhelper/api/internal/connectors"
	"github.com/myfusionhelper/api/internal/connectors/loader"
	authMiddleware "github.com/myfusionhelper/api/internal/middleware/auth"
	apitypes "github.com/myfusionhelper/api/internal/types"
)

// HandleWithAuth returns all available tags for a connection.
func HandleWithAuth(ctx context.Context, event events.APIGatewayV2HTTPRequest, authCtx *apitypes.AuthContext) (events.APIGatewayV2HTTPResponse, error) {
	connectionID := event.PathParameters["connection_id"]
	if connectionID == "" {
		return authMiddleware.CreateErrorResponse(400, "Connection ID is required"), nil
	}

	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(os.Getenv("COGNITO_REGION")))
	if err != nil {
		log.Printf("Failed to load AWS config: %v", err)
		return authMiddleware.CreateErrorResponse(500, "Internal server error"), nil
	}
	db := dynamodb.NewFromConfig(cfg)

	connector, err := loader.LoadConnector(ctx, db, connectionID, authCtx.AccountID)
	if err != nil {
		log.Printf("Failed to load connector: %v", err)
		if connErr, ok := err.(*connectors.ConnectorError); ok {
			return authMiddleware.CreateErrorResponse(connErr.StatusCode, connErr.Message), nil
		}
		return authMiddleware.CreateErrorResponse(500, "Failed to load connection"), nil
	}

	tags, err := connector.GetTags(ctx)
	if err != nil {
		log.Printf("Failed to get tags: %v", err)
		if connErr, ok := err.(*connectors.ConnectorError); ok {
			return authMiddleware.CreateErrorResponse(connErr.StatusCode, connErr.Message), nil
		}
		return authMiddleware.CreateErrorResponse(500, "Failed to retrieve tags"), nil
	}

	return authMiddleware.CreateSuccessResponse(200, "Tags retrieved successfully", tags), nil
}
