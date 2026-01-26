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

const CheckpointInfoSize = 48

// CheckpointInfo holds metadata about a single checkpoint (a persisted tree state).
// The checkpoint is identified by Version, since checkpoint == version.
type CheckpointInfo struct {
	Leaves   NodeSetInfo
	Branches NodeSetInfo
	// Version is the tree version at which this checkpoint was taken.
	// This also serves as the checkpoint identifier.
	Version uint32
	// RootID is the NodeID of the root node at this checkpoint.
	// This will be empty if the tree was empty at this checkpoint or if HaveRoot is false.
	// If HaveRoot is false, the root node is not stored in this changeset and must be obtained by replaying the WAL over
	// a previous changeset.
	RootID NodeID
	// HaveRoot indicates whether the root node is stored in this changeset.
	HaveRoot    bool
	KVEndOffset KVOffset // used to sanity check the length of the kv.dat file at this checkpoint
	CRC32       uint32   // checksum of the checkpoint info record for data integrity verification
}

type NodeSetInfo struct {
	StartOffset uint32
	Count       uint32
	StartIndex  uint32
	EndIndex    uint32
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
