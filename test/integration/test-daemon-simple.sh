#!/bin/bash
# Simple Daemon Test - Exact workflow

# Install latest version
/home/console/PhpstormProjects/CLI/xdebug-cli/install.sh

BREAKPOINT="booking/application/modules/default/models/Controller/Action/Helper/AllowedTabs.php:144"
CURL_URL="http://booking.previo.loc/coupon/index/select/?hotId=731541&currency=CZK&lang=cs&redirectType=iframe&showTabs=stay-hotels&PHPSESSID=8fc0edd2942ff8140966ecee51c6114c&helpCampaign=0"

# Start daemon with curl trigger and breakpoint
# (auto-kills all existing daemons, no race condition!)
xdebug-cli daemon start --curl "$CURL_URL" --commands "break $BREAKPOINT"

# Wait for connection to establish
sleep 1

xdebug-cli attach --commands "info stack" 
echo "\n"
echo "\n"

# Attach and execute context
xdebug-cli attach --commands "print \$this->_allowedTabs"
echo "\n"
echo "\n"

# Attach and execute context
xdebug-cli attach --commands "context local"
echo "\n"
echo "\n"


xdebug-cli attach --commands "breakpoint_list"
echo "\n"
echo "\n"

xdebug-cli attach --commands "property_get -n \$this->_allowedTabs"
echo "\n"
echo "\n"

xdebug-cli attach --commands "over"
echo "\n"
echo "\n"

xdebug-cli attach --commands "into"
echo "\n"
echo "\n"

xdebug-cli attach --commands "step_out"
echo "\n"
echo "\n"

xdebug-cli attach --commands "step_into"
echo "\n"
echo "\n"

xdebug-cli attach --commands "breakpoint_remove"
echo "\n"
echo "\n"

xdebug-cli attach --commands "clear $BREAKPOINT"
echo "\n"
echo "\n"

xdebug-cli attach --commands "continue"
echo "\n"
echo "\n"