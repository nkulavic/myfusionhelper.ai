# Infrastructure & CI/CD Setup - Implementation Summary

**Task ID**: 44509051
**Status**: Code Changes Complete - Manual Steps Required
**Date**: 2026-02-08

## ✅ Completed Steps

### 1. Updated GitHub Actions Workflow
**File**: `.github/workflows/sync-internal-secrets.yml`

**Changes**:
- Changed step name from "Build and Upload Stripe Secrets" to "Build and Upload Unified Internal Secrets"
- Added Groq secrets environment variables (lines 61-63):
  - `DEV_INTERNAL_GROQ_API_KEY`
  - `STAGING_INTERNAL_GROQ_API_KEY`
  - `MAIN_INTERNAL_GROQ_API_KEY`
- Added Twilio secrets environment variables (lines 65-76):
  - `DEV_INTERNAL_TWILIO_ACCOUNT_SID`
  - `DEV_INTERNAL_TWILIO_AUTH_TOKEN`
  - `DEV_INTERNAL_TWILIO_FROM_NUMBER`
  - `DEV_INTERNAL_TWILIO_MESSAGING_SID`
  - (Same for STAGING_ and MAIN_ prefixes)
- Updated verification step to check unified JSON at `/myfusionhelper/${STAGE}/secrets`
- Verification now validates JSON structure (stripe, groq, twilio sections)

### 2. Rewritten Secrets Build Script
**File**: `backend/golang/scripts/build-internal-secrets.sh`

**Changes**:
- Complete rewrite following listbackup-ai pattern
- Builds individual JSON objects for:
  - Stripe (secret_key, publishable_key, webhook_secret)
  - Groq (api_key)
  - Twilio (account_sid, auth_token, from_number, messaging_sid - optional)
- Combines into unified JSON structure
- Uploads to **ONE SSM parameter**: `/myfusionhelper/${STAGE}/secrets` (tier: Advanced)
- Creates backward compatibility parameters for existing code:
  - `/${STAGE}/stripe/secret_key`
  - `/${STAGE}/stripe/publishable_key`
  - `/${STAGE}/stripe/webhook_secret`
- Validates all required secrets before upload
- Provides detailed summary with usage examples

**Unified JSON Structure**:
```json
{
  "stripe": {
    "secret_key": "...",
    "publishable_key": "...",
    "webhook_secret": "..."
  },
  "groq": {
    "api_key": "..."
  },
  "twilio": {
    "account_sid": "...",
    "auth_token": "...",
    "from_number": "...",
    "messaging_sid": "..."
  }
}
```

### 3. Added DynamoDB Tables
**File**: `backend/golang/services/infrastructure/dynamodb/core/serverless.yml`

**New Tables**:

1. **ChatConversationsTable** (`mfh-${stage}-chat-conversations`)
   - PK: `conversation_id` (S)
   - GSI: `AccountIdIndex` (account_id + created_at)
   - TTL enabled (ttl attribute)
   - Deletion protection enabled

2. **ChatMessagesTable** (`mfh-${stage}-chat-messages`)
   - PK: `message_id` (S)
   - GSI: `ConversationIdIndex` (conversation_id + sequence)
   - Deletion protection enabled

3. **PhoneMappingsTable** (`mfh-${stage}-phone-mappings`)
   - PK: `phone_number` (S)
   - Deletion protection enabled

**CloudFormation Exports Added**:
- `ChatConversationsTableName`
- `ChatConversationsTableArn`
- `ChatMessagesTableName`
- `ChatMessagesTableArn`
- `PhoneMappingsTableName`
- `PhoneMappingsTableArn`

### 4. Added Go Dependencies
**File**: `backend/golang/go.mod`

**Added**:
- `github.com/sashabaranov/go-openai v1.17.9` (Groq API client - OpenAI-compatible)

### 5. Updated CI/CD Workflow
**File**: `.github/workflows/deploy-backend.yml`

**Changes**:
- Added `chat` to API services matrix (line 211)
- Added webhook workers to workers matrix (lines 274-276):
  - `sms-chat-webhook`
  - `alexa-webhook`
  - `google-assistant-webhook`

## 🔧 Manual Steps Required

### Step 6: Add GitHub Secrets

You must add the following secrets via GitHub UI (Settings → Secrets → Actions):

