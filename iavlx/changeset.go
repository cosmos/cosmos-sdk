package iavlx

import (
	"errors"
	"fmt"
	"sync/atomic"
	"unsafe"
)

type Changeset struct {
	files *ChangesetFiles
	info  *ChangesetInfo // we copy this so that it can be accessed even in shared changesets

	treeStore *TreeStore

	kvLog        *KVLog // TODO make sure we handle compaction here too
	branchesData *NodeMmap[BranchLayout]
	leavesData   *NodeMmap[LeafLayout]
	versionsData *StructMmap[VersionInfo]
	orphanWriter *OrphanWriter

	refCount      atomic.Int32
	evicted       atomic.Bool
	disposed      atomic.Bool
	dirtyBranches atomic.Bool
	dirtyLeaves   atomic.Bool
	needsSync     atomic.Bool
}

func NewChangeset(treeStore *TreeStore) *Changeset {
	return &Changeset{
		treeStore: treeStore,
	}
}

func OpenChangeset(treeStore *TreeStore, dir string) (*Changeset, error) {
	files, err := OpenChangesetFiles(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to open changeset files: %w", err)
	}
	cs := NewChangeset(treeStore)
	err = cs.InitOwned(files)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize changeset: %w", err)
	}
	return cs, nil
}

func (cr *Changeset) InitOwned(files *ChangesetFiles) error {
	err := cr.InitShared(files)
	if err != nil {
		return err
	}
	cr.files = files
	return nil
}

func (cr *Changeset) InitShared(files *ChangesetFiles) error {
	var err error

	cr.kvLog, err = NewKVLog(files.kvlogFile)
	if err != nil {
		return fmt.Errorf("failed to open KV data store: %w", err)
	}

	cr.leavesData, err = NewNodeReader[LeafLayout](files.leavesFile)
	if err != nil {
		return fmt.Errorf("failed to open leaves data file: %w", err)
	}

	cr.branchesData, err = NewNodeReader[BranchLayout](files.branchesFile)
	if err != nil {
		return fmt.Errorf("failed to open branches data file: %w", err)
	}

	cr.versionsData, err = NewStructReader[VersionInfo](files.versionsFile)
	if err != nil {
		return fmt.Errorf("failed to open versions data file: %w", err)
	}

	cr.orphanWriter = NewOrphanWriter(files.orphansFile)

	cr.info = files.info

	return nil
}

func (cr *Changeset) getVersionInfo(version uint32) (*VersionInfo, error) {
	info := cr.info
	if version < info.StartVersion || version >= info.StartVersion+uint32(cr.versionsData.Count()) {
		return nil, fmt.Errorf("version %d out of range for changeset (have %d..%d)", version, info.StartVersion, info.StartVersion+uint32(cr.versionsData.Count())-1)
	}
	return cr.versionsData.UnsafeItem(version - info.StartVersion), nil
}

func (cr *Changeset) ReadK(nodeId NodeID, offset uint32) (key []byte, err error) {
	if cr.evicted.Load() {
		return cr.treeStore.ReadK(nodeId)
	}
	cr.Pin()
	defer cr.Unpin()

	k, err := cr.kvLog.UnsafeReadK(offset)
	if err != nil {
		return nil, err
	}
	copyKey := make([]byte, len(k))
	copy(copyKey, k)
	return copyKey, nil
}

func (cr *Changeset) ReadKV(nodeId NodeID, offset uint32) (key, value []byte, err error) {
	if cr.evicted.Load() {
		return cr.treeStore.ReadKV(nodeId)
	}
	cr.Pin()
	defer cr.Unpin()

	k, v, err := cr.kvLog.UnsafeReadKV(offset)
	if err != nil {
		return nil, nil, err
	}
	copyKey := make([]byte, len(k))
	copy(copyKey, k)
	copyValue := make([]byte, len(v))
	copy(copyValue, v)
	return copyKey, copyValue, nil
}

