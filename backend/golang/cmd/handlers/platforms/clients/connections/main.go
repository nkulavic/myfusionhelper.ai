package connections

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	ddbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/google/uuid"
	mfhconfig "github.com/myfusionhelper/api/internal/config"
	authMiddleware "github.com/myfusionhelper/api/internal/middleware/auth"
	apitypes "github.com/myfusionhelper/api/internal/types"
)

var (
	connectionsTable = os.Getenv("CONNECTIONS_TABLE")
	authsTable       = os.Getenv("PLATFORM_CONNECTION_AUTHS_TABLE")
	platformsTable   = os.Getenv("PLATFORMS_TABLE")
	oauthStatesTable = os.Getenv("OAUTH_STATES_TABLE")
)

// ConnectionRequest represents the request body for creating/updating a connection
type ConnectionRequest struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	AuthType    string                 `json:"auth_type"`
	Credentials map[string]interface{} `json:"credentials"`
}

// HandleWithAuth handles platform connections CRUD and protected OAuth endpoints
func HandleWithAuth(ctx context.Context, event events.APIGatewayV2HTTPRequest, authCtx *apitypes.AuthContext) (events.APIGatewayV2HTTPResponse, error) {
	path := event.RequestContext.HTTP.Path
	method := event.RequestContext.HTTP.Method

	if !authCtx.Permissions.CanManageConnections {
		return authMiddleware.CreateErrorResponse(403, "Permission denied"), nil
	}

	// OAuth start
	if strings.Contains(path, "/oauth/start") && method == "POST" {
		return oauthStart(ctx, event, authCtx)
	}

	// Connection test
	if strings.HasSuffix(path, "/test") && method == "POST" {
		return testConnection(ctx, event, authCtx)
	}

	// List ALL connections (no platform filter)
	if path == "/platform-connections" && method == "GET" {
		return listAllConnections(ctx, event, authCtx)
	}

	// CRUD operations
	switch method {
	case "GET":
		if event.PathParameters["connection_id"] != "" {
			return getConnection(ctx, event, authCtx)
		}
		return listConnections(ctx, event, authCtx)
	case "POST":
		return createConnection(ctx, event, authCtx)
	case "PUT":
		return updateConnection(ctx, event, authCtx)
	case "DELETE":
		return deleteConnection(ctx, event, authCtx)
	default:
		return authMiddleware.CreateErrorResponse(405, "Method not allowed"), nil
	}
}

// HandlePublic handles OAuth callback (public endpoint, no auth)
func HandlePublic(ctx context.Context, event events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	if event.RequestContext.HTTP.Path == "/platforms/oauth/callback" {
		return oauthCallback(ctx, event)
	}
	return authMiddleware.CreateErrorResponse(404, "Not Found"), nil
}

// ========== LIST OPERATIONS ==========

func listAllConnections(ctx context.Context, event events.APIGatewayV2HTTPRequest, authCtx *apitypes.AuthContext) (events.APIGatewayV2HTTPResponse, error) {
	log.Printf("List all connections for account: %s", authCtx.AccountID)

	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(os.Getenv("COGNITO_REGION")))
	if err != nil {
		return authMiddleware.CreateErrorResponse(500, "Internal server error"), nil
	}
	db := dynamodb.NewFromConfig(cfg)

	result, err := db.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(connectionsTable),
		IndexName:              aws.String("AccountIdIndex"),
		KeyConditionExpression: aws.String("account_id = :account_id"),
		ExpressionAttributeValues: map[string]ddbtypes.AttributeValue{
			":account_id": &ddbtypes.AttributeValueMemberS{Value: authCtx.AccountID},
		},
	})
	if err != nil {
		log.Printf("Failed to query connections: %v", err)
		return authMiddleware.CreateErrorResponse(500, "Failed to list connections"), nil
	}

	connections := buildConnectionResponseList(result.Items)

	return authMiddleware.CreateSuccessResponse(200, "Connections listed successfully", map[string]interface{}{
		"connections": connections,
		"total":       len(connections),
	}), nil
}

func listConnections(ctx context.Context, event events.APIGatewayV2HTTPRequest, authCtx *apitypes.AuthContext) (events.APIGatewayV2HTTPResponse, error) {
	platformIDOrSlug := event.PathParameters["platform_id"]
	if platformIDOrSlug == "" {
		return authMiddleware.CreateErrorResponse(400, "Platform ID is required"), nil
	}

	log.Printf("List connections for platform %s, account: %s", platformIDOrSlug, authCtx.AccountID)

	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(os.Getenv("COGNITO_REGION")))
	if err != nil {
		return authMiddleware.CreateErrorResponse(500, "Internal server error"), nil
	}
	db := dynamodb.NewFromConfig(cfg)

	// Resolve slug to platform_id if needed
	platformID, err := resolvePlatformID(ctx, db, platformIDOrSlug)
	if err != nil {
		return authMiddleware.CreateErrorResponse(404, "Platform not found"), nil
	}

	result, err := db.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(connectionsTable),
		IndexName:              aws.String("AccountIdIndex"),
		KeyConditionExpression: aws.String("account_id = :account_id"),
		FilterExpression:       aws.String("platform_id = :platform_id"),
		ExpressionAttributeValues: map[string]ddbtypes.AttributeValue{
			":account_id":  &ddbtypes.AttributeValueMemberS{Value: authCtx.AccountID},
			":platform_id": &ddbtypes.AttributeValueMemberS{Value: platformID},
		},
	})
	if err != nil {
		log.Printf("Failed to query connections: %v", err)
		return authMiddleware.CreateErrorResponse(500, "Failed to list connections"), nil
	}

	connections := buildConnectionResponseList(result.Items)

	return authMiddleware.CreateSuccessResponse(200, "Connections listed successfully", map[string]interface{}{
		"connections": connections,
		"platform_id": platformID,
		"total":       len(connections),
	}), nil
}

// ========== CRUD OPERATIONS ==========

