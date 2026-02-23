# Source Tree Analysis

> Generated: 2026-02-23 | Scan Level: Deep

## Annotated Directory Tree

```
xdebug-cli/
├── cmd/
│   └── xdebug-cli/
│       └── main.go                    # [ENTRY POINT] Minimal - calls cli.Execute()
│
├── internal/
│   ├── cfg/
│   │   └── config.go                  # CLIParameter struct, Version constant (1.0.2)
│   │
│   ├── cli/                           # Cobra command definitions
│   │   ├── root.go                    # Root command, global flags, versionCmd, Execute()
│   │   ├── daemon.go                  # [LARGEST FILE] Daemon lifecycle, fork, curl, kill
│   │   ├── attach.go                  # Attach command - IPC client, result display
│   │   ├── install.go                 # Install binary to ~/.local/bin
│   │   ├── install_test.go            # Install command tests
│   │   └── daemon_integration_test.go # Daemon CLI integration tests
│   │
│   ├── dbgp/                          # DBGp protocol implementation
│   │   ├── server.go                  # TCP listener, port conflict detection, IDE detection
│   │   ├── connection.go              # Message framing (size\0xml\0), timeouts, validation
│   │   ├── client.go                  # High-level DBGp operations (run, step, break, etc.)
│   │   ├── protocol.go               # XML struct definitions (Init, Response, Property, etc.)
│   │   ├── session.go                 # Session state tracking (state, location, txn IDs)
│   │   ├── view_adapters.go           # Protocol types → view interface adapters
│   │   ├── client_test.go             # Client unit tests
│   │   ├── protocol_test.go           # Protocol parsing tests
│   │   └── session_test.go            # Session state tests
│   │
│   ├── daemon/                        # Daemon process management
│   │   ├── daemon.go                  # Daemon struct, Initialize, Fork, Shutdown, IPC handler
│   │   ├── executor.go                # [CORE] CommandExecutor - 30+ command handlers
│   │   ├── paths.go                   # BreakpointPathStore - relative path suggestions
│   │   ├── registry.go               # SessionRegistry - track active sessions (JSON file)
│   │   ├── daemon_test.go             # Daemon unit tests
│   │   └── daemon_ipc_integration_test.go  # IPC integration tests
│   │
│   ├── ipc/                           # Inter-process communication
│   │   ├── server.go                  # Unix socket server (mode 0600)
│   │   ├── client.go                  # Unix socket client with retry/backoff
│   │   ├── protocol.go               # JSON message types (Request, Response, Result)
│   │   ├── server_test.go             # Server tests
│   │   └── protocol_test.go           # Protocol serialization tests
│   │
│   └── view/                          # Terminal output formatting
│       ├── view.go                    # Base View struct (stdout/stderr writers)
│       ├── types.go                   # Interfaces: ProtocolBreakpoint, ProtocolProperty, ProtocolStack
│       ├── json.go                    # JSON output mode (JSONProperty, JSONBreakpoint, etc.)
│       ├── display.go                 # Formatted terminal output (tables, property trees)
│       ├── source.go                  # SourceFileCache - lazy file loading and line display
│       ├── README.md                  # Package documentation
│       ├── view_test.go               # View tests
│       ├── json_test.go               # JSON formatting tests
│       ├── source_test.go             # Source display tests
│       └── help_test.go               # Help output tests
│
├── test/
│   └── integration/
│       └── test-curl-trigger.sh       # Shell-based integration tests for --curl flag
│
├── scripts/
│   └── hooks/
│       ├── tmux-validate-session.sh   # Validate tmux window UUID
│       └── tmux-session-notify.sh     # Log session lifecycle events
│
├── go.mod                             # Module: github.com/console/xdebug-cli
├── go.sum                             # Dependency checksums
├── install.sh                         # Build + install script (ldflags version injection)
├── coverage.out                       # Test coverage data
├── README.md                          # User-facing documentation
├── CLAUDE.md                          # AI agent specification (comprehensive)
└── .xdebug-cli.yaml.example          # Configuration example (preview duration, verbose)
```

## Critical Folders

| Directory | Purpose | Key Files | LOC (approx) |
|-----------|---------|-----------|---------------|
| `internal/cli/` | CLI command definitions | daemon.go (1254), attach.go (299), root.go | ~1700 |
| `internal/daemon/` | Daemon process management | executor.go (1405), daemon.go (447), registry.go (279) | ~2300 |
| `internal/dbgp/` | DBGp protocol layer | server.go (342), connection.go, client.go, protocol.go | ~900 |
| `internal/ipc/` | Unix socket IPC | server.go, client.go, protocol.go | ~450 |
| `internal/view/` | Output formatting | display.go, json.go, source.go | ~500 |
| `internal/cfg/` | Configuration | config.go | ~50 |

## Entry Points

| Entry Point | Purpose |
|------------|---------|
| `cmd/xdebug-cli/main.go` | Binary entry point → `cli.Execute()` |
| `internal/cli/root.go:Execute()` | Cobra command tree initialization |
| `internal/cli/daemon.go:runDaemonStart()` | Daemon startup (parent process) |
| `internal/cli/daemon.go:runDaemonProcess()` | Daemon worker (child process) |
| `internal/cli/attach.go:runAttachCmd()` | Attach client entry |

## Package Relationship Summary

- **cfg**: Zero dependencies, used by all packages for configuration
- **view**: Zero internal dependencies, defines interfaces consumed by dbgp adapters
- **ipc**: Zero internal dependencies, provides transport layer
- **dbgp**: Depends on view (interface adapters)
- **daemon**: Depends on dbgp, ipc, view (orchestration layer)
- **cli**: Depends on all packages (top-level orchestration)
