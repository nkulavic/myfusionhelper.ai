# MyFusion Helper - Go Backend

Serverless Go backend on AWS Lambda, deployed via Serverless Framework v4.

## CRITICAL DEPLOYMENT POLICY

**ALL DEPLOYMENTS MUST GO THROUGH CI/CD PIPELINE**

- Push code to `dev` or `main` branch -> GitHub Actions deploys automatically
- NEVER run `npx sls deploy` manually (except for emergency debugging)
- NEVER deploy to any region other than `us-west-2`

**Region Lock**: ALL infrastructure and services are deployed ONLY to **us-west-2**. This is enforced at three levels:
1. **IAM Policy**: Claude Code's `mfh-claude` profile has an explicit Deny on all regional services outside us-west-2
2. **CI/CD**: GitHub Actions passes `--region us-west-2` on every command
3. **serverless.yml**: Every service specifies `region: us-west-2`

## AWS CLI Profile

**CRITICAL**: Always use the `mfh-claude` profile for ALL AWS commands. This is the only profile Claude Code is authorized to use. There is no default profile.

```bash
export AWS_PROFILE=mfh-claude
```

This profile is hard-locked to **us-west-2 only** via IAM policy. See root `CLAUDE.md` for full details.

## Quick Reference

```bash
cd backend/golang

# Build all handlers
go build ./...

# Run tests
go test ./...

# Build DuckDB-dependent service (needs Docker for AL2023 glibc)
bash scripts/build-duckdb-handler.sh
```

## Architecture Overview

The backend uses two distinct patterns:

1. **API Services** -- consolidated Lambda handlers that route HTTP requests (`cmd/handlers/`)
2. **Helper Workers** -- self-contained microservices, one per helper type (`services/workers/*-worker/`)

### Execution Flow

```
User triggers execution via API
  -> Execution record written to DynamoDB (status: "pending")
  -> DynamoDB Stream fires
  -> Stream Router Lambda reads stream event
  -> Stream Router constructs queue URL from helper_type using naming convention
  -> SQS FIFO message sent to helper-specific queue
  -> Helper Worker Lambda processes the job
  -> Execution record updated with results
```

## Project Layout

