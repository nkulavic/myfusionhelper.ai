package automation

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/myfusionhelper/api/internal/helpers"
)

func init() {
	helpers.Register("ip_notifications", func() helpers.Helper { return &IPNotifications{} })
}

// IPNotifications sends notifications based on IP geolocation
type IPNotifications struct{}

func (h *IPNotifications) GetName() string     { return "IP Notifications" }
func (h *IPNotifications) GetType() string     { return "ip_notifications" }
func (h *IPNotifications) GetCategory() string { return "automation" }
func (h *IPNotifications) GetDescription() string {
	return "Send notifications based on visitor IP geolocation (country, region, city)"
}
func (h *IPNotifications) RequiresCRM() bool       { return true }
func (h *IPNotifications) SupportedCRMs() []string { return nil }

func (h *IPNotifications) GetConfigSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"ip_address": map[string]interface{}{
				"type":        "string",
				"description": "IP address to look up",
			},
			"match_countries": map[string]interface{}{
				"type":        "array",
				"items":       map[string]interface{}{"type": "string"},
				"description": "Country codes to trigger notification (e.g., [\"US\", \"CA\", \"GB\"])",
			},
			"match_regions": map[string]interface{}{
				"type":        "array",
				"items":       map[string]interface{}{"type": "string"},
				"description": "Region/state names to trigger notification",
			},
			"apply_tag": map[string]interface{}{
				"type":        "string",
				"description": "Tag ID to apply when location matches",
			},
			"save_location_to": map[string]interface{}{
				"type":        "string",
				"description": "Field to save formatted location string",
			},
		},
		"required": []string{"ip_address"},
	}
}

func (h *IPNotifications) ValidateConfig(config map[string]interface{}) error {
	if _, ok := config["ip_address"].(string); !ok || config["ip_address"] == "" {
		return fmt.Errorf("ip_address is required")
	}
	return nil
}

func (h *IPNotifications) Execute(ctx context.Context, input helpers.HelperInput) (*helpers.HelperOutput, error) {
	ipAddress := input.Config["ip_address"].(string)
	applyTag, _ := input.Config["apply_tag"].(string)
	saveLocationTo, _ := input.Config["save_location_to"].(string)

	// Parse match criteria
	var matchCountries []string
	if countries, ok := input.Config["match_countries"].([]interface{}); ok {
		for _, c := range countries {
			if country, ok := c.(string); ok {
				matchCountries = append(matchCountries, country)
			}
		}
	}

	var matchRegions []string
	if regions, ok := input.Config["match_regions"].([]interface{}); ok {
		for _, r := range regions {
			if region, ok := r.(string); ok {
				matchRegions = append(matchRegions, region)
			}
		}
	}

	output := &helpers.HelperOutput{
		Actions: make([]helpers.HelperAction, 0),
		Logs:    make([]string, 0),
	}

	output.Logs = append(output.Logs, fmt.Sprintf("Looking up IP geolocation: %s", ipAddress))

	// Use ip-api.com for geolocation (free tier, no API key required)
	apiURL := fmt.Sprintf("http://ip-api.com/json/%s?fields=status,message,country,countryCode,region,regionName,city,lat,lon", ipAddress)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		output.Message = fmt.Sprintf("Failed to create geolocation request: %v", err)
		return output, err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		output.Message = fmt.Sprintf("Geolocation lookup failed: %v", err)
		return output, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		output.Message = fmt.Sprintf("Failed to read geolocation response: %v", err)
		return output, err
	}

	var geoData map[string]interface{}
	if err := json.Unmarshal(body, &geoData); err != nil {
		output.Message = fmt.Sprintf("Failed to parse geolocation response: %v", err)
		return output, err
	}

	// Check if lookup succeeded
	if status, ok := geoData["status"].(string); !ok || status != "success" {
		message, _ := geoData["message"].(string)
		output.Message = fmt.Sprintf("Geolocation lookup failed: %s", message)
		return output, fmt.Errorf("geolocation failed: %s", message)
	}

	countryCode, _ := geoData["countryCode"].(string)
	country, _ := geoData["country"].(string)
	regionName, _ := geoData["regionName"].(string)
	city, _ := geoData["city"].(string)

	locationString := fmt.Sprintf("%s, %s, %s", city, regionName, country)

	output.Logs = append(output.Logs, fmt.Sprintf("IP geolocation: %s (country: %s, region: %s)", locationString, countryCode, regionName))

	// Check if location matches criteria
	locationMatches := false

	if len(matchCountries) > 0 {
		for _, mc := range matchCountries {
			if mc == countryCode {
				locationMatches = true
				output.Logs = append(output.Logs, fmt.Sprintf("Country code matched: %s", countryCode))
				break
			}
		}
	}

	if len(matchRegions) > 0 && !locationMatches {
		for _, mr := range matchRegions {
			if mr == regionName {
				locationMatches = true
				output.Logs = append(output.Logs, fmt.Sprintf("Region matched: %s", regionName))
				break
			}
		}
	}

	// If no criteria specified, consider it a match
	if len(matchCountries) == 0 && len(matchRegions) == 0 {
		locationMatches = true
	}

	// Save location if configured
	if saveLocationTo != "" {
		err := input.Connector.SetContactFieldValue(ctx, input.ContactID, saveLocationTo, locationString)
		if err != nil {
			output.Logs = append(output.Logs, fmt.Sprintf("Warning: Failed to save location to field '%s': %v", saveLocationTo, err))
		} else {
			output.Actions = append(output.Actions, helpers.HelperAction{
				Type:   "field_updated",
				Target: saveLocationTo,
				Value:  locationString,
			})
			output.Logs = append(output.Logs, fmt.Sprintf("Saved location to field '%s'", saveLocationTo))
		}
	}

	// Apply tag if location matches
	if locationMatches && applyTag != "" {
		err := input.Connector.ApplyTag(ctx, input.ContactID, applyTag)
		if err != nil {
			output.Logs = append(output.Logs, fmt.Sprintf("Failed to apply tag '%s': %v", applyTag, err))
		} else {
			output.Actions = append(output.Actions, helpers.HelperAction{
				Type:   "tag_applied",
				Target: input.ContactID,
				Value:  applyTag,
			})
			output.Logs = append(output.Logs, fmt.Sprintf("Applied tag '%s' (location matched)", applyTag))
		}
	}

	output.Success = true
	output.Message = fmt.Sprintf("IP geolocation: %s", locationString)
	output.ModifiedData = map[string]interface{}{
		"ip_address":      ipAddress,
		"country":         country,
		"country_code":    countryCode,
		"region":          regionName,
		"city":            city,
		"location_string": locationString,
		"location_matches": locationMatches,
	}

	return output, nil
}
