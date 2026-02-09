package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	ddbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	"github.com/myfusionhelper/api/internal/connectors"
	"github.com/myfusionhelper/api/internal/connectors/loader"
	"github.com/myfusionhelper/api/internal/google"
	helperEngine "github.com/myfusionhelper/api/internal/helpers"
	stripeusage "github.com/myfusionhelper/api/internal/stripe"

	// Register all connectors via init()
	_ "github.com/myfusionhelper/api/internal/connectors"

	// Register all helpers via init()
	_ "github.com/myfusionhelper/api/internal/helpers/analytics"
	_ "github.com/myfusionhelper/api/internal/helpers/automation"
	_ "github.com/myfusionhelper/api/internal/helpers/contact"
	_ "github.com/myfusionhelper/api/internal/helpers/data"
	_ "github.com/myfusionhelper/api/internal/helpers/integration"
	_ "github.com/myfusionhelper/api/internal/helpers/notification"
	_ "github.com/myfusionhelper/api/internal/helpers/tagging"
)

var (
	executionsTable      = os.Getenv("EXECUTIONS_TABLE")
	helpersTable         = os.Getenv("HELPERS_TABLE")
	notificationQueueURL = os.Getenv("NOTIFICATION_QUEUE_URL")
)

// HelperExecutionJob represents a job from the SQS queue
type HelperExecutionJob struct {
	ExecutionID  string                 `json:"execution_id"`
	HelperID     string                 `json:"helper_id"`
	HelperType   string                 `json:"helper_type"`
	AccountID    string                 `json:"account_id"`
	UserID       string                 `json:"user_id"`
	ConnectionID string                 `json:"connection_id"`
	ContactID    string                 `json:"contact_id"`
	Config       map[string]interface{} `json:"config"`
	Input        map[string]interface{} `json:"input"`
	APIKey       string                 `json:"api_key"`        // Original x-api-key for relay helpers
	RetryCount   int                    `json:"retry_count"`
}

func main() {
	lambda.Start(handleSQSEvent)
}

func handleSQSEvent(ctx context.Context, event events.SQSEvent) error {
	log.Printf("Processing %d SQS messages", len(event.Records))

	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(os.Getenv("COGNITO_REGION")))
	if err != nil {
		log.Printf("Failed to load AWS config: %v", err)
		return err
	}
	db := dynamodb.NewFromConfig(cfg)
	sqsClient := sqs.NewFromConfig(cfg)

	for _, record := range event.Records {
		var job HelperExecutionJob
		if err := json.Unmarshal([]byte(record.Body), &job); err != nil {
			log.Printf("Failed to unmarshal SQS message: %v", err)
			continue
		}

		log.Printf("Processing execution %s (helper: %s, type: %s)", job.ExecutionID, job.HelperID, job.HelperType)

		// Update execution status to running
		updateExecutionStatus(ctx, db, job.ExecutionID, "running", "", 0)

		// Execute the helper
		result, execErr := processJob(ctx, db, job)

		// Update execution record with results
		now := time.Now().UTC()
		if execErr != nil {
			log.Printf("Execution %s failed: %v", job.ExecutionID, execErr)
			updateExecutionResult(ctx, db, job.ExecutionID, "failed", execErr.Error(), result, &now)
			sendFailureNotification(ctx, sqsClient, job, execErr.Error())
		} else if result != nil && result.Success {
			log.Printf("Execution %s completed successfully", job.ExecutionID)
			updateExecutionResult(ctx, db, job.ExecutionID, "completed", "", result, &now)
			// Report usage to Stripe (best-effort, non-blocking)
			go stripeusage.ReportExecution(ctx, db, job.ExecutionID, job.AccountID, now.Unix())
		} else {
			errMsg := "execution returned unsuccessful result"
			if result != nil && result.Error != "" {
				errMsg = result.Error
			}
			log.Printf("Execution %s completed with errors: %s", job.ExecutionID, errMsg)
			updateExecutionResult(ctx, db, job.ExecutionID, "failed", errMsg, result, &now)
			sendFailureNotification(ctx, sqsClient, job, errMsg)
		}

		// Update helper execution count
		updateHelperStats(ctx, db, job.HelperID, &now)
	}

	return nil
}

