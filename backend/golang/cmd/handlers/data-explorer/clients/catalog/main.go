package catalog

import (
	"context"
	"log"
	"os"
	"sort"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	ddbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	authMiddleware "github.com/myfusionhelper/api/internal/middleware/auth"
	"github.com/myfusionhelper/api/internal/services/parquet"
	apitypes "github.com/myfusionhelper/api/internal/types"
)

var (
	connectionsTable = os.Getenv("CONNECTIONS_TABLE")
	platformsTable   = os.Getenv("PLATFORMS_TABLE")
	analyticsBucket  = os.Getenv("ANALYTICS_BUCKET")
)

// objectTypeMeta defines the display metadata for each supported object type.
var objectTypeMeta = map[string]struct {
	Label string
	Icon  string
}{
	"contacts":      {Label: "Contacts", Icon: "users"},
	"tags":          {Label: "Tags", Icon: "tags"},
	"custom_fields": {Label: "Custom Fields", Icon: "list"},
}

// CatalogSource represents a single data source in the catalog response.
type CatalogSource struct {
	ObjectType     string          `json:"object_type"`
	Label          string          `json:"label"`
	Icon           string          `json:"icon"`
	RecordCount    int             `json:"record_count"`
	ConnectionID   string          `json:"connection_id"`
	ConnectionName string          `json:"connection_name"`
	PlatformID     string          `json:"platform_id"`
	PlatformName   string          `json:"platform_name"`
	PlatformSlug   string          `json:"platform_slug"`
	LastSyncedAt   string          `json:"last_synced_at,omitempty"`
	SyncStatus     string          `json:"sync_status,omitempty"`
	ColumnCount    int             `json:"column_count"`
	Columns        []SchemaColumn  `json:"columns,omitempty"`
}

// SchemaColumn is a lightweight column descriptor returned in the catalog.
type SchemaColumn struct {
	Name         string   `json:"name"`
	Type         string   `json:"type"`
	DisplayName  string   `json:"display_name"`
	SampleValues []string `json:"sample_values,omitempty"`
}

// HandleWithAuth returns the data catalog for the authenticated account.
func HandleWithAuth(ctx context.Context, event events.APIGatewayV2HTTPRequest, authCtx *apitypes.AuthContext) (events.APIGatewayV2HTTPResponse, error) {
	log.Printf("Catalog request for account: %s", authCtx.AccountID)

	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(os.Getenv("COGNITO_REGION")))
	if err != nil {
		log.Printf("Failed to load AWS config: %v", err)
		return authMiddleware.CreateErrorResponse(500, "Internal server error"), nil
	}

	ddb := dynamodb.NewFromConfig(cfg)
	s3Client := s3.NewFromConfig(cfg)

	// Query connections by account_id using AccountIdIndex GSI
	connectionsResult, err := ddb.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(connectionsTable),
		IndexName:              aws.String("AccountIdIndex"),
		KeyConditionExpression: aws.String("account_id = :account_id"),
		ExpressionAttributeValues: map[string]ddbtypes.AttributeValue{
			":account_id": &ddbtypes.AttributeValueMemberS{Value: authCtx.AccountID},
		},
	})
	if err != nil {
		log.Printf("Failed to query connections: %v", err)
		return authMiddleware.CreateErrorResponse(500, "Failed to retrieve connections"), nil
	}

	var connections []apitypes.PlatformConnection
	if err := attributevalue.UnmarshalListOfMaps(connectionsResult.Items, &connections); err != nil {
		log.Printf("Failed to unmarshal connections: %v", err)
		return authMiddleware.CreateErrorResponse(500, "Failed to parse connections"), nil
	}

	// Build a cache of platform info
	platformCache := make(map[string]*apitypes.Platform)

	var sources []CatalogSource

	for _, conn := range connections {
		if conn.Status != "active" {
			continue
		}

		// Fetch platform info if not cached
		if _, ok := platformCache[conn.PlatformID]; !ok {
			platformResult, err := ddb.GetItem(ctx, &dynamodb.GetItemInput{
				TableName: aws.String(platformsTable),
				Key: map[string]ddbtypes.AttributeValue{
					"platform_id": &ddbtypes.AttributeValueMemberS{Value: conn.PlatformID},
				},
			})
			if err != nil {
				log.Printf("Failed to get platform %s: %v", conn.PlatformID, err)
				continue
			}
			if platformResult.Item != nil {
				var platform apitypes.Platform
				if err := attributevalue.UnmarshalMap(platformResult.Item, &platform); err != nil {
					log.Printf("Failed to unmarshal platform %s: %v", conn.PlatformID, err)
					continue
				}
				platformCache[conn.PlatformID] = &platform
			}
		}

		platform := platformCache[conn.PlatformID]
		platformName := ""
		platformSlug := ""
		if platform != nil {
			platformName = platform.Name
			platformSlug = platform.Slug
		}

		// Check each object type for schema.json
		for objectType, meta := range objectTypeMeta {
			schemaKey := conn.AccountID + "/" + conn.ConnectionID + "/" + objectType + "/schema.json"
			schema, err := parquet.ReadSchema(ctx, s3Client, analyticsBucket, schemaKey)
			if err != nil {
				// Schema doesn't exist for this object type — skip
				continue
			}

			lastSynced := ""
			if schema.SyncedAt != "" {
				lastSynced = schema.SyncedAt
			}

			// Convert schema columns to sorted array
			var columns []SchemaColumn
			if len(schema.Columns) > 0 {
				colNames := make([]string, 0, len(schema.Columns))
				for name := range schema.Columns {
					colNames = append(colNames, name)
				}
				sort.Strings(colNames)
				for _, name := range colNames {
					col := schema.Columns[name]
					columns = append(columns, SchemaColumn{
						Name:         name,
						Type:         col.Type,
						DisplayName:  col.DisplayName,
						SampleValues: col.SampleValues,
					})
				}
			}

			sources = append(sources, CatalogSource{
				ObjectType:     objectType,
				Label:          meta.Label,
				Icon:           meta.Icon,
				RecordCount:    schema.RecordCount,
				ConnectionID:   conn.ConnectionID,
				ConnectionName: conn.Name,
				PlatformID:     conn.PlatformID,
				PlatformName:   platformName,
				PlatformSlug:   platformSlug,
				LastSyncedAt:   lastSynced,
				SyncStatus:     conn.SyncStatus,
				ColumnCount:    len(schema.Columns),
				Columns:        columns,
			})
		}
	}

	if sources == nil {
		sources = []CatalogSource{}
	}

	return authMiddleware.CreateSuccessResponse(200, "Catalog retrieved", map[string]interface{}{
		"sources": sources,
	}), nil
}
