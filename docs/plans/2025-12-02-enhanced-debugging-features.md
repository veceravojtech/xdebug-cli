# Enhanced Debugging Features Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use @superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add step_out command, conditional breakpoints, and multiple breakpoints to xdebug-cli

**Architecture:** Three-layer implementation following existing patterns: (1) DBGp protocol client methods, (2) CLI and daemon command handlers, (3) Help documentation. Uses TDD with tests before implementation.

**Tech Stack:** Go 1.x, DBGp protocol, Xdebug

---

## Task 1: Step Out - DBGp Protocol Layer

**Files:**
- Modify: `internal/dbgp/client.go:95-115` (after Next() method)
- Create: `internal/dbgp/client_test.go:195-230` (new test function)

**Step 1: Write the failing test**

File: `internal/dbgp/client_test.go`

Add after `TestClient_Next()` function (around line 195):

```go
func TestClient_StepOut(t *testing.T) {
	mockConn := &MockConnection{
		responseXML: `<?xml version="1.0" encoding="UTF-8"?>
          <response xmlns="urn:debugger_protocol_v1"
                    status="break"
                    reason="ok"
                    command="step_out"
                    transaction_id="1">
            <xdebug:message filename="file:///test.php" lineno="50"/>
          </response>`,
	}

	session := NewSession()
	client := &Client{
		conn:    mockConn,
		session: session,
	}

	response, err := client.StepOut()
	if err != nil {
		t.Fatalf("StepOut() failed: %v", err)
	}

	if response.Command != "step_out" {
		t.Errorf("Expected command 'step_out', got '%s'", response.Command)
	}

	sent := mockConn.LastSent()
	if !strings.Contains(sent, "step_out -i") {
		t.Errorf("Expected 'step_out -i' command, got '%s'", sent)
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/dbgp -run TestClient_StepOut -v`

Expected: FAIL with "client.StepOut undefined (type *Client has no field or method StepOut)"

**Step 3: Write minimal implementation**

File: `internal/dbgp/client.go`

Add after the `Next()` method (around line 115):

```go
// StepOut sends the step_out command
// Steps out of current scope and breaks after returning from current function
// Also known as "finish" in GDB
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

**Step 4: Run test to verify it passes**

Run: `go test ./internal/dbgp -run TestClient_StepOut -v`

Expected: PASS

**Step 5: Run all DBGp tests to ensure no regressions**

Run: `go test ./internal/dbgp -v`

Expected: All tests PASS

**Step 6: Commit**

```bash
git add internal/dbgp/client.go internal/dbgp/client_test.go
git commit -m "feat(dbgp): add StepOut method for step_out command

- Add Client.StepOut() sending step_out DBGp command
- Add test verifying command format and response handling
- Follows existing pattern from Step() and Next() methods"
```

---

## Task 2: Step Out - CLI Handler

**Files:**
- Modify: `internal/cli/listen.go:350-380` (add case), `~900` (add handler)
- Create: `internal/cli/listen_test.go:100-110` (add test case)

**Step 1: Write the failing test**

File: `internal/cli/listen_test.go`

Add to `TestParseCommand` test cases (around line 100):

```go
		{
			name:            "out long form",
			input:           "out",
			expectedCommand: "out",
		},
		{
			name:            "out short form",
			input:           "o",
			expectedCommand: "out",
		},
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/cli -run TestParseCommand -v`

Expected: FAIL with unhandled command

**Step 3: Add command dispatcher case**

File: `internal/cli/listen.go`

In `dispatchCommand()` function (around line 370), add after the "next" case:

```go
	case "out", "o":
		return handleStepOut(client, v)
```

**Step 4: Write test for command handler**

File: `internal/cli/listen_test.go`

Add to test cases array in `TestValidCommands` (around line 210):

```go
		{"out long form", "out", true, "out"},
		{"out short form", "o", true, "out"},
```

**Step 5: Add handleStepOut function**

File: `internal/cli/listen.go`

Add after `handleNext()` function (around line 430):

```go
// handleStepOut steps out of current function
func handleStepOut(client *dbgp.Client, v *view.View) bool {
	response, err := client.StepOut()
	if err != nil {
		if CLIArgs.JSON {
			v.OutputJSON("out", false, err.Error(), nil)
		} else {
			v.PrintErrorLn(fmt.Sprintf("Error: %v", err))
		}
		return false
	}

	return updateState(client, v, response, "out")
}
```

**Step 6: Run tests to verify they pass**

Run: `go test ./internal/cli -run TestParseCommand -v`
Run: `go test ./internal/cli -run TestValidCommands -v`

Expected: All PASS

**Step 7: Commit**

```bash
git add internal/cli/listen.go internal/cli/listen_test.go
git commit -m "feat(cli): add step out command handler

