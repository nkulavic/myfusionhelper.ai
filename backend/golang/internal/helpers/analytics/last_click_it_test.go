package analytics

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/myfusionhelper/api/internal/connectors"
	"github.com/myfusionhelper/api/internal/helpers"
)

// mockConnectorLastClick is a minimal mock for testing last_click_it
type mockConnectorLastClick struct {
	setContactFieldValueFunc func(ctx context.Context, contactID string, fieldKey string, value interface{}) error
	getContactFieldValueFunc func(ctx context.Context, contactID string, fieldKey string) (interface{}, error)
	fieldsSet                map[string]string // Track fields that were set
}

func (m *mockConnectorLastClick) SetContactFieldValue(ctx context.Context, contactID string, fieldKey string, value interface{}) error {
	if m.setContactFieldValueFunc != nil {
		return m.setContactFieldValueFunc(ctx, contactID, fieldKey, value)
	}
	if m.fieldsSet == nil {
		m.fieldsSet = make(map[string]string)
	}
	m.fieldsSet[fieldKey] = value.(string)
	return nil
}

func (m *mockConnectorLastClick) GetContactFieldValue(ctx context.Context, contactID string, fieldKey string) (interface{}, error) {
	if m.getContactFieldValueFunc != nil {
		return m.getContactFieldValueFunc(ctx, contactID, fieldKey)
	}
	return nil, nil
}

// Stub methods to satisfy CRMConnector interface
func (m *mockConnectorLastClick) GetContacts(ctx context.Context, opts connectors.QueryOptions) (*connectors.ContactList, error) {
	return nil, nil
}
func (m *mockConnectorLastClick) GetContact(ctx context.Context, contactID string) (*connectors.NormalizedContact, error) {
	return nil, nil
}
func (m *mockConnectorLastClick) CreateContact(ctx context.Context, contact connectors.CreateContactInput) (*connectors.NormalizedContact, error) {
	return nil, nil
}
func (m *mockConnectorLastClick) UpdateContact(ctx context.Context, contactID string, updates connectors.UpdateContactInput) (*connectors.NormalizedContact, error) {
	return nil, nil
}
func (m *mockConnectorLastClick) DeleteContact(ctx context.Context, contactID string) error {
	return nil
}
func (m *mockConnectorLastClick) GetTags(ctx context.Context) ([]connectors.Tag, error) {
	return nil, nil
}
func (m *mockConnectorLastClick) ApplyTag(ctx context.Context, contactID string, tagID string) error {
	return nil
}
func (m *mockConnectorLastClick) RemoveTag(ctx context.Context, contactID string, tagID string) error {
	return nil
}
func (m *mockConnectorLastClick) GetCustomFields(ctx context.Context) ([]connectors.CustomField, error) {
	return nil, nil
}
func (m *mockConnectorLastClick) TriggerAutomation(ctx context.Context, contactID string, automationID string) error {
	return nil
}
func (m *mockConnectorLastClick) AchieveGoal(ctx context.Context, contactID string, goalName string, integration string) error {
	return nil
}
func (m *mockConnectorLastClick) SetOptInStatus(ctx context.Context, contactID string, optIn bool, reason string) error {
	return nil
}
func (m *mockConnectorLastClick) TestConnection(ctx context.Context) error {
	return nil
}
func (m *mockConnectorLastClick) GetMetadata() connectors.ConnectorMetadata {
	return connectors.ConnectorMetadata{}
}
func (m *mockConnectorLastClick) GetCapabilities() []connectors.Capability {
	return nil
}

