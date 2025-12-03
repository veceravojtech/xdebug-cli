package view

import (
	"encoding/json"
	"fmt"
)

// JSONResponse represents a standardized JSON response for commands
type JSONResponse struct {
	Command string      `json:"command"`
	Success bool        `json:"success"`
	Error   string      `json:"error,omitempty"`
	Result  interface{} `json:"result,omitempty"`
}

// JSONProperty represents a property/variable in JSON format
type JSONProperty struct {
	Name        string         `json:"name"`
	FullName    string         `json:"fullname"`
	Type        string         `json:"type"`
	Value       string         `json:"value"`
	NumChildren int            `json:"num_children"`
	Children    []JSONProperty `json:"children,omitempty"`
}

// JSONBreakpoint represents a breakpoint in JSON format
type JSONBreakpoint struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	State    string `json:"state"`
	Filename string `json:"filename,omitempty"`
	Line     int    `json:"line,omitempty"`
	Function string `json:"function,omitempty"`
}

// JSONStack represents a stack frame in JSON format
type JSONStack struct {
	Level    int    `json:"level"`
	Where    string `json:"where"`
	Filename string `json:"filename"`
	Line     int    `json:"line"`
}

// JSONStateResult represents the result of a state-changing command (run, step, next)
type JSONStateResult struct {
	Status   string `json:"status"`
	Filename string `json:"filename,omitempty"`
	Line     int    `json:"line,omitempty"`
}

// JSONBreakpointResult represents the result of setting a breakpoint
type JSONBreakpointResult struct {
	ID        string `json:"id,omitempty"`
	Location  string `json:"location"`
	Condition string `json:"condition,omitempty"`
	Error     string `json:"error,omitempty"`
}

// OutputJSON outputs a JSON-formatted response
func (v *View) OutputJSON(command string, success bool, errorMsg string, result interface{}) {
	response := JSONResponse{
		Command: command,
		Success: success,
		Error:   errorMsg,
		Result:  result,
	}

	jsonBytes, err := json.Marshal(response)
	if err != nil {
		// Fallback error output
		fmt.Fprintf(v.stderr, `{"command":"%s","success":false,"error":"Failed to marshal JSON: %v"}`+"\n", command, err)
		return
	}

	fmt.Fprintln(v.stdout, string(jsonBytes))
}

// OutputJSONArray outputs an array of results as JSON
func (v *View) OutputJSONArray(command string, data interface{}) {
	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		// Fallback error output
		fmt.Fprintln(v.stdout, `{"error": "failed to marshal JSON"}`)
		return
	}
	fmt.Fprintln(v.stdout, string(jsonBytes))
}

// ConvertPropertyToJSON converts a ProtocolProperty to JSONProperty
// Base64-encoded string values are automatically decoded
func ConvertPropertyToJSON(prop ProtocolProperty) JSONProperty {
	// Decode base64 values for string types
	value := TryDecodeBase64(prop.GetValue(), prop.GetType())

	jsonProp := JSONProperty{
		Name:        prop.GetName(),
		FullName:    prop.GetFullName(),
		Type:        prop.GetType(),
		Value:       value,
		NumChildren: prop.GetNumChildren(),
	}

	// Convert children if present
	if prop.HasChildren() {
		children := prop.GetChildren()
		jsonProp.Children = make([]JSONProperty, 0, len(children))
		for _, child := range children {
			if childProp, ok := child.(ProtocolProperty); ok {
				jsonProp.Children = append(jsonProp.Children, ConvertPropertyToJSON(childProp))
			}
		}
	}

	return jsonProp
}

// ConvertBreakpointToJSON converts a ProtocolBreakpoint to JSONBreakpoint
func ConvertBreakpointToJSON(bp ProtocolBreakpoint) JSONBreakpoint {
	return JSONBreakpoint{
		ID:       bp.GetID(),
		Type:     bp.GetType(),
		State:    bp.GetState(),
		Filename: bp.GetFilename(),
		Line:     bp.GetLineNumber(),
		Function: bp.GetFunction(),
	}
}

// ConvertStackToJSON converts a ProtocolStack to JSON format
func ConvertStackToJSON(frame ProtocolStack) JSONStack {
	return JSONStack{
		Level:    frame.GetLevel(),
		Where:    frame.GetWhere(),
		Filename: frame.GetFilename(),
		Line:     frame.GetLineNumber(),
	}
}
