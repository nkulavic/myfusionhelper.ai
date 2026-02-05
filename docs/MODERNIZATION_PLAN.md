# MyFusionHelper.ai Modernization Plan

## Overview

Rebuild the MyFusion Helper platform from a WordPress-based CRM automation system to a modern, AI-first, multi-platform CRM automation platform.

**Current State**: `secure.myfusionsolutions.com` - WordPress + PHP with 90+ helper microservices, Keap-only integration
**Target State**: `myfusionhelper.ai` - Go backend + Next.js frontend, multi-CRM support, AI-powered automation

---

## Decisions Log

Key architectural decisions made during development:

| Decision | Choice | Rationale |
|----------|--------|-----------|
| **Authentication** | AWS Cognito | Free tier, integrates with AWS ecosystem, no monthly fee |
| **Database** | DynamoDB from day 1 | Build on target architecture, avoid migration later |
| **Infrastructure-as-Code** | Serverless Framework | Team familiarity from listbackup.ai, not ready for Terraform |
| **AWS Account** | Default account (570331155915) | Same account as secure.myfusionsolutions.com (this is an upgrade) |
| **AWS Region** | us-west-2 | All resources in Oregon |
| **AWS CLI Profile** | Default (no --profile flag) | Using default AWS credentials |
| **Repository** | github.com/nkulavic/myfusionhelper.ai | Public repo on personal account for free Vercel integration |
| **Frontend Hosting** | Vercel | Free tier, excellent Next.js support |
| **Frontend Auth Library** | AWS Amplify Auth (Cognito client) | Replaces Better Auth; connects to Cognito user pool |
| **Frontend State** | Zustand | Lightweight, simple API |
| **AI Provider** | Groq | Cost-effective, fast inference |
| **Monorepo Tool** | Turborepo | Fast builds, good Next.js integration |
| **UI Library** | shadcn/ui + Tailwind CSS | Beautiful defaults, fully customizable |
| **Resource Naming** | `mfh-{stage}-*` | e.g., mfh-dev-users, mfh-prod-helpers |

---

## Development Approach

**Build infrastructure and UI in parallel:**
- Frontend prototype is deployed on Vercel with mock data
- AWS infrastructure (Cognito, DynamoDB) is defined via serverless.yml
- Connect frontend to real backend as services come online
- DynamoDB from day 1 — no intermediate database

**CRM Platform Priority:**
1. Keap (existing users)
2. GoHighLevel (high demand)
3. ActiveCampaign
4. Ontraport
5. Any platform with HTTP POST webhook functionality

**AI Provider:** Groq (cost-effective, fast inference)

---

## MVP vs Full Feature Scope

**MVP (Phase 1-3):**
- UI prototype with mock data → connected to real backend
- Cognito auth (email/password + Google OAuth)
- Single user/account (no team features)
- API key generation (basic)
- 1-2 CRM connections (Keap, GoHighLevel)
- Core helpers (10-15 most used)
- Basic execution logging

**Full Product (Phase 4+):**
- Multi-user accounts with roles (Owner, Admin, Member, Viewer)
- Team invitations
- Full Stripe billing integration
- Usage-based billing & metering
- All 90+ helpers
- AI features (email, reports, chat)
- All CRM platforms
- Webhooks & advanced integrations

**Architecture Principle:** Design data models and APIs to support the full feature set from day 1, even if not implemented in MVP. This avoids painful migrations later.

---

## Architecture

### Domain Structure
- `myfusionhelper.ai` - Marketing/sales site (public)
- `app.myfusionhelper.ai` - Main application (authenticated)
- `api.myfusionhelper.ai` - Go backend API (API Gateway custom domain)

