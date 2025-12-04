package dbgp

import (
	"context"
	"fmt"
	"net"
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
