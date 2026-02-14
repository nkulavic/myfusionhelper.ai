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
	ontraportBaseURL = "https://api.ontraport.com/1"
	ontraportSlug    = "ontraport"
	// Ontraport object type IDs
	ontraportContactObjectID = "0"
	ontraportTagObjectID     = "14"
)

func init() {
	Register(ontraportSlug, NewOntraportConnector)
}

// OntraportConnector implements CRMConnector for Ontraport
type OntraportConnector struct {
	appID   string
	apiKey  string
	baseURL string
	client  *http.Client
}

// NewOntraportConnector creates a new Ontraport CRM connector
func NewOntraportConnector(config ConnectorConfig) (CRMConnector, error) {
	if config.APIKey == "" {
		return nil, fmt.Errorf("API key is required for Ontraport connector")
	}
	if config.APISecret == "" {
		return nil, fmt.Errorf("App ID (api_secret) is required for Ontraport connector")
	}

	baseURL := config.BaseURL
	if baseURL == "" {
		baseURL = ontraportBaseURL
	}

	return &OntraportConnector{
		appID:   config.APISecret, // App ID stored in APISecret field
		apiKey:  config.APIKey,
		baseURL: strings.TrimRight(baseURL, "/"),
		client:  &http.Client{Timeout: 30 * time.Second},
	}, nil
}

// ========== CONTACTS ==========

func (o *OntraportConnector) GetContacts(ctx context.Context, opts QueryOptions) (*ContactList, error) {
	params := url.Values{}
	params.Set("objectID", ontraportContactObjectID)
	if opts.Limit > 0 {
		params.Set("range", fmt.Sprintf("%d", opts.Limit))
	} else {
		params.Set("range", "25")
	}
	if opts.Offset > 0 {
		params.Set("start", fmt.Sprintf("%d", opts.Offset))
	}
	if opts.Email != "" {
		params.Set("condition", fmt.Sprintf(`[{"field":{"field":"email"},"op":"=","value":{"value":"%s"}}]`, opts.Email))
	}

	var result struct {
		Data []ontraportContact `json:"data"`
		Code int                `json:"code"`
	}

	if err := o.doRequest(ctx, "GET", "/objects?"+params.Encode(), nil, &result); err != nil {
		return nil, err
	}

	contacts := make([]NormalizedContact, 0, len(result.Data))
	for _, oc := range result.Data {
		contacts = append(contacts, oc.toNormalized())
	}

	return &ContactList{
		Contacts: contacts,
		Total:    len(result.Data),
		HasMore:  len(result.Data) == opts.Limit || (opts.Limit == 0 && len(result.Data) == 25),
	}, nil
}

func (o *OntraportConnector) GetContact(ctx context.Context, contactID string) (*NormalizedContact, error) {
	params := url.Values{}
	params.Set("objectID", ontraportContactObjectID)
	params.Set("id", contactID)

	var result struct {
		Data ontraportContact `json:"data"`
	}

	if err := o.doRequest(ctx, "GET", "/object?"+params.Encode(), nil, &result); err != nil {
		return nil, err
	}

	contact := result.Data.toNormalized()
	return &contact, nil
}

func (o *OntraportConnector) CreateContact(ctx context.Context, input CreateContactInput) (*NormalizedContact, error) {
	body := map[string]interface{}{
		"objectID":  ontraportContactObjectID,
		"firstname": input.FirstName,
		"lastname":  input.LastName,
		"email":     input.Email,
	}
	if input.Phone != "" {
		body["office_phone"] = input.Phone
	}
	if input.Company != "" {
		body["company"] = input.Company
	}
	if input.CustomFields != nil {
		for key, value := range input.CustomFields {
			body[key] = value
		}
	}

	var result struct {
		Data ontraportContact `json:"data"`
	}

	if err := o.doRequest(ctx, "POST", "/objects", body, &result); err != nil {
		return nil, err
	}

	// Apply tags if provided
	for _, tagID := range input.Tags {
		_ = o.ApplyTag(ctx, result.Data.ID, tagID)
	}

	contact := result.Data.toNormalized()
	return &contact, nil
}

func (o *OntraportConnector) UpdateContact(ctx context.Context, contactID string, updates UpdateContactInput) (*NormalizedContact, error) {
	body := map[string]interface{}{
		"objectID": ontraportContactObjectID,
		"id":       contactID,
	}

	if updates.FirstName != nil {
		body["firstname"] = *updates.FirstName
	}
	if updates.LastName != nil {
		body["lastname"] = *updates.LastName
	}
	if updates.Email != nil {
		body["email"] = *updates.Email
	}
	if updates.Phone != nil {
		body["office_phone"] = *updates.Phone
	}
	if updates.Company != nil {
		body["company"] = *updates.Company
	}
	if updates.CustomFields != nil {
		for key, value := range updates.CustomFields {
			body[key] = value
		}
	}

	var result struct {
		Data ontraportContact `json:"data"`
	}

	if err := o.doRequest(ctx, "PUT", "/objects", body, &result); err != nil {
		return nil, err
	}

	contact := result.Data.toNormalized()
	return &contact, nil
}