func getConnection(ctx context.Context, event events.APIGatewayV2HTTPRequest, authCtx *apitypes.AuthContext) (events.APIGatewayV2HTTPResponse, error) {
	connectionID := event.PathParameters["connection_id"]
	if connectionID == "" {
		return authMiddleware.CreateErrorResponse(400, "Connection ID is required"), nil
	}

	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(os.Getenv("COGNITO_REGION")))
	if err != nil {
		return authMiddleware.CreateErrorResponse(500, "Internal server error"), nil
	}
	db := dynamodb.NewFromConfig(cfg)

	connection, err := getConnectionByID(ctx, db, connectionID)
	if err != nil {
		return authMiddleware.CreateErrorResponse(404, "Connection not found"), nil
	}

	if connection.AccountID != authCtx.AccountID {
		return authMiddleware.CreateErrorResponse(404, "Connection not found"), nil
	}

	responseConn := buildConnectionResponse(connection)
	return authMiddleware.CreateSuccessResponse(200, "Connection retrieved successfully", responseConn), nil
}

func createConnection(ctx context.Context, event events.APIGatewayV2HTTPRequest, authCtx *apitypes.AuthContext) (events.APIGatewayV2HTTPResponse, error) {
	platformID := event.PathParameters["platform_id"]
	if platformID == "" {
		return authMiddleware.CreateErrorResponse(400, "Platform ID is required"), nil
	}

	var req ConnectionRequest
	if err := json.Unmarshal([]byte(event.Body), &req); err != nil {
		return authMiddleware.CreateErrorResponse(400, "Invalid request body"), nil
	}

	if req.Name == "" {
		return authMiddleware.CreateErrorResponse(400, "Connection name is required"), nil
	}
	if req.AuthType == "" {
		return authMiddleware.CreateErrorResponse(400, "Auth type is required (oauth2 or api_key)"), nil
	}

	log.Printf("Create connection for platform %s, account: %s", platformID, authCtx.AccountID)

	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(os.Getenv("COGNITO_REGION")))
	if err != nil {
		return authMiddleware.CreateErrorResponse(500, "Internal server error"), nil
	}
	db := dynamodb.NewFromConfig(cfg)

	// Verify platform exists
	_, err = db.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(platformsTable),
		Key: map[string]ddbtypes.AttributeValue{
			"platform_id": &ddbtypes.AttributeValueMemberS{Value: platformID},
		},
	})
	if err != nil {
		return authMiddleware.CreateErrorResponse(404, "Platform not found"), nil
	}

	// For API key connections, create auth record with credentials
	connectionID := "connection:" + uuid.Must(uuid.NewV7()).String()
	now := time.Now().UTC()
	var authID *string

	if req.AuthType == "api_key" && req.Credentials != nil {
		authRecord := apitypes.PlatformConnectionAuth{
			AuthID:       "auth:" + uuid.Must(uuid.NewV7()).String(),
			ConnectionID: connectionID,
			AccountID:    authCtx.AccountID,
			UserID:       authCtx.UserID,
			PlatformID:   platformID,
			AuthType:     "api_key",
			Status:       "active",
			CreatedAt:    now.Unix(),
			UpdatedAt:    now.Unix(),
		}

		if apiKey, ok := req.Credentials["api_key"].(string); ok {
			authRecord.APIKey = apiKey
		}
		if apiSecret, ok := req.Credentials["api_secret"].(string); ok {
			authRecord.APISecret = apiSecret
		}

		item, err := attributevalue.MarshalMap(authRecord)
		if err != nil {
			return authMiddleware.CreateErrorResponse(500, "Failed to create auth record"), nil
		}

		_, err = db.PutItem(ctx, &dynamodb.PutItemInput{
			TableName: aws.String(authsTable),
			Item:      item,
		})
		if err != nil {
			log.Printf("Failed to store auth record: %v", err)
			return authMiddleware.CreateErrorResponse(500, "Failed to store credentials"), nil
		}

		authID = &authRecord.AuthID
		log.Printf("Created auth record %s for connection %s", authRecord.AuthID, connectionID)
	}

	connection := apitypes.PlatformConnection{
		ConnectionID: connectionID,
		AccountID:    authCtx.AccountID,
		UserID:       authCtx.UserID,
		PlatformID:   platformID,
		Name:         req.Name,
		Status:       "active",
		AuthType:     req.AuthType,
		AuthID:       authID,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	item, err := attributevalue.MarshalMap(connection)
	if err != nil {
		return authMiddleware.CreateErrorResponse(500, "Failed to create connection"), nil
	}

	_, err = db.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(connectionsTable),
		Item:      item,
	})
	if err != nil {
		log.Printf("Failed to store connection: %v", err)
		return authMiddleware.CreateErrorResponse(500, "Failed to create connection"), nil
	}

	return authMiddleware.CreateSuccessResponse(201, "Connection created successfully", map[string]interface{}{
		"connection_id": connectionID,
		"platform_id":   platformID,
		"name":          req.Name,
		"status":        "active",
		"auth_type":     req.AuthType,
		"created_at":    now.Format(time.RFC3339),
	}), nil
}

