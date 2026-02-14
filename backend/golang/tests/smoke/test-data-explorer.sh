#!/bin/bash

# Smoke Test: Data Explorer
# Tests: Query CRM data, filters, exports

set -e

API_BASE_URL="${API_BASE_URL:-http://localhost:8080}"
TEST_EMAIL="smoketest+explorer+$(date +%s)@myfusionhelper.ai"
TEST_PASSWORD="TestPassword123!"
TEST_NAME="Explorer Test User"

echo "========================================"
echo "Data Explorer Smoke Test"
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

# Setup: Connect test CRM
echo "Setup: Connecting test CRM..."
curl -s -X POST "$API_BASE_URL/api/v1/crm/connect" \
    -H "Authorization: Bearer $ACCESS_TOKEN" \
    -H "Content-Type: application/json" \
    -d '{"crm_type":"hubspot","credentials":{"api_key":"test_key","test_mode":true}}' > /dev/null
echo "✓ Test CRM connected"
echo ""

# Test 1: List available data objects
echo "1. Testing list data objects..."
OBJECTS_RESPONSE=$(curl -s -w "\n%{http_code}" -X GET "$API_BASE_URL/api/v1/data-explorer/objects" \
    -H "Authorization: Bearer $ACCESS_TOKEN")

OBJECTS_BODY=$(echo "$OBJECTS_RESPONSE" | sed '$d')
OBJECTS_STATUS=$(echo "$OBJECTS_RESPONSE" | tail -n1)

if assert_response "List data objects" 200 "$OBJECTS_STATUS" "$OBJECTS_BODY"; then
    OBJECT_COUNT=$(echo "$OBJECTS_BODY" | grep -o '"name":"[^"]*' | wc -l | tr -d ' ')
    echo "  Available objects: $OBJECT_COUNT"
fi
echo ""

# Test 2: Get schema for specific object (contacts)
echo "2. Testing get object schema..."
SCHEMA_RESPONSE=$(curl -s -w "\n%{http_code}" -X GET "$API_BASE_URL/api/v1/data-explorer/objects/contacts/schema" \
    -H "Authorization: Bearer $ACCESS_TOKEN")

SCHEMA_BODY=$(echo "$SCHEMA_RESPONSE" | sed '$d')
SCHEMA_STATUS=$(echo "$SCHEMA_RESPONSE" | tail -n1)

if assert_response "Get object schema" 200 "$SCHEMA_STATUS" "$SCHEMA_BODY"; then
    FIELD_COUNT=$(echo "$SCHEMA_BODY" | grep -o '"field_name":"[^"]*' | wc -l | tr -d ' ')
    echo "  Fields in schema: $FIELD_COUNT"
fi
echo ""

# Test 3: Query contacts (no filters)
echo "3. Testing query contacts..."
QUERY_RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$API_BASE_URL/api/v1/data-explorer/query" \
    -H "Authorization: Bearer $ACCESS_TOKEN" \
    -H "Content-Type: application/json" \
    -d '{
        "object": "contacts",
        "limit": 10
    }')

QUERY_BODY=$(echo "$QUERY_RESPONSE" | sed '$d')
QUERY_STATUS=$(echo "$QUERY_RESPONSE" | tail -n1)

if assert_response "Query contacts" 200 "$QUERY_STATUS" "$QUERY_BODY"; then
    RECORD_COUNT=$(echo "$QUERY_BODY" | grep -o '"id":"[^"]*' | wc -l | tr -d ' ')
    echo "  Records returned: $RECORD_COUNT"
fi
echo ""

# Test 4: Query with filters
echo "4. Testing query with filters..."
FILTER_QUERY_RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$API_BASE_URL/api/v1/data-explorer/query" \
    -H "Authorization: Bearer $ACCESS_TOKEN" \
    -H "Content-Type: application/json" \
    -d '{
        "object": "contacts",
        "filters": [
            {
                "field": "company_size",
                "operator": "gt",
                "value": 50
            },
            {
                "field": "industry",
                "operator": "eq",
                "value": "Technology"
            }
        ],
        "limit": 10
    }')

FILTER_QUERY_BODY=$(echo "$FILTER_QUERY_RESPONSE" | sed '$d')
FILTER_QUERY_STATUS=$(echo "$FILTER_QUERY_RESPONSE" | tail -n1)

assert_response "Query with filters" 200 "$FILTER_QUERY_STATUS" "$FILTER_QUERY_BODY"
echo ""

# Test 5: Query with field selection
echo "5. Testing query with field selection..."
FIELD_SELECT_RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$API_BASE_URL/api/v1/data-explorer/query" \
    -H "Authorization: Bearer $ACCESS_TOKEN" \
    -H "Content-Type: application/json" \
    -d '{
        "object": "contacts",
        "fields": ["id", "email", "first_name", "last_name", "company"],
        "limit": 5
    }')

