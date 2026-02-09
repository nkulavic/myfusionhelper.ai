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
	keapBaseURL = "https://api.infusionsoft.com/crm/rest/v2"
	keapSlug    = "keap"
)

func init() {
	Register(keapSlug, NewKeapConnector)
}

// KeapConnector implements CRMConnector for Keap (Infusionsoft)
type KeapConnector struct {
	accessToken string
	baseURL     string
	client      *http.Client
}

// NewKeapConnector creates a new Keap CRM connector
func NewKeapConnector(config ConnectorConfig) (CRMConnector, error) {
	if config.AccessToken == "" {
		return nil, fmt.Errorf("access token is required for Keap connector")
	}

	baseURL := config.BaseURL
	if baseURL == "" {
		baseURL = keapBaseURL
	}

	return &KeapConnector{
		accessToken: config.AccessToken,
		baseURL:     baseURL,
		client:      &http.Client{Timeout: 30 * time.Second},
	}, nil
}

// ========== CONTACTS ==========

func (k *KeapConnector) GetContacts(ctx context.Context, opts QueryOptions) (*ContactList, error) {
	params := url.Values{}
	if opts.Limit > 0 {
		params.Set("limit", fmt.Sprintf("%d", opts.Limit))
	} else {
		params.Set("limit", "25")
	}
	if opts.Offset > 0 {
		params.Set("offset", fmt.Sprintf("%d", opts.Offset))
	}
	if opts.Email != "" {
		params.Set("email", opts.Email)
	}

	endpoint := "/contacts"
	if len(params) > 0 {
		endpoint += "?" + params.Encode()
	}

	var result struct {
		Contacts []keapContact `json:"contacts"`
		Count    int           `json:"count"`
		Next     string        `json:"next,omitempty"`
	}

	if err := k.doRequest(ctx, "GET", endpoint, nil, &result); err != nil {
		return nil, err
	}

	contacts := make([]NormalizedContact, 0, len(result.Contacts))
	for _, kc := range result.Contacts {
		contacts = append(contacts, kc.toNormalized())
	}

	return &ContactList{
		Contacts:   contacts,
		Total:      result.Count,
		NextCursor: result.Next,
		HasMore:    result.Next != "",
	}, nil
}

func (k *KeapConnector) GetContact(ctx context.Context, contactID string) (*NormalizedContact, error) {
	var kc keapContact
	if err := k.doRequest(ctx, "GET", "/contacts/"+contactID, nil, &kc); err != nil {
		return nil, err
	}

	contact := kc.toNormalized()
	return &contact, nil
}

func (k *KeapConnector) CreateContact(ctx context.Context, input CreateContactInput) (*NormalizedContact, error) {
	body := map[string]interface{}{
		"given_name":  input.FirstName,
		"family_name": input.LastName,
	}

	if input.Email != "" {
		body["email_addresses"] = []map[string]string{
			{"email": input.Email, "field": "EMAIL1"},
		}
	}
	if input.Phone != "" {
		body["phone_numbers"] = []map[string]string{
			{"number": input.Phone, "field": "PHONE1"},
		}
	}
	if input.Company != "" {
		body["company"] = map[string]string{"company_name": input.Company}
	}
	if input.CustomFields != nil {
		customFields := make([]map[string]interface{}, 0)
		for key, value := range input.CustomFields {
			customFields = append(customFields, map[string]interface{}{
				"id":      key,
				"content": value,
			})
		}
		body["custom_fields"] = customFields
	}

	var kc keapContact
	if err := k.doRequest(ctx, "POST", "/contacts", body, &kc); err != nil {
		return nil, err
	}

	contact := kc.toNormalized()
	return &contact, nil
}

func (k *KeapConnector) UpdateContact(ctx context.Context, contactID string, updates UpdateContactInput) (*NormalizedContact, error) {
	body := map[string]interface{}{}

	if updates.FirstName != nil {
		body["given_name"] = *updates.FirstName
	}
	if updates.LastName != nil {
		body["family_name"] = *updates.LastName
	}
	if updates.Email != nil {
		body["email_addresses"] = []map[string]string{
			{"email": *updates.Email, "field": "EMAIL1"},
		}
	}
	if updates.Phone != nil {
		body["phone_numbers"] = []map[string]string{
			{"number": *updates.Phone, "field": "PHONE1"},
		}
	}
	if updates.CustomFields != nil {
		customFields := make([]map[string]interface{}, 0)
		for key, value := range updates.CustomFields {
			customFields = append(customFields, map[string]interface{}{
				"id":      key,
				"content": value,
			})
		}
		body["custom_fields"] = customFields
	}

	var kc keapContact
	if err := k.doRequest(ctx, "PATCH", "/contacts/"+contactID, body, &kc); err != nil {
		return nil, err
	}

	contact := kc.toNormalized()
	return &contact, nil
}

