package contact

import (
	"context"
	"fmt"

	"github.com/myfusionhelper/api/internal/helpers"
)

func init() {
	helpers.Register("own_it", func() helpers.Helper { return &OwnIt{} })
}

// OwnIt updates the owner (assigned user) of a contact in the CRM.
// Ported from legacy PHP own_it helper.
type OwnIt struct{}

func (h *OwnIt) GetName() string        { return "Own It" }
func (h *OwnIt) GetType() string        { return "own_it" }
func (h *OwnIt) GetCategory() string    { return "contact" }
func (h *OwnIt) GetDescription() string { return "Update the contact owner/assigned user ID" }
func (h *OwnIt) RequiresCRM() bool      { return true }
func (h *OwnIt) SupportedCRMs() []string { return nil }

func (h *OwnIt) GetConfigSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"owner_id": map[string]interface{}{
				"type":        "string",
				"description": "The user/owner ID to assign to the contact",
			},
		},
		"required": []string{"owner_id"},
	}
}

func (h *OwnIt) ValidateConfig(config map[string]interface{}) error {
	ownerID, ok := config["owner_id"].(string)
	if !ok || ownerID == "" {
		// Also accept numeric types
		if _, okNum := config["owner_id"].(float64); !okNum {
			return fmt.Errorf("owner_id is required")
		}
	}
	return nil
}

func (h *OwnIt) Execute(ctx context.Context, input helpers.HelperInput) (*helpers.HelperOutput, error) {
	ownerID := fmt.Sprintf("%v", input.Config["owner_id"])

	output := &helpers.HelperOutput{
		Logs: make([]string, 0),
	}

	if ownerID == "" || ownerID == "<nil>" {
		output.Message = "owner_id is empty, skipping"
		output.Success = true
		return output, nil
	}

	// Set the owner field on the contact
	err := input.Connector.SetContactFieldValue(ctx, input.ContactID, "OwnerID", ownerID)
	if err != nil {
		output.Message = fmt.Sprintf("Failed to update contact owner: %v", err)
		return output, err
	}

	output.Success = true
	output.Message = fmt.Sprintf("Contact owner updated to %s", ownerID)
	output.Actions = []helpers.HelperAction{
		{
			Type:   "field_updated",
			Target: "OwnerID",
			Value:  ownerID,
		},
	}
	output.ModifiedData = map[string]interface{}{
		"OwnerID": ownerID,
	}
	output.Logs = append(output.Logs, fmt.Sprintf("Updated owner of contact %s to user %s", input.ContactID, ownerID))

	return output, nil
}
