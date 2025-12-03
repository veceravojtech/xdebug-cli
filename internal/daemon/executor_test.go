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
