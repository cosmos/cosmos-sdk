package internal

import (
	"context"
	"errors"
	"fmt"
)

// Compaction merges one or more sealed changesets into a single new changeset, pruning orphaned
// nodes along the way. This serves two purposes:
//
//  1. Disk reclamation: nodes that were replaced or deleted in newer versions (orphans) can be
//     removed from the data files, freeing disk space.
//  2. File consolidation: multiple small changeset directories are merged into fewer, larger ones,
//     which reduces the number of open file handles and simplifies the version-to-changeset lookup.
//
// The compaction process for each changeset being merged:
//   - Rewrite the WAL: copy WAL entries starting from WALStartVersion into the new WAL file,
//     recording how key/value offsets shifted so we can remap node references.
//   - Process orphans: read the orphan log and use RetainCriteria to decide which orphaned nodes
//     to keep (still needed for historical queries) vs delete (prunable). Build a deleteMap.
//   - Copy checkpoint data: iterate each checkpoint's leaves and branches, skipping nodes in the
//     deleteMap. For surviving nodes, remap their key/value offsets (which may now point into the
//     new WAL or new kv.dat), and remap branch child offsets to reflect new file positions.
//   - Track new file positions in offsetCache so branch nodes can find their children.
//
// After all changesets are processed, Seal() finalizes the new changeset, atomically swaps it
// in for the originals, and marks the old changesets for deletion.
//
// Compaction runs in the background (see compactIfNeeded in commit_multi_tree.go) and does not
// block commits. The orphan processor lock is held only briefly during orphan preprocessing and
// during the final switchover, to prevent orphan writes from going to a changeset that's about
// to be replaced.

type CompactorOptions struct {
	// RetainCriteria decides whether an orphaned node should be kept.
	// It receives the checkpoint the node was created in and the version it was orphaned at.
	// Returns true to retain (keep the node in the compacted output), false to prune it.
	RetainCriteria RetainCriteria
	// CompactedAt is stamped into the new changeset directory name so we can tell when it was compacted.
	CompactedAt uint32
	// WALStartVersion is the first version whose WAL entries are copied to the new WAL.
	// Earlier WAL entries are dropped because their state is already captured in checkpoints.
	WALStartVersion uint32
}

// RetainCriteria decides whether an orphaned node should be kept during compaction.
// createCheckpoint is the checkpoint the node was originally written in.
// orphanVersion is the version at which the node became orphaned (was replaced or deleted).
type RetainCriteria func(createCheckpoint, orphanVersion uint32) bool

type Compactor struct {
	shouldRetain    RetainCriteria
	walStartVersion uint32

	processedChangesets []pendingCompactionEntry
	treeStore           *TreeStore

	// Output files: the new compacted changeset being built.
	files          *ChangesetFiles
	leavesWriter   *StructWriter[LeafLayout]
	branchesWriter *StructWriter[BranchLayout]
	cpInfoWriter   *StructWriter[CheckpointInfo]
	walWriter      *WALWriter
	kvlogWriter    *KVDataWriter
	orphanWriter   *StructWriter[OrphanEntry]

	endVersion uint32

	// offsetCache maps NodeID → new 1-based file offset in the compacted output.
	// When a node is written to the new leaves or branches file, its new offset is recorded here.
	// Branch nodes use this to update their left/right child offsets, since the children may have
	// moved to different file positions during compaction.
	offsetCache map[NodeID]uint32

	ctx context.Context
}

type pendingCompactionEntry struct {
	orig                *Changeset
	orphanRewriter      *OrphanRewriter
	origStartCheckpoint uint32
}

func NewCompactor(ctx context.Context, reader *ChangesetReader, opts CompactorOptions, store *TreeStore) (*Compactor, error) {
	files := reader.changeset.files
	startingVersion := files.StartVersion()
	lastCompactedAt := files.CompactedAtVersion()
	if lastCompactedAt >= opts.CompactedAt {
		return nil, fmt.Errorf("cannot compact changeset starting at version %d which was last compacted at %d to an earlier or same version %d",
			startingVersion, lastCompactedAt, opts.CompactedAt)
	}

	newFiles, err := CreateChangesetFiles(files.TreeDir(), files.StartVersion(), opts.CompactedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to open new changeset files: %w", err)
	}

	c := &Compactor{
		ctx:             ctx,
		shouldRetain:    opts.RetainCriteria,
		walStartVersion: opts.WALStartVersion,
		treeStore:       store,
		files:           newFiles,
		walWriter:       NewWALWriter(newFiles.WALFile()),
		kvlogWriter:     NewKVDataWriter(newFiles.KVDataFile()),
		leavesWriter:    NewStructWriter[LeafLayout](newFiles.leavesFile),
		branchesWriter:  NewStructWriter[BranchLayout](newFiles.branchesFile),
		cpInfoWriter:    NewStructWriter[CheckpointInfo](newFiles.checkpointsFile),
		offsetCache:     make(map[NodeID]uint32),
		orphanWriter:    NewStructWriter[OrphanEntry](newFiles.orphansFile),
	}

	err = c.AddChangeset(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to add initial changeset to compactor: %w", err)
	}

	return c, nil
}

