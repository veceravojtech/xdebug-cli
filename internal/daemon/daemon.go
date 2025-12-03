package daemon

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/console/xdebug-cli/internal/dbgp"
	"github.com/console/xdebug-cli/internal/ipc"
)

// Daemon represents a background daemon process that manages debug sessions
type Daemon struct {
	server     *dbgp.Server
	ipcServer  *ipc.Server
	client     *dbgp.Client
	executor   *CommandExecutor
	registry   *SessionRegistry
	port       int
	pidFile    string
	socketPath string
	shutdown   chan os.Signal
	ctx        context.Context
	cancel     context.CancelFunc
	mu         sync.Mutex
}

// NewDaemon creates a new daemon instance
func NewDaemon(server *dbgp.Server, port int) (*Daemon, error) {
	// Create session registry
	registry, err := NewSessionRegistry()
	if err != nil {
		return nil, fmt.Errorf("failed to create session registry: %w", err)
	}

	// Generate PID file path
	pidFile := fmt.Sprintf("/tmp/xdebug-cli-daemon-%d.pid", port)

	// Generate socket path
	socketPath := fmt.Sprintf("/tmp/xdebug-cli-session-%d.sock", port)

	ctx, cancel := context.WithCancel(context.Background())

	return &Daemon{
		server:     server,
		registry:   registry,
		port:       port,
		pidFile:    pidFile,
		socketPath: socketPath,
		shutdown:   make(chan os.Signal, 1),
		ctx:        ctx,
		cancel:     cancel,
	}, nil
}

// CheckExisting checks if a daemon is already running on this port
func (d *Daemon) CheckExisting() (bool, int, error) {
	// Check if PID file exists
	if _, err := os.Stat(d.pidFile); os.IsNotExist(err) {
		return false, 0, nil
	}

	// Read PID from file
	data, err := os.ReadFile(d.pidFile)
	if err != nil {
		return false, 0, fmt.Errorf("failed to read PID file: %w", err)
	}

	pid, err := strconv.Atoi(string(data))
	if err != nil {
		return false, 0, fmt.Errorf("invalid PID in file: %w", err)
	}

	// Check if process exists
	if processExists(pid) {
		return true, pid, nil
	}

	// PID file is stale, remove it
	os.Remove(d.pidFile)
	return false, 0, nil
}

// Initialize initializes the daemon infrastructure (PID, registry, IPC server)
// This should be called before waiting for the first Xdebug connection
func (d *Daemon) Initialize() error {
	// Ensure cleanup happens even on panic or early return
	cleanupStarted := false
	defer func() {
		if r := recover(); r != nil {
			// Panic occurred, ensure cleanup
			if !cleanupStarted {
				d.emergencyCleanup()
			}
			panic(r) // Re-panic after cleanup
		}
	}()

	// Write PID file
	if err := d.writePIDFile(); err != nil {
		return fmt.Errorf("failed to write PID file: %w", err)
	}

	// Register in session registry
	sessionInfo := SessionInfo{
		PID:        os.Getpid(),
		Port:       d.port,
		SocketPath: d.socketPath,
		StartedAt:  time.Now(),
	}
	if err := d.registry.Add(sessionInfo); err != nil {
		d.removePIDFile()
		return fmt.Errorf("failed to register session: %w", err)
	}

	// Setup signal handlers
	d.setupSignalHandlers()

	// Create IPC server
	ipcServer := ipc.NewServer(d.socketPath, d.handleIPCRequest)
	d.ipcServer = ipcServer

	// Start IPC server
	if err := ipcServer.Listen(); err != nil {
		d.Shutdown()
		return fmt.Errorf("failed to start IPC server: %w", err)
	}

	// Start IPC server in background
	go func() {
		if err := ipcServer.Serve(); err != nil {
			// IPC server error, initiate shutdown
			cleanupStarted = true
			d.Shutdown()
		}
	}()

	return nil
}

// SetClient sets the active DBGp client for this daemon
// This should be called after an Xdebug connection is established
func (d *Daemon) SetClient(client *dbgp.Client) {
	d.mu.Lock()
	d.client = client
	d.executor = NewCommandExecutor(client)
	d.mu.Unlock()
}

// Start starts the daemon process in the current process (no fork)
// This should be called after forking to run the daemon logic
// DEPRECATED: Use Initialize() and SetClient() instead for better control
func (d *Daemon) Start(client *dbgp.Client) error {
	// For backward compatibility, call Initialize first
	if d.ipcServer == nil {
		if err := d.Initialize(); err != nil {
			return err
		}
	}

	// Set the client
	d.SetClient(client)

	return nil
}

// emergencyCleanup performs best-effort cleanup without locks (for panic recovery)
func (d *Daemon) emergencyCleanup() {
	// Best-effort cleanup without acquiring locks (may already be held in panic scenario)
	os.Remove(d.pidFile)
	os.Remove(d.socketPath)
	// Note: Registry cleanup skipped in emergency - will be cleaned up on next start
}

