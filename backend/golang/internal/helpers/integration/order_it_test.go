package integration

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/myfusionhelper/api/internal/connectors"
	"github.com/myfusionhelper/api/internal/helpers"
)

// Mock connector for order_it testing
type mockConnectorForOrder struct {
	contact       *connectors.NormalizedContact
	appliedTags   []string
	updatedFields map[string]interface{}
}

func (m *mockConnectorForOrder) GetContact(ctx context.Context, contactID string) (*connectors.NormalizedContact, error) {
	if m.contact != nil {
		return m.contact, nil
	}
	return nil, fmt.Errorf("contact not found")
}

func (m *mockConnectorForOrder) ApplyTag(ctx context.Context, contactID, tagID string) error {
	m.appliedTags = append(m.appliedTags, tagID)
	return nil
}

// Implement remaining interface methods as stubs
func (m *mockConnectorForOrder) GetContacts(ctx context.Context, opts connectors.QueryOptions) (*connectors.ContactList, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForOrder) CreateContact(ctx context.Context, input connectors.CreateContactInput) (*connectors.NormalizedContact, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForOrder) UpdateContact(ctx context.Context, contactID string, updates connectors.UpdateContactInput) (*connectors.NormalizedContact, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForOrder) DeleteContact(ctx context.Context, contactID string) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorForOrder) GetTags(ctx context.Context) ([]connectors.Tag, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForOrder) RemoveTag(ctx context.Context, contactID, tagID string) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorForOrder) GetCustomFields(ctx context.Context) ([]connectors.CustomField, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForOrder) GetContactFieldValue(ctx context.Context, contactID, fieldKey string) (interface{}, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForOrder) SetContactFieldValue(ctx context.Context, contactID, fieldKey string, value interface{}) error {
	if m.updatedFields == nil {
		m.updatedFields = make(map[string]interface{})
	}
	m.updatedFields[fieldKey] = value
	return nil
}
func (m *mockConnectorForOrder) TriggerAutomation(ctx context.Context, contactID, automationID string) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorForOrder) AchieveGoal(ctx context.Context, contactID, goalName, integration string) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorForOrder) TestConnection(ctx context.Context) error {
	return nil
}
func (m *mockConnectorForOrder) GetMetadata() connectors.ConnectorMetadata {
	return connectors.ConnectorMetadata{}
}
func (m *mockConnectorForOrder) GetCapabilities() []connectors.Capability {
	return nil
}

// Test helper metadata
func TestOrderIt_Metadata(t *testing.T) {
	helper := &OrderIt{}

	if helper.GetName() != "Order It" {
		t.Errorf("Expected name 'Order It', got '%s'", helper.GetName())
	}
	if helper.GetType() != "order_it" {
		t.Errorf("Expected type 'order_it', got '%s'", helper.GetType())
	}
	if helper.GetCategory() != "integration" {
		t.Errorf("Expected category 'integration', got '%s'", helper.GetCategory())
	}
	if !helper.RequiresCRM() {
		t.Error("Expected RequiresCRM to be true")
	}

	supportedCRMs := helper.SupportedCRMs()
	if len(supportedCRMs) != 1 || supportedCRMs[0] != "keap" {
		t.Errorf("Expected SupportedCRMs to be ['keap'], got %v", supportedCRMs)
	}

	schema := helper.GetConfigSchema()
	if schema == nil {
		t.Fatal("Expected config schema, got nil")
	}
}

// Test validation - missing product_id
func TestOrderIt_ValidateConfig_MissingProductId(t *testing.T) {
	helper := &OrderIt{}
	config := map[string]interface{}{
		"quantity": 2,
	}

	err := helper.ValidateConfig(config)
	if err == nil {
		t.Error("Expected validation error for missing product_id")
	}
	if !strings.Contains(err.Error(), "product_id") {
		t.Errorf("Expected error about product_id, got: %v", err)
	}
}

// Test validation - invalid product_id (negative)
func TestOrderIt_ValidateConfig_InvalidProductId(t *testing.T) {
	helper := &OrderIt{}
	config := map[string]interface{}{
		"product_id": -5.0,
	}

	err := helper.ValidateConfig(config)
	if err == nil {
		t.Error("Expected validation error for negative product_id")
	}
	if !strings.Contains(err.Error(), "positive") {
		t.Errorf("Expected error about positive number, got: %v", err)
	}
}

// Test validation - product_id as string
func TestOrderIt_ValidateConfig_ProductIdAsString(t *testing.T) {
	helper := &OrderIt{}
	config := map[string]interface{}{
		"product_id": "123",
	}

	err := helper.ValidateConfig(config)
	if err != nil {
		t.Errorf("Expected no validation error for string product_id, got: %v", err)
	}
}

