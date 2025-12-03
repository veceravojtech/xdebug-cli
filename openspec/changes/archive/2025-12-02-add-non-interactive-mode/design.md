# Design: Non-Interactive Mode for CLI Debugging

## Context
The xdebug-cli currently only supports interactive REPL debugging. Users (especially LLM agents like Claude, CI/CD systems, and automation scripts) need to execute debugging operations programmatically without manual input.

**Constraints:**
- Must maintain backward compatibility (interactive mode is default)
- Should work with existing DBGp client/server infrastructure
- Output must be parseable by LLMs and scripts

**Stakeholders:**
- LLM agents (Claude) - primary use case for automated debugging
- DevOps engineers - CI/CD integration
- Script authors - automated testing and debugging workflows

## Goals / Non-Goals

**Goals:**
- Enable command-line driven debugging without REPL
- Provide structured output for LLM/script consumption
- Support JSON output for machine parsing
- Maintain all existing debugging capabilities in non-interactive mode

**Non-Goals:**
- Stdin piping or script file input (deferred to future if needed)
- Websocket/API server mode
- Multi-session management
- Interactive mode modifications (keep existing behavior unchanged)

## Decisions

### Decision 1: Command Input via CLI Arguments
**Choice:** Use `--commands` flag with variadic string arguments
```bash
xdebug-cli listen --non-interactive --commands "run" "step" "print $x"
```

**Rationale:**
- Most explicit and clear for LLM agents
- Shell-friendly (no file creation needed)
- Easier to debug (visible in process list)
- Aligns with user preference for "Claude can use CLI better"

**Alternatives considered:**
- Stdin piping: More Unix-like but harder for LLMs to construct
- Script file: Requires file management, adds complexity
- Environment variable: Poor UX, hard to escape special characters

### Decision 2: Mode Trigger via Flag
**Choice:** Add `--non-interactive` flag to existing `listen` command

**Rationale:**
- Minimal API surface (reuse existing command)
- Clear opt-in behavior
- Easier migration path

**Alternatives considered:**
- New `debug` command: More separation but API fragmentation
- Auto-detect: Too magical, harder to debug

### Decision 3: JSON Output Structure
**Choice:** Each command produces a JSON object with standardized fields:
```json
{
  "command": "run|step|print|...",
  "success": true|false,
  "error": "error message if success=false",
  "result": {
    // Command-specific structured data
  }
}
```

**Rationale:**
- Consistent structure across all commands
- `success` field enables error detection without parsing
- `result` field allows command-specific data
- LLM-friendly: clear context for each operation

**Alternatives considered:**
- Plain text with markers: Harder to parse reliably
- One JSON array: Loses per-command context
- NDJSON stream: Over-engineering for single-connection use case

### Decision 4: Output Mode Flag Placement
**Choice:** Make `--json` a global flag on root command

**Rationale:**
- Works across all commands (future-proof)
- Standard practice in CLI tools
- Easier for users to remember

## Risks / Trade-offs

### Risk: Command Escaping Complexity
Shell escaping of special characters in commands (e.g., `print $x`, quotes in strings)

**Mitigation:**
- Document escaping requirements clearly
- Provide examples in CLAUDE.md for common cases
- Consider adding validation/helpful error messages

### Risk: Sequence Length Limits
Very long command sequences may exceed shell argument limits

**Mitigation:**
- Start with CLI args (covers 99% of use cases)
- Document limit in help text
- Can add script file support in future if needed

### Trade-off: No Intermediate Output
In non-interactive mode, users can't see intermediate state during execution

**Acceptance:**
- JSON output provides all relevant state
- Use case is automation, not human debugging
- Interactive mode remains available for exploratory debugging

## Migration Plan

**Phase 1: Add feature (this change)**
1. Add flags to existing command
2. Implement non-interactive execution path
3. Add JSON output formatters
4. Update tests and documentation

**Phase 2: User adoption**
- No breaking changes, so existing workflows unaffected
- Users opt-in by adding `--non-interactive` flag

**Rollback:**
- If issues arise, remove flags and revert changes
- No data migration needed
- No breaking changes to rollback

## Open Questions

1. **Should we limit command sequence length?**
   - Defer to implementation; can add validation if becomes issue

2. **Should JSON output be pretty-printed or compact?**
   - Start with compact (one line per command result)
   - Can add `--pretty` flag later if needed

3. **Should we support command aliases (r, s, n) in non-interactive mode?**
   - Yes, same command parser should work
   - Maintains consistency with interactive mode
