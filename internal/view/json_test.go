package view

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestOutputJSON(t *testing.T) {
	tests := []struct {
		name    string
		command string
		success bool
		error   string
		result  interface{}
	}{
		{
			name:    "successful command with result",
			command: "run",
			success: true,
			error:   "",
			result:  map[string]interface{}{"status": "break", "line": 42},
		},
		{
			name:    "failed command with error",
			command: "step",
			success: false,
			error:   "session ended",
			result:  nil,
		},
		{
			name:    "command with no result",
			command: "quit",
			success: true,
			error:   "",
			result:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			v := &View{
				stdout: &buf,
				stderr: &buf,
			}

			v.OutputJSON(tt.command, tt.success, tt.error, tt.result)

			output := buf.String()
			if output == "" {
				t.Fatal("expected non-empty output")
			}

			// Parse JSON to validate structure
			var response JSONResponse
			if err := json.Unmarshal([]byte(strings.TrimSpace(output)), &response); err != nil {
				t.Fatalf("failed to parse JSON: %v\nOutput: %s", err, output)
			}

			if response.Command != tt.command {
				t.Errorf("expected command %q, got %q", tt.command, response.Command)
			}

			if response.Success != tt.success {
				t.Errorf("expected success=%v, got %v", tt.success, response.Success)
			}

			if response.Error != tt.error {
				t.Errorf("expected error %q, got %q", tt.error, response.Error)
			}
		})
	}
}

func TestJSONPropertyConversion(t *testing.T) {
	// Mock property for testing
	mockProp := &mockProperty{
		name:        "myVar",
		fullName:    "$myVar",
		propType:    "string",
		value:       "test value",
		numChildren: 0,
		hasChildren: false,
	}

	jsonProp := ConvertPropertyToJSON(mockProp)

	if jsonProp.Name != mockProp.name {
		t.Errorf("expected name %q, got %q", mockProp.name, jsonProp.Name)
	}

	if jsonProp.FullName != mockProp.fullName {
		t.Errorf("expected fullName %q, got %q", mockProp.fullName, jsonProp.FullName)
	}

	if jsonProp.Type != mockProp.propType {
		t.Errorf("expected type %q, got %q", mockProp.propType, jsonProp.Type)
	}

	if jsonProp.Value != mockProp.value {
		t.Errorf("expected value %q, got %q", mockProp.value, jsonProp.Value)
	}

	if jsonProp.NumChildren != mockProp.numChildren {
		t.Errorf("expected numChildren %d, got %d", mockProp.numChildren, jsonProp.NumChildren)
	}
}

func TestJSONBreakpointConversion(t *testing.T) {
	// Mock breakpoint for testing
	mockBp := &mockBreakpoint{
		id:       "1",
		bpType:   "line",
		state:    "enabled",
		filename: "/path/to/file.php",
		line:     42,
		function: "",
	}

	jsonBp := ConvertBreakpointToJSON(mockBp)

	if jsonBp.ID != mockBp.id {
		t.Errorf("expected ID %q, got %q", mockBp.id, jsonBp.ID)
	}

	if jsonBp.Type != mockBp.bpType {
		t.Errorf("expected type %q, got %q", mockBp.bpType, jsonBp.Type)
	}

	if jsonBp.State != mockBp.state {
		t.Errorf("expected state %q, got %q", mockBp.state, jsonBp.State)
	}

	if jsonBp.Filename != mockBp.filename {
		t.Errorf("expected filename %q, got %q", mockBp.filename, jsonBp.Filename)
	}

	if jsonBp.Line != mockBp.line {
		t.Errorf("expected line %d, got %d", mockBp.line, jsonBp.Line)
	}
}
