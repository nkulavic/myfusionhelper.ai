package webhook

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	ddbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/myfusionhelper/api/internal/apiutil"
	"github.com/myfusionhelper/api/internal/billing"
	appConfig "github.com/myfusionhelper/api/internal/config"
	authMiddleware "github.com/myfusionhelper/api/internal/middleware/auth"
	"github.com/myfusionhelper/api/internal/notifications"
	stripe "github.com/stripe/stripe-go/v82"
	"github.com/stripe/stripe-go/v82/checkout/session"
	stripeSub "github.com/stripe/stripe-go/v82/subscription"
	"github.com/stripe/stripe-go/v82/webhook"
)

var (
	accountsTable      = os.Getenv("ACCOUNTS_TABLE")
	usersTable         = os.Getenv("USERS_TABLE")
	webhookEventsTable = os.Getenv("WEBHOOK_EVENTS_TABLE")
	cognitoUserPoolID  = os.Getenv("COGNITO_USER_POOL_ID")
)

// planRank maps plan names to numeric ranks for upgrade/downgrade detection.
var planRank = map[string]int{
	"free": 0, "trial": 0, "start": 1, "grow": 2, "deliver": 3,
}

// Handle processes Stripe webhook events (public, verified by signature)
func Handle(ctx context.Context, event events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	log.Printf("Stripe webhook handler called")

	if event.RequestContext.HTTP.Method != "POST" {
		return authMiddleware.CreateErrorResponse(405, "Method not allowed"), nil
	}

	secrets, err := appConfig.LoadSecrets(ctx)
	if err != nil {
		log.Printf("Failed to load secrets: %v", err)
		return authMiddleware.CreateErrorResponse(500, "Config error"), nil
	}

	if secrets.Stripe.WebhookSecret == "" {
		log.Printf("ERROR: Stripe webhook secret not configured")
		return authMiddleware.CreateErrorResponse(500, "Webhook not configured"), nil
	}

	// Verify webhook signature
	sig := event.Headers["stripe-signature"]
	if sig == "" {
		return authMiddleware.CreateErrorResponse(400, "Missing Stripe signature"), nil
	}

	stripeEvent, err := webhook.ConstructEvent([]byte(apiutil.GetBody(event)), sig, secrets.Stripe.WebhookSecret)
	if err != nil {
		log.Printf("Webhook signature verification failed: %v", err)
		return authMiddleware.CreateErrorResponse(400, "Invalid signature"), nil
	}

	stripeKey := secrets.Stripe.SecretKey

	log.Printf("Received Stripe event: %s (id: %s)", stripeEvent.Type, stripeEvent.ID)

	// Create shared DynamoDB client for all handlers
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		log.Printf("Failed to load AWS config: %v", err)
		return authMiddleware.CreateErrorResponse(500, "Internal error"), nil
	}
	dbClient := dynamodb.NewFromConfig(cfg)

	// Idempotency check -- skip if already processed
	alreadySeen, err := checkAndRecordEvent(ctx, dbClient, stripeEvent)
	if err != nil {
		log.Printf("Idempotency check failed: %v", err)
		// Continue processing -- better to risk a duplicate than to drop the event
	}
	if alreadySeen {
		log.Printf("Event %s already processed, skipping", stripeEvent.ID)
		return authMiddleware.CreateSuccessResponse(200, "OK", nil), nil
	}

	// Route to handler
	var handlerErr error
	switch stripeEvent.Type {
	case "customer.subscription.created":
		handlerErr = handleSubscriptionCreated(ctx, dbClient, stripeEvent)
	case "customer.subscription.updated":
		handlerErr = handleSubscriptionUpdated(ctx, dbClient, stripeEvent)
	case "customer.subscription.deleted":
		handlerErr = handleSubscriptionCancelled(ctx, dbClient, stripeEvent)
	case "customer.subscription.trial_will_end":
		handlerErr = handleTrialWillEnd(ctx, dbClient, stripeEvent)
	case "checkout.session.completed":
		handlerErr = handleCheckoutSessionCompleted(ctx, dbClient, stripeEvent, stripeKey)
	case "invoice.payment_failed":
		handlerErr = handlePaymentFailed(ctx, dbClient, stripeEvent)
	case "invoice.paid":
		handlerErr = handleInvoicePaid(ctx, dbClient, stripeEvent)
	case "charge.refunded":
		handlerErr = handleChargeRefunded(ctx, dbClient, stripeEvent)
	default:
		log.Printf("Unhandled event type: %s", stripeEvent.Type)
	}

	// Update event status
	if handlerErr != nil {
		log.Printf("Handler error for event %s: %v", stripeEvent.ID, handlerErr)
		markEventStatus(ctx, dbClient, stripeEvent.ID, "failed", handlerErr.Error())
		return authMiddleware.CreateErrorResponse(500, "Failed to process event"), nil
	}

	markEventStatus(ctx, dbClient, stripeEvent.ID, "processed", "")
	return authMiddleware.CreateSuccessResponse(200, "OK", nil), nil
}

