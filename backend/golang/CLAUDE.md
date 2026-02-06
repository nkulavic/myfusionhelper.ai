# MyFusion Helper - Go Backend

Serverless Go backend on AWS Lambda, deployed via Serverless Framework v4.

## Quick Reference

```bash
cd backend/golang

# Build all handlers
CGO_ENABLED=1 go build ./...

# Run tests
CGO_ENABLED=1 go test ./...

# Deploy a single service
npm install                          # first time only (installs serverless-go-plugin)
cd services/api/auth
npx sls deploy --stage dev

# Deploy DuckDB-dependent service (needs Docker for AL2023 glibc)
bash scripts/build-duckdb-handler.sh
cd services/api/data-explorer
npx sls deploy --stage dev
```

## Project Layout

```
backend/golang/
├── cmd/handlers/                  # Lambda entry points (one per service)
│   ├── auth/
│   │   ├── main.go                # Consolidated router for /auth/* paths
│   │   └── clients/               # Individual endpoint handlers
│   │       ├── login/main.go
│   │       ├── register/main.go
│   │       ├── refresh/main.go
│   │       ├── status/main.go
│   │       ├── logout/main.go
│   │       ├── profile/main.go
│   │       ├── forgot-password/main.go
│   │       ├── reset-password/main.go
│   │       └── health/main.go
│   ├── accounts/                  # /accounts/* endpoints
│   ├── api-keys/                  # /api-keys/* endpoints
│   ├── helpers/                   # /helpers/* + /executions/* endpoints
│   ├── platforms/                 # /platforms/* + /platform-connections/*
│   ├── billing/                   # /billing/* endpoints (Stripe integration)
│   │   ├── main.go                # Consolidated router
│   │   └── clients/
│   │       ├── get-billing/main.go
│   │       ├── checkout/main.go
│   │       ├── portal-session/main.go
│   │       ├── invoices/main.go
│   │       └── webhook/main.go
│   ├── data-explorer/             # /data/* endpoints (DuckDB + Parquet)
│   ├── data-sync/                 # SQS worker: sync CRM data → S3/Parquet
│   ├── data-sync-scheduler/       # EventBridge trigger for data-sync
│   └── helper-worker/             # SQS worker: execute helpers async
├── internal/
│   ├── connectors/                # CRM platform adapters
│   │   ├── interface.go           # CRMConnector interface
│   │   ├── keap.go
│   │   ├── gohighlevel.go
│   │   ├── activecampaign.go
│   │   ├── ontraport.go
│   │   ├── models.go              # NormalizedContact, Tag, CustomField
│   │   ├── registry.go            # Connector factory registry
│   │   ├── loader/loader.go       # Load connector with credentials
│   │   └── translate/             # Data normalization layer
│   ├── database/                  # DynamoDB repositories
│   │   ├── client.go              # NewDynamoDBClient, generic helpers
│   │   ├── users_repository.go
│   │   ├── accounts_repository.go
│   │   ├── user_accounts_repository.go
│   │   ├── connections_repository.go
│   │   ├── connection_auths_repository.go
│   │   ├── helpers_repository.go
│   │   ├── executions_repository.go
│   │   ├── platforms_repository.go
│   │   ├── apikeys_repository.go
│   │   └── oauth_states_repository.go
│   ├── helpers/                   # Helper implementations
│   │   ├── interface.go           # Helper interface
│   │   ├── registry.go            # Helper factory registry
│   │   ├── executor.go            # Execution orchestrator
│   │   ├── contact/               # Contact helpers (tag_it, copy_it, etc.)
│   │   ├── data/                  # Data helpers (format_it, math_it, etc.)
│   │   ├── tagging/               # Tag helpers (score_it, group_it, etc.)
│   │   ├── automation/            # Automation helpers (trigger_it, etc.)
│   │   ├── integration/           # Integration helpers (slack, email, etc.)
│   │   ├── notification/          # Notification helpers
│   │   └── analytics/             # Analytics helpers (RFM, CLV)
│   ├── middleware/auth/           # JWT auth middleware
│   │   └── auth.go                # WithAuth wrapper, response helpers
│   ├── services/parquet/          # Parquet file writer for data sync
│   └── types/types.go             # All shared Go types
├── services/                      # Serverless Framework service configs
│   ├── api/
│   │   ├── gateway/serverless.yml # Shared API Gateway + Cognito authorizer
│   │   ├── auth/serverless.yml
│   │   ├── accounts/serverless.yml
│   │   ├── api-keys/serverless.yml
│   │   ├── helpers/serverless.yml
│   │   ├── platforms/serverless.yml
│   │   ├── billing/serverless.yml
│   │   └── data-explorer/serverless.yml
│   ├── infrastructure/
│   │   ├── cognito/serverless.yml
│   │   ├── dynamodb/core/serverless.yml
│   │   ├── s3/serverless.yml
│   │   └── sqs/serverless.yml
│   └── workers/
│       ├── data-sync/serverless.yml
│       └── helper-worker/serverless.yml
├── docker/
│   └── Dockerfile.lambda-builder  # AL2023 container for DuckDB CGO builds
├── scripts/
│   ├── build-duckdb-handler.sh    # Build data-explorer in Docker
│   └── verify-deploy.sh           # Post-deploy health checks
├── go.mod
├── go.sum
└── package.json                   # serverless-go-plugin dependency
```

