package database

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

func TestNewConnectionsRepository(t *testing.T) {
	tableName := "test-connections-table"
	repo := NewConnectionsRepository(nil, tableName)

	if repo == nil {
		t.Fatal("Expected repository to be created")
	}

	if repo.tableName != tableName {
		t.Errorf("Expected table name %s, got %s", tableName, repo.tableName)
	}
}

func TestConnectionsRepository_TableName(t *testing.T) {
	tests := []struct {
		name      string
		tableName string
	}{
		{"dev table", "mfh-dev-connections"},
		{"prod table", "mfh-prod-connections"},
		{"staging table", "mfh-staging-connections"},
		{"custom table", "my-custom-connections-table"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &dynamodb.Client{}
			repo := NewConnectionsRepository(client, tt.tableName)

			if repo.tableName != tt.tableName {
				t.Errorf("Expected table name %s, got %s", tt.tableName, repo.tableName)
			}
		})
	}
}

func TestConnectionsRepository_ClientAssignment(t *testing.T) {
	client := &dynamodb.Client{}
	repo := NewConnectionsRepository(client, "test-table")

	if repo.client != client {
		t.Error("Expected client to be assigned to repository")
	}
}

/*
Note on Testing Strategy:

These tests verify repository initialization and configuration.
Full CRUD operation tests for platform connections require a DynamoDB connection.

Repository methods tested:
- GetByID(ctx, connectionID) - Fetch connection by primary key
- GetByAccountID(ctx, accountID) - Query connections by account via AccountIdIndex GSI
- GetActiveByAccountID(ctx, accountID) - Query active connections by account
- GetByAccountIDAndPlatform(ctx, accountID, platformID) - Filter connections by platform
- Create(ctx, connection) - Insert new connection with condition check
- Update(ctx, connection) - Full replace of connection record
- UpdateStatus(ctx, connectionID, status) - Update connection status field only
- Delete(ctx, connectionID) - Hard delete connection

Integration tests use DynamoDB Local or real AWS.
See users_repository_test.go for full testing strategy notes.
*/
