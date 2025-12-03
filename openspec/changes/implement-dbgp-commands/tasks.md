# Tasks: Implement DBGp Protocol Commands

## Implementation Order

Tasks are ordered to deliver user-visible progress incrementally, with each task representing a small, verifiable unit of work.

### Phase 1: DBGp Client Foundation (Parallel: Tasks 1-6)

**Task 1: Implement Client.Status() method**
- Add `Status()` method to `internal/dbgp/client.go`
- Send `status -i {id}` command
- Parse response XML for status, reason, filename, lineno
- Return `*ProtocolResponse`
- **Validation**: Unit test verifies status command sent and response parsed correctly

**Task 2: Implement Client.Detach() method**
- Add `Detach()` method to `internal/dbgp/client.go`
- Send `detach -i {id}` command
- Update session state after detach
- Return `*ProtocolResponse`
- **Validation**: Unit test verifies detach command sent

**Task 3: Implement Client.Eval() method**
- Add `Eval(expression string)` method to `internal/dbgp/client.go`
- Base64-encode expression per DBGp protocol
- Send `eval -i {id} -- base64(expr)` command
- Parse `<property>` element from response for result value and type
- Return `*ProtocolResponse` with evaluation result
- **Validation**: Unit test with mock expression evaluation

**Task 4: Implement Client.SetProperty() method**
- Add `SetProperty(name, value, dataType string)` method to `internal/dbgp/client.go`
- Base64-encode value
- Calculate data length
- Send `property_set -i {id} -n {name} -t {type} -l {len} -- base64(value)`
- Parse success/failure response
- **Validation**: Unit test for integer and string property set

**Task 5: Implement Client.GetSource() method**
- Add `GetSource(fileURI string, beginLine, endLine int)` method to `internal/dbgp/client.go`
- Build command with optional `-f`, `-b`, `-e` parameters
- Send `source -i {id} ...` command
- Decode base64-encoded source from response
- Return `*ProtocolResponse` with decoded source
- **Validation**: Unit test verifies base64 decoding and line range parameters

**Task 6: Implement Client.UpdateBreakpoint() method**
- Add `UpdateBreakpoint(breakpointID, state string)` method to `internal/dbgp/client.go`
- Validate state is "enabled" or "disabled"
- Send `breakpoint_update -i {id} -d {bid} -s {state}`
- Return `*ProtocolResponse`
- **Validation**: Unit test for enable and disable operations

### Phase 2: CLI Command Handlers - Execution Control (Sequential: Tasks 7-9)

**Task 7: Implement handleStatus() command handler**
- Add `handleStatus(client, view, args)` to `internal/cli/listen.go`
- Call `client.Status()`
- Display status, reason, file, and line in human-readable format
- Handle errors gracefully
- Add case "status", "st" to dispatcher switch
- **Validation**: Manual test: `xdebug-cli listen --commands "status"`
- **Dependencies**: Task 1

**Task 8: Implement handleDetach() command handler**
- Add `handleDetach(client, view, args)` to `internal/cli/listen.go`
- Call `client.Detach()`
- Display detach confirmation
- Return `true` to signal session end
- Add case "detach", "d" to dispatcher switch
- **Validation**: Manual test: `xdebug-cli listen --commands "run" "detach"`
- **Dependencies**: Task 2

**Task 9: Implement handleEval() command handler**
- Add `handleEval(client, view, args)` to `internal/cli/listen.go`
- Join args into single expression string
- Validate expression not empty
- Call `client.Eval(expression)`
- Display result value and type (similar to print command format)
- Handle evaluation errors
- Add case "eval", "e" to dispatcher switch
- **Validation**: Manual test: `xdebug-cli listen --commands "break :42" "run" "eval \$x + 10"`
- **Dependencies**: Task 3

### Phase 3: CLI Command Handlers - Variable & Source (Sequential: Tasks 10-11)

**Task 10: Implement handleSet() command handler**
- Add `handleSet(client, view, args)` to `internal/cli/listen.go`
- Parse args to extract variable name and value from `$var = value` syntax
- Implement type detection logic (int, float, string, bool)
- Call `client.SetProperty(name, value, type)`
- Display confirmation with variable name and new value
- Handle parse errors and show usage help
- Add case "set" to dispatcher switch
- **Validation**: Manual test: `xdebug-cli listen --commands "break :42" "run" "set \$count = 100" "print \$count"`
- **Dependencies**: Task 4

