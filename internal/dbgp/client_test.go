package dbgp

import (
	"encoding/base64"
	"fmt"
	"strings"
	"testing"
)

func TestNewClient(t *testing.T) {
	mockConn := newMockConn()
	conn := NewConnection(mockConn)
	client := NewClient(conn)

	if client == nil {
		t.Fatal("Expected non-nil client")
	}

	if client.conn == nil {
		t.Error("Expected non-nil connection")
	}

	if client.session == nil {
		t.Error("Expected non-nil session")
	}

	if client.session.GetState() != StateNone {
		t.Errorf("Expected initial state StateNone, got %v", client.session.GetState())
	}
}

func TestClient_Init(t *testing.T) {
	initXML := `<?xml version="1.0" encoding="iso-8859-1"?>
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
	message := fmt.Sprintf("%d\x00%s\x00", len(initXML), initXML)
	mockConn.readBuf.WriteString(message)

	conn := NewConnection(mockConn)
	client := NewClient(conn)

	init, err := client.Init()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if init == nil {
		t.Fatal("Expected non-nil init")
	}

	if init.AppID != "12345" {
		t.Errorf("Expected AppID '12345', got '%s'", init.AppID)
	}

	if init.IDEKey != "PHPSTORM" {
		t.Errorf("Expected IDEKey 'PHPSTORM', got '%s'", init.IDEKey)
	}

	// Check session was updated
	if client.session.GetState() != StateStarting {
		t.Errorf("Expected state StateStarting, got %v", client.session.GetState())
	}

	if client.session.GetIDEKey() != "PHPSTORM" {
		t.Errorf("Expected session IDE key 'PHPSTORM', got '%s'", client.session.GetIDEKey())
	}

	if client.session.GetAppID() != "12345" {
		t.Errorf("Expected session app ID '12345', got '%s'", client.session.GetAppID())
	}
}

func TestClient_Run(t *testing.T) {
	responseXML := `<?xml version="1.0" encoding="iso-8859-1"?>
<response xmlns="urn:debugger_protocol_v1" xmlns:xdebug="https://xdebug.org/dbgp/xdebug"
          command="run"
          transaction_id="1"
          status="break"
          reason="ok"
          filename="file:///test.php"
          lineno="10">
</response>`

	mockConn := newMockConn()
	message := fmt.Sprintf("%d\x00%s\x00", len(responseXML), responseXML)
	mockConn.readBuf.WriteString(message)

	conn := NewConnection(mockConn)
	client := NewClient(conn)

	response, err := client.Run()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if response.Command != "run" {
		t.Errorf("Expected command 'run', got '%s'", response.Command)
	}

	if response.Status != "break" {
		t.Errorf("Expected status 'break', got '%s'", response.Status)
	}

	// Check command was sent
	sent := mockConn.writeBuf.String()
	if !strings.Contains(sent, "run -i") {
		t.Errorf("Expected 'run -i' command, got '%s'", sent)
	}

	// Check session state was updated
	if client.session.GetState() != StateBreak {
		t.Errorf("Expected state StateBreak, got %v", client.session.GetState())
	}

	// Check location was updated
	file, line := client.session.GetCurrentLocation()
	if !strings.Contains(file, "test.php") {
		t.Errorf("Expected file to contain 'test.php', got '%s'", file)
	}
	if line != 10 {
		t.Errorf("Expected line 10, got %d", line)
	}
}

func TestClient_Step(t *testing.T) {
	responseXML := `<?xml version="1.0" encoding="iso-8859-1"?>
<response xmlns="urn:debugger_protocol_v1" xmlns:xdebug="https://xdebug.org/dbgp/xdebug"
          command="step_into"
          transaction_id="1"
          status="break"
          reason="ok">
</response>`

	mockConn := newMockConn()
	message := fmt.Sprintf("%d\x00%s\x00", len(responseXML), responseXML)
	mockConn.readBuf.WriteString(message)

	conn := NewConnection(mockConn)
	client := NewClient(conn)

	response, err := client.Step()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if response.Command != "step_into" {
		t.Errorf("Expected command 'step_into', got '%s'", response.Command)
	}

	sent := mockConn.writeBuf.String()
	if !strings.Contains(sent, "step_into -i") {
		t.Errorf("Expected 'step_into -i' command, got '%s'", sent)
	}
}

func TestClient_Next(t *testing.T) {
	responseXML := `<?xml version="1.0" encoding="iso-8859-1"?>
<response xmlns="urn:debugger_protocol_v1" xmlns:xdebug="https://xdebug.org/dbgp/xdebug"
          command="step_over"
          transaction_id="1"
          status="break"
          reason="ok">
</response>`

	mockConn := newMockConn()
	message := fmt.Sprintf("%d\x00%s\x00", len(responseXML), responseXML)
	mockConn.readBuf.WriteString(message)

	conn := NewConnection(mockConn)
	client := NewClient(conn)

	response, err := client.Next()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if response.Command != "step_over" {
		t.Errorf("Expected command 'step_over', got '%s'", response.Command)
	}

	sent := mockConn.writeBuf.String()
	if !strings.Contains(sent, "step_over -i") {
		t.Errorf("Expected 'step_over -i' command, got '%s'", sent)
	}
}

func TestClient_StepOut(t *testing.T) {
	responseXML := `<?xml version="1.0" encoding="iso-8859-1"?>
<response xmlns="urn:debugger_protocol_v1" xmlns:xdebug="https://xdebug.org/dbgp/xdebug"
          command="step_out"
          transaction_id="1"
          status="break"
          reason="ok">
  <xdebug:message filename="file:///test.php" lineno="50"></xdebug:message>
</response>`

	mockConn := newMockConn()
	message := fmt.Sprintf("%d\x00%s\x00", len(responseXML), responseXML)
	mockConn.readBuf.WriteString(message)

	conn := NewConnection(mockConn)
	client := NewClient(conn)

	response, err := client.StepOut()
	if err != nil {
		t.Fatalf("StepOut() failed: %v", err)
	}

	if response.Command != "step_out" {
		t.Errorf("Expected command 'step_out', got '%s'", response.Command)
	}

	sent := mockConn.writeBuf.String()
	if !strings.Contains(sent, "step_out -i") {
		t.Errorf("Expected 'step_out -i' command, got '%s'", sent)
	}
}

func TestClient_Finish(t *testing.T) {
	responseXML := `<?xml version="1.0" encoding="iso-8859-1"?>
<response xmlns="urn:debugger_protocol_v1" xmlns:xdebug="https://xdebug.org/dbgp/xdebug"
          command="stop"
          transaction_id="1"
          status="stopped"
          reason="ok">
</response>`

	mockConn := newMockConn()
	message := fmt.Sprintf("%d\x00%s\x00", len(responseXML), responseXML)
	mockConn.readBuf.WriteString(message)

	conn := NewConnection(mockConn)
	client := NewClient(conn)

	// Session should be set to stopping before sending
	response, err := client.Finish()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if response.Command != "stop" {
		t.Errorf("Expected command 'stop', got '%s'", response.Command)
	}

	sent := mockConn.writeBuf.String()
	if !strings.Contains(sent, "stop -i") {
		t.Errorf("Expected 'stop -i' command, got '%s'", sent)
	}

	// Check session state
	if client.session.GetState() != StateStopped {
		t.Errorf("Expected state StateStopped, got %v", client.session.GetState())
	}
}

func TestClient_SetBreakpoint(t *testing.T) {
	responseXML := `<?xml version="1.0" encoding="iso-8859-1"?>
<response xmlns="urn:debugger_protocol_v1" xmlns:xdebug="https://xdebug.org/dbgp/xdebug"
          command="breakpoint_set"
          transaction_id="1"
          id="1">
</response>`

	tests := []struct {
		name      string
		file      string
		line      int
		condition string
		contains  string
	}{
		{
			name:     "simple breakpoint",
			file:     "file:///test.php",
			line:     10,
			contains: "breakpoint_set -i 1 -t line -f file:///test.php -n 10",
		},
		{
			name:      "breakpoint with condition",
			file:      "file:///test.php",
			line:      20,
			condition: "$x > 5",
			contains:  "breakpoint_set -i 1 -t line -f file:///test.php -n 20 --",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockConn := newMockConn()
			message := fmt.Sprintf("%d\x00%s\x00", len(responseXML), responseXML)
			mockConn.readBuf.WriteString(message)

			conn := NewConnection(mockConn)
			client := NewClient(conn)

			response, err := client.SetBreakpoint(tt.file, tt.line, tt.condition)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if response.Command != "breakpoint_set" {
				t.Errorf("Expected command 'breakpoint_set', got '%s'", response.Command)
			}

			sent := mockConn.writeBuf.String()
			if !strings.Contains(sent, tt.contains) {
				t.Errorf("Expected command to contain '%s', got '%s'", tt.contains, sent)
			}
		})
	}
}

func TestClient_SetBreakpointToCall(t *testing.T) {
	responseXML := `<?xml version="1.0" encoding="iso-8859-1"?>
<response xmlns="urn:debugger_protocol_v1" xmlns:xdebug="https://xdebug.org/dbgp/xdebug"
          command="breakpoint_set"
          transaction_id="1"
          id="2">
</response>`

	mockConn := newMockConn()
	message := fmt.Sprintf("%d\x00%s\x00", len(responseXML), responseXML)
	mockConn.readBuf.WriteString(message)

	conn := NewConnection(mockConn)
	client := NewClient(conn)

	response, err := client.SetBreakpointToCall("myFunction")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if response.Command != "breakpoint_set" {
		t.Errorf("Expected command 'breakpoint_set', got '%s'", response.Command)
	}

	sent := mockConn.writeBuf.String()
	if !strings.Contains(sent, "breakpoint_set -i 1 -t call -m myFunction") {
		t.Errorf("Expected call breakpoint command, got '%s'", sent)
	}
}

func TestClient_GetBreakpointList(t *testing.T) {
	responseXML := `<?xml version="1.0" encoding="iso-8859-1"?>
<response xmlns="urn:debugger_protocol_v1" xmlns:xdebug="https://xdebug.org/dbgp/xdebug"
          command="breakpoint_list"
          transaction_id="1">
    <breakpoint id="1" type="line" state="enabled" filename="file:///test.php" lineno="10"></breakpoint>
</response>`

	mockConn := newMockConn()
	message := fmt.Sprintf("%d\x00%s\x00", len(responseXML), responseXML)
	mockConn.readBuf.WriteString(message)

	conn := NewConnection(mockConn)
	client := NewClient(conn)

	response, err := client.GetBreakpointList()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if response.Command != "breakpoint_list" {
		t.Errorf("Expected command 'breakpoint_list', got '%s'", response.Command)
	}

	if len(response.Breakpoints) != 1 {
		t.Errorf("Expected 1 breakpoint, got %d", len(response.Breakpoints))
	}
}

func TestClient_GetProperty(t *testing.T) {
	responseXML := `<?xml version="1.0" encoding="iso-8859-1"?>
<response xmlns="urn:debugger_protocol_v1" xmlns:xdebug="https://xdebug.org/dbgp/xdebug"
          command="property_get"
          transaction_id="1">
    <property name="var1" fullname="$var1" type="string" size="5" encoding="base64">
        <![CDATA[aGVsbG8=]]>
    </property>
</response>`

	mockConn := newMockConn()
	message := fmt.Sprintf("%d\x00%s\x00", len(responseXML), responseXML)
	mockConn.readBuf.WriteString(message)

	conn := NewConnection(mockConn)
	client := NewClient(conn)

	response, err := client.GetProperty("$var1")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if response.Command != "property_get" {
		t.Errorf("Expected command 'property_get', got '%s'", response.Command)
	}

	sent := mockConn.writeBuf.String()
	if !strings.Contains(sent, "property_get -i 1 -d 0 -n $var1") {
		t.Errorf("Expected property_get command, got '%s'", sent)
	}
}

func TestClient_GetContext(t *testing.T) {
	responseXML := `<?xml version="1.0" encoding="iso-8859-1"?>
<response xmlns="urn:debugger_protocol_v1" xmlns:xdebug="https://xdebug.org/dbgp/xdebug"
          command="context_get"
          transaction_id="1"
          context="0">
    <property name="var1" fullname="$var1" type="string"></property>
</response>`

	mockConn := newMockConn()
	message := fmt.Sprintf("%d\x00%s\x00", len(responseXML), responseXML)
	mockConn.readBuf.WriteString(message)

	conn := NewConnection(mockConn)
	client := NewClient(conn)

	response, err := client.GetContext(0)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if response.Command != "context_get" {
		t.Errorf("Expected command 'context_get', got '%s'", response.Command)
	}

	sent := mockConn.writeBuf.String()
	if !strings.Contains(sent, "context_get -i 1 -d 0 -c 0") {
		t.Errorf("Expected context_get command, got '%s'", sent)
	}
}

func TestClient_GetContextNames(t *testing.T) {
	responseXML := `<?xml version="1.0" encoding="iso-8859-1"?>
<response xmlns="urn:debugger_protocol_v1" xmlns:xdebug="https://xdebug.org/dbgp/xdebug"
          command="context_names"
          transaction_id="1">
    <context name="Local" id="0"></context>
    <context name="Global" id="1"></context>
</response>`

	mockConn := newMockConn()
	message := fmt.Sprintf("%d\x00%s\x00", len(responseXML), responseXML)
	mockConn.readBuf.WriteString(message)

	conn := NewConnection(mockConn)
	client := NewClient(conn)

	response, err := client.GetContextNames()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if response.Command != "context_names" {
		t.Errorf("Expected command 'context_names', got '%s'", response.Command)
	}

	if len(response.Contexts) != 2 {
		t.Errorf("Expected 2 contexts, got %d", len(response.Contexts))
	}
}

func TestClient_Eval(t *testing.T) {
	responseXML := `<?xml version="1.0" encoding="iso-8859-1"?>
<response xmlns="urn:debugger_protocol_v1" xmlns:xdebug="https://xdebug.org/dbgp/xdebug"
          command="eval"
          transaction_id="1">
    <property type="int"><![CDATA[42]]></property>
</response>`

	mockConn := newMockConn()
	message := fmt.Sprintf("%d\x00%s\x00", len(responseXML), responseXML)
	mockConn.readBuf.WriteString(message)

	conn := NewConnection(mockConn)
	client := NewClient(conn)

	response, err := client.Eval("$x + 1")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if response.Command != "eval" {
		t.Errorf("Expected command 'eval', got '%s'", response.Command)
	}

	sent := mockConn.writeBuf.String()
	// Check that the expression was base64 encoded
	expectedEncoded := base64.StdEncoding.EncodeToString([]byte("$x + 1"))
	if !strings.Contains(sent, expectedEncoded) {
		t.Errorf("Expected base64-encoded expression in command, got '%s'", sent)
	}
}

func TestDecodePropertyValue(t *testing.T) {
	tests := []struct {
		name        string
		property    ProtocolProperty
		expected    string
		expectError bool
	}{
		{
			name: "base64 encoded",
			property: ProtocolProperty{
				Encoding: "base64",
				Value:    "aGVsbG8=",
			},
			expected: "hello",
		},
		{
			name: "not encoded",
			property: ProtocolProperty{
				Encoding: "",
				Value:    "hello",
			},
			expected: "hello",
		},
		{
			name: "invalid base64",
			property: ProtocolProperty{
				Encoding: "base64",
				Value:    "invalid!!!",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := DecodePropertyValue(&tt.property)

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

func TestClient_GetSession(t *testing.T) {
	mockConn := newMockConn()
	conn := NewConnection(mockConn)
	client := NewClient(conn)

	session := client.GetSession()
	if session == nil {
		t.Error("Expected non-nil session")
	}
	if session != client.session {
		t.Error("Expected GetSession to return client's session")
	}
}

func TestClient_Close(t *testing.T) {
	mockConn := newMockConn()
	conn := NewConnection(mockConn)
	client := NewClient(conn)

	err := client.Close()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}
