# Adding a New Helper - Runbook

This document describes the process for adding a new helper to the MyFusion Helper platform.

## Architecture Overview

Each helper runs as its own **self-contained microservice**: one Serverless Framework service that creates its own SQS FIFO queue, DLQ, and Lambda function.

```
User Action -> API -> Execution record in DynamoDB (status: "pending")
                           |
                    DynamoDB Stream fires
                           |
                    Stream Router Lambda
                           |
            Constructs queue URL from helper_type naming convention
                           |
                    SQS FIFO queue (per-helper)
                           |
              Individual Helper Worker Lambda
                           |
                  Execute helper logic
                           |
               Update execution status in DynamoDB
```

**Key Components:**
- **Stream Router**: Routes DynamoDB stream events to individual helper SQS queues
- **Per-Helper SQS FIFO Queue**: `mfh-{stage}-{kebab-name}-executions.fifo`
- **Per-Helper Lambda**: Each helper has its own worker Lambda
- **Shared Worker Handler**: `internal/worker/handler.go` -- shared SQS processing logic
- **Helper Registry**: Runtime lookup of helper implementation by `helper_type`

## Prerequisites

- Go 1.24+ installed
- AWS CLI configured with appropriate credentials
- Push access to `dev` branch (triggers CI/CD deployment)

## Step 1: Determine Helper Category

Choose the category for your helper:

- **contact**: Contact manipulation (copy_it, merge_it, assign_it, etc.)
- **data**: Data transformation (format_it, math_it, split_it, etc.)
- **tagging**: Tag management (tag_it, score_it, group_it, etc.)
- **automation**: Automation triggers (trigger_it, action_it, chain_it, etc.)
- **integration**: External integrations (hook_it, mail_it, slack_it, etc.)
- **notification**: Notifications (notify_me, email_engagement)
- **analytics**: Analytics (rfm_calculation, customer_lifetime_value)
- **platform**: Platform-specific (keap_backup)

## Step 2: Implement Helper Logic

Create your helper in `backend/golang/internal/helpers/{category}/`:

```go
// backend/golang/internal/helpers/contact/my_new_helper.go
package contact

import (
    "context"
    "github.com/myfusionhelper/api/internal/helpers"
)

type MyNewHelper struct{}

func NewMyNewHelper() helpers.Helper { return &MyNewHelper{} }

func (h *MyNewHelper) GetName() string        { return "My New Helper" }
func (h *MyNewHelper) GetType() string        { return "my_new_helper" }
func (h *MyNewHelper) GetCategory() string    { return "contact" }
func (h *MyNewHelper) GetDescription() string { return "Description of what this helper does" }

func (h *MyNewHelper) GetConfigSchema() map[string]interface{} {
    return map[string]interface{}{
        "type": "object",
        "properties": map[string]interface{}{
            "field_name": map[string]interface{}{
                "type":        "string",
                "description": "Field to operate on",
            },
        },
        "required": []string{"field_name"},
    }
}

func (h *MyNewHelper) Execute(ctx context.Context, input helpers.HelperInput) (*helpers.HelperOutput, error) {
    // Implementation here
    return &helpers.HelperOutput{
        Success: true,
        Message: "Helper executed successfully",
    }, nil
}

func (h *MyNewHelper) ValidateConfig(config map[string]interface{}) error {
    return nil
}

func (h *MyNewHelper) RequiresCRM() bool    { return true }
func (h *MyNewHelper) SupportedCRMs() []string {
    return []string{"keap", "gohighlevel", "activecampaign", "ontraport", "hubspot"}
}

// init() registers in the global registry (backward compatibility with monolith)
func init() {
    helpers.Register("my_new_helper", func() helpers.Helper { return &MyNewHelper{} })
}
```

**Important**: You need BOTH:
- `NewMyNewHelper()` -- exported factory function used by the individual worker
- `init()` -- registers in global registry for backward compatibility

## Step 3: Create the Worker Service

Create `services/workers/my-new-helper-worker/` with two files:

### main.go

```go
package main

import (
    "github.com/aws/aws-lambda-go/lambda"
    "github.com/myfusionhelper/api/internal/helpers"
    "github.com/myfusionhelper/api/internal/helpers/contact"
    "github.com/myfusionhelper/api/internal/worker"

    _ "github.com/myfusionhelper/api/internal/connectors"
)

func main() {
    helpers.Register("my_new_helper", contact.NewMyNewHelper)
    lambda.Start(worker.HandleSQSEvent)
}
```

### serverless.yml

Copy from `services/workers/tag-it-worker/serverless.yml` and change these 6 values:

