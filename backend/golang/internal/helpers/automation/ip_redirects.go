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
	helpers.Register("ip_redirects", func() helpers.Helper { return &IPRedirects{} })
}

// IPRedirects generates redirect URLs based on IP geolocation
type IPRedirects struct{}

func (h *IPRedirects) GetName() string     { return "IP Redirects" }
func (h *IPRedirects) GetType() string     { return "ip_redirects" }
func (h *IPRedirects) GetCategory() string { return "automation" }
func (h *IPRedirects) GetDescription() string {
	return "Generate dynamic redirect URLs based on visitor IP geolocation (country-specific landing pages)"
}
func (h *IPRedirects) RequiresCRM() bool       { return false }
func (h *IPRedirects) SupportedCRMs() []string { return nil }

func (h *IPRedirects) GetConfigSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"ip_address": map[string]interface{}{
				"type":        "string",
				"description": "IP address to look up",
			},
			"country_urls": map[string]interface{}{
				"type":        "object",
				"description": "Map of country codes to redirect URLs (e.g., {\"US\": \"https://example.com/us\", \"CA\": \"https://example.com/ca\"})",
			},
			"default_url": map[string]interface{}{
				"type":        "string",
				"description": "Default URL if country doesn't match any rules",
			},
			"save_redirect_to": map[string]interface{}{
				"type":        "string",
				"description": "Optional CRM field to save the redirect URL (requires CRM connection)",
			},
		},
		"required": []string{"ip_address"},
	}
}

func (h *IPRedirects) ValidateConfig(config map[string]interface{}) error {
	if _, ok := config["ip_address"].(string); !ok || config["ip_address"] == "" {
		return fmt.Errorf("ip_address is required")
	}
	return nil
}

func (h *IPRedirects) Execute(ctx context.Context, input helpers.HelperInput) (*helpers.HelperOutput, error) {
	ipAddress := input.Config["ip_address"].(string)
	defaultURL, _ := input.Config["default_url"].(string)
	saveRedirectTo, _ := input.Config["save_redirect_to"].(string)

	// Parse country URLs
	countryURLs := make(map[string]string)
	if urls, ok := input.Config["country_urls"].(map[string]interface{}); ok {
		for country, url := range urls {
			if urlStr, ok := url.(string); ok {
				countryURLs[country] = urlStr
			}
		}
	}

	output := &helpers.HelperOutput{
		Actions: make([]helpers.HelperAction, 0),
		Logs:    make([]string, 0),
	}

	output.Logs = append(output.Logs, fmt.Sprintf("Looking up IP geolocation for redirect: %s", ipAddress))

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
		// On geolocation failure, use default URL
		if defaultURL != "" {
			output.Success = true
			output.Message = fmt.Sprintf("Geolocation failed, using default URL: %s", defaultURL)
			output.ModifiedData = map[string]interface{}{
				"redirect_url": defaultURL,
				"ip_address":   ipAddress,
				"reason":       "geolocation_failed",
			}
			return output, nil
		}
		message, _ := geoData["message"].(string)
		output.Message = fmt.Sprintf("Geolocation lookup failed and no default URL: %s", message)
		return output, fmt.Errorf("geolocation failed: %s", message)
	}

	countryCode, _ := geoData["countryCode"].(string)
	country, _ := geoData["country"].(string)

	output.Logs = append(output.Logs, fmt.Sprintf("IP country: %s (%s)", country, countryCode))

	// Find matching redirect URL
	var redirectURL string
	var matchReason string

	if url, ok := countryURLs[countryCode]; ok {
		redirectURL = url
		matchReason = fmt.Sprintf("country_match:%s", countryCode)
		output.Logs = append(output.Logs, fmt.Sprintf("Country matched: %s -> %s", countryCode, redirectURL))
	} else if defaultURL != "" {
		redirectURL = defaultURL
		matchReason = "default"
		output.Logs = append(output.Logs, fmt.Sprintf("No country match, using default: %s", redirectURL))
	} else {
		output.Message = fmt.Sprintf("No redirect URL found for country %s and no default URL configured", countryCode)
		return output, fmt.Errorf("no redirect URL")
	}

	// Save redirect URL to CRM field if configured
	if saveRedirectTo != "" && input.Connector != nil && input.ContactID != "" {
		err := input.Connector.SetContactFieldValue(ctx, input.ContactID, saveRedirectTo, redirectURL)
		if err != nil {
			output.Logs = append(output.Logs, fmt.Sprintf("Warning: Failed to save redirect URL to field '%s': %v", saveRedirectTo, err))
		} else {
			output.Actions = append(output.Actions, helpers.HelperAction{
				Type:   "field_updated",
				Target: saveRedirectTo,
				Value:  redirectURL,
			})
			output.Logs = append(output.Logs, fmt.Sprintf("Saved redirect URL to field '%s'", saveRedirectTo))
		}
	}

	output.Success = true
	output.Message = fmt.Sprintf("Redirect URL determined for %s: %s", country, redirectURL)
	output.ModifiedData = map[string]interface{}{
		"redirect_url": redirectURL,
		"ip_address":   ipAddress,
		"country":      country,
		"country_code": countryCode,
		"match_reason": matchReason,
	}

	output.Actions = append(output.Actions, helpers.HelperAction{
		Type:   "redirect_determined",
		Target: countryCode,
		Value:  redirectURL,
	})

	return output, nil
}
