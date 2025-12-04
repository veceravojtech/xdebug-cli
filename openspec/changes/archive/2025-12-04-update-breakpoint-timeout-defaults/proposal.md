# Update Breakpoint Timeout Defaults

## Why

The current 10-second breakpoint timeout is too short for many real-world scenarios:
- Cold PHP opcache requires 5-10 seconds for first compilation
- Framework bootstrap (Laravel, Symfony) takes 2-3 seconds
- Database connection pools need warm-up time
- Deep code paths may not hit breakpoint quickly

When timeout occurs, daemon exits silently with code 1, leaving users confused about what happened.

## What Changes

- Increase default `--breakpoint-timeout` from 10 to 30 seconds
- Add `--wait-forever` convenience flag (equivalent to `--breakpoint-timeout 0`)
- Document timeout behavior in help text

## Impact

- Affected specs: `cfg` (CLI Parameters requirement)
- Affected code: `internal/cli/daemon.go` (flag definition)
- No breaking changes - existing `--breakpoint-timeout` flag behavior unchanged
- Reduces false-positive "daemon not running" errors
