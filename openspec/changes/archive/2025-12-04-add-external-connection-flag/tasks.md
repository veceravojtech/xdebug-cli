## 1. Implementation

- [x] 1.1 Add `--enable-external-connection` flag to `startCmd` in `internal/cli/daemon.go`
- [x] 1.2 Add `EnableExternalConnection` field to `CLIArgs` in `internal/cfg/config.go`
- [x] 1.3 Modify `runDaemonStart()` validation to accept either `--curl` OR `--enable-external-connection`
- [x] 1.4 Skip curl execution when `--enable-external-connection` is set
- [x] 1.5 Update help text to document the new flag

## 2. Testing

- [x] 2.1 Test `daemon start --enable-external-connection` starts and waits for connection
- [x] 2.2 Test that `--curl` and `--enable-external-connection` are mutually exclusive alternatives
- [x] 2.3 Test daemon start without either flag shows appropriate error

## 3. Documentation

- [x] 3.1 Update CLAUDE.md with `--enable-external-connection` usage
- [x] 3.2 Add examples for external trigger workflows
