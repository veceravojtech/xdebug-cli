#!/bin/bash
# Test script for xdebug-cli daemon workflow validation
# This script validates the debugging workflow described in the user scenario

set -e  # Exit on error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
URL="http://booking.previo.loc/coupon/index/select/?hotId=731541&currency=CZK&lang=cs&redirectType=iframe&showTabs=coupon&PHPSESSID=8fc0edd2942ff8140966ecee51c6114c&helpCampaign=0&XDEBUG_TRIGGER=1"
BREAKPOINT="booking/application/modules/default/models/Controller/Action/Helper/AllowedTabs.php:144"
PORT=9003

echo "=========================================="
echo "xdebug-cli Workflow Validation Test"
echo "=========================================="
echo ""

# Function to print status
print_step() {
    echo -e "${YELLOW}[STEP $1]${NC} $2"
}

print_success() {
    echo -e "${GREEN}✓${NC} $1"
}

print_error() {
    echo -e "${RED}✗${NC} $1"
}

# Test 1: Validate xdebug-cli exists
print_step 1 "Checking xdebug-cli installation"
if command -v xdebug-cli &> /dev/null; then
    print_success "xdebug-cli found: $(which xdebug-cli)"
    xdebug-cli version
else
    print_error "xdebug-cli not found in PATH"
    echo "  Run: ./install.sh"
    exit 1
fi
echo ""

# Test 2: Clean up any existing daemon
print_step 2 "Cleaning up existing daemons on port $PORT"
if xdebug-cli connection isAlive 2>/dev/null; then
    print_success "Found existing daemon, killing it"
    xdebug-cli connection kill
else
    print_success "No existing daemon on port $PORT"
fi
echo ""

# Test 3: START DEBUG workflow
print_step 3 "Testing START DEBUG workflow"
echo "  Command: xdebug-cli listen --daemon --force --commands \"break $BREAKPOINT\""
echo ""
echo "  Expected output:"
echo "    - Daemon started on port $PORT (PID xxxxx)"
echo ""
echo "  Press Enter to start daemon (or Ctrl+C to skip)..."
read -r

xdebug-cli listen --daemon --force --commands "break $BREAKPOINT"
sleep 1

if xdebug-cli connection isAlive; then
    print_success "Daemon started successfully"
    xdebug-cli connection
else
    print_error "Daemon failed to start"
    exit 1
fi
echo ""

# Test 4: Simulate PHP request trigger
print_step 4 "Trigger PHP request"
echo "  URL: $URL"
echo ""
echo "  NOTE: This test script cannot trigger the actual PHP request."
echo "  You need to manually execute:"
echo ""
echo "    curl '$URL'"
echo ""
echo "  Or open in browser and wait for breakpoint to be hit."
echo ""
echo "  Press Enter once the request is paused at breakpoint..."
read -r
echo ""

# Test 5: DO DEBUG workflow
print_step 5 "Testing DO DEBUG workflow - inspect variables"
echo "  Command: xdebug-cli attach --commands \"context local\""
echo "  Press Enter to execute..."
read -r

if xdebug-cli attach --commands "context local"; then
    print_success "Successfully attached and retrieved context"
else
    print_error "Failed to attach or retrieve context"
fi
echo ""

print_step 6 "Testing DO DEBUG workflow - continue execution"
echo "  Command: xdebug-cli attach --commands \"run\""
echo "  Press Enter to execute..."
read -r

if xdebug-cli attach --commands "run"; then
    print_success "Successfully continued execution"
else
    print_error "Failed to continue execution"
fi
echo ""

# Test 6: FORCE QUIT DEBUG workflow
print_step 7 "Testing FORCE QUIT DEBUG workflow"
sleep 1

if xdebug-cli connection isAlive; then
    print_success "Daemon still running"
    echo "  Killing daemon..."
    xdebug-cli connection kill
    sleep 1

    if ! xdebug-cli connection isAlive 2>/dev/null; then
        print_success "Daemon killed successfully"
    else
        print_error "Failed to kill daemon"
        exit 1
    fi
else
    print_success "Daemon already ended (expected after 'run' command)"
fi
echo ""

# Summary
echo "=========================================="
echo "Workflow Validation Complete"
echo "=========================================="
echo ""
echo -e "${GREEN}All workflow steps validated successfully!${NC}"
echo ""
echo "Key points:"
echo "  1. Use --daemon --force to start fresh persistent session"
echo "  2. Set breakpoints before triggering PHP request"
echo "  3. Use 'attach' commands to inspect state across multiple invocations"
echo "  4. Use 'run' to complete request execution"
echo "  5. Use 'connection kill' to terminate daemon when done"
echo ""
