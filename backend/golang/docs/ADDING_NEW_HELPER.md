# Adding a New Helper - Runbook

This document describes the process for adding a new helper to the MyFusion Helper platform after the microservices migration.

## Architecture Overview

As of the microservices migration (Feb 2026), each helper type has its own dedicated Lambda worker and SQS queue:

```
User Action → API → DynamoDB Executions Table (with Stream)
                           ↓
                    Stream Router Lambda
                           ↓
              (routes by helper_type to appropriate queue)
                           ↓
            Helper-specific SQS FIFO Queue
                           ↓
         Helper-specific Lambda Worker
                           ↓
              Execute helper logic
```

## Prerequisites

- Go 1.24+ installed
- AWS CLI configured with appropriate credentials
- Node.js 20+ (for Serverless Framework)
- Access to deploy to us-west-2 region

## Step 1: Determine Helper Category

First, determine which category your helper belongs to:

- **contact**: Contact manipulation (tag_it, copy_it, merge_it, etc.)
- **data**: Data transformation (format_it, math_it, split_it, etc.)
- **tagging**: Tag management (score_it, group_it, count_tags, etc.)
- **automation**: Automation triggers (trigger_it, action_it, chain_it, etc.)
- **integration**: External integrations (hook_it, mail_it, slack_it, etc.)
- **notification**: Notifications (notify_me, email_engagement, etc.)
- **analytics**: Analytics (rfm_calculation, customer_lifetime_value, etc.)

## Step 2: Implement Helper Logic

Create your helper implementation in `backend/golang/internal/helpers/{category}/`:

```go
// backend/golang/internal/helpers/contact/my_new_helper.go
package contact

import (
    "context"
    "github.com/myfusionhelper/api/internal/helpers"
)

type MyNewHelper struct{}

func (h *MyNewHelper) GetName() string {
    return "My New Helper"
}

func (h *MyNewHelper) GetType() string {
    return "my_new_helper"
}

func (h *MyNewHelper) GetCategory() string {
    return "contact"
}

func (h *MyNewHelper) GetDescription() string {
    return "Description of what this helper does"
}

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
    // Validation logic
    return nil
}

func (h *MyNewHelper) RequiresCRM() bool {
    return true // or false if it doesn't need CRM connection
}

func (h *MyNewHelper) SupportedCRMs() []string {
    return []string{"keap", "gohighlevel", "activecampaign", "ontraport", "hubspot"}
}

func init() {
    helpers.Register("my_new_helper", func() helpers.Helper {
        return &MyNewHelper{}
    })
}
```

## Step 3: Scaffold Worker Service

Use the template to create the worker service:

```bash
cd backend/golang/services/workers

# Copy template
cp -r _templates/helper-worker my-new-helper-worker

# Update service name in serverless.yml
cd my-new-helper-worker
```

Edit `serverless.yml`:

