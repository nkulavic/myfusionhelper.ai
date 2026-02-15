package query

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	ddbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	_ "github.com/marcboeker/go-duckdb"

	"github.com/myfusionhelper/api/internal/apiutil"
	authMiddleware "github.com/myfusionhelper/api/internal/middleware/auth"
	"github.com/myfusionhelper/api/internal/services/parquet"
	apitypes "github.com/myfusionhelper/api/internal/types"
)

var (
	connectionsTable = os.Getenv("CONNECTIONS_TABLE")
	analyticsBucket  = os.Getenv("ANALYTICS_BUCKET")
)

// QueryRequest is the expected POST body for data queries.
type QueryRequest struct {
	ConnectionID     string            `json:"connectionId"`
	ObjectType       string            `json:"objectType"`
	Page             int               `json:"page"`
	PageSize         int               `json:"pageSize"`
	SortBy           string            `json:"sortBy"`
	SortOrder        string            `json:"sortOrder"`
	FilterConditions []FilterCondition `json:"filterConditions"`
	Search           string            `json:"search"`
}

// FilterCondition represents a single filter clause.
type FilterCondition struct {
	Column   string      `json:"column"`
	Operator string      `json:"operator"`
	Value    interface{} `json:"value"`
	Value2   interface{} `json:"value2"`
}

// SchemaColumn is returned in the response for the frontend to understand column types.
type SchemaColumn struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	DisplayName string `json:"display_name"`
}

// allowedObjectTypes prevents injection of arbitrary S3 paths.
var allowedObjectTypes = map[string]bool{
	"contacts":      true,
	"tags":          true,
	"custom_fields": true,
}

