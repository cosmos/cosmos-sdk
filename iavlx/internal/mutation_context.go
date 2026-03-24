package internal

import (
	"fmt"
)

// MutationContext is a small helper that keeps track of the current version and orphaned nodes.
// It is used in all mutation operations such as insertion, deletion, and rebalancing.
type MutationContext struct {
	// version is the version new nodes will be created with
	version uint32
	// cowVersion is the version below which nodes should be copied when they are mutated
	// That means nodes with version >= cowVersion will be mutated in places as an optimization
	cowVersion uint32
	// orphans are the orphaned node pointers that are collected when we mutate existing nodes
	orphans []*NodePointer
}

// NewMutationContext creates a new MutationContext for the given version.
// The cowVersion indicates the version below which copy-on-write should occur.
// Any nodes with version less than cowVersion will be copied when mutated,
// while nodes with version greater than or equal to cowVersion will be reused as-is.
// Callers of NewMutationContext MUST ensure that cowVersion is set correctly otherwise data WILL get corrupted.
// If you are unsure even a tiny bit, set cowVersion to version.
// NOTE: if cowVersion is set to a value higher than version, then the IAVL tree could be used as a cache layer
// (instead of the B-tree used in cachekv) where we will always copy nodes even if they are from the current version.
// This would allow us to have nested caches and rollbacks because we would assume that it isn't safe to mutate any nodes.
func NewMutationContext(version, cowVersion uint32) *MutationContext {
	return &MutationContext{
		version:    version,
		cowVersion: cowVersion,
	}
}

// mutateBranch should be used when we have an existing branch which we want to "mutate"
// such that it has different left and right children.
// "Mutation" in an immutable tree generally means copying a node and returning a new node
// because the tree is "immutable".
// However, if we know that the node hasn't been shared with any other consumers yet,
// we can safely mutate it in place as an optimization.
// If the node's version is >= the cowVersion (copy-on-write threshold version) and it is a *MemNode,
// then it will be mutated in place and returned with its version field updated and hash cleared.
// If the node is not being mutated in-place (the ordinary case), a safe copy will be returned
// with the current version set and hash cleared, and the node will be tracked as an orphan
// if it has been persisted to disk already.
func (ctx *MutationContext) mutateBranch(node Node, nodePtr *NodePointer) (*MemNode, error) {
	if node.Version() >= ctx.cowVersion {
		memNode, ok := node.(*MemNode)
		if ok {
			memNode.version = ctx.version
			memNode.hash = nil
			return memNode, nil
		}
	}
	// Track the node as an orphan (if needed).
	ctx.addOrphan(nodePtr)
	// Safely copy the branch using the node's MutateBranch method.
	mem, err := node.MutateBranch(ctx.version)
	if err != nil {
		return nil, fmt.Errorf("mutate branch node %s: %w", node.ID(), err)
	}
	return mem, nil
}

// addOrphan adds the given node's ID to the list of orphaned nodes.
// This is to be called when a node is deleted without being replaced; use mutateBranch for nodes that are replaced.
// Only nodes with a valid checkpoint number in their NodeID are tracked as orphans.
// When a node has a valid NodeID with a valid checkpoint, this implies that the node has been persisted to disk in a checkpoint
// (or will be soon in a background thread), and thus there is some on-disk data to clean up.
// If there is no checkpoint number, and we are orphaning the node, the node was never saved on disk anyway,
// so the only garbage collection is in-memory, and we can let the go GC take care of that.
// When checkpoints are taken, NodeID's are ALWAYS set for the checkpoint before the next version is created,
// so if a node is missing a valid NodeID, it definitively means that it is not stored on disk in any checkpoint.
func (ctx *MutationContext) addOrphan(nodePtr *NodePointer) bool {
	if nodePtr == nil {
		return false
	}
	// We only track orphans for nodes with a valid checkpoint.
	// Nodes without a valid checkpoint aren't persisted on disk anyway, so nothing to clean up.
	if nodePtr.id.Checkpoint() > 0 {
		ctx.orphans = append(ctx.orphans, nodePtr)
		return true
	}
	return false
}

// NewLeafNode is a helper for creating a new leaf MemNode at the correct version.
func (ctx *MutationContext) NewLeafNode(key, value []byte) *MemNode {
	return &MemNode{
		height:  0,
		size:    1,
		key:     key,
		value:   value,
		version: ctx.version,
	}
}

// newBranchNode is a helper for creating a new branch MemNode at the correct version.
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