| Field | Change to |
|-------|-----------|
| `service` | `mfh-my-new-helper-worker` |
| `HELPER_TYPE` | `my_new_helper` |
| `handler` | `services/workers/my-new-helper-worker/main.go` |
| `description` | `"Process my_new_helper helper execution jobs"` |
| `QueueName` (queue) | `mfh-${self:provider.stage}-my-new-helper-executions.fifo` |
| `QueueName` (DLQ) | `mfh-${self:provider.stage}-my-new-helper-dlq.fifo` |

**CRITICAL serverless-go-plugin settings (DO NOT CHANGE):**
- `cmd` must NOT include `-o bootstrap` (plugin handles output naming)
- `handler` must be Go source path relative to baseDir, NOT `bootstrap`
- `baseDir: ../../..` (points to `backend/golang/` from worker directory)

## Step 4: Verify Build

```bash
cd backend/golang
go build ./services/workers/my-new-helper-worker
go build ./...  # verify nothing else broke
```

## Step 5: Deploy via CI/CD

**All deployments MUST go through CI/CD. Never run `npx sls deploy` manually.**

```bash
git add backend/golang/internal/helpers/contact/my_new_helper.go
git add backend/golang/services/workers/my-new-helper-worker/
git commit -m "feat: Add my_new_helper worker"
git push origin dev
```

GitHub Actions will:
1. Build the Go code
2. Auto-detect the new worker directory via `git diff`
3. Deploy the new worker's Lambda + SQS queue + DLQ to us-west-2
4. The stream router will automatically route `my_new_helper` executions to the new queue

## Step 6: Verify Deployment

```bash
# Check queue exists
aws sqs get-queue-url --queue-name mfh-dev-my-new-helper-executions.fifo --region us-west-2

# Check Lambda exists
aws lambda get-function --function-name mfh-my-new-helper-worker-dev-worker --region us-west-2

# Test via API
POST /helpers
{
  "name": "My Test Helper",
  "helper_type": "my_new_helper",
  "config": { "field_name": "email" },
  "connection_id": "conn_xxx",
  "enabled": true
}

POST /helpers/{helper_id}/execute
GET /executions/{execution_id}

# Monitor logs
aws logs tail /aws/lambda/mfh-my-new-helper-worker-dev-worker --follow --region us-west-2
```

## Naming Conventions

| Item | Convention | Example |
|------|-----------|---------|
| helper_type | `snake_case` | `my_new_helper` |
| Worker directory | `{kebab-case}-worker/` | `my-new-helper-worker/` |
| Service name | `mfh-{kebab}-worker` | `mfh-my-new-helper-worker` |
| SQS queue | `mfh-{stage}-{kebab}-executions.fifo` | `mfh-dev-my-new-helper-executions.fifo` |
| SQS DLQ | `mfh-{stage}-{kebab}-dlq.fifo` | `mfh-dev-my-new-helper-dlq.fifo` |
| Factory function | `New{PascalCase}()` | `NewMyNewHelper()` |

**Kebab conversion**: `my_new_helper` -> `my-new-helper` (replace `_` with `-`)

## Verification Checklist

- [ ] Helper implementation created in `internal/helpers/{category}/`
- [ ] Factory function `NewXxx()` exported
- [ ] Helper registered in `init()` function
- [ ] Worker `main.go` created (registers one helper, calls shared handler)
- [ ] Worker `serverless.yml` created (self-contained with queue + DLQ + Lambda)
- [ ] `go build ./services/workers/my-new-helper-worker` compiles
- [ ] `go build ./...` passes
- [ ] Code committed and pushed to `dev` branch
- [ ] CI/CD deployment succeeded
- [ ] SQS queue created in us-west-2
- [ ] Lambda function created and healthy
- [ ] Helper tested via API execution

## Troubleshooting

### Helper not found error
1. Verify `helpers.Register()` is called in worker `main.go`
2. Verify factory function name matches the import
3. Check CI/CD deployed the worker successfully

### Messages not arriving at worker
1. Check stream router logs -- is it routing to the correct queue URL?
2. Verify queue name follows convention: `mfh-{stage}-{kebab}-executions.fifo`
3. Check the stream router's fallback -- messages may be going to the old monolith queue

### Worker Lambda errors
1. Check CloudWatch logs for the specific worker Lambda
2. Verify all environment variables are set (check serverless.yml)
3. Verify IAM permissions include all required DynamoDB tables

## Related Documentation

- [Go Backend CLAUDE.md](../CLAUDE.md) -- full architecture reference
- [Helpers Interface](../internal/helpers/interface.go)
- [Helper Registry](../internal/helpers/registry.go)
- [Shared Worker Handler](../internal/worker/handler.go)
- [Reference Implementation](../services/workers/tag-it-worker/)