func (cr *Changeset) ReadV(nodeId NodeID, offset uint32) (value []byte, err error) {
	if cr.evicted.Load() {
		return cr.treeStore.ReadV(nodeId)
	}
	cr.Pin()
	defer cr.Unpin()

	_, v, err := cr.kvLog.UnsafeReadKV(offset)
	if err != nil {
		return nil, err
	}
	copyValue := make([]byte, len(v))
	copy(copyValue, v)
	return copyValue, nil
}

func (cr *Changeset) ResolveLeaf(nodeId NodeID, fileIdx uint32) (LeafLayout, error) {
	if cr.evicted.Load() {
		return cr.treeStore.ResolveLeaf(nodeId)
	}
	cr.Pin()
	defer cr.Unpin()

	if fileIdx == 0 {
		version := uint32(nodeId.Version())
		vi, err := cr.getVersionInfo(version)
		if err != nil {
			return LeafLayout{}, err
		}
		leaf, err := cr.leavesData.FindByID(nodeId, &vi.Leaves)
		if err != nil {
			return LeafLayout{}, err
		}
		return *leaf, nil
	} else {
		fileIdx-- // convert to 0-based index
		return *cr.leavesData.UnsafeItem(fileIdx), nil
	}
}

func (cr *Changeset) ResolveBranch(nodeId NodeID, fileIdx uint32) (BranchLayout, error) {
	if cr.evicted.Load() {
		return cr.treeStore.ResolveBranch(nodeId)
	}

	layout, _, err := cr.resolveBranchWithIdx(nodeId, fileIdx)
	return layout, err
}

func (cr *Changeset) resolveBranchWithIdx(nodeId NodeID, fileIdx uint32) (BranchLayout, uint32, error) {
	cr.Pin()
	defer cr.Unpin()

	if fileIdx == 0 {
		version := uint32(nodeId.Version())
		vi, err := cr.getVersionInfo(version)
		if err != nil {
			return BranchLayout{}, 0, err
		}
		branch, err := cr.branchesData.FindByID(nodeId, &vi.Branches)
		if err != nil {
			return BranchLayout{}, 0, err
		}
		// Compute the actual file index from the pointer
		itemIdx := uint32((uintptr(unsafe.Pointer(branch)) - uintptr(unsafe.Pointer(&cr.branchesData.items[0]))) / uintptr(cr.branchesData.size))
		return *branch, itemIdx + 1, nil // +1 to convert back to 1-based
	} else {
		itemIdx := fileIdx - 1                                    // convert to 0-based index
		return *cr.branchesData.UnsafeItem(itemIdx), fileIdx, nil // return original fileIdx
	}
}

func (cr *Changeset) Resolve(nodeId NodeID, fileIdx uint32) (Node, error) {
	if cr.evicted.Load() {
		return cr.treeStore.Resolve(nodeId, fileIdx)
	}
	cr.Pin()
	defer cr.Unpin()

	// we don't have a fileIdx, so its probably not in this changeset.
	if fileIdx == 0 {
		// load up the changeset for this node
		cs := cr.treeStore.getChangesetForVersion(uint32(nodeId.Version()))
		cs.Pin()
		defer cs.Unpin()

		// get version data
		version := uint32(nodeId.Version())
		vi, err := cs.getVersionInfo(version)
		if err != nil {
			return nil, err
		}
		if nodeId.IsLeaf() {
			leaf, err := cs.leavesData.FindByID(nodeId, &vi.Leaves)
			if err != nil {
				return nil, err
			}
			return &LeafPersisted{
				store:   cs,
				selfIdx: 0,
				layout:  *leaf,
			}, nil
		} else {
			branch, err := cs.branchesData.FindByID(nodeId, &vi.Branches)
			if err != nil {
				return nil, err
			}
			return &BranchPersisted{
				store:  cs,
				layout: *branch,
			}, nil
		}
	} else {
		// since we have the fileIdx, we know it's in this changeset.
		// we can just directly index in this changeset's leaf/branch data.
		if nodeId.IsLeaf() {
			itemIdx := fileIdx - 1
			leafLayout := *cr.leavesData.UnsafeItem(itemIdx)
			return &LeafPersisted{layout: leafLayout, store: cr}, nil
		} else {
			itemIdx := fileIdx - 1
			branchLayout := *cr.branchesData.UnsafeItem(itemIdx)
			return &BranchPersisted{
				layout: branchLayout,
				store:  cr,
			}, nil
		}
	}
}

