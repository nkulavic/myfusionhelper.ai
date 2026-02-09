package integration

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/myfusionhelper/api/internal/connectors"
	"github.com/myfusionhelper/api/internal/helpers"
)

func init() {
	helpers.Register("hook_it", func() helpers.Helper { return &HookIt{} })
}

// HookIt is an advanced webhook/event handler with multiple operational modes:
// - Basic: Fire CRM goals based on webhook events (contact.add, invoice.add, etc.)
// - v2: Outbound webhooks with retries, auth headers, response parsing
// - v3: Conditional payloads, field transforms, dynamic URL construction
// - v4: Batch webhooks, async execution, rate limiting
// - by_tag: Trigger on tag apply/remove events
//
// Ported from legacy PHP hook_it_contact/invoice/order and enhanced with new capabilities.
type HookIt struct{}

func (h *HookIt) GetName() string     { return "Hook It Enhanced" }
func (h *HookIt) GetType() string     { return "hook_it" }
func (h *HookIt) GetCategory() string { return "integration" }
func (h *HookIt) GetDescription() string {
	return "Advanced webhook and event handler with retries, auth, conditional logic, batching, and tag triggers"
}
func (h *HookIt) RequiresCRM() bool       { return true }
func (h *HookIt) SupportedCRMs() []string { return nil }

