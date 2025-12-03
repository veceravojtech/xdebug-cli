# Design: Implement DBGp Protocol Commands

## Architecture Overview
This change extends the existing three-layer architecture without modifying its structure:

```
┌─────────────────────────────────────────────────────┐
│  CLI Layer (internal/cli/)                          │
│  - Command parsing and validation                   │
│  - User-facing command handlers                     │
│  - Output formatting                                │
└─────────────────────────────────────────────────────┘
                       ↓
┌─────────────────────────────────────────────────────┐
│  DBGp Client Layer (internal/dbgp/client.go)        │
│  - Protocol command methods                         │
│  - Response parsing                                 │
│  - Session state updates                            │
└─────────────────────────────────────────────────────┘
                       ↓
┌─────────────────────────────────────────────────────┐
│  Protocol Layer (internal/dbgp/protocol.go)         │
│  - Raw DBGp message framing                         │
│  - XML parsing                                      │
└─────────────────────────────────────────────────────┘
```

## Command Design Patterns

### Pattern 1: Simple Execution Control Commands
**Example**: `status`, `detach`

**User Input**: `xdebug-cli attach --commands "status"`

**Flow**:
1. CLI dispatcher routes to `handleStatus()`
2. Handler calls `client.Status()`
3. Client sends `status -i {id}`
4. Response parsed and displayed via view layer

**Implementation**:
- Add `Status()` and `Detach()` methods to `dbgp.Client`
- Add `handleStatus()` and `handleDetach()` to `internal/cli/listen.go`
- Add cases to dispatcher switch statements
- Add JSON output formatting

### Pattern 2: Commands with Arguments
**Example**: `eval <expression>`

**User Input**: `xdebug-cli attach --commands "eval \$x + 10"`

**Flow**:
1. CLI parses command: `["eval", "$x", "+", "10"]`
2. Rejoin args after command name: `"$x + 10"`
3. Base64-encode expression per DBGp protocol
4. Client sends `eval -i {id} -- base64(expression)`
5. Response contains evaluation result
6. Display formatted output

**Implementation**:
- Add `Eval(expression string)` to `dbgp.Client`
- Implement base64 encoding for expression data
- Parse result from `<property>` XML element
- Format output similar to `print` command

### Pattern 3: Breakpoint Management
**Example**: `delete 123`, `disable 456`, `enable 789`

**User Input**: `xdebug-cli attach --commands "delete 123"`

**Flow**:
1. CLI parses breakpoint ID from args
2. Validate ID is numeric
3. Call appropriate client method
4. Confirm success/failure to user

**Implementation**:
- `RemoveBreakpoint(id)` already exists in client.go
- Add `UpdateBreakpoint(id, enabled bool)` to client.go
- Add `handleDelete()`, `handleDisable()`, `handleEnable()` handlers
- Map to DBGp commands:
  - `delete` → `breakpoint_remove -i {tid} -d {bid}`
  - `disable` → `breakpoint_update -i {tid} -d {bid} -s disabled`
  - `enable` → `breakpoint_update -i {tid} -d {bid} -s enabled`

### Pattern 4: Variable Modification
**Example**: `set $myVar = "new value"`

**User Input**: `xdebug-cli attach --commands "set \$count = 42"`

**Flow**:
1. Parse variable name and new value
2. Determine data type (string, int, bool, etc.)
3. Encode value as base64 per DBGp protocol
4. Send `property_set` command
5. Confirm success

**Parsing Strategy**:
```
Input: "set $myVar = 42"
Split: ["set", "$myVar", "=", "42"]
Extract: variable="$myVar", value="42"
Detect type: integer
```

**Implementation**:
- Add `SetProperty(name, value, dataType string)` to client
- Implement type detection heuristic (int, float, string, bool)
- Handle base64 encoding of value
- Send `property_set -i {id} -n {name} -t {type} -l {length} -- base64(value)`

### Pattern 5: Source Code Display
**Example**: `source`, `source file.php`, `source file.php:10-20`

**User Input**: `xdebug-cli attach --commands "source app.php:100-120"`

**Flow**:
1. Parse optional file and line range
2. Convert file path to file:// URI
3. Send `source` command with parameters
4. Receive base64-encoded source
5. Decode and display with line numbers

**Syntax Options**:
- `source` → current file, all lines
- `source file.php` → specific file, all lines
- `source file.php:100-120` → specific file, line range
- `source :100-120` → current file, line range

