package database

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	ddbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/myfusionhelper/api/internal/types"
)

// UserAccountsRepository provides access to the user-accounts DynamoDB table.
type UserAccountsRepository struct {
	client    *dynamodb.Client
	tableName string
}

// NewUserAccountsRepository creates a new UserAccountsRepository.
func NewUserAccountsRepository(client *dynamodb.Client, tableName string) *UserAccountsRepository {
	return &UserAccountsRepository{client: client, tableName: tableName}
}

// Get fetches a user-account relationship by composite key (user_id + account_id).
func (r *UserAccountsRepository) Get(ctx context.Context, userID, accountID string) (*types.UserAccount, error) {
	key := map[string]ddbtypes.AttributeValue{
		"user_id":    stringVal(userID),
		"account_id": stringVal(accountID),
	}
	return getItem[types.UserAccount](ctx, r.client, r.tableName, key)
}

// ListByUser fetches all accounts for a given user by querying the partition key.
func (r *UserAccountsRepository) ListByUser(ctx context.Context, userID string) ([]types.UserAccount, error) {
	return queryIndex[types.UserAccount](ctx, r.client, &dynamodb.QueryInput{
		TableName:              &r.tableName,
		KeyConditionExpression: aws.String("user_id = :user_id"),
		ExpressionAttributeValues: map[string]ddbtypes.AttributeValue{
			":user_id": stringVal(userID),
		},
	})
}

// ListByAccount fetches all users for a given account using the AccountIdIndex GSI.
func (r *UserAccountsRepository) ListByAccount(ctx context.Context, accountID string) ([]types.UserAccount, error) {
	indexName := "AccountIdIndex"
	return queryIndex[types.UserAccount](ctx, r.client, &dynamodb.QueryInput{
		TableName:              &r.tableName,
		IndexName:              &indexName,
		KeyConditionExpression: aws.String("account_id = :account_id"),
		ExpressionAttributeValues: map[string]ddbtypes.AttributeValue{
			":account_id": stringVal(accountID),
		},
	})
}

// Create inserts a new user-account relationship.
func (r *UserAccountsRepository) Create(ctx context.Context, ua *types.UserAccount) error {
	return putItem(ctx, r.client, r.tableName, ua)
}

// Update performs a full replace of the user-account record.
func (r *UserAccountsRepository) Update(ctx context.Context, ua *types.UserAccount) error {
	return putItem(ctx, r.client, r.tableName, ua)
}

// Delete removes a user-account relationship.
func (r *UserAccountsRepository) Delete(ctx context.Context, userID, accountID string) error {
	key := map[string]ddbtypes.AttributeValue{
		"user_id":    stringVal(userID),
		"account_id": stringVal(accountID),
	}
	_, err := r.client.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: &r.tableName,
		Key:       key,
	})
	return err
}
