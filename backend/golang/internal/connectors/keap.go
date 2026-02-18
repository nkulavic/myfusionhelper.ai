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

// ========== Flexible JSON helpers ==========
// The Keap API is inconsistent with types (e.g. id can be int or string).
// We unmarshal into map[string]interface{} and extract values flexibly.

// jsonStr extracts a string from a map, converting from other types if needed.
func jsonStr(m map[string]interface{}, key string) string {
	v, ok := m[key]
	if !ok || v == nil {
		return ""
	}
	switch val := v.(type) {
	case string:
		return val
	case float64:
		if val == float64(int(val)) {
			return fmt.Sprintf("%d", int(val))
		}
		return fmt.Sprintf("%g", val)
	case json.Number:
		return val.String()
	case bool:
		if val {
			return "true"
		}
		return "false"
	default:
		return fmt.Sprintf("%v", val)
	}
}

// jsonArr extracts a JSON array as []interface{} from a map.
func jsonArr(m map[string]interface{}, key string) []interface{} {
	v, ok := m[key]
	if !ok || v == nil {
		return nil
	}
	arr, ok := v.([]interface{})
	if !ok {
		return nil
	}
	return arr
}

// jsonObj extracts a nested object as map[string]interface{} from a map.
func jsonObj(m map[string]interface{}, key string) map[string]interface{} {
	v, ok := m[key]
	if !ok || v == nil {
		return nil
	}
	obj, ok := v.(map[string]interface{})
	if !ok {
		return nil
	}
	return obj
}

// asMap converts an interface{} to map[string]interface{} if possible.
func asMap(v interface{}) map[string]interface{} {
	if v == nil {
		return nil
	}
	m, ok := v.(map[string]interface{})
	if !ok {
		return nil
	}
	return m
}

// ========== CONTACTS ==========

func (k *KeapConnector) GetContacts(ctx context.Context, opts QueryOptions) (*ContactList, error) {
	var raw map[string]interface{}

	// If a cursor (full next URL) is provided, use it directly.
	if opts.Cursor != "" && strings.HasPrefix(opts.Cursor, "http") {
		if err := k.doRequestURL(ctx, "GET", opts.Cursor, nil, &raw); err != nil {
			return nil, err
		}
	} else {
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

		if err := k.doRequest(ctx, "GET", endpoint, nil, &raw); err != nil {
			return nil, err
		}
	}

	contactsRaw := jsonArr(raw, "contacts")
	contacts := make([]NormalizedContact, 0, len(contactsRaw))
	for _, c := range contactsRaw {
		cm := asMap(c)
		if cm == nil {
			continue
		}
		contacts = append(contacts, keapContactToNormalized(cm))
	}

	nextCursor := jsonStr(raw, "next")
	return &ContactList{
		Contacts:   contacts,
		Total:      int(jsonFloat(raw, "count")),
		NextCursor: nextCursor,
		HasMore:    nextCursor != "",
	}, nil
}

func (k *KeapConnector) GetContact(ctx context.Context, contactID string) (*NormalizedContact, error) {
	var raw map[string]interface{}
	if err := k.doRequest(ctx, "GET", "/contacts/"+contactID, nil, &raw); err != nil {
		return nil, err
	}

	contact := keapContactToNormalized(raw)
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

	var raw map[string]interface{}
	if err := k.doRequest(ctx, "POST", "/contacts", body, &raw); err != nil {
		return nil, err
	}

	contact := keapContactToNormalized(raw)
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

	var raw map[string]interface{}
	if err := k.doRequest(ctx, "PATCH", "/contacts/"+contactID, body, &raw); err != nil {
		return nil, err
	}

	contact := keapContactToNormalized(raw)
	return &contact, nil
}

func (k *KeapConnector) DeleteContact(ctx context.Context, contactID string) error {
	return k.doRequest(ctx, "DELETE", "/contacts/"+contactID, nil, nil)
}

// ========== TAGS ==========

