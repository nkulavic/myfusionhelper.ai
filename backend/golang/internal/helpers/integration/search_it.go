package integration

import (
	"context"
	"fmt"
	"strings"

	"github.com/myfusionhelper/api/internal/connectors"
	"github.com/myfusionhelper/api/internal/helpers"
)

// NewSearchIt creates a new SearchIt helper instance
func NewSearchIt() helpers.Helper { return &SearchIt{} }

func init() {
	helpers.Register("search_it", func() helpers.Helper { return &SearchIt{} })
}

// SearchIt performs advanced contact searches with multiple criteria
type SearchIt struct{}

func (h *SearchIt) GetName() string     { return "Search It" }
func (h *SearchIt) GetType() string     { return "search_it" }
func (h *SearchIt) GetCategory() string { return "integration" }
func (h *SearchIt) GetDescription() string {
	return "Perform advanced contact searches with multiple criteria (name, email, company, custom fields)"
}
func (h *SearchIt) RequiresCRM() bool       { return true }
func (h *SearchIt) SupportedCRMs() []string { return nil }

func (h *SearchIt) GetConfigSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"search_term": map[string]interface{}{
				"type":        "string",
				"description": "Search term to find in contact fields",
			},
			"search_fields": map[string]interface{}{
				"type":        "array",
				"items":       map[string]interface{}{"type": "string"},
				"description": "Fields to search in (e.g., [\"email\", \"first_name\", \"last_name\", \"company\"])",
				"default":     []string{"email", "first_name", "last_name"},
			},
			"match_mode": map[string]interface{}{
				"type":        "string",
				"enum":        []string{"exact", "contains", "starts_with"},
				"description": "How to match the search term",
				"default":     "contains",
			},
			"tag_id": map[string]interface{}{
				"type":        "string",
				"description": "Optional: Filter results by tag",
			},
			"limit": map[string]interface{}{
				"type":        "number",
				"description": "Maximum number of contacts to return",
				"default":     50,
			},
			"apply_tag_to_results": map[string]interface{}{
				"type":        "string",
				"description": "Optional: Tag ID to apply to all found contacts",
			},
			"save_count_to": map[string]interface{}{
				"type":        "string",
				"description": "Optional field to save the result count",
			},
		},
		"required": []string{"search_term"},
	}
}

func (h *SearchIt) ValidateConfig(config map[string]interface{}) error {
	if _, ok := config["search_term"].(string); !ok || config["search_term"] == "" {
		return fmt.Errorf("search_term is required")
	}
	return nil
}

