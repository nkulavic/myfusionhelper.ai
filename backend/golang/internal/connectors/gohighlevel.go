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
	ghlBaseURL    = "https://services.leadconnectorhq.com"
	ghlSlug       = "gohighlevel"
	ghlAPIVersion = "2021-07-28"
)

func init() {
	Register(ghlSlug, NewGoHighLevelConnector)
}

// GoHighLevelConnector implements CRMConnector for GoHighLevel
type GoHighLevelConnector struct {
	accessToken string
	baseURL     string
	locationID  string
	client      *http.Client
}

// NewGoHighLevelConnector creates a new GoHighLevel CRM connector
func NewGoHighLevelConnector(config ConnectorConfig) (CRMConnector, error) {
	if config.AccessToken == "" {
		return nil, fmt.Errorf("access token is required for GoHighLevel connector")
	}

	baseURL := config.BaseURL
	if baseURL == "" {
		baseURL = ghlBaseURL
	}

	return &GoHighLevelConnector{
		accessToken: config.AccessToken,
		baseURL:     baseURL,
		locationID:  config.AccountID, // GHL uses locationId
		client:      &http.Client{Timeout: 30 * time.Second},
	}, nil
}

// ========== CONTACTS ==========

func (g *GoHighLevelConnector) GetContacts(ctx context.Context, opts QueryOptions) (*ContactList, error) {
	params := url.Values{}
	if opts.Limit > 0 {
		params.Set("limit", fmt.Sprintf("%d", opts.Limit))
	} else {
		params.Set("limit", "20")
	}
	if opts.Cursor != "" {
		params.Set("startAfterId", opts.Cursor)
	}
	if opts.Email != "" {
		params.Set("email", opts.Email)
	}
	if g.locationID != "" {
		params.Set("locationId", g.locationID)
	}

	endpoint := "/contacts/?" + params.Encode()

	var result struct {
		Contacts []ghlContact `json:"contacts"`
		Meta     struct {
			Total        int    `json:"total"`
			NextPageUrl  string `json:"nextPageUrl"`
			StartAfterId string `json:"startAfterId"`
		} `json:"meta"`
	}

	if err := g.doRequest(ctx, "GET", endpoint, nil, &result); err != nil {
		return nil, err
	}

	contacts := make([]NormalizedContact, 0, len(result.Contacts))
	for _, gc := range result.Contacts {
		contacts = append(contacts, gc.toNormalized())
	}

	return &ContactList{
		Contacts:   contacts,
		Total:      result.Meta.Total,
		NextCursor: result.Meta.StartAfterId,
		HasMore:    result.Meta.NextPageUrl != "",
	}, nil
}

func (g *GoHighLevelConnector) GetContact(ctx context.Context, contactID string) (*NormalizedContact, error) {
	var result struct {
		Contact ghlContact `json:"contact"`
	}
	if err := g.doRequest(ctx, "GET", "/contacts/"+contactID, nil, &result); err != nil {
		return nil, err
	}

	contact := result.Contact.toNormalized()
	return &contact, nil
}

func (g *GoHighLevelConnector) CreateContact(ctx context.Context, input CreateContactInput) (*NormalizedContact, error) {
	body := map[string]interface{}{
		"firstName": input.FirstName,
		"lastName":  input.LastName,
		"email":     input.Email,
	}
	if input.Phone != "" {
		body["phone"] = input.Phone
	}
	if input.Company != "" {
		body["companyName"] = input.Company
	}
	if g.locationID != "" {
		body["locationId"] = g.locationID
	}
	if input.Tags != nil {
		body["tags"] = input.Tags
	}
	if input.CustomFields != nil {
		customFields := make([]map[string]interface{}, 0)
		for key, value := range input.CustomFields {
			customFields = append(customFields, map[string]interface{}{
				"id":          key,
				"field_value": value,
			})
		}
		body["customFields"] = customFields
	}

	var result struct {
		Contact ghlContact `json:"contact"`
	}
	if err := g.doRequest(ctx, "POST", "/contacts/", body, &result); err != nil {
		return nil, err
	}

	contact := result.Contact.toNormalized()
	return &contact, nil
}

