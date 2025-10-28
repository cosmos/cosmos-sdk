package iavlx

import (
	"errors"
	"fmt"

	"cosmossdk.io/log"
)

type CompactOptions struct {
	RetainCriteria RetainCriteria
	CompactWAL     bool
	CompactedAt    uint32 // version at which compaction is done
}

type RetainCriteria func(createVersion, orphanVersion uint32) bool

type Compactor struct {
	logger log.Logger

	criteria   RetainCriteria
	compactWAL bool

	processedChangesets []*Changeset
	treeStore           *TreeStore

	originalKvLogPath string
	files             *ChangesetFiles
	leavesWriter      *StructWriter[LeafLayout]
	branchesWriter    *StructWriter[BranchLayout]
	versionsWriter    *StructWriter[VersionInfo]
	kvlogWriter       *KVLogWriter

	keyCache map[string]uint32
	// offsetCache holds the updated 1-based offsets of nodes affected by compacting.
	// these are then used to update BranchLayout's left and right offsets.
	offsetCache map[NodeID]uint32

	// Running totals across all processed changesets
	leafOrphanCount          uint32
	branchOrphanCount        uint32
	leafOrphanVersionTotal   uint64
	branchOrphanVersionTotal uint64
}

func NewCompacter(logger log.Logger, reader *Changeset, opts CompactOptions, store *TreeStore) (*Compactor, error) {
	if reader.files == nil {
		return nil, fmt.Errorf("changeset has no associated files, cannot compact a shared changeset reader which files set to nil")
	}
	files := reader.files
	startingVersion := files.StartVersion()
	lastCompactedAt := files.CompactedAtVersion()
	if lastCompactedAt >= opts.CompactedAt {
		return nil, fmt.Errorf("cannot compact changeset starting at version %d which was last compacted at %d to an earlier or same version %d",
			startingVersion, lastCompactedAt, opts.CompactedAt)
	}

	// if we're not compacting the WAL, we can reuse the existing KV log path
	kvlogPath := reader.files.KVLogPath()
	// if we're compacting the WAL, create a new KV log path
	if opts.CompactWAL {
		kvlogPath = ""
	}

	newFiles, err := OpenChangesetFiles(files.TreeDir(), files.StartVersion(), opts.CompactedAt, kvlogPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open new changeset files: %w", err)
	}

	var kvlogWriter *KVLogWriter
	// we only need a new KV log writer if we're compacting the WAL, otherwise it should be nil
	if opts.CompactWAL {
		kvlogWriter = NewKVDataWriter(newFiles.kvlogFile)
	}

	c := &Compactor{
		logger:            logger,
		criteria:          opts.RetainCriteria,
		compactWAL:        opts.CompactWAL,
		treeStore:         store,
		files:             newFiles,
		originalKvLogPath: reader.files.KVLogPath(),
		kvlogWriter:       kvlogWriter,
		leavesWriter:      NewStructWriter[LeafLayout](newFiles.leavesFile),
		branchesWriter:    NewStructWriter[BranchLayout](newFiles.branchesFile),
		versionsWriter:    NewStructWriter[VersionInfo](newFiles.versionsFile),
		keyCache:          make(map[string]uint32),
		offsetCache:       make(map[NodeID]uint32),
	}

	// Process first changeset immediately
	err = c.processChangeset(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to process initial changeset: %w", err)
	}

	return c, nil
}

