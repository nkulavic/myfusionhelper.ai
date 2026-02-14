package data

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/myfusionhelper/api/internal/connectors"
	"github.com/myfusionhelper/api/internal/helpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Mock connector for when_is_it testing
type mockConnectorForWhenIsIt struct {
	fieldValues   map[string]interface{}
	setFieldCalls map[string]interface{}
	setFieldError error
}

func (m *mockConnectorForWhenIsIt) GetContactFieldValue(ctx context.Context, contactID, fieldKey string) (interface{}, error) {
	if m.fieldValues == nil {
		return nil, nil
	}
	return m.fieldValues[fieldKey], nil
}

func (m *mockConnectorForWhenIsIt) SetContactFieldValue(ctx context.Context, contactID, fieldKey string, value interface{}) error {
	if m.setFieldCalls == nil {
		m.setFieldCalls = make(map[string]interface{})
	}
	m.setFieldCalls[fieldKey] = value
	return m.setFieldError
}

// Stub implementations
func (m *mockConnectorForWhenIsIt) GetContact(ctx context.Context, contactID string) (*connectors.NormalizedContact, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForWhenIsIt) GetContacts(ctx context.Context, opts connectors.QueryOptions) (*connectors.ContactList, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForWhenIsIt) CreateContact(ctx context.Context, input connectors.CreateContactInput) (*connectors.NormalizedContact, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForWhenIsIt) UpdateContact(ctx context.Context, contactID string, updates connectors.UpdateContactInput) (*connectors.NormalizedContact, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForWhenIsIt) DeleteContact(ctx context.Context, contactID string) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorForWhenIsIt) GetTags(ctx context.Context) ([]connectors.Tag, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForWhenIsIt) ApplyTag(ctx context.Context, contactID, tagID string) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorForWhenIsIt) RemoveTag(ctx context.Context, contactID, tagID string) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorForWhenIsIt) GetCustomFields(ctx context.Context) ([]connectors.CustomField, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForWhenIsIt) TriggerAutomation(ctx context.Context, contactID, automationID string) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorForWhenIsIt) AchieveGoal(ctx context.Context, contactID, goalName, integration string) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorForWhenIsIt) TestConnection(ctx context.Context) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorForWhenIsIt) GetMetadata() connectors.ConnectorMetadata {
	return connectors.ConnectorMetadata{}
}
func (m *mockConnectorForWhenIsIt) GetCapabilities() []connectors.Capability {
	return nil
}

func TestWhenIsIt_Metadata(t *testing.T) {
	h := &WhenIsIt{}

	assert.Equal(t, "When Is It", h.GetName())
	assert.Equal(t, "when_is_it", h.GetType())
	assert.Equal(t, "data", h.GetCategory())
	assert.NotEmpty(t, h.GetDescription())
	assert.True(t, h.RequiresCRM())
	assert.Nil(t, h.SupportedCRMs())
}

func TestWhenIsIt_GetConfigSchema(t *testing.T) {
	h := &WhenIsIt{}
	schema := h.GetConfigSchema()

	assert.Equal(t, "object", schema["type"])

	props, ok := schema["properties"].(map[string]interface{})
	require.True(t, ok)
	assert.Contains(t, props, "source_field")
	assert.Contains(t, props, "target_field")
	assert.Contains(t, props, "to_timezone")
	assert.Contains(t, props, "output_format")

	required, ok := schema["required"].([]string)
	require.True(t, ok)
	assert.Contains(t, required, "source_field")
	assert.Contains(t, required, "target_field")
	assert.Contains(t, required, "to_timezone")
	assert.Contains(t, required, "output_format")
}

func TestWhenIsIt_ValidateConfig_Success(t *testing.T) {
	h := &WhenIsIt{}

	config := map[string]interface{}{
		"source_field":  "created_date",
		"target_field":  "formatted_date",
		"to_timezone":   "America/New_York",
		"output_format": "2006-01-02",
	}

	err := h.ValidateConfig(config)
	assert.NoError(t, err)
}

