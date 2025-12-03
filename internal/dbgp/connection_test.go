package dbgp

import (
	"bytes"
	"fmt"
	"net"
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
		name        string
		input       string
		expected    string
		expectError bool
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
			name:        "invalid size",
			input:       "abc\x00content\x00",
			expectError: true,
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