func (c *Compactor) AddChangeset(reader *ChangesetReader) error {
	err := c.doAddChangeset(reader)
	if err != nil {
		abortErr := c.Abort()
		if abortErr != nil {
			return fmt.Errorf("failed to add changeset to compactor: %w; additionally failed to abort compactor during cleanup: %w", err, abortErr)
		}
		return fmt.Errorf("failed to add changeset to compactor: %w", err)
	}
	return nil
}

// doAddChangeset processes a single source changeset and writes its surviving data to the compacted output.
// This is the core of the compaction algorithm:
//
// Step 1: Rewrite the WAL — copy entries from walStartVersion onward into the new WAL file.
//
//	The rewrite returns offset remapping tables so we can update node key/value references.
//
// Step 2: Process orphans — read the source changeset's orphan log to determine which nodes
//
//	have been replaced/deleted. The RetainCriteria function decides: nodes orphaned before
//	the retain version are prunable (added to deleteMap), nodes orphaned after are kept
//	(copied to the new orphan log for future compaction cycles).
//
// Step 3: Copy checkpoint data — for each checkpoint in the source changeset, iterate its leaves
//
//	and branches. Skip any node in the deleteMap. For surviving nodes, remap their key/value
//	offsets (from old WAL/kv.dat positions to new ones) and remap branch child offsets
//	(from old file positions to new ones via offsetCache).
func (c *Compactor) doAddChangeset(reader *ChangesetReader) error {
	// Step 1: Rewrite WAL entries starting from walStartVersion.
	// RewriteWAL returns remapping tables that tell us where each key/value ended up in the new WAL.
	walRewriteInfo, err := RewriteWAL(c.walWriter, reader.changeset.files.WALFile(), uint64(c.walStartVersion))
	if err != nil {
		return fmt.Errorf("failed to rewrite WAL during compaction: %w", err)
	}

	c.endVersion = reader.Changeset().Files().EndVersion() // this will only be non-zero in already compacted changesets
	if c.endVersion == 0 {
		// in original changesets, we get the end version from WAL rewrite
		c.endVersion = uint32(walRewriteInfo.EndVersion)
	}

	cpInfo := reader.checkpointsInfo
	numCheckpoints := cpInfo.Count()
	leavesData := reader.leavesData
	branchesData := reader.branchesData

	// Step 2: Process orphans to build a deleteMap of nodes to prune.
	// We hold the orphan proc lock briefly here to prevent the live orphan processor from
	// writing to the source changeset's orphan file while we're reading it.
	// TODO the deleteMap from the orphan rewriter contains the exact number of nodes that will be pruned
	// so we could use this to implement threshold based pruning and only prune based on whether we
	// are pruning enough orphans
	// in that case this orphan pre-processing logic should happen before anything else and
	// be a condition on whether we add this changeset or not
	orphanRewriter, err := NewOrphanRewriter(reader.changeset.orphanWriter)
	if err != nil {
		return fmt.Errorf("failed to create orphan rewriter: %w", err)
	}
	c.treeStore.LockOrphanProc()
	deleteMap, err := orphanRewriter.Preprocess(c.shouldRetain, c.orphanWriter)
	if err != nil {
		c.treeStore.UnlockOrphanProc()
		return fmt.Errorf("failed to preprocess orphans for compaction: %w", err)
	}
	c.treeStore.UnlockOrphanProc()

	// Step 3: Copy checkpoint data, skipping pruned nodes and remapping offsets.
	// We iterate checkpoints in order. For each checkpoint, we iterate its leaves first,
	// then branches (branches reference leaves, so leaves must be written first to populate offsetCache).
	c.treeStore.logger.DebugContext(c.ctx, "processing changeset for compaction", "numCheckpoints", numCheckpoints)
	for i := 0; i < numCheckpoints; i++ {
		cpInfo := cpInfo.UnsafeItem(i) // copy
		newLeafStartIdx := uint32(0)
		newLeafEndIdx := uint32(0)
		leafStartOffset := cpInfo.Leaves.StartOffset
		leafCount := cpInfo.Leaves.Count
		newLeafStartOffset := uint32(c.leavesWriter.Count())
		newLeafCount := uint32(0)
		// Iterate leaves
		// For each leaf, check if it should be retained
		for j := uint32(0); j < leafCount; j++ {
			existingLeaf := LeafPersisted{store: reader, layout: leavesData.UnsafeItem(int(leafStartOffset + j))}
			leaf := *existingLeaf.layout // copy so we don't modify original
			id := leaf.ID
			_, toDelete := deleteMap[id]
			if toDelete {
				continue
			}

			if newLeafStartIdx == 0 {
				newLeafStartIdx = id.Index()
			}
			newLeafEndIdx = id.Index()
			newLeafCount++

			// remap key offset
			keyOffset, keyInKVData, err := c.remapBlob(leaf.KeyInKVData(), leaf.KeyOffset.ToUint64(), walRewriteInfo, true, &existingLeaf)
			if err != nil {
				return fmt.Errorf("failed to remap key blob for leaf %s: %w", id, err)
			}
			leaf.KeyOffset = NewUint40(keyOffset)
			leaf.SetKeyInKVData(keyInKVData)

			//  remap value offset
			valOffset, valInKVData, err := c.remapBlob(leaf.ValueInKVData(), leaf.ValueOffset.ToUint64(), walRewriteInfo, false, &existingLeaf)
			if err != nil {
				return fmt.Errorf("failed to remap value blob for leaf %s: %w", id, err)
			}
			leaf.ValueOffset = NewUint40(valOffset)
			leaf.SetValueInKVData(valInKVData)

			err = c.leavesWriter.Append(&leaf)
			if err != nil {
				return fmt.Errorf("failed to append leaf %s: %w", id, err)
			}

			c.offsetCache[id] = uint32(c.leavesWriter.Count())
		}

		newBranchStartIdx := uint32(0)
		newBranchEndIdx := uint32(0)
		branchStartOffset := cpInfo.Branches.StartOffset
		branchCount := cpInfo.Branches.Count
		newBranchStartOffset := uint32(c.branchesWriter.Count())
		newBranchCount := uint32(0)
		for j := uint32(0); j < branchCount; j++ {
			existingBranch := BranchPersisted{store: reader, layout: branchesData.UnsafeItem(int(branchStartOffset + j))}
			branch := *existingBranch.layout // copy so we don't modify original
			id := branch.ID
			_, toDelete := deleteMap[id]
			if toDelete {
				continue
			}

			if newBranchStartIdx == 0 {
				newBranchStartIdx = id.Index()
			}
			newBranchEndIdx = id.Index()
			newBranchCount++

			if newLeftOffset, ok := c.offsetCache[branch.Left]; ok {
				branch.LeftOffset = newLeftOffset
			} else {
				branch.LeftOffset = 0
			}
			if newRightOffset, ok := c.offsetCache[branch.Right]; ok {
				branch.RightOffset = newRightOffset
			} else {
				branch.RightOffset = 0
			}

			// remap key offset
			keyOffset, keyInKVData, err := c.remapBlob(branch.KeyInKVData(), branch.KeyOffset.ToUint64(), walRewriteInfo, true, &existingBranch)
			if err != nil {
				return fmt.Errorf("failed to remap key blob for branch %s: %w", id, err)
			}
			branch.KeyOffset = NewUint40(keyOffset)
			branch.SetKeyInKVData(keyInKVData)

			err = c.branchesWriter.Append(&branch)
			if err != nil {
				return fmt.Errorf("failed to append branch %s: %w", id, err)
			}
			c.offsetCache[id] = uint32(c.branchesWriter.Count())
		}

		checkpointHasNoNodes := newBranchCount == 0 && newLeafCount == 0
		if checkpointHasNoNodes {
			// we attempt to delete checkpoint infos that have no nodes
			inRetainRange := cpInfo.Version+1 >= c.walStartVersion
			// we need to keep checkpoints in the retain range even if they're empty
			// either they represent an empty tree or they point to some earlier
			// tree root which was retained
			// we need to preserve this for WAL replay from this checkpoint
			if !inRetainRange {
				continue
			}
		}

		cpInfo = &CheckpointInfo{
			Leaves: NodeSetInfo{
				StartIndex:  newLeafStartIdx,
				EndIndex:    newLeafEndIdx,
				StartOffset: newLeafStartOffset,
				Count:       newLeafCount,
			},
			Branches: NodeSetInfo{
				StartIndex:  newBranchStartIdx,
				EndIndex:    newBranchEndIdx,
				StartOffset: newBranchStartOffset,
				Count:       newBranchCount,
			},
			Checkpoint:  cpInfo.Checkpoint,
			Version:     cpInfo.Version,
			RootID:      cpInfo.RootID,
			KVEndOffset: uint64(c.kvlogWriter.Size()),
		}
		cpInfo.SetCRC32()

		err := c.cpInfoWriter.Append(cpInfo)
		if err != nil {
			return fmt.Errorf("failed to append checkpoint info for checkpoint %d: %w", cpInfo.Version, err)
		}
	}

	// track this changeset as processed
	c.processedChangesets = append(c.processedChangesets, pendingCompactionEntry{
		orig:                reader.Changeset(),
		orphanRewriter:      orphanRewriter,
		origStartCheckpoint: reader.FirstCheckpoint(),
	})

	return nil
}