// ---------------------------------------------------------------------------
// Idempotency layer
// ---------------------------------------------------------------------------

// checkAndRecordEvent attempts to insert the event into the webhook-events table.
// Returns (true, nil) if the event was already recorded (duplicate).
func checkAndRecordEvent(ctx context.Context, dbClient *dynamodb.Client, event stripe.Event) (bool, error) {
	if webhookEventsTable == "" {
		return false, nil // table not configured, skip idempotency
	}

	now := time.Now().UTC()
	ttl := now.Add(90 * 24 * time.Hour).Unix() // 90-day retention

	_, err := dbClient.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(webhookEventsTable),
		Item: map[string]ddbtypes.AttributeValue{
			"event_id":            &ddbtypes.AttributeValueMemberS{Value: event.ID},
			"event_type":          &ddbtypes.AttributeValueMemberS{Value: string(event.Type)},
			"status":              &ddbtypes.AttributeValueMemberS{Value: "pending"},
			"received_at":         &ddbtypes.AttributeValueMemberS{Value: now.Format(time.RFC3339)},
			"ttl":                 &ddbtypes.AttributeValueMemberN{Value: fmt.Sprintf("%d", ttl)},
		},
		ConditionExpression: aws.String("attribute_not_exists(event_id)"),
	})
	if err != nil {
		var condErr *ddbtypes.ConditionalCheckFailedException
		if errors.As(err, &condErr) {
			return true, nil // already exists
		}
		return false, fmt.Errorf("failed to record event: %w", err)
	}

	return false, nil
}

// markEventStatus updates the processing status of a recorded webhook event.
func markEventStatus(ctx context.Context, dbClient *dynamodb.Client, eventID string, status, errorMsg string) {
	if webhookEventsTable == "" {
		return
	}

	now := time.Now().UTC().Format(time.RFC3339)
	updateExpr := "SET #s = :status, processed_at = :now"
	exprValues := map[string]ddbtypes.AttributeValue{
		":status": &ddbtypes.AttributeValueMemberS{Value: status},
		":now":    &ddbtypes.AttributeValueMemberS{Value: now},
	}

	if errorMsg != "" {
		updateExpr += ", error_message = :err"
		exprValues[":err"] = &ddbtypes.AttributeValueMemberS{Value: errorMsg}
	}

	_, err := dbClient.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(webhookEventsTable),
		Key: map[string]ddbtypes.AttributeValue{
			"event_id": &ddbtypes.AttributeValueMemberS{Value: eventID},
		},
		UpdateExpression: aws.String(updateExpr),
		ExpressionAttributeNames: map[string]string{
			"#s": "status",
		},
		ExpressionAttributeValues: exprValues,
	})
	if err != nil {
		log.Printf("Failed to update event status for %s: %v", eventID, err)
	}
}

// ---------------------------------------------------------------------------
// Shared helpers
// ---------------------------------------------------------------------------

// accountLookupResult holds the result of looking up an account by Stripe customer ID.
type accountLookupResult struct {
	AccountID   string
	OwnerUserID string
	CurrentPlan string
	Status      string
}

// lookupAccountByStripeCustomer queries the StripeCustomerIdIndex GSI.
func lookupAccountByStripeCustomer(ctx context.Context, dbClient *dynamodb.Client, customerID string) (*accountLookupResult, error) {
	queryResult, err := dbClient.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(accountsTable),
		IndexName:              aws.String("StripeCustomerIdIndex"),
		KeyConditionExpression: aws.String("stripe_customer_id = :cid"),
		ExpressionAttributeValues: map[string]ddbtypes.AttributeValue{
			":cid": &ddbtypes.AttributeValueMemberS{Value: customerID},
		},
		Limit: aws.Int32(1),
	})
	if err != nil {
		return nil, err
	}
	if len(queryResult.Items) == 0 {
		log.Printf("No account found for Stripe customer %s", customerID)
		return nil, nil
	}

	result := &accountLookupResult{}
	if attr, ok := queryResult.Items[0]["account_id"]; ok {
		result.AccountID = attr.(*ddbtypes.AttributeValueMemberS).Value
	}
	if attr, ok := queryResult.Items[0]["owner_user_id"]; ok {
		result.OwnerUserID = attr.(*ddbtypes.AttributeValueMemberS).Value
	}
	if attr, ok := queryResult.Items[0]["plan"]; ok {
		result.CurrentPlan = attr.(*ddbtypes.AttributeValueMemberS).Value
	}
	if attr, ok := queryResult.Items[0]["status"]; ok {
		result.Status = attr.(*ddbtypes.AttributeValueMemberS).Value
	}

	return result, nil
}