func (k *KeapConnector) DeleteContact(ctx context.Context, contactID string) error {
	return k.doRequest(ctx, "DELETE", "/contacts/"+contactID, nil, nil)
}

// ========== TAGS ==========

func (k *KeapConnector) GetTags(ctx context.Context) ([]Tag, error) {
	var result struct {
		Tags []struct {
			ID          int    `json:"id"`
			Name        string `json:"name"`
			Description string `json:"description"`
			Category    struct {
				ID   int    `json:"id"`
				Name string `json:"name"`
			} `json:"category"`
		} `json:"tags"`
	}

	if err := k.doRequest(ctx, "GET", "/tags?limit=1000", nil, &result); err != nil {
		return nil, err
	}

	tags := make([]Tag, 0, len(result.Tags))
	for _, t := range result.Tags {
		tags = append(tags, Tag{
			ID:          fmt.Sprintf("%d", t.ID),
			Name:        t.Name,
			Description: t.Description,
			Category:    t.Category.Name,
		})
	}
	return tags, nil
}

func (k *KeapConnector) ApplyTag(ctx context.Context, contactID string, tagID string) error {
	body := map[string]interface{}{
		"tagIds": []string{tagID},
	}
	return k.doRequest(ctx, "POST", "/contacts/"+contactID+"/tags", body, nil)
}

func (k *KeapConnector) RemoveTag(ctx context.Context, contactID string, tagID string) error {
	return k.doRequest(ctx, "DELETE", "/contacts/"+contactID+"/tags/"+tagID, nil, nil)
}

// ========== CUSTOM FIELDS ==========

func (k *KeapConnector) GetCustomFields(ctx context.Context) ([]CustomField, error) {
	var result struct {
		CustomFields []struct {
			ID        int    `json:"id"`
			FieldName string `json:"field_name"`
			Label     string `json:"label"`
			FieldType string `json:"field_type"`
			GroupID   int    `json:"group_id"`
		} `json:"custom_fields"`
	}

	if err := k.doRequest(ctx, "GET", "/contacts/model", nil, &result); err != nil {
		return nil, err
	}

	fields := make([]CustomField, 0, len(result.CustomFields))
	for _, f := range result.CustomFields {
		fields = append(fields, CustomField{
			ID:        fmt.Sprintf("%d", f.ID),
			Key:       f.FieldName,
			Label:     f.Label,
			FieldType: f.FieldType,
		})
	}
	return fields, nil
}

func (k *KeapConnector) GetContactFieldValue(ctx context.Context, contactID string, fieldKey string) (interface{}, error) {
	contact, err := k.GetContact(ctx, contactID)
	if err != nil {
		return nil, err
	}

	// Check standard fields first
	switch fieldKey {
	case "first_name", "given_name":
		return contact.FirstName, nil
	case "last_name", "family_name":
		return contact.LastName, nil
	case "email":
		return contact.Email, nil
	case "phone":
		return contact.Phone, nil
	case "company":
		return contact.Company, nil
	}

	// Check custom fields
	if val, ok := contact.CustomFields[fieldKey]; ok {
		return val, nil
	}

	return nil, nil
}

func (k *KeapConnector) SetContactFieldValue(ctx context.Context, contactID string, fieldKey string, value interface{}) error {
	updates := UpdateContactInput{}

	switch fieldKey {
	case "first_name", "given_name":
		v := fmt.Sprintf("%v", value)
		updates.FirstName = &v
	case "last_name", "family_name":
		v := fmt.Sprintf("%v", value)
		updates.LastName = &v
	case "email":
		v := fmt.Sprintf("%v", value)
		updates.Email = &v
	case "phone":
		v := fmt.Sprintf("%v", value)
		updates.Phone = &v
	default:
		// Treat as custom field
		updates.CustomFields = map[string]interface{}{
			fieldKey: value,
		}
	}

	_, err := k.UpdateContact(ctx, contactID, updates)
	return err
}

// ========== AUTOMATIONS ==========

func (k *KeapConnector) TriggerAutomation(ctx context.Context, contactID string, automationID string) error {
	// Keap v2 API: Add contact to a campaign sequence
	body := map[string]interface{}{
		"contact_id": contactID,
	}
	return k.doRequest(ctx, "POST", "/campaigns/"+automationID+"/sequences/1/contacts", body, nil)
}

func (k *KeapConnector) AchieveGoal(ctx context.Context, contactID string, goalName string, integration string) error {
	// Keap goal achievement via legacy API endpoint
	if integration == "" {
		integration = "mfh"
	}

	params := url.Values{}
	params.Set("contact_id", contactID)
	params.Set("call_name", goalName)
	params.Set("integration", integration)

	return k.doRequest(ctx, "POST", "/funnel/achieve?"+params.Encode(), nil, nil)
}

// ========== MARKETING ==========

