package iavlx

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"go.opentelemetry.io/otel"
)

var (
	tracer = otel.Tracer("iavlx")
)

type cleanupProc struct {
	*TreeStore
	closeCleanupProc chan struct{}
	cleanupProcDone  chan struct{}

	// Split orphan queues based on whether versions are readable
	orphanWriteQueue  []markOrphansReq // For versions <= savedVersion (can process immediately)
	stagedOrphanQueue []markOrphansReq // For versions > savedVersion (need to wait)
	orphanQueueLock   sync.Mutex

	toDelete        map[*Changeset]ChangesetDeleteArgs
	activeCompactor *Compactor
	beingCompacted  []compactionEntry

	// Disposal queue for evicted changesets awaiting refcount=0
	disposalQueue sync.Map // *Changeset -> struct{}
}

type compactionEntry struct {
	entry *changesetEntry
	cs    *Changeset
}

func newCleanupProc(treeStore *TreeStore) *cleanupProc {
	cp := &cleanupProc{
		TreeStore:        treeStore,
		closeCleanupProc: make(chan struct{}),
		cleanupProcDone:  make(chan struct{}),
		toDelete:         make(map[*Changeset]ChangesetDeleteArgs),
	}
	go cp.run()
	return cp
}

func (cp *cleanupProc) run() {
	ctx, span := tracer.Start(context.Background(), "cleanupProc")
	defer span.End()
	// before we shutdown save any pending orphans
	defer func() {
		err := cp.doMarkOrphans()
		if err != nil {
			cp.logger.Error("failed to mark orphans at shutdown", "error", err)
		}
	}()
	defer close(cp.cleanupProcDone)

	minCompactorInterval := time.Second * time.Duration(cp.opts.MinCompactionSeconds)
	var lastCompactorStart time.Time

	for {
		sleepTime := time.Duration(0)
		if time.Since(lastCompactorStart) < minCompactorInterval {
			sleepTime = minCompactorInterval - time.Since(lastCompactorStart)
		}
		select {
		case <-cp.closeCleanupProc:
			return
		case <-time.After(sleepTime):
		}

		lastCompactorStart = time.Now()

		// process any pending orphans at the start of each cycle
		err := cp.doMarkOrphans()
		if err != nil {
			cp.logger.Error("failed to mark orphans at start of cycle", "error", err)
		}

		// collect current entries
		cp.changesetsMapLock.RLock()
		var entries []*changesetEntry
		cp.changesets.Scan(func(version uint32, entry *changesetEntry) bool {
			entries = append(entries, entry)
			return true
		})
		cp.changesetsMapLock.RUnlock()

		for i := 0; i < len(entries); i++ {
			entry := entries[i]
			var nextEntry *changesetEntry
			if i+1 < len(entries) {
				nextEntry = entries[i+1]
			}
			err := cp.processEntry(ctx, entry, nextEntry)
			if err != nil {
				cp.logger.Error("failed to process changeset entry", "error", err)
				// on error, clean up any failed compaction and stop processing further entries this round
				cp.cleanupFailedCompaction()
				break
			}
		}
		if cp.activeCompactor != nil {
			err := cp.sealActiveCompactor()
			if err != nil {
				cp.logger.Error("failed to seal active compactor", "error", err)
			}
		}

		cp.processToDelete()
		cp.processDisposalQueue()
	}
}

func (cp *cleanupProc) markOrphans(version uint32, nodeIds [][]NodeID) {
	req := markOrphansReq{
		version: version,
		orphans: nodeIds,
	}

	cp.orphanQueueLock.Lock()
	defer cp.orphanQueueLock.Unlock()

	cp.orphanWriteQueue = append(cp.orphanWriteQueue, req)
}

// doMarkOrphans must only be called from the cleanupProc
func (cp *cleanupProc) doMarkOrphans() error {
	var orphanQueue []markOrphansReq
	cp.orphanQueueLock.Lock()
	orphanQueue, cp.orphanWriteQueue = cp.orphanWriteQueue, nil
	cp.orphanQueueLock.Unlock()

	orphanQueue = append(orphanQueue, cp.stagedOrphanQueue...)

	savedVersion := cp.savedVersion.Load()
	var newStagedOrphans []markOrphansReq

	for _, req := range orphanQueue {
		for _, nodeSet := range req.orphans {
			var stagedNodes []NodeID
			for _, nodeId := range nodeSet {
				nodeVersion := uint32(nodeId.Version())

				// Route to staged queue if version not yet readable
				if nodeVersion > savedVersion {
					stagedNodes = append(stagedNodes, nodeId)
					continue
				}

				ce := cp.getChangesetEntryForVersion(nodeVersion)
				if ce == nil {
					return fmt.Errorf("no changeset found for version %d", nodeVersion)
				}
				// this somewhat awkward retry loop is needed to handle a race condition where
				// we have disposed of a changeset between getting the entry and marking the orphan
				retries := 0
				for {
					err := ce.changeset.Load().MarkOrphan(req.version, nodeId)
					if errors.Is(err, ErrDisposed) {
						if retries > 3 {
							return fmt.Errorf("changeset for version %d disposed while marking orphan %s", nodeVersion, nodeId.String())
						}
						retries++
						continue
					} else if err != nil {
						return err
					}
					break
				}
			}
			// Add any staged nodes back to the staged queue
			if len(stagedNodes) > 0 {
				newStagedOrphans = append(newStagedOrphans, markOrphansReq{
					version: req.version,
					orphans: [][]NodeID{stagedNodes},
				})
			}
		}
	}

	cp.stagedOrphanQueue = newStagedOrphans

	return nil
}

