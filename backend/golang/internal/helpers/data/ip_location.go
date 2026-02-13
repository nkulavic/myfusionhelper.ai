package data

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/myfusionhelper/api/internal/helpers"
)

// NewIPLocation creates a new IPLocation helper instance
func NewIPLocation() helpers.Helper { return &IPLocation{} }

func init() {
	helpers.Register("ip_location", func() helpers.Helper { return &IPLocation{} })
}

// IPLocation looks up geolocation data for an IP address stored on a contact field.
// Uses the ip-api.com free API to resolve city, state, country, and zip code.
type IPLocation struct{}

func (h *IPLocation) GetName() string     { return "IP Location" }
func (h *IPLocation) GetType() string     { return "ip_location" }
func (h *IPLocation) GetCategory() string { return "data" }
func (h *IPLocation) GetDescription() string {
	return "Lookup IP address geolocation and store city, state, country, and zip on the contact"
}
func (h *IPLocation) RequiresCRM() bool       { return true }
func (h *IPLocation) SupportedCRMs() []string { return nil }

func (h *IPLocation) GetConfigSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"ip_field": map[string]interface{}{
				"type":        "string",
				"description": "The contact field containing the IP address to look up",
			},
			"city_field": map[string]interface{}{
				"type":        "string",
				"description": "The contact field to store the resolved city",
			},
			"state_field": map[string]interface{}{
				"type":        "string",
				"description": "The contact field to store the resolved state/region",
			},
			"country_field": map[string]interface{}{
				"type":        "string",
				"description": "The contact field to store the resolved country",
			},
			"zip_field": map[string]interface{}{
				"type":        "string",
				"description": "The contact field to store the resolved zip/postal code",
			},
		},
		"required": []string{"ip_field"},
	}
}

func (h *IPLocation) ValidateConfig(config map[string]interface{}) error {
	if _, ok := config["ip_field"].(string); !ok || config["ip_field"] == "" {
		return fmt.Errorf("ip_field is required")
	}
	return nil
}

// ipAPIResponse represents the JSON response from ip-api.com
type ipAPIResponse struct {
	Status      string `json:"status"`
	Message     string `json:"message,omitempty"`
	Country     string `json:"country"`
	CountryCode string `json:"countryCode"`
	Region      string `json:"region"`
	RegionName  string `json:"regionName"`
	City        string `json:"city"`
	Zip         string `json:"zip"`
	Lat         float64 `json:"lat"`
	Lon         float64 `json:"lon"`
	Timezone    string `json:"timezone"`
	ISP         string `json:"isp"`
	Query       string `json:"query"`
}

