package team

import (
	"context"
	"encoding/json"
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
	ddbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	authMiddleware "github.com/myfusionhelper/api/internal/middleware/auth"
	apitypes "github.com/myfusionhelper/api/internal/types"
)

var (
	usersTable        = os.Getenv("USERS_TABLE")
	userAccountsTable = os.Getenv("USER_ACCOUNTS_TABLE")
)

type InviteTeamMemberRequest struct {
	Email string `json:"email"`
	Role  string `json:"role"`
}

type UpdateTeamMemberRequest struct {
	Role string `json:"role"`
}

// HandleWithAuth routes team management requests
func HandleWithAuth(ctx context.Context, event events.APIGatewayV2HTTPRequest, authCtx *apitypes.AuthContext) (events.APIGatewayV2HTTPResponse, error) {
	path := event.RequestContext.HTTP.Path
	method := event.RequestContext.HTTP.Method

	switch {
	case strings.HasSuffix(path, "/team") && method == "GET":
		return listTeamMembers(ctx, authCtx)
	case strings.HasSuffix(path, "/team") && method == "POST":
		return inviteTeamMember(ctx, event, authCtx)
	case strings.Contains(path, "/team/") && method == "PUT":
		return updateTeamMember(ctx, event, authCtx)
	case strings.Contains(path, "/team/") && method == "DELETE":
		return removeTeamMember(ctx, event, authCtx)
	default:
		return authMiddleware.CreateErrorResponse(404, "Not Found"), nil
	}
}

func listTeamMembers(ctx context.Context, authCtx *apitypes.AuthContext) (events.APIGatewayV2HTTPResponse, error) {
	log.Printf("List team members for account: %s", authCtx.AccountID)

	region := os.Getenv("COGNITO_REGION")
	if region == "" {
		region = "us-west-2"
	}

	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return authMiddleware.CreateErrorResponse(500, "Internal server error"), nil
	}
	db := dynamodb.NewFromConfig(cfg)

	// Query user_accounts by account_id GSI to find all users in this account
	result, err := db.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(userAccountsTable),
		IndexName:              aws.String("AccountIdIndex"),
		KeyConditionExpression: aws.String("account_id = :account_id"),
		ExpressionAttributeValues: map[string]ddbtypes.AttributeValue{
			":account_id": &ddbtypes.AttributeValueMemberS{Value: authCtx.AccountID},
		},
	})
	if err != nil {
		log.Printf("Failed to query team members: %v", err)
		return authMiddleware.CreateErrorResponse(500, "Failed to list team members"), nil
	}

	var members []map[string]interface{}
	for _, item := range result.Items {
		var ua apitypes.UserAccount
		if err := attributevalue.UnmarshalMap(item, &ua); err != nil {
			continue
		}

		// Fetch user details
		userResult, err := db.GetItem(ctx, &dynamodb.GetItemInput{
			TableName: aws.String(usersTable),
			Key: map[string]ddbtypes.AttributeValue{
				"user_id": &ddbtypes.AttributeValueMemberS{Value: ua.UserID},
			},
		})
		if err != nil || userResult.Item == nil {
			continue
		}

		var user struct {
			UserID string `dynamodbav:"user_id"`
			Email  string `dynamodbav:"email"`
			Name   string `dynamodbav:"name"`
			Status string `dynamodbav:"status"`
		}
		if err := attributevalue.UnmarshalMap(userResult.Item, &user); err != nil {
			continue
		}

		members = append(members, map[string]interface{}{
			"user_id":   user.UserID,
			"email":     user.Email,
			"name":      user.Name,
			"role":      ua.Role,
			"status":    ua.Status,
			"linked_at": ua.LinkedAt,
		})
	}

	return authMiddleware.CreateSuccessResponse(200, "Team members retrieved successfully", map[string]interface{}{
		"members":     members,
		"total_count": len(members),
	}), nil
}

