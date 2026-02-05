package main

import (
	"context"
	"encoding/json"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	ddbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go-v2/service/sqs"

	apitypes "github.com/myfusionhelper/api/internal/types"
)

var (
	connectionsTable = os.Getenv("CONNECTIONS_TABLE")
	dataSyncQueueURL = os.Getenv("DATA_SYNC_QUEUE_URL")
)

// SyncMessage matches the message format expected by the data-sync worker.
type SyncMessage struct {
	AccountID    string   `json:"account_id"`
	ConnectionID string   `json:"connection_id"`
	ObjectTypes  []string `json:"object_types"`
}

func main() {
	lambda.Start(handleScheduleEvent)
}

func handleScheduleEvent(ctx context.Context) error {
	log.Println("Data sync scheduler triggered")

	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(os.Getenv("COGNITO_REGION")))
	if err != nil {
		log.Printf("Failed to load AWS config: %v", err)
		return err
	}

	db := dynamodb.NewFromConfig(cfg)
	sqsClient := sqs.NewFromConfig(cfg)

	// Scan for all active connections
	activeConnections, err := scanActiveConnections(ctx, db)
	if err != nil {
		log.Printf("Failed to scan active connections: %v", err)
		return err
	}

	log.Printf("Found %d active connections to sync", len(activeConnections))

	sentCount := 0
	for _, conn := range activeConnections {
		msg := SyncMessage{
			AccountID:    conn.AccountID,
			ConnectionID: conn.ConnectionID,
			ObjectTypes:  []string{"contacts", "tags", "custom_fields"},
		}

		msgBody, err := json.Marshal(msg)
		if err != nil {
			log.Printf("Failed to marshal sync message for connection %s: %v", conn.ConnectionID, err)
			continue
		}

		_, err = sqsClient.SendMessage(ctx, &sqs.SendMessageInput{
			QueueUrl:    aws.String(dataSyncQueueURL),
			MessageBody: aws.String(string(msgBody)),
		})
		if err != nil {
			log.Printf("Failed to send sync message for connection %s: %v", conn.ConnectionID, err)
			continue
		}

		sentCount++
	}

	log.Printf("Sent %d sync messages out of %d active connections", sentCount, len(activeConnections))
	return nil
}

// scanActiveConnections scans the connections table for all connections with status="active".
func scanActiveConnections(ctx context.Context, db *dynamodb.Client) ([]apitypes.PlatformConnection, error) {
	var connections []apitypes.PlatformConnection
	var lastEvaluatedKey map[string]ddbtypes.AttributeValue

	for {
		input := &dynamodb.ScanInput{
			TableName:        aws.String(connectionsTable),
			FilterExpression: aws.String("#s = :active"),
			ExpressionAttributeNames: map[string]string{
				"#s": "status",
			},
			ExpressionAttributeValues: map[string]ddbtypes.AttributeValue{
				":active": &ddbtypes.AttributeValueMemberS{Value: "active"},
			},
		}
		if lastEvaluatedKey != nil {
			input.ExclusiveStartKey = lastEvaluatedKey
		}

		result, err := db.Scan(ctx, input)
		if err != nil {
			return nil, err
		}

		var page []apitypes.PlatformConnection
		if err := attributevalue.UnmarshalListOfMaps(result.Items, &page); err != nil {
			return nil, err
		}
		connections = append(connections, page...)

		if result.LastEvaluatedKey == nil {
			break
		}
		lastEvaluatedKey = result.LastEvaluatedKey
	}

	return connections, nil
}
