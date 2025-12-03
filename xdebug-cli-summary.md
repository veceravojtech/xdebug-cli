# xdebug-cli Issues Summary

## Environment
- **Xdebug Config**: `client_host=127.17.0.1`, `port=9003`, `mode=debug,develop,trace`, `start_with_request=trigger`
- **PHP Path in Container**: `/home/users/previo/current/`

## Problems Encountered

### 1. Daemon Mode Instability
```bash
xdebug-cli listen --daemon --force --commands "break /path:919"
# Daemon starts but exits before attach can connect
```
**Error**: `no daemon running on port 9003`

### 2. Wrong File Path
```bash
# Local path (wrong):
break /var/www/html/booking/...

# Container path (correct):
break /home/users/previo/current/booking/...
```

### 3. Port Binding Race
```bash
# After killing daemon, port stays bound
Error: listen tcp 0.0.0.0:9003: bind: address already in use
```

### 4. Command Sequence Issue
```bash
xdebug-cli listen --commands "break ..." "run" "stack" "detach"
# Only breakpoint output captured, run/stack never execute
```
Breakpoint sets successfully but `run` continues to completion without pausing.

### 5. Invalid Command
```bash
xdebug-cli listen --commands "step_into" ...
# Error: Unknown command: step_into
```
Correct command is `step`, not `step_into`.

## Working Approach

```bash
# 1. Add temporary debug code
file_put_contents('/tmp/debug_stack.txt', (new \Exception())->getTraceAsString());

# 2. Trigger request
curl 'http://booking.previo.loc/index/get-reservation-price/' ...

# 3. Read from container
docker exec previo-previo_php-1 cat /tmp/debug_stack.txt
```

## Successful xdebug-cli Pattern

```bash
# Step command works - stops at first line
xdebug-cli listen --commands "step" "stack" "detach" &
sleep 2
curl ... -b 'XDEBUG_SESSION=PHPSTORM' ...
wait
```
**Output**: `Breakpoint hit at file:///home/users/previo/current/booking/www/index.php:4`

## Recommendations

1. Use `step` for initial stop, not breakpoints
2. Always use container paths (`/home/users/previo/current/...`)
3. Add `sleep 2-3` between listener start and curl trigger
4. For complex debugging, use temporary `Exception()->getTraceAsString()` method
5. Kill stale processes: `pkill -f xdebug-cli; sleep 2` before retrying
