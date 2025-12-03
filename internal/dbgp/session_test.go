package dbgp

import (
	"sync"
	"testing"
)

func TestNewSession(t *testing.T) {
	session := NewSession()

	if session == nil {
		t.Fatal("Expected non-nil session")
	}

	if session.GetState() != StateNone {
		t.Errorf("Expected initial state StateNone, got %v", session.GetState())
	}

	if len(session.commands) != 0 {
		t.Errorf("Expected empty commands list, got %d items", len(session.commands))
	}
}

func TestSessionStateType_String(t *testing.T) {
	tests := []struct {
		state    SessionStateType
		expected string
	}{
		{StateNone, "none"},
		{StateStarting, "starting"},
		{StateRunning, "running"},
		{StateBreak, "break"},
		{StateStopping, "stopping"},
		{StateStopped, "stopped"},
		{SessionStateType(99), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.state.String(); got != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, got)
			}
		})
	}
}

func TestSession_GetSetState(t *testing.T) {
	session := NewSession()

	states := []SessionStateType{
		StateStarting,
		StateRunning,
		StateBreak,
		StateStopping,
		StateStopped,
	}

	for _, state := range states {
		session.SetState(state)
		if got := session.GetState(); got != state {
			t.Errorf("Expected state %v, got %v", state, got)
		}
	}
}

func TestSession_NextTransactionIDInt(t *testing.T) {
	session := NewSession()

	// Test sequential transaction IDs
	for i := 1; i <= 5; i++ {
		id := session.NextTransactionIDInt()
		if id != i {
			t.Errorf("Expected transaction ID %d, got %d", i, id)
		}
	}
}

func TestSession_NextTransactionID(t *testing.T) {
	session := NewSession()

	// Test that transaction IDs are generated
	id1 := session.NextTransactionID()
	id2 := session.NextTransactionID()

	if id1 == id2 {
		t.Error("Expected different transaction IDs")
	}

	// Both should be non-empty strings
	if id1 == "" || id2 == "" {
		t.Error("Expected non-empty transaction IDs")
	}
}

func TestSession_AddAndGetCommand(t *testing.T) {
	session := NewSession()

	// Initially no commands
	if cmd := session.GetLastCommand(); cmd != nil {
		t.Error("Expected nil for GetLastCommand on empty session")
	}

	// Add commands
	session.AddCommand("1", "run")
	session.AddCommand("2", "step_into")
	session.AddCommand("3", "step_over")

	// Get last command
	lastCmd := session.GetLastCommand()
	if lastCmd == nil {
		t.Fatal("Expected non-nil command")
	}

	if lastCmd.TransactionID != "3" {
		t.Errorf("Expected transaction ID '3', got '%s'", lastCmd.TransactionID)
	}

	if lastCmd.Command != "step_over" {
		t.Errorf("Expected command 'step_over', got '%s'", lastCmd.Command)
	}
}

func TestSession_GetCommandByTransactionID(t *testing.T) {
	session := NewSession()

	session.AddCommand("1", "run")
	session.AddCommand("2", "step_into")
	session.AddCommand("3", "step_over")

	// Test getting existing command
	cmd := session.GetCommandByTransactionID("2")
	if cmd == nil {
		t.Fatal("Expected non-nil command")
	}

	if cmd.Command != "step_into" {
		t.Errorf("Expected command 'step_into', got '%s'", cmd.Command)
	}

	// Test getting non-existent command
	cmd = session.GetCommandByTransactionID("999")
	if cmd != nil {
		t.Error("Expected nil for non-existent transaction ID")
	}
}

func TestSession_SetGetTargetFiles(t *testing.T) {
	session := NewSession()

	// Initially empty
	files := session.GetTargetFiles()
	if len(files) != 0 {
		t.Errorf("Expected empty target files, got %d", len(files))
	}

	// Set target file
	session.SetTargetFiles("/path/to/script.php")

	files = session.GetTargetFiles()
	if len(files) != 1 {
		t.Fatalf("Expected 1 target file, got %d", len(files))
	}

	if files[0] != "/path/to/script.php" {
		t.Errorf("Expected '/path/to/script.php', got '%s'", files[0])
	}
}

