# MyFusion Helper -- API Reference

Base URL: `https://a95gb181u4.execute-api.us-west-2.amazonaws.com`

All responses follow the standard envelope:

```json
// Success
{ "success": true, "message": "OK", "data": { ... } }

// Error
{ "success": false, "error": "Error message" }
```

All responses include `Access-Control-Allow-Origin: *` CORS header.

---

## Authentication

JWT tokens are issued by AWS Cognito. Protected endpoints require an `Authorization: Bearer <token>` header. The auth middleware extracts the user's `sub` claim, constructs `user:<sub>` as the internal user ID, and resolves the user's current account and permissions from DynamoDB.

---

## 1. Auth Service (`mfh-auth`)

### POST /auth/register

Create a new user account.

**Auth**: None

**Request**:
```json
{
  "email": "user@example.com",     // required, must contain @
  "password": "SecurePass1!",      // required, min 8 chars
  "name": "Jane Doe",             // required
  "phone_number": "+15551234567", // optional, must start with +
  "company": "Acme Inc"           // optional
}
```

**Response** (200):
```json
{
  "token": "<access_token>",
  "refresh_token": "<refresh_token>",
  "user": {
    "user_id": "user:<uuid>",
    "email": "user@example.com",
    "name": "Jane Doe",
    "status": "active",
    "current_account_id": "account:<uuid>",
    "created_at": "2026-02-05T00:00:00Z",
    "updated_at": "2026-02-05T00:00:00Z"
  },
  "account": {
    "account_id": "account:<uuid>",
    "name": "Acme Inc",
    "plan": "free",
    "status": "active",
    "created_at": "2026-02-05T00:00:00Z"
  }
}
```

**Errors**: `400` invalid input, `409` email already registered

**Notes**: Auto-creates a default account with `free` plan and `Owner` role. Sends a welcome email via SES.

---

### POST /auth/login

Authenticate with email and password.

**Auth**: None

**Request**:
```json
{
  "email": "user@example.com",  // required
  "password": "SecurePass1!"    // required
}
```

**Response** (200):
```json
{
  "token": "<access_token>",
  "refresh_token": "<refresh_token>",
  "user": { ... },
  "account": { ... }
}
```

**Errors**: `401` invalid credentials or unverified email, `429` too many attempts

---

### POST /auth/refresh

Exchange a refresh token for a new access token.

**Auth**: None

**Request**:
```json
{
  "refresh_token": "<refresh_token>"  // required
}
```

**Response** (200):
```json
{
  "token": "<new_access_token>",
  "token_type": "Bearer",
  "refresh_token": "<new_refresh_token>"  // only if Cognito returns one
}
```

**Errors**: `401` invalid or expired refresh token

---

### GET /auth/status

Get the authenticated user's profile and account context.

**Auth**: JWT required

**Response** (200):
```json
{
  "user_id": "user:<uuid>",
  "email": "user@example.com",
  "name": "Jane Doe",
  "status": "active",
  "current_account_id": "account:<uuid>",
  "created_at": "...",
  "updated_at": "...",
  "account_context": {
    "account_id": "account:<uuid>",
    "role": "Owner",
    "permissions": { ... }
  },
  "available_accounts": [
    {
      "account_id": "account:<uuid>",
      "account_name": "Acme Inc",
      "role": "Owner",
      "permissions": { ... },
      "is_current": true
    }
  ]
}
```

---

### POST /auth/logout

Sign out the user globally (invalidates all sessions).

**Auth**: JWT required

**Response** (200):
```json
{
  "user_id": "user:<uuid>"
}
```

---

### PUT /auth/profile

Update the user's name and/or email.

**Auth**: JWT required

**Request**:
```json
{
  "name": "New Name",       // optional
  "email": "new@email.com"  // optional, updates both Cognito and DynamoDB
}
```

**Response** (200):
```json
{
  "user_id": "user:<uuid>",
  "name": "New Name",
  "email": "new@email.com",
  "updated_at": "..."
}
```

