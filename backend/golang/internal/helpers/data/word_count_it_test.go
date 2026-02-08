package data

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/myfusionhelper/api/internal/connectors"
	"github.com/myfusionhelper/api/internal/helpers"
)

type mockConnectorForWordCount struct {
	fieldValues   map[string]interface{}
	updatedFields map[string]interface{}
}

func (m *mockConnectorForWordCount) GetContactFieldValue(ctx context.Context, contactID, fieldKey string) (interface{}, error) {
	if m.fieldValues == nil {
		return nil, nil
	}
	val, ok := m.fieldValues[fieldKey]
	if !ok {
		return nil, nil
	}
	return val, nil
}

func (m *mockConnectorForWordCount) SetContactFieldValue(ctx context.Context, contactID, fieldKey string, value interface{}) error {
	if m.updatedFields == nil {
		m.updatedFields = make(map[string]interface{})
	}
	m.updatedFields[fieldKey] = value
	return nil
}

func (m *mockConnectorForWordCount) GetContact(ctx context.Context, contactID string) (*connectors.NormalizedContact, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForWordCount) GetContacts(ctx context.Context, opts connectors.QueryOptions) (*connectors.ContactList, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForWordCount) CreateContact(ctx context.Context, input connectors.CreateContactInput) (*connectors.NormalizedContact, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForWordCount) UpdateContact(ctx context.Context, contactID string, updates connectors.UpdateContactInput) (*connectors.NormalizedContact, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForWordCount) DeleteContact(ctx context.Context, contactID string) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorForWordCount) GetTags(ctx context.Context) ([]connectors.Tag, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForWordCount) ApplyTag(ctx context.Context, contactID, tagID string) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorForWordCount) RemoveTag(ctx context.Context, contactID, tagID string) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorForWordCount) GetCustomFields(ctx context.Context) ([]connectors.CustomField, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForWordCount) TriggerAutomation(ctx context.Context, contactID, automationID string) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorForWordCount) AchieveGoal(ctx context.Context, contactID, goalName, integration string) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorForWordCount) TestConnection(ctx context.Context) error {
	return nil
}
func (m *mockConnectorForWordCount) GetMetadata() connectors.ConnectorMetadata {
	return connectors.ConnectorMetadata{}
}
func (m *mockConnectorForWordCount) GetCapabilities() []connectors.Capability {
	return nil
}

func TestWordCountIt_Metadata(t *testing.T) {
	helper := &WordCountIt{}

	if helper.GetName() != "Word Count It" {
		t.Errorf("Expected name 'Word Count It', got '%s'", helper.GetName())
	}
	if helper.GetType() != "word_count_it" {
		t.Errorf("Expected type 'word_count_it', got '%s'", helper.GetType())
	}
	if helper.GetCategory() != "data" {
		t.Errorf("Expected category 'data', got '%s'", helper.GetCategory())
	}
	if !helper.RequiresCRM() {
		t.Error("Expected RequiresCRM to be true")
	}
}

func TestWordCountIt_ValidateConfig_MissingSourceField(t *testing.T) {
	helper := &WordCountIt{}
	config := map[string]interface{}{
		"target_field": "word_count",
		"count_type":   "words",
	}

	err := helper.ValidateConfig(config)
	if err == nil || !strings.Contains(err.Error(), "source_field") {
		t.Errorf("Expected error about source_field, got: %v", err)
	}
}

func TestWordCountIt_ValidateConfig_MissingTargetField(t *testing.T) {
	helper := &WordCountIt{}
	config := map[string]interface{}{
		"source_field": "description",
		"count_type":   "words",
	}

	err := helper.ValidateConfig(config)
	if err == nil || !strings.Contains(err.Error(), "target_field") {
		t.Errorf("Expected error about target_field, got: %v", err)
	}
}

func TestWordCountIt_ValidateConfig_MissingCountType(t *testing.T) {
	helper := &WordCountIt{}
	config := map[string]interface{}{
		"source_field": "description",
		"target_field": "word_count",
	}

	err := helper.ValidateConfig(config)
	if err == nil || !strings.Contains(err.Error(), "count_type") {
		t.Errorf("Expected error about count_type, got: %v", err)
	}
}

func TestWordCountIt_ValidateConfig_InvalidCountType(t *testing.T) {
	helper := &WordCountIt{}
	config := map[string]interface{}{
		"source_field": "description",
		"target_field": "word_count",
		"count_type":   "invalid",
	}

	err := helper.ValidateConfig(config)
	if err == nil || !strings.Contains(err.Error(), "must be") {
		t.Errorf("Expected error about invalid count_type, got: %v", err)
	}
}

