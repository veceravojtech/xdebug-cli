<!-- OPENSPEC:START -->
# OpenSpec Instructions

These instructions are for AI assistants working in this project.

Always open `@/openspec/AGENTS.md` when the request:
- Mentions planning or proposals (words like proposal, spec, change, plan)
- Introduces new capabilities, breaking changes, architecture shifts, or big performance/security work
- Sounds ambiguous and you need the authoritative spec before coding

Use `@/openspec/AGENTS.md` to learn:
- How to create and apply change proposals
- Spec format and conventions
- Project structure and guidelines

Keep this managed block so 'openspec update' can refresh the instructions.

<!-- OPENSPEC:END -->


## Features
* DBGp protocol client for PHP debugging with Xdebug
* Interactive REPL debugging session
* Non-interactive commands for scripting
* Daemon mode for persistent debug sessions (multi-step workflows)
* TCP server for accepting Xdebug connections
* Full debugging operations: run, step (into/over/out), breakpoints, variable inspection
* Conditional breakpoints with PHP expressions
* Multiple breakpoints in single command
* Source code display with line numbers
* TDD with comprehensive test coverage
* Install command (`xdebug-cli install`) installs CLI to `~/.local/bin` with build timestamp

## Available Commands

```bash
xdebug-cli listen --commands "cmd1" "cmd2"               # Execute debugging commands
xdebug-cli listen --commands "cmd1" --force              # Kill existing daemon, run commands
xdebug-cli listen --json --commands "run"                # JSON output for automation
xdebug-cli daemon start --commands "break file.php:100"    # Start persistent daemon session
xdebug-cli attach --commands "context local" "print \$var"    # Attach to daemon and execute commands
xdebug-cli connection                                    # Show connection status
xdebug-cli connection list [--json]                      # List all daemon sessions
xdebug-cli connection isAlive                            # Check if session active (exit 0/1)
xdebug-cli connection kill                               # Terminate active session
xdebug-cli connection kill --all [--force]               # Terminate all daemon sessions
xdebug-cli install                                       # Install binary to ~/.local/bin
xdebug-cli version                                       # Show version and build timestamp
```

## Debugging Commands

Available commands for use with `--commands` flag:

```
run, r              # Continue execution
step, s             # Step into
next, n             # Step over
out, o              # Step out of current function
break, b <target>   # Set breakpoint (:line, file:line, call func, exception)
delete, del <id>    # Delete breakpoint by ID
disable <id>        # Disable breakpoint by ID
enable <id>         # Enable breakpoint by ID
print, p <var>      # Print variable value
set $var = value    # Set variable value
eval, e <expr>      # Evaluate PHP expression
context, c [type]   # Show variables (local/global/constant)
list, l             # Show source code
source, src [file]  # Display source code (alternative to list)
stack               # Show call stack
status, st          # Show current execution status
info, i [topic]     # Show info (breakpoints)
detach, d           # Detach from debug session
finish, f           # Stop debugging
help, h, ?          # Show help
```

### Breakpoint Syntax

**Basic breakpoints:**
```bash
break :42                # Line in current file
break file.php:100       # Specific file and line
break call myFunction    # Function call
break exception          # Any exception
```

**Conditional breakpoints:**
```bash
break :42 if $count > 10
break file.php:100 if $user->isAdmin()
```

**Multiple breakpoints:**
```bash
break :42 :100 :150
break file.php:10 file.php:20 if $debug
```

### Breakpoint Management

Manage breakpoints during debugging sessions:

```bash
# List all breakpoints
xdebug-cli listen --commands "info breakpoints"

# Delete breakpoint by ID
xdebug-cli attach --commands "delete 1"
xdebug-cli attach --commands "del 2"

# Disable breakpoint (keeps it but won't trigger)
xdebug-cli attach --commands "disable 1"

# Enable previously disabled breakpoint
xdebug-cli attach --commands "enable 1"
```

### Execution Control

Control program execution and inspect current state:

```bash
# Show current execution status
xdebug-cli listen --commands "status"
xdebug-cli attach --commands "st"

# Show call stack
xdebug-cli attach --commands "stack"

# Detach from debug session without stopping
xdebug-cli attach --commands "detach"
```

### Variable Inspection and Modification

Inspect and modify variables during debugging:

```bash
# Evaluate PHP expressions
xdebug-cli attach --commands "eval \$x + 10"
xdebug-cli attach --commands "e strlen(\$name)"

# Set variable values
xdebug-cli attach --commands "set \$count = 100"
xdebug-cli attach --commands "set \$debug = true"

# Print variables
xdebug-cli attach --commands "print \$myVar"
xdebug-cli attach --commands "p \$obj->property"
```

### Source Code Display

Display and navigate source code:

