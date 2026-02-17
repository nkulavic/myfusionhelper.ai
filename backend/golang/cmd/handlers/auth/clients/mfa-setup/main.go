package mfasetup

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	cognitotypes "github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider/types"
	"github.com/myfusionhelper/api/internal/apiutil"
	authMiddleware "github.com/myfusionhelper/api/internal/middleware/auth"
	"github.com/myfusionhelper/api/internal/types"
)

var (
	cognitoUserPoolID = os.Getenv("COGNITO_USER_POOL_ID")
)

// HandleWithAuth routes MFA setup requests based on HTTP method and path
func HandleWithAuth(ctx context.Context, event events.APIGatewayV2HTTPRequest, authCtx *types.AuthContext) (events.APIGatewayV2HTTPResponse, error) {
	path := event.RequestContext.HTTP.Path
	method := event.RequestContext.HTTP.Method

	log.Printf("MFA Setup handler: path=%s method=%s", path, method)

	switch {
	case path == "/auth/mfa/status" && method == "GET":
		return handleGetStatus(ctx, event, authCtx)
	case path == "/auth/mfa/setup-totp" && method == "POST":
		return handleSetupTotp(ctx, event, authCtx)
	case path == "/auth/mfa/verify-totp" && method == "POST":
		return handleVerifyTotp(ctx, event, authCtx)
	case path == "/auth/mfa/enable-sms" && method == "POST":
		return handleEnableSms(ctx, event, authCtx)
	case path == "/auth/mfa/disable" && method == "POST":
		return handleDisable(ctx, event, authCtx)
	default:
		return authMiddleware.CreateErrorResponse(404, "Not Found"), nil
	}
}

func getCognitoClient(ctx context.Context) (*cognitoidentityprovider.Client, error) {
	region := os.Getenv("COGNITO_REGION")
	if region == "" {
		region = "us-west-2"
	}
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return nil, err
	}
	return cognitoidentityprovider.NewFromConfig(cfg), nil
}

// handleGetStatus returns the current MFA state for the user
func handleGetStatus(ctx context.Context, _ events.APIGatewayV2HTTPRequest, authCtx *types.AuthContext) (events.APIGatewayV2HTTPResponse, error) {
	client, err := getCognitoClient(ctx)
	if err != nil {
		log.Printf("Failed to create Cognito client: %v", err)
		return authMiddleware.CreateErrorResponse(500, "Failed to get MFA status"), nil
	}

	// Get user's MFA settings from Cognito
	cognitoSub := authCtx.UserID[len("user:"):]
	userResult, err := client.AdminGetUser(ctx, &cognitoidentityprovider.AdminGetUserInput{
		UserPoolId: aws.String(cognitoUserPoolID),
		Username:   aws.String(cognitoSub),
	})
	if err != nil {
		log.Printf("Failed to get user from Cognito: %v", err)
		return authMiddleware.CreateErrorResponse(500, "Failed to get MFA status"), nil
	}

	enabled := false
	var method *string
	phoneVerified := false

	// Check MFA preferences
	for _, mfaPref := range userResult.UserMFASettingList {
		if mfaPref == "SOFTWARE_TOKEN_MFA" {
			enabled = true
			m := "totp"
			method = &m
		} else if mfaPref == "SMS_MFA" {
			enabled = true
			m := "sms"
			method = &m
		}
	}

	// Check if phone is verified
	for _, attr := range userResult.UserAttributes {
		if aws.ToString(attr.Name) == "phone_number_verified" && aws.ToString(attr.Value) == "true" {
			phoneVerified = true
		}
	}

	response := map[string]interface{}{
		"enabled":        enabled,
		"method":         method,
		"phone_verified": phoneVerified,
	}

	return authMiddleware.CreateSuccessResponse(200, "MFA status retrieved", response), nil
}

// handleSetupTotp initiates TOTP setup and returns the secret for QR code
func handleSetupTotp(ctx context.Context, event events.APIGatewayV2HTTPRequest, authCtx *types.AuthContext) (events.APIGatewayV2HTTPResponse, error) {
	client, err := getCognitoClient(ctx)
	if err != nil {
		log.Printf("Failed to create Cognito client: %v", err)
		return authMiddleware.CreateErrorResponse(500, "Failed to setup TOTP"), nil
	}

	// Get the access token from the Authorization header
	accessToken := event.Headers["authorization"]
	if len(accessToken) > 7 && accessToken[:7] == "Bearer " {
		accessToken = accessToken[7:]
	}

	result, err := client.AssociateSoftwareToken(ctx, &cognitoidentityprovider.AssociateSoftwareTokenInput{
		AccessToken: aws.String(accessToken),
	})
	if err != nil {
		log.Printf("Failed to associate software token: %v", err)
		return authMiddleware.CreateErrorResponse(500, "Failed to setup TOTP"), nil
	}

	secret := aws.ToString(result.SecretCode)
	qrCodeUri := fmt.Sprintf("otpauth://totp/MyFusionHelper:%s?secret=%s&issuer=MyFusionHelper", authCtx.Email, secret)

	return authMiddleware.CreateSuccessResponse(200, "TOTP setup initiated", map[string]interface{}{
		"secret":       secret,
		"qr_code_uri": qrCodeUri,
	}), nil
}

