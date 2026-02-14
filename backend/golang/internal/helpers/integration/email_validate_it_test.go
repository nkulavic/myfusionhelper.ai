package integration

import (
	"context"
	"fmt"
	"testing"

	"github.com/myfusionhelper/api/internal/connectors"
	"github.com/myfusionhelper/api/internal/helpers"
)

// mockConnectorForEmailValidateIt implements the CRMConnector interface for testing email_validate_it
type mockConnectorForEmailValidateIt struct {
	fieldValues     map[string]interface{}
	fieldsSet       map[string]interface{}
	goalsAchieved   []string
	getFieldError   error
	setFieldError   error
	achieveGoalError error
}

func (m *mockConnectorForEmailValidateIt) GetContactFieldValue(ctx context.Context, contactID, fieldKey string) (interface{}, error) {
	if m.getFieldError != nil {
		return nil, m.getFieldError
	}
	if m.fieldValues == nil {
		return nil, nil
	}
	return m.fieldValues[fieldKey], nil
}

func (m *mockConnectorForEmailValidateIt) SetContactFieldValue(ctx context.Context, contactID, fieldKey string, value interface{}) error {
	if m.setFieldError != nil {
		return m.setFieldError
	}
	if m.fieldsSet == nil {
		m.fieldsSet = make(map[string]interface{})
	}
	m.fieldsSet[fieldKey] = value
	return nil
}

func (m *mockConnectorForEmailValidateIt) AchieveGoal(ctx context.Context, contactID string, goalName string, integration string) error {
	if m.achieveGoalError != nil {
		return m.achieveGoalError
	}
	if m.goalsAchieved == nil {
		m.goalsAchieved = make([]string, 0)
	}
	m.goalsAchieved = append(m.goalsAchieved, goalName)
	return nil
}

// Stub implementations
func (m *mockConnectorForEmailValidateIt) GetContacts(ctx context.Context, opts connectors.QueryOptions) (*connectors.ContactList, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *mockConnectorForEmailValidateIt) GetContact(ctx context.Context, contactID string) (*connectors.NormalizedContact, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *mockConnectorForEmailValidateIt) CreateContact(ctx context.Context, input connectors.CreateContactInput) (*connectors.NormalizedContact, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *mockConnectorForEmailValidateIt) UpdateContact(ctx context.Context, contactID string, updates connectors.UpdateContactInput) (*connectors.NormalizedContact, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *mockConnectorForEmailValidateIt) DeleteContact(ctx context.Context, contactID string) error {
	return fmt.Errorf("not implemented")
}

func (m *mockConnectorForEmailValidateIt) GetTags(ctx context.Context) ([]connectors.Tag, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *mockConnectorForEmailValidateIt) ApplyTag(ctx context.Context, contactID string, tagID string) error {
	return fmt.Errorf("not implemented")
}

func (m *mockConnectorForEmailValidateIt) RemoveTag(ctx context.Context, contactID string, tagID string) error {
	return fmt.Errorf("not implemented")
}

func (m *mockConnectorForEmailValidateIt) GetCustomFields(ctx context.Context) ([]connectors.CustomField, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *mockConnectorForEmailValidateIt) TriggerAutomation(ctx context.Context, contactID string, automationID string) error {
	return fmt.Errorf("not implemented")
}

func (m *mockConnectorForEmailValidateIt) TestConnection(ctx context.Context) error {
	return fmt.Errorf("not implemented")
}

func (m *mockConnectorForEmailValidateIt) GetMetadata() connectors.ConnectorMetadata {
	return connectors.ConnectorMetadata{
		PlatformSlug: "test",
		PlatformName: "Test Platform",
	}
}

func (m *mockConnectorForEmailValidateIt) GetCapabilities() []connectors.Capability {
	return []connectors.Capability{connectors.CapContacts, connectors.CapCustomFields}
}

// Tests

func TestEmailValidateIt_GetMetadata(t *testing.T) {
	helper := &EmailValidateIt{}

	if helper.GetName() != "Email Validate It" {
		t.Errorf("Expected name 'Email Validate It', got '%s'", helper.GetName())
	}
	if helper.GetType() != "email_validate_it" {
		t.Errorf("Expected type 'email_validate_it', got '%s'", helper.GetType())
	}
	if helper.GetCategory() != "integration" {
		t.Errorf("Expected category 'integration', got '%s'", helper.GetCategory())
	}
	if !helper.RequiresCRM() {
		t.Error("Expected RequiresCRM to be true")
	}
}

