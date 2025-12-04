## 1. Implementation

- [x] 1.1 Add `validateProcess(pid int)` helper in `internal/daemon/registry.go`
- [x] 1.2 Check process exists via `/proc/<pid>` existence
- [x] 1.3 Verify process is xdebug-cli via `/proc/<pid>/comm` content
- [x] 1.4 Add `CleanupStaleEntries()` method to SessionRegistry
- [x] 1.5 Call `CleanupStaleEntries()` at start of `daemon start` command
- [x] 1.6 Clean up orphaned socket files when removing stale entries

## 2. Testing

- [x] 2.1 Add test: stale entry removed when process doesn't exist
- [x] 2.2 Add test: stale entry removed when PID recycled to different process
- [x] 2.3 Add test: valid entries preserved during cleanup
- [x] 2.4 Add test: orphaned files cleaned up with stale entries
- [x] 2.5 Run existing test suite to verify no regressions

## 3. Verification

- [ ] 3.1 Manual test: kill -9 daemon, then start new daemon works
- [ ] 3.2 Manual test: restart after OOM killer works without workaround
