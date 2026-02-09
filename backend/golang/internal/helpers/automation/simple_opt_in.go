package automation

import (
	"context"
	"fmt"

	"github.com/myfusionhelper/api/internal/helpers"
)

func init() {
	helpers.Register("simple_opt_in", func() helpers.Helper { return &SimpleOptIn{} })
}

// SimpleOptIn sets the marketable status for a contact (email opt-in automation)
type SimpleOptIn struct{}

func (h *SimpleOptIn) GetName() string        { return "Simple Opt-In" }
func (h *SimpleOptIn) GetType() string        { return "simple_opt_in" }
func (h *SimpleOptIn) GetCategory() string    { return "automation" }
func (h *SimpleOptIn) GetDescription() string { return "Set marketable status for email opt-in" }
func (h *SimpleOptIn) RequiresCRM() bool      { return true }
func (h *SimpleOptIn) SupportedCRMs() []string { return nil } // All CRMs

func (h *SimpleOptIn) GetConfigSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"opt_in": map[string]interface{}{
				"type":        "boolean",
				"description": "Set to true to opt in, false to opt out",
				"default":     true,
			},
			"reason": map[string]interface{}{
				"type":        "string",
				"description": "Reason for opt-in/opt-out (optional)",
			},
		},
		"required": []string{"opt_in"},
	}
}

func (h *SimpleOptIn) ValidateConfig(config map[string]interface{}) error {
	if _, ok := config["opt_in"].(bool); !ok {
		return fmt.Errorf("opt_in must be a boolean")
	}
	return nil
}

func (h *SimpleOptIn) Execute(ctx context.Context, input helpers.HelperInput) (*helpers.HelperOutput, error) {
	optIn, _ := input.Config["opt_in"].(bool)
	reason := ""
	if r, ok := input.Config["reason"].(string); ok {
		reason = r
	}

	output := &helpers.HelperOutput{
		Logs: make([]string, 0),
	}

	err := input.Connector.SetOptInStatus(ctx, input.ContactID, optIn, reason)
	if err != nil {
		output.Message = fmt.Sprintf("Failed to set opt-in status: %v", err)
		output.Logs = append(output.Logs, output.Message)
		return output, err
	}

	status := "opted out"
	if optIn {
		status = "opted in"
	}

	output.Success = true
	output.Message = fmt.Sprintf("Contact %s for email marketing", status)
	output.Actions = []helpers.HelperAction{
		{
			Type:   "opt_in_status_changed",
			Target: input.ContactID,
			Value:  status,
		},
	}

	logMsg := fmt.Sprintf("Set opt-in status to %v for contact %s", optIn, input.ContactID)
	if reason != "" {
		logMsg += fmt.Sprintf(" (reason: %s)", reason)
	}
	output.Logs = append(output.Logs, logMsg)

	return output, nil
}
