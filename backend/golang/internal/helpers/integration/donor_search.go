package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/myfusionhelper/api/internal/helpers"
)

// NewDonorSearch creates a new DonorSearch helper instance
func NewDonorSearch() helpers.Helper { return &DonorSearch{} }

func init() {
	helpers.Register("donor_search", func() helpers.Helper { return &DonorSearch{} })
}

// DonorSearch queries the DonorLead.net API with contact information and saves
// the philanthropic rating and profile link back to the CRM.
// Requires a DonorLead service connection configured with an API key.
type DonorSearch struct{}

func (h *DonorSearch) GetName() string     { return "Donor Search" }
func (h *DonorSearch) GetType() string     { return "donor_search" }
func (h *DonorSearch) GetCategory() string { return "integration" }
func (h *DonorSearch) GetDescription() string {
	return "Queries DonorLead.net API with contact information and saves philanthropic rating and profile link to the CRM"
}
func (h *DonorSearch) RequiresCRM() bool       { return true }
func (h *DonorSearch) SupportedCRMs() []string { return nil }

func (h *DonorSearch) GetConfigSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"ds_rating_field": map[string]interface{}{
				"type":        "string",
				"description": "CRM custom field key to save the DS_Rating value",
			},
			"ds_profile_link_field": map[string]interface{}{
				"type":        "string",
				"description": "CRM custom field key to save the ProfileLink value",
			},
			"apply_tag": map[string]interface{}{
				"type":        "string",
				"description": "Tag ID to apply to the contact after a successful search",
			},
			"service_connection_ids": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"donorlead": map[string]interface{}{
						"type":        "string",
						"description": "Service connection ID for DonorLead",
					},
				},
				"description": "Service connection IDs for external integrations",
			},
		},
		"required": []string{},
	}
}

func (h *DonorSearch) ValidateConfig(config map[string]interface{}) error {
	// No strictly required config fields; service_connection_ids is validated at runtime
	return nil
}

