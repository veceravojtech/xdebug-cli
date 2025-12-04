## 1. Implementation

- [x] 1.1 Add `MaxMessageSize` constant (100MB) in `internal/dbgp/connection.go`
- [x] 1.2 Add size field format validation (digits only) before `strconv.Atoi()`
- [x] 1.3 Add bounds check after parsing: reject negative or > MaxMessageSize
- [x] 1.4 Improve error message to show first 50 bytes of invalid size field

## 2. Testing

- [x] 2.1 Add test: reject size field with non-digit characters
- [x] 2.2 Add test: reject negative size values
- [x] 2.3 Add test: reject size values exceeding MaxMessageSize
- [x] 2.4 Add test: valid messages still parse correctly
- [x] 2.5 Run existing test suite to verify no regressions

## 3. Verification

- [ ] 3.1 Manual test with large Xdebug responses (deep stack traces)
- [ ] 3.2 Verify error messages are actionable
