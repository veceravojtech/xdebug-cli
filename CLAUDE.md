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
* Daemon-based persistent debug sessions for multi-step workflows
* TCP server for accepting Xdebug connections
* Full debugging operations: run, step (into/over/out), breakpoints, variable inspection
* Conditional breakpoints with PHP expressions
* Multiple breakpoints in single command
* Source code display with line numbers
* JSON output mode for automation and scripting
* TDD with comprehensive test coverage
* Install command (`xdebug-cli install`) installs CLI to `~/.local/bin` with build timestamp

## Available Commands

```bash
xdebug-cli daemon start --curl "<curl-args>" [--commands "cmd1" "cmd2"]  # Start daemon with HTTP trigger
xdebug-cli attach --commands "context local" "print \$var"  # Execute commands on daemon session
xdebug-cli daemon status                                 # Show daemon status
xdebug-cli daemon list [--json]                          # List all daemon sessions
xdebug-cli daemon isAlive                                # Check if daemon active (exit 0/1)
xdebug-cli daemon kill                                   # Terminate active daemon
xdebug-cli daemon kill --all [--force]                   # Terminate all daemon sessions
xdebug-cli install                                       # Install binary to ~/.local/bin
xdebug-cli version                                       # Show version and build timestamp
```

## Debugging Commands

Available commands for use with `--commands` flag:

```
run, r              # Continue execution (aliases: continue, cont)
step, s             # Step into (aliases: into, step_into)
next, n             # Step over (alias: over)
out, o              # Step out of current function (alias: step_out)
break, b <target>   # Set breakpoint (:line, file:line, call func, exception)
delete, del <id>    # Delete breakpoint by ID (alias: breakpoint_remove)
clear <location>    # Delete breakpoint by location (GDB-style)
disable <id>        # Disable breakpoint by ID
enable <id>         # Enable breakpoint by ID
print, p <var>      # Print variable value
property_get -n $v  # Print variable (DBGp-style)
set $var = value    # Set variable value
eval, e <expr>      # Evaluate PHP expression
context, c [type]   # Show variables (local/global/constant)
list, l             # Show source code
source, src [file]  # Display source code (alternative to list)
stack               # Show call stack
status, st          # Show current execution status
info, i [topic]     # Show info (breakpoints)
breakpoint_list     # List breakpoints (DBGp-style)
detach, d           # Detach from debug session
finish, f           # Stop debugging
help, h, ?          # Show help
```

### Command Aliases

xdebug-cli supports multiple naming conventions for commands, making it easier for users familiar with other debuggers:

**GDB-style aliases:**
```bash
continue            # Same as 'run' - continue execution
cont                # Same as 'run' - continue execution (abbreviated)
clear :42           # Delete breakpoint at line 42 (by location, not ID)
clear file.php:100  # Delete breakpoint at specific location
```

**DBGp protocol-style aliases:**
```bash
property_get -n \$var    # Same as 'print $var'
breakpoint_list          # Same as 'info breakpoints'
breakpoint_remove <id>   # Same as 'delete <id>'
```

**Alternative stepping commands:**
```bash
into                # Same as 'step' - step into function
step_into           # Same as 'step' - step into function
over                # Same as 'next' - step over function call
step_out            # Same as 'out' - step out of current function
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
xdebug-cli attach --commands "info breakpoints"

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
xdebug-cli attach --commands "status"
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
xdebug-cli attach --commands "list"
xdebug-cli attach --commands "l"

# Show specific file source
xdebug-cli attach --commands "source app.php"
xdebug-cli attach --commands "src app.php"

# Show source code range
xdebug-cli attach --commands "source app.php:100-120"
xdebug-cli attach --commands "list :50-:75"
```

## Daemon Workflow

All debugging in xdebug-cli uses a daemon-based workflow. The daemon runs in the background and maintains your debug session, allowing you to execute commands across multiple CLI invocations.

### Basic Usage

The `--curl` flag is **required** and triggers the HTTP request automatically with the XDEBUG_TRIGGER cookie:

```bash
# 1. Start daemon with curl trigger (single command - no race conditions!)
xdebug-cli daemon start --curl "http://localhost/app.php"

# 2. Execute debugging commands via attach
xdebug-cli attach --commands "run"
xdebug-cli attach --commands "step" "print \$x"
xdebug-cli attach --commands "context local"

# 3. Stop the daemon when done
xdebug-cli daemon kill
```

### Starting with Breakpoints

Set breakpoints when starting the daemon:

```bash
# Start daemon with breakpoint
xdebug-cli daemon start --curl "http://localhost/app.php" --commands "break /path/file.php:100"

# Start daemon with multiple breakpoints
xdebug-cli daemon start --curl "http://localhost/app.php" --commands "break :42" "break :100"

# Commands execute when the Xdebug connection is established
```

### Complex HTTP Requests

The `--curl` flag supports all curl arguments:

```bash
# POST request with data
xdebug-cli daemon start --curl "http://localhost/api -X POST -d 'name=value'"

# POST with JSON payload
xdebug-cli daemon start --curl "http://localhost/api -X POST -H 'Content-Type: application/json' -d '{\"key\":\"value\"}'"

# With custom headers
xdebug-cli daemon start --curl "http://localhost/api -H 'Authorization: Bearer token'"
```

The XDEBUG_TRIGGER cookie is automatically appended to all requests.

The daemon automatically kills any existing daemon on the same port before starting, so you never need to worry about stale processes.

### JSON Output Mode

Enable JSON output for machine parsing and automation:

