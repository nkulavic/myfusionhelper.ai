# MyFusionHelper.ai — Implementation Plan

Based on audits completed Feb 5, 2026. This plan covers all remaining gaps between the current codebase and a production-ready application.

---

## Current State Summary

### What's DONE:
- **Go Backend**: 32 Lambda handlers across 9 services with real business logic (auth, accounts, api-keys, helpers, platforms, data-explorer, workers)
- **CRM Connectors**: 4 fully implemented (Keap, GoHighLevel, ActiveCampaign, Ontraport) + registry + translation layer
- **Helpers**: 58+ implementations across 7 categories with real execution logic
- **Data Explorer**: Parquet writer (Arrow v17) + DuckDB query engine + sync workers
- **Infrastructure**: All 13 serverless.yml configs (Cognito, DynamoDB, S3, SQS, API Gateway)
- **Frontend UI**: All dashboard pages (helpers, connections, executions, insights, data-explorer, settings) with React Query hooks hitting real API endpoints
- **API Client**: Well-structured with auth headers, error handling, snake/camelCase transforms

### What's BROKEN or MISSING:
1. Auth is completely broken — login/register use better-auth (no database), not the Go backend
2. `internal/database/` is empty — no repository layer (handlers do raw DynamoDB inline — functional but messy)
3. Zero Go tests
4. Settings > Billing, Team, Notifications are mock data
5. Dead packages in frontend (better-auth, neon, drizzle)
6. Marketing site not started

---

## Phase A: Cognito Auth Migration (CRITICAL — App Unusable Without This)

**Priority**: P0 — Must be done first
**Estimated scope**: ~12 files changed, ~4 files deleted

### A1. Remove Dead Packages
- Remove from `apps/web/package.json`: `better-auth`, `@neondatabase/serverless`, `drizzle-orm`, `drizzle-kit`
- Delete `apps/web/src/lib/db.ts` (dead code)
- Run `npm install` to clean lock file

### A2. Install Cognito Client
- Add `aws-amplify` (v6) to `apps/web/package.json`
- Only import `@aws-amplify/auth` submodule to keep bundle small

### A3. Replace Auth Configuration
- **Rewrite** `apps/web/src/lib/auth.ts`:
  - Configure Amplify with Cognito User Pool ID, Client ID, region
  - Export configured auth instance
- **Rewrite** `apps/web/src/lib/auth-client.ts`:
  - Export Cognito-based `signIn`, `signUp`, `signOut`, `confirmSignUp`, `fetchAuthSession`, `getCurrentUser`
  - Wrap Amplify methods for consistent error handling

### A4. Update Auth Store
- **Update** `apps/web/src/lib/stores/auth-store.ts`:
  - Remove manual localStorage token management
  - Use Amplify's built-in token storage
  - Add `fetchSession()` method that gets current Cognito session
  - Keep user/account state but source tokens from Amplify

### A5. Update Login Page
- **Update** `apps/web/src/app/(auth)/login/page.tsx`:
  - Replace `signIn.email()` from better-auth with Amplify `signIn({ username, password })`
  - After Cognito sign-in succeeds, call Go backend `POST /auth/login` to get user/account context
  - Replace `signIn.social({ provider: 'google' })` with Amplify `signInWithRedirect({ provider: 'Google' })`
  - Handle Cognito-specific errors (UserNotConfirmedException, NotAuthorizedException, etc.)

### A6. Update Register Page
- **Update** `apps/web/src/app/(auth)/register/page.tsx`:
  - Replace `signUp.email()` with Amplify `signUp({ username, password, options: { userAttributes: { email, name } } })`
  - Add email verification step (Cognito sends code)
  - After confirmation, call Go backend `POST /auth/register` to create user/account records
  - Add confirmation code page/modal

### A7. Update Middleware
- **Update** `apps/web/src/middleware.ts`:
  - Remove `better-auth.session_token` cookie check
  - Check for Cognito tokens (CognitoIdentityServiceProvider cookies or check auth session)
  - Redirect to `/login` if no valid session
  - Note: Next.js middleware runs on edge, Amplify may need server-side session check approach

### A8. Update API Client Token Flow
- **Update** `apps/web/src/lib/api/client.ts`:
  - Instead of reading `auth_token` from localStorage directly, use Amplify `fetchAuthSession()` to get current access token
  - Tokens auto-refresh via Amplify

