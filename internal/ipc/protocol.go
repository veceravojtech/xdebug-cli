package ipc

import (
	"encoding/json"
	"fmt"
)

// CommandRequest represents a request to execute commands in the daemon
type CommandRequest struct {
	Type       string   `json:"type"`        // Request type (e.g., "execute_commands", "kill")
	Commands   []string `json:"commands"`    // Commands to execute
	JSONOutput bool     `json:"json_output"` // Whether to return JSON output
}

// CommandResponse represents the response from the daemon
type CommandResponse struct {
	Success bool            `json:"success"` // Whether all commands succeeded
	Results []CommandResult `json:"results"` // Results for each command
	Error   string          `json:"error"`   // Error message if success=false
}

// CommandResult represents the result of a single command execution
type CommandResult struct {
	Command string      `json:"command"` // Command that was executed
	Success bool        `json:"success"` // Whether command succeeded
	Result  interface{} `json:"result"`  // Command result (structure varies by command)
	Error   string      `json:"error"`   // Error message if success=false
}

// Validate validates the CommandRequest fields
func (r *CommandRequest) Validate() error {
	if r.Type == "" {
		return fmt.Errorf("request type is required")
	}

	if r.Type == "execute_commands" && len(r.Commands) == 0 {
		return fmt.Errorf("at least one command is required for execute_commands request")
	}

	return nil
}

// ToJSON serializes the CommandRequest to JSON
func (r *CommandRequest) ToJSON() ([]byte, error) {
	return json.Marshal(r)
}

// FromJSON deserializes a CommandRequest from JSON
func (r *CommandRequest) FromJSON(data []byte) error {
	if err := json.Unmarshal(data, r); err != nil {
		return fmt.Errorf("failed to unmarshal request: %w", err)
	}
	return r.Validate()
}

// ToJSON serializes the CommandResponse to JSON
func (r *CommandResponse) ToJSON() ([]byte, error) {
	return json.Marshal(r)
}

// FromJSON deserializes a CommandResponse from JSON
func (r *CommandResponse) FromJSON(data []byte) error {
	if err := json.Unmarshal(data, r); err != nil {
		return fmt.Errorf("failed to unmarshal response: %w", err)
	}
	return nil
}

// NewExecuteCommandsRequest creates a new CommandRequest for executing commands
func NewExecuteCommandsRequest(commands []string, jsonOutput bool) *CommandRequest {
	return &CommandRequest{
		Type:       "execute_commands",
		Commands:   commands,
		JSONOutput: jsonOutput,
	}
}

// NewKillRequest creates a new CommandRequest for killing the daemon
func NewKillRequest() *CommandRequest {
	return &CommandRequest{
		Type: "kill",
	}
}

// NewSuccessResponse creates a successful CommandResponse
func NewSuccessResponse(results []CommandResult) *CommandResponse {
	return &CommandResponse{
		Success: true,
		Results: results,
	}
}

// NewErrorResponse creates an error CommandResponse
func NewErrorResponse(err string) *CommandResponse {
	return &CommandResponse{
		Success: false,
		Error:   err,
	}
}
