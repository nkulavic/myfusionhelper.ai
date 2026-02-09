package connectors

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	acSlug = "activecampaign"
)

func init() {
	Register(acSlug, NewActiveCampaignConnector)
}

// ActiveCampaignConnector implements CRMConnector for ActiveCampaign
type ActiveCampaignConnector struct {
	apiKey  string
	baseURL string
	client  *http.Client
}

// NewActiveCampaignConnector creates a new ActiveCampaign CRM connector
func NewActiveCampaignConnector(config ConnectorConfig) (CRMConnector, error) {
	if config.APIKey == "" {
		return nil, fmt.Errorf("API key is required for ActiveCampaign connector")
	}
	if config.BaseURL == "" {
		return nil, fmt.Errorf("base URL (account URL) is required for ActiveCampaign connector")
	}

	// Ensure base URL includes /api/3
	baseURL := strings.TrimRight(config.BaseURL, "/")
	if !strings.HasSuffix(baseURL, "/api/3") {
		baseURL += "/api/3"
	}

	return &ActiveCampaignConnector{
		apiKey:  config.APIKey,
		baseURL: baseURL,
		client:  &http.Client{Timeout: 30 * time.Second},
	}, nil
}

// ========== CONTACTS ==========

func (a *ActiveCampaignConnector) GetContacts(ctx context.Context, opts QueryOptions) (*ContactList, error) {
	params := url.Values{}
	if opts.Limit > 0 {
		params.Set("limit", fmt.Sprintf("%d", opts.Limit))
	} else {
		params.Set("limit", "20")
	}
	if opts.Offset > 0 {
		params.Set("offset", fmt.Sprintf("%d", opts.Offset))
	}
	if opts.Email != "" {
		params.Set("email", opts.Email)
	}

	var result struct {
		Contacts []acContact `json:"contacts"`
		Meta     struct {
			Total string `json:"total"`
		} `json:"meta"`
	}

	if err := a.doRequest(ctx, "GET", "/contacts?"+params.Encode(), nil, &result); err != nil {
		return nil, err
	}

	contacts := make([]NormalizedContact, 0, len(result.Contacts))
	for _, ac := range result.Contacts {
		contacts = append(contacts, ac.toNormalized())
	}

	total := 0
	fmt.Sscanf(result.Meta.Total, "%d", &total)

	return &ContactList{
		Contacts: contacts,
		Total:    total,
		HasMore:  len(result.Contacts) == opts.Limit || (opts.Limit == 0 && len(result.Contacts) == 20),
	}, nil
}

func (a *ActiveCampaignConnector) GetContact(ctx context.Context, contactID string) (*NormalizedContact, error) {
	var result struct {
		Contact acContact `json:"contact"`
	}
	if err := a.doRequest(ctx, "GET", "/contacts/"+contactID, nil, &result); err != nil {
		return nil, err
	}

	contact := result.Contact.toNormalized()
	return &contact, nil
}

func (a *ActiveCampaignConnector) CreateContact(ctx context.Context, input CreateContactInput) (*NormalizedContact, error) {
	body := map[string]interface{}{
		"contact": map[string]interface{}{
			"firstName": input.FirstName,
			"lastName":  input.LastName,
			"email":     input.Email,
			"phone":     input.Phone,
		},
	}

	var result struct {
		Contact acContact `json:"contact"`
	}
	if err := a.doRequest(ctx, "POST", "/contacts", body, &result); err != nil {
		return nil, err
	}

	// Apply tags if provided
	for _, tagID := range input.Tags {
		_ = a.ApplyTag(ctx, result.Contact.ID, tagID)
	}

	// Set custom fields if provided
	if input.CustomFields != nil {
		for key, value := range input.CustomFields {
			fieldBody := map[string]interface{}{
				"fieldValue": map[string]interface{}{
					"contact": result.Contact.ID,
					"field":   key,
					"value":   value,
				},
			}
			_ = a.doRequest(ctx, "POST", "/fieldValues", fieldBody, nil)
		}
	}

	contact := result.Contact.toNormalized()
	return &contact, nil
}

func (a *ActiveCampaignConnector) UpdateContact(ctx context.Context, contactID string, updates UpdateContactInput) (*NormalizedContact, error) {
	contactData := map[string]interface{}{}

	if updates.FirstName != nil {
		contactData["firstName"] = *updates.FirstName
	}
	if updates.LastName != nil {
		contactData["lastName"] = *updates.LastName
	}
	if updates.Email != nil {
		contactData["email"] = *updates.Email
	}
	if updates.Phone != nil {
		contactData["phone"] = *updates.Phone
	}

	body := map[string]interface{}{
		"contact": contactData,
	}

	var result struct {
		Contact acContact `json:"contact"`
	}
	if err := a.doRequest(ctx, "PUT", "/contacts/"+contactID, body, &result); err != nil {
		return nil, err
	}

	// Update custom fields
	if updates.CustomFields != nil {
		for key, value := range updates.CustomFields {
			fieldBody := map[string]interface{}{
				"fieldValue": map[string]interface{}{
					"contact": contactID,
					"field":   key,
					"value":   value,
				},
			}
			_ = a.doRequest(ctx, "POST", "/fieldValues", fieldBody, nil)
		}
	}

	contact := result.Contact.toNormalized()
	return &contact, nil
}

