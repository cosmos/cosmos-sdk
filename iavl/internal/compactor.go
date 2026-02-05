package internal

import (
	"context"
	"fmt"
)

type CompactOptions struct {
	RetainCriteria  RetainCriteria
	CompactedAt     uint32 // version at which compaction is done
	WALStartVersion uint32
}

type RetainCriteria func(createVersion, orphanVersion uint32) bool

type Compactor struct {
	criteria        RetainCriteria
	walStartVersion uint32

	processedChangesets []*Changeset
	treeStore           *TreeStore

	originalKvLogPath string
	files             *ChangesetFiles
	leavesWriter      *StructWriter[LeafLayout]
	branchesWriter    *StructWriter[BranchLayout]
	cpInfoWriter      *StructWriter[CheckpointInfo]
	walWriter         *WALWriter
	kvlogWriter       *KVDataWriter
	orphanWriter      *OrphanWriter

	keyCache map[string]uint32
	// offsetCache holds the updated 1-based offsets of nodes affected by compacting.
	// these are then used to update BranchLayout's left and right offsets.
	offsetCache map[NodeID]uint32

	// Running totals across all processed changesets
	leafOrphanCount          uint32
	branchOrphanCount        uint32
	leafOrphanVersionTotal   uint64
	branchOrphanVersionTotal uint64
	ctx                      context.Context
}

func NewCompacter(ctx context.Context, reader *ChangesetReader, opts CompactOptions, store *TreeStore) (*Compactor, error) {
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
		orphanWriter:    NewOrphanWriter(newFiles.orphansFile),
		offsetCache:     make(map[NodeID]uint32),
	}

	// Process first changeset immediately
	err = c.processChangeset(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to process initial changeset: %w", err)
	}

	return c, nil
}

func (c *Compactor) processChangeset(reader *ChangesetReader) error {
	// TODO rewrite WAL
	_, err := RewriteWAL(c.walWriter, reader.changeset.files.WALFile(), uint64(c.walStartVersion))
	if err != nil {
		return fmt.Errorf("failed to rewrite WAL during compaction: %w", err)
	}

	cpInfo := reader.checkpointsInfo
	numCheckpoints := cpInfo.Count()
	leavesData := reader.leavesData
	branchesData := reader.branchesData

	// flush orphan writer to ensure all orphans are written before reading
	err = reader.changeset.orphanWriter.Flush()
	if err != nil {
		return fmt.Errorf("failed to flush orphan writer before reading orphan map: %w", err)
	}
	orphanMap, err := ReadOrphanLog(reader.changeset.files.orphansFile)
	if err != nil {
		return fmt.Errorf("failed to read orphan map: %w", err)
	}

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
			orphanVersion := orphanMap[id]
			retain := orphanVersion == 0 || c.criteria(leaf.Version, orphanVersion)
			if !retain {
				continue
			}

			if orphanVersion != 0 {
				c.leafOrphanCount++
				c.leafOrphanVersionTotal += uint64(orphanVersion)
			}

			if newLeafStartIdx == 0 {
				newLeafStartIdx = id.Index()
			}
			newLeafEndIdx = id.Index()
			newLeafCount++

			//var keyOffset
			//keyOffset, ok := walRewriteInfo.KeyOffsetRemapping[uint64(leaf.KeyOffset)]
			//if !ok {
			//	key, err := existingLeaf.Key()
			//	if err != nil {
			//		return fmt.Errorf("failed to read key for leaf %s: %w", id, err)
			//	}
			//	c.kvlogWriter.WriteKeyBlob(key.UnsafeBytes())
			//}
			//k, v, err := reader.ReadKV(id, leaf.KeyOffset)
			//if err != nil {
			//	return fmt.Errorf("failed to read KV for leaf %s: %w", id, err)
			//}
			//
			//	offset, err := c.kvlogWriter.WriteKV(k, v)
			//	if err != nil {
			//		return fmt.Errorf("failed to write KV for leaf %s: %w", id, err)
			//	}
			//
			//	leaf.KeyOffset = offset
			//	c.keyCache[unsafeBytesToString(k)] = offset
			//} else {
			//	// When not compacting WAL, add offset delta
			//	leaf.KeyOffset += kvOffsetDelta
			//}

			err := c.leavesWriter.Append(&leaf)
			if err != nil {
				return fmt.Errorf("failed to append leaf %s: %w", id, err)
			}

			c.offsetCache[id] = uint32(c.leavesWriter.Count())

			if orphanVersion != 0 {
				if err := c.orphanWriter.WriteOrphan(orphanVersion, id); err != nil {
					return fmt.Errorf("failed to write retained orphan leaf %s: %w", id, err)
				}
			}
		}

		newBranchStartIdx := uint32(0)
		newBranchEndIdx := uint32(0)
		branchStartOffset := cpInfo.Branches.StartOffset
		branchCount := cpInfo.Branches.Count
		newBranchStartOffset := uint32(c.branchesWriter.Count())
		newBranchCount := uint32(0)
		for j := uint32(0); j < branchCount; j++ {
			branch := *branchesData.UnsafeItem(branchStartOffset + j) // copy
			id := branch.ID
			orphanVersion := orphanMap[id]
			retain := orphanVersion == 0 || c.criteria(branch.Version, orphanVersion)
			if !retain {
				continue
			}

			if orphanVersion != 0 {
				c.branchOrphanCount++
				c.branchOrphanVersionTotal += uint64(orphanVersion)
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

			// TODO lookup new key offset
			//if c.compactWAL {
			//	k, err := reader.ReadK(id, branch.KeyOffset)
			//	if err != nil {
			//		return fmt.Errorf("failed to read key for branch %s: %w", id, err)
			//	}
			//	offset, ok := c.keyCache[unsafeBytesToString(k)]
			//	if !ok {
			//		offset, err = c.kvlogWriter.WriteK(k)
			//	}
			//	if err != nil {
			//		return fmt.Errorf("failed to write key for branch %s: %w", id, err)
			//	}
			//	branch.KeyOffset = offset
			//} else {
			//	// When not compacting WAL, add offset delta
			//	branch.KeyOffset += kvOffsetDelta
			//}

			err := c.branchesWriter.Append(&branch)
			if err != nil {
				return fmt.Errorf("failed to append branch %s: %w", id, err)
			}
			c.offsetCache[id] = uint32(c.branchesWriter.Count())

			if orphanVersion != 0 {
				if err := c.orphanWriter.WriteOrphan(orphanVersion, id); err != nil {
					return fmt.Errorf("failed to write retained orphan leaf %s: %w", id, err)
				}
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
			RootID: cpInfo.RootID,
		}

		err := c.cpInfoWriter.Append(cpInfo)
		if err != nil {
			return fmt.Errorf("failed to append checkpoint info for checkpoint %d: %w", cpInfo.Version, err)
		}
	}

	// Track this changeset as processed
	c.processedChangesets = append(c.processedChangesets, reader.Changeset())

	return nil
}

