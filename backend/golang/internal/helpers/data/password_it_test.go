package data

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/myfusionhelper/api/internal/connectors"
	"github.com/myfusionhelper/api/internal/helpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Mock connector for password_it testing
type mockConnectorForPasswordIt struct {
	fieldValues   map[string]interface{}
	getFieldError error
	setFieldError error
	setFieldCalls map[string]interface{}
}

func (m *mockConnectorForPasswordIt) GetContactFieldValue(ctx context.Context, contactID, fieldKey string) (interface{}, error) {
	if m.getFieldError != nil {
		return nil, m.getFieldError
	}
	if m.fieldValues == nil {
		return nil, nil
	}
	return m.fieldValues[fieldKey], nil
}

func (m *mockConnectorForPasswordIt) SetContactFieldValue(ctx context.Context, contactID, fieldKey string, value interface{}) error {
	if m.setFieldCalls == nil {
		m.setFieldCalls = make(map[string]interface{})
	}
	m.setFieldCalls[fieldKey] = value
	return m.setFieldError
}

// Stub implementations
func (m *mockConnectorForPasswordIt) GetContact(ctx context.Context, contactID string) (*connectors.NormalizedContact, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForPasswordIt) GetContacts(ctx context.Context, opts connectors.QueryOptions) (*connectors.ContactList, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForPasswordIt) CreateContact(ctx context.Context, input connectors.CreateContactInput) (*connectors.NormalizedContact, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForPasswordIt) UpdateContact(ctx context.Context, contactID string, updates connectors.UpdateContactInput) (*connectors.NormalizedContact, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForPasswordIt) DeleteContact(ctx context.Context, contactID string) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorForPasswordIt) GetTags(ctx context.Context) ([]connectors.Tag, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForPasswordIt) ApplyTag(ctx context.Context, contactID, tagID string) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorForPasswordIt) RemoveTag(ctx context.Context, contactID, tagID string) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorForPasswordIt) GetCustomFields(ctx context.Context) ([]connectors.CustomField, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForPasswordIt) TriggerAutomation(ctx context.Context, contactID, automationID string) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorForPasswordIt) AchieveGoal(ctx context.Context, contactID, goalName, integration string) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorForPasswordIt) TestConnection(ctx context.Context) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorForPasswordIt) GetMetadata() connectors.ConnectorMetadata {
	return connectors.ConnectorMetadata{}
}
func (m *mockConnectorForPasswordIt) GetCapabilities() []connectors.Capability {
	return nil
}

func TestPasswordIt_Metadata(t *testing.T) {
	h := &PasswordIt{}

	assert.Equal(t, "Password It", h.GetName())
	assert.Equal(t, "password_it", h.GetType())
	assert.Equal(t, "data", h.GetCategory())
	assert.NotEmpty(t, h.GetDescription())
	assert.True(t, h.RequiresCRM())
	assert.Nil(t, h.SupportedCRMs())
}

func TestPasswordIt_GetConfigSchema(t *testing.T) {
	h := &PasswordIt{}
	schema := h.GetConfigSchema()

	assert.Equal(t, "object", schema["type"])

	props, ok := schema["properties"].(map[string]interface{})
	require.True(t, ok)

	// Verify target_field
	targetField, ok := props["target_field"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "string", targetField["type"])

	// Verify length
	length, ok := props["length"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "integer", length["type"])
	assert.Equal(t, 12, length["default"])

	// Verify required fields
	required, ok := schema["required"].([]string)
	require.True(t, ok)
	assert.Contains(t, required, "target_field")
}

