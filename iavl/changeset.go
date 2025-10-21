package iavlx

import (
	"errors"
	"fmt"
	"sync/atomic"
	"unsafe"
)

type Changeset struct {
	files *ChangesetFiles

	treeStore *TreeStore

	info         *ChangesetInfo
	infoMmap     *StructMmap[ChangesetInfo]
	kvLog        *KVLog // TODO make sure we handle compaction here too
	branchesData *NodeMmap[BranchLayout]
	leavesData   *NodeMmap[LeafLayout]
	versionsData *StructMmap[VersionInfo]

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

	// we need a reference to the changeset info mmap to be able to flush it later when orphans are marked
	cr.info = files.info
	cr.infoMmap = files.infoMmap

	return nil
}

func (cr *Changeset) getVersionInfo(version uint32) (*VersionInfo, error) {
	if version < cr.info.StartVersion || version >= cr.info.StartVersion+uint32(cr.versionsData.Count()) {
		return nil, fmt.Errorf("version %d out of range for changeset (have %d..%d)", version, cr.info.StartVersion, cr.info.StartVersion+uint32(cr.versionsData.Count())-1)
	}
	return cr.versionsData.UnsafeItem(version - cr.info.StartVersion), nil
}

func (cr *Changeset) ReadK(nodeId NodeID, offset uint32) (key []byte, err error) {
	if cr.evicted.Load() {
		return cr.treeStore.ReadK(nodeId, offset)
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
		return cr.treeStore.ReadKV(nodeId, offset)
	}
	cr.Pin()
	defer cr.Unpin()

	// TODO add an optimization when we only want to read and copy value
	k, v, err := cr.kvLog.ReadKV(offset)
	if err != nil {
		return nil, nil, err
	}
	copyKey := make([]byte, len(k))
	copy(copyKey, k)
	copyValue := make([]byte, len(v))
	copy(copyValue, v)
	return copyKey, copyValue, nil
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

	if nodeId.IsLeaf() {
		layout, err := cr.ResolveLeaf(nodeId, fileIdx)
		if err != nil {
			return nil, err
		}
		return &LeafPersisted{layout: layout, store: cr}, nil
	} else {
		layout, _, err := cr.resolveBranchWithIdx(nodeId, fileIdx)
		if err != nil {
			return nil, err
		}

		return &BranchPersisted{
			layout: layout,
			store:  cr,
		}, nil
	}
}

var ErrDisposed = errors.New("changeset disposed")

func (cr *Changeset) MarkOrphan(version uint32, nodeId NodeID) error {
	if cr.evicted.Load() {
		return ErrDisposed
	}
	cr.Pin()
	defer cr.Unpin()

	nodeVersion := uint32(nodeId.Version())
	vi, err := cr.getVersionInfo(nodeVersion)
	if err != nil {
		return err
	}

	if nodeId.IsLeaf() {
		leaf, err := cr.leavesData.FindByID(nodeId, &vi.Leaves)
		if err != nil {
			return err
		}

		if leaf.OrphanVersion == 0 {
			leaf.OrphanVersion = version
			cr.info.LeafOrphans++
			cr.info.LeafOrphanVersionTotal += uint64(version)
			cr.dirtyLeaves.Store(true)
		}
	} else {
		branch, err := cr.branchesData.FindByID(nodeId, &vi.Branches)
		if err != nil {
			return err
		}

		if branch.OrphanVersion == 0 {
			branch.OrphanVersion = version
			cr.info.BranchOrphans++
			cr.info.BranchOrphanVersionTotal += uint64(version)
			cr.dirtyBranches.Store(true)
		}
	}

	return nil
}

func (cr *Changeset) ReadyToCompact(orphanPercentTarget float64, orphanAgeTarget uint32) bool {
	leafOrphanCount := cr.info.LeafOrphans
	if leafOrphanCount > 0 {
		leafOrphanPercent := float64(leafOrphanCount) / float64(cr.leavesData.Count())
		leafOrphanAge := uint32(cr.info.LeafOrphanVersionTotal / uint64(cr.info.LeafOrphans))

		if leafOrphanPercent >= orphanPercentTarget && leafOrphanAge <= orphanAgeTarget {
			return true
		}
	}

	branchOrphanCount := cr.info.BranchOrphans
	if branchOrphanCount > 0 {
		branchOrphanPercent := float64(branchOrphanCount) / float64(cr.branchesData.Count())
		branchOrphanAge := uint32(cr.info.BranchOrphanVersionTotal / uint64(cr.info.BranchOrphans))
		if branchOrphanPercent >= orphanPercentTarget && branchOrphanAge <= orphanAgeTarget {
			return true
		}
	}

	return false
}

func (cr *Changeset) FlushOrphans() error {
	cr.Pin()
	defer cr.Unpin()

	wasDirty := false
	if cr.dirtyLeaves.Load() {
		wasDirty = true
		err := cr.leavesData.Flush()
		if err != nil {
			return fmt.Errorf("failed to flush leaf data: %w", err)
		}
		cr.dirtyLeaves.Store(false)
	}
	if cr.dirtyBranches.Load() {
		wasDirty = true
		err := cr.branchesData.Flush()
		if err != nil {
			return fmt.Errorf("failed to flush branch data: %w", err)
		}
		cr.dirtyBranches.Store(false)
	}
	if wasDirty {
		err := cr.infoMmap.Flush()
		if err != nil {
			return fmt.Errorf("failed to flush changeset info: %w", err)
		}
	}
	return nil
}

func (cr *Changeset) Close() error {
	errs := []error{
		cr.kvLog.Close(),
		cr.leavesData.Close(),
		cr.branchesData.Close(),
		cr.versionsData.Close(),
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
			cr.info = nil
			cr.infoMmap = nil
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
	return cr.info.LeafOrphans > 0 || cr.info.BranchOrphans > 0
}
