## 1. Implementation

- [x] 1.1 Add warning message to stderr when timeout occurs in `internal/cli/daemon.go`
- [x] 1.2 Include timeout duration and pending breakpoints in message
- [x] 1.3 Suggest `--breakpoint-timeout` or `--wait-forever` flags
- [x] 1.4 Change exit code from 1 to 124 (Unix timeout convention)
- [x] 1.5 Write timeout event to `/tmp/xdebug-cli-daemon-<port>.log`

## 2. Testing

- [x] 2.1 Add test: timeout produces warning on stderr
- [x] 2.2 Add test: exit code is 124 on timeout
- [x] 2.3 Add test: log file contains timeout details
- [x] 2.4 Run existing test suite to verify no regressions

## 3. Documentation

- [x] 3.1 Document exit code 124 in CLAUDE.md
- [x] 3.2 Add troubleshooting section for timeout scenarios

## 4. Verification

- [ ] 4.1 Manual test: timeout message appears on stderr
- [ ] 4.2 Manual test: message provides actionable guidance
