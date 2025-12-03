# Persistent Debug Sessions Design

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────┐
│                     CLI Invocation Layer                     │
├─────────────────────────────────────────────────────────────┤
│  xdebug-cli listen --daemon    │    xdebug-cli attach       │
│           (spawns)             │        (connects)           │
└──────────┬──────────────────────┴────────────┬──────────────┘
           │                                    │
           ▼                                    ▼
┌─────────────────────────────────────────────────────────────┐
│                    IPC Layer (Unix Socket)                   │
│         /tmp/xdebug-cli-session-<pid>.sock                   │
└──────────┬──────────────────────────────────┬───────────────┘
           │                                    │
           ▼                                    ▼
┌─────────────────────────────────────────────────────────────┐
│                    Daemon Process                            │
│  ┌────────────────┐  ┌──────────────┐  ┌─────────────────┐ │
│  │  DBGp Server   │  │ IPC Server   │  │ Session Manager │ │
│  │ (port 9003)    │◄─┤ (Unix socket)│◄─┤  (state store)  │ │
│  └────────┬───────┘  └──────────────┘  └─────────────────┘ │
│           │                                                  │
│           ▼                                                  │
│  ┌────────────────────┐                                     │
│  │   DBGp Client      │                                     │
│  │  (active session)  │                                     │
│  └────────────────────┘                                     │
└─────────────────────────────────────────────────────────────┘
           │
           ▼
┌─────────────────────────────────────────────────────────────┐
│                PHP Process (Xdebug)                          │
└─────────────────────────────────────────────────────────────┘
```

## Core Components

### 1. Daemon Process Management

**Location**: `internal/daemon/daemon.go` (new package)

**Responsibilities**:
- Fork process to background
- Manage PID file (`/tmp/xdebug-cli-daemon-<port>.pid`)
- Handle signal cleanup (SIGTERM, SIGINT)
- Maintain session registry

**Key Decisions**:
- Use `syscall.ForkExec` for daemonization
- Single daemon per port (prevents conflicts)
- PID file includes port number for identification
- Daemon exits when debug session ends or is killed

**Trade-offs**:
- **Chosen**: Single session per daemon
  - **Pro**: Simple state management, clear lifecycle
  - **Con**: Can't debug multiple PHP requests simultaneously
  - **Rationale**: Primary use case is sequential debugging, complexity of multi-session not justified

### 2. IPC Communication

**Location**: `internal/ipc/server.go`, `internal/ipc/client.go` (new package)

**Protocol**:
```json
// Request
{
  "type": "execute_commands",
  "commands": ["break :42", "run", "context local"],
  "json_output": true
}

// Response
{
  "success": true,
  "results": [
    {"command": "break", "success": true, "result": {...}},
    {"command": "run", "success": true, "result": {...}},
    {"command": "context", "success": true, "result": {...}}
  ]
}
```

**Key Decisions**:
- Unix domain socket (not TCP) for local-only, fast IPC
- JSON protocol for human-readable debugging
- Request/response pattern with command batching
- Socket path: `/tmp/xdebug-cli-session-<pid>.sock`

**Trade-offs**:
- **Chosen**: Unix sockets (vs shared memory, TCP)
  - **Pro**: Simple, secure (filesystem permissions), cross-platform (Unix)
  - **Con**: Linux/macOS only, not Windows
  - **Rationale**: Target audience primarily uses Unix systems, simplicity wins

### 3. Session State Management

**Location**: `internal/daemon/session_manager.go` (new file)

**State**:
- Session metadata (IDE key, App ID, target files)
- Current location (file, line)
- Active breakpoints
- Session state (starting, break, running, stopped)
- Connection status

**Key Decisions**:
- In-memory state (not persisted to disk)
- Thread-safe access with mutex (concurrent CLI invocations)
- Session registry file tracks active daemons: `~/.xdebug-cli/sessions.json`

**Trade-offs**:
- **Chosen**: In-memory only (vs persistent)
  - **Pro**: Simple, fast, no serialization complexity
  - **Con**: State lost if daemon crashes
  - **Rationale**: Debug sessions are ephemeral, crash recovery not critical

### 4. CLI Command Flow

#### Daemon Start (`xdebug-cli listen --daemon`)
1. Check for existing daemon on port (via PID file)
2. Fork to background (parent exits, child continues)
3. Start DBGp server (existing logic)
4. Start IPC server (Unix socket)
5. Wait for Xdebug connection
6. Execute initial commands (if provided)
7. Write PID and socket path to registry
8. Run IPC server loop (handle attach commands)

#### Attach (`xdebug-cli attach --commands "..."`)
1. Read session registry to find active daemon
2. Connect to Unix socket
3. Send command batch request
4. Receive and display results
5. Exit (daemon continues running)

#### Connection Kill (`xdebug-cli connection kill`)
1. Read session registry
2. Connect to Unix socket, send kill request
3. Daemon sends DBGp stop, closes connection, exits
4. Remove PID file and registry entry

## Implementation Patterns

### Pattern 1: Command Dispatcher Reuse
Reuse existing `dispatchCommand()` logic from listen.go, but:
- Make it thread-safe for concurrent IPC access
- Return result objects instead of printing directly
- Separate view rendering from command execution

### Pattern 2: Graceful Shutdown
```go
type Daemon struct {
    server    *dbgp.Server
    ipcServer *ipc.Server
    shutdown  chan os.Signal
}

func (d *Daemon) Run() {
    signal.Notify(d.shutdown, syscall.SIGTERM, syscall.SIGINT)
    go d.handleShutdown()
    // ... run servers
}
```

### Pattern 3: Registry Management
```go
type SessionRegistry struct {
    Path     string // ~/.xdebug-cli/sessions.json
    Sessions []SessionInfo
}

type SessionInfo struct {
    PID        int
    Port       int
    SocketPath string
    StartedAt  time.Time
}
```

## Security Considerations

1. **Unix socket permissions**: 0600 (owner only)
2. **PID file verification**: Check process actually exists before trusting PID file
3. **Socket path validation**: Reject paths outside /tmp or user home
4. **Command injection**: Already handled by existing command parsing

## Error Handling

1. **Daemon already running**: Exit with error, display existing PID
2. **Socket connection failure**: Clear error message with troubleshooting steps
3. **Session ended during command**: Return error, daemon exits
4. **Stale PID/socket files**: Auto-cleanup on startup if process doesn't exist

## Testing Strategy

1. **Unit tests**: IPC protocol, session manager, command dispatcher
2. **Integration tests**: Full daemon lifecycle (start, attach, kill)
3. **Manual testing**: Real PHP debugging scenarios with curl triggers

## Open Questions

None - design is complete and ready for implementation.
