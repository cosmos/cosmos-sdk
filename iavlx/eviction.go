package iavlx

import (
	"context"
	"sync/atomic"

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

type evictTracker struct {
	evictedBytes int
	budget       *atomic.Int64
}

func (t *evictTracker) trackEvict(memNode *MemNode) bool {
	var sz int
	if memNode.IsLeaf() {
		sz = SizeLeaf + len(memNode.key) + len(memNode.value)
	} else {
		sz = SizeBranch + len(memNode.key)
	}
	t.evictedBytes += sz
	if t.budget.Add(int64(-sz)) <= 0 {
		return false
	}
	return true
}

func (t *evictTracker) done() bool {
	return t.budget.Load() <= 0
}

func evictTraverse(np *NodePointer, tracker *evictTracker, evictVersion uint32) bool {
	if tracker.done() {
		return false
	}

	memNode := np.mem.Load()
	if memNode == nil {
		return true
	}

	if !evictTraverse(memNode.left, tracker, evictVersion) {
		return false
	}
	if !evictTraverse(memNode.right, tracker, evictVersion) {
		return false
	}

	if memNode.version <= evictVersion {
		np.mem.Store(nil)
		if !tracker.trackEvict(memNode) {
			return false
		}
	}

	return true
}
