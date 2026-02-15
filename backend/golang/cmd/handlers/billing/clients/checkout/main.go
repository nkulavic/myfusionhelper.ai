package checkout

import (
	"context"
	"encoding/json"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	ddbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	appConfig "github.com/myfusionhelper/api/internal/config"
	authMiddleware "github.com/myfusionhelper/api/internal/middleware/auth"
	"github.com/myfusionhelper/api/internal/types"
	stripe "github.com/stripe/stripe-go/v82"
	"github.com/stripe/stripe-go/v82/checkout/session"
	"github.com/stripe/stripe-go/v82/customer"
)

var (
	accountsTable = os.Getenv("ACCOUNTS_TABLE")
	appURL        = os.Getenv("APP_URL")
)

// CreateCheckoutRequest is the request body for POST /billing/checkout/sessions
type CreateCheckoutRequest struct {
	Plan          string `json:"plan"`            // "start", "grow", or "deliver"
	ReturnURL     string `json:"return_url"`      // optional: redirect back to this path after checkout (e.g., "/onboarding")
	BillingPeriod string `json:"billing_period"`  // "monthly" or "annual" (default: "monthly")
}

// HandleWithAuth creates a Stripe Checkout session for a new subscription
func HandleWithAuth(ctx context.Context, event events.APIGatewayV2HTTPRequest, authCtx *types.AuthContext) (events.APIGatewayV2HTTPResponse, error) {
	log.Printf("Checkout handler called for account %s", authCtx.AccountID)

	if event.RequestContext.HTTP.Method != "POST" {
		return authMiddleware.CreateErrorResponse(405, "Method not allowed"), nil
	}

	secrets, err := appConfig.LoadSecrets(ctx)
	if err != nil {
		log.Printf("Failed to load secrets: %v", err)
		return authMiddleware.CreateErrorResponse(500, "Config error"), nil
	}

	if secrets.Stripe.SecretKey == "" {
		return authMiddleware.CreateErrorResponse(503, "Billing service not configured"), nil
	}

	var req CreateCheckoutRequest
	if err := json.Unmarshal([]byte(event.Body), &req); err != nil {
		return authMiddleware.CreateErrorResponse(400, "Invalid request body"), nil
	}

	billingPeriod := req.BillingPeriod
	if billingPeriod == "" {
		billingPeriod = "monthly"
	}
	if billingPeriod != "monthly" && billingPeriod != "annual" {
		return authMiddleware.CreateErrorResponse(400, "Invalid billing_period. Must be 'monthly' or 'annual'"), nil
	}

	priceID := getPriceID(req.Plan, billingPeriod, secrets)
	if priceID == "" {
		return authMiddleware.CreateErrorResponse(400, "Invalid plan. Must be one of: start, grow, deliver"), nil
	}

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		log.Printf("Failed to load AWS config: %v", err)
		return authMiddleware.CreateErrorResponse(500, "Internal error"), nil
	}

	dbClient := dynamodb.NewFromConfig(cfg)

	// Fetch account
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

	stripe.Key = secrets.Stripe.SecretKey

	// Get or create Stripe customer
	customerID := account.StripeCustomerID
	if customerID == "" {
		cust, err := customer.New(&stripe.CustomerParams{
			Email: stripe.String(authCtx.Email),
			Name:  stripe.String(account.Name),
			Metadata: map[string]string{
				"account_id": authCtx.AccountID,
			},
		})
		if err != nil {
			log.Printf("Failed to create Stripe customer: %v", err)
			return authMiddleware.CreateErrorResponse(500, "Failed to create billing customer"), nil
		}
		customerID = cust.ID

		// Save Stripe customer ID to account
		_, err = dbClient.UpdateItem(ctx, &dynamodb.UpdateItemInput{
			TableName: aws.String(accountsTable),
			Key: map[string]ddbtypes.AttributeValue{
				"account_id": &ddbtypes.AttributeValueMemberS{Value: authCtx.AccountID},
			},
			UpdateExpression: aws.String("SET stripe_customer_id = :cid"),
			ExpressionAttributeValues: map[string]ddbtypes.AttributeValue{
				":cid": &ddbtypes.AttributeValueMemberS{Value: customerID},
			},
		})
		if err != nil {
			log.Printf("Failed to save Stripe customer ID: %v", err)
			// Non-fatal -- proceed with checkout
		}
	}

	baseURL := appURL
	if baseURL == "" {
		baseURL = "https://app.myfusionhelper.ai"
	}
	successURL := baseURL + "/settings/billing/success?session_id={CHECKOUT_SESSION_ID}"
	cancelURL := baseURL + "/settings?tab=billing&billing=cancelled"

	// Allow caller to specify a return URL (e.g., onboarding flow)
	if req.ReturnURL != "" {
		successURL = baseURL + req.ReturnURL + "?session_id={CHECKOUT_SESSION_ID}"
		cancelURL = baseURL + req.ReturnURL + "&billing=cancelled"
	}

	params := &stripe.CheckoutSessionParams{
		Customer: stripe.String(customerID),
		Mode:     stripe.String(string(stripe.CheckoutSessionModeSubscription)),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Price:    stripe.String(priceID),
				Quantity: stripe.Int64(1),
			},
		},
		SuccessURL: stripe.String(successURL),
		CancelURL:  stripe.String(cancelURL),
		SubscriptionData: &stripe.CheckoutSessionSubscriptionDataParams{
			TrialPeriodDays: stripe.Int64(14),
			Metadata: map[string]string{
				"account_id": authCtx.AccountID,
				"plan":       req.Plan,
			},
		},
	}

	s, err := session.New(params)
	if err != nil {
		log.Printf("Failed to create checkout session: %v", err)
		return authMiddleware.CreateErrorResponse(500, "Failed to create checkout session"), nil
	}

	return authMiddleware.CreateSuccessResponse(200, "OK", map[string]string{
		"url":        s.URL,
		"session_id": s.ID,
	}), nil
}

func getPriceID(plan string, billingPeriod string, secrets *appConfig.SecretsConfig) string {
	if billingPeriod == "annual" {
		switch plan {
		case "start":
			return secrets.Stripe.PriceStartAnnual
		case "grow":
			return secrets.Stripe.PriceGrowAnnual
		case "deliver":
			return secrets.Stripe.PriceDeliverAnnual
		default:
			return ""
		}
	}
	switch plan {
	case "start":
		return secrets.Stripe.PriceStart
	case "grow":
		return secrets.Stripe.PriceGrow
	case "deliver":
		return secrets.Stripe.PriceDeliver
	default:
		return ""
	}
}
