package dbgp

import (
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
)

// Client represents a DBGp client for debugging operations
type Client struct {
	conn    *Connection
	session *Session
}

// NewClient creates a new DBGp client
func NewClient(conn *Connection) *Client {
	return &Client{
		conn:    conn,
		session: NewSession(),
	}
}

// Init reads the initial protocol message and sets up the session
func (c *Client) Init() (*ProtocolInit, error) {
	// Read the init message
	xmlData, err := c.conn.ReadMessage()
	if err != nil {
		return nil, fmt.Errorf("failed to read init message: %w", err)
	}

	// Parse the init message
	result, err := CreateProtocolFromXML(xmlData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse init message: %w", err)
	}

	init, ok := result.(*ProtocolInit)
	if !ok {
		return nil, fmt.Errorf("expected init message, got %T", result)
	}

	// Update session with init data
	c.session.SetState(StateStarting)
	c.session.SetIDEKey(init.IDEKey)
	c.session.SetAppID(init.AppID)
	c.session.SetTargetFiles(init.FileURI)

	return init, nil
}

// Run sends the run command to continue execution
func (c *Client) Run() (*ProtocolResponse, error) {
	txID := c.session.NextTransactionIDInt()
	command := fmt.Sprintf("run -i %d", txID)
	c.session.AddCommand(strconv.Itoa(txID), "run")

	err := c.conn.SendMessage(command)
	if err != nil {
		return nil, err
	}

	response, err := c.conn.GetResponse()
	if err != nil {
		return nil, err
	}

	// Update session state based on response
	c.updateSessionFromResponse(response)

	return response, nil
}

