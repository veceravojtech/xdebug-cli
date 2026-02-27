package mcpserver

import (
	"context"
	"os/exec"
	"strconv"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// run executes the xdebug-cli binary with the given arguments and returns
// combined stdout+stderr output. On non-zero exit the error is returned
// alongside the output so the caller can decide whether it is IsError.
func run(ctx context.Context, binary string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, binary, args...)
	out, err := cmd.CombinedOutput()
	return strings.TrimSpace(string(out)), err
}

func textResult(output string) (*mcp.CallToolResult, error) {
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: output}},
	}, nil
}

func errorResult(output string) (*mcp.CallToolResult, error) {
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: output}},
		IsError: true,
	}, nil
}

// --- Input structs ---

// DaemonStartInput defines parameters for xdebug_daemon_start.
type DaemonStartInput struct {
	Curl                     string   `json:"curl,omitempty" jsonschema:"Curl arguments to trigger Xdebug connection"`
	Port                     int      `json:"port,omitempty" jsonschema:"Port number for Xdebug connection (default 9003)"`
	Commands                 []string `json:"commands,omitempty" jsonschema:"Commands to execute when connection established"`
	EnableExternalConnection bool     `json:"enable_external_connection,omitempty" jsonschema:"Wait for external Xdebug trigger instead of using curl"`
	BreakpointTimeout        int      `json:"breakpoint_timeout,omitempty" jsonschema:"Timeout in seconds for breakpoint validation (default 30)"`
	WaitForever              bool     `json:"wait_forever,omitempty" jsonschema:"Disable breakpoint timeout (wait indefinitely)"`
}

// DaemonKillInput defines parameters for xdebug_daemon_kill.
type DaemonKillInput struct {
	Port  int  `json:"port,omitempty" jsonschema:"Port number (default 9003)"`
	All   bool `json:"all,omitempty" jsonschema:"Kill all daemon sessions"`
	Force bool `json:"force,omitempty" jsonschema:"Skip confirmation prompt (always true when all=true)"`
}

// DaemonStatusInput defines parameters for xdebug_daemon_status.
type DaemonStatusInput struct {
	Port int `json:"port,omitempty" jsonschema:"Port number (default 9003)"`
}

// DaemonListInput has no parameters (always returns JSON).
type DaemonListInput struct{}

// DaemonIsAliveInput defines parameters for xdebug_daemon_is_alive.
type DaemonIsAliveInput struct {
	Port int `json:"port,omitempty" jsonschema:"Port number (default 9003)"`
}

// ExecuteInput defines parameters for xdebug_execute.
type ExecuteInput struct {
	Commands []string `json:"commands" jsonschema:"Debug commands to execute"`
	Port     int      `json:"port,omitempty" jsonschema:"Port number (default 9003)"`
}

// --- Arg builders (pure functions) ---

func buildDaemonStartArgs(input DaemonStartInput) []string {
	args := []string{"daemon", "start"}
	if input.Curl != "" {
		args = append(args, "--curl", input.Curl)
	}
	if input.EnableExternalConnection {
		args = append(args, "--enable-external-connection")
	}
	if input.Port != 0 {
		args = append(args, "-p", strconv.Itoa(input.Port))
	}
	for _, cmd := range input.Commands {
		args = append(args, "--commands", cmd)
	}
	if input.BreakpointTimeout != 0 {
		args = append(args, "--breakpoint-timeout", strconv.Itoa(input.BreakpointTimeout))
	}
	if input.WaitForever {
		args = append(args, "--wait-forever")
	}
	return args
}

func buildDaemonKillArgs(input DaemonKillInput) []string {
	args := []string{"daemon", "kill"}
	if input.Port != 0 {
		args = append(args, "-p", strconv.Itoa(input.Port))
	}
	if input.All {
		args = append(args, "--all", "--force")
	} else if input.Force {
		args = append(args, "--force")
	}
	return args
}

func buildDaemonStatusArgs(input DaemonStatusInput) []string {
	args := []string{"daemon", "status"}
	if input.Port != 0 {
		args = append(args, "-p", strconv.Itoa(input.Port))
	}
	return args
}

func buildDaemonListArgs() []string {
	return []string{"daemon", "list", "--json"}
}

func buildDaemonIsAliveArgs(input DaemonIsAliveInput) []string {
	args := []string{"daemon", "isAlive"}
	if input.Port != 0 {
		args = append(args, "-p", strconv.Itoa(input.Port))
	}
	return args
}

