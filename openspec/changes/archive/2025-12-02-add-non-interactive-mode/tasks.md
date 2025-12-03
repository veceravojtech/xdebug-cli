## 1. Add Flags and Configuration
- [x] 1.1 Add `--non-interactive` flag to listen command in `internal/cli/listen.go`
- [x] 1.2 Add `--commands` flag (accepts string array) to listen command
- [x] 1.3 Add global `--json` flag to root command in `internal/cli/root.go`
- [x] 1.4 Add flag validation (error if `--commands` without `--non-interactive`)
- [x] 1.5 Update CLIParameter struct in `internal/cfg/` to hold new flags

## 2. Implement Non-Interactive Command Execution
- [x] 2.1 Create `executeNonInteractive()` function in `internal/cli/listen.go`
- [x] 2.2 Modify `listenAccept()` to branch between REPL and non-interactive based on flag
- [x] 2.3 Implement sequential command processing from `--commands` array
- [x] 2.4 Add error handling with appropriate exit codes (0 = success, 1+ = error)
- [x] 2.5 Suppress interactive prompts and banners in non-interactive mode

## 3. Add JSON Output Support
- [x] 3.1 Create `internal/view/json.go` with JSON formatter structs
- [x] 3.2 Add `OutputJSON()` method to View for structured output
- [x] 3.3 Update `handleRun()`, `handleStep()`, `handleNext()` to support JSON output
- [x] 3.4 Update `handlePrint()` to output JSON variable structure
- [x] 3.5 Update `handleContext()` to output JSON variable list
- [x] 3.6 Update `handleInfo()` to output JSON breakpoint list
- [x] 3.7 Add error response JSON format `{"success": false, "error": "..."}`

## 4. Testing
- [x] 4.1 Write unit tests for non-interactive flag parsing
- [x] 4.2 Write unit tests for command sequence execution
- [x] 4.3 Write unit tests for JSON output formatters
- [x] 4.4 Write integration test for `listen --non-interactive --commands "run"`
- [x] 4.5 Write integration test for JSON output mode
- [x] 4.6 Test error handling and exit codes
- [x] 4.7 Verify backward compatibility (interactive mode still works)

## 5. Documentation
- [x] 5.1 Update CLAUDE.md with non-interactive mode examples
- [x] 5.2 Update help text in listen command with new flags
- [x] 5.3 Add usage examples for automation and scripting
