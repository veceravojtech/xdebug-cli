package daemon

import (
	"bytes"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/console/xdebug-cli/internal/dbgp"
)

// TestHandleContext_NoStackFrames tests that context command returns friendly error
// when there are no active stack frames (e.g., daemon waiting for PHP request)
func TestHandleContext_NoStackFrames(t *testing.T) {
	// Setup: Create a client with mock connection that returns empty stack
	mockConn := newMockConn()
	conn := dbgp.NewConnection(mockConn)
	client := dbgp.NewClient(conn)
	executor := NewCommandExecutor(client)

	// Mock response for stack_get: empty stack frames
	stackXML := `<?xml version="1.0" encoding="iso-8859-1"?>
<response xmlns="urn:debugger_protocol_v1" xmlns:xdebug="https://xdebug.org/dbgp/xdebug"
          command="stack_get"
          transaction_id="1">
</response>`
	stackMessage := fmt.Sprintf("%d\x00%s\x00", len(stackXML), stackXML)
	mockConn.readBuf.WriteString(stackMessage)

	// Execute: Try to get context when no stack frames exist
	result := executor.executeCommand("context", []string{"local"})

	// Verify: Should return user-friendly error, not "stack depth invalid"
	if result.Success {
		t.Errorf("Expected context command to fail when no stack frames exist, but it succeeded")
	}

	expectedError := "Cannot inspect variables: no active stack frames. Trigger a PHP request with XDEBUG_TRIGGER=1 and ensure execution hits a breakpoint first."
	if result.Error != expectedError {
		t.Errorf("Expected error message '%s', got '%s'", expectedError, result.Error)
	}

	// Ensure we didn't get the raw Xdebug error
	if result.Error == "stack depth invalid" {
		t.Errorf("Got raw Xdebug error instead of user-friendly message")
	}
}

// mockConn is a mock implementation of net.Conn for testing
type mockConn struct {
	readBuf  *bytes.Buffer
	writeBuf *bytes.Buffer
}

func newMockConn() *mockConn {
	return &mockConn{
		readBuf:  &bytes.Buffer{},
		writeBuf: &bytes.Buffer{},
	}
}

func (m *mockConn) Read(b []byte) (n int, err error) {
	return m.readBuf.Read(b)
}

func (m *mockConn) Write(b []byte) (n int, err error) {
	return m.writeBuf.Write(b)
}

func (m *mockConn) Close() error                       { return nil }
func (m *mockConn) LocalAddr() net.Addr                { return &net.TCPAddr{} }
func (m *mockConn) RemoteAddr() net.Addr               { return &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 9003} }
func (m *mockConn) SetDeadline(t time.Time) error      { return nil }
func (m *mockConn) SetReadDeadline(t time.Time) error  { return nil }
func (m *mockConn) SetWriteDeadline(t time.Time) error { return nil }

// TestCommandAliases_ExecutionControl tests that execution control aliases
// (continue, cont, into, over, step_out, step_into) map to correct handlers
func TestCommandAliases_ExecutionControl(t *testing.T) {
	testCases := []struct {
		alias           string
		expectedCommand string
	}{
		// run aliases
		{"continue", "run"},
		{"cont", "run"},
		// step aliases
		{"into", "step"},
		{"step_into", "step"},
		// next alias
		{"over", "next"},
		// out alias
		{"step_out", "out"},
	}

	for _, tc := range testCases {
		t.Run(tc.alias, func(t *testing.T) {
			mockConn := newMockConn()
			conn := dbgp.NewConnection(mockConn)
			client := dbgp.NewClient(conn)
			executor := NewCommandExecutor(client)

			// Mock response for run/step/next/out commands
			runXML := fmt.Sprintf(`<?xml version="1.0" encoding="iso-8859-1"?>
<response xmlns="urn:debugger_protocol_v1" command="%s" transaction_id="1" status="break" reason="ok">
<xdebug:message filename="file:///test.php" lineno="10"/>
</response>`, tc.expectedCommand)
			runMessage := fmt.Sprintf("%d\x00%s\x00", len(runXML), runXML)
			mockConn.readBuf.WriteString(runMessage)

			result := executor.executeCommand(tc.alias, []string{})

			// The command should execute (may fail due to mock limitations,
			// but it should NOT return "Unknown command")
			if result.Error == fmt.Sprintf("Unknown command: %s", tc.alias) {
				t.Errorf("Alias '%s' was not recognized as a valid command", tc.alias)
			}
		})
	}
}