```yaml
service: mfh-my-new-helper-worker

provider:
  name: aws
  runtime: provided.al2023
  architecture: arm64
  region: us-west-2
  stage: ${opt:stage, 'dev'}
  memorySize: 512
  timeout: 300

  environment:
    STAGE: ${self:provider.stage}
    HELPER_TYPE: my_new_helper
    CONNECTIONS_TABLE: ${cf:mfh-infrastructure-dynamodb-core-${self:provider.stage}.ConnectionsTableName}
    PLATFORM_CONNECTION_AUTHS_TABLE: ${cf:mfh-infrastructure-dynamodb-core-${self:provider.stage}.PlatformConnectionAuthsTableName}
    EXECUTIONS_TABLE: ${cf:mfh-infrastructure-dynamodb-core-${self:provider.stage}.ExecutionsTableName}
    HELPERS_TABLE: ${cf:mfh-infrastructure-dynamodb-core-${self:provider.stage}.HelpersTableName}
    PLATFORMS_TABLE: ${cf:mfh-infrastructure-dynamodb-core-${self:provider.stage}.PlatformsTableName}
    INTERNAL_SECRETS_PARAM: /myfusionhelper/${self:provider.stage}/secrets

  iam:
    role:
      statements:
        # DynamoDB permissions
        - Effect: Allow
          Action:
            - dynamodb:GetItem
            - dynamodb:PutItem
            - dynamodb:UpdateItem
            - dynamodb:Query
          Resource:
            - ${cf:mfh-infrastructure-dynamodb-core-${self:provider.stage}.ConnectionsTableArn}
            - ${cf:mfh-infrastructure-dynamodb-core-${self:provider.stage}.PlatformConnectionAuthsTableArn}
            - ${cf:mfh-infrastructure-dynamodb-core-${self:provider.stage}.ExecutionsTableArn}
            - ${cf:mfh-infrastructure-dynamodb-core-${self:provider.stage}.HelpersTableArn}
            - ${cf:mfh-infrastructure-dynamodb-core-${self:provider.stage}.PlatformsTableArn}
        # SSM for secrets
        - Effect: Allow
          Action:
            - ssm:GetParameter
          Resource:
            - "arn:aws:ssm:${self:provider.region}:*:parameter/myfusionhelper/${self:provider.stage}/secrets"

functions:
  worker:
    handler: bootstrap
    events:
      - sqs:
          arn: ${cf:mfh-sqs-{category}-helpers-${self:provider.stage}.MyNewHelperQueueArn}
          batchSize: 10
          maximumBatchingWindowInSeconds: 5
          functionResponseType: ReportBatchItemFailures

plugins:
  - serverless-go-plugin

custom:
  go:
    baseDir: ../../..
    supportedRuntimes: ["provided.al2023"]
    buildProvidedRuntimeAsBootstrap: true
    cmd: 'GOARCH=arm64 GOOS=linux go build -ldflags="-s -w" -o .bin/bootstrap ./cmd/handlers/workers/my-new-helper-worker'
```

## Step 4: Create Handler Entry Point

Create `backend/golang/cmd/handlers/workers/my-new-helper-worker/main.go`:

```go
package main

import (
    "context"
    "encoding/json"
    "log"
    "os"

    "github.com/aws/aws-lambda-go/events"
    "github.com/aws/aws-lambda-go/lambda"
    "github.com/myfusionhelper/api/internal/helpers"

    // Import your helper category to register it
    _ "github.com/myfusionhelper/api/internal/helpers/contact"
)

func main() {
    lambda.Start(HandleSQSEvent)
}

func HandleSQSEvent(ctx context.Context, sqsEvent events.SQSEvent) (events.SQSEventResponse, error) {
    helperType := os.Getenv("HELPER_TYPE")

    failures := []events.SQSBatchItemFailure{}

    for _, record := range sqsEvent.Records {
        var input helpers.HelperInput
        if err := json.Unmarshal([]byte(record.Body), &input); err != nil {
            log.Printf("Failed to unmarshal message: %v", err)
            failures = append(failures, events.SQSBatchItemFailure{
                ItemIdentifier: record.MessageId,
            })
            continue
        }

        if err := helpers.Execute(ctx, helperType, input); err != nil {
            log.Printf("Helper execution failed: %v", err)
            failures = append(failures, events.SQSBatchItemFailure{
                ItemIdentifier: record.MessageId,
            })
            continue
        }

        log.Printf("Successfully executed helper %s for execution %s", helperType, input.ExecutionID)
    }

    return events.SQSEventResponse{
        BatchItemFailures: failures,
    }, nil
}
```

## Step 5: Add SQS Queue to Infrastructure

Add queue definition to appropriate category in `backend/golang/services/infrastructure/sqs/{category}-helpers/serverless.yml`:

```yaml
MyNewHelperQueue:
  Type: AWS::SQS::Queue
  Properties:
    QueueName: mfh-${self:provider.stage}-my-new-helper-executions.fifo
    FifoQueue: true
    ContentBasedDeduplication: true
    MessageRetentionPeriod: 1209600  # 14 days
    VisibilityTimeout: 900  # 15 minutes (3x worker timeout)
    RedrivePolicy:
      deadLetterTargetArn: !GetAtt MyNewHelperDLQ.Arn
      maxReceiveCount: 3

MyNewHelperDLQ:
  Type: AWS::SQS::Queue
  Properties:
    QueueName: mfh-${self:provider.stage}-my-new-helper-executions-dlq.fifo
    FifoQueue: true
    MessageRetentionPeriod: 1209600  # 14 days

# In Outputs section:
MyNewHelperQueueArn:
  Value: !GetAtt MyNewHelperQueue.Arn
  Export:
    Name: ${self:service}-${self:provider.stage}-MyNewHelperQueueArn

MyNewHelperQueueUrl:
  Value: !Ref MyNewHelperQueue
  Export:
    Name: ${self:service}-${self:provider.stage}-MyNewHelperQueueUrl
```

