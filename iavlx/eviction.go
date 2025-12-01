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
			ev.logger.DebugContext(ev.ctx, "evictor waiting", "budget", ev.memMonitor.evictBudget.Load())
			if !ev.memMonitor.Wait() {
				ev.logger.DebugContext(ev.ctx, "evictor exiting (context cancelled)")
				return
			}
			ev.logger.DebugContext(ev.ctx, "evictor woke from Wait", "budget", ev.memMonitor.evictBudget.Load())
		}

		ev.latestMapLock.RLock()
		version, tree, ok := ev.latest.Min()
		ev.latestMapLock.RUnlock()
		evictVersion := ev.savedVersion.Load()
		ev.logger.DebugContext(ev.ctx, "evictor checking", "latestMinVersion", version, "latestOk", ok, "evictVersion", evictVersion, "budget", ev.memMonitor.evictBudget.Load())
		if !ok || version > evictVersion {
			ev.logger.DebugContext(ev.ctx, "evictor needs reader", "version", version, "evictVersion", evictVersion, "ok", ok)
			ev.needReader.Store(true)
			select {
			case <-ev.wakeCh:
				ev.logger.DebugContext(ev.ctx, "evictor woke from wakeCh")
				continue
			case <-ev.ctx.Done():
				ev.logger.DebugContext(ev.ctx, "evictor exiting (context cancelled while waiting for reader)")
				return
			}
		}

		budgetBefore := ev.memMonitor.evictBudget.Load()
		evictTraverse(tree, ev.memMonitor, evictVersion)
		budgetAfter := ev.memMonitor.evictBudget.Load()
		ev.logger.DebugContext(ev.ctx, "evictTraverse completed", "version", version, "evictVersion", evictVersion, "budgetBefore", budgetBefore, "budgetAfter", budgetAfter, "evicted", budgetBefore-budgetAfter)

		if tree.mem.Load() == nil {
			ev.logger.DebugContext(ev.ctx, "removing fully evicted version from latest", "version", version)
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
