#!/bin/bash

# Smoke Test: API Key Management
# Tests: Create, list, rotate, revoke API keys

set -e

API_BASE_URL="${API_BASE_URL:-http://localhost:8080}"
TEST_EMAIL="smoketest+apikeys+$(date +%s)@myfusionhelper.ai"
TEST_PASSWORD="TestPassword123!"
TEST_NAME="API Keys Test User"

echo "========================================"
echo "API Key Management Smoke Test"
echo "========================================"
echo "API: $API_BASE_URL"
echo ""

# Track test results
TESTS_PASSED=0
TESTS_FAILED=0

# Helper function for test assertions
assert_response() {
    local test_name=$1
    local expected_status=$2
    local actual_status=$3
    local response=$4

    echo -n "Testing: $test_name ... "
    if [ "$actual_status" -eq "$expected_status" ]; then
        echo "✓ PASSED"
        TESTS_PASSED=$((TESTS_PASSED + 1))
        return 0
    else
        echo "✗ FAILED"
        echo "  Expected status: $expected_status"
        echo "  Actual status: $actual_status"
        echo "  Response: $response"
        TESTS_FAILED=$((TESTS_FAILED + 1))
        return 1
    fi
}

# Setup: Register and login user
echo "Setup: Creating test user..."
curl -s -X POST "$API_BASE_URL/api/v1/auth/register" \
    -H "Content-Type: application/json" \
    -d "{\"email\":\"$TEST_EMAIL\",\"password\":\"$TEST_PASSWORD\",\"name\":\"$TEST_NAME\"}" > /dev/null

LOGIN_RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$API_BASE_URL/api/v1/auth/login" \
    -H "Content-Type: application/json" \
    -d "{\"email\":\"$TEST_EMAIL\",\"password\":\"$TEST_PASSWORD\"}")

LOGIN_BODY=$(echo "$LOGIN_RESPONSE" | sed '$d')
ACCESS_TOKEN=$(echo "$LOGIN_BODY" | grep -o '"access_token":"[^"]*' | cut -d'"' -f4)

if [ -z "$ACCESS_TOKEN" ]; then
    echo "✗ FAILED: Could not authenticate test user"
    exit 1
fi
echo "✓ Test user authenticated"
echo ""

# Test 1: List API keys (should be empty initially)
echo "1. Testing list API keys (empty state)..."
LIST_KEYS_RESPONSE=$(curl -s -w "\n%{http_code}" -X GET "$API_BASE_URL/api/v1/api-keys" \
    -H "Authorization: Bearer $ACCESS_TOKEN")

LIST_KEYS_BODY=$(echo "$LIST_KEYS_RESPONSE" | sed '$d')
LIST_KEYS_STATUS=$(echo "$LIST_KEYS_RESPONSE" | tail -n1)

if assert_response "List API keys" 200 "$LIST_KEYS_STATUS" "$LIST_KEYS_BODY"; then
    KEY_COUNT=$(echo "$LIST_KEYS_BODY" | grep -o '"id":"[^"]*' | wc -l | tr -d ' ')
    echo "  Initial key count: $KEY_COUNT"
fi
echo ""

# Test 2: Create new API key
echo "2. Testing create API key..."
CREATE_KEY_RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$API_BASE_URL/api/v1/api-keys" \
    -H "Authorization: Bearer $ACCESS_TOKEN" \
    -H "Content-Type: application/json" \
    -d '{
        "name": "Production API Key",
        "description": "Key for production integrations",
        "scopes": ["read:contacts", "write:contacts", "read:helpers"],
        "expires_at": "2026-12-31T23:59:59Z"
    }')

CREATE_KEY_BODY=$(echo "$CREATE_KEY_RESPONSE" | sed '$d')
CREATE_KEY_STATUS=$(echo "$CREATE_KEY_RESPONSE" | tail -n1)

if assert_response "Create API key" 201 "$CREATE_KEY_STATUS" "$CREATE_KEY_BODY"; then
    API_KEY=$(echo "$CREATE_KEY_BODY" | grep -o '"key":"[^"]*' | cut -d'"' -f4)
    KEY_ID=$(echo "$CREATE_KEY_BODY" | grep -o '"id":"[^"]*' | cut -d'"' -f4)
    echo "  API Key: ${API_KEY:0:20}..."
    echo "  Key ID: $KEY_ID"
fi
echo ""

