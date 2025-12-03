# DBGp Protocol Commands Reference

Complete reference of all Xdebug DBGp protocol commands extracted from [xdebug.org/docs/dbgp](https://xdebug.org/docs/dbgp).

---

## Core Commands (Required Support)

### Status Command
**Command:** `status -i transaction_id`

**Purpose:** Query debugger engine execution state

**Description:** Determines if code execution can continue. Returns status values: starting, stopping, stopped, running, or break. Reason attributes include ok, error, aborted, or exception.

---

### Feature Commands

#### feature_get
**Command:** `feature_get -i transaction_id -n feature_name`

**Purpose:** Query supported features and capabilities

**Returns:** Feature support status and configuration details

#### feature_set
**Command:** `feature_set -i transaction_id -n feature_name -v value`

**Purpose:** Configure debugger engine capabilities

**Returns:** Success/failure confirmation

---

### Continuation Commands

These commands resume script execution:

| Command | Syntax | Purpose |
|---------|--------|---------|
| **run** | `run -i transaction_id` | Execute until breakpoint or script end |
| **step_into** | `step_into -i transaction_id` | Step to next statement, entering functions |
| **step_over** | `step_over -i transaction_id` | Step to next statement, skipping function calls |
| **step_out** | `step_out -i transaction_id` | Exit current scope, stop after return |
| **stop** | `stop -i transaction_id` | Terminate execution immediately |
| **detach** | `detach -i transaction_id` | Cease debugging without stopping script |

---

### Breakpoint Commands

#### breakpoint_set
**Command:** `breakpoint_set -i ID -t TYPE [-f FILENAME] [-n LINENO] [-m FUNCTION] [-x EXCEPTION] [-h HIT_VALUE] [-o HIT_CONDITION] [-r 0|1] -- base64(expression)`

**Purpose:** Create new breakpoint with specified conditions

**Types:** line, call, return, exception, conditional, watch

**Parameters:**
- `-i` : Transaction ID (required)
- `-t` : Breakpoint type (required)
- `-f` : Filename (optional)
- `-n` : Line number (optional)
- `-m` : Function name (optional)
- `-x` : Exception name (optional)
- `-h` : Hit value (optional)
- `-o` : Hit condition (optional)
- `-r` : Temporary breakpoint flag (optional)
- `--` : Expression data separator (base64-encoded)

#### breakpoint_get
**Command:** `breakpoint_get -i ID -d BREAKPOINT_ID`

**Purpose:** Retrieve breakpoint details

#### breakpoint_update
**Command:** `breakpoint_update -i ID -d BREAKPOINT_ID [-s STATE] [-n LINENO] [-h HIT_VALUE] [-o HIT_CONDITION]`

**Purpose:** Modify existing breakpoint attributes

#### breakpoint_remove
**Command:** `breakpoint_remove -i ID -d BREAKPOINT_ID`

**Purpose:** Delete specified breakpoint

#### breakpoint_list
**Command:** `breakpoint_list -i transaction_id`

**Purpose:** Retrieve all active breakpoints

---

### Stack Commands

#### stack_depth
**Command:** `stack_depth -i transaction_id`

**Purpose:** Obtain maximum stack depth available

#### stack_get
**Command:** `stack_get [-d NUM] -i transaction_id`

**Purpose:** Retrieve stack frame information at specified depth

**Parameters:**
- `-d` : Stack depth level (optional, defaults to current)
- `-i` : Transaction ID (required)

---

### Context Commands

#### context_names
**Command:** `context_names [-d NUM] -i transaction_id`

**Purpose:** List available variable contexts (Local, Global, Class)

**Parameters:**
- `-d` : Stack depth (optional)
- `-i` : Transaction ID (required)

#### context_get
**Command:** `context_get [-d NUM] [-c NUM] -i transaction_id`

**Purpose:** Retrieve all properties within specified context

**Parameters:**
- `-d` : Stack depth (optional)
- `-c` : Context ID (optional)
- `-i` : Transaction ID (required)

---

### Property/Variable Commands

#### property_get
**Command:** `property_get -n property_name [-d NUM] [-c NUM] [-m SIZE] [-p PAGE] [-k KEY] -i transaction_id`

**Purpose:** Retrieve variable/property value with optional depth limits

**Parameters:**
- `-n` : Property name (required)
- `-d` : Stack depth (optional)
- `-c` : Context ID (optional)
- `-m` : Maximum data size (optional)
- `-p` : Page number for large data (optional)
- `-k` : Key for hash/array access (optional)
- `-i` : Transaction ID (required)

#### property_set
**Command:** `property_set -n property_name [-d NUM] [-c NUM] [-t TYPE] -l length -- DATA -i transaction_id`

**Purpose:** Modify variable/property value

**Parameters:**
- `-n` : Property name (required)
- `-d` : Stack depth (optional)
- `-c` : Context ID (optional)
- `-t` : Data type (optional)
- `-l` : Data length (required)
- `--` : Data separator
- `-i` : Transaction ID (required)

#### property_value
**Command:** `property_value -n property_name [-d NUM] [-c NUM] [-p PAGE] [-k KEY] -i transaction_id`

**Purpose:** Obtain complete property data when truncated

---

### Source Command

**Command:** `source -i transaction_id [-b BEGIN_LINE] [-e END_LINE] [-f FILE_URI]`

**Purpose:** Retrieve source code contents for specified file or current context, optionally limited to line range

**Parameters:**
- `-i` : Transaction ID (required)
- `-b` : Beginning line number (optional)
- `-e` : Ending line number (optional)
- `-f` : File URI (optional, defaults to current file)

---

### Stream Redirection Commands

#### stdout
**Command:** `stdout -i transaction_id -c [0|1|2]`

**Purpose:** Control standard output capture

**Modes:**
- `0` : Disable stdout capture
- `1` : Copy stdout to IDE
- `2` : Redirect stdout to IDE

#### stderr
**Command:** `stderr -i transaction_id -c [0|1|2]`

**Purpose:** Control standard error capture

**Modes:**
- `0` : Disable stderr capture
- `1` : Copy stderr to IDE
- `2` : Redirect stderr to IDE

---

## Extended Commands (Optional Support)

### stdin Command

**Command:** `stdin -i transaction_id -c [0|1]` (for mode setting)
**Command:** `stdin -i transaction_id -- base64(data)` (for data transmission)

**Purpose:** Enable input redirection or transmit input data to debugger engine

**Modes:**
- `0` : Disable stdin redirection
- `1` : Enable stdin redirection

---

### break Command

**Command:** `break -i transaction_id`

**Purpose:** Interrupt script execution while debugger engine operates in run state

**Description:** Allows breaking into running code without a predefined breakpoint.

---

### eval Command

**Command:** `eval -i transaction_id [-d DEPTH] [-p PAGE] -- base64(code)`

**Purpose:** Execute arbitrary code within current execution context

**Parameters:**
- `-i` : Transaction ID (required)
- `-d` : Stack depth for evaluation context (optional)
- `-p` : Page number for paginated results (optional)
- `--` : Code separator (base64-encoded code follows)

**Variants:**
- **expr** : Expression evaluation variant
- **exec** : Code execution variant

---

### Spawnpoint Commands

Similar structure to breakpoints but track process spawning points:

#### spawnpoint_set
**Purpose:** Create new spawnpoint

#### spawnpoint_get
**Purpose:** Retrieve spawnpoint details

#### spawnpoint_update
**Purpose:** Modify existing spawnpoint attributes

#### spawnpoint_remove
**Purpose:** Delete specified spawnpoint

#### spawnpoint_list
**Purpose:** Retrieve all active spawnpoints

---

### interact Command

**Command:** `interact -i transaction_id [-m MODE] -- base64(code)`

**Purpose:** Buffer and compile code incrementally, mimicking interactive console input

**Parameters:**
- `-i` : Transaction ID (required)
- `-m` : Mode (optional)
- `--` : Code separator (base64-encoded code follows)

---

### Notifications

**Purpose:** Receive asynchronous messages from debugger engine

**Activation:** Requires `feature_set -n notify_ok -v 1`

**Standard notification types:**
- `stdin` : Standard input notifications
- `breakpoint_resolved` : Breakpoint resolution notifications
- `error` : Error notifications

---

## Common Parameters

Parameters used across multiple commands:

| Parameter | Description |
|-----------|-------------|
| `-i` | Transaction identifier (required on all commands) |
| `-d` | Stack depth or debugger ID |
| `-c` | Context identifier |
| `-n` | Property/feature name |
| `-v` | Feature value to set |
| `-f` | File URI |
| `-m` | Function name or mode |
| `-t` | Type specification |
| `-l` | Data length |
| `-p` | Page number for pagination |
| `-k` | Key for hash/array access |
| `-h` | Hit value for breakpoints |
| `-o` | Hit condition for breakpoints |
| `-s` | State for breakpoint updates |
| `-r` | Temporary breakpoint flag (0 or 1) |
| `-b` | Beginning line number |
| `-e` | Ending line number |
| `-x` | Exception name |
| `--` | Data separator (following data is base64-encoded) |

---

## Command Categories Summary

### Essential Debugging Commands
- `status` - Check execution state
- `run`, `step_into`, `step_over`, `step_out` - Control execution flow
- `breakpoint_*` - Manage breakpoints
- `context_get`, `property_get` - Inspect variables

### Advanced Inspection
- `stack_get`, `stack_depth` - Call stack analysis
- `source` - View source code
- `eval` - Execute code dynamically

### Session Management
- `stop`, `detach` - End debugging session
- `feature_get`, `feature_set` - Configure debugger capabilities
- `stdout`, `stderr`, `stdin` - Stream redirection

### Specialized Features
- `spawnpoint_*` - Process spawning tracking
- `interact` - Interactive console mode
- `break` - Interrupt running execution
- Notifications - Asynchronous event handling

---

## Implementation Status

### CLI Commands to DBGp Protocol Mapping

This section documents which CLI commands have been implemented and their corresponding DBGp protocol commands.

| CLI Command | DBGp Protocol Command | Status | Usage | Notes |
|-------------|----------------------|--------|-------|-------|
| `status` | `status -i {id}` | ✓ Implemented | Query debugger execution state | Returns: starting, stopping, stopped, running, break |
| `run` / `r` | `run -i {id}` | ✓ Implemented | Continue execution until breakpoint or script end | Standard execution control |
| `step` / `s` | `step_into -i {id}` | ✓ Implemented | Step into next statement (enter functions) | Fundamental debugging step |
| `next` / `n` | `step_over -i {id}` | ✓ Implemented | Step over next statement (skip functions) | Fundamental debugging step |
| `out` / `o` | `step_out -i {id}` | ✓ Implemented | Step out of current function scope | Return to caller |
| `detach` | `detach -i {id}` | ✓ Implemented | Cease debugging without stopping script | Clean session termination |
| `eval` | `eval -i {id} -- base64(expression)` | ✓ Implemented | Execute arbitrary code in current context | Stack depth support with `-d` flag |
| `set` | `property_set -i {id} -n {name} -t {type} -l {len} -- base64(value)` | ✓ Implemented | Modify variable/property value | Type and length parameters required |
| `break` / `b` | `breakpoint_set -i {id} -t {type} ...` | ✓ Implemented | Set breakpoint (line, function, exception) | Supports conditional breakpoints, multiple targets |
| `disable` | `breakpoint_update -i {id} -d {bid} -s disabled` | ✓ Implemented | Disable existing breakpoint | Preserves breakpoint for later re-enable |
| `enable` | `breakpoint_update -i {id} -d {bid} -s enabled` | ✓ Implemented | Re-enable disabled breakpoint | Activates previously disabled breakpoint |
| `delete` | `breakpoint_remove -i {id} -d {bid}` | ✓ Implemented | Remove breakpoint permanently | Cannot be restored |
| `info` | `breakpoint_list -i {id}` | ✓ Implemented | List all active breakpoints | Shows current breakpoint state |
| `stack` | `stack_get -i {id} [-d {depth}]` | ✓ Implemented | Retrieve stack frame information | Optional depth parameter for specific frames |
| `list` / `l` | `source -i {id} [-f {file}] [-b {begin}] [-e {end}]` | ✓ Implemented | Display source code with optional range | File, begin line, end line parameters |
| `print` / `p` | `property_get -n {name} -i {id}` | ✓ Implemented | Print variable value | Supports nested property access |
| `context` / `c` | `context_get -i {id} [-c {type}]` | ✓ Implemented | Show variables (local/global/constant) | Optional context type filter |
| `finish` / `f` | `stop -i {id}` | ✓ Implemented | Stop debugging, terminate execution | Ends debug session |
| `help` / `h` / `?` | (Internal) | ✓ Implemented | Show command help | CLI-level command |
| `version` | (Internal) | ✓ Implemented | Show version and build timestamp | CLI-level command |

### Known Limitations and Notes

1. **Base64 Encoding**: Commands that use `--` separator require base64-encoded data (e.g., `eval`, `property_set`). The CLI handles encoding/decoding automatically.

2. **Transaction IDs**: All DBGp commands require `-i {id}` parameter. The CLI session manager automatically assigns and tracks transaction IDs.

3. **Conditional Breakpoints**: When using `break :line if $expression`, the expression is automatically base64-encoded and passed to `breakpoint_set` with the conditional type.

4. **Multiple Breakpoints**: Commands like `break :42 :100 :150` internally issue multiple `breakpoint_set` commands, one for each target.

5. **Stack Depth Context**: Commands like `context`, `property_get`, and `eval` support optional stack depth (`-d` flag) to inspect different stack frames. Defaults to current frame if not specified.

6. **Property Paths**: Variable names can include object/array notation (e.g., `$obj->property`, `$array['key']`), which the CLI properly encodes.

### Commands Not Yet Implemented

The following DBGp commands are documented but not yet integrated into the CLI command interface:

| DBGp Command | Reason | Alternative |
|--------------|--------|-------------|
| `feature_get` / `feature_set` | Debugger capability negotiation handled internally | Session manager handles feature negotiation |
| `stdout` / `stderr` / `stdin` | Stream redirection requires complex buffering | Output captured through existing commands |
| `break` (interrupt) | Run-state interruption not essential for current workflow | Use breakpoints instead |
| `interact` | Interactive console mode requires shell-like parsing | Use `eval` for dynamic code execution |
| `spawnpoint_*` | Process spawning tracking rarely used in PHP debugging | Not prioritized |
| `context_names` | Context enumeration handled internally | `context` command infers available contexts |
| `stack_depth` | Depth tracking handled internally | Stack operations handle auto-discovery |

### Error Handling

All CLI commands return:
- **Exit code 0**: Command executed successfully
- **Exit code 1**: Command failed (invalid syntax, session ended, DBGp error)

JSON output mode (`--json`) provides structured error information:
```json
{
  "command": "print",
  "success": false,
  "error": "variable not found",
  "details": "undefined variable: $unknownVar"
}
```

---

*Generated from Xdebug DBGp Protocol Documentation*
*Implementation Status Last Updated: 2025-12-03*