func updateConnection(ctx context.Context, event events.APIGatewayV2HTTPRequest, authCtx *apitypes.AuthContext) (events.APIGatewayV2HTTPResponse, error) {
	connectionID := event.PathParameters["connection_id"]
	if connectionID == "" {
		return authMiddleware.CreateErrorResponse(400, "Connection ID is required"), nil
	}

	var req ConnectionRequest
	if err := json.Unmarshal([]byte(event.Body), &req); err != nil {
		return authMiddleware.CreateErrorResponse(400, "Invalid request body"), nil
	}

	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(os.Getenv("COGNITO_REGION")))
	if err != nil {
		return authMiddleware.CreateErrorResponse(500, "Internal server error"), nil
	}
	db := dynamodb.NewFromConfig(cfg)

	existing, err := getConnectionByID(ctx, db, connectionID)
	if err != nil {
		return authMiddleware.CreateErrorResponse(404, "Connection not found"), nil
	}

	if existing.AccountID != authCtx.AccountID {
		return authMiddleware.CreateErrorResponse(404, "Connection not found"), nil
	}

	// Build update expression
	updateParts := []string{"updated_at = :updated_at"}
	expressionValues := map[string]ddbtypes.AttributeValue{
		":updated_at": &ddbtypes.AttributeValueMemberS{Value: time.Now().UTC().Format(time.RFC3339)},
	}
	expressionNames := map[string]string{}

	if req.Name != "" {
		updateParts = append(updateParts, "#n = :name")
		expressionValues[":name"] = &ddbtypes.AttributeValueMemberS{Value: req.Name}
		expressionNames["#n"] = "name"
	}

	// Update credentials if provided
	if req.Credentials != nil && existing.AuthID != nil && *existing.AuthID != "" {
		authUpdateParts := []string{"updated_at = :auth_updated_at"}
		authExprValues := map[string]ddbtypes.AttributeValue{
			":auth_updated_at": &ddbtypes.AttributeValueMemberN{Value: fmt.Sprintf("%d", time.Now().Unix())},
		}

		if apiKey, ok := req.Credentials["api_key"].(string); ok {
			authUpdateParts = append(authUpdateParts, "api_key = :api_key")
			authExprValues[":api_key"] = &ddbtypes.AttributeValueMemberS{Value: apiKey}
		}
		if apiSecret, ok := req.Credentials["api_secret"].(string); ok {
			authUpdateParts = append(authUpdateParts, "api_secret = :api_secret")
			authExprValues[":api_secret"] = &ddbtypes.AttributeValueMemberS{Value: apiSecret}
		}

		_, err = db.UpdateItem(ctx, &dynamodb.UpdateItemInput{
			TableName: aws.String(authsTable),
			Key: map[string]ddbtypes.AttributeValue{
				"auth_id": &ddbtypes.AttributeValueMemberS{Value: *existing.AuthID},
			},
			UpdateExpression:          aws.String("SET " + strings.Join(authUpdateParts, ", ")),
			ExpressionAttributeValues: authExprValues,
		})
		if err != nil {
			log.Printf("Failed to update auth record: %v", err)
			return authMiddleware.CreateErrorResponse(500, "Failed to update credentials"), nil
		}
	}

	updateInput := &dynamodb.UpdateItemInput{
		TableName: aws.String(connectionsTable),
		Key: map[string]ddbtypes.AttributeValue{
			"connection_id": &ddbtypes.AttributeValueMemberS{Value: connectionID},
		},
		UpdateExpression:          aws.String("SET " + strings.Join(updateParts, ", ")),
		ExpressionAttributeValues: expressionValues,
	}
	if len(expressionNames) > 0 {
		updateInput.ExpressionAttributeNames = expressionNames
	}

	_, err = db.UpdateItem(ctx, updateInput)
	if err != nil {
		log.Printf("Failed to update connection: %v", err)
		return authMiddleware.CreateErrorResponse(500, "Failed to update connection"), nil
	}

	return authMiddleware.CreateSuccessResponse(200, "Connection updated successfully", map[string]interface{}{
		"connection_id": connectionID,
		"updated_at":    time.Now().UTC().Format(time.RFC3339),
	}), nil
}

func deleteConnection(ctx context.Context, event events.APIGatewayV2HTTPRequest, authCtx *apitypes.AuthContext) (events.APIGatewayV2HTTPResponse, error) {
	connectionID := event.PathParameters["connection_id"]
	if connectionID == "" {
		return authMiddleware.CreateErrorResponse(400, "Connection ID is required"), nil
	}

	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(os.Getenv("COGNITO_REGION")))
	if err != nil {
		return authMiddleware.CreateErrorResponse(500, "Internal server error"), nil
	}
	db := dynamodb.NewFromConfig(cfg)

	existing, err := getConnectionByID(ctx, db, connectionID)
	if err != nil {
		return authMiddleware.CreateErrorResponse(404, "Connection not found"), nil
	}

	if existing.AccountID != authCtx.AccountID {
		return authMiddleware.CreateErrorResponse(404, "Connection not found"), nil
	}

	// Revoke the auth record if it exists
	if existing.AuthID != nil && *existing.AuthID != "" {
		_, err = db.UpdateItem(ctx, &dynamodb.UpdateItemInput{
			TableName: aws.String(authsTable),
			Key: map[string]ddbtypes.AttributeValue{
				"auth_id": &ddbtypes.AttributeValueMemberS{Value: *existing.AuthID},
			},
			UpdateExpression: aws.String("SET #s = :status, revoked_at = :revoked_at"),
			ExpressionAttributeNames: map[string]string{
				"#s": "status",
			},
			ExpressionAttributeValues: map[string]ddbtypes.AttributeValue{
				":status":     &ddbtypes.AttributeValueMemberS{Value: "revoked"},
				":revoked_at": &ddbtypes.AttributeValueMemberN{Value: fmt.Sprintf("%d", time.Now().Unix())},
			},
		})
		if err != nil {
			log.Printf("Failed to revoke auth record %s: %v", *existing.AuthID, err)
		}
	}

	// Delete the connection
	_, err = db.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: aws.String(connectionsTable),
		Key: map[string]ddbtypes.AttributeValue{
			"connection_id": &ddbtypes.AttributeValueMemberS{Value: connectionID},
		},
	})
	if err != nil {
		log.Printf("Failed to delete connection: %v", err)
		return authMiddleware.CreateErrorResponse(500, "Failed to delete connection"), nil
	}

	return authMiddleware.CreateSuccessResponse(200, "Connection deleted successfully", map[string]interface{}{
		"connection_id": connectionID,
		"deleted_at":    time.Now().UTC().Format(time.RFC3339),
	}), nil
}

// ========== CONNECTION TESTING ==========