// HandleWithAuth handles POST /data/query.
func HandleWithAuth(ctx context.Context, event events.APIGatewayV2HTTPRequest, authCtx *apitypes.AuthContext) (events.APIGatewayV2HTTPResponse, error) {
	start := time.Now()

	var req QueryRequest
	if err := json.Unmarshal([]byte(apiutil.GetBody(event)), &req); err != nil {
		return authMiddleware.CreateErrorResponse(400, "Invalid request body"), nil
	}

	// Validate required fields
	if req.ConnectionID == "" || req.ObjectType == "" {
		return authMiddleware.CreateErrorResponse(400, "connectionId and objectType are required"), nil
	}
	if !allowedObjectTypes[req.ObjectType] {
		return authMiddleware.CreateErrorResponse(400, "Invalid objectType"), nil
	}

	// Defaults
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 || req.PageSize > 500 {
		req.PageSize = 50
	}
	if req.SortOrder != "asc" && req.SortOrder != "desc" {
		req.SortOrder = "asc"
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

	// Read schema for column metadata
	s3Client := s3.NewFromConfig(cfg)
	schemaKey := authCtx.AccountID + "/" + req.ConnectionID + "/" + req.ObjectType + "/schema.json"
	schema, err := parquet.ReadSchema(ctx, s3Client, analyticsBucket, schemaKey)
	if err != nil {
		log.Printf("Failed to read schema: %v", err)
		return authMiddleware.CreateErrorResponse(404, "No data available for this object type"), nil
	}

	// Build column list and schema response
	var columnNames []string
	var schemaColumns []SchemaColumn
	var stringColumns []string
	for colName, colInfo := range schema.Columns {
		columnNames = append(columnNames, colName)
		schemaColumns = append(schemaColumns, SchemaColumn{
			Name:        colName,
			Type:        colInfo.Type,
			DisplayName: colInfo.DisplayName,
		})
		if colInfo.Type == "string" {
			stringColumns = append(stringColumns, colName)
		}
	}

	// Validate sortBy against known columns
	if req.SortBy != "" {
		found := false
		for _, c := range columnNames {
			if c == req.SortBy {
				found = true
				break
			}
		}
		if !found {
			req.SortBy = ""
		}
	}

	// Build parquet path
	parquetPath := fmt.Sprintf("s3://%s/%s/%s/%s/data.parquet", analyticsBucket, authCtx.AccountID, req.ConnectionID, req.ObjectType)

	// Open DuckDB in-memory
	db, err := sql.Open("duckdb", "")
	if err != nil {
		log.Printf("Failed to open DuckDB: %v", err)
		return authMiddleware.CreateErrorResponse(500, "Query engine error"), nil
	}
	defer db.Close()

	// Install and load httpfs, configure S3
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

	// Build WHERE clause
	whereClause, params := buildWhereClause(req, stringColumns)

	// Count query
	countSQL := fmt.Sprintf("SELECT COUNT(*) FROM read_parquet('%s')", parquetPath)
	if whereClause != "" {
		countSQL += " WHERE " + whereClause
	}

	var totalRecords int
	err = db.QueryRowContext(ctx, countSQL, params...).Scan(&totalRecords)
	if err != nil {
		log.Printf("Count query failed: %v", err)
		return authMiddleware.CreateErrorResponse(500, "Query execution error"), nil
	}

	// Data query
	dataSQL := fmt.Sprintf("SELECT * FROM read_parquet('%s')", parquetPath)
	if whereClause != "" {
		dataSQL += " WHERE " + whereClause
	}
	if req.SortBy != "" {
		dataSQL += fmt.Sprintf(" ORDER BY \"%s\" %s", req.SortBy, strings.ToUpper(req.SortOrder))
	}
	offset := (req.Page - 1) * req.PageSize
	dataSQL += fmt.Sprintf(" LIMIT %d OFFSET %d", req.PageSize, offset)

	rows, err := db.QueryContext(ctx, dataSQL, params...)
	if err != nil {
		log.Printf("Data query failed: %v", err)
		return authMiddleware.CreateErrorResponse(500, "Query execution error"), nil
	}
	defer rows.Close()

	records, cols, err := scanRows(rows)
	if err != nil {
		log.Printf("Failed to scan rows: %v", err)
		return authMiddleware.CreateErrorResponse(500, "Failed to read query results"), nil
	}

	totalPages := int(math.Ceil(float64(totalRecords) / float64(req.PageSize)))
	queryTimeMs := time.Since(start).Milliseconds()

	return authMiddleware.CreateSuccessResponse(200, "Query executed", map[string]interface{}{
		"records":       records,
		"total_records": totalRecords,
		"page":          req.Page,
		"page_size":     req.PageSize,
		"total_pages":   totalPages,
		"has_next_page": req.Page < totalPages,
		"has_prev_page": req.Page > 1,
		"columns":       cols,
		"schema":        schemaColumns,
		"query_time_ms": queryTimeMs,
	}), nil
}

// buildWhereClause constructs a parameterised WHERE clause from the request filters and search term.
func buildWhereClause(req QueryRequest, stringColumns []string) (string, []interface{}) {
	var clauses []string
	var params []interface{}

	for _, f := range req.FilterConditions {
		if f.Column == "" {
			continue
		}
		col := fmt.Sprintf("\"%s\"", f.Column)

		switch f.Operator {
		case "eq":
			clauses = append(clauses, col+" = ?")
			params = append(params, f.Value)
		case "neq":
			clauses = append(clauses, col+" <> ?")
			params = append(params, f.Value)
		case "contains":
			clauses = append(clauses, col+" ILIKE '%' || ? || '%'")
			params = append(params, f.Value)
		case "startswith":
			clauses = append(clauses, col+" ILIKE ? || '%'")
			params = append(params, f.Value)
		case "gt":
			clauses = append(clauses, col+" > ?")
			params = append(params, f.Value)
		case "gte":
			clauses = append(clauses, col+" >= ?")
			params = append(params, f.Value)
		case "lt":
			clauses = append(clauses, col+" < ?")
			params = append(params, f.Value)
		case "lte":
			clauses = append(clauses, col+" <= ?")
			params = append(params, f.Value)
		case "between":
			clauses = append(clauses, col+" BETWEEN ? AND ?")
			params = append(params, f.Value, f.Value2)
		case "in":
			if values, ok := f.Value.([]interface{}); ok && len(values) > 0 {
				placeholders := make([]string, len(values))
				for i, v := range values {
					placeholders[i] = "?"
					params = append(params, v)
				}
				clauses = append(clauses, col+" IN ("+strings.Join(placeholders, ", ")+")")
			}
		}
	}

	// Global search across all string columns
	if req.Search != "" && len(stringColumns) > 0 {
		var searchParts []string
		for _, sc := range stringColumns {
			searchParts = append(searchParts, fmt.Sprintf("\"%s\" ILIKE '%%' || ? || '%%'", sc))
			params = append(params, req.Search)
		}
		clauses = append(clauses, "("+strings.Join(searchParts, " OR ")+")")
	}

	return strings.Join(clauses, " AND "), params
}

// scanRows reads all rows from the result set into a slice of maps.
func scanRows(rows *sql.Rows) ([]map[string]interface{}, []string, error) {
	cols, err := rows.Columns()
	if err != nil {
		return nil, nil, err
	}

	var records []map[string]interface{}
	for rows.Next() {
		values := make([]interface{}, len(cols))
		valuePtrs := make([]interface{}, len(cols))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, nil, err
		}

		record := make(map[string]interface{}, len(cols))
		for i, col := range cols {
			val := values[i]
			// Convert []byte to string for JSON serialisation
			if b, ok := val.([]byte); ok {
				record[col] = string(b)
			} else {
				record[col] = val
			}
		}
		records = append(records, record)
	}

	if err := rows.Err(); err != nil {
		return nil, nil, err
	}

	if records == nil {
		records = []map[string]interface{}{}
	}

	return records, cols, nil
}
