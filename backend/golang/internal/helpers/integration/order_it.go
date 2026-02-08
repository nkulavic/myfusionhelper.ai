package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/myfusionhelper/api/internal/helpers"
)

func init() {
	helpers.Register("order_it", func() helpers.Helper { return &OrderIt{} })
}

// OrderIt creates a Keap order for a CRM contact via the Keap REST API.
// Requires a Keap service connection for API access since the CRMConnector
// interface does not expose order creation methods.
type OrderIt struct{}

func (h *OrderIt) GetName() string     { return "Order It" }
func (h *OrderIt) GetType() string     { return "order_it" }
func (h *OrderIt) GetCategory() string { return "integration" }
func (h *OrderIt) GetDescription() string {
	return "Creates a Keap order for a CRM contact"
}
func (h *OrderIt) RequiresCRM() bool       { return true }
func (h *OrderIt) SupportedCRMs() []string { return []string{"keap"} }

func (h *OrderIt) GetConfigSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"product_id": map[string]interface{}{
				"type":        "number",
				"description": "The Keap product ID for the order",
			},
			"quantity": map[string]interface{}{
				"type":        "number",
				"description": "Quantity of the product to order (default: 1)",
				"default":     1,
			},
			"promo_codes": map[string]interface{}{
				"type": "array",
				"items": map[string]interface{}{
					"type": "string",
				},
				"description": "Optional promotional codes to apply to the order",
			},
			"apply_tag": map[string]interface{}{
				"type":        "string",
				"description": "Tag ID to apply to the contact after order creation",
			},
		},
		"required": []string{"product_id"},
	}
}

func (h *OrderIt) ValidateConfig(config map[string]interface{}) error {
	productID, ok := config["product_id"]
	if !ok {
		return fmt.Errorf("product_id is required")
	}
	// product_id can arrive as float64 (from JSON) or as a string
	switch v := productID.(type) {
	case float64:
		if v <= 0 {
			return fmt.Errorf("product_id must be a positive number")
		}
	case string:
		if v == "" {
			return fmt.Errorf("product_id is required")
		}
		if _, err := strconv.Atoi(v); err != nil {
			return fmt.Errorf("product_id must be a valid number")
		}
	default:
		return fmt.Errorf("product_id must be a number")
	}
	return nil
}

func (h *OrderIt) Execute(ctx context.Context, input helpers.HelperInput) (*helpers.HelperOutput, error) {
	output := &helpers.HelperOutput{
		Actions: make([]helpers.HelperAction, 0),
		Logs:    make([]string, 0),
	}

	// 1. Parse product_id (handles both float64 and string from JSON)
	var productID int
	switch v := input.Config["product_id"].(type) {
	case float64:
		productID = int(v)
	case string:
		parsed, err := strconv.Atoi(v)
		if err != nil {
			output.Message = fmt.Sprintf("Invalid product_id: %v", err)
			return output, err
		}
		productID = parsed
	default:
		output.Message = "Invalid product_id type"
		return output, fmt.Errorf("invalid product_id type: %T", input.Config["product_id"])
	}

	// Parse quantity (default to 1)
	quantity := 1
	if q, ok := input.Config["quantity"].(float64); ok && q > 0 {
		quantity = int(q)
	}

	// Parse promo codes
	var promoCodes []string
	if codes, ok := input.Config["promo_codes"].([]interface{}); ok {
		for _, c := range codes {
			if code, ok := c.(string); ok && code != "" {
				promoCodes = append(promoCodes, code)
			}
		}
	}

	// 2. Get Keap service auth
	auth := input.ServiceAuths["keap"]
	if auth == nil {
		output.Message = "Keap service connection required for order creation"
		return output, fmt.Errorf("Keap service connection required for order creation")
	}

	// 3. Get contact data
	contact := input.ContactData
	if contact == nil {
		var err error
		contact, err = input.Connector.GetContact(ctx, input.ContactID)
		if err != nil {
			output.Message = fmt.Sprintf("Failed to get contact data: %v", err)
			return output, err
		}
	}

	output.Logs = append(output.Logs, fmt.Sprintf("Creating Keap order for contact %s (product: %d, qty: %d)", input.ContactID, productID, quantity))

	// 4. Convert contact ID to integer for the Keap API
	contactIDInt, err := strconv.Atoi(input.ContactID)
	if err != nil {
		output.Message = fmt.Sprintf("Invalid contact ID for Keap order (must be numeric): %v", err)
		return output, err
	}

	// 5. Build the order payload for the Keap REST API
	orderItem := map[string]interface{}{
		"product_id": productID,
		"quantity":   quantity,
	}

	orderPayload := map[string]interface{}{
		"contact_id":  contactIDInt,
		"order_items": []map[string]interface{}{orderItem},
	}

	if len(promoCodes) > 0 {
		orderPayload["promo_codes"] = promoCodes
		output.Logs = append(output.Logs, fmt.Sprintf("Applying promo codes: %v", promoCodes))
	}

	payloadBytes, err := json.Marshal(orderPayload)
	if err != nil {
		output.Message = fmt.Sprintf("Failed to build order payload: %v", err)
		return output, err
	}

	// 6. POST to the Keap orders API
	baseURL := auth.BaseURL
	if baseURL == "" {
		baseURL = "https://api.infusionsoft.com/crm/rest/v1"
	}
	apiURL := fmt.Sprintf("%s/orders", baseURL)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, bytes.NewReader(payloadBytes))
	if err != nil {
		output.Message = fmt.Sprintf("Failed to create HTTP request: %v", err)
		return output, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", auth.AccessToken))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		output.Message = fmt.Sprintf("Keap order API request failed: %v", err)
		return output, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		output.Message = fmt.Sprintf("Failed to read Keap order response: %v", err)
		return output, err
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		output.Message = fmt.Sprintf("Keap order API returned status %d: %s", resp.StatusCode, string(body))
		return output, fmt.Errorf("Keap order API returned status %d", resp.StatusCode)
	}

	// 7. Parse the response for order details
	var orderResult map[string]interface{}
	if err := json.Unmarshal(body, &orderResult); err != nil {
		output.Message = fmt.Sprintf("Failed to parse Keap order response: %v", err)
		return output, err
	}

	orderID := ""
	if id, ok := orderResult["id"].(float64); ok {
		orderID = strconv.Itoa(int(id))
	} else if id, ok := orderResult["order_id"].(float64); ok {
		orderID = strconv.Itoa(int(id))
	}

	output.Logs = append(output.Logs, fmt.Sprintf("Keap order created: order_id=%s", orderID))

	// 8. Apply tag if configured
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

	// 9. Return success with order details
	output.Success = true
	output.Message = fmt.Sprintf("Keap order created successfully (order_id: %s)", orderID)
	output.ModifiedData = map[string]interface{}{
		"order_id":    orderID,
		"product_id":  productID,
		"quantity":    quantity,
		"contact_id":  input.ContactID,
		"promo_codes": promoCodes,
	}

	return output, nil
}
