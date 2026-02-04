package internal

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type Evictor interface {
	Evict(root *NodePointer, checkpoint uint32)
}

type BasicEvictor struct {
	EvictDepth uint8
}

func (be BasicEvictor) Evict(root *NodePointer, checkpoint uint32) {
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
		_, span := tracer.Start(context.Background(), "Evict",
			trace.WithAttributes(
				attribute.Int("evictDepth", int(be.EvictDepth)),
				attribute.Int("treeHeight", int(height)),
			),
		)
		defer span.End()

		count := evictTraverse(root, 0, be.EvictDepth, checkpoint)

		span.SetAttributes(
			attribute.Int("nodesEvicted", count),
		)
	}()
}

func evictTraverse(np *NodePointer, depth, evictionDepth uint8, evictCheckpoint uint32) (count int) {
	memNode := np.Mem.Load()
	if memNode == nil {
		return 0
	}

	nodeCheckpoint := memNode.nodeId.checkpoint
	if nodeCheckpoint == 0 || nodeCheckpoint > evictCheckpoint {
		panic(fmt.Sprintf("fatal logic error: evictTraverse reached node %s with invalid checkpoint %d (evictCheckpoint %d)", memNode.nodeId.String(), nodeCheckpoint, evictCheckpoint))
	}

	// evict nodes at or below the eviction depth
	if depth >= evictionDepth {
		if np.changeset == nil {
			panic(fmt.Sprintf("fatal logic error: nnot evict node %s at checkpoint %d without changeset", memNode.nodeId.String(), nodeCheckpoint))
		}
		np.Mem.Store(nil)
		count = 1
	}

	if memNode.IsLeaf() {
		return count
	}

	// continue traversing to find nodes to evict
	count += evictTraverse(memNode.left, depth+1, evictionDepth, evictCheckpoint)
	count += evictTraverse(memNode.right, depth+1, evictionDepth, evictCheckpoint)
	return count
}
