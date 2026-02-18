package parquet

import (
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strings"
)

// TransformService flattens nested JSON structures into flat, queryable columns
// for parquet writing. Ported from listbackup-ai's transform_service.go.
//
// Example:
//
//	{"address": {"state": "CA"}, "tags": [1, 2]} →
//	{"address_state": "CA", "tags": "[1,2]"}
type TransformService struct {
	maxDepth            int
	maxArrayIndex       int
	maxColumnNameLength int
}

// NewTransformService creates a new transform service with default settings.
func NewTransformService() *TransformService {
	return &TransformService{
		maxDepth:            0,   // 0 = unlimited
		maxArrayIndex:       -1,  // -1 = unlimited
		maxColumnNameLength: 128, // DuckDB/Parquet compatibility
	}
}

// TransformConfig allows customization of transform behavior.
type TransformConfig struct {
	MaxDepth            int
	MaxArrayIndex       int
	MaxColumnNameLength int
	FieldMappings       map[string]string // Custom field name mappings
	ExcludeFields       []string          // Fields to exclude
	FlattenArrays       bool              // Whether to flatten array elements (default: true)
	PreserveNulls       bool              // Whether to include null fields in output
}

// FlattenedRecord represents a transformed record with flattened fields.
type FlattenedRecord struct {
	Fields          map[string]interface{} // Flattened field name -> value
	TimestampFields map[string]int64       // Field name -> epoch milliseconds
	FieldList       []string               // Ordered list of field names
}

// TransformRecord flattens a nested JSON record into queryable columns.
func (s *TransformService) TransformRecord(record map[string]interface{}, config *TransformConfig) (*FlattenedRecord, error) {
	if config == nil {
		config = &TransformConfig{
			MaxDepth:            s.maxDepth,
			MaxArrayIndex:       s.maxArrayIndex,
			MaxColumnNameLength: s.maxColumnNameLength,
			FlattenArrays:       true,
		}
	}

	if config.MaxColumnNameLength == 0 {
		config.MaxColumnNameLength = s.maxColumnNameLength
	}

	result := &FlattenedRecord{
		Fields:          make(map[string]interface{}),
		TimestampFields: make(map[string]int64),
		FieldList:       make([]string, 0),
	}

	s.flattenMap(record, "", 0, config, result)
	sort.Strings(result.FieldList)
	return result, nil
}

// TransformRecords transforms multiple records and returns unified field list.
func (s *TransformService) TransformRecords(records []map[string]interface{}, config *TransformConfig) ([]*FlattenedRecord, []string, error) {
	if len(records) == 0 {
		return nil, nil, nil
	}

	results := make([]*FlattenedRecord, 0, len(records))
	allFields := make(map[string]struct{})

	for _, record := range records {
		transformed, err := s.TransformRecord(record, config)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to transform record: %w", err)
		}
		results = append(results, transformed)

		for _, field := range transformed.FieldList {
			allFields[field] = struct{}{}
		}
	}

	fieldList := make([]string, 0, len(allFields))
	for field := range allFields {
		fieldList = append(fieldList, field)
	}
	sort.Strings(fieldList)

	return results, fieldList, nil
}

// GetAllTimestampFields collects all unique timestamp field names from transformed records.
func GetAllTimestampFields(records []*FlattenedRecord) []string {
	timestampFields := make(map[string]struct{})
	for _, record := range records {
		for fieldName := range record.TimestampFields {
			timestampFields[fieldName] = struct{}{}
		}
	}

	fieldList := make([]string, 0, len(timestampFields))
	for field := range timestampFields {
		fieldList = append(fieldList, field)
	}
	sort.Strings(fieldList)
	return fieldList
}

func (s *TransformService) flattenMap(data map[string]interface{}, prefix string, depth int, config *TransformConfig, result *FlattenedRecord) {
	for key, value := range data {
		fieldName := s.buildFieldName(prefix, key, config)

		if s.isExcluded(fieldName, config.ExcludeFields) {
			continue
		}

		if mapping, ok := config.FieldMappings[fieldName]; ok {
			fieldName = mapping
		}

		s.flattenValue(fieldName, value, depth, config, result)
	}
}

func (s *TransformService) flattenValue(fieldName string, value interface{}, depth int, config *TransformConfig, result *FlattenedRecord) {
	if value == nil {
		if config.PreserveNulls {
			result.Fields[fieldName] = nil
			result.FieldList = append(result.FieldList, fieldName)
		}
		return
	}

	switch v := value.(type) {
	case map[string]interface{}:
		if config.MaxDepth == 0 || depth < config.MaxDepth {
			s.flattenMap(v, fieldName, depth+1, config, result)
		} else {
			jsonBytes, _ := json.Marshal(v)
			result.Fields[fieldName+"_json"] = string(jsonBytes)
			result.FieldList = append(result.FieldList, fieldName+"_json")
		}

	case []interface{}:
		if config.FlattenArrays {
			s.flattenArray(fieldName, v, depth, config, result)
		} else {
			jsonBytes, _ := json.Marshal(v)
			result.Fields[fieldName] = string(jsonBytes)
			result.FieldList = append(result.FieldList, fieldName)
		}

	default:
		normalizedValue := s.normalizeValue(v)
		result.Fields[fieldName] = normalizedValue
		result.FieldList = append(result.FieldList, fieldName)

		parser := GetTimestampParser()
		if epochMillis, ok := parser.DetectAndParseTimestamp(fieldName, v); ok {
			result.TimestampFields[fieldName] = epochMillis
		}
	}
}

