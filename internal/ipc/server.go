package ipc

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"sync"
)

// RequestHandler is a function that processes IPC requests
type RequestHandler func(*CommandRequest) *CommandResponse

// Server represents an IPC server using Unix domain sockets
type Server struct {
	socketPath string
	listener   net.Listener
	handler    RequestHandler
	mu         sync.Mutex
	ctx        context.Context
	cancel     context.CancelFunc
	wg         sync.WaitGroup
}

// NewServer creates a new IPC server
func NewServer(socketPath string, handler RequestHandler) *Server {
	ctx, cancel := context.WithCancel(context.Background())
	return &Server{
		socketPath: socketPath,
		handler:    handler,
		ctx:        ctx,
		cancel:     cancel,
	}
}

// Listen starts the IPC server and binds to the Unix socket
func (s *Server) Listen() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.listener != nil {
		return fmt.Errorf("server already listening")
	}

	// Remove existing socket file if it exists
	if err := os.Remove(s.socketPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove existing socket: %w", err)
	}

	// Create Unix socket listener
	listener, err := net.Listen("unix", s.socketPath)
	if err != nil {
		return fmt.Errorf("failed to create Unix socket: %w", err)
	}

	// Set socket permissions to 0600 (owner only)
	if err := os.Chmod(s.socketPath, 0600); err != nil {
		listener.Close()
		return fmt.Errorf("failed to set socket permissions: %w", err)
	}

	s.listener = listener
	return nil
}

// Serve accepts and handles incoming connections
func (s *Server) Serve() error {
	if s.listener == nil {
		return fmt.Errorf("server not listening, call Listen() first")
	}

	for {
		select {
		case <-s.ctx.Done():
			return nil
		default:
		}

		conn, err := s.listener.Accept()
		if err != nil {
			// Check if server was shut down
			select {
			case <-s.ctx.Done():
				return nil
			default:
				return fmt.Errorf("failed to accept connection: %w", err)
			}
		}

		s.wg.Add(1)
		go s.handleConnection(conn)
	}
}

// handleConnection processes a single client connection
func (s *Server) handleConnection(conn net.Conn) {
	defer s.wg.Done()
	defer conn.Close()

	reader := bufio.NewReader(conn)

	// Read request (JSON terminated by newline)
	line, err := reader.ReadBytes('\n')
	if err != nil {
		if err != io.EOF {
			s.writeError(conn, fmt.Sprintf("failed to read request: %v", err))
		}
		return
	}

	// Parse request
	var req CommandRequest
	if err := req.FromJSON(line); err != nil {
		s.writeError(conn, fmt.Sprintf("invalid request: %v", err))
		return
	}

	// Handle request
	resp := s.handler(&req)

	// Send response
	if err := s.writeResponse(conn, resp); err != nil {
		// Can't send error response if write failed
		return
	}
}

// writeResponse writes a CommandResponse to the connection
func (s *Server) writeResponse(conn net.Conn, resp *CommandResponse) error {
	data, err := resp.ToJSON()
	if err != nil {
		return fmt.Errorf("failed to serialize response: %w", err)
	}

	// Write JSON followed by newline
	if _, err := conn.Write(append(data, '\n')); err != nil {
		return fmt.Errorf("failed to write response: %w", err)
	}

	return nil
}

// writeError writes an error response to the connection
func (s *Server) writeError(conn net.Conn, errMsg string) {
	resp := NewErrorResponse(errMsg)
	_ = s.writeResponse(conn, resp) // Ignore error since we're already handling an error
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Signal shutdown
	s.cancel()

	// Close listener
	if s.listener != nil {
		if err := s.listener.Close(); err != nil {
			return fmt.Errorf("failed to close listener: %w", err)
		}
	}

	// Wait for active connections to finish
	s.wg.Wait()

	// Remove socket file
	if err := os.Remove(s.socketPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove socket file: %w", err)
	}

	return nil
}

// SocketPath returns the Unix socket path
func (s *Server) SocketPath() string {
	return s.socketPath
}

// IsRunning returns whether the server is currently running
func (s *Server) IsRunning() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.listener != nil
}
