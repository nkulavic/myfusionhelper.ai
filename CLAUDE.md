# MyFusion Helper - Monorepo

AI-powered CRM automation platform. Users connect CRM platforms (Keap, GoHighLevel, ActiveCampaign, Ontraport, HubSpot) and configure "Helpers" -- small automation units that manipulate contacts, tags, data, and integrations.

## Monorepo Structure

```
myfusionhelper.ai/
├── apps/
│   ├── web/                    # Next.js 15 dashboard (app.myfusionhelper.ai)
│   └── marketing/              # Marketing site [placeholder]
├── backend/
│   └── golang/                 # Go Lambda handlers + Serverless Framework
├── packages/
│   ├── types/                  # Shared TypeScript types (@myfusionhelper/types)
│   └── ui/                     # Shared UI primitives (@myfusionhelper/ui)
├── docs/                       # Architecture docs
└── .github/workflows/          # CI/CD (deploy-backend.yml, sync-internal-secrets.yml)
```

## Tech Stack

| Layer | Technology |
|-------|-----------|
| Frontend | Next.js 15, React 19, TypeScript 5.7 |
| Styling | Tailwind CSS 3 + shadcn/ui + CSS variables (HSL) |
| State | Zustand (client state) + React Query (server state) |
| Forms | react-hook-form + zod validation |
| Backend | Go 1.24, AWS Lambda (ARM64, provided.al2023) |
| Database | DynamoDB (PAY_PER_REQUEST, single-table-ish per entity) |
| Auth | AWS Cognito (USER_PASSWORD_AUTH), JWT tokens |
| API | API Gateway v2 (HTTP API) with Cognito JWT authorizer |
| Deploy | Serverless Framework v4, serverless-go-plugin |
| CI/CD | GitHub Actions with OIDC → `GitHubActions-Deploy-Dev` IAM role |
| Hosting | Vercel (frontend), AWS (backend) |
| Monorepo | npm workspaces + Turborepo |

## Key Infrastructure

- **AWS Account**: 570331155915, region us-west-2
- **API Gateway**:
  - Default endpoint: `https://a95gb181u4.execute-api.us-west-2.amazonaws.com`
  - Custom domain (dev): `https://api-dev.myfusionhelper.ai`
  - Custom domain (main): `https://api.myfusionhelper.ai`
- **Route53 Hosted Zone**: `myfusionhelper.ai` (ID: Z071462818IPQJBH38AMK)
- **ACM Certificate**: Managed via `services/infrastructure/acm` (DNS validation)
- **Cognito User Pool**: `us-west-2_1E74cZW97`
- **DynamoDB table prefix**: `mfh-{stage}-` (e.g., `mfh-dev-users`)
- **S3 data bucket**: `mfh-{stage}-data` (also stores email templates at `email-templates/`)
- **Unified Secrets SSM**: `/myfusionhelper/{stage}/secrets` (single SecureString JSON -- see Secrets section below)

## AWS CLI Profile for Claude Code

**CRITICAL**: Claude Code MUST always use the `mfh-claude` AWS profile. This is the ONLY profile Claude Code is authorized to use.

```bash
# ALWAYS prefix AWS commands with this:
AWS_PROFILE=mfh-claude aws <command>

# Or set it for the session:
export AWS_PROFILE=mfh-claude
```

**There is no default AWS profile configured.** Any `aws` command without `AWS_PROFILE=mfh-claude` will fail.

### IAM User & Permissions

- **IAM User**: `myfusion-helper-ai-claude-code`
- **Policy**: `mfh-claude-code-us-west-2-only`
- **Region**: Hard-locked to **us-west-2 only**

### What's Allowed

