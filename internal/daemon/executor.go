package daemon

import (
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/console/xdebug-cli/internal/cfg"
	"github.com/console/xdebug-cli/internal/dbgp"
	"github.com/console/xdebug-cli/internal/ipc"
	"github.com/console/xdebug-cli/internal/view"
)

// CommandExecutor executes debug commands and returns structured results
type CommandExecutor struct {
	client     *dbgp.Client
	mu         sync.Mutex
	jsonOutput bool
}

// NewCommandExecutor creates a new command executor
func NewCommandExecutor(client *dbgp.Client) *CommandExecutor {
	return &CommandExecutor{
		client: client,
	}
}

// ExecuteCommands executes a batch of commands and returns results
// This is thread-safe and can be called from multiple IPC requests
func (e *CommandExecutor) ExecuteCommands(commands []string, jsonOutput bool) []ipc.CommandResult {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.jsonOutput = jsonOutput
	results := make([]ipc.CommandResult, 0, len(commands))

	for _, cmdStr := range commands {
		// Parse command
		parts := strings.Fields(cmdStr)
		if len(parts) == 0 {
			continue
		}

		command := parts[0]
		args := parts[1:]

		// Execute command
		result := e.executeCommand(command, args)
		results = append(results, result)

		// If command failed or ended session, stop executing
		if !result.Success {
			break
		}

		// Check if session ended
		if e.isSessionEndingCommand(command) {
			break
		}
	}

	return results
}

// executeCommand executes a single command and returns the result
func (e *CommandExecutor) executeCommand(command string, args []string) ipc.CommandResult {
	switch command {
	case "run", "r":
		return e.handleRun()
	case "step", "s":
		return e.handleStep()
	case "next", "n":
		return e.handleNext()
	case "out", "o":
		return e.handleStepOut()
	case "break", "b":
		return e.handleBreak(args)
	case "print", "p":
		return e.handlePrint(args)
	case "context", "c":
		return e.handleContext(args)
	case "list", "l":
		return e.handleList()
	case "info", "i":
		return e.handleInfo(args)
	case "finish", "f":
		return e.handleFinish()
	case "help", "h", "?":
		return e.handleHelp(args)
	case "status", "st":
		return e.handleStatus()
	case "detach", "d":
		return e.handleDetach()
	case "eval", "e":
		return e.handleEval(args)
	case "set":
		return e.handleSet(args)
	case "source", "src":
		return e.handleSource(args)
	case "delete", "del":
		return e.handleDelete(args)
	case "disable":
		return e.handleDisable(args)
	case "enable":
		return e.handleEnable(args)
	case "stack":
		return e.handleStack()
	default:
		return ipc.CommandResult{
			Command: command,
			Success: false,
			Error:   fmt.Sprintf("Unknown command: %s", command),
		}
	}
}

// isSessionEndingCommand checks if a command ends the session
func (e *CommandExecutor) isSessionEndingCommand(command string) bool {
	return command == "finish" || command == "f" ||
	       command == "detach" || command == "d" ||
	       command == "quit" || command == "q"
}

// handleRun continues execution to next breakpoint
func (e *CommandExecutor) handleRun() ipc.CommandResult {
	response, err := e.client.Run()
	if err != nil {
		return ipc.CommandResult{
			Command: "run",
			Success: false,
			Error:   err.Error(),
		}
	}

	if response.HasError() {
		return ipc.CommandResult{
			Command: "run",
			Success: false,
			Error:   response.GetErrorMessage(),
		}
	}

	file, line := e.client.GetSession().GetCurrentLocation()
	return ipc.CommandResult{
		Command: "run",
		Success: true,
		Result: map[string]interface{}{
			"status":   response.Status,
			"filename": file,
			"line":     line,
		},
	}
}

// handleStep steps into next statement
func (e *CommandExecutor) handleStep() ipc.CommandResult {
	response, err := e.client.Step()
	if err != nil {
		return ipc.CommandResult{
			Command: "step",
			Success: false,
			Error:   err.Error(),
		}
	}

	if response.HasError() {
		return ipc.CommandResult{
			Command: "step",
			Success: false,
			Error:   response.GetErrorMessage(),
		}
	}

	file, line := e.client.GetSession().GetCurrentLocation()
	return ipc.CommandResult{
		Command: "step",
		Success: true,
		Result: map[string]interface{}{
			"status":   response.Status,
			"filename": file,
			"line":     line,
		},
	}
}

