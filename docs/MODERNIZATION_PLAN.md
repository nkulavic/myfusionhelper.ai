# MyFusionHelper.ai Modernization Plan

## Overview

Rebuild the MyFusion Helper platform from a WordPress-based CRM automation system to a modern, AI-first, multi-platform CRM automation platform.

**Current State**: `secure.myfusionsolutions.com` - WordPress + PHP with 90+ helper microservices, Keap-only integration
**Target State**: `myfusionhelper.ai` - Go backend + Next.js frontend, multi-CRM support, AI-powered automation

## Development Approach

**Phase 1 Priority: UI/UX Prototype on Vercel**
- Start with frontend design and prototype
- Use Vercel-hosted Next.js with Vercel KV/Upstash for initial database
- Iterate on UI/UX before building full backend
- Migrate to DynamoDB + Go backend when ready

**CRM Platform Priority:**
1. Keap (existing users)
2. GoHighLevel (high demand)
3. ActiveCampaign
4. Ontraport
5. Any platform with HTTP POST webhook functionality

**AI Provider:** Groq (cost-effective, fast inference)

---

## MVP vs Full Feature Scope

**MVP (Phase 1-2):**
- UI prototype with mock data
- Basic auth (Cognito + JWT)
- Single user/account (no team features)
- API key generation (basic)
- 1-2 CRM connections (Keap, GoHighLevel)
- Core helpers (10-15 most used)
- Basic execution logging

**Full Product (Phase 3+):**
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

## Architecture Decision

### Domain Structure
- `myfusionhelper.ai` - Marketing/sales site (public)
- `app.myfusionhelper.ai` - Main application (authenticated)
- `api.myfusionhelper.ai` - Go backend API

### Monorepo Structure
```
myfusionhelper.ai/
├── apps/
│   ├── marketing/              # Next.js - myfusionhelper.ai
│   │   ├── app/
│   │   ├── components/
│   │   └── package.json
│   └── web/                    # Next.js - app.myfusionhelper.ai
│       ├── app/
│       │   ├── (auth)/         # Login, register, forgot password
│       │   ├── (dashboard)/    # Protected routes
│       │   └── api/            # API route proxies
│       ├── components/
│       │   ├── ui/             # shadcn/ui components
│       │   ├── helpers/        # Helper configuration UI
│       │   └── crm/            # CRM-specific components
│       ├── lib/
│       │   ├── api/            # API client modules
│       │   └── stores/         # Zustand state stores
│       └── package.json
├── backend/
│   └── golang/
│       ├── cmd/
│       │   └── handlers/       # Lambda handlers (1 per service)
│       │       ├── auth/
│       │       ├── accounts/
│       │       ├── helpers/    # Helper execution service
│       │       ├── crm/        # CRM operations
│       │       └── platforms/  # Platform connections
│       ├── internal/
│       │   ├── connectors/     # Unified CRM connector interface
│       │   ├── database/       # DynamoDB repositories
│       │   ├── helpers/        # Helper implementations (90+)
│       │   ├── middleware/     # Auth, logging, rate limiting
│       │   └── services/       # Business logic
│       └── services/
│           └── api/            # Serverless service definitions
│               └── platforms/
│                   └── ci_cd/
│                       └── seed/   # Platform definitions
│                           ├── keap/
│                           ├── hubspot/
│                           ├── salesforce/
│                           └── pipedrive/
├── packages/
│   ├── ui/                     # Shared UI components
│   ├── types/                  # Shared TypeScript types
│   └── config/                 # Shared configs (eslint, tsconfig)
├── docs/
│   ├── architecture/
│   └── api/
├── turbo.json                  # Turborepo config
└── package.json                # Root package.json
```

---

## Phase 1: UI/UX Prototype on Vercel (Weeks 1-4)

**PRIORITY #1: UI/UX Design First**

Build a beautiful, modern interface before worrying about backend. Get the design right, then build to support it.

**Goal**: Get a working UI prototype deployed on Vercel with mock data, iterate on design before building full backend.

### 1.1 Repository Setup
- [ ] Initialize monorepo with Turborepo in `myfusionhelper.ai` repo
- [ ] Set up `apps/web` with Next.js 15 + React 19
- [ ] Set up `apps/marketing` with Next.js (simpler, static-focused)
- [ ] Deploy to Vercel (app.myfusionhelper.ai, myfusionhelper.ai)
- [ ] Configure shadcn/ui component library