// Test validation - product_id as invalid string
func TestOrderIt_ValidateConfig_ProductIdAsInvalidString(t *testing.T) {
	helper := &OrderIt{}
	config := map[string]interface{}{
		"product_id": "abc",
	}

	err := helper.ValidateConfig(config)
	if err == nil {
		t.Error("Expected validation error for non-numeric string product_id")
	}
	if !strings.Contains(err.Error(), "valid number") {
		t.Errorf("Expected error about valid number, got: %v", err)
	}
}

// Test validation - valid config with float product_id
func TestOrderIt_ValidateConfig_ValidFloat(t *testing.T) {
	helper := &OrderIt{}
	config := map[string]interface{}{
		"product_id": 456.0,
	}

	err := helper.ValidateConfig(config)
	if err != nil {
		t.Errorf("Expected no validation error, got: %v", err)
	}
}

// Test validation - valid config with all fields
func TestOrderIt_ValidateConfig_ValidComplete(t *testing.T) {
	helper := &OrderIt{}
	config := map[string]interface{}{
		"product_id":  123.0,
		"quantity":    2.0,
		"promo_codes": []interface{}{"SAVE10", "WELCOME"},
		"apply_tag":   "order-created",
	}

	err := helper.ValidateConfig(config)
	if err != nil {
		t.Errorf("Expected no validation error, got: %v", err)
	}
}

// Test execution - missing Keap connection
func TestOrderIt_Execute_MissingConnection(t *testing.T) {
	helper := &OrderIt{}

	input := helpers.HelperInput{
		ContactID:    "123",
		ServiceAuths: map[string]*connectors.ConnectorConfig{},
		Config: map[string]interface{}{
			"product_id": 456.0,
		},
	}

	output, err := helper.Execute(context.Background(), input)
	if err == nil {
		t.Error("Expected error for missing Keap connection")
	}
	if !strings.Contains(output.Message, "Keap service connection required") {
		t.Errorf("Expected error message about Keap connection, got: %s", output.Message)
	}
}

// Test execution - invalid contact ID (non-numeric)
func TestOrderIt_Execute_InvalidContactId(t *testing.T) {
	helper := &OrderIt{}

	mockConn := &mockConnectorForOrder{
		contact: &connectors.NormalizedContact{
			ID:        "abc-not-numeric",
			FirstName: "Test",
			LastName:  "User",
		},
	}

	input := helpers.HelperInput{
		ContactID: "abc-not-numeric",
		ContactData: &connectors.NormalizedContact{
			ID:        "abc-not-numeric",
			FirstName: "Test",
			LastName:  "User",
		},
		Connector: mockConn,
		ServiceAuths: map[string]*connectors.ConnectorConfig{
			"keap": {
				AccessToken: "test-token",
				BaseURL:     "https://api.test.com",
			},
		},
		Config: map[string]interface{}{
			"product_id": 789.0,
		},
	}

	output, err := helper.Execute(context.Background(), input)
	if err == nil {
		t.Error("Expected error for non-numeric contact ID")
	}
	if !strings.Contains(output.Message, "Invalid contact ID") {
		t.Errorf("Expected error about invalid contact ID, got: %s", output.Message)
	}
}

// Test execution - with contact fetch
func TestOrderIt_Execute_FetchContact(t *testing.T) {
	helper := &OrderIt{}

	mockConn := &mockConnectorForOrder{
		contact: &connectors.NormalizedContact{
			ID:        "456",
			FirstName: "Jane",
			LastName:  "Doe",
			Email:     "jane@example.com",
		},
	}

	// Create mock Keap API server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"id": 12345, "contact_id": 456, "total": 99.99}`))
	}))
	defer server.Close()

	input := helpers.HelperInput{
		ContactID: "456",
		Connector: mockConn,
		ServiceAuths: map[string]*connectors.ConnectorConfig{
			"keap": {
				AccessToken: "test-token",
				BaseURL:     server.URL,
			},
		},
		Config: map[string]interface{}{
			"product_id": 123.0,
		},
	}

	output, err := helper.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if !output.Success {
		t.Error("Expected success=true")
	}

	// Verify the contact was fetched
	if len(output.Logs) == 0 {
		t.Error("Expected logs to be generated")
	}
}