**Errors**: `400` no fields provided or invalid email format

**Notes**: At least one field is required. Email changes are applied to Cognito first.

---

### PUT /auth/password

Change the authenticated user's password.

**Auth**: JWT required

**Request**:
```json
{
  "current_password": "OldPass1!",  // required
  "new_password": "NewPass2!"       // required, min 8 chars, upper+lower+digit
}
```

**Response** (200): `null` data

**Errors**: `401` current password incorrect, `400` password policy violation, `429` rate limited

---

### POST /auth/forgot-password

Initiate password reset (sends verification code via Cognito email).

**Auth**: None

**Request**:
```json
{
  "email": "user@example.com"  // required
}
```

**Response** (200): Always returns success to prevent email enumeration.
```json
null
```

**Errors**: `429` rate limited

---

### POST /auth/reset-password

Confirm password reset with verification code.

**Auth**: None

**Request**:
```json
{
  "email": "user@example.com",  // required
  "code": "123456",             // required (from email)
  "new_password": "NewPass1!"   // required
}
```

**Response** (200): `null` data

**Errors**: `400` invalid/expired code or password policy, `429` rate limited

---

## 2. Accounts Service (`mfh-accounts`)

### GET /accounts

List all accounts the authenticated user has access to.

**Auth**: JWT required

**Response** (200):
```json
{
  "accounts": [
    {
      "account_id": "account:<uuid>",
      "name": "Acme Inc",
      "company": "Acme Inc",
      "plan": "grow",
      "status": "active",
      "user_role": "Owner",
      "user_status": "active",
      "is_current": true,
      "linked_at": "...",
      "updated_at": "..."
    }
  ],
  "total_count": 1,
  "current_account_id": "account:<uuid>"
}
```

---

### GET /accounts/{account_id}

Get a specific account's details.

**Auth**: JWT required (user must have access to the account)

**Response** (200):
```json
{
  "account_id": "account:<uuid>",
  "name": "Acme Inc",
  "company": "Acme Inc",
  "plan": "grow",
  "status": "active",
  "settings": { "max_helpers": 50, "max_connections": 5, "max_executions": 50000, "max_team_members": 10 },
  "usage": { "helpers_count": 5, "connections_count": 2, "executions_count": 150 },
  "created_at": "...",
  "updated_at": "..."
}
```

---

### PUT /accounts/{account_id}

Update an account's name or company.

**Auth**: JWT required

**Request**:
```json
{
  "name": "New Name",      // optional
  "company": "New Co"      // optional
}
```

**Response** (200):
```json
{
  "account_id": "account:<uuid>",
  "name": "New Name",
  "company": "New Co",
  "updated_at": "..."
}
```

---

### POST /accounts/switch

Switch the user's active account.

**Auth**: JWT required

**Request**:
```json
{
  "account_id": "account:<uuid>"  // required
}
```

**Response** (200):
```json
{
  "user_id": "user:<uuid>",
  "current_account_id": "account:<uuid>"
}
```

**Errors**: `403` no access to target account

---

### GET /accounts/preferences

Get notification preferences for the authenticated user.

**Auth**: JWT required

**Response** (200):
```json
{
  "execution_failures": true,
  "connection_issues": true,
  "usage_alerts": true,
  "weekly_summary": false,
  "new_features": true,
  "team_activity": false,
  "realtime_status": false,
  "ai_insights": true,
  "system_maintenance": true
}
```

---

### PUT /accounts/preferences

Update notification preferences.

**Auth**: JWT required

**Request**: Same shape as the GET response (all fields optional; partial update).

**Response** (200): Returns the updated preferences object.

---

### GET /accounts/{account_id}/team

List all team members for an account.

**Auth**: JWT required

**Response** (200):
```json
{
  "members": [
    {
      "user_id": "user:<uuid>",
      "email": "owner@example.com",
      "name": "Jane Doe",
      "role": "Owner",
      "status": "active",
      "linked_at": "..."
    }
  ],
  "total_count": 1
}
```

