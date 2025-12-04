# Add Attach Retry Logic

## Why

The `attach` command currently makes a single connection attempt to the daemon socket. This causes failures in scenarios where:
1. User runs `attach` immediately after `daemon start` (daemon still initializing)
2. Transient filesystem/network issues cause momentary connection failure
3. Daemon is briefly busy and doesn't accept connection immediately

Users see "failed to connect to daemon socket" and must manually retry.

## What Changes

- Add retry logic with exponential backoff to IPC client connection
- Default: 3 attempts with 100ms, 200ms, 400ms delays
- Add `--retry` flag to customize number of attempts
- Improve error message to indicate all attempts failed

## Impact

- Affected specs: `persistent-sessions` (IPC Communication requirement)
- Affected code: `internal/ipc/client.go`, `internal/cli/attach.go`
- No breaking changes - retry is transparent improvement
- Eliminates race condition between daemon start and attach
