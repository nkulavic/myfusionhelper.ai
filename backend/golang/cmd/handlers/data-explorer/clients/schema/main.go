package schema

import (
	"context"
	"log"
	"os"
	"sort"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	authMiddleware "github.com/myfusionhelper/api/internal/middleware/auth"
	"github.com/myfusionhelper/api/internal/services/parquet"
	apitypes "github.com/myfusionhelper/api/internal/types"
)

var analyticsBucket = os.Getenv("ANALYTICS_BUCKET")

// SchemaResponse is the full schema for a single object type.
type SchemaResponse struct {
	ConnectionID string                    `json:"connection_id"`
	ObjectType   string                    `json:"object_type"`
	RecordCount  int                       `json:"record_count"`
	SyncedAt     string                    `json:"synced_at"`
	Columns      map[string]SchemaColumn   `json:"columns"`
}

// SchemaColumn mirrors the parquet ColumnInfo with the fields the frontend needs.
type SchemaColumn struct {
	Type         string   `json:"type"`
	DisplayName  string   `json:"display_name"`
	Nullable     bool     `json:"nullable"`
	SampleValues []string `json:"sample_values,omitempty"`
}

// HandleWithAuth handles GET /data/schema?connection_id=X&object_type=Y.
func HandleWithAuth(ctx context.Context, event events.APIGatewayV2HTTPRequest, authCtx *apitypes.AuthContext) (events.APIGatewayV2HTTPResponse, error) {
	connectionID := event.QueryStringParameters["connection_id"]
	objectType := event.QueryStringParameters["object_type"]

	if connectionID == "" || objectType == "" {
		return authMiddleware.CreateErrorResponse(400, "connection_id and object_type are required"), nil
	}

	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(os.Getenv("COGNITO_REGION")))
	if err != nil {
		log.Printf("Failed to load AWS config: %v", err)
		return authMiddleware.CreateErrorResponse(500, "Internal server error"), nil
	}

	s3Client := s3.NewFromConfig(cfg)

	schemaKey := authCtx.AccountID + "/" + connectionID + "/" + objectType + "/schema.json"
	schemaInfo, err := parquet.ReadSchema(ctx, s3Client, analyticsBucket, schemaKey)
	if err != nil {
		log.Printf("Failed to read schema for %s/%s: %v", connectionID, objectType, err)
		return authMiddleware.CreateErrorResponse(404, "Schema not found — data may not be synced yet"), nil
	}

	// Convert parquet.ColumnInfo map to our response columns, sorted by name
	columns := make(map[string]SchemaColumn, len(schemaInfo.Columns))
	colNames := make([]string, 0, len(schemaInfo.Columns))
	for name := range schemaInfo.Columns {
		colNames = append(colNames, name)
	}
	sort.Strings(colNames)

	for _, name := range colNames {
		col := schemaInfo.Columns[name]
		columns[name] = SchemaColumn{
			Type:         col.Type,
			DisplayName:  col.DisplayName,
			Nullable:     col.Nullable,
			SampleValues: col.SampleValues,
		}
	}

	resp := SchemaResponse{
		ConnectionID: connectionID,
		ObjectType:   objectType,
		RecordCount:  schemaInfo.RecordCount,
		SyncedAt:     schemaInfo.SyncedAt,
		Columns:      columns,
	}

	return authMiddleware.CreateSuccessResponse(200, "Schema retrieved", resp), nil
}
