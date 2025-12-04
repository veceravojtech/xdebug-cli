package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/console/xdebug-cli/internal/daemon"
	"github.com/console/xdebug-cli/internal/dbgp"
	"github.com/console/xdebug-cli/internal/ipc"
	"github.com/spf13/cobra"
)

var daemonCmd = &cobra.Command{
	Use:   "daemon",
	Short: "Start and manage debug daemon sessions",
	Long: `Start and manage background debug daemon sessions.

All debugging sessions in xdebug-cli run as background daemons. The daemon
persists your debug session, allowing you to execute multiple commands
across separate CLI invocations without losing the debug connection.

Available subcommands:
  start     Start a new daemon session
  status    Show current daemon status
  list      List all active daemon sessions
  kill      Terminate daemon session(s)
  isAlive   Check if daemon is running

Example usage:
  xdebug-cli daemon start
  xdebug-cli daemon start --commands "break :42"
  xdebug-cli daemon start -p 9004
  xdebug-cli daemon status
  xdebug-cli daemon list --json
  xdebug-cli daemon kill
  xdebug-cli daemon kill --all --force`,
}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start a debug daemon session",
	Long: `Start a background daemon that listens for Xdebug connections.

This is the primary entry point for all debugging sessions. The daemon runs
in the background and keeps your debug session alive, allowing you to execute
multiple commands via 'attach' without losing the connection.

REQUIRED (one of):
  --curl                       Curl arguments to trigger Xdebug connection
  --enable-external-connection Wait for external Xdebug trigger (browser, IDE, manual)

Features:
- Automatically kills any existing daemon on the same port
- Listens on 0.0.0.0:9003 by default (all interfaces)
- Supports initial breakpoint/command setup via --commands flag
- Port can be changed with -p/--port flag
- Auto-appends XDEBUG_TRIGGER cookie to curl command (when using --curl)

Breakpoint timeout options:
- Default 30-second timeout handles slow PHP bootstrap (opcache, frameworks)
- Use --wait-forever for cold starts or when breakpoint timing is unpredictable
- Use --breakpoint-timeout N to set a custom timeout in seconds (0 = disabled)

Example workflow with curl trigger:
  # 1. Start daemon with curl trigger
  xdebug-cli daemon start --curl "http://localhost/app.php"

  # 2. Interact with the session (daemon triggered PHP automatically)
  xdebug-cli attach --commands "context local"
  xdebug-cli attach --commands "print \$myVar"
  xdebug-cli attach --commands "run"

  # 3. Stop the daemon when done
  xdebug-cli daemon kill

Example workflow with external trigger:
  # 1. Start daemon waiting for external connection (--commands required)
  xdebug-cli daemon start --enable-external-connection --commands "break /app/file.php:42"

  # 2. Trigger PHP from browser/IDE/manual curl with XDEBUG_TRIGGER

  # 3. Interact with the session
  xdebug-cli attach --commands "context local"
  xdebug-cli attach --commands "run"

  # 4. Stop the daemon when done
  xdebug-cli daemon kill

Additional examples:
  xdebug-cli daemon start --curl "http://localhost/app.php"
  xdebug-cli daemon start --curl "http://localhost/app.php" -p 9004
  xdebug-cli daemon start --curl "http://localhost/api -X POST -d 'data'" --commands "break :42"
  xdebug-cli daemon start --enable-external-connection --commands "break /app/file.php:42"
  xdebug-cli daemon start --enable-external-connection -p 9004 --commands "break :100"
  xdebug-cli daemon start --curl "http://localhost/app.php" --wait-forever --commands "break :42"`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := runDaemonStart(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show current daemon status",
	Long: `Display information about the active daemon session on the current port.

Shows PID, port, socket path, and start time for daemon sessions.
For daemon sessions, provides information about the background process.`,
	Run: func(cmd *cobra.Command, args []string) {
		runDaemonStatus()
	},
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all active daemon sessions",
	Long: `List all active daemon debugging sessions across all ports.

Shows PID, port, socket path, and start time for each session.
Supports JSON output for automation with --json flag.

Example usage:
  xdebug-cli daemon list
  xdebug-cli daemon list --json`,
	Run: func(cmd *cobra.Command, args []string) {
		runDaemonList()
	},
}

