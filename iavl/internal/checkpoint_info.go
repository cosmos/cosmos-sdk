package internal

import (
	"fmt"
	"hash/crc32"
	"unsafe"
)

func init() {
	if unsafe.Sizeof(CheckpointInfo{}) != CheckpointInfoSize {
		panic(fmt.Sprintf("invalid CheckpointInfo size: got %d, want %d", unsafe.Sizeof(CheckpointInfo{}), CheckpointInfoSize))
	}
}

const CheckpointInfoSize = 64

// CheckpointInfo holds metadata about a single checkpoint (a persisted tree state).
// The checkpoint is identified by Version, since checkpoint == version.
type CheckpointInfo struct {
	Leaves   NodeSetInfo `json:"leaves"`
	Branches NodeSetInfo `json:"branches"`
	// Checkpoint is the identifier of this checkpoint. It is incremented for each checkpoint saved
	// and shouldn't be expected to align with the Version field.
	Checkpoint uint32 `json:"checkpoint"`
	// Version is the tree version at which this checkpoint was taken.
	Version uint32 `json:"version"`
	// RootID is the NodeID of the root node at this checkpoint.
	// This will be empty if the tree was empty at this checkpoint or if HaveRoot is false.
	// If HaveRoot is false, the root node is not stored in this changeset and must be obtained by replaying the WAL over
	// a previous changeset.
	RootID NodeID `json:"root_id"`
	// used to sanity check the length of the kv.dat file at this checkpoint
	KVEndOffset uint64 `json:"kv_end_offset"`
	// checksum of the checkpoint info record for data integrity verification
	CRC32 uint32 `json:"crc32"`
}

type NodeSetInfo struct {
	// StartOffset is the starting offset (in number of nodes) of this node set in the corresponding data file.
	StartOffset uint32 `json:"start_offset"`
	// Count is the total number of retained nodes in this node set.
	Count uint32 `json:"count"`
	// StartIndex is the 1-based indexe of the first retained node in this node set.
	StartIndex uint32 `json:"start_index"`
	// EndIndex is the 1-based index of the last retained node in this node set.
	EndIndex uint32 `json:"end_index"`
}

func (cp *CheckpointInfo) ComputeCRC32() uint32 {
	data := unsafe.Slice((*byte)(unsafe.Pointer(cp)), CheckpointInfoSize)
	return crc32.ChecksumIEEE(data[:CheckpointInfoSize-4]) // exclude CRC32 field itself
}

func (cp *CheckpointInfo) SetCRC32() {
	cp.CRC32 = cp.ComputeCRC32()
}

func (cp *CheckpointInfo) VerifyCRC32() bool {
	return cp.CRC32 == cp.ComputeCRC32()
}
