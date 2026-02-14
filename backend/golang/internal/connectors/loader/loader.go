package loader

import (
	"context"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	ddbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	"github.com/myfusionhelper/api/internal/connectors"
	"github.com/myfusionhelper/api/internal/connectors/translate"
	apitypes "github.com/myfusionhelper/api/internal/types"
)

var (
	connectionsTable     = os.Getenv("CONNECTIONS_TABLE")
	connectionAuthsTable = os.Getenv("PLATFORM_CONNECTION_AUTHS_TABLE")
	platformsTable       = os.Getenv("PLATFORMS_TABLE")
)

// LoadConnector loads a CRM connector by looking up the connection, auth credentials,
// and platform definition from DynamoDB. It verifies account ownership.
func LoadConnector(ctx context.Context, db *dynamodb.Client, connectionID, accountID string) (connectors.CRMConnector, error) {
	// Get the connection record
	connResult, err := db.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(connectionsTable),
		Key: map[string]ddbtypes.AttributeValue{
			"connection_id": &ddbtypes.AttributeValueMemberS{Value: connectionID},
		},
	})
	if err != nil || connResult.Item == nil {
		return nil, &connectors.ConnectorError{
			Code: "CONNECTION_NOT_FOUND", Message: "connection not found",
			StatusCode: 404, Platform: "unknown",
		}
	}

	var connection apitypes.PlatformConnection
	if err := attributevalue.UnmarshalMap(connResult.Item, &connection); err != nil {
		return nil, err
	}

	// Verify account ownership
	if connection.AccountID != accountID {
		return nil, &connectors.ConnectorError{
			Code: "FORBIDDEN", Message: "connection does not belong to account",
			StatusCode: 403, Platform: "unknown",
		}
	}

	// Get the auth credentials
	if connection.AuthID == nil || *connection.AuthID == "" {
		return nil, &connectors.ConnectorError{
			Code: "NO_AUTH", Message: "connection has no auth credentials",
			StatusCode: 400, Platform: "unknown",
		}
	}

	authResult, err := db.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(connectionAuthsTable),
		Key: map[string]ddbtypes.AttributeValue{
			"auth_id": &ddbtypes.AttributeValueMemberS{Value: *connection.AuthID},
		},
	})
	if err != nil || authResult.Item == nil {
		return nil, &connectors.ConnectorError{
			Code: "AUTH_NOT_FOUND", Message: "auth credentials not found",
			StatusCode: 404, Platform: "unknown",
		}
	}

	var auth apitypes.PlatformConnectionAuth
	if err := attributevalue.UnmarshalMap(authResult.Item, &auth); err != nil {
		return nil, err
	}

	// Get the platform to determine slug
	platformResult, err := db.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(platformsTable),
		Key: map[string]ddbtypes.AttributeValue{
			"platform_id": &ddbtypes.AttributeValueMemberS{Value: connection.PlatformID},
		},
	})
	if err != nil || platformResult.Item == nil {
		return nil, &connectors.ConnectorError{
			Code: "PLATFORM_NOT_FOUND", Message: "platform not found",
			StatusCode: 404, Platform: "unknown",
		}
	}

	var platform apitypes.Platform
	if err := attributevalue.UnmarshalMap(platformResult.Item, &platform); err != nil {
		return nil, err
	}

	// Build connector config
	connConfig := connectors.ConnectorConfig{
		AccessToken: auth.AccessToken,
		APIKey:      auth.APIKey,
		APISecret:   auth.APISecret,
		BaseURL:     platform.APIConfig.BaseURL,
		AccountID:   connection.ExternalAppID,
	}

	return connectors.NewConnector(platform.Slug, connConfig)
}

// LoadConnectorWithTranslation loads a connector and wraps it with the translation
// layer for field name standardization, custom field resolution, and data normalization.
func LoadConnectorWithTranslation(ctx context.Context, db *dynamodb.Client, connectionID, accountID string) (connectors.CRMConnector, error) {
	connector, err := LoadConnector(ctx, db, connectionID, accountID)
	if err != nil {
		return nil, err
	}
	return translate.NewTranslatingConnector(connector), nil
}

// LoadServiceAuth loads auth credentials for a non-CRM service connection.
// Returns the raw ConnectorConfig (access_token, api_key, etc.) without
// creating a CRMConnector. Used by helpers that integrate with external
// services like Zoom, Google Sheets, Trello, etc.
func LoadServiceAuth(ctx context.Context, db *dynamodb.Client, connectionID, accountID string) (*connectors.ConnectorConfig, error) {
	// Get the connection record
	connResult, err := db.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(connectionsTable),
		Key: map[string]ddbtypes.AttributeValue{
			"connection_id": &ddbtypes.AttributeValueMemberS{Value: connectionID},
		},
	})
	if err != nil || connResult.Item == nil {
		return nil, &connectors.ConnectorError{
			Code: "CONNECTION_NOT_FOUND", Message: "service connection not found",
			StatusCode: 404, Platform: "unknown",
		}
	}

	var connection apitypes.PlatformConnection
	if err := attributevalue.UnmarshalMap(connResult.Item, &connection); err != nil {
		return nil, err
	}

	// Verify account ownership
	if connection.AccountID != accountID {
		return nil, &connectors.ConnectorError{
			Code: "FORBIDDEN", Message: "connection does not belong to account",
			StatusCode: 403, Platform: "unknown",
		}
	}

	// Get the auth credentials
	if connection.AuthID == nil || *connection.AuthID == "" {
		return nil, &connectors.ConnectorError{
			Code: "NO_AUTH", Message: "service connection has no auth credentials",
			StatusCode: 400, Platform: "unknown",
		}
	}

	authResult, err := db.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(connectionAuthsTable),
		Key: map[string]ddbtypes.AttributeValue{
			"auth_id": &ddbtypes.AttributeValueMemberS{Value: *connection.AuthID},
		},
	})
	if err != nil || authResult.Item == nil {
		return nil, &connectors.ConnectorError{
			Code: "AUTH_NOT_FOUND", Message: "service auth credentials not found",
			StatusCode: 404, Platform: "unknown",
		}
	}

	var auth apitypes.PlatformConnectionAuth
	if err := attributevalue.UnmarshalMap(authResult.Item, &auth); err != nil {
		return nil, err
	}

	// Get the platform for base URL
	platformResult, err := db.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(platformsTable),
		Key: map[string]ddbtypes.AttributeValue{
			"platform_id": &ddbtypes.AttributeValueMemberS{Value: connection.PlatformID},
		},
	})
	if err != nil || platformResult.Item == nil {
		return nil, &connectors.ConnectorError{
			Code: "PLATFORM_NOT_FOUND", Message: "platform not found",
			StatusCode: 404, Platform: "unknown",
		}
	}

	var platform apitypes.Platform
	if err := attributevalue.UnmarshalMap(platformResult.Item, &platform); err != nil {
		return nil, err
	}

	return &connectors.ConnectorConfig{
		AccessToken:  auth.AccessToken,
		RefreshToken: auth.RefreshToken,
		APIKey:       auth.APIKey,
		APISecret:    auth.APISecret,
		BaseURL:      platform.APIConfig.BaseURL,
		AccountID:    connection.ExternalAppID,
	}, nil
}
