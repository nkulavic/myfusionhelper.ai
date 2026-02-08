package helpers

import (
	"context"

	"github.com/myfusionhelper/api/internal/connectors"
)

// Helper defines the interface for all automation helpers.
// Each helper type (tag_it, copy_it, format_it, etc.) implements this interface.
type Helper interface {
	// Metadata
	GetName() string
	GetType() string
	GetCategory() string
	GetDescription() string
	GetConfigSchema() map[string]interface{}

	// Execution
	Execute(ctx context.Context, input HelperInput) (*HelperOutput, error)

	// Validation
	ValidateConfig(config map[string]interface{}) error

	// Capabilities
	RequiresCRM() bool
	SupportedCRMs() []string // empty = all CRMs
}

// HelperInput provides all context needed to execute a helper
type HelperInput struct {
	ContactID    string                                  `json:"contact_id"`
	ContactData  *connectors.NormalizedContact            `json:"contact_data,omitempty"`
	Config       map[string]interface{}                   `json:"config"`
	Connector    connectors.CRMConnector                  `json:"-"`
	ServiceAuths map[string]*connectors.ConnectorConfig   `json:"-"` // keyed by platform slug (e.g. "zoom", "trello")
	UserID       string                                  `json:"user_id"`
	AccountID    string                                  `json:"account_id"`
	HelperID     string                                  `json:"helper_id"`
}

// HelperOutput represents the result of a helper execution
type HelperOutput struct {
	Success      bool                   `json:"success"`
	Message      string                 `json:"message"`
	ModifiedData map[string]interface{} `json:"modified_data,omitempty"`
	Actions      []HelperAction         `json:"actions,omitempty"`
	Logs         []string               `json:"logs,omitempty"`
}

// HelperAction represents an action taken during helper execution
type HelperAction struct {
	Type   string      `json:"type"`   // "tag_applied", "tag_removed", "field_updated", "automation_triggered"
	Target string      `json:"target"` // Contact ID, field name, tag name, etc.
	Value  interface{} `json:"value,omitempty"`
}
