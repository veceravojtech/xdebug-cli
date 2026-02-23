# xdebug-cli - Project Documentation Index

> Generated: 2026-02-23 | Scan Level: Deep | Workflow: initial_scan

## Project Overview

- **Type:** Monolith CLI tool
- **Primary Language:** Go 1.25.4
- **Architecture:** Layered daemon with Unix socket IPC
- **Version:** 1.0.2
- **License:** MIT

## Quick Reference

- **Framework:** Cobra (spf13/cobra v1.10.1)
- **Protocol:** DBGp (PHP Xdebug debug protocol)
- **Entry Point:** `cmd/xdebug-cli/main.go`
- **Architecture Pattern:** Background daemon + IPC client
- **Key Dependencies:** cobra, golang.org/x/net (minimal)

## Generated Documentation

- [Project Overview](./project-overview.md) - Summary, key facts, features
- [Architecture](./architecture.md) - Design decisions, process flows, layer diagram
- [Source Tree Analysis](./source-tree-analysis.md) - Annotated directory structure
- [Component Inventory](./component-inventory.md) - All packages, types, and functions
- [Development Guide](./development-guide.md) - Build, test, contribute

## Existing Project Documentation

- [README.md](../README.md) - User-facing quick start and usage guide
- [CLAUDE.md](../CLAUDE.md) - Comprehensive AI agent specification with all commands, workflows, and examples
- [internal/view/README.md](../internal/view/README.md) - View package documentation
- [.xdebug-cli.yaml.example](../.xdebug-cli.yaml.example) - Configuration example
- [_bmad-output/project-context.md](../_bmad-output/project-context.md) - AI project context with 47 implementation rules

## Getting Started

```bash
# Build
go build -o xdebug-cli ./cmd/xdebug-cli

# Install with version info
./install.sh

# Run tests
go test ./...

# Start debugging a PHP application
xdebug-cli daemon start --curl "http://localhost/app.php" --commands "break /var/www/app.php:42"
xdebug-cli attach --commands "context local"
xdebug-cli attach --commands "run"
xdebug-cli daemon kill
```

## Package Map

```
cmd/xdebug-cli/    → Entry point (main.go calls cli.Execute())
internal/cfg/       → Configuration structs, version constant
internal/cli/       → Cobra commands: daemon, attach, install, version
internal/daemon/    → Daemon lifecycle, command executor, session registry
internal/dbgp/      → DBGp protocol: TCP server, client, XML parsing
internal/ipc/       → Unix socket IPC: server, client, JSON protocol
internal/view/      → Terminal output: formatting, JSON, source display
```