func (a *ActiveCampaignConnector) DeleteContact(ctx context.Context, contactID string) error {
	return a.doRequest(ctx, "DELETE", "/contacts/"+contactID, nil, nil)
}

// ========== TAGS ==========

func (a *ActiveCampaignConnector) GetTags(ctx context.Context) ([]Tag, error) {
	var result struct {
		Tags []struct {
			ID          string `json:"id"`
			Tag         string `json:"tag"`
			Description string `json:"description"`
			TagType     string `json:"tagType"`
		} `json:"tags"`
	}

	if err := a.doRequest(ctx, "GET", "/tags?limit=100", nil, &result); err != nil {
		return nil, err
	}

	tags := make([]Tag, 0, len(result.Tags))
	for _, t := range result.Tags {
		tags = append(tags, Tag{
			ID:          t.ID,
			Name:        t.Tag,
			Description: t.Description,
			Category:    t.TagType,
		})
	}
	return tags, nil
}

func (a *ActiveCampaignConnector) ApplyTag(ctx context.Context, contactID string, tagID string) error {
	body := map[string]interface{}{
		"contactTag": map[string]interface{}{
			"contact": contactID,
			"tag":     tagID,
		},
	}
	return a.doRequest(ctx, "POST", "/contactTags", body, nil)
}

func (a *ActiveCampaignConnector) RemoveTag(ctx context.Context, contactID string, tagID string) error {
	// AC requires the contactTag ID, not the tag ID directly
	// First find the contactTag association
	var result struct {
		ContactTags []struct {
			ID  string `json:"id"`
			Tag string `json:"tag"`
		} `json:"contactTags"`
	}

	if err := a.doRequest(ctx, "GET", "/contacts/"+contactID+"/contactTags", nil, &result); err != nil {
		return err
	}

	for _, ct := range result.ContactTags {
		if ct.Tag == tagID {
			return a.doRequest(ctx, "DELETE", "/contactTags/"+ct.ID, nil, nil)
		}
	}

	return NewConnectorError(acSlug, 404, "tag not found on contact", false)
}

// ========== CUSTOM FIELDS ==========

func (a *ActiveCampaignConnector) GetCustomFields(ctx context.Context) ([]CustomField, error) {
	var result struct {
		Fields []struct {
			ID       string   `json:"id"`
			Title    string   `json:"title"`
			Perstag  string   `json:"perstag"`
			Type     string   `json:"type"`
			Options  []string `json:"options"`
			Defval   string   `json:"defval"`
		} `json:"fields"`
	}

	if err := a.doRequest(ctx, "GET", "/fields?limit=100", nil, &result); err != nil {
		return nil, err
	}

	fields := make([]CustomField, 0, len(result.Fields))
	for _, f := range result.Fields {
		fields = append(fields, CustomField{
			ID:           f.ID,
			Key:          f.Perstag,
			Label:        f.Title,
			FieldType:    f.Type,
			Options:      f.Options,
			DefaultValue: f.Defval,
		})
	}
	return fields, nil
}

func (a *ActiveCampaignConnector) GetContactFieldValue(ctx context.Context, contactID string, fieldKey string) (interface{}, error) {
	contact, err := a.GetContact(ctx, contactID)
	if err != nil {
		return nil, err
	}

	switch fieldKey {
	case "first_name", "firstName":
		return contact.FirstName, nil
	case "last_name", "lastName":
		return contact.LastName, nil
	case "email":
		return contact.Email, nil
	case "phone":
		return contact.Phone, nil
	}

	if val, ok := contact.CustomFields[fieldKey]; ok {
		return val, nil
	}
	return nil, nil
}

func (a *ActiveCampaignConnector) SetContactFieldValue(ctx context.Context, contactID string, fieldKey string, value interface{}) error {
	updates := UpdateContactInput{}

	switch fieldKey {
	case "first_name", "firstName":
		v := fmt.Sprintf("%v", value)
		updates.FirstName = &v
	case "last_name", "lastName":
		v := fmt.Sprintf("%v", value)
		updates.LastName = &v
	case "email":
		v := fmt.Sprintf("%v", value)
		updates.Email = &v
	case "phone":
		v := fmt.Sprintf("%v", value)
		updates.Phone = &v
	default:
		// Custom field via fieldValues endpoint
		body := map[string]interface{}{
			"fieldValue": map[string]interface{}{
				"contact": contactID,
				"field":   fieldKey,
				"value":   value,
			},
		}
		return a.doRequest(ctx, "POST", "/fieldValues", body, nil)
	}

	_, err := a.UpdateContact(ctx, contactID, updates)
	return err
}

