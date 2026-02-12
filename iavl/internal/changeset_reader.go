package internal

import (
	"errors"
	"fmt"
)

type ChangesetReader struct {
	changeset *Changeset // we keep a reference to the parent changeset handle

	walReader             *KVDataReader
	kvDataReader          *KVDataReader
	branchesData          *NodeMmap[BranchLayout]
	leavesData            *NodeMmap[LeafLayout]
	firstCheckpoint       uint32
	lastCheckpoint        uint32
	checkpointsContiguous bool // true if checkpoint numbers are contiguous (no gaps from compaction)
	checkpointsInfo       *StructMmap[CheckpointInfo]
}

func NewChangesetReader(changeset *Changeset) (*ChangesetReader, error) {
	cr := &ChangesetReader{changeset: changeset}

	var err error

	files := changeset.files

	cr.walReader, err = NewKVDataReader(files.WALFile())
	if err != nil {
		return nil, fmt.Errorf("failed to open WAL data store: %w", err)
	}

	cr.kvDataReader, err = NewKVDataReader(files.KVDataFile())
	if err != nil {
		return nil, fmt.Errorf("failed to open KV data store: %w", err)
	}

	cr.leavesData, err = NewNodeReader[LeafLayout](files.LeavesFile())
	if err != nil {
		return nil, fmt.Errorf("failed to open leaves data file: %w", err)
	}

	cr.branchesData, err = NewNodeReader[BranchLayout](files.BranchesFile())
	if err != nil {
		return nil, fmt.Errorf("failed to open branches data file: %w", err)
	}

	cr.checkpointsInfo, err = NewStructMmap[CheckpointInfo](files.CheckpointsFile())
	if err != nil {
		return nil, fmt.Errorf("failed to open checkpoints data file: %w", err)
	}

	if cr.checkpointsInfo.Count() > 0 {
		count := uint32(cr.checkpointsInfo.Count())
		firstInfo := cr.checkpointsInfo.UnsafeItem(0)
		cr.firstCheckpoint = firstInfo.Checkpoint
		lastInfo := cr.checkpointsInfo.UnsafeItem(count - 1)
		cr.lastCheckpoint = lastInfo.Checkpoint
		cr.checkpointsContiguous = cr.lastCheckpoint-cr.firstCheckpoint+1 == count
	}

	return cr, nil
}

func (cr *ChangesetReader) WALData() *KVDataReader {
	return cr.walReader
}

func (cr *ChangesetReader) KVData() *KVDataReader {
	return cr.kvDataReader
}

func (cr *ChangesetReader) Changeset() *Changeset {
	return cr.changeset
}

func (cr *ChangesetReader) ResolveLeafByIndex(fileIdx uint32) (*LeafLayout, error) {
	if fileIdx == 0 || fileIdx > uint32(cr.leavesData.Count()) {
		return nil, fmt.Errorf("leaf file index %d out of range (have 1..%d)", fileIdx, cr.leavesData.Count())
	}

	fileIdx-- // convert to 0-based index
	return cr.leavesData.UnsafeItem(fileIdx), nil
}

func (cr *ChangesetReader) ResolveBranchByIndex(fileIdx uint32) (*BranchLayout, error) {
	if fileIdx == 0 || fileIdx > uint32(cr.branchesData.Count()) {
		return nil, fmt.Errorf("branch file index %d out of range (have 1..%d)", fileIdx, cr.branchesData.Count())
	}

	fileIdx-- // convert to 0-based index
	return cr.branchesData.UnsafeItem(fileIdx), nil
}

func (cr *ChangesetReader) ResolveLeafByID(id NodeID) (*LeafLayout, error) {
	if !id.IsLeaf() {
		return nil, fmt.Errorf("node ID %s is not a leaf", id.String())
	}
	info, err := cr.GetCheckpointInfo(id.Checkpoint())
	if err != nil {
		return nil, err
	}
	return cr.leavesData.FindByID(id, &info.Leaves)
}

func (cr *ChangesetReader) ResolveBranchByID(id NodeID) (*BranchLayout, error) {
	if id.IsLeaf() {
		return nil, fmt.Errorf("node ID %s is not a branch", id.String())
	}

	info, err := cr.GetCheckpointInfo(id.Checkpoint())
	if err != nil {
		return nil, err
	}
	return cr.branchesData.FindByID(id, &info.Branches)
}

