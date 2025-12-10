#!/bin/bash
# Integration Test: Breakpoint path warning validation
# Tests that non-absolute breakpoint paths show warnings and fail-fast behavior

# Use gtimeout on macOS if timeout is not available
if ! command -v timeout &> /dev/null && command -v gtimeout &> /dev/null; then
    timeout() { gtimeout "$@"; }
fi

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
    rm -f /tmp/test-breakpoint-warning.php
}

trap cleanup EXIT

echo "=== Testing breakpoint path warning functionality ==="
echo

# Test 1: Non-absolute breakpoint path shows warning
echo "Test 1: Non-absolute breakpoint path shows warning"
OUTPUT=$(timeout 5 xdebug-cli daemon start --curl "http://localhost:59999/nonexistent" --commands "break file.php:100" 2>&1 || true)
if echo "$OUTPUT" | grep -q "Warning: breakpoint path 'file.php:100' is not absolute"; then
    pass "Non-absolute path warning displayed"
else
    fail "Should show warning for non-absolute path"
    echo "Got: $OUTPUT"
fi

# Test 2: Warning mentions Xdebug requires absolute paths
echo "Test 2: Warning mentions Xdebug requires absolute paths"
if echo "$OUTPUT" | grep -q "Xdebug requires absolute paths"; then
    pass "Warning explains Xdebug requirement"
else
    fail "Warning should explain Xdebug requirement"
    echo "Got: $OUTPUT"
fi

# Test 3: Warning mentions timeout
echo "Test 3: Warning mentions timeout for breakpoint hit"
if echo "$OUTPUT" | grep -q "Will wait.*seconds for breakpoint hit"; then
    pass "Warning mentions timeout"
else
    fail "Warning should mention timeout"
    echo "Got: $OUTPUT"
fi

# Test 4: Absolute path does NOT show warning
echo "Test 4: Absolute path does not show warning"
OUTPUT=$(timeout 5 xdebug-cli daemon start --curl "http://localhost:59999/nonexistent" --commands "break /var/www/file.php:100" 2>&1 || true)
if ! echo "$OUTPUT" | grep -q "Warning: breakpoint path"; then
    pass "No warning for absolute path"
else
    fail "Should not warn for absolute path"
    echo "Got: $OUTPUT"
fi

# Test 5: :line format does NOT show warning (uses current file)
echo "Test 5: :line format does not show warning"
OUTPUT=$(timeout 5 xdebug-cli daemon start --curl "http://localhost:59999/nonexistent" --commands "break :42" 2>&1 || true)
if ! echo "$OUTPUT" | grep -q "Warning: breakpoint path"; then
    pass "No warning for :line format"
else
    fail "Should not warn for :line format"
    echo "Got: $OUTPUT"
fi

# Test 6: "break call" does NOT show warning
echo "Test 6: break call does not show warning"
OUTPUT=$(timeout 5 xdebug-cli daemon start --curl "http://localhost:59999/nonexistent" --commands "break call myFunction" 2>&1 || true)
if ! echo "$OUTPUT" | grep -q "Warning: breakpoint path"; then
    pass "No warning for break call"
else
    fail "Should not warn for break call"
    echo "Got: $OUTPUT"
fi

# Test 7: "break exception" does NOT show warning
echo "Test 7: break exception does not show warning"
OUTPUT=$(timeout 5 xdebug-cli daemon start --curl "http://localhost:59999/nonexistent" --commands "break exception" 2>&1 || true)
if ! echo "$OUTPUT" | grep -q "Warning: breakpoint path"; then
    pass "No warning for break exception"
else
    fail "Should not warn for break exception"
    echo "Got: $OUTPUT"
fi

# Test 8: Short form "b" also triggers warning
echo "Test 8: Short form 'b' triggers warning"
OUTPUT=$(timeout 5 xdebug-cli daemon start --curl "http://localhost:59999/nonexistent" --commands "b myfile.php:50" 2>&1 || true)
if echo "$OUTPUT" | grep -q "Warning: breakpoint path 'myfile.php:50' is not absolute"; then
    pass "Short form 'b' also triggers warning"
else
    fail "Short form 'b' should also trigger warning"
    echo "Got: $OUTPUT"
fi

# Test 9: --breakpoint-timeout=0 disables timeout validation
echo "Test 9: --breakpoint-timeout=0 disables timeout message"
OUTPUT=$(timeout 5 xdebug-cli daemon start --curl "http://localhost:59999/nonexistent" --commands "break file.php:100" --breakpoint-timeout=0 2>&1 || true)
if ! echo "$OUTPUT" | grep -q "Will wait.*seconds for breakpoint hit"; then
    pass "--breakpoint-timeout=0 disables timeout message"
else
    fail "--breakpoint-timeout=0 should disable timeout message"
    echo "Got: $OUTPUT"
fi

# Summary
echo
echo "=== Test Summary ==="
echo "Passed: $TESTS_PASSED"
echo "Failed: $TESTS_FAILED"

if [ $TESTS_FAILED -gt 0 ]; then
    exit 1
fi

echo "All tests passed!"
exit 0
