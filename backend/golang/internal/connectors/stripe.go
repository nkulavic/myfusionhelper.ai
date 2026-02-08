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
	stripeBaseURL = "https://api.stripe.com/v1"
	stripeSlug    = "stripe"
)

func init() {
	Register(stripeSlug, NewStripeConnector)
}

// StripeConnector implements CRMConnector for Stripe.
// Stripe is a payments platform, so we map customers → contacts,
// metadata → custom fields, and use labels/tags where applicable.
type StripeConnector struct {
	apiKey  string
	baseURL string
	client  *http.Client
}

// NewStripeConnector creates a new Stripe connector
func NewStripeConnector(config ConnectorConfig) (CRMConnector, error) {
	key := config.APIKey
	if key == "" {
		key = config.AccessToken
	}
	if key == "" {
		return nil, fmt.Errorf("API key is required for Stripe connector")
	}

	baseURL := config.BaseURL
	if baseURL == "" {
		baseURL = stripeBaseURL
	}

	return &StripeConnector{
		apiKey:  key,
		baseURL: strings.TrimRight(baseURL, "/"),
		client:  &http.Client{Timeout: 30 * time.Second},
	}, nil
}

// ========== CONTACTS (Stripe Customers) ==========

func (s *StripeConnector) GetContacts(ctx context.Context, opts QueryOptions) (*ContactList, error) {
	limit := opts.Limit
	if limit <= 0 {
		limit = 25
	}
	if limit > 100 {
		limit = 100
	}

	params := url.Values{}
	params.Set("limit", fmt.Sprintf("%d", limit))
	if opts.Cursor != "" {
		params.Set("starting_after", opts.Cursor)
	}
	if opts.Email != "" {
		params.Set("email", opts.Email)
	}

	var result stripeCustomerList
	if err := s.doRequest(ctx, "GET", "/customers?"+params.Encode(), nil, &result); err != nil {
		return nil, err
	}

	contacts := make([]NormalizedContact, 0, len(result.Data))
	for _, sc := range result.Data {
		contacts = append(contacts, sc.toNormalized())
	}

	cl := &ContactList{
		Contacts: contacts,
		Total:    len(result.Data),
		HasMore:  result.HasMore,
	}
	if result.HasMore && len(result.Data) > 0 {
		cl.NextCursor = result.Data[len(result.Data)-1].ID
	}
	return cl, nil
}

func (s *StripeConnector) GetContact(ctx context.Context, contactID string) (*NormalizedContact, error) {
	var result stripeCustomer
	if err := s.doRequest(ctx, "GET", "/customers/"+contactID, nil, &result); err != nil {
		return nil, err
	}
	contact := result.toNormalized()
	return &contact, nil
}

func (s *StripeConnector) CreateContact(ctx context.Context, input CreateContactInput) (*NormalizedContact, error) {
	form := url.Values{}
	if input.Email != "" {
		form.Set("email", input.Email)
	}
	name := strings.TrimSpace(input.FirstName + " " + input.LastName)
	if name != "" {
		form.Set("name", name)
	}
	if input.Phone != "" {
		form.Set("phone", input.Phone)
	}
	if input.CustomFields != nil {
		for key, value := range input.CustomFields {
			form.Set(fmt.Sprintf("metadata[%s]", key), fmt.Sprintf("%v", value))
		}
	}

	var result stripeCustomer
	if err := s.doFormRequest(ctx, "POST", "/customers", form, &result); err != nil {
		return nil, err
	}

	contact := result.toNormalized()
	return &contact, nil
}

func (s *StripeConnector) UpdateContact(ctx context.Context, contactID string, updates UpdateContactInput) (*NormalizedContact, error) {
	form := url.Values{}

	// Build name from updates
	var nameParts []string
	if updates.FirstName != nil {
		nameParts = append(nameParts, *updates.FirstName)
	}
	if updates.LastName != nil {
		nameParts = append(nameParts, *updates.LastName)
	}
	if len(nameParts) > 0 {
		form.Set("name", strings.Join(nameParts, " "))
	}

	if updates.Email != nil {
		form.Set("email", *updates.Email)
	}
	if updates.Phone != nil {
		form.Set("phone", *updates.Phone)
	}
	if updates.CustomFields != nil {
		for key, value := range updates.CustomFields {
			form.Set(fmt.Sprintf("metadata[%s]", key), fmt.Sprintf("%v", value))
		}
	}

	path := "/customers/" + contactID
	var result stripeCustomer
	if err := s.doFormRequest(ctx, "POST", path, form, &result); err != nil {
		return nil, err
	}

	contact := result.toNormalized()
	return &contact, nil
}

func (s *StripeConnector) DeleteContact(ctx context.Context, contactID string) error {
	return s.doRequest(ctx, "DELETE", "/customers/"+contactID, nil, nil)
}

// ========== TAGS ==========

// Stripe doesn't have native tags — return empty
func (s *StripeConnector) GetTags(_ context.Context) ([]Tag, error) {
	return []Tag{}, nil
}

func (s *StripeConnector) ApplyTag(ctx context.Context, contactID string, tagID string) error {
	// Store tags in customer metadata
	form := url.Values{}
	form.Set(fmt.Sprintf("metadata[tag_%s]", tagID), tagID)

	return s.doFormRequest(ctx, "POST", "/customers/"+contactID, form, nil)
}

