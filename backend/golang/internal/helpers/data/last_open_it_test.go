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

// Mock connector for last_open_it testing
type mockConnectorForLastOpenIt struct {
	fieldValues    map[string]interface{}
	getFieldError  map[string]error
	setFieldError  error
	setFieldCalls  map[string]interface{}
	setFieldCount  int
}

func (m *mockConnectorForLastOpenIt) GetContactFieldValue(ctx context.Context, contactID, fieldKey string) (interface{}, error) {
	if m.getFieldError != nil {
		if err, ok := m.getFieldError[fieldKey]; ok {
			return nil, err
		}
	}
	if m.fieldValues == nil {
		return nil, fmt.Errorf("field not found")
	}
	val, ok := m.fieldValues[fieldKey]
	if !ok {
		return nil, fmt.Errorf("field '%s' not found", fieldKey)
	}
	return val, nil
}

func (m *mockConnectorForLastOpenIt) SetContactFieldValue(ctx context.Context, contactID, fieldKey string, value interface{}) error {
	m.setFieldCount++
	if m.setFieldCalls == nil {
		m.setFieldCalls = make(map[string]interface{})
	}
	m.setFieldCalls[fieldKey] = value
	return m.setFieldError
}

// Stub implementations for CRMConnector interface
func (m *mockConnectorForLastOpenIt) GetContact(ctx context.Context, contactID string) (*connectors.NormalizedContact, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForLastOpenIt) GetContacts(ctx context.Context, opts connectors.QueryOptions) (*connectors.ContactList, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForLastOpenIt) CreateContact(ctx context.Context, input connectors.CreateContactInput) (*connectors.NormalizedContact, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForLastOpenIt) UpdateContact(ctx context.Context, contactID string, updates connectors.UpdateContactInput) (*connectors.NormalizedContact, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForLastOpenIt) DeleteContact(ctx context.Context, contactID string) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorForLastOpenIt) GetTags(ctx context.Context) ([]connectors.Tag, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForLastOpenIt) ApplyTag(ctx context.Context, contactID, tagID string) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorForLastOpenIt) RemoveTag(ctx context.Context, contactID, tagID string) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorForLastOpenIt) GetCustomFields(ctx context.Context) ([]connectors.CustomField, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForLastOpenIt) TriggerAutomation(ctx context.Context, contactID, automationID string) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorForLastOpenIt) AchieveGoal(ctx context.Context, contactID, goalName, integration string) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorForLastOpenIt) TestConnection(ctx context.Context) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorForLastOpenIt) GetMetadata() connectors.ConnectorMetadata {
	return connectors.ConnectorMetadata{}
}
func (m *mockConnectorForLastOpenIt) GetCapabilities() []connectors.Capability {
	return nil
}

func TestLastOpenIt_Metadata(t *testing.T) {
	h := &LastOpenIt{}

	assert.Equal(t, "Last Open It", h.GetName())
	assert.Equal(t, "last_open_it", h.GetType())
	assert.Equal(t, "data", h.GetCategory())
	assert.NotEmpty(t, h.GetDescription())
	assert.True(t, h.RequiresCRM())
	assert.Nil(t, h.SupportedCRMs())
}

func TestLastOpenIt_GetConfigSchema(t *testing.T) {
	h := &LastOpenIt{}
	schema := h.GetConfigSchema()

	assert.Equal(t, "object", schema["type"])

	props, ok := schema["properties"].(map[string]interface{})
	require.True(t, ok)

	// Verify email_field
	emailField, ok := props["email_field"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "string", emailField["type"])
	assert.Equal(t, "Email", emailField["default"])

	// Verify save_to
	saveTo, ok := props["save_to"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "string", saveTo["type"])

	// Verify required fields
	required, ok := schema["required"].([]string)
	require.True(t, ok)
	assert.Contains(t, required, "save_to")
}

func TestLastOpenIt_ValidateConfig_Success(t *testing.T) {
	h := &LastOpenIt{}

	tests := []struct {
		name   string
		config map[string]interface{}
	}{
		{
			name: "save_to only",
			config: map[string]interface{}{
				"save_to": "last_click_date",
			},
		},
		{
			name: "with email_field",
			config: map[string]interface{}{
				"save_to":     "last_click_date",
				"email_field": "Email",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := h.ValidateConfig(tt.config)
			assert.NoError(t, err)
		})
	}
}