func (s *TransformService) flattenArray(fieldName string, arr []interface{}, depth int, config *TransformConfig, result *FlattenedRecord) {
	if len(arr) == 0 {
		result.Fields[fieldName] = "[]"
		result.FieldList = append(result.FieldList, fieldName)
		return
	}

	hasObjects := false
	hasPrimitives := false
	for _, item := range arr {
		if _, ok := item.(map[string]interface{}); ok {
			hasObjects = true
		} else {
			hasPrimitives = true
		}
	}

	if hasPrimitives && !hasObjects {
		jsonBytes, _ := json.Marshal(arr)
		result.Fields[fieldName] = string(jsonBytes)
		result.FieldList = append(result.FieldList, fieldName)
		return
	}

	if hasObjects {
		for i, item := range arr {
			if config.MaxArrayIndex >= 0 && i > config.MaxArrayIndex {
				break
			}
			if obj, ok := item.(map[string]interface{}); ok {
				indexedPrefix := fmt.Sprintf("%s_%d", fieldName, i)
				if config.MaxDepth == 0 || depth < config.MaxDepth {
					s.flattenMap(obj, indexedPrefix, depth+1, config, result)
				} else {
					jsonBytes, _ := json.Marshal(obj)
					result.Fields[indexedPrefix] = string(jsonBytes)
					result.FieldList = append(result.FieldList, indexedPrefix)
				}
			}
		}
		result.Fields[fieldName+"_count"] = len(arr)
		result.FieldList = append(result.FieldList, fieldName+"_count")
	}
}

func (s *TransformService) buildFieldName(prefix, key string, config *TransformConfig) string {
	key = sanitizeFieldName(key)

	var fullName string
	if prefix == "" {
		fullName = key
	} else {
		fullName = prefix + "_" + key
	}

	maxLen := config.MaxColumnNameLength
	if maxLen == 0 {
		maxLen = s.maxColumnNameLength
	}

	if len(fullName) > maxLen {
		hash := s.hashString(fullName)
		p := fullName[:maxLen/2-4]
		suffix := fullName[len(fullName)-maxLen/2+4:]
		fullName = p + "_" + hash + "_" + suffix
		if len(fullName) > maxLen {
			fullName = fullName[:maxLen]
		}
	}

	return fullName
}

func (s *TransformService) hashString(str string) string {
	var sum uint32
	for _, b := range []byte(str) {
		sum = sum*31 + uint32(b)
	}
	return fmt.Sprintf("%04d", sum%10000)
}

// sanitizeFieldName ensures a field name is valid for SQL/Parquet.
func sanitizeFieldName(name string) string {
	name = strings.ToLower(name)
	name = strings.ReplaceAll(name, "-", "_")
	name = strings.ReplaceAll(name, ".", "_")
	name = strings.ReplaceAll(name, " ", "_")

	reg := regexp.MustCompile(`[^a-z0-9_]`)
	name = reg.ReplaceAllString(name, "")

	for strings.Contains(name, "__") {
		name = strings.ReplaceAll(name, "__", "_")
	}
	name = strings.Trim(name, "_")

	if len(name) > 0 && name[0] >= '0' && name[0] <= '9' {
		name = "_" + name
	}

	return name
}

func (s *TransformService) normalizeValue(value interface{}) interface{} {
	switch v := value.(type) {
	case float64:
		if v == float64(int64(v)) {
			return int64(v)
		}
		return v
	case string:
		return strings.TrimSpace(v)
	default:
		return v
	}
}

func (s *TransformService) isExcluded(fieldName string, excludeList []string) bool {
	for _, excluded := range excludeList {
		if fieldName == excluded || strings.HasPrefix(fieldName, excluded+"_") {
			return true
		}
	}
	return false
}

// PlatformTransformConfigs provides platform-specific transform configurations.
var PlatformTransformConfigs = map[string]*TransformConfig{
	"keap": {
		MaxDepth:      0,  // Unlimited
		MaxArrayIndex: -1, // Unlimited
		FlattenArrays: true,
		FieldMappings: map[string]string{
			"date_created": "created_at",
		},
	},
	"gohighlevel": {
		MaxDepth:      0,
		MaxArrayIndex: -1,
		FlattenArrays: true,
	},
	"activecampaign": {
		MaxDepth:      0,
		MaxArrayIndex: -1,
		FlattenArrays: true,
	},
	"default": {
		MaxDepth:      0,
		MaxArrayIndex: -1,
		FlattenArrays: true,
	},
}

// GetPlatformConfig returns the transform config for a platform.
func GetPlatformConfig(platform string) *TransformConfig {
	if config, ok := PlatformTransformConfigs[strings.ToLower(platform)]; ok {
		return config
	}
	return PlatformTransformConfigs["default"]
}
