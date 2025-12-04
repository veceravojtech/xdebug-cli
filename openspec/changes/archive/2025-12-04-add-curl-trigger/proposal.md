# Add --curl Flag for Integrated HTTP Trigger

## Problem

The current workflow requires users to run `curl` in a separate terminal to trigger Xdebug connections:

```bash
# Terminal 1
xdebug-cli daemon start

# Terminal 2 (race condition!)
curl http://localhost/app.php -b "XDEBUG_TRIGGER=1"

# Terminal 1
xdebug-cli attach --commands "..."
```

This creates race conditions and is difficult for Claude/AI assistants to orchestrate reliably.

## Solution

Make `--curl` a required flag on `daemon start` that handles the HTTP request internally:

```bash
xdebug-cli daemon start --curl "http://localhost/app.php"
xdebug-cli daemon start --curl "http://localhost/api -X POST -d 'data'" --commands "break :42"
```

## Behavior

1. `--curl` is **required** - command fails with helpful error if missing
2. Auto-appends `-b "XDEBUG_TRIGGER=1"` to curl command
3. Shells out to actual `curl` binary (full compatibility with all curl features)
4. Fire-and-forget: starts curl after server is listening, doesn't wait for HTTP response
5. **Fail-fast**: if curl exits with error, daemon terminates with error
6. Existing auto-kill behavior (same port) remains unchanged

## Example Workflow

```bash
# Single command - no race condition!
xdebug-cli daemon start --curl "http://localhost/app.php" --commands "break :42"

# Wait for Xdebug to connect and hit breakpoint...

xdebug-cli attach --commands "context local"
xdebug-cli attach --commands "run"
xdebug-cli daemon kill
```

## Error Handling

When `--curl` is not provided:
```
Error: --curl flag is required

Usage:
  xdebug-cli daemon start --curl "<curl-args>"

Examples:
  xdebug-cli daemon start --curl "http://localhost/app.php"
  xdebug-cli daemon start --curl "http://localhost/api -X POST -d 'data'"
  xdebug-cli daemon start --curl "http://localhost/app.php" --commands "break :42"

The --curl flag specifies the HTTP request to trigger Xdebug.
XDEBUG_TRIGGER cookie is added automatically.
```

When curl fails:
```
Error: curl failed with exit code 7: Could not connect to host
Daemon terminated.
```

## Scope

- Modifies: `internal/cli/daemon.go` (add flag, execute curl)
- Updates: CLAUDE.md documentation
- No changes to: daemon forking, IPC, attach command