var ErrDisposed = errors.New("changeset disposed")

func (cr *Changeset) MarkOrphan(version uint32, nodeId NodeID) error {
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

func (cr *Changeset) ReadyToCompact(orphanPercentTarget float64, orphanAgeTarget uint32) bool {
	info := cr.info
	leafOrphanCount := info.LeafOrphans
	if leafOrphanCount > 0 {
		leafOrphanPercent := float64(leafOrphanCount) / float64(cr.leavesData.Count())
		leafOrphanAge := uint32(info.LeafOrphanVersionTotal / uint64(info.LeafOrphans))

		if leafOrphanPercent >= orphanPercentTarget && leafOrphanAge <= orphanAgeTarget {
			return true
		}
	}

	branchOrphanCount := info.BranchOrphans
	if branchOrphanCount > 0 {
		branchOrphanPercent := float64(branchOrphanCount) / float64(cr.branchesData.Count())
		branchOrphanAge := uint32(info.BranchOrphanVersionTotal / uint64(info.BranchOrphans))
		if branchOrphanPercent >= orphanPercentTarget && branchOrphanAge <= orphanAgeTarget {
			return true
		}
	}

	return false
}

func (cr *Changeset) Close() error {
	errs := []error{
		cr.kvLog.Close(),
		cr.leavesData.Close(),
		cr.branchesData.Close(),
		cr.versionsData.Close(),
		cr.orphanWriter.Flush(),
	}
	if cr.files != nil {
		errs = append(errs, cr.files.Close())
	}
	return errors.Join(errs...)
}

func (cr *Changeset) Pin() {
	cr.refCount.Add(1)
}

func (cr *Changeset) Unpin() {
	cr.refCount.Add(-1)
}

func (cr *Changeset) Evict() {
	cr.evicted.Store(true)
}

func (cr *Changeset) TryDispose() bool {
	if cr.disposed.Load() {
		return true
	}
	if cr.refCount.Load() <= 0 {
		if cr.disposed.CompareAndSwap(false, true) {
			_ = cr.Close()
			cr.versionsData = nil
			cr.branchesData = nil
			cr.leavesData = nil
			cr.kvLog = nil
			// DO NOT set treeStore to nil, as deposed changesets should still forward calls to the main tree store
			// DO NOT set files to nil, as we might need to delete them later
			return true
		}
	}
	return false
}

func (cr *Changeset) TotalBytes() int {
	return cr.leavesData.TotalBytes() +
		cr.branchesData.TotalBytes() +
		cr.kvLog.TotalBytes() +
		cr.versionsData.TotalBytes()
}

func (cr *Changeset) HasOrphans() bool {
	info := cr.info
	return info.LeafOrphans > 0 || info.BranchOrphans > 0
}

func (cr *Changeset) ResolveRoot(version uint32) (*NodePointer, error) {
	startVersion := cr.info.StartVersion
	endVersion := startVersion + uint32(cr.versionsData.Count()) - 1
	if version < startVersion || version > endVersion {
		return nil, fmt.Errorf("version %d out of range for changeset (have %d..%d)", version, startVersion, endVersion)
	}
	vi, err := cr.getVersionInfo(version)
	if err != nil {
		return nil, err
	}
	if vi.RootID.IsEmpty() {
		// empty tree
		return nil, nil
	}
	return &NodePointer{
		id:    vi.RootID,
		store: cr,
	}, nil
}
