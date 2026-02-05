package contact

import (
	"context"
	"fmt"

	"github.com/myfusionhelper/api/internal/helpers"
)

func init() {
	helpers.Register("company_link", func() helpers.Helper { return &CompanyLink{} })
}

// CompanyLink links a contact to a company based on a field value
type CompanyLink struct{}

func (h *CompanyLink) GetName() string        { return "Company Link" }
func (h *CompanyLink) GetType() string        { return "company_link" }
func (h *CompanyLink) GetCategory() string    { return "contact" }
func (h *CompanyLink) GetDescription() string { return "Link a contact to a company based on a field value" }
func (h *CompanyLink) RequiresCRM() bool      { return true }
func (h *CompanyLink) SupportedCRMs() []string { return []string{"keap"} }

func (h *CompanyLink) GetConfigSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"company_field": map[string]interface{}{
				"type":        "string",
				"description": "The field containing the company name to link",
			},
		},
		"required": []string{"company_field"},
	}
}

func (h *CompanyLink) ValidateConfig(config map[string]interface{}) error {
	if _, ok := config["company_field"].(string); !ok || config["company_field"] == "" {
		return fmt.Errorf("company_field is required")
	}
	return nil
}

func (h *CompanyLink) Execute(ctx context.Context, input helpers.HelperInput) (*helpers.HelperOutput, error) {
	companyField := input.Config["company_field"].(string)

	output := &helpers.HelperOutput{
		Logs: make([]string, 0),
	}

	// Get the company name from the contact field
	companyValue, err := input.Connector.GetContactFieldValue(ctx, input.ContactID, companyField)
	if err != nil {
		output.Message = fmt.Sprintf("Failed to read company field '%s': %v", companyField, err)
		return output, err
	}

	companyName := fmt.Sprintf("%v", companyValue)
	if companyValue == nil || companyName == "" || companyName == "<nil>" {
		output.Success = true
		output.Message = fmt.Sprintf("Company field '%s' is empty, nothing to link", companyField)
		output.Logs = append(output.Logs, output.Message)
		return output, nil
	}

	// Set the company field on the contact (triggers CRM-side company association)
	err = input.Connector.SetContactFieldValue(ctx, input.ContactID, "company", companyName)
	if err != nil {
		output.Message = fmt.Sprintf("Failed to link company '%s': %v", companyName, err)
		return output, err
	}

	output.Success = true
	output.Message = fmt.Sprintf("Linked contact to company '%s'", companyName)
	output.Actions = []helpers.HelperAction{
		{
			Type:   "company_linked",
			Target: input.ContactID,
			Value:  companyName,
		},
	}
	output.ModifiedData = map[string]interface{}{
		"company": companyName,
	}
	output.Logs = append(output.Logs, fmt.Sprintf("Linked contact %s to company '%s' via field '%s'", input.ContactID, companyName, companyField))

	return output, nil
}