func testConnection(ctx context.Context, event events.APIGatewayV2HTTPRequest, authCtx *apitypes.AuthContext) (events.APIGatewayV2HTTPResponse, error) {
	connectionID := event.PathParameters["connection_id"]
	if connectionID == "" {
		return authMiddleware.CreateErrorResponse(400, "Connection ID is required"), nil
	}

	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(os.Getenv("COGNITO_REGION")))
	if err != nil {
		return authMiddleware.CreateErrorResponse(500, "Internal server error"), nil
	}
	db := dynamodb.NewFromConfig(cfg)

	connection, err := getConnectionByID(ctx, db, connectionID)
	if err != nil {
		return authMiddleware.CreateErrorResponse(404, "Connection not found"), nil
	}

	if connection.AccountID != authCtx.AccountID {
		return authMiddleware.CreateErrorResponse(404, "Connection not found"), nil
	}

	if connection.AuthID == nil || *connection.AuthID == "" {
		return authMiddleware.CreateErrorResponse(400, "Connection has no auth record"), nil
	}

	// Get auth record
	authResult, err := db.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(authsTable),
		Key: map[string]ddbtypes.AttributeValue{
			"auth_id": &ddbtypes.AttributeValueMemberS{Value: *connection.AuthID},
		},
	})
	if err != nil || authResult.Item == nil {
		return authMiddleware.CreateErrorResponse(500, "Failed to retrieve credentials"), nil
	}

	var auth apitypes.PlatformConnectionAuth
	if err := attributevalue.UnmarshalMap(authResult.Item, &auth); err != nil {
		return authMiddleware.CreateErrorResponse(500, "Failed to parse credentials"), nil
	}

	if auth.Status != "active" {
		return authMiddleware.CreateErrorResponse(400, fmt.Sprintf("Auth record is not active (status: %s)", auth.Status)), nil
	}

	// Get platform config for test endpoint
	platformResult, err := db.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(platformsTable),
		Key: map[string]ddbtypes.AttributeValue{
			"platform_id": &ddbtypes.AttributeValueMemberS{Value: connection.PlatformID},
		},
	})
	if err != nil || platformResult.Item == nil {
		return authMiddleware.CreateErrorResponse(500, "Failed to retrieve platform config"), nil
	}

	var platform apitypes.Platform
	if err := attributevalue.UnmarshalMap(platformResult.Item, &platform); err != nil {
		return authMiddleware.CreateErrorResponse(500, "Failed to parse platform config"), nil
	}

	// Test the connection using platform test endpoint
	testResult, err := executeConnectionTest(ctx, &platform, &auth)
	if err != nil {
		log.Printf("Connection test failed: %v", err)
		// Update connection status to error
		updateConnectionStatus(ctx, db, connectionID, "error")
		return authMiddleware.CreateErrorResponse(401, "Connection test failed"), nil
	}

	// Update connection status and last connected
	now := time.Now().UTC()
	_, _ = db.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(connectionsTable),
		Key: map[string]ddbtypes.AttributeValue{
			"connection_id": &ddbtypes.AttributeValueMemberS{Value: connectionID},
		},
		UpdateExpression: aws.String("SET #s = :status, last_connected = :last_connected, updated_at = :updated_at"),
		ExpressionAttributeNames: map[string]string{
			"#s": "status",
		},
		ExpressionAttributeValues: map[string]ddbtypes.AttributeValue{
			":status":         &ddbtypes.AttributeValueMemberS{Value: "active"},
			":last_connected": &ddbtypes.AttributeValueMemberS{Value: now.Format(time.RFC3339)},
			":updated_at":     &ddbtypes.AttributeValueMemberS{Value: now.Format(time.RFC3339)},
		},
	})

	return authMiddleware.CreateSuccessResponse(200, "Connection test successful", map[string]interface{}{
		"connection_id": connectionID,
		"platform_id":   connection.PlatformID,
		"status":        "success",
		"tested_at":     now.Format(time.RFC3339),
		"result":        testResult,
	}), nil
}

func executeConnectionTest(ctx context.Context, platform *apitypes.Platform, auth *apitypes.PlatformConnectionAuth) (map[string]interface{}, error) {
	testEndpoint := platform.APIConfig.TestEndpoint
	if testEndpoint == "" {
		testEndpoint = platform.APIConfig.BaseURL
	}
	if testEndpoint == "" {
		return map[string]interface{}{"status": "valid", "message": "No test endpoint configured"}, nil
	}

	req, err := http.NewRequestWithContext(ctx, "GET", testEndpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create test request: %w", err)
	}

	// Set auth header based on type
	switch auth.AuthType {
	case "oauth2":
		req.Header.Set("Authorization", "Bearer "+auth.AccessToken)
	case "api_key":
		req.Header.Set("Authorization", "Bearer "+auth.APIKey)
	}
	req.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("test request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("test failed with status %d: %s", resp.StatusCode, string(body))
	}

	return map[string]interface{}{
		"status":      "valid",
		"platform_id": platform.PlatformID,
		"tested_at":   time.Now().UTC().Format(time.RFC3339),
	}, nil
}

// ========== OAUTH FLOW ==========

