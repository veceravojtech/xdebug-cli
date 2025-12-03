package dbgp

import (
	"strings"
	"testing"
)

func TestCreateProtocolFromXML_Init(t *testing.T) {
	xmlData := `<?xml version="1.0" encoding="iso-8859-1"?>
<init xmlns="urn:debugger_protocol_v1" xmlns:xdebug="https://xdebug.org/dbgp/xdebug"
      fileuri="file:///path/to/script.php"
      language="PHP"
      xdebug:language_version="7.4.3"
      protocol_version="1.0"
      appid="12345"
      idekey="PHPSTORM">
    <engine version="3.0.0">
        <![CDATA[Xdebug]]>
    </engine>
</init>`

	result, err := CreateProtocolFromXML(xmlData)
	if err != nil {
		t.Fatalf("Failed to parse init XML: %v", err)
	}

	init, ok := result.(*ProtocolInit)
	if !ok {
		t.Fatalf("Expected *ProtocolInit, got %T", result)
	}

	if init.AppID != "12345" {
		t.Errorf("Expected AppID '12345', got '%s'", init.AppID)
	}

	if init.IDEKey != "PHPSTORM" {
		t.Errorf("Expected IDEKey 'PHPSTORM', got '%s'", init.IDEKey)
	}

	if init.Language != "PHP" {
		t.Errorf("Expected Language 'PHP', got '%s'", init.Language)
	}

	if init.ProtocolV != "1.0" {
		t.Errorf("Expected ProtocolV '1.0', got '%s'", init.ProtocolV)
	}

	if !strings.Contains(init.FileURI, "script.php") {
		t.Errorf("Expected FileURI to contain 'script.php', got '%s'", init.FileURI)
	}
}

func TestCreateProtocolFromXML_Response(t *testing.T) {
	xmlData := `<?xml version="1.0" encoding="iso-8859-1"?>
<response xmlns="urn:debugger_protocol_v1" xmlns:xdebug="https://xdebug.org/dbgp/xdebug"
          command="step_into"
          transaction_id="1"
          status="break"
          reason="ok">
</response>`

	result, err := CreateProtocolFromXML(xmlData)
	if err != nil {
		t.Fatalf("Failed to parse response XML: %v", err)
	}

	response, ok := result.(*ProtocolResponse)
	if !ok {
		t.Fatalf("Expected *ProtocolResponse, got %T", result)
	}

	if response.Command != "step_into" {
		t.Errorf("Expected Command 'step_into', got '%s'", response.Command)
	}

	if response.TransactionID != "1" {
		t.Errorf("Expected TransactionID '1', got '%s'", response.TransactionID)
	}

	if response.Status != "break" {
		t.Errorf("Expected Status 'break', got '%s'", response.Status)
	}

	if response.Reason != "ok" {
		t.Errorf("Expected Reason 'ok', got '%s'", response.Reason)
	}
}

func TestCreateProtocolFromXML_ResponseWithError(t *testing.T) {
	xmlData := `<?xml version="1.0" encoding="iso-8859-1"?>
<response xmlns="urn:debugger_protocol_v1" xmlns:xdebug="https://xdebug.org/dbgp/xdebug"
          command="breakpoint_set"
          transaction_id="2">
    <error code="5">
        <message>Invalid breakpoint</message>
    </error>
</response>`

	result, err := CreateProtocolFromXML(xmlData)
	if err != nil {
		t.Fatalf("Failed to parse response XML: %v", err)
	}

	response, ok := result.(*ProtocolResponse)
	if !ok {
		t.Fatalf("Expected *ProtocolResponse, got %T", result)
	}

	if !response.HasError() {
		t.Error("Expected response to have error")
	}

	if response.Error == nil {
		t.Fatal("Expected non-nil Error")
	}

	if response.Error.Code != "5" {
		t.Errorf("Expected ErrorCode '5', got '%s'", response.Error.Code)
	}

	if response.Error.Message != "Invalid breakpoint" {
		t.Errorf("Expected ErrorMessage 'Invalid breakpoint', got '%s'", response.Error.Message)
	}

	errMsg := response.GetErrorMessage()
	if errMsg != "Invalid breakpoint" {
		t.Errorf("Expected GetErrorMessage() 'Invalid breakpoint', got '%s'", errMsg)
	}
}

