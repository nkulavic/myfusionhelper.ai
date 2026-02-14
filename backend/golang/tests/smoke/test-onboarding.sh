#!/bin/bash

# Smoke Test: Onboarding Flow
# Tests: Complete user onboarding process

set -e

API_BASE_URL="${API_BASE_URL:-http://localhost:8080}"
TEST_EMAIL="smoketest+onboarding+$(date +%s)@myfusionhelper.ai"
TEST_PASSWORD="TestPassword123!"
TEST_NAME="Onboarding Test User"

echo "========================================"
echo "Onboarding Flow Smoke Test"
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

# Test 1: Get onboarding status (should be incomplete)
echo "1. Testing initial onboarding status..."
ONBOARDING_STATUS_RESPONSE=$(curl -s -w "\n%{http_code}" -X GET "$API_BASE_URL/api/v1/onboarding/status" \
    -H "Authorization: Bearer $ACCESS_TOKEN")

ONBOARDING_STATUS_BODY=$(echo "$ONBOARDING_STATUS_RESPONSE" | sed '$d')
ONBOARDING_STATUS=$(echo "$ONBOARDING_STATUS_RESPONSE" | tail -n1)

if assert_response "Get onboarding status" 200 "$ONBOARDING_STATUS" "$ONBOARDING_STATUS_BODY"; then
    IS_COMPLETE=$(echo "$ONBOARDING_STATUS_BODY" | grep -o '"completed":[^,}]*' | cut -d':' -f2)
    echo "  Onboarding completed: $IS_COMPLETE"
fi
echo ""

# Test 2: Update user preferences
echo "2. Testing user preferences update..."
PREFERENCES_RESPONSE=$(curl -s -w "\n%{http_code}" -X PUT "$API_BASE_URL/api/v1/onboarding/preferences" \
    -H "Authorization: Bearer $ACCESS_TOKEN" \
    -H "Content-Type: application/json" \
    -d '{"industry":"technology","company_size":"10-50","role":"sales_manager"}')

PREFERENCES_BODY=$(echo "$PREFERENCES_RESPONSE" | sed '$d')
PREFERENCES_STATUS=$(echo "$PREFERENCES_RESPONSE" | tail -n1)

assert_response "Update preferences" 200 "$PREFERENCES_STATUS" "$PREFERENCES_BODY"
echo ""

# Test 3: Select CRM type
echo "3. Testing CRM selection..."
CRM_SELECT_RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$API_BASE_URL/api/v1/onboarding/crm-select" \
    -H "Authorization: Bearer $ACCESS_TOKEN" \
    -H "Content-Type: application/json" \
    -d '{"crm_type":"salesforce"}')

CRM_SELECT_BODY=$(echo "$CRM_SELECT_RESPONSE" | sed '$d')
CRM_SELECT_STATUS=$(echo "$CRM_SELECT_RESPONSE" | tail -n1)

assert_response "Select CRM" 200 "$CRM_SELECT_STATUS" "$CRM_SELECT_BODY"
echo ""

# Test 4: Skip CRM connection (for testing purposes)
echo "4. Testing skip CRM connection..."
SKIP_CRM_RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$API_BASE_URL/api/v1/onboarding/skip-crm" \
    -H "Authorization: Bearer $ACCESS_TOKEN")

SKIP_CRM_BODY=$(echo "$SKIP_CRM_RESPONSE" | sed '$d')
SKIP_CRM_STATUS=$(echo "$SKIP_CRM_RESPONSE" | tail -n1)

assert_response "Skip CRM connection" 200 "$SKIP_CRM_STATUS" "$SKIP_CRM_BODY"
echo ""

# Test 5: Select use cases
echo "5. Testing use case selection..."
USECASE_RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$API_BASE_URL/api/v1/onboarding/use-cases" \
    -H "Authorization: Bearer $ACCESS_TOKEN" \
    -H "Content-Type: application/json" \
    -d '{"use_cases":["lead_qualification","data_enrichment","reporting"]}')

USECASE_BODY=$(echo "$USECASE_RESPONSE" | sed '$d')
USECASE_STATUS=$(echo "$USECASE_RESPONSE" | tail -n1)

assert_response "Select use cases" 200 "$USECASE_STATUS" "$USECASE_BODY"
echo ""

# Test 6: Complete onboarding
echo "6. Testing complete onboarding..."
COMPLETE_RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$API_BASE_URL/api/v1/onboarding/complete" \
    -H "Authorization: Bearer $ACCESS_TOKEN")

COMPLETE_BODY=$(echo "$COMPLETE_RESPONSE" | sed '$d')
COMPLETE_STATUS=$(echo "$COMPLETE_RESPONSE" | tail -n1)

assert_response "Complete onboarding" 200 "$COMPLETE_STATUS" "$COMPLETE_BODY"
echo ""

# Test 7: Verify onboarding is marked complete
echo "7. Testing final onboarding status..."
FINAL_STATUS_RESPONSE=$(curl -s -w "\n%{http_code}" -X GET "$API_BASE_URL/api/v1/onboarding/status" \
    -H "Authorization: Bearer $ACCESS_TOKEN")

FINAL_STATUS_BODY=$(echo "$FINAL_STATUS_RESPONSE" | sed '$d')
FINAL_STATUS=$(echo "$FINAL_STATUS_RESPONSE" | tail -n1)

if assert_response "Verify onboarding complete" 200 "$FINAL_STATUS" "$FINAL_STATUS_BODY"; then
    IS_COMPLETE=$(echo "$FINAL_STATUS_BODY" | grep -o '"completed":[^,}]*' | cut -d':' -f2)
    if [ "$IS_COMPLETE" = "true" ]; then
        echo "  ✓ Onboarding marked as complete"
    else
        echo "  ✗ Onboarding NOT marked as complete"
        TESTS_FAILED=$((TESTS_FAILED + 1))
    fi
fi
echo ""

# Test 8: Attempt to complete onboarding again (idempotency)
echo "8. Testing onboarding idempotency..."
REPEAT_RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$API_BASE_URL/api/v1/onboarding/complete" \
    -H "Authorization: Bearer $ACCESS_TOKEN")

REPEAT_BODY=$(echo "$REPEAT_RESPONSE" | sed '$d')
REPEAT_STATUS=$(echo "$REPEAT_RESPONSE" | tail -n1)

assert_response "Onboarding idempotency" 200 "$REPEAT_STATUS" "$REPEAT_BODY"
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
