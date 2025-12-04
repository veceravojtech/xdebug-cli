# Add DBGp Message Read Timeout

## Why

The `Connection.ReadMessage()` method can hang indefinitely if:
1. Xdebug becomes unresponsive (e.g., PHP process stuck, infinite loop)
2. Network connection stalls without TCP timeout
3. Incomplete message frames (no null terminator arrives)

Both `bufio.ReadBytes(0)` and `io.ReadFull()` block forever with no deadline set. This makes the CLI appear frozen with no way to recover except killing the process.

## What Changes

- Add `ReadMessageWithTimeout(timeout time.Duration)` method to Connection
- Set read deadline on underlying connection before reading
- Clear deadline after successful read
- Provide sensible default timeout (30 seconds) with option to customize
- Existing `ReadMessage()` remains for backward compatibility

## Impact

- Affected specs: `dbgp` (Connection Message Framing requirement)
- Affected code: `internal/dbgp/connection.go`
- No breaking changes - new method added
- Prevents indefinite hangs when Xdebug is unresponsive