func (k *KeapConnector) GetTags(ctx context.Context) ([]Tag, error) {
	var raw map[string]interface{}
	if err := k.doRequest(ctx, "GET", "/tags?limit=1000", nil, &raw); err != nil {
		return nil, err
	}

	tagsRaw := jsonArr(raw, "tags")
	tags := make([]Tag, 0, len(tagsRaw))
	for _, t := range tagsRaw {
		tm := asMap(t)
		if tm == nil {
			continue
		}
		tag := Tag{
			ID:          jsonStr(tm, "id"),
			Name:        jsonStr(tm, "name"),
			Description: jsonStr(tm, "description"),
		}
		if cat := jsonObj(tm, "category"); cat != nil {
			tag.Category = jsonStr(cat, "name")
		}
		tags = append(tags, tag)
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
	var raw map[string]interface{}
	if err := k.doRequest(ctx, "GET", "/contacts/model", nil, &raw); err != nil {
		return nil, err
	}

	fieldsRaw := jsonArr(raw, "custom_fields")
	fields := make([]CustomField, 0, len(fieldsRaw))
	for _, f := range fieldsRaw {
		fm := asMap(f)
		if fm == nil {
			continue
		}
		fields = append(fields, CustomField{
			ID:        jsonStr(fm, "id"),
			Key:       jsonStr(fm, "field_name"),
			Label:     jsonStr(fm, "label"),
			FieldType: jsonStr(fm, "field_type"),
		})
	}
	return fields, nil
}

func (k *KeapConnector) GetContactFieldValue(ctx context.Context, contactID string, fieldKey string) (interface{}, error) {
	contact, err := k.GetContact(ctx, contactID)
	if err != nil {
		return nil, err
	}

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
		updates.CustomFields = map[string]interface{}{
			fieldKey: value,
		}
	}

	_, err := k.UpdateContact(ctx, contactID, updates)
	return err
}

// ========== AUTOMATIONS ==========

func (k *KeapConnector) TriggerAutomation(ctx context.Context, contactID string, automationID string) error {
	body := map[string]interface{}{
		"contact_id": contactID,
	}
	return k.doRequest(ctx, "POST", "/campaigns/"+automationID+"/sequences/1/contacts", body, nil)
}

