# Change: Remove Listen Command - Daemon-Only Workflow

## Why

The `xdebug-cli listen` command creates confusion by offering two distinct execution models:
1. Command-based execution (one-shot mode with `--commands`)
2. Daemon mode (background process with `--daemon`)

This dual-mode approach complicates the CLI interface and creates inconsistent user workflows. The daemon-based workflow is more powerful, supporting multi-step debugging across multiple CLI invocations, making the one-shot `listen` mode redundant.

Simplifying to a daemon-only workflow reduces cognitive load, eliminates code duplication, and provides a single, consistent pattern for all debugging sessions.

## What Changes

- **BREAKING**: Remove `xdebug-cli listen` command entirely
- Introduce `xdebug-cli daemon start` command as the sole entry point for debug sessions
- All debugging sessions run as background daemons by default
- Users interact with sessions via `xdebug-cli attach --commands` for executing debugging commands
- Update documentation to reflect daemon-first workflow
- Remove `--force` flag (was specific to `listen` command for killing daemons)
- Session management remains through `xdebug-cli connection` commands

## Impact

**Affected specs:**
- `cli` - Remove Listen Command requirement, modify daemon and attach commands
- `persistent-sessions` - Update to reflect daemon as primary workflow (not alternative mode)

**Affected code:**
- `internal/cli/listen.go` - Remove entirely
- `internal/cli/listen_test.go` - Remove entirely
- `internal/cli/listen_force_test.go` - Remove entirely
- `internal/cli/listen_force_integration_test.go` - Remove entirely
- `internal/cli/daemon.go` - Becomes primary command, remove `--daemon` flag references
- `internal/cli/root.go` - Remove `listenCmd` registration
- `CLAUDE.md` - Update all examples to use daemon workflow

**Migration path:**
- Old: `xdebug-cli listen --commands "run" "print $x"`
- New: `xdebug-cli daemon start && xdebug-cli attach --commands "run" "print $x"`

**User impact:**
- Users must adapt to two-step workflow (start daemon, then attach)
- More explicit lifecycle management (start/stop daemon)
- Better support for multi-step debugging workflows
- Consistent behavior across all use cases
