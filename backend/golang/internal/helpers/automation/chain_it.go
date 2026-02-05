package automation

import (
	"context"
	"fmt"

	"github.com/myfusionhelper/api/internal/helpers"
)

func init() {
	helpers.Register("chain_it", func() helpers.Helper { return &ChainIt{} })
}

// ChainIt chains multiple helper executions together.
// Returns output with actions listing helpers to chain (actual chaining done by execution layer).
type ChainIt struct{}

func (h *ChainIt) GetName() string        { return "Chain It" }
func (h *ChainIt) GetType() string        { return "chain_it" }
func (h *ChainIt) GetCategory() string    { return "automation" }
func (h *ChainIt) GetDescription() string { return "Chain multiple helper executions in sequence" }
func (h *ChainIt) RequiresCRM() bool      { return false }
func (h *ChainIt) SupportedCRMs() []string { return nil }

func (h *ChainIt) GetConfigSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"helpers": map[string]interface{}{
				"type":        "array",
				"items":       map[string]interface{}{"type": "string"},
				"description": "List of helper type strings to chain in sequence",
			},
		},
		"required": []string{"helpers"},
	}
}

func (h *ChainIt) ValidateConfig(config map[string]interface{}) error {
	helpersList, ok := config["helpers"]
	if !ok {
		return fmt.Errorf("helpers is required")
	}

	switch v := helpersList.(type) {
	case []interface{}:
		if len(v) == 0 {
			return fmt.Errorf("helpers must contain at least one helper type")
		}
	case []string:
		if len(v) == 0 {
			return fmt.Errorf("helpers must contain at least one helper type")
		}
	default:
		return fmt.Errorf("helpers must be an array of strings")
	}

	return nil
}

func (h *ChainIt) Execute(ctx context.Context, input helpers.HelperInput) (*helpers.HelperOutput, error) {
	helperTypes := extractStringSlice(input.Config["helpers"])

	output := &helpers.HelperOutput{
		Actions: make([]helpers.HelperAction, 0),
		Logs:    make([]string, 0),
	}

	// Build chain actions for the execution layer to process
	for i, helperType := range helperTypes {
		output.Actions = append(output.Actions, helpers.HelperAction{
			Type:   "helper_chain",
			Target: helperType,
			Value:  i,
		})
		output.Logs = append(output.Logs, fmt.Sprintf("Chained helper %d: %s", i+1, helperType))
	}

	output.Success = true
	output.Message = fmt.Sprintf("Chained %d helper(s) for sequential execution", len(helperTypes))

	return output, nil
}
