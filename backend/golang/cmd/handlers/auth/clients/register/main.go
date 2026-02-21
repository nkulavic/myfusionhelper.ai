package register

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"log"
	"math/big"
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
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/google/uuid"
	"github.com/myfusionhelper/api/internal/apiutil"
	"github.com/myfusionhelper/api/internal/billing"
	appConfig "github.com/myfusionhelper/api/internal/config"
	"github.com/myfusionhelper/api/internal/database"
	authMiddleware "github.com/myfusionhelper/api/internal/middleware/auth"
	"github.com/myfusionhelper/api/internal/types"
	stripe "github.com/stripe/stripe-go/v82"
	"github.com/stripe/stripe-go/v82/customer"
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
	EmailVerified    bool   `json:"email_verified" dynamodbav:"email_verified"`
	CreatedAt        string `json:"created_at" dynamodbav:"created_at"`
	UpdatedAt        string `json:"updated_at" dynamodbav:"updated_at"`
}

type AccountSettings struct {
	MaxHelpers     int  `json:"max_helpers" dynamodbav:"max_helpers"`
	MaxConnections int  `json:"max_connections" dynamodbav:"max_connections"`
	MaxAPIKeys     int  `json:"max_api_keys" dynamodbav:"max_api_keys"`
	MaxTeamMembers int  `json:"max_team_members" dynamodbav:"max_team_members"`
	MaxExecutions  int  `json:"max_executions" dynamodbav:"max_executions"`
	WebhooksEnabled bool `json:"webhooks_enabled" dynamodbav:"webhooks_enabled"`
}

type AccountUsage struct {
	Helpers            int `json:"helpers" dynamodbav:"helpers"`
	Connections        int `json:"connections" dynamodbav:"connections"`
	APIKeys            int `json:"api_keys" dynamodbav:"api_keys"`
	TeamMembers        int `json:"team_members" dynamodbav:"team_members"`
	MonthlyExecutions  int `json:"monthly_executions" dynamodbav:"monthly_executions"`
	MonthlyAPIRequests int `json:"monthly_api_requests" dynamodbav:"monthly_api_requests"`
}

type Account struct {
	AccountID        string          `json:"account_id" dynamodbav:"account_id"`
	OwnerUserID      string          `json:"owner_user_id" dynamodbav:"owner_user_id"`
	CreatedByUserID  string          `json:"created_by_user_id" dynamodbav:"created_by_user_id"`
	Name             string          `json:"name" dynamodbav:"name"`
	Company          string          `json:"company" dynamodbav:"company"`
	Plan             string          `json:"plan" dynamodbav:"plan"`
	Status           string          `json:"status" dynamodbav:"status"`
	StripeCustomerID string          `json:"stripe_customer_id,omitempty" dynamodbav:"stripe_customer_id,omitempty"`
	TrialStartedAt   *time.Time      `json:"trial_started_at,omitempty" dynamodbav:"trial_started_at,omitempty"`
	TrialEndsAt      *time.Time      `json:"trial_ends_at,omitempty" dynamodbav:"trial_ends_at,omitempty"`
	TrialExpired     bool            `json:"trial_expired" dynamodbav:"trial_expired"`
	Settings         AccountSettings `json:"settings" dynamodbav:"settings"`
	Usage            AccountUsage    `json:"usage" dynamodbav:"usage"`
	CreatedAt        string          `json:"created_at" dynamodbav:"created_at"`
	UpdatedAt        string          `json:"updated_at" dynamodbav:"updated_at"`
}

var (
	usersTable              = os.Getenv("USERS_TABLE")
	accountsTable           = os.Getenv("ACCOUNTS_TABLE")
	userAccountsTable       = os.Getenv("USER_ACCOUNTS_TABLE")
	emailVerificationsTable = os.Getenv("EMAIL_VERIFICATIONS_TABLE")
	userPoolID              = os.Getenv("COGNITO_USER_POOL_ID")
	cognitoClientID         = os.Getenv("COGNITO_CLIENT_ID")
)

