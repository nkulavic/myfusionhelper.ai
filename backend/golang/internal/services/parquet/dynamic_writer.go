package parquet

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/apache/arrow/go/v17/arrow"
	"github.com/apache/arrow/go/v17/arrow/array"
	"github.com/apache/arrow/go/v17/arrow/memory"
	arrowparquet "github.com/apache/arrow/go/v17/parquet"
	"github.com/apache/arrow/go/v17/parquet/compress"
	"github.com/apache/arrow/go/v17/parquet/pqarrow"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// WriteDynamicParquet writes raw API records to a parquet file using a dynamic
// schema. All CRM fields are flattened and written as columns. This captures
// every field from the API (unlike the typed WriteContactsParquet which only
// captures fields in NormalizedContact).
//
// The flow:
//  1. TransformService flattens nested JSON to flat column names
//  2. Arrow schema is built dynamically from discovered fields
//  3. Parquet file is written with Snappy compression
//  4. SchemaInfo is returned for writing schema.json
//
// System columns (_record_id, _sync_timestamp) are always present.
// Timestamp columns (timestamp_<field>) are added for detected date fields.
func WriteDynamicParquet(
	ctx context.Context,
	s3Client *s3.Client,
	bucket, key string,
	records []map[string]interface{},
	platformSlug string,
) (*SchemaInfo, error) {
	if len(records) == 0 {
		return &SchemaInfo{RecordCount: 0, Columns: map[string]ColumnInfo{}}, nil
	}

	transformSvc := NewTransformService()
	transformConfig := GetPlatformConfig(platformSlug)

	// Transform all records
	transformedRecords, allFields, err := transformSvc.TransformRecords(records, transformConfig)
	if err != nil {
		return nil, fmt.Errorf("dynamic parquet: transform failed: %w", err)
	}

	log.Printf("[PARQUET] Transformed %d records with %d flattened fields", len(transformedRecords), len(allFields))

	syncTimestamp := time.Now().UTC().Format(time.RFC3339)

	// Get all detected timestamp source fields
	timestampSourceFields := GetAllTimestampFields(transformedRecords)

	// Build column list: system + flattened + timestamp
	systemCols := []string{"_record_id", "_sync_timestamp"}
	timestampCols := make([]string, 0, len(timestampSourceFields))
	for _, field := range timestampSourceFields {
		timestampCols = append(timestampCols, "timestamp_"+field)
	}

	allCols := make([]string, 0, len(systemCols)+len(allFields)+len(timestampCols))
	allCols = append(allCols, systemCols...)
	allCols = append(allCols, allFields...)
	allCols = append(allCols, timestampCols...)

	// Build Arrow schema
	arrowFields := buildDynamicSchemaFields(allCols)
	schema := arrow.NewSchema(arrowFields, nil)

	mem := memory.DefaultAllocator
	builders := createDynamicBuilders(mem, allCols)
	defer func() {
		for _, b := range builders {
			b.Release()
		}
	}()

	// Write each record
	for idx, rawRecord := range records {
		transformed := transformedRecords[idx]
		colIdx := 0

		// _record_id (index 0)
		recordID := extractRecordID(rawRecord)
		builders[colIdx].(*array.StringBuilder).Append(recordID)
		colIdx++

		// _sync_timestamp (index 1)
		builders[colIdx].(*array.StringBuilder).Append(syncTimestamp)
		colIdx++

		// Flattened fields (indices 2 .. 2+len(allFields)-1) - all STRING
		for _, colName := range allFields {
			value := ""
			if v, ok := transformed.Fields[colName]; ok {
				value = convertAnyToString(v)
			}
			builders[colIdx].(*array.StringBuilder).Append(value)
			colIdx++
		}

		// Timestamp columns - INT64 epoch milliseconds
		for _, sourceField := range timestampSourceFields {
			if epochMillis, ok := transformed.TimestampFields[sourceField]; ok && epochMillis > 0 {
				builders[colIdx].(*array.Int64Builder).Append(epochMillis)
			} else {
				builders[colIdx].(*array.Int64Builder).AppendNull()
			}
			colIdx++
		}
	}

	// Build arrays
	arrays := make([]arrow.Array, len(allCols))
	for i, builder := range builders {
		arrays[i] = builder.NewArray()
	}
	defer func() {
		for _, a := range arrays {
			a.Release()
		}
	}()

	// Create record batch
	record := array.NewRecord(schema, arrays, int64(len(records)))
	defer record.Release()

	// Write to parquet
	var buf bytes.Buffer
	props := arrowparquet.NewWriterProperties(
		arrowparquet.WithCompression(compress.Codecs.Snappy),
	)
	arrowProps := pqarrow.NewArrowWriterProperties(pqarrow.WithStoreSchema())

	writer, err := pqarrow.NewFileWriter(schema, &buf, props, arrowProps)
	if err != nil {
		return nil, fmt.Errorf("dynamic parquet: failed to create writer: %w", err)
	}
	if err := writer.Write(record); err != nil {
		writer.Close()
		return nil, fmt.Errorf("dynamic parquet: failed to write record batch: %w", err)
	}
	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("dynamic parquet: failed to close writer: %w", err)
	}

	log.Printf("[PARQUET] Wrote %d records with %d columns (2 system + %d flattened + %d timestamp)",
		len(records), len(allCols), len(allFields), len(timestampCols))

	// Upload to S3
	if err := putParquetToS3(ctx, s3Client, bucket, key, &buf); err != nil {
		return nil, err
	}

	// Build SchemaInfo
	schemaInfo := buildDynamicSchemaInfo(allFields, timestampSourceFields, transformedRecords, len(records), syncTimestamp)
	return schemaInfo, nil
}