### Monorepo Structure
```
myfusionhelper.ai/
├── apps/
│   ├── marketing/              # Next.js - myfusionhelper.ai (future)
│   └── web/                    # Next.js - app.myfusionhelper.ai
│       ├── src/
│       │   ├── app/
│       │   │   ├── (auth)/     # Login, register, forgot password
│       │   │   ├── (dashboard)/# Protected routes
│       │   │   └── api/        # API route handlers
│       │   ├── components/
│       │   │   └── ui/         # shadcn/ui components
│       │   └── lib/
│       │       ├── api/        # API client modules
│       │       └── stores/     # Zustand state stores
│       └── package.json
├── backend/
│   └── golang/
│       ├── cmd/
│       │   └── handlers/       # Lambda handlers (1 per service)
│       │       ├── auth/
│       │       ├── accounts/
│       │       ├── helpers/
│       │       ├── connections/
│       │       └── executions/
│       ├── internal/
│       │   ├── connectors/     # Unified CRM connector interface
│       │   ├── database/       # DynamoDB repositories
│       │   ├── helpers/        # Helper implementations (90+)
│       │   ├── middleware/     # Auth, logging, rate limiting
│       │   └── services/       # Business logic
│       └── services/
│           ├── api/            # API service serverless.yml files
│           │   ├── auth/
│           │   ├── users/
│           │   ├── accounts/
│           │   ├── helpers/
│           │   ├── connections/
│           │   └── gateway/
│           └── infrastructure/ # Infrastructure serverless.yml files
│               ├── cognito/
│               ├── dynamodb/
│               │   └── core/
│               ├── s3/
│               └── sqs/
├── packages/
│   ├── ui/                     # Shared UI components
│   ├── types/                  # Shared TypeScript types
│   └── config/                 # Shared configs (eslint, tsconfig)
├── docs/
│   └── MODERNIZATION_PLAN.md
├── turbo.json
└── package.json
```

### Serverless Service Structure (following listbackup.ai pattern)

Each service is an independent Serverless Framework stack with its own `serverless.yml`. Services reference each other via CloudFormation cross-stack exports (`${cf:...}`).

**Infrastructure services** (deploy first, rarely change):
```
backend/golang/services/infrastructure/
├── cognito/serverless.yml          # User pool, client, groups
├── dynamodb/core/serverless.yml    # All DynamoDB tables
├── s3/serverless.yml               # S3 buckets (future)
└── sqs/serverless.yml              # SQS queues (future)
```

**API services** (deploy frequently, contain business logic):
```
backend/golang/services/api/
├── auth/serverless.yml             # Login, register, refresh, verify
├── users/serverless.yml            # User CRUD
├── accounts/serverless.yml         # Account management
├── helpers/serverless.yml          # Helper CRUD + execution
├── connections/serverless.yml      # CRM connection management
└── gateway/serverless.yml          # API Gateway custom domain
```

**Deployment order:**
1. `infrastructure/cognito` → exports UserPoolId, UserPoolArn, ClientId
2. `infrastructure/dynamodb/core` → exports all table names and ARNs
3. `infrastructure/s3` → exports bucket names and ARNs
4. `infrastructure/sqs` → exports queue URLs and ARNs
5. `api/gateway` → creates API Gateway, custom domain
6. `api/auth` → references Cognito and DynamoDB exports
7. `api/users`, `api/accounts`, etc. → reference infrastructure exports

---

## Phase 1: UI/UX Prototype on Vercel ✅ COMPLETED

**Status**: Deployed to Vercel at app.myfusionhelper.ai

### 1.1 Repository Setup ✅
- [x] Initialize monorepo with Turborepo
- [x] Set up `apps/web` with Next.js 15 + React 19
- [x] Deploy to Vercel
- [x] Configure shadcn/ui component library
- [x] Set up shared packages (ui, types)
- [x] Push to GitHub (nkulavic/myfusionhelper.ai)
- [x] Create branch structure (main, dev, staging)

### 1.2 UI Prototype Pages ✅
Built core UI with mock data:

**Core Automation:**
- [x] **Helpers Library** (`/helpers`) - Browse helpers by category with search
- [x] **Connections** (`/connections`) - CRM connection management
- [x] **Executions** (`/executions`) - Execution history with stats

**Business Intelligence:**
- [x] **Insights** (`/insights`) - AI-powered analytics and reports

**Settings:**
- [x] **Settings** (`/settings`) - Profile, Account, Team, API Keys, Billing, Notifications tabs

### 1.3 Design System ✅
- [x] shadcn/ui Button, Card, Input components
- [x] Tailwind CSS with CSS variables for theming
- [x] Geist font family (Sans + Mono)
- [x] Dark mode CSS variables ready

