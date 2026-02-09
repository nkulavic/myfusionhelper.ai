#!/bin/bash

# Build and Upload Unified Internal Secrets to AWS SSM
# Combines individual GitHub Secrets into unified JSON structure
# Follows naming convention: {STAGE}_INTERNAL_{CATEGORY}_{KEY}

set -euo pipefail

# Color codes
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m'

print_colored() {
    local color=$1
    local message=$2
    echo -e "${color}${message}${NC}"
}

print_header() {
    echo ""
    print_colored "$CYAN" "════════════════════════════════════════════════════════════════"
    print_colored "$CYAN" "  $1"
    print_colored "$CYAN" "════════════════════════════════════════════════════════════════"
}

# Get stage from environment variable
STAGE="${STAGE:-dev}"
STAGE_UPPER=$(echo "$STAGE" | tr '[:lower:]' '[:upper:]')

print_header "Building Unified Internal Secrets for $STAGE_UPPER"

# Function to get secret value from environment variable
get_secret() {
    local category=$1
    local key=$2
    local env_var="${STAGE_UPPER}_INTERNAL_${category}_${key}"
    echo "${!env_var:-}"
}

# Build Stripe secrets object
print_colored "$BLUE" "Building Stripe secrets..."
STRIPE_SECRET_KEY=$(get_secret "STRIPE" "SECRET_KEY")
STRIPE_PUBLISHABLE_KEY=$(get_secret "STRIPE" "PUBLISHABLE_KEY")
STRIPE_WEBHOOK_SECRET=$(get_secret "STRIPE" "WEBHOOK_SECRET")

if [ -z "$STRIPE_SECRET_KEY" ] || [ -z "$STRIPE_PUBLISHABLE_KEY" ]; then
    print_colored "$RED" "❌ Missing required Stripe secrets for $STAGE_UPPER"
    exit 1
fi

# Webhook secret is optional (warn but don't fail)
if [ -z "$STRIPE_WEBHOOK_SECRET" ]; then
    print_colored "$YELLOW" "⚠️  Warning: STRIPE_WEBHOOK_SECRET not set for $STAGE_UPPER"
    STRIPE_JSON=$(jq -n \
        --arg secret_key "$STRIPE_SECRET_KEY" \
        --arg publishable_key "$STRIPE_PUBLISHABLE_KEY" \
        '{
            "secret_key": $secret_key,
            "publishable_key": $publishable_key
        }')
else
    STRIPE_JSON=$(jq -n \
        --arg secret_key "$STRIPE_SECRET_KEY" \
        --arg publishable_key "$STRIPE_PUBLISHABLE_KEY" \
        --arg webhook_secret "$STRIPE_WEBHOOK_SECRET" \
        '{
            "secret_key": $secret_key,
            "publishable_key": $publishable_key,
            "webhook_secret": $webhook_secret
        }')
fi

print_colored "$GREEN" "✓ Stripe secrets loaded"

# Build Groq secrets object
print_colored "$BLUE" "Building Groq secrets..."
GROQ_API_KEY=$(get_secret "GROQ" "API_KEY")

if [ -z "$GROQ_API_KEY" ]; then
    print_colored "$RED" "❌ Missing Groq API key for $STAGE_UPPER"
    exit 1
fi

GROQ_JSON=$(jq -n \
    --arg api_key "$GROQ_API_KEY" \
    '{
        "api_key": $api_key
    }')

print_colored "$GREEN" "✓ Groq secrets loaded"

# Build Twilio secrets object (optional - don't fail if not present)
print_colored "$BLUE" "Building Twilio secrets..."
TWILIO_ACCOUNT_SID=$(get_secret "TWILIO" "ACCOUNT_SID")
TWILIO_AUTH_TOKEN=$(get_secret "TWILIO" "AUTH_TOKEN")
TWILIO_FROM_NUMBER=$(get_secret "TWILIO" "FROM_NUMBER")
TWILIO_MESSAGING_SID=$(get_secret "TWILIO" "MESSAGING_SID")

TWILIO_JSON=""
if [ -n "$TWILIO_ACCOUNT_SID" ] && [ -n "$TWILIO_AUTH_TOKEN" ] && [ -n "$TWILIO_FROM_NUMBER" ]; then
    if [ -n "$TWILIO_MESSAGING_SID" ]; then
        TWILIO_JSON=$(jq -n \
            --arg account_sid "$TWILIO_ACCOUNT_SID" \
            --arg auth_token "$TWILIO_AUTH_TOKEN" \
            --arg from_number "$TWILIO_FROM_NUMBER" \
            --arg messaging_sid "$TWILIO_MESSAGING_SID" \
            '{
                "account_sid": $account_sid,
                "auth_token": $auth_token,
                "from_number": $from_number,
                "messaging_sid": $messaging_sid
            }')
    else
        TWILIO_JSON=$(jq -n \
            --arg account_sid "$TWILIO_ACCOUNT_SID" \
            --arg auth_token "$TWILIO_AUTH_TOKEN" \
            --arg from_number "$TWILIO_FROM_NUMBER" \
            '{
                "account_sid": $account_sid,
                "auth_token": $auth_token,
                "from_number": $from_number
            }')
    fi
    print_colored "$GREEN" "✓ Twilio secrets loaded"
