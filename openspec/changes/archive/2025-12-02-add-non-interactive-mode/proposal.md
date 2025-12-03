# Change: Add Non-Interactive Mode for CLI Debugging

## Why
Currently, `xdebug-cli listen` always starts an interactive REPL after establishing a connection, making it impossible to automate debugging operations via scripts, CI/CD pipelines, or LLM agents like Claude. This forces users to manually type commands even for repetitive debugging tasks.

## What Changes
- Add `--non-interactive` flag to `listen` command that accepts debugging commands as CLI arguments
- Process commands sequentially without REPL prompt
- Provide structured output optimized for LLM/script consumption (clear context for each operation)
- Exit automatically after all commands complete or on error
- Support JSON output format with `--json` flag for machine parsing
- Maintain backward compatibility (default behavior remains interactive REPL)

## Impact
- **Affected specs**: `cli` (listen command, new flags, output modes)
- **Affected code**:
  - `internal/cli/listen.go` - Add flag parsing and command execution logic
  - `internal/cli/root.go` - Add global `--json` flag
  - `internal/view/` - Add JSON output formatters
- **Breaking changes**: None (new optional flag)
- **New capabilities**: Enables scripting, automation, CI/CD integration, and LLM agent debugging
