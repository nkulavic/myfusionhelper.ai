package parquet

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// SchemaInfo holds metadata about a parquet file's schema and contents.
// It is stored alongside parquet data files as schema.json.
type SchemaInfo struct {
	ConnectionID string                `json:"connection_id"`
	ObjectType   string                `json:"object_type"`
	RecordCount  int                   `json:"record_count"`
	SyncedAt     string                `json:"synced_at"`
	Columns      map[string]ColumnInfo `json:"columns"`
}

// ColumnInfo describes a single column in the parquet schema.
type ColumnInfo struct {
	Type         string   `json:"type"`                    // "string", "number", "date", "boolean"
	DisplayName  string   `json:"display_name"`
	Nullable     bool     `json:"nullable"`
	SampleValues []string `json:"sample_values,omitempty"`
}

// WriteSchema serializes a SchemaInfo as JSON and writes it to S3.
func WriteSchema(ctx context.Context, s3Client *s3.Client, bucket, key string, schema *SchemaInfo) error {
	data, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		return fmt.Errorf("parquet schema: failed to marshal schema: %w", err)
	}

	_, err = s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(data),
		ContentType: aws.String("application/json"),
	})
	if err != nil {
		return fmt.Errorf("parquet schema: failed to write schema to s3://%s/%s: %w", bucket, key, err)
	}

	return nil
}

// ReadSchema reads a SchemaInfo JSON file from S3 and deserializes it.
func ReadSchema(ctx context.Context, s3Client *s3.Client, bucket, key string) (*SchemaInfo, error) {
	result, err := s3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("parquet schema: failed to read schema from s3://%s/%s: %w", bucket, key, err)
	}
	defer result.Body.Close()

	data, err := io.ReadAll(result.Body)
	if err != nil {
		return nil, fmt.Errorf("parquet schema: failed to read response body: %w", err)
	}

	var schema SchemaInfo
	if err := json.Unmarshal(data, &schema); err != nil {
		return nil, fmt.Errorf("parquet schema: failed to unmarshal schema: %w", err)
	}

	return &schema, nil
}