// remapBlob translates a key or value reference from the old changeset to the new compacted one.
//
// A node's key/value can live in one of two places: the WAL file or the kv.dat file.
// During compaction, both files are being rewritten, so offsets change. The logic:
//   - If the blob was in the OLD WAL: look up its new offset in the WAL rewrite remapping table.
//     If found, the blob is still in the new WAL. If not found (the WAL entry was before
//     walStartVersion and was dropped), we need to read the actual bytes and write them to kv.dat.
//   - If the blob was in kv.dat: read the actual bytes and write them to the NEW kv.dat
//     (since the old kv.dat file won't exist after compaction).
func (c *Compactor) remapBlob(origIsInKVData bool, origOffset uint64, walRewriteInfo *WALRewriteInfo, isKey bool, node Node) (newOffset uint64, newInKVData bool, err error) {
	newInKVData = origIsInKVData
	if !origIsInKVData {
		// try to find it in the wal
		var isInWAL bool
		if isKey {
			newOffset, isInWAL = walRewriteInfo.KeyOffsetRemapping[origOffset]
		} else {
			newOffset, isInWAL = walRewriteInfo.ValueOffsetRemapping[origOffset]
		}
		newInKVData = !isInWAL
	}
	if newInKVData {
		// if it's not in the wal, we need to read the original value and write it to the new kv log
		var blob UnsafeBytes
		var err error
		if isKey {
			blob, err = node.Key()
		} else {
			blob, err = node.Value()
		}
		if err != nil {
			return 0, false, fmt.Errorf("failed to read original blob during remap: %w", err)
		}
		if isKey {
			newOffset, err = c.kvlogWriter.WriteKeyBlob(blob)
		} else {
			newOffset, err = c.kvlogWriter.WriteValueBlob(blob)
		}
		if err != nil {
			return 0, false, fmt.Errorf("failed to write blob to new kv log during remap: %w", err)
		}
	}
	return newOffset, newInKVData, nil
}

