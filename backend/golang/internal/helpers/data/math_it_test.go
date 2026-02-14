package data

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/myfusionhelper/api/internal/connectors"
	"github.com/myfusionhelper/api/internal/helpers"
)

// Mock connector for math_it testing
type mockConnectorForMath struct {
	fieldValues   map[string]interface{}
	updatedFields map[string]interface{}
}

func (m *mockConnectorForMath) GetContactFieldValue(ctx context.Context, contactID, fieldKey string) (interface{}, error) {
	if m.fieldValues == nil {
		return nil, fmt.Errorf("field not found")
	}
	val, ok := m.fieldValues[fieldKey]
	if !ok {
		return nil, fmt.Errorf("field '%s' not found", fieldKey)
	}
	return val, nil
}

func (m *mockConnectorForMath) SetContactFieldValue(ctx context.Context, contactID, fieldKey string, value interface{}) error {
	if m.updatedFields == nil {
		m.updatedFields = make(map[string]interface{})
	}
	m.updatedFields[fieldKey] = value
	return nil
}

// Stub implementations for CRMConnector interface
func (m *mockConnectorForMath) GetContact(ctx context.Context, contactID string) (*connectors.NormalizedContact, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForMath) GetContacts(ctx context.Context, opts connectors.QueryOptions) (*connectors.ContactList, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForMath) CreateContact(ctx context.Context, input connectors.CreateContactInput) (*connectors.NormalizedContact, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForMath) UpdateContact(ctx context.Context, contactID string, updates connectors.UpdateContactInput) (*connectors.NormalizedContact, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForMath) DeleteContact(ctx context.Context, contactID string) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorForMath) GetTags(ctx context.Context) ([]connectors.Tag, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForMath) ApplyTag(ctx context.Context, contactID, tagID string) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorForMath) RemoveTag(ctx context.Context, contactID, tagID string) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorForMath) GetCustomFields(ctx context.Context) ([]connectors.CustomField, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForMath) TriggerAutomation(ctx context.Context, contactID, automationID string) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorForMath) AchieveGoal(ctx context.Context, contactID, goalName, integration string) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorForMath) TestConnection(ctx context.Context) error {
	return nil
}
func (m *mockConnectorForMath) GetMetadata() connectors.ConnectorMetadata {
	return connectors.ConnectorMetadata{}
}
func (m *mockConnectorForMath) GetCapabilities() []connectors.Capability {
	return nil
}

// Test helper metadata
func TestMathIt_Metadata(t *testing.T) {
	helper := &MathIt{}

	if helper.GetName() != "Math It" {
		t.Errorf("Expected name 'Math It', got '%s'", helper.GetName())
	}
	if helper.GetType() != "math_it" {
		t.Errorf("Expected type 'math_it', got '%s'", helper.GetType())
	}
	if helper.GetCategory() != "data" {
		t.Errorf("Expected category 'data', got '%s'", helper.GetCategory())
	}
	if !helper.RequiresCRM() {
		t.Error("Expected RequiresCRM to be true")
	}

	schema := helper.GetConfigSchema()
	if schema == nil {
		t.Fatal("Expected config schema, got nil")
	}
}

// Test validation - missing field
func TestMathIt_ValidateConfig_MissingField(t *testing.T) {
	helper := &MathIt{}
	config := map[string]interface{}{
		"operation": "add",
		"operand":   10,
	}

	err := helper.ValidateConfig(config)
	if err == nil {
		t.Error("Expected validation error for missing field")
	}
	if !strings.Contains(err.Error(), "field") {
		t.Errorf("Expected error about field, got: %v", err)
	}
}

// Test validation - missing operation
func TestMathIt_ValidateConfig_MissingOperation(t *testing.T) {
	helper := &MathIt{}
	config := map[string]interface{}{
		"field":   "score",
		"operand": 10,
	}

	err := helper.ValidateConfig(config)
	if err == nil {
		t.Error("Expected validation error for missing operation")
	}
	if !strings.Contains(err.Error(), "operation") {
		t.Errorf("Expected error about operation, got: %v", err)
	}
}

// Test validation - invalid operation
func TestMathIt_ValidateConfig_InvalidOperation(t *testing.T) {
	helper := &MathIt{}
	config := map[string]interface{}{
		"field":     "score",
		"operation": "power",
		"operand":   2,
	}

	err := helper.ValidateConfig(config)
	if err == nil {
		t.Error("Expected validation error for invalid operation")
	}
	if !strings.Contains(err.Error(), "invalid operation") {
		t.Errorf("Expected error about invalid operation, got: %v", err)
	}
}

