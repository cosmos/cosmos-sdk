package iavlx

import (
	"context"
	"sync/atomic"
)

type evictor struct {
	*TreeStore
	wakeChan chan struct{}
}

func newEvictor(ts *TreeStore, ctx context.Context) *evictor {
	ev := &evictor{
		TreeStore: ts,
		wakeChan:  make(chan struct{}, 1),
	}
	go ev.evictLoop(ctx)
	return ev
}

func (ev *evictor) wake() {
	select {
	case ev.wakeChan <- struct{}{}:
	default: // already woken
	}
}

func (ev *evictor) evictLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-ev.wakeChan:
			ev.processEviction()
		}
	}
}

func (ev *evictor) processEviction() {
	_, span := tracer.Start(context.Background(), "processEviction")
	defer span.End()

	tracker := &evictTracker{
		budget: ev.memMonitor.EvictBudget(),
	}

	for {
		if tracker.done() {
			return
		}

		ev.latestMapLock.RLock()
		version, tree, ok := ev.latest.Min()
		ev.latestMapLock.RUnlock()
		evictVersion := ev.savedVersion.Load()
		if !ok || version > evictVersion {
			// no more versions to evict
			return
		}

		evictTraverse(tree, tracker, evictVersion)

		if tree.mem.Load() == nil {
			// if we fully evicted this version, remove it from latest map
			ev.latestMapLock.Lock()
			ev.latest.Delete(version)
			ev.latestMapLock.Unlock()
		}
	}
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

	if !memNode.IsLeaf() {
		if !evictTraverse(memNode.left, tracker, evictVersion) {
			return false
		}
		if !evictTraverse(memNode.right, tracker, evictVersion) {
			return false
		}
	}

	if memNode.version <= evictVersion {
		np.mem.Store(nil)
		if !tracker.trackEvict(memNode) {
			return false
		}
	}

	return true
}
