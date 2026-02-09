package data

import (
	"context"
	"fmt"
	"math"
	"testing"

	"github.com/myfusionhelper/api/internal/connectors"
	"github.com/myfusionhelper/api/internal/helpers"
)

// mockConnectorAdvanceMath is a mock CRM connector for testing advanced math operations
type mockConnectorAdvanceMath struct {
	fields map[string]interface{}
}

func (m *mockConnectorAdvanceMath) GetContactFieldValue(ctx context.Context, contactID, fieldKey string) (interface{}, error) {
	if val, ok := m.fields[fieldKey]; ok {
		return val, nil
	}
	return nil, fmt.Errorf("field not found: %s", fieldKey)
}

func (m *mockConnectorAdvanceMath) SetContactFieldValue(ctx context.Context, contactID, fieldKey string, value interface{}) error {
	m.fields[fieldKey] = value
	return nil
}

// Stub methods for CRMConnector interface
func (m *mockConnectorAdvanceMath) GetContact(ctx context.Context, contactID string) (*connectors.NormalizedContact, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorAdvanceMath) GetContacts(ctx context.Context, opts connectors.QueryOptions) (*connectors.ContactList, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorAdvanceMath) CreateContact(ctx context.Context, input connectors.CreateContactInput) (*connectors.NormalizedContact, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorAdvanceMath) UpdateContact(ctx context.Context, contactID string, updates connectors.UpdateContactInput) (*connectors.NormalizedContact, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorAdvanceMath) DeleteContact(ctx context.Context, contactID string) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorAdvanceMath) GetTags(ctx context.Context) ([]connectors.Tag, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorAdvanceMath) ApplyTag(ctx context.Context, contactID, tagID string) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorAdvanceMath) RemoveTag(ctx context.Context, contactID, tagID string) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorAdvanceMath) GetCustomFields(ctx context.Context) ([]connectors.CustomField, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorAdvanceMath) TriggerAutomation(ctx context.Context, contactID, automationID string) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorAdvanceMath) AchieveGoal(ctx context.Context, contactID, goalName, integration string) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorAdvanceMath) TestConnection(ctx context.Context) error {
	return nil
}
func (m *mockConnectorAdvanceMath) GetMetadata() connectors.ConnectorMetadata {
	return connectors.ConnectorMetadata{}
}
func (m *mockConnectorAdvanceMath) GetCapabilities() []connectors.Capability {
	return nil
}