func (c *Compactor) processChangeset(reader *Changeset) error {
	// Compute KV offset delta for non-CompactWAL mode
	kvOffsetDelta := uint32(0)
	if c.kvlogWriter != nil && !c.compactWAL {
		kvOffsetDelta = uint32(c.kvlogWriter.Size())
	}

	versionsData := reader.versionsData
	numVersions := versionsData.Count()
	leavesData := reader.leavesData
	branchesData := reader.branchesData
	skippedBranches := 0

	c.logger.Debug("processing changeset for compaction", "versions", numVersions)
	for i := 0; i < numVersions; i++ {
		c.logger.Debug("compacting version", "version", reader.info.StartVersion+uint32(i))
		verInfo := *versionsData.UnsafeItem(uint32(i)) // copy
		newLeafStartIdx := uint32(0)
		newLeafEndIdx := uint32(0)
		leafStartOffset := verInfo.Leaves.StartOffset
		leafCount := verInfo.Leaves.Count
		newLeafStartOffset := uint32(c.leavesWriter.Count())
		newLeafCount := uint32(0)
		// Iterate leaves
		// For each leaf, check if it should be retained
		for j := uint32(0); j < leafCount; j++ {
			leaf := *leavesData.UnsafeItem(leafStartOffset + j) // copy
			id := leaf.Id
			retain := leaf.OrphanVersion == 0 || c.criteria(uint32(id.Version()), leaf.OrphanVersion)
			if !retain {
				continue
			}

			if leaf.OrphanVersion != 0 {
				c.leafOrphanCount++
				c.leafOrphanVersionTotal += uint64(leaf.OrphanVersion)
			}

			if newLeafStartIdx == 0 {
				newLeafStartIdx = id.Index()
			}
			newLeafEndIdx = id.Index()
			newLeafCount++

			if c.compactWAL {
				k, v, err := reader.ReadKV(id, leaf.KeyOffset)
				if err != nil {
					return fmt.Errorf("failed to read KV for leaf %s: %w", id, err)
				}

				offset, err := c.kvlogWriter.WriteKV(k, v)
				if err != nil {
					return fmt.Errorf("failed to write KV for leaf %s: %w", id, err)
				}

				leaf.KeyOffset = offset
				c.keyCache[string(k)] = offset
			} else {
				// When not compacting WAL, add offset delta
				leaf.KeyOffset += kvOffsetDelta
			}

			err := c.leavesWriter.Append(&leaf)
			if err != nil {
				return fmt.Errorf("failed to append leaf %s: %w", id, err)
			}

			c.offsetCache[id] = uint32(c.leavesWriter.Count())
		}

		newBranchStartIdx := uint32(0)
		newBranchEndIdx := uint32(0)
		branchStartOffset := verInfo.Branches.StartOffset
		branchCount := verInfo.Branches.Count
		newBranchStartOffset := uint32(c.branchesWriter.Count())
		newBranchCount := uint32(0)
		for j := uint32(0); j < branchCount; j++ {
			branch := *branchesData.UnsafeItem(branchStartOffset + j) // copy
			id := branch.Id
			retain := branch.OrphanVersion == 0 || c.criteria(uint32(id.Version()), branch.OrphanVersion)
			if !retain {
				skippedBranches++
				continue
			}

			if branch.OrphanVersion != 0 {
				c.branchOrphanCount++
				c.branchOrphanVersionTotal += uint64(branch.OrphanVersion)
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

			if c.compactWAL {
				k, err := reader.ReadK(id, branch.KeyOffset)
				if err != nil {
					return fmt.Errorf("failed to read key for branch %s: %w", id, err)
				}
				offset, ok := c.keyCache[string(k)]
				if !ok {
					offset, err = c.kvlogWriter.WriteK(k)
				}
				if err != nil {
					return fmt.Errorf("failed to write key for branch %s: %w", id, err)
				}
				branch.KeyOffset = offset
			} else {
				// When not compacting WAL, add offset delta
				branch.KeyOffset += kvOffsetDelta
			}

			err := c.branchesWriter.Append(&branch)
			if err != nil {
				return fmt.Errorf("failed to append branch %s: %w", id, err)
			}
			c.offsetCache[id] = uint32(c.branchesWriter.Count())
		}

		verInfo = VersionInfo{
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
			RootID: verInfo.RootID,
		}

		err := c.versionsWriter.Append(&verInfo)
		if err != nil {
			return fmt.Errorf("failed to append version info for version %d: %w", reader.info.StartVersion+uint32(i), err)
		}
	}

	// Track this changeset as processed
	c.processedChangesets = append(c.processedChangesets, reader)

	return nil
}

func (c *Compactor) AddChangeset(cs *Changeset) error {
	// TODO: Support joining changesets when CompactWAL=false
	// This requires copying the entire KV log and tracking cumulative offsets
	if !c.compactWAL {
		return fmt.Errorf("joining changesets is only supported when CompactWAL=true")
	}
	return c.processChangeset(cs)
}

func (c *Compactor) Seal() (*Changeset, error) {
	if len(c.processedChangesets) == 0 {
		return nil, fmt.Errorf("no changesets processed")
	}

	info := c.files.info
	info.StartVersion = c.processedChangesets[0].info.StartVersion
	info.EndVersion = c.processedChangesets[len(c.processedChangesets)-1].info.EndVersion
	info.LeafOrphans = c.leafOrphanCount
	info.BranchOrphans = c.branchOrphanCount
	info.LeafOrphanVersionTotal = c.leafOrphanVersionTotal
	info.BranchOrphanVersionTotal = c.branchOrphanVersionTotal

	errs := []error{
		c.leavesWriter.Flush(),
		c.branchesWriter.Flush(),
		c.versionsWriter.Flush(),
		c.files.infoMmap.Flush(),
	}
	if c.kvlogWriter != nil {
		errs = append(errs, c.kvlogWriter.Flush())
	}
	if err := errors.Join(errs...); err != nil {
		return nil, fmt.Errorf("failed to flush data during compaction seal: %w", err)
	}

	cs := NewChangeset(c.treeStore)
	err := cs.InitOwned(c.files)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize sealed changeset: %w", err)
	}

	return cs, nil
}

func (c *Compactor) Abort() error {
	err := c.files.Close()
	if err != nil {
		return fmt.Errorf("failed to close compactor files during cleanup: %w", err)
	}
	return c.files.DeleteFiles(ChangesetDeleteArgs{
		SaveKVLogPath: c.originalKvLogPath,
	})
}

func (c *Compactor) TotalBytes() int {
	total := c.leavesWriter.Size() + c.branchesWriter.Size() + c.versionsWriter.Size()
	if c.kvlogWriter != nil {
		total += c.kvlogWriter.Size()
	}
	return total
}