else
    print_colored "$YELLOW" "⚠️  Warning: Twilio secrets not configured for $STAGE_UPPER (SMS features will be disabled)"
fi

# Combine into unified JSON
print_header "Creating Unified JSON Structure"

if [ -n "$TWILIO_JSON" ]; then
    UNIFIED_JSON=$(jq -n \
        --argjson stripe "$STRIPE_JSON" \
        --argjson groq "$GROQ_JSON" \
        --argjson twilio "$TWILIO_JSON" \
        '{
            "stripe": $stripe,
            "groq": $groq,
            "twilio": $twilio
        }')
else
    UNIFIED_JSON=$(jq -n \
        --argjson stripe "$STRIPE_JSON" \
        --argjson groq "$GROQ_JSON" \
        '{
            "stripe": $stripe,
            "groq": $groq
        }')
fi

echo "$UNIFIED_JSON" | jq '.'

# Upload unified JSON to SSM
print_header "Uploading to AWS SSM"

UNIFIED_PARAM_NAME="/myfusionhelper/${STAGE}/secrets"

print_colored "$BLUE" "Uploading unified JSON to $UNIFIED_PARAM_NAME..."

aws ssm put-parameter \
    --name "$UNIFIED_PARAM_NAME" \
    --value "$UNIFIED_JSON" \
    --type "SecureString" \
    --overwrite \
    --description "Unified internal secrets for $STAGE_UPPER environment" \
    --tier "Advanced"

print_colored "$GREEN" "✓ Uploaded unified JSON to SSM"

# Create backward compatibility individual parameters (for existing code)
print_header "Creating Backward Compatibility Parameters"

# Stripe secret key
print_colored "$BLUE" "Creating /${STAGE}/stripe/secret_key..."
aws ssm put-parameter \
    --name "/${STAGE}/stripe/secret_key" \
    --value "$STRIPE_SECRET_KEY" \
    --type "SecureString" \
    --overwrite \
    --description "Stripe secret key for $STAGE_UPPER (backward compatibility)"

# Stripe publishable key
print_colored "$BLUE" "Creating /${STAGE}/stripe/publishable_key..."
aws ssm put-parameter \
    --name "/${STAGE}/stripe/publishable_key" \
    --value "$STRIPE_PUBLISHABLE_KEY" \
    --type "SecureString" \
    --overwrite \
    --description "Stripe publishable key for $STAGE_UPPER (backward compatibility)"

# Stripe webhook secret (if available)
if [ -n "$STRIPE_WEBHOOK_SECRET" ]; then
    print_colored "$BLUE" "Creating /${STAGE}/stripe/webhook_secret..."
    aws ssm put-parameter \
        --name "/${STAGE}/stripe/webhook_secret" \
        --value "$STRIPE_WEBHOOK_SECRET" \
        --type "SecureString" \
        --overwrite \
        --description "Stripe webhook secret for $STAGE_UPPER (backward compatibility)"
fi

print_colored "$GREEN" "✓ Backward compatibility parameters created"

# Summary
print_header "Upload Summary"

print_colored "$GREEN" "✅ Successfully uploaded secrets for $STAGE_UPPER environment"
echo ""
print_colored "$CYAN" "Unified JSON Path:"
print_colored "$BLUE" "  $UNIFIED_PARAM_NAME"
echo ""
print_colored "$CYAN" "Backward Compatibility Paths:"
print_colored "$BLUE" "  /${STAGE}/stripe/secret_key"
print_colored "$BLUE" "  /${STAGE}/stripe/publishable_key"
if [ -n "$STRIPE_WEBHOOK_SECRET" ]; then
    print_colored "$BLUE" "  /${STAGE}/stripe/webhook_secret"
fi
echo ""
print_colored "$CYAN" "Secrets Included:"
print_colored "$BLUE" "  • Stripe: secret_key, publishable_key$([ -n "$STRIPE_WEBHOOK_SECRET" ] && echo ', webhook_secret' || echo '')"
print_colored "$BLUE" "  • Groq: api_key"
if [ -n "$TWILIO_JSON" ]; then
    print_colored "$BLUE" "  • Twilio: account_sid, auth_token, from_number$([ -n "$TWILIO_MESSAGING_SID" ] && echo ', messaging_sid' || echo '')"
fi
echo ""
print_colored "$CYAN" "Usage in Lambda (Go):"
print_colored "$BLUE" "  // Load unified JSON once at init"
print_colored "$BLUE" "  secretsJSON := getSSMParameter(\"/myfusionhelper/\" + stage + \"/secrets\")"
print_colored "$BLUE" "  json.Unmarshal(secretsJSON, &secrets)"
print_colored "$BLUE" "  groqKey := secrets.Groq.APIKey"
echo ""

exit 0
