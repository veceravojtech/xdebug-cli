## 1. Implementation

- [x] 1.1 Add `DefaultMessageTimeout` constant (30 seconds) in `internal/dbgp/connection.go`
- [x] 1.2 Add `ReadMessageWithTimeout(timeout time.Duration)` method
- [x] 1.3 Set read deadline via `c.conn.SetReadDeadline()` before reading
- [x] 1.4 Clear deadline with `SetReadDeadline(time.Time{})` after successful read
- [x] 1.5 Update `GetResponse()` to use `ReadMessageWithTimeout`

## 2. Testing

- [x] 2.1 Add test: timeout fires when no data received
- [x] 2.2 Add test: successful read within timeout works normally
- [x] 2.3 Add test: deadline is cleared after successful read
- [x] 2.4 Run existing test suite to verify no regressions

## 3. Integration

- [x] 3.1 Update daemon command executor to use timeout-aware reads
- [x] 3.2 Ensure timeout errors are reported clearly to user

## 4. Verification

- [ ] 4.1 Manual test: verify timeout fires when PHP is stuck
- [ ] 4.2 Manual test: verify normal debugging unaffected