func oauthStart(ctx context.Context, event events.APIGatewayV2HTTPRequest, authCtx *apitypes.AuthContext) (events.APIGatewayV2HTTPResponse, error) {
	platformIDOrSlug := event.PathParameters["platform_id"]
	if platformIDOrSlug == "" {
		return authMiddleware.CreateErrorResponse(400, "Platform ID is required"), nil
	}

	log.Printf("Starting OAuth flow for platform %s, user %s", platformIDOrSlug, authCtx.UserID)

	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(os.Getenv("COGNITO_REGION")))
	if err != nil {
		return authMiddleware.CreateErrorResponse(500, "Internal server error"), nil
	}
	db := dynamodb.NewFromConfig(cfg)

	// Resolve platform
	platformID, err := resolvePlatformID(ctx, db, platformIDOrSlug)
	if err != nil {
		return authMiddleware.CreateErrorResponse(404, "Platform not found"), nil
	}

	// Get platform details
	platformResult, err := db.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(platformsTable),
		Key: map[string]ddbtypes.AttributeValue{
			"platform_id": &ddbtypes.AttributeValueMemberS{Value: platformID},
		},
	})
	if err != nil || platformResult.Item == nil {
		return authMiddleware.CreateErrorResponse(404, "Platform not found"), nil
	}

	var platform apitypes.Platform
	if err := attributevalue.UnmarshalMap(platformResult.Item, &platform); err != nil {
		return authMiddleware.CreateErrorResponse(500, "Failed to parse platform"), nil
	}

	if platform.OAuth == nil {
		return authMiddleware.CreateErrorResponse(400, "Platform does not support OAuth"), nil
	}

	// Load OAuth credentials from unified SSM parameter
	oauthConfig, err := mfhconfig.GetPlatformOAuth(ctx, platform.Slug)
	if err != nil {
		log.Printf("Failed to load OAuth credentials for %s: %v", platform.Slug, err)
		return authMiddleware.CreateErrorResponse(500, "OAuth not configured for this platform"), nil
	}
	clientID := oauthConfig.ClientID
	clientSecret := oauthConfig.ClientSecret

	// Parse redirect URLs from request body
	successRedirect := ""
	failureRedirect := ""
	if event.Body != "" {
		var bodyParams struct {
			SuccessRedirect string `json:"success_redirect"`
			FailureRedirect string `json:"failure_redirect"`
		}
		if err := json.Unmarshal([]byte(event.Body), &bodyParams); err == nil {
			successRedirect = bodyParams.SuccessRedirect
			failureRedirect = bodyParams.FailureRedirect
		}
	}

	// Generate state token
	state := "state:" + uuid.Must(uuid.NewV7()).String()
	now := time.Now()

	var metadata map[string]interface{}
	if failureRedirect != "" {
		metadata = map[string]interface{}{
			"failure_redirect": failureRedirect,
		}
	}

	oauthState := apitypes.OAuthState{
		State:       state,
		UserID:      authCtx.UserID,
		AccountID:   authCtx.AccountID,
		PlatformID:  platformID,
		RedirectURI: successRedirect,
		Metadata:    metadata,
		CreatedAt:   now.Unix(),
		ExpiresAt:   now.Add(15 * time.Minute).Unix(),
		TTL:         now.Add(1 * time.Hour).Unix(),
	}

	stateItem, err := attributevalue.MarshalMap(oauthState)
	if err != nil {
		return authMiddleware.CreateErrorResponse(500, "Failed to create OAuth state"), nil
	}

	_, err = db.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(oauthStatesTable),
		Item:      stateItem,
	})
	if err != nil {
		log.Printf("Failed to store OAuth state: %v", err)
		return authMiddleware.CreateErrorResponse(500, "Failed to initiate OAuth flow"), nil
	}

	// Build authorization URL
	stage := os.Getenv("STAGE")
	if stage == "" {
		stage = "dev"
	}
	redirectURI := os.Getenv("OAUTH_REDIRECT_URI")
	if redirectURI == "" {
		redirectURI = fmt.Sprintf("https://%s.api.myfusionhelper.ai/platforms/oauth/callback", stage)
	}

	params := url.Values{}
	params.Set("client_id", clientID)
	params.Set("redirect_uri", redirectURI)
	params.Set("response_type", "code")
	params.Set("state", state)

	if len(platform.OAuth.Scopes) > 0 {
		params.Set("scope", strings.Join(platform.OAuth.Scopes, " "))
	}

	// Platform-specific parameters
	switch platform.Slug {
	case "keap":
		// Keap-specific params
	case "gohighlevel":
		// GHL-specific params
	}

	authURL := fmt.Sprintf("%s?%s", platform.OAuth.AuthURL, params.Encode())

	_ = clientSecret // Used during token exchange in callback

	return authMiddleware.CreateSuccessResponse(200, "OAuth flow initiated", map[string]interface{}{
		"authorization_url": authURL,
		"state":             state,
		"platform_id":       platformID,
		"platform_name":     platform.Name,
		"expires_in":        900,
	}), nil
}