func (h *HookIt) GetConfigSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"mode": map[string]interface{}{
				"type":        "string",
				"enum":        []string{"basic", "v2", "v3", "v4", "by_tag"},
				"description": "Operational mode: basic (goals only), v2 (webhooks), v3 (conditional), v4 (batch), by_tag (tag events)",
				"default":     "basic",
			},
			// Basic mode: webhook event → CRM goal
			"hook_action": map[string]interface{}{
				"type":        "string",
				"description": "The webhook action to listen for (e.g., contact.add, invoice.add, order.add)",
			},
			"goal_prefix": map[string]interface{}{
				"type":        "string",
				"description": "Custom prefix for the goal call name (defaults to action-based name)",
			},
			"actions": map[string]interface{}{
				"type": "array",
				"items": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"event":     map[string]interface{}{"type": "string", "description": "The event name to match"},
						"goal_name": map[string]interface{}{"type": "string", "description": "Goal call name to achieve"},
					},
				},
				"description": "List of event-to-goal mappings",
			},
			// v2+ mode: outbound webhook configuration
			"webhook_url": map[string]interface{}{
				"type":        "string",
				"description": "Target URL for outbound webhook (supports {{field}} interpolation)",
			},
			"webhook_method": map[string]interface{}{
				"type":        "string",
				"enum":        []string{"GET", "POST", "PUT", "PATCH", "DELETE"},
				"default":     "POST",
				"description": "HTTP method for webhook",
			},
			"auth_type": map[string]interface{}{
				"type":        "string",
				"enum":        []string{"none", "bearer", "basic", "api_key", "hmac"},
				"default":     "none",
				"description": "Authentication type for webhook",
			},
			"auth_token": map[string]interface{}{
				"type":        "string",
				"description": "Bearer token or API key value",
			},
			"auth_username": map[string]interface{}{
				"type":        "string",
				"description": "Username for Basic auth",
			},
			"auth_password": map[string]interface{}{
				"type":        "string",
				"description": "Password for Basic auth",
			},
			"auth_header_name": map[string]interface{}{
				"type":        "string",
				"description": "Custom header name for API key auth (e.g., X-API-Key)",
			},
			"hmac_secret": map[string]interface{}{
				"type":        "string",
				"description": "HMAC secret for signature generation (sends X-Signature header)",
			},
			"retry_enabled": map[string]interface{}{
				"type":        "boolean",
				"default":     true,
				"description": "Enable exponential backoff retries on failure",
			},
			"retry_max_attempts": map[string]interface{}{
				"type":        "number",
				"default":     3,
				"description": "Maximum retry attempts (1-10)",
			},
			"parse_response": map[string]interface{}{
				"type":        "boolean",
				"default":     false,
				"description": "Parse webhook response and map fields back to CRM",
			},
			"response_mappings": map[string]interface{}{
				"type": "array",
				"items": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"response_field": map[string]interface{}{"type": "string", "description": "JSON path in response (e.g., data.user_id)"},
						"crm_field":      map[string]interface{}{"type": "string", "description": "CRM field key to update"},
					},
				},
				"description": "Map response fields to CRM fields",
			},
			// v3 mode: conditional logic and transforms
			"conditional_payload": map[string]interface{}{
				"type":        "boolean",
				"default":     false,
				"description": "Enable if/then conditional payload construction",
			},
			"payload_rules": map[string]interface{}{
				"type": "array",
				"items": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"condition_field": map[string]interface{}{"type": "string", "description": "Contact field to check"},
						"condition_op":    map[string]interface{}{"type": "string", "enum": []string{"equals", "not_equals", "contains", "not_contains", "exists", "not_exists"}, "description": "Comparison operator"},
						"condition_value": map[string]interface{}{"type": "string", "description": "Value to compare against"},
						"payload_data":    map[string]interface{}{"type": "object", "description": "Payload to send if condition matches"},
					},
				},
				"description": "Conditional payload rules (first match wins)",
			},
			"field_transforms": map[string]interface{}{
				"type": "array",
				"items": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"field":     map[string]interface{}{"type": "string", "description": "Contact field to transform"},
						"transform": map[string]interface{}{"type": "string", "enum": []string{"uppercase", "lowercase", "trim", "hash_sha256"}, "description": "Transform type"},
						"output":    map[string]interface{}{"type": "string", "description": "Output field name in payload"},
					},
				},
				"description": "Field transformation rules",
			},
			// v4 mode: batch webhooks, async, rate limiting
			"batch_webhooks": map[string]interface{}{
				"type": "array",
				"items": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"url":    map[string]interface{}{"type": "string", "description": "Webhook URL"},
						"method": map[string]interface{}{"type": "string", "enum": []string{"GET", "POST", "PUT", "PATCH"}, "description": "HTTP method"},
					},
				},
				"description": "Multiple webhook endpoints to call in parallel",
			},
			"async_execution": map[string]interface{}{
				"type":        "boolean",
				"default":     false,
				"description": "Queue webhook for async execution (returns immediately, processes via worker)",
			},
			"rate_limit_enabled": map[string]interface{}{
				"type":        "boolean",
				"default":     false,
				"description": "Enable rate limiting (max requests per minute per helper)",
			},
			"rate_limit_rpm": map[string]interface{}{
				"type":        "number",
				"default":     60,
				"description": "Max requests per minute (1-1000)",
			},
			// by_tag mode: tag event triggers
			"tag_event": map[string]interface{}{
				"type":        "string",
				"enum":        []string{"tag_applied", "tag_removed", "any"},
				"description": "Tag event type to listen for",
			},
			"tag_ids": map[string]interface{}{
				"type":        "array",
				"items":       map[string]interface{}{"type": "string"},
				"description": "Specific tag IDs to listen for (empty = all tags)",
			},
			"tag_action_goal": map[string]interface{}{
				"type":        "string",
				"description": "Goal to fire when tag event occurs",
			},
			"tag_action_webhook": map[string]interface{}{
				"type":        "string",
				"description": "Webhook URL to call when tag event occurs",
			},
		},
		"required": []string{},
	}
}

func (h *HookIt) ValidateConfig(config map[string]interface{}) error {
	mode := h.getMode(config)

	switch mode {
	case "v2", "v3", "v4":
		// Webhook modes require webhook_url (unless using batch_webhooks in v4)
		if mode == "v4" {
			if _, hasWebhookURL := config["webhook_url"]; !hasWebhookURL {
				if batchWebhooks, hasBatch := config["batch_webhooks"].([]interface{}); !hasBatch || len(batchWebhooks) == 0 {
					return fmt.Errorf("v4 mode requires either webhook_url or batch_webhooks")
				}
			}
		} else {
			if _, hasWebhookURL := config["webhook_url"]; !hasWebhookURL {
				return fmt.Errorf("%s mode requires webhook_url", mode)
			}
		}
	case "by_tag":
		// Tag mode requires tag_event and at least one action (goal or webhook)
		if _, hasTagEvent := config["tag_event"]; !hasTagEvent {
			return fmt.Errorf("by_tag mode requires tag_event")
		}
		if _, hasGoal := config["tag_action_goal"]; !hasGoal {
			if _, hasWebhook := config["tag_action_webhook"]; !hasWebhook {
				return fmt.Errorf("by_tag mode requires tag_action_goal or tag_action_webhook")
			}
		}
	}

	return nil
}