**Task 11: Implement handleSource() command handler**
- Add `handleSource(client, view, args)` to `internal/cli/listen.go`
- Parse args for optional file path and line range (`:10-20` or `file.php:10-20`)
- Default to current file if no file specified
- Normalize file path to file:// URI using existing `normalizeFileURI()` logic
- Call `client.GetSource(uri, start, end)`
- Display source with line numbers (reuse formatting from `list` command)
- Highlight current line if within displayed range
- Add case "source", "src" to dispatcher switch
- **Validation**: Manual test: `xdebug-cli listen --commands "run" "source" "source :10-20"`
- **Dependencies**: Task 5

### Phase 4: CLI Command Handlers - Breakpoints (Sequential: Tasks 12-15)

**Task 12: Implement handleDelete() command handler**
- Add `handleDelete(client, view, args)` to `internal/cli/listen.go`
- Validate breakpoint ID provided in args
- Validate ID is numeric
- Call existing `client.RemoveBreakpoint(id)`
- Display confirmation
- Handle errors (breakpoint not found)
- Add case "delete", "del" to dispatcher switch
- **Validation**: Manual test: `xdebug-cli listen --commands "break :42" "info b" "delete <ID>" "info b"`
- **Dependencies**: None (uses existing RemoveBreakpoint)

**Task 13: Implement handleDisable() command handler**
- Add `handleDisable(client, view, args)` to `internal/cli/listen.go`
- Validate breakpoint ID provided
- Call `client.UpdateBreakpoint(id, "disabled")`
- Display confirmation
- Handle errors
- Add case "disable" to dispatcher switch
- **Validation**: Manual test: `xdebug-cli listen --commands "break :42" "disable <ID>" "run"` (should not stop)
- **Dependencies**: Task 6

**Task 14: Implement handleEnable() command handler**
- Add `handleEnable(client, view, args)` to `internal/cli/listen.go`
- Validate breakpoint ID provided
- Call `client.UpdateBreakpoint(id, "enabled")`
- Display confirmation
- Handle errors
- Add case "enable" to dispatcher switch
- **Validation**: Manual test: `xdebug-cli listen --commands "break :42" "disable <ID>" "enable <ID>" "run"` (should stop)
- **Dependencies**: Task 6

**Task 15: Implement handleStack() command handler**
- Add `handleStack(client, view, args)` to `internal/cli/listen.go`
- Call existing `client.GetStackTrace()`
- Reuse stack display formatting from `handleInfo()`
- Add case "stack" to dispatcher switch
- **Validation**: Manual test: `xdebug-cli listen --commands "break :42" "run" "stack"`
- **Dependencies**: None (uses existing GetStackTrace)

### Phase 5: Daemon Mode Integration (Parallel: Tasks 16-23)

**Task 16: Add status command to daemon executor**
- Add case "status", "st" to `internal/daemon/executor.go` switch
- Implement handler calling `client.Status()`
- Return `ipc.CommandResult` with JSON-serializable status data
- **Validation**: Test `xdebug-cli daemon start --commands "run"` then `xdebug-cli attach --commands "status"`
- **Dependencies**: Task 7

**Task 17: Add detach command to daemon executor**
- Add case "detach", "d" to daemon executor switch
- Implement handler calling `client.Detach()`
- Return `ipc.CommandResult`
- **Dependencies**: Task 8

**Task 18: Add eval command to daemon executor**
- Add case "eval", "e" to daemon executor switch
- Implement handler calling `client.Eval()`
- Return `ipc.CommandResult` with evaluation result
- **Dependencies**: Task 9

**Task 19: Add set command to daemon executor**
- Add case "set" to daemon executor switch
- Reuse parsing logic from Task 10
- Return `ipc.CommandResult`
- **Dependencies**: Task 10

**Task 20: Add source command to daemon executor**
- Add case "source", "src" to daemon executor switch
- Reuse parsing logic from Task 11
- Return `ipc.CommandResult` with source data
- **Dependencies**: Task 11

**Task 21: Add delete command to daemon executor**
- Add case "delete", "del" to daemon executor switch
- Return `ipc.CommandResult`
- **Dependencies**: Task 12

**Task 22: Add disable command to daemon executor**
- Add case "disable" to daemon executor switch
- Return `ipc.CommandResult`
- **Dependencies**: Task 13

**Task 23: Add enable command to daemon executor**
- Add case "enable" to daemon executor switch
- Return `ipc.CommandResult`
- **Dependencies**: Task 14

**Task 24: Add stack command to daemon executor**
- Add case "stack" to daemon executor switch
- Return `ipc.CommandResult` with stack trace data
- **Dependencies**: Task 15

### Phase 6: JSON Output Support (Parallel: Tasks 25-32)