func TestSession_SetGetCurrentLocation(t *testing.T) {
	session := NewSession()

	// Initially empty
	file, line := session.GetCurrentLocation()
	if file != "" || line != 0 {
		t.Errorf("Expected empty location, got file='%s', line=%d", file, line)
	}

	// Set location
	session.SetCurrentLocation("/path/to/file.php", 42)

	file, line = session.GetCurrentLocation()
	if file != "/path/to/file.php" {
		t.Errorf("Expected '/path/to/file.php', got '%s'", file)
	}
	if line != 42 {
		t.Errorf("Expected line 42, got %d", line)
	}
}

func TestSession_SetGetIDEKey(t *testing.T) {
	session := NewSession()

	// Initially empty
	if key := session.GetIDEKey(); key != "" {
		t.Errorf("Expected empty IDE key, got '%s'", key)
	}

	// Set IDE key
	session.SetIDEKey("PHPSTORM")

	if key := session.GetIDEKey(); key != "PHPSTORM" {
		t.Errorf("Expected 'PHPSTORM', got '%s'", key)
	}
}

func TestSession_SetGetAppID(t *testing.T) {
	session := NewSession()

	// Initially empty
	if id := session.GetAppID(); id != "" {
		t.Errorf("Expected empty app ID, got '%s'", id)
	}

	// Set app ID
	session.SetAppID("12345")

	if id := session.GetAppID(); id != "12345" {
		t.Errorf("Expected '12345', got '%s'", id)
	}
}

func TestSession_Reset(t *testing.T) {
	session := NewSession()

	// Set various properties
	session.SetState(StateRunning)
	session.AddCommand("1", "run")
	session.SetTargetFiles("/path/to/file.php")
	session.SetCurrentLocation("/path/to/file.php", 42)
	session.SetIDEKey("PHPSTORM")
	session.SetAppID("12345")
	session.NextTransactionIDInt()

	// Reset
	session.Reset()

	// Verify everything is reset
	if session.GetState() != StateNone {
		t.Errorf("Expected state StateNone after reset, got %v", session.GetState())
	}

	if cmd := session.GetLastCommand(); cmd != nil {
		t.Error("Expected no commands after reset")
	}

	if files := session.GetTargetFiles(); len(files) != 0 {
		t.Errorf("Expected no target files after reset, got %d", len(files))
	}

	file, line := session.GetCurrentLocation()
	if file != "" || line != 0 {
		t.Errorf("Expected empty location after reset, got file='%s', line=%d", file, line)
	}

	if key := session.GetIDEKey(); key != "" {
		t.Errorf("Expected empty IDE key after reset, got '%s'", key)
	}

	if id := session.GetAppID(); id != "" {
		t.Errorf("Expected empty app ID after reset, got '%s'", id)
	}
}

func TestSession_ConcurrentAccess(t *testing.T) {
	session := NewSession()
	var wg sync.WaitGroup

	// Test concurrent state changes
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(state SessionStateType) {
			defer wg.Done()
			session.SetState(state)
			_ = session.GetState()
		}(SessionStateType(i % 6))
	}

	// Test concurrent transaction ID generation
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = session.NextTransactionIDInt()
		}()
	}

	// Test concurrent command addition
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			session.AddCommand(string(rune('0'+id)), "test")
		}(i)
	}

	wg.Wait()

	// Verify transaction IDs were generated correctly (should be 10)
	if session.transactionID != 10 {
		t.Errorf("Expected 10 transaction IDs, got %d", session.transactionID)
	}

	// Verify commands were added (should be 10)
	if len(session.commands) != 10 {
		t.Errorf("Expected 10 commands, got %d", len(session.commands))
	}
}
