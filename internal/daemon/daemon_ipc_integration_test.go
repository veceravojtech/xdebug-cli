package daemon

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"testing"
	"time"

	"github.com/console/xdebug-cli/internal/dbgp"
	"github.com/console/xdebug-cli/internal/ipc"
)

// TestDaemon_IPC_Integration_ServerCreation tests that IPC server is created and started
// This validates Task 4.1: Integrate IPC Server with Listen Flow
//
// This test verifies:
// 1. IPC server is created when daemon starts
// 2. IPC server is running in background
// 3. Socket file is created
// 4. IPC client can connect and execute commands
// 5. Multiple concurrent IPC requests work correctly
func TestDaemon_IPC_Integration_ServerCreation(t *testing.T) {
	tempDir := t.TempDir()
	os.Setenv("HOME", tempDir)
	defer os.Unsetenv("HOME")

	// Create DBGp server
	server := dbgp.NewServer("127.0.0.1", 9020)
	if err := server.Listen(); err != nil {
		t.Fatalf("Failed to start DBGp server: %v", err)
	}
	defer server.Close()

	// Create daemon
	daemon, err := NewDaemon(server, 9020)
	if err != nil {
		t.Fatalf("NewDaemon() error = %v", err)
	}

	// Create mock DBGp connection
	mockConn := createMockConnection()
	client := dbgp.NewClient(mockConn)

	// Initialize client
	_, err = client.Init()
	if err != nil {
		t.Fatalf("Failed to initialize client: %v", err)
	}

	// Start daemon - this should create and start IPC server
	if err := daemon.Start(client); err != nil {
		t.Fatalf("Failed to start daemon: %v", err)
	}

	// Give IPC server time to start
	time.Sleep(100 * time.Millisecond)

	// 1. Verify IPC server was created
	if daemon.ipcServer == nil {
		t.Fatal("IPC server was not created by daemon.Start()")
	}

	// 2. Verify IPC server is running
	if !daemon.ipcServer.IsRunning() {
		t.Fatal("IPC server is not running")
	}

	// 3. Verify socket file exists
	if _, err := os.Stat(daemon.socketPath); os.IsNotExist(err) {
		t.Error("Socket file does not exist")
	}

	// 4. Test command execution via IPC
	t.Run("ExecuteCommands", func(t *testing.T) {
		ipcClient := ipc.NewClient(daemon.socketPath)

		// Send test commands
		commands := []string{"help"}
		response, err := ipcClient.SendCommands(commands, false)
		if err != nil {
			t.Fatalf("Failed to send commands: %v", err)
		}

		if !response.Success {
			t.Errorf("Command execution failed: %s", response.Error)
		}

		if len(response.Results) != 1 {
			t.Errorf("Expected 1 result, got %d", len(response.Results))
		}

		if response.Results[0].Command != "help" {
			t.Errorf("Expected command 'help', got '%s'", response.Results[0].Command)
		}
	})

	// 5. Test concurrent IPC requests
	t.Run("ConcurrentAttach", func(t *testing.T) {
		done := make(chan bool, 2)
		errors := make(chan error, 2)

		// Simulate two concurrent attach commands
		for i := 0; i < 2; i++ {
			go func(id int) {
				ipcClient := ipc.NewClient(daemon.socketPath)
				commands := []string{"help"}
				response, err := ipcClient.SendCommands(commands, false)
				if err != nil {
					errors <- fmt.Errorf("client %d: %w", id, err)
					return
				}
				if !response.Success {
					errors <- fmt.Errorf("client %d: command failed: %s", id, response.Error)
					return
				}
				done <- true
			}(i)
		}

		// Wait for both to complete
		completed := 0
		for completed < 2 {
			select {
			case <-done:
				completed++
			case err := <-errors:
				t.Errorf("Concurrent request failed: %v", err)
				completed++
			case <-time.After(2 * time.Second):
				t.Fatal("Concurrent requests timed out")
			}
		}
	})

	// Cleanup
	daemon.Shutdown()
}

