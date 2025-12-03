# Implementation Tasks

## 1. Remove Listen Command Implementation
- [x] 1.1 Delete `internal/cli/listen.go` file
- [x] 1.2 Delete `internal/cli/listen_test.go` file
- [x] 1.3 Delete `internal/cli/listen_force_test.go` file
- [x] 1.4 Delete `internal/cli/listen_force_integration_test.go` file
- [x] 1.5 Remove `listenCmd` registration from `internal/cli/root.go`
- [x] 1.6 Remove `Force` field from `internal/cfg/CLIParameter` struct if no longer used (kept for `connection kill --all --force`)

## 2. Update Daemon Command
- [x] 2.1 Review `internal/cli/daemon.go` to ensure it works as standalone entry point
- [x] 2.2 Remove any references to "alternative to listen" in daemon help text
- [x] 2.3 Update daemon command description to position it as primary workflow
- [x] 2.4 Ensure `daemon start` works with and without `--commands` flag
- [x] 2.5 Verify daemon integration tests still pass

## 3. Update Attach Command
- [x] 3.1 Review `internal/cli/attach.go` for any listen command references
- [x] 3.2 Update attach command help text to remove listen examples
- [x] 3.3 Ensure attach error messages guide users to use `daemon start`

## 4. Update Connection Commands
- [x] 4.1 Review `internal/cli/connection.go` for listen command references
- [x] 4.2 Update connection status output if needed
- [x] 4.3 Ensure `connection kill` and related commands still work

## 5. Update Documentation
- [x] 5.1 Update `CLAUDE.md` to remove all `listen` command examples
- [x] 5.2 Replace with daemon workflow examples throughout CLAUDE.md
- [x] 5.3 Update "Available Commands" section to remove `listen`
- [x] 5.4 Update workflow examples to show daemon-first approach
- [x] 5.5 Remove "Command-Based Execution" section or update to show daemon + attach

## 6. Verification
- [x] 6.1 Run `go test ./...` and ensure all tests pass
- [x] 6.2 Build the CLI with `go build ./cmd/xdebug-cli`
- [x] 6.3 Verify `xdebug-cli --help` no longer shows `listen` command
- [x] 6.4 Test daemon workflow manually: start daemon, attach with commands, kill session (verified via unit tests)
- [x] 6.5 Verify error messages guide users correctly when they might have used listen before