func (k *KeapConnector) AchieveGoal(ctx context.Context, contactID string, goalName string, integration string) error {
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
	status := "UnMarketable"
	if optIn {
		status = "Single Opt In"
	}

	updates := map[string]interface{}{
		"email_status": status,
	}

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

// ========== Contact normalization ==========

// keapContactToNormalized builds a NormalizedContact from a raw Keap JSON map.
// This is type-safe against API inconsistencies (int vs string fields, etc.).
func keapContactToNormalized(m map[string]interface{}) NormalizedContact {
	id := jsonStr(m, "id")
	contact := NormalizedContact{
		ID:           id,
		FirstName:    jsonStr(m, "given_name"),
		LastName:     jsonStr(m, "family_name"),
		Company:      jsonStr(m, "company_name"),
		JobTitle:     jsonStr(m, "job_title"),
		SourceCRM:    keapSlug,
		SourceID:     id,
		CustomFields: make(map[string]interface{}),
	}

	// Email: first entry in email_addresses array
	if emails := jsonArr(m, "email_addresses"); len(emails) > 0 {
		if em := asMap(emails[0]); em != nil {
			contact.Email = jsonStr(em, "email")
		}
	}

	// Phone: first entry in phone_numbers array
	if phones := jsonArr(m, "phone_numbers"); len(phones) > 0 {
		if pm := asMap(phones[0]); pm != nil {
			contact.Phone = jsonStr(pm, "number")
		}
	}

	// Tags
	if tags := jsonArr(m, "tag_ids"); len(tags) > 0 {
		for _, t := range tags {
			tm := asMap(t)
			if tm == nil {
				continue
			}
			contact.Tags = append(contact.Tags, TagRef{
				ID:   jsonStr(tm, "id"),
				Name: jsonStr(tm, "name"),
			})
		}
	}

	// Custom fields
	if cfs := jsonArr(m, "custom_fields"); len(cfs) > 0 {
		for _, cf := range cfs {
			cfm := asMap(cf)
			if cfm == nil {
				continue
			}
			contact.CustomFields[jsonStr(cfm, "id")] = cfm["content"]
		}
	}

	// Timestamps
	if dateCreated := jsonStr(m, "date_created"); dateCreated != "" {
		if t, err := time.Parse(time.RFC3339, dateCreated); err == nil {
			contact.CreatedAt = &t
		}
	}
	if lastUpdated := jsonStr(m, "last_updated"); lastUpdated != "" {
		if t, err := time.Parse(time.RFC3339, lastUpdated); err == nil {
			contact.UpdatedAt = &t
		}
	}

	return contact
}

// jsonFloat extracts a float64 from a map (handles int, float64, string).
func jsonFloat(m map[string]interface{}, key string) float64 {
	v, ok := m[key]
	if !ok || v == nil {
		return 0
	}
	switch val := v.(type) {
	case float64:
		return val
	case int:
		return float64(val)
	case string:
		var f float64
		fmt.Sscanf(val, "%f", &f)
		return f
	case json.Number:
		f, _ := val.Float64()
		return f
	default:
		return 0
	}
}

// ========== Raw Data Provider (for data sync pipeline) ==========

// GetRawPage implements RawDataProvider. Returns raw API records as maps
// to capture ALL fields for dynamic parquet writing.
func (k *KeapConnector) GetRawPage(ctx context.Context, objectType string, opts QueryOptions) (*RawPageResult, error) {
	switch objectType {
	case "contacts":
		return k.getRawContacts(ctx, opts)
	case "tags":
		return k.getRawTags(ctx, opts)
	case "custom_fields":
		return k.getRawCustomFields(ctx, opts)
	default:
		return nil, fmt.Errorf("unsupported object type: %s", objectType)
	}
}

func (k *KeapConnector) getRawContacts(ctx context.Context, opts QueryOptions) (*RawPageResult, error) {
	var raw map[string]interface{}

	if opts.Cursor != "" && strings.HasPrefix(opts.Cursor, "http") {
		if err := k.doRequestURL(ctx, "GET", opts.Cursor, nil, &raw); err != nil {
			return nil, err
		}
	} else {
		params := url.Values{}
		if opts.Limit > 0 {
			params.Set("limit", fmt.Sprintf("%d", opts.Limit))
		} else {
			params.Set("limit", "25")
		}
		if opts.Offset > 0 {
			params.Set("offset", fmt.Sprintf("%d", opts.Offset))
		}

		endpoint := "/contacts"
		if len(params) > 0 {
			endpoint += "?" + params.Encode()
		}

		if err := k.doRequest(ctx, "GET", endpoint, nil, &raw); err != nil {
			return nil, err
		}
	}

	contactsRaw := jsonArr(raw, "contacts")
	records := make([]map[string]interface{}, 0, len(contactsRaw))
	for _, c := range contactsRaw {
		if cm := asMap(c); cm != nil {
			records = append(records, cm)
		}
	}

	nextCursor := jsonStr(raw, "next")
	return &RawPageResult{
		Records:    records,
		NextCursor: nextCursor,
		HasMore:    nextCursor != "",
		Total:      int(jsonFloat(raw, "count")),
	}, nil
}

func (k *KeapConnector) getRawTags(ctx context.Context, _ QueryOptions) (*RawPageResult, error) {
	var raw map[string]interface{}
	if err := k.doRequest(ctx, "GET", "/tags?limit=1000", nil, &raw); err != nil {
		return nil, err
	}

	tagsRaw := jsonArr(raw, "tags")
	records := make([]map[string]interface{}, 0, len(tagsRaw))
	for _, t := range tagsRaw {
		if tm := asMap(t); tm != nil {
			records = append(records, tm)
		}
	}

	return &RawPageResult{
		Records: records,
		HasMore: false,
		Total:   len(records),
	}, nil
}

func (k *KeapConnector) getRawCustomFields(ctx context.Context, _ QueryOptions) (*RawPageResult, error) {
	var raw map[string]interface{}
	if err := k.doRequest(ctx, "GET", "/contacts/model", nil, &raw); err != nil {
		return nil, err
	}

	fieldsRaw := jsonArr(raw, "custom_fields")
	records := make([]map[string]interface{}, 0, len(fieldsRaw))
	for _, f := range fieldsRaw {
		if fm := asMap(f); fm != nil {
			records = append(records, fm)
		}
	}

	return &RawPageResult{
		Records: records,
		HasMore: false,
		Total:   len(records),
	}, nil
}

// ========== HTTP helpers ==========

// doRequest makes an HTTP request to a relative path under the Keap base URL.
func (k *KeapConnector) doRequest(ctx context.Context, method, path string, body interface{}, result interface{}) error {
	return k.doRequestURL(ctx, method, k.baseURL+path, body, result)
}

// doRequestURL makes an HTTP request to an absolute URL.
func (k *KeapConnector) doRequestURL(ctx context.Context, method, fullURL string, body interface{}, result interface{}) error {
	var bodyReader io.Reader
	if body != nil {
		bodyJSON, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = strings.NewReader(string(bodyJSON))
	}

	req, err := http.NewRequestWithContext(ctx, method, fullURL, bodyReader)
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