func TestEmailValidateIt_GetConfigSchema(t *testing.T) {
	helper := &EmailValidateIt{}
	schema := helper.GetConfigSchema()

	if schema == nil {
		t.Fatal("Schema should not be nil")
	}

	props, ok := schema["properties"].(map[string]interface{})
	if !ok {
		t.Fatal("Schema should have properties")
	}

	if _, ok := props["email_field"]; !ok {
		t.Error("Schema should have email_field property")
	}
	if _, ok := props["result_field"]; !ok {
		t.Error("Schema should have result_field property")
	}
	if _, ok := props["check_mx"]; !ok {
		t.Error("Schema should have check_mx property")
	}
}

func TestEmailValidateIt_ValidateConfig_MissingResultField(t *testing.T) {
	helper := &EmailValidateIt{}

	config := map[string]interface{}{}
	err := helper.ValidateConfig(config)
	if err == nil {
		t.Error("Expected error for missing result_field")
	}
}

func TestEmailValidateIt_ValidateConfig_EmptyResultField(t *testing.T) {
	helper := &EmailValidateIt{}

	config := map[string]interface{}{
		"result_field": "",
	}
	err := helper.ValidateConfig(config)
	if err == nil {
		t.Error("Expected error for empty result_field")
	}
}

func TestEmailValidateIt_ValidateConfig_Valid(t *testing.T) {
	helper := &EmailValidateIt{}

	config := map[string]interface{}{
		"result_field": "email_status",
	}
	err := helper.ValidateConfig(config)
	if err != nil {
		t.Errorf("Expected no error for valid config, got: %v", err)
	}
}

func TestEmailValidateIt_Execute_ValidEmail(t *testing.T) {
	helper := &EmailValidateIt{}
	mockConn := &mockConnectorForEmailValidateIt{
		fieldValues: map[string]interface{}{
			"Email": "test@gmail.com",
		},
	}
	ctx := context.Background()

	input := helpers.HelperInput{
		ContactID: "contact-123",
		Config: map[string]interface{}{
			"result_field": "email_status",
			"check_mx":     false, // Disable MX check for unit test
		},
		Connector: mockConn,
	}

	output, err := helper.Execute(ctx, input)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if !output.Success {
		t.Error("Expected success to be true")
	}

	if mockConn.fieldsSet["email_status"] != "valid" {
		t.Errorf("Expected email_status to be 'valid', got: %v", mockConn.fieldsSet["email_status"])
	}

	if output.ModifiedData["status"] != "valid" {
		t.Errorf("Expected status to be 'valid', got: %v", output.ModifiedData["status"])
	}
}

func TestEmailValidateIt_Execute_InvalidEmailFormat(t *testing.T) {
	helper := &EmailValidateIt{}
	mockConn := &mockConnectorForEmailValidateIt{
		fieldValues: map[string]interface{}{
			"Email": "not-an-email",
		},
	}
	ctx := context.Background()

	input := helpers.HelperInput{
		ContactID: "contact-123",
		Config: map[string]interface{}{
			"result_field": "email_status",
		},
		Connector: mockConn,
	}

	output, err := helper.Execute(ctx, input)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if !output.Success {
		t.Error("Expected success to be true")
	}

	if mockConn.fieldsSet["email_status"] != "invalid" {
		t.Errorf("Expected email_status to be 'invalid', got: %v", mockConn.fieldsSet["email_status"])
	}

	if output.ModifiedData["status"] != "invalid" {
		t.Errorf("Expected status to be 'invalid', got: %v", output.ModifiedData["status"])
	}

	if output.ModifiedData["reason"] != "format" {
		t.Errorf("Expected reason to be 'format', got: %v", output.ModifiedData["reason"])
	}
}

func TestEmailValidateIt_Execute_EmptyEmail(t *testing.T) {
	helper := &EmailValidateIt{}
	mockConn := &mockConnectorForEmailValidateIt{
		fieldValues: map[string]interface{}{
			"Email": "",
		},
	}
	ctx := context.Background()

	input := helpers.HelperInput{
		ContactID: "contact-123",
		Config: map[string]interface{}{
			"result_field": "email_status",
		},
		Connector: mockConn,
	}

	output, err := helper.Execute(ctx, input)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if !output.Success {
		t.Error("Expected success to be true")
	}

	if mockConn.fieldsSet["email_status"] != "invalid" {
		t.Errorf("Expected email_status to be 'invalid', got: %v", mockConn.fieldsSet["email_status"])
	}

	if output.ModifiedData["reason"] != "empty" {
		t.Errorf("Expected reason to be 'empty', got: %v", output.ModifiedData["reason"])
	}
}