// handleNext steps over next statement
func (e *CommandExecutor) handleNext() ipc.CommandResult {
	response, err := e.client.Next()
	if err != nil {
		return ipc.CommandResult{
			Command: "next",
			Success: false,
			Error:   err.Error(),
		}
	}

	if response.HasError() {
		return ipc.CommandResult{
			Command: "next",
			Success: false,
			Error:   response.GetErrorMessage(),
		}
	}

	file, line := e.client.GetSession().GetCurrentLocation()
	return ipc.CommandResult{
		Command: "next",
		Success: true,
		Result: map[string]interface{}{
			"status":   response.Status,
			"filename": file,
			"line":     line,
		},
	}
}

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

	file, line := e.client.GetSession().GetCurrentLocation()
	return ipc.CommandResult{
		Command: "out",
		Success: true,
		Result: map[string]interface{}{
			"status":   response.Status,
			"filename": file,
			"line":     line,
		},
	}
}

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

// handleBreak sets breakpoints
func (e *CommandExecutor) handleBreak(args []string) ipc.CommandResult {
	if len(args) == 0 {
		return ipc.CommandResult{
			Command: "break",
			Success: false,
			Error:   "Usage: break <line> | break :<line> | break <file>:<line> | break call <function> | break exception",
		}
	}

	// Handle "break call <function>"
	if args[0] == "call" {
		if len(args) < 2 {
			return ipc.CommandResult{
				Command: "break",
				Success: false,
				Error:   "Usage: break call <function>",
			}
		}
		funcName := args[1]
		response, err := e.client.SetBreakpointToCall(funcName)
		if err != nil {
			return ipc.CommandResult{
				Command: "break",
				Success: false,
				Error:   err.Error(),
			}
		}
		if response.HasError() {
			return ipc.CommandResult{
				Command: "break",
				Success: false,
				Error:   response.GetErrorMessage(),
			}
		}
		return ipc.CommandResult{
			Command: "break",
			Success: true,
			Result: map[string]interface{}{
				"id":       response.ID,
				"location": fmt.Sprintf("call %s", funcName),
			},
		}
	}

	// Handle "break exception"
	if args[0] == "exception" {
		exceptionName := ""
		if len(args) > 1 {
			exceptionName = args[1]
		}
		response, err := e.client.SetExceptionBreakpoint(exceptionName)
		if err != nil {
			return ipc.CommandResult{
				Command: "break",
				Success: false,
				Error:   err.Error(),
			}
		}
		if response.HasError() {
			return ipc.CommandResult{
				Command: "break",
				Success: false,
				Error:   response.GetErrorMessage(),
			}
		}
		location := "exception"
		if exceptionName != "" {
			location = fmt.Sprintf("exception %s", exceptionName)
		}
		return ipc.CommandResult{
			Command: "break",
			Success: true,
			Result: map[string]interface{}{
				"id":       response.ID,
				"location": location,
			},
		}
	}

	// Parse locations and condition
	locations, condition, err := parseBreakpointArgs(args)
	if err != nil {
		return ipc.CommandResult{
			Command: "break",
			Success: false,
			Error:   err.Error(),
		}
	}

	// For now, handle only single location (multiple in next task)
	if len(locations) != 1 {
		return ipc.CommandResult{
			Command: "break",
			Success: false,
			Error:   "multiple breakpoints not yet supported",
		}
	}

	location := locations[0]

	// Parse line or file:line format
	var file string
	var line int

	// Check for :line format (current file)
	if strings.HasPrefix(location, ":") {
		lineStr := strings.TrimPrefix(location, ":")
		parsedLine, err := strconv.Atoi(lineStr)
		if err != nil {
			return ipc.CommandResult{
				Command: "break",
				Success: false,
				Error:   fmt.Sprintf("Invalid line number: %s", lineStr),
			}
		}
		line = parsedLine
		file, _ = e.client.GetSession().GetCurrentLocation()
		if file == "" {
			return ipc.CommandResult{
				Command: "break",
				Success: false,
				Error:   "No current file. Use format: break <file>:<line>",
			}
		}
	} else if strings.Contains(location, ":") {
		// file:line format
		parts := strings.SplitN(location, ":", 2)
		file = parts[0]
		parsedLine, err := strconv.Atoi(parts[1])
		if err != nil {
			return ipc.CommandResult{
				Command: "break",
				Success: false,
				Error:   fmt.Sprintf("Invalid line number: %s", parts[1]),
			}
		}
		line = parsedLine
	} else {
		// Just a line number
		parsedLine, err := strconv.Atoi(location)
		if err != nil {
			return ipc.CommandResult{
				Command: "break",
				Success: false,
				Error:   fmt.Sprintf("Invalid breakpoint format: %s", location),
			}
		}
		line = parsedLine
		file, _ = e.client.GetSession().GetCurrentLocation()
		if file == "" {
			return ipc.CommandResult{
				Command: "break",
				Success: false,
				Error:   "No current file. Use format: break <file>:<line>",
			}
		}
	}

	// Set the breakpoint with condition
	response, err := e.client.SetBreakpoint(file, line, condition)
	if err != nil {
		return ipc.CommandResult{
			Command: "break",
			Success: false,
			Error:   err.Error(),
		}
	}

	if response.HasError() {
		return ipc.CommandResult{
			Command: "break",
			Success: false,
			Error:   response.GetErrorMessage(),
		}
	}

	result := map[string]interface{}{
		"id":       response.ID,
		"location": fmt.Sprintf("%s:%d", file, line),
	}
	if condition != "" {
		result["condition"] = condition
	}

	return ipc.CommandResult{
		Command: "break",
		Success: true,
		Result:  result,
	}
}