func (cp *cleanupProc) processEntry(ctx context.Context, entry, nextEntry *changesetEntry) error {
	ctx, span := tracer.Start(ctx, "cleanupProc.processEntry")
	defer span.End()

	cs := entry.changeset.Load()

	if cs.files == nil {
		// skipping incomplete changeset which is still open for writing
		return nil
	}

	// safety check - skip if evicted or disposed
	if cs.evicted.Load() || cs.disposed.Load() {
		return fmt.Errorf("evicted/disposed changeset: %s found in queue", cs.files.dir)
	}

	if cp.opts.DisableCompaction {
		return nil
	}

	// skip if still pending sync
	if cs.needsSync.Load() {
		return nil
	}

	if cp.activeCompactor != nil {
		if cp.opts.CompactWAL &&
			cs.TotalBytes()+cp.activeCompactor.TotalBytes() <= int(cp.opts.GetCompactionMaxTarget()) {
			// add to active compactor
			slog.DebugContext(ctx, "joining changeset to active compactor", "info", cs.info, "size", cs.TotalBytes(), "dir", cs.files.dir,
				"newDir", cp.activeCompactor.files.dir)
			err := cp.activeCompactor.AddChangeset(cs)
			if err != nil {
				return fmt.Errorf("failed to add changeset to active compactor: %w", err)
			}
			cp.beingCompacted = append(cp.beingCompacted, compactionEntry{entry: entry, cs: cs})
			return nil
		} else {
			err := cp.sealActiveCompactor()
			if err != nil {
				cp.cleanupFailedCompaction()
				return fmt.Errorf("failed to seal active compactor: %w", err)
			}
		}
	}

	// mark any pending orphans here when we don't have an active compactor
	err := cp.doMarkOrphans()
	if err != nil {
		cp.logger.Error("failed to mark orphans", "error", err)
	}

	// check if other triggers apply for a new compaction
	savedVersion := cp.savedVersion.Load()
	retainVersions := cp.opts.RetainVersions
	retentionWindowBottom := savedVersion - retainVersions
	if retainVersions == 0 {
		// retain everything
		retentionWindowBottom = 0
	}

	compactOrphanAge := cp.opts.GetCompactionOrphanAge()
	compactOrphanThreshold := cp.opts.GetCompactionOrphanRatio()

	// Age target relative to bottom of retention window
	ageTarget := retentionWindowBottom - compactOrphanAge

	// Check orphan-based trigger
	shouldCompact := cs.ReadyToCompact(compactOrphanThreshold, ageTarget)
	if !shouldCompact {
		lastCompactedAt := cs.files.CompactedAtVersion()
		if savedVersion-lastCompactedAt >= cp.opts.GetCompactAfterVersions() {
			shouldCompact = cs.HasOrphans()
		}
	}

	// Check size-based joining trigger
	maxSize := cp.opts.GetCompactionMaxTarget()

	canJoin := false
	if !shouldCompact && cp.opts.CompactWAL && nextEntry != nil {
		nextCs := nextEntry.changeset.Load()
		if nextCs.files != nil && // we can't compact a changeset that's still being written
			nextCs.info.StartVersion == cs.info.EndVersion+1 {
			if uint64(cs.TotalBytes())+uint64(nextCs.TotalBytes()) <= maxSize {
				canJoin = true
			}
		}
	}

	if !shouldCompact && !canJoin {
		return nil
	}

	retainVersion := retentionWindowBottom
	retainCriteria := func(createVersion, orphanVersion uint32) bool {
		// orphanVersion should be non-zero
		if orphanVersion >= retainVersion {
			// keep the orphan if it's in the retain window
			return true
		} else {
			// otherwise, we can remove it
			return false
		}
	}

	slog.Info("compacting changeset", "info", cs.info, "size", cs.TotalBytes(), "dir", cs.files.dir)

	cp.activeCompactor, err = NewCompacter(ctx, cs, CompactOptions{
		RetainCriteria: retainCriteria,
		CompactWAL:     cp.opts.CompactWAL,
		CompactedAt:    savedVersion,
	}, cp.TreeStore)
	if err != nil {
		return fmt.Errorf("failed to create compactor: %w", err)
	}
	cp.beingCompacted = []compactionEntry{{entry: entry, cs: cs}}
	return nil
}