func oauthCallback(ctx context.Context, event events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	queryParams := event.QueryStringParameters
	if queryParams == nil {
		return redirectWithError("Missing callback parameters", ""), nil
	}

	code := queryParams["code"]
	state := queryParams["state"]
	errorParam := queryParams["error"]

	if errorParam != "" {
		errorDesc := queryParams["error_description"]
		log.Printf("OAuth error: %s - %s", errorParam, errorDesc)
		return redirectWithError(fmt.Sprintf("OAuth failed: %s", errorParam), ""), nil
	}

	if code == "" || state == "" {
		return redirectWithError("Missing code or state parameter", ""), nil
	}

	log.Printf("Processing OAuth callback with state: %s", state)

	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(os.Getenv("COGNITO_REGION")))
	if err != nil {
		return redirectWithError("Internal server error", ""), nil
	}
	db := dynamodb.NewFromConfig(cfg)

	// Retrieve and validate state
	stateResult, err := db.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(oauthStatesTable),
		Key: map[string]ddbtypes.AttributeValue{
			"state": &ddbtypes.AttributeValueMemberS{Value: state},
		},
	})
	if err != nil || stateResult.Item == nil {
		return redirectWithError("Invalid or expired OAuth state", ""), nil
	}

	var oauthState apitypes.OAuthState
	if err := attributevalue.UnmarshalMap(stateResult.Item, &oauthState); err != nil {
		return redirectWithError("Invalid OAuth state", ""), nil
	}

	failureRedirect := getFailureRedirect(&oauthState)

	// Verify state hasn't expired
	if time.Now().Unix() > oauthState.ExpiresAt {
		return redirectWithError("OAuth session expired", failureRedirect), nil
	}

	// Get platform details
	platformResult, err := db.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(platformsTable),
		Key: map[string]ddbtypes.AttributeValue{
			"platform_id": &ddbtypes.AttributeValueMemberS{Value: oauthState.PlatformID},
		},
	})
	if err != nil || platformResult.Item == nil {
		return redirectWithError("Platform not found", failureRedirect), nil
	}

	var platform apitypes.Platform
	if err := attributevalue.UnmarshalMap(platformResult.Item, &platform); err != nil {
		return redirectWithError("Failed to parse platform", failureRedirect), nil
	}

	// Load OAuth credentials from unified SSM parameter
	oauthConfig, err := mfhconfig.GetPlatformOAuth(ctx, platform.Slug)
	if err != nil {
		return redirectWithError("OAuth configuration error", failureRedirect), nil
	}
	clientID := oauthConfig.ClientID
	clientSecret := oauthConfig.ClientSecret

	// Exchange code for tokens
	stage := os.Getenv("STAGE")
	if stage == "" {
		stage = "dev"
	}
	redirectURI := os.Getenv("OAUTH_REDIRECT_URI")
	if redirectURI == "" {
		redirectURI = fmt.Sprintf("https://%s.api.myfusionhelper.ai/platforms/oauth/callback", stage)
	}

	tokens, err := exchangeCodeForTokens(ctx, platform.OAuth.TokenURL, clientID, clientSecret, code, redirectURI)
	if err != nil {
		log.Printf("Failed to exchange code for tokens: %v", err)
		return redirectWithError("Failed to complete OAuth", failureRedirect), nil
	}

	// Fetch user info from provider
	var externalUserID, externalUserEmail string
	if platform.OAuth.UserInfoURL != "" {
		userInfo, err := fetchUserInfo(ctx, platform.OAuth.UserInfoURL, tokens.AccessToken)
		if err != nil {
			log.Printf("Warning: Failed to fetch user info: %v", err)
		} else if userInfo != nil {
			externalUserID = userInfo.ID
			externalUserEmail = userInfo.Email
		}
	}

	// Check for existing connection
	existingConn, _ := findExistingConnection(ctx, db, oauthState.AccountID, oauthState.PlatformID, externalUserID)

	now := time.Now().UTC()
	var connectionID string

	if existingConn != nil {
		// Update existing connection
		connectionID = existingConn.ConnectionID
		log.Printf("Found existing connection %s, updating auth", connectionID)

		if existingConn.AuthID != nil && *existingConn.AuthID != "" {
			var expiresAt int64
			if tokens.ExpiresAt != nil {
				expiresAt = tokens.ExpiresAt.Unix()
			}

			_, err = db.UpdateItem(ctx, &dynamodb.UpdateItemInput{
				TableName: aws.String(authsTable),
				Key: map[string]ddbtypes.AttributeValue{
					"auth_id": &ddbtypes.AttributeValueMemberS{Value: *existingConn.AuthID},
				},
				UpdateExpression: aws.String("SET access_token = :at, refresh_token = :rt, expires_at = :ea, #s = :status, updated_at = :ua"),
				ExpressionAttributeNames: map[string]string{
					"#s": "status",
				},
				ExpressionAttributeValues: map[string]ddbtypes.AttributeValue{
					":at":     &ddbtypes.AttributeValueMemberS{Value: tokens.AccessToken},
					":rt":     &ddbtypes.AttributeValueMemberS{Value: tokens.RefreshToken},
					":ea":     &ddbtypes.AttributeValueMemberN{Value: fmt.Sprintf("%d", expiresAt)},
					":status": &ddbtypes.AttributeValueMemberS{Value: "active"},
					":ua":     &ddbtypes.AttributeValueMemberN{Value: fmt.Sprintf("%d", now.Unix())},
				},
			})
			if err != nil {
				return redirectWithError("Failed to update credentials", failureRedirect), nil
			}
		}

		// Update connection metadata
		updateParts := []string{"#s = :status", "updated_at = :ua"}
		updateValues := map[string]ddbtypes.AttributeValue{
			":status": &ddbtypes.AttributeValueMemberS{Value: "active"},
			":ua":     &ddbtypes.AttributeValueMemberS{Value: now.Format(time.RFC3339)},
		}
		updateNames := map[string]string{"#s": "status"}

		if externalUserID != "" {
			updateParts = append(updateParts, "external_user_id = :euid")
			updateValues[":euid"] = &ddbtypes.AttributeValueMemberS{Value: externalUserID}
		}
		if externalUserEmail != "" {
			updateParts = append(updateParts, "external_user_email = :eue")
			updateValues[":eue"] = &ddbtypes.AttributeValueMemberS{Value: externalUserEmail}
		}

		_, _ = db.UpdateItem(ctx, &dynamodb.UpdateItemInput{
			TableName: aws.String(connectionsTable),
			Key: map[string]ddbtypes.AttributeValue{
				"connection_id": &ddbtypes.AttributeValueMemberS{Value: connectionID},
			},
			UpdateExpression:          aws.String("SET " + strings.Join(updateParts, ", ")),
			ExpressionAttributeValues: updateValues,
			ExpressionAttributeNames:  updateNames,
		})

	} else {
		// Create new connection and auth record
		connectionID = "connection:" + uuid.Must(uuid.NewV7()).String()
		authID := "auth:" + uuid.Must(uuid.NewV7()).String()

		var expiresAt int64
		if tokens.ExpiresAt != nil {
			expiresAt = tokens.ExpiresAt.Unix()
		}

		authRecord := apitypes.PlatformConnectionAuth{
			AuthID:       authID,
			ConnectionID: connectionID,
			AccountID:    oauthState.AccountID,
			UserID:       oauthState.UserID,
			PlatformID:   oauthState.PlatformID,
			AuthType:     "oauth2",
			AccessToken:  tokens.AccessToken,
			RefreshToken: tokens.RefreshToken,
			ExpiresAt:    expiresAt,
			Status:       "active",
			CreatedAt:    now.Unix(),
			UpdatedAt:    now.Unix(),
		}

		authItem, _ := attributevalue.MarshalMap(authRecord)
		_, err = db.PutItem(ctx, &dynamodb.PutItemInput{
			TableName: aws.String(authsTable),
			Item:      authItem,
		})
		if err != nil {
			return redirectWithError("Failed to save credentials", failureRedirect), nil
		}

		connectionName := fmt.Sprintf("%s Connection", platform.Name)

		connection := apitypes.PlatformConnection{
			ConnectionID:      connectionID,
			AccountID:         oauthState.AccountID,
			UserID:            oauthState.UserID,
			PlatformID:        oauthState.PlatformID,
			ExternalUserID:    externalUserID,
			ExternalUserEmail: externalUserEmail,
			Name:              connectionName,
			Status:            "active",
			AuthType:          "oauth2",
			AuthID:            &authID,
			CreatedAt:         now,
			UpdatedAt:         now,
		}

		connItem, _ := attributevalue.MarshalMap(connection)
		_, err = db.PutItem(ctx, &dynamodb.PutItemInput{
			TableName: aws.String(connectionsTable),
			Item:      connItem,
		})
		if err != nil {
			// Clean up auth record
			_, _ = db.DeleteItem(ctx, &dynamodb.DeleteItemInput{
				TableName: aws.String(authsTable),
				Key: map[string]ddbtypes.AttributeValue{
					"auth_id": &ddbtypes.AttributeValueMemberS{Value: authID},
				},
			})
			return redirectWithError("Failed to create connection", failureRedirect), nil
		}

		log.Printf("Created OAuth connection %s for platform %s", connectionID, platform.Name)
	}

	// Delete the state token (one-time use)
	_, _ = db.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: aws.String(oauthStatesTable),
		Key: map[string]ddbtypes.AttributeValue{
			"state": &ddbtypes.AttributeValueMemberS{Value: state},
		},
	})

	return redirectWithSuccess(connectionID, oauthState.RedirectURI), nil
}