func TestEmailValidateIt_Execute_CustomEmailField(t *testing.T) {
	helper := &EmailValidateIt{}
	mockConn := &mockConnectorForEmailValidateIt{
		fieldValues: map[string]interface{}{
			"work_email": "work@company.com",
		},
	}
	ctx := context.Background()

	input := helpers.HelperInput{
		ContactID: "contact-123",
		Config: map[string]interface{}{
			"email_field":  "work_email",
			"result_field": "email_status",
			"check_mx":     false,
		},
		Connector: mockConn,
	}

	output, err := helper.Execute(ctx, input)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if !output.Success {
		t.Error("Expected success to be true")
	}

	if output.ModifiedData["email"] != "work@company.com" {
		t.Errorf("Expected email to be 'work@company.com', got: %v", output.ModifiedData["email"])
	}
}

func TestEmailValidateIt_Execute_ValidGoalAchieved(t *testing.T) {
	helper := &EmailValidateIt{}
	mockConn := &mockConnectorForEmailValidateIt{
		fieldValues: map[string]interface{}{
			"Email": "valid@example.com",
		},
	}
	ctx := context.Background()

	input := helpers.HelperInput{
		ContactID: "contact-123",
		Config: map[string]interface{}{
			"result_field": "email_status",
			"valid_goal":   "Email Valid",
			"check_mx":     false,
		},
		Connector: mockConn,
	}

	output, err := helper.Execute(ctx, input)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(mockConn.goalsAchieved) == 0 {
		t.Fatal("Expected goal to be achieved")
	}

	if mockConn.goalsAchieved[0] != "Email Valid" {
		t.Errorf("Expected goal 'Email Valid' to be achieved, got: %s", mockConn.goalsAchieved[0])
	}

	// Check for goal_achieved action
	goalActionFound := false
	for _, action := range output.Actions {
		if action.Type == "goal_achieved" {
			goalActionFound = true
			break
		}
	}
	if !goalActionFound {
		t.Error("Expected goal_achieved action to be recorded")
	}
}

func TestEmailValidateIt_Execute_InvalidGoalAchieved(t *testing.T) {
	helper := &EmailValidateIt{}
	mockConn := &mockConnectorForEmailValidateIt{
		fieldValues: map[string]interface{}{
			"Email": "invalid-email",
		},
	}
	ctx := context.Background()

	input := helpers.HelperInput{
		ContactID: "contact-123",
		Config: map[string]interface{}{
			"result_field": "email_status",
			"invalid_goal": "Email Invalid",
		},
		Connector: mockConn,
	}

	_, err := helper.Execute(ctx, input)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(mockConn.goalsAchieved) == 0 {
		t.Fatal("Expected goal to be achieved")
	}

	if mockConn.goalsAchieved[0] != "Email Invalid" {
		t.Errorf("Expected goal 'Email Invalid' to be achieved, got: %s", mockConn.goalsAchieved[0])
	}
}

func TestEmailValidateIt_Execute_GetFieldError(t *testing.T) {
	helper := &EmailValidateIt{}
	mockConn := &mockConnectorForEmailValidateIt{
		getFieldError: fmt.Errorf("field read error"),
	}
	ctx := context.Background()

	input := helpers.HelperInput{
		ContactID: "contact-123",
		Config: map[string]interface{}{
			"result_field": "email_status",
		},
		Connector: mockConn,
	}

	_, err := helper.Execute(ctx, input)
	if err != nil {
		t.Fatalf("Expected no error (graceful handling), got: %v", err)
	}

	// Should treat as invalid when field read fails
	if mockConn.fieldsSet["email_status"] != "invalid" {
		t.Errorf("Expected email_status to be 'invalid', got: %v", mockConn.fieldsSet["email_status"])
	}
}

