package export

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	ddbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	_ "github.com/marcboeker/go-duckdb"

	"github.com/myfusionhelper/api/internal/apiutil"
	authMiddleware "github.com/myfusionhelper/api/internal/middleware/auth"
	apitypes "github.com/myfusionhelper/api/internal/types"
)

var (
	connectionsTable = os.Getenv("CONNECTIONS_TABLE")
	analyticsBucket  = os.Getenv("ANALYTICS_BUCKET")
)

// ExportRequest is the expected POST body.
type ExportRequest struct {
	ConnectionID string `json:"connectionId"`
	ObjectType   string `json:"objectType"`
	Format       string `json:"format"` // "csv" or "json"
}

// allowedObjectTypes prevents injection of arbitrary S3 paths.
var allowedObjectTypes = map[string]bool{
	"contacts":      true,
	"tags":          true,
	"custom_fields": true,
}

// HandleWithAuth handles POST /data/export.
func HandleWithAuth(ctx context.Context, event events.APIGatewayV2HTTPRequest, authCtx *apitypes.AuthContext) (events.APIGatewayV2HTTPResponse, error) {
	var req ExportRequest
	if err := json.Unmarshal([]byte(apiutil.GetBody(event)), &req); err != nil {
		return authMiddleware.CreateErrorResponse(400, "Invalid request body"), nil
	}

	if req.ConnectionID == "" || req.ObjectType == "" {
		return authMiddleware.CreateErrorResponse(400, "connectionId and objectType are required"), nil
	}
	if !allowedObjectTypes[req.ObjectType] {
		return authMiddleware.CreateErrorResponse(400, "Invalid objectType"), nil
	}
	if req.Format != "csv" && req.Format != "json" {
		req.Format = "csv"
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
			"connection_id": &ddbtypes.AttributeValueMemberS{Value: req.ConnectionID},
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

	parquetPath := fmt.Sprintf("s3://%s/%s/%s/%s/data.parquet", analyticsBucket, authCtx.AccountID, req.ConnectionID, req.ObjectType)

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

	querySQL := fmt.Sprintf("SELECT * FROM read_parquet('%s')", parquetPath)
	rows, err := db.QueryContext(ctx, querySQL)
	if err != nil {
		log.Printf("Export query failed: %v", err)
		return authMiddleware.CreateErrorResponse(500, "Query execution error"), nil
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return authMiddleware.CreateErrorResponse(500, "Failed to read columns"), nil
	}

	// Read all records
	var records []map[string]interface{}
	for rows.Next() {
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
		records = append(records, record)
	}
	if err := rows.Err(); err != nil {
		return authMiddleware.CreateErrorResponse(500, "Failed to read query results"), nil
	}

	filename := fmt.Sprintf("%s_%s.%s", req.ConnectionID, req.ObjectType, req.Format)

	if req.Format == "json" {
		body, err := json.Marshal(records)
		if err != nil {
			return authMiddleware.CreateErrorResponse(500, "Failed to marshal JSON"), nil
		}
		return events.APIGatewayV2HTTPResponse{
			StatusCode: 200,
			Headers: map[string]string{
				"Content-Type":                 "application/json",
				"Content-Disposition":          fmt.Sprintf("attachment; filename=\"%s\"", filename),
				"Access-Control-Allow-Origin":  "*",
				"Access-Control-Allow-Methods": "GET, POST, PUT, DELETE, OPTIONS",
				"Access-Control-Allow-Headers": "Content-Type, Authorization, X-Account-Context",
			},
			Body: string(body),
		}, nil
	}

	// CSV format
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	// Write header
	if err := writer.Write(cols); err != nil {
		return authMiddleware.CreateErrorResponse(500, "Failed to write CSV header"), nil
	}

	// Write records
	for _, record := range records {
		row := make([]string, len(cols))
		for i, col := range cols {
			row[i] = fmt.Sprintf("%v", record[col])
		}
		if err := writer.Write(row); err != nil {
			return authMiddleware.CreateErrorResponse(500, "Failed to write CSV row"), nil
		}
	}
	writer.Flush()

	if err := writer.Error(); err != nil {
		return authMiddleware.CreateErrorResponse(500, "Failed to flush CSV"), nil
	}

	return events.APIGatewayV2HTTPResponse{
		StatusCode: 200,
		Headers: map[string]string{
			"Content-Type":                 "text/csv",
			"Content-Disposition":          fmt.Sprintf("attachment; filename=\"%s\"", filename),
			"Access-Control-Allow-Origin":  "*",
			"Access-Control-Allow-Methods": "GET, POST, PUT, DELETE, OPTIONS",
			"Access-Control-Allow-Headers": "Content-Type, Authorization, X-Account-Context",
		},
		Body: buf.String(),
	}, nil
}
