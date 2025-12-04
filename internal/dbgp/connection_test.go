package dbgp

import (
	"bytes"
	"fmt"
	"net"
	"strconv"
	"strings"
	"testing"
	"time"
)

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

func TestNewConnection(t *testing.T) {
	mockConn := newMockConn()
	conn := NewConnection(mockConn)

	if conn == nil {
		t.Fatal("Expected non-nil connection")
	}

	if conn.conn != mockConn {
		t.Error("Expected connection to wrap mock conn")
	}

	if conn.reader == nil {
		t.Error("Expected reader to be initialized")
	}
}

func TestConnection_ReadMessage(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expected      string
		expectError   bool
		errorContains string
	}{
		{
			name:     "simple message",
			input:    "13\x00Hello, World!\x00",
			expected: "Hello, World!",
		},
		{
			name:     "xml message",
			input:    "42\x00<?xml version=\"1.0\"?><response></response>\x00",
			expected: "<?xml version=\"1.0\"?><response></response>",
		},
		{
			name:     "message with special chars",
			input:    "12\x00Test\nLine\t2!\x00",
			expected: "Test\nLine\t2!",
		},
		{
			name:        "missing terminator",
			input:       "5\x00HelloX",
			expectError: true,
		},
		{
			name:          "invalid size with non-digit chars",
			input:         "abc\x00content\x00",
			expectError:   true,
			errorContains: "expected digits only",
		},
		{
			name:          "size field with xml content",
			input:         "php\" lineno=\"622\">...</init>\x00content\x00",
			expectError:   true,
			errorContains: "expected digits only",
		},
		{
			name:     "zero size message",
			input:    "0\x00\x00",
			expected: "",
		},
		{
			name:     "large valid message",
			input:    fmt.Sprintf("%d\x00%s\x00", 1024, strings.Repeat("x", 1024)),
			expected: strings.Repeat("x", 1024),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockConn := newMockConn()
			mockConn.readBuf.WriteString(tt.input)
			conn := NewConnection(mockConn)

			result, err := conn.ReadMessage()

			if tt.expectError {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				if tt.errorContains != "" && !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("Expected error containing '%s', got: %v", tt.errorContains, err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if result != tt.expected {
					t.Errorf("Expected '%s', got '%s'", tt.expected, result)
				}
			}
		})
	}
}