var killCmd = &cobra.Command{
	Use:   "kill",
	Short: "Terminate daemon session(s)",
	Long: `Terminate the currently active daemon session and shut down the daemon.

Sends a kill request via IPC and shuts down the daemon process.
Use --all flag to terminate all daemon sessions at once.
Use --force to skip confirmation prompts when using --all.

Example usage:
  xdebug-cli daemon kill
  xdebug-cli daemon kill --all
  xdebug-cli daemon kill --all --force`,
	Run: func(cmd *cobra.Command, args []string) {
		runDaemonKill()
	},
}

var isAliveCmd = &cobra.Command{
	Use:   "isAlive",
	Short: "Check if daemon is running",
	Long: `Check if there is an active daemon process on the current port.

Exits with code 0 if daemon is running, code 1 if not running.
Useful for shell scripts and automation.

Example usage:
  xdebug-cli daemon isAlive
  if xdebug-cli daemon isAlive; then echo "Daemon running"; fi`,
	Run: func(cmd *cobra.Command, args []string) {
		runDaemonIsAlive()
	},
}

func init() {
	// Register subcommands under daemon parent
	daemonCmd.AddCommand(startCmd)
	daemonCmd.AddCommand(statusCmd)
	daemonCmd.AddCommand(listCmd)
	daemonCmd.AddCommand(killCmd)
	daemonCmd.AddCommand(isAliveCmd)

	// Add flags to start subcommand
	startCmd.Flags().StringVar(&CLIArgs.Curl, "curl", "", "Curl arguments to trigger Xdebug connection")
	startCmd.Flags().BoolVar(&CLIArgs.EnableExternalConnection, "enable-external-connection", false, "Wait for external Xdebug connection (bypasses --curl requirement)")
	startCmd.Flags().StringArrayVar(&CLIArgs.Commands, "commands", []string{}, "Commands to execute when connection established (optional)")
	startCmd.Flags().IntVar(&CLIArgs.BreakpointTimeout, "breakpoint-timeout", 30, "Timeout in seconds to wait for breakpoint hit (0 = disabled, default handles slow bootstrap)")
	startCmd.Flags().BoolVar(&CLIArgs.WaitForever, "wait-forever", false, "Disable breakpoint timeout (wait indefinitely, useful for cold starts)")

	// Add flags to list subcommand
	listCmd.Flags().BoolVar(&CLIArgs.JSON, "json", false, "Output in JSON format")

	// Add flags to kill subcommand
	killCmd.Flags().BoolVar(&CLIArgs.KillAll, "all", false, "Kill all daemon sessions")
	killCmd.Flags().BoolVar(&CLIArgs.Force, "force", false, "Skip confirmation prompt")

	// Register daemon parent command with root
	rootCmd.AddCommand(daemonCmd)
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

// killAllDaemonsSilent kills all active daemon sessions silently.
// Used by daemon start to ensure clean state before starting.
func killAllDaemonsSilent() {
	registry, err := daemon.NewSessionRegistry()
	if err != nil {
		return
	}

	sessions := registry.List()
	for _, session := range sessions {
		if !processExists(session.PID) {
			// Stale entry, just remove from registry
			registry.Remove(session.Port)
			continue
		}

		// Try IPC kill first
		client := ipc.NewClient(session.SocketPath)
		resp, err := client.Kill()
		if err != nil || !resp.Success {
			// IPC failed, try direct process kill
			if process, err := os.FindProcess(session.PID); err == nil {
				process.Kill()
			}
		}

		// Wait for process to actually terminate (up to 2 seconds)
		for i := 0; i < 20; i++ {
			if !processExists(session.PID) {
				break
			}
			time.Sleep(100 * time.Millisecond)
		}

		// Clean up registry
		registry.Remove(session.Port)
	}
}

// runDaemonStart handles the daemon start command execution
func runDaemonStart() error {
	// Apply --wait-forever flag (sets breakpoint timeout to 0)
	if CLIArgs.WaitForever {
		CLIArgs.BreakpointTimeout = 0
	}

	// Check if we're already in daemon mode (child process)
	// If so, run the daemon directly - don't do parent-only validation
	if daemon.IsDaemonMode() {
		// Create and start server (child process owns the server)
		server := dbgp.NewServer(CLIArgs.Host, CLIArgs.Port)
		if err := server.Listen(); err != nil {
			return fmt.Errorf("failed to start server: %w", err)
		}
		defer server.Close()

		d, err := daemon.NewDaemon(server, CLIArgs.Port)
		if err != nil {
			return fmt.Errorf("failed to create daemon: %w", err)
		}

		return runDaemonProcess(d, server)
	}

	// Parent process - do validation and fork

	// Validate that either --curl or --enable-external-connection is provided
	if CLIArgs.Curl == "" && !CLIArgs.EnableExternalConnection {
		return fmt.Errorf(`either --curl or --enable-external-connection is required

Usage:
  xdebug-cli daemon start --curl "<curl-args>"
  xdebug-cli daemon start --enable-external-connection --commands "break :42"

Examples:
  xdebug-cli daemon start --curl "http://localhost/app.php"
  xdebug-cli daemon start --curl "http://localhost/api -X POST -d 'data'"
  xdebug-cli daemon start --enable-external-connection --commands "break /app/file.php:42"

Use --curl to trigger Xdebug via HTTP request (XDEBUG_TRIGGER cookie added automatically).
Use --enable-external-connection to wait for external triggers (browser, IDE, manual).`)
	}

	// Verify curl binary exists in PATH (only if --curl is used)
	if CLIArgs.Curl != "" {
		if _, err := exec.LookPath("curl"); err != nil {
			return fmt.Errorf("curl not found in PATH")
		}
	}

	// Clean up stale registry entries (crashed/killed daemons)
	registry, err := daemon.NewSessionRegistry()
	if err == nil {
		registry.CleanupStaleEntries()
	}

	// Auto-kill all existing daemons before starting
	killAllDaemonsSilent()

	// Fork the daemon (child will create and start server)
	return forkDaemon()
}

// forkDaemon handles the parent process forking and waiting for daemon status
func forkDaemon() error {
	// Create a temporary daemon instance just for forking (no server needed)
	d, err := daemon.NewDaemon(nil, CLIArgs.Port)
	if err != nil {
		return fmt.Errorf("failed to create daemon: %w", err)
	}

	// Check for non-absolute breakpoint paths and show warning BEFORE forking
	// (child process has stderr redirected to log file)
	hasNonAbsolute, nonAbsPath := daemon.HasNonAbsoluteBreakpoint(CLIArgs.Commands)
	if hasNonAbsolute {
		pathStore, err := daemon.NewBreakpointPathStore()
		var suggestedPath string
		if err == nil {
			// Extract just the filename part for lookup
			filename := nonAbsPath
			if strings.Contains(nonAbsPath, ":") {
				filename = strings.Split(nonAbsPath, ":")[0]
			}
			suggestedPath = pathStore.LoadBreakpointPath(filename)
		}

		// Show warning
		fmt.Fprintf(os.Stderr, "Warning: breakpoint path '%s' is not absolute.\n", nonAbsPath)
		if suggestedPath != "" {
			// Extract line number from original path
			lineNum := ""
			if strings.Contains(nonAbsPath, ":") {
				lineNum = ":" + strings.Split(nonAbsPath, ":")[1]
			}
			fmt.Fprintf(os.Stderr, "Suggestion: use '%s%s' instead.\n", suggestedPath, lineNum)
		}
		fmt.Fprintf(os.Stderr, "Xdebug requires absolute paths for breakpoints.\n")
		if CLIArgs.BreakpointTimeout > 0 {
			fmt.Fprintf(os.Stderr, "Will wait %d seconds for breakpoint hit, then fail if not hit.\n", CLIArgs.BreakpointTimeout)
		}
		fmt.Fprintln(os.Stderr, "")
	}

	// Clean up any old status file before forking
	daemon.CleanupStatusFile(CLIArgs.Port)

	// Check if we have breakpoint commands that need validation
	hasBreakpointCommand := false
	for _, cmd := range CLIArgs.Commands {
		if strings.HasPrefix(cmd, "break ") || strings.HasPrefix(cmd, "b ") {
			hasBreakpointCommand = true
			break
		}
	}

	// Fork the process
	args := os.Args
	if err := d.Fork(args); err != nil {
		return fmt.Errorf("failed to fork daemon: %w", err)
	}

	// If we have breakpoint commands and a timeout, wait for daemon to report status
	if hasBreakpointCommand && CLIArgs.BreakpointTimeout > 0 {
		// Wait for status file with timeout
		timeout := time.Duration(CLIArgs.BreakpointTimeout+5) * time.Second // Extra 5s for daemon overhead
		deadline := time.Now().Add(timeout)

		for time.Now().Before(deadline) {
			status, exists, err := daemon.ReadStatus(CLIArgs.Port)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error reading daemon status: %v\n", err)
				os.Exit(1)
			}

			if exists {
				daemon.CleanupStatusFile(CLIArgs.Port)
				if strings.HasPrefix(status, "ready:") {
					// Breakpoint hit successfully - show location
					location := strings.TrimPrefix(status, "ready:")
					fmt.Printf("Breakpoint hit at %s\n", location)
					os.Exit(0)
				} else if status == "ready" {
					// Legacy ready without location
					fmt.Println("Breakpoint hit")
					os.Exit(0)
				} else if strings.HasPrefix(status, "error:") {
					// Error occurred
					errorMsg := strings.TrimPrefix(status, "error:")
					fmt.Fprintf(os.Stderr, "Error: %s\n", errorMsg)
					os.Exit(1)
				}
			}

			time.Sleep(100 * time.Millisecond)
		}

		// Timeout waiting for daemon status
		fmt.Fprintf(os.Stderr, "Timeout waiting for daemon status\n")
		os.Exit(124)
	}

	// No breakpoint validation needed, parent exits successfully
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

	// Check for non-absolute breakpoint paths (warning already shown in parent process)
	hasNonAbsolute, nonAbsPath := daemon.HasNonAbsoluteBreakpoint(CLIArgs.Commands)
	var pathStore *daemon.BreakpointPathStore
	var suggestedPath string

	if hasNonAbsolute {
		var err error
		pathStore, err = daemon.NewBreakpointPathStore()
		if err == nil {
			// Extract just the filename part for lookup
			filename := nonAbsPath
			if strings.Contains(nonAbsPath, ":") {
				filename = strings.Split(nonAbsPath, ":")[0]
			}
			suggestedPath = pathStore.LoadBreakpointPath(filename)
		}
	}

	// Execute curl to trigger Xdebug connection (CLIArgs.Curl is passed via command line)
	var curlErrCh <-chan error
	if CLIArgs.Curl != "" {
		curlErrCh = executeCurl(CLIArgs.Curl)

		// Monitor curl for errors in background - terminate daemon if curl fails
		go func() {
			if err := <-curlErrCh; err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\nDaemon terminated.\n", err)
				d.Shutdown()
				os.Exit(1)
			}
		}()
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

			// Check if any command sets a breakpoint and collect breakpoint locations
			hasBreakpoint := false
			hasRunCommand := false
			var breakpointLocations []string
			for _, cmd := range CLIArgs.Commands {
				if strings.HasPrefix(cmd, "break ") || strings.HasPrefix(cmd, "b ") {
					hasBreakpoint = true
					// Extract the location from the command
					parts := strings.Fields(cmd)
					if len(parts) >= 2 {
						breakpointLocations = append(breakpointLocations, parts[1])
					}
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

			// Set up breakpoint validation timeout for ALL breakpoints (not just non-absolute)
			var timeoutCh <-chan time.Time
			if hasBreakpoint && CLIArgs.BreakpointTimeout > 0 {
				timeoutCh = time.After(time.Duration(CLIArgs.BreakpointTimeout) * time.Second)
			}

			results := executor.ExecuteCommands(commandsToExecute, CLIArgs.JSON)

			// Check for command failures
			for _, result := range results {
				if !result.Success {
					daemonErr = fmt.Errorf("command '%s' failed", result.Command)
					return
				}
			}

			// After run command, check if we hit a breakpoint (validate for ALL breakpoints)
			if hasBreakpoint && CLIArgs.BreakpointTimeout > 0 {
				// Check the status - if we're in "break" status, the breakpoint was hit
				statusResp, err := client.Status()
				if err != nil {
					daemonErr = fmt.Errorf("failed to get status: %w", err)
					return
				}

				// Build breakpoint location string for error messages
				breakpointStr := strings.Join(breakpointLocations, ", ")

				if statusResp.Status == "break" {
					// Breakpoint was hit! Save the full path for future suggestions
					currentFile, currentLine := client.GetSession().GetCurrentLocation()
					if currentFile != "" && pathStore != nil {
						pathStore.SaveBreakpointPath(currentFile)
					}
					// Signal success to parent process with location
					d.WriteStatus(fmt.Sprintf("ready:%s:%d", currentFile, currentLine))
				} else if statusResp.Status == "stopping" || statusResp.Status == "stopped" {
					// Script ended without hitting breakpoint - this is the fail-fast case
					errorMsg := fmt.Sprintf("Breakpoint at '%s' was not hit - script completed.", breakpointStr)
					if hasNonAbsolute {
						if suggestedPath != "" {
							lineNum := ""
							if strings.Contains(nonAbsPath, ":") {
								lineNum = ":" + strings.Split(nonAbsPath, ":")[1]
							}
							errorMsg += fmt.Sprintf(" Use full path: %s%s", suggestedPath, lineNum)
						} else {
							errorMsg += " Ensure you use an absolute path (starting with /)."
						}
					} else {
						errorMsg += " Verify the breakpoint location is correct and the code path is executed."
					}
					// Signal error to parent process
					d.WriteStatus("error:" + errorMsg)
					d.Shutdown()
					os.Exit(1)
				} else {
					// Still running - wait for timeout or breakpoint hit
					select {
					case <-timeoutCh:
						// Timeout expired - check status one more time
						statusResp, err := client.Status()
						if err != nil || statusResp.Status != "break" {
							// Build timeout error message
							errorMsg := fmt.Sprintf("Breakpoint not hit within %d seconds. Pending: %s", CLIArgs.BreakpointTimeout, breakpointStr)

							// Write timeout event to log file
							logFilePath := fmt.Sprintf("/tmp/xdebug-cli-daemon-%d.log", CLIArgs.Port)
							logEntry := fmt.Sprintf("[%s] Timeout: %s\n", time.Now().Format("2006-01-02 15:04:05"), errorMsg)
							if logFile, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err == nil {
								logFile.WriteString(logEntry)
								logFile.Close()
							}

							// Signal timeout error to parent process
							d.WriteStatus("error:" + errorMsg)

							// Exit with code 124 (Unix timeout convention)
							d.Shutdown()
							os.Exit(124)
						}
						// Breakpoint was hit in time
						currentFile, currentLine := client.GetSession().GetCurrentLocation()
						if currentFile != "" && pathStore != nil {
							pathStore.SaveBreakpointPath(currentFile)
						}
						// Signal success to parent process with location
						d.WriteStatus(fmt.Sprintf("ready:%s:%d", currentFile, currentLine))
					default:
						// No timeout yet, continue normally
					}
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

// activeSession holds the currently active debugging session
// This allows daemon subcommands to check status without running listen
var activeSession struct {
	sync.RWMutex
	client *dbgp.Client
	active bool
}

// setActiveSession sets the global active session
func setActiveSession(client *dbgp.Client) {
	activeSession.Lock()
	defer activeSession.Unlock()
	activeSession.client = client
	activeSession.active = true
}

// clearActiveSession clears the global active session
func clearActiveSession() {
	activeSession.Lock()
	defer activeSession.Unlock()
	activeSession.client = nil
	activeSession.active = false
}

// getActiveSession returns the current active client and whether it's active
func getActiveSession() (*dbgp.Client, bool) {
	activeSession.RLock()
	defer activeSession.RUnlock()
	return activeSession.client, activeSession.active
}

// processExists checks if a process with the given PID exists
func processExists(pid int) bool {
	// Check if /proc/<pid> exists (Linux-specific)
	procPath := fmt.Sprintf("/proc/%d", pid)
	_, err := os.Stat(procPath)
	return err == nil
}

// findStaleProcessOnPort checks if there's a stale xdebug-cli process on the given port
// Returns the PID if found, or 0 if not found
func findStaleProcessOnPort(port int) int {
	// Use lsof to find process using the port
	cmd := exec.Command("lsof", "-i", fmt.Sprintf(":%d", port), "-t")
	output, err := cmd.Output()
	if err != nil {
		// No process found or lsof not available
		return 0
	}

	// Parse PID from output
	pidStr := strings.TrimSpace(string(output))
	if pidStr == "" {
		return 0
	}

	pid, err := strconv.Atoi(pidStr)
	if err != nil {
		return 0
	}

	// Verify it's an xdebug-cli process by checking /proc/<pid>/comm
	commPath := fmt.Sprintf("/proc/%d/comm", pid)
	commData, err := os.ReadFile(commPath)
	if err != nil {
		return 0
	}

	comm := strings.TrimSpace(string(commData))
	if strings.Contains(comm, "xdebug-cli") {
		return pid
	}

	return 0
}

// executeCurl executes the curl command with XDEBUG_TRIGGER cookie appended.
// It runs asynchronously and returns a channel that receives the error (or nil on success).
// The curl command is parsed using shell-style splitting to handle complex arguments.
func executeCurl(curlArgs string) <-chan error {
	errCh := make(chan error, 1)

	go func() {
		defer close(errCh)

		// Parse curl args using shell-style splitting
		args, err := parseShellArgs(curlArgs)
		if err != nil {
			errCh <- fmt.Errorf("failed to parse curl arguments: %w", err)
			return
		}

		// Append XDEBUG_TRIGGER cookie
		args = append(args, "-b", "XDEBUG_TRIGGER=1")

		// Execute curl
		cmd := exec.Command("curl", args...)
		output, err := cmd.CombinedOutput()
		if err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				errCh <- fmt.Errorf("curl failed with exit code %d: %s", exitErr.ExitCode(), strings.TrimSpace(string(output)))
			} else {
				errCh <- fmt.Errorf("curl failed: %w", err)
			}
			return
		}

		errCh <- nil
	}()

	return errCh
}

// parseShellArgs parses a string into shell-style arguments.
// Handles single quotes, double quotes, and backslash escaping.
func parseShellArgs(s string) ([]string, error) {
	var args []string
	var current strings.Builder
	var inSingleQuote, inDoubleQuote, escaped bool

	for i := 0; i < len(s); i++ {
		c := s[i]

		if escaped {
			current.WriteByte(c)
			escaped = false
			continue
		}

		switch c {
		case '\\':
			if inSingleQuote {
				current.WriteByte(c)
			} else {
				escaped = true
			}
		case '\'':
			if inDoubleQuote {
				current.WriteByte(c)
			} else {
				inSingleQuote = !inSingleQuote
			}
		case '"':
			if inSingleQuote {
				current.WriteByte(c)
			} else {
				inDoubleQuote = !inDoubleQuote
			}
		case ' ', '\t':
			if inSingleQuote || inDoubleQuote {
				current.WriteByte(c)
			} else if current.Len() > 0 {
				args = append(args, current.String())
				current.Reset()
			}
		default:
			current.WriteByte(c)
		}
	}

	if inSingleQuote || inDoubleQuote {
		return nil, fmt.Errorf("unclosed quote in arguments")
	}

	if current.Len() > 0 {
		args = append(args, current.String())
	}

	return args, nil
}

// runDaemonStatus displays the current daemon status
func runDaemonStatus() {
	// First check if there's a daemon session on the current port
	registry, err := daemon.NewSessionRegistry()
	if err == nil {
		sessionInfo, err := registry.Get(CLIArgs.Port)
		if err == nil {
			// Daemon session found
			fmt.Println("Connection Status: Daemon Mode")
			fmt.Println("")
			fmt.Printf("PID: %d\n", sessionInfo.PID)
			fmt.Printf("Port: %d\n", sessionInfo.Port)
			fmt.Printf("Socket Path: %s\n", sessionInfo.SocketPath)
			fmt.Printf("Started: %s\n", sessionInfo.StartedAt.Format("2006-01-02 15:04:05"))
			fmt.Println("")
			fmt.Println("This session is running as a daemon in the background.")
			fmt.Println("Use 'xdebug-cli daemon kill' to terminate the daemon.")
			return
		}
	}

	// Fall back to checking in-process session
	client, active := getActiveSession()

	if !active || client == nil {
		fmt.Println("Connection Status: Not connected")
		fmt.Println("")
		fmt.Println("No active debugging session.")
		fmt.Println("")
		fmt.Println("Start a session with:")
		fmt.Println("  xdebug-cli daemon start")
		return
	}

	session := client.GetSession()
	state := session.GetState()
	file, line := session.GetCurrentLocation()

	fmt.Println("Connection Status: Connected")
	fmt.Println("")
	fmt.Printf("Session State: %s\n", state.String())
	fmt.Printf("IDE Key: %s\n", session.GetIDEKey())
	fmt.Printf("App ID: %s\n", session.GetAppID())

	if file != "" {
		fmt.Printf("Current Location: %s:%d\n", file, line)
	} else {
		fmt.Println("Current Location: Not available")
	}

	targetFiles := session.GetTargetFiles()
	if len(targetFiles) > 0 {
		fmt.Printf("Target File: %s\n", targetFiles[0])
	}

	fmt.Println("")
}

// runDaemonList lists all active daemon sessions
func runDaemonList() {
	registry, err := daemon.NewSessionRegistry()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to access session registry: %v\n", err)
		os.Exit(1)
	}

	sessions := registry.List()

	// Clean up stale sessions (dead PIDs) and get updated list
	activeSessions := []daemon.SessionInfo{}
	for _, session := range sessions {
		if processExists(session.PID) {
			activeSessions = append(activeSessions, session)
		}
	}

	if CLIArgs.JSON {
		// JSON output
		if err := outputSessionListJSON(activeSessions); err != nil {
			fmt.Fprintf(os.Stderr, "Error: Failed to output JSON: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Text output
	if len(activeSessions) == 0 {
		fmt.Println("No active daemon sessions found.")
		fmt.Println("")
		fmt.Println("Start a daemon session with:")
		fmt.Println("  xdebug-cli daemon start")
		return
	}

	fmt.Println("Active Daemon Sessions:")
	fmt.Printf("%-8s %-8s %-20s %s\n", "PID", "Port", "Started", "Socket Path")
	fmt.Println(strings.Repeat("-", 80))

	for _, session := range activeSessions {
		fmt.Printf("%-8d %-8d %-20s %s\n",
			session.PID,
			session.Port,
			session.StartedAt.Format("2006-01-02 15:04:05"),
			session.SocketPath,
		)
	}

	fmt.Println("")
	fmt.Printf("%d session(s) found\n", len(activeSessions))
}

// outputSessionListJSON outputs session list in JSON format
func outputSessionListJSON(sessions []daemon.SessionInfo) error {
	data, err := json.Marshal(sessions)
	if err != nil {
		return err
	}
	fmt.Println(string(data))
	return nil
}

// runDaemonIsAlive checks if a session is active and exits with appropriate code
func runDaemonIsAlive() {
	// First check if there's a daemon session on the current port
	registry, err := daemon.NewSessionRegistry()
	if err == nil {
		sessionInfo, err := registry.Get(CLIArgs.Port)
		if err == nil {
			// Verify the process still exists
			if processExists(sessionInfo.PID) {
				fmt.Println("connected")
				os.Exit(0)
			}
			// Process doesn't exist, fall through to check in-process session
		}
	}

	// Fall back to checking in-process session
	_, active := getActiveSession()

	if active {
		fmt.Println("connected")
		os.Exit(0)
	} else {
		fmt.Println("not connected")
		os.Exit(1)
	}
}

// runDaemonKill terminates the active session
func runDaemonKill() {
	// Handle --all flag
	if CLIArgs.KillAll {
		runDaemonKillAll()
		return
	}

	// First check if there's a daemon session on the current port
	registry, err := daemon.NewSessionRegistry()
	if err == nil {
		sessionInfo, err := registry.Get(CLIArgs.Port)
		if err == nil {
			// Daemon session found - use IPC to kill it
			client := ipc.NewClient(sessionInfo.SocketPath)

			fmt.Printf("Sending kill request to daemon (PID %d)...\n", sessionInfo.PID)

			resp, err := client.Kill()
			if err != nil || !resp.Success {
				// Socket communication failed - try SIGTERM as fallback
				if processExists(sessionInfo.PID) {
					proc, procErr := os.FindProcess(sessionInfo.PID)
					if procErr == nil {
						if sigErr := proc.Signal(syscall.SIGTERM); sigErr == nil {
							// Give process time to terminate
							time.Sleep(100 * time.Millisecond)
							if !processExists(sessionInfo.PID) {
								// Clean up registry entry
								registry.Remove(sessionInfo.Port)
								// Clean up socket file if it exists
								os.Remove(sessionInfo.SocketPath)
								fmt.Println("Daemon terminated successfully (via SIGTERM).")
								return
							}
						}
					}
				}
				if err != nil {
					fmt.Fprintf(os.Stderr, "Failed to send kill request to daemon: %v\n", err)
				} else {
					fmt.Fprintf(os.Stderr, "Kill request failed: %s\n", resp.Error)
				}
				fmt.Fprintf(os.Stderr, "The daemon process (PID %d) may need to be killed manually.\n", sessionInfo.PID)
				os.Exit(1)
			}

			fmt.Println("Daemon terminated successfully.")
			return
		}
	}

	// Check for stale processes on the port
	stalePID := findStaleProcessOnPort(CLIArgs.Port)
	if stalePID != 0 {
		fmt.Printf("Found stale xdebug-cli process (PID %d) on port %d\n", stalePID, CLIArgs.Port)
		fmt.Printf("Killing stale process...\n")

		// Kill the stale process using SIGTERM
		process, err := os.FindProcess(stalePID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: Failed to find process: %v\n", err)
			os.Exit(1)
		}

		err = process.Kill()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: Failed to kill stale process: %v\n", err)
			fmt.Fprintf(os.Stderr, "Try manually: kill %d\n", stalePID)
			os.Exit(1)
		}

		fmt.Println("Stale process terminated successfully.")
		return
	}

	// Fall back to in-process session
	client, active := getActiveSession()

	if !active || client == nil {
		fmt.Fprintf(os.Stderr, "No active session to kill.\n")
		fmt.Fprintf(os.Stderr, "\nCheck session status with: xdebug-cli daemon status\n")
		os.Exit(1)
	}

	// Try to send stop command
	_, err = client.Finish()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Failed to send stop command: %v\n", err)
	}

	// Close the connection
	err = client.Close()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Failed to close connection: %v\n", err)
	}

	// Clear the session
	clearActiveSession()

	fmt.Println("Session terminated.")
}

// runDaemonKillAll terminates all daemon sessions
func runDaemonKillAll() {
	registry, err := daemon.NewSessionRegistry()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to access session registry: %v\n", err)
		os.Exit(1)
	}

	sessions := registry.List()

	// Filter active sessions (verify PIDs exist)
	activeSessions := []daemon.SessionInfo{}
	for _, session := range sessions {
		if processExists(session.PID) {
			activeSessions = append(activeSessions, session)
		}
	}

	if len(activeSessions) == 0 {
		fmt.Println("No active daemon sessions found.")
		return
	}

	// Show confirmation prompt unless --force is used
	if !CLIArgs.Force {
		fmt.Printf("Found %d active session(s). Terminate all? (y/N): ", len(activeSessions))
		var response string
		fmt.Scanln(&response)
		response = strings.ToLower(strings.TrimSpace(response))
		if response != "y" && response != "yes" {
			fmt.Println("Operation cancelled.")
			return
		}
	}

	// Kill each session
	successCount := 0
	failedSessions := []daemon.SessionInfo{}

	for _, session := range activeSessions {
		fmt.Printf("Killing daemon on port %d (PID %d)... ", session.Port, session.PID)

		client := ipc.NewClient(session.SocketPath)
		resp, err := client.Kill()

		if err != nil || !resp.Success {
			// Socket communication failed - try SIGTERM as fallback
			if processExists(session.PID) {
				proc, procErr := os.FindProcess(session.PID)
				if procErr == nil {
					if sigErr := proc.Signal(syscall.SIGTERM); sigErr == nil {
						// Give process time to terminate
						time.Sleep(100 * time.Millisecond)
						if !processExists(session.PID) {
							fmt.Println("done (via SIGTERM)")
							// Clean up registry entry
							registry.Remove(session.Port)
							// Clean up socket file if it exists
							os.Remove(session.SocketPath)
							successCount++
							continue
						}
					}
				}
			}
			fmt.Println("failed")
			failedSessions = append(failedSessions, session)
			if err != nil {
				fmt.Fprintf(os.Stderr, "  Error: %v\n", err)
			} else {
				fmt.Fprintf(os.Stderr, "  Error: %s\n", resp.Error)
			}
		} else {
			fmt.Println("done")
			successCount++
		}
	}

	// Summary
	fmt.Println("")
	if successCount == len(activeSessions) {
		fmt.Printf("All %d session(s) terminated successfully.\n", successCount)
	} else {
		fmt.Printf("%d of %d session(s) terminated.\n", successCount, len(activeSessions))
		if len(failedSessions) > 0 {
			fmt.Println("\nFailed sessions:")
			for _, session := range failedSessions {
				fmt.Printf("  Port %d (PID %d)\n", session.Port, session.PID)
			}
			os.Exit(1)
		}
	}
}