func (g *GoHighLevelConnector) UpdateContact(ctx context.Context, contactID string, updates UpdateContactInput) (*NormalizedContact, error) {
	body := map[string]interface{}{}

	if updates.FirstName != nil {
		body["firstName"] = *updates.FirstName
	}
	if updates.LastName != nil {
		body["lastName"] = *updates.LastName
	}
	if updates.Email != nil {
		body["email"] = *updates.Email
	}
	if updates.Phone != nil {
		body["phone"] = *updates.Phone
	}
	if updates.Company != nil {
		body["companyName"] = *updates.Company
	}
	if updates.CustomFields != nil {
		customFields := make([]map[string]interface{}, 0)
		for key, value := range updates.CustomFields {
			customFields = append(customFields, map[string]interface{}{
				"id":          key,
				"field_value": value,
			})
		}
		body["customFields"] = customFields
	}

	var result struct {
		Contact ghlContact `json:"contact"`
	}
	if err := g.doRequest(ctx, "PUT", "/contacts/"+contactID, body, &result); err != nil {
		return nil, err
	}

	contact := result.Contact.toNormalized()
	return &contact, nil
}

func (g *GoHighLevelConnector) DeleteContact(ctx context.Context, contactID string) error {
	return g.doRequest(ctx, "DELETE", "/contacts/"+contactID, nil, nil)
}

// ========== TAGS ==========