- Add 'out' and 'o' command aliases
- Add handleStepOut() following pattern of handleStep/handleNext
- Add test cases for command parsing and validation"
```

---

## Task 3: Step Out - Daemon Handler

**Files:**
- Modify: `internal/daemon/executor.go:65-75` (add case), `~120` (add handler)

**Step 1: Add command case to executor**

File: `internal/daemon/executor.go`

In `ExecuteCommand()` function (around line 70), add after the "next" case:

```go
	case "out", "o":
		return e.handleStepOut()
```

**Step 2: Add handleStepOut function**

File: `internal/daemon/executor.go`

Add after `handleNext()` function (around line 125):

```go
// handleStepOut steps out of current function
func (e *CommandExecutor) handleStepOut() ipc.CommandResult {
	response, err := e.client.StepOut()
	if err != nil {
		return ipc.CommandResult{
			Command: "out",
			Success: false,
			Error:   err.Error(),
		}
	}

	if response.HasError() {
		return ipc.CommandResult{
			Command: "out",
			Success: false,
			Error:   response.GetErrorMessage(),
		}
	}

	// Update session state
	file, line := response.GetLocation()
	e.session.SetCurrentLocation(file, line)

	return ipc.CommandResult{
		Command: "out",
		Success: true,
		Data: map[string]interface{}{
			"status":   response.Status,
			"filename": file,
			"line":     line,
		},
	}
}
```

**Step 3: Run daemon tests**

Run: `go test ./internal/daemon -v`

Expected: All PASS (no new tests needed, existing executor tests cover new handler)

**Step 4: Commit**

```bash
git add internal/daemon/executor.go
git commit -m "feat(daemon): add step out command to executor

- Add 'out'/'o' command case to ExecuteCommand
- Add handleStepOut() following existing step/next pattern
- Updates session location after step out completes"
```

---

## Task 4: Step Out - Help Documentation

**Files:**
- Modify: `internal/view/help.go:10-20` (main help), `~43-70` (step help)
- Modify: `internal/view/help_test.go:15-30` (test expectations)

**Step 1: Update main help message**

File: `internal/view/help.go`

In `ShowHelpMessage()` function (around line 10), update to:

```go
	help := `
Available Commands:
  run, r          Continue execution to next breakpoint
  step, s         Step into next statement (enter functions)
  next, n         Step over next statement (skip functions)
  out, o          Step out of current function (return to caller)
  finish, f       Stop the debugging session
  break, b        Set a breakpoint (see 'help break' for details)
  print, p        Print variable value (see 'help print' for details)
  context, c      Show variables in current context (see 'help context' for details)
  list, l         Show source code around current line
  info, i         Show debugging information (see 'help info' for details)
  help, h, ?      Show this help message or help for specific command
  quit, q         Exit the debugger

For detailed help on a specific command, type: help <command>
Examples: help break, help print, help info
`
```

**Step 2: Update step help message**

File: `internal/view/help.go`

In `ShowStepHelpMessage()` function (around line 43), update to:

```go
	help := `
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

Examples:
  (at line 10) step     # Enter function if line 10 has a call
  (at line 10) next     # Execute line 10, don't enter functions
  (inside func) out     # Return from current function
`
```

**Step 3: Update help test expectations**

File: `internal/view/help_test.go`

In `TestView_ShowHelpMessage()` function (around line 15), add to expected commands:

```go
	expectedCommands := []string{
		"run, r",
		"step, s",
		"next, n",
		"out, o",          // Add this line
		"finish, f",
		"break, b",
		// ... rest of commands
	}
```

**Step 4: Run help tests**

Run: `go test ./internal/view -run TestView_ShowHelpMessage -v`
Run: `go test ./internal/view -run TestView_ShowStepHelpMessage -v`

Expected: All PASS

**Step 5: Commit**

```bash
git add internal/view/help.go internal/view/help_test.go
git commit -m "docs(help): add step out command documentation

