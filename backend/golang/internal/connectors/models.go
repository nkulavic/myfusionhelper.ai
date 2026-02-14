package connectors

import (
	"time"
)

// NormalizedContact represents a contact normalized across CRM platforms
type NormalizedContact struct {
	ID           string                 `json:"id"`
	FirstName    string                 `json:"first_name"`
	LastName     string                 `json:"last_name"`
	Email        string                 `json:"email"`
	Phone        string                 `json:"phone,omitempty"`
	Company      string                 `json:"company,omitempty"`
	JobTitle     string                 `json:"job_title,omitempty"`
	Tags         []TagRef               `json:"tags,omitempty"`
	CustomFields map[string]interface{} `json:"custom_fields,omitempty"`
	SourceCRM    string                 `json:"source_crm"`
	SourceID     string                 `json:"source_id"`
	CreatedAt    *time.Time             `json:"created_at,omitempty"`
	UpdatedAt    *time.Time             `json:"updated_at,omitempty"`
}

// TagRef represents a tag reference on a contact
type TagRef struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// Tag represents a tag/label in the CRM system
type Tag struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Category    string `json:"category,omitempty"`
}

// CustomField represents a custom field definition in the CRM
type CustomField struct {
	ID           string   `json:"id"`
	Key          string   `json:"key"`
	Label        string   `json:"label"`
	FieldType    string   `json:"field_type"`
	GroupName    string   `json:"group_name,omitempty"`
	Options      []string `json:"options,omitempty"`
	DefaultValue string   `json:"default_value,omitempty"`
}

// ConnectorError represents an error from a CRM connector
type ConnectorError struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	StatusCode int    `json:"status_code"`
	Platform   string `json:"platform"`
	Retryable  bool   `json:"retryable"`
}

func (e *ConnectorError) Error() string {
	return e.Message
}

// NewConnectorError creates a new connector error
func NewConnectorError(platform string, statusCode int, message string, retryable bool) *ConnectorError {
	return &ConnectorError{
		Code:       "CONNECTOR_ERROR",
		Message:    message,
		StatusCode: statusCode,
		Platform:   platform,
		Retryable:  retryable,
	}
}