// handlePrint prints variable value
func (e *CommandExecutor) handlePrint(args []string) ipc.CommandResult {
	if len(args) == 0 {
		return ipc.CommandResult{
			Command: "print",
			Success: false,
			Error:   "Usage: print <variable>",
		}
	}

	// Check if there are active stack frames before attempting to get property
	// When no frames exist (e.g., daemon waiting for PHP request), property_get fails with "stack depth invalid"
	stackResponse, err := e.client.GetStackTrace()
	if err != nil {
		return ipc.CommandResult{
			Command: "print",
			Success: false,
			Error:   fmt.Sprintf("Failed to check execution state: %v. Trigger a PHP request with XDEBUG_TRIGGER=1 and ensure execution hits a breakpoint first.", err),
		}
	}

	// Check if the stack is empty (no active execution context)
	// This happens when daemon is waiting for a PHP request or execution hasn't hit a breakpoint
	if len(stackResponse.Stack) == 0 {
		return ipc.CommandResult{
			Command: "print",
			Success: false,
			Error:   "Cannot inspect variables: no active stack frames. Trigger a PHP request with XDEBUG_TRIGGER=1 and ensure execution hits a breakpoint first.",
		}
	}

	varName := strings.Join(args, " ")
	varName = strings.TrimPrefix(varName, "$")

	response, err := e.client.GetProperty(varName)
	if err != nil {
		return ipc.CommandResult{
			Command: "print",
			Success: false,
			Error:   err.Error(),
		}
	}

	if response.HasError() {
		return ipc.CommandResult{
			Command: "print",
			Success: false,
			Error:   response.GetErrorMessage(),
		}
	}

	if len(response.Properties) == 0 {
		return ipc.CommandResult{
			Command: "print",
			Success: false,
			Error:   "Variable not found or has no value",
		}
	}

	prop := &response.Properties[0]
	jsonProp := view.ConvertPropertyToJSON(prop)

	return ipc.CommandResult{
		Command: "print",
		Success: true,
		Result:  jsonProp,
	}
}

