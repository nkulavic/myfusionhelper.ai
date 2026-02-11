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
  - Custom domain (prod): `https://api.myfusionhelper.ai`
- **Route53 Hosted Zone**: `myfusionhelper.ai` (ID: Z071462818IPQJBH38AMK)
- **ACM Certificate**: Managed via `services/infrastructure/acm` (DNS validation)
- **Cognito User Pool**: `us-west-2_1E74cZW97`
- **DynamoDB table prefix**: `mfh-{stage}-` (e.g., `mfh-dev-users`)
- **S3 data bucket**: `mfh-{stage}-data`
- **Stripe SSM params**: `/{stage}/stripe/secret_key`, `/{stage}/stripe/webhook_secret`, `/{stage}/stripe/price_start`, `/{stage}/stripe/price_grow`, `/{stage}/stripe/price_deliver`

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

## Backend Deploy Order

Infrastructure must deploy before API services. The CI pipeline enforces this:

1. **Build & test** Go code
2. **Infrastructure** (parallel): cognito, dynamodb-core, s3, sqs, ses, monitoring, acm (ACM certificate for custom domain)
3. **Pre-gateway services**: api-key-authorizer, scheduler
4. **API Gateway** (creates HttpApi + Cognito authorizer + custom domain mapping)
5. **Route53** (creates DNS records pointing to API Gateway)
6. **API services** (parallel, max 3): auth, accounts, api-keys, helpers, platforms, data-explorer, billing, chat
7. **Workers** (parallel): helper-worker, notification-worker, data-sync, executions-stream, sms-chat-webhook, alexa-webhook, google-assistant-webhook
8. **Post-deploy**: seed platform data + health check verification

### Local Backend Deploy

```bash
cd backend/golang
npm install                    # installs serverless-go-plugin + glob
cd services/api/auth
npx sls deploy --stage dev     # deploys single service
```

## Supported CRM Platforms

| Platform | Auth Type | Status |
|----------|-----------|--------|
| Keap (Infusionsoft) | OAuth 2.0 | Active |
| GoHighLevel | API Key | Active |
| ActiveCampaign | API Key | Active |
| Ontraport | API Key | Active |
| HubSpot | API Key (Private App) | Active |
| Stripe | API Key (Secret Key) | Active |

## Helper Categories

Helpers are organized into categories: contact, data, tagging, automation, integration, notification, analytics. Each helper implements the `helpers.Helper` Go interface and is registered in a global registry via `helpers.Register()`.

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
