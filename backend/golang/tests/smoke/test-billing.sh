#!/bin/bash

# Smoke Test: Billing Flow with Stripe
# Tests: Plans, subscriptions, payment methods, usage tracking

set -e

API_BASE_URL="${API_BASE_URL:-http://localhost:8080}"
TEST_EMAIL="smoketest+billing+$(date +%s)@myfusionhelper.ai"
TEST_PASSWORD="TestPassword123!"
TEST_NAME="Billing Test User"

echo "========================================"
echo "Billing Flow Smoke Test"
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

# Test 1: List available pricing plans
echo "1. Testing list pricing plans..."
PLANS_RESPONSE=$(curl -s -w "\n%{http_code}" -X GET "$API_BASE_URL/api/v1/billing/plans" \
    -H "Authorization: Bearer $ACCESS_TOKEN")

PLANS_BODY=$(echo "$PLANS_RESPONSE" | sed '$d')
PLANS_STATUS=$(echo "$PLANS_RESPONSE" | tail -n1)

if assert_response "List pricing plans" 200 "$PLANS_STATUS" "$PLANS_BODY"; then
    PLAN_COUNT=$(echo "$PLANS_BODY" | grep -o '"id":"[^"]*' | wc -l | tr -d ' ')
    echo "  Available plans: $PLAN_COUNT"
fi
echo ""

# Test 2: Get current subscription status (should be free tier)
echo "2. Testing get subscription status..."
SUBSCRIPTION_RESPONSE=$(curl -s -w "\n%{http_code}" -X GET "$API_BASE_URL/api/v1/billing/subscription" \
    -H "Authorization: Bearer $ACCESS_TOKEN")

SUBSCRIPTION_BODY=$(echo "$SUBSCRIPTION_RESPONSE" | sed '$d')
SUBSCRIPTION_STATUS=$(echo "$SUBSCRIPTION_RESPONSE" | tail -n1)

if assert_response "Get subscription status" 200 "$SUBSCRIPTION_STATUS" "$SUBSCRIPTION_BODY"; then
    CURRENT_PLAN=$(echo "$SUBSCRIPTION_BODY" | grep -o '"plan":"[^"]*' | cut -d'"' -f4)
    echo "  Current plan: $CURRENT_PLAN"
fi
echo ""

# Test 3: Get usage metrics
echo "3. Testing get usage metrics..."
USAGE_RESPONSE=$(curl -s -w "\n%{http_code}" -X GET "$API_BASE_URL/api/v1/billing/usage" \
    -H "Authorization: Bearer $ACCESS_TOKEN")

USAGE_BODY=$(echo "$USAGE_RESPONSE" | sed '$d')
USAGE_STATUS=$(echo "$USAGE_RESPONSE" | tail -n1)

if assert_response "Get usage metrics" 200 "$USAGE_STATUS" "$USAGE_BODY"; then
    echo "  Usage metrics retrieved"
fi
echo ""

# Test 4: Create Stripe checkout session
echo "4. Testing create checkout session..."
CHECKOUT_RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$API_BASE_URL/api/v1/billing/checkout" \
    -H "Authorization: Bearer $ACCESS_TOKEN" \
    -H "Content-Type: application/json" \
    -d '{
        "plan_id": "pro_monthly",
        "success_url": "https://app.myfusionhelper.ai/billing/success",
        "cancel_url": "https://app.myfusionhelper.ai/billing/cancel"
    }')

CHECKOUT_BODY=$(echo "$CHECKOUT_RESPONSE" | sed '$d')
CHECKOUT_STATUS=$(echo "$CHECKOUT_RESPONSE" | tail -n1)

if assert_response "Create checkout session" 200 "$CHECKOUT_STATUS" "$CHECKOUT_BODY"; then
    CHECKOUT_URL=$(echo "$CHECKOUT_BODY" | grep -o '"checkout_url":"[^"]*' | cut -d'"' -f4)
    echo "  Checkout URL: ${CHECKOUT_URL:0:50}..."
fi
echo ""

# Test 5: Create customer portal session
echo "5. Testing create customer portal session..."
PORTAL_RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$API_BASE_URL/api/v1/billing/portal" \
    -H "Authorization: Bearer $ACCESS_TOKEN" \
    -H "Content-Type: application/json" \
    -d '{
        "return_url": "https://app.myfusionhelper.ai/settings/billing"
    }')