### A9. Update Auth Hooks
- **Update** `apps/web/src/lib/hooks/use-auth.ts`:
  - `useLogin()`: Call Amplify signIn → then Go backend `/auth/login`
  - `useRegister()`: Call Amplify signUp → confirm → then Go backend `/auth/register`
  - `useLogout()`: Call Amplify signOut → then Go backend `/auth/logout`
  - `useAuthStatus()`: Use Amplify `getCurrentUser()` + Go backend `/auth/status`

### A10. Add Environment Variables
- Add to Vercel and `.env.local`:
  ```
  NEXT_PUBLIC_COGNITO_USER_POOL_ID=us-west-2_XXXXXX
  NEXT_PUBLIC_COGNITO_CLIENT_ID=xxxxxxxxxxxxxxxxx
  NEXT_PUBLIC_AWS_REGION=us-west-2
  NEXT_PUBLIC_API_URL=https://api.myfusionhelper.ai
  ```

### A11. Delete Unused Auth Files
- Delete `apps/web/src/app/api/auth/[...all]/route.ts` (better-auth API route handler)

---

## Phase B: DynamoDB Repository Layer (Code Quality)

**Priority**: P1 — Important for maintainability, not a blocker
**Estimated scope**: ~10 new files

### B1. Base Repository Setup
- **Create** `backend/golang/internal/database/client.go`:
  - DynamoDB client singleton/factory
  - Table name resolver (reads from environment variables)
  - Common helpers: `marshalItem`, `unmarshalItem`, `buildKey`

### B2. Repository Implementations
Create one repository per table with standard CRUD + query methods:

- **Create** `backend/golang/internal/database/users_repository.go`:
  - `GetByID(userID)`, `GetByEmail(email)` (via EmailIndex), `GetByCognitoID(cognitoID)` (via CognitoUserIdIndex)
  - `Create(user)`, `Update(user)`, `UpdateCurrentAccount(userID, accountID)`

- **Create** `backend/golang/internal/database/accounts_repository.go`:
  - `GetByID(accountID)`, `GetByOwner(ownerUserID)` (via OwnerUserIdIndex)
  - `Create(account)`, `Update(account)`

- **Create** `backend/golang/internal/database/user_accounts_repository.go`:
  - `GetByUserAndAccount(userID, accountID)`, `ListByUser(userID)`, `ListByAccount(accountID)` (via AccountIdIndex)
  - `Create(userAccount)`, `Update(userAccount)`, `Delete(userID, accountID)`

- **Create** `backend/golang/internal/database/apikeys_repository.go`:
  - `GetByID(keyID)`, `GetByHash(keyHash)` (via KeyHashIndex), `ListByAccount(accountID)` (via AccountIdIndex)
  - `Create(apiKey)`, `Revoke(keyID)`

- **Create** `backend/golang/internal/database/connections_repository.go`:
  - `GetByID(connectionID)`, `ListByAccount(accountID)` (via AccountIdIndex)
  - `Create(connection)`, `Update(connection)`, `Delete(connectionID)`
  - `UpdateSyncStatus(connectionID, status, counts)`

- **Create** `backend/golang/internal/database/helpers_repository.go`:
  - `GetByID(helperID)`, `ListByAccount(accountID)` (via AccountIdIndex)
  - `Create(helper)`, `Update(helper)`, `SoftDelete(helperID)`
  - `UpdateExecutionStats(helperID, count, lastExecutedAt)`

- **Create** `backend/golang/internal/database/executions_repository.go`:
  - `GetByID(executionID)`, `ListByAccount(accountID, cursor, limit)` (via AccountIdCreatedAtIndex)
  - `ListByHelper(helperID, cursor, limit)` (via HelperIdCreatedAtIndex)
  - `Create(execution)`, `UpdateResult(executionID, status, result, duration)`

- **Create** `backend/golang/internal/database/platforms_repository.go`:
  - `GetByID(platformID)`, `GetBySlug(slug)` (via SlugIndex), `ListAll(filters)`
  - `Create(platform)`, `Update(platform)`

- **Create** `backend/golang/internal/database/oauth_states_repository.go`:
  - `Create(state)`, `GetAndDelete(state)` (atomic get + delete for one-time use)

### B3. Refactor Handlers (Optional, Lower Priority)
- Gradually replace inline DynamoDB calls in handlers with repository calls
- Start with auth service (most critical path), then expand
- This can be done incrementally — no need to refactor all handlers at once

---

## Phase C: Go Tests

**Priority**: P2 — Important for confidence before production deployment
**Estimated scope**: ~15-20 test files

### C1. Repository Tests
- Test each repository against DynamoDB Local (Docker)
- CRUD operations, GSI queries, error handling