func (cp *cleanupProc) sealActiveCompactor() error {
	// seal compactor and finish
	newCs, err := cp.activeCompactor.Seal()
	if err != nil {
		return fmt.Errorf("failed to seal active compactor: %w", err)
	}

	// update all processed entries to point to new changeset
	oldSize := uint64(0)
	for i, procEntry := range cp.beingCompacted {
		cp.logger.Debug("updating changeset entry to compacted changeset and trying to delete",
			"old_dir", procEntry.cs.files.dir, "new_dir", newCs.files.dir)

		oldCs := procEntry.cs
		oldDir := oldCs.files.dir
		oldSize += uint64(oldCs.TotalBytes())

		if i == 0 {
			procEntry.entry.changeset.Store(newCs)
		} else {
			cp.changesetsMapLock.Lock()
			cp.changesets.Delete(oldCs.files.StartVersion())
			cp.changesetsMapLock.Unlock()
		}
		oldCs.Evict()

		// try to delete now or schedule for later
		if !oldCs.TryDispose() {
			cp.logger.Debug("changeset has active references, scheduling for deletion", "path", oldDir, "refcount", oldCs.refCount.Load())
			cp.toDelete[oldCs] = ChangesetDeleteArgs{newCs.files.KVLogPath()}
		} else {
			cp.logger.Info("changeset disposed, deleting files", "path", oldDir)
			err = oldCs.files.DeleteFiles(ChangesetDeleteArgs{SaveKVLogPath: newCs.files.KVLogPath()})
			if err != nil {
				cp.logger.Error("failed to delete old changeset files", "error", err, "path", oldDir)
			}
		}
	}

	cp.logger.Info("compacted changeset", "dir", newCs.files.dir, "new_size", newCs.TotalBytes(), "old_size", oldSize, "joined", len(cp.beingCompacted))

	// Clear compactor state after successful seal
	cp.activeCompactor = nil
	cp.beingCompacted = nil
	return nil
}

func (cp *cleanupProc) cleanupFailedCompaction() {
	// clean up any partial compactor state and remove temporary files
	if cp.activeCompactor != nil && cp.activeCompactor.files != nil {
		cp.logger.Warn("cleaning up failed compaction", "dir", cp.activeCompactor.files.dir, "changesets_attempted", len(cp.beingCompacted))
		err := cp.activeCompactor.Abort()
		if err != nil {
			cp.logger.Error("failed to abort active compactor", "error", err)
		}
	}
	cp.activeCompactor = nil
	cp.beingCompacted = nil
}

func (cp *cleanupProc) processToDelete() {
	if len(cp.toDelete) > 0 {
		cp.logger.Debug("processing delete queue", "size", len(cp.toDelete))
	}

	for oldCs, args := range cp.toDelete {
		select {
		case <-cp.closeCleanupProc:
			return
		default:
		}

		if !oldCs.TryDispose() {
			cp.logger.Warn("old changeset not disposed, skipping delete", "path", oldCs.files.dir, "refcount", oldCs.refCount.Load())
			continue
		}

		cp.logger.Info("deleting old changeset files", "path", oldCs.files.dir)
		err := oldCs.files.DeleteFiles(args)
		if err != nil {
			cp.logger.Error("failed to delete old changeset files", "error", err)
		}
		delete(cp.toDelete, oldCs)
	}
}

func (cp *cleanupProc) shutdown() {
	close(cp.closeCleanupProc)
	<-cp.cleanupProcDone
}

// addPendingDisposal adds an evicted changeset to the disposal queue
func (cp *cleanupProc) addPendingDisposal(cs *Changeset) {
	cp.disposalQueue.Store(cs, struct{}{})
}

// processDisposalQueue tries to dispose changesets waiting for refcount=0
func (cp *cleanupProc) processDisposalQueue() {
	disposalCount := 0
	cp.disposalQueue.Range(func(key, value interface{}) bool {
		disposalCount++
		return true
	})

	if disposalCount > 0 {
		cp.logger.Debug("processing disposal queue", "size", disposalCount)
	}

	cp.disposalQueue.Range(func(key, value interface{}) bool {
		cs := key.(*Changeset)
		if cs.TryDispose() {
			cp.disposalQueue.Delete(cs)
			cp.logger.Debug("disposed shared changeset from queue")
		} else {
			cp.logger.Debug("shared changeset still has references", "refcount", cs.refCount.Load())
		}
		return true
	})

	// Warn if the disposal queue is getting large
	if disposalCount > 100 {
		cp.logger.Warn("disposal queue is large", "size", disposalCount)
	}
}
