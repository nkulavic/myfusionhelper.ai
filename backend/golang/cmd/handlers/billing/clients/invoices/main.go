package invoices

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
	"github.com/stripe/stripe-go/v82/invoice"
)

var (
	accountsTable = os.Getenv("ACCOUNTS_TABLE")
	stripeKey     = os.Getenv("STRIPE_SECRET_KEY")
)

// InvoiceItem is a simplified invoice for the frontend
type InvoiceItem struct {
	ID         string `json:"id"`
	Amount     int64  `json:"amount"`
	Currency   string `json:"currency"`
	Status     string `json:"status"`
	Date       int64  `json:"date"`
	PDFURL     string `json:"pdf_url,omitempty"`
	HostedURL  string `json:"hosted_url,omitempty"`
}

// HandleWithAuth returns invoice history from Stripe
func HandleWithAuth(ctx context.Context, event events.APIGatewayV2HTTPRequest, authCtx *types.AuthContext) (events.APIGatewayV2HTTPResponse, error) {
	log.Printf("Invoices handler called for account %s", authCtx.AccountID)

	if event.RequestContext.HTTP.Method != "GET" {
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
		return authMiddleware.CreateSuccessResponse(200, "OK", []InvoiceItem{}), nil
	}

	stripe.Key = stripeKey
	params := &stripe.InvoiceListParams{}
	params.Customer = stripe.String(account.StripeCustomerID)
	params.Filters.AddFilter("limit", "", "24")

	var items []InvoiceItem
	iter := invoice.List(params)
	for iter.Next() {
		inv := iter.Invoice()
		item := InvoiceItem{
			ID:       inv.ID,
			Amount:   inv.AmountPaid / 100, // cents to dollars
			Currency: string(inv.Currency),
			Status:   string(inv.Status),
			Date:     inv.Created,
		}
		if inv.InvoicePDF != "" {
			item.PDFURL = inv.InvoicePDF
		}
		if inv.HostedInvoiceURL != "" {
			item.HostedURL = inv.HostedInvoiceURL
		}
		items = append(items, item)
	}
	if err := iter.Err(); err != nil {
		log.Printf("Failed to list invoices: %v", err)
		return authMiddleware.CreateErrorResponse(500, "Failed to fetch invoices"), nil
	}

	if items == nil {
		items = []InvoiceItem{}
	}

	return authMiddleware.CreateSuccessResponse(200, "OK", items), nil
}
