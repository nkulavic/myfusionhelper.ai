// +build integration

package data

import (
	"testing"

	"github.com/myfusionhelper/api/internal/helpers"
)

func TestGetTheFirstIntegration_Registration(t *testing.T) {
	// Verify get_the_first is registered
	if !helpers.IsRegistered("get_the_first") {
		t.Fatal("get_the_first should be registered")
	}

	// Verify we can create a new instance
	helper, err := helpers.NewHelper("get_the_first")
	if err != nil {
		t.Fatalf("NewHelper should not return error: %v", err)
	}
	if helper == nil {
		t.Fatal("NewHelper should return a get_the_first instance")
	}

	// Verify it implements the Helper interface
	if _, ok := helper.(helpers.Helper); !ok {
		t.Error("get_the_first should implement helpers.Helper interface")
	}

	// Verify type
	if helper.GetType() != "get_the_first" {
		t.Errorf("Expected type 'get_the_first', got '%s'", helper.GetType())
	}
}

func TestGetTheFirstIntegration_ListHelpers(t *testing.T) {
	// Get all registered helpers
	helpersInfo := helpers.ListHelperInfo()

	// Find get_the_first in the list
	found := false
	for _, info := range helpersInfo {
		if info.Type == "get_the_first" {
			found = true

			// Verify basic info
			if info.Name != "Get The First" {
				t.Errorf("Expected name 'Get The First', got '%s'", info.Name)
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
		t.Error("get_the_first should be in the helpers list")
	}
}

func TestGetTheFirstIntegration_HelperInfo(t *testing.T) {
	helper, err := helpers.NewHelper("get_the_first")
	if err != nil {
		t.Fatalf("Failed to create get_the_first helper: %v", err)
	}
	if helper == nil {
		t.Fatal("Failed to create get_the_first helper")
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

	requiredFields := []string{"type", "from_field", "to_field"}
	for _, field := range requiredFields {
		if _, exists := props[field]; !exists {
			t.Errorf("Schema should define '%s' field", field)
		}
	}

	// Verify type field has enum
	typeField, ok := props["type"].(map[string]interface{})
	if !ok {
		t.Fatal("type field should be defined")
	}
	enum, ok := typeField["enum"].([]string)
	if !ok {
		t.Fatal("type field should have enum")
	}
	if len(enum) != 6 {
		t.Errorf("Expected 6 record types in enum, got %d", len(enum))
	}
}
