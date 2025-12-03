#!/bin/bash
# Integration Test: --curl flag for daemon start
# Tests that --curl flag properly triggers HTTP requests with XDEBUG_TRIGGER cookie

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m' # No Color

# Test counter
TESTS_PASSED=0
TESTS_FAILED=0

pass() {
    echo -e "${GREEN}PASS${NC}: $1"
    TESTS_PASSED=$((TESTS_PASSED + 1))
}

fail() {
    echo -e "${RED}FAIL${NC}: $1"
    TESTS_FAILED=$((TESTS_FAILED + 1))
}

# Cleanup function
cleanup() {
    xdebug-cli daemon kill --all --force >/dev/null 2>&1 || true
}

trap cleanup EXIT

echo "=== Testing --curl flag functionality ==="
echo

# Test 1: Missing --curl flag should show error
echo "Test 1: Missing --curl flag shows error"
OUTPUT=$(xdebug-cli daemon start 2>&1 || true)
if echo "$OUTPUT" | grep -q "\-\-curl flag is required"; then
    pass "Missing --curl flag shows proper error message"
else
    fail "Missing --curl flag should show error message"
    echo "Got: $OUTPUT"
fi

# Test 2: Missing --curl flag shows usage examples
echo "Test 2: Error message includes usage examples"
if echo "$OUTPUT" | grep -q "xdebug-cli daemon start --curl"; then
    pass "Error message includes usage examples"
else
    fail "Error message should include usage examples"
fi

# Test 3: Error message mentions XDEBUG_TRIGGER
echo "Test 3: Error message mentions XDEBUG_TRIGGER"
if echo "$OUTPUT" | grep -q "XDEBUG_TRIGGER"; then
    pass "Error message mentions XDEBUG_TRIGGER is added automatically"
else
    fail "Error message should mention XDEBUG_TRIGGER"
fi

# Test 4: --curl with invalid URL (curl fails) should terminate daemon
echo "Test 4: Curl failure terminates daemon"
# Use a URL that curl will fail to connect to
xdebug-cli daemon start --curl "http://localhost:59999/nonexistent" 2>&1 &
sleep 2

# Check if daemon is still running (it shouldn't be if curl failed)
if ! xdebug-cli daemon isAlive 2>/dev/null; then
    pass "Daemon terminated after curl failure"
else
    fail "Daemon should terminate when curl fails"
    xdebug-cli daemon kill --force 2>/dev/null || true
fi

# Summary
echo
echo "=== Test Summary ==="
echo "Passed: $TESTS_PASSED"
echo "Failed: $TESTS_FAILED"

if [ $TESTS_FAILED -gt 0 ]; then
    exit 1
fi