func TestAdvanceMath_ValidateConfig(t *testing.T) {
	helper := &AdvanceMath{}

	tests := []struct {
		name    string
		config  map[string]interface{}
		wantErr bool
	}{
		{
			name: "valid sqrt operation",
			config: map[string]interface{}{
				"operation":    "sqrt",
				"source_field": "amount",
				"target_field": "sqrt_amount",
			},
			wantErr: false,
		},
		{
			name: "valid power operation with operand",
			config: map[string]interface{}{
				"operation":    "power",
				"source_field": "base",
				"target_field": "result",
				"operand":      2.0,
			},
			wantErr: false,
		},
		{
			name: "valid min operation with second_field",
			config: map[string]interface{}{
				"operation":    "min",
				"source_field": "field1",
				"second_field": "field2",
				"target_field": "min_value",
			},
			wantErr: false,
		},
		{
			name: "missing operation",
			config: map[string]interface{}{
				"source_field": "amount",
				"target_field": "result",
			},
			wantErr: true,
		},
		{
			name: "invalid operation",
			config: map[string]interface{}{
				"operation":    "invalid_op",
				"source_field": "amount",
				"target_field": "result",
			},
			wantErr: true,
		},
		{
			name: "missing source_field",
			config: map[string]interface{}{
				"operation":    "sqrt",
				"target_field": "result",
			},
			wantErr: true,
		},
		{
			name: "missing target_field",
			config: map[string]interface{}{
				"operation":    "sqrt",
				"source_field": "amount",
			},
			wantErr: true,
		},
		{
			name: "power without operand or second_field",
			config: map[string]interface{}{
				"operation":    "power",
				"source_field": "amount",
				"target_field": "result",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := helper.ValidateConfig(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAdvanceMath_Execute_Sqrt(t *testing.T) {
	helper := &AdvanceMath{}
	connector := &mockConnectorAdvanceMath{
		fields: map[string]interface{}{
			"amount": 16.0,
		},
	}

	input := helpers.HelperInput{
		ContactID: "contact_123",
		Connector: connector,
		Config: map[string]interface{}{
			"operation":    "sqrt",
			"source_field": "amount",
			"target_field": "sqrt_amount",
		},
	}

	output, err := helper.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if !output.Success {
		t.Errorf("Expected Success = true, got false")
	}

	if connector.fields["sqrt_amount"] != 4.0 {
		t.Errorf("Expected sqrt_amount = 4.0, got %v", connector.fields["sqrt_amount"])
	}
}

func TestAdvanceMath_Execute_Abs(t *testing.T) {
	helper := &AdvanceMath{}
	connector := &mockConnectorAdvanceMath{
		fields: map[string]interface{}{
			"balance": -42.5,
		},
	}

	input := helpers.HelperInput{
		ContactID: "contact_123",
		Connector: connector,
		Config: map[string]interface{}{
			"operation":    "abs",
			"source_field": "balance",
			"target_field": "abs_balance",
		},
	}

	output, err := helper.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if !output.Success {
		t.Errorf("Expected Success = true, got false")
	}

	if connector.fields["abs_balance"] != 42.5 {
		t.Errorf("Expected abs_balance = 42.5, got %v", connector.fields["abs_balance"])
	}
}

func TestAdvanceMath_Execute_Round(t *testing.T) {
	helper := &AdvanceMath{}
	connector := &mockConnectorAdvanceMath{
		fields: map[string]interface{}{
			"price": 19.67,
		},
	}

	input := helpers.HelperInput{
		ContactID: "contact_123",
		Connector: connector,
		Config: map[string]interface{}{
			"operation":    "round",
			"source_field": "price",
			"target_field": "rounded_price",
		},
	}

	output, err := helper.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if !output.Success {
		t.Errorf("Expected Success = true, got false")
	}

	if connector.fields["rounded_price"] != 20.0 {
		t.Errorf("Expected rounded_price = 20.0, got %v", connector.fields["rounded_price"])
	}
}

func TestAdvanceMath_Execute_Ceil(t *testing.T) {
	helper := &AdvanceMath{}
	connector := &mockConnectorAdvanceMath{
		fields: map[string]interface{}{
			"value": 3.2,
		},
	}

	input := helpers.HelperInput{
		ContactID: "contact_123",
		Connector: connector,
		Config: map[string]interface{}{
			"operation":    "ceil",
			"source_field": "value",
			"target_field": "ceil_value",
		},
	}

	output, err := helper.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if !output.Success {
		t.Errorf("Expected Success = true, got false")
	}

	if connector.fields["ceil_value"] != 4.0 {
		t.Errorf("Expected ceil_value = 4.0, got %v", connector.fields["ceil_value"])
	}
}

func TestAdvanceMath_Execute_Floor(t *testing.T) {
	helper := &AdvanceMath{}
	connector := &mockConnectorAdvanceMath{
		fields: map[string]interface{}{
			"value": 3.9,
		},
	}

	input := helpers.HelperInput{
		ContactID: "contact_123",
		Connector: connector,
		Config: map[string]interface{}{
			"operation":    "floor",
			"source_field": "value",
			"target_field": "floor_value",
		},
	}

	output, err := helper.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if !output.Success {
		t.Errorf("Expected Success = true, got false")
	}

	if connector.fields["floor_value"] != 3.0 {
		t.Errorf("Expected floor_value = 3.0, got %v", connector.fields["floor_value"])
	}
}

func TestAdvanceMath_Execute_Power(t *testing.T) {
	helper := &AdvanceMath{}
	connector := &mockConnectorAdvanceMath{
		fields: map[string]interface{}{
			"base": 2.0,
		},
	}

	input := helpers.HelperInput{
		ContactID: "contact_123",
		Connector: connector,
		Config: map[string]interface{}{
			"operation":    "power",
			"source_field": "base",
			"operand":      3.0,
			"target_field": "result",
		},
	}

	output, err := helper.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if !output.Success {
		t.Errorf("Expected Success = true, got false")
	}

	if connector.fields["result"] != 8.0 {
		t.Errorf("Expected result = 8.0, got %v", connector.fields["result"])
	}
}

func TestAdvanceMath_Execute_Min(t *testing.T) {
	helper := &AdvanceMath{}
	connector := &mockConnectorAdvanceMath{
		fields: map[string]interface{}{
			"field1": 10.0,
			"field2": 5.0,
		},
	}

	input := helpers.HelperInput{
		ContactID: "contact_123",
		Connector: connector,
		Config: map[string]interface{}{
			"operation":    "min",
			"source_field": "field1",
			"second_field": "field2",
			"target_field": "min_value",
		},
	}

	output, err := helper.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if !output.Success {
		t.Errorf("Expected Success = true, got false")
	}

	if connector.fields["min_value"] != 5.0 {
		t.Errorf("Expected min_value = 5.0, got %v", connector.fields["min_value"])
	}
}

func TestAdvanceMath_Execute_Max(t *testing.T) {
	helper := &AdvanceMath{}
	connector := &mockConnectorAdvanceMath{
		fields: map[string]interface{}{
			"field1": 10.0,
			"field2": 5.0,
		},
	}

	input := helpers.HelperInput{
		ContactID: "contact_123",
		Connector: connector,
		Config: map[string]interface{}{
			"operation":    "max",
			"source_field": "field1",
			"second_field": "field2",
			"target_field": "max_value",
		},
	}

	output, err := helper.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if !output.Success {
		t.Errorf("Expected Success = true, got false")
	}

	if connector.fields["max_value"] != 10.0 {
		t.Errorf("Expected max_value = 10.0, got %v", connector.fields["max_value"])
	}
}

func TestAdvanceMath_Execute_MinWithOperand(t *testing.T) {
	helper := &AdvanceMath{}
	connector := &mockConnectorAdvanceMath{
		fields: map[string]interface{}{
			"value": 15.0,
		},
	}

	input := helpers.HelperInput{
		ContactID: "contact_123",
		Connector: connector,
		Config: map[string]interface{}{
			"operation":    "min",
			"source_field": "value",
			"operand":      10.0,
			"target_field": "min_result",
		},
	}

	output, err := helper.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if !output.Success {
		t.Errorf("Expected Success = true, got false")
	}

	if connector.fields["min_result"] != 10.0 {
		t.Errorf("Expected min_result = 10.0, got %v", connector.fields["min_result"])
	}
}

func TestAdvanceMath_Execute_NegativeSqrt(t *testing.T) {
	helper := &AdvanceMath{}
	connector := &mockConnectorAdvanceMath{
		fields: map[string]interface{}{
			"value": -4.0,
		},
	}

	input := helpers.HelperInput{
		ContactID: "contact_123",
		Connector: connector,
		Config: map[string]interface{}{
			"operation":    "sqrt",
			"source_field": "value",
			"target_field": "result",
		},
	}

	output, err := helper.Execute(context.Background(), input)
	if err == nil {
		t.Fatalf("Expected error for negative sqrt, got nil")
	}

	if output.Success {
		t.Errorf("Expected Success = false for negative sqrt")
	}
}

func TestAdvanceMath_Execute_StringConversion(t *testing.T) {
	helper := &AdvanceMath{}
	connector := &mockConnectorAdvanceMath{
		fields: map[string]interface{}{
			"amount": "25",
		},
	}

	input := helpers.HelperInput{
		ContactID: "contact_123",
		Connector: connector,
		Config: map[string]interface{}{
			"operation":    "sqrt",
			"source_field": "amount",
			"target_field": "result",
		},
	}

	output, err := helper.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if !output.Success {
		t.Errorf("Expected Success = true, got false")
	}

	if connector.fields["result"] != 5.0 {
		t.Errorf("Expected result = 5.0, got %v", connector.fields["result"])
	}
}

func TestAdvanceMath_GetMetadata(t *testing.T) {
	helper := &AdvanceMath{}

	if helper.GetName() != "Advance Math" {
		t.Errorf("Expected name 'Advance Math', got '%s'", helper.GetName())
	}

	if helper.GetType() != "advance_math" {
		t.Errorf("Expected type 'advance_math', got '%s'", helper.GetType())
	}

	if helper.GetCategory() != "data" {
		t.Errorf("Expected category 'data', got '%s'", helper.GetCategory())
	}

	if !helper.RequiresCRM() {
		t.Errorf("Expected RequiresCRM = true")
	}
}

func TestAdvanceMath_ConfigSchema(t *testing.T) {
	helper := &AdvanceMath{}
	schema := helper.GetConfigSchema()

	if schema == nil {
		t.Fatal("Expected non-nil schema")
	}

	// Verify schema has required properties
	props, ok := schema["properties"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected properties in schema")
	}

	requiredFields := []string{"operation", "source_field", "target_field", "operand", "second_field"}
	for _, field := range requiredFields {
		if _, ok := props[field]; !ok {
			t.Errorf("Expected field '%s' in schema properties", field)
		}
	}

	// Verify operation enum
	opProp, ok := props["operation"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected operation property")
	}

	enum, ok := opProp["enum"].([]string)
	if !ok {
		t.Fatal("Expected enum in operation property")
	}

	expectedOps := []string{"power", "sqrt", "abs", "round", "ceil", "floor", "min", "max"}
	if len(enum) != len(expectedOps) {
		t.Errorf("Expected %d operations in enum, got %d", len(expectedOps), len(enum))
	}
}

func TestToFloat64Advanced(t *testing.T) {
	tests := []struct {
		name    string
		input   interface{}
		want    float64
		wantErr bool
	}{
		{
			name:    "float64",
			input:   42.5,
			want:    42.5,
			wantErr: false,
		},
		{
			name:    "float32",
			input:   float32(10.5),
			want:    10.5,
			wantErr: false,
		},
		{
			name:    "int",
			input:   100,
			want:    100.0,
			wantErr: false,
		},
		{
			name:    "int64",
			input:   int64(200),
			want:    200.0,
			wantErr: false,
		},
		{
			name:    "string number",
			input:   "15.75",
			want:    15.75,
			wantErr: false,
		},
		{
			name:    "invalid string",
			input:   "not a number",
			want:    0,
			wantErr: true,
		},
		{
			name:    "bool (unsupported)",
			input:   true,
			want:    0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := toFloat64Advanced(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("toFloat64Advanced() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && math.Abs(got-tt.want) > 0.0001 {
				t.Errorf("toFloat64Advanced() = %v, want %v", got, tt.want)
			}
		})
	}
}