// ========== OAUTH HELPERS ==========

// OAuthTokenResponse represents the response from token exchange
type OAuthTokenResponse struct {
	AccessToken  string     `json:"access_token"`
	RefreshToken string     `json:"refresh_token,omitempty"`
	ExpiresIn    int        `json:"expires_in"`
	TokenType    string     `json:"token_type"`
	Scope        string     `json:"scope,omitempty"`
	ExpiresAt    *time.Time `json:"-"`
}

func exchangeCodeForTokens(ctx context.Context, tokenURL, clientID, clientSecret, code, redirectURI string) (*OAuthTokenResponse, error) {
	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)
	data.Set("redirect_uri", redirectURI)
	data.Set("client_id", clientID)
	data.Set("client_secret", clientSecret)

	req, err := http.NewRequestWithContext(ctx, "POST", tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create token request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("token exchange failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read token response: %w", err)
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("token exchange failed: %s - %s", resp.Status, string(body))
	}

	var tokenResp OAuthTokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, fmt.Errorf("failed to parse token response: %w", err)
	}

	if tokenResp.ExpiresIn > 0 {
		expiresAt := time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)
		tokenResp.ExpiresAt = &expiresAt
	}

	return &tokenResp, nil
}

// OAuthUserInfo represents the user info from an OAuth provider
type OAuthUserInfo struct {
	ID    string `json:"id"`
	Sub   string `json:"sub"`
	Email string `json:"email"`
}

func fetchUserInfo(ctx context.Context, userInfoURL, accessToken string) (*OAuthUserInfo, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", userInfoURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("user info fetch returned %d", resp.StatusCode)
	}

	var raw map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, err
	}

	userInfo := &OAuthUserInfo{}

	// Try common field names for user ID
	for _, field := range []string{"global_user_id", "sub", "id", "user_id"} {
		if id, ok := raw[field].(string); ok && id != "" {
			userInfo.ID = id
			break
		} else if id, ok := raw[field].(float64); ok {
			userInfo.ID = fmt.Sprintf("%.0f", id)
			break
		}
	}

	if email, ok := raw["email"].(string); ok {
		userInfo.Email = email
	}

	return userInfo, nil
}

// loadOAuthCredentials is deprecated. Use config.GetPlatformOAuth() instead.
// Kept as fallback for backward compatibility.
func loadOAuthCredentials(ctx context.Context, ssmClient *ssm.Client, stage, slug string) (clientID, clientSecret string, err error) {
	oauthConfig, err := mfhconfig.GetPlatformOAuth(ctx, slug)
	if err != nil {
		return "", "", err
	}
	return oauthConfig.ClientID, oauthConfig.ClientSecret, nil
}

func findExistingConnection(ctx context.Context, db *dynamodb.Client, accountID, platformID, externalUserID string) (*apitypes.PlatformConnection, error) {
	if externalUserID == "" {
		return nil, nil
	}

	result, err := db.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(connectionsTable),
		IndexName:              aws.String("AccountIdIndex"),
		KeyConditionExpression: aws.String("account_id = :account_id"),
		FilterExpression:       aws.String("platform_id = :platform_id AND external_user_id = :external_user_id"),
		ExpressionAttributeValues: map[string]ddbtypes.AttributeValue{
			":account_id":       &ddbtypes.AttributeValueMemberS{Value: accountID},
			":platform_id":      &ddbtypes.AttributeValueMemberS{Value: platformID},
			":external_user_id": &ddbtypes.AttributeValueMemberS{Value: externalUserID},
		},
	})
	if err != nil {
		return nil, err
	}

	if len(result.Items) == 0 {
		return nil, nil
	}

	var conn apitypes.PlatformConnection
	if err := attributevalue.UnmarshalMap(result.Items[0], &conn); err != nil {
		return nil, err
	}
	return &conn, nil
}

// ========== REDIRECT HELPERS ==========

func redirectWithError(errorMsg, customRedirectURL string) events.APIGatewayV2HTTPResponse {
	var errorURL string

	if customRedirectURL != "" {
		separator := "?"
		if strings.Contains(customRedirectURL, "?") {
			separator = "&"
		}
		errorURL = fmt.Sprintf("%s%soauth=error&error=%s", customRedirectURL, separator, url.QueryEscape(errorMsg))
	} else {
		frontendErrorURL := os.Getenv("FRONTEND_ERROR_URL")
		if frontendErrorURL == "" {
			frontendErrorURL = "https://app.myfusionhelper.ai/connections?oauth=error"
		}
		errorURL = fmt.Sprintf("%s&error=%s", frontendErrorURL, url.QueryEscape(errorMsg))
	}

	return events.APIGatewayV2HTTPResponse{
		StatusCode: 302,
		Headers: map[string]string{
			"Location":                    errorURL,
			"Access-Control-Allow-Origin": "*",
		},
		Body: "",
	}
}