// handleContext shows variables in current context
func (e *CommandExecutor) handleContext(args []string) ipc.CommandResult {
	contextType := "local"
	if len(args) > 0 {
		contextType = strings.ToLower(args[0])
	}

	var contextID int
	var scopeName string
	switch contextType {
	case "local":
		contextID = 0
		scopeName = "Local"
	case "global":
		contextID = 1
		scopeName = "Global"
	case "constant":
		contextID = 2
		scopeName = "Constant"
	default:
		return ipc.CommandResult{
			Command: "context",
			Success: false,
			Error:   fmt.Sprintf("Unknown context type: %s. Valid types: local, global, constant", contextType),
		}
	}

	// Check if there are active stack frames before attempting to get context
	// When no frames exist (e.g., daemon waiting for PHP request), context_get fails with "stack depth invalid"
	stackResponse, err := e.client.GetStackTrace()
	if err != nil {
		return ipc.CommandResult{
			Command: "context",
			Success: false,
			Error:   err.Error(),
		}
	}

	if stackResponse.HasError() {
		return ipc.CommandResult{
			Command: "context",
			Success: false,
			Error:   stackResponse.GetErrorMessage(),
		}
	}

	if len(stackResponse.Stack) == 0 {
		return ipc.CommandResult{
			Command: "context",
			Success: false,
			Error:   "Cannot inspect variables: no active stack frames. Trigger a PHP request with XDEBUG_TRIGGER=1 and ensure execution hits a breakpoint first.",
		}
	}

	response, err := e.client.GetContext(contextID)
	if err != nil {
		return ipc.CommandResult{
			Command: "context",
			Success: false,
			Error:   err.Error(),
		}
	}

	if response.HasError() {
		return ipc.CommandResult{
			Command: "context",
			Success: false,
			Error:   response.GetErrorMessage(),
		}
	}

	// Convert properties to JSON
	var viewProps []view.ProtocolProperty
	for i := range response.Properties {
		viewProps = append(viewProps, &response.Properties[i])
	}

	jsonProps := make([]view.JSONProperty, 0, len(viewProps))
	for _, prop := range viewProps {
		jsonProps = append(jsonProps, view.ConvertPropertyToJSON(prop))
	}

	return ipc.CommandResult{
		Command: "context",
		Success: true,
		Result: map[string]interface{}{
			"scope":     scopeName,
			"variables": jsonProps,
		},
	}
}

// handleList shows source code around current line
func (e *CommandExecutor) handleList() ipc.CommandResult {
	file, line := e.client.GetSession().GetCurrentLocation()
	if file == "" {
		return ipc.CommandResult{
			Command: "list",
			Success: false,
			Error:   "No current location available",
		}
	}

	return ipc.CommandResult{
		Command: "list",
		Success: true,
		Result: map[string]interface{}{
			"file": file,
			"line": line,
		},
	}
}

// handleInfo shows debugging information
func (e *CommandExecutor) handleInfo(args []string) ipc.CommandResult {
	if len(args) == 0 {
		return ipc.CommandResult{
			Command: "info",
			Success: false,
			Error:   "Usage: info <type>. Valid types: breakpoints (or 'b'), stack (or 's')",
		}
	}

	infoType := args[0]
	switch infoType {
	case "breakpoints", "b":
		response, err := e.client.GetBreakpointList()
		if err != nil {
			return ipc.CommandResult{
				Command: "info",
				Success: false,
				Error:   err.Error(),
			}
		}

		if response.HasError() {
			return ipc.CommandResult{
				Command: "info",
				Success: false,
				Error:   response.GetErrorMessage(),
			}
		}

		// Convert breakpoints to JSON
		var viewBps []view.ProtocolBreakpoint
		for i := range response.Breakpoints {
			viewBps = append(viewBps, &response.Breakpoints[i])
		}

		jsonBps := make([]view.JSONBreakpoint, 0, len(viewBps))
		for _, bp := range viewBps {
			jsonBps = append(jsonBps, view.ConvertBreakpointToJSON(bp))
		}

		return ipc.CommandResult{
			Command: "info",
			Success: true,
			Result: map[string]interface{}{
				"type":        "breakpoints",
				"breakpoints": jsonBps,
			},
		}

	case "stack", "s":
		response, err := e.client.GetStackTrace()
		if err != nil {
			return ipc.CommandResult{
				Command: "info",
				Success: false,
				Error:   err.Error(),
			}
		}

		if response.HasError() {
			return ipc.CommandResult{
				Command: "info",
				Success: false,
				Error:   response.GetErrorMessage(),
			}
		}

		// Convert stack to JSON
		var viewStack []view.ProtocolStack
		for i := range response.Stack {
			viewStack = append(viewStack, &response.Stack[i])
		}

		jsonStack := make([]view.JSONStack, 0, len(viewStack))
		for _, frame := range viewStack {
			jsonStack = append(jsonStack, view.ConvertStackToJSON(frame))
		}

		return ipc.CommandResult{
			Command: "info",
			Success: true,
			Result: map[string]interface{}{
				"type":   "stack",
				"frames": jsonStack,
			},
		}

	default:
		return ipc.CommandResult{
			Command: "info",
			Success: false,
			Error:   fmt.Sprintf("Unknown info type: %s. Valid types: breakpoints (or 'b'), stack (or 's')", infoType),
		}
	}
}