### 1.2 Initial Database (Vercel-Compatible)
Use Vercel KV (Redis) or Upstash for initial development - easy to migrate to DynamoDB later:
- [ ] `users` - User accounts
- [ ] `accounts` - Workspaces/teams
- [ ] `connections` - CRM connections (mock initially)
- [ ] `helpers` - Helper configurations
- [ ] `executions` - Execution logs

### 1.3 UI Prototype Pages
Build out the core UI with mock data:

**Core Automation:**
- [ ] **Dashboard** (`/`) - Overview, recent executions, quick stats, AI insights
- [ ] **Helpers Library** (`/helpers`) - Browse all 90+ helpers by category
- [ ] **Helper Builder** (`/helpers/new`) - Visual helper configuration (see below)
- [ ] **Helper Detail** (`/helpers/[id]`) - Edit existing helper
- [ ] **Connections** (`/connections`) - CRM connection management
- [ ] **Executions** (`/executions`) - Execution history with logs

**Helper Configuration UI Options:**

*Option A: Single Page App (Recommended)*
- One unified builder interface
- Select helper type from dropdown/sidebar
- Config form dynamically loads based on helper type
- Preview panel shows what will happen
- All in one screen, no page navigation

*Option B: Individual Pages*
- Each helper has its own route (`/helpers/tag-it/new`)
- More focused, but more pages to build
- Easier for complex helpers with lots of config

**Recommendation:** Start with SPA approach, design for extensibility:
- All helpers in one unified builder for MVP
- Dynamic config forms based on helper type (JSON Schema driven)
- "Advanced mode" toggle for complex helpers with more options
- Expandable sections for optional configurations
- If specific helpers prove too complex, break them out later
- Real-time preview of what the helper will do

**Business Intelligence (AI-First):**
- [ ] **Insights** (`/insights`) - AI-powered analytics and reports
- [ ] **Reports** (`/reports`) - Custom report builder
- [ ] **Reports Detail** (`/reports/[id]`) - Individual report view

**Communication:**
- [ ] **Emails** (`/emails`) - AI email composer and templates
- [ ] **Email Templates** (`/emails/templates`) - Template library

**Settings:**
- [ ] **Settings** (`/settings`) - Account, billing, team
- [ ] **API Keys** (`/settings/api-keys`) - Manage API keys for helper execution
- [ ] **Team** (`/settings/team`) - Invite users, manage roles
- [ ] **Billing** (`/settings/billing`) - Stripe subscription, usage
- [ ] **AI Chat** (floating component) - Conversational CRM assistant

### 1.4 Design System & Component Library

**Design Principles (AI-First, Clean, Modern):**
- Minimal, focused interfaces with lots of whitespace
- Dark mode support from day 1
- AI suggestions surface contextually (not overwhelming)
- Progressive disclosure - simple by default, powerful when needed
- Real-time feedback and animations

**Core Components (shadcn/ui base):**
- [ ] `HelperCard` - Helper preview in library
- [ ] `HelperConfigForm` - Dynamic form based on helper schema
- [ ] `ConnectionCard` - CRM connection status
- [ ] `ExecutionLog` - Execution timeline/details
- [ ] `CRMSelector` - Platform picker dropdown

**AI-First Components:**
- [ ] `AIInsightCard` - Surfaced AI insight with action
- [ ] `AIChatBubble` - Floating chat assistant
- [ ] `AIEmailComposer` - Email writing interface with AI
- [ ] `AIReportBuilder` - Natural language report creation
- [ ] `AISuggestionBanner` - Contextual suggestions
- [ ] `MetricCard` - KPI display with trend indicator
- [ ] `ChartWidget` - Configurable data visualization
- [ ] `DataTable` - Sortable, filterable data grid with AI search

**Design Inspiration:**
- Linear (clean, minimal, keyboard-first)
- Vercel Dashboard (modern, dark mode)
- Notion (AI integration patterns)
- Raycast (command palette, quick actions)

### 1.5 State Management (Zustand)
Set up stores for prototype:
```typescript
// lib/stores/
├── auth-store.ts       // User auth state
├── workspace-store.ts  // Current workspace/account
├── helper-store.ts     // Helper configuration state
└── ui-store.ts         // UI state (modals, sidebars)
```

---

