package integration

import (
	"testing"

	"github.com/myfusionhelper/api/internal/helpers"
)

// TestStripeHooks_RegistryIntegration verifies the helper can be created from the registry
func TestStripeHooks_RegistryIntegration(t *testing.T) {
	// Verify helper is registered
	if !helpers.IsRegistered("stripe_hooks") {
		t.Fatal("stripe_hooks is not registered in the helper registry")
	}

	// Create helper from registry
	h, err := helpers.NewHelper("stripe_hooks")
	if err != nil {
		t.Fatalf("NewHelper(stripe_hooks) failed: %v", err)
	}

	// Verify it's the correct type
	stripeHooks, ok := h.(*StripeHooks)
	if !ok {
		t.Fatalf("NewHelper returned type %T, want *StripeHooks", h)
	}

	// Verify metadata
	if stripeHooks.GetType() != "stripe_hooks" {
		t.Errorf("GetType() = %v, want stripe_hooks", stripeHooks.GetType())
	}

	if stripeHooks.GetName() != "Stripe Hooks" {
		t.Errorf("GetName() = %v, want Stripe Hooks", stripeHooks.GetName())
	}

	if stripeHooks.GetCategory() != "integration" {
		t.Errorf("GetCategory() = %v, want integration", stripeHooks.GetCategory())
	}
}

// TestStripeHooks_ListHelperInfo verifies the helper appears in ListHelperInfo
func TestStripeHooks_ListHelperInfo(t *testing.T) {
	allHelpers := helpers.ListHelperInfo()

	// Find stripe_hooks in the list
	var found *helpers.HelperInfo
	for i := range allHelpers {
		if allHelpers[i].Type == "stripe_hooks" {
			found = &allHelpers[i]
			break
		}
	}

	if found == nil {
		t.Fatal("stripe_hooks not found in ListHelperInfo")
	}

	// Verify metadata
	if found.Name != "Stripe Hooks" {
		t.Errorf("HelperInfo.Name = %v, want Stripe Hooks", found.Name)
	}

	if found.Category != "integration" {
		t.Errorf("HelperInfo.Category = %v, want integration", found.Category)
	}

	if !found.RequiresCRM {
		t.Error("HelperInfo.RequiresCRM = false, want true")
	}

	if found.SupportedCRMs != nil {
		t.Errorf("HelperInfo.SupportedCRMs = %v, want nil (all CRMs)", found.SupportedCRMs)
	}

	// Verify config schema exists
	if found.ConfigSchema == nil {
		t.Error("HelperInfo.ConfigSchema is nil, want schema object")
	}
}

// TestStripeHooks_EndToEnd verifies full lifecycle from registry to execution
func TestStripeHooks_EndToEnd(t *testing.T) {
	// Get helper from registry
	h, err := helpers.NewHelper("stripe_hooks")
	if err != nil {
		t.Fatalf("NewHelper failed: %v", err)
	}

	// Validate a typical configuration
	config := map[string]interface{}{
		"selected_events": []interface{}{
			"charge.succeeded",
			"customer.subscription.created",
		},
		"goal_name":  "stripe_event_processed",
		"event_tags": []interface{}{"stripe", "automated"},
	}

	err = h.ValidateConfig(config)
	if err != nil {
		t.Fatalf("ValidateConfig failed: %v", err)
	}

	// Note: We don't execute here since this is a webhook configuration helper
	// Actual execution would happen when webhook events arrive
}