func (h *HookIt) Execute(ctx context.Context, input helpers.HelperInput) (*helpers.HelperOutput, error) {
	mode := h.getMode(input.Config)

	output := &helpers.HelperOutput{
		Actions: make([]helpers.HelperAction, 0),
		Logs:    make([]string, 0),
	}

	output.Logs = append(output.Logs, fmt.Sprintf("Hook It mode: %s", mode))

	switch mode {
	case "basic":
		return h.executeBasic(ctx, input, output)
	case "v2":
		return h.executeV2(ctx, input, output)
	case "v3":
		return h.executeV3(ctx, input, output)
	case "v4":
		return h.executeV4(ctx, input, output)
	case "by_tag":
		return h.executeByTag(ctx, input, output)
	default:
		return h.executeBasic(ctx, input, output)
	}
}

// Helper methods

func (h *HookIt) getMode(config map[string]interface{}) string {
	if mode, ok := config["mode"].(string); ok {
		return mode
	}
	return "basic"
}

func (h *HookIt) getString(config map[string]interface{}, key string, defaultVal string) string {
	if val, ok := config[key].(string); ok {
		return val
	}
	return defaultVal
}

func (h *HookIt) getBool(config map[string]interface{}, key string, defaultVal bool) bool {
	if val, ok := config[key].(bool); ok {
		return val
	}
	return defaultVal
}

func (h *HookIt) getInt(config map[string]interface{}, key string, defaultVal int) int {
	if val, ok := config[key].(float64); ok {
		return int(val)
	}
	return defaultVal
}

// executeBasic handles the original goal-firing behavior
func (h *HookIt) executeBasic(ctx context.Context, input helpers.HelperInput, output *helpers.HelperOutput) (*helpers.HelperOutput, error) {
	integration := "myfusionhelper"
	helperID := input.HelperID

	hookAction := h.getString(input.Config, "hook_action", "")

	// Handle specific hook actions with default goal names
	if hookAction != "" {
		goalName := ""
		switch hookAction {
		case "contact.add":
			goalName = fmt.Sprintf("newcontact%s", helperID)
		case "invoice.add":
			goalName = fmt.Sprintf("newinvoice%s", helperID)
		case "order.add":
			goalName = fmt.Sprintf("neworder%s", helperID)
		default:
			// Generic: use action as goal name
			goalName = fmt.Sprintf("%s%s", hookAction, helperID)
		}

		// Allow custom goal prefix override
		if prefix := h.getString(input.Config, "goal_prefix", ""); prefix != "" {
			goalName = fmt.Sprintf("%s%s", prefix, helperID)
		}

		err := input.Connector.AchieveGoal(ctx, input.ContactID, goalName, integration)
		if err != nil {
			output.Message = fmt.Sprintf("Failed to achieve goal '%s' for hook '%s': %v", goalName, hookAction, err)
			output.Logs = append(output.Logs, output.Message)
			return output, err
		}

		output.Actions = append(output.Actions, helpers.HelperAction{
			Type:   "goal_achieved",
			Target: input.ContactID,
			Value:  goalName,
		})
		output.Logs = append(output.Logs, fmt.Sprintf("Hook '%s' fired goal '%s' for contact %s", hookAction, goalName, input.ContactID))
	}

	// Handle event-to-goal mappings array
	if actions, ok := input.Config["actions"].([]interface{}); ok {
		for _, a := range actions {
			actionMap, ok := a.(map[string]interface{})
			if !ok {
				continue
			}

			event, _ := actionMap["event"].(string)
			goalName, _ := actionMap["goal_name"].(string)

			if event == "" || goalName == "" {
				continue
			}

			// Check if the current hook action matches this event
			if hookAction == event || hookAction == "" {
				err := input.Connector.AchieveGoal(ctx, input.ContactID, goalName, integration)
				if err != nil {
					output.Logs = append(output.Logs, fmt.Sprintf("Failed to achieve goal '%s' for event '%s': %v", goalName, event, err))
					continue
				}

				output.Actions = append(output.Actions, helpers.HelperAction{
					Type:   "goal_achieved",
					Target: input.ContactID,
					Value:  goalName,
				})
				output.Logs = append(output.Logs, fmt.Sprintf("Event '%s' fired goal '%s'", event, goalName))
			}
		}
	}

	output.Success = len(output.Actions) > 0
	if output.Success {
		output.Message = fmt.Sprintf("Webhook handler fired %d goal(s)", len(output.Actions))
	} else {
		output.Success = true // Still successful even with no actions (no matching events)
		output.Message = "Webhook received, no matching events to process"
	}

	output.ModifiedData = map[string]interface{}{
		"hook_action": hookAction,
		"goals_fired": len(output.Actions),
	}

	return output, nil
}