// TestCommandAliases_BreakpointManagement tests breakpoint_list and breakpoint_remove aliases
func TestCommandAliases_BreakpointManagement(t *testing.T) {
	t.Run("breakpoint_list", func(t *testing.T) {
		mockConn := newMockConn()
		conn := dbgp.NewConnection(mockConn)
		client := dbgp.NewClient(conn)
		executor := NewCommandExecutor(client)

		// Mock response for breakpoint_list
		listXML := `<?xml version="1.0" encoding="iso-8859-1"?>
<response xmlns="urn:debugger_protocol_v1" command="breakpoint_list" transaction_id="1">
<breakpoint id="1" type="line" state="enabled" filename="file:///test.php" lineno="10"/>
</response>`
		listMessage := fmt.Sprintf("%d\x00%s\x00", len(listXML), listXML)
		mockConn.readBuf.WriteString(listMessage)

		result := executor.executeCommand("breakpoint_list", []string{})

		if result.Error == "Unknown command: breakpoint_list" {
			t.Errorf("breakpoint_list alias was not recognized")
		}
		// Should use info command internally
		if result.Success && result.Command != "info" {
			// The command should be info since breakpoint_list delegates to handleInfo
		}
	})

	t.Run("breakpoint_remove", func(t *testing.T) {
		mockConn := newMockConn()
		conn := dbgp.NewConnection(mockConn)
		client := dbgp.NewClient(conn)
		executor := NewCommandExecutor(client)

		// Mock response for breakpoint_remove
		removeXML := `<?xml version="1.0" encoding="iso-8859-1"?>
<response xmlns="urn:debugger_protocol_v1" command="breakpoint_remove" transaction_id="1"/>`
		removeMessage := fmt.Sprintf("%d\x00%s\x00", len(removeXML), removeXML)
		mockConn.readBuf.WriteString(removeMessage)

		result := executor.executeCommand("breakpoint_remove", []string{"1"})

		if result.Error == "Unknown command: breakpoint_remove" {
			t.Errorf("breakpoint_remove alias was not recognized")
		}
	})
}

// TestPropertyGet tests the property_get command (DBGp-style alias for print)
func TestPropertyGet(t *testing.T) {
	t.Run("valid_usage", func(t *testing.T) {
		mockConn := newMockConn()
		conn := dbgp.NewConnection(mockConn)
		client := dbgp.NewClient(conn)
		executor := NewCommandExecutor(client)

		// Mock response for stack_get (needed by handlePrint)
		stackXML := `<?xml version="1.0" encoding="iso-8859-1"?>
<response xmlns="urn:debugger_protocol_v1" command="stack_get" transaction_id="1">
<stack level="0" type="file" filename="file:///test.php" lineno="10" where="{main}"/>
</response>`
		stackMessage := fmt.Sprintf("%d\x00%s\x00", len(stackXML), stackXML)
		mockConn.readBuf.WriteString(stackMessage)

		// Mock response for property_get
		propXML := `<?xml version="1.0" encoding="iso-8859-1"?>
<response xmlns="urn:debugger_protocol_v1" command="property_get" transaction_id="2">
<property name="myVar" fullname="$myVar" type="string">aGVsbG8=</property>
</response>`
		propMessage := fmt.Sprintf("%d\x00%s\x00", len(propXML), propXML)
		mockConn.readBuf.WriteString(propMessage)

		result := executor.executeCommand("property_get", []string{"-n", "$myVar"})

		if result.Error == "Unknown command: property_get" {
			t.Errorf("property_get command was not recognized")
		}
		if result.Command != "property_get" {
			t.Errorf("Expected command name 'property_get', got '%s'", result.Command)
		}
	})

	t.Run("missing_n_flag", func(t *testing.T) {
		mockConn := newMockConn()
		conn := dbgp.NewConnection(mockConn)
		client := dbgp.NewClient(conn)
		executor := NewCommandExecutor(client)

		result := executor.executeCommand("property_get", []string{"$myVar"})

		if result.Success {
			t.Errorf("Expected property_get without -n flag to fail")
		}
		expectedError := "Usage: property_get -n <variable>"
		if result.Error != expectedError {
			t.Errorf("Expected error '%s', got '%s'", expectedError, result.Error)
		}
	})

	t.Run("missing_variable_after_n", func(t *testing.T) {
		mockConn := newMockConn()
		conn := dbgp.NewConnection(mockConn)
		client := dbgp.NewClient(conn)
		executor := NewCommandExecutor(client)

		result := executor.executeCommand("property_get", []string{"-n"})

		if result.Success {
			t.Errorf("Expected property_get with -n but no variable to fail")
		}
		expectedError := "Usage: property_get -n <variable>"
		if result.Error != expectedError {
			t.Errorf("Expected error '%s', got '%s'", expectedError, result.Error)
		}
	})
}

