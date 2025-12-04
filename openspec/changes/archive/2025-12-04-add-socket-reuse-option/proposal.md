# Add SO_REUSEADDR Socket Option

## Why

After daemon shutdown (normal or crash), the TCP port (default 9003) often remains in `TIME_WAIT` state for 60+ seconds on Linux. During this period, `daemon start` fails with "bind: address already in use" even though no process is using the port.

This forces users to either wait or use workarounds like `pkill -9 -f xdebug-cli && sleep 3`.

## What Changes

- Modify `dbgp.Server.Listen()` to use `net.ListenConfig` with `Control` function
- Set `SO_REUSEADDR` socket option before binding
- Allows immediate port reuse after daemon termination

## Impact

- Affected specs: `dbgp` (TCP Server requirement)
- Affected code: `internal/dbgp/server.go:25-34`
- No breaking changes - same Listen() signature
- Improves reliability for rapid daemon restart cycles
