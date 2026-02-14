package connectors

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	hubspotBaseURL = "https://api.hubapi.com"
	hubspotSlug    = "hubspot"
)

func init() {
	Register(hubspotSlug, NewHubSpotConnector)
}

// HubSpotConnector implements CRMConnector for HubSpot
type HubSpotConnector struct {
	accessToken string
	baseURL     string
	client      *http.Client
}

// NewHubSpotConnector creates a new HubSpot CRM connector
func NewHubSpotConnector(config ConnectorConfig) (CRMConnector, error) {
	token := config.AccessToken
	if token == "" {
		token = config.APIKey
	}
	if token == "" {
		return nil, fmt.Errorf("access token or API key is required for HubSpot connector")
	}

	baseURL := config.BaseURL
	if baseURL == "" {
		baseURL = hubspotBaseURL
	}

	return &HubSpotConnector{
		accessToken: token,
		baseURL:     strings.TrimRight(baseURL, "/"),
		client:      &http.Client{Timeout: 30 * time.Second},
	}, nil
}

// ========== CONTACTS ==========

func (h *HubSpotConnector) GetContacts(ctx context.Context, opts QueryOptions) (*ContactList, error) {
	limit := opts.Limit
	if limit <= 0 {
		limit = 25
	}
	if limit > 100 {
		limit = 100
	}

	path := fmt.Sprintf("/crm/v3/objects/contacts?limit=%d&properties=firstname,lastname,email,phone,company,jobtitle,createdate,lastmodifieddate", limit)
	if opts.Cursor != "" {
		path += "&after=" + opts.Cursor
	}

	var result hubspotContactsResponse
	if err := h.doRequest(ctx, "GET", path, nil, &result); err != nil {
		return nil, err
	}

	contacts := make([]NormalizedContact, 0, len(result.Results))
	for _, hc := range result.Results {
		contacts = append(contacts, hc.toNormalized())
	}

	cl := &ContactList{
		Contacts: contacts,
		Total:    len(result.Results),
		HasMore:  result.Paging.Next.After != "",
	}
	if result.Paging.Next.After != "" {
		cl.NextCursor = result.Paging.Next.After
	}
	return cl, nil
}

func (h *HubSpotConnector) GetContact(ctx context.Context, contactID string) (*NormalizedContact, error) {
	path := fmt.Sprintf("/crm/v3/objects/contacts/%s?properties=firstname,lastname,email,phone,company,jobtitle,createdate,lastmodifieddate", contactID)

	var result hubspotContact
	if err := h.doRequest(ctx, "GET", path, nil, &result); err != nil {
		return nil, err
	}

	contact := result.toNormalized()
	return &contact, nil
}

func (h *HubSpotConnector) CreateContact(ctx context.Context, input CreateContactInput) (*NormalizedContact, error) {
	properties := map[string]string{
		"firstname": input.FirstName,
		"lastname":  input.LastName,
		"email":     input.Email,
	}
	if input.Phone != "" {
		properties["phone"] = input.Phone
	}
	if input.Company != "" {
		properties["company"] = input.Company
	}
	if input.CustomFields != nil {
		for key, value := range input.CustomFields {
			properties[key] = fmt.Sprintf("%v", value)
		}
	}

	body := map[string]interface{}{
		"properties": properties,
	}

	var result hubspotContact
	if err := h.doRequest(ctx, "POST", "/crm/v3/objects/contacts", body, &result); err != nil {
		return nil, err
	}

	contact := result.toNormalized()
	return &contact, nil
}

func (h *HubSpotConnector) UpdateContact(ctx context.Context, contactID string, updates UpdateContactInput) (*NormalizedContact, error) {
	properties := map[string]string{}

	if updates.FirstName != nil {
		properties["firstname"] = *updates.FirstName
	}
	if updates.LastName != nil {
		properties["lastname"] = *updates.LastName
	}
	if updates.Email != nil {
		properties["email"] = *updates.Email
	}
	if updates.Phone != nil {
		properties["phone"] = *updates.Phone
	}
	if updates.Company != nil {
		properties["company"] = *updates.Company
	}
	if updates.CustomFields != nil {
		for key, value := range updates.CustomFields {
			properties[key] = fmt.Sprintf("%v", value)
		}
	}

	body := map[string]interface{}{
		"properties": properties,
	}

	path := fmt.Sprintf("/crm/v3/objects/contacts/%s", contactID)
	var result hubspotContact
	if err := h.doRequest(ctx, "PATCH", path, body, &result); err != nil {
		return nil, err
	}

	contact := result.toNormalized()
	return &contact, nil
}

