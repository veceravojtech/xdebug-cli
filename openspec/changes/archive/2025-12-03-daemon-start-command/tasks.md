# Implementation Tasks: Daemon Start Command

## Overview
This implementation introduces the `daemon start` command to replace `listen --daemon`, providing better discoverability and improved defaults (auto-force enabled).

## Tasks

### 1. Create daemon.go with parent command structure [DONE]
**What**: Create `internal/cli/daemon.go` with parent `daemon` command
**Why**: Establishes command hierarchy for future daemon subcommands
**Location**: `internal/cli/daemon.go`
**Changes**:
- Create `daemonCmd` as parent command with help text
- Register with `rootCmd` in `init()`
- No flags on parent (all flags on subcommands)
**Validation**: `xdebug-cli daemon` shows help listing subcommands

### 2. Implement daemon start subcommand [DONE]
**What**: Add `startCmd` as subcommand under `daemonCmd`
**Why**: Provides the actual daemon starting functionality
**Location**: `internal/cli/daemon.go`
**Changes**:
- Create `startCmd` with usage/help text
- Add `--commands` flag (string array)
- Inherit `-p`/`--port` and `--json` from global flags
- Register as subcommand: `daemonCmd.AddCommand(startCmd)`
**Validation**: `xdebug-cli daemon start --help` shows correct usage

### 3. Move daemon mode logic from listen.go to daemon.go [DONE]
**What**: Extract daemon startup logic from `listen.go` into reusable functions
**Why**: Avoids code duplication and keeps listen.go focused
**Location**: `internal/cli/daemon.go` and `internal/cli/listen.go`
**Changes**:
- Extract `runDaemonMode()`, `runDaemonChild()` to shared functions
- Rename to `startDaemonMode()` and `runDaemonProcess()` for clarity
- Keep functions in `listen.go` initially, then move to new file if needed
- Update `startCmd` Run handler to call `startDaemonMode()`
**Validation**: Daemon starts correctly from new command

### 4. Add auto-force behavior to daemon start [DONE]
**What**: Make daemon start always kill existing daemons on same port
**Why**: Eliminates need for manual `--force` flag, improves UX
**Location**: `internal/cli/daemon.go`
**Changes**:
- In `startCmd` Run handler, call `killDaemonOnPort()` before starting
- Use existing `killDaemonOnPort()` function from `listen.go`
- Move `killDaemonOnPort()` to shared location if needed
- Always call it (no conditional flag check)
**Validation**: Starting daemon on occupied port auto-kills old daemon

### 5. Remove --daemon flag from listen command [DONE]
**What**: Remove `--daemon` flag and all daemon mode handling from `listenCmd`
**Why**: Daemon mode now exclusively uses `daemon start` command
**Location**: `internal/cli/listen.go`
**Changes**:
- Remove `--daemon` flag from `listenCmd.Flags()`
- Remove `CLIArgs.Daemon` checks in `validateListenFlags()`
- Remove daemon mode branch in `runListeningCmd()`
- Update validation error messages to suggest `daemon start`
- Keep command-based execution logic intact
**Validation**: `xdebug-cli listen` requires `--commands`, no `--daemon` flag exists

### 6. Update CLIParameter struct [DONE]
**What**: Keep or remove `Daemon` field from `CLIParameter`
**Why**: Field may still be used by `daemon start` command
**Location**: `internal/cfg/parameter.go` (or wherever `CLIParameter` is defined)
**Changes**:
- Check if `Daemon` field is still needed
- If not needed (daemon start doesn't use it), remove field
- If needed, keep but update documentation
**Validation**: Code compiles without errors

### 7. Remove --force flag validation from listen [DONE]
**What**: Update `validateListenFlags()` to remove `--force` validation
**Why**: `--force` was only relevant with `--daemon` flag
**Location**: `internal/cli/listen.go`
**Changes**:
- Remove validation that checks `--force` requires `--daemon`
- Keep `--force` functionality in listen.go for backwards compatibility if needed
- Or remove entirely if only used for daemon mode
**Validation**: `xdebug-cli listen --commands "run"` works without errors

### 8. Write unit tests for daemon command [DONE]
**What**: Create `internal/cli/daemon_test.go` with test coverage
**Why**: Ensure daemon start command works correctly
**Location**: `internal/cli/daemon_test.go`
**Changes**:
- Test parent command shows help
- Test start subcommand validates flags
- Test start subcommand handles --commands parsing
- Test auto-force behavior (mock killDaemonOnPort)
**Validation**: `go test ./internal/cli -run TestDaemon` passes

### 9. Update integration tests [DONE]
**What**: Update `daemon_integration_test.go` to use new command
**Why**: Existing tests use `listen --daemon` which is being removed
**Location**: `internal/cli/daemon_integration_test.go`
**Changes**:
- Replace all `listen --daemon` invocations with `daemon start`
- Update expected behavior for auto-force
- Verify tests still pass with new command
**Validation**: `go test ./internal/cli -run Integration` passes

### 10. Update listen.go tests [DONE]
**What**: Update tests to remove daemon mode coverage
**Why**: Daemon mode no longer part of listen command
**Location**: `internal/cli/listen_test.go`, `internal/cli/listen_force_test.go`
**Changes**:
- Remove tests for `--daemon` flag
- Remove tests for `listen --force --daemon` combination
- Keep tests for command-based execution
- Update error message assertions to match new validation
**Validation**: `go test ./internal/cli/listen_test.go` passes

### 11. Update CLAUDE.md documentation [DONE]
**What**: Replace all `listen --daemon` examples with `daemon start`
**Why**: Documentation must reflect new command structure
**Location**: `CLAUDE.md`
**Changes**:
- Search for "listen --daemon" and replace with "daemon start"
- Update workflow examples
- Update Available Commands section
- Remove `--force` flag from examples (now implicit)
**Validation**: Review documentation for consistency

### 12. Update README.md if exists [DONE]
**What**: Update README examples to use new command
**Why**: User-facing documentation must be current
**Location**: `README.md` (if exists)
**Changes**:
- Replace `listen --daemon` with `daemon start`
- Update command reference table
- Add migration note for existing users
**Validation**: Review README for accuracy

### 13. Run full test suite [DONE]
**What**: Verify all tests pass
**Why**: Ensure no regressions introduced
**Command**: `go test ./...`
**Expected**: All tests pass
**If failures**: Debug and fix failing tests before proceeding

### 14. Build and smoke test [DONE]
**What**: Build binary and manually test daemon start command
**Why**: Final validation of functionality
**Commands**:
```bash
go build -o xdebug-cli ./cmd/xdebug-cli
./xdebug-cli daemon start
./xdebug-cli daemon start -p 9004
./xdebug-cli daemon start --commands "break :42"
./xdebug-cli connection
./xdebug-cli connection kill
```
**Expected**: All commands work as documented

## Dependencies
- Task 2 depends on Task 1 (parent command must exist)
- Task 3 depends on Task 2 (subcommand must exist before adding logic)
- Task 4 depends on Task 3 (daemon logic must be in place)
- Task 5 can be done in parallel with Tasks 3-4
- Task 9 depends on Task 3 (new command must work)
- Task 10 depends on Task 5 (listen changes must be done)
- Tasks 11-12 can be done in parallel after Tasks 1-10
- Task 13 depends on all code tasks (1-10)
- Task 14 depends on Task 13 (tests must pass)

## Rollback Plan
If issues are discovered:
1. Revert changes to `listen.go` to restore `--daemon` flag
2. Remove `daemon.go` and related tests
3. Update documentation back to original examples
