package parquet

import (
	"bytes"
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/apache/arrow/go/v17/arrow"
	"github.com/apache/arrow/go/v17/arrow/array"
	"github.com/apache/arrow/go/v17/arrow/memory"
	"github.com/apache/arrow/go/v17/parquet/compress"
	"github.com/apache/arrow/go/v17/parquet/pqarrow"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"github.com/apache/arrow/go/v17/parquet"

	"github.com/myfusionhelper/api/internal/connectors"
)

// maxSampleValues is the maximum number of sample values stored per column
// in SchemaInfo metadata.
const maxSampleValues = 5

// WriteContactsParquet writes a slice of NormalizedContact as a parquet file
// to S3. Fixed columns include standard contact fields. Dynamic columns are
// derived from CustomFields across all contacts. All values are stored as
// UTF8 strings for maximum DuckDB compatibility.
//
// Returns a SchemaInfo describing the written schema and record count.
func WriteContactsParquet(
	ctx context.Context,
	s3Client *s3.Client,
	bucket, key string,
	contacts []connectors.NormalizedContact,
) (*SchemaInfo, error) {
	syncTimestamp := time.Now().UTC().Format(time.RFC3339)

	// Collect all unique custom field keys across every contact so that the
	// schema is a superset of all custom fields present in the dataset.
	customKeySet := make(map[string]struct{})
	for i := range contacts {
		for k := range contacts[i].CustomFields {
			customKeySet[k] = struct{}{}
		}
	}
	customKeys := make([]string, 0, len(customKeySet))
	for k := range customKeySet {
		customKeys = append(customKeys, k)
	}
	sort.Strings(customKeys)

	// Build Arrow schema: fixed columns + dynamic custom field columns + _sync_timestamp.
	fields := []arrow.Field{
		{Name: "_record_id", Type: arrow.BinaryTypes.String, Nullable: false},
		{Name: "first_name", Type: arrow.BinaryTypes.String, Nullable: true},
		{Name: "last_name", Type: arrow.BinaryTypes.String, Nullable: true},
		{Name: "email", Type: arrow.BinaryTypes.String, Nullable: true},
		{Name: "phone", Type: arrow.BinaryTypes.String, Nullable: true},
		{Name: "company", Type: arrow.BinaryTypes.String, Nullable: true},
		{Name: "job_title", Type: arrow.BinaryTypes.String, Nullable: true},
		{Name: "source_crm", Type: arrow.BinaryTypes.String, Nullable: true},
		{Name: "source_id", Type: arrow.BinaryTypes.String, Nullable: true},
		{Name: "tags", Type: arrow.BinaryTypes.String, Nullable: true},
		{Name: "created_at", Type: arrow.BinaryTypes.String, Nullable: true},
		{Name: "updated_at", Type: arrow.BinaryTypes.String, Nullable: true},
	}
	for _, k := range customKeys {
		fields = append(fields, arrow.Field{
			Name:     "cf_" + k,
			Type:     arrow.BinaryTypes.String,
			Nullable: true,
		})
	}
	fields = append(fields, arrow.Field{
		Name:     "_sync_timestamp",
		Type:     arrow.BinaryTypes.String,
		Nullable: false,
	})

	arrowSchema := arrow.NewSchema(fields, nil)

	// Build arrays for every column.
	mem := memory.DefaultAllocator
	numRows := len(contacts)

	builders := make([]*array.StringBuilder, len(fields))
	for i := range builders {
		builders[i] = array.NewStringBuilder(mem)
	}
	defer func() {
		for _, b := range builders {
			b.Release()
		}
	}()

	for _, c := range contacts {
		idx := 0
		builders[idx].Append(c.ID); idx++
		builders[idx].Append(c.FirstName); idx++
		builders[idx].Append(c.LastName); idx++
		builders[idx].Append(c.Email); idx++
		builders[idx].Append(c.Phone); idx++
		builders[idx].Append(c.Company); idx++
		builders[idx].Append(c.JobTitle); idx++
		builders[idx].Append(c.SourceCRM); idx++
		builders[idx].Append(c.SourceID); idx++

		// Tags: comma-separated tag names.
		tagNames := make([]string, 0, len(c.Tags))
		for _, t := range c.Tags {
			tagNames = append(tagNames, t.Name)
		}
		builders[idx].Append(strings.Join(tagNames, ",")); idx++

		// created_at / updated_at as RFC3339 or empty string.
		if c.CreatedAt != nil {
			builders[idx].Append(c.CreatedAt.UTC().Format(time.RFC3339))
		} else {
			builders[idx].AppendNull()
		}
		idx++

		if c.UpdatedAt != nil {
			builders[idx].Append(c.UpdatedAt.UTC().Format(time.RFC3339))
		} else {
			builders[idx].AppendNull()
		}
		idx++

		// Dynamic custom field columns (sorted order).
		for _, k := range customKeys {
			if v, ok := c.CustomFields[k]; ok && v != nil {
				builders[idx].Append(fmt.Sprintf("%v", v))
			} else {
				builders[idx].AppendNull()
			}
			idx++
		}

		// _sync_timestamp
		builders[idx].Append(syncTimestamp)
	}

	// Build the arrays and create a record batch.
	arrays := make([]arrow.Array, len(builders))
	for i, b := range builders {
		arrays[i] = b.NewArray()
	}
	defer func() {
		for _, a := range arrays {
			a.Release()
		}
	}()

	record := array.NewRecord(arrowSchema, arrays, int64(numRows))
	defer record.Release()

	// Write the record batch to a parquet buffer.
	buf, err := writeRecordToParquet(arrowSchema, record)
	if err != nil {
		return nil, fmt.Errorf("parquet writer: contacts: %w", err)
	}

	// Upload to S3.
	if err := putParquetToS3(ctx, s3Client, bucket, key, buf); err != nil {
		return nil, err
	}

	// Build SchemaInfo.
	schema := buildSchemaInfo("contacts", numRows, syncTimestamp, fields, contacts, customKeys)
	return schema, nil
}