PORTAL_BODY=$(echo "$PORTAL_RESPONSE" | sed '$d')
PORTAL_STATUS=$(echo "$PORTAL_RESPONSE" | tail -n1)

if assert_response "Create portal session" 200 "$PORTAL_STATUS" "$PORTAL_BODY"; then
    PORTAL_URL=$(echo "$PORTAL_BODY" | grep -o '"portal_url":"[^"]*' | cut -d'"' -f4)
    echo "  Portal URL: ${PORTAL_URL:0:50}..."
fi
echo ""

# Test 6: Get billing history
echo "6. Testing get billing history..."
HISTORY_RESPONSE=$(curl -s -w "\n%{http_code}" -X GET "$API_BASE_URL/api/v1/billing/history" \
    -H "Authorization: Bearer $ACCESS_TOKEN")

HISTORY_BODY=$(echo "$HISTORY_RESPONSE" | sed '$d')
HISTORY_STATUS=$(echo "$HISTORY_RESPONSE" | tail -n1)

if assert_response "Get billing history" 200 "$HISTORY_STATUS" "$HISTORY_BODY"; then
    INVOICE_COUNT=$(echo "$HISTORY_BODY" | grep -o '"invoice_id":"[^"]*' | wc -l | tr -d ' ')
    echo "  Invoices found: $INVOICE_COUNT"
fi
echo ""

# Test 7: Get upcoming invoice preview
echo "7. Testing get upcoming invoice..."
UPCOMING_RESPONSE=$(curl -s -w "\n%{http_code}" -X GET "$API_BASE_URL/api/v1/billing/upcoming-invoice" \
    -H "Authorization: Bearer $ACCESS_TOKEN")

UPCOMING_BODY=$(echo "$UPCOMING_RESPONSE" | sed '$d')
UPCOMING_STATUS=$(echo "$UPCOMING_RESPONSE" | tail -n1)

# Upcoming invoice might return 404 if no subscription exists (expected)
if [ "$UPCOMING_STATUS" -eq 200 ] || [ "$UPCOMING_STATUS" -eq 404 ]; then
    echo -n "Testing: Get upcoming invoice ... "
    echo "✓ PASSED"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -n "Testing: Get upcoming invoice ... "
    echo "✗ FAILED"
    echo "  Expected status: 200 or 404"
    echo "  Actual status: $UPCOMING_STATUS"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
echo ""

# Test 8: Track usage event (API call)
echo "8. Testing track usage event..."
TRACK_RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$API_BASE_URL/api/v1/billing/usage/track" \
    -H "Authorization: Bearer $ACCESS_TOKEN" \
    -H "Content-Type: application/json" \
    -d '{
        "event_type": "helper_execution",
        "quantity": 1,
        "metadata": {
            "helper_id": "test_helper_123"
        }
    }')

TRACK_BODY=$(echo "$TRACK_RESPONSE" | sed '$d')
TRACK_STATUS=$(echo "$TRACK_RESPONSE" | tail -n1)

assert_response "Track usage event" 200 "$TRACK_STATUS" "$TRACK_BODY"
echo ""

# Test 9: Get updated usage after tracking
echo "9. Testing updated usage metrics..."
UPDATED_USAGE_RESPONSE=$(curl -s -w "\n%{http_code}" -X GET "$API_BASE_URL/api/v1/billing/usage" \
    -H "Authorization: Bearer $ACCESS_TOKEN")

UPDATED_USAGE_BODY=$(echo "$UPDATED_USAGE_RESPONSE" | sed '$d')
UPDATED_USAGE_STATUS=$(echo "$UPDATED_USAGE_RESPONSE" | tail -n1)

assert_response "Get updated usage" 200 "$UPDATED_USAGE_STATUS" "$UPDATED_USAGE_BODY"
echo ""

# Test 10: Validate plan limits
echo "10. Testing plan limits validation..."
LIMITS_RESPONSE=$(curl -s -w "\n%{http_code}" -X GET "$API_BASE_URL/api/v1/billing/limits" \
    -H "Authorization: Bearer $ACCESS_TOKEN")

LIMITS_BODY=$(echo "$LIMITS_RESPONSE" | sed '$d')
LIMITS_STATUS=$(echo "$LIMITS_RESPONSE" | tail -n1)

if assert_response "Get plan limits" 200 "$LIMITS_STATUS" "$LIMITS_BODY"; then
    echo "  Plan limits retrieved"
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