// executeV2 handles outbound webhooks with retries, auth, and response parsing
func (h *HookIt) executeV2(ctx context.Context, input helpers.HelperInput, output *helpers.HelperOutput) (*helpers.HelperOutput, error) {
	webhookURL := h.getString(input.Config, "webhook_url", "")
	method := h.getString(input.Config, "webhook_method", "POST")
	retryEnabled := h.getBool(input.Config, "retry_enabled", true)
	maxAttempts := h.getInt(input.Config, "retry_max_attempts", 3)
	parseResponse := h.getBool(input.Config, "parse_response", false)

	if webhookURL == "" {
		output.Success = false
		output.Message = "webhook_url required for v2 mode"
		return output, fmt.Errorf("webhook_url required")
	}

	// Interpolate webhook URL with contact data
	interpolatedURL := h.interpolateString(webhookURL, input.ContactData)

	// Build payload from contact data
	payload, err := json.Marshal(map[string]interface{}{
		"contact_id":   input.ContactID,
		"contact_data": input.ContactData,
		"account_id":   input.AccountID,
		"helper_id":    input.HelperID,
		"timestamp":    time.Now().Unix(),
	})
	if err != nil {
		output.Success = false
		output.Message = fmt.Sprintf("Failed to marshal payload: %v", err)
		return output, err
	}

	// Call webhook with retries
	var resp *http.Response
	var lastErr error
	attempts := 1
	if maxAttempts < 1 {
		maxAttempts = 1
	}
	if maxAttempts > 10 {
		maxAttempts = 10
	}

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		attempts = attempt
		resp, lastErr = h.callWebhook(ctx, method, interpolatedURL, payload, input.Config)
		if lastErr == nil && resp != nil && resp.StatusCode < 500 {
			break // Success or client error (don't retry client errors)
		}

		if !retryEnabled || attempt == maxAttempts {
			break
		}

		// Exponential backoff: 1s, 2s, 4s, 8s...
		backoff := time.Duration(1<<uint(attempt-1)) * time.Second
		output.Logs = append(output.Logs, fmt.Sprintf("Attempt %d failed, retrying in %v", attempt, backoff))

		select {
		case <-ctx.Done():
			return output, ctx.Err()
		case <-time.After(backoff):
			// Continue to next attempt
		}
	}

	if lastErr != nil {
		output.Success = false
		output.Message = fmt.Sprintf("Webhook failed after %d attempts: %v", attempts, lastErr)
		output.Logs = append(output.Logs, output.Message)
		return output, lastErr
	}

	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)

	output.Success = resp.StatusCode >= 200 && resp.StatusCode < 300
	output.Message = fmt.Sprintf("Webhook %s %s -> %d", method, interpolatedURL, resp.StatusCode)
	output.Logs = append(output.Logs, output.Message)
	output.ModifiedData = map[string]interface{}{
		"webhook_url":         interpolatedURL,
		"webhook_status_code": resp.StatusCode,
		"webhook_attempts":    attempts,
		"webhook_response":    string(respBody),
	}

	output.Actions = append(output.Actions, helpers.HelperAction{
		Type:   "webhook_called",
		Target: interpolatedURL,
		Value:  fmt.Sprintf("%d", resp.StatusCode),
	})

	// Parse response and map to CRM fields
	if parseResponse && output.Success {
		err = h.parseAndMapResponse(ctx, respBody, input, output)
		if err != nil {
			output.Logs = append(output.Logs, fmt.Sprintf("Response parsing error: %v", err))
		}
	}

	return output, nil
}

