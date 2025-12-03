# Enhanced Debugging Features Design

**Date:** 2025-12-02
**Status:** Design Validated
**Features:** step_out, conditional breakpoints, multiple breakpoints, daemon mode improvements

## Overview

This document describes the design for enhancing xdebug-cli with advanced debugging features that improve workflow efficiency and match standard debugger capabilities.

## Requirements

### Features to Implement

1. **step_out command** - Step out of current function and return to caller
2. **Conditional breakpoints** - Break only when a condition is true
3. **Multiple breakpoints** - Set multiple breakpoints in one command
4. **Daemon mode breakpoints** - Validate breakpoints work during active sessions

### Feature Already Covered

- **continue command** - Already exists as `run`/`r` (continues execution until next breakpoint)

## Design Decisions

### 1. Step Out Command

**Command Syntax:**
- `out` - Full command name
- `o` - Short form alias

**Behavior:**
- Steps out of current scope
- Breaks on statement after returning from current function
- Equivalent to GDB's "finish" command

**DBGp Protocol:**
- Sends `step_out -i {transaction_id}` command
- Supported by Xdebug 2.1+

**Use Case:**
```bash
# You stepped into a function
step              # Enter calculateTotal()

# Inside function, realize you don't need to debug it
out               # Return to caller, stop after function returns

# Now at line after calculateTotal() call
```

### 2. Conditional Breakpoints

**Command Syntax:**
```bash
break <location> if <condition>
```

**Examples:**
```bash
break :42 if $count > 10
break file.php:100 if $user->isAdmin()
break :150 if $items !== null && count($items) > 0
```

**Behavior:**
- Breakpoint only triggers when condition evaluates to true
- Condition is PHP expression evaluated in current scope
- Works with all breakpoint types (line, call, exception)

**Implementation:**
- Reuses existing `SetBreakpoint(file, line, condition)` parameter
- Condition is base64-encoded per DBGp protocol
- Parser splits args on `if` keyword

**Error Handling:**
- Empty condition after `if`: Show error
- Invalid PHP syntax: Xdebug returns error (propagate to user)

**Output:**
- Text: `Breakpoint set at file.php:42 with condition '$count > 10' (ID: 12345)`
- JSON: `{"id": "12345", "location": "file.php:42", "condition": "$count > 10"}`

### 3. Multiple Breakpoints

**Command Syntax:**
```bash
break <location1> <location2> <location3> ...
```

**Examples:**
```bash
break :42 :100 :150
break file.php:42 file.php:100 other.php:50
break :10 :20 :30 if $debug
```

**Behavior:**
- Sets breakpoints at all specified locations
- If condition provided, applies to ALL breakpoints
- Returns array of breakpoint IDs

**Implementation:**
- Parser splits args on `if` keyword
- Everything before `if` = locations
- Everything after `if` = condition
- Loop through locations, call `SetBreakpoint()` for each

**Output:**
- Text: `3 breakpoints set: file.php:42 (ID: 101), file.php:100 (ID: 102), file.php:150 (ID: 103)`
- JSON: Array of breakpoint results

**Design Choice:**
- Same condition applies to all breakpoints in one command
- Keeps parsing simple and matches common use case
- If different conditions needed, use multiple commands

### 4. Daemon Mode Breakpoints

**Current Behavior:**
Daemon mode already supports setting breakpoints during paused sessions.

**Validation:**
Confirm existing code works correctly with new features (conditional, multiple).

**Examples:**
```bash
# Start daemon with initial breakpoint
xdebug-cli listen --daemon --commands "break :42"

# Trigger PHP request (pauses at line 42)
curl http://localhost/app.php -b "XDEBUG_TRIGGER=1"

# Add more breakpoints while paused (with conditions)
xdebug-cli attach --commands "break :100 :150 if $debug" "run"
```

**Requirements:**
- Session must be paused (at breakpoint or just connected)
- DBGp protocol allows `breakpoint_set` at any time
- No special handling needed

**Error Cases:**
- Session ended: `Error: session ended`
- Connection lost: `Error: failed to connect to daemon`
- Invalid breakpoint: Xdebug error propagated to user

## Architecture

### Layered Implementation

**Layer 1: DBGp Protocol (internal/dbgp/client.go)**
- Add `StepOut()` method sending `step_out -i {txid}`
- `SetBreakpoint()` already supports conditions (just needs to be used)
- No changes needed for multiple breakpoints (called in loop)

**Layer 2: Command Handlers (internal/daemon/executor.go & internal/cli/listen.go)**
- Add `handleStepOut()` following pattern of `handleStep()` and `handleNext()`
- Modify `handleBreak()` to parse conditions (split on `if` keyword)
- Modify `handleBreak()` to parse multiple locations (loop before `if`)
- No changes for daemon mode (already works)

**Layer 3: CLI Interface (commands & help)**
- Add `out`/`o` command case to dispatcher
- Update `break` command parser for `if` conditions
- Update `break` command parser for multiple locations
- Update help documentation

### Parsing Flow for Enhanced Break Command