// WriteTagsParquet writes a slice of Tag as a parquet file to S3.
func WriteTagsParquet(
	ctx context.Context,
	s3Client *s3.Client,
	bucket, key string,
	tags []connectors.Tag,
) (*SchemaInfo, error) {
	syncTimestamp := time.Now().UTC().Format(time.RFC3339)

	fields := []arrow.Field{
		{Name: "_record_id", Type: arrow.BinaryTypes.String, Nullable: false},
		{Name: "name", Type: arrow.BinaryTypes.String, Nullable: true},
		{Name: "description", Type: arrow.BinaryTypes.String, Nullable: true},
		{Name: "category", Type: arrow.BinaryTypes.String, Nullable: true},
		{Name: "_sync_timestamp", Type: arrow.BinaryTypes.String, Nullable: false},
	}
	arrowSchema := arrow.NewSchema(fields, nil)

	mem := memory.DefaultAllocator
	numRows := len(tags)

	idBuilder := array.NewStringBuilder(mem)
	nameBuilder := array.NewStringBuilder(mem)
	descBuilder := array.NewStringBuilder(mem)
	catBuilder := array.NewStringBuilder(mem)
	syncBuilder := array.NewStringBuilder(mem)
	defer func() {
		idBuilder.Release()
		nameBuilder.Release()
		descBuilder.Release()
		catBuilder.Release()
		syncBuilder.Release()
	}()

	for _, t := range tags {
		idBuilder.Append(t.ID)
		nameBuilder.Append(t.Name)
		descBuilder.Append(t.Description)
		catBuilder.Append(t.Category)
		syncBuilder.Append(syncTimestamp)
	}

	arrays := []arrow.Array{
		idBuilder.NewArray(),
		nameBuilder.NewArray(),
		descBuilder.NewArray(),
		catBuilder.NewArray(),
		syncBuilder.NewArray(),
	}
	defer func() {
		for _, a := range arrays {
			a.Release()
		}
	}()

	record := array.NewRecord(arrowSchema, arrays, int64(numRows))
	defer record.Release()

	buf, err := writeRecordToParquet(arrowSchema, record)
	if err != nil {
		return nil, fmt.Errorf("parquet writer: tags: %w", err)
	}

	if err := putParquetToS3(ctx, s3Client, bucket, key, buf); err != nil {
		return nil, err
	}

	columns := map[string]ColumnInfo{
		"_record_id": {
			Type:         "string",
			DisplayName:  "Record ID",
			Nullable:     false,
			SampleValues: sampleStrings(tags, func(t connectors.Tag) string { return t.ID }),
		},
		"name": {
			Type:         "string",
			DisplayName:  "Name",
			Nullable:     true,
			SampleValues: sampleStrings(tags, func(t connectors.Tag) string { return t.Name }),
		},
		"description": {
			Type:         "string",
			DisplayName:  "Description",
			Nullable:     true,
			SampleValues: sampleStrings(tags, func(t connectors.Tag) string { return t.Description }),
		},
		"category": {
			Type:         "string",
			DisplayName:  "Category",
			Nullable:     true,
			SampleValues: sampleStrings(tags, func(t connectors.Tag) string { return t.Category }),
		},
		"_sync_timestamp": {
			Type:        "date",
			DisplayName: "Sync Timestamp",
			Nullable:    false,
		},
	}

	return &SchemaInfo{
		ObjectType:  "tags",
		RecordCount: numRows,
		SyncedAt:    syncTimestamp,
		Columns:     columns,
	}, nil
}