func (o *OntraportConnector) DeleteContact(ctx context.Context, contactID string) error {
	body := map[string]interface{}{
		"objectID": ontraportContactObjectID,
		"id":       contactID,
	}
	return o.doRequest(ctx, "DELETE", "/object", body, nil)
}

// ========== TAGS ==========

func (o *OntraportConnector) GetTags(ctx context.Context) ([]Tag, error) {
	params := url.Values{}
	params.Set("objectID", ontraportTagObjectID)
	params.Set("range", "1000")

	var result struct {
		Data []struct {
			ID      string `json:"tag_id"`
			TagName string `json:"tag_name"`
		} `json:"data"`
	}

	if err := o.doRequest(ctx, "GET", "/objects?"+params.Encode(), nil, &result); err != nil {
		return nil, err
	}

	tags := make([]Tag, 0, len(result.Data))
	for _, t := range result.Data {
		tags = append(tags, Tag{
			ID:   t.ID,
			Name: t.TagName,
		})
	}
	return tags, nil
}

func (o *OntraportConnector) ApplyTag(ctx context.Context, contactID string, tagID string) error {
	body := map[string]interface{}{
		"objectID": ontraportContactObjectID,
		"ids":      contactID,
		"add_list": tagID,
	}
	return o.doRequest(ctx, "PUT", "/objects/tag", body, nil)
}

func (o *OntraportConnector) RemoveTag(ctx context.Context, contactID string, tagID string) error {
	body := map[string]interface{}{
		"objectID":    ontraportContactObjectID,
		"ids":         contactID,
		"remove_list": tagID,
	}
	return o.doRequest(ctx, "DELETE", "/objects/tag", body, nil)
}

// ========== CUSTOM FIELDS ==========

func (o *OntraportConnector) GetCustomFields(ctx context.Context) ([]CustomField, error) {
	params := url.Values{}
	params.Set("objectID", ontraportContactObjectID)

	var result struct {
		Data map[string]struct {
			Alias string `json:"alias"`
			Type  string `json:"type"`
		} `json:"data"`
	}

	if err := o.doRequest(ctx, "GET", "/objects/fieldeditor?"+params.Encode(), nil, &result); err != nil {
		return nil, err
	}

	fields := make([]CustomField, 0)
	for key, f := range result.Data {
		// Only include custom fields (f_ prefix)
		if strings.HasPrefix(key, "f") {
			fields = append(fields, CustomField{
				ID:        key,
				Key:       key,
				Label:     f.Alias,
				FieldType: f.Type,
			})
		}
	}
	return fields, nil
}

func (o *OntraportConnector) GetContactFieldValue(ctx context.Context, contactID string, fieldKey string) (interface{}, error) {
	contact, err := o.GetContact(ctx, contactID)
	if err != nil {
		return nil, err
	}

	switch fieldKey {
	case "first_name", "firstname":
		return contact.FirstName, nil
	case "last_name", "lastname":
		return contact.LastName, nil
	case "email":
		return contact.Email, nil
	case "phone", "office_phone":
		return contact.Phone, nil
	case "company":
		return contact.Company, nil
	}

	if val, ok := contact.CustomFields[fieldKey]; ok {
		return val, nil
	}
	return nil, nil
}

func (o *OntraportConnector) SetContactFieldValue(ctx context.Context, contactID string, fieldKey string, value interface{}) error {
	updates := UpdateContactInput{}

	switch fieldKey {
	case "first_name", "firstname":
		v := fmt.Sprintf("%v", value)
		updates.FirstName = &v
	case "last_name", "lastname":
		v := fmt.Sprintf("%v", value)
		updates.LastName = &v
	case "email":
		v := fmt.Sprintf("%v", value)
		updates.Email = &v
	case "phone", "office_phone":
		v := fmt.Sprintf("%v", value)
		updates.Phone = &v
	case "company":
		v := fmt.Sprintf("%v", value)
		updates.Company = &v
	default:
		updates.CustomFields = map[string]interface{}{
			fieldKey: value,
		}
	}

	_, err := o.UpdateContact(ctx, contactID, updates)
	return err
}

// ========== AUTOMATIONS ==========

func (o *OntraportConnector) TriggerAutomation(ctx context.Context, contactID string, automationID string) error {
	// Ontraport: Add contact to a sequence
	body := map[string]interface{}{
		"objectID": ontraportContactObjectID,
		"ids":      contactID,
		"add_list": automationID,
	}
	return o.doRequest(ctx, "PUT", "/objects/sequence", body, nil)
}

