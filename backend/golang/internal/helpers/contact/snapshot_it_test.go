package contact

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/myfusionhelper/api/internal/connectors"
	"github.com/myfusionhelper/api/internal/helpers"
)

// mockConnectorForSnapshotIt mocks the CRMConnector interface for testing snapshot_it
type mockConnectorForSnapshotIt struct {
	contact         *connectors.NormalizedContact
	getContactError error
}

func (m *mockConnectorForSnapshotIt) GetContact(ctx context.Context, contactID string) (*connectors.NormalizedContact, error) {
	if m.getContactError != nil {
		return nil, m.getContactError
	}
	return m.contact, nil
}

func (m *mockConnectorForSnapshotIt) SetContactFieldValue(ctx context.Context, contactID, fieldKey string, value interface{}) error {
	return fmt.Errorf("not implemented")
}

func (m *mockConnectorForSnapshotIt) GetContactFieldValue(ctx context.Context, contactID, fieldKey string) (interface{}, error) {
	return nil, fmt.Errorf("not implemented")
}

// Stub implementations
func (m *mockConnectorForSnapshotIt) GetContacts(ctx context.Context, opts connectors.QueryOptions) (*connectors.ContactList, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *mockConnectorForSnapshotIt) CreateContact(ctx context.Context, input connectors.CreateContactInput) (*connectors.NormalizedContact, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *mockConnectorForSnapshotIt) UpdateContact(ctx context.Context, contactID string, updates connectors.UpdateContactInput) (*connectors.NormalizedContact, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *mockConnectorForSnapshotIt) DeleteContact(ctx context.Context, contactID string) error {
	return fmt.Errorf("not implemented")
}

func (m *mockConnectorForSnapshotIt) GetTags(ctx context.Context) ([]connectors.Tag, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *mockConnectorForSnapshotIt) ApplyTag(ctx context.Context, contactID, tagID string) error {
	return fmt.Errorf("not implemented")
}

func (m *mockConnectorForSnapshotIt) RemoveTag(ctx context.Context, contactID, tagID string) error {
	return fmt.Errorf("not implemented")
}

func (m *mockConnectorForSnapshotIt) GetCustomFields(ctx context.Context) ([]connectors.CustomField, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *mockConnectorForSnapshotIt) TriggerAutomation(ctx context.Context, contactID, automationID string) error {
	return fmt.Errorf("not implemented")
}

func (m *mockConnectorForSnapshotIt) AchieveGoal(ctx context.Context, contactID, goalName, integration string) error {
	return fmt.Errorf("not implemented")
}

func (m *mockConnectorForSnapshotIt) TestConnection(ctx context.Context) error {
	return fmt.Errorf("not implemented")
}

func (m *mockConnectorForSnapshotIt) GetMetadata() connectors.ConnectorMetadata {
	return connectors.ConnectorMetadata{PlatformSlug: "mock", PlatformName: "Mock"}
}

func (m *mockConnectorForSnapshotIt) GetCapabilities() []connectors.Capability {
	return []connectors.Capability{}
}

func TestSnapshotIt_GetMetadata(t *testing.T) {
	helper := &SnapshotIt{}

	if helper.GetName() != "Snapshot It" {
		t.Errorf("Expected name 'Snapshot It', got '%s'", helper.GetName())
	}
	if helper.GetType() != "snapshot_it" {
		t.Errorf("Expected type 'snapshot_it', got '%s'", helper.GetType())
	}
	if helper.GetCategory() != "contact" {
		t.Errorf("Expected category 'contact', got '%s'", helper.GetCategory())
	}
	if !helper.RequiresCRM() {
		t.Error("Expected RequiresCRM to be true")
	}
}

func TestSnapshotIt_ValidateConfig_Success(t *testing.T) {
	helper := &SnapshotIt{}

	// Empty config is valid - all fields are optional
	config := map[string]interface{}{}

	err := helper.ValidateConfig(config)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
}