func (h *IPLocation) Execute(ctx context.Context, input helpers.HelperInput) (*helpers.HelperOutput, error) {
	ipField := input.Config["ip_field"].(string)

	cityField := ""
	if cf, ok := input.Config["city_field"].(string); ok {
		cityField = cf
	}

	stateField := ""
	if sf, ok := input.Config["state_field"].(string); ok {
		stateField = sf
	}

	countryField := ""
	if cf, ok := input.Config["country_field"].(string); ok {
		countryField = cf
	}

	zipField := ""
	if zf, ok := input.Config["zip_field"].(string); ok {
		zipField = zf
	}

	output := &helpers.HelperOutput{
		Actions: make([]helpers.HelperAction, 0),
		Logs:    make([]string, 0),
	}

	// Get the IP address from the contact field
	ipValue, err := input.Connector.GetContactFieldValue(ctx, input.ContactID, ipField)
	if err != nil || ipValue == nil || fmt.Sprintf("%v", ipValue) == "" {
		output.Success = true
		output.Message = fmt.Sprintf("IP field '%s' is empty, nothing to look up", ipField)
		output.Logs = append(output.Logs, output.Message)
		return output, nil
	}

	ipAddress := fmt.Sprintf("%v", ipValue)

	// Basic IP format validation
	if net.ParseIP(ipAddress) == nil {
		output.Message = fmt.Sprintf("Invalid IP address format: %s", ipAddress)
		output.Logs = append(output.Logs, output.Message)
		return output, fmt.Errorf("invalid IP address format: %s", ipAddress)
	}

	output.Logs = append(output.Logs, fmt.Sprintf("Looking up IP address: %s", ipAddress))

	// Call ip-api.com
	apiURL := fmt.Sprintf("http://ip-api.com/json/%s", ipAddress)
	client := &http.Client{Timeout: 10 * time.Second}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		output.Message = fmt.Sprintf("Failed to create IP lookup request: %v", err)
		return output, err
	}

	resp, err := client.Do(req)
	if err != nil {
		output.Message = fmt.Sprintf("IP lookup request failed: %v", err)
		return output, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		output.Message = fmt.Sprintf("Failed to read IP lookup response: %v", err)
		return output, err
	}

	var geoData ipAPIResponse
	if err := json.Unmarshal(body, &geoData); err != nil {
		output.Message = fmt.Sprintf("Failed to parse IP lookup response: %v", err)
		return output, err
	}

	if geoData.Status != "success" {
		errMsg := geoData.Message
		if errMsg == "" {
			errMsg = "unknown error"
		}
		output.Message = fmt.Sprintf("IP lookup failed for %s: %s", ipAddress, errMsg)
		output.Logs = append(output.Logs, output.Message)
		return output, fmt.Errorf("IP lookup failed: %s", errMsg)
	}

	output.Logs = append(output.Logs, fmt.Sprintf("IP lookup result: %s, %s, %s %s", geoData.City, geoData.RegionName, geoData.Country, geoData.Zip))

	// Store results in contact fields
	if cityField != "" && geoData.City != "" {
		setErr := input.Connector.SetContactFieldValue(ctx, input.ContactID, cityField, geoData.City)
		if setErr != nil {
			output.Logs = append(output.Logs, fmt.Sprintf("Failed to set city field: %v", setErr))
		} else {
			output.Actions = append(output.Actions, helpers.HelperAction{
				Type:   "field_updated",
				Target: cityField,
				Value:  geoData.City,
			})
		}
	}

	if stateField != "" && geoData.RegionName != "" {
		setErr := input.Connector.SetContactFieldValue(ctx, input.ContactID, stateField, geoData.RegionName)
		if setErr != nil {
			output.Logs = append(output.Logs, fmt.Sprintf("Failed to set state field: %v", setErr))
		} else {
			output.Actions = append(output.Actions, helpers.HelperAction{
				Type:   "field_updated",
				Target: stateField,
				Value:  geoData.RegionName,
			})
		}
	}

	if countryField != "" && geoData.Country != "" {
		setErr := input.Connector.SetContactFieldValue(ctx, input.ContactID, countryField, geoData.Country)
		if setErr != nil {
			output.Logs = append(output.Logs, fmt.Sprintf("Failed to set country field: %v", setErr))
		} else {
			output.Actions = append(output.Actions, helpers.HelperAction{
				Type:   "field_updated",
				Target: countryField,
				Value:  geoData.Country,
			})
		}
	}

	if zipField != "" && geoData.Zip != "" {
		setErr := input.Connector.SetContactFieldValue(ctx, input.ContactID, zipField, geoData.Zip)
		if setErr != nil {
			output.Logs = append(output.Logs, fmt.Sprintf("Failed to set zip field: %v", setErr))
		} else {
			output.Actions = append(output.Actions, helpers.HelperAction{
				Type:   "field_updated",
				Target: zipField,
				Value:  geoData.Zip,
			})
		}
	}

	output.Success = true
	output.Message = fmt.Sprintf("IP location resolved for %s: %s, %s, %s", ipAddress, geoData.City, geoData.RegionName, geoData.Country)
	output.ModifiedData = map[string]interface{}{
		"ip_address":   ipAddress,
		"city":         geoData.City,
		"state":        geoData.RegionName,
		"country":      geoData.Country,
		"country_code": geoData.CountryCode,
		"zip":          geoData.Zip,
		"latitude":     geoData.Lat,
		"longitude":    geoData.Lon,
		"timezone":     geoData.Timezone,
		"isp":          geoData.ISP,
	}
	output.Logs = append(output.Logs, fmt.Sprintf("IP location for contact %s: %s -> %s, %s, %s %s", input.ContactID, ipAddress, geoData.City, geoData.RegionName, geoData.Country, geoData.Zip))

	return output, nil
}
