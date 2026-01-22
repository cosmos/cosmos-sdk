package internal

import (
	"errors"
	"fmt"
)

type ChangesetReader struct {
	changeset *Changeset      // we keep a reference to the parent changeset handle
	files     *ChangesetFiles // files is only non-nil if this reader "owns" the changeset files
	info      *ChangesetInfo  // we reference this so that it can be accessed even in shared changesets

	walReader       *KVDataReader
	kvDataReader    *KVDataReader
	branchesData    *NodeMmap[BranchLayout]
	leavesData      *NodeMmap[LeafLayout]
	checkpointsInfo *StructMmap[CheckpointInfo]
	orphanWriter    *OrphanWriter
}

func NewChangesetReader(changeset *Changeset, files *ChangesetFiles, owned bool) (*ChangesetReader, error) {
	cr := &ChangesetReader{changeset: changeset}

	var err error

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

	cr.orphanWriter = NewOrphanWriter(files.orphansFile)

	cr.info = files.info

	if owned {
		cr.files = files
	}

	return cr, nil
}

//func OpenChangeset(dir string) (*ChangesetReader, error) {
//	files, err := OpenChangesetFiles(dir)
//	if err != nil {
//		return nil, fmt.Errorf("failed to open changeset files: %w", err)
//	}
//	cs := &ChangesetReader{}
//	err = cs.InitOwned(files)
//	if err != nil {
//		return nil, fmt.Errorf("failed to initialize changeset: %w", err)
//	}
//	return cs, nil
//}

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
	info, err := cr.getCheckpointInfo(id.Checkpoint())
	if err != nil {
		return nil, err
	}
	return cr.leavesData.FindByID(id, &info.Leaves)
}

func (cr *ChangesetReader) ResolveBranchByID(id NodeID) (*BranchLayout, error) {
	if id.IsLeaf() {
		return nil, fmt.Errorf("node ID %s is not a branch", id.String())
	}

	info, err := cr.getCheckpointInfo(id.Checkpoint())
	if err != nil {
		return nil, err
	}
	return cr.branchesData.FindByID(id, &info.Branches)
}

func (cr *ChangesetReader) getCheckpointInfo(checkpoint uint32) (*CheckpointInfo, error) {
	info := cr.info
	firstCheckpoint := info.FirstCheckpoint
	lastCheckpoint := firstCheckpoint + uint32(cr.checkpointsInfo.Count()) - 1
	if checkpoint < firstCheckpoint || checkpoint > lastCheckpoint {
		return nil, fmt.Errorf("checkpoint %d out of range for changeset (have %d..%d)", checkpoint, firstCheckpoint, lastCheckpoint)
	}
	return cr.checkpointsInfo.UnsafeItem(checkpoint - firstCheckpoint), nil
}

func (cr *ChangesetReader) Changeset() *Changeset {
	return cr.changeset
}

//var ErrDisposed = errors.New("changeset disposed")

func (cr *ChangesetReader) MarkOrphan(version uint32, nodeId NodeID) error {
	err := cr.orphanWriter.WriteOrphan(version, nodeId)
	if err != nil {
		return fmt.Errorf("failed to write orphan node: %w", err)
	}

	info := cr.info
	if nodeId.IsLeaf() {
		info.LeafOrphans++
		info.LeafOrphanVersionTotal += uint64(version)
	} else {
		info.BranchOrphans++
		info.BranchOrphanVersionTotal += uint64(version)
	}

	return nil
}

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

func (cr *ChangesetReader) Close() error {
	errs := []error{
		cr.kvDataReader.Close(),
		cr.leavesData.Close(),
		cr.branchesData.Close(),
		cr.checkpointsInfo.Close(),
		//cr.orphanWriter.Flush(),
	}
	if cr.files != nil {
		errs = append(errs, cr.files.Close())
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