```bash
# Get structured JSON output
xdebug-cli attach --json --commands "run" "print \$myVar"

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
xdebug-cli daemon start --curl "http://localhost/test.php" --commands "break /app/test.php:50"
xdebug-cli attach --json --commands "context local" > debug-output.json
xdebug-cli daemon kill

# Check variable value
xdebug-cli daemon start --curl "http://localhost/app.php"
xdebug-cli attach --json --commands "run" "print \$result" | jq '.result.value'
xdebug-cli daemon kill

# Automated debugging workflow
xdebug-cli daemon start --curl "http://localhost/api" --commands "break :100"
xdebug-cli attach --commands "context local" "run"
xdebug-cli daemon kill
```

### Shell Escaping

When using special characters in commands, ensure proper shell escaping:

```bash
# Escape dollar signs in variable names
xdebug-cli attach --commands "print \$myVar"

# Use quotes for complex expressions
xdebug-cli attach --commands "print \$obj->property"

# File paths with spaces (use quotes)
xdebug-cli daemon start --curl "http://localhost/app.php" --commands "break /path/with spaces/file.php:42"

# Complex curl arguments with JSON
xdebug-cli daemon start --curl "http://localhost/api -X POST -d '{\"key\":\"value\"}'"
```

### Exit Codes

- `0`: All commands executed successfully
- `1`: Command execution failed or session ended prematurely

## Managing Daemon Sessions

```bash
# Check daemon status
xdebug-cli daemon status

# Example output:
# Connection Status: Daemon Mode
#
# PID: 12345
# Port: 9003
# Socket Path: /tmp/xdebug-cli-session-9003.sock
# Started: 2025-12-02 10:30:15

# List all daemon sessions
xdebug-cli daemon list

# Example output:
# Active Daemon Sessions:
# PID      Port    Started              Socket Path
# --------------------------------------------------------------------------------
# 12345    9003    2025-12-02 10:30:15  /tmp/xdebug-cli-session-9003.sock
# 67890    9004    2025-12-02 11:45:22  /tmp/xdebug-cli-session-9004.sock
#
# 2 session(s) found

# List all sessions in JSON format
xdebug-cli daemon list --json

# Example JSON output:
# [
#   {"pid":12345,"port":9003,"socket_path":"/tmp/xdebug-cli-session-9003.sock","started_at":"2025-12-02T10:30:15Z"},
#   {"pid":67890,"port":9004,"socket_path":"/tmp/xdebug-cli-session-9004.sock","started_at":"2025-12-02T11:45:22Z"}
# ]

# Check if daemon is alive (exit code 0 if running, 1 if not)
xdebug-cli daemon isAlive

# Kill daemon on current port
xdebug-cli daemon kill

# Kill all daemon sessions (with confirmation)
xdebug-cli daemon kill --all

# Example output:
# Found 2 active session(s). Terminate all? (y/N): y
# Killing daemon on port 9003 (PID 12345)... done
# Killing daemon on port 9004 (PID 67890)... done
#
# All 2 session(s) terminated successfully.

# Kill all daemon sessions (skip confirmation)
xdebug-cli daemon kill --all --force
```

### Workflow Examples

#### Workflow 1: Set breakpoint, trigger request, inspect

```bash
# 1. Start daemon with curl trigger and breakpoint (single command!)
xdebug-cli daemon start --curl "http://localhost/app.php" --commands "break /var/www/app.php:100"

# 2. Inspect state (attaches to existing session after breakpoint hit)
xdebug-cli attach --commands "context local" "print \$user"

# 3. Continue execution
xdebug-cli attach --commands "run"

# 4. Kill session when done
xdebug-cli daemon kill
```

#### Workflow 2: Incremental debugging

```bash
# Start daemon with curl trigger
xdebug-cli daemon start --curl "http://localhost/app.php"

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
# Start daemon with curl trigger and breakpoint
xdebug-cli daemon start --curl "http://localhost/test.php" --commands "break /app/critical.php:50"

# Collect debug data
xdebug-cli attach --json --commands "context local" > debug-data.json

# Analyze and continue
if jq -e '.result.variables[] | select(.name == "error")' debug-data.json; then
  echo "Error variable detected"
  xdebug-cli attach --commands "print \$error" "print \$trace"
fi

# Continue or kill
xdebug-cli attach --commands "run"
xdebug-cli daemon kill
```

## Error Messages

**Missing --curl flag:**
```
Error: --curl flag is required

Usage:
  xdebug-cli daemon start --curl "<curl-args>"

Examples:
  xdebug-cli daemon start --curl "http://localhost/app.php"
  xdebug-cli daemon start --curl "http://localhost/api -X POST -d 'data'"

The --curl flag specifies the HTTP request to trigger Xdebug.
XDEBUG_TRIGGER cookie is added automatically.
```

**Curl failure:**
```
Error: curl failed with exit code 7: Could not connect to host
Daemon terminated.
```

**Curl not found:**
```
Error: curl not found in PATH
```

**No daemon running:**
```
Error: no daemon running on port 9003. Start with:
  xdebug-cli daemon start --curl "http://localhost/app.php"
```

**Daemon already running (auto-killed by daemon start):**

The daemon automatically kills any existing daemon on the same port, showing:
```
Killed daemon on port 9003 (PID 12345)
Daemon started on port 9003
```

**Connection failed:**
```
Error: failed to connect to daemon socket: /tmp/xdebug-cli-session-9003.sock
The daemon may have crashed or ended. Check 'xdebug-cli daemon status' for status.
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
internal/cli/              # Cobra commands (root, daemon, attach, connection, install)
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
