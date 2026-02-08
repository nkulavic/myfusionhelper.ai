package automation

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/myfusionhelper/api/internal/helpers"
)

func init() {
	helpers.Register("route_it_geo", func() helpers.Helper { return &RouteItGeo{} })
}

// RouteItGeo routes contacts based on geographic location (IP geolocation)
type RouteItGeo struct{}

func (h *RouteItGeo) GetName() string     { return "Route It - Geographic" }
func (h *RouteItGeo) GetType() string     { return "route_it_geo" }
func (h *RouteItGeo) GetCategory() string { return "automation" }
func (h *RouteItGeo) GetDescription() string {
	return "Route contacts to different URLs based on geographic location (country, region, city)"
}
func (h *RouteItGeo) RequiresCRM() bool       { return false }
func (h *RouteItGeo) SupportedCRMs() []string { return nil }

func (h *RouteItGeo) GetConfigSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"ip_address": map[string]interface{}{
				"type":        "string",
				"description": "IP address to geolocate (required)",
			},
			"country_routes": map[string]interface{}{
				"type":        "object",
				"description": "Map of country codes to redirect URLs (e.g., {\"US\": \"https://example.com/us\", \"CA\": \"https://example.com/ca\"})",
			},
			"region_routes": map[string]interface{}{
				"type":        "object",
				"description": "Map of region names to redirect URLs (e.g., {\"California\": \"https://example.com/ca-promo\"})",
			},
			"fallback_url": map[string]interface{}{
				"type":        "string",
				"description": "Default URL if no geographic match",
			},
			"save_to_field": map[string]interface{}{
				"type":        "string",
				"description": "Optional: CRM field to save the selected URL to",
			},
			"save_location_to": map[string]interface{}{
				"type":        "string",
				"description": "Optional: CRM field to save the formatted location string to",
			},
			"apply_tag": map[string]interface{}{
				"type":        "string",
				"description": "Optional: Tag ID to apply when routing occurs",
			},
		},
		"required": []string{"ip_address"},
	}
}

func (h *RouteItGeo) ValidateConfig(config map[string]interface{}) error {
	if _, ok := config["ip_address"].(string); !ok || config["ip_address"] == "" {
		return fmt.Errorf("ip_address is required")
	}
	return nil
}

