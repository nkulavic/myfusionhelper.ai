package getbilling

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
	"github.com/stripe/stripe-go/v82/subscription"
)

var (
	accountsTable = os.Getenv("ACCOUNTS_TABLE")
	stripeKey     = os.Getenv("STRIPE_SECRET_KEY")
)

// BillingResponse is the response for GET /billing
type BillingResponse struct {
	Plan             string                    `json:"plan"`
	Status           string                    `json:"status"`
	PriceMonthly     int                       `json:"price_monthly"`
	RenewsAt         *int64                    `json:"renews_at,omitempty"`
	TrialEndsAt      *int64                    `json:"trial_ends_at,omitempty"`
	CancelAt         *int64                    `json:"cancel_at,omitempty"`
	StripeCustomerID string                    `json:"stripe_customer_id,omitempty"`
	Usage            types.AccountUsage        `json:"usage"`
	Limits           types.AccountSettings     `json:"limits"`
}

// HandleWithAuth returns billing info for the current account
func HandleWithAuth(ctx context.Context, event events.APIGatewayV2HTTPRequest, authCtx *types.AuthContext) (events.APIGatewayV2HTTPResponse, error) {
	log.Printf("GetBilling handler called for account %s", authCtx.AccountID)

	if event.RequestContext.HTTP.Method != "GET" {
		return authMiddleware.CreateErrorResponse(405, "Method not allowed"), nil
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
		return authMiddleware.CreateErrorResponse(500, "Failed to fetch billing info"), nil
	}
	if result.Item == nil {
		return authMiddleware.CreateErrorResponse(404, "Account not found"), nil
	}

	var account types.Account
	if err := attributevalue.UnmarshalMap(result.Item, &account); err != nil {
		log.Printf("Failed to unmarshal account: %v", err)
		return authMiddleware.CreateErrorResponse(500, "Internal error"), nil
	}

	resp := BillingResponse{
		Plan:             account.Plan,
		Status:           account.Status,
		PriceMonthly:     getPlanPrice(account.Plan),
		StripeCustomerID: account.StripeCustomerID,
		Usage:            account.Usage,
		Limits:           account.Settings,
	}

	// If Stripe is configured and customer exists, enrich with subscription data
	if stripeKey != "" && account.StripeCustomerID != "" {
		stripe.Key = stripeKey
		params := &stripe.SubscriptionListParams{}
		params.Customer = stripe.String(account.StripeCustomerID)
		params.Filters.AddFilter("limit", "", "1")

		iter := subscription.List(params)
		if iter.Next() {
			sub := iter.Subscription()
			// In stripe-go v82, CurrentPeriodEnd moved to SubscriptionItem
			if sub.Items != nil && len(sub.Items.Data) > 0 {
				periodEnd := sub.Items.Data[0].CurrentPeriodEnd
				if periodEnd > 0 {
					resp.RenewsAt = &periodEnd
				}
			}
			if sub.TrialEnd > 0 {
				resp.TrialEndsAt = &sub.TrialEnd
			}
			if sub.CancelAt > 0 {
				resp.CancelAt = &sub.CancelAt
			}
		}
	}

	return authMiddleware.CreateSuccessResponse(200, "OK", resp), nil
}

func getPlanPrice(plan string) int {
	switch plan {
	case "start":
		return 39
	case "grow":
		return 59
	case "deliver":
		return 79
	default:
		return 0
	}
}
