package data

import (
	"context"
	"fmt"
	"testing"

	"github.com/myfusionhelper/api/internal/connectors"
	"github.com/myfusionhelper/api/internal/helpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Mock connector for phone_lookup testing
type mockConnectorForPhoneLookup struct {
	fieldValues      map[string]interface{}
	getFieldError    error
	setFieldError    error
	setFieldCalls    map[string]interface{}
	achieveGoalCalls []string
}

func (m *mockConnectorForPhoneLookup) GetContactFieldValue(ctx context.Context, contactID, fieldKey string) (interface{}, error) {
	if m.getFieldError != nil {
		return nil, m.getFieldError
	}
	if m.fieldValues == nil {
		return nil, nil
	}
	return m.fieldValues[fieldKey], nil
}

func (m *mockConnectorForPhoneLookup) SetContactFieldValue(ctx context.Context, contactID, fieldKey string, value interface{}) error {
	if m.setFieldCalls == nil {
		m.setFieldCalls = make(map[string]interface{})
	}
	m.setFieldCalls[fieldKey] = value
	return m.setFieldError
}

func (m *mockConnectorForPhoneLookup) AchieveGoal(ctx context.Context, contactID, goalName, integration string) error {
	if m.achieveGoalCalls == nil {
		m.achieveGoalCalls = make([]string, 0)
	}
	m.achieveGoalCalls = append(m.achieveGoalCalls, goalName)
	return nil
}

// Stub implementations
func (m *mockConnectorForPhoneLookup) GetContact(ctx context.Context, contactID string) (*connectors.NormalizedContact, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForPhoneLookup) GetContacts(ctx context.Context, opts connectors.QueryOptions) (*connectors.ContactList, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForPhoneLookup) CreateContact(ctx context.Context, input connectors.CreateContactInput) (*connectors.NormalizedContact, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForPhoneLookup) UpdateContact(ctx context.Context, contactID string, updates connectors.UpdateContactInput) (*connectors.NormalizedContact, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForPhoneLookup) DeleteContact(ctx context.Context, contactID string) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorForPhoneLookup) GetTags(ctx context.Context) ([]connectors.Tag, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForPhoneLookup) ApplyTag(ctx context.Context, contactID, tagID string) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorForPhoneLookup) RemoveTag(ctx context.Context, contactID, tagID string) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorForPhoneLookup) GetCustomFields(ctx context.Context) ([]connectors.CustomField, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForPhoneLookup) TriggerAutomation(ctx context.Context, contactID, automationID string) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorForPhoneLookup) TestConnection(ctx context.Context) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorForPhoneLookup) GetMetadata() connectors.ConnectorMetadata {
	return connectors.ConnectorMetadata{}
}
func (m *mockConnectorForPhoneLookup) GetCapabilities() []connectors.Capability {
	return nil
}

func TestPhoneLookup_Metadata(t *testing.T) {
	h := &PhoneLookup{}

	assert.Equal(t, "Phone Lookup", h.GetName())
	assert.Equal(t, "phone_lookup", h.GetType())
	assert.Equal(t, "data", h.GetCategory())
	assert.NotEmpty(t, h.GetDescription())
	assert.True(t, h.RequiresCRM())
	assert.Nil(t, h.SupportedCRMs())
}

func TestPhoneLookup_GetConfigSchema(t *testing.T) {
	h := &PhoneLookup{}
	schema := h.GetConfigSchema()

	assert.Equal(t, "object", schema["type"])

	props, ok := schema["properties"].(map[string]interface{})
	require.True(t, ok)

	// Verify phone_field
	phoneField, ok := props["phone_field"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "string", phoneField["type"])

	// Verify goals
	assert.Contains(t, props, "valid_goal")
	assert.Contains(t, props, "invalid_goal")
	assert.Contains(t, props, "empty_goal")

	// Verify required
	required, ok := schema["required"].([]string)
	require.True(t, ok)
	assert.Contains(t, required, "phone_field")
}

func TestPhoneLookup_ValidateConfig_Success(t *testing.T) {
	h := &PhoneLookup{}

	config := map[string]interface{}{
		"phone_field": "Phone1",
	}

	err := h.ValidateConfig(config)
	assert.NoError(t, err)
}

