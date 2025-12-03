package dbgp

import (
	"sync"
)

// SessionStateType represents the current state of a debugging session
type SessionStateType int

const (
	// StateNone represents the initial state before initialization
	StateNone SessionStateType = iota
	// StateStarting represents the state after init but before first run
	StateStarting
	// StateRunning represents the state when code is executing
	StateRunning
	// StateBreak represents the state when paused at a breakpoint or step
	StateBreak
	// StateStopping represents the state when stop command has been sent
	StateStopping
	// StateStopped represents the state when the session has ended
	StateStopped
)

// String returns the string representation of the session state
func (s SessionStateType) String() string {
	switch s {
	case StateNone:
		return "none"
	case StateStarting:
		return "starting"
	case StateRunning:
		return "running"
	case StateBreak:
		return "break"
	case StateStopping:
		return "stopping"
	case StateStopped:
		return "stopped"
	default:
		return "unknown"
	}
}

// CommandRecord represents a sent command and its transaction ID
type CommandRecord struct {
	TransactionID string
	Command       string
}

// Session manages the state of a debugging session
type Session struct {
	mu             sync.RWMutex
	state          SessionStateType
	transactionID  int
	commands       []CommandRecord
	targetFiles    []string
	currentFile    string
	currentLine    int
	ideKey         string
	appID          string
}

// NewSession creates a new debugging session
func NewSession() *Session {
	return &Session{
		state:         StateNone,
		transactionID: 0,
		commands:      make([]CommandRecord, 0),
		targetFiles:   make([]string, 0),
	}
}

// GetState returns the current session state
func (s *Session) GetState() SessionStateType {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.state
}

// SetState sets the current session state
func (s *Session) SetState(state SessionStateType) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state = state
}

// NextTransactionID generates and returns the next transaction ID
func (s *Session) NextTransactionID() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.transactionID++
	return string(rune('0' + s.transactionID%10)) + string(rune('0' + (s.transactionID/10)%10)) + string(rune('0' + (s.transactionID/100)%10))
}

// NextTransactionIDInt generates and returns the next transaction ID as integer
func (s *Session) NextTransactionIDInt() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.transactionID++
	return s.transactionID
}

// AddCommand records a sent command with its transaction ID
func (s *Session) AddCommand(transactionID, command string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.commands = append(s.commands, CommandRecord{
		TransactionID: transactionID,
		Command:       command,
	})
}

// GetLastCommand returns the most recently sent command
func (s *Session) GetLastCommand() *CommandRecord {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if len(s.commands) == 0 {
		return nil
	}
	return &s.commands[len(s.commands)-1]
}

// GetCommandByTransactionID returns the command with the given transaction ID
func (s *Session) GetCommandByTransactionID(transactionID string) *CommandRecord {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for i := len(s.commands) - 1; i >= 0; i-- {
		if s.commands[i].TransactionID == transactionID {
			return &s.commands[i]
		}
	}
	return nil
}

// SetTargetFiles sets the list of target files for the session
func (s *Session) SetTargetFiles(rootFile string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.targetFiles = []string{rootFile}
}

// GetTargetFiles returns the list of target files
func (s *Session) GetTargetFiles() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	files := make([]string, len(s.targetFiles))
	copy(files, s.targetFiles)
	return files
}

// SetCurrentLocation sets the current file and line number
func (s *Session) SetCurrentLocation(file string, line int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.currentFile = file
	s.currentLine = line
}

// GetCurrentLocation returns the current file and line number
func (s *Session) GetCurrentLocation() (string, int) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.currentFile, s.currentLine
}

// SetIDEKey sets the IDE key from the init message
func (s *Session) SetIDEKey(ideKey string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ideKey = ideKey
}

// GetIDEKey returns the IDE key
func (s *Session) GetIDEKey() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.ideKey
}

// SetAppID sets the application ID from the init message
func (s *Session) SetAppID(appID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.appID = appID
}

// GetAppID returns the application ID
func (s *Session) GetAppID() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.appID
}

// Reset resets the session to initial state
func (s *Session) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state = StateNone
	s.transactionID = 0
	s.commands = make([]CommandRecord, 0)
	s.targetFiles = make([]string, 0)
	s.currentFile = ""
	s.currentLine = 0
	s.ideKey = ""
	s.appID = ""
}