func TestEmailValidateIt_Execute_ActionsRecorded(t *testing.T) {
	helper := &EmailValidateIt{}
	mockConn := &mockConnectorForEmailValidateIt{
		fieldValues: map[string]interface{}{
			"Email": "test@example.com",
		},
	}
	ctx := context.Background()

	input := helpers.HelperInput{
		ContactID: "contact-123",
		Config: map[string]interface{}{
			"result_field": "email_status",
			"check_mx":     false,
		},
		Connector: mockConn,
	}

	output, err := helper.Execute(ctx, input)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Should have at least field_updated action
	fieldUpdateFound := false
	for _, action := range output.Actions {
		if action.Type == "field_updated" {
			fieldUpdateFound = true
			if action.Target != "email_status" {
				t.Errorf("Expected field_updated target to be 'email_status', got: %s", action.Target)
			}
			break
		}
	}

	if !fieldUpdateFound {
		t.Error("Expected field_updated action to be recorded")
	}
}

func TestEmailValidateIt_Execute_LogsRecorded(t *testing.T) {
	helper := &EmailValidateIt{}
	mockConn := &mockConnectorForEmailValidateIt{
		fieldValues: map[string]interface{}{
			"Email": "test@example.com",
		},
	}
	ctx := context.Background()

	input := helpers.HelperInput{
		ContactID: "contact-123",
		Config: map[string]interface{}{
			"result_field": "email_status",
			"check_mx":     false,
		},
		Connector: mockConn,
	}

	output, err := helper.Execute(ctx, input)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(output.Logs) == 0 {
		t.Error("Expected logs to be recorded")
	}
}

func TestEmailValidateIt_Execute_CheckMXDefault(t *testing.T) {
	helper := &EmailValidateIt{}
	mockConn := &mockConnectorForEmailValidateIt{
		fieldValues: map[string]interface{}{
			"Email": "test@gmail.com",
		},
	}
	ctx := context.Background()

	input := helpers.HelperInput{
		ContactID: "contact-123",
		Config: map[string]interface{}{
			"result_field": "email_status",
			// check_mx not specified, should default to true
		},
		Connector: mockConn,
	}

	output, err := helper.Execute(ctx, input)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Should perform MX check by default
	if checkMX, ok := output.ModifiedData["check_mx"].(bool); ok {
		if !checkMX {
			t.Error("Expected check_mx to default to true")
		}
	}
}

func TestEmailValidateIt_Execute_ValidEmailMultipleFormats(t *testing.T) {
	testCases := []string{
		"test@example.com",
		"user.name@example.com",
		"user+tag@example.co.uk",
		"user_name@sub.example.com",
		"123@example.com",
	}

	for _, email := range testCases {
		t.Run(email, func(t *testing.T) {
			helper := &EmailValidateIt{}
			mockConn := &mockConnectorForEmailValidateIt{
				fieldValues: map[string]interface{}{
					"Email": email,
				},
			}
			ctx := context.Background()

			input := helpers.HelperInput{
				ContactID: "contact-123",
				Config: map[string]interface{}{
					"result_field": "email_status",
					"check_mx":     false,
				},
				Connector: mockConn,
			}

			_, err := helper.Execute(ctx, input)
			if err != nil {
				t.Fatalf("Expected no error for email '%s', got: %v", email, err)
			}

			if mockConn.fieldsSet["email_status"] != "valid" {
				t.Errorf("Expected email '%s' to be valid, got: %v", email, mockConn.fieldsSet["email_status"])
			}
		})
	}
}

func TestEmailValidateIt_Execute_InvalidEmailMultipleFormats(t *testing.T) {
	testCases := []string{
		"notanemail",
		"@example.com",
		"user@",
		"user name@example.com",
		"user@example",
	}

	for _, email := range testCases {
		t.Run(email, func(t *testing.T) {
			helper := &EmailValidateIt{}
			mockConn := &mockConnectorForEmailValidateIt{
				fieldValues: map[string]interface{}{
					"Email": email,
				},
			}
			ctx := context.Background()

			input := helpers.HelperInput{
				ContactID: "contact-123",
				Config: map[string]interface{}{
					"result_field": "email_status",
				},
				Connector: mockConn,
			}

			_, err := helper.Execute(ctx, input)
			if err != nil {
				t.Fatalf("Expected no error for email '%s', got: %v", email, err)
			}

			if mockConn.fieldsSet["email_status"] != "invalid" {
				t.Errorf("Expected email '%s' to be invalid, got: %v", email, mockConn.fieldsSet["email_status"])
			}
		})
	}
}
