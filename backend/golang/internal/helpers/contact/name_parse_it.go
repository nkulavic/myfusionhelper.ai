package contact

import (
	"context"
	"fmt"
	"strings"

	"github.com/myfusionhelper/api/internal/helpers"
)

func init() {
	helpers.Register("name_parse_it", func() helpers.Helper { return &NameParseIt{} })
}

// NameParseIt parses a full name into first/last/suffix components
type NameParseIt struct{}

func (h *NameParseIt) GetName() string        { return "Name Parse It" }
func (h *NameParseIt) GetType() string        { return "name_parse_it" }
func (h *NameParseIt) GetCategory() string    { return "contact" }
func (h *NameParseIt) GetDescription() string { return "Parse a full name into first, last, and suffix components" }
func (h *NameParseIt) RequiresCRM() bool      { return true }
func (h *NameParseIt) SupportedCRMs() []string { return nil }

func (h *NameParseIt) GetConfigSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"source_field": map[string]interface{}{
				"type":        "string",
				"description": "The field containing the full name",
			},
			"first_name_field": map[string]interface{}{
				"type":        "string",
				"description": "Field to store parsed first name",
				"default":     "first_name",
			},
			"last_name_field": map[string]interface{}{
				"type":        "string",
				"description": "Field to store parsed last name",
				"default":     "last_name",
			},
			"suffix_field": map[string]interface{}{
				"type":        "string",
				"description": "Optional field to store parsed suffix (Jr, Sr, III, etc.)",
			},
		},
		"required": []string{"source_field"},
	}
}

func (h *NameParseIt) ValidateConfig(config map[string]interface{}) error {
	if _, ok := config["source_field"].(string); !ok || config["source_field"] == "" {
		return fmt.Errorf("source_field is required")
	}
	return nil
}

func (h *NameParseIt) Execute(ctx context.Context, input helpers.HelperInput) (*helpers.HelperOutput, error) {
	sourceField := input.Config["source_field"].(string)
	firstNameField := "first_name"
	if f, ok := input.Config["first_name_field"].(string); ok && f != "" {
		firstNameField = f
	}
	lastNameField := "last_name"
	if f, ok := input.Config["last_name_field"].(string); ok && f != "" {
		lastNameField = f
	}
	suffixField := ""
	if f, ok := input.Config["suffix_field"].(string); ok {
		suffixField = f
	}

	output := &helpers.HelperOutput{
		Actions:      make([]helpers.HelperAction, 0),
		ModifiedData: make(map[string]interface{}),
		Logs:         make([]string, 0),
	}

	// Get the full name value
	rawValue, err := input.Connector.GetContactFieldValue(ctx, input.ContactID, sourceField)
	if err != nil {
		output.Message = fmt.Sprintf("Failed to read field '%s': %v", sourceField, err)
		return output, err
	}

	fullName := strings.TrimSpace(fmt.Sprintf("%v", rawValue))
	if rawValue == nil || fullName == "" || fullName == "<nil>" {
		output.Success = true
		output.Message = fmt.Sprintf("Field '%s' is empty, nothing to parse", sourceField)
		return output, nil
	}

	// Parse name parts
	parts := strings.Fields(fullName)
	firstName := ""
	lastName := ""
	suffix := ""

	// Known suffixes
	suffixes := map[string]bool{
		"jr": true, "jr.": true, "sr": true, "sr.": true,
		"ii": true, "iii": true, "iv": true, "v": true,
		"phd": true, "md": true, "esq": true, "dds": true,
	}

	// Check last part for suffix
	if len(parts) > 2 {
		lastPart := strings.ToLower(parts[len(parts)-1])
		if suffixes[lastPart] {
			suffix = parts[len(parts)-1]
			parts = parts[:len(parts)-1]
		}
	}

	switch len(parts) {
	case 0:
		// Nothing to do
	case 1:
		firstName = parts[0]
	default:
		firstName = parts[0]
		lastName = strings.Join(parts[1:], " ")
	}

	// Set first name
	err = input.Connector.SetContactFieldValue(ctx, input.ContactID, firstNameField, firstName)
	if err != nil {
		output.Logs = append(output.Logs, fmt.Sprintf("Failed to set first name: %v", err))
	} else {
		output.Actions = append(output.Actions, helpers.HelperAction{
			Type:   "field_updated",
			Target: firstNameField,
			Value:  firstName,
		})
		output.ModifiedData[firstNameField] = firstName
	}

	// Set last name
	err = input.Connector.SetContactFieldValue(ctx, input.ContactID, lastNameField, lastName)
	if err != nil {
		output.Logs = append(output.Logs, fmt.Sprintf("Failed to set last name: %v", err))
	} else {
		output.Actions = append(output.Actions, helpers.HelperAction{
			Type:   "field_updated",
			Target: lastNameField,
			Value:  lastName,
		})
		output.ModifiedData[lastNameField] = lastName
	}

	// Set suffix if configured and present
	if suffixField != "" && suffix != "" {
		err = input.Connector.SetContactFieldValue(ctx, input.ContactID, suffixField, suffix)
		if err != nil {
			output.Logs = append(output.Logs, fmt.Sprintf("Failed to set suffix: %v", err))
		} else {
			output.Actions = append(output.Actions, helpers.HelperAction{
				Type:   "field_updated",
				Target: suffixField,
				Value:  suffix,
			})
			output.ModifiedData[suffixField] = suffix
		}
	}

	output.Success = len(output.Actions) > 0
	output.Message = fmt.Sprintf("Parsed name '%s' into first='%s', last='%s'", fullName, firstName, lastName)
	if suffix != "" {
		output.Message += fmt.Sprintf(", suffix='%s'", suffix)
	}
	output.Logs = append(output.Logs, output.Message)

	return output, nil
}