func (o *OntraportConnector) AchieveGoal(_ context.Context, _ string, _ string, _ string) error {
	return NewConnectorError(ontraportSlug, 501, "Ontraport does not support goal achievement", false)
}

// ========== MARKETING ==========

func (o *OntraportConnector) SetOptInStatus(ctx context.Context, contactID string, optIn bool, reason string) error {
	// Ontraport manages bulk mail status via specific fields
	// https://api.ontraport.com/doc/#update-an-object
	// bulk_mail: 0 = Unsubscribed, 1 = Single Opt-In, 2 = Double Opt-In, 3 = Pending (requires confirmation)

	var bulkMailValue string
	if optIn {
		bulkMailValue = "1" // Single Opt-In
	} else {
		bulkMailValue = "0" // Unsubscribed
	}

	updates := map[string]interface{}{
		"objectID": ontraportContactObjectID,
		"id":       contactID,
		"bulk_mail": bulkMailValue,
	}

	// Add reason to custom field if provided
	if reason != "" {
		updates["note"] = fmt.Sprintf("Opt-in status updated: %s. Reason: %s", bulkMailValue, reason)
	}

	return o.doRequest(ctx, "PUT", "/objects", updates, nil)
}

// ========== HEALTH ==========

func (o *OntraportConnector) TestConnection(ctx context.Context) error {
	params := url.Values{}
	params.Set("objectID", ontraportContactObjectID)
	params.Set("range", "1")

	var result map[string]interface{}
	return o.doRequest(ctx, "GET", "/objects?"+params.Encode(), nil, &result)
}

func (o *OntraportConnector) GetMetadata() ConnectorMetadata {
	return ConnectorMetadata{
		PlatformSlug: ontraportSlug,
		PlatformName: "Ontraport",
		APIVersion:   "v1",
		BaseURL:      o.baseURL,
	}
}

func (o *OntraportConnector) GetCapabilities() []Capability {
	return []Capability{
		CapContacts,
		CapTags,
		CapCustomFields,
		CapAutomations,
		CapDeals,
	}
}

// ========== INTERNAL TYPES ==========

type ontraportContact struct {
	ID        string `json:"id"`
	FirstName string `json:"firstname"`
	LastName  string `json:"lastname"`
	Email     string `json:"email"`
	Phone     string `json:"office_phone"`
	Company   string `json:"company"`
	DateAdded string `json:"date"`
	LastActivity string `json:"dla"`
}

func (oc *ontraportContact) toNormalized() NormalizedContact {
	contact := NormalizedContact{
		ID:           oc.ID,
		FirstName:    oc.FirstName,
		LastName:     oc.LastName,
		Email:        oc.Email,
		Phone:        oc.Phone,
		Company:      oc.Company,
		SourceCRM:    ontraportSlug,
		SourceID:     oc.ID,
		CustomFields: make(map[string]interface{}),
	}

	// Ontraport uses Unix timestamps
	if oc.DateAdded != "" {
		var ts int64
		if _, err := fmt.Sscanf(oc.DateAdded, "%d", &ts); err == nil && ts > 0 {
			t := time.Unix(ts, 0).UTC()
			contact.CreatedAt = &t
		}
	}
	if oc.LastActivity != "" {
		var ts int64
		if _, err := fmt.Sscanf(oc.LastActivity, "%d", &ts); err == nil && ts > 0 {
			t := time.Unix(ts, 0).UTC()
			contact.UpdatedAt = &t
		}
	}

	return contact
}

// ========== HTTP HELPER ==========

func (o *OntraportConnector) doRequest(ctx context.Context, method, path string, body interface{}, result interface{}) error {
	var bodyReader io.Reader
	if body != nil {
		bodyJSON, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = strings.NewReader(string(bodyJSON))
	}

	apiURL := o.baseURL + path
	req, err := http.NewRequestWithContext(ctx, method, apiURL, bodyReader)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Api-Appid", o.appID)
	req.Header.Set("Api-Key", o.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := o.client.Do(req)
	if err != nil {
		return NewConnectorError(ontraportSlug, 0, fmt.Sprintf("request failed: %v", err), true)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return NewConnectorError(ontraportSlug, resp.StatusCode, "failed to read response", true)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		retryable := resp.StatusCode == 429 || resp.StatusCode >= 500
		return NewConnectorError(ontraportSlug, resp.StatusCode,
			fmt.Sprintf("Ontraport API error (%d): %s", resp.StatusCode, string(respBody)), retryable)
	}

	if result != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("failed to parse response: %w", err)
		}
	}

	return nil
}
