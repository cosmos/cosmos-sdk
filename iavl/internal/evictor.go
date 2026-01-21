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
	go func() {
		_, span := Tracer.Start(context.Background(), "Evict",
			trace.WithAttributes(attribute.Int("evictDepth", int(be.EvictDepth))),
		)
		defer span.End()

		mem := root.Mem.Load()
		if mem == nil {
			return
		}
		height := mem.Height()
		if height < be.EvictDepth {
			// shortcut when tree is too short
			return
		}
		count := evictTraverse(root, 0, be.EvictDepth, evictLayer)

		span.SetAttributes(
			attribute.Int("nodesEvicted", count),
			attribute.Int("treeHeight", int(height)),
		)
	}()
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
