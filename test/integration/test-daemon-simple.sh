#!/bin/bash
# Simple Daemon Test - Exact workflow

# Install latest version
/home/console/PhpstormProjects/CLI/xdebug-cli/install.sh

BREAKPOINT="booking/application/modules/default/models/Controller/Action/Helper/AllowedTabs.php:144"

# Start daemon with breakpoint
xdebug-cli daemon start --commands "break $BREAKPOINT"

# Wait 1 second
sleep 1

# Trigger curl in background
curl "http://booking.previo.loc/coupon/index/select/?hotId=731541&currency=CZK&lang=cs&redirectType=iframe&showTabs=stay-hotels&PHPSESSID=8fc0edd2942ff8140966ecee51c6114c&helpCampaign=0&XDEBUG_TRIGGER=1" > /dev/null 2>&1 &

# Wait 1 second
sleep 1

xdebug-cli attach --json --commands "info stack" 


# Attach and execute context
xdebug-cli attach --commands "print \$this->_allowedTabs"

# Attach and execute context
xdebug-cli attach --json --commands "print \$this->_allowedTabs"

# Attach and execute context
xdebug-cli attach --commands "context"

#kill it
xdebug-cli connection kill