### 1.4 Temporary Auth (to be replaced) ⚠️
- Better Auth was set up as placeholder for prototype
- **Will be replaced with Cognito** in Phase 2
- Files to update: `src/lib/auth.ts`, `src/lib/auth-client.ts`, `src/middleware.ts`
- Remove packages: `better-auth`, `@neondatabase/serverless`, `drizzle-orm`, `drizzle-kit`

### Still TODO from Phase 1:
- [ ] **Dashboard** (`/`) - Currently just landing page, needs dashboard overview
- [ ] **Helper Builder** (`/helpers/new`) - Visual helper configuration
- [ ] **Helper Detail** (`/helpers/[id]`) - Edit existing helper
- [ ] Set up `apps/marketing` - Marketing site (lower priority)
- [ ] Zustand stores (auth-store, workspace-store, helper-store, ui-store)

---

## Phase 2: AWS Infrastructure & Auth

**Goal**: Deploy all AWS infrastructure via Serverless Framework, replace Better Auth with Cognito, connect frontend to real auth.

### 2.1 Infrastructure - Cognito ✅ (serverless.yml created)
**File**: `backend/golang/services/infrastructure/cognito/serverless.yml`

Already created with:
- User pool (`mfh-{stage}-user-pool`) with email auth
- Password policy (8+ chars, upper, lower, numbers)
- Optional MFA (software token)
- Web client with OAuth flows
- User groups: owner, admin, member, viewer
- CloudFormation exports: UserPoolId, UserPoolArn, ClientId, JwksUri, Issuer

**Deploy**: `cd backend/golang/services/infrastructure/cognito && sls deploy --stage dev`

### 2.2 Infrastructure - DynamoDB ✅ (serverless.yml created)
**File**: `backend/golang/services/infrastructure/dynamodb/core/serverless.yml`

Already created with 8 tables:

| Table | Partition Key | Sort Key | GSIs |
|-------|--------------|----------|------|
| `mfh-{stage}-users` | user_id | - | EmailIndex, CognitoUserIdIndex |
| `mfh-{stage}-accounts` | account_id | - | OwnerUserIdIndex |
| `mfh-{stage}-user-accounts` | user_id | account_id | AccountIdIndex |
| `mfh-{stage}-api-keys` | key_id | - | AccountIdIndex, KeyHashIndex |
| `mfh-{stage}-connections` | connection_id | - | AccountIdIndex |
| `mfh-{stage}-helpers` | helper_id | - | AccountIdIndex |
| `mfh-{stage}-executions` | execution_id | - | AccountIdCreatedAtIndex, HelperIdCreatedAtIndex |
| `mfh-{stage}-oauth-states` | state | - | - (TTL enabled) |

All tables have: PAY_PER_REQUEST billing, deletion protection, point-in-time recovery, DeletionPolicy: Retain

**Deploy**: `cd backend/golang/services/infrastructure/dynamodb/core && sls deploy --stage dev`

### 2.3 Infrastructure - SQS (TODO)
**File to create**: `backend/golang/services/infrastructure/sqs/serverless.yml`
- [ ] `mfh-{stage}-helper-execution.fifo` - Main execution queue
- [ ] `mfh-{stage}-helper-execution-dlq.fifo` - Dead letter queue
- [ ] `mfh-{stage}-webhooks` - Inbound webhook processing

### 2.4 Infrastructure - S3 (TODO)
**File to create**: `backend/golang/services/infrastructure/s3/serverless.yml`
- [ ] `mfh-{stage}-uploads` - User file uploads
- [ ] `mfh-{stage}-exports` - Report exports

### 2.5 Frontend Auth Migration (TODO)
Replace Better Auth with Cognito:
- [ ] Install `@aws-amplify/auth` or build custom Cognito client
- [ ] Update `src/lib/auth.ts` → Cognito configuration
- [ ] Update `src/lib/auth-client.ts` → Cognito sign-in/sign-up/sign-out
- [ ] Update `src/middleware.ts` → JWT validation against Cognito
- [ ] Update login/register pages to use Cognito
- [ ] Remove Better Auth, Neon, Drizzle packages
- [ ] Add environment variables: NEXT_PUBLIC_COGNITO_USER_POOL_ID, NEXT_PUBLIC_COGNITO_CLIENT_ID, NEXT_PUBLIC_AWS_REGION