```
Input: "break :42 :100 file.php:150 if $count > 10"

Step 1: Split on "if" keyword
  - locations = [":42", ":100", "file.php:150"]
  - condition = "$count > 10"

Step 2: For each location:
  - Parse location (existing logic handles :42, file.php:150, etc.)
  - Extract file and line
  - Call SetBreakpoint(file, line, condition)
  - Collect breakpoint ID and location

Step 3: Output results
  - Text: "3 breakpoints set: ..."
  - JSON: Array of breakpoint results
```

## Implementation Details

### Step Out Command

**DBGp Client (internal/dbgp/client.go)**

```go
// StepOut sends the step_out command
// Steps out of current scope and breaks after returning from current function
func (c *Client) StepOut() (*ProtocolResponse, error) {
    txID := c.session.NextTransactionIDInt()
    command := fmt.Sprintf("step_out -i %d", txID)
    c.session.AddCommand(strconv.Itoa(txID), "step_out")

    err := c.conn.SendMessage(command)
    if err != nil {
        return nil, err
    }

    response, err := c.conn.GetResponse()
    if err != nil {
        return nil, err
    }

    c.updateSessionFromResponse(response)
    return response, nil
}
```

**Command Handler Pattern**

Both `internal/cli/listen.go` and `internal/daemon/executor.go` need:

```go
case "out", "o":
    return handleStepOut(client, v)

func handleStepOut(client *dbgp.Client, v *view.View) bool {
    response, err := client.StepOut()
    if err != nil {
        v.PrintErrorLn(fmt.Sprintf("Error: %v", err))
        return false
    }
    return updateState(client, v, response, "step_out")
}
```

### Conditional Breakpoints

**Parser Enhancement (internal/cli/listen.go line ~530)**

```go
// After parsing args, check for condition
var condition string
var locations []string

// Find "if" keyword
ifIndex := -1
for i, arg := range args {
    if arg == "if" {
        ifIndex = i
        break
    }
}

if ifIndex > 0 {
    locations = args[:ifIndex]
    condition = strings.Join(args[ifIndex+1:], " ")
    if condition == "" {
        v.PrintErrorLn("Error: condition cannot be empty after 'if'")
        return
    }
} else {
    locations = args
}

// Rest of parsing continues with locations and condition
```

### Multiple Breakpoints

**Loop Through Locations**

```go
type breakpointResult struct {
    ID       string
    Location string
    Error    string
}

var results []breakpointResult

for _, loc := range locations {
    // Parse location (existing logic)
    file, line, err := parseLocation(loc, client)
    if err != nil {
        results = append(results, breakpointResult{
            Location: loc,
            Error:    err.Error(),
        })
        continue
    }

    // Set breakpoint
    response, err := client.SetBreakpoint(file, line, condition)
    if err != nil {
        results = append(results, breakpointResult{
            Location: fmt.Sprintf("%s:%d", file, line),
            Error:    err.Error(),
        })
        continue
    }

    if response.HasError() {
        results = append(results, breakpointResult{
            Location: fmt.Sprintf("%s:%d", file, line),
            Error:    response.GetErrorMessage(),
        })
        continue
    }

    results = append(results, breakpointResult{
        ID:       response.ID,
        Location: fmt.Sprintf("%s:%d", file, line),
    })
}

// Output results
if CLIArgs.JSON {
    v.OutputJSONArray("break", results)
} else {
    successCount := 0
    for _, r := range results {
        if r.Error == "" {
            v.PrintLn(fmt.Sprintf("Breakpoint set at %s (ID: %s)", r.Location, r.ID))
            successCount++
        } else {
            v.PrintErrorLn(fmt.Sprintf("Failed to set breakpoint at %s: %s", r.Location, r.Error))
        }
    }
    if successCount > 1 {
        v.PrintLn(fmt.Sprintf("\n%d breakpoints set successfully", successCount))
    }
}
```

## Help Documentation Updates

### Main Help Message

```
Available Commands:
  run, r          Continue execution to next breakpoint
  step, s         Step into next statement (enter functions)
  next, n         Step over next statement (skip functions)
  out, o          Step out of current function (return to caller)
  finish, f       Stop the debugging session
  break, b        Set breakpoint(s) (see 'help break' for details)
  ...
```

### Step Help Message

```
step/next/out - Control execution flow

Commands:
  step, s    Step into next statement (enters function calls)
  next, n    Step over next statement (executes functions without stepping through)
  out, o     Step out of current function (returns to caller)

Usage:
  step       Execute one statement, entering any functions
  next       Execute one statement, treating function calls as single steps
  out        Execute until current function returns

The difference:
  - 'step' enters function definitions so you can debug inside them
  - 'next' executes the entire function and stops at the next line
  - 'out' finishes the current function and stops after it returns
```

### Break Help Message