**Groq Secrets** (get from https://console.groq.com/keys):
```
DEV_INTERNAL_GROQ_API_KEY=gsk_...
STAGING_INTERNAL_GROQ_API_KEY=gsk_...
MAIN_INTERNAL_GROQ_API_KEY=gsk_...
```

**Twilio Secrets** (get from https://console.twilio.com):
```
DEV_INTERNAL_TWILIO_ACCOUNT_SID=AC...
DEV_INTERNAL_TWILIO_AUTH_TOKEN=...
DEV_INTERNAL_TWILIO_FROM_NUMBER=+1...
DEV_INTERNAL_TWILIO_MESSAGING_SID=MG... (optional)

STAGING_INTERNAL_TWILIO_ACCOUNT_SID=AC...
STAGING_INTERNAL_TWILIO_AUTH_TOKEN=...
STAGING_INTERNAL_TWILIO_FROM_NUMBER=+1...
STAGING_INTERNAL_TWILIO_MESSAGING_SID=MG... (optional)

MAIN_INTERNAL_TWILIO_ACCOUNT_SID=AC...
MAIN_INTERNAL_TWILIO_AUTH_TOKEN=...
MAIN_INTERNAL_TWILIO_FROM_NUMBER=+1...
MAIN_INTERNAL_TWILIO_MESSAGING_SID=MG... (optional)
```

**Note**: Twilio secrets are optional. If not provided, SMS features will be disabled but deployment will succeed.

### Step 7: Run Secrets Sync Workflow

1. Go to GitHub Actions → "Sync Internal Secrets to SSM"
2. Click "Run workflow"
3. Select stage: `dev`
4. Run workflow
5. Verify output shows:
   - ✓ Stripe secrets loaded
   - ✓ Groq secrets loaded
   - ✓ Twilio secrets loaded (or warning if not configured)
   - ✓ Uploaded unified JSON to SSM

### Step 8: Deploy Infrastructure to Dev

Deploy DynamoDB tables:
```bash
cd backend/golang/services/infrastructure/dynamodb/core
npx sls deploy --stage dev --region us-west-2
```

Expected output:
- ✓ ChatConversationsTable created
- ✓ ChatMessagesTable created
- ✓ PhoneMappingsTable created
- CloudFormation exports available

### Step 9: Verify Deployment

**Verify Unified JSON in SSM**:
```bash
aws ssm get-parameter \
  --name "/myfusionhelper/dev/secrets" \
  --with-decryption \
  --query 'Parameter.Value' \
  --output text \
  --profile listbackup.ai \
  --region us-west-2 | jq '.'
```

Expected output:
```json
{
  "stripe": {
    "secret_key": "sk_test_...",
    "publishable_key": "pk_test_...",
    "webhook_secret": "whsec_..."
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

**Verify DynamoDB Tables**:
```bash
aws cloudformation describe-stacks \
  --stack-name mfh-infrastructure-dynamodb-core-dev \
  --region us-west-2 \
  --profile listbackup.ai \
  --query 'Stacks[0].Outputs[?OutputKey==`ChatConversationsTableName`]'

aws dynamodb describe-table \
  --table-name mfh-dev-chat-conversations \
  --region us-west-2 \
  --profile listbackup.ai \
  --query 'Table.TableStatus'
```

Expected: `"ACTIVE"`

**Run Go Tests**:
```bash
cd backend/golang
go mod download
CGO_ENABLED=1 go test ./...
```

## 📋 Next Steps (After Infrastructure is Deployed)

Once infrastructure is verified, proceed with:

1. **Task 44509052: Implement MCP Service Foundation** (blocked by this task)
   - Create `internal/services/mcp_service.go`
   - Define tool schemas
   - Implement ExecuteTool() router
   - Integration tests with Data Explorer API

2. **Task 44509053: Implement Chat Service Backend** (blocked by 44509051, 44509052)
   - Create chat service handlers
   - Implement conversation CRUD
   - Implement streaming messages endpoint
   - Groq API integration

3. **Other tasks** (see Teamwork task list 3256220)

## 🔗 References

- Plan file: `~/.claude/plans/lovely-gathering-boole.md`
- Teamwork Project: 674054 (myfusionhelper.ai)
- Teamwork Task List: 3256220 (P1 - Apps & Voice Assistants)
- Teamwork Notebook: 417849 (Voice Assistants & Chat-with-Data Integration Plan)

## 💡 Lambda Usage Pattern (for future services)

When implementing services that need secrets:

```go
package main

import (
    "context"
    "encoding/json"
    "os"

    "github.com/aws/aws-sdk-go-v2/config"
    "github.com/aws/aws-sdk-go-v2/service/ssm"
)

type SecretsConfig struct {
    Stripe struct {
        SecretKey      string `json:"secret_key"`
        PublishableKey string `json:"publishable_key"`
        WebhookSecret  string `json:"webhook_secret"`
    } `json:"stripe"`
    Groq struct {
        APIKey string `json:"api_key"`
    } `json:"groq"`
    Twilio struct {
        AccountSID   string `json:"account_sid"`
        AuthToken    string `json:"auth_token"`
        FromNumber   string `json:"from_number"`
        MessagingSID string `json:"messaging_sid"`
    } `json:"twilio"`
}

var secrets SecretsConfig

func init() {
    stage := os.Getenv("STAGE")

    // Load unified secrets once at init
    cfg, _ := config.LoadDefaultConfig(context.Background())
    client := ssm.NewFromConfig(cfg)

    result, _ := client.GetParameter(context.Background(), &ssm.GetParameterInput{
        Name:           aws.String("/myfusionhelper/" + stage + "/secrets"),
        WithDecryption: aws.Bool(true),
    })

    json.Unmarshal([]byte(*result.Parameter.Value), &secrets)
}

func HandleRequest(ctx context.Context, event events.APIGatewayV2HTTPRequest) {
    // Use secrets
    groqAPIKey := secrets.Groq.APIKey
    twilioSID := secrets.Twilio.AccountSID

    // ... rest of handler
}
```

## ✅ Task Completion Checklist

- [x] Update GitHub Actions workflow (add Groq + Twilio env vars)
- [x] Rewrite build-internal-secrets.sh (unified JSON pattern)
- [x] Add DynamoDB table definitions (chat tables)
- [x] Add go-openai dependency to go.mod
- [x] Update deploy-backend.yml (add chat + webhook services)
- [ ] Add GitHub Secrets (manual - user action required)
- [ ] Run sync-internal-secrets workflow (manual - user action required)
- [ ] Deploy infrastructure to dev (manual - user action required)
- [ ] Verify unified JSON in SSM (manual - user action required)
- [ ] Verify DynamoDB tables created (manual - user action required)

**Status**: Ready for manual deployment steps