```
backend/golang/
├── cmd/handlers/                  # API Lambda entry points (consolidated routers)
│   ├── auth/                      # /auth/* endpoints
│   ├── accounts/                  # /accounts/* endpoints
│   ├── api-keys/                  # /api-keys/* endpoints
│   ├── helpers/                   # /helpers/* + /executions/* endpoints
│   ├── platforms/                 # /platforms/* + /platform-connections/*
│   ├── billing/                   # /billing/* endpoints (Stripe integration)
│   ├── data-explorer/             # /data/* endpoints (DuckDB + Parquet)
│   ├── data-sync/                 # SQS worker: sync CRM data -> S3/Parquet
│   └── data-sync-scheduler/       # EventBridge trigger for data-sync
│
├── internal/
│   ├── worker/                    # Shared worker handler (used by ALL helper workers)
│   │   └── handler.go             # HandleSQSEvent -- SQS message parsing, execution
│   ├── connectors/                # CRM platform adapters
│   │   ├── interface.go           # CRMConnector interface
│   │   ├── keap.go, gohighlevel.go, activecampaign.go, ontraport.go
│   │   ├── models.go              # NormalizedContact, Tag, CustomField
│   │   ├── registry.go            # Connector factory registry
│   │   ├── loader/loader.go       # Load connector with credentials
│   │   └── translate/             # Data normalization layer
│   ├── database/                  # DynamoDB repositories
│   ├── helpers/                   # Helper implementations
│   │   ├── interface.go           # Helper interface
│   │   ├── registry.go            # Helper factory registry
│   │   ├── executor.go            # Execution orchestrator
│   │   ├── contact/               # 17 helpers: tag_it, copy_it, merge_it, etc.
│   │   ├── data/                  # 18 helpers: format_it, math_it, split_it, etc.
│   │   ├── tagging/               # 6 helpers: tag_it, score_it, group_it, etc.
│   │   ├── automation/            # 22 helpers: trigger_it, action_it, chain_it, etc.
│   │   ├── integration/           # 29 helpers: hook_it, slack_it, zoom_webinar, etc.
│   │   ├── notification/          # 2 helpers: notify_me, email_engagement
│   │   ├── analytics/             # 2 helpers: rfm_calculation, customer_lifetime_value
│   │   └── platform/              # 1 helper: keap_backup
│   ├── middleware/auth/           # JWT auth middleware
│   ├── config/                    # Secrets loading (SSM)
│   ├── services/parquet/          # Parquet file writer for data sync
│   └── types/types.go             # All shared Go types
│
├── services/
│   ├── api/                       # API Gateway + HTTP API services
│   │   ├── gateway/               # Shared API Gateway + Cognito authorizer
│   │   ├── auth/, accounts/, api-keys/, helpers/, platforms/, billing/
│   │   └── data-explorer/
│   ├── infrastructure/            # Core AWS resources
│   │   ├── cognito/               # User pool
│   │   ├── dynamodb/core/         # All DynamoDB tables
│   │   ├── s3/                    # Data bucket
│   │   ├── sqs/                   # Notification queue + fallback queue
│   │   ├── ses/                   # Email sending
│   │   ├── monitoring/            # CloudWatch alarms
│   │   └── acm/                   # SSL certificates
│   └── workers/                   # ALL worker services
│       ├── stream-router/         # Routes DynamoDB stream events to SQS queues
│       ├── tag-it-worker/         # REFERENCE IMPLEMENTATION for helper workers
│       ├── copy-it-worker/        # Each helper has its own self-contained worker
│       ├── score-it-worker/
│       ├── ... (97 total helper workers)
│       ├── data-sync/             # Non-helper: CRM data sync
│       ├── notification-worker/   # Non-helper: send notifications
│       └── helper-worker/         # DEPRECATED: old monolith (fallback only)
│
├── go.mod, go.sum
└── package.json                   # serverless-go-plugin dependency
```

---

## Helper Worker Architecture (Self-Contained Microservices)

Every helper type runs as its own self-contained Serverless Framework service. Each service creates its own SQS FIFO queue, DLQ, and Lambda function in a single `serverless.yml`.

### Reference Implementation: tag-it-worker

**ALL helper workers MUST follow this exact pattern.**

#### Directory Structure
```
services/workers/tag-it-worker/
├── main.go           # Registers ONE helper, starts SQS handler
└── serverless.yml    # Self-contained: queue + DLQ + Lambda + IAM
```

#### main.go Pattern
```go
package main

import (
    "github.com/aws/aws-lambda-go/lambda"
    "github.com/myfusionhelper/api/internal/helpers"
    "github.com/myfusionhelper/api/internal/helpers/tagging"
    "github.com/myfusionhelper/api/internal/worker"

    // Register all connectors via init()
    _ "github.com/myfusionhelper/api/internal/connectors"
)

func main() {
    helpers.Register("tag_it", tagging.NewTagIt)
    lambda.Start(worker.HandleSQSEvent)
}
```

Key points:
- Import `internal/worker` for the shared `HandleSQSEvent` handler
- Import the specific helper package (e.g., `tagging`, `contact`, `data`, `automation`, `integration`)
- Import `_ "github.com/myfusionhelper/api/internal/connectors"` for CRM connector registration
- Call `helpers.Register()` with the helper_type string and the exported factory function
- Call `lambda.Start(worker.HandleSQSEvent)` -- the shared handler does all the work

#### Factory Function Pattern

Each helper file exports a factory function used by its worker:

```go
// In internal/helpers/tagging/tag_it.go
func NewTagIt() helpers.Helper { return &TagIt{} }

func init() {
    helpers.Register("tag_it", func() helpers.Helper { return &TagIt{} })
}
```

The `NewXxx()` factory is used by individual workers. The `init()` registration is kept for backward compatibility with the monolith.

#### serverless.yml Pattern

