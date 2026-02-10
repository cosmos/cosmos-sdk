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

	originalKvLogPath string
	files             *ChangesetFiles
	leavesWriter      *StructWriter[LeafLayout]
	branchesWriter    *StructWriter[BranchLayout]
	cpInfoWriter      *StructWriter[CheckpointInfo]
	walWriter         *WALWriter
	kvlogWriter       *KVDataWriter
	orphanWriter      *OrphanWriter

	// offsetCache holds the updated 1-based offsets of nodes affected by compacting.
	// these are then used to update BranchLayout's left and right offsets.
	offsetCache map[NodeID]uint32

	ctx context.Context
}

type pendingCompactionEntry struct {
	orig           *Changeset
	orphanRewriter *OrphanRewriter
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
	walRewriteInfo, err := RewriteWAL(c.walWriter, reader.changeset.files.WALFile(), uint64(c.walStartVersion))
	if err != nil {
		return fmt.Errorf("failed to rewrite WAL during compaction: %w", err)
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
			RootID: cpInfo.RootID,
		}

		err := c.cpInfoWriter.Append(cpInfo)
		if err != nil {
			return fmt.Errorf("failed to append checkpoint info for checkpoint %d: %w", cpInfo.Version, err)
		}
	}

	// track this changeset as processed
	c.processedChangesets = append(c.processedChangesets, pendingCompactionEntry{
		orig:           reader.Changeset(),
		orphanRewriter: orphanRewriter,
	})

	return nil
}

func (c *Compactor) Seal() (*Changeset, error) {
	c.treeStore.LockOrphanProc()
	defer c.treeStore.UnlockOrphanProc()

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
		return nil, fmt.Errorf("failed to flush data during compaction seal: %w", err)
	}

	for _, entry := range c.processedChangesets {
		err := entry.orphanRewriter.FinishRewrite(c.orphanWriter)
		if err != nil {
			c.treeStore.UnlockOrphanProc()
			return nil, fmt.Errorf("failed to finish orphan rewrite for a processed changeset during compaction seal: %w", err)
		}
	}
	err := c.orphanWriter.Sync()
	if err != nil {
		return nil, fmt.Errorf("failed to flush orphan data during compaction seal: %w", err)
	}

	cs, err := NewChangeset(c.treeStore, c.files)
	if err != nil {
		return nil, fmt.Errorf("failed to create new changeset for compacted data during compaction seal: %w", err)
	}

	if err := c.files.MarkReady(); err != nil {
		return nil, fmt.Errorf("failed to mark changeset as ready during compaction seal: %w", err)
	}

	for _, entry := range c.processedChangesets {
		entry.orig.MarkCompacted(cs)
	}

	return cs, nil
}

//
//func (c *Compactor) Abort() error {
//	err := c.files.Close()
//	if err != nil {
//		return fmt.Errorf("failed to close compactor files during cleanup: %w", err)
//	}
//	return c.files.DeleteFiles(ChangesetDeleteArgs{
//		SaveKVLogPath: c.originalKvLogPath,
//	})
//}
//
//func (c *Compactor) TotalBytes() int {
//	total := c.leavesWriter.Size() + c.branchesWriter.Size() + c.cpInfoWriter.Size()
//	if c.kvlogWriter != nil {
//		total += c.kvlogWriter.Size()
//	}
//	return total
//}
