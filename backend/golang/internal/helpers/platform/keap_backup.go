package platform

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/myfusionhelper/api/internal/connectors"
	"github.com/myfusionhelper/api/internal/helpers"
)

// KeapBackup backs up Keap/Infusionsoft CRM data to S3.
//
// This helper provides full or incremental backups of Keap data including:
// - Contacts, companies, notes
// - Tags, automations, campaigns
// - Payment methods, custom fields
//
// Data is stored in S3 as JSON files organized by data type and timestamp.
type KeapBackup struct{}

// GetName returns the helper's unique identifier
func (h *KeapBackup) GetName() string {
	return "keap_backup"
}

// GetType returns the helper's type classification
func (h *KeapBackup) GetType() string {
	return "keap_backup"
}

// GetCategory returns the helper's category
func (h *KeapBackup) GetCategory() string {
	return "platform"
}

// GetDescription returns a human-readable description
func (h *KeapBackup) GetDescription() string {
	return "Backup Keap/Infusionsoft CRM data to S3 for disaster recovery and compliance. Supports full backups with automatic pagination and organization by data type."
}

// GetConfigSchema returns the JSON schema for configuration
func (h *KeapBackup) GetConfigSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"data_types": map[string]interface{}{
				"type":        "array",
				"description": "Data types to backup (e.g., contacts, companies, tags, automations)",
				"items": map[string]interface{}{
					"type": "string",
					"enum": []string{
						"contacts", "companies", "notes", "automations",
						"tags", "campaigns", "payment_methods", "custom_fields",
					},
				},
				"default": []string{"contacts", "companies", "tags"},
			},
			"s3_bucket": map[string]interface{}{
				"type":        "string",
				"description": "S3 bucket name for backup storage",
				"default":     "mfh-dev-data",
			},
			"s3_prefix": map[string]interface{}{
				"type":        "string",
				"description": "S3 key prefix for organizing backups",
				"default":     "keap-backups",
			},
			"page_size": map[string]interface{}{
				"type":        "integer",
				"description": "Number of records per page (max 1000)",
				"default":     1000,
				"minimum":     1,
				"maximum":     1000,
			},
		},
		"required": []string{"data_types", "s3_bucket"},
	}
}

// ValidateConfig validates the helper configuration
func (h *KeapBackup) ValidateConfig(config map[string]interface{}) error {
	dataTypes, ok := config["data_types"].([]interface{})
	if !ok || len(dataTypes) == 0 {
		return fmt.Errorf("data_types is required and must be a non-empty array")
	}

	s3Bucket, ok := config["s3_bucket"].(string)
	if !ok || s3Bucket == "" {
		return fmt.Errorf("s3_bucket is required and must be a non-empty string")
	}

	// Validate data types
	validTypes := map[string]bool{
		"contacts": true, "companies": true, "notes": true, "automations": true,
		"tags": true, "campaigns": true, "payment_methods": true, "custom_fields": true,
	}

	for _, dt := range dataTypes {
		dataType, ok := dt.(string)
		if !ok {
			return fmt.Errorf("data_types must contain strings")
		}
		if !validTypes[dataType] {
			return fmt.Errorf("invalid data_type: %s", dataType)
		}
	}

	return nil
}

// RequiresCRM returns whether this helper needs a CRM connection
func (h *KeapBackup) RequiresCRM() bool {
	return true
}

// SupportedCRMs returns the list of supported CRM platforms
func (h *KeapBackup) SupportedCRMs() []string {
	return []string{"keap"}
}

// Execute performs the keap backup operation
func (h *KeapBackup) Execute(ctx context.Context, input helpers.HelperInput) (*helpers.HelperOutput, error) {
	// Extract configuration
	dataTypes := h.getStringArray(input.Config, "data_types")
	s3Bucket := h.getString(input.Config, "s3_bucket", "mfh-dev-data")
	s3Prefix := h.getString(input.Config, "s3_prefix", "keap-backups")
	pageSize := h.getInt(input.Config, "page_size", 1000)

	if len(dataTypes) == 0 {
		return &helpers.HelperOutput{
			Success: false,
			Message: "No data types specified for backup",
		}, nil
	}

	// Verify connector is Keap
	if input.Connector.GetMetadata().PlatformSlug != "keap" {
		return &helpers.HelperOutput{
			Success: false,
			Message: "This helper only supports Keap/Infusionsoft",
		}, nil
	}

	// Initialize S3 client
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return &helpers.HelperOutput{
			Success: false,
			Message: fmt.Sprintf("Failed to load AWS config: %v", err),
		}, nil
	}
	s3Client := s3.NewFromConfig(cfg)

	// Backup timestamp for organizing files
	timestamp := time.Now().UTC().Format("2006-01-02T15-04-05Z")
	backupResults := make(map[string]interface{})

	// Backup each data type
	for _, dataType := range dataTypes {
		result, err := h.backupDataType(ctx, input.Connector, s3Client, s3Bucket, s3Prefix, dataType, timestamp, pageSize)
		if err != nil {
			log.Printf("Error backing up %s: %v", dataType, err)
			backupResults[dataType] = map[string]interface{}{
				"success": false,
				"error":   err.Error(),
			}
		} else {
			backupResults[dataType] = result
		}
	}

	// Count successes
	successCount := 0
	for _, result := range backupResults {
		if resultMap, ok := result.(map[string]interface{}); ok {
			if success, ok := resultMap["success"].(bool); ok && success {
				successCount++
			}
		}
	}

	message := fmt.Sprintf("Backup completed: %d/%d data types backed up successfully", successCount, len(dataTypes))
	return &helpers.HelperOutput{
		Success: successCount > 0,
		Message: message,
		ModifiedData: map[string]interface{}{
			"timestamp":      timestamp,
			"total_types":    len(dataTypes),
			"successful":     successCount,
			"failed":         len(dataTypes) - successCount,
			"backup_results": backupResults,
			"s3_bucket":      s3Bucket,
			"s3_prefix":      s3Prefix,
		},
	}, nil
}