### 2.6 API Gateway (TODO)
**File to create**: `backend/golang/services/api/gateway/serverless.yml`
- [ ] HTTP API (API Gateway v2)
- [ ] Custom domain: api.myfusionhelper.ai
- [ ] CORS configuration for app.myfusionhelper.ai
- [ ] Stage-based deployment (dev, staging, prod)

---

## Phase 3: Backend Services (Go)

**Goal**: Build Go Lambda services for auth, users, accounts, helpers, connections. Follow consolidated handler pattern from listbackup.ai.

### 3.1 Authentication Service
**Files to create:**
- `backend/golang/cmd/handlers/auth/main.go` - Consolidated Lambda handler
- `backend/golang/cmd/handlers/auth/clients/login/main.go`
- `backend/golang/cmd/handlers/auth/clients/register/main.go`
- `backend/golang/cmd/handlers/auth/clients/refresh/main.go`
- `backend/golang/cmd/handlers/auth/clients/verify/main.go`
- `backend/golang/services/api/auth/serverless.yml`

**Consolidated Handler Pattern** (from listbackup.ai):
```go
func HandlePublic(ctx context.Context, event events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
    path := event.RequestContext.HTTP.Path
    method := event.RequestContext.HTTP.Method

    switch {
    case method == "POST" && strings.HasSuffix(path, "/login"):
        return login.Handle(ctx, event)
    case method == "POST" && strings.HasSuffix(path, "/register"):
        return register.Handle(ctx, event)
    case method == "POST" && strings.HasSuffix(path, "/refresh"):
        return refresh.Handle(ctx, event)
    }
}
```

### 3.2 Auth Middleware
**Files to create:**
- `backend/golang/internal/middleware/auth/auth.go` - JWT validation (Cognito)
- `backend/golang/internal/middleware/apikey/apikey.go` - API key validation

**Two Authentication Methods:**

1. **JWT (App Authentication)** - For web app users
   - Cognito User Pool issues JWTs
   - Validate against Cognito JWKS endpoint
   - AuthContext includes UserID, AccountID, Permissions

2. **API Keys (Helper Execution)** - For CRM webhooks
   - Format: `mfh_live_xxxxxxxxxxxx` / `mfh_test_xxxxxxxxxxxx`
   - Header: `X-API-Key: mfh_live_xxxxx`
   - Tied to Account, not User
   - Scoped permissions

### 3.3 Core Data Models
```go
type User struct {
    UserID           string    `json:"user_id"`           // "user:uuid"
    CognitoUserID    string    `json:"cognito_user_id"`
    Email            string    `json:"email"`
    Name             string    `json:"name"`
    CurrentAccountID string    `json:"current_account_id"`
    CreatedAt        time.Time `json:"created_at"`
}

type Account struct {
    AccountID        string    `json:"account_id"`        // "account:uuid"
    OwnerUserID      string    `json:"owner_user_id"`
    Name             string    `json:"name"`
    Company          string    `json:"company"`
    Plan             string    `json:"plan"`              // "free", "pro", "business"
    Status           string    `json:"status"`            // "active", "suspended"
    StripeCustomerID string    `json:"stripe_customer_id"`
    Settings         AccountSettings `json:"settings"`
    Usage            AccountUsage    `json:"usage"`
}

type UserAccount struct {
    UserID      string          `json:"user_id"`
    AccountID   string          `json:"account_id"`
    Role        string          `json:"role"`        // "owner", "admin", "member", "viewer"
    Permissions UserPermissions `json:"permissions"`
    LinkedAt    time.Time       `json:"linked_at"`
}

type APIKey struct {
    KeyID       string     `json:"key_id"`
    AccountID   string     `json:"account_id"`
    CreatedBy   string     `json:"created_by"`
    Name        string     `json:"name"`
    KeyHash     string     `json:"key_hash"`
    KeyPrefix   string     `json:"key_prefix"`
    Permissions []string   `json:"permissions"`
    Status      string     `json:"status"`
    LastUsedAt  *time.Time `json:"last_used_at"`
    CreatedAt   time.Time  `json:"created_at"`
    ExpiresAt   *time.Time `json:"expires_at"`
}
```

