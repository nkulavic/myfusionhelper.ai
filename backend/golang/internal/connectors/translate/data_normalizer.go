package translate

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// DataNormalizer handles bidirectional data format conversion between
// CRM-specific representations and a standardized internal format.
type DataNormalizer struct {
	platformSlug string
}

// NewDataNormalizer creates a normalizer for the given platform.
func NewDataNormalizer(platformSlug string) *DataNormalizer {
	return &DataNormalizer{platformSlug: platformSlug}
}

// NormalizeRead converts a CRM-specific value to a standardized format.
// fieldType is the CRM's declared field type (e.g., "Date", "Phone", "DateTime").
func (n *DataNormalizer) NormalizeRead(value interface{}, fieldType string) interface{} {
	if value == nil {
		return nil
	}

	switch strings.ToLower(fieldType) {
	case "date", "datetime", "timestamp", "datefield":
		return n.normalizeDateRead(value)
	default:
		return value
	}
}

// NormalizeWrite converts a standardized value to the CRM-specific format.
func (n *DataNormalizer) NormalizeWrite(value interface{}, fieldType string) interface{} {
	if value == nil {
		return nil
	}

	switch strings.ToLower(fieldType) {
	case "date", "datetime", "timestamp", "datefield":
		return n.normalizeDateWrite(value)
	default:
		return value
	}
}

// normalizeDateRead parses CRM-specific date formats into RFC3339.
func (n *DataNormalizer) normalizeDateRead(value interface{}) interface{} {
	// Handle numeric timestamps (Ontraport)
	if ts, ok := toInt64(value); ok && ts > 0 {
		return time.Unix(ts, 0).UTC().Format(time.RFC3339)
	}

	strVal, ok := value.(string)
	if !ok {
		return value
	}

	if strVal == "" {
		return value
	}

	// Ontraport may return Unix timestamps as strings
	if n.platformSlug == "ontraport" {
		if ts, err := strconv.ParseInt(strVal, 10, 64); err == nil && ts > 0 {
			return time.Unix(ts, 0).UTC().Format(time.RFC3339)
		}
	}

	// Try platform-specific date formats
	formats := n.readDateFormats()
	for _, f := range formats {
		if t, err := time.Parse(f, strVal); err == nil {
			return t.Format(time.RFC3339)
		}
	}

	// Return as-is if no format matched
	return value
}

// readDateFormats returns the expected date formats for the platform, ordered by likelihood.
func (n *DataNormalizer) readDateFormats() []string {
	switch n.platformSlug {
	case "keap":
		return []string{
			time.RFC3339,
			"2006-01-02T15:04:05.000Z",
			"2006-01-02T15:04:05Z",
			"2006-01-02",
		}
	case "gohighlevel":
		return []string{
			time.RFC3339,
			"2006-01-02T15:04:05.000Z",
			"2006-01-02",
		}
	case "activecampaign":
		return []string{
			"2006-01-02T15:04:05-07:00",
			time.RFC3339,
			"2006-01-02 15:04:05",
			"2006-01-02",
		}
	case "ontraport":
		return []string{
			time.RFC3339,
			"2006-01-02",
		}
	default:
		return []string{
			time.RFC3339,
			"2006-01-02T15:04:05.000Z",
			"2006-01-02 15:04:05",
			"2006-01-02",
		}
	}
}

// normalizeDateWrite converts a standardized date to the CRM-specific format.
func (n *DataNormalizer) normalizeDateWrite(value interface{}) interface{} {
	strVal, ok := value.(string)
	if !ok {
		return value
	}

	if strVal == "" {
		return value
	}

	// Try parsing as RFC3339 first (our standard format)
	t, err := time.Parse(time.RFC3339, strVal)
	if err != nil {
		// Try other common formats the helper might produce
		for _, f := range []string{"2006-01-02", "01/02/2006", "2006-01-02 15:04:05"} {
			if parsed, parseErr := time.Parse(f, strVal); parseErr == nil {
				t = parsed
				break
			}
		}
		if t.IsZero() {
			return value // Cannot parse, pass through
		}
	}

	// Convert to CRM-specific format
	switch n.platformSlug {
	case "keap":
		return t.Format("2006-01-02T15:04:05.000Z")
	case "gohighlevel":
		return t.Format("2006-01-02T15:04:05.000Z")
	case "activecampaign":
		return t.Format("2006-01-02 15:04:05")
	case "ontraport":
		return fmt.Sprintf("%d", t.Unix())
	default:
		return t.Format(time.RFC3339)
	}
}

// toInt64 attempts to convert a value to int64.
func toInt64(value interface{}) (int64, bool) {
	switch v := value.(type) {
	case int:
		return int64(v), true
	case int32:
		return int64(v), true
	case int64:
		return v, true
	case float64:
		return int64(v), true
	case float32:
		return int64(v), true
	default:
		return 0, false
	}
}