```yaml
service: mfh-tag-it-worker
frameworkVersion: '4'

provider:
  name: aws
  runtime: provided.al2023
  architecture: arm64
  region: us-west-2
  stage: ${opt:stage, 'dev'}
  memorySize: 256
  timeout: 300
  tracing:
    lambda: true
  environment:
    STAGE: ${self:provider.stage}
    HELPER_TYPE: tag_it
    COGNITO_REGION: ${self:provider.region}
    EXECUTIONS_TABLE: ${cf:mfh-infrastructure-dynamodb-core-${self:provider.stage}.ExecutionsTableName}
    HELPERS_TABLE: ${cf:mfh-infrastructure-dynamodb-core-${self:provider.stage}.HelpersTableName}
    CONNECTIONS_TABLE: ${cf:mfh-infrastructure-dynamodb-core-${self:provider.stage}.ConnectionsTableName}
    PLATFORMS_TABLE: ${cf:mfh-infrastructure-dynamodb-core-${self:provider.stage}.PlatformsTableName}
    PLATFORM_CONNECTION_AUTHS_TABLE: ${cf:mfh-infrastructure-dynamodb-core-${self:provider.stage}.PlatformConnectionAuthsTableName}
    ACCOUNTS_TABLE: ${cf:mfh-infrastructure-dynamodb-core-${self:provider.stage}.AccountsTableName}
    NOTIFICATION_QUEUE_URL: ${cf:mfh-infrastructure-sqs-${self:provider.stage}.NotificationQueueUrl}
    INTERNAL_SECRETS_PARAM: /myfusionhelper/${self:provider.stage}/secrets
  iam:
    role:
      statements:
        - Effect: Allow
          Action:
            - dynamodb:GetItem
            - dynamodb:PutItem
            - dynamodb:UpdateItem
            - dynamodb:Query
            - dynamodb:BatchGetItem
          Resource:
            - ${cf:mfh-infrastructure-dynamodb-core-${self:provider.stage}.ExecutionsTableArn}
            - ${cf:mfh-infrastructure-dynamodb-core-${self:provider.stage}.HelpersTableArn}
            - ${cf:mfh-infrastructure-dynamodb-core-${self:provider.stage}.ConnectionsTableArn}
            - ${cf:mfh-infrastructure-dynamodb-core-${self:provider.stage}.PlatformsTableArn}
            - ${cf:mfh-infrastructure-dynamodb-core-${self:provider.stage}.PlatformConnectionAuthsTableArn}
            - ${cf:mfh-infrastructure-dynamodb-core-${self:provider.stage}.AccountsTableArn}
            - "${cf:mfh-infrastructure-dynamodb-core-${self:provider.stage}.HelpersTableArn}/index/*"
            - "${cf:mfh-infrastructure-dynamodb-core-${self:provider.stage}.ExecutionsTableArn}/index/*"
            - "${cf:mfh-infrastructure-dynamodb-core-${self:provider.stage}.ConnectionsTableArn}/index/*"
            - "${cf:mfh-infrastructure-dynamodb-core-${self:provider.stage}.PlatformsTableArn}/index/*"
            - "${cf:mfh-infrastructure-dynamodb-core-${self:provider.stage}.PlatformConnectionAuthsTableArn}/index/*"
        - Effect: Allow
          Action:
            - sqs:SendMessage
          Resource:
            - ${cf:mfh-infrastructure-sqs-${self:provider.stage}.NotificationQueueArn}
        - Effect: Allow
          Action:
            - ssm:GetParameter
          Resource:
            - "arn:aws:ssm:${self:provider.region}:*:parameter/myfusionhelper/${self:provider.stage}/secrets"

functions:
  worker:
    handler: services/workers/tag-it-worker/main.go
    description: "Process tag_it helper execution jobs"
    events:
      - sqs:
          arn: !GetAtt HelperQueue.Arn
          batchSize: 1
          functionResponseType: ReportBatchItemFailures

resources:
  Resources:
    HelperQueue:
      Type: AWS::SQS::Queue
      Properties:
        QueueName: mfh-${self:provider.stage}-tag-it-executions.fifo
        FifoQueue: true
        ContentBasedDeduplication: true
        VisibilityTimeout: 360
        MessageRetentionPeriod: 1209600
        RedrivePolicy:
          deadLetterTargetArn: !GetAtt HelperDLQ.Arn
          maxReceiveCount: 3

    HelperDLQ:
      Type: AWS::SQS::Queue
      Properties:
        QueueName: mfh-${self:provider.stage}-tag-it-dlq.fifo
        FifoQueue: true
        ContentBasedDeduplication: true
        MessageRetentionPeriod: 1209600

  Outputs:
    HelperQueueArn:
      Value: !GetAtt HelperQueue.Arn
    HelperQueueUrl:
      Value: !Ref HelperQueue

plugins:
  - serverless-go-plugin

custom:
  go:
    baseDir: ../../..
    cmd: 'GOARCH=arm64 GOOS=linux go build -ldflags="-s -w"'
    supportedRuntimes: ["provided.al2023"]
    buildProvidedRuntimeAsBootstrap: true
```

