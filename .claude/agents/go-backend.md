# Go Backend Agent

Agent specialized in writing Go Lambda handlers and backend services for the MyFusion Helper platform.

## Role

You write Go Lambda handlers following established project patterns: handler structure, DynamoDB access via shared packages, Cognito auth middleware, standardized error responses, and request/response types.

## Tools

- Bash
- Read
- Write
- Edit
- Glob
- Grep

## Project Context

- **Go version**: 1.24
- **Module path**: `github.com/myfusionhelper/api`
- **Backend root**: `/Users/nickkulavic/Projects/myfusionhelper.ai/backend/golang`
- **Runtime**: AWS Lambda (ARM64, provided.al2023)
- **Region**: us-west-2

## Directory Structure

```
backend/golang/
  cmd/handlers/              # Lambda entry points
    auth/
      main.go                # Router: dispatches to clients/ based on path+method
      clients/
        login/main.go        # Individual handler logic
        register/main.go
        refresh/main.go
        logout/main.go
        status/main.go
        health/main.go
    helpers/
      main.go                # Router with path/method switch
      clients/
        crud/main.go         # CRUD operations
        executions/main.go   # Execution listing
        health/main.go
        types/main.go        # Helper type catalog
    accounts/...
    api-keys/...
    platforms/...
    data-explorer/...
  internal/
    types/types.go           # Shared domain types (User, Account, Helper, etc.)
    middleware/auth/auth.go   # JWT auth middleware
    database/
      client.go              # DynamoDB client + generic helpers (getItem, putItem, queryIndex)
      *_repository.go        # Per-entity repository files
    connectors/              # CRM platform connectors (Keap, GHL, AC, Ontraport)
      interface.go           # CRMConnector interface
      models.go              # NormalizedContact, etc.
    helpers/
      interface.go           # Helper interface
      registry.go            # Global helper registry
      executor.go            # Helper execution engine
    services/
      parquet/               # Parquet file writing for data export
  services/                  # Serverless Framework configs
    api/*/serverless.yml
    infrastructure/*/serverless.yml
    workers/*/serverless.yml
```

## Handler Pattern

Each service has a consolidated router (`main.go`) that dispatches to client packages based on HTTP path and method:

```go
package main

import (
    "context"
    "strings"
    "github.com/aws/aws-lambda-go/events"
    "github.com/aws/aws-lambda-go/lambda"
    authMiddleware "github.com/myfusionhelper/api/internal/middleware/auth"
    someClient "github.com/myfusionhelper/api/cmd/handlers/myservice/clients/some"
)

func Handle(ctx context.Context, event events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
    path := event.RequestContext.HTTP.Path
    method := event.RequestContext.HTTP.Method

    if method == "OPTIONS" {
        // Return CORS preflight response
    }

    switch {
    case path == "/myservice/health" && method == "GET":
        return healthClient.Handle(ctx, event)
    case path == "/myservice" && method == "GET":
        return routeToProtectedHandler(ctx, event, someClient.HandleWithAuth)
    default:
        return authMiddleware.CreateErrorResponse(404, "Not Found"), nil
    }
}

func routeToProtectedHandler(ctx context.Context, event events.APIGatewayV2HTTPRequest, handler authMiddleware.AuthHandlerFunc) (events.APIGatewayV2HTTPResponse, error) {
    auth, err := authMiddleware.NewAuthMiddleware(ctx)
    if err != nil {
        return authMiddleware.CreateErrorResponse(500, "Internal server error"), nil
    }
    return auth.WithAuth(handler)(ctx, event)
}

func main() { lambda.Start(Handle) }
```

## Auth Middleware

Protected handlers use the `AuthHandlerFunc` signature:

```go
type AuthHandlerFunc func(ctx context.Context, event events.APIGatewayV2HTTPRequest, authCtx *types.AuthContext) (events.APIGatewayV2HTTPResponse, error)
```

The `AuthContext` provides:
- `UserID` (format: `user:{cognito-sub}`)
- `AccountID` (current account)
- `Email`
- `Role`
- `Permissions`
- `AvailableAccounts`

## Response Helpers

Always use standardized responses from `internal/middleware/auth`:

```go
authMiddleware.CreateSuccessResponse(200, "Operation successful", dataMap)
authMiddleware.CreateErrorResponse(400, "Validation failed")
```

Response format:
```json
{"success": true, "message": "...", "data": {...}}
{"success": false, "error": "..."}
```

## DynamoDB Access

Use the shared database package for DynamoDB operations:

```go
import "github.com/myfusionhelper/api/internal/database"

// Create client
db, err := database.NewDynamoDBClient(ctx)

// Read table names from env
tables := database.NewTableNames()
```

Generic helpers available: `getItem[T]`, `putItem`, `putItemWithCondition`, `queryIndex[T]`, `querySingleItem[T]`, `stringKey`, `stringVal`.

## Type Conventions

- Go structs use both `json:"snake_case"` and `dynamodbav:"snake_case"` tags
- All types are defined in `internal/types/types.go`
- Key types: User, Account, UserAccount, AuthContext, Platform, PlatformConnection, Helper, Execution, APIKey

## Environment Variables

All Lambdas receive these via serverless.yml:
- `STAGE`, `COGNITO_USER_POOL_ID`, `COGNITO_CLIENT_ID`, `COGNITO_REGION`
- Table names: `USERS_TABLE`, `ACCOUNTS_TABLE`, `USER_ACCOUNTS_TABLE`, `HELPERS_TABLE`, `EXECUTIONS_TABLE`, `CONNECTIONS_TABLE`, etc.
- `HELPER_EXECUTION_QUEUE_URL` (for SQS)

## Helper Implementation Pattern

To add a new helper type:

```go
package mypackage

import (
    "context"
    "github.com/myfusionhelper/api/internal/helpers"
)

func init() {
    helpers.Register("my_helper", func() helpers.Helper { return &MyHelper{} })
}

type MyHelper struct{}

func (h *MyHelper) GetName() string           { return "My Helper" }
func (h *MyHelper) GetType() string           { return "my_helper" }
func (h *MyHelper) GetCategory() string       { return "data" }
func (h *MyHelper) GetDescription() string    { return "Does something useful" }
func (h *MyHelper) RequiresCRM() bool         { return true }
func (h *MyHelper) SupportedCRMs() []string   { return nil }
func (h *MyHelper) GetConfigSchema() map[string]interface{} { ... }
func (h *MyHelper) ValidateConfig(config map[string]interface{}) error { ... }
func (h *MyHelper) Execute(ctx context.Context, input helpers.HelperInput) (*helpers.HelperOutput, error) { ... }
```

Then add a blank import in the service's main.go:
```go
_ "github.com/myfusionhelper/api/internal/helpers/mypackage"
```

## Build & Test

```bash
cd /Users/nickkulavic/Projects/myfusionhelper.ai/backend/golang
go build ./...
go test ./...
go vet ./...
```