func (h *HubSpotConnector) DeleteContact(ctx context.Context, contactID string) error {
	path := fmt.Sprintf("/crm/v3/objects/contacts/%s", contactID)
	return h.doRequest(ctx, "DELETE", path, nil, nil)
}

// ========== TAGS ==========

// HubSpot doesn't have native tags — we map contact lists as tags
func (h *HubSpotConnector) GetTags(ctx context.Context) ([]Tag, error) {
	var result struct {
		Lists []struct {
			ListID int    `json:"listId"`
			Name   string `json:"name"`
		} `json:"lists"`
	}

	if err := h.doRequest(ctx, "GET", "/contacts/v1/lists?count=250", nil, &result); err != nil {
		return nil, err
	}

	tags := make([]Tag, 0, len(result.Lists))
	for _, l := range result.Lists {
		tags = append(tags, Tag{
			ID:   fmt.Sprintf("%d", l.ListID),
			Name: l.Name,
		})
	}
	return tags, nil
}

func (h *HubSpotConnector) ApplyTag(ctx context.Context, contactID string, tagID string) error {
	path := fmt.Sprintf("/contacts/v1/lists/%s/add", tagID)
	body := map[string]interface{}{
		"vids": []string{contactID},
	}
	return h.doRequest(ctx, "POST", path, body, nil)
}

func (h *HubSpotConnector) RemoveTag(ctx context.Context, contactID string, tagID string) error {
	path := fmt.Sprintf("/contacts/v1/lists/%s/remove", tagID)
	body := map[string]interface{}{
		"vids": []string{contactID},
	}
	return h.doRequest(ctx, "POST", path, body, nil)
}

// ========== CUSTOM FIELDS ==========

func (h *HubSpotConnector) GetCustomFields(ctx context.Context) ([]CustomField, error) {
	var result []struct {
		Name      string `json:"name"`
		Label     string `json:"label"`
		Type      string `json:"type"`
		GroupName string `json:"groupName"`
	}

	if err := h.doRequest(ctx, "GET", "/crm/v3/properties/contacts", nil, &result); err != nil {
		return nil, err
	}

	fields := make([]CustomField, 0, len(result))
	for _, f := range result {
		fields = append(fields, CustomField{
			ID:        f.Name,
			Key:       f.Name,
			Label:     f.Label,
			FieldType: f.Type,
			GroupName: f.GroupName,
		})
	}
	return fields, nil
}

func (h *HubSpotConnector) GetContactFieldValue(ctx context.Context, contactID string, fieldKey string) (interface{}, error) {
	path := fmt.Sprintf("/crm/v3/objects/contacts/%s?properties=%s", contactID, fieldKey)

	var result hubspotContact
	if err := h.doRequest(ctx, "GET", path, nil, &result); err != nil {
		return nil, err
	}

	if val, ok := result.Properties[fieldKey]; ok {
		return val, nil
	}
	return nil, nil
}

func (h *HubSpotConnector) SetContactFieldValue(ctx context.Context, contactID string, fieldKey string, value interface{}) error {
	updates := UpdateContactInput{
		CustomFields: map[string]interface{}{
			fieldKey: value,
		},
	}
	_, err := h.UpdateContact(ctx, contactID, updates)
	return err
}

// ========== AUTOMATIONS ==========

func (h *HubSpotConnector) TriggerAutomation(ctx context.Context, contactID string, automationID string) error {
	path := fmt.Sprintf("/automation/v2/workflows/%s/enrollments/contacts/%s", automationID, contactID)
	return h.doRequest(ctx, "POST", path, map[string]interface{}{}, nil)
}

func (h *HubSpotConnector) AchieveGoal(_ context.Context, _ string, _ string, _ string) error {
	return NewConnectorError(hubspotSlug, 501, "HubSpot does not support goal achievement", false)
}