func TestLastClickIt_GetMetadata(t *testing.T) {
	helper := &LastClickIt{}

	if helper.GetName() != "Last Click It" {
		t.Errorf("Expected name 'Last Click It', got '%s'", helper.GetName())
	}

	if helper.GetType() != "last_click_it" {
		t.Errorf("Expected type 'last_click_it', got '%s'", helper.GetType())
	}

	if helper.GetCategory() != "analytics" {
		t.Errorf("Expected category 'analytics', got '%s'", helper.GetCategory())
	}

	if !helper.RequiresCRM() {
		t.Error("Expected RequiresCRM to be true")
	}

	if helper.SupportedCRMs() != nil {
		t.Errorf("Expected SupportedCRMs to be nil (all CRMs), got %v", helper.SupportedCRMs())
	}
}

func TestLastClickIt_ValidateConfig(t *testing.T) {
	helper := &LastClickIt{}

	tests := []struct {
		name      string
		config    map[string]interface{}
		expectErr bool
	}{
		{
			name:      "valid minimal config",
			config:    map[string]interface{}{"utm_source": "google", "utm_medium": "cpc"},
			expectErr: false,
		},
		{
			name: "valid full config",
			config: map[string]interface{}{
				"utm_source":   "facebook",
				"utm_medium":   "social",
				"utm_campaign": "spring_sale",
				"utm_term":     "running shoes",
				"utm_content":  "ad_variant_a",
			},
			expectErr: false,
		},
		{
			name:      "missing utm_source",
			config:    map[string]interface{}{"utm_medium": "cpc"},
			expectErr: true,
		},
		{
			name:      "missing utm_medium",
			config:    map[string]interface{}{"utm_source": "google"},
			expectErr: true,
		},
		{
			name:      "empty utm_source",
			config:    map[string]interface{}{"utm_source": "", "utm_medium": "cpc"},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := helper.ValidateConfig(tt.config)
			if tt.expectErr && err == nil {
				t.Error("Expected error, got nil")
			}
			if !tt.expectErr && err != nil {
				t.Errorf("Expected no error, got %v", err)
			}
		})
	}
}

func TestLastClickIt_Execute_Success(t *testing.T) {
	helper := &LastClickIt{}
	connector := &mockConnectorLastClick{}

	input := helpers.HelperInput{
		Connector: connector,
		ContactID: "contact_123",
		Config: map[string]interface{}{
			"utm_source":   "google",
			"utm_medium":   "cpc",
			"utm_campaign": "summer_sale",
			"utm_term":     "buy shoes",
			"utm_content":  "text_ad_1",
		},
	}

	output, err := helper.Execute(context.Background(), input)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if !output.Success {
		t.Error("Expected success to be true")
	}

	if !strings.Contains(output.Message, "google/cpc") {
		t.Errorf("Expected message to contain 'google/cpc', got '%s'", output.Message)
	}

	// Verify all expected fields were set
	expectedFields := []string{
		"last_click_source",
		"last_click_medium",
		"last_click_campaign",
		"last_click_term",
		"last_click_content",
		"last_click_timestamp",
	}

	for _, field := range expectedFields {
		if _, exists := connector.fieldsSet[field]; !exists {
			t.Errorf("Expected field '%s' to be set", field)
		}
	}

	// Verify actions were recorded
	if len(output.Actions) != len(expectedFields) {
		t.Errorf("Expected %d actions, got %d", len(expectedFields), len(output.Actions))
	}
}

func TestLastClickIt_Execute_MinimalConfig(t *testing.T) {
	helper := &LastClickIt{}
	connector := &mockConnectorLastClick{}

	input := helpers.HelperInput{
		Connector: connector,
		ContactID: "contact_456",
		Config: map[string]interface{}{
			"utm_source": "newsletter",
			"utm_medium": "email",
		},
	}

	output, err := helper.Execute(context.Background(), input)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if !output.Success {
		t.Error("Expected success to be true")
	}

	// Only source, medium, and timestamp should be set
	expectedFields := []string{
		"last_click_source",
		"last_click_medium",
		"last_click_timestamp",
	}

	if len(connector.fieldsSet) != len(expectedFields) {
		t.Errorf("Expected %d fields to be set, got %d", len(expectedFields), len(connector.fieldsSet))
	}

	for _, field := range expectedFields {
		if _, exists := connector.fieldsSet[field]; !exists {
			t.Errorf("Expected field '%s' to be set", field)
		}
	}
}

