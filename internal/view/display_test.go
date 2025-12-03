package view

import (
	"bytes"
	"strings"
	"testing"
)

// Mock implementations for testing

type mockBreakpoint struct {
	id       string
	bpType   string
	state    string
	filename string
	line     int
	function string
}

func (m *mockBreakpoint) GetID() string         { return m.id }
func (m *mockBreakpoint) GetType() string       { return m.bpType }
func (m *mockBreakpoint) GetState() string      { return m.state }
func (m *mockBreakpoint) GetFilename() string   { return m.filename }
func (m *mockBreakpoint) GetLineNumber() int    { return m.line }
func (m *mockBreakpoint) GetFunction() string   { return m.function }

type mockProperty struct {
	name        string
	fullName    string
	propType    string
	value       string
	children    []interface{}
	hasChildren bool
	numChildren int
}

func (m *mockProperty) GetName() string        { return m.name }
func (m *mockProperty) GetFullName() string    { return m.fullName }
func (m *mockProperty) GetType() string        { return m.propType }
func (m *mockProperty) GetValue() string       { return m.value }
func (m *mockProperty) GetChildren() []interface{} { return m.children }
func (m *mockProperty) HasChildren() bool      { return m.hasChildren }
func (m *mockProperty) GetNumChildren() int    { return m.numChildren }

func TestView_ShowInfoBreakpoints(t *testing.T) {
	tests := []struct {
		name        string
		breakpoints []ProtocolBreakpoint
		wantOutput  []string
	}{
		{
			name:        "no breakpoints",
			breakpoints: []ProtocolBreakpoint{},
			wantOutput:  []string{"No breakpoints set."},
		},
		{
			name: "single line breakpoint",
			breakpoints: []ProtocolBreakpoint{
				&mockBreakpoint{
					id:       "1",
					bpType:   "line",
					state:    "enabled",
					filename: "/path/to/file.php",
					line:     42,
				},
			},
			wantOutput: []string{
				"Breakpoints:",
				"ID",
				"Type",
				"State",
				"Location",
				"1",
				"line",
				"enabled",
				"file.php:42",
			},
		},
		{
			name: "multiple breakpoints",
			breakpoints: []ProtocolBreakpoint{
				&mockBreakpoint{
					id:       "1",
					bpType:   "line",
					state:    "enabled",
					filename: "/path/to/file.php",
					line:     42,
				},
				&mockBreakpoint{
					id:       "2",
					bpType:   "call",
					state:    "enabled",
					function: "myFunction",
				},
			},
			wantOutput: []string{
				"Breakpoints:",
				"1",
				"2",
				"line",
				"call",
				"myFunction",
			},
		},
		{
			name: "long path truncation",
			breakpoints: []ProtocolBreakpoint{
				&mockBreakpoint{
					id:       "1",
					bpType:   "line",
					state:    "enabled",
					filename: "/very/long/path/to/some/deeply/nested/directory/structure/file.php",
					line:     100,
				},
			},
			wantOutput: []string{
				"...",
				"file.php:100",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			v := &View{stdout: &buf}

			v.ShowInfoBreakpoints(tt.breakpoints)

			output := buf.String()
			for _, want := range tt.wantOutput {
				if !strings.Contains(output, want) {
					t.Errorf("ShowInfoBreakpoints() output missing %q\nGot: %s", want, output)
				}
			}
		})
	}
}

