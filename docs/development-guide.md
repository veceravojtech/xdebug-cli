# Development Guide

> Generated: 2026-02-23 | Scan Level: Deep

## Prerequisites

- **Go 1.25.4** or later
- **PHP with Xdebug** (for integration testing)
- **Unix-like OS** (Linux/macOS - uses syscall.ForkExec, Unix sockets)

## Getting Started

### Clone and Build

```bash
# Build binary
go build -o xdebug-cli ./cmd/xdebug-cli

# Or install with version info
./install.sh
```

### Install Script Details

The `install.sh` script:
1. Checks Go is installed
2. Downloads dependencies (`go mod download`)
3. Builds with ldflags for version injection:
   ```
   -X github.com/console/xdebug-cli/internal/cli.Version={VERSION}
   -X github.com/console/xdebug-cli/internal/cli.BuildTime={BUILD_TIME}
   ```
4. Installs to `~/.local/bin/xdebug-cli`
5. Warns if `~/.local/bin` is not in PATH

### Configuration

Optional configuration via `.xdebug-cli.yaml`:

```yaml
preview_duration: 5   # Source code preview duration
verbose: false         # Verbose output
```

See `.xdebug-cli.yaml.example` for reference.

## Running Tests

```bash
# Run all unit tests
go test ./...

# Run with coverage
go test ./... -coverprofile=coverage.out

# View coverage report
go tool cover -html=coverage.out

# Run specific package tests
go test ./internal/dbgp/...
go test ./internal/daemon/...
go test ./internal/ipc/...
go test ./internal/view/...
go test ./internal/cli/...
```

### Test Organization

| Package | Test Files | Focus |
|---------|-----------|-------|
| `internal/cli/` | `install_test.go`, `daemon_integration_test.go` | CLI command behavior, daemon lifecycle |
| `internal/dbgp/` | `client_test.go`, `protocol_test.go`, `session_test.go` | Protocol parsing, session state |
| `internal/daemon/` | `daemon_test.go`, `daemon_ipc_integration_test.go` | Daemon operations, IPC integration |
| `internal/ipc/` | `server_test.go`, `protocol_test.go` | Socket communication, message serialization |
| `internal/view/` | `view_test.go`, `json_test.go`, `source_test.go`, `help_test.go` | Output formatting |

### Integration Tests

Shell-based integration tests in `test/integration/`:

```bash
# Run curl trigger integration tests
bash test/integration/test-curl-trigger.sh
```

Tests cover:
- Missing `--curl` flag error messages
- Error includes usage examples
- Curl failure terminates daemon
- XDEBUG_TRIGGER cookie handling

## Project Structure

```
cmd/xdebug-cli/main.go     # Entry point
internal/cfg/               # Configuration (CLIParameter, Version)
internal/cli/               # Cobra commands (root, daemon, attach, install)
internal/dbgp/              # DBGp protocol layer (server, client, session, protocol)
internal/daemon/            # Daemon process management (fork, IPC, registry)
internal/ipc/               # Inter-process communication (Unix sockets, protocol)
internal/view/              # Terminal view (output, source display, help, formatting)
```

## Code Conventions

### Go Standards
- Standard Go project layout (`cmd/`, `internal/`)
- All shared state protected by mutexes
- Interfaces defined in consumer package (`view/types.go`)
- Adapters in producer package (`dbgp/view_adapters.go`)
- Factory functions: `NewXxx()` pattern

### Error Handling
- Exit code 0: success
- Exit code 1: command/runtime error
- Exit code 124: breakpoint timeout (Unix convention)
- Detailed error messages with troubleshooting suggestions

### Testing Pattern
- Unit tests alongside source files (`*_test.go`)
- Integration tests in `test/integration/` (shell scripts)
- Test binaries accepted in process validation (`.test` suffix)

## Build Commands

| Command | Purpose |
|---------|---------|
| `go build -o xdebug-cli ./cmd/xdebug-cli` | Build binary |
| `go test ./...` | Run all tests |
| `go test ./... -coverprofile=coverage.out` | Test with coverage |
| `./install.sh` | Build + install to ~/.local/bin |
| `go mod tidy` | Clean up dependencies |
| `go vet ./...` | Static analysis |

## PHP Configuration for Testing

Configure Xdebug in `php.ini`:

```ini
[xdebug]
zend_extension=xdebug.so
xdebug.mode=debug
xdebug.client_host=127.0.0.1
xdebug.client_port=9003
xdebug.start_with_request=trigger
```

## Common Development Tasks

### Adding a New Debugging Command

1. Add handler in `internal/daemon/executor.go`: `handleNewCommand()`
2. Register command alias in the switch statement in `executeCommand()`
3. Add command to help output in `handleHelp()`
4. Update CLAUDE.md debugging commands table
5. Add tests

### Adding a New CLI Subcommand

1. Create command in `internal/cli/` (follow Cobra pattern from existing commands)
2. Register as subcommand in `root.go` init()
3. Add to README.md and CLAUDE.md

### Modifying the IPC Protocol

1. Update message types in `internal/ipc/protocol.go`
2. Update daemon handler in `internal/daemon/daemon.go:handleIPCRequest()`
3. Update client in `internal/cli/attach.go`
4. Ensure backward compatibility or bump version
