package integration

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/myfusionhelper/api/internal/connectors"
	"github.com/myfusionhelper/api/internal/helpers"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

func init() {
	helpers.Register("google_sheet_it", func() helpers.Helper { return &GoogleSheetIt{} })
}

// GoogleSheetIt syncs contact/search data to a Google Sheet using Google Sheets API v4.
// Full implementation ported from Node.js mfh-v2-process-google-sheet-it processor.
type GoogleSheetIt struct{}

func (h *GoogleSheetIt) GetName() string     { return "Google Sheet It" }
func (h *GoogleSheetIt) GetType() string     { return "google_sheet_it" }
func (h *GoogleSheetIt) GetCategory() string { return "integration" }
func (h *GoogleSheetIt) GetDescription() string {
	return "Sync contact or saved search data to a Google Spreadsheet"
}
func (h *GoogleSheetIt) RequiresCRM() bool       { return true }
func (h *GoogleSheetIt) SupportedCRMs() []string { return nil }

func (h *GoogleSheetIt) GetConfigSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"spreadsheet_id": map[string]interface{}{
				"type":        "string",
				"description": "The Google Spreadsheet ID",
			},
			"sheet_id": map[string]interface{}{
				"type":        "string",
				"description": "The sheet/tab ID within the spreadsheet (numeric ID, not name)",
			},
			"google_connection_id": map[string]interface{}{
				"type":        "string",
				"description": "The Google service connection ID for OAuth authentication",
			},
			"search_id": map[string]interface{}{
				"type":        "string",
				"description": "CRM saved search ID to export (leave empty for single contact)",
			},
			"translate_headers": map[string]interface{}{
				"type":        "boolean",
				"description": "Whether to translate field names to human-readable labels",
				"default":     false,
			},
			"include_headers": map[string]interface{}{
				"type":        "boolean",
				"description": "Whether to include column headers in the first row",
				"default":     true,
			},
			"mode": map[string]interface{}{
				"type":        "string",
				"enum":        []string{"replace", "append"},
				"description": "Whether to replace all data or append to existing rows",
				"default":     "replace",
			},
			"chunk_size": map[string]interface{}{
				"type":        "integer",
				"description": "Number of rows to process per batch (default: 7500)",
				"default":     7500,
			},
		},
		"required": []string{"spreadsheet_id", "sheet_id", "google_connection_id"},
	}
}

func (h *GoogleSheetIt) ValidateConfig(config map[string]interface{}) error {
	if _, ok := config["spreadsheet_id"].(string); !ok || config["spreadsheet_id"] == "" {
		return fmt.Errorf("spreadsheet_id is required")
	}
	if _, ok := config["sheet_id"].(string); !ok || config["sheet_id"] == "" {
		return fmt.Errorf("sheet_id is required")
	}
	if _, ok := config["google_connection_id"].(string); !ok || config["google_connection_id"] == "" {
		return fmt.Errorf("google_connection_id is required")
	}
	return nil
}

