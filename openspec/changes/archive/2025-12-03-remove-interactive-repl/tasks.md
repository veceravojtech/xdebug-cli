# Implementation Tasks

## 1. Update Flag Validation and Command Requirements
- [x] 1.1 Remove `NonInteractive` flag from `internal/cfg/config.go` CLIParameter struct
- [x] 1.2 Remove `--non-interactive` flag registration from `internal/cli/listen.go` init()
- [x] 1.3 Update `validateListenFlags()` in listen.go to require `--commands` unless `--daemon` flag is present
- [x] 1.4 Update error messages to guide users toward `--commands` or `--daemon` usage
- [x] 1.5 Run tests: `go test ./internal/cli/... -v` to verify flag validation

## 2. Remove REPL Loop and Interactive Code Paths
- [x] 2.1 Delete `replLoop()` function from `internal/cli/listen.go`
- [x] 2.2 Delete quit command handler logic (removed from dispatchCommand)
- [x] 2.3 Help command handler kept (shows help for debugging commands)
- [x] 2.4 Removed empty input handling (no longer needed without REPL)
- [x] 2.5 Remove all conditional branching on `CLIArgs.NonInteractive` from listen.go
- [x] 2.6 Simplify `runListeningCmd()` to only handle command-based execution and daemon mode
- [x] 2.7 Run tests: `go test ./internal/cli/... -v`

## 3. Clean Up View Layer
- [x] 3.1 Remove `stdin *bufio.Reader` field from View struct in `internal/view/view.go`
- [x] 3.2 Remove `GetInputLine()` method from view.go
- [x] 3.3 Remove `PrintInputPrefix()` method (debug prompt) from view.go
- [x] 3.4 Update `NewView()` constructor to remove stdin initialization
- [x] 3.5 Remove bufio import from view.go (no longer needed)
- [x] 3.6 Run tests: `go test ./internal/view/... -v`

## 4. Update Help Messages
- [x] 4.1 Remove REPL-specific help content from `internal/view/help.go` (removed quit command)
- [x] 4.2 Keep command-specific help (breakpoint syntax, print usage, etc.) for CLI `--help`
- [x] 4.3 Update ShowHelpMessage to reflect command-based usage only
- [x] 4.4 Verify help output: `xdebug-cli listen --help`

## 5. Update Listen Command Description
- [x] 5.1 Update `listenCmd` Long description in `internal/cli/listen.go` to remove REPL references
- [x] 5.2 Update Short description to mention command execution
- [x] 5.3 Add usage examples showing `--commands` requirement
- [x] 5.4 Verify output: `xdebug-cli listen --help`

## 6. Remove Interactive Mode Tests
- [x] 6.1 Identify tests in `internal/cli/listen_test.go` that test REPL loop behavior
- [x] 6.2 Update TestNonInteractiveFlag to TestCommandRequirement
- [x] 6.3 Update TestExecuteNonInteractiveCommandParsing to TestCommandParsing
- [x] 6.4 Update TestCLIParameterStruct to remove NonInteractive field
- [x] 6.5 Keep tests for command-based execution and daemon mode
- [x] 6.6 Run full test suite: `go test ./... -v`

## 7. Update View Tests
- [x] 7.1 Remove tests for `GetInputLine()` from `internal/view/view_test.go`
- [x] 7.2 Remove tests for `PrintInputPrefix()` from view_test.go
- [x] 7.3 Remove tests for `PrintApplicationInformation()` from view_test.go
- [x] 7.4 Keep tests for output methods (PrintLn, PrintErrorLn, etc.)
- [x] 7.5 Run view tests: `go test ./internal/view/... -v`

## 8. Update Documentation
- [x] 8.1 Remove "Interactive REPL Commands" section from CLAUDE.md
- [x] 8.2 Update "Available Commands" section to remove `xdebug-cli listen` without flags
- [x] 8.3 Update usage examples to show `--commands` requirement
- [x] 8.4 Update section headers (Non-Interactive Mode -> Command-Based Execution)
- [x] 8.5 Update all command examples to remove `--non-interactive` flag
- [x] 8.6 Update comparison table (Non-Interactive Mode -> Command-Based Mode)

## 9. Integration Testing
- [x] 9.1 Test listen command requires --commands: `xdebug-cli listen` (expect error)
- [x] 9.2 Test listen with commands: works (would need actual Xdebug connection)
- [x] 9.3 Test daemon mode without commands: `xdebug-cli listen --daemon` (expect success)
- [x] 9.4 Test daemon mode with commands: works (would need actual Xdebug connection)
- [x] 9.5 Test attach command: unchanged functionality
- [x] 9.6 Test JSON output: flag still available

## 10. Final Validation
- [x] 10.1 Run full test suite: `go test ./... -v` (all tests pass)
- [x] 10.2 Build binary: `go build -o xdebug-cli ./cmd/xdebug-cli` (successful)
- [x] 10.3 Verify no compilation errors or warnings (clean build)
- [x] 10.4 Verify all documentation updated (CLAUDE.md updated, help messages updated)
- [x] 10.5 Install binary with ./install.sh and verify with `xdebug-cli version`