- **All AWS services in us-west-2**: Lambda, DynamoDB, SQS, S3, CloudFormation, CloudWatch, SSM, API Gateway, SES, IAM, etc.
- **Global services**: IAM, STS, Route53, CloudFront, ACM (these don't have regional endpoints)

### What's Denied

- **All regional services in ANY region other than us-west-2** -- enforced with an explicit `Deny` + `StringNotEquals` on `aws:RequestedRegion`, which cannot be overridden by any other policy
- Do NOT attempt to use any other AWS profile
- Do NOT attempt to access resources in other regions

## Development Setup

```bash
# Prerequisites: Node 20+, npm 10+
node --version   # must be >= 20 (see .nvmrc)

# Install all workspace dependencies
npm install

# Run the web app (dev server on port 3001)
npm run web

# Run all apps
npm run dev

# Lint / type-check
npm run lint
npm run type-check

# Format code
npm run format
```

### Environment Variables

Copy `apps/web/.env.example` to `apps/web/.env.local` and fill in:
- `NEXT_PUBLIC_COGNITO_USER_POOL_ID`
- `NEXT_PUBLIC_COGNITO_CLIENT_ID`
- `NEXT_PUBLIC_AWS_REGION`
- `NEXT_PUBLIC_API_URL`

## Code Conventions

### Formatting (Prettier)
- No semicolons
- Single quotes
- 2-space indent
- Trailing commas (es5)
- 100 char print width
- Tailwind class sorting plugin

### TypeScript
- Strict mode enabled
- Path aliases: `@/*` → `./src/*` (in web app)
- Shared packages: `@myfusionhelper/types`, `@myfusionhelper/ui`

### API Data Flow
- Backend returns **snake_case** JSON
- Frontend `apiClient` auto-converts to **camelCase** on response
- Frontend auto-converts to **snake_case** on request
- Types in `packages/types/src/index.ts` use camelCase (frontend convention)
- Go structs use `json:"snake_case"` tags

### Branching
- `main` -- production
- `staging` -- QA
- `dev` -- development (default working branch)

## ⚠️ CRITICAL: Deployment Policy

**ALL DEPLOYMENTS MUST GO THROUGH CI/CD PIPELINE**

- ✅ Push code to `dev` or `main` branch → GitHub Actions deploys automatically to **us-west-2**
- ❌ NEVER run `npx sls deploy` manually (except for emergency debugging)
- ❌ NEVER deploy to any region other than **us-west-2**

**Region Lock**: ALL infrastructure and services are deployed ONLY to **us-west-2**. This is enforced in CI/CD (`--region us-west-2` on every command) and must be set in all `serverless.yml` files (`region: us-west-2`).

## Backend Deploy Order (CI/CD Automated)

Infrastructure must deploy before API services. The CI pipeline (`deploy-backend.yml`) enforces this order:

1. **Build & test** Go code (Go 1.23, `CGO_ENABLED=1 go build ./...`)
2. **Infrastructure** (parallel): cognito, dynamodb-core, s3, sqs, ses, monitoring, acm
3. **Stripe webhooks**: `setup-stripe-webhooks.sh` -- creates/updates Stripe webhook endpoint with all 9 events, verifies events match
4. **Pre-gateway** (parallel): api-key-authorizer, scheduler, executions-stream, stream-router
5. **API Gateway** (creates HttpApi + Cognito authorizer + custom domain mapping)
6. **Route53** (creates DNS records for custom domain after gateway)
7. **API services** (parallel, max 3): auth, accounts, api-keys, helpers, platforms, data-explorer, billing, chat, emails, internal-email
8. **Helper workers** (parallel, max 10): 97 individual self-contained workers, auto-detected from changed `services/workers/*-worker/` directories
9. **Non-helper workers** (parallel): helper-worker (deprecated monolith), notification-worker, data-sync, trial-expiration
10. **Post-deploy**: `verify-deploy.sh` health check + Stripe webhook event verification

**All deployments target us-west-2 region exclusively.**

### Emergency Manual Deploy (Debugging Only)

Manual deployment should ONLY be used for emergency debugging. Always specify `--region us-west-2`:

```bash
cd backend/golang
npm install                                           # installs serverless-go-plugin + glob
cd services/api/auth
npx sls deploy --stage dev --region us-west-2        # MUST specify us-west-2
```

## CI/CD Pipeline Detail

Three GitHub Actions workflows in `.github/workflows/`:

### 1. `deploy-backend.yml` -- Main Deployment

**Triggers**: Push to `dev`/`main` (paths: `backend/golang/**`), or manual dispatch
**Auth**: GitHub OIDC token → assumes `GitHubActions-Deploy-Dev` IAM role (account 570331155915)
**Region**: `us-west-2` hardcoded in workflow env

**Job dependency graph**:
```
build-test
  ├── deploy-infra (parallel: cognito, dynamodb-core, s3, sqs, ses, acm)
  │     ├── setup-stripe-webhooks (creates/updates Stripe webhook endpoint, verifies events)
  │     ├── deploy-pre-gateway (parallel: api-key-authorizer, scheduler, executions-stream, stream-router)
  │     │     ├── deploy-gateway (API Gateway + Cognito authorizer + custom domain)
  │     │     │     ├── deploy-route53 (DNS records for custom domain)
  │     │     │     ├── deploy-monitoring (CloudWatch alarms)
  │     │     │     └── deploy-api (parallel max 3: auth, accounts, ..., billing, internal-email)
  │     │     └── deploy-helpers (parallel max 10: auto-detected changed *-worker/ directories)
  │     └── deploy-workers (parallel: helper-worker [monolith], notification-worker, data-sync, trial-expiration)
  └── detect-changed-helpers (git diff to find changed worker dirs)
        └── post-deploy (verify-deploy.sh + Stripe webhook verification)
```

**Helper auto-detection**: On push, uses `git diff` to detect changed `services/workers/*-worker/` directories. On manual dispatch, deploys ALL helpers. Excludes: helper-worker, notification-worker, data-sync, executions-stream, scheduler, voice assistant webhooks.

**Route53**: Deploys after gateway. Creates DNS A records for custom API domain (`api-dev.myfusionhelper.ai`, `api.myfusionhelper.ai`).

### 2. `sync-internal-secrets.yml` -- Secrets Sync to SSM

**Trigger**: Manual dispatch only (choose stage: dev, staging, main)
**Auth**: Same OIDC → `GitHubActions-Deploy-Dev` role

Reads stage-prefixed GitHub secrets, builds unified JSON via `scripts/build-internal-secrets.sh`, uploads to SSM as a single SecureString parameter. See **Secrets Architecture** section below for full detail.

### 3. `seed-platforms.yml` -- Platform Data Seeding

**Triggers**: Push to `dev`/`main` (paths: `platforms/ci_cd/seed/**`), or manual dispatch
**Purpose**: Seeds CRM platform configuration data into the Platforms DynamoDB table
**Platforms**: keap, gohighlevel, activecampaign, ontraport, hubspot, stripe, zoom, trello, google_sheets, etc.
**Detection**: Auto-detects changed `platform.json` files via git diff; manual can seed all or specific platform

### GitHub Secrets Required

| Secret | Used By | Purpose |
|--------|---------|---------|
| `SERVERLESS_ACCESS_KEY` | All 3 workflows | Serverless Framework registry auth |
| `DEV_INTERNAL_STRIPE_SECRET_KEY` | sync-internal-secrets | Stripe API secret key (dev) |
| `DEV_INTERNAL_STRIPE_PUBLISHABLE_KEY` | sync-internal-secrets | Stripe publishable key (dev) |
| `DEV_INTERNAL_STRIPE_WEBHOOK_SECRET` | sync-internal-secrets | Stripe webhook signing secret (dev) |
| `DEV_INTERNAL_STRIPE_PRICE_START` | sync-internal-secrets | Stripe price ID for Start plan (dev) |
| `DEV_INTERNAL_STRIPE_PRICE_GROW` | sync-internal-secrets | Stripe price ID for Grow plan (dev) |
| `DEV_INTERNAL_STRIPE_PRICE_DELIVER` | sync-internal-secrets | Stripe price ID for Deliver plan (dev) |
| `DEV_INTERNAL_GROQ_API_KEY` | sync-internal-secrets | Groq API key for voice assistants (dev, optional) |
| `DEV_INTERNAL_TWILIO_ACCOUNT_SID` | sync-internal-secrets | Twilio account SID (dev, **account suspended — needs reactivation**) |
| `DEV_INTERNAL_TWILIO_AUTH_TOKEN` | sync-internal-secrets | Twilio auth token (dev, **needs update after reactivation**) |
| `DEV_INTERNAL_TWILIO_FROM_NUMBER` | sync-internal-secrets | Twilio phone number (dev, **not yet set**) |
| `DEV_INTERNAL_TWILIO_MESSAGING_SID` | sync-internal-secrets | Twilio messaging service SID (dev, **not yet set**) |

Same pattern for `STAGING_INTERNAL_*` and `MAIN_INTERNAL_*` (36 stage-scoped + 1 global = **37 total GitHub secrets**).

## Secrets Architecture

### Unified Secrets (SSM Parameter Store)

All internal API secrets are consolidated into ONE SSM parameter per stage:

**Parameter**: `/myfusionhelper/{stage}/secrets` (SecureString, KMS-encrypted, Advanced tier)

**JSON structure**:
```json
{
  "stripe": {
    "secret_key": "sk_...",
    "publishable_key": "pk_...",
    "webhook_secret": "whsec_...",
    "price_start": "price_...",
    "price_grow": "price_...",
    "price_deliver": "price_..."
  },
  "groq": {
    "api_key": "gsk_..."
  },
  "twilio": {
    "account_sid": "AC...",
    "auth_token": "...",
    "from_number": "+1...",
    "messaging_sid": "MG..."
  }
}
```

### Secrets Flow

```
GitHub Secrets (master source, AES-256)
     ↓  (manual trigger: sync-internal-secrets.yml)
scripts/build-internal-secrets.sh
     ↓  (builds JSON, uploads via aws ssm put-parameter)
SSM Parameter Store: /myfusionhelper/{stage}/secrets (SecureString)
     ↓  (Lambda reads at runtime via INTERNAL_SECRETS_PARAM env var)
Go: config.LoadSecrets(ctx) → SecretsConfig struct (singleton, cached)
```

### How Lambda Functions Access Secrets

Every service that needs secrets adds to its `serverless.yml`:
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

Go code: `secrets, err := config.LoadSecrets(ctx)` then `secrets.Stripe.SecretKey`, etc.

### OAuth Credentials (Platform Connections)

OAuth credentials for CRM platform connections (e.g., Keap OAuth 2.0) are stored as **separate SSM parameters** (not in the unified JSON):

**Pattern**: `/{stage}/platforms/{platform_slug}/oauth/client_id` and `/{stage}/platforms/{platform_slug}/oauth/client_secret`

These are loaded by `loadOAuthCredentials()` in the platforms connection handler. Currently only Keap uses OAuth 2.0; other platforms use API keys entered directly by the user.

**OAuth flow** (Keap example):
1. Frontend calls `POST /platforms/{id}/oauth/start` (JWT-protected)
2. Backend generates state token, stores in `mfh-{stage}-oauth-states` DynamoDB table (TTL: 15 min)
3. Backend returns authorization URL → user redirected to Keap consent screen
4. Keap redirects to `GET /platforms/oauth/callback?code=...&state=...` (public endpoint)
5. Backend validates state, exchanges code for tokens, stores connection + auth records in DynamoDB
6. User redirected to success URL with `connection_id`

### Legacy SSM Parameters (Cleanup Needed)

These old SSM parameters still exist but are **no longer used**:
- `/dev/stripe/price_*` -- superseded by unified `/myfusionhelper/dev/secrets`
- `/mfh/dev/sqs/*/queue-url` -- superseded by convention-based queue URL construction in stream router

## Supported CRM Platforms

| Platform | Auth Type | Status |
|----------|-----------|--------|
| Keap (Infusionsoft) | OAuth 2.0 | Active |
| GoHighLevel | API Key | Active |
| ActiveCampaign | API Key | Active |
| Ontraport | API Key | Active |
| HubSpot | API Key (Private App) | Active |
| Stripe | API Key (Secret Key) | Active |

## Billing & Trial Model

**Plans**: `trial` (default for new accounts), `start`, `grow`, `deliver`. Legacy `free` plan exists but new registrations use `trial`.

**14-Day Free Trial**:
- No credit card required at registration
- Stripe customer created at registration (no subscription, for tracking/analytics)
- Start-level limits during trial: 10 helpers, 2 connections, 5k executions/mo
- `Account` fields: `TrialStartedAt`, `TrialEndsAt`, `TrialExpired`
- Trial expiration worker (`mfh-trial-expiration`) runs every 6 hours via EventBridge, scans for expired trials, sets `trial_expired = true`, and sends `trial_expired` email

**Soft Lock on Expiry**: Expired trial users can log in, view dashboard, access `/plans` and `/settings`, but can't create/run helpers. Other routes redirect to `/plans`.

**Enforcement**: `internal/billing/enforce.go` checks `account.Settings.Max*` vs `account.Usage.*` (pure DynamoDB reads). Trial expiration is an early return in all limit functions. `IsTrialPlan()` helper treats both `"trial"` and `"free"` as trial plans.

**Checkout**: When an active trial user subscribes, Stripe checkout uses remaining trial days (not a fresh 14 days) to prevent double-trial abuse. Expired trials get no trial period.

### Stripe Webhook Architecture

**Endpoint**: `POST /billing/webhook` (public, verified by Stripe signature)
**Handler**: `cmd/handlers/billing/clients/webhook/main.go`
**Setup script**: `scripts/setup-stripe-webhooks.sh` (idempotent, creates/updates endpoint + verifies events)

**Idempotency**: Every webhook event is recorded in `mfh-{stage}-webhook-events` DynamoDB table with `attribute_not_exists(event_id)` conditional write. Duplicate events return 200 immediately. Events tracked with status (`pending` → `processed`/`failed`) and 90-day TTL.

**Race condition fix**: `checkout.session.completed` and `customer.subscription.created` fire simultaneously. Checkout sets `subscription_email_sent=true` flag on account; subscription.created checks and clears it to prevent duplicate emails.

**9 Stripe events handled**:

| Stripe Event | Handler | Email Sent |
|---|---|---|
| `checkout.session.completed` | Activate plan, store metered item, sync Cognito | `subscription_created` |
| `customer.subscription.created` | Update plan/limits (conditional email) | `subscription_created` (if checkout didn't send) |
| `customer.subscription.updated` | Detect upgrade/downgrade via `planRank` map | `plan_upgraded` or `plan_downgraded` |
| `customer.subscription.deleted` | Downgrade to trial, mark expired, sync Cognito | `subscription_cancelled` |
| `customer.subscription.trial_will_end` | No data change | `trial_ending` |
| `invoice.paid` | Reset `past_due` → `active`; send receipt for all payments | `payment_recovered` or `payment_receipt` |
| `invoice.payment_failed` | No data change (Stripe retries) | `payment_failed` (with hosted invoice URL) |
| `charge.refunded` | No data change | `refund_processed` (amount, reason) |

### Transactional Email System (S3 Liquid Templates)

**Template engine**: Liquid templates (`github.com/osteele/liquid`) stored in S3, rendered at runtime. Templates are cached for 5 minutes by the `TemplateLoader`.

**S3 location**: `s3://mfh-{stage}-data/email-templates/{template_type}/` — each template has `subject.liquid`, `body.html.liquid`, `body.txt.liquid` (optional), and `meta.json` (header, CTA, icon).

**Template source files**: `backend/golang/email-templates/` — synced to S3 during deployment.

**Internal email service**: `cmd/handlers/internal-email/clients/send/main.go` — receives template type + data, resolves S3 path, renders Liquid templates, sends via SES.

**Notification service**: `internal/notifications/notifications.go` — HTTP client that calls the internal email API. Used by webhook handler, trial-expiration worker, and other services.

**Email flow**:
```
Handler → SQS (NotificationQueue)
  → Notification Worker → HTTP POST to /internal/emails/send
  → ResolveTemplatePath() → TemplateLoader (S3 + cache)
  → Liquid render with TemplateData bindings (snake_case)
  → meta.json (header_title, cta_url, etc.) injected into HTML wrapper
  → SES sends email
```

**9 email template types** (each a directory in S3):
- `welcome`, `password_reset`, `execution_alert`, `connection_alert`, `usage_alert`, `weekly_summary`, `team_invite`
- `billing_event/` — contains 12 sub-type subdirectories: `subscription_created`, `subscription_cancelled`, `payment_failed`, `payment_recovered`, `payment_receipt`, `trial_ending`, `trial_expired`, `plan_upgraded`, `plan_downgraded`, `card_expiring`, `refund_processed`, `default`

**Liquid template variables**: All use snake_case (e.g., `{{ user_name }}`, `{{ reset_code }}`, `{{ base_url }}`, `{{ plan_name }}`). Converted from Go `TemplateData` struct via `TemplateDataToBindings()`.

### Password Reset Flow (Self-Managed Verification Codes)

Password reset does **NOT** use Cognito's `ForgotPassword`/`ConfirmForgotPassword` APIs. Instead, the system generates and manages its own 6-digit verification codes:

1. `POST /auth/forgot-password` — generates crypto/rand 6-digit code, stores in `EMAIL_VERIFICATIONS_TABLE` (15-min TTL), enqueues branded email via SQS with the code. Always returns 200 (email enumeration prevention).
2. User receives branded email with `{{ reset_code }}` rendered in the template.
3. `POST /auth/reset-password` — verifies code from DynamoDB (`GetPendingByEmail` GSI query + in-memory match), calls Cognito `AdminSetUserPassword`, marks code as verified.

**Key files**: `cmd/handlers/auth/clients/forgot-password/main.go`, `cmd/handlers/auth/clients/reset-password/main.go`
**DynamoDB table**: `mfh-{stage}-email-verifications` (partition key: `verification_id`, GSI: `EmailIndex` on `email`)

## Helper Categories

Helpers are organized into categories: contact, data, tagging, automation, integration, notification, analytics. Each helper implements the `helpers.Helper` Go interface and is registered in a global registry via `helpers.Register()`.

## Data Explorer

Browse, query, filter, and export CRM data synced from connected platforms. Data is stored as Parquet files in S3 and queried in-process via DuckDB.

### Data Pipeline

```
CRM Platform (Keap, GHL, etc.)
  → data-sync worker pulls records via connector
  → Writes Parquet files to s3://mfh-{stage}-data/{connection_id}/{object_type}/
  → Data Explorer queries Parquet via DuckDB (in-Lambda, CGO required)
```

### API Endpoints

| Method | Route | Handler | Purpose |
|--------|-------|---------|---------|
| GET | `/data/catalog` | catalog | List all object types across user's connections (record counts, columns, sync status) |
| GET | `/data/schema` | schema | Column metadata for a specific object type |
| POST | `/data/query` | query | Query records with filters, sorting, pagination (DuckDB SQL on Parquet) |
| GET | `/data/record/{connId}/{objectType}/{recordId}` | record | Single record detail (full JSON) |
| POST | `/data/export` | export | Export matching records as CSV or JSON |
| POST | `/data/sync` | sync | Trigger manual data sync (enqueues SQS job) |
| POST | `/data/aggregate` | aggregate | Group/count aggregations for Studio charts |
| POST | `/data/timeseries` | timeseries | Time-bucketed aggregations for line/area charts |

### Key Files

| Layer | Path | Notes |
|-------|------|-------|
| Frontend page | `apps/web/src/app/(dashboard)/data-explorer/page.tsx` | Two-panel resizable layout |
| Zustand store | `apps/web/src/lib/stores/data-explorer-store.ts` | Selection, filters, pagination (persisted) |
| API client | `apps/web/src/lib/api/data-explorer.ts` | HTTP endpoints + request/response types |
| React Query hooks | `apps/web/src/lib/hooks/use-data-explorer.ts` | `useDataCatalog`, `useDataQuery`, `useDataRecord`, `useDataSchema`, `useTriggerSync` |
| Components | `apps/web/src/components/data-explorer/` | HierarchicalNav, DataTable, NLQueryBar, RecordDetail, JsonViewer, ExportUtils, etc. |
| Backend router | `backend/golang/cmd/handlers/data-explorer/main.go` | Consolidated Lambda handler dispatching by path+method |
| Client handlers | `backend/golang/cmd/handlers/data-explorer/clients/*/main.go` | 9 sub-handlers (catalog, schema, query, record, export, sync, aggregate, timeseries, health) |
| Serverless config | `backend/golang/services/api/data-explorer/serverless.yml` | **x86_64** arch (required for DuckDB/CGO), 512-1024MB memory |
| Smoke tests | `backend/golang/tests/smoke/test-data-explorer.sh` | Catalog, query, filter, sort, pagination, export, aggregate tests |

### Architecture Notes

- **x86_64 only**: Data explorer Lambda runs on x86_64 (not ARM64) because DuckDB requires CGO. A pre-built binary is copied during build instead of compiling from source.
- **Hierarchical navigation**: Frontend tree is Platform → Connection → ObjectType. Selection level drives which view renders (PlatformOverview, ConnectionOverview, DataTable, RecordDetail).
- **Filter operators**: `eq`, `neq`, `gt`, `gte`, `lt`, `lte`, `contains`, `startswith`, `in`, `between`, `daterange`.
- **NL Query bar**: UI exists with example suggestions but no LLM backend yet (placeholder).
- **DynamoDB table**: `mfh-{stage}-dashboards` (also used by Studio for its dashboards).

## Studio (Dashboard Builder)

Visual dashboard builder for creating custom CRM analytics dashboards with charts, scorecards, and tables. Data sourced from the Data Explorer's aggregate/timeseries endpoints.

### How It Works

1. Users create dashboards (blank or from platform-specific templates)
2. Add widgets: scorecard, bar, line, area, pie, funnel, or data table
3. Each widget configured with: data source, metric, dimension/grouping, date range, size
4. Widgets rendered on a 12-column responsive grid (Recharts for charts)
5. Dashboards persisted in DynamoDB; widgets stored as a list on the dashboard record

### API Endpoints

| Method | Route | Handler | Purpose |
|--------|-------|---------|---------|
| GET | `/studio/dashboards` | dashboards | List user's dashboards |
| POST | `/studio/dashboards` | dashboards | Create new dashboard |
| GET | `/studio/dashboards/{id}` | dashboards | Get single dashboard |
| PUT | `/studio/dashboards/{id}` | dashboards | Update dashboard (name, description, widgets) |
| DELETE | `/studio/dashboards/{id}` | dashboards | Soft-delete dashboard |
| GET | `/studio/templates` | templates | List templates filtered by user's connected platforms |
| POST | `/studio/templates/{id}/apply` | templates | Create dashboard from template (injects connection_id) |

Data endpoints (`/data/aggregate`, `/data/timeseries`) are served by the Data Explorer service.

### Templates

6 built-in templates in `cmd/handlers/studio/clients/templates/registry.go`:
- **Generic CRM Overview** -- works with any platform
- **Keap**, **GoHighLevel**, **ActiveCampaign**, **Ontraport**, **HubSpot** -- platform-specific dashboards
- Templates filtered dynamically based on user's connected platforms

### Widget Types & Config

| Chart Type | Data Sources | Metrics | Notes |
|------------|-------------|---------|-------|
| Scorecard | contacts, deals, tags | count, sum, avg | Single KPI value |
| Bar/Pie | contacts, deals, tags | count, sum, avg | Group by dimension (status, source, company, etc.) |
| Line/Area | contacts, deals, tags | count, sum, avg | Time series only (date_histogram) |
| Funnel | contacts, deals, tags | count | Ordered stages |
| Data Table | contacts, deals, tags | count, sum, avg | Tabular aggregation |

Widget sizes: `sm` (3-col), `md` (6-col), `lg` (9-col), `full` (12-col).

### Key Files

| Layer | Path | Notes |
|-------|------|-------|
| Frontend list page | `apps/web/src/app/(dashboard)/studio/page.tsx` | Dashboard grid + create dialog |
| Frontend detail page | `apps/web/src/app/(dashboard)/studio/[id]/page.tsx` | Canvas editor |
| Zustand store | `apps/web/src/lib/stores/studio-store.ts` | UI state (activeDateRange), type definitions |
| API client | `apps/web/src/lib/api/studio.ts` | CRUD + template + aggregate/timeseries APIs |
| React Query hooks | `apps/web/src/lib/hooks/use-studio.ts` | `useDashboards`, `useDashboard`, `useCreateDashboard`, `useUpdateDashboard`, `useDeleteDashboard`, `useTemplates`, `useApplyTemplate` |
| Components | `apps/web/src/components/studio/` | DashboardList, DashboardCanvas, AddWidgetModal, WidgetCard, WidgetRenderer, DateRangeFilter |
| Chart components | `apps/web/src/components/studio/charts/` | Scorecard, BarChart, LineChart, AreaChart, PieChart, FunnelChart, DataTable (all Recharts) |
| Mock data | `apps/web/src/lib/mock-data/studio-mock-data.ts` | Frontend mock aggregation (currently used for chart rendering) |
| Backend router | `backend/golang/cmd/handlers/studio/main.go` | Lambda handler routing `/studio/*` |
| Dashboard handler | `backend/golang/cmd/handlers/studio/clients/dashboards/main.go` | CRUD with ownership checks, soft delete, optimistic locking |
| Template handler | `backend/golang/cmd/handlers/studio/clients/templates/main.go` | List (filtered by connections) + apply |
| Template registry | `backend/golang/cmd/handlers/studio/clients/templates/registry.go` | 6 static template definitions |

### DynamoDB

- **Table**: `mfh-{stage}-dashboards` (PK: `dashboard_id`, GSI: `AccountIdIndex` on `account_id`)
- **Dashboard ID format**: `dash:` + UUIDv7
- **Widget ID format**: `wdg-` + UUID
- Soft deletes (status: `active` → `deleted`)

### Current State

- Dashboard CRUD and template system are fully functional (backend + frontend)
- Chart rendering currently uses **frontend mock data** (`studio-mock-data.ts`) -- not yet wired to backend aggregate/timeseries endpoints
- Backend `/data/aggregate` and `/data/timeseries` endpoints exist and are ready for integration

## Project Management & Planning Workflow

**Teamwork Project ID**: 674054 (myfusionhelper.ai)

### Plan Files Live in Teamwork Notebooks

**IMPORTANT**: All implementation plans, roadmaps, and project status documents MUST be stored as Teamwork notebooks in project 674054 -- NOT as local markdown files in the repo. This ensures plans persist across conversation sessions and are accessible from Teamwork's UI.

**Workflow**:
1. Before starting significant work, check Teamwork notebooks for existing plans: `teamwork_notebooks list projectId=674054`
2. When creating a new plan, save it as a Teamwork notebook: `teamwork_notebooks create projectId=674054`
3. When updating progress, update the relevant notebook: `teamwork_notebooks update notebookId=...`
4. Use version comments when updating notebooks to track changes over time
5. Do NOT create `*_PLAN.md`, `*_ROADMAP.md`, or similar planning files in the repo

**Notebook naming convention** (with IDs):
- `Implementation Plan` (417843) -- current implementation priorities and status
- `Product Plan` (417844) -- product vision, features, pricing, target users
- `Test Plan` (417845) -- QA checklists and test procedures
- `Architecture Notes` (417842) -- technical architecture decisions
- `Helper Architecture & Migration Plan` (417846) -- helper execution system, service connections, legacy-to-Go migration plan

**Task Lists** (8 active):
- P0 - Git & Code Hygiene (3256057)
- P0 - Infrastructure Deployment (3256058)
- P0 - End-to-End Smoke Testing (3256059)
- P1 - Backend Gaps (3256060)
- P1 - Go Test Suite (3256061)
- P1 - Helper Migration - Legacy to Go (3256064)
- P2 - Frontend Polish (3256062)
- P2 - Launch Prep (3256063)

### Teamwork Task Tracking

Use Teamwork tasks (project 674054) for tracking work items, bugs, and feature requests. Claude Code has access to Teamwork MCP tools for creating and managing tasks.

### What Belongs in the Repo vs Teamwork

| Content | Location |
|---------|----------|
| Code documentation (CLAUDE.md, README) | Repo |
| API reference docs | Repo (`docs/API.md`) |
| Implementation plans & roadmaps | Teamwork notebooks |
| Task tracking & status | Teamwork tasks |
| Product strategy & vision | Teamwork notebooks |
| QA test plans | Teamwork notebooks |
| Landing page copy | Repo (`docs/LANDING_COPY.md`) |
