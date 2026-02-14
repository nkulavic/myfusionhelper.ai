package profile

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
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	cognitotypes "github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider/types"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	ddbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	authMiddleware "github.com/myfusionhelper/api/internal/middleware/auth"
	apitypes "github.com/myfusionhelper/api/internal/types"
)

type UpdateProfileRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

// HandleWithAuth handles profile update requests
func HandleWithAuth(ctx context.Context, event events.APIGatewayV2HTTPRequest, authCtx *apitypes.AuthContext) (events.APIGatewayV2HTTPResponse, error) {
	method := event.RequestContext.HTTP.Method

	switch method {
	case "PUT":
		return updateProfile(ctx, event, authCtx)
	default:
		return authMiddleware.CreateErrorResponse(405, "Method not allowed"), nil
	}
}

func updateProfile(ctx context.Context, event events.APIGatewayV2HTTPRequest, authCtx *apitypes.AuthContext) (events.APIGatewayV2HTTPResponse, error) {
	log.Printf("Update profile for user: %s", authCtx.UserID)

	var req UpdateProfileRequest
	if err := json.Unmarshal([]byte(event.Body), &req); err != nil {
		return authMiddleware.CreateErrorResponse(400, "Invalid request format"), nil
	}

	if req.Name == "" && req.Email == "" {
		return authMiddleware.CreateErrorResponse(400, "At least one field to update is required"), nil
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
	usersTable := os.Getenv("USERS_TABLE")

	// Build DynamoDB update expression
	updateParts := []string{"updated_at = :updated_at"}
	exprValues := map[string]ddbtypes.AttributeValue{
		":updated_at": &ddbtypes.AttributeValueMemberS{Value: time.Now().UTC().Format(time.RFC3339)},
	}
	exprNames := map[string]string{}

	if req.Name != "" {
		updateParts = append(updateParts, "#n = :name")
		exprValues[":name"] = &ddbtypes.AttributeValueMemberS{Value: req.Name}
		exprNames["#n"] = "name"
	}

	if req.Email != "" {
		// Validate email format
		if !strings.Contains(req.Email, "@") || !strings.Contains(req.Email, ".") {
			return authMiddleware.CreateErrorResponse(400, "Invalid email format"), nil
		}

		// Update email in Cognito first
		cognitoSub := strings.TrimPrefix(authCtx.UserID, "user:")
		cognitoClient := cognitoidentityprovider.NewFromConfig(cfg)

		_, err := cognitoClient.AdminUpdateUserAttributes(ctx, &cognitoidentityprovider.AdminUpdateUserAttributesInput{
			UserPoolId: aws.String(os.Getenv("COGNITO_USER_POOL_ID")),
			Username:   aws.String(cognitoSub),
			UserAttributes: []cognitotypes.AttributeType{
				{Name: aws.String("email"), Value: aws.String(req.Email)},
				{Name: aws.String("email_verified"), Value: aws.String("true")},
			},
		})
		if err != nil {
			log.Printf("Failed to update Cognito email: %v", err)
			return authMiddleware.CreateErrorResponse(500, "Failed to update email"), nil
		}

		updateParts = append(updateParts, "email = :email")
		exprValues[":email"] = &ddbtypes.AttributeValueMemberS{Value: req.Email}
	}

	updateInput := &dynamodb.UpdateItemInput{
		TableName: aws.String(usersTable),
		Key: map[string]ddbtypes.AttributeValue{
			"user_id": &ddbtypes.AttributeValueMemberS{Value: authCtx.UserID},
		},
		UpdateExpression:          aws.String("SET " + strings.Join(updateParts, ", ")),
		ExpressionAttributeValues: exprValues,
		ReturnValues:              ddbtypes.ReturnValueAllNew,
	}
	if len(exprNames) > 0 {
		updateInput.ExpressionAttributeNames = exprNames
	}

	result, err := db.UpdateItem(ctx, updateInput)
	if err != nil {
		log.Printf("Failed to update user profile: %v", err)
		return authMiddleware.CreateErrorResponse(500, "Failed to update profile"), nil
	}

	// Extract updated fields from response
	responseData := map[string]interface{}{
		"user_id": authCtx.UserID,
	}
	if v, ok := result.Attributes["name"]; ok {
		if sv, ok := v.(*ddbtypes.AttributeValueMemberS); ok {
			responseData["name"] = sv.Value
		}
	}
	if v, ok := result.Attributes["email"]; ok {
		if sv, ok := v.(*ddbtypes.AttributeValueMemberS); ok {
			responseData["email"] = sv.Value
		}
	}
	if v, ok := result.Attributes["updated_at"]; ok {
		if sv, ok := v.(*ddbtypes.AttributeValueMemberS); ok {
			responseData["updated_at"] = sv.Value
		}
	}

	return authMiddleware.CreateSuccessResponse(200, "Profile updated successfully", responseData), nil
}