// updateAccountPlan updates the plan, limits, and status for an account.
func updateAccountPlan(ctx context.Context, dbClient *dynamodb.Client, accountID, plan, status string) error {
	planCfg := billing.GetPlan(plan)

	updateExpr := "SET #plan = :plan, #status = :status, settings.max_helpers = :mh, settings.max_connections = :mc, settings.max_executions = :me, updated_at = :now"
	exprValues := map[string]ddbtypes.AttributeValue{
		":plan":   &ddbtypes.AttributeValueMemberS{Value: plan},
		":status": &ddbtypes.AttributeValueMemberS{Value: status},
		":mh":     &ddbtypes.AttributeValueMemberN{Value: intToStr(planCfg.MaxHelpers)},
		":mc":     &ddbtypes.AttributeValueMemberN{Value: intToStr(planCfg.MaxConnections)},
		":me":     &ddbtypes.AttributeValueMemberN{Value: intToStr(planCfg.MaxExecutions)},
		":now":    &ddbtypes.AttributeValueMemberS{Value: time.Now().Format(time.RFC3339)},
	}

	// Clear trial_expired when activating a paid subscription
	if billing.IsPaidPlan(plan) {
		updateExpr += ", trial_expired = :te"
		exprValues[":te"] = &ddbtypes.AttributeValueMemberBOOL{Value: false}
	}

	_, err := dbClient.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(accountsTable),
		Key: map[string]ddbtypes.AttributeValue{
			"account_id": &ddbtypes.AttributeValueMemberS{Value: accountID},
		},
		UpdateExpression: aws.String(updateExpr),
		ExpressionAttributeNames: map[string]string{
			"#plan":   "plan",
			"#status": "status",
		},
		ExpressionAttributeValues: exprValues,
	})
	return err
}

// classifyPlanChange determines if a plan change is an upgrade, downgrade, or same-tier.
func classifyPlanChange(oldPlan, newPlan string) string {
	oldRank := planRank[oldPlan]
	newRank := planRank[newPlan]

	if newRank > oldRank {
		return "plan_upgraded"
	}
	if newRank < oldRank {
		return "plan_downgraded"
	}
	return "" // same tier, no email
}

// extractPlanFromSubscription determines the plan name from subscription metadata/price metadata.
func extractPlanFromSubscription(sub *stripe.Subscription) string {
	plan := "free"
	if sub.Metadata != nil {
		if p, ok := sub.Metadata["plan"]; ok && p != "" {
			plan = p
		}
	}
	if plan == "free" && len(sub.Items.Data) > 0 {
		item := sub.Items.Data[0]
		if item.Price != nil && item.Price.Metadata != nil {
			if p, ok := item.Price.Metadata["plan"]; ok && p != "" {
				plan = p
			}
		}
	}
	return plan
}

// subscriptionStatusToAccountStatus maps Stripe subscription status to our account status.
func subscriptionStatusToAccountStatus(status stripe.SubscriptionStatus) string {
	switch status {
	case stripe.SubscriptionStatusCanceled:
		return "cancelled"
	case stripe.SubscriptionStatusPastDue:
		return "past_due"
	default:
		return "active"
	}
}

// ---------------------------------------------------------------------------
// Event handlers
// ---------------------------------------------------------------------------

// handleSubscriptionCreated handles customer.subscription.created events.
// Skips the email if checkout.session.completed already sent one (race condition fix).
func handleSubscriptionCreated(ctx context.Context, dbClient *dynamodb.Client, event stripe.Event) error {
	var sub stripe.Subscription
	if err := json.Unmarshal(event.Data.Raw, &sub); err != nil {
		return fmt.Errorf("failed to parse subscription: %w", err)
	}

	customerID := sub.Customer.ID
	if customerID == "" {
		log.Printf("No customer ID in subscription.created event")
		return nil
	}

	plan := extractPlanFromSubscription(&sub)
	status := subscriptionStatusToAccountStatus(sub.Status)

	acct, err := lookupAccountByStripeCustomer(ctx, dbClient, customerID)
	if err != nil {
		return fmt.Errorf("failed to look up account: %w", err)
	}
	if acct == nil {
		return nil
	}

	if err := updateAccountPlan(ctx, dbClient, acct.AccountID, plan, status); err != nil {
		return fmt.Errorf("failed to update account: %w", err)
	}

	if acct.OwnerUserID != "" {
		syncCognitoPlanGroup(ctx, dbClient, acct.OwnerUserID, plan)

		// Only send email if checkout hasn't already sent one.
		// Check for subscription_email_sent flag on the account.
		emailAlreadySent := checkAndClearEmailSentFlag(ctx, dbClient, acct.AccountID)
		if !emailAlreadySent {
			sendBillingEmail(ctx, dbClient, acct.OwnerUserID, "subscription_created", plan)
		} else {
			log.Printf("Skipping subscription_created email -- already sent by checkout handler")
		}
	}

	return nil
}