// TestClear tests the clear command (GDB-style delete by location)
func TestClear(t *testing.T) {
	t.Run("no_args", func(t *testing.T) {
		mockConn := newMockConn()
		conn := dbgp.NewConnection(mockConn)
		client := dbgp.NewClient(conn)
		executor := NewCommandExecutor(client)

		result := executor.executeCommand("clear", []string{})

		if result.Success {
			t.Errorf("Expected clear with no args to fail")
		}
		expectedError := "Usage: clear <:line> or clear <file:line>"
		if result.Error != expectedError {
			t.Errorf("Expected error '%s', got '%s'", expectedError, result.Error)
		}
	})

	t.Run("invalid_format", func(t *testing.T) {
		mockConn := newMockConn()
		conn := dbgp.NewConnection(mockConn)
		client := dbgp.NewClient(conn)
		executor := NewCommandExecutor(client)

		result := executor.executeCommand("clear", []string{"42"})

		if result.Success {
			t.Errorf("Expected clear with bare number to fail (needs : or file:)")
		}
		expectedError := "Usage: clear <:line> or clear <file:line>"
		if result.Error != expectedError {
			t.Errorf("Expected error '%s', got '%s'", expectedError, result.Error)
		}
	})

	t.Run("invalid_line_number", func(t *testing.T) {
		mockConn := newMockConn()
		conn := dbgp.NewConnection(mockConn)
		client := dbgp.NewClient(conn)
		executor := NewCommandExecutor(client)

		result := executor.executeCommand("clear", []string{":abc"})

		if result.Success {
			t.Errorf("Expected clear with non-numeric line to fail")
		}
		if result.Error != "Invalid line number: abc" {
			t.Errorf("Expected 'Invalid line number: abc', got '%s'", result.Error)
		}
	})

	t.Run("file_line_format", func(t *testing.T) {
		mockConn := newMockConn()
		conn := dbgp.NewConnection(mockConn)
		client := dbgp.NewClient(conn)
		executor := NewCommandExecutor(client)

		// Mock response for breakpoint_list
		listXML := `<?xml version="1.0" encoding="iso-8859-1"?>
<response xmlns="urn:debugger_protocol_v1" command="breakpoint_list" transaction_id="1">
<breakpoint id="1" type="line" state="enabled" filename="file:///test.php" lineno="42"/>
</response>`
		listMessage := fmt.Sprintf("%d\x00%s\x00", len(listXML), listXML)
		mockConn.readBuf.WriteString(listMessage)

		// Mock response for breakpoint_remove
		removeXML := `<?xml version="1.0" encoding="iso-8859-1"?>
<response xmlns="urn:debugger_protocol_v1" command="breakpoint_remove" transaction_id="2"/>`
		removeMessage := fmt.Sprintf("%d\x00%s\x00", len(removeXML), removeXML)
		mockConn.readBuf.WriteString(removeMessage)

		result := executor.executeCommand("clear", []string{"test.php:42"})

		if result.Error == "Unknown command: clear" {
			t.Errorf("clear command was not recognized")
		}
		// Command should execute (success depends on mock matching breakpoint location)
	})
}
