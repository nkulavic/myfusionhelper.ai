package webhook

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	ddbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	authMiddleware "github.com/myfusionhelper/api/internal/middleware/auth"
	stripe "github.com/stripe/stripe-go/v82"
	"github.com/stripe/stripe-go/v82/webhook"
)

var (
	accountsTable        = os.Getenv("ACCOUNTS_TABLE")
	stripeWebhookSecret  = os.Getenv("STRIPE_WEBHOOK_SECRET")
)

// Handle processes Stripe webhook events (public, verified by signature)
func Handle(ctx context.Context, event events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	log.Printf("Stripe webhook handler called")

	if event.RequestContext.HTTP.Method != "POST" {
		return authMiddleware.CreateErrorResponse(405, "Method not allowed"), nil
	}

	if stripeWebhookSecret == "" {
		log.Printf("ERROR: STRIPE_WEBHOOK_SECRET not configured")
		return authMiddleware.CreateErrorResponse(500, "Webhook not configured"), nil
	}

	// Verify webhook signature
	sig := event.Headers["stripe-signature"]
	if sig == "" {
		return authMiddleware.CreateErrorResponse(400, "Missing Stripe signature"), nil
	}

	stripeEvent, err := webhook.ConstructEvent([]byte(event.Body), sig, stripeWebhookSecret)
	if err != nil {
		log.Printf("Webhook signature verification failed: %v", err)
		return authMiddleware.CreateErrorResponse(400, "Invalid signature"), nil
	}

	log.Printf("Received Stripe event: %s", stripeEvent.Type)

	switch stripeEvent.Type {
	case "customer.subscription.created",
		"customer.subscription.updated":
		return handleSubscriptionUpdate(ctx, stripeEvent)
	case "customer.subscription.deleted":
		return handleSubscriptionCancelled(ctx, stripeEvent)
	case "invoice.payment_failed":
		return handlePaymentFailed(ctx, stripeEvent)
	default:
		log.Printf("Unhandled event type: %s", stripeEvent.Type)
	}

	return authMiddleware.CreateSuccessResponse(200, "OK", nil), nil
}

func handleSubscriptionUpdate(ctx context.Context, event stripe.Event) (events.APIGatewayV2HTTPResponse, error) {
	var sub stripe.Subscription
	if err := json.Unmarshal(event.Data.Raw, &sub); err != nil {
		log.Printf("Failed to parse subscription: %v", err)
		return authMiddleware.CreateErrorResponse(400, "Invalid event data"), nil
	}

	customerID := sub.Customer.ID
	if customerID == "" {
		log.Printf("No customer ID in subscription event")
		return authMiddleware.CreateSuccessResponse(200, "OK", nil), nil
	}

	// Determine plan from price metadata or product
	plan := "free"
	if len(sub.Items.Data) > 0 {
		item := sub.Items.Data[0]
		if item.Price != nil && item.Price.Metadata != nil {
			if p, ok := item.Price.Metadata["plan"]; ok {
				plan = p
			}
		}
	}

	status := "active"
	if sub.Status == stripe.SubscriptionStatusPastDue {
		status = "active" // still active but payment issue
	} else if sub.Status == stripe.SubscriptionStatusCanceled {
		status = "cancelled"
	} else if sub.Status == stripe.SubscriptionStatusTrialing {
		status = "active"
	}

	if err := updateAccountByStripeCustomer(ctx, customerID, plan, status); err != nil {
		log.Printf("Failed to update account: %v", err)
		return authMiddleware.CreateErrorResponse(500, "Failed to process event"), nil
	}

	return authMiddleware.CreateSuccessResponse(200, "OK", nil), nil
}

func handleSubscriptionCancelled(ctx context.Context, event stripe.Event) (events.APIGatewayV2HTTPResponse, error) {
	var sub stripe.Subscription
	if err := json.Unmarshal(event.Data.Raw, &sub); err != nil {
		log.Printf("Failed to parse subscription: %v", err)
		return authMiddleware.CreateErrorResponse(400, "Invalid event data"), nil
	}

	customerID := sub.Customer.ID
	if customerID == "" {
		return authMiddleware.CreateSuccessResponse(200, "OK", nil), nil
	}

	if err := updateAccountByStripeCustomer(ctx, customerID, "free", "active"); err != nil {
		log.Printf("Failed to downgrade account: %v", err)
		return authMiddleware.CreateErrorResponse(500, "Failed to process event"), nil
	}

	return authMiddleware.CreateSuccessResponse(200, "OK", nil), nil
}

