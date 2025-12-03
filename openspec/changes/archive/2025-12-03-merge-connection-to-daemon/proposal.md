# Change: Merge Connection Command to Daemon

## Why

The `xdebug-cli connection` command provides session management functionality (status, list, kill, isAlive) that is conceptually tied to daemon operations. Currently, these two commands are separate:

- `xdebug-cli daemon start` - Creates a daemon session
- `xdebug-cli connection` - Manages daemon sessions
- `xdebug-cli connection list` - Lists daemon sessions
- `xdebug-cli connection kill` - Terminates daemon sessions
- `xdebug-cli connection isAlive` - Checks daemon status

This separation creates an inconsistent user experience where daemon lifecycle operations are split across two top-level commands. Users must remember which operations belong under `daemon` vs `connection`, despite all operations relating to daemon session management.

Merging connection functionality into daemon subcommands creates a unified, intuitive interface where all daemon-related operations live under a single command namespace. This aligns with the daemon-first architecture established by removing the `listen` command.

## What Changes

- **BREAKING**: Remove `xdebug-cli connection` top-level command entirely
- Introduce new daemon subcommands for session management:
  - `xdebug-cli daemon status` (replaces `connection` with no args)
  - `xdebug-cli daemon list [--json]` (replaces `connection list`)
  - `xdebug-cli daemon kill [--all] [--force]` (replaces `connection kill`)
  - `xdebug-cli daemon isAlive` (replaces `connection isAlive`)
  - `xdebug-cli daemon start` (existing, unchanged)
- All existing functionality preserved, only command paths change
- Update documentation to reflect unified daemon command structure
- Maintain backward compatibility in exit codes and JSON output formats

## Impact

**Affected specs:**
- `cli` - Remove Connection Command requirement, add daemon subcommands
- `persistent-sessions` - Update Enhanced Connection Commands requirement to reference daemon subcommands

**Affected code:**
- `internal/cli/connection.go` - Refactor into daemon subcommands
- `internal/cli/connection_test.go` - Update to test daemon subcommands
- `internal/cli/daemon.go` - Add new subcommands (status, list, kill, isAlive)
- `internal/cli/root.go` - Remove connectionCmd registration
- `CLAUDE.md` - Update all examples to use daemon subcommands

**Migration path:**
- Old: `xdebug-cli connection` → New: `xdebug-cli daemon status`
- Old: `xdebug-cli connection list` → New: `xdebug-cli daemon list`
- Old: `xdebug-cli connection kill` → New: `xdebug-cli daemon kill`
- Old: `xdebug-cli connection kill --all` → New: `xdebug-cli daemon kill --all`
- Old: `xdebug-cli connection isAlive` → New: `xdebug-cli daemon isAlive`

**User impact:**
- Simpler mental model: all daemon operations under one command
- More discoverable: `xdebug-cli daemon --help` shows all lifecycle operations
- Consistent with daemon-first architecture
- Users must update scripts and documentation to use new command paths

**Dependencies:**
- Should be applied after `remove-listen-command` change is completed
- No conflicts with other pending changes
