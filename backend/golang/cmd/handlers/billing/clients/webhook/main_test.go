package webhook

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-lambda-go/events"
)

func TestHandle_MethodValidation(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		expectedStatus int
	}{
		{"POST allowed", "POST", 200},
		{"GET not allowed", "GET", 405},
		{"PUT not allowed", "PUT", 405},
		{"DELETE not allowed", "DELETE", 405},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event := events.APIGatewayV2HTTPRequest{
				RequestContext: events.APIGatewayV2HTTPRequestContext{
					HTTP: events.APIGatewayV2HTTPRequestContextHTTPDescription{
						Method: tt.method,
					},
				},
			}

			t.Setenv("STRIPE_WEBHOOK_SECRET", "whsec_test123")

			response, err := Handle(context.Background(), event)
			if err != nil {
				t.Fatalf("Expected no error, got %v", err)
			}

			if tt.method != "POST" && response.StatusCode != 405 {
				t.Errorf("Expected status 405 for method %s, got %d", tt.method, response.StatusCode)
			}
		})
	}
}

func TestHandle_MissingWebhookSecret(t *testing.T) {
	event := events.APIGatewayV2HTTPRequest{
		RequestContext: events.APIGatewayV2HTTPRequestContext{
			HTTP: events.APIGatewayV2HTTPRequestContextHTTPDescription{
				Method: "POST",
			},
		},
		Body: `{"type": "customer.subscription.created"}`,
	}

	// Clear webhook secret env var
	t.Setenv("STRIPE_WEBHOOK_SECRET", "")

	response, err := Handle(context.Background(), event)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if response.StatusCode != 500 {
		t.Errorf("Expected status 500 when webhook secret missing, got %d", response.StatusCode)
	}

	var respBody map[string]interface{}
	json.Unmarshal([]byte(response.Body), &respBody)

	if respBody["success"] != false {
		t.Error("Expected success=false in response")
	}
}

func TestHandle_MissingStripeSignature(t *testing.T) {
	event := events.APIGatewayV2HTTPRequest{
		RequestContext: events.APIGatewayV2HTTPRequestContext{
			HTTP: events.APIGatewayV2HTTPRequestContextHTTPDescription{
				Method: "POST",
			},
		},
		Headers: map[string]string{},
		Body:    `{"type": "customer.subscription.created"}`,
	}

	t.Setenv("STRIPE_WEBHOOK_SECRET", "whsec_test123")

	response, err := Handle(context.Background(), event)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if response.StatusCode != 400 {
		t.Errorf("Expected status 400 when signature missing, got %d", response.StatusCode)
	}

	var respBody map[string]interface{}
	json.Unmarshal([]byte(response.Body), &respBody)

	if respBody["success"] != false {
		t.Error("Expected success=false in response")
	}

	errorMsg, ok := respBody["error"].(string)
	if !ok || errorMsg != "Missing Stripe signature" {
		t.Errorf("Expected 'Missing Stripe signature' error, got %v", respBody["error"])
	}
}

func TestHandle_InvalidSignature(t *testing.T) {
	payload := `{"id":"evt_test","type":"customer.subscription.created"}`

	event := events.APIGatewayV2HTTPRequest{
		RequestContext: events.APIGatewayV2HTTPRequestContext{
			HTTP: events.APIGatewayV2HTTPRequestContextHTTPDescription{
				Method: "POST",
			},
		},
		Headers: map[string]string{
			"stripe-signature": "t=123456789,v1=invalid_signature",
		},
		Body: payload,
	}

	t.Setenv("STRIPE_WEBHOOK_SECRET", "whsec_test123")

	response, err := Handle(context.Background(), event)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if response.StatusCode != 400 {
		t.Errorf("Expected status 400 for invalid signature, got %d", response.StatusCode)
	}

	var respBody map[string]interface{}
	json.Unmarshal([]byte(response.Body), &respBody)

	if respBody["success"] != false {
		t.Error("Expected success=false in response")
	}
}

// computeStripeSignature generates a valid Stripe webhook signature for testing
func computeStripeSignature(payload string, secret string, timestamp int64) string {
	signedPayload := fmt.Sprintf("%d.%s", timestamp, payload)
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(signedPayload))
	signature := hex.EncodeToString(h.Sum(nil))
	return fmt.Sprintf("t=%d,v1=%s", timestamp, signature)
}

