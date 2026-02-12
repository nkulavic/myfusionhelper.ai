#!/bin/bash

# Smoke Test: CRM Connection Lifecycle
# Tests: Connect, verify, disconnect CRM integrations

set -e

API_BASE_URL="${API_BASE_URL:-http://localhost:8080}"
TEST_EMAIL="smoketest+crm+$(date +%s)@myfusionhelper.ai"
TEST_PASSWORD="TestPassword123!"
TEST_NAME="CRM Test User"

echo "========================================"
echo "CRM Connection Lifecycle Smoke Test"
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

# Test 1: List available CRM integrations
echo "1. Testing list available CRMs..."
LIST_CRM_RESPONSE=$(curl -s -w "\n%{http_code}" -X GET "$API_BASE_URL/api/v1/crm/available" \
    -H "Authorization: Bearer $ACCESS_TOKEN")

LIST_CRM_BODY=$(echo "$LIST_CRM_RESPONSE" | sed '$d')
LIST_CRM_STATUS=$(echo "$LIST_CRM_RESPONSE" | tail -n1)

if assert_response "List available CRMs" 200 "$LIST_CRM_STATUS" "$LIST_CRM_BODY"; then
    echo "  Available CRMs: $(echo "$LIST_CRM_BODY" | grep -o '"name":"[^"]*' | cut -d'"' -f4 | tr '\n' ',' | sed 's/,$//')"
fi
echo ""

# Test 2: Get current CRM connections (should be empty)
echo "2. Testing get CRM connections (empty state)..."
GET_CONNECTIONS_RESPONSE=$(curl -s -w "\n%{http_code}" -X GET "$API_BASE_URL/api/v1/crm/connections" \
    -H "Authorization: Bearer $ACCESS_TOKEN")

GET_CONNECTIONS_BODY=$(echo "$GET_CONNECTIONS_RESPONSE" | sed '$d')
GET_CONNECTIONS_STATUS=$(echo "$GET_CONNECTIONS_RESPONSE" | tail -n1)

assert_response "Get CRM connections" 200 "$GET_CONNECTIONS_STATUS" "$GET_CONNECTIONS_BODY"
echo ""

# Test 3: Initiate OAuth flow for Salesforce
echo "3. Testing initiate Salesforce OAuth..."
OAUTH_INIT_RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$API_BASE_URL/api/v1/crm/salesforce/oauth/init" \
    -H "Authorization: Bearer $ACCESS_TOKEN")

OAUTH_INIT_BODY=$(echo "$OAUTH_INIT_RESPONSE" | sed '$d')
OAUTH_INIT_STATUS=$(echo "$OAUTH_INIT_RESPONSE" | tail -n1)

if assert_response "Initiate Salesforce OAuth" 200 "$OAUTH_INIT_STATUS" "$OAUTH_INIT_BODY"; then
    OAUTH_URL=$(echo "$OAUTH_INIT_BODY" | grep -o '"authorization_url":"[^"]*' | cut -d'"' -f4)
    echo "  OAuth URL received: ${OAUTH_URL:0:50}..."
fi
echo ""

# Test 4: Test manual CRM connection (test mode)
echo "4. Testing manual CRM connection..."
CONNECT_RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$API_BASE_URL/api/v1/crm/connect" \
    -H "Authorization: Bearer $ACCESS_TOKEN" \
    -H "Content-Type: application/json" \
    -d '{"crm_type":"hubspot","credentials":{"api_key":"test_key_12345","test_mode":true}}')

CONNECT_BODY=$(echo "$CONNECT_RESPONSE" | sed '$d')
CONNECT_STATUS=$(echo "$CONNECT_RESPONSE" | tail -n1)

if assert_response "Connect CRM (test mode)" 200 "$CONNECT_STATUS" "$CONNECT_BODY"; then
    CONNECTION_ID=$(echo "$CONNECT_BODY" | grep -o '"connection_id":"[^"]*' | cut -d'"' -f4)
    echo "  Connection ID: $CONNECTION_ID"
fi
echo ""

# Test 5: Verify CRM connection status
echo "5. Testing verify CRM connection..."
if [ -n "$CONNECTION_ID" ]; then
    VERIFY_RESPONSE=$(curl -s -w "\n%{http_code}" -X GET "$API_BASE_URL/api/v1/crm/connections/$CONNECTION_ID/verify" \
        -H "Authorization: Bearer $ACCESS_TOKEN")

    VERIFY_BODY=$(echo "$VERIFY_RESPONSE" | sed '$d')
    VERIFY_STATUS=$(echo "$VERIFY_RESPONSE" | tail -n1)

    if assert_response "Verify CRM connection" 200 "$VERIFY_STATUS" "$VERIFY_BODY"; then
        IS_VALID=$(echo "$VERIFY_BODY" | grep -o '"valid":[^,}]*' | cut -d':' -f2)
        echo "  Connection valid: $IS_VALID"
    fi
else
    echo "✗ SKIPPED (no connection ID)"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
echo ""

# Test 6: Get CRM connection details
echo "6. Testing get connection details..."
if [ -n "$CONNECTION_ID" ]; then
    DETAILS_RESPONSE=$(curl -s -w "\n%{http_code}" -X GET "$API_BASE_URL/api/v1/crm/connections/$CONNECTION_ID" \
        -H "Authorization: Bearer $ACCESS_TOKEN")

    DETAILS_BODY=$(echo "$DETAILS_RESPONSE" | sed '$d')
    DETAILS_STATUS=$(echo "$DETAILS_RESPONSE" | tail -n1)

    assert_response "Get connection details" 200 "$DETAILS_STATUS" "$DETAILS_BODY"
else
    echo "✗ SKIPPED (no connection ID)"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
echo ""

# Test 7: Test invalid connection attempt
echo "7. Testing invalid CRM connection..."
INVALID_CONNECT_RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$API_BASE_URL/api/v1/crm/connect" \
    -H "Authorization: Bearer $ACCESS_TOKEN" \
    -H "Content-Type: application/json" \
    -d '{"crm_type":"invalid_crm","credentials":{"api_key":"test"}}')

INVALID_CONNECT_BODY=$(echo "$INVALID_CONNECT_RESPONSE" | sed '$d')
INVALID_CONNECT_STATUS=$(echo "$INVALID_CONNECT_RESPONSE" | tail -n1)

assert_response "Invalid CRM rejection" 400 "$INVALID_CONNECT_STATUS" "$INVALID_CONNECT_BODY"
echo ""

# Test 8: Refresh CRM connection
echo "8. Testing refresh CRM connection..."
if [ -n "$CONNECTION_ID" ]; then
    REFRESH_RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$API_BASE_URL/api/v1/crm/connections/$CONNECTION_ID/refresh" \
        -H "Authorization: Bearer $ACCESS_TOKEN")

    REFRESH_BODY=$(echo "$REFRESH_RESPONSE" | sed '$d')
    REFRESH_STATUS=$(echo "$REFRESH_RESPONSE" | tail -n1)

    assert_response "Refresh CRM connection" 200 "$REFRESH_STATUS" "$REFRESH_BODY"
else
    echo "✗ SKIPPED (no connection ID)"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
echo ""

# Test 9: Disconnect CRM
echo "9. Testing disconnect CRM..."
if [ -n "$CONNECTION_ID" ]; then
    DISCONNECT_RESPONSE=$(curl -s -w "\n%{http_code}" -X DELETE "$API_BASE_URL/api/v1/crm/connections/$CONNECTION_ID" \
        -H "Authorization: Bearer $ACCESS_TOKEN")

    DISCONNECT_BODY=$(echo "$DISCONNECT_RESPONSE" | sed '$d')
    DISCONNECT_STATUS=$(echo "$DISCONNECT_RESPONSE" | tail -n1)

    assert_response "Disconnect CRM" 200 "$DISCONNECT_STATUS" "$DISCONNECT_BODY"
else
    echo "✗ SKIPPED (no connection ID)"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
echo ""

# Test 10: Verify connection is removed
echo "10. Testing verify disconnection..."
if [ -n "$CONNECTION_ID" ]; then
    VERIFY_DISCONNECT_RESPONSE=$(curl -s -w "\n%{http_code}" -X GET "$API_BASE_URL/api/v1/crm/connections/$CONNECTION_ID" \
        -H "Authorization: Bearer $ACCESS_TOKEN")

    VERIFY_DISCONNECT_BODY=$(echo "$VERIFY_DISCONNECT_RESPONSE" | sed '$d')
    VERIFY_DISCONNECT_STATUS=$(echo "$VERIFY_DISCONNECT_RESPONSE" | tail -n1)

    assert_response "Verify disconnection" 404 "$VERIFY_DISCONNECT_STATUS" "$VERIFY_DISCONNECT_BODY"
else
    echo "✗ SKIPPED (no connection ID)"
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
