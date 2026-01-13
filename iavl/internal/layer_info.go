package internal

import (
	"fmt"
	"unsafe"
)

func init() {
	if unsafe.Sizeof(Layerinfo{}) != LayerInfoSize {
		panic(fmt.Sprintf("invalid Layerinfo size: got %d, want %d", unsafe.Sizeof(Layerinfo{}), LayerInfoSize))
	}
}

const LayerInfoSize = 48

type Layerinfo struct {
	Leaves   NodeSetInfo
	Branches NodeSetInfo
	Layer    uint32
	Version  uint32
	RootID   NodeID
}

type NodeSetInfo struct {
	StartOffset uint32
	Count       uint32
	StartIndex  uint32
	EndIndex    uint32
}
