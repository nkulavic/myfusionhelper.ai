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
│   ├── data-sync-scheduler/       # EventBridge trigger for data-sync
│   ├── internal-email/            # Internal email service (template rendering + SES)
│   └── trial-expiration/          # EventBridge worker: expire trial accounts + send email
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
│   ├── email/                     # S3 Liquid template loader, renderer, bindings, SES client
│   ├── notifications/             # Internal email service HTTP client (Send* methods)
│   ├── billing/                   # Plan configs, limit enforcement, helpers
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
│       ├── notification-worker/   # Non-helper: SQS → email dispatch (welcome, password_reset, billing, etc.)
│       ├── trial-expiration/      # Non-helper: EventBridge cron, expires trial accounts
│       └── helper-worker/         # DEPRECATED: old monolith (fallback only)
│
├── email-templates/               # S3 Liquid email templates (synced to s3://mfh-{stage}-data/email-templates/)
│   ├── welcome/                   # subject.liquid, body.html.liquid, meta.json
│   ├── password_reset/
│   ├── execution_alert/
│   ├── connection_alert/
│   ├── usage_alert/
│   ├── weekly_summary/
│   ├── team_invite/
│   └── billing_event/             # 12 sub-type subdirectories
│       ├── subscription_created/
│       ├── payment_failed/
│       ├── card_expiring/
│       └── ... (12 total)
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

## Password Reset (Self-Managed Verification Codes)

Password reset does **NOT** use Cognito's `ForgotPassword`/`ConfirmForgotPassword`. The system generates and manages its own 6-digit codes.

### Flow

```
POST /auth/forgot-password (public)
├─ Validate email
├─ Query Users table (EmailIndex) → check exists (silent failure for enumeration prevention)
├─ Expire prior pending codes (GetPendingByEmail → MarkAsExpired)
├─ generateSecureCode() → crypto/rand 6-digit code (0-padded)
├─ Create EmailVerification record (TTL: 15 minutes)
├─ Enqueue SQS: type="password_reset", data={user_email, user_name, reset_code}
└─ Return 200 (always, even if user doesn't exist)

POST /auth/reset-password (public)
├─ Validate email, code, new_password
├─ Query EmailVerifications (EmailIndex GSI) → GetPendingByEmail()
├─ Match code + validate not expired
├─ Cognito AdminSetUserPassword(permanent=true)
├─ MarkAsVerified() on the verification record
└─ Return 200 or error (400 invalid code, 429 rate limited, etc.)
```

### Key Files

- `cmd/handlers/auth/clients/forgot-password/main.go` — code generation + storage
- `cmd/handlers/auth/clients/reset-password/main.go` — code verification + password set
- `internal/database/email_verifications_repository.go` — DynamoDB CRUD

### EmailVerification Record

```go
type EmailVerification struct {
    VerificationID string  // "verify:" + UUIDv7
    Email          string  // EmailIndex GSI partition key
    Token          string  // 6-digit code
    ExpiresAt      int64   // Unix timestamp (15 min TTL)
    Status         string  // "pending" | "verified" | "expired"
    CreatedAt      string  // RFC3339
    VerifiedAt     string  // RFC3339 (set on successful reset)
}
```

### Security

- `crypto/rand` for code generation (not `math/rand`)
- 15-minute expiry (DynamoDB TTL auto-cleanup)
- One active code per email (prior codes expired on new request)
- Always returns 200 from forgot-password (prevents email enumeration)
- Password validation delegated to Cognito (8+ chars, mixed case, numbers, symbols)

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
| `WEBHOOK_EVENTS_TABLE` | `mfh-{stage}-webhook-events` | `event_id` | -- |
| `EMAIL_VERIFICATIONS_TABLE` | `mfh-{stage}-email-verifications` | `verification_id` | -- |
| `EMAIL_TEMPLATES_TABLE` | `mfh-{stage}-email-templates` | `template_id` | -- |
| `EMAIL_LOGS_TABLE` | `mfh-{stage}-email-logs` | `email_id` | -- |

### GSI Patterns

