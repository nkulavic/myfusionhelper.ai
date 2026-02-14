package data

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"

	"github.com/myfusionhelper/api/internal/helpers"
)

// NewPasswordIt creates a new PasswordIt helper instance
func NewPasswordIt() helpers.Helper { return &PasswordIt{} }

func init() {
	helpers.Register("password_it", func() helpers.Helper { return &PasswordIt{} })
}

// PasswordIt generates a random password and stores it in a contact field
type PasswordIt struct{}

func (h *PasswordIt) GetName() string        { return "Password It" }
func (h *PasswordIt) GetType() string        { return "password_it" }
func (h *PasswordIt) GetCategory() string    { return "data" }
func (h *PasswordIt) GetDescription() string { return "Generate a random password and store it in a contact field" }
func (h *PasswordIt) RequiresCRM() bool      { return true }
func (h *PasswordIt) SupportedCRMs() []string { return nil }

func (h *PasswordIt) GetConfigSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"target_field": map[string]interface{}{
				"type":        "string",
				"description": "The field to store the generated password",
			},
			"length": map[string]interface{}{
				"type":        "integer",
				"description": "Password length",
				"default":     12,
			},
			"include_special": map[string]interface{}{
				"type":        "boolean",
				"description": "Whether to include special characters",
				"default":     true,
			},
			"overwrite": map[string]interface{}{
				"type":        "boolean",
				"description": "Whether to overwrite existing field value",
				"default":     false,
			},
		},
		"required": []string{"target_field"},
	}
}

func (h *PasswordIt) ValidateConfig(config map[string]interface{}) error {
	if _, ok := config["target_field"].(string); !ok || config["target_field"] == "" {
		return fmt.Errorf("target_field is required")
	}
	return nil
}

func (h *PasswordIt) Execute(ctx context.Context, input helpers.HelperInput) (*helpers.HelperOutput, error) {
	targetField := input.Config["target_field"].(string)
	length := 12
	if l, ok := input.Config["length"].(float64); ok && l > 0 {
		length = int(l)
	}
	includeSpecial := true
	if is, ok := input.Config["include_special"].(bool); ok {
		includeSpecial = is
	}
	overwrite := false
	if ow, ok := input.Config["overwrite"].(bool); ok {
		overwrite = ow
	}

	output := &helpers.HelperOutput{
		Logs: make([]string, 0),
	}

	// Check if field already has a value
	if !overwrite {
		existing, err := input.Connector.GetContactFieldValue(ctx, input.ContactID, targetField)
		if err == nil && existing != nil && existing != "" {
			strVal := fmt.Sprintf("%v", existing)
			if strVal != "" && strVal != "<nil>" {
				output.Success = true
				output.Message = fmt.Sprintf("Field '%s' already has a value, skipping (overwrite=false)", targetField)
				output.Logs = append(output.Logs, output.Message)
				return output, nil
			}
		}
	}

	// Generate password
	password, err := generatePassword(length, includeSpecial)
	if err != nil {
		output.Message = fmt.Sprintf("Failed to generate password: %v", err)
		return output, err
	}

	// Set field value
	err = input.Connector.SetContactFieldValue(ctx, input.ContactID, targetField, password)
	if err != nil {
		output.Message = fmt.Sprintf("Failed to set password field '%s': %v", targetField, err)
		return output, err
	}

	output.Success = true
	output.Message = fmt.Sprintf("Generated %d-character password in '%s'", length, targetField)
	output.Actions = []helpers.HelperAction{
		{
			Type:   "field_updated",
			Target: targetField,
			Value:  password,
		},
	}
	output.ModifiedData = map[string]interface{}{
		targetField: password,
	}
	output.Logs = append(output.Logs, fmt.Sprintf("Generated password of length %d for contact %s in field '%s'", length, input.ContactID, targetField))

	return output, nil
}

func generatePassword(length int, includeSpecial bool) (string, error) {
	charset := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	if includeSpecial {
		charset += "!@#$%^&*()-_=+[]{}|;:,.<>?"
	}

	password := make([]byte, length)
	for i := 0; i < length; i++ {
		idx, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", fmt.Errorf("crypto/rand error: %w", err)
		}
		password[i] = charset[idx.Int64()]
	}

	return string(password), nil
}
