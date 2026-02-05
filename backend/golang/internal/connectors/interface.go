package connectors

import (
	"context"
)

// CRMConnector defines the unified interface for all CRM platform integrations.
// Each CRM platform (Keap, GoHighLevel, ActiveCampaign, etc.) implements this interface
// to normalize operations across different APIs.
type CRMConnector interface {
	// Contacts
	GetContacts(ctx context.Context, opts QueryOptions) (*ContactList, error)
	GetContact(ctx context.Context, contactID string) (*NormalizedContact, error)
	CreateContact(ctx context.Context, contact CreateContactInput) (*NormalizedContact, error)
	UpdateContact(ctx context.Context, contactID string, updates UpdateContactInput) (*NormalizedContact, error)
	DeleteContact(ctx context.Context, contactID string) error

	// Tags
	GetTags(ctx context.Context) ([]Tag, error)
	ApplyTag(ctx context.Context, contactID string, tagID string) error
	RemoveTag(ctx context.Context, contactID string, tagID string) error

	// Custom Fields
	GetCustomFields(ctx context.Context) ([]CustomField, error)
	GetContactFieldValue(ctx context.Context, contactID string, fieldKey string) (interface{}, error)
	SetContactFieldValue(ctx context.Context, contactID string, fieldKey string, value interface{}) error

	// Automations / Sequences
	TriggerAutomation(ctx context.Context, contactID string, automationID string) error
	AchieveGoal(ctx context.Context, contactID string, goalName string, integration string) error

	// Health & Metadata
	TestConnection(ctx context.Context) error
	GetMetadata() ConnectorMetadata
	GetCapabilities() []Capability
}

// QueryOptions provides filtering and pagination for list operations
type QueryOptions struct {
	Limit    int               `json:"limit"`
	Offset   int               `json:"offset"`
	Cursor   string            `json:"cursor,omitempty"`
	OrderBy  string            `json:"order_by,omitempty"`
	Filters  map[string]string `json:"filters,omitempty"`
	TagID    string            `json:"tag_id,omitempty"`
	Email    string            `json:"email,omitempty"`
}

// ContactList represents a paginated list of contacts
type ContactList struct {
	Contacts   []NormalizedContact `json:"contacts"`
	Total      int                 `json:"total"`
	NextCursor string              `json:"next_cursor,omitempty"`
	HasMore    bool                `json:"has_more"`
}

// CreateContactInput represents data for creating a contact
type CreateContactInput struct {
	FirstName    string                 `json:"first_name"`
	LastName     string                 `json:"last_name"`
	Email        string                 `json:"email"`
	Phone        string                 `json:"phone,omitempty"`
	Company      string                 `json:"company,omitempty"`
	CustomFields map[string]interface{} `json:"custom_fields,omitempty"`
	Tags         []string               `json:"tags,omitempty"`
}

// UpdateContactInput represents data for updating a contact
type UpdateContactInput struct {
	FirstName    *string                `json:"first_name,omitempty"`
	LastName     *string                `json:"last_name,omitempty"`
	Email        *string                `json:"email,omitempty"`
	Phone        *string                `json:"phone,omitempty"`
	Company      *string                `json:"company,omitempty"`
	CustomFields map[string]interface{} `json:"custom_fields,omitempty"`
}

// ConnectorMetadata provides information about the CRM connector
type ConnectorMetadata struct {
	PlatformSlug string `json:"platform_slug"`
	PlatformName string `json:"platform_name"`
	APIVersion   string `json:"api_version"`
	BaseURL      string `json:"base_url"`
}

// Capability represents a feature supported by a CRM connector
type Capability string

const (
	CapContacts    Capability = "contacts"
	CapTags        Capability = "tags"
	CapCustomFields Capability = "custom_fields"
	CapAutomations Capability = "automations"
	CapGoals       Capability = "goals"
	CapDeals       Capability = "deals"
	CapEmails      Capability = "emails"
	CapWebhooks    Capability = "webhooks"
)

// ConnectorConfig holds authentication and configuration for a connector instance
type ConnectorConfig struct {
	AccessToken string `json:"access_token"`
	APIKey      string `json:"api_key,omitempty"`
	APISecret   string `json:"api_secret,omitempty"`
	BaseURL     string `json:"base_url,omitempty"`
	AccountID   string `json:"account_id,omitempty"`
}