func (cr *ChangesetReader) GetCheckpointInfo(checkpoint uint32) (*CheckpointInfo, error) {
	if checkpoint < cr.firstCheckpoint || checkpoint > cr.lastCheckpoint {
		return nil, fmt.Errorf("checkpoint %d out of range for changeset %s (have %d..%d)",
			checkpoint, cr.changeset.Files().Dir(), cr.firstCheckpoint, cr.lastCheckpoint)
	}
	if cr.checkpointsContiguous {
		item := cr.checkpointsInfo.UnsafeItem(checkpoint - cr.firstCheckpoint)
		if item.Checkpoint != checkpoint {
			return nil, fmt.Errorf("checkpoint data corruption in changeset %s: expected checkpoint %d at index %d, got %d",
				cr.changeset.Files().Dir(), checkpoint, checkpoint-cr.firstCheckpoint, item.Checkpoint)
		}
		return item, nil
	}
	// non-contiguous: binary search by checkpoint number
	idx := cr.checkpointsInfo.BinarySearch(func(c *CheckpointInfo) bool {
		return c.Checkpoint >= checkpoint
	})
	if idx < cr.checkpointsInfo.Count() {
		item := cr.checkpointsInfo.UnsafeItem(uint32(idx))
		if item.Checkpoint == checkpoint {
			return item, nil
		}
	}
	return nil, fmt.Errorf("checkpoint %d not found in changeset %s (have %d..%d, likely pruned during compaction)",
		checkpoint, cr.changeset.Files().Dir(), cr.firstCheckpoint, cr.lastCheckpoint)
}

func (cr *ChangesetReader) ResolveByID(id NodeID) (Node, error) {
	if id.IsLeaf() {
		leafLayout, err := cr.ResolveLeafByID(id)
		if err != nil {
			return nil, err
		}
		return &LeafPersisted{
			store:  cr,
			layout: leafLayout,
		}, nil
	} else {
		branchLayout, err := cr.ResolveBranchByID(id)
		if err != nil {
			return nil, err
		}
		return &BranchPersisted{
			store:  cr,
			layout: branchLayout,
		}, nil
	}
}

func (cr *ChangesetReader) ResolveByFileIndex(id NodeID, idx uint32) (Node, error) {
	if id.IsLeaf() {
		leafLayout, err := cr.ResolveLeafByIndex(idx)
		if err != nil {
			return nil, err
		}
		return &LeafPersisted{
			store:  cr,
			layout: leafLayout,
		}, nil
	} else {
		branchLayout, err := cr.ResolveBranchByIndex(idx)
		if err != nil {
			return nil, err
		}
		return &BranchPersisted{
			store:  cr,
			layout: branchLayout,
		}, nil
	}
}

// FirstCheckpoint returns the first checkpoint number in this changeset.
// If there are no checkpoints, 0 is returned.
func (cr *ChangesetReader) FirstCheckpoint() uint32 {
	return cr.firstCheckpoint
}

func (cr *ChangesetReader) LastCheckpoint() uint32 {
	return cr.lastCheckpoint
}

// CheckpointRootInfo contains the resolved root NodePointer for a checkpoint along with the checkpoint's version and number.
// If we could not resolve a checkpoint at all, Version will be zero.
// If we resolved a checkpoint with an empty tree, Root will be nil but Version will be set to the checkpoint version.
// Checkpoint should be non-zero whenever Version is non-zero.
type CheckpointRootInfo struct {
	Root       *NodePointer
	Version    uint32
	Checkpoint uint32
}

// LatestCheckpointRoot returns the latest checkpoint root in the changeset if one is available.
// If there is no available checkpoint, CheckpointRootInfo.Version will be zero.
func (cr *ChangesetReader) LatestCheckpointRoot() CheckpointRootInfo {
	count := cr.checkpointsInfo.Count()
	if count == 0 {
		return CheckpointRootInfo{}
	}
	return cr.latestValidCheckpoint(count - 1)
}