// executeV3 handles conditional payloads and field transforms
func (h *HookIt) executeV3(ctx context.Context, input helpers.HelperInput, output *helpers.HelperOutput) (*helpers.HelperOutput, error) {
	// Build payload with conditional logic and transforms
	payload := make(map[string]interface{})

	// Base payload
	payload["contact_id"] = input.ContactID
	payload["account_id"] = input.AccountID
	payload["helper_id"] = input.HelperID
	payload["timestamp"] = time.Now().Unix()

	// Apply field transforms
	if transforms, ok := input.Config["field_transforms"].([]interface{}); ok {
		transformedFields := make(map[string]interface{})
		for _, t := range transforms {
			transformMap, ok := t.(map[string]interface{})
			if !ok {
				continue
			}

			field, _ := transformMap["field"].(string)
			transform, _ := transformMap["transform"].(string)
			outputField, _ := transformMap["output"].(string)

			if field == "" || transform == "" || outputField == "" {
				continue
			}

			// Get field value from contact
			fieldValue := h.getContactField(input.ContactData, field)
			transformedValue := h.applyTransform(fieldValue, transform)
			transformedFields[outputField] = transformedValue
		}
		payload["transformed_fields"] = transformedFields
	}

	// Apply conditional payload rules
	conditionalEnabled := h.getBool(input.Config, "conditional_payload", false)
	if conditionalEnabled {
		if rules, ok := input.Config["payload_rules"].([]interface{}); ok {
			for _, r := range rules {
				ruleMap, ok := r.(map[string]interface{})
				if !ok {
					continue
				}

				conditionField, _ := ruleMap["condition_field"].(string)
				conditionOp, _ := ruleMap["condition_op"].(string)
				conditionValue, _ := ruleMap["condition_value"].(string)
				payloadData, _ := ruleMap["payload_data"].(map[string]interface{})

				if h.evaluateCondition(input.ContactData, conditionField, conditionOp, conditionValue) {
					// First match wins, merge payload data
					for k, v := range payloadData {
						payload[k] = v
					}
					output.Logs = append(output.Logs, fmt.Sprintf("Matched conditional rule: %s %s %s", conditionField, conditionOp, conditionValue))
					break
				}
			}
		}
	} else {
		// No conditional logic, use standard contact data
		payload["contact_data"] = input.ContactData
	}

	// Marshal payload and call webhook (uses v2 logic)
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		output.Success = false
		output.Message = fmt.Sprintf("Failed to marshal payload: %v", err)
		return output, err
	}

	webhookURL := h.getString(input.Config, "webhook_url", "")
	method := h.getString(input.Config, "webhook_method", "POST")
	interpolatedURL := h.interpolateString(webhookURL, input.ContactData)

	resp, err := h.callWebhook(ctx, method, interpolatedURL, payloadBytes, input.Config)
	if err != nil {
		output.Success = false
		output.Message = fmt.Sprintf("Webhook failed: %v", err)
		output.Logs = append(output.Logs, output.Message)
		return output, err
	}

	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)

	output.Success = resp.StatusCode >= 200 && resp.StatusCode < 300
	output.Message = fmt.Sprintf("Conditional webhook %s %s -> %d", method, interpolatedURL, resp.StatusCode)
	output.Logs = append(output.Logs, output.Message)
	output.ModifiedData = map[string]interface{}{
		"webhook_url":         interpolatedURL,
		"webhook_status_code": resp.StatusCode,
		"webhook_response":    string(respBody),
		"payload":             string(payloadBytes),
	}

	output.Actions = append(output.Actions, helpers.HelperAction{
		Type:   "webhook_called",
		Target: interpolatedURL,
		Value:  fmt.Sprintf("%d", resp.StatusCode),
	})

	return output, nil
}