func TestCreateProtocolFromXML_ResponseWithProperties(t *testing.T) {
	xmlData := `<?xml version="1.0" encoding="iso-8859-1"?>
<response xmlns="urn:debugger_protocol_v1" xmlns:xdebug="https://xdebug.org/dbgp/xdebug"
          command="context_get"
          transaction_id="3"
          context="0">
    <property name="var1" fullname="$var1" type="string" size="5" encoding="base64">
        <![CDATA[aGVsbG8=]]>
    </property>
    <property name="var2" fullname="$var2" type="int">
        <![CDATA[42]]>
    </property>
</response>`

	result, err := CreateProtocolFromXML(xmlData)
	if err != nil {
		t.Fatalf("Failed to parse response XML: %v", err)
	}

	response, ok := result.(*ProtocolResponse)
	if !ok {
		t.Fatalf("Expected *ProtocolResponse, got %T", result)
	}

	if len(response.Properties) != 2 {
		t.Fatalf("Expected 2 properties, got %d", len(response.Properties))
	}

	prop1 := response.Properties[0]
	if prop1.Name != "var1" {
		t.Errorf("Expected property name 'var1', got '%s'", prop1.Name)
	}
	if prop1.Type != "string" {
		t.Errorf("Expected property type 'string', got '%s'", prop1.Type)
	}
	if prop1.Size != "5" {
		t.Errorf("Expected property size '5', got '%s'", prop1.Size)
	}

	prop2 := response.Properties[1]
	if prop2.Name != "var2" {
		t.Errorf("Expected property name 'var2', got '%s'", prop2.Name)
	}
	if prop2.Type != "int" {
		t.Errorf("Expected property type 'int', got '%s'", prop2.Type)
	}
}

func TestCreateProtocolFromXML_ResponseWithBreakpoints(t *testing.T) {
	xmlData := `<?xml version="1.0" encoding="iso-8859-1"?>
<response xmlns="urn:debugger_protocol_v1" xmlns:xdebug="https://xdebug.org/dbgp/xdebug"
          command="breakpoint_list"
          transaction_id="4">
    <breakpoint id="1" type="line" state="enabled" filename="file:///path/to/file.php" lineno="10"></breakpoint>
    <breakpoint id="2" type="call" state="enabled" function="myFunction"></breakpoint>
</response>`

	result, err := CreateProtocolFromXML(xmlData)
	if err != nil {
		t.Fatalf("Failed to parse response XML: %v", err)
	}

	response, ok := result.(*ProtocolResponse)
	if !ok {
		t.Fatalf("Expected *ProtocolResponse, got %T", result)
	}

	if len(response.Breakpoints) != 2 {
		t.Fatalf("Expected 2 breakpoints, got %d", len(response.Breakpoints))
	}

	bp1 := response.Breakpoints[0]
	if bp1.ID != "1" {
		t.Errorf("Expected breakpoint ID '1', got '%s'", bp1.ID)
	}
	if bp1.Type != "line" {
		t.Errorf("Expected breakpoint type 'line', got '%s'", bp1.Type)
	}
	if bp1.Lineno != "10" {
		t.Errorf("Expected breakpoint lineno '10', got '%s'", bp1.Lineno)
	}

	bp2 := response.Breakpoints[1]
	if bp2.ID != "2" {
		t.Errorf("Expected breakpoint ID '2', got '%s'", bp2.ID)
	}
	if bp2.Type != "call" {
		t.Errorf("Expected breakpoint type 'call', got '%s'", bp2.Type)
	}
	if bp2.Function != "myFunction" {
		t.Errorf("Expected breakpoint function 'myFunction', got '%s'", bp2.Function)
	}
}

func TestProtocolResponse_HasError(t *testing.T) {
	tests := []struct {
		name      string
		response  ProtocolResponse
		wantError bool
	}{
		{
			name:      "no error",
			response:  ProtocolResponse{},
			wantError: false,
		},
		{
			name: "with error code",
			response: ProtocolResponse{
				Error: &ProtocolError{
					Code: "5",
				},
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.response.HasError(); got != tt.wantError {
				t.Errorf("HasError() = %v, want %v", got, tt.wantError)
			}
		})
	}
}

func TestProtocolResponse_GetErrorMessage(t *testing.T) {
	tests := []struct {
		name     string
		response ProtocolResponse
		want     string
	}{
		{
			name:     "no error",
			response: ProtocolResponse{},
			want:     "",
		},
		{
			name: "error with message",
			response: ProtocolResponse{
				Error: &ProtocolError{
					Code:    "5",
					Message: "Invalid breakpoint",
				},
			},
			want: "Invalid breakpoint",
		},
		{
			name: "error without message",
			response: ProtocolResponse{
				Error: &ProtocolError{
					Code: "404",
				},
			},
			want: "Error code: 404",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.response.GetErrorMessage(); got != tt.want {
				t.Errorf("GetErrorMessage() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCreateProtocolFromXML_InvalidXML(t *testing.T) {
	xmlData := `<invalid xml`

	_, err := CreateProtocolFromXML(xmlData)
	if err == nil {
		t.Error("Expected error for invalid XML, got nil")
	}
}
