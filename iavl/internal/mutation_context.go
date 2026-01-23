package internal

import (
	"fmt"
	"unsafe"
)

// MutationContext is a small helper that keeps track of the current version and orphaned nodes.
// It is used in all mutation operations such as insertion, deletion, and rebalancing.
type MutationContext struct {
	version  uint32
	orphans  []NodeID
	memUsage int64
}

// NewMutationContext creates a new MutationContext for the given version.
func NewMutationContext(version uint32) *MutationContext {
	return &MutationContext{
		version: version,
	}
}

// mutateBranch mutates the given branch node for the current version
// and tracks the existing node as an orphan.
// If the node's ID is from the current version or is empty (meaning it hasn't been persisted yet),
// the node is returned as-is without mutation or orphan tracking.
// NOTE: if we do decide to implement nested cache wrapper functionality
// directly using the IAVL tree structures (instead of a btree wrapper),
// then we MUST change this code to ALWAYS mutate nodes here,
// even if they are from the current version.
func (ctx *MutationContext) mutateBranch(node Node) (*MemNode, error) {
	if node.Version() == ctx.version {
		// node is already at the current version; no mutation needed.
		memNode, ok := node.(*MemNode)
		if !ok {
			return nil, fmt.Errorf("expected MemNode, got %T", node)
		}
		return memNode, nil
	}
	ctx.addOrphanId(node.ID())
	mem, err := node.MutateBranch(ctx.version)
	if err != nil {
		return nil, fmt.Errorf("mutate branch node %s: %w", node.ID(), err)
	}
	ctx.memUsage += memNodeOverhead + int64(len(mem.key))
	return mem, nil
}

func (ctx *MutationContext) addOrphan(nodePtr *NodePointer) bool {
	if nodePtr == nil {
		return false
	}
	mem := nodePtr.Mem.Load()
	if mem != nil {
		ctx.memUsage -= memNodeOverhead + int64(len(mem.key)) + int64(len(mem.value))
	}
	return ctx.addOrphanId(nodePtr.id)
}

// addOrphan adds the given node's ID to the list of orphaned nodes.
// This is to be called when a node is deleted without being replaced; use mutateBranch for nodes that are replaced.
// Only nodes with a version older than the current mutation version are considered orphans.
// Nodes with version 0 are uncommitted and don't need orphan tracking since they were never persisted.
// Returns true if the node was a valid orphan, false otherwise.
func (ctx *MutationContext) addOrphanId(id NodeID) bool {
	// checkpoint == version, so this gives us the version at which the node was persisted
	checkpoint := id.Checkpoint()
	if checkpoint > 0 && checkpoint < ctx.version {
		ctx.orphans = append(ctx.orphans, id)
		return true
	}
	return false
}

// NewLeafNode creates a new leaf MemNode with the given key, value, and version.
func (ctx *MutationContext) NewLeafNode(key, value []byte) *MemNode {
	ctx.memUsage += memNodeOverhead + int64(len(key)) + int64(len(value))
	return &MemNode{
		height:  0,
		size:    1,
		key:     key,
		value:   value,
		version: ctx.version,
	}
}

func (ctx *MutationContext) newBranchNode(left, right *NodePointer, key []byte) *MemNode {
	ctx.memUsage += memNodeOverhead + int64(len(key))
	return &MemNode{
		height:  1,
		size:    2,
		version: ctx.version,
		left:    left,
		right:   right,
		key:     key,
	}
}

const memNodeOverhead = int64(unsafe.Sizeof(MemNode{})) + int64(unsafe.Sizeof(NodePointer{}))*2 + 32 /* hash size */
