package google

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// SheetsClient is a client for Google Sheets API v4
type SheetsClient struct {
	httpClient  *http.Client
	accessToken string
	baseURL     string
}

// NewSheetsClient creates a new Google Sheets API client
func NewSheetsClient(accessToken string) *SheetsClient {
	return &SheetsClient{
		httpClient:  &http.Client{Timeout: 30 * time.Second},
		accessToken: accessToken,
		baseURL:     "https://sheets.googleapis.com/v4/spreadsheets",
	}
}

// WorksheetMetadata represents metadata about a worksheet
type WorksheetMetadata struct {
	SheetID    int64                  `json:"sheetId"`
	Title      string                 `json:"title"`
	Index      int                    `json:"index"`
	SheetType  string                 `json:"sheetType"`
	GridProps  map[string]interface{} `json:"gridProperties,omitempty"`
	Properties map[string]interface{} `json:"properties,omitempty"`
}

// GetWorksheetByID retrieves metadata for a specific worksheet by sheet ID
func (c *SheetsClient) GetWorksheetByID(ctx context.Context, spreadsheetID, sheetID string) (*WorksheetMetadata, error) {
	url := fmt.Sprintf("%s/%s?includeGridData=false", c.baseURL, spreadsheetID)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.accessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get spreadsheet: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	var result struct {
		Sheets []struct {
			Properties WorksheetMetadata `json:"properties"`
		} `json:"sheets"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Find the sheet with matching ID
	for _, sheet := range result.Sheets {
		if fmt.Sprintf("%d", sheet.Properties.SheetID) == sheetID {
			return &sheet.Properties, nil
		}
	}

	return nil, fmt.Errorf("sheet with ID %s not found", sheetID)
}

// ClearWorksheet clears all rows in the specified worksheet
func (c *SheetsClient) ClearWorksheet(ctx context.Context, spreadsheetID, sheetID string) error {
	// First get the worksheet to validate it exists and get its grid properties
	metadata, err := c.GetWorksheetByID(ctx, spreadsheetID, sheetID)
	if err != nil {
		return fmt.Errorf("failed to get worksheet metadata: %w", err)
	}

	// Convert sheetID string to int64
	var sheetIDInt int64
	_, err = fmt.Sscanf(sheetID, "%d", &sheetIDInt)
	if err != nil {
		return fmt.Errorf("invalid sheet ID format: %w", err)
	}

	// Build a DeleteRange request to clear all content
	requestBody := map[string]interface{}{
		"requests": []map[string]interface{}{
			{
				"updateCells": map[string]interface{}{
					"range": map[string]interface{}{
						"sheetId": metadata.SheetID,
					},
					"fields": "*",
				},
			},
		},
	}

	return c.batchUpdate(ctx, spreadsheetID, requestBody)
}

// WriteRows writes rows to the specified worksheet using AppendCells
func (c *SheetsClient) WriteRows(ctx context.Context, spreadsheetID, sheetID string, rows [][]interface{}) error {
	if len(rows) == 0 {
		return nil
	}

	// Convert sheetID string to int64
	var sheetIDInt int64
	_, err := fmt.Sscanf(sheetID, "%d", &sheetIDInt)
	if err != nil {
		return fmt.Errorf("invalid sheet ID format: %w", err)
	}

	// Build cell data for each row
	rowData := make([]map[string]interface{}, 0, len(rows))
	for _, row := range rows {
		values := make([]map[string]interface{}, 0, len(row))
		for _, cell := range row {
			cellValue := map[string]interface{}{
				"userEnteredValue": formatCellValue(cell),
			}
			values = append(values, cellValue)
		}
		rowData = append(rowData, map[string]interface{}{
			"values": values,
		})
	}

	// Build AppendCells request
	requestBody := map[string]interface{}{
		"requests": []map[string]interface{}{
			{
				"appendCells": map[string]interface{}{
					"sheetId": sheetIDInt,
					"rows":    rowData,
					"fields":  "*",
				},
			},
		},
	}

	return c.batchUpdate(ctx, spreadsheetID, requestBody)
}

// formatCellValue converts a Go value to a Google Sheets cell value format
func formatCellValue(value interface{}) map[string]interface{} {
	switch v := value.(type) {
	case string:
		return map[string]interface{}{"stringValue": v}
	case int, int8, int16, int32, int64:
		return map[string]interface{}{"numberValue": v}
	case float32, float64:
		return map[string]interface{}{"numberValue": v}
	case bool:
		return map[string]interface{}{"boolValue": v}
	default:
		// Default to string representation
		return map[string]interface{}{"stringValue": fmt.Sprintf("%v", v)}
	}
}

// batchUpdate executes a batchUpdate request
func (c *SheetsClient) batchUpdate(ctx context.Context, spreadsheetID string, requestBody map[string]interface{}) error {
	url := fmt.Sprintf("%s/%s:batchUpdate", c.baseURL, spreadsheetID)

	bodyJSON, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(bodyJSON))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.accessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	return nil
}
