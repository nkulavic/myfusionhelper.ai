package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/myfusionhelper/api/internal/helpers"
)

// NewTrelloIt creates a new TrelloIt helper instance
func NewTrelloIt() helpers.Helper { return &TrelloIt{} }

func init() {
	helpers.Register("trello_it", func() helpers.Helper { return &TrelloIt{} })
}

// TrelloIt creates a Trello card populated with CRM contact data.
// Card name and description support placeholder interpolation for contact fields.
// Requires a Trello service connection configured with API key and token.
type TrelloIt struct{}

func (h *TrelloIt) GetName() string     { return "Trello It" }
func (h *TrelloIt) GetType() string     { return "trello_it" }
func (h *TrelloIt) GetCategory() string { return "integration" }
func (h *TrelloIt) GetDescription() string {
	return "Creates a Trello card with CRM contact data"
}
func (h *TrelloIt) RequiresCRM() bool       { return true }
func (h *TrelloIt) SupportedCRMs() []string { return nil }

func (h *TrelloIt) GetConfigSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"board_id": map[string]interface{}{
				"type":        "string",
				"description": "The Trello board ID where the card will be created",
			},
			"list_id": map[string]interface{}{
				"type":        "string",
				"description": "The Trello list ID where the card will be placed",
			},
			"card_name_template": map[string]interface{}{
				"type":        "string",
				"description": "Template for the card name. Supports {first_name}, {last_name}, {email}, {phone}, {company} placeholders",
			},
			"card_description_template": map[string]interface{}{
				"type":        "string",
				"description": "Optional template for the card description. Supports the same placeholders as card_name_template",
			},
			"apply_tag": map[string]interface{}{
				"type":        "string",
				"description": "Tag ID to apply to the contact after card creation",
			},
			"service_connection_ids": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"trello": map[string]interface{}{
						"type":        "string",
						"description": "Service connection ID for Trello",
					},
				},
				"description": "Service connection IDs for external integrations",
			},
		},
		"required": []string{"board_id", "list_id", "card_name_template"},
	}
}

func (h *TrelloIt) ValidateConfig(config map[string]interface{}) error {
	if _, ok := config["board_id"].(string); !ok || config["board_id"] == "" {
		return fmt.Errorf("board_id is required")
	}
	if _, ok := config["list_id"].(string); !ok || config["list_id"] == "" {
		return fmt.Errorf("list_id is required")
	}
	if _, ok := config["card_name_template"].(string); !ok || config["card_name_template"] == "" {
		return fmt.Errorf("card_name_template is required")
	}
	return nil
}

func (h *TrelloIt) Execute(ctx context.Context, input helpers.HelperInput) (*helpers.HelperOutput, error) {
	boardID := input.Config["board_id"].(string)
	listID := input.Config["list_id"].(string)
	cardNameTemplate := input.Config["card_name_template"].(string)

	cardDescTemplate := ""
	if cdt, ok := input.Config["card_description_template"].(string); ok {
		cardDescTemplate = cdt
	}

	output := &helpers.HelperOutput{
		Actions: make([]helpers.HelperAction, 0),
		Logs:    make([]string, 0),
	}

	// 1. Get the Trello service auth
	auth := input.ServiceAuths["trello"]
	if auth == nil {
		output.Message = "Trello connection required"
		return output, fmt.Errorf("Trello connection required")
	}

	// 2. Get contact data
	contact := input.ContactData
	if contact == nil {
		var err error
		contact, err = input.Connector.GetContact(ctx, input.ContactID)
		if err != nil {
			output.Message = fmt.Sprintf("Failed to get contact data: %v", err)
			return output, err
		}
	}

	// 3. Interpolate templates with contact data
	replacer := strings.NewReplacer(
		"{first_name}", contact.FirstName,
		"{last_name}", contact.LastName,
		"{email}", contact.Email,
		"{phone}", contact.Phone,
		"{company}", contact.Company,
	)

	cardName := replacer.Replace(cardNameTemplate)
	cardDesc := replacer.Replace(cardDescTemplate)

	output.Logs = append(output.Logs, fmt.Sprintf("Creating Trello card '%s' on board %s, list %s", cardName, boardID, listID))

	// 4. POST to Trello API to create the card
	apiURL := "https://api.trello.com/1/cards"

	params := url.Values{}
	params.Set("key", auth.APIKey)
	params.Set("token", auth.APISecret)
	params.Set("idList", listID)
	params.Set("name", cardName)
	if cardDesc != "" {
		params.Set("desc", cardDesc)
	}

	fullURL := fmt.Sprintf("%s?%s", apiURL, params.Encode())

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fullURL, nil)
	if err != nil {
		output.Message = fmt.Sprintf("Failed to create HTTP request: %v", err)
		return output, err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		output.Message = fmt.Sprintf("Trello API request failed: %v", err)
		return output, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		output.Message = fmt.Sprintf("Failed to read Trello response: %v", err)
		return output, err
	}

	if resp.StatusCode != http.StatusOK {
		output.Message = fmt.Sprintf("Trello API returned status %d: %s", resp.StatusCode, string(body))
		return output, fmt.Errorf("Trello API returned status %d", resp.StatusCode)
	}

	// 5. Parse the response for card details
	var cardResult map[string]interface{}
	if err := json.Unmarshal(body, &cardResult); err != nil {
		output.Message = fmt.Sprintf("Failed to parse Trello response: %v", err)
		return output, err
	}

	cardID := ""
	if id, ok := cardResult["id"].(string); ok {
		cardID = id
	}
	cardShortURL := ""
	if shortURL, ok := cardResult["shortUrl"].(string); ok {
		cardShortURL = shortURL
	}

	output.Logs = append(output.Logs, fmt.Sprintf("Trello card created: id=%s, url=%s", cardID, cardShortURL))

	// 6. Apply tag if configured
	if applyTag, ok := input.Config["apply_tag"].(string); ok && applyTag != "" {
		err := input.Connector.ApplyTag(ctx, input.ContactID, applyTag)
		if err != nil {
			output.Logs = append(output.Logs, fmt.Sprintf("Failed to apply tag '%s': %v", applyTag, err))
		} else {
			output.Actions = append(output.Actions, helpers.HelperAction{
				Type:   "tag_applied",
				Target: input.ContactID,
				Value:  applyTag,
			})
			output.Logs = append(output.Logs, fmt.Sprintf("Applied tag '%s' to contact %s", applyTag, input.ContactID))
		}
	}

	// 7. Return success with card details
	output.Success = true
	output.Message = fmt.Sprintf("Trello card created: %s", cardName)
	output.ModifiedData = map[string]interface{}{
		"card_id":    cardID,
		"card_url":   cardShortURL,
		"card_name":  cardName,
		"board_id":   boardID,
		"list_id":    listID,
		"contact_id": input.ContactID,
	}

	return output, nil
}
