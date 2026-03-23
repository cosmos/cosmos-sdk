package internal

import (
	"fmt"
)

// MutationContext is a small helper that keeps track of the current version and orphaned nodes.
// It is used in all mutation operations such as insertion, deletion, and rebalancing.
type MutationContext struct {
	version    uint32
	cowVersion uint32
	orphans    []*NodePointer
}

// NewMutationContext creates a new MutationContext for the given version.
// The cowVersion indicates the version below which copy-on-write should occur.
// Any nodes with version less than cowVersion will be copied when mutated,
// while nodes with version greater than or equal to cowVersion will be reused as-is.
func NewMutationContext(version, cowVersion uint32) *MutationContext {
	return &MutationContext{
		version:    version,
		cowVersion: cowVersion,
	}
}

// mutateBranch mutates the given branch node for the current version
// and tracks the existing node as an orphan.
// If the node's ID is from the current version or is empty (meaning it hasn't been persisted yet),
// the node is returned as-is without mutation or orphan tracking.
func (ctx *MutationContext) mutateBranch(node Node, nodePtr *NodePointer) (*MemNode, error) {
	if node.Version() >= ctx.cowVersion {
		memNode, ok := node.(*MemNode)
		if ok {
			// node is not shared, can mutate in place but must update version and clear hash
			memNode.version = ctx.version
			memNode.hash = nil
			return memNode, nil
		}
	}
	ctx.addOrphan(nodePtr)
	mem, err := node.MutateBranch(ctx.version)
	if err != nil {
		return nil, fmt.Errorf("mutate branch node %s: %w", node.ID(), err)
	}
	return mem, nil
}

// addOrphan adds the given node's ID to the list of orphaned nodes.
// This is to be called when a node is deleted without being replaced; use mutateBranch for nodes that are replaced.
// Only nodes with a version older than the current mutation version are considered orphans.
// Nodes with version 0 are uncommitted and don't need orphan tracking since they were never persisted.
// Returns true if the node was a valid orphan, false otherwise.
func (ctx *MutationContext) addOrphan(nodePtr *NodePointer) bool {
	if nodePtr == nil {
		return false
	}
	// we only track orphans for nodes with a valid checkpoint
	if nodePtr.id.Checkpoint() > 0 {
		ctx.orphans = append(ctx.orphans, nodePtr)
		return true
	}
	return false
}

// NewLeafNode creates a new leaf MemNode with the given key, value, and version.
func (ctx *MutationContext) NewLeafNode(key, value []byte) *MemNode {
	return &MemNode{
		height:  0,
		size:    1,
		key:     key,
		value:   value,
		version: ctx.version,
	}
}

func (ctx *MutationContext) newBranchNode(left, right *NodePointer, key []byte) *MemNode {
	return &MemNode{
		height:  1,
		size:    2,
		version: ctx.version,
		left:    left,
		right:   right,
		key:     key,
	}
}
