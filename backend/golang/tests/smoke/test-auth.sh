#!/bin/bash

# Smoke Test: Authentication Flow
# Tests: Register, Login, Refresh Token, Logout

set -e

API_BASE_URL="${API_BASE_URL:-http://localhost:8080}"
TEST_EMAIL="smoketest+$(date +%s)@myfusionhelper.ai"
TEST_PASSWORD="TestPassword123!"
TEST_NAME="Smoke Test User"

echo "========================================"
echo "Authentication Flow Smoke Test"
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

# Test 1: Register new user
echo "1. Testing user registration..."
REGISTER_RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$API_BASE_URL/api/v1/auth/register" \
    -H "Content-Type: application/json" \
    -d "{\"email\":\"$TEST_EMAIL\",\"password\":\"$TEST_PASSWORD\",\"name\":\"$TEST_NAME\"}")

REGISTER_BODY=$(echo "$REGISTER_RESPONSE" | sed '$d')
REGISTER_STATUS=$(echo "$REGISTER_RESPONSE" | tail -n1)

if assert_response "User registration" 201 "$REGISTER_STATUS" "$REGISTER_BODY"; then
    USER_ID=$(echo "$REGISTER_BODY" | grep -o '"id":"[^"]*' | cut -d'"' -f4)
    echo "  Created user ID: $USER_ID"
fi
echo ""

# Test 2: Login with credentials
echo "2. Testing user login..."
LOGIN_RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$API_BASE_URL/api/v1/auth/login" \
    -H "Content-Type: application/json" \
    -d "{\"email\":\"$TEST_EMAIL\",\"password\":\"$TEST_PASSWORD\"}")

LOGIN_BODY=$(echo "$LOGIN_RESPONSE" | sed '$d')
LOGIN_STATUS=$(echo "$LOGIN_RESPONSE" | tail -n1)

if assert_response "User login" 200 "$LOGIN_STATUS" "$LOGIN_BODY"; then
    ACCESS_TOKEN=$(echo "$LOGIN_BODY" | grep -o '"access_token":"[^"]*' | cut -d'"' -f4)
    REFRESH_TOKEN=$(echo "$LOGIN_BODY" | grep -o '"refresh_token":"[^"]*' | cut -d'"' -f4)
    echo "  Obtained access token: ${ACCESS_TOKEN:0:20}..."
    echo "  Obtained refresh token: ${REFRESH_TOKEN:0:20}..."
fi
echo ""

# Test 3: Login with invalid credentials (error case)
echo "3. Testing login with invalid credentials..."
INVALID_LOGIN_RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$API_BASE_URL/api/v1/auth/login" \
    -H "Content-Type: application/json" \
    -d "{\"email\":\"$TEST_EMAIL\",\"password\":\"WrongPassword123!\"}")

INVALID_LOGIN_BODY=$(echo "$INVALID_LOGIN_RESPONSE" | sed '$d')
INVALID_LOGIN_STATUS=$(echo "$INVALID_LOGIN_RESPONSE" | tail -n1)

assert_response "Invalid login rejection" 401 "$INVALID_LOGIN_STATUS" "$INVALID_LOGIN_BODY"
echo ""

# Test 4: Access protected endpoint with valid token
echo "4. Testing authenticated request..."
if [ -n "$ACCESS_TOKEN" ]; then
    AUTH_RESPONSE=$(curl -s -w "\n%{http_code}" -X GET "$API_BASE_URL/api/v1/user/profile" \
        -H "Authorization: Bearer $ACCESS_TOKEN")

    AUTH_BODY=$(echo "$AUTH_RESPONSE" | sed '$d')
    AUTH_STATUS=$(echo "$AUTH_RESPONSE" | tail -n1)

    assert_response "Authenticated request" 200 "$AUTH_STATUS" "$AUTH_BODY"
else
    echo "✗ SKIPPED (no access token)"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
echo ""

# Test 5: Refresh access token
echo "5. Testing token refresh..."
if [ -n "$REFRESH_TOKEN" ]; then
    REFRESH_RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$API_BASE_URL/api/v1/auth/refresh" \
        -H "Content-Type: application/json" \
        -d "{\"refresh_token\":\"$REFRESH_TOKEN\"}")

    REFRESH_BODY=$(echo "$REFRESH_RESPONSE" | sed '$d')
    REFRESH_STATUS=$(echo "$REFRESH_RESPONSE" | tail -n1)

    if assert_response "Token refresh" 200 "$REFRESH_STATUS" "$REFRESH_BODY"; then
        NEW_ACCESS_TOKEN=$(echo "$REFRESH_BODY" | grep -o '"access_token":"[^"]*' | cut -d'"' -f4)
        echo "  New access token: ${NEW_ACCESS_TOKEN:0:20}..."
        ACCESS_TOKEN=$NEW_ACCESS_TOKEN
    fi
else
    echo "✗ SKIPPED (no refresh token)"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
echo ""

# Test 6: Logout
echo "6. Testing logout..."
if [ -n "$ACCESS_TOKEN" ]; then
    LOGOUT_RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$API_BASE_URL/api/v1/auth/logout" \
        -H "Authorization: Bearer $ACCESS_TOKEN")

    LOGOUT_BODY=$(echo "$LOGOUT_RESPONSE" | sed '$d')
    LOGOUT_STATUS=$(echo "$LOGOUT_RESPONSE" | tail -n1)

    assert_response "User logout" 200 "$LOGOUT_STATUS" "$LOGOUT_BODY"
else
    echo "✗ SKIPPED (no access token)"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
echo ""

# Test 7: Access protected endpoint after logout (should fail)
echo "7. Testing request after logout..."
if [ -n "$ACCESS_TOKEN" ]; then
    POST_LOGOUT_RESPONSE=$(curl -s -w "\n%{http_code}" -X GET "$API_BASE_URL/api/v1/user/profile" \
        -H "Authorization: Bearer $ACCESS_TOKEN")

    POST_LOGOUT_BODY=$(echo "$POST_LOGOUT_RESPONSE" | sed '$d')
    POST_LOGOUT_STATUS=$(echo "$POST_LOGOUT_RESPONSE" | tail -n1)

    assert_response "Request after logout rejection" 401 "$POST_LOGOUT_STATUS" "$POST_LOGOUT_BODY"
else
    echo "✗ SKIPPED (no access token)"
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
