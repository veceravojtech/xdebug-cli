## 1. Implementation

- [x] 1.1 Import `syscall` package in `internal/dbgp/server.go`
- [x] 1.2 Modify `Listen()` to use `net.ListenConfig` with `Control` callback
- [x] 1.3 Set `SO_REUSEADDR` via `syscall.SetsockoptInt()` in Control callback
- [x] 1.4 Update error handling for the new Listen pattern

## 2. Testing

- [x] 2.1 Add unit test: server can rebind immediately after close
- [ ] 2.2 Manual test: `daemon start` works after `daemon kill` without delay
- [x] 2.3 Run existing test suite to verify no regressions

## 3. Verification

- [x] 3.1 Verify port is reusable immediately after daemon termination
- [x] 3.2 Verify `TIME_WAIT` state no longer blocks restart
