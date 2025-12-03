# Change: Implement Xdebug CLI Debugging Functionality

## Why
The project needs to become a fully functional Xdebug/DBGp client for non-interactive CLI debugging of PHP applications. Currently it's a skeleton with only preview/install commands. We need to implement the complete DBGp protocol stack and provide scriptable CLI commands.

## What Changes

### New Capabilities

**1. DBGp Protocol Layer (`internal/dbgp/`)**
- TCP server that listens for Xdebug connections
- Connection wrapper with DBGp message framing (size\0xml\0)
- Protocol parsing for init/response XML messages
- Client with debugging commands (run, step, next, breakpoints, eval)
- Session state management

**2. Terminal View Layer (`internal/view/`)**
- Console output formatting
- Source code display with line numbers
- Property/variable tree display
- Breakpoint table display
- User input handling

**3. CLI Commands (`internal/cli/`)**
- `xdebug-cli listen -p 9003` - Start DBGp server and wait for connections
- `xdebug-cli connection` - Show connection status
- `xdebug-cli connection isAlive` - Check if session is active
- `xdebug-cli connection kill` - Terminate session
- Interactive REPL with commands: run, step, next, break, print, context, list, info, finish

**4. Configuration (`internal/cfg/`)**
- CLI parameters (host, port, trigger)
- Version info

### Command Structure

```
xdebug-cli
├── listen [-p port] [-l host]     # Start DBGp server
├── connection                      # Connection management
│   ├── (no args)                   # Show connection info
│   ├── isAlive                     # Check if connected
│   └── kill                        # Terminate session
├── version                         # Show version
└── install                         # Install binary (existing)
```

### Interactive Debugging Commands (in listen mode REPL)

| Command | Alias | Description |
|---------|-------|-------------|
| run | r | Continue to next breakpoint |
| step | s | Step into next statement |
| next | n | Step over next statement |
| break | b | Set breakpoint (`:line`, `file:line`, `call func`) |
| print | p | Print variable value |
| context | c | Show variables (local/global/constant) |
| list | l | Show source code around current line |
| info | i | Show info (breakpoints) |
| finish | f | Stop debugging session |
| help | h,? | Show help |
| quit | q | Exit debugger |

## Impact

- Affected specs: cfg, dbgp, view, cli
- Affected code:
  - New: `internal/dbgp/` (server, connection, client, session, protocol)
  - New: `internal/view/` (terminal, help, display)
  - New: `internal/cfg/config.go`
  - Modified: `internal/cli/root.go` (add global flags)
  - New: `internal/cli/listen.go`, `internal/cli/connection.go`
  - Remove: `internal/cli/preview.go` (placeholder functionality)
- **BREAKING**: Removes preview command, changes CLI purpose entirely