func TestWhenIsIt_ValidateConfig_Errors(t *testing.T) {
	h := &WhenIsIt{}

	tests := []struct {
		name    string
		config  map[string]interface{}
		wantErr string
	}{
		{
			name:    "missing source_field",
			config:  map[string]interface{}{"target_field": "t", "to_timezone": "UTC", "output_format": "2006"},
			wantErr: "source_field is required",
		},
		{
			name:    "missing target_field",
			config:  map[string]interface{}{"source_field": "s", "to_timezone": "UTC", "output_format": "2006"},
			wantErr: "target_field is required",
		},
		{
			name:    "missing to_timezone",
			config:  map[string]interface{}{"source_field": "s", "target_field": "t", "output_format": "2006"},
			wantErr: "to_timezone is required",
		},
		{
			name:    "missing output_format",
			config:  map[string]interface{}{"source_field": "s", "target_field": "t", "to_timezone": "UTC"},
			wantErr: "output_format is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := h.ValidateConfig(tt.config)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErr)
		})
	}
}

func TestWhenIsIt_Execute_Success_UTC_to_EST(t *testing.T) {
	h := &WhenIsIt{}
	mockConnector := &mockConnectorForWhenIsIt{
		fieldValues: map[string]interface{}{
			"created_date": "2024-01-15T10:00:00",
		},
	}

	input := helpers.HelperInput{
		ContactID: "contact_123",
		Connector: mockConnector,
		Config: map[string]interface{}{
			"source_field":  "created_date",
			"target_field":  "formatted_date",
			"from_timezone": "UTC",
			"to_timezone":   "America/New_York",
			"output_format": time.RFC3339,
		},
	}

	output, err := h.Execute(context.Background(), input)

	assert.NoError(t, err)
	assert.True(t, output.Success)
	assert.NotNil(t, mockConnector.setFieldCalls["formatted_date"])
}

func TestWhenIsIt_Execute_EmptyField(t *testing.T) {
	h := &WhenIsIt{}
	mockConnector := &mockConnectorForWhenIsIt{
		fieldValues: map[string]interface{}{
			"created_date": "",
		},
	}

	input := helpers.HelperInput{
		ContactID: "contact_456",
		Connector: mockConnector,
		Config: map[string]interface{}{
			"source_field":  "created_date",
			"target_field":  "formatted_date",
			"to_timezone":   "UTC",
			"output_format": "2006-01-02",
		},
	}

	output, err := h.Execute(context.Background(), input)

	assert.NoError(t, err)
	assert.True(t, output.Success)
	assert.Contains(t, output.Message, "empty")
	assert.Nil(t, mockConnector.setFieldCalls["formatted_date"])
}

func TestWhenIsIt_Execute_InvalidTimezone(t *testing.T) {
	h := &WhenIsIt{}
	mockConnector := &mockConnectorForWhenIsIt{
		fieldValues: map[string]interface{}{
			"created_date": "2024-01-15",
		},
	}

	input := helpers.HelperInput{
		ContactID: "contact_789",
		Connector: mockConnector,
		Config: map[string]interface{}{
			"source_field":  "created_date",
			"target_field":  "formatted_date",
			"to_timezone":   "Invalid/Timezone",
			"output_format": "2006-01-02",
		},
	}

	output, err := h.Execute(context.Background(), input)

	assert.Error(t, err)
	assert.Contains(t, output.Message, "Invalid target timezone")
}

func TestWhenIsIt_Execute_CustomFormat(t *testing.T) {
	h := &WhenIsIt{}
	mockConnector := &mockConnectorForWhenIsIt{
		fieldValues: map[string]interface{}{
			"created_date": "2024-01-15T10:30:00",
		},
	}

	input := helpers.HelperInput{
		ContactID: "contact_format",
		Connector: mockConnector,
		Config: map[string]interface{}{
			"source_field":  "created_date",
			"target_field":  "formatted_date",
			"to_timezone":   "UTC",
			"output_format": "Jan 2, 2006",
		},
	}

	output, err := h.Execute(context.Background(), input)

	assert.NoError(t, err)
	assert.True(t, output.Success)
	result := mockConnector.setFieldCalls["formatted_date"].(string)
	assert.Contains(t, result, "Jan")
	assert.Contains(t, result, "2024")
}

