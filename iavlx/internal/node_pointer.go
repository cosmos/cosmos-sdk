package internal

import (
	"fmt"
	"sync/atomic"
)

// NodePointer is the central indirection for tree nodes. A node can exist in three states:
//
//   - In-memory only: freshly created during a commit, not yet checkpointed. Mem is set,
//     changeset/fileIdx are zero. The node lives purely on the heap.
//
//   - In-memory + on-disk: the node has been checkpointed (written to a changeset's data files)
//     but is still cached in memory. Mem is set, changeset and fileIdx point to the on-disk copy.
//     Reads use the fast in-memory path; the on-disk copy is the fallback after eviction.
//
//   - On-disk only: the evictor has cleared Mem (set to nil) to free heap memory. The node must
//     be resolved from disk via the changeset + fileIdx (O(1) lookup) or changeset + NodeID
//     (requires checkpoint metadata lookup). This is the cold path.
//
// The Resolve method handles all three cases transparently. It also handles the case where the
// node's changeset has been compacted — if the original changeset's reader is evicted, Resolve
// follows the compaction pointer and falls back to resolving by NodeID in the compacted changeset
// (since file offsets change during compaction).
type NodePointer struct {
	// Mem holds the in-memory node, or nil if the node has been evicted.
	// Accessed atomically because the evictor clears it from a background goroutine.
	Mem atomic.Pointer[MemNode]
	// changeset is the changeset that contains this node's on-disk data. Nil for nodes that
	// haven't been checkpointed yet.
	changeset *Changeset
	// fileIdx is the 1-based offset into the changeset's leaves or branches file.
	// Zero means we don't have a direct file offset for this node in its changeset. This happens
	// when the node is in a different changeset than its parent (cross-changeset reference) — the
	// parent's BranchLayout only stores offsets for children in the same changeset. It also happens
	// for root NodePointers constructed during checkpoint lookup.
	// When fileIdx is zero, Resolve falls back to looking up the node by its NodeID via the
	// checkpoint metadata.
	fileIdx uint32
	// id is the stable identifier for this node (checkpoint + leaf/branch flag + index).
	id NodeID
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
// Resolve has three resolution paths depending on the node's state:
//  1. In-memory (fast path): Mem is set, return it directly. No disk IO.
//  2. On-disk with file offset: fileIdx is set, do O(1) lookup in the original changeset.
//     If the original changeset was compacted (reader evicted), follow the compacted pointer
//     and resolve by NodeID in the compacted changeset (file offsets change during compaction).
//  3. On-disk without file offset: the node is a cross-changeset reference (child in a different
//     changeset than its parent) or a root pointer from checkpoint lookup. We find the correct
//     changeset by checkpoint number and resolve by NodeID.
func (p *NodePointer) Resolve() (Node, Pin, error) {
	// Path 1: in-memory fast path
	mem := p.Mem.Load()
	if mem != nil {
		return mem, Pin{}, nil
	}
	if p.fileIdx != 0 {
		// Path 2: have a direct file offset — try O(1) lookup in the original changeset
		rdr, pin := p.changeset.TryPinUncompactedReader()
		if rdr != nil {
			node, err := rdr.ResolveByFileIndex(p.id, p.fileIdx)
			return node, pin, err
		} else {
			// Original changeset was compacted (reader evicted).
			// Follow the compacted pointer and resolve by NodeID since file offsets
			// changed during compaction.
			compacted := p.changeset.Compacted()
			if compacted == nil {
				return nil, Pin{}, fmt.Errorf("unable to pin ChangesetReader for checkpoint %d, no compaction found", p.id.Checkpoint())
			}
			rdr, pin := compacted.TryPinReader()
			if rdr != nil {
				node, err := rdr.ResolveByID(p.id)
				return node, pin, err
			}
			return nil, Pin{}, fmt.Errorf("unable to pin ChangesetReader for checkpoint %d, likely it has been closed and can't be reopened", p.id.Checkpoint())
		}
	} else {
		// Path 3: no file offset — cross-changeset reference or root from checkpoint lookup.
		// Look up the changeset by checkpoint number and resolve by NodeID.
		cs := p.changeset.TreeStore().ChangesetForCheckpoint(p.id.Checkpoint())
		if cs == nil {
			return nil, Pin{}, fmt.Errorf("unable to find Changeset for checkpoint %d", p.id.Checkpoint())
		}
		rdr, pin := cs.TryPinReader()
		if rdr != nil {
			node, err := rdr.ResolveByID(p.id)
			return node, pin, err
		} else {
			return nil, Pin{}, fmt.Errorf("unable to pin ChangesetReader for checkpoint %d, likely it has been closed and can't be reopened", p.id.Checkpoint())
		}
	}
}

// String implements the fmt.Stringer interface.
func (p *NodePointer) String() string {
	return fmt.Sprintf("NodePointer{id: %s, fileIdx: %d}", p.id.String(), p.fileIdx)
}
