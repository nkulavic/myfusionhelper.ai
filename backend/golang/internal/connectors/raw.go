package connectors

import "context"

// RawDataProvider is an optional interface that connectors can implement
// to provide raw API responses for the data sync pipeline. Unlike the typed
// CRMConnector methods (which normalize to NormalizedContact, Tag, etc.),
// this returns the raw JSON maps from the CRM API. This captures ALL fields
// for writing to parquet via the dynamic schema writer.
type RawDataProvider interface {
	// GetRawPage fetches a page of records for the given object type as raw JSON maps.
	// objectType is one of: "contacts", "tags", "custom_fields"
	GetRawPage(ctx context.Context, objectType string, opts QueryOptions) (*RawPageResult, error)
}

// RawPageResult holds a page of raw API records with pagination info.
type RawPageResult struct {
	Records    []map[string]interface{} `json:"records"`
	NextCursor string                   `json:"next_cursor,omitempty"`
	HasMore    bool                     `json:"has_more"`
	Total      int                      `json:"total"`
}
