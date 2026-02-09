package database

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	ddbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/myfusionhelper/api/internal/types"
)

// EmailVerificationsRepository provides access to the email_verifications DynamoDB table.
type EmailVerificationsRepository struct {
	client    *dynamodb.Client
	tableName string
}

// NewEmailVerificationsRepository creates a new EmailVerificationsRepository.
func NewEmailVerificationsRepository(client *dynamodb.Client, tableName string) *EmailVerificationsRepository {
	return &EmailVerificationsRepository{client: client, tableName: tableName}
}

// GetByID fetches an email verification by verification_id (primary key).
func (r *EmailVerificationsRepository) GetByID(ctx context.Context, verificationID string) (*types.EmailVerification, error) {
	return getItem[types.EmailVerification](ctx, r.client, r.tableName, stringKey("verification_id", verificationID))
}

// GetByEmail retrieves the most recent verification for an email using the EmailIndex GSI.
func (r *EmailVerificationsRepository) GetByEmail(ctx context.Context, email string) (*types.EmailVerification, error) {
	indexName := "EmailIndex"
	return querySingleItem[types.EmailVerification](ctx, r.client, &dynamodb.QueryInput{
		TableName:              &r.tableName,
		IndexName:              &indexName,
		KeyConditionExpression: aws.String("email = :email"),
		ExpressionAttributeValues: map[string]ddbtypes.AttributeValue{
			":email": stringVal(email),
		},
		ScanIndexForward: aws.Bool(false), // Most recent first
		Limit:            aws.Int32(1),
	})
}

// GetPendingByEmail retrieves pending verifications for an email.
func (r *EmailVerificationsRepository) GetPendingByEmail(ctx context.Context, email string) ([]*types.EmailVerification, error) {
	indexName := "EmailIndex"
	return queryIndex[types.EmailVerification](ctx, r.client, &dynamodb.QueryInput{
		TableName:              &r.tableName,
		IndexName:              &indexName,
		KeyConditionExpression: aws.String("email = :email"),
		FilterExpression:       aws.String("#status = :status AND expires_at > :now"),
		ExpressionAttributeNames: map[string]string{
			"#status": "status",
		},
		ExpressionAttributeValues: map[string]ddbtypes.AttributeValue{
			":email":  stringVal(email),
			":status": stringVal("pending"),
			":now":    &ddbtypes.AttributeValueMemberN{Value: aws.String(formatInt64(time.Now().Unix()))},
		},
	})
}

// GetByToken retrieves a verification by token (note: this requires scanning, not efficient).
// For production, consider adding a TokenIndex GSI.
func (r *EmailVerificationsRepository) GetByToken(ctx context.Context, token string) (*types.EmailVerification, error) {
	// Scanning is expensive - in production, add a GSI on token field
	output, err := r.client.Scan(ctx, &dynamodb.ScanInput{
		TableName:        &r.tableName,
		FilterExpression: aws.String("token = :token"),
		ExpressionAttributeValues: map[string]ddbtypes.AttributeValue{
			":token": stringVal(token),
		},
		Limit: aws.Int32(1),
	})
	if err != nil {
		return nil, err
	}

	if len(output.Items) == 0 {
		return nil, nil
	}

	var verification types.EmailVerification
	if err := unmarshalMap(output.Items[0], &verification); err != nil {
		return nil, err
	}

	return &verification, nil
}

// Create inserts a new email verification.
func (r *EmailVerificationsRepository) Create(ctx context.Context, verification *types.EmailVerification) error {
	now := time.Now().UTC().Format(time.RFC3339)
	verification.CreatedAt = now
	verification.Status = "pending"

	// Set TTL to expires_at (already unix timestamp)
	if verification.ExpiresAt == 0 {
		verification.ExpiresAt = time.Now().Add(24 * time.Hour).Unix()
	}

	return putItemWithCondition(ctx, r.client, r.tableName, verification, "attribute_not_exists(verification_id)")
}

// Update performs a full replace of the verification record.
func (r *EmailVerificationsRepository) Update(ctx context.Context, verification *types.EmailVerification) error {
	return putItem(ctx, r.client, r.tableName, verification)
}

// MarkAsVerified updates a verification to verified status.
func (r *EmailVerificationsRepository) MarkAsVerified(ctx context.Context, verificationID string) error {
	_, err := r.client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: &r.tableName,
		Key:       stringKey("verification_id", verificationID),
		UpdateExpression: aws.String("SET #status = :status, verified_at = :verified_at"),
		ExpressionAttributeNames: map[string]string{
			"#status": "status",
		},
		ExpressionAttributeValues: map[string]ddbtypes.AttributeValue{
			":status":      stringVal("verified"),
			":verified_at": stringVal(time.Now().UTC().Format(time.RFC3339)),
		},
	})
	return err
}

// MarkAsExpired updates a verification to expired status.
func (r *EmailVerificationsRepository) MarkAsExpired(ctx context.Context, verificationID string) error {
	_, err := r.client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: &r.tableName,
		Key:       stringKey("verification_id", verificationID),
		UpdateExpression: aws.String("SET #status = :status"),
		ExpressionAttributeNames: map[string]string{
			"#status": "status",
		},
		ExpressionAttributeValues: map[string]ddbtypes.AttributeValue{
			":status": stringVal("expired"),
		},
	})
	return err
}

// Delete removes a verification (hard delete).
func (r *EmailVerificationsRepository) Delete(ctx context.Context, verificationID string) error {
	_, err := r.client.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: &r.tableName,
		Key:       stringKey("verification_id", verificationID),
	})
	return err
}

// Helper function to format int64 as string for DynamoDB N type
func formatInt64(n int64) string {
	return aws.String(string(rune(n)))
}