Most tables have an `AccountIdIndex` GSI. Other notable GSIs:
- Users: `EmailIndex`, `CognitoUserIdIndex`
- Accounts: `OwnerUserIdIndex`, `StripeCustomerIdIndex`
- API Keys: `KeyHashIndex`
- Executions: `AccountIdCreatedAtIndex`, `HelperIdCreatedAtIndex`
- Platforms: `SlugIndex`
- Platform Connection Auths: `ConnectionIdIndex`
- Email Verifications: `EmailIndex` (used by password reset to query pending codes by email)
- Email Logs: `AccountIdIndex`, `RecipientEmailIndex`

### Database Client (`internal/database/client.go`)

Generic helpers: `getItem[T]()`, `putItem()`, `putItemWithCondition()`, `queryIndex[T]()`, `querySingleItem[T]()`, `stringKey()`, `stringVal()`, `numVal()`.

## Go Types (`internal/types/types.go`)

All struct fields use dual tags: `json:"snake_case" dynamodbav:"snake_case"`.

Key types: `User`, `Account`, `UserAccount`, `Permissions`, `AuthContext`, `Platform`, `PlatformConnection`, `PlatformConnectionAuth`, `Helper`, `Execution`, `APIKey`, `PlanLimits`.

**Account trial fields**: `TrialStartedAt *time.Time`, `TrialEndsAt *time.Time`, `TrialExpired bool`. Set at registration when plan is `"trial"`.

## Billing & Trial Enforcement (`internal/billing/`)

**Plans** (`plans.go`): `Plans` map with configs for `"trial"`, `"free"`, `"start"`, `"grow"`, `"deliver"`. Key helpers: `GetPlan()`, `GetPlanLabel()`, `IsTrialPlan()` (returns true for `"trial"` and `"free"`), `IsPaidPlan()`.

**Plan ranking** (used for upgrade/downgrade detection in webhook handler):
```go
var planRank = map[string]int{"free": 0, "trial": 0, "start": 1, "grow": 2, "deliver": 3}
```

**Enforcement** (`enforce.go`): All limit functions (`CheckHelperLimit`, `CheckConnectionLimit`, `CheckAPIKeyLimit`, `CheckExecutionLimit`) call `checkTrialExpired()` as an early return. Trial expiration checks both the `TrialExpired` flag and `TrialEndsAt` timestamp. Returns `LimitExceededError` with resource `"trial"`.

**Registration** (`cmd/handlers/auth/clients/register/main.go`): Creates Stripe customer at registration (no subscription). Sets `Plan: "trial"`, `TrialStartedAt`, `TrialEndsAt: +14 days`, `TrialExpired: false`, `Settings` from Start-level plan limits.

**Checkout** (`cmd/handlers/billing/clients/checkout/main.go`): Active trial users get remaining trial days on Stripe subscription (not fresh 14 days). Expired trials get no trial period.

**Trial Expiration Worker** (`cmd/handlers/trial-expiration/main.go`): EventBridge-triggered every 6 hours. Scans accounts table for `plan="trial" AND trial_expired=false AND trial_ends_at < now`, marks each as `trial_expired=true` with conditional update. Also sends `trial_expired` email via the notification service.

## Stripe Webhook System (`cmd/handlers/billing/clients/webhook/main.go`)

**Endpoint**: `POST /billing/webhook` (public, verified by Stripe signature)
**Setup**: `scripts/setup-stripe-webhooks.sh <stage>` -- idempotent script that creates/updates Stripe webhook endpoint and verifies all events are registered. Auto-triggered by CI when billing code or the script changes.

### Idempotency

Every webhook event is recorded in `mfh-{stage}-webhook-events` DynamoDB table via conditional PutItem (`attribute_not_exists(event_id)`). Duplicates return HTTP 200 immediately. Events are tracked with status (`pending` → `processed`/`failed`), timestamps, and 90-day TTL.

Key functions:
- `checkAndRecordEvent()` -- conditional PutItem, returns `(alreadySeen, error)`
- `markEventStatus()` -- UpdateItem to set final status + processed_at

### Race Condition Fix

`checkout.session.completed` and `customer.subscription.created` fire simultaneously from Stripe when a new subscription is created. Both update the same account record. The checkout handler sets `subscription_email_sent=true` on the account. The subscription.created handler checks this flag via `checkAndClearEmailSentFlag()` — if set, it skips the email and clears the flag. Both handlers write identical plan/limits/status, so the DynamoDB SET is naturally idempotent.

### Event Handlers (8 Stripe events)

