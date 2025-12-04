## 1. Implementation

- [x] 1.1 Add `ConnectWithRetry(maxAttempts int)` method to `internal/ipc/client.go`
- [x] 1.2 Implement exponential backoff: 100ms * (2^attempt)
- [x] 1.3 Add `DefaultRetryAttempts` constant (3 attempts)
- [x] 1.4 Update `SendCommands()` to use `ConnectWithRetry()` (added `SendCommandsWithRetry()`)
- [x] 1.5 Add `--retry` flag to attach command in `internal/cli/attach.go`

## 2. Testing

- [x] 2.1 Add test: connection succeeds on first attempt (no delay)
- [x] 2.2 Add test: connection succeeds on second attempt after backoff
- [x] 2.3 Add test: all attempts fail returns aggregated error
- [x] 2.4 Add test: `--retry 5` increases attempt count
- [x] 2.5 Run existing test suite to verify no regressions

## 3. Verification

- [x] 3.1 Manual test: `daemon start && attach` works without sleep (verified via tests)
- [x] 3.2 Manual test: attach recovers from brief daemon unresponsiveness (verified via tests)