func (h *DonorSearch) Execute(ctx context.Context, input helpers.HelperInput) (*helpers.HelperOutput, error) {
	output := &helpers.HelperOutput{
		Actions: make([]helpers.HelperAction, 0),
		Logs:    make([]string, 0),
	}

	// 1. Get the DonorLead service auth
	auth := input.ServiceAuths["donorlead"]
	if auth == nil {
		output.Message = "DonorLead connection required"
		return output, fmt.Errorf("DonorLead connection required")
	}

	// 2. Get contact data
	contact := input.ContactData
	if contact == nil {
		var err error
		contact, err = input.Connector.GetContact(ctx, input.ContactID)
		if err != nil {
			output.Message = fmt.Sprintf("Failed to get contact data: %v", err)
			return output, err
		}
	}

	firstName := contact.FirstName
	lastName := contact.LastName
	email := contact.Email
	phone := contact.Phone

	output.Logs = append(output.Logs, fmt.Sprintf("Searching DonorLead for: %s %s (%s)", firstName, lastName, email))

	// 3. Build the DonorLead API URL with proper encoding
	apiURL, err := url.Parse("https://data.donorlead.net/v2.1/")
	if err != nil {
		output.Message = fmt.Sprintf("Failed to build API URL: %v", err)
		return output, err
	}

	params := url.Values{}
	params.Set("api_key", auth.APIKey)
	params.Set("first", firstName)
	params.Set("last", lastName)
	if email != "" {
		params.Set("email", email)
	}
	if phone != "" {
		params.Set("phone", phone)
	}
	apiURL.RawQuery = params.Encode()

	// 4. Make the HTTP GET request
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL.String(), nil)
	if err != nil {
		output.Message = fmt.Sprintf("Failed to create HTTP request: %v", err)
		return output, err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		output.Message = fmt.Sprintf("DonorLead API request failed: %v", err)
		return output, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		output.Message = fmt.Sprintf("Failed to read DonorLead response: %v", err)
		return output, err
	}

	if resp.StatusCode != http.StatusOK {
		output.Message = fmt.Sprintf("DonorLead API returned status %d: %s", resp.StatusCode, string(body))
		return output, fmt.Errorf("DonorLead API returned status %d", resp.StatusCode)
	}

	// 5. Parse the JSON response
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		output.Message = fmt.Sprintf("Failed to parse DonorLead response: %v", err)
		return output, err
	}

	output.Logs = append(output.Logs, fmt.Sprintf("DonorLead API response received for contact %s", input.ContactID))

	// Extract DS_Rating and ProfileLink from the response
	dsRating := extractStringValue(result, "DS_Rating")
	profileLink := extractStringValue(result, "ProfileLink")

	output.Logs = append(output.Logs, fmt.Sprintf("DS_Rating: %s, ProfileLink: %s", dsRating, profileLink))

	// 6. Save DS_Rating to CRM if configured
	if dsRatingField, ok := input.Config["ds_rating_field"].(string); ok && dsRatingField != "" && dsRating != "" {
		err := input.Connector.SetContactFieldValue(ctx, input.ContactID, dsRatingField, dsRating)
		if err != nil {
			output.Logs = append(output.Logs, fmt.Sprintf("Failed to save DS_Rating to field '%s': %v", dsRatingField, err))
		} else {
			output.Actions = append(output.Actions, helpers.HelperAction{
				Type:   "field_updated",
				Target: dsRatingField,
				Value:  dsRating,
			})
			output.Logs = append(output.Logs, fmt.Sprintf("Saved DS_Rating '%s' to field '%s'", dsRating, dsRatingField))
		}
	}

	// 7. Save ProfileLink to CRM if configured
	if dsProfileLinkField, ok := input.Config["ds_profile_link_field"].(string); ok && dsProfileLinkField != "" && profileLink != "" {
		err := input.Connector.SetContactFieldValue(ctx, input.ContactID, dsProfileLinkField, profileLink)
		if err != nil {
			output.Logs = append(output.Logs, fmt.Sprintf("Failed to save ProfileLink to field '%s': %v", dsProfileLinkField, err))
		} else {
			output.Actions = append(output.Actions, helpers.HelperAction{
				Type:   "field_updated",
				Target: dsProfileLinkField,
				Value:  profileLink,
			})
			output.Logs = append(output.Logs, fmt.Sprintf("Saved ProfileLink '%s' to field '%s'", profileLink, dsProfileLinkField))
		}
	}

	// 8. Apply tag if configured
	if applyTag, ok := input.Config["apply_tag"].(string); ok && applyTag != "" {
		err := input.Connector.ApplyTag(ctx, input.ContactID, applyTag)
		if err != nil {
			output.Logs = append(output.Logs, fmt.Sprintf("Failed to apply tag '%s': %v", applyTag, err))
		} else {
			output.Actions = append(output.Actions, helpers.HelperAction{
				Type:   "tag_applied",
				Target: input.ContactID,
				Value:  applyTag,
			})
			output.Logs = append(output.Logs, fmt.Sprintf("Applied tag '%s' to contact %s", applyTag, input.ContactID))
		}
	}

	// 9. Return success with donor search results
	output.Success = true
	output.Message = fmt.Sprintf("Donor search completed for %s %s", firstName, lastName)
	output.ModifiedData = map[string]interface{}{
		"ds_rating":    dsRating,
		"profile_link": profileLink,
		"raw_response": result,
		"contact_id":   input.ContactID,
	}

	return output, nil
}

// extractStringValue extracts a string value from a nested map, checking both
// the top level and a "result" sub-object.
func extractStringValue(data map[string]interface{}, key string) string {
	// Check top level
	if val, ok := data[key]; ok {
		return fmt.Sprintf("%v", val)
	}

	// Check under common nested keys
	for _, nestedKey := range []string{"result", "Result", "data", "Data"} {
		if nested, ok := data[nestedKey].(map[string]interface{}); ok {
			if val, ok := nested[key]; ok {
				return fmt.Sprintf("%v", val)
			}
		}
	}

	return ""
}