// Fork creates a background daemon process using fork/exec
func (d *Daemon) Fork(args []string) error {
	// Check for existing daemon
	exists, pid, err := d.CheckExisting()
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("daemon already running on port %d (PID %d)\nUse 'xdebug-cli daemon kill' to terminate it first.", d.port, pid)
	}

	// Get current executable path
	executable, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	// Resolve symlinks
	executable, err = filepath.EvalSymlinks(executable)
	if err != nil {
		return fmt.Errorf("failed to resolve executable path: %w", err)
	}

	// Prepare environment
	env := os.Environ()

	// Add marker to indicate we're in daemon mode
	env = append(env, "XDEBUG_CLI_DAEMON_MODE=1")

	// Setup file descriptors for background process
	devNull, err := os.OpenFile(os.DevNull, os.O_RDWR, 0)
	if err != nil {
		return fmt.Errorf("failed to open /dev/null: %w", err)
	}
	defer devNull.Close()

	// Fork process
	attr := &syscall.ProcAttr{
		Dir: "",
		Env: env,
		Files: []uintptr{
			devNull.Fd(), // stdin
			devNull.Fd(), // stdout
			devNull.Fd(), // stderr
		},
		Sys: &syscall.SysProcAttr{
			Setsid: true, // Create new session (detach from terminal)
		},
	}

	pid, err = syscall.ForkExec(executable, args, attr)
	if err != nil {
		return fmt.Errorf("failed to fork daemon process: %w", err)
	}

	// Parent process: return success with daemon PID
	fmt.Printf("Daemon started on port %d (PID %d)\n", d.port, pid)
	return nil
}

// IsDaemonMode checks if we're running in daemon mode
func IsDaemonMode() bool {
	return os.Getenv("XDEBUG_CLI_DAEMON_MODE") == "1"
}

// setupSignalHandlers registers signal handlers for graceful shutdown
func (d *Daemon) setupSignalHandlers() {
	signal.Notify(d.shutdown, syscall.SIGTERM, syscall.SIGINT)
	go d.handleShutdownSignal()
}

// handleShutdownSignal waits for shutdown signal and initiates cleanup
func (d *Daemon) handleShutdownSignal() {
	<-d.shutdown
	d.Shutdown()
}

// handleIPCRequest processes incoming IPC requests
func (d *Daemon) handleIPCRequest(req *ipc.CommandRequest) *ipc.CommandResponse {
	// Validate request type first
	switch req.Type {
	case "execute_commands", "kill":
		// Valid request types
	default:
		return ipc.NewErrorResponse(fmt.Sprintf("unknown request type: %s", req.Type))
	}

	// Handle kill request (doesn't require active session)
	if req.Type == "kill" {
		go func() {
			time.Sleep(100 * time.Millisecond) // Give time to send response
			d.Shutdown()
		}()
		return ipc.NewSuccessResponse([]ipc.CommandResult{
			{
				Command: "kill",
				Success: true,
				Result:  map[string]interface{}{"message": "Daemon shutting down"},
			},
		})
	}

	// For execute_commands, check if client is available
	d.mu.Lock()
	executor := d.executor
	client := d.client
	d.mu.Unlock()

	if client == nil || executor == nil {
		return ipc.NewErrorResponse("no active debug session")
	}

	// Execute commands
	results := executor.ExecuteCommands(req.Commands, req.JSONOutput)
	return ipc.NewSuccessResponse(results)
}

// Shutdown performs graceful shutdown of the daemon with timeout
func (d *Daemon) Shutdown() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	// Cancel context to signal shutdown
	d.cancel()

	// Create shutdown context with 5 second timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	var errors []error
	done := make(chan bool, 1)

	// Perform shutdown in goroutine with timeout
	go func() {
		// Close IPC server (this also removes Unix socket)
		if d.ipcServer != nil {
			if err := d.ipcServer.Shutdown(); err != nil {
				errors = append(errors, fmt.Errorf("IPC server shutdown error: %w", err))
			}
		}

		// Close DBGp client (sends stop command)
		if d.client != nil {
			if err := d.client.Close(); err != nil {
				errors = append(errors, fmt.Errorf("client close error: %w", err))
			}
		}

		// Close DBGp server
		if d.server != nil {
			if err := d.server.Close(); err != nil {
				errors = append(errors, fmt.Errorf("server close error: %w", err))
			}
		}

		done <- true
	}()

	// Wait for shutdown or timeout
	select {
	case <-done:
		// Graceful shutdown completed
	case <-shutdownCtx.Done():
		// Timeout exceeded, force cleanup
		errors = append(errors, fmt.Errorf("shutdown timeout exceeded, forcing cleanup"))
	}

	// Always cleanup registry and PID file, even on timeout
	if err := d.registry.Remove(d.port); err != nil {
		errors = append(errors, fmt.Errorf("registry removal error: %w", err))
	}

	if err := d.removePIDFile(); err != nil {
		errors = append(errors, fmt.Errorf("PID file removal error: %w", err))
	}

	// Ensure socket file is removed (defensive cleanup)
	if err := os.Remove(d.socketPath); err != nil && !os.IsNotExist(err) {
		errors = append(errors, fmt.Errorf("socket file removal error: %w", err))
	}

	if len(errors) > 0 {
		// Return first error (could be improved to return multiple)
		return errors[0]
	}

	return nil
}

// Wait blocks until daemon shutdown is complete
func (d *Daemon) Wait() {
	<-d.ctx.Done()
}

// writePIDFile writes the current process PID to the PID file
func (d *Daemon) writePIDFile() error {
	pid := os.Getpid()
	data := []byte(strconv.Itoa(pid))
	if err := os.WriteFile(d.pidFile, data, 0600); err != nil {
		return fmt.Errorf("failed to write PID file: %w", err)
	}
	return nil
}

// removePIDFile removes the PID file
func (d *Daemon) removePIDFile() error {
	if err := os.Remove(d.pidFile); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove PID file: %w", err)
	}
	return nil
}

// GetSocketPath returns the Unix socket path for IPC
func (d *Daemon) GetSocketPath() string {
	return d.socketPath
}

// GetPort returns the DBGp server port
func (d *Daemon) GetPort() int {
	return d.port
}

// GetPIDFile returns the PID file path
func (d *Daemon) GetPIDFile() string {
	return d.pidFile
}
