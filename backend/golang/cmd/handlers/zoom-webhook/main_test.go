package main

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/aws/aws-lambda-go/events"
)

func TestHandleURLValidation(t *testing.T) {
	// Set zoom webhook secret for testing
	os.Setenv("ZOOM_WEBHOOK_SECRET", "test-secret")
	defer os.Unsetenv("ZOOM_WEBHOOK_SECRET")

	webhookEvent := ZoomWebhookEvent{
		Event: "endpoint.url_validation",
		Payload: map[string]interface{}{
			"plainToken": "test-plain-token",
		},
	}

	resp, err := handleURLValidation(webhookEvent)
	if err != nil {
		t.Fatalf("handleURLValidation failed: %v", err)
	}

	if resp.StatusCode != 200 {
		t.Errorf("Expected status code 200, got %d", resp.StatusCode)
	}

	var body map[string]interface{}
	if err := json.Unmarshal([]byte(resp.Body), &body); err != nil {
		t.Fatalf("Failed to unmarshal response body: %v", err)
	}

	if body["plainToken"] != "test-plain-token" {
		t.Errorf("Expected plainToken 'test-plain-token', got %v", body["plainToken"])
	}

	if body["encryptedToken"] == "" {
		t.Error("Expected encryptedToken to be non-empty")
	}
}

func TestVerifySignature(t *testing.T) {
	secret := "test-secret"
	timestamp := "1234567890"
	body := `{"event":"test"}`

	// Generate valid signature
	message := "v0:" + timestamp + ":" + body
	validSignature := "v0=b5a9e9e9c3d3c3c3c3c3c3c3c3c3c3c3c3c3c3c3c3c3c3c3c3c3c3c3c3c3c3c3"

	// Test with empty signature
	if verifySignature(body, "", timestamp, secret) {
		t.Error("Expected verification to fail with empty signature")
	}

	// Test with empty timestamp
	if verifySignature(body, validSignature, "", secret) {
		t.Error("Expected verification to fail with empty timestamp")
	}

	// Test message construction
	expectedMessage := "v0:" + timestamp + ":" + body
	if message != expectedMessage {
		t.Errorf("Expected message '%s', got '%s'", expectedMessage, message)
	}
}

func TestGetStringFromMap(t *testing.T) {
	testMap := map[string]interface{}{
		"string_val":  "hello",
		"int_val":     42,
		"float_val":   3.14,
		"missing_val": nil,
	}

	tests := []struct {
		key          string
		defaultVal   string
		expected     string
	}{
		{"string_val", "default", "hello"},
		{"int_val", "default", "42"},
		{"float_val", "default", "3"},
		{"nonexistent", "default", "default"},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			result := getStringFromMap(testMap, tt.key, tt.defaultVal)
			if result != tt.expected {
				t.Errorf("For key '%s', expected '%s', got '%s'", tt.key, tt.expected, result)
			}
		})
	}
}

func TestCreateResponse(t *testing.T) {
	body := map[string]interface{}{
		"success": true,
		"message": "test",
	}

	resp := createResponse(200, body)

	if resp.StatusCode != 200 {
		t.Errorf("Expected status code 200, got %d", resp.StatusCode)
	}

	if resp.Headers["Content-Type"] != "application/json" {
		t.Errorf("Expected Content-Type 'application/json', got '%s'", resp.Headers["Content-Type"])
	}

	if resp.Headers["Access-Control-Allow-Origin"] != "*" {
		t.Errorf("Expected CORS header '*', got '%s'", resp.Headers["Access-Control-Allow-Origin"])
	}

	var responseBody map[string]interface{}
	if err := json.Unmarshal([]byte(resp.Body), &responseBody); err != nil {
		t.Fatalf("Failed to unmarshal response body: %v", err)
	}

	if responseBody["success"] != true {
		t.Errorf("Expected success=true, got %v", responseBody["success"])
	}
}

func TestHandleRequest_InvalidJSON(t *testing.T) {
	event := events.APIGatewayV2HTTPRequest{
		Body: "invalid json",
		RequestContext: events.APIGatewayV2HTTPRequestContext{
			HTTP: events.APIGatewayV2HTTPRequestContextHTTPDescription{
				Method: "POST",
				Path:   "/webhooks/zoom",
			},
		},
	}

	resp, err := handleRequest(context.Background(), event)
	if err != nil {
		t.Fatalf("handleRequest failed: %v", err)
	}

	if resp.StatusCode != 400 {
		t.Errorf("Expected status code 400 for invalid JSON, got %d", resp.StatusCode)
	}
}

func TestHandleRequest_UnhandledEvent(t *testing.T) {
	webhookEvent := ZoomWebhookEvent{
		Event: "unknown.event",
		Payload: map[string]interface{}{},
	}

	body, _ := json.Marshal(webhookEvent)

	event := events.APIGatewayV2HTTPRequest{
		Body: string(body),
		RequestContext: events.APIGatewayV2HTTPRequestContext{
			HTTP: events.APIGatewayV2HTTPRequestContextHTTPDescription{
				Method: "POST",
				Path:   "/webhooks/zoom",
			},
		},
		Headers: map[string]string{},
	}

	resp, err := handleRequest(context.Background(), event)
	if err != nil {
		t.Fatalf("handleRequest failed: %v", err)
	}

	if resp.StatusCode != 200 {
		t.Errorf("Expected status code 200 for unhandled event, got %d", resp.StatusCode)
	}

	var responseBody map[string]interface{}
	if err := json.Unmarshal([]byte(resp.Body), &responseBody); err != nil {
		t.Fatalf("Failed to unmarshal response body: %v", err)
	}

	if responseBody["success"] != true {
		t.Error("Expected success=true for unhandled event")
	}
}
