# xdebug-cli - Project Overview

> Generated: 2026-02-23 | Scan Level: Deep

## Summary

**xdebug-cli** is a command-line DBGp protocol client for debugging PHP applications with Xdebug. It provides daemon-based persistent debug sessions that allow multi-step debugging workflows across multiple CLI invocations.

## Key Facts

| Property | Value |
|----------|-------|
| **Project Name** | xdebug-cli |
| **Version** | 1.0.2 |
| **Language** | Go 1.25.4 |
| **CLI Framework** | Cobra (spf13/cobra v1.10.1) |
| **Repository Type** | Monolith |
| **Project Type** | CLI tool |
| **Architecture** | Layered daemon with IPC |
| **License** | MIT |
| **Entry Point** | `cmd/xdebug-cli/main.go` |

## What It Does

1. **Starts a background daemon** that listens for Xdebug TCP connections
2. **Triggers PHP execution** via curl (with XDEBUG_TRIGGER cookie) or waits for external triggers
3. **Accepts debugging commands** via Unix socket IPC from the `attach` subcommand
4. **Executes DBGp protocol commands**: breakpoints, stepping, variable inspection, expression evaluation
5. **Outputs results** in human-readable format or JSON for automation

## Core Features

- DBGp protocol client for PHP debugging with Xdebug
- Daemon-based persistent debug sessions for multi-step workflows
- TCP server for accepting Xdebug connections
- Full debugging operations: run, step (into/over/out), breakpoints, variable inspection
- Conditional breakpoints with PHP expressions
- Multiple breakpoints in single command
- Source code display with line numbers
- JSON output mode for automation and scripting
- Command aliases (GDB-style, DBGp-style)
- Semicolon-separated command syntax
- Port conflict detection with IDE awareness
- Breakpoint path suggestions for relative paths
- Breakpoint validation with configurable timeouts

## Technology Stack Summary

| Category | Technology | Purpose |
|----------|-----------|---------|
| Language | Go 1.25.4 | Core implementation |
| CLI | Cobra v1.10.1 | Command parsing |
| Protocol | DBGp (TCP, XML) | PHP debug protocol |
| IPC | Unix sockets (JSON) | Daemon communication |
| Process Mgmt | syscall.ForkExec | Background daemon |
| Build | Go modules + shell | Build and install |

## Architecture Overview

The application is organized in 6 internal packages following Go conventions:

- **cli**: Cobra command definitions (daemon, attach, install, version)
- **daemon**: Daemon lifecycle, command execution, session registry
- **dbgp**: DBGp protocol (TCP server, client, XML parsing, session state)
- **ipc**: Unix socket server/client, JSON protocol
- **view**: Terminal output formatting, JSON output, source display
- **cfg**: Configuration structs and version constants

## Documentation Index

- [Architecture](./architecture.md) - Detailed architecture and design decisions
- [Source Tree Analysis](./source-tree-analysis.md) - Annotated project structure
- [Development Guide](./development-guide.md) - Build, test, and contribute
- [Component Inventory](./component-inventory.md) - All packages and key types

## Existing Project Documentation

- [README.md](../README.md) - User-facing quick start and usage
- [CLAUDE.md](../CLAUDE.md) - Comprehensive AI agent specification
- [internal/view/README.md](../internal/view/README.md) - View package documentation
- [.xdebug-cli.yaml.example](../.xdebug-cli.yaml.example) - Configuration example