func inviteTeamMember(ctx context.Context, event events.APIGatewayV2HTTPRequest, authCtx *apitypes.AuthContext) (events.APIGatewayV2HTTPResponse, error) {
	log.Printf("Invite team member to account: %s", authCtx.AccountID)

	if !authCtx.Permissions.CanManageTeam {
		return authMiddleware.CreateErrorResponse(403, "Permission denied: cannot manage team"), nil
	}

	var req InviteTeamMemberRequest
	if err := json.Unmarshal([]byte(event.Body), &req); err != nil {
		return authMiddleware.CreateErrorResponse(400, "Invalid request format"), nil
	}

	if req.Email == "" {
		return authMiddleware.CreateErrorResponse(400, "Email is required"), nil
	}
	if req.Role == "" {
		req.Role = "member"
	}

	validRoles := map[string]bool{"admin": true, "member": true, "viewer": true}
	if !validRoles[req.Role] {
		return authMiddleware.CreateErrorResponse(400, "Invalid role. Must be admin, member, or viewer"), nil
	}

	region := os.Getenv("COGNITO_REGION")
	if region == "" {
		region = "us-west-2"
	}

	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return authMiddleware.CreateErrorResponse(500, "Internal server error"), nil
	}
	db := dynamodb.NewFromConfig(cfg)

	// Look up user by email in the EmailIndex GSI
	emailResult, err := db.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(usersTable),
		IndexName:              aws.String("EmailIndex"),
		KeyConditionExpression: aws.String("email = :email"),
		ExpressionAttributeValues: map[string]ddbtypes.AttributeValue{
			":email": &ddbtypes.AttributeValueMemberS{Value: req.Email},
		},
		Limit: aws.Int32(1),
	})
	if err != nil {
		log.Printf("Failed to look up user by email: %v", err)
		return authMiddleware.CreateErrorResponse(500, "Failed to invite team member"), nil
	}

	var targetUserID string

	if len(emailResult.Items) > 0 {
		// User already exists -- add them to this account
		var existingUser struct {
			UserID string `dynamodbav:"user_id"`
		}
		if err := attributevalue.UnmarshalMap(emailResult.Items[0], &existingUser); err != nil {
			return authMiddleware.CreateErrorResponse(500, "Internal server error"), nil
		}
		targetUserID = existingUser.UserID

		// Check if already a member
		existingUA, err := db.GetItem(ctx, &dynamodb.GetItemInput{
			TableName: aws.String(userAccountsTable),
			Key: map[string]ddbtypes.AttributeValue{
				"user_id":    &ddbtypes.AttributeValueMemberS{Value: targetUserID},
				"account_id": &ddbtypes.AttributeValueMemberS{Value: authCtx.AccountID},
			},
		})
		if err == nil && existingUA.Item != nil {
			return authMiddleware.CreateErrorResponse(409, "User is already a member of this account"), nil
		}
	} else {
		// User doesn't exist -- create them in Cognito with a temporary password
		cognitoClient := cognitoidentityprovider.NewFromConfig(cfg)

		createResult, err := cognitoClient.AdminCreateUser(ctx, &cognitoidentityprovider.AdminCreateUserInput{
			UserPoolId: aws.String(os.Getenv("COGNITO_USER_POOL_ID")),
			Username:   aws.String(req.Email),
			UserAttributes: []cognitotypes.AttributeType{
				{Name: aws.String("email"), Value: aws.String(req.Email)},
				{Name: aws.String("email_verified"), Value: aws.String("true")},
			},
			DesiredDeliveryMediums: []cognitotypes.DeliveryMediumType{
				cognitotypes.DeliveryMediumTypeEmail,
			},
		})
		if err != nil {
			log.Printf("Failed to create Cognito user: %v", err)
			return authMiddleware.CreateErrorResponse(500, "Failed to invite team member"), nil
		}

		// Extract the sub from the Cognito response
		var cognitoSub string
		for _, attr := range createResult.User.Attributes {
			if *attr.Name == "sub" {
				cognitoSub = *attr.Value
				break
			}
		}
		targetUserID = "user:" + cognitoSub

		// Create user record in DynamoDB
		now := time.Now().UTC()
		newUser := apitypes.User{
			UserID:           targetUserID,
			CognitoUserID:    cognitoSub,
			Email:            req.Email,
			Name:             req.Email, // Default name to email
			Status:           "invited",
			CurrentAccountID: authCtx.AccountID,
			CreatedAt:        now,
			UpdatedAt:        now,
		}

		userItem, err := attributevalue.MarshalMap(newUser)
		if err != nil {
			return authMiddleware.CreateErrorResponse(500, "Internal server error"), nil
		}

		_, err = db.PutItem(ctx, &dynamodb.PutItemInput{
			TableName: aws.String(usersTable),
			Item:      userItem,
		})
		if err != nil {
			log.Printf("Failed to create user record: %v", err)
			return authMiddleware.CreateErrorResponse(500, "Failed to invite team member"), nil
		}
	}

	// Create user-account relationship
	now := time.Now().UTC()
	permissions := permissionsForRole(req.Role)

	ua := apitypes.UserAccount{
		UserID:      targetUserID,
		AccountID:   authCtx.AccountID,
		Role:        req.Role,
		Status:      "active",
		Permissions: permissions,
		LinkedAt:    now.Format(time.RFC3339),
		UpdatedAt:   now.Format(time.RFC3339),
	}

	uaItem, err := attributevalue.MarshalMap(ua)
	if err != nil {
		return authMiddleware.CreateErrorResponse(500, "Internal server error"), nil
	}

	_, err = db.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(userAccountsTable),
		Item:      uaItem,
	})
	if err != nil {
		log.Printf("Failed to create user-account link: %v", err)
		return authMiddleware.CreateErrorResponse(500, "Failed to invite team member"), nil
	}

	return authMiddleware.CreateSuccessResponse(201, "Team member invited successfully", map[string]interface{}{
		"user_id": targetUserID,
		"email":   req.Email,
		"role":    req.Role,
		"status":  "active",
	}), nil
}

