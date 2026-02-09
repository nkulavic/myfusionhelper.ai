package automation

import (
	"context"
	"strings"
	"testing"

	"github.com/myfusionhelper/api/internal/connectors"
	"github.com/myfusionhelper/api/internal/helpers"
)

func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

// mockConnectorForRouting implements the minimal CRMConnector interface needed for routing tests
type mockConnectorForRouting struct {
	fieldValues        map[string]interface{}
	setFieldCalls      []setFieldCall
	applyTagCalls      []string
	setFieldError      error
	getFieldError      error
	applyTagError      error
}

type setFieldCall struct {
	contactID string
	fieldKey  string
	value     interface{}
}

func (m *mockConnectorForRouting) GetContactFieldValue(ctx context.Context, contactID string, fieldKey string) (interface{}, error) {
	if m.getFieldError != nil {
		return nil, m.getFieldError
	}
	if val, ok := m.fieldValues[fieldKey]; ok {
		return val, nil
	}
	return "", nil
}

func (m *mockConnectorForRouting) SetContactFieldValue(ctx context.Context, contactID string, fieldKey string, value interface{}) error {
	m.setFieldCalls = append(m.setFieldCalls, setFieldCall{
		contactID: contactID,
		fieldKey:  fieldKey,
		value:     value,
	})
	return m.setFieldError
}

func (m *mockConnectorForRouting) ApplyTag(ctx context.Context, contactID string, tagID string) error {
	m.applyTagCalls = append(m.applyTagCalls, tagID)
	return m.applyTagError
}

// Minimal implementations of other required methods
func (m *mockConnectorForRouting) GetContacts(ctx context.Context, opts connectors.QueryOptions) (*connectors.ContactList, error) {
	return nil, nil
}
func (m *mockConnectorForRouting) GetContact(ctx context.Context, contactID string) (*connectors.NormalizedContact, error) {
	return nil, nil
}
func (m *mockConnectorForRouting) CreateContact(ctx context.Context, contact connectors.CreateContactInput) (*connectors.NormalizedContact, error) {
	return nil, nil
}
func (m *mockConnectorForRouting) UpdateContact(ctx context.Context, contactID string, updates connectors.UpdateContactInput) (*connectors.NormalizedContact, error) {
	return nil, nil
}
func (m *mockConnectorForRouting) DeleteContact(ctx context.Context, contactID string) error {
	return nil
}
func (m *mockConnectorForRouting) GetTags(ctx context.Context) ([]connectors.Tag, error) {
	return nil, nil
}
func (m *mockConnectorForRouting) RemoveTag(ctx context.Context, contactID string, tagID string) error {
	return nil
}
func (m *mockConnectorForRouting) GetCustomFields(ctx context.Context) ([]connectors.CustomField, error) {
	return nil, nil
}
func (m *mockConnectorForRouting) TriggerAutomation(ctx context.Context, contactID string, automationID string) error {
	return nil
}
func (m *mockConnectorForRouting) AchieveGoal(ctx context.Context, contactID string, goalName string, integration string) error {
	return nil
}
func (m *mockConnectorForRouting) TestConnection(ctx context.Context) error {
	return nil
}
func (m *mockConnectorForRouting) GetMetadata() connectors.ConnectorMetadata {
	return connectors.ConnectorMetadata{}
}
func (m *mockConnectorForRouting) GetCapabilities() []connectors.Capability {
	return nil
}
func (m *mockConnectorForRouting) SetOptInStatus(ctx context.Context, contactID string, optIn bool, reason string) error {
	return nil
}