// WriteCustomFieldsParquet writes a slice of CustomField as a parquet file to S3.
func WriteCustomFieldsParquet(
	ctx context.Context,
	s3Client *s3.Client,
	bucket, key string,
	customFields []connectors.CustomField,
) (*SchemaInfo, error) {
	syncTimestamp := time.Now().UTC().Format(time.RFC3339)

	fields := []arrow.Field{
		{Name: "_record_id", Type: arrow.BinaryTypes.String, Nullable: false},
		{Name: "key", Type: arrow.BinaryTypes.String, Nullable: true},
		{Name: "label", Type: arrow.BinaryTypes.String, Nullable: true},
		{Name: "field_type", Type: arrow.BinaryTypes.String, Nullable: true},
		{Name: "group_name", Type: arrow.BinaryTypes.String, Nullable: true},
		{Name: "default_value", Type: arrow.BinaryTypes.String, Nullable: true},
		{Name: "options", Type: arrow.BinaryTypes.String, Nullable: true},
		{Name: "_sync_timestamp", Type: arrow.BinaryTypes.String, Nullable: false},
	}
	arrowSchema := arrow.NewSchema(fields, nil)

	mem := memory.DefaultAllocator
	numRows := len(customFields)

	idBuilder := array.NewStringBuilder(mem)
	keyBuilder := array.NewStringBuilder(mem)
	labelBuilder := array.NewStringBuilder(mem)
	ftBuilder := array.NewStringBuilder(mem)
	gnBuilder := array.NewStringBuilder(mem)
	dvBuilder := array.NewStringBuilder(mem)
	optBuilder := array.NewStringBuilder(mem)
	syncBuilder := array.NewStringBuilder(mem)
	defer func() {
		idBuilder.Release()
		keyBuilder.Release()
		labelBuilder.Release()
		ftBuilder.Release()
		gnBuilder.Release()
		dvBuilder.Release()
		optBuilder.Release()
		syncBuilder.Release()
	}()

	for _, cf := range customFields {
		idBuilder.Append(cf.ID)
		keyBuilder.Append(cf.Key)
		labelBuilder.Append(cf.Label)
		ftBuilder.Append(cf.FieldType)
		gnBuilder.Append(cf.GroupName)
		dvBuilder.Append(cf.DefaultValue)
		optBuilder.Append(strings.Join(cf.Options, ","))
		syncBuilder.Append(syncTimestamp)
	}

	arrays := []arrow.Array{
		idBuilder.NewArray(),
		keyBuilder.NewArray(),
		labelBuilder.NewArray(),
		ftBuilder.NewArray(),
		gnBuilder.NewArray(),
		dvBuilder.NewArray(),
		optBuilder.NewArray(),
		syncBuilder.NewArray(),
	}
	defer func() {
		for _, a := range arrays {
			a.Release()
		}
	}()

	record := array.NewRecord(arrowSchema, arrays, int64(numRows))
	defer record.Release()

	buf, err := writeRecordToParquet(arrowSchema, record)
	if err != nil {
		return nil, fmt.Errorf("parquet writer: custom_fields: %w", err)
	}

	if err := putParquetToS3(ctx, s3Client, bucket, key, buf); err != nil {
		return nil, err
	}

	columns := map[string]ColumnInfo{
		"_record_id": {
			Type:         "string",
			DisplayName:  "Record ID",
			Nullable:     false,
			SampleValues: sampleStrings(customFields, func(cf connectors.CustomField) string { return cf.ID }),
		},
		"key": {
			Type:         "string",
			DisplayName:  "Key",
			Nullable:     true,
			SampleValues: sampleStrings(customFields, func(cf connectors.CustomField) string { return cf.Key }),
		},
		"label": {
			Type:         "string",
			DisplayName:  "Label",
			Nullable:     true,
			SampleValues: sampleStrings(customFields, func(cf connectors.CustomField) string { return cf.Label }),
		},
		"field_type": {
			Type:         "string",
			DisplayName:  "Field Type",
			Nullable:     true,
			SampleValues: sampleStrings(customFields, func(cf connectors.CustomField) string { return cf.FieldType }),
		},
		"group_name": {
			Type:         "string",
			DisplayName:  "Group Name",
			Nullable:     true,
			SampleValues: sampleStrings(customFields, func(cf connectors.CustomField) string { return cf.GroupName }),
		},
		"default_value": {
			Type:         "string",
			DisplayName:  "Default Value",
			Nullable:     true,
			SampleValues: sampleStrings(customFields, func(cf connectors.CustomField) string { return cf.DefaultValue }),
		},
		"options": {
			Type:        "string",
			DisplayName: "Options",
			Nullable:    true,
		},
		"_sync_timestamp": {
			Type:        "date",
			DisplayName: "Sync Timestamp",
			Nullable:    false,
		},
	}

	return &SchemaInfo{
		ObjectType:  "custom_fields",
		RecordCount: numRows,
		SyncedAt:    syncTimestamp,
		Columns:     columns,
	}, nil
}

