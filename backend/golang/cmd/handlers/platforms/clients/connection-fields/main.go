package connectionfields

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

// Standard contact fields common across all CRM platforms
var standardFields = []connectors.CustomField{
	{Key: "first_name", Label: "First Name", FieldType: "text", GroupName: "Contact Info"},
	{Key: "last_name", Label: "Last Name", FieldType: "text", GroupName: "Contact Info"},
	{Key: "email", Label: "Email", FieldType: "email", GroupName: "Contact Info"},
	{Key: "phone", Label: "Phone", FieldType: "phone", GroupName: "Contact Info"},
	{Key: "company", Label: "Company", FieldType: "text", GroupName: "Contact Info"},
	{Key: "job_title", Label: "Job Title", FieldType: "text", GroupName: "Contact Info"},
}

// HandleWithAuth returns all available fields (standard + custom) for a connection.
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

	customFields, err := connector.GetCustomFields(ctx)
	if err != nil {
		log.Printf("Failed to get custom fields: %v", err)
		if connErr, ok := err.(*connectors.ConnectorError); ok {
			return authMiddleware.CreateErrorResponse(connErr.StatusCode, connErr.Message), nil
		}
		return authMiddleware.CreateErrorResponse(500, "Failed to retrieve fields"), nil
	}

	return authMiddleware.CreateSuccessResponse(200, "Fields retrieved successfully", map[string]interface{}{
		"standard_fields": standardFields,
		"custom_fields":   customFields,
	}), nil
}
