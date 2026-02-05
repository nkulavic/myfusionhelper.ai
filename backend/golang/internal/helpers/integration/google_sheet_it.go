package integration

import (
	"context"
	"fmt"
	"strings"

	"github.com/myfusionhelper/api/internal/helpers"
)

func init() {
	helpers.Register("google_sheet_it", func() helpers.Helper { return &GoogleSheetIt{} })
}

// GoogleSheetIt syncs contact/search data to a Google Sheet. It prepares the data
// and produces a sync request for downstream processing by the execution layer
// which handles OAuth token refresh and Google Sheets API calls.
// Ported from legacy PHP google_sheet_it helper.
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
				"description": "The sheet/tab ID within the spreadsheet",
			},
			"google_account_id": map[string]interface{}{
				"type":        "string",
				"description": "The Google account connection ID for authentication",
			},
			"search_data": map[string]interface{}{
				"type":        "string",
				"description": "Comma-separated saved search ID and user ID (e.g., 'searchId,userId')",
			},
			"translate": map[string]interface{}{
				"type":        "string",
				"enum":        []string{"true", "false"},
				"description": "Whether to translate field names to human-readable labels",
				"default":     "false",
			},
			"fields": map[string]interface{}{
				"type":        "array",
				"items":       map[string]interface{}{"type": "string"},
				"description": "Specific fields to include in the export",
			},
			"mode": map[string]interface{}{
				"type":        "string",
				"enum":        []string{"replace", "append"},
				"description": "Whether to replace all data or append to existing rows",
				"default":     "replace",
			},
		},
		"required": []string{"spreadsheet_id", "sheet_id", "google_account_id"},
	}
}

func (h *GoogleSheetIt) ValidateConfig(config map[string]interface{}) error {
	if _, ok := config["spreadsheet_id"].(string); !ok || config["spreadsheet_id"] == "" {
		return fmt.Errorf("spreadsheet_id is required")
	}
	if _, ok := config["sheet_id"].(string); !ok || config["sheet_id"] == "" {
		return fmt.Errorf("sheet_id is required")
	}
	if _, ok := config["google_account_id"].(string); !ok || config["google_account_id"] == "" {
		return fmt.Errorf("google_account_id is required")
	}
	return nil
}

func (h *GoogleSheetIt) Execute(ctx context.Context, input helpers.HelperInput) (*helpers.HelperOutput, error) {
	spreadsheetID := input.Config["spreadsheet_id"].(string)
	sheetID := input.Config["sheet_id"].(string)
	googleAccountID := input.Config["google_account_id"].(string)

	translate := false
	if t, ok := input.Config["translate"].(string); ok && t == "true" {
		translate = true
	}

	mode := "replace"
	if m, ok := input.Config["mode"].(string); ok && m != "" {
		mode = m
	}

	output := &helpers.HelperOutput{
		Actions: make([]helpers.HelperAction, 0),
		Logs:    make([]string, 0),
	}

	// Parse search data if provided
	searchID := ""
	searchUserID := ""
	if searchData, ok := input.Config["search_data"].(string); ok && searchData != "" {
		parts := strings.SplitN(searchData, ",", 2)
		if len(parts) >= 1 {
			searchID = strings.TrimSpace(parts[0])
		}
		if len(parts) >= 2 {
			searchUserID = strings.TrimSpace(parts[1])
		}
	}

	// Build the sync request for downstream processing
	syncRequest := map[string]interface{}{
		"spreadsheet_id":    spreadsheetID,
		"sheet_id":          sheetID,
		"google_account_id": googleAccountID,
		"mode":              mode,
		"translate":         translate,
		"helper_id":         input.HelperID,
		"account_id":        input.AccountID,
		"user_id":           input.UserID,
	}

	if searchID != "" {
		syncRequest["search_id"] = searchID
		syncRequest["search_user_id"] = searchUserID
	}

	// If specific fields are configured, include them
	if fields, ok := input.Config["fields"]; ok {
		syncRequest["fields"] = fields
	}

	// If a specific contact is being processed (not a bulk report), get contact data
	if input.ContactID != "" && input.ContactID != "google_report" {
		contact, err := input.Connector.GetContact(ctx, input.ContactID)
		if err != nil {
			output.Logs = append(output.Logs, fmt.Sprintf("Failed to get contact: %v", err))
		} else {
			contactRow := map[string]interface{}{
				"Id":        contact.ID,
				"FirstName": contact.FirstName,
				"LastName":  contact.LastName,
				"Email":     contact.Email,
				"Phone":     contact.Phone,
				"Company":   contact.Company,
			}
			if contact.CustomFields != nil {
				for k, v := range contact.CustomFields {
					contactRow[k] = v
				}
			}
			syncRequest["contact_data"] = contactRow
		}
	}

	output.Success = true
	output.Message = fmt.Sprintf("Google Sheet sync request prepared for spreadsheet %s", spreadsheetID)
	output.Actions = []helpers.HelperAction{
		{
			Type:   "google_sheet_sync_queued",
			Target: spreadsheetID,
			Value:  syncRequest,
		},
	}
	output.ModifiedData = syncRequest
	output.Logs = append(output.Logs, fmt.Sprintf("Google Sheet sync for spreadsheet %s, sheet %s (mode: %s)", spreadsheetID, sheetID, mode))

	return output, nil
}
