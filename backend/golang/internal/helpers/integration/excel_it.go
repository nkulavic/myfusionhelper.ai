package integration

import (
	"context"
	"fmt"
	"strings"

	"github.com/myfusionhelper/api/internal/helpers"
)

func init() {
	helpers.Register("excel_it", func() helpers.Helper { return &ExcelIt{} })
}

// ExcelIt exports contact data to CSV or Excel format. It prepares the data
// and produces an export request for downstream processing by the execution layer
// which handles file generation and delivery.
// Supports configurable field selection, format, and delimiter.
type ExcelIt struct{}

func (h *ExcelIt) GetName() string     { return "Excel It" }
func (h *ExcelIt) GetType() string     { return "excel_it" }
func (h *ExcelIt) GetCategory() string { return "integration" }
func (h *ExcelIt) GetDescription() string {
	return "Export contact data to CSV or Excel format for download"
}
func (h *ExcelIt) RequiresCRM() bool       { return true }
func (h *ExcelIt) SupportedCRMs() []string { return nil }

func (h *ExcelIt) GetConfigSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"fields": map[string]interface{}{
				"type":        "array",
				"items":       map[string]interface{}{"type": "string"},
				"description": "List of contact field keys to export (e.g., FirstName, LastName, Email)",
			},
			"format": map[string]interface{}{
				"type":        "string",
				"enum":        []string{"csv", "xlsx"},
				"description": "Export file format",
				"default":     "csv",
			},
			"delimiter": map[string]interface{}{
				"type":        "string",
				"description": "Delimiter character for CSV format (default: comma)",
				"default":     ",",
			},
			"include_headers": map[string]interface{}{
				"type":        "boolean",
				"description": "Whether to include column headers in the export",
				"default":     true,
			},
		},
		"required": []string{"fields"},
	}
}

func (h *ExcelIt) ValidateConfig(config map[string]interface{}) error {
	fields, ok := config["fields"]
	if !ok {
		return fmt.Errorf("fields is required")
	}

	// Fields can be a []interface{} from JSON deserialization
	fieldSlice, ok := fields.([]interface{})
	if !ok || len(fieldSlice) == 0 {
		return fmt.Errorf("fields must be a non-empty array of field names")
	}

	if format, ok := config["format"].(string); ok && format != "" {
		if format != "csv" && format != "xlsx" {
			return fmt.Errorf("format must be 'csv' or 'xlsx', got '%s'", format)
		}
	}

	return nil
}

func (h *ExcelIt) Execute(ctx context.Context, input helpers.HelperInput) (*helpers.HelperOutput, error) {
	output := &helpers.HelperOutput{
		Actions: make([]helpers.HelperAction, 0),
		Logs:    make([]string, 0),
	}

	// Parse config
	format := "csv"
	if f, ok := input.Config["format"].(string); ok && f != "" {
		format = f
	}

	delimiter := ","
	if d, ok := input.Config["delimiter"].(string); ok && d != "" {
		delimiter = d
	}

	includeHeaders := true
	if ih, ok := input.Config["include_headers"].(bool); ok {
		includeHeaders = ih
	}

	// Parse fields list
	var fieldKeys []string
	if fields, ok := input.Config["fields"].([]interface{}); ok {
		for _, f := range fields {
			if s, ok := f.(string); ok && s != "" {
				fieldKeys = append(fieldKeys, s)
			}
		}
	}

	if len(fieldKeys) == 0 {
		output.Message = "No valid fields configured for export"
		return output, fmt.Errorf("no valid fields configured for export")
	}

	// Get contact data
	contact, err := input.Connector.GetContact(ctx, input.ContactID)
	if err != nil {
		output.Message = fmt.Sprintf("Failed to get contact data: %v", err)
		return output, err
	}

	// Build field data map
	fieldData := map[string]string{
		"Id":        contact.ID,
		"FirstName": contact.FirstName,
		"LastName":  contact.LastName,
		"Email":     contact.Email,
		"Phone":     contact.Phone,
		"Company":   contact.Company,
		"JobTitle":  contact.JobTitle,
	}

	// Add custom fields
	if contact.CustomFields != nil {
		for key, value := range contact.CustomFields {
			fieldData[key] = fmt.Sprintf("%v", value)
		}
	}

	// Build CSV row
	var values []string
	for _, key := range fieldKeys {
		val, exists := fieldData[key]
		if !exists {
			val = ""
		}
		// Escape values containing delimiter or quotes
		if strings.Contains(val, delimiter) || strings.Contains(val, "\"") || strings.Contains(val, "\n") {
			val = "\"" + strings.ReplaceAll(val, "\"", "\"\"") + "\""
		}
		values = append(values, val)
	}

	// Build export data
	exportData := map[string]interface{}{
		"format":          format,
		"delimiter":       delimiter,
		"include_headers": includeHeaders,
		"fields":          fieldKeys,
		"contact_id":      input.ContactID,
		"account_id":      input.AccountID,
		"user_id":         input.UserID,
		"helper_id":       input.HelperID,
	}

	if includeHeaders {
		exportData["header_row"] = strings.Join(fieldKeys, delimiter)
	}
	exportData["data_row"] = strings.Join(values, delimiter)

	output.Success = true
	output.Message = fmt.Sprintf("Export prepared for contact %s in %s format (%d fields)", input.ContactID, format, len(fieldKeys))
	output.Actions = []helpers.HelperAction{
		{
			Type:   "export_queued",
			Target: input.ContactID,
			Value:  exportData,
		},
	}
	output.ModifiedData = exportData
	output.Logs = append(output.Logs, fmt.Sprintf("Excel export for contact %s: %d fields in %s format", input.ContactID, len(fieldKeys), format))

	return output, nil
}
