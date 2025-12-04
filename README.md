# xdebug-cli

A command-line DBGp protocol client for debugging PHP applications with Xdebug. Uses daemon-based persistent sessions for multi-step debugging workflows.

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

## Quick Start

```bash
# Start daemon with HTTP trigger and breakpoint
xdebug-cli daemon start --curl "http://localhost/app.php" --commands "break /var/www/app.php:42"

# Inspect variables when breakpoint hits
xdebug-cli attach --commands "context local"

# Continue execution
xdebug-cli attach --commands "run"

# Stop daemon
xdebug-cli daemon kill
```

## Usage

### Daemon Start

Start a daemon session with either `--curl` (HTTP trigger) or `--enable-external-connection` (external trigger):

```bash
# HTTP trigger (recommended)
xdebug-cli daemon start --curl "http://localhost/app.php"
xdebug-cli daemon start --curl "http://localhost/api -X POST -d 'data'"

# External trigger (browser, IDE, manual)
xdebug-cli daemon start --enable-external-connection --commands "break /app/file.php:42"
```

**Flags:**
- `-p, --port int` - Port to listen on (default: 9003)
- `--curl string` - Curl arguments for HTTP trigger
- `--enable-external-connection` - Wait for external Xdebug connection
- `--commands strings` - Initial commands to execute
- `--breakpoint-timeout int` - Timeout for breakpoint validation (default: 30s)
- `--wait-forever` - Disable breakpoint timeout

### Attach

Execute commands on an active daemon session:

```bash
xdebug-cli attach --commands "run"
xdebug-cli attach --commands "step" "print \$x"
xdebug-cli attach --json --commands "context local"
```

**Flags:**
- `--commands strings` - Commands to execute
- `--json` - Output in JSON format

### Daemon Management

```bash
xdebug-cli daemon status              # Show daemon status
xdebug-cli daemon list [--json]       # List all daemon sessions
xdebug-cli daemon isAlive             # Check if daemon active (exit 0/1)
xdebug-cli daemon kill                # Terminate daemon on current port
xdebug-cli daemon kill --all [--force] # Terminate all daemons
```

### Other Commands

```bash
xdebug-cli install    # Install binary to ~/.local/bin
xdebug-cli version    # Show version and build timestamp
```

## Debugging Commands

Available commands for use with `--commands` flag:

| Command | Aliases | Description |
|---------|---------|-------------|
| `run` | `r`, `continue`, `cont` | Continue execution |
| `step` | `s`, `into`, `step_into` | Step into |
| `next` | `n`, `over` | Step over |
| `out` | `o`, `step_out` | Step out |
| `break <target>` | `b` | Set breakpoint |
| `delete <id>` | `del`, `breakpoint_remove` | Delete breakpoint by ID |
| `clear <location>` | | Delete breakpoint by location |
| `disable <id>` | | Disable breakpoint |
| `enable <id>` | | Enable breakpoint |
| `print <var>` | `p`, `property_get -n` | Print variable |
| `set $var = value` | | Set variable value |
| `eval <expr>` | `e` | Evaluate PHP expression |
| `context [type]` | `c` | Show variables (local/global/constant) |
| `list` | `l` | Show source code |
| `source [file]` | `src` | Display source code |
| `stack` | | Show call stack |
| `status` | `st` | Show execution status |
| `info [topic]` | `i` | Show info (breakpoints) |
| `detach` | `d` | Detach from session |
| `finish` | `f` | Stop debugging |
| `help` | `h`, `?` | Show help |

### Breakpoint Syntax

```bash
break :42                    # Line in current file
break /path/file.php:100     # Specific file and line
break call myFunction        # Function call
break exception              # Any exception
break :42 if $count > 10     # Conditional breakpoint
break :42 :100 :150          # Multiple breakpoints
```

## PHP Configuration

Configure Xdebug in `php.ini`:

```ini
[xdebug]
zend_extension=xdebug.so
xdebug.mode=debug
xdebug.client_host=127.0.0.1
xdebug.client_port=9003
xdebug.start_with_request=trigger
```

## Exit Codes

- `0`: Success
- `1`: Command execution failed or session ended
- `124`: Breakpoint timeout

## Development

```bash
go test ./...              # Run tests
go build -o xdebug-cli ./cmd/xdebug-cli  # Build
./install.sh               # Install with version info
```

## Project Structure

```
cmd/xdebug-cli/main.go     # Entry point
internal/cli/              # Cobra commands (root, daemon, attach, install)
internal/dbgp/             # DBGp protocol layer (server, client, session)
internal/daemon/           # Daemon process management (fork, IPC, registry)
internal/ipc/              # Inter-process communication (Unix sockets)
internal/view/             # Terminal view (output, source display, help)
internal/cfg/              # Configuration (CLIParameter, Version)
```

## License

MIT
