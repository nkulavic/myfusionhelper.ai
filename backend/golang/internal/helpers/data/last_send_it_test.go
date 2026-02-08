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

// Mock connector for last_send_it testing
type mockConnectorForLastSendIt struct {
	fieldValues    map[string]interface{}
	getFieldError  map[string]error
	setFieldError  error
	setFieldCalls  map[string]interface{}
	setFieldCount  int
}

func (m *mockConnectorForLastSendIt) GetContactFieldValue(ctx context.Context, contactID, fieldKey string) (interface{}, error) {
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

func (m *mockConnectorForLastSendIt) SetContactFieldValue(ctx context.Context, contactID, fieldKey string, value interface{}) error {
	m.setFieldCount++
	if m.setFieldCalls == nil {
		m.setFieldCalls = make(map[string]interface{})
	}
	m.setFieldCalls[fieldKey] = value
	return m.setFieldError
}

// Stub implementations for CRMConnector interface
func (m *mockConnectorForLastSendIt) GetContact(ctx context.Context, contactID string) (*connectors.NormalizedContact, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForLastSendIt) GetContacts(ctx context.Context, opts connectors.QueryOptions) (*connectors.ContactList, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForLastSendIt) CreateContact(ctx context.Context, input connectors.CreateContactInput) (*connectors.NormalizedContact, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForLastSendIt) UpdateContact(ctx context.Context, contactID string, updates connectors.UpdateContactInput) (*connectors.NormalizedContact, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForLastSendIt) DeleteContact(ctx context.Context, contactID string) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorForLastSendIt) GetTags(ctx context.Context) ([]connectors.Tag, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForLastSendIt) ApplyTag(ctx context.Context, contactID, tagID string) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorForLastSendIt) RemoveTag(ctx context.Context, contactID, tagID string) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorForLastSendIt) GetCustomFields(ctx context.Context) ([]connectors.CustomField, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForLastSendIt) TriggerAutomation(ctx context.Context, contactID, automationID string) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorForLastSendIt) AchieveGoal(ctx context.Context, contactID, goalName, integration string) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorForLastSendIt) TestConnection(ctx context.Context) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorForLastSendIt) GetMetadata() connectors.ConnectorMetadata {
	return connectors.ConnectorMetadata{}
}
func (m *mockConnectorForLastSendIt) GetCapabilities() []connectors.Capability {
	return nil
}

func TestLastSendIt_Metadata(t *testing.T) {
	h := &LastSendIt{}

	assert.Equal(t, "Last Send It", h.GetName())
	assert.Equal(t, "last_send_it", h.GetType())
	assert.Equal(t, "data", h.GetCategory())
	assert.NotEmpty(t, h.GetDescription())
	assert.True(t, h.RequiresCRM())
	assert.Nil(t, h.SupportedCRMs())
}