```bash
# Show source code (basic)
xdebug-cli listen --commands "list"
xdebug-cli attach --commands "l"

# Show specific file source
xdebug-cli attach --commands "source app.php"
xdebug-cli attach --commands "src app.php"

# Show source code range
xdebug-cli attach --commands "source app.php:100-120"
xdebug-cli attach --commands "list :50-:75"
```

## Command-Based Execution

Execute debugging commands for scripting, automation, and CI/CD:

### Basic Usage

```bash
# Run a single command
xdebug-cli listen --commands "run"

# Execute multiple commands sequentially
xdebug-cli listen --commands "run" "step" "print \$x"

# Set breakpoint and continue
xdebug-cli listen --commands "break :42" "run" "context local"
```

### Force Flag

Use `--force` to automatically kill any existing daemon on the same port before starting:

```bash
# Kill stale daemon on port 9003, then run commands
xdebug-cli listen --force --commands "run" "print \$x"

# With custom port
xdebug-cli listen -p 9004 --force --commands "break :42" "run"
```

The `--force` flag:
- Kills only the daemon on the same port (e.g., port 9003)
- Shows warning if no daemon exists, but continues
- Never fails - always proceeds with the new session
- Useful for automation scripts and CI/CD where stale processes may exist

**Output Examples:**

Daemon killed successfully:
```
Killed daemon on port 9003 (PID 12345)
Server listening on 0.0.0.0:9003
```

No daemon running:
```
Warning: no daemon running on port 9003
Server listening on 0.0.0.0:9003
```

Stale daemon (process already dead):
```
Warning: daemon on port 9003 is stale (PID 12345 no longer exists), cleaning up
Server listening on 0.0.0.0:9003
```

### JSON Output Mode

Enable JSON output for machine parsing and LLM consumption:

```bash
# Get structured JSON output
xdebug-cli listen --json --commands "run" "print \$myVar"

# Example JSON output for run command:
# {"command":"run","success":true,"result":{"status":"break","filename":"/path/to/file.php","line":42}}

# Example JSON output for print command:
# {"command":"print","success":true,"result":{"name":"myVar","fullname":"$myVar","type":"string","value":"hello","num_children":0}}

# Example JSON error output:
# {"command":"step","success":false,"error":"session ended"}
```

### Automation Examples

```bash
# CI/CD debugging script
xdebug-cli listen --json --commands "run" "context local" > debug-output.json

# Check variable value in test
xdebug-cli listen --json --commands "run" "print \$result" | jq '.result.value'

# Set breakpoint and inspect
xdebug-cli listen --commands "break :100" "run" "context local" "finish"
```

### Shell Escaping

When using special characters in commands, ensure proper shell escaping:

```bash
# Escape dollar signs in variable names
xdebug-cli listen --commands "print \$myVar"

# Use quotes for complex expressions
xdebug-cli listen --commands "print \$obj->property"

# File paths with spaces (use quotes)
xdebug-cli listen --commands "break /path/with spaces/file.php:42"
```

### Exit Codes

- `0`: All commands executed successfully
- `1`: Command execution failed or session ended prematurely

## Daemon Mode (Persistent Sessions)

Start a persistent debugging session that runs in the background, allowing multiple CLI invocations to interact with the same debug session without terminating the connection.

### Starting a Daemon

```bash
# Start daemon (waits for Xdebug connection in background)
xdebug-cli daemon start

# Start daemon with initial breakpoint
xdebug-cli daemon start --commands "break /path/to/file.php:100"

# Start daemon with JSON output
xdebug-cli daemon start --json --commands "break :42"

# Kill old daemon and start fresh
xdebug-cli daemon start --commands "break :42"
```

The daemon:
- Runs in the background (detaches from terminal)
- Keeps the debug connection alive across multiple `attach` commands
- Persists until explicitly killed or the debug session ends
- Creates a Unix socket for IPC communication
- Registers its PID and socket path for discovery

### Attaching to a Daemon

Once a daemon is running, attach to it to execute commands:

```bash
# Trigger PHP request (connects to daemon, hits breakpoint)
curl http://localhost/app.php -b "XDEBUG_TRIGGER=1"

# Inspect variables
xdebug-cli attach --commands "context local" "print \$myVar"

# Continue execution
xdebug-cli attach --commands "run"

# Get JSON output for automation
xdebug-cli attach --json --commands "context local"
```

### Managing Daemon Sessions