### 3.4 DynamoDB Repositories
**Files to create:**
- `backend/golang/internal/database/users_repository.go`
- `backend/golang/internal/database/accounts_repository.go`
- `backend/golang/internal/database/user_accounts_repository.go`
- `backend/golang/internal/database/apikeys_repository.go`
- `backend/golang/internal/database/helpers_repository.go`
- `backend/golang/internal/database/connections_repository.go`
- `backend/golang/internal/database/executions_repository.go`

### 3.5 API Services
Each service gets its own `serverless.yml` and consolidated Lambda handler:

| Service | Endpoints | Lambda |
|---------|-----------|--------|
| auth | POST /login, /register, /refresh, /verify | auth-public |
| users | GET/PUT /users/me | users-private |
| accounts | GET/PUT/POST /accounts | accounts-private |
| helpers | GET/POST/PUT/DELETE /helpers, POST /helpers/{id}/execute | helpers-private |
| connections | GET/POST/PUT/DELETE /connections, GET /connections/{id}/callback | connections-private |
| api-keys | GET/POST/DELETE /api-keys | api-keys-private |

### 3.6 Queue System (SQS + Lambda)
**All helper executions go through a queue:**

```
┌─────────────────┐      ┌─────────────────┐      ┌─────────────────┐
│  API Gateway    │ ──▶  │  SQS Queue      │ ──▶  │  Worker Lambda  │
│  (receives req) │      │  (FIFO)         │      │  (executes)     │
└─────────────────┘      └─────────────────┘      └─────────────────┘
                                                           │
                                                           ▼
                                                  ┌─────────────────┐
                                                  │  CRM API        │
                                                  │  (Keap, etc.)   │
                                                  └─────────────────┘
```

---

## Phase 4: CRM Abstraction Layer

### 4.1 Unified Connector Interface
**File**: `backend/golang/internal/connectors/unified_interface.go`

```go
type CRMConnector interface {
    Authenticate(ctx context.Context, credentials AuthCredentials) error
    RefreshAuthentication(ctx context.Context) error
    GetAuthStatus() AuthenticationStatus

    GetContacts(ctx context.Context, opts QueryOptions) ([]NormalizedContact, string, error)
    GetContact(ctx context.Context, id string) (*NormalizedContact, error)
    CreateContact(ctx context.Context, contact NormalizedContact) (*NormalizedContact, error)
    UpdateContact(ctx context.Context, id string, contact NormalizedContact) (*NormalizedContact, error)

    GetTags(ctx context.Context) ([]Tag, error)
    ApplyTag(ctx context.Context, contactID, tagID string) error
    RemoveTag(ctx context.Context, contactID, tagID string) error

    GetCustomFields(ctx context.Context) ([]CustomField, error)

    TriggerAutomation(ctx context.Context, contactID, automationID string) error
    AchieveGoal(ctx context.Context, contactID, goalName string) error

    GetHealthStatus(ctx context.Context) (*HealthStatus, error)
    GetMetadata() *ConnectorMetadata
    GetCapabilities() []string
}
```

### 4.2 Platform Implementations
Priority order:
1. **Keap** - `backend/golang/internal/connectors/keap.go`
2. **GoHighLevel** - `backend/golang/internal/connectors/gohighlevel.go`
3. **ActiveCampaign** - `backend/golang/internal/connectors/activecampaign.go`
4. **Ontraport** - `backend/golang/internal/connectors/ontraport.go`
5. **Generic HTTP POST** - `backend/golang/internal/connectors/webhook.go`

### 4.3 Platform Seed Data
**Location**: `backend/golang/services/api/platforms/ci_cd/seed/`

Platform definitions (auth config, API URLs, rate limits, capabilities) stored as JSON seed data.

---

## Phase 5: Helper System Modernization

### 5.1 Helper Architecture
**Location**: `backend/golang/internal/helpers/`