func handlePaymentFailed(ctx context.Context, event stripe.Event) (events.APIGatewayV2HTTPResponse, error) {
	log.Printf("Payment failed event received -- account status unchanged, Stripe will retry")
	// Stripe handles retry logic. We just log it. The subscription.updated event
	// will fire if the subscription status changes to past_due or cancelled.
	return authMiddleware.CreateSuccessResponse(200, "OK", nil), nil
}

func updateAccountByStripeCustomer(ctx context.Context, stripeCustomerID, plan, status string) error {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return err
	}

	dbClient := dynamodb.NewFromConfig(cfg)

	// Query accounts by stripe_customer_id using a scan with filter
	// In production, you'd add a GSI on stripe_customer_id. For now, we use the
	// fact that we know the account_id from the subscription metadata.
	// This is a simplified approach -- in a full implementation, store account_id
	// in Stripe subscription metadata during checkout creation.

	// For now, scan with filter (acceptable for low volume)
	scanResult, err := dbClient.Scan(ctx, &dynamodb.ScanInput{
		TableName:        aws.String(accountsTable),
		FilterExpression: aws.String("stripe_customer_id = :cid"),
		ExpressionAttributeValues: map[string]ddbtypes.AttributeValue{
			":cid": &ddbtypes.AttributeValueMemberS{Value: stripeCustomerID},
		},
		Limit: aws.Int32(1),
	})
	if err != nil {
		return err
	}
	if len(scanResult.Items) == 0 {
		log.Printf("No account found for Stripe customer %s", stripeCustomerID)
		return nil
	}

	// Get the account_id from the result
	accountIDAttr, ok := scanResult.Items[0]["account_id"]
	if !ok {
		log.Printf("Account missing account_id field")
		return nil
	}
	accountID := accountIDAttr.(*ddbtypes.AttributeValueMemberS).Value

	limits := getPlanLimits(plan)

	_, err = dbClient.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(accountsTable),
		Key: map[string]ddbtypes.AttributeValue{
			"account_id": &ddbtypes.AttributeValueMemberS{Value: accountID},
		},
		UpdateExpression: aws.String("SET #plan = :plan, #status = :status, settings.max_helpers = :mh, settings.max_connections = :mc, settings.max_executions = :me, updated_at = :now"),
		ExpressionAttributeNames: map[string]string{
			"#plan":   "plan",
			"#status": "status",
		},
		ExpressionAttributeValues: map[string]ddbtypes.AttributeValue{
			":plan":   &ddbtypes.AttributeValueMemberS{Value: plan},
			":status": &ddbtypes.AttributeValueMemberS{Value: status},
			":mh":     &ddbtypes.AttributeValueMemberN{Value: intToStr(limits.MaxHelpers)},
			":mc":     &ddbtypes.AttributeValueMemberN{Value: intToStr(limits.MaxConnections)},
			":me":     &ddbtypes.AttributeValueMemberN{Value: intToStr(limits.MaxExecutions)},
			":now":    &ddbtypes.AttributeValueMemberS{Value: time.Now().Format(time.RFC3339)},
		},
	})
	return err
}

type planLimits struct {
	MaxHelpers     int
	MaxConnections int
	MaxExecutions  int
}

func getPlanLimits(plan string) planLimits {
	switch plan {
	case "start":
		return planLimits{MaxHelpers: 10, MaxConnections: 2, MaxExecutions: 10000}
	case "grow":
		return planLimits{MaxHelpers: 50, MaxConnections: 5, MaxExecutions: 50000}
	case "deliver":
		return planLimits{MaxHelpers: 999999, MaxConnections: 20, MaxExecutions: 500000}
	default: // free
		return planLimits{MaxHelpers: 3, MaxConnections: 1, MaxExecutions: 1000}
	}
}

func intToStr(n int) string {
	return fmt.Sprintf("%d", n)
}
