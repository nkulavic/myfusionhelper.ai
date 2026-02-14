package database

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	ddbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/myfusionhelper/api/internal/types"
)

// OAuthStatesRepository provides access to the oauth-states DynamoDB table.
type OAuthStatesRepository struct {
	client    *dynamodb.Client
	tableName string
}

// NewOAuthStatesRepository creates a new OAuthStatesRepository.
func NewOAuthStatesRepository(client *dynamodb.Client, tableName string) *OAuthStatesRepository {
	return &OAuthStatesRepository{client: client, tableName: tableName}
}

// Create inserts a new OAuth state record. The TTL field on the OAuthState struct
// controls automatic expiration via DynamoDB's TTL feature.
func (r *OAuthStatesRepository) Create(ctx context.Context, state *types.OAuthState) error {
	return putItem(ctx, r.client, r.tableName, state)
}

// GetAndDelete retrieves an OAuth state by its state ID and then deletes it (one-time use).
// Returns (nil, nil) if the state is not found.
func (r *OAuthStatesRepository) GetAndDelete(ctx context.Context, stateID string) (*types.OAuthState, error) {
	// Note: the OAuthState PK attribute in DynamoDB is "state" per the serverless.yml KeySchema,
	// but the struct tag maps State -> "state_id". The DynamoDB table uses "state" as the
	// attribute name in KeySchema. We need to use the DynamoDB attribute name here.
	key := map[string]ddbtypes.AttributeValue{
		"state": stringVal(stateID),
	}

	state, err := getItem[types.OAuthState](ctx, r.client, r.tableName, key)
	if err != nil {
		return nil, fmt.Errorf("get oauth state: %w", err)
	}
	if state == nil {
		return nil, nil
	}

	_, err = r.client.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: &r.tableName,
		Key:       key,
	})
	if err != nil {
		return nil, fmt.Errorf("delete oauth state: %w", err)
	}

	return state, nil
}