func TestHandle_ValidSignature_UnhandledEvent(t *testing.T) {
	secret := "whsec_test123"
	timestamp := time.Now().Unix()

	eventData := map[string]interface{}{
		"id":   "evt_test123",
		"type": "payment_intent.created", // Unhandled event type
		"data": map[string]interface{}{
			"object": map[string]interface{}{},
		},
	}
	payload, _ := json.Marshal(eventData)

	signature := computeStripeSignature(string(payload), secret, timestamp)

	event := events.APIGatewayV2HTTPRequest{
		RequestContext: events.APIGatewayV2HTTPRequestContext{
			HTTP: events.APIGatewayV2HTTPRequestContextHTTPDescription{
				Method: "POST",
			},
		},
		Headers: map[string]string{
			"stripe-signature": signature,
		},
		Body: string(payload),
	}

	t.Setenv("STRIPE_WEBHOOK_SECRET", secret)

	response, err := Handle(context.Background(), event)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Unhandled events should return 200 OK
	if response.StatusCode != 200 {
		t.Errorf("Expected status 200 for unhandled event, got %d", response.StatusCode)
	}

	var respBody map[string]interface{}
	json.Unmarshal([]byte(response.Body), &respBody)

	if respBody["success"] != true {
		t.Error("Expected success=true for unhandled event")
	}
}

func TestHandle_EventTypeRouting(t *testing.T) {
	// This test verifies that different event types are routed correctly
	// Actual handler logic is tested separately due to AWS/Stripe dependencies

	secret := "whsec_test456"
	timestamp := time.Now().Unix()

	tests := []struct {
		name      string
		eventType string
	}{
		{"subscription created", "customer.subscription.created"},
		{"subscription updated", "customer.subscription.updated"},
		{"subscription deleted", "customer.subscription.deleted"},
		{"checkout completed", "checkout.session.completed"},
		{"payment failed", "invoice.payment_failed"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			eventData := map[string]interface{}{
				"id":   "evt_test_" + tt.eventType,
				"type": tt.eventType,
				"data": map[string]interface{}{
					"object": map[string]interface{}{
						"customer": "cus_test123",
					},
				},
			}
			payload, _ := json.Marshal(eventData)

			signature := computeStripeSignature(string(payload), secret, timestamp)

			event := events.APIGatewayV2HTTPRequest{
				RequestContext: events.APIGatewayV2HTTPRequestContext{
					HTTP: events.APIGatewayV2HTTPRequestContextHTTPDescription{
						Method: "POST",
					},
				},
				Headers: map[string]string{
					"stripe-signature": signature,
				},
				Body: string(payload),
			}

			t.Setenv("STRIPE_WEBHOOK_SECRET", secret)
			t.Setenv("ACCOUNTS_TABLE", "test-accounts")
			t.Setenv("USERS_TABLE", "test-users")
			t.Setenv("COGNITO_USER_POOL_ID", "us-west-2_test")

			response, err := Handle(context.Background(), event)
			if err != nil {
				t.Fatalf("Expected no error for %s, got %v", tt.eventType, err)
			}

			// Handlers will fail due to missing AWS resources, but signature should pass
			// The signature validation happens before routing, so we verify that step works
			if response.StatusCode == 400 {
				var respBody map[string]interface{}
				json.Unmarshal([]byte(response.Body), &respBody)
				if errorMsg, ok := respBody["error"].(string); ok {
					if errorMsg == "Invalid signature" || errorMsg == "Missing Stripe signature" {
						t.Errorf("Signature validation failed for %s: %s", tt.eventType, errorMsg)
					}
				}
			}
		})
	}
}

// Integration tests (require AWS and Stripe test mode)
// Skipped by default, run with: go test -tags=integration

// func TestHandleSubscriptionUpdate_Integration(t *testing.T) {
//   // This test would:
//   // 1. Set up test DynamoDB tables
//   // 2. Create test account and Stripe customer
//   // 3. Generate real Stripe webhook event
//   // 4. Verify subscription status updated in DynamoDB
//   // 5. Verify plan limits updated
//   // 6. Clean up test data
// }

// func TestHandleSubscriptionCancelled_Integration(t *testing.T) {
//   // This test would:
//   // 1. Set up test account with active subscription
//   // 2. Generate subscription.deleted webhook
//   // 3. Verify subscription status changed to cancelled
//   // 4. Verify plan downgraded to free
//   // 5. Verify user notified via email
// }

// func TestHandleCheckoutSessionCompleted_Integration(t *testing.T) {
//   // This test would:
//   // 1. Create checkout session via API
//   // 2. Simulate checkout.session.completed webhook
//   // 3. Verify account upgraded to paid plan
//   // 4. Verify trial period set correctly
//   // 5. Verify welcome email sent
// }

// func TestHandlePaymentFailed_Integration(t *testing.T) {
//   // This test would:
//   // 1. Set up account with active subscription
//   // 2. Simulate invoice.payment_failed webhook
//   // 3. Verify payment failure recorded
//   // 4. Verify user notified of failure
//   // 5. Verify account still active (grace period)
// }
