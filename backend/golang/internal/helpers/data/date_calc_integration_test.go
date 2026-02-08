// +build integration

package data

import (
	"testing"

	"github.com/myfusionhelper/api/internal/helpers"
)

func TestDateCalcIntegration_Registration(t *testing.T) {
	// Verify date_calc is registered
	if !helpers.IsRegistered("date_calc") {
		t.Fatal("date_calc should be registered")
	}

	// Verify we can create a new instance
	helper, err := helpers.NewHelper("date_calc")
	if err != nil {
		t.Fatalf("NewHelper should not return error: %v", err)
	}
	if helper == nil {
		t.Fatal("NewHelper should return a date_calc instance")
	}

	// Verify it implements the Helper interface
	if _, ok := helper.(helpers.Helper); !ok {
		t.Error("date_calc should implement helpers.Helper interface")
	}

	// Verify type
	if helper.GetType() != "date_calc" {
		t.Errorf("Expected type 'date_calc', got '%s'", helper.GetType())
	}
}

func TestDateCalcIntegration_ListHelpers(t *testing.T) {
	// Get all registered helpers
	helpersInfo := helpers.ListHelperInfo()

	// Find date_calc in the list
	found := false
	for _, info := range helpersInfo {
		if info.Type == "date_calc" {
			found = true

			// Verify basic info
			if info.Name != "Date Calc" {
				t.Errorf("Expected name 'Date Calc', got '%s'", info.Name)
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
		t.Error("date_calc should be in the helpers list")
	}
}

func TestDateCalcIntegration_HelperInfo(t *testing.T) {
	helper, err := helpers.NewHelper("date_calc")
	if err != nil {
		t.Fatalf("Failed to create date_calc helper: %v", err)
	}
	if helper == nil {
		t.Fatal("Failed to create date_calc helper")
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

	requiredFields := []string{"operation", "field", "amount", "compare_field", "output_format", "target_field"}
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
	if len(enum) != 9 {
		t.Errorf("Expected 9 operations in enum, got %d", len(enum))
	}
}
