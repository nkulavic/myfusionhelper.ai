package contact

import (
	"context"
	"fmt"
	"testing"

	"github.com/myfusionhelper/api/internal/connectors"
	"github.com/myfusionhelper/api/internal/helpers"
)

// mockConnectorForMoveIt mocks the CRMConnector interface for testing move_it
type mockConnectorForMoveIt struct {
	fieldValues     map[string]interface{}
	getFieldError   error
	setFieldError   error
	contact         *connectors.NormalizedContact
	getContactError error
}

func (m *mockConnectorForMoveIt) GetContact(ctx context.Context, contactID string) (*connectors.NormalizedContact, error) {
	if m.getContactError != nil {
		return nil, m.getContactError
	}
	return m.contact, nil
}

func (m *mockConnectorForMoveIt) SetContactFieldValue(ctx context.Context, contactID, fieldKey string, value interface{}) error {
	if m.setFieldError != nil {
		return m.setFieldError
	}
	if m.fieldValues == nil {
		m.fieldValues = make(map[string]interface{})
	}
	m.fieldValues[fieldKey] = value
	return nil
}

func (m *mockConnectorForMoveIt) GetContactFieldValue(ctx context.Context, contactID, fieldKey string) (interface{}, error) {
	if m.getFieldError != nil {
		return nil, m.getFieldError
	}
	if m.fieldValues == nil {
		return nil, nil
	}
	return m.fieldValues[fieldKey], nil
}

// Stub implementations
func (m *mockConnectorForMoveIt) GetContacts(ctx context.Context, opts connectors.QueryOptions) (*connectors.ContactList, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *mockConnectorForMoveIt) CreateContact(ctx context.Context, input connectors.CreateContactInput) (*connectors.NormalizedContact, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *mockConnectorForMoveIt) UpdateContact(ctx context.Context, contactID string, updates connectors.UpdateContactInput) (*connectors.NormalizedContact, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *mockConnectorForMoveIt) DeleteContact(ctx context.Context, contactID string) error {
	return fmt.Errorf("not implemented")
}

func (m *mockConnectorForMoveIt) GetTags(ctx context.Context) ([]connectors.Tag, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *mockConnectorForMoveIt) ApplyTag(ctx context.Context, contactID, tagID string) error {
	return fmt.Errorf("not implemented")
}

func (m *mockConnectorForMoveIt) RemoveTag(ctx context.Context, contactID, tagID string) error {
	return fmt.Errorf("not implemented")
}

func (m *mockConnectorForMoveIt) GetCustomFields(ctx context.Context) ([]connectors.CustomField, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *mockConnectorForMoveIt) TriggerAutomation(ctx context.Context, contactID, automationID string) error {
	return fmt.Errorf("not implemented")
}

func (m *mockConnectorForMoveIt) AchieveGoal(ctx context.Context, contactID, goalName, integration string) error {
	return fmt.Errorf("not implemented")
}

func (m *mockConnectorForMoveIt) TestConnection(ctx context.Context) error {
	return fmt.Errorf("not implemented")
}

func (m *mockConnectorForMoveIt) GetMetadata() connectors.ConnectorMetadata {
	return connectors.ConnectorMetadata{PlatformSlug: "mock", PlatformName: "Mock"}
}

func (m *mockConnectorForMoveIt) GetCapabilities() []connectors.Capability {
	return []connectors.Capability{}
}

func TestMoveIt_GetMetadata(t *testing.T) {
	helper := &MoveIt{}

	if helper.GetName() != "Move It" {
		t.Errorf("Expected name 'Move It', got '%s'", helper.GetName())
	}
	if helper.GetType() != "move_it" {
		t.Errorf("Expected type 'move_it', got '%s'", helper.GetType())
	}
	if helper.GetCategory() != "contact" {
		t.Errorf("Expected category 'contact', got '%s'", helper.GetCategory())
	}
	if !helper.RequiresCRM() {
		t.Error("Expected RequiresCRM to be true")
	}
}