// handleSubscriptionUpdated handles customer.subscription.updated events.
// Detects upgrade vs downgrade and sends the correct email.
func handleSubscriptionUpdated(ctx context.Context, dbClient *dynamodb.Client, event stripe.Event) error {
	var sub stripe.Subscription
	if err := json.Unmarshal(event.Data.Raw, &sub); err != nil {
		return fmt.Errorf("failed to parse subscription: %w", err)
	}

	customerID := sub.Customer.ID
	if customerID == "" {
		log.Printf("No customer ID in subscription.updated event")
		return nil
	}

	plan := extractPlanFromSubscription(&sub)
	status := subscriptionStatusToAccountStatus(sub.Status)

	acct, err := lookupAccountByStripeCustomer(ctx, dbClient, customerID)
	if err != nil {
		return fmt.Errorf("failed to look up account: %w", err)
	}
	if acct == nil {
		return nil
	}

	previousPlan := acct.CurrentPlan

	if err := updateAccountPlan(ctx, dbClient, acct.AccountID, plan, status); err != nil {
		return fmt.Errorf("failed to update account: %w", err)
	}

	if acct.OwnerUserID != "" {
		syncCognitoPlanGroup(ctx, dbClient, acct.OwnerUserID, plan)

		// Determine correct email type based on plan change direction
		emailType := classifyPlanChange(previousPlan, plan)
		if emailType != "" {
			sendBillingEmail(ctx, dbClient, acct.OwnerUserID, emailType, plan)
		} else {
			log.Printf("Subscription updated but plan unchanged (%s), no email sent", plan)
		}
	}

	return nil
}

// handleSubscriptionCancelled handles customer.subscription.deleted events.
func handleSubscriptionCancelled(ctx context.Context, dbClient *dynamodb.Client, event stripe.Event) error {
	var sub stripe.Subscription
	if err := json.Unmarshal(event.Data.Raw, &sub); err != nil {
		return fmt.Errorf("failed to parse subscription: %w", err)
	}

	customerID := sub.Customer.ID
	if customerID == "" {
		return nil
	}

	acct, err := lookupAccountByStripeCustomer(ctx, dbClient, customerID)
	if err != nil {
		return fmt.Errorf("failed to look up account: %w", err)
	}
	if acct == nil {
		return nil
	}

	// Downgrade to trial plan
	if err := updateAccountPlan(ctx, dbClient, acct.AccountID, "trial", "active"); err != nil {
		return fmt.Errorf("failed to downgrade account: %w", err)
	}

	// Mark trial as expired since they had a subscription before
	setTrialExpired(ctx, dbClient, acct.AccountID)

	if acct.OwnerUserID != "" {
		syncCognitoPlanGroup(ctx, dbClient, acct.OwnerUserID, "trial")
		sendBillingEmail(ctx, dbClient, acct.OwnerUserID, "subscription_cancelled", "trial")
	}

	return nil
}

// handleTrialWillEnd handles customer.subscription.trial_will_end events (3 days before end).
func handleTrialWillEnd(ctx context.Context, dbClient *dynamodb.Client, event stripe.Event) error {
	var sub stripe.Subscription
	if err := json.Unmarshal(event.Data.Raw, &sub); err != nil {
		return fmt.Errorf("failed to parse subscription: %w", err)
	}

	customerID := sub.Customer.ID
	if customerID == "" {
		return nil
	}

	acct, err := lookupAccountByStripeCustomer(ctx, dbClient, customerID)
	if err != nil {
		return fmt.Errorf("failed to look up account: %w", err)
	}
	if acct == nil {
		return nil
	}

	// No data changes, just send the email
	if acct.OwnerUserID != "" {
		sendBillingEmail(ctx, dbClient, acct.OwnerUserID, "trial_ending", acct.CurrentPlan)
	}

	return nil
}