func TestPhoneLookup_ValidateConfig_Error(t *testing.T) {
	h := &PhoneLookup{}

	tests := []struct {
		name    string
		config  map[string]interface{}
		wantErr string
	}{
		{
			name:    "missing phone_field",
			config:  map[string]interface{}{},
			wantErr: "phone_field is required",
		},
		{
			name: "empty phone_field",
			config: map[string]interface{}{
				"phone_field": "",
			},
			wantErr: "phone_field is required",
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

func TestPhoneLookup_Execute_EmptyPhone(t *testing.T) {
	h := &PhoneLookup{}
	mockConnector := &mockConnectorForPhoneLookup{
		fieldValues: map[string]interface{}{
			"Phone1": "",
		},
	}

	input := helpers.HelperInput{
		ContactID: "contact_123",
		Connector: mockConnector,
		Config: map[string]interface{}{
			"phone_field": "Phone1",
			"empty_goal":  "EmptyPhoneGoal",
		},
	}

	output, err := h.Execute(context.Background(), input)

	assert.NoError(t, err)
	assert.True(t, output.Success)
	assert.Contains(t, output.Message, "empty")
	assert.Contains(t, mockConnector.achieveGoalCalls, "EmptyPhoneGoal")
	assert.Equal(t, "empty", output.ModifiedData["status"])
}

func TestPhoneLookup_Execute_ValidUSPhone(t *testing.T) {
	h := &PhoneLookup{}
	mockConnector := &mockConnectorForPhoneLookup{
		fieldValues: map[string]interface{}{
			"Phone1": "(555) 123-4567",
		},
	}

	input := helpers.HelperInput{
		ContactID: "contact_456",
		Connector: mockConnector,
		Config: map[string]interface{}{
			"phone_field":        "Phone1",
			"country_code":       "US",
			"valid_goal":         "ValidPhoneGoal",
			"save_formatted_to": "CleanPhone",
		},
	}

	output, err := h.Execute(context.Background(), input)

	assert.NoError(t, err)
	assert.True(t, output.Success)
	assert.Contains(t, mockConnector.achieveGoalCalls, "ValidPhoneGoal")
	assert.Equal(t, "5551234567", mockConnector.setFieldCalls["CleanPhone"])
	assert.Equal(t, "valid", output.ModifiedData["status"])
}

func TestPhoneLookup_Execute_InvalidPhone(t *testing.T) {
	h := &PhoneLookup{}
	mockConnector := &mockConnectorForPhoneLookup{
		fieldValues: map[string]interface{}{
			"Phone1": "123",
		},
	}

	input := helpers.HelperInput{
		ContactID: "contact_789",
		Connector: mockConnector,
		Config: map[string]interface{}{
			"phone_field":   "Phone1",
			"country_code":  "US",
			"invalid_goal": "InvalidPhoneGoal",
		},
	}

	output, err := h.Execute(context.Background(), input)

	assert.NoError(t, err)
	assert.True(t, output.Success)
	assert.Contains(t, mockConnector.achieveGoalCalls, "InvalidPhoneGoal")
	assert.Equal(t, "invalid", output.ModifiedData["status"])
}

func TestPhoneLookup_Execute_OnlySpecialChars(t *testing.T) {
	h := &PhoneLookup{}
	mockConnector := &mockConnectorForPhoneLookup{
		fieldValues: map[string]interface{}{
			"Phone1": "()---",
		},
	}

	input := helpers.HelperInput{
		ContactID: "contact_999",
		Connector: mockConnector,
		Config: map[string]interface{}{
			"phone_field":   "Phone1",
			"invalid_goal": "InvalidPhoneGoal",
		},
	}

	output, err := h.Execute(context.Background(), input)

	assert.NoError(t, err)
	assert.True(t, output.Success)
	assert.Contains(t, output.Message, "no valid digits")
	assert.Contains(t, mockConnector.achieveGoalCalls, "InvalidPhoneGoal")
}

func TestPhoneLookup_Execute_ValidCAPhone(t *testing.T) {
	h := &PhoneLookup{}
	mockConnector := &mockConnectorForPhoneLookup{
		fieldValues: map[string]interface{}{
			"Phone1": "416-555-1234",
		},
	}

	input := helpers.HelperInput{
		ContactID: "contact_ca",
		Connector: mockConnector,
		Config: map[string]interface{}{
			"phone_field":  "Phone1",
			"country_code": "CA",
			"valid_goal":   "ValidPhoneGoal",
		},
	}

	output, err := h.Execute(context.Background(), input)

	assert.NoError(t, err)
	assert.True(t, output.Success)
	assert.Equal(t, "valid", output.ModifiedData["status"])
}

func TestPhoneLookup_Execute_ValidGBPhone(t *testing.T) {
	h := &PhoneLookup{}
	mockConnector := &mockConnectorForPhoneLookup{
		fieldValues: map[string]interface{}{
			"Phone1": "020 7946 0958",
		},
	}

	input := helpers.HelperInput{
		ContactID: "contact_gb",
		Connector: mockConnector,
		Config: map[string]interface{}{
			"phone_field":  "Phone1",
			"country_code": "GB",
			"valid_goal":   "ValidPhoneGoal",
		},
	}

	output, err := h.Execute(context.Background(), input)

	assert.NoError(t, err)
	assert.True(t, output.Success)
	assert.Equal(t, "valid", output.ModifiedData["status"])
}

func TestPhoneLookup_Execute_NoGoals(t *testing.T) {
	h := &PhoneLookup{}
	mockConnector := &mockConnectorForPhoneLookup{
		fieldValues: map[string]interface{}{
			"Phone1": "555-1234",
		},
	}

	input := helpers.HelperInput{
		ContactID: "contact_nogoals",
		Connector: mockConnector,
		Config: map[string]interface{}{
			"phone_field":  "Phone1",
			"country_code": "US",
		},
	}

	output, err := h.Execute(context.Background(), input)

	assert.NoError(t, err)
	assert.True(t, output.Success)
	assert.Len(t, mockConnector.achieveGoalCalls, 0)
}

func TestPhoneLookup_Execute_DefaultCountryUS(t *testing.T) {
	h := &PhoneLookup{}
	mockConnector := &mockConnectorForPhoneLookup{
		fieldValues: map[string]interface{}{
			"Phone1": "5551234567",
		},
	}

	input := helpers.HelperInput{
		ContactID: "contact_default",
		Connector: mockConnector,
		Config: map[string]interface{}{
			"phone_field": "Phone1",
			// No country_code - should default to US
		},
	}

	output, err := h.Execute(context.Background(), input)

	assert.NoError(t, err)
	assert.True(t, output.Success)
	assert.Equal(t, "US", output.ModifiedData["country_code"])
}

func TestPhoneLookup_Execute_CleanedNumber(t *testing.T) {
	h := &PhoneLookup{}
	mockConnector := &mockConnectorForPhoneLookup{
		fieldValues: map[string]interface{}{
			"Phone1": "+1 (555) 123-4567 ext. 890",
		},
	}

	input := helpers.HelperInput{
		ContactID: "contact_clean",
		Connector: mockConnector,
		Config: map[string]interface{}{
			"phone_field": "Phone1",
		},
	}

	output, err := h.Execute(context.Background(), input)

	assert.NoError(t, err)
	assert.True(t, output.Success)
	// Should strip all non-digits
	assert.Equal(t, "15551234567890", output.ModifiedData["number"])
}

func TestPhoneLookup_Execute_ModifiedDataStructure(t *testing.T) {
	h := &PhoneLookup{}
	mockConnector := &mockConnectorForPhoneLookup{
		fieldValues: map[string]interface{}{
			"Phone1": "555-123-4567",
		},
	}

	input := helpers.HelperInput{
		ContactID: "contact_data",
		Connector: mockConnector,
		Config: map[string]interface{}{
			"phone_field": "Phone1",
		},
	}

	output, err := h.Execute(context.Background(), input)

	assert.NoError(t, err)
	require.NotNil(t, output.ModifiedData)
	assert.Contains(t, output.ModifiedData, "number")
	assert.Contains(t, output.ModifiedData, "country_code")
	assert.Contains(t, output.ModifiedData, "raw_number")
	assert.Contains(t, output.ModifiedData, "status")
	assert.Contains(t, output.ModifiedData, "contact_id")
}

func TestPhoneLookup_Execute_Logs(t *testing.T) {
	h := &PhoneLookup{}
	mockConnector := &mockConnectorForPhoneLookup{
		fieldValues: map[string]interface{}{
			"Phone1": "555-1234567",
		},
	}

	input := helpers.HelperInput{
		ContactID: "contact_logs",
		Connector: mockConnector,
		Config: map[string]interface{}{
			"phone_field": "Phone1",
		},
	}

	output, err := h.Execute(context.Background(), input)

	assert.NoError(t, err)
	assert.NotEmpty(t, output.Logs)
	assert.Contains(t, output.Logs[0], "Cleaned phone number")
}