func (g *GoHighLevelConnector) GetTags(ctx context.Context) ([]Tag, error) {
	params := url.Values{}
	if g.locationID != "" {
		params.Set("locationId", g.locationID)
	}

	var result struct {
		Tags []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"tags"`
	}

	if err := g.doRequest(ctx, "GET", "/tags/?"+params.Encode(), nil, &result); err != nil {
		return nil, err
	}

	tags := make([]Tag, 0, len(result.Tags))
	for _, t := range result.Tags {
		tags = append(tags, Tag{
			ID:   t.ID,
			Name: t.Name,
		})
	}
	return tags, nil
}

func (g *GoHighLevelConnector) ApplyTag(ctx context.Context, contactID string, tagID string) error {
	body := map[string]interface{}{
		"tags": []string{tagID},
	}
	return g.doRequest(ctx, "POST", "/contacts/"+contactID+"/tags", body, nil)
}

func (g *GoHighLevelConnector) RemoveTag(ctx context.Context, contactID string, tagID string) error {
	return g.doRequest(ctx, "DELETE", "/contacts/"+contactID+"/tags", map[string]interface{}{
		"tags": []string{tagID},
	}, nil)
}

// ========== CUSTOM FIELDS ==========

func (g *GoHighLevelConnector) GetCustomFields(ctx context.Context) ([]CustomField, error) {
	params := url.Values{}
	if g.locationID != "" {
		params.Set("locationId", g.locationID)
	}

	var result struct {
		CustomFields []struct {
			ID        string   `json:"id"`
			Name      string   `json:"name"`
			FieldKey  string   `json:"fieldKey"`
			DataType  string   `json:"dataType"`
			Options   []string `json:"picklistOptions,omitempty"`
		} `json:"customFields"`
	}

	if err := g.doRequest(ctx, "GET", "/locations/"+g.locationID+"/customFields?"+params.Encode(), nil, &result); err != nil {
		return nil, err
	}

	fields := make([]CustomField, 0, len(result.CustomFields))
	for _, f := range result.CustomFields {
		fields = append(fields, CustomField{
			ID:        f.ID,
			Key:       f.FieldKey,
			Label:     f.Name,
			FieldType: f.DataType,
			Options:   f.Options,
		})
	}
	return fields, nil
}

func (g *GoHighLevelConnector) GetContactFieldValue(ctx context.Context, contactID string, fieldKey string) (interface{}, error) {
	contact, err := g.GetContact(ctx, contactID)
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
	case "company", "companyName":
		return contact.Company, nil
	}

	if val, ok := contact.CustomFields[fieldKey]; ok {
		return val, nil
	}
	return nil, nil
}

func (g *GoHighLevelConnector) SetContactFieldValue(ctx context.Context, contactID string, fieldKey string, value interface{}) error {
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
	case "company", "companyName":
		v := fmt.Sprintf("%v", value)
		updates.Company = &v
	default:
		updates.CustomFields = map[string]interface{}{
			fieldKey: value,
		}
	}

	_, err := g.UpdateContact(ctx, contactID, updates)
	return err
}

// ========== AUTOMATIONS ==========

func (g *GoHighLevelConnector) TriggerAutomation(ctx context.Context, contactID string, automationID string) error {
	// GHL uses workflows - add contact to a workflow
	body := map[string]interface{}{
		"contactId": contactID,
	}
	return g.doRequest(ctx, "POST", "/workflows/"+automationID+"/contacts", body, nil)
}

func (g *GoHighLevelConnector) AchieveGoal(_ context.Context, _ string, _ string, _ string) error {
	return NewConnectorError(ghlSlug, 501, "GoHighLevel does not support goal achievement", false)
}

// ========== HEALTH ==========

func (g *GoHighLevelConnector) TestConnection(ctx context.Context) error {
	params := url.Values{}
	params.Set("limit", "1")
	if g.locationID != "" {
		params.Set("locationId", g.locationID)
	}
	var result map[string]interface{}
	return g.doRequest(ctx, "GET", "/contacts/?"+params.Encode(), nil, &result)
}

func (g *GoHighLevelConnector) GetMetadata() ConnectorMetadata {
	return ConnectorMetadata{
		PlatformSlug: ghlSlug,
		PlatformName: "GoHighLevel",
		APIVersion:   "v2",
		BaseURL:      g.baseURL,
	}
}

func (g *GoHighLevelConnector) GetCapabilities() []Capability {
	return []Capability{
		CapContacts,
		CapTags,
		CapCustomFields,
		CapAutomations,
		CapWebhooks,
	}
}

// ========== INTERNAL TYPES ==========

type ghlContact struct {
	ID          string   `json:"id"`
	FirstName   string   `json:"firstName"`
	LastName    string   `json:"lastName"`
	Email       string   `json:"email"`
	Phone       string   `json:"phone"`
	CompanyName string   `json:"companyName"`
	Tags        []string `json:"tags"`
	CustomFields []struct {
		ID    string      `json:"id"`
		Value interface{} `json:"value"`
	} `json:"customFields"`
	DateAdded    string `json:"dateAdded"`
	DateUpdated  string `json:"dateUpdated"`
	LocationID   string `json:"locationId"`
}

func (gc *ghlContact) toNormalized() NormalizedContact {
	contact := NormalizedContact{
		ID:           gc.ID,
		FirstName:    gc.FirstName,
		LastName:     gc.LastName,
		Email:        gc.Email,
		Phone:        gc.Phone,
		Company:      gc.CompanyName,
		SourceCRM:    ghlSlug,
		SourceID:     gc.ID,
		CustomFields: make(map[string]interface{}),
	}

	for _, tag := range gc.Tags {
		contact.Tags = append(contact.Tags, TagRef{
			ID:   tag,
			Name: tag,
		})
	}

	for _, cf := range gc.CustomFields {
		contact.CustomFields[cf.ID] = cf.Value
	}

	if gc.DateAdded != "" {
		if t, err := time.Parse(time.RFC3339, gc.DateAdded); err == nil {
			contact.CreatedAt = &t
		}
	}
	if gc.DateUpdated != "" {
		if t, err := time.Parse(time.RFC3339, gc.DateUpdated); err == nil {
			contact.UpdatedAt = &t
		}
	}

	return contact
}

// ========== HTTP HELPER ==========

func (g *GoHighLevelConnector) doRequest(ctx context.Context, method, path string, body interface{}, result interface{}) error {
	var bodyReader io.Reader
	if body != nil {
		bodyJSON, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = strings.NewReader(string(bodyJSON))
	}

	apiURL := g.baseURL + path
	req, err := http.NewRequestWithContext(ctx, method, apiURL, bodyReader)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+g.accessToken)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Version", ghlAPIVersion)

	resp, err := g.client.Do(req)
	if err != nil {
		return NewConnectorError(ghlSlug, 0, fmt.Sprintf("request failed: %v", err), true)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return NewConnectorError(ghlSlug, resp.StatusCode, "failed to read response", true)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		retryable := resp.StatusCode == 429 || resp.StatusCode >= 500
		return NewConnectorError(ghlSlug, resp.StatusCode,
			fmt.Sprintf("GoHighLevel API error (%d): %s", resp.StatusCode, string(respBody)), retryable)
	}

	if result != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("failed to parse response: %w", err)
		}
	}

	return nil
}