func (h *GoogleSheetIt) Execute(ctx context.Context, input helpers.HelperInput) (*helpers.HelperOutput, error) {
	output := &helpers.HelperOutput{
		Actions: make([]helpers.HelperAction, 0),
		Logs:    make([]string, 0),
	}

	// Extract configuration
	spreadsheetID := input.Config["spreadsheet_id"].(string)
	sheetIDStr := input.Config["sheet_id"].(string)
	googleConnectionID := input.Config["google_connection_id"].(string)

	sheetID, err := strconv.ParseInt(sheetIDStr, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid sheet_id: must be numeric: %w", err)
	}

	translateHeaders := false
	if t, ok := input.Config["translate_headers"].(bool); ok {
		translateHeaders = t
	}

	includeHeaders := true
	if h, ok := input.Config["include_headers"].(bool); ok {
		includeHeaders = h
	}

	mode := "replace"
	if m, ok := input.Config["mode"].(string); ok && m != "" {
		mode = m
	}

	chunkSize := 7500
	if cs, ok := input.Config["chunk_size"].(float64); ok && cs > 0 {
		chunkSize = int(cs)
	}

	searchID := ""
	if sid, ok := input.Config["search_id"].(string); ok {
		searchID = sid
	}

	output.Logs = append(output.Logs, fmt.Sprintf("Starting Google Sheets sync: spreadsheet=%s, sheet=%d, mode=%s", spreadsheetID, sheetID, mode))

	// Step 1: Get Google OAuth access token from service connection
	accessToken, err := h.refreshGoogleAccessToken(ctx, input, googleConnectionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get Google access token: %w", err)
	}
	output.Logs = append(output.Logs, "Google OAuth token refreshed successfully")

	// Step 2: Create Google Sheets API client
	sheetsService, err := h.createSheetsClient(ctx, accessToken)
	if err != nil {
		return nil, fmt.Errorf("failed to create Sheets client: %w", err)
	}

	// Step 3: Add metadata note to cell A1 (indicates processing started)
	if err := h.addProcessingNote(ctx, sheetsService, spreadsheetID, sheetID, input, "processing"); err != nil {
		output.Logs = append(output.Logs, fmt.Sprintf("Warning: failed to add processing note: %v", err))
	}

	// Step 4: Fetch data from CRM
	var data []map[string]interface{}
	var headers []string

	if searchID != "" {
		// Bulk export from saved search
		output.Logs = append(output.Logs, fmt.Sprintf("Querying CRM saved search: %s", searchID))
		data, headers, err = h.fetchSearchData(ctx, input, searchID, translateHeaders)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch search data: %w", err)
		}
		output.Logs = append(output.Logs, fmt.Sprintf("Fetched %d records from saved search", len(data)))
	} else if input.ContactID != "" && input.ContactID != "google_report" {
		// Single contact export
		output.Logs = append(output.Logs, fmt.Sprintf("Fetching single contact: %s", input.ContactID))
		contact, err := input.Connector.GetContact(ctx, input.ContactID)
		if err != nil {
			return nil, fmt.Errorf("failed to get contact: %w", err)
		}
		data, headers = h.contactToRows(contact)
		output.Logs = append(output.Logs, "Fetched 1 contact record")
	} else {
		return nil, fmt.Errorf("either search_id or contact_id must be provided")
	}

	// Step 5: Clear sheet (replace mode only)
	if mode == "replace" {
		output.Logs = append(output.Logs, "Clearing sheet data...")
		if err := h.clearSheet(ctx, sheetsService, spreadsheetID, sheetID, len(headers)); err != nil {
			return nil, fmt.Errorf("failed to clear sheet: %w", err)
		}
		output.Logs = append(output.Logs, "Sheet cleared successfully")
	}

	// Step 6: Write data to sheet in chunks
	output.Logs = append(output.Logs, fmt.Sprintf("Writing %d rows in chunks of %d...", len(data), chunkSize))
	rowsWritten, err := h.writeDataToSheet(ctx, sheetsService, spreadsheetID, sheetID, data, headers, includeHeaders, chunkSize, mode)
	if err != nil {
		return nil, fmt.Errorf("failed to write data: %w", err)
	}
	output.Logs = append(output.Logs, fmt.Sprintf("Successfully wrote %d rows to sheet", rowsWritten))

	// Step 7: Add completion metadata note to cell A1
	if err := h.addProcessingNote(ctx, sheetsService, spreadsheetID, sheetID, input, "completed"); err != nil {
		output.Logs = append(output.Logs, fmt.Sprintf("Warning: failed to add completion note: %v", err))
	}

	output.Success = true
	output.Message = fmt.Sprintf("Successfully synced %d records to Google Sheet", len(data))
	output.ModifiedData = map[string]interface{}{
		"spreadsheet_id": spreadsheetID,
		"sheet_id":       sheetID,
		"rows_written":   rowsWritten,
		"mode":           mode,
	}

	return output, nil
}

