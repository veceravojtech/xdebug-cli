# Add Timeout Exit Feedback

## Why

When breakpoint validation timeout occurs, the daemon exits silently with code 1. Users then see "no daemon running" on the next `attach` command with no indication of what happened.

This makes it difficult to diagnose whether:
- The breakpoint path was wrong
- The code path wasn't exercised
- The timeout was simply too short
- Some other error occurred

## What Changes

- Print warning message to stderr when breakpoint timeout occurs
- Include timeout duration and hint about `--breakpoint-timeout` flag
- Write timeout event to a log file for post-mortem analysis
- Exit with distinct code (124, matching Unix timeout convention)

## Impact

- Affected specs: `cli` (Daemon Start Command requirement)
- Affected code: `internal/cli/daemon.go` (breakpoint validation section)
- No breaking changes - adds user feedback only
- Improves debuggability of timeout scenarios
