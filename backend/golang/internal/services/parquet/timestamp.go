package parquet

import (
	"regexp"
	"strconv"
	"strings"
	"time"
)

// TimestampParser provides date detection and parsing for ETL.
// All parsed dates are converted to Unix epoch milliseconds (UTC).
// Ported from listbackup-ai's timestamp_parser.go.
type TimestampParser struct {
	dateFieldPatterns []*regexp.Regexp
}

// NewTimestampParser creates a new timestamp parser with compiled patterns.
func NewTimestampParser() *TimestampParser {
	patterns := []string{
		`(?i)_at$`,        // created_at, updated_at, deleted_at
		`(?i)_date$`,      // birth_date, start_date
		`(?i)^date_`,      // date_created, date_modified
		`(?i)_time$`,      // start_time, end_time
		`(?i)_timestamp$`, // last_login_timestamp
		`(?i)birthday`,
		`(?i)expir`,     // expires, expiration
		`(?i)_on$`,      // subscribed_on
		`(?i)^last_`,    // last_login, last_activity
		`(?i)due`,       // due_date
		`(?i)scheduled`, // scheduled_at
		`(?i)created`,   // created, date_created
		`(?i)updated`,   // updated, last_updated
		`(?i)modified`,  // modified, date_modified
		`(?i)published`,
		`(?i)deleted`,
	}

	compiled := make([]*regexp.Regexp, 0, len(patterns))
	for _, p := range patterns {
		if re, err := regexp.Compile(p); err == nil {
			compiled = append(compiled, re)
		}
	}

	return &TimestampParser{dateFieldPatterns: compiled}
}

// dateFormats lists supported date formats in priority order.
var dateFormats = []string{
	time.RFC3339,
	time.RFC3339Nano,
	"2006-01-02T15:04:05Z",
	"2006-01-02T15:04:05",
	"2006-01-02 15:04:05",
	"2006-01-02 15:04",
	"2006-01-02",
	"01/02/2006",
	"20060102T150405Z",
	"20060102",
}

// IsLikelyDateField checks if a field name suggests it contains date/time data.
func (p *TimestampParser) IsLikelyDateField(fieldName string) bool {
	for _, pattern := range p.dateFieldPatterns {
		if pattern.MatchString(fieldName) {
			return true
		}
	}
	return false
}

// TryParseTimestamp attempts to parse a value as a timestamp.
// Returns epoch milliseconds (UTC) and success boolean.
func (p *TimestampParser) TryParseTimestamp(value interface{}) (int64, bool) {
	if value == nil {
		return 0, false
	}

	switch v := value.(type) {
	case string:
		return p.parseStringTimestamp(v)
	case float64:
		return p.parseNumericTimestamp(v)
	case int64:
		return p.parseNumericTimestamp(float64(v))
	case int:
		return p.parseNumericTimestamp(float64(v))
	case time.Time:
		if v.IsZero() {
			return 0, false
		}
		return v.UnixMilli(), true
	default:
		return 0, false
	}
}

func (p *TimestampParser) parseStringTimestamp(s string) (int64, bool) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, false
	}

	// Check if it's a numeric string (epoch)
	if i, err := strconv.ParseInt(s, 10, 64); err == nil {
		return p.parseNumericTimestamp(float64(i))
	}
	if f, err := strconv.ParseFloat(s, 64); err == nil {
		return p.parseNumericTimestamp(f)
	}

	for _, format := range dateFormats {
		if t, err := time.Parse(format, s); err == nil {
			return t.UTC().UnixMilli(), true
		}
	}
	return 0, false
}

func (p *TimestampParser) parseNumericTimestamp(v float64) (int64, bool) {
	if v <= 0 {
		return 0, false
	}

	const (
		minReasonableSeconds = 946684800  // Jan 1, 2000
		maxReasonableSeconds = 4102444800 // Jan 1, 2100
		millisThreshold      = 1e12
	)

	var epochMillis int64
	if v > millisThreshold {
		epochMillis = int64(v)
	} else if v >= minReasonableSeconds && v <= maxReasonableSeconds {
		epochMillis = int64(v * 1000)
	} else {
		return 0, false
	}

	if epochMillis < minReasonableSeconds*1000 || epochMillis > maxReasonableSeconds*1000 {
		return 0, false
	}
	return epochMillis, true
}

// DetectAndParseTimestamp combines field name heuristics with value parsing.
func (p *TimestampParser) DetectAndParseTimestamp(fieldName string, value interface{}) (int64, bool) {
	if value == nil {
		return 0, false
	}
	if p.IsLikelyDateField(fieldName) {
		return p.TryParseTimestamp(value)
	}
	// For non-date-named fields, only parse if it's clearly a recognizable format
	if s, ok := value.(string); ok {
		s = strings.TrimSpace(s)
		if len(s) >= 6 && len(s) <= 50 && strings.Contains(s, "T") && strings.Contains(s, "-") {
			return p.parseStringTimestamp(s)
		}
	}
	return 0, false
}

var defaultTimestampParser = NewTimestampParser()

// GetTimestampParser returns the default timestamp parser instance.
func GetTimestampParser() *TimestampParser {
	return defaultTimestampParser
}
