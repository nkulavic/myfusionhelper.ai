package database

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	ddbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/myfusionhelper/api/internal/types"
)

// PlatformsRepository provides access to the platforms DynamoDB table.
type PlatformsRepository struct {
	client    *dynamodb.Client
	tableName string
}

// NewPlatformsRepository creates a new PlatformsRepository.
func NewPlatformsRepository(client *dynamodb.Client, tableName string) *PlatformsRepository {
	return &PlatformsRepository{client: client, tableName: tableName}
}

// GetByID fetches a platform by its platform_id (primary key).
func (r *PlatformsRepository) GetByID(ctx context.Context, platformID string) (*types.Platform, error) {
	return getItem[types.Platform](ctx, r.client, r.tableName, stringKey("platform_id", platformID))
}

// GetBySlug fetches a platform by slug using the SlugIndex GSI.
func (r *PlatformsRepository) GetBySlug(ctx context.Context, slug string) (*types.Platform, error) {
	indexName := "SlugIndex"
	return querySingleItem[types.Platform](ctx, r.client, &dynamodb.QueryInput{
		TableName:              &r.tableName,
		IndexName:              &indexName,
		KeyConditionExpression: aws.String("slug = :slug"),
		ExpressionAttributeValues: map[string]ddbtypes.AttributeValue{
			":slug": stringVal(slug),
		},
	})
}

// ListAll performs a full table scan to return all platforms.
func (r *PlatformsRepository) ListAll(ctx context.Context) ([]types.Platform, error) {
	result, err := r.client.Scan(ctx, &dynamodb.ScanInput{
		TableName: &r.tableName,
	})
	if err != nil {
		return nil, err
	}

	platforms := make([]types.Platform, 0, len(result.Items))
	for _, item := range result.Items {
		var p types.Platform
		if err := attributevalue.UnmarshalMap(item, &p); err != nil {
			return nil, err
		}
		platforms = append(platforms, p)
	}
	return platforms, nil
}

// Create inserts a new platform.
func (r *PlatformsRepository) Create(ctx context.Context, platform *types.Platform) error {
	return putItemWithCondition(ctx, r.client, r.tableName, platform, "attribute_not_exists(platform_id)")
}

// Update performs a full replace of the platform record.
func (r *PlatformsRepository) Update(ctx context.Context, platform *types.Platform) error {
	return putItem(ctx, r.client, r.tableName, platform)
}