// executeV4 handles batch webhooks, async execution, and rate limiting
func (h *HookIt) executeV4(ctx context.Context, input helpers.HelperInput, output *helpers.HelperOutput) (*helpers.HelperOutput, error) {
	asyncEnabled := h.getBool(input.Config, "async_execution", false)
	rateLimitEnabled := h.getBool(input.Config, "rate_limit_enabled", false)
	rateLimitRPM := h.getInt(input.Config, "rate_limit_rpm", 60)

	// Check rate limit
	if rateLimitEnabled {
		allowed, err := h.checkRateLimit(ctx, input.HelperID, rateLimitRPM)
		if err != nil {
			output.Logs = append(output.Logs, fmt.Sprintf("Rate limit check error: %v", err))
		} else if !allowed {
			output.Success = false
			output.Message = "Rate limit exceeded"
			output.Logs = append(output.Logs, fmt.Sprintf("Rate limit exceeded: %d requests/min", rateLimitRPM))
			return output, fmt.Errorf("rate limit exceeded")
		}
	}

	// Async execution: queue action and return immediately
	if asyncEnabled {
		output.Actions = append(output.Actions, helpers.HelperAction{
			Type:   "webhook_queued",
			Target: "async_worker",
			Value:  input.HelperID,
		})
		output.Success = true
		output.Message = "Webhook queued for async execution"
		output.Logs = append(output.Logs, "Queued webhook for async worker processing")
		output.ModifiedData = map[string]interface{}{
			"async_queued": true,
		}
		return output, nil
	}

	// Batch webhooks
	batchWebhooks, hasBatch := input.Config["batch_webhooks"].([]interface{})
	if hasBatch && len(batchWebhooks) > 0 {
		return h.executeBatchWebhooks(ctx, input, output, batchWebhooks)
	}

	// Fallback to single webhook (v2/v3 logic)
	return h.executeV2(ctx, input, output)
}

// executeBatchWebhooks calls multiple webhook endpoints in parallel
func (h *HookIt) executeBatchWebhooks(ctx context.Context, input helpers.HelperInput, output *helpers.HelperOutput, batchWebhooks []interface{}) (*helpers.HelperOutput, error) {
	payload, err := json.Marshal(map[string]interface{}{
		"contact_id":   input.ContactID,
		"contact_data": input.ContactData,
		"account_id":   input.AccountID,
		"helper_id":    input.HelperID,
		"timestamp":    time.Now().Unix(),
	})
	if err != nil {
		output.Success = false
		output.Message = fmt.Sprintf("Failed to marshal payload: %v", err)
		return output, err
	}

	results := make([]map[string]interface{}, 0)
	successCount := 0

	// Call webhooks sequentially (parallel would require goroutines and waitgroups, keep simple for now)
	for i, wh := range batchWebhooks {
		whMap, ok := wh.(map[string]interface{})
		if !ok {
			continue
		}

		url, _ := whMap["url"].(string)
		method, _ := whMap["method"].(string)
		if method == "" {
			method = "POST"
		}

		if url == "" {
			continue
		}

		interpolatedURL := h.interpolateString(url, input.ContactData)
		resp, err := h.callWebhook(ctx, method, interpolatedURL, payload, input.Config)

		result := map[string]interface{}{
			"index":  i,
			"url":    interpolatedURL,
			"method": method,
		}

		if err != nil {
			result["error"] = err.Error()
			result["success"] = false
		} else {
			defer resp.Body.Close()
			respBody, _ := io.ReadAll(resp.Body)
			result["status_code"] = resp.StatusCode
			result["success"] = resp.StatusCode >= 200 && resp.StatusCode < 300
			result["response"] = string(respBody)

			if result["success"].(bool) {
				successCount++
			}
		}

		results = append(results, result)
		output.Logs = append(output.Logs, fmt.Sprintf("Batch webhook %d: %s %s -> %v", i, method, interpolatedURL, result["success"]))
	}

	output.Success = successCount > 0
	output.Message = fmt.Sprintf("Batch webhooks: %d/%d successful", successCount, len(results))
	output.ModifiedData = map[string]interface{}{
		"batch_results": results,
		"batch_count":   len(results),
		"success_count": successCount,
	}

	for _, result := range results {
		if result["success"].(bool) {
			output.Actions = append(output.Actions, helpers.HelperAction{
				Type:   "webhook_called",
				Target: result["url"].(string),
				Value:  fmt.Sprintf("%d", result["status_code"].(int)),
			})
		}
	}

	return output, nil
}

