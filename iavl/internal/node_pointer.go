package internal

import (
	"fmt"
	"sync/atomic"
)

// NodePointer is a pointer to a Node, which may be either in-memory, on-disk or both.
type NodePointer struct {
	mem       atomic.Pointer[MemNode]
	changeset *Changeset
	fileIdx   uint32 // absolute index in file, 1-based, zero means we don't have an offset
	id        NodeID
}

// NewNodePointer creates a new NodePointer pointing to the given in-memory node.
func NewNodePointer(memNode *MemNode) *NodePointer {
	n := &NodePointer{}
	n.mem.Store(memNode)
	return n
}

// Resolve resolves the NodePointer to a Node, loading from memory or disk as necessary.
func (p *NodePointer) Resolve() (Node, Pin, error) {
	mem := p.mem.Load()
	if mem != nil {
		return mem, NoopPin{}, nil
	}
	panic("not implemented: loading from disk")
}

// String implements the fmt.Stringer interface.
func (p *NodePointer) String() string {
	return fmt.Sprintf("NodePointer{id: %s, fileIdx: %d}", p.id.String(), p.fileIdx)
}
