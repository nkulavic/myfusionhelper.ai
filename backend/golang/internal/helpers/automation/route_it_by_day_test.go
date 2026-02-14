package automation

import (
	"context"
	"testing"
	"time"

	"github.com/myfusionhelper/api/internal/helpers"
)

func TestRouteItByDay_Metadata(t *testing.T) {
	helper := &RouteItByDay{}

	if name := helper.GetName(); name != "Route It By Day" {
		t.Errorf("expected name 'Route It By Day', got '%s'", name)
	}

	if helperType := helper.GetType(); helperType != "route_it_by_day" {
		t.Errorf("expected type 'route_it_by_day', got '%s'", helperType)
	}

	if category := helper.GetCategory(); category != "automation" {
		t.Errorf("expected category 'automation', got '%s'", category)
	}

	if helper.RequiresCRM() {
		t.Error("expected RequiresCRM to be false")
	}

	schema := helper.GetConfigSchema()
	if schema == nil {
		t.Fatal("expected config schema to be non-nil")
	}
}

func TestRouteItByDay_ValidateConfig(t *testing.T) {
	helper := &RouteItByDay{}

	tests := []struct {
		name        string
		config      map[string]interface{}
		expectError bool
	}{
		{
			name: "valid config with day routes",
			config: map[string]interface{}{
				"day_routes": map[string]interface{}{
					"Monday":    "https://example.com/monday",
					"Wednesday": "https://example.com/wednesday",
				},
			},
			expectError: false,
		},
		{
			name: "valid config with timezone",
			config: map[string]interface{}{
				"day_routes": map[string]interface{}{
					"Friday": "https://example.com/friday",
				},
				"timezone": "America/New_York",
			},
			expectError: false,
		},
		{
			name: "valid config with fallback URL",
			config: map[string]interface{}{
				"day_routes": map[string]interface{}{
					"Sunday": "https://example.com/sunday",
				},
				"fallback_url": "https://example.com/default",
			},
			expectError: false,
		},
		{
			name: "missing day_routes",
			config: map[string]interface{}{
				"timezone": "UTC",
			},
			expectError: true,
		},
		{
			name: "empty day_routes",
			config: map[string]interface{}{
				"day_routes": map[string]interface{}{},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := helper.ValidateConfig(tt.config)
			if tt.expectError && err == nil {
				t.Error("expected validation error but got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("expected no validation error but got: %v", err)
			}
		})
	}
}

func TestRouteItByDay_Execute_MatchingDay(t *testing.T) {
	helper := &RouteItByDay{}

	// Get current day of week
	now := time.Now().UTC()
	currentDay := now.Weekday().String()

	config := map[string]interface{}{
		"day_routes": map[string]interface{}{
			currentDay: "https://example.com/today",
		},
		"timezone": "UTC",
	}

	input := helpers.HelperInput{
		Config:     config,
		ContactID:  "contact123",
		Connector:  nil, // No connector needed for basic routing
		AccountID:  "account456",
		UserID: "user789",
	}

	output, err := helper.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if !output.Success {
		t.Error("expected success to be true")
	}

	modData := output.ModifiedData

	redirectURL, ok := modData["redirect_url"].(string)
	if !ok || redirectURL != "https://example.com/today" {
		t.Errorf("expected redirect_url 'https://example.com/today', got '%v'", redirectURL)
	}

	routingReason, ok := modData["routing_reason"].(string)
	if !ok || routingReason == "" {
		t.Errorf("expected routing_reason to be set, got '%v'", routingReason)
	}
}

func TestRouteItByDay_Execute_FallbackURL(t *testing.T) {
	helper := &RouteItByDay{}

	config := map[string]interface{}{
		"day_routes": map[string]interface{}{
			// Use a day that's definitely not today (if today is Monday, use Tuesday, etc.)
			"NonExistentDay": "https://example.com/never",
		},
		"fallback_url": "https://example.com/default",
		"timezone":     "UTC",
	}

	input := helpers.HelperInput{
		Config:     config,
		ContactID:  "contact123",
		Connector:  nil,
		AccountID:  "account456",
		UserID: "user789",
	}

	output, err := helper.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	modData := output.ModifiedData
	redirectURL := modData["redirect_url"].(string)
	if redirectURL != "https://example.com/default" {
		t.Errorf("expected fallback URL, got '%s'", redirectURL)
	}

	routingReason := modData["routing_reason"].(string)
	if routingReason != "fallback" {
		t.Errorf("expected routing_reason 'fallback', got '%s'", routingReason)
	}
}

func TestRouteItByDay_Execute_WithSaveToField(t *testing.T) {
	helper := &RouteItByDay{}
	mockConn := &mockConnectorForRouting{}

	now := time.Now().UTC()
	currentDay := now.Weekday().String()

	config := map[string]interface{}{
		"day_routes": map[string]interface{}{
			currentDay: "https://example.com/today",
		},
		"timezone":      "UTC",
		"save_to_field": "last_redirect",
	}

	input := helpers.HelperInput{
		Config:     config,
		ContactID:  "contact123",
		Connector:  mockConn,
		AccountID:  "account456",
		UserID: "user789",
	}

	output, err := helper.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if !output.Success {
		t.Error("expected success to be true")
	}

	// Verify SetContactFieldValue was called
	if len(mockConn.setFieldCalls) != 1 {
		t.Fatalf("expected 1 SetContactFieldValue call, got %d", len(mockConn.setFieldCalls))
	}

	call := mockConn.setFieldCalls[0]
	if call.fieldKey != "last_redirect" {
		t.Errorf("expected field key 'last_redirect', got '%s'", call.fieldKey)
	}
	if call.value != "https://example.com/today" {
		t.Errorf("expected value 'https://example.com/today', got '%v'", call.value)
	}
}

func TestRouteItByDay_Execute_WithApplyTag(t *testing.T) {
	helper := &RouteItByDay{}
	mockConn := &mockConnectorForRouting{}

	now := time.Now().UTC()
	currentDay := now.Weekday().String()

	config := map[string]interface{}{
		"day_routes": map[string]interface{}{
			currentDay: "https://example.com/today",
		},
		"timezone":  "UTC",
		"apply_tag": "routed_by_day",
	}

	input := helpers.HelperInput{
		Config:     config,
		ContactID:  "contact123",
		Connector:  mockConn,
		AccountID:  "account456",
		UserID: "user789",
	}

	output, err := helper.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if !output.Success {
		t.Error("expected success to be true")
	}

	// Verify ApplyTag was called
	if len(mockConn.applyTagCalls) != 1 {
		t.Fatalf("expected 1 ApplyTag call, got %d", len(mockConn.applyTagCalls))
	}

	if mockConn.applyTagCalls[0] != "routed_by_day" {
		t.Errorf("expected tag 'routed_by_day', got '%s'", mockConn.applyTagCalls[0])
	}
}

func TestRouteItByDay_Execute_DifferentTimezones(t *testing.T) {
	helper := &RouteItByDay{}

	// Test with a specific timezone to verify it's being used
	config := map[string]interface{}{
		"day_routes": map[string]interface{}{
			"Monday":    "https://example.com/monday",
			"Tuesday":   "https://example.com/tuesday",
			"Wednesday": "https://example.com/wednesday",
			"Thursday":  "https://example.com/thursday",
			"Friday":    "https://example.com/friday",
			"Saturday":  "https://example.com/saturday",
			"Sunday":    "https://example.com/sunday",
		},
		"timezone":     "America/New_York",
		"fallback_url": "https://example.com/default",
	}

	input := helpers.HelperInput{
		Config:     config,
		ContactID:  "contact123",
		Connector:  nil,
		AccountID:  "account456",
		UserID: "user789",
	}

	output, err := helper.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if !output.Success {
		t.Error("expected success to be true")
	}

	modData := output.ModifiedData
	redirectURL := modData["redirect_url"].(string)

	// Should get one of the day URLs, not the fallback
	if redirectURL == "https://example.com/default" {
		t.Error("expected a day-specific URL, not the fallback")
	}

	// Verify we got a valid day URL
	validURLs := []string{
		"https://example.com/monday",
		"https://example.com/tuesday",
		"https://example.com/wednesday",
		"https://example.com/thursday",
		"https://example.com/friday",
		"https://example.com/saturday",
		"https://example.com/sunday",
	}

	found := false
	for _, validURL := range validURLs {
		if redirectURL == validURL {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("expected one of the valid day URLs, got '%s'", redirectURL)
	}
}