| Stripe Event | Handler | Account Change | Email |
|---|---|---|---|
| `checkout.session.completed` | `handleCheckoutSessionCompleted` | Plan activated, metered item stored, Cognito synced, email_sent flag set | `subscription_created` |
| `customer.subscription.created` | `handleSubscriptionCreated` | Plan/limits updated | `subscription_created` (conditional on flag) |
| `customer.subscription.updated` | `handleSubscriptionUpdated` | Plan/limits updated, detects upgrade/downgrade | `plan_upgraded` or `plan_downgraded` |
| `customer.subscription.deleted` | `handleSubscriptionCancelled` | Downgrade to trial, mark expired, Cognito synced | `subscription_cancelled` |
| `customer.subscription.trial_will_end` | `handleTrialWillEnd` | None | `trial_ending` |
| `invoice.paid` | `handleInvoicePaid` | Reset `past_due` → `active` if applicable | `payment_recovered` (recovery) or `payment_receipt` (all payments) |
| `invoice.payment_failed` | `handlePaymentFailed` | None (Stripe retries) | `payment_failed` (with hosted invoice URL if available) |
| `charge.refunded` | `handleChargeRefunded` | None | `refund_processed` (amount, reason) |

### Shared Helpers

- `lookupAccountByStripeCustomer()` -- queries `StripeCustomerIdIndex` GSI, returns `accountLookupResult{AccountID, OwnerUserID, CurrentPlan, Status}`
- `updateAccountPlan()` -- updates plan, limits, status; clears `trial_expired` for paid plans
- `classifyPlanChange(oldPlan, newPlan)` -- returns `"plan_upgraded"`, `"plan_downgraded"`, or `""` using `planRank` map
- `sendBillingEmail()` -- looks up user by ownerUserID, calls notification service. Accepts variadic `extraData ...map[string]interface{}` for passing template-specific data (e.g., InvoiceURL, Amount, CardLast4)
- `formatStripeAmount()` -- converts Stripe integer cents to formatted string (e.g., `3900` → `"$39.00"`)
- `syncCognitoPlanGroup()` -- removes user from all `plan-*` Cognito groups, adds to new `plan-{name}` group

## Email & Notification System (S3 Liquid Templates)

### Architecture Overview

Emails are rendered using **Liquid templates stored in S3**, not hardcoded Go templates. The rendering pipeline uses `github.com/osteele/liquid` with a 5-minute template cache.

```
Service/Worker → SQS (NotificationQueue)
  → Notification Worker → notifications.Service.Send*()
  → HTTP POST to /internal/emails/send
  → ResolveTemplatePath(templateType, data) → S3 path
  → TemplateLoader.RenderTemplate(ctx, path, bindings) [cached 5min]
  → meta.json (header_title, cta_url, icon) rendered + injected
  → SES sends email
```

### Key Files

| Component | Path |
|-----------|------|
| Notification Worker | `cmd/handlers/notification-worker/main.go` |
| Internal Email Handler | `cmd/handlers/internal-email/clients/send/main.go` |
| Notifications Service (HTTP client) | `internal/notifications/notifications.go` |
| Template Loader (S3 + cache) | `internal/email/template_loader.go` |
| Template Rendering + Path Resolution | `internal/email/templates.go` |
| Template Bindings (Go → snake_case) | `internal/email/template_bindings.go` |
| SES Client | `internal/email/ses_client.go` |
| S3 Template Source Files | `email-templates/` (synced to S3) |

### S3 Template Structure

**Bucket**: `mfh-{stage}-data` (env var: `TEMPLATE_BUCKET`)
**Prefix**: `email-templates/{template_type}/`

Each template directory contains:
```
template-name/
├── subject.liquid          # Email subject (Liquid template)
├── body.html.liquid        # HTML email body (Liquid template)
├── body.txt.liquid         # Plain text alternative (optional)
└── meta.json              # Metadata (header_title, header_subtitle, icon, cta_text, cta_url)
```

**Directory tree**:
```
email-templates/
├── welcome/
├── password_reset/
├── execution_alert/
├── connection_alert/
├── usage_alert/
├── weekly_summary/
├── team_invite/
└── billing_event/
    ├── subscription_created/
    ├── subscription_cancelled/
    ├── payment_failed/
    ├── payment_recovered/
    ├── payment_receipt/
    ├── card_expiring/
    ├── refund_processed/
    ├── plan_upgraded/
    ├── plan_downgraded/
    ├── trial_ending/
    ├── trial_expired/
    └── default/
```

