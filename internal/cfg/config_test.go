package cfg

import "testing"

// TestVersion verifies that the Version constant is defined and non-empty
func TestVersion(t *testing.T) {
	if Version == "" {
		t.Error("Version constant should not be empty")
	}

	expectedVersion := "1.0.0"
	if Version != expectedVersion {
		t.Errorf("Expected version %q, got %q", expectedVersion, Version)
	}
}

// TestCLIParameterInitialization verifies that CLIParameter can be initialized
func TestCLIParameterInitialization(t *testing.T) {
	param := CLIParameter{
		Host:    "localhost",
		Port:    9003,
		Trigger: "PHPSTORM",
	}

	if param.Host != "localhost" {
		t.Errorf("Expected Host to be 'localhost', got %q", param.Host)
	}

	if param.Port != 9003 {
		t.Errorf("Expected Port to be 9003, got %d", param.Port)
	}

	if param.Trigger != "PHPSTORM" {
		t.Errorf("Expected Trigger to be 'PHPSTORM', got %q", param.Trigger)
	}
}

// TestCLIParameterZeroValues verifies zero values work correctly
func TestCLIParameterZeroValues(t *testing.T) {
	var param CLIParameter

	if param.Host != "" {
		t.Errorf("Expected Host zero value to be empty string, got %q", param.Host)
	}

	if param.Port != 0 {
		t.Errorf("Expected Port zero value to be 0, got %d", param.Port)
	}

	if param.Trigger != "" {
		t.Errorf("Expected Trigger zero value to be empty string, got %q", param.Trigger)
	}
}

// TestCLIParameterPartialInitialization verifies partial initialization works
func TestCLIParameterPartialInitialization(t *testing.T) {
	param := CLIParameter{
		Port: 9000,
	}

	if param.Host != "" {
		t.Errorf("Expected Host to be empty string, got %q", param.Host)
	}

	if param.Port != 9000 {
		t.Errorf("Expected Port to be 9000, got %d", param.Port)
	}

	if param.Trigger != "" {
		t.Errorf("Expected Trigger to be empty string, got %q", param.Trigger)
	}
}

// TestCLIParameterFieldTypes verifies that fields have correct types
func TestCLIParameterFieldTypes(t *testing.T) {
	param := CLIParameter{
		Host:    "0.0.0.0",
		Port:    9003,
		Trigger: "xdebug",
	}

	// Type assertions - will compile only if types are correct
	var _ string = param.Host
	var _ int = param.Port
	var _ string = param.Trigger
}

// TestCLIParameterMutability verifies that struct fields can be modified
func TestCLIParameterMutability(t *testing.T) {
	param := CLIParameter{
		Host:    "127.0.0.1",
		Port:    9003,
		Trigger: "PHPSTORM",
	}

	// Modify fields
	param.Host = "0.0.0.0"
	param.Port = 9000
	param.Trigger = "VSCODE"

	if param.Host != "0.0.0.0" {
		t.Errorf("Expected Host to be updated to '0.0.0.0', got %q", param.Host)
	}

	if param.Port != 9000 {
		t.Errorf("Expected Port to be updated to 9000, got %d", param.Port)
	}

	if param.Trigger != "VSCODE" {
		t.Errorf("Expected Trigger to be updated to 'VSCODE', got %q", param.Trigger)
	}
}

// TestCLIParameterCommonValues tests common real-world values
func TestCLIParameterCommonValues(t *testing.T) {
	tests := []struct {
		name    string
		host    string
		port    int
		trigger string
	}{
		{
			name:    "Default Xdebug configuration",
			host:    "localhost",
			port:    9003,
			trigger: "PHPSTORM",
		},
		{
			name:    "Legacy Xdebug 2.x port",
			host:    "127.0.0.1",
			port:    9000,
			trigger: "xdebug",
		},
		{
			name:    "All network interfaces",
			host:    "0.0.0.0",
			port:    9003,
			trigger: "",
		},
		{
			name:    "Custom port",
			host:    "localhost",
			port:    9999,
			trigger: "IDE_KEY",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			param := CLIParameter{
				Host:    tt.host,
				Port:    tt.port,
				Trigger: tt.trigger,
			}

			if param.Host != tt.host {
				t.Errorf("Expected Host %q, got %q", tt.host, param.Host)
			}

			if param.Port != tt.port {
				t.Errorf("Expected Port %d, got %d", tt.port, param.Port)
			}

			if param.Trigger != tt.trigger {
				t.Errorf("Expected Trigger %q, got %q", tt.trigger, param.Trigger)
			}
		})
	}
}
