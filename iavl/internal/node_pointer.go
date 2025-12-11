package internal

import (
	"fmt"
	"sync/atomic"
)

// NodePointer is a pointer to a Node, which may be either in-memory, on-disk or both.
type NodePointer struct {
	mem atomic.Pointer[MemNode]
	// changeset *Changeset // commented to satisfy linter, will uncomment in a future PR when we wire it up
	fileIdx uint32 // absolute index in file, 1-based, zero means we don't have an offset
	id      NodeID
}

// NewNodePointer creates a new NodePointer pointing to the given in-memory node.
func NewNodePointer(memNode *MemNode) *NodePointer {
	n := &NodePointer{}
	n.mem.Store(memNode)
	return n
}

// Resolve resolves the NodePointer to a Node, loading from memory or disk as necessary
// as well as a Pin which MUST be unpinned after the caller is done using the node.
// Resolve will ALWAYS return a valid Pin even if there is an error. For clarity and
// consistency it is recommended to introduce a defer pin.Unpin() immediately after
// calling Resolve and BEFORE checking the error return value like this:
//
//	node, pin, err := nodePointer.Resolve()
//	defer pin.Unpin()
//	if err != nil {
//	    // handle error
//	}
func (p *NodePointer) Resolve() (Node, Pin, error) {
	mem := p.mem.Load()
	if mem != nil {
		return mem, NoopPin{}, nil
	}
	return nil, NoopPin{}, fmt.Errorf("node not in memory and on-disk loading will be implemented in a future PR")
}

// String implements the fmt.Stringer interface.
func (p *NodePointer) String() string {
	return fmt.Sprintf("NodePointer{id: %s, fileIdx: %d}", p.id.String(), p.fileIdx)
}
