package internal

import "fmt"

// MutationContext is a small helper that keeps track of the current version and orphaned nodes.
// It is used in all mutation operations such as insertion, deletion, and rebalancing.
type MutationContext struct {
	version uint32
	orphans []NodeID
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
	if ctx.addOrphan(node.ID()) {
		return node.MutateBranch(ctx.version)
	}
	memNode, ok := node.(*MemNode)
	if !ok {
		return nil, fmt.Errorf("expected MemNode, got %T", node)
	}
	return memNode, nil
}

// addOrphan adds the given node's ID to the list of orphaned nodes.
// This is to be called when a node is deleted without being replaced; use mutateBranch for nodes that are replaced.
// Only nodes with a version older than the current mutation version are considered orphans.
// Nodes with version 0 are uncommitted and don't need orphan tracking since they were never persisted.
// Returns true if the node was added as an orphan, false otherwise.
func (ctx *MutationContext) addOrphan(id NodeID) bool {
	version := id.Version()
	if version > 0 && version < ctx.version {
		ctx.orphans = append(ctx.orphans, id)
		return true
	}
	return false
}
