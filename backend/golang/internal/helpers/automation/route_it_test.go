package automation

import (
	"context"
	"testing"

	"github.com/myfusionhelper/api/internal/helpers"
)

func TestRouteIt_Metadata(t *testing.T) {
	helper := &RouteIt{}

	if name := helper.GetName(); name != "Route It" {
		t.Errorf("expected name 'Route It', got '%s'", name)
	}

	if helperType := helper.GetType(); helperType != "route_it" {
		t.Errorf("expected type 'route_it', got '%s'", helperType)
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

func TestRouteIt_ValidateConfig(t *testing.T) {
	helper := &RouteIt{}

	tests := []struct {
		name        string
		config      map[string]interface{}
		expectError bool
	}{
		{
			name: "valid config with routes",
			config: map[string]interface{}{
				"routes": []interface{}{
					map[string]interface{}{
						"label":       "Option A",
						"redirectUrl": "https://example.com/a",
					},
					map[string]interface{}{
						"label":       "Option B",
						"redirectUrl": "https://example.com/b",
					},
				},
			},
			expectError: false,
		},
		{
			name: "valid config with fallback URL",
			config: map[string]interface{}{
				"routes": []interface{}{
					map[string]interface{}{
						"label":       "Primary",
						"redirectUrl": "https://example.com/primary",
					},
				},
				"fallback_url": "https://example.com/fallback",
			},
			expectError: false,
		},
		{
			name: "valid config with save_to_field and apply_tag",
			config: map[string]interface{}{
				"routes": []interface{}{
					map[string]interface{}{
						"label":       "Route",
						"redirectUrl": "https://example.com/route",
					},
				},
				"save_to_field": "redirect_url",
				"apply_tag":     "routed",
			},
			expectError: false,
		},
		{
			name: "missing routes",
			config: map[string]interface{}{
				"fallback_url": "https://example.com/default",
			},
			expectError: true,
		},
		{
			name: "empty routes array",
			config: map[string]interface{}{
				"routes": []interface{}{},
			},
			expectError: true,
		},
		{
			name: "route missing redirectUrl",
			config: map[string]interface{}{
				"routes": []interface{}{
					map[string]interface{}{
						"label": "Invalid",
					},
				},
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

func TestRouteIt_Execute_FirstRouteSelected(t *testing.T) {
	helper := &RouteIt{}

	config := map[string]interface{}{
		"routes": []interface{}{
			map[string]interface{}{
				"label":       "Primary",
				"redirectUrl": "https://example.com/primary",
			},
			map[string]interface{}{
				"label":       "Secondary",
				"redirectUrl": "https://example.com/secondary",
			},
		},
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

	redirectURL, ok := modData["redirect_url"].(string)
	if !ok || redirectURL != "https://example.com/primary" {
		t.Errorf("expected redirect_url 'https://example.com/primary', got '%v'", redirectURL)
	}

	routingReason, ok := modData["routing_reason"].(string)
	if !ok || routingReason == "" {
		t.Errorf("expected routing_reason to be set, got '%v'", routingReason)
	}
}

func TestRouteIt_Execute_WithFallback(t *testing.T) {
	helper := &RouteIt{}

	config := map[string]interface{}{
		"routes": []interface{}{
			map[string]interface{}{
				"label":       "Primary",
				"redirectUrl": "https://example.com/primary",
			},
		},
		"fallback_url": "https://example.com/fallback",
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

	// First route should be selected, not fallback
	modData := output.ModifiedData
	redirectURL := modData["redirect_url"].(string)
	if redirectURL != "https://example.com/primary" {
		t.Errorf("expected primary URL to be selected, got '%s'", redirectURL)
	}

	routingReason := modData["routing_reason"].(string)
	if routingReason == "" {
		t.Error("expected routing_reason to be set")
	}
}

func TestRouteIt_Execute_EmptyRoutesUsesFallback(t *testing.T) {
	helper := &RouteIt{}

	// Note: This would fail validation, but testing Execute behavior directly
	config := map[string]interface{}{
		"routes":       []interface{}{},
		"fallback_url": "https://example.com/fallback",
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
	if redirectURL != "https://example.com/fallback" {
		t.Errorf("expected fallback URL, got '%s'", redirectURL)
	}

	routingReason := modData["routing_reason"].(string)
	if routingReason != "fallback" {
		t.Errorf("expected routing_reason 'fallback', got '%s'", routingReason)
	}
}

func TestRouteIt_Execute_WithSaveToField(t *testing.T) {
	helper := &RouteIt{}
	mockConn := &mockConnectorForRouting{}

	config := map[string]interface{}{
		"routes": []interface{}{
			map[string]interface{}{
				"label":       "Primary",
				"redirectUrl": "https://example.com/primary",
			},
		},
		"save_to_field": "last_redirect_url",
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
	if call.fieldKey != "last_redirect_url" {
		t.Errorf("expected field key 'last_redirect_url', got '%s'", call.fieldKey)
	}
	if call.value != "https://example.com/primary" {
		t.Errorf("expected value 'https://example.com/primary', got '%v'", call.value)
	}
}

func TestRouteIt_Execute_WithApplyTag(t *testing.T) {
	helper := &RouteIt{}
	mockConn := &mockConnectorForRouting{}

	config := map[string]interface{}{
		"routes": []interface{}{
			map[string]interface{}{
				"label":       "Primary",
				"redirectUrl": "https://example.com/primary",
			},
		},
		"apply_tag": "routed_tag",
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

	if mockConn.applyTagCalls[0] != "routed_tag" {
		t.Errorf("expected tag 'routed_tag', got '%s'", mockConn.applyTagCalls[0])
	}
}

func TestRouteIt_Execute_MultipleRoutesSequential(t *testing.T) {
	helper := &RouteIt{}

	config := map[string]interface{}{
		"routes": []interface{}{
			map[string]interface{}{
				"label":       "Route 1",
				"redirectUrl": "https://example.com/route1",
			},
			map[string]interface{}{
				"label":       "Route 2",
				"redirectUrl": "https://example.com/route2",
			},
			map[string]interface{}{
				"label":       "Route 3",
				"redirectUrl": "https://example.com/route3",
			},
		},
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

	// Should always select the first route
	if redirectURL != "https://example.com/route1" {
		t.Errorf("expected first route to be selected, got '%s'", redirectURL)
	}
}

func TestRouteIt_Execute_SavesRoutedAtTimestamp(t *testing.T) {
	helper := &RouteIt{}

	config := map[string]interface{}{
		"routes": []interface{}{
			map[string]interface{}{
				"label":       "Primary",
				"redirectUrl": "https://example.com/primary",
			},
		},
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

	// Check that routed_at timestamp exists
	routedAt, ok := modData["routed_at"].(string)
	if !ok || routedAt == "" {
		t.Error("expected routed_at timestamp to be present")
	}
}
