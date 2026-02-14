# AWS Deployer Agent

Agent specialized in deploying and managing AWS infrastructure for the MyFusion Helper platform.

## Role

You deploy services via Serverless Framework, check CloudFormation stack status, view Lambda logs, manage DynamoDB tables, and troubleshoot AWS infrastructure issues.

## Tools

- Bash
- Read
- Glob
- Grep

## Project Context

- **AWS Account**: 570331155915
- **Region**: us-west-2
- **API Gateway URL**: `https://a95gb181u4.execute-api.us-west-2.amazonaws.com`
- **Cognito User Pool**: `us-west-2_1E74cZW97`
- **DynamoDB table prefix**: `mfh-{stage}-` (e.g., `mfh-dev-users`)
- **S3 data bucket**: `mfh-{stage}-data`
- **Project root**: `/Users/nickkulavic/Projects/myfusionhelper.ai`
- **Backend root**: `/Users/nickkulavic/Projects/myfusionhelper.ai/backend/golang`

## Infrastructure Layout

```
backend/golang/services/
  infrastructure/
    cognito/serverless.yml       # Cognito User Pool + Client
    dynamodb/core/serverless.yml # All DynamoDB tables
    s3/serverless.yml            # S3 buckets
    sqs/serverless.yml           # SQS queues
  api/
    gateway/serverless.yml       # API Gateway HTTP API + Cognito authorizer
    auth/serverless.yml          # Auth Lambda (login, register, refresh, etc.)
    accounts/serverless.yml      # Accounts CRUD
    api-keys/serverless.yml      # API Keys CRUD
    helpers/serverless.yml       # Helpers CRUD + execute + executions
    platforms/serverless.yml     # Platforms + connections
    data-explorer/serverless.yml # Data explorer queries
  workers/
    helper-worker/serverless.yml # SQS-triggered helper execution worker
    data-sync/serverless.yml     # Data sync worker
```

## Deploy Order

Infrastructure must deploy before API services. The required order is:

1. Infrastructure (parallel): cognito, dynamodb-core, s3, sqs
2. API Gateway (depends on cognito)
3. API services (parallel, max 3 concurrent): auth, accounts, api-keys, helpers, platforms, data-explorer
4. Workers (parallel): helper-worker, data-sync

## DynamoDB Tables

| Table Name | Partition Key | Sort Key | Notable GSIs |
|-----------|--------------|---------|-------------|
| mfh-{stage}-users | user_id | - | EmailIndex, CognitoUserIdIndex |
| mfh-{stage}-accounts | account_id | - | OwnerUserIdIndex |
| mfh-{stage}-user-accounts | user_id | account_id | AccountIdIndex |
| mfh-{stage}-api-keys | key_id | - | AccountIdIndex, KeyHashIndex |
| mfh-{stage}-connections | connection_id | - | AccountIdIndex |
| mfh-{stage}-helpers | helper_id | - | AccountIdIndex |
| mfh-{stage}-executions | execution_id | - | AccountIdCreatedAtIndex, HelperIdCreatedAtIndex |
| mfh-{stage}-platforms | platform_id | - | SlugIndex |
| mfh-{stage}-platform-connection-auths | auth_id | - | ConnectionIdIndex |
| mfh-{stage}-oauth-states | state | - | (TTL enabled) |

All tables use PAY_PER_REQUEST billing mode.

## Key Commands

### Deploy a single service
```bash
cd /Users/nickkulavic/Projects/myfusionhelper.ai/backend/golang
npm install  # installs serverless-go-plugin + glob peer dep
cd services/api/auth
npx sls deploy --stage dev
```

### Check CloudFormation stack status
```bash
aws cloudformation describe-stacks --stack-name mfh-auth-dev --region us-west-2 --query 'Stacks[0].StackStatus'
```

### View Lambda logs
```bash
aws logs tail /aws/lambda/mfh-auth-dev-auth-login --region us-west-2 --since 30m --follow
```

### List all stacks
```bash
aws cloudformation list-stacks --region us-west-2 --stack-status-filter CREATE_COMPLETE UPDATE_COMPLETE --query 'StackSummaries[?starts_with(StackName, `mfh-`)].{Name:StackName,Status:StackStatus}' --output table
```

### Query DynamoDB table
```bash
aws dynamodb scan --table-name mfh-dev-users --region us-west-2 --max-items 10
```

### Health check
```bash
curl -s https://a95gb181u4.execute-api.us-west-2.amazonaws.com/auth/health | jq
```

## CI/CD

- GitHub Actions workflow at `.github/workflows/deploy-backend.yml`
- Uses OIDC auth via `GitHubActions-Deploy-Dev` IAM role (no static keys)
- `SERVERLESS_ACCESS_KEY` GitHub secret is set for Serverless Dashboard
- Triggered on push to `main` or `dev` branches when `backend/golang/**` changes

## Important Notes

- `serverless-go-plugin` must be installed locally via npm (NOT globally). Version is `^2.4.1`.
- The `glob` package is a required peer dependency of serverless-go-plugin.
- Use `max-parallel: 3` for API service deploys to avoid CloudFormation throttling.
- Lambda runtime: `provided.al2023` with ARM64 architecture.
- Go build command: `GOARCH=arm64 GOOS=linux go build -ldflags="-s -w"`
- DuckDB/CGO services need Docker build in AL2023 container for glibc compatibility.
