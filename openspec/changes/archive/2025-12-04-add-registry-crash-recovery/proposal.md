# Add Registry Crash Recovery

## Why

When daemon crashes (SIGSEGV, SIGKILL, OOM), the registry cleanup in `Shutdown()` is never executed. This leaves:
1. Stale entries in `~/.xdebug-cli/sessions.json`
2. Orphaned socket files in `/tmp/`
3. PID files pointing to dead or recycled processes

This causes "daemon already running" errors when trying to start a new daemon, even though no daemon is actually running.

## What Changes

- Add registry validation on daemon start: verify each entry's process exists AND is xdebug-cli
- Remove stale entries before attempting to start new daemon
- Check if PID has been recycled to different process using `/proc/<pid>/comm`
- Clean up orphaned socket and PID files

## Impact

- Affected specs: `persistent-sessions` (Session Registry requirement)
- Affected code: `internal/daemon/daemon.go`, `internal/daemon/registry.go`
- No breaking changes - improved robustness
- Enables clean recovery after crashes without manual intervention
