// +build integration

package data

import (
	"testing"

	"github.com/myfusionhelper/api/internal/helpers"
)

func TestWordCountItIntegration_Registration(t *testing.T) {
	if !helpers.IsRegistered("word_count_it") {
		t.Fatal("word_count_it should be registered")
	}

	helper, err := helpers.NewHelper("word_count_it")
	if err != nil {
		t.Fatalf("NewHelper should not return error: %v", err)
	}
	if helper == nil {
		t.Fatal("NewHelper should return a word_count_it instance")
	}

	if helper.GetType() != "word_count_it" {
		t.Errorf("Expected type 'word_count_it', got '%s'", helper.GetType())
	}
}

func TestWordCountItIntegration_ListHelpers(t *testing.T) {
	helpersInfo := helpers.ListHelperInfo()

	found := false
	for _, info := range helpersInfo {
		if info.Type == "word_count_it" {
			found = true
			if info.Name != "Word Count It" {
				t.Errorf("Expected name 'Word Count It', got '%s'", info.Name)
			}
			if info.Category != "data" {
				t.Errorf("Expected category 'data', got '%s'", info.Category)
			}
			break
		}
	}

	if !found {
		t.Error("word_count_it should be in the helpers list")
	}
}

func TestWordCountItIntegration_HelperInfo(t *testing.T) {
	helper, err := helpers.NewHelper("word_count_it")
	if err != nil {
		t.Fatalf("Failed to create word_count_it helper: %v", err)
	}

	schema := helper.GetConfigSchema()
	if schema == nil {
		t.Fatal("Config schema should not be nil")
	}

	props, ok := schema["properties"].(map[string]interface{})
	if !ok {
		t.Fatal("Schema should have properties")
	}

	requiredFields := []string{"source_field", "target_field", "count_type"}
	for _, field := range requiredFields {
		if _, exists := props[field]; !exists {
			t.Errorf("Schema should define '%s' field", field)
		}
	}
}
