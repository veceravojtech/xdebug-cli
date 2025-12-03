# Change: Remove Interactive REPL Mode

## Why
The interactive REPL mode is legacy functionality that is no longer needed. All use cases are better served by non-interactive mode (for automation, scripting, CI/CD) and daemon mode with attach (for multi-step debugging workflows). Removing the REPL simplifies the codebase, eliminates maintenance burden, and provides a more focused tool designed for automation and programmatic debugging.

## What Changes
- **BREAKING**: Remove interactive REPL loop (`replLoop` function) from `listen` command
- **BREAKING**: Remove `--non-interactive` flag (non-interactive becomes the only mode)
- **BREAKING**: Make `--commands` flag mandatory for `listen` command
- **BREAKING**: Remove REPL-specific help messages and prompts
- Remove `GetInputLine()` and `PrintInputPrefix()` methods from View
- Remove input buffer (`stdin *bufio.Reader`) from View struct
- Remove help command handler (help becomes CLI-level `--help` only)
- Remove quit command handler (session ends when commands complete)
- Update CLI long description to reflect command-based usage only
- Simplify listen command flow (no REPL vs non-interactive branching)
- Update all documentation to remove REPL examples

The change keeps daemon mode and attach command functionality unchanged. Users who need multi-step debugging workflows should use daemon mode.

## Impact
- **Affected specs**: `cli`, `view`
- **Affected code**:
  - `internal/cli/listen.go` - Remove `replLoop`, simplify command flow
  - `internal/view/view.go` - Remove `GetInputLine`, `PrintInputPrefix`, stdin field
  - `internal/view/help.go` - Remove REPL-specific help for help/quit commands
  - `CLAUDE.md` - Remove interactive REPL section, update available commands
  - `README.md` - Remove interactive mode examples (if exists)
- **Breaking changes**:
  - Users who run `xdebug-cli listen` without `--commands` will get an error
  - Users relying on interactive prompt must migrate to either:
    - Non-interactive mode: `xdebug-cli listen --commands "run" "step"`
    - Daemon mode: `xdebug-cli listen --daemon && xdebug-cli attach --commands "run"`
- **Migration path**: All interactive REPL use cases map cleanly to non-interactive or daemon mode
