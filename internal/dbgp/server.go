package dbgp

import (
	"context"
	"fmt"
	"net"
	"os/exec"
	"regexp"
	"runtime"
	"strings"
	"syscall"
	"time"
)

// Server represents a DBGp protocol server that listens for connections
type Server struct {
	address  string
	port     int
	listener net.Listener
}

// NewServer creates a new DBGp server
func NewServer(address string, port int) *Server {
	return &Server{
		address: address,
		port:    port,
	}
}

// Listen starts listening for connections on the configured address and port
func (s *Server) Listen() error {
	addr := fmt.Sprintf("%s:%d", s.address, s.port)

	// Use ListenConfig with SO_REUSEADDR to allow immediate port reuse
	lc := net.ListenConfig{
		Control: func(network, address string, c syscall.RawConn) error {
			var sockoptErr error
			err := c.Control(func(fd uintptr) {
				sockoptErr = syscall.SetsockoptInt(int(fd), syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1)
			})
			if err != nil {
				return err
			}
			return sockoptErr
		},
	}

	listener, err := lc.Listen(context.Background(), "tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", addr, err)
	}

	s.listener = listener
	return nil
}

// Accept waits for and accepts incoming connections, calling the handler for each
func (s *Server) Accept(handler func(*Connection)) error {
	if s.listener == nil {
		return fmt.Errorf("server not listening")
	}

	for {
		listener := s.listener
		if listener == nil {
			return nil
		}

		conn, err := listener.Accept()
		if err != nil {
			// Check if the error is due to the listener being closed
			if opErr, ok := err.(*net.OpError); ok && opErr.Err != nil && opErr.Err.Error() == "use of closed network connection" {
				return nil
			}
			// Also check for common close errors
			if s.listener == nil {
				return nil
			}
			return fmt.Errorf("failed to accept connection: %w", err)
		}

		// Handle connection
		dbgpConn := NewConnection(conn)
		handler(dbgpConn)
	}
}

// AcceptWithTimeout waits for incoming connections with a timeout
// Returns an error if timeout is reached before a connection arrives
func (s *Server) AcceptWithTimeout(timeout time.Duration, handler func(*Connection)) error {
	if s.listener == nil {
		return fmt.Errorf("server not listening")
	}

	// Set deadline for accept
	tcpListener, ok := s.listener.(*net.TCPListener)
	if !ok {
		return fmt.Errorf("listener is not a TCP listener")
	}

	deadline := time.Now().Add(timeout)
	if err := tcpListener.SetDeadline(deadline); err != nil {
		return fmt.Errorf("failed to set deadline: %w", err)
	}

	// Try to accept a connection
	conn, err := s.listener.Accept()
	if err != nil {
		// Check if it's a timeout error
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			return fmt.Errorf("timeout waiting for connection after %v", timeout)
		}
		// Check if the error is due to the listener being closed
		if opErr, ok := err.(*net.OpError); ok && opErr.Err != nil && opErr.Err.Error() == "use of closed network connection" {
			return nil
		}
		if s.listener == nil {
			return nil
		}
		return fmt.Errorf("failed to accept connection: %w", err)
	}

	// Clear deadline for subsequent operations
	if err := tcpListener.SetDeadline(time.Time{}); err != nil {
		return fmt.Errorf("failed to clear deadline: %w", err)
	}

	// Handle connection
	dbgpConn := NewConnection(conn)
	handler(dbgpConn)

	return nil
}

// Close stops the server and closes the listener
func (s *Server) Close() error {
	if s.listener != nil {
		err := s.listener.Close()
		s.listener = nil
		return err
	}
	return nil
}

// GetAddress returns the server's listen address
func (s *Server) GetAddress() string {
	return fmt.Sprintf("%s:%d", s.address, s.port)
}

// IsListening returns true if the server is currently listening
func (s *Server) IsListening() bool {
	return s.listener != nil
}

// PortConflictInfo contains information about a process using a port
type PortConflictInfo struct {
	Port        int
	ProcessName string
	PID         string
	IsIDE       bool
}

// CheckPortConflict checks if another process is listening on the port
// Returns nil if the port is free, or PortConflictInfo if occupied
func (s *Server) CheckPortConflict() *PortConflictInfo {
	return CheckPortInUse(s.port)
}

// CheckPortInUse checks if a port is in use by another process
func CheckPortInUse(port int) *PortConflictInfo {
	// Try to connect to the port - if successful, something is listening
	addr := fmt.Sprintf("127.0.0.1:%d", port)
	conn, err := net.DialTimeout("tcp", addr, 100*time.Millisecond)
	if err != nil {
		// Connection refused means port is free
		return nil
	}
	conn.Close()

	// Port is in use, try to identify the process
	info := &PortConflictInfo{Port: port}
	identifyProcessOnPort(port, info)

	return info
}

