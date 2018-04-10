package types

import ()

type Codespacer struct {
	next     CodespaceType
	reserved map[CodespaceType]bool
}

// NewCodespacer generates a new Codespacer with the starting codespace
func NewCodespacer() *Codespacer {
	return &Codespacer{
		next:     CodespaceType(1),
		reserved: make(map[CodespaceType]bool),
	}
}

// Reserve reserves a specified codespace
func (c *Codespacer) reserve(codespace CodespaceType) {
	if codespace == c.next {
		c.next++
	} else if codespace < c.next || c.reserved[codespace] {
		panic("Cannot reserve codespace - already taken!")
	} else {
		c.reserved[codespace] = true
	}
}

// Register registers a provided codespace
func (c *Codespacer) Register(codespace CodespaceType) {
	c.reserve(codespace)
}

// RegisterDefault registers and returns the next available codespace
func (c *Codespacer) RegisterDefault() CodespaceType {
	def := c.next
	c.reserve(def)
	return def
}