func (h *RouteItGeo) Execute(ctx context.Context, input helpers.HelperInput) (*helpers.HelperOutput, error) {
	ipAddress := input.Config["ip_address"].(string)
	fallbackURL, _ := input.Config["fallback_url"].(string)
	saveToField, _ := input.Config["save_to_field"].(string)
	saveLocationTo, _ := input.Config["save_location_to"].(string)
	applyTag, _ := input.Config["apply_tag"].(string)

	// Parse country routes
	countryRoutes := make(map[string]string)
	if cr, ok := input.Config["country_routes"].(map[string]interface{}); ok {
		for country, url := range cr {
			if urlStr, ok := url.(string); ok {
				countryRoutes[country] = urlStr
			}
		}
	}

	// Parse region routes
	regionRoutes := make(map[string]string)
	if rr, ok := input.Config["region_routes"].(map[string]interface{}); ok {
		for region, url := range rr {
			if urlStr, ok := url.(string); ok {
				regionRoutes[region] = urlStr
			}
		}
	}

	output := &helpers.HelperOutput{
		Actions: make([]helpers.HelperAction, 0),
		Logs:    make([]string, 0),
	}

	output.Logs = append(output.Logs, fmt.Sprintf("Looking up IP geolocation for routing: %s", ipAddress))

	// Use ip-api.com for geolocation
	apiURL := fmt.Sprintf("http://ip-api.com/json/%s?fields=status,message,country,countryCode,region,regionName,city", ipAddress)

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
		// On geolocation failure, use fallback URL
		if fallbackURL != "" {
			output.Success = true
			output.Message = fmt.Sprintf("Geolocation failed, using fallback URL: %s", fallbackURL)
			output.ModifiedData = map[string]interface{}{
				"redirect_url":   fallbackURL,
				"routing_reason": "geolocation_failed",
			}
			return output, nil
		}
		message, _ := geoData["message"].(string)
		output.Message = fmt.Sprintf("Geolocation lookup failed and no fallback URL: %s", message)
		return output, fmt.Errorf("geolocation failed: %s", message)
	}

	countryCode, _ := geoData["countryCode"].(string)
	country, _ := geoData["country"].(string)
	regionName, _ := geoData["regionName"].(string)
	city, _ := geoData["city"].(string)

	locationString := fmt.Sprintf("%s, %s, %s", city, regionName, country)
	output.Logs = append(output.Logs, fmt.Sprintf("IP geolocation: %s (country: %s, region: %s)", locationString, countryCode, regionName))

	// Find matching redirect URL
	var selectedURL string
	var routingReason string

	// Check region routes first (more specific)
	if url, ok := regionRoutes[regionName]; ok {
		selectedURL = url
		routingReason = fmt.Sprintf("region_match:%s", regionName)
		output.Logs = append(output.Logs, fmt.Sprintf("Region matched: %s -> %s", regionName, selectedURL))
	} else if url, ok := countryRoutes[countryCode]; ok {
		selectedURL = url
		routingReason = fmt.Sprintf("country_match:%s", countryCode)
		output.Logs = append(output.Logs, fmt.Sprintf("Country matched: %s -> %s", countryCode, selectedURL))
	} else if fallbackURL != "" {
		selectedURL = fallbackURL
		routingReason = "fallback"
		output.Logs = append(output.Logs, fmt.Sprintf("No geographic match, using fallback: %s", selectedURL))
	} else {
		output.Message = fmt.Sprintf("No redirect URL found for %s (%s) and no fallback URL configured", country, countryCode)
		return output, fmt.Errorf("no redirect URL")
	}

	// Optional: Save location to CRM field
	if saveLocationTo != "" && input.Connector != nil && input.ContactID != "" {
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

	// Optional: Save URL to CRM field
	if saveToField != "" && input.Connector != nil && input.ContactID != "" {
		err := input.Connector.SetContactFieldValue(ctx, input.ContactID, saveToField, selectedURL)
		if err != nil {
			output.Logs = append(output.Logs, fmt.Sprintf("Warning: Failed to save URL to field '%s': %v", saveToField, err))
		} else {
			output.Actions = append(output.Actions, helpers.HelperAction{
				Type:   "field_updated",
				Target: saveToField,
				Value:  selectedURL,
			})
			output.Logs = append(output.Logs, fmt.Sprintf("Saved URL to field '%s'", saveToField))
		}
	}

	// Optional: Apply tag
	if applyTag != "" && input.Connector != nil && input.ContactID != "" {
		err := input.Connector.ApplyTag(ctx, input.ContactID, applyTag)
		if err != nil {
			output.Logs = append(output.Logs, fmt.Sprintf("Warning: Failed to apply tag '%s': %v", applyTag, err))
		} else {
			output.Actions = append(output.Actions, helpers.HelperAction{
				Type:   "tag_applied",
				Target: input.ContactID,
				Value:  applyTag,
			})
			output.Logs = append(output.Logs, fmt.Sprintf("Applied tag '%s'", applyTag))
		}
	}

	output.Success = true
	output.Message = fmt.Sprintf("Routed to %s based on geographic location: %s", selectedURL, locationString)
	output.ModifiedData = map[string]interface{}{
		"redirect_url":   selectedURL,
		"routing_reason": routingReason,
		"country":        country,
		"country_code":   countryCode,
		"region":         regionName,
		"city":           city,
		"location":       locationString,
	}
	output.Actions = append(output.Actions, helpers.HelperAction{
		Type:   "contact_routed",
		Target: strings.ToLower(countryCode),
		Value:  selectedURL,
	})

	return output, nil
}