### CRITICAL: serverless-go-plugin Configuration

These settings are learned from deployment failures and MUST NOT be changed:

| Setting | Value | Why |
|---------|-------|-----|
| `cmd` | `'GOARCH=arm64 GOOS=linux go build -ldflags="-s -w"'` | Must NOT include `-o bootstrap` (plugin handles output naming) |
| `handler` | `services/workers/tag-it-worker/main.go` | Must be Go source path relative to baseDir, NOT `bootstrap` |
| `baseDir` | `../../..` | Points to `backend/golang/` from `services/workers/xxx-worker/` |
| `buildProvidedRuntimeAsBootstrap` | `true` | Plugin builds output as `bootstrap` automatically |

### Naming Conventions

| Item | Convention | Example |
|------|-----------|---------|
| Worker directory | `{kebab-name}-worker/` | `tag-it-worker/` |
| Serverless service | `mfh-{kebab-name}-worker` | `mfh-tag-it-worker` |
| SQS queue | `mfh-{stage}-{kebab-name}-executions.fifo` | `mfh-dev-tag-it-executions.fifo` |
| SQS DLQ | `mfh-{stage}-{kebab-name}-dlq.fifo` | `mfh-dev-tag-it-dlq.fifo` |
| Lambda function | `mfh-{kebab-name}-worker-{stage}-worker` | `mfh-tag-it-worker-dev-worker` |
| HELPER_TYPE env | `{snake_case}` | `tag_it` |
| Factory function | `New{PascalCase}()` | `NewTagIt()` |

**Kebab conversion**: helper_type `tag_it` -> kebab name `tag-it` (replace `_` with `-`)

### Stream Router (Convention-Based URL Discovery)

The stream router (`services/workers/stream-router/`) reads DynamoDB Stream events from the Executions table and routes them to individual helper SQS queues.

**Queue URL construction** (no SSM parameters needed):
```go
func buildQueueURL(helperType string) string {
    queueName := strings.ReplaceAll(helperType, "_", "-")
    return fmt.Sprintf("https://sqs.%s.amazonaws.com/%s/mfh-%s-%s-executions.fifo",
        region, accountID, stage, queueName)
}
```

The stream router has a **fallback queue** for helpers that don't have individual workers yet. Once all helpers are deployed, this fallback will be removed.

### How to Add a New Helper Worker

1. Create the helper implementation in `internal/helpers/{category}/my_helper.go`
2. Add the factory function: `func NewMyHelper() helpers.Helper { return &MyHelper{} }`
3. Add the `init()` registration: `helpers.Register("my_helper", func() helpers.Helper { return &MyHelper{} })`
4. Create `services/workers/my-helper-worker/main.go` (copy tag-it-worker pattern, change import + Register call)
5. Create `services/workers/my-helper-worker/serverless.yml` (copy tag-it-worker, change 6 values: service, HELPER_TYPE, handler, description, QueueName x2)
6. Push to `dev` -- CI/CD auto-detects the new worker directory and deploys it

