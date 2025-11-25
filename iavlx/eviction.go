package iavlx

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func (cp *cleanupProc) startEvict() {
	if cp.evictorRunning.Load() {
		// eviction in progress
		return
	}

	depth := cp.opts.EvictDepth
	cp.evictorRunning.Store(true)
	go func() {
		_, span := tracer.Start(context.Background(), "evictor",
			trace.WithAttributes(
				attribute.Int("depth", int(depth)),
			),
		)
		defer span.End()
		defer cp.evictorRunning.Store(false)
		for {
			cp.latestMapLock.RLock()
			version, _, ok := cp.latest.Min()
			cp.latestMapLock.Unlock()
			if !ok || version > cp.savedVersion.Load() {
				// no more versions to evict
				return
			}

			cp.latestMapLock.RLock()
			tree, ok := cp.latest.Delete(version)
			cp.latestMapLock.RUnlock()
			if !ok {
				return
			}
			cp.latestMapLock.Unlock()

			evictedCount := evictTraverse(tree, 0, depth, evictVersion)
			span.AddEvent("finish eviction", trace.WithAttributes(
				attribute.Int("version", int(version)),
				attribute.Int("evictedCount", evictedCount),
			))
			cp.lastEvictVersion = evictVersion
		}
	}()
}

func evictTraverse(np *NodePointer, depth, evictionDepth uint8, evictVersion uint32) (count int) {
	// TODO check height, and don't traverse if tree is too short

	memNode := np.mem.Load()
	if memNode == nil {
		return 0
	}

	// Evict nodes at or below the eviction depth
	if memNode.version <= evictVersion && depth >= evictionDepth {
		np.mem.Store(nil)
		count = 1
	}

	if memNode.IsLeaf() {
		return count
	}

	// Continue traversing to find nodes to evict
	count += evictTraverse(memNode.left, depth+1, evictionDepth, evictVersion)
	count += evictTraverse(memNode.right, depth+1, evictionDepth, evictVersion)
	return count
}