## Phase 2: Backend Foundation (Weeks 5-8)

When UI is validated, build the real backend.

### 2.1 Account & User Model (like listbackup.ai)

**Core Entities:**

```go
// User - Individual person (linked to Cognito)
type User struct {
    UserID           string    `json:"user_id"`           // "user:uuid"
    CognitoUserID    string    `json:"cognito_user_id"`
    Email            string    `json:"email"`
    Name             string    `json:"name"`
    CurrentAccountID string    `json:"current_account_id"` // Active workspace
    CreatedAt        time.Time `json:"created_at"`
}

// Account - Billing entity (workspace)
type Account struct {
    AccountID        string    `json:"account_id"`        // "account:uuid"
    OwnerUserID      string    `json:"owner_user_id"`
    Name             string    `json:"name"`              // "My Business"
    Company          string    `json:"company"`
    Plan             string    `json:"plan"`              // "free", "pro", "business"
    Status           string    `json:"status"`            // "active", "suspended"
    StripeCustomerID string    `json:"stripe_customer_id"` // Billing tied to account
    Settings         AccountSettings `json:"settings"`
    Usage            AccountUsage    `json:"usage"`
}

// UserAccount - Many-to-many with roles
type UserAccount struct {
    UserID      string          `json:"user_id"`
    AccountID   string          `json:"account_id"`
    Role        string          `json:"role"`        // "owner", "admin", "member", "viewer"
    Permissions UserPermissions `json:"permissions"`
    LinkedAt    time.Time       `json:"linked_at"`
}

// UserPermissions - Granular permissions
type UserPermissions struct {
    CanManageHelpers      bool `json:"can_manage_helpers"`
    CanExecuteHelpers     bool `json:"can_execute_helpers"`
    CanManageConnections  bool `json:"can_manage_connections"`
    CanManageTeam         bool `json:"can_manage_team"`
    CanManageBilling      bool `json:"can_manage_billing"`
    CanViewAnalytics      bool `json:"can_view_analytics"`
    CanManageAPIKeys      bool `json:"can_manage_api_keys"`
}
```

### 2.2 Authentication Strategy

**Two Authentication Methods:**

1. **JWT (App Authentication)** - For web app users
   - Cognito User Pool for identity
   - JWT tokens with `sub` claim
   - AuthContext includes UserID, AccountID, Permissions
   - Used for: Dashboard, settings, configuration

2. **API Keys (Helper Execution)** - For external triggers (CRM webhooks, HTTP calls)
   - API key tied to Account (not user)
   - Format: `mfh_live_xxxxxxxxxxxx` or `mfh_test_xxxxxxxxxxxx`
   - Scoped permissions (execute-only, full access)
   - Rate limited per account/plan
   - Used for: Helper execution from CRM automations

**API Key Model:**
```go
type APIKey struct {
    KeyID       string    `json:"key_id"`        // "apikey:uuid"
    AccountID   string    `json:"account_id"`
    CreatedBy   string    `json:"created_by"`    // UserID who created it
    Name        string    `json:"name"`          // "Keap Production"
    KeyHash     string    `json:"key_hash"`      // Hashed key (never store raw)
    KeyPrefix   string    `json:"key_prefix"`    // "mfh_live_abc" for display
    Permissions []string  `json:"permissions"`   // ["execute_helpers", "read_logs"]
    Status      string    `json:"status"`        // "active", "revoked"
    LastUsedAt  *time.Time `json:"last_used_at"`
    CreatedAt   time.Time `json:"created_at"`
    ExpiresAt   *time.Time `json:"expires_at"`   // Optional expiration
}
```

**Authentication Flow:**

```
┌─────────────────────────────────────────────────────────────┐
│                     Web App (Dashboard)                      │
│  JWT Token → Cognito Validation → AuthContext → Handler     │
└─────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────┐
│              CRM Webhook / Helper Execution                  │
│  API Key → Key Validation → AccountContext → Execute Helper │
│  POST /api/v1/helpers/{helper_id}/execute                   │
│  Header: X-API-Key: mfh_live_xxxxx                          │
└─────────────────────────────────────────────────────────────┘
```

### 2.3 Billing Integration (Stripe) - Full Architecture

**Note:** Full billing is Phase 3+. MVP can start with single plan or free tier. Architecture designed to support full billing from day 1.

