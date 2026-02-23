# Component Inventory

> Generated: 2026-02-23 | Scan Level: Deep

## Package: internal/cfg

Configuration and version management.

### Types

| Type | Kind | Purpose |
|------|------|---------|
| `CLIParameter` | struct | All CLI flags and arguments |

### Key Fields of CLIParameter

| Field | Type | Default | Purpose |
|-------|------|---------|---------|
| `Host` | string | "0.0.0.0" | Listen address |
| `Port` | int | 9003 | DBGp listen port |
| `Trigger` | string | - | Xdebug IDE key |
| `Commands` | []string | - | Commands to execute |
| `JSON` | bool | false | JSON output mode |
| `Curl` | string | - | Curl arguments for HTTP trigger |
| `BreakpointTimeout` | int | 30 | Timeout for breakpoint validation (seconds) |
| `WaitForever` | bool | false | Disable breakpoint timeout |
| `EnableExternalConnection` | bool | false | Wait for external trigger |
| `KillAll` | bool | false | Kill all daemons |
| `Force` | bool | false | Skip confirmations |

---

## Package: internal/cli

Cobra CLI command definitions. Top-level orchestration.

### Commands

| Command | Function | File | Purpose |
|---------|----------|------|---------|
| `root` | `Execute()` | root.go | Entry point, global flags |
| `version` | `versionCmd` | root.go | Version + build timestamp |
| `daemon start` | `runDaemonStart()` | daemon.go | Start daemon (parent/fork) |
| `daemon status` | `runDaemonStatus()` | daemon.go | Show session info |
| `daemon list` | `runDaemonList()` | daemon.go | List all sessions |
| `daemon kill` | `runDaemonKill()` | daemon.go | Terminate session |
| `daemon kill --all` | `runDaemonKillAll()` | daemon.go | Terminate all sessions |
| `daemon isAlive` | `runDaemonIsAlive()` | daemon.go | Check daemon running |
| `attach` | `runAttachCmd()` | attach.go | Execute commands on daemon |
| `install` | `runInstall()` | install.go | Install binary |

### Key Internal Functions

| Function | File | Purpose |
|----------|------|---------|
| `forkDaemon()` | daemon.go | Fork child process via syscall.ForkExec |
| `runDaemonProcess()` | daemon.go | Daemon worker logic (child) |
| `killDaemonOnPort()` | daemon.go | Graceful daemon termination |
| `executeCurl()` | daemon.go | Async HTTP trigger with XDEBUG_TRIGGER |
| `parseShellArgs()` | daemon.go | Shell-style argument parsing |
| `displayCommandResult()` | attach.go | Format command results for display |

---

## Package: internal/daemon

Daemon process management and command execution.

### Types

| Type | Kind | Purpose |
|------|------|---------|
| `Daemon` | struct | Daemon lifecycle manager |
| `CommandExecutor` | struct | Execute debugging commands |
| `BreakpointPathStore` | struct | Relative path suggestion cache |
| `SessionRegistry` | struct | Active session tracker |
| `SessionInfo` | struct | Serializable session metadata |

### Daemon Struct

| Field | Type | Purpose |
|-------|------|---------|
| `server` | *dbgp.Server | TCP listener for Xdebug |
| `ipcServer` | *ipc.Server | Unix socket server |
| `client` | *dbgp.Client | Active DBGp connection |
| `executor` | *CommandExecutor | Command handler |
| `registry` | *SessionRegistry | Session tracking |
| `port` | int | Listening port |
| `shutdown` | chan struct{} | Shutdown signal |

### CommandExecutor - 30+ Command Handlers

| Handler | Commands | DBGp Command |
|---------|----------|--------------|
| `handleRun()` | run, r, continue, cont | run |
| `handleStep()` | step, s, into, step_into | step_into |
| `handleNext()` | next, n, over | step_over |
| `handleStepOut()` | out, o, step_out | step_out |
| `handleBreak()` | break, b | breakpoint_set |
| `handleDelete()` | delete, del, breakpoint_remove | breakpoint_remove |
| `handleClear()` | clear | breakpoint_remove (by location) |
| `handleEnable()` | enable | breakpoint_update |
| `handleDisable()` | disable | breakpoint_update |
| `handlePrint()` | print, p | property_get |
| `handlePropertyGet()` | property_get | property_get |
| `handleContext()` | context, c | context_get |
| `handleEval()` | eval, e | eval |
| `handleSet()` | set | property_set |
| `handleStatus()` | status, st | status |
| `handleInfo()` | info, i | breakpoint_list / stack_get |
| `handleStack()` | stack | stack_get |
| `handleList()` | list, l | source (display) |
| `handleSource()` | source, src | source |
| `handleDetach()` | detach, d | detach |
| `handleFinish()` | finish, f | stop |
| `handleHelp()` | help, h, ? | (local) |

