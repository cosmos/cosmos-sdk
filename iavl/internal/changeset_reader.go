package internal

import (
	"errors"
	"fmt"
)

type ChangesetReader struct {
	changeset *Changeset // we keep a reference to the parent changeset handle

	walReader       *KVDataReader
	kvDataReader    *KVDataReader
	branchesData    *NodeMmap[BranchLayout]
	leavesData      *NodeMmap[LeafLayout]
	firstCheckpoint uint32
	lastCheckpoint  uint32
	checkpointsInfo *StructMmap[CheckpointInfo]
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
		firstInfo := cr.checkpointsInfo.UnsafeItem(0)
		cr.firstCheckpoint = firstInfo.Checkpoint
		cr.lastCheckpoint = cr.firstCheckpoint + uint32(cr.checkpointsInfo.Count()) - 1
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
		return nil, fmt.Errorf("checkpoint %d out of range for changeset (have %d..%d)", checkpoint, cr.firstCheckpoint, cr.lastCheckpoint)
	}
	return cr.checkpointsInfo.UnsafeItem(checkpoint - cr.firstCheckpoint), nil
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
	return cr.LatestValidCheckpoint(cr.lastCheckpoint)
}

// LatestValidCheckpoint finds the latest checkpoint <= targetCheckpoint that has a root.
// A valid checkpoint is considered to be one that has a root retained in the changeset,
// where the root ID's checkpoint corresponds to the actual checkpoint.
// If for example we have checkpoint 100, but its root ID has checkpoint number 99,
// we actually need to navigate to checkpoint 99 to confirm that the root was retained.
// If there is no available checkpoint, CheckpointRootInfo.Version will be zero.
func (cr *ChangesetReader) LatestValidCheckpoint(targetCheckpoint uint32) CheckpointRootInfo {
	if targetCheckpoint < cr.firstCheckpoint || targetCheckpoint > cr.lastCheckpoint {
		return CheckpointRootInfo{}
	}
	i := int(targetCheckpoint - cr.firstCheckpoint)
	for ; i >= 0; i-- {
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
			if rootCheckpoint >= cr.firstCheckpoint {
				sourceInfo = cr.checkpointsInfo.UnsafeItem(rootCheckpoint - cr.firstCheckpoint)
			} else {
				// root's source checkpoint is not in this changeset (compacted away entirely)
				// we likely don't have any valid checkpoint roots in this changeset,
				// but we'll keep looking at earlier checkpoints just in case
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

	// binary search for nearest checkpoint <= targetVersion
	low := 0
	high := count - 1
	resultCheckpoint := uint32(0)

	for low <= high {
		mid := (low + high) / 2
		midInfo := cr.checkpointsInfo.UnsafeItem(uint32(mid))

		if midInfo.Version <= targetVersion {
			// this checkpoint is a candidate, let's see if we can find a higher one, search right
			resultCheckpoint = midInfo.Checkpoint
			low = mid + 1
		} else {
			// this checkpoint is too high, search left
			high = mid - 1
		}
	}

	if resultCheckpoint == 0 {
		// no checkpoint found <= targetVersion
		return CheckpointRootInfo{}
	}

	return cr.LatestValidCheckpoint(resultCheckpoint)
}

func (cr *ChangesetReader) TotalBytes() int {
	return cr.leavesData.TotalBytes() +
		cr.branchesData.TotalBytes() +
		cr.checkpointsInfo.TotalBytes() +
		cr.kvDataReader.Len() +
		cr.walReader.Len()
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
