package dbgp

import (
	"strconv"
)

// View adapter methods for ProtocolBreakpoint
// These methods allow ProtocolBreakpoint to be used with the view package

// GetID returns the breakpoint ID
func (b *ProtocolBreakpoint) GetID() string {
	return b.ID
}

// GetType returns the breakpoint type
func (b *ProtocolBreakpoint) GetType() string {
	return b.Type
}

// GetState returns the breakpoint state
func (b *ProtocolBreakpoint) GetState() string {
	return b.State
}

// GetFilename returns the breakpoint filename
func (b *ProtocolBreakpoint) GetFilename() string {
	return b.Filename
}

// GetLineNumber returns the breakpoint line number as int
func (b *ProtocolBreakpoint) GetLineNumber() int {
	if b.Lineno == "" {
		return 0
	}
	line, err := strconv.Atoi(b.Lineno)
	if err != nil {
		return 0
	}
	return line
}

// GetFunction returns the breakpoint function name
func (b *ProtocolBreakpoint) GetFunction() string {
	return b.Function
}

// View adapter methods for ProtocolProperty
// These methods allow ProtocolProperty to be used with the view package

// GetName returns the property name
func (p *ProtocolProperty) GetName() string {
	return p.Name
}

// GetFullName returns the full property name
func (p *ProtocolProperty) GetFullName() string {
	return p.FullName
}

// GetType returns the property type
func (p *ProtocolProperty) GetType() string {
	return p.Type
}

// GetValue returns the property value
func (p *ProtocolProperty) GetValue() string {
	return p.Value
}

// GetChildren returns child properties as an interface slice
func (p *ProtocolProperty) GetChildren() []interface{} {
	children := make([]interface{}, len(p.Children))
	for i := range p.Children {
		children[i] = &p.Children[i]
	}
	return children
}

// HasChildren returns true if property has children
func (p *ProtocolProperty) HasChildren() bool {
	if p.NumChildren == "" {
		return len(p.Children) > 0
	}
	count, err := strconv.Atoi(p.NumChildren)
	if err != nil {
		return len(p.Children) > 0
	}
	return count > 0
}

// GetNumChildren returns the number of children
func (p *ProtocolProperty) GetNumChildren() int {
	if p.NumChildren == "" {
		return len(p.Children)
	}
	count, err := strconv.Atoi(p.NumChildren)
	if err != nil {
		return len(p.Children)
	}
	return count
}

// View adapter methods for ProtocolStack
// These methods allow ProtocolStack to be used with the view package

// GetWhere returns the function/method name
func (s *ProtocolStack) GetWhere() string {
	return s.Where
}

// GetLevel returns the stack level as int
func (s *ProtocolStack) GetLevel() int {
	if s.Level == "" {
		return 0
	}
	level, err := strconv.Atoi(s.Level)
	if err != nil {
		return 0
	}
	return level
}

// GetType returns the stack frame type
func (s *ProtocolStack) GetType() string {
	return s.Type
}

// GetFilename returns the filename without file:// prefix
func (s *ProtocolStack) GetFilename() string {
	// Strip file:// prefix if present
	filename := s.Filename
	if len(filename) > 7 && filename[:7] == "file://" {
		return filename[7:]
	}
	return filename
}

// GetLineNumber returns the line number as int
func (s *ProtocolStack) GetLineNumber() int {
	if s.Lineno == "" {
		return 0
	}
	line, err := strconv.Atoi(s.Lineno)
	if err != nil {
		return 0
	}
	return line
}
