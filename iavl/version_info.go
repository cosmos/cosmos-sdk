package iavlx

import (
	"fmt"
	"unsafe"
)

func init() {
	if unsafe.Sizeof(VersionInfo{}) != VersionInfoSize {
		panic(fmt.Sprintf("invalid VersionInfo size: got %d, want %d", unsafe.Sizeof(VersionInfo{}), VersionInfoSize))
	}
}

const VersionInfoSize = 40

type VersionInfo struct {
	Leaves   NodeSetInfo
	Branches NodeSetInfo
	RootID   NodeID
}

type NodeSetInfo struct {
	StartOffset uint32
	Count       uint32
	StartIndex  uint32
	EndIndex    uint32
}