### Template Variables (Liquid Bindings)

All template variables use **snake_case** (converted from Go `TemplateData` via `TemplateDataToBindings()`):

| Liquid Variable | Purpose |
|-----------------|---------|
| `user_name` | User's display name |
| `user_email` | User's email |
| `app_name` | "MyFusion Helper" |
| `base_url` | Stage-specific app URL |
| `reset_code` | 6-digit password reset code |
| `plan_name` | Billing plan name |
| `helper_name` | Helper name (execution alerts) |
| `error_msg` | Error details |
| `invoice_url`, `amount`, `invoice_number` | Billing |
| `card_last4`, `card_brand`, `card_exp_month`, `card_exp_year` | Card expiring |
| `refund_reason` | Refund details |
| `inviter_name`, `inviter_email`, `role_name`, `account_name`, `invite_token` | Team invite |
| `resource_name`, `usage_percent`, `usage_current`, `usage_limit` | Usage alerts |
| `total_helpers`, `total_executed`, `total_succeeded`, `total_failed`, `success_rate`, `top_helper`, `week_start`, `week_end` | Weekly summary |
| `current_year` | Auto-injected |

### Notification Worker Dispatch

The notification worker (`cmd/handlers/notification-worker/main.go`) processes SQS jobs by type:

| Job Type | Service Method | Notes |
|----------|---------------|-------|
| `welcome` | `SendWelcomeEmail` | |
| `password_reset` | `SendPasswordResetEmail` | Passes `reset_code` from job data |
| `execution_failure` | `SendHelperExecutionAlert` | |
| `connection_issue` | `SendConnectionAlert` | |
| `usage_alert` | `SendUsageAlert` | |
| `billing_event` | `SendBillingEvent` | `event_type` sub-field → resolves to billing_event/{sub_type} |
| `weekly_summary` | `SendWeeklySummary` | |
| `team_invite` | `SendTeamInvite` | |

### Notification Service Methods (`internal/notifications/notifications.go`)

- `SendWelcomeEmail(ctx, name, email)`
- `SendPasswordResetEmail(ctx, email, resetCode)`
- `SendEmailVerificationEmail(ctx, email, verifyCode)`
- `SendHelperExecutionAlert(ctx, accountID, email, helperName, errorMsg)`
- `SendBillingEvent(ctx, accountID, email, eventType, planName, extraData...)`
- `SendConnectionAlert(ctx, accountID, email, connectionName)`
- `SendUsageAlert(ctx, accountID, email, resourceName, usagePercent, usageCurrent, usageLimit)`
- `SendWeeklySummary(ctx, accountID, email, summaryData)`
- `SendTeamInvite(ctx, inviterName, inviterEmail, inviteeEmail, roleName, accountName, inviteToken)`

### Stage-Specific Email Configuration

| Stage | FROM Address | CTA Base URL |
|-------|-------------|--------------|
| dev | `noreply@dev.myfusionhelper.ai` | `https://dev.myfusionhelper.ai` |
| staging | `noreply@staging.myfusionhelper.ai` | `https://staging.myfusionhelper.ai` |
| main | `noreply@myfusionhelper.ai` | `https://app.myfusionhelper.ai` |

Configured via `fromEmail` and `appUrl` custom vars in each service's `serverless.yml`. Go code fallbacks in `ses_client.go` (`getDefaultFromEmail()`) and `templates.go` (`getAppBaseURL()`) derive from the `STAGE` env var.

### SES Domain Verification

SES domain identity for `myfusionhelper.ai` is verified via Easy DKIM (3 CNAME records in Route53). Parent domain verification automatically covers all subdomains (`dev.myfusionhelper.ai`, `staging.myfusionhelper.ai`, etc.). DKIM, SPF, and DMARC records are managed in the SES CloudFormation stack (`IsMain` condition).

**SES Sandbox**: Check `ProductionAccessEnabled` status. In sandbox mode, SES can only send to verified email addresses or the test inbox domain.

### Test Email Inbox (dev only)

