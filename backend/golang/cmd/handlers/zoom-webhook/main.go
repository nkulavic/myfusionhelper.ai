package main

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/myfusionhelper/api/internal/apiutil"
)

var (
	zoomWebhookSecret = os.Getenv("ZOOM_WEBHOOK_SECRET")
	helperQueueURL    = os.Getenv("HELPER_QUEUE_URL")
	connectionsTable  = os.Getenv("CONNECTIONS_TABLE")
)

// ZoomWebhookEvent represents the incoming Zoom webhook payload
type ZoomWebhookEvent struct {
	Event          string                 `json:"event"`
	EventTimestamp int64                  `json:"event_ts"`
	Payload        map[string]interface{} `json:"payload"`
}

// ZoomWebinarObject represents webinar data in the payload
type ZoomWebinarObject struct {
	ID              string `json:"id"`
	UUID            string `json:"uuid"`
	HostID          string `json:"host_id"`
	Topic           string `json:"topic"`
	Type            int    `json:"type"`
	StartTime       string `json:"start_time"`
	Timezone        string `json:"timezone"`
	Duration        int    `json:"duration"`
	TotalSize       int    `json:"total_size"`
	RecordingCount  int    `json:"recording_count"`
	ShareURL        string `json:"share_url,omitempty"`
	RecordingFiles  []interface{} `json:"recording_files,omitempty"`
}

func main() {
	lambda.Start(handleRequest)
}

func handleRequest(ctx context.Context, event events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	log.Printf("Received Zoom webhook: %s %s", event.RequestContext.HTTP.Method, event.RequestContext.HTTP.Path)

	// Verify webhook signature
	body := apiutil.GetBody(event)
	if zoomWebhookSecret != "" {
		signature := event.Headers["x-zm-signature"]
		timestamp := event.Headers["x-zm-request-timestamp"]

		if !verifySignature(body, signature, timestamp, zoomWebhookSecret) {
			log.Printf("Invalid webhook signature")
			return createResponse(401, map[string]interface{}{
				"success": false,
				"error":   "Invalid signature",
			}), nil
		}
	}

	// Parse webhook event
	var webhookEvent ZoomWebhookEvent
	if err := json.Unmarshal([]byte(body), &webhookEvent); err != nil {
		log.Printf("Failed to parse webhook: %v", err)
		return createResponse(400, map[string]interface{}{
			"success": false,
			"error":   "Invalid JSON",
		}), nil
	}

	log.Printf("Zoom event: %s at %d", webhookEvent.Event, webhookEvent.EventTimestamp)

	// Handle different event types
	switch webhookEvent.Event {
	case "webinar.ended":
		return handleWebinarEnded(ctx, webhookEvent)
	case "endpoint.url_validation":
		// Zoom sends this to validate the webhook URL
		return handleURLValidation(webhookEvent)
	default:
		log.Printf("Unhandled event type: %s", webhookEvent.Event)
		return createResponse(200, map[string]interface{}{
			"success": true,
			"message": "Event received but not processed",
		}), nil
	}
}

// handleURLValidation responds to Zoom's URL validation challenge
func handleURLValidation(webhookEvent ZoomWebhookEvent) (events.APIGatewayV2HTTPResponse, error) {
	// Extract plainToken and encryptedToken from payload
	plainToken := ""
	if payload, ok := webhookEvent.Payload["plainToken"].(string); ok {
		plainToken = payload
	}

	// Create hash using webhook secret
	hash := hmac.New(sha256.New, []byte(zoomWebhookSecret))
	hash.Write([]byte(plainToken))
	encryptedToken := hex.EncodeToString(hash.Sum(nil))

	log.Printf("URL validation: plainToken=%s, encryptedToken=%s", plainToken, encryptedToken)

	return createResponse(200, map[string]interface{}{
		"plainToken":     plainToken,
		"encryptedToken": encryptedToken,
	}), nil
}