// identifyProcessOnPort uses lsof to identify the process using a port
func identifyProcessOnPort(port int, info *PortConflictInfo) {
	if runtime.GOOS == "windows" {
		// Windows: use netstat
		identifyProcessWindows(port, info)
		return
	}

	// macOS/Linux: use lsof
	cmd := exec.Command("lsof", "-i", fmt.Sprintf(":%d", port), "-sTCP:LISTEN", "-n", "-P")
	output, err := cmd.Output()
	if err != nil {
		info.ProcessName = "unknown process"
		return
	}

	parseLosfOutput(string(output), info)
}

// parseLosfOutput extracts process info from lsof output
func parseLosfOutput(output string, info *PortConflictInfo) {
	lines := strings.Split(output, "\n")
	if len(lines) < 2 {
		info.ProcessName = "unknown process"
		return
	}

	// Parse the data line (skip header)
	// Example: phpstorm 68135 vecera  735u  IPv6 0xc869... TCP *:9003 (LISTEN)
	fields := strings.Fields(lines[1])
	if len(fields) >= 2 {
		info.ProcessName = fields[0]
		info.PID = fields[1]
	} else {
		info.ProcessName = "unknown process"
		return
	}

	// Check if it's a known IDE
	checkIfIDE(info)
}

// identifyProcessWindows uses netstat on Windows
func identifyProcessWindows(port int, info *PortConflictInfo) {
	cmd := exec.Command("netstat", "-ano")
	output, err := cmd.Output()
	if err != nil {
		info.ProcessName = "unknown process"
		return
	}

	// Find the line with our port
	portStr := fmt.Sprintf(":%d", port)
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, portStr) && strings.Contains(line, "LISTENING") {
			fields := strings.Fields(line)
			if len(fields) >= 5 {
				info.PID = fields[4]
				// Try to get process name from PID
				info.ProcessName = getProcessNameWindows(info.PID)
			}
			break
		}
	}

	if info.ProcessName == "" {
		info.ProcessName = "unknown process"
	}
	checkIfIDE(info)
}

// getProcessNameWindows gets process name from PID on Windows
func getProcessNameWindows(pid string) string {
	cmd := exec.Command("tasklist", "/FI", fmt.Sprintf("PID eq %s", pid), "/FO", "CSV", "/NH")
	output, err := cmd.Output()
	if err != nil {
		return "unknown"
	}
	// Parse CSV: "process.exe","12345",...
	re := regexp.MustCompile(`"([^"]+)"`)
	matches := re.FindStringSubmatch(string(output))
	if len(matches) >= 2 {
		return strings.TrimSuffix(matches[1], ".exe")
	}
	return "unknown"
}

// checkIfIDE checks if the process is a known IDE debugger
func checkIfIDE(info *PortConflictInfo) {
	name := strings.ToLower(info.ProcessName)

	// Known IDE process names
	ideNames := []string{
		"phpstorm",
		"idea",          // IntelliJ IDEA
		"webstorm",
		"pycharm",
		"goland",
		"rubymine",
		"clion",
		"rider",
		"datagrip",
		"code",          // VS Code
		"vscodium",
		"eclipse",
		"netbeans",
		"sublime",
		"atom",
		"java",          // Eclipse/Java debuggers
	}

	for _, ide := range ideNames {
		if strings.Contains(name, ide) {
			info.IsIDE = true
			return
		}
	}
}

// FormatPortConflictError creates a user-friendly error message for port conflicts
func FormatPortConflictError(info *PortConflictInfo) string {
	if info == nil {
		return ""
	}

	var msg strings.Builder
	msg.WriteString(fmt.Sprintf("Error: port %d is already in use", info.Port))

	if info.ProcessName != "" && info.ProcessName != "unknown process" {
		msg.WriteString(fmt.Sprintf(" by %s", info.ProcessName))
		if info.PID != "" {
			msg.WriteString(fmt.Sprintf(" (PID %s)", info.PID))
		}
	}
	msg.WriteString("\n")

	if info.IsIDE {
		msg.WriteString("\nAnother debugger is listening on this port.\n")
		msg.WriteString("Xdebug connections will go to that debugger instead of xdebug-cli.\n\n")

		switch strings.ToLower(info.ProcessName) {
		case "phpstorm":
			msg.WriteString("To fix: In PhpStorm, click the phone icon in the toolbar to stop listening,\n")
			msg.WriteString("        or go to Run > Stop Listening for PHP Debug Connections\n")
		default:
			msg.WriteString("To fix: Stop the debug listener in your IDE, then retry.\n")
		}
	} else {
		msg.WriteString("\nTo fix: Stop the process using the port, or use a different port:\n")
		msg.WriteString(fmt.Sprintf("        xdebug-cli daemon start -p %d ...\n", info.Port+1))
	}

	return msg.String()
}