# Test 3: Create another API key with different scopes
echo "3. Testing create second API key..."
CREATE_KEY2_RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$API_BASE_URL/api/v1/api-keys" \
    -H "Authorization: Bearer $ACCESS_TOKEN" \
    -H "Content-Type: application/json" \
    -d '{
        "name": "Read-Only API Key",
        "description": "Limited read-only access",
        "scopes": ["read:contacts", "read:helpers"]
    }')

CREATE_KEY2_BODY=$(echo "$CREATE_KEY2_RESPONSE" | sed '$d')
CREATE_KEY2_STATUS=$(echo "$CREATE_KEY2_RESPONSE" | tail -n1)

if assert_response "Create second API key" 201 "$CREATE_KEY2_STATUS" "$CREATE_KEY2_BODY"; then
    API_KEY2=$(echo "$CREATE_KEY2_BODY" | grep -o '"key":"[^"]*' | cut -d'"' -f4)
    KEY_ID2=$(echo "$CREATE_KEY2_BODY" | grep -o '"id":"[^"]*' | cut -d'"' -f4)
    echo "  API Key 2: ${API_KEY2:0:20}..."
    echo "  Key ID 2: $KEY_ID2"
fi
echo ""

# Test 4: List API keys (should show 2)
echo "4. Testing list API keys (with keys)..."
LIST_KEYS2_RESPONSE=$(curl -s -w "\n%{http_code}" -X GET "$API_BASE_URL/api/v1/api-keys" \
    -H "Authorization: Bearer $ACCESS_TOKEN")

LIST_KEYS2_BODY=$(echo "$LIST_KEYS2_RESPONSE" | sed '$d')
LIST_KEYS2_STATUS=$(echo "$LIST_KEYS2_RESPONSE" | tail -n1)

if assert_response "List API keys" 200 "$LIST_KEYS2_STATUS" "$LIST_KEYS2_BODY"; then
    CURRENT_KEY_COUNT=$(echo "$LIST_KEYS2_BODY" | grep -o '"id":"[^"]*' | wc -l | tr -d ' ')
    echo "  Current key count: $CURRENT_KEY_COUNT"
fi
echo ""

# Test 5: Get specific API key details
echo "5. Testing get API key details..."
if [ -n "$KEY_ID" ]; then
    GET_KEY_RESPONSE=$(curl -s -w "\n%{http_code}" -X GET "$API_BASE_URL/api/v1/api-keys/$KEY_ID" \
        -H "Authorization: Bearer $ACCESS_TOKEN")

    GET_KEY_BODY=$(echo "$GET_KEY_RESPONSE" | sed '$d')
    GET_KEY_STATUS=$(echo "$GET_KEY_RESPONSE" | tail -n1)

    assert_response "Get API key details" 200 "$GET_KEY_STATUS" "$GET_KEY_BODY"
else
    echo "✗ SKIPPED (no key ID)"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
echo ""

# Test 6: Use API key to authenticate
echo "6. Testing authenticate with API key..."
if [ -n "$API_KEY" ]; then
    API_AUTH_RESPONSE=$(curl -s -w "\n%{http_code}" -X GET "$API_BASE_URL/api/v1/user/profile" \
        -H "X-API-Key: $API_KEY")

    API_AUTH_BODY=$(echo "$API_AUTH_RESPONSE" | sed '$d')
    API_AUTH_STATUS=$(echo "$API_AUTH_RESPONSE" | tail -n1)

    assert_response "Authenticate with API key" 200 "$API_AUTH_STATUS" "$API_AUTH_BODY"
else
    echo "✗ SKIPPED (no API key)"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
echo ""

# Test 7: Test scope enforcement (write with read-only key)
echo "7. Testing scope enforcement..."
if [ -n "$API_KEY2" ]; then
    SCOPE_TEST_RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$API_BASE_URL/api/v1/helpers" \
        -H "X-API-Key: $API_KEY2" \
        -H "Content-Type: application/json" \
        -d '{"name":"Test","description":"Test","prompt":"Test"}')

    SCOPE_TEST_BODY=$(echo "$SCOPE_TEST_RESPONSE" | sed '$d')
    SCOPE_TEST_STATUS=$(echo "$SCOPE_TEST_RESPONSE" | tail -n1)

    assert_response "Scope enforcement (should fail)" 403 "$SCOPE_TEST_STATUS" "$SCOPE_TEST_BODY"
