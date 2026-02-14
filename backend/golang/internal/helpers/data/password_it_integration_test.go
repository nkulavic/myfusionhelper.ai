// +build integration

package data

import (
	"testing"

	"github.com/myfusionhelper/api/internal/helpers"
)

func TestPasswordItIntegration_Registration(t *testing.T) {
	if !helpers.IsRegistered("password_it") {
		t.Fatal("password_it should be registered")
	}

	helper, err := helpers.NewHelper("password_it")
	if err != nil {
		t.Fatalf("NewHelper should not return error: %v", err)
	}
	if helper.GetType() != "password_it" {
		t.Errorf("Expected type 'password_it', got '%s'", helper.GetType())
	}
}

func TestPasswordItIntegration_ListHelpers(t *testing.T) {
	helpersInfo := helpers.ListHelperInfo()

	found := false
	for _, info := range helpersInfo {
		if info.Type == "password_it" {
			found = true
			if info.Category != "data" {
				t.Errorf("Expected category 'data', got '%s'", info.Category)
			}
			break
		}
	}

	if !found {
		t.Error("password_it should be in the helpers list")
	}
}

func TestPasswordItIntegration_HelperInfo(t *testing.T) {
	helper, err := helpers.NewHelper("password_it")
	if err != nil {
		t.Fatalf("Failed to create password_it helper: %v", err)
	}

	schema := helper.GetConfigSchema()
	if schema == nil {
		t.Fatal("Config schema should not be nil")
	}

	props, ok := schema["properties"].(map[string]interface{})
	if !ok {
		t.Fatal("Schema should have properties")
	}

	if _, exists := props["target_field"]; !exists {
		t.Error("Schema should define 'target_field' field")
	}
}
