package internal

import (
	"fmt"
	"sync/atomic"
)

// NodePointer is a pointer to a Node, which may be either in-memory, on-disk or both.
type NodePointer struct {
	Mem       atomic.Pointer[MemNode]
	changeset *Changeset
	fileIdx   uint32 // absolute index in file, 1-based, zero means we don't have an offset
	id        NodeID
}

// NewNodePointer creates a new NodePointer pointing to the given in-memory node.
func NewNodePointer(memNode *MemNode) *NodePointer {
	n := &NodePointer{}
	n.Mem.Store(memNode)
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
	mem := p.Mem.Load()
	if mem != nil {
		return mem, NoopPin{}, nil
	}
	if p.fileIdx != 0 {
		rdr, pin := p.changeset.TryPinUncompactedReader()
		if rdr != nil {
			node, err := rdr.ResolveByFileIndex(p.id, p.fileIdx)
			return node, pin, err
		} else {
			compacted := p.changeset.Compacted()
			if compacted == nil {
				return nil, NoopPin{}, fmt.Errorf("unable to pin ChangesetReader for checkpoint %d, no compaction found", p.id.Checkpoint())
			}
			rdr, pin := compacted.TryPinReader()
			if rdr != nil {
				node, err := rdr.ResolveByID(p.id)
				return node, pin, err
			}
			return nil, NoopPin{}, fmt.Errorf("unable to pin ChangesetReader for checkpoint %d, likely it has been closed and can't be reopened", p.id.Checkpoint())
		}
	} else {
		cs := p.changeset.TreeStore().ChangesetForCheckpoint(p.id.Checkpoint())
		if cs == nil {
			return nil, NoopPin{}, fmt.Errorf("unable to find Changeset for checkpoint %d", p.id.Checkpoint())
		}
		rdr, pin := cs.TryPinReader()
		if rdr != nil {
			node, err := rdr.ResolveByID(p.id)
			return node, pin, err
		} else {
			return nil, NoopPin{}, fmt.Errorf("unable to pin ChangesetReader for checkpoint %d, likely it has been closed and can't be reopened", p.id.Checkpoint())
		}
	}
}

// String implements the fmt.Stringer interface.
func (p *NodePointer) String() string {
	return fmt.Sprintf("NodePointer{id: %s, fileIdx: %d}", p.id.String(), p.fileIdx)
}