else
    echo "✗ SKIPPED (no read-only API key)"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
echo ""

# Test 8: Update API key
echo "8. Testing update API key..."
if [ -n "$KEY_ID" ]; then
    UPDATE_KEY_RESPONSE=$(curl -s -w "\n%{http_code}" -X PUT "$API_BASE_URL/api/v1/api-keys/$KEY_ID" \
        -H "Authorization: Bearer $ACCESS_TOKEN" \
        -H "Content-Type: application/json" \
        -d '{
            "name": "Production API Key (Updated)",
            "description": "Updated description"
        }')

    UPDATE_KEY_BODY=$(echo "$UPDATE_KEY_RESPONSE" | sed '$d')
    UPDATE_KEY_STATUS=$(echo "$UPDATE_KEY_RESPONSE" | tail -n1)

    assert_response "Update API key" 200 "$UPDATE_KEY_STATUS" "$UPDATE_KEY_BODY"
else
    echo "✗ SKIPPED (no key ID)"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
echo ""

# Test 9: Rotate API key
echo "9. Testing rotate API key..."
if [ -n "$KEY_ID" ]; then
    ROTATE_RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$API_BASE_URL/api/v1/api-keys/$KEY_ID/rotate" \
        -H "Authorization: Bearer $ACCESS_TOKEN")

    ROTATE_BODY=$(echo "$ROTATE_RESPONSE" | sed '$d')
    ROTATE_STATUS=$(echo "$ROTATE_RESPONSE" | tail -n1)

    if assert_response "Rotate API key" 200 "$ROTATE_STATUS" "$ROTATE_BODY"; then
        NEW_API_KEY=$(echo "$ROTATE_BODY" | grep -o '"key":"[^"]*' | cut -d'"' -f4)
        echo "  New API Key: ${NEW_API_KEY:0:20}..."

        # Old key should no longer work
        echo -n "  Verifying old key revoked ... "
        OLD_KEY_TEST=$(curl -s -w "\n%{http_code}" -X GET "$API_BASE_URL/api/v1/user/profile" \
            -H "X-API-Key: $API_KEY")
        OLD_KEY_STATUS=$(echo "$OLD_KEY_TEST" | tail -n1)

        if [ "$OLD_KEY_STATUS" -eq 401 ]; then
            echo "✓ Old key revoked"
        else
            echo "✗ Old key still valid"
            TESTS_FAILED=$((TESTS_FAILED + 1))
        fi
    fi
else
    echo "✗ SKIPPED (no key ID)"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
echo ""

# Test 10: Revoke API key
echo "10. Testing revoke API key..."
if [ -n "$KEY_ID2" ]; then
    REVOKE_RESPONSE=$(curl -s -w "\n%{http_code}" -X DELETE "$API_BASE_URL/api/v1/api-keys/$KEY_ID2" \
        -H "Authorization: Bearer $ACCESS_TOKEN")

    REVOKE_BODY=$(echo "$REVOKE_RESPONSE" | sed '$d')
    REVOKE_STATUS=$(echo "$REVOKE_RESPONSE" | tail -n1)

    if assert_response "Revoke API key" 200 "$REVOKE_STATUS" "$REVOKE_BODY"; then
        # Verify key is revoked
        echo -n "  Verifying key revoked ... "
        REVOKED_TEST=$(curl -s -w "\n%{http_code}" -X GET "$API_BASE_URL/api/v1/user/profile" \
            -H "X-API-Key: $API_KEY2")
        REVOKED_STATUS=$(echo "$REVOKED_TEST" | tail -n1)

        if [ "$REVOKED_STATUS" -eq 401 ]; then
            echo "✓ Key revoked"
        else
            echo "✗ Key still valid"
            TESTS_FAILED=$((TESTS_FAILED + 1))
        fi
    fi
else
    echo "✗ SKIPPED (no key ID)"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
echo ""

# Summary
echo "========================================"
echo "Test Summary"
echo "========================================"
echo "Passed: $TESTS_PASSED"
echo "Failed: $TESTS_FAILED"
echo "Total:  $((TESTS_PASSED + TESTS_FAILED))"
echo ""

if [ $TESTS_FAILED -eq 0 ]; then
    echo "✓ All tests passed!"
    exit 0
else
    echo "✗ Some tests failed"
    exit 1
fi
