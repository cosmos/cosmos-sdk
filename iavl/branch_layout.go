package iavlx

import (
	"fmt"
	"unsafe"
)

func init() {
	if unsafe.Sizeof(BranchLayout{}) != SizeBranch {
		panic(fmt.Sprintf("invalid BranchLayout size: got %d, want %d", unsafe.Sizeof(BranchLayout{}), SizeBranch))
	}
}

const (
	SizeBranch = 88
)

type BranchLayout struct {
	Id            NodeID
	Left          NodeID
	Right         NodeID
	LeftOffset    uint32 // absolute offset
	RightOffset   uint32 // absolute offset
	KeyOffset     uint32
	Height        uint8
	Size          uint32 // TODO 5 bytes?
	OrphanVersion uint32 // TODO 5 bytes?
	Hash          [32]byte
}

func (b BranchLayout) ID() NodeID {
	return b.Id
}
