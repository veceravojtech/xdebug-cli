# Implementation Tasks

## 1. Create daemon subcommand structure
- [x] Add `statusCmd`, `listCmd`, `killCmd`, `isAliveCmd` to `internal/cli/daemon.go`
- [x] Register subcommands under `daemonCmd` in `init()`
- [x] Add command flags: `--json` for list, `--all` and `--force` for kill
- [x] Add command descriptions and usage examples
- [x] Verify `daemon --help` shows all subcommands

**Validation:** Run `xdebug-cli daemon --help` and verify all subcommands are listed

## 2. Move status functionality to daemon
- [x] Copy `runConnectionStatus()` from `connection.go` to `daemon.go`
- [x] Rename to `runDaemonStatus()` and wire to `statusCmd`
- [x] Update output messages to reference daemon command paths
- [x] Test status display when daemon exists
- [x] Test status display when no daemon exists
- [x] Verify PID, port, socket path, and timestamp display correctly

**Validation:** Start daemon, run `xdebug-cli daemon status`, verify output matches expected format

## 3. Move list functionality to daemon
- [x] Copy `runConnectionList()` from `connection.go` to `daemon.go`
- [x] Rename to `runDaemonList()` and wire to `listCmd`
- [x] Copy `outputSessionListJSON()` helper function
- [x] Update help text references to daemon commands
- [x] Test list output with multiple sessions
- [x] Test list output with no sessions
- [x] Test JSON output format

**Validation:** Start multiple daemons on different ports, run `xdebug-cli daemon list` and verify all sessions shown

## 4. Move kill functionality to daemon
- [x] Copy `runConnectionKill()` from `connection.go` to `daemon.go`
- [x] Rename to `runDaemonKill()` and wire to `killCmd`
- [x] Copy `runConnectionKillAll()` and rename to `runDaemonKillAll()`
- [x] Copy helper functions: `findStaleProcessOnPort()`, `processExists()`
- [x] Update error messages to reference daemon commands
- [x] Test killing single daemon
- [x] Test killing all daemons with confirmation
- [x] Test killing all daemons with --force flag
- [x] Test kill when no daemon exists

**Validation:** Start daemon, run `xdebug-cli daemon kill`, verify daemon terminates

## 5. Move isAlive functionality to daemon
- [x] Copy `runConnectionIsAlive()` from `connection.go` to `daemon.go`
- [x] Rename to `runDaemonIsAlive()` and wire to `isAliveCmd`
- [x] Preserve exit codes: 0 for alive, 1 for not alive
- [x] Preserve output format: "connected" or "not connected"
- [x] Test with running daemon (exit code 0)
- [x] Test without daemon (exit code 1)

**Validation:** Start daemon, run `xdebug-cli daemon isAlive`, verify exit code 0 and output "connected"

## 6. Update tests
- [x] Rename `internal/cli/connection_test.go` to `daemon_lifecycle_test.go`
- [x] Update all test function names to reference daemon subcommands
- [x] Update test command invocations to use `daemon status/list/kill/isAlive`
- [x] Add test coverage for new subcommand structure
- [x] Verify all existing test scenarios pass with new command paths
- [x] Run `go test ./internal/cli/...` and verify all tests pass

**Validation:** Run `go test ./internal/cli/... -v` and verify 100% pass rate

## 7. Remove connection command code
- [x] Remove `connectionCmd` and all subcommands from `connection.go`
- [x] Remove `setActiveSession()`, `clearActiveSession()`, `getActiveSession()` if unused
- [x] Remove `connectionCmd` registration from `root.go`
- [x] Delete `internal/cli/connection.go` file
- [x] Update `daemon.go` to include any needed helper functions
- [x] Verify project builds successfully: `go build ./...`

**Validation:** Run `xdebug-cli connection` and verify "unknown command" error

## 8. Update documentation
- [x] Update `CLAUDE.md` examples to use daemon subcommands
- [x] Replace all `connection` references with `daemon status/list/kill/isAlive`
- [x] Update "Managing Daemon Sessions" section
- [x] Update workflow examples to use new commands
- [x] Update error message examples
- [x] Search for remaining `connection` references: `rg "connection (kill|list|isAlive)" CLAUDE.md`

**Validation:** Search CLAUDE.md for "connection kill|list|isAlive" and verify no matches

## 9. Integration testing
- [x] Test complete workflow: start → status → list → attach → kill
- [x] Test multiple daemon sessions on different ports
- [x] Test kill --all with confirmation prompt
- [x] Test kill --all --force without confirmation
- [x] Test JSON output for list command
- [x] Test isAlive exit codes in shell scripts
- [x] Verify backward compatibility: exit codes, JSON formats preserved

**Validation:** Run full integration test suite and verify all scenarios pass

## 10. Final verification
- [x] Run `go build ./...` and verify clean build
- [x] Run `go test ./...` and verify all tests pass
- [x] Run `xdebug-cli --help` and verify no `connection` command listed
- [x] Run `xdebug-cli daemon --help` and verify all subcommands listed
- [x] Test all daemon subcommands with real Xdebug connections
- [x] Verify no breaking changes to JSON output formats
- [x] Verify exit codes remain consistent with previous implementation

**Validation:** Install built binary and test all documented workflows from CLAUDE.md
