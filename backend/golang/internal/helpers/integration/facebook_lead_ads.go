package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/myfusionhelper/api/internal/connectors"
	"github.com/myfusionhelper/api/internal/helpers"
)

// NewFacebookLeadAds creates a new FacebookLeadAds helper instance
func NewFacebookLeadAds() helpers.Helper { return &FacebookLeadAds{} }

func init() {
	helpers.Register("facebook_lead_ads", func() helpers.Helper { return &FacebookLeadAds{} })
}

// FacebookLeadAds syncs Facebook Lead Ads into CRM
type FacebookLeadAds struct{}

func (h *FacebookLeadAds) GetName() string     { return "Facebook Lead Ads" }
func (h *FacebookLeadAds) GetType() string     { return "facebook_lead_ads" }
func (h *FacebookLeadAds) GetCategory() string { return "integration" }
func (h *FacebookLeadAds) GetDescription() string {
	return "Sync Facebook Lead Ads form submissions into CRM as new contacts"
}
func (h *FacebookLeadAds) RequiresCRM() bool       { return true }
func (h *FacebookLeadAds) SupportedCRMs() []string { return nil }

func (h *FacebookLeadAds) GetConfigSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"lead_id": map[string]interface{}{
				"type":        "string",
				"description": "Facebook Lead ID from webhook",
			},
			"form_id": map[string]interface{}{
				"type":        "string",
				"description": "Facebook Form ID",
			},
			"apply_tag": map[string]interface{}{
				"type":        "string",
				"description": "Tag ID to apply to the new contact",
			},
			"field_mapping": map[string]interface{}{
				"type":        "object",
				"description": "Map Facebook lead fields to CRM fields (e.g., {\"full_name\": \"Name\", \"email\": \"Email\"})",
			},
			"service_connection_ids": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"facebook": map[string]interface{}{
						"type":        "string",
						"description": "Service connection ID for Facebook",
					},
				},
			},
		},
		"required": []string{"lead_id"},
	}
}

func (h *FacebookLeadAds) ValidateConfig(config map[string]interface{}) error {
	if _, ok := config["lead_id"].(string); !ok || config["lead_id"] == "" {
		return fmt.Errorf("lead_id is required")
	}
	return nil
}

func (h *FacebookLeadAds) Execute(ctx context.Context, input helpers.HelperInput) (*helpers.HelperOutput, error) {
	leadID := input.Config["lead_id"].(string)
	applyTag, _ := input.Config["apply_tag"].(string)
	fieldMapping, _ := input.Config["field_mapping"].(map[string]interface{})

	output := &helpers.HelperOutput{
		Actions: make([]helpers.HelperAction, 0),
		Logs:    make([]string, 0),
	}

	// Get Facebook service auth
	auth := input.ServiceAuths["facebook"]
	if auth == nil {
		output.Message = "Facebook connection required"
		return output, fmt.Errorf("Facebook connection required")
	}

	output.Logs = append(output.Logs, fmt.Sprintf("Fetching Facebook lead: %s", leadID))

	// Fetch lead data from Facebook Graph API
	apiURL := fmt.Sprintf("https://graph.facebook.com/v18.0/%s", leadID)
	params := url.Values{}
	params.Set("access_token", auth.AccessToken)
	params.Set("fields", "id,created_time,field_data,form_id")

	fullURL := fmt.Sprintf("%s?%s", apiURL, params.Encode())

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fullURL, nil)
	if err != nil {
		output.Message = fmt.Sprintf("Failed to create request: %v", err)
		return output, err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		output.Message = fmt.Sprintf("Facebook API request failed: %v", err)
		return output, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		output.Message = fmt.Sprintf("Failed to read Facebook response: %v", err)
		return output, err
	}

	if resp.StatusCode != http.StatusOK {
		output.Message = fmt.Sprintf("Facebook API returned status %d: %s", resp.StatusCode, string(body))
		return output, fmt.Errorf("Facebook API error: %d", resp.StatusCode)
	}

	var leadData map[string]interface{}
	if err := json.Unmarshal(body, &leadData); err != nil {
		output.Message = fmt.Sprintf("Failed to parse Facebook response: %v", err)
		return output, err
	}

	output.Logs = append(output.Logs, "Lead data fetched successfully")

	// Extract field_data array
	fieldDataArray, ok := leadData["field_data"].([]interface{})
	if !ok || len(fieldDataArray) == 0 {
		output.Message = "No field data in lead response"
		return output, fmt.Errorf("no field data")
	}

	// Convert field_data to map
	leadFields := make(map[string]string)
	for _, fieldInterface := range fieldDataArray {
		field, ok := fieldInterface.(map[string]interface{})
		if !ok {
			continue
		}
		name, _ := field["name"].(string)
		values, ok := field["values"].([]interface{})
		if ok && len(values) > 0 {
			value, _ := values[0].(string)
			leadFields[name] = value
		}
	}

	output.Logs = append(output.Logs, fmt.Sprintf("Extracted %d lead fields", len(leadFields)))

	// Build contact data for CRM
	contactInput := connectors.CreateContactInput{
		CustomFields: make(map[string]interface{}),
	}

	// Map standard fields
	if email, ok := leadFields["email"]; ok {
		contactInput.Email = email
	}
	if phone, ok := leadFields["phone_number"]; ok {
		contactInput.Phone = phone
	}
	if fullName, ok := leadFields["full_name"]; ok {
		// Split full name into first/last
		parts := strings.SplitN(fullName, " ", 2)
		contactInput.FirstName = parts[0]
		if len(parts) > 1 {
			contactInput.LastName = parts[1]
		}
	}
	if firstName, ok := leadFields["first_name"]; ok {
		contactInput.FirstName = firstName
	}
	if lastName, ok := leadFields["last_name"]; ok {
		contactInput.LastName = lastName
	}
	if company, ok := leadFields["company_name"]; ok {
		contactInput.Company = company
	}

	// Apply custom field mapping
	if fieldMapping != nil {
		for fbField, crmField := range fieldMapping {
			fbFieldStr, _ := fbField, true
			crmFieldStr, _ := crmField.(string)
			if value, ok := leadFields[fbFieldStr]; ok {
				contactInput.CustomFields[crmFieldStr] = value
			}
		}
	}

	// Create contact in CRM
	contact, err := input.Connector.CreateContact(ctx, contactInput)
	if err != nil {
		output.Message = fmt.Sprintf("Failed to create contact: %v", err)
		return output, err
	}

	output.Logs = append(output.Logs, fmt.Sprintf("Created contact: %s %s (%s)", contact.FirstName, contact.LastName, contact.Email))

	// Apply tag if configured
	if applyTag != "" {
		err := input.Connector.ApplyTag(ctx, contact.ID, applyTag)
		if err != nil {
			output.Logs = append(output.Logs, fmt.Sprintf("Failed to apply tag '%s': %v", applyTag, err))
		} else {
			output.Actions = append(output.Actions, helpers.HelperAction{
				Type:   "tag_applied",
				Target: contact.ID,
				Value:  applyTag,
			})
			output.Logs = append(output.Logs, fmt.Sprintf("Applied tag '%s'", applyTag))
		}
	}

	output.Success = true
	output.Message = fmt.Sprintf("Facebook lead synced to CRM: %s %s", contact.FirstName, contact.LastName)
	output.ModifiedData = map[string]interface{}{
		"contact_id":  contact.ID,
		"lead_id":     leadID,
		"email":       contact.Email,
		"first_name":  contact.FirstName,
		"last_name":   contact.LastName,
		"lead_fields": leadFields,
	}

	return output, nil
}
