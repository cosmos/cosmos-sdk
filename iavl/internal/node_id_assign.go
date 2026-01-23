package internal

// AssignNodeIDs assigns unique NodeIDs to all nodes in the tree rooted at root.
// Leaf nodes receive IDs first in in-order traversal, followed by branch nodes in post-order traversal.
// The checkpoint parameter specifies the checkpoint (version) at which these nodes are being assigned IDs.
// This function must complete before any code that must read IDs can run.
func AssignNodeIDs(root *NodePointer, checkpoint uint32) {
	ctx := &nodeIdContext{}
	assignNodeIDsRecursive(root, checkpoint, ctx)
}

type nodeIdContext struct {
	lastLeafIdx   uint32
	lastBranchIdx uint32
}

func assignNodeIDsRecursive(np *NodePointer, checkpoint uint32, ctx *nodeIdContext) {
	if np == nil {
		return
	}
	memNode := np.Mem.Load()
	if memNode == nil {
		return
	}

	if !memNode.nodeId.IsEmpty() {
		return
	}

	if memNode.IsLeaf() {
		ctx.lastLeafIdx++
		id := NewNodeID(true, checkpoint, ctx.lastLeafIdx)
		memNode.nodeId = id
		np.id = id
		return
	}

	// assign node IDs in post-order traversal
	assignNodeIDsRecursive(memNode.left, checkpoint, ctx)
	assignNodeIDsRecursive(memNode.right, checkpoint, ctx)

	ctx.lastBranchIdx++
	id := NewNodeID(false, checkpoint, ctx.lastBranchIdx)
	memNode.nodeId = id
	np.id = id
}