## Handler Pattern

Each API service uses a **consolidated handler** pattern: one Lambda binary routes to multiple endpoint handlers based on path.

```go
// cmd/handlers/auth/main.go
func Handle(ctx context.Context, event events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
    switch event.RequestContext.HTTP.Path {
    case "/auth/login":
        return loginClient.Handle(ctx, event)
    case "/auth/status":
        return routeToProtectedHandler(ctx, event, statusClient.HandleWithAuth)
    // ...
    }
}
```

**Public endpoints**: Use `Handle(ctx, event)` signature directly.

**Protected endpoints**: Use `HandleWithAuth(ctx, event, authCtx)` signature, wrapped by `routeToProtectedHandler` which runs the auth middleware.

### Client Handler Pattern

Each endpoint handler lives in `cmd/handlers/{service}/clients/{endpoint}/main.go`:

```go
package login

func Handle(ctx context.Context, event events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
    // Parse request, validate, execute, return response
}
```

For protected endpoints:
```go
package status

func HandleWithAuth(ctx context.Context, event events.APIGatewayV2HTTPRequest, authCtx *types.AuthContext) (events.APIGatewayV2HTTPResponse, error) {
    // authCtx contains UserID, AccountID, Email, Role, Permissions
}
```

## Auth Middleware (`internal/middleware/auth/auth.go`)

The `WithAuth` wrapper:
1. Extracts `sub` claim from JWT (tries API Gateway authorizer context first, falls back to Bearer token)
2. Constructs `userID` as `"user:" + sub`
3. Fetches user from DynamoDB to get `current_account_id`
4. Fetches user-account relationship for permissions
5. Builds `AuthContext` and passes to handler

**Important**: User IDs are prefixed with `user:` (e.g., `user:abc-123-def`).

## Response Helpers

All responses use standardized helpers from the auth middleware package:

```go
import authMiddleware "github.com/myfusionhelper/api/internal/middleware/auth"

// Success
return authMiddleware.CreateSuccessResponse(200, "OK", data), nil

// Error
return authMiddleware.CreateErrorResponse(400, "Bad request"), nil
```

Response format:
```json
{
  "success": true,
  "message": "OK",
  "data": { ... }
}
```

Error format:
```json
{
  "success": false,
  "error": "Error message"
}
```

All responses include CORS headers (`Access-Control-Allow-Origin: *`).

## DynamoDB Conventions

### Table Names

Table names come from environment variables, set via CloudFormation exports:

