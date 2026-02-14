package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	ddbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/golang-jwt/jwt/v5"
	apitypes "github.com/myfusionhelper/api/internal/types"
)

// AuthHandlerFunc is the standard signature for authenticated handlers
type AuthHandlerFunc func(ctx context.Context, event events.APIGatewayV2HTTPRequest, authCtx *apitypes.AuthContext) (events.APIGatewayV2HTTPResponse, error)

// AuthMiddleware handles JWT validation and user context loading
type AuthMiddleware struct {
	db *dynamodb.Client
}

// NewAuthMiddleware creates a new auth middleware instance
func NewAuthMiddleware(ctx context.Context) (*AuthMiddleware, error) {
	region := os.Getenv("COGNITO_REGION")
	if region == "" {
		region = "us-west-2"
	}

	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %v", err)
	}

	return &AuthMiddleware{db: dynamodb.NewFromConfig(cfg)}, nil
}

// extractSubFromJWT extracts the 'sub' claim from JWT Bearer token or V2 authorizer
func extractSubFromJWT(event events.APIGatewayV2HTTPRequest) (string, error) {
	// First try to extract from V2 authorizer context
	if event.RequestContext.Authorizer != nil && event.RequestContext.Authorizer.JWT != nil {
		claims := event.RequestContext.Authorizer.JWT.Claims
		if sub, ok := claims["sub"]; ok {
			return sub, nil
		}
	}

	// Fallback to Bearer token extraction
	authHeader := event.Headers["Authorization"]
	if authHeader == "" {
		authHeader = event.Headers["authorization"]
	}
	if authHeader == "" {
		return "", fmt.Errorf("missing Authorization header")
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	if tokenString == authHeader {
		return "", fmt.Errorf("invalid Bearer token format")
	}

	token, err := jwt.ParseWithClaims(tokenString, jwt.MapClaims{}, func(token *jwt.Token) (interface{}, error) {
		return nil, nil
	}, jwt.WithoutClaimsValidation())
	if err != nil {
		return "", fmt.Errorf("failed to parse JWT: %v", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", fmt.Errorf("invalid JWT claims")
	}

	sub, ok := claims["sub"].(string)
	if !ok {
		return "", fmt.Errorf("missing sub claim")
	}

	return sub, nil
}

// WithAuth wraps a handler with authentication
func (auth *AuthMiddleware) WithAuth(handler AuthHandlerFunc) func(ctx context.Context, event events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	return func(ctx context.Context, event events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
		sub, err := extractSubFromJWT(event)
		if err != nil {
			log.Printf("JWT extraction failed: %v", err)
			return CreateErrorResponse(401, "Authentication required"), nil
		}

		userID := "user:" + sub

		// Get user's current account ID from users table
		usersTable := os.Getenv("USERS_TABLE")
		userResult, err := auth.db.GetItem(ctx, &dynamodb.GetItemInput{
			TableName: aws.String(usersTable),
			Key: map[string]ddbtypes.AttributeValue{
				"user_id": &ddbtypes.AttributeValueMemberS{Value: userID},
			},
		})
		if err != nil || userResult.Item == nil {
			log.Printf("Failed to get user %s: %v", userID, err)
			return CreateErrorResponse(403, "User not found"), nil
		}

		var user struct {
			CurrentAccountID string `dynamodbav:"current_account_id"`
			Email            string `dynamodbav:"email"`
		}
		if err := attributevalue.UnmarshalMap(userResult.Item, &user); err != nil {
			log.Printf("Failed to unmarshal user: %v", err)
			return CreateErrorResponse(500, "Internal server error"), nil
		}

		if user.CurrentAccountID == "" {
			return CreateErrorResponse(403, "No current account set for user"), nil
		}

		// Get user-account relationship for permissions
		userAccountsTable := os.Getenv("USER_ACCOUNTS_TABLE")
		uaResult, err := auth.db.GetItem(ctx, &dynamodb.GetItemInput{
			TableName: aws.String(userAccountsTable),
			Key: map[string]ddbtypes.AttributeValue{
				"user_id":    &ddbtypes.AttributeValueMemberS{Value: userID},
				"account_id": &ddbtypes.AttributeValueMemberS{Value: user.CurrentAccountID},
			},
		})
		if err != nil || uaResult.Item == nil {
			log.Printf("Failed to get user-account relationship: %v", err)
			return CreateErrorResponse(403, "Access denied"), nil
		}

		var userAccount apitypes.UserAccount
		if err := attributevalue.UnmarshalMap(uaResult.Item, &userAccount); err != nil {
			log.Printf("Failed to unmarshal user-account: %v", err)
			return CreateErrorResponse(500, "Internal server error"), nil
		}

		// Get all available accounts
		accountsResult, err := auth.db.Query(ctx, &dynamodb.QueryInput{
			TableName:              aws.String(userAccountsTable),
			KeyConditionExpression: aws.String("user_id = :user_id"),
			ExpressionAttributeValues: map[string]ddbtypes.AttributeValue{
				":user_id": &ddbtypes.AttributeValueMemberS{Value: userID},
			},
		})

		var availableAccounts []apitypes.AccountAccess
		if err == nil && len(accountsResult.Items) > 0 {
			for _, item := range accountsResult.Items {
				var ua apitypes.UserAccount
				if err := attributevalue.UnmarshalMap(item, &ua); err == nil {
					availableAccounts = append(availableAccounts, apitypes.AccountAccess{
						AccountID:   ua.AccountID,
						Role:        ua.Role,
						Permissions: ua.Permissions,
						IsCurrent:   ua.AccountID == user.CurrentAccountID,
					})
				}
			}
		}

		authCtx := &apitypes.AuthContext{
			UserID:            userID,
			AccountID:         user.CurrentAccountID,
			Email:             user.Email,
			Role:              userAccount.Role,
			Permissions:       userAccount.Permissions,
			AvailableAccounts: availableAccounts,
		}

		return handler(ctx, event, authCtx)
	}
}

// CreateSuccessResponse creates a standardized success response
func CreateSuccessResponse(statusCode int, message string, data interface{}) events.APIGatewayV2HTTPResponse {
	responseBody := map[string]interface{}{
		"success": true,
		"message": message,
	}

	if data != nil {
		responseBody["data"] = data
	}

	body, _ := json.Marshal(responseBody)

	return events.APIGatewayV2HTTPResponse{
		StatusCode: statusCode,
		Headers: map[string]string{
			"Content-Type":                 "application/json",
			"Access-Control-Allow-Origin":  "*",
			"Access-Control-Allow-Methods": "GET, POST, PUT, DELETE, OPTIONS",
			"Access-Control-Allow-Headers": "Content-Type, Authorization, X-Account-Context",
		},
		Body: string(body),
	}
}

// CreateErrorResponse creates a standardized error response
func CreateErrorResponse(statusCode int, message string) events.APIGatewayV2HTTPResponse {
	responseBody := map[string]interface{}{
		"success": false,
		"error":   message,
	}

	body, _ := json.Marshal(responseBody)

	return events.APIGatewayV2HTTPResponse{
		StatusCode: statusCode,
		Headers: map[string]string{
			"Content-Type":                 "application/json",
			"Access-Control-Allow-Origin":  "*",
			"Access-Control-Allow-Methods": "GET, POST, PUT, DELETE, OPTIONS",
			"Access-Control-Allow-Headers": "Content-Type, Authorization, X-Account-Context",
		},
		Body: string(body),
	}
}