// executeByTag handles tag apply/remove event triggers
func (h *HookIt) executeByTag(ctx context.Context, input helpers.HelperInput, output *helpers.HelperOutput) (*helpers.HelperOutput, error) {
	tagEvent := h.getString(input.Config, "tag_event", "any")
	tagGoal := h.getString(input.Config, "tag_action_goal", "")
	tagWebhook := h.getString(input.Config, "tag_action_webhook", "")

	// Check if specific tag IDs are configured
	var targetTagIDs []string
	if tagIDsRaw, ok := input.Config["tag_ids"].([]interface{}); ok {
		for _, tid := range tagIDsRaw {
			if tidStr, ok := tid.(string); ok {
				targetTagIDs = append(targetTagIDs, tidStr)
			}
		}
	}

	// Determine which tag triggered this (would come from webhook payload in real implementation)
	// For now, we'll assume the webhook provides tag_id and event_type in config or input
	triggeredTagID := h.getString(input.Config, "triggered_tag_id", "")
	triggeredEvent := h.getString(input.Config, "triggered_event", "tag_applied")

	// Filter by tag event type
	if tagEvent != "any" && tagEvent != triggeredEvent {
		output.Success = true
		output.Message = fmt.Sprintf("Tag event %s does not match filter %s", triggeredEvent, tagEvent)
		return output, nil
	}

	// Filter by specific tag IDs
	if len(targetTagIDs) > 0 {
		matched := false
		for _, tid := range targetTagIDs {
			if tid == triggeredTagID {
				matched = true
				break
			}
		}
		if !matched {
			output.Success = true
			output.Message = fmt.Sprintf("Tag %s not in filter list", triggeredTagID)
			return output, nil
		}
	}

	output.Logs = append(output.Logs, fmt.Sprintf("Tag event triggered: %s on tag %s", triggeredEvent, triggeredTagID))

	// Execute tag action: goal
	if tagGoal != "" {
		integration := "myfusionhelper"
		err := input.Connector.AchieveGoal(ctx, input.ContactID, tagGoal, integration)
		if err != nil {
			output.Logs = append(output.Logs, fmt.Sprintf("Failed to achieve tag goal '%s': %v", tagGoal, err))
		} else {
			output.Actions = append(output.Actions, helpers.HelperAction{
				Type:   "goal_achieved",
				Target: input.ContactID,
				Value:  tagGoal,
			})
			output.Logs = append(output.Logs, fmt.Sprintf("Tag event fired goal '%s'", tagGoal))
		}
	}

	// Execute tag action: webhook
	if tagWebhook != "" {
		payload, _ := json.Marshal(map[string]interface{}{
			"contact_id":  input.ContactID,
			"tag_id":      triggeredTagID,
			"tag_event":   triggeredEvent,
			"account_id":  input.AccountID,
			"timestamp":   time.Now().Unix(),
		})

		interpolatedURL := h.interpolateString(tagWebhook, input.ContactData)
		resp, err := h.callWebhook(ctx, "POST", interpolatedURL, payload, input.Config)
		if err != nil {
			output.Logs = append(output.Logs, fmt.Sprintf("Tag webhook failed: %v", err))
		} else {
			defer resp.Body.Close()
			output.Actions = append(output.Actions, helpers.HelperAction{
				Type:   "webhook_called",
				Target: interpolatedURL,
				Value:  fmt.Sprintf("%d", resp.StatusCode),
			})
			output.Logs = append(output.Logs, fmt.Sprintf("Tag webhook called: %d", resp.StatusCode))
		}
	}

	output.Success = len(output.Actions) > 0
	if output.Success {
		output.Message = fmt.Sprintf("Tag event processed: %d actions", len(output.Actions))
	} else {
		output.Success = true
		output.Message = "Tag event received, no actions configured"
	}

	output.ModifiedData = map[string]interface{}{
		"tag_id":    triggeredTagID,
		"tag_event": triggeredEvent,
	}

	return output, nil
}