---

### POST /accounts/{account_id}/team

Invite a team member by email.

**Auth**: JWT required (requires `can_manage_team` permission)

**Request**:
```json
{
  "email": "newmember@example.com",  // required
  "role": "member"                    // optional, default "member". Values: admin, member, viewer
}
```

**Response** (201):
```json
{
  "user_id": "user:<uuid>",
  "email": "newmember@example.com",
  "role": "member",
  "status": "active"
}
```

**Errors**: `403` permission denied or team limit reached, `409` already a member

**Notes**: If the email is not registered, a new Cognito user is created with a temporary password and an invitation email is sent.

---

### PUT /accounts/{account_id}/team/{user_id}

Update a team member's role.

**Auth**: JWT required (requires `can_manage_team` permission)

**Request**:
```json
{
  "role": "admin"  // required. Values: admin, member, viewer
}
```

**Response** (200):
```json
{
  "user_id": "user:<uuid>",
  "role": "admin"
}
```

**Errors**: `403` cannot change account owner's role

---

### DELETE /accounts/{account_id}/team/{user_id}

Remove a team member from the account.

**Auth**: JWT required (requires `can_manage_team` permission)

**Response** (200):
```json
{
  "user_id": "user:<uuid>"
}
```

**Errors**: `400` cannot remove yourself, `403` cannot remove account owner

---

## 3. Platforms & Connections Service (`mfh-platforms`)

### GET /platforms

List all available CRM platforms.

**Auth**: JWT required

**Response** (200): Array of platform objects with `platform_id`, `name`, `slug`, `status`, `api_config`, `oauth`, etc.

---

### GET /platforms/{platform_id}

Get a single platform's details.

**Auth**: JWT required

---

### GET /platform-connections

List all connections for the current account (across all platforms).

**Auth**: JWT required (requires `can_manage_connections` permission)

**Response** (200):
```json
{
  "connections": [
    {
      "connection_id": "connection:<uuid>",
      "platform_id": "platform:<uuid>",
      "account_id": "account:<uuid>",
      "name": "My Keap Connection",
      "status": "active",
      "auth_type": "oauth2",
      "created_at": "...",
      "updated_at": "...",
      "external_user_id": "...",
      "external_user_email": "..."
    }
  ],
  "total": 1
}
```

---

### GET /platforms/{platform_id}/connections

List connections for a specific platform.

**Auth**: JWT required

**Response** (200): Same shape as above, with additional `platform_id` field.

---

### POST /platforms/{platform_id}/connections

Create a new API key-based connection.

**Auth**: JWT required

**Request**:
```json
{
  "name": "My Connection",       // required
  "auth_type": "api_key",        // required: "oauth2" or "api_key"
  "credentials": {                // required for api_key
    "api_key": "...",
    "api_secret": "..."
  }
}
```

**Response** (201):
```json
{
  "connection_id": "connection:<uuid>",
  "platform_id": "platform:<uuid>",
  "name": "My Connection",
  "status": "active",
  "auth_type": "api_key",
  "created_at": "..."
}
```

---

### GET /platforms/{platform_id}/connections/{connection_id}

Get a single connection's details.

**Auth**: JWT required

---

### PUT /platforms/{platform_id}/connections/{connection_id}

Update a connection's name or credentials.

**Auth**: JWT required

**Request**:
```json
{
  "name": "Updated Name",    // optional
  "credentials": {            // optional, updates auth record
    "api_key": "...",
    "api_secret": "..."
  }
}
```

**Response** (200):
```json
{
  "connection_id": "connection:<uuid>",
  "updated_at": "..."
}
```

---

### DELETE /platforms/{platform_id}/connections/{connection_id}

Delete a connection and revoke its auth record.

**Auth**: JWT required

**Response** (200):
```json
{
  "connection_id": "connection:<uuid>",
  "deleted_at": "..."
}
```

---

### POST /platforms/{platform_id}/connections/{connection_id}/test