// handleFinish stops the debugging session
func (e *CommandExecutor) handleFinish() ipc.CommandResult {
	response, err := e.client.Finish()
	if err != nil {
		return ipc.CommandResult{
			Command: "finish",
			Success: false,
			Error:   err.Error(),
		}
	}

	if response.HasError() {
		return ipc.CommandResult{
			Command: "finish",
			Success: false,
			Error:   response.GetErrorMessage(),
		}
	}

	return ipc.CommandResult{
		Command: "finish",
		Success: true,
		Result: map[string]interface{}{
			"message": "Session stopped",
		},
	}
}

// handleHelp returns help information
func (e *CommandExecutor) handleHelp(args []string) ipc.CommandResult {
	helpText := fmt.Sprintf(`xdebug-cli version %s

Available commands:
  run, r              Continue execution
  step, s             Step into
  next, n             Step over
  break, b <target>   Set breakpoint
  print, p <var>      Print variable value
  context, c [type]   Show variables (local/global/constant)
  list, l             Show source code
  info, i [topic]     Show info (breakpoints)
  finish, f           Stop debugging
  help, h, ?          Show help

For detailed help on a specific command, use: help <command>
`, cfg.Version)

	if len(args) > 0 {
		// TODO: Add detailed help for specific commands
		helpText = fmt.Sprintf("Help for command: %s\n(Detailed help not yet implemented)", args[0])
	}

	return ipc.CommandResult{
		Command: "help",
		Success: true,
		Result: map[string]interface{}{
			"help": helpText,
		},
	}
}

// handleStatus returns the current execution status
func (e *CommandExecutor) handleStatus() ipc.CommandResult {
	response, err := e.client.Status()
	if err != nil {
		return ipc.CommandResult{
			Command: "status",
			Success: false,
			Error:   err.Error(),
		}
	}

	if response.HasError() {
		return ipc.CommandResult{
			Command: "status",
			Success: false,
			Error:   response.GetErrorMessage(),
		}
	}

	file, line := e.client.GetSession().GetCurrentLocation()
	return ipc.CommandResult{
		Command: "status",
		Success: true,
		Result: map[string]interface{}{
			"status":   response.Status,
			"reason":   response.Reason,
			"filename": file,
			"line":     line,
		},
	}
}

// handleDetach detaches from the debug session
func (e *CommandExecutor) handleDetach() ipc.CommandResult {
	response, err := e.client.Detach()
	if err != nil {
		return ipc.CommandResult{
			Command: "detach",
			Success: false,
			Error:   err.Error(),
		}
	}

	if response.HasError() {
		return ipc.CommandResult{
			Command: "detach",
			Success: false,
			Error:   response.GetErrorMessage(),
		}
	}

	return ipc.CommandResult{
		Command: "detach",
		Success: true,
	}
}

