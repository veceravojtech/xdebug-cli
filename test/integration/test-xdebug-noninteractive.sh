#!/bin/bash
# Non-Interactive xdebug-cli Test Script

TEST_URL="http://booking.previo.loc/coupon/index/select/?hotId=731541&currency=CZK&lang=cs&redirectType=iframe&showTabs=coupon&PHPSESSID=8fc0edd2942ff8140966ecee51c6114c&helpCampaign=0&XDEBUG_TRIGGER=1"
BREAKPOINT="booking/application/modules/default/models/Controller/Action/Helper/AllowedTabs.php:144"

# Function to run a test
run_test() {
  local test_name="$1"
  shift

  # Check if first argument is a URL (starts with http)
  local test_url="$TEST_URL"
  if [[ "$1" == http* ]]; then
    test_url="$1"
    shift
  fi

  local commands=("$@")

  echo "=========================================="
  echo "$test_name"
  echo "=========================================="

  # Create temporary file for output
  local output_file=$(mktemp)

  # Build --commands arguments (each command needs its own --commands flag)
  local cmd_args=()
  for cmd in "${commands[@]}"; do
    cmd_args+=(--commands "$cmd")
  done

  # Start xdebug-cli in background (listen on 0.0.0.0 for Docker containers)
  xdebug-cli listen -l 0.0.0.0 --force "${cmd_args[@]}" > "$output_file" 2>&1 &
  XDEBUG_PID=$!

  # Give it a moment to start listening
  sleep 0.5

  # Trigger the PHP request (will connect to xdebug-cli)
  curl -s "$test_url" > /dev/null 2>&1 &
  CURL_PID=$!

  # Wait for xdebug-cli to complete (or timeout after 10s)
  for i in {1..100}; do
    if ! ps -p $XDEBUG_PID > /dev/null 2>&1; then
      wait $XDEBUG_PID 2>/dev/null
      EXIT_CODE=$?
      break
    fi
    sleep 0.1
  done

  # If still running after 10 seconds, consider it timed out
  if ps -p $XDEBUG_PID > /dev/null 2>&1; then
    EXIT_CODE=124
  fi

  # Kill curl if still running
  kill $CURL_PID 2>/dev/null || true

  # Kill xdebug-cli if still running
  if ps -p $XDEBUG_PID > /dev/null 2>&1; then
    kill $XDEBUG_PID 2>/dev/null || true
    sleep 0.5
  fi

  # Display captured output
  echo ""
  echo "--- Output ---"
  cat "$output_file"
  echo "--- End Output ---"
  echo ""

  # Clean up temporary file
  rm -f "$output_file"

  if [ $EXIT_CODE -eq 0 ]; then
    echo "✓ Test passed"
  else
    echo "✗ Test failed (exit code: $EXIT_CODE)"
  fi

  echo ""

  return $EXIT_CODE
}

# Run tests sequentially
run_test "Test 1: Basic breakpoint" \
  "break $BREAKPOINT" "run"

run_test "Test 2: Variable inspection" \
  "break $BREAKPOINT" "run" "print \$this->_allowedTabs"

run_test "Test 3: JSON output" \
  "break $BREAKPOINT" "run" "context local" "print \$this->_allowedTabs"

run_test "Test 4: Stack trace" \
  "break $BREAKPOINT" "run" "info stack"

# Test 5: 10-Step Debugging Workflow - Tracing tab name processing bug
# This scenario demonstrates real debugging by stepping through the _parse() method
# to discover why tab names aren't being properly normalized.
#
# Bug: The array_walk on line 170 doesn't use &$tab reference, so trim/strtolower
#      don't actually modify the array elements.
#
# Target: AllowedTabs.php _parse() method (lines 161-184)
# Trigger: Use showTabs parameter to invoke _parse()
PARSE_URL="http://booking.previo.loc/coupon/index/select/?hotId=731541&currency=CZK&lang=cs&redirectType=iframe&showTabs=stay-hotels&PHPSESSID=8fc0edd2942ff8140966ecee51c6114c&helpCampaign=0&XDEBUG_TRIGGER=1"
PARSE_BREAKPOINT="booking/application/modules/default/models/Controller/Action/Helper/AllowedTabs.php:168"

run_test "Test 5: Step-Through Debugging (10 steps - Tab Name Processing Bug)" \
  "$PARSE_URL" \
  "break $PARSE_BREAKPOINT" \
  "run" \
  "print \$param" \
  "step" \
  "print \$tabNames" \
  "next" \
  "print \$tabNames" \
  "next" \
  "next" \
  "next" \
  "print \$tabName" \
  "print \$tabs"

echo "=========================================="
echo "All tests completed"
echo "=========================================="
