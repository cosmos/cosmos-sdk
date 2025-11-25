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
		if !ev.memMonitor.UnderPressure() {
			if !ev.memMonitor.Wait() {
				// done
				return
			}
		}

		ev.latestMapLock.RLock()
		version, tree, ok := ev.latest.Min()
		ev.latestMapLock.RUnlock()
		evictVersion := ev.savedVersion.Load()
		if !ok || version > evictVersion {
			ev.needReader.Store(true)
			select {
			case <-ev.wakeCh:
				continue
			case <-ev.ctx.Done():
				return
			}
		}

		evictTraverse(tree, ev.memMonitor, evictVersion)

		if tree.mem.Load() == nil {
			// if we fully evicted this version, remove it from latest map
			ev.latestMapLock.Lock()
			ev.latest.Delete(version)
			ev.latestMapLock.Unlock()
		}
	}
}

func evictTraverse(np *NodePointer, tracker *memoryMonitor, evictVersion uint32) bool {
	if !tracker.UnderPressure() {
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
		if !tracker.TrackEviction(memNode) {
			return false
		}
	}

	return true
}
