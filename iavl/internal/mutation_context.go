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

// addOrphan adds the given node ID to the list of orphaned nodes.
// this is to be called when a node is deleted without being replaced, use mutateBranch for nodes that are replaced.
// an empty NodeID is ignored, which would be the case when a node is inserted transiently and dropped before it is
// even committed to a version.
func (ctx *mutationContext) addOrphan(id NodeID) {
	if !id.IsEmpty() {
		ctx.orphans = append(ctx.orphans, id)
	}
}