### Complete Helper Inventory (97 helpers)

**Tagging** (6): tag_it, clear_tags, count_it_tags, count_tags, group_it, score_it

**Contact** (17): assign_it, clear_it, combine_it, company_link, contact_updater, copy_it, default_to_field, field_to_field, found_it, merge_it, move_it, name_parse_it, note_it, opt_in, opt_out, own_it, snapshot_it

**Data** (18): advance_math, date_calc, format_it, get_the_first, get_the_last, ip_location, last_click_it, last_open_it, last_send_it, math_it, password_it, phone_lookup, quote_it, split_it, split_it_basic, text_it, when_is_it, word_count_it

**Automation** (22): action_it, chain_it, countdown_timer, drip_it, goal_it, ip_notifications, ip_redirects, limit_it, match_it, route_it, route_it_by_custom, route_it_by_day, route_it_by_time, route_it_geo, route_it_score, route_it_source, simple_opt_in, simple_opt_out, stage_it, timezone_triggers, trigger_it, video_trigger_it

**Integration** (29): calendly_it, donor_search, dropbox_it, email_attach_it, email_validate_it, everwebinar, excel_it, facebook_lead_ads, google_sheet_it, gotowebinar, hook_it, hook_it_by_tag, hook_it_v2, hook_it_v3, hook_it_v4, mail_it, order_it, query_it_basic, search_it, slack_it, stripe_hooks, trello_it, twilio_sms, upload_it, webinar_jam, zoom_meeting, zoom_webinar, zoom_webinar_absentee, zoom_webinar_participant

**Notification** (2): email_engagement, notify_me

**Analytics** (2): customer_lifetime_value, rfm_calculation

**Platform** (1): keap_backup

---

## API Handler Pattern

Each API service uses a **consolidated handler** pattern: one Lambda binary routes to multiple endpoint handlers based on path.

```go
// cmd/handlers/auth/main.go
func Handle(ctx context.Context, event events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
    switch event.RequestContext.HTTP.Path {
    case "/auth/login":
        return loginClient.Handle(ctx, event)
    case "/auth/status":
        return routeToProtectedHandler(ctx, event, statusClient.HandleWithAuth)
    }
}
```

**Public endpoints**: Use `Handle(ctx, event)` signature directly.
**Protected endpoints**: Use `HandleWithAuth(ctx, event, authCtx)` signature, wrapped by `routeToProtectedHandler`.

## Auth Middleware (`internal/middleware/auth/auth.go`)

The `WithAuth` wrapper:
1. Extracts `sub` claim from JWT (tries API Gateway authorizer context first, falls back to Bearer token)
2. Constructs `userID` as `"user:" + sub`
3. Fetches user from DynamoDB to get `current_account_id`
4. Fetches user-account relationship for permissions
5. Builds `AuthContext` and passes to handler

**Important**: User IDs are prefixed with `user:` (e.g., `user:abc-123-def`).

## Response Helpers

```go
import authMiddleware "github.com/myfusionhelper/api/internal/middleware/auth"

return authMiddleware.CreateSuccessResponse(200, "OK", data), nil
return authMiddleware.CreateErrorResponse(400, "Bad request"), nil
```

All responses include CORS headers (`Access-Control-Allow-Origin: *`).

## DynamoDB Conventions

### Table Names

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

Most tables have an `AccountIdIndex` GSI. Other notable GSIs:
- Users: `EmailIndex`, `CognitoUserIdIndex`
- Accounts: `OwnerUserIdIndex`
- API Keys: `KeyHashIndex`
- Executions: `AccountIdCreatedAtIndex`, `HelperIdCreatedAtIndex`
- Platforms: `SlugIndex`
- Platform Connection Auths: `ConnectionIdIndex`

### Database Client (`internal/database/client.go`)

Generic helpers: `getItem[T]()`, `putItem()`, `putItemWithCondition()`, `queryIndex[T]()`, `querySingleItem[T]()`, `stringKey()`, `stringVal()`, `numVal()`.

