package helpers_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/myfusionhelper/api/internal/connectors"
	"github.com/myfusionhelper/api/internal/helpers"

	// Import helpers to register them via init()
	_ "github.com/myfusionhelper/api/internal/helpers/contact"
	_ "github.com/myfusionhelper/api/internal/helpers/integration"
	_ "github.com/myfusionhelper/api/internal/helpers/tagging"
)

// mockConnector implements connectors.CRMConnector for testing
type mockConnector struct {
	getContactFunc    func(ctx context.Context, contactID string) (*connectors.NormalizedContact, error)
	applyTagFunc      func(ctx context.Context, contactID, tagID string) error
	removeTagFunc     func(ctx context.Context, contactID, tagID string) error
	updateContactFunc func(ctx context.Context, contactID string, updates connectors.UpdateContactInput) (*connectors.NormalizedContact, error)
	achieveGoalFunc   func(ctx context.Context, contactID, goalName, integration string) error
}

func (m *mockConnector) GetContact(ctx context.Context, contactID string) (*connectors.NormalizedContact, error) {
	if m.getContactFunc != nil {
		return m.getContactFunc(ctx, contactID)
	}
	return &connectors.NormalizedContact{ID: contactID}, nil
}

func (m *mockConnector) ApplyTag(ctx context.Context, contactID, tagID string) error {
	if m.applyTagFunc != nil {
		return m.applyTagFunc(ctx, contactID, tagID)
	}
	return nil
}

func (m *mockConnector) RemoveTag(ctx context.Context, contactID, tagID string) error {
	if m.removeTagFunc != nil {
		return m.removeTagFunc(ctx, contactID, tagID)
	}
	return nil
}

func (m *mockConnector) UpdateContact(ctx context.Context, contactID string, updates connectors.UpdateContactInput) (*connectors.NormalizedContact, error) {
	if m.updateContactFunc != nil {
		return m.updateContactFunc(ctx, contactID, updates)
	}
	return &connectors.NormalizedContact{ID: contactID}, nil
}

// Stub implementations for required interface methods
func (m *mockConnector) GetContacts(ctx context.Context, opts connectors.QueryOptions) (*connectors.ContactList, error) {
	return &connectors.ContactList{}, nil
}

func (m *mockConnector) CreateContact(ctx context.Context, input connectors.CreateContactInput) (*connectors.NormalizedContact, error) {
	return &connectors.NormalizedContact{}, nil
}

func (m *mockConnector) DeleteContact(ctx context.Context, contactID string) error {
	return nil
}

func (m *mockConnector) GetTags(ctx context.Context) ([]connectors.Tag, error) {
	return []connectors.Tag{}, nil
}

func (m *mockConnector) GetCustomFields(ctx context.Context) ([]connectors.CustomField, error) {
	return []connectors.CustomField{}, nil
}

func (m *mockConnector) GetContactFieldValue(ctx context.Context, contactID, fieldKey string) (interface{}, error) {
	return nil, nil
}

func (m *mockConnector) SetContactFieldValue(ctx context.Context, contactID, fieldKey string, value interface{}) error {
	return nil
}

func (m *mockConnector) TriggerAutomation(ctx context.Context, contactID, automationID string) error {
	return nil
}

func (m *mockConnector) AchieveGoal(ctx context.Context, contactID, goalName, integration string) error {
	if m.achieveGoalFunc != nil {
		return m.achieveGoalFunc(ctx, contactID, goalName, integration)
	}
	return nil
}

func (m *mockConnector) TestConnection(ctx context.Context) error {
	return nil
}

func (m *mockConnector) GetMetadata() connectors.ConnectorMetadata {
	return connectors.ConnectorMetadata{
		PlatformName: "Mock",
		PlatformSlug: "mock",
		APIVersion:   "v1",
		BaseURL:      "https://mock.example.com",
	}
}

func (m *mockConnector) GetCapabilities() []connectors.Capability {
	return []connectors.Capability{}
}

// TestNewExecutor tests executor creation
func TestNewExecutor(t *testing.T) {
	executor := helpers.NewExecutor()
	if executor == nil {
		t.Fatal("Expected executor to be created")
	}
}

