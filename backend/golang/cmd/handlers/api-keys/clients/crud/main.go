package crud

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	ddbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"
	authMiddleware "github.com/myfusionhelper/api/internal/middleware/auth"
	apitypes "github.com/myfusionhelper/api/internal/types"
)

var apiKeysTable = os.Getenv("API_KEYS_TABLE")

type CreateAPIKeyRequest struct {
	Name        string   `json:"name"`
	Permissions []string `json:"permissions"`
}

// HandleWithAuth routes to the appropriate operation based on path and method
func HandleWithAuth(ctx context.Context, event events.APIGatewayV2HTTPRequest, authCtx *apitypes.AuthContext) (events.APIGatewayV2HTTPResponse, error) {
	path := event.RequestContext.HTTP.Path
	method := event.RequestContext.HTTP.Method

	switch {
	case path == "/api-keys" && method == "GET":
		return listAPIKeys(ctx, event, authCtx)
	case path == "/api-keys" && method == "POST":
		return createAPIKey(ctx, event, authCtx)
	case strings.HasPrefix(path, "/api-keys/") && method == "DELETE":
		return revokeAPIKey(ctx, event, authCtx)
	default:
		return authMiddleware.CreateErrorResponse(404, "Not Found"), nil
	}
}

func listAPIKeys(ctx context.Context, event events.APIGatewayV2HTTPRequest, authCtx *apitypes.AuthContext) (events.APIGatewayV2HTTPResponse, error) {
	log.Printf("List API keys for account: %s", authCtx.AccountID)

	if !authCtx.Permissions.CanManageAPIKeys {
		return authMiddleware.CreateErrorResponse(403, "Permission denied"), nil
	}

	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(os.Getenv("COGNITO_REGION")))
	if err != nil {
		return authMiddleware.CreateErrorResponse(500, "Internal server error"), nil
	}
	db := dynamodb.NewFromConfig(cfg)

	// Query API keys by account_id using GSI
	result, err := db.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(apiKeysTable),
		IndexName:              aws.String("AccountIdIndex"),
		KeyConditionExpression: aws.String("account_id = :account_id"),
		ExpressionAttributeValues: map[string]ddbtypes.AttributeValue{
			":account_id": &ddbtypes.AttributeValueMemberS{Value: authCtx.AccountID},
		},
	})
	if err != nil {
		log.Printf("Failed to query API keys: %v", err)
		return authMiddleware.CreateErrorResponse(500, "Failed to list API keys"), nil
	}

	var keys []map[string]interface{}
	for _, item := range result.Items {
		var key apitypes.APIKey
		if err := attributevalue.UnmarshalMap(item, &key); err != nil {
			continue
		}
		keys = append(keys, map[string]interface{}{
			"key_id":      key.KeyID,
			"name":        key.Name,
			"key_prefix":  key.KeyPrefix,
			"permissions": key.Permissions,
			"status":      key.Status,
			"last_used_at": key.LastUsedAt,
			"created_at":  key.CreatedAt,
			"expires_at":  key.ExpiresAt,
		})
	}

	return authMiddleware.CreateSuccessResponse(200, "API keys retrieved successfully", map[string]interface{}{
		"api_keys":    keys,
		"total_count": len(keys),
	}), nil
}

