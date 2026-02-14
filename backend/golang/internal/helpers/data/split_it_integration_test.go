// +build integration

package data

import (
	"testing"

	"github.com/myfusionhelper/api/internal/helpers"
)

func TestSplitItIntegration_Registration(t *testing.T) {
	// Verify split_it is registered
	if !helpers.IsRegistered("split_it") {
		t.Fatal("split_it should be registered")
	}

	// Verify we can create a new instance
	helper, err := helpers.NewHelper("split_it")
	if err != nil {
		t.Fatalf("NewHelper should not return error: %v", err)
	}
	if helper == nil {
		t.Fatal("NewHelper should return a split_it instance")
	}

	// Verify it implements the Helper interface
	if _, ok := helper.(helpers.Helper); !ok {
		t.Error("split_it should implement helpers.Helper interface")
	}

	// Verify type
	if helper.GetType() != "split_it" {
		t.Errorf("Expected type 'split_it', got '%s'", helper.GetType())
	}
}

func TestSplitItIntegration_ListHelpers(t *testing.T) {
	// Get all registered helpers
	helpersInfo := helpers.ListHelperInfo()

	// Find split_it in the list
	found := false
	for _, info := range helpersInfo {
		if info.Type == "split_it" {
			found = true

			// Verify basic info
			if info.Name != "Split It" {
				t.Errorf("Expected name 'Split It', got '%s'", info.Name)
			}
			if info.Category != "data" {
				t.Errorf("Expected category 'data', got '%s'", info.Category)
			}
			if !info.RequiresCRM {
				t.Error("Expected RequiresCRM to be true")
			}
			break
		}
	}

	if !found {
		t.Error("split_it should be in the helpers list")
	}
}

func TestSplitItIntegration_HelperInfo(t *testing.T) {
	helper, err := helpers.NewHelper("split_it")
	if err != nil {
		t.Fatalf("Failed to create split_it helper: %v", err)
	}
	if helper == nil {
		t.Fatal("Failed to create split_it helper")
	}

	// Verify metadata
	if helper.GetName() == "" {
		t.Error("Helper name should not be empty")
	}
	if helper.GetDescription() == "" {
		t.Error("Helper description should not be empty")
	}
	if helper.GetCategory() == "" {
		t.Error("Helper category should not be empty")
	}

	// Verify config schema
	schema := helper.GetConfigSchema()
	if schema == nil {
		t.Fatal("Config schema should not be nil")
	}

	// Verify required fields are in schema
	props, ok := schema["properties"].(map[string]interface{})
	if !ok {
		t.Fatal("Schema should have properties")
	}

	requiredFields := []string{"mode", "option_a", "option_b", "state_field"}
	for _, field := range requiredFields {
		if _, exists := props[field]; !exists {
			t.Errorf("Schema should define '%s' field", field)
		}
	}
}
