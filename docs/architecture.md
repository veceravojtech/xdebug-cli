# xdebug-cli Architecture

> Generated: 2026-02-23 | Scan Level: Deep | Project Type: CLI

## Executive Summary

xdebug-cli is a Go-based command-line DBGp protocol client for debugging PHP applications with Xdebug. It uses a **daemon-based architecture** where debug sessions persist in the background, allowing users to issue debugging commands across multiple CLI invocations via Unix socket IPC.

## Technology Stack

| Category | Technology | Version | Purpose |
|----------|-----------|---------|---------|
| Language | Go | 1.25.4 | Core implementation language |
| CLI Framework | spf13/cobra | 1.10.1 | Command parsing, flags, subcommands |
| XML Encoding | golang.org/x/net | 0.47.0 | charset.NewReaderLabel for DBGp XML |
| Text Processing | golang.org/x/text | 0.31.0 | Unicode text handling (indirect) |
| Build System | Go modules | - | Dependency management |
| Installation | Shell script | - | Binary build and install |

## Architecture Pattern: Layered Daemon with IPC

The application follows a layered architecture with daemon process management:

```
┌─────────────────────────────────────────────────┐
│                  CLI Layer (Cobra)                │
│  root.go │ daemon.go │ attach.go │ install.go    │
├─────────────────────────────────────────────────┤
│              IPC Layer (Unix Sockets)             │
│  client.go │ server.go │ protocol.go             │
├─────────────────────────────────────────────────┤
│            Command Executor Layer                 │
│  executor.go (30+ command handlers)               │
├─────────────────────────────────────────────────┤
│             DBGp Protocol Layer                   │
│  client.go │ server.go │ session.go │ protocol.go │
├─────────────────────────────────────────────────┤
│              Network Layer (TCP)                  │
│  connection.go (message framing)                  │
└─────────────────────────────────────────────────┘
```

## Core Architectural Decisions

### 1. Daemon-Based Sessions

**Decision:** Debug sessions run in a long-lived background daemon process rather than a single CLI invocation.

**Rationale:**
- PHP debugging requires persistent connections - Xdebug maintains a TCP connection throughout the debug session
- Users need to issue multiple debugging commands across separate terminal invocations
- Daemon allows non-blocking HTTP trigger + debug session in one atomic operation

**Implementation:**
- Parent process validates, forks child via `syscall.ForkExec`, then monitors
- Child process creates DBGp server, IPC server, and runs until killed
- Environment variable `XDEBUG_CLI_DAEMON_MODE=1` differentiates parent/child

### 2. Unix Socket IPC

**Decision:** Use Unix domain sockets for inter-process communication between `attach` client and daemon.

**Rationale:**
- Fast, low-latency local communication
- File-based socket paths allow session discovery
- No network overhead or port conflicts
- Socket permissions (0600) for security

**Protocol:** JSON messages terminated by newline over Unix socket.

### 3. Parent-Child Status Communication

**Decision:** Use status files (`/tmp/xdebug-cli-daemon-{port}.status`) for parent-child breakpoint validation.

**Rationale:**
- Parent needs to know if breakpoints were hit before exiting
- Status files are simple and reliable for one-time communication
- Avoids complexity of additional IPC channel for startup validation

### 4. Session Registry

**Decision:** Centralized JSON file (`~/.xdebug-cli/sessions.json`) tracks all active daemon sessions.

**Rationale:**
- Multiple daemon sessions can run on different ports
- `attach` command needs to find the correct daemon socket
- Registry enables `daemon list`, `daemon kill --all` operations
- Stale entry cleanup validates process existence

## Process Architecture

### Daemon Startup Flow