| Env Var | Table | Partition Key | Sort Key |
|---------|-------|--------------|----------|
| `USERS_TABLE` | `mfh-{stage}-users` | `user_id` | -- |
| `ACCOUNTS_TABLE` | `mfh-{stage}-accounts` | `account_id` | -- |
| `USER_ACCOUNTS_TABLE` | `mfh-{stage}-user-accounts` | `user_id` | `account_id` |
| `API_KEYS_TABLE` | `mfh-{stage}-api-keys` | `key_id` | -- |
| `CONNECTIONS_TABLE` | `mfh-{stage}-connections` | `connection_id` | -- |
| `PLATFORM_CONNECTION_AUTHS_TABLE` | `mfh-{stage}-platform-connection-auths` | `auth_id` | -- |
| `HELPERS_TABLE` | `mfh-{stage}-helpers` | `helper_id` | -- |
| `EXECUTIONS_TABLE` | `mfh-{stage}-executions` | `execution_id` | -- |
| `PLATFORMS_TABLE` | `mfh-{stage}-platforms` | `platform_id` | -- |
| `OAUTH_STATES_TABLE` | `mfh-{stage}-oauth-states` | `state` | -- |

### GSI Patterns

Most tables have an `AccountIdIndex` GSI for querying by account. Other notable GSIs:
- Users: `EmailIndex`, `CognitoUserIdIndex`
- Accounts: `OwnerUserIdIndex`
- API Keys: `KeyHashIndex`
- Executions: `AccountIdCreatedAtIndex`, `HelperIdCreatedAtIndex`
- Platforms: `SlugIndex`
- Platform Connection Auths: `ConnectionIdIndex`

### Database Client (`internal/database/client.go`)

Generic helpers for DynamoDB operations:
- `getItem[T]()` -- fetch + unmarshal single item
- `putItem()` -- marshal + write item
- `putItemWithCondition()` -- conditional write
- `queryIndex[T]()` -- query GSI + unmarshal results
- `querySingleItem[T]()` -- query + return first result
- `stringKey()`, `stringVal()`, `numVal()` -- key/value builders

Each entity has its own repository file (e.g., `users_repository.go`) that uses these generic helpers.

## Go Types (`internal/types/types.go`)

All struct fields use dual tags: `json:"snake_case" dynamodbav:"snake_case"`.

Key types: `User`, `Account`, `UserAccount`, `Permissions`, `AuthContext`, `Platform`, `PlatformConnection`, `PlatformConnectionAuth`, `Helper`, `Execution`, `APIKey`, `PlanLimits`.

## CRM Connector System (`internal/connectors/`)

### Interface

```go
type CRMConnector interface {
    GetContacts(ctx, opts) (*ContactList, error)
    GetContact(ctx, contactID) (*NormalizedContact, error)
    CreateContact(ctx, input) (*NormalizedContact, error)
    UpdateContact(ctx, contactID, updates) (*NormalizedContact, error)
    DeleteContact(ctx, contactID) error
    GetTags(ctx) ([]Tag, error)
    ApplyTag(ctx, contactID, tagID) error
    RemoveTag(ctx, contactID, tagID) error
    GetCustomFields(ctx) ([]CustomField, error)
    GetContactFieldValue(ctx, contactID, fieldKey) (interface{}, error)
    SetContactFieldValue(ctx, contactID, fieldKey, value) error
    TriggerAutomation(ctx, contactID, automationID) error
    AchieveGoal(ctx, contactID, goalName, integration) error
    TestConnection(ctx) error
    GetMetadata() ConnectorMetadata
    GetCapabilities() []Capability
}
```

### Implementations

Each CRM has its own file: `keap.go`, `gohighlevel.go`, `activecampaign.go`, `ontraport.go`. They translate platform-specific APIs into the normalized `CRMConnector` interface.

The `translate/` directory handles cross-platform data normalization (field mapping, custom field resolution, tag resolution).

## Helper System (`internal/helpers/`)

### Interface

```go
type Helper interface {
    GetName() string
    GetType() string
    GetCategory() string
    GetDescription() string
    GetConfigSchema() map[string]interface{}
    Execute(ctx, input HelperInput) (*HelperOutput, error)
    ValidateConfig(config map[string]interface{}) error
    RequiresCRM() bool
    SupportedCRMs() []string
}
```

### Registry

Helpers self-register via `init()` functions:
```go
func init() {
    helpers.Register("tag_it", func() helpers.Helper { return &TagIt{} })
}
```

### Categories and Examples

