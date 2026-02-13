#!/bin/bash
set -euo pipefail

# Build and Upload OAuth Platform Credentials to AWS SSM
# Combines GitHub Secrets (client_id/client_secret) with platform config (scopes, URLs, etc.)
# from ci_cd/seed/*/platform.json files
#
# Usage: ./build-oauth-secrets.sh [stage]
# Env vars: {STAGE_UPPER}_OAUTH_{PLATFORM}_CLIENT_ID / CLIENT_SECRET

STAGE="${STAGE:-${1:-dev}}"
STAGE_UPPER=$(echo "$STAGE" | tr '[:lower:]' '[:upper:]')

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SEED_DIR="$SCRIPT_DIR/../services/api/platforms/ci_cd/seed"

echo "Building OAuth platform credentials for $STAGE_UPPER"
echo "Reading platform configs from: $SEED_DIR"

# Get OAuth secret value from environment variable
get_oauth_secret() {
    local platform=$1
    local type=$2
    local env_var="${STAGE_UPPER}_OAUTH_${platform}_${type}"
    echo "${!env_var:-}"
}

# Map of environment variable platform names to seed directory names
declare -A PLATFORM_SEED_MAP
PLATFORM_SEED_MAP["KEAP"]="keap"
PLATFORM_SEED_MAP["GOHIGHLEVEL"]="gohighlevel"
PLATFORM_SEED_MAP["HUBSPOT"]="hubspot"
PLATFORM_SEED_MAP["GOOGLE_SHEETS"]="google_sheets"
PLATFORM_SEED_MAP["ZOOM"]="zoom"
PLATFORM_SEED_MAP["DROPBOX"]="dropbox"
PLATFORM_SEED_MAP["BOX"]="box"
PLATFORM_SEED_MAP["GOTOWEBINAR"]="gotowebinar"

# Platforms to process
PLATFORMS=("KEAP" "GOHIGHLEVEL" "HUBSPOT" "GOOGLE_SHEETS" "ZOOM" "DROPBOX" "BOX" "GOTOWEBINAR")

# Find platforms with credentials
declare -A OAUTH_PLATFORMS
for platform in "${PLATFORMS[@]}"; do
    CLIENT_ID=$(get_oauth_secret "$platform" "CLIENT_ID")
    CLIENT_SECRET=$(get_oauth_secret "$platform" "CLIENT_SECRET")

    if [ -n "$CLIENT_ID" ] && [ -n "$CLIENT_SECRET" ]; then
        echo "  + $platform credentials found"
        OAUTH_PLATFORMS[$platform]=1
    else
        echo "  - $platform skipped (credentials not set)"
    fi
done

if [ ${#OAUTH_PLATFORMS[@]} -eq 0 ]; then
    echo "ERROR: No OAuth platforms configured for $STAGE_UPPER"
    exit 1
fi

# Build unified JSON
echo ""
echo "Building unified OAuth credentials JSON..."

OAUTH_JSON="{"
FIRST=true

for platform in "${!OAUTH_PLATFORMS[@]}"; do
    CLIENT_ID=$(get_oauth_secret "$platform" "CLIENT_ID")
    CLIENT_SECRET=$(get_oauth_secret "$platform" "CLIENT_SECRET")

    SEED_NAME="${PLATFORM_SEED_MAP[$platform]:-}"
    SEED_FILE="$SEED_DIR/$SEED_NAME/platform.json"

    if [ -n "$SEED_NAME" ] && [ -f "$SEED_FILE" ]; then
        # Extract slug from seed file
        PLATFORM_SLUG=$(jq -r '.slug // ""' "$SEED_FILE")
        if [ -z "$PLATFORM_SLUG" ]; then
            PLATFORM_SLUG="$SEED_NAME"
        fi

        # Extract oauth config from seed file
        OAUTH_CONFIG=$(jq -r '.oauth // {}' "$SEED_FILE")
        AUTH_URL=$(echo "$OAUTH_CONFIG" | jq -r '.auth_url // ""')
        TOKEN_URL=$(echo "$OAUTH_CONFIG" | jq -r '.token_url // ""')
        USER_INFO_URL=$(echo "$OAUTH_CONFIG" | jq -r '.user_info_url // ""')
        SCOPES=$(echo "$OAUTH_CONFIG" | jq -c '.scopes // []')

        PLATFORM_JSON=$(jq -n \
            --arg client_id "$CLIENT_ID" \
            --arg client_secret "$CLIENT_SECRET" \
            --arg slug "$PLATFORM_SLUG" \
            --arg auth_url "$AUTH_URL" \
            --arg token_url "$TOKEN_URL" \
            --arg user_info_url "$USER_INFO_URL" \
            --argjson scopes "$SCOPES" \
            '{
                client_id: $client_id,
                client_secret: $client_secret,
                slug: $slug,
                auth_url: $auth_url,
                token_url: $token_url,
                scopes: $scopes
            } + (if $user_info_url != "" then {user_info_url: $user_info_url} else {} end)')

        JSON_KEY="$PLATFORM_SLUG"
        echo "  $PLATFORM_SLUG (full config from seed)"
    else
        PLATFORM_LOWER=$(echo "$platform" | tr '[:upper:]' '[:lower:]' | tr '_' '-')
        PLATFORM_JSON=$(jq -n \
            --arg client_id "$CLIENT_ID" \
            --arg client_secret "$CLIENT_SECRET" \
            --arg slug "$PLATFORM_LOWER" \
            '{
                client_id: $client_id,
                client_secret: $client_secret,
                slug: $slug
            }')
        JSON_KEY="$PLATFORM_LOWER"
        echo "  $PLATFORM_LOWER (minimal config, no seed file)"
    fi

    if [ "$FIRST" = false ]; then
        OAUTH_JSON="$OAUTH_JSON,"
    fi
    FIRST=false
    OAUTH_JSON="$OAUTH_JSON\"$JSON_KEY\":$PLATFORM_JSON"
done

OAUTH_JSON="$OAUTH_JSON}"

# Show structure (redacted)
echo ""
echo "OAuth credentials structure:"
echo "$OAUTH_JSON" | jq 'walk(if type == "string" then "***" else . end)'

# Upload to SSM
PARAM_NAME="/myfusionhelper/${STAGE}/platforms/oauth/credentials"
echo ""
echo "Uploading to SSM parameter: $PARAM_NAME"

aws ssm put-parameter \
    --name "$PARAM_NAME" \
    --value "$OAUTH_JSON" \
    --type "SecureString" \
    --tier "Advanced" \
    --overwrite \
    --region us-west-2

echo "Done! OAuth credentials uploaded to $PARAM_NAME"
