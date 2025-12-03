# Design: Xdebug CLI Architecture

## Context

We're building a command-line DBGp client for debugging PHP applications via Xdebug. The design follows the existing library specs but adapted for this specific CLI structure.

## Goals

- Implement DBGp protocol for Xdebug communication
- Provide non-interactive CLI commands for scripting
- Support interactive REPL debugging session
- Keep code modular and testable

## Non-Goals

- GUI interface
- Multiple simultaneous connections
- Watch expressions (can add later)
- Conditional breakpoint GUI

## Architecture

```
cmd/xdebug-cli/main.go
    │
    ▼
internal/cli/           (Cobra commands)
    │
    ├── root.go         (global flags: host, port)
    ├── listen.go       (start server, REPL loop)
    ├── connection.go   (connection status commands)
    ├── version.go      (existing)
    └── install.go      (existing)
    │
    ▼
internal/dbgp/          (Protocol layer)
    │
    ├── server.go       (TCP listener)
    ├── connection.go   (message framing)
    ├── client.go       (debugging operations)
    ├── session.go      (state management)
    └── protocol.go     (XML parsing)
    │
    ▼
internal/view/          (Terminal UI)
    │
    ├── view.go         (output/input facade)
    ├── help.go         (help messages)
    ├── display.go      (property/breakpoint display)
    └── source.go       (source file cache)
    │
    ▼
internal/cfg/           (Configuration)
    │
    └── config.go       (CLIParameter, Version)
```

## Key Decisions

### 1. Package Structure

**Decision**: Use `internal/` directory for all packages to prevent external imports.

**Rationale**: This is a CLI application, not a library. Internal packages enforce encapsulation.

### 2. DBGp Message Framing

**Decision**: Handle DBGp's unique framing format (size\0content\0) in Connection layer.

**Format**:
- Commands: `{command}\0`
- Responses: `{size}\0{xml}\0`

### 3. Session State Machine

**Decision**: Use explicit state enum for session lifecycle.

```go
const (
    StateNone     = iota  // Before init
    StateStarting         // After init, before first run
    StateRunning          // Executing code
    StateBreak            // Paused at breakpoint/step
    StateStopping         // Stop command sent
    StateStopped          // Session ended
)
```

### 4. Connection Management for Non-Interactive Use

**Decision**: Track active connection in global state accessible by subcommands.

**Implementation**:
```go
var activeSession struct {
    sync.RWMutex
    client *dbgp.Client
    active bool
}
```

This allows `xdebug-cli connection isAlive` to check status without running the listen command.

### 5. REPL Command Dispatch

**Decision**: Simple switch-based dispatch in listen command handler.

**Rationale**: Cobra is for CLI commands, not REPL. Internal REPL uses direct string matching for simplicity.

## Data Flow

### Listen Command Flow

```
1. User runs: xdebug-cli listen -p 9003
2. Server.Listen() binds to port
3. Server.Accept() waits for connection
4. PHP script connects (with Xdebug)
5. Connection receives init protocol
6. Client.Init() sets up session
7. REPL loop starts
8. User types commands (step, run, etc.)
9. Commands sent to Xdebug, responses displayed
10. User types 'quit' or session ends
11. Connection closes
```

### Connection Status Flow

```
1. User runs: xdebug-cli connection isAlive
2. Check activeSession.active
3. Print "connected" or "not connected"
4. Exit with code 0 (connected) or 1 (not connected)
```

## Error Handling

- Network errors: Display error, return to REPL (don't crash)
- Protocol errors: Display error message from Xdebug
- User errors: Display help, continue REPL

## Testing Strategy

- Unit tests for protocol parsing
- Unit tests for session state transitions
- Integration tests with mock Xdebug server
- Table-driven tests for REPL command parsing

## Migration Plan

1. Remove `internal/progress/` (preview functionality)
2. Remove `internal/cli/preview.go`
3. Add new packages incrementally
4. Update root.go with new flags

## Open Questions

- Should we support multiple connections queued? (Initial: No, single connection)
- Should we add `run` command for launching PHP? (Can add later as enhancement)
