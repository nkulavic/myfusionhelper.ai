// +build integration

package data

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIPLocation_Integration_RealAPI_PublicIP(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	h := &IPLocation{}
	mockConnector := &mockConnectorForIPLocation{
		fieldValues: map[string]interface{}{
			"ip_address": "8.8.8.8", // Google DNS - known public IP
		},
		setFieldErrors: map[string]error{},
	}

	input := helpers.HelperInput{
		ContactID: "contact_123",
		Connector: mockConnector,
		Config: map[string]interface{}{
			"ip_field":      "ip_address",
			"city_field":    "city",
			"state_field":   "state",
			"country_field": "country",
			"zip_field":     "zip",
		},
	}

	output, err := h.Execute(context.Background(), input)

	require.NoError(t, err)
	assert.True(t, output.Success)
	assert.NotEmpty(t, output.Message)

	// Verify actions were created for each configured field
	assert.GreaterOrEqual(t, len(output.Actions), 1, "Should have at least one field_updated action")

	// Verify ModifiedData contains expected fields
	require.NotNil(t, output.ModifiedData)
	assert.Equal(t, "8.8.8.8", output.ModifiedData["ip_address"])
	assert.NotEmpty(t, output.ModifiedData["country"])
	assert.Equal(t, "US", output.ModifiedData["country_code"])

	// Verify logs
	assert.NotEmpty(t, output.Logs)
	assert.Contains(t, output.Logs[0], "Looking up IP address: 8.8.8.8")

	// Verify connector methods were called
	assert.Equal(t, 4, mockConnector.setFieldCount, "Should have called SetContactFieldValue for city, state, country, zip")
}

func TestIPLocation_Integration_RealAPI_AlternateIP(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	h := &IPLocation{}
	mockConnector := &mockConnectorForIPLocation{
		fieldValues: map[string]interface{}{
			"ip_address": "1.1.1.1", // Cloudflare DNS
		},
		setFieldErrors: map[string]error{},
	}

	input := helpers.HelperInput{
		ContactID: "contact_456",
		Connector: mockConnector,
		Config: map[string]interface{}{
			"ip_field":      "ip_address",
			"city_field":    "city",
			"country_field": "country",
		},
	}

	output, err := h.Execute(context.Background(), input)

	require.NoError(t, err)
	assert.True(t, output.Success)

	// Verify ModifiedData
	require.NotNil(t, output.ModifiedData)
	assert.Equal(t, "1.1.1.1", output.ModifiedData["ip_address"])
	assert.NotEmpty(t, output.ModifiedData["country"])

	// Should have set city and country (state and zip not configured)
	assert.LessOrEqual(t, mockConnector.setFieldCount, 2)
}

func TestIPLocation_Integration_RealAPI_PrivateIP(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	h := &IPLocation{}
	mockConnector := &mockConnectorForIPLocation{
		fieldValues: map[string]interface{}{
			"ip_address": "192.168.1.1", // Private IP
		},
		setFieldErrors: map[string]error{},
	}

	input := helpers.HelperInput{
		ContactID: "contact_789",
		Connector: mockConnector,
		Config: map[string]interface{}{
			"ip_field":   "ip_address",
			"city_field": "city",
		},
	}

	output, err := h.Execute(context.Background(), input)

	// Private IPs may fail or return limited data
	// The API typically returns an error for private/reserved IPs
	if err != nil {
		assert.Contains(t, err.Error(), "IP lookup failed")
	} else {
		// If it succeeds, it might return "private range" or similar
		assert.NotNil(t, output)
	}
}