// TestDaemon_IPC_Integration_NoSession validates error handling when no session is active
func TestDaemon_IPC_Integration_NoSession(t *testing.T) {
	tempDir := t.TempDir()
	os.Setenv("HOME", tempDir)
	defer os.Unsetenv("HOME")

	server := dbgp.NewServer("127.0.0.1", 9021)
	if err := server.Listen(); err != nil {
		t.Fatalf("Failed to start DBGp server: %v", err)
	}
	defer server.Close()

	daemon, err := NewDaemon(server, 9021)
	if err != nil {
		t.Fatalf("NewDaemon() error = %v", err)
	}

	// Create IPC server manually without calling Start (no client)
	daemon.ipcServer = ipc.NewServer(daemon.socketPath, daemon.handleIPCRequest)
	if err := daemon.ipcServer.Listen(); err != nil {
		t.Fatalf("Failed to start IPC server: %v", err)
	}
	defer daemon.ipcServer.Shutdown()

	go daemon.ipcServer.Serve()
	time.Sleep(50 * time.Millisecond)

	// Try to execute commands without an active session
	client := ipc.NewClient(daemon.socketPath)
	commands := []string{"run"}
	response, err := client.SendCommands(commands, false)
	if err != nil {
		t.Fatalf("Failed to send commands: %v", err)
	}

	// Should fail with "no active debug session" error
	if response.Success {
		t.Error("Expected command to fail without active session")
	}

	if response.Error != "no active debug session" {
		t.Errorf("Expected 'no active debug session' error, got: %s", response.Error)
	}
}

// TestDaemon_IPC_Integration_KillRequest tests the kill request via IPC
func TestDaemon_IPC_Integration_KillRequest(t *testing.T) {
	tempDir := t.TempDir()
	os.Setenv("HOME", tempDir)
	defer os.Unsetenv("HOME")

	server := dbgp.NewServer("127.0.0.1", 9022)
	if err := server.Listen(); err != nil {
		t.Fatalf("Failed to start DBGp server: %v", err)
	}
	defer server.Close()

	daemon, err := NewDaemon(server, 9022)
	if err != nil {
		t.Fatalf("NewDaemon() error = %v", err)
	}

	// Create mock connection and client
	mockConn := createMockConnection()
	client := dbgp.NewClient(mockConn)
	_, err = client.Init()
	if err != nil {
		t.Fatalf("Failed to initialize client: %v", err)
	}

	// Start daemon
	if err := daemon.Start(client); err != nil {
		t.Fatalf("Failed to start daemon: %v", err)
	}
	time.Sleep(100 * time.Millisecond)

	// Send kill request via IPC
	ipcClient := ipc.NewClient(daemon.socketPath)
	response, err := ipcClient.Kill()
	if err != nil {
		t.Fatalf("Failed to send kill request: %v", err)
	}

	if !response.Success {
		t.Errorf("Kill request failed: %s", response.Error)
	}

	// Give daemon time to shutdown
	time.Sleep(300 * time.Millisecond)

	// Verify daemon shut down
	if _, err := os.Stat(daemon.pidFile); !os.IsNotExist(err) {
		t.Error("PID file still exists after kill")
	}

	if _, err := os.Stat(daemon.socketPath); !os.IsNotExist(err) {
		t.Error("Socket file still exists after kill")
	}

	// Verify context cancelled
	select {
	case <-daemon.ctx.Done():
		// Success
	default:
		t.Error("Context not cancelled after kill")
	}
}

// createMockConnection creates a mock network connection for testing
func createMockConnection() *dbgp.Connection {
	// Create a pipe for bidirectional communication
	server, client := net.Pipe()

	// Create the DBGp connection wrapper
	conn := dbgp.NewConnection(client)

	// Start a goroutine to simulate Xdebug responses
	go func() {
		defer server.Close()
		reader := bufio.NewReader(server)

		// Send init message first
		initXML := `<?xml version="1.0" encoding="UTF-8"?>
<init xmlns="urn:debugger_protocol_v1" xmlns:xdebug="https://xdebug.org/dbgp/xdebug"
      appid="test" idekey="test" language="PHP" protocol_version="1.0"
      fileuri="file:///test.php">
</init>`
		initMsg := fmt.Sprintf("%d\x00%s\x00", len(initXML), initXML)
		server.Write([]byte(initMsg))

		// Echo back any commands (for now)
		for {
			_, err := reader.ReadString('\x00')
			if err != nil {
				return
			}
			// Send a simple OK response
			responseXML := `<?xml version="1.0" encoding="UTF-8"?>
<response xmlns="urn:debugger_protocol_v1" command="status" status="break"></response>`
			responseMsg := fmt.Sprintf("%d\x00%s\x00", len(responseXML), responseXML)
			server.Write([]byte(responseMsg))
		}
	}()

	return conn
}
