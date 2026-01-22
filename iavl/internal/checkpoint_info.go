package internal

import (
	"fmt"
	"unsafe"
)

func init() {
	if unsafe.Sizeof(CheckpointInfo{}) != CheckpointInfoSize {
		panic(fmt.Sprintf("invalid CheckpointInfo size: got %d, want %d", unsafe.Sizeof(CheckpointInfo{}), CheckpointInfoSize))
	}
}

const CheckpointInfoSize = 44

// CheckpointInfo holds metadata about a single checkpoint (a persisted tree state).
// The checkpoint is identified by Version, since checkpoint == version.
type CheckpointInfo struct {
	Leaves   NodeSetInfo
	Branches NodeSetInfo
	// Version is the tree version at which this checkpoint was taken.
	// This also serves as the checkpoint identifier.
	Version uint32
	RootID  NodeID
}

type NodeSetInfo struct {
	StartOffset uint32
	Count       uint32
	StartIndex  uint32
	EndIndex    uint32
}
