// +build integration

package data

import (
	"context"
	"testing"

	"github.com/myfusionhelper/api/internal/helpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLastOpenIt_Integration_WithEmailStats(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	h := &LastOpenIt{}
	mockConnector := &mockConnectorForLastOpenIt{
		fieldValues: map[string]interface{}{
			"Email": "integration@example.com",
			"_email_stats.integration@example.com.LastOpenDate": "2024-02-08T10:30:00Z",
		},
	}

	input := helpers.HelperInput{
		ContactID: "integration_contact_1",
		Connector: mockConnector,
		Config: map[string]interface{}{
			"save_to": "last_click_date",
		},
	}

	output, err := h.Execute(context.Background(), input)

	require.NoError(t, err)
	assert.True(t, output.Success)
	assert.NotEmpty(t, output.Message)
	assert.Len(t, output.Actions, 1)
	assert.Equal(t, "field_updated", output.Actions[0].Type)
	assert.Equal(t, "last_click_date", output.Actions[0].Target)
	assert.Equal(t, "2024-02-08T10:30:00Z", output.Actions[0].Value)
	assert.Equal(t, 1, mockConnector.setFieldCount)
}

func TestLastOpenIt_Integration_FallbackGenericField(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	h := &LastOpenIt{}
	mockConnector := &mockConnectorForLastOpenIt{
		fieldValues: map[string]interface{}{
			"Email":         "fallback@example.com",
			"LastOpenDate": "2024-02-08T11:45:00Z",
		},
		getFieldError: map[string]error{
			"_email_stats.fallback@example.com.LastOpenDate": fmt.Errorf("stats not supported"),
		},
	}

	input := helpers.HelperInput{
		ContactID: "integration_contact_2",
		Connector: mockConnector,
		Config: map[string]interface{}{
			"save_to":     "click_date",
			"email_field": "Email",
		},
	}

	output, err := h.Execute(context.Background(), input)

	require.NoError(t, err)
	assert.True(t, output.Success)
	assert.Equal(t, "2024-02-08T11:45:00Z", mockConnector.setFieldCalls["click_date"])
}

func TestLastOpenIt_Integration_NoDataAvailable(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	h := &LastOpenIt{}
	mockConnector := &mockConnectorForLastOpenIt{
		fieldValues: map[string]interface{}{
			"Email": "nodata@example.com",
		},
		getFieldError: map[string]error{
			"_email_stats.nodata@example.com.LastOpenDate": fmt.Errorf("not found"),
			"LastOpenDate": fmt.Errorf("not found"),
		},
	}

	input := helpers.HelperInput{
		ContactID: "integration_contact_3",
		Connector: mockConnector,
		Config: map[string]interface{}{
			"save_to": "last_click_date",
		},
	}

	output, err := h.Execute(context.Background(), input)

	require.NoError(t, err)
	assert.True(t, output.Success)
	assert.Contains(t, output.Message, "Could not retrieve email click stats")
	assert.Empty(t, output.Actions)
}
