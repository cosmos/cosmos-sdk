package iavlx

type evictor struct {
	*TreeStore
	wakeCh           chan struct{}
	lastEvictVersion uint32
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
		// check for context cancellation
		select {
		case <-ev.ctx.Done():
			return
		default:
		}

		ev.latestMapLock.RLock()
		version, _, ok := ev.latest.Min()
		ev.latestMapLock.RUnlock()
		savedVersion := ev.savedVersion.Load()
		latestVersion := ev.latestVersion.Load()
		if ok && version <= savedVersion && version < latestVersion {
			// delete the saved version that is not the latest
			ev.latestMapLock.Lock()
			ev.latest.Delete(version)
			ev.latestMapLock.Unlock()
		} else if latestVersion > ev.lastEvictVersion {
			// do an evict traverse on the latest tree if it is newer than last version that was evict traversed
			tree, _ := ev.latest.Get(latestVersion)
			evictCount := evictTraverse(tree, 0, ev.opts.EvictDepth, savedVersion)
			ev.logger.DebugContext(ev.ctx, "evicted nodes", "version", version, "count", evictCount, "evictVersion", savedVersion, "lastEvictVersion", ev.lastEvictVersion)
			ev.lastEvictVersion = version
		} else {
			// wait for wake signal or context cancellation
			select {
			case <-ev.wakeCh:
				continue
			case <-ev.ctx.Done():
				return
			}
		}
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