func TestWordCountIt_Execute_WordCount(t *testing.T) {
	helper := &WordCountIt{}

	mockConn := &mockConnectorForWordCount{
		fieldValues: map[string]interface{}{
			"description": "This is a test description with seven words",
		},
	}

	input := helpers.HelperInput{
		ContactID: "123",
		Connector: mockConn,
		Config: map[string]interface{}{
			"source_field": "description",
			"target_field": "word_count",
			"count_type":   "words",
		},
	}

	output, err := helper.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if !output.Success {
		t.Error("Expected success=true")
	}

	result := mockConn.updatedFields["word_count"]
	if result != "8" {
		t.Errorf("Expected '8', got: %v", result)
	}
}

func TestWordCountIt_Execute_CharacterCount(t *testing.T) {
	helper := &WordCountIt{}

	mockConn := &mockConnectorForWordCount{
		fieldValues: map[string]interface{}{
			"message": "Hello!",
		},
	}

	input := helpers.HelperInput{
		ContactID: "123",
		Connector: mockConn,
		Config: map[string]interface{}{
			"source_field": "message",
			"target_field": "char_count",
			"count_type":   "characters",
		},
	}

	output, err := helper.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if !output.Success {
		t.Error("Expected success=true")
	}

	result := mockConn.updatedFields["char_count"]
	if result != "6" {
		t.Errorf("Expected '6', got: %v", result)
	}
}

func TestWordCountIt_Execute_EmptyString_Words(t *testing.T) {
	helper := &WordCountIt{}

	mockConn := &mockConnectorForWordCount{
		fieldValues: map[string]interface{}{
			"notes": "",
		},
	}

	input := helpers.HelperInput{
		ContactID: "123",
		Connector: mockConn,
		Config: map[string]interface{}{
			"source_field": "notes",
			"target_field": "word_count",
			"count_type":   "words",
		},
	}

	_, err := helper.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	result := mockConn.updatedFields["word_count"]
	if result != "0" {
		t.Errorf("Expected '0', got: %v", result)
	}
}

func TestWordCountIt_Execute_NilValue(t *testing.T) {
	helper := &WordCountIt{}

	mockConn := &mockConnectorForWordCount{
		fieldValues: map[string]interface{}{
			"bio": nil,
		},
	}

	input := helpers.HelperInput{
		ContactID: "123",
		Connector: mockConn,
		Config: map[string]interface{}{
			"source_field": "bio",
			"target_field": "word_count",
			"count_type":   "words",
		},
	}

	_, err := helper.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	result := mockConn.updatedFields["word_count"]
	if result != "0" {
		t.Errorf("Expected '0' for nil value, got: %v", result)
	}
}

func TestWordCountIt_Execute_MultipleSpaces(t *testing.T) {
	helper := &WordCountIt{}

	mockConn := &mockConnectorForWordCount{
		fieldValues: map[string]interface{}{
			"text": "word1    word2   word3",
		},
	}

	input := helpers.HelperInput{
		ContactID: "123",
		Connector: mockConn,
		Config: map[string]interface{}{
			"source_field": "text",
			"target_field": "word_count",
			"count_type":   "words",
		},
	}

	_, err := helper.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	result := mockConn.updatedFields["word_count"]
	if result != "3" {
		t.Errorf("Expected '3' (Fields handles multiple spaces), got: %v", result)
	}
}

func TestWordCountIt_Execute_UnicodeCharacters(t *testing.T) {
	helper := &WordCountIt{}

	mockConn := &mockConnectorForWordCount{
		fieldValues: map[string]interface{}{
			"message": "Hello 世界",
		},
	}

	input := helpers.HelperInput{
		ContactID: "123",
		Connector: mockConn,
		Config: map[string]interface{}{
			"source_field": "message",
			"target_field": "char_count",
			"count_type":   "characters",
		},
	}

	_, err := helper.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	result := mockConn.updatedFields["char_count"]
	if result != "8" {
		t.Errorf("Expected '8' (rune count includes unicode), got: %v", result)
	}
}

func TestWordCountIt_Execute_ActionLogging(t *testing.T) {
	helper := &WordCountIt{}

	mockConn := &mockConnectorForWordCount{
		fieldValues: map[string]interface{}{
			"description": "Test content",
		},
	}

	input := helpers.HelperInput{
		ContactID: "123",
		Connector: mockConn,
		Config: map[string]interface{}{
			"source_field": "description",
			"target_field": "word_count",
			"count_type":   "words",
		},
	}

	output, err := helper.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(output.Actions) != 1 {
		t.Fatalf("Expected 1 action, got %d", len(output.Actions))
	}

	action := output.Actions[0]
	if action.Type != "field_updated" {
		t.Errorf("Expected action type 'field_updated', got '%s'", action.Type)
	}
	if action.Target != "word_count" {
		t.Errorf("Expected action target 'word_count', got '%s'", action.Target)
	}

	if len(output.Logs) == 0 {
		t.Error("Expected logs to be generated")
	}

	if output.ModifiedData == nil {
		t.Fatal("Expected ModifiedData to be set")
	}
}