func TestView_PrintPropertyListWithDetails(t *testing.T) {
	tests := []struct {
		name       string
		scope      string
		properties []ProtocolProperty
		wantOutput []string
	}{
		{
			name:       "no properties",
			scope:      "Local",
			properties: []ProtocolProperty{},
			wantOutput: []string{"No local variables."},
		},
		{
			name:  "simple properties",
			scope: "Local",
			properties: []ProtocolProperty{
				&mockProperty{
					name:     "$var",
					propType: "int",
					value:    "42",
				},
				&mockProperty{
					name:     "$str",
					propType: "string",
					value:    "hello",
				},
			},
			wantOutput: []string{
				"Local Variables:",
				"$var (int) = 42",
				"$str (string) = hello",
			},
		},
		{
			name:  "nested properties",
			scope: "Local",
			properties: []ProtocolProperty{
				&mockProperty{
					name:        "$arr",
					propType:    "array",
					hasChildren: true,
					numChildren: 2,
					children: []interface{}{
						&mockProperty{
							name:     "0",
							propType: "int",
							value:    "1",
						},
						&mockProperty{
							name:     "1",
							propType: "int",
							value:    "2",
						},
					},
				},
			},
			wantOutput: []string{
				"Local Variables:",
				"$arr (array) [2 children]",
				"  0 (int) = 1",
				"  1 (int) = 2",
			},
		},
		{
			name:  "deeply nested properties",
			scope: "Local",
			properties: []ProtocolProperty{
				&mockProperty{
					name:        "$obj",
					propType:    "object",
					hasChildren: true,
					numChildren: 1,
					children: []interface{}{
						&mockProperty{
							name:        "prop",
							propType:    "array",
							hasChildren: true,
							numChildren: 1,
							children: []interface{}{
								&mockProperty{
									name:     "key",
									propType: "string",
									value:    "value",
								},
							},
						},
					},
				},
			},
			wantOutput: []string{
				"$obj (object) [1 children]",
				"  prop (array) [1 children]",
				"    key (string) = value",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			v := &View{stdout: &buf}

			v.PrintPropertyListWithDetails(tt.scope, tt.properties)

			output := buf.String()
			for _, want := range tt.wantOutput {
				if !strings.Contains(output, want) {
					t.Errorf("PrintPropertyListWithDetails() output missing %q\nGot: %s", want, output)
				}
			}
		})
	}
}

func TestView_PrintProperty(t *testing.T) {
	var buf bytes.Buffer
	v := &View{stdout: &buf}

	prop := &mockProperty{
		name:     "$var",
		propType: "string",
		value:    "test",
	}

	v.PrintProperty(prop)

	output := buf.String()
	if !strings.Contains(output, "$var (string) = test") {
		t.Errorf("PrintProperty() output missing expected content\nGot: %s", output)
	}
}

func TestPrintProperty_ValueTruncation(t *testing.T) {
	var buf bytes.Buffer
	v := &View{stdout: &buf}

	// Create a property with a very long value
	longValue := strings.Repeat("x", 200)
	prop := &mockProperty{
		name:     "$long",
		propType: "string",
		value:    longValue,
	}

	v.PrintProperty(prop)

	output := buf.String()
	// Should truncate and add "..."
	if !strings.Contains(output, "...") {
		t.Error("PrintProperty() should truncate long values")
	}
	// Original long value should not be fully present
	if strings.Contains(output, longValue) {
		t.Error("PrintProperty() should not contain full long value")
	}
}

func TestTryDecodeBase64(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		propType string
		want     string
	}{
		{
			name:     "valid base64 string",
			value:    "RMOhcmtvdsO9IHBvdWtheg==",
			propType: "string",
			want:     "Dárkový poukaz",
		},
		{
			name:     "non-base64 string",
			value:    "hello world",
			propType: "string",
			want:     "hello world",
		},
		{
			name:     "invalid base64",
			value:    "not!valid@base64",
			propType: "string",
			want:     "not!valid@base64",
		},
		{
			name:     "non-string type",
			value:    "RMOhcmtvdsO9IHBvdWtheg==",
			propType: "int",
			want:     "RMOhcmtvdsO9IHBvdWtheg==",
		},
		{
			name:     "empty string",
			value:    "",
			propType: "string",
			want:     "",
		},
		{
			name:     "short string",
			value:    "abc",
			propType: "string",
			want:     "abc",
		},
		{
			name:     "base64 with non-utf8",
			value:    "////",
			propType: "string",
			want:     "////", // Should return original if decoded is not valid UTF-8
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tryDecodeBase64(tt.value, tt.propType)
			if got != tt.want {
				t.Errorf("tryDecodeBase64() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestPrintProperty_Base64Decoding(t *testing.T) {
	var buf bytes.Buffer
	v := &View{stdout: &buf}

	prop := &mockProperty{
		name:     "coupon",
		propType: "string",
		value:    "RMOhcmtvdsO9IHBvdWtheg==",
	}

	v.PrintProperty(prop)

	output := buf.String()
	// Should contain decoded value
	if !strings.Contains(output, "Dárkový poukaz") {
		t.Errorf("PrintProperty() should decode base64 string\nGot: %s", output)
	}
	// Should NOT contain the base64 encoded value
	if strings.Contains(output, "RMOhcmtvdsO9IHBvdWtheg==") {
		t.Errorf("PrintProperty() should not show base64 encoded value\nGot: %s", output)
	}
}
