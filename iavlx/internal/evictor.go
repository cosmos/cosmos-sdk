package internal

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// Evictor controls which in-memory tree nodes are evicted after a checkpoint is saved.
//
// After a checkpoint is written to disk, all nodes from that checkpoint can be resolved from
// disk instead of memory. Evicting them frees heap memory at the cost of future disk reads
// (which may hit the OS mmap cache and still be fast).
//
// Eviction is depth-based: nodes deeper than a configured threshold are evicted, while nodes
// near the root are kept in memory. This is a good heuristic because:
//   - Shallow nodes are accessed on every read/write (they're on the path from root to any key).
//   - Deep nodes are only accessed for specific key ranges and can be loaded on demand.
//   - Leaf nodes can be evicted more aggressively than branch nodes because branch nodes
//     are needed just to navigate to the right subtree, while leaf nodes are only needed
//     for the final key/value read.
//
// The eviction depths are configured via TreeOptions.LeafEvictDepth and BranchEvictDepth.
// Higher values retain more nodes in memory (better read performance, more memory usage).
type Evictor interface {
	Evict(root *NodePointer, checkpoint uint32)
}

// BasicEvictor evicts nodes beyond a configurable depth threshold after each checkpoint.
// It runs asynchronously in a goroutine to avoid blocking the commit path.
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