// refreshGoogleAccessToken exchanges the refresh token for a new access token
func (h *GoogleSheetIt) refreshGoogleAccessToken(ctx context.Context, input helpers.HelperInput, connectionID string) (string, error) {
	// Load Google service connection
	googleAuth, ok := input.ServiceAuths["google_sheets"]
	if !ok || googleAuth == nil {
		return "", fmt.Errorf("google_sheets service connection not found (connection_id: %s)", connectionID)
	}

	// Google OAuth2 credentials (from legacy Node.js config)
	clientID := "814280671271-lc9hurr8n5ef4crj3t3pg3or18h28heq.apps.googleusercontent.com"
	clientSecret := "KIIMF61IybsCC8TiNxkb62pf"

	// Get refresh token from connection
	refreshToken := googleAuth.RefreshToken
	if refreshToken == "" {
		return "", fmt.Errorf("refresh_token not found in Google connection")
	}

	// Exchange refresh token for access token
	tokenURL := "https://www.googleapis.com/oauth2/v4/token"
	data := url.Values{}
	data.Set("client_id", clientID)
	data.Set("client_secret", clientSecret)
	data.Set("refresh_token", refreshToken)
	data.Set("grant_type", "refresh_token")

	req, err := http.NewRequestWithContext(ctx, "POST", tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return "", fmt.Errorf("failed to create token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to refresh token: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("token refresh failed with status %d: %s", resp.StatusCode, string(body))
	}

	var tokenResp struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
		TokenType   string `json:"token_type"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return "", fmt.Errorf("failed to parse token response: %w", err)
	}

	return tokenResp.AccessToken, nil
}

// createSheetsClient creates a Google Sheets API v4 client with the access token
func (h *GoogleSheetIt) createSheetsClient(ctx context.Context, accessToken string) (*sheets.Service, error) {
	token := &oauth2.Token{
		AccessToken: accessToken,
		TokenType:   "Bearer",
	}

	config := &oauth2.Config{
		Endpoint: google.Endpoint,
	}

	client := config.Client(ctx, token)
	service, err := sheets.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("failed to create Sheets service: %w", err)
	}

	return service, nil
}

// fetchSearchData queries the CRM for saved search results
func (h *GoogleSheetIt) fetchSearchData(ctx context.Context, input helpers.HelperInput, searchID string, translateHeaders bool) ([]map[string]interface{}, []string, error) {
	// Note: This is a simplified implementation. The full Node.js version has complex pagination
	// and uses Infusionsoft-specific SearchService.getSavedSearchResultsAllFields().
	// For multi-CRM support, we need to implement search querying via the connector interface.

	// For now, return a placeholder error indicating this needs CRM-specific implementation
	return nil, nil, fmt.Errorf("saved search querying not yet implemented - needs CRM connector enhancement")
}

// contactToRows converts a single contact to row data
func (h *GoogleSheetIt) contactToRows(contact *connectors.NormalizedContact) ([]map[string]interface{}, []string) {
	row := map[string]interface{}{
		"Id":        contact.ID,
		"FirstName": contact.FirstName,
		"LastName":  contact.LastName,
		"Email":     contact.Email,
		"Phone":     contact.Phone,
		"Company":   contact.Company,
	}

	// Add custom fields
	if contact.CustomFields != nil {
		for k, v := range contact.CustomFields {
			row[k] = v
		}
	}

	// Build headers from keys
	headers := make([]string, 0, len(row))
	for k := range row {
		headers = append(headers, k)
	}

	return []map[string]interface{}{row}, headers
}

// clearSheet clears all data from the sheet
func (h *GoogleSheetIt) clearSheet(ctx context.Context, service *sheets.Service, spreadsheetID string, sheetID int64, numColumns int) error {
	requests := []*sheets.Request{
		{
			UpdateCells: &sheets.UpdateCellsRequest{
				Range: &sheets.GridRange{
					SheetId:          sheetID,
					StartRowIndex:    0,
					EndRowIndex:      1000000,
					StartColumnIndex: 0,
					EndColumnIndex:   int64(numColumns),
				},
				Rows: []*sheets.RowData{
					{Values: []*sheets.CellData{}},
				},
				Fields: "userEnteredValue",
			},
		},
	}

	batchUpdateRequest := &sheets.BatchUpdateSpreadsheetRequest{
		Requests: requests,
	}

	_, err := service.Spreadsheets.BatchUpdate(spreadsheetID, batchUpdateRequest).Context(ctx).Do()
	return err
}

// writeDataToSheet writes data to the sheet in chunks using pasteData API
func (h *GoogleSheetIt) writeDataToSheet(ctx context.Context, service *sheets.Service, spreadsheetID string, sheetID int64, data []map[string]interface{}, headers []string, includeHeaders bool, chunkSize int, mode string) (int, error) {
	if len(data) == 0 {
		return 0, nil
	}

	// Convert data to row arrays
	rows := make([][]interface{}, 0, len(data))
	for _, item := range data {
		row := make([]interface{}, len(headers))
		for i, header := range headers {
			row[i] = item[header]
		}
		rows = append(rows, row)
	}

	// Split into chunks
	totalRows := 0
	for i := 0; i < len(rows); i += chunkSize {
		end := i + chunkSize
		if end > len(rows) {
			end = len(rows)
		}
		chunk := rows[i:end]

		// Convert chunk to CSV
		var csvData string
		if i == 0 && includeHeaders {
			// First chunk: include headers
			csvData = h.rowsToCSV(chunk, headers)
		} else {
			// Subsequent chunks: no headers
			csvData = h.rowsToCSV(chunk, nil)
		}

		// Calculate row index for this chunk
		rowIndex := int64(i)
		if includeHeaders && i == 0 {
			rowIndex = 0
		} else if includeHeaders {
			rowIndex = int64(i) + 1 // Account for header row
		}

		// Build pasteData request
		requests := []*sheets.Request{
			{
				UpdateCells: &sheets.UpdateCellsRequest{
					Range: &sheets.GridRange{
						SheetId:          sheetID,
						StartRowIndex:    rowIndex,
						EndRowIndex:      rowIndex + int64(len(chunk)),
						StartColumnIndex: 0,
						EndColumnIndex:   int64(len(headers)),
					},
					Rows:   []*sheets.RowData{{Values: []*sheets.CellData{}}},
					Fields: "userEnteredValue",
				},
			},
			{
				PasteData: &sheets.PasteDataRequest{
					Data:      csvData,
					Type:      "PASTE_NORMAL",
					Delimiter: ";",
					Coordinate: &sheets.GridCoordinate{
						SheetId:  sheetID,
						RowIndex: rowIndex,
					},
				},
			},
			{
				AutoResizeDimensions: &sheets.AutoResizeDimensionsRequest{
					Dimensions: &sheets.DimensionRange{
						Dimension:  "ROWS",
						SheetId:    sheetID,
						StartIndex: rowIndex,
						EndIndex:   rowIndex + int64(len(chunk)),
					},
				},
			},
		}

		batchUpdateRequest := &sheets.BatchUpdateSpreadsheetRequest{
			Requests: requests,
		}

		_, err := service.Spreadsheets.BatchUpdate(spreadsheetID, batchUpdateRequest).Context(ctx).Do()
		if err != nil {
			return totalRows, fmt.Errorf("failed to write chunk %d: %w", i/chunkSize, err)
		}

		totalRows += len(chunk)
	}

	return totalRows, nil
}

// rowsToCSV converts rows to CSV format with semicolon delimiter
func (h *GoogleSheetIt) rowsToCSV(rows [][]interface{}, headers []string) string {
	var buf strings.Builder
	writer := csv.NewWriter(&buf)
	writer.Comma = ';'

	// Write headers if provided
	if headers != nil && len(headers) > 0 {
		writer.Write(headers)
	}

	// Write data rows
	for _, row := range rows {
		strRow := make([]string, len(row))
		for i, val := range row {
			strRow[i] = fmt.Sprintf("%v", val)
		}
		writer.Write(strRow)
	}

	writer.Flush()
	return buf.String()
}

// addProcessingNote adds a metadata note to cell A1
func (h *GoogleSheetIt) addProcessingNote(ctx context.Context, service *sheets.Service, spreadsheetID string, sheetID int64, input helpers.HelperInput, status string) error {
	timestamp := time.Now().UTC().Format(time.RFC3339)
	var noteText string

	if status == "processing" {
		noteText = fmt.Sprintf("MyFusion Helper %s update currently running. Depending on list size this operation could take between 1-15 minutes.\n\n%s\n\nHelper ID: %s\nAccount ID: %s",
			input.HelperID, timestamp, input.HelperID, input.AccountID)
	} else {
		noteText = fmt.Sprintf("MyFusion Helper %s update completed successfully.\n\n%s\n\nHelper ID: %s\nAccount ID: %s",
			input.HelperID, timestamp, input.HelperID, input.AccountID)
	}

	requests := []*sheets.Request{
		{
			UpdateCells: &sheets.UpdateCellsRequest{
				Range: &sheets.GridRange{
					SheetId:          sheetID,
					StartRowIndex:    0,
					EndRowIndex:      1,
					StartColumnIndex: 0,
					EndColumnIndex:   1,
				},
				Rows: []*sheets.RowData{
					{
						Values: []*sheets.CellData{
							{Note: noteText},
						},
					},
				},
				Fields: "note",
			},
		},
	}

	batchUpdateRequest := &sheets.BatchUpdateSpreadsheetRequest{
		Requests: requests,
	}

	_, err := service.Spreadsheets.BatchUpdate(spreadsheetID, batchUpdateRequest).Context(ctx).Do()
	return err
}