**Stripe tied to Account (not User):**
- StripeCustomerID stored on Account
- Subscriptions managed at Account level
- Usage-based billing for helper executions
- Plan limits enforced at Account level

**Billing Service Endpoints (like listbackup.ai):**
```
/billing/plans                    - List available plans
/billing/subscriptions            - Get current subscription
/billing/subscriptions/create     - Create new subscription
/billing/subscriptions/upgrade    - Upgrade plan
/billing/subscriptions/downgrade  - Downgrade plan
/billing/subscriptions/cancel     - Cancel subscription
/billing/subscriptions/resume     - Resume cancelled subscription
/billing/checkout                 - Create checkout session
/billing/portal                   - Create Stripe portal session
/billing/invoices                 - List invoices
/billing/payment-methods          - Manage payment methods
/billing/usage                    - Get usage metrics
/billing/webhook                  - Stripe webhook handler
```

**Plans:**
```go
var Plans = map[string]PlanLimits{
    "free": {
        MaxHelpers:       5,
        MaxExecutions:    1000,  // per month
        MaxConnections:   1,
        MaxTeamMembers:   1,
        MaxAPIKeys:       1,
    },
    "pro": {
        MaxHelpers:       50,
        MaxExecutions:    50000,
        MaxConnections:   5,
        MaxTeamMembers:   5,
        MaxAPIKeys:       10,
    },
    "business": {
        MaxHelpers:       -1,    // unlimited
        MaxExecutions:    -1,
        MaxConnections:   -1,
        MaxTeamMembers:   -1,
        MaxAPIKeys:       -1,
    },
}
```

**Usage Tracking:**
- Track executions per account per month
- Track API calls per API key
- Enforce limits before execution
- Overage handling (block or charge)

### 2.4 Queue System (SQS + Lambda)

**All helper executions go through a queue:**
- Prevents timeouts on long-running helpers
- Enables retry logic
- Rate limiting per account
- Batch processing for bulk operations

**Architecture:**
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

**Queue Types:**
- `mfh-{stage}-helper-execution.fifo` - Main execution queue (FIFO for ordering)
- `mfh-{stage}-helper-execution-dlq.fifo` - Dead letter queue for failed executions
- `mfh-{stage}-webhooks` - Inbound webhook processing

**Worker Lambda:**
```go
// cmd/handlers/helper-worker/main.go
func HandleSQSEvent(ctx context.Context, event events.SQSEvent) error {
    for _, record := range event.Records {
        var job HelperExecutionJob
        json.Unmarshal([]byte(record.Body), &job)

        // Execute helper
        result, err := executeHelper(ctx, job)

        // Log execution
        logExecution(ctx, job, result, err)

        // Handle retry if needed
        if err != nil && job.RetryCount < 3 {
            requeueWithDelay(job)
        }
    }
    return nil
}
```

### 2.5 Infrastructure (AWS)
- [ ] DynamoDB tables:
  - `mfh-{stage}-users`
  - `mfh-{stage}-accounts`
  - `mfh-{stage}-user-accounts` (relationship table)
  - `mfh-{stage}-api-keys`
  - `mfh-{stage}-platform-connections`
  - `mfh-{stage}-helpers`
  - `mfh-{stage}-executions`
  - `mfh-{stage}-oauth-states`
- [ ] S3 buckets for file storage
- [ ] SQS queues:
  - `mfh-{stage}-helper-execution.fifo`
  - `mfh-{stage}-helper-execution-dlq.fifo`
  - `mfh-{stage}-webhooks`
- [ ] Cognito user pool
- [ ] API Gateway HTTP API
- [ ] Lambda functions:
  - API handlers (consolidated pattern)
  - SQS worker (helper execution)
  - Scheduled jobs (token refresh, cleanup)

### 1.3 Authentication Service
**Files to create:**
- `backend/golang/cmd/handlers/auth/main.go` - Router
- `backend/golang/cmd/handlers/auth/clients/login/main.go`
- `backend/golang/cmd/handlers/auth/clients/register/main.go`
- `backend/golang/cmd/handlers/auth/clients/refresh/main.go`
- `backend/golang/internal/middleware/auth/auth.go`