func TestSnapshotIt_Execute_SuccessBasicContact(t *testing.T) {
	helper := &SnapshotIt{}
	now := time.Now()
	mock := &mockConnectorForSnapshotIt{
		contact: &connectors.NormalizedContact{
			ID:        "123",
			FirstName: "John",
			LastName:  "Doe",
			Email:     "john@example.com",
			Phone:     "+1234567890",
			Company:   "Acme Corp",
			JobTitle:  "Developer",
			SourceCRM: "keap",
			CreatedAt: &now,
			UpdatedAt: &now,
		},
	}

	input := helpers.HelperInput{
		ContactID: "123",
		AccountID: "acc-456",
		HelperID:  "helper-789",
		UserID:    "user-111",
		Config:    map[string]interface{}{},
		Connector: mock,
	}

	output, err := helper.Execute(context.Background(), input)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if !output.Success {
		t.Error("Expected success to be true")
	}

	// Verify snapshot structure
	snapshot := output.ModifiedData
	if snapshot == nil {
		t.Fatal("Expected ModifiedData to not be nil")
	}

	// Verify required fields
	if snapshot["contact_id"] != "123" {
		t.Errorf("Expected contact_id to be '123', got: %v", snapshot["contact_id"])
	}
	if snapshot["first_name"] != "John" {
		t.Errorf("Expected first_name to be 'John', got: %v", snapshot["first_name"])
	}
	if snapshot["last_name"] != "Doe" {
		t.Errorf("Expected last_name to be 'Doe', got: %v", snapshot["last_name"])
	}
	if snapshot["email"] != "john@example.com" {
		t.Errorf("Expected email to be 'john@example.com', got: %v", snapshot["email"])
	}
	if snapshot["full_name"] != "John Doe" {
		t.Errorf("Expected full_name to be 'John Doe', got: %v", snapshot["full_name"])
	}
	if snapshot["source_crm"] != "keap" {
		t.Errorf("Expected source_crm to be 'keap', got: %v", snapshot["source_crm"])
	}

	// Verify IDs
	if snapshot["account_id"] != "acc-456" {
		t.Errorf("Expected account_id to be 'acc-456', got: %v", snapshot["account_id"])
	}
	if snapshot["helper_id"] != "helper-789" {
		t.Errorf("Expected helper_id to be 'helper-789', got: %v", snapshot["helper_id"])
	}

	// Verify snapshot metadata
	if snapshot["table_name"] != "myfusion_helper_contact_snapshots" {
		t.Errorf("Expected table_name to be 'myfusion_helper_contact_snapshots', got: %v", snapshot["table_name"])
	}

	// Verify action
	if len(output.Actions) != 1 {
		t.Errorf("Expected 1 action, got %d", len(output.Actions))
	}
	if output.Actions[0].Type != "snapshot_captured" {
		t.Errorf("Expected action type 'snapshot_captured', got: %s", output.Actions[0].Type)
	}
}

func TestSnapshotIt_Execute_WithTags(t *testing.T) {
	helper := &SnapshotIt{}
	mock := &mockConnectorForSnapshotIt{
		contact: &connectors.NormalizedContact{
			ID:        "123",
			FirstName: "Jane",
			LastName:  "Smith",
			Email:     "jane@example.com",
			Tags: []connectors.TagRef{
				{ID: "tag1", Name: "VIP"},
				{ID: "tag2", Name: "Customer"},
			},
		},
	}

	input := helpers.HelperInput{
		ContactID: "123",
		AccountID: "acc-456",
		HelperID:  "helper-789",
		UserID:    "user-111",
		Config: map[string]interface{}{
			"include_tags": true,
		},
		Connector: mock,
	}

	output, err := helper.Execute(context.Background(), input)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	snapshot := output.ModifiedData

	// Verify tags are included
	tags, ok := snapshot["tags"].([]map[string]string)
	if !ok {
		t.Fatal("Expected tags to be []map[string]string")
	}
	if len(tags) != 2 {
		t.Errorf("Expected 2 tags, got %d", len(tags))
	}
	if tags[0]["name"] != "VIP" {
		t.Errorf("Expected first tag name to be 'VIP', got: %s", tags[0]["name"])
	}
	if snapshot["tag_count"] != 2 {
		t.Errorf("Expected tag_count to be 2, got: %v", snapshot["tag_count"])
	}
}