```
internal/helpers/
├── interface.go              # Helper interface
├── registry.go               # Helper registration/lookup
├── executor.go               # Execution engine
├── contact/                  # Contact manipulation
│   ├── assign_it.go
│   ├── copy_it.go
│   ├── merge_it.go
│   └── field_to_field.go
├── data/                     # Data transformation
│   ├── format_it.go
│   ├── math.go
│   ├── date_calc.go
│   └── text_it.go
├── tagging/                  # Tag management
│   ├── tag_it.go
│   └── clear_tags.go
├── automation/               # Automation triggers
│   ├── goal_it.go
│   ├── trigger_it.go
│   └── action_it.go
├── integration/              # External integrations
│   ├── google_sheet_it.go
│   ├── slack_it.go
│   ├── trello_it.go
│   └── sms_it.go
├── notification/
│   └── notify_me.go
└── analytics/
    ├── rfm_calculation.go
    └── customer_lifetime_value.go
```

### 5.2 Helper Interface
```go
type Helper interface {
    GetName() string
    GetCategory() string
    GetDescription() string
    GetConfigSchema() map[string]interface{}

    Execute(ctx context.Context, input HelperInput) (*HelperOutput, error)
    ValidateConfig(config map[string]interface{}) error

    RequiresCRM() bool
    SupportedCRMs() []string
}
```

### 5.3 Migration from PHP
For each helper in legacy `/core-api/app/mfh_functions/`:
1. Analyze PHP implementation
2. Create Go implementation with same logic
3. Map field indices to named config fields
4. Add error handling and logging
5. Write tests

---

## Phase 6: AI-First Features

### 6.1 Natural Language Automation Builder
- "When a contact is tagged with 'New Lead', send them to my Google Sheet and notify me on Slack"
- AI parses intent → generates helper chain configuration
- Uses **Groq** for fast, cost-effective inference

### 6.2 AI-Powered Business Intelligence
- Natural language queries on CRM data
- Auto-generated reports and summaries
- Trend analysis and forecasting
- Custom report builder

### 6.3 AI Email Assistant
- AI-assisted email composition with CRM personalization
- Template generation
- Subject line optimization
- Response suggestions

### 6.4 Conversational CRM Assistant
- Floating chat interface
- Quick data lookups
- Action execution via natural language
- Help & guidance

---

## Phase 7: Billing & Teams

### 7.1 Stripe Integration
- Stripe tied to Account (not User)
- Subscription management
- Usage-based billing for helper executions
- Plan limits enforcement

**Plans:**
| Plan | Helpers | Executions/mo | Connections | Team Members | API Keys |
|------|---------|---------------|-------------|--------------|----------|
| Free | 5 | 1,000 | 1 | 1 | 1 |
| Pro | 50 | 50,000 | 5 | 5 | 10 |
| Business | Unlimited | Unlimited | Unlimited | Unlimited | Unlimited |

### 7.2 Multi-User Accounts
- User groups: owner, admin, member, viewer (already in Cognito)
- Team invitations
- Granular permissions per user per account
- Account switching

---

## Phase 8: Migration Strategy

### 8.1 Parallel Operation
- Run both systems simultaneously during transition
- New users → myfusionhelper.ai
- Existing users → option to migrate

### 8.2 Data Migration
- Export user configurations from Gravity Forms
- Transform to new helper configuration format
- Import to DynamoDB

### 8.3 Legacy API Compatibility
- Maintain compatible endpoint for transition
- Proxy to new Go backend
- Gradual deprecation

### 8.4 User Migration Flow
1. User logs into new system
2. Connect CRM (re-auth OAuth)
3. Import existing helpers (automated)
4. Test helpers work correctly
5. Deprecate old account

---

## Key Files Reference

### Infrastructure (Serverless Framework)
| Path | Purpose | Status |
|------|---------|--------|
| `backend/golang/services/infrastructure/cognito/serverless.yml` | Cognito user pool + client | ✅ Created |
| `backend/golang/services/infrastructure/dynamodb/core/serverless.yml` | All DynamoDB tables | ✅ Created |
| `backend/golang/services/infrastructure/s3/serverless.yml` | S3 buckets | TODO |
| `backend/golang/services/infrastructure/sqs/serverless.yml` | SQS queues | TODO |
| `backend/golang/services/api/gateway/serverless.yml` | API Gateway + custom domain | TODO |
| `backend/golang/services/api/auth/serverless.yml` | Auth Lambda service | TODO |

