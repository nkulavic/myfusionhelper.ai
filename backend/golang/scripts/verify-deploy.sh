#!/bin/bash
set -euo pipefail

STAGE="${1:-dev}"
REGION="${2:-us-west-2}"

echo "Verifying deployment: stage=${STAGE}, region=${REGION}"
echo "================================================"

# Get API Gateway endpoint
API_URL=$(aws cloudformation describe-stacks \
  --stack-name "mfh-api-gateway-${STAGE}" \
  --region "${REGION}" \
  --query "Stacks[0].Outputs[?OutputKey=='HttpApiEndpoint'].OutputValue" \
  --output text 2>/dev/null || echo "")

if [ -z "$API_URL" ]; then
  echo "FAIL: Could not retrieve API Gateway endpoint"
  exit 1
fi

echo "API Gateway: ${API_URL}"
echo ""

ENDPOINTS=(
  "/health"
  "/auth/health"
  "/accounts/health"
  "/api-keys/health"
  "/helpers/health"
  "/platforms/health"
  "/data/health"
)

PASS=0
FAIL=0

for ep in "${ENDPOINTS[@]}"; do
  STATUS=$(curl -s -o /dev/null -w "%{http_code}" --max-time 10 "${API_URL}${ep}" 2>/dev/null || echo "000")
  if [ "$STATUS" -eq 200 ]; then
    echo "  PASS: ${ep} (${STATUS})"
    PASS=$((PASS + 1))
  else
    echo "  FAIL: ${ep} (${STATUS})"
    FAIL=$((FAIL + 1))
  fi
done

echo ""

# Check worker stacks
WORKER_STACKS=("mfh-helper-worker" "mfh-data-sync")

for STACK_NAME in "${WORKER_STACKS[@]}"; do
  WORKER_STATUS=$(aws cloudformation describe-stacks \
    --stack-name "${STACK_NAME}-${STAGE}" \
    --region "${REGION}" \
    --query "Stacks[0].StackStatus" \
    --output text 2>/dev/null || echo "NOT_FOUND")

  if [[ "$WORKER_STATUS" == *"COMPLETE"* ]]; then
    echo "  PASS: ${STACK_NAME} stack (${WORKER_STATUS})"
    PASS=$((PASS + 1))
  else
    echo "  FAIL: ${STACK_NAME} stack (${WORKER_STATUS})"
    FAIL=$((FAIL + 1))
  fi
done

echo ""
echo "================================================"
echo "Results: ${PASS} passed, ${FAIL} failed"
exit $FAIL
