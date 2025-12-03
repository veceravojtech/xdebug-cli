# Tasks: Default Listen All Interfaces

## Implementation Tasks

- [x] **Update default host value in root.go**
   - Change default from `"127.0.0.1"` to `"0.0.0.0"` in flag definition
   - File: `internal/cli/root.go:37`
   - Verification: `go build ./...` succeeds

- [x] **Update tests for new default**
   - Verify any tests that assert on default host value
   - File: `internal/cfg/config_test.go`, `internal/cli/listen_test.go`
   - Verification: `go test ./...` passes

- [x] **Update CLI spec**
   - Modify scenario for host flag to reflect new default
   - File: `openspec/specs/cli/spec.md`
   - Verification: `openspec validate default-listen-all-interfaces --strict`

- [x] **Update CLAUDE.md documentation**
   - Update PHP configuration example comment if needed
   - Update any references to default host
   - File: `CLAUDE.md`
   - Verification: Documentation reflects `0.0.0.0` default

- [x] **Update README.md**
   - Update any examples or descriptions referencing default host
   - File: `README.md`
   - Verification: Documentation is consistent

## Verification
- [x] `go build ./...` - Build succeeds
- [x] `go test ./...` - All tests pass
- [x] `xdebug-cli listen --help` - Shows `0.0.0.0` as default for `-l`
- [x] `./install.sh` - Binary installed to `~/.local/bin/`
- [x] `xdebug-cli version` - Returns version 1.0.0
