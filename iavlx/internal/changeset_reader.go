package internal

import (
	"errors"
	"fmt"
)

// ChangesetReader provides read access to a changeset's on-disk data via memory-mapped files.
//
// A changeset contains the persisted state for a range of tree versions. The reader mmaps all
// of the changeset's data files and provides methods to:
//   - Resolve nodes by ID or by file index (for reading tree structure from disk)
//   - Look up checkpoint metadata (for finding roots at specific versions)
//   - Read key/value blobs from either the WAL or kv.dat
//   - Verify checkpoint integrity (for crash recovery)
//
// Node resolution: there are two ways to find a node in the data files:
//   - By file index (ResolveLeafByIndex/ResolveBranchByIndex): O(1) lookup using the 1-based
//     offset stored in a parent branch's LeftOffset/RightOffset. This is the fast path used
//     when traversing the tree from a parent that was written in the same changeset.
//   - By NodeID (ResolveLeafByID/ResolveBranchByID): looks up the checkpoint's NodeSetInfo
//     to find the offset range, then searches within that range. Used when the parent was
//     written in a different changeset or when navigating from a checkpoint root.
//
// Checkpoint lookup: checkpoints may be contiguous (checkpoint N, N+1, N+2...) or have gaps
// (after compaction removes some). When contiguous, lookup is O(1) by index arithmetic.
// When non-contiguous, we fall back to binary search.
type ChangesetReader struct {
	changeset *Changeset // we keep a reference to the parent changeset handle

	// walReader and kvDataReader provide access to key/value blob data.
	// A node's key or value may be in either file — the node's layout flags indicate which.
	walReader    *KVDataReader
	kvDataReader *KVDataReader
	// branchesData and leavesData are mmapped arrays of fixed-size node records.
	// Nodes are addressed by 1-based file index (for O(1) parent→child navigation)
	// or by NodeID (which requires a checkpoint lookup first).
	branchesData *NodeMmap[BranchLayout]
	leavesData   *NodeMmap[LeafLayout]
	// Checkpoint bookkeeping: the range of checkpoint numbers in this changeset.
	firstCheckpoint       uint32
	lastCheckpoint        uint32
	checkpointsContiguous bool // true if checkpoint numbers are contiguous (no gaps from compaction)
	// checkpointsInfo is the mmapped array of CheckpointInfo entries — one per checkpoint.
	// Each entry records the version, root NodeID, node count/offset ranges, and CRC32.
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
		count := cr.checkpointsInfo.Count()
		firstInfo := cr.checkpointsInfo.UnsafeItem(0)
		cr.firstCheckpoint = firstInfo.Checkpoint
		lastInfo := cr.checkpointsInfo.UnsafeItem(count - 1)
		cr.lastCheckpoint = lastInfo.Checkpoint
		cr.checkpointsContiguous = int(cr.lastCheckpoint-cr.firstCheckpoint+1) == count
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

// ResolveLeafByIndex returns a leaf node by its 1-based file offset.
// This is the fast O(1) path used when a parent branch has a LeftOffset/RightOffset
// pointing to a child in the same changeset.
func (cr *ChangesetReader) ResolveLeafByIndex(fileIdx uint32) (*LeafLayout, error) {
	if fileIdx == 0 || fileIdx > uint32(cr.leavesData.Count()) {
		return nil, fmt.Errorf("leaf file index %d out of range (have 1..%d)", fileIdx, cr.leavesData.Count())
	}

	fileIdx-- // convert to 0-based index
	return cr.leavesData.UnsafeItem(int(fileIdx)), nil
}

func (cr *ChangesetReader) ResolveBranchByIndex(fileIdx uint32) (*BranchLayout, error) {
	if fileIdx == 0 || fileIdx > uint32(cr.branchesData.Count()) {
		return nil, fmt.Errorf("branch file index %d out of range (have 1..%d)", fileIdx, cr.branchesData.Count())
	}

	fileIdx-- // convert to 0-based index
	return cr.branchesData.UnsafeItem(int(fileIdx)), nil
}

// ResolveLeafByID finds a leaf node by its NodeID. This requires looking up the checkpoint's
// NodeSetInfo to determine the offset range where this leaf could be, then searching within it.
// Slower than ResolveLeafByIndex but works across changeset boundaries.
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

// GetCheckpointInfo returns the metadata for a specific checkpoint number.
// Uses O(1) index arithmetic when checkpoints are contiguous, falls back to binary search otherwise.
func (cr *ChangesetReader) GetCheckpointInfo(checkpoint uint32) (*CheckpointInfo, error) {
	if checkpoint < cr.firstCheckpoint || checkpoint > cr.lastCheckpoint {
		return nil, fmt.Errorf("checkpoint %d out of range for changeset %s (have %d..%d)",
			checkpoint, cr.changeset.Files().Dir(), cr.firstCheckpoint, cr.lastCheckpoint)
	}
	if cr.checkpointsContiguous {
		item := cr.checkpointsInfo.UnsafeItem(int(checkpoint - cr.firstCheckpoint))
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
		item := cr.checkpointsInfo.UnsafeItem(idx)
		if item.Checkpoint == checkpoint {
			return item, nil
		}
	}
	return nil, fmt.Errorf("checkpoint %d not found in changeset %s (have %d..%d, likely removed during compaction)",
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

// latestValidCheckpoint walks backwards from startIdx to find the latest checkpoint whose
// root node is actually present in this changeset's data files.
//
// This is more complex than it might seem because of three subtleties:
//
//  1. Empty tree sentinel: a checkpoint can record that the tree was empty at that version.
//     This is a valid result (Root=nil, Version set) — it means "the tree existed but had no keys".
//
//  2. Zero-value RootID (IsEmpty): means the checkpoint entry has no root information at all.
//     This is distinct from IsEmptyTree (which means "the tree has zero keys"). No current code
//     path produces a zero-value RootID — SaveCheckpoint always sets either a real root or the
//     empty-tree sentinel, and compaction copies RootID through. This check is defensive; if we
//     encounter it, we skip to an earlier checkpoint.
//
//  3. Root from a different checkpoint: when no keys change between versions, the new checkpoint
//     re-uses the same root node from an earlier checkpoint. For example, checkpoint 100 might
//     have a RootID whose checkpoint field is 99 — meaning "the root is the same node that was
//     written in checkpoint 99". To verify the root still exists, we need to check the NODE SET
//     bounds of checkpoint 99 (not 100), because that's where the node lives on disk.
//     If checkpoint 99 was compacted away, the root is gone and we must keep looking backwards.
//
// startIdx is a 0-based index into the checkpoints info array.
// If no valid checkpoint is found, CheckpointRootInfo.Version will be zero.
func (cr *ChangesetReader) latestValidCheckpoint(startIdx int) CheckpointRootInfo {
	for i := startIdx; i >= 0; i-- {
		info := cr.checkpointsInfo.UnsafeItem(i)
		rootID := info.RootID

		// Case 1: the tree was explicitly empty at this checkpoint — valid result, no root node.
		if rootID.IsEmptyTree() {
			return CheckpointRootInfo{
				Version:    info.Version,
				Checkpoint: info.Checkpoint,
			}
		}

		// Case 2: zero-value RootID — no root information available (defensive; shouldn't happen
		// in practice since SaveCheckpoint always sets a root or empty-tree sentinel).
		if rootID.IsEmpty() {
			continue
		}

		// Case 3: we have a RootID — but we need to verify the actual node still exists in this
		// changeset's data files. The root may have been written in a DIFFERENT checkpoint than
		// this one (if no tree changes happened at this version), so we look up the source
		// checkpoint to check its node set bounds.
		rootCheckpoint := rootID.Checkpoint()
		sourceInfo := info
		if rootCheckpoint != info.Checkpoint {
			// Root points to a node created in an earlier checkpoint.
			// Look up that checkpoint's metadata to get the node offset ranges.
			var err error
			sourceInfo, err = cr.GetCheckpointInfo(rootCheckpoint)
			if err != nil {
				// The source checkpoint was compacted away — this root is no longer resolvable.
				continue
			}
		}

		// Check whether the root node's index falls within the node set bounds for its checkpoint.
		// After compaction, some nodes may have been pruned, shrinking the index range.
		var nodeSetInfo *NodeSetInfo
		if rootID.IsLeaf() {
			nodeSetInfo = &sourceInfo.Leaves
		} else {
			nodeSetInfo = &sourceInfo.Branches
		}
		idx := rootID.Index()
		if idx < nodeSetInfo.StartIndex || idx > nodeSetInfo.EndIndex {
			// Root node was pruned during compaction — keep looking backwards.
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

// Verify checks the integrity of the latest checkpoint in this changeset.
//
// It validates two things:
//  1. CRC32 integrity: the checkpoint metadata entry has a CRC32 checksum covering its fields.
//     If this doesn't match, the checkpoint entry itself was partially written (crash mid-write).
//  2. File size consistency: the checkpoint metadata records how many branches, leaves, and bytes
//     of KV data should exist. If the actual files are smaller, checkpoint writing was interrupted
//     before all data was flushed.
//
// Only the LAST checkpoint is verified because that's the only one that could be incomplete —
// earlier checkpoints were fully written before any subsequent commit started.
//
// If verification fails, the caller (VerifyAndFix) can repair by rolling back the last checkpoint,
// which truncates the data files back to the previous checkpoint's offsets.
func (cr *ChangesetReader) Verify() error {
	n := cr.checkpointsInfo.Count()
	if n == 0 {
		// no checkpoints, nothing to verify
		return nil
	}

	lastInfo := cr.checkpointsInfo.UnsafeItem(n - 1)
	if !lastInfo.VerifyCRC32() {
		return fmt.Errorf("changeset checkpoint info failed CRC32 check during verification: checkpoint %d, expected CRC32 %08x, actual CRC32 %08x",
			lastInfo.Checkpoint, lastInfo.CRC32, lastInfo.ComputeCRC32())
	}
	// Check that the data files are at least as large as the checkpoint claims.
	// A crash during checkpoint writing could leave these files truncated.
	expectedLeafCount := lastInfo.Leaves.StartOffset + lastInfo.Leaves.Count
	if expectedLeafCount > uint32(cr.leavesData.Count()) {
		return fmt.Errorf("changeset leaves data count mismatch during verification: expected %d, actual %d", expectedLeafCount, cr.leavesData.Count())
	}
	expectedBranchCount := lastInfo.Branches.StartOffset + lastInfo.Branches.Count
	if expectedBranchCount > uint32(cr.branchesData.Count()) {
		return fmt.Errorf("changeset branches data count mismatch during verification: expected %d, actual %d", expectedBranchCount, cr.branchesData.Count())
	}
	expectedKVDataSize := lastInfo.KVEndOffset
	if expectedKVDataSize > uint64(cr.kvDataReader.Len()) {
		return fmt.Errorf("changeset KV data size mismatch during verification: expected at least %d bytes, actual %d bytes", expectedKVDataSize, cr.kvDataReader.Len())
	}
	return nil
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