func TestMoveIt_ValidateConfig_Success(t *testing.T) {
	helper := &MoveIt{}

	config := map[string]interface{}{
		"source_field": "old_email",
		"target_field": "new_email",
	}

	err := helper.ValidateConfig(config)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
}

func TestMoveIt_ValidateConfig_MissingSourceField(t *testing.T) {
	helper := &MoveIt{}

	config := map[string]interface{}{
		"target_field": "new_email",
	}

	err := helper.ValidateConfig(config)
	if err == nil {
		t.Error("Expected error for missing source_field")
	}
	if err != nil && err.Error() != "source_field is required" {
		t.Errorf("Expected 'source_field is required', got: %v", err)
	}
}

func TestMoveIt_ValidateConfig_EmptySourceField(t *testing.T) {
	helper := &MoveIt{}

	config := map[string]interface{}{
		"source_field": "",
		"target_field": "new_email",
	}

	err := helper.ValidateConfig(config)
	if err == nil {
		t.Error("Expected error for empty source_field")
	}
}

func TestMoveIt_ValidateConfig_MissingTargetField(t *testing.T) {
	helper := &MoveIt{}

	config := map[string]interface{}{
		"source_field": "old_email",
	}

	err := helper.ValidateConfig(config)
	if err == nil {
		t.Error("Expected error for missing target_field")
	}
	if err != nil && err.Error() != "target_field is required" {
		t.Errorf("Expected 'target_field is required', got: %v", err)
	}
}

func TestMoveIt_ValidateConfig_EmptyTargetField(t *testing.T) {
	helper := &MoveIt{}

	config := map[string]interface{}{
		"source_field": "old_email",
		"target_field": "",
	}

	err := helper.ValidateConfig(config)
	if err == nil {
		t.Error("Expected error for empty target_field")
	}
}

func TestMoveIt_Execute_Success(t *testing.T) {
	helper := &MoveIt{}
	mock := &mockConnectorForMoveIt{
		fieldValues: map[string]interface{}{
			"old_email": "user@example.com",
		},
	}

	input := helpers.HelperInput{
		ContactID: "123",
		Config: map[string]interface{}{
			"source_field": "old_email",
			"target_field": "new_email",
		},
		Connector: mock,
	}

	output, err := helper.Execute(context.Background(), input)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if !output.Success {
		t.Error("Expected success to be true")
	}

	// Verify target field was set
	if mock.fieldValues["new_email"] != "user@example.com" {
		t.Errorf("Expected target field to be 'user@example.com', got: %v", mock.fieldValues["new_email"])
	}

	// Verify source field was cleared
	if mock.fieldValues["old_email"] != "" {
		t.Errorf("Expected source field to be cleared, got: %v", mock.fieldValues["old_email"])
	}

	// Verify actions
	if len(output.Actions) != 2 {
		t.Errorf("Expected 2 actions, got %d", len(output.Actions))
	}
}

func TestMoveIt_Execute_EmptySourceField(t *testing.T) {
	helper := &MoveIt{}
	mock := &mockConnectorForMoveIt{
		fieldValues: map[string]interface{}{
			"old_email": "",
		},
	}

	input := helpers.HelperInput{
		ContactID: "123",
		Config: map[string]interface{}{
			"source_field": "old_email",
			"target_field": "new_email",
		},
		Connector: mock,
	}

	output, err := helper.Execute(context.Background(), input)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if !output.Success {
		t.Error("Expected success to be true")
	}
	if output.Message != "Source field 'old_email' is empty, nothing to move" {
		t.Errorf("Unexpected message: %s", output.Message)
	}

	// Verify no fields were set
	if len(mock.fieldValues) > 1 {
		t.Error("Expected no additional fields to be set")
	}
}

