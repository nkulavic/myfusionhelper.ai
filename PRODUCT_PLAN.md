# MyFusionHelper.ai -- Product Plan & Executive Summary

**Last updated**: February 2026
**Status**: Active development (pre-launch)

---

## 1. Vision

MyFusionHelper.ai is an AI-powered CRM automation platform that gives small business owners, marketers, and entrepreneurs the power to extend their CRM far beyond its native capabilities -- without writing code.

The product evolved from **MyFusion Helper** (myfusionhelper.com), a WordPress/PHP-based system with 90+ helper microservices built exclusively for Keap/Infusionsoft. The new platform is a ground-up rebuild with three strategic upgrades:

1. **Multi-CRM support**: One platform works across Keap, GoHighLevel, ActiveCampaign, Ontraport, and HubSpot -- with more to come.
2. **Modern architecture**: Go backend on AWS Lambda, Next.js frontend on Vercel, DynamoDB, SQS queues. Fast, scalable, and cheap to operate.
3. **AI-first features**: Natural language automation building, AI-powered business intelligence, email composition, and a conversational CRM assistant.

**Why it exists**: CRM platforms are powerful but rigid. Users constantly need to do things their CRM cannot do natively -- format data, calculate dates, sync to Google Sheets, apply conditional tags, split-test, score leads, chain automations together. MyFusionHelper fills that gap with pre-built, configurable "helpers" that plug into CRM workflows via webhooks. The AI layer makes it even easier: describe what you want in plain English, and the system builds the automation for you.

---

## 2. Target Users

**Primary**: Small business owners and solo entrepreneurs using a CRM (especially Keap) who need automation beyond what the CRM provides natively. Non-technical. Value time savings over technical sophistication.

**Secondary**: Marketing consultants and agencies managing multiple client CRM accounts. They need multi-CRM support and team features.

**Tertiary**: Technical marketers and developers who use the API/webhook system to build custom integrations.

**User profile**:
- Runs a small business or consultancy (1-20 employees)
- Uses Keap, GoHighLevel, ActiveCampaign, Ontraport, or HubSpot as their CRM
- Understands CRM concepts (contacts, tags, automations, custom fields) but is not a developer
- Willing to pay $39-79/month for tools that save hours of manual work each week
- Existing legacy user base: ~1,200+ users on the current myfusionhelper.com platform

---

## 3. Core Value Proposition

**"58+ automation helpers that extend your CRM, now enhanced with AI."**

Key differentiators:
- **No code required**: Every helper is configured through a UI, not code
- **Multi-CRM**: Same helpers work across 5+ CRM platforms through a unified connector layer
- **AI-powered**: Natural language automation building, AI insights, email composition
- **Webhook-triggered**: Helpers fire from CRM automations in real-time via API
- **Battle-tested**: Based on a system that has served 1,200+ users and millions of executions

---

## 4. Product Structure

### 4.1 Helpers (Core Product)

Helpers are the fundamental unit of the product. Each helper is a configurable automation action that a user sets up once, then triggers repeatedly from their CRM via webhook.

**7 categories, 58+ helpers implemented in Go:**

| Category | Count | Examples |
|----------|-------|---------|
| Contact | 16 | Copy It, Merge It, Found It, Name Parse It, Assign It, Clear It, Own It, Snapshot It |
| Data | 16 | Format It, Math It, Date Calculator, Text It, Split It (A/B), When Is It, Password It |
| Tagging | 6 | Tag It, Clear Tags, Score It, Group It, Count Tags, Count It Tags |
| Automation | 7 | Trigger It, Goal It, Chain It, Action It, Drip It, Stage It, Timezone Triggers |
| Integration | 9 | Google Sheet It, Slack It, Hook It (webhooks), Excel It, Zoom Webinar, Calendly, Twilio SMS, Email Validate, Phone Lookup |
| Notification | 3 | Notify Me, Twilio SMS, Email Engagement |
| Analytics | 3 | RFM Calculation, Customer Lifetime Value, IP Location |

**How helpers work**:
1. User configures a helper in the web UI (selects fields, sets rules, picks tags)
2. System generates a webhook URL for that helper
3. User places the webhook URL into a CRM automation/campaign
4. When the CRM automation fires, it POSTs contact data to the webhook
5. The helper executes (via SQS queue -> Lambda worker), performs the action on the CRM, and logs the result

