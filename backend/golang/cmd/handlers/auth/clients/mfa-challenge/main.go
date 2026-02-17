package mfachallenge

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

type MfaChallengeRequest struct {
	Session       string `json:"session"`
	Code          string `json:"code"`
	ChallengeName string `json:"challenge_name"`
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
	cognitoClientID   = os.Getenv("COGNITO_CLIENT_ID")
	usersTable        = os.Getenv("USERS_TABLE")
	accountsTable     = os.Getenv("ACCOUNTS_TABLE")
	userAccountsTable = os.Getenv("USER_ACCOUNTS_TABLE")
)

// Handle processes MFA challenge responses (public, no JWT required)
func Handle(ctx context.Context, event events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	log.Printf("MFA Challenge handler called")

	body := apiutil.GetBody(event)
	if body == "" {
		return authMiddleware.CreateErrorResponse(400, "Request body is required"), nil
	}

	var req MfaChallengeRequest
	if err := json.Unmarshal([]byte(body), &req); err != nil {
		return authMiddleware.CreateErrorResponse(400, "Invalid JSON format"), nil
	}

	if req.Session == "" || req.Code == "" || req.ChallengeName == "" {
		return authMiddleware.CreateErrorResponse(400, "session, code, and challenge_name are required"), nil
	}

	// Validate challenge name
	var challengeName cognitotypes.ChallengeNameType
	var codeKey string
	switch req.ChallengeName {
	case "SMS_MFA":
		challengeName = cognitotypes.ChallengeNameTypeSmsMfa
		codeKey = "SMS_MFA_CODE"
	case "SOFTWARE_TOKEN_MFA":
		challengeName = cognitotypes.ChallengeNameTypeSoftwareTokenMfa
		codeKey = "SOFTWARE_TOKEN_MFA_CODE"
	default:
		return authMiddleware.CreateErrorResponse(400, "Invalid challenge type"), nil
	}

	region := os.Getenv("COGNITO_REGION")
	if region == "" {
		region = "us-west-2"
	}

	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		log.Printf("Failed to load AWS config: %v", err)
		return authMiddleware.CreateErrorResponse(500, "MFA verification failed"), nil
	}

	cognitoClient := cognitoidentityprovider.NewFromConfig(cfg)

	result, err := cognitoClient.RespondToAuthChallenge(ctx, &cognitoidentityprovider.RespondToAuthChallengeInput{
		ClientId:      aws.String(cognitoClientID),
		ChallengeName: challengeName,
		Session:       aws.String(req.Session),
		ChallengeResponses: map[string]string{
			codeKey:    req.Code,
			"USERNAME": "", // Cognito fills this from the session
		},
	})
	if err != nil {
		log.Printf("MFA challenge failed: %v", err)
		var codeMismatch *cognitotypes.CodeMismatchException
		var expiredCode *cognitotypes.ExpiredCodeException
		switch {
		case errors.As(err, &codeMismatch):
			return authMiddleware.CreateErrorResponse(401, "Invalid verification code"), nil
		case errors.As(err, &expiredCode):
			return authMiddleware.CreateErrorResponse(401, "Verification code expired. Please log in again."), nil
		default:
			return authMiddleware.CreateErrorResponse(401, "MFA verification failed"), nil
		}
	}

	if result.AuthenticationResult == nil {
		return authMiddleware.CreateErrorResponse(500, "MFA verification failed"), nil
	}

	accessToken := *result.AuthenticationResult.AccessToken

	// Extract user ID and fetch user/account (same as login handler)
	dbClient := dynamodb.NewFromConfig(cfg)
	userID, err := extractUserID(accessToken)
	if err != nil {
		log.Printf("Failed to extract user ID: %v", err)
		return authMiddleware.CreateErrorResponse(500, "MFA verification failed"), nil
	}

	user, err := getUser(ctx, dbClient, userID)
	if err != nil {
		log.Printf("Failed to fetch user: %v", err)
		return authMiddleware.CreateErrorResponse(500, "MFA verification failed"), nil
	}

	if user.CurrentAccountID == "" {
		if err := setFirstAccount(ctx, dbClient, userID); err != nil {
			log.Printf("Warning: could not set current account: %v", err)
		}
		user, _ = getUser(ctx, dbClient, userID)
	}

	var account *Account
	if user != nil && user.CurrentAccountID != "" {
		account, err = getAccount(ctx, dbClient, user.CurrentAccountID)
		if err != nil {
			log.Printf("Warning: could not fetch account: %v", err)
		}
	}

	return authMiddleware.CreateSuccessResponse(200, "MFA verification successful", map[string]interface{}{
		"token":         accessToken,
		"refresh_token": *result.AuthenticationResult.RefreshToken,
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
