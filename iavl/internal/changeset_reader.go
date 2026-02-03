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
	if files.WALFile() != nil {
		cr.walReader, err = NewKVDataReader(files.WALFile())
		if err != nil {
			return nil, fmt.Errorf("failed to open WAL data store: %w", err)
		}
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
		firstInfo, err := cr.checkpointsInfo.UnsafeItem(0), error(nil)
		if err != nil {
			return nil, fmt.Errorf("failed to read first checkpoint info: %w", err)
		}
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

func (cr *ChangesetReader) Changeset() *Changeset {
	return cr.changeset
}

func (cr *ChangesetReader) Close() error {
	errs := []error{
		cr.kvDataReader.Close(),
		cr.leavesData.Close(),
		cr.branchesData.Close(),
		cr.checkpointsInfo.Close(),
	}
	return errors.Join(errs...)
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

// LatestCheckpointRoot returns the latest checkpoint root NodePointer and its version.
// If there are no checkpoints with roots, (nil, 0) is returned.
// If the latest checkpoint has an empty tree, (nil, version) is returned.
func (cr *ChangesetReader) LatestCheckpointRoot() (*NodePointer, uint32) {
	count := cr.checkpointsInfo.Count()
	if count == 0 {
		return nil, 0
	}
	return cr.LatestValidCheckpoint(cr.lastCheckpoint)
}

// LatestValidCheckpoint finds the latest checkpoint <= targetCheckpoint that has a root.
// If found, it returns the root NodePointer and the checkpoint version.
// If no such checkpoint exists, (nil, 0) is returned.
// If the checkpoint has an empty tree, (nil, version) is returned.
func (cr *ChangesetReader) LatestValidCheckpoint(targetCheckpoint uint32) (*NodePointer, uint32) {
	if targetCheckpoint < cr.firstCheckpoint || targetCheckpoint > cr.lastCheckpoint {
		return nil, 0
	}
	i := int(targetCheckpoint - cr.firstCheckpoint)
	for ; i >= 0; i-- {
		info := cr.checkpointsInfo.UnsafeItem(uint32(i))
		rootID := info.RootID
		if rootID.IsEmpty() {
			// if root is empty, we have an empty tree at this checkpoint
			return nil, info.Version
		}
		if rootID.Checkpoint() != info.Checkpoint {
			// if root ID checkpoint does not match, skip because this root is actually in a different checkpoint - we have a checkpoint with no changes
			continue
		}
		// check if the root was retained in this changeset, if it was compacted, maybe not
		var nodeSetInfo *NodeSetInfo
		if rootID.IsLeaf() {
			nodeSetInfo = &info.Leaves
		} else {
			nodeSetInfo = &info.Branches
		}
		idx := rootID.Index()
		if idx < nodeSetInfo.StartIndex || idx > nodeSetInfo.EndIndex {
			// root node was compacted away
			continue
		}
		return &NodePointer{
			id:        info.RootID,
			changeset: cr.changeset,
		}, info.Version
	}
	return nil, 0
}

// CheckpointForVersion finds the nearest checkpoint <= targetVersion that has a root.
// If found, it returns the root NodePointer and the checkpoint version.
// If no such checkpoint exists, (nil, 0) is returned.
// If the checkpoint has an empty tree, (nil, version) is returned.
func (cr *ChangesetReader) CheckpointForVersion(targetVersion uint32) (cpRoot *NodePointer, checkpointVersion uint32) {
	count := cr.checkpointsInfo.Count()
	if count == 0 {
		return nil, 0
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
		return nil, 0
	}

	return cr.LatestValidCheckpoint(resultCheckpoint)
}

//func (cr *Changeset) TotalBytes() int {
//	return cr.leavesData.TotalBytes() +
//		cr.branchesData.TotalBytes() +
//		cr.kvLog.TotalBytes() +
//		cr.versionsData.TotalBytes()
//}
//
//func (cr *Changeset) HasOrphans() bool {
//	info := cr.info
//	return info.LeafOrphans > 0 || info.BranchOrphans > 0
//}
//
//func (cr *Changeset) ResolveRoot(version uint32) (*NodePointer, error) {
//	startVersion := cr.info.StartVersion
//	endVersion := startVersion + uint32(cr.versionsData.Count()) - 1
//	if version < startVersion || version > endVersion {
//		return nil, fmt.Errorf("version %d out of range for changeset (have %d..%d)", version, startVersion, endVersion)
//	}
//	vi, err := cr.getVersionInfo(version)
//	if err != nil {
//		return nil, err
//	}
//	if vi.RootID == 0 {
//		// empty tree
//		return nil, nil
//	}
//	return &NodePointer{
//		id:    vi.RootID,
//		store: cr,
//	}, nil
//}

//var ErrDisposed = errors.New("changeset disposed")

//
//func (cr *Changeset) ReadyToCompact(orphanPercentTarget float64, orphanAgeTarget uint32) bool {
//	info := cr.info
//	leafOrphanCount := info.LeafOrphans
//	if leafOrphanCount > 0 {
//		leafOrphanPercent := float64(leafOrphanCount) / float64(cr.leavesData.Count())
//		leafOrphanAge := uint32(info.LeafOrphanVersionTotal / uint64(info.LeafOrphans))
//
//		if leafOrphanPercent >= orphanPercentTarget && leafOrphanAge <= orphanAgeTarget {
//			return true
//		}
//	}
//
//	branchOrphanCount := info.BranchOrphans
//	if branchOrphanCount > 0 {
//		branchOrphanPercent := float64(branchOrphanCount) / float64(cr.branchesData.Count())
//		branchOrphanAge := uint32(info.BranchOrphanVersionTotal / uint64(info.BranchOrphans))
//		if branchOrphanPercent >= orphanPercentTarget && branchOrphanAge <= orphanAgeTarget {
//			return true
//		}
//	}
//
//	return false
//}
//
