// +build integration

package integration

import (
	"testing"

	"github.com/myfusionhelper/api/internal/helpers"
)

// TestOrderItIntegration_Registration verifies order_it is properly registered in the helper registry
func TestOrderItIntegration_Registration(t *testing.T) {
	if !helpers.IsRegistered("order_it") {
		t.Fatal("order_it helper should be registered")
	}

	helper, err := helpers.NewHelper("order_it")
	if err != nil {
		t.Fatalf("NewHelper should succeed for order_it: %v", err)
	}

	if helper == nil {
		t.Fatal("order_it helper factory should return a non-nil helper")
	}

	if helper.GetType() != "order_it" {
		t.Errorf("Expected type 'order_it', got '%s'", helper.GetType())
	}

	if helper.GetCategory() != "integration" {
		t.Errorf("Expected category 'integration', got '%s'", helper.GetCategory())
	}
}

// TestOrderItIntegration_Listing verifies order_it appears in the global helper list
func TestOrderItIntegration_Listing(t *testing.T) {
	allHelpers := helpers.ListHelperInfo()

	found := false
	for _, h := range allHelpers {
		if h.Type == "order_it" {
			found = true
			if h.Name != "Order It" {
				t.Errorf("Expected name 'Order It', got '%s'", h.Name)
			}
			if h.Category != "integration" {
				t.Errorf("Expected category 'integration', got '%s'", h.Category)
			}
			break
		}
	}

	if !found {
		t.Error("order_it should appear in ListHelperInfo() output")
	}
}

// TestOrderItIntegration_HelperInfo verifies ListHelperInfo returns complete metadata
func TestOrderItIntegration_HelperInfo(t *testing.T) {
	allHelpers := helpers.ListHelperInfo()

	var info *helpers.HelperInfo
	for _, h := range allHelpers {
		if h.Type == "order_it" {
			info = &h
			break
		}
	}

	if info == nil {
		t.Fatal("order_it should be in ListHelperInfo() output")
	}

	if info.Type != "order_it" {
		t.Errorf("Expected type 'order_it', got '%s'", info.Type)
	}

	if info.Name != "Order It" {
		t.Errorf("Expected name 'Order It', got '%s'", info.Name)
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

	if len(info.SupportedCRMs) != 1 || info.SupportedCRMs[0] != "keap" {
		t.Errorf("Expected SupportedCRMs to be ['keap'], got %v", info.SupportedCRMs)
	}
}