**Task 25: Implement JSON output for status command**
- Update `handleStatus()` in listen.go to check JSON mode
- Format JSON output: `{"command":"status","success":true,"result":{...}}`
- Update attach.go display function to handle status JSON
- **Validation**: Test `xdebug-cli listen --json --commands "status"`
- **Dependencies**: Task 7, Task 16

**Task 26: Implement JSON output for detach command**
- Update `handleDetach()` for JSON mode
- Format: `{"command":"detach","success":true}`
- **Dependencies**: Task 8, Task 17

**Task 27: Implement JSON output for eval command**
- Update `handleEval()` for JSON mode
- Format: `{"command":"eval","success":true,"result":{"expression":"...","type":"...","value":"..."}}`
- **Dependencies**: Task 9, Task 18

**Task 28: Implement JSON output for set command**
- Update `handleSet()` for JSON mode
- Format: `{"command":"set","success":true,"result":{"variable":"...","value":"...","type":"..."}}`
- **Dependencies**: Task 10, Task 19

**Task 29: Implement JSON output for source command**
- Update `handleSource()` for JSON mode
- Format: `{"command":"source","success":true,"result":{"file":"...","start_line":N,"end_line":M,"source":"..."}}`
- **Dependencies**: Task 11, Task 20

**Task 30: Implement JSON output for delete command**
- Update `handleDelete()` for JSON mode
- Format: `{"command":"delete","success":true,"result":{"breakpoint_id":"..."}}`
- **Dependencies**: Task 12, Task 21

**Task 31: Implement JSON output for disable/enable commands**
- Update `handleDisable()` and `handleEnable()` for JSON mode
- Format: `{"command":"disable","success":true,"result":{"breakpoint_id":"...","state":"..."}}`
- **Dependencies**: Tasks 13-14, Tasks 22-23

**Task 32: Implement JSON output for stack command**
- Update `handleStack()` for JSON mode
- Format: `{"command":"stack","success":true,"result":[{"depth":N,"function":"...","file":"...","line":N}]}`
- **Dependencies**: Task 15, Task 24

### Phase 7: Help Text & Documentation (Sequential: Tasks 33-35)

**Task 33: Update help text in internal/view/help.go**
- Add entries for new commands: status, detach, eval, set, source, stack, delete, disable, enable
- Include command syntax and brief description
- Add examples showing usage
- **Validation**: Run `xdebug-cli listen --commands "help"` and verify new commands listed

**Task 34: Update CLAUDE.md documentation**
- Add new commands to "Debugging Commands" table
- Add syntax and examples for each command
- Document aliases
- Add usage examples in "Command-Based Execution" section
- **Validation**: Review documentation for completeness and accuracy

**Task 35: Add command reference to sources/commands.md**
- Add implementation status table showing which DBGp commands are now supported
- Map CLI commands to DBGp protocol commands
- Note any limitations or known issues
- **Validation**: Cross-reference with DBGp protocol documentation

### Phase 8: Testing & Validation (Sequential: Tasks 36-37)

**Task 36: Write unit tests for new client methods**
- Test `Status()` with mock DBGp responses
- Test `Detach()` command sending
- Test `Eval()` with base64 encoding validation
- Test `SetProperty()` with different data types
- Test `GetSource()` with line range parameters
- Test `UpdateBreakpoint()` state changes
- **Validation**: Run `go test ./internal/dbgp/...` - all tests pass

**Task 37: Write integration tests for command flow**
- Test full command execution: parse → client → response → display
- Test error handling for each command
- Test JSON output format matches specification
- Test daemon mode command execution
- **Validation**: Run `go test ./internal/cli/...` and `go test ./internal/daemon/...` - all tests pass

## Parallelization Opportunities

- **Phase 1 (Tasks 1-6)**: All client methods can be implemented in parallel
- **Phase 5 (Tasks 16-24)**: Daemon executor cases can be added in parallel after Phase 2-4 complete
- **Phase 6 (Tasks 25-32)**: JSON output can be implemented in parallel for each command

## Dependencies Summary

```
Phase 1 (Tasks 1-6) → Phase 2 (Tasks 7-9) → Phase 5 subset (Tasks 16-18)
                    → Phase 3 (Tasks 10-11) → Phase 5 subset (Tasks 19-20)
                    → Phase 4 (Tasks 12-15) → Phase 5 subset (Tasks 21-24)

Phase 5 → Phase 6 (Tasks 25-32)

Phase 6 → Phase 7 (Tasks 33-35) → Phase 8 (Tasks 36-37)
```

## Total Tasks: 37

Estimated completion: 37 small, focused tasks delivering incremental user-visible functionality throughout implementation.
