package iavlx

type evictor struct {
	*TreeStore
	wakeCh chan struct{}
}

func newEvictor(ts *TreeStore) *evictor {
	ev := &evictor{
		TreeStore: ts,
		wakeCh:    make(chan struct{}, 1),
	}
	go ev.evictLoop()
	return ev
}

func (ev *evictor) wake() {
	select {
	case ev.wakeCh <- struct{}{}:
	default:
	}
}

func (ev *evictor) evictLoop() {
	for {
		ev.latestMapLock.RLock()
		version, _, ok := ev.latest.Min()
		ev.latestMapLock.RUnlock()
		evictVersion := ev.savedVersion.Load()
		if !ok || version > evictVersion {
			select {
			case <-ev.wakeCh:
				continue
			case <-ev.ctx.Done():
				return
			}
		}

		ev.latest.Delete(version)
	}
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
