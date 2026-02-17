package timeseries

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

// TimeSeriesRequest is the expected POST body for time series queries.
type TimeSeriesRequest struct {
	ConnectionID string     `json:"connection_id"`
	ObjectType   string     `json:"object_type"`
	Metric       string     `json:"metric"`       // count, sum, average
	MetricField  string     `json:"metric_field"`  // field to sum/avg
	Interval     string     `json:"interval"`      // day, week, month, quarter, year
	DateRange    *DateRange `json:"date_range"`
	DateColumn   string     `json:"date_column"`   // defaults to "created_at"
}

// DateRange filters data by date.
type DateRange struct {
	Start string `json:"start"` // YYYY-MM-DD
	End   string `json:"end"`
}

// TimeSeriesPoint is a single time bucket.
type TimeSeriesPoint struct {
	Date  string  `json:"date"`
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

var allowedIntervals = map[string]bool{
	"day":     true,
	"week":    true,
	"month":   true,
	"quarter": true,
	"year":    true,
}

// HandleWithAuth handles POST /data/timeseries.
func HandleWithAuth(ctx context.Context, event events.APIGatewayV2HTTPRequest, authCtx *apitypes.AuthContext) (events.APIGatewayV2HTTPResponse, error) {
	start := time.Now()

	var req TimeSeriesRequest
	if err := json.Unmarshal([]byte(apiutil.GetBody(event)), &req); err != nil {
		return authMiddleware.CreateErrorResponse(400, "Invalid request body"), nil
	}

	if req.ConnectionID == "" || req.ObjectType == "" {
		return authMiddleware.CreateErrorResponse(400, "connection_id and object_type are required"), nil
	}
	if !allowedObjectTypes[req.ObjectType] {
		return authMiddleware.CreateErrorResponse(400, "Invalid object_type"), nil
	}
	if req.Metric == "" {
		req.Metric = "count"
	}
	if !allowedMetrics[req.Metric] {
		return authMiddleware.CreateErrorResponse(400, "Invalid metric"), nil
	}
	if (req.Metric == "sum" || req.Metric == "average") && req.MetricField == "" {
		return authMiddleware.CreateErrorResponse(400, "metric_field is required for sum/average"), nil
	}
	if req.Interval == "" {
		req.Interval = "month"
	}
	if !allowedIntervals[req.Interval] {
		return authMiddleware.CreateErrorResponse(400, "Invalid interval (day, week, month, quarter, year)"), nil
	}
	if req.DateColumn == "" {
		req.DateColumn = "created_at"
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

	// Build metric expression
	var metricExpr string
	switch req.Metric {
	case "count":
		metricExpr = "COUNT(*)"
	case "sum":
		metricExpr = fmt.Sprintf("SUM(CAST(\"%s\" AS DOUBLE))", req.MetricField)
	case "average":
		metricExpr = fmt.Sprintf("AVG(CAST(\"%s\" AS DOUBLE))", req.MetricField)
	}

	// Build date format for bucket label
	var dateFmt string
	switch req.Interval {
	case "day":
		dateFmt = "'%Y-%m-%d'"
	case "week":
		dateFmt = "'%Y-W%W'"
	case "month":
		dateFmt = "'%Y-%m'"
	case "quarter":
		dateFmt = "'%Y-Q'"
	case "year":
		dateFmt = "'%Y'"
	}

	// Build WHERE clause for date range
	whereClause, params := buildDateWhereClause(req.DateRange, req.DateColumn)

	// Build time series SQL
	bucketExpr := fmt.Sprintf(`DATE_TRUNC('%s', TRY_CAST("%s" AS TIMESTAMP))`, req.Interval, req.DateColumn)

	var querySQL string
	if req.Interval == "quarter" {
		// Special formatting for quarter
		querySQL = fmt.Sprintf(
			`SELECT CONCAT(CAST(YEAR(%s) AS VARCHAR), '-Q', CAST(QUARTER(%s) AS VARCHAR)) AS bucket_label, %s AS value FROM read_parquet('%s')`,
			bucketExpr, bucketExpr, metricExpr, parquetPath,
		)
	} else {
		querySQL = fmt.Sprintf(
			`SELECT STRFTIME(%s, %s) AS bucket_label, %s AS value FROM read_parquet('%s')`,
			bucketExpr, dateFmt, metricExpr, parquetPath,
		)
	}

	// Filter out NULL dates
	nullFilter := fmt.Sprintf(`TRY_CAST("%s" AS TIMESTAMP) IS NOT NULL`, req.DateColumn)
	if whereClause != "" {
		querySQL += " WHERE " + whereClause + " AND " + nullFilter
	} else {
		querySQL += " WHERE " + nullFilter
	}

	querySQL += " GROUP BY bucket_label ORDER BY bucket_label"

	rows, err := db.QueryContext(ctx, querySQL, params...)
	if err != nil {
		log.Printf("Timeseries query failed: %v (SQL: %s)", err, querySQL)
		return authMiddleware.CreateErrorResponse(500, "Query execution error"), nil
	}
	defer rows.Close()

	var points []TimeSeriesPoint
	for rows.Next() {
		var p TimeSeriesPoint
		if err := rows.Scan(&p.Date, &p.Value); err != nil {
			log.Printf("Failed to scan timeseries row: %v", err)
			return authMiddleware.CreateErrorResponse(500, "Failed to read results"), nil
		}
		points = append(points, p)
	}
	if err := rows.Err(); err != nil {
		log.Printf("Rows error: %v", err)
		return authMiddleware.CreateErrorResponse(500, "Failed to read results"), nil
	}

	if points == nil {
		points = []TimeSeriesPoint{}
	}

	queryTimeMs := time.Since(start).Milliseconds()

	return authMiddleware.CreateSuccessResponse(200, "Time series complete", map[string]interface{}{
		"points":        points,
		"interval":      req.Interval,
		"query_time_ms": queryTimeMs,
	}), nil
}

func buildDateWhereClause(dr *DateRange, dateColumn string) (string, []interface{}) {
	if dr == nil || (dr.Start == "" && dr.End == "") {
		return "", nil
	}

	var clauses []string
	var params []interface{}

	if dr.Start != "" {
		clauses = append(clauses, fmt.Sprintf(`CAST("%s" AS VARCHAR) >= ?`, dateColumn))
		params = append(params, dr.Start)
	}
	if dr.End != "" {
		clauses = append(clauses, fmt.Sprintf(`CAST("%s" AS VARCHAR) <= ?`, dateColumn))
		params = append(params, dr.End)
	}

	result := ""
	for i, c := range clauses {
		if i > 0 {
			result += " AND "
		}
		result += c
	}
	return "(" + result + ")", params
}