// handlePaymentFailed handles invoice.payment_failed events.
func handlePaymentFailed(ctx context.Context, dbClient *dynamodb.Client, event stripe.Event) error {
	var invoice stripe.Invoice
	if err := json.Unmarshal(event.Data.Raw, &invoice); err != nil {
		return fmt.Errorf("failed to parse invoice: %w", err)
	}

	customerID := ""
	if invoice.Customer != nil {
		customerID = invoice.Customer.ID
	}
	if customerID == "" {
		log.Printf("Payment failed event with no customer ID")
		return nil
	}

	acct, err := lookupAccountByStripeCustomer(ctx, dbClient, customerID)
	if err != nil {
		return fmt.Errorf("failed to look up account: %w", err)
	}
	if acct == nil {
		return nil
	}

	// No account status change -- Stripe handles retry logic.
	// Send payment_failed email with direct link to pay the invoice.
	if acct.OwnerUserID != "" {
		var extraData map[string]interface{}
		if invoice.HostedInvoiceURL != "" {
			extraData = map[string]interface{}{
				"InvoiceURL": invoice.HostedInvoiceURL,
			}
		}
		sendBillingEmail(ctx, dbClient, acct.OwnerUserID, "payment_failed", acct.CurrentPlan, extraData)
	}

	log.Printf("Payment failed for customer %s -- email sent, Stripe will retry", customerID)
	return nil
}

// handleInvoicePaid handles invoice.paid events.
// Resets past_due accounts back to active, and sends a payment receipt for all paid invoices.
func handleInvoicePaid(ctx context.Context, dbClient *dynamodb.Client, event stripe.Event) error {
	var invoice stripe.Invoice
	if err := json.Unmarshal(event.Data.Raw, &invoice); err != nil {
		return fmt.Errorf("failed to parse invoice: %w", err)
	}

	customerID := ""
	if invoice.Customer != nil {
		customerID = invoice.Customer.ID
	}
	if customerID == "" {
		return nil
	}

	acct, err := lookupAccountByStripeCustomer(ctx, dbClient, customerID)
	if err != nil {
		return fmt.Errorf("failed to look up account: %w", err)
	}
	if acct == nil {
		return nil
	}

	// Reset past_due accounts back to active
	if acct.Status == "past_due" {
		_, err := dbClient.UpdateItem(ctx, &dynamodb.UpdateItemInput{
			TableName: aws.String(accountsTable),
			Key: map[string]ddbtypes.AttributeValue{
				"account_id": &ddbtypes.AttributeValueMemberS{Value: acct.AccountID},
			},
			UpdateExpression: aws.String("SET #status = :active, updated_at = :now"),
			ExpressionAttributeNames: map[string]string{
				"#status": "status",
			},
			ExpressionAttributeValues: map[string]ddbtypes.AttributeValue{
				":active": &ddbtypes.AttributeValueMemberS{Value: "active"},
				":now":    &ddbtypes.AttributeValueMemberS{Value: time.Now().Format(time.RFC3339)},
			},
		})
		if err != nil {
			return fmt.Errorf("failed to reset account status: %w", err)
		}

		log.Printf("Reset account %s from past_due to active", acct.AccountID)

		if acct.OwnerUserID != "" {
			sendBillingEmail(ctx, dbClient, acct.OwnerUserID, "payment_recovered", acct.CurrentPlan)
		}
		return nil
	}

	// For all other paid invoices, send a payment receipt
	if acct.OwnerUserID != "" && invoice.AmountPaid > 0 {
		extraData := map[string]interface{}{
			"Amount":        formatStripeAmount(invoice.AmountPaid, string(invoice.Currency)),
			"InvoiceNumber": invoice.Number,
		}
		if invoice.HostedInvoiceURL != "" {
			extraData["InvoiceURL"] = invoice.HostedInvoiceURL
		}
		sendBillingEmail(ctx, dbClient, acct.OwnerUserID, "payment_receipt", acct.CurrentPlan, extraData)
		log.Printf("Sent payment receipt for invoice %s (amount: %d %s)", invoice.Number, invoice.AmountPaid, invoice.Currency)
	}

	return nil
}

// handleChargeRefunded handles charge.refunded events.
// Sends a refund confirmation email to the account owner.
func handleChargeRefunded(ctx context.Context, dbClient *dynamodb.Client, event stripe.Event) error {
	var charge stripe.Charge
	if err := json.Unmarshal(event.Data.Raw, &charge); err != nil {
		return fmt.Errorf("failed to parse charge: %w", err)
	}

	customerID := ""
	if charge.Customer != nil {
		customerID = charge.Customer.ID
	}
	if customerID == "" {
		log.Printf("Charge refunded event with no customer ID")
		return nil
	}

	acct, err := lookupAccountByStripeCustomer(ctx, dbClient, customerID)
	if err != nil {
		return fmt.Errorf("failed to look up account: %w", err)
	}
	if acct == nil {
		return nil
	}

	if acct.OwnerUserID != "" {
		refundAmount := charge.AmountRefunded
		refundReason := ""

		// Get reason from the most recent refund if available
		if charge.Refunds != nil && len(charge.Refunds.Data) > 0 {
			latestRefund := charge.Refunds.Data[0]
			if refundAmount == 0 {
				refundAmount = latestRefund.Amount
			}
			if latestRefund.Reason != "" {
				refundReason = formatRefundReason(string(latestRefund.Reason))
			}
		}

		extraData := map[string]interface{}{
			"Amount": formatStripeAmount(refundAmount, string(charge.Currency)),
		}
		if refundReason != "" {
			extraData["RefundReason"] = refundReason
		}

		sendBillingEmail(ctx, dbClient, acct.OwnerUserID, "refund_processed", acct.CurrentPlan, extraData)
		log.Printf("Sent refund_processed email for customer %s (amount: %d %s)",
			customerID, refundAmount, charge.Currency)
	}

	return nil
}

