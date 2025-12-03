package cli

import (
	"encoding/base64"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/console/xdebug-cli/internal/daemon"
	"github.com/console/xdebug-cli/internal/dbgp"
	"github.com/console/xdebug-cli/internal/view"
	"github.com/spf13/cobra"
)

var listenCmd = &cobra.Command{
	Use:   "listen",
	Short: "Start DBGp server and execute debugging commands",
	Long: `Start a DBGp protocol server on the specified host and port.
The server will wait for Xdebug connections from PHP applications and
execute the specified debugging commands.

Commands are executed sequentially:
  xdebug-cli listen --commands "run" "step" "print $x"

For multi-step debugging workflows, use daemon mode:
  xdebug-cli daemon start --commands "break /path/file.php:100"
  xdebug-cli attach --commands "run"
  xdebug-cli attach --commands "context local"

Daemon mode runs in the background and persists the session, allowing multiple
attach invocations without terminating the debug connection.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Validate flags
		if err := validateListenFlags(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		if err := runListeningCmd(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	listenCmd.Flags().StringArrayVar(&CLIArgs.Commands, "commands", []string{}, "Commands to execute (required)")
	listenCmd.Flags().BoolVar(&CLIArgs.Force, "force", false, "Kill existing daemon on same port before starting")
	rootCmd.AddCommand(listenCmd)
}

// validateListenFlags validates command-line flag combinations
func validateListenFlags() error {
	// listen requires --commands
	if len(CLIArgs.Commands) == 0 {
		return fmt.Errorf(`listen command requires --commands flag

Example usage:
  xdebug-cli listen --commands "run" "print $var"

For multi-step debugging, use daemon mode:
  xdebug-cli daemon start
  xdebug-cli attach --commands "run"
  xdebug-cli attach --commands "print $var"`)
	}

	return nil
}

// killDaemonOnPort attempts to kill any daemon running on the specified port.
// Always returns nil (never fails) - shows warnings/errors but continues.
func killDaemonOnPort(port int) error {
	registry, err := daemon.NewSessionRegistry()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to load session registry: %v\n", err)
		return nil
	}

	session, err := registry.Get(port)
	if err != nil {
		// No session found - continue silently
		return nil
	}

	// Check if process exists (handle stale registry entries)
	if !processExists(session.PID) {
		fmt.Fprintf(os.Stderr, "Warning: daemon on port %d is stale (PID %d no longer exists), cleaning up\n", port, session.PID)
		registry.Remove(port)
		return nil
	}

	// Kill the process
	process, err := os.FindProcess(session.PID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to find daemon process (PID %d): %v\n", session.PID, err)
		return nil
	}

	if err := process.Kill(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to kill daemon on port %d (PID %d): %v\nContinuing anyway...\n", port, session.PID, err)
		return nil
	}

	// Clean up registry
	registry.Remove(port)
	fmt.Printf("Killed daemon on port %d (PID %d)\n", port, session.PID)
	return nil
}

// runListeningCmd starts the DBGp server and waits for connections
func runListeningCmd() error {
	v := view.NewView()

	// Kill existing daemon if --force is set
	if CLIArgs.Force {
		killDaemonOnPort(CLIArgs.Port)
	}

	// Create and start server
	server := dbgp.NewServer(CLIArgs.Host, CLIArgs.Port)
	if err := server.Listen(); err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}
	defer server.Close()

	// Command-based execution mode (with timeout to avoid hanging)
	timeout := 5 * time.Second
	return server.AcceptWithTimeout(timeout, func(conn *dbgp.Connection) {
		listenAccept(conn, v)
	})
}

// listenAccept handles an incoming connection and executes commands
func listenAccept(conn *dbgp.Connection, v *view.View) {
	defer conn.Close()

	// Create client and initialize
	client := dbgp.NewClient(conn)
	_, err := client.Init()
	if err != nil {
		v.PrintErrorLn(fmt.Sprintf("Failed to initialize session: %v", err))
		os.Exit(1)
		return
	}

	// Update global session state
	setActiveSession(client)
	defer clearActiveSession()

	// Execute commands
	exitCode := executeCommands(client, v)
	os.Exit(exitCode)
}

// executeCommands executes commands sequentially
// Returns exit code: 0 for success, 1+ for errors
func executeCommands(client *dbgp.Client, v *view.View) int {
	// If no commands specified, just exit successfully
	if len(CLIArgs.Commands) == 0 {
		return 0
	}

	for i, cmdStr := range CLIArgs.Commands {
		// Parse command and arguments
		parts := strings.Fields(cmdStr)
		if len(parts) == 0 {
			continue
		}

		command := parts[0]
		args := parts[1:]

		// Execute command
		shouldExit := dispatchCommand(client, v, command, args)

		// Check if command caused session to end
		if shouldExit {
			// If this was intended (quit, finish), return success
			// If there are more commands, return error
			if i < len(CLIArgs.Commands)-1 {
				if CLIArgs.JSON {
					v.OutputJSON(command, false, "Session ended before all commands completed", nil)
				} else {
					v.PrintErrorLn("Session ended before all commands completed")
				}
				return 1
			}
			return 0
		}
	}

	return 0
}

// dispatchCommand routes commands to their handlers
// Returns true if the session should end
func dispatchCommand(client *dbgp.Client, v *view.View, command string, args []string) bool {
	switch command {
	case "run", "r":
		return handleRun(client, v)
	case "step", "s":
		return handleStep(client, v)
	case "next", "n":
		return handleNext(client, v)
	case "out", "o":
		return handleStepOut(client, v)
	case "break", "b":
		handleBreak(client, v, args)
	case "delete", "del":
		handleDelete(client, v, args)
	case "disable", "d":
		handleDisable(client, v, args)
	case "enable":
		handleEnable(client, v, args)
	case "print", "p":
		handlePrint(client, v, args)
	case "eval", "e":
		handleEval(client, v, args)
	case "context", "c":
		handleContext(client, v, args)
	case "list", "l":
		handleList(client, v)
	case "source", "src":
		handleSource(client, v, args)
	case "info", "i":
		handleInfo(client, v, args)
	case "stack":
		handleStack(client, v)
	case "finish", "f":
		return handleFinish(client, v)
	case "status", "st":
		return handleStatus(client, v)
	case "help", "h", "?":
		handleHelp(v, args)
	default:
		v.PrintErrorLn(fmt.Sprintf("Unknown command: %s", command))
		v.PrintLn("Use --help flag for available commands.")
		v.PrintErrorLn(fmt.Sprintf("Unknown command: %s", command))
		v.PrintLn("Use --help flag for available commands.")
	}
	return false
}

// handleRun continues execution to next breakpoint
func handleRun(client *dbgp.Client, v *view.View) bool {
	if !CLIArgs.JSON {
		v.PrintLn("Continuing execution...")
	}
	response, err := client.Run()
	if err != nil {
		if CLIArgs.JSON {
			v.OutputJSON("run", false, err.Error(), nil)
		} else {
			v.PrintErrorLn(fmt.Sprintf("Error: %v", err))
		}
		return false
	}

	return updateState(client, v, response, "run")
}

// handleStep steps into next statement
func handleStep(client *dbgp.Client, v *view.View) bool {
	response, err := client.Step()
	if err != nil {
		if CLIArgs.JSON {
			v.OutputJSON("step", false, err.Error(), nil)
		} else {
			v.PrintErrorLn(fmt.Sprintf("Error: %v", err))
		}
		return false
	}

	return updateState(client, v, response, "step")
}

// handleNext steps over next statement
func handleNext(client *dbgp.Client, v *view.View) bool {
	response, err := client.Next()
	if err != nil {
		if CLIArgs.JSON {
			v.OutputJSON("next", false, err.Error(), nil)
		} else {
			v.PrintErrorLn(fmt.Sprintf("Error: %v", err))
		}
		return false
	}

	return updateState(client, v, response, "next")
}

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
func handleBreak(client *dbgp.Client, v *view.View, args []string) {
	if len(args) == 0 {
		if CLIArgs.JSON {
			v.OutputJSON("break", false, "Usage: break <line> | break :<line> | break <file>:<line> | break call <function> | break exception", nil)
		} else {
			v.PrintErrorLn("Usage: break <line> | break :<line> | break <file>:<line> | break call <function> | break exception")
			v.PrintLn("Type 'help break' for more information.")
		}
		return
	}

	// Handle "break call <function>"
	if args[0] == "call" {
		if len(args) < 2 {
			if CLIArgs.JSON {
				v.OutputJSON("break", false, "Usage: break call <function>", nil)
			} else {
				v.PrintErrorLn("Usage: break call <function>")
			}
			return
		}
		funcName := args[1]
		response, err := client.SetBreakpointToCall(funcName)
		if err != nil {
			if CLIArgs.JSON {
				v.OutputJSON("break", false, fmt.Sprintf("Error setting breakpoint: %v", err), nil)
			} else {
				v.PrintErrorLn(fmt.Sprintf("Error setting breakpoint: %v", err))
			}
			return
		}
		if response.HasError() {
			if CLIArgs.JSON {
				v.OutputJSON("break", false, response.GetErrorMessage(), nil)
			} else {
				v.PrintErrorLn(fmt.Sprintf("Error: %s", response.GetErrorMessage()))
			}
			return
		}
		if CLIArgs.JSON {
			result := view.JSONBreakpointResult{
				ID:       response.ID,
				Location: fmt.Sprintf("call %s", funcName),
			}
			v.OutputJSON("break", true, "", result)
		} else {
			v.PrintLn(fmt.Sprintf("Breakpoint set on function call: %s (ID: %s)", funcName, response.ID))
		}
		return
	}

	// Handle "break exception"
	if args[0] == "exception" {
		exceptionName := ""
		if len(args) > 1 {
			exceptionName = args[1]
		}
		response, err := client.SetExceptionBreakpoint(exceptionName)
		if err != nil {
			if CLIArgs.JSON {
				v.OutputJSON("break", false, fmt.Sprintf("Error setting breakpoint: %v", err), nil)
			} else {
				v.PrintErrorLn(fmt.Sprintf("Error setting breakpoint: %v", err))
			}
			return
		}
		if response.HasError() {
			if CLIArgs.JSON {
				v.OutputJSON("break", false, response.GetErrorMessage(), nil)
			} else {
				v.PrintErrorLn(fmt.Sprintf("Error: %s", response.GetErrorMessage()))
			}
			return
		}
		location := "exception"
		if exceptionName != "" {
			location = fmt.Sprintf("exception %s", exceptionName)
		}
		if CLIArgs.JSON {
			result := view.JSONBreakpointResult{
				ID:       response.ID,
				Location: location,
			}
			v.OutputJSON("break", true, "", result)
		} else {
			if exceptionName != "" {
				v.PrintLn(fmt.Sprintf("Breakpoint set on exception: %s (ID: %s)", exceptionName, response.ID))
			} else {
				v.PrintLn(fmt.Sprintf("Breakpoint set on all exceptions (ID: %s)", response.ID))
			}
		}
		return
	}

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

	type breakpointResult struct {
		ID       string
		Location string
		Error    string
	}

	var results []breakpointResult

	for _, location := range locations {
		var file string
		var line int

		// Parse location format
		if strings.HasPrefix(location, ":") {
			// :line format (current file)
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
			// file:line format
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
			// Just a line number
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

		// Set the breakpoint with condition
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
}

// handleDelete removes a breakpoint by ID
func handleDelete(client *dbgp.Client, v *view.View, args []string) {
	if len(args) == 0 {
		if CLIArgs.JSON {
			v.OutputJSON("delete", false, "Usage: delete <breakpoint_id>", nil)
		} else {
			v.PrintErrorLn("Usage: delete <breakpoint_id>")
		}
		return
	}

	breakpointID := args[0]

	// Validate ID is numeric
	if _, err := strconv.Atoi(breakpointID); err != nil {
		if CLIArgs.JSON {
			v.OutputJSON("delete", false, "Breakpoint ID must be numeric", nil)
		} else {
			v.PrintErrorLn("Breakpoint ID must be numeric")
		}
		return
	}

	response, err := client.RemoveBreakpoint(breakpointID)
	if err != nil {
		if CLIArgs.JSON {
			v.OutputJSON("delete", false, err.Error(), nil)
		} else {
			v.PrintErrorLn(fmt.Sprintf("Error: %v", err))
		}
		return
	}

	if response.HasError() {
		if CLIArgs.JSON {
			v.OutputJSON("delete", false, response.GetErrorMessage(), nil)
		} else {
			v.PrintErrorLn(fmt.Sprintf("Error: %s", response.GetErrorMessage()))
		}
		return
	}

	if CLIArgs.JSON {
		result := map[string]string{
			"breakpoint_id": breakpointID,
		}
		v.OutputJSON("delete", true, "", result)
	} else {
		v.PrintLn(fmt.Sprintf("Deleted breakpoint %s", breakpointID))
	}
}

// handleDisable disables a breakpoint
func handleDisable(client *dbgp.Client, v *view.View, args []string) {
	if len(args) == 0 {
		if CLIArgs.JSON {
			v.OutputJSON("disable", false, "Usage: disable <breakpoint_id>", nil)
		} else {
			v.PrintErrorLn("Usage: disable <breakpoint_id>")
		}
		return
	}

	breakpointID := args[0]

	response, err := client.UpdateBreakpoint(breakpointID, "disabled")
	if err != nil {
		if CLIArgs.JSON {
			v.OutputJSON("disable", false, err.Error(), nil)
		} else {
			v.PrintErrorLn(fmt.Sprintf("Error: %v", err))
		}
		return
	}

	if response.HasError() {
		if CLIArgs.JSON {
			v.OutputJSON("disable", false, response.GetErrorMessage(), nil)
		} else {
			v.PrintErrorLn(fmt.Sprintf("Error: %s", response.GetErrorMessage()))
		}
		return
	}

	if CLIArgs.JSON {
		result := map[string]string{
			"breakpoint_id": breakpointID,
			"state":         "disabled",
		}
		v.OutputJSON("disable", true, "", result)
	} else {
		v.PrintLn(fmt.Sprintf("Disabled breakpoint %s", breakpointID))
	}
}

// handleEnable enables a breakpoint
func handleEnable(client *dbgp.Client, v *view.View, args []string) {
	if len(args) == 0 {
		if CLIArgs.JSON {
			v.OutputJSON("enable", false, "Usage: enable <breakpoint_id>", nil)
		} else {
			v.PrintErrorLn("Usage: enable <breakpoint_id>")
		}
		return
	}

	breakpointID := args[0]

	response, err := client.UpdateBreakpoint(breakpointID, "enabled")
	if err != nil {
		if CLIArgs.JSON {
			v.OutputJSON("enable", false, err.Error(), nil)
		} else {
			v.PrintErrorLn(fmt.Sprintf("Error: %v", err))
		}
		return
	}

	if response.HasError() {
		if CLIArgs.JSON {
			v.OutputJSON("enable", false, response.GetErrorMessage(), nil)
		} else {
			v.PrintErrorLn(fmt.Sprintf("Error: %s", response.GetErrorMessage()))
		}
		return
	}

	if CLIArgs.JSON {
		result := map[string]string{
			"breakpoint_id": breakpointID,
			"state":         "enabled",
		}
		v.OutputJSON("enable", true, "", result)
	} else {
		v.PrintLn(fmt.Sprintf("Enabled breakpoint %s", breakpointID))
	}
}

// handlePrint prints variable value
func handlePrint(client *dbgp.Client, v *view.View, args []string) {
	if len(args) == 0 {
		if CLIArgs.JSON {
			v.OutputJSON("print", false, "Usage: print <variable>", nil)
		} else {
			v.PrintErrorLn("Usage: print <variable>")
			v.PrintLn("Type 'help print' for more information.")
		}
		return
	}

	varName := strings.Join(args, " ")
	// Remove $ prefix if present (PHP variables)
	varName = strings.TrimPrefix(varName, "$")

	response, err := client.GetProperty(varName)
	if err != nil {
		if CLIArgs.JSON {
			v.OutputJSON("print", false, err.Error(), nil)
		} else {
			v.PrintErrorLn(fmt.Sprintf("Error: %v", err))
		}
		return
	}

	if response.HasError() {
		if CLIArgs.JSON {
			v.OutputJSON("print", false, response.GetErrorMessage(), nil)
		} else {
			v.PrintErrorLn(fmt.Sprintf("Error: %s", response.GetErrorMessage()))
		}
		return
	}

	if len(response.Properties) == 0 {
		if CLIArgs.JSON {
			v.OutputJSON("print", false, "Variable not found or has no value", nil)
		} else {
			v.PrintLn("Variable not found or has no value.")
		}
		return
	}

	// Display the property
	prop := &response.Properties[0]
	if CLIArgs.JSON {
		jsonProp := view.ConvertPropertyToJSON(prop)
		v.OutputJSON("print", true, "", jsonProp)
	} else {
		v.PrintProperty(prop)
	}
}

// handleEval evaluates PHP expressions
func handleEval(client *dbgp.Client, v *view.View, args []string) {
	if len(args) == 0 {
		if CLIArgs.JSON {
			v.OutputJSON("eval", false, "Usage: eval <expression>", nil)
		} else {
			v.PrintErrorLn("Usage: eval <expression>")
			v.PrintLn("Type 'help eval' for more information.")
		}
		return
	}

	expression := strings.Join(args, " ")

	response, err := client.Eval(expression)
	if err != nil {
		if CLIArgs.JSON {
			v.OutputJSON("eval", false, err.Error(), nil)
		} else {
			v.PrintErrorLn(fmt.Sprintf("Error: %v", err))
		}
		return
	}

	if response.HasError() {
		if CLIArgs.JSON {
			v.OutputJSON("eval", false, response.GetErrorMessage(), nil)
		} else {
			v.PrintErrorLn(fmt.Sprintf("Error: %s", response.GetErrorMessage()))
		}
		return
	}

	if len(response.Properties) == 0 {
		if CLIArgs.JSON {
			v.OutputJSON("eval", false, "No result returned", nil)
		} else {
			v.PrintLn("No result returned.")
		}
		return
	}

	prop := response.Properties[0]
	decodedValue, err := dbgp.DecodePropertyValue(&prop)
	if err != nil {
		decodedValue = prop.Value
	}

	if CLIArgs.JSON {
		result := map[string]string{
			"expression": expression,
			"type":       prop.Type,
			"value":      decodedValue,
		}
		v.OutputJSON("eval", true, "", result)
	} else {
		v.PrintLn(fmt.Sprintf("%s = %s (%s)", expression, decodedValue, prop.Type))
	}
}

// handleContext shows variables in current context
func handleContext(client *dbgp.Client, v *view.View, args []string) {
	contextType := "local"
	if len(args) > 0 {
		contextType = strings.ToLower(args[0])
	}

	// Map context types to context IDs
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
		if CLIArgs.JSON {
			v.OutputJSON("context", false, fmt.Sprintf("Unknown context type: %s. Valid types: local, global, constant", contextType), nil)
		} else {
			v.PrintErrorLn(fmt.Sprintf("Unknown context type: %s", contextType))
			v.PrintLn("Valid types: local, global, constant")
		}
		return
	}

	response, err := client.GetContext(contextID)
	if err != nil {
		if CLIArgs.JSON {
			v.OutputJSON("context", false, err.Error(), nil)
		} else {
			v.PrintErrorLn(fmt.Sprintf("Error: %v", err))
		}
		return
	}

	if response.HasError() {
		if CLIArgs.JSON {
			v.OutputJSON("context", false, response.GetErrorMessage(), nil)
		} else {
			v.PrintErrorLn(fmt.Sprintf("Error: %s", response.GetErrorMessage()))
		}
		return
	}

	// Convert to view types
	var viewProps []view.ProtocolProperty
	for i := range response.Properties {
		viewProps = append(viewProps, &response.Properties[i])
	}

	if CLIArgs.JSON {
		// Convert all properties to JSON
		jsonProps := make([]view.JSONProperty, 0, len(viewProps))
		for _, prop := range viewProps {
			jsonProps = append(jsonProps, view.ConvertPropertyToJSON(prop))
		}
		result := map[string]interface{}{
			"scope":     scopeName,
			"variables": jsonProps,
		}
		v.OutputJSON("context", true, "", result)
	} else {
		v.PrintPropertyListWithDetails(scopeName, viewProps)
	}
}

// handleList shows source code around current line
func handleList(client *dbgp.Client, v *view.View) {
	file, line := client.GetSession().GetCurrentLocation()
	if file == "" {
		v.PrintErrorLn("No current location available.")
		return
	}

	v.PrintSourceLn(file, line, 5)
}

// handleInfo shows debugging information

// handleSource shows source code with optional file and line range
func handleSource(client *dbgp.Client, v *view.View, args []string) {
	var fileURI string
	var beginLine, endLine int

	// Parse args: "source", "source :10-20", "source file.php:10-20"
	if len(args) > 0 {
		arg := args[0]

		// Check if it contains line range
		if strings.Contains(arg, ":") {
			parts := strings.Split(arg, ":")
			if len(parts) == 2 {
				// Parse line range (e.g., "10-20" or just "10")
				if strings.Contains(parts[1], "-") {
					lineParts := strings.Split(parts[1], "-")
					beginLine, _ = strconv.Atoi(lineParts[0])
					if len(lineParts) > 1 {
						endLine, _ = strconv.Atoi(lineParts[1])
					}
				} else {
					beginLine, _ = strconv.Atoi(parts[1])
				}

				// parts[0] is either empty (":10-20") or file path
				if parts[0] != "" {
					fileURI = parts[0]
				}
			}
		} else {
			// Just a file path, no line range
			fileURI = arg
		}
	}

	// If no file specified, use current location
	if fileURI == "" {
		file, _ := client.GetSession().GetCurrentLocation()
		if file == "" {
			if CLIArgs.JSON {
				v.OutputJSON("source", false, "No file specified and no current location", nil)
			} else {
				v.PrintErrorLn("No file specified and no current location")
			}
			return
		}
		fileURI = file
	}

	response, err := client.GetSource(fileURI, beginLine, endLine)
	if err != nil {
		if CLIArgs.JSON {
			v.OutputJSON("source", false, err.Error(), nil)
		} else {
			v.PrintErrorLn(fmt.Sprintf("Error: %v", err))
		}
		return
	}

	if response.HasError() {
		if CLIArgs.JSON {
			v.OutputJSON("source", false, response.GetErrorMessage(), nil)
		} else {
			v.PrintErrorLn(fmt.Sprintf("Error: %s", response.GetErrorMessage()))
		}
		return
	}

	// Decode base64 source
	sourceData := ""
	if response.Source != "" {
		decoded, err := base64.StdEncoding.DecodeString(strings.TrimSpace(response.Source))
		if err == nil {
			sourceData = string(decoded)
		}
	}

	if CLIArgs.JSON {
		result := map[string]interface{}{
			"file":       fileURI,
			"start_line": beginLine,
			"end_line":   endLine,
			"source":     sourceData,
		}
		v.OutputJSON("source", true, "", result)
	} else {
		// Display source with line numbers (similar to list command)
		lines := strings.Split(sourceData, "\n")
		startNum := beginLine
		if startNum == 0 {
			startNum = 1
		}
		for i, line := range lines {
			lineNum := startNum + i
			fmt.Printf("%4d: %s\n", lineNum, line)
		}
	}
}

func handleInfo(client *dbgp.Client, v *view.View, args []string) {
	if len(args) == 0 {
		if CLIArgs.JSON {
			v.OutputJSON("info", false, "Usage: info <type>. Valid types: breakpoints (or 'b'), stack (or 's')", nil)
		} else {
			v.PrintErrorLn("Usage: info <type>")
			v.PrintLn("Type 'help info' for more information.")
		}
		return
	}

	infoType := args[0]
	switch infoType {
	case "breakpoints", "b":
		response, err := client.GetBreakpointList()
		if err != nil {
			if CLIArgs.JSON {
				v.OutputJSON("info", false, err.Error(), nil)
			} else {
				v.PrintErrorLn(fmt.Sprintf("Error: %v", err))
			}
			return
		}

		if response.HasError() {
			if CLIArgs.JSON {
				v.OutputJSON("info", false, response.GetErrorMessage(), nil)
			} else {
				v.PrintErrorLn(fmt.Sprintf("Error: %s", response.GetErrorMessage()))
			}
			return
		}

		// Convert to view types
		var viewBps []view.ProtocolBreakpoint
		for i := range response.Breakpoints {
			viewBps = append(viewBps, &response.Breakpoints[i])
		}

		if CLIArgs.JSON {
			// Convert breakpoints to JSON
			jsonBps := make([]view.JSONBreakpoint, 0, len(viewBps))
			for _, bp := range viewBps {
				jsonBps = append(jsonBps, view.ConvertBreakpointToJSON(bp))
			}
			result := map[string]interface{}{
				"type":        "breakpoints",
				"breakpoints": jsonBps,
			}
			v.OutputJSON("info", true, "", result)
		} else {
			v.ShowInfoBreakpoints(viewBps)
		}

	case "stack", "s":
		response, err := client.GetStackTrace()
		if err != nil {
			if CLIArgs.JSON {
				v.OutputJSON("info", false, err.Error(), nil)
			} else {
				v.PrintErrorLn(fmt.Sprintf("Error: %v", err))
			}
			return
		}

		if response.HasError() {
			if CLIArgs.JSON {
				v.OutputJSON("info", false, response.GetErrorMessage(), nil)
			} else {
				v.PrintErrorLn(fmt.Sprintf("Error: %s", response.GetErrorMessage()))
			}
			return
		}

		// Convert to view types
		var viewStack []view.ProtocolStack
		for i := range response.Stack {
			viewStack = append(viewStack, &response.Stack[i])
		}

		if CLIArgs.JSON {
			// Convert stack to JSON
			jsonStack := make([]view.JSONStack, 0, len(viewStack))
			for _, frame := range viewStack {
				jsonStack = append(jsonStack, view.ConvertStackToJSON(frame))
			}
			result := map[string]interface{}{
				"type":   "stack",
				"frames": jsonStack,
			}
			v.OutputJSON("info", true, "", result)
		} else {
			v.ShowInfoStack(viewStack)
		}

	default:
		if CLIArgs.JSON {
			v.OutputJSON("info", false, fmt.Sprintf("Unknown info type: %s. Valid types: breakpoints (or 'b'), stack (or 's')", infoType), nil)
		} else {
			v.PrintErrorLn(fmt.Sprintf("Unknown info type: %s", infoType))
			v.PrintLn("Valid types: breakpoints (or 'b'), stack (or 's')")
		}
	}
}

// handleStack displays the stack trace
func handleStack(client *dbgp.Client, v *view.View) {
	response, err := client.GetStackTrace()
	if err != nil {
		if CLIArgs.JSON {
			v.OutputJSON("stack", false, err.Error(), nil)
		} else {
			v.PrintErrorLn(fmt.Sprintf("Error: %v", err))
		}
		return
	}

	if response.HasError() {
		if CLIArgs.JSON {
			v.OutputJSON("stack", false, response.GetErrorMessage(), nil)
		} else {
			v.PrintErrorLn(fmt.Sprintf("Error: %s", response.GetErrorMessage()))
		}
		return
	}

	// Convert to view types
	var viewStack []view.ProtocolStack
	for i := range response.Stack {
		viewStack = append(viewStack, &response.Stack[i])
	}

	if CLIArgs.JSON {
		// Convert stack to JSON
		jsonStack := make([]view.JSONStack, 0, len(viewStack))
		for _, frame := range viewStack {
			jsonStack = append(jsonStack, view.ConvertStackToJSON(frame))
		}
		result := map[string]interface{}{
			"frames": jsonStack,
		}
		v.OutputJSON("stack", true, "", result)
	} else {
		v.ShowInfoStack(viewStack)
	}
}

// handleFinish stops the debugging session
func handleFinish(client *dbgp.Client, v *view.View) bool {
	v.PrintLn("Stopping session...")
	response, err := client.Finish()
	if err != nil {
		v.PrintErrorLn(fmt.Sprintf("Error: %v", err))
		return false
	}

	if response.HasError() {
		v.PrintErrorLn(fmt.Sprintf("Error: %s", response.GetErrorMessage()))
	}

	v.PrintLn("Session stopped.")
	return true
}

// handleStatus returns the current execution status
func handleStatus(client *dbgp.Client, v *view.View) bool {
	response, err := client.Status()
	if err != nil {
		if CLIArgs.JSON {
			v.OutputJSON("status", false, err.Error(), nil)
		} else {
			v.PrintErrorLn(fmt.Sprintf("Error: %v", err))
		}
		return false
	}

	return updateState(client, v, response, "status")
}

// handleHelp shows help messages
func handleHelp(v *view.View, args []string) {
	if len(args) == 0 {
		v.ShowHelpMessage()
		return
	}

	// Show help for specific command
	v.ShowCommandHelp(args[0])
}

// updateState updates the display based on the response status
// Returns true if the session should end
func updateState(client *dbgp.Client, v *view.View, response *dbgp.ProtocolResponse, command string) bool {
	if response.HasError() {
		if CLIArgs.JSON {
			v.OutputJSON(command, false, response.GetErrorMessage(), nil)
		} else {
			v.PrintErrorLn(fmt.Sprintf("Error: %s", response.GetErrorMessage()))
		}
		return false
	}

	session := client.GetSession()
	state := session.GetState()
	file, line := session.GetCurrentLocation()

	if CLIArgs.JSON {
		result := view.JSONStateResult{
			Status:   response.Status,
			Filename: file,
			Line:     line,
		}
		v.OutputJSON(command, true, "", result)
	} else {
		switch response.Status {
		case "break":
			v.PrintLn(fmt.Sprintf("Breakpoint hit at %s:%d", file, line))
			v.PrintSourceLn(file, line, 5)

		case "stopping", "stopped":
			v.PrintLn("Execution finished.")
			return true

		case "running":
			v.PrintLn("Execution completed.")
			return true

		default:
			v.PrintLn(fmt.Sprintf("Status: %s", state.String()))
		}
	}

	// Check if session should end
	if response.Status == "stopping" || response.Status == "stopped" || response.Status == "running" {
		return true
	}

	return false
}
