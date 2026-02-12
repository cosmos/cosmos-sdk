package internal

import (
	"context"
	"errors"
	"fmt"
)

type CompactOptions struct {
	RetainCriteria  RetainCriteria
	CompactedAt     uint32 // version at which compaction is done
	WALStartVersion uint32
}

type RetainCriteria func(createCheckpoint, orphanVersion uint32) bool

type Compactor struct {
	criteria        RetainCriteria
	walStartVersion uint32

	processedChangesets []pendingCompactionEntry
	treeStore           *TreeStore

	files          *ChangesetFiles
	leavesWriter   *StructWriter[LeafLayout]
	branchesWriter *StructWriter[BranchLayout]
	cpInfoWriter   *StructWriter[CheckpointInfo]
	walWriter      *WALWriter
	kvlogWriter    *KVDataWriter
	orphanWriter   *OrphanWriter

	endVersion uint32

	// offsetCache holds the updated 1-based offsets of nodes affected by compacting.
	// these are then used to update BranchLayout's left and right offsets.
	offsetCache map[NodeID]uint32

	ctx context.Context
}

type pendingCompactionEntry struct {
	orig                *Changeset
	orphanRewriter      *OrphanRewriter
	origStartCheckpoint uint32
}

func NewCompactor(ctx context.Context, reader *ChangesetReader, opts CompactOptions, store *TreeStore) (*Compactor, error) {
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
		criteria:        opts.RetainCriteria,
		walStartVersion: opts.WALStartVersion,
		treeStore:       store,
		files:           newFiles,
		walWriter:       NewWALWriter(newFiles.WALFile()),
		kvlogWriter:     NewKVDataWriter(newFiles.KVDataFile()),
		leavesWriter:    NewStructWriter[LeafLayout](newFiles.leavesFile),
		branchesWriter:  NewStructWriter[BranchLayout](newFiles.branchesFile),
		cpInfoWriter:    NewStructWriter[CheckpointInfo](newFiles.checkpointsFile),
		offsetCache:     make(map[NodeID]uint32),
		orphanWriter:    NewOrphanWriter(newFiles.OrphansFile()),
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
			return fmt.Errorf("failed to add changeset to compactor: %v; additionally failed to abort compactor during cleanup: %w", err, abortErr)
		}
		return fmt.Errorf("failed to add changeset to compactor: %w", err)
	}
	return nil
}

