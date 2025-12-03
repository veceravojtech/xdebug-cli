# Design: Remove Interactive REPL Mode

## Context
The xdebug-cli currently supports two modes for the `listen` command:
1. **Interactive mode** - REPL loop that reads commands from stdin, displays prompts
2. **Non-interactive mode** - Executes commands from `--commands` flag and exits

Additionally, daemon mode allows background sessions with attach-based command execution, which already provides the benefits of multi-step debugging without requiring an interactive prompt.

The interactive REPL was built as a familiar debugging interface, but in practice:
- Automation and scripting use cases require non-interactive mode
- Multi-step workflows are better served by daemon mode (persistent sessions)
- The REPL adds complexity with minimal real-world benefit
- Documentation and testing burden increases with dual modes

## Goals / Non-Goals

### Goals
- Remove all interactive REPL code paths from the codebase
- Make command-based execution the only way to use `listen` command
- Simplify View layer by removing stdin input handling
- Maintain full backwards compatibility with daemon mode and attach commands
- Provide clear migration guidance for existing users

### Non-Goals
- Changing daemon mode or attach command behavior
- Removing JSON output support
- Modifying DBGp protocol layer
- Changing command syntax or available debugging operations

## Decisions

### Decision 1: Remove --non-interactive flag entirely
**Rationale**: Since non-interactive becomes the only mode, the flag is redundant. Users simply provide `--commands` to specify what to execute.

**Alternative considered**: Keep flag as no-op for backwards compatibility
- Rejected because it creates confusion and leaves dead code

### Decision 2: Make --commands mandatory for listen
**Rationale**: Without interactive input, there must be predefined commands to execute. This makes the contract explicit.

**Alternative considered**: Allow empty commands and just wait for connection
- Rejected because it would leave the session in an unusable state (no way to issue commands without REPL or daemon mode)

**Exception**: Daemon mode (`--daemon` flag) has different behavior and doesn't require `--commands` to be mandatory, since daemon sessions accept commands via `attach`.

### Decision 3: Keep daemon mode completely unchanged
**Rationale**: Daemon mode with attach already provides the multi-step debugging workflow users need. It's orthogonal to the interactive/non-interactive distinction.

**Benefits**:
- Migration path for users who need multi-step debugging
- No breaking changes to daemon functionality
- Clean separation of concerns

### Decision 4: Remove stdin handling from View
**Rationale**: View was designed to abstract terminal I/O. With no interactive input, stdin buffering and `GetInputLine()` become unnecessary.

**Impact**:
- Remove `stdin *bufio.Reader` field from View struct
- Remove `GetInputLine()` method
- Remove `PrintInputPrefix()` method (no prompt needed)
- Keep all output methods (PrintLn, PrintErrorLn, etc.) for command results

### Decision 5: Flag validation changes
Current validation:
```go
// --commands requires --non-interactive or --daemon
if len(Commands) > 0 && !NonInteractive && !Daemon {
    return error
}
```

New validation:
```go
// listen requires --commands unless in --daemon mode
if !Daemon && len(Commands) == 0 {
    return error("listen command requires --commands flag (or use --daemon mode)")
}
```

## Risks / Trade-offs

### Risk: Breaking change for interactive users
**Impact**: Users who rely on `xdebug-cli listen` entering a REPL will experience breakage

**Mitigation**:
- Clear error message when `--commands` is missing
- Updated documentation with migration examples
- Daemon mode provides equivalent functionality for exploration

**Example error message**:
```
Error: listen command requires --commands flag

Example usage:
  xdebug-cli listen --commands "run" "print \$var"

For multi-step debugging, use daemon mode:
  xdebug-cli listen --daemon
  xdebug-cli attach --commands "run"
  xdebug-cli attach --commands "print \$var"
```

### Trade-off: Less discoverable for beginners
Interactive REPL provides discoverability (user can type `help`, explore commands)

**Mitigation**:
- CLI-level `--help` still available
- Documentation provides comprehensive examples
- Target audience is automation/tooling, not beginner exploration

## Migration Plan

### For users running: `xdebug-cli listen`
**Old behavior**: Enters interactive REPL, waits for user commands

**New behavior**: Error - must provide `--commands`

**Migration**:
```bash
# Option 1: Non-interactive mode (single invocation)
xdebug-cli listen --commands "break :42" "run" "context local"

# Option 2: Daemon mode (multiple invocations)
xdebug-cli listen --daemon --commands "break :42"
xdebug-cli attach --commands "run"
xdebug-cli attach --commands "context local"
```

### For users running: `xdebug-cli listen --non-interactive --commands "..."`
**Old behavior**: Executes commands and exits (non-interactive mode)

**New behavior**: Same, but drop `--non-interactive` flag

**Migration**:
```bash
# Before
xdebug-cli listen --non-interactive --commands "run"

# After (--non-interactive removed, same behavior)
xdebug-cli listen --commands "run"
```

### For daemon mode users: `xdebug-cli listen --daemon`
**No migration needed** - daemon mode unchanged

## Implementation Sequence

1. **Update flag validation** - Change to require `--commands` unless `--daemon`
2. **Remove REPL loop** - Delete `replLoop()` function from listen.go
3. **Simplify command flow** - Remove branching logic between interactive/non-interactive
4. **Clean View layer** - Remove stdin, GetInputLine, PrintInputPrefix
5. **Update help messages** - Remove REPL-specific help text
6. **Update tests** - Remove interactive mode tests, update remaining tests
7. **Update documentation** - CLAUDE.md, README.md remove REPL sections

## Open Questions

None - scope is well-defined and migration path is clear.
