#!/bin/bash

# Build and Upload Stripe Secrets to AWS SSM
# Reads GitHub Secrets via environment variables and writes to SSM Parameters.
# Follows naming convention: {STAGE}_INTERNAL_{CATEGORY}_{KEY}

set -euo pipefail

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
    print_colored "$CYAN" "================================================================"
    print_colored "$CYAN"  "  $1"
    print_colored "$CYAN" "================================================================"
}

# Get stage from environment variable
STAGE="${STAGE:-dev}"
STAGE_UPPER=$(echo "$STAGE" | tr '[:lower:]' '[:upper:]')

print_header "Syncing Stripe Secrets for ${STAGE_UPPER}"

# Function to get secret value from environment variable
get_secret() {
    local category=$1
    local key=$2
    local env_var="${STAGE_UPPER}_INTERNAL_${category}_${key}"
    echo "${!env_var:-}"
}

# Read Stripe secrets from environment
STRIPE_SECRET_KEY=$(get_secret "STRIPE" "SECRET_KEY")
STRIPE_PUBLISHABLE_KEY=$(get_secret "STRIPE" "PUBLISHABLE_KEY")
STRIPE_WEBHOOK_SECRET=$(get_secret "STRIPE" "WEBHOOK_SECRET")

# Validate required secrets
if [ -z "$STRIPE_SECRET_KEY" ]; then
    print_colored "$RED" "ERROR: ${STAGE_UPPER}_INTERNAL_STRIPE_SECRET_KEY is not set"
    exit 1
fi

if [ -z "$STRIPE_PUBLISHABLE_KEY" ]; then
    print_colored "$RED" "ERROR: ${STAGE_UPPER}_INTERNAL_STRIPE_PUBLISHABLE_KEY is not set"
    exit 1
fi

if [ -z "$STRIPE_WEBHOOK_SECRET" ]; then
    print_colored "$YELLOW" "WARNING: ${STAGE_UPPER}_INTERNAL_STRIPE_WEBHOOK_SECRET is not set (webhook verification will be disabled)"
fi

# Upload to SSM
print_header "Uploading to AWS SSM"

print_colored "$BLUE" "Writing /${STAGE}/stripe/secret_key..."
aws ssm put-parameter \
    --name "/${STAGE}/stripe/secret_key" \
    --value "$STRIPE_SECRET_KEY" \
    --type "SecureString" \
    --overwrite \
    --description "Stripe secret key for ${STAGE_UPPER}"

print_colored "$GREEN" "  Done"

print_colored "$BLUE" "Writing /${STAGE}/stripe/publishable_key..."
aws ssm put-parameter \
    --name "/${STAGE}/stripe/publishable_key" \
    --value "$STRIPE_PUBLISHABLE_KEY" \
    --type "SecureString" \
    --overwrite \
    --description "Stripe publishable key for ${STAGE_UPPER}"

print_colored "$GREEN" "  Done"

if [ -n "$STRIPE_WEBHOOK_SECRET" ]; then
    print_colored "$BLUE" "Writing /${STAGE}/stripe/webhook_secret..."
    aws ssm put-parameter \
        --name "/${STAGE}/stripe/webhook_secret" \
        --value "$STRIPE_WEBHOOK_SECRET" \
        --type "SecureString" \
        --overwrite \
        --description "Stripe webhook secret for ${STAGE_UPPER}"

    print_colored "$GREEN" "  Done"
else
    print_colored "$YELLOW" "  Skipping webhook_secret (not set)"
fi

# Summary
print_header "Upload Summary"

print_colored "$GREEN" "Successfully uploaded Stripe secrets for ${STAGE_UPPER}"
echo ""
print_colored "$CYAN" "SSM Parameters:"
print_colored "$BLUE" "  /${STAGE}/stripe/secret_key"
print_colored "$BLUE" "  /${STAGE}/stripe/publishable_key"
if [ -n "$STRIPE_WEBHOOK_SECRET" ]; then
    print_colored "$BLUE" "  /${STAGE}/stripe/webhook_secret"
fi
echo ""
print_colored "$CYAN" "Usage in serverless.yml:"
print_colored "$BLUE" "  STRIPE_SECRET_KEY: \${ssm:/${STAGE}/stripe/secret_key}"
print_colored "$BLUE" "  STRIPE_WEBHOOK_SECRET: \${ssm:/${STAGE}/stripe/webhook_secret}"
echo ""

exit 0
