package helpers

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	ddbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	apitypes "github.com/myfusionhelper/api/internal/types"
)

// ResolveHelper looks up a helper by any identifier format:
//   - NanoID short key (<=20 chars) → query ShortKeyIndex GSI
//   - UUID without prefix (36 chars) → prepend "helper:", direct GetItem
//   - Full ID with prefix ("helper:...") → direct GetItem
func ResolveHelper(ctx context.Context, db *dynamodb.Client, tableName string, identifier string) (*apitypes.Helper, error) {
	if identifier == "" {
		return nil, fmt.Errorf("identifier is required")
	}

	if len(identifier) <= 20 && !strings.HasPrefix(identifier, "helper:") {
		// NanoID short key — query ShortKeyIndex GSI
		return queryByShortKey(ctx, db, tableName, identifier)
	}

	// Full or prefix-less UUID — normalize to full ID
	helperID := identifier
	if !strings.HasPrefix(identifier, "helper:") {
		helperID = "helper:" + identifier
	}
	return getHelperByID(ctx, db, tableName, helperID)
}

func queryByShortKey(ctx context.Context, db *dynamodb.Client, tableName string, shortKey string) (*apitypes.Helper, error) {
	result, err := db.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(tableName),
		IndexName:              aws.String("ShortKeyIndex"),
		KeyConditionExpression: aws.String("short_key = :sk"),
		ExpressionAttributeValues: map[string]ddbtypes.AttributeValue{
			":sk": &ddbtypes.AttributeValueMemberS{Value: shortKey},
		},
		Limit: aws.Int32(1),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query ShortKeyIndex: %w", err)
	}
	if len(result.Items) == 0 {
		return nil, fmt.Errorf("helper not found")
	}

	var helper apitypes.Helper
	if err := attributevalue.UnmarshalMap(result.Items[0], &helper); err != nil {
		return nil, fmt.Errorf("failed to unmarshal helper: %w", err)
	}
	return &helper, nil
}

func getHelperByID(ctx context.Context, db *dynamodb.Client, tableName string, helperID string) (*apitypes.Helper, error) {
	result, err := db.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(tableName),
		Key: map[string]ddbtypes.AttributeValue{
			"helper_id": &ddbtypes.AttributeValueMemberS{Value: helperID},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get helper: %w", err)
	}
	if result.Item == nil {
		return nil, fmt.Errorf("helper not found")
	}

	var helper apitypes.Helper
	if err := attributevalue.UnmarshalMap(result.Item, &helper); err != nil {
		return nil, fmt.Errorf("failed to unmarshal helper: %w", err)
	}
	return &helper, nil
}
