// +build integration

package data

import (
	"testing"

	"github.com/myfusionhelper/api/internal/helpers"
)

func TestGetTheLastIntegration_Registration(t *testing.T) {
	if !helpers.IsRegistered("get_the_last") {
		t.Fatal("get_the_last should be registered")
	}

	helper, err := helpers.NewHelper("get_the_last")
	if err != nil {
		t.Fatalf("NewHelper should not return error: %v", err)
	}
	if helper == nil {
		t.Fatal("NewHelper should return a get_the_last instance")
	}

	if _, ok := helper.(helpers.Helper); !ok {
		t.Error("get_the_last should implement helpers.Helper interface")
	}

	if helper.GetType() != "get_the_last" {
		t.Errorf("Expected type 'get_the_last', got '%s'", helper.GetType())
	}
}

func TestGetTheLastIntegration_ListHelpers(t *testing.T) {
	helpersInfo := helpers.ListHelperInfo()

	found := false
	for _, info := range helpersInfo {
		if info.Type == "get_the_last" {
			found = true

			if info.Name != "Get The Last" {
				t.Errorf("Expected name 'Get The Last', got '%s'", info.Name)
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
		t.Error("get_the_last should be in the helpers list")
	}
}

func TestGetTheLastIntegration_HelperInfo(t *testing.T) {
	helper, err := helpers.NewHelper("get_the_last")
	if err != nil {
		t.Fatalf("Failed to create get_the_last helper: %v", err)
	}
	if helper == nil {
		t.Fatal("Failed to create get_the_last helper")
	}

	if helper.GetName() == "" {
		t.Error("Helper name should not be empty")
	}
	if helper.GetDescription() == "" {
		t.Error("Helper description should not be empty")
	}
	if helper.GetCategory() == "" {
		t.Error("Helper category should not be empty")
	}

	schema := helper.GetConfigSchema()
	if schema == nil {
		t.Fatal("Config schema should not be nil")
	}

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
