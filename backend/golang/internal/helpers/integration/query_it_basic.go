package integration

import (
	"context"
	"fmt"

	"github.com/myfusionhelper/api/internal/connectors"
	"github.com/myfusionhelper/api/internal/helpers"
)

// NewQueryItBasic creates a new QueryItBasic helper instance
func NewQueryItBasic() helpers.Helper { return &QueryItBasic{} }

func init() {
	helpers.Register("query_it_basic", func() helpers.Helper { return &QueryItBasic{} })
}

// QueryItBasic performs basic contact queries in the CRM
type QueryItBasic struct{}

func (h *QueryItBasic) GetName() string     { return "Query It Basic" }
func (h *QueryItBasic) GetType() string     { return "query_it_basic" }
func (h *QueryItBasic) GetCategory() string { return "integration" }
func (h *QueryItBasic) GetDescription() string {
	return "Perform basic contact queries in the CRM (search by email, tag, field value)"
}
func (h *QueryItBasic) RequiresCRM() bool       { return true }
func (h *QueryItBasic) SupportedCRMs() []string { return nil }

func (h *QueryItBasic) GetConfigSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"query_type": map[string]interface{}{
				"type":        "string",
				"enum":        []string{"email", "tag", "all"},
				"description": "Type of query to perform",
				"default":     "all",
			},
			"email": map[string]interface{}{
				"type":        "string",
				"description": "Email address to search for (required if query_type is 'email')",
			},
			"tag_id": map[string]interface{}{
				"type":        "string",
				"description": "Tag ID to filter by (required if query_type is 'tag')",
			},
			"limit": map[string]interface{}{
				"type":        "number",
				"description": "Maximum number of contacts to return",
				"default":     100,
			},
			"save_count_to": map[string]interface{}{
				"type":        "string",
				"description": "Optional field to save the result count",
			},
		},
		"required": []string{},
	}
}

func (h *QueryItBasic) ValidateConfig(config map[string]interface{}) error {
	queryType, _ := config["query_type"].(string)
	if queryType == "" {
		queryType = "all"
	}

	if queryType == "email" {
		if _, ok := config["email"].(string); !ok || config["email"] == "" {
			return fmt.Errorf("email is required when query_type is 'email'")
		}
	}

	if queryType == "tag" {
		if _, ok := config["tag_id"].(string); !ok || config["tag_id"] == "" {
			return fmt.Errorf("tag_id is required when query_type is 'tag'")
		}
	}

	return nil
}

func (h *QueryItBasic) Execute(ctx context.Context, input helpers.HelperInput) (*helpers.HelperOutput, error) {
	queryType, _ := input.Config["query_type"].(string)
	if queryType == "" {
		queryType = "all"
	}

	limit := 100
	if l, ok := input.Config["limit"].(float64); ok && l > 0 {
		limit = int(l)
	}

	saveCountTo, _ := input.Config["save_count_to"].(string)

	output := &helpers.HelperOutput{
		Actions: make([]helpers.HelperAction, 0),
		Logs:    make([]string, 0),
	}

	// Build query options
	queryOpts := connectors.QueryOptions{
		Limit: limit,
	}

	switch queryType {
	case "email":
		email := input.Config["email"].(string)
		queryOpts.Email = email
		output.Logs = append(output.Logs, fmt.Sprintf("Querying contacts by email: %s", email))

	case "tag":
		tagID := input.Config["tag_id"].(string)
		queryOpts.TagID = tagID
		output.Logs = append(output.Logs, fmt.Sprintf("Querying contacts by tag: %s", tagID))

	case "all":
		output.Logs = append(output.Logs, fmt.Sprintf("Querying all contacts (limit: %d)", limit))
	}

	// Execute query
	results, err := input.Connector.GetContacts(ctx, queryOpts)
	if err != nil {
		output.Message = fmt.Sprintf("Query failed: %v", err)
		return output, err
	}

	output.Logs = append(output.Logs, fmt.Sprintf("Query returned %d contacts", len(results.Contacts)))

	// Save count to field if configured
	if saveCountTo != "" && input.ContactID != "" {
		err := input.Connector.SetContactFieldValue(ctx, input.ContactID, saveCountTo, len(results.Contacts))
		if err != nil {
			output.Logs = append(output.Logs, fmt.Sprintf("Warning: Failed to save count to field '%s': %v", saveCountTo, err))
		} else {
			output.Actions = append(output.Actions, helpers.HelperAction{
				Type:   "field_updated",
				Target: saveCountTo,
				Value:  len(results.Contacts),
			})
			output.Logs = append(output.Logs, fmt.Sprintf("Saved result count to field '%s'", saveCountTo))
		}
	}

	// Build contact summary
	contactSummaries := make([]map[string]interface{}, 0, len(results.Contacts))
	for _, contact := range results.Contacts {
		contactSummaries = append(contactSummaries, map[string]interface{}{
			"id":         contact.ID,
			"first_name": contact.FirstName,
			"last_name":  contact.LastName,
			"email":      contact.Email,
		})
	}

	output.Success = true
	output.Message = fmt.Sprintf("Query completed: found %d contacts", len(results.Contacts))
	output.ModifiedData = map[string]interface{}{
		"query_type":   queryType,
		"result_count": len(results.Contacts),
		"contacts":     contactSummaries,
		"has_more":     results.HasMore,
	}

	output.Actions = append(output.Actions, helpers.HelperAction{
		Type:   "query_executed",
		Target: queryType,
		Value:  len(results.Contacts),
	})

	return output, nil
}
