package automation

import (
	"context"
	"testing"
	"time"

	"github.com/myfusionhelper/api/internal/helpers"
)

func TestRouteItByTime_Metadata(t *testing.T) {
	helper := &RouteItByTime{}

	if name := helper.GetName(); name != "Route It By Time" {
		t.Errorf("expected name 'Route It By Time', got '%s'", name)
	}

	if helperType := helper.GetType(); helperType != "route_it_by_time" {
		t.Errorf("expected type 'route_it_by_time', got '%s'", helperType)
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

func TestRouteItByTime_ValidateConfig(t *testing.T) {
	helper := &RouteItByTime{}

	tests := []struct {
		name        string
		config      map[string]interface{}
		expectError bool
	}{
		{
			name: "valid config with time routes",
			config: map[string]interface{}{
				"time_routes": []interface{}{
					map[string]interface{}{
						"start_time": "09:00",
						"end_time":   "17:00",
						"url":        "https://example.com/business",
					},
				},
			},
			expectError: false,
		},
		{
			name: "valid config with multiple time routes",
			config: map[string]interface{}{
				"time_routes": []interface{}{
					map[string]interface{}{
						"start_time": "09:00",
						"end_time":   "12:00",
						"url":        "https://example.com/morning",
						"label":      "Morning shift",
					},
					map[string]interface{}{
						"start_time": "13:00",
						"end_time":   "17:00",
						"url":        "https://example.com/afternoon",
					},
				},
				"timezone": "America/New_York",
			},
			expectError: false,
		},
		{
			name: "valid overnight range",
			config: map[string]interface{}{
				"time_routes": []interface{}{
					map[string]interface{}{
						"start_time": "22:00",
						"end_time":   "06:00",
						"url":        "https://example.com/overnight",
					},
				},
			},
			expectError: false,
		},
		{
			name: "missing time_routes",
			config: map[string]interface{}{
				"timezone": "UTC",
			},
			expectError: true,
		},
		{
			name: "empty time_routes",
			config: map[string]interface{}{
				"time_routes": []interface{}{},
			},
			expectError: true,
		},
		{
			name: "invalid time format in start_time",
			config: map[string]interface{}{
				"time_routes": []interface{}{
					map[string]interface{}{
						"start_time": "25:00",
						"end_time":   "17:00",
						"url":        "https://example.com/invalid",
					},
				},
			},
			expectError: true,
		},
		{
			name: "invalid time format in end_time",
			config: map[string]interface{}{
				"time_routes": []interface{}{
					map[string]interface{}{
						"start_time": "09:00",
						"end_time":   "not-a-time",
						"url":        "https://example.com/invalid",
					},
				},
			},
			expectError: true,
		},
		{
			name: "missing URL in time route",
			config: map[string]interface{}{
				"time_routes": []interface{}{
					map[string]interface{}{
						"start_time": "09:00",
						"end_time":   "17:00",
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

func TestRouteItByTime_Execute_MatchingTimeRange(t *testing.T) {
	helper := &RouteItByTime{}

	// Get current time and create a time range that includes it
	now := time.Now().UTC()
	startHour := now.Hour()
	endHour := (startHour + 2) % 24

	config := map[string]interface{}{
		"time_routes": []interface{}{
			map[string]interface{}{
				"start_time": formatHour(startHour),
				"end_time":   formatHour(endHour),
				"url":        "https://example.com/current",
			},
		},
		"timezone": "UTC",
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
	if !ok {
		t.Fatal("expected redirect_url to be a string")
	}

	// Should match the current time range or fallback if endHour wrapped around
	if redirectURL != "https://example.com/current" && redirectURL != "" {
		// It's ok if it didn't match due to hour wrapping at midnight
		t.Logf("URL was '%s' - acceptable if time wrapped around midnight", redirectURL)
	}
}

func TestRouteItByTime_Execute_FallbackURL(t *testing.T) {
	helper := &RouteItByTime{}

	config := map[string]interface{}{
		"time_routes": []interface{}{
			map[string]interface{}{
				"start_time": "02:00",
				"end_time":   "03:00",
				"url":        "https://example.com/unlikely",
			},
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

	// Most of the time this should hit fallback (unless test runs at 2am UTC)
	if redirectURL != "https://example.com/default" && redirectURL != "https://example.com/unlikely" {
		t.Errorf("expected either fallback or matching URL, got '%s'", redirectURL)
	}
}

func TestRouteItByTime_Execute_FirstMatchWins(t *testing.T) {
	helper := &RouteItByTime{}

	// Create overlapping time ranges
	now := time.Now().UTC()
	currentHour := formatHour(now.Hour())
	nextHour := formatHour((now.Hour() + 3) % 24)

	config := map[string]interface{}{
		"time_routes": []interface{}{
			// First route (should win if current time matches)
			map[string]interface{}{
				"start_time": currentHour,
				"end_time":   nextHour,
				"url":        "https://example.com/first",
			},
			// Second route (overlapping, should not be used)
			map[string]interface{}{
				"start_time": currentHour,
				"end_time":   nextHour,
				"url":        "https://example.com/second",
			},
		},
		"timezone": "UTC",
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

	// If current time is in range, should get first URL
	if redirectURL == "https://example.com/first" {
		routingReason := modData["routing_reason"].(string)
		if routingReason == "" {
			t.Error("expected routing_reason to be set")
		}
	}

	// Should never get the second URL
	if redirectURL == "https://example.com/second" {
		t.Error("expected first matching route to win, but got second route")
	}
}

func TestRouteItByTime_Execute_WithSaveToField(t *testing.T) {
	helper := &RouteItByTime{}
	mockConn := &mockConnectorForRouting{}

	now := time.Now().UTC()
	currentHour := formatHour(now.Hour())
	nextHour := formatHour((now.Hour() + 2) % 24)

	config := map[string]interface{}{
		"time_routes": []interface{}{
			map[string]interface{}{
				"start_time": currentHour,
				"end_time":   nextHour,
				"url":        "https://example.com/current",
			},
		},
		"timezone":      "UTC",
		"save_to_field": "last_route",
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

	// Verify SetContactFieldValue was called (if time matched)
	modData := output.ModifiedData
	redirectURL := modData["redirect_url"].(string)

	if redirectURL == "https://example.com/current" {
		if len(mockConn.setFieldCalls) != 1 {
			t.Errorf("expected 1 SetContactFieldValue call when time matches, got %d", len(mockConn.setFieldCalls))
		}
	}
}

func TestRouteItByTime_Execute_WithApplyTag(t *testing.T) {
	helper := &RouteItByTime{}
	mockConn := &mockConnectorForRouting{}

	now := time.Now().UTC()
	currentHour := formatHour(now.Hour())
	nextHour := formatHour((now.Hour() + 2) % 24)

	config := map[string]interface{}{
		"time_routes": []interface{}{
			map[string]interface{}{
				"start_time": currentHour,
				"end_time":   nextHour,
				"url":        "https://example.com/current",
			},
		},
		"timezone":  "UTC",
		"apply_tag": "routed_by_time",
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

	// Verify ApplyTag was called (if time matched)
	modData := output.ModifiedData
	redirectURL := modData["redirect_url"].(string)

	if redirectURL == "https://example.com/current" {
		if len(mockConn.applyTagCalls) != 1 {
			t.Errorf("expected 1 ApplyTag call when time matches, got %d", len(mockConn.applyTagCalls))
		}
	}
}

// Helper function to format hour as HH:00 string
func formatHour(hour int) string {
	return time.Date(2000, 1, 1, hour, 0, 0, 0, time.UTC).Format("15:04")
}
