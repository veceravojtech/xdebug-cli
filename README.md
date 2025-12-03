# xdebug-cli

A command-line DBGp protocol client for debugging PHP applications with Xdebug. Provides both interactive REPL debugging and non-interactive scriptable commands.

## Installation

### Using install script

```bash
./install.sh
```

This builds the binary with version information and installs it to `~/.local/bin/`.

### Manual build

```bash
go build -o xdebug-cli ./cmd/xdebug-cli
```

## Usage

### Listen Command

Start a DBGp server and wait for Xdebug connections:

```bash
xdebug-cli listen [flags]
```

**Flags:**
- `-p, --port int` - Port to listen on (default: 9003)
- `-l, --host string` - Host address to bind to (default: "0.0.0.0")

**Example:**
```bash
# Start server on default port 9003
xdebug-cli listen

# Listen on custom port
xdebug-cli listen -p 9000

# Listen on all interfaces
xdebug-cli listen -l 0.0.0.0 -p 9003
```

Once connected, an interactive REPL provides debugging commands:

**Debugging Commands:**
- `run` or `r` - Continue execution to next breakpoint
- `step` or `s` - Step into next statement
- `next` or `n` - Step over next statement
- `break` or `b` - Set breakpoint (see formats below)
- `print` or `p` - Print variable value
- `context` or `c` - Show variables in context (local/global/constant)
- `list` or `l` - Show source code around current line
- `info` or `i` - Show debugging information
- `finish` or `f` - Stop debugging session
- `help` or `h` or `?` - Show help
- `quit` or `q` - Exit debugger

**Breakpoint Formats:**
```
break :42              # Line 42 in current file
break 42               # Line 42 in current file
break /path/file.php:10  # Specific file and line
break call myFunction  # Function call breakpoint
break exception        # Break on any exception
break exception ValueError  # Break on specific exception
```

**Context Types:**
```
context local          # Show local variables
context global         # Show global variables
context constant       # Show constants
```

**Info Commands:**
```
info breakpoints       # List all breakpoints
info b                 # Shorthand for breakpoints
```

### Connection Command

Manage and inspect active debugging connections:

```bash
# Show connection status
xdebug-cli connection

# Check if session is alive (exit code 0=connected, 1=not connected)
xdebug-cli connection isAlive

# Terminate active session
xdebug-cli connection kill
```

### Install Command

Install the binary to `~/.local/bin/`:

```bash
xdebug-cli install
```

### Version

Show version and build information:

```bash
xdebug-cli version
```

## PHP Configuration

To use xdebug-cli with your PHP application, configure Xdebug in `php.ini`:

```ini
[xdebug]
zend_extension=xdebug.so
xdebug.mode=debug
xdebug.client_host=127.0.0.1
xdebug.client_port=9003
xdebug.start_with_request=yes
```

## Development

### Run tests

```bash
go test ./...
```

### Build with version info

```bash
go build -o xdebug-cli ./cmd/xdebug-cli
```

Or use the install script which handles version tagging:

```bash
./install.sh
```

## Project Structure

```
cmd/xdebug-cli/main.go      # Entry point
internal/cli/               # Cobra commands
  root.go                   # Root command with global flags
  listen.go                 # Listen command with REPL
  connection.go             # Connection management commands
  install.go                # Install command
internal/dbgp/              # DBGp protocol implementation
  server.go                 # TCP listener
  connection.go             # Message framing
  client.go                 # Debugging operations
  session.go                # State management
  protocol.go               # XML parsing
internal/view/              # Terminal UI
  view.go                   # Output/input facade
  source.go                 # Source file display
  help.go                   # Help messages
  display.go                # Property/breakpoint formatting
internal/cfg/               # Configuration
  config.go                 # CLIParameter, Version
```

## Example Workflow

```bash
# Terminal 1: Start debugger
xdebug-cli listen -p 9003

# Terminal 2: Run PHP script (configured with Xdebug)
php my-script.php

# Back in Terminal 1: Interactive debugging
(xdbg) break /path/to/my-script.php:10
(xdbg) run
(xdbg) print $myVariable
(xdbg) context local
(xdbg) step
(xdbg) list
(xdbg) finish
```

## License

MIT