// Step sends the step_into command
func (c *Client) Step() (*ProtocolResponse, error) {
	txID := c.session.NextTransactionIDInt()
	command := fmt.Sprintf("step_into -i %d", txID)
	c.session.AddCommand(strconv.Itoa(txID), "step_into")

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

// Next sends the step_over command
func (c *Client) Next() (*ProtocolResponse, error) {
	txID := c.session.NextTransactionIDInt()
	command := fmt.Sprintf("step_over -i %d", txID)
	c.session.AddCommand(strconv.Itoa(txID), "step_over")

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

// Finish sends the stop command to end the debugging session
func (c *Client) Finish() (*ProtocolResponse, error) {
	txID := c.session.NextTransactionIDInt()
	command := fmt.Sprintf("stop -i %d", txID)
	c.session.AddCommand(strconv.Itoa(txID), "stop")
	c.session.SetState(StateStopping)

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

// Status sends the status command to get the current debug session status
func (c *Client) Status() (*ProtocolResponse, error) {
	txID := c.session.NextTransactionIDInt()
	command := fmt.Sprintf("status -i %d", txID)
	c.session.AddCommand(strconv.Itoa(txID), "status")

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

// Detach sends the detach command to detach from the debugging session
func (c *Client) Detach() (*ProtocolResponse, error) {
	txID := c.session.NextTransactionIDInt()
	command := fmt.Sprintf("detach -i %d", txID)
	c.session.AddCommand(strconv.Itoa(txID), "detach")
	c.session.SetState(StateStopping)

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

// SetBreakpoint sets a line breakpoint
func (c *Client) SetBreakpoint(file string, line int, condition string) (*ProtocolResponse, error) {
	txID := c.session.NextTransactionIDInt()

	// Convert file path to file:// URI if needed
	fileURI := c.normalizeFileURI(file)

	command := fmt.Sprintf("breakpoint_set -i %d -t line -f %s -n %d", txID, fileURI, line)

	if condition != "" {
		// Encode condition in base64
		encoded := base64.StdEncoding.EncodeToString([]byte(condition))
		command += fmt.Sprintf(" -- %s", encoded)
	}

	c.session.AddCommand(strconv.Itoa(txID), "breakpoint_set")

	err := c.conn.SendMessage(command)
	if err != nil {
		return nil, err
	}

	return c.conn.GetResponse()
}

// normalizeFileURI converts a file path to a file:// URI
// If the path is already a file:// URI, returns it as-is
// If it's an absolute path, converts to file:// URI
// If it's a relative path, resolves it against the project root from init FileURI
func (c *Client) normalizeFileURI(file string) string {
	// Already a file:// URI
	if strings.HasPrefix(file, "file://") {
		return file
	}

	// Absolute path - convert to file:// URI
	if strings.HasPrefix(file, "/") {
		return "file://" + file
	}

	// Relative path - need to resolve against project root
	// Get project root from init FileURI
	targetFiles := c.session.GetTargetFiles()
	if len(targetFiles) == 0 {
		// No init file, just add file:// prefix and hope for the best
		return "file:///" + file
	}

	// Extract directory path from file URI
	// e.g., file:///home/users/previo/current/booking/www/index.php
	// -> /home/users/previo/current/booking/www/
	projectRoot := extractProjectRoot(targetFiles[0])

	// Resolve relative path against project root
	absolutePath := resolveRelativePath(projectRoot, file)

	return "file://" + absolutePath
}

// extractProjectRoot extracts the project root directory from an init file:// URI
// Example: file:///home/users/previo/current/booking/www/index.php
// Returns: /home/users/previo/current/booking/
func extractProjectRoot(fileURI string) string {
	// Remove file:// prefix
	path := strings.TrimPrefix(fileURI, "file://")

	// Find the directory containing www/ or public/ or just use parent directory
	// Look for common web root directories
	if idx := strings.Index(path, "/www/"); idx != -1 {
		return path[:idx+1] // Include the slash before www
	}
	if idx := strings.Index(path, "/public/"); idx != -1 {
		return path[:idx+1]
	}

	// Fall back to just the directory of the init file
	lastSlash := strings.LastIndex(path, "/")
	if lastSlash > 0 {
		return path[:lastSlash+1]
	}

	return "/"
}

// resolveRelativePath resolves a relative file path against a project root
func resolveRelativePath(projectRoot string, relativePath string) string {
	if !strings.HasSuffix(projectRoot, "/") {
		projectRoot += "/"
	}

	// Check if the relative path starts with the project directory name
	// e.g., projectRoot="/home/users/previo/current/booking/"
	//       relativePath="booking/application/..."
	// This is redundant, so we should strip "booking/"
	parts := strings.Split(strings.Trim(relativePath, "/"), "/")
	if len(parts) > 0 {
		projectDirName := getLastDirName(projectRoot)
		if parts[0] == projectDirName {
			// Strip the redundant project directory name
			relativePath = strings.Join(parts[1:], "/")
		}
	}

	return projectRoot + relativePath
}

// getLastDirName extracts the last directory name from a path
// e.g., "/home/users/previo/current/booking/" -> "booking"
func getLastDirName(path string) string {
	path = strings.TrimSuffix(path, "/")
	lastSlash := strings.LastIndex(path, "/")
	if lastSlash >= 0 && lastSlash < len(path)-1 {
		return path[lastSlash+1:]
	}
	return path
}

// SetBreakpointToCall sets a function call breakpoint
func (c *Client) SetBreakpointToCall(funcName string) (*ProtocolResponse, error) {
	txID := c.session.NextTransactionIDInt()
	command := fmt.Sprintf("breakpoint_set -i %d -t call -m %s", txID, funcName)
	c.session.AddCommand(strconv.Itoa(txID), "breakpoint_set")

	err := c.conn.SendMessage(command)
	if err != nil {
		return nil, err
	}

	return c.conn.GetResponse()
}

// SetExceptionBreakpoint sets an exception breakpoint
func (c *Client) SetExceptionBreakpoint(exceptionName string) (*ProtocolResponse, error) {
	txID := c.session.NextTransactionIDInt()

	command := fmt.Sprintf("breakpoint_set -i %d -t exception", txID)
	if exceptionName != "" {
		command += fmt.Sprintf(" -x %s", exceptionName)
	}

	c.session.AddCommand(strconv.Itoa(txID), "breakpoint_set")

	err := c.conn.SendMessage(command)
	if err != nil {
		return nil, err
	}

	return c.conn.GetResponse()
}

// GetBreakpointList retrieves the list of all breakpoints
func (c *Client) GetBreakpointList() (*ProtocolResponse, error) {
	txID := c.session.NextTransactionIDInt()
	command := fmt.Sprintf("breakpoint_list -i %d", txID)
	c.session.AddCommand(strconv.Itoa(txID), "breakpoint_list")

	err := c.conn.SendMessage(command)
	if err != nil {
		return nil, err
	}

	return c.conn.GetResponse()
}

// RemoveBreakpoint removes a breakpoint by ID
func (c *Client) RemoveBreakpoint(breakpointID string) (*ProtocolResponse, error) {
	txID := c.session.NextTransactionIDInt()
	command := fmt.Sprintf("breakpoint_remove -i %d -d %s", txID, breakpointID)
	c.session.AddCommand(strconv.Itoa(txID), "breakpoint_remove")

	err := c.conn.SendMessage(command)
	if err != nil {
		return nil, err
	}

	return c.conn.GetResponse()
}

// UpdateBreakpoint updates a breakpoint state (enabled/disabled)
func (c *Client) UpdateBreakpoint(breakpointID, state string) (*ProtocolResponse, error) {
	// Validate state
	if state != "enabled" && state != "disabled" {
		return nil, fmt.Errorf("invalid breakpoint state: %s (must be 'enabled' or 'disabled')", state)
	}

	txID := c.session.NextTransactionIDInt()
	command := fmt.Sprintf("breakpoint_update -i %d -d %s -s %s", txID, breakpointID, state)
	c.session.AddCommand(strconv.Itoa(txID), "breakpoint_update")

	err := c.conn.SendMessage(command)
	if err != nil {
		return nil, err
	}

	return c.conn.GetResponse()
}

// GetProperty retrieves the value of a property/variable
func (c *Client) GetProperty(name string) (*ProtocolResponse, error) {
	txID := c.session.NextTransactionIDInt()
	command := fmt.Sprintf("property_get -i %d -d 0 -n %s", txID, name)
	c.session.AddCommand(strconv.Itoa(txID), "property_get")

	err := c.conn.SendMessage(command)
	if err != nil {
		return nil, err
	}

	return c.conn.GetResponse()
}

// SetProperty sets a variable value
func (c *Client) SetProperty(name, value, dataType string) (*ProtocolResponse, error) {
	txID := c.session.NextTransactionIDInt()

	// Base64-encode the value
	encodedValue := base64.StdEncoding.EncodeToString([]byte(value))
	dataLength := len(encodedValue)

	command := fmt.Sprintf("property_set -i %d -n %s -t %s -l %d -- %s",
		txID, name, dataType, dataLength, encodedValue)
	c.session.AddCommand(strconv.Itoa(txID), "property_set")

	err := c.conn.SendMessage(command)
	if err != nil {
		return nil, err
	}

	return c.conn.GetResponse()
}

// GetContext retrieves all variables in a specific context
func (c *Client) GetContext(contextID int) (*ProtocolResponse, error) {
	txID := c.session.NextTransactionIDInt()
	command := fmt.Sprintf("context_get -i %d -d 0 -c %d", txID, contextID)
	c.session.AddCommand(strconv.Itoa(txID), "context_get")

	err := c.conn.SendMessage(command)
	if err != nil {
		return nil, err
	}

	return c.conn.GetResponse()
}

// GetContextNames retrieves the list of available contexts
func (c *Client) GetContextNames() (*ProtocolResponse, error) {
	txID := c.session.NextTransactionIDInt()
	command := fmt.Sprintf("context_names -i %d", txID)
	c.session.AddCommand(strconv.Itoa(txID), "context_names")

	err := c.conn.SendMessage(command)
	if err != nil {
		return nil, err
	}

	return c.conn.GetResponse()
}

// Eval evaluates an expression
func (c *Client) Eval(expression string) (*ProtocolResponse, error) {
	txID := c.session.NextTransactionIDInt()

	// Encode expression in base64
	encoded := base64.StdEncoding.EncodeToString([]byte(expression))
	command := fmt.Sprintf("eval -i %d -- %s", txID, encoded)
	c.session.AddCommand(strconv.Itoa(txID), "eval")

	err := c.conn.SendMessage(command)
	if err != nil {
		return nil, err
	}

	return c.conn.GetResponse()
}

// GetStackDepth retrieves the current stack depth
func (c *Client) GetStackDepth() (*ProtocolResponse, error) {
	txID := c.session.NextTransactionIDInt()
	command := fmt.Sprintf("stack_depth -i %d", txID)
	c.session.AddCommand(strconv.Itoa(txID), "stack_depth")

	err := c.conn.SendMessage(command)
	if err != nil {
		return nil, err
	}

	return c.conn.GetResponse()
}

// GetStackTrace retrieves the call stack
func (c *Client) GetStackTrace() (*ProtocolResponse, error) {
	txID := c.session.NextTransactionIDInt()
	command := fmt.Sprintf("stack_get -i %d", txID)
	c.session.AddCommand(strconv.Itoa(txID), "stack_get")

	err := c.conn.SendMessage(command)
	if err != nil {
		return nil, err
	}

	return c.conn.GetResponse()
}

// GetSource retrieves the source code of a file with optional line range
// If fileURI is empty, the current file is used
// If beginLine is 0, source from the start
// If endLine is 0, source to the end
func (c *Client) GetSource(fileURI string, beginLine, endLine int) (*ProtocolResponse, error) {
	txID := c.session.NextTransactionIDInt()
	command := fmt.Sprintf("source -i %d", txID)

	// Add optional file URI parameter
	if fileURI != "" {
		command += fmt.Sprintf(" -f %s", fileURI)
	}

	// Add optional begin line parameter
	if beginLine > 0 {
		command += fmt.Sprintf(" -b %d", beginLine)
	}

	// Add optional end line parameter
	if endLine > 0 {
		command += fmt.Sprintf(" -e %d", endLine)
	}

	c.session.AddCommand(strconv.Itoa(txID), "source")

	err := c.conn.SendMessage(command)
	if err != nil {
		return nil, err
	}

	return c.conn.GetResponse()
}

// GetSession returns the session object
func (c *Client) GetSession() *Session {
	return c.session
}

// Close closes the connection
func (c *Client) Close() error {
	return c.conn.Close()
}

// updateSessionFromResponse updates session state based on response
func (c *Client) updateSessionFromResponse(response *ProtocolResponse) {
	// Update state based on status
	switch response.Status {
	case "starting":
		c.session.SetState(StateStarting)
	case "running":
		c.session.SetState(StateRunning)
	case "break":
		c.session.SetState(StateBreak)
	case "stopping":
		c.session.SetState(StateStopping)
	case "stopped":
		c.session.SetState(StateStopped)
	}

	// Update current location if provided
	if response.Filename != "" && response.Lineno != "" {
		line, _ := strconv.Atoi(response.Lineno)
		c.session.SetCurrentLocation(response.Filename, line)
	}

	// Handle message element for location (used in some responses)
	if len(response.Message) > 0 {
		msg := response.Message[0]
		if msg.Filename != "" && msg.Lineno != "" {
			line, _ := strconv.Atoi(msg.Lineno)
			c.session.SetCurrentLocation(msg.Filename, line)
		}
	}
}

// DecodePropertyValue decodes a base64-encoded property value
func DecodePropertyValue(prop *ProtocolProperty) (string, error) {
	if prop.Encoding == "base64" {
		decoded, err := base64.StdEncoding.DecodeString(strings.TrimSpace(prop.Value))
		if err != nil {
			return "", fmt.Errorf("failed to decode base64 value: %w", err)
		}
		return string(decoded), nil
	}
	return prop.Value, nil
}
