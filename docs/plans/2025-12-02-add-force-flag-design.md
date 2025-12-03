# Design: Add --force Flag to listen Command

**Date**: 2025-12-02
**Status**: Approved

## Overview

Add a `--force` flag to `xdebug-cli listen --non-interactive` that automatically kills any existing daemon session on the same port before starting the non-interactive session. This ensures clean state for automation scenarios where stale daemon processes may exist.

## Problem Statement

When running `listen --non-interactive` in automation scripts or CI/CD, old daemon sessions from previous runs may still be active on the target port. Currently, users must manually detect and kill these daemons before starting new sessions. The `--force` flag automates this cleanup.

## Goals

- Provide automatic cleanup of stale daemon sessions on the same port
- Never fail - always continue after cleanup attempt
- Show clear feedback about cleanup actions (killed daemon or no daemon found)
- Work with both `--non-interactive` and `--daemon` modes

## Behavior Specification

### Core Behavior

When `--force` is used with `--non-interactive`:
1. Check if a daemon exists on the target port (default 9003 or custom via `-p`)
2. If daemon exists: kill it, show confirmation message, proceed with session
3. If no daemon exists: show warning message, proceed with session
4. Never fail - always continue to start the non-interactive session

### Usage Examples

```bash
# Clean slate: kill any daemon on port 9003, then run commands
xdebug-cli listen --non-interactive --force --commands "run" "print \$x"

# With custom port
xdebug-cli listen -p 9004 --non-interactive --force --commands "break :42" "run"

# Daemon mode with force (kill old daemon before forking new one)
xdebug-cli listen --daemon --force --commands "break file.php:100"
```

### Output Messages

**Daemon killed successfully**:
```
Killed daemon on port 9003 (PID 12345)
Server listening on 127.0.0.1:9003
```

**No daemon running**:
```
Warning: no daemon running on port 9003
Server listening on 127.0.0.1:9003
```

**Kill failed (permission denied)**:
```
Error: failed to kill daemon on port 9003 (PID 12345): permission denied
Continuing anyway...
Server listening on 127.0.0.1:9003
```

## Implementation Plan

### 1. Flag Addition

**File**: `internal/cli/listen.go`

Add flag to `listenCmd.Flags()`:
```go
listenCmd.Flags().BoolVar(&CLIArgs.Force, "force", false, "Kill existing daemon on same port before starting")
```

**File**: `internal/cfg/config.go` (or wherever CLIParameter is defined)

Add field to `CLIParameter`:
```go
type CLIParameter struct {
    // ... existing fields ...
    Force bool
}
```

### 2. Cleanup Logic

**File**: `internal/cli/listen.go`

Create new helper function:
```go
// killDaemonOnPort attempts to kill any daemon running on the specified port.
// Always returns nil (never fails) - shows warnings/errors but continues.
func killDaemonOnPort(port int) error {
    registry, err := daemon.NewSessionRegistry()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Warning: failed to load session registry: %v\n", err)
        return nil
    }

    session, err := registry.Get(port)
    if err != nil {
        // No session found - just warn and continue
        fmt.Fprintf(os.Stderr, "Warning: no daemon running on port %d\n", port)
        return nil
    }

    // Check if process exists (handle stale registry entries)
    if !daemon.ProcessExists(session.PID) {
        fmt.Fprintf(os.Stderr, "Warning: daemon on port %d is stale (PID %d no longer exists), cleaning up\n", port, session.PID)
        registry.Remove(port)
        return nil
    }

    // Kill the process
    process, err := os.FindProcess(session.PID)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Warning: failed to find daemon process (PID %d): %v\n", session.PID, err)
        return nil
    }

    if err := process.Kill(); err != nil {
        fmt.Fprintf(os.Stderr, "Error: failed to kill daemon on port %d (PID %d): %v\nContinuing anyway...\n", port, session.PID, err)
        return nil
    }

    // Clean up registry
    registry.Remove(port)
    fmt.Printf("Killed daemon on port %d (PID %d)\n", port, session.PID)
    return nil
}
```

Modify `runListeningCmd()`:
```go
func runListeningCmd() error {
    v := view.NewView()

    // Validate flags
    if len(CLIArgs.Commands) > 0 && !CLIArgs.NonInteractive && !CLIArgs.Daemon {
        fmt.Fprintf(os.Stderr, "Error: --commands requires --non-interactive or --daemon flag\n")
        os.Exit(1)
    }

    // NEW: Validate --force flag
    if CLIArgs.Force && !CLIArgs.NonInteractive && !CLIArgs.Daemon {
        fmt.Fprintf(os.Stderr, "Error: --force requires --non-interactive or --daemon flag\n")
        os.Exit(1)
    }

    // NEW: Kill existing daemon if --force is set
    if CLIArgs.Force {
        killDaemonOnPort(CLIArgs.Port)
    }

    // Display startup information only in interactive mode
    if !CLIArgs.NonInteractive && !CLIArgs.Daemon {
        v.PrintApplicationInformation(cfg.Version, CLIArgs.Host, CLIArgs.Port)
    }

    // ... rest of existing code ...
}
```

