# Add --enable-external-connection Flag

## Why

The current `daemon start` command requires `--curl` to trigger the Xdebug connection. However, some debugging workflows involve external triggers:
- Triggering PHP from a browser with XDEBUG_TRIGGER cookie
- IDE-initiated debugging (PhpStorm, VS Code)
- Manual curl from another terminal
- CLI script execution with xdebug.start_with_request=trigger

These workflows need the daemon to listen and wait without initiating any HTTP request.

## What Changes

- Add `--enable-external-connection` flag to `daemon start`
- When provided, bypasses the `--curl` requirement
- Daemon starts, listens on port, and waits indefinitely for an external Xdebug connection
- All other daemon functionality remains unchanged (IPC, attach, kill, etc.)

## Impact

- Affected specs: `cli` (Daemon Start Command requirement)
- Affected code: `internal/cli/daemon.go`
- No breaking changes - `--curl` remains the default required flag
- No changes to daemon forking, IPC, registry, or attach command
