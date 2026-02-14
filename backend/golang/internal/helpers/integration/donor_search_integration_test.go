// +build integration

package integration

import (
	"testing"

	"github.com/myfusionhelper/api/internal/helpers"
)

// Test that donor_search is registered in the helper registry
func TestDonorSearch_Registration(t *testing.T) {
	if !helpers.IsRegistered("donor_search") {
		t.Fatal("donor_search helper is not registered")
	}

	helper, err := helpers.NewHelper("donor_search")
	if err != nil {
		t.Fatalf("Failed to create donor_search helper: %v", err)
	}

	if helper == nil {
		t.Fatal("Helper instance is nil")
	}

	if helper.GetType() != "donor_search" {
		t.Errorf("Expected type 'donor_search', got '%s'", helper.GetType())
	}
}

// Test donor_search appears in helper listing
func TestDonorSearch_InListing(t *testing.T) {
	helperTypes := helpers.ListHelperTypes()

	found := false
	for _, helperType := range helperTypes {
		if helperType == "donor_search" {
			found = true
			break
		}
	}

	if !found {
		t.Error("donor_search not found in helper type listing")
	}
}

// Test donor_search metadata in helper info
func TestDonorSearch_HelperInfo(t *testing.T) {
	infos := helpers.ListHelperInfo()

	var donorSearchInfo *helpers.HelperInfo
	for _, info := range infos {
		if info.Type == "donor_search" {
			donorSearchInfo = &info
			break
		}
	}

	if donorSearchInfo == nil {
		t.Fatal("donor_search not found in helper info listing")
	}

	if donorSearchInfo.Name != "Donor Search" {
		t.Errorf("Expected name 'Donor Search', got '%s'", donorSearchInfo.Name)
	}

	if donorSearchInfo.Category != "integration" {
		t.Errorf("Expected category 'integration', got '%s'", donorSearchInfo.Category)
	}

	if !donorSearchInfo.RequiresCRM {
		t.Error("Expected RequiresCRM to be true")
	}

	if donorSearchInfo.ConfigSchema == nil {
		t.Error("Expected config schema to be non-nil")
	}
}
