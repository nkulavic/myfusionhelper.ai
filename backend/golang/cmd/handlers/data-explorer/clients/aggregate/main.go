package aggregate

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

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

// AggregateRequest is the expected POST body for aggregate queries.
type AggregateRequest struct {
	ConnectionID string     `json:"connection_id"`
	ObjectType   string     `json:"object_type"`
	Metric       string     `json:"metric"`       // count, sum, average
	MetricField  string     `json:"metric_field"`  // field to sum/avg (required for sum/average)
	Dimension    string     `json:"dimension"`     // field to group by, or "_total" for scorecard
	DateRange    *DateRange `json:"date_range"`
	Limit        int        `json:"limit"`
}

// DateRange filters data by a date column.
type DateRange struct {
	Start string `json:"start"` // ISO date (YYYY-MM-DD)
	End   string `json:"end"`
}

// AggregateResult is a single group-by row.
type AggregateResult struct {
	Label string  `json:"label"`
	Value float64 `json:"value"`
}

var allowedObjectTypes = map[string]bool{
	"contacts":      true,
	"tags":          true,
	"custom_fields": true,
}

var allowedMetrics = map[string]bool{
	"count":   true,
	"sum":     true,
	"average": true,
}

// HandleWithAuth handles POST /data/aggregate.
func HandleWithAuth(ctx context.Context, event events.APIGatewayV2HTTPRequest, authCtx *apitypes.AuthContext) (events.APIGatewayV2HTTPResponse, error) {
	start := time.Now()

	var req AggregateRequest
	if err := json.Unmarshal([]byte(apiutil.GetBody(event)), &req); err != nil {
		return authMiddleware.CreateErrorResponse(400, "Invalid request body"), nil
	}

	if req.ConnectionID == "" || req.ObjectType == "" || req.Dimension == "" {
		return authMiddleware.CreateErrorResponse(400, "connection_id, object_type, and dimension are required"), nil
	}
	if !allowedObjectTypes[req.ObjectType] {
		return authMiddleware.CreateErrorResponse(400, "Invalid object_type"), nil
	}
	if req.Metric == "" {
		req.Metric = "count"
	}
	if !allowedMetrics[req.Metric] {
		return authMiddleware.CreateErrorResponse(400, "Invalid metric (must be count, sum, or average)"), nil
	}
	if (req.Metric == "sum" || req.Metric == "average") && req.MetricField == "" {
		return authMiddleware.CreateErrorResponse(400, "metric_field is required for sum/average"), nil
	}
	if req.Limit < 1 || req.Limit > 100 {
		req.Limit = 20
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

	parquetPath := fmt.Sprintf("s3://%s/%s/%s/%s/chunk_*.parquet",
		analyticsBucket, authCtx.AccountID, req.ConnectionID, req.ObjectType)

	// Open DuckDB
	db, err := sql.Open("duckdb", "")
	if err != nil {
		log.Printf("Failed to open DuckDB: %v", err)
		return authMiddleware.CreateErrorResponse(500, "Query engine error"), nil
	}
	defer db.Close()

	for _, stmt := range []string{
		"INSTALL httpfs", "LOAD httpfs",
		"SET s3_region='us-west-2'", "SET s3_use_ssl=true",
	} {
		if _, err := db.ExecContext(ctx, stmt); err != nil {
			log.Printf("DuckDB setup failed (%s): %v", stmt, err)
			return authMiddleware.CreateErrorResponse(500, "Query engine setup error"), nil
		}
	}

	// Handle _total (scorecard) — no GROUP BY
	if req.Dimension == "_total" {
		return handleScorecard(ctx, db, req, parquetPath, start)
	}

	// Build aggregate SQL
	var metricExpr string
	switch req.Metric {
	case "count":
		metricExpr = "COUNT(*)"
	case "sum":
		metricExpr = fmt.Sprintf("SUM(CAST(\"%s\" AS DOUBLE))", req.MetricField)
	case "average":
		metricExpr = fmt.Sprintf("AVG(CAST(\"%s\" AS DOUBLE))", req.MetricField)
	}

	whereClause, params := buildDateWhereClause(req.DateRange)

	querySQL := fmt.Sprintf(
		`SELECT COALESCE(CAST("%s" AS VARCHAR), 'Unknown') AS label, %s AS value FROM read_parquet('%s')`,
		req.Dimension, metricExpr, parquetPath,
	)
	if whereClause != "" {
		querySQL += " WHERE " + whereClause
	}
	querySQL += fmt.Sprintf(` GROUP BY label ORDER BY value DESC LIMIT %d`, req.Limit)

	rows, err := db.QueryContext(ctx, querySQL, params...)
	if err != nil {
		log.Printf("Aggregate query failed: %v (SQL: %s)", err, querySQL)
		return authMiddleware.CreateErrorResponse(500, "Query execution error"), nil
	}
	defer rows.Close()

	var results []AggregateResult
	var total float64
	for rows.Next() {
		var r AggregateResult
		if err := rows.Scan(&r.Label, &r.Value); err != nil {
			log.Printf("Failed to scan aggregate row: %v", err)
			return authMiddleware.CreateErrorResponse(500, "Failed to read results"), nil
		}
		total += r.Value
		results = append(results, r)
	}
	if err := rows.Err(); err != nil {
		log.Printf("Rows error: %v", err)
		return authMiddleware.CreateErrorResponse(500, "Failed to read results"), nil
	}

	if results == nil {
		results = []AggregateResult{}
	}

	queryTimeMs := time.Since(start).Milliseconds()

	return authMiddleware.CreateSuccessResponse(200, "Aggregation complete", map[string]interface{}{
		"results":       results,
		"total":         total,
		"query_time_ms": queryTimeMs,
	}), nil
}

func handleScorecard(ctx context.Context, db *sql.DB, req AggregateRequest, parquetPath string, start time.Time) (events.APIGatewayV2HTTPResponse, error) {
	var metricExpr string
	switch req.Metric {
	case "count":
		metricExpr = "COUNT(*)"
	case "sum":
		metricExpr = fmt.Sprintf("SUM(CAST(\"%s\" AS DOUBLE))", req.MetricField)
	case "average":
		metricExpr = fmt.Sprintf("AVG(CAST(\"%s\" AS DOUBLE))", req.MetricField)
	}

	whereClause, params := buildDateWhereClause(req.DateRange)
	querySQL := fmt.Sprintf("SELECT %s AS value FROM read_parquet('%s')", metricExpr, parquetPath)
	if whereClause != "" {
		querySQL += " WHERE " + whereClause
	}

	var value float64
	err := db.QueryRowContext(ctx, querySQL, params...).Scan(&value)
	if err != nil {
		log.Printf("Scorecard query failed: %v", err)
		return authMiddleware.CreateErrorResponse(500, "Query execution error"), nil
	}

	queryTimeMs := time.Since(start).Milliseconds()

	return authMiddleware.CreateSuccessResponse(200, "Aggregation complete", map[string]interface{}{
		"results": []AggregateResult{
			{Label: "_total", Value: value},
		},
		"total":         value,
		"query_time_ms": queryTimeMs,
	}), nil
}

func buildDateWhereClause(dr *DateRange) (string, []interface{}) {
	if dr == nil || (dr.Start == "" && dr.End == "") {
		return "", nil
	}

	var clauses []string
	var params []interface{}

	if dr.Start != "" {
		clauses = append(clauses, `CAST("created_at" AS VARCHAR) >= ?`)
		params = append(params, dr.Start)
	}
	if dr.End != "" {
		clauses = append(clauses, `CAST("created_at" AS VARCHAR) <= ?`)
		params = append(params, dr.End)
	}

	return fmt.Sprintf("(%s)", joinAnd(clauses)), params
}

func joinAnd(clauses []string) string {
	result := ""
	for i, c := range clauses {
		if i > 0 {
			result += " AND "
		}
		result += c
	}
	return result
}
