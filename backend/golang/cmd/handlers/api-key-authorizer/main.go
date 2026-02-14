package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"log"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	ddbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	helperResolve "github.com/myfusionhelper/api/internal/helpers"
	apitypes "github.com/myfusionhelper/api/internal/types"
)

var (
	apiKeysTable  = os.Getenv("API_KEYS_TABLE")
	accountsTable = os.Getenv("ACCOUNTS_TABLE")
	helpersTable  = os.Getenv("HELPERS_TABLE")
)

func main() {
	lambda.Start(handler)
}

func handler(ctx context.Context, event events.APIGatewayV2CustomAuthorizerV2Request) (events.APIGatewayV2CustomAuthorizerSimpleResponse, error) {
	denied := events.APIGatewayV2CustomAuthorizerSimpleResponse{IsAuthorized: false}

	// 1. Extract API key from header or path parameter
	rawKey := extractAPIKey(event)
	if rawKey == "" {
		log.Printf("No API key found in request")
		return denied, nil
	}

	// 2. Hash and look up in DynamoDB
	keyHash := hashAPIKey(rawKey)

	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(os.Getenv("COGNITO_REGION")))
	if err != nil {
		log.Printf("Failed to load AWS config: %v", err)
		return denied, nil
	}
	db := dynamodb.NewFromConfig(cfg)

	// Query KeyHashIndex GSI
	result, err := db.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(apiKeysTable),
		IndexName:              aws.String("KeyHashIndex"),
		KeyConditionExpression: aws.String("key_hash = :kh"),
		ExpressionAttributeValues: map[string]ddbtypes.AttributeValue{
			":kh": &ddbtypes.AttributeValueMemberS{Value: keyHash},
		},
		Limit: aws.Int32(1),
	})
	if err != nil || len(result.Items) == 0 {
		log.Printf("API key not found")
		return denied, nil
	}

	var apiKey apitypes.APIKey
	if err := attributevalue.UnmarshalMap(result.Items[0], &apiKey); err != nil {
		log.Printf("Failed to unmarshal API key: %v", err)
		return denied, nil
	}

	// 3. Validate key status and expiry
	if apiKey.Status != "active" {
		log.Printf("API key %s is not active (status: %s)", apiKey.KeyID, apiKey.Status)
		return denied, nil
	}
	if apiKey.ExpiresAt != nil && apiKey.ExpiresAt.Before(time.Now().UTC()) {
		log.Printf("API key %s has expired", apiKey.KeyID)
		return denied, nil
	}

	// 4. Validate account subscription
	accountResult, err := db.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(accountsTable),
		Key: map[string]ddbtypes.AttributeValue{
			"account_id": &ddbtypes.AttributeValueMemberS{Value: apiKey.AccountID},
		},
	})
	if err != nil || accountResult.Item == nil {
		log.Printf("Account %s not found", apiKey.AccountID)
		return denied, nil
	}

	var account apitypes.Account
	if err := attributevalue.UnmarshalMap(accountResult.Item, &account); err != nil {
		log.Printf("Failed to unmarshal account: %v", err)
		return denied, nil
	}

	if account.Status != "active" {
		log.Printf("Account %s is not active (status: %s)", account.AccountID, account.Status)
		return denied, nil
	}

	// 5. Resolve helper from path and verify ownership
	identifier := extractHelperIdentifier(event)
	if identifier == "" {
		log.Printf("No helper identifier found in path")
		return denied, nil
	}

	helper, err := helperResolve.ResolveHelper(ctx, db, helpersTable, identifier)
	if err != nil {
		log.Printf("Failed to resolve helper %s: %v", identifier, err)
		return denied, nil
	}

	if helper.AccountID != apiKey.AccountID {
		log.Printf("Helper %s does not belong to account %s", helper.HelperID, apiKey.AccountID)
		return denied, nil
	}

	if helper.Status != "active" || !helper.Enabled {
		log.Printf("Helper %s is not active/enabled", helper.HelperID)
		return denied, nil
	}

	// 6. Fire-and-forget: update LastUsedAt on the API key
	go func() {
		bgCtx := context.Background()
		now := time.Now().UTC().Format(time.RFC3339)
		_, _ = db.UpdateItem(bgCtx, &dynamodb.UpdateItemInput{
			TableName: aws.String(apiKeysTable),
			Key: map[string]ddbtypes.AttributeValue{
				"key_id": &ddbtypes.AttributeValueMemberS{Value: apiKey.KeyID},
			},
			UpdateExpression: aws.String("SET last_used_at = :now"),
			ExpressionAttributeValues: map[string]ddbtypes.AttributeValue{
				":now": &ddbtypes.AttributeValueMemberS{Value: now},
			},
		})
	}()

	// 7. Return authorized with context for downstream handler
	return events.APIGatewayV2CustomAuthorizerSimpleResponse{
		IsAuthorized: true,
		Context: map[string]interface{}{
			"accountId":  apiKey.AccountID,
			"apiKeyId":   apiKey.KeyID,
			"helperId":   helper.HelperID,
			"helperType": helper.HelperType,
			"permissions": strings.Join(apiKey.Permissions, ","),
		},
	}, nil
}

// extractAPIKey gets the API key from x-api-key header or {api_key} path param.
func extractAPIKey(event events.APIGatewayV2CustomAuthorizerV2Request) string {
	if key, ok := event.Headers["x-api-key"]; ok && key != "" {
		return key
	}
	if key, ok := event.PathParameters["api_key"]; ok && key != "" {
		return key
	}
	return ""
}

// extractHelperIdentifier gets the {identifier} path param.
func extractHelperIdentifier(event events.APIGatewayV2CustomAuthorizerV2Request) string {
	if id, ok := event.PathParameters["identifier"]; ok && id != "" {
		return id
	}
	return ""
}

func hashAPIKey(key string) string {
	hash := sha256.Sum256([]byte(key))
	return hex.EncodeToString(hash[:])
}