// formatStripeAmount converts Stripe's integer cents to a formatted dollar string.
func formatStripeAmount(amountCents int64, currency string) string {
	dollars := float64(amountCents) / 100.0
	symbol := "$"
	switch currency {
	case "eur":
		symbol = "\u20ac"
	case "gbp":
		symbol = "\u00a3"
	}
	return fmt.Sprintf("%s%.2f", symbol, dollars)
}

// formatRefundReason converts Stripe refund reason codes to human-readable text.
func formatRefundReason(reason string) string {
	switch reason {
	case "duplicate":
		return "Duplicate charge"
	case "fraudulent":
		return "Fraudulent charge"
	case "requested_by_customer":
		return "Requested by customer"
	default:
		return reason
	}
}

// handleCheckoutSessionCompleted handles checkout.session.completed events.
func handleCheckoutSessionCompleted(ctx context.Context, dbClient *dynamodb.Client, event stripe.Event, stripeKey string) error {
	var sess stripe.CheckoutSession
	if err := json.Unmarshal(event.Data.Raw, &sess); err != nil {
		return fmt.Errorf("failed to parse checkout session: %w", err)
	}

	log.Printf("Checkout session completed: %s, mode: %s", sess.ID, sess.Mode)

	if sess.Mode != stripe.CheckoutSessionModeSubscription {
		log.Printf("Ignoring non-subscription checkout session")
		return nil
	}

	// Extract account_id and plan from session metadata or subscription metadata
	accountID := ""
	plan := ""

	if sess.Metadata != nil {
		accountID = sess.Metadata["account_id"]
		plan = sess.Metadata["plan"]
	}

	// If not on the session itself, retrieve subscription to check its metadata
	if (accountID == "" || plan == "") && sess.Subscription != nil && sess.Subscription.ID != "" {
		stripe.Key = stripeKey
		expandedSess, err := session.Get(sess.ID, &stripe.CheckoutSessionParams{
			Params: stripe.Params{
				Expand: []*string{stripe.String("subscription")},
			},
		})
		if err != nil {
			log.Printf("Failed to retrieve checkout session with expansion: %v", err)
		} else if expandedSess.Subscription != nil && expandedSess.Subscription.Metadata != nil {
			if accountID == "" {
				accountID = expandedSess.Subscription.Metadata["account_id"]
			}
			if plan == "" {
				plan = expandedSess.Subscription.Metadata["plan"]
			}
		}
	}

	if accountID == "" {
		log.Printf("No account_id found in checkout session metadata, falling back to customer lookup")
		return nil
	}

	if plan == "" {
		plan = "start" // default
	}

	log.Printf("Activating subscription for account %s, plan: %s", accountID, plan)

	planCfg := billing.GetPlan(plan)

	// Activate the account -- set plan, limits, clear trial_expired, and set email_sent flag
	_, err := dbClient.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(accountsTable),
		Key: map[string]ddbtypes.AttributeValue{
			"account_id": &ddbtypes.AttributeValueMemberS{Value: accountID},
		},
		UpdateExpression: aws.String("SET #plan = :plan, #status = :status, settings.max_helpers = :mh, settings.max_connections = :mc, settings.max_executions = :me, trial_expired = :te, subscription_email_sent = :emailSent, updated_at = :now"),
		ExpressionAttributeNames: map[string]string{
			"#plan":   "plan",
			"#status": "status",
		},
		ExpressionAttributeValues: map[string]ddbtypes.AttributeValue{
			":plan":      &ddbtypes.AttributeValueMemberS{Value: plan},
			":status":    &ddbtypes.AttributeValueMemberS{Value: "active"},
			":mh":        &ddbtypes.AttributeValueMemberN{Value: intToStr(planCfg.MaxHelpers)},
			":mc":        &ddbtypes.AttributeValueMemberN{Value: intToStr(planCfg.MaxConnections)},
			":me":        &ddbtypes.AttributeValueMemberN{Value: intToStr(planCfg.MaxExecutions)},
			":te":        &ddbtypes.AttributeValueMemberBOOL{Value: false},
			":emailSent": &ddbtypes.AttributeValueMemberBOOL{Value: true},
			":now":       &ddbtypes.AttributeValueMemberS{Value: time.Now().Format(time.RFC3339)},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to activate account %s: %w", accountID, err)
	}

	log.Printf("Successfully activated account %s on plan %s", accountID, plan)

	// Extract and store stripe_metered_item_id from subscription items
	if sess.Subscription != nil && sess.Subscription.ID != "" {
		storeMeteredItemID(ctx, dbClient, accountID, sess.Subscription.ID, stripeKey)
	}

	// Send welcome/activation email
	ownerUserID := getAccountOwnerUserID(ctx, dbClient, accountID)
	if ownerUserID != "" {
		syncCognitoPlanGroup(ctx, dbClient, ownerUserID, plan)
		sendBillingEmail(ctx, dbClient, ownerUserID, "subscription_created", plan)
	}

	return nil
}

