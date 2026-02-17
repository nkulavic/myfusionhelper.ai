package record

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	ddbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	_ "github.com/marcboeker/go-duckdb"

	authMiddleware "github.com/myfusionhelper/api/internal/middleware/auth"
	apitypes "github.com/myfusionhelper/api/internal/types"
)

var (
	connectionsTable = os.Getenv("CONNECTIONS_TABLE")
	analyticsBucket  = os.Getenv("ANALYTICS_BUCKET")
)

// allowedObjectTypes prevents injection of arbitrary S3 paths.
var allowedObjectTypes = map[string]bool{
	"contacts":      true,
	"tags":          true,
	"custom_fields": true,
}

// HandleWithAuth handles GET /data/record/{connectionId}/{objectType}/{recordId}.
func HandleWithAuth(ctx context.Context, event events.APIGatewayV2HTTPRequest, authCtx *apitypes.AuthContext) (events.APIGatewayV2HTTPResponse, error) {
	// Extract path parameters: /data/record/{connectionId}/{objectType}/{recordId}
	path := event.RequestContext.HTTP.Path
	trimmed := strings.TrimPrefix(path, "/data/record/")
	parts := strings.SplitN(trimmed, "/", 3)
	if len(parts) != 3 || parts[0] == "" || parts[1] == "" || parts[2] == "" {
		return authMiddleware.CreateErrorResponse(400, "Path must be /data/record/{connectionId}/{objectType}/{recordId}"), nil
	}

	connectionID := parts[0]
	objectType := parts[1]
	recordID := parts[2]

	if !allowedObjectTypes[objectType] {
		return authMiddleware.CreateErrorResponse(400, "Invalid objectType"), nil
	}

	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(os.Getenv("COGNITO_REGION")))
	if err != nil {
		log.Printf("Failed to load AWS config: %v", err)
		return authMiddleware.CreateErrorResponse(500, "Internal server error"), nil
	}

	// Verify account owns connection
	ddb := dynamodb.NewFromConfig(cfg)
	connResult, err := ddb.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(connectionsTable),
		Key: map[string]ddbtypes.AttributeValue{
			"connection_id": &ddbtypes.AttributeValueMemberS{Value: connectionID},
		},
	})
	if err != nil || connResult.Item == nil {
		return authMiddleware.CreateErrorResponse(404, "Connection not found"), nil
	}

	var conn apitypes.PlatformConnection
	if err := attributevalue.UnmarshalMap(connResult.Item, &conn); err != nil {
		return authMiddleware.CreateErrorResponse(500, "Failed to parse connection"), nil
	}
	if conn.AccountID != authCtx.AccountID {
		return authMiddleware.CreateErrorResponse(403, "Access denied"), nil
	}

	parquetPath := fmt.Sprintf("s3://%s/%s/%s/%s/chunk_*.parquet", analyticsBucket, authCtx.AccountID, connectionID, objectType)

	// Open DuckDB in-memory
	db, err := sql.Open("duckdb", "")
	if err != nil {
		log.Printf("Failed to open DuckDB: %v", err)
		return authMiddleware.CreateErrorResponse(500, "Query engine error"), nil
	}
	defer db.Close()

	setupStatements := []string{
		"INSTALL httpfs",
		"LOAD httpfs",
		"SET s3_region='us-west-2'",
		"SET s3_use_ssl=true",
	}
	for _, stmt := range setupStatements {
		if _, err := db.ExecContext(ctx, stmt); err != nil {
			log.Printf("DuckDB setup failed (%s): %v", stmt, err)
			return authMiddleware.CreateErrorResponse(500, "Query engine setup error"), nil
		}
	}

	query := fmt.Sprintf("SELECT * FROM read_parquet('%s') WHERE \"_record_id\" = ? LIMIT 1", parquetPath)
	rows, err := db.QueryContext(ctx, query, recordID)
	if err != nil {
		log.Printf("Record query failed: %v", err)
		return authMiddleware.CreateErrorResponse(500, "Query execution error"), nil
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return authMiddleware.CreateErrorResponse(500, "Failed to read columns"), nil
	}

	if !rows.Next() {
		return authMiddleware.CreateErrorResponse(404, "Record not found"), nil
	}

	values := make([]interface{}, len(cols))
	valuePtrs := make([]interface{}, len(cols))
	for i := range values {
		valuePtrs[i] = &values[i]
	}
	if err := rows.Scan(valuePtrs...); err != nil {
		return authMiddleware.CreateErrorResponse(500, "Failed to scan record"), nil
	}

	record := make(map[string]interface{}, len(cols))
	for i, col := range cols {
		val := values[i]
		if b, ok := val.([]byte); ok {
			record[col] = string(b)
		} else {
			record[col] = val
		}
	}

	return authMiddleware.CreateSuccessResponse(200, "Record retrieved", map[string]interface{}{
		"record": record,
	}), nil
}
