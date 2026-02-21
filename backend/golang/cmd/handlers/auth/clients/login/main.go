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
	"github.com/myfusionhelper/api/internal/apiutil"
	authMiddleware "github.com/myfusionhelper/api/internal/middleware/auth"
)

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type User struct {
	UserID             string `json:"user_id" dynamodbav:"user_id"`
	CognitoUserID      string `json:"cognito_user_id" dynamodbav:"cognito_user_id"`
	Email              string `json:"email" dynamodbav:"email"`
	Name               string `json:"name" dynamodbav:"name"`
	PhoneNumber        string `json:"phone_number,omitempty" dynamodbav:"phone_number,omitempty"`
	Company            string `json:"company,omitempty" dynamodbav:"company,omitempty"`
	Status             string `json:"status" dynamodbav:"status"`
	CurrentAccountID   string `json:"current_account_id" dynamodbav:"current_account_id"`
	OnboardingComplete bool   `json:"onboarding_complete" dynamodbav:"onboarding_complete"`
	EmailVerified      bool   `json:"email_verified" dynamodbav:"email_verified"`
	CreatedAt          string `json:"created_at" dynamodbav:"created_at"`
	UpdatedAt          string `json:"updated_at" dynamodbav:"updated_at"`
}

type Account struct {
	AccountID       string `json:"account_id" dynamodbav:"account_id"`
	OwnerUserID     string `json:"owner_user_id" dynamodbav:"owner_user_id"`
	CreatedByUserID string `json:"created_by_user_id" dynamodbav:"created_by_user_id"`
	Name            string `json:"name" dynamodbav:"name"`
	Company         string `json:"company" dynamodbav:"company"`
	Plan            string `json:"plan" dynamodbav:"plan"`
	Status          string `json:"status" dynamodbav:"status"`
	CreatedAt       string `json:"created_at" dynamodbav:"created_at"`
	UpdatedAt       string `json:"updated_at" dynamodbav:"updated_at"`
}

var (
	cognitoUserPoolID = os.Getenv("COGNITO_USER_POOL_ID")
	cognitoClientID   = os.Getenv("COGNITO_CLIENT_ID")
	usersTable        = os.Getenv("USERS_TABLE")
	accountsTable     = os.Getenv("ACCOUNTS_TABLE")
	userAccountsTable = os.Getenv("USER_ACCOUNTS_TABLE")
)

// Handle is the login handler (public, no auth required)
func Handle(ctx context.Context, event events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	log.Printf("Login handler called")

	if cognitoUserPoolID == "" || cognitoClientID == "" {
		log.Printf("ERROR: Missing Cognito configuration")
		return authMiddleware.CreateErrorResponse(500, "Authentication service not configured"), nil
	}

	body := apiutil.GetBody(event)
	if body == "" {
		return authMiddleware.CreateErrorResponse(400, "Request body is required"), nil
	}

	var req LoginRequest
	if err := json.Unmarshal([]byte(body), &req); err != nil {
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
		log.Printf("MFA challenge required: %s", authResult.ChallengeName)
		return authMiddleware.CreateSuccessResponse(200, "MFA required", map[string]interface{}{
			"mfa_required":   true,
			"challenge_name": string(authResult.ChallengeName),
			"session":        aws.ToString(authResult.Session),
			"username":       req.Email,
		}), nil
	}

	if authResult.AuthenticationResult == nil {
		return authMiddleware.CreateErrorResponse(500, "Authentication failed"), nil
	}

	accessToken := *authResult.AuthenticationResult.AccessToken

	// Extract user ID from the access token
	dbClient := dynamodb.NewFromConfig(cfg)
	userID, err := extractUserID(accessToken)
	if err != nil {
		log.Printf("Failed to extract user ID from token: %v", err)
		return authMiddleware.CreateErrorResponse(500, "Login failed"), nil
	}

	// Fetch user record from DynamoDB
	user, err := getUser(ctx, dbClient, userID)
	if err != nil {
		log.Printf("Failed to fetch user: %v", err)
		return authMiddleware.CreateErrorResponse(500, "Login failed"), nil
	}

	// Ensure user has a current account
	if user.CurrentAccountID == "" {
		if err := setFirstAccount(ctx, dbClient, userID); err != nil {
			log.Printf("Warning: could not set current account: %v", err)
		}
		// Re-fetch user to get updated current_account_id
		user, _ = getUser(ctx, dbClient, userID)
	}

	// Fetch account record
	var account *Account
	if user != nil && user.CurrentAccountID != "" {
		account, err = getAccount(ctx, dbClient, user.CurrentAccountID)
		if err != nil {
			log.Printf("Warning: could not fetch account: %v", err)
		}
	}

	return authMiddleware.CreateSuccessResponse(200, "Login successful", map[string]interface{}{
		"token":         accessToken,
		"refresh_token": *authResult.AuthenticationResult.RefreshToken,
		"user":          user,
		"account":       account,
	}), nil
}

func extractUserID(accessToken string) (string, error) {
	token, _, err := jwt.NewParser().ParseUnverified(accessToken, jwt.MapClaims{})
	if err != nil {
		return "", err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", errors.New("invalid token claims")
	}

	sub, ok := claims["sub"].(string)
	if !ok || sub == "" {
		return "", errors.New("missing sub claim")
	}

	return "user:" + sub, nil
}

func getUser(ctx context.Context, dbClient *dynamodb.Client, userID string) (*User, error) {
	if usersTable == "" {
		return nil, errors.New("users table not configured")
	}

	result, err := dbClient.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(usersTable),
		Key: map[string]ddbtypes.AttributeValue{
			"user_id": &ddbtypes.AttributeValueMemberS{Value: userID},
		},
	})
	if err != nil {
		return nil, err
	}
	if result.Item == nil {
		return nil, errors.New("user not found")
	}

	var user User
	if err := attributevalue.UnmarshalMap(result.Item, &user); err != nil {
		return nil, err
	}
	return &user, nil
}

func getAccount(ctx context.Context, dbClient *dynamodb.Client, accountID string) (*Account, error) {
	if accountsTable == "" {
		return nil, errors.New("accounts table not configured")
	}

	result, err := dbClient.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(accountsTable),
		Key: map[string]ddbtypes.AttributeValue{
			"account_id": &ddbtypes.AttributeValueMemberS{Value: accountID},
		},
	})
	if err != nil {
		return nil, err
	}
	if result.Item == nil {
		return nil, errors.New("account not found")
	}

	var account Account
	if err := attributevalue.UnmarshalMap(result.Item, &account); err != nil {
		return nil, err
	}
	return &account, nil
}

func setFirstAccount(ctx context.Context, dbClient *dynamodb.Client, userID string) error {
	if userAccountsTable == "" || usersTable == "" {
		return nil
	}

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