func processJob(ctx context.Context, db *dynamodb.Client, job HelperExecutionJob) (*helperEngine.ExecutionResult, error) {
	// Load CRM connector if connection ID is specified
	var connector connectors.CRMConnector
	if job.ConnectionID != "" {
		var err error
		connector, err = loader.LoadConnectorWithTranslation(ctx, db, job.ConnectionID, job.AccountID)
		if err != nil {
			return nil, err
		}
	}

	// Pre-load service connection credentials (for non-CRM integrations like Zoom, Trello, etc.)
	serviceAuths := loadServiceAuths(ctx, db, job.Config, job.AccountID)

	// Execute via the helper engine
	executor := helperEngine.NewExecutor()
	execReq := helperEngine.ExecutionRequest{
		HelperType:   job.HelperType,
		ContactID:    job.ContactID,
		Config:       job.Config,
		UserID:       job.UserID,
		AccountID:    job.AccountID,
		HelperID:     job.HelperID,
		ConnectionID: job.ConnectionID,
		ServiceAuths: serviceAuths,
		APIKey:       job.APIKey,
	}

	result, err := executor.Execute(ctx, execReq, connector)

	// Process post-execution actions (e.g., google_sheet_sync_queued)
	if err == nil && result != nil && result.Output != nil && len(result.Output.Actions) > 0 {
		processPostExecutionActions(ctx, db, result.Output.Actions, job, connector, serviceAuths)
	}

	return result, err
}

// loadServiceAuths reads service_connection_ids from helper config and pre-loads
// auth credentials for each service. Returns a map keyed by platform slug.
func loadServiceAuths(ctx context.Context, db *dynamodb.Client, config map[string]interface{}, accountID string) map[string]*connectors.ConnectorConfig {
	raw, ok := config["service_connection_ids"]
	if !ok {
		return nil
	}

	connMap, ok := raw.(map[string]interface{})
	if !ok {
		return nil
	}

	auths := make(map[string]*connectors.ConnectorConfig, len(connMap))
	for slug, connIDRaw := range connMap {
		connID, ok := connIDRaw.(string)
		if !ok || connID == "" {
			continue
		}

		auth, err := loader.LoadServiceAuth(ctx, db, connID, accountID)
		if err != nil {
			log.Printf("Warning: failed to load service auth for %s (connection %s): %v", slug, connID, err)
			continue
		}
		auths[slug] = auth
	}

	if len(auths) == 0 {
		return nil
	}
	return auths
}

func updateExecutionStatus(ctx context.Context, db *dynamodb.Client, executionID, status, errorMsg string, durationMs int64) {
	updateExpr := "SET #s = :status"
	exprNames := map[string]string{"#s": "status"}
	exprValues := map[string]ddbtypes.AttributeValue{
		":status": &ddbtypes.AttributeValueMemberS{Value: status},
	}

	if errorMsg != "" {
		updateExpr += ", error_message = :error"
		exprValues[":error"] = &ddbtypes.AttributeValueMemberS{Value: errorMsg}
	}
	if durationMs > 0 {
		updateExpr += ", duration_ms = :duration"
		exprValues[":duration"] = &ddbtypes.AttributeValueMemberN{Value: fmt.Sprintf("%d", durationMs)}
	}

	_, err := db.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(executionsTable),
		Key: map[string]ddbtypes.AttributeValue{
			"execution_id": &ddbtypes.AttributeValueMemberS{Value: executionID},
		},
		UpdateExpression:          aws.String(updateExpr),
		ExpressionAttributeNames:  exprNames,
		ExpressionAttributeValues: exprValues,
	})
	if err != nil {
		log.Printf("Failed to update execution status: %v", err)
	}
}

func updateExecutionResult(ctx context.Context, db *dynamodb.Client, executionID, status, errorMsg string, result *helperEngine.ExecutionResult, completedAt *time.Time) {
	updateExpr := "SET #s = :status, completed_at = :completed_at"
	exprNames := map[string]string{"#s": "status"}
	exprValues := map[string]ddbtypes.AttributeValue{
		":status":       &ddbtypes.AttributeValueMemberS{Value: status},
		":completed_at": &ddbtypes.AttributeValueMemberS{Value: completedAt.Format(time.RFC3339)},
	}

	if errorMsg != "" {
		updateExpr += ", error_message = :error"
		exprValues[":error"] = &ddbtypes.AttributeValueMemberS{Value: errorMsg}
	}

	if result != nil {
		updateExpr += ", duration_ms = :duration"
		exprValues[":duration"] = &ddbtypes.AttributeValueMemberN{Value: fmt.Sprintf("%d", result.DurationMs)}

		if result.Output != nil {
			outputJSON, err := json.Marshal(result.Output)
			if err == nil {
				updateExpr += ", output = :output"
				exprValues[":output"] = &ddbtypes.AttributeValueMemberS{Value: string(outputJSON)}
			}
		}
	}

	_, err := db.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(executionsTable),
		Key: map[string]ddbtypes.AttributeValue{
			"execution_id": &ddbtypes.AttributeValueMemberS{Value: executionID},
		},
		UpdateExpression:          aws.String(updateExpr),
		ExpressionAttributeNames:  exprNames,
		ExpressionAttributeValues: exprValues,
	})
	if err != nil {
		log.Printf("Failed to update execution result: %v", err)
	}
}

