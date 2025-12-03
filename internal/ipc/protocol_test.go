package ipc

import (
	"encoding/json"
	"testing"
)

func TestCommandRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		req     CommandRequest
		wantErr bool
	}{
		{
			name: "valid execute_commands request",
			req: CommandRequest{
				Type:     "execute_commands",
				Commands: []string{"run", "step"},
			},
			wantErr: false,
		},
		{
			name: "valid kill request",
			req: CommandRequest{
				Type: "kill",
			},
			wantErr: false,
		},
		{
			name: "missing type",
			req: CommandRequest{
				Commands: []string{"run"},
			},
			wantErr: true,
		},
		{
			name: "execute_commands with no commands",
			req: CommandRequest{
				Type:     "execute_commands",
				Commands: []string{},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCommandRequest_JSON(t *testing.T) {
	req := NewExecuteCommandsRequest([]string{"run", "step"}, true)

	// Serialize
	data, err := req.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON() error = %v", err)
	}

	// Deserialize
	var decoded CommandRequest
	if err := decoded.FromJSON(data); err != nil {
		t.Fatalf("FromJSON() error = %v", err)
	}

	// Verify fields
	if decoded.Type != "execute_commands" {
		t.Errorf("Type = %v, want execute_commands", decoded.Type)
	}
	if len(decoded.Commands) != 2 {
		t.Errorf("Commands length = %v, want 2", len(decoded.Commands))
	}
	if decoded.Commands[0] != "run" || decoded.Commands[1] != "step" {
		t.Errorf("Commands = %v, want [run step]", decoded.Commands)
	}
	if !decoded.JSONOutput {
		t.Errorf("JSONOutput = %v, want true", decoded.JSONOutput)
	}
}

func TestCommandRequest_FromJSON_InvalidJSON(t *testing.T) {
	var req CommandRequest
	err := req.FromJSON([]byte("invalid json"))
	if err == nil {
		t.Error("FromJSON() expected error for invalid JSON")
	}
}

func TestCommandResponse_JSON(t *testing.T) {
	resp := NewSuccessResponse([]CommandResult{
		{
			Command: "run",
			Success: true,
			Result:  map[string]interface{}{"status": "break", "line": 42},
		},
	})

	// Serialize
	data, err := resp.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON() error = %v", err)
	}

	// Deserialize
	var decoded CommandResponse
	if err := decoded.FromJSON(data); err != nil {
		t.Fatalf("FromJSON() error = %v", err)
	}

	// Verify fields
	if !decoded.Success {
		t.Errorf("Success = %v, want true", decoded.Success)
	}
	if len(decoded.Results) != 1 {
		t.Errorf("Results length = %v, want 1", len(decoded.Results))
	}
	if decoded.Results[0].Command != "run" {
		t.Errorf("Results[0].Command = %v, want run", decoded.Results[0].Command)
	}
}

func TestCommandResponse_Error(t *testing.T) {
	resp := NewErrorResponse("session ended")

	if resp.Success {
		t.Errorf("Success = %v, want false", resp.Success)
	}
	if resp.Error != "session ended" {
		t.Errorf("Error = %v, want 'session ended'", resp.Error)
	}
}

func TestNewExecuteCommandsRequest(t *testing.T) {
	req := NewExecuteCommandsRequest([]string{"run"}, false)

	if req.Type != "execute_commands" {
		t.Errorf("Type = %v, want execute_commands", req.Type)
	}
	if len(req.Commands) != 1 || req.Commands[0] != "run" {
		t.Errorf("Commands = %v, want [run]", req.Commands)
	}
	if req.JSONOutput {
		t.Errorf("JSONOutput = %v, want false", req.JSONOutput)
	}
}

func TestNewKillRequest(t *testing.T) {
	req := NewKillRequest()

	if req.Type != "kill" {
		t.Errorf("Type = %v, want kill", req.Type)
	}
}

func TestCommandResult_ComplexResult(t *testing.T) {
	// Test that CommandResult can handle complex nested structures
	result := CommandResult{
		Command: "context",
		Success: true,
		Result: map[string]interface{}{
			"variables": []map[string]interface{}{
				{"name": "x", "value": "42"},
				{"name": "y", "value": "hello"},
			},
		},
	}

	// Serialize and deserialize
	data, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("Marshal error = %v", err)
	}

	var decoded CommandResult
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal error = %v", err)
	}

	// Verify structure
	if !decoded.Success {
		t.Errorf("Success = %v, want true", decoded.Success)
	}
	if decoded.Command != "context" {
		t.Errorf("Command = %v, want context", decoded.Command)
	}

	// Verify nested result
	resultMap, ok := decoded.Result.(map[string]interface{})
	if !ok {
		t.Fatalf("Result is not a map")
	}
	variables, ok := resultMap["variables"].([]interface{})
	if !ok {
		t.Fatalf("variables is not an array")
	}
	if len(variables) != 2 {
		t.Errorf("variables length = %v, want 2", len(variables))
	}
}
