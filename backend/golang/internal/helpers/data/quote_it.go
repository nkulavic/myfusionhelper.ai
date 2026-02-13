package data

import (
	"context"
	"fmt"
	"strings"

	"github.com/myfusionhelper/api/internal/helpers"
)

// NewQuoteIt creates a new QuoteIt helper instance
func NewQuoteIt() helpers.Helper { return &QuoteIt{} }

func init() {
	helpers.Register("quote_it", func() helpers.Helper { return &QuoteIt{} })
}

// QuoteIt wraps text in quotes or other delimiters, useful for formatting data
type QuoteIt struct{}

func (h *QuoteIt) GetName() string     { return "Quote It" }
func (h *QuoteIt) GetType() string     { return "quote_it" }
func (h *QuoteIt) GetCategory() string { return "data" }
func (h *QuoteIt) GetDescription() string {
	return "Wrap text fields in quotes or custom delimiters for formatting and data export"
}
func (h *QuoteIt) RequiresCRM() bool       { return true }
func (h *QuoteIt) SupportedCRMs() []string { return nil }

func (h *QuoteIt) GetConfigSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"source_field": map[string]interface{}{
				"type":        "string",
				"description": "Source field containing the text to quote",
			},
			"target_field": map[string]interface{}{
				"type":        "string",
				"description": "Target field to store the quoted result",
			},
			"quote_style": map[string]interface{}{
				"type":        "string",
				"enum":        []string{"double", "single", "backtick", "parentheses", "brackets", "braces", "custom"},
				"description": "Style of quotes to use",
				"default":     "double",
			},
			"custom_left": map[string]interface{}{
				"type":        "string",
				"description": "Custom left delimiter (required if quote_style is 'custom')",
			},
			"custom_right": map[string]interface{}{
				"type":        "string",
				"description": "Custom right delimiter (required if quote_style is 'custom')",
			},
			"escape_quotes": map[string]interface{}{
				"type":        "boolean",
				"description": "Whether to escape existing quotes in the text",
				"default":     false,
			},
		},
		"required": []string{"source_field", "target_field"},
	}
}

func (h *QuoteIt) ValidateConfig(config map[string]interface{}) error {
	if _, ok := config["source_field"].(string); !ok || config["source_field"] == "" {
		return fmt.Errorf("source_field is required")
	}
	if _, ok := config["target_field"].(string); !ok || config["target_field"] == "" {
		return fmt.Errorf("target_field is required")
	}

	quoteStyle, _ := config["quote_style"].(string)
	if quoteStyle == "custom" {
		if _, ok := config["custom_left"].(string); !ok {
			return fmt.Errorf("custom_left is required when quote_style is 'custom'")
		}
		if _, ok := config["custom_right"].(string); !ok {
			return fmt.Errorf("custom_right is required when quote_style is 'custom'")
		}
	}

	return nil
}

func (h *QuoteIt) Execute(ctx context.Context, input helpers.HelperInput) (*helpers.HelperOutput, error) {
	sourceField := input.Config["source_field"].(string)
	targetField := input.Config["target_field"].(string)
	quoteStyle, _ := input.Config["quote_style"].(string)
	if quoteStyle == "" {
		quoteStyle = "double"
	}
	escapeQuotes, _ := input.Config["escape_quotes"].(bool)

	output := &helpers.HelperOutput{
		Logs: make([]string, 0),
	}

	// Get source field value
	sourceValue, err := input.Connector.GetContactFieldValue(ctx, input.ContactID, sourceField)
	if err != nil {
		output.Message = fmt.Sprintf("Failed to get source field '%s': %v", sourceField, err)
		return output, err
	}

	// Convert to string
	sourceText := fmt.Sprintf("%v", sourceValue)
	output.Logs = append(output.Logs, fmt.Sprintf("Source text: %s = '%s'", sourceField, sourceText))

	// Determine delimiters
	var leftQuote, rightQuote string
	switch quoteStyle {
	case "double":
		leftQuote = "\""
		rightQuote = "\""
	case "single":
		leftQuote = "'"
		rightQuote = "'"
	case "backtick":
		leftQuote = "`"
		rightQuote = "`"
	case "parentheses":
		leftQuote = "("
		rightQuote = ")"
	case "brackets":
		leftQuote = "["
		rightQuote = "]"
	case "braces":
		leftQuote = "{"
		rightQuote = "}"
	case "custom":
		leftQuote = input.Config["custom_left"].(string)
		rightQuote = input.Config["custom_right"].(string)
	default:
		leftQuote = "\""
		rightQuote = "\""
	}

	// Escape existing quotes if requested
	processedText := sourceText
	if escapeQuotes {
		if quoteStyle == "double" || quoteStyle == "custom" && leftQuote == "\"" {
			processedText = strings.ReplaceAll(processedText, "\"", "\\\"")
		}
		if quoteStyle == "single" || quoteStyle == "custom" && leftQuote == "'" {
			processedText = strings.ReplaceAll(processedText, "'", "\\'")
		}
		output.Logs = append(output.Logs, "Escaped existing quotes")
	}

	// Apply quotes
	quotedText := leftQuote + processedText + rightQuote

	// Save to target field
	err = input.Connector.SetContactFieldValue(ctx, input.ContactID, targetField, quotedText)
	if err != nil {
		output.Message = fmt.Sprintf("Failed to set target field '%s': %v", targetField, err)
		return output, err
	}

	output.Success = true
	output.Message = fmt.Sprintf("Quoted text with %s style, saved to '%s'", quoteStyle, targetField)
	output.Actions = []helpers.HelperAction{
		{
			Type:   "field_updated",
			Target: targetField,
			Value:  quotedText,
		},
	}
	output.ModifiedData = map[string]interface{}{
		targetField:    quotedText,
		"quote_style":  quoteStyle,
		"source_text":  sourceText,
		"quoted_text":  quotedText,
	}
	output.Logs = append(output.Logs, fmt.Sprintf("Result: %s = %s", targetField, quotedText))

	return output, nil
}