func TestSnapshotIt_Execute_WithoutTags(t *testing.T) {
	helper := &SnapshotIt{}
	mock := &mockConnectorForSnapshotIt{
		contact: &connectors.NormalizedContact{
			ID:        "123",
			FirstName: "Jane",
			LastName:  "Smith",
			Tags: []connectors.TagRef{
				{ID: "tag1", Name: "VIP"},
			},
		},
	}

	input := helpers.HelperInput{
		ContactID: "123",
		AccountID: "acc-456",
		HelperID:  "helper-789",
		UserID:    "user-111",
		Config: map[string]interface{}{
			"include_tags": false,
		},
		Connector: mock,
	}

	output, err := helper.Execute(context.Background(), input)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	snapshot := output.ModifiedData

	// Verify tags are not included
	if _, exists := snapshot["tags"]; exists {
		t.Error("Expected tags to not be included when include_tags is false")
	}
	if _, exists := snapshot["tag_count"]; exists {
		t.Error("Expected tag_count to not be included when include_tags is false")
	}
}

func TestSnapshotIt_Execute_WithCustomFields(t *testing.T) {
	helper := &SnapshotIt{}
	mock := &mockConnectorForSnapshotIt{
		contact: &connectors.NormalizedContact{
			ID:        "123",
			FirstName: "Bob",
			LastName:  "Johnson",
			CustomFields: map[string]interface{}{
				"industry":    "Technology",
				"company_size": 100,
				"website":     "https://example.com",
			},
		},
	}

	input := helpers.HelperInput{
		ContactID: "123",
		AccountID: "acc-456",
		HelperID:  "helper-789",
		UserID:    "user-111",
		Config: map[string]interface{}{
			"include_custom_fields": true,
		},
		Connector: mock,
	}

	output, err := helper.Execute(context.Background(), input)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	snapshot := output.ModifiedData

	// Verify custom fields are included
	if snapshot["industry"] != "Technology" {
		t.Errorf("Expected industry to be 'Technology', got: %v", snapshot["industry"])
	}
	if snapshot["company_size"] != 100 {
		t.Errorf("Expected company_size to be 100, got: %v", snapshot["company_size"])
	}
	if snapshot["website"] != "https://example.com" {
		t.Errorf("Expected website to be 'https://example.com', got: %v", snapshot["website"])
	}
}

func TestSnapshotIt_Execute_WithoutCustomFields(t *testing.T) {
	helper := &SnapshotIt{}
	mock := &mockConnectorForSnapshotIt{
		contact: &connectors.NormalizedContact{
			ID:        "123",
			FirstName: "Bob",
			LastName:  "Johnson",
			CustomFields: map[string]interface{}{
				"industry": "Technology",
			},
		},
	}

	input := helpers.HelperInput{
		ContactID: "123",
		AccountID: "acc-456",
		HelperID:  "helper-789",
		UserID:    "user-111",
		Config: map[string]interface{}{
			"include_custom_fields": false,
		},
		Connector: mock,
	}

	output, err := helper.Execute(context.Background(), input)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	snapshot := output.ModifiedData

	// Verify custom fields are not included (but standard fields are)
	if _, exists := snapshot["industry"]; exists {
		t.Error("Expected industry to not be included when include_custom_fields is false")
	}
	if snapshot["first_name"] != "Bob" {
		t.Error("Expected standard fields to still be present")
	}
}

func TestSnapshotIt_Execute_GetContactError(t *testing.T) {
	helper := &SnapshotIt{}
	mock := &mockConnectorForSnapshotIt{
		getContactError: fmt.Errorf("contact not found"),
	}

	input := helpers.HelperInput{
		ContactID: "123",
		AccountID: "acc-456",
		HelperID:  "helper-789",
		UserID:    "user-111",
		Config:    map[string]interface{}{},
		Connector: mock,
	}

	output, err := helper.Execute(context.Background(), input)
	if err == nil {
		t.Error("Expected error for GetContact failure")
	}
	if output.Success {
		t.Error("Expected success to be false")
	}
	if output.Message != "Failed to get contact: contact not found" {
		t.Errorf("Unexpected message: %s", output.Message)
	}
}

