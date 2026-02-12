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
	leafEvictDepth   uint8
	branchEvictDepth uint8
}

func NewBasicEvictor(leafEvictDepth, branchEvictDepth uint8) BasicEvictor {
	if branchEvictDepth < leafEvictDepth {
		branchEvictDepth = leafEvictDepth
	}
	return BasicEvictor{
		leafEvictDepth:   leafEvictDepth,
		branchEvictDepth: branchEvictDepth,
	}
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
	// use the smaller of the two depths for the shortcut check
	minDepth := be.leafEvictDepth
	if be.branchEvictDepth < minDepth {
		minDepth = be.branchEvictDepth
	}
	if height < minDepth {
		// shortcut when tree is too short
		return
	}

	go func() {
		_, span := tracer.Start(context.Background(), "Evict",
			trace.WithAttributes(
				attribute.Int("leafEvictDepth", int(be.leafEvictDepth)),
				attribute.Int("branchEvictDepth", int(be.branchEvictDepth)),
				attribute.Int("treeHeight", int(height)),
			),
		)
		defer span.End()

		count := evictTraverse(root, 0, be.leafEvictDepth, be.branchEvictDepth, checkpoint)

		span.SetAttributes(
			attribute.Int("nodesEvicted", count),
		)
	}()
}

func evictTraverse(np *NodePointer, depth, leafEvictDepth, branchEvictDepth uint8, evictCheckpoint uint32) (count int) {
	memNode := np.Mem.Load()
	if memNode == nil {
		return 0
	}

	nodeCheckpoint := memNode.nodeId.checkpoint
	if nodeCheckpoint == 0 || nodeCheckpoint > evictCheckpoint {
		panic(fmt.Sprintf("fatal logic error: evictTraverse reached node %s with invalid checkpoint %d (evictCheckpoint %d)", memNode.nodeId.String(), nodeCheckpoint, evictCheckpoint))
	}

	isLeaf := memNode.IsLeaf()
	evictDepth := branchEvictDepth
	if isLeaf {
		evictDepth = leafEvictDepth
	}

	// evict nodes at or beyond the eviction depth
	if depth >= evictDepth {
		if np.changeset == nil {
			panic(fmt.Sprintf("fatal logic error: cannot evict node %s at checkpoint %d without changeset", memNode.nodeId.String(), nodeCheckpoint))
		}
		np.Mem.Store(nil)
		count = 1
	}

	if isLeaf {
		return count
	}

	// continue traversing to find nodes to evict
	count += evictTraverse(memNode.left, depth+1, leafEvictDepth, branchEvictDepth, evictCheckpoint)
	count += evictTraverse(memNode.right, depth+1, leafEvictDepth, branchEvictDepth, evictCheckpoint)
	return count
}