## Go Types (`internal/types/types.go`)

All struct fields use dual tags: `json:"snake_case" dynamodbav:"snake_case"`.

Key types: `User`, `Account`, `UserAccount`, `Permissions`, `AuthContext`, `Platform`, `PlatformConnection`, `PlatformConnectionAuth`, `Helper`, `Execution`, `APIKey`, `PlanLimits`.

## CRM Connector System (`internal/connectors/`)

Implements `CRMConnector` interface for: Keap, GoHighLevel, ActiveCampaign, Ontraport, HubSpot.

Connectors register via `init()` and are loaded dynamically based on platform type. The `translate/` directory handles cross-platform data normalization.

## Serverless Framework Conventions

### Critical Requirements

**EVERY serverless.yml file MUST include**:
```yaml
provider:
  name: aws
  region: us-west-2
  runtime: provided.al2023
  architecture: arm64
```

### Plugin Configuration (serverless-go-plugin)

```yaml
plugins:
  - serverless-go-plugin

custom:
  go:
    baseDir: ../../..
    cmd: 'GOARCH=arm64 GOOS=linux go build -ldflags="-s -w"'
    supportedRuntimes: ["provided.al2023"]
    buildProvidedRuntimeAsBootstrap: true
```

**CRITICAL**: `cmd` must NOT include `-o bootstrap`. The plugin handles output naming. The `handler` must be the Go source file path relative to baseDir (e.g., `services/workers/tag-it-worker/main.go`), NOT `bootstrap`.

### Service Naming

- Infrastructure: `mfh-infrastructure-{name}` (cognito, dynamodb-core, s3, sqs, ses, monitoring, acm)
- API Gateway: `mfh-api-gateway`
- API services: `mfh-{name}` (auth, accounts, api-keys, helpers, platforms, data-explorer, billing)
- Helper workers: `mfh-{kebab-name}-worker` (tag-it-worker, copy-it-worker, etc.)
- Other workers: `mfh-stream-router`, `mfh-data-sync`, `mfh-notification-worker`

### Cross-Stack References

```yaml
httpApi:
  id: ${cf:mfh-api-gateway-${self:provider.stage}.HttpApiId}
USERS_TABLE: ${cf:mfh-infrastructure-dynamodb-core-${self:provider.stage}.UsersTableName}
COGNITO_USER_POOL_ID: ${cf:mfh-infrastructure-cognito-${self:provider.stage}.CognitoUserPoolId}
```

## Secrets Architecture

Two categories of secrets exist:

### 1. Unified Secrets (Internal API Keys)

All internal API secrets consolidated into ONE SSM parameter per stage:

**Parameter**: `/myfusionhelper/{stage}/secrets` (SecureString, KMS-encrypted)
**Source**: GitHub Secrets → `sync-internal-secrets.yml` workflow → `scripts/build-internal-secrets.sh`

**JSON structure**:
```json
{
  "stripe": { "secret_key", "publishable_key", "webhook_secret", "price_start", "price_grow", "price_deliver" },
  "groq": { "api_key" },
  "twilio": { "account_sid", "auth_token", "from_number", "messaging_sid" }
}
```

**Go code** (`internal/config/secrets.go`):
```go
secrets, err := config.LoadSecrets(ctx) // Singleton, cached after first call
stripeKey := secrets.Stripe.SecretKey
groqKey := secrets.Groq.APIKey
twilioSID := secrets.Twilio.AccountSID
```

**serverless.yml** (required for any service using secrets):
```yaml
environment:
  INTERNAL_SECRETS_PARAM: /myfusionhelper/${self:provider.stage}/secrets
iam:
  role:
    statements:
      - Effect: Allow
        Action: [ssm:GetParameter]
        Resource: "arn:aws:ssm:${self:provider.region}:*:parameter/myfusionhelper/${self:provider.stage}/secrets"
```

**To sync secrets**: Run `sync-internal-secrets.yml` workflow manually in GitHub Actions, selecting the target stage.