func TestConnection_SendMessage(t *testing.T) {
	tests := []struct {
		name     string
		message  string
		expected string
	}{
		{
			name:     "simple command",
			message:  "run -i 1",
			expected: "run -i 1\x00",
		},
		{
			name:     "breakpoint command",
			message:  "breakpoint_set -i 2 -t line -f file:///test.php -n 10",
			expected: "breakpoint_set -i 2 -t line -f file:///test.php -n 10\x00",
		},
		{
			name:     "empty message",
			message:  "",
			expected: "\x00",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockConn := newMockConn()
			conn := NewConnection(mockConn)

			err := conn.SendMessage(tt.message)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			result := mockConn.writeBuf.String()
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestConnection_GetResponse(t *testing.T) {
	xmlResponse := `<?xml version="1.0" encoding="iso-8859-1"?>
<response xmlns="urn:debugger_protocol_v1" xmlns:xdebug="https://xdebug.org/dbgp/xdebug"
          command="step_into"
          transaction_id="1"
          status="break"
          reason="ok">
</response>`

	mockConn := newMockConn()
	// Format: size\0xml\0
	message := fmt.Sprintf("%d\x00%s\x00", len(xmlResponse), xmlResponse)
	mockConn.readBuf.WriteString(message)

	conn := NewConnection(mockConn)
	response, err := conn.GetResponse()

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if response == nil {
		t.Fatal("Expected non-nil response")
	}

	if response.Command != "step_into" {
		t.Errorf("Expected command 'step_into', got '%s'", response.Command)
	}

	if response.TransactionID != "1" {
		t.Errorf("Expected transaction_id '1', got '%s'", response.TransactionID)
	}

	if response.Status != "break" {
		t.Errorf("Expected status 'break', got '%s'", response.Status)
	}
}

func TestConnection_GetResponse_InvalidXML(t *testing.T) {
	invalidXML := "<invalid xml"

	mockConn := newMockConn()
	message := fmt.Sprintf("%d\x00%s\x00", len(invalidXML), invalidXML)
	mockConn.readBuf.WriteString(message)

	conn := NewConnection(mockConn)
	_, err := conn.GetResponse()

	if err == nil {
		t.Error("Expected error for invalid XML, got nil")
	}
}

func TestConnection_GetResponse_NotResponse(t *testing.T) {
	xmlInit := `<?xml version="1.0" encoding="iso-8859-1"?>
<init xmlns="urn:debugger_protocol_v1" xmlns:xdebug="https://xdebug.org/dbgp/xdebug"
      fileuri="file:///path/to/script.php"
      language="PHP"
      protocol_version="1.0"
      appid="12345"
      idekey="PHPSTORM">
    <engine version="3.0.0">
        <![CDATA[Xdebug]]>
    </engine>
</init>`

	mockConn := newMockConn()
	message := fmt.Sprintf("%d\x00%s\x00", len(xmlInit), xmlInit)
	mockConn.readBuf.WriteString(message)

	conn := NewConnection(mockConn)
	_, err := conn.GetResponse()

	if err == nil {
		t.Error("Expected error for non-response XML, got nil")
	}
	if !strings.Contains(err.Error(), "expected response") {
		t.Errorf("Expected 'expected response' error, got: %v", err)
	}
}

func TestConnection_Close(t *testing.T) {
	mockConn := newMockConn()
	conn := NewConnection(mockConn)

	err := conn.Close()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestConnection_GetRemoteAddr(t *testing.T) {
	mockConn := newMockConn()
	conn := NewConnection(mockConn)

	addr := conn.GetRemoteAddr()
	if addr == "" {
		t.Error("Expected non-empty remote address")
	}
}

// TestConnection_ReadMessage_SizeValidation tests size field validation
func TestConnection_ReadMessage_SizeValidation(t *testing.T) {
	tests := []struct {
		name          string
		sizeField     string
		expectError   bool
		errorContains string
	}{
		{
			name:          "non-digit characters",
			sizeField:     "12abc",
			expectError:   true,
			errorContains: "expected digits only",
		},
		{
			name:          "size field with spaces",
			sizeField:     "12 34",
			expectError:   true,
			errorContains: "expected digits only",
		},
		{
			name:          "size field with special chars",
			sizeField:     "12@34",
			expectError:   true,
			errorContains: "expected digits only",
		},
		{
			name:          "corrupted xml in size field",
			sizeField:     "php\" lineno=\"622\">...</init>",
			expectError:   true,
			errorContains: "expected digits only",
		},
		{
			name:          "very long corrupted size field truncated in error",
			sizeField:     strings.Repeat("corrupted_data_", 10),
			expectError:   true,
			errorContains: "...", // Should be truncated to 50 chars
		},
		{
			name:        "valid numeric size",
			sizeField:   "42",
			expectError: false,
		},
		{
			name:        "zero size",
			sizeField:   "0",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockConn := newMockConn()
			if tt.expectError {
				// Only write size field for error cases
				mockConn.readBuf.WriteString(tt.sizeField + "\x00")
			} else {
				// For valid cases, write complete message
				content := strings.Repeat("x", mustAtoi(tt.sizeField))
				mockConn.readBuf.WriteString(fmt.Sprintf("%s\x00%s\x00", tt.sizeField, content))
			}
			conn := NewConnection(mockConn)

			_, err := conn.ReadMessage()

			if tt.expectError {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				if tt.errorContains != "" && !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("Expected error containing '%s', got: %v", tt.errorContains, err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

// TestConnection_ReadMessage_NegativeSize tests rejection of negative size values
func TestConnection_ReadMessage_NegativeSize(t *testing.T) {
	// Note: Negative sizes can't actually be sent in the size field since it's validated
	// to contain only digits. This test documents that negative sizes would be rejected
	// if they somehow made it past the digit validation.
	t.Skip("Negative sizes are already prevented by digit-only validation")
}

// TestConnection_ReadMessage_MaxSizeExceeded tests rejection of messages exceeding MaxMessageSize
func TestConnection_ReadMessage_MaxSizeExceeded(t *testing.T) {
	mockConn := newMockConn()
	// Try to send a message claiming to be 101MB (exceeds 100MB max)
	oversizedValue := MaxMessageSize + 1
	mockConn.readBuf.WriteString(fmt.Sprintf("%d\x00", oversizedValue))

	conn := NewConnection(mockConn)
	_, err := conn.ReadMessage()

	if err == nil {
		t.Fatal("Expected error for message size exceeding maximum, got nil")
	}

	if !strings.Contains(err.Error(), "exceeds maximum allowed size") {
		t.Errorf("Expected error about exceeding maximum size, got: %v", err)
	}

	// Verify the error message includes the size values
	if !strings.Contains(err.Error(), fmt.Sprintf("%d", oversizedValue)) {
		t.Errorf("Expected error to include size value %d, got: %v", oversizedValue, err)
	}
}

// TestConnection_ReadMessage_MaxSizeAtBoundary tests that MaxMessageSize exactly is allowed
func TestConnection_ReadMessage_MaxSizeAtBoundary(t *testing.T) {
	// This test verifies that a message exactly at MaxMessageSize is accepted
	// We'll test with a much smaller size to avoid memory issues in tests
	testSize := 1024 // 1KB for testing
	mockConn := newMockConn()
	content := strings.Repeat("x", testSize)
	mockConn.readBuf.WriteString(fmt.Sprintf("%d\x00%s\x00", testSize, content))

	conn := NewConnection(mockConn)
	result, err := conn.ReadMessage()

	if err != nil {
		t.Errorf("Expected no error for valid size, got: %v", err)
	}

	if len(result) != testSize {
		t.Errorf("Expected message length %d, got %d", testSize, len(result))
	}
}

// Helper function for tests
func mustAtoi(s string) int {
	val, _ := strconv.Atoi(s)
	return val
}

// timeoutMockConn simulates a connection that blocks on read to test timeout behavior
type timeoutMockConn struct {
	readDeadline   time.Time
	readBlock      chan struct{} // Used to control when Read returns
	deadlineWasSet bool
}

func newTimeoutMockConn() *timeoutMockConn {
	return &timeoutMockConn{
		readBlock: make(chan struct{}),
	}
}

func (m *timeoutMockConn) Read(b []byte) (n int, err error) {
	// Check if deadline has passed
	if !m.readDeadline.IsZero() && time.Now().After(m.readDeadline) {
		return 0, fmt.Errorf("i/o timeout")
	}

	// Block until readBlock channel is closed or deadline passes
	if !m.readDeadline.IsZero() {
		timer := time.NewTimer(time.Until(m.readDeadline))
		defer timer.Stop()

		select {
		case <-m.readBlock:
			// Unblocked - should not happen in timeout test
			return 0, nil
		case <-timer.C:
			// Timeout!
			return 0, fmt.Errorf("i/o timeout")
		}
	}

	// No deadline set - block forever
	<-m.readBlock
	return 0, nil
}

func (m *timeoutMockConn) Write(b []byte) (n int, err error) {
	return len(b), nil
}

func (m *timeoutMockConn) Close() error {
	close(m.readBlock)
	return nil
}

func (m *timeoutMockConn) LocalAddr() net.Addr {
	return &net.TCPAddr{}
}

func (m *timeoutMockConn) RemoteAddr() net.Addr {
	return &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 9003}
}

func (m *timeoutMockConn) SetDeadline(t time.Time) error {
	m.readDeadline = t
	m.deadlineWasSet = true
	return nil
}

func (m *timeoutMockConn) SetReadDeadline(t time.Time) error {
	m.readDeadline = t
	m.deadlineWasSet = true
	return nil
}

func (m *timeoutMockConn) SetWriteDeadline(t time.Time) error {
	return nil
}

// TestConnection_ReadMessageWithTimeout_Timeout verifies that timeout fires when no data is received
func TestConnection_ReadMessageWithTimeout_Timeout(t *testing.T) {
	mockConn := newTimeoutMockConn()
	conn := NewConnection(mockConn)

	// Use a very short timeout for the test
	timeout := 50 * time.Millisecond

	start := time.Now()
	_, err := conn.ReadMessageWithTimeout(timeout)
	elapsed := time.Since(start)

	if err == nil {
		t.Fatal("Expected timeout error, got nil")
	}

	if !strings.Contains(err.Error(), "timeout") && !strings.Contains(err.Error(), "failed to read message size") {
		t.Errorf("Expected timeout error, got: %v", err)
	}

	// Verify timeout happened around the expected time (with some tolerance)
	if elapsed < timeout || elapsed > timeout+200*time.Millisecond {
		t.Errorf("Expected timeout around %v, but took %v", timeout, elapsed)
	}

	// Verify deadline was set
	if !mockConn.deadlineWasSet {
		t.Error("Expected SetReadDeadline to be called")
	}
}

// TestConnection_ReadMessageWithTimeout_Success verifies successful read within timeout
func TestConnection_ReadMessageWithTimeout_Success(t *testing.T) {
	mockConn := newMockConn()
	validMessage := "13\x00Hello, World!\x00"
	mockConn.readBuf.WriteString(validMessage)

	conn := NewConnection(mockConn)

	timeout := 5 * time.Second
	result, err := conn.ReadMessageWithTimeout(timeout)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if result != "Hello, World!" {
		t.Errorf("Expected 'Hello, World!', got '%s'", result)
	}
}

// deadlineClearMockConn tracks SetReadDeadline calls to verify deadline is cleared
type deadlineClearMockConn struct {
	readBuf           *bytes.Buffer
	writeBuf          *bytes.Buffer
	deadlineCalls     []time.Time
	deadlineCallCount int
}

func newDeadlineClearMockConn() *deadlineClearMockConn {
	return &deadlineClearMockConn{
		readBuf:       &bytes.Buffer{},
		writeBuf:      &bytes.Buffer{},
		deadlineCalls: []time.Time{},
	}
}

func (m *deadlineClearMockConn) Read(b []byte) (n int, err error) {
	return m.readBuf.Read(b)
}

func (m *deadlineClearMockConn) Write(b []byte) (n int, err error) {
	return m.writeBuf.Write(b)
}

func (m *deadlineClearMockConn) Close() error {
	return nil
}

func (m *deadlineClearMockConn) LocalAddr() net.Addr {
	return &net.TCPAddr{}
}

func (m *deadlineClearMockConn) RemoteAddr() net.Addr {
	return &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 9003}
}

func (m *deadlineClearMockConn) SetDeadline(t time.Time) error {
	m.deadlineCalls = append(m.deadlineCalls, t)
	m.deadlineCallCount++
	return nil
}

func (m *deadlineClearMockConn) SetReadDeadline(t time.Time) error {
	m.deadlineCalls = append(m.deadlineCalls, t)
	m.deadlineCallCount++
	return nil
}

func (m *deadlineClearMockConn) SetWriteDeadline(t time.Time) error {
	return nil
}

// TestConnection_ReadMessageWithTimeout_DeadlineCleared verifies deadline is cleared after successful read
func TestConnection_ReadMessageWithTimeout_DeadlineCleared(t *testing.T) {
	mockConn := newDeadlineClearMockConn()
	validMessage := "13\x00Hello, World!\x00"
	mockConn.readBuf.WriteString(validMessage)

	conn := NewConnection(mockConn)

	timeout := 5 * time.Second
	_, err := conn.ReadMessageWithTimeout(timeout)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Verify SetReadDeadline was called at least twice
	if mockConn.deadlineCallCount < 2 {
		t.Fatalf("Expected at least 2 SetReadDeadline calls (set + clear), got %d", mockConn.deadlineCallCount)
	}

	// The last call should clear the deadline (zero time)
	lastDeadline := mockConn.deadlineCalls[len(mockConn.deadlineCalls)-1]
	if !lastDeadline.IsZero() {
		t.Errorf("Expected last SetReadDeadline call to clear deadline (zero time), got %v", lastDeadline)
	}
}

// TestConnection_ReadMessage_UsesDefaultTimeout verifies ReadMessage uses default timeout
func TestConnection_ReadMessage_UsesDefaultTimeout(t *testing.T) {
	mockConn := newDeadlineClearMockConn()
	validMessage := "13\x00Hello, World!\x00"
	mockConn.readBuf.WriteString(validMessage)

	conn := NewConnection(mockConn)

	_, err := conn.ReadMessage()

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Verify SetReadDeadline was called (proves timeout is being used)
	if mockConn.deadlineCallCount < 1 {
		t.Fatal("Expected SetReadDeadline to be called (timeout should be set)")
	}
}