Test a connection by hitting the platform's test endpoint.

**Auth**: JWT required

**Response** (200):
```json
{
  "connection_id": "connection:<uuid>",
  "platform_id": "platform:<uuid>",
  "status": "success",
  "tested_at": "...",
  "result": { "status": "valid", ... }
}
```

**Errors**: `401` connection test failed (updates connection status to `error`)

---

### POST /platforms/{platform_id}/oauth/start

Initiate an OAuth2 flow for the given platform.

**Auth**: JWT required

**Request** (optional):
```json
{
  "success_redirect": "https://app.myfusionhelper.ai/connections?tab=keap",
  "failure_redirect": "https://app.myfusionhelper.ai/connections?error=1"
}
```

**Response** (200):
```json
{
  "authorization_url": "https://accounts.infusionsoft.com/app/oauth/authorize?...",
  "state": "state:<uuid>",
  "platform_id": "platform:<uuid>",
  "platform_name": "Keap",
  "expires_in": 900
}
```

**Notes**: The frontend should redirect the user to `authorization_url`. OAuth credentials are loaded from SSM at `/{stage}/platforms/{slug}/oauth/client_id` and `client_secret`.

---

### GET /platforms/oauth/callback

OAuth callback endpoint (called by the CRM provider, not the frontend directly).

**Auth**: None (verified via state token)

**Response**: 302 redirect to the frontend with `?oauth=success&connection_id=...` or `?oauth=error&error=...`.

---

## 4. Helpers Service (`mfh-helpers`)

### GET /helpers

List all helpers for the current account.

**Auth**: JWT required

**Response** (200):
```json
{
  "helpers": [
    {
      "helper_id": "helper:<uuid>",
      "name": "Tag New Leads",
      "description": "Applies the 'new-lead' tag to contacts",
      "helper_type": "tag_it",
      "category": "contact",
      "status": "active",
      "enabled": true,
      "execution_count": 42,
      "last_executed_at": "...",
      "created_at": "...",
      "updated_at": "..."
    }
  ],
  "total_count": 1
}
```

---

### POST /helpers

Create a new helper.

**Auth**: JWT required (requires `can_manage_helpers` permission)

**Request**:
```json
{
  "name": "Tag New Leads",           // required
  "description": "...",              // optional
  "helper_type": "tag_it",           // required, must be a registered type
  "category": "contact",             // optional, auto-populated from registry
  "connection_id": "connection:...", // optional
  "config": { "tag_name": "new-lead" } // optional, validated against helper schema
}
```

**Response** (201):
```json
{
  "helper_id": "helper:<uuid>",
  "name": "Tag New Leads",
  "helper_type": "tag_it",
  "category": "contact",
  "created_at": "..."
}
```

**Errors**: `400` unknown helper type or invalid config, `403` permission denied

---

### GET /helpers/{helper_id}

Get a single helper with full config details.

**Auth**: JWT required

**Response** (200): Same fields as list, plus `config`, `config_schema`, `connection_id`.

---

### PUT /helpers/{helper_id}

Update a helper's name, description, config, enabled state, or connection.

**Auth**: JWT required (requires `can_manage_helpers`)

**Request**:
```json
{
  "name": "Updated Name",     // optional
  "description": "...",       // optional
  "config": { ... },          // optional, validated against schema
  "enabled": false,           // optional
  "connection_id": "..."      // optional
}
```

**Response** (200):
```json
{
  "helper_id": "helper:<uuid>"
}
```

---

### DELETE /helpers/{helper_id}

Soft-delete a helper (sets status to `deleted`, enabled to `false`).

**Auth**: JWT required (requires `can_manage_helpers`)

**Response** (200):
```json
{
  "helper_id": "helper:<uuid>"
}
```

---

### POST /helpers/{helper_id}/execute

Queue a helper for async execution via SQS.

**Auth**: JWT required (requires `can_execute_helpers`)

**Request**:
```json
{
  "contact_id": "12345",        // optional
  "input": { "key": "value" }   // optional
}
```