func updateTeamMember(ctx context.Context, event events.APIGatewayV2HTTPRequest, authCtx *apitypes.AuthContext) (events.APIGatewayV2HTTPResponse, error) {
	// Extract user_id from path: /accounts/{account_id}/team/{user_id}
	path := event.RequestContext.HTTP.Path
	parts := strings.Split(path, "/team/")
	if len(parts) < 2 || parts[1] == "" {
		return authMiddleware.CreateErrorResponse(400, "User ID is required"), nil
	}
	targetUserID := parts[1]

	log.Printf("Update team member %s in account: %s", targetUserID, authCtx.AccountID)

	if !authCtx.Permissions.CanManageTeam {
		return authMiddleware.CreateErrorResponse(403, "Permission denied: cannot manage team"), nil
	}

	var req UpdateTeamMemberRequest
	if err := json.Unmarshal([]byte(event.Body), &req); err != nil {
		return authMiddleware.CreateErrorResponse(400, "Invalid request format"), nil
	}

	validRoles := map[string]bool{"admin": true, "member": true, "viewer": true}
	if !validRoles[req.Role] {
		return authMiddleware.CreateErrorResponse(400, "Invalid role. Must be admin, member, or viewer"), nil
	}

	region := os.Getenv("COGNITO_REGION")
	if region == "" {
		region = "us-west-2"
	}

	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return authMiddleware.CreateErrorResponse(500, "Internal server error"), nil
	}
	db := dynamodb.NewFromConfig(cfg)

	permissions := permissionsForRole(req.Role)
	permAV, err := attributevalue.MarshalMap(permissions)
	if err != nil {
		return authMiddleware.CreateErrorResponse(500, "Internal server error"), nil
	}

	_, err = db.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(userAccountsTable),
		Key: map[string]ddbtypes.AttributeValue{
			"user_id":    &ddbtypes.AttributeValueMemberS{Value: targetUserID},
			"account_id": &ddbtypes.AttributeValueMemberS{Value: authCtx.AccountID},
		},
		UpdateExpression: aws.String("SET #r = :role, permissions = :permissions, updated_at = :updated_at"),
		ExpressionAttributeNames: map[string]string{
			"#r": "role",
		},
		ExpressionAttributeValues: map[string]ddbtypes.AttributeValue{
			":role":        &ddbtypes.AttributeValueMemberS{Value: req.Role},
			":permissions": &ddbtypes.AttributeValueMemberM{Value: permAV},
			":updated_at":  &ddbtypes.AttributeValueMemberS{Value: time.Now().UTC().Format(time.RFC3339)},
		},
		ConditionExpression: aws.String("attribute_exists(user_id)"),
	})
	if err != nil {
		log.Printf("Failed to update team member: %v", err)
		return authMiddleware.CreateErrorResponse(500, "Failed to update team member"), nil
	}

	return authMiddleware.CreateSuccessResponse(200, "Team member updated successfully", map[string]interface{}{
		"user_id": targetUserID,
		"role":    req.Role,
	}), nil
}