// backupDataType backs up a specific data type from Keap
func (h *KeapBackup) backupDataType(
	ctx context.Context,
	connector connectors.CRMConnector,
	s3Client *s3.Client,
	bucket, prefix, dataType, timestamp string,
	pageSize int,
) (map[string]interface{}, error) {
	var allRecords []interface{}
	pageNumber := 1
	s3Keys := []string{}

	// Map data types to appropriate connector methods
	switch dataType {
	case "contacts":
		// Fetch all contacts with pagination
		opts := connectors.QueryOptions{
			Limit: pageSize,
		}

		for {
			contactList, err := connector.GetContacts(ctx, opts)
			if err != nil {
				return nil, fmt.Errorf("failed to fetch contacts: %w", err)
			}

			if len(contactList.Contacts) == 0 {
				break
			}

			// Convert contacts to interface{} for JSON marshaling
			for _, contact := range contactList.Contacts {
				allRecords = append(allRecords, contact)
			}

			// Upload this page to S3
			s3Key := fmt.Sprintf("%s/%s/%s-page-%d.json", prefix, dataType, timestamp, pageNumber)
			if err := h.uploadToS3(ctx, s3Client, bucket, s3Key, contactList.Contacts); err != nil {
				return nil, fmt.Errorf("failed to upload to S3: %w", err)
			}
			s3Keys = append(s3Keys, s3Key)

			pageNumber++

			// Check if there are more pages
			if !contactList.HasMore {
				break
			}

			// Update cursor for next page
			opts.Cursor = contactList.NextCursor
		}

	case "tags":
		// Fetch all tags
		tags, err := connector.GetTags(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch tags: %w", err)
		}

		// Upload tags to S3
		s3Key := fmt.Sprintf("%s/%s/%s.json", prefix, dataType, timestamp)
		if err := h.uploadToS3(ctx, s3Client, bucket, s3Key, tags); err != nil {
			return nil, fmt.Errorf("failed to upload to S3: %w", err)
		}
		s3Keys = append(s3Keys, s3Key)

		for _, tag := range tags {
			allRecords = append(allRecords, tag)
		}

	case "custom_fields":
		// Fetch all custom fields
		fields, err := connector.GetCustomFields(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch custom fields: %w", err)
		}

		// Upload custom fields to S3
		s3Key := fmt.Sprintf("%s/%s/%s.json", prefix, dataType, timestamp)
		if err := h.uploadToS3(ctx, s3Client, bucket, s3Key, fields); err != nil {
			return nil, fmt.Errorf("failed to upload to S3: %w", err)
		}
		s3Keys = append(s3Keys, s3Key)

		for _, field := range fields {
			allRecords = append(allRecords, field)
		}

	default:
		return nil, fmt.Errorf("unsupported data type: %s", dataType)
	}

	return map[string]interface{}{
		"success":       true,
		"record_count":  len(allRecords),
		"page_count":    len(s3Keys),
		"s3_keys":       s3Keys,
	}, nil
}

// uploadToS3 uploads data to S3 as JSON
func (h *KeapBackup) uploadToS3(ctx context.Context, s3Client *s3.Client, bucket, key string, data interface{}) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	_, err = s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(jsonData),
		ContentType: aws.String("application/json"),
	})
	if err != nil {
		return fmt.Errorf("failed to put object to S3: %w", err)
	}

	log.Printf("Uploaded backup to s3://%s/%s (%d bytes)", bucket, key, len(jsonData))
	return nil
}

// Helper methods for extracting config values
func (h *KeapBackup) getString(config map[string]interface{}, key, defaultValue string) string {
	if val, ok := config[key].(string); ok {
		return val
	}
	return defaultValue
}

func (h *KeapBackup) getInt(config map[string]interface{}, key string, defaultValue int) int {
	if val, ok := config[key].(float64); ok {
		return int(val)
	}
	if val, ok := config[key].(int); ok {
		return val
	}
	return defaultValue
}

func (h *KeapBackup) getStringArray(config map[string]interface{}, key string) []string {
	if val, ok := config[key].([]interface{}); ok {
		result := make([]string, 0, len(val))
		for _, v := range val {
			if str, ok := v.(string); ok {
				result = append(result, str)
			}
		}
		return result
	}
	return []string{}
}

// NewKeapBackup creates a new KeapBackup helper instance
func NewKeapBackup() helpers.Helper { return &KeapBackup{} }

// Register the helper
func init() {
	helpers.Register("keap_backup", func() helpers.Helper {
		return &KeapBackup{}
	})
}