// ---------------------------------------------------------------------------
// Helper functions
// ---------------------------------------------------------------------------

// getAccountOwnerUserID fetches the owner_user_id for an account.
func getAccountOwnerUserID(ctx context.Context, dbClient *dynamodb.Client, accountID string) string {
	result, err := dbClient.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(accountsTable),
		Key: map[string]ddbtypes.AttributeValue{
			"account_id": &ddbtypes.AttributeValueMemberS{Value: accountID},
		},
		ProjectionExpression: aws.String("owner_user_id"),
	})
	if err != nil || result.Item == nil {
		return ""
	}
	if attr, ok := result.Item["owner_user_id"]; ok {
		return attr.(*ddbtypes.AttributeValueMemberS).Value
	}
	return ""
}

// checkAndClearEmailSentFlag checks if the subscription_email_sent flag is set
// on the account and clears it. Returns true if it was set.
func checkAndClearEmailSentFlag(ctx context.Context, dbClient *dynamodb.Client, accountID string) bool {
	result, err := dbClient.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(accountsTable),
		Key: map[string]ddbtypes.AttributeValue{
			"account_id": &ddbtypes.AttributeValueMemberS{Value: accountID},
		},
		UpdateExpression:    aws.String("REMOVE subscription_email_sent"),
		ConditionExpression: aws.String("subscription_email_sent = :true"),
		ExpressionAttributeValues: map[string]ddbtypes.AttributeValue{
			":true": &ddbtypes.AttributeValueMemberBOOL{Value: true},
		},
		ReturnValues: ddbtypes.ReturnValueAllOld,
	})
	if err != nil {
		// Condition failed = flag wasn't set, or other error
		return false
	}
	// If we got here, the condition passed and the flag was cleared
	_ = result
	return true
}

func intToStr(n int) string {
	return fmt.Sprintf("%d", n)
}

// sendBillingEmail looks up the account owner and sends a billing notification email.
// Optional extraData is forwarded to the notification service (e.g., InvoiceURL).
func sendBillingEmail(ctx context.Context, dbClient *dynamodb.Client, ownerUserID, eventType, plan string, extraData ...map[string]interface{}) {
	if usersTable == "" {
		log.Printf("USERS_TABLE not set, skipping billing email")
		return
	}

	// Look up owner user to get their email
	userResult, err := dbClient.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(usersTable),
		Key: map[string]ddbtypes.AttributeValue{
			"user_id": &ddbtypes.AttributeValueMemberS{Value: ownerUserID},
		},
		ProjectionExpression: aws.String("#n, email"),
		ExpressionAttributeNames: map[string]string{
			"#n": "name",
		},
	})
	if err != nil {
		log.Printf("Failed to look up user %s for billing email: %v", ownerUserID, err)
		return
	}
	if userResult.Item == nil {
		log.Printf("User %s not found, skipping billing email", ownerUserID)
		return
	}

	emailAttr, ok := userResult.Item["email"]
	if !ok {
		return
	}
	userEmail := emailAttr.(*ddbtypes.AttributeValueMemberS).Value

	userName := "there"
	if nameAttr, ok := userResult.Item["name"]; ok {
		userName = nameAttr.(*ddbtypes.AttributeValueMemberS).Value
	}

	notifSvc, err := notifications.New(ctx)
	if err != nil {
		log.Printf("Failed to create notification service: %v", err)
		return
	}

	if err := notifSvc.SendBillingEvent(ctx, userName, userEmail, eventType, billing.GetPlanLabel(plan), extraData...); err != nil {
		log.Printf("Failed to send %s billing email to %s: %v", eventType, userEmail, err)
	} else {
		log.Printf("Sent %s billing email to %s (plan: %s)", eventType, userEmail, plan)
	}
}

