package view

// ProtocolBreakpoint represents a breakpoint in the DBGp protocol.
// This interface is implemented by dbgp.ProtocolBreakpoint.
type ProtocolBreakpoint interface {
	GetID() string
	GetType() string
	GetState() string
	GetFilename() string
	GetLineNumber() int
	GetFunction() string
}

// ProtocolProperty represents a variable/property in the DBGp protocol.
// This interface is implemented by dbgp.ProtocolProperty.
type ProtocolProperty interface {
	GetName() string
	GetFullName() string
	GetType() string
	GetValue() string
	GetChildren() []interface{}
	HasChildren() bool
	GetNumChildren() int
}

// ProtocolStack represents a stack frame in the DBGp protocol.
// This interface is implemented by dbgp.ProtocolStack.
type ProtocolStack interface {
	GetWhere() string
	GetLevel() int
	GetType() string
	GetFilename() string
	GetLineNumber() int
}