- Add 'out, o' to main help message
- Update step help to include out command
- Add examples demonstrating step/next/out differences
- Update help tests for new command"
```

---

## Task 5: Conditional Breakpoints - Parser Foundation

**Files:**
- Modify: `internal/cli/listen.go:530-600` (enhance handleBreak parser)
- Create: `internal/cli/listen_test.go:320-380` (new test function)

**Step 1: Write failing tests for conditional parsing**

File: `internal/cli/listen_test.go`

Add new test function after existing break tests:

```go
func TestParseBreakpointCondition(t *testing.T) {
	tests := []struct {
		name              string
		args              []string
		expectedLocations []string
		expectedCondition string
		expectError       bool
	}{
		{
			name:              "single location with condition",
			args:              []string{":42", "if", "$count", ">", "10"},
			expectedLocations: []string{":42"},
			expectedCondition: "$count > 10",
			expectError:       false,
		},
		{
			name:              "single location no condition",
			args:              []string{":42"},
			expectedLocations: []string{":42"},
			expectedCondition: "",
			expectError:       false,
		},
		{
			name:              "empty condition after if",
			args:              []string{":42", "if"},
			expectedLocations: []string{":42"},
			expectedCondition: "",
			expectError:       true,
		},
		{
			name:              "complex condition",
			args:              []string{"file.php:100", "if", "$user->isAdmin()", "&&", "$debug"},
			expectedLocations: []string{"file.php:100"},
			expectedCondition: "$user->isAdmin() && $debug",
			expectError:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			locations, condition, err := parseBreakpointArgs(tt.args)

			if tt.expectError && err == nil {
				t.Error("Expected error but got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if !reflect.DeepEqual(locations, tt.expectedLocations) {
				t.Errorf("Expected locations %v, got %v", tt.expectedLocations, locations)
			}

			if condition != tt.expectedCondition {
				t.Errorf("Expected condition '%s', got '%s'", tt.expectedCondition, condition)
			}
		})
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/cli -run TestParseBreakpointCondition -v`

Expected: FAIL with "undefined: parseBreakpointArgs"

**Step 3: Extract parser function**

File: `internal/cli/listen.go`

Add before `handleBreak()` function (around line 520):

```go
// parseBreakpointArgs splits args into locations and condition
// Returns (locations, condition, error)
func parseBreakpointArgs(args []string) ([]string, string, error) {
	if len(args) == 0 {
		return nil, "", fmt.Errorf("no arguments provided")
	}

	// Find "if" keyword to split locations from condition
	ifIndex := -1
	for i, arg := range args {
		if arg == "if" {
			ifIndex = i
			break
		}
	}

	var locations []string
	var condition string

	if ifIndex >= 0 {
		// Has condition
		locations = args[:ifIndex]
		if ifIndex+1 < len(args) {
			condition = strings.Join(args[ifIndex+1:], " ")
		}

		// Validate condition not empty
		if condition == "" {
			return locations, "", fmt.Errorf("condition cannot be empty after 'if'")
		}
	} else {
		// No condition
		locations = args
	}

	return locations, condition, nil
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/cli -run TestParseBreakpointCondition -v`

Expected: PASS

**Step 5: Commit**

```bash
git add internal/cli/listen.go internal/cli/listen_test.go
git commit -m "feat(cli): add breakpoint condition parser

- Add parseBreakpointArgs() to extract locations and condition
- Splits args on 'if' keyword
- Validates condition is not empty
- Add comprehensive tests for parsing logic"
```

---

## Task 6: Conditional Breakpoints - Integrate Parser

**Files:**
- Modify: `internal/cli/listen.go:530-610` (refactor handleBreak to use parser)
- Modify: `internal/daemon/executor.go:195-350` (refactor daemon handleBreak)

**Step 1: Refactor CLI handleBreak to use parser**

File: `internal/cli/listen.go`

In `handleBreak()` function, replace the beginning (after the special cases for "call" and "exception") with:

```go
	// Parse locations and condition
	locations, condition, err := parseBreakpointArgs(args)
	if err != nil {
		if CLIArgs.JSON {
			v.OutputJSON("break", false, err.Error(), nil)
		} else {
			v.PrintErrorLn(fmt.Sprintf("Error: %s", err.Error()))
		}
		return
	}

	// For now, handle only single location (multiple in next task)
	if len(locations) != 1 {
		if CLIArgs.JSON {
			v.OutputJSON("break", false, "multiple breakpoints not yet supported", nil)
		} else {
			v.PrintErrorLn("Error: multiple breakpoints not yet supported")
		}
		return
	}

	location := locations[0]

	// Rest of existing parsing logic for location...
	// (existing code for :line, file:line, etc.)
	// ...

	// At the end where SetBreakpoint is called, pass condition instead of ""
	response, err := client.SetBreakpoint(file, line, condition)
```

**Step 2: Update output message to show condition**

File: `internal/cli/listen.go`

In `handleBreak()` after successful breakpoint set (around line 600):

```go
	if CLIArgs.JSON {
		result := view.JSONBreakpointResult{
			ID:        response.ID,
			Location:  fmt.Sprintf("%s:%d", file, line),
			Condition: condition,  // Add this field
		}
		v.OutputJSON("break", true, "", result)
	} else {
		if condition != "" {
			v.PrintLn(fmt.Sprintf("Breakpoint set at %s:%d with condition '%s' (ID: %s)", file, line, condition, response.ID))
		} else {
			v.PrintLn(fmt.Sprintf("Breakpoint set at %s:%d (ID: %s)", file, line, response.ID))
		}
	}
```

**Step 3: Add Condition field to JSONBreakpointResult**

File: `internal/view/json.go`

Update `JSONBreakpointResult` struct (around line 30):

```go
type JSONBreakpointResult struct {
	ID        string `json:"id"`
	Location  string `json:"location"`
	Condition string `json:"condition,omitempty"`  // Add this field
}
```

**Step 4: Run CLI tests**

Run: `go test ./internal/cli -v`

Expected: All PASS

**Step 5: Apply same pattern to daemon executor**

File: `internal/daemon/executor.go`

In `handleBreak()` function, add the same parser integration and condition passing.

**Step 6: Run daemon tests**

Run: `go test ./internal/daemon -v`

Expected: All PASS

**Step 7: Commit**

```bash
git add internal/cli/listen.go internal/daemon/executor.go internal/view/json.go
git commit -m "feat: integrate conditional breakpoint parser

- Refactor handleBreak to use parseBreakpointArgs
- Pass condition to SetBreakpoint (was always empty string)
- Update output to show condition when set
- Add Condition field to JSON response
- Apply changes to both CLI and daemon handlers"
```

---

## Task 7: Multiple Breakpoints - Parser Enhancement

**Files:**
- Modify: `internal/cli/listen_test.go:380-450` (add multi-location tests)

**Step 1: Write failing tests for multiple locations**

File: `internal/cli/listen_test.go`

Add to `TestParseBreakpointCondition` test cases:

```go
		{
			name:              "multiple locations no condition",
			args:              []string{":42", ":100", ":150"},
			expectedLocations: []string{":42", ":100", ":150"},
			expectedCondition: "",
			expectError:       false,
		},
		{
			name:              "multiple locations with condition",
			args:              []string{":10", ":20", ":30", "if", "$debug"},
			expectedLocations: []string{":10", ":20", ":30"},
			expectedCondition: "$debug",
			expectError:       false,
		},
		{
			name:              "mixed format locations",
			args:              []string{":42", "file.php:100", "other.php:50"},
			expectedLocations: []string{":42", "file.php:100", "other.php:50"},
			expectedCondition: "",
			expectError:       false,
		},
```

**Step 2: Run tests to verify they pass**

Run: `go test ./internal/cli -run TestParseBreakpointCondition -v`

Expected: PASS (parser already handles multiple locations!)

**Step 3: Commit**

```bash
git add internal/cli/listen_test.go
git commit -m "test: add tests for multiple breakpoint locations

- Test multiple locations without condition
- Test multiple locations with shared condition
- Test mixed location formats in one command
- Parser already supports these, tests verify behavior"
```

---

## Task 8: Multiple Breakpoints - Handler Implementation

**Files:**
- Modify: `internal/cli/listen.go:530-650` (refactor to loop over locations)

**Step 1: Remove single-location restriction**

File: `internal/cli/listen.go`

In `handleBreak()`, remove this block:

```go
	// For now, handle only single location (multiple in next task)
	if len(locations) != 1 {
		if CLIArgs.JSON {
			v.OutputJSON("break", false, "multiple breakpoints not yet supported", nil)
		} else {
			v.PrintErrorLn("Error: multiple breakpoints not yet supported")
		}
		return
	}

	location := locations[0]
```

**Step 2: Refactor to loop over locations**

File: `internal/cli/listen.go`

Replace the location parsing and breakpoint setting with:

```go
	type breakpointResult struct {
		ID       string
		Location string
		Error    string
	}

	var results []breakpointResult

	for _, location := range locations {
		var file string
		var line int

		// Parse location (existing logic)
		if strings.HasPrefix(location, ":") {
			lineStr := strings.TrimPrefix(location, ":")
			parsedLine, err := strconv.Atoi(lineStr)
			if err != nil {
				results = append(results, breakpointResult{
					Location: location,
					Error:    fmt.Sprintf("Invalid line number: %s", lineStr),
				})
				continue
			}
			line = parsedLine
			file, _ = client.GetSession().GetCurrentLocation()
			if file == "" {
				results = append(results, breakpointResult{
					Location: location,
					Error:    "No current file. Use format: break <file>:<line>",
				})
				continue
			}
		} else if strings.Contains(location, ":") {
			parts := strings.SplitN(location, ":", 2)
			file = parts[0]
			parsedLine, err := strconv.Atoi(parts[1])
			if err != nil {
				results = append(results, breakpointResult{
					Location: location,
					Error:    fmt.Sprintf("Invalid line number: %s", parts[1]),
				})
				continue
			}
			line = parsedLine
		} else {
			parsedLine, err := strconv.Atoi(location)
			if err != nil {
				results = append(results, breakpointResult{
					Location: location,
					Error:    fmt.Sprintf("Invalid breakpoint format: %s", location),
				})
				continue
			}
			line = parsedLine
			file, _ = client.GetSession().GetCurrentLocation()
			if file == "" {
				results = append(results, breakpointResult{
					Location: location,
					Error:    "No current file. Use format: break <file>:<line>",
				})
				continue
			}
		}

		// Set breakpoint with condition
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
		jsonResults := make([]view.JSONBreakpointResult, 0, len(results))
		for _, r := range results {
			jsonResults = append(jsonResults, view.JSONBreakpointResult{
				ID:        r.ID,
				Location:  r.Location,
				Error:     r.Error,
				Condition: condition,
			})
		}
		v.OutputJSONArray("break", jsonResults)
	} else {
		successCount := 0
		for _, r := range results {
			if r.Error == "" {
				if condition != "" {
					v.PrintLn(fmt.Sprintf("Breakpoint set at %s with condition '%s' (ID: %s)", r.Location, condition, r.ID))
				} else {
					v.PrintLn(fmt.Sprintf("Breakpoint set at %s (ID: %s)", r.Location, r.ID))
				}
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

**Step 3: Add Error field to JSONBreakpointResult**

File: `internal/view/json.go`

Update struct:

```go
type JSONBreakpointResult struct {
	ID        string `json:"id,omitempty"`
	Location  string `json:"location"`
	Condition string `json:"condition,omitempty"`
	Error     string `json:"error,omitempty"`
}
```

**Step 4: Add OutputJSONArray method**

File: `internal/view/json.go`

Add after `OutputJSON()` method (around line 75):

```go
// OutputJSONArray outputs an array of results as JSON
func (v *View) OutputJSONArray(command string, data interface{}) {
	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		// Fallback error output
		fmt.Fprintln(v.stdout, `{"error": "failed to marshal JSON"}`)
		return
	}
	fmt.Fprintln(v.stdout, string(jsonBytes))
}
```

**Step 5: Run tests**

Run: `go test ./internal/cli -v`
Run: `go test ./internal/view -v`

Expected: All PASS

**Step 6: Commit**

```bash
git add internal/cli/listen.go internal/view/json.go
git commit -m "feat: implement multiple breakpoints in one command

- Refactor handleBreak to loop over all locations
- Collect results for each breakpoint (success or error)
- Show individual results and total count
- Support mixed valid/invalid locations (partial success)
- Add Error field to JSON response
- Add OutputJSONArray for array responses"
```

---

## Task 9: Multiple Breakpoints - Daemon Handler

**Files:**
- Modify: `internal/daemon/executor.go:195-350` (apply same pattern to daemon)

**Step 1: Apply multiple breakpoint pattern to daemon**

File: `internal/daemon/executor.go`

Refactor `handleBreak()` following the same pattern as CLI:
- Parse with `parseBreakpointArgs()` (need to copy/extract to shared location)
- Loop over locations
- Collect results
- Return array of breakpoint results

**Step 2: Extract parseBreakpointArgs to shared location**

Create: `internal/common/breakpoint_parser.go`

Move `parseBreakpointArgs()` from `internal/cli/listen.go` to new shared file:

```go
package common

import (
	"fmt"
	"strings"
)

// ParseBreakpointArgs splits args into locations and condition
// Returns (locations, condition, error)
func ParseBreakpointArgs(args []string) ([]string, string, error) {
	if len(args) == 0 {
		return nil, "", fmt.Errorf("no arguments provided")
	}

	// Find "if" keyword to split locations from condition
	ifIndex := -1
	for i, arg := range args {
		if arg == "if" {
			ifIndex = i
			break
		}
	}

	var locations []string
	var condition string

	if ifIndex >= 0 {
		// Has condition
		locations = args[:ifIndex]
		if ifIndex+1 < len(args) {
			condition = strings.Join(args[ifIndex+1:], " ")
		}

		// Validate condition not empty
		if condition == "" {
			return locations, "", fmt.Errorf("condition cannot be empty after 'if'")
		}
	} else {
		// No condition
		locations = args
	}

	return locations, condition, nil
}
```

**Step 3: Update imports in CLI and daemon**

Update `internal/cli/listen.go` and `internal/daemon/executor.go` to use `common.ParseBreakpointArgs()`

**Step 4: Move tests to common package**

Create: `internal/common/breakpoint_parser_test.go`

Move tests from `internal/cli/listen_test.go` to new file.

**Step 5: Update daemon executor**

File: `internal/daemon/executor.go`

Apply the loop pattern to handle multiple breakpoints in daemon.

**Step 6: Run all tests**

Run: `go test ./internal/common -v`
Run: `go test ./internal/cli -v`
Run: `go test ./internal/daemon -v`

Expected: All PASS

**Step 7: Commit**

```bash
git add internal/common/ internal/cli/listen.go internal/daemon/executor.go
git commit -m "refactor: extract breakpoint parser to shared package

- Create internal/common/breakpoint_parser.go
- Move ParseBreakpointArgs to common package
- Update CLI and daemon to use shared parser
- Move tests to common package
- Apply multiple breakpoint pattern to daemon executor"
```

---

## Task 10: Break Help Documentation

**Files:**
- Modify: `internal/view/help.go:80-160` (add ShowBreakHelpMessage)
- Modify: `internal/view/help.go:160-180` (update ShowHelp to route to break help)

**Step 1: Add break help message**

File: `internal/view/help.go`

Add new function after existing help functions:

```go
// ShowBreakHelpMessage displays help for the break command
func (v *View) ShowBreakHelpMessage() {
	help := `
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
`
	v.PrintLn(help)
}
```

**Step 2: Update ShowHelp router**

File: `internal/view/help.go`

In `ShowHelp()` function, add case for "break":

```go
func (v *View) ShowHelp(topic string) {
	switch topic {
	case "":
		v.ShowHelpMessage()
	case "break", "b":
		v.ShowBreakHelpMessage()
	case "print", "p":
		v.ShowPrintHelpMessage()
	case "context", "c":
		v.ShowContextHelpMessage()
	case "info", "i":
		v.ShowInfoHelpMessage()
	default:
		v.PrintLn(fmt.Sprintf("No help available for '%s'", topic))
		v.PrintLn("Available help topics: break, print, context, info")
	}
}
```

**Step 3: Add test for break help**

File: `internal/view/help_test.go`

Add new test function:

```go
func TestView_ShowBreakHelpMessage(t *testing.T) {
	var buf bytes.Buffer
	v := &View{stdout: &buf}

	v.ShowBreakHelpMessage()

	output := buf.String()

	expectedContent := []string{
		"break - Set breakpoint",
		"Multiple Breakpoints",
		"Conditional Breakpoints",
		"break :42 if $count > 10",
		"break call myFunction",
	}

	for _, content := range expectedContent {
		if !strings.Contains(output, content) {
			t.Errorf("ShowBreakHelpMessage() missing expected content: %q", content)
		}
	}
}
```

**Step 4: Run help tests**

Run: `go test ./internal/view -run TestView_ShowBreakHelpMessage -v`

Expected: PASS

**Step 5: Commit**

```bash
git add internal/view/help.go internal/view/help_test.go
git commit -m "docs(help): add comprehensive break command help

- Add ShowBreakHelpMessage with examples
- Document conditional breakpoints syntax
- Document multiple breakpoints syntax
- Update help router to handle 'help break'
- Add test for break help content"
```

---

## Task 11: Integration Test - Step Out

**Files:**
- Create: `test/integration/test-step-out.sh`

**Step 1: Create step out integration test**

File: `test/integration/test-step-out.sh`

```bash
#!/bin/bash
# Step Out Integration Test

TEST_URL="http://booking.previo.loc/coupon/index/select/?hotId=731541&currency=CZK&lang=cs&XDEBUG_TRIGGER=1"
OUTER_BREAKPOINT="booking/application/modules/default/models/Controller/Action/Helper/AllowedTabs.php:144"
INNER_BREAKPOINT="booking/application/modules/default/models/Controller/Action/Helper/AllowedTabs.php:168"

echo "=========================================="
echo "Step Out Integration Test"
echo "=========================================="

# Test: Step into function, then step out
output=$(mktemp)

xdebug-cli listen -l 0.0.0.0 --non-interactive --force \
  --commands "break $OUTER_BREAKPOINT" \
             "run" \
             "step" \
             "step" \
             "step" \
             "out" \
             "list" > "$output" 2>&1 &
XDEBUG_PID=$!

sleep 0.5
curl -s "$TEST_URL" > /dev/null 2>&1 &
CURL_PID=$!

# Wait for xdebug-cli
for i in {1..100}; do
  if ! ps -p $XDEBUG_PID > /dev/null 2>&1; then
    wait $XDEBUG_PID 2>/dev/null
    EXIT_CODE=$?
    break
  fi
  sleep 0.1
done

kill $CURL_PID 2>/dev/null || true

echo "--- Output ---"
cat "$output"
echo "--- End Output ---"

# Check that we stepped out successfully
if grep -q "out" "$output"; then
  echo "✓ Step out command executed"
else
  echo "✗ Step out command not found in output"
  EXIT_CODE=1
fi

rm -f "$output"

if [ $EXIT_CODE -eq 0 ]; then
  echo "✓ Test passed"
else
  echo "✗ Test failed (exit code: $EXIT_CODE)"
fi

exit $EXIT_CODE
```

**Step 2: Make test executable**

Run: `chmod +x test/integration/test-step-out.sh`

**Step 3: Run integration test**

Run: `./test/integration/test-step-out.sh`

Expected: Test passes, shows step out execution

**Step 4: Commit**

```bash
git add test/integration/test-step-out.sh
git commit -m "test: add step out integration test

- Test step into nested function then step out
- Verify execution returns to caller
- Verify list shows correct location after step out"
```

---

## Task 12: Integration Test - Conditional Breakpoints

**Files:**
- Create: `test/integration/test-conditional-breakpoint.sh`

**Step 1: Create conditional breakpoint test**

File: `test/integration/test-conditional-breakpoint.sh`

```bash
#!/bin/bash
# Conditional Breakpoint Integration Test

TEST_URL="http://booking.previo.loc/coupon/index/select/?hotId=731541&currency=CZK&lang=cs&XDEBUG_TRIGGER=1"
BREAKPOINT="booking/application/modules/default/models/Controller/Action/Helper/AllowedTabs.php:144"

echo "=========================================="
echo "Conditional Breakpoint Test"
echo "=========================================="

output=$(mktemp)

# Test: Set breakpoint with condition
xdebug-cli listen -l 0.0.0.0 --non-interactive --force \
  --commands "break $BREAKPOINT if \$hotId > 700000" \
             "run" \
             "print \$hotId" > "$output" 2>&1 &
XDEBUG_PID=$!

sleep 0.5
curl -s "$TEST_URL" > /dev/null 2>&1 &
CURL_PID=$!

for i in {1..100}; do
  if ! ps -p $XDEBUG_PID > /dev/null 2>&1; then
    wait $XDEBUG_PID 2>/dev/null
    EXIT_CODE=$?
    break
  fi
  sleep 0.1
done

kill $CURL_PID 2>/dev/null || true

echo "--- Output ---"
cat "$output"
echo "--- End Output ---"

# Check for condition in output
if grep -q "with condition" "$output"; then
  echo "✓ Conditional breakpoint set"
else
  echo "✗ Condition not shown in output"
  EXIT_CODE=1
fi

# Check breakpoint was hit (condition was true)
if grep -q "731541" "$output"; then
  echo "✓ Breakpoint hit with correct value"
else
  echo "✗ Breakpoint did not trigger as expected"
  EXIT_CODE=1
fi

rm -f "$output"

if [ $EXIT_CODE -eq 0 ]; then
  echo "✓ Test passed"
else
  echo "✗ Test failed (exit code: $EXIT_CODE)"
fi

exit $EXIT_CODE
```

**Step 2: Make test executable and run**

Run: `chmod +x test/integration/test-conditional-breakpoint.sh`
Run: `./test/integration/test-conditional-breakpoint.sh`

Expected: Test passes, shows conditional breakpoint working

**Step 3: Commit**

```bash
git add test/integration/test-conditional-breakpoint.sh
git commit -m "test: add conditional breakpoint integration test

- Test breakpoint with condition expression
- Verify breakpoint only triggers when condition true
- Verify condition shown in output"
```

---

## Task 13: Integration Test - Multiple Breakpoints

**Files:**
- Create: `test/integration/test-multiple-breakpoints.sh`

**Step 1: Create multiple breakpoints test**

File: `test/integration/test-multiple-breakpoints.sh`

```bash
#!/bin/bash
# Multiple Breakpoints Integration Test

TEST_URL="http://booking.previo.loc/coupon/index/select/?hotId=731541&currency=CZK&lang=cs&XDEBUG_TRIGGER=1"
FILE="booking/application/modules/default/models/Controller/Action/Helper/AllowedTabs.php"

echo "=========================================="
echo "Multiple Breakpoints Test"
echo "=========================================="

output=$(mktemp)

# Test: Set 3 breakpoints in one command
xdebug-cli listen -l 0.0.0.0 --non-interactive --force \
  --commands "break $FILE:57 $FILE:144 $FILE:168" \
             "run" \
             "run" \
             "run" > "$output" 2>&1 &
XDEBUG_PID=$!

sleep 0.5
curl -s "$TEST_URL" > /dev/null 2>&1 &
CURL_PID=$!

for i in {1..100}; do
  if ! ps -p $XDEBUG_PID > /dev/null 2>&1; then
    wait $XDEBUG_PID 2>/dev/null
    EXIT_CODE=$?
    break
  fi
  sleep 0.1
done

kill $CURL_PID 2>/dev/null || true

echo "--- Output ---"
cat "$output"
echo "--- End Output ---"

# Count breakpoints set
bp_count=$(grep -c "Breakpoint set at" "$output" || echo 0)

if [ "$bp_count" -eq 3 ]; then
  echo "✓ All 3 breakpoints set"
else
  echo "✗ Expected 3 breakpoints, got $bp_count"
  EXIT_CODE=1
fi

# Check for success count message
if grep -q "3 breakpoints set successfully" "$output"; then
  echo "✓ Success count shown"
else
  echo "! Success count message not found (may be OK)"
fi

rm -f "$output"

if [ $EXIT_CODE -eq 0 ]; then
  echo "✓ Test passed"
else
  echo "✗ Test failed (exit code: $EXIT_CODE)"
fi

exit $EXIT_CODE
```

**Step 2: Make test executable and run**

Run: `chmod +x test/integration/test-multiple-breakpoints.sh`
Run: `./test/integration/test-multiple-breakpoints.sh`

Expected: Test passes, shows 3 breakpoints set

**Step 3: Commit**

```bash
git add test/integration/test-multiple-breakpoints.sh
git commit -m "test: add multiple breakpoints integration test

- Test setting 3 breakpoints in one command
- Verify all breakpoints created successfully
- Verify count message shown"
```

---

## Task 14: Update CLAUDE.md Documentation

**Files:**
- Modify: `CLAUDE.md:18-30` (update command list), `~47-85` (update REPL commands)

**Step 1: Update Interactive REPL Commands section**

File: `CLAUDE.md`

Update the "Interactive REPL Commands" section (around line 47):

```markdown
## Interactive REPL Commands

Once connected via `xdebug-cli listen`, use these commands:

```
run, r              # Continue execution
step, s             # Step into
next, n             # Step over
out, o              # Step out of current function
break, b <target>   # Set breakpoint (:line, file:line, call func, exception)
print, p <var>      # Print variable value
context, c [type]   # Show variables (local/global/constant)
list, l             # Show source code
info, i [topic]     # Show info (breakpoints)
finish, f           # Stop debugging
help, h, ?          # Show help
quit, q             # Exit debugger
```

### Breakpoint Syntax

**Basic breakpoints:**
```bash
break :42                # Line in current file
break file.php:100       # Specific file and line
break call myFunction    # Function call
break exception          # Any exception
```

**Conditional breakpoints:**
```bash
break :42 if $count > 10
break file.php:100 if $user->isAdmin()
```

**Multiple breakpoints:**
```bash
break :42 :100 :150
break file.php:10 file.php:20 if $debug
```
```

**Step 2: Update Features section**

File: `CLAUDE.md`

Update the Features section (around line 10):

```markdown
## Features
* DBGp protocol client for PHP debugging with Xdebug
* Interactive REPL debugging session
* Non-interactive commands for scripting
* Daemon mode for persistent debug sessions (multi-step workflows)
* TCP server for accepting Xdebug connections
* Full debugging operations: run, step (into/over/out), breakpoints, variable inspection
* Conditional breakpoints with PHP expressions
* Multiple breakpoints in single command
* Source code display with line numbers
* TDD with comprehensive test coverage
* Install command (`xdebug-cli install`) installs CLI to `~/.local/bin` with build timestamp
```

**Step 3: Run documentation check**

Run: `grep -n "step out\|out, o\|conditional\|multiple breakpoint" CLAUDE.md`

Expected: Shows new documentation entries

**Step 4: Commit**

```bash
git add CLAUDE.md
git commit -m "docs: update CLAUDE.md with new debugging features

- Add step out command to REPL commands list
- Document conditional breakpoint syntax
- Document multiple breakpoint syntax
- Add examples for new features
- Update Features list"
```

---

## Task 15: Final Testing and Verification

**Files:**
- All previously created files

**Step 1: Run full test suite**

Run: `go test ./...`

Expected: All tests PASS

**Step 2: Run all integration tests**

Run: `./test/integration/test-xdebug-noninteractive.sh`
Run: `./test/integration/test-step-out.sh`
Run: `./test/integration/test-conditional-breakpoint.sh`
Run: `./test/integration/test-multiple-breakpoints.sh`

Expected: All integration tests PASS

**Step 3: Build binary**

Run: `go build -o xdebug-cli ./cmd/xdebug-cli`

Expected: Build succeeds

**Step 4: Manual smoke test - step out**

Run: `./xdebug-cli listen`

In another terminal, trigger PHP with breakpoint, try:
```
break :100
run
step
step
out
list
```

Expected: Steps out of function successfully

**Step 5: Manual smoke test - conditional breakpoint**

Run: `./xdebug-cli listen`

Try:
```
break :42 if $count > 10
run
```

Expected: Breakpoint set with condition shown

**Step 6: Manual smoke test - multiple breakpoints**

Run: `./xdebug-cli listen`

Try:
```
break :42 :100 :150
```

Expected: 3 breakpoints set, IDs shown

**Step 7: Final commit**

```bash
git add .
git commit -m "feat: complete enhanced debugging features

Summary of changes:
- Add step_out command (out/o aliases)
- Add conditional breakpoints (if syntax)
- Add multiple breakpoints (space-separated)
- Comprehensive tests and documentation

All tests passing. Ready for use.

🤖 Generated with [Claude Code](https://claude.com/claude-code)

Co-Authored-By: Claude <noreply@anthropic.com>"
```

---

## Implementation Complete

This plan implements all four enhanced debugging features:
1. ✅ Step out command
2. ✅ Conditional breakpoints
3. ✅ Multiple breakpoints
4. ✅ Daemon mode validation (existing behavior confirmed)

**Total estimated time:** 4-6 hours
**Test coverage:** Unit tests + integration tests + manual verification
**Backward compatibility:** 100% maintained
