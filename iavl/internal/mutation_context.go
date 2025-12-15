package internal

// mutationContext is a small helper that keeps track of the current mutation version and orphaned nodes.
// it is used in all mutation operations such as insertion, deletion, and rebalancing.
type mutationContext struct {
	version uint32
	orphans []NodeID
}

// mutateBranch mutates the given branch node for the current version and tracks the existing node as an orphan.
func (ctx *mutationContext) mutateBranch(node Node) (*MemNode, error) {
	id := node.ID()
	if !id.IsEmpty() {
		ctx.orphans = append(ctx.orphans, id)
	}
	return node.MutateBranch(ctx.version)
}

// addOrphan adds the given node's ID to the list of orphaned nodes.
// this is to be called when a node is deleted without being replaced, use mutateBranch for nodes that are replaced.
// only nodes with an empty version or a version older than the current mutation version are considered orphans.
func (ctx *mutationContext) addOrphan(id NodeID) {
	version := id.Version()
	if version == 0 || version < ctx.version {
		ctx.orphans = append(ctx.orphans, id)
	}
}