**Response** (202):
```json
{
  "execution_id": "exec:<uuid>",
  "helper_id": "helper:<uuid>",
  "status": "queued",
  "started_at": "..."
}
```

**Errors**: `400` helper is disabled, `403` permission denied

---

### GET /helpers/types

List all registered helper types with their schemas.

**Auth**: JWT required

**Response** (200):
```json
{
  "types": [
    {
      "type": "tag_it",
      "name": "Tag It",
      "category": "contact",
      "description": "Apply or remove tags on a contact",
      "requires_crm": true,
      "supported_crms": ["keap", "gohighlevel", "activecampaign", "ontraport"],
      "config_schema": { ... }
    }
  ],
  "total_count": 47,
  "categories": ["analytics", "automation", "contact", "data", "integration", "notification", "tagging"]
}
```

---

### GET /helpers/types/{type}

Get a single helper type's details and config schema.

**Auth**: JWT required

**Response** (200): Single type object (same shape as items in the list above).

---

## 5. Executions (part of `mfh-helpers`)

### GET /executions

List executions for the current account with pagination and filtering.

**Auth**: JWT required

**Query Parameters**:
| Param | Type | Default | Description |
|-------|------|---------|-------------|
| `helper_id` | string | -- | Filter by helper |
| `status` | string | -- | Filter by status (pending, queued, running, completed, failed) |
| `limit` | int | 20 | Page size (max 100) |
| `next_token` | string | -- | Cursor for next page (base64-encoded) |

**Response** (200):
```json
{
  "executions": [
    {
      "execution_id": "exec:<uuid>",
      "helper_id": "helper:<uuid>",
      "account_id": "account:<uuid>",
      "user_id": "user:<uuid>",
      "connection_id": "connection:<uuid>",
      "contact_id": "12345",
      "status": "completed",
      "trigger_type": "manual",
      "error_message": "",
      "duration_ms": 1250,
      "created_at": "...",
      "started_at": "...",
      "completed_at": "..."
    }
  ],
  "total_count": 1,
  "next_token": "eyJl...",
  "has_more": false
}
```

---

### GET /executions/{execution_id}

Get a single execution with full input/output data.

**Auth**: JWT required

**Response** (200): Same fields as list, plus `input` and `output` objects.

---

## 6. Billing Service (`mfh-billing`)

### GET /billing

Get billing info and subscription details for the current account.

**Auth**: JWT required

**Response** (200):
```json
{
  "plan": "grow",
  "status": "active",
  "price_monthly": 59,
  "renews_at": 1709251200,
  "trial_ends_at": null,
  "cancel_at": null,
  "stripe_customer_id": "cus_...",
  "usage": { "helpers_count": 5, "connections_count": 2, "executions_count": 150 },
  "limits": { "max_helpers": 50, "max_connections": 5, "max_executions": 50000 }
}
```

**Notes**: If a Stripe customer exists, enriches with live subscription data (renewal date, trial end, cancellation).

---

### POST /billing/checkout/sessions

Create a Stripe Checkout session for a new subscription.

**Auth**: JWT required

**Request**:
```json
{
  "plan": "start"  // required: "start", "grow", or "deliver"
}
```

**Response** (200):
```json
{
  "url": "https://checkout.stripe.com/c/pay/...",
  "session_id": "cs_..."
}
```

**Notes**: Creates a Stripe customer if one doesn't exist. Includes a 14-day trial. Redirects to `/settings?tab=billing&session_id={CHECKOUT_SESSION_ID}` on success.

**Errors**: `400` invalid plan, `503` billing not configured

---

### POST /billing/portal-session

Create a Stripe Customer Portal session for managing subscriptions.

**Auth**: JWT required

**Response** (200):
```json
{
  "url": "https://billing.stripe.com/p/session/..."
}
```

**Errors**: `400` no Stripe customer (must subscribe first), `503` billing not configured

---

### GET /billing/invoices

List up to 24 recent invoices from Stripe.

