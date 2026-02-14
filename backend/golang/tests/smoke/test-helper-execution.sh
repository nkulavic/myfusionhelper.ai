#!/bin/bash

# Smoke Test: Helper Creation and Execution
# Tests: Create helper, execute helper, get results

set -e

API_BASE_URL="${API_BASE_URL:-http://localhost:8080}"
TEST_EMAIL="smoketest+helper+$(date +%s)@myfusionhelper.ai"
TEST_PASSWORD="TestPassword123!"
TEST_NAME="Helper Test User"

echo "========================================"
echo "Helper Creation & Execution Smoke Test"
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

# Test 1: List available helper templates
echo "1. Testing list helper templates..."
TEMPLATES_RESPONSE=$(curl -s -w "\n%{http_code}" -X GET "$API_BASE_URL/api/v1/helpers/templates" \
    -H "Authorization: Bearer $ACCESS_TOKEN")

TEMPLATES_BODY=$(echo "$TEMPLATES_RESPONSE" | sed '$d')
TEMPLATES_STATUS=$(echo "$TEMPLATES_RESPONSE" | tail -n1)

if assert_response "List helper templates" 200 "$TEMPLATES_STATUS" "$TEMPLATES_BODY"; then
    echo "  Templates available: $(echo "$TEMPLATES_BODY" | grep -o '"name":"[^"]*' | wc -l | tr -d ' ')"
fi
echo ""

# Test 2: Create a new helper
echo "2. Testing create helper..."
CREATE_HELPER_RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$API_BASE_URL/api/v1/helpers" \
    -H "Authorization: Bearer $ACCESS_TOKEN" \
    -H "Content-Type: application/json" \
    -d '{
        "name": "Lead Qualifier",
        "description": "Qualifies leads based on company size and industry",
        "prompt": "You are a lead qualification assistant. Analyze the provided lead data and determine if they meet our qualification criteria: company size > 50 employees, industry in [Technology, Finance, Healthcare]. Return a score from 0-100.",
        "type": "data_analysis",
        "settings": {
            "temperature": 0.7,
            "max_tokens": 500
        }
    }')

CREATE_HELPER_BODY=$(echo "$CREATE_HELPER_RESPONSE" | sed '$d')
CREATE_HELPER_STATUS=$(echo "$CREATE_HELPER_RESPONSE" | tail -n1)

if assert_response "Create helper" 201 "$CREATE_HELPER_STATUS" "$CREATE_HELPER_BODY"; then
    HELPER_ID=$(echo "$CREATE_HELPER_BODY" | grep -o '"id":"[^"]*' | cut -d'"' -f4)
    echo "  Helper ID: $HELPER_ID"
fi
echo ""

# Test 3: List user's helpers
echo "3. Testing list helpers..."
LIST_HELPERS_RESPONSE=$(curl -s -w "\n%{http_code}" -X GET "$API_BASE_URL/api/v1/helpers" \
    -H "Authorization: Bearer $ACCESS_TOKEN")

LIST_HELPERS_BODY=$(echo "$LIST_HELPERS_RESPONSE" | sed '$d')
LIST_HELPERS_STATUS=$(echo "$LIST_HELPERS_RESPONSE" | tail -n1)

if assert_response "List helpers" 200 "$LIST_HELPERS_STATUS" "$LIST_HELPERS_BODY"; then
    HELPER_COUNT=$(echo "$LIST_HELPERS_BODY" | grep -o '"id":"[^"]*' | wc -l | tr -d ' ')
    echo "  Helper count: $HELPER_COUNT"
fi
echo ""

# Test 4: Get specific helper details
echo "4. Testing get helper details..."
if [ -n "$HELPER_ID" ]; then
    GET_HELPER_RESPONSE=$(curl -s -w "\n%{http_code}" -X GET "$API_BASE_URL/api/v1/helpers/$HELPER_ID" \
        -H "Authorization: Bearer $ACCESS_TOKEN")

    GET_HELPER_BODY=$(echo "$GET_HELPER_RESPONSE" | sed '$d')
    GET_HELPER_STATUS=$(echo "$GET_HELPER_RESPONSE" | tail -n1)

    assert_response "Get helper details" 200 "$GET_HELPER_STATUS" "$GET_HELPER_BODY"
else
    echo "✗ SKIPPED (no helper ID)"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
echo ""

# Test 5: Execute helper with test data
echo "5. Testing execute helper..."
if [ -n "$HELPER_ID" ]; then
    EXECUTE_RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$API_BASE_URL/api/v1/helpers/$HELPER_ID/execute" \
        -H "Authorization: Bearer $ACCESS_TOKEN" \
        -H "Content-Type: application/json" \
        -d '{
            "input_data": {
                "company_name": "Acme Corp",
                "company_size": 150,
                "industry": "Technology",
                "revenue": "$10M"
            }
        }')

    EXECUTE_BODY=$(echo "$EXECUTE_RESPONSE" | sed '$d')
    EXECUTE_STATUS=$(echo "$EXECUTE_RESPONSE" | tail -n1)

    if assert_response "Execute helper" 200 "$EXECUTE_STATUS" "$EXECUTE_BODY"; then
        EXECUTION_ID=$(echo "$EXECUTE_BODY" | grep -o '"execution_id":"[^"]*' | cut -d'"' -f4)
        echo "  Execution ID: $EXECUTION_ID"
    fi
