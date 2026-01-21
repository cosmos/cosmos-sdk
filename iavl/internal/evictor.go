package internal

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type Evictor interface {
	Evict(root *NodePointer, evictLayer uint32)
}

type BasicEvictor struct {
	EvictDepth uint8
}

func (be BasicEvictor) Evict(root *NodePointer, evictLayer uint32) {
	if root == nil {
		return
	}
	mem := root.Mem.Load()
	if mem == nil {
		return
	}
	height := mem.Height()
	if height < be.EvictDepth {
		// shortcut when tree is too short
		return
	}

	go func() {
		_, span := Tracer.Start(context.Background(), "Evict",
			trace.WithAttributes(
				attribute.Int("evictDepth", int(be.EvictDepth)),
				attribute.Int("treeHeight", int(height)),
			),
		)
		defer span.End()

		count := evictTraverse(root, 0, be.EvictDepth, evictLayer)

		span.SetAttributes(
			attribute.Int("nodesEvicted", count),
		)
	}()
}

func evictTraverse(np *NodePointer, depth, evictionDepth uint8, evictLayer uint32) (count int) {
	memNode := np.Mem.Load()
	if memNode == nil {
		return 0
	}

	if memNode.nodeId.layer == 0 {
		// node has not been assigned an ID yet, so cannot be evicted
		return 0
	}

	// evict nodes at or below the eviction depth
	if memNode.nodeId.layer <= evictLayer && depth >= evictionDepth {
		np.Mem.Store(nil)
		count = 1
	}

	if memNode.IsLeaf() {
		return count
	}

	// continue traversing to find nodes to evict
	count += evictTraverse(memNode.left, depth+1, evictionDepth, evictLayer)
	count += evictTraverse(memNode.right, depth+1, evictionDepth, evictLayer)
	return count
}
