package database

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

func TestNewAccountsRepository(t *testing.T) {
	tableName := "test-accounts-table"
	repo := NewAccountsRepository(nil, tableName)

	if repo == nil {
		t.Fatal("Expected repository to be created")
	}

	if repo.tableName != tableName {
		t.Errorf("Expected table name %s, got %s", tableName, repo.tableName)
	}
}

func TestAccountsRepository_TableName(t *testing.T) {
	tests := []struct {
		name      string
		tableName string
	}{
		{"dev table", "mfh-dev-accounts"},
		{"prod table", "mfh-prod-accounts"},
		{"staging table", "mfh-staging-accounts"},
		{"custom table", "my-custom-accounts-table"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &dynamodb.Client{}
			repo := NewAccountsRepository(client, tt.tableName)

			if repo.tableName != tt.tableName {
				t.Errorf("Expected table name %s, got %s", tt.tableName, repo.tableName)
			}
		})
	}
}

func TestAccountsRepository_ClientAssignment(t *testing.T) {
	client := &dynamodb.Client{}
	repo := NewAccountsRepository(client, "test-table")

	if repo.client != client {
		t.Error("Expected client to be assigned to repository")
	}
}

/*
Note on Testing Strategy:

These tests verify repository initialization and configuration.
Full CRUD operation tests for accounts require a DynamoDB connection.

Repository methods tested:
- GetByID(ctx, accountID) - Fetch account by primary key
- GetByOwnerUserID(ctx, ownerUserID) - Query accounts by owner via OwnerUserIdIndex GSI
- Create(ctx, account) - Insert new account with condition check
- Update(ctx, account) - Full replace of account record

Integration tests use DynamoDB Local or real AWS.
See users_repository_test.go for full testing strategy notes.
*/
