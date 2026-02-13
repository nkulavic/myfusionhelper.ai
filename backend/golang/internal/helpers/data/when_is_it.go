package data

import (
	"context"
	"fmt"
	"time"

	"github.com/myfusionhelper/api/internal/helpers"
)

// NewWhenIsIt creates a new WhenIsIt helper instance
func NewWhenIsIt() helpers.Helper { return &WhenIsIt{} }

func init() {
	helpers.Register("when_is_it", func() helpers.Helper { return &WhenIsIt{} })
}

// WhenIsIt performs timezone-aware date formatting
type WhenIsIt struct{}

func (h *WhenIsIt) GetName() string        { return "When Is It" }
func (h *WhenIsIt) GetType() string        { return "when_is_it" }
func (h *WhenIsIt) GetCategory() string    { return "data" }
func (h *WhenIsIt) GetDescription() string { return "Convert dates between timezones and format output" }
func (h *WhenIsIt) RequiresCRM() bool      { return true }
func (h *WhenIsIt) SupportedCRMs() []string { return nil }

func (h *WhenIsIt) GetConfigSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"source_field": map[string]interface{}{
				"type":        "string",
				"description": "The field containing the date to convert",
			},
			"target_field": map[string]interface{}{
				"type":        "string",
				"description": "The field to store the converted date",
			},
			"from_timezone": map[string]interface{}{
				"type":        "string",
				"description": "Source timezone (IANA timezone name)",
				"default":     "UTC",
			},
			"to_timezone": map[string]interface{}{
				"type":        "string",
				"description": "Target timezone (IANA timezone name)",
			},
			"output_format": map[string]interface{}{
				"type":        "string",
				"description": "Output date format (Go time format string)",
			},
		},
		"required": []string{"source_field", "target_field", "to_timezone", "output_format"},
	}
}

func (h *WhenIsIt) ValidateConfig(config map[string]interface{}) error {
	if _, ok := config["source_field"].(string); !ok || config["source_field"] == "" {
		return fmt.Errorf("source_field is required")
	}
	if _, ok := config["target_field"].(string); !ok || config["target_field"] == "" {
		return fmt.Errorf("target_field is required")
	}
	if _, ok := config["to_timezone"].(string); !ok || config["to_timezone"] == "" {
		return fmt.Errorf("to_timezone is required")
	}
	if _, ok := config["output_format"].(string); !ok || config["output_format"] == "" {
		return fmt.Errorf("output_format is required")
	}
	return nil
}

func (h *WhenIsIt) Execute(ctx context.Context, input helpers.HelperInput) (*helpers.HelperOutput, error) {
	sourceField := input.Config["source_field"].(string)
	targetField := input.Config["target_field"].(string)
	fromTZ := "UTC"
	if tz, ok := input.Config["from_timezone"].(string); ok && tz != "" {
		fromTZ = tz
	}
	toTZ := input.Config["to_timezone"].(string)
	outputFormat := input.Config["output_format"].(string)

	output := &helpers.HelperOutput{
		Logs: make([]string, 0),
	}

	// Load timezones
	fromLoc, err := time.LoadLocation(fromTZ)
	if err != nil {
		output.Message = fmt.Sprintf("Invalid source timezone '%s': %v", fromTZ, err)
		return output, err
	}

	toLoc, err := time.LoadLocation(toTZ)
	if err != nil {
		output.Message = fmt.Sprintf("Invalid target timezone '%s': %v", toTZ, err)
		return output, err
	}

	// Get source date value
	rawValue, err := input.Connector.GetContactFieldValue(ctx, input.ContactID, sourceField)
	if err != nil {
		output.Message = fmt.Sprintf("Failed to read field '%s': %v", sourceField, err)
		return output, err
	}

	dateStr := fmt.Sprintf("%v", rawValue)
	if rawValue == nil || dateStr == "" || dateStr == "<nil>" {
		output.Success = true
		output.Message = fmt.Sprintf("Field '%s' is empty, nothing to convert", sourceField)
		return output, nil
	}

	// Parse the date
	parsedDate, err := parseWhenDate(dateStr)
	if err != nil {
		output.Message = fmt.Sprintf("Could not parse date '%s': %v", dateStr, err)
		return output, err
	}

	// Apply source timezone and convert to target
	dateInFromTZ := time.Date(
		parsedDate.Year(), parsedDate.Month(), parsedDate.Day(),
		parsedDate.Hour(), parsedDate.Minute(), parsedDate.Second(),
		parsedDate.Nanosecond(), fromLoc,
	)
	dateInToTZ := dateInFromTZ.In(toLoc)

	// Format output
	resultStr := dateInToTZ.Format(outputFormat)

	// Set target field
	err = input.Connector.SetContactFieldValue(ctx, input.ContactID, targetField, resultStr)
	if err != nil {
		output.Message = fmt.Sprintf("Failed to set result field '%s': %v", targetField, err)
		return output, err
	}

	output.Success = true
	output.Message = fmt.Sprintf("Converted date from %s to %s: %s", fromTZ, toTZ, resultStr)
	output.Actions = []helpers.HelperAction{
		{
			Type:   "field_updated",
			Target: targetField,
			Value:  resultStr,
		},
	}
	output.ModifiedData = map[string]interface{}{
		targetField: resultStr,
	}
	output.Logs = append(output.Logs, fmt.Sprintf("Converted '%s' from %s to %s -> '%s' on contact %s", dateStr, fromTZ, toTZ, resultStr, input.ContactID))

	return output, nil
}

func parseWhenDate(s string) (time.Time, error) {
	formats := []string{
		time.RFC3339,
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
		"2006-01-02",
		"01/02/2006",
		"01-02-2006",
		"01/02/2006 15:04:05",
		"01-02-2006 15:04:05",
		"Jan 2, 2006",
		"January 2, 2006",
		"02 Jan 2006",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, s); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("unrecognized date format: %s", s)
}