### 3. Expose ProcessExists

**File**: `internal/daemon/registry.go`

Change `processExists` from private to public:
```go
// ProcessExists checks if a process with the given PID exists
func ProcessExists(pid int) bool {
    // Check if /proc/<pid> exists (Linux-specific)
    procPath := fmt.Sprintf("/proc/%d", pid)
    _, err := os.Stat(procPath)
    return err == nil
}
```

Update call site in `cleanupStale()`:
```go
if ProcessExists(s.PID) {
    activeSessions = append(activeSessions, s)
}
```

## Edge Cases Handled

### 1. Stale Registry Entry
**Scenario**: Registry says daemon exists but process is dead
**Solution**: Check `ProcessExists(pid)` before kill attempt, clean up registry entry

### 2. Permission Denied
**Scenario**: Can't kill process (different user owns it)
**Solution**: Show error message but continue with session start

### 3. Race Condition
**Scenario**: Daemon dies between registry check and kill
**Solution**: Ignore "process not found" errors, clean up registry, continue

### 4. Invalid Flag Combination
**Scenario**: `--force` used without `--non-interactive` or `--daemon`
**Solution**: Validate flag combo in `runListeningCmd()`, show error and exit

### 5. Registry Load Failure
**Scenario**: Can't load session registry file
**Solution**: Show warning but continue (assume no daemons running)

## Testing Requirements

### Unit Tests

**File**: `internal/cli/listen_test.go`

1. Test `killDaemonOnPort()` with existing daemon
2. Test `killDaemonOnPort()` with no daemon (warning message)
3. Test `killDaemonOnPort()` with stale registry entry
4. Test flag validation (--force requires --non-interactive or --daemon)

### Integration Tests

**File**: `internal/cli/daemon_integration_test.go`

1. `listen --daemon` followed by `listen --non-interactive --force` (kills daemon)
2. `listen --non-interactive --force` with no daemon (shows warning, continues)
3. `listen --daemon --force` kills old daemon before forking new one

### Test Coverage Goals
- All edge cases documented above
- Both success and failure paths
- Output message verification

## Documentation Updates

### 1. CLAUDE.md - Available Commands Section

Update line 36:
```bash
xdebug-cli listen --non-interactive --force --commands "cmd1" "cmd2"  # Kill existing daemon, run commands
```

### 2. CLAUDE.md - Non-Interactive Mode Section

Add new subsection after "Basic Usage":

```markdown
### Force Flag

Use `--force` to automatically kill any existing daemon on the same port before starting:

```bash
# Kill stale daemon on port 9003, then run commands
xdebug-cli listen --non-interactive --force --commands "run" "print \$x"

# With custom port
xdebug-cli listen -p 9004 --non-interactive --force --commands "break :42" "run"
```

The `--force` flag:
- Kills only the daemon on the same port (e.g., port 9003)
- Shows warning if no daemon exists, but continues
- Never fails - always proceeds with the new session
- Useful for automation scripts and CI/CD where stale processes may exist
```

### 3. CLAUDE.md - Daemon Mode Section

Add example after "Starting a Daemon":
```bash
# Kill old daemon and start fresh
xdebug-cli listen --daemon --force --commands "break /path/to/file.php:100"
```

## Implementation Checklist

- [ ] Add `Force` field to `CLIParameter` struct
- [ ] Add `--force` flag to `listenCmd.Flags()`
- [ ] Create `killDaemonOnPort(port int)` function
- [ ] Add cleanup call in `runListeningCmd()` when `Force` is true
- [ ] Add flag validation (--force requires --non-interactive or --daemon)
- [ ] Make `processExists` public as `ProcessExists`
- [ ] Write unit tests for `killDaemonOnPort()`
- [ ] Write integration tests for force flag behavior
- [ ] Update CLAUDE.md documentation (3 sections)
- [ ] Manual testing of all edge cases

## Future Considerations

- **Windows support**: Current implementation uses `/proc` for process checking (Linux-specific). Windows would need different implementation.
- **Force-all flag**: Could add `--force-all` to kill all daemons on all ports (not just same port).
- **Timeout on kill**: Could add graceful shutdown with timeout before force kill (currently just kills immediately).

## Success Criteria

- `xdebug-cli listen --non-interactive --force` kills existing daemon and runs successfully
- Warning message shown when no daemon exists
- All edge cases handled gracefully (never fails)
- Clear, actionable output messages
- Full test coverage
- Documentation updated
