package database

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	ddbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/myfusionhelper/api/internal/types"
)

// EmailLogsRepository provides access to the email_logs DynamoDB table.
type EmailLogsRepository struct {
	client    *dynamodb.Client
	tableName string
}

// NewEmailLogsRepository creates a new EmailLogsRepository.
func NewEmailLogsRepository(client *dynamodb.Client, tableName string) *EmailLogsRepository {
	return &EmailLogsRepository{client: client, tableName: tableName}
}

// GetByID fetches an email log by email_id (primary key).
func (r *EmailLogsRepository) GetByID(ctx context.Context, emailID string) (*types.EmailLog, error) {
	return getItem[types.EmailLog](ctx, r.client, r.tableName, stringKey("email_id", emailID))
}

// GetByAccountID retrieves email logs for an account using the AccountIdIndex GSI.
func (r *EmailLogsRepository) GetByAccountID(ctx context.Context, accountID string, limit int32) ([]types.EmailLog, error) {
	indexName := "AccountIdIndex"
	input := &dynamodb.QueryInput{
		TableName:              &r.tableName,
		IndexName:              &indexName,
		KeyConditionExpression: aws.String("account_id = :account_id"),
		ExpressionAttributeValues: map[string]ddbtypes.AttributeValue{
			":account_id": stringVal(accountID),
		},
		ScanIndexForward: aws.Bool(false), // Most recent first
	}

	if limit > 0 {
		input.Limit = aws.Int32(limit)
	}

	return queryIndex[types.EmailLog](ctx, r.client, input)
}

// GetByRecipientEmail retrieves email logs by recipient email using the RecipientEmailIndex GSI.
func (r *EmailLogsRepository) GetByRecipientEmail(ctx context.Context, recipientEmail string, limit int32) ([]types.EmailLog, error) {
	indexName := "RecipientEmailIndex"
	input := &dynamodb.QueryInput{
		TableName:              &r.tableName,
		IndexName:              &indexName,
		KeyConditionExpression: aws.String("recipient_email = :email"),
		ExpressionAttributeValues: map[string]ddbtypes.AttributeValue{
			":email": stringVal(recipientEmail),
		},
		ScanIndexForward: aws.Bool(false), // Most recent first
	}

	if limit > 0 {
		input.Limit = aws.Int32(limit)
	}

	return queryIndex[types.EmailLog](ctx, r.client, input)
}

// GetByAccountIDAndStatus retrieves email logs filtered by status for an account.
func (r *EmailLogsRepository) GetByAccountIDAndStatus(ctx context.Context, accountID, status string, limit int32) ([]types.EmailLog, error) {
	indexName := "AccountIdIndex"
	input := &dynamodb.QueryInput{
		TableName:              &r.tableName,
		IndexName:              &indexName,
		KeyConditionExpression: aws.String("account_id = :account_id"),
		FilterExpression:       aws.String("#status = :status"),
		ExpressionAttributeNames: map[string]string{
			"#status": "status",
		},
		ExpressionAttributeValues: map[string]ddbtypes.AttributeValue{
			":account_id": stringVal(accountID),
			":status":     stringVal(status),
		},
		ScanIndexForward: aws.Bool(false), // Most recent first
	}

	if limit > 0 {
		input.Limit = aws.Int32(limit)
	}

	return queryIndex[types.EmailLog](ctx, r.client, input)
}

// Create inserts a new email log.
func (r *EmailLogsRepository) Create(ctx context.Context, log *types.EmailLog) error {
	return putItem(ctx, r.client, r.tableName, log)
}

// Update performs a full replace of the email log record.
func (r *EmailLogsRepository) Update(ctx context.Context, log *types.EmailLog) error {
	return putItem(ctx, r.client, r.tableName, log)
}

// UpdateStatus updates only the status field of an email log.
func (r *EmailLogsRepository) UpdateStatus(ctx context.Context, emailID, status string) error {
	_, err := r.client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: &r.tableName,
		Key:       stringKey("email_id", emailID),
		UpdateExpression: aws.String("SET #status = :status"),
		ExpressionAttributeNames: map[string]string{
			"#status": "status",
		},
		ExpressionAttributeValues: map[string]ddbtypes.AttributeValue{
			":status": stringVal(status),
		},
	})
	return err
}
