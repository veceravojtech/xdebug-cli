package cli

import (
	"context"
	"fmt"
	"os"
	"os/exec"

	"github.com/console/xdebug-cli/internal/mcpserver"

	"github.com/spf13/cobra"
)

var mcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "Start MCP server for AI assistant integration",
	Long: `Start an MCP (Model Context Protocol) server over stdio JSON-RPC.

This allows AI assistants like Claude Code and Claude Desktop to interact
with xdebug-cli debugging sessions programmatically through structured
tool calls.

The server exposes 6 MCP tools:
  xdebug_daemon_start    - Start a debug daemon session
  xdebug_daemon_kill     - Kill daemon session(s)
  xdebug_daemon_status   - Get daemon status
  xdebug_daemon_list     - List all active daemon sessions
  xdebug_daemon_is_alive - Check if daemon is running
  xdebug_execute         - Execute debug commands on a running session`,
	SilenceUsage:  true,
	SilenceErrors: true,
	Run: func(cmd *cobra.Command, args []string) {
		if err := runMCP(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(mcpCmd)
}

func runMCP() error {
	binary, err := resolveXdebugBinary()
	if err != nil {
		return err
	}
	srv := mcpserver.New(binary)
	return srv.Run(context.Background())
}

func resolveXdebugBinary() (string, error) {
	path, err := os.Executable()
	if err == nil {
		return path, nil
	}
	path, err = exec.LookPath("xdebug-cli")
	if err != nil {
		return "", fmt.Errorf("failed to resolve xdebug-cli binary: %w", err)
	}
	return path, nil
}