else
    echo "✗ SKIPPED (no helper ID)"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
echo ""

# Test 6: Get execution status
echo "6. Testing get execution status..."
if [ -n "$EXECUTION_ID" ]; then
    # Wait a moment for execution to complete
    sleep 2

    STATUS_RESPONSE=$(curl -s -w "\n%{http_code}" -X GET "$API_BASE_URL/api/v1/helpers/executions/$EXECUTION_ID" \
        -H "Authorization: Bearer $ACCESS_TOKEN")

    STATUS_BODY=$(echo "$STATUS_RESPONSE" | sed '$d')
    STATUS_STATUS=$(echo "$STATUS_RESPONSE" | tail -n1)

    if assert_response "Get execution status" 200 "$STATUS_STATUS" "$STATUS_BODY"; then
        EXEC_STATUS=$(echo "$STATUS_BODY" | grep -o '"status":"[^"]*' | cut -d'"' -f4)
        echo "  Execution status: $EXEC_STATUS"
    fi
else
    echo "✗ SKIPPED (no execution ID)"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
echo ""

# Test 7: Get execution results
echo "7. Testing get execution results..."
if [ -n "$EXECUTION_ID" ]; then
    RESULTS_RESPONSE=$(curl -s -w "\n%{http_code}" -X GET "$API_BASE_URL/api/v1/helpers/executions/$EXECUTION_ID/results" \
        -H "Authorization: Bearer $ACCESS_TOKEN")

    RESULTS_BODY=$(echo "$RESULTS_RESPONSE" | sed '$d')
    RESULTS_STATUS=$(echo "$RESULTS_RESPONSE" | tail -n1)

    if assert_response "Get execution results" 200 "$RESULTS_STATUS" "$RESULTS_BODY"; then
        echo "  Results retrieved successfully"
    fi
else
    echo "✗ SKIPPED (no execution ID)"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
echo ""

# Test 8: List execution history
echo "8. Testing list execution history..."
if [ -n "$HELPER_ID" ]; then
    HISTORY_RESPONSE=$(curl -s -w "\n%{http_code}" -X GET "$API_BASE_URL/api/v1/helpers/$HELPER_ID/executions" \
        -H "Authorization: Bearer $ACCESS_TOKEN")

    HISTORY_BODY=$(echo "$HISTORY_RESPONSE" | sed '$d')
    HISTORY_STATUS=$(echo "$HISTORY_RESPONSE" | tail -n1)

    if assert_response "List execution history" 200 "$HISTORY_STATUS" "$HISTORY_BODY"; then
        EXEC_COUNT=$(echo "$HISTORY_BODY" | grep -o '"execution_id":"[^"]*' | wc -l | tr -d ' ')
        echo "  Executions found: $EXEC_COUNT"
    fi
else
    echo "✗ SKIPPED (no helper ID)"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
echo ""

# Test 9: Update helper
echo "9. Testing update helper..."
if [ -n "$HELPER_ID" ]; then
    UPDATE_RESPONSE=$(curl -s -w "\n%{http_code}" -X PUT "$API_BASE_URL/api/v1/helpers/$HELPER_ID" \
        -H "Authorization: Bearer $ACCESS_TOKEN" \
        -H "Content-Type: application/json" \
        -d '{
            "name": "Lead Qualifier v2",
            "description": "Updated lead qualification logic"
        }')

    UPDATE_BODY=$(echo "$UPDATE_RESPONSE" | sed '$d')
    UPDATE_STATUS=$(echo "$UPDATE_RESPONSE" | tail -n1)

    assert_response "Update helper" 200 "$UPDATE_STATUS" "$UPDATE_BODY"
else
    echo "✗ SKIPPED (no helper ID)"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
echo ""

# Test 10: Delete helper
echo "10. Testing delete helper..."
if [ -n "$HELPER_ID" ]; then
    DELETE_RESPONSE=$(curl -s -w "\n%{http_code}" -X DELETE "$API_BASE_URL/api/v1/helpers/$HELPER_ID" \
        -H "Authorization: Bearer $ACCESS_TOKEN")

    DELETE_BODY=$(echo "$DELETE_RESPONSE" | sed '$d')
    DELETE_STATUS=$(echo "$DELETE_RESPONSE" | tail -n1)

    assert_response "Delete helper" 200 "$DELETE_STATUS" "$DELETE_BODY"
else
    echo "✗ SKIPPED (no helper ID)"
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