// Handle is the registration handler (public, no auth required)
func Handle(ctx context.Context, event events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	log.Printf("Register handler called")

	var req RegisterRequest
	if err := json.Unmarshal([]byte(apiutil.GetBody(event)), &req); err != nil {
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
		{Name: aws.String("email_verified"), Value: aws.String("false")},
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
	accountID := "account:" + uuid.Must(uuid.NewV7()).String()

	companyName := req.Company
	if companyName == "" {
		companyName = req.Name
	}

	// Create Stripe customer for tracking (no subscription yet)
	stripeCustomerID := ""
	secrets, secretsErr := appConfig.LoadSecrets(ctx)
	if secretsErr != nil {
		log.Printf("Failed to load secrets for Stripe customer creation: %v", secretsErr)
		// Non-fatal -- continue registration without Stripe customer
	} else if secrets.Stripe.SecretKey != "" {
		stripe.Key = secrets.Stripe.SecretKey
		cust, err := customer.New(&stripe.CustomerParams{
			Email: stripe.String(strings.ToLower(req.Email)),
			Name:  stripe.String(companyName),
			Metadata: map[string]string{
				"account_id": accountID,
			},
		})
		if err != nil {
			log.Printf("Failed to create Stripe customer (non-fatal): %v", err)
		} else {
			stripeCustomerID = cust.ID
			log.Printf("Created Stripe customer %s for account %s", stripeCustomerID, accountID)
		}
	}

	// Create account with 14-day free trial (Start-level limits)
	trialPlan := billing.GetPlan("trial")
	trialStart := time.Now().UTC()
	trialEnd := trialStart.Add(14 * 24 * time.Hour)
	account := Account{
		AccountID:        accountID,
		OwnerUserID:      userID,
		CreatedByUserID:  userID,
		Name:             companyName,
		Company:          companyName,
		Plan:             "trial",
		Status:           "active",
		StripeCustomerID: stripeCustomerID,
		TrialStartedAt:   &trialStart,
		TrialEndsAt:      &trialEnd,
		TrialExpired:     false,
		Settings: AccountSettings{
			MaxHelpers:     trialPlan.MaxHelpers,
			MaxConnections: trialPlan.MaxConnections,
			MaxAPIKeys:     trialPlan.MaxAPIKeys,
			MaxTeamMembers: trialPlan.MaxTeamMembers,
			MaxExecutions:  trialPlan.MaxExecutions,
		},
		Usage:     AccountUsage{},
		CreatedAt: now,
		UpdatedAt: now,
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
		EmailVerified:    false,
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

	// Generate email verification code and enqueue notification
	verifyCode, err := generateSecureCode()
	if err != nil {
		log.Printf("Failed to generate verification code: %v", err)
		// Non-fatal — user is created, they can resend later
	} else if emailVerificationsTable != "" {
		verificationsRepo := database.NewEmailVerificationsRepository(dynamoClient, emailVerificationsTable)

		verification := &types.EmailVerification{
			VerificationID: "verify:" + uuid.Must(uuid.NewV7()).String(),
			Email:          strings.ToLower(req.Email),
			Token:          verifyCode,
			ExpiresAt:      time.Now().Add(15 * time.Minute).Unix(),
		}
		if err := verificationsRepo.Create(ctx, verification); err != nil {
			log.Printf("Failed to store email verification code: %v", err)
		} else {
			log.Printf("Created email verification code for %s (verification_id: %s)", req.Email, verification.VerificationID)
		}

		// Enqueue email_verification notification via SQS
		notifQueueURL := os.Getenv("NOTIFICATION_QUEUE_URL")
		if notifQueueURL != "" {
			sqsClient := sqs.NewFromConfig(cfg)
			userName := req.Name
			if userName == "" {
				userName = "there"
			}
			jobJSON, _ := json.Marshal(map[string]interface{}{
				"type":    "email_verification",
				"user_id": userID,
				"data": map[string]interface{}{
					"user_email":  strings.ToLower(req.Email),
					"user_name":   userName,
					"verify_code": verifyCode,
				},
			})
			_, sqsErr := sqsClient.SendMessage(ctx, &sqs.SendMessageInput{
				QueueUrl:       aws.String(notifQueueURL),
				MessageGroupId: aws.String("email-verify"),
				MessageBody:    aws.String(string(jobJSON)),
			})
			if sqsErr != nil {
				log.Printf("Failed to enqueue email verification notification: %v", sqsErr)
			}
		}
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

	// Welcome email is triggered automatically by the DynamoDB stream on the
	// Users table INSERT. The notification-stream-processor detects the new
	// user record and enqueues a "welcome" notification via SQS.

	return authMiddleware.CreateSuccessResponse(200, "Registration successful", map[string]interface{}{
		"token":         *authResult.AuthenticationResult.AccessToken,
		"refresh_token": *authResult.AuthenticationResult.RefreshToken,
		"user":          user,
		"account":       account,
	}), nil
}

// generateSecureCode generates a cryptographically secure 6-digit numeric code
func generateSecureCode() (string, error) {
	max := big.NewInt(1000000) // 0-999999
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", err
	}
	return padCode(n.Int64()), nil
}

func padCode(n int64) string {
	s := make([]byte, 6)
	for i := 5; i >= 0; i-- {
		s[i] = byte('0' + n%10)
		n /= 10
	}
	return string(s)
}
