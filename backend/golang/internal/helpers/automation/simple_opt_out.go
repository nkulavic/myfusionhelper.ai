package automation

import (
	"context"
	"fmt"

	"github.com/myfusionhelper/api/internal/helpers"
)

func init() {
	helpers.Register("simple_opt_out", func() helpers.Helper { return &SimpleOptOut{} })
}

// SimpleOptOut removes the marketable status for a contact (email opt-out automation)
type SimpleOptOut struct{}

func (h *SimpleOptOut) GetName() string        { return "Simple Opt-Out" }
func (h *SimpleOptOut) GetType() string        { return "simple_opt_out" }
func (h *SimpleOptOut) GetCategory() string    { return "automation" }
func (h *SimpleOptOut) GetDescription() string { return "Remove marketable status for email opt-out" }
func (h *SimpleOptOut) RequiresCRM() bool      { return true }
func (h *SimpleOptOut) SupportedCRMs() []string { return nil } // All CRMs

func (h *SimpleOptOut) GetConfigSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"reason": map[string]interface{}{
				"type":        "string",
				"description": "Reason for opt-out (optional)",
			},
		},
	}
}

func (h *SimpleOptOut) ValidateConfig(config map[string]interface{}) error {
	// No required fields - reason is optional
	return nil
}

func (h *SimpleOptOut) Execute(ctx context.Context, input helpers.HelperInput) (*helpers.HelperOutput, error) {
	reason := ""
	if r, ok := input.Config["reason"].(string); ok {
		reason = r
	}

	output := &helpers.HelperOutput{
		Logs: make([]string, 0),
	}

	// Simple opt-out always sets optIn to false
	err := input.Connector.SetOptInStatus(ctx, input.ContactID, false, reason)
	if err != nil {
		output.Message = fmt.Sprintf("Failed to opt out contact: %v", err)
		output.Logs = append(output.Logs, output.Message)
		return output, err
	}

	output.Success = true
	output.Message = "Contact opted out from email marketing"
	output.Actions = []helpers.HelperAction{
		{
			Type:   "opt_out",
			Target: input.ContactID,
			Value:  "opted_out",
		},
	}

	logMsg := fmt.Sprintf("Opted out contact %s from email marketing", input.ContactID)
	if reason != "" {
		logMsg += fmt.Sprintf(" (reason: %s)", reason)
	}
	output.Logs = append(output.Logs, logMsg)

	return output, nil
}