// handleEval evaluates a PHP expression
func (e *CommandExecutor) handleEval(args []string) ipc.CommandResult {
	if len(args) == 0 {
		return ipc.CommandResult{
			Command: "eval",
			Success: false,
			Error:   "Usage: eval <expression>",
		}
	}

	expression := strings.Join(args, " ")
	response, err := e.client.Eval(expression)
	if err != nil {
		return ipc.CommandResult{
			Command: "eval",
			Success: false,
			Error:   err.Error(),
		}
	}

	if response.HasError() {
		return ipc.CommandResult{
			Command: "eval",
			Success: false,
			Error:   response.GetErrorMessage(),
		}
	}

	if len(response.Properties) == 0 {
		return ipc.CommandResult{
			Command: "eval",
			Success: false,
			Error:   "No result returned",
		}
	}

	prop := &response.Properties[0]
	decodedValue, err := dbgp.DecodePropertyValue(prop)
	if err != nil {
		decodedValue = prop.Value
	}

	return ipc.CommandResult{
		Command: "eval",
		Success: true,
		Result: map[string]interface{}{
			"expression": expression,
			"type":       prop.Type,
			"value":      decodedValue,
		},
	}
}

// handleSet sets a variable value
func (e *CommandExecutor) handleSet(args []string) ipc.CommandResult {
	if len(args) < 3 {
		return ipc.CommandResult{
			Command: "set",
			Success: false,
			Error:   "Usage: set $variable = value",
		}
	}

	varName := args[0]
	if len(args) < 2 || args[1] != "=" {
		return ipc.CommandResult{
			Command: "set",
			Success: false,
			Error:   "Usage: set $variable = value",
		}
	}

	value := strings.Join(args[2:], " ")
	varName = strings.TrimPrefix(varName, "$")
	dataType := e.detectType(value)

	response, err := e.client.SetProperty(varName, value, dataType)
	if err != nil {
		return ipc.CommandResult{
			Command: "set",
			Success: false,
			Error:   err.Error(),
		}
	}

	if response.HasError() {
		return ipc.CommandResult{
			Command: "set",
			Success: false,
			Error:   response.GetErrorMessage(),
		}
	}

	return ipc.CommandResult{
		Command: "set",
		Success: true,
		Result: map[string]interface{}{
			"variable": "$" + varName,
			"value":    value,
			"type":     dataType,
		},
	}
}

// detectType detects the data type from a value string
func (e *CommandExecutor) detectType(value string) string {
	lower := strings.ToLower(value)
	if lower == "true" || lower == "false" {
		return "bool"
	}
	if _, err := strconv.Atoi(value); err == nil {
		return "int"
	}
	if _, err := strconv.ParseFloat(value, 64); err == nil {
		return "float"
	}
	return "string"
}

// handleSource retrieves source code
func (e *CommandExecutor) handleSource(args []string) ipc.CommandResult {
	var fileURI string
	var beginLine, endLine int

	if len(args) > 0 {
		arg := args[0]
		if strings.Contains(arg, ":") {
			parts := strings.Split(arg, ":")
			if len(parts) == 2 {
				if strings.Contains(parts[1], "-") {
					lineParts := strings.Split(parts[1], "-")
					beginLine, _ = strconv.Atoi(lineParts[0])
					if len(lineParts) > 1 {
						endLine, _ = strconv.Atoi(lineParts[1])
					}
				} else {
					beginLine, _ = strconv.Atoi(parts[1])
				}
				if parts[0] != "" {
					fileURI = parts[0]
				}
			}
		} else {
			fileURI = arg
		}
	}

	if fileURI == "" {
		file, _ := e.client.GetSession().GetCurrentLocation()
		if file == "" {
			return ipc.CommandResult{
				Command: "source",
				Success: false,
				Error:   "No file specified and no current location",
			}
		}
		fileURI = file
	}

	response, err := e.client.GetSource(fileURI, beginLine, endLine)
	if err != nil {
		return ipc.CommandResult{
			Command: "source",
			Success: false,
			Error:   err.Error(),
		}
	}

	if response.HasError() {
		return ipc.CommandResult{
			Command: "source",
			Success: false,
			Error:   response.GetErrorMessage(),
		}
	}

	sourceData := ""
	if response.Source != "" {
		sourceData = view.TryDecodeBase64(response.Source, "string")
	}

	return ipc.CommandResult{
		Command: "source",
		Success: true,
		Result: map[string]interface{}{
			"file":       fileURI,
			"start_line": beginLine,
			"end_line":   endLine,
			"source":     sourceData,
		},
	}
}