**Pattern** (from listbackup.ai):
```go
// Consolidated handler - single Lambda entry point
func HandlePublic(ctx context.Context, event events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
    path := event.RequestContext.HTTP.Path
    method := event.RequestContext.HTTP.Method

    switch {
    case method == "POST" && strings.HasSuffix(path, "/login"):
        return login.Handle(ctx, event)
    case method == "POST" && strings.HasSuffix(path, "/register"):
        return register.Handle(ctx, event)
    // ... more routes
    }
}
```

---

## Phase 3: CRM Abstraction Layer (Weeks 9-12)

### 2.1 Unified Connector Interface
**File**: `backend/golang/internal/connectors/unified_interface.go`

```go
type CRMConnector interface {
    // Authentication
    Authenticate(ctx context.Context, credentials AuthCredentials) error
    RefreshAuthentication(ctx context.Context) error
    GetAuthStatus() AuthenticationStatus

    // Contacts
    GetContacts(ctx context.Context, opts QueryOptions) ([]NormalizedContact, string, error)
    GetContact(ctx context.Context, id string) (*NormalizedContact, error)
    CreateContact(ctx context.Context, contact NormalizedContact) (*NormalizedContact, error)
    UpdateContact(ctx context.Context, id string, contact NormalizedContact) (*NormalizedContact, error)

    // Tags
    GetTags(ctx context.Context) ([]Tag, error)
    ApplyTag(ctx context.Context, contactID, tagID string) error
    RemoveTag(ctx context.Context, contactID, tagID string) error

    // Custom Fields
    GetCustomFields(ctx context.Context) ([]CustomField, error)

    // Automations/Sequences
    TriggerAutomation(ctx context.Context, contactID, automationID string) error
    AchieveGoal(ctx context.Context, contactID, goalName string) error

    // Health & Metadata
    GetHealthStatus(ctx context.Context) (*HealthStatus, error)
    GetMetadata() *ConnectorMetadata
    GetCapabilities() []string
}
```

### 2.2 Normalized Data Models
**File**: `backend/golang/internal/connectors/models.go`

```go
type NormalizedContact struct {
    ID           string                 `json:"id"`
    FirstName    string                 `json:"first_name"`
    LastName     string                 `json:"last_name"`
    Email        string                 `json:"email"`
    Phone        string                 `json:"phone"`
    Company      string                 `json:"company"`
    Tags         []string               `json:"tags"`
    CustomFields map[string]interface{} `json:"custom_fields"`
    SourceCRM    string                 `json:"source_crm"` // keap, hubspot, etc.
    SourceID     string                 `json:"source_id"`
    CreatedAt    time.Time              `json:"created_at"`
    UpdatedAt    time.Time              `json:"updated_at"`
}
```

### 2.3 Platform Implementations
Priority order:
1. **Keap** (existing users) - `backend/golang/internal/connectors/keap.go`
2. **GoHighLevel** (high demand) - `backend/golang/internal/connectors/gohighlevel.go`
3. **ActiveCampaign** - `backend/golang/internal/connectors/activecampaign.go`
4. **Ontraport** - `backend/golang/internal/connectors/ontraport.go`
5. **Generic HTTP POST** - `backend/golang/internal/connectors/webhook.go` (any platform with HTTP POST capability)

**Note**: Many CRMs support triggering automations via HTTP POST/webhooks. The Generic HTTP POST connector enables support for any platform with this capability without building a full connector.

### 2.4 Platform Seed Data
**Location**: `backend/golang/services/api/platforms/ci_cd/seed/`

Example for Keap:
```json
// keap/platform.json
{
  "id": "keap",
  "name": "Keap (Infusionsoft)",
  "category": "crm",
  "auth_type": "oauth2",
  "oauth_config": {
    "authorization_url": "https://accounts.infusionsoft.com/app/oauth/authorize",
    "token_url": "https://api.infusionsoft.com/token",
    "scopes": ["full"]
  },
  "api_base_url": "https://api.infusionsoft.com/crm/rest/v2",
  "rate_limit": {
    "requests_per_second": 10,
    "daily_limit": 25000
  },
  "capabilities": ["contacts", "tags", "custom_fields", "automations", "goals", "deals"]
}
```

---

## Phase 4: Helper System Modernization (Weeks 13-16)

### 3.1 Helper Architecture
**Location**: `backend/golang/internal/helpers/`