func (c *Compactor) Seal() (*Changeset, error) {
	if len(c.processedChangesets) == 0 {
		return nil, fmt.Errorf("no changesets processed")
	}

	errs := []error{
		c.leavesWriter.Sync(),
		c.branchesWriter.Sync(),
		c.cpInfoWriter.Sync(),
		c.kvlogWriter.Sync(),
		c.walWriter.Sync(),
	}
	if err := errors.Join(errs...); err != nil {
		// if this fails we abort
		errAbort := c.Abort()
		if errAbort != nil {
			return nil, fmt.Errorf("failed to flush data during compaction seal: %w; additionally failed to abort compactor during cleanup: %w", err, errAbort)
		}
		return nil, fmt.Errorf("failed to flush data during compaction seal: %w, aborted compaction", err)
	}

	cs, err := c.switchoverChangesets()
	if err != nil {
		return nil, fmt.Errorf("failed to clean up orphans during compaction seal: %w", err)
	}

	c.updateChangesetVersionEntries(cs)
	c.updateChangesetCheckpointEntries(cs)

	return cs, nil
}

// switchoverChangesets atomically replaces the old changesets with the new compacted one.
// This is the critical section of compaction — it holds the orphan proc lock to ensure no
// orphan writes land in the old changesets between when we finalize and when we mark them
// as compacted.
//
// The sequence is:
//  1. finalize: finish writing orphan data, rename the -tmp directory to its final name,
//     open it as a proper Changeset.
//  2. markCompacted: point each old changeset to the new compacted one, so any future readers
//     get redirected. After this point the old changesets are queued for deletion.
//
// IMPORTANT: after markCompacted, we cannot abort — the old changesets are already marked
// for deletion. If we deleted the compacted output too, we'd lose data.
func (c *Compactor) switchoverChangesets() (*Changeset, error) {
	c.treeStore.LockOrphanProc()
	defer c.treeStore.UnlockOrphanProc()

	cs, err := c.finalize()
	if err != nil {
		// if we error at this point, we abort and cleanup, past this point there can be no aborting
		errAbort := c.Abort()
		if errAbort != nil {
			return nil, fmt.Errorf("failed to finalize changeset during compaction seal: %w; additionally failed to abort compactor during cleanup: %w", err, errAbort)
		}
		return nil, fmt.Errorf("failed to finalize changeset during compaction seal: %w, but aborted successfully", err)
	}

	// IMPORTANT: an abort CANNOT happen past this point, otherwise we can lose both the original and compacted changesets and cause data loss!
	// this operation does not error, but critically it marks the original changesets for deletion!
	// this operation MUST happen while we are holding the orphan proc lock to prevent orphans from going to the old changesets now that we've switched over
	for _, entry := range c.processedChangesets {
		entry.orig.markCompacted(cs)
	}

	return cs, nil
}