func TestPasswordIt_ValidateConfig_Success(t *testing.T) {
	h := &PasswordIt{}

	tests := []struct {
		name   string
		config map[string]interface{}
	}{
		{
			name: "target_field only",
			config: map[string]interface{}{
				"target_field": "password",
			},
		},
		{
			name: "with all options",
			config: map[string]interface{}{
				"target_field":    "password",
				"length":          16.0,
				"include_special": true,
				"overwrite":       false,
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

func TestPasswordIt_ValidateConfig_Errors(t *testing.T) {
	h := &PasswordIt{}

	tests := []struct {
		name    string
		config  map[string]interface{}
		wantErr string
	}{
		{
			name:    "missing target_field",
			config:  map[string]interface{}{},
			wantErr: "target_field is required",
		},
		{
			name: "empty target_field",
			config: map[string]interface{}{
				"target_field": "",
			},
			wantErr: "target_field is required",
		},
		{
			name: "target_field not string",
			config: map[string]interface{}{
				"target_field": 123,
			},
			wantErr: "target_field is required",
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

func TestPasswordIt_Execute_Success_DefaultLength(t *testing.T) {
	h := &PasswordIt{}
	mockConnector := &mockConnectorForPasswordIt{
		fieldValues: map[string]interface{}{},
	}

	input := helpers.HelperInput{
		ContactID: "contact_123",
		Connector: mockConnector,
		Config: map[string]interface{}{
			"target_field": "password",
		},
	}

	output, err := h.Execute(context.Background(), input)

	assert.NoError(t, err)
	assert.True(t, output.Success)
	assert.Contains(t, output.Message, "12-character password")

	// Verify password was set
	password, ok := mockConnector.setFieldCalls["password"].(string)
	require.True(t, ok)
	assert.Len(t, password, 12)

	// Verify actions
	assert.Len(t, output.Actions, 1)
	assert.Equal(t, "field_updated", output.Actions[0].Type)
	assert.Equal(t, "password", output.Actions[0].Target)
}

func TestPasswordIt_Execute_Success_CustomLength(t *testing.T) {
	h := &PasswordIt{}
	mockConnector := &mockConnectorForPasswordIt{
		fieldValues: map[string]interface{}{},
	}

	input := helpers.HelperInput{
		ContactID: "contact_456",
		Connector: mockConnector,
		Config: map[string]interface{}{
			"target_field": "password",
			"length":       20.0,
		},
	}

	output, err := h.Execute(context.Background(), input)

	assert.NoError(t, err)
	assert.True(t, output.Success)

	password, ok := mockConnector.setFieldCalls["password"].(string)
	require.True(t, ok)
	assert.Len(t, password, 20)
}

func TestPasswordIt_Execute_Success_NoSpecialChars(t *testing.T) {
	h := &PasswordIt{}
	mockConnector := &mockConnectorForPasswordIt{
		fieldValues: map[string]interface{}{},
	}

	input := helpers.HelperInput{
		ContactID: "contact_789",
		Connector: mockConnector,
		Config: map[string]interface{}{
			"target_field":    "password",
			"include_special": false,
		},
	}

	output, err := h.Execute(context.Background(), input)

	assert.NoError(t, err)
	assert.True(t, output.Success)

	password, ok := mockConnector.setFieldCalls["password"].(string)
	require.True(t, ok)

	// Verify no special characters
	specialChars := "!@#$%^&*()-_=+[]{}|;:,.<>?"
	for _, char := range specialChars {
		assert.False(t, strings.ContainsRune(password, char),
			"Password should not contain special character: %c", char)
	}
}

func TestPasswordIt_Execute_SkipExisting_OverwriteFalse(t *testing.T) {
	h := &PasswordIt{}
	mockConnector := &mockConnectorForPasswordIt{
		fieldValues: map[string]interface{}{
			"password": "existing_password",
		},
	}

	input := helpers.HelperInput{
		ContactID: "contact_101",
		Connector: mockConnector,
		Config: map[string]interface{}{
			"target_field": "password",
			"overwrite":    false,
		},
	}

	output, err := h.Execute(context.Background(), input)

	assert.NoError(t, err)
	assert.True(t, output.Success)
	assert.Contains(t, output.Message, "already has a value")
	assert.Len(t, output.Actions, 0)

	// Verify no new password was set
	assert.Nil(t, mockConnector.setFieldCalls["password"])
}

func TestPasswordIt_Execute_OverwriteExisting(t *testing.T) {
	h := &PasswordIt{}
	mockConnector := &mockConnectorForPasswordIt{
		fieldValues: map[string]interface{}{
			"password": "old_password",
		},
	}

	input := helpers.HelperInput{
		ContactID: "contact_202",
		Connector: mockConnector,
		Config: map[string]interface{}{
			"target_field": "password",
			"overwrite":    true,
		},
	}

	output, err := h.Execute(context.Background(), input)

	assert.NoError(t, err)
	assert.True(t, output.Success)

	// Verify new password was set
	password, ok := mockConnector.setFieldCalls["password"].(string)
	require.True(t, ok)
	assert.NotEqual(t, "old_password", password)
}

func TestPasswordIt_Execute_SetFieldError(t *testing.T) {
	h := &PasswordIt{}
	mockConnector := &mockConnectorForPasswordIt{
		fieldValues:   map[string]interface{}{},
		setFieldError: fmt.Errorf("failed to set field"),
	}

	input := helpers.HelperInput{
		ContactID: "contact_303",
		Connector: mockConnector,
		Config: map[string]interface{}{
			"target_field": "password",
		},
	}

	output, err := h.Execute(context.Background(), input)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to set field")
	assert.Contains(t, output.Message, "Failed to set password field")
}

func TestPasswordIt_Execute_ModifiedData(t *testing.T) {
	h := &PasswordIt{}
	mockConnector := &mockConnectorForPasswordIt{
		fieldValues: map[string]interface{}{},
	}

	input := helpers.HelperInput{
		ContactID: "contact_404",
		Connector: mockConnector,
		Config: map[string]interface{}{
			"target_field": "password",
		},
	}

	output, err := h.Execute(context.Background(), input)

	assert.NoError(t, err)
	require.NotNil(t, output.ModifiedData)
	assert.Contains(t, output.ModifiedData, "password")

	password := output.ModifiedData["password"]
	assert.NotNil(t, password)
}

func TestPasswordIt_Execute_Logs(t *testing.T) {
	h := &PasswordIt{}
	mockConnector := &mockConnectorForPasswordIt{
		fieldValues: map[string]interface{}{},
	}

	input := helpers.HelperInput{
		ContactID: "contact_505",
		Connector: mockConnector,
		Config: map[string]interface{}{
			"target_field": "password",
		},
	}

	output, err := h.Execute(context.Background(), input)

	assert.NoError(t, err)
	assert.NotEmpty(t, output.Logs)
	assert.Contains(t, output.Logs[0], "Generated password")
}

func TestPasswordIt_GeneratePassword_VariousLengths(t *testing.T) {
	lengths := []int{8, 12, 16, 20, 32}

	for _, length := range lengths {
		t.Run(fmt.Sprintf("length_%d", length), func(t *testing.T) {
			password, err := generatePassword(length, true)
			assert.NoError(t, err)
			assert.Len(t, password, length)
		})
	}
}

func TestPasswordIt_GeneratePassword_WithSpecial(t *testing.T) {
	// Generate multiple passwords and verify at least one has a special char
	hasSpecial := false
	for i := 0; i < 10; i++ {
		password, err := generatePassword(20, true)
		assert.NoError(t, err)

		specialChars := "!@#$%^&*()-_=+[]{}|;:,.<>?"
		for _, char := range specialChars {
			if strings.ContainsRune(password, char) {
				hasSpecial = true
				break
			}
		}
		if hasSpecial {
			break
		}
	}
	// With 10 attempts at 20 chars each, extremely likely to get a special char
	assert.True(t, hasSpecial, "Should generate at least one password with special characters")
}

func TestPasswordIt_GeneratePassword_WithoutSpecial(t *testing.T) {
	password, err := generatePassword(50, false)
	assert.NoError(t, err)

	// Verify no special characters
	specialChars := "!@#$%^&*()-_=+[]{}|;:,.<>?"
	for _, char := range specialChars {
		assert.False(t, strings.ContainsRune(password, char),
			"Password should not contain special character: %c", char)
	}
}