func TestMoveIt_Execute_NilSourceField(t *testing.T) {
	helper := &MoveIt{}
	mock := &mockConnectorForMoveIt{
		fieldValues: map[string]interface{}{
			"old_email": nil,
		},
	}

	input := helpers.HelperInput{
		ContactID: "123",
		Config: map[string]interface{}{
			"source_field": "old_email",
			"target_field": "new_email",
		},
		Connector: mock,
	}

	output, err := helper.Execute(context.Background(), input)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if !output.Success {
		t.Error("Expected success to be true")
	}
}

func TestMoveIt_Execute_PreserveTrue_TargetHasValue(t *testing.T) {
	helper := &MoveIt{}
	mock := &mockConnectorForMoveIt{
		fieldValues: map[string]interface{}{
			"old_email": "user@example.com",
			"new_email": "existing@example.com",
		},
	}

	input := helpers.HelperInput{
		ContactID: "123",
		Config: map[string]interface{}{
			"source_field": "old_email",
			"target_field": "new_email",
			"preserve":     true,
		},
		Connector: mock,
	}

	output, err := helper.Execute(context.Background(), input)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if !output.Success {
		t.Error("Expected success to be true")
	}
	if output.Message != "Target field 'new_email' already has a value, skipping (preserve=true)" {
		t.Errorf("Unexpected message: %s", output.Message)
	}

	// Verify target was not overwritten
	if mock.fieldValues["new_email"] != "existing@example.com" {
		t.Errorf("Expected target to remain 'existing@example.com', got: %v", mock.fieldValues["new_email"])
	}

	// Verify source was not cleared
	if mock.fieldValues["old_email"] != "user@example.com" {
		t.Errorf("Expected source to remain 'user@example.com', got: %v", mock.fieldValues["old_email"])
	}
}

func TestMoveIt_Execute_PreserveTrue_TargetEmpty(t *testing.T) {
	helper := &MoveIt{}
	mock := &mockConnectorForMoveIt{
		fieldValues: map[string]interface{}{
			"old_email": "user@example.com",
			"new_email": "",
		},
	}

	input := helpers.HelperInput{
		ContactID: "123",
		Config: map[string]interface{}{
			"source_field": "old_email",
			"target_field": "new_email",
			"preserve":     true,
		},
		Connector: mock,
	}

	output, err := helper.Execute(context.Background(), input)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if !output.Success {
		t.Error("Expected success to be true")
	}

	// Verify move happened even with preserve=true (target was empty)
	if mock.fieldValues["new_email"] != "user@example.com" {
		t.Errorf("Expected target to be 'user@example.com', got: %v", mock.fieldValues["new_email"])
	}
	if mock.fieldValues["old_email"] != "" {
		t.Errorf("Expected source to be cleared, got: %v", mock.fieldValues["old_email"])
	}
}

func TestMoveIt_Execute_GetSourceFieldError(t *testing.T) {
	helper := &MoveIt{}
	mock := &mockConnectorForMoveIt{
		getFieldError: fmt.Errorf("field read error"),
	}

	input := helpers.HelperInput{
		ContactID: "123",
		Config: map[string]interface{}{
			"source_field": "old_email",
			"target_field": "new_email",
		},
		Connector: mock,
	}

	output, err := helper.Execute(context.Background(), input)
	if err == nil {
		t.Error("Expected error for GetContactFieldValue failure")
	}
	if output.Success {
		t.Error("Expected success to be false")
	}
}

func TestMoveIt_Execute_SetTargetFieldError(t *testing.T) {
	helper := &MoveIt{}
	mock := &mockConnectorForMoveIt{
		fieldValues: map[string]interface{}{
			"old_email": "user@example.com",
		},
		setFieldError: fmt.Errorf("field write error"),
	}

	input := helpers.HelperInput{
		ContactID: "123",
		Config: map[string]interface{}{
			"source_field": "old_email",
			"target_field": "new_email",
		},
		Connector: mock,
	}

	output, err := helper.Execute(context.Background(), input)
	if err == nil {
		t.Error("Expected error for SetContactFieldValue failure")
	}
	if output.Success {
		t.Error("Expected success to be false")
	}
}