func (c *Compactor) finalize() (*Changeset, error) {
	for _, entry := range c.processedChangesets {
		err := entry.orphanRewriter.FinishRewrite(c.orphanWriter)
		if err != nil {
			return nil, fmt.Errorf("failed to finish orphan rewrite for a processed changeset during compaction seal: %w", err)
		}
	}
	err := c.orphanWriter.Sync()
	if err != nil {
		return nil, fmt.Errorf("failed to flush orphan data during compaction seal: %w", err)
	}

	finalDir, err := c.files.MarkReadyAndClose(c.endVersion)
	if err != nil {
		return nil, fmt.Errorf("failed to mark changeset as ready during compaction seal: %w", err)
	}

	cs, err := OpenChangeset(c.treeStore, finalDir, false)
	if err != nil {
		return nil, fmt.Errorf("failed to create new changeset for compacted data during compaction seal: %w", err)
	}

	return cs, nil
}

func (c *Compactor) updateChangesetVersionEntries(newCs *Changeset) {
	c.treeStore.changesetsLock.Lock()
	defer c.treeStore.changesetsLock.Unlock()
	// clear the old changesets that were compacted
	for _, cs := range c.processedChangesets {
		c.treeStore.changesetsByVersion.Delete(cs.orig.Files().StartVersion())
	}
	// set the new compacted changeset
	c.treeStore.changesetsByVersion.Set(newCs.Files().StartVersion(), newCs)
}

func (c *Compactor) updateChangesetCheckpointEntries(newCs *Changeset) {
	newRdr, pin := newCs.TryPinReader()
	defer pin.Unpin()
	checkpointer := c.treeStore.checkpointer
	checkpointer.changesetLock.Lock()
	defer checkpointer.changesetLock.Unlock()

	// clear the old changesets that were compacted
	for _, cs := range c.processedChangesets {
		firstCheckpoint := cs.origStartCheckpoint
		if firstCheckpoint != 0 {
			checkpointer.changesetsByCheckpoint.Delete(firstCheckpoint)
		}
	}
	checkpointer.changesetsByCheckpoint.Set(newRdr.FirstCheckpoint(), newCs)
}

func (c *Compactor) Abort() error {
	err := c.files.Close()
	if err != nil {
		return fmt.Errorf("failed to close compactor files during cleanup: %w", err)
	}
	return c.files.DeleteFiles()
}

func (c *Compactor) TotalBytes() int {
	total := c.leavesWriter.Size() + c.branchesWriter.Size() + c.cpInfoWriter.Size()
	total += c.kvlogWriter.Size()
	total += c.walWriter.Size()
	return total
}
