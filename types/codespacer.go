package sdk

//----------------------------------------
// Codespacer

// Codespacer is a simple struct to track reserved codespaces
type Codespacer struct {
	reserved map[CodespaceType]bool
}

// NewCodespacer generates a new Codespacer with the starting codespace
func NewCodespacer() *Codespacer {
	c := &Codespacer{
		reserved: make(map[CodespaceType]bool),
	}
	c.reserved[CodespaceUndefined] = true
	c.reserved[CodespaceRoot] = true

	return c
}

// RegisterNext reserves and returns the next available codespace, starting from a default, and panics if the maximum codespace is reached
func (c *Codespacer) RegisterNext(codespace CodespaceType) CodespaceType {
	for {
		if !c.reserved[codespace] {
			c.reserved[codespace] = true
			return codespace
		}
		codespace++
		if codespace == MaximumCodespace {
			panic("Maximum codespace reached!")
		}
	}
}

// RegisterOrPanic reser ved a codespace or panics if it is unavailable
func (c *Codespacer) RegisterOrPanic(codespace CodespaceType) {
	if c.reserved[codespace] {
		panic("Cannot register codespace, already reserved")
	}
	c.reserved[codespace] = true
}