func buildExecuteArgs(input ExecuteInput) []string {
	args := []string{"attach", "--json"}
	if input.Port != 0 {
		args = append(args, "-p", strconv.Itoa(input.Port))
	}
	for _, cmd := range input.Commands {
		args = append(args, "--commands", cmd)
	}
	return args
}

// --- Tool registrations and handlers ---

func (s *Server) registerDaemonStart() {
	mcp.AddTool(s.server, &mcp.Tool{
		Name:        "xdebug_daemon_start",
		Description: "Start a debug daemon session. Either curl or enable_external_connection is required.",
	}, s.handleDaemonStart)
}

func (s *Server) handleDaemonStart(ctx context.Context, _ *mcp.CallToolRequest, input DaemonStartInput) (*mcp.CallToolResult, any, error) {
	args := buildDaemonStartArgs(input)
	output, err := run(ctx, s.binary, args...)
	if err != nil {
		r, _ := errorResult(output)
		return r, nil, nil
	}
	r, _ := textResult(output)
	return r, nil, nil
}

func (s *Server) registerDaemonKill() {
	mcp.AddTool(s.server, &mcp.Tool{
		Name:        "xdebug_daemon_kill",
		Description: "Kill daemon session(s). Use all=true to kill all sessions.",
	}, s.handleDaemonKill)
}

func (s *Server) handleDaemonKill(ctx context.Context, _ *mcp.CallToolRequest, input DaemonKillInput) (*mcp.CallToolResult, any, error) {
	args := buildDaemonKillArgs(input)
	output, err := run(ctx, s.binary, args...)
	if err != nil {
		r, _ := errorResult(output)
		return r, nil, nil
	}
	r, _ := textResult(output)
	return r, nil, nil
}

func (s *Server) registerDaemonStatus() {
	mcp.AddTool(s.server, &mcp.Tool{
		Name:        "xdebug_daemon_status",
		Description: "Get status of the daemon on the specified port.",
	}, s.handleDaemonStatus)
}

func (s *Server) handleDaemonStatus(ctx context.Context, _ *mcp.CallToolRequest, input DaemonStatusInput) (*mcp.CallToolResult, any, error) {
	args := buildDaemonStatusArgs(input)
	output, err := run(ctx, s.binary, args...)
	if err != nil {
		r, _ := errorResult(output)
		return r, nil, nil
	}
	r, _ := textResult(output)
	return r, nil, nil
}

func (s *Server) registerDaemonList() {
	mcp.AddTool(s.server, &mcp.Tool{
		Name:        "xdebug_daemon_list",
		Description: "List all active daemon sessions with their PID, port, and socket path.",
	}, s.handleDaemonList)
}

func (s *Server) handleDaemonList(ctx context.Context, _ *mcp.CallToolRequest, _ DaemonListInput) (*mcp.CallToolResult, any, error) {
	args := buildDaemonListArgs()
	output, err := run(ctx, s.binary, args...)
	if err != nil {
		r, _ := errorResult(output)
		return r, nil, nil
	}
	r, _ := textResult(output)
	return r, nil, nil
}

func (s *Server) registerDaemonIsAlive() {
	mcp.AddTool(s.server, &mcp.Tool{
		Name:        "xdebug_daemon_is_alive",
		Description: "Check if a daemon is running on the specified port. Returns 'connected' or 'not connected'.",
	}, s.handleDaemonIsAlive)
}

func (s *Server) handleDaemonIsAlive(ctx context.Context, _ *mcp.CallToolRequest, input DaemonIsAliveInput) (*mcp.CallToolResult, any, error) {
	args := buildDaemonIsAliveArgs(input)
	output, err := run(ctx, s.binary, args...)
	if err != nil {
		// Non-zero exit for isAlive is informational (not alive), not an error.
		r, _ := textResult(output)
		return r, nil, nil
	}
	r, _ := textResult(output)
	return r, nil, nil
}

func (s *Server) registerExecute() {
	mcp.AddTool(s.server, &mcp.Tool{
		Name:        "xdebug_execute",
		Description: "Execute debug commands on a running daemon session. Commands: run, step, next, out, break, print, context, eval, list, stack, status, info, delete, clear, disable, enable, finish, detach, help.",
	}, s.handleExecute)
}

func (s *Server) handleExecute(ctx context.Context, _ *mcp.CallToolRequest, input ExecuteInput) (*mcp.CallToolResult, any, error) {
	args := buildExecuteArgs(input)
	output, err := run(ctx, s.binary, args...)
	if err != nil {
		r, _ := errorResult(output)
		return r, nil, nil
	}
	r, _ := textResult(output)
	return r, nil, nil
}