// handleWebinarEnded processes the webinar.ended event
func handleWebinarEnded(ctx context.Context, webhookEvent ZoomWebhookEvent) (events.APIGatewayV2HTTPResponse, error) {
	payload := webhookEvent.Payload

	// Extract webinar object
	webinarObj, ok := payload["object"].(map[string]interface{})
	if !ok {
		log.Printf("No webinar object in payload")
		return createResponse(200, map[string]interface{}{
			"success": true,
			"message": "No webinar object",
		}), nil
	}

	webinarID := getStringFromMap(webinarObj, "id", "")
	webinarUUID := getStringFromMap(webinarObj, "uuid", "")
	webinarTopic := getStringFromMap(webinarObj, "topic", "")

	log.Printf("Webinar ended: ID=%s, UUID=%s, Topic=%s", webinarID, webinarUUID, webinarTopic)

	// Initialize AWS clients
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(os.Getenv("COGNITO_REGION")))
	if err != nil {
		log.Printf("Failed to load AWS config: %v", err)
		return createResponse(500, map[string]interface{}{
			"success": false,
			"error":   "AWS config error",
		}), err
	}

	sqsClient := sqs.NewFromConfig(cfg)

	// Queue jobs to process participants and absentees
	// In a real implementation, we would:
	// 1. Fetch webinar participants from Zoom API
	// 2. Fetch webinar registrants from Zoom API
	// 3. Match participants to CRM contacts by email
	// 4. Queue helper execution jobs for each contact

	// For now, we'll create a placeholder job structure
	job := map[string]interface{}{
		"event_type":   "webinar.ended",
		"webinar_id":   webinarID,
		"webinar_uuid": webinarUUID,
		"webinar_topic": webinarTopic,
		"timestamp":    time.Now().UTC().Format(time.RFC3339),
	}

	jobJSON, err := json.Marshal(job)
	if err != nil {
		log.Printf("Failed to marshal job: %v", err)
		return createResponse(500, map[string]interface{}{
			"success": false,
			"error":   "Job marshal error",
		}), err
	}

	// Send to SQS queue for processing
	_, err = sqsClient.SendMessage(ctx, &sqs.SendMessageInput{
		QueueUrl:    aws.String(helperQueueURL),
		MessageBody: aws.String(string(jobJSON)),
	})
	if err != nil {
		log.Printf("Failed to send message to SQS: %v", err)
		return createResponse(500, map[string]interface{}{
			"success": false,
			"error":   "SQS send error",
		}), err
	}

	log.Printf("Queued webinar processing job for webinar %s", webinarID)

	return createResponse(200, map[string]interface{}{
		"success": true,
		"message": "Webinar processing queued",
		"webinar_id": webinarID,
	}), nil
}

// verifySignature verifies the Zoom webhook signature
func verifySignature(body, signature, timestamp, secret string) bool {
	if signature == "" || timestamp == "" {
		return false
	}

	// Construct message: v0:timestamp:body
	message := fmt.Sprintf("v0:%s:%s", timestamp, body)

	// Create HMAC SHA256 hash
	hash := hmac.New(sha256.New, []byte(secret))
	hash.Write([]byte(message))
	expectedSignature := "v0=" + hex.EncodeToString(hash.Sum(nil))

	return hmac.Equal([]byte(expectedSignature), []byte(signature))
}

func createResponse(statusCode int, body map[string]interface{}) events.APIGatewayV2HTTPResponse {
	bodyJSON, _ := json.Marshal(body)
	return events.APIGatewayV2HTTPResponse{
		StatusCode: statusCode,
		Headers: map[string]string{
			"Content-Type":                "application/json",
			"Access-Control-Allow-Origin": "*",
		},
		Body: string(bodyJSON),
	}
}

func getStringFromMap(m map[string]interface{}, key, defaultValue string) string {
	if val, ok := m[key]; ok {
		switch v := val.(type) {
		case string:
			return v
		case float64:
			return fmt.Sprintf("%.0f", v)
		case int:
			return fmt.Sprintf("%d", v)
		default:
			return fmt.Sprintf("%v", v)
		}
	}
	return defaultValue
}
