package iavlx

import (
	"fmt"
	"sync/atomic"
)

type NodePointer struct {
	mem     atomic.Pointer[MemNode]
	store   *Changeset
	fileIdx uint32 // absolute index in file, 1-based, zero means we don't have an offset
	id      NodeID
}

func NewNodePointer(memNode *MemNode) *NodePointer {
	n := &NodePointer{}
	n.mem.Store(memNode)
	return n
}

func (p *NodePointer) Resolve() (Node, error) {
	mem := p.mem.Load()
	if mem != nil {
		return mem, nil
	}
	return p.store.Resolve(p.id, p.fileIdx)
}

func (p *NodePointer) String() string {
	return fmt.Sprintf("NodePointer{id: %s, fileIdx: %d}", p.id.String(), p.fileIdx)
}
