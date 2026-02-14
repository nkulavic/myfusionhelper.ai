// +build integration

package data

import (
	"testing"

	"github.com/myfusionhelper/api/internal/helpers"
)

// TestFormatItIntegration_Registration verifies format_it is properly registered
func TestFormatItIntegration_Registration(t *testing.T) {
	if !helpers.IsRegistered("format_it") {
		t.Fatal("format_it helper should be registered")
	}

	helper, err := helpers.NewHelper("format_it")
	if err != nil {
		t.Fatalf("NewHelper should succeed for format_it: %v", err)
	}

	if helper == nil {
		t.Fatal("format_it helper factory should return a non-nil helper")
	}

	if helper.GetType() != "format_it" {
		t.Errorf("Expected type 'format_it', got '%s'", helper.GetType())
	}

	if helper.GetCategory() != "data" {
		t.Errorf("Expected category 'data', got '%s'", helper.GetCategory())
	}
}

// TestFormatItIntegration_Listing verifies format_it appears in the global helper list
func TestFormatItIntegration_Listing(t *testing.T) {
	allHelpers := helpers.ListHelperInfo()

	found := false
	for _, h := range allHelpers {
		if h.Type == "format_it" {
			found = true
			if h.Name != "Format It" {
				t.Errorf("Expected name 'Format It', got '%s'", h.Name)
			}
			if h.Category != "data" {
				t.Errorf("Expected category 'data', got '%s'", h.Category)
			}
			break
		}
	}

	if !found {
		t.Error("format_it should appear in ListHelperInfo() output")
	}
}

// TestFormatItIntegration_HelperInfo verifies ListHelperInfo returns complete metadata
func TestFormatItIntegration_HelperInfo(t *testing.T) {
	allHelpers := helpers.ListHelperInfo()

	var info *helpers.HelperInfo
	for _, h := range allHelpers {
		if h.Type == "format_it" {
			info = &h
			break
		}
	}

	if info == nil {
		t.Fatal("format_it should be in ListHelperInfo() output")
	}

	if info.Type != "format_it" {
		t.Errorf("Expected type 'format_it', got '%s'", info.Type)
	}

	if info.Name != "Format It" {
		t.Errorf("Expected name 'Format It', got '%s'", info.Name)
	}

	if info.Category != "data" {
		t.Errorf("Expected category 'data', got '%s'", info.Category)
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