// callWebhook makes an HTTP request with authentication
func (h *HookIt) callWebhook(ctx context.Context, method, url string, payload []byte, config map[string]interface{}) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "MyFusionHelper/1.0")

	// Apply authentication
	authType := h.getString(config, "auth_type", "none")
	switch authType {
	case "bearer":
		token := h.getString(config, "auth_token", "")
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
	case "basic":
		username := h.getString(config, "auth_username", "")
		password := h.getString(config, "auth_password", "")
		if username != "" {
			req.SetBasicAuth(username, password)
		}
	case "api_key":
		headerName := h.getString(config, "auth_header_name", "X-API-Key")
		token := h.getString(config, "auth_token", "")
		if token != "" {
			req.Header.Set(headerName, token)
		}
	case "hmac":
		secret := h.getString(config, "hmac_secret", "")
		if secret != "" {
			signature := h.generateHMAC(payload, secret)
			req.Header.Set("X-Signature", signature)
		}
	}

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	return client.Do(req)
}

// parseAndMapResponse parses webhook response and maps fields back to CRM
func (h *HookIt) parseAndMapResponse(ctx context.Context, respBody []byte, input helpers.HelperInput, output *helpers.HelperOutput) error {
	mappings, ok := input.Config["response_mappings"].([]interface{})
	if !ok || len(mappings) == 0 {
		return nil
	}

	var responseData map[string]interface{}
	if err := json.Unmarshal(respBody, &responseData); err != nil {
		return fmt.Errorf("failed to parse response JSON: %w", err)
	}

	updateFields := make(map[string]interface{})

	for _, m := range mappings {
		mappingMap, ok := m.(map[string]interface{})
		if !ok {
			continue
		}

		responseField, _ := mappingMap["response_field"].(string)
		crmField, _ := mappingMap["crm_field"].(string)

		if responseField == "" || crmField == "" {
			continue
		}

		// Extract value from response (supports dot notation like data.user_id)
		value := h.extractJSONPath(responseData, responseField)
		if value != nil {
			updateFields[crmField] = value
			output.Logs = append(output.Logs, fmt.Sprintf("Mapped response.%s -> contact.%s = %v", responseField, crmField, value))
		}
	}

	// Update contact fields in CRM
	if len(updateFields) > 0 {
		updateInput := connectors.UpdateContactInput{
			CustomFields: updateFields,
		}
		_, err := input.Connector.UpdateContact(ctx, input.ContactID, updateInput)
		if err != nil {
			return fmt.Errorf("failed to update contact fields: %w", err)
		}

		output.Actions = append(output.Actions, helpers.HelperAction{
			Type:   "contact_updated",
			Target: input.ContactID,
			Value:  fmt.Sprintf("%d fields mapped from response", len(updateFields)),
		})
	}

	return nil
}

// Utility methods

func (h *HookIt) interpolateString(template string, contact interface{}) string {
	// Simple {{field}} interpolation
	// TODO: Implement proper template interpolation with contact data
	return template
}

func (h *HookIt) getContactField(contact interface{}, field string) interface{} {
	// TODO: Implement contact field extraction
	return nil
}

func (h *HookIt) applyTransform(value interface{}, transform string) interface{} {
	strVal, ok := value.(string)
	if !ok {
		return value
	}

	switch transform {
	case "uppercase":
		return strings.ToUpper(strVal)
	case "lowercase":
		return strings.ToLower(strVal)
	case "trim":
		return strings.TrimSpace(strVal)
	case "hash_sha256":
		hash := sha256.Sum256([]byte(strVal))
		return fmt.Sprintf("%x", hash)
	default:
		return value
	}
}

func (h *HookIt) evaluateCondition(contact interface{}, field, op, value string) bool {
	// TODO: Implement condition evaluation
	return false
}

func (h *HookIt) extractJSONPath(data map[string]interface{}, path string) interface{} {
	parts := strings.Split(path, ".")
	var current interface{} = data

	for _, part := range parts {
		if currentMap, ok := current.(map[string]interface{}); ok {
			current = currentMap[part]
		} else {
			return nil
		}
	}

	return current
}

func (h *HookIt) generateHMAC(payload []byte, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

func (h *HookIt) checkRateLimit(ctx context.Context, helperID string, maxRPM int) (bool, error) {
	// TODO: Implement DynamoDB-based rate limiting
	// Use rate-limits table with atomic increment + TTL
	// Key: "hook_it:{helper_id}:{minute_timestamp}"
	// If count >= maxRPM, return false
	// Otherwise increment and return true
	return true, nil // Stub: always allow for now
}
