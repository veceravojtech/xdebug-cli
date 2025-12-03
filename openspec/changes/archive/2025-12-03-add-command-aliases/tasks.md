# Implementation Tasks

## 1. Add Command Aliases to Executor

- [x] 1.1 Add simple execution control aliases to `executeCommand()` switch in `internal/daemon/executor.go`
  - [x] 1.1.1 Add `continue`, `cont`, `c` cases that call `handleRun()`
  - [x] 1.1.2 Add `into` case that calls `handleStep()`
  - [x] 1.1.3 Add `over` case that calls `handleNext()`
  - [x] 1.1.4 Add `step_out` case that calls `handleStepOut()`
  - [x] 1.1.5 Add `step_into` case that calls `handleStep()`
- [x] 1.2 Add breakpoint management aliases
  - [x] 1.2.1 Add `breakpoint_list` case that calls `handleInfo([]string{"breakpoints"})`
  - [x] 1.2.2 Add `breakpoint_remove` case that calls `handleDelete(args)`
- [x] 1.3 Implement `property_get` command handler
  - [x] 1.3.1 Create `handlePropertyGet(args)` function that parses `-n <var>` syntax
  - [x] 1.3.2 Extract variable name after `-n` flag
  - [x] 1.3.3 Call `handlePrint()` with extracted variable name
  - [x] 1.3.4 Return error if `-n` flag missing or no variable specified
  - [x] 1.3.5 Add `property_get` case to switch statement
- [x] 1.4 Implement `clear` command handler
  - [x] 1.4.1 Create `handleClear(args)` function
  - [x] 1.4.2 Parse location argument (`:line` or `file:line` format)
  - [x] 1.4.3 Call `client.GetBreakpointList()` to fetch all breakpoints
  - [x] 1.4.4 Filter breakpoints matching the parsed file and line
  - [x] 1.4.5 Call `client.RemoveBreakpoint()` for each matching breakpoint
  - [x] 1.4.6 Return success with count of removed breakpoints
  - [x] 1.4.7 Return error if no breakpoint found at location
  - [x] 1.4.8 Add `clear` case to switch statement

## 2. Update Help Text

- [x] 2.1 Update `handleHelp()` in `internal/daemon/executor.go` to show aliases
- [x] 2.2 Update `ShowStepHelpMessage()` in `internal/view/help.go` to include new aliases
- [x] 2.3 Update `ShowCommandHelp()` in `internal/view/help.go` to recognize new command names
- [x] 2.4 Update help output in `internal/cli/attach.go` displayCommandResult() for new commands

## 3. Update Documentation

- [x] 3.1 Add "Command Aliases" section to CLAUDE.md
  - [x] 3.1.1 Document GDB-style commands (continue, clear)
  - [x] 3.1.2 Document DBGp protocol commands (property_get, breakpoint_list, breakpoint_remove)
  - [x] 3.1.3 Document alternative names (into, over, step_into, step_out)
  - [x] 3.1.4 Include examples showing both old and new command styles
- [x] 3.2 Update main debugging commands section to show primary commands with aliases

## 4. Add Tests

- [x] 4.1 Add unit tests for new command aliases in `internal/daemon/executor_test.go` (create if needed)
  - [x] 4.1.1 Test `continue` command maps to run behavior
  - [x] 4.1.2 Test `into`, `over`, `step_out`, `step_into` map correctly
  - [x] 4.1.3 Test `breakpoint_list` returns breakpoint list
  - [x] 4.1.4 Test `breakpoint_remove` deletes breakpoint by ID
- [x] 4.2 Add tests for `property_get` command
  - [x] 4.2.1 Test `property_get -n $var` extracts variable correctly
  - [x] 4.2.2 Test error when `-n` flag missing
  - [x] 4.2.3 Test error when variable name missing after `-n`
- [x] 4.3 Add tests for `clear` command
  - [x] 4.3.1 Test `clear :42` removes breakpoint at line 42 in current file
  - [x] 4.3.2 Test `clear file.php:100` removes breakpoint at specific location
  - [x] 4.3.3 Test error when no breakpoint exists at location
  - [x] 4.3.4 Test removes multiple breakpoints at same location

## 5. Integration Testing

- [x] 5.1 Test command aliases in daemon workflow
  - [x] 5.1.1 Start daemon, use `continue` instead of `run`
  - [x] 5.1.2 Test `breakpoint_list` shows breakpoints
  - [x] 5.1.3 Test `property_get -n $var` inspects variables
  - [x] 5.1.4 Test `clear` removes breakpoints by location
- [x] 5.2 Verify JSON output mode works with new commands
- [x] 5.3 Verify error messages are clear for invalid usage

## 6. Verification

- [x] 6.1 Run `go test ./...` - all tests pass
- [x] 6.2 Run `go build ./cmd/xdebug-cli` - builds successfully
- [x] 6.3 Manual test: try all commands from the original error table
- [x] 6.4 Run `openspec validate add-command-aliases --strict` - passes validation
