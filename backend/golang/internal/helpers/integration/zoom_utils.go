package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/myfusionhelper/api/internal/connectors"
)

// makeZoomAPIRequest makes an authenticated API request to Zoom.
// zoomAuth should be a *connectors.ConnectorConfig with AccessToken.
func makeZoomAPIRequest(ctx context.Context, method, url string, payload interface{}, zoomAuth interface{}) (map[string]interface{}, error) {
	// Extract access token from zoomAuth
	var accessToken string
	if authConfig, ok := zoomAuth.(*connectors.ConnectorConfig); ok && authConfig != nil {
		accessToken = authConfig.AccessToken
	} else {
		return nil, fmt.Errorf("invalid zoom auth configuration")
	}

	if accessToken == "" {
		return nil, fmt.Errorf("zoom access token is empty")
	}

	// Marshal payload
	var body io.Reader
	if payload != nil {
		payloadBytes, err := json.Marshal(payload)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal payload: %w", err)
		}
		body = bytes.NewReader(payloadBytes)
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	// Execute request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Check status code
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("zoom API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	// Parse response
	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return result, nil
}

// buildFieldDataMap builds a map of contact field data for field lookups.
func buildFieldDataMap(contact *connectors.NormalizedContact) map[string]string {
	fieldData := map[string]string{
		"Id":        contact.ID,
		"FirstName": contact.FirstName,
		"LastName":  contact.LastName,
		"Email":     contact.Email,
		"Phone1":    contact.Phone,
		"Company":   contact.Company,
		"JobTitle":  contact.JobTitle,
		"full_name": strings.TrimSpace(contact.FirstName + " " + contact.LastName),
	}

	// Add custom fields
	if contact.CustomFields != nil {
		for key, value := range contact.CustomFields {
			fieldData[key] = fmt.Sprintf("%v", value)
		}
	}

	return fieldData
}

// Helper functions to safely extract config values with defaults
func getStringConfigValue(config map[string]interface{}, key, defaultValue string) string {
	if val, ok := config[key].(string); ok && val != "" {
		return val
	}
	return defaultValue
}

func getIntConfigValue(config map[string]interface{}, key string, defaultValue int) int {
	if val, ok := config[key].(float64); ok {
		return int(val)
	}
	if val, ok := config[key].(int); ok {
		return val
	}
	return defaultValue
}

func getFloatConfigValue(config map[string]interface{}, key string, defaultValue float64) float64 {
	if val, ok := config[key].(float64); ok {
		return val
	}
	if val, ok := config[key].(int); ok {
		return float64(val)
	}
	return defaultValue
}

func getBoolConfigValue(config map[string]interface{}, key string, defaultValue bool) bool {
	if val, ok := config[key].(bool); ok {
		return val
	}
	return defaultValue
}

func getStringFromMap(m map[string]interface{}, key, defaultValue string) string {
	if val, ok := m[key]; ok {
		switch v := val.(type) {
		case string:
			return v
		case float64:
			return fmt.Sprintf("%.0f", v)
		case int:
			return fmt.Sprintf("%d", v)
		default:
			return fmt.Sprintf("%v", v)
		}
	}
	return defaultValue
}
