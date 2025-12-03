# Proposal: Implement DBGp Protocol Commands

## Change ID
`implement-dbgp-commands`

## Problem Statement
The xdebug-cli currently implements only a subset of DBGp protocol commands documented in `sources/commands.md`. Many essential debugging capabilities are missing, limiting the tool's usefulness for advanced debugging workflows:

- **Missing execution control**: No `status` or `detach` commands
- **Missing code evaluation**: No `eval` command for runtime PHP execution
- **Limited breakpoint management**: Cannot update, disable, or remove individual breakpoints
- **No variable modification**: Cannot use `property_set` to change values during debugging
- **Missing source display**: Cannot fetch source code from debugger using `source` command
- **No standalone stack command**: Stack trace only available via `info stack`

Users who need these capabilities must use alternative debugging tools or manually craft DBGp protocol messages.

## Proposed Solution
Implement all essential DBGp protocol commands from the reference documentation to provide complete debugging functionality. This change will:

1. **Add new CLI commands**: `status`, `detach`, `eval`, `stack`, `delete`, `disable`, `enable`, `set`, `source`
2. **Extend DBGp client**: Add missing protocol methods to `internal/dbgp/client.go`
3. **Update command dispatcher**: Integrate new commands into `internal/cli/listen.go` and `internal/daemon/executor.go`
4. **Maintain consistency**: Ensure all commands work in both listen mode and daemon/attach mode with JSON output support

## Success Criteria
- All commands from `sources/commands.md` marked as "Core Commands (Required Support)" are implemented
- New commands available via `xdebug-cli listen --commands` and `xdebug-cli attach --commands`
- JSON output mode (`--json`) works for all new commands
- Documentation updated to reflect new capabilities
- Tests cover new command functionality

## Scope
**In Scope:**
- CLI command handlers for: `status`, `detach`, `eval`, `stack`, `delete`, `disable`, `enable`, `set`, `source`
- DBGp client methods for: `Status()`, `Detach()`, `Eval()`, `SetProperty()`, `GetSource()`, `UpdateBreakpoint()`
- Command parsing and validation
- JSON output formatting for new commands
- Help text and documentation updates

**Out of Scope:**
- Extended/optional commands (stdin, stdout, stderr redirection)
- Spawnpoint commands
- `interact` command (interactive console mode)
- Notification system
- `break` command (interrupt running execution)
- Feature get/set commands (advanced configuration)

## Dependencies
- Existing DBGp protocol infrastructure (`internal/dbgp/`)
- Current command dispatch architecture (`internal/cli/listen.go`, `internal/daemon/executor.go`)
- View layer for output formatting (`internal/view/`)

## Risks and Mitigations
**Risk**: Some DBGp commands may not be supported by all Xdebug versions
**Mitigation**: Document minimum Xdebug version requirements; handle unsupported command errors gracefully

**Risk**: Command syntax complexity may confuse users
**Mitigation**: Provide clear help text with examples; follow existing command patterns

**Risk**: Variable modification (`set`) could lead to unexpected debugging behavior
**Mitigation**: Include warnings in help text; require explicit variable name syntax

## Related Changes
None currently. Future changes may add:
- Stream redirection support
- Feature negotiation capabilities
- Advanced breakpoint hit conditions