// ========== AUTOMATIONS ==========

func (a *ActiveCampaignConnector) TriggerAutomation(ctx context.Context, contactID string, automationID string) error {
	body := map[string]interface{}{
		"contactAutomation": map[string]interface{}{
			"contact":    contactID,
			"automation": automationID,
		},
	}
	return a.doRequest(ctx, "POST", "/contactAutomations", body, nil)
}

func (a *ActiveCampaignConnector) AchieveGoal(_ context.Context, _ string, _ string, _ string) error {
	return NewConnectorError(acSlug, 501, "ActiveCampaign does not support goal achievement", false)
}

// ========== MARKETING ==========

func (a *ActiveCampaignConnector) SetOptInStatus(ctx context.Context, contactID string, optIn bool, reason string) error {
	// ActiveCampaign manages email status via contact fields
	// https://developers.activecampaign.com/reference/update-a-contact-new
	// Fields: 0 = Unconfirmed, 1 = Active, 2 = Unsubscribed, 3 = Bounced

	var fieldValue string
	if optIn {
		fieldValue = "1" // Active
	} else {
		fieldValue = "2" // Unsubscribed
	}

	updates := map[string]interface{}{
		"contact": map[string]interface{}{
			"fieldValues": []map[string]interface{}{
				{
					"field": "email_status",
					"value": fieldValue,
				},
			},
		},
	}

	// Add reason to note if provided
	if reason != "" {
		updates["contact"].(map[string]interface{})["note"] = fmt.Sprintf("Opt-in status updated: %s. Reason: %s", fieldValue, reason)
	}

	return a.doRequest(ctx, "PUT", "/contacts/"+contactID, updates, nil)
}

// ========== HEALTH ==========

func (a *ActiveCampaignConnector) TestConnection(ctx context.Context) error {
	var result map[string]interface{}
	return a.doRequest(ctx, "GET", "/users/me", nil, &result)
}

func (a *ActiveCampaignConnector) GetMetadata() ConnectorMetadata {
	return ConnectorMetadata{
		PlatformSlug: acSlug,
		PlatformName: "ActiveCampaign",
		APIVersion:   "v3",
		BaseURL:      a.baseURL,
	}
}

func (a *ActiveCampaignConnector) GetCapabilities() []Capability {
	return []Capability{
		CapContacts,
		CapTags,
		CapCustomFields,
		CapAutomations,
		CapDeals,
		CapEmails,
	}
}

// ========== INTERNAL TYPES ==========

type acContact struct {
	ID        string `json:"id"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Email     string `json:"email"`
	Phone     string `json:"phone"`
	CDate     string `json:"cdate"`
	UDate     string `json:"udate"`
}

func (ac *acContact) toNormalized() NormalizedContact {
	contact := NormalizedContact{
		ID:           ac.ID,
		FirstName:    ac.FirstName,
		LastName:     ac.LastName,
		Email:        ac.Email,
		Phone:        ac.Phone,
		SourceCRM:    acSlug,
		SourceID:     ac.ID,
		CustomFields: make(map[string]interface{}),
	}

	if ac.CDate != "" {
		if t, err := time.Parse("2006-01-02T15:04:05-07:00", ac.CDate); err == nil {
			contact.CreatedAt = &t
		}
	}
	if ac.UDate != "" {
		if t, err := time.Parse("2006-01-02T15:04:05-07:00", ac.UDate); err == nil {
			contact.UpdatedAt = &t
		}
	}

	return contact
}

// ========== HTTP HELPER ==========

func (a *ActiveCampaignConnector) doRequest(ctx context.Context, method, path string, body interface{}, result interface{}) error {
	var bodyReader io.Reader
	if body != nil {
		bodyJSON, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = strings.NewReader(string(bodyJSON))
	}

	apiURL := a.baseURL + path
	req, err := http.NewRequestWithContext(ctx, method, apiURL, bodyReader)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Api-Token", a.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := a.client.Do(req)
	if err != nil {
		return NewConnectorError(acSlug, 0, fmt.Sprintf("request failed: %v", err), true)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return NewConnectorError(acSlug, resp.StatusCode, "failed to read response", true)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		retryable := resp.StatusCode == 429 || resp.StatusCode >= 500
		return NewConnectorError(acSlug, resp.StatusCode,
			fmt.Sprintf("ActiveCampaign API error (%d): %s", resp.StatusCode, string(respBody)), retryable)
	}

	if result != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("failed to parse response: %w", err)
		}
	}

	return nil
}
