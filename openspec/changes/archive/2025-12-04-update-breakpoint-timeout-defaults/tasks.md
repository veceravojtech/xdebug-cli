## 1. Implementation

- [x] 1.1 Change default `--breakpoint-timeout` from 10 to 30 in `internal/cli/daemon.go`
- [x] 1.2 Add `--wait-forever` boolean flag as alias for `--breakpoint-timeout 0`
- [x] 1.3 Update flag help text to explain timeout behavior
- [x] 1.4 Add note about cold start scenarios in CLAUDE.md

## 2. Testing

- [x] 2.1 Verify new default timeout is 30 seconds
- [x] 2.2 Test `--wait-forever` sets timeout to 0 (disabled)
- [x] 2.3 Test `--breakpoint-timeout` still overrides both defaults
- [x] 2.4 Run existing test suite to verify no regressions

## 3. Documentation

- [x] 3.1 Update CLAUDE.md usage examples if needed
- [x] 3.2 Ensure help text explains when to use each option

## 4. Verification

- [ ] 4.1 Manual test with slow PHP bootstrap (cold cache)
- [ ] 4.2 Verify timeout still works when explicitly set