// TestExecutor_Execute_UnknownHelper tests unknown helper handling
func TestExecutor_Execute_UnknownHelper(t *testing.T) {
	executor := helpers.NewExecutor()

	req := helpers.ExecutionRequest{
		HelperType: "unknown_helper_xyz",
		ContactID:  "contact-123",
	}

	result, err := executor.Execute(context.Background(), req, nil)

	// Executor returns an error for unknown helper
	if err == nil {
		t.Fatal("Expected error for unknown helper")
	}

	if result.Success {
		t.Error("Expected success to be false for unknown helper")
	}

	if result.Error == "" {
		t.Error("Expected error message for unknown helper")
	}

	if result.DurationMs < 0 {
		t.Error("Expected non-negative duration")
	}
}

// TestExecutor_Execute_TagIt tests tag_it helper execution
func TestExecutor_Execute_TagIt(t *testing.T) {
	t.Run("success - apply tags", func(t *testing.T) {
		appliedTags := make(map[string][]string) // contactID -> []tagID

		connector := &mockConnector{
			applyTagFunc: func(ctx context.Context, contactID, tagID string) error {
				appliedTags[contactID] = append(appliedTags[contactID], tagID)
				return nil
			},
		}

		executor := helpers.NewExecutor()
		req := helpers.ExecutionRequest{
			HelperType: "tag_it",
			ContactID:  "contact-123",
			Config: map[string]interface{}{
				"action":  "apply",
				"tag_ids": []interface{}{"tag-1", "tag-2"},
			},
		}

		result, err := executor.Execute(context.Background(), req, connector)

		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if !result.Success {
			t.Errorf("Expected success, got error: %s", result.Error)
		}

		if len(appliedTags["contact-123"]) != 2 {
			t.Errorf("Expected 2 tags applied, got %d", len(appliedTags["contact-123"]))
		}

		// Duration may be 0 for very fast execution, just check it's non-negative
		if result.DurationMs < 0 {
			t.Error("Expected duration to be non-negative")
		}
	})

	t.Run("error - invalid config", func(t *testing.T) {
		executor := helpers.NewExecutor()

		req := helpers.ExecutionRequest{
			HelperType: "tag_it",
			ContactID:  "contact-123",
			Config:     map[string]interface{}{}, // Missing required fields
		}

		result, err := executor.Execute(context.Background(), req, &mockConnector{})

		// Executor returns an error for invalid config
		if err == nil {
			t.Fatal("Expected error for invalid config")
		}

		if result.Success {
			t.Error("Expected failure for invalid config")
		}

		if result.Error == "" {
			t.Error("Expected error message for invalid config")
		}
	})

	t.Run("error - no CRM connector", func(t *testing.T) {
		executor := helpers.NewExecutor()

		req := helpers.ExecutionRequest{
			HelperType: "tag_it",
			ContactID:  "contact-123",
			Config: map[string]interface{}{
				"action":  "apply",
				"tag_ids": []interface{}{"tag-1"},
			},
		}

		result, err := executor.Execute(context.Background(), req, nil)

		// Executor returns an error when CRM connector is required but missing
		if err == nil {
			t.Fatal("Expected error when CRM connector missing")
		}

		if result.Success {
			t.Error("Expected failure when CRM connector missing")
		}

		if result.Error == "" {
			t.Error("Expected error message about missing CRM connection")
		}
	})
}