### Backend (Go)
| Path | Purpose | Status |
|------|---------|--------|
| `backend/golang/cmd/handlers/auth/main.go` | Auth service router | TODO |
| `backend/golang/cmd/handlers/accounts/main.go` | Accounts service router | TODO |
| `backend/golang/cmd/handlers/helpers/main.go` | Helper service router | TODO |
| `backend/golang/internal/connectors/unified_interface.go` | CRM interface | TODO |
| `backend/golang/internal/connectors/keap.go` | Keap implementation | TODO |
| `backend/golang/internal/helpers/interface.go` | Helper interface | TODO |
| `backend/golang/internal/middleware/auth/auth.go` | JWT auth middleware | TODO |
| `backend/golang/internal/middleware/apikey/apikey.go` | API key auth middleware | TODO |
| `backend/golang/internal/database/users_repository.go` | Users DB access | TODO |

### Frontend (Next.js)
| Path | Purpose | Status |
|------|---------|--------|
| `apps/web/src/app/(dashboard)/helpers/page.tsx` | Helper library page | ✅ Created |
| `apps/web/src/app/(dashboard)/connections/page.tsx` | CRM connections | ✅ Created |
| `apps/web/src/app/(dashboard)/executions/page.tsx` | Execution history | ✅ Created |
| `apps/web/src/app/(dashboard)/insights/page.tsx` | AI insights | ✅ Created |
| `apps/web/src/app/(dashboard)/settings/page.tsx` | Settings | ✅ Created |
| `apps/web/src/lib/auth.ts` | Auth config (Better Auth → Cognito) | ⚠️ Needs migration |
| `apps/web/src/lib/auth-client.ts` | Auth client (Better Auth → Cognito) | ⚠️ Needs migration |
| `apps/web/src/lib/db.ts` | DB connection (Neon → remove) | ⚠️ Remove |

---

## Design System

### Design Principles
- Minimal, focused interfaces with whitespace
- Dark mode support from day 1
- AI suggestions surface contextually
- Progressive disclosure - simple by default, powerful when needed
- Real-time feedback and animations

### Design Inspiration
- Linear (clean, minimal, keyboard-first)
- Vercel Dashboard (modern, dark mode)
- Notion (AI integration patterns)
- Raycast (command palette, quick actions)

### Core Components (shadcn/ui base)
- `HelperCard` - Helper preview in library
- `HelperConfigForm` - Dynamic form based on helper schema (JSON Schema driven)
- `ConnectionCard` - CRM connection status
- `ExecutionLog` - Execution timeline/details
- `CRMSelector` - Platform picker dropdown

### AI-First Components
- `AIInsightCard` - Surfaced AI insight with action
- `AIChatBubble` - Floating chat assistant
- `AIEmailComposer` - Email writing interface with AI
- `AIReportBuilder` - Natural language report creation
- `MetricCard` - KPI display with trend indicator
- `DataTable` - Sortable, filterable data grid with AI search

### Helper Configuration UI
- Unified builder interface (SPA approach)
- Dynamic config forms based on helper type (JSON Schema driven)
- "Advanced mode" toggle for complex helpers
- Real-time preview of what the helper will do
- Break out individual pages only if specific helpers prove too complex

---

## Verification Plan

### Backend Testing
1. Unit tests for each CRM connector
2. Unit tests for each helper
3. Integration tests for helper execution flow
4. Load testing for concurrent executions

### Frontend Testing
1. Component tests with React Testing Library
2. E2E tests with Playwright
3. Visual regression tests

### Manual Verification
1. Create CRM connection (OAuth flow)
2. Configure a helper (e.g., tag_it)
3. Execute helper on test contact
4. Verify CRM reflects changes
5. Check execution logs

---

## Success Metrics

**Core Platform:**
- [ ] All 90+ helpers migrated and functional
- [ ] At least 4 CRM platforms supported
- [ ] <500ms average helper execution time
- [ ] <100ms cold start (consolidated handlers)
- [ ] 99.9% uptime

**AI-First Features:**
- [ ] AI chat assistant responds in <2 seconds
- [ ] Email composer generates drafts in <3 seconds
- [ ] Reports can be created via natural language
- [ ] Insights surface proactively on dashboard
- [ ] Dark mode + light mode fully implemented

**User Experience:**
- [ ] Mobile-responsive design
- [ ] Keyboard shortcuts for power users
- [ ] <3 second page load times
- [ ] Existing users successfully migrated
