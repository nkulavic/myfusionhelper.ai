package database

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	ddbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/myfusionhelper/api/internal/types"
)

// AccountsRepository provides access to the accounts DynamoDB table.
type AccountsRepository struct {
	client    *dynamodb.Client
	tableName string
}

// NewAccountsRepository creates a new AccountsRepository.
func NewAccountsRepository(client *dynamodb.Client, tableName string) *AccountsRepository {
	return &AccountsRepository{client: client, tableName: tableName}
}

// GetByID fetches an account by its account_id (primary key).
func (r *AccountsRepository) GetByID(ctx context.Context, accountID string) (*types.Account, error) {
	return getItem[types.Account](ctx, r.client, r.tableName, stringKey("account_id", accountID))
}

// GetByOwner fetches all accounts owned by a user using the OwnerUserIdIndex GSI.
func (r *AccountsRepository) GetByOwner(ctx context.Context, ownerUserID string) ([]types.Account, error) {
	indexName := "OwnerUserIdIndex"
	return queryIndex[types.Account](ctx, r.client, &dynamodb.QueryInput{
		TableName:              &r.tableName,
		IndexName:              &indexName,
		KeyConditionExpression: aws.String("owner_user_id = :owner_user_id"),
		ExpressionAttributeValues: map[string]ddbtypes.AttributeValue{
			":owner_user_id": stringVal(ownerUserID),
		},
	})
}

// Create inserts a new account with a condition that the account_id does not already exist.
func (r *AccountsRepository) Create(ctx context.Context, account *types.Account) error {
	return putItemWithCondition(ctx, r.client, r.tableName, account, "attribute_not_exists(account_id)")
}

// Update performs a full replace of the account record.
func (r *AccountsRepository) Update(ctx context.Context, account *types.Account) error {
	return putItem(ctx, r.client, r.tableName, account)
}
