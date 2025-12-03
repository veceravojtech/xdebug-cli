package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/console/xdebug-cli/internal/daemon"
	"github.com/console/xdebug-cli/internal/dbgp"
	"github.com/spf13/cobra"
)

var daemonCmd = &cobra.Command{
	Use:   "daemon",
	Short: "Manage persistent debug daemon sessions",
	Long: `Manage background debug daemon sessions for multi-step debugging workflows.

The daemon command provides subcommands for starting and managing persistent
debug sessions that run in the background, allowing multiple attach invocations
without terminating the debug connection.

Available subcommands:
  start    Start a new daemon session

Example usage:
  xdebug-cli daemon start
  xdebug-cli daemon start --commands "break :42"
  xdebug-cli daemon start -p 9004`,
}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start a persistent debug daemon session",
	Long: `Start a background daemon that listens for Xdebug connections.

The daemon runs in the background and persists the debug session, allowing
multiple attach invocations to interact with the same session without
terminating the connection.

Features:
- Automatically kills any existing daemon on the same port (--force implicit)
- Listens on 0.0.0.0:9003 by default (all interfaces)
- Supports initial breakpoint/command setup via --commands flag
- Port can be changed with -p/--port flag

Example usage:
  xdebug-cli daemon start
  xdebug-cli daemon start --commands "break /path/file.php:100"
  xdebug-cli daemon start -p 9004
  xdebug-cli daemon start --json --commands "break :42"

After starting the daemon:
  1. Trigger PHP request with Xdebug enabled
  2. Use 'xdebug-cli attach --commands "..."' to interact with session
  3. Use 'xdebug-cli connection kill' to stop daemon`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := runDaemonStart(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	// Register start subcommand under daemon parent
	daemonCmd.AddCommand(startCmd)

	// Add flags to start subcommand
	startCmd.Flags().StringArrayVar(&CLIArgs.Commands, "commands", []string{}, "Commands to execute when connection established (optional)")

	// Register daemon parent command with root
	rootCmd.AddCommand(daemonCmd)
}

// runDaemonStart handles the daemon start command execution
func runDaemonStart() error {
	// Auto-force: always kill existing daemon on the same port
	killDaemonOnPort(CLIArgs.Port)

	// Create and start server
	server := dbgp.NewServer(CLIArgs.Host, CLIArgs.Port)
	if err := server.Listen(); err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}
	defer server.Close()

	// Start daemon mode
	return startDaemonMode(server)
}

// startDaemonMode handles daemon mode: fork if parent, or start daemon if child
func startDaemonMode(server *dbgp.Server) error {
	// Create daemon instance
	d, err := daemon.NewDaemon(server, CLIArgs.Port)
	if err != nil {
		return fmt.Errorf("failed to create daemon: %w", err)
	}

	// Check if we're already in daemon mode (child process)
	if daemon.IsDaemonMode() {
		// We're the child process, run the daemon
		return runDaemonProcess(d, server)
	}

	// We're the parent process, fork to background
	// Build command line args to pass to child
	args := os.Args

	// Fork the process
	if err := d.Fork(args); err != nil {
		return fmt.Errorf("failed to fork daemon: %w", err)
	}

	// Parent process exits successfully
	os.Exit(0)
	return nil
}

// runDaemonProcess runs the daemon logic in the child process
func runDaemonProcess(d *daemon.Daemon, server *dbgp.Server) error {
	// Initialize daemon infrastructure (PID, registry, IPC server) before waiting for connection
	// This allows 'attach' commands to connect even before the first Xdebug connection
	if err := d.Initialize(); err != nil {
		return fmt.Errorf("failed to initialize daemon: %w", err)
	}

	// Accept first connection (blocking)
	var daemonErr error
	err := server.Accept(func(conn *dbgp.Connection) {
		// Create client and initialize
		client := dbgp.NewClient(conn)
		_, err := client.Init()
		if err != nil {
			daemonErr = fmt.Errorf("failed to initialize session: %w", err)
			return
		}

		// Update global session state
		setActiveSession(client)
		defer clearActiveSession()

		// Set the client for the daemon (now that connection is established)
		d.SetClient(client)

		// Execute initial commands if provided
		if len(CLIArgs.Commands) > 0 {
			executor := daemon.NewCommandExecutor(client)

			// Check if any command sets a breakpoint
			hasBreakpoint := false
			hasRunCommand := false
			for _, cmd := range CLIArgs.Commands {
				if strings.HasPrefix(cmd, "break ") || strings.HasPrefix(cmd, "b ") {
					hasBreakpoint = true
				}
				if cmd == "run" || cmd == "r" {
					hasRunCommand = true
				}
			}

			// If breakpoint set without run, automatically add run command
			commandsToExecute := CLIArgs.Commands
			if hasBreakpoint && !hasRunCommand {
				commandsToExecute = append(commandsToExecute, "run")
			}

			results := executor.ExecuteCommands(commandsToExecute, CLIArgs.JSON)

			// Check for command failures
			for _, result := range results {
				if !result.Success {
					daemonErr = fmt.Errorf("command '%s' failed", result.Command)
					return
				}
			}
		}

		// Wait for daemon shutdown
		d.Wait()
	})

	if daemonErr != nil {
		return daemonErr
	}

	return err
}
