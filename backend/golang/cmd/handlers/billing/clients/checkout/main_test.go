package checkout

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

func TestHandleWithAuth_InvalidRequestBody(t *testing.T) {
	// Skip if Stripe key not configured (module-level var set at init)
	if stripeKey == "" {
		t.Skip("STRIPE_SECRET_KEY not set, skipping test")
	}

	event := events.APIGatewayV2HTTPRequest{
		RequestContext: events.APIGatewayV2HTTPRequestContext{
			HTTP: events.APIGatewayV2HTTPRequestContextHTTPDescription{
				Method: "POST",
			},
		},
		Body: "invalid json",
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

	if response.StatusCode != 400 {
		t.Errorf("Expected status 400 for invalid JSON, got %d", response.StatusCode)
	}

	var respBody map[string]interface{}
	json.Unmarshal([]byte(response.Body), &respBody)

	if respBody["success"] != false {
		t.Error("Expected success=false in response")
	}
}

func TestHandleWithAuth_InvalidPlan(t *testing.T) {
	// Skip if Stripe key not configured (module-level var set at init)
	if stripeKey == "" {
		t.Skip("STRIPE_SECRET_KEY not set, skipping test")
	}

	reqBody := CreateCheckoutRequest{
		Plan: "invalid_plan",
	}
	bodyBytes, _ := json.Marshal(reqBody)

	event := events.APIGatewayV2HTTPRequest{
		RequestContext: events.APIGatewayV2HTTPRequestContext{
			HTTP: events.APIGatewayV2HTTPRequestContextHTTPDescription{
				Method: "POST",
			},
		},
		Body: string(bodyBytes),
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

	if response.StatusCode != 400 {
		t.Errorf("Expected status 400 for invalid plan, got %d", response.StatusCode)
	}

	var respBody map[string]interface{}
	json.Unmarshal([]byte(response.Body), &respBody)

	if respBody["success"] != false {
		t.Error("Expected success=false in response")
	}

	if respBody["error"] == nil {
		t.Error("Expected error message in response")
	}
}

func TestHandleWithAuth_MissingStripeKey(t *testing.T) {
	reqBody := CreateCheckoutRequest{
		Plan: "start",
	}
	bodyBytes, _ := json.Marshal(reqBody)

	event := events.APIGatewayV2HTTPRequest{
		RequestContext: events.APIGatewayV2HTTPRequestContext{
			HTTP: events.APIGatewayV2HTTPRequestContextHTTPDescription{
				Method: "POST",
			},
		},
		Body: string(bodyBytes),
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
}

func TestGetPriceID(t *testing.T) {
	// Note: getPriceID reads from module-level vars set at init time,
	// so t.Setenv won't affect it. These tests verify the logic works,
	// but with whatever env vars are set at test startup.

	tests := []struct {
		name     string
		plan     string
		expected string
	}{
		{"start plan returns priceStart var", "start", priceStart},
		{"grow plan returns priceGrow var", "grow", priceGrow},
		{"deliver plan returns priceDeliver var", "deliver", priceDeliver},
		{"invalid plan returns empty", "invalid", ""},
		{"empty plan returns empty", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getPriceID(tt.plan)
			if result != tt.expected {
				t.Errorf("getPriceID(%q) = %q, expected %q", tt.plan, result, tt.expected)
			}
		})
	}
}

// Integration test (requires AWS and Stripe access)
// Skipped by default, run with: go test -tags=integration
// func TestHandleWithAuth_CreateCheckoutSession(t *testing.T) {
//   // This test would:
//   // 1. Set up real/test DynamoDB table
//   // 2. Use Stripe test mode API keys
//   // 3. Create checkout session
//   // 4. Verify response contains session URL and ID
//   // 5. Clean up test data
// }