Organize helpers by category:
```
internal/helpers/
├── interface.go              # Helper interface definition
├── registry.go               # Helper registration/lookup
├── executor.go               # Helper execution engine
├── contact/                  # Contact manipulation helpers
│   ├── assign_it.go
│   ├── copy_it.go
│   ├── merge_it.go
│   └── field_to_field.go
├── data/                     # Data transformation helpers
│   ├── format_it.go
│   ├── math.go
│   ├── date_calc.go
│   └── text_it.go
├── tagging/                  # Tag management helpers
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
├── notification/             # Notification helpers
│   └── notify_me.go
└── analytics/                # Analytics helpers
    ├── rfm_calculation.go
    └── customer_lifetime_value.go
```

### 3.2 Helper Interface
```go
type Helper interface {
    // Metadata
    GetName() string
    GetCategory() string
    GetDescription() string
    GetConfigSchema() map[string]interface{}  // JSON Schema for config

    // Execution
    Execute(ctx context.Context, input HelperInput) (*HelperOutput, error)

    // Validation
    ValidateConfig(config map[string]interface{}) error

    // Capabilities
    RequiresCRM() bool
    SupportedCRMs() []string  // Empty = all CRMs
}

type HelperInput struct {
    ContactID    string
    ContactData  NormalizedContact
    Config       map[string]interface{}
    CRMConnector CRMConnector
    UserID       string
    AccountID    string
}

type HelperOutput struct {
    Success       bool
    ModifiedData  map[string]interface{}
    Actions       []HelperAction  // Tags to apply, fields to update, etc.
    Logs          []string
}
```

### 3.3 Migration from PHP Helpers
For each helper in `/core-api/app/mfh_functions/`:
1. Analyze PHP implementation
2. Create Go implementation with same logic
3. Map field indices to named config fields
4. Add proper error handling and logging
5. Write tests

---

## Phase 5: AI-First Features (Weeks 17-20)

### 4.1 Natural Language Automation Builder
**Service**: `backend/golang/cmd/handlers/ai/`

Features:
- "When a contact is tagged with 'New Lead', send them to my Google Sheet and notify me on Slack"
- AI parses intent → generates helper chain configuration
- Uses **Groq** for fast, cost-effective inference (LLaMA 3.x models)

**Groq Integration:**
```go
// internal/services/ai/groq.go
type GroqService struct {
    apiKey string
    model  string // "llama-3.1-70b-versatile" or "llama-3.1-8b-instant"
}

func (s *GroqService) ParseAutomationIntent(prompt string) (*AutomationConfig, error) {
    // Call Groq API with structured output
    // Return parsed helper chain configuration
}
```

### 4.2 AI-Powered Data Analysis & Business Intelligence
**Location**: `apps/web/app/(dashboard)/insights/`

Natural language queries on CRM data:
- "Show me contacts who haven't been emailed in 30 days"
- "What's my best performing lead source this month?"
- "Compare this month's sales to last month"

**Smart Reports Dashboard:**
- Auto-generated daily/weekly/monthly summaries
- Trend analysis and forecasting
- Custom report builder with AI suggestions
- Export to PDF, CSV, Google Sheets

### 4.3 AI Email Assistant
**Location**: `apps/web/app/(dashboard)/emails/`

Features:
- **Email Composer**: AI-assisted email writing
  - "Write a follow-up email for cold leads"
  - Tone adjustment (professional, friendly, urgent)
  - Personalization tokens from CRM data
- **Email Templates**: AI-generated templates library
- **Subject Line Generator**: Test multiple variations
- **Response Suggestions**: Quick reply recommendations

**Implementation:**
```go
// internal/services/ai/email_assistant.go
type EmailAssistant struct {
    groq *GroqService
}

func (e *EmailAssistant) ComposeEmail(ctx context.Context, req EmailRequest) (*EmailDraft, error) {
    // Use Groq to generate email based on context
    // Include CRM contact data for personalization
}

func (e *EmailAssistant) SuggestSubjectLines(ctx context.Context, body string) ([]string, error) {
    // Generate 3-5 subject line options
}
```

### 4.4 Intelligent Suggestions & Insights
- Suggest automations based on user patterns
- Identify optimization opportunities
- Anomaly detection on CRM data
- "You have 47 contacts without a tag - want to segment them?"
- "Your email open rate dropped 15% - here's why"

### 4.5 Conversational CRM Assistant
**Location**: `apps/web/components/chat/`

Floating AI chat interface for:
- Quick data lookups: "How many new leads this week?"
- Action execution: "Tag all contacts from California with 'West Coast'"
- Report generation: "Create a report of top customers"
- Help & guidance: "How do I set up a webhook?"

