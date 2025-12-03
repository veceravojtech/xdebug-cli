package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"

	"github.com/console/xdebug-cli/internal/daemon"
	"github.com/console/xdebug-cli/internal/dbgp"
	"github.com/console/xdebug-cli/internal/ipc"
	"github.com/spf13/cobra"
)

// activeSession holds the currently active debugging session
// This allows connection subcommands to check status without running listen
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

var connectionCmd = &cobra.Command{
	Use:   "connection",
	Short: "Manage debugging connections",
	Long: `Show connection status, check if alive, or kill active debugging session.

Supports both interactive sessions and background daemon sessions.
For daemon sessions, displays PID, port, and socket information.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Default behavior: show connection status
		runConnectionStatus()
	},
}

var connectionIsAliveCmd = &cobra.Command{
	Use:   "isAlive",
	Short: "Check if a debugging session is active",
	Long: `Check if there is an active debugging connection.
Checks both interactive sessions and daemon processes.
Exits with code 0 if connected, code 1 if not connected.`,
	Run: func(cmd *cobra.Command, args []string) {
		runConnectionIsAlive()
	},
}

var connectionListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all active daemon sessions",
	Long: `List all active daemon debugging sessions across all ports.

Shows PID, port, socket path, and start time for each session.
Supports JSON output for automation with --json flag.`,
	Run: func(cmd *cobra.Command, args []string) {
		runConnectionList()
	},
}

var connectionKillCmd = &cobra.Command{
	Use:   "kill",
	Short: "Terminate the active debugging session",
	Long: `Terminate the currently active debugging session and close the connection.

For daemon sessions, sends a kill request via IPC and shuts down the daemon.
For interactive sessions, sends a stop command and closes the connection.
Use --all flag to terminate all daemon sessions at once.`,
	Run: func(cmd *cobra.Command, args []string) {
		runConnectionKill()
	},
}

func init() {
	connectionCmd.AddCommand(connectionIsAliveCmd)
	connectionCmd.AddCommand(connectionListCmd)
	connectionCmd.AddCommand(connectionKillCmd)

	// Add flags for list command
	connectionListCmd.Flags().BoolVar(&CLIArgs.JSON, "json", false, "Output in JSON format")

	// Add flags for kill command
	connectionKillCmd.Flags().BoolVar(&CLIArgs.KillAll, "all", false, "Kill all daemon sessions")
	connectionKillCmd.Flags().BoolVar(&CLIArgs.Force, "force", false, "Skip confirmation prompt")

	rootCmd.AddCommand(connectionCmd)
}

// runConnectionStatus displays the current connection status
func runConnectionStatus() {
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
			fmt.Println("Use 'xdebug-cli connection kill' to terminate the daemon.")
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
		fmt.Println("  xdebug-cli listen --commands \"run\"  # Command-based execution")
		fmt.Println("  xdebug-cli daemon start            # Background daemon mode")
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

// runConnectionList lists all active daemon sessions
func runConnectionList() {
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

// runConnectionIsAlive checks if a session is active and exits with appropriate code
func runConnectionIsAlive() {
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

// runConnectionKill terminates the active session
func runConnectionKill() {
	// Handle --all flag
	if CLIArgs.KillAll {
		runConnectionKillAll()
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
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to send kill request to daemon: %v\n", err)
				fmt.Fprintf(os.Stderr, "The daemon process (PID %d) may need to be killed manually.\n", sessionInfo.PID)
				os.Exit(1)
			}

			if !resp.Success {
				fmt.Fprintf(os.Stderr, "Kill request failed: %s\n", resp.Error)
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
		fmt.Fprintf(os.Stderr, "\nCheck session status with: xdebug-cli connection\n")
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

// runConnectionKillAll terminates all daemon sessions
func runConnectionKillAll() {
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
