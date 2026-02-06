package register

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	cognitotypes "github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider/types"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/google/uuid"
	authMiddleware "github.com/myfusionhelper/api/internal/middleware/auth"
	"github.com/myfusionhelper/api/internal/notifications"
)

type RegisterRequest struct {
	Email       string `json:"email"`
	Password    string `json:"password"`
	Name        string `json:"name"`
	PhoneNumber string `json:"phone_number"`
	Company     string `json:"company"`
}

type User struct {
	UserID           string `json:"user_id" dynamodbav:"user_id"`
	CognitoUserID    string `json:"cognito_user_id" dynamodbav:"cognito_user_id"`
	Email            string `json:"email" dynamodbav:"email"`
	Name             string `json:"name" dynamodbav:"name"`
	PhoneNumber      string `json:"phone_number,omitempty" dynamodbav:"phone_number,omitempty"`
	Company          string `json:"company,omitempty" dynamodbav:"company,omitempty"`
	Status           string `json:"status" dynamodbav:"status"`
	CurrentAccountID string `json:"current_account_id" dynamodbav:"current_account_id"`
	CreatedAt        string `json:"created_at" dynamodbav:"created_at"`
	UpdatedAt        string `json:"updated_at" dynamodbav:"updated_at"`
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
	usersTable        = os.Getenv("USERS_TABLE")
	accountsTable     = os.Getenv("ACCOUNTS_TABLE")
	userAccountsTable = os.Getenv("USER_ACCOUNTS_TABLE")
	userPoolID        = os.Getenv("COGNITO_USER_POOL_ID")
	cognitoClientID   = os.Getenv("COGNITO_CLIENT_ID")
)