```bash
# Check daemon status
xdebug-cli connection

# Example output:
# Connection Status: Daemon Mode
#
# PID: 12345
# Port: 9003
# Socket Path: /tmp/xdebug-cli-session-9003.sock
# Started: 2025-12-02 10:30:15

# List all daemon sessions
xdebug-cli connection list

# Example output:
# Active Daemon Sessions:
# PID      Port    Started              Socket Path
# --------------------------------------------------------------------------------
# 12345    9003    2025-12-02 10:30:15  /tmp/xdebug-cli-session-9003.sock
# 67890    9004    2025-12-02 11:45:22  /tmp/xdebug-cli-session-9004.sock
#
# 2 session(s) found

# List all sessions in JSON format
xdebug-cli connection list --json

# Example JSON output:
# [
#   {"pid":12345,"port":9003,"socket_path":"/tmp/xdebug-cli-session-9003.sock","started_at":"2025-12-02T10:30:15Z"},
#   {"pid":67890,"port":9004,"socket_path":"/tmp/xdebug-cli-session-9004.sock","started_at":"2025-12-02T11:45:22Z"}
# ]

# Check if daemon is alive (exit code 0 if running, 1 if not)
xdebug-cli connection isAlive

# Kill daemon on current port
xdebug-cli connection kill

# Kill all daemon sessions (with confirmation)
xdebug-cli connection kill --all

# Example output:
# Found 2 active session(s). Terminate all? (y/N): y
# Killing daemon on port 9003 (PID 12345)... done
# Killing daemon on port 9004 (PID 67890)... done
#
# All 2 session(s) terminated successfully.

# Kill all daemon sessions (skip confirmation)
xdebug-cli connection kill --all --force
```

### Workflow Examples

#### Workflow 1: Set breakpoint, trigger request, inspect

```bash
# 1. Start daemon and set breakpoint (CLI exits, daemon stays)
xdebug-cli daemon start --commands "break /var/www/app.php:100"

# 2. Trigger PHP request (connects to daemon, hits breakpoint, stays paused)
curl http://localhost/app.php -b "XDEBUG_TRIGGER=1"

# 3. Inspect state (attaches to existing session)
xdebug-cli attach --commands "context local" "print \$user"

# 4. Continue execution
xdebug-cli attach --commands "run"

# 5. Kill session when done
xdebug-cli connection kill
```

#### Workflow 2: Incremental debugging

```bash
# Start daemon
xdebug-cli daemon start

# Trigger request
curl http://localhost/app.php -b "XDEBUG_TRIGGER=1"

# Multiple commands across separate invocations
xdebug-cli attach --commands "break :42"
xdebug-cli attach --commands "run"
xdebug-cli attach --commands "context local"
xdebug-cli attach --commands "step" "step" "print \$x"
xdebug-cli attach --commands "finish"
```

#### Workflow 3: Automated testing with daemon

```bash
#!/bin/bash
# Start daemon in background
xdebug-cli daemon start --commands "break /app/critical.php:50"

# Run tests (triggers Xdebug connections)
php vendor/bin/phpunit tests/CriticalTest.php

# Collect debug data
xdebug-cli attach --json --commands "context local" > debug-data.json

# Analyze and continue
if jq -e '.result.variables[] | select(.name == "error")' debug-data.json; then
  echo "Error variable detected"
  xdebug-cli attach --commands "print \$error" "print \$trace"
fi

# Continue or kill
xdebug-cli attach --commands "run"
xdebug-cli connection kill
```

### Daemon Mode vs. Command-Based Execution

| Feature | Daemon Mode | Command-Based Mode |
|---------|-------------|---------------------|
| Session Persistence | Persists across invocations | Terminates after commands |
| Multiple Commands | Across multiple CLI calls | Single CLI invocation |
| Background Process | Yes (detached daemon) | No (foreground process) |
| Use Case | Multi-step workflows | One-shot command execution |
| PHP State | Remains paused at breakpoint | Executes all commands in sequence |

### Error Messages

**No daemon running:**
```
Error: no daemon running on port 9003. Start with:
  xdebug-cli daemon start
```

**Daemon already running:**
```
Error: daemon already running on port 9003 (PID 12345)
Use 'xdebug-cli connection kill' to terminate it first.
```

**Connection failed:**
```
Error: failed to connect to daemon socket: /tmp/xdebug-cli-session-9003.sock
The daemon may have crashed or ended. Check 'xdebug-cli connection' for status.
```

## Development

```bash
# Build
go build -o xdebug-cli ./cmd/xdebug-cli

# Run tests
go test ./...

# Install with version info
./install.sh
```

## Project Structure

```
cmd/xdebug-cli/main.go     # Entry point
internal/cli/              # Cobra commands (root, listen, attach, connection, install)
internal/dbgp/             # DBGp protocol layer (server, client, session, protocol)
internal/daemon/           # Daemon process management (fork, IPC, registry)
internal/ipc/              # Inter-process communication (Unix sockets, protocol)
internal/view/             # Terminal view (output, source display, help, formatting)
internal/cfg/              # Configuration (CLIParameter, Version)
```

## PHP Configuration

Configure Xdebug in php.ini:
```ini
[xdebug]
zend_extension=xdebug.so
xdebug.mode=debug
xdebug.client_host=127.0.0.1
xdebug.client_port=9003
xdebug.start_with_request=yes
```
