package middleware

import (
	"encoding/json"
	"testing"

	"github.com/aws/aws-lambda-go/events"
)

// TestExtractSubFromJWT_AuthorizerContext tests JWT extraction from API Gateway authorizer context
func TestExtractSubFromJWT_AuthorizerContext(t *testing.T) {
	event := events.APIGatewayV2HTTPRequest{
		RequestContext: events.APIGatewayV2HTTPRequestContext{
			Authorizer: &events.APIGatewayV2HTTPRequestContextAuthorizerDescription{
				JWT: &events.APIGatewayV2HTTPRequestContextAuthorizerJWTDescription{
					Claims: map[string]string{
						"sub": "user-abc-123",
					},
				},
			},
		},
	}

	sub, err := extractSubFromJWT(event)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if sub != "user-abc-123" {
		t.Errorf("Expected sub 'user-abc-123', got '%s'", sub)
	}
}

// Note: Testing Bearer token parsing is complex because the middleware's key function returns nil
// which causes the JWT library to fail validation even with WithoutClaimsValidation().
// In production, API Gateway validates JWTs and passes claims via authorizer context (tested above).
// The Bearer token fallback exists for development/testing but is hard to unit test without
// refactoring the middleware to accept a key function parameter.

// TestExtractSubFromJWT_MissingAuthHeader tests error when Authorization header is missing
func TestExtractSubFromJWT_MissingAuthHeader(t *testing.T) {
	event := events.APIGatewayV2HTTPRequest{
		Headers: map[string]string{},
	}

	_, err := extractSubFromJWT(event)
	if err == nil {
		t.Fatal("Expected error for missing Authorization header, got nil")
	}

	if err.Error() != "missing Authorization header" {
		t.Errorf("Expected 'missing Authorization header', got '%s'", err.Error())
	}
}

// TestExtractSubFromJWT_InvalidBearerFormat tests error when Bearer token format is invalid
func TestExtractSubFromJWT_InvalidBearerFormat(t *testing.T) {
	event := events.APIGatewayV2HTTPRequest{
		Headers: map[string]string{
			"Authorization": "InvalidFormat token123",
		},
	}

	_, err := extractSubFromJWT(event)
	if err == nil {
		t.Fatal("Expected error for invalid Bearer format, got nil")
	}

	if err.Error() != "invalid Bearer token format" {
		t.Errorf("Expected 'invalid Bearer token format', got '%s'", err.Error())
	}
}


// Note: WithAuth integration tests require DynamoDB mocking which is complex with concrete types.
// Those tests are better suited for integration testing against actual DynamoDB or LocalStack.
// The core JWT extraction and response formatting functions are tested above.

// TestCreateSuccessResponse tests success response formatting
func TestCreateSuccessResponse(t *testing.T) {
	data := map[string]string{
		"key": "value",
	}

	resp := CreateSuccessResponse(200, "Operation successful", data)

	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	if resp.Headers["Content-Type"] != "application/json" {
		t.Errorf("Expected Content-Type 'application/json', got '%s'", resp.Headers["Content-Type"])
	}

	if resp.Headers["Access-Control-Allow-Origin"] != "*" {
		t.Errorf("Expected CORS header '*', got '%s'", resp.Headers["Access-Control-Allow-Origin"])
	}

	var body map[string]interface{}
	if err := json.Unmarshal([]byte(resp.Body), &body); err != nil {
		t.Fatalf("Failed to parse response body: %v", err)
	}

	if body["success"] != true {
		t.Error("Expected success=true")
	}

	if body["message"] != "Operation successful" {
		t.Errorf("Expected message 'Operation successful', got '%v'", body["message"])
	}

	if body["data"] == nil {
		t.Error("Expected data to be present")
	}
}

// TestCreateSuccessResponse_NilData tests success response with nil data
func TestCreateSuccessResponse_NilData(t *testing.T) {
	resp := CreateSuccessResponse(204, "No content", nil)

	if resp.StatusCode != 204 {
		t.Errorf("Expected status 204, got %d", resp.StatusCode)
	}

	var body map[string]interface{}
	if err := json.Unmarshal([]byte(resp.Body), &body); err != nil {
		t.Fatalf("Failed to parse response body: %v", err)
	}

	if body["success"] != true {
		t.Error("Expected success=true")
	}

	if body["message"] != "No content" {
		t.Errorf("Expected message 'No content', got '%v'", body["message"])
	}

	if _, exists := body["data"]; exists {
		t.Error("Expected data key to not exist when nil")
	}
}

// TestCreateErrorResponse tests error response formatting
func TestCreateErrorResponse(t *testing.T) {
	resp := CreateErrorResponse(400, "Bad request")

	if resp.StatusCode != 400 {
		t.Errorf("Expected status 400, got %d", resp.StatusCode)
	}

	if resp.Headers["Content-Type"] != "application/json" {
		t.Errorf("Expected Content-Type 'application/json', got '%s'", resp.Headers["Content-Type"])
	}

	if resp.Headers["Access-Control-Allow-Origin"] != "*" {
		t.Errorf("Expected CORS header '*', got '%s'", resp.Headers["Access-Control-Allow-Origin"])
	}

	var body map[string]interface{}
	if err := json.Unmarshal([]byte(resp.Body), &body); err != nil {
		t.Fatalf("Failed to parse response body: %v", err)
	}

	if body["success"] != false {
		t.Error("Expected success=false")
	}

	if body["error"] != "Bad request" {
		t.Errorf("Expected error 'Bad request', got '%v'", body["error"])
	}
}