func createAPIKey(ctx context.Context, event events.APIGatewayV2HTTPRequest, authCtx *apitypes.AuthContext) (events.APIGatewayV2HTTPResponse, error) {
	log.Printf("Create API key for account: %s", authCtx.AccountID)

	if !authCtx.Permissions.CanManageAPIKeys {
		return authMiddleware.CreateErrorResponse(403, "Permission denied"), nil
	}

	var req CreateAPIKeyRequest
	if err := json.Unmarshal([]byte(event.Body), &req); err != nil {
		return authMiddleware.CreateErrorResponse(400, "Invalid request format"), nil
	}

	if req.Name == "" {
		return authMiddleware.CreateErrorResponse(400, "Name is required"), nil
	}

	// Default permissions
	if len(req.Permissions) == 0 {
		req.Permissions = []string{"execute_helpers"}
	}

	// Generate the raw API key
	rawKey, err := generateAPIKey()
	if err != nil {
		log.Printf("Failed to generate API key: %v", err)
		return authMiddleware.CreateErrorResponse(500, "Failed to create API key"), nil
	}

	// Hash the key for storage
	keyHash := hashAPIKey(rawKey)
	keyPrefix := rawKey[:16] // Show first 16 chars for identification

	now := time.Now().UTC()
	keyID := "apikey:" + uuid.Must(uuid.NewV7()).String()

	apiKey := apitypes.APIKey{
		KeyID:       keyID,
		AccountID:   authCtx.AccountID,
		CreatedBy:   authCtx.UserID,
		Name:        req.Name,
		KeyHash:     keyHash,
		KeyPrefix:   keyPrefix,
		Permissions: req.Permissions,
		Status:      "active",
		CreatedAt:   now,
	}

	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(os.Getenv("COGNITO_REGION")))
	if err != nil {
		return authMiddleware.CreateErrorResponse(500, "Internal server error"), nil
	}
	db := dynamodb.NewFromConfig(cfg)

	item, err := attributevalue.MarshalMap(apiKey)
	if err != nil {
		return authMiddleware.CreateErrorResponse(500, "Failed to create API key"), nil
	}

	_, err = db.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(apiKeysTable),
		Item:      item,
	})
	if err != nil {
		log.Printf("Failed to store API key: %v", err)
		return authMiddleware.CreateErrorResponse(500, "Failed to create API key"), nil
	}

	// Return the raw key ONLY on creation — it's never retrievable again
	return authMiddleware.CreateSuccessResponse(201, "API key created successfully", map[string]interface{}{
		"key_id":      keyID,
		"name":        req.Name,
		"api_key":     rawKey,
		"key_prefix":  keyPrefix,
		"permissions": req.Permissions,
		"created_at":  now,
	}), nil
}

func revokeAPIKey(ctx context.Context, event events.APIGatewayV2HTTPRequest, authCtx *apitypes.AuthContext) (events.APIGatewayV2HTTPResponse, error) {
	keyID := event.PathParameters["key_id"]
	if keyID == "" {
		return authMiddleware.CreateErrorResponse(400, "Key ID is required"), nil
	}

	log.Printf("Revoke API key %s for account: %s", keyID, authCtx.AccountID)

	if !authCtx.Permissions.CanManageAPIKeys {
		return authMiddleware.CreateErrorResponse(403, "Permission denied"), nil
	}

	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(os.Getenv("COGNITO_REGION")))
	if err != nil {
		return authMiddleware.CreateErrorResponse(500, "Internal server error"), nil
	}
	db := dynamodb.NewFromConfig(cfg)

	// Verify the key belongs to this account
	result, err := db.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(apiKeysTable),
		Key: map[string]ddbtypes.AttributeValue{
			"key_id": &ddbtypes.AttributeValueMemberS{Value: keyID},
		},
	})
	if err != nil || result.Item == nil {
		return authMiddleware.CreateErrorResponse(404, "API key not found"), nil
	}

	var existingKey apitypes.APIKey
	if err := attributevalue.UnmarshalMap(result.Item, &existingKey); err != nil {
		return authMiddleware.CreateErrorResponse(500, "Internal server error"), nil
	}

	if existingKey.AccountID != authCtx.AccountID {
		return authMiddleware.CreateErrorResponse(404, "API key not found"), nil
	}

	// Update status to revoked
	_, err = db.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(apiKeysTable),
		Key: map[string]ddbtypes.AttributeValue{
			"key_id": &ddbtypes.AttributeValueMemberS{Value: keyID},
		},
		UpdateExpression: aws.String("SET #s = :status"),
		ExpressionAttributeNames: map[string]string{
			"#s": "status",
		},
		ExpressionAttributeValues: map[string]ddbtypes.AttributeValue{
			":status": &ddbtypes.AttributeValueMemberS{Value: "revoked"},
		},
	})
	if err != nil {
		log.Printf("Failed to revoke API key: %v", err)
		return authMiddleware.CreateErrorResponse(500, "Failed to revoke API key"), nil
	}

	return authMiddleware.CreateSuccessResponse(200, "API key revoked successfully", map[string]interface{}{
		"key_id": keyID,
		"status": "revoked",
	}), nil
}

// generateAPIKey creates a cryptographically secure API key
func generateAPIKey() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return fmt.Sprintf("mfh_live_%s", hex.EncodeToString(bytes)), nil
}

// hashAPIKey creates a SHA-256 hash of the API key
func hashAPIKey(key string) string {
	hash := sha256.Sum256([]byte(key))
	return hex.EncodeToString(hash[:])
}