func TestLastClickIt_Execute_CustomFieldPrefix(t *testing.T) {
	helper := &LastClickIt{}
	connector := &mockConnectorLastClick{}

	input := helpers.HelperInput{
		Connector: connector,
		ContactID: "contact_789",
		Config: map[string]interface{}{
			"utm_source":   "linkedin",
			"utm_medium":   "social",
			"field_prefix": "attribution_",
		},
	}

	_, err := helper.Execute(context.Background(), input)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify custom prefix was used
	if _, exists := connector.fieldsSet["attribution_source"]; !exists {
		t.Error("Expected 'attribution_source' to be set with custom prefix")
	}
	if _, exists := connector.fieldsSet["attribution_medium"]; !exists {
		t.Error("Expected 'attribution_medium' to be set with custom prefix")
	}
}

func TestLastClickIt_Execute_NoOverwrite(t *testing.T) {
	helper := &LastClickIt{}

	// Mock connector that returns existing data
	connector := &mockConnectorLastClick{
		getContactFieldValueFunc: func(ctx context.Context, contactID string, fieldKey string) (interface{}, error) {
			if fieldKey == "last_click_source" {
				return "existing_source", nil
			}
			return nil, nil
		},
	}

	input := helpers.HelperInput{
		Connector: connector,
		ContactID: "contact_existing",
		Config: map[string]interface{}{
			"utm_source": "new_source",
			"utm_medium": "new_medium",
			"overwrite":  false,
		},
	}

	output, err := helper.Execute(context.Background(), input)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if !output.Success {
		t.Error("Expected success to be true")
	}

	if !strings.Contains(output.Message, "not overwriting") {
		t.Errorf("Expected message to indicate not overwriting, got '%s'", output.Message)
	}

	// No fields should have been set
	if len(connector.fieldsSet) != 0 {
		t.Errorf("Expected no fields to be set when overwrite=false and data exists, got %d", len(connector.fieldsSet))
	}
}

func TestLastClickIt_Execute_ConnectorError(t *testing.T) {
	helper := &LastClickIt{}

	connector := &mockConnectorLastClick{
		setContactFieldValueFunc: func(ctx context.Context, contactID string, fieldKey string, value interface{}) error {
			return errors.New("CRM API error")
		},
	}

	input := helpers.HelperInput{
		Connector: connector,
		ContactID: "contact_error",
		Config: map[string]interface{}{
			"utm_source": "google",
			"utm_medium": "cpc",
		},
	}

	output, err := helper.Execute(context.Background(), input)

	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	if output.Success {
		t.Error("Expected success to be false")
	}

	if !strings.Contains(output.Message, "Partially updated") || !strings.Contains(output.Message, "failed") {
		t.Errorf("Expected error message about partial update, got '%s'", output.Message)
	}
}

func TestLastClickIt_Execute_TimestampFormat(t *testing.T) {
	helper := &LastClickIt{}
	connector := &mockConnectorLastClick{}

	input := helpers.HelperInput{
		Connector: connector,
		ContactID: "contact_timestamp",
		Config: map[string]interface{}{
			"utm_source": "twitter",
			"utm_medium": "social",
		},
	}

	_, err := helper.Execute(context.Background(), input)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify timestamp field exists and is in RFC3339 format
	timestamp, exists := connector.fieldsSet["last_click_timestamp"]
	if !exists {
		t.Fatal("Expected timestamp field to be set")
	}

	// Timestamp should be a valid RFC3339 string
	if !strings.Contains(timestamp, "T") || !strings.Contains(timestamp, "Z") {
		t.Errorf("Expected RFC3339 timestamp format, got '%s'", timestamp)
	}
}
