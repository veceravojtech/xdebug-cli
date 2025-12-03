# Change: Add breakpoint validation with timeout - fail fast on bad paths

## Why
Users often specify breakpoints with just filename:line (e.g., `break PriceLoader.php:369`) instead of a full path. Xdebug requires file:// URIs with absolute paths to properly match breakpoints. Breakpoints with relative or filename-only paths silently fail to trigger, leaving a useless daemon running and causing confusion.

## What Changes
- When `daemon start` is called with `--commands "break ..."` containing a non-absolute path:
  1. Show warning about non-absolute path with suggestion from history (if available)
  2. Start daemon and set the breakpoint
  3. Wait for breakpoint to be hit (with configurable timeout, default 10s)
  4. If breakpoint hits within timeout: daemon continues normally
  5. If timeout expires without hit: terminate daemon and exit with error
- Store successfully hit breakpoint paths for future suggestions
- Error message includes the suggested full path for easy retry

## Impact
- Affected specs: cli, persistent-sessions
- Affected code:
  - `internal/daemon/daemon.go` (breakpoint validation on start)
  - `internal/daemon/executor.go` (path validation, timeout handling)
  - `internal/dbgp/session.go` (store breakpoint paths)