func sendFailureNotification(ctx context.Context, sqsClient *sqs.Client, job HelperExecutionJob, errorMsg string) {
	if notificationQueueURL == "" {
		log.Printf("NOTIFICATION_QUEUE_URL not set, skipping failure notification")
		return
	}

	notification := map[string]interface{}{
		"type":       "execution_failure",
		"user_id":    job.UserID,
		"account_id": job.AccountID,
		"data": map[string]interface{}{
			"helper_name":   job.HelperType,
			"error_message": errorMsg,
			"execution_id":  job.ExecutionID,
			"helper_id":     job.HelperID,
		},
	}

	body, err := json.Marshal(notification)
	if err != nil {
		log.Printf("Failed to marshal notification: %v", err)
		return
	}

	_, err = sqsClient.SendMessage(ctx, &sqs.SendMessageInput{
		QueueUrl:       aws.String(notificationQueueURL),
		MessageBody:    aws.String(string(body)),
		MessageGroupId: aws.String(job.AccountID),
	})
	if err != nil {
		log.Printf("Failed to send notification to SQS: %v", err)
	} else {
		log.Printf("Sent failure notification for execution %s", job.ExecutionID)
	}
}

func updateHelperStats(ctx context.Context, db *dynamodb.Client, helperID string, executedAt *time.Time) {
	_, err := db.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(helpersTable),
		Key: map[string]ddbtypes.AttributeValue{
			"helper_id": &ddbtypes.AttributeValueMemberS{Value: helperID},
		},
		UpdateExpression: aws.String("SET execution_count = if_not_exists(execution_count, :zero) + :one, last_executed_at = :executed_at"),
		ExpressionAttributeValues: map[string]ddbtypes.AttributeValue{
			":zero":        &ddbtypes.AttributeValueMemberN{Value: "0"},
			":one":         &ddbtypes.AttributeValueMemberN{Value: "1"},
			":executed_at": &ddbtypes.AttributeValueMemberS{Value: executedAt.Format(time.RFC3339)},
		},
	})
	if err != nil {
		log.Printf("Failed to update helper stats: %v", err)
	}
}

// processPostExecutionActions handles post-execution actions like google_sheet_sync_queued
func processPostExecutionActions(
	ctx context.Context,
	db *dynamodb.Client,
	actions []helperEngine.HelperAction,
	job HelperExecutionJob,
	connector connectors.CRMConnector,
	serviceAuths map[string]*connectors.ConnectorConfig,
) {
	for _, action := range actions {
		switch action.Type {
		case "google_sheet_sync_queued":
			handleGoogleSheetSync(ctx, db, action, job, connector, serviceAuths)
		default:
			log.Printf("Skipping unknown post-execution action type: %s", action.Type)
		}
	}
}