// buildDynamicSchemaFields creates Arrow schema fields with hybrid typing:
// system strings, timestamp columns as INT64, everything else STRING.
func buildDynamicSchemaFields(columns []string) []arrow.Field {
	fields := make([]arrow.Field, len(columns))
	for i, colName := range columns {
		var dataType arrow.DataType
		switch {
		case strings.HasPrefix(colName, "timestamp_"):
			dataType = arrow.PrimitiveTypes.Int64
		default:
			dataType = arrow.BinaryTypes.String
		}
		fields[i] = arrow.Field{
			Name:     colName,
			Type:     dataType,
			Nullable: true,
		}
	}
	return fields
}

// createDynamicBuilders creates appropriate array builders based on column types.
func createDynamicBuilders(mem memory.Allocator, columns []string) []array.Builder {
	builders := make([]array.Builder, len(columns))
	for i, colName := range columns {
		if strings.HasPrefix(colName, "timestamp_") {
			builders[i] = array.NewInt64Builder(mem)
		} else {
			builders[i] = array.NewStringBuilder(mem)
		}
	}
	return builders
}

// extractRecordID extracts a record ID from a raw API map.
func extractRecordID(record map[string]interface{}) string {
	idFields := []string{"id", "Id", "ID", "_id", "uuid"}
	for _, field := range idFields {
		if val, ok := record[field]; ok {
			switch v := val.(type) {
			case string:
				return v
			case float64:
				if v == float64(int64(v)) {
					return strconv.FormatInt(int64(v), 10)
				}
				return fmt.Sprintf("%g", v)
			case int:
				return strconv.Itoa(v)
			case int64:
				return strconv.FormatInt(v, 10)
			case json.Number:
				return v.String()
			}
		}
	}
	return ""
}

// convertAnyToString converts any value to a string for parquet storage.
func convertAnyToString(value interface{}) string {
	if value == nil {
		return ""
	}
	switch v := value.(type) {
	case string:
		return v
	case bool:
		return strconv.FormatBool(v)
	case int:
		return strconv.Itoa(v)
	case int64:
		return strconv.FormatInt(v, 10)
	case float32:
		if v == float32(int64(v)) {
			return strconv.FormatInt(int64(v), 10)
		}
		return strconv.FormatFloat(float64(v), 'f', -1, 32)
	case float64:
		if v == float64(int64(v)) {
			return strconv.FormatInt(int64(v), 10)
		}
		return strconv.FormatFloat(v, 'f', -1, 64)
	case json.Number:
		return v.String()
	case time.Time:
		if v.IsZero() {
			return ""
		}
		return v.Format(time.RFC3339)
	case []byte:
		return string(v)
	default:
		jsonBytes, err := json.Marshal(v)
		if err != nil {
			return fmt.Sprintf("%v", v)
		}
		return string(jsonBytes)
	}
}