### 4.2 Connections (Integration Layer)

Connections are OAuth/API-key authenticated links between MyFusionHelper and a user's CRM account.

**5 supported CRM platforms**:
- **Keap** (OAuth2) -- primary, most mature
- **GoHighLevel** (OAuth2) -- high demand
- **ActiveCampaign** (API key)
- **Ontraport** (API key)
- **HubSpot** (OAuth2) -- includes platform-specific helpers (Deal Stager, List Sync, Property Mapper, Workflow Trigger)

Each CRM is implemented behind a **unified connector interface** (`CRMConnector`) with a translation layer that normalizes contacts, tags, custom fields, and automations across platforms. This means helpers written once work across all CRMs.

### 4.3 AI Features

AI capabilities powered by Groq (fast, cost-effective inference):

- **AI Chat Assistant**: Floating chat panel in the dashboard for quick lookups, natural language queries, and guided help
- **AI Insights**: Proactive recommendations on the Insights page -- anomaly detection, optimization suggestions, trend analysis
- **AI Email Composer**: AI-assisted email writing with CRM personalization tokens (planned)
- **AI Reports**: Natural language report building -- "Show me contacts tagged 'Hot Lead' in the last 30 days" (planned)
- **Natural Language Automation Builder**: Describe what you want and the AI generates the helper configuration (planned)

### 4.4 Data Explorer

A DuckDB-powered analytics engine that syncs CRM data to Parquet files on S3 and lets users query their data with SQL or natural language. Includes:
- Parquet writer using Apache Arrow v17
- DuckDB query engine
- Data sync workers
- Catalog, export, and record endpoints

### 4.5 Dashboard

The main dashboard (`/`) provides:
- **Stats grid**: Active helpers, recent executions, success rate, active connections
- **Recent executions**: Live feed of the latest helper runs with status, timing, and contact details
- **Quick actions**: Create helper, add connection, view executions, browse helpers
- **Connection health**: Real-time status of all CRM connections

### 4.6 Additional Pages

| Page | Path | Purpose |
|------|------|---------|
| Helpers Library | `/helpers` | Browse, search, and configure helpers by category |
| Connections | `/connections` | Manage CRM connections, OAuth flow, connection health |
| Executions | `/executions` | Execution history with filtering, stats, and detail view |
| Insights | `/insights` | AI-powered analytics and recommendations |
| Data Explorer | `/data-explorer` | SQL/natural language queries on synced CRM data |
| Reports | `/reports` | Scheduled and on-demand reports |
| Emails | `/emails` | AI-assisted email composition |
| Settings | `/settings` | Profile, Account, Team, API Keys, Billing, Notifications |

---

## 5. Pricing

Three tiers, all with a 14-day free trial. Monthly and annual billing (annual saves ~17%).

| | Start | Grow | Deliver |
|---|---|---|---|
| **Monthly** | $39/mo | $59/mo | $79/mo |
| **Annual** | $33/mo | $49/mo | $66/mo |
| **Active Helpers** | 10 | 50 | Unlimited |
| **Executions/mo** | 5,000 | 50,000 | Unlimited |
| **CRM Connections** | 1 | 3 | Unlimited |
| **AI Features** | -- | AI insights & email composer | AI insights & email composer |
| **Reports** | -- | Scheduled reports | Scheduled reports |
| **Execution Logs** | 7 days | 30 days | 90 days |
| **Support** | Email | Priority | Phone & priority |

**Billing integration**: Stripe, tied to Account (not User). Subscription management, usage metering for helper executions, and plan limits enforcement.

---

## 6. Technical Architecture

### 6.1 Frontend
- **Framework**: Next.js 15 + React 19
- **Styling**: Tailwind CSS + shadcn/ui component library
- **State**: Zustand stores (auth, workspace, UI)
- **Data fetching**: React Query with API client
- **Auth**: AWS Cognito (migrating from Better Auth placeholder)
- **Hosting**: Vercel
- **Domains**: `app.myfusionhelper.ai` (app), `myfusionhelper.ai` (marketing, planned)