FIELD_SELECT_BODY=$(echo "$FIELD_SELECT_RESPONSE" | sed '$d')
FIELD_SELECT_STATUS=$(echo "$FIELD_SELECT_RESPONSE" | tail -n1)

assert_response "Query with field selection" 200 "$FIELD_SELECT_STATUS" "$FIELD_SELECT_BODY"
echo ""

# Test 6: Query with sorting
echo "6. Testing query with sorting..."
SORT_RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$API_BASE_URL/api/v1/data-explorer/query" \
    -H "Authorization: Bearer $ACCESS_TOKEN" \
    -H "Content-Type: application/json" \
    -d '{
        "object": "contacts",
        "sort": [
            {
                "field": "created_at",
                "direction": "desc"
            }
        ],
        "limit": 5
    }')

SORT_BODY=$(echo "$SORT_RESPONSE" | sed '$d')
SORT_STATUS=$(echo "$SORT_RESPONSE" | tail -n1)

assert_response "Query with sorting" 200 "$SORT_STATUS" "$SORT_BODY"
echo ""

# Test 7: Query with pagination
echo "7. Testing query with pagination..."
PAGE_RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$API_BASE_URL/api/v1/data-explorer/query" \
    -H "Authorization: Bearer $ACCESS_TOKEN" \
    -H "Content-Type: application/json" \
    -d '{
        "object": "contacts",
        "limit": 5,
        "offset": 5
    }')

PAGE_BODY=$(echo "$PAGE_RESPONSE" | sed '$d')
PAGE_STATUS=$(echo "$PAGE_RESPONSE" | tail -n1)

assert_response "Query with pagination" 200 "$PAGE_STATUS" "$PAGE_BODY"
echo ""

# Test 8: Export query results (CSV)
echo "8. Testing export to CSV..."
EXPORT_RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$API_BASE_URL/api/v1/data-explorer/export" \
    -H "Authorization: Bearer $ACCESS_TOKEN" \
    -H "Content-Type: application/json" \
    -d '{
        "object": "contacts",
        "format": "csv",
        "filters": [],
        "limit": 100
    }')

EXPORT_BODY=$(echo "$EXPORT_RESPONSE" | sed '$d')
EXPORT_STATUS=$(echo "$EXPORT_RESPONSE" | tail -n1)

if assert_response "Export to CSV" 200 "$EXPORT_STATUS" "$EXPORT_BODY"; then
    EXPORT_ID=$(echo "$EXPORT_BODY" | grep -o '"export_id":"[^"]*' | cut -d'"' -f4)
    echo "  Export ID: $EXPORT_ID"
fi
echo ""

# Test 9: Get export status
echo "9. Testing get export status..."
if [ -n "$EXPORT_ID" ]; then
    sleep 2  # Wait for export to process

    EXPORT_STATUS_RESPONSE=$(curl -s -w "\n%{http_code}" -X GET "$API_BASE_URL/api/v1/data-explorer/exports/$EXPORT_ID" \
        -H "Authorization: Bearer $ACCESS_TOKEN")

    EXPORT_STATUS_BODY=$(echo "$EXPORT_STATUS_RESPONSE" | sed '$d')
    EXPORT_STATUS_CODE=$(echo "$EXPORT_STATUS_RESPONSE" | tail -n1)

    if assert_response "Get export status" 200 "$EXPORT_STATUS_CODE" "$EXPORT_STATUS_BODY"; then
        STATUS=$(echo "$EXPORT_STATUS_BODY" | grep -o '"status":"[^"]*' | cut -d'"' -f4)
        echo "  Export status: $STATUS"
    fi
else
    echo "✗ SKIPPED (no export ID)"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
echo ""

# Test 10: Aggregate query (count)
echo "10. Testing aggregate query..."
AGGREGATE_RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$API_BASE_URL/api/v1/data-explorer/aggregate" \
    -H "Authorization: Bearer $ACCESS_TOKEN" \
    -H "Content-Type: application/json" \
    -d '{
        "object": "contacts",
        "aggregations": [
            {
                "function": "count",
                "field": "id"
            },
            {
                "function": "count",
                "field": "industry",
                "group_by": true
            }
        ]
    }')

AGGREGATE_BODY=$(echo "$AGGREGATE_RESPONSE" | sed '$d')
AGGREGATE_STATUS=$(echo "$AGGREGATE_RESPONSE" | tail -n1)

assert_response "Aggregate query" 200 "$AGGREGATE_STATUS" "$AGGREGATE_BODY"
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