func (h *SearchIt) Execute(ctx context.Context, input helpers.HelperInput) (*helpers.HelperOutput, error) {
	searchTerm := input.Config["search_term"].(string)
	matchMode, _ := input.Config["match_mode"].(string)
	if matchMode == "" {
		matchMode = "contains"
	}

	limit := 50
	if l, ok := input.Config["limit"].(float64); ok && l > 0 {
		limit = int(l)
	}

	tagID, _ := input.Config["tag_id"].(string)
	applyTagToResults, _ := input.Config["apply_tag_to_results"].(string)
	saveCountTo, _ := input.Config["save_count_to"].(string)

	// Parse search fields
	searchFields := []string{"email", "first_name", "last_name"}
	if fields, ok := input.Config["search_fields"].([]interface{}); ok && len(fields) > 0 {
		searchFields = make([]string, 0, len(fields))
		for _, f := range fields {
			if field, ok := f.(string); ok {
				searchFields = append(searchFields, field)
			}
		}
	}

	output := &helpers.HelperOutput{
		Actions: make([]helpers.HelperAction, 0),
		Logs:    make([]string, 0),
	}

	output.Logs = append(output.Logs, fmt.Sprintf("Searching for '%s' in fields: %v (match mode: %s)", searchTerm, searchFields, matchMode))

	// Build query options
	queryOpts := connectors.QueryOptions{
		Limit: limit,
	}
	if tagID != "" {
		queryOpts.TagID = tagID
		output.Logs = append(output.Logs, fmt.Sprintf("Filtering by tag: %s", tagID))
	}

	// Get contacts (we'll filter client-side for now; future: add search to connector interface)
	results, err := input.Connector.GetContacts(ctx, queryOpts)
	if err != nil {
		output.Message = fmt.Sprintf("Search failed: %v", err)
		return output, err
	}

	output.Logs = append(output.Logs, fmt.Sprintf("Retrieved %d contacts for filtering", len(results.Contacts)))

	// Client-side filtering
	matchedContacts := make([]*connectors.NormalizedContact, 0)
	for _, contact := range results.Contacts {
		if h.contactMatches(&contact, searchTerm, searchFields, matchMode) {
			matchedContacts = append(matchedContacts, &contact)
		}
	}

	output.Logs = append(output.Logs, fmt.Sprintf("Found %d matching contacts", len(matchedContacts)))

	// Apply tag to results if configured
	if applyTagToResults != "" {
		for _, contact := range matchedContacts {
			err := input.Connector.ApplyTag(ctx, contact.ID, applyTagToResults)
			if err != nil {
				output.Logs = append(output.Logs, fmt.Sprintf("Failed to apply tag to contact %s: %v", contact.ID, err))
			}
		}
		output.Actions = append(output.Actions, helpers.HelperAction{
			Type:   "bulk_tag_applied",
			Target: applyTagToResults,
			Value:  len(matchedContacts),
		})
		output.Logs = append(output.Logs, fmt.Sprintf("Applied tag '%s' to %d contacts", applyTagToResults, len(matchedContacts)))
	}

	// Save count to field if configured
	if saveCountTo != "" && input.ContactID != "" {
		err := input.Connector.SetContactFieldValue(ctx, input.ContactID, saveCountTo, len(matchedContacts))
		if err != nil {
			output.Logs = append(output.Logs, fmt.Sprintf("Warning: Failed to save count to field '%s': %v", saveCountTo, err))
		} else {
			output.Actions = append(output.Actions, helpers.HelperAction{
				Type:   "field_updated",
				Target: saveCountTo,
				Value:  len(matchedContacts),
			})
			output.Logs = append(output.Logs, fmt.Sprintf("Saved result count to field '%s'", saveCountTo))
		}
	}

	// Build contact summary
	contactSummaries := make([]map[string]interface{}, 0, len(matchedContacts))
	for _, contact := range matchedContacts {
		contactSummaries = append(contactSummaries, map[string]interface{}{
			"id":         contact.ID,
			"first_name": contact.FirstName,
			"last_name":  contact.LastName,
			"email":      contact.Email,
			"company":    contact.Company,
		})
	}

	output.Success = true
	output.Message = fmt.Sprintf("Search completed: found %d contacts matching '%s'", len(matchedContacts), searchTerm)
	output.ModifiedData = map[string]interface{}{
		"search_term":  searchTerm,
		"match_mode":   matchMode,
		"result_count": len(matchedContacts),
		"contacts":     contactSummaries,
	}

	output.Actions = append(output.Actions, helpers.HelperAction{
		Type:   "search_executed",
		Target: searchTerm,
		Value:  len(matchedContacts),
	})

	return output, nil
}

// contactMatches checks if a contact matches the search criteria
func (h *SearchIt) contactMatches(contact *connectors.NormalizedContact, searchTerm string, searchFields []string, matchMode string) bool {
	searchTermLower := strings.ToLower(searchTerm)

	for _, field := range searchFields {
		var fieldValue string
		switch field {
		case "email":
			fieldValue = contact.Email
		case "first_name":
			fieldValue = contact.FirstName
		case "last_name":
			fieldValue = contact.LastName
		case "company":
			fieldValue = contact.Company
		case "phone":
			fieldValue = contact.Phone
		default:
			continue
		}

		fieldValueLower := strings.ToLower(fieldValue)

		matched := false
		switch matchMode {
		case "exact":
			matched = fieldValueLower == searchTermLower
		case "contains":
			matched = strings.Contains(fieldValueLower, searchTermLower)
		case "starts_with":
			matched = strings.HasPrefix(fieldValueLower, searchTermLower)
		}

		if matched {
			return true
		}
	}

	return false
}