**Auth**: JWT required

**Response** (200):
```json
[
  {
    "id": "in_...",
    "amount": 59,
    "currency": "usd",
    "status": "paid",
    "date": 1706745600,
    "pdf_url": "https://...",
    "hosted_url": "https://..."
  }
]
```

**Notes**: Returns empty array if no Stripe customer exists.

---

### POST /billing/webhook

Stripe webhook receiver. Verified by Stripe signature, not JWT.

**Auth**: None (verified by `Stripe-Signature` header)

**Handled Events**:
| Event | Action |
|-------|--------|
| `checkout.session.completed` | Activates subscription, updates plan and limits |
| `customer.subscription.created` | Updates account plan and status |
| `customer.subscription.updated` | Updates account plan and status |
| `customer.subscription.deleted` | Downgrades to free plan |
| `invoice.payment_failed` | Logged only (Stripe handles retry) |

**Plan Limits**:
| Plan | Helpers | Connections | Executions |
|------|---------|-------------|------------|
| free | 3 | 1 | 1,000 |
| start | 10 | 2 | 10,000 |
| grow | 50 | 5 | 50,000 |
| deliver | unlimited | 20 | 500,000 |

**Response**: Always returns `200 OK` to Stripe.

---

## 7. API Keys Service (`mfh-api-keys`)

Managed through the accounts service. Endpoints follow standard CRUD patterns at `/api-keys`.

---

## Common Error Codes

| Code | Meaning |
|------|---------|
| 400 | Bad request (validation error) |
| 401 | Unauthorized (missing/invalid token, wrong password) |
| 403 | Forbidden (insufficient permissions) |
| 404 | Not found (or hidden due to account isolation) |
| 405 | Method not allowed |
| 409 | Conflict (duplicate resource) |
| 429 | Rate limited |
| 500 | Internal server error |
| 503 | Service not configured (missing Stripe keys, etc.) |

## Permissions

Each user-account relationship has a `permissions` object:

```json
{
  "can_manage_helpers": true,
  "can_execute_helpers": true,
  "can_manage_connections": true,
  "can_manage_team": true,
  "can_manage_billing": true,
  "can_view_analytics": true,
  "can_manage_api_keys": true
}
```

Role defaults:
- **Owner/Admin**: all permissions
- **Member**: manage + execute helpers, manage connections, view analytics
- **Viewer**: view analytics only

---

## 8. Chat Service (`mfh-chat`)

AI-powered chat interface with MCP (Model Context Protocol) tool calling for querying CRM data and invoking helpers.

### GET /chat/health

Health check endpoint.

**Auth**: None

**Response** (200):
```json
{
  "status": "healthy",
  "service": "chat",
  "version": "2025.02.09.0001"
}
```

---

### POST /chat/conversations

Create a new conversation.

**Auth**: Required

**Request**:
```json
{
  "title": "My Conversation"  // optional, auto-generated if not provided
}
```

**Response** (200):
```json
{
  "conversation_id": "conv:uuid",
  "account_id": "account:uuid",
  "title": "My Conversation",
  "created_at": "2026-02-09T00:00:00Z",
  "updated_at": "2026-02-09T00:00:00Z"
}
```

---

### GET /chat/conversations

List all conversations for the current user.

**Auth**: Required

**Query Parameters**:
- `limit` (optional, default: 50): Max conversations to return
- `offset` (optional, default: 0): Pagination offset

**Response** (200):
```json
{
  "conversations": [
    {
      "conversation_id": "conv:uuid",
      "title": "My Conversation",
      "created_at": "2026-02-09T00:00:00Z",
      "updated_at": "2026-02-09T00:00:00Z",
      "message_count": 5
    }
  ],
  "total": 10,
  "has_more": false
}
```

---

### GET /chat/conversations/{id}

Get conversation details with message history.

**Auth**: Required