// buildDynamicSchemaInfo constructs SchemaInfo from dynamic columns.
func buildDynamicSchemaInfo(
	allFields []string,
	timestampSourceFields []string,
	transformedRecords []*FlattenedRecord,
	recordCount int,
	syncedAt string,
) *SchemaInfo {
	columns := make(map[string]ColumnInfo)

	// System columns
	columns["_record_id"] = ColumnInfo{
		Type:        "string",
		DisplayName: "Record ID",
		Nullable:    false,
	}
	columns["_sync_timestamp"] = ColumnInfo{
		Type:        "date",
		DisplayName: "Sync Timestamp",
		Nullable:    false,
	}

	// Flattened data columns with sample values
	sampleSize := 5
	if len(transformedRecords) < sampleSize {
		sampleSize = len(transformedRecords)
	}

	for _, fieldName := range allFields {
		col := ColumnInfo{
			Type:        "string",
			DisplayName: generateDisplayName(fieldName),
			Nullable:    true,
		}

		// Collect sample values
		var samples []string
		seen := make(map[string]bool)
		for i := 0; i < sampleSize; i++ {
			if val, ok := transformedRecords[i].Fields[fieldName]; ok {
				valStr := fmt.Sprintf("%v", val)
				if valStr != "" && !seen[valStr] && len(valStr) < 100 {
					samples = append(samples, valStr)
					seen[valStr] = true
				}
			}
		}
		if len(samples) > 0 {
			col.SampleValues = samples
		}
		columns[fieldName] = col
	}

	// Timestamp columns
	for _, sourceField := range timestampSourceFields {
		colName := "timestamp_" + sourceField
		columns[colName] = ColumnInfo{
			Type:        "date",
			DisplayName: generateDisplayName(sourceField) + " (Timestamp)",
			Nullable:    true,
		}
	}

	return &SchemaInfo{
		RecordCount: recordCount,
		SyncedAt:    syncedAt,
		Columns:     columns,
	}
}

// generateDisplayName converts a snake_case field name to Title Case.
// Array-indexed fields like "addresses_0_city" become "Address 1 - City".
func generateDisplayName(fieldName string) string {
	parts := strings.Split(fieldName, "_")
	result := make([]string, 0, len(parts))

	for _, part := range parts {
		if idx, err := strconv.Atoi(part); err == nil {
			// Numeric index — append 1-based to previous word
			if len(result) > 0 {
				prev := result[len(result)-1]
				if len(prev) > 1 && strings.HasSuffix(prev, "s") {
					result[len(result)-1] = prev[:len(prev)-1]
				}
				result[len(result)-1] = fmt.Sprintf("%s %d", result[len(result)-1], idx+1)
			}
		} else if len(part) > 0 {
			result = append(result, strings.ToUpper(part[:1])+part[1:])
		}
	}

	// If there's an array index, join with " - "
	hasIdx := false
	for _, r := range result {
		if matched, _ := regexp.MatchString(`\d+$`, r); matched {
			hasIdx = true
			break
		}
	}
	if hasIdx && len(result) > 1 {
		for i, r := range result {
			if matched, _ := regexp.MatchString(`\d+$`, r); matched {
				prefix := strings.Join(result[:i+1], " ")
				suffix := strings.Join(result[i+1:], " ")
				if suffix != "" {
					return prefix + " - " + suffix
				}
				return prefix
			}
		}
	}

	return strings.Join(result, " ")
}

// extractTimestampFromMap extracts a timestamp from raw record, checking multiple field names.
func extractTimestampFromMap(record map[string]interface{}, fields ...string) int64 {
	for _, field := range fields {
		if val, ok := record[field]; ok {
			switch v := val.(type) {
			case string:
				formats := []string{
					time.RFC3339,
					"2006-01-02T15:04:05Z",
					"2006-01-02T15:04:05",
					"2006-01-02 15:04:05",
					"2006-01-02",
				}
				for _, format := range formats {
					if t, err := time.Parse(format, v); err == nil {
						return t.UnixMilli()
					}
				}
			case float64:
				if v > 1e12 {
					return int64(v)
				}
				return int64(v * 1000)
			case int64:
				if v > 1e12 {
					return v
				}
				return v * 1000
			}
		}
	}
	return 0
}

// SortedKeys returns sorted keys from a string-keyed map.
func SortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// WriteDynamicSchema writes schema.json for dynamically-written parquet to S3.
func WriteDynamicSchema(ctx context.Context, s3Client *s3.Client, bucket, key string, schema *SchemaInfo) error {
	data, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		return fmt.Errorf("dynamic schema: failed to marshal: %w", err)
	}
	_, err = s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(data),
		ContentType: aws.String("application/json"),
	})
	if err != nil {
		return fmt.Errorf("dynamic schema: failed to upload to s3://%s/%s: %w", bucket, key, err)
	}
	return nil
}
