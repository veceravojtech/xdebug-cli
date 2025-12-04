package dbgp

import (
	"fmt"
	"net"
	"sync"
	"testing"
	"time"
)

func TestNewServer(t *testing.T) {
	server := NewServer("127.0.0.1", 9003)

	if server == nil {
		t.Fatal("Expected non-nil server")
	}

	if server.address != "127.0.0.1" {
		t.Errorf("Expected address '127.0.0.1', got '%s'", server.address)
	}

	if server.port != 9003 {
		t.Errorf("Expected port 9003, got %d", server.port)
	}
}

func TestServer_Listen(t *testing.T) {
	server := NewServer("127.0.0.1", 0) // Use port 0 for automatic assignment

	err := server.Listen()
	if err != nil {
		t.Fatalf("Failed to start listening: %v", err)
	}
	defer server.Close()

	if !server.IsListening() {
		t.Error("Expected server to be listening")
	}

	if server.listener == nil {
		t.Error("Expected listener to be non-nil")
	}
}

func TestServer_Listen_InvalidAddress(t *testing.T) {
	// Try to listen on an invalid address
	server := NewServer("999.999.999.999", 9003)

	err := server.Listen()
	if err == nil {
		t.Error("Expected error for invalid address, got nil")
		server.Close()
	}
}

func TestServer_Accept(t *testing.T) {
	server := NewServer("127.0.0.1", 0)
	err := server.Listen()
	if err != nil {
		t.Fatalf("Failed to start listening: %v", err)
	}
	defer server.Close()

	// Get the actual port assigned
	addr := server.listener.Addr().(*net.TCPAddr)
	port := addr.Port

	var wg sync.WaitGroup
	var receivedConn *Connection

	// Start accepting connections in a goroutine
	wg.Add(1)
	go func() {
		defer wg.Done()
		err := server.Accept(func(conn *Connection) {
			receivedConn = conn
			conn.Close()
			server.Close() // Close server to stop Accept loop
		})
		// Accept returns nil when server is closed normally
		if err != nil {
			t.Errorf("Unexpected error from Accept: %v", err)
		}
	}()

	// Give the server time to start accepting
	time.Sleep(50 * time.Millisecond)

	// Connect to the server
	conn, err := net.Dial("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		t.Fatalf("Failed to connect to server: %v", err)
	}
	defer conn.Close()

	// Wait for the handler to be called
	wg.Wait()

	if receivedConn == nil {
		t.Error("Expected connection to be received by handler")
	}
}

func TestServer_Accept_NotListening(t *testing.T) {
	server := NewServer("127.0.0.1", 9003)

	err := server.Accept(func(conn *Connection) {})
	if err == nil {
		t.Error("Expected error when accepting without listening, got nil")
	}
}

func TestServer_Close(t *testing.T) {
	server := NewServer("127.0.0.1", 0)
	err := server.Listen()
	if err != nil {
		t.Fatalf("Failed to start listening: %v", err)
	}

	err = server.Close()
	if err != nil {
		t.Errorf("Unexpected error closing server: %v", err)
	}

	// Closing again should not error
	err = server.Close()
	if err != nil {
		t.Errorf("Unexpected error closing already-closed server: %v", err)
	}
}

func TestServer_GetAddress(t *testing.T) {
	server := NewServer("127.0.0.1", 9003)

	addr := server.GetAddress()
	expected := "127.0.0.1:9003"

	if addr != expected {
		t.Errorf("Expected address '%s', got '%s'", expected, addr)
	}
}

func TestServer_IsListening(t *testing.T) {
	server := NewServer("127.0.0.1", 0)

	if server.IsListening() {
		t.Error("Expected server to not be listening initially")
	}

	err := server.Listen()
	if err != nil {
		t.Fatalf("Failed to start listening: %v", err)
	}
	defer server.Close()

	if !server.IsListening() {
		t.Error("Expected server to be listening after Listen()")
	}
}

func TestServer_MultipleConnections(t *testing.T) {
	server := NewServer("127.0.0.1", 0)
	err := server.Listen()
	if err != nil {
		t.Fatalf("Failed to start listening: %v", err)
	}
	defer server.Close()

	addr := server.listener.Addr().(*net.TCPAddr)
	port := addr.Port

	connectionCount := 0
	var mu sync.Mutex
	var wg sync.WaitGroup

	// Start accepting connections
	wg.Add(1)
	go func() {
		defer wg.Done()
		server.Accept(func(conn *Connection) {
			mu.Lock()
			connectionCount++
			count := connectionCount
			mu.Unlock()

			conn.Close()

			// Close server after receiving 3 connections
			if count >= 3 {
				server.Close()
			}
		})
	}()

	// Give the server time to start accepting
	time.Sleep(50 * time.Millisecond)

	// Connect multiple times
	for i := 0; i < 3; i++ {
		conn, err := net.Dial("tcp", fmt.Sprintf("127.0.0.1:%d", port))
		if err != nil {
			t.Errorf("Failed to connect (attempt %d): %v", i+1, err)
			continue
		}
		conn.Close()
		time.Sleep(10 * time.Millisecond)
	}

	// Wait for all connections to be processed
	wg.Wait()

	mu.Lock()
	finalCount := connectionCount
	mu.Unlock()

	if finalCount != 3 {
		t.Errorf("Expected 3 connections, got %d", finalCount)
	}
}

func TestServer_ImmediateRebind(t *testing.T) {
	// Pick a specific port to test rebinding
	port := 19003 // Use a non-standard port to avoid conflicts

	// First server: bind, close
	server1 := NewServer("127.0.0.1", port)
	err := server1.Listen()
	if err != nil {
		t.Fatalf("Failed to start first server: %v", err)
	}

	// Verify server1 is listening
	if !server1.IsListening() {
		t.Fatal("Expected first server to be listening")
	}

	// Close the server immediately
	err = server1.Close()
	if err != nil {
		t.Fatalf("Failed to close first server: %v", err)
	}

	// Second server: attempt immediate rebind
	// Without SO_REUSEADDR, this would fail with "address already in use"
	server2 := NewServer("127.0.0.1", port)
	err = server2.Listen()
	if err != nil {
		t.Fatalf("Failed to rebind immediately after close: %v", err)
	}
	defer server2.Close()

	// Verify server2 is listening
	if !server2.IsListening() {
		t.Fatal("Expected second server to be listening")
	}

	// Verify we can accept connections on the rebound port
	addr := server2.listener.Addr().(*net.TCPAddr)
	if addr.Port != port {
		t.Errorf("Expected port %d, got %d", port, addr.Port)
	}

	// Test that we can actually connect to the rebound server
	var wg sync.WaitGroup
	var connReceived bool

	wg.Add(1)
	go func() {
		defer wg.Done()
		server2.Accept(func(conn *Connection) {
			connReceived = true
			conn.Close()
			server2.Close()
		})
	}()

	// Give the server time to start accepting
	time.Sleep(50 * time.Millisecond)

	// Connect to verify the server is functional
	conn, err := net.Dial("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		t.Fatalf("Failed to connect to rebound server: %v", err)
	}
	conn.Close()

	wg.Wait()

	if !connReceived {
		t.Error("Expected connection to be received by rebound server")
	}
}