**Response** (200):
```json
{
  "conversation_id": "conv:uuid",
  "title": "My Conversation",
  "created_at": "2026-02-09T00:00:00Z",
  "messages": [
    {
      "message_id": "msg:uuid",
      "role": "user",
      "content": "Show my Keap contacts",
      "created_at": "2026-02-09T00:00:00Z"
    },
    {
      "message_id": "msg:uuid",
      "role": "assistant",
      "content": "You have 142 contacts in Keap...",
      "tool_calls": [
        {
          "id": "call_abc",
          "type": "function",
          "function": {
            "name": "query_crm_data",
            "arguments": "{\"connection_id\":\"conn:xyz\"}"
          }
        }
      ],
      "created_at": "2026-02-09T00:00:01Z"
    }
  ]
}
```

---

### DELETE /chat/conversations/{id}

Delete a conversation and all its messages.

**Auth**: Required

**Response** (200):
```json
{
  "success": true,
  "message": "Conversation deleted"
}
```

---

### POST /chat/conversations/{id}/messages

Send a message and get AI response (with Server-Sent Events streaming).

**Auth**: Required

**Request**:
```json
{
  "content": "Show my Keap contacts",
  "connection_id": "conn:xyz"  // optional, for CRM-specific queries
}
```

**Response** (200, text/event-stream):
```
event: message
data: {"type":"content","delta":"You"}

event: message
data: {"type":"content","delta":" have"}

event: message
data: {"type":"tool_call","tool":"query_crm_data","arguments":"..."}

event: message
data: {"type":"done"}
```

**Notes**:
- Uses Groq LLM (llama-3.3-70b-versatile)
- Supports tool calling for CRM operations
- Available tools: `query_crm_data`, `get_contacts`, `get_contact_detail`, `invoke_helper`, `list_helpers`, `get_helper_config`, `get_connections`
- Conversation history persists in DynamoDB with 90-day TTL

---

### GET /chat/conversations/{id}/messages

Get all messages in a conversation.

**Auth**: Required

**Query Parameters**:
- `limit` (optional, default: 100): Max messages to return
- `offset` (optional, default: 0): Pagination offset

**Response** (200):
```json
{
  "messages": [
    {
      "message_id": "msg:uuid",
      "role": "user",
      "content": "Show my contacts",
      "sequence": 1,
      "created_at": "2026-02-09T00:00:00Z"
    }
  ],
  "total": 5,
  "has_more": false
}
```

---

## 9. Internal Email Service (`mfh-internal-email`)

Internal-only service for sending transactional emails (welcome, password reset, notifications). Not exposed to frontend.

**Base Path**: `/internal/emails`

**Auth**: Service-to-service (no Cognito auth)

### POST /internal/emails/send

Send a transactional email using predefined templates.

**Request**:
```json
{
  "template_type": "welcome",
  "to": "user@example.com",
  "data": {
    "user_name": "John Doe",
    "login_url": "https://app.myfusionhelper.ai/login"
  },
  "account_id": "account:uuid"  // for logging
}
```

**Template Types**:
- `welcome`: New user welcome email
- `password_reset`: Password reset notification
- `execution_alert`: Helper execution result
- `billing_event`: Subscription changes
- `connection_alert`: CRM connection issues
- `usage_alert`: Approaching plan limits
- `weekly_summary`: Weekly activity digest
- `team_invite`: Team member invitation

**Response** (200):
```json
{
  "email_id": "email:uuid",
  "message_id": "<ses-message-id>",
  "status": "sent"
}
```

---

### GET /internal/emails/history

Get email delivery history (admin/debugging only).

**Query Parameters**:
- `account_id` (optional): Filter by account
- `start_date` (optional): ISO date
- `end_date` (optional): ISO date
- `status` (optional): `sent`, `failed`, `bounced`
- `limit` (optional, default: 50)

**Response** (200):
```json
{
  "logs": [
    {
      "email_id": "email:uuid",
      "account_id": "account:uuid",
      "recipient": "user@example.com",
      "template_type": "welcome",
      "status": "sent",
      "message_id": "<ses-message-id>",
      "sent_at": "2026-02-09T00:00:00Z"
    }
  ]
}
```