// handleGoogleSheetSync processes the google_sheet_sync_queued action
func handleGoogleSheetSync(
	ctx context.Context,
	db *dynamodb.Client,
	action helperEngine.HelperAction,
	job HelperExecutionJob,
	connector connectors.CRMConnector,
	serviceAuths map[string]*connectors.ConnectorConfig,
) {
	log.Printf("Processing Google Sheet sync for spreadsheet %s", action.Target)

	// Extract sync request from action value
	syncRequest, ok := action.Value.(map[string]interface{})
	if !ok {
		log.Printf("Invalid sync request format for Google Sheet action")
		return
	}

	// Extract required fields
	spreadsheetID, _ := syncRequest["spreadsheet_id"].(string)
	sheetID, _ := syncRequest["sheet_id"].(string)
	mode, _ := syncRequest["mode"].(string)

	if spreadsheetID == "" || sheetID == "" {
		log.Printf("Missing required fields in sync request")
		return
	}

	// Get Google Sheets OAuth token from ServiceAuths
	googleAuth, ok := serviceAuths["google_sheets"]
	if !ok || googleAuth == nil {
		log.Printf("Google Sheets authentication not found in ServiceAuths")
		return
	}

	accessToken := googleAuth.AccessToken
	if accessToken == "" {
		log.Printf("Google Sheets access token is empty")
		return
	}

	// Create Sheets client
	sheetsClient := google.NewSheetsClient(accessToken)

	// Step 1: Clear worksheet if mode is "replace"
	if mode == "replace" {
		log.Printf("Clearing worksheet %s in spreadsheet %s", sheetID, spreadsheetID)
		if err := sheetsClient.ClearWorksheet(ctx, spreadsheetID, sheetID); err != nil {
			log.Printf("Failed to clear worksheet: %v", err)
			return
		}
	}

	// Step 2: Get contact data from CRM
	var rows [][]interface{}

	// Check if we have a search query or individual contact
	searchID, hasSearch := syncRequest["search_id"].(string)
	contactData, hasContact := syncRequest["contact_data"].(map[string]interface{})

	if hasSearch && searchID != "" {
		// Execute search query to get contacts
		log.Printf("Executing CRM search for search_id: %s", searchID)

		// For now, we'll use GetContacts with default options
		// In a production system, you'd load the saved search query and apply filters
		contactList, err := connector.GetContacts(ctx, connectors.QueryOptions{
			Limit: 1000, // reasonable default
		})
		if err != nil {
			log.Printf("Failed to fetch contacts from CRM: %v", err)
			return
		}

		// Convert contacts slice to pointer slice
		contactPtrs := make([]*connectors.NormalizedContact, len(contactList.Contacts))
		for i := range contactList.Contacts {
			contactPtrs[i] = &contactList.Contacts[i]
		}

		// Convert contacts to rows
		rows = contactsToRows(contactPtrs, syncRequest)
	} else if hasContact {
		// Single contact data provided
		log.Printf("Processing single contact data")
		rows = [][]interface{}{
			contactToRow(contactData, syncRequest),
		}
	} else {
		log.Printf("No contact data or search query provided")
		return
	}

	if len(rows) == 0 {
		log.Printf("No rows to write to Google Sheet")
		return
	}

	// Step 3: Write rows to worksheet
	log.Printf("Writing %d rows to worksheet %s", len(rows), sheetID)
	if err := sheetsClient.WriteRows(ctx, spreadsheetID, sheetID, rows); err != nil {
		log.Printf("Failed to write rows to worksheet: %v", err)
		return
	}

	log.Printf("Successfully synced %d rows to Google Sheet %s", len(rows), spreadsheetID)
}

// contactsToRows converts a list of normalized contacts to spreadsheet rows
func contactsToRows(contacts []*connectors.NormalizedContact, syncRequest map[string]interface{}) [][]interface{} {
	if len(contacts) == 0 {
		return [][]interface{}{}
	}

	// Build header row
	headerRow := []interface{}{"ID", "First Name", "Last Name", "Email", "Phone", "Company"}

	// Collect all unique custom field keys
	customFieldKeys := make(map[string]bool)
	for _, contact := range contacts {
		if contact.CustomFields != nil {
			for key := range contact.CustomFields {
				customFieldKeys[key] = true
			}
		}
	}

	// Add custom field headers
	for key := range customFieldKeys {
		headerRow = append(headerRow, key)
	}

	rows := [][]interface{}{headerRow}

	// Add data rows
	for _, contact := range contacts {
		row := []interface{}{
			contact.ID,
			contact.FirstName,
			contact.LastName,
			contact.Email,
			contact.Phone,
			contact.Company,
		}

		// Add custom field values
		for key := range customFieldKeys {
			value := ""
			if contact.CustomFields != nil {
				if v, ok := contact.CustomFields[key]; ok {
					value = fmt.Sprintf("%v", v)
				}
			}
			row = append(row, value)
		}

		rows = append(rows, row)
	}

	return rows
}

// contactToRow converts a single contact map to a spreadsheet row
func contactToRow(contactData map[string]interface{}, syncRequest map[string]interface{}) []interface{} {
	row := []interface{}{
		getStringValue(contactData, "Id"),
		getStringValue(contactData, "FirstName"),
		getStringValue(contactData, "LastName"),
		getStringValue(contactData, "Email"),
		getStringValue(contactData, "Phone"),
		getStringValue(contactData, "Company"),
	}
	return row
}

// getStringValue safely extracts a string value from a map
func getStringValue(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok {
		return fmt.Sprintf("%v", v)
	}
	return ""
}
