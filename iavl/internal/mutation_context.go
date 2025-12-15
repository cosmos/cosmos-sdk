package internal

import "fmt"

// mutationContext is a small helper that keeps track of the current mutation version and orphaned nodes.
// it is used in all mutation operations such as insertion, deletion, and rebalancing.
type mutationContext struct {
	version uint32
	orphans []NodeID
}

// mutateBranch mutates the given branch node for the current version and tracks the existing node as an orphan.
func (ctx *mutationContext) mutateBranch(node Node) (*MemNode, error) {
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
func (ctx *mutationContext) addOrphan(id NodeID) bool {
	version := id.Version()
	if version > 0 && version < ctx.version {
		ctx.orphans = append(ctx.orphans, id)
		return true
	}
	return false
}
