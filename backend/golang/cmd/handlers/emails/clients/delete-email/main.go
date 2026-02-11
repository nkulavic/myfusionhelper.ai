package deleteEmail

import (
	"context"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	authMiddleware "github.com/myfusionhelper/api/internal/middleware/auth"
)

func HandleWithAuth(ctx context.Context, event events.APIGatewayV2HTTPRequest, authCtx *authMiddleware.AuthContext) (events.APIGatewayV2HTTPResponse, error) {
	emailID := event.PathParameters["id"]
	if emailID == "" {
		return authMiddleware.CreateErrorResponse(400, "Email ID is required"), nil
	}

	// Initialize AWS config
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(os.Getenv("COGNITO_REGION")))
	if err != nil {
		log.Printf("Failed to load AWS config: %v", err)
		return authMiddleware.CreateErrorResponse(500, "Failed to initialize AWS"), nil
	}

	dbClient := dynamodb.NewFromConfig(cfg)

	// Delete email log from DynamoDB
	_, err = dbClient.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: aws.String(os.Getenv("EMAIL_LOGS_TABLE")),
		Key: map[string]types.AttributeValue{
			"email_id": &types.AttributeValueMemberS{Value: emailID},
		},
		// Security: ensure the email belongs to the user's account
		ConditionExpression: aws.String("account_id = :account_id"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":account_id": &types.AttributeValueMemberS{Value: authCtx.AccountID},
		},
	})

	if err != nil {
		log.Printf("Failed to delete email: %v", err)
		return authMiddleware.CreateErrorResponse(500, "Failed to delete email"), nil
	}

	return authMiddleware.CreateSuccessResponse(200, "Email deleted successfully", nil), nil
}
