package crud

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
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	ddbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	authMiddleware "github.com/myfusionhelper/api/internal/middleware/auth"
	apitypes "github.com/myfusionhelper/api/internal/types"
)

var (
	usersTable        = os.Getenv("USERS_TABLE")
	accountsTable     = os.Getenv("ACCOUNTS_TABLE")
	userAccountsTable = os.Getenv("USER_ACCOUNTS_TABLE")
)

type UpdateAccountRequest struct {
	Name    string `json:"name"`
	Company string `json:"company"`
}

type SwitchAccountRequest struct {
	AccountID string `json:"account_id"`
}

// HandleWithAuth routes to the appropriate CRUD operation based on path and method
func HandleWithAuth(ctx context.Context, event events.APIGatewayV2HTTPRequest, authCtx *apitypes.AuthContext) (events.APIGatewayV2HTTPResponse, error) {
	path := event.RequestContext.HTTP.Path
	method := event.RequestContext.HTTP.Method

	switch {
	case path == "/accounts/switch" && method == "POST":
		return switchAccount(ctx, event, authCtx)
	case path == "/accounts" && method == "GET":
		return listAccounts(ctx, event, authCtx)
	case strings.HasPrefix(path, "/accounts/") && method == "GET":
		return getAccount(ctx, event, authCtx)
	case strings.HasPrefix(path, "/accounts/") && method == "PUT":
		return updateAccount(ctx, event, authCtx)
	default:
		return authMiddleware.CreateErrorResponse(404, "Not Found"), nil
	}
}

func listAccounts(ctx context.Context, event events.APIGatewayV2HTTPRequest, authCtx *apitypes.AuthContext) (events.APIGatewayV2HTTPResponse, error) {
	log.Printf("List accounts for user: %s", authCtx.UserID)

	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(os.Getenv("COGNITO_REGION")))
	if err != nil {
		return authMiddleware.CreateErrorResponse(500, "Internal server error"), nil
	}
	db := dynamodb.NewFromConfig(cfg)

	// Query user-accounts table for all accounts this user has access to
	result, err := db.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(userAccountsTable),
		KeyConditionExpression: aws.String("user_id = :user_id"),
		ExpressionAttributeValues: map[string]ddbtypes.AttributeValue{
			":user_id": &ddbtypes.AttributeValueMemberS{Value: authCtx.UserID},
		},
	})
	if err != nil {
		log.Printf("Failed to query user accounts: %v", err)
		return authMiddleware.CreateErrorResponse(500, "Failed to list accounts"), nil
	}

	var accounts []map[string]interface{}
	for _, item := range result.Items {
		var ua apitypes.UserAccount
		if err := attributevalue.UnmarshalMap(item, &ua); err != nil {
			continue
		}

		// Fetch account details
		accountResult, err := db.GetItem(ctx, &dynamodb.GetItemInput{
			TableName: aws.String(accountsTable),
			Key: map[string]ddbtypes.AttributeValue{
				"account_id": &ddbtypes.AttributeValueMemberS{Value: ua.AccountID},
			},
		})
		if err != nil || accountResult.Item == nil {
			continue
		}

		var account apitypes.Account
		if err := attributevalue.UnmarshalMap(accountResult.Item, &account); err != nil {
			continue
		}

		accounts = append(accounts, map[string]interface{}{
			"account_id":  account.AccountID,
			"name":        account.Name,
			"company":     account.Company,
			"plan":        account.Plan,
			"status":      account.Status,
			"user_role":   ua.Role,
			"user_status": ua.Status,
			"is_current":  ua.AccountID == authCtx.AccountID,
			"linked_at":   ua.LinkedAt,
			"updated_at":  account.UpdatedAt,
		})
	}

	return authMiddleware.CreateSuccessResponse(200, "Accounts retrieved successfully", map[string]interface{}{
		"accounts":           accounts,
		"total_count":        len(accounts),
		"current_account_id": authCtx.AccountID,
	}), nil
}

func getAccount(ctx context.Context, event events.APIGatewayV2HTTPRequest, authCtx *apitypes.AuthContext) (events.APIGatewayV2HTTPResponse, error) {
	accountID := event.PathParameters["account_id"]
	if accountID == "" {
		return authMiddleware.CreateErrorResponse(400, "Account ID is required"), nil
	}

	log.Printf("Get account %s for user: %s", accountID, authCtx.UserID)

	// Verify user has access to this account
	if !userHasAccountAccess(authCtx, accountID) {
		return authMiddleware.CreateErrorResponse(404, "Account not found"), nil
	}

	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(os.Getenv("COGNITO_REGION")))
	if err != nil {
		return authMiddleware.CreateErrorResponse(500, "Internal server error"), nil
	}
	db := dynamodb.NewFromConfig(cfg)

	result, err := db.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(accountsTable),
		Key: map[string]ddbtypes.AttributeValue{
			"account_id": &ddbtypes.AttributeValueMemberS{Value: accountID},
		},
	})
	if err != nil || result.Item == nil {
		return authMiddleware.CreateErrorResponse(404, "Account not found"), nil
	}

	var account apitypes.Account
	if err := attributevalue.UnmarshalMap(result.Item, &account); err != nil {
		return authMiddleware.CreateErrorResponse(500, "Internal server error"), nil
	}

	return authMiddleware.CreateSuccessResponse(200, "Account retrieved successfully", map[string]interface{}{
		"account_id":  account.AccountID,
		"name":        account.Name,
		"company":     account.Company,
		"plan":        account.Plan,
		"status":      account.Status,
		"settings":    account.Settings,
		"usage":       account.Usage,
		"created_at":  account.CreatedAt,
		"updated_at":  account.UpdatedAt,
	}), nil
}

