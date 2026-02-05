package login

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	cognitotypes "github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider/types"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	ddbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/golang-jwt/jwt/v5"
	authMiddleware "github.com/myfusionhelper/api/internal/middleware/auth"
)

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

var (
	cognitoUserPoolID = os.Getenv("COGNITO_USER_POOL_ID")
	cognitoClientID   = os.Getenv("COGNITO_CLIENT_ID")
	usersTable        = os.Getenv("USERS_TABLE")
	userAccountsTable = os.Getenv("USER_ACCOUNTS_TABLE")
)

// Handle is the login handler (public, no auth required)
func Handle(ctx context.Context, event events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	log.Printf("Login handler called")

	if cognitoUserPoolID == "" || cognitoClientID == "" {
		log.Printf("ERROR: Missing Cognito configuration")
		return authMiddleware.CreateErrorResponse(500, "Authentication service not configured"), nil
	}

	if event.Body == "" {
		return authMiddleware.CreateErrorResponse(400, "Request body is required"), nil
	}

	var req LoginRequest
	if err := json.Unmarshal([]byte(event.Body), &req); err != nil {
		return authMiddleware.CreateErrorResponse(400, "Invalid JSON format"), nil
	}

	if req.Email == "" {
		return authMiddleware.CreateErrorResponse(400, "Email is required"), nil
	}
	if req.Password == "" {
		return authMiddleware.CreateErrorResponse(400, "Password is required"), nil
	}

	region := os.Getenv("COGNITO_REGION")
	if region == "" {
		region = "us-west-2"
	}

	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		log.Printf("Failed to load AWS config: %v", err)
		return authMiddleware.CreateErrorResponse(500, "Login failed"), nil
	}

	cognitoClient := cognitoidentityprovider.NewFromConfig(cfg)

	authResult, err := cognitoClient.InitiateAuth(ctx, &cognitoidentityprovider.InitiateAuthInput{
		AuthFlow: cognitotypes.AuthFlowTypeUserPasswordAuth,
		ClientId: aws.String(cognitoClientID),
		AuthParameters: map[string]string{
			"USERNAME": req.Email,
			"PASSWORD": req.Password,
		},
	})
	if err != nil {
		log.Printf("Authentication failed: %v", err)
		return handleLoginError(err), nil
	}

	if authResult.ChallengeName != "" {
		return authMiddleware.CreateErrorResponse(400, "Additional authentication required"), nil
	}

	if authResult.AuthenticationResult != nil {
		accessToken := *authResult.AuthenticationResult.AccessToken

		// Ensure user has a current account set
		dbClient := dynamodb.NewFromConfig(cfg)
		if err := ensureUserHasCurrentAccount(ctx, dbClient, accessToken); err != nil {
			log.Printf("Warning: could not ensure current account: %v", err)
		}

		return authMiddleware.CreateSuccessResponse(200, "Login successful", map[string]interface{}{
			"access_token":  accessToken,
			"id_token":      *authResult.AuthenticationResult.IdToken,
			"refresh_token": *authResult.AuthenticationResult.RefreshToken,
			"expires_in":    authResult.AuthenticationResult.ExpiresIn,
			"token_type":    *authResult.AuthenticationResult.TokenType,
			"user_email":    req.Email,
		}), nil
	}

	return authMiddleware.CreateErrorResponse(500, "Authentication failed"), nil
}

func handleLoginError(err error) events.APIGatewayV2HTTPResponse {
	if err == nil {
		return authMiddleware.CreateErrorResponse(500, "Unknown login error")
	}

	var notAuth *cognitotypes.NotAuthorizedException
	var userNotFound *cognitotypes.UserNotFoundException
	var userNotConfirmed *cognitotypes.UserNotConfirmedException
	var tooMany *cognitotypes.TooManyRequestsException

	switch {
	case errors.As(err, &notAuth):
		return authMiddleware.CreateErrorResponse(401, "Invalid email or password")
	case errors.As(err, &userNotFound):
		return authMiddleware.CreateErrorResponse(401, "Invalid email or password")
	case errors.As(err, &userNotConfirmed):
		return authMiddleware.CreateErrorResponse(401, "Email not verified")
	case errors.As(err, &tooMany):
		return authMiddleware.CreateErrorResponse(429, "Too many login attempts")
	default:
		return authMiddleware.CreateErrorResponse(500, "Login failed")
	}
}

func ensureUserHasCurrentAccount(ctx context.Context, dbClient *dynamodb.Client, accessToken string) error {
	if usersTable == "" || userAccountsTable == "" {
		return nil
	}

	token, _, err := jwt.NewParser().ParseUnverified(accessToken, jwt.MapClaims{})
	if err != nil {
		return err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil
	}

	sub, ok := claims["sub"].(string)
	if !ok || sub == "" {
		return nil
	}

	userID := "user:" + sub

	userResult, err := dbClient.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(usersTable),
		Key: map[string]ddbtypes.AttributeValue{
			"user_id": &ddbtypes.AttributeValueMemberS{Value: userID},
		},
		ProjectionExpression: aws.String("user_id, current_account_id"),
	})
	if err != nil || userResult.Item == nil {
		return err
	}

	var user struct {
		CurrentAccountID string `dynamodbav:"current_account_id"`
	}
	if err := attributevalue.UnmarshalMap(userResult.Item, &user); err != nil {
		return err
	}

	if user.CurrentAccountID != "" {
		return nil
	}

	// Set first available account as current
	accountsResult, err := dbClient.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(userAccountsTable),
		KeyConditionExpression: aws.String("user_id = :user_id"),
		ExpressionAttributeValues: map[string]ddbtypes.AttributeValue{
			":user_id": &ddbtypes.AttributeValueMemberS{Value: userID},
		},
		Limit: aws.Int32(1),
	})
	if err != nil || len(accountsResult.Items) == 0 {
		return err
	}

	var ua struct {
		AccountID string `dynamodbav:"account_id"`
	}
	if err := attributevalue.UnmarshalMap(accountsResult.Items[0], &ua); err != nil {
		return err
	}

	_, err = dbClient.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(usersTable),
		Key: map[string]ddbtypes.AttributeValue{
			"user_id": &ddbtypes.AttributeValueMemberS{Value: userID},
		},
		UpdateExpression: aws.String("SET current_account_id = :account_id, updated_at = :updated_at"),
		ExpressionAttributeValues: map[string]ddbtypes.AttributeValue{
			":account_id": &ddbtypes.AttributeValueMemberS{Value: ua.AccountID},
			":updated_at": &ddbtypes.AttributeValueMemberS{Value: time.Now().Format(time.RFC3339)},
		},
	})
	return err
}