// Test validation - missing operand for operations that require it
func TestMathIt_ValidateConfig_MissingOperand(t *testing.T) {
	helper := &MathIt{}
	opsThatNeedOperand := []string{"add", "subtract", "multiply", "divide", "percent"}

	for _, op := range opsThatNeedOperand {
		config := map[string]interface{}{
			"field":     "score",
			"operation": op,
		}

		err := helper.ValidateConfig(config)
		if err == nil {
			t.Errorf("Expected validation error for operation '%s' without operand", op)
		}
		if !strings.Contains(err.Error(), "operand is required") {
			t.Errorf("Expected error about required operand for '%s', got: %v", op, err)
		}
	}
}

// Test validation - valid configs
func TestMathIt_ValidateConfig_Valid(t *testing.T) {
	helper := &MathIt{}

	testCases := []map[string]interface{}{
		{"field": "score", "operation": "add", "operand": 10.0},
		{"field": "score", "operation": "round"},
		{"field": "score", "operation": "ceil"},
		{"field": "score", "operation": "floor"},
		{"field": "score", "operation": "abs"},
	}

	for _, config := range testCases {
		err := helper.ValidateConfig(config)
		if err != nil {
			t.Errorf("Expected no validation error for %v, got: %v", config, err)
		}
	}
}

// Test execution - add
func TestMathIt_Execute_Add(t *testing.T) {
	helper := &MathIt{}

	mockConn := &mockConnectorForMath{
		fieldValues: map[string]interface{}{
			"score": 50.0,
		},
	}

	input := helpers.HelperInput{
		ContactID: "123",
		Connector: mockConn,
		Config: map[string]interface{}{
			"field":     "score",
			"operation": "add",
			"operand":   10.0,
		},
	}

	output, err := helper.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if !output.Success {
		t.Error("Expected success=true")
	}

	if updated := mockConn.updatedFields["score"]; updated != "60" {
		t.Errorf("Expected score '60', got: %v", updated)
	}
}

// Test execution - subtract
func TestMathIt_Execute_Subtract(t *testing.T) {
	helper := &MathIt{}

	mockConn := &mockConnectorForMath{
		fieldValues: map[string]interface{}{
			"score": 100.0,
		},
	}

	input := helpers.HelperInput{
		ContactID: "123",
		Connector: mockConn,
		Config: map[string]interface{}{
			"field":     "score",
			"operation": "subtract",
			"operand":   25.0,
		},
	}

	_, err := helper.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if updated := mockConn.updatedFields["score"]; updated != "75" {
		t.Errorf("Expected score '75', got: %v", updated)
	}
}

// Test execution - multiply
func TestMathIt_Execute_Multiply(t *testing.T) {
	helper := &MathIt{}

	mockConn := &mockConnectorForMath{
		fieldValues: map[string]interface{}{
			"score": 5.0,
		},
	}

	input := helpers.HelperInput{
		ContactID: "123",
		Connector: mockConn,
		Config: map[string]interface{}{
			"field":     "score",
			"operation": "multiply",
			"operand":   3.0,
		},
	}

	_, err := helper.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if updated := mockConn.updatedFields["score"]; updated != "15" {
		t.Errorf("Expected score '15', got: %v", updated)
	}
}

// Test execution - divide
func TestMathIt_Execute_Divide(t *testing.T) {
	helper := &MathIt{}

	mockConn := &mockConnectorForMath{
		fieldValues: map[string]interface{}{
			"score": 100.0,
		},
	}

	input := helpers.HelperInput{
		ContactID: "123",
		Connector: mockConn,
		Config: map[string]interface{}{
			"field":     "score",
			"operation": "divide",
			"operand":   4.0,
		},
	}

	_, err := helper.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if updated := mockConn.updatedFields["score"]; updated != "25" {
		t.Errorf("Expected score '25', got: %v", updated)
	}
}

// Test execution - divide by zero error
func TestMathIt_Execute_DivideByZero(t *testing.T) {
	helper := &MathIt{}

	mockConn := &mockConnectorForMath{
		fieldValues: map[string]interface{}{
			"score": 100.0,
		},
	}

	input := helpers.HelperInput{
		ContactID: "123",
		Connector: mockConn,
		Config: map[string]interface{}{
			"field":     "score",
			"operation": "divide",
			"operand":   0.0,
		},
	}

	output, err := helper.Execute(context.Background(), input)
	if err == nil {
		t.Error("Expected error for divide by zero")
	}

	if !strings.Contains(output.Message, "divide by zero") {
		t.Errorf("Expected error message about divide by zero, got: %s", output.Message)
	}
}