// Test successful execution with mock server
func TestOrderIt_Execute_Success(t *testing.T) {
	// Create mock Keap API server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request method
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST method, got: %s", r.Method)
		}

		// Verify authorization header
		authHeader := r.Header.Get("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			t.Errorf("Expected Bearer token in Authorization header, got: %s", authHeader)
		}

		// Verify content type
		contentType := r.Header.Get("Content-Type")
		if contentType != "application/json" {
			t.Errorf("Expected Content-Type application/json, got: %s", contentType)
		}

		// Return mock response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{
			"id": 12345,
			"contact_id": 123,
			"total": 149.99,
			"status": "PAID"
		}`))
	}))
	defer server.Close()

	helper := &OrderIt{}

	mockConn := &mockConnectorForOrder{
		contact: &connectors.NormalizedContact{
			ID:        "123",
			FirstName: "Test",
			LastName:  "User",
			Email:     "test@example.com",
		},
	}

	input := helpers.HelperInput{
		ContactID: "123",
		ContactData: &connectors.NormalizedContact{
			ID:        "123",
			FirstName: "Test",
			LastName:  "User",
			Email:     "test@example.com",
		},
		Connector: mockConn,
		ServiceAuths: map[string]*connectors.ConnectorConfig{
			"keap": {
				AccessToken: "test-api-token",
				BaseURL:     server.URL,
			},
		},
		Config: map[string]interface{}{
			"product_id": 789.0,
			"quantity":   2.0,
		},
	}

	output, err := helper.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if !output.Success {
		t.Error("Expected success=true")
	}

	if output.ModifiedData == nil {
		t.Fatal("Expected ModifiedData to be non-nil")
	}

	// Verify order_id is set
	orderID, ok := output.ModifiedData["order_id"].(string)
	if !ok || orderID != "12345" {
		t.Errorf("Expected order_id '12345', got: %v", output.ModifiedData["order_id"])
	}

	// Verify product_id
	productID, ok := output.ModifiedData["product_id"].(int)
	if !ok || productID != 789 {
		t.Errorf("Expected product_id 789, got: %v", output.ModifiedData["product_id"])
	}

	// Verify quantity
	quantity, ok := output.ModifiedData["quantity"].(int)
	if !ok || quantity != 2 {
		t.Errorf("Expected quantity 2, got: %v", output.ModifiedData["quantity"])
	}
}

// Test execution with promo codes
func TestOrderIt_Execute_WithPromoCodes(t *testing.T) {
	// Create mock Keap API server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"id": 99999, "contact_id": 555}`))
	}))
	defer server.Close()

	helper := &OrderIt{}

	mockConn := &mockConnectorForOrder{
		contact: &connectors.NormalizedContact{
			ID:        "555",
			FirstName: "Promo",
			LastName:  "User",
		},
	}

	input := helpers.HelperInput{
		ContactID: "555",
		ContactData: &connectors.NormalizedContact{
			ID:        "555",
			FirstName: "Promo",
			LastName:  "User",
		},
		Connector: mockConn,
		ServiceAuths: map[string]*connectors.ConnectorConfig{
			"keap": {
				AccessToken: "test-token",
				BaseURL:     server.URL,
			},
		},
		Config: map[string]interface{}{
			"product_id":  100.0,
			"promo_codes": []interface{}{"SAVE20", "FIRSTORDER"},
		},
	}

	output, err := helper.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if !output.Success {
		t.Error("Expected success=true")
	}

	// Verify promo codes in logs
	foundPromoLog := false
	for _, log := range output.Logs {
		if strings.Contains(log, "promo codes") {
			foundPromoLog = true
			break
		}
	}
	if !foundPromoLog {
		t.Error("Expected log about promo codes")
	}

	// Verify promo codes in modified data
	promoCodes, ok := output.ModifiedData["promo_codes"].([]string)
	if !ok || len(promoCodes) != 2 {
		t.Errorf("Expected 2 promo codes in ModifiedData, got: %v", output.ModifiedData["promo_codes"])
	}
}

// Test tag application
func TestOrderIt_Execute_TagApplication(t *testing.T) {
	// Create mock Keap API server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"id": 77777, "contact_id": 999}`))
	}))
	defer server.Close()

	helper := &OrderIt{}

	mockConn := &mockConnectorForOrder{
		contact: &connectors.NormalizedContact{
			ID:        "999",
			FirstName: "Tag",
			LastName:  "Test",
		},
		appliedTags: make([]string, 0),
	}

	input := helpers.HelperInput{
		ContactID: "999",
		ContactData: &connectors.NormalizedContact{
			ID:        "999",
			FirstName: "Tag",
			LastName:  "Test",
		},
		Connector: mockConn,
		ServiceAuths: map[string]*connectors.ConnectorConfig{
			"keap": {
				AccessToken: "test-token",
				BaseURL:     server.URL,
			},
		},
		Config: map[string]interface{}{
			"product_id": 200.0,
			"apply_tag":  "order-completed",
		},
	}

	output, err := helper.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if !output.Success {
		t.Error("Expected success=true")
	}

	// Verify tag was applied
	if len(mockConn.appliedTags) != 1 || mockConn.appliedTags[0] != "order-completed" {
		t.Errorf("Expected tag 'order-completed' to be applied, got: %v", mockConn.appliedTags)
	}

	// Verify tag action in output
	foundTagAction := false
	for _, action := range output.Actions {
		if action.Type == "tag_applied" && action.Value == "order-completed" {
			foundTagAction = true
			break
		}
	}
	if !foundTagAction {
		t.Error("Expected tag_applied action in output")
	}
}