func (s *StripeConnector) RemoveTag(ctx context.Context, contactID string, tagID string) error {
	// Remove tag from customer metadata by setting to empty
	form := url.Values{}
	form.Set(fmt.Sprintf("metadata[tag_%s]", tagID), "")

	return s.doFormRequest(ctx, "POST", "/customers/"+contactID, form, nil)
}

// ========== CUSTOM FIELDS ==========

// Stripe uses metadata as custom fields
func (s *StripeConnector) GetCustomFields(_ context.Context) ([]CustomField, error) {
	// Stripe metadata is schemaless — no predefined custom fields
	return []CustomField{}, nil
}

func (s *StripeConnector) GetContactFieldValue(ctx context.Context, contactID string, fieldKey string) (interface{}, error) {
	contact, err := s.GetContact(ctx, contactID)
	if err != nil {
		return nil, err
	}

	switch fieldKey {
	case "first_name":
		return contact.FirstName, nil
	case "last_name":
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

func (s *StripeConnector) SetContactFieldValue(ctx context.Context, contactID string, fieldKey string, value interface{}) error {
	updates := UpdateContactInput{}

	switch fieldKey {
	case "first_name":
		v := fmt.Sprintf("%v", value)
		updates.FirstName = &v
	case "last_name":
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

	_, err := s.UpdateContact(ctx, contactID, updates)
	return err
}

// ========== AUTOMATIONS ==========

func (s *StripeConnector) TriggerAutomation(_ context.Context, _ string, _ string) error {
	return NewConnectorError(stripeSlug, 501, "Stripe does not support automation triggers", false)
}

func (s *StripeConnector) AchieveGoal(_ context.Context, _ string, _ string, _ string) error {
	return NewConnectorError(stripeSlug, 501, "Stripe does not support goal achievement", false)
}

// ========== HEALTH ==========

func (s *StripeConnector) TestConnection(ctx context.Context) error {
	var result map[string]interface{}
	return s.doRequest(ctx, "GET", "/balance", nil, &result)
}

func (s *StripeConnector) GetMetadata() ConnectorMetadata {
	return ConnectorMetadata{
		PlatformSlug: stripeSlug,
		PlatformName: "Stripe",
		APIVersion:   "v1",
		BaseURL:      s.baseURL,
	}
}

func (s *StripeConnector) GetCapabilities() []Capability {
	return []Capability{
		CapContacts,
		CapCustomFields,
	}
}

// ========== INTERNAL TYPES ==========

type stripeCustomer struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Email       string            `json:"email"`
	Phone       string            `json:"phone"`
	Description string            `json:"description"`
	Metadata    map[string]string `json:"metadata"`
	Created     int64             `json:"created"`
}

type stripeCustomerList struct {
	Data    []stripeCustomer `json:"data"`
	HasMore bool             `json:"has_more"`
}

func (sc *stripeCustomer) toNormalized() NormalizedContact {
	// Split name into first/last
	firstName, lastName := "", ""
	if sc.Name != "" {
		parts := strings.SplitN(sc.Name, " ", 2)
		firstName = parts[0]
		if len(parts) > 1 {
			lastName = parts[1]
		}
	}

	contact := NormalizedContact{
		ID:           sc.ID,
		FirstName:    firstName,
		LastName:     lastName,
		Email:        sc.Email,
		Phone:        sc.Phone,
		SourceCRM:    stripeSlug,
		SourceID:     sc.ID,
		CustomFields: make(map[string]interface{}),
	}

	if sc.Created > 0 {
		t := time.Unix(sc.Created, 0).UTC()
		contact.CreatedAt = &t
	}

	// Map metadata to custom fields
	for key, value := range sc.Metadata {
		contact.CustomFields[key] = value
	}

	return contact
}

// ========== HTTP HELPERS ==========

func (s *StripeConnector) doRequest(ctx context.Context, method, path string, body interface{}, result interface{}) error {
	var bodyReader io.Reader
	if body != nil {
		bodyJSON, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = strings.NewReader(string(bodyJSON))
	}

	apiURL := s.baseURL + path
	req, err := http.NewRequestWithContext(ctx, method, apiURL, bodyReader)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+s.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	return s.executeRequest(req, result)
}

func (s *StripeConnector) doFormRequest(ctx context.Context, method, path string, form url.Values, result interface{}) error {
	apiURL := s.baseURL + path
	req, err := http.NewRequestWithContext(ctx, method, apiURL, strings.NewReader(form.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+s.apiKey)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	return s.executeRequest(req, result)
}

func (s *StripeConnector) executeRequest(req *http.Request, result interface{}) error {
	resp, err := s.client.Do(req)
	if err != nil {
		return NewConnectorError(stripeSlug, 0, fmt.Sprintf("request failed: %v", err), true)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return NewConnectorError(stripeSlug, resp.StatusCode, "failed to read response", true)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		retryable := resp.StatusCode == 429 || resp.StatusCode >= 500
		return NewConnectorError(stripeSlug, resp.StatusCode,
			fmt.Sprintf("Stripe API error (%d): %s", resp.StatusCode, string(respBody)), retryable)
	}

	if result != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("failed to parse response: %w", err)
		}
	}

	return nil
}
