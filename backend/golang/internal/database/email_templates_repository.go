package database

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	ddbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/myfusionhelper/api/internal/types"
)

// EmailTemplatesRepository provides access to the email_templates DynamoDB table.
type EmailTemplatesRepository struct {
	client    *dynamodb.Client
	tableName string
}

// NewEmailTemplatesRepository creates a new EmailTemplatesRepository.
func NewEmailTemplatesRepository(client *dynamodb.Client, tableName string) *EmailTemplatesRepository {
	return &EmailTemplatesRepository{client: client, tableName: tableName}
}

// GetByID fetches an email template by template_id (primary key).
func (r *EmailTemplatesRepository) GetByID(ctx context.Context, templateID string) (*types.EmailTemplate, error) {
	return getItem[types.EmailTemplate](ctx, r.client, r.tableName, stringKey("template_id", templateID))
}

// GetByAccountID retrieves all templates for an account using the AccountIdIndex GSI.
func (r *EmailTemplatesRepository) GetByAccountID(ctx context.Context, accountID string) ([]types.EmailTemplate, error) {
	indexName := "AccountIdIndex"
	return queryIndex[types.EmailTemplate](ctx, r.client, &dynamodb.QueryInput{
		TableName:              &r.tableName,
		IndexName:              &indexName,
		KeyConditionExpression: aws.String("account_id = :account_id"),
		ExpressionAttributeValues: map[string]ddbtypes.AttributeValue{
			":account_id": stringVal(accountID),
		},
	})
}

// GetSystemTemplates retrieves all system templates (templates with is_system = true).
func (r *EmailTemplatesRepository) GetSystemTemplates(ctx context.Context) ([]*types.EmailTemplate, error) {
	// System templates have account_id as empty string or specific marker
	// Using Scan with filter for is_system = true
	output, err := r.client.Scan(ctx, &dynamodb.ScanInput{
		TableName:        &r.tableName,
		FilterExpression: aws.String("is_system = :is_system AND is_active = :is_active"),
		ExpressionAttributeValues: map[string]ddbtypes.AttributeValue{
			":is_system":  &ddbtypes.AttributeValueMemberBOOL{Value: true},
			":is_active":  &ddbtypes.AttributeValueMemberBOOL{Value: true},
		},
	})
	if err != nil {
		return nil, err
	}

	templates := make([]*types.EmailTemplate, 0, len(output.Items))
	for _, item := range output.Items {
		var tmpl types.EmailTemplate
		if err := attributevalue.UnmarshalMap(item, &tmpl); err != nil {
			continue
		}
		templates = append(templates, &tmpl)
	}

	return templates, nil
}

// GetActiveTemplates retrieves all active templates for an account.
func (r *EmailTemplatesRepository) GetActiveTemplates(ctx context.Context, accountID string) ([]types.EmailTemplate, error) {
	indexName := "AccountIdIndex"
	return queryIndex[types.EmailTemplate](ctx, r.client, &dynamodb.QueryInput{
		TableName:              &r.tableName,
		IndexName:              &indexName,
		KeyConditionExpression: aws.String("account_id = :account_id"),
		FilterExpression:       aws.String("is_active = :is_active"),
		ExpressionAttributeValues: map[string]ddbtypes.AttributeValue{
			":account_id": stringVal(accountID),
			":is_active":  &ddbtypes.AttributeValueMemberBOOL{Value: true},
		},
	})
}

// Create inserts a new email template with a condition that the template_id does not already exist.
func (r *EmailTemplatesRepository) Create(ctx context.Context, template *types.EmailTemplate) error {
	now := time.Now().UTC().Format(time.RFC3339)
	template.CreatedAt = now
	template.UpdatedAt = now
	return putItemWithCondition(ctx, r.client, r.tableName, template, "attribute_not_exists(template_id)")
}

// Update performs a full replace of the template record.
func (r *EmailTemplatesRepository) Update(ctx context.Context, template *types.EmailTemplate) error {
	template.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
	return putItem(ctx, r.client, r.tableName, template)
}

// UpdateActive updates the is_active status of a template (soft delete).
func (r *EmailTemplatesRepository) UpdateActive(ctx context.Context, templateID string, isActive bool) error {
	_, err := r.client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: &r.tableName,
		Key:       stringKey("template_id", templateID),
		UpdateExpression: aws.String("SET is_active = :is_active, updated_at = :updated_at"),
		ExpressionAttributeValues: map[string]ddbtypes.AttributeValue{
			":is_active":   &ddbtypes.AttributeValueMemberBOOL{Value: isActive},
			":updated_at":  stringVal(time.Now().UTC().Format(time.RFC3339)),
		},
	})
	return err
}

// Delete performs a hard delete of a template (use UpdateActive for soft delete).
func (r *EmailTemplatesRepository) Delete(ctx context.Context, templateID string) error {
	_, err := r.client.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: &r.tableName,
		Key:       stringKey("template_id", templateID),
	})
	return err
}