// ========== MARKETING ==========

func (h *HubSpotConnector) SetOptInStatus(ctx context.Context, contactID string, optIn bool, reason string) error {
	// HubSpot manages marketing email opt-in via specific properties
	// https://developers.hubspot.com/docs/api/crm/contacts
	// hs_email_optout: true = opted out, false = opted in

	optOutValue := "true"
	if optIn {
		optOutValue = "false"
	}

	updates := map[string]interface{}{
		"properties": map[string]interface{}{
			"hs_email_optout": optOutValue,
		},
	}

	// Add reason to notes if provided
	if reason != "" {
		updates["properties"].(map[string]interface{})["hs_note"] = fmt.Sprintf("Opt-in status updated: %s. Reason: %s", optOutValue, reason)
	}

	return h.doRequest(ctx, "PATCH", "/crm/v3/objects/contacts/"+contactID, updates, nil)
}

// ========== HEALTH ==========

func (h *HubSpotConnector) TestConnection(ctx context.Context) error {
	var result map[string]interface{}
	return h.doRequest(ctx, "GET", "/crm/v3/objects/contacts?limit=1", nil, &result)
}

func (h *HubSpotConnector) GetMetadata() ConnectorMetadata {
	return ConnectorMetadata{
		PlatformSlug: hubspotSlug,
		PlatformName: "HubSpot",
		APIVersion:   "v3",
		BaseURL:      h.baseURL,
	}
}

func (h *HubSpotConnector) GetCapabilities() []Capability {
	return []Capability{
		CapContacts,
		CapTags,
		CapCustomFields,
		CapAutomations,
		CapDeals,
	}
}

// ========== INTERNAL TYPES ==========

type hubspotContact struct {
	ID         string            `json:"id"`
	Properties map[string]string `json:"properties"`
	CreatedAt  string            `json:"createdAt"`
	UpdatedAt  string            `json:"updatedAt"`
}

type hubspotContactsResponse struct {
	Results []hubspotContact `json:"results"`
	Paging  struct {
		Next struct {
			After string `json:"after"`
		} `json:"next"`
	} `json:"paging"`
}

func (hc *hubspotContact) toNormalized() NormalizedContact {
	contact := NormalizedContact{
		ID:           hc.ID,
		FirstName:    hc.Properties["firstname"],
		LastName:     hc.Properties["lastname"],
		Email:        hc.Properties["email"],
		Phone:        hc.Properties["phone"],
		Company:      hc.Properties["company"],
		JobTitle:     hc.Properties["jobtitle"],
		SourceCRM:    hubspotSlug,
		SourceID:     hc.ID,
		CustomFields: make(map[string]interface{}),
	}

	if hc.CreatedAt != "" {
		if t, err := time.Parse(time.RFC3339Nano, hc.CreatedAt); err == nil {
			contact.CreatedAt = &t
		}
	}
	if hc.UpdatedAt != "" {
		if t, err := time.Parse(time.RFC3339Nano, hc.UpdatedAt); err == nil {
			contact.UpdatedAt = &t
		}
	}

	return contact
}

// ========== HTTP HELPER ==========

func (h *HubSpotConnector) doRequest(ctx context.Context, method, path string, body interface{}, result interface{}) error {
	var bodyReader io.Reader
	if body != nil {
		bodyJSON, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = strings.NewReader(string(bodyJSON))
	}

	apiURL := h.baseURL + path
	req, err := http.NewRequestWithContext(ctx, method, apiURL, bodyReader)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+h.accessToken)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := h.client.Do(req)
	if err != nil {
		return NewConnectorError(hubspotSlug, 0, fmt.Sprintf("request failed: %v", err), true)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return NewConnectorError(hubspotSlug, resp.StatusCode, "failed to read response", true)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		retryable := resp.StatusCode == 429 || resp.StatusCode >= 500
		return NewConnectorError(hubspotSlug, resp.StatusCode,
			fmt.Sprintf("HubSpot API error (%d): %s", resp.StatusCode, string(respBody)), retryable)
	}

	if result != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("failed to parse response: %w", err)
		}
	}

	return nil
}