// handleDelete deletes a breakpoint
func (e *CommandExecutor) handleDelete(args []string) ipc.CommandResult {
	if len(args) == 0 {
		return ipc.CommandResult{
			Command: "delete",
			Success: false,
			Error:   "Usage: delete <breakpoint_id>",
		}
	}

	breakpointID := args[0]
	if _, err := strconv.Atoi(breakpointID); err != nil {
		return ipc.CommandResult{
			Command: "delete",
			Success: false,
			Error:   "Breakpoint ID must be numeric",
		}
	}

	response, err := e.client.RemoveBreakpoint(breakpointID)
	if err != nil {
		return ipc.CommandResult{
			Command: "delete",
			Success: false,
			Error:   err.Error(),
		}
	}

	if response.HasError() {
		return ipc.CommandResult{
			Command: "delete",
			Success: false,
			Error:   response.GetErrorMessage(),
		}
	}

	return ipc.CommandResult{
		Command: "delete",
		Success: true,
		Result: map[string]interface{}{
			"breakpoint_id": breakpointID,
		},
	}
}

// handleDisable disables a breakpoint
func (e *CommandExecutor) handleDisable(args []string) ipc.CommandResult {
	if len(args) == 0 {
		return ipc.CommandResult{
			Command: "disable",
			Success: false,
			Error:   "Usage: disable <breakpoint_id>",
		}
	}

	breakpointID := args[0]
	response, err := e.client.UpdateBreakpoint(breakpointID, "disabled")
	if err != nil {
		return ipc.CommandResult{
			Command: "disable",
			Success: false,
			Error:   err.Error(),
		}
	}

	if response.HasError() {
		return ipc.CommandResult{
			Command: "disable",
			Success: false,
			Error:   response.GetErrorMessage(),
		}
	}

	return ipc.CommandResult{
		Command: "disable",
		Success: true,
		Result: map[string]interface{}{
			"breakpoint_id": breakpointID,
			"state":         "disabled",
		},
	}
}

// handleEnable enables a breakpoint
func (e *CommandExecutor) handleEnable(args []string) ipc.CommandResult {
	if len(args) == 0 {
		return ipc.CommandResult{
			Command: "enable",
			Success: false,
			Error:   "Usage: enable <breakpoint_id>",
		}
	}

	breakpointID := args[0]
	response, err := e.client.UpdateBreakpoint(breakpointID, "enabled")
	if err != nil {
		return ipc.CommandResult{
			Command: "enable",
			Success: false,
			Error:   err.Error(),
		}
	}

	if response.HasError() {
		return ipc.CommandResult{
			Command: "enable",
			Success: false,
			Error:   response.GetErrorMessage(),
		}
	}

	return ipc.CommandResult{
		Command: "enable",
		Success: true,
		Result: map[string]interface{}{
			"breakpoint_id": breakpointID,
			"state":         "enabled",
		},
	}
}

// handleStack retrieves the call stack
func (e *CommandExecutor) handleStack() ipc.CommandResult {
	response, err := e.client.GetStackTrace()
	if err != nil {
		return ipc.CommandResult{
			Command: "stack",
			Success: false,
			Error:   err.Error(),
		}
	}

	if response.HasError() {
		return ipc.CommandResult{
			Command: "stack",
			Success: false,
			Error:   response.GetErrorMessage(),
		}
	}

	stackFrames := make([]map[string]interface{}, 0, len(response.Stack))
	for _, frame := range response.Stack {
		lineNo, _ := strconv.Atoi(frame.Lineno)
		stackFrames = append(stackFrames, map[string]interface{}{
			"depth":    frame.Level,
			"function": frame.Where,
			"file":     frame.Filename,
			"line":     lineNo,
		})
	}

	return ipc.CommandResult{
		Command: "stack",
		Success: true,
		Result:  stackFrames,
	}
}
