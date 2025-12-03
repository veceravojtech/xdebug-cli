# Change: Add Standard Debugger Command Aliases

## Why

Users coming from different debugging backgrounds (GDB, DBGp protocol, VS Code, other debuggers) naturally try command names from their familiar tools. Currently these fail with "Unknown command" errors, creating friction and requiring users to learn xdebug-cli's specific vocabulary.

Real-world usage shows users attempting:
- `continue` instead of `run`
- `into`/`over`/`step_out`/`step_into` instead of `step`/`next`/`out`
- `breakpoint_list`, `breakpoint_remove` (DBGp protocol style)
- `property_get -n $var` (DBGp protocol style)
- `clear file:line` (GDB style for removing breakpoints by location)

## What Changes

Add command aliases to support multiple debugger naming conventions without breaking existing commands:

- **Execution control aliases:** `continue`, `cont`, `c` → `run`; `into` → `step`; `over` → `next`; `step_out` → `out`; `step_into` → `step`
- **DBGp protocol style:** `breakpoint_list` → `info breakpoints`; `breakpoint_remove` → `delete`; `property_get -n <var>` → `print <var>`
- **GDB style:** `clear <location>` - new command to delete breakpoint by file:line location instead of ID

All existing commands continue to work unchanged. This is purely additive - no breaking changes.

## Impact

- Affected specs: `cli` (new debugging command requirements)
- Affected code:
  - `internal/daemon/executor.go` - add cases to `executeCommand()` switch
  - `internal/view/help.go` - update help text to show aliases
  - `CLAUDE.md` - document command aliases
- Breaking changes: None
- Migration: None required (backward compatible)
