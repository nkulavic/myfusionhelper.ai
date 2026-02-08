// +build integration

package integration

import (
	"testing"

	"github.com/myfusionhelper/api/internal/helpers"
)

// TestTrelloItIntegration_Registration verifies trello_it is properly registered in the helper registry
func TestTrelloItIntegration_Registration(t *testing.T) {
	if !helpers.IsRegistered("trello_it") {
		t.Fatal("trello_it helper should be registered")
	}

	helper, err := helpers.NewHelper("trello_it")
	if err != nil {
		t.Fatalf("NewHelper should succeed for trello_it: %v", err)
	}

	if helper == nil {
		t.Fatal("trello_it helper factory should return a non-nil helper")
	}

	if helper.GetType() != "trello_it" {
		t.Errorf("Expected type 'trello_it', got '%s'", helper.GetType())
	}

	if helper.GetCategory() != "integration" {
		t.Errorf("Expected category 'integration', got '%s'", helper.GetCategory())
	}
}

// TestTrelloItIntegration_Listing verifies trello_it appears in the global helper list
func TestTrelloItIntegration_Listing(t *testing.T) {
	allHelpers := helpers.ListHelperInfo()

	found := false
	for _, h := range allHelpers {
		if h.Type == "trello_it" {
			found = true
			if h.Name != "Trello It" {
				t.Errorf("Expected name 'Trello It', got '%s'", h.Name)
			}
			if h.Category != "integration" {
				t.Errorf("Expected category 'integration', got '%s'", h.Category)
			}
			break
		}
	}

	if !found {
		t.Error("trello_it should appear in ListHelperInfo() output")
	}
}

// TestTrelloItIntegration_HelperInfo verifies ListHelperInfo returns complete metadata
func TestTrelloItIntegration_HelperInfo(t *testing.T) {
	allHelpers := helpers.ListHelperInfo()

	var info *helpers.HelperInfo
	for _, h := range allHelpers {
		if h.Type == "trello_it" {
			info = &h
			break
		}
	}

	if info == nil {
		t.Fatal("trello_it should be in ListHelperInfo() output")
	}

	if info.Type != "trello_it" {
		t.Errorf("Expected type 'trello_it', got '%s'", info.Type)
	}

	if info.Name != "Trello It" {
		t.Errorf("Expected name 'Trello It', got '%s'", info.Name)
	}

	if info.Category != "integration" {
		t.Errorf("Expected category 'integration', got '%s'", info.Category)
	}

	if info.Description == "" {
		t.Error("Expected non-empty description")
	}

	if info.ConfigSchema == nil {
		t.Error("Expected non-nil config schema")
	}

	if !info.RequiresCRM {
		t.Error("Expected RequiresCRM to be true")
	}
}
