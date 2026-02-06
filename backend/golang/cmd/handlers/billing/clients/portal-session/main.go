package portalsession

import (
	"context"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	ddbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	authMiddleware "github.com/myfusionhelper/api/internal/middleware/auth"
	"github.com/myfusionhelper/api/internal/types"
	stripe "github.com/stripe/stripe-go/v82"
	"github.com/stripe/stripe-go/v82/billingportal/session"
)

var (
	accountsTable = os.Getenv("ACCOUNTS_TABLE")
	stripeKey     = os.Getenv("STRIPE_SECRET_KEY")
	appURL        = os.Getenv("APP_URL")
)

// HandleWithAuth creates a Stripe Customer Portal session
func HandleWithAuth(ctx context.Context, event events.APIGatewayV2HTTPRequest, authCtx *types.AuthContext) (events.APIGatewayV2HTTPResponse, error) {
	log.Printf("PortalSession handler called for account %s", authCtx.AccountID)

	if event.RequestContext.HTTP.Method != "POST" {
		return authMiddleware.CreateErrorResponse(405, "Method not allowed"), nil
	}

	if stripeKey == "" {
		return authMiddleware.CreateErrorResponse(503, "Billing service not configured"), nil
	}

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		log.Printf("Failed to load AWS config: %v", err)
		return authMiddleware.CreateErrorResponse(500, "Internal error"), nil
	}

	dbClient := dynamodb.NewFromConfig(cfg)

	// Fetch account to get Stripe customer ID
	result, err := dbClient.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(accountsTable),
		Key: map[string]ddbtypes.AttributeValue{
			"account_id": &ddbtypes.AttributeValueMemberS{Value: authCtx.AccountID},
		},
	})
	if err != nil {
		log.Printf("Failed to fetch account: %v", err)
		return authMiddleware.CreateErrorResponse(500, "Internal error"), nil
	}
	if result.Item == nil {
		return authMiddleware.CreateErrorResponse(404, "Account not found"), nil
	}

	var account types.Account
	if err := attributevalue.UnmarshalMap(result.Item, &account); err != nil {
		log.Printf("Failed to unmarshal account: %v", err)
		return authMiddleware.CreateErrorResponse(500, "Internal error"), nil
	}

	if account.StripeCustomerID == "" {
		return authMiddleware.CreateErrorResponse(400, "No billing account found. Please subscribe to a plan first."), nil
	}

	baseURL := appURL
	if baseURL == "" {
		baseURL = "https://app.myfusionhelper.ai"
	}
	returnURL := baseURL + "/settings?tab=billing"

	stripe.Key = stripeKey
	params := &stripe.BillingPortalSessionParams{
		Customer:  stripe.String(account.StripeCustomerID),
		ReturnURL: stripe.String(returnURL),
	}

	s, err := session.New(params)
	if err != nil {
		log.Printf("Failed to create portal session: %v", err)
		return authMiddleware.CreateErrorResponse(500, "Failed to create billing portal session"), nil
	}

	return authMiddleware.CreateSuccessResponse(200, "OK", map[string]string{
		"url": s.URL,
	}), nil
}
