package database

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	ddbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/myfusionhelper/api/internal/types"
)

// UsersRepository provides access to the users DynamoDB table.
type UsersRepository struct {
	client    *dynamodb.Client
	tableName string
}

// NewUsersRepository creates a new UsersRepository.
func NewUsersRepository(client *dynamodb.Client, tableName string) *UsersRepository {
	return &UsersRepository{client: client, tableName: tableName}
}

// GetByID fetches a user by their user_id (primary key).
func (r *UsersRepository) GetByID(ctx context.Context, userID string) (*types.User, error) {
	return getItem[types.User](ctx, r.client, r.tableName, stringKey("user_id", userID))
}

// GetByEmail fetches a user by email using the EmailIndex GSI.
func (r *UsersRepository) GetByEmail(ctx context.Context, email string) (*types.User, error) {
	indexName := "EmailIndex"
	return querySingleItem[types.User](ctx, r.client, &dynamodb.QueryInput{
		TableName:              &r.tableName,
		IndexName:              &indexName,
		KeyConditionExpression: aws.String("email = :email"),
		ExpressionAttributeValues: map[string]ddbtypes.AttributeValue{
			":email": stringVal(email),
		},
	})
}

// GetByCognitoID fetches a user by cognito_user_id using the CognitoUserIdIndex GSI.
func (r *UsersRepository) GetByCognitoID(ctx context.Context, cognitoID string) (*types.User, error) {
	indexName := "CognitoUserIdIndex"
	return querySingleItem[types.User](ctx, r.client, &dynamodb.QueryInput{
		TableName:              &r.tableName,
		IndexName:              &indexName,
		KeyConditionExpression: aws.String("cognito_user_id = :cognito_user_id"),
		ExpressionAttributeValues: map[string]ddbtypes.AttributeValue{
			":cognito_user_id": stringVal(cognitoID),
		},
	})
}

// Create inserts a new user with a condition that the user_id does not already exist.
func (r *UsersRepository) Create(ctx context.Context, user *types.User) error {
	return putItemWithCondition(ctx, r.client, r.tableName, user, "attribute_not_exists(user_id)")
}

// Update performs a full replace of the user record.
func (r *UsersRepository) Update(ctx context.Context, user *types.User) error {
	return putItem(ctx, r.client, r.tableName, user)
}

// UpdateCurrentAccount updates only the current_account_id and updated_at fields.
func (r *UsersRepository) UpdateCurrentAccount(ctx context.Context, userID, accountID string) error {
	_, err := r.client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: &r.tableName,
		Key:       stringKey("user_id", userID),
		UpdateExpression: aws.String("SET current_account_id = :account_id, updated_at = :updated_at"),
		ExpressionAttributeValues: map[string]ddbtypes.AttributeValue{
			":account_id":  stringVal(accountID),
			":updated_at":  stringVal(time.Now().UTC().Format(time.RFC3339)),
		},
		ConditionExpression: aws.String("attribute_exists(user_id)"),
	})
	if err != nil {
		return fmt.Errorf("update current account: %w", err)
	}
	return nil
}
