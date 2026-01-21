package internal

type Evictor interface {
	Evict(root *NodePointer, evictLayer uint32) (count int)
}

type BasicEvictor struct {
	EvictDepth uint8
}

func (be BasicEvictor) Evict(root *NodePointer, evictLayer uint32) (count int) {
	mem := root.Mem.Load()
	if mem == nil {
		return 0
	}
	if mem.Height() < be.EvictDepth {
		// shortcut when tree is too short
		return 0
	}
	return evictTraverse(root, 0, be.EvictDepth, evictLayer)
}

func evictTraverse(np *NodePointer, depth, evictionDepth uint8, evictLayer uint32) (count int) {
	// TODO check height, and don't traverse if tree is too short

	memNode := np.Mem.Load()
	if memNode == nil {
		return 0
	}

	// Evict nodes at or below the eviction depth
	if memNode.nodeId.layer <= evictLayer && depth >= evictionDepth {
		np.Mem.Store(nil)
		count = 1
	}

	if memNode.IsLeaf() {
		return count
	}

	// Continue traversing to find nodes to evict
	count += evictTraverse(memNode.left, depth+1, evictionDepth, evictLayer)
	count += evictTraverse(memNode.right, depth+1, evictionDepth, evictLayer)
	return count
}
