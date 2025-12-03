#!/bin/bash
# Integration test script for the attach command
# This script demonstrates how to test the attach command with a running daemon

set -e

echo "=== Attach Command Integration Test ==="
echo ""

# Build the binary
echo "Building xdebug-cli..."
go build -o xdebug-cli ./cmd/xdebug-cli
echo "✓ Build successful"
echo ""

# Test 1: Try to attach without a daemon
echo "Test 1: Attach without daemon (should fail)"
if ./xdebug-cli attach --commands "run" 2>&1 | grep -q "no daemon running"; then
    echo "✓ Correct error message when no daemon is running"
else
    echo "✗ Expected 'no daemon running' error"
    exit 1
fi
echo ""

# Test 2: Check help output
echo "Test 2: Verify attach command help"
if ./xdebug-cli attach --help 2>&1 | grep -q "Attach to an existing daemon"; then
    echo "✓ Help text is correct"
else
    echo "✗ Help text is missing or incorrect"
    exit 1
fi
echo ""

# Test 3: Verify command is registered
echo "Test 3: Verify attach command is listed"
if ./xdebug-cli --help 2>&1 | grep -q "attach.*Attach to a running daemon"; then
    echo "✓ Attach command is registered in command list"
else
    echo "✗ Attach command not found in command list"
    exit 1
fi
echo ""

# Test 4: Test --commands flag validation
echo "Test 4: Verify --commands flag is required"
if ./xdebug-cli attach 2>&1 | grep -q "commands.*required"; then
    echo "✓ Correctly requires --commands flag"
else
    echo "✗ Should require --commands flag"
    exit 1
fi
echo ""

# Test 5: Test different port
echo "Test 5: Test with custom port"
if ./xdebug-cli attach --port 9999 --commands "run" 2>&1 | grep -q "no daemon running on port 9999"; then
    echo "✓ Correctly uses custom port"
else
    echo "✗ Port flag not working correctly"
    exit 1
fi
echo ""

echo "=== All basic tests passed ==="
echo ""
echo "Note: Full integration tests require a running daemon."
echo "To test with a real daemon:"
echo "  1. Start daemon: xdebug-cli daemon --port 9003"
echo "  2. Start PHP app with Xdebug enabled"
echo "  3. Run: xdebug-cli attach --commands 'run' 'context local'"
echo ""