**Implementation**:
- Add `GetSource(fileURI, startLine, endLine int)` to client
- Parse file path and line range from args
- Normalize file paths to file:// URIs
- Display similar to `list` command output

### Pattern 6: Stack Command
**Example**: `stack`

**User Input**: `xdebug-cli attach --commands "stack"`

**Flow**:
1. Call existing `client.GetStackTrace()`
2. Format output consistently with `info stack`
3. Make standalone command for convenience

**Implementation**:
- Reuse existing `GetStackTrace()` from client
- Extract formatting logic from `handleInfo()`
- Create `handleStack()` wrapper
- Add to dispatcher

## Command Alias Strategy
Follow existing patterns:
- `status` → `st`
- `detach` → `d`
- `eval` → `e`
- `stack` → No alias (conflicts with `step`/`s`)
- `delete` → `del`
- `disable` → No alias
- `enable` → No alias
- `set` → No alias
- `source` → `src`

## JSON Output Format Design

### Status Command
```json
{
  "command": "status",
  "success": true,
  "result": {
    "status": "break",
    "reason": "ok",
    "filename": "/var/www/app.php",
    "lineno": 42
  }
}
```

### Eval Command
```json
{
  "command": "eval",
  "success": true,
  "result": {
    "expression": "$x + 10",
    "type": "int",
    "value": "52"
  }
}
```

### Set Command
```json
{
  "command": "set",
  "success": true,
  "result": {
    "variable": "$count",
    "value": "42",
    "type": "int"
  }
}
```

### Delete/Disable/Enable Commands
```json
{
  "command": "delete",
  "success": true,
  "result": {
    "breakpoint_id": "123"
  }
}
```

### Source Command
```json
{
  "command": "source",
  "success": true,
  "result": {
    "file": "file:///var/www/app.php",
    "start_line": 100,
    "end_line": 120,
    "source": "<?php\n// line 100\n..."
  }
}
```

## Error Handling Design

### Common Error Cases
1. **Session ended**: "Error: debug session has ended"
2. **Invalid breakpoint ID**: "Error: breakpoint 999 not found"
3. **Invalid variable name**: "Error: variable $invalid not found"
4. **Eval syntax error**: "Error: syntax error in expression"
5. **Command not supported**: "Error: command not supported by Xdebug version"

### Error Format (JSON mode)
```json
{
  "command": "eval",
  "success": false,
  "error": "syntax error in expression"
}
```

## Implementation Strategy

### Phase 1: DBGp Client Methods
Add low-level protocol methods to `internal/dbgp/client.go`:
1. `Status()` - query debugger state
2. `Detach()` - detach from session
3. `Eval()` - evaluate expression
4. `SetProperty()` - modify variable
5. `GetSource()` - retrieve source code
6. `UpdateBreakpoint()` - enable/disable breakpoint

### Phase 2: CLI Command Handlers
Add handlers to `internal/cli/listen.go`:
1. `handleStatus()` - status command
2. `handleDetach()` - detach command
3. `handleEval()` - eval command
4. `handleSet()` - set command
5. `handleSource()` - source command
6. `handleStack()` - stack command (wrapper)
7. `handleDelete()` - delete breakpoint
8. `handleDisable()` - disable breakpoint
9. `handleEnable()` - enable breakpoint

### Phase 3: Daemon Mode Integration
Mirror handlers in `internal/daemon/executor.go`:
- Copy command cases to daemon executor switch
- Return `ipc.CommandResult` instead of view output
- Ensure JSON serialization works

### Phase 4: View Layer Updates
Add display functions to `internal/view/`:
- Format eval results
- Format status information
- Format source code with line numbers
- Add help text for new commands

## Testing Strategy

### Unit Tests
- Test client methods with mock DBGp responses
- Test command parsing edge cases
- Test type detection for `set` command
- Test file path parsing for `source` command

### Integration Tests
- Test full command flow: parse → client → response → display
- Test JSON output format matches spec
- Test error handling and error messages
- Test daemon mode command execution

### Manual Testing Checklist
- Verify each command works in listen mode
- Verify each command works in daemon/attach mode
- Verify JSON output for each command
- Verify help text displays correctly
- Test with actual Xdebug connection

## Backward Compatibility
No breaking changes. All existing commands continue to work unchanged. New commands are additive only.

## Documentation Updates
- Update `CLAUDE.md` with new commands in "Debugging Commands" section
- Add examples for each new command
- Update help text with new command descriptions
- Add to sources/commands.md implementation status table