// Test execution - round
func TestMathIt_Execute_Round(t *testing.T) {
	helper := &MathIt{}

	testCases := []struct {
		value          float64
		decimalPlaces  int
		expectedResult string
	}{
		{3.14159, 2, "3.14"},
		{3.14159, 3, "3.142"},
		{3.14159, 0, "3"},
		{2.5, 0, "3"},      // rounds up
		{2.4, 0, "2"},      // rounds down
		{1.555, 2, "1.56"}, // rounds up
	}

	for _, tc := range testCases {
		mockConn := &mockConnectorForMath{
			fieldValues: map[string]interface{}{
				"value": tc.value,
			},
		}

		input := helpers.HelperInput{
			ContactID: "123",
			Connector: mockConn,
			Config: map[string]interface{}{
				"field":          "value",
				"operation":      "round",
				"decimal_places": float64(tc.decimalPlaces),
			},
		}

		_, err := helper.Execute(context.Background(), input)
		if err != nil {
			t.Fatalf("Expected no error for value %v, got: %v", tc.value, err)
		}

		if updated := mockConn.updatedFields["value"]; updated != tc.expectedResult {
			t.Errorf("Round %v to %d decimal places: expected '%s', got '%v'", tc.value, tc.decimalPlaces, tc.expectedResult, updated)
		}
	}
}

// Test execution - ceil
func TestMathIt_Execute_Ceil(t *testing.T) {
	helper := &MathIt{}

	testCases := []struct {
		value          float64
		expectedResult string
	}{
		{3.14, "4"},
		{3.99, "4"},
		{-2.5, "-2"},
		{5.0, "5"},
	}

	for _, tc := range testCases {
		mockConn := &mockConnectorForMath{
			fieldValues: map[string]interface{}{
				"value": tc.value,
			},
		}

		input := helpers.HelperInput{
			ContactID: "123",
			Connector: mockConn,
			Config: map[string]interface{}{
				"field":     "value",
				"operation": "ceil",
			},
		}

		_, err := helper.Execute(context.Background(), input)
		if err != nil {
			t.Fatalf("Expected no error for value %v, got: %v", tc.value, err)
		}

		if updated := mockConn.updatedFields["value"]; updated != tc.expectedResult {
			t.Errorf("Ceil(%v): expected '%s', got '%v'", tc.value, tc.expectedResult, updated)
		}
	}
}

// Test execution - floor
func TestMathIt_Execute_Floor(t *testing.T) {
	helper := &MathIt{}

	testCases := []struct {
		value          float64
		expectedResult string
	}{
		{3.14, "3"},
		{3.99, "3"},
		{-2.5, "-3"},
		{5.0, "5"},
	}

	for _, tc := range testCases {
		mockConn := &mockConnectorForMath{
			fieldValues: map[string]interface{}{
				"value": tc.value,
			},
		}

		input := helpers.HelperInput{
			ContactID: "123",
			Connector: mockConn,
			Config: map[string]interface{}{
				"field":     "value",
				"operation": "floor",
			},
		}

		_, err := helper.Execute(context.Background(), input)
		if err != nil {
			t.Fatalf("Expected no error for value %v, got: %v", tc.value, err)
		}

		if updated := mockConn.updatedFields["value"]; updated != tc.expectedResult {
			t.Errorf("Floor(%v): expected '%s', got '%v'", tc.value, tc.expectedResult, updated)
		}
	}
}

// Test execution - abs
func TestMathIt_Execute_Abs(t *testing.T) {
	helper := &MathIt{}

	testCases := []struct {
		value          float64
		expectedResult string
	}{
		{-5.0, "5"},
		{5.0, "5"},
		{-3.14, "3.14"},
		{0.0, "0"},
	}

	for _, tc := range testCases {
		mockConn := &mockConnectorForMath{
			fieldValues: map[string]interface{}{
				"value": tc.value,
			},
		}

		input := helpers.HelperInput{
			ContactID: "123",
			Connector: mockConn,
			Config: map[string]interface{}{
				"field":     "value",
				"operation": "abs",
			},
		}

		_, err := helper.Execute(context.Background(), input)
		if err != nil {
			t.Fatalf("Expected no error for value %v, got: %v", tc.value, err)
		}

		if updated := mockConn.updatedFields["value"]; updated != tc.expectedResult {
			t.Errorf("Abs(%v): expected '%s', got '%v'", tc.value, tc.expectedResult, updated)
		}
	}
}