func TestLastSendIt_GetConfigSchema(t *testing.T) {
	h := &LastSendIt{}
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

func TestLastSendIt_ValidateConfig_Success(t *testing.T) {
	h := &LastSendIt{}

	tests := []struct {
		name   string
		config map[string]interface{}
	}{
		{
			name: "save_to only",
			config: map[string]interface{}{
				"save_to": "last_send_date",
			},
		},
		{
			name: "with email_field",
			config: map[string]interface{}{
				"save_to":     "last_send_date",
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

func TestLastSendIt_ValidateConfig_Errors(t *testing.T) {
	h := &LastSendIt{}

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

func TestLastSendIt_Execute_EmptyEmail(t *testing.T) {
	h := &LastSendIt{}
	mockConnector := &mockConnectorForLastSendIt{
		fieldValues: map[string]interface{}{
			"Email": "", // empty email
		},
	}

	input := helpers.HelperInput{
		ContactID: "contact_123",
		Connector: mockConnector,
		Config: map[string]interface{}{
			"save_to": "last_send_date",
		},
	}

	output, err := h.Execute(context.Background(), input)

	assert.NoError(t, err)
	assert.True(t, output.Success)
	assert.Contains(t, output.Message, "Email field 'Email' is empty")
	assert.Len(t, output.Actions, 0)
}

func TestLastSendIt_Execute_MissingEmail(t *testing.T) {
	h := &LastSendIt{}
	mockConnector := &mockConnectorForLastSendIt{
		fieldValues: map[string]interface{}{},
		getFieldError: map[string]error{
			"Email": fmt.Errorf("field not found"),
		},
	}

	input := helpers.HelperInput{
		ContactID: "contact_123",
		Connector: mockConnector,
		Config: map[string]interface{}{
			"save_to": "last_send_date",
		},
	}

	output, err := h.Execute(context.Background(), input)

	assert.NoError(t, err)
	assert.True(t, output.Success)
	assert.Contains(t, output.Message, "Email field 'Email' is empty")
}

func TestLastSendIt_Execute_Success_DefaultEmail(t *testing.T) {
	h := &LastSendIt{}
	mockConnector := &mockConnectorForLastSendIt{
		fieldValues: map[string]interface{}{
			"Email": "test@example.com",
			"_email_stats.test@example.com.LastSentDate": "2024-02-08T10:00:00Z",
		},
	}

	input := helpers.HelperInput{
		ContactID: "contact_123",
		Connector: mockConnector,
		Config: map[string]interface{}{
			"save_to": "last_send_date",
		},
	}

	output, err := h.Execute(context.Background(), input)

	assert.NoError(t, err)
	assert.True(t, output.Success)
	assert.Contains(t, output.Message, "Last send date saved")
	assert.Len(t, output.Actions, 1)
	assert.Equal(t, "field_updated", output.Actions[0].Type)
	assert.Equal(t, "last_send_date", output.Actions[0].Target)
	assert.Equal(t, "2024-02-08T10:00:00Z", output.Actions[0].Value)
	assert.Equal(t, 1, mockConnector.setFieldCount)
}

func TestLastSendIt_Execute_Success_CustomEmailField(t *testing.T) {
	h := &LastSendIt{}
	mockConnector := &mockConnectorForLastSendIt{
		fieldValues: map[string]interface{}{
			"CustomEmail": "custom@example.com",
			"_email_stats.custom@example.com.LastSentDate": "2024-02-08T11:00:00Z",
		},
	}

	input := helpers.HelperInput{
		ContactID: "contact_456",
		Connector: mockConnector,
		Config: map[string]interface{}{
			"save_to":     "send_date",
			"email_field": "CustomEmail",
		},
	}

	output, err := h.Execute(context.Background(), input)

	assert.NoError(t, err)
	assert.True(t, output.Success)
	assert.Equal(t, "2024-02-08T11:00:00Z", mockConnector.setFieldCalls["send_date"])
}

func TestLastSendIt_Execute_Email2Normalization(t *testing.T) {
	h := &LastSendIt{}
	mockConnector := &mockConnectorForLastSendIt{
		fieldValues: map[string]interface{}{
			"EmailAddress2": "email2@example.com",
			"_email_stats.email2@example.com.LastSentDate": "2024-02-08T12:00:00Z",
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

func TestLastSendIt_Execute_Email3Normalization(t *testing.T) {
	h := &LastSendIt{}
	mockConnector := &mockConnectorForLastSendIt{
		fieldValues: map[string]interface{}{
			"EmailAddress3": "email3@example.com",
			"_email_stats.email3@example.com.LastSentDate": "2024-02-08T13:00:00Z",
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

func TestLastSendIt_Execute_FallbackToGenericField(t *testing.T) {
	h := &LastSendIt{}
	mockConnector := &mockConnectorForLastSendIt{
		fieldValues: map[string]interface{}{
			"Email":         "test@example.com",
			"LastSentDate": "2024-02-08T14:00:00Z",
		},
		getFieldError: map[string]error{
			"_email_stats.test@example.com.LastSentDate": fmt.Errorf("not supported"),
		},
	}

	input := helpers.HelperInput{
		ContactID: "contact_202",
		Connector: mockConnector,
		Config: map[string]interface{}{
			"save_to": "last_send_date",
		},
	}

	output, err := h.Execute(context.Background(), input)

	assert.NoError(t, err)
	assert.True(t, output.Success)
	assert.Equal(t, "2024-02-08T14:00:00Z", mockConnector.setFieldCalls["last_send_date"])
}

func TestLastSendIt_Execute_NoSendDateFound(t *testing.T) {
	h := &LastSendIt{}
	mockConnector := &mockConnectorForLastSendIt{
		fieldValues: map[string]interface{}{
			"Email": "test@example.com",
		},
		getFieldError: map[string]error{
			"_email_stats.test@example.com.LastSentDate": fmt.Errorf("not found"),
			"LastSentDate": fmt.Errorf("not found"),
		},
	}

	input := helpers.HelperInput{
		ContactID: "contact_303",
		Connector: mockConnector,
		Config: map[string]interface{}{
			"save_to": "last_send_date",
		},
	}

	output, err := h.Execute(context.Background(), input)

	assert.NoError(t, err)
	assert.True(t, output.Success)
	assert.Contains(t, output.Message, "Could not retrieve email send stats")
	assert.Len(t, output.Actions, 0)
}

func TestLastSendIt_Execute_EmptySendDate(t *testing.T) {
	h := &LastSendIt{}
	mockConnector := &mockConnectorForLastSendIt{
		fieldValues: map[string]interface{}{
			"Email":                                        "test@example.com",
			"_email_stats.test@example.com.LastSentDate": "",
		},
	}

	input := helpers.HelperInput{
		ContactID: "contact_404",
		Connector: mockConnector,
		Config: map[string]interface{}{
			"save_to": "last_send_date",
		},
	}

	output, err := h.Execute(context.Background(), input)

	assert.NoError(t, err)
	assert.True(t, output.Success)
	assert.Contains(t, output.Message, "No send date found")
	assert.Len(t, output.Actions, 0)
}

func TestLastSendIt_Execute_SetFieldError(t *testing.T) {
	h := &LastSendIt{}
	mockConnector := &mockConnectorForLastSendIt{
		fieldValues: map[string]interface{}{
			"Email": "test@example.com",
			"_email_stats.test@example.com.LastSentDate": "2024-02-08T15:00:00Z",
		},
		setFieldError: fmt.Errorf("failed to set field"),
	}

	input := helpers.HelperInput{
		ContactID: "contact_505",
		Connector: mockConnector,
		Config: map[string]interface{}{
			"save_to": "last_send_date",
		},
	}

	output, err := h.Execute(context.Background(), input)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to set field")
	assert.Contains(t, output.Message, "Failed to save last send date")
}

func TestLastSendIt_Execute_ModifiedData(t *testing.T) {
	h := &LastSendIt{}
	mockConnector := &mockConnectorForLastSendIt{
		fieldValues: map[string]interface{}{
			"Email": "test@example.com",
			"_email_stats.test@example.com.LastSentDate": "2024-02-08T16:00:00Z",
		},
	}

	input := helpers.HelperInput{
		ContactID: "contact_606",
		Connector: mockConnector,
		Config: map[string]interface{}{
			"save_to": "last_send_date",
		},
	}

	output, err := h.Execute(context.Background(), input)

	assert.NoError(t, err)
	require.NotNil(t, output.ModifiedData)
	assert.Equal(t, "2024-02-08T16:00:00Z", output.ModifiedData["last_send_date"])
}

func TestLastSendIt_Execute_Logs(t *testing.T) {
	h := &LastSendIt{}
	mockConnector := &mockConnectorForLastSendIt{
		fieldValues: map[string]interface{}{
			"Email": "test@example.com",
			"_email_stats.test@example.com.LastSentDate": "2024-02-08T17:00:00Z",
		},
	}

	input := helpers.HelperInput{
		ContactID: "contact_707",
		Connector: mockConnector,
		Config: map[string]interface{}{
			"save_to": "last_send_date",
		},
	}

	output, err := h.Execute(context.Background(), input)

	assert.NoError(t, err)
	assert.NotEmpty(t, output.Logs)
	assert.Contains(t, output.Logs[0], "Last send date")
}