| Category | Helpers |
|----------|---------|
| contact | tag_it, copy_it, merge_it, move_it, name_parse_it, note_it, assign_it, clear_it, combine_it, found_it, opt_in, opt_out, own_it, snapshot_it, field_to_field, default_to_field, company_link |
| data | format_it, math_it, split_it, text_it, date_calc, when_is_it, word_count_it, password_it, phone_lookup, ip_location, get_the_first, get_the_last, last_click_it, last_open_it, last_send_it |
| tagging | score_it, group_it, count_tags, count_it_tags, clear_tags |
| automation | trigger_it, action_it, chain_it, drip_it, goal_it, stage_it, timezone_triggers |
| integration | hook_it, mail_it, slack_it, twilio_sms, zoom_webinar, calendly_it, email_validate_it, excel_it, google_sheet_it |
| notification | notify_me, email_engagement |
| analytics | rfm_calculation, customer_lifetime_value |

## Serverless Framework Conventions

### Plugin

Uses `serverless-go-plugin` (installed via `npm install` in `backend/golang/`):
```yaml
plugins:
  - serverless-go-plugin

custom:
  go:
    baseDir: ../../..
    supportedRuntimes: ["provided.al2023"]
    buildProvidedRuntimeAsBootstrap: true
    cmd: 'GOARCH=arm64 GOOS=linux go build -ldflags="-s -w"'
```

### Service Naming

All services use the `mfh-` prefix:
- Infrastructure: `mfh-infrastructure-cognito`, `mfh-infrastructure-dynamodb-core`, `mfh-infrastructure-s3`, `mfh-infrastructure-sqs`
- API Gateway: `mfh-api-gateway`
- API services: `mfh-auth`, `mfh-accounts`, `mfh-api-keys`, `mfh-helpers`, `mfh-platforms`, `mfh-data-explorer`, `mfh-billing`
- Workers: `mfh-helper-worker`, `mfh-data-sync`

### Cross-Stack References

Services reference each other's CloudFormation exports:
```yaml
# Reference API Gateway ID from gateway service
httpApi:
  id: ${cf:mfh-api-gateway-${self:provider.stage}.HttpApiId}

# Reference DynamoDB table from infrastructure
USERS_TABLE: ${cf:mfh-infrastructure-dynamodb-core-${self:provider.stage}.UsersTableName}

# Reference Cognito from infrastructure
COGNITO_USER_POOL_ID: ${cf:mfh-infrastructure-cognito-${self:provider.stage}.CognitoUserPoolId}
```

### Lambda Configuration

- Runtime: `provided.al2023`
- Architecture: `arm64` (Graviton)
- Default memory: 512 MB (health checks: 128 MB, auth: 256 MB)
- Default timeout: 29 seconds
- X-Ray tracing enabled

### DuckDB / CGO Services

The data-explorer service uses DuckDB (CGO) and must be built in an AL2023 Docker container for glibc compatibility:
```bash
docker build -t mfh-lambda-builder -f docker/Dockerfile.lambda-builder .
bash scripts/build-duckdb-handler.sh
```

## Environment Variables

Common env vars set by Serverless on all Lambda functions:
- `STAGE` -- deployment stage (dev/prod)
- `COGNITO_USER_POOL_ID`, `COGNITO_CLIENT_ID`, `COGNITO_REGION`
- `USERS_TABLE`, `ACCOUNTS_TABLE`, `USER_ACCOUNTS_TABLE`
- Service-specific tables as needed
- Billing service: `STRIPE_SECRET_KEY`, `STRIPE_WEBHOOK_SECRET`, `STRIPE_PRICE_START`, `STRIPE_PRICE_GROW`, `STRIPE_PRICE_DELIVER`, `APP_URL` (from SSM)

## CI/CD

GitHub Actions workflows:

**`deploy-backend.yml`** -- main deploy pipeline:
- Triggered on push to `main`/`dev` (paths: `backend/golang/**`)
- OIDC authentication: `GitHubActions-Deploy-Dev` IAM role
- Branch mapping: `dev` -> dev stage, `main` -> prod stage
- API services deploy with `max-parallel: 3` to avoid CloudFormation throttling
- Post-deploy: seeds platform data + runs health check verification

**`sync-internal-secrets.yml`** -- manual secrets sync:
- `workflow_dispatch` triggered, syncs GitHub secrets to AWS SSM Parameter Store
- Writes Stripe keys per stage (`/{stage}/stripe/secret_key`, etc.)
