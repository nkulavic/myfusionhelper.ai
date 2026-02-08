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
- **API Gateway**: `https://a95gb181u4.execute-api.us-west-2.amazonaws.com`
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
2. **Infrastructure** (parallel): cognito, dynamodb-core, s3, sqs
3. **API Gateway** (creates HttpApi + Cognito authorizer)
4. **API services** (parallel, max 3): auth, accounts, api-keys, helpers, platforms, data-explorer, billing
5. **Workers** (parallel): helper-worker, data-sync
6. **Post-deploy**: seed platform data + health check verification

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