// Handle is the registration handler (public, no auth required)
func Handle(ctx context.Context, event events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	log.Printf("Register handler called")

	var req RegisterRequest
	if err := json.Unmarshal([]byte(event.Body), &req); err != nil {
		return authMiddleware.CreateErrorResponse(400, "Invalid request format"), nil
	}

	// Validate required fields
	if req.Email == "" {
		return authMiddleware.CreateErrorResponse(400, "Email is required"), nil
	}
	if req.Password == "" {
		return authMiddleware.CreateErrorResponse(400, "Password is required"), nil
	}
	if req.Name == "" {
		return authMiddleware.CreateErrorResponse(400, "Name is required"), nil
	}
	if !strings.Contains(req.Email, "@") {
		return authMiddleware.CreateErrorResponse(400, "Invalid email format"), nil
	}
	if len(req.Password) < 8 {
		return authMiddleware.CreateErrorResponse(400, "Password must be at least 8 characters"), nil
	}
	if req.PhoneNumber != "" && !strings.HasPrefix(req.PhoneNumber, "+") {
		return authMiddleware.CreateErrorResponse(400, "Phone number must start with +"), nil
	}

	region := os.Getenv("COGNITO_REGION")
	if region == "" {
		region = "us-west-2"
	}

	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		log.Printf("Failed to load AWS config: %v", err)
		return authMiddleware.CreateErrorResponse(500, "Registration failed"), nil
	}

	cognitoClient := cognitoidentityprovider.NewFromConfig(cfg)
	dynamoClient := dynamodb.NewFromConfig(cfg)

	// Build user attributes
	userAttributes := []cognitotypes.AttributeType{
		{Name: aws.String("email"), Value: aws.String(strings.ToLower(req.Email))},
		{Name: aws.String("email_verified"), Value: aws.String("true")},
		{Name: aws.String("name"), Value: aws.String(req.Name)},
	}
	if req.PhoneNumber != "" {
		userAttributes = append(userAttributes,
			cognitotypes.AttributeType{Name: aws.String("phone_number"), Value: aws.String(req.PhoneNumber)},
			cognitotypes.AttributeType{Name: aws.String("phone_number_verified"), Value: aws.String("true")},
		)
	}

	// Create user in Cognito
	createUserResult, err := cognitoClient.AdminCreateUser(ctx, &cognitoidentityprovider.AdminCreateUserInput{
		UserPoolId:     aws.String(userPoolID),
		Username:       aws.String(strings.ToLower(req.Email)),
		UserAttributes: userAttributes,
		MessageAction:  cognitotypes.MessageActionTypeSuppress,
	})
	if err != nil {
		log.Printf("Failed to create user in Cognito: %v", err)
		var usernameExists *cognitotypes.UsernameExistsException
		if errors.As(err, &usernameExists) {
			return authMiddleware.CreateErrorResponse(409, "Email already registered"), nil
		}
		return authMiddleware.CreateErrorResponse(400, "Failed to create user"), nil
	}

	// Set user password
	_, err = cognitoClient.AdminSetUserPassword(ctx, &cognitoidentityprovider.AdminSetUserPasswordInput{
		UserPoolId: aws.String(userPoolID),
		Username:   aws.String(strings.ToLower(req.Email)),
		Password:   aws.String(req.Password),
		Permanent:  true,
	})
	if err != nil {
		log.Printf("Failed to set user password: %v", err)
		return authMiddleware.CreateErrorResponse(500, "Failed to set user password"), nil
	}

	now := time.Now().UTC().Format(time.RFC3339)
	userID := "user:" + *createUserResult.User.Username
	accountID := "account:" + uuid.New().String()

	companyName := req.Company
	if companyName == "" {
		companyName = req.Name
	}

	// Create account
	account := Account{
		AccountID:       accountID,
		OwnerUserID:     userID,
		CreatedByUserID: userID,
		Name:            companyName,
		Company:         companyName,
		Plan:            "free",
		Status:          "active",
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	accountItem, err := attributevalue.MarshalMap(account)
	if err != nil {
		log.Printf("Failed to marshal account: %v", err)
		return authMiddleware.CreateErrorResponse(500, "Failed to create account"), nil
	}

	_, err = dynamoClient.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(accountsTable),
		Item:      accountItem,
	})
	if err != nil {
		log.Printf("Failed to store account: %v", err)
		return authMiddleware.CreateErrorResponse(500, "Failed to create account"), nil
	}

	// Create user record
	user := User{
		UserID:           userID,
		CognitoUserID:    *createUserResult.User.Username,
		Email:            strings.ToLower(req.Email),
		Name:             req.Name,
		PhoneNumber:      req.PhoneNumber,
		Company:          req.Company,
		Status:           "active",
		CurrentAccountID: accountID,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	userItem, err := attributevalue.MarshalMap(user)
	if err != nil {
		log.Printf("Failed to marshal user: %v", err)
		return authMiddleware.CreateErrorResponse(500, "Failed to create user"), nil
	}

	_, err = dynamoClient.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(usersTable),
		Item:      userItem,
	})
	if err != nil {
		log.Printf("Failed to store user: %v", err)
		return authMiddleware.CreateErrorResponse(500, "Failed to create user"), nil
	}

	// Create user-account relationship
	userAccount := map[string]interface{}{
		"user_id":    userID,
		"account_id": accountID,
		"role":       "Owner",
		"status":     "Active",
		"permissions": map[string]interface{}{
			"can_manage_helpers":     true,
			"can_execute_helpers":    true,
			"can_manage_connections": true,
			"can_manage_team":        true,
			"can_manage_billing":     true,
			"can_view_analytics":     true,
			"can_manage_api_keys":    true,
		},
		"linked_at":  now,
		"updated_at": now,
	}

	userAccountItem, err := attributevalue.MarshalMap(userAccount)
	if err != nil {
		log.Printf("Failed to marshal user-account: %v", err)
		return authMiddleware.CreateErrorResponse(500, "Failed to create user-account relationship"), nil
	}

	_, err = dynamoClient.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(userAccountsTable),
		Item:      userAccountItem,
	})
	if err != nil {
		log.Printf("Failed to store user-account: %v", err)
		return authMiddleware.CreateErrorResponse(500, "Failed to create user-account relationship"), nil
	}

	// Authenticate the new user to get tokens
	authResult, err := cognitoClient.InitiateAuth(ctx, &cognitoidentityprovider.InitiateAuthInput{
		AuthFlow: cognitotypes.AuthFlowTypeUserPasswordAuth,
		ClientId: aws.String(cognitoClientID),
		AuthParameters: map[string]string{
			"USERNAME": strings.ToLower(req.Email),
			"PASSWORD": req.Password,
		},
	})
	if err != nil {
		log.Printf("Failed to authenticate new user: %v", err)
		return authMiddleware.CreateErrorResponse(500, "Account created but login failed"), nil
	}

	if authResult.AuthenticationResult == nil {
		return authMiddleware.CreateErrorResponse(500, "Account created but login failed"), nil
	}

	log.Printf("Registration successful for user: %s", req.Email)

	// Send welcome email asynchronously
	go func() {
		notifSvc, err := notifications.New(ctx)
		if err != nil {
			log.Printf("Failed to create notification service for welcome email: %v", err)
			return
		}
		if err := notifSvc.SendWelcomeEmail(ctx, req.Name, req.Email); err != nil {
			log.Printf("Failed to send welcome email: %v", err)
		}
	}()

	return authMiddleware.CreateSuccessResponse(200, "Registration successful", map[string]interface{}{
		"token":         *authResult.AuthenticationResult.AccessToken,
		"refresh_token": *authResult.AuthenticationResult.RefreshToken,
		"user":          user,
		"account":       account,
	}), nil
}