// handleVerifyTotp verifies the TOTP code and enables TOTP MFA
func handleVerifyTotp(ctx context.Context, event events.APIGatewayV2HTTPRequest, authCtx *types.AuthContext) (events.APIGatewayV2HTTPResponse, error) {
	body := apiutil.GetBody(event)
	var req struct {
		Code string `json:"code"`
	}
	if err := json.Unmarshal([]byte(body), &req); err != nil || req.Code == "" {
		return authMiddleware.CreateErrorResponse(400, "Verification code is required"), nil
	}

	client, err := getCognitoClient(ctx)
	if err != nil {
		log.Printf("Failed to create Cognito client: %v", err)
		return authMiddleware.CreateErrorResponse(500, "Failed to verify TOTP"), nil
	}

	accessToken := event.Headers["authorization"]
	if len(accessToken) > 7 && accessToken[:7] == "Bearer " {
		accessToken = accessToken[7:]
	}

	// Verify the software token
	_, err = client.VerifySoftwareToken(ctx, &cognitoidentityprovider.VerifySoftwareTokenInput{
		AccessToken: aws.String(accessToken),
		UserCode:    aws.String(req.Code),
	})
	if err != nil {
		log.Printf("TOTP verification failed: %v", err)
		return authMiddleware.CreateErrorResponse(401, "Invalid verification code"), nil
	}

	// Enable TOTP MFA for the user
	cognitoSub := authCtx.UserID[len("user:"):]
	_, err = client.AdminSetUserMFAPreference(ctx, &cognitoidentityprovider.AdminSetUserMFAPreferenceInput{
		UserPoolId: aws.String(cognitoUserPoolID),
		Username:   aws.String(cognitoSub),
		SoftwareTokenMfaSettings: &cognitotypes.SoftwareTokenMfaSettingsType{
			Enabled:      true,
			PreferredMfa: true,
		},
	})
	if err != nil {
		log.Printf("Failed to set MFA preference: %v", err)
		return authMiddleware.CreateErrorResponse(500, "Failed to enable TOTP"), nil
	}

	return authMiddleware.CreateSuccessResponse(200, "TOTP MFA enabled", nil), nil
}

// handleEnableSms enables SMS MFA for the user
func handleEnableSms(ctx context.Context, _ events.APIGatewayV2HTTPRequest, authCtx *types.AuthContext) (events.APIGatewayV2HTTPResponse, error) {
	client, err := getCognitoClient(ctx)
	if err != nil {
		log.Printf("Failed to create Cognito client: %v", err)
		return authMiddleware.CreateErrorResponse(500, "Failed to enable SMS MFA"), nil
	}

	cognitoSub := authCtx.UserID[len("user:"):]

	// Check if user has a verified phone number
	userResult, err := client.AdminGetUser(ctx, &cognitoidentityprovider.AdminGetUserInput{
		UserPoolId: aws.String(cognitoUserPoolID),
		Username:   aws.String(cognitoSub),
	})
	if err != nil {
		log.Printf("Failed to get user: %v", err)
		return authMiddleware.CreateErrorResponse(500, "Failed to enable SMS MFA"), nil
	}

	phoneVerified := false
	for _, attr := range userResult.UserAttributes {
		if aws.ToString(attr.Name) == "phone_number_verified" && aws.ToString(attr.Value) == "true" {
			phoneVerified = true
		}
	}

	if !phoneVerified {
		return authMiddleware.CreateErrorResponse(400, "A verified phone number is required for SMS MFA"), nil
	}

	_, err = client.AdminSetUserMFAPreference(ctx, &cognitoidentityprovider.AdminSetUserMFAPreferenceInput{
		UserPoolId: aws.String(cognitoUserPoolID),
		Username:   aws.String(cognitoSub),
		SMSMfaSettings: &cognitotypes.SMSMfaSettingsType{
			Enabled:      true,
			PreferredMfa: true,
		},
	})
	if err != nil {
		log.Printf("Failed to set SMS MFA preference: %v", err)
		return authMiddleware.CreateErrorResponse(500, "Failed to enable SMS MFA"), nil
	}

	return authMiddleware.CreateSuccessResponse(200, "SMS MFA enabled", nil), nil
}

// handleDisable disables all MFA for the user
func handleDisable(ctx context.Context, _ events.APIGatewayV2HTTPRequest, authCtx *types.AuthContext) (events.APIGatewayV2HTTPResponse, error) {
	client, err := getCognitoClient(ctx)
	if err != nil {
		log.Printf("Failed to create Cognito client: %v", err)
		return authMiddleware.CreateErrorResponse(500, "Failed to disable MFA"), nil
	}

	cognitoSub := authCtx.UserID[len("user:"):]

	_, err = client.AdminSetUserMFAPreference(ctx, &cognitoidentityprovider.AdminSetUserMFAPreferenceInput{
		UserPoolId: aws.String(cognitoUserPoolID),
		Username:   aws.String(cognitoSub),
		SoftwareTokenMfaSettings: &cognitotypes.SoftwareTokenMfaSettingsType{
			Enabled:      false,
			PreferredMfa: false,
		},
		SMSMfaSettings: &cognitotypes.SMSMfaSettingsType{
			Enabled:      false,
			PreferredMfa: false,
		},
	})
	if err != nil {
		log.Printf("Failed to disable MFA: %v", err)
		return authMiddleware.CreateErrorResponse(500, "Failed to disable MFA"), nil
	}

	return authMiddleware.CreateSuccessResponse(200, "MFA disabled", nil), nil
}