func (k *KeapConnector) SetOptInStatus(ctx context.Context, contactID string, optIn bool, reason string) error {
	// Keap's email status field: https://developer.infusionsoft.com/docs/rest/#!/Contact/updatePropertiesOnContactUsingPATCH
	// email_status values: Single Opt In, Double Opt In, Confirmed, UnMarketable, NonMarketable
	// For simple opt-in automation, we use: Single Opt In (opted in) or UnMarketable (opted out)

	status := "UnMarketable"
	if optIn {
		status = "Single Opt In"
	}

	updates := map[string]interface{}{
		"email_status": status,
	}

	// Add reason to notes if provided
	if reason != "" {
		updates["notes"] = fmt.Sprintf("Opt-in status updated: %s. Reason: %s", status, reason)
	}

	return k.doRequest(ctx, "PATCH", "/contacts/"+contactID, updates, nil)
}

// ========== HEALTH ==========

func (k *KeapConnector) TestConnection(ctx context.Context) error {
	var result map[string]interface{}
	return k.doRequest(ctx, "GET", "/oauth/connect/userinfo", nil, &result)
}

func (k *KeapConnector) GetMetadata() ConnectorMetadata {
	return ConnectorMetadata{
		PlatformSlug: keapSlug,
		PlatformName: "Keap (Infusionsoft)",
		APIVersion:   "v2",
		BaseURL:      k.baseURL,
	}
}

func (k *KeapConnector) GetCapabilities() []Capability {
	return []Capability{
		CapContacts,
		CapTags,
		CapCustomFields,
		CapAutomations,
		CapGoals,
		CapDeals,
		CapEmails,
	}
}

// ========== INTERNAL TYPES ==========

type keapContact struct {
	ID          int    `json:"id"`
	GivenName   string `json:"given_name"`
	FamilyName  string `json:"family_name"`
	CompanyName string `json:"company_name,omitempty"`
	JobTitle    string `json:"job_title,omitempty"`
	Emails      []struct {
		Email string `json:"email"`
		Field string `json:"field"`
	} `json:"email_addresses"`
	Phones []struct {
		Number string `json:"number"`
		Field  string `json:"field"`
	} `json:"phone_numbers"`
	Tags []struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"tag_ids"`
	CustomFields []struct {
		ID      int         `json:"id"`
		Content interface{} `json:"content"`
	} `json:"custom_fields"`
	DateCreated string `json:"date_created"`
	LastUpdated string `json:"last_updated"`
}

func (kc *keapContact) toNormalized() NormalizedContact {
	contact := NormalizedContact{
		ID:           fmt.Sprintf("%d", kc.ID),
		FirstName:    kc.GivenName,
		LastName:     kc.FamilyName,
		Company:      kc.CompanyName,
		JobTitle:     kc.JobTitle,
		SourceCRM:    keapSlug,
		SourceID:     fmt.Sprintf("%d", kc.ID),
		CustomFields: make(map[string]interface{}),
	}

	if len(kc.Emails) > 0 {
		contact.Email = kc.Emails[0].Email
	}
	if len(kc.Phones) > 0 {
		contact.Phone = kc.Phones[0].Number
	}

	for _, tag := range kc.Tags {
		contact.Tags = append(contact.Tags, TagRef{
			ID:   fmt.Sprintf("%d", tag.ID),
			Name: tag.Name,
		})
	}

	for _, cf := range kc.CustomFields {
		contact.CustomFields[fmt.Sprintf("%d", cf.ID)] = cf.Content
	}

	if kc.DateCreated != "" {
		if t, err := time.Parse(time.RFC3339, kc.DateCreated); err == nil {
			contact.CreatedAt = &t
		}
	}
	if kc.LastUpdated != "" {
		if t, err := time.Parse(time.RFC3339, kc.LastUpdated); err == nil {
			contact.UpdatedAt = &t
		}
	}

	return contact
}

// ========== HTTP HELPER ==========

func (k *KeapConnector) doRequest(ctx context.Context, method, path string, body interface{}, result interface{}) error {
	var bodyReader io.Reader
	if body != nil {
		bodyJSON, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = strings.NewReader(string(bodyJSON))
	}

	apiURL := k.baseURL + path
	req, err := http.NewRequestWithContext(ctx, method, apiURL, bodyReader)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+k.accessToken)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := k.client.Do(req)
	if err != nil {
		return NewConnectorError(keapSlug, 0, fmt.Sprintf("request failed: %v", err), true)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return NewConnectorError(keapSlug, resp.StatusCode, "failed to read response", true)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		retryable := resp.StatusCode == 429 || resp.StatusCode >= 500
		return NewConnectorError(keapSlug, resp.StatusCode,
			fmt.Sprintf("Keap API error (%d): %s", resp.StatusCode, string(respBody)), retryable)
	}

	if result != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("failed to parse response: %w", err)
		}
	}

	return nil
}
