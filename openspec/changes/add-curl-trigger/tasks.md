# Tasks

## Implementation

1. [x] Add `--curl` flag to `startCmd` in `internal/cli/daemon.go`
   - String flag for curl arguments
   - Mark as required (validate in Run function)

2. [x] Add curl validation in `runDaemonStart()`
   - Check if `--curl` flag is empty
   - Display helpful error message with examples if missing
   - Exit with code 1

3. [x] Implement `executeCurl()` function in `internal/cli/daemon.go`
   - Accept curl args string
   - Append `-b "XDEBUG_TRIGGER=1"` to args
   - Shell out to `curl` binary using `exec.Command`
   - Return error channel for async monitoring

4. [x] Integrate curl execution in `runDaemonProcess()`
   - After `server.Listen()` succeeds, before `server.Accept()`
   - Start curl in goroutine
   - Monitor for curl failure, kill daemon if curl exits with error

5. [x] Add unit tests for curl flag validation
   - Test missing --curl shows proper error
   - Test curl args are passed correctly

6. [x] Add integration test for curl trigger flow
   - Test daemon start with --curl triggers request
   - Test curl failure terminates daemon

7. [x] Update CLAUDE.md documentation
   - Update command examples to use --curl
   - Remove manual curl instructions from workflows

## Verification

- [x] `go build ./...` succeeds
- [x] `go test ./...` passes
- [x] Manual test: `xdebug-cli daemon start --curl "http://localhost/test.php"` works
- [x] Manual test: `xdebug-cli daemon start` (without --curl) shows helpful error
