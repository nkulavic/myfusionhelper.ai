#!/bin/bash
set -e

# Setup Stripe webhook endpoints for a given stage.
# Usage: ./setup-stripe-webhooks.sh <stage>
#
# This script is idempotent -- safe to run multiple times.
# It ensures the correct Stripe webhook endpoint exists with the right events.
#
# The webhook signing secret follows the standard secrets flow:
#   1. This script creates the endpoint and outputs the signing secret
#   2. Store the secret as a GitHub Secret: <STAGE>_INTERNAL_STRIPE_WEBHOOK_SECRET
#   3. Run sync-internal-secrets.yml to push it to SSM
#   4. The billing Lambda reads it from SSM at runtime
#
# On subsequent runs (endpoint already exists), the script updates the
# subscribed events but does NOT output or modify the signing secret.
#
# Prerequisites:
#   - curl and jq installed
#   - STRIPE_SECRET_KEY env var set, OR Stripe key in SSM at /myfusionhelper/<stage>/secrets
#   - AWS CLI configured (only needed if reading Stripe key from SSM)

STAGE=${1:-dev}
REGION=${2:-us-west-2}

echo "=== Setting up Stripe webhooks for stage: $STAGE ==="

# Determine API domain based on stage
case "$STAGE" in
  dev)
    API_DOMAIN="api-dev.myfusionhelper.ai"
    ;;
  staging)
    API_DOMAIN="api-staging.myfusionhelper.ai"
    ;;
  main|prod|production)
    API_DOMAIN="api.myfusionhelper.ai"
    ;;
  *)
    echo "ERROR: Unknown stage '$STAGE'. Expected: dev, staging, main"
    exit 1
    ;;
esac

WEBHOOK_URL="https://${API_DOMAIN}/billing/webhook"
echo "Webhook URL: $WEBHOOK_URL"

# Get Stripe secret key -- prefer env var, fall back to SSM
if [ -n "$STRIPE_SECRET_KEY" ]; then
  echo "Using STRIPE_SECRET_KEY from environment"
else
  echo "Loading Stripe secret key from SSM..."
  SECRETS_JSON=$(aws ssm get-parameter \
    --name "/myfusionhelper/${STAGE}/secrets" \
    --with-decryption \
    --query 'Parameter.Value' \
    --output text \
    --region "$REGION" 2>/dev/null || echo "")

  if [ -z "$SECRETS_JSON" ]; then
    echo "ERROR: Could not load secrets from SSM. Set STRIPE_SECRET_KEY env var or run sync-internal-secrets workflow first."
    exit 1
  fi

  STRIPE_SECRET_KEY=$(echo "$SECRETS_JSON" | jq -r '.stripe.secret_key // empty')
  if [ -z "$STRIPE_SECRET_KEY" ]; then
    echo "ERROR: Stripe secret key not found in SSM secrets."
    exit 1
  fi
fi

echo "Stripe key loaded (${STRIPE_SECRET_KEY:0:7}...)"

# Define the events our webhook handler supports
WEBHOOK_EVENTS=(
  "checkout.session.completed"
  "customer.subscription.created"
  "customer.subscription.updated"
  "customer.subscription.deleted"
  "invoice.payment_failed"
  "invoice.paid"
  "customer.created"
  "customer.updated"
  "customer.deleted"
)

# Check for existing webhook endpoint with our URL
echo "Checking for existing webhook endpoints..."
EXISTING=$(curl -s "https://api.stripe.com/v1/webhook_endpoints?limit=100" \
  -u "${STRIPE_SECRET_KEY}:" 2>/dev/null)

EXISTING_ID=$(echo "$EXISTING" | jq -r --arg url "$WEBHOOK_URL" '.data[] | select(.url == $url) | .id' 2>/dev/null | head -1)

if [ -n "$EXISTING_ID" ] && [ "$EXISTING_ID" != "null" ]; then
  echo "Found existing webhook endpoint: $EXISTING_ID"
  echo "Updating subscribed events..."

  # Build the update request
  UPDATE_ARGS=(-s "https://api.stripe.com/v1/webhook_endpoints/${EXISTING_ID}" -u "${STRIPE_SECRET_KEY}:")
  for event in "${WEBHOOK_EVENTS[@]}"; do
    UPDATE_ARGS+=(-d "enabled_events[]=${event}")
  done
  UPDATE_ARGS+=(-d "description=MyFusion Helper ${STAGE} - Billing Webhooks")

  RESULT=$(curl "${UPDATE_ARGS[@]}" 2>/dev/null)

  if echo "$RESULT" | jq -e '.error' >/dev/null 2>&1; then
    echo "ERROR updating webhook: $(echo "$RESULT" | jq -r '.error.message')"
    exit 1
  fi

  ENDPOINT_STATUS=$(echo "$RESULT" | jq -r '.status')
  ENABLED_EVENTS=$(echo "$RESULT" | jq -r '.enabled_events | join(", ")')
  echo "Updated webhook endpoint. Status: $ENDPOINT_STATUS"
  echo "Events: $ENABLED_EVENTS"

else
  echo "No existing webhook endpoint found. Creating new one..."

  # Build the create request
  CREATE_ARGS=(-s "https://api.stripe.com/v1/webhook_endpoints" -u "${STRIPE_SECRET_KEY}:")
  CREATE_ARGS+=(-d "url=${WEBHOOK_URL}")
  for event in "${WEBHOOK_EVENTS[@]}"; do
    CREATE_ARGS+=(-d "enabled_events[]=${event}")
  done
  CREATE_ARGS+=(-d "description=MyFusion Helper ${STAGE} - Billing Webhooks")
  CREATE_ARGS+=(-d "api_version=2025-08-27.basil")

  RESULT=$(curl "${CREATE_ARGS[@]}" 2>/dev/null)

  if echo "$RESULT" | jq -e '.error' >/dev/null 2>&1; then
    echo "ERROR creating webhook: $(echo "$RESULT" | jq -r '.error.message')"
    exit 1
  fi

  ENDPOINT_ID=$(echo "$RESULT" | jq -r '.id')
  WEBHOOK_SECRET=$(echo "$RESULT" | jq -r '.secret')
  ENDPOINT_STATUS=$(echo "$RESULT" | jq -r '.status')
  ENABLED_EVENTS=$(echo "$RESULT" | jq -r '.enabled_events | join(", ")')

  echo ""
  echo "Created webhook endpoint:"
  echo "  ID:     $ENDPOINT_ID"
  echo "  Status: $ENDPOINT_STATUS"
  echo "  Events: $ENABLED_EVENTS"
  echo ""
  echo "========================================================"
  echo "  ACTION REQUIRED: Store the webhook signing secret"
  echo "========================================================"
  echo ""
  echo "  Add this as a GitHub Secret:"
  echo ""
  STAGE_UPPER=$(echo "$STAGE" | tr '[:lower:]' '[:upper:]')
  echo "    Name:  ${STAGE_UPPER}_INTERNAL_STRIPE_WEBHOOK_SECRET"
  echo "    Value: $WEBHOOK_SECRET"
  echo ""
  echo "  Then run the sync-internal-secrets workflow for stage '$STAGE'"
  echo "  to push it to SSM."
  echo ""
  echo "========================================================"
fi

echo ""
echo "=== Stripe webhook setup complete for stage: $STAGE ==="
echo "  URL:    $WEBHOOK_URL"
echo "  Events: ${WEBHOOK_EVENTS[*]}"
