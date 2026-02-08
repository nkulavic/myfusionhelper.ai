package integration

import (
	"context"
	"testing"

	"github.com/myfusionhelper/api/internal/helpers"
)

func TestStripeHooks_Metadata(t *testing.T) {
	helper := &StripeHooks{}

	if name := helper.GetName(); name != "Stripe Hooks" {
		t.Errorf("expected name 'Stripe Hooks', got '%s'", name)
	}

	if helperType := helper.GetType(); helperType != "stripe_hooks" {
		t.Errorf("expected type 'stripe_hooks', got '%s'", helperType)
	}

	if category := helper.GetCategory(); category != "integration" {
		t.Errorf("expected category 'integration', got '%s'", category)
	}

	if !helper.RequiresCRM() {
		t.Error("expected RequiresCRM to be true")
	}

	if crms := helper.SupportedCRMs(); crms != nil {
		t.Errorf("expected SupportedCRMs to be nil (all CRMs), got %v", crms)
	}

	schema := helper.GetConfigSchema()
	if schema == nil {
		t.Fatal("expected config schema to be non-nil")
	}
}

func TestStripeHooks_ValidateConfig(t *testing.T) {
	helper := &StripeHooks{}

	tests := []struct {
		name        string
		config      map[string]interface{}
		expectError bool
	}{
		{
			name: "valid config with single event",
			config: map[string]interface{}{
				"selected_events": []interface{}{"charge.succeeded"},
			},
			expectError: false,
		},
		{
			name: "valid config with multiple events",
			config: map[string]interface{}{
				"selected_events": []interface{}{
					"charge.succeeded",
					"charge.failed",
					"customer.subscription.created",
				},
			},
			expectError: false,
		},
		{
			name: "valid config with goal",
			config: map[string]interface{}{
				"selected_events": []interface{}{"invoice.payment_succeeded"},
				"goal_name":       "payment_received",
			},
			expectError: false,
		},
		{
			name: "valid config with tags",
			config: map[string]interface{}{
				"selected_events": []interface{}{"customer.subscription.created"},
				"event_tags":      []interface{}{"subscriber", "active"},
			},
			expectError: false,
		},
		{
			name: "valid config with all options",
			config: map[string]interface{}{
				"selected_events": []interface{}{"charge.succeeded"},
				"goal_name":       "charge_completed",
				"event_tags":      []interface{}{"paid"},
			},
			expectError: false,
		},
		{
			name: "missing selected_events",
			config: map[string]interface{}{
				"goal_name": "test_goal",
			},
			expectError: true,
		},
		{
			name: "empty selected_events",
			config: map[string]interface{}{
				"selected_events": []interface{}{},
			},
			expectError: true,
		},
		{
			name: "selected_events not an array",
			config: map[string]interface{}{
				"selected_events": "charge.succeeded",
			},
			expectError: true,
		},
		{
			name: "selected_events contains non-string",
			config: map[string]interface{}{
				"selected_events": []interface{}{"charge.succeeded", 123},
			},
			expectError: true,
		},
		{
			name: "event_tags contains non-string",
			config: map[string]interface{}{
				"selected_events": []interface{}{"charge.succeeded"},
				"event_tags":      []interface{}{"valid_tag", 456},
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

func TestStripeHooks_Execute_SingleEvent(t *testing.T) {
	helper := &StripeHooks{}

	config := map[string]interface{}{
		"selected_events": []interface{}{"charge.succeeded"},
	}

	input := helpers.HelperInput{
		Config:     config,
		ContactID:  "contact123",
		Connector:  nil,
		AccountID:  "account456",
		UserID:     "user789",
	}

	output, err := helper.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if !output.Success {
		t.Error("expected success to be true")
	}

	modData := output.ModifiedData
	configuredEvents, ok := modData["configured_events"].([]interface{})
	if !ok {
		t.Fatal("expected configured_events in ModifiedData")
	}

	if len(configuredEvents) != 1 {
		t.Errorf("expected 1 configured event, got %d", len(configuredEvents))
	}

	if configuredEvents[0] != "charge.succeeded" {
		t.Errorf("expected event 'charge.succeeded', got '%v'", configuredEvents[0])
	}
}

func TestStripeHooks_Execute_MultipleEvents(t *testing.T) {
	helper := &StripeHooks{}

	events := []interface{}{
		"charge.succeeded",
		"charge.failed",
		"customer.subscription.created",
		"invoice.payment_succeeded",
	}

	config := map[string]interface{}{
		"selected_events": events,
	}

	input := helpers.HelperInput{
		Config:     config,
		ContactID:  "contact123",
		Connector:  nil,
		AccountID:  "account456",
		UserID:     "user789",
	}

	output, err := helper.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if !output.Success {
		t.Error("expected success to be true")
	}

	modData := output.ModifiedData
	configuredEvents, ok := modData["configured_events"].([]interface{})
	if !ok {
		t.Fatal("expected configured_events in ModifiedData")
	}

	if len(configuredEvents) != 4 {
		t.Errorf("expected 4 configured events, got %d", len(configuredEvents))
	}
}

func TestStripeHooks_Execute_WithGoal(t *testing.T) {
	helper := &StripeHooks{}

	config := map[string]interface{}{
		"selected_events": []interface{}{"invoice.payment_succeeded"},
		"goal_name":       "payment_received",
	}

	input := helpers.HelperInput{
		Config:     config,
		ContactID:  "contact123",
		Connector:  nil,
		AccountID:  "account456",
		UserID:     "user789",
	}

	output, err := helper.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if !output.Success {
		t.Error("expected success to be true")
	}

	modData := output.ModifiedData
	goalName, ok := modData["goal_name"].(string)
	if !ok {
		t.Fatal("expected goal_name in ModifiedData")
	}

	if goalName != "payment_received" {
		t.Errorf("expected goal 'payment_received', got '%s'", goalName)
	}

	// Check logs mention goal
	foundGoalLog := false
	for _, log := range output.Logs {
		if log == "Will trigger goal: payment_received" {
			foundGoalLog = true
			break
		}
	}
	if !foundGoalLog {
		t.Error("expected log message about triggering goal")
	}
}

func TestStripeHooks_Execute_WithTags(t *testing.T) {
	helper := &StripeHooks{}

	config := map[string]interface{}{
		"selected_events": []interface{}{"customer.subscription.created"},
		"event_tags":      []interface{}{"subscriber", "active", "stripe_customer"},
	}

	input := helpers.HelperInput{
		Config:     config,
		ContactID:  "contact123",
		Connector:  nil,
		AccountID:  "account456",
		UserID:     "user789",
	}

	output, err := helper.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if !output.Success {
		t.Error("expected success to be true")
	}

	modData := output.ModifiedData
	eventTags, ok := modData["event_tags"].([]string)
	if !ok {
		t.Fatal("expected event_tags in ModifiedData")
	}

	if len(eventTags) != 3 {
		t.Errorf("expected 3 event tags, got %d", len(eventTags))
	}

	expectedTags := map[string]bool{
		"subscriber":      true,
		"active":          true,
		"stripe_customer": true,
	}

	for _, tag := range eventTags {
		if !expectedTags[tag] {
			t.Errorf("unexpected tag '%s'", tag)
		}
	}

	// Check logs mention tags
	foundTagsLog := false
	for _, log := range output.Logs {
		if log == "Will apply 3 tags on events" {
			foundTagsLog = true
			break
		}
	}
	if !foundTagsLog {
		t.Error("expected log message about applying tags")
	}
}

func TestStripeHooks_Execute_WithAllOptions(t *testing.T) {
	helper := &StripeHooks{}

	config := map[string]interface{}{
		"selected_events": []interface{}{"charge.succeeded", "invoice.payment_succeeded"},
		"goal_name":       "stripe_payment",
		"event_tags":      []interface{}{"paid", "customer"},
	}

	input := helpers.HelperInput{
		Config:     config,
		ContactID:  "contact123",
		Connector:  nil,
		AccountID:  "account456",
		UserID:     "user789",
	}

	output, err := helper.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if !output.Success {
		t.Error("expected success to be true")
	}

	// Verify all components in output
	modData := output.ModifiedData

	configuredEvents := modData["configured_events"].([]interface{})
	if len(configuredEvents) != 2 {
		t.Errorf("expected 2 events, got %d", len(configuredEvents))
	}

	goalName := modData["goal_name"].(string)
	if goalName != "stripe_payment" {
		t.Errorf("expected goal 'stripe_payment', got '%s'", goalName)
	}

	eventTags := modData["event_tags"].([]string)
	if len(eventTags) != 2 {
		t.Errorf("expected 2 tags, got %d", len(eventTags))
	}

	// Verify logs
	if len(output.Logs) < 3 {
		t.Errorf("expected at least 3 log entries, got %d", len(output.Logs))
	}
}

func TestStripeHooks_Execute_EmptyOptionalFields(t *testing.T) {
	helper := &StripeHooks{}

	config := map[string]interface{}{
		"selected_events": []interface{}{"charge.succeeded"},
		"goal_name":       "",
		"event_tags":      []interface{}{},
	}

	input := helpers.HelperInput{
		Config:     config,
		ContactID:  "contact123",
		Connector:  nil,
		AccountID:  "account456",
		UserID:     "user789",
	}

	output, err := helper.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if !output.Success {
		t.Error("expected success to be true")
	}

	modData := output.ModifiedData

	// Goal should be empty string
	goalName := modData["goal_name"].(string)
	if goalName != "" {
		t.Errorf("expected empty goal name, got '%s'", goalName)
	}

	// Event tags should be empty slice or nil
	eventTags, ok := modData["event_tags"].([]string)
	if ok && len(eventTags) > 0 {
		t.Errorf("expected no event tags, got %d", len(eventTags))
	}

	// Should only have one log entry (the configured events count)
	if len(output.Logs) != 1 {
		t.Errorf("expected 1 log entry, got %d", len(output.Logs))
	}
}
