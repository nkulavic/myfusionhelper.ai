package database

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

func TestNewUsersRepository(t *testing.T) {
	// Test repository initialization without a functional client
	// The client will be nil/uninitialized, but we can verify the struct is created

	tableName := "test-users-table"
	repo := NewUsersRepository(nil, tableName)

	if repo == nil {
		t.Fatal("Expected repository to be created")
	}

	if repo.tableName != tableName {
		t.Errorf("Expected table name %s, got %s", tableName, repo.tableName)
	}
}

func TestUsersRepository_TableName(t *testing.T) {
	tests := []struct {
		name      string
		tableName string
	}{
		{"dev table", "mfh-dev-users"},
		{"prod table", "mfh-prod-users"},
		{"staging table", "mfh-staging-users"},
		{"custom table", "my-custom-users-table"},
		{"empty table name", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &dynamodb.Client{}
			repo := NewUsersRepository(client, tt.tableName)

			if repo.tableName != tt.tableName {
				t.Errorf("Expected table name %s, got %s", tt.tableName, repo.tableName)
			}
		})
	}
}

func TestUsersRepository_ClientAssignment(t *testing.T) {
	client := &dynamodb.Client{}
	repo := NewUsersRepository(client, "test-table")

	if repo.client != client {
		t.Error("Expected client to be assigned to repository")
	}
}

/*
Note on Testing Strategy:

These tests verify repository initialization and configuration.
Full CRUD operation tests require a DynamoDB connection and are tested via:

1. Integration tests with DynamoDB Local:
   docker run -d -p 8000:8000 amazon/dynamodb-local
   AWS_ACCESS_KEY_ID=test AWS_SECRET_ACCESS_KEY=test \
   AWS_ENDPOINT_URL=http://localhost:8000 \
   go test -tags=integration ./internal/database/...

2. End-to-end API tests:
   Test the full stack including DynamoDB operations via HTTP endpoints.
   See backend/golang/cmd/handlers/{service}/{endpoint}/main_test.go files

3. Handler-level tests:
   Mock the repository layer to test handler logic in isolation.

The repository methods (GetByID, GetByEmail, Create, Update, etc.) are thin
wrappers around generic DynamoDB helpers. Their correctness is verified by:
- Type safety (compile-time checks)
- Integration tests (runtime behavior)
- End-to-end tests (full system behavior)
*/