// ---------- internal helpers ----------

// writeRecordToParquet serialises an Arrow record batch into a parquet byte
// buffer using Snappy compression.
func writeRecordToParquet(schema *arrow.Schema, record arrow.Record) (*bytes.Buffer, error) {
	var buf bytes.Buffer

	writerProps := parquet.NewWriterProperties(
		parquet.WithCompression(compress.Codecs.Snappy),
	)
	arrowProps := pqarrow.NewArrowWriterProperties(pqarrow.WithStoreSchema())

	writer, err := pqarrow.NewFileWriter(schema, &buf, writerProps, arrowProps)
	if err != nil {
		return nil, fmt.Errorf("failed to create parquet file writer: %w", err)
	}

	if err := writer.Write(record); err != nil {
		return nil, fmt.Errorf("failed to write record batch: %w", err)
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("failed to close parquet writer: %w", err)
	}

	return &buf, nil
}

// putParquetToS3 uploads a bytes.Buffer to S3 as application/octet-stream.
func putParquetToS3(ctx context.Context, s3Client *s3.Client, bucket, key string, buf *bytes.Buffer) error {
	_, err := s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(buf.Bytes()),
		ContentType: aws.String("application/octet-stream"),
	})
	if err != nil {
		return fmt.Errorf("parquet writer: failed to upload to s3://%s/%s: %w", bucket, key, err)
	}
	return nil
}