func updateAccount(ctx context.Context, event events.APIGatewayV2HTTPRequest, authCtx *apitypes.AuthContext) (events.APIGatewayV2HTTPResponse, error) {
	accountID := event.PathParameters["account_id"]
	if accountID == "" {
		return authMiddleware.CreateErrorResponse(400, "Account ID is required"), nil
	}

	log.Printf("Update account %s for user: %s", accountID, authCtx.UserID)

	// Verify user has access and permission
	if !userHasAccountAccess(authCtx, accountID) {
		return authMiddleware.CreateErrorResponse(404, "Account not found"), nil
	}

	var req UpdateAccountRequest
	if err := json.Unmarshal([]byte(event.Body), &req); err != nil {
		return authMiddleware.CreateErrorResponse(400, "Invalid request format"), nil
	}

	if req.Name == "" && req.Company == "" {
		return authMiddleware.CreateErrorResponse(400, "At least one field to update is required"), nil
	}

	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(os.Getenv("COGNITO_REGION")))
	if err != nil {
		return authMiddleware.CreateErrorResponse(500, "Internal server error"), nil
	}
	db := dynamodb.NewFromConfig(cfg)

	// Build update expression
	updateParts := []string{"updated_at = :updated_at"}
	exprValues := map[string]ddbtypes.AttributeValue{
		":updated_at": &ddbtypes.AttributeValueMemberS{Value: time.Now().UTC().Format(time.RFC3339)},
	}

	if req.Name != "" {
		updateParts = append(updateParts, "#n = :name")
		exprValues[":name"] = &ddbtypes.AttributeValueMemberS{Value: req.Name}
	}
	if req.Company != "" {
		updateParts = append(updateParts, "company = :company")
		exprValues[":company"] = &ddbtypes.AttributeValueMemberS{Value: req.Company}
	}

	updateInput := &dynamodb.UpdateItemInput{
		TableName: aws.String(accountsTable),
		Key: map[string]ddbtypes.AttributeValue{
			"account_id": &ddbtypes.AttributeValueMemberS{Value: accountID},
		},
		UpdateExpression:          aws.String("SET " + strings.Join(updateParts, ", ")),
		ExpressionAttributeValues: exprValues,
		ReturnValues:              ddbtypes.ReturnValueAllNew,
	}

	// Add expression attribute names if 'name' is being updated (reserved word)
	if req.Name != "" {
		updateInput.ExpressionAttributeNames = map[string]string{
			"#n": "name",
		}
	}

	result, err := db.UpdateItem(ctx, updateInput)
	if err != nil {
		log.Printf("Failed to update account: %v", err)
		return authMiddleware.CreateErrorResponse(500, "Failed to update account"), nil
	}

	var updated apitypes.Account
	if err := attributevalue.UnmarshalMap(result.Attributes, &updated); err != nil {
		return authMiddleware.CreateErrorResponse(500, "Internal server error"), nil
	}

	return authMiddleware.CreateSuccessResponse(200, "Account updated successfully", map[string]interface{}{
		"account_id": updated.AccountID,
		"name":       updated.Name,
		"company":    updated.Company,
		"updated_at": updated.UpdatedAt,
	}), nil
}

func switchAccount(ctx context.Context, event events.APIGatewayV2HTTPRequest, authCtx *apitypes.AuthContext) (events.APIGatewayV2HTTPResponse, error) {
	log.Printf("Switch account for user: %s", authCtx.UserID)

	var req SwitchAccountRequest
	if err := json.Unmarshal([]byte(event.Body), &req); err != nil {
		return authMiddleware.CreateErrorResponse(400, "Invalid request format"), nil
	}

	if req.AccountID == "" {
		return authMiddleware.CreateErrorResponse(400, "Account ID is required"), nil
	}

	// Verify user has access to the target account
	if !userHasAccountAccess(authCtx, req.AccountID) {
		return authMiddleware.CreateErrorResponse(403, "You don't have access to this account"), nil
	}

	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(os.Getenv("COGNITO_REGION")))
	if err != nil {
		return authMiddleware.CreateErrorResponse(500, "Internal server error"), nil
	}
	db := dynamodb.NewFromConfig(cfg)

	// Update user's current_account_id
	_, err = db.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(usersTable),
		Key: map[string]ddbtypes.AttributeValue{
			"user_id": &ddbtypes.AttributeValueMemberS{Value: authCtx.UserID},
		},
		UpdateExpression: aws.String("SET current_account_id = :account_id, updated_at = :updated_at"),
		ExpressionAttributeValues: map[string]ddbtypes.AttributeValue{
			":account_id": &ddbtypes.AttributeValueMemberS{Value: req.AccountID},
			":updated_at": &ddbtypes.AttributeValueMemberS{Value: time.Now().UTC().Format(time.RFC3339)},
		},
	})
	if err != nil {
		log.Printf("Failed to switch account: %v", err)
		return authMiddleware.CreateErrorResponse(500, "Failed to switch account"), nil
	}

	return authMiddleware.CreateSuccessResponse(200, "Account switched successfully", map[string]interface{}{
		"user_id":            authCtx.UserID,
		"current_account_id": req.AccountID,
	}), nil
}

// userHasAccountAccess checks if the authenticated user has access to a specific account
func userHasAccountAccess(authCtx *apitypes.AuthContext, accountID string) bool {
	for _, acc := range authCtx.AvailableAccounts {
		if acc.AccountID == accountID {
			return true
		}
	}
	return false
}
