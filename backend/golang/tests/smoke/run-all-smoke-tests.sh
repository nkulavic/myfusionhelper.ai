#!/bin/bash

# Master Smoke Test Runner
# Runs all smoke tests and provides summary report

set -e

# Configuration
API_BASE_URL="${API_BASE_URL:-http://localhost:8080}"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

echo "========================================"
echo "MyFusionHelper.ai - Smoke Test Suite"
echo "========================================"
echo "API: $API_BASE_URL"
echo "Date: $(date)"
echo ""

# Track overall results
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0
SUITE_START_TIME=$(date +%s)

# Array to store test results
declare -a TEST_RESULTS

# Function to run a single test suite
run_test_suite() {
    local test_script=$1
    local test_name=$(basename "$test_script" .sh)

    echo ""
    echo "========================================="
    echo "Running: $test_name"
    echo "========================================="

    SUITE_START=$(date +%s)

    # Run the test and capture exit code
    set +e
    "$test_script"
    TEST_EXIT_CODE=$?
    set -e

    SUITE_END=$(date +%s)
    SUITE_DURATION=$((SUITE_END - SUITE_START))

    # Store result
    if [ $TEST_EXIT_CODE -eq 0 ]; then
        TEST_RESULTS+=("✓ PASSED - $test_name (${SUITE_DURATION}s)")
        PASSED_TESTS=$((PASSED_TESTS + 1))
    else
        TEST_RESULTS+=("✗ FAILED - $test_name (${SUITE_DURATION}s)")
        FAILED_TESTS=$((FAILED_TESTS + 1))
    fi

    TOTAL_TESTS=$((TOTAL_TESTS + 1))
}

# Check if API is reachable
echo "Checking API connectivity..."
if curl -s -f -o /dev/null "$API_BASE_URL/health" 2>/dev/null; then
    echo "✓ API is reachable at $API_BASE_URL"
else
    echo "✗ WARNING: API may not be reachable at $API_BASE_URL"
    echo "  Tests may fail. Press Ctrl+C to abort or wait 5 seconds to continue..."
    sleep 5
fi
echo ""

# Run all test suites in order
echo "Starting test execution..."
echo ""

run_test_suite "$SCRIPT_DIR/test-auth.sh"
run_test_suite "$SCRIPT_DIR/test-onboarding.sh"
run_test_suite "$SCRIPT_DIR/test-crm-connection.sh"
run_test_suite "$SCRIPT_DIR/test-helper-execution.sh"
run_test_suite "$SCRIPT_DIR/test-billing.sh"
run_test_suite "$SCRIPT_DIR/test-data-explorer.sh"
run_test_suite "$SCRIPT_DIR/test-api-keys.sh"

# Calculate total duration
SUITE_END_TIME=$(date +%s)
TOTAL_DURATION=$((SUITE_END_TIME - SUITE_START_TIME))
TOTAL_MINUTES=$((TOTAL_DURATION / 60))
TOTAL_SECONDS=$((TOTAL_DURATION % 60))

# Print summary
echo ""
echo ""
echo "========================================="
echo "SMOKE TEST SUITE SUMMARY"
echo "========================================="
echo "Date: $(date)"
echo "API: $API_BASE_URL"
echo "Duration: ${TOTAL_MINUTES}m ${TOTAL_SECONDS}s"
echo ""
echo "Test Suites:"
for result in "${TEST_RESULTS[@]}"; do
    echo "  $result"
done
echo ""
echo "========================================="
echo "Total Test Suites: $TOTAL_TESTS"
echo "Passed: $PASSED_TESTS"
echo "Failed: $FAILED_TESTS"
echo "Success Rate: $(awk "BEGIN {printf \"%.1f\", ($PASSED_TESTS/$TOTAL_TESTS)*100}")%"
echo "========================================="
echo ""

# Exit with appropriate code
if [ $FAILED_TESTS -eq 0 ]; then
    echo "✓ ALL SMOKE TESTS PASSED!"
    echo ""
    echo "The system is ready for deployment."
    exit 0
else
    echo "✗ SOME TESTS FAILED"
    echo ""
    echo "Please review the failed tests above before deploying."
    exit 1
fi
