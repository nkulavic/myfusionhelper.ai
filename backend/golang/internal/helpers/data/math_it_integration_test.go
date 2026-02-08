// +build integration

package data

import (
	"testing"

	"github.com/myfusionhelper/api/internal/helpers"
)

// TestMathItIntegration_Registration verifies math_it is properly registered
func TestMathItIntegration_Registration(t *testing.T) {
	if !helpers.IsRegistered("math_it") {
		t.Fatal("math_it helper should be registered")
	}

	helper, err := helpers.NewHelper("math_it")
	if err != nil {
		t.Fatalf("NewHelper should succeed for math_it: %v", err)
	}

	if helper == nil {
		t.Fatal("math_it helper factory should return a non-nil helper")
	}

	if helper.GetType() != "math_it" {
		t.Errorf("Expected type 'math_it', got '%s'", helper.GetType())
	}

	if helper.GetCategory() != "data" {
		t.Errorf("Expected category 'data', got '%s'", helper.GetCategory())
	}
}

// TestMathItIntegration_Listing verifies math_it appears in the global helper list
func TestMathItIntegration_Listing(t *testing.T) {
	allHelpers := helpers.ListHelperInfo()

	found := false
	for _, h := range allHelpers {
		if h.Type == "math_it" {
			found = true
			if h.Name != "Math It" {
				t.Errorf("Expected name 'Math It', got '%s'", h.Name)
			}
			if h.Category != "data" {
				t.Errorf("Expected category 'data', got '%s'", h.Category)
			}
			break
		}
	}

	if !found {
		t.Error("math_it should appear in ListHelperInfo() output")
	}
}

// TestMathItIntegration_HelperInfo verifies ListHelperInfo returns complete metadata
func TestMathItIntegration_HelperInfo(t *testing.T) {
	allHelpers := helpers.ListHelperInfo()

	var info *helpers.HelperInfo
	for _, h := range allHelpers {
		if h.Type == "math_it" {
			info = &h
			break
		}
	}

	if info == nil {
		t.Fatal("math_it should be in ListHelperInfo() output")
	}

	if info.Type != "math_it" {
		t.Errorf("Expected type 'math_it', got '%s'", info.Type)
	}

	if info.Name != "Math It" {
		t.Errorf("Expected name 'Math It', got '%s'", info.Name)
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