func removeTeamMember(ctx context.Context, event events.APIGatewayV2HTTPRequest, authCtx *apitypes.AuthContext) (events.APIGatewayV2HTTPResponse, error) {
	path := event.RequestContext.HTTP.Path
	parts := strings.Split(path, "/team/")
	if len(parts) < 2 || parts[1] == "" {
		return authMiddleware.CreateErrorResponse(400, "User ID is required"), nil
	}
	targetUserID := parts[1]

	log.Printf("Remove team member %s from account: %s", targetUserID, authCtx.AccountID)

	if !authCtx.Permissions.CanManageTeam {
		return authMiddleware.CreateErrorResponse(403, "Permission denied: cannot manage team"), nil
	}

	// Prevent removing yourself
	if targetUserID == authCtx.UserID {
		return authMiddleware.CreateErrorResponse(400, "Cannot remove yourself from the account"), nil
	}

	region := os.Getenv("COGNITO_REGION")
	if region == "" {
		region = "us-west-2"
	}

	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return authMiddleware.CreateErrorResponse(500, "Internal server error"), nil
	}
	db := dynamodb.NewFromConfig(cfg)

	_, err = db.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: aws.String(userAccountsTable),
		Key: map[string]ddbtypes.AttributeValue{
			"user_id":    &ddbtypes.AttributeValueMemberS{Value: targetUserID},
			"account_id": &ddbtypes.AttributeValueMemberS{Value: authCtx.AccountID},
		},
		ConditionExpression: aws.String("attribute_exists(user_id)"),
	})
	if err != nil {
		log.Printf("Failed to remove team member: %v", err)
		return authMiddleware.CreateErrorResponse(500, "Failed to remove team member"), nil
	}

	return authMiddleware.CreateSuccessResponse(200, "Team member removed successfully", map[string]interface{}{
		"user_id": targetUserID,
	}), nil
}

// permissionsForRole returns the default permissions for a given role
func permissionsForRole(role string) apitypes.Permissions {
	switch role {
	case "admin":
		return apitypes.Permissions{
			CanManageHelpers:     true,
			CanExecuteHelpers:    true,
			CanManageConnections: true,
			CanManageTeam:        true,
			CanManageBilling:     true,
			CanViewAnalytics:     true,
			CanManageAPIKeys:     true,
		}
	case "member":
		return apitypes.Permissions{
			CanManageHelpers:     true,
			CanExecuteHelpers:    true,
			CanManageConnections: true,
			CanManageTeam:        false,
			CanManageBilling:     false,
			CanViewAnalytics:     true,
			CanManageAPIKeys:     false,
		}
	case "viewer":
		return apitypes.Permissions{
			CanManageHelpers:     false,
			CanExecuteHelpers:    false,
			CanManageConnections: false,
			CanManageTeam:        false,
			CanManageBilling:     false,
			CanViewAnalytics:     true,
			CanManageAPIKeys:     false,
		}
	default:
		return apitypes.Permissions{}
	}
}