### 6.2 Backend
- **Language**: Go
- **Compute**: AWS Lambda (32 handlers across 9 services)
- **Database**: DynamoDB (8 tables with GSIs)
- **Queue**: SQS FIFO for helper execution
- **Storage**: S3 for Parquet files, exports, uploads
- **Auth**: Cognito User Pool (JWT) + API keys (`mfh_live_xxxx`)
- **Infrastructure-as-Code**: Serverless Framework (13 serverless.yml configs)
- **API Gateway**: HTTP API (v2) at `api.myfusionhelper.ai`
- **AI**: Groq for inference

### 6.3 Monorepo Structure
```
myfusionhelper.ai/
├── apps/
│   ├── web/              # Next.js app (app.myfusionhelper.ai)
│   └── marketing/        # Marketing site (planned)
├── backend/
│   └── golang/
│       ├── cmd/handlers/  # 9 Lambda handler services
│       ├── internal/      # Shared Go packages
│       │   ├── connectors/  # 5 CRM connector implementations
│       │   ├── database/    # DynamoDB repository layer
│       │   ├── helpers/     # 58+ helper implementations
│       │   ├── middleware/  # Auth, logging
│       │   ├── services/    # Business logic
│       │   └── types/       # Shared types
│       └── services/      # Serverless Framework configs
│           ├── api/         # 7 API services
│           ├── infrastructure/  # 4 infra services
│           └── workers/     # 2 worker services
├── packages/
│   ├── ui/               # Shared UI components
│   ├── types/            # Shared TypeScript types
│   └── config/           # Shared configs
└── docs/                 # Documentation
```

### 6.4 AWS Resources (Account 570331155915, us-west-2)
- Cognito User Pool: `us-west-2_1E74cZW97`
- API Gateway: `https://a95gb181u4.execute-api.us-west-2.amazonaws.com`
- 8 DynamoDB tables (users, accounts, user-accounts, api-keys, connections, helpers, executions, oauth-states)
- S3 buckets for uploads and exports
- SQS FIFO queues for helper execution and webhooks
- CI/CD via GitHub Actions with OIDC auth

---

## 7. Current State (as of February 2026)

### What Is Built

**Backend (Go)**:
- 32 Lambda handlers across 9 services (auth, accounts, api-keys, helpers, platforms, data-explorer, workers)
- 5 CRM connectors fully implemented (Keap, GoHighLevel, ActiveCampaign, Ontraport, HubSpot) with unified interface and translation layer
- 58+ helper implementations across 7 categories with real execution logic
- Data Explorer with Parquet writer, DuckDB query engine, and sync workers
- All 13 serverless.yml infrastructure configs
- DynamoDB repository layer (client + all repositories)
- Helper execution engine with registry pattern

**Frontend (Next.js)**:
- All dashboard pages built with React Query hooks hitting real API endpoints
- Dashboard with live stats, recent executions, connection health
- Helpers library with category browsing and search
- Connections management page
- Executions history with filtering
- Insights page (AI analytics)
- Data Explorer page
- Reports page
- Emails page
- Settings page (Profile, Account, Team, API Keys, Billing, Notifications tabs)
- Landing page with hero, features, pricing, how-it-works sections
- API client with auth headers, error handling, snake/camelCase transforms
- Dark mode support
- AI chat panel (floating assistant)

**Infrastructure**:
- Monorepo with Turborepo
- Vercel deployment configured
- GitHub Actions CI/CD pipeline
- All Serverless Framework configs for AWS resources

### What Is Broken or Missing

1. **Auth is broken** (P0): Login/register still wired to Better Auth (placeholder), not the Go backend + Cognito. The app is unusable until this is fixed.
2. **Cognito auth migration**: Need to replace Better Auth with AWS Amplify Auth client connecting to Cognito
3. **Settings backend**: Billing (Stripe), Team management, and Notifications tabs use mock data
4. **Zero Go tests**: No unit or integration tests for backend
5. **Marketing site**: `apps/marketing/` not started (landing components exist in `apps/web`)
6. **Onboarding flow**: No guided first-run experience for new users
7. **Helper builder UI**: No visual configuration interface for creating/editing helpers (`/helpers/new`, `/helpers/[id]`)
8. **Dead packages**: better-auth, @neondatabase/serverless, drizzle-orm still in frontend deps

---

## 8. Roadmap & Priorities