---

### GET /internal/emails/health

Health check endpoint.

**Response** (200):
```json
{
  "status": "healthy",
  "service": "internal-email",
  "ses_configured": true
}
```

---

## 10. Data Explorer Service (`mfh-data-explorer`)

Query CRM data with natural language or structured filters using DuckDB on Parquet files.

**Note**: Some endpoints return large datasets. Server-side pagination is enforced.

### POST /data/query

Execute a data query with optional natural language processing.

**Auth**: Required

**Request**:
```json
{
  "connection_id": "conn:uuid",
  "nl_query": "show contacts tagged as VIP",  // optional, natural language
  "filters": {                                // optional, structured filters
    "object_type": "contacts",
    "tags": ["VIP"],
    "date_field": "created_at",
    "date_from": "2026-01-01",
    "date_to": "2026-02-09"
  },
  "limit": 100,
  "offset": 0
}
```

**Response** (200):
```json
{
  "results": [
    {
      "id": "12345",
      "email": "john@example.com",
      "first_name": "John",
      "last_name": "Doe",
      "tags": ["VIP", "Customer"],
      "created_at": "2026-01-15T00:00:00Z"
    }
  ],
  "total": 42,
  "query_time_ms": 150,
  "has_more": false
}
```

**Notes**:
- Uses DuckDB for fast Parquet queries (97-99% cost reduction vs Athena)
- Analytics bucket: `listbackup-analytics-{stage}`
- Max limit: 1000 rows per request

---

### GET /data/catalog

Get available data objects for a connection.

**Auth**: Required

**Query Parameters**:
- `connection_id` (required)

**Response** (200):
```json
{
  "objects": [
    {
      "object_type": "contacts",
      "record_count": 1523,
      "last_synced": "2026-02-09T00:00:00Z",
      "fields": ["id", "email", "first_name", "last_name", "tags"]
    },
    {
      "object_type": "tags",
      "record_count": 45,
      "last_synced": "2026-02-09T00:00:00Z",
      "fields": ["id", "name", "category"]
    }
  ]
}
```

---

### POST /data/export

Export query results as CSV/JSON.

**Auth**: Required

**Request**:
```json
{
  "connection_id": "conn:uuid",
  "query": { /* same as /data/query */ },
  "format": "csv"  // or "json"
}
```

**Response** (200):
- Content-Type: `text/csv` or `application/json`
- Content-Disposition: `attachment; filename="export.csv"`

**Notes**: Exports are limited to 10,000 rows max.

---

## 11. SMS Chat Webhook (`mfh-sms-chat-webhook`)

Twilio webhook for two-way SMS chat interface (internal worker, not directly accessible).

**Base Path**: `/sms-webhook`

**Auth**: Twilio signature validation

**Features**:
- Phone number to account mapping
- Rate limiting (10 messages/hour per phone)
- MCP service integration with tool calling
- Conversation history tracking
- SMS response truncation (1600 char limit)

**Note**: This is a worker service called by Twilio, not a user-facing API.

---

## Data Sync & Workers

### Data Sync Worker (`mfh-data-sync`)

SQS-triggered worker that syncs CRM data to S3 Parquet files.

**Trigger**: SQS queue messages from scheduled EventBridge rules

**Process**:
1. Receives sync job from SQS
2. Fetches data from CRM API (Keap, GoHighLevel, etc.)
3. Writes Parquet files to `mfh-{stage}-data` S3 bucket
4. Updates connection sync metadata

**Note**: Not directly accessible via API.

---

### Helper Execution Worker (`mfh-helper-worker`)

SQS-triggered worker that executes helpers asynchronously.

**Trigger**: SQS queue messages from POST /helpers/{id}/execute

**Process**:
1. Receives execution request from SQS
2. Loads helper configuration
3. Executes helper logic via CRM connectors
4. Updates execution status in DynamoDB
5. Sends notification if configured

**Note**: Not directly accessible via API.