A complete email receiving system exists for testing transactional emails in dev. Test accounts use `@test.myfusionhelper.ai` addresses.

**Infrastructure** (all `IsDev` conditional in `services/infrastructure/ses/serverless.yml`):
- Route53 MX record: `test.myfusionhelper.ai` → `inbound-smtp.us-west-2.amazonaws.com`
- SES Receipt Rule Set (`mfh-test-email-rules`) catches `@test.myfusionhelper.ai`
- S3 bucket `mfh-test-emails` stores raw emails (30-day auto-expiry)
- Lambda organizer (`mfh-test-email-organizer`) copies emails to `by-recipient/<email>/<timestamp>_<subject>`
- SNS topic wires SES → Lambda

**How to use for testing:**

1. Sign up a test account with email `anything@test.myfusionhelper.ai`
2. The app sends emails (welcome, billing, etc.) to that address
3. SES receives the email and stores it in S3

**Reading test emails:**

```bash
# List all emails for a specific test address
AWS_PROFILE=mfh-claude aws s3 ls s3://mfh-test-emails/by-recipient/user1@test.myfusionhelper.ai/

# Read the most recent email (raw MIME format)
AWS_PROFILE=mfh-claude aws s3 cp s3://mfh-test-emails/by-recipient/user1@test.myfusionhelper.ai/<key> -

# List all test addresses that have received email
AWS_PROFILE=mfh-claude aws s3 ls s3://mfh-test-emails/by-recipient/

# List raw incoming emails (by message ID)
AWS_PROFILE=mfh-claude aws s3 ls s3://mfh-test-emails/incoming/
```

**Parsing email content from S3:**

The stored emails are in raw MIME format (RFC 822). To extract HTML body or headers:

```bash
# Download and view headers + text
AWS_PROFILE=mfh-claude aws s3 cp s3://mfh-test-emails/by-recipient/user1@test.myfusionhelper.ai/<key> /tmp/email.eml

# Extract just the subject and from
grep -E "^(Subject|From|To|Date):" /tmp/email.eml

# Extract HTML body (between boundaries) -- or use Python:
python3 -c "
import email, sys
with open('/tmp/email.eml', 'rb') as f:
    msg = email.message_from_binary_file(f)
print('Subject:', msg['Subject'])
print('From:', msg['From'])
print('To:', msg['To'])
for part in msg.walk():
    if part.get_content_type() == 'text/html':
        print(part.get_payload(decode=True).decode())
        break
"
```

**Verifying email content in automated checks:**

```bash
# Check that a welcome email was received
AWS_PROFILE=mfh-claude aws s3 ls s3://mfh-test-emails/by-recipient/testuser@test.myfusionhelper.ai/ | grep -i welcome

# Count emails received by a test address
AWS_PROFILE=mfh-claude aws s3 ls s3://mfh-test-emails/by-recipient/testuser@test.myfusionhelper.ai/ | wc -l
```

**Important notes:**
- Only exists in `dev` stage (all resources gated by `IsDev` CloudFormation condition)
- Emails auto-expire after 30 days (S3 lifecycle rule)
- The receipt rule set must be active (`aws ses set-active-receipt-rule-set --rule-set-name mfh-test-email-rules`) -- CI/CD handles this automatically
- SES sandbox mode may prevent sending to non-verified addresses; `@test.myfusionhelper.ai` is received directly by SES, bypassing sandbox restrictions for receiving

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
- Other workers: `mfh-stream-router`, `mfh-data-sync`, `mfh-notification-worker`, `mfh-trial-expiration`

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
3. **Stripe webhooks**: `setup-stripe-webhooks.sh` -- creates/updates Stripe webhook endpoint, verifies all 9 events registered
4. **Pre-gateway** (parallel): api-key-authorizer, scheduler, executions-stream, stream-router
5. **API Gateway** (creates HttpApi + Cognito authorizer + custom domain)
6. **API services** (parallel, max 3): auth, accounts, api-keys, helpers, platforms, data-explorer, billing, chat, emails, internal-email
7. **Helper workers** (parallel, max 10): auto-detected from changed `services/workers/*-worker/` directories
8. **Non-helper workers** (parallel): helper-worker (monolith), notification-worker, data-sync, trial-expiration
9. **Post-deploy**: `verify-deploy.sh` health check + Stripe webhook event verification

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
