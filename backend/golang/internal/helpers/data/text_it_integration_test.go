// +build integration

package data

import (
	"testing"

	"github.com/myfusionhelper/api/internal/helpers"
)

func TestTextItIntegration_Registration(t *testing.T) {
	// Verify text_it is registered
	if !helpers.IsRegistered("text_it") {
		t.Fatal("text_it should be registered")
	}

	// Verify we can create a new instance
	helper, err := helpers.NewHelper("text_it")
	if err != nil {
		t.Fatalf("NewHelper should not return error: %v", err)
	}
	if helper == nil {
		t.Fatal("NewHelper should return a text_it instance")
	}

	// Verify it implements the Helper interface
	if _, ok := helper.(helpers.Helper); !ok {
		t.Error("text_it should implement helpers.Helper interface")
	}

	// Verify type
	if helper.GetType() != "text_it" {
		t.Errorf("Expected type 'text_it', got '%s'", helper.GetType())
	}
}

func TestTextItIntegration_ListHelpers(t *testing.T) {
	// Get all registered helpers
	helpersInfo := helpers.ListHelperInfo()

	// Find text_it in the list
	found := false
	for _, info := range helpersInfo {
		if info.Type == "text_it" {
			found = true

			// Verify basic info
			if info.Name != "Text It" {
				t.Errorf("Expected name 'Text It', got '%s'", info.Name)
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
		t.Error("text_it should be in the helpers list")
	}
}

func TestTextItIntegration_HelperInfo(t *testing.T) {
	helper, err := helpers.NewHelper("text_it")
	if err != nil {
		t.Fatalf("Failed to create text_it helper: %v", err)
	}
	if helper == nil {
		t.Fatal("Failed to create text_it helper")
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

	requiredFields := []string{"field", "operation"}
	for _, field := range requiredFields {
		if _, exists := props[field]; !exists {
			t.Errorf("Schema should define '%s' field", field)
		}
	}

	// Verify operation field has enum
	opField, ok := props["operation"].(map[string]interface{})
	if !ok {
		t.Fatal("operation field should be defined")
	}
	enum, ok := opField["enum"].([]string)
	if !ok {
		t.Fatal("operation field should have enum")
	}
	if len(enum) == 0 {
		t.Error("operation enum should not be empty")
	}
}