func redirectWithSuccess(connectionID, customRedirectURI string) events.APIGatewayV2HTTPResponse {
	var successURL string

	if customRedirectURI != "" {
		separator := "?"
		if strings.Contains(customRedirectURI, "?") {
			separator = "&"
		}
		successURL = fmt.Sprintf("%s%soauth=success&connection_id=%s", customRedirectURI, separator, connectionID)
	} else {
		frontendSuccessURL := os.Getenv("FRONTEND_SUCCESS_URL")
		if frontendSuccessURL == "" {
			frontendSuccessURL = "https://app.myfusionhelper.ai/connections?oauth=success"
		}
		successURL = fmt.Sprintf("%s&connection_id=%s", frontendSuccessURL, connectionID)
	}

	return events.APIGatewayV2HTTPResponse{
		StatusCode: 302,
		Headers: map[string]string{
			"Location":                    successURL,
			"Access-Control-Allow-Origin": "*",
		},
		Body: "",
	}
}

func getFailureRedirect(oauthState *apitypes.OAuthState) string {
	if oauthState == nil || oauthState.Metadata == nil {
		return ""
	}
	if failureRedirect, ok := oauthState.Metadata["failure_redirect"].(string); ok {
		return failureRedirect
	}
	return ""
}

// ========== UTILITY HELPERS ==========

func getConnectionByID(ctx context.Context, db *dynamodb.Client, connectionID string) (*apitypes.PlatformConnection, error) {
	result, err := db.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(connectionsTable),
		Key: map[string]ddbtypes.AttributeValue{
			"connection_id": &ddbtypes.AttributeValueMemberS{Value: connectionID},
		},
	})
	if err != nil {
		return nil, err
	}
	if result.Item == nil {
		return nil, fmt.Errorf("connection not found")
	}

	var conn apitypes.PlatformConnection
	if err := attributevalue.UnmarshalMap(result.Item, &conn); err != nil {
		return nil, err
	}
	return &conn, nil
}

func resolvePlatformID(ctx context.Context, db *dynamodb.Client, platformIDOrSlug string) (string, error) {
	// If it already looks like a platform ID, validate and return
	if strings.HasPrefix(platformIDOrSlug, "platform:") {
		result, err := db.GetItem(ctx, &dynamodb.GetItemInput{
			TableName: aws.String(platformsTable),
			Key: map[string]ddbtypes.AttributeValue{
				"platform_id": &ddbtypes.AttributeValueMemberS{Value: platformIDOrSlug},
			},
		})
		if err != nil || result.Item == nil {
			return "", fmt.Errorf("platform not found: %s", platformIDOrSlug)
		}
		return platformIDOrSlug, nil
	}

	// Try slug lookup via GSI
	slugResult, err := db.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(platformsTable),
		IndexName:              aws.String("SlugIndex"),
		KeyConditionExpression: aws.String("slug = :slug"),
		ExpressionAttributeValues: map[string]ddbtypes.AttributeValue{
			":slug": &ddbtypes.AttributeValueMemberS{Value: platformIDOrSlug},
		},
	})
	if err != nil {
		return "", err
	}
	if len(slugResult.Items) == 0 {
		return "", fmt.Errorf("platform not found for slug: %s", platformIDOrSlug)
	}

	var platform struct {
		PlatformID string `dynamodbav:"platform_id"`
	}
	if err := attributevalue.UnmarshalMap(slugResult.Items[0], &platform); err != nil {
		return "", err
	}
	return platform.PlatformID, nil
}

func updateConnectionStatus(ctx context.Context, db *dynamodb.Client, connectionID, status string) {
	_, _ = db.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(connectionsTable),
		Key: map[string]ddbtypes.AttributeValue{
			"connection_id": &ddbtypes.AttributeValueMemberS{Value: connectionID},
		},
		UpdateExpression: aws.String("SET #s = :status, updated_at = :updated_at"),
		ExpressionAttributeNames: map[string]string{
			"#s": "status",
		},
		ExpressionAttributeValues: map[string]ddbtypes.AttributeValue{
			":status":     &ddbtypes.AttributeValueMemberS{Value: status},
			":updated_at": &ddbtypes.AttributeValueMemberS{Value: time.Now().UTC().Format(time.RFC3339)},
		},
	})
}

func buildConnectionResponseList(items []map[string]ddbtypes.AttributeValue) []map[string]interface{} {
	connections := make([]map[string]interface{}, 0, len(items))
	for _, item := range items {
		var conn apitypes.PlatformConnection
		if err := attributevalue.UnmarshalMap(item, &conn); err != nil {
			continue
		}
		connections = append(connections, buildConnectionResponse(&conn))
	}
	return connections
}

func buildConnectionResponse(conn *apitypes.PlatformConnection) map[string]interface{} {
	resp := map[string]interface{}{
		"connection_id": conn.ConnectionID,
		"platform_id":   conn.PlatformID,
		"account_id":    conn.AccountID,
		"name":          conn.Name,
		"status":        conn.Status,
		"auth_type":     conn.AuthType,
		"created_at":    conn.CreatedAt.Format(time.RFC3339),
		"updated_at":    conn.UpdatedAt.Format(time.RFC3339),
	}

	if conn.ExternalUserID != "" {
		resp["external_user_id"] = conn.ExternalUserID
	}
	if conn.ExternalUserEmail != "" {
		resp["external_user_email"] = conn.ExternalUserEmail
	}
	if conn.ExternalAppID != "" {
		resp["external_app_id"] = conn.ExternalAppID
	}
	if conn.ExternalAppName != "" {
		resp["external_app_name"] = conn.ExternalAppName
	}
	if conn.AuthID != nil && *conn.AuthID != "" {
		resp["auth_id"] = *conn.AuthID
	}
	if conn.LastConnected != nil {
		resp["last_connected"] = conn.LastConnected.Format(time.RFC3339)
	}
	if conn.ExpiresAt != nil {
		resp["expires_at"] = conn.ExpiresAt.Format(time.RFC3339)
	}

	return resp
}