## Step 6: Update helpers-inventory.json

Add your helper to the inventory:

```bash
cd backend/golang

# Add to appropriate category in helpers-inventory.json
# Example for contact category:
{
  "contact": [
    "assign_it",
    ...
    "my_new_helper"  // Add here
  ],
  ...
}
```

## Step 7: Deploy Infrastructure (SQS Queue)

```bash
cd backend/golang/services/infrastructure/sqs/{category}-helpers
npx sls deploy --stage dev --region us-west-2
```

## Step 8: Create SSM Parameter for Queue URL

```bash
# Get the queue URL from CloudFormation
QUEUE_URL=$(aws cloudformation describe-stacks \
  --region us-west-2 \
  --stack-name mfh-sqs-{category}-helpers-dev \
  --query 'Stacks[0].Outputs[?OutputKey==`MyNewHelperQueueUrl`].OutputValue' \
  --output text)

# Create SSM parameter
aws ssm put-parameter \
  --region us-west-2 \
  --name "/mfh/dev/sqs/my_new_helper/queue-url" \
  --value "$QUEUE_URL" \
  --type String \
  --overwrite
```

## Step 9: Deploy Worker Lambda

```bash
cd backend/golang/services/workers/my-new-helper-worker
npx sls deploy --stage dev --region us-west-2
```

## Step 10: Test Helper

Test the helper by creating an execution:

```bash
# Create a test helper in DynamoDB
# Execute via API: POST /helpers/{helper_id}/execute
# Check CloudWatch Logs for worker execution
```

## Step 11: Update CI/CD Pipeline

Add the new worker to `.github/workflows/deploy-backend.yml`:

```yaml
# In the workers deployment section, add:
- name: Deploy my-new-helper-worker
  working-directory: backend/golang/services/workers/my-new-helper-worker
  run: npx sls deploy --stage ${{ env.STAGE }} --region us-west-2
```

## Verification Checklist

- [ ] Helper implementation created in `internal/helpers/{category}/`
- [ ] Helper registered in `init()` function
- [ ] Worker service scaffolded in `services/workers/`
- [ ] Handler created in `cmd/handlers/workers/`
- [ ] SQS queue added to infrastructure
- [ ] Queue deployed to us-west-2
- [ ] SSM parameter created for queue URL
- [ ] Worker Lambda deployed to us-west-2
- [ ] Helper added to `helpers-inventory.json`
- [ ] CI/CD pipeline updated
- [ ] Helper tested successfully
- [ ] CloudWatch logs show successful execution

## Troubleshooting

### Worker not receiving messages

1. Check SSM parameter exists:
   ```bash
   aws ssm get-parameter --region us-west-2 --name "/mfh/dev/sqs/my_new_helper/queue-url"
   ```

2. Verify stream-router can access SSM:
   ```bash
   aws lambda get-function-configuration --region us-west-2 --function-name mfh-stream-router-dev-router
   ```

3. Check CloudWatch Logs for stream-router errors

### Helper execution fails

1. Check worker CloudWatch Logs:
   ```bash
   aws logs tail /aws/lambda/mfh-my-new-helper-worker-dev-worker --follow
   ```

2. Verify DynamoDB permissions in IAM role

3. Check helper configuration schema validation

### Messages going to DLQ

1. Check DLQ:
   ```bash
   aws sqs get-queue-attributes \
     --region us-west-2 \
     --queue-url <DLQ-URL> \
     --attribute-names ApproximateNumberOfMessages
   ```

2. Review error logs in CloudWatch

3. Increase worker timeout if needed (max 900 seconds)

## Architecture Notes

- Each helper type has its own dedicated Lambda worker
- Stream-router reads from DynamoDB Executions stream
- Stream-router routes messages based on `helper_type` field
- Workers are triggered by SQS events (batch size 10)
- All infrastructure must be in **us-west-2** region
- SSM parameters enable dynamic queue URL lookup

## Related Documentation

- [Helper Architecture & Migration Plan](https://teamwork.com/notebooks/417846) (Teamwork)
- [Go Backend CLAUDE.md](../CLAUDE.md)
- [Helpers Interface](../internal/helpers/interface.go)