// syncCognitoPlanGroup updates a user's Cognito plan group.
// Removes from all plan-* groups, then adds to the new plan group.
func syncCognitoPlanGroup(ctx context.Context, dbClient *dynamodb.Client, ownerUserID, newPlan string) {
	if cognitoUserPoolID == "" {
		log.Printf("COGNITO_USER_POOL_ID not set, skipping Cognito plan group sync")
		return
	}

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		log.Printf("Failed to load AWS config for Cognito sync: %v", err)
		return
	}
	cognitoClient := cognitoidentityprovider.NewFromConfig(cfg)

	// Look up cognito_user_id from DynamoDB user record
	userResult, err := dbClient.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(usersTable),
		Key: map[string]ddbtypes.AttributeValue{
			"user_id": &ddbtypes.AttributeValueMemberS{Value: ownerUserID},
		},
	})
	if err != nil || userResult.Item == nil {
		log.Printf("Failed to get user %s for Cognito sync: %v", ownerUserID, err)
		return
	}

	var user struct {
		CognitoUserID string `dynamodbav:"cognito_user_id"`
	}
	if err := attributevalue.UnmarshalMap(userResult.Item, &user); err != nil || user.CognitoUserID == "" {
		log.Printf("Failed to get cognito_user_id for user %s: %v", ownerUserID, err)
		return
	}

	cognitoUsername := user.CognitoUserID

	// Remove from all plan groups
	planGroups := []string{"plan-free", "plan-trial", "plan-start", "plan-grow", "plan-deliver"}
	for _, group := range planGroups {
		_, _ = cognitoClient.AdminRemoveUserFromGroup(ctx, &cognitoidentityprovider.AdminRemoveUserFromGroupInput{
			UserPoolId: aws.String(cognitoUserPoolID),
			Username:   aws.String(cognitoUsername),
			GroupName:  aws.String(group),
		})
	}

	// Add to new plan group
	newGroup := "plan-" + newPlan
	_, err = cognitoClient.AdminAddUserToGroup(ctx, &cognitoidentityprovider.AdminAddUserToGroupInput{
		UserPoolId: aws.String(cognitoUserPoolID),
		Username:   aws.String(cognitoUsername),
		GroupName:  aws.String(newGroup),
	})
	if err != nil {
		log.Printf("Failed to add user %s to Cognito group %s: %v", cognitoUsername, newGroup, err)
	} else {
		log.Printf("Synced user %s to Cognito group %s", cognitoUsername, newGroup)
	}
}

// setTrialExpired marks an account's trial as expired.
func setTrialExpired(ctx context.Context, dbClient *dynamodb.Client, accountID string) {
	_, err := dbClient.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(accountsTable),
		Key: map[string]ddbtypes.AttributeValue{
			"account_id": &ddbtypes.AttributeValueMemberS{Value: accountID},
		},
		UpdateExpression: aws.String("SET trial_expired = :te, updated_at = :now"),
		ExpressionAttributeValues: map[string]ddbtypes.AttributeValue{
			":te":  &ddbtypes.AttributeValueMemberBOOL{Value: true},
			":now": &ddbtypes.AttributeValueMemberS{Value: time.Now().Format(time.RFC3339)},
		},
	})
	if err != nil {
		log.Printf("Failed to set trial_expired for account %s: %v", accountID, err)
	} else {
		log.Printf("Set trial_expired=true for account %s", accountID)
	}
}

// storeMeteredItemID retrieves the subscription, finds the metered price item,
// and stores its ID on the account for usage reporting.
func storeMeteredItemID(ctx context.Context, dbClient *dynamodb.Client, accountID, subscriptionID, stripeKey string) {
	stripe.Key = stripeKey
	if stripe.Key == "" {
		return
	}

	params := &stripe.SubscriptionParams{}
	params.AddExpand("items")

	sub, err := stripeSub.Get(subscriptionID, params)
	if err != nil {
		log.Printf("Failed to get subscription %s: %v", subscriptionID, err)
		return
	}

	// Find the metered subscription item
	var meteredItemID string
	for _, item := range sub.Items.Data {
		if item.Price != nil && item.Price.Recurring != nil && item.Price.Recurring.UsageType == "metered" {
			meteredItemID = item.ID
			break
		}
	}

	if meteredItemID == "" {
		log.Printf("No metered item found in subscription %s for account %s", subscriptionID, accountID)
		return
	}

	// Store on account settings
	_, err = dbClient.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(accountsTable),
		Key: map[string]ddbtypes.AttributeValue{
			"account_id": &ddbtypes.AttributeValueMemberS{Value: accountID},
		},
		UpdateExpression: aws.String("SET settings.stripe_metered_item_id = :id"),
		ExpressionAttributeValues: map[string]ddbtypes.AttributeValue{
			":id": &ddbtypes.AttributeValueMemberS{Value: meteredItemID},
		},
	})
	if err != nil {
		log.Printf("Failed to store metered item ID for account %s: %v", accountID, err)
	} else {
		log.Printf("Stored metered item %s for account %s", meteredItemID, accountID)
	}
}
