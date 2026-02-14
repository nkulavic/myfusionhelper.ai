package analytics

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/myfusionhelper/api/internal/helpers"
)

func init() {
	helpers.Register("last_click_it", func() helpers.Helper { return &LastClickIt{} })
}

// LastClickIt tracks last-click attribution for marketing touchpoints
// Stores UTM parameters and timestamp for conversion attribution
type LastClickIt struct{}

func (h *LastClickIt) GetName() string        { return "Last Click It" }
func (h *LastClickIt) GetType() string        { return "last_click_it" }
func (h *LastClickIt) GetCategory() string    { return "analytics" }
func (h *LastClickIt) GetDescription() string { return "Track last-click attribution for marketing touchpoints" }
func (h *LastClickIt) RequiresCRM() bool      { return true }
func (h *LastClickIt) SupportedCRMs() []string { return nil } // All CRMs

func (h *LastClickIt) GetConfigSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"utm_source": map[string]interface{}{
				"type":        "string",
				"description": "UTM source parameter (e.g., google, facebook, newsletter)",
			},
			"utm_medium": map[string]interface{}{
				"type":        "string",
				"description": "UTM medium parameter (e.g., cpc, social, email)",
			},
			"utm_campaign": map[string]interface{}{
				"type":        "string",
				"description": "UTM campaign parameter (e.g., spring_sale, product_launch)",
			},
			"utm_term": map[string]interface{}{
				"type":        "string",
				"description": "UTM term parameter for paid search keywords (optional)",
			},
			"utm_content": map[string]interface{}{
				"type":        "string",
				"description": "UTM content parameter for A/B testing (optional)",
			},
			"field_prefix": map[string]interface{}{
				"type":        "string",
				"description": "Custom field prefix for storing UTM data (default: 'last_click_')",
				"default":     "last_click_",
			},
			"overwrite": map[string]interface{}{
				"type":        "boolean",
				"description": "Overwrite existing attribution data (default: true for last-click)",
				"default":     true,
			},
		},
		"required": []string{"utm_source", "utm_medium"},
	}
}

func (h *LastClickIt) ValidateConfig(config map[string]interface{}) error {
	if source, ok := config["utm_source"].(string); !ok || source == "" {
		return fmt.Errorf("utm_source is required")
	}
	if medium, ok := config["utm_medium"].(string); !ok || medium == "" {
		return fmt.Errorf("utm_medium is required")
	}
	return nil
}

func (h *LastClickIt) Execute(ctx context.Context, input helpers.HelperInput) (*helpers.HelperOutput, error) {
	// Extract UTM parameters
	utmSource := input.Config["utm_source"].(string)
	utmMedium := input.Config["utm_medium"].(string)
	utmCampaign := ""
	utmTerm := ""
	utmContent := ""
	fieldPrefix := "last_click_"
	overwrite := true

	if val, ok := input.Config["utm_campaign"].(string); ok {
		utmCampaign = val
	}
	if val, ok := input.Config["utm_term"].(string); ok {
		utmTerm = val
	}
	if val, ok := input.Config["utm_content"].(string); ok {
		utmContent = val
	}
	if val, ok := input.Config["field_prefix"].(string); ok && val != "" {
		fieldPrefix = val
	}
	if val, ok := input.Config["overwrite"].(bool); ok {
		overwrite = val
	}

	output := &helpers.HelperOutput{
		Logs:    make([]string, 0),
		Actions: make([]helpers.HelperAction, 0),
	}

	// Build field names
	fields := map[string]string{
		fieldPrefix + "source":    utmSource,
		fieldPrefix + "medium":    utmMedium,
		fieldPrefix + "timestamp": time.Now().UTC().Format(time.RFC3339),
	}

	if utmCampaign != "" {
		fields[fieldPrefix+"campaign"] = utmCampaign
	}
	if utmTerm != "" {
		fields[fieldPrefix+"term"] = utmTerm
	}
	if utmContent != "" {
		fields[fieldPrefix+"content"] = utmContent
	}

	// Check if we should overwrite existing attribution
	if !overwrite {
		// Check if last_click_source already exists
		existingSource, err := input.Connector.GetContactFieldValue(ctx, input.ContactID, fieldPrefix+"source")
		if err == nil && existingSource != nil && existingSource != "" {
			output.Success = true
			output.Message = "Attribution already set (not overwriting)"
			output.Logs = append(output.Logs, fmt.Sprintf("Skipping attribution update for contact %s (overwrite=false and data exists)", input.ContactID))
			return output, nil
		}
	}

	// Set all UTM fields on contact
	errorCount := 0
	for fieldKey, fieldValue := range fields {
		err := input.Connector.SetContactFieldValue(ctx, input.ContactID, fieldKey, fieldValue)
		if err != nil {
			output.Logs = append(output.Logs, fmt.Sprintf("Failed to set %s: %v", fieldKey, err))
			errorCount++
		} else {
			output.Actions = append(output.Actions, helpers.HelperAction{
				Type:   "field_updated",
				Target: input.ContactID,
				Value:  fmt.Sprintf("%s=%s", fieldKey, fieldValue),
			})
			output.Logs = append(output.Logs, fmt.Sprintf("Set %s = %s", fieldKey, fieldValue))
		}
	}

	if errorCount > 0 {
		output.Message = fmt.Sprintf("Partially updated attribution (%d/%d fields failed)", errorCount, len(fields))
		return output, fmt.Errorf("failed to set %d attribution fields", errorCount)
	}

	output.Success = true
	output.Message = fmt.Sprintf("Last-click attribution tracked: %s/%s", utmSource, utmMedium)

	// Summary log
	attrSummary := fmt.Sprintf("source=%s, medium=%s", utmSource, utmMedium)
	if utmCampaign != "" {
		attrSummary += fmt.Sprintf(", campaign=%s", utmCampaign)
	}
	output.Logs = append(output.Logs, fmt.Sprintf("Updated last-click attribution for contact %s: %s", input.ContactID, attrSummary))

	return output, nil
}

// SanitizeFieldName converts a field name to CRM-safe format
func sanitizeFieldName(name string) string {
	// Replace spaces and special chars with underscores
	name = strings.ToLower(name)
	name = strings.ReplaceAll(name, " ", "_")
	name = strings.ReplaceAll(name, "-", "_")
	return name
}