### C2. Handler Integration Tests
- Test auth flow: register → login → status → refresh → logout
- Test helper CRUD + execution flow
- Test connection CRUD + OAuth flow
- Use httptest for Lambda event simulation

### C3. Helper Unit Tests
- Test each helper category with mock connector
- Verify config validation, execution logic, error handling
- Focus on the most-used helpers first (tag_it, format_it, copy_it, field_to_field)

### C4. Connector Tests
- Test each connector against mock HTTP server
- Verify request formatting, response parsing, error handling, pagination

---

## Phase D: Settings Backend Integration

**Priority**: P2 — Needed for full product, not MVP blocker
**Estimated scope**: ~6 files modified

### D1. Team Management API
- Backend: Add team endpoints to accounts service (invite, list members, update role, remove)
- Frontend: Replace mock `listTeamMembers`/`inviteTeamMember` in `lib/api/settings.ts` with real API calls

### D2. Notification Preferences
- Backend: Add notification preferences to user/account DynamoDB records
- Frontend: Replace mock `getNotificationPreferences` with real API calls
- Wire up toggle persistence

### D3. Stripe Billing Integration
- Backend: New billing service or add to accounts service
- Stripe Customer creation on account creation
- Subscription management (create, update, cancel)
- Usage metering for helper executions
- Webhook handler for Stripe events
- Frontend: Replace mock `getBillingInfo`/`createPortalSession` with real Stripe integration

### D4. Profile Save
- Fix Settings ProfileTab "Save Changes" button — add onClick handler that calls `PUT /accounts/{id}`

---

## Phase E: Infrastructure Deployment

**Priority**: P1 — Required before any backend works in production
**Estimated scope**: CLI commands, no code changes

### E1. Deploy Infrastructure (in order)
1. `cd backend/golang/services/infrastructure/cognito && sls deploy --stage dev`
2. `cd backend/golang/services/infrastructure/dynamodb/core && sls deploy --stage dev`
3. `cd backend/golang/services/infrastructure/s3 && sls deploy --stage dev`
4. `cd backend/golang/services/infrastructure/sqs && sls deploy --stage dev`

### E2. Deploy API Gateway
5. `cd backend/golang/services/api/gateway && sls deploy --stage dev`

### E3. Deploy API Services
6. `cd backend/golang/services/api/auth && sls deploy --stage dev`
7. `cd backend/golang/services/api/accounts && sls deploy --stage dev`
8. `cd backend/golang/services/api/api-keys && sls deploy --stage dev`
9. `cd backend/golang/services/api/platforms && sls deploy --stage dev`
10. `cd backend/golang/services/api/helpers && sls deploy --stage dev`
11. `cd backend/golang/services/api/data-explorer && sls deploy --stage dev`

### E4. Deploy Workers
12. `cd backend/golang/services/workers/helper-worker && sls deploy --stage dev`
13. `cd backend/golang/services/workers/data-sync && sls deploy --stage dev`

### E5. Seed Platform Data
14. `cd backend/golang/services/api/platforms/ci_cd/seed && ./seed.sh dev`

### E6. Configure Custom Domains
- Set up `api.myfusionhelper.ai` → API Gateway custom domain
- Update Vercel env vars with deployed API URL + Cognito IDs

---

## Phase F: Marketing Site (Low Priority)

**Priority**: P3 — Can launch app without this
- Create `apps/marketing/` Next.js app
- Landing page content already exists as components in `apps/web/src/components/landing/`
- Move landing components to marketing app
- Deploy to `myfusionhelper.ai` (root domain) on Vercel

---

## Implementation Order & Dependencies

```
Phase E (Deploy Infra) ─── must be first, provides Cognito IDs + API URL
       │
       ├──→ Phase A (Cognito Auth Migration) ─── unblocks all frontend usage
       │         │
       │         └──→ Phase D4 (Profile Save) ─── quick fix after auth works
       │
       ├──→ Phase B (Repository Layer) ─── can run in parallel with A
       │         │
       │         └──→ Phase C (Go Tests) ─── depends on repositories
       │
       └──→ Phase D1-D3 (Settings Backend) ─── can start after infra deployed

Phase F (Marketing Site) ─── independent, anytime
```

---

## Recommended Team Structure for Implementation

| Agent | Workstream | Phases |
|-------|-----------|--------|
| **infra-deployer** | AWS infrastructure deployment | E1-E6 |
| **auth-migrator** | Frontend Cognito migration | A1-A11 |
| **backend-repos** | Go DynamoDB repository layer | B1-B3 |
| **test-writer** | Go test suite (after repos done) | C1-C4 |

Infra must deploy first. Then auth-migrator and backend-repos can work in parallel.
