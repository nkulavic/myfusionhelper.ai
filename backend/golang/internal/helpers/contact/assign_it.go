package contact

import (
	"context"
	"fmt"

	"github.com/myfusionhelper/api/internal/helpers"
)

// NewAssignIt creates a new AssignIt helper instance
func NewAssignIt() helpers.Helper { return &AssignIt{} }

func init() {
	helpers.Register("assign_it", func() helpers.Helper { return &AssignIt{} })
}

// AssignIt assigns a contact owner by setting an owner field
type AssignIt struct{}

func (h *AssignIt) GetName() string        { return "Assign It" }
func (h *AssignIt) GetType() string        { return "assign_it" }
func (h *AssignIt) GetCategory() string    { return "contact" }
func (h *AssignIt) GetDescription() string { return "Assign a contact owner by setting the owner field" }
func (h *AssignIt) RequiresCRM() bool      { return true }
func (h *AssignIt) SupportedCRMs() []string { return nil }

func (h *AssignIt) GetConfigSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"owner_id": map[string]interface{}{
				"type":        "string",
				"description": "The user/owner ID to assign the contact to",
			},
			"owner_field": map[string]interface{}{
				"type":        "string",
				"description": "The CRM field that stores the owner (defaults to 'owner_id')",
				"default":     "owner_id",
			},
		},
		"required": []string{"owner_id"},
	}
}

func (h *AssignIt) ValidateConfig(config map[string]interface{}) error {
	if _, ok := config["owner_id"].(string); !ok || config["owner_id"] == "" {
		return fmt.Errorf("owner_id is required")
	}
	return nil
}

func (h *AssignIt) Execute(ctx context.Context, input helpers.HelperInput) (*helpers.HelperOutput, error) {
	ownerID := input.Config["owner_id"].(string)
	ownerField := "owner_id"
	if f, ok := input.Config["owner_field"].(string); ok && f != "" {
		ownerField = f
	}

	output := &helpers.HelperOutput{
		Logs: make([]string, 0),
	}

	err := input.Connector.SetContactFieldValue(ctx, input.ContactID, ownerField, ownerID)
	if err != nil {
		output.Message = fmt.Sprintf("Failed to assign owner: %v", err)
		return output, err
	}

	output.Success = true
	output.Message = fmt.Sprintf("Assigned contact to owner %s", ownerID)
	output.Actions = []helpers.HelperAction{
		{
			Type:   "field_updated",
			Target: ownerField,
			Value:  ownerID,
		},
	}
	output.ModifiedData = map[string]interface{}{
		ownerField: ownerID,
	}
	output.Logs = append(output.Logs, fmt.Sprintf("Assigned contact %s to owner %s via field '%s'", input.ContactID, ownerID, ownerField))

	return output, nil
}
