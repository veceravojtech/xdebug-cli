# Enable Persistent Debug Sessions

## Problem Statement

The current non-interactive mode implementation calls `os.Exit(exitCode)` immediately after executing commands (internal/cli/listen.go:111), which:

- Terminates the entire process
- Severs the debug connection while PHP is still at breakpoint
- Makes it impossible to send additional commands or continue debugging
- Requires restarting the entire debug session to execute more commands

This severely limits the usefulness of non-interactive mode for automation, scripting, and multi-step debugging workflows where you want to:
- Set breakpoints, trigger the request, then inspect state
- Execute debug commands incrementally across multiple CLI invocations
- Integrate with external tools that issue debug commands separately
- Keep PHP execution paused while performing analysis

## Solution Overview

Enable persistent debug sessions by:

1. **Daemon Mode**: Add `--daemon` flag to listen command that keeps the server and debug session alive in the background after executing initial commands
2. **Session Management**: Implement Unix socket-based IPC to persist session state and allow multiple CLI invocations to communicate with the same debug session
3. **Background Process**: Fork the listen process to background, allowing CLI to exit while keeping session alive
4. **Session Commands**: Extend connection commands to attach to, query, and kill persistent sessions

## User Workflows

### Workflow 1: Set breakpoint, trigger request, inspect
```bash
# Start daemon and set breakpoint (CLI exits, daemon stays)
xdebug-cli listen --daemon --commands "break /path/file.php:100"

# Trigger PHP request (connects to daemon, hits breakpoint, stays paused)
curl http://localhost/trigger.php -b "XDEBUG_TRIGGER=1"

# Inspect state (attaches to existing session)
xdebug-cli attach --commands "context local" "print \$myVar"

# Continue execution
xdebug-cli attach --commands "run"

# Kill session when done
xdebug-cli connection kill
```

### Workflow 2: Incremental debugging
```bash
# Start daemon
xdebug-cli listen --daemon

# Multiple commands across separate invocations
xdebug-cli attach --commands "break :42"
xdebug-cli attach --commands "run"
xdebug-cli attach --commands "context local"
xdebug-cli attach --commands "step" "step" "print \$x"
xdebug-cli attach --commands "finish"
```

## Capabilities Affected

### New Capability: persistent-sessions
- Session lifecycle management (daemon mode, background process)
- IPC mechanism for session communication
- Attach command for connecting to existing sessions
- Session state persistence and recovery

### Modified Capability: cli
- Listen command daemon mode flag
- Connection commands extended for persistent sessions
- Session lifecycle integration

### Modified Capability: dbgp
- Session state serialization for IPC
- Thread-safe session access from multiple CLI instances

## Dependencies

- Unix domain sockets for IPC (Linux/macOS)
- Process forking and daemonization
- Session state serialization (JSON or gob)
- File-based session registry

## Out of Scope

- Windows support (Unix sockets not available)
- Multiple concurrent debug sessions (single session per daemon)
- Remote session access (local Unix socket only)
- Session persistence across daemon restarts

## Success Criteria

1. User can start daemon, trigger request, and issue commands across multiple CLI invocations
2. PHP execution remains paused at breakpoint while CLI exits
3. Session state is accessible from subsequent CLI invocations
4. Connection commands correctly report and manage persistent sessions
5. Daemon process cleans up properly when session ends or is killed