func (c *Compactor) doAddChangeset(reader *ChangesetReader) error {
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

	orphanRewriter, err := NewOrphanRewriter(reader.changeset.orphanWriter)
	if err != nil {
		return fmt.Errorf("failed to create orphan rewriter: %w", err)
	}
	c.treeStore.LockOrphanProc()
	deleteMap, err := orphanRewriter.Preprocess(c.criteria, c.orphanWriter)
	if err != nil {
		c.treeStore.UnlockOrphanProc()
		return fmt.Errorf("failed to preprocess orphans for compaction: %w", err)
	}
	c.treeStore.UnlockOrphanProc()

	logger.DebugContext(c.ctx, "processing changeset for compaction", "numCheckpoints", numCheckpoints)
	for i := 0; i < numCheckpoints; i++ {
		cpInfo := cpInfo.UnsafeItem(uint32(i)) // copy
		newLeafStartIdx := uint32(0)
		newLeafEndIdx := uint32(0)
		leafStartOffset := cpInfo.Leaves.StartOffset
		leafCount := cpInfo.Leaves.Count
		newLeafStartOffset := uint32(c.leavesWriter.Count())
		newLeafCount := uint32(0)
		// Iterate leaves
		// For each leaf, check if it should be retained
		for j := uint32(0); j < leafCount; j++ {
			existingLeaf := LeafPersisted{store: reader, layout: leavesData.UnsafeItem(leafStartOffset + j)}
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
			keyOffset, keyInWAL := walRewriteInfo.KeyOffsetRemapping[leaf.KeyOffset.ToUint64()]
			if !keyInWAL {
				key, err := existingLeaf.Key()
				if err != nil {
					return fmt.Errorf("failed to read key for leaf %s: %w", id, err)
				}
				keyOffset, err = c.kvlogWriter.WriteKeyBlob(key.UnsafeBytes())
				if err != nil {
					return fmt.Errorf("failed to write key blob for leaf %s: %w", id, err)
				}
			}
			leaf.KeyOffset = NewUint40(keyOffset)
			leaf.SetKeyInKVData(!keyInWAL)

			//  remap value offset
			valOffset, valInWAL := walRewriteInfo.ValueOffsetRemapping[leaf.ValueOffset.ToUint64()]
			if !valInWAL {
				val, err := existingLeaf.Value()
				if err != nil {
					return fmt.Errorf("failed to read value for leaf %s: %w", id, err)
				}
				valOffset, err = c.kvlogWriter.WriteValueBlob(val.UnsafeBytes())
				if err != nil {
					return fmt.Errorf("failed to write value blob for leaf %s: %w", id, err)
				}
			}
			leaf.ValueOffset = NewUint40(valOffset)
			leaf.SetValueInKVData(!valInWAL)

			err := c.leavesWriter.Append(&leaf)
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
			existingBranch := BranchPersisted{store: reader, layout: branchesData.UnsafeItem(branchStartOffset + j)}
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
			}
			if newRightOffset, ok := c.offsetCache[branch.Right]; ok {
				branch.RightOffset = newRightOffset
			}

			// remap key offset
			keyOffset, keyInWAL := walRewriteInfo.KeyOffsetRemapping[branch.KeyOffset.ToUint64()]
			if !keyInWAL {
				key, err := existingBranch.Key()
				if err != nil {
					return fmt.Errorf("failed to read key for branch %s: %w", id, err)
				}
				keyOffset, err = c.kvlogWriter.WriteKeyBlob(key.UnsafeBytes())
				if err != nil {
					return fmt.Errorf("failed to write key blob for branch %s: %w", id, err)
				}
			}
			branch.KeyOffset = NewUint40(keyOffset)
			branch.SetKeyInKVData(!keyInWAL)

			err := c.branchesWriter.Append(&branch)
			if err != nil {
				return fmt.Errorf("failed to append branch %s: %w", id, err)
			}
			c.offsetCache[id] = uint32(c.branchesWriter.Count())
		}

		if newBranchCount == 0 && newLeafCount == 0 && c.cpInfoWriter.Count() == 0 {
			// This is the first checkpoint in the compacted output and it has no nodes.
			// Skip it if the checkpoint is unreachable because WAL entries before
			// walStartVersion were truncated. To replay from checkpoint version V,
			// we need WAL entries starting at V+1.
			if cpInfo.Version+1 < c.walStartVersion {
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
			Checkpoint: cpInfo.Checkpoint,
			Version:    cpInfo.Version,
			RootID:     cpInfo.RootID,
		}

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
			return nil, fmt.Errorf("failed to flush data during compaction seal: %v; additionally failed to abort compactor during cleanup: %w", err, errAbort)
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

func (c *Compactor) switchoverChangesets() (*Changeset, error) {
	c.treeStore.LockOrphanProc()
	defer c.treeStore.UnlockOrphanProc()

	cs, err := c.finalize()
	if err != nil {
		// if we error at this point, we abort and cleanup, past this point there can be no aborting
		errAbort := c.Abort()
		if errAbort != nil {
			return nil, fmt.Errorf("failed to finalize changeset during compaction seal: %v; additionally failed to abort compactor during cleanup: %w", err, errAbort)
		}
		return nil, fmt.Errorf("failed to finalize changeset during compaction seal: %w, but aborted successfully", err)
	}

	// IMPORTANT: an abort CANNOT happen past this point, otherwise we can lose both the original and compacted changesets and cause data loss!
	// this operation does not error, but critically it marks the original changesets for deletion!
	// this operation MUST happen while we are holding the orphan proc lock to prevent orphans from going to the old changesets now that we've switched over
	for _, entry := range c.processedChangesets {
		entry.orig.MarkCompacted(cs)
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

	cs, err := OpenChangeset(c.treeStore, finalDir)
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
