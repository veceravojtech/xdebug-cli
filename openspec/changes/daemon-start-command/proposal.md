# Proposal: Daemon Start Command

## Summary
Introduce a new `daemon start` command as the primary interface for starting persistent debug sessions, replacing the `listen --daemon` flag. This command provides better discoverability, clearer semantics, and improved defaults for daemon mode operations.

## Motivation
Currently, daemon mode is activated via `xdebug-cli listen --daemon`, which has several usability issues:

1. **Discoverability**: The `--daemon` flag is hidden within the `listen` command, making it harder for users to discover daemon functionality
2. **Inconsistent defaults**: Users must remember to add `--force` to kill stale daemons, creating friction
3. **Semantic clarity**: The `listen` command conceptually waits for a connection and exits, while daemon mode is fundamentally different (backgrounded, persistent)
4. **Command hierarchy**: Daemon operations (start, attach, kill) would be better grouped under a single `daemon` subcommand namespace

The new `daemon start` command addresses these issues by:
- Providing a dedicated, discoverable command for starting daemons
- Enabling `--force` behavior by default (always kills stale daemons)
- Creating a logical command hierarchy: `daemon start`, `daemon attach` (future), `daemon kill` (future)

## Scope
- **In Scope**:
  - Create new `daemon start` command that replaces `listen --daemon`
  - Remove `--daemon` flag from `listen` command
  - Make `--force` behavior default in `daemon start` (no flag needed)
  - Support `--commands` flag for initial breakpoints/commands
  - Support `-p`/`--port` flag to change listen port (inherits from global flags)
  - Default to listening on `0.0.0.0:9003`
- **Out of Scope**:
  - Restructuring `attach` command (remains as `xdebug-cli attach`)
  - Restructuring `connection` commands (remains as `xdebug-cli connection`)
  - Adding new daemon management features beyond current capabilities

## Impact Analysis
- **Breaking Change**: Yes - removes `--daemon` flag from `listen` command
  - Users currently using `xdebug-cli listen --daemon` must migrate to `xdebug-cli daemon start`
  - Migration path is straightforward: replace `listen --daemon` → `daemon start`
- **Behavioral Change**: Yes - `--force` is now implicit (always enabled)
  - Stale daemons are automatically killed without requiring explicit flag
  - This improves UX but changes default behavior
- **Documentation Impact**: All examples using `listen --daemon` need updating

## Files Affected
- `internal/cli/listen.go` - Remove `--daemon` flag and daemon mode logic
- `internal/cli/daemon.go` - New file for `daemon` parent command and `start` subcommand
- `internal/cli/root.go` - Register new `daemon` command
- `openspec/specs/cli/spec.md` - Update requirements to reflect new command structure
- `openspec/specs/persistent-sessions/spec.md` - Update daemon startup scenarios
- `CLAUDE.md` - Update all documentation examples
- `README.md` - Update command reference

## Alternatives Considered
1. **Keep `listen --daemon`**: Rejected - doesn't address discoverability or semantic clarity issues
2. **Add both `listen --daemon` and `daemon start`**: Rejected - creates two ways to do the same thing, confusing for users
3. **Make `--force` explicit in `daemon start`**: Rejected - forcing users to type `--force` every time adds unnecessary friction
4. **Create `daemon` parent with `attach`, `kill` subcommands too**: Deferred - out of scope for this change, but enables future restructuring

## Decision
Proceed with introducing `daemon start` command and removing `--daemon` flag from `listen`. This is a breaking change but provides significant UX improvements and creates a better foundation for future daemon management features.

## Migration Guide
**Before:**
```bash
xdebug-cli listen --daemon
xdebug-cli listen --daemon --commands "break :42"
xdebug-cli listen --daemon --force -p 9004
```

**After:**
```bash
xdebug-cli daemon start
xdebug-cli daemon start --commands "break :42"
xdebug-cli daemon start -p 9004  # --force is automatic
```