### P0 -- Launch Blockers (Must complete before any user can use the app)

1. **Cognito Auth Migration**: Replace Better Auth with Cognito. Install aws-amplify, rewrite auth client/store/hooks, update login/register pages, update middleware, update API client token flow. ~12 files changed, ~4 deleted.

2. **Infrastructure Deployment**: Deploy all AWS resources in order: Cognito -> DynamoDB -> S3 -> SQS -> API Gateway -> API services -> Workers -> Seed platform data. Configure custom domains.

3. **Onboarding Flow**: First-run experience that guides new users through: connect CRM -> browse helpers -> configure first helper -> test execution. Critical for activation and retention.

### P1 -- Core Product Completeness

4. **Helper Builder UI**: Visual configuration interface at `/helpers/new` and `/helpers/[id]` for creating and editing helpers. JSON Schema-driven dynamic forms based on helper type. This is how users actually use the product.

5. **Settings Backend Integration**: Wire up Profile save, Team management, Notification preferences, and Stripe billing to real API endpoints.

6. **Go Test Suite**: Unit tests for repositories, handler integration tests, helper unit tests with mock connectors, connector tests with mock HTTP. ~15-20 test files.

### P2 -- Growth & Polish

7. **AI Features Activation**: Connect AI chat, insights, and email composer to Groq backend. Natural language automation builder.

8. **Data Explorer Polish**: Improve the DuckDB query UX, add saved queries, scheduled data syncs.

9. **Marketing Site**: Move landing page components from `apps/web` to `apps/marketing`, deploy to root domain.

10. **Mobile Responsive**: Ensure all pages work well on mobile devices.

### P3 -- Scale

11. **Legacy User Migration**: Migration flow from myfusionhelper.com -- import existing helpers, re-auth OAuth, verify helpers work.

12. **Multi-User & Teams**: Team invitations, role-based access (owner, admin, member, viewer), account switching.

13. **Advanced Billing**: Usage-based billing, overage handling, plan upgrades/downgrades via Stripe portal.

14. **Additional CRM Platforms**: Generic HTTP POST connector, Salesforce, Mailchimp, and any platform with webhook functionality.

---

## 9. Success Metrics

### Launch Metrics (First 90 days)
- App is functional end-to-end: register -> connect CRM -> configure helper -> execute helper -> see results
- Auth works reliably (Cognito, no auth-related support tickets)
- First 50 users successfully onboarded
- <500ms average helper execution time
- 99.9% uptime
- Zero data-loss incidents

### Growth Metrics (6 months)
- 200+ active paying users
- 80%+ of legacy myfusionhelper.com users migrated
- Average user has 5+ active helpers
- <5% monthly churn rate
- At least 3 CRM platforms with active users (Keap + 2 others)
- AI features used by 30%+ of active users

### Product Health Metrics
- <3 second page load times
- <100ms Lambda cold start (consolidated handlers)
- AI chat responds in <2 seconds
- Email composer generates drafts in <3 seconds
- Execution success rate >99%
- Net Promoter Score >40

### Revenue Targets
- MRR of $10K within 6 months of launch
- Average revenue per user (ARPU) of $50+/month
- Annual plan adoption >40% (improves cash flow and reduces churn)

---

## Appendix: Key Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Authentication | AWS Cognito | Free tier, integrates with AWS ecosystem |
| Database | DynamoDB | Serverless, pay-per-request, no ops overhead |
| Backend Language | Go | Fast cold starts on Lambda, strong typing, team experience from listbackup.ai |
| Frontend Framework | Next.js 15 | React ecosystem, Vercel hosting, SSR/ISR |
| Infrastructure-as-Code | Serverless Framework | Team familiarity, rapid iteration |
| AI Provider | Groq | Cost-effective, fast inference |
| CRM Abstraction | Unified Connector Interface | Write helpers once, support all CRMs |
| Helper Execution | SQS FIFO + Lambda Worker | Reliable, ordered execution with dead letter queue |
| UI Library | shadcn/ui + Tailwind | Beautiful defaults, fully customizable, no runtime CSS |
| Monorepo | Turborepo | Fast builds, good Next.js integration |
| Resource Naming | `mfh-{stage}-*` | Consistent, stage-aware naming across all AWS resources |