// Test execution - percent
func TestMathIt_Execute_Percent(t *testing.T) {
	helper := &MathIt{}

	testCases := []struct {
		value          float64
		percent        float64
		expectedResult string
	}{
		{100.0, 10.0, "10"},   // 10% of 100
		{50.0, 20.0, "10"},    // 20% of 50
		{200.0, 5.0, "10"},    // 5% of 200
		{150.0, 33.33, "50"},  // 33.33% of 150
	}

	for _, tc := range testCases {
		mockConn := &mockConnectorForMath{
			fieldValues: map[string]interface{}{
				"value": tc.value,
			},
		}

		input := helpers.HelperInput{
			ContactID: "123",
			Connector: mockConn,
			Config: map[string]interface{}{
				"field":     "value",
				"operation": "percent",
				"operand":   tc.percent,
			},
		}

		_, err := helper.Execute(context.Background(), input)
		if err != nil {
			t.Fatalf("Expected no error for %v%% of %v, got: %v", tc.percent, tc.value, err)
		}

		if updated := mockConn.updatedFields["value"]; updated != tc.expectedResult {
			t.Errorf("%v%% of %v: expected '%s', got '%v'", tc.percent, tc.value, tc.expectedResult, updated)
		}
	}
}

// Test execution - target_field
func TestMathIt_Execute_TargetField(t *testing.T) {
	helper := &MathIt{}

	mockConn := &mockConnectorForMath{
		fieldValues: map[string]interface{}{
			"price": 100.0,
		},
	}

	input := helpers.HelperInput{
		ContactID: "123",
		Connector: mockConn,
		Config: map[string]interface{}{
			"field":        "price",
			"operation":    "multiply",
			"operand":      1.1,
			"target_field": "discounted_price",
		},
	}

	_, err := helper.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Verify target field was updated (not source field)
	if updated := mockConn.updatedFields["discounted_price"]; updated != "110" {
		t.Errorf("Expected discounted_price '110', got: %v", updated)
	}
	if _, ok := mockConn.updatedFields["price"]; ok {
		t.Error("Source field should not have been updated when target_field is specified")
	}
}

// Test execution - different value types (string, int, float)
func TestMathIt_Execute_DifferentValueTypes(t *testing.T) {
	helper := &MathIt{}

	testCases := []struct {
		name     string
		value    interface{}
		expected string
	}{
		{"string", "50", "60"},
		{"int", 50, "60"},
		{"float64", 50.0, "60"},
	}

	for _, tc := range testCases {
		mockConn := &mockConnectorForMath{
			fieldValues: map[string]interface{}{
				"value": tc.value,
			},
		}

		input := helpers.HelperInput{
			ContactID: "123",
			Connector: mockConn,
			Config: map[string]interface{}{
				"field":     "value",
				"operation": "add",
				"operand":   10.0,
			},
		}

		_, err := helper.Execute(context.Background(), input)
		if err != nil {
			t.Fatalf("Expected no error for type %s, got: %v", tc.name, err)
		}

		if updated := mockConn.updatedFields["value"]; updated != tc.expected {
			t.Errorf("Type %s: expected '%s', got '%v'", tc.name, tc.expected, updated)
		}
	}
}

// Test action logging
func TestMathIt_Execute_ActionLogging(t *testing.T) {
	helper := &MathIt{}

	mockConn := &mockConnectorForMath{
		fieldValues: map[string]interface{}{
			"score": 75.0,
		},
	}

	input := helpers.HelperInput{
		ContactID: "123",
		Connector: mockConn,
		Config: map[string]interface{}{
			"field":     "score",
			"operation": "add",
			"operand":   25.0,
		},
	}

	output, err := helper.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Verify action was logged
	if len(output.Actions) == 0 {
		t.Fatal("Expected actions to be logged")
	}

	action := output.Actions[0]
	if action.Type != "field_updated" {
		t.Errorf("Expected action type 'field_updated', got '%s'", action.Type)
	}
	if action.Target != "score" {
		t.Errorf("Expected action target 'score', got '%s'", action.Target)
	}

	// Verify logs
	if len(output.Logs) == 0 {
		t.Fatal("Expected logs to be generated")
	}
}
