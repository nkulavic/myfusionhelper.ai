package sync

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	ddbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go-v2/service/sqs"

	"github.com/myfusionhelper/api/internal/apiutil"
	authMiddleware "github.com/myfusionhelper/api/internal/middleware/auth"
	apitypes "github.com/myfusionhelper/api/internal/types"
)

var (
	connectionsTable  = os.Getenv("CONNECTIONS_TABLE")
	dataSyncQueueURL  = os.Getenv("DATA_SYNC_QUEUE_URL")
)

// SyncRequest is the expected POST body.
type SyncRequest struct {
	ConnectionID string `json:"connection_id"`
}

// SyncMessage is the message sent to the data sync SQS queue.
// Must include object_types so the data-sync worker knows what to sync.
type SyncMessage struct {
	AccountID    string   `json:"account_id"`
	ConnectionID string   `json:"connection_id"`
	ObjectTypes  []string `json:"object_types"`
	TriggerType  string   `json:"trigger_type"`
	RequestedAt  string   `json:"requested_at"`
}

// HandleWithAuth handles POST /data/sync.
func HandleWithAuth(ctx context.Context, event events.APIGatewayV2HTTPRequest, authCtx *apitypes.AuthContext) (events.APIGatewayV2HTTPResponse, error) {
	var req SyncRequest
	if err := json.Unmarshal([]byte(apiutil.GetBody(event)), &req); err != nil {
		return authMiddleware.CreateErrorResponse(400, "Invalid request body"), nil
	}

	if req.ConnectionID == "" {
		return authMiddleware.CreateErrorResponse(400, "connectionId is required"), nil
	}

	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(os.Getenv("COGNITO_REGION")))
	if err != nil {
		log.Printf("Failed to load AWS config: %v", err)
		return authMiddleware.CreateErrorResponse(500, "Internal server error"), nil
	}

	// Verify account owns connection
	ddb := dynamodb.NewFromConfig(cfg)
	connResult, err := ddb.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(connectionsTable),
		Key: map[string]ddbtypes.AttributeValue{
			"connection_id": &ddbtypes.AttributeValueMemberS{Value: req.ConnectionID},
		},
	})
	if err != nil || connResult.Item == nil {
		return authMiddleware.CreateErrorResponse(404, "Connection not found"), nil
	}

	var conn apitypes.PlatformConnection
	if err := attributevalue.UnmarshalMap(connResult.Item, &conn); err != nil {
		return authMiddleware.CreateErrorResponse(500, "Failed to parse connection"), nil
	}
	if conn.AccountID != authCtx.AccountID {
		return authMiddleware.CreateErrorResponse(403, "Access denied"), nil
	}

	// Build sync message
	msg := SyncMessage{
		AccountID:    authCtx.AccountID,
		ConnectionID: req.ConnectionID,
		ObjectTypes:  []string{"contacts", "tags", "custom_fields"},
		TriggerType:  "manual",
		RequestedAt:  time.Now().UTC().Format(time.RFC3339),
	}

	msgBody, err := json.Marshal(msg)
	if err != nil {
		log.Printf("Failed to marshal sync message: %v", err)
		return authMiddleware.CreateErrorResponse(500, "Failed to create sync message"), nil
	}

	// Send to SQS
	sqsClient := sqs.NewFromConfig(cfg)
	_, err = sqsClient.SendMessage(ctx, &sqs.SendMessageInput{
		QueueUrl:    aws.String(dataSyncQueueURL),
		MessageBody: aws.String(string(msgBody)),
	})
	if err != nil {
		log.Printf("Failed to send SQS message: %v", err)
		return authMiddleware.CreateErrorResponse(500, "Failed to trigger sync"), nil
	}

	return authMiddleware.CreateSuccessResponse(202, "Sync triggered", map[string]interface{}{
		"connection_id": req.ConnectionID,
		"status":        "queued",
	}), nil
}