// buildSchemaInfo constructs a SchemaInfo for the contacts object type.
func buildSchemaInfo(
	objectType string,
	recordCount int,
	syncedAt string,
	fields []arrow.Field,
	contacts []connectors.NormalizedContact,
	customKeys []string,
) *SchemaInfo {
	columns := map[string]ColumnInfo{
		"_record_id": {
			Type:         "string",
			DisplayName:  "Record ID",
			Nullable:     false,
			SampleValues: sampleStrings(contacts, func(c connectors.NormalizedContact) string { return c.ID }),
		},
		"first_name": {
			Type:         "string",
			DisplayName:  "First Name",
			Nullable:     true,
			SampleValues: sampleStrings(contacts, func(c connectors.NormalizedContact) string { return c.FirstName }),
		},
		"last_name": {
			Type:         "string",
			DisplayName:  "Last Name",
			Nullable:     true,
			SampleValues: sampleStrings(contacts, func(c connectors.NormalizedContact) string { return c.LastName }),
		},
		"email": {
			Type:         "string",
			DisplayName:  "Email",
			Nullable:     true,
			SampleValues: sampleStrings(contacts, func(c connectors.NormalizedContact) string { return c.Email }),
		},
		"phone": {
			Type:         "string",
			DisplayName:  "Phone",
			Nullable:     true,
			SampleValues: sampleStrings(contacts, func(c connectors.NormalizedContact) string { return c.Phone }),
		},
		"company": {
			Type:         "string",
			DisplayName:  "Company",
			Nullable:     true,
			SampleValues: sampleStrings(contacts, func(c connectors.NormalizedContact) string { return c.Company }),
		},
		"job_title": {
			Type:         "string",
			DisplayName:  "Job Title",
			Nullable:     true,
			SampleValues: sampleStrings(contacts, func(c connectors.NormalizedContact) string { return c.JobTitle }),
		},
		"source_crm": {
			Type:         "string",
			DisplayName:  "Source CRM",
			Nullable:     true,
			SampleValues: sampleStrings(contacts, func(c connectors.NormalizedContact) string { return c.SourceCRM }),
		},
		"source_id": {
			Type:         "string",
			DisplayName:  "Source ID",
			Nullable:     true,
			SampleValues: sampleStrings(contacts, func(c connectors.NormalizedContact) string { return c.SourceID }),
		},
		"tags": {
			Type:        "string",
			DisplayName: "Tags",
			Nullable:    true,
		},
		"created_at": {
			Type:        "date",
			DisplayName: "Created At",
			Nullable:    true,
		},
		"updated_at": {
			Type:        "date",
			DisplayName: "Updated At",
			Nullable:    true,
		},
	}

	// Add custom field columns.
	for _, k := range customKeys {
		colName := "cf_" + k
		columns[colName] = ColumnInfo{
			Type:        "string",
			DisplayName: formatDisplayName(k),
			Nullable:    true,
			SampleValues: sampleStrings(contacts, func(c connectors.NormalizedContact) string {
				if v, ok := c.CustomFields[k]; ok && v != nil {
					return fmt.Sprintf("%v", v)
				}
				return ""
			}),
		}
	}

	columns["_sync_timestamp"] = ColumnInfo{
		Type:        "date",
		DisplayName: "Sync Timestamp",
		Nullable:    false,
	}

	return &SchemaInfo{
		ObjectType:  objectType,
		RecordCount: recordCount,
		SyncedAt:    syncedAt,
		Columns:     columns,
	}
}

// sampleStrings collects up to maxSampleValues unique non-empty values using
// the provided extractor function.
func sampleStrings[T any](items []T, extract func(T) string) []string {
	seen := make(map[string]struct{})
	var samples []string
	for _, item := range items {
		if len(samples) >= maxSampleValues {
			break
		}
		v := extract(item)
		if v == "" {
			continue
		}
		if _, exists := seen[v]; exists {
			continue
		}
		seen[v] = struct{}{}
		samples = append(samples, v)
	}
	return samples
}

// formatDisplayName converts a snake_case key into a Title Case display name.
func formatDisplayName(key string) string {
	parts := strings.Split(key, "_")
	for i, p := range parts {
		if len(p) > 0 {
			parts[i] = strings.ToUpper(p[:1]) + p[1:]
		}
	}
	return strings.Join(parts, " ")
}
