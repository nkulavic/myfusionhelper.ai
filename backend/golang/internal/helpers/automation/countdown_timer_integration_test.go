// +build integration

package automation

import (
	"testing"

	"github.com/myfusionhelper/api/internal/helpers"
)

// Test that countdown_timer is registered in the helper registry
func TestCountdownTimer_Registration(t *testing.T) {
	if !helpers.IsRegistered("countdown_timer") {
		t.Fatal("countdown_timer helper is not registered")
	}

	helper, err := helpers.NewHelper("countdown_timer")
	if err != nil {
		t.Fatalf("Failed to create countdown_timer helper: %v", err)
	}

	if helper == nil {
		t.Fatal("Helper instance is nil")
	}

	if helper.GetType() != "countdown_timer" {
		t.Errorf("Expected type 'countdown_timer', got '%s'", helper.GetType())
	}
}

// Test countdown_timer appears in helper listing
func TestCountdownTimer_InListing(t *testing.T) {
	helperTypes := helpers.ListHelperTypes()

	found := false
	for _, helperType := range helperTypes {
		if helperType == "countdown_timer" {
			found = true
			break
		}
	}

	if !found {
		t.Error("countdown_timer not found in helper type listing")
	}
}

// Test countdown_timer metadata in helper info
func TestCountdownTimer_HelperInfo(t *testing.T) {
	infos := helpers.ListHelperInfo()

	var countdownTimerInfo *helpers.HelperInfo
	for _, info := range infos {
		if info.Type == "countdown_timer" {
			countdownTimerInfo = &info
			break
		}
	}

	if countdownTimerInfo == nil {
		t.Fatal("countdown_timer not found in helper info listing")
	}

	if countdownTimerInfo.Name != "Countdown Timer" {
		t.Errorf("Expected name 'Countdown Timer', got '%s'", countdownTimerInfo.Name)
	}

	if countdownTimerInfo.Category != "automation" {
		t.Errorf("Expected category 'automation', got '%s'", countdownTimerInfo.Category)
	}

	if countdownTimerInfo.RequiresCRM {
		t.Error("Expected RequiresCRM to be false")
	}

	if countdownTimerInfo.ConfigSchema == nil {
		t.Error("Expected config schema to be non-nil")
	}
}
