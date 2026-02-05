package database

import (
	"context"
	"os"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	ddbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// NewDynamoDBClient creates a new DynamoDB client using the default AWS credential chain.
func NewDynamoDBClient(ctx context.Context) (*dynamodb.Client, error) {
	region := os.Getenv("COGNITO_REGION")
	if region == "" {
		region = "us-west-2"
	}

	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return nil, err
	}

	return dynamodb.NewFromConfig(cfg), nil
}

// TableNames holds all DynamoDB table names read from environment variables.
type TableNames struct {
	Users                    string
	Accounts                 string
	UserAccounts             string
	APIKeys                  string
	Connections              string
	PlatformConnectionAuths  string
	Helpers                  string
	Executions               string
	Platforms                string
	OAuthStates              string
}

// NewTableNames reads table names from environment variables.
func NewTableNames() TableNames {
	return TableNames{
		Users:                   os.Getenv("USERS_TABLE"),
		Accounts:                os.Getenv("ACCOUNTS_TABLE"),
		UserAccounts:            os.Getenv("USER_ACCOUNTS_TABLE"),
		APIKeys:                 os.Getenv("API_KEYS_TABLE"),
		Connections:             os.Getenv("CONNECTIONS_TABLE"),
		PlatformConnectionAuths: os.Getenv("PLATFORM_CONNECTION_AUTHS_TABLE"),
		Helpers:                 os.Getenv("HELPERS_TABLE"),
		Executions:              os.Getenv("EXECUTIONS_TABLE"),
		Platforms:               os.Getenv("PLATFORMS_TABLE"),
		OAuthStates:             os.Getenv("OAUTH_STATES_TABLE"),
	}
}

// getItem fetches a single item by key and unmarshals it into T.
// Returns (nil, nil) if the item is not found.
func getItem[T any](ctx context.Context, client *dynamodb.Client, tableName string, key map[string]ddbtypes.AttributeValue) (*T, error) {
	result, err := client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: &tableName,
		Key:       key,
	})
	if err != nil {
		return nil, err
	}
	if result.Item == nil {
		return nil, nil
	}

	var item T
	if err := attributevalue.UnmarshalMap(result.Item, &item); err != nil {
		return nil, err
	}
	return &item, nil
}

// putItem marshals an item and writes it to the table.
func putItem(ctx context.Context, client *dynamodb.Client, tableName string, item interface{}) error {
	av, err := attributevalue.MarshalMap(item)
	if err != nil {
		return err
	}

	_, err = client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: &tableName,
		Item:      av,
	})
	return err
}

// putItemWithCondition marshals an item and writes it with a condition expression.
func putItemWithCondition(ctx context.Context, client *dynamodb.Client, tableName string, item interface{}, condition string) error {
	av, err := attributevalue.MarshalMap(item)
	if err != nil {
		return err
	}

	_, err = client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName:           &tableName,
		Item:                av,
		ConditionExpression: &condition,
	})
	return err
}

// queryIndex queries a GSI and unmarshals the results into a slice of T.
func queryIndex[T any](ctx context.Context, client *dynamodb.Client, input *dynamodb.QueryInput) ([]T, error) {
	result, err := client.Query(ctx, input)
	if err != nil {
		return nil, err
	}

	items := make([]T, 0, len(result.Items))
	for _, item := range result.Items {
		var t T
		if err := attributevalue.UnmarshalMap(item, &t); err != nil {
			return nil, err
		}
		items = append(items, t)
	}
	return items, nil
}

// querySingleItem queries and returns the first result, or nil if none found.
func querySingleItem[T any](ctx context.Context, client *dynamodb.Client, input *dynamodb.QueryInput) (*T, error) {
	limit := int32(1)
	input.Limit = &limit

	result, err := client.Query(ctx, input)
	if err != nil {
		return nil, err
	}
	if len(result.Items) == 0 {
		return nil, nil
	}

	var item T
	if err := attributevalue.UnmarshalMap(result.Items[0], &item); err != nil {
		return nil, err
	}
	return &item, nil
}

// stringKey builds a single-attribute string key for DynamoDB operations.
func stringKey(name, value string) map[string]ddbtypes.AttributeValue {
	return map[string]ddbtypes.AttributeValue{
		name: &ddbtypes.AttributeValueMemberS{Value: value},
	}
}

// stringVal returns a DynamoDB string attribute value.
func stringVal(value string) ddbtypes.AttributeValue {
	return &ddbtypes.AttributeValueMemberS{Value: value}
}

// numVal returns a DynamoDB number attribute value.
func numVal(value string) ddbtypes.AttributeValue {
	return &ddbtypes.AttributeValueMemberN{Value: value}
}