### 2. OAuth Credentials (Platform Connections)

Platform OAuth credentials (for CRM connections like Keap) are stored as **separate SSM parameters**:

**Pattern**:
```
/{stage}/platforms/{platform_slug}/oauth/client_id     (SecureString)
/{stage}/platforms/{platform_slug}/oauth/client_secret  (SecureString)
```

**Loaded by**: `loadOAuthCredentials()` in `cmd/handlers/platforms/clients/connections/main.go`

**OAuth flow** (Keap example):
1. `POST /platforms/{id}/oauth/start` → generates state token, stores in `mfh-{stage}-oauth-states` table (TTL: 15 min), returns authorization URL
2. User authorizes at Keap → redirected to `GET /platforms/oauth/callback?code=...&state=...`
3. Backend validates state (one-time use), exchanges code for tokens
4. Stores connection in `connections` table + auth record in `platform-connection-auths` table
5. User redirected to success URL with `connection_id`

**Key DynamoDB tables for OAuth**:
- `mfh-{stage}-oauth-states` -- temporary state tokens (DynamoDB TTL auto-cleanup)
- `mfh-{stage}-connections` -- platform connections (connection_id, account_id, platform_id, auth_type)
- `mfh-{stage}-platform-connection-auths` -- OAuth tokens (access_token, refresh_token, expires_at)

### Legacy SSM Parameters (Cleanup Needed)

These old parameters still exist but are **no longer used** by current code:
- `/dev/stripe/price_*` -- superseded by unified JSON
- `/mfh/dev/sqs/*/queue-url` -- superseded by convention-based queue URL construction

## CI/CD Pipeline

See root `CLAUDE.md` for full CI/CD documentation with all 3 workflows, GitHub secrets inventory, and job dependency graph.

**Quick reference**:
- **Trigger**: Push to `main`/`dev` branches (paths: `backend/golang/**`)
- **Auth**: OIDC → `GitHubActions-Deploy-Dev` IAM role (account 570331155915)
- **Region**: `us-west-2` hardcoded in workflow

### Deployment Order

1. **Build & test** (Go 1.23, `CGO_ENABLED=1 go build ./...`)
2. **Infrastructure** (parallel): cognito, dynamodb-core, s3, sqs, ses, monitoring, acm
3. **Pre-gateway** (parallel): api-key-authorizer, scheduler, executions-stream, stream-router
4. **API Gateway** (creates HttpApi + Cognito authorizer + custom domain)
5. **API services** (parallel, max 3): auth, accounts, api-keys, helpers, platforms, data-explorer, billing, chat, emails
6. **Helper workers** (parallel, max 10): auto-detected from changed `services/workers/*-worker/` directories
7. **Non-helper workers** (parallel): helper-worker (monolith), notification-worker, data-sync
8. **Post-deploy**: `verify-deploy.sh` health check

### Helper Worker Auto-Detection

The CI pipeline uses `git diff` to detect which worker directories changed and deploys only those. On manual trigger, ALL helpers are deployed. Excludes: helper-worker, notification-worker, data-sync, executions-stream, scheduler, voice assistant webhooks.

### Deployment Workflow

```bash
git add .
git commit -m "feat: add new helper"
git push origin dev
# GitHub Actions deploys automatically to us-west-2
```

## Known Cleanup Items

- `_archive/` directory (empty, can be deleted)
- `integration/countdown_timer.go` -- duplicate of `automation/countdown_timer.go` (canonical is automation)
- `analytics/last_click_it.go` -- duplicate of `data/last_click_it.go` (canonical is data)
- `services/workers/helper-worker/` -- deprecated monolith, kept as fallback until all workers verified
- Old category SQS infrastructure stacks have been **deleted** (mfh-sqs-contact-helpers-dev, mfh-sqs-data-helpers-dev, mfh-sqs-automation-helpers-dev, mfh-sqs-integration-helpers-dev, mfh-sqs-analytics-helpers-dev, mfh-sqs-tagging-helpers-dev) -- queues are now self-contained in each worker's serverless.yml