### SessionRegistry Methods

| Method | Purpose |
|--------|---------|
| `Add()` | Register new session |
| `Remove()` | Unregister session |
| `Get()` | Lookup by port |
| `List()` | Get all sessions |
| `CleanupStaleEntries()` | Remove dead processes |

### BreakpointPathStore Methods

| Method | Purpose |
|--------|---------|
| `SaveBreakpointPath()` | Store successful path |
| `LoadBreakpointPath()` | Get suggested path |
| `HasNonAbsoluteBreakpoint()` | Detect relative paths |

---

## Package: internal/dbgp

DBGp protocol implementation.

### Types

| Type | Kind | Purpose |
|------|------|---------|
| `Server` | struct | TCP listener for Xdebug connections |
| `Connection` | struct | DBGp message framing |
| `Client` | struct | High-level DBGp operations |
| `Session` | struct | Session state tracking |
| `PortConflictInfo` | struct | Port conflict details |
| `ProtocolInit` | struct | Xdebug init message |
| `ProtocolResponse` | struct | DBGp response message |
| `ProtocolProperty` | struct | Variable/property data |
| `ProtocolBreakpoint` | struct | Breakpoint data |
| `ProtocolStack` | struct | Stack frame data |
| `SessionStateType` | string enum | none, starting, running, break, stopping, stopped |

### Server Methods

| Method | Purpose |
|--------|---------|
| `Listen()` | Start TCP listener (SO_REUSEADDR) |
| `Accept()` | Wait for connection |
| `AcceptWithTimeout()` | Accept with deadline |
| `CheckPortInUse()` | Detect port conflicts |
| `FormatPortConflictError()` | User-friendly error |

### Connection Constants

| Constant | Value | Purpose |
|----------|-------|---------|
| `MaxMessageSize` | 100MB | Safety limit for message parsing |
| `DefaultMessageTimeout` | 30s | Read timeout |

### Protocol Adapters (view_adapters.go)

Protocol types implement `view.ProtocolBreakpoint`, `view.ProtocolProperty`, `view.ProtocolStack` interfaces.

---

## Package: internal/ipc

Inter-process communication via Unix sockets.

### Types

| Type | Kind | Purpose |
|------|------|---------|
| `Server` | struct | Unix socket listener |
| `Client` | struct | Unix socket client |
| `CommandRequest` | struct | Request message |
| `CommandResponse` | struct | Response message |
| `CommandResult` | struct | Single command result |
| `RequestHandler` | func type | Request processing callback |

### Request Types

| Type Value | Purpose |
|-----------|---------|
| `"execute_commands"` | Execute debugging commands |
| `"kill"` | Terminate daemon |

### Client Retry Logic

- Default 3 retry attempts
- Exponential backoff: 100ms * 2^attempt
- 5-second connection/read timeout

---

## Package: internal/view

Terminal output formatting.

### Types

| Type | Kind | Purpose |
|------|------|---------|
| `View` | struct | Output handler (stdout/stderr) |
| `SourceFileCache` | struct | Lazy file loader and line display |
| `JSONProperty` | struct | JSON-serializable variable |
| `JSONBreakpoint` | struct | JSON-serializable breakpoint |
| `JSONStack` | struct | JSON-serializable stack frame |
| `JSONResponse` | struct | Standard JSON response |

### Interfaces (types.go)

| Interface | Methods | Implemented By |
|-----------|---------|---------------|
| `ProtocolBreakpoint` | GetID, GetType, GetState, GetFilename, GetLineNumber, GetFunction | dbgp.ProtocolBreakpoint |
| `ProtocolProperty` | GetName, GetFullName, GetType, GetValue, GetChildren, HasChildren, GetNumChildren | dbgp.ProtocolProperty |
| `ProtocolStack` | GetWhere, GetLevel, GetType, GetFilename, GetLineNumber | dbgp.ProtocolStack |

### Key Functions

| Function | Purpose |
|----------|---------|
| `OutputJSON()` | Write single JSON result |
| `OutputJSONArray()` | Write JSON array |
| `ConvertPropertyToJSON()` | Recursive property → JSON |
| `ShowInfoBreakpoints()` | Breakpoint table display |
| `PrintPropertyListWithDetails()` | Variable tree display |
| `TryDecodeBase64()` | Decode Xdebug base64 strings |