func TestMoveIt_Execute_ClearSourceFieldError(t *testing.T) {
	helper := &MoveIt{}

	// Create custom mock that fails on second SetContactFieldValue call
	callCount := 0
	mock := &mockConnectorForMoveItCustom{
		fieldValues: map[string]interface{}{
			"old_email": "user@example.com",
		},
		callCount: &callCount,
	}

	input := helpers.HelperInput{
		ContactID: "123",
		Config: map[string]interface{}{
			"source_field": "old_email",
			"target_field": "new_email",
		},
		Connector: mock,
	}

	output, err := helper.Execute(context.Background(), input)
	if err != nil {
		t.Errorf("Expected no error (graceful degradation), got: %v", err)
	}
	// Should still succeed even if source clear fails
	if !output.Success {
		t.Error("Expected success to be true despite clear failure")
	}
	if len(output.Message) == 0 || len(output.Logs) == 0 {
		t.Error("Expected warning message in logs about failed clear")
	}
}

// mockConnectorForMoveItCustom allows custom behavior for SetContactFieldValue
type mockConnectorForMoveItCustom struct {
	fieldValues map[string]interface{}
	callCount   *int
}

func (m *mockConnectorForMoveItCustom) GetContact(ctx context.Context, contactID string) (*connectors.NormalizedContact, error) {
	return nil, nil
}

func (m *mockConnectorForMoveItCustom) SetContactFieldValue(ctx context.Context, contactID, fieldKey string, value interface{}) error {
	*m.callCount++
	if *m.callCount == 2 {
		return fmt.Errorf("failed to clear source")
	}
	if m.fieldValues == nil {
		m.fieldValues = make(map[string]interface{})
	}
	m.fieldValues[fieldKey] = value
	return nil
}

func (m *mockConnectorForMoveItCustom) GetContactFieldValue(ctx context.Context, contactID, fieldKey string) (interface{}, error) {
	if m.fieldValues == nil {
		return nil, nil
	}
	return m.fieldValues[fieldKey], nil
}

func (m *mockConnectorForMoveItCustom) GetContacts(ctx context.Context, opts connectors.QueryOptions) (*connectors.ContactList, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *mockConnectorForMoveItCustom) CreateContact(ctx context.Context, input connectors.CreateContactInput) (*connectors.NormalizedContact, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *mockConnectorForMoveItCustom) UpdateContact(ctx context.Context, contactID string, updates connectors.UpdateContactInput) (*connectors.NormalizedContact, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *mockConnectorForMoveItCustom) DeleteContact(ctx context.Context, contactID string) error {
	return fmt.Errorf("not implemented")
}

func (m *mockConnectorForMoveItCustom) GetTags(ctx context.Context) ([]connectors.Tag, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *mockConnectorForMoveItCustom) ApplyTag(ctx context.Context, contactID, tagID string) error {
	return fmt.Errorf("not implemented")
}

func (m *mockConnectorForMoveItCustom) RemoveTag(ctx context.Context, contactID, tagID string) error {
	return fmt.Errorf("not implemented")
}

func (m *mockConnectorForMoveItCustom) GetCustomFields(ctx context.Context) ([]connectors.CustomField, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *mockConnectorForMoveItCustom) TriggerAutomation(ctx context.Context, contactID, automationID string) error {
	return fmt.Errorf("not implemented")
}

func (m *mockConnectorForMoveItCustom) AchieveGoal(ctx context.Context, contactID, goalName, integration string) error {
	return fmt.Errorf("not implemented")
}

func (m *mockConnectorForMoveItCustom) TestConnection(ctx context.Context) error {
	return fmt.Errorf("not implemented")
}

func (m *mockConnectorForMoveItCustom) GetMetadata() connectors.ConnectorMetadata {
	return connectors.ConnectorMetadata{PlatformSlug: "mock", PlatformName: "Mock"}
}

func (m *mockConnectorForMoveItCustom) GetCapabilities() []connectors.Capability {
	return []connectors.Capability{}
}
