package integration

import (
	"context"
	"fmt"
	"strings"

	"github.com/myfusionhelper/api/internal/helpers"
)

func init() {
	helpers.Register("slack_it", func() helpers.Helper { return &SlackIt{} })
}

// SlackIt sends a message to a Slack channel via webhook with contact data merge fields.
// Supports {{field}} and @field merge syntax for dynamic message content.
// Ported from legacy PHP slack_it helper.
type SlackIt struct{}

func (h *SlackIt) GetName() string     { return "Slack It" }
func (h *SlackIt) GetType() string     { return "slack_it" }
func (h *SlackIt) GetCategory() string { return "integration" }
func (h *SlackIt) GetDescription() string {
	return "Send a Slack message via webhook with contact data merge fields"
}
func (h *SlackIt) RequiresCRM() bool       { return true }
func (h *SlackIt) SupportedCRMs() []string { return nil }

func (h *SlackIt) GetConfigSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"webhook": map[string]interface{}{
				"type":        "string",
				"description": "The Slack incoming webhook URL",
			},
			"message": map[string]interface{}{
				"type":        "string",
				"description": "The message to send. Supports {{field_name}} and @field_name merge fields",
			},
			"username": map[string]interface{}{
				"type":        "string",
				"description": "The display name for the bot posting the message",
			},
			"channel": map[string]interface{}{
				"type":        "string",
				"description": "Optional Slack channel override",
			},
			"icon_emoji": map[string]interface{}{
				"type":        "string",
				"description": "Optional emoji icon for the bot (e.g., :robot_face:)",
			},
		},
		"required": []string{"webhook", "message", "username"},
	}
}

func (h *SlackIt) ValidateConfig(config map[string]interface{}) error {
	if _, ok := config["webhook"].(string); !ok || config["webhook"] == "" {
		return fmt.Errorf("webhook URL is required")
	}
	if _, ok := config["message"].(string); !ok || config["message"] == "" {
		return fmt.Errorf("message is required")
	}
	if _, ok := config["username"].(string); !ok || config["username"] == "" {
		return fmt.Errorf("username is required")
	}
	return nil
}

func (h *SlackIt) Execute(ctx context.Context, input helpers.HelperInput) (*helpers.HelperOutput, error) {
	webhook := input.Config["webhook"].(string)
	message := input.Config["message"].(string)
	username := input.Config["username"].(string)

	channel := ""
	if c, ok := input.Config["channel"].(string); ok {
		channel = c
	}

	iconEmoji := ""
	if ie, ok := input.Config["icon_emoji"].(string); ok {
		iconEmoji = ie
	}

	output := &helpers.HelperOutput{
		Logs: make([]string, 0),
	}

	// Get contact data for merge field interpolation
	contact, err := input.Connector.GetContact(ctx, input.ContactID)
	if err != nil {
		output.Message = fmt.Sprintf("Failed to get contact data: %v", err)
		return output, err
	}

	// Build field data map for merge field replacement
	fieldData := map[string]string{
		"Id":        contact.ID,
		"FirstName": contact.FirstName,
		"LastName":  contact.LastName,
		"Email":     contact.Email,
		"Phone1":    contact.Phone,
		"Company":   contact.Company,
		"JobTitle":  contact.JobTitle,
		"full_name": strings.TrimSpace(contact.FirstName + " " + contact.LastName),
	}

	// Add custom fields
	if contact.CustomFields != nil {
		for key, value := range contact.CustomFields {
			fieldData[key] = fmt.Sprintf("%v", value)
		}
	}

	// Replace {{field}} syntax
	for key, value := range fieldData {
		message = strings.ReplaceAll(message, "{{"+key+"}}", value)
	}

	// Replace @field syntax
	for key, value := range fieldData {
		message = strings.ReplaceAll(message, "@"+key, value)
	}

	// Build the Slack payload
	slackPayload := map[string]interface{}{
		"text":     message,
		"username": username,
	}
	if channel != "" {
		slackPayload["channel"] = channel
	}
	if iconEmoji != "" {
		slackPayload["icon_emoji"] = iconEmoji
	}

	output.Success = true
	output.Message = fmt.Sprintf("Slack message prepared for webhook")
	output.Actions = []helpers.HelperAction{
		{
			Type:   "webhook_queued",
			Target: webhook,
			Value: map[string]interface{}{
				"method":  "POST",
				"url":     webhook,
				"payload": slackPayload,
			},
		},
	}
	output.ModifiedData = map[string]interface{}{
		"webhook":   webhook,
		"message":   message,
		"username":  username,
		"channel":   channel,
		"payload":   slackPayload,
	}
	output.Logs = append(output.Logs, fmt.Sprintf("Slack message for contact %s: %s", input.ContactID, message))

	return output, nil
}