// latestValidCheckpoint finds the latest checkpoint at or before startIdx that has a root.
// startIdx is a 0-based index into the checkpoints info array.
// A valid checkpoint is considered to be one that has a root retained in the changeset,
// where the root ID's checkpoint corresponds to the actual checkpoint.
// If for example we have checkpoint 100, but its root ID has checkpoint number 99,
// we actually need to navigate to checkpoint 99 to confirm that the root was retained.
// If there is no available checkpoint, CheckpointRootInfo.Version will be zero.
func (cr *ChangesetReader) latestValidCheckpoint(startIdx int) CheckpointRootInfo {
	for i := startIdx; i >= 0; i-- {
		info := cr.checkpointsInfo.UnsafeItem(uint32(i))
		rootID := info.RootID
		if rootID.IsEmpty() {
			// if root is empty, we have an empty tree at this checkpoint
			return CheckpointRootInfo{
				Version:    info.Version,
				Checkpoint: info.Checkpoint,
			}
		}

		// check if the root was retained in this changeset by looking at the bounds
		// of the checkpoint where the root was actually created
		rootCheckpoint := rootID.Checkpoint()
		sourceInfo := info
		if rootCheckpoint != info.Checkpoint {
			// root is from a different checkpoint (no tree changes at this checkpoint),
			// look up bounds from the root's source checkpoint
			var err error
			sourceInfo, err = cr.GetCheckpointInfo(rootCheckpoint)
			if err != nil {
				// root's source checkpoint is not in this changeset (compacted away or pruned)
				// keep looking at earlier checkpoints just in case
				continue
			}
		}

		var nodeSetInfo *NodeSetInfo
		if rootID.IsLeaf() {
			nodeSetInfo = &sourceInfo.Leaves
		} else {
			nodeSetInfo = &sourceInfo.Branches
		}
		idx := rootID.Index()
		if idx < nodeSetInfo.StartIndex || idx > nodeSetInfo.EndIndex {
			// root node was compacted away
			// keep looking for an earlier checkpoint with a root that was retained
			continue
		}
		return CheckpointRootInfo{
			Root: &NodePointer{
				id:        rootID,
				changeset: cr.changeset,
			},
			Version:    info.Version,
			Checkpoint: info.Checkpoint,
		}
	}
	return CheckpointRootInfo{}
}

// CheckpointForVersion finds the nearest checkpoint <= targetVersion that has a root.
// If there is no available checkpoint in this changeset, CheckpointRootInfo.Version will be zero.
func (cr *ChangesetReader) CheckpointForVersion(targetVersion uint32) CheckpointRootInfo {
	count := cr.checkpointsInfo.Count()
	if count == 0 {
		return CheckpointRootInfo{}
	}

	// binary search for the first checkpoint with version > targetVersion,
	// then step back one to get the floor (last checkpoint with version <= targetVersion)
	floorIdx := cr.checkpointsInfo.BinarySearch(func(c *CheckpointInfo) bool {
		return c.Version > targetVersion
	}) - 1

	if floorIdx < 0 {
		// no checkpoint found with version <= targetVersion
		return CheckpointRootInfo{}
	}

	return cr.latestValidCheckpoint(floorIdx)
}

func (cr *ChangesetReader) TotalBytes() int {
	return cr.leavesData.TotalBytes() +
		cr.branchesData.TotalBytes() +
		cr.checkpointsInfo.TotalBytes() +
		cr.kvDataReader.Len() +
		cr.walReader.Len()
}

func (cr *ChangesetReader) Describe() ChangesetDescription {
	checkpoints := make([]CheckpointInfo, 0, cr.checkpointsInfo.Count())
	for i := 0; i < cr.checkpointsInfo.Count(); i++ {
		info := cr.checkpointsInfo.UnsafeItem(uint32(i))
		checkpoints = append(checkpoints, *info) // copy
	}
	return ChangesetDescription{
		StartVersion:  cr.changeset.Files().StartVersion(),
		EndVersion:    cr.changeset.Files().EndVersion(),
		CompactedAt:   cr.changeset.Files().CompactedAtVersion(),
		TotalLeaves:   cr.leavesData.Count(),
		TotalBranches: cr.branchesData.Count(),
		TotalBytes:    cr.TotalBytes(),
		KVLogSize:     cr.kvDataReader.Len(),
		WALSize:       cr.walReader.Len(),
		Checkpoints:   checkpoints,
	}
}

func (cr *ChangesetReader) Close() error {
	errs := []error{
		cr.leavesData.Close(),
		cr.branchesData.Close(),
		cr.checkpointsInfo.Close(),
		cr.kvDataReader.Close(),
		cr.walReader.Close(),
	}
	return errors.Join(errs...)
}
