# Implementation Tasks

## Implementation Status

**Status**: ✅ **COMPLETE** - All tasks implemented and tested

### Completion Summary
- ✅ Phase 1: IPC Infrastructure (Tasks 1.1, 1.2, 1.3)
- ✅ Phase 2: Daemon Management (Tasks 2.1, 2.2, 2.3)
- ✅ Phase 3: CLI Integration (Tasks 3.1, 3.2, 3.3)
- ✅ Phase 4: Integration and Polish (Tasks 4.1, 4.2, 4.3, 4.4)

**Total**: 13 tasks completed
**Test Coverage**: 82+ tests (all passing)
**Documentation**: Complete with examples

---

## Phase 1: IPC Infrastructure (Foundation)

### Task 1.1: Create IPC Protocol Package
**Files**: `internal/ipc/protocol.go`
- Define IPC request/response structs (CommandRequest, CommandResponse, CommandResult)
- Implement JSON marshaling for protocol types
- Add validation for request fields
- **Validation**: Unit tests for protocol serialization/deserialization

### Task 1.2: Implement IPC Server
**Files**: `internal/ipc/server.go`
- Create Server struct with Unix socket listener
- Implement `Listen(socketPath)` to bind Unix socket with 0600 permissions
- Implement `Accept()` loop to handle incoming connections
- Implement request handler that reads JSON, processes, returns JSON
- Add graceful shutdown via context cancellation
- **Validation**: Unit tests for socket creation, connection handling, shutdown

### Task 1.3: Implement IPC Client
**Files**: `internal/ipc/client.go`
- Create Client struct with Unix socket connection
- Implement `Connect(socketPath)` to dial Unix socket
- Implement `SendCommands(commands, jsonOutput)` to send request and receive response
- Add connection timeout and error handling
- **Validation**: Unit tests for connect, send, receive, error cases

## Phase 2: Daemon Management

### Task 2.1: Create Session Registry
**Files**: `internal/daemon/registry.go`
- Define SessionInfo struct (PID, Port, SocketPath, StartedAt)
- Implement registry file operations (load, save, add, remove)
- Create registry directory `~/.xdebug-cli/` if missing
- Implement stale entry cleanup (check if PID exists)
- **Validation**: Unit tests for registry CRUD, cleanup logic

### Task 2.2: Implement Daemon Process Management
**Files**: `internal/daemon/daemon.go`
- Create Daemon struct (server, ipcServer, client, shutdown channel)
- Implement `Start()` to fork process using syscall.ForkExec
- Write PID file `/tmp/xdebug-cli-daemon-<port>.pid`
- Register signal handlers (SIGTERM, SIGINT) for cleanup
- Implement `Shutdown()` to close servers, remove PID, update registry
- **Validation**: Integration test for fork, PID creation, signal handling

### Task 2.3: Implement Command Executor for Daemon
**Files**: `internal/daemon/executor.go`
- Extract command dispatching logic from `internal/cli/listen.go`
- Create `ExecuteCommands(client, commands, jsonOutput)` that returns results
- Make thread-safe with mutex for concurrent IPC requests
- Return structured result objects instead of printing
- **Validation**: Unit tests for command execution, result capture, concurrency

## Phase 3: CLI Integration

### Task 3.1: Add Daemon Flag to Listen Command
**Files**: `internal/cli/listen.go`
- Add `--daemon` flag to listen command
- Check for existing daemon on same port (read PID file)
- Fork to daemon if flag set, otherwise run normally
- Pass control to daemon.Start() in child process
- **Validation**: Manual test daemon start, PID file creation

### Task 3.2: Implement Attach Command
**Files**: `internal/cli/attach.go`
- Create `attachCmd` cobra command
- Add `--commands` and `--json` flags (reuse from listen)
- Read registry to find active session socket
- Use IPC client to send commands and display results
- Handle "no session" error gracefully
- **Validation**: Integration test with running daemon

### Task 3.3: Update Connection Commands for Daemon
**Files**: `internal/cli/connection.go`
- Modify `runConnectionStatus()` to detect daemon mode from registry
- Display daemon PID, socket path when in daemon mode
- Modify `runConnectionKill()` to send IPC kill request for daemon
- Update `runConnectionIsAlive()` to check PID file + process existence
- **Validation**: Manual tests with daemon running

## Phase 4: Integration and Polish

### Task 4.1: Integrate IPC Server with Listen Flow
**Files**: `internal/cli/listen.go`
- Start IPC server alongside DBGp server in daemon mode
- Handle IPC requests in background goroutine
- Route IPC commands to executor with active client
- Close IPC server when session ends
- **Validation**: End-to-end test: daemon start, attach commands, kill

### Task 4.2: Add Daemon Lifecycle Cleanup
**Files**: `internal/daemon/daemon.go`
- Remove Unix socket file on shutdown
- Remove PID file on shutdown
- Remove registry entry on shutdown
- Add timeout for graceful shutdown (5 seconds)
- **Validation**: Test cleanup on normal exit, signal, and error conditions

### Task 4.3: Error Messages and Documentation
**Files**: Various
- Add clear error messages for common failure cases:
  - "No daemon running. Start with: xdebug-cli listen --daemon"
  - "Daemon already running on port 9003 (PID 12345)"
  - "Failed to connect to daemon. Session may have ended."
- Update CLAUDE.md with daemon mode examples
- Update help text for listen, attach, connection commands
- **Validation**: Review all error paths, test error messages

### Task 4.4: Integration Testing
**Files**: `internal/cli/daemon_integration_test.go`
- Test full workflow: start daemon → trigger PHP → attach → kill
- Test concurrent attach commands
- Test daemon restart (cleanup stale files)
- Test session end behavior (daemon exits)
- **Validation**: All integration tests pass

## Dependencies Between Tasks

```
Phase 1 (parallel)
  ├─ 1.1 → 1.2, 1.3 (protocol must exist first)

Phase 2 (mostly parallel after 1.x)
  ├─ 2.1 (independent)
  ├─ 2.2 requires 2.1 (daemon uses registry)
  └─ 2.3 (independent, can happen with 2.1/2.2)

Phase 3 (requires Phase 2)
  ├─ 3.1 requires 2.2 (listen needs daemon management)
  ├─ 3.2 requires 1.3, 2.1 (attach needs IPC client + registry)
  └─ 3.3 requires 2.1 (connection commands need registry)

Phase 4 (requires Phase 3)
  ├─ 4.1 requires 1.2, 2.3, 3.1 (integrate IPC with listen)
  ├─ 4.2 requires 2.2 (cleanup logic)
  ├─ 4.3 (independent)
  └─ 4.4 requires all previous tasks
```

## Parallelizable Work

- Phase 1 tasks (1.2 and 1.3) can be done in parallel after 1.1
- Task 2.1 and 2.3 can be done in parallel
- Task 3.2 and 3.3 can be done in parallel
- Documentation (4.3) can happen throughout development

## Estimated Complexity

- **Small tasks** (1-2 hours): 1.1, 1.3, 2.1, 3.1, 3.3, 4.2, 4.3
- **Medium tasks** (2-4 hours): 1.2, 2.3, 3.2, 4.1
- **Large tasks** (4-6 hours): 2.2, 4.4

**Note**: Complexity is for guidance only. Focus on completing tasks fully rather than hitting time estimates.
