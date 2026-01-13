package internal

import (
	"fmt"
)

type ChangesetReader struct {
	changeset *Changeset      // we keep a reference to the parent changeset handle
	files     *ChangesetFiles // files is only non-nil if this reader "owns" the changeset files
	info      *ChangesetInfo  // we reference this so that it can be accessed even in shared changesets

	kvDataReader *KVDataReader
	branchesData *NodeMmap[BranchLayout]
	leavesData   *NodeMmap[LeafLayout]
	layersInfo   *StructMmap[LayerInfo]
	//orphanWriter *OrphanWriter
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

func (cr *ChangesetReader) InitOwned(files *ChangesetFiles) error {
	err := cr.InitShared(files)
	if err != nil {
		return err
	}
	cr.files = files
	return nil
}

func (cr *ChangesetReader) InitShared(files *ChangesetFiles) error {
	var err error

	cr.kvDataReader, err = NewKVDataReader(files.KVDataFile())
	if err != nil {
		return fmt.Errorf("failed to open KV data store: %w", err)
	}

	cr.leavesData, err = NewNodeReader[LeafLayout](files.LeavesFile())
	if err != nil {
		return fmt.Errorf("failed to open leaves data file: %w", err)
	}

	cr.branchesData, err = NewNodeReader[BranchLayout](files.BranchesFile())
	if err != nil {
		return fmt.Errorf("failed to open branches data file: %w", err)
	}

	cr.layersInfo, err = NewStructMmap[LayerInfo](files.LayersFile())
	if err != nil {
		return fmt.Errorf("failed to open versions data file: %w", err)
	}

	//cr.orphanWriter = NewOrphanWriter(files.orphansFile)
	//
	cr.info = files.info

	return nil
}

//func (cr *ChangesetReader) getVersionInfo(version uint32) (*VersionInfo, error) {
//	info := cr.info
//	if version < info.StartVersion || version >= info.StartVersion+uint32(cr.versionsData.Count()) {
//		return nil, fmt.Errorf("version %d out of range for changeset (have %d..%d)", version, info.StartVersion, info.StartVersion+uint32(cr.versionsData.Count())-1)
//	}
//	return cr.versionsData.Item(version - info.StartVersion), nil
//}

func (cr *ChangesetReader) KVData() *KVDataReader {
	return cr.kvDataReader
}

//func (cr *ChangesetReader) ResolveLeafByIndex(fileIdx uint32) (*LeafLayout, error) {
//	if fileIdx == 0 || fileIdx > uint32(cr.leavesData.Count()) {
//		return nil, fmt.Errorf("leaf file index %d out of range (have 1..%d)", fileIdx, cr.leavesData.Count())
//	}
//
//	fileIdx-- // convert to 0-based index
//	return cr.leavesData.Item(fileIdx), nil
//}
//
//func (cr *ChangesetReader) ResolveBranchByIndex(fileIdx uint32) (*BranchLayout, error) {
//	if fileIdx == 0 || fileIdx > uint32(cr.branchesData.Count()) {
//		return nil, fmt.Errorf("branch file index %d out of range (have 1..%d)", fileIdx, cr.branchesData.Count())
//	}
//
//	fileIdx-- // convert to 0-based index
//	return cr.branchesData.Item(fileIdx), nil
//}
//
//func (cr *ChangesetReader) ResolveLeafByID(id NodeID) (*LeafLayout, error) {
//	if !id.IsLeaf() {
//		return nil, fmt.Errorf("node ID %s is not a leaf", id.String())
//	}
//	info, err := cr.getVersionInfo(id.Version)
//	if err != nil {
//		return nil, err
//	}
//	return cr.leavesData.FindByID(id, &info.Leaves)
//}
//
//func (cr *ChangesetReader) ResolveBranchByID(id NodeID) (*BranchLayout, error) {
//	if id.IsLeaf() {
//		return nil, fmt.Errorf("node ID %s is not a branch", id.String())
//	}
//
//	info, err := cr.getVersionInfo(id.Version)
//	if err != nil {
//		return nil, err
//	}
//	return cr.branchesData.FindByID(id, &info.Branches)
//}
//
//func (cr *ChangesetReader) Changeset() *Changeset {
//	return cr.changeset
//}
//
////var ErrDisposed = errors.New("changeset disposed")
////
////func (cr *Changeset) MarkOrphan(version uint32, nodeId NodeID) error {
////	err := cr.orphanWriter.WriteOrphan(version, nodeId)
////	if err != nil {
////		return fmt.Errorf("failed to write orphan node: %w", err)
////	}
////
////	info := cr.info
////	if nodeId.IsLeaf() {
////		info.LeafOrphans++
////		info.LeafOrphanVersionTotal += uint64(version)
////	} else {
////		info.BranchOrphans++
////		info.BranchOrphanVersionTotal += uint64(version)
////	}
////
////	return nil
////}
////
////func (cr *Changeset) ReadyToCompact(orphanPercentTarget float64, orphanAgeTarget uint32) bool {
////	info := cr.info
////	leafOrphanCount := info.LeafOrphans
////	if leafOrphanCount > 0 {
////		leafOrphanPercent := float64(leafOrphanCount) / float64(cr.leavesData.Count())
////		leafOrphanAge := uint32(info.LeafOrphanVersionTotal / uint64(info.LeafOrphans))
////
////		if leafOrphanPercent >= orphanPercentTarget && leafOrphanAge <= orphanAgeTarget {
////			return true
////		}
////	}
////
////	branchOrphanCount := info.BranchOrphans
////	if branchOrphanCount > 0 {
////		branchOrphanPercent := float64(branchOrphanCount) / float64(cr.branchesData.Count())
////		branchOrphanAge := uint32(info.BranchOrphanVersionTotal / uint64(info.BranchOrphans))
////		if branchOrphanPercent >= orphanPercentTarget && branchOrphanAge <= orphanAgeTarget {
////			return true
////		}
////	}
////
////	return false
////}
////
//
//func (cr *ChangesetReader) Close() error {
//	errs := []error{
//		cr.kvDataReader.Close(),
//		cr.leavesData.Close(),
//		cr.branchesData.Close(),
//		cr.versionsData.Close(),
//		//cr.orphanWriter.Flush(),
//	}
//	if cr.files != nil {
//		errs = append(errs, cr.files.Close())
//	}
//	return errors.Join(errs...)
//}
//
////
////func (cr *Changeset) Pin() {
////	cr.refCount.Add(1)
////}
////
////func (cr *Changeset) Unpin() {
////	cr.refCount.Add(-1)
////}
////
////func (cr *Changeset) Evict() {
////	cr.evicted.Store(true)
////}
////
////func (cr *Changeset) TryDispose() bool {
////	if cr.disposed.Load() {
////		return true
////	}
////	if cr.refCount.Load() <= 0 {
////		if cr.disposed.CompareAndSwap(false, true) {
////			_ = cr.Close()
////			cr.versionsData = nil
////			cr.branchesData = nil
////			cr.leavesData = nil
////			cr.kvLog = nil
////			// DO NOT set treeStore to nil, as deposed changesets should still forward calls to the main tree store
////			// DO NOT set files to nil, as we might need to delete them later
////			return true
////		}
////	}
////	return false
////}
////
////func (cr *Changeset) TotalBytes() int {
////	return cr.leavesData.TotalBytes() +
////		cr.branchesData.TotalBytes() +
////		cr.kvLog.TotalBytes() +
////		cr.versionsData.TotalBytes()
////}
////
////func (cr *Changeset) HasOrphans() bool {
////	info := cr.info
////	return info.LeafOrphans > 0 || info.BranchOrphans > 0
////}
////
////func (cr *Changeset) ResolveRoot(version uint32) (*NodePointer, error) {
////	startVersion := cr.info.StartVersion
////	endVersion := startVersion + uint32(cr.versionsData.Count()) - 1
////	if version < startVersion || version > endVersion {
////		return nil, fmt.Errorf("version %d out of range for changeset (have %d..%d)", version, startVersion, endVersion)
////	}
////	vi, err := cr.getVersionInfo(version)
////	if err != nil {
////		return nil, err
////	}
////	if vi.RootID == 0 {
////		// empty tree
////		return nil, nil
////	}
////	return &NodePointer{
////		id:    vi.RootID,
////		store: cr,
////	}, nil
////}
