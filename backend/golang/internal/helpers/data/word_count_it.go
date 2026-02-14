package data

import (
	"context"
	"fmt"
	"strings"

	"github.com/myfusionhelper/api/internal/helpers"
)

// NewWordCountIt creates a new WordCountIt helper instance
func NewWordCountIt() helpers.Helper { return &WordCountIt{} }

func init() {
	helpers.Register("word_count_it", func() helpers.Helper { return &WordCountIt{} })
}

// WordCountIt counts words or characters in a field and stores the count
type WordCountIt struct{}

func (h *WordCountIt) GetName() string        { return "Word Count It" }
func (h *WordCountIt) GetType() string        { return "word_count_it" }
func (h *WordCountIt) GetCategory() string    { return "data" }
func (h *WordCountIt) GetDescription() string { return "Count words or characters in a field and store the count" }
func (h *WordCountIt) RequiresCRM() bool      { return true }
func (h *WordCountIt) SupportedCRMs() []string { return nil }

func (h *WordCountIt) GetConfigSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"source_field": map[string]interface{}{
				"type":        "string",
				"description": "The field to count words/characters in",
			},
			"target_field": map[string]interface{}{
				"type":        "string",
				"description": "The field to store the count",
			},
			"count_type": map[string]interface{}{
				"type":        "string",
				"enum":        []string{"words", "characters"},
				"description": "Whether to count words or characters",
			},
		},
		"required": []string{"source_field", "target_field", "count_type"},
	}
}

func (h *WordCountIt) ValidateConfig(config map[string]interface{}) error {
	if _, ok := config["source_field"].(string); !ok || config["source_field"] == "" {
		return fmt.Errorf("source_field is required")
	}
	if _, ok := config["target_field"].(string); !ok || config["target_field"] == "" {
		return fmt.Errorf("target_field is required")
	}

	countType, ok := config["count_type"].(string)
	if !ok || countType == "" {
		return fmt.Errorf("count_type is required")
	}
	if countType != "words" && countType != "characters" {
		return fmt.Errorf("count_type must be 'words' or 'characters'")
	}

	return nil
}

func (h *WordCountIt) Execute(ctx context.Context, input helpers.HelperInput) (*helpers.HelperOutput, error) {
	sourceField := input.Config["source_field"].(string)
	targetField := input.Config["target_field"].(string)
	countType := input.Config["count_type"].(string)

	output := &helpers.HelperOutput{
		Logs: make([]string, 0),
	}

	// Get source field value
	rawValue, err := input.Connector.GetContactFieldValue(ctx, input.ContactID, sourceField)
	if err != nil {
		output.Message = fmt.Sprintf("Failed to read field '%s': %v", sourceField, err)
		return output, err
	}

	strValue := fmt.Sprintf("%v", rawValue)
	if rawValue == nil || strValue == "<nil>" {
		strValue = ""
	}

	// Count
	var count int
	switch countType {
	case "words":
		if strings.TrimSpace(strValue) == "" {
			count = 0
		} else {
			count = len(strings.Fields(strValue))
		}
	case "characters":
		count = len([]rune(strValue))
	}

	countStr := fmt.Sprintf("%d", count)

	// Set target field
	err = input.Connector.SetContactFieldValue(ctx, input.ContactID, targetField, countStr)
	if err != nil {
		output.Message = fmt.Sprintf("Failed to set count field '%s': %v", targetField, err)
		return output, err
	}

	output.Success = true
	output.Message = fmt.Sprintf("Counted %d %s in '%s'", count, countType, sourceField)
	output.Actions = []helpers.HelperAction{
		{
			Type:   "field_updated",
			Target: targetField,
			Value:  countStr,
		},
	}
	output.ModifiedData = map[string]interface{}{
		targetField: countStr,
	}
	output.Logs = append(output.Logs, fmt.Sprintf("Counted %d %s in field '%s', stored in '%s' on contact %s", count, countType, sourceField, targetField, input.ContactID))

	return output, nil
}
