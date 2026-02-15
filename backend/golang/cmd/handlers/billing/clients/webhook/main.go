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
	accountsTable     = os.Getenv("ACCOUNTS_TABLE")
	usersTable        = os.Getenv("USERS_TABLE")
	cognitoUserPoolID = os.Getenv("COGNITO_USER_POOL_ID")
)

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

	log.Printf("Received Stripe event: %s", stripeEvent.Type)

	switch stripeEvent.Type {
	case "customer.subscription.created",
		"customer.subscription.updated":
		return handleSubscriptionUpdate(ctx, stripeEvent)
	case "customer.subscription.deleted":
		return handleSubscriptionCancelled(ctx, stripeEvent)
	case "checkout.session.completed":
		return handleCheckoutSessionCompleted(ctx, stripeEvent, stripeKey)
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

	// Determine plan from subscription metadata, then fall back to price metadata
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

	status := "active"
	if sub.Status == stripe.SubscriptionStatusPastDue {
		status = "active" // still active but payment issue
	} else if sub.Status == stripe.SubscriptionStatusCanceled {
		status = "cancelled"
	} else if sub.Status == stripe.SubscriptionStatusTrialing {
		status = "active"
	}

	ownerUserID, err := updateAccountByStripeCustomer(ctx, customerID, plan, status)
	if err != nil {
		log.Printf("Failed to update account: %v", err)
		return authMiddleware.CreateErrorResponse(500, "Failed to process event"), nil
	}

	// Sync Cognito plan group for the owner user
	if ownerUserID != "" {
		go syncCognitoPlanGroup(ctx, ownerUserID, plan)
	}

	// Send billing email notification asynchronously
	if ownerUserID != "" {
		cfg, _ := config.LoadDefaultConfig(ctx)
		dbClient := dynamodb.NewFromConfig(cfg)
		eventType := "subscription_created"
		if event.Type == "customer.subscription.updated" {
			eventType = "subscription_created" // same template for update
		}
		go sendBillingEmail(ctx, dbClient, ownerUserID, eventType, plan)
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

	ownerUserID, err := updateAccountByStripeCustomer(ctx, customerID, "free", "active")
	if err != nil {
		log.Printf("Failed to downgrade account: %v", err)
		return authMiddleware.CreateErrorResponse(500, "Failed to process event"), nil
	}

	if ownerUserID != "" {
		go syncCognitoPlanGroup(ctx, ownerUserID, "free")
		cfg, _ := config.LoadDefaultConfig(ctx)
		dbClient := dynamodb.NewFromConfig(cfg)
		go sendBillingEmail(ctx, dbClient, ownerUserID, "subscription_cancelled", "free")
	}

	return authMiddleware.CreateSuccessResponse(200, "OK", nil), nil
}

func handlePaymentFailed(ctx context.Context, event stripe.Event) (events.APIGatewayV2HTTPResponse, error) {
	log.Printf("Payment failed event received -- account status unchanged, Stripe will retry")
	// Stripe handles retry logic. We just log it. The subscription.updated event
	// will fire if the subscription status changes to past_due or cancelled.
	return authMiddleware.CreateSuccessResponse(200, "OK", nil), nil
}

func handleCheckoutSessionCompleted(ctx context.Context, event stripe.Event, stripeKey string) (events.APIGatewayV2HTTPResponse, error) {
	var sess stripe.CheckoutSession
	if err := json.Unmarshal(event.Data.Raw, &sess); err != nil {
		log.Printf("Failed to parse checkout session: %v", err)
		return authMiddleware.CreateErrorResponse(400, "Invalid event data"), nil
	}

	log.Printf("Checkout session completed: %s, mode: %s", sess.ID, sess.Mode)

	if sess.Mode != stripe.CheckoutSessionModeSubscription {
		log.Printf("Ignoring non-subscription checkout session")
		return authMiddleware.CreateSuccessResponse(200, "OK", nil), nil
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
		// Fall through -- subscription.created event will handle via customer ID lookup
		return authMiddleware.CreateSuccessResponse(200, "OK", nil), nil
	}

	if plan == "" {
		plan = "start" // default
	}

	log.Printf("Activating subscription for account %s, plan: %s", accountID, plan)

	planCfg := billing.GetPlan(plan)

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		log.Printf("Failed to load AWS config: %v", err)
		return authMiddleware.CreateErrorResponse(500, "Internal error"), nil
	}

	dbClient := dynamodb.NewFromConfig(cfg)

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
			":status": &ddbtypes.AttributeValueMemberS{Value: "active"},
			":mh":     &ddbtypes.AttributeValueMemberN{Value: intToStr(planCfg.MaxHelpers)},
			":mc":     &ddbtypes.AttributeValueMemberN{Value: intToStr(planCfg.MaxConnections)},
			":me":     &ddbtypes.AttributeValueMemberN{Value: intToStr(planCfg.MaxExecutions)},
			":now":    &ddbtypes.AttributeValueMemberS{Value: time.Now().Format(time.RFC3339)},
		},
	})
	if err != nil {
		log.Printf("Failed to activate account %s: %v", accountID, err)
		return authMiddleware.CreateErrorResponse(500, "Failed to activate subscription"), nil
	}

	log.Printf("Successfully activated account %s on plan %s", accountID, plan)

	// Extract and store stripe_metered_item_id from subscription items
	if sess.Subscription != nil && sess.Subscription.ID != "" {
		storeMeteredItemID(ctx, dbClient, accountID, sess.Subscription.ID, stripeKey)
	}

	// Send welcome/activation email
	ownerUserID := ""
	acctResult, err := dbClient.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(accountsTable),
		Key: map[string]ddbtypes.AttributeValue{
			"account_id": &ddbtypes.AttributeValueMemberS{Value: accountID},
		},
		ProjectionExpression: aws.String("owner_user_id"),
	})
	if err == nil && acctResult.Item != nil {
		if attr, ok := acctResult.Item["owner_user_id"]; ok {
			ownerUserID = attr.(*ddbtypes.AttributeValueMemberS).Value
		}
	}

	if ownerUserID != "" {
		go syncCognitoPlanGroup(ctx, ownerUserID, plan)
		go sendBillingEmail(ctx, dbClient, ownerUserID, "subscription_created", plan)
	}

	return authMiddleware.CreateSuccessResponse(200, "OK", nil), nil
}