**UI Component:**
```typescript
// components/chat/CRMAssistant.tsx
- Floating button in bottom-right
- Chat interface with message history
- Action previews before execution
- Voice input support (optional)
```

---

## Phase 6: Frontend Polish & Integration (Ongoing)

### 5.1 Core Pages
- `/` - Dashboard overview
- `/helpers` - Helper library and configuration
- `/helpers/[id]` - Individual helper setup
- `/connections` - CRM connections management
- `/connections/[platform]/callback` - OAuth callbacks
- `/executions` - Execution history/logs
- `/settings` - Account settings

### 5.2 State Management (Zustand)
```typescript
// lib/stores/auth-store.ts
interface AuthState {
  user: User | null
  isAuthenticated: boolean
  login: (email: string, password: string) => Promise<void>
  logout: () => void
}

// lib/stores/workspace-store.ts
interface WorkspaceState {
  currentAccount: Account | null
  connections: PlatformConnection[]
  activeConnection: PlatformConnection | null
  setActiveConnection: (id: string) => void
}

// lib/stores/helper-store.ts
interface HelperState {
  helpers: Helper[]
  selectedHelper: Helper | null
  helperConfig: Record<string, any>
  updateConfig: (key: string, value: any) => void
}
```

### 5.3 API Client Pattern
```typescript
// lib/api/helpers.ts
export const helpersApi = {
  list: () => apiClient.get<Helper[]>('/helpers'),
  get: (id: string) => apiClient.get<Helper>(`/helpers/${id}`),
  create: (data: CreateHelperDto) => apiClient.post<Helper>('/helpers', data),
  execute: (id: string, contactId: string) =>
    apiClient.post<ExecutionResult>(`/helpers/${id}/execute`, { contactId }),
}
```

---

## Phase 7: Migration Strategy (Future)

### 6.1 Parallel Operation
- Run both systems simultaneously during transition
- New users → myfusionhelper.ai
- Existing users → option to migrate

### 6.2 Data Migration
- Export user configurations from Gravity Forms
- Transform to new helper configuration format
- Import to DynamoDB

### 6.3 Legacy API Compatibility
- Maintain `/core-api/app/mfh-app.php` compatible endpoint
- Proxy to new Go backend
- Gradual deprecation

### 6.4 User Migration Flow
1. User logs into new system
2. Connect CRM (re-auth OAuth)
3. Import existing helpers (automated)
4. Test helpers work correctly
5. Deprecate old account

---

## Key Files Reference

### Backend (Go)
| Path | Purpose |
|------|---------|
| `backend/golang/cmd/handlers/auth/main.go` | Auth service router |
| `backend/golang/cmd/handlers/accounts/main.go` | Accounts service router |
| `backend/golang/cmd/handlers/helpers/main.go` | Helper service router |
| `backend/golang/cmd/handlers/api-keys/main.go` | API key management |
| `backend/golang/internal/connectors/unified_interface.go` | CRM interface |
| `backend/golang/internal/connectors/keap.go` | Keap implementation |
| `backend/golang/internal/helpers/interface.go` | Helper interface |
| `backend/golang/internal/helpers/registry.go` | Helper registry |
| `backend/golang/internal/middleware/auth/auth.go` | JWT auth middleware |
| `backend/golang/internal/middleware/apikey/apikey.go` | API key auth middleware |
| `backend/golang/internal/database/users_repository.go` | Users DB access |
| `backend/golang/internal/database/accounts_repository.go` | Accounts DB access |
| `backend/golang/internal/database/apikeys_repository.go` | API keys DB access |
| `backend/golang/internal/database/helpers_repository.go` | Helpers DB access |

### Frontend (Next.js)
| Path | Purpose |
|------|---------|
| `apps/web/app/(dashboard)/helpers/page.tsx` | Helper library page |
| `apps/web/app/(dashboard)/connections/page.tsx` | CRM connections |
| `apps/web/components/helpers/HelperConfigForm.tsx` | Helper config UI |
| `apps/web/lib/api/helpers.ts` | Helpers API client |
| `apps/web/lib/stores/helper-store.ts` | Helper state |

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
- [ ] At least 4 CRM platforms supported (Keap, GoHighLevel, ActiveCampaign, Ontraport)
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