//func (c *Compactor) AddChangeset(cs *Changeset) error {
//	// TODO: Support joining changesets when CompactWAL=false
//	// This requires copying the entire KV log and tracking cumulative offsets
//	if !c.compactWAL {
//		return fmt.Errorf("joining changesets is only supported when CompactWAL=true")
//	}
//	return c.processChangeset(cs)
//}
//
//func (c *Compactor) Seal() (*Changeset, error) {
//	if len(c.processedChangesets) == 0 {
//		return nil, fmt.Errorf("no changesets processed")
//	}
//
//	info := c.files.info
//	info.StartVersion = c.processedChangesets[0].files.info.StartVersion
//	info.EndVersion = c.processedChangesets[len(c.processedChangesets)-1].files.info.EndVersion
//	info.LeafOrphans = c.leafOrphanCount
//	info.BranchOrphans = c.branchOrphanCount
//	info.LeafOrphanVersionTotal = c.leafOrphanVersionTotal
//	info.BranchOrphanVersionTotal = c.branchOrphanVersionTotal
//
//	errs := []error{
//		c.leavesWriter.Flush(),
//		c.branchesWriter.Flush(),
//		c.cpInfoWriter.Flush(),
//		c.orphanWriter.Flush(),
//		c.files.RewriteInfo(),
//	}
//	if c.kvlogWriter != nil {
//		errs = append(errs, c.kvlogWriter.Flush())
//	}
//	if err := errors.Join(errs...); err != nil {
//		return nil, fmt.Errorf("failed to flush data during compaction seal: %w", err)
//	}
//	if err := c.files.MarkReady(); err != nil {
//		return nil, fmt.Errorf("failed to mark changeset as ready during compaction seal: %w", err)
//	}
//
//	cs := NewChangeset(c.treeStore)
//	err := cs.InitOwned(c.files)
//	if err != nil {
//		return nil, fmt.Errorf("failed to initialize sealed changeset: %w", err)
//	}
//
//}
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
