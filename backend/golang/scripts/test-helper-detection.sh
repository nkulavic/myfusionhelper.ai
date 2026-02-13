#!/bin/bash
#
# Test script for CI/CD helper worker detection logic
# Simulates the detect-changed-helpers job behavior
#
# Usage:
#   bash scripts/test-helper-detection.sh [before_sha] [after_sha]
#
# If no SHAs provided, uses HEAD~1 and HEAD

set -e

BEFORE_SHA=${1:-HEAD~1}
AFTER_SHA=${2:-HEAD}

echo "Testing helper worker detection logic"
echo "======================================"
echo "Before: $BEFORE_SHA"
echo "After:  $AFTER_SHA"
echo ""

# Change to repo root
cd "$(dirname "$0")/../../../"

# Test 1: Find all helper workers (excluding core workers)
echo "Test 1: Finding all helper workers..."
ALL_HELPERS=$(find backend/golang/services/workers -maxdepth 1 -type d -name "*-worker" \
  ! -name "helper-worker" \
  ! -name "notification-worker" \
  ! -name "data-sync" \
  ! -name "executions-stream" \
  ! -name "scheduler" \
  ! -name "sms-chat-webhook" \
  ! -name "alexa-webhook" \
  ! -name "google-assistant-webhook" \
  ! -name "zoom-webhook" \
  2>/dev/null | xargs -I {} basename {} | sed 's/-worker$//' | jq -R -s -c 'split("\n") | map(select(length > 0))')

echo "All helper workers found: $ALL_HELPERS"
HELPER_COUNT=$(echo "$ALL_HELPERS" | jq 'length')
echo "Total count: $HELPER_COUNT"
echo ""

# Test 2: Detect changed helper workers via git diff
echo "Test 2: Detecting changed helper workers (git diff)..."
if git rev-parse "$BEFORE_SHA" >/dev/null 2>&1 && git rev-parse "$AFTER_SHA" >/dev/null 2>&1; then
  # Note: Using sed for macOS compatibility (GitHub Actions uses Linux grep -P)
  CHANGED=$(git diff --name-only "$BEFORE_SHA" "$AFTER_SHA" \
    | grep '^backend/golang/services/workers/' \
    | sed -n 's|^backend/golang/services/workers/\([^/]*\)-worker/.*|\1|p' \
    | grep -v -E '^(helper|notification|data-sync|executions-stream|scheduler|sms-chat-webhook|alexa-webhook|google-assistant-webhook|zoom-webhook)$' \
    | sort -u \
    | jq -R -s -c 'split("\n") | map(select(length > 0))' || echo '[]')

  echo "Changed helper workers: $CHANGED"
  CHANGED_COUNT=$(echo "$CHANGED" | jq 'length')
  echo "Changed count: $CHANGED_COUNT"

  if [ "$CHANGED" = "[]" ] || [ -z "$CHANGED" ]; then
    echo "has_changes=false"
  else
    echo "has_changes=true"
  fi
else
  echo "⚠️  Warning: Invalid git SHAs provided, skipping git diff test"
fi

echo ""
echo "======================================"
echo "Detection logic test complete!"
echo ""
echo "Expected behavior in CI/CD:"
echo "- Manual trigger (workflow_dispatch): Deploy all $HELPER_COUNT helpers"
echo "- Git push trigger: Deploy only changed helpers"
echo "- If no changes: Skip deploy-helpers job"
