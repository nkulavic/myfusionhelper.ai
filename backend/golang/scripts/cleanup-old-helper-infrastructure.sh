#!/bin/bash

# Cleanup Old Helper Infrastructure
# This script removes the old monolith helper-worker and its SQS queue
# after verifying that all 97 individual helper workers are deployed

set -e

REGION="us-west-2"
STAGE="${1:-dev}"

echo "=== Helper Infrastructure Cleanup Script ==="
echo "Region: $REGION"
echo "Stage: $STAGE"
echo ""

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Step 1: Verify all 97 helper workers are deployed
echo -e "${YELLOW}Step 1: Verifying 97 helper workers are deployed...${NC}"

WORKER_COUNT=$(aws lambda list-functions --region $REGION --output json | \
  jq -r '.Functions[].FunctionName' | \
  grep -E "mfh-.*-worker-${STAGE}-worker$" | \
  grep -v "helper-worker" | \
  grep -v "notification-worker" | \
  grep -v "data-sync" | \
  wc -l | tr -d ' ')

echo "Found $WORKER_COUNT individual helper workers deployed"

if [ "$WORKER_COUNT" -lt 97 ]; then
  echo -e "${RED}ERROR: Only $WORKER_COUNT workers found, expected 97${NC}"
  echo "Cannot proceed with cleanup. Deploy all workers first."
  exit 1
fi

echo -e "${GREEN}✓ All 97 helper workers verified${NC}"
echo ""

# Step 2: Check old queue message count
echo -e "${YELLOW}Step 2: Checking old queue for remaining messages...${NC}"

OLD_QUEUE_NAME="mfh-${STAGE}-helper-executions.fifo"
OLD_QUEUE_URL=$(aws sqs get-queue-url \
  --region $REGION \
  --queue-name "$OLD_QUEUE_NAME" \
  --query 'QueueUrl' \
  --output text 2>/dev/null || echo "")

if [ -z "$OLD_QUEUE_URL" ]; then
  echo -e "${YELLOW}Old queue not found (may have been deleted already)${NC}"
else
  MESSAGE_COUNT=$(aws sqs get-queue-attributes \
    --region $REGION \
    --queue-url "$OLD_QUEUE_URL" \
    --attribute-names ApproximateNumberOfMessages \
    --query 'Attributes.ApproximateNumberOfMessages' \
    --output text)

  echo "Messages in old queue: $MESSAGE_COUNT"

  if [ "$MESSAGE_COUNT" -gt 0 ]; then
    echo -e "${RED}WARNING: Old queue still has $MESSAGE_COUNT messages${NC}"
    read -p "Continue with cleanup anyway? (yes/no): " CONTINUE
    if [ "$CONTINUE" != "yes" ]; then
      echo "Cleanup cancelled"
      exit 1
    fi
  else
    echo -e "${GREEN}✓ Old queue is empty${NC}"
  fi
fi
echo ""

# Step 3: Backup old infrastructure configuration
echo -e "${YELLOW}Step 3: Backing up old infrastructure configuration...${NC}"

BACKUP_DIR="backup/old-helper-infrastructure-$(date +%Y%m%d-%H%M%S)"
mkdir -p "$BACKUP_DIR"

# Backup old Lambda configuration
echo "Backing up old helper-worker Lambda configuration..."
aws lambda get-function-configuration \
  --region $REGION \
  --function-name "mfh-helper-worker-${STAGE}-helper-worker" \
  --output json > "$BACKUP_DIR/helper-worker-config.json" 2>/dev/null || true

# Backup old queue attributes
if [ -n "$OLD_QUEUE_URL" ]; then
  echo "Backing up old queue configuration..."
  aws sqs get-queue-attributes \
    --region $REGION \
    --queue-url "$OLD_QUEUE_URL" \
    --attribute-names All \
    --output json > "$BACKUP_DIR/old-queue-attributes.json" 2>/dev/null || true
fi

# Backup old CloudFormation stack
echo "Backing up old CloudFormation stack..."
aws cloudformation describe-stacks \
  --region $REGION \
  --stack-name "mfh-helper-worker-${STAGE}" \
  --output json > "$BACKUP_DIR/cloudformation-stack.json" 2>/dev/null || true

echo -e "${GREEN}✓ Configuration backed up to $BACKUP_DIR${NC}"
echo ""

# Step 4: Delete old helper-worker Lambda
echo -e "${YELLOW}Step 4: Deleting old helper-worker Lambda...${NC}"

STACK_NAME="mfh-helper-worker-${STAGE}"

if aws cloudformation describe-stacks --region $REGION --stack-name "$STACK_NAME" &>/dev/null; then
  echo "Deleting CloudFormation stack: $STACK_NAME"
  aws cloudformation delete-stack --region $REGION --stack-name "$STACK_NAME"

  echo "Waiting for stack deletion to complete..."
  aws cloudformation wait stack-delete-complete --region $REGION --stack-name "$STACK_NAME"

  echo -e "${GREEN}✓ Old helper-worker Lambda deleted${NC}"
else
  echo -e "${YELLOW}Stack $STACK_NAME not found (may have been deleted already)${NC}"
fi
echo ""

# Step 5: Delete old SQS queue
echo -e "${YELLOW}Step 5: Deleting old SQS queue...${NC}"

if [ -n "$OLD_QUEUE_URL" ]; then
  echo "Deleting queue: $OLD_QUEUE_NAME"
  aws sqs delete-queue --region $REGION --queue-url "$OLD_QUEUE_URL"

  # Also delete DLQ if it exists
  OLD_DLQ_NAME="mfh-${STAGE}-helper-executions-dlq.fifo"
  OLD_DLQ_URL=$(aws sqs get-queue-url \
    --region $REGION \
    --queue-name "$OLD_DLQ_NAME" \
    --query 'QueueUrl' \
    --output text 2>/dev/null || echo "")

  if [ -n "$OLD_DLQ_URL" ]; then
    echo "Deleting DLQ: $OLD_DLQ_NAME"
    aws sqs delete-queue --region $REGION --queue-url "$OLD_DLQ_URL"
  fi

  echo -e "${GREEN}✓ Old SQS queue deleted${NC}"
else
  echo -e "${YELLOW}Old queue already deleted${NC}"
fi
echo ""

# Step 6: Verify cleanup
echo -e "${YELLOW}Step 6: Verifying cleanup...${NC}"

# Check Lambda doesn't exist
if aws lambda get-function --region $REGION --function-name "mfh-helper-worker-${STAGE}-helper-worker" &>/dev/null; then
  echo -e "${RED}✗ Old helper-worker Lambda still exists${NC}"
else
  echo -e "${GREEN}✓ Old helper-worker Lambda removed${NC}"
fi

# Check queue doesn't exist
if aws sqs get-queue-url --region $REGION --queue-name "$OLD_QUEUE_NAME" &>/dev/null; then
  echo -e "${RED}✗ Old queue still exists${NC}"
else
  echo -e "${GREEN}✓ Old queue removed${NC}"
fi

echo ""
echo -e "${GREEN}=== Cleanup Complete ===${NC}"
echo ""
echo "Next steps:"
echo "1. Remove passthrough mode from stream-router"
echo "2. Monitor helper executions to ensure new workers are functioning"
echo "3. Review CloudWatch Logs for any errors"
echo ""
echo "Backup location: $BACKUP_DIR"