// Test action logging
func TestOrderIt_Execute_ActionLogging(t *testing.T) {
	// Create mock Keap API server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"id": 88888, "contact_id": 888}`))
	}))
	defer server.Close()

	helper := &OrderIt{}

	mockConn := &mockConnectorForOrder{
		contact: &connectors.NormalizedContact{
			ID:        "888",
			FirstName: "Action",
			LastName:  "Logger",
		},
	}

	input := helpers.HelperInput{
		ContactID: "888",
		ContactData: &connectors.NormalizedContact{
			ID:        "888",
			FirstName: "Action",
			LastName:  "Logger",
		},
		Connector: mockConn,
		ServiceAuths: map[string]*connectors.ConnectorConfig{
			"keap": {
				AccessToken: "test-token",
				BaseURL:     server.URL,
			},
		},
		Config: map[string]interface{}{
			"product_id": 300.0,
			"quantity":   5.0,
		},
	}

	output, err := helper.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Verify logs were created
	if len(output.Logs) == 0 {
		t.Error("Expected logs to be generated")
	}

	// First log should mention creating the order
	if !strings.Contains(output.Logs[0], "Creating Keap order") {
		t.Errorf("Expected first log to mention order creation, got: %s", output.Logs[0])
	}
	if !strings.Contains(output.Logs[0], "product: 300") {
		t.Errorf("Expected first log to mention product ID, got: %s", output.Logs[0])
	}
	if !strings.Contains(output.Logs[0], "qty: 5") {
		t.Errorf("Expected first log to mention quantity, got: %s", output.Logs[0])
	}

	// Last log should mention order creation
	lastLog := output.Logs[len(output.Logs)-1]
	if !strings.Contains(lastLog, "order created") {
		t.Errorf("Expected last log to mention order creation, got: %s", lastLog)
	}
}

// Test default quantity (should be 1)
func TestOrderIt_Execute_DefaultQuantity(t *testing.T) {
	// Create mock Keap API server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"id": 66666, "contact_id": 777}`))
	}))
	defer server.Close()

	helper := &OrderIt{}

	mockConn := &mockConnectorForOrder{
		contact: &connectors.NormalizedContact{
			ID:        "777",
			FirstName: "Default",
			LastName:  "Quantity",
		},
	}

	input := helpers.HelperInput{
		ContactID: "777",
		ContactData: &connectors.NormalizedContact{
			ID:        "777",
			FirstName: "Default",
			LastName:  "Quantity",
		},
		Connector: mockConn,
		ServiceAuths: map[string]*connectors.ConnectorConfig{
			"keap": {
				AccessToken: "test-token",
				BaseURL:     server.URL,
			},
		},
		Config: map[string]interface{}{
			"product_id": 400.0,
			// No quantity specified - should default to 1
		},
	}

	output, err := helper.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if !output.Success {
		t.Error("Expected success=true")
	}

	// Verify quantity defaults to 1
	quantity, ok := output.ModifiedData["quantity"].(int)
	if !ok || quantity != 1 {
		t.Errorf("Expected default quantity 1, got: %v", output.ModifiedData["quantity"])
	}
}

// Test product_id as string (conversion)
func TestOrderIt_Execute_ProductIdAsString(t *testing.T) {
	// Create mock Keap API server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"id": 55555, "contact_id": 666}`))
	}))
	defer server.Close()

	helper := &OrderIt{}

	mockConn := &mockConnectorForOrder{
		contact: &connectors.NormalizedContact{
			ID:        "666",
			FirstName: "String",
			LastName:  "ProductID",
		},
	}

	input := helpers.HelperInput{
		ContactID: "666",
		ContactData: &connectors.NormalizedContact{
			ID:        "666",
			FirstName: "String",
			LastName:  "ProductID",
		},
		Connector: mockConn,
		ServiceAuths: map[string]*connectors.ConnectorConfig{
			"keap": {
				AccessToken: "test-token",
				BaseURL:     server.URL,
			},
		},
		Config: map[string]interface{}{
			"product_id": "500", // String instead of float
		},
	}

	output, err := helper.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if !output.Success {
		t.Error("Expected success=true")
	}

	// Verify product_id was converted correctly
	productID, ok := output.ModifiedData["product_id"].(int)
	if !ok || productID != 500 {
		t.Errorf("Expected product_id 500, got: %v", output.ModifiedData["product_id"])
	}
}