func TestRouteItByCustom_Metadata(t *testing.T) {
	helper := &RouteItByCustom{}

	if name := helper.GetName(); name != "Route It By Custom Field" {
		t.Errorf("expected name 'Route It By Custom Field', got '%s'", name)
	}

	if helperType := helper.GetType(); helperType != "route_it_by_custom" {
		t.Errorf("expected type 'route_it_by_custom', got '%s'", helperType)
	}

	if category := helper.GetCategory(); category != "automation" {
		t.Errorf("expected category 'automation', got '%s'", category)
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

func TestRouteItByCustom_ValidateConfig(t *testing.T) {
	helper := &RouteItByCustom{}

	tests := []struct {
		name        string
		config      map[string]interface{}
		expectError bool
	}{
		{
			name: "valid config with field name and value routes",
			config: map[string]interface{}{
				"field_name": "status",
				"value_routes": map[string]interface{}{
					"premium": "https://example.com/premium",
					"basic":   "https://example.com/basic",
				},
			},
			expectError: false,
		},
		{
			name: "valid config with fallback URL",
			config: map[string]interface{}{
				"field_name": "tier",
				"value_routes": map[string]interface{}{
					"gold": "https://example.com/gold",
				},
				"fallback_url": "https://example.com/default",
			},
			expectError: false,
		},
		{
			name: "missing field_name",
			config: map[string]interface{}{
				"value_routes": map[string]interface{}{
					"premium": "https://example.com/premium",
				},
			},
			expectError: true,
		},
		{
			name: "missing value_routes",
			config: map[string]interface{}{
				"field_name": "status",
			},
			expectError: true,
		},
		{
			name: "empty value_routes",
			config: map[string]interface{}{
				"field_name":   "status",
				"value_routes": map[string]interface{}{},
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

func TestRouteItByCustom_Execute_MatchingValue(t *testing.T) {
	helper := &RouteItByCustom{}
	mockConn := &mockConnectorForRouting{
		fieldValues: map[string]interface{}{
			"status": "premium",
		},
	}

	config := map[string]interface{}{
		"field_name": "status",
		"value_routes": map[string]interface{}{
			"premium": "https://example.com/premium",
			"basic":   "https://example.com/basic",
		},
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

	modData := output.ModifiedData

	redirectURL, ok := modData["redirect_url"].(string)
	if !ok || redirectURL != "https://example.com/premium" {
		t.Errorf("expected redirect_url 'https://example.com/premium', got '%v'", redirectURL)
	}

	routingReason, ok := modData["routing_reason"].(string)
	if !ok || routingReason == "" {
		t.Errorf("expected routing_reason to be set, got '%v'", routingReason)
	}
}

func TestRouteItByCustom_Execute_FallbackURL(t *testing.T) {
	helper := &RouteItByCustom{}
	mockConn := &mockConnectorForRouting{
		fieldValues: map[string]interface{}{
			"status": "trial",
		},
	}

	config := map[string]interface{}{
		"field_name": "status",
		"value_routes": map[string]interface{}{
			"premium": "https://example.com/premium",
			"basic":   "https://example.com/basic",
		},
		"fallback_url": "https://example.com/default",
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

func TestRouteItByCustom_Execute_WithSaveToField(t *testing.T) {
	helper := &RouteItByCustom{}
	mockConn := &mockConnectorForRouting{
		fieldValues: map[string]interface{}{
			"status": "premium",
		},
	}

	config := map[string]interface{}{
		"field_name": "status",
		"value_routes": map[string]interface{}{
			"premium": "https://example.com/premium",
		},
		"save_to_field": "redirect_url",
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
	if call.fieldKey != "redirect_url" {
		t.Errorf("expected field key 'redirect_url', got '%s'", call.fieldKey)
	}
	if call.value != "https://example.com/premium" {
		t.Errorf("expected value 'https://example.com/premium', got '%v'", call.value)
	}
}

func TestRouteItByCustom_Execute_WithApplyTag(t *testing.T) {
	helper := &RouteItByCustom{}
	mockConn := &mockConnectorForRouting{
		fieldValues: map[string]interface{}{
			"status": "premium",
		},
	}

	config := map[string]interface{}{
		"field_name": "status",
		"value_routes": map[string]interface{}{
			"premium": "https://example.com/premium",
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