```
break - Set breakpoint(s)

Usage:
  break <line>                    Set breakpoint at line in current file
  break :<line>                   Set breakpoint at line in current file
  break <file>:<line>             Set breakpoint at specific file and line
  break call <function>           Set breakpoint on function call
  break exception [<name>]        Set breakpoint on exception

Multiple Breakpoints:
  break <loc1> <loc2> <loc3>      Set multiple breakpoints

Conditional Breakpoints:
  break <location> if <condition>

Examples:
  break 42                        Breakpoint at line 42 in current file
  break :100                      Breakpoint at line 100 in current file
  break app.php:50                Breakpoint at app.php line 50
  break :42 :100 :150             Three breakpoints in current file
  break file.php:10 file.php:20   Two breakpoints in file.php
  break :42 if $count > 10        Conditional breakpoint
  break app.php:100 if $user->isAdmin()  Conditional in specific file
  break :10 :20 if $debug         Multiple conditional breakpoints
  break call myFunction           Breakpoint on function call
  break exception                 Break on any exception
  break exception MyException     Break on specific exception
```

## Testing Strategy

### Unit Tests

**Step Out Command:**
- `TestClient_StepOut` - Verify DBGp command format
- `TestHandleStepOut` - Verify command handler logic

**Conditional Breakpoints:**
- `TestHandleBreak_WithCondition` - Parse single location with condition
- `TestHandleBreak_EmptyCondition` - Reject empty condition after `if`
- `TestHandleBreak_ComplexCondition` - Handle multi-word conditions

**Multiple Breakpoints:**
- `TestHandleBreak_MultipleLocations` - Parse multiple locations
- `TestHandleBreak_MultipleWithCondition` - Multiple locations with shared condition
- `TestHandleBreak_MixedFormats` - Different location formats in one command

### Integration Tests

**Test Scenarios:**

1. **Step Out Workflow**
   - Set breakpoint in nested function
   - Step into function
   - Use `out` to return to caller
   - Verify stopped at correct line

2. **Conditional Breakpoint**
   - Set breakpoint with condition
   - Trigger code with condition false (should not break)
   - Trigger code with condition true (should break)
   - Verify stopped only when condition met

3. **Multiple Breakpoints**
   - Set 3 breakpoints in one command
   - Verify all 3 IDs returned
   - Run code that hits each breakpoint
   - Verify stops at each location

4. **Daemon Mode with New Features**
   - Start daemon
   - Attach and set conditional breakpoint
   - Attach and set multiple breakpoints
   - Verify breakpoints active in session

## Edge Cases and Error Handling

### Step Out
- **At top-level scope:** Xdebug returns error (propagate to user)
- **Session ended:** Return "Error: session ended"

### Conditional Breakpoints
- **Empty condition after `if`:** Show "Error: condition cannot be empty after 'if'"
- **Invalid PHP syntax:** Xdebug validates, return error
- **Condition with special chars:** Already handled (base64 encoded)

### Multiple Breakpoints
- **Mixed valid/invalid locations:** Set valid ones, report errors for invalid
- **Same location multiple times:** Xdebug handles (may return duplicate IDs or error)
- **Too many breakpoints:** No client-side limit, Xdebug may have limit

### Daemon Mode
- **Session not paused:** Command fails with error
- **Connection lost during multi-breakpoint:** Partial success reported

## Backward Compatibility

**All changes are backward compatible:**

- `break :42` still works (no condition)
- `break file.php:100` still works
- Existing breakpoint commands unchanged
- New features are additive, not breaking

## Performance Considerations

**Multiple Breakpoints:**
- N sequential calls to `SetBreakpoint()`
- Network round-trip for each
- Acceptable for reasonable counts (< 100)

**Conditional Breakpoints:**
- Condition evaluated by Xdebug on each pass
- Performance depends on condition complexity
- User responsibility to write efficient conditions

## Future Enhancements (Out of Scope)

- **Different conditions per breakpoint:** Require multiple commands
- **Breakpoint templates:** Save/load breakpoint sets
- **Hit count conditions:** `break :42 hitcount 5` (requires DBGp extension)
- **Temporary breakpoints:** Auto-delete after first hit

## Implementation Checklist

- [ ] Add `StepOut()` to `internal/dbgp/client.go`
- [ ] Add `handleStepOut()` to `internal/cli/listen.go`
- [ ] Add `handleStepOut()` to `internal/daemon/executor.go`
- [ ] Modify `handleBreak()` to parse conditions
- [ ] Modify `handleBreak()` to parse multiple locations
- [ ] Update help messages (main, step, break)
- [ ] Add unit tests for step_out
- [ ] Add unit tests for conditional parsing
- [ ] Add unit tests for multiple breakpoint parsing
- [ ] Add integration test for step_out workflow
- [ ] Add integration test for conditional breakpoints
- [ ] Add integration test for multiple breakpoints
- [ ] Add integration test for daemon mode with new features
- [ ] Update CLAUDE.md documentation
- [ ] Update README if needed

## Conclusion

This design adds powerful debugging features while maintaining simplicity and backward compatibility. The implementation reuses existing patterns and protocols, minimizing complexity and risk.

**Key Benefits:**
- **step_out:** Navigate call stack efficiently
- **Conditional breakpoints:** Reduce manual breakpoint management
- **Multiple breakpoints:** Set up debugging sessions faster
- **Daemon mode:** Validates existing behavior works with new features

**Implementation Effort:**
- Low risk - extends existing patterns
- No breaking changes
- Well-defined test cases