```
User: xdebug-cli daemon start --curl "http://localhost/app.php" --commands "break :42"
  │
  ├── Parent Process
  │   ├── Validate flags (--curl or --enable-external-connection required)
  │   ├── Check port availability (detect IDE conflicts)
  │   ├── Auto-kill existing daemon on same port
  │   ├── Fork child process (syscall.ForkExec)
  │   ├── Monitor status file for breakpoint validation
  │   └── Exit with status (0=success, 124=timeout, 1=error)
  │
  └── Child Process (Daemon)
      ├── Write PID file (/tmp/xdebug-cli-daemon-{port}.pid)
      ├── Register in session registry
      ├── Start DBGp TCP server (listen for Xdebug)
      ├── Start IPC Unix socket server (listen for attach)
      ├── Execute curl in background goroutine (if --curl)
      ├── Accept Xdebug connection
      ├── Execute initial --commands
      ├── Write status file (breakpoint validation result)
      ├── Enter command loop (wait for IPC requests)
      └── Shutdown on signal or kill request
```

### Command Execution Flow

```
User: xdebug-cli attach --commands "step" "print $x"
  │
  ├── Look up session in registry (by port)
  ├── Connect to daemon Unix socket
  ├── Send CommandRequest (JSON)
  │
  └── Daemon receives request
      ├── Expand semicolons ("a; b" → ["a", "b"])
      ├── Execute each command via CommandExecutor
      │   ├── handleStep() → DBGp step_into command
      │   └── handlePrint() → DBGp property_get command
      ├── Collect CommandResult array
      ├── Send CommandResponse (JSON)
      │
      └── Client receives response
          └── Display results (formatted text or JSON)
```

## Package Dependencies

```
cmd/xdebug-cli/main.go
    └── internal/cli          (Execute)

internal/cli
    ├── internal/cfg          (CLIParameter, Version)
    ├── internal/daemon       (Daemon, SessionRegistry, CommandExecutor)
    ├── internal/dbgp         (Server, Client, Session)
    ├── internal/ipc          (Client, protocol types)
    └── internal/view         (View, JSON formatting)

internal/daemon
    ├── internal/dbgp         (Server, Client)
    ├── internal/ipc          (Server, protocol types)
    └── internal/view         (JSON conversion)

internal/dbgp
    └── internal/view         (interface adapters)

internal/ipc
    └── (no internal deps)

internal/view
    └── (no internal deps)

internal/cfg
    └── (no internal deps)
```

## Thread Safety Model

All shared state is protected by mutexes:

| Component | Lock Type | Protected State |
|-----------|-----------|-----------------|
| Session (dbgp) | RWMutex | Execution state, location, transaction IDs |
| SessionRegistry | Mutex | Session file read/write |
| CommandExecutor | Mutex | Command execution serialization |
| Daemon | Mutex | Daemon lifecycle state |
| BreakpointPathStore | RWMutex | Path cache read/write |

## Error Handling Strategy

| Error Type | Exit Code | Behavior |
|-----------|-----------|----------|
| Success | 0 | All commands executed |
| Command failure | 1 | Session ended or command error |
| Breakpoint timeout | 124 | Unix timeout convention |
| Port conflict | 1 | Identifies blocking process (IDE detection) |
| Curl failure | 1 | Daemon terminates, error logged |
| EOF during execution | 1 | Suggests path issues or PHP errors |

## File System Artifacts

| Path | Purpose | Lifecycle |
|------|---------|-----------|
| `/tmp/xdebug-cli-daemon-{port}.pid` | Process ID file | Created on start, removed on shutdown |
| `/tmp/xdebug-cli-session-{port}.sock` | IPC Unix socket | Created on start, removed on shutdown |
| `/tmp/xdebug-cli-daemon-{port}.status` | Parent-child communication | Created during startup, temporary |
| `/tmp/xdebug-cli-daemon-{port}.log` | Daemon log file | Created on start, persists |
| `~/.xdebug-cli/sessions.json` | Session registry | Persistent, cleaned up on stale detection |
| `~/.xdebug-cli/breakpoint-paths.json` | Breakpoint path suggestions | Persistent, grows with usage |

## External Dependencies (Minimal)

- **spf13/cobra** (v1.10.1): CLI framework - commands, flags, help generation
- **golang.org/x/net** (v0.47.0): HTML charset detection for DBGp XML encoding
- **Standard library only** for: TCP networking, Unix sockets, process management, XML parsing, JSON serialization, signal handling, file I/O