func TestWhenIsIt_Execute_DateOnly(t *testing.T) {
	h := &WhenIsIt{}
	mockConnector := &mockConnectorForWhenIsIt{
		fieldValues: map[string]interface{}{
			"created_date": "2024-01-15",
		},
	}

	input := helpers.HelperInput{
		ContactID: "contact_date",
		Connector: mockConnector,
		Config: map[string]interface{}{
			"source_field":  "created_date",
			"target_field":  "formatted_date",
			"to_timezone":   "UTC",
			"output_format": "2006-01-02",
		},
	}

	output, err := h.Execute(context.Background(), input)

	assert.NoError(t, err)
	assert.True(t, output.Success)
	assert.Equal(t, "2024-01-15", mockConnector.setFieldCalls["formatted_date"])
}

func TestWhenIsIt_Execute_DefaultFromTimezone(t *testing.T) {
	h := &WhenIsIt{}
	mockConnector := &mockConnectorForWhenIsIt{
		fieldValues: map[string]interface{}{
			"created_date": "2024-01-15T10:00:00",
		},
	}

	input := helpers.HelperInput{
		ContactID: "contact_default",
		Connector: mockConnector,
		Config: map[string]interface{}{
			"source_field":  "created_date",
			"target_field":  "formatted_date",
			"to_timezone":   "America/Los_Angeles",
			"output_format": "2006-01-02 15:04:05",
		},
	}

	output, err := h.Execute(context.Background(), input)

	assert.NoError(t, err)
	assert.True(t, output.Success)
}

func TestWhenIsIt_Execute_Actions(t *testing.T) {
	h := &WhenIsIt{}
	mockConnector := &mockConnectorForWhenIsIt{
		fieldValues: map[string]interface{}{
			"created_date": "2024-01-15",
		},
	}

	input := helpers.HelperInput{
		ContactID: "contact_actions",
		Connector: mockConnector,
		Config: map[string]interface{}{
			"source_field":  "created_date",
			"target_field":  "formatted_date",
			"to_timezone":   "UTC",
			"output_format": "2006-01-02",
		},
	}

	output, err := h.Execute(context.Background(), input)

	assert.NoError(t, err)
	assert.Len(t, output.Actions, 1)
	assert.Equal(t, "field_updated", output.Actions[0].Type)
	assert.Equal(t, "formatted_date", output.Actions[0].Target)
}

func TestWhenIsIt_Execute_ModifiedData(t *testing.T) {
	h := &WhenIsIt{}
	mockConnector := &mockConnectorForWhenIsIt{
		fieldValues: map[string]interface{}{
			"created_date": "2024-01-15",
		},
	}

	input := helpers.HelperInput{
		ContactID: "contact_data",
		Connector: mockConnector,
		Config: map[string]interface{}{
			"source_field":  "created_date",
			"target_field":  "formatted_date",
			"to_timezone":   "UTC",
			"output_format": "2006-01-02",
		},
	}

	output, err := h.Execute(context.Background(), input)

	assert.NoError(t, err)
	require.NotNil(t, output.ModifiedData)
	assert.Contains(t, output.ModifiedData, "formatted_date")
}

func TestWhenIsIt_Execute_Logs(t *testing.T) {
	h := &WhenIsIt{}
	mockConnector := &mockConnectorForWhenIsIt{
		fieldValues: map[string]interface{}{
			"created_date": "2024-01-15",
		},
	}

	input := helpers.HelperInput{
		ContactID: "contact_logs",
		Connector: mockConnector,
		Config: map[string]interface{}{
			"source_field":  "created_date",
			"target_field":  "formatted_date",
			"to_timezone":   "UTC",
			"output_format": "2006-01-02",
		},
	}

	output, err := h.Execute(context.Background(), input)

	assert.NoError(t, err)
	assert.NotEmpty(t, output.Logs)
	assert.Contains(t, output.Logs[0], "Converted")
}

func TestWhenIsIt_Execute_SetFieldError(t *testing.T) {
	h := &WhenIsIt{}
	mockConnector := &mockConnectorForWhenIsIt{
		fieldValues: map[string]interface{}{
			"created_date": "2024-01-15",
		},
		setFieldError: fmt.Errorf("failed to set field"),
	}

	input := helpers.HelperInput{
		ContactID: "contact_error",
		Connector: mockConnector,
		Config: map[string]interface{}{
			"source_field":  "created_date",
			"target_field":  "formatted_date",
			"to_timezone":   "UTC",
			"output_format": "2006-01-02",
		},
	}

	output, err := h.Execute(context.Background(), input)

	assert.Error(t, err)
	assert.Contains(t, output.Message, "Failed to set result field")
}