func updateAccountByStripeCustomer(ctx context.Context, stripeCustomerID, plan, status string) (string, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return "", err
	}

	dbClient := dynamodb.NewFromConfig(cfg)

	// Query accounts by stripe_customer_id using StripeCustomerIdIndex GSI
	queryResult, err := dbClient.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(accountsTable),
		IndexName:              aws.String("StripeCustomerIdIndex"),
		KeyConditionExpression: aws.String("stripe_customer_id = :cid"),
		ExpressionAttributeValues: map[string]ddbtypes.AttributeValue{
			":cid": &ddbtypes.AttributeValueMemberS{Value: stripeCustomerID},
		},
		Limit: aws.Int32(1),
	})
	if err != nil {
		return "", err
	}
	if len(queryResult.Items) == 0 {
		log.Printf("No account found for Stripe customer %s", stripeCustomerID)
		return "", nil
	}

	// Get the account_id and owner_user_id from the result
	accountIDAttr, ok := queryResult.Items[0]["account_id"]
	if !ok {
		log.Printf("Account missing account_id field")
		return "", nil
	}
	accountID := accountIDAttr.(*ddbtypes.AttributeValueMemberS).Value

	ownerUserID := ""
	if ownerAttr, ok := queryResult.Items[0]["owner_user_id"]; ok {
		ownerUserID = ownerAttr.(*ddbtypes.AttributeValueMemberS).Value
	}

	planCfg := billing.GetPlan(plan)

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
			":mh":     &ddbtypes.AttributeValueMemberN{Value: intToStr(planCfg.MaxHelpers)},
			":mc":     &ddbtypes.AttributeValueMemberN{Value: intToStr(planCfg.MaxConnections)},
			":me":     &ddbtypes.AttributeValueMemberN{Value: intToStr(planCfg.MaxExecutions)},
			":now":    &ddbtypes.AttributeValueMemberS{Value: time.Now().Format(time.RFC3339)},
		},
	})
	return ownerUserID, err
}

func intToStr(n int) string {
	return fmt.Sprintf("%d", n)
}

// sendBillingEmail looks up the account owner and sends a billing notification email
func sendBillingEmail(ctx context.Context, dbClient *dynamodb.Client, ownerUserID, eventType, plan string) {
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

	if err := notifSvc.SendBillingEvent(ctx, userName, userEmail, eventType, billing.GetPlanLabel(plan)); err != nil {
		log.Printf("Failed to send billing email: %v", err)
	}
}

// syncCognitoPlanGroup updates a user's Cognito plan group.
// Removes from all plan-* groups, then adds to the new plan group.
func syncCognitoPlanGroup(ctx context.Context, ownerUserID, newPlan string) {
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
	dbClient := dynamodb.NewFromConfig(cfg)

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
	planGroups := []string{"plan-free", "plan-start", "plan-grow", "plan-deliver"}
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