func TestSnapshotIt_Execute_SnapshotIDUniqueness(t *testing.T) {
	helper := &SnapshotIt{}
	mock := &mockConnectorForSnapshotIt{
		contact: &connectors.NormalizedContact{
			ID:        "123",
			FirstName: "Test",
			LastName:  "User",
		},
	}

	input := helpers.HelperInput{
		ContactID: "123",
		AccountID: "acc-456",
		HelperID:  "helper-789",
		UserID:    "user-111",
		Config:    map[string]interface{}{},
		Connector: mock,
	}

	// Run twice to verify different snapshot IDs
	output1, _ := helper.Execute(context.Background(), input)
	output2, _ := helper.Execute(context.Background(), input)

	snapshot1 := output1.ModifiedData
	snapshot2 := output2.ModifiedData

	id1 := snapshot1["id"]
	id2 := snapshot2["id"]

	if id1 == id2 {
		t.Error("Expected different snapshot IDs for separate executions")
	}
}

func TestSnapshotIt_Execute_TimestampPresent(t *testing.T) {
	helper := &SnapshotIt{}
	mock := &mockConnectorForSnapshotIt{
		contact: &connectors.NormalizedContact{
			ID:        "123",
			FirstName: "Test",
			LastName:  "User",
		},
	}

	input := helpers.HelperInput{
		ContactID: "123",
		AccountID: "acc-456",
		HelperID:  "helper-789",
		UserID:    "user-111",
		Config:    map[string]interface{}{},
		Connector: mock,
	}

	output, err := helper.Execute(context.Background(), input)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	snapshot := output.ModifiedData

	// Verify timestamp exists and is recent
	timestamp, ok := snapshot["timestamp"].(int64)
	if !ok {
		t.Fatal("Expected timestamp to be int64")
	}

	now := time.Now().Unix()
	if timestamp < now-10 || timestamp > now+10 {
		t.Errorf("Expected timestamp to be recent (within 10s), got %d, now: %d", timestamp, now)
	}
}

func TestSnapshotIt_Execute_DefaultBehavior(t *testing.T) {
	helper := &SnapshotIt{}
	mock := &mockConnectorForSnapshotIt{
		contact: &connectors.NormalizedContact{
			ID:        "123",
			FirstName: "Default",
			LastName:  "Test",
			Tags: []connectors.TagRef{
				{ID: "tag1", Name: "Tag1"},
			},
			CustomFields: map[string]interface{}{
				"custom1": "value1",
			},
		},
	}

	input := helpers.HelperInput{
		ContactID: "123",
		AccountID: "acc-456",
		HelperID:  "helper-789",
		UserID:    "user-111",
		Config:    map[string]interface{}{
			// No explicit config - should default to include_tags=true, include_custom_fields=true
		},
		Connector: mock,
	}

	output, err := helper.Execute(context.Background(), input)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	snapshot := output.ModifiedData

	// Verify defaults include tags and custom fields
	if _, exists := snapshot["tags"]; !exists {
		t.Error("Expected tags to be included by default")
	}
	if _, exists := snapshot["custom1"]; !exists {
		t.Error("Expected custom fields to be included by default")
	}
}

func TestSnapshotIt_Execute_EmptyFullName(t *testing.T) {
	helper := &SnapshotIt{}
	mock := &mockConnectorForSnapshotIt{
		contact: &connectors.NormalizedContact{
			ID:        "123",
			FirstName: "",
			LastName:  "",
			Email:     "test@example.com",
		},
	}

	input := helpers.HelperInput{
		ContactID: "123",
		AccountID: "acc-456",
		HelperID:  "helper-789",
		UserID:    "user-111",
		Config:    map[string]interface{}{},
		Connector: mock,
	}

	output, err := helper.Execute(context.Background(), input)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	snapshot := output.ModifiedData

	// Verify full_name is empty string (trimmed)
	if snapshot["full_name"] != "" {
		t.Errorf("Expected full_name to be empty, got: '%v'", snapshot["full_name"])
	}
}
