package portalsession

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/myfusionhelper/api/internal/types"
)

func TestHandleWithAuth_MethodValidation(t *testing.T) {
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

			authCtx := &types.AuthContext{
				UserID:    "user:123",
				AccountID: "account-123",
				Email:     "test@example.com",
			}

			response, err := HandleWithAuth(context.Background(), event, authCtx)
			if err != nil {
				t.Fatalf("Expected no error, got %v", err)
			}

			if tt.method != "POST" && response.StatusCode != 405 {
				t.Errorf("Expected status 405 for method %s, got %d", tt.method, response.StatusCode)
			}
		})
	}
}

func TestHandleWithAuth_MissingStripeKey(t *testing.T) {
	event := events.APIGatewayV2HTTPRequest{
		RequestContext: events.APIGatewayV2HTTPRequestContext{
			HTTP: events.APIGatewayV2HTTPRequestContextHTTPDescription{
				Method: "POST",
			},
		},
	}

	authCtx := &types.AuthContext{
		UserID:    "user:123",
		AccountID: "account-123",
		Email:     "test@example.com",
	}

	// Clear STRIPE_SECRET_KEY env var
	t.Setenv("STRIPE_SECRET_KEY", "")

	response, err := HandleWithAuth(context.Background(), event, authCtx)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if response.StatusCode != 503 {
		t.Errorf("Expected status 503 when Stripe key missing, got %d", response.StatusCode)
	}

	var respBody map[string]interface{}
	json.Unmarshal([]byte(response.Body), &respBody)

	if respBody["success"] != false {
		t.Error("Expected success=false in response")
	}

	errorMsg, ok := respBody["error"].(string)
	if !ok || errorMsg != "Billing service not configured" {
		t.Errorf("Expected 'Billing service not configured' error, got %v", respBody["error"])
	}
}

func TestHandleWithAuth_ResponseStructure(t *testing.T) {
	event := events.APIGatewayV2HTTPRequest{
		RequestContext: events.APIGatewayV2HTTPRequestContext{
			HTTP: events.APIGatewayV2HTTPRequestContextHTTPDescription{
				Method: "POST",
			},
		},
	}

	authCtx := &types.AuthContext{
		UserID:    "user:123",
		AccountID: "account-123",
		Email:     "test@example.com",
	}

	t.Setenv("STRIPE_SECRET_KEY", "sk_test_fake")
	t.Setenv("ACCOUNTS_TABLE", "test-accounts")

	response, err := HandleWithAuth(context.Background(), event, authCtx)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify response has correct CORS headers
	if response.Headers["Access-Control-Allow-Origin"] == "" {
		t.Error("Expected CORS headers in response")
	}

	// Verify response body is valid JSON
	var respBody map[string]interface{}
	if err := json.Unmarshal([]byte(response.Body), &respBody); err != nil {
		t.Errorf("Response body is not valid JSON: %v", err)
	}

	// Verify response has success field
	if _, ok := respBody["success"]; !ok {
		t.Error("Response missing 'success' field")
	}
}

// Integration test (requires AWS and Stripe access)
// Skipped by default, run with: go test -tags=integration

// func TestHandleWithAuth_CreatePortalSession_Integration(t *testing.T) {
//   // This test would:
//   // 1. Set up test DynamoDB table with account that has stripe_customer_id
//   // 2. Use Stripe test mode API keys
//   // 3. Create portal session
//   // 4. Verify response contains portal URL
//   // 5. Verify URL is valid Stripe portal URL
//   // 6. Clean up test data
// }

// func TestHandleWithAuth_NoStripeCustomer_Integration(t *testing.T) {
//   // This test would:
//   // 1. Set up test account without stripe_customer_id
//   // 2. Attempt to create portal session
//   // 3. Verify returns 400 error
//   // 4. Verify error message instructs user to subscribe first
// }