func TestLastOpenIt_ValidateConfig_Errors(t *testing.T) {
	h := &LastOpenIt{}

	tests := []struct {
		name    string
		config  map[string]interface{}
		wantErr string
	}{
		{
			name:    "missing save_to",
			config:  map[string]interface{}{},
			wantErr: "save_to field is required",
		},
		{
			name: "empty save_to",
			config: map[string]interface{}{
				"save_to": "",
			},
			wantErr: "save_to field is required",
		},
		{
			name: "save_to not string",
			config: map[string]interface{}{
				"save_to": 123,
			},
			wantErr: "save_to field is required",
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

func TestLastOpenIt_Execute_EmptyEmail(t *testing.T) {
	h := &LastOpenIt{}
	mockConnector := &mockConnectorForLastOpenIt{
		fieldValues: map[string]interface{}{
			"Email": "", // empty email
		},
	}

	input := helpers.HelperInput{
		ContactID: "contact_123",
		Connector: mockConnector,
		Config: map[string]interface{}{
			"save_to": "last_click_date",
		},
	}

	output, err := h.Execute(context.Background(), input)

	assert.NoError(t, err)
	assert.True(t, output.Success)
	assert.Contains(t, output.Message, "Email field 'Email' is empty")
	assert.Len(t, output.Actions, 0)
}

func TestLastOpenIt_Execute_MissingEmail(t *testing.T) {
	h := &LastOpenIt{}
	mockConnector := &mockConnectorForLastOpenIt{
		fieldValues: map[string]interface{}{},
		getFieldError: map[string]error{
			"Email": fmt.Errorf("field not found"),
		},
	}

	input := helpers.HelperInput{
		ContactID: "contact_123",
		Connector: mockConnector,
		Config: map[string]interface{}{
			"save_to": "last_click_date",
		},
	}

	output, err := h.Execute(context.Background(), input)

	assert.NoError(t, err)
	assert.True(t, output.Success)
	assert.Contains(t, output.Message, "Email field 'Email' is empty")
}

func TestLastOpenIt_Execute_Success_DefaultEmail(t *testing.T) {
	h := &LastOpenIt{}
	mockConnector := &mockConnectorForLastOpenIt{
		fieldValues: map[string]interface{}{
			"Email": "test@example.com",
			"_email_stats.test@example.com.LastOpenDate": "2024-02-08T10:00:00Z",
		},
	}

	input := helpers.HelperInput{
		ContactID: "contact_123",
		Connector: mockConnector,
		Config: map[string]interface{}{
			"save_to": "last_click_date",
		},
	}

	output, err := h.Execute(context.Background(), input)

	assert.NoError(t, err)
	assert.True(t, output.Success)
	assert.Contains(t, output.Message, "Last open date saved")
	assert.Len(t, output.Actions, 1)
	assert.Equal(t, "field_updated", output.Actions[0].Type)
	assert.Equal(t, "last_click_date", output.Actions[0].Target)
	assert.Equal(t, "2024-02-08T10:00:00Z", output.Actions[0].Value)
	assert.Equal(t, 1, mockConnector.setFieldCount)
}

func TestLastOpenIt_Execute_Success_CustomEmailField(t *testing.T) {
	h := &LastOpenIt{}
	mockConnector := &mockConnectorForLastOpenIt{
		fieldValues: map[string]interface{}{
			"CustomEmail": "custom@example.com",
			"_email_stats.custom@example.com.LastOpenDate": "2024-02-08T11:00:00Z",
		},
	}

	input := helpers.HelperInput{
		ContactID: "contact_456",
		Connector: mockConnector,
		Config: map[string]interface{}{
			"save_to":     "click_date",
			"email_field": "CustomEmail",
		},
	}

	output, err := h.Execute(context.Background(), input)

	assert.NoError(t, err)
	assert.True(t, output.Success)
	assert.Equal(t, "2024-02-08T11:00:00Z", mockConnector.setFieldCalls["click_date"])
}

func TestLastOpenIt_Execute_Email2Normalization(t *testing.T) {
	h := &LastOpenIt{}
	mockConnector := &mockConnectorForLastOpenIt{
		fieldValues: map[string]interface{}{
			"EmailAddress2": "email2@example.com",
			"_email_stats.email2@example.com.LastOpenDate": "2024-02-08T12:00:00Z",
		},
	}

	input := helpers.HelperInput{
		ContactID: "contact_789",
		Connector: mockConnector,
		Config: map[string]interface{}{
			"save_to":     "last_click",
			"email_field": "Email2",
		},
	}

	output, err := h.Execute(context.Background(), input)

	assert.NoError(t, err)
	assert.True(t, output.Success)
}

func TestLastOpenIt_Execute_Email3Normalization(t *testing.T) {
	h := &LastOpenIt{}
	mockConnector := &mockConnectorForLastOpenIt{
		fieldValues: map[string]interface{}{
			"EmailAddress3": "email3@example.com",
			"_email_stats.email3@example.com.LastOpenDate": "2024-02-08T13:00:00Z",
		},
	}

	input := helpers.HelperInput{
		ContactID: "contact_101",
		Connector: mockConnector,
		Config: map[string]interface{}{
			"save_to":     "last_click",
			"email_field": "Email3",
		},
	}

	output, err := h.Execute(context.Background(), input)

	assert.NoError(t, err)
	assert.True(t, output.Success)
}

func TestLastOpenIt_Execute_FallbackToGenericField(t *testing.T) {
	h := &LastOpenIt{}
	mockConnector := &mockConnectorForLastOpenIt{
		fieldValues: map[string]interface{}{
			"Email":         "test@example.com",
			"LastOpenDate": "2024-02-08T14:00:00Z",
		},
		getFieldError: map[string]error{
			"_email_stats.test@example.com.LastOpenDate": fmt.Errorf("not supported"),
		},
	}

	input := helpers.HelperInput{
		ContactID: "contact_202",
		Connector: mockConnector,
		Config: map[string]interface{}{
			"save_to": "last_click_date",
		},
	}

	output, err := h.Execute(context.Background(), input)

	assert.NoError(t, err)
	assert.True(t, output.Success)
	assert.Equal(t, "2024-02-08T14:00:00Z", mockConnector.setFieldCalls["last_click_date"])
}

func TestLastOpenIt_Execute_NoOpenDateFound(t *testing.T) {
	h := &LastOpenIt{}
	mockConnector := &mockConnectorForLastOpenIt{
		fieldValues: map[string]interface{}{
			"Email": "test@example.com",
		},
		getFieldError: map[string]error{
			"_email_stats.test@example.com.LastOpenDate": fmt.Errorf("not found"),
			"LastOpenDate": fmt.Errorf("not found"),
		},
	}

	input := helpers.HelperInput{
		ContactID: "contact_303",
		Connector: mockConnector,
		Config: map[string]interface{}{
			"save_to": "last_click_date",
		},
	}

	output, err := h.Execute(context.Background(), input)

	assert.NoError(t, err)
	assert.True(t, output.Success)
	assert.Contains(t, output.Message, "Could not retrieve email open stats")
	assert.Len(t, output.Actions, 0)
}

func TestLastOpenIt_Execute_EmptyOpenDate(t *testing.T) {
	h := &LastOpenIt{}
	mockConnector := &mockConnectorForLastOpenIt{
		fieldValues: map[string]interface{}{
			"Email":                                        "test@example.com",
			"_email_stats.test@example.com.LastOpenDate": "",
		},
	}

	input := helpers.HelperInput{
		ContactID: "contact_404",
		Connector: mockConnector,
		Config: map[string]interface{}{
			"save_to": "last_click_date",
		},
	}

	output, err := h.Execute(context.Background(), input)

	assert.NoError(t, err)
	assert.True(t, output.Success)
	assert.Contains(t, output.Message, "No open date found")
	assert.Len(t, output.Actions, 0)
}

func TestLastOpenIt_Execute_SetFieldError(t *testing.T) {
	h := &LastOpenIt{}
	mockConnector := &mockConnectorForLastOpenIt{
		fieldValues: map[string]interface{}{
			"Email": "test@example.com",
			"_email_stats.test@example.com.LastOpenDate": "2024-02-08T15:00:00Z",
		},
		setFieldError: fmt.Errorf("failed to set field"),
	}

	input := helpers.HelperInput{
		ContactID: "contact_505",
		Connector: mockConnector,
		Config: map[string]interface{}{
			"save_to": "last_click_date",
		},
	}

	output, err := h.Execute(context.Background(), input)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to set field")
	assert.Contains(t, output.Message, "Failed to save last open date")
}

func TestLastOpenIt_Execute_ModifiedData(t *testing.T) {
	h := &LastOpenIt{}
	mockConnector := &mockConnectorForLastOpenIt{
		fieldValues: map[string]interface{}{
			"Email": "test@example.com",
			"_email_stats.test@example.com.LastOpenDate": "2024-02-08T16:00:00Z",
		},
	}

	input := helpers.HelperInput{
		ContactID: "contact_606",
		Connector: mockConnector,
		Config: map[string]interface{}{
			"save_to": "last_click_date",
		},
	}

	output, err := h.Execute(context.Background(), input)

	assert.NoError(t, err)
	require.NotNil(t, output.ModifiedData)
	assert.Equal(t, "2024-02-08T16:00:00Z", output.ModifiedData["last_click_date"])
}

func TestLastOpenIt_Execute_Logs(t *testing.T) {
	h := &LastOpenIt{}
	mockConnector := &mockConnectorForLastOpenIt{
		fieldValues: map[string]interface{}{
			"Email": "test@example.com",
			"_email_stats.test@example.com.LastOpenDate": "2024-02-08T17:00:00Z",
		},
	}

	input := helpers.HelperInput{
		ContactID: "contact_707",
		Connector: mockConnector,
		Config: map[string]interface{}{
			"save_to": "last_click_date",
		},
	}

	output, err := h.Execute(context.Background(), input)

	assert.NoError(t, err)
	assert.NotEmpty(t, output.Logs)
	assert.Contains(t, output.Logs[0], "Last open date")
}
