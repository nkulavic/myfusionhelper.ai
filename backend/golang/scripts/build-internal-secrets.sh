#!/bin/bash
set -e

# Build unified internal secrets JSON and upload to SSM
# Usage: ./build-internal-secrets.sh <stage>
# Based on listbackup-ai pattern: ONE SSM parameter with ALL secrets as JSON

STAGE=${1:-dev}
STAGE_UPPER=$(echo "$STAGE" | tr '[:lower:]' '[:upper:]')

echo "Building unified secrets for stage: $STAGE"

# Map stage-specific env vars to generic names
export STRIPE_SECRET_KEY="${STAGE_UPPER}_INTERNAL_STRIPE_SECRET_KEY"
export STRIPE_PUBLISHABLE_KEY="${STAGE_UPPER}_INTERNAL_STRIPE_PUBLISHABLE_KEY"
export STRIPE_WEBHOOK_SECRET="${STAGE_UPPER}_INTERNAL_STRIPE_WEBHOOK_SECRET"
export STRIPE_PRICE_START="${STAGE_UPPER}_INTERNAL_STRIPE_PRICE_START"
export STRIPE_PRICE_GROW="${STAGE_UPPER}_INTERNAL_STRIPE_PRICE_GROW"
export STRIPE_PRICE_DELIVER="${STAGE_UPPER}_INTERNAL_STRIPE_PRICE_DELIVER"
export GROQ_API_KEY="${STAGE_UPPER}_INTERNAL_GROQ_API_KEY"
export TWILIO_ACCOUNT_SID="${STAGE_UPPER}_INTERNAL_TWILIO_ACCOUNT_SID"
export TWILIO_AUTH_TOKEN="${STAGE_UPPER}_INTERNAL_TWILIO_AUTH_TOKEN"
export TWILIO_FROM_NUMBER="${STAGE_UPPER}_INTERNAL_TWILIO_FROM_NUMBER"
export TWILIO_MESSAGING_SID="${STAGE_UPPER}_INTERNAL_TWILIO_MESSAGING_SID"

# Dereference the variable names to get actual values
STRIPE_SECRET_KEY="${!STRIPE_SECRET_KEY}"
STRIPE_PUBLISHABLE_KEY="${!STRIPE_PUBLISHABLE_KEY}"
STRIPE_WEBHOOK_SECRET="${!STRIPE_WEBHOOK_SECRET}"
STRIPE_PRICE_START="${!STRIPE_PRICE_START:-price_start_placeholder}"
STRIPE_PRICE_GROW="${!STRIPE_PRICE_GROW:-price_grow_placeholder}"
STRIPE_PRICE_DELIVER="${!STRIPE_PRICE_DELIVER:-price_deliver_placeholder}"
GROQ_API_KEY="${!GROQ_API_KEY:-}"
TWILIO_ACCOUNT_SID="${!TWILIO_ACCOUNT_SID:-}"
TWILIO_AUTH_TOKEN="${!TWILIO_AUTH_TOKEN:-}"
TWILIO_FROM_NUMBER="${!TWILIO_FROM_NUMBER:-}"
TWILIO_MESSAGING_SID="${!TWILIO_MESSAGING_SID:-}"

# Build Stripe JSON object
STRIPE_JSON=$(jq -n \
  --arg sk "$STRIPE_SECRET_KEY" \
  --arg pk "$STRIPE_PUBLISHABLE_KEY" \
  --arg ws "$STRIPE_WEBHOOK_SECRET" \
  --arg ps "$STRIPE_PRICE_START" \
  --arg pg "$STRIPE_PRICE_GROW" \
  --arg pd "$STRIPE_PRICE_DELIVER" \
  '{
    secret_key: $sk,
    publishable_key: $pk,
    webhook_secret: $ws,
    price_start: $ps,
    price_grow: $pg,
    price_deliver: $pd
  }')

# Build Groq JSON object (optional)
if [ -n "$GROQ_API_KEY" ]; then
  GROQ_JSON=$(jq -n --arg key "$GROQ_API_KEY" '{api_key: $key}')
else
  GROQ_JSON='{}'
fi

# Build Twilio JSON object (optional)
if [ -n "$TWILIO_ACCOUNT_SID" ]; then
  TWILIO_JSON=$(jq -n \
    --arg sid "$TWILIO_ACCOUNT_SID" \
    --arg token "$TWILIO_AUTH_TOKEN" \
    --arg from "$TWILIO_FROM_NUMBER" \
    --arg msg_sid "$TWILIO_MESSAGING_SID" \
    '{
      account_sid: $sid,
      auth_token: $token,
      from_number: $from,
      messaging_sid: $msg_sid
    }')
else
  TWILIO_JSON='{}'
fi

# Build unified secrets JSON
SECRETS_JSON=$(jq -n \
  --argjson stripe "$STRIPE_JSON" \
  --argjson groq "$GROQ_JSON" \
  --argjson twilio "$TWILIO_JSON" \
  '{
    stripe: $stripe,
    groq: $groq,
    twilio: $twilio
  }')

echo "Secrets JSON structure:"
echo "$SECRETS_JSON" | jq 'walk(if type == "string" then "***" else . end)'

# Upload to SSM Parameter Store
PARAM_NAME="/myfusionhelper/${STAGE}/secrets"
echo "Uploading to SSM parameter: $PARAM_NAME"

aws ssm put-parameter \
  --name "$PARAM_NAME" \
  --value "$SECRETS_JSON" \
  --type "SecureString" \
  --tier "Advanced" \
  --overwrite \
  --region us-west-2

echo "✅ Unified secrets uploaded successfully to $PARAM_NAME"

# Optionally create backward-compatible individual parameters for existing code
if [ "${CREATE_LEGACY_PARAMS:-false}" = "true" ]; then
  echo "Creating backward-compatible individual parameters..."
  
  aws ssm put-parameter --name "/dev/stripe/secret_key" --value "$STRIPE_SECRET_KEY" --type "SecureString" --overwrite --region us-west-2 || true
  aws ssm put-parameter --name "/dev/stripe/webhook_secret" --value "$STRIPE_WEBHOOK_SECRET" --type "SecureString" --overwrite --region us-west-2 || true
  aws ssm put-parameter --name "/dev/stripe/price_start" --value "$STRIPE_PRICE_START" --type "String" --overwrite --region us-west-2 || true
  aws ssm put-parameter --name "/dev/stripe/price_grow" --value "$STRIPE_PRICE_GROW" --type "String" --overwrite --region us-west-2 || true
  aws ssm put-parameter --name "/dev/stripe/price_deliver" --value "$STRIPE_PRICE_DELIVER" --type "String" --overwrite --region us-west-2 || true
  
  echo "✅ Legacy parameters created"
fi

echo "Done!"