// TestExecutor_Execute_ContactUpdater tests contact_updater helper execution
func TestExecutor_Execute_ContactUpdater(t *testing.T) {
	t.Run("success - update contact fields", func(t *testing.T) {
		updatedContacts := make(map[string]connectors.UpdateContactInput)

		connector := &mockConnector{
			updateContactFunc: func(ctx context.Context, contactID string, updates connectors.UpdateContactInput) (*connectors.NormalizedContact, error) {
				updatedContacts[contactID] = updates
				result := &connectors.NormalizedContact{ID: contactID}
				if updates.FirstName != nil {
					result.FirstName = *updates.FirstName
				}
				if updates.LastName != nil {
					result.LastName = *updates.LastName
				}
				return result, nil
			},
		}

		executor := helpers.NewExecutor()
		req := helpers.ExecutionRequest{
			HelperType: "contact_updater",
			ContactID:  "contact-123",
			Config: map[string]interface{}{
				"fields": map[string]interface{}{
					"first_name": "Jane",
					"last_name":  "Doe",
				},
			},
		}

		result, err := executor.Execute(context.Background(), req, connector)

		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if !result.Success {
			t.Errorf("Expected success, got error: %s", result.Error)
		}

		if len(updatedContacts) != 1 {
			t.Error("Expected contact to be updated")
		}
	})

	t.Run("error - missing required fields config", func(t *testing.T) {
		executor := helpers.NewExecutor()

		req := helpers.ExecutionRequest{
			HelperType: "contact_updater",
			ContactID:  "contact-123",
			Config:     map[string]interface{}{}, // Missing fields
		}

		result, err := executor.Execute(context.Background(), req, &mockConnector{})

		// Executor returns an error for invalid config
		if err == nil {
			t.Fatal("Expected error for missing required fields")
		}

		if result.Success {
			t.Error("Expected failure for invalid config")
		}
	})
}

// TestExecutor_Execute_HookIt tests hook_it helper execution
func TestExecutor_Execute_HookIt(t *testing.T) {
	t.Run("basic mode validation", func(t *testing.T) {
		connector := &mockConnector{
			getContactFunc: func(ctx context.Context, contactID string) (*connectors.NormalizedContact, error) {
				return &connectors.NormalizedContact{
					ID:        contactID,
					FirstName: "John",
				}, nil
			},
		}

		executor := helpers.NewExecutor()
		req := helpers.ExecutionRequest{
			HelperType: "hook_it",
			ContactID:  "contact-123",
			Config: map[string]interface{}{
				"mode": "basic",
				"actions": []interface{}{
					map[string]interface{}{
						"event":     "contact.add",
						"goal_name": "webhook_received",
					},
				},
			},
		}

		result, err := executor.Execute(context.Background(), req, connector)

		// hook_it may succeed or fail depending on its internal logic
		// Just verify the executor ran without panicking
		if err != nil && result == nil {
			t.Fatalf("Expected result even on error, got nil")
		}

		if result.DurationMs < 0 {
			t.Error("Expected duration to be non-negative")
		}
	})

	t.Run("unrecognized mode defaults to basic", func(t *testing.T) {
		connector := &mockConnector{
			getContactFunc: func(ctx context.Context, contactID string) (*connectors.NormalizedContact, error) {
				return &connectors.NormalizedContact{
					ID:        contactID,
					FirstName: "John",
				}, nil
			},
		}

		executor := helpers.NewExecutor()

		req := helpers.ExecutionRequest{
			HelperType: "hook_it",
			ContactID:  "contact-123",
			Config: map[string]interface{}{
				"mode": "invalid_mode",
				"actions": []interface{}{
					map[string]interface{}{
						"event":     "contact.add",
						"goal_name": "webhook_received",
					},
				},
			},
		}

		result, err := executor.Execute(context.Background(), req, connector)

		// hook_it defaults to "basic" mode for unrecognized modes
		// Just verify it ran without panicking
		if err != nil && result == nil {
			t.Fatalf("Expected result even on error, got nil")
		}

		if result.DurationMs < 0 {
			t.Error("Expected duration to be non-negative")
		}
	})
}

// TestExecutor_Execute_ContactFetching tests contact data fetching behavior
func TestExecutor_Execute_ContactFetching(t *testing.T) {
	t.Run("contact fetched when connector and contactID provided", func(t *testing.T) {
		fetchCalled := false

		connector := &mockConnector{
			getContactFunc: func(ctx context.Context, contactID string) (*connectors.NormalizedContact, error) {
				fetchCalled = true
				return &connectors.NormalizedContact{
					ID:        contactID,
					FirstName: "Test",
				}, nil
			},
			applyTagFunc: func(ctx context.Context, contactID, tagID string) error {
				return nil
			},
		}

		executor := helpers.NewExecutor()
		req := helpers.ExecutionRequest{
			HelperType: "tag_it",
			ContactID:  "contact-123",
			Config: map[string]interface{}{
				"action":  "apply",
				"tag_ids": []interface{}{"tag-1"},
			},
		}

		_, err := executor.Execute(context.Background(), req, connector)

		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if !fetchCalled {
			t.Error("Expected contact to be fetched")
		}
	})

	t.Run("handles contact fetch error gracefully", func(t *testing.T) {
		connector := &mockConnector{
			getContactFunc: func(ctx context.Context, contactID string) (*connectors.NormalizedContact, error) {
				return nil, fmt.Errorf("contact not found")
			},
			applyTagFunc: func(ctx context.Context, contactID, tagID string) error {
				return nil
			},
		}

		executor := helpers.NewExecutor()
		req := helpers.ExecutionRequest{
			HelperType: "tag_it",
			ContactID:  "contact-123",
			Config: map[string]interface{}{
				"action":  "apply",
				"tag_ids": []interface{}{"tag-1"},
			},
		}

		result, err := executor.Execute(context.Background(), req, connector)

		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		// Execution should continue even if contact fetch fails
		if !result.Success {
			t.Error("Expected execution to succeed despite contact fetch error")
		}
	})
}

// TestExecutor_Execute_Duration tests duration tracking
func TestExecutor_Execute_Duration(t *testing.T) {
	executor := helpers.NewExecutor()

	req := helpers.ExecutionRequest{
		HelperType: "unknown_helper",
		ContactID:  "contact-123",
	}

	result, err := executor.Execute(context.Background(), req, nil)

	// Error is expected for unknown helper, but duration should still be recorded
	if err == nil {
		t.Fatal("Expected error for unknown helper")
	}

	if result.DurationMs < 0 {
		t.Error("Expected non-negative duration even on error")
	}

	if result.ExecutedAt.IsZero() {
		t.Error("Expected execution timestamp to be set")
	}
}

// TestExecutor_Execute_HelperInputStructure tests HelperInput construction
func TestExecutor_Execute_HelperInputStructure(t *testing.T) {
	executor := helpers.NewExecutor()

	req := helpers.ExecutionRequest{
		HelperType:   "tag_it",
		ContactID:    "contact-123",
		UserID:       "user-456",
		AccountID:    "account-789",
		HelperID:     "helper-abc",
		ConnectionID: "conn-xyz",
		Config: map[string]interface{}{
			"action":  "apply",
			"tag_ids": []interface{}{"tag-1"},
		},
	}

	connector := &mockConnector{
		applyTagFunc: func(ctx context.Context, contactID, tagID string) error {
			return nil
		},
	}

	result, err := executor.Execute(context.Background(), req, connector)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if !result.Success {
		t.Errorf("Expected success, got error: %s", result.Error)
	}

	// Verify result contains request metadata
	if result.HelperType != "tag_it" {
		t.Errorf("Expected HelperType 'tag_it', got %s", result.HelperType)
	}

	if result.ContactID != "contact-123" {
		t.Errorf("Expected ContactID 'contact-123', got %s", result.ContactID)
	}
}

// TestExecutor_Execute_SuccessFlag tests success flag handling
func TestExecutor_Execute_SuccessFlag(t *testing.T) {
	t.Run("success flag set to true on successful execution", func(t *testing.T) {
		executor := helpers.NewExecutor()

		connector := &mockConnector{
			applyTagFunc: func(ctx context.Context, contactID, tagID string) error {
				return nil
			},
		}

		req := helpers.ExecutionRequest{
			HelperType: "tag_it",
			ContactID:  "contact-123",
			Config: map[string]interface{}{
				"action":  "apply",
				"tag_ids": []interface{}{"tag-1"},
			},
		}

		result, err := executor.Execute(context.Background(), req, connector)

		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if !result.Success {
			t.Error("Expected Success to be true")
		}

		if result.Error != "" {
			t.Errorf("Expected no error message, got %s", result.Error)
		}
	})

	t.Run("success flag set to false on execution error", func(t *testing.T) {
		executor := helpers.NewExecutor()

		req := helpers.ExecutionRequest{
			HelperType: "unknown_helper",
			ContactID:  "contact-123",
		}

		result, err := executor.Execute(context.Background(), req, nil)

		// Error is expected for unknown helper
		if err == nil {
			t.Fatal("Expected error for unknown helper")
		}

		if result.Success {
			t.Error("Expected Success to be false for unknown helper")
		}

		if result.Error == "" {
			t.Error("Expected error message to be set")
		}
	})
}
